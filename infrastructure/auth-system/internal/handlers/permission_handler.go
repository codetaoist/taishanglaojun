package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/service"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permissionService service.PermissionService
	logger            *zap.Logger
}

// NewPermissionHandler 创建权限处理器实例
func NewPermissionHandler(permissionService service.PermissionService, logger *zap.Logger) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		logger:            logger,
	}
}

// CreatePermission 创建权限
// @Summary 创建权限
// @Description 创建新的权限
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body service.CreatePermissionRequest true "创建权限请求"
// @Success 201 {object} models.Permission
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req service.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	permission, err := h.permissionService.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrPermissionExists:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "permission_exists",
				Message: "Permission already exists",
			})
		default:
			h.logger.Error("Failed to create permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to create permission",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// GetPermission 获取权限详情
// @Summary 获取权限详情
// @Description 根据ID获取权限详情
// @Tags permissions
// @Produce json
// @Param id path string true "权限ID"
// @Success 200 {object} models.Permission
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid permission ID format",
		})
		return
	}

	permission, err := h.permissionService.GetPermission(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrPermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "permission_not_found",
				Message: "Permission not found",
			})
		default:
			h.logger.Error("Failed to get permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to get permission",
			})
		}
		return
	}

	c.JSON(http.StatusOK, permission)
}

// ListPermissions 获取权限列表
// @Summary 获取权限列表
// @Description 获取权限列表，支持分页和过滤
// @Tags permissions
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param resource query string false "资源过滤"
// @Param action query string false "操作过滤"
// @Param search query string false "搜索关键词"
// @Success 200 {object} ListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	req := &service.ListPermissionsRequest{
		Page:     page,
		PageSize: pageSize,
		Resource: c.Query("resource"),
		Action:   c.Query("action"),
		Search:   c.Query("search"),
	}

	permissions, total, err := h.permissionService.ListPermissions(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list permissions",
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Data:     permissions,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// UpdatePermission 更新权限
// @Summary 更新权限
// @Description 更新权限信息
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "权限ID"
// @Param request body service.UpdatePermissionRequest true "更新权限请求"
// @Success 200 {object} models.Permission
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid permission ID format",
		})
		return
	}

	var req service.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	permission, err := h.permissionService.UpdatePermission(c.Request.Context(), id, &req)
	if err != nil {
		switch err {
		case service.ErrPermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "permission_not_found",
				Message: "Permission not found",
			})
		case service.ErrPermissionExists:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "permission_exists",
				Message: "Permission name already exists",
			})
		default:
			h.logger.Error("Failed to update permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to update permission",
			})
		}
		return
	}

	c.JSON(http.StatusOK, permission)
}

// DeletePermission 删除权限
// @Summary 删除权限
// @Description 删除权限及其相关的角色和用户权限
// @Tags permissions
// @Produce json
// @Param id path string true "权限ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid permission ID format",
		})
		return
	}

	err = h.permissionService.DeletePermission(c.Request.Context(), id)
	if err != nil {
		switch err {
		case service.ErrPermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "permission_not_found",
				Message: "Permission not found",
			})
		default:
			h.logger.Error("Failed to delete permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to delete permission",
			})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限
// @Description 获取指定角色的所有权限
// @Tags role-permissions
// @Produce json
// @Param role path string true "角色名称"
// @Success 200 {array} models.Permission
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /role-permissions/{role} [get]
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleStr := c.Param("role")
	role := models.UserRole(roleStr)

	// 验证角色是否有效
	if !isValidRole(role) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_role",
			Message: "Invalid role",
		})
		return
	}

	permissions, err := h.permissionService.GetRolePermissions(c.Request.Context(), role)
	if err != nil {
		h.logger.Error("Failed to get role permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get role permissions",
		})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// AssignRolePermissionRequest 分配角色权限请求
type AssignRolePermissionRequest struct {
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
}

// AssignRolePermission 分配角色权限
// @Summary 分配角色权限
// @Description 为角色分配权限
// @Tags role-permissions
// @Accept json
// @Produce json
// @Param role path string true "角色名称"
// @Param request body AssignRolePermissionRequest true "分配权限请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /role-permissions/{role} [post]
func (h *PermissionHandler) AssignRolePermission(c *gin.Context) {
	roleStr := c.Param("role")
	role := models.UserRole(roleStr)

	// 验证角色是否有效
	if !isValidRole(role) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_role",
			Message: "Invalid role",
		})
		return
	}

	var req AssignRolePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	err := h.permissionService.AssignRolePermission(c.Request.Context(), role, req.PermissionID)
	if err != nil {
		switch err {
		case service.ErrPermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "permission_not_found",
				Message: "Permission not found",
			})
		case service.ErrRolePermissionExists:
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "role_permission_exists",
				Message: "Role permission already exists",
			})
		default:
			h.logger.Error("Failed to assign role permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to assign role permission",
			})
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Role permission assigned successfully",
	})
}

// RevokeRolePermission 撤销角色权限
// @Summary 撤销角色权限
// @Description 撤销角色的指定权限
// @Tags role-permissions
// @Produce json
// @Param role path string true "角色名称"
// @Param permissionId path string true "权限ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /role-permissions/{role}/{permissionId} [delete]
func (h *PermissionHandler) RevokeRolePermission(c *gin.Context) {
	roleStr := c.Param("role")
	role := models.UserRole(roleStr)

	// 验证角色是否有效
	if !isValidRole(role) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_role",
			Message: "Invalid role",
		})
		return
	}

	permissionIDStr := c.Param("permissionId")
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_permission_id",
			Message: "Invalid permission ID format",
		})
		return
	}

	err = h.permissionService.RevokeRolePermission(c.Request.Context(), role, permissionID)
	if err != nil {
		switch err {
		case service.ErrRolePermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "role_permission_not_found",
				Message: "Role permission not found",
			})
		default:
			h.logger.Error("Failed to revoke role permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to revoke role permission",
			})
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Role permission revoked successfully",
	})
}

// GetUserPermissions 获取用户权限
// @Summary 获取用户权限
// @Description 获取用户的有效权限（角色权限 + 用户权限）
// @Tags user-permissions
// @Produce json
// @Param userId path string true "用户ID"
// @Success 200 {array} models.Permission
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /user-permissions/{userId} [get]
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	permissions, err := h.permissionService.GetUserEffectivePermissions(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get user permissions",
		})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

// AssignUserPermissionRequest 分配用户权限请求
type AssignUserPermissionRequest struct {
	PermissionID uuid.UUID `json:"permission_id" binding:"required"`
	Granted      bool      `json:"granted"`
}

// AssignUserPermission 分配用户权限
// @Summary 分配用户权限
// @Description 为用户分配或撤销权限
// @Tags user-permissions
// @Accept json
// @Produce json
// @Param userId path string true "用户ID"
// @Param request body AssignUserPermissionRequest true "分配权限请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /user-permissions/{userId} [post]
func (h *PermissionHandler) AssignUserPermission(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	var req AssignUserPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	err = h.permissionService.AssignUserPermission(c.Request.Context(), userID, req.PermissionID, req.Granted)
	if err != nil {
		switch err {
		case service.ErrPermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "permission_not_found",
				Message: "Permission not found",
			})
		default:
			h.logger.Error("Failed to assign user permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to assign user permission",
			})
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "User permission assigned successfully",
	})
}

// RevokeUserPermission 撤销用户权限
// @Summary 撤销用户权限
// @Description 撤销用户的指定权限
// @Tags user-permissions
// @Produce json
// @Param userId path string true "用户ID"
// @Param permissionId path string true "权限ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /user-permissions/{userId}/{permissionId} [delete]
func (h *PermissionHandler) RevokeUserPermission(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	permissionIDStr := c.Param("permissionId")
	permissionID, err := uuid.Parse(permissionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_permission_id",
			Message: "Invalid permission ID format",
		})
		return
	}

	err = h.permissionService.RevokeUserPermission(c.Request.Context(), userID, permissionID)
	if err != nil {
		switch err {
		case service.ErrUserPermissionNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_permission_not_found",
				Message: "User permission not found",
			})
		default:
			h.logger.Error("Failed to revoke user permission", zap.Error(err))
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to revoke user permission",
			})
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "User permission revoked successfully",
	})
}

// 辅助函数：验证角色是否有效
func isValidRole(role models.UserRole) bool {
	switch role {
	case models.RoleSuperAdmin, models.RoleAdmin, models.RoleModerator, models.RoleUser, models.RoleGuest:
		return true
	default:
		return false
	}
}

// ListResponse 列表响应结构体
type ListResponse struct {
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}