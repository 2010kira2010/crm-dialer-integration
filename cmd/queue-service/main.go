package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/services/queue"
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

	// Initialize Queue service
	queueService := queue.NewQueueService(log)

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	// Subscribe to queue requests
	_, err = nc.Subscribe("queue.amocrm.*", func(msg *nats.Msg) {
		log.Info("Received queue request",
			zap.String("subject", msg.Subject))

		var request struct {
			Type     string        `json:"type"`
			Entities []interface{} `json:"entities"`
		}

		if err := json.Unmarshal(msg.Data, &request); err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			return
		}

		// Add to queue with rate limiting
		ctx := context.Background()
		if err := queueService.AddToQueue(ctx, request.Type, request.Entities); err != nil {
			log.Error("Failed to add to queue", zap.Error(err))
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	log.Info("Queue Service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Queue Service...")
}
