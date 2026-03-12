package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/accursedgalaxy/forge/internal/orchestrator"
)

// resumeSessionHandler processes TypeResumeSession jobs.
type resumeSessionHandler struct {
	orch *orchestrator.Orchestrator
}

func (h *resumeSessionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ResumeSessionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal resume session payload: %w", err)
	}

	slog.Info("job received: session:resume", "session_id", payload.SessionID)

	if err := h.orch.ResumeSession(ctx, payload.SessionID, payload.CorrectionPrompt); err != nil {
		slog.Error("worker: ResumeSession failed", "session_id", payload.SessionID, "err", err)
		return err
	}
	return nil
}
