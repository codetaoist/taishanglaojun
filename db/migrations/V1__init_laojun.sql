-- 初始化老君基础域表与索引（含多租户、RLS、审计字段）

-- 启用必要扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 统一审计字段函数
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 1. 用户表（租户级隔离）
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

-- 2. 会话表（租户级隔离）
CREATE TABLE IF NOT EXISTS lao_sessions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  user_id UUID NOT NULL REFERENCES lao_users(id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. 令牌黑名单表（租户级隔离）
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

-- 4. 系统配置表（租户级隔离）
CREATE TABLE IF NOT EXISTS lao_configs (
  key VARCHAR(128) NOT NULL,
  value JSONB,
  scope VARCHAR(64) NOT NULL DEFAULT 'system', -- system / tenant / plugin
  tenant_id VARCHAR(64) DEFAULT 'default',
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (key, tenant_id)
);

-- 5. 插件主表（租户级隔离）
CREATE TABLE IF NOT EXISTS lao_plugins (
  id VARCHAR(64) PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id VARCHAR(64) NOT NULL DEFAULT 'default',
  name VARCHAR(128) NOT NULL,
  description TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'inactive',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (tenant_id, name)
);

-- 6. 插件版本表（强外键 + 唯一版本）
CREATE TABLE IF NOT EXISTS lao_plugin_versions (
  id SERIAL PRIMARY KEY,
  plugin_id VARCHAR(64) NOT NULL,
  version VARCHAR(64) NOT NULL,
  manifest JSONB NOT NULL,
  signature TEXT,
  package_url TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_lao_plugin_versions_plugin
    FOREIGN KEY(plugin_id) REFERENCES lao_plugins(id) ON DELETE CASCADE,
  CONSTRAINT uq_lao_plugin_versions_plugin_version
    UNIQUE (plugin_id, version)
);

-- 7. 审计日志表（分区模板）
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

-- 按月分区示例（可脚本化自动创建）
CREATE TABLE IF NOT EXISTS lao_audit_logs_2024_01
  PARTITION OF lao_audit_logs
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 8. 老君域表（租户级隔离）
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

-- 创建索引
-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_lao_users_tenant_username ON lao_users(tenant_id, username);
CREATE INDEX IF NOT EXISTS idx_lao_users_tenant_email ON lao_users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_lao_users_tenant_role ON lao_users(tenant_id, role);
CREATE INDEX IF NOT EXISTS idx_lao_users_tenant_active ON lao_users(tenant_id, is_active);

-- 会话表索引
CREATE INDEX IF NOT EXISTS idx_lao_sessions_tenant_user ON lao_sessions(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_lao_sessions_tenant_token ON lao_sessions(tenant_id, token_hash);
CREATE INDEX IF NOT EXISTS idx_lao_sessions_tenant_expires ON lao_sessions(tenant_id, expires_at);

-- 令牌黑名单表索引
CREATE INDEX IF NOT EXISTS idx_lao_token_blacklist_tenant_token ON lao_token_blacklist(tenant_id, token_hash);
CREATE INDEX IF NOT EXISTS idx_lao_token_blacklist_tenant_user ON lao_token_blacklist(tenant_id, user_id);

-- 系统配置表索引
CREATE INDEX IF NOT EXISTS idx_lao_configs_tenant_scope ON lao_configs(tenant_id, scope);

-- 插件表索引
CREATE INDEX IF NOT EXISTS idx_lao_plugins_tenant_status ON lao_plugins(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_lao_plugins_name_gin ON lao_plugins USING gin (name gin_trgm_ops);

-- 插件版本表索引
CREATE INDEX IF NOT EXISTS idx_lao_plugin_versions_created_at ON lao_plugin_versions(created_at DESC);

-- 审计日志表索引
CREATE INDEX IF NOT EXISTS idx_lao_audit_logs_tenant_actor_created_at
  ON lao_audit_logs(tenant_id, actor, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lao_audit_logs_target
  ON lao_audit_logs(target_type, target_id, created_at DESC);

-- 老君域表索引
CREATE INDEX IF NOT EXISTS idx_lao_domains_tenant_owner ON lao_domains(tenant_id, owner_id);
CREATE INDEX IF NOT EXISTS idx_lao_domains_tenant_active ON lao_domains(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_lao_domains_tenant_name ON lao_domains(tenant_id, name);

-- 行级安全策略（RLS）
ALTER TABLE lao_users ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_users_tenant_isolation ON lao_users
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE lao_sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_sessions_tenant_isolation ON lao_sessions
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE lao_token_blacklist ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_token_blacklist_tenant_isolation ON lao_token_blacklist
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE lao_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_configs_tenant_isolation ON lao_configs
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE lao_plugins ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_plugins_tenant_isolation ON lao_plugins
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE lao_audit_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_audit_logs_tenant_isolation ON lao_audit_logs
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

ALTER TABLE lao_domains ENABLE ROW LEVEL SECURITY;
CREATE POLICY lao_domains_tenant_isolation ON lao_domains
  FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::varchar);

-- 触发器：自动更新 updated_at
CREATE TRIGGER trg_lao_users_updated_at
  BEFORE UPDATE ON lao_users
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_lao_sessions_updated_at
  BEFORE UPDATE ON lao_sessions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_lao_configs_updated_at
  BEFORE UPDATE ON lao_configs
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_lao_plugins_updated_at
  BEFORE UPDATE ON lao_plugins
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_lao_domains_updated_at
  BEFORE UPDATE ON lao_domains
  FOR EACH ROW EXECUTE FUNCTION update_updated_at();