package mock

import (
	"context"
	"time"
)

type AgentStatusStore struct {
	SetHeartbeatFn func(ctx context.Context, agentID string, ttl time.Duration) error
	IsOnlineFn     func(ctx context.Context, agentID string) (bool, error)
	SetOnlineFn    func(ctx context.Context, agentID string) error
	SetOfflineFn   func(ctx context.Context, agentID string) error
	ListOnlineFn   func(ctx context.Context) ([]string, error)
}

func (m *AgentStatusStore) SetHeartbeat(ctx context.Context, agentID string, ttl time.Duration) error {
	return m.SetHeartbeatFn(ctx, agentID, ttl)
}

func (m *AgentStatusStore) IsOnline(ctx context.Context, agentID string) (bool, error) {
	return m.IsOnlineFn(ctx, agentID)
}

func (m *AgentStatusStore) SetOnline(ctx context.Context, agentID string) error {
	return m.SetOnlineFn(ctx, agentID)
}

func (m *AgentStatusStore) SetOffline(ctx context.Context, agentID string) error {
	return m.SetOfflineFn(ctx, agentID)
}

func (m *AgentStatusStore) ListOnline(ctx context.Context) ([]string, error) {
	return m.ListOnlineFn(ctx)
}
