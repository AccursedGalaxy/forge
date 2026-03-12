CREATE TABLE IF NOT EXISTS sessions (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id           UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    project_id        UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    session_type      TEXT        NOT NULL DEFAULT 'plan',
    status            TEXT        NOT NULL DEFAULT 'pending',
    claude_session_id TEXT,
    plan_steps        JSONB,
    error             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at      TIMESTAMPTZ
);
