-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (owner_id, name, description, repo_url, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1 LIMIT 1;

-- name: GetProjectWithTaskCounts :one
SELECT
    p.*,
    COUNT(t.id) FILTER (WHERE t.status = 'backlog')     AS backlog_count,
    COUNT(t.id) FILTER (WHERE t.status = 'in_progress') AS in_progress_count,
    COUNT(t.id) FILTER (WHERE t.status = 'in_review')   AS in_review_count,
    COUNT(t.id) FILTER (WHERE t.status = 'done')        AS done_count
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id
WHERE p.id = $1
GROUP BY p.id;

-- name: UpdateProject :one
UPDATE projects
SET
    name        = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    repo_url    = COALESCE(sqlc.narg('repo_url'), repo_url),
    status      = COALESCE(sqlc.narg('status'), status),
    updated_at  = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;
