# 太上域（taishang）深入设计

定位：AI能力域，负责模型注册、向量集合管理、任务编排与监控，路径 `/api/taishang/...`，表前缀 `tai_`。[0]

## 业务职责
- 模型注册与管理：模型来源、版本、状态；兼容性与退役策略。
- 向量集合：维度与索引类型；集合生命周期与性能策略。
- 任务编排：任务类型、状态流转、重试与回滚；监控与告警。

## 核心实体
- Model（模型）
  - 字段：`id`, `name`, `provider`, `version`, `status`, `created_at`
  - 约束：`name+provider` 唯一；状态 `inactive/active/deprecated`。
- VectorCollection（向量集合）
  - 字段：`id`, `name`, `dims`, `index_type`, `created_at`
  - 约束：`name` 唯一；`dims` 必填。
- Task（任务）
  - 字段：`id`, `type`, `status`, `payload`, `result`, `created_at`, `updated_at`
  - 索引：`status, created_at`；必要时分表。

## 数据落库
- 表：`tai_models`, `tai_vector_collections`, `tai_tasks`
- 索引：`name+provider` 唯一、`name` 唯一、`status+created_at` 复合。

## 接口与路由
- 模型：`GET/POST /api/taishang/models`, `PUT/DELETE /api/taishang/models/{id}`
- 向量集合：`GET/POST /api/taishang/vectors`, `PUT/DELETE /api/taishang/vectors/{id}`
- 任务：`GET/POST /api/taishang/tasks`, `PUT /api/taishang/tasks/{id}`（状态变更）

## 关键流程
- 模型注册：来源校验→版本策略→状态切换（启用/停用/退役）。
- 向量集合：创建→索引构建→导入→查询；兼顾性能与存储成本。
- 任务编排：挂起/运行/成功/失败/取消；失败采样与重试策略；结果幂等。

## 性能与资源
- 限流与队列：接口限流、异步处理；重试指数退避。
- 向量库接入：Milvus/Faiss；写入批量、查询并发与缓存。
- 监控：任务吞吐/失败率、索引构建耗时、模型响应时间。

## 规划总览
- 目标：提供可治理的模型/向量/任务编排域，实现数据管道、索引与检索、训练评估与作业管理的闭环。
- 契约：`/api/taishang/*`（models/vectors/tasks），统一响应包装与错误码映射。

## 业务流程（高层）
- 模型：注册 → 配置（资源/精度/费用）→ 上架/下架 → 更新/回滚。
- 向量：集合创建 → 数据摄入 → 切分与嵌入 → 索引构建 → 搜索/维护。
- 任务：作业创建 → 调度（优先级/配额/重试）→ 运行 → 监控与审计 → 终止/取消。

## 功能分解
- 模型管理：`register/list/update/delete`，版本语义化与兼容策略，缓存与路由。
- 向量管理：集合/分片/度量（cosine/L2/dot），`upsert/search/delete`，TTL 与压缩策略。
- 任务编排：类型（训练/嵌入/索引/评估），状态机（创建/队列/运行/完成/失败/取消），重试与回退。
- 数据管道：来源（文件/仓/API），清洗与分块（窗口/语义），嵌入器与模型选择，监控与告警。
- 成本与性能：预算与配额，并发/回压，冷热数据分层与缓存策略。
- 安全与合规：数据分级与脱敏，密钥管理与访问审计，跨域共享治理。

## 细节逻辑与约束
- 一致性：集合与索引构建需幂等；任务状态机禁止不合法跳转。
- 并发：写多读多场景下的锁与分片策略；后台合并与重建。
- 监控：吞吐/延时/召回/精度指标；阈值告警与自愈策略。
- 持久化：集合元数据、索引参数、任务轨迹；参见 `db/migrations/V2__init_tai.sql`。
- 错误码：统一错误码表，模型/向量/任务端点映射与响应包装返回。

## 接口契约映射
- 模型：`GET/POST /api/taishang/models`，`PATCH /api/taishang/models/{id}`，`DELETE /api/taishang/models/{id}`。
- 向量：`POST /api/taishang/vectors/collections`，`POST /api/taishang/vectors/upsert`，`POST /api/taishang/vectors/search`，`DELETE /api/taishang/vectors/collections/{id}`。
- 任务：`POST /api/taishang/tasks`，`GET /api/taishang/tasks/{id}`，`POST /api/taishang/tasks/{id}/cancel`，`GET /api/taishang/tasks?page=...`。

> 参考：`openapi/taishang.yaml`，`docs/data/data-model-design.md`，`docs/testing/strategy.md`

## 能力列表（What we do）
- 模型治理：注册/版本/状态与退役策略，动态路由与缓存。
- 向量检索：集合生命周期、嵌入/索引/搜索闭环，性能与成本治理。
- 任务编排：长任务状态机、优先级与配额、失败重试与降级、观测与告警。

## 实现路径（How we do it）
- 契约与生成：使用 `openapi/taishang.yaml` 作为事实源，生成客户端并驱动前后端联调；合同校验与 diff 接入 CI 门禁。
- 数据管道与嵌入：来源（文件/仓/API）→ 清洗与分块 → 嵌入器选择与模型切换 → 索引构建 → 搜索。
- 状态机与观测：任务 `create/queue/run/succeed/fail/cancel` 状态；进度、成本与吞吐指标；端到端 trace 与审计事件。
- 成本与性能：预算/配额/回压与并发控制；冷热数据分层与缓存策略；异常降级与切换（本地→远程模型）。

> 参阅：`docs/interfaces/taishang-api-spec.md`、`docs/testing/strategy.md`、`scripts/openapi_contract_diff.py`