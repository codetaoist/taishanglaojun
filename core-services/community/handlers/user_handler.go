package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/codetaoist/taishanglaojun/core-services/community/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler 
type UserHandler struct {
	userService *services.UserService
	logger      *zap.Logger
}

// NewUserHandler 
func NewUserHandler(userService *services.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// GetUserProfile 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} models.UserProfileResponse ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id} [get]
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// ID
	var viewerID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			viewerID = &uidStr
		}
	}

	profile, err := h.userService.GetUserProfile(userID, viewerID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, profile.ToResponse())
}

// GetMyProfile 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} models.UserProfileResponse ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/me [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	userIDStr := userID.(string)
	profile, err := h.userService.GetUserProfile(userIDStr, &userIDStr)
	if err != nil {
		h.logger.Error("Failed to get my profile", zap.String("user_id", userIDStr), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, profile.ToResponse())
}

// UpdateUserProfile 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body models.UpdateUserProfileRequest true ""
// @Success 200 {object} models.UserProfileResponse ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/me [put]
func (h *UserHandler) UpdateUserProfile(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	var req models.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	profile, err := h.userService.UpdateUserProfile(userID.(string), &req)
	if err != nil {
		h.logger.Error("Failed to update user profile", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile.ToResponse())
}

// GetUsers 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Param status query string false "" Enums(active,inactive,all)
// @Param sort query string false "" Enums(latest,active,popular)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	// 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	status := c.Query("status")
	sort := c.DefaultQuery("sort", "latest")

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 
	req := &models.UserListRequest{
		Page:     page,
		PageSize: pageSize,
		Keyword:  keyword,
		Status:   status,
		SortBy:   sort,
	}

	// ID
	var viewerID *string
	if uid, exists := c.Get("user_id"); exists {
		if uidStr, ok := uid.(string); ok {
			viewerID = &uidStr
		}
	}

	// 
	response, err := h.userService.GetUsers(req, viewerID)
	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
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

// GetUserStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} models.UserStatsResponse ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		h.logger.Error("Failed to get user stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SearchUsers 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param q query string true ""
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	response, err := h.userService.SearchUsers(keyword, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to search users", zap.String("keyword", keyword), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
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

// BanUser 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param request body map[string]string true "" example({"reason": ""})
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id}/ban [post]
func (h *UserHandler) BanUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	err := h.userService.BanUser(userID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to ban user", zap.String("user_id", userID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// UnbanUser 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id}/unban [post]
func (h *UserHandler) UnbanUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// ID
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	err := h.userService.UnbanUser(userID, adminID.(string))
	if err != nil {
		h.logger.Error("Failed to unban user", zap.String("user_id", userID), zap.Error(err))
		if err.Error() == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

// GetUserPosts 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param user_id path string true "ID"
// @Param page query int false "" default(1)
// @Param page_size query int false "" default(20)
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/{user_id}/posts [get]
func (h *UserHandler) GetUserPosts(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 鿴ID
	var viewerID *string
	if id, exists := c.Get("user_id"); exists {
		idStr := id.(string)
		viewerID = &idStr
	}

	// 
	response, err := h.userService.GetUserPosts(userID, page, pageSize, viewerID)
	if err != nil {
		h.logger.Error("Failed to get user posts", zap.String("user_id", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	// 
	totalPages := (int(response.Total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"posts": response.Posts,
		"pagination": gin.H{
			"page":        response.Page,
			"page_size":   response.PageSize,
			"total":       response.Total,
			"total_pages": totalPages,
		},
	})
}

// UpdateUserActivity 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Failure 401 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /community/users/me/activity [post]
func (h *UserHandler) UpdateUserActivity(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ""})
		return
	}

	err := h.userService.UpdateUserActivity(userID.(string))
	if err != nil {
		h.logger.Error("Failed to update user activity", zap.String("user_id", userID.(string)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": ""})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": ""})
}

