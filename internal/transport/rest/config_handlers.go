package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type ConfigHandlers struct {
	svc    service.ConfigService
	logger *slog.Logger
}

func NewConfigHandlers(svc service.ConfigService, logger *slog.Logger) *ConfigHandlers {
	return &ConfigHandlers{svc: svc, logger: logger}
}

type createConfigRequest struct {
	ModelProvider      string                   `json:"model_provider"`
	ModelID            string                   `json:"model_id"`
	SystemPrompt       string                   `json:"system_prompt"`
	ModelParams        domain.ModelParams       `json:"model_params"`
	APICredential      string                   `json:"api_credential,omitempty"`
	MCPServers         []domain.MCPServerConfig `json:"mcp_servers,omitempty"`
	ToolPermissions    domain.ToolPermissions   `json:"tool_permissions"`
	MaxIterations      int                      `json:"max_iterations"`
	TokenBudget        int                      `json:"token_budget"`
	TaskTimeoutSeconds int                      `json:"task_timeout_seconds"`
	EscalationRules    []domain.EscalationRule  `json:"escalation_rules,omitempty"`
}

func (h *ConfigHandlers) CreateConfig(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	var req createConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	config := &domain.AgentConfig{
		AgentID:            agentID,
		ModelProvider:      domain.ModelProvider(req.ModelProvider),
		ModelID:            req.ModelID,
		SystemPrompt:       req.SystemPrompt,
		ModelParams:        req.ModelParams,
		MCPServers:         req.MCPServers,
		ToolPermissions:    req.ToolPermissions,
		MaxIterations:      req.MaxIterations,
		TokenBudget:        req.TokenBudget,
		TaskTimeoutSeconds: req.TaskTimeoutSeconds,
		EscalationRules:    req.EscalationRules,
	}

	if err := h.svc.CreateConfig(r.Context(), config, req.APICredential); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, config)
}

func (h *ConfigHandlers) GetConfig(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	config, err := h.svc.GetConfig(r.Context(), agentID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, config)
}

type updateConfigRequest struct {
	ModelProvider      string                   `json:"model_provider"`
	ModelID            string                   `json:"model_id"`
	SystemPrompt       string                   `json:"system_prompt"`
	ModelParams        domain.ModelParams       `json:"model_params"`
	APICredential      string                   `json:"api_credential,omitempty"`
	MCPServers         []domain.MCPServerConfig `json:"mcp_servers,omitempty"`
	ToolPermissions    domain.ToolPermissions   `json:"tool_permissions"`
	MaxIterations      int                      `json:"max_iterations"`
	TokenBudget        int                      `json:"token_budget"`
	TaskTimeoutSeconds int                      `json:"task_timeout_seconds"`
	EscalationRules    []domain.EscalationRule  `json:"escalation_rules,omitempty"`
}

func (h *ConfigHandlers) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	var req updateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	config := &domain.AgentConfig{
		AgentID:            agentID,
		ModelProvider:      domain.ModelProvider(req.ModelProvider),
		ModelID:            req.ModelID,
		SystemPrompt:       req.SystemPrompt,
		ModelParams:        req.ModelParams,
		MCPServers:         req.MCPServers,
		ToolPermissions:    req.ToolPermissions,
		MaxIterations:      req.MaxIterations,
		TokenBudget:        req.TokenBudget,
		TaskTimeoutSeconds: req.TaskTimeoutSeconds,
		EscalationRules:    req.EscalationRules,
	}

	if err := h.svc.UpdateConfig(r.Context(), config, req.APICredential); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, config)
}

func (h *ConfigHandlers) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	if err := h.svc.DeleteConfig(r.Context(), agentID); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ConfigHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAgentNotFound):
		writeError(w, http.StatusNotFound, "agent not found")
	case errors.Is(err, domain.ErrConfigNotFound):
		writeError(w, http.StatusNotFound, "agent config not found")
	case errors.Is(err, domain.ErrConfigAlreadyExists):
		writeError(w, http.StatusConflict, "agent config already exists")
	case errors.Is(err, domain.ErrInvalidConfig):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
