package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taishanglaojun/health-management/internal/application"
)

// HealthDashboardHandler 崦
type HealthDashboardHandler struct {
	dashboardService *application.HealthDashboardService
}

// NewHealthDashboardHandler 崦
func NewHealthDashboardHandler(dashboardService *application.HealthDashboardService) *HealthDashboardHandler {
	return &HealthDashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardRequest ?
type GetDashboardRequest struct {
	Period string `form:"period" json:"period"`
}

// GetDashboard ?
// @Summary ?
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Param period query string false "" Enums(day,week,month,year)
// @Success 200 {object} application.GetDashboardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-dashboard [get]
func (h *HealthDashboardHandler) GetDashboard(c *gin.Context) {
	var req GetDashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	// ID
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	// 
	var period application.DashboardPeriod
	switch req.Period {
	case "day":
		period = application.DashboardPeriodDay
	case "week":
		period = application.DashboardPeriodWeek
	case "month":
		period = application.DashboardPeriodMonth
	case "year":
		period = application.DashboardPeriodYear
	default:
		period = application.DashboardPeriodMonth
	}

	// 
	serviceReq := &application.GetDashboardRequest{
		UserID: userID,
		Period: period,
	}

	response, err := h.dashboardService.GetDashboard(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get dashboard",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardByUser ID?
// @Summary ID?
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param period query string false "" Enums(day,week,month,year)
// @Success 200 {object} application.GetDashboardResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{user_id}/health-dashboard [get]
func (h *HealthDashboardHandler) GetDashboardByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	var req GetDashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	// 
	var period application.DashboardPeriod
	switch req.Period {
	case "day":
		period = application.DashboardPeriodDay
	case "week":
		period = application.DashboardPeriodWeek
	case "month":
		period = application.DashboardPeriodMonth
	case "year":
		period = application.DashboardPeriodYear
	default:
		period = application.DashboardPeriodMonth
	}

	// 
	serviceReq := &application.GetDashboardRequest{
		UserID: userID,
		Period: period,
	}

	response, err := h.dashboardService.GetDashboard(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get dashboard",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardSummary ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Success 200 {object} application.GetDashboardSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-dashboard/summary [get]
func (h *HealthDashboardHandler) GetDashboardSummary(c *gin.Context) {
	// ID
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	// 
	serviceReq := &application.GetDashboardSummaryRequest{
		UserID: userID,
	}

	response, err := h.dashboardService.GetDashboardSummary(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get dashboard summary",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardSummaryByUser ID?
// @Summary ID?
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} application.GetDashboardSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{user_id}/health-dashboard/summary [get]
func (h *HealthDashboardHandler) GetDashboardSummaryByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	// 
	serviceReq := &application.GetDashboardSummaryRequest{
		UserID: userID,
	}

	response, err := h.dashboardService.GetDashboardSummary(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get dashboard summary",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboardMetrics ?
// @Summary ?
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Param period query string false "" Enums(day,week,month,year)
// @Success 200 {object} GetDashboardMetricsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-dashboard/metrics [get]
func (h *HealthDashboardHandler) GetDashboardMetrics(c *gin.Context) {
	var req GetDashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	// ID
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	// 
	var period application.DashboardPeriod
	switch req.Period {
	case "day":
		period = application.DashboardPeriodDay
	case "week":
		period = application.DashboardPeriodWeek
	case "month":
		period = application.DashboardPeriodMonth
	case "year":
		period = application.DashboardPeriodYear
	default:
		period = application.DashboardPeriodMonth
	}

	// ?
	serviceReq := &application.GetDashboardRequest{
		UserID: userID,
		Period: period,
	}

	response, err := h.dashboardService.GetDashboard(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get dashboard metrics",
			Message: err.Error(),
		})
		return
	}

	// ?
	metricsResponse := GetDashboardMetricsResponse{
		Metrics:   response.Dashboard.KeyMetrics,
		Timestamp: response.Timestamp,
	}

	c.JSON(http.StatusOK, metricsResponse)
}

// GetDashboardCharts ?
// @Summary ?
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Param period query string false "" Enums(day,week,month,year)
// @Success 200 {object} GetDashboardChartsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-dashboard/charts [get]
func (h *HealthDashboardHandler) GetDashboardCharts(c *gin.Context) {
	var req GetDashboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}

	// ID
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: err.Error(),
		})
		return
	}

	// 
	var period application.DashboardPeriod
	switch req.Period {
	case "day":
		period = application.DashboardPeriodDay
	case "week":
		period = application.DashboardPeriodWeek
	case "month":
		period = application.DashboardPeriodMonth
	case "year":
		period = application.DashboardPeriodYear
	default:
		period = application.DashboardPeriodMonth
	}

	// ?
	serviceReq := &application.GetDashboardRequest{
		UserID: userID,
		Period: period,
	}

	response, err := h.dashboardService.GetDashboard(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get dashboard charts",
			Message: err.Error(),
		})
		return
	}

	// ?
	chartsResponse := GetDashboardChartsResponse{
		Charts:    response.Dashboard.TrendCharts,
		Timestamp: response.Timestamp,
	}

	c.JSON(http.StatusOK, chartsResponse)
}

// HealthDashboardHealthCheckHandler 彡?
// @Summary ?
// @Description 齡?
// @Tags ?
// @Produce json
// @Success 200 {object} HealthCheckResponse
// @Router /health/dashboard [get]
func (h *HealthDashboardHandler) HealthDashboardHealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:    "healthy",
		Service:   "health-dashboard",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	})
}

// ?
type GetDashboardMetricsResponse struct {
	Metrics   []application.KeyMetric `json:"metrics"`
	Timestamp time.Time               `json:"timestamp"`
}

type GetDashboardChartsResponse struct {
	Charts    []application.TrendChart `json:"charts"`
	Timestamp time.Time                `json:"timestamp"`
}

