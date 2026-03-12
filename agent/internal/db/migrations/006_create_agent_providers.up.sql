CREATE TABLE IF NOT EXISTS agent_providers (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL UNIQUE,
    provider_type TEXT        NOT NULL DEFAULT 'claude',
    config        JSONB       NOT NULL DEFAULT '{}',
    is_default    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
