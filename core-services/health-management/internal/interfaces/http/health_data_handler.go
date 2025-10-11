package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/application"
	"github.com/taishanglaojun/health-management/internal/domain"
)

// HealthDataHandler 健康数据HTTP处理�?
type HealthDataHandler struct {
	healthDataService *application.HealthDataService
}

// NewHealthDataHandler 创建健康数据HTTP处理�?
func NewHealthDataHandler(healthDataService *application.HealthDataService) *HealthDataHandler {
	return &HealthDataHandler{
		healthDataService: healthDataService,
	}
}

// CreateHealthDataRequest 创建健康数据请求
type CreateHealthDataRequest struct {
	UserID     string                   `json:"user_id" binding:"required"`
	DataType   domain.HealthDataType    `json:"data_type" binding:"required"`
	Value      float64                  `json:"value" binding:"required"`
	Unit       string                   `json:"unit" binding:"required"`
	Source     domain.HealthDataSource  `json:"source" binding:"required"`
	DeviceID   *string                  `json:"device_id,omitempty"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
	RecordedAt *time.Time               `json:"recorded_at,omitempty"`
}

// UpdateHealthDataRequest 更新健康数据请求
type UpdateHealthDataRequest struct {
	Value    float64                `json:"value" binding:"required"`
	Unit     string                 `json:"unit" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HealthDataResponse 健康数据响应
type HealthDataResponse struct {
	ID         string                   `json:"id"`
	UserID     string                   `json:"user_id"`
	DataType   domain.HealthDataType    `json:"data_type"`
	Value      float64                  `json:"value"`
	Unit       string                   `json:"unit"`
	Source     domain.HealthDataSource  `json:"source"`
	DeviceID   *string                  `json:"device_id,omitempty"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
	RecordedAt time.Time                `json:"recorded_at"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	RiskLevel  string                   `json:"risk_level"`
	IsAbnormal bool                     `json:"is_abnormal"`
}

// HealthDataListResponse 健康数据列表响应
type HealthDataListResponse struct {
	Data       []HealthDataResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// HealthDataStatisticsResponse 健康数据统计响应
type HealthDataStatisticsResponse struct {
	UserID    string                   `json:"user_id"`
	DataType  domain.HealthDataType    `json:"data_type"`
	Count     int64                    `json:"count"`
	Average   float64                  `json:"average"`
	Min       float64                  `json:"min"`
	Max       float64                  `json:"max"`
	StartTime time.Time                `json:"start_time"`
	EndTime   time.Time                `json:"end_time"`
}

// CreateHealthData 创建健康数据
// @Summary 创建健康数据
// @Description 创建新的健康数据记录
// @Tags health-data
// @Accept json
// @Produce json
// @Param request body CreateHealthDataRequest true "创建健康数据请求"
// @Success 201 {object} HealthDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-data [post]
func (h *HealthDataHandler) CreateHealthData(c *gin.Context) {
	var req CreateHealthDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	createReq := &application.CreateHealthDataRequest{
		UserID:     userID,
		DataType:   req.DataType,
		Value:      req.Value,
		Unit:       req.Unit,
		Source:     req.Source,
		DeviceID:   req.DeviceID,
		Metadata:   req.Metadata,
		RecordedAt: req.RecordedAt,
	}

	resp, err := h.healthDataService.CreateHealthData(c.Request.Context(), createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "CREATE_FAILED",
			Message: "创建健康数据失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, h.toHealthDataResponse(resp))
}

// GetHealthData 获取健康数据
// @Summary 获取健康数据
// @Description 根据ID获取健康数据
// @Tags health-data
// @Produce json
// @Param id path string true "健康数据ID"
// @Success 200 {object} HealthDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-data/{id} [get]
func (h *HealthDataHandler) GetHealthData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康数据ID格式无效",
		})
		return
	}

	resp, err := h.healthDataService.GetHealthData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_FAILED",
			Message: "获取健康数据失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康数据不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthDataResponse(resp))
}

// UpdateHealthData 更新健康数据
// @Summary 更新健康数据
// @Description 更新健康数据记录
// @Tags health-data
// @Accept json
// @Produce json
// @Param id path string true "健康数据ID"
// @Param request body UpdateHealthDataRequest true "更新健康数据请求"
// @Success 200 {object} HealthDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-data/{id} [put]
func (h *HealthDataHandler) UpdateHealthData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康数据ID格式无效",
		})
		return
	}

	var req UpdateHealthDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数无效",
			Details: err.Error(),
		})
		return
	}

	updateReq := &application.UpdateHealthDataRequest{
		ID:       id,
		Value:    req.Value,
		Unit:     req.Unit,
		Metadata: req.Metadata,
	}

	resp, err := h.healthDataService.UpdateHealthData(c.Request.Context(), updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: "更新健康数据失败",
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "健康数据不存�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthDataResponse(resp))
}

// DeleteHealthData 删除健康数据
// @Summary 删除健康数据
// @Description 删除健康数据记录
// @Tags health-data
// @Produce json
// @Param id path string true "健康数据ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/health-data/{id} [delete]
func (h *HealthDataHandler) DeleteHealthData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_ID",
			Message: "健康数据ID格式无效",
		})
		return
	}

	err = h.healthDataService.DeleteHealthData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "DELETE_FAILED",
			Message: "删除健康数据失败",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetHealthDataByUser 获取用户的健康数�?
// @Summary 获取用户的健康数�?
// @Description 分页获取用户的健康数据列�?
// @Tags health-data
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Param data_type query string false "数据类型"
// @Param start_time query string false "开始时�?(RFC3339格式)"
// @Param end_time query string false "结束时间 (RFC3339格式)"
// @Success 200 {object} HealthDataListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-data [get]
func (h *HealthDataHandler) GetHealthDataByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	dataTypeStr := c.Query("data_type")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &application.GetHealthDataByUserRequest{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}

	// 解析数据类型
	if dataTypeStr != "" {
		dataType := domain.HealthDataType(dataTypeStr)
		req.DataType = &dataType
	}

	// 解析时间范围
	if startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_START_TIME",
				Message: "开始时间格式无�?,
			})
			return
		}
		req.StartTime = &startTime
	}

	if endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_END_TIME",
				Message: "结束时间格式无效",
			})
			return
		}
		req.EndTime = &endTime
	}

	resp, err := h.healthDataService.GetHealthDataByUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_FAILED",
			Message: "获取健康数据失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthDataListResponse(resp, page, pageSize))
}

// GetLatestHealthData 获取最新健康数�?
// @Summary 获取最新健康数�?
// @Description 获取用户特定类型的最新健康数�?
// @Tags health-data
// @Produce json
// @Param user_id path string true "用户ID"
// @Param data_type path string true "数据类型"
// @Success 200 {object} HealthDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-data/latest/{data_type} [get]
func (h *HealthDataHandler) GetLatestHealthData(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	dataType := domain.HealthDataType(c.Param("data_type"))

	req := &application.GetLatestHealthDataRequest{
		UserID:   userID,
		DataType: dataType,
	}

	resp, err := h.healthDataService.GetLatestHealthData(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_FAILED",
			Message: "获取最新健康数据失�?,
			Details: err.Error(),
		})
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "未找到健康数�?,
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthDataResponse(resp))
}

// GetHealthDataStatistics 获取健康数据统计
// @Summary 获取健康数据统计
// @Description 获取用户健康数据的统计信�?
// @Tags health-data
// @Produce json
// @Param user_id path string true "用户ID"
// @Param data_type path string true "数据类型"
// @Param start_time query string true "开始时�?(RFC3339格式)"
// @Param end_time query string true "结束时间 (RFC3339格式)"
// @Success 200 {object} HealthDataStatisticsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-data/statistics/{data_type} [get]
func (h *HealthDataHandler) GetHealthDataStatistics(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	dataType := domain.HealthDataType(c.Param("data_type"))
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_TIME_RANGE",
			Message: "开始时间和结束时间不能为空",
		})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_START_TIME",
			Message: "开始时间格式无�?,
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_END_TIME",
			Message: "结束时间格式无效",
		})
		return
	}

	req := &application.GetHealthDataStatisticsRequest{
		UserID:    userID,
		DataType:  dataType,
		StartTime: startTime,
		EndTime:   endTime,
	}

	resp, err := h.healthDataService.GetHealthDataStatistics(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_STATISTICS_FAILED",
			Message: "获取健康数据统计失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, HealthDataStatisticsResponse{
		UserID:    resp.UserID.String(),
		DataType:  resp.DataType,
		Count:     resp.Count,
		Average:   resp.Average,
		Min:       resp.Min,
		Max:       resp.Max,
		StartTime: resp.StartTime,
		EndTime:   resp.EndTime,
	})
}

// GetAbnormalHealthData 获取异常健康数据
// @Summary 获取异常健康数据
// @Description 获取用户的异常健康数�?
// @Tags health-data
// @Produce json
// @Param user_id path string true "用户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页大小" default(20)
// @Success 200 {object} HealthDataListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/users/{user_id}/health-data/abnormal [get]
func (h *HealthDataHandler) GetAbnormalHealthData(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_USER_ID",
			Message: "用户ID格式无效",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &application.GetAbnormalHealthDataRequest{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}

	resp, err := h.healthDataService.GetAbnormalHealthData(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "GET_ABNORMAL_DATA_FAILED",
			Message: "获取异常健康数据失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, h.toHealthDataListResponse(resp, page, pageSize))
}

// 辅助方法

func (h *HealthDataHandler) toHealthDataResponse(data *application.HealthDataResponse) HealthDataResponse {
	return HealthDataResponse{
		ID:         data.ID.String(),
		UserID:     data.UserID.String(),
		DataType:   data.DataType,
		Value:      data.Value,
		Unit:       data.Unit,
		Source:     data.Source,
		DeviceID:   data.DeviceID,
		Metadata:   data.Metadata,
		RecordedAt: data.RecordedAt,
		CreatedAt:  data.CreatedAt,
		UpdatedAt:  data.UpdatedAt,
		RiskLevel:  data.RiskLevel,
		IsAbnormal: data.IsAbnormal,
	}
}

func (h *HealthDataHandler) toHealthDataListResponse(data []*application.HealthDataResponse, page, pageSize int) HealthDataListResponse {
	responses := make([]HealthDataResponse, len(data))
	for i, item := range data {
		responses[i] = h.toHealthDataResponse(item)
	}

	total := int64(len(data)) // 这里简化处理，实际应该从服务层获取总数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return HealthDataListResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
