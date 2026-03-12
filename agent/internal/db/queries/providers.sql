-- name: ListProviders :many
SELECT * FROM agent_providers ORDER BY created_at DESC;

-- name: CreateProvider :one
INSERT INTO agent_providers (name, provider_type, config, is_default)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProvider :one
SELECT * FROM agent_providers WHERE id = $1 LIMIT 1;

-- name: UpdateProvider :one
UPDATE agent_providers
SET
    name          = COALESCE(sqlc.narg('name'), name),
    provider_type = COALESCE(sqlc.narg('provider_type'), provider_type),
    config        = COALESCE(sqlc.narg('config'), config),
    is_default    = COALESCE(sqlc.narg('is_default'), is_default),
    updated_at    = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: GetDefaultProvider :one
SELECT * FROM agent_providers WHERE is_default = TRUE LIMIT 1;

-- name: ClearDefaultProvider :exec
UPDATE agent_providers SET is_default = FALSE, updated_at = NOW();
