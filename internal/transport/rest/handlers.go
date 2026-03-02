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

type Handlers struct {
	svc    service.AgentService
	logger *slog.Logger
}

func NewHandlers(svc service.AgentService, logger *slog.Logger) *Handlers {
	return &Handlers{svc: svc, logger: logger}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createAgentRequest struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

func (h *Handlers) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req createAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	agent, err := h.svc.Create(r.Context(), req.Name, req.Metadata, req.Labels)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, agent)
}

func (h *Handlers) GetAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	agent, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (h *Handlers) ListAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, agents)
}

type updateAgentRequest struct {
	Name     string            `json:"name"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

func (h *Handlers) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	agent, err := h.svc.Update(r.Context(), id, req.Name, req.Metadata, req.Labels)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, agent)
}

func (h *Handlers) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) StartAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Start(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *Handlers) StopAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Stop(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

func (h *Handlers) AgentStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	status, err := h.svc.Status(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *Handlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAgentNotFound):
		writeError(w, http.StatusNotFound, "agent not found")
	case errors.Is(err, domain.ErrInvalidTransition):
		writeError(w, http.StatusConflict, "invalid status transition")
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
