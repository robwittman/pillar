CREATE TABLE agent_configs (
    agent_id             TEXT PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
    model_provider       TEXT NOT NULL,
    model_id             TEXT NOT NULL,
    system_prompt        TEXT NOT NULL DEFAULT '',
    model_params         JSONB NOT NULL DEFAULT '{}',
    api_credential_ref   TEXT NOT NULL DEFAULT '',
    mcp_servers          JSONB NOT NULL DEFAULT '[]',
    tool_permissions     JSONB NOT NULL DEFAULT '{}',
    max_iterations       INTEGER NOT NULL DEFAULT 50,
    token_budget         INTEGER NOT NULL DEFAULT 0,
    task_timeout_seconds INTEGER NOT NULL DEFAULT 0,
    escalation_rules     JSONB NOT NULL DEFAULT '[]',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER agent_configs_updated_at
    BEFORE UPDATE ON agent_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TABLE agent_secrets (
    name       TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER agent_secrets_updated_at
    BEFORE UPDATE ON agent_secrets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
