-- 太上序列零数据库初始化脚本
-- 创建数据库和基础表结构

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS taishang_sequence_zero;
\c taishang_sequence_zero;

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    permission_level INTEGER DEFAULT 1,
    trust_score DECIMAL(3,2) DEFAULT 0.5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建权限表
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    required_level INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建意识融合历史表
CREATE TABLE IF NOT EXISTS consciousness_fusion_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_type VARCHAR(100) NOT NULL,
    start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP,
    effectiveness DECIMAL(3,2),
    feedback TEXT,
    metrics JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建智慧问答表
CREATE TABLE IF NOT EXISTS wisdom_qa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    source VARCHAR(200),
    category VARCHAR(100),
    wisdom_level VARCHAR(50),
    references JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文化知识表
CREATE TABLE IF NOT EXISTS cultural_knowledge (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    tags JSONB DEFAULT '[]',
    difficulty VARCHAR(50) DEFAULT 'intermediate',
    source VARCHAR(200),
    author VARCHAR(100),
    period VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建修养计划表
CREATE TABLE IF NOT EXISTS cultivation_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    goals JSONB DEFAULT '[]',
    practices JSONB DEFAULT '[]',
    duration INTEGER NOT NULL DEFAULT 30,
    difficulty VARCHAR(50) DEFAULT 'intermediate',
    progress DECIMAL(3,2) DEFAULT 0.0,
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建修行记录表
CREATE TABLE IF NOT EXISTS practice_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID REFERENCES cultivation_plans(id) ON DELETE SET NULL,
    practice_type VARCHAR(100) NOT NULL,
    duration INTEGER NOT NULL,
    quality DECIMAL(3,2) NOT NULL,
    reflection TEXT,
    insights JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文化传承故事表
CREATE TABLE IF NOT EXISTS heritage_stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100),
    region VARCHAR(100),
    period VARCHAR(100),
    characters JSONB DEFAULT '[]',
    moral_lesson TEXT,
    tags JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文化体验分享表
CREATE TABLE IF NOT EXISTS cultural_experiences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    images JSONB DEFAULT '[]',
    location VARCHAR(200),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建用户会话表
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建意识状态记录表
CREATE TABLE IF NOT EXISTS consciousness_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    emotional_state JSONB NOT NULL DEFAULT '{}',
    cognitive_level DECIMAL(3,2) NOT NULL DEFAULT 0.5,
    spiritual_depth DECIMAL(3,2) NOT NULL DEFAULT 0.5,
    personality_traits JSONB NOT NULL DEFAULT '{}',
    consciousness_type VARCHAR(50) NOT NULL DEFAULT 'beginner',
    fusion_level DECIMAL(3,2) NOT NULL DEFAULT 0.0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文化分析记录表
CREATE TABLE IF NOT EXISTS cultural_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    input_text TEXT NOT NULL,
    analysis_result JSONB,
    confidence_score DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    result VARCHAR(20) NOT NULL, -- 'granted' or 'denied'
    reason TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 插入基础权限数据
INSERT INTO permissions (name, resource, action, required_level) VALUES
('查看个人资料', 'user_profile', 'read', 1),
('修改个人资料', 'user_profile', 'update', 2),
('查看意识数据', 'consciousness_data', 'read', 5),
('修改意识数据', 'consciousness_data', 'write', 7),
('查看系统控制', 'system_control', 'read', 8),
('修改系统控制', 'system_control', 'write', 9);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_consciousness_states_user_time ON consciousness_states(user_id, recorded_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time ON audit_logs(user_id, created_at);

COMMIT;