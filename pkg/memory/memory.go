package memory

import "github.com/tonni17/wxagent/pkg/llm"

type Memory interface {
	Update(messages []*llm.ChatMessage)
	GetAllMessages() []*llm.ChatMessage
}
