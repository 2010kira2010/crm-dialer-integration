package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// NATS
	NatsURL string

	// JWT
	JWTSecret string

	// CORS
	CORSOrigins string

	// AmoCRM
	AmoCRMDomain       string
	AmoCRMClientID     string
	AmoCRMClientSecret string
	AmoCRMRedirectURI  string
	AmoCRMAuthCode     string // Код авторизации для первичного получения токена

	// Dialer
	DialerAPIURL string
	DialerAPIKey string

	// Logging
	LogLevel string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/crm_dialer?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),

		JWTSecret:   getEnv("JWT_SECRET", "default-secret-key"),
		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:3000"),

		AmoCRMDomain:       getEnv("AMOCRM_DOMAIN", ""),
		AmoCRMClientID:     getEnv("AMOCRM_CLIENT_ID", ""),
		AmoCRMClientSecret: getEnv("AMOCRM_CLIENT_SECRET", ""),
		AmoCRMRedirectURI:  getEnv("AMOCRM_REDIRECT_URI", ""),
		AmoCRMAuthCode:     getEnv("AMOCRM_AUTH_CODE", ""),

		DialerAPIURL: getEnv("DIALER_API_URL", ""),
		DialerAPIKey: getEnv("DIALER_API_KEY", ""),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := strings.ToLower(getEnv(key, ""))
	if strValue == "true" || strValue == "1" {
		return true
	} else if strValue == "false" || strValue == "0" {
		return false
	}
	return defaultValue
}
