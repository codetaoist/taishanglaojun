-- 安全模块数据库迁移文件
-- 创建时间: 2024-01-01
-- 版本: 001

-- 创建威胁告警表
CREATE TABLE IF NOT EXISTS threat_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    threat_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    source VARCHAR(255) NOT NULL,
    target VARCHAR(255),
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'ignored')),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(255)
);

-- 创建检测规则表
CREATE TABLE IF NOT EXISTS detection_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    rule_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    pattern TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,
    threshold INTEGER DEFAULT 1,
    time_window INTEGER DEFAULT 300, -- 秒
    action VARCHAR(50) DEFAULT 'alert' CHECK (action IN ('alert', 'block', 'log')),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

-- 创建漏洞表
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    cvss_score DECIMAL(3,1),
    cve_id VARCHAR(50),
    target VARCHAR(255) NOT NULL,
    vulnerability_type VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'fixed', 'ignored', 'false_positive')),
    proof_of_concept TEXT,
    remediation TEXT,
    references TEXT[],
    metadata JSONB,
    discovered_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    fixed_at TIMESTAMP WITH TIME ZONE,
    scan_job_id UUID
);

-- 创建扫描任务表
CREATE TABLE IF NOT EXISTS scan_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    scan_type VARCHAR(50) NOT NULL CHECK (scan_type IN ('web', 'network', 'api', 'mobile')),
    target VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    config JSONB,
    results JSONB,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    timeout_at TIMESTAMP WITH TIME ZONE
);

-- 创建渗透测试项目表
CREATE TABLE IF NOT EXISTS pentest_projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    target VARCHAR(255) NOT NULL,
    scope TEXT[],
    status VARCHAR(20) NOT NULL DEFAULT 'planning' CHECK (status IN ('planning', 'reconnaissance', 'scanning', 'enumeration', 'exploitation', 'post_exploitation', 'reporting', 'completed', 'cancelled')),
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    methodology VARCHAR(100) DEFAULT 'OWASP',
    config JSONB,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    assigned_to VARCHAR(255)
);

-- 创建渗透测试结果表
CREATE TABLE IF NOT EXISTS pentest_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES pentest_projects(id) ON DELETE CASCADE,
    phase VARCHAR(50) NOT NULL,
    tool VARCHAR(100),
    command TEXT,
    output TEXT,
    findings JSONB,
    severity VARCHAR(20) CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'verified', 'false_positive', 'fixed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- 创建安全课程表
CREATE TABLE IF NOT EXISTS security_courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    difficulty VARCHAR(20) NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
    duration INTEGER, -- 分钟
    content JSONB,
    prerequisites TEXT[],
    learning_objectives TEXT[],
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255)
);

-- 创建实验环境表
CREATE TABLE IF NOT EXISTS lab_environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    lab_type VARCHAR(100) NOT NULL,
    difficulty VARCHAR(20) NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
    config JSONB,
    docker_image VARCHAR(255),
    port_mappings JSONB,
    environment_variables JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('inactive', 'active', 'maintenance')),
    max_instances INTEGER DEFAULT 10,
    current_instances INTEGER DEFAULT 0,
    timeout_minutes INTEGER DEFAULT 60,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- 创建安全认证表
CREATE TABLE IF NOT EXISTS security_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    requirements JSONB,
    validity_period INTEGER, -- 天数
    certificate_template TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'deprecated')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- 创建审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(255),
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'blocked')),
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建安全事件表
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    source VARCHAR(255) NOT NULL,
    target VARCHAR(255),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'detected' CHECK (status IN ('detected', 'investigating', 'resolved', 'false_positive')),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(255)
);

-- 创建合规报告表
CREATE TABLE IF NOT EXISTS compliance_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard VARCHAR(100) NOT NULL,
    version VARCHAR(50),
    scope TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'in_progress', 'completed', 'approved')),
    compliance_score DECIMAL(5,2),
    total_controls INTEGER,
    passed_controls INTEGER,
    failed_controls INTEGER,
    findings JSONB,
    recommendations JSONB,
    report_data JSONB,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    generated_by VARCHAR(255),
    approved_by VARCHAR(255),
    approved_at TIMESTAMP WITH TIME ZONE
);

-- 创建用户安全档案表
CREATE TABLE IF NOT EXISTS user_security_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL UNIQUE,
    security_level VARCHAR(20) DEFAULT 'basic' CHECK (security_level IN ('basic', 'intermediate', 'advanced', 'expert')),
    completed_courses UUID[],
    earned_certifications UUID[],
    lab_progress JSONB,
    achievements JSONB,
    last_activity TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建威胁情报表
CREATE TABLE IF NOT EXISTS threat_intelligence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    indicator_type VARCHAR(50) NOT NULL CHECK (indicator_type IN ('ip', 'domain', 'url', 'hash', 'email')),
    indicator_value VARCHAR(500) NOT NULL,
    threat_type VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    confidence INTEGER CHECK (confidence >= 0 AND confidence <= 100),
    source VARCHAR(255) NOT NULL,
    description TEXT,
    tags TEXT[],
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_threat_alerts_severity ON threat_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_threat_alerts_status ON threat_alerts(status);
CREATE INDEX IF NOT EXISTS idx_threat_alerts_created_at ON threat_alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_threat_alerts_source ON threat_alerts(source);

CREATE INDEX IF NOT EXISTS idx_detection_rules_enabled ON detection_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_detection_rules_rule_type ON detection_rules(rule_type);

CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity ON vulnerabilities(severity);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_status ON vulnerabilities(status);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_target ON vulnerabilities(target);
CREATE INDEX IF NOT EXISTS idx_vulnerabilities_scan_job_id ON vulnerabilities(scan_job_id);

CREATE INDEX IF NOT EXISTS idx_scan_jobs_status ON scan_jobs(status);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_scan_type ON scan_jobs(scan_type);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_created_at ON scan_jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_created_by ON scan_jobs(created_by);

CREATE INDEX IF NOT EXISTS idx_pentest_projects_status ON pentest_projects(status);
CREATE INDEX IF NOT EXISTS idx_pentest_projects_created_by ON pentest_projects(created_by);
CREATE INDEX IF NOT EXISTS idx_pentest_projects_assigned_to ON pentest_projects(assigned_to);

CREATE INDEX IF NOT EXISTS idx_pentest_results_project_id ON pentest_results(project_id);
CREATE INDEX IF NOT EXISTS idx_pentest_results_phase ON pentest_results(phase);
CREATE INDEX IF NOT EXISTS idx_pentest_results_severity ON pentest_results(severity);

CREATE INDEX IF NOT EXISTS idx_security_courses_category ON security_courses(category);
CREATE INDEX IF NOT EXISTS idx_security_courses_difficulty ON security_courses(difficulty);
CREATE INDEX IF NOT EXISTS idx_security_courses_status ON security_courses(status);

CREATE INDEX IF NOT EXISTS idx_lab_environments_lab_type ON lab_environments(lab_type);
CREATE INDEX IF NOT EXISTS idx_lab_environments_status ON lab_environments(status);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);

CREATE INDEX IF NOT EXISTS idx_security_events_event_type ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events(severity);
CREATE INDEX IF NOT EXISTS idx_security_events_status ON security_events(status);
CREATE INDEX IF NOT EXISTS idx_security_events_created_at ON security_events(created_at);

CREATE INDEX IF NOT EXISTS idx_compliance_reports_standard ON compliance_reports(standard);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_status ON compliance_reports(status);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_generated_at ON compliance_reports(generated_at);

CREATE INDEX IF NOT EXISTS idx_user_security_profiles_user_id ON user_security_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_security_profiles_security_level ON user_security_profiles(security_level);

CREATE INDEX IF NOT EXISTS idx_threat_intelligence_indicator_type ON threat_intelligence(indicator_type);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_indicator_value ON threat_intelligence(indicator_value);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_threat_type ON threat_intelligence(threat_type);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_severity ON threat_intelligence(severity);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_source ON threat_intelligence(source);
CREATE INDEX IF NOT EXISTS idx_threat_intelligence_expires_at ON threat_intelligence(expires_at);

-- 创建触发器函数用于更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建触发器
CREATE TRIGGER update_threat_alerts_updated_at BEFORE UPDATE ON threat_alerts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_detection_rules_updated_at BEFORE UPDATE ON detection_rules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_vulnerabilities_updated_at BEFORE UPDATE ON vulnerabilities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_scan_jobs_updated_at BEFORE UPDATE ON scan_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pentest_projects_updated_at BEFORE UPDATE ON pentest_projects FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_pentest_results_updated_at BEFORE UPDATE ON pentest_results FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_security_courses_updated_at BEFORE UPDATE ON security_courses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lab_environments_updated_at BEFORE UPDATE ON lab_environments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_security_certifications_updated_at BEFORE UPDATE ON security_certifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_security_events_updated_at BEFORE UPDATE ON security_events FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_compliance_reports_updated_at BEFORE UPDATE ON compliance_reports FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_security_profiles_updated_at BEFORE UPDATE ON user_security_profiles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_threat_intelligence_updated_at BEFORE UPDATE ON threat_intelligence FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入默认检测规则
INSERT INTO detection_rules (name, rule_type, severity, pattern, description, threshold, time_window, action) VALUES
('SQL注入检测', 'sql_injection', 'high', '(?i)(union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|drop\s+table)', 'SQL注入攻击检测规则', 3, 300, 'block'),
('XSS攻击检测', 'xss', 'medium', '(?i)(<script[^>]*>|javascript:|onload\s*=|onerror\s*=)', 'XSS跨站脚本攻击检测规则', 3, 300, 'alert'),
('路径遍历检测', 'path_traversal', 'medium', '(\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e%5c)', '路径遍历攻击检测规则', 5, 300, 'block'),
('命令注入检测', 'command_injection', 'high', '(?i)(;|\||\&\&|\|\|)\s*(ls|dir|cat|type|echo|ping|wget|curl)', '命令注入攻击检测规则', 3, 300, 'block'),
('暴力破解检测', 'brute_force', 'high', '', '暴力破解攻击检测规则', 10, 60, 'block'),
('DDoS攻击检测', 'ddos', 'critical', '', 'DDoS攻击检测规则', 100, 60, 'block')
ON CONFLICT (name) DO NOTHING;

-- 插入默认安全课程
INSERT INTO security_courses (title, description, category, difficulty, duration, content, learning_objectives, status, created_by) VALUES
('网络安全基础', '介绍网络安全的基本概念和原理', '基础安全', 'beginner', 120, '{"modules": ["安全概述", "威胁模型", "防护措施"]}', '{"了解网络安全基本概念", "掌握常见威胁类型", "学习基本防护方法"}', 'published', 'system'),
('Web应用安全', '深入学习Web应用安全漏洞和防护技术', 'Web安全', 'intermediate', 180, '{"modules": ["OWASP Top 10", "SQL注入", "XSS攻击", "CSRF攻击"]}', '{"掌握Web安全漏洞", "学习漏洞利用技术", "了解防护措施"}', 'published', 'system'),
('渗透测试入门', '学习渗透测试的基本方法和工具使用', '渗透测试', 'intermediate', 240, '{"modules": ["信息收集", "漏洞扫描", "漏洞利用", "后渗透"]}', '{"掌握渗透测试流程", "学习常用工具使用", "了解法律法规"}', 'published', 'system'),
('安全编码实践', '学习安全编码的最佳实践和常见陷阱', '安全开发', 'advanced', 200, '{"modules": ["输入验证", "输出编码", "认证授权", "加密技术"]}', '{"掌握安全编码原则", "避免常见安全漏洞", "提高代码安全性"}', 'published', 'system')
ON CONFLICT (title) DO NOTHING;

-- 插入默认实验环境
INSERT INTO lab_environments (name, description, lab_type, difficulty, config, docker_image, status, timeout_minutes, created_by) VALUES
('SQL注入实验', 'SQL注入漏洞练习环境', 'web_security', 'beginner', '{"database": "mysql", "vulnerable_app": "dvwa"}', 'vulnerables/web-dvwa', 'active', 60, 'system'),
('XSS攻击实验', 'XSS跨站脚本攻击练习环境', 'web_security', 'beginner', '{"framework": "php", "vulnerable_points": ["reflected", "stored"]}', 'webgoat/webgoat-8.0', 'active', 60, 'system'),
('Linux提权实验', 'Linux系统提权练习环境', 'system_security', 'intermediate', '{"os": "ubuntu", "kernel_version": "4.4.0", "vulnerable_services": ["sudo", "cron"]}', 'vulhub/ubuntu:16.04', 'active', 90, 'system'),
('网络渗透实验', '网络渗透测试综合练习环境', 'network_security', 'advanced', '{"network": "192.168.1.0/24", "services": ["ssh", "http", "ftp", "smb"]}', 'metasploitable/metasploitable2', 'active', 120, 'system')
ON CONFLICT (name) DO NOTHING;

-- 插入默认安全认证
INSERT INTO security_certifications (name, description, requirements, validity_period, status, created_by) VALUES
('网络安全基础认证', '网络安全基础知识认证', '{"courses": ["网络安全基础"], "score": 80, "labs": 3}', 365, 'active', 'system'),
('Web安全专家认证', 'Web应用安全专家认证', '{"courses": ["Web应用安全", "安全编码实践"], "score": 85, "labs": 5, "projects": 2}', 730, 'active', 'system'),
('渗透测试工程师认证', '渗透测试工程师专业认证', '{"courses": ["渗透测试入门", "网络安全基础"], "score": 90, "labs": 10, "projects": 3}', 730, 'active', 'system')
ON CONFLICT (name) DO NOTHING;

-- 插入威胁情报数据
INSERT INTO threat_intelligence (indicator_type, indicator_value, threat_type, severity, confidence, source, description, tags) VALUES
('ip', '192.168.1.100', 'malware', 'high', 90, 'internal_honeypot', '内部蜜罐检测到的恶意IP', '{"malware", "botnet", "internal"}'),
('domain', 'malicious-site.com', 'phishing', 'medium', 75, 'threat_feed', '钓鱼网站域名', '{"phishing", "social_engineering"}'),
('hash', 'd41d8cd98f00b204e9800998ecf8427e', 'malware', 'critical', 95, 'antivirus', '已知恶意软件哈希值', '{"malware", "trojan"}')
ON CONFLICT DO NOTHING;

COMMIT;