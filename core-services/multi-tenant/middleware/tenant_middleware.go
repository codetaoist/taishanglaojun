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

// TenantMiddleware 租户中间件
type TenantMiddleware struct {
	tenantService multitenant.TenantService
	logger        *zap.Logger
	config        TenantMiddlewareConfig
}

// TenantMiddlewareConfig 租户中间件配置
type TenantMiddlewareConfig struct {
	// 租户识别策略
	IdentificationStrategy TenantIdentificationStrategy `json:"identification_strategy"`
	
	// 租户头部名称
	TenantHeaderName string `json:"tenant_header_name"`
	
	// 子域名模式
	SubdomainPattern string `json:"subdomain_pattern"`
	
	// 路径前缀模式
	PathPrefixPattern string `json:"path_prefix_pattern"`
	
	// 查询参数名称
	QueryParamName string `json:"query_param_name"`
	
	// 默认租户ID
	DefaultTenantID string `json:"default_tenant_id"`
	
	// 是否允许无租户
	AllowNoTenant bool `json:"allow_no_tenant"`
	
	// 跳过的路径
	SkipPaths []string `json:"skip_paths"`
	
	// 错误处理器
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

// NewTenantMiddleware 创建新的租户中间件
func NewTenantMiddleware(
	tenantService multitenant.TenantService,
	config TenantMiddlewareConfig,
	logger *zap.Logger,
) *TenantMiddleware {
	// 设置默认配置
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

// Handler 获取中间件处理器
func (m *TenantMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过此路径
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

		// ID?
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

		// ID?
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

			// ?
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

			// ?
			userID := m.getUserID(c)
			tenantContext, err := m.tenantService.GetTenantContext(c.Request.Context(), tenantID, userID)
			if err != nil {
				m.logger.Error("Failed to get tenant context",
					zap.String("tenant_id", tenantID),
					zap.String("user_id", userID),
					zap.Error(err))
			}

			// ?
			ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
			ctx = context.WithValue(ctx, TenantKey, tenant)
			if tenantContext != nil {
				ctx = context.WithValue(ctx, TenantContextKey, tenantContext)
			}

			c.Request = c.Request.WithContext(ctx)

			// ?
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

// identifyTenant 
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

// identifyByHeader 
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

// identifyBySubdomain ?
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

	// ?
	tenantID := parts[0]
	
	// ID
	if tenantID == "" || tenantID == "www" || tenantID == "api" {
		return "", &TenantError{
			Code:    "INVALID_TENANT_SUBDOMAIN",
			Message: "Invalid tenant subdomain",
			Details: map[string]interface{}{"subdomain": tenantID},
		}
	}

	return tenantID, nil
}

// identifyByPath 
func (m *TenantMiddleware) identifyByPath(c *gin.Context) (string, error) {
	path := c.Request.URL.Path
	
	//  /tenant/{tenant_id}/...
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

// identifyByQuery 
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

// identifyByMultiple 
func (m *TenantMiddleware) identifyByMultiple(c *gin.Context) (string, error) {
	// ?
	
	// 1. 
	if tenantID := c.GetHeader(m.config.TenantHeaderName); tenantID != "" {
		return tenantID, nil
	}

	// 2. 
	if tenantID := c.Query(m.config.QueryParamName); tenantID != "" {
		return tenantID, nil
	}

	// 3. 
	if tenantID, err := m.identifyByPath(c); err == nil && tenantID != "" {
		return tenantID, nil
	}

	// 4. ?
	if tenantID, err := m.identifyBySubdomain(c); err == nil && tenantID != "" {
		return tenantID, nil
	}

	return "", &TenantError{
		Code:    "TENANT_NOT_IDENTIFIED",
		Message: "Could not identify tenant using any method",
	}
}

// shouldSkip ?
func (m *TenantMiddleware) shouldSkip(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// getUserID ID
func (m *TenantMiddleware) getUserID(c *gin.Context) string {
	// JWT tokensessionID
	// ?
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	
	// ?
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	
	return ""
}

// TenantError 
type TenantError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *TenantError) Error() string {
	return e.Message
}

// defaultErrorHandler ?
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

// 

// GetTenantID ID
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	return ""
}

// GetTenant 
func GetTenant(ctx context.Context) *multitenant.Tenant {
	if tenant, ok := ctx.Value(TenantKey).(*multitenant.Tenant); ok {
		return tenant
	}
	return nil
}

// GetTenantContext ?
func GetTenantContext(ctx context.Context) *multitenant.TenantContext {
	if tenantContext, ok := ctx.Value(TenantContextKey).(*multitenant.TenantContext); ok {
		return tenantContext
	}
	return nil
}

// RequireTenant ?
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

// RequireActiveTenant ?
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

