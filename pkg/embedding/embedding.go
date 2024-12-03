package embedding

import "context"

type Model interface {
	CreateEmbeddings(ctx context.Context, model string, content string) ([]float32, error)
}

func New(provider string) Model {
	switch provider {
	case "openai":
		return NewOpenAI()
	default:
		return NewNotImplemented()
	}
}
