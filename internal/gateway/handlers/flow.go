package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func SetupFlowRoutes(router fiber.Router, logger *zap.Logger) {
	flows := router.Group("/flows")

	// Get all flows
	flows.Get("/", func(c *fiber.Ctx) error {
		// TODO: Implement actual database query
		return c.JSON([]fiber.Map{
			{
				"id":         uuid.New().String(),
				"name":       "Основной поток",
				"is_active":  true,
				"created_at": time.Now(),
				"updated_at": time.Now(),
			},
		})
	})

	// Get flow by ID
	flows.Get("/:id", func(c *fiber.Ctx) error {
		flowID := c.Params("id")

		// TODO: Implement actual database query
		return c.JSON(fiber.Map{
			"id":   flowID,
			"name": "Основной поток",
			"flow_data": fiber.Map{
				"nodes": []interface{}{},
				"edges": []interface{}{},
			},
			"is_active":  true,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		})
	})

	// Create new flow
	flows.Post("/", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// TODO: Save to database
		body["id"] = uuid.New().String()
		body["created_at"] = time.Now()
		body["updated_at"] = time.Now()

		return c.Status(fiber.StatusCreated).JSON(body)
	})

	// Update flow
	flows.Put("/:id", func(c *fiber.Ctx) error {
		flowID := c.Params("id")

		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// TODO: Update in database
		body["id"] = flowID
		body["updated_at"] = time.Now()

		return c.JSON(body)
	})

	// Delete flow
	flows.Delete("/:id", func(c *fiber.Ctx) error {
		flowID := c.Params("id")

		// TODO: Delete from database
		logger.Info("Deleting flow", zap.String("flow_id", flowID))

		return c.SendStatus(fiber.StatusNoContent)
	})
}
