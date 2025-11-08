-- 分域数据模型（太上老君）
-- 约定：老君基础域表前缀 `lao_`，太上域表前缀 `tai_`

-- =====================
-- 老君基础域（laojun）
-- =====================

-- 用户表（老君基础）
CREATE TABLE IF NOT EXISTS lao_users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  full_name VARCHAR(100),
  avatar_url VARCHAR(255),
  role VARCHAR(20) NOT NULL DEFAULT 'user',
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, username)
);

-- 会话表（老君基础）
CREATE TABLE IF NOT EXISTS lao_sessions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  user_id UUID NOT NULL REFERENCES lao_users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 令牌黑名单表（老君基础）
CREATE TABLE IF NOT EXISTS lao_token_blacklist (
  id SERIAL PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  token_hash VARCHAR(255) NOT NULL,
  user_id UUID NOT NULL,
  reason VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  CONSTRAINT fk_lao_token_blacklist_user
    FOREIGN KEY(user_id) REFERENCES lao_users(id) ON DELETE CASCADE,
  UNIQUE (tenant_id, token_hash)
);

-- 系统配置表（老君基础）
CREATE TABLE IF NOT EXISTS lao_configs (
  key VARCHAR(128) NOT NULL,
  value JSONB,
  scope VARCHAR(64) NOT NULL DEFAULT 'system', -- system / tenant / plugin
  tenant_id VARCHAR(64) DEFAULT 'default',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (key, tenant_id)
);

-- 插件主表（老君基础）
CREATE TABLE IF NOT EXISTS lao_plugins (
  id VARCHAR(64) PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(128) NOT NULL,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'inactive', -- inactive/active/installed
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, name)
);

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
  id BIGSERIAL PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  actor VARCHAR(64) NOT NULL,
  actor_type VARCHAR(32) NOT NULL DEFAULT 'user', -- user / apikey / plugin
  action VARCHAR(64) NOT NULL,
  target_type VARCHAR(64), -- plugin / model / vector / task
  target_id VARCHAR(64),
  payload JSONB,
  result VARCHAR(32) NOT NULL, -- success / fail
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- 老君域表（老君基础）
CREATE TABLE IF NOT EXISTS lao_domains (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(100) NOT NULL,
  description TEXT,
  owner_id UUID NOT NULL REFERENCES lao_users(id) ON DELETE CASCADE,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, name)
);

-- =====================
-- 太上域（taishang）
-- =====================

-- 太上域表（太上）
CREATE TABLE IF NOT EXISTS tai_domains (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(100) NOT NULL,
  description TEXT,
  owner_id UUID NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, name)
);

-- 模型注册（太上）
CREATE TABLE IF NOT EXISTS tai_models (
  id VARCHAR(64) PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(128) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  version VARCHAR(64),
  status VARCHAR(32) NOT NULL DEFAULT 'inactive', -- inactive/active/deprecated
  meta JSONB,               -- 额外超参、能力描述
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, name, provider)
);

-- 模型配置（太上）
CREATE TABLE IF NOT EXISTS tai_model_configs (
  id SERIAL PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  model_id VARCHAR(64),
  name VARCHAR(255) NOT NULL,
  service_type VARCHAR(50) NOT NULL,
  endpoint VARCHAR(500) NOT NULL,
  api_key VARCHAR(500),
  max_tokens INTEGER DEFAULT 4096,
  temperature FLOAT DEFAULT 0.7,
  is_default BOOLEAN DEFAULT FALSE,
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tai_model_configs_model
    FOREIGN KEY(model_id) REFERENCES tai_models(id) ON DELETE SET NULL,
  UNIQUE (tenant_id, name)
);

-- 向量集合（太上）
CREATE TABLE IF NOT EXISTS tai_vector_collections (
  id SERIAL PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(128) NOT NULL,
  model_id VARCHAR(64) NOT NULL,
  dims INT NOT NULL CHECK (dims > 0 AND dims <= 1536),
  index_type VARCHAR(64) DEFAULT 'ivfflat', -- ivfflat / hnsw
  metric_type VARCHAR(16) DEFAULT 'cosine', -- cosine / l2 / ip
  extra_index_args JSONB,                   -- 构建参数
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tai_vector_collections_model
    FOREIGN KEY(model_id) REFERENCES tai_models(id) ON DELETE CASCADE,
  UNIQUE (tenant_id, name)
);

-- 向量数据（太上）
CREATE TABLE IF NOT EXISTS tai_vectors (
  id BIGSERIAL PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  collection_id INT NOT NULL,
  external_id VARCHAR(256), -- 业务主键
  embedding vector(1536) NOT NULL,
  metadata JSONB,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tai_vectors_collection
    FOREIGN KEY(collection_id) REFERENCES tai_vector_collections(id) ON DELETE CASCADE,
  UNIQUE (collection_id, external_id)
) PARTITION BY HASH (collection_id);

-- 任务编排（太上）
CREATE TYPE task_status AS ENUM ('pending','running','success','failed','cancelled');
CREATE TYPE task_priority AS ENUM ('low','normal','high','urgent');

CREATE TABLE IF NOT EXISTS tai_tasks (
  id BIGSERIAL PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  type VARCHAR(64) NOT NULL, -- index / train / batch_infer
  status task_status NOT NULL DEFAULT 'pending',
  priority task_priority NOT NULL DEFAULT 'normal',
  payload JSONB NOT NULL,    -- 输入参数
  result JSONB,              -- 输出结果或错误信息
  worker_id VARCHAR(128),    -- 执行节点
  started_at TIMESTAMP,
  finished_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- 对话管理（太上）
CREATE TABLE IF NOT EXISTS tai_conversations (
  id VARCHAR(255) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  title VARCHAR(255) NOT NULL,
  user_id VARCHAR(255) NOT NULL DEFAULT '',
  model_config JSONB,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 消息管理（太上）
CREATE TABLE IF NOT EXISTS tai_messages (
  id VARCHAR(255) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  conversation_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(255) NOT NULL DEFAULT '',
  role VARCHAR(50) NOT NULL,
  content TEXT NOT NULL,
  metadata JSONB,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tai_messages_conversation
    FOREIGN KEY (conversation_id) REFERENCES tai_conversations(id) ON DELETE CASCADE
);