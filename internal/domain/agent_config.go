package domain

import (
	"context"
	"time"
)

type ModelProvider string

const (
	ModelProviderClaude ModelProvider = "claude"
	ModelProviderOpenAI ModelProvider = "openai"
)

type MCPTransportType string

const (
	MCPTransportStdio MCPTransportType = "stdio"
	MCPTransportSSE   MCPTransportType = "sse"
)

type EscalationAction string

const (
	EscalationActionPause  EscalationAction = "pause"
	EscalationActionNotify EscalationAction = "notify"
	EscalationActionStop   EscalationAction = "stop"
)

type AgentConfig struct {
	AgentID            string            `json:"agent_id"`
	ModelProvider      ModelProvider     `json:"model_provider"`
	ModelID            string            `json:"model_id"`
	SystemPrompt       string            `json:"system_prompt"`
	ModelParams        ModelParams       `json:"model_params"`
	APICredentialRef   string            `json:"api_credential_ref"`
	MCPServers         []MCPServerConfig `json:"mcp_servers"`
	ToolPermissions    ToolPermissions   `json:"tool_permissions"`
	MaxIterations      int               `json:"max_iterations"`
	TokenBudget        int               `json:"token_budget"`
	TaskTimeoutSeconds int               `json:"task_timeout_seconds"`
	EscalationRules    []EscalationRule  `json:"escalation_rules"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type ModelParams struct {
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	MaxTokens   int     `json:"max_tokens"`
}

type MCPServerConfig struct {
	Name          string            `json:"name"`
	TransportType MCPTransportType  `json:"transport_type"`
	Command       string            `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	URL           string            `json:"url,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

type ToolPermissions struct {
	AllowedTools []string `json:"allowed_tools,omitempty"`
	DeniedTools  []string `json:"denied_tools,omitempty"`
}

type EscalationRule struct {
	Name      string           `json:"name"`
	Condition string           `json:"condition"`
	Action    EscalationAction `json:"action"`
	Message   string           `json:"message,omitempty"`
}

type AgentConfigRepository interface {
	Create(ctx context.Context, config *AgentConfig) error
	Get(ctx context.Context, agentID string) (*AgentConfig, error)
	Update(ctx context.Context, config *AgentConfig) error
	Delete(ctx context.Context, agentID string) error
}

type SecretProvider interface {
	Put(ctx context.Context, name string, value string) error
	Get(ctx context.Context, name string) (string, error)
	Delete(ctx context.Context, name string) error
}
