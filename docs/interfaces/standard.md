# 接口规范（统一契约与跨域通信）

本文档定义了混合架构（Go + Python）下的统一接口规范，包括响应包装、分页、安全契约、幂等性、观测审计以及跨语言通信协议。所有域（laojun/taishang/taishang-ai）均需遵循此规范，确保前后端一致性和跨域协作。

## 统一响应包装
- 结构：`{ code: number, data: any, message: string }`。
- 约定：成功 `code=0`；错误码统一于平台错误码表，域端映射并返回包装。
- 错误码分类（示例）：
  - 通用：`1000 参数错误`、`1001 鉴权失败`、`1002 权限不足`、`1003 资源不存在`、`1004 并发冲突`。
  - Laojun：`2000 清单校验失败`、`2001 签名不合法`、`2002 资源配额不足`、`2003 生命周期异常`。
  - Taishang：`3000 索引构建失败`、`3001 嵌入异常`、`3002 任务状态不合法`、`3003 成本配额不足`。

## 分页与列表
- 结构：`{ items: [], page: number, pageSize: number, total: number }`。
- 查询参数：`page`, `pageSize`, `sortBy`, `order`, `filters`。

## 安全契约（OpenAPI）
- `securitySchemes` 示例：
```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```
- 在受保护端点添加：
```yaml
security:
  - bearerAuth: []
```
- 标准请求头：
  - `Authorization: Bearer <token>`
  - `X-Request-Id: <uuid>`
  - `X-Workspace-Id: <id>`（可选，多租户/工作空间）

## 幂等与重试
- 幂等键：对安装/升级/任务创建等写操作使用 `Idempotency-Key`（请求头）实现幂等。
- 重试策略：对可重试错误（5xx/网络异常）执行指数退避；长任务通过进度流轮询或推送。

## 观测与审计
- 日志：请求与响应关键字段记录；敏感字段脱敏。
- 指标：吞吐、延迟、错误率；按域和端点维度聚合。
- 审计：敏感写操作记录用户、动作、资源、结果与 traceId。

> 参见：`openapi/laojun.yaml`、`openapi/taishang.yaml`、`docs/security/permission-model.md`、`docs/ops/ci-cd-pipeline.md`


## 错误码到 HTTP 状态映射（统一约定）
- 统一响应格式：所有端点返回统一包装
```json
{
  "code": 0,
  "message": "ok",
  "data": { },
  "traceId": "<uuid>"
}
```
- 成功约定：
  - 成功返回：`HTTP 200`（查询/更新）、`HTTP 201`（创建）；`code=0`；`message=ok`。
  - 长任务触发：`HTTP 202`（Accepted），`code=0`，`data` 包含 `taskId` 与初始状态。
- 错误码分段：
  - `1xxx` 请求与契约错误 → `HTTP 400`（Bad Request）
    - `1001` 参数缺失或非法；`1002` 契约校验失败；`1003` 不支持的媒体类型。
  - `2xxx` 鉴权与权限 → `HTTP 401/403/409`
    - `2001` 未认证（缺少/无效令牌）→ `401`
    - `2002` 令牌过期 → `401`
    - `2003` 禁止访问（越权）→ `403`
    - `2009` 工作空间冲突（切换/绑定异常）→ `409`
  - `3xxx` 资源状态与存在性 → `HTTP 404/409`
    - `3001` 资源不存在 → `404`
    - `3002` 资源已存在（唯一约束冲突）→ `409`
  - `4xxx` 业务规则错误 → `HTTP 422`（Unprocessable Entity）
    - `4001` 状态不允许；`4002` 配置不合法；`4003` 配额不足。
  - `5xxx` 系统与依赖错误 → `HTTP 500/502/503`
    - `5001` 内部错误 → `500`
    - `5002` 依赖服务错误（上游响应异常）→ `502`
    - `5003` 服务不可用（降级/熔断中）→ `503`
  - `7xxx` 老君域（插件）错误 → 依规则映射到 `4xx/5xx`
    - `7001` 插件安装失败；`7002` 插件启动失败；`7003` 插件签名校验失败。
  - `8xxx` 太上域（模型/向量/任务）错误 → 依规则映射到 `4xx/5xx`
    - `8001` 模型版本不兼容；`8002` 向量集合不存在/维度不匹配；`8003` 任务幂等冲突。

### HTTP 状态与错误码映射表（摘要）
- `400` ↔ `1xxx` 合同/参数错误
- `401` ↔ `2001/2002` 未认证/过期
- `403` ↔ `2003` 越权
- `404` ↔ `3001` 未找到
- `409` ↔ `3002/2009` 冲突
- `422` ↔ `4xxx` 业务规则
- `500` ↔ `5001` 内部错误
- `502` ↔ `5002` 依赖异常
- `503` ↔ `5003` 服务不可用

### 处理规范（服务端）
- 中间件统一设置：
  - 解析 `Authorization: Bearer <JWT>`，将 `traceId` 注入请求上下文与响应。
  - 失败时根据错误类型映射 HTTP 状态，并填充统一包装。
- 伪代码：
```go
func WriteUnified(w http.ResponseWriter, r *http.Request, err error, data any) {
  traceID := getTraceID(r.Context())
  if err == nil {
    status := http.StatusOK
    if isCreated(r) { status = http.StatusCreated }
    if isAccepted(r) { status = http.StatusAccepted }
    writeJSON(w, status, Response{Code: 0, Message: "ok", Data: data, TraceID: traceID})
    return
  }
  code, status := mapError(err) // 如上分段映射
  writeJSON(w, status, Response{Code: code, Message: err.Error(), Data: nil, TraceID: traceID})
}
```
- 统一性要求：
  - 所有端点（含错误）均返回统一包装；`/health` 亦返回包装并包含版本与时间戳。
  - 批量接口对部分失败以 `code != 0` 表示，并在 `data.failures[]` 列出失败项。

## 跨语言通信（gRPC协议）

### 服务间通信架构
- **Laojun域 (Go)**: 提供插件管理、审计、配置等核心功能
- **Taishang-core (Go)**: 提供模型、向量、任务管理功能
- **Taishang-AI (Python)**: 提供AI推理、对话、嵌入等AI功能
- **通信方式**: taishang-ai与taishang-core之间通过gRPC协议通信

### gRPC接口定义
- **协议定义**: 使用Protocol Buffers定义服务接口和数据结构
- **服务发现**: 通过服务注册中心发现和连接服务
- **负载均衡**: 支持多种负载均衡策略（轮询、一致性哈希等）
- **错误处理**: 统一的错误码映射和传播机制

### 关键gRPC服务接口
```protobuf
// 模型服务
service ModelService {
  rpc GetModel(GetModelRequest) returns (ModelResponse);
  rpc ListModels(ListModelsRequest) returns (ListModelsResponse);
}

// 向量服务
service VectorService {
  rpc QueryVector(QueryRequest) returns (QueryResponse);
  rpc UpsertVector(UpsertRequest) returns (UpsertResponse);
}

// AI推理服务
service InferenceService {
  rpc TextGeneration(TextGenerationRequest) returns (TextGenerationResponse);
  rpc EmbeddingGeneration(EmbeddingRequest) returns (EmbeddingResponse);
}
```

### 跨语言通信最佳实践
- **超时控制**: 设置合理的请求超时和重试策略
- **流式处理**: 对长时间运行的AI推理任务使用流式响应
- **认证传递**: 通过metadata传递认证信息和上下文
- **监控与追踪**: 实现分布式追踪，监控跨服务调用链路

### 错误处理与降级策略
- **服务降级**: 当AI服务不可用时，提供基础功能降级方案
- **熔断机制**: 实现熔断器模式，防止级联故障
- **重试策略**: 对临时性错误实现指数退避重试
- **错误映射**: 将gRPC错误码映射到统一的HTTP错误码