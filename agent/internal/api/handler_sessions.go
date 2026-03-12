package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/stream"
	"github.com/accursedgalaxy/forge/internal/worker"
)

// SessionHandler handles session lifecycle endpoints.
type SessionHandler struct {
	db          *db.Queries
	pool        *pgxpool.Pool
	asynqClient *asynq.Client
	broadcaster *stream.Broadcaster
}

// List returns all sessions for a task.
func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	sessions, err := h.db.ListSessionsByTask(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []db.Session{}
	}
	writeJSON(w, http.StatusOK, sessions)
}

// Create inserts a new session and enqueues a plan job.
func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TaskID      string `json:"task_id"`
		ProjectID   string `json:"project_id"`
		SessionType string `json:"session_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	taskID, err := uuid.Parse(body.TaskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task_id")
		return
	}
	projectID, err := uuid.Parse(body.ProjectID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project_id")
		return
	}
	sessionType := body.SessionType
	if sessionType == "" {
		sessionType = "plan"
	}

	session, err := h.db.CreateSession(r.Context(), db.CreateSessionParams{
		TaskID:      taskID,
		ProjectID:   projectID,
		SessionType: sessionType,
		Status:      "pending",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	task, err := worker.NewPlanSessionTask(worker.PlanSessionPayload{
		SessionID: session.ID,
		TaskID:    session.TaskID,
		ProjectID: session.ProjectID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create job")
		return
	}
	if _, err := h.asynqClient.EnqueueContext(r.Context(), task); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enqueue job")
		return
	}

	writeJSON(w, http.StatusCreated, session)
}

// Get returns a single session.
func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	session, err := h.db.GetSession(r.Context(), id)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// Approve validates a session is awaiting approval and enqueues the execute job.
func (h *SessionHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	session, err := h.db.GetSession(r.Context(), id)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if session.Status != "awaiting_approval" {
		writeError(w, http.StatusConflict, "session is not awaiting approval")
		return
	}

	updated, err := h.db.UpdateSessionStatus(r.Context(), db.UpdateSessionStatusParams{
		ID:     id,
		Status: "approved",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to approve session")
		return
	}

	// Enqueue TypeExecuteSession (not TypePlanSession).
	task, err := worker.NewExecuteSessionTask(worker.ExecuteSessionPayload{
		SessionID: session.ID,
		TaskID:    session.TaskID,
		ProjectID: session.ProjectID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create job")
		return
	}
	if _, err := h.asynqClient.EnqueueContext(r.Context(), task); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enqueue job")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// Interrupt pauses a running session.
func (h *SessionHandler) Interrupt(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	updated, err := h.db.UpdateSessionStatus(r.Context(), db.UpdateSessionStatusParams{
		ID:     id,
		Status: "paused",
	})
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to interrupt session")
		return
	}

	_ = h.broadcaster.Publish(r.Context(), id.String(), "session:status", map[string]string{
		"status": "paused",
	})

	writeJSON(w, http.StatusOK, updated)
}

// Resume re-enqueues a paused session with an optional correction prompt.
func (h *SessionHandler) Resume(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	var body struct {
		Prompt string `json:"prompt"`
	}
	// Ignore decode errors — prompt is optional.
	_ = json.NewDecoder(r.Body).Decode(&body)

	session, err := h.db.GetSession(r.Context(), id)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session")
		return
	}
	if session.Status != "paused" {
		writeError(w, http.StatusConflict, "session is not paused")
		return
	}

	task, err := worker.NewResumeSessionTask(worker.ResumeSessionPayload{
		SessionID:        session.ID,
		CorrectionPrompt: body.Prompt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create job")
		return
	}
	if _, err := h.asynqClient.EnqueueContext(r.Context(), task); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to enqueue job")
		return
	}

	writeJSON(w, http.StatusOK, session)
}

// Stream serves SSE events for a session.
func (h *SessionHandler) Stream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	h.broadcaster.SSEHandler(id)(w, r)
}
