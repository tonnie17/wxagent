package tool

import "context"

var tools = make(map[string]Tool)

func init() {
	for _, tool := range []Tool{
		NewGetWeather(),
		NewGoogleSearch(),
		NewWebPageSummary(),
		NewGetHADevices(),
		NewExecuteHADevice(),
	} {
		tools[tool.Name()] = tool
	}
}

type Tool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	Execute(context.Context, string) (string, error)
}

type Call struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func DefaultTools() []Tool {
	res := make([]Tool, 0, len(tools))
	for _, tool := range tools {
		res = append(res, tool)
	}
	return res
}

func GetTools(names []string) []Tool {
	res := make([]Tool, 0, len(names))
	for _, name := range names {
		if _, ok := tools[name]; ok {
			res = append(res, tools[name])
		}
	}
	return res
}
