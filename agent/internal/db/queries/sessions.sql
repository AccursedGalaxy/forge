-- name: ListSessionsByTask :many
SELECT * FROM sessions WHERE task_id = $1 ORDER BY created_at DESC;

-- name: CreateSession :one
INSERT INTO sessions (task_id, project_id, session_type, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = $1 LIMIT 1;

-- name: UpdateSessionStatus :one
UPDATE sessions
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateSessionCompleted :one
UPDATE sessions
SET status = 'completed', completed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateSessionError :one
UPDATE sessions
SET status = 'error', error = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateSessionClaudeID :one
UPDATE sessions
SET claude_session_id = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateSessionPlanSteps :one
UPDATE sessions
SET plan_steps = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
