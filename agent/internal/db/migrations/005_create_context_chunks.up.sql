CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS context_chunks (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    content    TEXT        NOT NULL,
    source     TEXT        NOT NULL DEFAULT '',
    embedding  vector(1536),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
