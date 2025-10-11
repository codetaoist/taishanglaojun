package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// UserBehaviorHandler 用户行为处理�?
type UserBehaviorHandler struct {
	userBehaviorService *services.UserBehaviorService
	logger              *zap.Logger
}

// NewUserBehaviorHandler 创建用户行为处理器实�?
func NewUserBehaviorHandler(userBehaviorService *services.UserBehaviorService, logger *zap.Logger) *UserBehaviorHandler {
	return &UserBehaviorHandler{
		userBehaviorService: userBehaviorService,
		logger:              logger,
	}
}

// RecordBehavior 记录用户行为
// @Summary 记录用户行为
// @Description 记录用户的各种行为（浏览、点赞、分享等�?
// @Tags 用户行为
// @Accept json
// @Produce json
// @Param request body services.BehaviorRequest true "行为记录请求"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user-behavior/record [post]
func (h *UserBehaviorHandler) RecordBehavior(c *gin.Context) {
	var req services.BehaviorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 验证必需参数
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "用户ID不能为空",
		})
		return
	}

	if req.WisdomID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "智慧ID不能为空",
		})
		return
	}

	if req.ActionType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "行为类型不能为空",
		})
		return
	}

	// 从请求头获取IP和User-Agent
	if req.IPAddress == "" {
		req.IPAddress = c.ClientIP()
	}
	if req.UserAgent == "" {
		req.UserAgent = c.GetHeader("User-Agent")
	}

	// 记录行为
	if err := h.userBehaviorService.RecordBehavior(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to record user behavior",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("wisdom_id", req.WisdomID),
			zap.String("action_type", req.ActionType))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RECORD_BEHAVIOR_ERROR",
			Message: "记录用户行为失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "用户行为记录成功",
	})
}

// GetUserProfile 获取用户画像
// @Summary 获取用户画像
// @Description 获取用户的偏好画像和行为分析
// @Tags 用户行为
// @Accept json
// @Produce json
// @Param user_id query string true "用户ID"
// @Success 200 {object} UserProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user-behavior/profile [get]
func (h *UserBehaviorHandler) GetUserProfile(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "用户ID不能为空",
		})
		return
	}

	// 获取用户画像
	profile, err := h.userBehaviorService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile",
			zap.Error(err),
			zap.String("user_id", userID))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_PROFILE_ERROR",
			Message: "获取用户画像失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		Code:    200,
		Message: "获取用户画像成功",
		Data:    profile,
	})
}

// GetSimilarUsers 获取相似用户
// @Summary 获取相似用户
// @Description 获取与指定用户相似的其他用户
// @Tags 用户行为
// @Accept json
// @Produce json
// @Param user_id query string true "用户ID"
// @Param limit query int false "返回数量限制" default(10)
// @Success 200 {object} SimilarUsersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user-behavior/similar-users [get]
func (h *UserBehaviorHandler) GetSimilarUsers(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "用户ID不能为空",
		})
		return
	}

	// 解析限制参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// 获取相似用户
	similarUsers, err := h.userBehaviorService.FindSimilarUsers(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to find similar users",
			zap.Error(err),
			zap.String("user_id", userID))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "FIND_SIMILAR_USERS_ERROR",
			Message: "查找相似用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SimilarUsersResponse{
		Code:    200,
		Message: "获取相似用户成功",
		Data:    similarUsers,
		Total:   len(similarUsers),
	})
}

// 响应结构�?
type UserProfileResponse struct {
	Code    int                          `json:"code"`
	Message string                       `json:"message"`
	Data    *services.UserProfile        `json:"data"`
}

type SimilarUsersResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
	Total   int      `json:"total"`
}
