package memory

import (
	"github.com/pkoukk/tiktoken-go"
	"github.com/tonnie17/wxagent/pkg/llm"
	"log/slog"
)

type TokenBase struct {
	BaseLock
	messages    []*llm.ChatMessage
	maxTokens   int
	totalTokens int
	encoding    *tiktoken.Tiktoken
}

func NewTokenBase(model string, maxTokens int) *TokenBase {
	encoding, err := tiktoken.EncodingForModel(model)
	if err != nil {
		slog.Warn("failed to get encoding for model")
		encoding, _ = tiktoken.EncodingForModel("gpt-4o")
	}
	return &TokenBase{
		messages:    []*llm.ChatMessage{},
		maxTokens:   maxTokens,
		totalTokens: 0,
		encoding:    encoding,
	}
}

func (m *TokenBase) History() ([]*llm.ChatMessage, error) {
	return m.messages, nil
}

func (m *TokenBase) Update(messages []*llm.ChatMessage) error {
	m.totalTokens = m.getMessageTokens(messages)
	m.messages = messages
	m.truncate()
	return nil
}

func (m *TokenBase) truncate() {
	start := 0
	for start < len(m.messages) && m.maxTokens > 0 && m.getMessageTokens(m.messages) > m.maxTokens {
		start++
		for start < len(m.messages) && (m.messages[start].Role == llm.RoleTool || m.messages[start].Role == llm.RoleAssistant) {
			start++
		}
		m.messages = m.messages[start:]
	}
}

func (m *TokenBase) getMessageTokens(messages []*llm.ChatMessage) int {
	var total int
	for _, message := range messages {
		tokens := m.encoding.Encode(message.Content, nil, nil)
		total += len(tokens)
	}
	return total
}
