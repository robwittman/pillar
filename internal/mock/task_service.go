package mock

import (
	"context"
	"encoding/json"

	"github.com/robwittman/pillar/internal/domain"
)

type TaskService struct {
	CreateFn         func(ctx context.Context, agentID string, prompt string, taskContext json.RawMessage, triggerID *string) (*domain.Task, error)
	GetFn            func(ctx context.Context, id string) (*domain.Task, error)
	ListFn           func(ctx context.Context) ([]*domain.Task, error)
	ListByAgentFn    func(ctx context.Context, agentID string) ([]*domain.Task, error)
	CompleteFn       func(ctx context.Context, id string, result string, success bool) (*domain.Task, error)
	CancelFn         func(ctx context.Context, id string) error
	DeliverPendingFn func(ctx context.Context, agentID string) error
}

func (m *TaskService) Create(ctx context.Context, agentID string, prompt string, taskContext json.RawMessage, triggerID *string) (*domain.Task, error) {
	return m.CreateFn(ctx, agentID, prompt, taskContext, triggerID)
}

func (m *TaskService) Get(ctx context.Context, id string) (*domain.Task, error) {
	return m.GetFn(ctx, id)
}

func (m *TaskService) List(ctx context.Context) ([]*domain.Task, error) {
	return m.ListFn(ctx)
}

func (m *TaskService) ListByAgent(ctx context.Context, agentID string) ([]*domain.Task, error) {
	return m.ListByAgentFn(ctx, agentID)
}

func (m *TaskService) Complete(ctx context.Context, id string, result string, success bool) (*domain.Task, error) {
	return m.CompleteFn(ctx, id, result, success)
}

func (m *TaskService) Cancel(ctx context.Context, id string) error {
	return m.CancelFn(ctx, id)
}

func (m *TaskService) DeliverPending(ctx context.Context, agentID string) error {
	return m.DeliverPendingFn(ctx, agentID)
}
