package services

import (
	"context"
	"strings"
	"time"

	"crm-dialer-integration/pkg/nats"
	"go.uber.org/zap"
)

type WebhookService struct {
	logger     *zap.Logger
	natsClient *nats.Client
}

func NewWebhookService(logger *zap.Logger) *WebhookService {
	// TODO: Inject NATS client properly
	natsClient, _ := nats.NewClient("nats://nats:4222", logger)

	return &WebhookService{
		logger:     logger,
		natsClient: natsClient,
	}
}

func (s *WebhookService) ProcessWebhook(ctx context.Context, eventType string, payload map[string]interface{}) error {
	s.logger.Info("Processing webhook",
		zap.String("event_type", eventType),
		zap.Any("payload", payload))

	// Create event
	event := map[string]interface{}{
		"type":      eventType,
		"payload":   payload,
		"timestamp": time.Now().Unix(),
	}

	// Publish to NATS
	subject := "webhooks.amocrm." + strings.ReplaceAll(eventType, ".", "_")
	if err := s.natsClient.Publish(subject, event); err != nil {
		s.logger.Error("Failed to publish event", zap.Error(err))
		return err
	}

	s.logger.Info("Webhook published to NATS",
		zap.String("subject", subject),
		zap.String("event_type", eventType))

	return nil
}
