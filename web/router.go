package web

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/rag"
	"log/slog"
	"net/http"
	"strings"
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

	chatHandler := NewChatHandler(config, ragClient)
	r.With(apiKeyAuth(config.ChatAPIKEY)).Post("/chat/completions", chatHandler.Completions)
}

func apiKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string
			splits := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(splits) == 2 {
				token = splits[1]
			}
			if apiKey == "" || token != apiKey {
				http.Error(w, "miss api key", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
