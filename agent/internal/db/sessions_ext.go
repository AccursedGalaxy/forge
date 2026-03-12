package db

// sessions_ext.go — hand-written DB helpers that complement the sqlc-generated code.
// These are kept separate so they never get clobbered by `sqlc generate`.

import (
	"context"

	"github.com/google/uuid"
)

// CompleteSession transitions a session to status='done', sets completed_at, and returns the updated row.
func (q *Queries) CompleteSession(ctx context.Context, id uuid.UUID) (Session, error) {
	const sql = `
		UPDATE sessions
		SET status = 'done', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
		RETURNING id, task_id, project_id, session_type, status, claude_session_id, plan_steps, error, created_at, updated_at, completed_at`

	row := q.db.QueryRowContext(ctx, sql, id)
	var s Session
	err := row.Scan(
		&s.ID, &s.TaskID, &s.ProjectID, &s.SessionType, &s.Status,
		&s.ClaudeSessionID, &s.PlanSteps, &s.Error,
		&s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	return s, err
}

// UpdateSessionType changes the session_type of an existing session.
func (q *Queries) UpdateSessionType(ctx context.Context, id uuid.UUID, sessionType string) (Session, error) {
	const sql = `
		UPDATE sessions
		SET session_type = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, task_id, project_id, session_type, status, claude_session_id, plan_steps, error, created_at, updated_at, completed_at`

	row := q.db.QueryRowContext(ctx, sql, id, sessionType)
	var s Session
	err := row.Scan(
		&s.ID, &s.TaskID, &s.ProjectID, &s.SessionType, &s.Status,
		&s.ClaudeSessionID, &s.PlanSteps, &s.Error,
		&s.CreatedAt, &s.UpdatedAt, &s.CompletedAt,
	)
	return s, err
}
