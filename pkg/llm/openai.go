package llm

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/tonnie17/wxagent/pkg/tool"
)

type OpenAI struct {
	client *openai.Client
}

func NewOpenAI() LLM {
	openaiConfig := openai.DefaultConfig(getAPIKey("openai"))
	openaiConfig.BaseURL = getAPIBaseURL("openai")
	client := openai.NewClientWithConfig(openaiConfig)
	return &OpenAI{
		client: client,
	}
}

func (o *OpenAI) Chat(ctx context.Context, model string, chatMessages []*ChatMessage, options ...ChatOption) (*ChatMessage, error) {
	message := make([]openai.ChatCompletionMessage, 0, len(chatMessages))
	for _, chatMessage := range chatMessages {
		msg := openai.ChatCompletionMessage{
			Role:       string(chatMessage.Role),
			Content:    chatMessage.Content,
			ToolCallID: chatMessage.ToolCallID,
		}
		for _, toolCall := range chatMessage.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
				ID:   toolCall.ID,
				Type: openai.ToolType(toolCall.Type),
				Function: openai.FunctionCall{
					Name:      toolCall.Name,
					Arguments: toolCall.Arguments,
				},
			})
		}
		message = append(message, msg)
	}

	var option chatOptions
	for _, o := range options {
		o(&option)
	}

	requestTools := make([]openai.Tool, 0, len(option.tools))
	for _, t := range option.tools {
		requestTools = append(requestTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Schema(),
			},
		})
	}

	response, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    message,
		Tools:       requestTools,
		MaxTokens:   option.maxTokens,
		Temperature: option.temperature,
		TopP:        option.topP,
	})
	if err != nil {
		return nil, err
	}

	return o.convertResponse(response)
}

func (o *OpenAI) convertResponse(response openai.ChatCompletionResponse) (*ChatMessage, error) {
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	message := response.Choices[0].Message
	if message.Content != "" {
		rc := &ChatMessage{
			Role:    Role(message.Role),
			Content: message.Content,
		}
		return rc, nil
	}

	if len(message.ToolCalls) > 0 {
		rc := &ChatMessage{
			Role:    Role(message.Role),
			Content: message.Content,
		}
		for _, toolCall := range message.ToolCalls {
			rc.ToolCalls = append(rc.ToolCalls, &tool.Call{
				ID:        toolCall.ID,
				Type:      string(toolCall.Type),
				Name:      toolCall.Function.Name,
				Arguments: toolCall.Function.Arguments,
			})
		}
		return rc, nil
	}

	return nil, fmt.Errorf("no content or tool calls")
}
