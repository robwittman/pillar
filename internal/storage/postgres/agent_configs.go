package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robwittman/pillar/internal/domain"
)

type AgentConfigRepository struct {
	pool *pgxpool.Pool
}

func NewAgentConfigRepository(pool *pgxpool.Pool) *AgentConfigRepository {
	return &AgentConfigRepository{pool: pool}
}

func (r *AgentConfigRepository) Create(ctx context.Context, config *domain.AgentConfig) error {
	modelParams, err := json.Marshal(config.ModelParams)
	if err != nil {
		return fmt.Errorf("marshaling model_params: %w", err)
	}
	mcpServers, err := json.Marshal(config.MCPServers)
	if err != nil {
		return fmt.Errorf("marshaling mcp_servers: %w", err)
	}
	toolPerms, err := json.Marshal(config.ToolPermissions)
	if err != nil {
		return fmt.Errorf("marshaling tool_permissions: %w", err)
	}
	escalationRules, err := json.Marshal(config.EscalationRules)
	if err != nil {
		return fmt.Errorf("marshaling escalation_rules: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	err = r.pool.QueryRow(ctx,
		`INSERT INTO agent_configs (agent_id, model_provider, model_id, system_prompt, model_params,
		 api_credential_ref, mcp_servers, tool_permissions, max_iterations, token_budget,
		 task_timeout_seconds, escalation_rules, org_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING created_at, updated_at`,
		config.AgentID, config.ModelProvider, config.ModelID, config.SystemPrompt, modelParams,
		config.APICredentialRef, mcpServers, toolPerms, config.MaxIterations, config.TokenBudget,
		config.TaskTimeoutSeconds, escalationRules, nullIfEmpty(orgID),
	).Scan(&config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConfigAlreadyExists
		}
		return fmt.Errorf("inserting agent config: %w", err)
	}
	return nil
}

func (r *AgentConfigRepository) Get(ctx context.Context, agentID string) (*domain.AgentConfig, error) {
	config := &domain.AgentConfig{}
	var modelParams, mcpServers, toolPerms, escalationRules []byte

	orgID := orgIDFromContext(ctx)
	query := `SELECT agent_id, model_provider, model_id, system_prompt, model_params,
		 api_credential_ref, mcp_servers, tool_permissions, max_iterations, token_budget,
		 task_timeout_seconds, escalation_rules, created_at, updated_at
		 FROM agent_configs WHERE agent_id = $1`
	args := []any{agentID}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&config.AgentID, &config.ModelProvider, &config.ModelID, &config.SystemPrompt, &modelParams,
		&config.APICredentialRef, &mcpServers, &toolPerms, &config.MaxIterations, &config.TokenBudget,
		&config.TaskTimeoutSeconds, &escalationRules, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, fmt.Errorf("querying agent config: %w", err)
	}

	if err := json.Unmarshal(modelParams, &config.ModelParams); err != nil {
		return nil, fmt.Errorf("unmarshaling model_params: %w", err)
	}
	if err := json.Unmarshal(mcpServers, &config.MCPServers); err != nil {
		return nil, fmt.Errorf("unmarshaling mcp_servers: %w", err)
	}
	if err := json.Unmarshal(toolPerms, &config.ToolPermissions); err != nil {
		return nil, fmt.Errorf("unmarshaling tool_permissions: %w", err)
	}
	if err := json.Unmarshal(escalationRules, &config.EscalationRules); err != nil {
		return nil, fmt.Errorf("unmarshaling escalation_rules: %w", err)
	}
	return config, nil
}

func (r *AgentConfigRepository) Update(ctx context.Context, config *domain.AgentConfig) error {
	modelParams, err := json.Marshal(config.ModelParams)
	if err != nil {
		return fmt.Errorf("marshaling model_params: %w", err)
	}
	mcpServers, err := json.Marshal(config.MCPServers)
	if err != nil {
		return fmt.Errorf("marshaling mcp_servers: %w", err)
	}
	toolPerms, err := json.Marshal(config.ToolPermissions)
	if err != nil {
		return fmt.Errorf("marshaling tool_permissions: %w", err)
	}
	escalationRules, err := json.Marshal(config.EscalationRules)
	if err != nil {
		return fmt.Errorf("marshaling escalation_rules: %w", err)
	}

	orgID := orgIDFromContext(ctx)
	query := `UPDATE agent_configs SET model_provider = $2, model_id = $3, system_prompt = $4,
		 model_params = $5, api_credential_ref = $6, mcp_servers = $7, tool_permissions = $8,
		 max_iterations = $9, token_budget = $10, task_timeout_seconds = $11, escalation_rules = $12
		 WHERE agent_id = $1`
	args := []any{config.AgentID, config.ModelProvider, config.ModelID, config.SystemPrompt, modelParams,
		config.APICredentialRef, mcpServers, toolPerms, config.MaxIterations, config.TokenBudget,
		config.TaskTimeoutSeconds, escalationRules}
	if orgID != "" {
		query += ` AND org_id = $13`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating agent config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConfigNotFound
	}
	return nil
}

func (r *AgentConfigRepository) Delete(ctx context.Context, agentID string) error {
	orgID := orgIDFromContext(ctx)
	query := `DELETE FROM agent_configs WHERE agent_id = $1`
	args := []any{agentID}
	if orgID != "" {
		query += ` AND org_id = $2`
		args = append(args, orgID)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("deleting agent config: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConfigNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
