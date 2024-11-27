package main

import (
	"context"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/tonni17/wxagent/pkg/agent"
	"github.com/tonni17/wxagent/pkg/config"
	"github.com/tonni17/wxagent/pkg/llm"
	"github.com/tonni17/wxagent/pkg/memory"
	"github.com/tonni17/wxagent/pkg/tool"
	"log"
	"log/slog"
	"strings"
)

func main() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt: "> ",
	})

	if err != nil {
		log.Fatal("readline create failed:", err)
	}
	defer rl.Close()

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("config load failed", slog.Any("err", err))
		return
	}

	a := agent.NewAgent(cfg, llm.NewOpenAI(), memory.NewBuffer(6), tool.DefaultTools())
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		input := strings.TrimSpace(line)

		output, err := a.Process(context.Background(), input)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(output)
	}

}
