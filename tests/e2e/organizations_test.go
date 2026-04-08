//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListOrganizations(t *testing.T) {
	env := SetupTestEnv(t)

	resp := env.Client.MustGet(t, "/api/v1/organizations")
	RequireStatus(t, resp, http.StatusOK)
	var orgs []map[string]any
	DecodeJSON(t, resp, &orgs)
	require.NotEmpty(t, orgs)

	// Admin should have at least a personal org.
	hasPersonal := false
	for _, org := range orgs {
		if org["personal"].(bool) {
			hasPersonal = true
		}
	}
	assert.True(t, hasPersonal, "admin should have a personal org")
}

func TestCreateOrganization(t *testing.T) {
	env := SetupTestEnv(t)
	slug := "e2e-org-" + randomSuffix()

	// Create.
	resp := env.Client.MustPost(t, "/api/v1/organizations", map[string]any{
		"name": "E2E Test Org",
		"slug": slug,
	})
	RequireStatus(t, resp, http.StatusCreated)
	var org map[string]any
	DecodeJSON(t, resp, &org)
	orgID := org["id"].(string)
	assert.Equal(t, "E2E Test Org", org["name"])
	assert.Equal(t, slug, org["slug"])
	assert.False(t, org["personal"].(bool))

	t.Cleanup(func() {
		env.Client.Delete("/api/v1/organizations/" + orgID)
	})

	// Should appear in list.
	resp = env.Client.MustGet(t, "/api/v1/organizations")
	RequireStatus(t, resp, http.StatusOK)
	var orgs []map[string]any
	DecodeJSON(t, resp, &orgs)
	found := false
	for _, o := range orgs {
		if o["id"] == orgID {
			found = true
		}
	}
	assert.True(t, found)

	// Creator should be owner.
	resp = env.Client.MustGet(t, "/api/v1/organizations/"+orgID+"/members")
	RequireStatus(t, resp, http.StatusOK)
	var members []map[string]any
	DecodeJSON(t, resp, &members)
	require.Len(t, members, 1)
	assert.Equal(t, env.UserID, members[0]["user_id"])
	assert.Equal(t, "owner", members[0]["role"])

	// Delete.
	resp = env.Client.MustDelete(t, "/api/v1/organizations/"+orgID)
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestCannotDeletePersonalOrg(t *testing.T) {
	env := SetupTestEnv(t)

	resp := env.Client.MustDelete(t, "/api/v1/organizations/"+env.OrgID)
	RequireStatus(t, resp, http.StatusBadRequest)
	resp.Body.Close()
}

func TestOrgTeams(t *testing.T) {
	env := SetupTestEnv(t)
	slug := "e2e-team-org-" + randomSuffix()

	// Create an org for this test.
	resp := env.Client.MustPost(t, "/api/v1/organizations", map[string]any{
		"name": "Team Test Org",
		"slug": slug,
	})
	RequireStatus(t, resp, http.StatusCreated)
	var org map[string]any
	DecodeJSON(t, resp, &org)
	orgID := org["id"].(string)

	t.Cleanup(func() {
		env.Client.Delete("/api/v1/organizations/" + orgID)
	})

	// Create team.
	resp = env.Client.MustPost(t, "/api/v1/organizations/"+orgID+"/teams", map[string]any{
		"name": "engineering",
	})
	RequireStatus(t, resp, http.StatusCreated)
	var team map[string]any
	DecodeJSON(t, resp, &team)
	teamID := team["id"].(string)
	assert.Equal(t, "engineering", team["name"])

	// List teams.
	resp = env.Client.MustGet(t, "/api/v1/organizations/"+orgID+"/teams")
	RequireStatus(t, resp, http.StatusOK)
	var teams []map[string]any
	DecodeJSON(t, resp, &teams)
	require.Len(t, teams, 1)

	// Delete team.
	resp = env.Client.MustDelete(t, "/api/v1/organizations/"+orgID+"/teams/"+teamID)
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Verify gone.
	resp = env.Client.MustGet(t, "/api/v1/organizations/"+orgID+"/teams")
	RequireStatus(t, resp, http.StatusOK)
	DecodeJSON(t, resp, &teams)
	assert.Empty(t, teams)
}
