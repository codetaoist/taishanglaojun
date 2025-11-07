# 数据模型设计（扩充版）

本设计将域拆分为「老君基础域（laojun）」与「太上域（taishang）」，并实施表前缀：老君基础功能表以 `lao_` 开头，太上相关表以 `tai_` 开头。[0]

## 设计原则
- 按域分表，明确数据边界与权限边界。
- 索引与约束在 SQL 中明确；审计与恢复策略完善。[0]
- 兼顾原生端的本地存储需求与同步模式。

## 老君基础域（laojun）
- 插件管理：
  - `lao_plugins`（插件主表）：`id`, `name`, `description`, `status`, `created_at`, `updated_at`
  - `lao_plugin_versions`（版本表）：`id`, `plugin_id`, `version`, `manifest`, `signature`, `created_at`
- 审计日志：
  - `lao_audit_logs`：`id`, `actor`, `action`, `target`, `payload`, `result`, `created_at`
- 系统配置：
  - `lao_configs`：`key`, `value`, `scope`, `updated_at`

## 太上域（taishang）
- 模型与向量：
  - `tai_models`（模型注册）：`id`, `name`, `provider`, `version`, `status`, `created_at`
  - `tai_vector_collections`（向量集合）：`id`, `name`, `dims`, `index_type`, `created_at`
- 任务编排：
  - `tai_tasks`：`id`, `type`, `status`, `payload`, `result`, `created_at`, `updated_at`

## 索引与约束建议
- `lao_plugins(name)` 唯一索引；`lao_plugin_versions(plugin_id, version)` 唯一索引。
- `lao_audit_logs(actor, created_at)` 复合索引；按时间分区（可选）。
- `tai_models(name, provider)` 唯一索引；`tai_vector_collections(name)` 唯一索引。
- `tai_tasks(status, created_at)` 复合索引；必要时分表。

## 原生端本地存储建议
- iOS/Android：使用轻量 KV（设置/令牌）+ 持久化（CoreData/Room）。
- 机器人/手表端：只存必要状态与缓存，避免长时写入。

## 模块拆分与实体
- 老君域（plugins/审计/配置）
  - 插件：`plugin(id, name, version, status, checksum, created_at, updated_at)`
  - 审计：`plugin_audit(id, plugin_id, action, actor, result, trace_id, created_at)`
  - 配置：`plugin_config(id, plugin_id, key, value, scope, created_at, updated_at)`
- 太上域（模型/向量/任务）
  - 模型：`model(id, name, version, status, family, quantization, params_json, created_at, updated_at)`
  - 向量集合：`vector_collection(id, name, dim, index_type, metric, replication, created_at, updated_at)`
  - 向量：`vector(id, collection_id, namespace, values, metadata_json, created_at, updated_at)`
  - 任务：`task(id, type, status, priority, progress, result_json, ttl, created_at, updated_at)`

## 约束与索引
- 主键/唯一性：`plugin(name, version)` 组合唯一；`model(name, version)` 组合唯一；`vector(id)` 主键；`task(id)` 主键。
- 事务与并发：向量 `upsert` 采用批量写 + 幂等键；任务状态机使用 `version` 字段或悲观锁防止并发写穿。
- 索引建议：
  - 向量集合：`(name)`、`(dim, index_type, metric)` 复合索引以支持检索；
  - 任务：`(status, created_at)`、`(type)` 辅助查询与队列扫描。

## 原生终端本地存储映射
- iOS：
  - 秘密/令牌：`Keychain`
  - 结构化缓存：`CoreData` 或 `SQLite`（只读索引 + 轻量写入）
  - KV/高速缓存：`NSUserDefaults`/`MMKV`
- Android：
  - 秘密/令牌：`EncryptedSharedPreferences`
  - 结构化缓存：`Room`
  - KV/高速缓存：`MMKV`
- Harmony/鸿蒙：
  - 秘密/令牌：系统凭证能力
  - 结构化缓存：`Preferences` + 轻量 `KV` 存储
  - 分布式能力：跨设备数据同步需显式开关 + 权限声明
- 桌面（macOS/Windows/Linux）：
  - 秘密：系统钥匙串或凭据管理器
  - 结构化缓存：本地 `SQLite`/`LevelDB`
  - KV：`INI/JSON` 配置 + 文件权限控制

## 跨端同步与冲突处理
- 同步策略：
  - 在线优先：请求成功即刷新本地缓存；失败保留离线队列重试。
  - 离线优先：关键路径可读本地缓存；上线后以 `ETag/If-None-Match` 增量拉取。
- 冲突解决：
  - 向量写入：以后端 `upsert` 为准，客户端仅缓冲写入队列；
  - 任务状态：以服务端状态机为准；本地仅作为只读镜像；
  - 配置变更：采用 `last-write-wins` + 审计日志回溯，避免复杂 CRDT。

## 迁移与版本化
- 表迁移：使用 `db/migrations/V1__init_lao.sql` 与 `V2__init_tai.sql` 版本化；回滚脚本与验收剧本见运维手册。
- 合同驱动：字段变更通过 OpenAPI 契约与接口规范先行评审。
- 兼容性：新增字段默认可空；移除字段需给出兼容适配与 deprecation 周期。

> 参考：[0] https://www.doubao.com/thread/w30e30a4dcadbb935