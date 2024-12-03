package embedding

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/tonnie17/wxagent/pkg/provider"
)

type OpenAI struct {
	client *openai.Client
}

func NewOpenAI() Model {
	openaiConfig := openai.DefaultConfig(provider.GetAPIKey("openai"))
	openaiConfig.BaseURL = provider.GetAPIBaseURL("openai")
	client := openai.NewClientWithConfig(openaiConfig)
	return &OpenAI{
		client: client,
	}
}

func (o *OpenAI) CreateEmbeddings(ctx context.Context, model string, content string) ([]float32, error) {
	resp, err := o.client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input:          []string{content},
		Model:          openai.EmbeddingModel(model),
		EncodingFormat: openai.EmbeddingEncodingFormatFloat,
	})

	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	return resp.Data[0].Embedding, nil
}
