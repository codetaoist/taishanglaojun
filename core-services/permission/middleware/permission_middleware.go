package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/permission"
)

// PermissionMiddleware 权限中间�?
type PermissionMiddleware struct {
	service permission.PermissionService
	logger  *zap.Logger
	config  PermissionMiddlewareConfig
}

// PermissionMiddlewareConfig 权限中间件配�?
type PermissionMiddlewareConfig struct {
	// 基本配置
	Enabled          bool     `json:"enabled"`
	SkipPaths        []string `json:"skip_paths"`
	SkipMethods      []string `json:"skip_methods"`
	
	// 权限检查配�?
	RequireAuth      bool   `json:"require_auth"`
	DefaultResource  string `json:"default_resource"`
	DefaultAction    string `json:"default_action"`
	
	// 用户信息提取配置
	UserIDHeader     string `json:"user_id_header"`
	UserIDClaim      string `json:"user_id_claim"`
	TenantIDHeader   string `json:"tenant_id_header"`
	TenantIDClaim    string `json:"tenant_id_claim"`
	
	// 错误处理配置
	UnauthorizedCode int    `json:"unauthorized_code"`
	ForbiddenCode    int    `json:"forbidden_code"`
	ErrorMessage     string `json:"error_message"`
	
	// 性能配置
	CacheEnabled     bool          `json:"cache_enabled"`
	CacheTTL         time.Duration `json:"cache_ttl"`
	Timeout          time.Duration `json:"timeout"`
	
	// 审计配置
	EnableAuditLog   bool   `json:"enable_audit_log"`
	AuditLogLevel    string `json:"audit_log_level"`
	
	// 调试配置
	DebugMode        bool `json:"debug_mode"`
	LogPermissions   bool `json:"log_permissions"`
}

// PermissionRequirement 权限要求
type PermissionRequirement struct {
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	AllowGuest  bool                   `json:"allow_guest,omitempty"`
	RequireAll  bool                   `json:"require_all,omitempty"` // 是否需要所有权�?
}

// NewPermissionMiddleware 创建权限中间�?
func NewPermissionMiddleware(service permission.PermissionService, logger *zap.Logger, config PermissionMiddlewareConfig) *PermissionMiddleware {
	// 设置默认配置
	if config.UserIDHeader == "" {
		config.UserIDHeader = "X-User-ID"
	}
	if config.UserIDClaim == "" {
		config.UserIDClaim = "user_id"
	}
	if config.TenantIDHeader == "" {
		config.TenantIDHeader = "X-Tenant-ID"
	}
	if config.TenantIDClaim == "" {
		config.TenantIDClaim = "tenant_id"
	}
	if config.UnauthorizedCode == 0 {
		config.UnauthorizedCode = http.StatusUnauthorized
	}
	if config.ForbiddenCode == 0 {
		config.ForbiddenCode = http.StatusForbidden
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = "Permission denied"
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.AuditLogLevel == "" {
		config.AuditLogLevel = "info"
	}

	return &PermissionMiddleware{
		service: service,
		logger:  logger,
		config:  config,
	}
}

// RequirePermission 要求特定权限的中间件
func (m *PermissionMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return m.RequirePermissions(&PermissionRequirement{
		Resource: resource,
		Action:   action,
	})
}

// RequirePermissions 要求多个权限的中间件
func (m *PermissionMiddleware) RequirePermissions(requirements ...*PermissionRequirement) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// 检查是否跳�?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 提取用户信息
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		// 检查是否需要认�?
		if m.config.RequireAuth && userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// 检查权�?
		for _, req := range requirements {
			if !m.checkPermission(c, userID, tenantID, req) {
				return // 权限检查失败，已处理错误响�?
			}
		}

		// 权限检查通过，继续处�?
		c.Next()
	}
}

// RequireRole 要求特定角色的中间件
func (m *PermissionMiddleware) RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// 检查是否跳�?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 提取用户信息
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		if userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// 检查角�?
		if !m.checkRole(c, userID, tenantID, roleCode) {
			return // 角色检查失败，已处理错误响�?
		}

		// 角色检查通过，继续处�?
		c.Next()
	}
}

// RequireAnyRole 要求任意角色的中间件
func (m *PermissionMiddleware) RequireAnyRole(roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// 检查是否跳�?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 提取用户信息
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		if userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// 检查是否拥有任意角�?
		hasRole := false
		for _, roleCode := range roleCodes {
			if m.checkRole(c, userID, tenantID, roleCode) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.handleError(c, m.config.ForbiddenCode, fmt.Sprintf("Requires one of roles: %s", strings.Join(roleCodes, ", ")))
			return
		}

		// 角色检查通过，继续处�?
		c.Next()
	}
}

// RequireResourceOwner 要求资源所有者权限的中间�?
func (m *PermissionMiddleware) RequireResourceOwner(resourceType string, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// 检查是否跳�?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 提取用户信息
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		if userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// 获取资源ID
		resourceID := c.Param(resourceIDParam)
		if resourceID == "" {
			resourceID = c.Query(resourceIDParam)
		}

		if resourceID == "" {
			m.handleError(c, http.StatusBadRequest, "Resource ID is required")
			return
		}

		// 检查资源所有权
		if !m.checkResourceOwnership(c, userID, tenantID, resourceType, resourceID) {
			return // 所有权检查失败，已处理错误响�?
		}

		// 所有权检查通过，继续处�?
		c.Next()
	}
}

// 检查是否应该跳过权限检�?
func (m *PermissionMiddleware) shouldSkip(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	// 检查跳过路�?
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// 检查跳过方�?
	for _, skipMethod := range m.config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// 提取用户信息
func (m *PermissionMiddleware) extractUserInfo(c *gin.Context) (userID, tenantID string, err error) {
	// 从Header提取
	userID = c.GetHeader(m.config.UserIDHeader)
	tenantID = c.GetHeader(m.config.TenantIDHeader)

	// 从JWT Claims提取（如果存在）
	if userID == "" {
		if claims, exists := c.Get("claims"); exists {
			if claimsMap, ok := claims.(map[string]interface{}); ok {
				if uid, exists := claimsMap[m.config.UserIDClaim]; exists {
					if uidStr, ok := uid.(string); ok {
						userID = uidStr
					}
				}
				if tid, exists := claimsMap[m.config.TenantIDClaim]; exists {
					if tidStr, ok := tid.(string); ok {
						tenantID = tidStr
					}
				}
			}
		}
	}

	// 从Context提取
	if userID == "" {
		if uid, exists := c.Get("user_id"); exists {
			if uidStr, ok := uid.(string); ok {
				userID = uidStr
			}
		}
	}
	if tenantID == "" {
		if tid, exists := c.Get("tenant_id"); exists {
			if tidStr, ok := tid.(string); ok {
				tenantID = tidStr
			}
		}
	}

	return userID, tenantID, nil
}

// 检查权�?
func (m *PermissionMiddleware) checkPermission(c *gin.Context, userID, tenantID string, req *PermissionRequirement) bool {
	// 如果允许访客且用户未认证，则通过
	if req.AllowGuest && userID == "" {
		return true
	}

	// 构建权限检查请�?
	checkReq := &permission.PermissionCheckRequest{
		UserID:     userID,
		TenantID:   tenantID,
		Resource:   req.Resource,
		Action:     req.Action,
		Conditions: req.Conditions,
		Context: map[string]interface{}{
			"request_path":   c.Request.URL.Path,
			"request_method": c.Request.Method,
			"client_ip":      c.ClientIP(),
			"user_agent":     c.GetHeader("User-Agent"),
		},
	}

	// 使用默认资源和动�?
	if checkReq.Resource == "" {
		checkReq.Resource = m.config.DefaultResource
	}
	if checkReq.Action == "" {
		checkReq.Action = m.config.DefaultAction
	}

	// 创建超时上下�?
	ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.Timeout)
	defer cancel()

	// 执行权限检�?
	result, err := m.service.CheckPermission(ctx, checkReq)
	if err != nil {
		m.logger.Error("Permission check failed", 
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.String("resource", checkReq.Resource),
			zap.String("action", checkReq.Action),
			zap.Error(err))
		m.handleError(c, http.StatusInternalServerError, "Permission check failed")
		return false
	}

	// 记录权限检查结�?
	if m.config.EnableAuditLog {
		m.logPermissionCheck(userID, tenantID, checkReq, result)
	}

	// 检查权限结�?
	if !result.Allowed {
		m.handleError(c, m.config.ForbiddenCode, result.Reason)
		return false
	}

	// 将权限检查结果存储到上下�?
	c.Set("permission_check_result", result)

	return true
}

// 检查角�?
func (m *PermissionMiddleware) checkRole(c *gin.Context, userID, tenantID, roleCode string) bool {
	// 创建超时上下�?
	ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.Timeout)
	defer cancel()

	// 获取用户角色
	roles, err := m.service.GetUserRoles(ctx, &permission.GetUserRolesRequest{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		m.logger.Error("Failed to get user roles", 
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		m.handleError(c, http.StatusInternalServerError, "Failed to check role")
		return false
	}

	// 检查是否拥有指定角�?
	for _, role := range roles.Roles {
		if role.Code == roleCode && role.IsActive {
			return true
		}
	}

	m.handleError(c, m.config.ForbiddenCode, fmt.Sprintf("Requires role: %s", roleCode))
	return false
}

// 检查资源所有权
func (m *PermissionMiddleware) checkResourceOwnership(c *gin.Context, userID, tenantID, resourceType, resourceID string) bool {
	// 构建资源权限检查请�?
	checkReq := &permission.PermissionCheckRequest{
		UserID:   userID,
		TenantID: tenantID,
		Resource: resourceType,
		Action:   "own",
		Conditions: map[string]interface{}{
			"resource_id": resourceID,
		},
		Context: map[string]interface{}{
			"request_path":   c.Request.URL.Path,
			"request_method": c.Request.Method,
			"resource_type":  resourceType,
			"resource_id":    resourceID,
		},
	}

	// 创建超时上下�?
	ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.Timeout)
	defer cancel()

	// 执行权限检�?
	result, err := m.service.CheckPermission(ctx, checkReq)
	if err != nil {
		m.logger.Error("Resource ownership check failed", 
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.String("resource_type", resourceType),
			zap.String("resource_id", resourceID),
			zap.Error(err))
		m.handleError(c, http.StatusInternalServerError, "Ownership check failed")
		return false
	}

	// 检查权限结�?
	if !result.Allowed {
		m.handleError(c, m.config.ForbiddenCode, "Resource access denied")
		return false
	}

	return true
}

// 处理错误响应
func (m *PermissionMiddleware) handleError(c *gin.Context, code int, message string) {
	if m.config.DebugMode {
		m.logger.Debug("Permission middleware error", 
			zap.Int("code", code),
			zap.String("message", message),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method))
	}

	c.JSON(code, gin.H{
		"error":   true,
		"code":    code,
		"message": message,
	})
	c.Abort()
}

// 记录权限检查日�?
func (m *PermissionMiddleware) logPermissionCheck(userID, tenantID string, req *permission.PermissionCheckRequest, result *permission.PermissionCheckResult) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action),
		zap.Bool("allowed", result.Allowed),
		zap.String("reason", result.Reason),
	}

	if m.config.LogPermissions {
		fields = append(fields, zap.Any("permissions", result.Permissions))
	}

	switch m.config.AuditLogLevel {
	case "debug":
		m.logger.Debug("Permission check", fields...)
	case "info":
		m.logger.Info("Permission check", fields...)
	case "warn":
		m.logger.Warn("Permission check", fields...)
	case "error":
		if !result.Allowed {
			m.logger.Error("Permission denied", fields...)
		}
	}
}

// GetUserPermissions 获取用户权限的辅助函�?
func GetUserPermissions(c *gin.Context) ([]*permission.Permission, bool) {
	if result, exists := c.Get("permission_check_result"); exists {
		if checkResult, ok := result.(*permission.PermissionCheckResult); ok {
			return checkResult.Permissions, true
		}
	}
	return nil, false
}

// HasPermission 检查用户是否拥有特定权限的辅助函数
func HasPermission(c *gin.Context, resource, action string) bool {
	permissions, exists := GetUserPermissions(c)
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return perm.Effect == permission.PermissionEffectAllow
		}
	}
	return false
}

// GetUserID 获取用户ID的辅助函�?
func GetUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid, true
		}
	}
	return "", false
}

// GetTenantID 获取租户ID的辅助函�?
func GetTenantID(c *gin.Context) (string, bool) {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			return tid, true
		}
	}
	return "", false
}
