package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"github.com/tonnie17/wxagent/pkg/config"
	"log/slog"
	"time"
)

func SetupRouter(r chi.Router, config *config.Config, logger *slog.Logger) {
	r.Use(slogchi.New(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	memStore := NewUserMemoryStore(config.WechatMemTTL)
	memStore.CheckAndClear(time.Second)

	wechatHandler := NewWechatHandler(config, memStore)
	r.Get("/wechat/receive", wechatHandler.Receive)
	r.Post("/wechat/receive", wechatHandler.Receive)
}
