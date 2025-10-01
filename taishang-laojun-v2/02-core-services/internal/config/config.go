package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	AI       AIConfig       `mapstructure:"ai"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name           string   `mapstructure:"name"`
	Version        string   `mapstructure:"version"`
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	Timeout        int      `mapstructure:"timeout"`
	MaxBodySize    int      `mapstructure:"max_body_size"`
	EnableCORS     bool     `mapstructure:"enable_cors"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	EnableLog      bool     `mapstructure:"enable_request_log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Primary  PostgresConfig `mapstructure:"primary"`
	ReadOnly PostgresConfig `mapstructure:"readonly"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AIConfig AI服务配置
type AIConfig struct {
	Providers []AIProviderConfig `mapstructure:"providers"`
	Default   string             `mapstructure:"default"`
	Timeout   time.Duration      `mapstructure:"timeout"`
}

// AIProviderConfig AI提供商配置
type AIProviderConfig struct {
	Name     string            `mapstructure:"name"`
	Type     string            `mapstructure:"type"`
	APIKey   string            `mapstructure:"api_key"`
	BaseURL  string            `mapstructure:"base_url"`
	Model    string            `mapstructure:"model"`
	Settings map[string]string `mapstructure:"settings"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")

	// 设置环境变量前缀
	viper.SetEnvPrefix("CORE")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，使用默认值和环境变量
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// 从环境变量覆盖敏感配置
	overrideFromEnv(&config)

	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.name", "taishang-laojun-core-services")
	viper.SetDefault("server.version", "1.0.0")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "development")
	viper.SetDefault("server.timeout", 30)
	viper.SetDefault("server.max_body_size", 10)
	viper.SetDefault("server.enable_cors", true)
	viper.SetDefault("server.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("server.enable_request_log", true)

	// 数据库默认配置
	viper.SetDefault("database.primary.driver", "postgres")
	viper.SetDefault("database.primary.host", "localhost")
	viper.SetDefault("database.primary.port", 5432)
	viper.SetDefault("database.primary.database", "taishanglaojun")
	viper.SetDefault("database.primary.username", "postgres")
	viper.SetDefault("database.primary.password", "password")
	viper.SetDefault("database.primary.max_open_conns", 25)
	viper.SetDefault("database.primary.max_idle_conns", 10)
	viper.SetDefault("database.primary.conn_max_lifetime", "1h")
	viper.SetDefault("database.primary.ssl_mode", "disable")
	viper.SetDefault("database.primary.connect_timeout", "10s")

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 2)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")

	// AI默认配置
	viper.SetDefault("ai.default", "openai")
	viper.SetDefault("ai.timeout", "30s")

	// 日志默认配置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 7)
	viper.SetDefault("log.compress", true)
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(config *Config) {
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Primary.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		viper.Set("database.primary.port", port)
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		config.Database.Primary.Database = dbname
	}
	if username := os.Getenv("DB_USER"); username != "" {
		config.Database.Primary.Username = username
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Primary.Password = password
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Redis.Password = password
	}

	// AI配置
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		for i := range config.AI.Providers {
			if config.AI.Providers[i].Type == "openai" {
				config.AI.Providers[i].APIKey = apiKey
			}
		}
	}
}

// GetDSN 获取数据库连接字符串
func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}