package main

import (
	"context"
	"crm-dialer-integration/internal/models"
	"crm-dialer-integration/internal/services/amocrm"
	"encoding/json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"

	"crm-dialer-integration/internal/repository"
	"crm-dialer-integration/pkg/config"
	"crm-dialer-integration/pkg/logger"
)

func main() {
	// Load .env file
	godotenv.Load()

	// Initialize config
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.LogLevel)

	// Initialize repository
	repo, err := repository.New(cfg.DatabaseURL, log)
	if err != nil {
		log.Fatal("Failed to initialize repository", zap.Error(err))
	}

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	// Initialize AmoCRM service
	amocrmService, err := amocrm.NewService(cfg, log)
	if err != nil {
		log.Error("Failed to initialize AmoCRM service", zap.Error(err))
	}

	// Initialize webhook processor with NATS support
	var processor *amocrm.WebhookProcessor
	if amocrmService != nil {
		processor = amocrm.NewWebhookProcessorWithNATS(amocrmService, log, nc)
	}

	// Subscribe to webhook events
	_, err = nc.Subscribe("webhooks.amocrm.*", func(msg *nats.Msg) {
		log.Info("Received webhook event",
			zap.String("subject", msg.Subject),
			zap.Int("size", len(msg.Data)))

		// Parse webhook data
		var webhookData struct {
			Type      string                 `json:"type"`
			Payload   map[string]interface{} `json:"payload"`
			Timestamp int64                  `json:"timestamp"`
		}

		if err := json.Unmarshal(msg.Data, &webhookData); err != nil {
			log.Error("Failed to unmarshal webhook data", zap.Error(err))
			return
		}

		// Save to database
		webhookLog := &models.WebhookLog{
			ID:          uuid.New().String(),
			WebhookType: webhookData.Type,
			RawData:     json.RawMessage(msg.Data),
			ProcessedAt: time.Now(),
			Status:      "processing",
		}

		ctx := context.Background()
		if err := repo.SaveWebhookLog(ctx, webhookLog); err != nil {
			log.Error("Failed to save webhook log", zap.Error(err))
		}

		// Process webhook with AmoCRM processor if available
		if processor != nil {
			if err := processor.ProcessLeadWebhook(ctx, webhookData.Type, webhookData.Payload); err != nil {
				log.Error("Failed to process webhook",
					zap.String("type", webhookData.Type),
					zap.Error(err))
				webhookLog.Status = "failed"
			} else {
				webhookLog.Status = "processed"
			}

			// Update webhook status
			if err := repo.SaveWebhookLog(ctx, webhookLog); err != nil {
				log.Error("Failed to update webhook status", zap.Error(err))
			}
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	log.Info("Webhook Service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Webhook Service...")
}
