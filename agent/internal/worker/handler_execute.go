package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/accursedgalaxy/forge/internal/orchestrator"
)

// executeSessionHandler processes TypeExecuteSession jobs.
type executeSessionHandler struct {
	orch *orchestrator.Orchestrator
}

func (h *executeSessionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ExecuteSessionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal execute session payload: %w", err)
	}

	slog.Info("job received: session:execute", "session_id", payload.SessionID)

	if err := h.orch.ExecuteSession(ctx, payload.SessionID); err != nil {
		slog.Error("worker: ExecuteSession failed", "session_id", payload.SessionID, "err", err)
		return err
	}
	return nil
}
