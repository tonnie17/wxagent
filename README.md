# wxagent

让公众号瞬间变身为智能助手，支持特性：

- 接近零成本部署，只需要一个域名即可绑定到公众号
- 支持对话记忆，超时自动回复，以及对话结果回溯
- 支持多种工具调用，包括：获取天气，关键字搜索，网页总结，Home Assitant 设备控制

一键部署到 Vercel：

[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https%3A%2F%2Fgithub.com%2Ftonnie17%2Fwxagent&env=WECHAT_TOKEN,AGENT_TOOLS,LLM_MODEL,OPENAI_API_KEY,OPENAI_BASE_URL)

绑定域名到服务器地址，然后配置公众号服务器地址为：`{domain}/wechat/receive/`

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

</details>

## 配置

所有配置通过环境变量指定，支持通过`.env`文件自动加载：

- `WECHAT_TOKEN`：公众号服务器配置的令牌(Token)
- `AGENT_TOOLS`：Agent 可以使用的 Tools，用逗号分隔，需要配置相关的环境变量，支持：
    - `google_search`：Google 搜索
    - `get_weather`：天气查询
    - `webpage_summary`：网页文本总结
    - `get_devices`：获取 Home Assistant 设备列表
    - `execute_device`：执行 Home Assistant 设备动作
- `LLM_PROVIDER`：LLM 提供者，支持：`openai`，默认为`openai`
- `LLM_MODEL`：LLM 模型名称，默认为`gpt-3.5-turbo`
- `OPENAI_API_KEY`：OpenAI（兼容接口）API KEY
- `OPENAI_BASE_URL`：OpenAI（兼容接口）Base URL
- `SERVER_ADDR`：服务器模式的启动地址，默认为`localhost:8082`
- `AGENT_TIMEOUT`：Agent 对话超时时间，默认为`30s`
- `AGENT_MAX_TOOL_ITER`：Agent 调用工具最大迭代次数，默认为`3`
- `TOOL_TIMEOUT`：工具调用超时时间，默认为`10s`
- `LLM_MAX_TOKENS`：最大输出 Token 数量，默认为`500`
- `LLM_TEMPERATURE`：Temperature 参数，默认为`0.95`
- `LLM_TOP_P`：Top P 参数，默认为`0.5`
- `WECHAT_ALLOW_LIST`：允许交互的微信账号（openid），用逗号分隔，默认无限制
- `WECHAT_MEM_TTL`：公众号单轮对话记忆保存时间，默认为`5m`
- `WECHAT_MEM_MSG_SIZE`：公众号单轮对话记忆消息记录上限（包括工具消息），默认为`6`
- `WECHAT_TIMEOUT`：公众号单轮对话超时时间（公众号限制回复时间不能超过5秒），默认为`4s`



### 工具配置

#### Google 搜索（google_search）

- `GOOGLE_SEARCH_ENGINE`：Google 搜索引擎
- `GOOGLE_SEARCH_API_KEY`：Google 搜索 API Key



#### 天气查询（get_weather）

- `OPENWEATHERMAP_API_KEY`：OpenWeatherMap API Key



#### Home Assistant（get_devices，execute_device）

- `HA_BASE_URL`：Home Assistant 服务器地址
- `HA_BEARER_TOKEN`：Home Assistant API 验证 Bearer Token



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
