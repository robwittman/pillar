package domain

import (
	"context"
	"encoding/json"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID          string          `json:"id"`
	AgentID     string          `json:"agent_id"`
	TriggerID   *string         `json:"trigger_id,omitempty"`
	Status      TaskStatus      `json:"status"`
	Prompt      string          `json:"prompt"`
	Context     json.RawMessage `json:"context,omitempty"`
	Result      *string         `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	Get(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context) ([]*Task, error)
	ListByAgent(ctx context.Context, agentID string) ([]*Task, error)
	ListByStatus(ctx context.Context, status TaskStatus) ([]*Task, error)
	ListPendingByAgent(ctx context.Context, agentID string) ([]*Task, error)
	Update(ctx context.Context, task *Task) error
	Complete(ctx context.Context, id string, result string, success bool) error
}
