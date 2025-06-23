package flowengine

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type FlowEngine struct {
	logger *zap.Logger
}

type FlowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
}

type FlowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type FlowConfig struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

func NewFlowEngine(logger *zap.Logger) *FlowEngine {
	return &FlowEngine{
		logger: logger,
	}
}

func (fe *FlowEngine) ExecuteFlow(ctx context.Context, flowData json.RawMessage, inputData map[string]interface{}) (bool, error) {
	var config FlowConfig
	if err := json.Unmarshal(flowData, &config); err != nil {
		return false, fmt.Errorf("failed to unmarshal flow config: %w", err)
	}

	// Find start node
	var startNode *FlowNode
	for _, node := range config.Nodes {
		if node.Type == "start" {
			startNode = &node
			break
		}
	}

	if startNode == nil {
		return false, fmt.Errorf("no start node found")
	}

	// Execute flow from start node
	return fe.executeNode(ctx, startNode, &config, inputData)
}

func (fe *FlowEngine) executeNode(ctx context.Context, node *FlowNode, config *FlowConfig, data map[string]interface{}) (bool, error) {
	switch node.Type {
	case "start":
		// Find next node
		nextNode := fe.findNextNode(node.ID, config)
		if nextNode == nil {
			return true, nil
		}
		return fe.executeNode(ctx, nextNode, config, data)

	case "condition":
		result := fe.evaluateCondition(node.Data, data)

		// Find appropriate next node based on condition result
		var nextNodeID string
		for _, edge := range config.Edges {
			if edge.Source == node.ID {
				if (result && edge.Type == "true") || (!result && edge.Type == "false") {
					nextNodeID = edge.Target
					break
				}
			}
		}

		if nextNodeID == "" {
			return false, nil
		}

		nextNode := fe.findNodeByID(nextNodeID, config)
		if nextNode == nil {
			return false, fmt.Errorf("node not found: %s", nextNodeID)
		}

		return fe.executeNode(ctx, nextNode, config, data)

	case "action":
		// Execute action (e.g., send to dialer)
		if err := fe.executeAction(ctx, node.Data, data); err != nil {
			return false, err
		}

		// Continue to next node
		nextNode := fe.findNextNode(node.ID, config)
		if nextNode == nil {
			return true, nil
		}
		return fe.executeNode(ctx, nextNode, config, data)

	case "end":
		return true, nil

	default:
		return false, fmt.Errorf("unknown node type: %s", node.Type)
	}
}

func (fe *FlowEngine) evaluateCondition(nodeData, inputData map[string]interface{}) bool {
	field, _ := nodeData["field"].(string)
	operator, _ := nodeData["operator"].(string)
	value := nodeData["value"]

	inputValue, exists := inputData[field]
	if !exists {
		return false
	}

	switch operator {
	case "equals":
		return fmt.Sprintf("%v", inputValue) == fmt.Sprintf("%v", value)
	case "not_equals":
		return fmt.Sprintf("%v", inputValue) != fmt.Sprintf("%v", value)
	case "greater_than":
		return compareNumeric(inputValue, value, ">")
	case "less_than":
		return compareNumeric(inputValue, value, "<")
	case "contains":
		return contains(fmt.Sprintf("%v", inputValue), fmt.Sprintf("%v", value))
	default:
		return false
	}
}

func (fe *FlowEngine) findNextNode(nodeID string, config *FlowConfig) *FlowNode {
	for _, edge := range config.Edges {
		if edge.Source == nodeID {
			return fe.findNodeByID(edge.Target, config)
		}
	}
	return nil
}

func (fe *FlowEngine) findNodeByID(nodeID string, config *FlowConfig) *FlowNode {
	for _, node := range config.Nodes {
		if node.ID == nodeID {
			return &node
		}
	}
	return nil
}
