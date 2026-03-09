package rest

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

type AttributeHandlers struct {
	svc    service.AttributeService
	logger *slog.Logger
}

func NewAttributeHandlers(svc service.AttributeService, logger *slog.Logger) *AttributeHandlers {
	return &AttributeHandlers{svc: svc, logger: logger}
}

func (h *AttributeHandlers) SetAttribute(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	namespace := chi.URLParam(r, "namespace")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	if !json.Valid(body) {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	attr, err := h.svc.Set(r.Context(), agentID, namespace, json.RawMessage(body))
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, attr)
}

func (h *AttributeHandlers) ListAttributes(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	attrs, err := h.svc.List(r.Context(), agentID)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if attrs == nil {
		attrs = []*domain.AgentAttribute{}
	}
	writeJSON(w, http.StatusOK, attrs)
}

func (h *AttributeHandlers) GetAttribute(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	namespace := chi.URLParam(r, "namespace")
	attr, err := h.svc.Get(r.Context(), agentID, namespace)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, attr)
}

func (h *AttributeHandlers) DeleteAttribute(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "id")
	namespace := chi.URLParam(r, "namespace")
	if err := h.svc.Delete(r.Context(), agentID, namespace); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AttributeHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAgentNotFound):
		writeError(w, http.StatusNotFound, "agent not found")
	case errors.Is(err, domain.ErrAttributeNotFound):
		writeError(w, http.StatusNotFound, "attribute not found")
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
