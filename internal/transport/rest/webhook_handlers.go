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

type WebhookHandlers struct {
	svc    service.WebhookService
	logger *slog.Logger
}

func NewWebhookHandlers(svc service.WebhookService, logger *slog.Logger) *WebhookHandlers {
	return &WebhookHandlers{svc: svc, logger: logger}
}

type createWebhookRequest struct {
	URL         string   `json:"url"`
	Description string   `json:"description"`
	EventTypes  []string `json:"event_types"`
}

type createWebhookResponse struct {
	*domain.Webhook
	Secret string `json:"secret"`
}

func (h *WebhookHandlers) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var req createWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	webhook, secret, err := h.svc.Create(r.Context(), req.URL, req.Description, req.EventTypes)
	if err != nil {
		h.handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createWebhookResponse{Webhook: webhook, Secret: secret})
}

func (h *WebhookHandlers) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, err)
		return
	}
	if webhooks == nil {
		webhooks = []*domain.Webhook{}
	}
	writeJSON(w, http.StatusOK, webhooks)
}

func (h *WebhookHandlers) GetWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	webhook, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, webhook)
}

type updateWebhookRequest struct {
	Description string                `json:"description"`
	EventTypes  []string              `json:"event_types"`
	Status      domain.WebhookStatus  `json:"status"`
}

func (h *WebhookHandlers) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	var req updateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	webhook, err := h.svc.Update(r.Context(), id, req.Description, req.EventTypes, req.Status)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, webhook)
}

func (h *WebhookHandlers) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type rotateSecretResponse struct {
	*domain.Webhook
	Secret string `json:"secret"`
}

func (h *WebhookHandlers) RotateSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	webhook, secret, err := h.svc.RotateSecret(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rotateSecretResponse{Webhook: webhook, Secret: secret})
}

func (h *WebhookHandlers) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "webhookID")
	deliveries, err := h.svc.ListDeliveries(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}
	if deliveries == nil {
		deliveries = []*domain.WebhookDelivery{}
	}
	writeJSON(w, http.StatusOK, deliveries)
}

func (h *WebhookHandlers) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrWebhookNotFound):
		writeError(w, http.StatusNotFound, "webhook not found")
	case errors.Is(err, domain.ErrInvalidWebhook):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
