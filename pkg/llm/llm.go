package llm

import (
	"context"
	"github.com/tonnie17/wxagent/pkg/tool"
)

type LLM interface {
	Chat(ctx context.Context, model string, messages []*ChatMessage, options ...ChatOption) (*ChatMessage, error)
}

func New(provider string) LLM {
	switch provider {
	case "openai":
		return NewOpenAI()
	default:
		return NewNotImplemented()
	}
}

type chatOptions struct {
	tools       []tool.Tool
	maxTokens   int
	temperature float32
	topP        float32
}

type ChatOption func(*chatOptions)

func Tools(tools []tool.Tool) ChatOption {
	return func(o *chatOptions) {
		o.tools = tools
	}
}

func MaxTokens(maxTokens int) ChatOption {
	return func(o *chatOptions) {
		o.maxTokens = maxTokens
	}
}

func Temperature(temperature float32) ChatOption {
	return func(o *chatOptions) {
		o.temperature = temperature
	}
}

func TopP(topP float32) ChatOption {
	return func(o *chatOptions) {
		o.topP = topP
	}
}
