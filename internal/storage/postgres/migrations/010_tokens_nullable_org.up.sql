-- API tokens are user-scoped, not org-scoped. Make org_id nullable.
ALTER TABLE api_tokens ALTER COLUMN org_id DROP NOT NULL;
