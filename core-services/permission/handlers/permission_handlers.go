package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/permission"
)

// PermissionHandlers 权限处理�?
type PermissionHandlers struct {
	service permission.PermissionService
	logger  *zap.Logger
}

// NewPermissionHandlers 创建权限处理�?
func NewPermissionHandlers(service permission.PermissionService, logger *zap.Logger) *PermissionHandlers {
	return &PermissionHandlers{
		service: service,
		logger:  logger,
	}
}

// CheckPermission 检查权�?
func (h *PermissionHandlers) CheckPermission(c *gin.Context) {
	var req permission.PermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.UserID == "" || req.Resource == "" || req.Action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserID, Resource, and Action are required"})
		return
	}

	result, err := h.service.CheckPermission(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to check permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CheckPermissions 批量检查权�?
func (h *PermissionHandlers) CheckPermissions(c *gin.Context) {
	var req permission.BatchPermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证请求
	if len(req.Checks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one permission check is required"})
		return
	}

	results, err := h.service.CheckPermissions(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to check permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// CreateRole 创建角色
func (h *PermissionHandlers) CreateRole(c *gin.Context) {
	var req permission.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and Code are required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "role", "create") {
		return
	}

	result, err := h.service.CreateRole(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetRole 获取角色
func (h *PermissionHandlers) GetRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "role", "read") {
		return
	}

	req := &permission.GetRoleRequest{RoleID: roleID}
	result, err := h.service.GetRole(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get role", zap.String("role_id", roleID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateRole 更新角色
func (h *PermissionHandlers) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	var req permission.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	req.RoleID = roleID

	// 检查权�?
	if !h.checkPermission(c, "role", "update") {
		return
	}

	result, err := h.service.UpdateRole(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to update role", zap.String("role_id", roleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteRole 删除角色
func (h *PermissionHandlers) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "role", "delete") {
		return
	}

	req := &permission.DeleteRoleRequest{RoleID: roleID}
	err := h.service.DeleteRole(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to delete role", zap.String("role_id", roleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// ListRoles 列出角色
func (h *PermissionHandlers) ListRoles(c *gin.Context) {
	// 检查权�?
	if !h.checkPermission(c, "role", "list") {
		return
	}

	// 解析查询参数
	filter, err := h.parseRoleFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters: " + err.Error()})
		return
	}

	req := &permission.ListRolesRequest{Filter: filter}
	result, err := h.service.ListRoles(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreatePermission 创建权限
func (h *PermissionHandlers) CreatePermission(c *gin.Context) {
	var req permission.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Code == "" || req.Resource == "" || req.Action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, Code, Resource, and Action are required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "permission", "create") {
		return
	}

	result, err := h.service.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPermission 获取权限
func (h *PermissionHandlers) GetPermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission ID is required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "permission", "read") {
		return
	}

	req := &permission.GetPermissionRequest{PermissionID: permissionID}
	result, err := h.service.GetPermission(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get permission", zap.String("permission_id", permissionID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdatePermission 更新权限
func (h *PermissionHandlers) UpdatePermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission ID is required"})
		return
	}

	var req permission.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	req.PermissionID = permissionID

	// 检查权�?
	if !h.checkPermission(c, "permission", "update") {
		return
	}

	result, err := h.service.UpdatePermission(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to update permission", zap.String("permission_id", permissionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeletePermission 删除权限
func (h *PermissionHandlers) DeletePermission(c *gin.Context) {
	permissionID := c.Param("id")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Permission ID is required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "permission", "delete") {
		return
	}

	req := &permission.DeletePermissionRequest{PermissionID: permissionID}
	err := h.service.DeletePermission(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to delete permission", zap.String("permission_id", permissionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

// ListPermissions 列出权限
func (h *PermissionHandlers) ListPermissions(c *gin.Context) {
	// 检查权�?
	if !h.checkPermission(c, "permission", "list") {
		return
	}

	// 解析查询参数
	filter, err := h.parsePermissionFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters: " + err.Error()})
		return
	}

	req := &permission.ListPermissionsRequest{Filter: filter}
	result, err := h.service.ListPermissions(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list permissions"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AssignPermissionToRole 分配权限给角�?
func (h *PermissionHandlers) AssignPermissionToRole(c *gin.Context) {
	var req permission.AssignPermissionToRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.RoleID == "" || req.PermissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RoleID and PermissionID are required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "role_permission", "assign") {
		return
	}

	err := h.service.AssignPermissionToRole(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to assign permission to role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permission to role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission assigned to role successfully"})
}

// RevokePermissionFromRole 从角色撤销权限
func (h *PermissionHandlers) RevokePermissionFromRole(c *gin.Context) {
	var req permission.RevokePermissionFromRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.RoleID == "" || req.PermissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "RoleID and PermissionID are required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "role_permission", "revoke") {
		return
	}

	err := h.service.RevokePermissionFromRole(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to revoke permission from role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke permission from role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission revoked from role successfully"})
}

// GetRolePermissions 获取角色权限
func (h *PermissionHandlers) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "role_permission", "read") {
		return
	}

	req := &permission.GetRolePermissionsRequest{RoleID: roleID}
	result, err := h.service.GetRolePermissions(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get role permissions", zap.String("role_id", roleID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AssignRoleToUser 分配角色给用�?
func (h *PermissionHandlers) AssignRoleToUser(c *gin.Context) {
	var req permission.AssignRoleToUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.UserID == "" || req.RoleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserID and RoleID are required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "user_role", "assign") {
		return
	}

	err := h.service.AssignRoleToUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to assign role to user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role to user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned to user successfully"})
}

// RevokeRoleFromUser 从用户撤销角色
func (h *PermissionHandlers) RevokeRoleFromUser(c *gin.Context) {
	var req permission.RevokeRoleFromUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 验证必填字段
	if req.UserID == "" || req.RoleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserID and RoleID are required"})
		return
	}

	// 检查权�?
	if !h.checkPermission(c, "user_role", "revoke") {
		return
	}

	err := h.service.RevokeRoleFromUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to revoke role from user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke role from user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role revoked from user successfully"})
}

// GetUserRoles 获取用户角色
func (h *PermissionHandlers) GetUserRoles(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	tenantID := c.Query("tenant_id")

	// 检查权�?
	if !h.checkPermission(c, "user_role", "read") {
		return
	}

	req := &permission.GetUserRolesRequest{
		UserID:   userID,
		TenantID: tenantID,
	}
	result, err := h.service.GetUserRoles(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetUserPermissions 获取用户权限
func (h *PermissionHandlers) GetUserPermissions(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	tenantID := c.Query("tenant_id")

	// 检查权�?
	if !h.checkPermission(c, "user_permission", "read") {
		return
	}

	req := &permission.GetUserPermissionsRequest{
		UserID:   userID,
		TenantID: tenantID,
	}
	result, err := h.service.GetUserPermissions(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get user permissions", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HealthCheck 健康检�?
func (h *PermissionHandlers) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// 解析角色过滤�?
func (h *PermissionHandlers) parseRoleFilter(c *gin.Context) (*permission.RoleFilter, error) {
	filter := &permission.RoleFilter{
		TenantID: c.Query("tenant_id"),
		Search:   c.Query("search"),
		Pagination: permission.PaginationRequest{
			Page:     1,
			PageSize: 20,
		},
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Pagination.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filter.Pagination.PageSize = pageSize
		}
	}

	// 解析角色类型
	if typeStr := c.Query("type"); typeStr != "" {
		roleType := permission.RoleType(typeStr)
		filter.Type = &roleType
	}

	// 解析是否激�?
	if activeStr := c.Query("is_active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filter.IsActive = &active
		}
	}

	// 解析是否系统角色
	if systemStr := c.Query("is_system"); systemStr != "" {
		if system, err := strconv.ParseBool(systemStr); err == nil {
			filter.IsSystem = &system
		}
	}

	// 解析父角色ID
	if parentID := c.Query("parent_id"); parentID != "" {
		filter.ParentID = &parentID
	}

	return filter, nil
}

// 解析权限过滤�?
func (h *PermissionHandlers) parsePermissionFilter(c *gin.Context) (*permission.PermissionFilter, error) {
	filter := &permission.PermissionFilter{
		TenantID: c.Query("tenant_id"),
		Category: c.Query("category"),
		Resource: c.Query("resource"),
		Action:   c.Query("action"),
		Search:   c.Query("search"),
		Pagination: permission.PaginationRequest{
			Page:     1,
			PageSize: 20,
		},
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Pagination.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			filter.Pagination.PageSize = pageSize
		}
	}

	// 解析权限效果
	if effectStr := c.Query("effect"); effectStr != "" {
		effect := permission.PermissionEffect(effectStr)
		filter.Effect = &effect
	}

	return filter, nil
}

// 检查权限的辅助函数
func (h *PermissionHandlers) checkPermission(c *gin.Context, resource, action string) bool {
	// 这里应该从上下文中获取用户信息并检查权�?
	// 为了简化，这里假设权限检查通过
	// 在实际应用中，应该使用权限中间件或在这里调用权限服务
	
	// 从上下文获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return false
	}

	tenantID, _ := c.Get("tenant_id")

	// 构建权限检查请�?
	checkReq := &permission.PermissionCheckRequest{
		UserID:   userID.(string),
		TenantID: tenantID.(string),
		Resource: resource,
		Action:   action,
	}

	// 执行权限检�?
	result, err := h.service.CheckPermission(c.Request.Context(), checkReq)
	if err != nil {
		h.logger.Error("Permission check failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
		return false
	}

	if !result.Allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied: " + result.Reason})
		return false
	}

	return true
}
