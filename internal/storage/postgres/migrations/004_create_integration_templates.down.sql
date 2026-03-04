ALTER TABLE integrations DROP COLUMN IF EXISTS template_id;

DROP TRIGGER IF EXISTS integration_templates_updated_at ON integration_templates;
DROP TABLE IF EXISTS integration_templates;
