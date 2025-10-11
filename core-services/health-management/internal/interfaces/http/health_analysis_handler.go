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

// HealthAnalysisHandler 健康分析HTTP处理�?
type HealthAnalysisHandler struct {
	healthAnalysisService *application.HealthAnalysisService
}

// NewHealthAnalysisHandler 创建健康分析HTTP处理�?
func NewHealthAnalysisHandler(healthAnalysisService *application.HealthAnalysisService) *HealthAnalysisHandler {
	return &HealthAnalysisHandler{
		healthAnalysisService: healthAnalysisService,
	}
}

// AnalyzeHealthTrendRequest 健康趋势分析请求
type AnalyzeHealthTrendRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	DataType  string `json:"data_type" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Period    string `json:"period" binding:"required,oneof=daily weekly monthly"`
}

// AssessHealthRiskRequest 健康风险评估请求
type AssessHealthRiskRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// GenerateHealthInsightsRequest 健康洞察生成请求
type GenerateHealthInsightsRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

// AnalyzeHealthTrend 分析健康趋势
// @Summary 分析健康趋势
// @Description 分析用户指定时间范围内的健康数据趋势
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param request body AnalyzeHealthTrendRequest true "健康趋势分析请求"
// @Success 200 {object} application.HealthTrendAnalysisResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-analysis/trend [post]
func (h *HealthAnalysisHandler) AnalyzeHealthTrend(c *gin.Context) {
	var req AnalyzeHealthTrendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 解析用户ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
			Details: err.Error(),
		})
		return
	}

	// 解析数据类型
	dataType, err := parseHealthDataType(req.DataType)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_DATA_TYPE",
			Message: "健康数据类型无效",
			Details: err.Error(),
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "开始时间格式无�?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "结束时间格式无效",
			Details: err.Error(),
		})
		return
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TIME_RANGE",
			Message: "结束时间不能早于开始时�?,
		})
		return
	}

	// 构建服务请求
	serviceReq := &application.HealthTrendAnalysisRequest{
		UserID:    userID,
		DataType:  dataType,
		StartTime: startTime,
		EndTime:   endTime,
		Period:    req.Period,
	}

	// 调用服务
	response, err := h.healthAnalysisService.AnalyzeHealthTrend(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ANALYSIS_FAILED",
			Message: "健康趋势分析失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AssessHealthRisk 评估健康风险
// @Summary 评估健康风险
// @Description 评估用户的整体健康风�?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param request body AssessHealthRiskRequest true "健康风险评估请求"
// @Success 200 {object} application.HealthRiskAssessmentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-analysis/risk-assessment [post]
func (h *HealthAnalysisHandler) AssessHealthRisk(c *gin.Context) {
	var req AssessHealthRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 解析用户ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
			Details: err.Error(),
		})
		return
	}

	// 构建服务请求
	serviceReq := &application.HealthRiskAssessmentRequest{
		UserID: userID,
	}

	// 调用服务
	response, err := h.healthAnalysisService.AssessHealthRisk(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RISK_ASSESSMENT_FAILED",
			Message: "健康风险评估失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GenerateHealthInsights 生成健康洞察
// @Summary 生成健康洞察
// @Description 生成用户的个性化健康洞察报告
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param request body GenerateHealthInsightsRequest true "健康洞察生成请求"
// @Success 200 {object} application.HealthInsightsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-analysis/insights [post]
func (h *HealthAnalysisHandler) GenerateHealthInsights(c *gin.Context) {
	var req GenerateHealthInsightsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	// 解析用户ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
			Details: err.Error(),
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "开始时间格式无�?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "结束时间格式无效",
			Details: err.Error(),
		})
		return
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TIME_RANGE",
			Message: "结束时间不能早于开始时�?,
		})
		return
	}

	// 构建服务请求
	serviceReq := &application.HealthInsightsRequest{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 调用服务
	response, err := h.healthAnalysisService.GenerateHealthInsights(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INSIGHTS_GENERATION_FAILED",
			Message: "健康洞察生成失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthTrendByUser 获取用户健康趋势
// @Summary 获取用户健康趋势
// @Description 获取指定用户的健康数据趋势分�?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param data_type query string true "数据类型"
// @Param start_time query string true "开始时�?
// @Param end_time query string true "结束时间"
// @Param period query string true "时间周期" Enums(daily, weekly, monthly)
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

	// 验证必需参数
	if userIDStr == "" || dataTypeStr == "" || startTimeStr == "" || endTimeStr == "" || period == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETERS",
			Message: "缺少必需的查询参�?,
		})
		return
	}

	// 解析用户ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
			Details: err.Error(),
		})
		return
	}

	// 解析数据类型
	dataType, err := parseHealthDataType(dataTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_DATA_TYPE",
			Message: "健康数据类型无效",
			Details: err.Error(),
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "开始时间格式无�?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "结束时间格式无效",
			Details: err.Error(),
		})
		return
	}

	// 验证周期参数
	if period != "daily" && period != "weekly" && period != "monthly" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PERIOD",
			Message: "时间周期必须�?daily、weekly �?monthly",
		})
		return
	}

	// 构建服务请求
	serviceReq := &application.HealthTrendAnalysisRequest{
		UserID:    userID,
		DataType:  dataType,
		StartTime: startTime,
		EndTime:   endTime,
		Period:    period,
	}

	// 调用服务
	response, err := h.healthAnalysisService.AnalyzeHealthTrend(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "ANALYSIS_FAILED",
			Message: "健康趋势分析失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthRiskByUser 获取用户健康风险评估
// @Summary 获取用户健康风险评估
// @Description 获取指定用户的健康风险评估结�?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} application.HealthRiskAssessmentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-analysis/risk-assessment [get]
func (h *HealthAnalysisHandler) GetHealthRiskByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")

	// 解析用户ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
			Details: err.Error(),
		})
		return
	}

	// 构建服务请求
	serviceReq := &application.HealthRiskAssessmentRequest{
		UserID: userID,
	}

	// 调用服务
	response, err := h.healthAnalysisService.AssessHealthRisk(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "RISK_ASSESSMENT_FAILED",
			Message: "健康风险评估失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHealthInsightsByUser 获取用户健康洞察
// @Summary 获取用户健康洞察
// @Description 获取指定用户的健康洞察报�?
// @Tags health-analysis
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param start_time query string true "开始时�?
// @Param end_time query string true "结束时间"
// @Success 200 {object} application.HealthInsightsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-analysis/insights [get]
func (h *HealthAnalysisHandler) GetHealthInsightsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 验证必需参数
	if userIDStr == "" || startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_PARAMETERS",
			Message: "缺少必需的查询参�?,
		})
		return
	}

	// 解析用户ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
			Details: err.Error(),
		})
		return
	}

	// 解析时间
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "开始时间格式无�?,
			Details: err.Error(),
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "结束时间格式无效",
			Details: err.Error(),
		})
		return
	}

	// 验证时间范围
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_TIME_RANGE",
			Message: "结束时间不能早于开始时�?,
		})
		return
	}

	// 构建服务请求
	serviceReq := &application.HealthInsightsRequest{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 调用服务
	response, err := h.healthAnalysisService.GenerateHealthInsights(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INSIGHTS_GENERATION_FAILED",
			Message: "健康洞察生成失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// 辅助函数

// parseHealthDataType 解析健康数据类型
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

// HealthAnalysisHealthCheckHandler 健康分析服务健康检�?
func HealthAnalysisHealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "health-analysis",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}
