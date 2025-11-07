package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the gateway
type Config struct {
	Port        string                   `mapstructure:"port"`
	Environment string                   `mapstructure:"environment"`
	LogLevel    string                   `mapstructure:"log_level"`
	JWT         JWTConfig                `mapstructure:"jwt"`
	Discovery   DiscoveryConfig          `mapstructure:"discovery"`
	Proxy       ProxyConfig              `mapstructure:"proxy"`
	CORS        CORSConfig               `mapstructure:"cors"`
	Metrics     MetricsConfig            `mapstructure:"metrics"`
	Services    map[string]ServiceConfig `mapstructure:"services"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret         string `mapstructure:"secret"`
	ExpirationTime int    `mapstructure:"expiration_time"`
}

// DiscoveryConfig holds service discovery configuration
type DiscoveryConfig struct {
	Type       string `mapstructure:"type"` // consul, etcd, etc.
	Address    string `mapstructure:"address"`
	Datacenter string `mapstructure:"datacenter"`
	Token      string `mapstructure:"token"`
}

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Timeout             time.Duration `mapstructure:"timeout"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	RetryAttempts       int           `mapstructure:"retry_attempts"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    string `mapstructure:"port"`
}

// ServiceConfig holds configuration for a specific service
type ServiceConfig struct {
	PathPrefix       string          `mapstructure:"path_prefix"`
	AuthRequired     bool            `mapstructure:"auth_required"`
	RateLimitEnabled bool            `mapstructure:"rate_limit_enabled"`
	RateLimit        RateLimitConfig `mapstructure:"rate_limit"`
	Timeout          time.Duration   `mapstructure:"timeout"`
	RetryAttempts    int             `mapstructure:"retry_attempts"`
	LoadBalancer     string          `mapstructure:"load_balancer"` // round_robin, random, etc.
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Set default values
	setDefaults()

	// Load configuration from file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/gateway")

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GATEWAY")

	// Load configuration
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file is not found, we'll use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server configuration
	viper.SetDefault("port", "8080")
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")

	// JWT configuration
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expiration_time", 86400) // 24 hours

	// Discovery configuration
	viper.SetDefault("discovery.type", "consul")
	viper.SetDefault("discovery.address", "localhost:8500")
	viper.SetDefault("discovery.datacenter", "dc1")

	// Proxy configuration
	viper.SetDefault("proxy.timeout", "30s")
	viper.SetDefault("proxy.health_check_interval", "10s")
	viper.SetDefault("proxy.retry_attempts", 3)

	// CORS configuration
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"*"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)

	// Metrics configuration
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.port", "9090")

	// Default services configuration
	viper.SetDefault("services.api.path_prefix", "/api")
	viper.SetDefault("services.api.auth_required", false)
	viper.SetDefault("services.api.rate_limit_enabled", false)
	viper.SetDefault("services.api.timeout", "30s")
	viper.SetDefault("services.api.retry_attempts", 3)
	viper.SetDefault("services.api.load_balancer", "round_robin")

	viper.SetDefault("services.auth.path_prefix", "/auth")
	viper.SetDefault("services.auth.auth_required", false)
	viper.SetDefault("services.auth.rate_limit_enabled", true)
	viper.SetDefault("services.auth.rate_limit.requests", 10)
	viper.SetDefault("services.auth.rate_limit.window", "1m")
	viper.SetDefault("services.auth.timeout", "10s")
	viper.SetDefault("services.auth.retry_attempts", 2)
	viper.SetDefault("services.auth.load_balancer", "round_robin")

	viper.SetDefault("services.notification.path_prefix", "/notification")
	viper.SetDefault("services.notification.auth_required", true)
	viper.SetDefault("services.notification.rate_limit_enabled", true)
	viper.SetDefault("services.notification.rate_limit.requests", 5)
	viper.SetDefault("services.notification.rate_limit.window", "1m")
	viper.SetDefault("services.notification.timeout", "15s")
	viper.SetDefault("services.notification.retry_attempts", 2)
	viper.SetDefault("services.notification.load_balancer", "round_robin")
}

// Close closes the configuration and releases any resources
func (c *Config) Close() error {
	// Nothing to close for now
	return nil
}
