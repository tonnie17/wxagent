package llm

import "github.com/tonni17/wxagent/pkg/tool"

type Role string

const (
	RoleUser      Role = "user"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
	RoleAssistant Role = "assistant"
)

type ChatMessage struct {
	Role       Role         `json:"role"`
	Content    string       `json:"content"`
	ToolCallID string       `json:"tool_call_id"`
	ToolCalls  []*tool.Call `json:"tool_calls"`
}
