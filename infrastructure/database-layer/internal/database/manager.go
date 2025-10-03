package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager 数据库管理器
type Manager struct {
	postgres *PostgresDB
	redis    *RedisDB
	config   *DatabaseManagerConfig
	logger   *zap.Logger

	// 健康检查间隔相关
	healthCheckTicker *time.Ticker
	healthCheckStop   chan bool
	healthMutex       sync.RWMutex
	healthStatus      map[string]bool

	// 监控指标
	metrics *DatabaseMetrics
}

// DatabaseMetrics 数据库监控指标
type DatabaseMetrics struct {
	PostgresConnections int32
	RedisConnections    int32
	PostgresHealthy     bool
	RedisHealthy        bool
	LastHealthCheck     time.Time
	TotalQueries        int64
	FailedQueries       int64
	AverageResponseTime time.Duration
	mutex               sync.RWMutex
}

// DatabaseManagerConfig 数据库管理器配置
type DatabaseManagerConfig struct {
	HealthCheckInterval time.Duration
	MetricsInterval     time.Duration
}

// NewManager 创建新的数据库管理器
func NewManager(config *Config, logger *zap.Logger) (*Manager, error) {
	// 创建管理器配置
	managerConfig := &DatabaseManagerConfig{
		HealthCheckInterval: config.Manager.HealthCheckInterval,
		MetricsInterval:     config.Manager.MetricsInterval,
	}

	// 设置默认值值
	if managerConfig.HealthCheckInterval == 0 {
		managerConfig.HealthCheckInterval = 30 * time.Second
	}
	if managerConfig.MetricsInterval == 0 {
		managerConfig.MetricsInterval = 1 * time.Minute
	}

	manager := &Manager{
		config:          managerConfig,
		logger:          logger,
		healthCheckStop: make(chan bool),
		healthStatus:    make(map[string]bool),
		metrics: &DatabaseMetrics{
			LastHealthCheck: time.Now(),
		},
	}

	// 初始化PostgreSQL
	if config.Postgres.Host != "" {
		postgres, err := NewPostgresDB(&config.Postgres, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
		}
		manager.postgres = postgres
		manager.healthStatus["postgres"] = true
	}

	// 初始化Redis
	if config.Redis.Host != "" {
		redis, err := NewRedisDB(&config.Redis, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Redis: %w", err)
		}
		manager.redis = redis
		manager.healthStatus["redis"] = true
	}

	// 启动统一健康检查间隔
	manager.startUnifiedHealthCheck()

	logger.Info("Database manager initialized successfully",
		zap.Bool("postgres_enabled", config.Postgres.Host != ""),
		zap.Bool("redis_enabled", config.Redis.Host != ""),
		zap.Duration("health_check_interval", managerConfig.HealthCheckInterval),
	)

	return manager, nil
}

// startUnifiedHealthCheck 启动统一健康检查间隔
func (m *Manager) startUnifiedHealthCheck() {
	m.healthCheckTicker = time.NewTicker(m.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-m.healthCheckTicker.C:
				m.performUnifiedHealthCheck()
			case <-m.healthCheckStop:
				m.healthCheckTicker.Stop()
				return
			}
		}
	}()
}

// performUnifiedHealthCheck 执行统一健康检查间隔
func (m *Manager) performUnifiedHealthCheck() {
	m.healthMutex.Lock()
	defer m.healthMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查PostgreSQL健康状态
	if m.postgres != nil {
		if err := m.postgres.Health(); err != nil {
			m.logger.Error("PostgreSQL health check failed", zap.Error(err))
			m.healthStatus["postgres"] = false
			m.metrics.PostgresHealthy = false
		} else {
			if !m.healthStatus["postgres"] {
				m.logger.Info("PostgreSQL connection restored")
			}
			m.healthStatus["postgres"] = true
			m.metrics.PostgresHealthy = true
		}
	}

	// 检查Redis健康状态态
	if m.redis != nil {
		if err := m.redis.Health(); err != nil {
			m.logger.Error("Redis health check failed", zap.Error(err))
			m.healthStatus["redis"] = false
			m.metrics.RedisHealthy = false
		} else {
			if !m.healthStatus["redis"] {
				m.logger.Info("Redis connection restored")
			}
			m.healthStatus["redis"] = true
			m.metrics.RedisHealthy = true
		}
	}

	// 更新监控指标
	m.updateMetrics(ctx)

	m.metrics.LastHealthCheck = time.Now()

	// 记录整体健康状态
	m.logger.Debug("Database health check completed",
		zap.Bool("postgres_healthy", m.healthStatus["postgres"]),
		zap.Bool("redis_healthy", m.healthStatus["redis"]),
		zap.Time("last_check", m.metrics.LastHealthCheck),
	)
}

// updateMetrics 更新监控指标
func (m *Manager) updateMetrics(ctx context.Context) {
	m.metrics.mutex.Lock()
	defer m.metrics.mutex.Unlock()

	// 更新PostgreSQL连接数
	if m.postgres != nil {
		stats := m.postgres.GetStats()
		if openConns, ok := stats["open_connections"].(int); ok {
			m.metrics.PostgresConnections = int32(openConns)
		}
	}

	// 更新Redis连接数
	if m.redis != nil {
		stats := m.redis.GetStats()
		if totalConns, ok := stats["total_conns"].(uint32); ok {
			m.metrics.RedisConnections = int32(totalConns)
		}
	}
}

// GetPostgresDB 获取值PostgreSQL数据库实例
func (m *Manager) GetPostgresDB() *PostgresDB {
	return m.postgres
}

// GetRedisDB 获取值Redis数据库实例
func (m *Manager) GetRedisDB() *RedisDB {
	return m.redis
}

// Close 关闭所有数据库连接
func (m *Manager) Close() error {
	m.logger.Info("Shutting down database manager")

	// 停止健康检查查间隔
	if m.healthCheckTicker != nil {
		close(m.healthCheckStop)
	}

	var errors []error

	// 关闭PostgreSQL连接
	if m.postgres != nil {
		if err := m.postgres.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close PostgreSQL: %w", err))
		}
	}

	// 关闭Redis连接
	if m.redis != nil {
		if err := m.redis.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	m.logger.Info("Database manager shutdown completed")
	return nil
}

// Health 检查所有数据库的健康状态
func (m *Manager) Health() error {
	m.healthMutex.RLock()
	defer m.healthMutex.RUnlock()

	var errors []error

	// 检查PostgreSQL
	if m.postgres != nil && !m.healthStatus["postgres"] {
		errors = append(errors, fmt.Errorf("PostgreSQL is unhealthy"))
	}

	// 检查Redis
	if m.redis != nil && !m.healthStatus["redis"] {
		errors = append(errors, fmt.Errorf("Redis is unhealthy"))
	}

	if len(errors) > 0 {
		return fmt.Errorf("database health check failed: %v", errors)
	}

	return nil
}

// GetHealthStatus 获取值详细健康状态
func (m *Manager) GetHealthStatus() map[string]interface{} {
	m.healthMutex.RLock()
	defer m.healthMutex.RUnlock()

	status := map[string]interface{}{
		"overall_healthy": len(m.healthStatus) > 0,
		"last_check":      m.metrics.LastHealthCheck,
		"databases":       make(map[string]interface{}),
	}

	// PostgreSQL健康状态
	if m.postgres != nil {
		postgresStats := m.postgres.GetStats()
		status["databases"].(map[string]interface{})["postgres"] = map[string]interface{}{
			"healthy": m.healthStatus["postgres"],
			"stats":   postgresStats,
		}
	}

	// Redis健康状态
	if m.redis != nil {
		redisStats := m.redis.GetStats()
		status["databases"].(map[string]interface{})["redis"] = map[string]interface{}{
			"healthy": m.healthStatus["redis"],
			"stats":   redisStats,
		}
	}

	// 检查整体健康状态
	overallHealthy := true
	for _, healthy := range m.healthStatus {
		if !healthy {
			overallHealthy = false
			break
		}
	}
	status["overall_healthy"] = overallHealthy

	return status
}

// GetMetrics 获取值监控指标
func (m *Manager) GetMetrics() *DatabaseMetrics {
	m.metrics.mutex.RLock()
	defer m.metrics.mutex.RUnlock()

	// 返回指标的副本
	return &DatabaseMetrics{
		PostgresConnections: m.metrics.PostgresConnections,
		RedisConnections:    m.metrics.RedisConnections,
		PostgresHealthy:     m.metrics.PostgresHealthy,
		RedisHealthy:        m.metrics.RedisHealthy,
		LastHealthCheck:     m.metrics.LastHealthCheck,
		TotalQueries:        m.metrics.TotalQueries,
		FailedQueries:       m.metrics.FailedQueries,
		AverageResponseTime: m.metrics.AverageResponseTime,
	}
}

// IsHealthy 检查管理器是否健康
func (m *Manager) IsHealthy() bool {
	m.healthMutex.RLock()
	defer m.healthMutex.RUnlock()

	for _, healthy := range m.healthStatus {
		if !healthy {
			return false
		}
	}
	return len(m.healthStatus) > 0
}

// GetDefaultPostgresConfig 获取值默认PostgreSQL配置
func GetDefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		Username:     "postgres",
		Password:     "password",
		Database:     "taishang",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}
}

// GetDefaultRedisConfig 获取值默认Redis配置
func GetDefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// CacheService 缓存服务接口
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
}

// GetCacheService 获取值缓存服务
func (m *Manager) GetCacheService() CacheService {
	if m.redis != nil {
		return m.redis
	}
	// 如果Redis不可用，返回内存缓存实现
	return &memoryCache{
		data: make(map[string]cacheItem),
	}
}

// memoryCache 内存缓存实现（Redis不可用时的备选方案）
type memoryCache struct {
	data map[string]cacheItem
}

type cacheItem struct {
	value      string
	expiration time.Time
}

func (m *memoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.data[key] = cacheItem{
		value:      fmt.Sprintf("%v", value),
		expiration: exp,
	}
	return nil
}

func (m *memoryCache) Get(ctx context.Context, key string) (string, error) {
	item, exists := m.data[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(m.data, key)
		return "", fmt.Errorf("key expired")
	}

	return item.value, nil
}

func (m *memoryCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *memoryCache) Exists(ctx context.Context, keys ...string) (int64, error) {
	var count int64
	for _, key := range keys {
		if item, exists := m.data[key]; exists {
			if item.expiration.IsZero() || time.Now().Before(item.expiration) {
				count++
			}
		}
	}
	return count, nil
}

