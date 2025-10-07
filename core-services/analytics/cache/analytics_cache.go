package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/taishanglaojun/core-services/analytics"
)

// RedisAnalyticsCache Redis数据分析缓存实现
type RedisAnalyticsCache struct {
	client *redis.Client
	config *RedisAnalyticsCacheConfig
}

// RedisAnalyticsCacheConfig Redis数据分析缓存配置
type RedisAnalyticsCacheConfig struct {
	// Redis连接配置
	Address      string `json:"address"`
	Password     string `json:"password"`
	Database     int    `json:"database"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	MaxRetries   int    `json:"max_retries"`
	
	// 缓存配置
	KeyPrefix           string        `json:"key_prefix"`
	DefaultTTL          time.Duration `json:"default_ttl"`
	DataPointTTL        time.Duration `json:"data_point_ttl"`
	AggregatedDataTTL   time.Duration `json:"aggregated_data_ttl"`
	ReportTTL           time.Duration `json:"report_ttl"`
	AnalysisResultTTL   time.Duration `json:"analysis_result_ttl"`
	QueryResultTTL      time.Duration `json:"query_result_ttl"`
	
	// 性能配置
	EnableCompression   bool          `json:"enable_compression"`
	EnableBatching      bool          `json:"enable_batching"`
	BatchSize           int           `json:"batch_size"`
	BatchTimeout        time.Duration `json:"batch_timeout"`
	
	// 序列化配置
	SerializationFormat string        `json:"serialization_format"` // json, msgpack, protobuf
	
	// 监控配置
	EnableMetrics       bool          `json:"enable_metrics"`
	MetricsPrefix       string        `json:"metrics_prefix"`
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Hits              int64   `json:"hits"`
	Misses            int64   `json:"misses"`
	HitRate           float64 `json:"hit_rate"`
	TotalKeys         int64   `json:"total_keys"`
	UsedMemory        int64   `json:"used_memory"`
	MaxMemory         int64   `json:"max_memory"`
	MemoryUsageRate   float64 `json:"memory_usage_rate"`
	ConnectedClients  int64   `json:"connected_clients"`
	CommandsProcessed int64   `json:"commands_processed"`
	KeyspaceHits      int64   `json:"keyspace_hits"`
	KeyspaceMisses    int64   `json:"keyspace_misses"`
}

// NewRedisAnalyticsCache 创建Redis数据分析缓存
func NewRedisAnalyticsCache(config *RedisAnalyticsCacheConfig) (*RedisAnalyticsCache, error) {
	if config == nil {
		config = &RedisAnalyticsCacheConfig{
			Address:             "localhost:6379",
			Database:            0,
			PoolSize:            10,
			MinIdleConns:        2,
			MaxRetries:          3,
			KeyPrefix:           "analytics:",
			DefaultTTL:          1 * time.Hour,
			DataPointTTL:        30 * time.Minute,
			AggregatedDataTTL:   2 * time.Hour,
			ReportTTL:           24 * time.Hour,
			AnalysisResultTTL:   4 * time.Hour,
			QueryResultTTL:      15 * time.Minute,
			EnableCompression:   true,
			EnableBatching:      true,
			BatchSize:           100,
			BatchTimeout:        100 * time.Millisecond,
			SerializationFormat: "json",
			EnableMetrics:       true,
			MetricsPrefix:       "analytics_cache",
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisAnalyticsCache{
		client: client,
		config: config,
	}, nil
}

// SetDataPoint 设置数据点缓存
func (c *RedisAnalyticsCache) SetDataPoint(ctx context.Context, dataPoint *analytics.DataPoint) error {
	if dataPoint == nil {
		return fmt.Errorf("data point cannot be nil")
	}

	key := c.buildKey("datapoint", dataPoint.ID)
	data, err := c.serialize(dataPoint)
	if err != nil {
		return fmt.Errorf("failed to serialize data point: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.DataPointTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set data point cache: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("datapoint_set")
	}

	return nil
}

// GetDataPoint 获取数据点缓存
func (c *RedisAnalyticsCache) GetDataPoint(ctx context.Context, dataPointID string) (*analytics.DataPoint, error) {
	key := c.buildKey("datapoint", dataPointID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			if c.config.EnableMetrics {
				c.incrementMetric("datapoint_miss")
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get data point cache: %w", err)
	}

	var dataPoint analytics.DataPoint
	if err := c.deserialize([]byte(data), &dataPoint); err != nil {
		return nil, fmt.Errorf("failed to deserialize data point: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("datapoint_hit")
	}

	return &dataPoint, nil
}

// DeleteDataPoint 删除数据点缓存
func (c *RedisAnalyticsCache) DeleteDataPoint(ctx context.Context, dataPointID string) error {
	key := c.buildKey("datapoint", dataPointID)
	return c.client.Del(ctx, key).Err()
}

// SetAggregatedData 设置聚合数据缓存
func (c *RedisAnalyticsCache) SetAggregatedData(ctx context.Context, aggregatedData *analytics.AggregatedData) error {
	if aggregatedData == nil {
		return fmt.Errorf("aggregated data cannot be nil")
	}

	key := c.buildKey("aggregated", aggregatedData.ID)
	data, err := c.serialize(aggregatedData)
	if err != nil {
		return fmt.Errorf("failed to serialize aggregated data: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.AggregatedDataTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set aggregated data cache: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("aggregated_set")
	}

	return nil
}

// GetAggregatedData 获取聚合数据缓存
func (c *RedisAnalyticsCache) GetAggregatedData(ctx context.Context, aggregatedDataID string) (*analytics.AggregatedData, error) {
	key := c.buildKey("aggregated", aggregatedDataID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			if c.config.EnableMetrics {
				c.incrementMetric("aggregated_miss")
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get aggregated data cache: %w", err)
	}

	var aggregatedData analytics.AggregatedData
	if err := c.deserialize([]byte(data), &aggregatedData); err != nil {
		return nil, fmt.Errorf("failed to deserialize aggregated data: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("aggregated_hit")
	}

	return &aggregatedData, nil
}

// DeleteAggregatedData 删除聚合数据缓存
func (c *RedisAnalyticsCache) DeleteAggregatedData(ctx context.Context, aggregatedDataID string) error {
	key := c.buildKey("aggregated", aggregatedDataID)
	return c.client.Del(ctx, key).Err()
}

// SetReport 设置报表缓存
func (c *RedisAnalyticsCache) SetReport(ctx context.Context, report *analytics.Report) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	key := c.buildKey("report", report.ID)
	data, err := c.serialize(report)
	if err != nil {
		return fmt.Errorf("failed to serialize report: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.ReportTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set report cache: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("report_set")
	}

	return nil
}

// GetReport 获取报表缓存
func (c *RedisAnalyticsCache) GetReport(ctx context.Context, reportID string) (*analytics.Report, error) {
	key := c.buildKey("report", reportID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			if c.config.EnableMetrics {
				c.incrementMetric("report_miss")
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get report cache: %w", err)
	}

	var report analytics.Report
	if err := c.deserialize([]byte(data), &report); err != nil {
		return nil, fmt.Errorf("failed to deserialize report: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("report_hit")
	}

	return &report, nil
}

// DeleteReport 删除报表缓存
func (c *RedisAnalyticsCache) DeleteReport(ctx context.Context, reportID string) error {
	key := c.buildKey("report", reportID)
	return c.client.Del(ctx, key).Err()
}

// SetAnalysisResult 设置分析结果缓存
func (c *RedisAnalyticsCache) SetAnalysisResult(ctx context.Context, analysisResult *analytics.AnalysisResult) error {
	if analysisResult == nil {
		return fmt.Errorf("analysis result cannot be nil")
	}

	key := c.buildKey("analysis", analysisResult.ID)
	data, err := c.serialize(analysisResult)
	if err != nil {
		return fmt.Errorf("failed to serialize analysis result: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.AnalysisResultTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set analysis result cache: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("analysis_set")
	}

	return nil
}

// GetAnalysisResult 获取分析结果缓存
func (c *RedisAnalyticsCache) GetAnalysisResult(ctx context.Context, analysisResultID string) (*analytics.AnalysisResult, error) {
	key := c.buildKey("analysis", analysisResultID)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			if c.config.EnableMetrics {
				c.incrementMetric("analysis_miss")
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get analysis result cache: %w", err)
	}

	var analysisResult analytics.AnalysisResult
	if err := c.deserialize([]byte(data), &analysisResult); err != nil {
		return nil, fmt.Errorf("failed to deserialize analysis result: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("analysis_hit")
	}

	return &analysisResult, nil
}

// DeleteAnalysisResult 删除分析结果缓存
func (c *RedisAnalyticsCache) DeleteAnalysisResult(ctx context.Context, analysisResultID string) error {
	key := c.buildKey("analysis", analysisResultID)
	return c.client.Del(ctx, key).Err()
}

// SetQueryResult 设置查询结果缓存
func (c *RedisAnalyticsCache) SetQueryResult(ctx context.Context, queryKey string, result interface{}) error {
	if result == nil {
		return fmt.Errorf("query result cannot be nil")
	}

	key := c.buildKey("query", queryKey)
	data, err := c.serialize(result)
	if err != nil {
		return fmt.Errorf("failed to serialize query result: %w", err)
	}

	err = c.client.Set(ctx, key, data, c.config.QueryResultTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set query result cache: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("query_set")
	}

	return nil
}

// GetQueryResult 获取查询结果缓存
func (c *RedisAnalyticsCache) GetQueryResult(ctx context.Context, queryKey string, result interface{}) error {
	key := c.buildKey("query", queryKey)
	
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			if c.config.EnableMetrics {
				c.incrementMetric("query_miss")
			}
			return fmt.Errorf("cache miss")
		}
		return fmt.Errorf("failed to get query result cache: %w", err)
	}

	if err := c.deserialize([]byte(data), result); err != nil {
		return fmt.Errorf("failed to deserialize query result: %w", err)
	}

	// 更新指标
	if c.config.EnableMetrics {
		c.incrementMetric("query_hit")
	}

	return nil
}

// DeleteQueryResult 删除查询结果缓存
func (c *RedisAnalyticsCache) DeleteQueryResult(ctx context.Context, queryKey string) error {
	key := c.buildKey("query", queryKey)
	return c.client.Del(ctx, key).Err()
}

// InvalidateByPattern 根据模式删除缓存
func (c *RedisAnalyticsCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern, "*")
	
	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 分批删除
	batchSize := c.config.BatchSize
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := c.client.Del(ctx, batch...).Err(); err != nil {
			return fmt.Errorf("failed to delete batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

// InvalidateUserData 删除用户相关缓存
func (c *RedisAnalyticsCache) InvalidateUserData(ctx context.Context, userID string) error {
	patterns := []string{
		fmt.Sprintf("datapoint:*:user:%s", userID),
		fmt.Sprintf("report:*:user:%s", userID),
		fmt.Sprintf("analysis:*:user:%s", userID),
	}

	for _, pattern := range patterns {
		if err := c.InvalidateByPattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// InvalidateTenantData 删除租户相关缓存
func (c *RedisAnalyticsCache) InvalidateTenantData(ctx context.Context, tenantID string) error {
	patterns := []string{
		fmt.Sprintf("datapoint:*:tenant:%s", tenantID),
		fmt.Sprintf("aggregated:*:tenant:%s", tenantID),
		fmt.Sprintf("report:*:tenant:%s", tenantID),
		fmt.Sprintf("analysis:*:tenant:%s", tenantID),
	}

	for _, pattern := range patterns {
		if err := c.InvalidateByPattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// ClearAll 清空所有缓存
func (c *RedisAnalyticsCache) ClearAll(ctx context.Context) error {
	pattern := c.buildKey("*")
	
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get all keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 分批删除
	batchSize := c.config.BatchSize
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := c.client.Del(ctx, batch...).Err(); err != nil {
			return fmt.Errorf("failed to delete batch %d-%d: %w", i, end-1, err)
		}
	}

	return nil
}

// GetStats 获取缓存统计信息
func (c *RedisAnalyticsCache) GetStats(ctx context.Context) (*CacheStats, error) {
	info, err := c.client.Info(ctx, "stats", "memory", "clients").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	stats := &CacheStats{}
	
	// 解析Redis INFO输出
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			switch key {
			case "keyspace_hits":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.KeyspaceHits = v
				}
			case "keyspace_misses":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.KeyspaceMisses = v
				}
			case "used_memory":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.UsedMemory = v
				}
			case "maxmemory":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.MaxMemory = v
				}
			case "connected_clients":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.ConnectedClients = v
				}
			case "total_commands_processed":
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.CommandsProcessed = v
				}
			}
		}
	}

	// 计算命中率
	totalRequests := stats.KeyspaceHits + stats.KeyspaceMisses
	if totalRequests > 0 {
		stats.HitRate = float64(stats.KeyspaceHits) / float64(totalRequests)
	}

	// 计算内存使用率
	if stats.MaxMemory > 0 {
		stats.MemoryUsageRate = float64(stats.UsedMemory) / float64(stats.MaxMemory)
	}

	// 获取键总数
	dbSize, err := c.client.DBSize(ctx).Result()
	if err == nil {
		stats.TotalKeys = dbSize
	}

	return stats, nil
}

// HealthCheck 健康检查
func (c *RedisAnalyticsCache) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 测试连接
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis ping failed: %w", err)
	}

	// 测试读写
	testKey := c.buildKey("healthcheck", "test")
	testValue := "test_value"
	
	if err := c.client.Set(ctx, testKey, testValue, time.Minute).Err(); err != nil {
		return fmt.Errorf("Redis set failed: %w", err)
	}

	result, err := c.client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("Redis get failed: %w", err)
	}

	if result != testValue {
		return fmt.Errorf("Redis value mismatch: expected %s, got %s", testValue, result)
	}

	// 清理测试键
	c.client.Del(ctx, testKey)

	return nil
}

// Close 关闭缓存连接
func (c *RedisAnalyticsCache) Close() error {
	return c.client.Close()
}

// 私有方法

func (c *RedisAnalyticsCache) buildKey(parts ...string) string {
	allParts := append([]string{c.config.KeyPrefix}, parts...)
	return strings.Join(allParts, ":")
}

func (c *RedisAnalyticsCache) serialize(data interface{}) ([]byte, error) {
	switch c.config.SerializationFormat {
	case "json":
		return json.Marshal(data)
	default:
		return json.Marshal(data)
	}
}

func (c *RedisAnalyticsCache) deserialize(data []byte, result interface{}) error {
	switch c.config.SerializationFormat {
	case "json":
		return json.Unmarshal(data, result)
	default:
		return json.Unmarshal(data, result)
	}
}

func (c *RedisAnalyticsCache) incrementMetric(metric string) {
	// 这里可以集成到监控系统，如Prometheus
	// 暂时使用Redis计数器
	key := c.buildKey("metrics", metric)
	c.client.Incr(context.Background(), key)
	c.client.Expire(context.Background(), key, 24*time.Hour)
}