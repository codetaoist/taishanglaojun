package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 
type Config struct {
	App     AppConfig     `yaml:"app"`
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Log     LogConfig     `yaml:"log"`
	Auth    AuthConfig    `yaml:"auth"`
	AI      AIConfig      `yaml:"ai"`
}

// AppConfig 
type AppConfig struct {
	Name        string `yaml:"name" env:"APP_NAME" default:"intelligent-learning"`
	Version     string `yaml:"version" env:"APP_VERSION" default:"1.0.0"`
	Environment string `yaml:"environment" env:"APP_ENV" default:"development"`
	Debug       bool   `yaml:"debug" env:"APP_DEBUG" default:"true"`
}

// ServerConfig ?
type ServerConfig struct {
	Host            string        `yaml:"host" env:"SERVER_HOST" default:"0.0.0.0"`
	Port            int           `yaml:"port" env:"SERVER_PORT" default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" default:"30s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT" default:"30s"`
	CORS            CORSConfig    `yaml:"cors"`
}

// CORSConfig CORS
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" env:"CORS_ALLOW_ORIGINS" default:"*"`
	AllowMethods     []string `yaml:"allow_methods" env:"CORS_ALLOW_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowHeaders     []string `yaml:"allow_headers" env:"CORS_ALLOW_HEADERS" default:"*"`
	ExposeHeaders    []string `yaml:"expose_headers" env:"CORS_EXPOSE_HEADERS" default:""`
	AllowCredentials bool     `yaml:"allow_credentials" env:"CORS_ALLOW_CREDENTIALS" default:"true"`
	MaxAge           int      `yaml:"max_age" env:"CORS_MAX_AGE" default:"86400"`
}

// LogConfig 
type LogConfig struct {
	Level      string `yaml:"level" env:"LOG_LEVEL" default:"info"`
	Format     string `yaml:"format" env:"LOG_FORMAT" default:"json"`
	Output     string `yaml:"output" env:"LOG_OUTPUT" default:"stdout"`
	Filename   string `yaml:"filename" env:"LOG_FILENAME" default:"app.log"`
	MaxSize    int    `yaml:"max_size" env:"LOG_MAX_SIZE" default:"100"`
	MaxBackups int    `yaml:"max_backups" env:"LOG_MAX_BACKUPS" default:"3"`
	MaxAge     int    `yaml:"max_age" env:"LOG_MAX_AGE" default:"28"`
	Compress   bool   `yaml:"compress" env:"LOG_COMPRESS" default:"true"`
}

// AuthConfig 
type AuthConfig struct {
	JWTSecret     string        `yaml:"jwt_secret" env:"JWT_SECRET" default:"your-secret-key"`
	JWTExpiration time.Duration `yaml:"jwt_expiration" env:"JWT_EXPIRATION" default:"24h"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration" env:"REFRESH_EXPIRATION" default:"168h"`
	Issuer        string        `yaml:"issuer" env:"JWT_ISSUER" default:"intelligent-learning"`
}

// AIConfig AI
type AIConfig struct {
	Provider string            `yaml:"provider" env:"AI_PROVIDER" default:"openai"`
	APIKey   string            `yaml:"api_key" env:"AI_API_KEY" default:""`
	BaseURL  string            `yaml:"base_url" env:"AI_BASE_URL" default:""`
	Model    string            `yaml:"model" env:"AI_MODEL" default:"gpt-3.5-turbo"`
	Options  map[string]string `yaml:"options"`
}

// Address 
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment ?
func (c *AppConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction ?
func (c *AppConfig) IsProduction() bool {
	return c.Environment == "production"
}

// Load 
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// ?
	setDefaults(config)

	// 
	if configPath != "" && fileExists(configPath) {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// ?
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load config from env: %w", err)
	}

	// 
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// setDefaults ?
func setDefaults(config *Config) {
	config.App = AppConfig{
		Name:        "intelligent-learning",
		Version:     "1.0.0",
		Environment: "development",
		Debug:       true,
	}

	config.Server = ServerConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		CORS: CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: true,
			MaxAge:           86400,
		},
	}

	config.Storage = StorageConfig{
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Username:        "postgres",
			Database:        "intelligent_learning",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
		},
		Redis: RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			IdleTimeout:  5 * time.Minute,
		},
		Elasticsearch: ElasticsearchConfig{
			URLs:  []string{"http://localhost:9200"},
			Index: "intelligent_learning",
		},
		Neo4j: Neo4jConfig{
			URI:      "bolt://localhost:7687",
			Username: "neo4j",
			Password: "password",
			Database: "neo4j",
		},
	}

	config.Log = LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		Filename:   "app.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	config.Auth = AuthConfig{
		JWTSecret:         "your-secret-key",
		JWTExpiration:     24 * time.Hour,
		RefreshExpiration: 168 * time.Hour,
		Issuer:            "intelligent-learning",
	}

	config.AI = AIConfig{
		Provider: "openai",
		Model:    "gpt-3.5-turbo",
		Options:  make(map[string]string),
	}
}

// loadFromFile ?
func loadFromFile(config *Config, configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv ?
func loadFromEnv(config *Config) error {
	// App
	if val := os.Getenv("APP_NAME"); val != "" {
		config.App.Name = val
	}
	if val := os.Getenv("APP_VERSION"); val != "" {
		config.App.Version = val
	}
	if val := os.Getenv("APP_ENV"); val != "" {
		config.App.Environment = val
	}
	if val := os.Getenv("APP_DEBUG"); val != "" {
		config.App.Debug = val == "true"
	}

	// Server
	if val := os.Getenv("SERVER_HOST"); val != "" {
		config.Server.Host = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Server.Port = port
		}
	}

	// Database
	if val := os.Getenv("DB_HOST"); val != "" {
		config.Storage.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Storage.Database.Port = port
		}
	}
	if val := os.Getenv("DB_USERNAME"); val != "" {
		config.Storage.Database.Username = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		config.Storage.Database.Password = val
	}
	if val := os.Getenv("DB_DATABASE"); val != "" {
		config.Storage.Database.Database = val
	}

	// Redis
	if val := os.Getenv("REDIS_HOST"); val != "" {
		config.Storage.Redis.Host = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Storage.Redis.Port = port
		}
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		config.Storage.Redis.Password = val
	}

	// Elasticsearch
	if val := os.Getenv("ES_URLS"); val != "" {
		config.Storage.Elasticsearch.URLs = strings.Split(val, ",")
	}

	// Neo4j
	if val := os.Getenv("NEO4J_URI"); val != "" {
		config.Storage.Neo4j.URI = val
	}
	if val := os.Getenv("NEO4J_USERNAME"); val != "" {
		config.Storage.Neo4j.Username = val
	}
	if val := os.Getenv("NEO4J_PASSWORD"); val != "" {
		config.Storage.Neo4j.Password = val
	}

	// Auth
	if val := os.Getenv("JWT_SECRET"); val != "" {
		config.Auth.JWTSecret = val
	}

	// AI
	if val := os.Getenv("AI_PROVIDER"); val != "" {
		config.AI.Provider = val
	}
	if val := os.Getenv("AI_API_KEY"); val != "" {
		config.AI.APIKey = val
	}
	if val := os.Getenv("AI_MODEL"); val != "" {
		config.AI.Model = val
	}

	return nil
}

// validate 
func validate(config *Config) error {
	if config.App.Name == "" {
		return fmt.Errorf("app name is required")
	}

	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Storage.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Storage.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	if config.Storage.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if config.Auth.JWTSecret == "" || config.Auth.JWTSecret == "your-secret-key" {
		return fmt.Errorf("JWT secret must be set and not use default value")
	}

	return nil
}

// fileExists ?
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

