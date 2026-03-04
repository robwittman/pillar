package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

type IntegrationTemplateService interface {
	Create(ctx context.Context, template *domain.IntegrationTemplate) error
	Get(ctx context.Context, id string) (*domain.IntegrationTemplate, error)
	List(ctx context.Context) ([]*domain.IntegrationTemplate, error)
	Update(ctx context.Context, template *domain.IntegrationTemplate) error
	Delete(ctx context.Context, id string) error
	Preview(ctx context.Context, id string) ([]*domain.Agent, error)
	ProvisionForAgent(ctx context.Context, agentID string, labels map[string]string) error
}

type integrationTemplateService struct {
	templates    domain.IntegrationTemplateRepository
	integrations domain.IntegrationRepository
	agents       domain.AgentRepository
	logger       *slog.Logger
}

func NewIntegrationTemplateService(
	templates domain.IntegrationTemplateRepository,
	integrations domain.IntegrationRepository,
	agents domain.AgentRepository,
	logger *slog.Logger,
) IntegrationTemplateService {
	return &integrationTemplateService{
		templates:    templates,
		integrations: integrations,
		agents:       agents,
		logger:       logger,
	}
}

func (s *integrationTemplateService) Create(ctx context.Context, template *domain.IntegrationTemplate) error {
	if err := s.validateTemplate(template); err != nil {
		return err
	}

	template.ID = uuid.New().String()

	if template.Config == nil {
		template.Config = map[string]any{}
	}
	if template.Selector == nil {
		template.Selector = map[string]string{}
	}

	if err := s.templates.Create(ctx, template); err != nil {
		return err
	}

	s.logger.Info("integration template created", "id", template.ID, "type", template.Type, "name", template.Name)
	return nil
}

func (s *integrationTemplateService) Get(ctx context.Context, id string) (*domain.IntegrationTemplate, error) {
	return s.templates.Get(ctx, id)
}

func (s *integrationTemplateService) List(ctx context.Context) ([]*domain.IntegrationTemplate, error) {
	return s.templates.List(ctx)
}

func (s *integrationTemplateService) Update(ctx context.Context, template *domain.IntegrationTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalidIntegrationTemplate)
	}

	if template.Config == nil {
		template.Config = map[string]any{}
	}
	if template.Selector == nil {
		template.Selector = map[string]string{}
	}

	if err := s.templates.Update(ctx, template); err != nil {
		return err
	}

	s.logger.Info("integration template updated", "id", template.ID)
	return nil
}

func (s *integrationTemplateService) Delete(ctx context.Context, id string) error {
	if err := s.templates.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("integration template deleted", "id", id)
	return nil
}

func (s *integrationTemplateService) Preview(ctx context.Context, id string) ([]*domain.Agent, error) {
	template, err := s.templates.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	agents, err := s.agents.List(ctx)
	if err != nil {
		return nil, err
	}

	var matched []*domain.Agent
	for _, agent := range agents {
		if labelsMatch(agent.Labels, template.Selector) {
			matched = append(matched, agent)
		}
	}
	return matched, nil
}

func (s *integrationTemplateService) ProvisionForAgent(ctx context.Context, agentID string, labels map[string]string) error {
	templates, err := s.templates.FindMatchingTemplates(ctx, labels)
	if err != nil {
		return err
	}

	for _, tmpl := range templates {
		integration := &domain.Integration{
			ID:         uuid.New().String(),
			AgentID:    agentID,
			Type:       tmpl.Type,
			Name:       tmpl.Name,
			Config:     tmpl.Config,
			TemplateID: tmpl.ID,
		}
		if integration.Config == nil {
			integration.Config = map[string]any{}
		}

		if err := s.integrations.Create(ctx, integration); err != nil {
			if err == domain.ErrInvalidIntegration {
				s.logger.Debug("skipping duplicate integration from template",
					"agent_id", agentID, "template_id", tmpl.ID, "type", tmpl.Type, "name", tmpl.Name)
				continue
			}
			s.logger.Warn("failed to provision integration from template",
				"agent_id", agentID, "template_id", tmpl.ID, "error", err)
			continue
		}

		s.logger.Info("provisioned integration from template",
			"agent_id", agentID, "integration_id", integration.ID, "template_id", tmpl.ID)
	}

	return nil
}

func (s *integrationTemplateService) validateTemplate(template *domain.IntegrationTemplate) error {
	if template.Type == "" {
		return fmt.Errorf("%w: type is required", domain.ErrInvalidIntegrationTemplate)
	}
	if template.Name == "" {
		return fmt.Errorf("%w: name is required", domain.ErrInvalidIntegrationTemplate)
	}
	return nil
}

func labelsMatch(agentLabels, selector map[string]string) bool {
	for k, v := range selector {
		if agentLabels[k] != v {
			return false
		}
	}
	return true
}
