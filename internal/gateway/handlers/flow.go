package handlers

import (
	"crm-dialer-integration/internal/models"
	"crm-dialer-integration/internal/repository"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func SetupFlowRoutes(router fiber.Router, repo *repository.Repository, logger *zap.Logger) {
	flows := router.Group("/flows")

	// Get all flows
	flows.Get("/", func(c *fiber.Ctx) error {
		ctx := c.Context()
		flowsList, err := repo.GetIntegrationFlows(ctx)
		if err != nil {
			logger.Error("Failed to get flows", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get flows",
			})
		}
		return c.JSON(flowsList)
	})

	// Get flow by ID
	flows.Get("/:id", func(c *fiber.Ctx) error {
		flowID := c.Params("id")
		ctx := c.Context()

		flow, err := repo.GetIntegrationFlowByID(ctx, flowID)
		if err != nil {
			logger.Error("Failed to get flow", zap.String("flow_id", flowID), zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get flow",
			})
		}

		if flow == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Flow not found",
			})
		}

		return c.JSON(flow)
	})

	// Create new flow
	flows.Post("/", func(c *fiber.Ctx) error {
		var body models.IntegrationFlow
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Generate new ID
		body.ID = uuid.New().String()
		body.CreatedAt = time.Now()
		body.UpdatedAt = time.Now()

		ctx := c.Context()
		if err := repo.CreateIntegrationFlow(ctx, &body); err != nil {
			logger.Error("Failed to create flow", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create flow",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(body)
	})

	// Update flow
	flows.Put("/:id", func(c *fiber.Ctx) error {
		flowID := c.Params("id")

		var body models.IntegrationFlow
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Ensure ID matches
		body.ID = flowID
		body.UpdatedAt = time.Now()

		ctx := c.Context()
		if err := repo.UpdateIntegrationFlow(ctx, &body); err != nil {
			logger.Error("Failed to update flow", zap.String("flow_id", flowID), zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update flow",
			})
		}

		return c.JSON(body)
	})

	// Delete flow
	flows.Delete("/:id", func(c *fiber.Ctx) error {
		flowID := c.Params("id")
		ctx := c.Context()

		if err := repo.DeleteIntegrationFlow(ctx, flowID); err != nil {
			logger.Error("Failed to delete flow", zap.String("flow_id", flowID), zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete flow",
			})
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
