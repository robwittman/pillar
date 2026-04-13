//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentCRUD(t *testing.T) {
	env := SetupTestEnv(t)
	c := env.Client
	agentName := "e2e-agent-" + randomSuffix()

	// Create.
	resp := c.MustPost(t, "/api/v1/agents", map[string]any{
		"name":     agentName,
		"metadata": map[string]string{"env": "test"},
		"labels":   map[string]string{"team": "e2e"},
	})
	RequireStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	DecodeJSON(t, resp, &agent)
	agentID := agent["id"].(string)
	assert.NotEmpty(t, agentID)
	assert.Equal(t, agentName, agent["name"])
	assert.Equal(t, "pending", agent["status"])

	t.Cleanup(func() {
		c.Delete("/api/v1/agents/" + agentID)
	})

	// Get.
	resp = c.MustGet(t, "/api/v1/agents/"+agentID)
	RequireStatus(t, resp, http.StatusOK)
	var fetched map[string]any
	DecodeJSON(t, resp, &fetched)
	assert.Equal(t, agentName, fetched["name"])
	metadata := fetched["metadata"].(map[string]any)
	assert.Equal(t, "test", metadata["env"])

	// Update.
	resp = c.MustPut(t, "/api/v1/agents/"+agentID, map[string]any{
		"name": agentName + "-updated",
	})
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// List.
	resp = c.MustGet(t, "/api/v1/agents")
	RequireStatus(t, resp, http.StatusOK)
	var agents []map[string]any
	DecodeJSON(t, resp, &agents)
	found := false
	for _, a := range agents {
		if a["id"] == agentID {
			found = true
			assert.Equal(t, agentName+"-updated", a["name"])
		}
	}
	assert.True(t, found, "created agent should appear in list")

	// Delete.
	resp = c.MustDelete(t, "/api/v1/agents/"+agentID)
	RequireStatus(t, resp, http.StatusNoContent)
	resp.Body.Close()

	// Verify gone.
	resp = c.MustGet(t, "/api/v1/agents/"+agentID)
	RequireStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

func TestAgentNotFoundReturns404(t *testing.T) {
	env := SetupTestEnv(t)
	resp := env.Client.MustGet(t, "/api/v1/agents/nonexistent-id")
	RequireStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}

func TestAgentOrgIsolation(t *testing.T) {
	env := SetupTestEnv(t)

	// Create an agent in the admin's org.
	agentName := "e2e-isolated-" + randomSuffix()
	resp := env.Client.MustPost(t, "/api/v1/agents", map[string]any{
		"name": agentName,
	})
	RequireStatus(t, resp, http.StatusCreated)
	var agent map[string]any
	DecodeJSON(t, resp, &agent)
	agentID := agent["id"].(string)

	t.Cleanup(func() {
		env.Client.Delete("/api/v1/agents/" + agentID)
	})

	// Register a second user (gets their own personal org).
	c2 := NewTestClient(testURL)
	resp = c2.MustPost(t, "/auth/register", map[string]any{
		"email":    "isolation-" + randomSuffix() + "@test.local",
		"password": "testpassword123",
	})
	RequireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	// Get user2's org.
	resp = c2.MustGet(t, "/api/v1/organizations")
	RequireStatus(t, resp, http.StatusOK)
	var orgs []map[string]any
	DecodeJSON(t, resp, &orgs)
	require.NotEmpty(t, orgs)
	c2.OrgID = orgs[0]["id"].(string)

	// User2 should NOT see admin's agent.
	resp = c2.MustGet(t, "/api/v1/agents")
	RequireStatus(t, resp, http.StatusOK)
	var user2Agents []map[string]any
	DecodeJSON(t, resp, &user2Agents)
	for _, a := range user2Agents {
		assert.NotEqual(t, agentID, a["id"], "user2 should not see admin's agent")
	}

	// Direct get should also fail.
	resp = c2.MustGet(t, "/api/v1/agents/"+agentID)
	RequireStatus(t, resp, http.StatusNotFound)
	resp.Body.Close()
}
