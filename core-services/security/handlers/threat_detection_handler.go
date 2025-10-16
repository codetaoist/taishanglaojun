package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/security/services"
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// ThreatDetectionHandler 
type ThreatDetectionHandler struct {
	threatService *services.ThreatDetectionService
}

// NewThreatDetectionHandler 
func NewThreatDetectionHandler(threatService *services.ThreatDetectionService) *ThreatDetectionHandler {
	return &ThreatDetectionHandler{
		threatService: threatService,
	}
}

// CreateThreatAlert 澯
// @Summary 澯
// @Description 澯
// @Tags ?
// @Accept json
// @Produce json
// @Param alert body models.ThreatAlert true "澯"
// @Success 201 {object} models.ThreatAlert
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/alerts [post]
func (h *ThreatDetectionHandler) CreateThreatAlert(c *gin.Context) {
	var alert models.ThreatAlert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.threatService.CreateThreatAlert(c.Request.Context(), &alert); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create threat alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

// GetThreatAlerts 澯
// @Summary 澯
// @Description 澯?
// @Tags ?
// @Accept json
// @Produce json
// @Param severity query string false ""
// @Param status query string false "?
// @Param start_time query string false "?
// @Param end_time query string false ""
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/alerts [get]
func (h *ThreatDetectionHandler) GetThreatAlerts(c *gin.Context) {
	// 
	severity := c.Query("severity")
	status := c.Query("status")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	// 
	var startTime, endTime time.Time
	var err error
	
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid start_time format",
				"details": "Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)",
			})
			return
		}
	}
	
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid end_time format",
				"details": "Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)",
			})
			return
		}
	}

	alerts, err := h.threatService.GetThreatAlerts(c.Request.Context(), severity, status, startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get threat alerts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(alerts),
		},
	})
}

// GetThreatAlert 澯
// @Summary 澯
// @Description ID澯
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "澯ID"
// @Success 200 {object} models.ThreatAlert
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/alerts/{id} [get]
func (h *ThreatDetectionHandler) GetThreatAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	alert, err := h.threatService.GetThreatAlert(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Threat alert not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// UpdateThreatAlert 澯
// @Summary 澯
// @Description 澯
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "澯ID"
// @Param updates body map[string]interface{} true ""
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/alerts/{id} [put]
func (h *ThreatDetectionHandler) UpdateThreatAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.threatService.UpdateThreatAlert(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update threat alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Threat alert updated successfully",
	})
}

// DeleteThreatAlert 澯
// @Summary 澯
// @Description ID澯
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "澯ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/alerts/{id} [delete]
func (h *ThreatDetectionHandler) DeleteThreatAlert(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alert ID is required",
		})
		return
	}

	if err := h.threatService.DeleteThreatAlert(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete threat alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Threat alert deleted successfully",
	})
}

// CreateDetectionRule ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Param rule body models.DetectionRule true "?
// @Success 201 {object} models.DetectionRule
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/rules [post]
func (h *ThreatDetectionHandler) CreateDetectionRule(c *gin.Context) {
	var rule models.DetectionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.threatService.CreateDetectionRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create detection rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetDetectionRules ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Param enabled query bool false ""
// @Param category query string false ""
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/rules [get]
func (h *ThreatDetectionHandler) GetDetectionRules(c *gin.Context) {
	enabledStr := c.Query("enabled")
	category := c.Query("category")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit
	
	var enabled *bool
	if enabledStr != "" {
		e, err := strconv.ParseBool(enabledStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid enabled parameter",
				"details": "Must be true or false",
			})
			return
		}
		enabled = &e
	}

	rules, err := h.threatService.GetDetectionRules(c.Request.Context(), enabled, category, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get detection rules",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rules,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(rules),
		},
	})
}

// GetDetectionRule ?
// @Summary ?
// @Description ID?
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.DetectionRule
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/rules/{id} [get]
func (h *ThreatDetectionHandler) GetDetectionRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
		})
		return
	}

	rule, err := h.threatService.GetDetectionRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Detection rule not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateDetectionRule ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param updates body map[string]interface{} true ""
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/rules/{id} [put]
func (h *ThreatDetectionHandler) UpdateDetectionRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.threatService.UpdateDetectionRule(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update detection rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Detection rule updated successfully",
	})
}

// DeleteDetectionRule ?
// @Summary ?
// @Description ID?
// @Tags ?
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/rules/{id} [delete]
func (h *ThreatDetectionHandler) DeleteDetectionRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
		})
		return
	}

	if err := h.threatService.DeleteDetectionRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete detection rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Detection rule deleted successfully",
	})
}

// AnalyzeSecurityEvent 
// @Summary 
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Param event body map[string]interface{} true ""
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/analyze [post]
func (h *ThreatDetectionHandler) AnalyzeSecurityEvent(c *gin.Context) {
	var eventData map[string]interface{}
	if err := c.ShouldBindJSON(&eventData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := h.threatService.AnalyzeSecurityEvent(c.Request.Context(), eventData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to analyze security event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis_result": result,
	})
}

// GetThreatStats 
// @Summary 
// @Description 
// @Tags ?
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/stats [get]
func (h *ThreatDetectionHandler) GetThreatStats(c *gin.Context) {
	stats, err := h.threatService.GetThreatStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get threat statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// StartThreatDetection ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/start [post]
func (h *ThreatDetectionHandler) StartThreatDetection(c *gin.Context) {
	h.threatService.Start()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Threat detection service started successfully",
	})
}

// StopThreatDetection ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/stop [post]
func (h *ThreatDetectionHandler) StopThreatDetection(c *gin.Context) {
	h.threatService.Stop()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Threat detection service stopped successfully",
	})
}

// GetThreatDetectionStatus ?
// @Summary ?
// @Description ?
// @Tags ?
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/security/threats/status [get]
func (h *ThreatDetectionHandler) GetThreatDetectionStatus(c *gin.Context) {
	// 
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now(),
	})
}

