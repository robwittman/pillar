package mock

import "context"

type AgentRuntime struct {
	EnsureRunningFn func(ctx context.Context, agentID string) error
	EnsureStoppedFn func(ctx context.Context, agentID string) error
	RemoveFn        func(ctx context.Context, agentID string) error
}

func (m *AgentRuntime) EnsureRunning(ctx context.Context, agentID string) error {
	if m.EnsureRunningFn != nil {
		return m.EnsureRunningFn(ctx, agentID)
	}
	return nil
}

func (m *AgentRuntime) EnsureStopped(ctx context.Context, agentID string) error {
	if m.EnsureStoppedFn != nil {
		return m.EnsureStoppedFn(ctx, agentID)
	}
	return nil
}

func (m *AgentRuntime) Remove(ctx context.Context, agentID string) error {
	if m.RemoveFn != nil {
		return m.RemoveFn(ctx, agentID)
	}
	return nil
}
