package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/models"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"github.com/codetaoist/taishanglaojun/core-services/internal/response"
)

// TenantContextKey 租户上下文键
type TenantContextKey string

const (
	// TenantIDKey 租户ID键
	TenantIDKey TenantContextKey = "tenant_id"
	// TenantKey 租户键
	TenantKey TenantContextKey = "tenant"
	// TenantUserKey 租户用户键
	TenantUserKey TenantContextKey = "tenant_user"
)

// TenantMiddleware 租户中间件
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
// Header: X-Tenant-ID		
func (m *TenantMiddleware) TenantIdentificationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenant *models.Tenant
		var err error

		// 1. URL?ID
		if tenantID := c.Param("tenant_id"); tenantID != "" {
			if id, parseErr := strconv.ParseUint(tenantID, 10, 32); parseErr == nil {
				tenant, err = m.tenantService.GetTenant(c.Request.Context(), uint(id))
			}
		}

		// 2. Header: X-Tenant-ID	
		if tenant == nil {
			if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
				if id, parseErr := strconv.ParseUint(tenantID, 10, 32); parseErr == nil {
					tenant, err = m.tenantService.GetTenant(c.Request.Context(), uint(id))
				}
			}
		}

		// 3. Subdomain
		if tenant == nil {
			host := c.Request.Host
			if subdomain := extractSubdomain(host); subdomain != "" {
				tenant, err = m.tenantService.GetTenantBySubdomain(c.Request.Context(), subdomain)
			}
		}

		// 4. Domain
		if tenant == nil {
			host := c.Request.Host
			tenant, err = m.tenantService.GetTenantByDomain(c.Request.Context(), host)
		}

		// 5. ?	
		if err != nil && err != gorm.ErrRecordNotFound {
			m.logger.Error("Failed to identify tenant", "error", err, "host", c.Request.Host)
			response.Error(c, http.StatusInternalServerError, "", err)
			c.Abort()
			return
		}

		// ?
		if tenant != nil {
			if !tenant.IsActive() {
				response.Error(c, http.StatusForbidden, "?, nil)
				c.Abort()
				return
			}

			// 洢?
			c.Set(string(TenantIDKey), tenant.ID)
			c.Set(string(TenantKey), tenant)

			// ?
			ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenant.ID)
			ctx = context.WithValue(ctx, TenantKey, tenant)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}

// TenantRequiredMiddleware ?
// Header: X-Tenant-ID
func (m *TenantMiddleware) TenantRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantContextMiddleware 
// Header: X-Tenant-ID
func (m *TenantMiddleware) TenantContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ?
		m.TenantIdentificationMiddleware()(c)
		if c.IsAborted() {
			return
		}

		// ?
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "?, nil)
			c.Abort()
			return
		}

		tenantObj := tenant.(*models.Tenant)

		// ?
		isolationCtx, err := m.tenantService.GetDataIsolationContext(c.Request.Context(), tenantObj.ID)
		if err != nil {
			m.logger.Error("Failed to get data isolation context", "error", err, "tenant_id", tenantObj.ID)
			response.Error(c, http.StatusInternalServerError, "?, err)
			c.Abort()
			return
		}

		// 
		c.Request = c.Request.WithContext(isolationCtx)

		c.Next()
	}
}

// TenantAccessMiddleware ?
// Header: X-Tenant-ID
func (m *TenantMiddleware) TenantAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "?, nil)
			c.Abort()
			return
		}

		// 
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "", nil)
			c.Abort()
			return
		}

		tenantObj := tenant.(*models.Tenant)
		userIDUint := userID.(uint)

		// ?
		hasAccess, err := m.tenantService.ValidateUserAccess(c.Request.Context(), tenantObj.ID, userIDUint)
		if err != nil {
			m.logger.Error("Failed to validate user access", "error", err, "user_id", userIDUint, "tenant_id", tenantObj.ID)
			response.Error(c, http.StatusInternalServerError, "", err)
			c.Abort()
			return
		}

		if !hasAccess {
			response.Error(c, http.StatusForbidden, "", nil)
			c.Abort()
			return
		}

		// 
		tenantUser, err := m.tenantService.GetTenantUser(c.Request.Context(), tenantObj.ID, userIDUint)
		if err != nil {
			m.logger.Error("Failed to get tenant user", "error", err, "user_id", userIDUint, "tenant_id", tenantObj.ID)
			response.Error(c, http.StatusInternalServerError, "", err)
			c.Abort()
			return
		}

		// 洢?
		c.Set(string(TenantUserKey), tenantUser)

		c.Next()
	}
}

// TenantPermissionMiddleware ?
// Header: X-Tenant-ID
func TenantPermissionMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 
		tenantUser, exists := c.Get(string(TenantUserKey))
		if !exists || tenantUser == nil {
			response.Error(c, http.StatusForbidden, "", nil)
			c.Abort()
			return
		}

		tenantUserObj := tenantUser.(*models.TenantUser)

		// ?
		for _, permission := range requiredPermissions {
			if !hasPermission(tenantUserObj.Role, permission) {
				response.Error(c, http.StatusForbidden, "", nil)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// AdminPermissionMiddleware ?
// Header: X-Tenant-ID
func AdminPermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "", nil)
			c.Abort()
			return
		}

		// ?
		if userRole != "admin" && userRole != "super_admin" {
			response.Error(c, http.StatusForbidden, "", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantQuotaMiddleware ?
// Header: X-Tenant-ID
func (m *TenantMiddleware) TenantQuotaMiddleware(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			response.Error(c, http.StatusBadRequest, "", nil)
			c.Abort()
			return
		}

		tenantObj := tenant.(*models.Tenant)

		// ?
		canProceed, err := m.tenantService.CheckQuota(c.Request.Context(), tenantObj.ID, resourceType, 1)
		if err != nil {
			m.logger.Error("Failed to check quota", "error", err, "tenant_id", tenantObj.ID, "resource_type", resourceType)
			response.Error(c, http.StatusInternalServerError, "?, err)
			c.Abort()
			return
		}

		if !canProceed {
			response.Error(c, http.StatusTooManyRequests, "?, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantRateLimitMiddleware ?
func (m *TenantMiddleware) TenantRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 
		tenant, exists := c.Get(string(TenantKey))
		if !exists || tenant == nil {
			c.Next()
			return
		}

		tenantObj := tenant.(*models.Tenant)

		// TODO: 
		// 
		_ = tenantObj

		c.Next()
	}
}

// 

// extractSubdomain 
func extractSubdomain(host string) string {
	// ?
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// 
	parts := strings.Split(host, ".")
	if len(parts) > 2 {
		return parts[0]
	}

	return ""
}

// hasPermission 
func hasPermission(role, permission string) bool {
	// 
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

// GetTenantFromContext 
func GetTenantFromContext(c *gin.Context) (*models.Tenant, bool) {
	tenant, exists := c.Get(string(TenantKey))
	if !exists || tenant == nil {
		return nil, false
	}
	return tenant.(*models.Tenant), true
}

// GetTenantIDFromContext ID
func GetTenantIDFromContext(c *gin.Context) (uint, bool) {
	tenantID, exists := c.Get(string(TenantIDKey))
	if !exists || tenantID == nil {
		return 0, false
	}
	return tenantID.(uint), true
}

// GetTenantUserFromContext 
func GetTenantUserFromContext(c *gin.Context) (*models.TenantUser, bool) {
	tenantUser, exists := c.Get(string(TenantUserKey))
	if !exists || tenantUser == nil {
		return nil, false
	}
	return tenantUser.(*models.TenantUser), true
}

