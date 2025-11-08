-- 初始化太上域表与索引（向量、模型、任务，含多租户、RLS、分区）

-- 启用向量扩展（pgvector）
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. 太上域表（租户级隔离）
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

-- 2. 模型表（多租户）
CREATE TABLE IF NOT EXISTS tai_models (
  id VARCHAR(64) PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(128) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  version VARCHAR(64),
  status VARCHAR(32) NOT NULL DEFAULT 'inactive',
  meta JSONB,               -- 额外超参、能力描述
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, name, provider)
);

-- 3. 模型配置表（多租户）
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

-- 4. 向量集合表（绑定模型）
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

-- 5. 向量数据表（分区 + 向量索引）
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

-- 初始分区（可按需扩容）
CREATE TABLE IF NOT EXISTS tai_vectors_0 PARTITION OF tai_vectors FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE IF NOT EXISTS tai_vectors_1 PARTITION OF tai_vectors FOR VALUES WITH (MODULUS 4, REMAINDER 1);
CREATE TABLE IF NOT EXISTS tai_vectors_2 PARTITION OF tai_vectors FOR VALUES WITH (MODULUS 4, REMAINDER 2);
CREATE TABLE IF NOT EXISTS tai_vectors_3 PARTITION OF tai_vectors FOR VALUES WITH (MODULUS 4, REMAINDER 3);

-- 6. 任务表（分区 + 状态机）
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

-- 按月分区示例
CREATE TABLE IF NOT EXISTS tai_tasks_2024_01
  PARTITION OF tai_tasks
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 创建索引
-- 太上域表索引
CREATE INDEX IF NOT EXISTS idx_tai_domains_tenant_owner ON tai_domains(tenant_id, owner_id);
CREATE INDEX IF NOT EXISTS idx_tai_domains_tenant_active ON tai_domains(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_tai_domains_tenant_name ON tai_domains(tenant_id, name);

-- 模型表索引
CREATE INDEX IF NOT EXISTS idx_tai_models_tenant_status ON tai_models(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_tai_models_provider_gin ON tai_models USING gin (provider gin_trgm_ops);

-- 模型配置表索引
CREATE INDEX IF NOT EXISTS idx_tai_model_configs_tenant_name ON tai_model_configs(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_tai_model_configs_active_default ON tai_model_configs(is_active, is_default);
CREATE INDEX IF NOT EXISTS idx_tai_model_configs_tenant_model ON tai_model_configs(tenant_id, model_id);

-- 向量集合表索引
CREATE INDEX IF NOT EXISTS idx_tai_vector_collections_tenant_model ON tai_vector_collections(tenant_id, model_id);
CREATE INDEX IF NOT EXISTS idx_tai_vector_collections_tenant_name ON tai_vector_collections(tenant_id, name);

-- 向量索引（cosine 示例）
CREATE INDEX IF NOT EXISTS idx_tai_vectors_embedding_cosine
  ON tai_vectors USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100);

-- 任务表索引
CREATE INDEX IF NOT EXISTS idx_tai_tasks_tenant_status_created ON tai_tasks(tenant_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tai_tasks_worker_status ON tai_tasks(worker_id, status);

-- 行级安全策略（RLS）
ALTER TABLE tai_domains ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_domains_tenant_isolation ON tai_domains
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE tai_models ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_models_tenant_isolation ON tai_models
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE tai_model_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_model_configs_tenant_isolation ON tai_model_configs
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE tai_vector_collections ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_vector_collections_tenant_isolation ON tai_vector_collections
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE tai_vectors ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_vectors_tenant_isolation ON tai_vectors
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE tai_tasks ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_tasks_tenant_isolation ON tai_tasks
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

-- 触发器：自动更新 updated_at
CREATE TRIGGER trg_tai_domains_updated_at
  BEFORE UPDATE ON tai_domains
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_tai_models_updated_at
  BEFORE UPDATE ON tai_models
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_tai_model_configs_updated_at
  BEFORE UPDATE ON tai_model_configs
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_tai_tasks_updated_at
  BEFORE UPDATE ON tai_tasks
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();