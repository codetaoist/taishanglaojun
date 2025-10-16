package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/audit"
)

// AuditHandlers 
type AuditHandlers struct {
	service audit.AuditService
	logger  *zap.Logger
}

// NewAuditHandlers 
func NewAuditHandlers(service audit.AuditService, logger *zap.Logger) *AuditHandlers {
	return &AuditHandlers{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 
func (h *AuditHandlers) RegisterRoutes(router *gin.RouterGroup) {
	audit := router.Group("/audit")
	{
		// 
		audit.GET("/logs", h.QueryLogs)
		audit.GET("/logs/:id", h.GetLog)

		// 
		audit.GET("/statistics", h.GetStatistics)
		audit.GET("/statistics/users", h.GetUserActivity)
		audit.GET("/statistics/resources", h.GetResourceActivity)
		audit.GET("/statistics/security", h.GetSecurityStatistics)

		// 
		audit.POST("/export", h.ExportLogs)
		audit.GET("/export/:id", h.GetExportStatus)
		audit.GET("/export/:id/download", h.DownloadExport)

		// 
		audit.DELETE("/cleanup", h.CleanupLogs)
		audit.GET("/health", h.HealthCheck)
		audit.GET("/stats", h.GetServiceStats)
	}
}

// QueryLogs 
func (h *AuditHandlers) QueryLogs(c *gin.Context) {
	// 
	query, err := h.parseAuditQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// 
	response, err := h.service.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to query audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLog 
func (h *AuditHandlers) GetLog(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Log ID is required",
		})
		return
	}

	// 
	query := &audit.AuditQuery{
		EventIDs: []string{logID},
		Pagination: audit.PaginationRequest{
			Page:     1,
			PageSize: 1,
		},
	}

	// 
	response, err := h.service.QueryLogs(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get audit log", zap.String("log_id", logID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get audit log",
		})
		return
	}

	if len(response.Events) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Audit log not found",
		})
		return
	}

	c.JSON(http.StatusOK, response.Events[0])
}

// GetStatistics 
func (h *AuditHandlers) GetStatistics(c *gin.Context) {
	// 
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 
	stats, err := h.service.GetStatistics(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get audit statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get audit statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserActivity 
func (h *AuditHandlers) GetUserActivity(c *gin.Context) {
	// 
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 
	stats, err := h.service.GetStatistics(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get user activity statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user activity statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_activities": stats.UserActivities,
		"total_users":     len(stats.UserActivities),
	})
}

// GetResourceActivity 
func (h *AuditHandlers) GetResourceActivity(c *gin.Context) {
	// 
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 
	stats, err := h.service.GetStatistics(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get resource activity statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get resource activity statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resource_activities": stats.ResourceActivities,
		"total_resources":     len(stats.ResourceActivities),
	})
}

// GetSecurityStatistics 
func (h *AuditHandlers) GetSecurityStatistics(c *gin.Context) {
	// 
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 
	stats, err := h.service.GetStatistics(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get security statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get security statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats.SecurityStats)
}

// ExportLogs 
func (h *AuditHandlers) ExportLogs(c *gin.Context) {
	var request audit.ExportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 
	if err := h.validateExportRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid export request",
			"details": err.Error(),
		})
		return
	}

	// 
	response, err := h.service.ExportLogs(c.Request.Context(), &request)
	if err != nil {
		h.logger.Error("Failed to start export", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start export",
		})
		return
	}

	c.JSON(http.StatusAccepted, response)
}

// GetExportStatus 
func (h *AuditHandlers) GetExportStatus(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Export ID is required",
		})
		return
	}

	// 
	// 
	c.JSON(http.StatusOK, gin.H{
		"export_id": exportID,
		"status":    "processing",
		"message":   "Export status check not implemented",
	})
}

// DownloadExport 
func (h *AuditHandlers) DownloadExport(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Export ID is required",
		})
		return
	}

	// 
	// 
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Download not implemented",
		"message": "Export download functionality not implemented",
	})
}

// CleanupLogs 
func (h *AuditHandlers) CleanupLogs(c *gin.Context) {
	// 
	beforeStr := c.Query("before")
	if beforeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter 'before' is required",
		})
		return
	}

	before, err := time.Parse(time.RFC3339, beforeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid 'before' parameter",
			"details": "Expected RFC3339 format",
		})
		return
	}

	// 创建保留策略
	retentionPolicy := &audit.RetentionPolicy{
		DefaultRetention: time.Since(before),
	}
	count, err := h.service.CleanupLogs(c.Request.Context(), retentionPolicy)
	if err != nil {
		h.logger.Error("Failed to cleanup audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cleanup audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Cleanup completed successfully",
		"deleted_count": count,
	})
}

// HealthCheck 
func (h *AuditHandlers) HealthCheck(c *gin.Context) {
	health := h.service.HealthCheck(c.Request.Context())

	status := http.StatusOK
	if !health.Healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, health)
}

// GetServiceStats 
func (h *AuditHandlers) GetServiceStats(c *gin.Context) {
	stats := h.service.GetStats()
	c.JSON(http.StatusOK, stats)
}

// 

// parseAuditQuery 
func (h *AuditHandlers) parseAuditQuery(c *gin.Context) (*audit.AuditQuery, error) {
	query := &audit.AuditQuery{}

	// 
	if eventTypes := c.Query("event_types"); eventTypes != "" {
		types := strings.Split(eventTypes, ",")
		for _, t := range types {
			query.EventTypes = append(query.EventTypes, audit.EventType(strings.TrimSpace(t)))
		}
	}

	if eventActions := c.Query("event_actions"); eventActions != "" {
		query.EventActions = strings.Split(eventActions, ",")
	}

	if eventCategories := c.Query("event_categories"); eventCategories != "" {
		query.EventCategories = strings.Split(eventCategories, ",")
	}

	// 
	if userIDs := c.Query("user_ids"); userIDs != "" {
		query.UserIDs = strings.Split(userIDs, ",")
	}

	if userNames := c.Query("user_names"); userNames != "" {
		query.UserNames = strings.Split(userNames, ",")
	}

	if userRoles := c.Query("user_roles"); userRoles != "" {
		query.UserRoles = strings.Split(userRoles, ",")
	}

	// 
	if tenantIDs := c.Query("tenant_ids"); tenantIDs != "" {
		query.TenantIDs = strings.Split(tenantIDs, ",")
	}

	// 
	if resourceIDs := c.Query("resource_ids"); resourceIDs != "" {
		query.ResourceIDs = strings.Split(resourceIDs, ",")
	}

	if resourceTypes := c.Query("resource_types"); resourceTypes != "" {
		query.ResourceTypes = strings.Split(resourceTypes, ",")
	}

	// 
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format: %v", err)
		}
		query.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format: %v", err)
		}
		query.EndTime = &endTime
	}

	// 
	if securityLevels := c.Query("security_levels"); securityLevels != "" {
		levels := strings.Split(securityLevels, ",")
		for _, level := range levels {
			query.SecurityLevels = append(query.SecurityLevels, audit.SecurityLevel(strings.TrimSpace(level)))
		}
	}

	// 
	if minRiskStr := c.Query("min_risk_score"); minRiskStr != "" {
		minRisk, err := strconv.ParseFloat(minRiskStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid min_risk_score: %v", err)
		}
		query.MinRiskScore = &minRisk
	}

	if maxRiskStr := c.Query("max_risk_score"); maxRiskStr != "" {
		maxRisk, err := strconv.ParseFloat(maxRiskStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid max_risk_score: %v", err)
		}
		query.MaxRiskScore = &maxRisk
	}

	// 
	if ipAddresses := c.Query("ip_addresses"); ipAddresses != "" {
		query.IPAddresses = strings.Split(ipAddresses, ",")
	}

	// 
	if requestMethods := c.Query("request_methods"); requestMethods != "" {
		query.RequestMethods = strings.Split(requestMethods, ",")
	}

	if responseStatuses := c.Query("response_statuses"); responseStatuses != "" {
		statusStrs := strings.Split(responseStatuses, ",")
		for _, statusStr := range statusStrs {
			status, err := strconv.Atoi(strings.TrimSpace(statusStr))
			if err != nil {
				return nil, fmt.Errorf("invalid response_status: %v", err)
			}
			query.ResponseStatuses = append(query.ResponseStatuses, status)
		}
	}

	// 
	if searchTerm := c.Query("search"); searchTerm != "" {
		query.SearchTerm = searchTerm
	}

	// 
	if sortBy := c.Query("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	// 
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return nil, fmt.Errorf("invalid page number")
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 1000 {
			return nil, fmt.Errorf("invalid page_size (must be 1-1000)")
		}
	}

	query.Pagination = audit.PaginationRequest{
		Page:     page,
		PageSize: pageSize,
	}

	return query, nil
}

// parseStatisticsFilter 
func (h *AuditHandlers) parseStatisticsFilter(c *gin.Context) (*audit.StatisticsFilter, error) {
	filter := &audit.StatisticsFilter{}

	// 
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_time format: %v", err)
		}
		filter.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_time format: %v", err)
		}
		filter.EndTime = &endTime
	}

	// 
	if eventTypes := c.Query("event_types"); eventTypes != "" {
		types := strings.Split(eventTypes, ",")
		for _, t := range types {
			filter.EventTypes = append(filter.EventTypes, audit.EventType(strings.TrimSpace(t)))
		}
	}

	// 
	if userIDs := c.Query("user_ids"); userIDs != "" {
		filter.UserIDs = strings.Split(userIDs, ",")
	}

	// 
	if tenantIDs := c.Query("tenant_ids"); tenantIDs != "" {
		filter.TenantIDs = strings.Split(tenantIDs, ",")
	}

	// 
	if resourceTypes := c.Query("resource_types"); resourceTypes != "" {
		filter.ResourceTypes = strings.Split(resourceTypes, ",")
	}

	// 
	if groupBy := c.Query("group_by"); groupBy != "" {
		filter.GroupBy = groupBy
	}

	// 
	if interval := c.Query("interval"); interval != "" {
		filter.Interval = interval
	}

	// 
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return nil, fmt.Errorf("invalid limit")
		}
		filter.Limit = limit
	}

	return filter, nil
}

// validateExportRequest 
func (h *AuditHandlers) validateExportRequest(request *audit.ExportRequest) error {
	// 
	validFormats := map[string]bool{
		"json": true,
		"csv":  true,
		"xlsx": true,
		"pdf":  true,
	}

	if !validFormats[request.Format] {
		return fmt.Errorf("unsupported format: %s", request.Format)
	}

	// 
	if request.Query != nil && request.Query.StartTime != nil && request.Query.EndTime != nil {
		if request.Query.EndTime.Before(*request.Query.StartTime) {
			return fmt.Errorf("end_time must be after start_time")
		}

		// 90
		maxDuration := 90 * 24 * time.Hour
		if request.Query.EndTime.Sub(*request.Query.StartTime) > maxDuration {
			return fmt.Errorf("export time range cannot exceed 90 days")
		}
	}

	return nil
}

// 

// GetCurrentUserID ID
func GetCurrentUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}

// GetCurrentTenantID ID
func GetCurrentTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}

// RequirePermission 
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 
		// 
		c.Next()
	}
}

// RequireRole 
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user role",
			})
			c.Abort()
			return
		}

		// 
		for _, requiredRole := range roles {
			if role == requiredRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

