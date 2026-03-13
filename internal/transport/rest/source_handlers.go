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

type SourceHandlers struct {
	svc    service.SourceService
	logger *slog.Logger
}

func NewSourceHandlers(svc service.SourceService, logger *slog.Logger) *SourceHandlers {
	return &SourceHandlers{svc: svc, logger: logger}
}

type createSourceRequest struct {
	Name string `json:"name"`
}

type createSourceResponse struct {
	*domain.Source
	Secret string `json:"secret"`
}

func (h *SourceHandlers) CreateSource(w http.ResponseWriter, r *http.Request) {
	var req createSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	source, secret, err := h.svc.Create(r.Context(), req.Name)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createSourceResponse{Source: source, Secret: secret})
}

func (h *SourceHandlers) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	if sources == nil {
		sources = []*domain.Source{}
	}
	writeJSON(w, http.StatusOK, sources)
}

func (h *SourceHandlers) GetSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "sourceID")
	source, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, source)
}

type updateSourceRequest struct {
	Name string `json:"name"`
}

func (h *SourceHandlers) UpdateSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "sourceID")
	var req updateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	source, err := h.svc.Update(r.Context(), id, req.Name)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func (h *SourceHandlers) DeleteSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "sourceID")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type rotateSourceSecretResponse struct {
	*domain.Source
	Secret string `json:"secret"`
}

func (h *SourceHandlers) RotateSourceSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "sourceID")
	source, secret, err := h.svc.RotateSecret(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rotateSourceSecretResponse{Source: source, Secret: secret})
}

// HandleSourceEvent is the inbound webhook endpoint. External systems POST events here.
func (h *SourceHandlers) HandleSourceEvent(w http.ResponseWriter, r *http.Request) {
	sourceID := chi.URLParam(r, "sourceID")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// Accept GitHub/Gitea-style signature headers
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		signature = r.Header.Get("X-Signature-256")
	}

	taskIDs, err := h.svc.HandleEvent(r.Context(), sourceID, signature, body)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tasks_created": len(taskIDs),
		"task_ids":      taskIDs,
	})
}

func (h *SourceHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrSourceNotFound):
		writeError(w, http.StatusNotFound, "source not found")
	case errors.Is(err, domain.ErrInvalidSource):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		if err.Error() == "invalid signature" {
			writeError(w, http.StatusUnauthorized, "invalid signature")
			return
		}
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
