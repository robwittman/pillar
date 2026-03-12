package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupWebhookRouter(webhookSvc service.WebhookService) *chi.Mux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	h := NewHandlers(noopAgentService(), logger)
	ch := NewConfigHandlers(&mock.ConfigService{
		CreateConfigFn:         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error { return nil },
		GetConfigFn:            func(ctx context.Context, agentID string) (*domain.AgentConfig, error) { return nil, nil },
		GetConfigWithSecretsFn: func(ctx context.Context, agentID string) (*domain.AgentConfig, string, error) { return nil, "", nil },
		UpdateConfigFn:         func(ctx context.Context, config *domain.AgentConfig, apiCredential string) error { return nil },
		DeleteConfigFn:         func(ctx context.Context, agentID string) error { return nil },
	}, logger)
	wh := NewWebhookHandlers(webhookSvc, logger)
	ah := NewAttributeHandlers(noopAttributeService(), logger)
	RegisterRoutes(r, h, ch, wh, ah, nil)
	return r
}

func TestCreateWebhook_Success(t *testing.T) {
	svc := &mock.WebhookService{
		CreateFn: func(ctx context.Context, url, description string, eventTypes []string) (*domain.Webhook, string, error) {
			return &domain.Webhook{
				ID:          "wh-1",
				URL:         url,
				Description: description,
				EventTypes:  eventTypes,
				Status:      domain.WebhookStatusActive,
			}, "secret-123", nil
		},
	}
	r := setupWebhookRouter(svc)

	body := `{"url":"https://example.com","description":"test","event_types":["agent.created"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "wh-1", resp["id"])
	assert.Equal(t, "secret-123", resp["secret"])
}

func TestCreateWebhook_InvalidURL(t *testing.T) {
	svc := &mock.WebhookService{
		CreateFn: func(ctx context.Context, url, description string, eventTypes []string) (*domain.Webhook, string, error) {
			return nil, "", domain.ErrInvalidWebhook
		},
	}
	r := setupWebhookRouter(svc)

	body := `{"url":"","description":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListWebhooks_Success(t *testing.T) {
	svc := &mock.WebhookService{
		ListFn: func(ctx context.Context) ([]*domain.Webhook, error) {
			return []*domain.Webhook{{ID: "1"}, {ID: "2"}}, nil
		},
	}
	r := setupWebhookRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetWebhook_Success(t *testing.T) {
	svc := &mock.WebhookService{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return &domain.Webhook{ID: id, URL: "https://example.com"}, nil
		},
	}
	r := setupWebhookRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/wh-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetWebhook_NotFound(t *testing.T) {
	svc := &mock.WebhookService{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return nil, domain.ErrWebhookNotFound
		},
	}
	r := setupWebhookRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/missing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWebhook_Success(t *testing.T) {
	svc := &mock.WebhookService{
		UpdateFn: func(ctx context.Context, id string, description string, eventTypes []string, status domain.WebhookStatus) (*domain.Webhook, error) {
			return &domain.Webhook{ID: id, Description: description}, nil
		},
	}
	r := setupWebhookRouter(svc)

	body := `{"description":"updated","event_types":["agent.deleted"],"status":"inactive"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/webhooks/wh-1", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteWebhook_Success(t *testing.T) {
	svc := &mock.WebhookService{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	r := setupWebhookRouter(svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/webhooks/wh-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRotateSecret_Success(t *testing.T) {
	svc := &mock.WebhookService{
		RotateSecretFn: func(ctx context.Context, id string) (*domain.Webhook, string, error) {
			return &domain.Webhook{ID: id}, "new-secret", nil
		},
	}
	r := setupWebhookRouter(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/wh-1/rotate-secret", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "new-secret", resp["secret"])
}

func TestListDeliveries_Success(t *testing.T) {
	svc := &mock.WebhookService{
		ListDeliveriesFn: func(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error) {
			return []*domain.WebhookDelivery{{ID: "d-1"}}, nil
		},
	}
	r := setupWebhookRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/wh-1/deliveries", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
