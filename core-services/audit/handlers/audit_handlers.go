package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"../audit"
)

// AuditHandlers 审计处理器
type AuditHandlers struct {
	service audit.AuditService
	logger  *zap.Logger
}

// NewAuditHandlers 创建审计处理器
func NewAuditHandlers(service audit.AuditService, logger *zap.Logger) *AuditHandlers {
	return &AuditHandlers{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *AuditHandlers) RegisterRoutes(router *gin.RouterGroup) {
	audit := router.Group("/audit")
	{
		// 查询审计日志
		audit.GET("/logs", h.QueryLogs)
		audit.GET("/logs/:id", h.GetLog)
		
		// 统计信息
		audit.GET("/statistics", h.GetStatistics)
		audit.GET("/statistics/users", h.GetUserActivity)
		audit.GET("/statistics/resources", h.GetResourceActivity)
		audit.GET("/statistics/security", h.GetSecurityStatistics)
		
		// 导出功能
		audit.POST("/export", h.ExportLogs)
		audit.GET("/export/:id", h.GetExportStatus)
		audit.GET("/export/:id/download", h.DownloadExport)
		
		// 管理功能
		audit.DELETE("/cleanup", h.CleanupLogs)
		audit.GET("/health", h.HealthCheck)
		audit.GET("/stats", h.GetServiceStats)
	}
}

// QueryLogs 查询审计日志
func (h *AuditHandlers) QueryLogs(c *gin.Context) {
	// 解析查询参数
	query, err := h.parseAuditQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// 查询日志
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

// GetLog 获取单个审计日志
func (h *AuditHandlers) GetLog(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Log ID is required",
		})
		return
	}

	// 构建查询
	query := &audit.AuditQuery{
		EventIDs: []string{logID},
		Pagination: audit.PaginationRequest{
			Page:     1,
			PageSize: 1,
		},
	}

	// 查询日志
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

// GetStatistics 获取统计信息
func (h *AuditHandlers) GetStatistics(c *gin.Context) {
	// 解析统计过滤器
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 获取统计信息
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

// GetUserActivity 获取用户活动统计
func (h *AuditHandlers) GetUserActivity(c *gin.Context) {
	// 解析统计过滤器
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 获取统计信息
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

// GetResourceActivity 获取资源活动统计
func (h *AuditHandlers) GetResourceActivity(c *gin.Context) {
	// 解析统计过滤器
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 获取统计信息
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

// GetSecurityStatistics 获取安全统计
func (h *AuditHandlers) GetSecurityStatistics(c *gin.Context) {
	// 解析统计过滤器
	filter, err := h.parseStatisticsFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid filter parameters",
			"details": err.Error(),
		})
		return
	}

	// 获取统计信息
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

// ExportLogs 导出审计日志
func (h *AuditHandlers) ExportLogs(c *gin.Context) {
	var request audit.ExportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 验证导出请求
	if err := h.validateExportRequest(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid export request",
			"details": err.Error(),
		})
		return
	}

	// 开始导出
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

// GetExportStatus 获取导出状态
func (h *AuditHandlers) GetExportStatus(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Export ID is required",
		})
		return
	}

	// 这里需要实现获取导出状态的逻辑
	// 由于接口中没有定义，我们返回一个简单的响应
	c.JSON(http.StatusOK, gin.H{
		"export_id": exportID,
		"status":    "processing",
		"message":   "Export status check not implemented",
	})
}

// DownloadExport 下载导出文件
func (h *AuditHandlers) DownloadExport(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Export ID is required",
		})
		return
	}

	// 这里需要实现下载导出文件的逻辑
	// 由于接口中没有定义，我们返回一个简单的响应
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Download not implemented",
		"message": "Export download functionality not implemented",
	})
}

// CleanupLogs 清理审计日志
func (h *AuditHandlers) CleanupLogs(c *gin.Context) {
	// 解析清理参数
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

	// 执行清理
	count, err := h.service.CleanupLogs(c.Request.Context(), before)
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

// HealthCheck 健康检查
func (h *AuditHandlers) HealthCheck(c *gin.Context) {
	health := h.service.HealthCheck(c.Request.Context())
	
	status := http.StatusOK
	if !health.Healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, health)
}

// GetServiceStats 获取服务统计
func (h *AuditHandlers) GetServiceStats(c *gin.Context) {
	stats := h.service.GetStats()
	c.JSON(http.StatusOK, stats)
}

// 私有方法

// parseAuditQuery 解析审计查询参数
func (h *AuditHandlers) parseAuditQuery(c *gin.Context) (*audit.AuditQuery, error) {
	query := &audit.AuditQuery{}

	// 基础过滤
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

	// 用户过滤
	if userIDs := c.Query("user_ids"); userIDs != "" {
		query.UserIDs = strings.Split(userIDs, ",")
	}

	if userNames := c.Query("user_names"); userNames != "" {
		query.UserNames = strings.Split(userNames, ",")
	}

	if userRoles := c.Query("user_roles"); userRoles != "" {
		query.UserRoles = strings.Split(userRoles, ",")
	}

	// 租户过滤
	if tenantIDs := c.Query("tenant_ids"); tenantIDs != "" {
		query.TenantIDs = strings.Split(tenantIDs, ",")
	}

	// 资源过滤
	if resourceIDs := c.Query("resource_ids"); resourceIDs != "" {
		query.ResourceIDs = strings.Split(resourceIDs, ",")
	}

	if resourceTypes := c.Query("resource_types"); resourceTypes != "" {
		query.ResourceTypes = strings.Split(resourceTypes, ",")
	}

	// 时间范围
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

	// 安全级别
	if securityLevels := c.Query("security_levels"); securityLevels != "" {
		levels := strings.Split(securityLevels, ",")
		for _, level := range levels {
			query.SecurityLevels = append(query.SecurityLevels, audit.SecurityLevel(strings.TrimSpace(level)))
		}
	}

	// 风险评分范围
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

	// 网络信息
	if ipAddresses := c.Query("ip_addresses"); ipAddresses != "" {
		query.IPAddresses = strings.Split(ipAddresses, ",")
	}

	// 请求信息
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

	// 搜索
	if searchTerm := c.Query("search"); searchTerm != "" {
		query.SearchTerm = searchTerm
	}

	// 排序
	if sortBy := c.Query("sort_by"); sortBy != "" {
		query.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		query.SortOrder = sortOrder
	}

	// 分页
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

// parseStatisticsFilter 解析统计过滤器
func (h *AuditHandlers) parseStatisticsFilter(c *gin.Context) (*audit.StatisticsFilter, error) {
	filter := &audit.StatisticsFilter{}

	// 时间范围
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

	// 事件类型
	if eventTypes := c.Query("event_types"); eventTypes != "" {
		types := strings.Split(eventTypes, ",")
		for _, t := range types {
			filter.EventTypes = append(filter.EventTypes, audit.EventType(strings.TrimSpace(t)))
		}
	}

	// 用户过滤
	if userIDs := c.Query("user_ids"); userIDs != "" {
		filter.UserIDs = strings.Split(userIDs, ",")
	}

	// 租户过滤
	if tenantIDs := c.Query("tenant_ids"); tenantIDs != "" {
		filter.TenantIDs = strings.Split(tenantIDs, ",")
	}

	// 资源过滤
	if resourceTypes := c.Query("resource_types"); resourceTypes != "" {
		filter.ResourceTypes = strings.Split(resourceTypes, ",")
	}

	// 聚合级别
	if groupBy := c.Query("group_by"); groupBy != "" {
		filter.GroupBy = groupBy
	}

	// 时间间隔
	if interval := c.Query("interval"); interval != "" {
		filter.Interval = interval
	}

	// 限制结果数量
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			return nil, fmt.Errorf("invalid limit")
		}
		filter.Limit = limit
	}

	return filter, nil
}

// validateExportRequest 验证导出请求
func (h *AuditHandlers) validateExportRequest(request *audit.ExportRequest) error {
	// 验证格式
	validFormats := map[string]bool{
		"json": true,
		"csv":  true,
		"xlsx": true,
		"pdf":  true,
	}

	if !validFormats[request.Format] {
		return fmt.Errorf("unsupported format: %s", request.Format)
	}

	// 验证时间范围
	if request.Query != nil && request.Query.StartTime != nil && request.Query.EndTime != nil {
		if request.Query.EndTime.Before(*request.Query.StartTime) {
			return fmt.Errorf("end_time must be after start_time")
		}

		// 限制导出时间范围（例如最多90天）
		maxDuration := 90 * 24 * time.Hour
		if request.Query.EndTime.Sub(*request.Query.StartTime) > maxDuration {
			return fmt.Errorf("export time range cannot exceed 90 days")
		}
	}

	return nil
}

// 辅助函数

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return ""
}

// GetCurrentTenantID 获取当前租户ID
func GetCurrentTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}
	return ""
}

// RequirePermission 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里应该实现权限检查逻辑
		// 暂时简化处理
		c.Next()
	}
}

// RequireRole 角色检查中间件
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

		// 检查角色
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