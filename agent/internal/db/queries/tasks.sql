-- name: ListTasksByProject :many
SELECT * FROM tasks WHERE project_id = $1 ORDER BY status, position;

-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, status, position, autonomy_level)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 LIMIT 1;

-- name: UpdateTask :one
UPDATE tasks
SET
    title          = COALESCE(sqlc.narg('title'), title),
    description    = COALESCE(sqlc.narg('description'), description),
    status         = COALESCE(sqlc.narg('status'), status),
    autonomy_level = COALESCE(sqlc.narg('autonomy_level'), autonomy_level),
    updated_at     = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;

-- name: GetMaxPositionForStatus :one
SELECT COALESCE(MAX(position), -1)::int AS max_position
FROM tasks
WHERE project_id = $1 AND status = $2;

-- name: UpdateTaskPositionAndStatus :exec
UPDATE tasks
SET status = $2, position = $3, updated_at = NOW()
WHERE id = $1;

-- name: ShiftTaskPositions :exec
UPDATE tasks
SET position = position + $4, updated_at = NOW()
WHERE project_id = $1
  AND status = $2
  AND position >= $3
  AND id != $5;
