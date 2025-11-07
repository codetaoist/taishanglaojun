package config

import (
	"os"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Port            string `mapstructure:"port"`
	Environment     string `mapstructure:"environment"`
	LogLevel        string `mapstructure:"log_level"`
	DatabaseURL     string `mapstructure:"database_url"`
	JWTSecret       string `mapstructure:"jwt_secret"`
	JWTExpiration   int    `mapstructure:"jwt_expiration"`
	TraceHeader     string `mapstructure:"trace_header"`
	AllowedOrigins  []string `mapstructure:"allowed_origins"`
}

// Load loads configuration from environment variables and config file
func Load() (Config, error) {
	var cfg Config

	// Set defaults
	viper.SetDefault("port", "8081")
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("jwt_expiration", 86400) // 24 hours
	viper.SetDefault("trace_header", "X-Trace-ID")
	viper.SetDefault("allowed_origins", []string{"*"})

	// Configure viper to read from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AUTH")

	// Try to read config file if it exists
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, that's OK, we'll use env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, err
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		cfg.Port = port
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.JWTSecret = jwtSecret
	}

	return cfg, nil
}