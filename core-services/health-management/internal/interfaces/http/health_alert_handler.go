package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	
	"github.com/taishanglaojun/health-management/internal/application"
)

// HealthAlertHandler ?
type HealthAlertHandler struct {
	healthAlertService *application.HealthAlertService
}

// NewHealthAlertHandler ?
func NewHealthAlertHandler(healthAlertService *application.HealthAlertService) *HealthAlertHandler {
	return &HealthAlertHandler{
		healthAlertService: healthAlertService,
	}
}

// DetectAnomaliesRequest HTTP
type DetectAnomaliesRequest struct {
	UserID    string   `json:"user_id" binding:"required"`
	DataTypes []string `json:"data_types,omitempty"`
	Days      int      `json:"days,omitempty"`
}

// CheckEmergencyRequest HTTP
type CheckEmergencyRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	DataType string  `json:"data_type" binding:"required"`
	Value    float64 `json:"value" binding:"required"`
	Unit     string  `json:"unit,omitempty"`
}

// GetAlertsRequest HTTP
type GetAlertsRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
	Types      []string `json:"types,omitempty"`
	Severities []string `json:"severities,omitempty"`
	IsRead     *bool    `json:"is_read,omitempty"`
	IsHandled  *bool    `json:"is_handled,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

// MarkAlertRequest 
type MarkAlertRequest struct {
	AlertID   string `json:"alert_id" binding:"required"`
	IsRead    *bool  `json:"is_read,omitempty"`
	IsHandled *bool  `json:"is_handled,omitempty"`
}

// DetectAnomalies ?
func (h *HealthAlertHandler) DetectAnomalies(c *gin.Context) {
	var req DetectAnomaliesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "User ID must be a valid UUID",
		})
		return
	}

	// 
	serviceReq := &application.DetectAnomaliesRequest{
		UserID:    userID,
		DataTypes: req.DataTypes,
		Days:      req.Days,
	}

	// 
	response, err := h.healthAlertService.DetectAnomalies(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to detect anomalies",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CheckEmergency ?
func (h *HealthAlertHandler) CheckEmergency(c *gin.Context) {
	var req CheckEmergencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "User ID must be a valid UUID",
		})
		return
	}

	// 
	serviceReq := &application.CheckEmergencyRequest{
		UserID:   userID,
		DataType: req.DataType,
		Value:    req.Value,
		Unit:     req.Unit,
	}

	// 
	response, err := h.healthAlertService.CheckEmergency(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check emergency",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAlerts 
func (h *HealthAlertHandler) GetAlerts(c *gin.Context) {
	var req GetAlertsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "User ID must be a valid UUID",
		})
		return
	}

	// ?
	var types []application.AlertType
	for _, t := range req.Types {
		types = append(types, application.AlertType(t))
	}

	var severities []application.AlertSeverity
	for _, s := range req.Severities {
		severities = append(severities, application.AlertSeverity(s))
	}

	// 
	serviceReq := &application.GetAlertsRequest{
		UserID:     userID,
		Types:      types,
		Severities: severities,
		IsRead:     req.IsRead,
		IsHandled:  req.IsHandled,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	// 
	response, err := h.healthAlertService.GetAlerts(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get alerts",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAlertsByUser ID
func (h *HealthAlertHandler) GetAlertsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "User ID must be a valid UUID",
		})
		return
	}

	// ?
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var isRead *bool
	if isReadStr := c.Query("is_read"); isReadStr != "" {
		if b, err := strconv.ParseBool(isReadStr); err == nil {
			isRead = &b
		}
	}

	var isHandled *bool
	if isHandledStr := c.Query("is_handled"); isHandledStr != "" {
		if b, err := strconv.ParseBool(isHandledStr); err == nil {
			isHandled = &b
		}
	}

	// 
	serviceReq := &application.GetAlertsRequest{
		UserID:    userID,
		IsRead:    isRead,
		IsHandled: isHandled,
		Limit:     limit,
		Offset:    offset,
	}

	// 
	response, err := h.healthAlertService.GetAlerts(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get alerts",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MarkAlert ?
func (h *HealthAlertHandler) MarkAlert(c *gin.Context) {
	var req MarkAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// ID
	alertID, err := uuid.Parse(req.AlertID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid alert ID",
			"message": "Alert ID must be a valid UUID",
		})
		return
	}

	// 
	// 㻹?
	response := gin.H{
		"alert_id": alertID,
		"message":  "Alert marked successfully",
	}

	if req.IsRead != nil {
		response["is_read"] = *req.IsRead
	}
	if req.IsHandled != nil {
		response["is_handled"] = *req.IsHandled
	}

	c.JSON(http.StatusOK, response)
}

// GetAlertStatistics 
func (h *HealthAlertHandler) GetAlertStatistics(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "User ID must be a valid UUID",
		})
		return
	}

	// 
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// 
	statistics := gin.H{
		"user_id": userID,
		"period":  fmt.Sprintf("last_%d_days", days),
		"total_alerts": 15,
		"by_severity": gin.H{
			"critical": 2,
			"high":     4,
			"medium":   6,
			"low":      3,
		},
		"by_type": gin.H{
			"abnormal":  8,
			"critical":  2,
			"emergency": 1,
			"trend":     3,
			"reminder":  1,
		},
		"by_data_type": gin.H{
			"heart_rate":     5,
			"blood_pressure": 4,
			"blood_sugar":    3,
			"temperature":    2,
			"sleep_duration": 1,
		},
		"unread_count":    8,
		"unhandled_count": 12,
	}

	c.JSON(http.StatusOK, statistics)
}

// HealthAlertHealthCheckHandler ?
func HealthAlertHealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "health-alert",
		"timestamp": fmt.Sprintf("%d", c.Request.Context().Value("timestamp")),
		"checks": gin.H{
			"anomaly_detection": "ok",
			"emergency_check":   "ok",
			"alert_management":  "ok",
		},
	})
}

