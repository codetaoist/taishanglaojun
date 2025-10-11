package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/security/services"
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// ThreatDetectionHandler 威胁检测处理器
type ThreatDetectionHandler struct {
	threatService *services.ThreatDetectionService
}

// NewThreatDetectionHandler 创建威胁检测处理器
func NewThreatDetectionHandler(threatService *services.ThreatDetectionService) *ThreatDetectionHandler {
	return &ThreatDetectionHandler{
		threatService: threatService,
	}
}

// CreateThreatAlert 创建威胁告警
// @Summary 创建威胁告警
// @Description 创建新的威胁告警
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param alert body models.ThreatAlert true "威胁告警信息"
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

// GetThreatAlerts 获取威胁告警列表
// @Summary 获取威胁告警列表
// @Description 获取威胁告警列表，支持分页和筛�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param severity query string false "严重级别"
// @Param status query string false "状�?
// @Param start_time query string false "开始时�?
// @Param end_time query string false "结束时间"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/threats/alerts [get]
func (h *ThreatDetectionHandler) GetThreatAlerts(c *gin.Context) {
	// 解析查询参数
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
	
	// 解析时间参数
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

// GetThreatAlert 获取威胁告警详情
// @Summary 获取威胁告警详情
// @Description 根据ID获取威胁告警详情
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
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

// UpdateThreatAlert 更新威胁告警
// @Summary 更新威胁告警
// @Description 更新威胁告警状态或其他信息
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
// @Param updates body map[string]interface{} true "更新内容"
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

// DeleteThreatAlert 删除威胁告警
// @Summary 删除威胁告警
// @Description 根据ID删除威胁告警
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param id path string true "告警ID"
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

// CreateDetectionRule 创建检测规�?
// @Summary 创建检测规�?
// @Description 创建新的威胁检测规�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param rule body models.DetectionRule true "检测规则信�?
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

// GetDetectionRules 获取检测规则列�?
// @Summary 获取检测规则列�?
// @Description 获取威胁检测规则列�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param enabled query bool false "是否启用"
// @Param category query string false "规则类别"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
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

// GetDetectionRule 获取检测规则详�?
// @Summary 获取检测规则详�?
// @Description 根据ID获取检测规则详�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
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

// UpdateDetectionRule 更新检测规�?
// @Summary 更新检测规�?
// @Description 更新检测规则信�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
// @Param updates body map[string]interface{} true "更新内容"
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

// DeleteDetectionRule 删除检测规�?
// @Summary 删除检测规�?
// @Description 根据ID删除检测规�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
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

// AnalyzeSecurityEvent 分析安全事件
// @Summary 分析安全事件
// @Description 对安全事件进行威胁分�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Param event body map[string]interface{} true "安全事件数据"
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

// GetThreatStats 获取威胁统计信息
// @Summary 获取威胁统计信息
// @Description 获取威胁检测相关的统计信息
// @Tags 威胁检�?
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

// StartThreatDetection 启动威胁检�?
// @Summary 启动威胁检�?
// @Description 启动威胁检测服�?
// @Tags 威胁检�?
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

// StopThreatDetection 停止威胁检�?
// @Summary 停止威胁检�?
// @Description 停止威胁检测服�?
// @Tags 威胁检�?
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

// GetThreatDetectionStatus 获取威胁检测状�?
// @Summary 获取威胁检测状�?
// @Description 获取威胁检测服务的运行状�?
// @Tags 威胁检�?
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/security/threats/status [get]
func (h *ThreatDetectionHandler) GetThreatDetectionStatus(c *gin.Context) {
	// 这里可以添加获取服务状态的逻辑
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now(),
	})
}
