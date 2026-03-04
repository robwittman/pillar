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

type IntegrationHandlers struct {
	svc    service.IntegrationService
	logger *slog.Logger
}

func NewIntegrationHandlers(svc service.IntegrationService, logger *slog.Logger) *IntegrationHandlers {
	return &IntegrationHandlers{svc: svc, logger: logger}
}

type createIntegrationRequest struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Config map[string]any `json:"config"`
}

func (h *IntegrationHandlers) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	var req createIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	integration := &domain.Integration{
		AgentID: agentID,
		Type:    req.Type,
		Name:    req.Name,
		Config:  req.Config,
	}

	if err := h.svc.Create(r.Context(), integration); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, integration)
}

func (h *IntegrationHandlers) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	integrations, err := h.svc.List(r.Context(), agentID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if integrations == nil {
		integrations = []*domain.Integration{}
	}
	writeJSON(w, http.StatusOK, integrations)
}

func (h *IntegrationHandlers) GetIntegration(w http.ResponseWriter, r *http.Request) {
	integID := chi.URLParam(r, "integID")
	integration, err := h.svc.Get(r.Context(), integID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, integration)
}

type updateIntegrationRequest struct {
	Name   string         `json:"name"`
	Config map[string]any `json:"config"`
}

func (h *IntegrationHandlers) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	integID := chi.URLParam(r, "integID")
	var req updateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	integration := &domain.Integration{
		ID:     integID,
		Name:   req.Name,
		Config: req.Config,
	}

	if err := h.svc.Update(r.Context(), integration); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, integration)
}

func (h *IntegrationHandlers) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	integID := chi.URLParam(r, "integID")
	if err := h.svc.Delete(r.Context(), integID); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *IntegrationHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAgentNotFound):
		writeError(w, http.StatusNotFound, "agent not found")
	case errors.Is(err, domain.ErrIntegrationNotFound):
		writeError(w, http.StatusNotFound, "integration not found")
	case errors.Is(err, domain.ErrInvalidIntegration):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
