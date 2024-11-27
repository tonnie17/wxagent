package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/tonni17/wxagent/pkg/config"
	"github.com/tonni17/wxagent/web"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("config load failed", slog.Any("err", err))
		return
	}

	r := chi.NewRouter()
	web.SetupRouter(r, cfg, logger)

	server := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}
	go func() {
		slog.Info("running on " + cfg.ServerAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http serve failed", slog.Any("err", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("server shutdown...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Info("server shutdown failed")
	}
}
