package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taishanglaojun/health-management/internal/application"
)

// HealthRecommendationHandler 健康建议处理器
type HealthRecommendationHandler struct {
	recommendationService *application.HealthRecommendationService
}

// NewHealthRecommendationHandler 创建健康建议处理器
func NewHealthRecommendationHandler(recommendationService *application.HealthRecommendationService) *HealthRecommendationHandler {
	return &HealthRecommendationHandler{
		recommendationService: recommendationService,
	}
}

// GenerateRecommendationsRequest 生成建议请求
type GenerateRecommendationsRequest struct {
	UserID uuid.UUID                                `json:"user_id" binding:"required"`
	Types  []application.RecommendationType         `json:"types,omitempty"`
	Days   int                                      `json:"days,omitempty"`
	Limit  int                                      `json:"limit,omitempty"`
}

// GenerateRecommendationsResponse 生成建议响应
type GenerateRecommendationsResponse struct {
	Recommendations []application.HealthRecommendation `json:"recommendations"`
	Summary         string                             `json:"summary"`
	Count           int                                `json:"count"`
	Timestamp       string                             `json:"timestamp"`
}

// GetPersonalizedTipsRequest 获取个性化提示请求
type GetPersonalizedTipsRequest struct {
	UserID   uuid.UUID                       `json:"user_id" binding:"required"`
	Category application.RecommendationType  `json:"category,omitempty"`
	Limit    int                             `json:"limit,omitempty"`
}

// GetPersonalizedTipsResponse 获取个性化提示响应
type GetPersonalizedTipsResponse struct {
	Tips      []application.HealthTip `json:"tips"`
	Category  string                  `json:"category"`
	Count     int                     `json:"count"`
	Timestamp string                  `json:"timestamp"`
}

// GetRecommendationsByUserRequest 按用户获取建议请求
type GetRecommendationsByUserRequest struct {
	Types []application.RecommendationType `json:"types,omitempty"`
	Days  int                              `json:"days,omitempty"`
	Limit int                              `json:"limit,omitempty"`
}

// GetTipsByUserRequest 按用户获取提示请求
type GetTipsByUserRequest struct {
	Category application.RecommendationType `json:"category,omitempty"`
	Limit    int                            `json:"limit,omitempty"`
}

// GenerateRecommendations 生成健康建议
// @Summary 生成健康建议
// @Description 基于用户健康数据生成个性化健康建议
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param request body GenerateRecommendationsRequest true "生成建议请求"
// @Success 200 {object} GenerateRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-recommendations/generate [post]
func (h *HealthRecommendationHandler) GenerateRecommendations(c *gin.Context) {
	var req GenerateRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	serviceReq := &application.GenerateRecommendationsRequest{
		UserID: req.UserID,
		Types:  req.Types,
		Days:   req.Days,
		Limit:  req.Limit,
	}

	result, err := h.recommendationService.GenerateRecommendations(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate recommendations",
			Message: err.Error(),
		})
		return
	}

	response := GenerateRecommendationsResponse{
		Recommendations: result.Recommendations,
		Summary:         result.Summary,
		Count:           result.Count,
		Timestamp:       result.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// GetPersonalizedTips 获取个性化健康提示
// @Summary 获取个性化健康提示
// @Description 获取基于用户特征的个性化健康提示
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param request body GetPersonalizedTipsRequest true "获取提示请求"
// @Success 200 {object} GetPersonalizedTipsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-recommendations/tips [post]
func (h *HealthRecommendationHandler) GetPersonalizedTips(c *gin.Context) {
	var req GetPersonalizedTipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	serviceReq := &application.GetPersonalizedTipsRequest{
		UserID:   req.UserID,
		Category: req.Category,
		Limit:    req.Limit,
	}

	result, err := h.recommendationService.GetPersonalizedTips(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get personalized tips",
			Message: err.Error(),
		})
		return
	}

	response := GetPersonalizedTipsResponse{
		Tips:      result.Tips,
		Category:  result.Category,
		Count:     result.Count,
		Timestamp: result.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// GetRecommendationsByUser 按用户获取健康建议
// @Summary 按用户获取健康建议
// @Description 获取指定用户的健康建议
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body GetRecommendationsByUserRequest false "请求参数"
// @Success 200 {object} GenerateRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{user_id}/health-recommendations [post]
func (h *HealthRecommendationHandler) GetRecommendationsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	var req GetRecommendationsByUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认值
		req = GetRecommendationsByUserRequest{}
	}

	serviceReq := &application.GenerateRecommendationsRequest{
		UserID: userID,
		Types:  req.Types,
		Days:   req.Days,
		Limit:  req.Limit,
	}

	result, err := h.recommendationService.GenerateRecommendations(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get recommendations",
			Message: err.Error(),
		})
		return
	}

	response := GenerateRecommendationsResponse{
		Recommendations: result.Recommendations,
		Summary:         result.Summary,
		Count:           result.Count,
		Timestamp:       result.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// GetTipsByUser 按用户获取健康提示
// @Summary 按用户获取健康提示
// @Description 获取指定用户的个性化健康提示
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param category query string false "提示类别"
// @Param limit query int false "限制数量"
// @Success 200 {object} GetPersonalizedTipsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{user_id}/health-tips [get]
func (h *HealthRecommendationHandler) GetTipsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid UUID",
		})
		return
	}

	// 从查询参数获取类别和限制
	categoryStr := c.Query("category")
	limitStr := c.Query("limit")

	var category application.RecommendationType
	if categoryStr != "" {
		category = application.RecommendationType(categoryStr)
	}

	limit := 5 // 默认值
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	serviceReq := &application.GetPersonalizedTipsRequest{
		UserID:   userID,
		Category: category,
		Limit:    limit,
	}

	result, err := h.recommendationService.GetPersonalizedTips(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get tips",
			Message: err.Error(),
		})
		return
	}

	response := GetPersonalizedTipsResponse{
		Tips:      result.Tips,
		Category:  result.Category,
		Count:     result.Count,
		Timestamp: result.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// GetRecommendationTypes 获取建议类型列表
// @Summary 获取建议类型列表
// @Description 获取所有可用的健康建议类型
// @Tags health-recommendations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health-recommendations/types [get]
func (h *HealthRecommendationHandler) GetRecommendationTypes(c *gin.Context) {
	types := map[string]interface{}{
		"types": []map[string]string{
			{"value": "exercise", "label": "运动建议"},
			{"value": "diet", "label": "饮食建议"},
			{"value": "sleep", "label": "睡眠建议"},
			{"value": "stress", "label": "压力管理"},
			{"value": "medical", "label": "医疗建议"},
			{"value": "lifestyle", "label": "生活方式"},
			{"value": "prevention", "label": "预防建议"},
		},
		"priorities": []map[string]string{
			{"value": "high", "label": "高优先级"},
			{"value": "medium", "label": "中优先级"},
			{"value": "low", "label": "低优先级"},
		},
	}

	c.JSON(http.StatusOK, types)
}

// HealthRecommendationHealthCheckHandler 健康建议服务健康检查
// @Summary 健康建议服务健康检查
// @Description 检查健康建议服务的运行状态
// @Tags health-check
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health/recommendations [get]
func (h *HealthRecommendationHandler) HealthRecommendationHealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "health-recommendation",
		"status":  "healthy",
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z07:00"),
	})
}