//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthProviders(t *testing.T) {
	c := NewTestClient(testURL)
	resp := c.MustGet(t, "/auth/providers")
	RequireStatus(t, resp, http.StatusOK)
	var body map[string]any
	DecodeJSON(t, resp, &body)

	providers, ok := body["providers"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, providers)
}

func TestUnauthenticatedRequestReturns401(t *testing.T) {
	c := NewTestClient(testURL)
	resp := c.MustGet(t, "/api/v1/agents")
	RequireStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestRegisterLoginLogout(t *testing.T) {
	suffix := randomSuffix()
	email := "auth-test-" + suffix + "@test.local"
	password := "testpass-" + suffix

	c := NewTestClient(testURL)

	// Register.
	resp := c.MustPost(t, "/auth/register", map[string]any{
		"email":        email,
		"password":     password,
		"display_name": "Auth Test User",
	})
	RequireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	// Should be logged in via session cookie.
	resp = c.MustGet(t, "/auth/me")
	RequireStatus(t, resp, http.StatusOK)
	var me map[string]any
	DecodeJSON(t, resp, &me)
	assert.Equal(t, email, me["email"])

	// Logout.
	resp = c.MustPost(t, "/auth/logout", nil)
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Should be unauthenticated.
	resp = c.MustGet(t, "/auth/me")
	RequireStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()

	// Log back in with password.
	resp = c.MustPost(t, "/auth/login", map[string]any{
		"email":    email,
		"password": password,
	})
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Should be authenticated again.
	resp = c.MustGet(t, "/auth/me")
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}

func TestLoginInvalidCredentials(t *testing.T) {
	c := NewTestClient(testURL)
	resp := c.MustPost(t, "/auth/login", map[string]any{
		"email":    "nonexistent-" + randomSuffix() + "@test.local",
		"password": "wrongpassword",
	})
	RequireStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestRegisterCreatesPersonalOrg(t *testing.T) {
	env := SetupTestEnv(t)

	resp := env.Client.MustGet(t, "/api/v1/organizations")
	RequireStatus(t, resp, http.StatusOK)
	var orgs []map[string]any
	DecodeJSON(t, resp, &orgs)
	require.NotEmpty(t, orgs)
	assert.True(t, orgs[0]["personal"].(bool))
}

func TestAPITokenAuth(t *testing.T) {
	env := SetupTestEnv(t)

	// Create a token.
	resp := env.Client.MustPost(t, "/api/v1/auth/tokens", map[string]any{
		"name": "e2e-token-" + randomSuffix(),
	})
	RequireStatus(t, resp, http.StatusCreated)
	var tokenResp map[string]any
	DecodeJSON(t, resp, &tokenResp)
	rawToken := tokenResp["token"].(string)
	tokenID := tokenResp["meta"].(map[string]any)["id"].(string)
	assert.NotEmpty(t, rawToken)

	// Use token in a fresh client.
	tc := NewTestClient(testURL)
	tc.APIToken = rawToken
	tc.OrgID = env.OrgID

	resp = tc.MustGet(t, "/api/v1/agents")
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Revoke.
	resp = env.Client.MustDelete(t, "/api/v1/auth/tokens/"+tokenID)
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()

	// Token should no longer work.
	resp = tc.MustGet(t, "/api/v1/agents")
	RequireStatus(t, resp, http.StatusUnauthorized)
	resp.Body.Close()
}

func TestServiceAccountAuth(t *testing.T) {
	env := SetupTestEnv(t)

	// Create service account.
	resp := env.Client.MustPost(t, "/api/v1/auth/service-accounts", map[string]any{
		"name":        "e2e-sa-" + randomSuffix(),
		"description": "E2E test service account",
	})
	RequireStatus(t, resp, http.StatusCreated)
	var saResp map[string]any
	DecodeJSON(t, resp, &saResp)
	clientID := saResp["client_id"].(string)
	clientSecret := saResp["client_secret"].(string)
	assert.NotEmpty(t, clientID)
	assert.NotEmpty(t, clientSecret)

	// Authenticate with Basic auth.
	tc := NewTestClient(testURL)
	req, _ := http.NewRequest("GET", testURL+"/api/v1/agents", nil)
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("X-Org-ID", env.OrgID)
	resp2, err := tc.HTTPClient.Do(req)
	require.NoError(t, err)
	RequireStatus(t, resp2, http.StatusOK)
	resp2.Body.Close()

	// Cleanup.
	resp = env.Client.MustDelete(t, "/api/v1/auth/service-accounts/"+clientID)
	RequireStatus(t, resp, http.StatusOK)
	resp.Body.Close()
}
