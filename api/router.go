package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/embedding"
	"github.com/tonnie17/wxagent/pkg/rag"
	"github.com/tonnie17/wxagent/web"
	"log/slog"
	"net/http"
	"os"
)

var router chi.Router

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("config load failed", slog.Any("err", err))
		return
	}

	var ragClient *rag.Client
	if cfg.UseRAG {
		store, err := rag.NewPgVectorStore()
		if err != nil {
			slog.Error("init vector store failed", slog.Any("err", err))
			return
		}
		ragClient = rag.NewClient(embedding.New(cfg.EmbeddingProvider), store)
		if err := ragClient.LoadData(context.Background(), cfg.KnowledgeBasePath, cfg.EmbeddingModel); err != nil {
			slog.Error("load data failed", slog.Any("err", err))
			return
		}
	}

	router = chi.NewRouter()
	web.SetupRouter(router, cfg, logger, ragClient)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
