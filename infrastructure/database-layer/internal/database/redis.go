package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host                 string
	Port                 int
	Password             string
	Database             int
	PoolSize             int
	MinIdleConns         int
	MaxRetries           int
	DialTimeout          time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	PoolTimeout          time.Duration        // 连接池超时
	IdleTimeout          time.Duration        // 空闲连接超时
	IdleCheckFrequency   time.Duration        // 空闲连接检查频率
	MaxConnAge           time.Duration        // 连接最大生存时间
	HealthCheckInterval  time.Duration        // 健康检查间隔
	ReconnectInterval    time.Duration        // 重连间隔
	MaxReconnectAttempts int                  // 最大重连尝试次数
}

// RedisDB Redis数据库管理器
type RedisDB struct {
	client               *redis.Client
	config               *RedisConfig
	logger               *zap.Logger
	healthCheckTicker    *time.Ticker
	healthCheckStop      chan bool
	reconnectMutex       sync.RWMutex
	isHealthy            bool
	lastHealthCheck      time.Time
	connectionLeakDetector *RedisConnectionLeakDetector
}

// RedisConnectionLeakDetector Redis连接泄漏检测器
type RedisConnectionLeakDetector struct {
	maxConnections     int
	warningThreshold   float64
	checkInterval      time.Duration
	logger            *zap.Logger
	ticker            *time.Ticker
	stop              chan bool
}

// NewRedisDB 创建新的Redis数据库连接
func NewRedisDB(config *RedisConfig, log *zap.Logger) (*RedisDB, error) {
	// 设置默认值
	if config.PoolTimeout == 0 {
		config.PoolTimeout = 4 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.IdleCheckFrequency == 0 {
		config.IdleCheckFrequency = 1 * time.Minute
	}
	if config.MaxConnAge == 0 {
		config.MaxConnAge = 30 * time.Minute
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}
	if config.MaxReconnectAttempts == 0 {
		config.MaxReconnectAttempts = 3
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:               fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:           config.Password,
		DB:                 config.Database,
		PoolSize:           config.PoolSize,
		MinIdleConns:       config.MinIdleConns,
		MaxRetries:         config.MaxRetries,
		DialTimeout:        config.DialTimeout,
		ReadTimeout:        config.ReadTimeout,
		WriteTimeout:       config.WriteTimeout,
		PoolTimeout:        config.PoolTimeout,
		IdleTimeout:        config.IdleTimeout,
		IdleCheckFrequency: config.IdleCheckFrequency,
		MaxConnAge:         config.MaxConnAge,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	redisDB := &RedisDB{
		client:          client,
		config:          config,
		logger:          log,
		healthCheckStop: make(chan bool),
		isHealthy:       true,
		lastHealthCheck: time.Now(),
	}

	// 初始化连接泄漏检测器
	redisDB.connectionLeakDetector = &RedisConnectionLeakDetector{
		maxConnections:   config.PoolSize,
		warningThreshold: 0.8, // 80%阈值
		checkInterval:    1 * time.Minute,
		logger:          log,
		stop:            make(chan bool),
	}

	// 启动健康检查
	redisDB.startHealthCheck()
	
	// 启动连接泄漏检测
	redisDB.startConnectionLeakDetection()

	log.Info("Redis connected successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.Int("database", config.Database),
		zap.Int("pool_size", config.PoolSize),
		zap.Int("min_idle_conns", config.MinIdleConns),
		zap.Duration("pool_timeout", config.PoolTimeout),
		zap.Duration("idle_timeout", config.IdleTimeout),
	)

	return redisDB, nil
}

// startHealthCheck 启动健康检查
func (r *RedisDB) startHealthCheck() {
	r.healthCheckTicker = time.NewTicker(r.config.HealthCheckInterval)
	
	go func() {
		for {
			select {
			case <-r.healthCheckTicker.C:
				r.performHealthCheck()
			case <-r.healthCheckStop:
				r.healthCheckTicker.Stop()
				return
			}
		}
	}()
}

// performHealthCheck 执行健康检查
func (r *RedisDB) performHealthCheck() {
	r.reconnectMutex.Lock()
	defer r.reconnectMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		r.logger.Error("Redis health check failed", zap.Error(err))
		r.isHealthy = false
		r.attemptReconnect()
		return
	}

	if !r.isHealthy {
		r.logger.Info("Redis connection restored")
	}
	
	r.isHealthy = true
	r.lastHealthCheck = time.Now()
	
	// 记录连接池统计信息
	stats := r.client.PoolStats()
	r.logger.Debug("Redis connection pool stats",
		zap.Uint32("hits", stats.Hits),
		zap.Uint32("misses", stats.Misses),
		zap.Uint32("timeouts", stats.Timeouts),
		zap.Uint32("total_conns", stats.TotalConns),
		zap.Uint32("idle_conns", stats.IdleConns),
		zap.Uint32("stale_conns", stats.StaleConns),
	)
}

// attemptReconnect 尝试重连
func (r *RedisDB) attemptReconnect() {
	for attempt := 1; attempt <= r.config.MaxReconnectAttempts; attempt++ {
		r.logger.Info("Attempting to reconnect to Redis",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", r.config.MaxReconnectAttempts),
		)

		time.Sleep(r.config.ReconnectInterval)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := r.client.Ping(ctx).Err(); err != nil {
			cancel()
			r.logger.Error("Redis reconnection attempt failed", 
				zap.Int("attempt", attempt), 
				zap.Error(err),
			)
			continue
		}
		cancel()

		r.logger.Info("Successfully reconnected to Redis", zap.Int("attempt", attempt))
		r.isHealthy = true
		return
	}

	r.logger.Error("Failed to reconnect to Redis after maximum attempts",
		zap.Int("max_attempts", r.config.MaxReconnectAttempts),
	)
}

// startConnectionLeakDetection 启动连接泄漏检测
func (r *RedisDB) startConnectionLeakDetection() {
	detector := r.connectionLeakDetector
	detector.ticker = time.NewTicker(detector.checkInterval)
	
	go func() {
		for {
			select {
			case <-detector.ticker.C:
				r.checkConnectionLeak()
			case <-detector.stop:
				detector.ticker.Stop()
				return
			}
		}
	}()
}

// checkConnectionLeak 检查连接泄漏
func (r *RedisDB) checkConnectionLeak() {
	stats := r.client.PoolStats()
	detector := r.connectionLeakDetector
	
	usageRatio := float64(stats.TotalConns) / float64(detector.maxConnections)
	
	if usageRatio >= detector.warningThreshold {
		detector.logger.Warn("High Redis connection usage detected",
			zap.Uint32("total_conns", stats.TotalConns),
			zap.Int("max_connections", detector.maxConnections),
			zap.Float64("usage_ratio", usageRatio),
			zap.Uint32("idle_conns", stats.IdleConns),
			zap.Uint32("stale_conns", stats.StaleConns),
			zap.Uint32("timeouts", stats.Timeouts),
		)
	}
}

// GetClient 获取Redis客户端
func (r *RedisDB) GetClient() *redis.Client {
	return r.client
}

// Close 关闭Redis连接
func (r *RedisDB) Close() error {
	// 停止健康检查
	if r.healthCheckTicker != nil {
		close(r.healthCheckStop)
	}
	
	// 停止连接泄漏检测
	if r.connectionLeakDetector != nil && r.connectionLeakDetector.ticker != nil {
		close(r.connectionLeakDetector.stop)
	}

	r.logger.Info("Closing Redis database connections")
	return r.client.Close()
}

// Health 检查Redis健康状态
func (r *RedisDB) Health() error {
	r.reconnectMutex.RLock()
	defer r.reconnectMutex.RUnlock()
	
	if !r.isHealthy {
		return fmt.Errorf("Redis is unhealthy, last check: %v", r.lastHealthCheck)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	return r.client.Ping(ctx).Err()
}

// IsHealthy 返回Redis健康状态
func (r *RedisDB) IsHealthy() bool {
	r.reconnectMutex.RLock()
	defer r.reconnectMutex.RUnlock()
	return r.isHealthy
}

// GetStats 获取Redis统计信息
func (r *RedisDB) GetStats() map[string]interface{} {
	stats := r.client.PoolStats()
	return map[string]interface{}{
		"hits":                    stats.Hits,
		"misses":                  stats.Misses,
		"timeouts":                stats.Timeouts,
		"total_conns":             stats.TotalConns,
		"idle_conns":              stats.IdleConns,
		"stale_conns":             stats.StaleConns,
		"is_healthy":              r.isHealthy,
		"last_health_check":       r.lastHealthCheck,
		"connection_usage_ratio":  float64(stats.TotalConns) / float64(r.config.PoolSize),
		"hit_ratio":              float64(stats.Hits) / float64(stats.Hits + stats.Misses),
	}
}

// Set 设置键值对
func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del 删除键
func (r *RedisDB) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisDB) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置键过期时间
func (r *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func (r *RedisDB) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// HSet 设置哈希字段
func (r *RedisDB) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段值
func (r *RedisDB) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希所有字段
func (r *RedisDB) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisDB) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// LPush 从列表左侧推入元素
func (r *RedisDB) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 从列表右侧推入元素
func (r *RedisDB) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop 从列表左侧弹出元素
func (r *RedisDB) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 从列表右侧弹出元素
func (r *RedisDB) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LLen 获取列表长度
func (r *RedisDB) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

// SAdd 向集合添加成员
func (r *RedisDB) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *RedisDB) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SIsMember 判断元素是否是集合成员
func (r *RedisDB) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}

// SRem 从集合移除成员
func (r *RedisDB) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// ZAdd 向有序集合添加成员
func (r *RedisDB) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合指定范围的成员
func (r *RedisDB) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 从有序集合移除成员
func (r *RedisDB) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Incr 递增键的值
func (r *RedisDB) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decr 递减键的值
func (r *RedisDB) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// IncrBy 按指定值递增键的值
func (r *RedisDB) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// DecrBy 按指定值递减键的值
func (r *RedisDB) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.DecrBy(ctx, key, value).Result()
}