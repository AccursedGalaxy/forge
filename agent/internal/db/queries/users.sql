-- name: GetDefaultUser :one
SELECT * FROM users WHERE id = '00000000-0000-0000-0000-000000000001'::uuid LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (id, name, email)
VALUES ($1, $2, $3)
RETURNING *;
