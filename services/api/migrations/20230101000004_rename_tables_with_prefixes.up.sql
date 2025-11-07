-- Migration: 20230101000004_rename_tables_with_prefixes
-- Up

-- 重命名用户表为 lao_users
ALTER TABLE users RENAME TO lao_users;

-- 重命名会话表为 lao_sessions
ALTER TABLE sessions RENAME TO lao_sessions;

-- 重命名太上域表为 tai_domains
ALTER TABLE taishang_domains RENAME TO tai_domains;

-- 重命名老君域表为 lao_domains
ALTER TABLE laojun_domains RENAME TO lao_domains;

-- 更新索引名称以匹配新表名
-- 用户表索引
DROP INDEX IF EXISTS idx_users_username;
CREATE INDEX idx_lao_users_username ON lao_users(username);

DROP INDEX IF EXISTS idx_users_email;
CREATE INDEX idx_lao_users_email ON lao_users(email);

DROP INDEX IF EXISTS idx_users_role;
CREATE INDEX idx_lao_users_role ON lao_users(role);

DROP INDEX IF EXISTS idx_users_is_active;
CREATE INDEX idx_lao_users_is_active ON lao_users(is_active);

-- 会话表索引
DROP INDEX IF EXISTS idx_sessions_user_id;
CREATE INDEX idx_lao_sessions_user_id ON lao_sessions(user_id);

DROP INDEX IF EXISTS idx_sessions_token_hash;
CREATE INDEX idx_lao_sessions_token_hash ON lao_sessions(token_hash);

DROP INDEX IF EXISTS idx_sessions_expires_at;
CREATE INDEX idx_lao_sessions_expires_at ON lao_sessions(expires_at);

-- 太上域表索引
DROP INDEX IF EXISTS idx_taishang_domains_owner_id;
CREATE INDEX idx_tai_domains_owner_id ON tai_domains(owner_id);

DROP INDEX IF EXISTS idx_taishang_domains_is_active;
CREATE INDEX idx_tai_domains_is_active ON tai_domains(is_active);

DROP INDEX IF EXISTS idx_taishang_domains_name;
CREATE INDEX idx_tai_domains_name ON tai_domains(name);

-- 老君域表索引
DROP INDEX IF EXISTS idx_laojun_domains_owner_id;
CREATE INDEX idx_lao_domains_owner_id ON lao_domains(owner_id);

DROP INDEX IF EXISTS idx_laojun_domains_is_active;
CREATE INDEX idx_lao_domains_is_active ON lao_domains(is_active);

DROP INDEX IF EXISTS idx_laojun_domains_name;
CREATE INDEX idx_lao_domains_name ON lao_domains(name);

-- 更新外键约束名称（如果需要）
-- 注意：PostgreSQL会自动更新外键引用，但约束名称可能需要手动更新