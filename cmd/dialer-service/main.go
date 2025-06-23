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
	"crm-dialer-integration/internal/services/dialer"
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

	// Initialize Dialer service
	dialerService := dialer.NewService(cfg, repo, log)

	// Sync data on startup
	ctx := context.Background()
	if err := dialerService.SyncSchedulers(ctx); err != nil {
		log.Error("Failed to sync schedulers", zap.Error(err))
	}
	if err := dialerService.SyncCampaigns(ctx); err != nil {
		log.Error("Failed to sync campaigns", zap.Error(err))
	}

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	// Subscribe to dialer events
	_, err = nc.Subscribe("dialer.send_contact", func(msg *nats.Msg) {
		log.Info("Received send contact request")

		var data struct {
			SchedulerID string         `json:"scheduler_id"`
			CampaignID  string         `json:"campaign_id"`
			BucketID    string         `json:"bucket_id"`
			Contact     dialer.Contact `json:"contact"`
		}

		if err := json.Unmarshal(msg.Data, &data); err != nil {
			log.Error("Failed to unmarshal message", zap.Error(err))
			return
		}

		// Send contact to dialer
		if err := dialerService.SendContact(ctx, data.SchedulerID, data.CampaignID, data.BucketID, data.Contact); err != nil {
			log.Error("Failed to send contact to dialer", zap.Error(err))
			// TODO: Handle retry logic
		} else {
			log.Info("Contact sent to dialer successfully",
				zap.String("phone", data.Contact.Phone),
				zap.String("campaign_id", data.CampaignID))
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	log.Info("Dialer Service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Dialer Service...")
}
