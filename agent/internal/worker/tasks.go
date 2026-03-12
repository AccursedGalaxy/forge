package worker

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypePlanSession    = "session:plan"    // enqueued by POST /api/sessions
	TypeExecuteSession = "session:execute" // enqueued by POST /api/sessions/:id/approve
	TypeResumeSession  = "session:resume"  // enqueued by POST /api/sessions/:id/resume
)

// PlanSessionPayload is the payload for a TypePlanSession job.
type PlanSessionPayload struct {
	SessionID uuid.UUID `json:"session_id"`
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

// ExecuteSessionPayload is the payload for a TypeExecuteSession job.
type ExecuteSessionPayload struct {
	SessionID uuid.UUID `json:"session_id"`
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

// ResumeSessionPayload is the payload for a TypeResumeSession job.
type ResumeSessionPayload struct {
	SessionID        uuid.UUID `json:"session_id"`
	CorrectionPrompt string    `json:"correction_prompt"`
}

// NewPlanSessionTask creates an asynq task for the planning phase.
func NewPlanSessionTask(payload PlanSessionPayload) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker: marshal plan session payload: %w", err)
	}
	return asynq.NewTask(TypePlanSession, b), nil
}

// NewExecuteSessionTask creates an asynq task for the execution phase.
func NewExecuteSessionTask(payload ExecuteSessionPayload) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker: marshal execute session payload: %w", err)
	}
	return asynq.NewTask(TypeExecuteSession, b), nil
}

// NewResumeSessionTask creates an asynq task for resuming a paused session.
func NewResumeSessionTask(payload ResumeSessionPayload) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker: marshal resume session payload: %w", err)
	}
	return asynq.NewTask(TypeResumeSession, b), nil
}
