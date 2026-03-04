CREATE TABLE integration_templates (
    id         TEXT PRIMARY KEY,
    type       TEXT NOT NULL,
    name       TEXT NOT NULL,
    config     JSONB NOT NULL DEFAULT '{}',
    selector   JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_integration_templates_type_name ON integration_templates(type, name);
CREATE INDEX idx_integration_templates_selector ON integration_templates USING GIN (selector);

CREATE TRIGGER integration_templates_updated_at
    BEFORE UPDATE ON integration_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

ALTER TABLE integrations ADD COLUMN template_id TEXT NOT NULL DEFAULT '';
