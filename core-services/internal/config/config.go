package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	CORS     CORSConfig     `mapstructure:"cors"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AI       AIConfig       `mapstructure:"ai"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name           string   `mapstructure:"name"`
	Version        string   `mapstructure:"version"`
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	ReadTimeout    int      `mapstructure:"read_timeout"`
	WriteTimeout   int      `mapstructure:"write_timeout"`
	MaxHeaderBytes int      `mapstructure:"max_header_bytes"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type         string `mapstructure:"type"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	Database     int    `mapstructure:"database"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	Enabled      bool   `mapstructure:"enabled"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	ExpiresIn        time.Duration `mapstructure:"expires_in"`
	RefreshExpiresIn time.Duration `mapstructure:"refresh_expires_in"`
	Issuer           string        `mapstructure:"issuer"`
	Audience         string        `mapstructure:"audience"`
}

// AIConfig AI配置
type AIConfig struct {
	Providers map[string]ProviderConfig `mapstructure:"providers"`
}

// ProviderConfig AI提供商配置
type ProviderConfig struct {
	Name    string `mapstructure:"name"`
	Enabled bool   `mapstructure:"enabled"`
	Config  map[string]interface{} `mapstructure:"config"`
}

var globalConfig *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("../config")
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix("TAISHANG")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 手动处理环境变量替换
	if config.JWT.Secret == "" || strings.Contains(config.JWT.Secret, "${") {
		config.JWT.Secret = os.ExpandEnv(config.JWT.Secret)
		if config.JWT.Secret == "" {
			config.JWT.Secret = "your-super-secret-jwt-key-change-in-production"
		}
	}

	globalConfig = &config
	return &config, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.name", "taishang-core-services")
	viper.SetDefault("server.version", "1.0.0")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.write_timeout", 60)
	viper.SetDefault("server.max_header_bytes", 1048576)

	// 数据库默认配置
	viper.SetDefault("database.type", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "taishang")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 3600)

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// 日志默认配置
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output", "stdout")
	viper.SetDefault("logger.max_size", 100)
	viper.SetDefault("logger.max_age", 30)
	viper.SetDefault("logger.max_backups", 10)
	viper.SetDefault("logger.compress", true)

	// CORS默认配置
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"*"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)

	// JWT默认配置
	viper.SetDefault("jwt.secret", "taishang-secret-key")
	viper.SetDefault("jwt.issuer", "taishang-core-services")
	viper.SetDefault("jwt.expiration", "24h")
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		return configFile
	}
	
	// 查找配置文件
	paths := []string{
		"config.yaml",
		"./config/config.yaml",
		"../config/config.yaml",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}
	
	return "config.yaml"
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", c.Server.Port)
	}
	
	if c.Database.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}
	
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}
	
	return nil
}

