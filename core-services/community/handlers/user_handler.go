package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
	logger      *zap.Logger
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *services.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// GetUserProfile 获取用户资料
// @Summary 获取用户资料
// @Description 根据用户ID获取用户详细资料
// @Tags 用户管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} models.UserProfileResponse "获取成功"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/{user_id} [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 获取当前用户ID（如果已登录）
	var viewerID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			viewerID = &uidStr
		}
	}

	profile, err := h.userService.GetUserProfile(userID, viewerID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, profile.ToResponse())
}

// GetMyProfile 获取当前用户资料
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的详细资料
// @Tags 用户管理
// @Produce json
// @Success 200 {object} models.UserProfileResponse "获取成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/me [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	userIDStr := userID.(string)
	profile, err := h.userService.GetUserProfile(userIDStr, &userIDStr)
	if err != nil {
		h.logger.Error("Failed to get my profile", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, profile.ToResponse())
}

// UpdateUserProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前用户的资料信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body models.UpdateUserProfileRequest true "更新用户资料请求"
// @Success 200 {object} models.UserProfileResponse "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/me [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var req models.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 更新用户资料
	profile, err := h.userService.UpdateUserProfile(userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to update user profile", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户资料失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile.ToResponse())
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持按状态筛选
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "用户状态"
// @Param sort query string false "排序方式" Enums(latest,active,popular)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	sort := c.DefaultQuery("sort", "latest")

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建请求参数
	req := &models.UserListRequest{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
		Status:   status,
		SortBy:   sort,
	}

	// 获取当前用户ID（如果已登录）
	var viewerID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			viewerID = &uidStr
		}
	}

	// 获取用户列表
	response, err := h.userService.GetUsers(req, viewerID)
	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response.Users,
		"pagination": gin.H{
			"page":        response.Page,
			"page_size":   response.PageSize,
			"total":       response.Total,
			"total_pages": response.TotalPages,
		},
	})
}

// GetUserStats 获取用户统计
// @Summary 获取用户统计
// @Description 获取用户相关统计信息
// @Tags 用户管理
// @Produce json
// @Success 200 {object} models.UserStatsResponse "获取成功"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		h.logger.Error("Failed to get user stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户
// @Tags 用户管理
// @Produce json
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "搜索成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.userService.SearchUsers(keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search users", zap.String("keyword", keyword), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response.Users,
		"pagination": gin.H{
			"page":        response.Page,
			"page_size":   response.PageSize,
			"total":       response.Total,
			"total_pages": response.TotalPages,
		},
		"keyword": keyword,
	})
}

// BanUser 封禁用户
// @Summary 封禁用户
// @Description 封禁指定用户（管理员操作）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body map[string]string true "封禁原因" example({"reason": "违规行为"})
// @Success 200 {object} map[string]interface{} "封禁成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/{user_id}/ban [post]
func (h *UserHandler) BanUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := h.userService.BanUser(userID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to ban user", zap.String("user_id", userID), zap.Error(err))
		if err.Error() == "用户不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "封禁用户失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "封禁成功"})
}

// UnbanUser 解封用户
// @Summary 解封用户
// @Description 解封指定用户（管理员操作）
// @Tags 用户管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} map[string]interface{} "解封成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/{user_id}/unban [post]
func (h *UserHandler) UnbanUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 获取管理员ID
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	err := h.userService.UnbanUser(userID, adminID.(string))
	if err != nil {
		h.logger.Error("Failed to unban user", zap.String("user_id", userID), zap.Error(err))
		if err.Error() == "用户不存在" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解封用户失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "解封成功"})
}

// GetUserPosts 获取用户帖子列表
// @Summary 获取用户帖子列表
// @Description 分页获取指定用户的帖子列表
// @Tags 用户管理
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/{user_id}/posts [get]
func (h *UserHandler) GetUserPosts(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 验证参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取查看者ID（可选）
	var viewerID *string
	if id, exists := c.Get("user_id"); exists {
		idStr := id.(string)
		viewerID = &idStr
	}

	// 获取用户帖子列表
	response, err := h.userService.GetUserPosts(userID, page, pageSize, viewerID)
	if err != nil {
		h.logger.Error("Failed to get user posts", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户帖子失败"})
		return
	}

	// 计算分页信息
	totalPages := (int(response.Total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"posts": response.Posts,
		"pagination": gin.H{
			"page":       response.Page,
			"page_size":  response.PageSize,
			"total":      response.Total,
			"total_pages": totalPages,
		},
	})
}

// UpdateUserActivity 更新用户活跃度
// @Summary 更新用户活跃度
// @Description 更新当前用户的活跃度信息
// @Tags 用户管理
// @Produce json
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 401 {object} map[string]interface{} "未授权"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /community/users/me/activity [post]
func (h *UserHandler) UpdateUserActivity(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	err := h.userService.UpdateUserActivity(userID.(string))
	if err != nil {
		h.logger.Error("Failed to update user activity", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新活跃度失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "活跃度更新成功"})
}