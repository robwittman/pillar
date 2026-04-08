//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestClient is an HTTP client for e2e tests targeting a running Pillar server.
// It carries auth state (session cookie or API token) and org context.
type TestClient struct {
	BaseURL    string
	HTTPClient *http.Client
	APIToken   string
	OrgID      string
}

func NewTestClient(baseURL string) *TestClient {
	jar, _ := cookiejar.New(nil)
	return &TestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Jar: jar,
		},
	}
}

func (c *TestClient) Do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}
	if c.OrgID != "" {
		req.Header.Set("X-Org-ID", c.OrgID)
	}

	return c.HTTPClient.Do(req)
}

func (c *TestClient) Get(path string) (*http.Response, error) {
	return c.Do("GET", path, nil)
}

func (c *TestClient) Post(path string, body any) (*http.Response, error) {
	return c.Do("POST", path, body)
}

func (c *TestClient) Put(path string, body any) (*http.Response, error) {
	return c.Do("PUT", path, body)
}

func (c *TestClient) Delete(path string) (*http.Response, error) {
	return c.Do("DELETE", path, nil)
}

// MustGet performs a GET and fails the test on transport error.
func (c *TestClient) MustGet(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := c.Get(path)
	require.NoError(t, err)
	return resp
}

func (c *TestClient) MustPost(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	resp, err := c.Post(path, body)
	require.NoError(t, err)
	return resp
}

func (c *TestClient) MustPut(t *testing.T, path string, body any) *http.Response {
	t.Helper()
	resp, err := c.Put(path, body)
	require.NoError(t, err)
	return resp
}

func (c *TestClient) MustDelete(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := c.Delete(path)
	require.NoError(t, err)
	return resp
}

// RequireStatus asserts the response has the expected status code.
func RequireStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}

// DecodeJSON reads and decodes the response body into target.
func DecodeJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
}
