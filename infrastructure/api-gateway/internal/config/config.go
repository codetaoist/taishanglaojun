package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Server        ServerConfig                       `yaml:"server" mapstructure:"server"`
	Log           LogConfig                          `yaml:"log" mapstructure:"log"`
	Monitoring    MonitoringConfig                   `yaml:"monitoring" mapstructure:"monitoring"`
	Registry      RegistryConfig                     `yaml:"registry" mapstructure:"registry"`
	Proxy         ProxyConfig                        `yaml:"proxy" mapstructure:"proxy"`
	Security      SecurityConfig                     `yaml:"security" mapstructure:"security"`
	RateLimit     RateLimitConfig                    `yaml:"rate_limit" mapstructure:"rate_limit"`
	HealthCheck   HealthCheckConfig                  `yaml:"health_check" mapstructure:"health_check"`
	Services      []ServiceConfig                    `yaml:"services" mapstructure:"services"`
	StaticServices map[string][]StaticServiceInstance `yaml:"static_services" mapstructure:"static_services"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port" mapstructure:"port"`
	Mode         string        `yaml:"mode" mapstructure:"mode"`
	Debug        bool          `yaml:"debug" mapstructure:"debug"`
	ReadTimeout  time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"`
	Output     string `yaml:"output" mapstructure:"output"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`
	Compress   bool   `yaml:"compress" mapstructure:"compress"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Enabled     bool   `yaml:"enabled" mapstructure:"enabled"`
	Port        int    `yaml:"port" mapstructure:"port"`
	MetricsPath string `yaml:"metrics_path" mapstructure:"metrics_path"`
}

// RegistryConfig 服务注册配置
type RegistryConfig struct {
	Type     string            `yaml:"type" mapstructure:"type"`
	Endpoints []string         `yaml:"endpoints" mapstructure:"endpoints"`
	Options  map[string]string `yaml:"options" mapstructure:"options"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Timeout         time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout" mapstructure:"idle_conn_timeout"`
	LoadBalancer    string        `yaml:"load_balancer" mapstructure:"load_balancer"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Auth          AuthConfig `yaml:"auth" mapstructure:"auth"`
	CORSOrigins   []string   `yaml:"cors_origins" mapstructure:"cors_origins"`
	CORSMethods   []string   `yaml:"cors_methods" mapstructure:"cors_methods"`
	CORSHeaders   []string   `yaml:"cors_headers" mapstructure:"cors_headers"`
	TrustedProxies []string  `yaml:"trusted_proxies" mapstructure:"trusted_proxies"`
	
	// 保持向后兼容性
	JWTSecret   string `yaml:"jwt_secret" mapstructure:"jwt_secret"`
	AuthService string `yaml:"auth_service" mapstructure:"auth_service"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret     string        `yaml:"jwt_secret" mapstructure:"jwt_secret"`
	TokenExpiry   time.Duration `yaml:"token_expiry" mapstructure:"token_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry" mapstructure:"refresh_expiry"`
	RedisAddr     string        `yaml:"redis_addr" mapstructure:"redis_addr"`
	RedisPassword string        `yaml:"redis_password" mapstructure:"redis_password"`
	RedisDB       int           `yaml:"redis_db" mapstructure:"redis_db"`
	SkipPaths     []string      `yaml:"skip_paths" mapstructure:"skip_paths"`
	OptionalPaths []string      `yaml:"optional_paths" mapstructure:"optional_paths"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled     bool `yaml:"enabled" mapstructure:"enabled"`
	DefaultRate int  `yaml:"default_rate" mapstructure:"default_rate"`
	DefaultBurst int `yaml:"default_burst" mapstructure:"default_burst"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled" mapstructure:"enabled"`
	Interval time.Duration `yaml:"interval" mapstructure:"interval"`
	Timeout  time.Duration `yaml:"timeout" mapstructure:"timeout"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name           string                 `yaml:"name" mapstructure:"name"`
	URL            string                 `yaml:"url" mapstructure:"url"`
	HealthCheck    string                 `yaml:"health_check" mapstructure:"health_check"`
	Weight         int                    `yaml:"weight" mapstructure:"weight"`
	Routes         []RouteConfig          `yaml:"routes" mapstructure:"routes"`
	Middleware     []string               `yaml:"middleware" mapstructure:"middleware"`
	RateLimit      *ServiceRateLimit      `yaml:"rate_limit,omitempty" mapstructure:"rate_limit"`
	CircuitBreaker *CircuitBreakerConfig  `yaml:"circuit_breaker,omitempty" mapstructure:"circuit_breaker"`
	LoadBalancer   string                 `yaml:"load_balancer" mapstructure:"load_balancer"`
	Retry          *RetryConfig           `yaml:"retry,omitempty" mapstructure:"retry"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Path         string            `yaml:"path" mapstructure:"path"`
	Method       string            `yaml:"method" mapstructure:"method"`
	Rewrite      string            `yaml:"rewrite" mapstructure:"rewrite"`
	StripPrefix  bool              `yaml:"strip_prefix" mapstructure:"strip_prefix"`
	Timeout      int               `yaml:"timeout" mapstructure:"timeout"`
	Middleware   []string          `yaml:"middleware" mapstructure:"middleware"`
	Headers      map[string]string `yaml:"headers" mapstructure:"headers"`
}

// ServiceRateLimit 服务级别限流配置
type ServiceRateLimit struct {
	Rate  int `yaml:"rate" mapstructure:"rate"`
	Burst int `yaml:"burst" mapstructure:"burst"`
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Enabled           bool    `yaml:"enabled" mapstructure:"enabled"`
	Threshold         int     `yaml:"threshold" mapstructure:"threshold"`
	FailureRate       float64 `yaml:"failure_rate" mapstructure:"failure_rate"`
	RecoveryTimeout   int     `yaml:"recovery_timeout" mapstructure:"recovery_timeout"`
	HalfOpenRequests  int     `yaml:"half_open_requests" mapstructure:"half_open_requests"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled     bool          `yaml:"enabled" mapstructure:"enabled"`
	MaxAttempts int           `yaml:"max_attempts" mapstructure:"max_attempts"`
	Delay       time.Duration `yaml:"delay" mapstructure:"delay"`
	BackoffRate float64       `yaml:"backoff_rate" mapstructure:"backoff_rate"`
}

// StaticServiceInstance 静态服务实例配置
type StaticServiceInstance struct {
	ID      string            `yaml:"id" mapstructure:"id"`
	Address string            `yaml:"address" mapstructure:"address"`
	Port    int               `yaml:"port" mapstructure:"port"`
	Weight  int               `yaml:"weight" mapstructure:"weight"`
	Tags    []string          `yaml:"tags" mapstructure:"tags"`
	Meta    map[string]string `yaml:"meta" mapstructure:"meta"`
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{}

	// 设置默认值
	setDefaults(cfg)

	// 从环境变量加载
	loadFromEnv(cfg)

	// 从配置文件加载
	if err := loadFromFile(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults 设置默认值
func setDefaults(cfg *Config) {
	cfg.Server = ServerConfig{
		Port:         8081,
		Mode:         "development",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	cfg.Log = LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	cfg.Monitoring = MonitoringConfig{
		Enabled:     true,
		Port:        9090,
		MetricsPath: "/metrics",
	}

	cfg.Registry = RegistryConfig{
		Type:      "static",
		Endpoints: []string{},
		Options:   make(map[string]string),
	}

	cfg.Proxy = ProxyConfig{
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
		LoadBalancer:    "round_robin",
	}

	cfg.Security = SecurityConfig{
		Auth: AuthConfig{
			JWTSecret:     "your-secret-key",
			TokenExpiry:   24 * time.Hour,
			RefreshExpiry: 7 * 24 * time.Hour,
			RedisAddr:     "localhost:6379",
			RedisPassword: "",
			RedisDB:       0,
			SkipPaths:     []string{"/health", "/ready", "/metrics"},
			OptionalPaths: []string{},
		},
		CORSOrigins:   []string{"*"},
		CORSMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		TrustedProxies: []string{"127.0.0.1"},
		// 保持向后兼容性
		JWTSecret:   "your-secret-key",
		AuthService: "http://auth-service:8081",
	}

	cfg.RateLimit = RateLimitConfig{
		Enabled:     true,
		DefaultRate: 100,
		DefaultBurst: 10,
	}

	cfg.HealthCheck = HealthCheckConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  5 * time.Second,
	}
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	if port := getEnv("GATEWAY_PORT", ""); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	if mode := getEnv("GATEWAY_MODE", ""); mode != "" {
		cfg.Server.Mode = mode
	}

	if level := getEnv("LOG_LEVEL", ""); level != "" {
		cfg.Log.Level = level
	}

	if secret := getEnv("JWT_SECRET", ""); secret != "" {
		cfg.Security.Auth.JWTSecret = secret
		cfg.Security.JWTSecret = secret // 保持向后兼容性
	}

	if authService := getEnv("AUTH_SERVICE_URL", ""); authService != "" {
		cfg.Security.AuthService = authService
	}

	if origins := getEnv("CORS_ORIGINS", ""); origins != "" {
		cfg.Security.CORSOrigins = strings.Split(origins, ",")
	}
}

// loadFromFile 从配置文件加载
func loadFromFile(cfg *Config) error {
	viper.SetConfigName("gateway")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/gateway")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			return nil
		}
		return err
	}

	return viper.Unmarshal(cfg)
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	// 优先使用嵌套的Auth配置，如果为空则使用向后兼容的字段
	jwtSecret := c.Security.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = c.Security.JWTSecret
	}
	
	if jwtSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	authService := c.Security.AuthService
	if authService == "" {
		return fmt.Errorf("auth service URL is required")
	}

	// 验证服务配置
	for i, service := range c.Services {
		if service.Name == "" {
			return fmt.Errorf("service[%d]: name is required", i)
		}
		if service.URL == "" {
			return fmt.Errorf("service[%d]: URL is required", i)
		}
		if service.Weight <= 0 {
			c.Services[i].Weight = 1
		}
	}

	return nil
}

// GetServerAddr 获取服务器地址
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

// IsProduction 是否生产环境
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "production"
}

// IsDevelopment 是否开发环境
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "development"
}

// ToYAML 导出为YAML格式
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}