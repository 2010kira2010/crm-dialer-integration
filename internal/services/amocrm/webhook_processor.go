package amocrm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/2010kira2010/amocrm"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type WebhookProcessor struct {
	service *Service
	logger  *zap.Logger
	nc      *nats.Conn
}

func NewWebhookProcessor(service *Service, logger *zap.Logger) *WebhookProcessor {
	return &WebhookProcessor{
		service: service,
		logger:  logger,
	}
}

// NewWebhookProcessorWithNATS создает процессор с поддержкой NATS
func NewWebhookProcessorWithNATS(service *Service, logger *zap.Logger, nc *nats.Conn) *WebhookProcessor {
	return &WebhookProcessor{
		service: service,
		logger:  logger,
		nc:      nc,
	}
}

// SetNATSConnection устанавливает соединение с NATS
func (wp *WebhookProcessor) SetNATSConnection(nc *nats.Conn) {
	wp.nc = nc
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
		if eventLeads, ok := leads[eventType].([]interface{}); ok && len(eventLeads) > 0 {
			if lead, ok := eventLeads[0].(map[string]interface{}); ok {
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
	contactData := map[string]interface{}{}
	hasContact := false

	if lead.Embedded != nil && lead.Embedded.Contacts != nil && len(lead.Embedded.Contacts) > 0 {
		// В библиотеке контакты представлены как []*FieldValues, где FieldValues это map[string]interface{}
		// Получаем первый контакт
		var contactID int

		// Проходим по контактам
		for _, contactField := range lead.Embedded.Contacts {
			// contactField имеет тип *FieldValues, которое является map[string]interface{}
			if idValue, ok := (*contactField)["id"]; ok {
				switch v := idValue.(type) {
				case float64:
					contactID = int(v)
				case int:
					contactID = v
				}
				break
			}
		}

		if contactID > 0 {
			contact, err := wp.service.GetContactByID(ctx, contactID)
			if err != nil {
				wp.logger.Error("Failed to get contact",
					zap.Int("contact_id", contactID),
					zap.Error(err))
			} else {
				hasContact = true
				// Проверяем наличие телефона
				phoneNumber := ""

				// Проходим по кастомным полям контакта
				if contact.CustomFieldsValues != nil {
					for _, field := range contact.CustomFieldsValues {
						// Проверяем, является ли поле телефоном
						// В AmoCRM телефон обычно имеет определенный field_code
						if field.FieldName == "Телефон" || field.FieldName == "Phone" {
							// Проверяем код поля, если он есть
							if codeStr, ok := field.FieldCode.(string); ok && (codeStr == "PHONE" || codeStr == "phone") {
								// Это точно телефон
							}

							if len(field.Values) > 0 && field.Values[0] != nil {
								// field.Values[0] имеет тип *CustomsFieldsValues
								if phoneValue := field.Values[0].Value; phoneValue != nil {
									if phone, ok := phoneValue.(string); ok && phone != "" {
										phoneNumber = phone
										break
									}
								}
							}
						}
					}
				}

				contactData = map[string]interface{}{
					"id":    contact.ID,
					"name":  contact.Name,
					"phone": phoneNumber,
					"email": wp.extractEmail(contact),
				}

				if phoneNumber != "" {
					wp.logger.Info("Lead has contact with phone",
						zap.Int("lead_id", lead.ID),
						zap.Int("contact_id", contact.ID),
						zap.String("phone", phoneNumber))
				}
			}
		}
	}

	// Создаем событие для Flow Engine
	eventData := map[string]interface{}{
		"event_type":          "lead.add",
		"lead_id":             lead.ID,
		"lead_name":           lead.Name,
		"status_id":           lead.StatusID,
		"pipeline_id":         lead.PipelineID,
		"price":               lead.Price,
		"responsible_user_id": lead.ResponsibleUserID,
		"created_at":          lead.CreatedAt,
		"custom_fields":       wp.extractCustomFields(lead),
		"has_contact":         hasContact,
	}

	// Добавляем данные контакта, если есть
	if hasContact {
		eventData["contact"] = contactData
	}

	// Публикуем событие для обработки Flow Engine
	if err := wp.PublishEvent(ctx, eventData); err != nil {
		wp.logger.Error("Failed to publish event", zap.Error(err))
		return fmt.Errorf("failed to publish lead.add event: %w", err)
	}

	return nil
}

func (wp *WebhookProcessor) handleLeadUpdate(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("Lead updated",
		zap.Int("lead_id", lead.ID),
		zap.String("name", lead.Name))

	// Создаем событие для Flow Engine
	eventData := map[string]interface{}{
		"event_type":          "lead.update",
		"lead_id":             lead.ID,
		"lead_name":           lead.Name,
		"status_id":           lead.StatusID,
		"pipeline_id":         lead.PipelineID,
		"price":               lead.Price,
		"responsible_user_id": lead.ResponsibleUserID,
		"updated_at":          lead.UpdatedAt,
		"custom_fields":       wp.extractCustomFields(lead),
	}

	// Публикуем событие для обработки Flow Engine
	if err := wp.PublishEvent(ctx, eventData); err != nil {
		wp.logger.Error("Failed to publish event", zap.Error(err))
		return fmt.Errorf("failed to publish lead.update event: %w", err)
	}

	return nil
}

func (wp *WebhookProcessor) handleLeadDelete(ctx context.Context, leadID int) error {
	wp.logger.Info("Lead deleted",
		zap.Int("lead_id", leadID))

	// Создаем событие для Flow Engine
	eventData := map[string]interface{}{
		"event_type": "lead.delete",
		"lead_id":    leadID,
		"deleted_at": time.Now().Unix(),
	}

	// Публикуем событие для обработки Flow Engine
	if err := wp.PublishEvent(ctx, eventData); err != nil {
		wp.logger.Error("Failed to publish event", zap.Error(err))
		return fmt.Errorf("failed to publish lead.delete event: %w", err)
	}

	return nil
}

func (wp *WebhookProcessor) handleLeadStatusChange(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("Lead status changed",
		zap.Int("lead_id", lead.ID),
		zap.Int("status_id", lead.StatusID),
		zap.Int("pipeline_id", lead.PipelineID))

	// Создаем событие для Flow Engine
	eventData := map[string]interface{}{
		"event_type":          "lead.status",
		"lead_id":             lead.ID,
		"lead_name":           lead.Name,
		"status_id":           lead.StatusID,
		"pipeline_id":         lead.PipelineID,
		"responsible_user_id": lead.ResponsibleUserID,
		"price":               lead.Price,
		"updated_at":          lead.UpdatedAt,
		"custom_fields":       wp.extractCustomFields(lead),
	}

	// Добавляем информацию о контактах, если есть
	if lead.Embedded != nil && lead.Embedded.Contacts != nil && len(lead.Embedded.Contacts) > 0 {
		contactIDs := []int{}
		for _, contactField := range lead.Embedded.Contacts {
			if idValue, ok := (*contactField)["id"]; ok {
				switch v := idValue.(type) {
				case float64:
					contactIDs = append(contactIDs, int(v))
				case int:
					contactIDs = append(contactIDs, v)
				}
			}
		}
		eventData["contact_ids"] = contactIDs
	}

	// Публикуем событие для обработки Flow Engine
	if err := wp.PublishEvent(ctx, eventData); err != nil {
		wp.logger.Error("Failed to publish event", zap.Error(err))
		return fmt.Errorf("failed to publish lead.status event: %w", err)
	}

	return nil
}

func (wp *WebhookProcessor) handleLeadResponsibleChange(ctx context.Context, lead *amocrm.Lead) error {
	wp.logger.Info("Lead responsible changed",
		zap.Int("lead_id", lead.ID),
		zap.Int("responsible_user_id", lead.ResponsibleUserID))

	// Создаем событие для Flow Engine
	eventData := map[string]interface{}{
		"event_type":          "lead.responsible",
		"lead_id":             lead.ID,
		"lead_name":           lead.Name,
		"responsible_user_id": lead.ResponsibleUserID,
		"status_id":           lead.StatusID,
		"pipeline_id":         lead.PipelineID,
		"updated_at":          lead.UpdatedAt,
		"custom_fields":       wp.extractCustomFields(lead),
	}

	// Публикуем событие для обработки Flow Engine
	if err := wp.PublishEvent(ctx, eventData); err != nil {
		wp.logger.Error("Failed to publish event", zap.Error(err))
		return fmt.Errorf("failed to publish lead.responsible event: %w", err)
	}

	return nil
}

// extractCustomFields извлекает кастомные поля в удобном формате
func (wp *WebhookProcessor) extractCustomFields(lead *amocrm.Lead) map[string]interface{} {
	fields := make(map[string]interface{})

	if lead.CustomFieldsValues != nil {
		for _, field := range lead.CustomFieldsValues {
			if len(field.Values) > 0 && field.Values[0] != nil {
				// Извлекаем значение
				value := field.Values[0].Value

				// Сохраняем по ID поля
				fields[fmt.Sprintf("field_%d", field.FieldID)] = value

				// Если есть код поля, сохраняем и по коду
				if codeStr, ok := field.FieldCode.(string); ok && codeStr != "" {
					fields[codeStr] = value
				}

				// Также сохраняем по имени поля для удобства
				if field.FieldName != "" {
					fields[field.FieldName] = value
				}
			}
		}
	}

	return fields
}

// extractEmail извлекает email из контакта
func (wp *WebhookProcessor) extractEmail(contact *amocrm.Contact) string {
	if contact.CustomFieldsValues != nil {
		for _, field := range contact.CustomFieldsValues {
			// Проверяем по имени поля
			if field.FieldName == "Email" || field.FieldName == "E-mail" || field.FieldName == "Электронная почта" {
				// Дополнительно проверяем код поля, если есть
				if codeStr, ok := field.FieldCode.(string); ok && (codeStr == "EMAIL" || codeStr == "email") {
					// Это точно email
				}

				if len(field.Values) > 0 && field.Values[0] != nil {
					if emailValue := field.Values[0].Value; emailValue != nil {
						if email, ok := emailValue.(string); ok && email != "" {
							return email
						}
					}
				}
			}
		}
	}
	return ""
}

// PublishEvent публикует событие в NATS для обработки Flow Engine
func (wp *WebhookProcessor) PublishEvent(ctx context.Context, eventData map[string]interface{}) error {
	if wp.nc == nil {
		wp.logger.Warn("NATS connection not initialized, skipping event publication")
		return nil // Не возвращаем ошибку, чтобы не блокировать обработку
	}

	// Определяем subject на основе типа события
	eventType, ok := eventData["event_type"].(string)
	if !ok {
		return fmt.Errorf("event_type not found in event data")
	}

	// Формируем subject для NATS
	// Например: webhooks.amocrm.lead_add
	subject := fmt.Sprintf("webhooks.amocrm.%s", strings.ReplaceAll(eventType, ".", "_"))

	// Добавляем метаданные
	eventData["timestamp"] = time.Now().Unix()
	eventData["source"] = "amocrm"

	// Сериализуем данные
	data, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Публикуем в NATS
	if err := wp.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish to NATS: %w", err)
	}

	wp.logger.Info("Event published to NATS",
		zap.String("subject", subject),
		zap.String("event_type", eventType),
		zap.Int("data_size", len(data)))

	// Для критически важных событий можно использовать NATS JetStream для гарантированной доставки
	// Пример:
	// js, _ := wp.nc.JetStream()
	// _, err = js.Publish(subject, data)

	return nil
}

// ProcessContactWebhook обрабатывает вебхук контакта
func (wp *WebhookProcessor) ProcessContactWebhook(ctx context.Context, eventType string, data map[string]interface{}) error {
	wp.logger.Info("Processing contact webhook",
		zap.String("event_type", eventType),
		zap.Any("data", data))

	// Извлекаем ID контакта
	var contactID int

	if contacts, ok := data["contacts"].(map[string]interface{}); ok {
		if eventContacts, ok := contacts[eventType].([]interface{}); ok && len(eventContacts) > 0 {
			if contact, ok := eventContacts[0].(map[string]interface{}); ok {
				if id, ok := contact["id"].(float64); ok {
					contactID = int(id)
				}
			}
		}
	}

	if contactID == 0 {
		return fmt.Errorf("contact ID not found in webhook data")
	}

	// Получаем полные данные контакта
	contact, err := wp.service.GetContactByID(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to get contact data: %w", err)
	}

	// Создаем событие
	eventData := map[string]interface{}{
		"event_type":   fmt.Sprintf("contact.%s", eventType),
		"contact_id":   contact.ID,
		"contact_name": contact.Name,
		"phone":        wp.extractPhone(contact),
		"email":        wp.extractEmail(contact),
	}

	// Публикуем событие
	if err := wp.PublishEvent(ctx, eventData); err != nil {
		wp.logger.Error("Failed to publish contact event", zap.Error(err))
		return fmt.Errorf("failed to publish contact event: %w", err)
	}

	return nil
}

// extractPhone извлекает телефон из контакта
func (wp *WebhookProcessor) extractPhone(contact *amocrm.Contact) string {
	if contact.CustomFieldsValues != nil {
		for _, field := range contact.CustomFieldsValues {
			if field.FieldName == "Телефон" || field.FieldName == "Phone" {
				if codeStr, ok := field.FieldCode.(string); ok && (codeStr == "PHONE" || codeStr == "phone") {
					// Это точно телефон
				}

				if len(field.Values) > 0 && field.Values[0] != nil {
					if phoneValue := field.Values[0].Value; phoneValue != nil {
						if phone, ok := phoneValue.(string); ok && phone != "" {
							return phone
						}
					}
				}
			}
		}
	}
	return ""
}
