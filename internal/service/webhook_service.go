package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	"github.com/google/uuid"
	"github.com/robwittman/pillar/internal/domain"
)

type WebhookService interface {
	Create(ctx context.Context, url, description string, eventTypes []string) (*domain.Webhook, string, error)
	Get(ctx context.Context, id string) (*domain.Webhook, error)
	List(ctx context.Context) ([]*domain.Webhook, error)
	Update(ctx context.Context, id string, description string, eventTypes []string, status domain.WebhookStatus) (*domain.Webhook, error)
	Delete(ctx context.Context, id string) error
	RotateSecret(ctx context.Context, id string) (*domain.Webhook, string, error)
	ListDeliveries(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error)
}

type webhookService struct {
	repo         domain.WebhookRepository
	deliveryRepo domain.WebhookDeliveryRepository
	logger       *slog.Logger
}

func NewWebhookService(repo domain.WebhookRepository, deliveryRepo domain.WebhookDeliveryRepository, logger *slog.Logger) WebhookService {
	return &webhookService{
		repo:         repo,
		deliveryRepo: deliveryRepo,
		logger:       logger,
	}
}

func (s *webhookService) Create(ctx context.Context, url, description string, eventTypes []string) (*domain.Webhook, string, error) {
	if url == "" {
		return nil, "", domain.ErrInvalidWebhook
	}
	if len(eventTypes) == 0 {
		eventTypes = []string{}
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, "", err
	}

	webhook := &domain.Webhook{
		ID:          uuid.New().String(),
		URL:         url,
		Secret:      secret,
		EventTypes:  eventTypes,
		Status:      domain.WebhookStatusActive,
		Description: description,
	}

	if err := s.repo.Create(ctx, webhook); err != nil {
		return nil, "", err
	}

	s.logger.Info("webhook created", "id", webhook.ID, "url", webhook.URL)
	return webhook, secret, nil
}

func (s *webhookService) Get(ctx context.Context, id string) (*domain.Webhook, error) {
	return s.repo.Get(ctx, id)
}

func (s *webhookService) List(ctx context.Context) ([]*domain.Webhook, error) {
	return s.repo.List(ctx)
}

func (s *webhookService) Update(ctx context.Context, id string, description string, eventTypes []string, status domain.WebhookStatus) (*domain.Webhook, error) {
	webhook, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	webhook.Description = description
	if eventTypes != nil {
		webhook.EventTypes = eventTypes
	}
	if status != "" {
		webhook.Status = status
	}

	if err := s.repo.Update(ctx, webhook); err != nil {
		return nil, err
	}
	return webhook, nil
}

func (s *webhookService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *webhookService) RotateSecret(ctx context.Context, id string) (*domain.Webhook, string, error) {
	webhook, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, "", err
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, "", err
	}

	webhook.Secret = secret
	if err := s.repo.Update(ctx, webhook); err != nil {
		return nil, "", err
	}

	s.logger.Info("webhook secret rotated", "id", webhook.ID)
	return webhook, secret, nil
}

func (s *webhookService) ListDeliveries(ctx context.Context, webhookID string) ([]*domain.WebhookDelivery, error) {
	return s.deliveryRepo.ListByWebhook(ctx, webhookID)
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
