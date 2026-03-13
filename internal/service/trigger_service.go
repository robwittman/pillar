package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

type TriggerService interface {
	Create(ctx context.Context, sourceID, agentID, name string, filter domain.TriggerFilter, taskTemplate string) (*domain.Trigger, error)
	Get(ctx context.Context, id string) (*domain.Trigger, error)
	List(ctx context.Context) ([]*domain.Trigger, error)
	ListBySource(ctx context.Context, sourceID string) ([]*domain.Trigger, error)
	Update(ctx context.Context, id string, name string, filter *domain.TriggerFilter, taskTemplate *string, enabled *bool) (*domain.Trigger, error)
	Delete(ctx context.Context, id string) error
}

type triggerService struct {
	repo   domain.TriggerRepository
	logger *slog.Logger
}

func NewTriggerService(repo domain.TriggerRepository, logger *slog.Logger) TriggerService {
	return &triggerService{repo: repo, logger: logger}
}

func (s *triggerService) Create(ctx context.Context, sourceID, agentID, name string, filter domain.TriggerFilter, taskTemplate string) (*domain.Trigger, error) {
	if sourceID == "" || agentID == "" || name == "" {
		return nil, domain.ErrInvalidTrigger
	}

	if filter.Conditions == nil {
		filter.Conditions = []domain.FilterCondition{}
	}

	trigger := &domain.Trigger{
		ID:           uuid.New().String(),
		SourceID:     sourceID,
		AgentID:      agentID,
		Name:         name,
		Filter:       filter,
		TaskTemplate: taskTemplate,
		Enabled:      true,
	}

	if err := s.repo.Create(ctx, trigger); err != nil {
		return nil, err
	}

	s.logger.Info("trigger created", "id", trigger.ID, "source_id", sourceID, "agent_id", agentID)
	return trigger, nil
}

func (s *triggerService) Get(ctx context.Context, id string) (*domain.Trigger, error) {
	return s.repo.Get(ctx, id)
}

func (s *triggerService) List(ctx context.Context) ([]*domain.Trigger, error) {
	return s.repo.List(ctx)
}

func (s *triggerService) ListBySource(ctx context.Context, sourceID string) ([]*domain.Trigger, error) {
	return s.repo.ListBySource(ctx, sourceID)
}

func (s *triggerService) Update(ctx context.Context, id string, name string, filter *domain.TriggerFilter, taskTemplate *string, enabled *bool) (*domain.Trigger, error) {
	trigger, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		trigger.Name = name
	}
	if filter != nil {
		if filter.Conditions == nil {
			filter.Conditions = []domain.FilterCondition{}
		}
		trigger.Filter = *filter
	}
	if taskTemplate != nil {
		trigger.TaskTemplate = *taskTemplate
	}
	if enabled != nil {
		trigger.Enabled = *enabled
	}

	if err := s.repo.Update(ctx, trigger); err != nil {
		return nil, err
	}
	return trigger, nil
}

func (s *triggerService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
