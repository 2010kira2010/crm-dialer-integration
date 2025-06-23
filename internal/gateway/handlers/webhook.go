package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/gateway/services"
)

func SetupWebhookRoutes(router fiber.Router, logger *zap.Logger) {
	webhookService := services.NewWebhookService(logger)

	webhook := router.Group("/webhooks")

	// AmoCRM webhooks
	webhook.Post("/amocrm/lead/add", handleAmoCRMWebhook(webhookService, "lead.add"))
	webhook.Post("/amocrm/lead/update", handleAmoCRMWebhook(webhookService, "lead.update"))
	webhook.Post("/amocrm/lead/delete", handleAmoCRMWebhook(webhookService, "lead.delete"))
	webhook.Post("/amocrm/lead/status", handleAmoCRMWebhook(webhookService, "lead.status"))
	webhook.Post("/amocrm/lead/responsible", handleAmoCRMWebhook(webhookService, "lead.responsible"))
}

func handleAmoCRMWebhook(service *services.WebhookService, eventType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload map[string]interface{}
		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid payload",
			})
		}

		if err := service.ProcessWebhook(c.Context(), eventType, payload); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process webhook",
			})
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
