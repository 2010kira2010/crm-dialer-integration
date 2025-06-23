package main

import (
	"context"
	"crm-dialer-integration/internal/models"
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

	// Subscribe to webhook events
	_, err = nc.Subscribe("webhooks.amocrm.*", func(msg *nats.Msg) {
		log.Info("Received webhook event",
			zap.String("subject", msg.Subject),
			zap.Int("size", len(msg.Data)))

		// Save to database
		webhookLog := &models.WebhookLog{
			ID:          uuid.New().String(),
			WebhookType: msg.Subject,
			RawData:     json.RawMessage(msg.Data),
			ProcessedAt: time.Now(),
			Status:      "received",
		}

		if err := repo.SaveWebhookLog(context.Background(), webhookLog); err != nil {
			log.Error("Failed to save webhook log", zap.Error(err))
		}

		// TODO: Process webhook based on type
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
