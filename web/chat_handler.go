package web

import (
	"encoding/json"
	"github.com/tonnie17/wxagent/pkg/agent"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/rag"
	"github.com/tonnie17/wxagent/pkg/tool"
	"io"
	"net/http"
	"time"
)

type ChatHandler struct {
	config    *config.Config
	ragClient *rag.Client
}

func NewChatHandler(config *config.Config, ragClient *rag.Client) *ChatHandler {
	return &ChatHandler{
		config:    config,
		ragClient: ragClient,
	}
}

func (h *ChatHandler) Completions(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeResponse(w, http.StatusInternalServerError, nil, err)
		return
	}

	req := &struct {
		Model          string             `json:"model"`
		MaxInputTokens int                `json:"max_input_tokens"`
		MaxTokens      int                `json:"max_tokens"`
		Temperature    float32            `json:"temperature"`
		TopP           float32            `json:"top_p"`
		Messages       []*llm.ChatMessage `json:"messages"`
	}{}
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeResponse(w, http.StatusInternalServerError, nil, err)
		return
	}

	var maxInputTokens int
	if req.MaxInputTokens == 0 {
		maxInputTokens = 500
	}
	mem := memory.NewTokenBase(req.Model, maxInputTokens)
	mem.Update(req.Messages)

	agentCfg := h.config.AgentConfig
	if req.Model != "" {
		agentCfg.Model = req.Model
	}
	if req.MaxTokens > 0 {
		agentCfg.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		agentCfg.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		agentCfg.TopP = req.TopP
	}

	a := agent.NewAgent(&agentCfg, llm.New(h.config.LLMProvider), mem, tool.GetTools(h.config.AgentTools), h.ragClient)
	msg, err := a.Process(r.Context(), nil)
	if err != nil {
		h.writeResponse(w, http.StatusInternalServerError, nil, err)
		return
	}

	resp := struct {
		Choices []struct {
			Message      *llm.ChatMessage `json:"message"`
			FinishReason string           `json:"finish_reason"`
			Index        int              `json:"index"`
		} `json:"choices"`
		Created int64 `json:"created"`
	}{
		Choices: []struct {
			Message      *llm.ChatMessage `json:"message"`
			FinishReason string           `json:"finish_reason"`
			Index        int              `json:"index"`
		}{
			{
				Message:      msg,
				FinishReason: "stop",
				Index:        0,
			},
		},
		Created: time.Now().Unix(),
	}

	h.writeResponse(w, http.StatusOK, resp, nil)
}

func (h *ChatHandler) writeResponse(w http.ResponseWriter, statusCode int, resp interface{}, err error) {
	if err != nil {
		resp = map[string]interface{}{
			"error": map[string]interface{}{
				"message": err.Error(),
			},
		}
	}
	respJSON, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(respJSON)
}
