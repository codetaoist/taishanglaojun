package database

import (
	"time"
)

// Config 数据库配置结构体
type Config struct {
	Postgres PostgresConfig `yaml:"postgres" mapstructure:"postgres"`
	Redis    RedisConfig    `yaml:"redis" mapstructure:"redis"`
	Manager  ManagerConfig  `yaml:"manager" mapstructure:"manager"`
	Monitor  MonitorConfig  `yaml:"monitor" mapstructure:"monitor"`
}

// ManagerConfig 数据库管理器配置
type ManagerConfig struct {
	HealthCheckInterval time.Duration `yaml:"health_check_interval" mapstructure:"health_check_interval"`
	MetricsInterval     time.Duration `yaml:"metrics_interval" mapstructure:"metrics_interval"`
}

// MonitorConfig 连接监控配置
type MonitorConfig struct {
	LeakDetectionInterval     time.Duration `yaml:"leak_detection_interval" mapstructure:"leak_detection_interval"`
	LeakWarningThreshold      float64       `yaml:"leak_warning_threshold" mapstructure:"leak_warning_threshold"`
	LeakCriticalThreshold     float64       `yaml:"leak_critical_threshold" mapstructure:"leak_critical_threshold"`
	ShutdownTimeout           time.Duration `yaml:"shutdown_timeout" mapstructure:"shutdown_timeout"`
	MetricsCollectionInterval time.Duration `yaml:"metrics_collection_interval" mapstructure:"metrics_collection_interval"`
	EnableDetailedLogging     bool          `yaml:"enable_detailed_logging" mapstructure:"enable_detailed_logging"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Postgres: PostgresConfig{
			Host:                 "localhost",
			Port:                 5432,
			Username:             "postgres",
			Password:             "password",
			Database:             "taishang",
			SSLMode:              "disable",
			MaxOpenConns:         25,
			MaxIdleConns:         5,
			MaxLifetime:          5 * time.Minute,
			MaxIdleTime:          10 * time.Minute,
			ConnMaxIdleTime:      30 * time.Minute,
			HealthCheckInterval:  30 * time.Second,
			ReconnectInterval:    5 * time.Second,
			MaxReconnectAttempts: 3,
		},
		Redis: RedisConfig{
			Host:                 "localhost",
			Port:                 6379,
			Password:             "",
			Database:             0,
			PoolSize:             10,
			MinIdleConns:         2,
			MaxRetries:           3,
			DialTimeout:          5 * time.Second,
			ReadTimeout:          3 * time.Second,
			WriteTimeout:         3 * time.Second,
			PoolTimeout:          4 * time.Second,
			IdleTimeout:          5 * time.Minute,
			IdleCheckFrequency:   1 * time.Minute,
			MaxConnAge:           30 * time.Minute,
			HealthCheckInterval:  30 * time.Second,
			ReconnectInterval:    5 * time.Second,
			MaxReconnectAttempts: 3,
		},
		Manager: ManagerConfig{
			HealthCheckInterval: 30 * time.Second,
			MetricsInterval:     1 * time.Minute,
		},
		Monitor: MonitorConfig{
			LeakDetectionInterval:     1 * time.Minute,
			LeakWarningThreshold:      0.8,
			LeakCriticalThreshold:     0.95,
			ShutdownTimeout:           30 * time.Second,
			MetricsCollectionInterval: 30 * time.Second,
			EnableDetailedLogging:     false,
		},
	}
}

