package config

import (
	"os"
	"strconv"
	"time"
)

// LocationConfig 位置跟踪配置
type LocationConfig struct {
	// 数据库配�?
	Database DatabaseConfig `json:"database"`
	
	// 缓存配置
	Cache CacheConfig `json:"cache"`
	
	// 性能配置
	Performance PerformanceConfig `json:"performance"`
	
	// 安全配置
	Security SecurityConfig `json:"security"`
}

// DatabaseConfig 数据库配�?
type DatabaseConfig struct {
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enabled    bool          `json:"enabled"`
	TTL        time.Duration `json:"ttl"`
	MaxEntries int           `json:"max_entries"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	BatchSize           int           `json:"batch_size"`
	MaxPointsPerRequest int           `json:"max_points_per_request"`
	StatsUpdateInterval time.Duration `json:"stats_update_interval"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RateLimitEnabled bool `json:"rate_limit_enabled"`
	MaxRequestsPerMinute int `json:"max_requests_per_minute"`
	MaxTrajectoryPerUser int `json:"max_trajectory_per_user"`
	MaxPointsPerTrajectory int `json:"max_points_per_trajectory"`
}

// LoadLocationConfig 加载位置跟踪配置
func LoadLocationConfig() *LocationConfig {
	return &LocationConfig{
		Database: DatabaseConfig{
			MaxOpenConns:    getEnvInt("LOCATION_DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("LOCATION_DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("LOCATION_DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("LOCATION_DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		Cache: CacheConfig{
			Enabled:    getEnvBool("LOCATION_CACHE_ENABLED", true),
			TTL:        getEnvDuration("LOCATION_CACHE_TTL", 10*time.Minute),
			MaxEntries: getEnvInt("LOCATION_CACHE_MAX_ENTRIES", 1000),
		},
		Performance: PerformanceConfig{
			BatchSize:           getEnvInt("LOCATION_BATCH_SIZE", 100),
			MaxPointsPerRequest: getEnvInt("LOCATION_MAX_POINTS_PER_REQUEST", 1000),
			StatsUpdateInterval: getEnvDuration("LOCATION_STATS_UPDATE_INTERVAL", 30*time.Second),
		},
		Security: SecurityConfig{
			RateLimitEnabled:       getEnvBool("LOCATION_RATE_LIMIT_ENABLED", true),
			MaxRequestsPerMinute:   getEnvInt("LOCATION_MAX_REQUESTS_PER_MINUTE", 100),
			MaxTrajectoryPerUser:   getEnvInt("LOCATION_MAX_TRAJECTORY_PER_USER", 100),
			MaxPointsPerTrajectory: getEnvInt("LOCATION_MAX_POINTS_PER_TRAJECTORY", 10000),
		},
	}
}

// 辅助函数
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
