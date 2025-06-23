package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

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

	// Initialize repository (might be used for config storage in the future)
	_, err := repository.New(cfg.DatabaseURL, log)
	if err != nil {
		log.Warn("Failed to initialize repository, continuing without database support", zap.Error(err))
	}

	// Initialize Redis
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL", zap.Error(err))
	}

	rdb := redis.NewClient(opt)
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()

	// Subscribe to config requests
	_, err = nc.Subscribe("config.get", func(msg *nats.Msg) {
		var request struct {
			Key string `json:"key"`
		}

		if err := json.Unmarshal(msg.Data, &request); err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			return
		}

		// First check Redis cache
		val, err := rdb.Get(ctx, request.Key).Result()
		if err == redis.Nil {
			// Not in cache, get from database
			// TODO: Implement database lookup
			log.Debug("Config not found in cache", zap.String("key", request.Key))
		} else if err != nil {
			log.Error("Failed to get from Redis", zap.Error(err))
		} else {
			// Found in cache
			msg.Respond([]byte(val))
			return
		}

		// Default response
		msg.Respond([]byte("{}"))
	})

	_, err = nc.Subscribe("config.set", func(msg *nats.Msg) {
		var request struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			TTL   int    `json:"ttl"` // seconds
		}

		if err := json.Unmarshal(msg.Data, &request); err != nil {
			log.Error("Failed to unmarshal request", zap.Error(err))
			return
		}

		// Save to Redis with TTL
		if request.TTL > 0 {
			err = rdb.SetEX(ctx, request.Key, request.Value, time.Duration(request.TTL)*time.Second).Err()
		} else {
			err = rdb.Set(ctx, request.Key, request.Value, 0).Err()
		}

		if err != nil {
			log.Error("Failed to save to Redis", zap.Error(err))
			msg.Respond([]byte(`{"error": "Failed to save"}`))
		} else {
			msg.Respond([]byte(`{"status": "ok"}`))
		}
	})

	if err != nil {
		log.Fatal("Failed to subscribe to NATS", zap.Error(err))
	}

	log.Info("Config Service started")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Config Service...")
}
