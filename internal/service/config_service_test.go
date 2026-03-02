package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfigService(
	configRepo *mock.AgentConfigRepository,
	agentRepo *mock.AgentRepository,
	secrets *mock.SecretProvider,
) service.ConfigService {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return service.NewConfigService(configRepo, agentRepo, secrets, logger)
}

// --- CreateConfig ---

func TestCreateConfig_Success(t *testing.T) {
	var capturedConfig *domain.AgentConfig
	configRepo := &mock.AgentConfigRepository{
		CreateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			capturedConfig = config
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{
		PutFn: func(ctx context.Context, name string, value string) error {
			return nil
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
	}
	err := svc.CreateConfig(context.Background(), config, "sk-test-key")
	require.NoError(t, err)
	assert.Equal(t, "agent:agent-1:api_credential", capturedConfig.APICredentialRef)
	assert.NotNil(t, capturedConfig.MCPServers)
	assert.NotNil(t, capturedConfig.EscalationRules)
}

func TestCreateConfig_NoCredential(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		CreateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
	}
	err := svc.CreateConfig(context.Background(), config, "")
	require.NoError(t, err)
	assert.Equal(t, "", config.APICredentialRef)
}

func TestCreateConfig_AgentNotFound(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "missing",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
	}
	err := svc.CreateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

func TestCreateConfig_MissingProvider(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID: "agent-1",
		ModelID: "claude-sonnet-4-20250514",
	}
	err := svc.CreateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrInvalidConfig)
}

func TestCreateConfig_MissingModelID(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
	}
	err := svc.CreateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrInvalidConfig)
}

func TestCreateConfig_DefaultMaxIterations(t *testing.T) {
	var captured *domain.AgentConfig
	configRepo := &mock.AgentConfigRepository{
		CreateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			captured = config
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
		MaxIterations: 0,
	}
	err := svc.CreateConfig(context.Background(), config, "")
	require.NoError(t, err)
	assert.Equal(t, 50, captured.MaxIterations)
}

func TestCreateConfig_AlreadyExists(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		CreateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			return domain.ErrConfigAlreadyExists
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
	}
	err := svc.CreateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrConfigAlreadyExists)
}

func TestCreateConfig_SecretStoreError(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secretErr := errors.New("vault unavailable")
	secrets := &mock.SecretProvider{
		PutFn: func(ctx context.Context, name string, value string) error {
			return secretErr
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
	}
	err := svc.CreateConfig(context.Background(), config, "sk-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storing api credential")
}

// --- GetConfig ---

func TestGetConfig_Success(t *testing.T) {
	expected := &domain.AgentConfig{AgentID: "agent-1", ModelProvider: domain.ModelProviderClaude}
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return expected, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config, err := svc.GetConfig(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, expected, config)
}

func TestGetConfig_NotFound(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return nil, domain.ErrConfigNotFound
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	_, err := svc.GetConfig(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}

// --- GetConfigWithSecrets ---

func TestGetConfigWithSecrets_Success(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{
				AgentID:          agentID,
				APICredentialRef: "agent:agent-1:api_credential",
			}, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{
		GetFn: func(ctx context.Context, name string) (string, error) {
			return "sk-resolved-key", nil
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config, credential, err := svc.GetConfigWithSecrets(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, "agent-1", config.AgentID)
	assert.Equal(t, "sk-resolved-key", credential)
}

func TestGetConfigWithSecrets_NoRef(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{AgentID: agentID}, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config, credential, err := svc.GetConfigWithSecrets(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "", credential)
}

func TestGetConfigWithSecrets_SecretResolveFails(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{
				AgentID:          agentID,
				APICredentialRef: "agent:agent-1:api_credential",
			}, nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{
		GetFn: func(ctx context.Context, name string) (string, error) {
			return "", domain.ErrSecretNotFound
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config, credential, err := svc.GetConfigWithSecrets(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "", credential)
}

func TestGetConfigWithSecrets_ConfigNotFound(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return nil, domain.ErrConfigNotFound
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	_, _, err := svc.GetConfigWithSecrets(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}

// --- UpdateConfig ---

func TestUpdateConfig_Success(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		UpdateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{
		PutFn: func(ctx context.Context, name string, value string) error {
			return nil
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
		MaxIterations: 100,
	}
	err := svc.UpdateConfig(context.Background(), config, "new-key")
	require.NoError(t, err)
	assert.Equal(t, "agent:agent-1:api_credential", config.APICredentialRef)
}

func TestUpdateConfig_NoNewCredential(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		UpdateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:          "agent-1",
		ModelProvider:    domain.ModelProviderClaude,
		ModelID:          "claude-sonnet-4-20250514",
		APICredentialRef: "existing-ref",
		MaxIterations:    100,
	}
	err := svc.UpdateConfig(context.Background(), config, "")
	require.NoError(t, err)
	assert.Equal(t, "existing-ref", config.APICredentialRef)
}

func TestUpdateConfig_AgentNotFound(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return nil, domain.ErrAgentNotFound
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "missing",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
		MaxIterations: 50,
	}
	err := svc.UpdateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrAgentNotFound)
}

func TestUpdateConfig_ConfigNotFound(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		UpdateFn: func(ctx context.Context, config *domain.AgentConfig) error {
			return domain.ErrConfigNotFound
		},
	}
	agentRepo := &mock.AgentRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Agent, error) {
			return &domain.Agent{ID: id}, nil
		},
	}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID:       "agent-1",
		ModelProvider: domain.ModelProviderClaude,
		ModelID:       "claude-sonnet-4-20250514",
		MaxIterations: 50,
	}
	err := svc.UpdateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}

func TestUpdateConfig_ValidationFailure(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	config := &domain.AgentConfig{
		AgentID: "agent-1",
	}
	err := svc.UpdateConfig(context.Background(), config, "")
	assert.ErrorIs(t, err, domain.ErrInvalidConfig)
}

// --- DeleteConfig ---

func TestDeleteConfig_Success(t *testing.T) {
	var secretDeleted bool
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{
				AgentID:          agentID,
				APICredentialRef: "agent:agent-1:api_credential",
			}, nil
		},
		DeleteFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{
		DeleteFn: func(ctx context.Context, name string) error {
			secretDeleted = true
			return nil
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	err := svc.DeleteConfig(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.True(t, secretDeleted)
}

func TestDeleteConfig_NoCredentialRef(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{AgentID: agentID}, nil
		},
		DeleteFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	err := svc.DeleteConfig(context.Background(), "agent-1")
	require.NoError(t, err)
}

func TestDeleteConfig_SecretDeleteFailureNonFatal(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return &domain.AgentConfig{
				AgentID:          agentID,
				APICredentialRef: "ref",
			}, nil
		},
		DeleteFn: func(ctx context.Context, agentID string) error {
			return nil
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{
		DeleteFn: func(ctx context.Context, name string) error {
			return errors.New("vault error")
		},
	}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	err := svc.DeleteConfig(context.Background(), "agent-1")
	require.NoError(t, err)
}

func TestDeleteConfig_ConfigNotFound(t *testing.T) {
	configRepo := &mock.AgentConfigRepository{
		GetFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
			return nil, domain.ErrConfigNotFound
		},
	}
	agentRepo := &mock.AgentRepository{}
	secrets := &mock.SecretProvider{}
	svc := newTestConfigService(configRepo, agentRepo, secrets)

	err := svc.DeleteConfig(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrConfigNotFound)
}
