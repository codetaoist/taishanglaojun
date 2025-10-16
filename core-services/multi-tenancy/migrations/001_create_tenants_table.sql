-- 创建租户表
CREATE TABLE IF NOT EXISTS tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subdomain VARCHAR(100) UNIQUE NOT NULL,
    domain VARCHAR(255),
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    isolation_strategy VARCHAR(50) NOT NULL DEFAULT 'row_level',
    
    -- 租户设置 (JSON)
    settings JSONB NOT NULL DEFAULT '{}',
    
    -- 租户配额 (JSON)
    quota JSONB NOT NULL DEFAULT '{}',
    
    -- 租户使用情况 (JSON)
    usage JSONB NOT NULL DEFAULT '{}',
    
    -- 租户元数据 (JSON)
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tenants_subdomain ON tenants(subdomain);
CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);
CREATE INDEX IF NOT EXISTS idx_tenants_deleted_at ON tenants(deleted_at);

-- 创建租户用户关联表
CREATE TABLE IF NOT EXISTS tenant_users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    permissions JSONB NOT NULL DEFAULT '[]',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- 唯一约束
    UNIQUE(tenant_id, user_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_id ON tenant_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_users_user_id ON tenant_users(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_users_role ON tenant_users(role);
CREATE INDEX IF NOT EXISTS idx_tenant_users_created_at ON tenant_users(created_at);
CREATE INDEX IF NOT EXISTS idx_tenant_users_deleted_at ON tenant_users(deleted_at);

-- 创建租户订阅表
CREATE TABLE IF NOT EXISTS tenant_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan_name VARCHAR(100) NOT NULL,
    plan_type VARCHAR(50) NOT NULL DEFAULT 'monthly',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    
    -- 订阅详情
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    trial_end_date TIMESTAMP WITH TIME ZONE,
    
    -- 价格信息
    price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    
    -- 订阅元数据 (JSON)
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_tenant_id ON tenant_subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_status ON tenant_subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_plan_name ON tenant_subscriptions(plan_name);
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_start_date ON tenant_subscriptions(start_date);
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_end_date ON tenant_subscriptions(end_date);
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_created_at ON tenant_subscriptions(created_at);
CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_deleted_at ON tenant_subscriptions(deleted_at);

-- 创建租户审计日志表
CREATE TABLE IF NOT EXISTS tenant_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT REFERENCES tenants(id) ON DELETE CASCADE,
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    
    -- 操作详情
    old_values JSONB,
    new_values JSONB,
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- 请求信息
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(255),
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_tenant_id ON tenant_audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_user_id ON tenant_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_action ON tenant_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_resource_type ON tenant_audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_tenant_audit_logs_created_at ON tenant_audit_logs(created_at);

-- 创建租户配额使用历史表
CREATE TABLE IF NOT EXISTS tenant_quota_usage_history (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    resource_type VARCHAR(100) NOT NULL,
    usage_value BIGINT NOT NULL DEFAULT 0,
    quota_value BIGINT NOT NULL DEFAULT 0,
    usage_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    
    -- 时间戳
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tenant_quota_usage_history_tenant_id ON tenant_quota_usage_history(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_quota_usage_history_resource_type ON tenant_quota_usage_history(resource_type);
CREATE INDEX IF NOT EXISTS idx_tenant_quota_usage_history_recorded_at ON tenant_quota_usage_history(recorded_at);

-- 创建租户设置历史表
CREATE TABLE IF NOT EXISTS tenant_settings_history (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id BIGINT,
    setting_key VARCHAR(255) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    
    -- 时间戳
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_tenant_settings_history_tenant_id ON tenant_settings_history(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_settings_history_setting_key ON tenant_settings_history(setting_key);
CREATE INDEX IF NOT EXISTS idx_tenant_settings_history_changed_at ON tenant_settings_history(changed_at);

-- 创建触发器函数：更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表创建触发器
CREATE TRIGGER update_tenants_updated_at 
    BEFORE UPDATE ON tenants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_users_updated_at 
    BEFORE UPDATE ON tenant_users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_subscriptions_updated_at 
    BEFORE UPDATE ON tenant_subscriptions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建触发器函数：记录租户设置变更
CREATE OR REPLACE FUNCTION log_tenant_settings_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- 只有当设置实际发生变化时才记录
    IF OLD.settings IS DISTINCT FROM NEW.settings THEN
        INSERT INTO tenant_settings_history (
            tenant_id,
            user_id,
            setting_key,
            old_value,
            new_value,
            changed_at
        ) VALUES (
            NEW.id,
            NULL, -- 需要从应用层传入用户ID
            'settings',
            OLD.settings,
            NEW.settings,
            NOW()
        );
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为租户表创建设置变更触发器
CREATE TRIGGER log_tenant_settings_changes_trigger
    AFTER UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION log_tenant_settings_changes();

-- 创建视图：活跃租户统计
CREATE OR REPLACE VIEW active_tenants_stats AS
SELECT 
    COUNT(*) as total_active_tenants,
    COUNT(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 END) as new_tenants_last_30_days,
    COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new_tenants_last_7_days,
    AVG(EXTRACT(EPOCH FROM (NOW() - created_at))/86400) as avg_tenant_age_days
FROM tenants 
WHERE status = 'active' AND deleted_at IS NULL;

-- 创建视图：租户用户统计
CREATE OR REPLACE VIEW tenant_user_stats AS
SELECT 
    t.id as tenant_id,
    t.name as tenant_name,
    COUNT(tu.id) as total_users,
    COUNT(CASE WHEN tu.role = 'owner' THEN 1 END) as owners,
    COUNT(CASE WHEN tu.role = 'admin' THEN 1 END) as admins,
    COUNT(CASE WHEN tu.role = 'member' THEN 1 END) as members,
    COUNT(CASE WHEN tu.role = 'viewer' THEN 1 END) as viewers
FROM tenants t
LEFT JOIN tenant_users tu ON t.id = tu.tenant_id AND tu.deleted_at IS NULL
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.name;

-- 创建视图：租户配额使用情况
CREATE OR REPLACE VIEW tenant_quota_overview AS
SELECT 
    t.id as tenant_id,
    t.name as tenant_name,
    t.status,
    (t.quota->>'max_users')::bigint as max_users,
    (t.usage->>'current_users')::bigint as current_users,
    (t.quota->>'max_storage')::bigint as max_storage,
    (t.usage->>'current_storage')::bigint as current_storage,
    (t.quota->>'max_api_requests')::bigint as max_api_requests,
    (t.usage->>'current_api_requests')::bigint as current_api_requests,
    CASE 
        WHEN (t.quota->>'max_users')::bigint > 0 THEN 
            ROUND(((t.usage->>'current_users')::bigint * 100.0) / (t.quota->>'max_users')::bigint, 2)
        ELSE 0 
    END as users_usage_percentage,
    CASE 
        WHEN (t.quota->>'max_storage')::bigint > 0 THEN 
            ROUND(((t.usage->>'current_storage')::bigint * 100.0) / (t.quota->>'max_storage')::bigint, 2)
        ELSE 0 
    END as storage_usage_percentage
FROM tenants t
WHERE t.deleted_at IS NULL;

-- 插入默认数据（可选）
-- INSERT INTO tenants (name, subdomain, description, settings, quota, usage, metadata) VALUES
-- ('默认租户', 'default', '系统默认租户', 
--  '{"language": "zh-CN", "timezone": "Asia/Shanghai", "date_format": "YYYY-MM-DD", "time_format": "24h", "currency": "CNY"}',
--  '{"max_users": 100, "max_storage": 10737418240, "max_api_requests": 10000, "max_bandwidth": 1073741824}',
--  '{"current_users": 0, "current_storage": 0, "current_api_requests": 0, "current_bandwidth": 0}',
--  '{"created_by": "system", "version": "1.0.0"}'
-- );

-- 添加注释
COMMENT ON TABLE tenants IS '租户表';
COMMENT ON COLUMN tenants.name IS '租户名称';
COMMENT ON COLUMN tenants.subdomain IS '子域名';
COMMENT ON COLUMN tenants.domain IS '自定义域名';
COMMENT ON COLUMN tenants.status IS '租户状态: active, suspended, inactive';
COMMENT ON COLUMN tenants.isolation_strategy IS '数据隔离策略: row_level, schema, database';
COMMENT ON COLUMN tenants.settings IS '租户设置JSON';
COMMENT ON COLUMN tenants.quota IS '租户配额JSON';
COMMENT ON COLUMN tenants.usage IS '租户使用情况JSON';
COMMENT ON COLUMN tenants.metadata IS '租户元数据JSON';

COMMENT ON TABLE tenant_users IS '租户用户关联表';
COMMENT ON COLUMN tenant_users.role IS '用户角色: owner, admin, member, viewer';
COMMENT ON COLUMN tenant_users.permissions IS '用户权限JSON数组';

COMMENT ON TABLE tenant_subscriptions IS '租户订阅表';
COMMENT ON COLUMN tenant_subscriptions.plan_type IS '订阅类型: monthly, yearly, lifetime';
COMMENT ON COLUMN tenant_subscriptions.status IS '订阅状态: active, cancelled, expired, trial';

COMMENT ON TABLE tenant_audit_logs IS '租户审计日志表';
COMMENT ON TABLE tenant_quota_usage_history IS '租户配额使用历史表';
COMMENT ON TABLE tenant_settings_history IS '租户设置变更历史表';