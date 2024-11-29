package tool

import (
	"context"
	"fmt"
	"github.com/tonnie17/wxagent/pkg/ha"
	"strings"
)

type GetHADevices struct {
}

func NewGetHADevices() Tool {
	return &GetHADevices{}
}

func (g *GetHADevices) Name() string {
	return "get_devices"
}

func (g *GetHADevices) Description() string {
	return "Use this function to get devices in Home Assistant, including their state and entity_id"
}

func (g *GetHADevices) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}

func (g *GetHADevices) Execute(ctx context.Context, input string) (string, error) {
	entityStates, err := ha.GetEntityStates(ctx, g.defaultDomains())
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("An overview of the devices in this smart home:\n")
	builder.WriteString("```csv\n")
	builder.WriteString("entity_id,name,state\n")
	for _, entityState := range entityStates {
		builder.WriteString(fmt.Sprintf("%v,%v,%v\n", entityState.EntityID, entityState.Attributes.FriendlyName, entityState.State))
	}
	builder.WriteString("```\n")

	return builder.String(), nil
}

func (g *GetHADevices) defaultDomains() []string {
	return []string{
		"door",
		"lock",
		"occupancy",
		"motion",
		"climate",
		"light",
		"switch",
		"sensor",
		"speaker",
		"media_player",
		"temperature",
		"humidity",
		"battery",
		"tv",
		"remote",
		"light",
		"vacuum",
	}
}
