package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS配置
type CORSConfig struct {
	// 允许的源
	AllowOrigins []string `yaml:"allow_origins"`
	
	// 允许的方法
	AllowMethods []string `yaml:"allow_methods"`
	
	// 允许的头部
	AllowHeaders []string `yaml:"allow_headers"`
	
	// 暴露的头部
	ExposeHeaders []string `yaml:"expose_headers"`
	
	// 是否允许凭证
	AllowCredentials bool `yaml:"allow_credentials"`
	
	// 预检请求缓存时间（秒）
	MaxAge int `yaml:"max_age"`
	
	// 是否允许所有源（开发模式）
	AllowAllOrigins bool `yaml:"allow_all_origins"`
	
	// 是否允许私有网络请求
	AllowPrivateNetwork bool `yaml:"allow_private_network"`
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cache-Control",
			"Connection",
			"Host",
			"Pragma",
			"Referer",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		},
		AllowCredentials: true,
		MaxAge:          86400, // 24小时
		AllowAllOrigins: false,
		AllowPrivateNetwork: false,
	}
}

// CORSMiddleware CORS中间件
type CORSMiddleware struct {
	config *CORSConfig
}

// NewCORSMiddleware 创建CORS中间件
func NewCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	if config == nil {
		config = DefaultCORSConfig()
	}
	
	return &CORSMiddleware{
		config: config,
	}
}

// Handler 返回Gin中间件处理函数
func (cm *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		// 设置CORS头部
		cm.setCORSHeaders(c, origin)
		
		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			cm.handlePreflightRequest(c, origin)
			return
		}
		
		c.Next()
	}
}

// setCORSHeaders 设置CORS头部
func (cm *CORSMiddleware) setCORSHeaders(c *gin.Context, origin string) {
	// Access-Control-Allow-Origin
	if cm.config.AllowAllOrigins {
		c.Header("Access-Control-Allow-Origin", "*")
	} else if cm.isOriginAllowed(origin) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Vary", "Origin")
	}
	
	// Access-Control-Allow-Credentials
	if cm.config.AllowCredentials && !cm.config.AllowAllOrigins {
		c.Header("Access-Control-Allow-Credentials", "true")
	}
	
	// Access-Control-Expose-Headers
	if len(cm.config.ExposeHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(cm.config.ExposeHeaders, ", "))
	}
	
	// Access-Control-Allow-Private-Network
	if cm.config.AllowPrivateNetwork {
		c.Header("Access-Control-Allow-Private-Network", "true")
	}
}

// handlePreflightRequest 处理预检请求
func (cm *CORSMiddleware) handlePreflightRequest(c *gin.Context, origin string) {
	// 检查源是否被允许
	if !cm.config.AllowAllOrigins && !cm.isOriginAllowed(origin) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	
	// Access-Control-Allow-Methods
	if len(cm.config.AllowMethods) > 0 {
		c.Header("Access-Control-Allow-Methods", strings.Join(cm.config.AllowMethods, ", "))
	}
	
	// Access-Control-Allow-Headers
	requestHeaders := c.GetHeader("Access-Control-Request-Headers")
	if requestHeaders != "" {
		if cm.areHeadersAllowed(requestHeaders) {
			c.Header("Access-Control-Allow-Headers", requestHeaders)
		} else if len(cm.config.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cm.config.AllowHeaders, ", "))
		}
	} else if len(cm.config.AllowHeaders) > 0 {
		c.Header("Access-Control-Allow-Headers", strings.Join(cm.config.AllowHeaders, ", "))
	}
	
	// Access-Control-Max-Age
	if cm.config.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", strconv.Itoa(cm.config.MaxAge))
	}
	
	c.AbortWithStatus(http.StatusNoContent)
}

// isOriginAllowed 检查源是否被允许
func (cm *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	
	for _, allowedOrigin := range cm.config.AllowOrigins {
		if cm.matchOrigin(origin, allowedOrigin) {
			return true
		}
	}
	
	return false
}

// matchOrigin 匹配源（支持通配符）
func (cm *CORSMiddleware) matchOrigin(origin, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	if pattern == origin {
		return true
	}
	
	// 支持子域名通配符，如 *.example.com
	if strings.HasPrefix(pattern, "*.") {
		domain := strings.TrimPrefix(pattern, "*.")
		return strings.HasSuffix(origin, "."+domain) || origin == domain
	}
	
	return false
}

// areHeadersAllowed 检查请求头是否被允许
func (cm *CORSMiddleware) areHeadersAllowed(requestHeaders string) bool {
	if len(cm.config.AllowHeaders) == 0 {
		return true
	}
	
	headers := strings.Split(requestHeaders, ",")
	for _, header := range headers {
		header = strings.TrimSpace(header)
		if !cm.isHeaderAllowed(header) {
			return false
		}
	}
	
	return true
}

// isHeaderAllowed 检查单个头部是否被允许
func (cm *CORSMiddleware) isHeaderAllowed(header string) bool {
	header = strings.ToLower(strings.TrimSpace(header))
	
	for _, allowedHeader := range cm.config.AllowHeaders {
		if strings.ToLower(allowedHeader) == header {
			return true
		}
	}
	
	return false
}

// SetConfig 设置配置
func (cm *CORSMiddleware) SetConfig(config *CORSConfig) {
	cm.config = config
}

// GetConfig 获取配置
func (cm *CORSMiddleware) GetConfig() *CORSConfig {
	return cm.config
}

// AddAllowedOrigin 添加允许的源
func (cm *CORSMiddleware) AddAllowedOrigin(origin string) {
	for _, existing := range cm.config.AllowOrigins {
		if existing == origin {
			return
		}
	}
	cm.config.AllowOrigins = append(cm.config.AllowOrigins, origin)
}

// RemoveAllowedOrigin 移除允许的源
func (cm *CORSMiddleware) RemoveAllowedOrigin(origin string) {
	for i, existing := range cm.config.AllowOrigins {
		if existing == origin {
			cm.config.AllowOrigins = append(cm.config.AllowOrigins[:i], cm.config.AllowOrigins[i+1:]...)
			return
		}
	}
}

// AddAllowedHeader 添加允许的头部
func (cm *CORSMiddleware) AddAllowedHeader(header string) {
	for _, existing := range cm.config.AllowHeaders {
		if strings.ToLower(existing) == strings.ToLower(header) {
			return
		}
	}
	cm.config.AllowHeaders = append(cm.config.AllowHeaders, header)
}

// RemoveAllowedHeader 移除允许的头部
func (cm *CORSMiddleware) RemoveAllowedHeader(header string) {
	for i, existing := range cm.config.AllowHeaders {
		if strings.ToLower(existing) == strings.ToLower(header) {
			cm.config.AllowHeaders = append(cm.config.AllowHeaders[:i], cm.config.AllowHeaders[i+1:]...)
			return
		}
	}
}

// AddExposedHeader 添加暴露的头部
func (cm *CORSMiddleware) AddExposedHeader(header string) {
	for _, existing := range cm.config.ExposeHeaders {
		if strings.ToLower(existing) == strings.ToLower(header) {
			return
		}
	}
	cm.config.ExposeHeaders = append(cm.config.ExposeHeaders, header)
}

// RemoveExposedHeader 移除暴露的头部
func (cm *CORSMiddleware) RemoveExposedHeader(header string) {
	for i, existing := range cm.config.ExposeHeaders {
		if strings.ToLower(existing) == strings.ToLower(header) {
			cm.config.ExposeHeaders = append(cm.config.ExposeHeaders[:i], cm.config.ExposeHeaders[i+1:]...)
			return
		}
	}
}

// EnableAllOrigins 启用所有源
func (cm *CORSMiddleware) EnableAllOrigins() {
	cm.config.AllowAllOrigins = true
	cm.config.AllowCredentials = false // 当允许所有源时，不能允许凭证
}

// DisableAllOrigins 禁用所有源
func (cm *CORSMiddleware) DisableAllOrigins() {
	cm.config.AllowAllOrigins = false
}

// EnableCredentials 启用凭证
func (cm *CORSMiddleware) EnableCredentials() {
	if !cm.config.AllowAllOrigins {
		cm.config.AllowCredentials = true
	}
}

// DisableCredentials 禁用凭证
func (cm *CORSMiddleware) DisableCredentials() {
	cm.config.AllowCredentials = false
}

// SetMaxAge 设置预检请求缓存时间
func (cm *CORSMiddleware) SetMaxAge(maxAge time.Duration) {
	cm.config.MaxAge = int(maxAge.Seconds())
}

// ValidateConfig 验证配置
func (cm *CORSMiddleware) ValidateConfig() error {
	if cm.config.AllowAllOrigins && cm.config.AllowCredentials {
		return fmt.Errorf("cannot allow credentials when allowing all origins")
	}
	
	if cm.config.MaxAge < 0 {
		return fmt.Errorf("max age cannot be negative")
	}
	
	return nil
}

// GetOriginFromRequest 从请求中获取源
func GetOriginFromRequest(c *gin.Context) string {
	return c.GetHeader("Origin")
}

// IsPreflightRequest 检查是否为预检请求
func IsPreflightRequest(c *gin.Context) bool {
	return c.Request.Method == http.MethodOptions &&
		c.GetHeader("Access-Control-Request-Method") != ""
}

// IsCORSRequest 检查是否为CORS请求
func IsCORSRequest(c *gin.Context) bool {
	return c.GetHeader("Origin") != ""
}