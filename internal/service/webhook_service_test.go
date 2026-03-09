package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/mock"
	"github.com/robwittman/pillar/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testWebhookLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestWebhookCreate_Success(t *testing.T) {
	var captured *domain.Webhook
	repo := &mock.WebhookRepository{
		CreateFn: func(ctx context.Context, webhook *domain.Webhook) error {
			captured = webhook
			return nil
		},
	}
	deliveryRepo := &mock.WebhookDeliveryRepository{}
	svc := service.NewWebhookService(repo, deliveryRepo, testWebhookLogger())

	webhook, secret, err := svc.Create(context.Background(), "https://example.com/hook", "test hook", []string{"agent.created"})
	require.NoError(t, err)
	assert.NotEmpty(t, webhook.ID)
	assert.Equal(t, "https://example.com/hook", webhook.URL)
	assert.Equal(t, "test hook", webhook.Description)
	assert.Equal(t, []string{"agent.created"}, webhook.EventTypes)
	assert.Equal(t, domain.WebhookStatusActive, webhook.Status)
	assert.NotEmpty(t, secret)
	assert.Equal(t, 64, len(secret)) // 32 bytes hex encoded
	assert.Equal(t, captured.Secret, secret)
}

func TestWebhookCreate_EmptyURL(t *testing.T) {
	repo := &mock.WebhookRepository{}
	deliveryRepo := &mock.WebhookDeliveryRepository{}
	svc := service.NewWebhookService(repo, deliveryRepo, testWebhookLogger())

	_, _, err := svc.Create(context.Background(), "", "desc", nil)
	assert.ErrorIs(t, err, domain.ErrInvalidWebhook)
}

func TestWebhookCreate_NilEventTypes(t *testing.T) {
	repo := &mock.WebhookRepository{
		CreateFn: func(ctx context.Context, webhook *domain.Webhook) error {
			assert.NotNil(t, webhook.EventTypes)
			return nil
		},
	}
	deliveryRepo := &mock.WebhookDeliveryRepository{}
	svc := service.NewWebhookService(repo, deliveryRepo, testWebhookLogger())

	_, _, err := svc.Create(context.Background(), "https://example.com", "", nil)
	require.NoError(t, err)
}

func TestWebhookGet_Success(t *testing.T) {
	expected := &domain.Webhook{ID: "wh-1", URL: "https://example.com"}
	repo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return expected, nil
		},
	}
	svc := service.NewWebhookService(repo, nil, testWebhookLogger())

	webhook, err := svc.Get(context.Background(), "wh-1")
	require.NoError(t, err)
	assert.Equal(t, expected, webhook)
}

func TestWebhookGet_NotFound(t *testing.T) {
	repo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return nil, domain.ErrWebhookNotFound
		},
	}
	svc := service.NewWebhookService(repo, nil, testWebhookLogger())

	_, err := svc.Get(context.Background(), "missing")
	assert.ErrorIs(t, err, domain.ErrWebhookNotFound)
}

func TestWebhookList_Success(t *testing.T) {
	expected := []*domain.Webhook{{ID: "1"}, {ID: "2"}}
	repo := &mock.WebhookRepository{
		ListFn: func(ctx context.Context) ([]*domain.Webhook, error) {
			return expected, nil
		},
	}
	svc := service.NewWebhookService(repo, nil, testWebhookLogger())

	webhooks, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expected, webhooks)
}

func TestWebhookUpdate_Success(t *testing.T) {
	existing := &domain.Webhook{
		ID:          "wh-1",
		URL:         "https://example.com",
		EventTypes:  []string{"agent.created"},
		Status:      domain.WebhookStatusActive,
		Description: "old",
	}
	repo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, webhook *domain.Webhook) error {
			return nil
		},
	}
	svc := service.NewWebhookService(repo, nil, testWebhookLogger())

	webhook, err := svc.Update(context.Background(), "wh-1", "new desc", []string{"agent.deleted"}, domain.WebhookStatusInactive)
	require.NoError(t, err)
	assert.Equal(t, "new desc", webhook.Description)
	assert.Equal(t, []string{"agent.deleted"}, webhook.EventTypes)
	assert.Equal(t, domain.WebhookStatusInactive, webhook.Status)
}

func TestWebhookDelete_Success(t *testing.T) {
	repo := &mock.WebhookRepository{
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	svc := service.NewWebhookService(repo, nil, testWebhookLogger())

	err := svc.Delete(context.Background(), "wh-1")
	assert.NoError(t, err)
}

func TestWebhookRotateSecret_Success(t *testing.T) {
	existing := &domain.Webhook{
		ID:     "wh-1",
		Secret: "old-secret",
	}
	var updatedSecret string
	repo := &mock.WebhookRepository{
		GetFn: func(ctx context.Context, id string) (*domain.Webhook, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, webhook *domain.Webhook) error {
			updatedSecret = webhook.Secret
			return nil
		},
	}
	svc := service.NewWebhookService(repo, nil, testWebhookLogger())

	_, secret, err := svc.RotateSecret(context.Background(), "wh-1")
	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.NotEqual(t, "old-secret", secret)
	assert.Equal(t, updatedSecret, secret)
}

func TestWebhookListDeliveries_Success(t *testing.T) {
	expected := []*domain.WebhookDelivery{{ID: "d-1"}, {ID: "d-2"}}
	deliveryRepo := &mock.WebhookDeliveryRepository{
		ListByWebhookFn: func(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error) {
			assert.Equal(t, "wh-1", webhookID)
			return expected, nil
		},
	}
	svc := service.NewWebhookService(nil, deliveryRepo, testWebhookLogger())

	deliveries, err := svc.ListDeliveries(context.Background(), "wh-1")
	require.NoError(t, err)
	assert.Equal(t, expected, deliveries)
}
