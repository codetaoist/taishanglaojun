-- Migration: 20230101000004_rename_tables_with_prefixes
-- Down

-- 恢复表名
ALTER TABLE lao_users RENAME TO users;
ALTER TABLE lao_sessions RENAME TO sessions;
ALTER TABLE tai_domains RENAME TO taishang_domains;
ALTER TABLE lao_domains RENAME TO laojun_domains;

-- 恢复索引名称
-- 用户表索引
DROP INDEX IF EXISTS idx_lao_users_username;
CREATE INDEX idx_users_username ON users(username);

DROP INDEX IF EXISTS idx_lao_users_email;
CREATE INDEX idx_users_email ON users(email);

DROP INDEX IF EXISTS idx_lao_users_role;
CREATE INDEX idx_users_role ON users(role);

DROP INDEX IF EXISTS idx_lao_users_is_active;
CREATE INDEX idx_users_is_active ON users(is_active);

-- 会话表索引
DROP INDEX IF EXISTS idx_lao_sessions_user_id;
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

DROP INDEX IF EXISTS idx_lao_sessions_token_hash;
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);

DROP INDEX IF EXISTS idx_lao_sessions_expires_at;
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- 太上域表索引
DROP INDEX IF EXISTS idx_tai_domains_owner_id;
CREATE INDEX idx_taishang_domains_owner_id ON taishang_domains(owner_id);

DROP INDEX IF EXISTS idx_tai_domains_is_active;
CREATE INDEX idx_taishang_domains_is_active ON taishang_domains(is_active);

DROP INDEX IF EXISTS idx_tai_domains_name;
CREATE INDEX idx_taishang_domains_name ON taishang_domains(name);

-- 老君域表索引
DROP INDEX IF EXISTS idx_lao_domains_owner_id;
CREATE INDEX idx_laojun_domains_owner_id ON laojun_domains(owner_id);

DROP INDEX IF EXISTS idx_lao_domains_is_active;
CREATE INDEX idx_laojun_domains_is_active ON laojun_domains(is_active);

DROP INDEX IF EXISTS idx_lao_domains_name;
CREATE INDEX idx_laojun_domains_name ON laojun_domains(name);