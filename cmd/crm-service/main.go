package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/services/amocrm"
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

	// Initialize AmoCRM service
	amocrmService, err := amocrm.NewService(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize AmoCRM service", zap.Error(err))
	}

	// Load saved tokens
	ctx := context.Background()
	if tokens, err := amocrm.LoadTokens(ctx); err != nil {
		log.Error("Failed to load tokens", zap.Error(err))
	} else if tokens != nil {
		amocrmService.SetTokens(tokens)
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("CRM Service started")

	<-quit
	log.Info("Shutting down CRM Service...")
}
