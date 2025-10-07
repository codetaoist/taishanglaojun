package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"taishanglaojun/core-services/multi-tenancy/models"
	"taishanglaojun/core-services/multi-tenancy/services"
	"taishanglaojun/pkg/logger"
	"taishanglaojun/pkg/response"
)

// TenantContextKey 租户上下文键
type TenantContextKey string

const (
	// TenantIDKey 租户ID键
	TenantIDKey TenantContextKey = "tenant_id"
	// TenantKey 租户对象键
	TenantKey TenantContextKey = "tenant"
	// TenantUserKey 租户用户键
	TenantUserKey TenantContextKey = "tenant_user"
)

// TenantMiddleware 租户中间件配置
type TenantMiddleware struct {
	tenantService services.TenantServiceInterface
	logger        logger.Logger
}

// NewTenantMiddleware 创建租户中间件
func NewTenantMiddleware(tenantService services.TenantServiceInterface, logger logger.Logger) *TenantMiddleware {
	return &TenantMiddleware{
		tenantService: tenantService,
		logger:        logger,
	}
}

// TenantIdentificationMiddleware 租户识别中间件
// 从请求中识别租户（通过子域名、域名、Header等）
func (m *TenantMiddleware) TenantIdentificationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenant *models.Tenant
		var err error

		// 1. 从URL路径参数获取租户ID
		if tenantID := c.Param("tenant_id"); tenantID != "" {
			if id, parseErr := strconv.ParseUint(tenantID, 10, 32); parseErr == nil {
				tenant, err = m.tenantService.GetTenant(c.Request.Context(), uint(id))
			}
		}

		// 2. 从Header获取租户信息
		if tenant == nil {
			if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
				if id, parseErr := strconv.ParseUint(tenantID, 10, 32); parseErr == nil {
					tenant, err = m.tenantService.GetTenant(c.Request.Context(), uint(id))
				}
			}
		}

		// 3. 从子域名获取租户
		if tenant == nil {
			host := c.Request.Host
			if subdomain := extractSubdomain(host); subdomain != "" {
				tenant, err = m.tenantService.GetTenantBySubdomain(c.Request.Context(), subdomain)
			}
		}

		// 4. 从完整域名获取租户
		if tenant == nil {
			host := c.Request.Host
			tenant, err = m.tenantService.GetTenantByDomain(c.Request.Context(), host)
		}

		// 处理错误
		if err != nil && err != gorm.ErrRecordNotFound {
			m.logger.Error("Failed to identify tenant", "error", err, "host", c.Request.Host)
			response.Error(c, http.StatusInternalServerError, "租户识别失败", err)
			c.Abort()
			return
		}

		// 如果找到租户，检查租户状态
		if tenant != nil {
			if !tenant.IsActive() {
				response.Error(c, http.StatusForbidden, "租户已被暂停或停用", nil)
				c.Abort()
				return
			}

			// 将租户信息存储到上下文
			c.Set(string(TenantIDKey), tenant.ID)
			c.Set(string(TenantKey), tenant)

			// 设置租户上下文到请求上下文
			ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenant.ID)
			ctx = context.WithValue(ctx, TenantKey, tenant)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

// TenantRequiredMiddleware 租户必需中间件
// 确保请求必须有有效的租户
func (m *TenantMiddleware) TenantRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "缺少租户信息", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantContextMiddleware 租户上下文中间件
// 为请求设置租户上下文，用于数据隔离
func (m *TenantMiddleware) TenantContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先运行租户识别
		m.TenantIdentificationMiddleware()(c)
		if c.IsAborted() {
			return
		}

		// 确保有租户信息
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "缺少租户上下文", nil)
			c.Abort()
			return
		}

		tenantObj := tenant.(*models.Tenant)

		// 设置数据隔离上下文
		isolationCtx, err := m.tenantService.GetDataIsolationContext(c.Request.Context(), tenantObj.ID)
		if err != nil {
			m.logger.Error("Failed to get data isolation context", "error", err, "tenant_id", tenantObj.ID)
			response.Error(c, http.StatusInternalServerError, "数据隔离上下文设置失败", err)
			c.Abort()
			return
		}

		// 将隔离上下文添加到请求上下文
		c.Request = c.Request.WithContext(isolationCtx)

		c.Next()
	}
}

// TenantAccessMiddleware 租户访问权限中间件
// 检查用户是否有权限访问指定租户
func (m *TenantMiddleware) TenantAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID（假设从认证中间件设置）
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "用户未认证", nil)
			c.Abort()
			return
		}

		// 获取租户信息
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "缺少租户信息", nil)
			c.Abort()
			return
		}

		tenantObj := tenant.(*models.Tenant)
		userIDUint := userID.(uint)

		// 检查用户是否有权限访问该租户
		hasAccess, err := m.tenantService.ValidateUserAccess(c.Request.Context(), tenantObj.ID, userIDUint)
		if err != nil {
			m.logger.Error("Failed to validate user access", "error", err, "user_id", userIDUint, "tenant_id", tenantObj.ID)
			response.Error(c, http.StatusInternalServerError, "权限验证失败", err)
			c.Abort()
			return
		}

		if !hasAccess {
			response.Error(c, http.StatusForbidden, "无权限访问该租户", nil)
			c.Abort()
			return
		}

		// 获取用户在该租户中的角色信息
		tenantUser, err := m.tenantService.GetTenantUser(c.Request.Context(), tenantObj.ID, userIDUint)
		if err != nil {
			m.logger.Error("Failed to get tenant user", "error", err, "user_id", userIDUint, "tenant_id", tenantObj.ID)
			response.Error(c, http.StatusInternalServerError, "获取租户用户信息失败", err)
			c.Abort()
			return
		}

		// 将租户用户信息存储到上下文
		c.Set(string(TenantUserKey), tenantUser)

		c.Next()
	}
}

// TenantPermissionMiddleware 租户权限中间件
// 检查用户在租户中的权限级别
func TenantPermissionMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取租户用户信息
		tenantUser, exists := c.Get(string(TenantUserKey))
		if !exists || tenantUser == nil {
			response.Error(c, http.StatusForbidden, "缺少租户用户信息", nil)
			c.Abort()
			return
		}

		tenantUserObj := tenantUser.(*models.TenantUser)

		// 检查权限
		for _, permission := range requiredPermissions {
			if !hasPermission(tenantUserObj.Role, permission) {
				response.Error(c, http.StatusForbidden, "权限不足", nil)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// AdminPermissionMiddleware 管理员权限中间件
func AdminPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色（假设从认证中间件设置）
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "用户角色信息缺失", nil)
			c.Abort()
			return
		}

		// 检查是否为管理员
		if userRole != "admin" && userRole != "super_admin" {
			response.Error(c, http.StatusForbidden, "需要管理员权限", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantQuotaMiddleware 租户配额检查中间件
func (m *TenantMiddleware) TenantQuotaMiddleware(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取租户信息
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "缺少租户信息", nil)
			c.Abort()
			return
		}

		tenantObj := tenant.(*models.Tenant)

		// 检查配额
		canProceed, err := m.tenantService.CheckQuota(c.Request.Context(), tenantObj.ID, resourceType, 1)
		if err != nil {
			m.logger.Error("Failed to check quota", "error", err, "tenant_id", tenantObj.ID, "resource_type", resourceType)
			response.Error(c, http.StatusInternalServerError, "配额检查失败", err)
			c.Abort()
			return
		}

		if !canProceed {
			response.Error(c, http.StatusTooManyRequests, "已超出配额限制", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantRateLimitMiddleware 租户限流中间件
func (m *TenantMiddleware) TenantRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取租户信息
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			c.Next()
			return
		}

		tenantObj := tenant.(*models.Tenant)

		// TODO: 实现基于租户的限流逻辑
		// 可以根据租户的订阅级别设置不同的限流策略
		_ = tenantObj

		c.Next()
	}
}

// 辅助函数

// extractSubdomain 从主机名中提取子域名
func extractSubdomain(host string) string {
	// 移除端口号
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// 分割域名
	parts := strings.Split(host, ".")
	if len(parts) > 2 {
		return parts[0]
	}

	return ""
}

// hasPermission 检查角色是否有指定权限
func hasPermission(role, permission string) bool {
	// 定义角色权限映射
	rolePermissions := map[string][]string{
		"owner": {
			"tenant:read", "tenant:write", "tenant:delete",
			"user:read", "user:write", "user:delete",
			"settings:read", "settings:write",
			"quota:read", "quota:write",
			"stats:read",
		},
		"admin": {
			"tenant:read", "tenant:write",
			"user:read", "user:write",
			"settings:read", "settings:write",
			"quota:read",
			"stats:read",
		},
		"member": {
			"tenant:read",
			"user:read",
			"settings:read",
			"quota:read",
			"stats:read",
		},
		"viewer": {
			"tenant:read",
			"user:read",
			"settings:read",
			"quota:read",
		},
	}

	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// GetTenantFromContext 从上下文获取租户信息
func GetTenantFromContext(c *gin.Context) (*models.Tenant, bool) {
	tenant, exists := c.Get(string(TenantKey))
	if !exists || tenant == nil {
		return nil, false
	}
	return tenant.(*models.Tenant), true
}

// GetTenantIDFromContext 从上下文获取租户ID
func GetTenantIDFromContext(c *gin.Context) (uint, bool) {
	tenantID, exists := c.Get(string(TenantIDKey))
	if !exists || tenantID == nil {
		return 0, false
	}
	return tenantID.(uint), true
}

// GetTenantUserFromContext 从上下文获取租户用户信息
func GetTenantUserFromContext(c *gin.Context) (*models.TenantUser, bool) {
	tenantUser, exists := c.Get(string(TenantUserKey))
	if !exists || tenantUser == nil {
		return nil, false
	}
	return tenantUser.(*models.TenantUser), true
}