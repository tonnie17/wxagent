package api

import (
	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tonni17/wxagent/pkg/config"
	"github.com/tonni17/wxagent/web"
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

	router = chi.NewRouter()
	web.SetupRouter(router, cfg, logger)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
