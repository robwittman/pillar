package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
	assert.Equal(t, ":9090", cfg.GRPCAddr)
	assert.Empty(t, cfg.Plugins)
}

func TestLoad_YAMLFile(t *testing.T) {
	content := `
http_addr: ":9999"
plugins:
  - name: keycloak
    path: /usr/local/bin/pillar-plugin-keycloak
    config:
      realm: agents
      admin_url: https://keycloak.example.com
  - name: gitea
    path: /usr/local/bin/pillar-plugin-gitea
    config:
      base_url: https://gitea.example.com
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pillar.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	t.Setenv("PILLAR_CONFIG_FILE", configPath)

	cfg, err := Load()
	require.NoError(t, err)

	// YAML value should be loaded
	assert.Equal(t, ":9999", cfg.HTTPAddr)

	// Plugins
	require.Len(t, cfg.Plugins, 2)
	assert.Equal(t, "keycloak", cfg.Plugins[0].Name)
	assert.Equal(t, "/usr/local/bin/pillar-plugin-keycloak", cfg.Plugins[0].Path)
	assert.Equal(t, "agents", cfg.Plugins[0].Config["realm"])
	assert.Equal(t, "gitea", cfg.Plugins[1].Name)
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	content := `
http_addr: ":9999"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pillar.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	t.Setenv("PILLAR_CONFIG_FILE", configPath)
	t.Setenv("PILLAR_HTTP_ADDR", ":7777")

	cfg, err := Load()
	require.NoError(t, err)

	// Env var should override YAML
	assert.Equal(t, ":7777", cfg.HTTPAddr)
}

func TestLoad_AuthConfig(t *testing.T) {
	content := `
auth:
  enabled: true
  session_secret: "test-secret"
  session_ttl: "12h"
  allow_signup: true
  providers:
    - type: local
      name: local
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pillar.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0644))

	t.Setenv("PILLAR_CONFIG_FILE", configPath)

	cfg, err := Load()
	require.NoError(t, err)

	assert.True(t, cfg.Auth.Enabled, "Auth.Enabled should be true")
	assert.Equal(t, "test-secret", cfg.Auth.SessionSecret)
	assert.Equal(t, "12h", cfg.Auth.SessionTTL)
	assert.True(t, cfg.Auth.AllowSignup)
	require.Len(t, cfg.Auth.Providers, 1, "should have 1 provider")
	assert.Equal(t, "local", cfg.Auth.Providers[0].Type)
	assert.Equal(t, "local", cfg.Auth.Providers[0].Name)
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(":::invalid"), 0644))

	t.Setenv("PILLAR_CONFIG_FILE", configPath)

	_, err := Load()
	assert.Error(t, err)
}

func TestLoad_MissingConfigFile(t *testing.T) {
	t.Setenv("PILLAR_CONFIG_FILE", "/nonexistent/pillar.yaml")

	_, err := Load()
	assert.Error(t, err)
}
