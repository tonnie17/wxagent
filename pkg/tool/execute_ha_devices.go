package tool

import (
	"context"
	"encoding/json"
	"github.com/tonni17/wxagent/pkg/ha"
	"log/slog"
)

type ExecuteHADevice struct {
}

func NewExecuteHADevice() Tool {
	return &ExecuteHADevice{}
}

func (e *ExecuteHADevice) Name() string {
	return "execute_device"
}

func (e *ExecuteHADevice) Description() string {
	return "Use this function to execute service of devices in Home Assistant"
}

func (e *ExecuteHADevice) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"list": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"domain": map[string]interface{}{
							"type":        "string",
							"description": "The domain of the service",
						},
						"service": map[string]interface{}{
							"type":        "string",
							"description": "The service to be called",
						},
						"entity_id": map[string]interface{}{
							"type":        "string",
							"description": "The entity_id retrieved from available devices. It must start with domain, followed by dot character",
						},
					},
					"required": []string{"domain", "service", "entity_id"},
				},
			},
		},
	}
}

func (e *ExecuteHADevice) Execute(ctx context.Context, input string) (string, error) {
	var arguments struct {
		List []struct {
			Domain   string `json:"domain"`
			Service  string `json:"service"`
			EntityID string `json:"entity_id"`
		} `json:"list"`
	}
	if err := json.Unmarshal([]byte(input), &arguments); err != nil {
		slog.Error("unmarshal failed", slog.Any("err", err))
		return "", err
	}

	var entityStates []ha.EntityState
	for _, action := range arguments.List {
		states, err := ha.ExecuteService(ctx, action.Domain, action.Service, action.EntityID)
		if err != nil {
			return "", err
		}
		entityStates = append(entityStates, states...)
	}

	statesJSON, _ := json.Marshal(entityStates)

	return string(statesJSON), nil
}
