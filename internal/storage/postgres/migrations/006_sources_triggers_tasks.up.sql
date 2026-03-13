-- Sources (inbound webhook endpoints)
CREATE TABLE sources (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    secret     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER sources_updated_at BEFORE UPDATE ON sources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Triggers (routing rules: source events -> agent tasks)
CREATE TABLE triggers (
    id            TEXT PRIMARY KEY,
    source_id     TEXT NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
    agent_id      TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    filter        JSONB NOT NULL DEFAULT '{"conditions":[]}',
    task_template TEXT NOT NULL DEFAULT '',
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_triggers_source ON triggers(source_id) WHERE enabled = true;
CREATE INDEX idx_triggers_agent ON triggers(agent_id);
CREATE TRIGGER triggers_updated_at BEFORE UPDATE ON triggers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Tasks (work items delivered to agents)
CREATE TABLE tasks (
    id           TEXT PRIMARY KEY,
    agent_id     TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    trigger_id   TEXT REFERENCES triggers(id) ON DELETE SET NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    prompt       TEXT NOT NULL,
    context      JSONB,
    result       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_tasks_agent_status ON tasks(agent_id, status);
CREATE INDEX idx_tasks_status ON tasks(status) WHERE status = 'pending';
CREATE TRIGGER tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
