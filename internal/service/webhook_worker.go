package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/robwittman/pillar/internal/domain"
)

const (
	maxDeliveryAttempts = 5
	workerPollInterval  = 5 * time.Second
	workerBatchSize     = 50
)

type WebhookWorker struct {
	webhookRepo  domain.WebhookRepository
	deliveryRepo domain.WebhookDeliveryRepository
	httpClient   *http.Client
	logger       *slog.Logger
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

func NewWebhookWorker(webhookRepo domain.WebhookRepository, deliveryRepo domain.WebhookDeliveryRepository, logger *slog.Logger) *WebhookWorker {
	return &WebhookWorker{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		logger:       logger,
	}
}

func (w *WebhookWorker) Start(ctx context.Context) {
	ctx, w.cancel = context.WithCancel(ctx)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.run(ctx)
	}()
	w.logger.Info("webhook worker started")
}

func (w *WebhookWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	w.logger.Info("webhook worker stopped")
}

func (w *WebhookWorker) run(ctx context.Context) {
	ticker := time.NewTicker(workerPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *WebhookWorker) processBatch(ctx context.Context) {
	deliveries, err := w.deliveryRepo.ListPending(ctx, workerBatchSize)
	if err != nil {
		w.logger.Warn("failed to list pending deliveries", "error", err)
		return
	}

	for _, delivery := range deliveries {
		if ctx.Err() != nil {
			return
		}
		w.deliver(ctx, delivery)
	}
}

func (w *WebhookWorker) deliver(ctx context.Context, delivery *domain.WebhookDelivery) {
	webhook, err := w.webhookRepo.Get(ctx, delivery.WebhookID)
	if err != nil {
		w.logger.Warn("failed to get webhook for delivery", "delivery_id", delivery.ID, "webhook_id", delivery.WebhookID, "error", err)
		delivery.Status = domain.DeliveryStatusFailed
		now := time.Now()
		delivery.LastAttemptAt = &now
		w.deliveryRepo.Update(ctx, delivery)
		return
	}

	sig := signPayload(delivery.Payload, webhook.Secret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(delivery.Payload))
	if err != nil {
		w.logger.Warn("failed to create delivery request", "delivery_id", delivery.ID, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pillar-Event", delivery.EventType)
	req.Header.Set("X-Pillar-Delivery", delivery.ID)
	req.Header.Set("X-Pillar-Signature", "sha256="+sig)

	resp, err := w.httpClient.Do(req)
	now := time.Now()
	delivery.Attempts++
	delivery.LastAttemptAt = &now

	if err != nil {
		w.handleFailure(ctx, delivery, 0, err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	delivery.ResponseCode = resp.StatusCode
	delivery.ResponseBody = string(body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = domain.DeliveryStatusDelivered
		delivery.NextRetryAt = nil
		if err := w.deliveryRepo.Update(ctx, delivery); err != nil {
			w.logger.Warn("failed to update delivery", "delivery_id", delivery.ID, "error", err)
		}
		return
	}

	w.handleFailure(ctx, delivery, resp.StatusCode, string(body))
}

func (w *WebhookWorker) handleFailure(ctx context.Context, delivery *domain.WebhookDelivery, statusCode int, errMsg string) {
	if delivery.Attempts >= maxDeliveryAttempts {
		delivery.Status = domain.DeliveryStatusFailed
		delivery.NextRetryAt = nil
		w.logger.Warn("webhook delivery failed after max attempts",
			"delivery_id", delivery.ID,
			"webhook_id", delivery.WebhookID,
			"attempts", delivery.Attempts,
		)
	} else {
		backoff := time.Duration(math.Pow(2, float64(delivery.Attempts))) * time.Second
		nextRetry := time.Now().Add(backoff)
		delivery.NextRetryAt = &nextRetry
	}

	if err := w.deliveryRepo.Update(ctx, delivery); err != nil {
		w.logger.Warn("failed to update delivery after failure", "delivery_id", delivery.ID, "error", err)
	}
}

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func SignPayload(payload []byte, secret string) string {
	return signPayload(payload, secret)
}

func RetryBackoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

func MaxAttempts() int {
	return maxDeliveryAttempts
}

func SetHTTPClient(w *WebhookWorker, client *http.Client) {
	w.httpClient = client
}

func ProcessBatch(w *WebhookWorker, ctx context.Context) {
	w.processBatch(ctx)
}

func WorkerPollInterval() time.Duration {
	return workerPollInterval
}

func FormatSignatureHeader(sig string) string {
	return fmt.Sprintf("sha256=%s", sig)
}
