# 太上域模型服务API

太上域模型服务API提供了统一的接口来管理和调用多种AI模型，包括OpenAI、Hugging Face、Ollama等主流模型提供商。

## 功能特性

- 支持多种模型提供商（OpenAI、Hugging Face、Ollama等）
- 统一的模型配置管理
- 文本生成和流式文本生成
- 嵌入向量生成（单个和批量）
- 模型健康检查
- 模型服务注册和发现

## API端点

### 模型管理

#### 创建模型配置
```
POST /api/v1/taishang/models
```

请求体：
```json
{
  "name": "gpt-3.5-turbo",
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "apiKey": "your-api-key",
  "baseURL": "https://api.openai.com/v1",
  "enabled": true,
  "maxTokens": 2048,
  "temperature": 0.7,
  "description": "OpenAI GPT-3.5 Turbo模型",
  "timeout": 30
}
```

#### 获取模型列表
```
GET /api/v1/taishang/models
```

#### 获取模型详情
```
GET /api/v1/taishang/models/{id}
```

#### 更新模型配置
```
PUT /api/v1/taishang/models/{id}
```

#### 删除模型配置
```
DELETE /api/v1/taishang/models/{id}
```

#### 模型健康检查
```
GET /api/v1/taishang/models/{id}/health
```

#### 获取服务列表
```
GET /api/v1/taishang/models/services
```

### 文本生成

#### 生成文本
```
POST /api/v1/taishang/models/generate
```

请求体：
```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "system", "content": "你是一个有用的助手。"},
    {"role": "user", "content": "请介绍一下人工智能的发展历史。"}
  ],
  "maxTokens": 500,
  "temperature": 0.7,
  "stream": false
}
```

#### 流式生成文本
```
POST /api/v1/taishang/models/generate
```

请求体与上述相同，但设置`"stream": true`。响应将以Server-Sent Events (SSE)格式返回。

### 嵌入生成

#### 生成单个嵌入
```
POST /api/v1/taishang/models/embeddings
```

请求体：
```json
{
  "model": "text-embedding-ada-002",
  "text": "这是一个测试文本，用于生成嵌入向量。"
}
```

#### 批量生成嵌入
```
POST /api/v1/taishang/models/embeddings/batch
```

请求体：
```json
{
  "model": "text-embedding-ada-002",
  "texts": [
    "这是第一个测试文本。",
    "这是第二个测试文本。",
    "这是第三个测试文本。"
  ]
}
```

## 支持的模型提供商

### OpenAI
- 支持文本生成和嵌入生成
- 需要提供API密钥
- 可自定义API端点

### Hugging Face
- 支持文本生成和嵌入生成
- 支持流式响应
- 可使用Inference API或自托管实例

### Ollama
- 支持本地模型
- 支持文本生成和嵌入生成
- 无需API密钥

## 使用示例

### 创建OpenAI模型配置

```bash
curl -X POST http://localhost:8080/api/v1/taishang/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-3.5-turbo",
    "provider": "openai",
    "model": "gpt-3.5-turbo",
    "apiKey": "sk-your-api-key",
    "baseURL": "https://api.openai.com/v1",
    "enabled": true,
    "maxTokens": 2048,
    "temperature": 0.7,
    "description": "OpenAI GPT-3.5 Turbo模型",
    "timeout": 30
  }'
```

### 生成文本

```bash
curl -X POST http://localhost:8080/api/v1/taishang/models/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "system", "content": "你是一个有用的助手。"},
      {"role": "user", "content": "请介绍一下人工智能的发展历史。"}
    ],
    "maxTokens": 500,
    "temperature": 0.7,
    "stream": false
  }'
```

### 生成嵌入

```bash
curl -X POST http://localhost:8080/api/v1/taishang/models/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-ada-002",
    "text": "这是一个测试文本，用于生成嵌入向量。"
  }'
```

## 测试

运行测试脚本：

```bash
cd services/api
go run cmd/test/main.go
```

这将执行一系列API调用，测试模型服务的各项功能。

## 错误处理

API使用标准HTTP状态码表示请求状态：

- `200 OK`: 请求成功
- `201 Created`: 资源创建成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未授权访问
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

错误响应格式：
```json
{
  "error": {
    "code": "MODEL_NOT_FOUND",
    "message": "指定的模型不存在"
  }
}
```

## 注意事项

1. 请确保API密钥的安全性，不要在代码中硬编码密钥
2. 使用流式响应时，客户端需要正确处理Server-Sent Events格式
3. 不同模型提供商的参数可能略有不同，请参考相应提供商的文档
4. 建议为生产环境设置适当的超时时间和重试机制