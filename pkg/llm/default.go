package llm

import (
	"context"
	"fmt"
)

type NotImplemented struct {
}

func NewNotImplemented() LLM {
	return &NotImplemented{}
}

func (o *NotImplemented) Chat(ctx context.Context, model string, chatMessages []*ChatMessage, options ...ChatOption) (*ChatMessage, error) {
	return nil, fmt.Errorf("llm not implemented")
}
