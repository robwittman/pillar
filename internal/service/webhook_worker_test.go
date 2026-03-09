package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookWorker_DeliverySuccess(t *testing.T) {
	var receivedBody []byte
	var receivedHeaders http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		receivedBody = buf[:n]
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	webhook := &domain.Webhook{ID: "wh-1", URL: ts.URL, Secret: "test-secret"}
	payload, _ := json.Marshal(domain.Event{ID: "evt-1", Type: "agent.created"})
	delivery := &domain.WebhookDelivery{
		ID:        "d-1",
		WebhookID: "wh-1",
		EventType: "agent.created",
		Payload:   payload,
		Status:    domain.DeliveryStatusPending,
	}

	webhookRepo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return webhook, nil
		},
	}

	var updated *domain.WebhookDelivery
	deliveryRepo := &mock.WebhookDeliveryRepository{
		ListPendingFn: func(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
			return []*domain.WebhookDelivery{delivery}, nil
		},
		UpdateFn: func(ctx context.Context, d *domain.WebhookDelivery) error {
			updated = d
			return nil
		},
	}

	worker := service.NewWebhookWorker(webhookRepo, deliveryRepo, testWebhookLogger())
	service.SetHTTPClient(worker, ts.Client())

	service.ProcessBatch(worker, context.Background())

	require.NotNil(t, updated)
	assert.Equal(t, domain.DeliveryStatusDelivered, updated.Status)
	assert.Equal(t, 1, updated.Attempts)
	assert.Equal(t, 200, updated.ResponseCode)
	assert.NotNil(t, receivedBody)
	assert.Equal(t, "agent.created", receivedHeaders.Get("X-Pillar-Event"))
	assert.Equal(t, "d-1", receivedHeaders.Get("X-Pillar-Delivery"))

	// Verify HMAC signature
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write(payload)
	expectedSig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expectedSig, receivedHeaders.Get("X-Pillar-Signature"))
}

func TestWebhookWorker_DeliveryRetry(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer ts.Close()

	webhook := &domain.Webhook{ID: "wh-1", URL: ts.URL, Secret: "secret"}
	delivery := &domain.WebhookDelivery{
		ID:        "d-1",
		WebhookID: "wh-1",
		EventType: "agent.created",
		Payload:   json.RawMessage(`{}`),
		Status:    domain.DeliveryStatusPending,
		Attempts:  0,
	}

	webhookRepo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return webhook, nil
		},
	}

	var updated *domain.WebhookDelivery
	deliveryRepo := &mock.WebhookDeliveryRepository{
		ListPendingFn: func(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
			return []*domain.WebhookDelivery{delivery}, nil
		},
		UpdateFn: func(ctx context.Context, d *domain.WebhookDelivery) error {
			updated = d
			return nil
		},
	}

	worker := service.NewWebhookWorker(webhookRepo, deliveryRepo, testWebhookLogger())
	service.SetHTTPClient(worker, ts.Client())

	service.ProcessBatch(worker, context.Background())

	require.NotNil(t, updated)
	assert.Equal(t, domain.DeliveryStatusPending, updated.Status)
	assert.Equal(t, 1, updated.Attempts)
	assert.NotNil(t, updated.NextRetryAt)
	assert.Equal(t, 500, updated.ResponseCode)
}

func TestWebhookWorker_MaxAttemptsExceeded(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	webhook := &domain.Webhook{ID: "wh-1", URL: ts.URL, Secret: "secret"}
	delivery := &domain.WebhookDelivery{
		ID:        "d-1",
		WebhookID: "wh-1",
		EventType: "agent.created",
		Payload:   json.RawMessage(`{}`),
		Status:    domain.DeliveryStatusPending,
		Attempts:  service.MaxAttempts() - 1,
	}

	webhookRepo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return webhook, nil
		},
	}

	var updated *domain.WebhookDelivery
	deliveryRepo := &mock.WebhookDeliveryRepository{
		ListPendingFn: func(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
			return []*domain.WebhookDelivery{delivery}, nil
		},
		UpdateFn: func(ctx context.Context, d *domain.WebhookDelivery) error {
			updated = d
			return nil
		},
	}

	worker := service.NewWebhookWorker(webhookRepo, deliveryRepo, testWebhookLogger())
	service.SetHTTPClient(worker, ts.Client())

	service.ProcessBatch(worker, context.Background())

	require.NotNil(t, updated)
	assert.Equal(t, domain.DeliveryStatusFailed, updated.Status)
	assert.Equal(t, service.MaxAttempts(), updated.Attempts)
}

func TestSignPayload(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "my-secret"

	sig := service.SignPayload(payload, secret)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	assert.Equal(t, expected, sig)
}
