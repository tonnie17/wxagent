package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/rag"
	"github.com/tonnie17/wxagent/pkg/tool"
	"html/template"
	"log/slog"
	"strings"
	"time"
)

var (
	ErrMemoryInUse = errors.New("memory in use")
	PromptFuncMap  = template.FuncMap{
		"now": time.Now,
	}
)

type Agent struct {
	config    *config.AgentConfig
	llm       llm.LLM
	memory    memory.Memory
	tools     []tool.Tool
	ragClient *rag.Client
}

func NewAgent(config *config.AgentConfig, llm llm.LLM, memory memory.Memory, tools []tool.Tool, ragClient *rag.Client) *Agent {
	return &Agent{
		config:    config,
		llm:       llm,
		memory:    memory,
		tools:     tools,
		ragClient: ragClient,
	}
}

func (a *Agent) Process(ctx context.Context, input string) (string, error) {
	if a.config.AgentTimeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, a.config.AgentTimeout)
		defer cancel()
		ctx = timeoutCtx
	}

	if a.ragClient != nil {
		documents, err := a.ragClient.Query(ctx, a.config.EmbeddingModel, input, 3)
		if err != nil {
			return "", err
		}

		if len(documents) > 0 {
			contexts := make([]string, 0, len(documents))
			for _, doc := range documents {
				contexts = append(contexts, doc.Content)
			}
			input = a.buildRAGPrompt(contexts, input)
		}
	}

	if l, ok := a.memory.(memory.Lock); ok {
		if l.Lock() {
			defer l.Release()
		} else {
			return "", ErrMemoryInUse
		}
	}

	var messages []*llm.ChatMessage
	if a.config.SystemPrompt != "" {
		systemPrompt := a.config.SystemPrompt
		if tmp, err := template.New("systemPrompt").Funcs(PromptFuncMap).Parse(systemPrompt); err == nil {
			promptTpl := new(bytes.Buffer)
			if tmp.Execute(promptTpl, nil) == nil {
				systemPrompt = promptTpl.String()
			}
		}
		messages = append(messages, &llm.ChatMessage{
			Role:    llm.RoleSystem,
			Content: systemPrompt,
		})
	}

	history, err := a.memory.History()
	if err != nil {
		return "", err
	}

	messages = append(messages, history...)
	messages = append(messages, &llm.ChatMessage{
		Role:    llm.RoleUser,
		Content: input,
	})

	toolsMap := make(map[string]tool.Tool, len(a.tools))
	for _, t := range a.tools {
		toolsMap[t.Name()] = t
	}

	var content string
	for i := 0; i < a.config.MaxToolIter; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		slog.Debug("chat input", slog.String("model", a.config.Model), slog.Any("messages", messages), slog.Any("tools", a.tools))
		msg, err := a.llm.Chat(ctx, a.config.Model, messages,
			llm.Tools(a.tools),
			llm.MaxTokens(a.config.MaxTokens),
			llm.Temperature(a.config.Temperature),
			llm.TopP(a.config.TopP),
		)
		if err != nil {
			slog.Debug("chat error", slog.Any("err", err))
			return "", err
		}
		slog.Debug("chat output", slog.Any("result", msg))

		messages = append(messages, msg)

		if len(msg.ToolCalls) == 0 {
			content = msg.Content
			break
		}

		for _, toolCall := range msg.ToolCalls {
			t, ok := toolsMap[toolCall.Name]
			if !ok {
				slog.Error("tool not exist", slog.String("tool_call_id", toolCall.ID))
				continue
			}

			var (
				toolCtx       = ctx
				toolCtxCancel context.CancelFunc
			)
			if a.config.ToolTimeout != 0 {
				toolCtx, toolCtxCancel = context.WithTimeout(ctx, a.config.ToolTimeout)
			}

			output, err := t.Execute(toolCtx, toolCall.Arguments)
			if err != nil {
				slog.Error("tool call function execute failed", slog.String("tool_call_id", toolCall.ID), slog.Any("err", err))
			}

			if toolCtxCancel != nil {
				toolCtxCancel()
			}

			messages = append(messages, a.convertToolCallMessage(toolCall.ID, output, err))
		}
	}

	a.memory.Update(messages)

	return content, nil
}

func (a *Agent) ProcessContinue(ctx context.Context) (string, error) {
	if l, ok := a.memory.(memory.Lock); ok && l.IsLocked() {
		return "", ErrMemoryInUse
	}

	messages, err := a.memory.History()
	if err != nil {
		return "", err
	}

	for i := len(messages) - 1; i > 0; i-- {
		msg := messages[i]
		if msg.Role == llm.RoleAssistant {
			return msg.Content, nil
		}
	}
	return "", nil
}

func (a *Agent) convertToolCallMessage(toolCallID string, output string, err error) *llm.ChatMessage {
	status := "success"
	if err != nil {
		status = "failed"
		output = err.Error()
	}

	content, _ := json.Marshal(struct {
		Status string `json:"status"`
		Output string `json:"output"`
	}{
		Status: status,
		Output: output,
	})

	toolMessage := &llm.ChatMessage{
		Role:       llm.RoleTool,
		Content:    string(content),
		ToolCallID: toolCallID,
	}

	return toolMessage
}

func (a *Agent) buildRAGPrompt(contexts []string, question string) string {
	return fmt.Sprintf(`You are an assistant. Answer the question based on the given context.

Context:
%v

Question:
%v

Answer:`, strings.Join(contexts, "\n"), question)
}
