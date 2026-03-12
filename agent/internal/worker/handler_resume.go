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

// resumeSessionHandler processes TypeResumeSession jobs.
type resumeSessionHandler struct {
	queries     *db.Queries
	pool        *pgxpool.Pool
	broadcaster *stream.Broadcaster
}

func (h *resumeSessionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ResumeSessionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal resume session payload: %w", err)
	}

	slog.Info("job received: session:resume", "session_id", payload.SessionID)

	// Update session status to running
	_, err := h.queries.UpdateSessionStatus(ctx, db.UpdateSessionStatusParams{
		ID:     payload.SessionID,
		Status: "running",
	})
	if err != nil {
		slog.Error("worker: update resume session status failed", "session_id", payload.SessionID, "err", err)
		return err
	}

	_ = h.broadcaster.Publish(ctx, payload.SessionID.String(), "session:status", map[string]string{
		"status": "running",
	})

	// TODO(Step 3): resume claude-cli agent
	slog.Info("agent resume pending Step 3", "session_id", payload.SessionID)
	return nil
}
