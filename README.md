# n8n to openai

一个兼容 OpenAI API 格式的聊天服务，支持将 n8n chat Trigger 转为 OpenAI 接口，已测试 Open WebUI 兼容性。

## 功能特性

- ✅ **聊天完成接口** - 支持流式和非流式响应
- ✅ **模型列表接口** - 返回支持的模型信息
- ✅ **API Key 认证** - 可选的认证中间件
- ✅ **流式传输** - 实时 Server-Sent Events (SSE) 支持
- ✅ **环境变量配置** - 灵活的配置管理

## 快速开始

### 环境要求

- Go 1.24.1 或更高版本

### 安装依赖

```bash
go mod download
```

### 运行服务

```bash
go run .
```

服务将在 `http://localhost:8080` 启动。

### 配置环境变量

- API_TOKEN：服务 API 鉴权，默认不开启
- PORT：服务端口，默认 8080
- GIN_MODE：日志打印级别 release/debug，默认 debug，debug 模式会打印更多日志
- MODELS: 模型配置以;符号做分隔符，例子：name1=n8n-webhook-url;name2=n8n-webhook-url2

_支持使用.env 文件配置环境变量_

## API 接口

### 1. 聊天完成接口

**端点：** `POST /v1/chat/completions`

**请求示例：**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: your-api-key" \
  -d '{
    "model": "gpt-5",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }'
```

**流式请求示例：**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: your-api-key" \
  -d '{
    "model": "gpt-5",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": true
  }'
```

**响应格式：**

```json
{
  "id": "1b01c3c2-742e-11f0-8e2b-8fce466710ef",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-5",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 20,
    "total_tokens": 30
  }
}
```

### 2. 模型列表接口

**端点：** `GET /v1/models`

**请求示例：**

```bash
curl -X GET http://localhost:8080/v1/models \
  -H "Authorization: your-api-key"
```

**响应格式：**

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-5",
      "object": "model",
      "created": 1677652288
    },
    {
      "id": "sonnet-4",
      "object": "model",
      "created": 1677652288
    },
    {
      "id": "sonnet-4-thinking",
      "object": "model",
      "created": 1677652288
    }
  ]
}
```

## 流式响应

当设置 `"stream": true` 时，服务器将返回 Server-Sent Events (SSE) 格式的流式数据。

## 配置选项

### 环境变量

| 变量名      | 默认值 | 描述                 |
| ----------- | ------ | -------------------- |
| `API_TOKEN` | -      | API 认证令牌（可选） |
| `PORT`      | `8080` | 服务器端口           |

### 认证

如果设置了 `API_TOKEN` 环境变量，所有请求都需要在 `Authorization` 头部包含该令牌：

```bash
Authorization: your-api-token
```

## 项目结构

```
├── main.go # 主程序入口
├── chat_completions.go # 聊天完成接口实现
├── models.go # 模型列表接口实现
├── middleware.go # 中间件（认证等）
├── types.go # 数据结构定义
├── go.mod # Go 模块文件
```

## 开发

### 构建

```bash
go build -o cursor-api
```

### 运行测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
```

## 依赖项

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web 框架
- [godotenv](https://github.com/joho/godotenv) - 环境变量管理

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 支持

如果你遇到问题或有疑问，请创建 Issue 或联系维护者。
