package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/robwittman/pillar/internal/domain"
)

type AttributeService interface {
	Set(ctx context.Context, agentID, namespace string, value json.RawMessage) (*domain.AgentAttribute, error)
	Get(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error)
	List(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error)
	Delete(ctx context.Context, agentID, namespace string) error
}

type attributeService struct {
	repo      domain.AgentAttributeRepository
	agentRepo domain.AgentRepository
	logger    *slog.Logger
}

func NewAttributeService(repo domain.AgentAttributeRepository, agentRepo domain.AgentRepository, logger *slog.Logger) AttributeService {
	return &attributeService{
		repo:      repo,
		agentRepo: agentRepo,
		logger:    logger,
	}
}

func (s *attributeService) Set(ctx context.Context, agentID, namespace string, value json.RawMessage) (*domain.AgentAttribute, error) {
	if _, err := s.agentRepo.Get(ctx, agentID); err != nil {
		return nil, err
	}

	attr := &domain.AgentAttribute{
		AgentID:   agentID,
		Namespace: namespace,
		Value:     value,
	}

	if err := s.repo.Set(ctx, attr); err != nil {
		return nil, err
	}

	s.logger.Info("attribute set", "agent_id", agentID, "namespace", namespace)
	return attr, nil
}

func (s *attributeService) Get(ctx context.Context, agentID, namespace string) (*domain.AgentAttribute, error) {
	return s.repo.Get(ctx, agentID, namespace)
}

func (s *attributeService) List(ctx context.Context, agentID string) ([]*domain.AgentAttribute, error) {
	return s.repo.List(ctx, agentID)
}

func (s *attributeService) Delete(ctx context.Context, agentID, namespace string) error {
	return s.repo.Delete(ctx, agentID, namespace)
}
