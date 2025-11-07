# 分域数据模型与索引（初期规划）
- 范围与目标：支持多租户（Workspace）隔离、可审计、可扩展的插件治理与模型/向量/任务编排；读写分离与索引优化优先保证关键路径性能。
- 多租户与安全：推荐 PostgreSQL 行级安全（RLS）+ `X-Workspace-Id` 注入；审计与敏感字段脱敏。

## 通用域（common）
- `workspaces(id, name, owner_id, status, created_at, updated_at)`；唯一索引 `name`；活跃工作空间缓存。
- `users(id, email, display_name, status, created_at, updated_at)`；唯一索引 `email`；外键到 `workspaces_users` 关系表。
- `roles(id, workspace_id, name, description)` 与 `permissions(id, key, description)`；`role_permissions(role_id, permission_id)`；RBAC。
- `audit_logs(id, workspace_id, actor_id, action, resource, result, trace_id, created_at)`；按时间分区与 `workspace_id` 索引。
- `kv_configs(id, workspace_id, scope, key, value_jsonb, updated_at)`；唯一 `(workspace_id, scope, key)`；GIN on `value_jsonb`。

## 老君域（laojun）
- `plugins(id, workspace_id, name, version, status, manifest_jsonb, created_at, updated_at)`
  - 唯一 `(workspace_id, name)`；`status` 索引；GIN on `manifest_jsonb`；审计写入 `audit_logs`。
- `plugin_instances(id, workspace_id, plugin_id, enabled, config_jsonb, created_at, updated_at)`
  - 外键 `plugin_id`；唯一 `(workspace_id, plugin_id)`；`enabled` 索引。
- `plugin_audits(id, workspace_id, plugin_id, action, actor_id, diff_jsonb, result, trace_id, created_at)`
  - 关键写操作均记录；GIN on `diff_jsonb`；按月分区。

## 太上域（taishang）
- 模型管理：
  - `models(id, workspace_id, name, provider, version, metadata_jsonb, status, created_at, updated_at)`；唯一 `(workspace_id, name, version)`；GIN on `metadata_jsonb`。
- 向量集合与向量：
  - `vector_collections(id, workspace_id, name, dimension, metric, backend, shard_count, created_at, updated_at)`
    - 唯一 `(workspace_id, name)`；`metric in ('cosine','l2','ip')`；`backend in ('pgvector','milvus','faiss')`。
  - `vectors(id, collection_id, external_id, embedding VECTOR(1536), payload_jsonb, created_at)`
    - 外键 `collection_id`；GIN on `payload_jsonb`；根据 `backend` 选择索引策略（见下文）。
- 任务与编排：
  - `tasks(id UUID, workspace_id, type, state, priority, dedup_key, attempts, scheduled_at, payload_jsonb, created_at, updated_at)`
    - 索引：`(workspace_id, state, priority DESC)`、唯一 `dedup_key`、`scheduled_at`；状态机与幂等。
  - `task_events(id, task_id, event, detail_jsonb, trace_id, created_at)`；按任务聚合与溯源。
- 成本与配额：
  - `quotas(workspace_id PK, monthly_quota, usage, resets_at, updated_at)`；唯一 `workspace_id`；超限触发限流与降级。

## 索引与存储策略
- JSONB：统一使用 GIN 索引加速查询；为常用键额外建表达式索引（如 `((payload_jsonb->>'type'))`）。
- 时间序列：`audit_logs`、`task_events` 按月分区；关键查询保留最近 N 月热数据，冷数据归档。
- 向量：
  - `pgvector`：
    - 建议 `hnsw`（0.5+）或 `ivfflat` 索引，根据数据规模与召回率需求选择。
    - `ivfflat` 适合批量入库后查询：`CREATE INDEX ON vectors USING ivfflat (embedding) WITH (lists=100);`。
    - `hnsw` 适合在线持续入库：`CREATE INDEX ON vectors USING hnsw (embedding) WITH (m=16, ef_construction=200);`。
  - `milvus/faiss`：将 `vectors` 作为元数据与指针存储，真正 embedding 由外部引擎管理；维护 `external_id` 与一致性。
- 读写分离：主库写、只读副本查询；队列/异步任务（如索引重建）在 `jobs/vector-indexer` 承担。

## 行级安全（RLS）示例（PostgreSQL）
```sql
-- 在连接建立或网关转发时注入工作空间作用域
SET app.workspace_id = '12345';

ALTER TABLE plugins ENABLE ROW LEVEL SECURITY;
CREATE POLICY p_workspace_isolation ON plugins
  USING (workspace_id = current_setting('app.workspace_id')::bigint);

ALTER TABLE vector_collections ENABLE ROW LEVEL SECURITY;
CREATE POLICY vc_workspace_isolation ON vector_collections
  USING (workspace_id = current_setting('app.workspace_id')::bigint);
```

## SQL 示例（PostgreSQL + pgvector）
```sql
-- 需要：CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS vector_collections (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  name TEXT NOT NULL,
  dimension INT NOT NULL,
  metric TEXT NOT NULL CHECK (metric IN ('cosine','l2','ip')),
  backend TEXT NOT NULL CHECK (backend IN ('pgvector','milvus','faiss')),
  shard_count INT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (workspace_id, name)
);
CREATE INDEX IF NOT EXISTS idx_vc_workspace ON vector_collections(workspace_id);

-- 注意：VECTOR(1536) 为示例维度；如需多维度可按集合分表或统一维度策略
CREATE TABLE IF NOT EXISTS vectors (
  id BIGSERIAL PRIMARY KEY,
  collection_id BIGINT NOT NULL REFERENCES vector_collections(id) ON DELETE CASCADE,
  external_id TEXT,
  embedding VECTOR(1536) NOT NULL,
  payload_jsonb JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_vectors_collection ON vectors(collection_id);
CREATE INDEX IF NOT EXISTS idx_vectors_payload ON vectors USING GIN (payload_jsonb);
-- 根据场景选择其一：
-- CREATE INDEX IF NOT EXISTS idx_vectors_embedding_ivfflat ON vectors USING ivfflat (embedding) WITH (lists = 100);
-- CREATE INDEX IF NOT EXISTS idx_vectors_embedding_hnsw ON vectors USING hnsw (embedding) WITH (m=16, ef_construction=200);

CREATE TABLE IF NOT EXISTS tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id BIGINT NOT NULL,
  type TEXT NOT NULL,
  state TEXT NOT NULL CHECK (state IN ('queued','running','succeeded','failed','cancelled')),
  priority SMALLINT NOT NULL DEFAULT 50 CHECK (priority BETWEEN 0 AND 100),
  dedup_key TEXT,
  attempts INT NOT NULL DEFAULT 0,
  scheduled_at TIMESTAMPTZ,
  payload_jsonb JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (workspace_id, dedup_key)
);
CREATE INDEX IF NOT EXISTS idx_tasks_ws_state_priority ON tasks(workspace_id, state, priority DESC);
CREATE INDEX IF NOT EXISTS idx_tasks_schedule ON tasks(scheduled_at);
```

## 迁移与保留策略
- 每次变更均以迁移脚本记录于 `db/migrations/*`；确保向后兼容与回滚脚本。
- 热数据（近 90 天）保留在线查询；冷数据归档到对象存储并保留查询索引摘要。
- 向量索引重建与 VACUUM/ANALYZE 由 `jobs/vector-indexer` 周期性执行。