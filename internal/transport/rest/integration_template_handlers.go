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

type IntegrationTemplateHandlers struct {
	svc    service.IntegrationTemplateService
	logger *slog.Logger
}

func NewIntegrationTemplateHandlers(svc service.IntegrationTemplateService, logger *slog.Logger) *IntegrationTemplateHandlers {
	return &IntegrationTemplateHandlers{svc: svc, logger: logger}
}

type createIntegrationTemplateRequest struct {
	Type     string            `json:"type"`
	Name     string            `json:"name"`
	Config   map[string]any    `json:"config"`
	Selector map[string]string `json:"selector"`
}

func (h *IntegrationTemplateHandlers) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req createIntegrationTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	template := &domain.IntegrationTemplate{
		Type:     req.Type,
		Name:     req.Name,
		Config:   req.Config,
		Selector: req.Selector,
	}

	if err := h.svc.Create(r.Context(), template); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, template)
}

func (h *IntegrationTemplateHandlers) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	if templates == nil {
		templates = []*domain.IntegrationTemplate{}
	}
	writeJSON(w, http.StatusOK, templates)
}

func (h *IntegrationTemplateHandlers) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "templateID")
	template, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, template)
}

type updateIntegrationTemplateRequest struct {
	Name     string            `json:"name"`
	Config   map[string]any    `json:"config"`
	Selector map[string]string `json:"selector"`
}

func (h *IntegrationTemplateHandlers) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "templateID")
	var req updateIntegrationTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	template := &domain.IntegrationTemplate{
		ID:       id,
		Name:     req.Name,
		Config:   req.Config,
		Selector: req.Selector,
	}

	if err := h.svc.Update(r.Context(), template); err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, template)
}

func (h *IntegrationTemplateHandlers) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "templateID")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *IntegrationTemplateHandlers) PreviewTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "templateID")
	agents, err := h.svc.Preview(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if agents == nil {
		agents = []*domain.Agent{}
	}
	writeJSON(w, http.StatusOK, agents)
}

func (h *IntegrationTemplateHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrIntegrationTemplateNotFound):
		writeError(w, http.StatusNotFound, "integration template not found")
	case errors.Is(err, domain.ErrInvalidIntegrationTemplate):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
