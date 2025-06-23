package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func SetupDialerRoutes(router fiber.Router, logger *zap.Logger) {
	dialer := router.Group("/dialer")

	// Get schedulers
	dialer.Get("/schedulers", func(c *fiber.Ctx) error {
		// TODO: Implement actual service call
		return c.JSON([]fiber.Map{
			{
				"id":   "550e8400-e29b-41d4-a716-446655440000",
				"name": "Утренний обзвон",
			},
			{
				"id":   "550e8400-e29b-41d4-a716-446655440001",
				"name": "Вечерний обзвон",
			},
		})
	})

	// Get campaigns
	dialer.Get("/campaigns", func(c *fiber.Ctx) error {
		// TODO: Implement actual service call
		return c.JSON([]fiber.Map{
			{
				"id":   "650e8400-e29b-41d4-a716-446655440000",
				"name": "Холодные звонки",
			},
			{
				"id":   "650e8400-e29b-41d4-a716-446655440001",
				"name": "Повторные звонки",
			},
		})
	})

	// Get buckets
	dialer.Get("/buckets", func(c *fiber.Ctx) error {
		campaignID := c.Query("campaign_id")

		// TODO: Implement actual service call with filtering
		_ = campaignID

		return c.JSON([]fiber.Map{
			{
				"id":          "750e8400-e29b-41d4-a716-446655440000",
				"campaign_id": "650e8400-e29b-41d4-a716-446655440000",
				"name":        "Новые лиды",
			},
		})
	})

	// Sync data from dialer
	dialer.Post("/sync", func(c *fiber.Ctx) error {
		// TODO: Implement sync logic
		return c.JSON(fiber.Map{
			"message": "Dialer data sync started",
		})
	})
}
