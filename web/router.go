package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/rag"
	"log/slog"
	"net/http"
	"time"
)

func SetupRouter(r chi.Router, config *config.Config, logger *slog.Logger, ragClient *rag.Client) {
	r.Use(slogchi.New(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	memStore := NewUserMemoryStore(config.WechatMemTTL)
	memStore.CheckAndClear(time.Second)

	wechatHandler := NewWechatHandler(config, memStore, ragClient)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	r.Get("/wechat/receive", wechatHandler.Receive)
	r.Post("/wechat/receive", wechatHandler.Receive)
}
