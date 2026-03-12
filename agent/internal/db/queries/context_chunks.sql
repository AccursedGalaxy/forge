-- name: InsertContextChunk :one
INSERT INTO context_chunks (project_id, content, source, embedding)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListContextChunksByProject :many
SELECT * FROM context_chunks WHERE project_id = $1 ORDER BY created_at DESC;

-- name: DeleteContextChunk :exec
DELETE FROM context_chunks WHERE id = $1;

-- name: SearchSimilarChunks :many
SELECT *, (embedding <-> $2) AS distance
FROM context_chunks
WHERE project_id = $1
ORDER BY embedding <-> $2
LIMIT $3;
