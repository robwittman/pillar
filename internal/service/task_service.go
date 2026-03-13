package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

// TaskNotifier delivers tasks to connected agents via gRPC.
type TaskNotifier interface {
	SendTaskAssignment(agentID, taskID, prompt string) error
}

type TaskService interface {
	Create(ctx context.Context, agentID string, prompt string, taskContext json.RawMessage, triggerID *string) (*domain.Task, error)
	Get(ctx context.Context, id string) (*domain.Task, error)
	List(ctx context.Context) ([]*domain.Task, error)
	ListByAgent(ctx context.Context, agentID string) ([]*domain.Task, error)
	Complete(ctx context.Context, id string, result string, success bool) (*domain.Task, error)
	Cancel(ctx context.Context, id string) error
	DeliverPending(ctx context.Context, agentID string) error
}

type taskService struct {
	repo     domain.TaskRepository
	notifier TaskNotifier
	logger   *slog.Logger
}

func NewTaskService(repo domain.TaskRepository, notifier TaskNotifier, logger *slog.Logger) TaskService {
	return &taskService{
		repo:     repo,
		notifier: notifier,
		logger:   logger,
	}
}

func (s *taskService) Create(ctx context.Context, agentID string, prompt string, taskContext json.RawMessage, triggerID *string) (*domain.Task, error) {
	if agentID == "" || prompt == "" {
		return nil, domain.ErrInvalidTask
	}

	task := &domain.Task{
		ID:        uuid.New().String(),
		AgentID:   agentID,
		TriggerID: triggerID,
		Status:    domain.TaskStatusPending,
		Prompt:    prompt,
		Context:   taskContext,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	s.logger.Info("task created", "id", task.ID, "agent_id", agentID)

	// Try immediate delivery
	if s.notifier != nil {
		if err := s.notifier.SendTaskAssignment(agentID, task.ID, prompt); err != nil {
			s.logger.Debug("task queued (agent not connected)", "task_id", task.ID, "agent_id", agentID)
		} else {
			task.Status = domain.TaskStatusAssigned
			_ = s.repo.Update(ctx, task)
		}
	}

	return task, nil
}

func (s *taskService) Get(ctx context.Context, id string) (*domain.Task, error) {
	return s.repo.Get(ctx, id)
}

func (s *taskService) List(ctx context.Context) ([]*domain.Task, error) {
	return s.repo.List(ctx)
}

func (s *taskService) ListByAgent(ctx context.Context, agentID string) ([]*domain.Task, error) {
	return s.repo.ListByAgent(ctx, agentID)
}

func (s *taskService) Complete(ctx context.Context, id string, result string, success bool) (*domain.Task, error) {
	if err := s.repo.Complete(ctx, id, result, success); err != nil {
		return nil, err
	}

	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Info("task completed", "id", id, "success", success)
	return task, nil
}

func (s *taskService) Cancel(ctx context.Context, id string) error {
	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusFailed {
		return domain.ErrInvalidTask
	}

	task.Status = domain.TaskStatusFailed
	result := "cancelled"
	task.Result = &result
	return s.repo.Update(ctx, task)
}

func (s *taskService) DeliverPending(ctx context.Context, agentID string) error {
	tasks, err := s.repo.ListPendingByAgent(ctx, agentID)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if s.notifier == nil {
			break
		}
		if err := s.notifier.SendTaskAssignment(agentID, task.ID, task.Prompt); err != nil {
			s.logger.Warn("failed to deliver pending task", "task_id", task.ID, "error", err)
			break
		}
		task.Status = domain.TaskStatusAssigned
		if err := s.repo.Update(ctx, task); err != nil {
			s.logger.Warn("failed to update task status", "task_id", task.ID, "error", err)
		}
		s.logger.Info("delivered pending task", "task_id", task.ID, "agent_id", agentID)
	}

	return nil
}
