package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/analytics"
	"github.com/gin-gonic/gin"
)

// AnalyticsHandlers HTTP
type AnalyticsHandlers struct {
	service analytics.AnalyticsService
}

// NewAnalyticsHandlers 
func NewAnalyticsHandlers(service analytics.AnalyticsService) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		service: service,
	}
}

// RegisterRoutes 
func (h *AnalyticsHandlers) RegisterRoutes(router *gin.RouterGroup) {
	// 
	router.POST("/data", h.CollectData)
	router.POST("/data/batch", h.CollectBatchData)

	// 
	router.GET("/data", h.QueryData)
	router.GET("/data/aggregated", h.QueryAggregatedData)

	// 
	router.POST("/analysis", h.AnalyzeData)
	router.GET("/analysis/:id", h.GetAnalysisResult)
	router.POST("/analysis/batch", h.BatchAnalyzeData)

	// 
	router.POST("/realtime/analysis", h.StartRealtimeAnalysis)
	router.DELETE("/realtime/analysis/:id", h.StopRealtimeAnalysis)
	router.GET("/realtime/analysis/:id/status", h.GetRealtimeAnalysisStatus)

	// 
	router.POST("/reports", h.GenerateReport)
	router.GET("/reports", h.ListReports)
	router.GET("/reports/:id", h.GetReport)
	router.PUT("/reports/:id", h.UpdateReport)
	router.DELETE("/reports/:id", h.DeleteReport)
	router.GET("/reports/:id/download", h.DownloadReport)

	// 
	router.POST("/export", h.ExportData)
	router.GET("/export/:id/status", h.GetExportStatus)
	router.GET("/export/:id/download", h.DownloadExport)

	// 
	router.DELETE("/data/cleanup", h.CleanupData)

	// 
	router.GET("/health", h.HealthCheck)
	router.GET("/stats", h.GetSystemStats)
}

// CollectData 
func (h *AnalyticsHandlers) CollectData(c *gin.Context) {
	var req analytics.DataCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:data:collect") {
		return
	}

	// IDID
	req.DataPoint.TenantID = h.getTenantID(c)
	req.DataPoint.UserID = h.getUserID(c)

	resp, err := h.service.CollectData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// CollectBatchData 
func (h *AnalyticsHandlers) CollectBatchData(c *gin.Context) {
	var req analytics.BatchDataCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:data:collect") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)
	for _, dataPoint := range req.DataPoints {
		dataPoint.TenantID = req.TenantID
		dataPoint.UserID = req.UserID
	}

	resp, err := h.service.BatchCollectData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect batch data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// QueryData 
func (h *AnalyticsHandlers) QueryData(c *gin.Context) {
	// 
	if !h.checkPermission(c, "analytics:data:read") {
		return
	}

	filter := h.parseDataFilter(c)
	filter.TenantID = h.getTenantID(c)

	req := &analytics.DataQueryRequest{Filter: filter}
	resp, err := h.service.QueryData(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// QueryAggregatedData 
func (h *AnalyticsHandlers) QueryAggregatedData(c *gin.Context) {
	// 
	if !h.checkPermission(c, "analytics:data:read") {
		return
	}

	filter := h.parseAggregationFilter(c)
	filter.TenantID = h.getTenantID(c)

	req := &analytics.AggregatedDataQueryRequest{Filter: filter}
	resp, err := h.service.QueryAggregatedData(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query aggregated data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AnalyzeData 
func (h *AnalyticsHandlers) AnalyzeData(c *gin.Context) {
	var req analytics.DataAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:analysis:execute") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)

	resp, err := h.service.AnalyzeData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// GetAnalysisResult 
func (h *AnalyticsHandlers) GetAnalysisResult(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis ID is required"})
		return
	}

	// 简化实现，直接返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"analysis_id": analysisID,
		"status":      "completed",
		"result":      map[string]interface{}{"message": "Analysis completed"},
	})
}

// BatchAnalyzeData 
func (h *AnalyticsHandlers) BatchAnalyzeData(c *gin.Context) {
	var req analytics.BatchDataAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:analysis:execute") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)
	for _, request := range req.Requests {
		request.TenantID = req.TenantID
		request.UserID = req.UserID
	}

	resp, err := h.service.BatchAnalyzeData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch analyze data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// StartRealtimeAnalysis 
func (h *AnalyticsHandlers) StartRealtimeAnalysis(c *gin.Context) {
	var req analytics.RealTimeAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:realtime:manage") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)

	resp, err := h.service.StartRealTimeAnalysis(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start realtime analysis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// StopRealtimeAnalysis 
func (h *AnalyticsHandlers) StopRealtimeAnalysis(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:realtime:manage") {
		return
	}

	req := &analytics.StopRealTimeAnalysisRequest{
		ID:       analysisID,
		TenantID: h.getTenantID(c),
		UserID:   h.getUserID(c),
	}

	resp, err := h.service.StopRealTimeAnalysis(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop realtime analysis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetRealtimeAnalysisStatus 
func (h *AnalyticsHandlers) GetRealtimeAnalysisStatus(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:realtime:read") {
		return
	}

	req := &analytics.RealTimeAnalysisStatusRequest{
		ID:       analysisID,
		TenantID: h.getTenantID(c),
		UserID:   h.getUserID(c),
	}

	resp, err := h.service.GetRealTimeAnalysisStatus(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get realtime analysis status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateReport 
func (h *AnalyticsHandlers) GenerateReport(c *gin.Context) {
	var req analytics.ReportGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:reports:create") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)

	resp, err := h.service.GenerateReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// ListReports 
func (h *AnalyticsHandlers) ListReports(c *gin.Context) {
	// 
	if !h.checkPermission(c, "analytics:reports:read") {
		return
	}

	filter := h.parseReportFilter(c)
	filter.TenantID = h.getTenantID(c)

	req := &analytics.ListReportsRequest{Filter: filter}
	resp, err := h.service.ListReports(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list reports", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetReport 
func (h *AnalyticsHandlers) GetReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:reports:read") {
		return
	}

	req := &analytics.GetReportRequest{
		ReportID: reportID,
		TenantID: h.getTenantID(c),
	}

	resp, err := h.service.GetReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get report", "details": err.Error()})
		return
	}

	if resp.Report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateReport 
func (h *AnalyticsHandlers) UpdateReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	var req analytics.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:reports:update") {
		return
	}

	req.ReportID = reportID
	req.TenantID = h.getTenantID(c)

	resp, err := h.service.UpdateReport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteReport 
func (h *AnalyticsHandlers) DeleteReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:reports:delete") {
		return
	}

	req := &analytics.DeleteReportRequest{
		ReportID: reportID,
		TenantID: h.getTenantID(c),
	}

	resp, err := h.service.DeleteReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete report", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DownloadReport 
func (h *AnalyticsHandlers) DownloadReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:reports:download") {
		return
	}

	req := &analytics.DownloadReportRequest{
		ReportID: reportID,
		TenantID: h.getTenantID(c),
	}

	resp, err := h.service.DownloadReport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download report", "details": err.Error()})
		return
	}

	// 
	c.Header("Content-Type", resp.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.FileName))
	c.Header("Content-Length", strconv.FormatInt(resp.FileSize, 10))

	// 
	c.Data(http.StatusOK, resp.ContentType, resp.Data)
}

// ExportData 
func (h *AnalyticsHandlers) ExportData(c *gin.Context) {
	var req analytics.ExportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:data:export") {
		return
	}

	// IDID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)

	resp, err := h.service.ExportData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// GetExportStatus 
func (h *AnalyticsHandlers) GetExportStatus(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Export ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:data:export") {
		return
	}

	req := &analytics.GetExportStatusRequest{
		ExportID: exportID,
		TenantID: h.getTenantID(c),
	}

	resp, err := h.service.GetExportStatus(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get export status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DownloadExport 
func (h *AnalyticsHandlers) DownloadExport(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Export ID is required"})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:data:export") {
		return
	}

	req := &analytics.DownloadExportRequest{
		ExportID: exportID,
		TenantID: h.getTenantID(c),
	}

	resp, err := h.service.DownloadExport(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download export", "details": err.Error()})
		return
	}

	// 
	c.Header("Content-Type", resp.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.FileName))
	c.Header("Content-Length", strconv.FormatInt(resp.FileSize, 10))

	// 
	c.Data(http.StatusOK, resp.ContentType, resp.Data)
}

// CleanupData 
func (h *AnalyticsHandlers) CleanupData(c *gin.Context) {
	var req analytics.CleanupDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 
	if !h.checkPermission(c, "analytics:data:cleanup") {
		return
	}

	// ID
	req.TenantID = h.getTenantID(c)

	resp, err := h.service.CleanupData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HealthCheck 
func (h *AnalyticsHandlers) HealthCheck(c *gin.Context) {
	resp, err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Health check failed", "details": err.Error()})
		return
	}

	if resp.Status != "healthy" {
		c.JSON(http.StatusServiceUnavailable, resp)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSystemStats 
func (h *AnalyticsHandlers) GetSystemStats(c *gin.Context) {
	// 
	if !h.checkPermission(c, "analytics:system:read") {
		return
	}

	req := &analytics.GetSystemStatsRequest{
		TenantID: h.getTenantID(c),
	}

	resp, err := h.service.GetSystemStats(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system stats", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 

func (h *AnalyticsHandlers) parseDataFilter(c *gin.Context) *analytics.DataFilter {
	filter := &analytics.DataFilter{}

	// 
	if sources := c.QueryArray("sources"); len(sources) > 0 {
		filter.Sources = sources
	}

	// 
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, analytics.DataType(t))
		}
	}

	// 
	if categories := c.QueryArray("categories"); len(categories) > 0 {
		filter.Categories = categories
	}

	// 
	if startTime := c.Query("start_time"); startTime != "" {
		if start, err := time.Parse(time.RFC3339, startTime); err == nil {
			if filter.TimeRange == nil {
				filter.TimeRange = &analytics.TimeRange{}
			}
			filter.TimeRange.Start = start
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if end, err := time.Parse(time.RFC3339, endTime); err == nil {
			if filter.TimeRange == nil {
				filter.TimeRange = &analytics.TimeRange{}
			}
			filter.TimeRange.End = end
		}
	}

	// ID
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	// 
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// 
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	return filter
}

func (h *AnalyticsHandlers) parseAggregationFilter(c *gin.Context) *analytics.AggregationFilter {
	filter := &analytics.AggregationFilter{}

	// 
	if sources := c.QueryArray("sources"); len(sources) > 0 {
		filter.Sources = sources
	}

	// 
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, analytics.DataType(t))
		}
	}

	// 
	if aggregations := c.QueryArray("aggregations"); len(aggregations) > 0 {
		for _, a := range aggregations {
			filter.Aggregations = append(filter.Aggregations, analytics.AggregationType(a))
		}
	}

	// 
	if startTime := c.Query("start_time"); startTime != "" {
		if start, err := time.Parse(time.RFC3339, startTime); err == nil {
			if filter.TimeRange == nil {
				filter.TimeRange = &analytics.TimeRange{}
			}
			filter.TimeRange.Start = start
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if end, err := time.Parse(time.RFC3339, endTime); err == nil {
			if filter.TimeRange == nil {
				filter.TimeRange = &analytics.TimeRange{}
			}
			filter.TimeRange.End = end
		}
	}

	// 
	if groupBy := c.QueryArray("group_by"); len(groupBy) > 0 {
		filter.GroupBy = groupBy
	}

	return filter
}

func (h *AnalyticsHandlers) parseReportFilter(c *gin.Context) *analytics.ReportFilter {
	filter := &analytics.ReportFilter{}

	// 
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, analytics.ReportType(t))
		}
	}

	// 
	if statuses := c.QueryArray("statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, analytics.ReportStatus(s))
		}
	}

	// ID
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	// 
	if startTime := c.Query("start_time"); startTime != "" {
		if start, err := time.Parse(time.RFC3339, startTime); err == nil {
			if filter.TimeRange == nil {
				filter.TimeRange = &analytics.TimeRange{}
			}
			filter.TimeRange.Start = start
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if end, err := time.Parse(time.RFC3339, endTime); err == nil {
			if filter.TimeRange == nil {
				filter.TimeRange = &analytics.TimeRange{}
			}
			filter.TimeRange.End = end
		}
	}

	// 
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	return filter
}

func (h *AnalyticsHandlers) checkPermission(c *gin.Context, permission string) bool {
	// 
	// true

	// 
	// userID := h.getUserID(c)
	// tenantID := h.getTenantID(c)
	//
	// hasPermission, err := h.permissionService.CheckPermission(c.Request.Context(), &permission.CheckPermissionRequest{
	//     UserID:     userID,
	//     TenantID:   tenantID,
	//     Permission: permission,
	// })
	//
	// if err != nil || !hasPermission.Allowed {
	//     c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
	//     return false
	// }

	return true
}

func (h *AnalyticsHandlers) getUserID(c *gin.Context) string {
	// JWT tokenheadercontextID
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}

	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}

	return ""
}

func (h *AnalyticsHandlers) getTenantID(c *gin.Context) string {
	// JWT tokenheadercontextID
	if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	if tenantID, exists := c.Get("tenant_id"); exists {
		if tid, ok := tenantID.(string); ok {
			return tid
		}
	}

	return ""
}

