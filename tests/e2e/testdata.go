//go:build e2e

package e2e

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEnv holds the bootstrapped state for e2e tests.
// Each call to SetupTestEnv registers a fresh user — no pre-existing admin required.
type TestEnv struct {
	BaseURL  string
	Email    string
	Password string
	Client   *TestClient // Logged in
	OrgID    string      // User's personal org
	UserID   string
}

// SetupTestEnv registers a new user, logs in, and resolves the personal org.
// This makes every test self-contained — no dependency on admin credentials.
func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	suffix := randomSuffix()
	email := "e2e-" + suffix + "@test.local"
	password := "testpass-" + suffix

	env := &TestEnv{
		BaseURL:  testURL,
		Email:    email,
		Password: password,
		Client:   NewTestClient(testURL),
	}

	// Register (also logs in via session cookie).
	resp := env.Client.MustPost(t, "/auth/register", map[string]any{
		"email":        email,
		"password":     password,
		"display_name": "E2E User " + suffix,
	})
	RequireStatus(t, resp, http.StatusCreated)
	resp.Body.Close()

	// Get user info.
	resp = env.Client.MustGet(t, "/auth/me")
	RequireStatus(t, resp, http.StatusOK)
	var me map[string]any
	DecodeJSON(t, resp, &me)
	env.UserID = me["id"].(string)

	// List orgs to find personal org.
	resp = env.Client.MustGet(t, "/api/v1/organizations")
	RequireStatus(t, resp, http.StatusOK)
	var orgs []map[string]any
	DecodeJSON(t, resp, &orgs)
	require.NotEmpty(t, orgs, "user should have at least one org")
	env.OrgID = orgs[0]["id"].(string)
	env.Client.OrgID = env.OrgID

	return env
}

// randomSuffix returns a short random hex string for unique test names.
func randomSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
