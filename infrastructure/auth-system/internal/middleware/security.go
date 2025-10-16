package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableCSRF        bool     `json:"enable_csrf"`
	CSRFTokenLength   int      `json:"csrf_token_length"`
	TrustedProxies    []string `json:"trusted_proxies"`
	MaxRequestSize    int64    `json:"max_request_size"`
	EnableIPWhitelist bool     `json:"enable_ip_whitelist"`
	IPWhitelist       []string `json:"ip_whitelist"`
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		EnableCSRF:        false, // 默认关闭，可根据需要开启
		CSRFTokenLength:   32,
		TrustedProxies:    []string{"127.0.0.1", "::1"},
		MaxRequestSize:    10 << 20, // 10MB
		EnableIPWhitelist: false,
		IPWhitelist:       []string{},
	}
}

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	config *SecurityConfig
	logger *zap.Logger
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(config *SecurityConfig, logger *zap.Logger) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	
	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// RequestSizeLimit 请求大小限制中间件
func (s *SecurityMiddleware) RequestSizeLimit() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.ContentLength > s.config.MaxRequestSize {
			s.logger.Warn("Request size too large",
				zap.Int64("content_length", c.Request.ContentLength),
				zap.Int64("max_size", s.config.MaxRequestSize),
				zap.String("client_ip", c.ClientIP()),
			)
			
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "request_too_large",
				"message": "Request size exceeds maximum allowed size",
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// IPWhitelist IP白名单中间件
func (s *SecurityMiddleware) IPWhitelist() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !s.config.EnableIPWhitelist {
			c.Next()
			return
		}
		
		clientIP := c.ClientIP()
		allowed := false
		
		for _, ip := range s.config.IPWhitelist {
			if ip == clientIP {
				allowed = true
				break
			}
		}
		
		if !allowed {
			s.logger.Warn("IP not in whitelist",
				zap.String("client_ip", clientIP),
				zap.Strings("whitelist", s.config.IPWhitelist),
			)
			
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "ip_not_allowed",
				"message": "Your IP address is not allowed to access this resource",
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// CSRFProtection CSRF保护中间件
func (s *SecurityMiddleware) CSRFProtection() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if !s.config.EnableCSRF {
			c.Next()
			return
		}
		
		// 对于GET、HEAD、OPTIONS请求不需要CSRF保护
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		// 检查CSRF令牌
		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			token = c.PostForm("_csrf_token")
		}
		
		if token == "" {
			s.logger.Warn("Missing CSRF token",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "csrf_token_missing",
				"message": "CSRF token is required",
			})
			c.Abort()
			return
		}
		
		// 验证CSRF令牌（这里简化处理，实际应该验证令牌的有效性）
		if !s.validateCSRFToken(token) {
			s.logger.Warn("Invalid CSRF token",
				zap.String("token", token),
				zap.String("client_ip", c.ClientIP()),
			)
			
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "csrf_token_invalid",
				"message": "Invalid CSRF token",
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// GenerateCSRFToken 生成CSRF令牌
func (s *SecurityMiddleware) GenerateCSRFToken() (string, error) {
	bytes := make([]byte, s.config.CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// validateCSRFToken 验证CSRF令牌（简化实现）
func (s *SecurityMiddleware) validateCSRFToken(token string) bool {
	// 这里应该实现真正的令牌验证逻辑
	// 例如：检查令牌是否存在于会话中，是否过期等
	return len(token) >= 16 // 简化验证
}

// AntiClickjacking 防点击劫持中间件
func (s *SecurityMiddleware) AntiClickjacking() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置X-Frame-Options头
		c.Header("X-Frame-Options", "SAMEORIGIN")
		
		// 设置Content-Security-Policy的frame-ancestors指令
		existingCSP := c.GetHeader("Content-Security-Policy")
		if existingCSP != "" && !strings.Contains(existingCSP, "frame-ancestors") {
			c.Header("Content-Security-Policy", existingCSP+"; frame-ancestors 'self'")
		}
		
		c.Next()
	})
}

// RequestTimeout 请求超时中间件
func (s *SecurityMiddleware) RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 设置请求超时
		ctx := c.Request.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
}

// HidePoweredBy 隐藏服务器信息中间件
func (s *SecurityMiddleware) HidePoweredBy() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 移除可能泄露服务器信息的头
		c.Header("Server", "")
		c.Header("X-Powered-By", "")
		
		c.Next()
	})
}