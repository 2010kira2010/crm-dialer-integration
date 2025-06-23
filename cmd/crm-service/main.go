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

	// Проверяем наличие токенов
	ctx := context.Background()
	if tokens, err := amocrm.LoadTokens(ctx); err != nil {
		log.Warn("No saved tokens found", zap.Error(err))

		// Если нет сохраненных токенов и нет кода авторизации, выводим инструкцию
		if cfg.AmoCRMAuthCode == "" {
			log.Info("No auth code provided. Please authorize the application:")
			log.Info("1. Set AMOCRM_DOMAIN, AMOCRM_CLIENT_ID, AMOCRM_CLIENT_SECRET, AMOCRM_REDIRECT_URI in .env")
			log.Info("2. Get auth URL from API Gateway: GET /api/v1/amocrm/auth")
			log.Info("3. Complete authorization and get the code")
			log.Info("4. Set AMOCRM_AUTH_CODE in .env and restart the service")
		}
	} else {
		// Устанавливаем загруженные токены
		if err := amocrmService.SetTokens(tokens); err != nil {
			log.Error("Failed to set loaded tokens", zap.Error(err))
		} else {
			log.Info("Tokens loaded successfully",
				zap.Time("expires_at", tokens.ExpiresAt))
		}
	}

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("CRM Service started")

	// В реальном приложении здесь бы был gRPC сервер или другой способ
	// предоставления функциональности другим сервисам
	// Пока просто ждем сигнала завершения

	<-quit
	log.Info("Shutting down CRM Service...")

	// Здесь можно добавить graceful shutdown логику
}
