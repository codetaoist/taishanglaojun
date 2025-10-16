package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
	
	"github.com/codetaoist/taishanglaojun/infrastructure/api-gateway/internal/logger"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 全局限流
	GlobalRPS   int `yaml:"global_rps"`   // 每秒请求数
	GlobalBurst int `yaml:"global_burst"` // 突发请求数
	
	// 基于IP的限流
	IPRPS   int `yaml:"ip_rps"`
	IPBurst int `yaml:"ip_burst"`
	
	// 基于用户的限流
	UserRPS   int `yaml:"user_rps"`
	UserBurst int `yaml:"user_burst"`
	
	// 基于路径的限流
	PathLimits map[string]PathLimit `yaml:"path_limits"`
	
	// Redis配置（用于分布式限流）
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
	RedisDB       int    `yaml:"redis_db"`
	
	// 限流窗口大小
	WindowSize time.Duration `yaml:"window_size"`
	
	// 跳过限流的路径
	SkipPaths []string `yaml:"skip_paths"`
}

// PathLimit 路径限流配置
type PathLimit struct {
	RPS   int `yaml:"rps"`
	Burst int `yaml:"burst"`
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	config      *RateLimitConfig
	redisClient *redis.Client
	logger      logger.Logger
	
	// 内存限流器
	globalLimiter *rate.Limiter
	ipLimiters    sync.Map // map[string]*rate.Limiter
	userLimiters  sync.Map // map[string]*rate.Limiter
	pathLimiters  sync.Map // map[string]*rate.Limiter
	
	// 清理器
	cleanupTicker *time.Ticker
	stopCh        chan struct{}
}

// RateLimitInfo 限流信息
type RateLimitInfo struct {
	Limit     int   `json:"limit"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
	RetryAfter int  `json:"retry_after,omitempty"`
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(config *RateLimitConfig, log logger.Logger) (*RateLimitMiddleware, error) {
	rl := &RateLimitMiddleware{
		config: config,
		logger: log,
		stopCh: make(chan struct{}),
	}
	
	// 创建全局限流器
	if config.GlobalRPS > 0 {
		rl.globalLimiter = rate.NewLimiter(rate.Limit(config.GlobalRPS), config.GlobalBurst)
	}
	
	// 如果配置了Redis，则创建Redis客户端
	if config.RedisAddr != "" {
		rl.redisClient = redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		})
		
		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		if err := rl.redisClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}
	
	// 启动清理器
	rl.startCleanup()
	
	return rl, nil
}

// Handler 返回Gin中间件处理函数
func (rl *RateLimitMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// 检查是否跳过限流
		if rl.shouldSkipRateLimit(path) {
			c.Next()
			return
		}
		
		// 获取客户端标识
		clientIP := c.ClientIP()
		userID := rl.getUserID(c)
		
		// 检查各种限流
		if !rl.checkGlobalLimit(c) {
			return
		}
		
		if !rl.checkIPLimit(c, clientIP) {
			return
		}
		
		if userID != "" && !rl.checkUserLimit(c, userID) {
			return
		}
		
		if !rl.checkPathLimit(c, path) {
			return
		}
		
		c.Next()
	}
}

// shouldSkipRateLimit 检查是否应该跳过限流
func (rl *RateLimitMiddleware) shouldSkipRateLimit(path string) bool {
	for _, skipPath := range rl.config.SkipPaths {
		if rl.matchPath(path, skipPath) {
			return true
		}
	}
	return false
}

// matchPath 路径匹配（支持通配符）
func (rl *RateLimitMiddleware) matchPath(path, pattern string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}

// getUserID 获取用户ID
func (rl *RateLimitMiddleware) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// checkGlobalLimit 检查全局限流
func (rl *RateLimitMiddleware) checkGlobalLimit(c *gin.Context) bool {
	if rl.globalLimiter == nil {
		return true
	}
	
	if rl.redisClient != nil {
		return rl.checkRedisLimit(c, "global", rl.config.GlobalRPS, rl.config.GlobalBurst)
	}
	
	if !rl.globalLimiter.Allow() {
		rl.respondWithRateLimit(c, &RateLimitInfo{
			Limit:      rl.config.GlobalRPS,
			Remaining:  0,
			Reset:      time.Now().Add(time.Second).Unix(),
			RetryAfter: 1,
		})
		return false
	}
	
	return true
}

// checkIPLimit 检查IP限流
func (rl *RateLimitMiddleware) checkIPLimit(c *gin.Context, clientIP string) bool {
	if rl.config.IPRPS <= 0 {
		return true
	}
	
	if rl.redisClient != nil {
		key := fmt.Sprintf("ip:%s", clientIP)
		return rl.checkRedisLimit(c, key, rl.config.IPRPS, rl.config.IPBurst)
	}
	
	// 获取或创建IP限流器
	limiterInterface, _ := rl.ipLimiters.LoadOrStore(clientIP, rate.NewLimiter(
		rate.Limit(rl.config.IPRPS),
		rl.config.IPBurst,
	))
	limiter := limiterInterface.(*rate.Limiter)
	
	if !limiter.Allow() {
		rl.respondWithRateLimit(c, &RateLimitInfo{
			Limit:      rl.config.IPRPS,
			Remaining:  0,
			Reset:      time.Now().Add(time.Second).Unix(),
			RetryAfter: 1,
		})
		return false
	}
	
	return true
}

// checkUserLimit 检查用户限流
func (rl *RateLimitMiddleware) checkUserLimit(c *gin.Context, userID string) bool {
	if rl.config.UserRPS <= 0 {
		return true
	}
	
	if rl.redisClient != nil {
		key := fmt.Sprintf("user:%s", userID)
		return rl.checkRedisLimit(c, key, rl.config.UserRPS, rl.config.UserBurst)
	}
	
	// 获取或创建用户限流器
	limiterInterface, _ := rl.userLimiters.LoadOrStore(userID, rate.NewLimiter(
		rate.Limit(rl.config.UserRPS),
		rl.config.UserBurst,
	))
	limiter := limiterInterface.(*rate.Limiter)
	
	if !limiter.Allow() {
		rl.respondWithRateLimit(c, &RateLimitInfo{
			Limit:      rl.config.UserRPS,
			Remaining:  0,
			Reset:      time.Now().Add(time.Second).Unix(),
			RetryAfter: 1,
		})
		return false
	}
	
	return true
}

// checkPathLimit 检查路径限流
func (rl *RateLimitMiddleware) checkPathLimit(c *gin.Context, path string) bool {
	var pathLimit *PathLimit
	
	// 查找匹配的路径限制
	for pattern, limit := range rl.config.PathLimits {
		if rl.matchPath(path, pattern) {
			pathLimit = &limit
			break
		}
	}
	
	if pathLimit == nil || pathLimit.RPS <= 0 {
		return true
	}
	
	if rl.redisClient != nil {
		key := fmt.Sprintf("path:%s", path)
		return rl.checkRedisLimit(c, key, pathLimit.RPS, pathLimit.Burst)
	}
	
	// 获取或创建路径限流器
	limiterInterface, _ := rl.pathLimiters.LoadOrStore(path, rate.NewLimiter(
		rate.Limit(pathLimit.RPS),
		pathLimit.Burst,
	))
	limiter := limiterInterface.(*rate.Limiter)
	
	if !limiter.Allow() {
		rl.respondWithRateLimit(c, &RateLimitInfo{
			Limit:      pathLimit.RPS,
			Remaining:  0,
			Reset:      time.Now().Add(time.Second).Unix(),
			RetryAfter: 1,
		})
		return false
	}
	
	return true
}

// checkRedisLimit 使用Redis检查限流
func (rl *RateLimitMiddleware) checkRedisLimit(c *gin.Context, key string, rps, burst int) bool {
	ctx := c.Request.Context()
	now := time.Now()
	window := rl.config.WindowSize
	if window == 0 {
		window = time.Second
	}
	
	// 使用滑动窗口算法
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	windowStart := now.Truncate(window).Unix()
	
	pipe := rl.redisClient.Pipeline()
	
	// 增加计数
	incrCmd := pipe.Incr(ctx, redisKey)
	// 设置过期时间
	pipe.ExpireAt(ctx, redisKey, time.Unix(windowStart, 0).Add(window))
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		rl.logger.Errorf("Redis rate limit error: %v", err)
		// Redis错误时允许请求通过
		return true
	}
	
	count := int(incrCmd.Val())
	remaining := rps - count
	if remaining < 0 {
		remaining = 0
	}
	
	// 设置响应头
	c.Header("X-RateLimit-Limit", strconv.Itoa(rps))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(windowStart+int64(window.Seconds()), 10))
	
	if count > rps {
		retryAfter := int(window.Seconds()) - int(now.Unix()-windowStart)
		if retryAfter < 0 {
			retryAfter = 1
		}
		
		rl.respondWithRateLimit(c, &RateLimitInfo{
			Limit:      rps,
			Remaining:  0,
			Reset:      windowStart + int64(window.Seconds()),
			RetryAfter: retryAfter,
		})
		return false
	}
	
	return true
}

// respondWithRateLimit 返回限流响应
func (rl *RateLimitMiddleware) respondWithRateLimit(c *gin.Context, info *RateLimitInfo) {
	// 设置响应头
	c.Header("X-RateLimit-Limit", strconv.Itoa(info.Limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(info.Reset, 10))
	
	if info.RetryAfter > 0 {
		c.Header("Retry-After", strconv.Itoa(info.RetryAfter))
	}
	
	rl.logger.WithFields(map[string]interface{}{
		"path":        c.Request.URL.Path,
		"method":      c.Request.Method,
		"remote_addr": c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
		"limit":       info.Limit,
		"remaining":   info.Remaining,
	}).Warn("Rate limit exceeded")
	
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":       "Rate limit exceeded",
		"message":     "Too many requests",
		"code":        http.StatusTooManyRequests,
		"rate_limit":  info,
	})
	c.Abort()
}

// startCleanup 启动清理器
func (rl *RateLimitMiddleware) startCleanup() {
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	
	go func() {
		for {
			select {
			case <-rl.cleanupTicker.C:
				rl.cleanup()
			case <-rl.stopCh:
				return
			}
		}
	}()
}

// cleanup 清理过期的限流器
func (rl *RateLimitMiddleware) cleanup() {
	now := time.Now()
	
	// 清理IP限流器
	rl.ipLimiters.Range(func(key, value interface{}) bool {
		limiter := value.(*rate.Limiter)
		// 如果限流器长时间未使用，则删除
		if limiter.TokensAt(now) == float64(rl.config.IPBurst) {
			rl.ipLimiters.Delete(key)
		}
		return true
	})
	
	// 清理用户限流器
	rl.userLimiters.Range(func(key, value interface{}) bool {
		limiter := value.(*rate.Limiter)
		if limiter.TokensAt(now) == float64(rl.config.UserBurst) {
			rl.userLimiters.Delete(key)
		}
		return true
	})
	
	// 清理路径限流器
	rl.pathLimiters.Range(func(key, value interface{}) bool {
		limiter := value.(*rate.Limiter)
		// 获取对应的路径限制配置
		pathStr := key.(string)
		var burst int
		for pattern, limit := range rl.config.PathLimits {
			if rl.matchPath(pathStr, pattern) {
				burst = limit.Burst
				break
			}
		}
		if burst > 0 && limiter.TokensAt(now) == float64(burst) {
			rl.pathLimiters.Delete(key)
		}
		return true
	})
}

// GetRateLimitInfo 获取限流信息
func (rl *RateLimitMiddleware) GetRateLimitInfo(c *gin.Context) *RateLimitInfo {
	// 这里可以实现获取当前限流状态的逻辑
	return &RateLimitInfo{
		Limit:     rl.config.GlobalRPS,
		Remaining: rl.config.GlobalRPS, // 简化实现
		Reset:     time.Now().Add(time.Second).Unix(),
	}
}

// Close 关闭中间件资源
func (rl *RateLimitMiddleware) Close() error {
	close(rl.stopCh)
	
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
	
	if rl.redisClient != nil {
		return rl.redisClient.Close()
	}
	
	return nil
}