-- Drop old tables
DROP TRIGGER IF EXISTS integration_templates_updated_at ON integration_templates;
DROP TABLE IF EXISTS integration_templates;
DROP TRIGGER IF EXISTS integrations_updated_at ON integrations;
DROP TABLE IF EXISTS integrations;

-- Webhooks
CREATE TABLE webhooks (
    id          TEXT PRIMARY KEY,
    url         TEXT NOT NULL,
    secret      TEXT NOT NULL,
    event_types JSONB NOT NULL DEFAULT '[]',
    status      TEXT NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_webhooks_status ON webhooks(status);
CREATE TRIGGER webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Webhook Deliveries
CREATE TABLE webhook_deliveries (
    id              TEXT PRIMARY KEY,
    webhook_id      TEXT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    response_code   INT,
    response_body   TEXT,
    status          TEXT NOT NULL DEFAULT 'pending',
    attempts        INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    next_retry_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_webhook_deliveries_pending ON webhook_deliveries(status, next_retry_at) WHERE status = 'pending';
CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);

-- Agent Attributes
CREATE TABLE agent_attributes (
    agent_id   TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    namespace  TEXT NOT NULL,
    value      JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (agent_id, namespace)
);
CREATE TRIGGER agent_attributes_updated_at BEFORE UPDATE ON agent_attributes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
