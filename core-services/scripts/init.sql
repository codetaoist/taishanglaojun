-- MySQL 数据库初始化脚本
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS laojun CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE laojun;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    avatar_url TEXT,
    phone VARCHAR(20),
    status VARCHAR(20) DEFAULT 'active',
    silicon_level INT DEFAULT 1,
    consciousness_level INT DEFAULT 1,
    thinking_level INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP NULL,
    metadata JSON
);

-- 文化智慧表
CREATE TABLE IF NOT EXISTS cultural_wisdom (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100) NOT NULL,
    tags JSON,
    source VARCHAR(255),
    author VARCHAR(100),
    difficulty_level INT DEFAULT 1,
    silicon_level INT DEFAULT 1,
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    share_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_difficulty_level (difficulty_level),
    INDEX idx_silicon_level (silicon_level),
    INDEX idx_created_at (created_at)
);

-- 对话表
CREATE TABLE IF NOT EXISTS conversations (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    title VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    context JSON,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    conversation_id CHAR(36) NOT NULL,
    role ENUM('user', 'assistant', 'system') NOT NULL,
    content TEXT NOT NULL,
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    INDEX idx_conversation_id (conversation_id),
    INDEX idx_role (role),
    INDEX idx_created_at (created_at)
);

-- 用户行为表
CREATE TABLE IF NOT EXISTS user_behaviors (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50),
    target_id CHAR(36),
    metadata JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_target_type (target_type),
    INDEX idx_created_at (created_at)
);

-- 收藏表
CREATE TABLE IF NOT EXISTS favorites (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_favorite (user_id, target_type, target_id),
    INDEX idx_user_id (user_id),
    INDEX idx_target_type (target_type),
    INDEX idx_created_at (created_at)
);

-- 学习进度表
CREATE TABLE IF NOT EXISTS learning_progress (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    content_id CHAR(36) NOT NULL,
    content_type VARCHAR(50) NOT NULL,
    progress_percentage DECIMAL(5,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'in_progress',
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    metadata JSON,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_progress (user_id, content_id, content_type),
    INDEX idx_user_id (user_id),
    INDEX idx_content_type (content_type),
    INDEX idx_status (status),
    INDEX idx_progress_percentage (progress_percentage)
);

-- 健康记录表
CREATE TABLE IF NOT EXISTS health_records (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    record_type VARCHAR(50) NOT NULL,
    record_date DATE NOT NULL,
    metrics JSON NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_record_type (record_type),
    INDEX idx_record_date (record_date),
    INDEX idx_created_at (created_at)
);

-- 任务表
CREATE TABLE IF NOT EXISTS tasks (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    priority ENUM('low', 'medium', 'high', 'urgent') DEFAULT 'medium',
    status ENUM('pending', 'in_progress', 'completed', 'cancelled') DEFAULT 'pending',
    due_date TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    metadata JSON,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_category (category),
    INDEX idx_priority (priority),
    INDEX idx_status (status),
    INDEX idx_due_date (due_date),
    INDEX idx_created_at (created_at)
);

-- 帖子表
CREATE TABLE IF NOT EXISTS posts (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    category VARCHAR(100),
    tags JSON,
    status ENUM('draft', 'published', 'archived') DEFAULT 'draft',
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    published_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_category (category),
    INDEX idx_status (status),
    INDEX idx_published_at (published_at),
    INDEX idx_created_at (created_at)
);

-- 评论表
CREATE TABLE IF NOT EXISTS comments (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    post_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    parent_id CHAR(36) NULL,
    content TEXT NOT NULL,
    status ENUM('active', 'hidden', 'deleted') DEFAULT 'active',
    like_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES comments(id) ON DELETE CASCADE,
    INDEX idx_post_id (post_id),
    INDEX idx_user_id (user_id),
    INDEX idx_parent_id (parent_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    user_id CHAR(36),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id CHAR(36),
    old_values JSON,
    new_values JSON,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_resource_type (resource_type),
    INDEX idx_resource_id (resource_id),
    INDEX idx_created_at (created_at)
);

-- 系统日志表
CREATE TABLE IF NOT EXISTS system_logs (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    level ENUM('debug', 'info', 'warn', 'error', 'fatal') NOT NULL,
    message TEXT NOT NULL,
    module VARCHAR(100),
    user_id CHAR(36) NULL,
    ip VARCHAR(45),
    user_agent TEXT,
    extra JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_level (level),
    INDEX idx_module (module),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);

-- 系统问题表
CREATE TABLE IF NOT EXISTS system_issues (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity ENUM('low', 'medium', 'high', 'critical') NOT NULL,
    status ENUM('open', 'investigating', 'resolved', 'closed') DEFAULT 'open',
    category VARCHAR(100),
    affected_module VARCHAR(100),
    reporter_id CHAR(36),
    assignee_id CHAR(36),
    resolved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    metadata JSON,
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_severity (severity),
    INDEX idx_status (status),
    INDEX idx_category (category),
    INDEX idx_affected_module (affected_module),
    INDEX idx_reporter_id (reporter_id),
    INDEX idx_assignee_id (assignee_id),
    INDEX idx_created_at (created_at)
);

-- 插入示例数据
INSERT INTO users (id, username, email, password_hash, full_name, status, silicon_level, consciousness_level, thinking_level) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'admin', 'admin@taishanglaojun.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye', '系统管理员', 'active', 10, 10, 10),
('550e8400-e29b-41d4-a716-446655440001', 'testuser', 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye', '测试用户', 'active', 1, 1, 1);

INSERT INTO cultural_wisdom (id, title, content, category, tags, source, author, difficulty_level, silicon_level) VALUES
('660e8400-e29b-41d4-a716-446655440000', '道德经第一章', '道可道，非常道。名可名，非常名。', '道德经', '["道德经", "老子", "哲学"]', '道德经', '老子', 1, 1),
('660e8400-e29b-41d4-a716-446655440001', '论语学而篇', '学而时习之，不亦说乎？', '论语', '["论语", "孔子", "学习"]', '论语', '孔子', 1, 1);

INSERT INTO system_logs (id, level, message, module, user_id, ip, extra) VALUES
('770e8400-e29b-41d4-a716-446655440000', 'info', '系统启动成功', 'system', NULL, '127.0.0.1', '{"startup_time": "2024-01-01T00:00:00Z"}'),
('770e8400-e29b-41d4-a716-446655440001', 'info', '用户登录', 'auth', '550e8400-e29b-41d4-a716-446655440001', '192.168.1.100', '{"login_method": "password"}'),
('770e8400-e29b-41d4-a716-446655440002', 'warn', '数据库连接缓慢', 'database', NULL, '127.0.0.1', '{"response_time": 5000}'),
('770e8400-e29b-41d4-a716-446655440003', 'error', '文件上传失败', 'upload', '550e8400-e29b-41d4-a716-446655440001', '192.168.1.100', '{"file_size": 10485760, "error": "disk_full"}');

INSERT INTO system_issues (id, title, description, severity, status, category, affected_module, reporter_id) VALUES
('880e8400-e29b-41d4-a716-446655440000', '数据库连接超时', '在高并发情况下，数据库连接经常超时', 'high', 'investigating', 'performance', 'database', '550e8400-e29b-41d4-a716-446655440000'),
('880e8400-e29b-41d4-a716-446655440001', '用户登录缓慢', '用户反馈登录过程需要等待较长时间', 'medium', 'open', 'performance', 'auth', '550e8400-e29b-41d4-a716-446655440001'),
('880e8400-e29b-41d4-a716-446655440002', '文件上传失败', '大文件上传时经常失败', 'medium', 'resolved', 'functionality', 'upload', '550e8400-e29b-41d4-a716-446655440001');

-- 创建用户统计视图
CREATE VIEW user_stats AS
SELECT 
    u.id,
    u.username,
    u.email,
    u.status,
    u.silicon_level,
    u.consciousness_level,
    u.thinking_level,
    COUNT(DISTINCT c.id) as conversation_count,
    COUNT(DISTINCT m.id) as message_count,
    COUNT(DISTINCT f.id) as favorite_count,
    COUNT(DISTINCT lp.id) as learning_progress_count,
    u.created_at,
    u.last_login_at
FROM users u
LEFT JOIN conversations c ON u.id = c.user_id
LEFT JOIN messages m ON c.id = m.conversation_id
LEFT JOIN favorites f ON u.id = f.user_id
LEFT JOIN learning_progress lp ON u.id = lp.user_id
GROUP BY u.id, u.username, u.email, u.status, u.silicon_level, u.consciousness_level, u.thinking_level, u.created_at, u.last_login_at;