package worker

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeRunSession    = "session:run"
	TypeResumeSession = "session:resume"
)

// RunSessionPayload is the payload for a TypeRunSession job.
type RunSessionPayload struct {
	SessionID uuid.UUID `json:"session_id"`
	TaskID    uuid.UUID `json:"task_id"`
	ProjectID uuid.UUID `json:"project_id"`
}

// ResumeSessionPayload is the payload for a TypeResumeSession job.
type ResumeSessionPayload struct {
	SessionID uuid.UUID `json:"session_id"`
}

// NewRunSessionTask creates an asynq task for running a new session.
func NewRunSessionTask(payload RunSessionPayload) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker: marshal run session payload: %w", err)
	}
	return asynq.NewTask(TypeRunSession, b), nil
}

// NewResumeSessionTask creates an asynq task for resuming a paused session.
func NewResumeSessionTask(payload ResumeSessionPayload) (*asynq.Task, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker: marshal resume session payload: %w", err)
	}
	return asynq.NewTask(TypeResumeSession, b), nil
}
