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

type TriggerHandlers struct {
	svc    service.TriggerService
	logger *slog.Logger
}

func NewTriggerHandlers(svc service.TriggerService, logger *slog.Logger) *TriggerHandlers {
	return &TriggerHandlers{svc: svc, logger: logger}
}

type createTriggerRequest struct {
	SourceID     string               `json:"source_id"`
	AgentID      string               `json:"agent_id"`
	Name         string               `json:"name"`
	Filter       domain.TriggerFilter `json:"filter"`
	TaskTemplate string               `json:"task_template"`
}

func (h *TriggerHandlers) CreateTrigger(w http.ResponseWriter, r *http.Request) {
	var req createTriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	trigger, err := h.svc.Create(r.Context(), req.SourceID, req.AgentID, req.Name, req.Filter, req.TaskTemplate)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, trigger)
}

func (h *TriggerHandlers) ListTriggers(w http.ResponseWriter, r *http.Request) {
	sourceID := r.URL.Query().Get("source_id")

	var (
		triggers []*domain.Trigger
		err      error
	)
	if sourceID != "" {
		triggers, err = h.svc.ListBySource(r.Context(), sourceID)
	} else {
		triggers, err = h.svc.List(r.Context())
	}
	if err != nil {
		h.handleError(w, err)
		return
	}
	if triggers == nil {
		triggers = []*domain.Trigger{}
	}
	writeJSON(w, http.StatusOK, triggers)
}

func (h *TriggerHandlers) GetTrigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "triggerID")
	trigger, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, trigger)
}

type updateTriggerRequest struct {
	Name         string                `json:"name,omitempty"`
	Filter       *domain.TriggerFilter `json:"filter,omitempty"`
	TaskTemplate *string               `json:"task_template,omitempty"`
	Enabled      *bool                 `json:"enabled,omitempty"`
}

func (h *TriggerHandlers) UpdateTrigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "triggerID")
	var req updateTriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	trigger, err := h.svc.Update(r.Context(), id, req.Name, req.Filter, req.TaskTemplate, req.Enabled)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, trigger)
}

func (h *TriggerHandlers) DeleteTrigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "triggerID")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TriggerHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrTriggerNotFound):
		writeError(w, http.StatusNotFound, "trigger not found")
	case errors.Is(err, domain.ErrInvalidTrigger):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
