package amocrm

import (
	"context"
	"fmt"

	"github.com/2010kira2010/amocrm"
	"go.uber.org/zap"
)

type WebhookProcessor struct {
	service *Service
	logger  *zap.Logger
}

func NewWebhookProcessor(service *Service, logger *zap.Logger) *WebhookProcessor {
	return &WebhookProcessor{
		service: service,
		logger:  logger,
	}
}

// ProcessLeadWebhook обрабатывает вебхук сделки
func (wp *WebhookProcessor) ProcessLeadWebhook(ctx context.Context, eventType string, data map[string]interface{}) error {
	wp.logger.Info("Processing lead webhook",
		zap.String("event_type", eventType),
		zap.Any("data", data))

	// Извлекаем ID сделки из данных вебхука
	// Структура вебхука может отличаться в зависимости от события
	var leadID int

	// Пробуем разные варианты структуры данных
	if leads, ok := data["leads"].(map[string]interface{}); ok {
		// Формат: {"leads": {"add": [{"id": 123, ...}]}}
		if addLeads, ok := leads[eventType].([]interface{}); ok && len(addLeads) > 0 {
			if lead, ok := addLeads[0].(map[string]interface{}); ok {
				if id, ok := lead["id"].(float64); ok {
					leadID = int(id)
				}
			}
		}
	} else if id, ok := data["id"].(float64); ok {
		// Простой формат: {"id": 123, ...}
		leadID = int(id)
	}

	if leadID == 0 {
		return fmt.Errorf("lead ID not found in webhook data")
	}

	// Получаем полные данные сделки
	lead, err := wp.service.GetLeadByID(ctx, leadID)
	if err != nil {
		return fmt.Errorf("failed to get lead data: %w", err)
	}

	// Обрабатываем в зависимости от типа события
	switch eventType {
	case "add":
		return wp.handleLeadAdd(ctx, lead)
	case "update":
		return wp.handleLeadUpdate(ctx, lead)
	case "delete":
		return wp.handleLeadDelete(ctx, leadID)
	case "status":
		return wp.handleLeadStatusChange(ctx, lead)
	case "responsible":
		return wp.handleLeadResponsibleChange(ctx, lead)
	default:
		wp.logger.Warn("Unknown webhook event type",
			zap.String("event_type", eventType))
	}

	return nil
}

func (wp *WebhookProcessor) handleLeadAdd(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("New lead added",
		zap.Int("lead_id", lead.ID),
		zap.String("name", lead.Name))

	// Получаем связанные контакты
	if lead.Embedded.Contacts != nil && len(lead.Embedded.Contacts) > 0 {
		// Получаем первый контакт
		contactID := lead.Embedded.Contacts[0]["id"]
		contact, err := wp.service.GetContactByID(ctx, contactID)
		if err != nil {
			wp.logger.Error("Failed to get contact",
				zap.Int("contact_id", contactID),
				zap.Error(err))
		} else {
			// Проверяем наличие телефона
			hasPhone := false
			phoneNumber := ""

			// Проходим по кастомным полям контакта
			for _, field := range contact.CustomFieldsValues {
				// Проверяем, является ли поле телефоном
				// В AmoCRM телефон обычно имеет определенный field_id или field_code
				if field.FieldCode == "PHONE" || field.FieldName == "Телефон" {
					if len(field.Values) > 0 {
						if value, ok := field.Values[0].(map[string]interface{}); ok {
							if phone, ok := value["value"].(string); ok && phone != "" {
								hasPhone = true
								phoneNumber = phone
								break
							}
						}
					}
				}
			}

			if hasPhone {
				wp.logger.Info("Lead has contact with phone",
					zap.Int("lead_id", lead.ID),
					zap.Int("contact_id", contact.ID),
					zap.String("phone", phoneNumber))

				// TODO: Отправить в Flow Engine для обработки
				eventData := map[string]interface{}{
					"event_type": "lead.add",
					"lead_id":    lead.ID,
					"lead_name":  lead.Name,
					"contact": map[string]interface{}{
						"id":    contact.ID,
						"name":  contact.Name,
						"phone": phoneNumber,
						"email": wp.extractEmail(contact),
					},
					"custom_fields": wp.extractCustomFields(lead),
				}

				// Публикуем событие для обработки Flow Engine
				// wp.publishEvent(ctx, eventData)
			}
		}
	}

	return nil
}

func (wp *WebhookProcessor) handleLeadUpdate(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("Lead updated",
		zap.Int("lead_id", lead.ID),
		zap.String("name", lead.Name))

	// TODO: Отправить в Flow Engine для обработки
	eventData := map[string]interface{}{
		"event_type":    "lead.update",
		"lead_id":       lead.ID,
		"lead_name":     lead.Name,
		"status_id":     lead.StatusID,
		"pipeline_id":   lead.PipelineID,
		"price":         lead.Price,
		"custom_fields": wp.extractCustomFields(lead),
	}

	// wp.publishEvent(ctx, eventData)

	return nil
}

func (wp *WebhookProcessor) handleLeadDelete(ctx context.Context, leadID int) error {
	wp.logger.Info("Lead deleted",
		zap.Int("lead_id", leadID))

	// TODO: Обработать удаление сделки
	eventData := map[string]interface{}{
		"event_type": "lead.delete",
		"lead_id":    leadID,
	}

	// wp.publishEvent(ctx, eventData)

	return nil
}

func (wp *WebhookProcessor) handleLeadStatusChange(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("Lead status changed",
		zap.Int("lead_id", lead.ID),
		zap.Int("status_id", lead.StatusID),
		zap.Int("pipeline_id", lead.PipelineID))

	// TODO: Отправить в Flow Engine для обработки
	eventData := map[string]interface{}{
		"event_type":    "lead.status",
		"lead_id":       lead.ID,
		"lead_name":     lead.Name,
		"status_id":     lead.StatusID,
		"pipeline_id":   lead.PipelineID,
		"custom_fields": wp.extractCustomFields(lead),
	}

	// wp.publishEvent(ctx, eventData)

	return nil
}

func (wp *WebhookProcessor) handleLeadResponsibleChange(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("Lead responsible changed",
		zap.Int("lead_id", lead.ID),
		zap.Int("responsible_user_id", lead.ResponsibleUserID))

	// TODO: Отправить в Flow Engine для обработки
	eventData := map[string]interface{}{
		"event_type":          "lead.responsible",
		"lead_id":             lead.ID,
		"lead_name":           lead.Name,
		"responsible_user_id": lead.ResponsibleUserID,
		"custom_fields":       wp.extractCustomFields(lead),
	}

	// wp.publishEvent(ctx, eventData)

	return nil
}

// extractCustomFields извлекает кастомные поля в удобном формате
func (wp *WebhookProcessor) extractCustomFields(lead *amocrm.Lead) map[string]interface{} {
	fields := make(map[string]interface{})

	for _, field := range lead.CustomFieldsValues {
		if len(field.Values) > 0 {
			// Для простоты берем первое значение
			if value, ok := field.Values[0].(map[string]interface{}); ok {
				fields[fmt.Sprintf("field_%d", field.FieldID)] = value["value"]
			}
		}
	}

	return fields
}

// extractEmail извлекает email из контакта
func (wp *WebhookProcessor) extractEmail(contact *amocrm.Contact) string {
	for _, field := range contact.CustomFieldsValues {
		if field.FieldCode == "EMAIL" || field.FieldName == "Email" {
			if len(field.Values) > 0 {
				if value, ok := field.Values[0].(map[string]interface{}); ok {
					if email, ok := value["value"].(string); ok {
						return email
					}
				}
			}
		}
	}
	return ""
}
