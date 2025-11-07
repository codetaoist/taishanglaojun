-- 初始化太上域表与索引（向量、模型、任务，含多租户、RLS、分区）

-- 启用向量扩展（pgvector）
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. 模型表（多租户）
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
CREATE INDEX IF NOT EXISTS idx_tai_models_tenant_status ON tai_models(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_tai_models_provider_gin ON tai_models USING gin (provider gin_trgm_ops);

-- 2. 向量集合表（绑定模型）
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
CREATE INDEX IF NOT EXISTS idx_tai_vector_collections_tenant_model ON tai_vector_collections(tenant_id, model_id);

-- 3. 向量数据表（分区 + 向量索引）
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

-- 向量索引（cosine 示例）
CREATE INDEX IF NOT EXISTS idx_tai_vectors_embedding_cosine
  ON tai_vectors USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100);

-- 4. 任务表（分区 + 状态机）
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

CREATE INDEX IF NOT EXISTS idx_tai_tasks_tenant_status_created ON tai_tasks(tenant_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tai_tasks_worker_status ON tai_tasks(worker_id, status);

-- 5. 行级安全策略（RLS）
ALTER TABLE tai_models ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_models_tenant_isolation ON tai_models
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

-- 6. 触发器：自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tai_models_updated_at
  BEFORE UPDATE ON tai_models
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_tai_tasks_updated_at
  BEFORE UPDATE ON tai_tasks
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();