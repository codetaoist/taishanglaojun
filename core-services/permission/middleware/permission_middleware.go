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

// PermissionMiddleware ?
type PermissionMiddleware struct {
	service permission.PermissionService
	logger  *zap.Logger
	config  PermissionMiddlewareConfig
}

// PermissionMiddlewareConfig ?
type PermissionMiddlewareConfig struct {
	// 
	Enabled          bool     `json:"enabled"`
	SkipPaths        []string `json:"skip_paths"`
	SkipMethods      []string `json:"skip_methods"`
	
	// ?
	RequireAuth      bool   `json:"require_auth"`
	DefaultResource  string `json:"default_resource"`
	DefaultAction    string `json:"default_action"`
	
	// 
	UserIDHeader     string `json:"user_id_header"`
	UserIDClaim      string `json:"user_id_claim"`
	TenantIDHeader   string `json:"tenant_id_header"`
	TenantIDClaim    string `json:"tenant_id_claim"`
	
	// 
	UnauthorizedCode int    `json:"unauthorized_code"`
	ForbiddenCode    int    `json:"forbidden_code"`
	ErrorMessage     string `json:"error_message"`
	
	// 
	CacheEnabled     bool          `json:"cache_enabled"`
	CacheTTL         time.Duration `json:"cache_ttl"`
	Timeout          time.Duration `json:"timeout"`
	
	// 
	EnableAuditLog   bool   `json:"enable_audit_log"`
	AuditLogLevel    string `json:"audit_log_level"`
	
	// 
	DebugMode        bool `json:"debug_mode"`
	LogPermissions   bool `json:"log_permissions"`
}

// PermissionRequirement 
type PermissionRequirement struct {
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	AllowGuest  bool                   `json:"allow_guest,omitempty"`
	RequireAll  bool                   `json:"require_all,omitempty"` // ?
}

// NewPermissionMiddleware ?
func NewPermissionMiddleware(service permission.PermissionService, logger *zap.Logger, config PermissionMiddlewareConfig) *PermissionMiddleware {
	// 
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

// RequirePermission 
func (m *PermissionMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return m.RequirePermissions(&PermissionRequirement{
		Resource: resource,
		Action:   action,
	})
}

// RequirePermissions 
func (m *PermissionMiddleware) RequirePermissions(requirements ...*PermissionRequirement) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// ?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		// ?
		if m.config.RequireAuth && userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// ?
		for _, req := range requirements {
			if !m.checkPermission(c, userID, tenantID, req) {
				return // ?
			}
		}

		// ?
		c.Next()
	}
}

// RequireRole 
func (m *PermissionMiddleware) RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// ?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		if userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// ?
		if !m.checkRole(c, userID, tenantID, roleCode) {
			return // ?
		}

		// ?
		c.Next()
	}
}

// RequireAnyRole 
func (m *PermissionMiddleware) RequireAnyRole(roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// ?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		if userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// ?
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

		// ?
		c.Next()
	}
}

// RequireResourceOwner ?
func (m *PermissionMiddleware) RequireResourceOwner(resourceType string, resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.Enabled {
			c.Next()
			return
		}

		// ?
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 
		userID, tenantID, err := m.extractUserInfo(c)
		if err != nil {
			m.handleError(c, http.StatusUnauthorized, "Failed to extract user info: "+err.Error())
			return
		}

		if userID == "" {
			m.handleError(c, m.config.UnauthorizedCode, "Authentication required")
			return
		}

		// ID
		resourceID := c.Param(resourceIDParam)
		if resourceID == "" {
			resourceID = c.Query(resourceIDParam)
		}

		if resourceID == "" {
			m.handleError(c, http.StatusBadRequest, "Resource ID is required")
			return
		}

		// 
		if !m.checkResourceOwnership(c, userID, tenantID, resourceType, resourceID) {
			return // ?
		}

		// ?
		c.Next()
	}
}

// ?
func (m *PermissionMiddleware) shouldSkip(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	// ?
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// ?
	for _, skipMethod := range m.config.SkipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// 
func (m *PermissionMiddleware) extractUserInfo(c *gin.Context) (userID, tenantID string, err error) {
	// Header
	userID = c.GetHeader(m.config.UserIDHeader)
	tenantID = c.GetHeader(m.config.TenantIDHeader)

	// JWT Claims
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

	// Context
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

// ?
func (m *PermissionMiddleware) checkPermission(c *gin.Context, userID, tenantID string, req *PermissionRequirement) bool {
	// 
	if req.AllowGuest && userID == "" {
		return true
	}

	// ?
	checkReq := &permission.PermissionCheckRequest{
		UserID:   userID,
		TenantID: tenantID,
		Resource: req.Resource,
		Action:   req.Action,
		Context: map[string]interface{}{
			"request_path":   c.Request.URL.Path,
			"request_method": c.Request.Method,
			"client_ip":      c.ClientIP(),
			"user_agent":     c.GetHeader("User-Agent"),
		},
	}

	// ?
	if checkReq.Resource == "" {
		checkReq.Resource = m.config.DefaultResource
	}
	if checkReq.Action == "" {
		checkReq.Action = m.config.DefaultAction
	}

	// ?
	ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.Timeout)
	defer cancel()

	// ?
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

	// ?
	if m.config.EnableAuditLog {
		m.logPermissionCheck(userID, tenantID, checkReq, result)
	}

	// ?
	if !result.Allowed {
		m.handleError(c, m.config.ForbiddenCode, result.Reason)
		return false
	}

	// 洢?
	c.Set("permission_check_result", result)

	return true
}

// ?
func (m *PermissionMiddleware) checkRole(c *gin.Context, userID, tenantID, roleCode string) bool {
	// ?
	ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.Timeout)
	defer cancel()

	// 
	roles, err := m.service.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		m.logger.Error("Failed to get user roles", 
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		m.handleError(c, http.StatusInternalServerError, "Failed to check role")
		return false
	}

	// ?
	for _, role := range roles {
		if role.Code == roleCode && role.IsActive {
			return true
		}
	}

	m.handleError(c, m.config.ForbiddenCode, fmt.Sprintf("Requires role: %s", roleCode))
	return false
}

// 
func (m *PermissionMiddleware) checkResourceOwnership(c *gin.Context, userID, tenantID, resourceType, resourceID string) bool {
	// ?
	checkReq := &permission.PermissionCheckRequest{
		UserID:     userID,
		TenantID:   tenantID,
		Resource:   resourceType,
		Action:     "own",
		ResourceID: &resourceID,
		Context: map[string]interface{}{
			"request_path":   c.Request.URL.Path,
			"request_method": c.Request.Method,
			"resource_type":  resourceType,
			"resource_id":    resourceID,
		},
	}

	// ?
	ctx, cancel := context.WithTimeout(c.Request.Context(), m.config.Timeout)
	defer cancel()

	// ?
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

	// ?
	if !result.Allowed {
		m.handleError(c, m.config.ForbiddenCode, "Resource access denied")
		return false
	}

	return true
}

// 
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

// ?
func (m *PermissionMiddleware) logPermissionCheck(userID, tenantID string, req *permission.PermissionCheckRequest, result *permission.PermissionCheckResponse) {
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

// GetUserPermissions ?
func GetUserPermissions(c *gin.Context) ([]*permission.Permission, bool) {
	if result, exists := c.Get("permission_check_result"); exists {
		if checkResult, ok := result.(*permission.PermissionCheckResponse); ok {
			return checkResult.Permissions, true
		}
	}
	return nil, false
}

// HasPermission 
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

// GetUserID ID?
func GetUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid, true
		}
	}
	return "", false
}

// GetTenantID ID?
func GetTenantID(c *gin.Context) (string, bool) {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			return tid, true
		}
	}
	return "", false
}

