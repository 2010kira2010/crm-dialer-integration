package handlers

import (
	"strconv"
	"time"

	amocrmLib "github.com/2010kira2010/amocrm"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/repository"
	"crm-dialer-integration/internal/services/amocrm"
	"crm-dialer-integration/pkg/config"
)

type CRMHandler struct {
	amocrmService *amocrm.Service
	repo          *repository.Repository
	logger        *zap.Logger
}

func SetupCRMRoutes(router fiber.Router, cfg *config.Config, repo *repository.Repository, logger *zap.Logger) {
	// Initialize AmoCRM service
	amocrmService, err := amocrm.NewService(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize AmoCRM service", zap.Error(err))
	}

	handler := &CRMHandler{
		amocrmService: amocrmService,
		repo:          repo,
		logger:        logger,
	}

	crm := router.Group("/amocrm")

	// OAuth endpoints (эти эндпоинты должны быть публичными для callback)
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

	// Status endpoint
	crm.Get("/status", handler.GetStatus)

	// Pipelines endpoints
	crm.Get("/pipelines", handler.GetPipelines)
}

func (h *CRMHandler) GetAuthURL(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

	state := c.Query("state", "default")
	authURL := h.amocrmService.GetAuthURL(state)

	return c.JSON(fiber.Map{
		"auth_url": authURL,
	})
}

func (h *CRMHandler) AuthCallback(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

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

	// В реальном приложении здесь бы был редирект на фронтенд с успешным статусом
	return c.JSON(fiber.Map{
		"message": "Successfully authorized",
		"status":  "success",
	})
}

func (h *CRMHandler) GetStatus(c *fiber.Ctx) error {
	status := fiber.Map{
		"service":     "AmoCRM",
		"initialized": h.amocrmService != nil,
	}

	if h.amocrmService != nil {
		token := h.amocrmService.GetToken()
		status["authorized"] = token != nil

		if token != nil {
			status["token_expires_at"] = token.ExpiresAt()
		}
	}

	return c.JSON(status)
}

func (h *CRMHandler) GetPipelines(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

	/*
		// Сначала пробуем получить из базы данных
		pipelines, err := h.repo.GetAmoCRMPipelines(c.Context())
		if err != nil {
			h.logger.Error("Failed to get pipelines from database", zap.Error(err))
		}

		if len(pipelines) == 0 {
			// Если в базе нет, получаем из AmoCRM
			pipelines, err = h.amocrmService.GetPipelines(c.Context())
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Failed to get fields",
					"details": err,
				})
			}

			// Сохраняем в базу данных
			for _, pipeline := range pipelines {
				if err := h.repo.SaveAmoCRMPipeline(c.Context(), pipeline); err != nil {
					h.logger.Error("Failed to save pipeline", zap.Error(err))
				}
			}
		}

		return c.JSON(pipelines)

	*/
	return nil
}

func (h *CRMHandler) GetFields(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

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
				"error":   "Failed to get fields",
				"details": err.Error(),
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
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

	entityType := c.Query("entity_type", "leads")

	fields, err := h.amocrmService.GetCustomFields(c.Context(), entityType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to sync fields",
			"details": err.Error(),
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
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

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
	if pipelineID := c.Query("pipeline_id"); pipelineID != "" {
		params["filter[pipeline_id]"] = pipelineID
	}
	if responsibleUserID := c.Query("responsible_user_id"); responsibleUserID != "" {
		params["filter[responsible_user_id]"] = responsibleUserID
	}
	if query := c.Query("query"); query != "" {
		params["query"] = query
	}

	leads, err := h.amocrmService.GetLeads(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get leads",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  leads,
		"page":  page,
		"limit": limit,
		"count": len(leads),
	})
}

func (h *CRMHandler) GetLead(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

	leadID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid lead ID",
		})
	}

	lead, err := h.amocrmService.GetLeadByID(c.Context(), leadID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get lead",
			"details": err.Error(),
		})
	}

	return c.JSON(lead)
}

func (h *CRMHandler) UpdateLead(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

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
			"error":   "Failed to get lead",
			"details": err.Error(),
		})
	}

	// Обновляем поля
	if name, ok := updateData["name"].(string); ok {
		lead.Name = name
	}
	if price, ok := updateData["price"].(float64); ok {
		lead.Price = int(price)
	}
	if statusID, ok := updateData["status_id"].(float64); ok {
		lead.StatusID = int(statusID)
	}

	// Обновляем время модификации
	lead.UpdatedAt = int(time.Now().Unix())

	// Сохраняем изменения
	if err := h.amocrmService.UpdateLeads(c.Context(), []*amocrmLib.Lead{lead}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to update lead",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Lead updated successfully",
		"lead":    lead,
	})
}

func (h *CRMHandler) GetContacts(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

	params := make(map[string]string)

	if query := c.Query("query"); query != "" {
		params["query"] = query
	}
	if limit := c.Query("limit"); limit != "" {
		params["limit"] = limit
	}
	if page := c.Query("page"); page != "" {
		params["page"] = page
	}

	contacts, err := h.amocrmService.GetContacts(c.Context(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get contacts",
			"details": err.Error(),
		})
	}

	return c.JSON(contacts)
}

func (h *CRMHandler) GetContact(c *fiber.Ctx) error {
	if h.amocrmService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "AmoCRM service is not available",
		})
	}

	contactID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid contact ID",
		})
	}

	contact, err := h.amocrmService.GetContactByID(c.Context(), contactID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get contact",
			"details": err.Error(),
		})
	}

	return c.JSON(contact)
}
