# 后端架构（Go + Gin）

统一分域（laojun/taishang）接口层、服务层、数据访问层与中间件。

## 分层结构
- 接口层：Gin 路由与控制器；`/api/laojun`, `/api/taishang`。
- 服务层：领域服务（插件管理、模型管理、任务编排）。
- DAO 层：数据库访问与事务；遵循数据字典与迁移版本。
- 中间件：鉴权、请求日志、错误处理、限流、追踪、契约校验。

## 目录约定
- `internal/laojun/...`，`internal/taishang/...`：各域服务与DAO。
- `cmd/api/main.go`：入口；加载路由与中间件；读取配置。
- `pkg/middleware/...`：通用中间件；`pkg/contracts`：OpenAPI 契约校验。

## Gin 路由示意
```go
r := gin.New()
r.Use(mw.Recovery(), mw.RequestLog(), mw.Auth())
lao := r.Group("/api/laojun")
{
  lao.GET("/plugins", ctrl.ListPlugins)
  lao.POST("/plugins/install", ctrl.InstallPlugin)
  lao.POST("/plugins/:id/start", ctrl.StartPlugin)
  lao.POST("/plugins/:id/stop", ctrl.StopPlugin)
  lao.POST("/plugins/:id/upgrade", ctrl.UpgradePlugin)
  lao.DELETE("/plugins/:id", ctrl.UninstallPlugin)
  lao.GET("/audits", ctrl.ListAudits)
}
tai := r.Group("/api/taishang")
{
  tai.GET("/models", ctrl.ListModels)
  tai.POST("/models", ctrl.CreateModel)
  tai.GET("/vectors", ctrl.ListCollections)
  tai.POST("/vectors", ctrl.CreateCollection)
  tai.POST("/tasks", ctrl.SubmitTask)
  tai.GET("/tasks/:id", ctrl.GetTask)
  tai.POST("/tasks/:id/cancel", ctrl.CancelTask)
}
```

## 服务化分层与路由边界（初期规划）
- 服务职责：
  - `gateway`：JWT 鉴权、`X-Workspace-Id` 作用域、限流与熔断、统一错误包装与 traceId 注入、路由转发到域服务。
  - `laojun-api`：插件安装/启停/升级/卸载、审计日志、系统配置与权限校验；严格遵循 OpenAPI 契约与统一响应包装。
  - `taishang-api`：模型注册与版本、向量集合与检索、任务编排与状态机、成本与配额治理；长任务进度与观测指标暴露。
  - `jobs/*`：`vector-indexer`（索引构建与维护）、`task-engine`（编排执行与重试），可先与 `taishang-api` 同进程运行后再拆分。
- 路由前缀：
  - `/api/laojun/*` 与 `/api/taishang/*`，保持域边界清晰；统一中间件（鉴权/包装/审计）。
- 中间件规范：
  - 统一响应包装 `{ code, message, data, traceId }`；错误码到 HTTP 状态映射（见接口标准）。
  - 鉴权：`Authorization: Bearer <JWT>`；资源级权限（RBAC）与工作空间作用域。
  - 审计：敏感写操作记录用户/动作/资源/结果/traceId；脱敏与合规。
- 代码生成与 DTO：
  - 使用 `oapi-codegen`（Go）基于 `openapi/*.yaml` 生成请求/响应类型与 handler 骨架；路由处理严格使用生成类型。
- 观测：
  - 指标（吞吐/延迟/错误率/队列长度）、结构化日志、分布式 trace；域与端点维度聚合。
- 演进路线：
  - M1：最小骨架（路由/中间件/健康检查）、契约生成与 stub 返回；驱动契约测试通过。
  - M2：拆分 `jobs` 与队列、完善鉴权与审计、错误码一致性与降级策略。
  - M3：前端联调、插件生态打通、端到端观测仪表盘。
- 事务与一致性
- 安装/升级/卸载插件：事务包裹；审计日志与版本变更一致写入。
- 任务编排：状态机驱动；重试与幂等键；结果写入与索引更新。

## 配置与启动
- 环境变量：DB、JWT 秘钥、限流参数、日志等级等。
- 启动Probe：`/healthz`、`/readyz`；CI 包含契约与路由检查。