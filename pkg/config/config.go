package config

import (
	"github.com/caarlos0/env/v11"
	"time"
)

type Config struct {
	ServerAddr           string        `env:"SERVER_ADDR" envDefault:"0.0.0.0:8082"`
	WechatAppID          string        `env:"WECHAT_APP_ID"`
	WechatEncodingAESKey string        `env:"WECHAT_ENCODING_AES_KEY"`
	WechatToken          string        `env:"WECHAT_TOKEN"`
	WechatAllowList      []string      `env:"WECHAT_ALLOW_LIST"`
	WechatMemTTL         time.Duration `env:"WECHAT_MEM_TTL" envDefault:"5m"`
	WechatMemMsgSize     int           `env:"WECHAT_MEM_MSG_SIZE" envDefault:"6"`
	WechatTimeout        time.Duration `env:"WECHAT_TIMEOUT" envDefault:"4s"`
	LLMProvider          string        `env:"LLM_PROVIDER" envDefault:"openai"`
	UseRAG               bool          `env:"USE_RAG" envDefault:"false"`
	EmbeddingProvider    string        `env:"EMBEDDING_PROVIDER" envDefault:"openai"`
	KnowledgeBasePath    string        `env:"KNOWLEDGE_BASE_PATH" envDefault:"./knowledge_base"`
	AgentConfig
}

type AgentConfig struct {
	AgentTools     []string      `env:"AGENT_TOOLS"`
	AgentTimeout   time.Duration `env:"AGENT_TIMEOUT" envDefault:"30s"`
	MaxToolIter    int           `env:"MAX_TOOL_ITER" envDefault:"3"`
	ToolTimeout    time.Duration `env:"TOOL_TIMEOUT" envDefault:"10s"`
	Model          string        `env:"LLM_MODEL" envDefault:"gpt-3.5-turbo"`
	MaxTokens      int           `env:"LLM_MAX_TOKENS" envDefault:"500"`
	Temperature    float32       `env:"LLM_TEMPERATURE" envDefault:"0.95"`
	TopP           float32       `env:"LLM_TOP_P" envDefault:"0.5"`
	SystemPrompt   string        `env:"SYSTEM_PROMPT" envDefault:"当前时间: {{now.UTC}}"`
	EmbeddingModel string        `env:"EMBEDDING_MODEL" envDefault:"text-embedding-ada-002"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
