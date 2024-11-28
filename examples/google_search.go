//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"github.com/tonni17/wxagent/pkg/agent"
	"github.com/tonni17/wxagent/pkg/config"
	"github.com/tonni17/wxagent/pkg/llm"
	"github.com/tonni17/wxagent/pkg/memory"
	"github.com/tonni17/wxagent/pkg/tool"
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

	ctx := context.Background()
	tools := []tool.Tool{
		tool.NewGoogleSearch(),
	}
	a := agent.NewAgent(&cfg.AgentConfig, llm.NewOpenAI(), memory.NewBuffer(6), tools)
	output, err := a.Process(ctx, "搜索一下法国的首都在哪里")
	fmt.Println(output, err)
}
