package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type AgentService struct {
	CreateFn    func(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error)
	GetFn       func(ctx context.Context, id string) (*domain.Agent, error)
	ListFn      func(ctx context.Context) ([]*domain.Agent, error)
	UpdateFn    func(ctx context.Context, id string, name string, metadata, labels map[string]string) (*domain.Agent, error)
	DeleteFn    func(ctx context.Context, id string) error
	StartFn     func(ctx context.Context, id string) error
	StopFn      func(ctx context.Context, id string) error
	StatusFn    func(ctx context.Context, id string) (*service.AgentStatusInfo, error)
	HeartbeatFn func(ctx context.Context, agentID string) error
}

func (m *AgentService) Create(ctx context.Context, name string, metadata, labels map[string]string) (*domain.Agent, error) {
	return m.CreateFn(ctx, name, metadata, labels)
}

func (m *AgentService) Get(ctx context.Context, id string) (*domain.Agent, error) {
	return m.GetFn(ctx, id)
}

func (m *AgentService) List(ctx context.Context) ([]*domain.Agent, error) {
	return m.ListFn(ctx)
}

func (m *AgentService) Update(ctx context.Context, id string, name string, metadata, labels map[string]string) (*domain.Agent, error) {
	return m.UpdateFn(ctx, id, name, metadata, labels)
}

func (m *AgentService) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *AgentService) Start(ctx context.Context, id string) error {
	return m.StartFn(ctx, id)
}

func (m *AgentService) Stop(ctx context.Context, id string) error {
	return m.StopFn(ctx, id)
}

func (m *AgentService) Status(ctx context.Context, id string) (*service.AgentStatusInfo, error) {
	return m.StatusFn(ctx, id)
}

func (m *AgentService) Heartbeat(ctx context.Context, agentID string) error {
	return m.HeartbeatFn(ctx, agentID)
}
