package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/taishanglaojun/core-services/monitoring/models"
)

// ApplicationCollector 应用指标收集器
type ApplicationCollector struct {
	name     string
	interval time.Duration
	enabled  bool
	labels   map[string]string
	
	// 配置选项
	collectHTTP     bool
	collectGRPC     bool
	collectDatabase bool
	collectCache    bool
	collectRuntime  bool
	
	// 外部依赖
	db          *sql.DB
	redisClient *redis.Client
	
	// Prometheus指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	grpcRequestsTotal    *prometheus.CounterVec
	grpcRequestDuration  *prometheus.HistogramVec
	dbConnectionsActive  prometheus.Gauge
	dbConnectionsIdle    prometheus.Gauge
	dbQueriesTotal       *prometheus.CounterVec
	dbQueryDuration      *prometheus.HistogramVec
	cacheOperationsTotal *prometheus.CounterVec
	cacheHitRate         prometheus.Gauge
	
	// 内部统计
	httpStats    *HTTPStats
	grpcStats    *GRPCStats
	dbStats      *DatabaseStats
	cacheStats   *CacheStats
	runtimeStats *RuntimeStats
}

// ApplicationCollectorConfig 应用收集器配置
type ApplicationCollectorConfig struct {
	Interval        time.Duration     `yaml:"interval"`
	Enabled         bool              `yaml:"enabled"`
	Labels          map[string]string `yaml:"labels"`
	CollectHTTP     bool              `yaml:"collect_http"`
	CollectGRPC     bool              `yaml:"collect_grpc"`
	CollectDatabase bool              `yaml:"collect_database"`
	CollectCache    bool              `yaml:"collect_cache"`
	CollectRuntime  bool              `yaml:"collect_runtime"`
}

// HTTPStats HTTP统计
type HTTPStats struct {
	RequestsTotal    map[string]uint64 // method:status -> count
	RequestDurations []time.Duration
	ErrorRate        float64
	LastReset        time.Time
}

// GRPCStats GRPC统计
type GRPCStats struct {
	RequestsTotal    map[string]uint64 // method:status -> count
	RequestDurations []time.Duration
	ErrorRate        float64
	LastReset        time.Time
}

// DatabaseStats 数据库统计
type DatabaseStats struct {
	ConnectionsActive int64
	ConnectionsIdle   int64
	ConnectionsMax    int64
	QueriesTotal      uint64
	SlowQueries       uint64
	QueryDurations    []time.Duration
	LastReset         time.Time
}

// CacheStats 缓存统计
type CacheStats struct {
	HitCount    uint64
	MissCount   uint64
	KeysTotal   uint64
	MemoryUsage uint64
	Evictions   uint64
	LastReset   time.Time
}

// RuntimeStats 运行时统计
type RuntimeStats struct {
	Goroutines   int
	HeapSize     uint64
	HeapUsed     uint64
	HeapObjects  uint64
	GCCount      uint32
	GCDuration   time.Duration
	LastGC       time.Time
	LastReset    time.Time
}

// NewApplicationCollector 创建应用指标收集器
func NewApplicationCollector(config ApplicationCollectorConfig, db *sql.DB, redisClient *redis.Client) *ApplicationCollector {
	labels := map[string]string{
		"collector": "application",
		"service":   "core-services",
	}
	
	// 添加自定义标签
	for k, v := range config.Labels {
		labels[k] = v
	}
	
	collector := &ApplicationCollector{
		name:            "application",
		interval:        config.Interval,
		enabled:         config.Enabled,
		labels:          labels,
		collectHTTP:     config.CollectHTTP,
		collectGRPC:     config.CollectGRPC,
		collectDatabase: config.CollectDatabase,
		collectCache:    config.CollectCache,
		collectRuntime:  config.CollectRuntime,
		db:              db,
		redisClient:     redisClient,
		httpStats:       &HTTPStats{RequestsTotal: make(map[string]uint64), LastReset: time.Now()},
		grpcStats:       &GRPCStats{RequestsTotal: make(map[string]uint64), LastReset: time.Now()},
		dbStats:         &DatabaseStats{LastReset: time.Now()},
		cacheStats:      &CacheStats{LastReset: time.Now()},
		runtimeStats:    &RuntimeStats{LastReset: time.Now()},
	}
	
	collector.initPrometheusMetrics()
	return collector
}

// initPrometheusMetrics 初始化Prometheus指标
func (c *ApplicationCollector) initPrometheusMetrics() {
	// HTTP指标
	c.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "status", "endpoint"},
	)
	
	c.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	
	// GRPC指标
	c.grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)
	
	c.grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
	
	// 数据库指标
	c.dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)
	
	c.dbConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_idle",
			Help: "Number of idle database connections",
		},
	)
	
	c.dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)
	
	c.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)
	
	// 缓存指标
	c.cacheOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "result"},
	)
	
	c.cacheHitRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "Cache hit rate",
		},
	)
}

// GetName 获取收集器名称
func (c *ApplicationCollector) GetName() string {
	return c.name
}

// GetCategory 获取收集器分类
func (c *ApplicationCollector) GetCategory() models.MetricCategory {
	return models.CategoryApplication
}

// GetInterval 获取收集间隔
func (c *ApplicationCollector) GetInterval() time.Duration {
	return c.interval
}

// IsEnabled 检查是否启用
func (c *ApplicationCollector) IsEnabled() bool {
	return c.enabled
}

// Start 启动收集器
func (c *ApplicationCollector) Start(ctx context.Context) error {
	if !c.enabled {
		return nil
	}
	
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := c.Collect(ctx); err != nil {
				fmt.Printf("Application collector error: %v\n", err)
			}
		}
	}
}

// Stop 停止收集器
func (c *ApplicationCollector) Stop() error {
	c.enabled = false
	return nil
}

// Health 健康检查
func (c *ApplicationCollector) Health() error {
	if !c.enabled {
		return fmt.Errorf("application collector is disabled")
	}
	
	// 检查数据库连接
	if c.collectDatabase && c.db != nil {
		if err := c.db.Ping(); err != nil {
			return fmt.Errorf("database connection failed: %w", err)
		}
	}
	
	// 检查Redis连接
	if c.collectCache && c.redisClient != nil {
		if err := c.redisClient.Ping(context.Background()).Err(); err != nil {
			return fmt.Errorf("redis connection failed: %w", err)
		}
	}
	
	return nil
}

// Collect 收集指标
func (c *ApplicationCollector) Collect(ctx context.Context) ([]models.Metric, error) {
	if !c.enabled {
		return nil, nil
	}
	
	var metrics []models.Metric
	now := time.Now()
	
	// 收集HTTP指标
	if c.collectHTTP {
		httpMetrics, err := c.collectHTTPMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect HTTP metrics: %w", err)
		}
		metrics = append(metrics, httpMetrics...)
	}
	
	// 收集GRPC指标
	if c.collectGRPC {
		grpcMetrics, err := c.collectGRPCMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect gRPC metrics: %w", err)
		}
		metrics = append(metrics, grpcMetrics...)
	}
	
	// 收集数据库指标
	if c.collectDatabase {
		dbMetrics, err := c.collectDatabaseMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect database metrics: %w", err)
		}
		metrics = append(metrics, dbMetrics...)
	}
	
	// 收集缓存指标
	if c.collectCache {
		cacheMetrics, err := c.collectCacheMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect cache metrics: %w", err)
		}
		metrics = append(metrics, cacheMetrics...)
	}
	
	// 收集运行时指标
	if c.collectRuntime {
		runtimeMetrics, err := c.collectRuntimeMetrics(now)
		if err != nil {
			return nil, fmt.Errorf("failed to collect runtime metrics: %w", err)
		}
		metrics = append(metrics, runtimeMetrics...)
	}
	
	return metrics, nil
}

// collectHTTPMetrics 收集HTTP指标
func (c *ApplicationCollector) collectHTTPMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 总请求数
	var totalRequests uint64
	for _, count := range c.httpStats.RequestsTotal {
		totalRequests += count
	}
	
	metric := models.NewMetric("application_http_requests_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(totalRequests)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "requests"
	metric.Description = "Total HTTP requests"
	metrics = append(metrics, *metric)
	
	// 按状态码分组的请求数
	for key, count := range c.httpStats.RequestsTotal {
		labels := make(map[string]string)
		for k, v := range c.labels {
			labels[k] = v
		}
		// 解析key (method:status)
		// 这里简化处理，实际应该解析method和status
		labels["status"] = key
		
		metric := models.NewMetric("application_http_requests_by_status", models.MetricTypeCounter, models.CategoryApplication).
			WithLabels(labels).
			WithValue(float64(count)).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "requests"
		metric.Description = "HTTP requests by status code"
		metrics = append(metrics, *metric)
	}
	
	// 请求持续时间统计
	if len(c.httpStats.RequestDurations) > 0 {
		durations := c.httpStats.RequestDurations
		
		// 平均响应时间
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		avgDuration := total / time.Duration(len(durations))
		
		metric := models.NewMetric("application_http_request_duration_avg_seconds", models.MetricTypeGauge, models.CategoryApplication).
			WithLabels(c.labels).
			WithValue(avgDuration.Seconds()).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Average HTTP request duration"
		metrics = append(metrics, *metric)
		
		// 最大响应时间
		var maxDuration time.Duration
		for _, d := range durations {
			if d > maxDuration {
				maxDuration = d
			}
		}
		
		metric = models.NewMetric("application_http_request_duration_max_seconds", models.MetricTypeGauge, models.CategoryApplication).
			WithLabels(c.labels).
			WithValue(maxDuration.Seconds()).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Maximum HTTP request duration"
		metrics = append(metrics, *metric)
	}
	
	// 错误率
	metric = models.NewMetric("application_http_error_rate", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(c.httpStats.ErrorRate).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "percent"
	metric.Description = "HTTP error rate"
	metrics = append(metrics, *metric)
	
	// 每秒请求数
	timeSinceReset := timestamp.Sub(c.httpStats.LastReset).Seconds()
	if timeSinceReset > 0 {
		rps := float64(totalRequests) / timeSinceReset
		metric := models.NewMetric("application_http_requests_per_second", models.MetricTypeGauge, models.CategoryApplication).
			WithLabels(c.labels).
			WithValue(rps).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "requests_per_second"
		metric.Description = "HTTP requests per second"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectGRPCMetrics 收集GRPC指标
func (c *ApplicationCollector) collectGRPCMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// 总请求数
	var totalRequests uint64
	for _, count := range c.grpcStats.RequestsTotal {
		totalRequests += count
	}
	
	metric := models.NewMetric("application_grpc_requests_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(totalRequests)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "requests"
	metric.Description = "Total gRPC requests"
	metrics = append(metrics, *metric)
	
	// 错误率
	metric = models.NewMetric("application_grpc_error_rate", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(c.grpcStats.ErrorRate).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "percent"
	metric.Description = "gRPC error rate"
	metrics = append(metrics, *metric)
	
	// 平均响应时间
	if len(c.grpcStats.RequestDurations) > 0 {
		var total time.Duration
		for _, d := range c.grpcStats.RequestDurations {
			total += d
		}
		avgDuration := total / time.Duration(len(c.grpcStats.RequestDurations))
		
		metric := models.NewMetric("application_grpc_request_duration_avg_seconds", models.MetricTypeGauge, models.CategoryApplication).
			WithLabels(c.labels).
			WithValue(avgDuration.Seconds()).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Average gRPC request duration"
		metrics = append(metrics, *metric)
	}
	
	return metrics, nil
}

// collectDatabaseMetrics 收集数据库指标
func (c *ApplicationCollector) collectDatabaseMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	if c.db == nil {
		return metrics, nil
	}
	
	// 获取数据库统计信息
	stats := c.db.Stats()
	
	// 活跃连接数
	metric := models.NewMetric("application_database_connections_active", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(stats.InUse)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Active database connections"
	metrics = append(metrics, *metric)
	
	// 空闲连接数
	metric = models.NewMetric("application_database_connections_idle", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(stats.Idle)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Idle database connections"
	metrics = append(metrics, *metric)
	
	// 最大连接数
	metric = models.NewMetric("application_database_connections_max", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(stats.MaxOpenConnections)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Maximum database connections"
	metrics = append(metrics, *metric)
	
	// 等待连接数
	metric = models.NewMetric("application_database_connections_waiting", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(stats.WaitCount)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "connections"
	metric.Description = "Waiting database connections"
	metrics = append(metrics, *metric)
	
	// 等待时间
	metric = models.NewMetric("application_database_wait_duration_seconds", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(stats.WaitDuration.Seconds()).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "seconds"
	metric.Description = "Database connection wait duration"
	metrics = append(metrics, *metric)
	
	// 查询总数
	metric = models.NewMetric("application_database_queries_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(c.dbStats.QueriesTotal)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "queries"
	metric.Description = "Total database queries"
	metrics = append(metrics, *metric)
	
	// 慢查询数
	metric = models.NewMetric("application_database_slow_queries_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(c.dbStats.SlowQueries)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "queries"
	metric.Description = "Total slow database queries"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// collectCacheMetrics 收集缓存指标
func (c *ApplicationCollector) collectCacheMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	if c.redisClient == nil {
		return metrics, nil
	}
	
	ctx := context.Background()
	
	// 获取Redis信息
	info, err := c.redisClient.Info(ctx, "memory", "stats", "keyspace").Result()
	if err != nil {
		return nil, err
	}
	
	// 解析Redis信息（简化实现）
	// 实际应该解析info字符串获取具体指标
	
	// 命中率
	totalOps := c.cacheStats.HitCount + c.cacheStats.MissCount
	var hitRate float64
	if totalOps > 0 {
		hitRate = float64(c.cacheStats.HitCount) / float64(totalOps) * 100
	}
	
	metric := models.NewMetric("application_cache_hit_rate", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(hitRate).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "percent"
	metric.Description = "Cache hit rate"
	metrics = append(metrics, *metric)
	
	// 缓存操作总数
	metric = models.NewMetric("application_cache_operations_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(totalOps)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "operations"
	metric.Description = "Total cache operations"
	metrics = append(metrics, *metric)
	
	// 键总数
	metric = models.NewMetric("application_cache_keys_total", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(c.cacheStats.KeysTotal)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "keys"
	metric.Description = "Total cache keys"
	metrics = append(metrics, *metric)
	
	// 内存使用
	metric = models.NewMetric("application_cache_memory_usage_bytes", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(c.cacheStats.MemoryUsage)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Cache memory usage"
	metrics = append(metrics, *metric)
	
	// 驱逐数
	metric = models.NewMetric("application_cache_evictions_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(c.cacheStats.Evictions)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "evictions"
	metric.Description = "Total cache evictions"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// collectRuntimeMetrics 收集运行时指标
func (c *ApplicationCollector) collectRuntimeMetrics(timestamp time.Time) ([]models.Metric, error) {
	var metrics []models.Metric
	
	// Goroutine数量
	metric := models.NewMetric("application_goroutines_total", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(runtime.NumGoroutine())).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "goroutines"
	metric.Description = "Number of goroutines"
	metrics = append(metrics, *metric)
	
	// 内存统计
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// 堆大小
	metric = models.NewMetric("application_memory_heap_size_bytes", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(memStats.HeapSys)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Heap size in bytes"
	metrics = append(metrics, *metric)
	
	// 堆使用
	metric = models.NewMetric("application_memory_heap_used_bytes", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(memStats.HeapInuse)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Heap used in bytes"
	metrics = append(metrics, *metric)
	
	// 堆对象数
	metric = models.NewMetric("application_memory_heap_objects", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(memStats.HeapObjects)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "objects"
	metric.Description = "Number of heap objects"
	metrics = append(metrics, *metric)
	
	// GC次数
	metric = models.NewMetric("application_gc_count_total", models.MetricTypeCounter, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(memStats.NumGC)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "count"
	metric.Description = "Total garbage collections"
	metrics = append(metrics, *metric)
	
	// GC暂停时间
	if memStats.NumGC > 0 {
		// 最近一次GC暂停时间
		lastPause := memStats.PauseNs[(memStats.NumGC+255)%256]
		metric = models.NewMetric("application_gc_pause_duration_seconds", models.MetricTypeGauge, models.CategoryApplication).
			WithLabels(c.labels).
			WithValue(float64(lastPause) / 1e9).
			WithSource(c.name)
		metric.Timestamp = timestamp
		metric.Unit = "seconds"
		metric.Description = "Last GC pause duration"
		metrics = append(metrics, *metric)
	}
	
	// 下次GC阈值
	metric = models.NewMetric("application_memory_next_gc_bytes", models.MetricTypeGauge, models.CategoryApplication).
		WithLabels(c.labels).
		WithValue(float64(memStats.NextGC)).
		WithSource(c.name)
	metric.Timestamp = timestamp
	metric.Unit = "bytes"
	metric.Description = "Next GC threshold"
	metrics = append(metrics, *metric)
	
	return metrics, nil
}

// RecordHTTPRequest 记录HTTP请求
func (c *ApplicationCollector) RecordHTTPRequest(method, status, endpoint string, duration time.Duration) {
	if !c.collectHTTP {
		return
	}
	
	key := fmt.Sprintf("%s:%s", method, status)
	c.httpStats.RequestsTotal[key]++
	c.httpStats.RequestDurations = append(c.httpStats.RequestDurations, duration)
	
	// 更新Prometheus指标
	if c.httpRequestsTotal != nil {
		c.httpRequestsTotal.WithLabelValues(method, status, endpoint).Inc()
	}
	if c.httpRequestDuration != nil {
		c.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
	}
}

// RecordGRPCRequest 记录GRPC请求
func (c *ApplicationCollector) RecordGRPCRequest(method, status string, duration time.Duration) {
	if !c.collectGRPC {
		return
	}
	
	key := fmt.Sprintf("%s:%s", method, status)
	c.grpcStats.RequestsTotal[key]++
	c.grpcStats.RequestDurations = append(c.grpcStats.RequestDurations, duration)
	
	// 更新Prometheus指标
	if c.grpcRequestsTotal != nil {
		c.grpcRequestsTotal.WithLabelValues(method, status).Inc()
	}
	if c.grpcRequestDuration != nil {
		c.grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
	}
}

// RecordDatabaseQuery 记录数据库查询
func (c *ApplicationCollector) RecordDatabaseQuery(operation, table string, duration time.Duration, isSlowQuery bool) {
	if !c.collectDatabase {
		return
	}
	
	c.dbStats.QueriesTotal++
	c.dbStats.QueryDurations = append(c.dbStats.QueryDurations, duration)
	
	if isSlowQuery {
		c.dbStats.SlowQueries++
	}
	
	// 更新Prometheus指标
	if c.dbQueriesTotal != nil {
		c.dbQueriesTotal.WithLabelValues(operation, table).Inc()
	}
	if c.dbQueryDuration != nil {
		c.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	}
}

// RecordCacheOperation 记录缓存操作
func (c *ApplicationCollector) RecordCacheOperation(operation string, hit bool) {
	if !c.collectCache {
		return
	}
	
	if hit {
		c.cacheStats.HitCount++
	} else {
		c.cacheStats.MissCount++
	}
	
	// 更新Prometheus指标
	result := "miss"
	if hit {
		result = "hit"
	}
	if c.cacheOperationsTotal != nil {
		c.cacheOperationsTotal.WithLabelValues(operation, result).Inc()
	}
	
	// 更新命中率
	total := c.cacheStats.HitCount + c.cacheStats.MissCount
	if total > 0 && c.cacheHitRate != nil {
		hitRate := float64(c.cacheStats.HitCount) / float64(total)
		c.cacheHitRate.Set(hitRate)
	}
}

// 确保实现了接口
var _ interfaces.MetricCollector = (*ApplicationCollector)(nil)