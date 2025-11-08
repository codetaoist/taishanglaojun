package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// ServiceConfig represents configuration for a service
type ServiceConfig struct {
	Name     string            `mapstructure:"name"`
	Version  string            `mapstructure:"version"`
	Address  string            `mapstructure:"address"`
	Port     int               `mapstructure:"port"`
	Protocol string            `mapstructure:"protocol"`
	Tags     []string          `mapstructure:"tags"`
	Metadata map[string]string `mapstructure:"metadata"`
}

// DatabaseConfig represents configuration for database connections
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig represents configuration for Redis connections
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AIConfig represents configuration for AI services
type AIConfig struct {
	VectorService ServiceConfig `mapstructure:"vector_service"`
	ModelService  ServiceConfig `mapstructure:"model_service"`
}

// PluginConfig represents configuration for the plugin system
type PluginConfig struct {
	Directory string `mapstructure:"directory"`
	Repository struct {
		URL    string `mapstructure:"url"`
		Branch string `mapstructure:"branch"`
	} `mapstructure:"repository"`
	AutoBuild     bool `mapstructure:"auto_build"`
	Versioning    bool `mapstructure:"versioning"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
	Timeout       time.Duration `mapstructure:"timeout"`
	ResourceLimits struct {
		Memory string `mapstructure:"memory"`
		CPU    string `mapstructure:"cpu"`
	} `mapstructure:"resource_limits"`
}

// CIConfig represents configuration for CI/CD pipeline
type CIConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Timeout time.Duration `mapstructure:"timeout"`
	ArtifactRetention time.Duration `mapstructure:"artifact_retention"`
	RegistryURL string `mapstructure:"registry_url"`
	DefaultPipelineTemplate string `mapstructure:"default_pipeline_template"`
	
	// Build configuration
	BuildDir     string            `mapstructure:"build_dir"`
	ArtifactDir  string            `mapstructure:"artifact_dir"`
	Dockerfile   string            `mapstructure:"dockerfile"`
	ImageName    string            `mapstructure:"image_name"`
	ImageTag     string            `mapstructure:"image_tag"`
	BuildArgs    map[string]string `mapstructure:"build_args"`
	Environment  map[string]string `mapstructure:"environment"`
	CacheEnabled bool              `mapstructure:"cache_enabled"`
	CacheDir     string            `mapstructure:"cache_dir"`
	
	// Test configuration
	TestDir      string            `mapstructure:"test_dir"`
	TestPattern  string            `mapstructure:"test_pattern"`
	Coverage     bool              `mapstructure:"coverage"`
	CoverageFile string            `mapstructure:"coverage_file"`
	TestEnvironment map[string]string `mapstructure:"test_environment"`
	TestTimeout  time.Duration     `mapstructure:"test_timeout"`
	Parallel     bool              `mapstructure:"parallel"`
	Verbose      bool              `mapstructure:"verbose"`
	
	// Deploy configuration
	DeployEnvironment string            `mapstructure:"deploy_environment"`
	DeployNamespace   string            `mapstructure:"deploy_namespace"`
	KubeConfig        string            `mapstructure:"kube_config"`
	Manifests         []string          `mapstructure:"manifests"`
	HelmChart         string            `mapstructure:"helm_chart"`
	HelmValues        string            `mapstructure:"helm_values"`
	HelmRelease       string            `mapstructure:"helm_release"`
	DeployWait        bool              `mapstructure:"deploy_wait"`
	DeployTimeout     time.Duration     `mapstructure:"deploy_timeout"`
	Variables         map[string]string `mapstructure:"variables"`
	
	// Build environment
	BuildEnvironment struct {
		DockerImage string `mapstructure:"docker_image"`
		Workdir     string `mapstructure:"workdir"`
	} `mapstructure:"build_environment"`
}

// HybridConfig represents configuration for the hybrid architecture
type HybridConfig struct {
	ServiceRegistry struct {
		Type string `mapstructure:"type"` // "in-memory", "consul", "etcd", etc.
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"service_registry"`

	HealthCheck struct {
		Interval time.Duration `mapstructure:"interval"`
		Timeout  time.Duration `mapstructure:"timeout"`
	} `mapstructure:"health_check"`

	CircuitBreaker struct {
		MaxRequests    int           `mapstructure:"max_requests"`
		Interval       time.Duration `mapstructure:"interval"`
		Timeout        time.Duration `mapstructure:"timeout"`
		ReadyToTrip    float64       `mapstructure:"ready_to_trip"`
		OnStateChange  bool          `mapstructure:"on_state_change"`
	} `mapstructure:"circuit_breaker"`

	RateLimit struct {
		RequestsPerMinute int `mapstructure:"requests_per_minute"`
		BurstSize         int `mapstructure:"burst_size"`
	} `mapstructure:"rate_limit"`

	Tracing struct {
		Enabled    bool   `mapstructure:"enabled"`
		ServiceName string `mapstructure:"service_name"`
		Jaeger struct {
			Endpoint string `mapstructure:"endpoint"`
		} `mapstructure:"jaeger"`
	} `mapstructure:"tracing"`

	Metrics struct {
		Enabled bool   `mapstructure:"enabled"`
		Port    int    `mapstructure:"port"`
		Path    string `mapstructure:"path"`
	} `mapstructure:"metrics"`

	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"` // "json", "text"
	} `mapstructure:"logging"`
}

// Config represents the application configuration
type Config struct {
	Server struct {
		Host         string        `mapstructure:"host"`
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	} `mapstructure:"server"`

	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	AI       AIConfig       `mapstructure:"ai"`
	Plugin   PluginConfig   `mapstructure:"plugin"`
	CI       CIConfig       `mapstructure:"ci"`
	Hybrid   HybridConfig   `mapstructure:"hybrid"`

	Security struct {
		JWTSecret     string `mapstructure:"jwt_secret"`
		ExpiresIn     int    `mapstructure:"expires_in"` // in hours
		RefreshExpiresIn int  `mapstructure:"refresh_expires_in"` // in hours
	} `mapstructure:"security"`

	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Set default values
	setDefaults()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file doesn't exist, use environment variables and defaults
		viper.AutomaticEnv()
		if err := viper.Unmarshal(config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		return config, nil
	}

	// Load config from file
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TAOIST")

	// Replace dots with underscores in env variable names
	viper.SetEnvKeyReplacer(viper.NewKeyReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "taoist")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// AI service defaults
	viper.SetDefault("ai.vector_service.name", "vector-service")
	viper.SetDefault("ai.vector_service.version", "1.0.0")
	viper.SetDefault("ai.vector_service.address", "localhost")
	viper.SetDefault("ai.vector_service.port", 50051)
	viper.SetDefault("ai.vector_service.protocol", "grpc")
	viper.SetDefault("ai.vector_service.tags", []string{"ai", "vector", "database"})

	viper.SetDefault("ai.model_service.name", "model-service")
	viper.SetDefault("ai.model_service.version", "1.0.0")
	viper.SetDefault("ai.model_service.address", "localhost")
	viper.SetDefault("ai.model_service.port", 50051)
	viper.SetDefault("ai.model_service.protocol", "grpc")
	viper.SetDefault("ai.model_service.tags", []string{"ai", "model", "inference"})

	// Plugin system defaults
	viper.SetDefault("plugin.directory", "./plugins")
	viper.SetDefault("plugin.repository.url", "https://github.com/codetaoist/plugins")
	viper.SetDefault("plugin.repository.branch", "main")
	viper.SetDefault("plugin.auto_build", true)
	viper.SetDefault("plugin.versioning", true)
	viper.SetDefault("plugin.update_interval", "24h")
	viper.SetDefault("plugin.timeout", "10m")
	viper.SetDefault("plugin.resource_limits.memory", "512Mi")
	viper.SetDefault("plugin.resource_limits.cpu", "500m")

	// CI/CD defaults
	viper.SetDefault("ci.enabled", true)
	viper.SetDefault("ci.timeout", "30m")
	viper.SetDefault("ci.artifact_retention", "7d")
	viper.SetDefault("ci.registry_url", "docker.io/codetaoist")
	viper.SetDefault("ci.default_pipeline_template", "default")
	
	// Build defaults
	viper.SetDefault("ci.build_dir", "./build")
	viper.SetDefault("ci.artifact_dir", "./artifacts")
	viper.SetDefault("ci.dockerfile", "Dockerfile")
	viper.SetDefault("ci.image_name", "app")
	viper.SetDefault("ci.image_tag", "latest")
	viper.SetDefault("ci.cache_enabled", true)
	viper.SetDefault("ci.cache_dir", "./cache")
	
	// Test defaults
	viper.SetDefault("ci.test_dir", "./tests")
	viper.SetDefault("ci.test_pattern", "*_test.go")
	viper.SetDefault("ci.coverage", true)
	viper.SetDefault("ci.coverage_file", "coverage.out")
	viper.SetDefault("ci.test_timeout", "10m")
	viper.SetDefault("ci.parallel", false)
	viper.SetDefault("ci.verbose", false)
	
	// Deploy defaults
	viper.SetDefault("ci.deploy_environment", "development")
	viper.SetDefault("ci.deploy_namespace", "default")
	viper.SetDefault("ci.deploy_wait", true)
	viper.SetDefault("ci.deploy_timeout", "10m")
	
	// Build environment defaults
	viper.SetDefault("ci.build_environment.docker_image", "golang:1.19-alpine")
	viper.SetDefault("ci.build_environment.workdir", "/workspace")
	viper.SetDefault("ci.helm_values", "./helm/values.yaml")
	viper.SetDefault("ci.helm_release", "taoist-api")

	// Hybrid architecture defaults
	viper.SetDefault("hybrid.service_registry.type", "in-memory")
	viper.SetDefault("hybrid.service_registry.host", "localhost")
	viper.SetDefault("hybrid.service_registry.port", 8500)

	viper.SetDefault("hybrid.health_check.interval", "30s")
	viper.SetDefault("hybrid.health_check.timeout", "5s")

	viper.SetDefault("hybrid.circuit_breaker.max_requests", 5)
	viper.SetDefault("hybrid.circuit_breaker.interval", "60s")
	viper.SetDefault("hybrid.circuit_breaker.timeout", "30s")
	viper.SetDefault("hybrid.circuit_breaker.ready_to_trip", 0.5)
	viper.SetDefault("hybrid.circuit_breaker.on_state_change", true)

	viper.SetDefault("hybrid.rate_limit.requests_per_minute", 60)
	viper.SetDefault("hybrid.rate_limit.burst_size", 10)

	viper.SetDefault("hybrid.tracing.enabled", false)
	viper.SetDefault("hybrid.tracing.service_name", "taoist-api")
	viper.SetDefault("hybrid.tracing.jaeger.endpoint", "http://localhost:14268/api/traces")

	viper.SetDefault("hybrid.metrics.enabled", false)
	viper.SetDefault("hybrid.metrics.port", 9090)
	viper.SetDefault("hybrid.metrics.path", "/metrics")

	viper.SetDefault("hybrid.logging.level", "info")
	viper.SetDefault("hybrid.logging.format", "json")

	// Security defaults
	viper.SetDefault("security.jwt_secret", "your-secret-key")
	viper.SetDefault("security.expires_in", 24)
	viper.SetDefault("security.refresh_expires_in", 168) // 7 days

	// General defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("debug", true)
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDebug returns true if debug mode is enabled
func (c *Config) IsDebug() bool {
	return c.Debug
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}

	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", c.Redis.Port)
	}

	if c.AI.VectorService.Port <= 0 || c.AI.VectorService.Port > 65535 {
		return fmt.Errorf("invalid vector service port: %d", c.AI.VectorService.Port)
	}

	if c.AI.ModelService.Port <= 0 || c.AI.ModelService.Port > 65535 {
		return fmt.Errorf("invalid model service port: %d", c.AI.ModelService.Port)
	}

	if c.Security.JWTSecret == "" || c.Security.JWTSecret == "your-secret-key" {
		return fmt.Errorf("invalid jwt secret: must be set to a non-default value")
	}

	return nil
}