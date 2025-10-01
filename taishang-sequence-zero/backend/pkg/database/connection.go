package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// Config 数据库配置结构体
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewConnection 创建新的数据库连接
func NewConnection() (*sql.DB, error) {
	config := getConfigFromEnv()
	
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("数据库连接成功")
	return db, nil
}

// getConfigFromEnv 从环境变量获取数据库配置
func getConfigFromEnv() Config {
	config := Config{
		Host:     getEnv("DB_HOST", "localhost"),
		User:     getEnv("DB_USER", "taishang"),
		Password: getEnv("DB_PASSWORD", "taishang123"),
		DBName:   getEnv("DB_NAME", "taishang_sequence_zero"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		port = 5432
	}
	config.Port = port

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitializeSchema 初始化数据库架构
func InitializeSchema(db *sql.DB) error {
	// 创建用户表
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		permission_level INTEGER DEFAULT 1 CHECK (permission_level >= 1 AND permission_level <= 9),
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建权限表
	permissionTableSQL := `
	CREATE TABLE IF NOT EXISTS permissions (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) UNIQUE NOT NULL,
		description TEXT,
		level INTEGER NOT NULL CHECK (level >= 1 AND level <= 9),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建用户会话表
	sessionTableSQL := `
	CREATE TABLE IF NOT EXISTS user_sessions (
		id SERIAL PRIMARY KEY,
		session_id VARCHAR(32) UNIQUE NOT NULL,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		refresh_token VARCHAR(64) UNIQUE NOT NULL,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建意识状态表
	consciousnessTableSQL := `
	CREATE TABLE IF NOT EXISTS consciousness_states (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		state_type VARCHAR(50) NOT NULL,
		state_data JSONB NOT NULL,
		confidence_score DECIMAL(5,4) DEFAULT 0.0000,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建文化分析表
	culturalTableSQL := `
	CREATE TABLE IF NOT EXISTS cultural_analyses (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		analysis_type VARCHAR(50) NOT NULL,
		content TEXT NOT NULL,
		result JSONB NOT NULL,
		wisdom_level INTEGER DEFAULT 1 CHECK (wisdom_level >= 1 AND wisdom_level <= 9),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建审计日志表
	auditTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
		action VARCHAR(100) NOT NULL,
		details TEXT,
		ip_address INET,
		user_agent TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 执行表创建语句
	tables := []string{
		userTableSQL,
		permissionTableSQL,
		sessionTableSQL,
		consciousnessTableSQL,
		culturalTableSQL,
		auditTableSQL,
	}

	for _, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// 插入基础权限数据
	permissionsData := []struct {
		name        string
		description string
		level       int
	}{
		{"基础访问", "基础系统访问权限", 1},
		{"数据查看", "查看个人数据权限", 2},
		{"数据修改", "修改个人数据权限", 3},
		{"意识分析", "意识状态分析权限", 4},
		{"文化智慧", "文化智慧访问权限", 5},
		{"高级功能", "高级功能访问权限", 6},
		{"系统管理", "系统管理权限", 7},
		{"核心控制", "核心系统控制权限", 8},
		{"至高权限", "最高级别系统权限", 9},
	}

	for _, perm := range permissionsData {
		_, err := db.Exec(`
			INSERT INTO permissions (name, description, level) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (name) DO NOTHING`,
			perm.name, perm.description, perm.level)
		if err != nil {
			return fmt.Errorf("failed to insert permission data: %w", err)
		}
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_users_permission_level ON users(permission_level);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON user_sessions(expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_consciousness_user_id ON consciousness_states(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_consciousness_type ON consciousness_states(state_type);",
		"CREATE INDEX IF NOT EXISTS idx_cultural_user_id ON cultural_analyses(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_cultural_type ON cultural_analyses(analysis_type);",
		"CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_logs(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);",
		"CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	fmt.Println("数据库架构初始化完成")
	return nil
}