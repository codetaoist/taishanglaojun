# 太上（taishang）接口规范文档

以 `/api/taishang` 为前缀，定义模型（Models）/ 向量（Vectors）/ 任务（Tasks）模块的接口、参数、响应与错误码。遵循统一响应包装 `{ code, message, data, traceId }` 与 OpenAPI 契约。

## 通用规范
- 认证：`Authorization: Bearer <JWT>`；作用域按资源粒度授权（model:*, vector:*, task:*）。
- 分页：`page`, `pageSize`；响应 `{ total, page, pageSize, items }`。
- 过滤与排序：`status`, `name`, `created_at`；`sort=field:asc,created_at:desc`。
- 错误码：`OK | INVALID_ARGUMENT | UNAUTHENTICATED | PERMISSION_DENIED | NOT_FOUND | CONFLICT | FAILED_PRECONDITION | INTERNAL | UNAVAILABLE`。
- 版本与头部：`Accept: application/vnd.codetaoist.v1+json`；`X-Workspace-Id?`、`Idempotency-Key?`。

## 模块一：模型（Models）
- 注册模型
  - `POST /models`
  - Body：`{ name, version, family?, quantization?, params? }`
  - 200：`{"code":"OK","data":{"id":"m-123","name":"qwen","version":"2.5","status":"enabled"}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"name":"qwen","version":"2.5","family":"transformer","quantization":"int8","params":{"max_tokens":"4096"}}' \
      "$API/api/taishang/models"
    ```
- 列表模型
  - `GET /models`
  - Query：`status? (enabled|disabled)`, `name?`, `page`, `pageSize`
  - 200：`{"code":"OK","data":{"total":1,"items":[{"id":"m-123","name":"qwen","version":"2.5","status":"enabled"}]}}`
  - curl：
    ```bash
    curl -H "Authorization: Bearer $JWT" \
      "$API/api/taishang/models?page=1&pageSize=20&status=enabled"
    ```
- 获取模型详情
  - `GET /models/{id}`
  - 200：`{"code":"OK","data":{"id":"m-123","name":"qwen","versions":["2.5"],"status":"enabled","params":{"max_tokens":4096}}}`
- 禁用/启用模型
  - `POST /models/{id}/disable` / `POST /models/{id}/enable`
  - 200：`{"code":"OK","data":{"id":"m-123","status":"disabled"}}`

## 模块二：向量（Vectors）
- 创建集合
  - `POST /vectors/collections`
  - Body：`{ name, dim, indexType(HNSW|IVF|FLAT), metric(cosine|l2|dot), replication? }`
  - 200：`{"code":"OK","data":{"id":"vc-1","name":"docs","dim":1536,"indexType":"HNSW","metric":"cosine"}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"name":"docs","dim":1536,"indexType":"HNSW","metric":"cosine"}' \
      "$API/api/taishang/vectors/collections"
    ```
- 列表集合
  - `GET /vectors/collections`
  - Query：`name?`, `page`, `pageSize`
  - 200：`{"code":"OK","data":{"total":1,"page":1,"pageSize":20,"items":[{"id":"vc-1","name":"docs","dim":1536,"indexType":"HNSW","metric":"cosine"}]}}`
- Upsert 向量
  - `POST /vectors/collections/{id}/upsert`
  - Body：`{ namespace?, vectors:[{ id, values:[f32...], metadata? }] }`
  - 200：`{"code":"OK","data":{"upserted":1,"namespace":"default"}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"namespace":"default","vectors":[{"id":"v1","values":[0.1,0.2,0.3],"metadata":{"doc":"..."}}]}' \
      "$API/api/taishang/vectors/collections/vc-1/upsert"
    ```
- 查询向量
  - `POST /vectors/collections/{id}/query`
  - Body：`{ topK, query:{ text|string|embedding }, namespace? }`
  - 200：`{"code":"OK","data":{"matches":[{"id":"v1","score":0.93,"metadata":{"doc":"..."}}],"count":1}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"topK":3,"query":{"text":"how to backup"},"namespace":"default"}' \
      "$API/api/taishang/vectors/collections/vc-1/query"
    ```
- 删除向量
  - `POST /vectors/collections/{id}/delete`
  - Body：`{ ids:[string...], namespace? }`
  - 200：`{"code":"OK","data":{"deleted":2}}`

## 模块三：任务（Tasks）
- 提交任务
  - `POST /tasks`
  - Body：`{ type:string, payload:object, priority?(low|normal|high), ttl? }`
  - 200：`{"code":"OK","data":{"id":"t-1","type":"infer","status":"PENDING"}}`
  - curl：
    ```bash
    curl -X POST -H "Authorization: Bearer $JWT" -H "Content-Type: application/json" \
      -d '{"type":"infer","payload":{"text":"hello"},"priority":"high","ttl":600}' \
      "$API/api/taishang/tasks"
    ```
- 列表任务
  - `GET /tasks`
  - Query：`status? (PENDING|RUNNING|SUCCEEDED|FAILED|CANCELED)`, `type?`, `page`, `pageSize`
  - 200：`{"code":"OK","data":{"total":2,"items":[{"id":"t-1","type":"infer","status":"PENDING"}]}}`
- 获取任务详情
  - `GET /tasks/{id}`
  - 200：`{"code":"OK","data":{"id":"t-1","type":"infer","status":"PENDING"}}`
- 取消任务
  - `POST /tasks/{id}/cancel`
  - 200：`{"code":"OK","data":{"id":"t-1","status":"CANCELED"}}`

## 契约与枚举
- 模型状态：`enabled|disabled`
- 向量索引：`HNSW|IVF|FLAT`；距离度量：`cosine|l2|dot`
- 任务状态机：`PENDING→RUNNING→SUCCEEDED/FAILED|CANCELED`；失败可 `retry`
- 资源与限额：任务 `priority`（`low|normal|high`）；`ttl` 过期自动清理

## 错误码与示例
- `INVALID_ARGUMENT`：维度不匹配、非法参数范围、未知枚举
- `NOT_FOUND`：模型/集合/任务不存在
- `CONFLICT`：重复名称、重复提交（幂等键）
- `FAILED_PRECONDITION`：依赖未准备好（索引未构建、模型未启用）
- `UNAVAILABLE`：检索服务或队列不可用（熔断/限流）

错误响应示例：
```json
{"code":"FAILED_PRECONDITION","message":"index not built","traceId":"..."}
```

## 契约校验
- 以 OpenAPI 契约为事实源；新增/变更需通过契约测试与审查。
- 响应示例与枚举引用必须与契约保持一致，CI 中进行校验。