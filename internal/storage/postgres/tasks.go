package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	orgID := orgIDFromContext(ctx)
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (id, agent_id, trigger_id, status, prompt, context, org_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING created_at, updated_at`,
		task.ID, task.AgentID, task.TriggerID, task.Status, task.Prompt, task.Context, nullIfEmpty(orgID),
	).Scan(&task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Get(ctx context.Context, id string) (*domain.Task, error) {
	t := &domain.Task{}
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, agent_id, trigger_id, status, prompt, context, result, created_at, updated_at, completed_at
		 FROM tasks WHERE id = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.AgentID, &t.TriggerID, &t.Status, &t.Prompt, &t.Context,
		&t.Result, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("querying task: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) List(ctx context.Context) ([]*domain.Task, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, agent_id, trigger_id, status, prompt, context, result, created_at, updated_at, completed_at FROM tasks`
	var args []any
	if orgID != "" {
		query += ` WHERE org_id = $1`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) ListByAgent(ctx context.Context, agentID string) ([]*domain.Task, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, agent_id, trigger_id, status, prompt, context, result, created_at, updated_at, completed_at
		 FROM tasks WHERE agent_id = $1`
	args := []any{agentID}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks by agent: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) ListByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	orgID := orgIDFromContext(ctx)
	query := `SELECT id, agent_id, trigger_id, status, prompt, context, result, created_at, updated_at, completed_at
		 FROM tasks WHERE status = $1`
	args := []any{status}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks by status: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) ListPendingByAgent(ctx context.Context, agentID string) ([]*domain.Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, agent_id, trigger_id, status, prompt, context, result, created_at, updated_at, completed_at
		 FROM tasks WHERE agent_id = $1 AND status = 'pending' ORDER BY created_at ASC`, agentID)
	if err != nil {
		return nil, fmt.Errorf("querying pending tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE tasks SET status = $2, result = $3, completed_at = $4 WHERE id = $1`,
		task.ID, task.Status, task.Result, task.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepository) Complete(ctx context.Context, id string, result string, success bool) error {
	status := domain.TaskStatusCompleted
	if !success {
		status = domain.TaskStatusFailed
	}
	now := time.Now()

	tag, err := r.pool.Exec(ctx,
		`UPDATE tasks SET status = $2, result = $3, completed_at = $4 WHERE id = $1`,
		id, status, result, now,
	)
	if err != nil {
		return fmt.Errorf("completing task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func scanTasks(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*domain.Task, error) {
	var tasks []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := rows.Scan(&t.ID, &t.AgentID, &t.TriggerID, &t.Status, &t.Prompt, &t.Context, &t.Result, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
