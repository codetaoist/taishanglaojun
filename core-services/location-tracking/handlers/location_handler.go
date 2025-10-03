package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/models"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LocationHandler 位置处理器
type LocationHandler struct {
	service *services.LocationService
	logger  *zap.Logger
}

// NewLocationHandler 创建位置处理器实例
func NewLocationHandler(service *services.LocationService, logger *zap.Logger) *LocationHandler {
	return &LocationHandler{
		service: service,
		logger:  logger,
	}
}

// CreateTrajectory 创建轨迹
// @Summary 创建轨迹
// @Description 创建新的轨迹记录
// @Tags trajectories
// @Accept json
// @Produce json
// @Param request body models.TrajectoryRequest true "轨迹创建请求"
// @Success 201 {object} models.TrajectoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/trajectories [post]
func (h *LocationHandler) CreateTrajectory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.TrajectoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	trajectory, err := h.service.CreateTrajectory(userID, req)
	if err != nil {
		h.logger.Error("Failed to create trajectory", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trajectory"})
		return
	}

	c.JSON(http.StatusCreated, trajectory.ToResponse())
}

// GetTrajectories 获取轨迹列表
// @Summary 获取轨迹列表
// @Description 获取用户的轨迹列表，支持分页和过滤
// @Tags trajectories
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Param name query string false "轨迹名称过滤"
// @Param is_active query bool false "是否活跃"
// @Param start_time query int64 false "开始时间（毫秒时间戳）"
// @Param end_time query int64 false "结束时间（毫秒时间戳）"
// @Param order_by query string false "排序字段" default("created_at")
// @Param order query string false "排序方向" default("desc")
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/trajectories [get]
func (h *LocationHandler) GetTrajectories(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 解析查询参数
	query := models.TrajectoryQuery{
		Limit:   20,
		Offset:  0,
		OrderBy: "created_at",
		Order:   "desc",
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		query.Offset = (page - 1) * query.Limit
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		query.Limit = limit
	}

	query.Name = c.Query("name")
	query.OrderBy = c.DefaultQuery("order_by", "created_at")
	query.Order = c.DefaultQuery("order", "desc")

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			query.IsActive = &isActive
		}
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			query.EndTime = &endTime
		}
	}

	trajectories, total, err := h.service.GetTrajectories(userID, query)
	if err != nil {
		h.logger.Error("Failed to get trajectories", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trajectories"})
		return
	}

	// 转换为响应格式
	responses := make([]models.TrajectoryResponse, len(trajectories))
	for i, trajectory := range trajectories {
		responses[i] = trajectory.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"total": total,
		"page":  (query.Offset / query.Limit) + 1,
		"limit": query.Limit,
	})
}

// GetTrajectory 获取轨迹详情
// @Summary 获取轨迹详情
// @Description 获取指定轨迹的详细信息
// @Tags trajectories
// @Accept json
// @Produce json
// @Param id path string true "轨迹ID"
// @Param include_points query bool false "是否包含位置点" default(false)
// @Success 200 {object} models.TrajectoryDetailResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/trajectories/{id} [get]
func (h *LocationHandler) GetTrajectory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	trajectoryID := c.Param("id")
	includePoints, _ := strconv.ParseBool(c.DefaultQuery("include_points", "false"))

	trajectory, err := h.service.GetTrajectory(userID, trajectoryID, includePoints)
	if err != nil {
		if err.Error() == "trajectory not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trajectory not found"})
			return
		}
		h.logger.Error("Failed to get trajectory", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trajectory"})
		return
	}

	c.JSON(http.StatusOK, trajectory.ToDetailResponse())
}

// UpdateTrajectory 更新轨迹
// @Summary 更新轨迹
// @Description 更新轨迹信息
// @Tags trajectories
// @Accept json
// @Produce json
// @Param id path string true "轨迹ID"
// @Param request body models.TrajectoryUpdateRequest true "轨迹更新请求"
// @Success 200 {object} models.TrajectoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/trajectories/{id} [put]
func (h *LocationHandler) UpdateTrajectory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	trajectoryID := c.Param("id")

	var req models.TrajectoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	trajectory, err := h.service.UpdateTrajectory(userID, trajectoryID, req)
	if err != nil {
		if err.Error() == "trajectory not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trajectory not found"})
			return
		}
		h.logger.Error("Failed to update trajectory", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trajectory"})
		return
	}

	c.JSON(http.StatusOK, trajectory.ToResponse())
}

// DeleteTrajectory 删除轨迹
// @Summary 删除轨迹
// @Description 删除指定轨迹及其所有位置点
// @Tags trajectories
// @Accept json
// @Produce json
// @Param id path string true "轨迹ID"
// @Success 204
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/trajectories/{id} [delete]
func (h *LocationHandler) DeleteTrajectory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	trajectoryID := c.Param("id")

	err := h.service.DeleteTrajectory(userID, trajectoryID)
	if err != nil {
		if err.Error() == "trajectory not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trajectory not found"})
			return
		}
		h.logger.Error("Failed to delete trajectory", zap.Error(err), zap.String("trajectory_id", trajectoryID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete trajectory"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddLocationPoint 添加位置点
// @Summary 添加位置点
// @Description 向轨迹添加单个位置点
// @Tags location-points
// @Accept json
// @Produce json
// @Param request body models.LocationPointRequest true "位置点请求"
// @Success 201 {object} models.LocationPointResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/location-points [post]
func (h *LocationHandler) AddLocationPoint(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.LocationPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	point, err := h.service.AddLocationPoint(userID, req)
	if err != nil {
		if err.Error() == "trajectory not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trajectory not found"})
			return
		}
		h.logger.Error("Failed to add location point", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add location point"})
		return
	}

	c.JSON(http.StatusCreated, point.ToResponse())
}

// AddLocationPointsBatch 批量添加位置点
// @Summary 批量添加位置点
// @Description 向轨迹批量添加位置点
// @Tags location-points
// @Accept json
// @Produce json
// @Param request body models.LocationPointBatchRequest true "批量位置点请求"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/location-points/batch [post]
func (h *LocationHandler) AddLocationPointsBatch(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.LocationPointBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	points, err := h.service.AddLocationPointsBatch(userID, req)
	if err != nil {
		if err.Error() == "trajectory not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trajectory not found"})
			return
		}
		h.logger.Error("Failed to add location points batch", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add location points batch"})
		return
	}

	responses := make([]models.LocationPointResponse, len(points))
	for i, point := range points {
		responses[i] = point.ToResponse()
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":  responses,
		"count": len(responses),
	})
}

// GetLocationPoints 获取位置点
// @Summary 获取位置点
// @Description 获取位置点列表，支持分页和过滤
// @Tags location-points
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(100)
// @Param trajectory_id query string false "轨迹ID"
// @Param start_time query int64 false "开始时间（毫秒时间戳）"
// @Param end_time query int64 false "结束时间（毫秒时间戳）"
// @Param min_lat query float64 false "最小纬度"
// @Param max_lat query float64 false "最大纬度"
// @Param min_lng query float64 false "最小经度"
// @Param max_lng query float64 false "最大经度"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/location-points [get]
func (h *LocationHandler) GetLocationPoints(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 解析查询参数
	query := models.LocationPointQuery{
		Limit:  100,
		Offset: 0,
	}

	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		query.Offset = (page - 1) * query.Limit
	}

	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "100")); err == nil && limit > 0 && limit <= 1000 {
		query.Limit = limit
	}

	query.TrajectoryID = c.Query("trajectory_id")

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			query.EndTime = &endTime
		}
	}

	if minLatStr := c.Query("min_lat"); minLatStr != "" {
		if minLat, err := strconv.ParseFloat(minLatStr, 64); err == nil {
			query.MinLat = &minLat
		}
	}

	if maxLatStr := c.Query("max_lat"); maxLatStr != "" {
		if maxLat, err := strconv.ParseFloat(maxLatStr, 64); err == nil {
			query.MaxLat = &maxLat
		}
	}

	if minLngStr := c.Query("min_lng"); minLngStr != "" {
		if minLng, err := strconv.ParseFloat(minLngStr, 64); err == nil {
			query.MinLng = &minLng
		}
	}

	if maxLngStr := c.Query("max_lng"); maxLngStr != "" {
		if maxLng, err := strconv.ParseFloat(maxLngStr, 64); err == nil {
			query.MaxLng = &maxLng
		}
	}

	points, total, err := h.service.GetLocationPoints(userID, query)
	if err != nil {
		if err.Error() == "trajectory not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Trajectory not found"})
			return
		}
		h.logger.Error("Failed to get location points", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get location points"})
		return
	}

	responses := make([]models.LocationPointResponse, len(points))
	for i, point := range points {
		responses[i] = point.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"total": total,
		"page":  (query.Offset / query.Limit) + 1,
		"limit": query.Limit,
	})
}

// GetTrajectoryStats 获取轨迹统计信息
// @Summary 获取轨迹统计信息
// @Description 获取用户的轨迹统计信息
// @Tags trajectories
// @Accept json
// @Produce json
// @Success 200 {object} models.TrajectoryStats
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/trajectories/stats [get]
func (h *LocationHandler) GetTrajectoryStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	stats, err := h.service.GetTrajectoryStats(userID)
	if err != nil {
		h.logger.Error("Failed to get trajectory stats", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trajectory stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SyncData 数据同步
// @Summary 数据同步
// @Description 同步轨迹数据
// @Tags sync
// @Accept json
// @Produce json
// @Param request body models.SyncRequest true "同步请求"
// @Success 200 {object} models.SyncResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/sync [post]
func (h *LocationHandler) SyncData(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	response, err := h.service.SyncData(userID, req)
	if err != nil {
		h.logger.Error("Failed to sync data", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync data"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查位置跟踪服务状态
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/location/health [get]
func (h *LocationHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "location-tracking",
		"timestamp": time.Now().Unix(),
	})
}