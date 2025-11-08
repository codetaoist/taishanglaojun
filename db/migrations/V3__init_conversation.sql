-- 初始化对话相关表（多租户）

-- 1. 对话表（租户级隔离）
CREATE TABLE IF NOT EXISTS tai_conversations (
  id VARCHAR(255) PRIMARY KEY,
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  title VARCHAR(255) NOT NULL,
  user_id VARCHAR(255) NOT NULL DEFAULT '',
  model_config JSONB,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. 消息表（租户级隔离）
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

-- 创建索引
-- 对话表索引
CREATE INDEX IF NOT EXISTS idx_tai_conversations_tenant_created_at ON tai_conversations(tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_tai_conversations_tenant_updated_at ON tai_conversations(tenant_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_tai_conversations_tenant_user_id ON tai_conversations(tenant_id, user_id);

-- 消息表索引
CREATE INDEX IF NOT EXISTS idx_tai_messages_tenant_conversation_id ON tai_messages(tenant_id, conversation_id);
CREATE INDEX IF NOT EXISTS idx_tai_messages_tenant_created_at ON tai_messages(tenant_id, created_at);
CREATE INDEX IF NOT EXISTS idx_tai_messages_tenant_role ON tai_messages(tenant_id, role);
CREATE INDEX IF NOT EXISTS idx_tai_messages_tenant_user_id ON tai_messages(tenant_id, user_id);

-- 行级安全策略（RLS）
ALTER TABLE tai_conversations ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_conversations_tenant_isolation ON tai_conversations
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE tai_messages ENABLE ROW LEVEL SECURITY;
CREATE POLICY tai_messages_tenant_isolation ON tai_messages
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

-- 触发器：自动更新 updated_at
CREATE TRIGGER trg_tai_conversations_updated_at
  BEFORE UPDATE ON tai_conversations
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_tai_messages_updated_at
  BEFORE UPDATE ON tai_messages
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();