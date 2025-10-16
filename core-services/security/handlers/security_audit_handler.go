package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/security/services"
	"github.com/codetaoist/taishanglaojun/core-services/security/models"
)

// SecurityAuditHandler ?
type SecurityAuditHandler struct {
	auditService *services.SecurityAuditService
}

// NewSecurityAuditHandler ?
func NewSecurityAuditHandler(auditService *services.SecurityAuditService) *SecurityAuditHandler {
	return &SecurityAuditHandler{
		auditService: auditService,
	}
}

// LogAuditEvent 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param event body models.AuditLog true ""
// @Success 201 {object} models.AuditLog
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/events [post]
func (h *SecurityAuditHandler) LogAuditEvent(c *gin.Context) {
	var auditLog models.AuditLog
	if err := c.ShouldBindJSON(&auditLog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 
	if auditLog.Action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Action is required",
		})
		return
	}

	if auditLog.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	if err := h.auditService.LogAuditEvent(c.Request.Context(), &auditLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to log audit event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, auditLog)
}

// GetAuditLogs 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param user_id query string false "ID"
// @Param action query string false ""
// @Param resource query string false ""
// @Param start_time query string false "?
// @Param end_time query string false ""
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/logs [get]
func (h *SecurityAuditHandler) GetAuditLogs(c *gin.Context) {
	userID := c.Query("user_id")
	action := c.Query("action")
	resource := c.Query("resource")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit

	logs, err := h.auditService.GetAuditLogs(c.Request.Context(), userID, action, resource, startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get audit logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(logs),
		},
	})
}

// GetAuditLog 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.AuditLog
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/logs/{id} [get]
func (h *SecurityAuditHandler) GetAuditLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Audit log ID is required",
		})
		return
	}

	log, err := h.auditService.GetAuditLog(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Audit log not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, log)
}

// DeleteAuditLog 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/logs/{id} [delete]
func (h *SecurityAuditHandler) DeleteAuditLog(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Audit log ID is required",
		})
		return
	}

	if err := h.auditService.DeleteAuditLog(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete audit log",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Audit log deleted successfully",
	})
}

// LogSecurityEvent 
// @Summary 
// @Description İ
// @Tags 
// @Accept json
// @Produce json
// @Param event body models.SecurityEvent true ""
// @Success 201 {object} models.SecurityEvent
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/security-events [post]
func (h *SecurityAuditHandler) LogSecurityEvent(c *gin.Context) {
	var securityEvent models.SecurityEvent
	if err := c.ShouldBindJSON(&securityEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 
	if securityEvent.EventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Event type is required",
		})
		return
	}

	if securityEvent.Severity == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Severity is required",
		})
		return
	}

	if err := h.auditService.LogSecurityEvent(c.Request.Context(), &securityEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to log security event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, securityEvent)
}

// GetSecurityEvents 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param event_type query string false ""
// @Param severity query string false ""
// @Param status query string false "?
// @Param start_time query string false "?
// @Param end_time query string false ""
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/security-events [get]
func (h *SecurityAuditHandler) GetSecurityEvents(c *gin.Context) {
	eventType := c.Query("event_type")
	severity := c.Query("severity")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit

	events, err := h.auditService.GetSecurityEvents(c.Request.Context(), eventType, severity, status, startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get security events",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(events),
		},
	})
}

// GetSecurityEvent 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} models.SecurityEvent
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/security-events/{id} [get]
func (h *SecurityAuditHandler) GetSecurityEvent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Security event ID is required",
		})
		return
	}

	event, err := h.auditService.GetSecurityEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Security event not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, event)
}

// UpdateSecurityEvent 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param updates body map[string]interface{} true ""
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/security-events/{id} [put]
func (h *SecurityAuditHandler) UpdateSecurityEvent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Security event ID is required",
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

	if err := h.auditService.UpdateSecurityEvent(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update security event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Security event updated successfully",
	})
}

// DeleteSecurityEvent 
// @Summary 
// @Description ID
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/security-events/{id} [delete]
func (h *SecurityAuditHandler) DeleteSecurityEvent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Security event ID is required",
		})
		return
	}

	if err := h.auditService.DeleteSecurityEvent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete security event",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Security event deleted successfully",
	})
}

// RunComplianceCheck ?
// @Summary ?
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param check body map[string]interface{} true "?
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/compliance/check [post]
func (h *SecurityAuditHandler) RunComplianceCheck(c *gin.Context) {
	var checkParams map[string]interface{}
	if err := c.ShouldBindJSON(&checkParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	report, err := h.auditService.RunComplianceCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to run compliance check",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Compliance check completed",
		"report":  report,
	})
}

// GetComplianceReports 汨
// @Summary 汨
// @Description 鱨?
// @Tags 
// @Accept json
// @Produce json
// @Param standard query string false ""
// @Param status query string false "?
// @Param start_time query string false "?
// @Param end_time query string false ""
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/compliance/reports [get]
func (h *SecurityAuditHandler) GetComplianceReports(c *gin.Context) {
	standard := c.Query("standard")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	
	offset := (page - 1) * limit

	reports, err := h.auditService.GetComplianceReports(c.Request.Context(), standard, status, startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get compliance reports",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": reports,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(reports),
		},
	})
}

// GetComplianceReport 汨
// @Summary 汨
// @Description ID汨
// @Tags 
// @Accept json
// @Produce json
// @Param id path string true "汨ID"
// @Success 200 {object} models.ComplianceReport
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/compliance/reports/{id} [get]
func (h *SecurityAuditHandler) GetComplianceReport(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Compliance report ID is required",
		})
		return
	}

	report, err := h.auditService.GetComplianceReport(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Compliance report not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetAuditStats 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/stats [get]
func (h *SecurityAuditHandler) GetAuditStats(c *gin.Context) {
	stats, err := h.auditService.GetAuditStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get audit statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ExportAuditLogs 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param format query string false "" default(json)
// @Param user_id query string false "ID"
// @Param action query string false ""
// @Param start_time query string false "?
// @Param end_time query string false ""
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/logs/export [get]
func (h *SecurityAuditHandler) ExportAuditLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	userID := c.Query("user_id")
	action := c.Query("action")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 
	logs, err := h.auditService.GetAuditLogs(c.Request.Context(), userID, action, "", startTime, endTime, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get audit logs for export",
			"details": err.Error(),
		})
		return
	}

	// 
	switch format {
	case "json":
		c.Header("Content-Disposition", "attachment; filename=audit_logs.json")
		c.JSON(http.StatusOK, gin.H{
			"export_time": time.Now(),
			"total_count": len(logs),
			"audit_logs":  logs,
		})
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
		
		// CSV?
		c.String(http.StatusOK, "CSV export not implemented yet")
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported export format",
		})
	}
}

// ExportSecurityEvents 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param format query string false "" default(json)
// @Param event_type query string false ""
// @Param severity query string false ""
// @Param start_time query string false "?
// @Param end_time query string false ""
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/security-events/export [get]
func (h *SecurityAuditHandler) ExportSecurityEvents(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	eventType := c.Query("event_type")
	severity := c.Query("severity")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 
	events, err := h.auditService.GetSecurityEvents(c.Request.Context(), eventType, severity, "", startTime, endTime, 10000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get security events for export",
			"details": err.Error(),
		})
		return
	}

	// 
	switch format {
	case "json":
		c.Header("Content-Disposition", "attachment; filename=security_events.json")
		c.JSON(http.StatusOK, gin.H{
			"export_time":     time.Now(),
			"total_count":     len(events),
			"security_events": events,
		})
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=security_events.csv")
		
		// CSV?
		c.String(http.StatusOK, "CSV export not implemented yet")
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported export format",
		})
	}
}

// StartAuditService 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/service/start [post]
func (h *SecurityAuditHandler) StartAuditService(c *gin.Context) {
	h.auditService.Start()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Security audit service started successfully",
	})
}

// StopAuditService 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/security/audit/service/stop [post]
func (h *SecurityAuditHandler) StopAuditService(c *gin.Context) {
	h.auditService.Stop()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Security audit service stopped successfully",
	})
}

// GetServiceStatus ?
// @Summary ?
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/security/audit/service/status [get]
func (h *SecurityAuditHandler) GetServiceStatus(c *gin.Context) {
	// 
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now(),
	})
}

