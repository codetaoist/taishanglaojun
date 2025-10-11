package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/security/services"
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// SecurityMiddleware 安全中间�?
type SecurityMiddleware struct {
	threatService *services.ThreatDetectionService
	auditService  *services.SecurityAuditService
}

// NewSecurityMiddleware 创建安全中间�?
func NewSecurityMiddleware(threatService *services.ThreatDetectionService, auditService *services.SecurityAuditService) *SecurityMiddleware {
	return &SecurityMiddleware{
		threatService: threatService,
		auditService:  auditService,
	}
}

// ThreatDetection 威胁检测中间件
func (m *SecurityMiddleware) ThreatDetection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建安全事件
		event := &models.SecurityEvent{
			EventType:   "http_request",
			Source:      getClientIP(c),
			Target:      c.Request.URL.Path,
			Description: fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			Severity:    "info",
			Status:      "detected",
			Metadata: models.JSONB{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"user_agent": c.Request.UserAgent(),
				"referer":    c.Request.Referer(),
			},
		}

		// 分析安全事件
		if m.threatService != nil {
			threat, err := m.threatService.AnalyzeSecurityEvent(context.Background(), event)
			if err == nil && threat != nil {
				// 如果检测到威胁，记录并可能阻止请求
				if threat.Severity == "critical" || threat.Severity == "high" {
					// 记录审计日志
					if m.auditService != nil {
						auditLog := &models.AuditLog{
							UserID:      getUserID(c),
							Action:      "threat_blocked",
							Resource:    c.Request.URL.Path,
							IPAddress:   getClientIP(c),
							UserAgent:   c.Request.UserAgent(),
							Status:      "blocked",
							Description: fmt.Sprintf("Threat detected: %s", threat.ThreatType),
							Metadata: models.JSONB{
								"threat_id":   threat.ID,
								"threat_type": threat.ThreatType,
								"severity":    threat.Severity,
							},
						}
						m.auditService.LogAuditEvent(context.Background(), auditLog)
					}

					c.JSON(http.StatusForbidden, gin.H{
						"error":   "Request blocked due to security threat",
						"code":    "THREAT_DETECTED",
						"details": "Your request has been identified as a potential security threat",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// RateLimiting 速率限制中间�?
func (m *SecurityMiddleware) RateLimiting(maxRequests int, window time.Duration) gin.HandlerFunc {
	// 简单的内存存储，生产环境应该使用Redis
	requestCounts := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		now := time.Now()

		// 清理过期的请求记�?
		if requests, exists := requestCounts[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < window {
					validRequests = append(validRequests, reqTime)
				}
			}
			requestCounts[clientIP] = validRequests
		}

		// 检查是否超过限�?
		if len(requestCounts[clientIP]) >= maxRequests {
			// 记录审计日志
			if m.auditService != nil {
				auditLog := &models.AuditLog{
					UserID:      getUserID(c),
					Action:      "rate_limit_exceeded",
					Resource:    c.Request.URL.Path,
					IPAddress:   clientIP,
					UserAgent:   c.Request.UserAgent(),
					Status:      "blocked",
					Description: fmt.Sprintf("Rate limit exceeded: %d requests in %v", maxRequests, window),
					Metadata: models.JSONB{
						"max_requests": maxRequests,
						"window":       window.String(),
						"current_count": len(requestCounts[clientIP]),
					},
				}
				m.auditService.LogAuditEvent(context.Background(), auditLog)
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"code":    "RATE_LIMIT_EXCEEDED",
				"details": fmt.Sprintf("Maximum %d requests per %v allowed", maxRequests, window),
			})
			c.Abort()
			return
		}

		// 记录当前请求
		requestCounts[clientIP] = append(requestCounts[clientIP], now)

		c.Next()
	}
}

// AuditLogging 审计日志中间�?
func (m *SecurityMiddleware) AuditLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录审计日志
		if m.auditService != nil {
			duration := time.Since(start)
			status := "success"
			if c.Writer.Status() >= 400 {
				status = "failed"
			}

			auditLog := &models.AuditLog{
				UserID:      getUserID(c),
				Action:      strings.ToLower(c.Request.Method),
				Resource:    c.Request.URL.Path,
				IPAddress:   getClientIP(c),
				UserAgent:   c.Request.UserAgent(),
				Status:      status,
				Description: fmt.Sprintf("%s %s - %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status()),
				Metadata: models.JSONB{
					"method":        c.Request.Method,
					"path":          c.Request.URL.Path,
					"status_code":   c.Writer.Status(),
					"response_time": duration.Milliseconds(),
					"content_length": c.Writer.Size(),
				},
			}

			go func() {
				m.auditService.LogAuditEvent(context.Background(), auditLog)
			}()
		}
	}
}

// SecurityHeaders 安全头中间件
func (m *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全�?
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// CORS 跨域资源共享中间�?
func (m *SecurityMiddleware) CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查是否允许该来源
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// InputValidation 输入验证中间�?
func (m *SecurityMiddleware) InputValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查常见的恶意输入模式
		maliciousPatterns := []string{
			"<script",
			"javascript:",
			"onload=",
			"onerror=",
			"eval(",
			"document.cookie",
			"../",
			"..\\",
			"union select",
			"drop table",
			"insert into",
			"delete from",
		}

		// 检查URL参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				lowerValue := strings.ToLower(value)
				for _, pattern := range maliciousPatterns {
					if strings.Contains(lowerValue, pattern) {
						// 记录安全事件
						if m.auditService != nil {
							auditLog := &models.AuditLog{
								UserID:      getUserID(c),
								Action:      "malicious_input_detected",
								Resource:    c.Request.URL.Path,
								IPAddress:   getClientIP(c),
								UserAgent:   c.Request.UserAgent(),
								Status:      "blocked",
								Description: fmt.Sprintf("Malicious input detected in parameter %s: %s", key, pattern),
								Metadata: models.JSONB{
									"parameter": key,
									"value":     value,
									"pattern":   pattern,
								},
							}
							m.auditService.LogAuditEvent(context.Background(), auditLog)
						}

						c.JSON(http.StatusBadRequest, gin.H{
							"error":   "Invalid input detected",
							"code":    "MALICIOUS_INPUT",
							"details": "Your input contains potentially malicious content",
						})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// IPWhitelist IP白名单中间件
func (m *SecurityMiddleware) IPWhitelist(allowedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		// 检查IP是否在白名单�?
		allowed := false
		for _, allowedIP := range allowedIPs {
			if allowedIP == clientIP {
				allowed = true
				break
			}
		}

		if !allowed {
			// 记录审计日志
			if m.auditService != nil {
				auditLog := &models.AuditLog{
					UserID:      getUserID(c),
					Action:      "ip_blocked",
					Resource:    c.Request.URL.Path,
					IPAddress:   clientIP,
					UserAgent:   c.Request.UserAgent(),
					Status:      "blocked",
					Description: fmt.Sprintf("IP %s not in whitelist", clientIP),
					Metadata: models.JSONB{
						"client_ip":    clientIP,
						"allowed_ips":  allowedIPs,
					},
				}
				m.auditService.LogAuditEvent(context.Background(), auditLog)
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"code":    "IP_NOT_ALLOWED",
				"details": "Your IP address is not allowed to access this resource",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBlacklist IP黑名单中间件
func (m *SecurityMiddleware) IPBlacklist(blockedIPs []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		
		// 检查IP是否在黑名单�?
		for _, blockedIP := range blockedIPs {
			if blockedIP == clientIP {
				// 记录审计日志
				if m.auditService != nil {
					auditLog := &models.AuditLog{
						UserID:      getUserID(c),
						Action:      "ip_blocked",
						Resource:    c.Request.URL.Path,
						IPAddress:   clientIP,
						UserAgent:   c.Request.UserAgent(),
						Status:      "blocked",
						Description: fmt.Sprintf("IP %s is blacklisted", clientIP),
						Metadata: models.JSONB{
							"client_ip":   clientIP,
							"blocked_ips": blockedIPs,
						},
					}
					m.auditService.LogAuditEvent(context.Background(), auditLog)
				}

				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Access denied",
					"code":    "IP_BLOCKED",
					"details": "Your IP address has been blocked",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// 辅助函数

// getClientIP 获取客户端IP地址
func getClientIP(c *gin.Context) string {
	// 检查X-Forwarded-For�?
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP�?
	if xri := c.Request.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := c.Request.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

// getUserID 获取用户ID
func getUserID(c *gin.Context) string {
	// 从上下文中获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}

	// 从JWT token中获取用户ID
	if token := c.Request.Header.Get("Authorization"); token != "" {
		// 这里应该解析JWT token获取用户ID
		// 简化处理，返回空字符串
		return ""
	}

	return ""
}
