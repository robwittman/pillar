package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnvVars(t *testing.T) {
	t.Setenv("KC_CLIENT_ID", "pillar-admin")
	t.Setenv("KC_CLIENT_SECRET", "s3cret")

	config := map[string]string{
		"realm":         "agents",
		"client_id":     "${KC_CLIENT_ID}",
		"client_secret": "${KC_CLIENT_SECRET}",
		"admin_url":     "https://keycloak.example.com",
	}

	expanded := expandEnvVars(config)

	assert.Equal(t, "agents", expanded["realm"])
	assert.Equal(t, "pillar-admin", expanded["client_id"])
	assert.Equal(t, "s3cret", expanded["client_secret"])
	assert.Equal(t, "https://keycloak.example.com", expanded["admin_url"])
}

func TestExpandEnvVars_UnsetVar(t *testing.T) {
	config := map[string]string{
		"key": "${DEFINITELY_NOT_SET_12345}",
	}

	expanded := expandEnvVars(config)
	assert.Equal(t, "", expanded["key"])
}

func TestExpandEnvVars_MixedContent(t *testing.T) {
	t.Setenv("HOST", "keycloak.example.com")

	config := map[string]string{
		"url": "https://${HOST}/auth",
	}

	expanded := expandEnvVars(config)
	assert.Equal(t, "https://keycloak.example.com/auth", expanded["url"])
}
