package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/service"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userRepo    repository.UserRepository
	userService service.AuthService
	logger      *zap.Logger
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userRepo repository.UserRepository, userService service.AuthService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		userService: userService,
		logger:      logger,
	}
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 解析查询参数
	query := &models.UserQuery{
		Username: c.Query("username"),
		Email:    c.Query("email"),
		Status:   models.UserStatus(c.Query("status")),
		Role:     models.UserRole(c.Query("role")),
		Search:   c.Query("search"),
		OrderBy:  c.DefaultQuery("order_by", "created_at"),
		Order:    c.DefaultQuery("order", "desc"),
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		} else {
			query.Page = 1
		}
	} else {
		query.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.PageSize = pageSize
		} else {
			query.PageSize = 20
		}
	} else {
		query.PageSize = 20
	}

	// 获取用户列表
	users, total, err := h.userRepo.List(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户列表失败",
		})
		return
	}

	// 转换为公开用户信息
	publicUsers := make([]*models.PublicUser, len(users))
	for i, user := range users {
		publicUsers[i] = user.ToPublic()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    publicUsers,
		"meta": gin.H{
			"total":     total,
			"page":      query.Page,
			"page_size": query.PageSize,
			"pages":     (total + int64(query.PageSize) - 1) / int64(query.PageSize),
		},
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID格式",
		})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
			return
		}
		h.logger.Error("Failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user.ToPublic(),
	})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 调用认证服务创建用户
	response, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建用户失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response.User,
		"message": "用户创建成功",
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID格式",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 调用认证服务更新用户
	user, err := h.userService.UpdateUserProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
			return
		}
		h.logger.Error("Failed to update user", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
		"message": "用户信息更新成功",
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID格式",
		})
		return
	}

	// 获取当前管理员信息
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	admin := adminUser.(*models.User)

	// 防止管理员删除自己
	if userID == admin.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不能删除自己的账户",
		})
		return
	}

	// 软删除用户
	err = h.userRepo.SoftDelete(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
			return
		}
		h.logger.Error("Failed to delete user", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户删除成功",
	})
}

// BatchDeleteUsers 批量删除用户
func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var req struct {
		UserIDs []string `json:"userIds" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 获取当前管理员信息
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	admin := adminUser.(*models.User)

	// 转换用户ID
	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		userID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "无效的用户ID格式: " + idStr,
			})
			return
		}

		// 防止管理员删除自己
		if userID == admin.ID {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "不能删除自己的账户",
			})
			return
		}

		userIDs = append(userIDs, userID)
	}

	// 批量更新用户状态为已删除
	err := h.userRepo.BatchUpdateStatus(c.Request.Context(), userIDs, models.UserStatusInactive)
	if err != nil {
		h.logger.Error("Failed to batch delete users", zap.Int("count", len(userIDs)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "批量删除用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"success": len(userIDs),
			"failed":  0,
		},
		"message": "批量删除用户成功",
	})
}

// UpdateUserStatus 更新用户状态
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID格式",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive suspended"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 获取当前管理员信息
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	admin := adminUser.(*models.User)

	// 防止管理员修改自己的状态
	if userID == admin.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不能修改自己的状态",
		})
		return
	}

	// 更新用户状态
	err = h.userRepo.UpdateStatus(c.Request.Context(), userID, models.UserStatus(req.Status))
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
			return
		}
		h.logger.Error("Failed to update user status", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户状态失败",
		})
		return
	}

	// 获取更新后的用户信息
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get updated user", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取更新后的用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user.ToPublic(),
		"message": "用户状态更新成功",
	})
}

// UpdateUserRole 更新用户角色
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的用户ID格式",
		})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=user moderator admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	// 获取当前管理员信息
	adminUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未授权",
		})
		return
	}

	admin := adminUser.(*models.User)

	// 防止管理员修改自己的角色
	if userID == admin.ID {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "不能修改自己的角色",
		})
		return
	}

	// 获取目标用户
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "用户不存在",
			})
			return
		}
		h.logger.Error("Failed to get user", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户信息失败",
		})
		return
	}

	// 更新用户角色
	user.Role = models.UserRole(req.Role)
	err = h.userRepo.Update(c.Request.Context(), user)
	if err != nil {
		h.logger.Error("Failed to update user role", zap.String("user_id", userID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户角色失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user.ToPublic(),
		"message": "用户角色更新成功",
	})
}

// GetUserStats 获取用户统计信息
func (h *UserHandler) GetUserStats(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取各种统计数据
	totalUsers, err := h.userRepo.Count(ctx)
	if err != nil {
		h.logger.Error("Failed to count total users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户统计失败",
		})
		return
	}

	activeUsers, err := h.userRepo.CountByStatus(ctx, models.UserStatusActive)
	if err != nil {
		h.logger.Error("Failed to count active users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户统计失败",
		})
		return
	}

	adminUsers, err := h.userRepo.CountByRole(ctx, models.RoleAdmin)
	if err != nil {
		h.logger.Error("Failed to count admin users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取用户统计失败",
		})
		return
	}

	// TODO: 实现今日新用户和在线用户统计
	newUsersToday := int64(0)
	onlineUsers := int64(0)

	stats := gin.H{
		"totalUsers":     totalUsers,
		"activeUsers":    activeUsers,
		"adminUsers":     adminUsers,
		"newUsersToday":  newUsersToday,
		"onlineUsers":    onlineUsers,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}