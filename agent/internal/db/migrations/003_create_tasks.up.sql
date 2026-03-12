CREATE TABLE IF NOT EXISTS tasks (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id     UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title          TEXT        NOT NULL,
    description    TEXT        NOT NULL DEFAULT '',
    status         TEXT        NOT NULL DEFAULT 'backlog',
    position       INTEGER     NOT NULL DEFAULT 0,
    autonomy_level TEXT        NOT NULL DEFAULT 'supervised',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
