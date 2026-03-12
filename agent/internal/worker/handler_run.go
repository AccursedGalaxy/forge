package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/accursedgalaxy/forge/internal/db"
	"github.com/accursedgalaxy/forge/internal/stream"
)

// HandleRunSession processes a TypeRunSession job.
type runSessionHandler struct {
	queries     *db.Queries
	pool        *pgxpool.Pool
	broadcaster *stream.Broadcaster
}

func (h *runSessionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload RunSessionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal run session payload: %w", err)
	}

	slog.Info("job received: session:run", "session_id", payload.SessionID)

	// Update session status to planning
	_, err := h.queries.UpdateSessionStatus(ctx, db.UpdateSessionStatusParams{
		ID:     payload.SessionID,
		Status: "planning",
	})
	if err != nil {
		slog.Error("worker: update session status failed", "session_id", payload.SessionID, "err", err)
		return err
	}

	// Publish status event to SSE subscribers
	_ = h.broadcaster.Publish(ctx, payload.SessionID.String(), "session:status", map[string]string{
		"status": "planning",
	})

	// TODO(Step 3): spawn claude-cli agent execution
	slog.Info("agent exec pending Step 3", "session_id", payload.SessionID)
	return nil
}

// RegisterHandlers wires all worker task handlers into the asynq mux.
func RegisterHandlers(mux *asynq.ServeMux, queries *db.Queries, pool *pgxpool.Pool, broadcaster *stream.Broadcaster) {
	runHandler := &runSessionHandler{queries: queries, pool: pool, broadcaster: broadcaster}
	resumeHandler := &resumeSessionHandler{queries: queries, pool: pool, broadcaster: broadcaster}

	mux.HandleFunc(TypeRunSession, runHandler.ProcessTask)
	mux.HandleFunc(TypeResumeSession, resumeHandler.ProcessTask)
}
