package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter 创建新的限流器
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter 获取指定IP的限流器
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

// CleanupLimiters 清理过期的限流器
func (rl *RateLimiter) CleanupLimiters() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, limiter := range rl.limiters {
				// 如果限流器在过去1分钟内没有被使用，则删除它
				if limiter.Allow() {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// 全局限流器实例
var globalRateLimiter = NewRateLimiter(rate.Every(time.Second), 100) // 每秒100个请求

func init() {
	// 启动清理协程
	go globalRateLimiter.CleanupLimiters()
}

// RateLimit 限流中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := globalRateLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests from this IP",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithConfig 带配置的限流中间件
func RateLimitWithConfig(requestsPerSecond int, burst int) gin.HandlerFunc {
	rateLimiter := NewRateLimiter(rate.Every(time.Second/time.Duration(requestsPerSecond)), burst)
	go rateLimiter.CleanupLimiters()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rateLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests from this IP",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}