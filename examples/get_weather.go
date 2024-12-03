//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"github.com/tonnie17/wxagent/pkg/agent"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/tool"
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
		tool.NewGetWeather(),
	}
	a := agent.NewAgent(&cfg.AgentConfig, llm.NewOpenAI(), memory.NewBuffer(6), tools, nil)
	output, err := a.Process(ctx, "天气怎么样")
	fmt.Println(output, err)
	output, err = a.Process(ctx, "深圳")
	fmt.Println(output, err)
}
