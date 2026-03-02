package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/robwittman/pillar/internal/domain"
)

type ConfigService interface {
	CreateConfig(ctx context.Context, config *domain.AgentConfig, apiCredential string) error
	GetConfig(ctx context.Context, agentID string) (*domain.AgentConfig, error)
	GetConfigWithSecrets(ctx context.Context, agentID string) (*domain.AgentConfig, string, error)
	UpdateConfig(ctx context.Context, config *domain.AgentConfig, apiCredential string) error
	DeleteConfig(ctx context.Context, agentID string) error
}

type configService struct {
	configs domain.AgentConfigRepository
	agents  domain.AgentRepository
	secrets domain.SecretProvider
	logger  *slog.Logger
}

func NewConfigService(
	configs domain.AgentConfigRepository,
	agents domain.AgentRepository,
	secrets domain.SecretProvider,
	logger *slog.Logger,
) ConfigService {
	return &configService{
		configs: configs,
		agents:  agents,
		secrets: secrets,
		logger:  logger,
	}
}

func (s *configService) CreateConfig(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
	if err := s.validateConfig(config); err != nil {
		return err
	}

	if _, err := s.agents.Get(ctx, config.AgentID); err != nil {
		return err
	}

	if apiCredential != "" {
		ref := credentialRefKey(config.AgentID)
		if err := s.secrets.Put(ctx, ref, apiCredential); err != nil {
			return fmt.Errorf("storing api credential: %w", err)
		}
		config.APICredentialRef = ref
	}

	if config.MCPServers == nil {
		config.MCPServers = []domain.MCPServerConfig{}
	}
	if config.EscalationRules == nil {
		config.EscalationRules = []domain.EscalationRule{}
	}

	if err := s.configs.Create(ctx, config); err != nil {
		return err
	}

	s.logger.Info("agent config created", "agent_id", config.AgentID)
	return nil
}

func (s *configService) GetConfig(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
	return s.configs.Get(ctx, agentID)
}

func (s *configService) GetConfigWithSecrets(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) {
	config, err := s.configs.Get(ctx, agentID)
	if err != nil {
		return nil, "", err
	}

	var credential string
	if config.APICredentialRef != "" {
		credential, err = s.secrets.Get(ctx, config.APICredentialRef)
		if err != nil {
			s.logger.Warn("failed to resolve api credential", "agent_id", agentID, "error", err)
		}
	}

	return config, credential, nil
}

func (s *configService) UpdateConfig(ctx context.Context, config *domain.AgentConfig, apiCredential string) error {
	if err := s.validateConfig(config); err != nil {
		return err
	}

	if _, err := s.agents.Get(ctx, config.AgentID); err != nil {
		return err
	}

	if apiCredential != "" {
		ref := credentialRefKey(config.AgentID)
		if err := s.secrets.Put(ctx, ref, apiCredential); err != nil {
			return fmt.Errorf("storing api credential: %w", err)
		}
		config.APICredentialRef = ref
	}

	if config.MCPServers == nil {
		config.MCPServers = []domain.MCPServerConfig{}
	}
	if config.EscalationRules == nil {
		config.EscalationRules = []domain.EscalationRule{}
	}

	if err := s.configs.Update(ctx, config); err != nil {
		return err
	}

	s.logger.Info("agent config updated", "agent_id", config.AgentID)
	return nil
}

func (s *configService) DeleteConfig(ctx context.Context, agentID string) error {
	config, err := s.configs.Get(ctx, agentID)
	if err != nil {
		return err
	}

	if config.APICredentialRef != "" {
		if err := s.secrets.Delete(ctx, config.APICredentialRef); err != nil {
			s.logger.Warn("failed to delete api credential", "agent_id", agentID, "error", err)
		}
	}

	if err := s.configs.Delete(ctx, agentID); err != nil {
		return err
	}

	s.logger.Info("agent config deleted", "agent_id", agentID)
	return nil
}

func (s *configService) validateConfig(config *domain.AgentConfig) error {
	if config.ModelProvider == "" {
		return fmt.Errorf("%w: model_provider is required", domain.ErrInvalidConfig)
	}
	if config.ModelID == "" {
		return fmt.Errorf("%w: model_id is required", domain.ErrInvalidConfig)
	}
	if config.MaxIterations <= 0 {
		config.MaxIterations = 50
	}
	return nil
}

func credentialRefKey(agentID string) string {
	return "agent:" + agentID + ":api_credential"
}
