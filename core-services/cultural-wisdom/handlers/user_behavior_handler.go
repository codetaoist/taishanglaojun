package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
)

// UserBehaviorHandler 
type UserBehaviorHandler struct {
	userBehaviorService *services.UserBehaviorService
	logger              *zap.Logger
}

// NewUserBehaviorHandler 
func NewUserBehaviorHandler(userBehaviorService *services.UserBehaviorService, logger *zap.Logger) *UserBehaviorHandler {
	return &UserBehaviorHandler{
		userBehaviorService: userBehaviorService,
		logger:              logger,
	}
}

// RecordBehavior 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body services.BehaviorRequest true ""
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user-behavior/record [post]
func (h *UserBehaviorHandler) RecordBehavior(c *gin.Context) {
	var req services.BehaviorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "ID",
		})
		return
	}

	if req.WisdomID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "ID",
		})
		return
	}

	if req.ActionType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "",
		})
		return
	}

	// IPUser-Agent
	if req.IPAddress == "" {
		req.IPAddress = c.ClientIP()
	}
	if req.UserAgent == "" {
		req.UserAgent = c.GetHeader("User-Agent")
	}

	// 
	if err := h.userBehaviorService.RecordBehavior(c.Request.Context(), req); err != nil {
		h.logger.Error("Failed to record user behavior",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("wisdom_id", req.WisdomID),
			zap.String("action_type", req.ActionType))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RECORD_BEHAVIOR_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "",
	})
}

// GetUserProfile 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param user_id query string true "ID"
// @Success 200 {object} UserProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user-behavior/profile [get]
func (h *UserBehaviorHandler) GetUserProfile(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "ID",
		})
		return
	}

	// 
	profile, err := h.userBehaviorService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile",
			zap.Error(err),
			zap.String("user_id", userID))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_PROFILE_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UserProfileResponse{
		Code:    200,
		Message: "",
		Data:    profile,
	})
}

// GetSimilarUsers 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param user_id query string true "ID"
// @Param limit query int false "" default(10)
// @Success 200 {object} SimilarUsersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/user-behavior/similar-users [get]
func (h *UserBehaviorHandler) GetSimilarUsers(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETER",
			Message: "ID",
		})
		return
	}

	// 
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// 
	similarUsers, err := h.userBehaviorService.FindSimilarUsers(c.Request.Context(), userID, limit)
	if err != nil {
		h.logger.Error("Failed to find similar users",
			zap.Error(err),
			zap.String("user_id", userID))

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "FIND_SIMILAR_USERS_ERROR",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SimilarUsersResponse{
		Code:    200,
		Message: "",
		Data:    similarUsers,
		Total:   len(similarUsers),
	})
}

// 
type UserProfileResponse struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    *services.UserProfile `json:"data"`
}

type SimilarUsersResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
	Total   int      `json:"total"`
}

