package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

type IntegrationService interface {
	Create(ctx context.Context, integration *domain.Integration) error
	Get(ctx context.Context, id string) (*domain.Integration, error)
	List(ctx context.Context, agentID string) ([]*domain.Integration, error)
	Update(ctx context.Context, integration *domain.Integration) error
	Delete(ctx context.Context, id string) error
}

type integrationService struct {
	integrations domain.IntegrationRepository
	agents       domain.AgentRepository
	logger       *slog.Logger
}

func NewIntegrationService(
	integrations domain.IntegrationRepository,
	agents domain.AgentRepository,
	logger *slog.Logger,
) IntegrationService {
	return &integrationService{
		integrations: integrations,
		agents:       agents,
		logger:       logger,
	}
}

func (s *integrationService) Create(ctx context.Context, integration *domain.Integration) error {
	if err := s.validateIntegration(integration); err != nil {
		return err
	}

	if _, err := s.agents.Get(ctx, integration.AgentID); err != nil {
		return err
	}

	integration.ID = uuid.New().String()

	if integration.Config == nil {
		integration.Config = map[string]any{}
	}

	if err := s.integrations.Create(ctx, integration); err != nil {
		return err
	}

	s.logger.Info("integration created", "id", integration.ID, "agent_id", integration.AgentID, "type", integration.Type)
	return nil
}

func (s *integrationService) Get(ctx context.Context, id string) (*domain.Integration, error) {
	return s.integrations.Get(ctx, id)
}

func (s *integrationService) List(ctx context.Context, agentID string) ([]*domain.Integration, error) {
	return s.integrations.List(ctx, agentID)
}

func (s *integrationService) Update(ctx context.Context, integration *domain.Integration) error {
	if integration.Name == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalidIntegration)
	}

	if integration.Config == nil {
		integration.Config = map[string]any{}
	}

	if err := s.integrations.Update(ctx, integration); err != nil {
		return err
	}

	s.logger.Info("integration updated", "id", integration.ID)
	return nil
}

func (s *integrationService) Delete(ctx context.Context, id string) error {
	if err := s.integrations.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("integration deleted", "id", id)
	return nil
}

func (s *integrationService) validateIntegration(integration *domain.Integration) error {
	if integration.Type == "" {
		return fmt.Errorf("%w: type is required", domain.ErrInvalidIntegration)
	}
	if integration.Name == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalidIntegration)
	}
	return nil
}
