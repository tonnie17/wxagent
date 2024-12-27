//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"github.com/tonnie17/wxagent/pkg/agent"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/embedding"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/rag"
	"log/slog"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("config load failed", slog.Any("err", err))
		return
	}

	store, err := rag.NewPgVectorStore()
	if err != nil {
		slog.Error("init vector store failed", slog.Any("err", err))
		return
	}
	defer store.Release()

	client := rag.NewClient(embedding.NewOpenAI(), store)
	if err := client.BuildKnowledgeBase(context.Background(), cfg.KnowledgeBasePath, cfg.EmbeddingModel, false); err != nil {
		slog.Error("load failed", slog.Any("err", err))
		return
	}

	a := agent.NewAgent(&cfg.AgentConfig, llm.NewOpenAI(), memory.NewBuffer(6), nil, client)
	output, err := a.Chat(context.Background(), "摩洛哥的公路有多少公里")
	fmt.Println(output, err)
}
