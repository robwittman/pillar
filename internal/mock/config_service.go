package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type ConfigService struct {
	CreateConfigFn         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error
	GetConfigFn            func(ctx context.Context, agentID string) (*domain.AgentConfig, error)
	GetConfigWithSecretsFn func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error)
	UpdateConfigFn         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error
	DeleteConfigFn         func(ctx context.Context, agentID string) error
}

func (m *ConfigService) CreateConfig(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
	return m.CreateConfigFn(ctx, config, apiCredential)
}

func (m *ConfigService) GetConfig(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
	return m.GetConfigFn(ctx, agentID)
}

func (m *ConfigService) GetConfigWithSecrets(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) {
	return m.GetConfigWithSecretsFn(ctx, agentID)
}

func (m *ConfigService) UpdateConfig(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
	return m.UpdateConfigFn(ctx, config, apiCredential)
}

func (m *ConfigService) DeleteConfig(ctx context.Context, agentID string) error {
	return m.DeleteConfigFn(ctx, agentID)
}
