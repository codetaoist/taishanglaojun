# AI（taishang-ai）接口规范文档

以 `/api/taishang/ai` 为前缀，定义AI推理、对话、嵌入等功能的接口、参数、响应与错误码。遵循统一响应包装 `{ code, message, data, traceId }` 与 OpenAPI 契约。

## 通用规范
- 认证：`Authorization: Bearer <JWT>`；作用域按资源粒度授权（ai:*, inference:*, chat:*）。
- 流式响应：支持Server-Sent Events (SSE)和WebSocket两种方式。
- 错误码：`OK | INVALID_ARGUMENT | UNAUTHENTICATED | PERMISSION_DENIED | NOT_FOUND | CONFLICT | FAILED_PRECONDITION | INTERNAL | UNAVAILABLE`。
- 版本与头部：`Accept: application/vnd.codetaoist.v1+json`；`X-Workspace-Id?`、`Idempotency-Key?`。
- 跨域通信：通过gRPC与taishang-core服务通信，获取模型和向量数据。

## 模块一：推理（Inference）
- 文本生成
  - `POST /ai/inference/text`
  - Body：`{ modelId, prompt, maxTokens?, temperature?, topP?, stream?(boolean) }`
  - 200（非流式）：`{"code":"OK","data":{"id":"inf-123","modelId":"m-123","text":"生成的文本内容","usage":{"promptTokens":10,"completionTokens":20,"totalTokens":30}}}`
  - 200（流式，SSE）：`data: {"id":"inf-123","text":"生成的","delta":"文本"}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"modelId":"m-123","prompt":"写一首关于春天的诗","maxTokens":100,"temperature":0.7}' \
      "$API/api/taishang/ai/inference/text"
    ```
- 嵌入生成
  - `POST /ai/inference/embedding`
  - Body：`{ modelId, input: string|string[], dimensions? }`
  - 200：`{"code":"OK","data":{"id":"emb-123","modelId":"m-456","embeddings":[[0.1,0.2,...],[0.3,0.4,...]],"usage":{"promptTokens":15,"totalTokens":15}}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"modelId":"m-456","input":["hello world","goodbye"]}' \
      "$API/api/taishang/ai/inference/embedding"
    ```

## 模块二：对话（Chat）
- 创建对话会话
  - `POST /ai/chat/sessions`
  - Body：`{ modelId, systemPrompt?, metadata? }`
  - 200：`{"code":"OK","data":{"sessionId":"chat-123","modelId":"m-123","createdAt":"2025-06-20T10:00:00Z"}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"modelId":"m-123","systemPrompt":"你是一个有帮助的助手"}' \
      "$API/api/taishang/ai/chat/sessions"
    ```
- 发送消息
  - `POST /ai/chat/sessions/{sessionId}/messages`
  - Body：`{ content, role?(user|assistant), stream?(boolean) }`
  - 200（非流式）：`{"code":"OK","data":{"id":"msg-456","sessionId":"chat-123","content":"AI回复内容","role":"assistant","createdAt":"2025-06-20T10:01:00Z"}}`
  - 200（流式，SSE）：`data: {"id":"msg-456","content":"AI","delta":"回复"}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"content":"你好，请介绍一下你自己","role":"user"}' \
      "$API/api/taishang/ai/chat/sessions/chat-123/messages"
    ```
- 获取对话历史
  - `GET /ai/chat/sessions/{sessionId}/messages`
  - Query：`limit?`, `before?`（消息ID）
  - 200：`{"code":"OK","data":{"messages":[{"id":"msg-123","content":"你好","role":"user","createdAt":"..."}]}}`
- 删除会话
  - `DELETE /ai/chat/sessions/{sessionId}`
  - 200：`{"code":"OK","data":{"deleted":true}}`

## 模块三：RAG（检索增强生成）
- RAG查询
  - `POST /ai/rag/query`
  - Body：`{ query, collectionId, topK?, modelId, rerank?(boolean), maxTokens? }`
  - 200：`{"code":"OK","data":{"id":"rag-789","answer":"基于检索内容的回答","sources":[{"id":"doc-1","content":"...","score":0.95}]}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"query":"如何备份数据库","collectionId":"vc-1","topK":5,"modelId":"m-123"}' \
      "$API/api/taishang/ai/rag/query"
    ```
- 添加文档
  - `POST /ai/rag/documents`
  - Body：`{ collectionId, documents:[{ content, metadata? }] }`
  - 200：`{"code":"OK","data":{"added":3,"ids":["doc-1","doc-2","doc-3"]}}`

## 模块四：工具调用（Tool Calling）
- 工具定义
  - `GET /ai/tools`
  - 200：`{"code":"OK","data":{"tools":[{"name":"search_web","description":"搜索网络","parameters":{"type":"object","properties":{"query":{"type":"string"}}}}]}}`
- 执行工具调用
  - `POST /ai/tools/execute`
  - Body：`{ toolName, parameters, modelId }`
  - 200：`{"code":"OK","data":{"id":"tool-123","result":"搜索结果","toolName":"search_web"}}`

## 模块五：微调（Fine-tuning）
- 创建微调任务
  - `POST /ai/fine-tuning/jobs`
  - Body：`{ modelId, trainingData, validationData?, hyperparameters? }`
  - 200：`{"code":"OK","data":{"id":"ft-123","modelId":"m-123","status":"QUEUED"}}`
- 获取微调任务状态
  - `GET /ai/fine-tuning/jobs/{jobId}`
  - 200：`{"code":"OK","data":{"id":"ft-123","status":"RUNNING","progress":0.45,"metrics":{"loss":0.23}}}`
- 取消微调任务
  - `POST /ai/fine-tuning/jobs/{jobId}/cancel`
  - 200：`{"code":"OK","data":{"id":"ft-123","status":"CANCELED"}}}`

## 错误码与示例
- `INVALID_ARGUMENT`：非法参数、模型ID不存在、输入内容过长
- `NOT_FOUND`：会话不存在、任务不存在
- `FAILED_PRECONDITION`：模型未加载、配额不足
- `RESOURCE_EXHAUSTED`：令牌配额不足、请求频率超限
- `UNAVAILABLE`：AI服务不可用、模型推理超时

错误响应示例：
```json
{"code":"RESOURCE_EXHAUSTED","message":"Token quota exceeded","traceId":"..."}
```

## 流式响应格式
- SSE格式：`data: {JSON数据}\n\n`
- WebSocket格式：JSON消息，包含`type`字段（`message`、`error`、`done`）

## 契约校验
- 以 OpenAPI 契约为事实源；新增/变更需通过契约测试与审查。
- 与taishang-core的gRPC接口需保持版本兼容性。
- 流式响应需提供完整的非流式等效接口用于测试。