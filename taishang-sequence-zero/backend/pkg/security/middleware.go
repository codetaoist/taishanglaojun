package security

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	RateLimitPerMinute int
	MaxRequestSize     int64
	AllowedOrigins     []string
	SecureHeaders      bool
	EnableCSRF         bool
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimitPerMinute: 100,
		MaxRequestSize:     10 << 20, // 10MB
		AllowedOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		SecureHeaders:      true,
		EnableCSRF:         true,
	}
}

// RateLimiter 速率限制器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

// NewRateLimiter 创建新的速率限制器
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerMinute) / 60, // 每秒请求数
		burst:    requestsPerMinute / 10,             // 突发请求数
	}
}

// GetLimiter 获取指定IP的限制器
func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}
	return limiter
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(rateLimiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		limiter := rateLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"message": "请稍后再试",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 强制HTTPS (生产环境)
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// 引用者策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否为允许的源
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if gin.Mode() == gin.DebugMode {
			// 开发模式下允许所有源
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestSizeMiddleware 请求大小限制中间件
func RequestSizeMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "请求体过大",
				"message": fmt.Sprintf("请求体大小不能超过 %d 字节", maxSize),
				"code":    "REQUEST_TOO_LARGE",
			})
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// InputValidationMiddleware 输入验证中间件
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查常见的恶意模式
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsMaliciousPattern(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error":   "检测到恶意输入",
						"message": fmt.Sprintf("参数 %s 包含不允许的内容", key),
						"code":    "MALICIOUS_INPUT_DETECTED",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// AuditMiddleware 审计中间件
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ip := getClientIP(c)
		userAgent := c.Request.UserAgent()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// 记录敏感操作
		if isSensitiveOperation(method, path) || status >= 400 {
			logAuditEvent(AuditEvent{
				Timestamp: start,
				IP:        ip,
				UserAgent: userAgent,
				Method:    method,
				Path:      path,
				Status:    status,
				Latency:   latency,
				UserID:    getUserIDFromContext(c),
			})
		}
	}
}

// AuditEvent 审计事件
type AuditEvent struct {
	Timestamp time.Time     `json:"timestamp"`
	IP        string        `json:"ip"`
	UserAgent string        `json:"user_agent"`
	Method    string        `json:"method"`
	Path      string        `json:"path"`
	Status    int           `json:"status"`
	Latency   time.Duration `json:"latency"`
	UserID    *int          `json:"user_id,omitempty"`
}

// getClientIP 获取客户端真实IP
func getClientIP(c *gin.Context) string {
	// 检查 X-Forwarded-For 头
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查 X-Real-IP 头
	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用 RemoteAddr
	ip := c.Request.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// containsMaliciousPattern 检查是否包含恶意模式
func containsMaliciousPattern(input string) bool {
	maliciousPatterns := []string{
		"<script",
		"javascript:",
		"onload=",
		"onerror=",
		"onclick=",
		"SELECT * FROM",
		"DROP TABLE",
		"INSERT INTO",
		"UPDATE SET",
		"DELETE FROM",
		"UNION SELECT",
		"../",
		"../..",
	}

	for _, pattern := range maliciousPatterns {
		if strings.Contains(strings.ToLower(input), strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// isSensitiveOperation 检查是否为敏感操作
func isSensitiveOperation(method, path string) bool {
	sensitivePaths := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/logout",
		"/api/permission",
		"/api/user",
		"/api/admin",
	}

	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return true
		}
	}

	// POST, PUT, DELETE 操作都被认为是敏感的
	return method == "POST" || method == "PUT" || method == "DELETE"
}

// getUserIDFromContext 从上下文中获取用户ID
func getUserIDFromContext(c *gin.Context) *int {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok {
			return &id
		}
	}
	return nil
}

// logAuditEvent 记录审计事件
func logAuditEvent(event AuditEvent) {
	// 这里可以实现具体的日志记录逻辑
	// 例如写入数据库、发送到日志服务等
	fmt.Printf("[AUDIT] %s %s %s - Status: %d, Latency: %v, IP: %s\n",
		event.Timestamp.Format("2006-01-02 15:04:05"),
		event.Method,
		event.Path,
		event.Status,
		event.Latency,
		event.IP,
	)
}
