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

	"crm-dialer-integration/internal/repository"
	"crm-dialer-integration/internal/services/flowengine"
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

	// Initialize Flow Engine
	engine := flowengine.NewFlowEngine(log, repo)

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	// Subscribe to lead events
	_, err = nc.Subscribe("webhooks.amocrm.lead_*", func(msg *nats.Msg) {
		log.Info("Processing lead event", zap.String("subject", msg.Subject))

		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Error("Failed to unmarshal event", zap.Error(err))
			return
		}

		// Process through flow engine
		ctx := context.Background()
		if err := engine.ProcessEvent(ctx, event); err != nil {
			log.Error("Failed to process event", zap.Error(err))
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	log.Info("Flow Engine Service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Flow Engine Service...")
}
