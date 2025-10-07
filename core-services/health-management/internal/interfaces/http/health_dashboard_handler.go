package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taishanglaojun/health-management/internal/application"
)

// HealthDashboardHandler 健康仪表板处理器
type HealthDashboardHandler struct {
	dashboardService *application.HealthDashboardService
}

// NewHealthDashboardHandler 创建健康仪表板处理器
func NewHealthDashboardHandler(dashboardService *application.HealthDashboardService) *HealthDashboardHandler {
	return &HealthDashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardRequest 获取仪表板请求
type GetDashboardRequest struct {
	Period string `form:"period" json:"period"`
}

// GetDashboard 获取健康仪表板
// @Summary 获取健康仪表板
// @Description 获取用户的健康仪表板数据，包括概览、关键指标、趋势图表等
// @Tags 健康仪表板
// @Accept json
// @Produce json
// @Param period query string false "时间周期" Enums(day,week,month,year)
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

	// 从上下文获取用户ID（假设已通过中间件设置）
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

	// 转换周期参数
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

	// 调用服务
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

// GetDashboardByUser 根据用户ID获取健康仪表板
// @Summary 根据用户ID获取健康仪表板
// @Description 管理员或授权用户获取指定用户的健康仪表板
// @Tags 健康仪表板
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param period query string false "时间周期" Enums(day,week,month,year)
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

	// 转换周期参数
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

	// 调用服务
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

// GetDashboardSummary 获取仪表板摘要
// @Summary 获取仪表板摘要
// @Description 获取用户健康仪表板的简要摘要信息
// @Tags 健康仪表板
// @Accept json
// @Produce json
// @Success 200 {object} application.GetDashboardSummaryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /health-dashboard/summary [get]
func (h *HealthDashboardHandler) GetDashboardSummary(c *gin.Context) {
	// 从上下文获取用户ID
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

	// 调用服务
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

// GetDashboardSummaryByUser 根据用户ID获取仪表板摘要
// @Summary 根据用户ID获取仪表板摘要
// @Description 管理员或授权用户获取指定用户的健康仪表板摘要
// @Tags 健康仪表板
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
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

	// 调用服务
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

// GetDashboardMetrics 获取仪表板关键指标
// @Summary 获取仪表板关键指标
// @Description 获取用户健康仪表板的关键指标数据
// @Tags 健康仪表板
// @Accept json
// @Produce json
// @Param period query string false "时间周期" Enums(day,week,month,year)
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

	// 从上下文获取用户ID
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

	// 转换周期参数
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

	// 调用服务获取完整仪表板
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

	// 只返回关键指标
	metricsResponse := GetDashboardMetricsResponse{
		Metrics:   response.Dashboard.KeyMetrics,
		Timestamp: response.Timestamp,
	}

	c.JSON(http.StatusOK, metricsResponse)
}

// GetDashboardCharts 获取仪表板图表数据
// @Summary 获取仪表板图表数据
// @Description 获取用户健康仪表板的趋势图表数据
// @Tags 健康仪表板
// @Accept json
// @Produce json
// @Param period query string false "时间周期" Enums(day,week,month,year)
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

	// 从上下文获取用户ID
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

	// 转换周期参数
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

	// 调用服务获取完整仪表板
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

	// 只返回图表数据
	chartsResponse := GetDashboardChartsResponse{
		Charts:    response.Dashboard.TrendCharts,
		Timestamp: response.Timestamp,
	}

	c.JSON(http.StatusOK, chartsResponse)
}

// HealthDashboardHealthCheckHandler 健康仪表板健康检查
// @Summary 健康仪表板服务健康检查
// @Description 检查健康仪表板服务的运行状态
// @Tags 健康检查
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

// 响应结构体
type GetDashboardMetricsResponse struct {
	Metrics   []application.KeyMetric `json:"metrics"`
	Timestamp time.Time               `json:"timestamp"`
}

type GetDashboardChartsResponse struct {
	Charts    []application.TrendChart `json:"charts"`
	Timestamp time.Time                `json:"timestamp"`
}