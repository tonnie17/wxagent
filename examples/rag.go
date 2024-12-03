//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvector "github.com/pgvector/pgvector-go/pgx"
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

	pool, err := InitPool()
	if err != nil {
		slog.Error("init pool failed", slog.Any("err", err))
		return
	}
	defer pool.Close()

	store, err := rag.NewPgVectorStore()
	if err != nil {
		slog.Error("init vector store failed", slog.Any("err", err))
		return
	}

	client := rag.NewClient(embedding.NewOpenAI(), store)
	if err := client.LoadData(context.Background(), cfg.KnowledgeBasePath, cfg.EmbeddingModel); err != nil {
		slog.Error("load failed", slog.Any("err", err))
		return
	}

	a := agent.NewAgent(&cfg.AgentConfig, llm.NewOpenAI(), memory.NewBuffer(6), nil, client)
	output, err := a.Process(context.Background(), "摩洛哥的公路有多少公里")
	fmt.Println(output, err)
}

func InitPool() (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("POSTGRES_URL"))
	if err != nil {
		return nil, err
	}
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvector.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
