DROP TRIGGER IF EXISTS agent_attributes_updated_at ON agent_attributes;
DROP TABLE IF EXISTS agent_attributes;
DROP INDEX IF EXISTS idx_webhook_deliveries_pending;
DROP INDEX IF EXISTS idx_webhook_deliveries_webhook;
DROP TABLE IF EXISTS webhook_deliveries;
DROP TRIGGER IF EXISTS webhooks_updated_at ON webhooks;
DROP INDEX IF EXISTS idx_webhooks_status;
DROP TABLE IF EXISTS webhooks;

-- Recreate integration tables
CREATE TABLE integrations (
    id          TEXT PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    config      JSONB NOT NULL DEFAULT '{}',
    template_id TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (agent_id, type, name)
);
CREATE INDEX idx_integrations_agent ON integrations(agent_id);
CREATE TRIGGER integrations_updated_at BEFORE UPDATE ON integrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TABLE integration_templates (
    id        TEXT PRIMARY KEY,
    type      TEXT NOT NULL,
    name      TEXT NOT NULL,
    config    JSONB NOT NULL DEFAULT '{}',
    selector  JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (type, name)
);
CREATE TRIGGER integration_templates_updated_at BEFORE UPDATE ON integration_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
