package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/application"
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthAnalysisHandler HTTP?
type HealthAnalysisHandler struct {
	healthAnalysisService *application.HealthAnalysisService
}

// NewHealthAnalysisHandler HTTP?
func NewHealthAnalysisHandler(healthAnalysisService *application.HealthAnalysisService) *HealthAnalysisHandler {
	return &HealthAnalysisHandler{
		healthAnalysisService: healthAnalysisService,
	}
}

// AnalyzeHealthTrendRequest 
type AnalyzeHealthTrendRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	DataType  string `json:"data_type" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Period    string `json:"period" binding:"required,oneof=daily weekly monthly"`
}

// AssessHealthRiskRequest 
type AssessHealthRiskRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// GenerateHealthInsightsRequest 
type GenerateHealthInsightsRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

// AnalyzeHealthTrend 
// @Summary 
// @Description 
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param request body AnalyzeHealthTrendRequest true ""
// @Success 200 {object} application.HealthTrendAnalysisResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-analysis/trend [post]
func (h *HealthAnalysisHandler) AnalyzeHealthTrend(c *gin.Context) {
	var req AnalyzeHealthTrendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	dataType, err := parseHealthDataType(req.DataType)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_DATA_TYPE",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TIME_RANGE",
			Message: "䲻?,
		})
		return
	}

	// 
	serviceReq := &application.HealthTrendAnalysisRequest{
		UserID:    userID,
		DataType:  dataType,
		StartTime: startTime,
		EndTime:   endTime,
		Period:    req.Period,
	}

	// 
	response, err := h.healthAnalysisService.AnalyzeHealthTrend(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ANALYSIS_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AssessHealthRisk 
// @Summary 
// @Description 彡?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param request body AssessHealthRiskRequest true ""
// @Success 200 {object} application.HealthRiskAssessmentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-analysis/risk-assessment [post]
func (h *HealthAnalysisHandler) AssessHealthRisk(c *gin.Context) {
	var req AssessHealthRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	serviceReq := &application.HealthRiskAssessmentRequest{
		UserID: userID,
	}

	// 
	response, err := h.healthAnalysisService.AssessHealthRisk(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RISK_ASSESSMENT_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateHealthInsights 
// @Summary 
// @Description 
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param request body GenerateHealthInsightsRequest true ""
// @Success 200 {object} application.HealthInsightsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-analysis/insights [post]
func (h *HealthAnalysisHandler) GenerateHealthInsights(c *gin.Context) {
	var req GenerateHealthInsightsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TIME_RANGE",
			Message: "䲻?,
		})
		return
	}

	// 
	serviceReq := &application.HealthInsightsRequest{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 
	response, err := h.healthAnalysisService.GenerateHealthInsights(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INSIGHTS_GENERATION_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthTrendByUser 
// @Summary 
// @Description ?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param data_type query string true ""
// @Param start_time query string true "?
// @Param end_time query string true ""
// @Param period query string true "" Enums(daily, weekly, monthly)
// @Success 200 {object} application.HealthTrendAnalysisResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-analysis/trend [get]
func (h *HealthAnalysisHandler) GetHealthTrendByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	dataTypeStr := c.Query("data_type")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	period := c.Query("period")

	// 
	if userIDStr == "" || dataTypeStr == "" || startTimeStr == "" || endTimeStr == "" || period == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETERS",
			Message: "?,
		})
		return
	}

	// ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	dataType, err := parseHealthDataType(dataTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_DATA_TYPE",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if period != "daily" && period != "weekly" && period != "monthly" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PERIOD",
			Message: "?dailyweekly ?monthly",
		})
		return
	}

	// 
	serviceReq := &application.HealthTrendAnalysisRequest{
		UserID:    userID,
		DataType:  dataType,
		StartTime: startTime,
		EndTime:   endTime,
		Period:    period,
	}

	// 
	response, err := h.healthAnalysisService.AnalyzeHealthTrend(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ANALYSIS_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthRiskByUser 
// @Summary 
// @Description ?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} application.HealthRiskAssessmentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-analysis/risk-assessment [get]
func (h *HealthAnalysisHandler) GetHealthRiskByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")

	// ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	serviceReq := &application.HealthRiskAssessmentRequest{
		UserID: userID,
	}

	// 
	response, err := h.healthAnalysisService.AssessHealthRisk(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RISK_ASSESSMENT_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthInsightsByUser 
// @Summary 
// @Description ?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param start_time query string true "?
// @Param end_time query string true ""
// @Success 200 {object} application.HealthInsightsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-analysis/insights [get]
func (h *HealthAnalysisHandler) GetHealthInsightsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 
	if userIDStr == "" || startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETERS",
			Message: "?,
		})
		return
	}

	// ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TIME_RANGE",
			Message: "䲻?,
		})
		return
	}

	// 
	serviceReq := &application.HealthInsightsRequest{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 
	response, err := h.healthAnalysisService.GenerateHealthInsights(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INSIGHTS_GENERATION_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// 

// parseHealthDataType 
func parseHealthDataType(dataTypeStr string) (domain.HealthDataType, error) {
	switch dataTypeStr {
	case "heart_rate":
		return domain.HeartRate, nil
	case "blood_pressure":
		return domain.BloodPressure, nil
	case "steps":
		return domain.Steps, nil
	case "sleep_duration":
		return domain.SleepDuration, nil
	case "stress_level":
		return domain.StressLevel, nil
	default:
		return "", fmt.Errorf("unsupported health data type: %s", dataTypeStr)
	}
}

// HealthAnalysisHealthCheckHandler ?
func HealthAnalysisHealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "health-analysis",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

