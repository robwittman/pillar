// Package agent provides a reusable LLM agent loop that uses Pillar config
// and attributes to interact with Claude and execute tools.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
)

// Emitter receives structured LLM events.
type Emitter interface {
	Emit(Event)
}

// Runner executes an agentic loop using Claude with tool use.
type Runner struct {
	client     anthropic.Client
	config     *pillarv1.AgentConfig
	attributes map[string][]byte
	logger     *slog.Logger
	events     Emitter
	tools      []anthropic.ToolUnionParam
	handlers   map[string]ToolHandler
	httpClient *http.Client
}

// ToolHandler processes a tool call and returns the result string.
type ToolHandler func(ctx context.Context, input json.RawMessage) (string, error)

// RunnerOption configures a Runner.
type RunnerOption func(*Runner)

// WithEvents sets the event emitter for structured LLM event logging.
func WithEvents(e Emitter) RunnerOption {
	return func(r *Runner) { r.events = e }
}

// NewRunner creates a runner from Pillar agent config and attributes.
func NewRunner(cfg *pillarv1.AgentConfig, attributes map[string][]byte, logger *slog.Logger, opts ...RunnerOption) (*Runner, error) {
	if cfg.ApiCredential == "" {
		return nil, fmt.Errorf("api_credential is required in agent config")
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.ApiCredential))

	r := &Runner{
		client:     client,
		config:     cfg,
		attributes: attributes,
		logger:     logger,
		events:     NopEventWriter{},
		handlers:   make(map[string]ToolHandler),
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(r)
	}

	r.registerBuiltinTools()

	return r, nil
}

func (r *Runner) registerBuiltinTools() {
	r.RegisterTool(anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        "http_request",
			Description: anthropic.String("Make an HTTP request. Use this to interact with REST APIs."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"method": map[string]interface{}{
						"type":        "string",
						"description": "HTTP method (GET, POST, PUT, DELETE, PATCH)",
					},
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Full URL to request",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": "HTTP headers as key-value pairs",
						"additionalProperties": map[string]interface{}{
							"type": "string",
						},
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Request body (for POST/PUT/PATCH)",
					},
				},
				Required: []string{"method", "url"},
			},
		},
	}, r.handleHTTPRequest)

	r.RegisterTool(anthropic.ToolUnionParam{
		OfTool: &anthropic.ToolParam{
			Name:        "get_attribute",
			Description: anthropic.String("Read an agent attribute by namespace. Attributes contain credentials and configuration from external systems."),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "The attribute namespace (e.g., 'keycloak', 'redmine')",
					},
				},
				Required: []string{"namespace"},
			},
		},
	}, r.handleGetAttribute)
}

// RegisterTool adds a custom tool to the runner.
func (r *Runner) RegisterTool(tool anthropic.ToolUnionParam, handler ToolHandler) {
	r.tools = append(r.tools, tool)
	if tool.OfTool != nil {
		r.handlers[tool.OfTool.Name] = handler
	}
}

// Run executes the agent loop with the given task prompt.
// It runs until the model stops using tools, hits max iterations, or context is cancelled.
func (r *Runner) Run(ctx context.Context, task string) (string, error) {
	maxIter := int(r.config.MaxIterations)
	if maxIter <= 0 {
		maxIter = 20
	}

	model := anthropic.Model(r.config.ModelId)
	if r.config.ModelId == "" {
		model = anthropic.ModelClaudeSonnet4_6
	}

	systemPrompt := r.config.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful agent."
	}

	if len(r.attributes) > 0 {
		namespaces := make([]string, 0, len(r.attributes))
		for ns := range r.attributes {
			namespaces = append(namespaces, ns)
		}
		systemPrompt += fmt.Sprintf("\n\nYou have attributes available in the following namespaces: %s. Use the get_attribute tool to read them.", strings.Join(namespaces, ", "))
	}

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(task)),
	}

	var finalText string

	r.events.Emit(Event{
		Type:       EventLoopStart,
		Model:      string(model),
		Iterations: maxIter,
	})

	for i := range maxIter {
		r.logger.Info("agent iteration", "iteration", i+1, "max", maxIter)

		params := anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: 4096,
			System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
			Messages:  messages,
		}
		if len(r.tools) > 0 {
			params.Tools = r.tools
		}
		if r.config.ModelParams != nil && r.config.ModelParams.MaxTokens > 0 {
			params.MaxTokens = int64(r.config.ModelParams.MaxTokens)
		}

		r.events.Emit(Event{
			Type:         EventRequestStart,
			Iteration:    i + 1,
			Model:        string(model),
			MessageCount: len(messages),
			ToolCount:    len(r.tools),
		})

		resp, err := r.client.Messages.New(ctx, params)
		if err != nil {
			r.events.Emit(Event{Type: EventError, Iteration: i + 1, Error: err.Error()})
			return "", fmt.Errorf("claude API error on iteration %d: %w", i+1, err)
		}

		r.events.Emit(Event{
			Type:         EventRequestEnd,
			Iteration:    i + 1,
			StopReason:   string(resp.StopReason),
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		})

		var assistantBlocks []anthropic.ContentBlockParamUnion
		var toolUses []struct {
			id    string
			name  string
			input json.RawMessage
		}

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				finalText = block.Text
				assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(block.Text))
				r.logger.Info("assistant text", "text", block.Text)
				r.events.Emit(Event{Type: EventTextDelta, Iteration: i + 1, Text: block.Text})
			case "tool_use":
				assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(block.ID, block.Input, block.Name))
				toolUses = append(toolUses, struct {
					id    string
					name  string
					input json.RawMessage
				}{block.ID, block.Name, block.Input})
				r.logger.Info("tool use", "tool", block.Name, "id", block.ID, "input", string(block.Input))
				r.events.Emit(Event{Type: EventToolUse, Iteration: i + 1, ToolName: block.Name, ToolID: block.ID, ToolInput: block.Input})
			}
		}

		messages = append(messages, anthropic.NewAssistantMessage(assistantBlocks...))

		if resp.StopReason == "end_turn" || len(toolUses) == 0 {
			r.logger.Info("agent loop complete", "reason", resp.StopReason, "iterations", i+1)
			r.events.Emit(Event{Type: EventLoopEnd, Iterations: i + 1, StopReason: string(resp.StopReason), FinalText: finalText})
			return finalText, nil
		}

		var toolResults []anthropic.ContentBlockParamUnion
		for _, tu := range toolUses {
			result, err := r.executeTool(ctx, tu.name, tu.input)
			if err != nil {
				r.logger.Warn("tool execution error", "tool", tu.name, "error", err)
				toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.id, fmt.Sprintf("Error: %s", err), true))
				r.events.Emit(Event{Type: EventToolResult, Iteration: i + 1, ToolName: tu.name, ToolID: tu.id, IsError: true, Error: err.Error()})
			} else {
				r.logger.Debug("tool result", "tool", tu.name, "result", result)
				r.logger.Info("tool result", "tool", tu.name, "result_len", len(result), "result_preview", truncate(result, 200))
				toolResults = append(toolResults, anthropic.NewToolResultBlock(tu.id, result, false))
				r.events.Emit(Event{Type: EventToolResult, Iteration: i + 1, ToolName: tu.name, ToolID: tu.id, Result: result})
			}
		}

		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}

	r.events.Emit(Event{Type: EventLoopEnd, Iterations: maxIter, StopReason: "max_iterations", FinalText: finalText})
	return finalText, fmt.Errorf("max iterations (%d) reached", maxIter)
}

func (r *Runner) executeTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return handler(ctx, input)
}

func (r *Runner) handleHTTPRequest(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	var bodyReader io.Reader
	if params.Body != "" {
		bodyReader = strings.NewReader(params.Body)
	}

	req, err := http.NewRequestWithContext(ctx, params.Method, params.URL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	for k, v := range params.Headers {
		req.Header.Set(k, v)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 50_000))
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return fmt.Sprintf("HTTP %d\n\n%s", resp.StatusCode, string(body)), nil
}

func (r *Runner) handleGetAttribute(_ context.Context, input json.RawMessage) (string, error) {
	var params struct {
		Namespace string `json:"namespace"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	data, ok := r.attributes[params.Namespace]
	if !ok {
		return fmt.Sprintf("attribute namespace %q not found", params.Namespace), nil
	}

	return string(data), nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
