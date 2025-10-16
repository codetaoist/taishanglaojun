package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taishanglaojun/health-management/internal/application"
)

// HealthRecommendationHandler 鴦?
type HealthRecommendationHandler struct {
	recommendationService *application.HealthRecommendationService
}

// NewHealthRecommendationHandler 鴦?
func NewHealthRecommendationHandler(recommendationService *application.HealthRecommendationService) *HealthRecommendationHandler {
	return &HealthRecommendationHandler{
		recommendationService: recommendationService,
	}
}

// GenerateRecommendationsRequest 
type GenerateRecommendationsRequest struct {
	UserID uuid.UUID                                `json:"user_id" binding:"required"`
	Types  []application.RecommendationType         `json:"types,omitempty"`
	Days   int                                      `json:"days,omitempty"`
	Limit  int                                      `json:"limit,omitempty"`
}

// GenerateRecommendationsResponse 
type GenerateRecommendationsResponse struct {
	Recommendations []application.HealthRecommendation `json:"recommendations"`
	Summary         string                             `json:"summary"`
	Count           int                                `json:"count"`
	Timestamp       string                             `json:"timestamp"`
}

// GetPersonalizedTipsRequest 
type GetPersonalizedTipsRequest struct {
	UserID   uuid.UUID                       `json:"user_id" binding:"required"`
	Category application.RecommendationType  `json:"category,omitempty"`
	Limit    int                             `json:"limit,omitempty"`
}

// GetPersonalizedTipsResponse 
type GetPersonalizedTipsResponse struct {
	Tips      []application.HealthTip `json:"tips"`
	Category  string                  `json:"category"`
	Count     int                     `json:"count"`
	Timestamp string                  `json:"timestamp"`
}

// GetRecommendationsByUserRequest ?
type GetRecommendationsByUserRequest struct {
	Types []application.RecommendationType `json:"types,omitempty"`
	Days  int                              `json:"days,omitempty"`
	Limit int                              `json:"limit,omitempty"`
}

// GetTipsByUserRequest ?
type GetTipsByUserRequest struct {
	Category application.RecommendationType `json:"category,omitempty"`
	Limit    int                            `json:"limit,omitempty"`
}

// GenerateRecommendations 
// @Summary 
// @Description 
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param request body GenerateRecommendationsRequest true ""
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

// GetPersonalizedTips 
// @Summary 
// @Description 
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param request body GetPersonalizedTipsRequest true ""
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

// GetRecommendationsByUser ?
// @Summary ?
// @Description ?
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param request body GetRecommendationsByUserRequest false ""
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
		// ?
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

// GetTipsByUser ?
// @Summary ?
// @Description 
// @Tags health-recommendations
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param category query string false ""
// @Param limit query int false ""
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

	// 
	categoryStr := c.Query("category")
	limitStr := c.Query("limit")

	var category application.RecommendationType
	if categoryStr != "" {
		category = application.RecommendationType(categoryStr)
	}

	limit := 5 // ?
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

// GetRecommendationTypes 
// @Summary 
// @Description 
// @Tags health-recommendations
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health-recommendations/types [get]
func (h *HealthRecommendationHandler) GetRecommendationTypes(c *gin.Context) {
	types := map[string]interface{}{
		"types": []map[string]string{
			{"value": "exercise", "label": ""},
			{"value": "diet", "label": ""},
			{"value": "sleep", "label": ""},
			{"value": "stress", "label": ""},
			{"value": "medical", "label": ""},
			{"value": "lifestyle", "label": ""},
			{"value": "prevention", "label": ""},
		},
		"priorities": []map[string]string{
			{"value": "high", "label": ""},
			{"value": "medium", "label": ""},
			{"value": "low", "label": ""},
		},
	}

	c.JSON(http.StatusOK, types)
}

// HealthRecommendationHealthCheckHandler ?
// @Summary ?
// @Description 齡?
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

