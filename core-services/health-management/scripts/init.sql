-- 健康管理服务数据库初始化脚本

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS health_management;

-- 使用数据库
\c health_management;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- 创建枚举类型
CREATE TYPE health_data_type AS ENUM (
    'heart_rate',
    'blood_pressure',
    'blood_sugar',
    'body_temperature',
    'weight',
    'height',
    'bmi',
    'steps',
    'sleep_duration',
    'stress_level',
    'oxygen_saturation',
    'respiratory_rate'
);

CREATE TYPE health_data_source AS ENUM (
    'manual_input',
    'smart_watch',
    'fitness_tracker',
    'medical_device',
    'mobile_app',
    'iot_sensor',
    'hospital_system'
);

CREATE TYPE gender AS ENUM (
    'male',
    'female',
    'other',
    'prefer_not_to_say'
);

CREATE TYPE blood_type AS ENUM (
    'A+',
    'A-',
    'B+',
    'B-',
    'AB+',
    'AB-',
    'O+',
    'O-',
    'unknown'
);

CREATE TYPE alert_severity AS ENUM (
    'low',
    'medium',
    'high',
    'critical'
);

CREATE TYPE alert_status AS ENUM (
    'active',
    'acknowledged',
    'resolved',
    'dismissed'
);

-- 创建健康数据表
CREATE TABLE IF NOT EXISTS health_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    data_type health_data_type NOT NULL,
    value DECIMAL(10,2) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    source health_data_source NOT NULL DEFAULT 'manual_input',
    device_id VARCHAR(100),
    metadata JSONB,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建健康档案表
CREATE TABLE IF NOT EXISTS health_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL,
    gender gender,
    date_of_birth DATE,
    height DECIMAL(5,2),
    blood_type blood_type,
    emergency_contact JSONB,
    medical_history TEXT[],
    allergies TEXT[],
    medications TEXT[],
    health_goals JSONB,
    preferred_units JSONB,
    notification_prefs JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建健康报告表
CREATE TABLE IF NOT EXISTS health_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    report_type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content JSONB NOT NULL,
    summary TEXT,
    recommendations TEXT[],
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建健康警报表
CREATE TABLE IF NOT EXISTS health_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    health_data_id UUID REFERENCES health_data(id),
    alert_type VARCHAR(50) NOT NULL,
    severity alert_severity NOT NULL,
    status alert_status NOT NULL DEFAULT 'active',
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
-- 健康数据索引
CREATE INDEX IF NOT EXISTS idx_health_data_user_id ON health_data(user_id);
CREATE INDEX IF NOT EXISTS idx_health_data_type ON health_data(data_type);
CREATE INDEX IF NOT EXISTS idx_health_data_recorded_at ON health_data(recorded_at);
CREATE INDEX IF NOT EXISTS idx_health_data_user_type_time ON health_data(user_id, data_type, recorded_at);
CREATE INDEX IF NOT EXISTS idx_health_data_source ON health_data(source);
CREATE INDEX IF NOT EXISTS idx_health_data_device_id ON health_data(device_id);

-- 健康档案索引
CREATE INDEX IF NOT EXISTS idx_health_profiles_user_id ON health_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_health_profiles_gender ON health_profiles(gender);
CREATE INDEX IF NOT EXISTS idx_health_profiles_blood_type ON health_profiles(blood_type);
CREATE INDEX IF NOT EXISTS idx_health_profiles_date_of_birth ON health_profiles(date_of_birth);

-- 健康报告索引
CREATE INDEX IF NOT EXISTS idx_health_reports_user_id ON health_reports(user_id);
CREATE INDEX IF NOT EXISTS idx_health_reports_type ON health_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_health_reports_generated_at ON health_reports(generated_at);
CREATE INDEX IF NOT EXISTS idx_health_reports_period ON health_reports(period_start, period_end);

-- 健康警报索引
CREATE INDEX IF NOT EXISTS idx_health_alerts_user_id ON health_alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_health_alerts_status ON health_alerts(status);
CREATE INDEX IF NOT EXISTS idx_health_alerts_severity ON health_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_health_alerts_triggered_at ON health_alerts(triggered_at);
CREATE INDEX IF NOT EXISTS idx_health_alerts_type ON health_alerts(alert_type);

-- 创建触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建更新时间触发器
CREATE TRIGGER update_health_data_updated_at 
    BEFORE UPDATE ON health_data 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_health_profiles_updated_at 
    BEFORE UPDATE ON health_profiles 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_health_alerts_updated_at 
    BEFORE UPDATE ON health_alerts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建分区表（按月分区健康数据）
CREATE TABLE IF NOT EXISTS health_data_partitioned (
    LIKE health_data INCLUDING ALL
) PARTITION BY RANGE (recorded_at);

-- 创建当前月份的分区
CREATE TABLE IF NOT EXISTS health_data_y2024m01 PARTITION OF health_data_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE IF NOT EXISTS health_data_y2024m02 PARTITION OF health_data_partitioned
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

CREATE TABLE IF NOT EXISTS health_data_y2024m03 PARTITION OF health_data_partitioned
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

-- 创建视图
-- 最新健康数据视图
CREATE OR REPLACE VIEW latest_health_data AS
SELECT DISTINCT ON (user_id, data_type) 
    id, user_id, data_type, value, unit, source, device_id, 
    metadata, recorded_at, created_at, updated_at
FROM health_data
ORDER BY user_id, data_type, recorded_at DESC;

-- 用户健康概览视图
CREATE OR REPLACE VIEW user_health_overview AS
SELECT 
    hp.user_id,
    hp.gender,
    hp.date_of_birth,
    EXTRACT(YEAR FROM AGE(hp.date_of_birth)) as age,
    hp.height,
    hp.blood_type,
    COUNT(DISTINCT hd.data_type) as data_types_count,
    MAX(hd.recorded_at) as last_data_recorded,
    COUNT(ha.id) FILTER (WHERE ha.status = 'active') as active_alerts_count
FROM health_profiles hp
LEFT JOIN health_data hd ON hp.user_id = hd.user_id
LEFT JOIN health_alerts ha ON hp.user_id = ha.user_id
GROUP BY hp.user_id, hp.gender, hp.date_of_birth, hp.height, hp.blood_type;

-- 健康数据统计视图
CREATE OR REPLACE VIEW health_data_statistics AS
SELECT 
    user_id,
    data_type,
    COUNT(*) as total_records,
    AVG(value) as avg_value,
    MIN(value) as min_value,
    MAX(value) as max_value,
    STDDEV(value) as std_deviation,
    MIN(recorded_at) as first_recorded,
    MAX(recorded_at) as last_recorded
FROM health_data
GROUP BY user_id, data_type;

-- 创建存储过程
-- 清理过期数据的存储过程
CREATE OR REPLACE FUNCTION cleanup_expired_data(retention_days INTEGER DEFAULT 365)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM health_data 
    WHERE recorded_at < CURRENT_DATE - INTERVAL '1 day' * retention_days;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 生成健康报告的存储过程
CREATE OR REPLACE FUNCTION generate_health_report(
    p_user_id UUID,
    p_start_date TIMESTAMP WITH TIME ZONE,
    p_end_date TIMESTAMP WITH TIME ZONE
)
RETURNS UUID AS $$
DECLARE
    report_id UUID;
    report_content JSONB;
BEGIN
    -- 生成报告内容
    SELECT jsonb_build_object(
        'summary', jsonb_build_object(
            'total_records', COUNT(*),
            'data_types', array_agg(DISTINCT data_type),
            'avg_values', jsonb_object_agg(data_type, avg_value)
        ),
        'trends', jsonb_build_object(
            'heart_rate_trend', 'stable',
            'weight_trend', 'decreasing'
        )
    ) INTO report_content
    FROM (
        SELECT 
            data_type,
            AVG(value) as avg_value
        FROM health_data
        WHERE user_id = p_user_id 
        AND recorded_at BETWEEN p_start_date AND p_end_date
        GROUP BY data_type
    ) stats;
    
    -- 插入报告
    INSERT INTO health_reports (
        user_id, report_type, title, content, 
        generated_at, period_start, period_end
    ) VALUES (
        p_user_id, 'weekly', '周健康报告', report_content,
        CURRENT_TIMESTAMP, p_start_date, p_end_date
    ) RETURNING id INTO report_id;
    
    RETURN report_id;
END;
$$ LANGUAGE plpgsql;

-- 插入示例数据
INSERT INTO health_profiles (user_id, gender, date_of_birth, height, blood_type) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'male', '1990-01-15', 175.5, 'A+'),
('550e8400-e29b-41d4-a716-446655440002', 'female', '1985-03-22', 162.0, 'B+'),
('550e8400-e29b-41d4-a716-446655440003', 'male', '1992-07-08', 180.2, 'O+')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO health_data (user_id, data_type, value, unit, source, recorded_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'heart_rate', 72, 'bpm', 'smart_watch', CURRENT_TIMESTAMP - INTERVAL '1 hour'),
('550e8400-e29b-41d4-a716-446655440001', 'weight', 70.5, 'kg', 'manual_input', CURRENT_TIMESTAMP - INTERVAL '2 hours'),
('550e8400-e29b-41d4-a716-446655440002', 'heart_rate', 68, 'bpm', 'fitness_tracker', CURRENT_TIMESTAMP - INTERVAL '30 minutes'),
('550e8400-e29b-41d4-a716-446655440002', 'blood_pressure', 120, 'mmHg', 'medical_device', CURRENT_TIMESTAMP - INTERVAL '1 day'),
('550e8400-e29b-41d4-a716-446655440003', 'steps', 8500, 'steps', 'mobile_app', CURRENT_TIMESTAMP - INTERVAL '3 hours');

-- 创建用户和权限
CREATE USER health_app_user WITH PASSWORD 'health_app_password';
GRANT CONNECT ON DATABASE health_management TO health_app_user;
GRANT USAGE ON SCHEMA public TO health_app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO health_app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO health_app_user;

-- 设置默认权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO health_app_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO health_app_user;

-- 创建备份用户
CREATE USER health_backup_user WITH PASSWORD 'health_backup_password';
GRANT CONNECT ON DATABASE health_management TO health_backup_user;
GRANT USAGE ON SCHEMA public TO health_backup_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO health_backup_user;

COMMIT;