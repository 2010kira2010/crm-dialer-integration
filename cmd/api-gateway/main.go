package main

import (
	"go.uber.org/zap"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"

	"crm-dialer-integration/internal/gateway/handlers"
	"crm-dialer-integration/internal/gateway/middleware"
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

	// Create fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(cfg.LogLevel))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Static files for frontend
	app.Static("/", "./web/build", fiber.Static{
		Compress: true,
		Index:    "index.html",
	})

	// Health check (public)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	// Metrics endpoint (public)
	app.Get("/metrics", middleware.PrometheusHandler())

	// API routes
	api := app.Group("/api/v1")

	// Apply auth middleware to all API routes
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Setup routes (auth middleware will handle public endpoints internally)
	handlers.SetupAuthRoutes(api, cfg.JWTSecret, repo, log)
	handlers.SetupWebhookRoutes(api, log)
	handlers.SetupCRMRoutes(api, cfg, repo, log)
	handlers.SetupFlowRoutes(api, log)
	handlers.SetupDialerRoutes(api, log)

	// Fallback to index.html for SPA (should be last)
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./web/build/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("Starting API Gateway on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
