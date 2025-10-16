package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) *PermissionHandler {
	return &PermissionHandler{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

// Permission 权限模型
type Permission struct {
	ID          string    `json:"id" gorm:"primaryKey;type:char(36)"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Code        string    `json:"code" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Description string    `json:"description" gorm:"type:text"`
	Resource    string    `json:"resource" gorm:"type:varchar(255)"`
	Action      string    `json:"action" gorm:"type:varchar(255)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Role 角色模型
type Role struct {
	ID          string       `json:"id" gorm:"primaryKey;type:char(36)"`
	Name        string       `json:"name" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Code        string       `json:"code" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Description string       `json:"description" gorm:"type:text"`
	Type        string       `json:"type" gorm:"default:custom;type:varchar(50)"` // system, custom
	Level       int          `json:"level" gorm:"default:1"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	Status      string       `json:"status" gorm:"default:active;type:varchar(50)"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// UserRole 用户角色关联模型
type UserRole struct {
	ID        string    `json:"id" gorm:"primaryKey;type:char(36)"`
	UserID    string    `json:"user_id" gorm:"not null;type:char(36)"`
	RoleID    string    `json:"role_id" gorm:"not null;type:char(36)"`
	Role      Role      `json:"role" gorm:"foreignKey:RoleID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckPermissionRequest 权限检查请求
type CheckPermissionRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// CheckPermissionsRequest 批量权限检查请求
type CheckPermissionsRequest struct {
	UserID      string                   `json:"user_id" binding:"required"`
	Permissions []PermissionCheckRequest `json:"permissions" binding:"required"`
}

// PermissionCheckRequest 权限检查项
type PermissionCheckRequest struct {
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Level       int    `json:"level"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       int    `json:"level"`
	Status      string `json:"status"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

// AssignRolesRequest 分配角色请求
type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}

// CheckPermission 检查权限
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	var req CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户角色
	var userRoles []UserRole
	if err := h.db.Preload("Role.Permissions").Where("user_id = ?", req.UserID).Find(&userRoles).Error; err != nil {
		h.logger.Error("Failed to query user roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
		return
	}

	// 检查权限
	hasPermission := false
	for _, userRole := range userRoles {
		for _, permission := range userRole.Role.Permissions {
			if permission.Resource == req.Resource && permission.Action == req.Action {
				hasPermission = true
				break
			}
		}
		if hasPermission {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"has_permission": hasPermission,
		"user_id":        req.UserID,
		"resource":       req.Resource,
		"action":         req.Action,
	})
}

// CheckPermissions 批量检查权限
func (h *PermissionHandler) CheckPermissions(c *gin.Context) {
	var req CheckPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户角色
	var userRoles []UserRole
	if err := h.db.Preload("Role.Permissions").Where("user_id = ?", req.UserID).Find(&userRoles).Error; err != nil {
		h.logger.Error("Failed to query user roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions"})
		return
	}

	// 构建用户权限映射
	userPermissions := make(map[string]bool)
	for _, userRole := range userRoles {
		for _, permission := range userRole.Role.Permissions {
			key := permission.Resource + ":" + permission.Action
			userPermissions[key] = true
		}
	}

	// 检查每个权限
	results := make([]gin.H, len(req.Permissions))
	for i, perm := range req.Permissions {
		key := perm.Resource + ":" + perm.Action
		results[i] = gin.H{
			"resource":       perm.Resource,
			"action":         perm.Action,
			"has_permission": userPermissions[key],
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.UserID,
		"results": results,
	})
}

// ListPermissions 获取权限列表
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var permissions []Permission
	var total int64

	query := h.db.Model(&Permission{})
	if search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("Failed to count permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	// 获取权限列表
	if err := query.Offset(offset).Limit(limit).Find(&permissions).Error; err != nil {
		h.logger.Error("Failed to query permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"pages":       (total + int64(limit) - 1) / int64(limit),
	})
}

// CreatePermission 创建权限
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission := Permission{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		h.logger.Error("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Permission created successfully",
		"permission": permission,
	})
}

// GetPermission 获取权限详情
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var permission Permission
	if err := h.db.First(&permission, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
			return
		}
		h.logger.Error("Failed to get permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permission": permission})
}

// UpdatePermission 更新权限
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var permission Permission
	if err := h.db.First(&permission, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
			return
		}
		h.logger.Error("Failed to get permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permission"})
		return
	}

	// 更新字段
	if req.Name != "" {
		permission.Name = req.Name
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.Resource != "" {
		permission.Resource = req.Resource
	}
	if req.Action != "" {
		permission.Action = req.Action
	}

	if err := h.db.Save(&permission).Error; err != nil {
		h.logger.Error("Failed to update permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Permission updated successfully",
		"permission": permission,
	})
}

// DeletePermission 删除权限
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	if err := h.db.Delete(&Permission{}, "id = ?", id).Error; err != nil {
		h.logger.Error("Failed to delete permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

// ListRoles 获取角色列表
func (h *PermissionHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	roleType := c.Query("type")
	isActiveStr := c.Query("is_active")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var roles []Role
	var total int64

	query := h.db.Model(&Role{}).Preload("Permissions")
	
	// 搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	
	// 角色类型筛选
	if roleType != "" {
		query = query.Where("type = ?", roleType)
	}
	
	// 激活状态筛选
	if isActiveStr != "" {
		if isActiveStr == "true" {
			query = query.Where("is_active = ?", true)
		} else if isActiveStr == "false" {
			query = query.Where("is_active = ?", false)
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("Failed to count roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}

	// 获取角色列表
	if err := query.Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		h.logger.Error("Failed to query roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	})
}

// CreateRole 创建角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := Role{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Type:        "custom", // 默认为自定义角色
		Level:       req.Level,
		IsActive:    true,     // 默认激活
		Status:      "active",
	}

	if err := h.db.Create(&role).Error; err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role created successfully",
		"role":    role,
	})
}

// GetRole 获取角色详情
func (h *PermissionHandler) GetRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role Role
	if err := h.db.Preload("Permissions").First(&role, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": role})
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role Role
	if err := h.db.First(&role, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	// 更新字段
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Level > 0 {
		role.Level = req.Level
	}
	if req.Status != "" {
		role.Status = req.Status
	}

	if err := h.db.Save(&role).Error; err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Role updated successfully",
		"role":    role,
	})
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	// 检查是否有用户使用此角色
	var count int64
	if err := h.db.Model(&UserRole{}).Where("role_id = ?", id).Count(&count).Error; err != nil {
		h.logger.Error("Failed to check role usage", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete role that is assigned to users"})
		return
	}

	if err := h.db.Delete(&Role{}, "id = ?", id).Error; err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// GetRolePermissions 获取角色权限
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role Role
	if err := h.db.Preload("Permissions").First(&role, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role_id":     role.ID,
		"role_name":   role.Name,
		"permissions": role.Permissions,
	})
}

// AssignPermissionsToRole 为角色分配权限
func (h *PermissionHandler) AssignPermissionsToRole(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role Role
	if err := h.db.First(&role, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permissions"})
		return
	}

	// 获取权限
	var permissions []Permission
	if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
		h.logger.Error("Failed to get permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permissions"})
		return
	}

	// 替换角色权限
	if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		h.logger.Error("Failed to assign permissions to role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Permissions assigned successfully",
		"role_id":     role.ID,
		"permissions": permissions,
	})
}

// RemovePermissionFromRole 从角色移除权限
func (h *PermissionHandler) RemovePermissionFromRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	permissionID := c.Param("permissionId")
	if permissionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var role Role
	if err := h.db.First(&role, "id = ?", roleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		h.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission"})
		return
	}

	var permission Permission
	if err := h.db.First(&permission, "id = ?", permissionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
			return
		}
		h.logger.Error("Failed to get permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission"})
		return
	}

	// 移除权限
	if err := h.db.Model(&role).Association("Permissions").Delete(&permission); err != nil {
		h.logger.Error("Failed to remove permission from role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed successfully"})
}

// GetUserRoles 获取用户角色
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")

	var userRoles []UserRole
	if err := h.db.Preload("Role.Permissions").Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	roles := make([]Role, len(userRoles))
	for i, userRole := range userRoles {
		roles[i] = userRole.Role
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   roles,
	})
}

// AssignRolesToUser 为用户分配角色
func (h *PermissionHandler) AssignRolesToUser(c *gin.Context) {
	userID := c.Param("userId")

	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 删除现有角色
	if err := h.db.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
		h.logger.Error("Failed to remove existing user roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign roles"})
		return
	}

	// 分配新角色
	for _, roleID := range req.RoleIDs {
		userRole := UserRole{
			ID:     uuid.New().String(),
			UserID: userID,
			RoleID: roleID,
		}
		if err := h.db.Create(&userRole).Error; err != nil {
			h.logger.Error("Failed to assign role to user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign roles"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Roles assigned successfully",
		"user_id": userID,
		"role_ids": req.RoleIDs,
	})
}

// RemoveRoleFromUser 从用户移除角色
func (h *PermissionHandler) RemoveRoleFromUser(c *gin.Context) {
	userID := c.Param("userId")
	roleID := c.Param("roleId")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	if err := h.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&UserRole{}).Error; err != nil {
		h.logger.Error("Failed to remove role from user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}