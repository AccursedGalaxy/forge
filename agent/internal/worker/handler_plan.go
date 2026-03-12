package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"github.com/accursedgalaxy/forge/internal/orchestrator"
)

// planSessionHandler processes TypePlanSession jobs.
type planSessionHandler struct {
	orch *orchestrator.Orchestrator
}

func (h *planSessionHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload PlanSessionPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal plan session payload: %w", err)
	}

	slog.Info("job received: session:plan", "session_id", payload.SessionID)

	if err := h.orch.PlanSession(ctx, payload.SessionID); err != nil {
		slog.Error("worker: PlanSession failed", "session_id", payload.SessionID, "err", err)
		return err
	}
	return nil
}
