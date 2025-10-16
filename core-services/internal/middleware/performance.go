package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// PerformanceMiddleware 性能优化中间件
type PerformanceMiddleware struct {
	cache       *redis.Client
	logger      *zap.Logger
	rateLimiter *RateLimiter
	compressor  *GzipCompressor
	metrics     *PerformanceMetrics
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	EnableCache       bool
	EnableRateLimit   bool
	EnableCompression bool
	EnableMetrics     bool
	CacheTTL          time.Duration
	RateLimit         rate.Limit
	RateBurst         int
	CompressionLevel  int
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	mu                sync.RWMutex
	RequestCount      int64
	CacheHits         int64
	CacheMisses       int64
	CompressionSaved  int64
	AverageLatency    time.Duration
	TotalLatency      time.Duration
}

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
}

// GzipCompressor Gzip压缩器
type GzipCompressor struct {
	level int
	pool  sync.Pool
}

// NewPerformanceMiddleware 创建性能中间件
func NewPerformanceMiddleware(cache *redis.Client, logger *zap.Logger, config PerformanceConfig) *PerformanceMiddleware {
	pm := &PerformanceMiddleware{
		cache:   cache,
		logger:  logger,
		metrics: &PerformanceMetrics{},
	}

	if config.EnableRateLimit {
		pm.rateLimiter = NewRateLimiter(config.RateLimit, config.RateBurst)
	}

	if config.EnableCompression {
		pm.compressor = NewGzipCompressor(config.CompressionLevel)
	}

	return pm
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    limit,
		burst:    burst,
	}
}

// NewGzipCompressor 创建Gzip压缩器
func NewGzipCompressor(level int) *GzipCompressor {
	return &GzipCompressor{
		level: level,
		pool: sync.Pool{
			New: func() interface{} {
				w, _ := gzip.NewWriterLevel(io.Discard, level)
				return w
			},
		},
	}
}

// CacheMiddleware 缓存中间件
func (pm *PerformanceMiddleware) CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只缓存GET请求
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := pm.generateCacheKey(c)

		// 尝试从缓存获取
		if pm.cache != nil {
			cached, err := pm.cache.Get(context.Background(), cacheKey).Result()
			if err == nil {
				pm.metrics.incrementCacheHits()
				
				var response CachedResponse
				if err := json.Unmarshal([]byte(cached), &response); err == nil {
					// 设置响应头
					for key, value := range response.Headers {
						c.Header(key, value)
					}
					c.Data(response.StatusCode, response.ContentType, response.Body)
					c.Abort()
					return
				}
			}
			pm.metrics.incrementCacheMisses()
		}

		// 创建响应写入器
		writer := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBuffer(nil),
			headers:       make(map[string]string),
		}
		c.Writer = writer

		c.Next()

		// 缓存响应
		if pm.cache != nil && writer.statusCode == http.StatusOK {
			response := CachedResponse{
				StatusCode:  writer.statusCode,
				ContentType: writer.Header().Get("Content-Type"),
				Headers:     writer.headers,
				Body:        writer.body.Bytes(),
			}

			if data, err := json.Marshal(response); err == nil {
				pm.cache.Set(context.Background(), cacheKey, data, ttl)
			}
		}
	}
}

// RateLimitMiddleware 限流中间件
func (pm *PerformanceMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if pm.rateLimiter == nil {
			c.Next()
			return
		}

		// 获取客户端IP
		clientIP := c.ClientIP()
		
		// 检查限流
		if !pm.rateLimiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CompressionMiddleware 压缩中间件
func (pm *PerformanceMiddleware) CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if pm.compressor == nil {
			c.Next()
			return
		}

		// 检查客户端是否支持gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 创建压缩写入器
		writer := &CompressResponseWriter{
			ResponseWriter: c.Writer,
			compressor:     pm.compressor,
		}
		c.Writer = writer
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		defer writer.Close()
		c.Next()
	}
}

// MetricsMiddleware 指标中间件
func (pm *PerformanceMiddleware) MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// 记录指标
		latency := time.Since(start)
		pm.metrics.recordRequest(latency)

		// 记录日志
		pm.logger.Info("Request processed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
		)
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.limit, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter.Allow()
}

// generateCacheKey 生成缓存键
func (pm *PerformanceMiddleware) generateCacheKey(c *gin.Context) string {
	return fmt.Sprintf("cache:%s:%s:%s", 
		c.Request.Method, 
		c.Request.URL.Path, 
		c.Request.URL.RawQuery,
	)
}

// CachedResponse 缓存响应
type CachedResponse struct {
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
}

// CacheResponseWriter 缓存响应写入器
type CacheResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	headers    map[string]string
	statusCode int
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *CacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *CacheResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// CompressResponseWriter 压缩响应写入器
type CompressResponseWriter struct {
	gin.ResponseWriter
	compressor *GzipCompressor
	writer     *gzip.Writer
}

func (w *CompressResponseWriter) Write(data []byte) (int, error) {
	if w.writer == nil {
		w.writer = w.compressor.pool.Get().(*gzip.Writer)
		w.writer.Reset(w.ResponseWriter)
	}
	return w.writer.Write(data)
}

func (w *CompressResponseWriter) Close() error {
	if w.writer != nil {
		err := w.writer.Close()
		w.compressor.pool.Put(w.writer)
		return err
	}
	return nil
}

// PerformanceMetrics 方法
func (pm *PerformanceMetrics) incrementCacheHits() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.CacheHits++
}

func (pm *PerformanceMetrics) incrementCacheMisses() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.CacheMisses++
}

func (pm *PerformanceMetrics) recordRequest(latency time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.RequestCount++
	pm.TotalLatency += latency
	pm.AverageLatency = pm.TotalLatency / time.Duration(pm.RequestCount)
}

// GetMetrics 获取性能指标
func (pm *PerformanceMiddleware) GetMetrics() *PerformanceMetrics {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()
	
	return &PerformanceMetrics{
		RequestCount:     pm.metrics.RequestCount,
		CacheHits:        pm.metrics.CacheHits,
		CacheMisses:      pm.metrics.CacheMisses,
		CompressionSaved: pm.metrics.CompressionSaved,
		AverageLatency:   pm.metrics.AverageLatency,
		TotalLatency:     pm.metrics.TotalLatency,
	}
}

// DatabaseConnectionPool 数据库连接池优化
type DatabaseConnectionPool struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

// NewDatabaseConnectionPool 创建数据库连接池
func NewDatabaseConnectionPool() *DatabaseConnectionPool {
	return &DatabaseConnectionPool{
		maxOpenConns:    25,  // 最大打开连接数
		maxIdleConns:    10,  // 最大空闲连接数
		connMaxLifetime: time.Hour,   // 连接最大生存时间
		connMaxIdleTime: time.Minute * 10, // 连接最大空闲时间
	}
}

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
	cache  map[string]interface{}
	mu     sync.RWMutex
	logger *zap.Logger
}

// NewQueryOptimizer 创建查询优化器
func NewQueryOptimizer(logger *zap.Logger) *QueryOptimizer {
	return &QueryOptimizer{
		cache:  make(map[string]interface{}),
		logger: logger,
	}
}

// OptimizeQuery 优化查询
func (qo *QueryOptimizer) OptimizeQuery(query string, params ...interface{}) (string, []interface{}) {
	// 添加LIMIT子句（如果没有）
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		query += " LIMIT 1000"
	}

	// 添加索引提示
	if strings.Contains(strings.ToUpper(query), "WHERE") {
		// 这里可以添加更复杂的索引优化逻辑
	}

	return query, params
}

// BatchProcessor 批处理器
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	buffer    []interface{}
	mu        sync.Mutex
	processor func([]interface{}) error
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(batchSize int, timeout time.Duration, processor func([]interface{}) error) *BatchProcessor {
	bp := &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		buffer:    make([]interface{}, 0, batchSize),
		processor: processor,
	}

	// 启动定时处理
	go bp.startTimer()
	
	return bp
}

// Add 添加项目到批处理
func (bp *BatchProcessor) Add(item interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, item)
	
	if len(bp.buffer) >= bp.batchSize {
		return bp.flush()
	}
	
	return nil
}

// flush 刷新缓冲区
func (bp *BatchProcessor) flush() error {
	if len(bp.buffer) == 0 {
		return nil
	}

	items := make([]interface{}, len(bp.buffer))
	copy(items, bp.buffer)
	bp.buffer = bp.buffer[:0]

	return bp.processor(items)
}

// startTimer 启动定时器
func (bp *BatchProcessor) startTimer() {
	ticker := time.NewTicker(bp.timeout)
	defer ticker.Stop()

	for range ticker.C {
		bp.mu.Lock()
		bp.flush()
		bp.mu.Unlock()
	}
}