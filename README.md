# wxagent

让公众号瞬间变身为智能助手，支持特性：

- 接近零成本部署，只需要一个域名即可绑定到公众号
- 支持对话记忆，超时自动回复，以及对话结果回溯
- 支持多种工具调用，包括：获取天气，关键字搜索，网页总结，Home Assistant 设备控制
- 支持构建本地 RAG 知识库进行读取和检索

一键部署到 Vercel：

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Ftonnie17%2Fwxagent&env=WECHAT_TOKEN,LLM_MODEL,OPENAI_API_KEY,OPENAI_BASE_URL)

1. 参考[配置](#配置)填写环境变量，完成部署
2. 拿到Vercel生成的默认域名进行访问，没问题的话会输出`OK`
3. 绑定域名到服务器地址，然后配置公众号服务器地址为：`{domain}/wechat/receive/`

## 示例

<details><summary>点击查看对话效果</summary>

基础对话：

<img width="516" alt="basic_dialogue" src="https://github.com/user-attachments/assets/2670b1bd-e55f-4fee-b71f-0e000ba2625e">

获取天气：

<img width="513" alt="get_weather" src="https://github.com/user-attachments/assets/35dba473-1090-462a-8293-b9edd309723b">

文章总结：

<img width="514" alt="webpage_summary" src="https://github.com/user-attachments/assets/b4dbf8bb-e121-4049-b904-b2dc223e875e">

信息搜索：

<img width="516" alt="google_search" src="https://github.com/user-attachments/assets/d7514c5c-5b05-4075-9380-a79c12ff910b">

知识库检索：

<img width="513" alt="knowledge_base" src="https://github.com/user-attachments/assets/7a8f1994-3d5d-43e8-9ec9-5204c1906231">


</details>

## 配置

所有配置通过环境变量指定，支持通过`.env`文件自动加载：

- `WECHAT_TOKEN`：公众号服务器配置的令牌(Token)
- `WECHAT_ALLOW_LIST`：允许交互的微信账号（openid），用逗号分隔，默认无限制
- `WECHAT_MEM_TTL`：公众号单轮对话记忆保存时间，默认为`5m`
- `WECHAT_MEM_MSG_SIZE`：公众号单轮对话记忆消息记录上限（包括工具消息），默认为`6`
- `WECHAT_TIMEOUT`：公众号单轮对话超时时间（公众号限制回复时间不能超过5秒），默认为`4s`
- `WECHAT_APP_ID`：公众号 AppID，安全模式下需要指定
- `WECHAT_ENCODING_AES_KEY`：公众号消息加解密密钥 (EncodingAESKey)，安全模式下需要指定
- `LLM_PROVIDER`：LLM 提供者，支持：`openai`，默认为`openai`
- `OPENAI_API_KEY`：OpenAI（兼容接口）API KEY
- `OPENAI_BASE_URL`：OpenAI（兼容接口）Base URL
- `SERVER_ADDR`：服务器模式的启动地址，默认为`0.0.0.0:8082`
- `USE_RAG`：是否开启 RAG 从知识库检索查询
- `EMBEDDING_PROVIDER`：文本嵌入模型提供者，支持：`openai`，默认为 `openai`
- `KNOWLEDGE_BASE_PATH`：本地知识库目录路径，目前支持文件格式：`txt`，默认为 `./knowledge_base`

### Agent 配置

- `AGENT_TOOLS`：Agent 可以使用的 Tools，用逗号分隔，需要配置相关的环境变量，支持：
  - `google_search`：Google 搜索
  - `get_weather`：天气查询
  - `webpage_summary`：网页文本总结
  - `get_devices`：获取 Home Assistant 设备列表
  - `execute_device`：执行 Home Assistant 设备动作
- `AGENT_TIMEOUT`：Agent 对话超时时间，默认为`30s`
- `MAX_TOOL_ITER`：Agent 调用工具最大迭代次数，默认为`3`
- `TOOL_TIMEOUT`：工具调用超时时间，默认为`10s`
- `LLM_MODEL`：LLM 模型名称，默认为`gpt-3.5-turbo`
- `LLM_MAX_TOKENS`：最大输出 Token 数量，默认为`500`
- `LLM_TEMPERATURE`：Temperature 参数，默认为`0.2`
- `LLM_TOP_P`：Top P 参数，默认为`0.9`
- `SYSTEM_PROMPT`：设置 Agent 对话的 System Prompt
- `EMBEDDING_MODEL`：文本嵌入模型，支持：`openai`，默认为 `openai`


### 工具配置

#### Google 搜索（google_search）

- `GOOGLE_SEARCH_ENGINE`：Google 搜索引擎
- `GOOGLE_SEARCH_API_KEY`：Google 搜索 API Key


#### 天气查询（get_weather）

- `OPENWEATHERMAP_API_KEY`：OpenWeatherMap API Key


#### Home Assistant（get_devices，execute_device）

- `HA_BASE_URL`：Home Assistant 服务器地址
- `HA_BEARER_TOKEN`：Home Assistant API 验证 Bearer Token


## 扩展使用

<details><summary>与Agent交互</summary>

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tonnie17/wxagent/pkg/agent"
	"github.com/tonnie17/wxagent/pkg/config"
	"github.com/tonnie17/wxagent/pkg/llm"
	"github.com/tonnie17/wxagent/pkg/memory"
	"github.com/tonnie17/wxagent/pkg/tool"
)

func main() {
	tools := []tool.Tool{
		tool.NewWebPageSummary(),
	}

	agent := agent.NewAgent(&config.AgentConfig{
		AgentTools:   []string{"webpage_summary"},
		AgentTimeout: 30 * time.Second,
		MaxToolIter:  3,
		ToolTimeout:  10 * time.Second,
		Model:        "qwen-plus",
		MaxTokens:    500,
		Temperature:  0.2,
		TopP:         0.9,
	}, llm.NewOpenAI(), memory.NewBuffer(6), tools, nil)

	output, err := agent.Chat(context.Background(), "总结一下：https://golangnote.com/golang/golang-stringsbuilder-vs-bytesbuffer")
	if err != nil {
		log.Fatalf("chat failed: %v", err)
	}

	fmt.Println(output)
}
```

</details>

<details><summary>自定义工具</summary>


要定义一个工具，需要实现`Tool`定义的接口：

```go
type Tool interface {
	Name() string
	Description() string
	Schema() map[string]interface{}
	Execute(context.Context, string) (string, error)
}
```

方法含义：
 - `Name()`：工具名称
 - `Description()`：工具描述，描述尽量清晰以便模型了解选择工具进行调用
 - `Schema() map[string]interface{}`：提供工具的参数描述以及定义
 - `Execute(context.Context, string) (string, error)`：工具的执行逻辑，接收模型输入，返回执行结果

</details>

<details><summary>接入新模型</summary>


要接入新的模型，需要实现`LLM`定义的接口：

```go
type LLM interface {
    Chat(context.Context, model string, messages []*ChatMessage, options ...ChatOption) (*ChatMessage, error)
}
```

</details>

## 本地开发

创建配置文件`.env`，将项目配置写入到文件：

```sh
echo "{CONFIG_KEY}={CONFIG_VALUE}" > .env 
```

拉取依赖：

```sh
go mod tidy
```

启动服务器：

```sh
go run cmd/server/main.go
```

启动交互式命令行：

```sh
go run cmd/cli/main.go
```

## 构建

### 本地构建

编译二进制文件：

```sh
go build -o server /cmd/server
```

运行：

```sh
./server
```

### Docker 构建

创建镜像：

```sh
docker build -t wxagent .
```

运行容器：

```sh
docker run -it --rm -p 8082:8082 --env-file .env wxagent
```
