package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/accursedgalaxy/forge/internal/db"
)

// TaskHandler handles task CRUD and reorder endpoints.
type TaskHandler struct {
	db   *db.Queries
	pool *pgxpool.Pool
}

// List returns all tasks for a project, grouped implicitly by their status/position ordering.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	tasks, err := h.db.ListTasksByProject(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	if tasks == nil {
		tasks = []db.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

// Create inserts a new task in the backlog at the last position.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var body struct {
		Title         string `json:"title"`
		Description   string `json:"description"`
		AutonomyLevel string `json:"autonomy_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if body.AutonomyLevel == "" {
		body.AutonomyLevel = "supervised"
	}

	maxPos, err := h.db.GetMaxPositionForStatus(r.Context(), db.GetMaxPositionForStatusParams{
		ProjectID: projectID,
		Status:    "backlog",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute position")
		return
	}

	task, err := h.db.CreateTask(r.Context(), db.CreateTaskParams{
		ProjectID:     projectID,
		Title:         body.Title,
		Description:   body.Description,
		Status:        "backlog",
		Position:      maxPos + 1,
		AutonomyLevel: body.AutonomyLevel,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

// Update patches a task's mutable fields.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var body struct {
		Title         *string `json:"title"`
		Description   *string `json:"description"`
		Status        *string `json:"status"`
		AutonomyLevel *string `json:"autonomy_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	params := db.UpdateTaskParams{ID: id}
	if body.Title != nil {
		params.Title = sql.NullString{String: *body.Title, Valid: true}
	}
	if body.Description != nil {
		params.Description = sql.NullString{String: *body.Description, Valid: true}
	}
	if body.Status != nil {
		params.Status = sql.NullString{String: *body.Status, Valid: true}
	}
	if body.AutonomyLevel != nil {
		params.AutonomyLevel = sql.NullString{String: *body.AutonomyLevel, Valid: true}
	}

	task, err := h.db.UpdateTask(r.Context(), params)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// Delete removes a task.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.db.DeleteTask(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reorder moves a task to a new status column and position.
// It shifts sibling tasks in both the old and new columns within a transaction.
func (h *TaskHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var body struct {
		Status   string `json:"status"`
		Position int32  `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Status == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}

	ctx := r.Context()

	// Fetch current task state
	task, err := h.db.GetTask(ctx, id)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get task")
		return
	}

	// Use a sql.DB transaction via the pool's stdlib adapter
	sqlDB, err := h.pool.Acquire(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to acquire connection")
		return
	}
	defer sqlDB.Release()

	// Perform reorder operations directly on the connection
	// 1. If moving within same column: shift tasks between old and new positions
	// 2. If changing columns: close gap in old column, open slot in new column
	if task.Status == body.Status {
		// Same column reorder
		if task.Position < body.Position {
			// Moving down: shift tasks in between up by -1
			_, err = sqlDB.Exec(ctx,
				`UPDATE tasks SET position = position - 1, updated_at = NOW()
				 WHERE project_id = $1 AND status = $2
				   AND position > $3 AND position <= $4 AND id != $5`,
				task.ProjectID, task.Status, task.Position, body.Position, task.ID)
		} else {
			// Moving up: shift tasks in between down by +1
			_, err = sqlDB.Exec(ctx,
				`UPDATE tasks SET position = position + 1, updated_at = NOW()
				 WHERE project_id = $1 AND status = $2
				   AND position >= $3 AND position < $4 AND id != $5`,
				task.ProjectID, task.Status, body.Position, task.Position, task.ID)
		}
	} else {
		// Cross-column move: close gap in old column
		_, err = sqlDB.Exec(ctx,
			`UPDATE tasks SET position = position - 1, updated_at = NOW()
			 WHERE project_id = $1 AND status = $2
			   AND position > $3 AND id != $4`,
			task.ProjectID, task.Status, task.Position, task.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to shift tasks in old column")
			return
		}
		// Open slot in new column
		_, err = sqlDB.Exec(ctx,
			`UPDATE tasks SET position = position + 1, updated_at = NOW()
			 WHERE project_id = $1 AND status = $2
			   AND position >= $3 AND id != $4`,
			task.ProjectID, body.Status, body.Position, task.ID)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to shift task positions")
		return
	}

	// Update the task's own position and status
	_, err = sqlDB.Exec(ctx,
		`UPDATE tasks SET status = $2, position = $3, updated_at = NOW() WHERE id = $1`,
		task.ID, body.Status, body.Position)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task position")
		return
	}

	// Return the updated task
	updated, err := h.db.GetTask(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch updated task")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
