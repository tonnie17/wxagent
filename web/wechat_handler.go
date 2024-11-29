package web

import (
	"context"
	"encoding/xml"
	"errors"
	"github.com/tonnie17/wxagent/pkg/agent"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/tool"
	"github.com/tonnie17/wxagent/pkg/wechat"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WechatHandler struct {
	config   *config.Config
	memStore *UserMemoryStore
}

func NewWechatHandler(config *config.Config, memStore *UserMemoryStore) *WechatHandler {
	return &WechatHandler{
		config:   config,
		memStore: memStore,
	}
}

func (h *WechatHandler) Receive(w http.ResponseWriter, r *http.Request) {
	signature := r.URL.Query().Get("signature")
	msgSignature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	echoStr := r.URL.Query().Get("echostr")

	if signature != wechat.Signature(h.config.WechatToken, timestamp, nonce) {
		slog.Error("signature check failed",
			slog.String("signature", signature),
			slog.String("timestamp", timestamp),
			slog.String("nonce", nonce),
			slog.String("echostr", echoStr),
		)
		http.Error(w, "signature check failed", http.StatusUnauthorized)
		return
	}

	if echoStr != "" {
		w.Write([]byte(echoStr))
		return
	}

	var reqMessage wechat.TextMessage
	if err := xmlParseRequest(r.Body, &reqMessage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if reqMessage.Encrypt != "" {
		if msgSignature != wechat.Signature(h.config.WechatToken, timestamp, nonce, reqMessage.Encrypt) {
			slog.Error("msg signature check failed",
				slog.String("signature", signature),
				slog.String("timestamp", timestamp),
				slog.String("nonce", nonce),
				slog.String("echostr", echoStr),
			)
			http.Error(w, "signature check failed", http.StatusUnauthorized)
			return
		}

		content, err := wechat.DecryptMsg(h.config.WechatAppID, reqMessage.Encrypt, h.config.WechatEncodingAESKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := xmlParseRequest(strings.NewReader(content), &reqMessage); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	slog.Info("receive req", slog.Any("req", reqMessage))

	if len(h.config.WechatAllowList) > 0 {
		var isAllow bool
		for _, userID := range h.config.WechatAllowList {
			if userID == reqMessage.FromUserName {
				isAllow = true
				break
			}
		}
		if !isAllow {
			http.Error(w, "access denied", http.StatusUnauthorized)
			return
		}
	}

	mem := h.memStore.GetOrNew(reqMessage.FromUserName, func() memory.Memory {
		return memory.NewBuffer(h.config.WechatMemMsgSize)
	})

	a := agent.NewAgent(&h.config.AgentConfig, llm.New(h.config.LLMProvider), mem, tool.GetTools(h.config.AgentTools))

	input := strings.TrimSpace(reqMessage.Content)
	result := make(chan string)
	go func() {
		switch reqMessage.MsgType {
		case wechat.MsgTypeText:
			var (
				output string
				err    error
			)
			if detectContinue(input) {
				if output, err = a.ProcessContinue(context.Background()); output == "" {
					output = getContinueEmptyHint()
				}
			} else {
				output, err = a.Process(context.Background(), input)
			}
			if err != nil {
				if errors.Is(err, agent.ErrMemoryInUse) {
					result <- getProcessingHint()
				} else {
					result <- err.Error()
				}
			} else {
				result <- output
			}
		}
		close(result)
	}()

	var content string
	ticker := time.NewTicker(h.config.WechatTimeout)
	select {
	case <-ticker.C:
		content = getContinueHint()
	case content = <-result:
	}

	now := time.Now().Unix()
	var respMessage wechat.TextMessage
	respMessage.FromUserName = reqMessage.ToUserName
	respMessage.ToUserName = reqMessage.FromUserName
	respMessage.MsgType = reqMessage.MsgType
	respMessage.Content = content
	respMessage.CreateTime = now

	resp, err := xml.Marshal(respMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if reqMessage.Encrypt != "" {
		encrypt, err := wechat.EncryptMsg(h.config.WechatAppID, string(resp), h.config.WechatEncodingAESKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		encryptMessage := &wechat.EncryptMessage{
			Encrypt:      encrypt,
			MsgSignature: wechat.Signature(h.config.WechatToken, nonce, encrypt, strconv.FormatInt(now, 10)),
			Timestamp:    now,
			Nonce:        nonce,
		}
		resp, err = xml.Marshal(encryptMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
