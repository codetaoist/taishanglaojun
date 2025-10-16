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

// LocationHandler ?
type LocationHandler struct {
	locationService *services.LocationService
	logger          *zap.Logger
}

// NewLocationHandler 
func NewLocationHandler(service *services.LocationService, logger *zap.Logger) *LocationHandler {
	return &LocationHandler{
		locationService: service,
		logger:          logger,
	}
}

// CreateTrajectory 
// @Summary 
// @Description 
// @Tags trajectories
// @Accept json
// @Produce json
// @Param request body models.TrajectoryRequest true ""
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

	trajectory, err := h.locationService.CreateTrajectory(userID, req)
	if err != nil {
		h.logger.Error("Failed to create trajectory", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create trajectory"})
		return
	}

	c.JSON(http.StatusCreated, trajectory.ToResponse())
}

// GetTrajectories 
// @Summary 
// @Description ?
// @Tags trajectories
// @Accept json
// @Produce json
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Param name query string false ""
// @Param is_active query bool false ""
// @Param start_time query int64 false ""
// @Param end_time query int64 false "?
// @Param order_by query string false "" default("created_at")
// @Param order query string false "" default("desc")
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

	// 
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

	trajectories, total, err := h.locationService.GetTrajectories(userID, query)
	if err != nil {
		h.logger.Error("Failed to get trajectories", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trajectories"})
		return
	}

	// ?
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

// GetTrajectory 
// @Summary 
// @Description ?
// @Tags trajectories
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param include_points query bool false "? default(false)
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

	trajectory, err := h.locationService.GetTrajectory(userID, trajectoryID, includePoints)
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

// UpdateTrajectory 
// @Summary 
// @Description 
// @Tags trajectories
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body models.TrajectoryUpdateRequest true ""
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

	trajectory, err := h.locationService.UpdateTrajectory(userID, trajectoryID, req)
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

// DeleteTrajectory 
// @Summary 
// @Description 
// @Tags trajectories
// @Accept json
// @Produce json
// @Param id path string true "ID"
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

	err := h.locationService.DeleteTrajectory(userID, trajectoryID)
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

// AddLocationPoint ?
// @Summary ?
// @Description 
// @Tags location-points
// @Accept json
// @Produce json
// @Param request body models.LocationPointRequest true "?
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

	point, err := h.locationService.AddLocationPoint(userID, req)
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

// BatchUploadPoints ?
func (h *LocationHandler) BatchUploadPoints(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.LocationPointBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	points, err := h.locationService.AddLocationPointsBatch(userID, req)
	if err != nil {
		h.logger.Error("Failed to batch upload points", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload points"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"points": points})
}

// FinishTrajectory 
func (h *LocationHandler) FinishTrajectory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trajectoryID := c.Param("id")
	if trajectoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trajectory ID is required"})
		return
	}

	// 
	isActive := false
	updateReq := models.TrajectoryUpdateRequest{
		IsActive: &isActive,
	}

	trajectory, err := h.locationService.UpdateTrajectory(userID, trajectoryID, updateReq)
	if err != nil {
		h.logger.Error("Failed to finish trajectory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finish trajectory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"trajectory": trajectory})
}

// GetTrajectoryPoints 
func (h *LocationHandler) GetTrajectoryPoints(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	trajectoryID := c.Param("id")
	if trajectoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trajectory ID is required"})
		return
	}

	// 
	query := models.LocationPointQuery{
		TrajectoryID: trajectoryID,
		Limit:        100, // 
		Offset:       0,
	}

	// 
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	points, total, err := h.locationService.GetLocationPoints(userID, query)
	if err != nil {
		h.logger.Error("Failed to get trajectory points", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get points"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"points": points,
		"total":  total,
		"limit":  query.Limit,
		"offset": query.Offset,
	})
}

// GetSyncStatus ?
func (h *LocationHandler) GetSyncStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 
	// ?
	status := gin.H{
		"user_id":     userID,
		"last_sync":   time.Now().Unix(),
		"sync_status": "up_to_date",
		"pending_uploads": 0,
	}

	c.JSON(http.StatusOK, status)
}

// GetTrajectoryStats 
// @Summary 
// @Description ?
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

	stats, err := h.locationService.GetTrajectoryStats(userID)
	if err != nil {
		h.logger.Error("Failed to get trajectory stats", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trajectory stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SyncData 
// @Summary 
// @Description 
// @Tags sync
// @Accept json
// @Produce json
// @Param request body models.SyncRequest true ""
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

	response, err := h.locationService.SyncData(userID, req)
	if err != nil {
		h.logger.Error("Failed to sync data", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync data"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck ?
// @Summary ?
// @Description ?
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

// AddLocationPointsBatch ?
func (h *LocationHandler) AddLocationPointsBatch(c *gin.Context) {
	h.BatchUploadPoints(c)
}

// GetLocationPoints ?
func (h *LocationHandler) GetLocationPoints(c *gin.Context) {
	h.GetTrajectoryPoints(c)
}

// UploadLocationPoints ?
func (h *LocationHandler) UploadLocationPoints(c *gin.Context) {
	h.BatchUploadPoints(c)
}

// DeleteLocationPoint ?
func (h *LocationHandler) DeleteLocationPoint(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	
	pointID := c.Param("point_id")

	err := h.locationService.DeleteLocationPoint(userID.(string), pointID)
	if err != nil {
		h.logger.Error("Failed to delete location point", 
			zap.String("user_id", userID.(string)),
			zap.String("point_id", pointID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete location point"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Location point deleted successfully"})
}

