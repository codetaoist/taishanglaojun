package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	multitenant "github.com/codetaoist/taishanglaojun/core-services/multi-tenant"
	"go.uber.org/zap"
)

// TenantContextKey 租户上下文键
type TenantContextKey string

const (
	TenantIDKey      TenantContextKey = "tenant_id"
	TenantKey        TenantContextKey = "tenant"
	TenantContextKey TenantContextKey = "tenant_context"
)

// TenantMiddleware 租户中间�?
type TenantMiddleware struct {
	tenantService multitenant.TenantService
	logger        *zap.Logger
	config        TenantMiddlewareConfig
}

// TenantMiddlewareConfig 租户中间件配�?
type TenantMiddlewareConfig struct {
	// 租户识别策略
	IdentificationStrategy TenantIdentificationStrategy `json:"identification_strategy"`
	
	// 头部字段名称
	TenantHeaderName string `json:"tenant_header_name"`
	
	// 子域名模�?
	SubdomainPattern string `json:"subdomain_pattern"`
	
	// 路径前缀模式
	PathPrefixPattern string `json:"path_prefix_pattern"`
	
	// 查询参数名称
	QueryParamName string `json:"query_param_name"`
	
	// 默认租户ID
	DefaultTenantID string `json:"default_tenant_id"`
	
	// 是否允许无租户访�?
	AllowNoTenant bool `json:"allow_no_tenant"`
	
	// 跳过的路�?
	SkipPaths []string `json:"skip_paths"`
	
	// 错误处理
	ErrorHandler func(c *gin.Context, err error) `json:"-"`
}

// TenantIdentificationStrategy 租户识别策略
type TenantIdentificationStrategy string

const (
	StrategyHeader    TenantIdentificationStrategy = "header"
	StrategySubdomain TenantIdentificationStrategy = "subdomain"
	StrategyPath      TenantIdentificationStrategy = "path"
	StrategyQuery     TenantIdentificationStrategy = "query"
	StrategyMultiple  TenantIdentificationStrategy = "multiple"
)

// NewTenantMiddleware 创建租户中间�?
func NewTenantMiddleware(
	tenantService multitenant.TenantService,
	config TenantMiddlewareConfig,
	logger *zap.Logger,
) *TenantMiddleware {
	// 设置默认�?
	if config.TenantHeaderName == "" {
		config.TenantHeaderName = "X-Tenant-ID"
	}
	if config.QueryParamName == "" {
		config.QueryParamName = "tenant"
	}
	if config.PathPrefixPattern == "" {
		config.PathPrefixPattern = "/tenant/:tenant_id"
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}

	return &TenantMiddleware{
		tenantService: tenantService,
		config:        config,
		logger:        logger,
	}
}

// Handler 中间件处理器
func (m *TenantMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳�?
		if m.shouldSkip(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 识别租户
		tenantID, err := m.identifyTenant(c)
		if err != nil {
			m.logger.Error("Failed to identify tenant",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Error(err))
			
			if !m.config.AllowNoTenant {
				m.config.ErrorHandler(c, err)
				return
			}
		}

		// 如果没有找到租户ID且不允许无租户访�?
		if tenantID == "" && !m.config.AllowNoTenant {
			if m.config.DefaultTenantID != "" {
				tenantID = m.config.DefaultTenantID
			} else {
				m.config.ErrorHandler(c, &TenantError{
					Code:    "TENANT_NOT_FOUND",
					Message: "Tenant not found",
				})
				return
			}
		}

		// 如果有租户ID，获取租户信�?
		if tenantID != "" {
			tenant, err := m.tenantService.GetTenant(c.Request.Context(), tenantID)
			if err != nil {
				m.logger.Error("Failed to get tenant",
					zap.String("tenant_id", tenantID),
					zap.Error(err))
				
				m.config.ErrorHandler(c, &TenantError{
					Code:    "TENANT_NOT_FOUND",
					Message: "Tenant not found",
					Details: map[string]interface{}{"tenant_id": tenantID},
				})
				return
			}

			// 检查租户状�?
			if !tenant.IsActive() {
				m.config.ErrorHandler(c, &TenantError{
					Code:    "TENANT_INACTIVE",
					Message: "Tenant is not active",
					Details: map[string]interface{}{
						"tenant_id": tenantID,
						"status":    tenant.Status,
					},
				})
				return
			}

			// 获取租户上下�?
			userID := m.getUserID(c)
			tenantContext, err := m.tenantService.GetTenantContext(c.Request.Context(), tenantID, userID)
			if err != nil {
				m.logger.Error("Failed to get tenant context",
					zap.String("tenant_id", tenantID),
					zap.String("user_id", userID),
					zap.Error(err))
			}

			// 设置上下�?
			ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
			ctx = context.WithValue(ctx, TenantKey, tenant)
			if tenantContext != nil {
				ctx = context.WithValue(ctx, TenantContextKey, tenantContext)
			}

			c.Request = c.Request.WithContext(ctx)

			// 设置响应�?
			c.Header("X-Tenant-ID", tenantID)
			c.Header("X-Tenant-Name", tenant.Name)

			m.logger.Debug("Tenant identified",
				zap.String("tenant_id", tenantID),
				zap.String("tenant_name", tenant.Name),
				zap.String("user_id", userID),
				zap.String("path", c.Request.URL.Path))
		}

		c.Next()
	}
}

// identifyTenant 识别租户
func (m *TenantMiddleware) identifyTenant(c *gin.Context) (string, error) {
	switch m.config.IdentificationStrategy {
	case StrategyHeader:
		return m.identifyByHeader(c)
	case StrategySubdomain:
		return m.identifyBySubdomain(c)
	case StrategyPath:
		return m.identifyByPath(c)
	case StrategyQuery:
		return m.identifyByQuery(c)
	case StrategyMultiple:
		return m.identifyByMultiple(c)
	default:
		return "", &TenantError{
			Code:    "INVALID_STRATEGY",
			Message: "Invalid tenant identification strategy",
		}
	}
}

// identifyByHeader 通过头部识别租户
func (m *TenantMiddleware) identifyByHeader(c *gin.Context) (string, error) {
	tenantID := c.GetHeader(m.config.TenantHeaderName)
	if tenantID == "" {
		return "", &TenantError{
			Code:    "TENANT_HEADER_MISSING",
			Message: "Tenant header missing",
			Details: map[string]interface{}{"header": m.config.TenantHeaderName},
		}
	}
	return tenantID, nil
}

// identifyBySubdomain 通过子域名识别租�?
func (m *TenantMiddleware) identifyBySubdomain(c *gin.Context) (string, error) {
	host := c.Request.Host
	parts := strings.Split(host, ".")
	
	if len(parts) < 2 {
		return "", &TenantError{
			Code:    "INVALID_SUBDOMAIN",
			Message: "Invalid subdomain format",
			Details: map[string]interface{}{"host": host},
		}
	}

	// 假设第一部分是租户标�?
	tenantID := parts[0]
	
	// 验证租户ID格式
	if tenantID == "" || tenantID == "www" || tenantID == "api" {
		return "", &TenantError{
			Code:    "INVALID_TENANT_SUBDOMAIN",
			Message: "Invalid tenant subdomain",
			Details: map[string]interface{}{"subdomain": tenantID},
		}
	}

	return tenantID, nil
}

// identifyByPath 通过路径识别租户
func (m *TenantMiddleware) identifyByPath(c *gin.Context) (string, error) {
	path := c.Request.URL.Path
	
	// 简单的路径解析，假设格式为 /tenant/{tenant_id}/...
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	if len(parts) < 2 || parts[0] != "tenant" {
		return "", &TenantError{
			Code:    "TENANT_PATH_MISSING",
			Message: "Tenant path missing",
			Details: map[string]interface{}{"path": path},
		}
	}

	tenantID := parts[1]
	if tenantID == "" {
		return "", &TenantError{
			Code:    "EMPTY_TENANT_ID",
			Message: "Empty tenant ID in path",
		}
	}

	return tenantID, nil
}

// identifyByQuery 通过查询参数识别租户
func (m *TenantMiddleware) identifyByQuery(c *gin.Context) (string, error) {
	tenantID := c.Query(m.config.QueryParamName)
	if tenantID == "" {
		return "", &TenantError{
			Code:    "TENANT_QUERY_MISSING",
			Message: "Tenant query parameter missing",
			Details: map[string]interface{}{"param": m.config.QueryParamName},
		}
	}
	return tenantID, nil
}

// identifyByMultiple 通过多种方式识别租户
func (m *TenantMiddleware) identifyByMultiple(c *gin.Context) (string, error) {
	// 按优先级尝试不同的识别方�?
	
	// 1. 尝试头部
	if tenantID := c.GetHeader(m.config.TenantHeaderName); tenantID != "" {
		return tenantID, nil
	}

	// 2. 尝试查询参数
	if tenantID := c.Query(m.config.QueryParamName); tenantID != "" {
		return tenantID, nil
	}

	// 3. 尝试路径
	if tenantID, err := m.identifyByPath(c); err == nil && tenantID != "" {
		return tenantID, nil
	}

	// 4. 尝试子域�?
	if tenantID, err := m.identifyBySubdomain(c); err == nil && tenantID != "" {
		return tenantID, nil
	}

	return "", &TenantError{
		Code:    "TENANT_NOT_IDENTIFIED",
		Message: "Could not identify tenant using any method",
	}
}

// shouldSkip 检查是否应该跳�?
func (m *TenantMiddleware) shouldSkip(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getUserID 获取用户ID
func (m *TenantMiddleware) getUserID(c *gin.Context) string {
	// 从JWT token或session中获取用户ID
	// 这里简化实�?
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	
	// 从认证中间件设置的上下文中获�?
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	
	return ""
}

// TenantError 租户错误
type TenantError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *TenantError) Error() string {
	return e.Message
}

// defaultErrorHandler 默认错误处理�?
func defaultErrorHandler(c *gin.Context, err error) {
	if tenantErr, ok := err.(*TenantError); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    tenantErr.Code,
				"message": tenantErr.Message,
				"details": tenantErr.Details,
			},
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
		})
	}
	c.Abort()
}

// 辅助函数

// GetTenantID 从上下文获取租户ID
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// GetTenant 从上下文获取租户信息
func GetTenant(ctx context.Context) *multitenant.Tenant {
	if tenant, ok := ctx.Value(TenantKey).(*multitenant.Tenant); ok {
		return tenant
	}
	return nil
}

// GetTenantContext 从上下文获取租户上下�?
func GetTenantContext(ctx context.Context) *multitenant.TenantContext {
	if tenantContext, ok := ctx.Value(TenantContextKey).(*multitenant.TenantContext); ok {
		return tenantContext
	}
	return nil
}

// RequireTenant 要求租户中间�?
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantID(c.Request.Context())
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "TENANT_REQUIRED",
					"message": "Tenant is required for this operation",
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireActiveTenant 要求活跃租户中间�?
func RequireActiveTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := GetTenant(c.Request.Context())
		if tenant == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "TENANT_REQUIRED",
					"message": "Tenant is required for this operation",
				},
			})
			c.Abort()
			return
		}

		if !tenant.IsActive() {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "TENANT_INACTIVE",
					"message": "Tenant is not active",
					"details": gin.H{
						"tenant_id": tenant.ID,
						"status":    tenant.Status,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
