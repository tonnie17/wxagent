package memory

import (
	"github.com/tonnie17/wxagent/pkg/llm"
)

type Buffer struct {
	BaseLock
	maxMessages int
	messages    []*llm.ChatMessage
}

func NewBuffer(maxMessages int) *Buffer {
	return &Buffer{
		maxMessages: maxMessages,
		messages:    []*llm.ChatMessage{},
	}
}

func (m *Buffer) History() ([]*llm.ChatMessage, error) {
	return m.messages, nil
}

func (m *Buffer) Update(messages []*llm.ChatMessage) error {
	m.messages = messages
	m.truncate()
	return nil
}

func (m *Buffer) truncate() {
	start := 0
	for start < len(m.messages) && len(m.messages) > m.maxMessages {
		start++
		for start < len(m.messages) && (m.messages[start].Role == llm.RoleTool || m.messages[start].Role == llm.RoleAssistant) {
			start++
		}
		m.messages = m.messages[start:]
	}
}
