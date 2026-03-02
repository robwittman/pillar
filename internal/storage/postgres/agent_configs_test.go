//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robwittman/pillar/internal/domain"
)

func TestAgentConfigRepository_CreateAndGet(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	agentRepo := NewAgentRepository(pool)
	configRepo := NewAgentConfigRepository(pool)
	ctx := context.Background()

	// Create an agent first (FK constraint)
	agent := &domain.Agent{
		ID:       "cfg-agent-1",
		Name:     "config-test-agent",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, agentRepo.Create(ctx, agent))

	config := &domain.AgentConfig{
		AgentID:       "cfg-agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
		SystemPrompt:  "You are a helpful agent.",
		ModelParams:   domain.ModelParams{Temperature: 0.7, TopP: 0.9, MaxTokens: 4096},
		MCPServers: []domain.MCPServerConfig{
			{Name: "fs", TransportType: domain.MCPTransportStdio, Command: "mcp-fs", Args: []string{"--root", "/tmp"}},
		},
		ToolPermissions: domain.ToolPermissions{
			AllowedTools: []string{"read_file", "write_file"},
		},
		MaxIterations:      100,
		TokenBudget:        50000,
		TaskTimeoutSeconds: 300,
		EscalationRules: []domain.EscalationRule{
			{Name: "error-limit", Condition: "error_count > 3", Action: domain.EscalationActionPause, Message: "Too many errors"},
		},
	}

	err := configRepo.Create(ctx, config)
	require.NoError(t, err)
	assert.False(t, config.CreatedAt.IsZero())
	assert.False(t, config.UpdatedAt.IsZero())

	got, err := configRepo.Get(ctx, "cfg-agent-1")
	require.NoError(t, err)
	assert.Equal(t, domain.ModelProviderClaude, got.ModelProvider)
	assert.Equal(t, "claude-sonnet-4-20250514", got.ModelID)
	assert.Equal(t, "You are a helpful agent.", got.SystemPrompt)
	assert.InDelta(t, 0.7, got.ModelParams.Temperature, 0.001)
	assert.InDelta(t, 0.9, got.ModelParams.TopP, 0.001)
	assert.Equal(t, 4096, got.ModelParams.MaxTokens)
	require.Len(t, got.MCPServers, 1)
	assert.Equal(t, "fs", got.MCPServers[0].Name)
	assert.Equal(t, domain.MCPTransportStdio, got.MCPServers[0].TransportType)
	assert.Equal(t, []string{"--root", "/tmp"}, got.MCPServers[0].Args)
	assert.Equal(t, []string{"read_file", "write_file"}, got.ToolPermissions.AllowedTools)
	assert.Equal(t, 100, got.MaxIterations)
	assert.Equal(t, 50000, got.TokenBudget)
	assert.Equal(t, 300, got.TaskTimeoutSeconds)
	require.Len(t, got.EscalationRules, 1)
	assert.Equal(t, "error-limit", got.EscalationRules[0].Name)
}

func TestAgentConfigRepository_Update(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	agentRepo := NewAgentRepository(pool)
	configRepo := NewAgentConfigRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "cfg-upd-1",
		Name:     "update-test",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, agentRepo.Create(ctx, agent))

	config := &domain.AgentConfig{
		AgentID:         "cfg-upd-1",
		ModelProvider:   domain.ModelProviderClaude,
		ModelID:         "claude-sonnet-4-20250514",
		MaxIterations:   50,
		MCPServers:      []domain.MCPServerConfig{},
		EscalationRules: []domain.EscalationRule{},
	}
	require.NoError(t, configRepo.Create(ctx, config))

	config.ModelID = "claude-opus-4-20250514"
	config.MaxIterations = 200
	config.SystemPrompt = "Updated prompt"
	require.NoError(t, configRepo.Update(ctx, config))

	got, err := configRepo.Get(ctx, "cfg-upd-1")
	require.NoError(t, err)
	assert.Equal(t, "claude-opus-4-20250514", got.ModelID)
	assert.Equal(t, 200, got.MaxIterations)
	assert.Equal(t, "Updated prompt", got.SystemPrompt)
}

func TestAgentConfigRepository_Delete(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	agentRepo := NewAgentRepository(pool)
	configRepo := NewAgentConfigRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "cfg-del-1",
		Name:     "delete-test",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, agentRepo.Create(ctx, agent))

	config := &domain.AgentConfig{
		AgentID:         "cfg-del-1",
		ModelProvider:   domain.ModelProviderClaude,
		ModelID:         "claude-sonnet-4-20250514",
		MCPServers:      []domain.MCPServerConfig{},
		EscalationRules: []domain.EscalationRule{},
	}
	require.NoError(t, configRepo.Create(ctx, config))
	require.NoError(t, configRepo.Delete(ctx, "cfg-del-1"))

	_, err := configRepo.Get(ctx, "cfg-del-1")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}

func TestAgentConfigRepository_CascadeDeleteOnAgent(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	agentRepo := NewAgentRepository(pool)
	configRepo := NewAgentConfigRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "cfg-cascade-1",
		Name:     "cascade-test",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, agentRepo.Create(ctx, agent))

	config := &domain.AgentConfig{
		AgentID:         "cfg-cascade-1",
		ModelProvider:   domain.ModelProviderClaude,
		ModelID:         "claude-sonnet-4-20250514",
		MCPServers:      []domain.MCPServerConfig{},
		EscalationRules: []domain.EscalationRule{},
	}
	require.NoError(t, configRepo.Create(ctx, config))

	// Delete agent — config should cascade delete
	require.NoError(t, agentRepo.Delete(ctx, "cfg-cascade-1"))

	_, err := configRepo.Get(ctx, "cfg-cascade-1")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}

func TestAgentConfigRepository_NotFound(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	configRepo := NewAgentConfigRepository(pool)
	ctx := context.Background()

	_, err := configRepo.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)

	err = configRepo.Delete(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)

	err = configRepo.Update(ctx, &domain.AgentConfig{
		AgentID:         "nonexistent",
		ModelProvider:   domain.ModelProviderClaude,
		ModelID:         "x",
		MCPServers:      []domain.MCPServerConfig{},
		EscalationRules: []domain.EscalationRule{},
	})
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}

func TestAgentConfigRepository_DuplicateCreate(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	agentRepo := NewAgentRepository(pool)
	configRepo := NewAgentConfigRepository(pool)
	ctx := context.Background()

	agent := &domain.Agent{
		ID:       "cfg-dup-1",
		Name:     "dup-test",
		Status:   domain.AgentStatusPending,
		Metadata: map[string]string{},
		Labels:   map[string]string{},
	}
	require.NoError(t, agentRepo.Create(ctx, agent))

	config := &domain.AgentConfig{
		AgentID:         "cfg-dup-1",
		ModelProvider:   domain.ModelProviderClaude,
		ModelID:         "claude-sonnet-4-20250514",
		MCPServers:      []domain.MCPServerConfig{},
		EscalationRules: []domain.EscalationRule{},
	}
	require.NoError(t, configRepo.Create(ctx, config))

	err := configRepo.Create(ctx, config)
	assert.ErrorIs(t, err, domain.ErrConfigAlreadyExists)
}

// --- SecretStore integration tests ---

func TestSecretStore_PutGetDelete(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	store := NewSecretStore(pool)
	ctx := context.Background()

	require.NoError(t, store.Put(ctx, "test-secret", "secret-value"))

	val, err := store.Get(ctx, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, "secret-value", val)

	// Upsert
	require.NoError(t, store.Put(ctx, "test-secret", "updated-value"))
	val, err = store.Get(ctx, "test-secret")
	require.NoError(t, err)
	assert.Equal(t, "updated-value", val)

	require.NoError(t, store.Delete(ctx, "test-secret"))

	_, err = store.Get(ctx, "test-secret")
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
}

func TestSecretStore_GetNotFound(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	store := NewSecretStore(pool)
	ctx := context.Background()

	_, err := store.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, domain.ErrSecretNotFound)
}
