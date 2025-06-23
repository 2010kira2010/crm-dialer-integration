package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/repository"
	"crm-dialer-integration/internal/services/amocrm"
)

type CRMHandler struct {
	amocrmService *amocrm.Service
	repo          *repository.Repository
	logger        *zap.Logger
}

func SetupCRMRoutes(router fiber.Router, logger *zap.Logger) {
	// TODO: Properly inject dependencies
	handler := &CRMHandler{
		logger: logger,
	}

	crm := router.Group("/amocrm")

	// OAuth endpoints
	crm.Get("/auth", handler.GetAuthURL)
	crm.Get("/auth/callback", handler.AuthCallback)

	// Fields endpoints
	crm.Get("/fields", handler.GetFields)
	crm.Post("/fields/sync", handler.SyncFields)

	// Leads endpoints
	crm.Get("/leads", handler.GetLeads)
	crm.Get("/leads/:id", handler.GetLead)
	crm.Put("/leads/:id", handler.UpdateLead)

	// Contacts endpoints
	crm.Get("/contacts", handler.GetContacts)
	crm.Get("/contacts/:id", handler.GetContact)
}

func (h *CRMHandler) GetAuthURL(c *fiber.Ctx) error {
	state := c.Query("state", "default")
	authURL := h.amocrmService.GetAuthURL(state)

	return c.JSON(fiber.Map{
		"auth_url": authURL,
	})
}

func (h *CRMHandler) AuthCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Authorization code is required",
		})
	}

	if err := h.amocrmService.ExchangeCode(c.Context(), code); err != nil {
		h.logger.Error("Failed to exchange code", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to authorize",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully authorized",
	})
}

func (h *CRMHandler) GetFields(c *fiber.Ctx) error {
	entityType := c.Query("entity_type", "leads")

	// Сначала пробуем получить из базы данных
	fields, err := h.repo.GetAmoCRMFields(c.Context(), entityType)
	if err != nil {
		h.logger.Error("Failed to get fields from database", zap.Error(err))
	}

	if len(fields) == 0 {
		// Если в базе нет, получаем из AmoCRM
		fields, err = h.amocrmService.GetCustomFields(c.Context(), entityType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get fields",
			})
		}

		// Сохраняем в базу данных
		for _, field := range fields {
			if err := h.repo.SaveAmoCRMField(c.Context(), field); err != nil {
				h.logger.Error("Failed to save field", zap.Error(err))
			}
		}
	}

	return c.JSON(fields)
}

func (h *CRMHandler) SyncFields(c *fiber.Ctx) error {
	entityType := c.Query("entity_type", "leads")

	fields, err := h.amocrmService.GetCustomFields(c.Context(), entityType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to sync fields",
		})
	}

	// Обновляем поля в базе данных
	for _, field := range fields {
		if err := h.repo.UpsertAmoCRMField(c.Context(), field); err != nil {
			h.logger.Error("Failed to upsert field", zap.Error(err))
		}
	}

	return c.JSON(fiber.Map{
		"message": "Fields sync completed",
		"count":   len(fields),
	})
}

func (h *CRMHandler) GetLeads(c *fiber.Ctx) error {
	// Параметры пагинации
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	params := map[string]string{
		"page":  strconv.Itoa(page),
		"limit": strconv.Itoa(limit),
	}

	// Дополнительные фильтры
	if statusID := c.Query("status_id"); statusID != "" {
		params["filter[status_id]"] = statusID
	}
	if responsibleUserID := c.Query("responsible_user_id"); responsibleUserID != "" {
		params["filter[responsible_user_id]"] = responsibleUserID
	}

	leads, err := h.amocrmService.GetLeads(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get leads",
		})
	}

	return c.JSON(fiber.Map{
		"data":  leads,
		"page":  page,
		"limit": limit,
	})
}

func (h *CRMHandler) GetLead(c *fiber.Ctx) error {
	leadID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid lead ID",
		})
	}

	lead, err := h.amocrmService.GetLeadByID(c.Context(), leadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get lead",
		})
	}

	return c.JSON(lead)
}

func (h *CRMHandler) UpdateLead(c *fiber.Ctx) error {
	leadID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid lead ID",
		})
	}

	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Получаем текущую сделку
	lead, err := h.amocrmService.GetLeadByID(c.Context(), leadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get lead",
		})
	}

	// Обновляем поля
	// TODO: Правильно мапить поля из updateData в lead

	// Сохраняем изменения
	if err := h.amocrmService.UpdateLeads(c.Context(), []*amocrm.Lead{lead}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update lead",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Lead updated successfully",
	})
}

func (h *CRMHandler) GetContacts(c *fiber.Ctx) error {
	params := make(map[string]string)

	if query := c.Query("query"); query != "" {
		params["query"] = query
	}

	contacts, err := h.amocrmService.GetContacts(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get contacts",
		})
	}

	return c.JSON(contacts)
}

func (h *CRMHandler) GetContact(c *fiber.Ctx) error {
	contactID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid contact ID",
		})
	}

	contact, err := h.amocrmService.GetContactByID(c.Context(), contactID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get contact",
		})
	}

	return c.JSON(contact)
}
