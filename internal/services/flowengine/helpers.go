package flowengine

import (
	"context"
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

	// TODO: Implement actual dialer API call
	return nil
}

func (fe *FlowEngine) updateLead(ctx context.Context, actionData, inputData map[string]interface{}) error {
	leadID, _ := inputData["lead_id"].(float64)
	updateFields, _ := actionData["fields"].(map[string]interface{})

	fe.logger.Info("Updating lead",
		zap.Float64("lead_id", leadID),
		zap.Any("fields", updateFields))

	// TODO: Implement actual CRM API call
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
