-- Drop indexes
DROP INDEX IF EXISTS idx_api_tokens_org;
DROP INDEX IF EXISTS idx_service_accounts_org;
DROP INDEX IF EXISTS idx_tasks_org;
DROP INDEX IF EXISTS idx_triggers_org;
DROP INDEX IF EXISTS idx_webhooks_org;
DROP INDEX IF EXISTS idx_sources_org;
DROP INDEX IF EXISTS idx_agent_configs_org;
DROP INDEX IF EXISTS idx_agents_org;

-- Remove org_id columns
ALTER TABLE api_tokens DROP COLUMN IF EXISTS org_id;
ALTER TABLE service_accounts DROP COLUMN IF EXISTS org_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS org_id;
ALTER TABLE triggers DROP COLUMN IF EXISTS org_id;
ALTER TABLE webhook_deliveries DROP COLUMN IF EXISTS org_id;
ALTER TABLE webhooks DROP COLUMN IF EXISTS org_id;
ALTER TABLE sources DROP COLUMN IF EXISTS org_id;
ALTER TABLE agent_attributes DROP COLUMN IF EXISTS org_id;
ALTER TABLE agent_secrets DROP COLUMN IF EXISTS org_id;
ALTER TABLE agent_configs DROP COLUMN IF EXISTS org_id;
ALTER TABLE agents DROP COLUMN IF EXISTS org_id;

-- Drop tables
DROP TABLE IF EXISTS team_memberships;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS organizations;
