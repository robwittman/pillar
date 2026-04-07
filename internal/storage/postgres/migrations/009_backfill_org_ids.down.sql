-- Revert unique constraint change
DROP INDEX IF EXISTS idx_service_accounts_org_name;
ALTER TABLE service_accounts ADD CONSTRAINT service_accounts_name_key UNIQUE (name);

-- Make org_id nullable again
ALTER TABLE agents ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE agent_configs ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE agent_secrets ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE agent_attributes ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE sources ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE webhooks ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE webhook_deliveries ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE triggers ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE tasks ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE service_accounts ALTER COLUMN org_id DROP NOT NULL;
ALTER TABLE api_tokens ALTER COLUMN org_id DROP NOT NULL;

-- Note: default org and personal orgs are left in place.
-- Removing them would require clearing org_id values first.
