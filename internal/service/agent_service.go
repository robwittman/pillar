package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

const defaultHeartbeatTTL = 30 * time.Second

const (
	DirectiveStart = "start"
	DirectiveStop  = "stop"
)

// AgentNotifier sends directives to connected agents.
type AgentNotifier interface {
	NotifyDirective(agentID string, directiveType string, payload string) error
}

// AgentRuntime manages the lifecycle of agent processes.
type AgentRuntime interface {
	EnsureRunning(ctx context.Context, agentID string) error
	EnsureStopped(ctx context.Context, agentID string) error
	Remove(ctx context.Context, agentID string) error
}

// EventEmitter emits lifecycle events for webhook delivery.
type EventEmitter interface {
	Emit(ctx context.Context, event domain.Event)
}

// AgentService defines the operations available on agents.
type AgentService interface {
	Create(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error)
	Get(ctx context.Context, id string) (*domain.Agent, error)
	List(ctx context.Context) ([]*domain.Agent, error)
	Update(ctx context.Context, id string, name string, metadata, labels map[string]string) (*domain.Agent, error)
	Delete(ctx context.Context, id string) error
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Status(ctx context.Context, id string) (*AgentStatusInfo, error)
	Heartbeat(ctx context.Context, agentID string) error
}

// AgentServiceOption configures an agentService.
type AgentServiceOption func(*agentService)

// WithNotifier sets the notifier used to push directives to connected agents.
func WithNotifier(n AgentNotifier) AgentServiceOption {
	return func(s *agentService) {
		s.notifier = n
	}
}

// WithRuntime sets the runtime used to manage agent processes.
func WithRuntime(r AgentRuntime) AgentServiceOption {
	return func(s *agentService) {
		s.runtime = r
	}
}

// WithEventEmitter sets the event emitter for lifecycle events.
func WithEventEmitter(e EventEmitter) AgentServiceOption {
	return func(s *agentService) {
		s.emitter = e
	}
}

type agentService struct {
	repo     domain.AgentRepository
	status   domain.AgentStatusStore
	logger   *slog.Logger
	notifier AgentNotifier
	runtime  AgentRuntime
	emitter  EventEmitter
}

func NewAgentService(repo domain.AgentRepository, status domain.AgentStatusStore, logger *slog.Logger, opts ...AgentServiceOption) AgentService {
	s := &agentService{
		repo:   repo,
		status: status,
		logger: logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *agentService) Create(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error) {
	agent := &domain.Agent{
		ID:       uuid.New().String(),
		Name:     name,
		Status:   domain.AgentStatusPending,
		Metadata: metadata,
		Labels:   labels,
	}
	if agent.Metadata == nil {
		agent.Metadata = make(map[string]string)
	}
	if agent.Labels == nil {
		agent.Labels = make(map[string]string)
	}

	if err := s.repo.Create(ctx, agent); err != nil {
		return nil, err
	}

	s.emitEvent(ctx, "agent.created", agent)
	s.logger.Info("agent created", "id", agent.ID, "name", agent.Name)
	return agent, nil
}

func (s *agentService) Get(ctx context.Context, id string) (*domain.Agent, error) {
	return s.repo.Get(ctx, id)
}

func (s *agentService) List(ctx context.Context) ([]*domain.Agent, error) {
	return s.repo.List(ctx)
}

func (s *agentService) Update(ctx context.Context, id string, name string, metadata, labels map[string]string) (*domain.Agent, error) {
	agent, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	agent.Name = name
	if metadata != nil {
		agent.Metadata = metadata
	}
	if labels != nil {
		agent.Labels = labels
	}

	if err := s.repo.Update(ctx, agent); err != nil {
		return nil, err
	}
	s.emitEvent(ctx, "agent.updated", agent)
	return agent, nil
}

func (s *agentService) Delete(ctx context.Context, id string) error {
	if err := s.status.SetOffline(ctx, id); err != nil {
		s.logger.Warn("failed to set agent offline in redis", "id", id, "error", err)
	}
	if s.runtime != nil {
		if err := s.runtime.Remove(ctx, id); err != nil {
			s.logger.Warn("runtime failed to remove agent", "id", id, "error", err)
		}
	}
	s.emitEvent(ctx, "agent.deleted", map[string]string{"id": id})
	return s.repo.Delete(ctx, id)
}

func (s *agentService) Start(ctx context.Context, id string) error {
	agent, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if agent.Status != domain.AgentStatusPending && agent.Status != domain.AgentStatusStopped {
		return domain.ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, id, domain.AgentStatusRunning); err != nil {
		return err
	}

	if s.runtime != nil {
		if err := s.runtime.EnsureRunning(ctx, id); err != nil {
			s.logger.Warn("runtime failed to ensure agent running", "id", id, "error", err)
		}
	}

	if s.notifier != nil {
		if err := s.notifier.NotifyDirective(id, DirectiveStart, ""); err != nil {
			s.logger.Warn("failed to notify agent of start", "id", id, "error", err)
		}
	}

	s.emitEvent(ctx, "agent.started", map[string]string{"id": id})
	s.logger.Info("agent started", "id", id)
	return nil
}

func (s *agentService) Stop(ctx context.Context, id string) error {
	agent, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if agent.Status != domain.AgentStatusRunning {
		return domain.ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, id, domain.AgentStatusStopped); err != nil {
		return err
	}

	if err := s.status.SetOffline(ctx, id); err != nil {
		s.logger.Warn("failed to set agent offline", "id", id, "error", err)
	}

	if s.runtime != nil {
		if err := s.runtime.EnsureStopped(ctx, id); err != nil {
			s.logger.Warn("runtime failed to ensure agent stopped", "id", id, "error", err)
		}
	}

	if s.notifier != nil {
		if err := s.notifier.NotifyDirective(id, DirectiveStop, ""); err != nil {
			s.logger.Warn("failed to notify agent of stop", "id", id, "error", err)
		}
	}

	s.emitEvent(ctx, "agent.stopped", map[string]string{"id": id})
	s.logger.Info("agent stopped", "id", id)
	return nil
}

func (s *agentService) emitEvent(ctx context.Context, eventType string, data any) {
	if s.emitter == nil {
		return
	}
	s.emitter.Emit(ctx, domain.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	})
}

type AgentStatusInfo struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Online  bool   `json:"online"`
}

func (s *agentService) Status(ctx context.Context, id string) (*AgentStatusInfo, error) {
	agent, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	online, err := s.status.IsOnline(ctx, id)
	if err != nil {
		s.logger.Warn("failed to check online status", "id", id, "error", err)
	}

	return &AgentStatusInfo{
		AgentID: agent.ID,
		Status:  string(agent.Status),
		Online:  online,
	}, nil
}

func (s *agentService) Heartbeat(ctx context.Context, agentID string) error {
	if err := s.status.SetHeartbeat(ctx, agentID, defaultHeartbeatTTL); err != nil {
		return err
	}
	return s.status.SetOnline(ctx, agentID)
}
