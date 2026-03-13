package mock

import (
	"context"

	"github.com/robwittman/pillar/internal/domain"
)

type TaskRepository struct {
	CreateFn             func(ctx context.Context, task *domain.Task) error
	GetFn                func(ctx context.Context, id string) (*domain.Task, error)
	ListFn               func(ctx context.Context) ([]*domain.Task, error)
	ListByAgentFn        func(ctx context.Context, agentID string) ([]*domain.Task, error)
	ListByStatusFn       func(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error)
	ListPendingByAgentFn func(ctx context.Context, agentID string) ([]*domain.Task, error)
	UpdateFn             func(ctx context.Context, task *domain.Task) error
	CompleteFn           func(ctx context.Context, id string, result string, success bool) error
}

func (m *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	return m.CreateFn(ctx, task)
}

func (m *TaskRepository) Get(ctx context.Context, id string) (*domain.Task, error) {
	return m.GetFn(ctx, id)
}

func (m *TaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	return m.ListFn(ctx)
}

func (m *TaskRepository) ListByAgent(ctx context.Context, agentID string) ([]*domain.Task, error) {
	return m.ListByAgentFn(ctx, agentID)
}

func (m *TaskRepository) ListByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	return m.ListByStatusFn(ctx, status)
}

func (m *TaskRepository) ListPendingByAgent(ctx context.Context, agentID string) ([]*domain.Task, error) {
	return m.ListPendingByAgentFn(ctx, agentID)
}

func (m *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	return m.UpdateFn(ctx, task)
}

func (m *TaskRepository) Complete(ctx context.Context, id string, result string, success bool) error {
	return m.CompleteFn(ctx, id, result, success)
}
