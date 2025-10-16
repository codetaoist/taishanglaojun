-- 太上老君AI平台 - 数据库初始化脚本
-- 版本: 1.0.0
-- 创建时间: 2024-01-01

-- 设置数据库编码
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url TEXT,
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    role VARCHAR(20) DEFAULT 'user',
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE
);

-- 创建文化智慧表
CREATE TABLE IF NOT EXISTS cultural_wisdom (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100),
    tags TEXT[],
    source VARCHAR(255),
    author VARCHAR(100),
    dynasty VARCHAR(50),
    difficulty_level INTEGER DEFAULT 1,
    popularity_score FLOAT DEFAULT 0.0,
    embedding vector(1536),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'active'
);

-- 创建AI对话表
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    title VARCHAR(255),
    model VARCHAR(100),
    provider VARCHAR(50),
    context JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active'
);

-- 创建消息表
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- 'user', 'assistant', 'system'
    content TEXT NOT NULL,
    content_type VARCHAR(50) DEFAULT 'text', -- 'text', 'image', 'audio', 'video'
    metadata JSONB DEFAULT '{}',
    tokens_used INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    parent_message_id UUID REFERENCES messages(id)
);

-- 创建用户行为表
CREATE TABLE IF NOT EXISTS user_behaviors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建收藏表
CREATE TABLE IF NOT EXISTS favorites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, resource_type, resource_id)
);

-- 创建学习进度表
CREATE TABLE IF NOT EXISTS learning_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    content_id UUID,
    content_type VARCHAR(50),
    progress_percentage FLOAT DEFAULT 0.0,
    time_spent INTEGER DEFAULT 0, -- 秒
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, content_id, content_type)
);

-- 创建健康记录表
CREATE TABLE IF NOT EXISTS health_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    record_type VARCHAR(50) NOT NULL,
    data JSONB NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- 创建任务表
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    priority VARCHAR(20) DEFAULT 'medium',
    due_date TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建社区帖子表
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(50) DEFAULT 'text',
    category VARCHAR(100),
    tags TEXT[],
    status VARCHAR(20) DEFAULT 'published',
    view_count INTEGER DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建评论表
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    content TEXT NOT NULL,
    parent_comment_id UUID REFERENCES comments(id),
    status VARCHAR(20) DEFAULT 'published',
    like_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建系统日志表
CREATE TABLE IF NOT EXISTS system_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(20) NOT NULL DEFAULT 'info', -- debug, info, warn, error, fatal
    message TEXT NOT NULL,
    module VARCHAR(100), -- 模块名称，如 'auth', 'api', 'database'
    user_id UUID REFERENCES users(id), -- 关联用户（可选）
    ip_address INET, -- 客户端IP
    user_agent TEXT, -- 用户代理
    extra JSONB DEFAULT '{}', -- 额外的结构化数据
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建系统问题检测表
CREATE TABLE IF NOT EXISTS system_issues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    issue_type VARCHAR(50) NOT NULL, -- 问题类型：performance, security, error, warning
    severity VARCHAR(20) NOT NULL DEFAULT 'medium', -- low, medium, high, critical
    title VARCHAR(255) NOT NULL, -- 问题标题
    description TEXT, -- 问题描述
    source VARCHAR(100), -- 问题来源：log_analysis, health_check, monitoring
    affected_component VARCHAR(100), -- 受影响的组件
    status VARCHAR(20) DEFAULT 'open', -- open, investigating, resolved, closed
    metadata JSONB DEFAULT '{}', -- 问题相关的元数据
    first_detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_category ON cultural_wisdom(category);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_tags ON cultural_wisdom USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_status ON cultural_wisdom(status);
CREATE INDEX IF NOT EXISTS idx_cultural_wisdom_embedding ON cultural_wisdom USING ivfflat (embedding vector_cosine_ops);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

CREATE INDEX IF NOT EXISTS idx_user_behaviors_user_id ON user_behaviors(user_id);
CREATE INDEX IF NOT EXISTS idx_user_behaviors_action ON user_behaviors(action);
CREATE INDEX IF NOT EXISTS idx_user_behaviors_created_at ON user_behaviors(created_at);

CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_resource ON favorites(resource_type, resource_id);

CREATE INDEX IF NOT EXISTS idx_learning_progress_user_id ON learning_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_learning_progress_content ON learning_progress(content_id, content_type);

CREATE INDEX IF NOT EXISTS idx_health_records_user_id ON health_records(user_id);
CREATE INDEX IF NOT EXISTS idx_health_records_type ON health_records(record_type);
CREATE INDEX IF NOT EXISTS idx_health_records_recorded_at ON health_records(recorded_at);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_category ON posts(category);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);

CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_comment_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_system_logs_level ON system_logs(level);
CREATE INDEX IF NOT EXISTS idx_system_logs_module ON system_logs(module);
CREATE INDEX IF NOT EXISTS idx_system_logs_user_id ON system_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_system_logs_created_at ON system_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_system_issues_type ON system_issues(issue_type);
CREATE INDEX IF NOT EXISTS idx_system_issues_severity ON system_issues(severity);
CREATE INDEX IF NOT EXISTS idx_system_issues_status ON system_issues(status);
CREATE INDEX IF NOT EXISTS idx_system_issues_source ON system_issues(source);
CREATE INDEX IF NOT EXISTS idx_system_issues_component ON system_issues(affected_component);
CREATE INDEX IF NOT EXISTS idx_system_issues_first_detected ON system_issues(first_detected_at);
CREATE INDEX IF NOT EXISTS idx_system_issues_last_detected ON system_issues(last_detected_at);

-- 插入示例数据
INSERT INTO users (username, email, password_hash, full_name, role) VALUES
('admin', 'admin@taishanglaojun.com', crypt('admin123', gen_salt('bf')), '系统管理员', 'admin'),
('demo', 'demo@taishanglaojun.com', crypt('demo123', gen_salt('bf')), '演示用户', 'user')
ON CONFLICT (email) DO NOTHING;

-- 插入示例文化智慧数据
INSERT INTO cultural_wisdom (title, content, category, tags, source, author, dynasty) VALUES
('道德经第一章', '道可道，非常道；名可名，非常名。无名天地之始，有名万物之母。', '道德经', ARRAY['道德经', '老子', '哲学'], '道德经', '老子', '春秋'),
('论语学而篇', '学而时习之，不亦说乎？有朋自远方来，不亦乐乎？', '论语', ARRAY['论语', '孔子', '教育'], '论语', '孔子', '春秋'),
('庄子逍遥游', '北冥有鱼，其名为鲲。鲲之大，不知其几千里也。', '庄子', ARRAY['庄子', '逍遥', '哲学'], '庄子', '庄子', '战国')
ON CONFLICT DO NOTHING;

-- 插入示例系统日志数据
INSERT INTO system_logs (level, message, module, extra) VALUES
('info', '系统启动成功', 'system', '{"startup_time": "2024-01-15T10:00:00Z", "version": "1.0.0"}'),
('info', '用户登录成功', 'auth', '{"user_id": "admin", "login_method": "password"}'),
('warn', '数据库连接池使用率较高', 'database', '{"pool_usage": 85, "max_connections": 100}'),
('error', 'API请求处理失败', 'api', '{"endpoint": "/api/v1/data", "error": "timeout", "duration_ms": 5000}'),
('info', '定时任务执行完成', 'scheduler', '{"task": "cleanup_logs", "processed_records": 1000}')
ON CONFLICT DO NOTHING;

-- 插入示例系统问题数据
INSERT INTO system_issues (issue_type, severity, title, description, source, affected_component, status, metadata) VALUES
('performance', 'medium', '数据库查询响应时间过长', '部分查询响应时间超过2秒，影响用户体验', 'monitoring', 'database', 'open', '{"avg_response_time": 2.5, "threshold": 2.0, "affected_queries": ["user_stats", "log_analysis"]}'),
('error', 'high', 'API错误率异常', '过去1小时内API错误率达到5%，超过正常阈值', 'log_analysis', 'api_gateway', 'investigating', '{"error_rate": 0.05, "threshold": 0.01, "total_requests": 10000, "failed_requests": 500}'),
('security', 'critical', '检测到异常登录尝试', '来自多个IP的暴力破解登录尝试', 'security_monitor', 'auth_service', 'resolved', '{"failed_attempts": 100, "unique_ips": 15, "time_window": "1h", "blocked_ips": 15}'),
('warning', 'low', '磁盘空间使用率较高', '系统磁盘使用率达到80%，建议清理日志文件', 'health_check', 'storage', 'open', '{"disk_usage": 0.8, "threshold": 0.75, "available_space": "20GB", "total_space": "100GB"}')
ON CONFLICT DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_cultural_wisdom_updated_at BEFORE UPDATE ON cultural_wisdom FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_learning_progress_updated_at BEFORE UPDATE ON learning_progress FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_issues_updated_at BEFORE UPDATE ON system_issues FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建视图
CREATE OR REPLACE VIEW user_stats AS
SELECT 
    u.id,
    u.username,
    u.email,
    u.created_at,
    COUNT(DISTINCT c.id) as conversation_count,
    COUNT(DISTINCT f.id) as favorite_count,
    COUNT(DISTINCT p.id) as post_count,
    COUNT(DISTINCT cm.id) as comment_count
FROM users u
LEFT JOIN conversations c ON u.id = c.user_id
LEFT JOIN favorites f ON u.id = f.user_id
LEFT JOIN posts p ON u.id = p.user_id
LEFT JOIN comments cm ON u.id = cm.user_id
GROUP BY u.id, u.username, u.email, u.created_at;

-- 设置权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;

-- 完成初始化
SELECT 'Database initialization completed successfully!' as status;