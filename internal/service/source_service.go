package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

type SourceService interface {
	Create(ctx context.Context, name string) (*domain.Source, string, error)
	Get(ctx context.Context, id string) (*domain.Source, error)
	List(ctx context.Context) ([]*domain.Source, error)
	Update(ctx context.Context, id string, name string) (*domain.Source, error)
	Delete(ctx context.Context, id string) error
	RotateSecret(ctx context.Context, id string) (*domain.Source, string, error)
	HandleEvent(ctx context.Context, sourceID string, signature string, payload json.RawMessage) ([]string, error)
}

type sourceService struct {
	sourceRepo  domain.SourceRepository
	triggerRepo domain.TriggerRepository
	taskSvc     TaskService
	logger      *slog.Logger
}

func NewSourceService(sourceRepo domain.SourceRepository, triggerRepo domain.TriggerRepository, taskSvc TaskService, logger *slog.Logger) SourceService {
	return &sourceService{
		sourceRepo:  sourceRepo,
		triggerRepo: triggerRepo,
		taskSvc:     taskSvc,
		logger:      logger,
	}
}

func (s *sourceService) Create(ctx context.Context, name string) (*domain.Source, string, error) {
	if name == "" {
		return nil, "", domain.ErrInvalidSource
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, "", err
	}

	source := &domain.Source{
		ID:     uuid.New().String(),
		Name:   name,
		Secret: secret,
	}

	if err := s.sourceRepo.Create(ctx, source); err != nil {
		return nil, "", err
	}

	s.logger.Info("source created", "id", source.ID, "name", name)
	return source, secret, nil
}

func (s *sourceService) Get(ctx context.Context, id string) (*domain.Source, error) {
	return s.sourceRepo.Get(ctx, id)
}

func (s *sourceService) List(ctx context.Context) ([]*domain.Source, error) {
	return s.sourceRepo.List(ctx)
}

func (s *sourceService) Update(ctx context.Context, id string, name string) (*domain.Source, error) {
	source, err := s.sourceRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		source.Name = name
	}

	if err := s.sourceRepo.Update(ctx, source); err != nil {
		return nil, err
	}
	return source, nil
}

func (s *sourceService) Delete(ctx context.Context, id string) error {
	return s.sourceRepo.Delete(ctx, id)
}

func (s *sourceService) RotateSecret(ctx context.Context, id string) (*domain.Source, string, error) {
	source, err := s.sourceRepo.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, "", err
	}

	source.Secret = secret
	if err := s.sourceRepo.Update(ctx, source); err != nil {
		return nil, "", err
	}

	s.logger.Info("source secret rotated", "id", source.ID)
	return source, secret, nil
}

func (s *sourceService) HandleEvent(ctx context.Context, sourceID string, signature string, payload json.RawMessage) ([]string, error) {
	source, err := s.sourceRepo.Get(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// Verify HMAC signature if provided
	if signature != "" {
		if !verifySignature(payload, source.Secret, signature) {
			return nil, fmt.Errorf("invalid signature")
		}
	}

	triggers, err := s.triggerRepo.ListBySource(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("fetching triggers: %w", err)
	}

	var taskIDs []string
	for _, trigger := range triggers {
		if !EvaluateFilter(trigger.Filter, payload) {
			continue
		}

		prompt, err := RenderTaskTemplate(trigger.TaskTemplate, payload)
		if err != nil {
			s.logger.Warn("failed to render task template",
				"trigger_id", trigger.ID, "error", err)
			continue
		}

		triggerID := trigger.ID
		task, err := s.taskSvc.Create(ctx, trigger.AgentID, prompt, payload, &triggerID)
		if err != nil {
			s.logger.Warn("failed to create task",
				"trigger_id", trigger.ID, "agent_id", trigger.AgentID, "error", err)
			continue
		}

		taskIDs = append(taskIDs, task.ID)
		s.logger.Info("task created from trigger",
			"task_id", task.ID, "trigger_id", trigger.ID, "agent_id", trigger.AgentID)
	}

	return taskIDs, nil
}

func verifySignature(payload []byte, secret, signature string) bool {
	// Support "sha256=<hex>" format (GitHub/Gitea style)
	sig := strings.TrimPrefix(signature, "sha256=")
	expectedMAC, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hmac.Equal(mac.Sum(nil), expectedMAC)
}
