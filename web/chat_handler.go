package web

import (
	"encoding/json"
	"fmt"
	"github.com/tonnie17/wxagent/pkg/agent"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/rag"
	"github.com/tonnie17/wxagent/pkg/tool"
	"io"
	"log/slog"
	"net/http"
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

func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notify := r.Context().Done()
	req := &struct {
		Model          string             `json:"model"`
		MaxInputTokens int                `json:"max_input_tokens"`
		MaxTokens      int                `json:"max_tokens"`
		Temperature    float32            `json:"temperature"`
		TopP           float32            `json:"top_p"`
		Messages       []*llm.ChatMessage `json:"messages"`
	}{}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxInputTokens int
	if req.MaxInputTokens == 0 {
		maxInputTokens = 500
	}
	mem := memory.NewTokenBase(req.Model, maxInputTokens)
	mem.Update(req.Messages)

	a := agent.NewAgent(&h.config.AgentConfig, llm.New(h.config.LLMProvider), mem, tool.GetTools(h.config.AgentTools), h.ragClient)
	out := make(chan *llm.ChatMessage)
	go func() {
		if err = a.ProcessStream(r.Context(), nil, out); err != nil {
			return
		}
	}()

	for {
		select {
		case <-notify:
			slog.Info("client disconnected")
			return
		case msg, ok := <-out:
			if !ok {
				slog.Info("client finished")
				return
			}
			if err != nil {
				fmt.Fprintf(w, "%s\n\n", err.Error())
				return
			}
			msgJSON, _ := json.Marshal(msg)
			fmt.Fprintf(w, "%s\n\n", string(msgJSON))
			flusher.Flush()
		}
	}
}
