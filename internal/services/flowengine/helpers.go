package flowengine

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

func (fe *FlowEngine) executeAction(ctx context.Context, actionData, inputData map[string]interface{}) error {
	actionType, _ := actionData["type"].(string)

	switch actionType {
	case "send_to_dialer":
		return fe.sendToDialer(ctx, actionData, inputData)
	case "update_lead":
		return fe.updateLead(ctx, actionData, inputData)
	case "add_note":
		return fe.addNote(ctx, actionData, inputData)
	case "add_to_bucket":
		return fe.addToBucket(ctx, actionData, inputData)
	case "change_priority":
		return fe.changePriority(ctx, actionData, inputData)
	case "change_scheduler_step":
		return fe.changeSchedulerStep(ctx, actionData, inputData)
	case "remove_from_dialer":
		return fe.removeFromDialer(ctx, inputData)
	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}
}

func (fe *FlowEngine) sendToDialer(ctx context.Context, actionData, inputData map[string]interface{}) error {
	schedulerID, _ := actionData["scheduler_id"].(string)
	campaignID, _ := actionData["campaign_id"].(string)
	bucketID, _ := actionData["bucket_id"].(string)

	fe.logger.Info("Sending to dialer",
		zap.String("scheduler_id", schedulerID),
		zap.String("campaign_id", campaignID),
		zap.String("bucket_id", bucketID),
		zap.Any("data", inputData))

	// Подготавливаем контакт
	contact := map[string]interface{}{
		"phone": "",
		"name":  "",
		"email": "",
	}

	// Извлекаем данные контакта
	if contactData, ok := inputData["contact"].(map[string]interface{}); ok {
		if phone, ok := contactData["phone"].(string); ok {
			contact["phone"] = phone
		}
		if name, ok := contactData["name"].(string); ok {
			contact["name"] = name
		}
		if email, ok := contactData["email"].(string); ok {
			contact["email"] = email
		}
	}

	// Отправляем через NATS если подключен
	if fe.nc != nil {
		message := map[string]interface{}{
			"scheduler_id": schedulerID,
			"campaign_id":  campaignID,
			"bucket_id":    bucketID,
			"contact":      contact,
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := fe.nc.Publish("dialer.send_contact", data); err != nil {
			return fmt.Errorf("failed to publish to NATS: %w", err)
		}

		fe.logger.Info("Contact sent to dialer queue")
	}

	return nil
}

func (fe *FlowEngine) updateLead(ctx context.Context, actionData, inputData map[string]interface{}) error {
	leadID, _ := inputData["lead_id"].(float64)

	fe.logger.Info("Updating lead",
		zap.Float64("lead_id", leadID),
		zap.Any("action_data", actionData))

	// Подготавливаем данные для обновления
	updateData := map[string]interface{}{
		"lead_id": int(leadID),
	}

	// Обновление полей
	if fields, ok := actionData["fields"].(map[string]interface{}); ok {
		updateData["fields"] = fields
	}

	// Обновление статуса
	if statusID, ok := actionData["status_id"].(float64); ok && statusID > 0 {
		updateData["status_id"] = int(statusID)
	}

	// Обновление воронки
	if pipelineID, ok := actionData["pipeline_id"].(float64); ok && pipelineID > 0 {
		updateData["pipeline_id"] = int(pipelineID)
	}

	// Отправляем через NATS если подключен
	if fe.nc != nil {
		message := map[string]interface{}{
			"action": "update_lead",
			"data":   updateData,
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := fe.nc.Publish("crm.update_lead", data); err != nil {
			return fmt.Errorf("failed to publish to NATS: %w", err)
		}

		fe.logger.Info("Lead update request sent")
	}

	return nil
}

func (fe *FlowEngine) addToBucket(ctx context.Context, actionData, inputData map[string]interface{}) error {
	bucketID, _ := actionData["bucket_id"].(string)
	priority, _ := actionData["priority"].(float64)
	schedulerID, _ := actionData["scheduler_id"].(string)
	schedulerStep, _ := actionData["scheduler_step"].(float64)

	fe.logger.Info("Adding to bucket",
		zap.String("bucket_id", bucketID),
		zap.Float64("priority", priority),
		zap.String("scheduler_id", schedulerID),
		zap.Float64("scheduler_step", schedulerStep))

	// Подготавливаем контакт
	contact := map[string]interface{}{
		"phone":   "",
		"name":    "",
		"email":   "",
		"lead_id": inputData["lead_id"],
		"custom_data": map[string]interface{}{
			"amocrm_lead_id":    inputData["lead_id"],
			"amocrm_contact_id": inputData["contact_id"],
			"priority":          int(priority),
			"scheduler_step":    int(schedulerStep),
		},
	}

	// Извлекаем данные контакта
	if contactData, ok := inputData["contact"].(map[string]interface{}); ok {
		if phone, ok := contactData["phone"].(string); ok {
			contact["phone"] = phone
		}
		if name, ok := contactData["name"].(string); ok {
			contact["name"] = name
		}
		if email, ok := contactData["email"].(string); ok {
			contact["email"] = email
		}
	}

	// Добавляем custom fields из inputData
	if customFields, ok := inputData["custom_fields"].(map[string]interface{}); ok {
		customData := contact["custom_data"].(map[string]interface{})
		for k, v := range customFields {
			customData[k] = v
		}
	}

	// Отправляем через NATS если подключен
	if fe.nc != nil {
		message := map[string]interface{}{
			"action":         "add_to_bucket",
			"bucket_id":      bucketID,
			"priority":       int(priority),
			"scheduler_id":   schedulerID,
			"scheduler_step": int(schedulerStep),
			"contact":        contact,
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := fe.nc.Publish("dialer.add_to_bucket", data); err != nil {
			return fmt.Errorf("failed to publish to NATS: %w", err)
		}

		fe.logger.Info("Contact sent to bucket")
	}

	return nil
}

func (fe *FlowEngine) changePriority(ctx context.Context, actionData, inputData map[string]interface{}) error {
	priority, _ := actionData["priority"].(float64)
	leadID, _ := inputData["lead_id"].(float64)

	fe.logger.Info("Changing priority",
		zap.Float64("lead_id", leadID),
		zap.Float64("new_priority", priority))

	// Отправляем через NATS если подключен
	if fe.nc != nil {
		message := map[string]interface{}{
			"action":   "change_priority",
			"lead_id":  leadID,
			"priority": int(priority),
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := fe.nc.Publish("dialer.change_priority", data); err != nil {
			return fmt.Errorf("failed to publish to NATS: %w", err)
		}

		fe.logger.Info("Lead change priority sent")
	}

	return nil
}

func (fe *FlowEngine) changeSchedulerStep(ctx context.Context, actionData, inputData map[string]interface{}) error {
	schedulerStep, _ := actionData["scheduler_step"].(float64)
	leadID, _ := inputData["lead_id"].(float64)

	fe.logger.Info("Changing scheduler step",
		zap.Float64("lead_id", leadID),
		zap.Float64("new_scheduler_step", schedulerStep))

	// Отправляем через NATS если подключен
	if fe.nc != nil {
		message := map[string]interface{}{
			"action":         "change_scheduler_step",
			"lead_id":        leadID,
			"scheduler_step": int(schedulerStep),
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := fe.nc.Publish("dialer.change_scheduler_step", data); err != nil {
			return fmt.Errorf("failed to publish to NATS: %w", err)
		}

		fe.logger.Info("Lead change scheduler step sent")
	}

	return nil
}

func (fe *FlowEngine) removeFromDialer(ctx context.Context, inputData map[string]interface{}) error {
	leadID, _ := inputData["lead_id"].(float64)

	fe.logger.Info("Remove from dialer",
		zap.Float64("lead_id", leadID))

	// Отправляем через NATS если подключен
	if fe.nc != nil {
		message := map[string]interface{}{
			"action":  "remove_from_dialer",
			"lead_id": leadID,
		}

		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := fe.nc.Publish("dialer.remove_from_dialer", data); err != nil {
			return fmt.Errorf("failed to publish to NATS: %w", err)
		}

		fe.logger.Info("Lead remove from dialer sent")
	}

	return nil
}

func (fe *FlowEngine) addNote(ctx context.Context, actionData, inputData map[string]interface{}) error {
	leadID, _ := inputData["lead_id"].(float64)
	noteText, _ := actionData["text"].(string)

	fe.logger.Info("Adding note to lead",
		zap.Float64("lead_id", leadID),
		zap.String("note", noteText))

	// TODO: Implement actual CRM API call
	return nil
}

func compareNumeric(a, b interface{}, operator string) bool {
	aFloat, aErr := toFloat64(a)
	bFloat, bErr := toFloat64(b)

	if aErr != nil || bErr != nil {
		return false
	}

	switch operator {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}
