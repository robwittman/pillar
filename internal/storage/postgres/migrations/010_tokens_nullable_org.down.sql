UPDATE api_tokens SET org_id = 'default-org' WHERE org_id IS NULL;
ALTER TABLE api_tokens ALTER COLUMN org_id SET NOT NULL;
