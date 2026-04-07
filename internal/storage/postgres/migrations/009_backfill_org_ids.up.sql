-- Phase 5: Backfill org_ids for existing data and enforce NOT NULL.
--
-- Strategy:
-- 1. Create personal orgs for any users that don't have one
-- 2. Create a "default" org and assign orphaned resources to it
-- 3. Make org_id NOT NULL on all resource tables
-- 4. Update unique constraints to be per-org

-- Step 1: Create personal orgs for users that don't have one
INSERT INTO organizations (id, name, slug, personal, owner_id)
SELECT
    gen_random_uuid()::text,
    display_name || '''s Workspace',
    'personal-' || LEFT(id, 8),
    true,
    id
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM organizations WHERE personal = true AND owner_id = u.id
);

-- Create owner memberships for those personal orgs
INSERT INTO memberships (id, org_id, user_id, role)
SELECT
    gen_random_uuid()::text,
    o.id,
    o.owner_id,
    'owner'
FROM organizations o
WHERE o.personal = true
AND NOT EXISTS (
    SELECT 1 FROM memberships WHERE org_id = o.id AND user_id = o.owner_id
);

-- Step 2: Create a default org for orphaned resources (only if needed)
DO $$
DECLARE
    has_orphans boolean;
    first_user_id text;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM agents WHERE org_id IS NULL
        UNION ALL SELECT 1 FROM sources WHERE org_id IS NULL
        UNION ALL SELECT 1 FROM webhooks WHERE org_id IS NULL
        UNION ALL SELECT 1 FROM triggers WHERE org_id IS NULL
        UNION ALL SELECT 1 FROM tasks WHERE org_id IS NULL
    ) INTO has_orphans;

    IF has_orphans THEN
        SELECT id INTO first_user_id FROM users ORDER BY created_at ASC LIMIT 1;
        IF first_user_id IS NULL THEN
            first_user_id := 'system';
        END IF;

        INSERT INTO organizations (id, name, slug, personal, owner_id)
        VALUES ('default-org', 'Default Organization', 'default', false, first_user_id)
        ON CONFLICT (id) DO NOTHING;

        -- Make all existing users members of the default org
        INSERT INTO memberships (id, org_id, user_id, role)
        SELECT
            gen_random_uuid()::text,
            'default-org',
            u.id,
            CASE WHEN u.roles @> '["admin"]' THEN 'owner' ELSE 'member' END
        FROM users u
        WHERE NOT EXISTS (
            SELECT 1 FROM memberships WHERE org_id = 'default-org' AND user_id = u.id
        );
    END IF;
END $$;

-- Step 3: Backfill orphaned resources to default org
UPDATE agents SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE agent_configs SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE agent_secrets SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE agent_attributes SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE sources SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE webhooks SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE webhook_deliveries SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE triggers SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE tasks SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE service_accounts SET org_id = 'default-org' WHERE org_id IS NULL;
UPDATE api_tokens SET org_id = 'default-org' WHERE org_id IS NULL;

-- Step 4: Enforce NOT NULL
ALTER TABLE agents ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE agent_configs ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE agent_secrets ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE agent_attributes ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE sources ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE webhooks ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE webhook_deliveries ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE triggers ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE tasks ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE service_accounts ALTER COLUMN org_id SET NOT NULL;
ALTER TABLE api_tokens ALTER COLUMN org_id SET NOT NULL;

-- Step 5: Update service_accounts unique constraint to be per-org
ALTER TABLE service_accounts DROP CONSTRAINT IF EXISTS service_accounts_name_key;
CREATE UNIQUE INDEX idx_service_accounts_org_name ON service_accounts (org_id, name);
