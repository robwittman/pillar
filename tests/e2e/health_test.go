//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	c := NewTestClient(testURL)
	resp := c.MustGet(t, "/health")
	RequireStatus(t, resp, http.StatusOK)
	var body map[string]string
	DecodeJSON(t, resp, &body)
	assert.Equal(t, "ok", body["status"])
}
