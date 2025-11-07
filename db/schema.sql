-- 分域数据模型（太上老君）
-- 约定：老君基础域表前缀 `lao_`，太上域表前缀 `tai_`

-- =====================
-- 老君基础域（laojun）
-- =====================

-- 插件主表（老君基础）
CREATE TABLE IF NOT EXISTS lao_plugins (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'inactive', -- inactive/active/installed
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lao_plugins_name ON lao_plugins(name);

-- 插件版本表（老君基础）
CREATE TABLE IF NOT EXISTS lao_plugin_versions (
  id SERIAL PRIMARY KEY,
  plugin_id VARCHAR(64) NOT NULL,
  version VARCHAR(64) NOT NULL,
  manifest JSONB,
  signature TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_lao_plugin_versions_plugin
    FOREIGN KEY(plugin_id)
    REFERENCES lao_plugins(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lao_plugin_versions_unique ON lao_plugin_versions(plugin_id, version);

-- 审计日志（老君基础）
CREATE TABLE IF NOT EXISTS lao_audit_logs (
  id SERIAL PRIMARY KEY,
  actor VARCHAR(64) NOT NULL,
  action VARCHAR(64) NOT NULL,
  target VARCHAR(64),
  payload JSONB,
  result VARCHAR(32) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_lao_audit_logs_actor_created_at ON lao_audit_logs(actor, created_at);

-- =====================
-- 太上域（taishang）
-- =====================

-- 模型注册（太上）
CREATE TABLE IF NOT EXISTS tai_models (
  id VARCHAR(64) PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  version VARCHAR(64),
  status VARCHAR(32) NOT NULL DEFAULT 'inactive', -- inactive/active/deprecated
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tai_models_name_provider ON tai_models(name, provider);

-- 向量集合（太上）
CREATE TABLE IF NOT EXISTS tai_vector_collections (
  id SERIAL PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  dims INT NOT NULL,
  index_type VARCHAR(64),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tai_vector_collections_name ON tai_vector_collections(name);

-- 任务编排（太上）
CREATE TABLE IF NOT EXISTS tai_tasks (
  id SERIAL PRIMARY KEY,
  type VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL, -- pending/running/succeeded/failed/canceled
  payload JSONB,
  result JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tai_tasks_status_created_at ON tai_tasks(status, created_at);