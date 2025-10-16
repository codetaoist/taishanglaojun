-- 插入示例日志数据
INSERT INTO system_logs (level, message, module, ip, extra) VALUES 
('info', '系统启动成功', 'system', '127.0.0.1', '{"startup_time": "2024-01-01T00:00:00Z"}'),
('info', '用户登录', 'auth', '192.168.1.100', '{"login_method": "password"}'),
('warn', '数据库连接缓慢', 'database', '127.0.0.1', '{"response_time": 5000}'),
('error', '文件上传失败', 'upload', '192.168.1.100', '{"file_size": 10485760, "error": "disk_full"}'),
('info', '定时任务执行', 'scheduler', '127.0.0.1', '{"task": "backup", "duration": 120}'),
('warn', '内存使用率过高', 'system', '127.0.0.1', '{"memory_usage": 85}'),
('error', 'API调用失败', 'api', '192.168.1.200', '{"endpoint": "/api/users", "status_code": 500}');

-- 插入示例问题数据
INSERT INTO system_issues (title, description, severity, status, category, affected_module) VALUES 
('数据库连接超时', '在高并发情况下，数据库连接经常超时，影响用户体验', 'high', 'investigating', 'performance', 'database'),
('用户登录缓慢', '用户反馈登录过程需要等待较长时间，特别是在高峰期', 'medium', 'open', 'performance', 'auth'),
('文件上传失败', '大文件上传时经常失败，需要优化上传机制', 'medium', 'resolved', 'functionality', 'upload'),
('内存泄漏问题', '系统运行一段时间后内存使用率持续上升', 'high', 'open', 'bug', 'system'),
('API响应慢', '某些API接口响应时间过长，需要优化', 'medium', 'investigating', 'performance', 'api');