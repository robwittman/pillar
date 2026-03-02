CREATE TYPE agent_status AS ENUM ('pending', 'running', 'stopped', 'error');

CREATE TABLE agents (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    status     agent_status NOT NULL DEFAULT 'pending',
    metadata   JSONB NOT NULL DEFAULT '{}',
    labels     JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_agents_labels ON agents USING GIN (labels);
CREATE INDEX idx_agents_status ON agents (status);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER agents_updated_at
    BEFORE UPDATE ON agents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
