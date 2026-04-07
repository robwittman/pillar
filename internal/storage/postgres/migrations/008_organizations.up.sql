-- Organizations
CREATE TABLE organizations (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,
    personal   BOOLEAN NOT NULL DEFAULT false,
    owner_id   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_organizations_personal_owner
    ON organizations (owner_id) WHERE personal = true;

CREATE TRIGGER organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Memberships (user <-> org with role)
CREATE TABLE memberships (
    id         TEXT PRIMARY KEY,
    org_id     TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, user_id)
);

CREATE INDEX idx_memberships_user ON memberships(user_id);

CREATE TRIGGER memberships_updated_at
    BEFORE UPDATE ON memberships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Teams
CREATE TABLE teams (
    id         TEXT PRIMARY KEY,
    org_id     TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, name)
);

CREATE TRIGGER teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Team memberships
CREATE TABLE team_memberships (
    id         TEXT PRIMARY KEY,
    team_id    TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (team_id, user_id)
);

-- Add org_id to all resource tables (nullable for migration)
ALTER TABLE agents ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE agent_configs ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE agent_secrets ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE agent_attributes ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE sources ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE webhooks ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE webhook_deliveries ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE triggers ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE tasks ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE service_accounts ADD COLUMN org_id TEXT REFERENCES organizations(id);
ALTER TABLE api_tokens ADD COLUMN org_id TEXT REFERENCES organizations(id);

-- Indexes for org-scoped queries
CREATE INDEX idx_agents_org ON agents(org_id);
CREATE INDEX idx_agent_configs_org ON agent_configs(org_id);
CREATE INDEX idx_sources_org ON sources(org_id);
CREATE INDEX idx_webhooks_org ON webhooks(org_id);
CREATE INDEX idx_triggers_org ON triggers(org_id);
CREATE INDEX idx_tasks_org ON tasks(org_id);
CREATE INDEX idx_service_accounts_org ON service_accounts(org_id);
CREATE INDEX idx_api_tokens_org ON api_tokens(org_id);
