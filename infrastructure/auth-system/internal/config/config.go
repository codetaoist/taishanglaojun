package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config 应用户配置
type Config struct {
	// 服务器配�?
	Server ServerConfig `json:"server"`
	
	// 数量据库配�?
	Database DatabaseConfig `json:"database"`
	
	// Redis配置
	Redis RedisConfig `json:"redis"`
	
	// JWT配置
	JWT JWTConfig `json:"jwt"`
	
	// 日志配置
	Log LogConfig `json:"log"`
	
	// 邮箱件配置
	Email EmailConfig `json:"email"`
	
	// 安全配置
	Security SecurityConfig `json:"security"`
}

// ServerConfig 服务器配�?
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Mode         string        `json:"mode"` // debug, release, test
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig 数量据库配�?
type DatabaseConfig struct {
	Type         string        `json:"type"`         // postgres, mysql, sqlite, sqlserver, oracle
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	Database     string        `json:"database"`
	Path         string        `json:"path"`         // SQLite数量据库文件路由�?
	SSLMode      string        `json:"ssl_mode"`
	Charset      string        `json:"charset"`      // MySQL字符�?
	Collation    string        `json:"collation"`    // MySQL排序规则
	TimeZone     string        `json:"timezone"`     // 时区设置
	MaxOpenConns int           `json:"max_open_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey        string        `json:"secret_key"`
	AccessTokenTTL   time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `json:"refresh_token_ttl"`
	Issuer           string        `json:"issuer"`
	RefreshThreshold time.Duration `json:"refresh_threshold"`
	RequireAudience  bool          `json:"require_audience"`
	AllowedAudiences []string      `json:"allowed_audiences"`
	MaxTokenAge      time.Duration `json:"max_token_age"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`      // debug, info, warn, error
	Format     string `json:"format"`     // json, console
	Output     string `json:"output"`     // stdout, stderr, file
	Filename   string `json:"filename"`   // 日志文件�?
	MaxSize    int    `json:"max_size"`   // MB
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`    // days
	Compress   bool   `json:"compress"`
}

// EmailConfig 邮箱件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	UseTLS       bool   `json:"use_tls"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	PasswordMinLength    int           `json:"password_min_length"`
	MaxLoginAttempts     int           `json:"max_login_attempts"`
	LockoutDuration      time.Duration `json:"lockout_duration"`
	SessionTimeout       time.Duration `json:"session_timeout"`
	TokenCleanupInterval time.Duration `json:"token_cleanup_interval"`
	RateLimitRequests    int           `json:"rate_limit_requests"`
	RateLimitWindow      time.Duration `json:"rate_limit_window"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在不是错�?
		zap.L().Debug("No .env file found")
	}

	config := &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvAsInt("SERVER_PORT", 8082),
			Mode:         getEnv("SERVER_MODE", "debug"),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Type:         getEnv("DB_TYPE", "postgres"),
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			Username:     getEnv("DB_USERNAME", "postgres"),
			Password:     getEnv("DB_PASSWORD", "password"),
			Database:     getEnv("DB_DATABASE", "auth_system"),
			Path:         getEnv("DB_PATH", "./data/auth_system.db"),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
			Charset:      getEnv("DB_CHARSET", "utf8mb4"),
			Collation:    getEnv("DB_COLLATION", "utf8mb4_unicode_ci"),
			TimeZone:     getEnv("DB_TIMEZONE", "UTC"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvAsInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getEnvAsInt("REDIS_DATABASE", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
			DialTimeout:  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
			IdleTimeout:  getEnvAsDuration("REDIS_IDLE_TIMEOUT", 5*time.Minute),
		},
		JWT: JWTConfig{
			SecretKey:        getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			AccessTokenTTL:   getEnvAsDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:  getEnvAsDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
			Issuer:           getEnv("JWT_ISSUER", "taishang-auth-system"),
			RefreshThreshold: getEnvAsDuration("JWT_REFRESH_THRESHOLD", 5*time.Minute),
			RequireAudience:  getEnvAsBool("JWT_REQUIRE_AUDIENCE", true),
			AllowedAudiences: getEnvAsStringSlice("JWT_ALLOWED_AUDIENCES", []string{"taishang-system", "taishang-web", "taishang-mobile"}),
			MaxTokenAge:      getEnvAsDuration("JWT_MAX_TOKEN_AGE", 24*time.Hour),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			Filename:   getEnv("LOG_FILENAME", "auth-system.log"),
			MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvAsInt("LOG_MAX_AGE", 28),
			Compress:   getEnvAsBool("LOG_COMPRESS", true),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@example.com"),
			FromName:     getEnv("FROM_NAME", "Auth System"),
			UseTLS:       getEnvAsBool("SMTP_USE_TLS", true),
		},
		Security: SecurityConfig{
			PasswordMinLength:    getEnvAsInt("PASSWORD_MIN_LENGTH", 8),
			MaxLoginAttempts:     getEnvAsInt("MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:      getEnvAsDuration("LOCKOUT_DURATION", 15*time.Minute),
			SessionTimeout:       getEnvAsDuration("SESSION_TIMEOUT", 30*24*time.Hour),
			TokenCleanupInterval: getEnvAsDuration("TOKEN_CLEANUP_INTERVAL", 1*time.Hour),
			RateLimitRequests:    getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:      getEnvAsDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	if c.JWT.SecretKey == "" || c.JWT.SecretKey == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT secret key must be set and changed from default")
	}

	if c.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT access token TTL must be positive")
	}

	if c.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("JWT refresh token TTL must be positive")
	}

	return nil
}

// GetDSN 获取数量据库连接字符串
func (c *Config) GetDSN() string {
	switch c.Database.Type {
	case "sqlite":
		return c.Database.Path
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&parseTime=true&loc=%s",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
			c.Database.Charset,
			c.Database.Collation,
			c.Database.TimeZone,
		)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Database,
			c.Database.SSLMode,
			c.Database.TimeZone,
		)
	case "sqlserver":
		return fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			c.Database.Username,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.Database,
		)
	default:
		// 默认返回PostgreSQL格式
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Database,
			c.Database.SSLMode,
		)
	}
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsProduction 是否为生产环�?
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}

// IsDevelopment 是否为开发环�?
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "debug"
}

// IsTest 是否为测试环�?
func (c *Config) IsTest() bool {
	return c.Server.Mode == "test"
}

// 辅助函数量

// getEnv 获取环境变量，如果不存在则返回默认�?
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数量
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔�?
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration 获取环境变量并转换为时间间隔
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvAsStringSlice 获取环境变量并转换为字符串切片
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// 按逗号分割字符串
		parts := make([]string, 0)
		for _, part := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return defaultValue
}
