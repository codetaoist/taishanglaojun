package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	RequestsPerWindow int           `json:"requests_per_window"`
	WindowDuration    time.Duration `json:"window_duration"`
	BurstSize         int           `json:"burst_size"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	
	// 不同端点的特殊限制
	LoginLimit    int `json:"login_limit"`     // 登录接口特殊限制
	RegisterLimit int `json:"register_limit"`  // 注册接口特殊限制
}

// DefaultRateLimiterConfig 默认限流配置
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		RequestsPerWindow: 100,
		WindowDuration:    time.Minute,
		BurstSize:         10,
		CleanupInterval:   5 * time.Minute,
		LoginLimit:        5,  // 每分钟最多5次登录尝试
		RegisterLimit:     3,  // 每分钟最多3次注册尝试
	}
}

// ClientInfo 客户端信息
type ClientInfo struct {
	RequestCount int       `json:"request_count"`
	WindowStart  time.Time `json:"window_start"`
	LastRequest  time.Time `json:"last_request"`
	Blocked      bool      `json:"blocked"`
	BlockedUntil time.Time `json:"blocked_until"`
}

// RateLimiter 限流器
type RateLimiter struct {
	config  *RateLimiterConfig
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
	logger  *zap.Logger
	
	// 停止清理协程的通道
	stopCleanup chan struct{}
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config *RateLimiterConfig, logger *zap.Logger) *RateLimiter {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}
	
	rl := &RateLimiter{
		config:      config,
		clients:     make(map[string]*ClientInfo),
		logger:      logger,
		stopCleanup: make(chan struct{}),
	}
	
	// 启动清理协程
	go rl.startCleanup()
	
	return rl
}

// GetClientKey 获取客户端标识
func (rl *RateLimiter) GetClientKey(c *gin.Context) string {
	// 优先使用用户ID（如果已认证）
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	
	// 使用IP地址
	clientIP := c.ClientIP()
	
	// 对于敏感操作，结合User-Agent
	if rl.isSensitiveEndpoint(c.Request.URL.Path) {
		userAgent := c.GetHeader("User-Agent")
		return fmt.Sprintf("ip:%s:ua:%s", clientIP, userAgent)
	}
	
	return fmt.Sprintf("ip:%s", clientIP)
}

// isSensitiveEndpoint 判断是否为敏感端点
func (rl *RateLimiter) isSensitiveEndpoint(path string) bool {
	sensitiveEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/forgot-password",
		"/api/v1/auth/reset-password",
	}
	
	for _, endpoint := range sensitiveEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// GetLimit 获取特定端点的限制
func (rl *RateLimiter) GetLimit(path string) int {
	switch path {
	case "/api/v1/auth/login":
		return rl.config.LoginLimit
	case "/api/v1/auth/register":
		return rl.config.RegisterLimit
	default:
		return rl.config.RequestsPerWindow
	}
}

// IsAllowed 检查请求是否被允许
func (rl *RateLimiter) IsAllowed(clientKey string, path string) (bool, *ClientInfo) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	client, exists := rl.clients[clientKey]
	
	if !exists {
		// 新客户端
		client = &ClientInfo{
			RequestCount: 1,
			WindowStart:  now,
			LastRequest:  now,
			Blocked:      false,
		}
		rl.clients[clientKey] = client
		return true, client
	}
	
	// 检查是否被阻止
	if client.Blocked && now.Before(client.BlockedUntil) {
		return false, client
	}
	
	// 重置阻止状态
	if client.Blocked && now.After(client.BlockedUntil) {
		client.Blocked = false
		client.RequestCount = 0
		client.WindowStart = now
	}
	
	// 检查时间窗口
	if now.Sub(client.WindowStart) >= rl.config.WindowDuration {
		// 新的时间窗口
		client.RequestCount = 1
		client.WindowStart = now
		client.LastRequest = now
		return true, client
	}
	
	// 在当前时间窗口内
	limit := rl.GetLimit(path)
	if client.RequestCount >= limit {
		// 超过限制，阻止客户端
		client.Blocked = true
		client.BlockedUntil = now.Add(rl.config.WindowDuration * 2) // 阻止时间为窗口时间的2倍
		
		rl.logger.Warn("Client rate limited",
			zap.String("client_key", clientKey),
			zap.String("path", path),
			zap.Int("request_count", client.RequestCount),
			zap.Int("limit", limit),
			zap.Time("blocked_until", client.BlockedUntil),
		)
		
		return false, client
	}
	
	// 允许请求
	client.RequestCount++
	client.LastRequest = now
	return true, client
}

// Middleware 限流中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := rl.GetClientKey(c)
		path := c.Request.URL.Path
		
		allowed, clientInfo := rl.IsAllowed(clientKey, path)
		
		if !allowed {
			// 设置响应头
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.GetLimit(path)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(clientInfo.BlockedUntil.Unix(), 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "RATE_LIMIT_EXCEEDED",
				"message": "请求过于频繁，请稍后再试",
				"retry_after": int(time.Until(clientInfo.BlockedUntil).Seconds()),
			})
			c.Abort()
			return
		}
		
		// 设置响应头
		limit := rl.GetLimit(path)
		remaining := limit - clientInfo.RequestCount
		if remaining < 0 {
			remaining = 0
		}
		
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(clientInfo.WindowStart.Add(rl.config.WindowDuration).Unix(), 10))
		
		c.Next()
	}
}

// startCleanup 启动清理协程
func (rl *RateLimiter) startCleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			return
		}
	}
}

// cleanup 清理过期的客户端信息
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	expireTime := rl.config.WindowDuration * 2 // 保留时间为窗口时间的2倍
	
	for clientKey, client := range rl.clients {
		if now.Sub(client.LastRequest) > expireTime && !client.Blocked {
			delete(rl.clients, clientKey)
		}
	}
	
	rl.logger.Debug("Rate limiter cleanup completed",
		zap.Int("remaining_clients", len(rl.clients)),
	)
}

// Stop 停止限流器
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// GetStats 获取统计信息
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	
	totalClients := len(rl.clients)
	blockedClients := 0
	
	for _, client := range rl.clients {
		if client.Blocked && time.Now().Before(client.BlockedUntil) {
			blockedClients++
		}
	}
	
	return map[string]interface{}{
		"total_clients":   totalClients,
		"blocked_clients": blockedClients,
		"config":          rl.config,
	}
}