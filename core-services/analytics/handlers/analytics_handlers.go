package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/analytics"
)

// AnalyticsHandlers 数据分析HTTP处理�?
type AnalyticsHandlers struct {
	service analytics.AnalyticsService
}

// NewAnalyticsHandlers 创建数据分析处理�?
func NewAnalyticsHandlers(service analytics.AnalyticsService) *AnalyticsHandlers {
	return &AnalyticsHandlers{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *AnalyticsHandlers) RegisterRoutes(router *gin.RouterGroup) {
	// 数据收集
	router.POST("/data", h.CollectData)
	router.POST("/data/batch", h.CollectBatchData)
	
	// 数据查询
	router.GET("/data", h.QueryData)
	router.GET("/data/aggregated", h.QueryAggregatedData)
	
	// 数据分析
	router.POST("/analysis", h.AnalyzeData)
	router.GET("/analysis/:id", h.GetAnalysisResult)
	router.POST("/analysis/batch", h.BatchAnalyzeData)
	
	// 实时分析
	router.POST("/realtime/analysis", h.StartRealtimeAnalysis)
	router.DELETE("/realtime/analysis/:id", h.StopRealtimeAnalysis)
	router.GET("/realtime/analysis/:id/status", h.GetRealtimeAnalysisStatus)
	
	// 报表管理
	router.POST("/reports", h.GenerateReport)
	router.GET("/reports", h.ListReports)
	router.GET("/reports/:id", h.GetReport)
	router.PUT("/reports/:id", h.UpdateReport)
	router.DELETE("/reports/:id", h.DeleteReport)
	router.GET("/reports/:id/download", h.DownloadReport)
	
	// 数据导出
	router.POST("/export", h.ExportData)
	router.GET("/export/:id/status", h.GetExportStatus)
	router.GET("/export/:id/download", h.DownloadExport)
	
	// 数据清理
	router.DELETE("/data/cleanup", h.CleanupData)
	
	// 系统管理
	router.GET("/health", h.HealthCheck)
	router.GET("/stats", h.GetSystemStats)
}

// CollectData 收集数据
func (h *AnalyticsHandlers) CollectData(c *gin.Context) {
	var req analytics.CollectDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:data:collect") {
		return
	}

	// 设置租户ID和用户ID
	req.DataPoint.TenantID = h.getTenantID(c)
	req.DataPoint.UserID = h.getUserID(c)

	resp, err := h.service.CollectData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// CollectBatchData 批量收集数据
func (h *AnalyticsHandlers) CollectBatchData(c *gin.Context) {
	var req analytics.CollectBatchDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:data:collect") {
		return
	}

	// 设置租户ID和用户ID
	tenantID := h.getTenantID(c)
	userID := h.getUserID(c)
	for _, dataPoint := range req.DataPoints {
		dataPoint.TenantID = tenantID
		dataPoint.UserID = userID
	}

	resp, err := h.service.CollectBatchData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to collect batch data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// QueryData 查询数据
func (h *AnalyticsHandlers) QueryData(c *gin.Context) {
	// 验证权限
	if !h.checkPermission(c, "analytics:data:read") {
		return
	}

	filter := h.parseDataFilter(c)
	filter.TenantID = h.getTenantID(c)

	req := &analytics.QueryDataRequest{Filter: filter}
	resp, err := h.service.QueryData(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// QueryAggregatedData 查询聚合数据
func (h *AnalyticsHandlers) QueryAggregatedData(c *gin.Context) {
	// 验证权限
	if !h.checkPermission(c, "analytics:data:read") {
		return
	}

	filter := h.parseAggregationFilter(c)
	filter.TenantID = h.getTenantID(c)

	req := &analytics.QueryAggregatedDataRequest{Filter: filter}
	resp, err := h.service.QueryAggregatedData(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query aggregated data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AnalyzeData 分析数据
func (h *AnalyticsHandlers) AnalyzeData(c *gin.Context) {
	var req analytics.AnalyzeDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
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

// GetAnalysisResult 获取分析结果
func (h *AnalyticsHandlers) GetAnalysisResult(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis ID is required"})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:analysis:read") {
		return
	}

	req := &analytics.GetAnalysisResultRequest{
		AnalysisID: analysisID,
		TenantID:   h.getTenantID(c),
	}

	resp, err := h.service.GetAnalysisResult(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analysis result", "details": err.Error()})
		return
	}

	if resp.Result == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analysis result not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// BatchAnalyzeData 批量分析数据
func (h *AnalyticsHandlers) BatchAnalyzeData(c *gin.Context) {
	var req analytics.BatchAnalyzeDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:analysis:execute") {
		return
	}

	// 设置租户ID和用户ID
	tenantID := h.getTenantID(c)
	userID := h.getUserID(c)
	for _, analysis := range req.Analyses {
		analysis.TenantID = tenantID
		analysis.UserID = userID
	}

	resp, err := h.service.BatchAnalyzeData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch analyze data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// StartRealtimeAnalysis 开始实时分�?
func (h *AnalyticsHandlers) StartRealtimeAnalysis(c *gin.Context) {
	var req analytics.StartRealtimeAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:realtime:manage") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)

	resp, err := h.service.StartRealtimeAnalysis(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start realtime analysis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// StopRealtimeAnalysis 停止实时分析
func (h *AnalyticsHandlers) StopRealtimeAnalysis(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis ID is required"})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:realtime:manage") {
		return
	}

	req := &analytics.StopRealtimeAnalysisRequest{
		AnalysisID: analysisID,
		TenantID:   h.getTenantID(c),
	}

	resp, err := h.service.StopRealtimeAnalysis(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop realtime analysis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetRealtimeAnalysisStatus 获取实时分析状�?
func (h *AnalyticsHandlers) GetRealtimeAnalysisStatus(c *gin.Context) {
	analysisID := c.Param("id")
	if analysisID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Analysis ID is required"})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:realtime:read") {
		return
	}

	req := &analytics.GetRealtimeAnalysisStatusRequest{
		AnalysisID: analysisID,
		TenantID:   h.getTenantID(c),
	}

	resp, err := h.service.GetRealtimeAnalysisStatus(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get realtime analysis status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateReport 生成报表
func (h *AnalyticsHandlers) GenerateReport(c *gin.Context) {
	var req analytics.GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
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

// ListReports 列出报表
func (h *AnalyticsHandlers) ListReports(c *gin.Context) {
	// 验证权限
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

// GetReport 获取报表
func (h *AnalyticsHandlers) GetReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// 验证权限
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

// UpdateReport 更新报表
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

	// 验证权限
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

// DeleteReport 删除报表
func (h *AnalyticsHandlers) DeleteReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// 验证权限
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

// DownloadReport 下载报表
func (h *AnalyticsHandlers) DownloadReport(c *gin.Context) {
	reportID := c.Param("id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Report ID is required"})
		return
	}

	// 验证权限
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

	// 设置响应�?
	c.Header("Content-Type", resp.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.FileName))
	c.Header("Content-Length", strconv.FormatInt(resp.FileSize, 10))

	// 返回文件内容
	c.Data(http.StatusOK, resp.ContentType, resp.Data)
}

// ExportData 导出数据
func (h *AnalyticsHandlers) ExportData(c *gin.Context) {
	var req analytics.ExportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:data:export") {
		return
	}

	// 设置租户ID和用户ID
	req.TenantID = h.getTenantID(c)
	req.UserID = h.getUserID(c)

	resp, err := h.service.ExportData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// GetExportStatus 获取导出状�?
func (h *AnalyticsHandlers) GetExportStatus(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Export ID is required"})
		return
	}

	// 验证权限
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

// DownloadExport 下载导出文件
func (h *AnalyticsHandlers) DownloadExport(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Export ID is required"})
		return
	}

	// 验证权限
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

	// 设置响应�?
	c.Header("Content-Type", resp.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.FileName))
	c.Header("Content-Length", strconv.FormatInt(resp.FileSize, 10))

	// 返回文件内容
	c.Data(http.StatusOK, resp.ContentType, resp.Data)
}

// CleanupData 清理数据
func (h *AnalyticsHandlers) CleanupData(c *gin.Context) {
	var req analytics.CleanupDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 验证权限
	if !h.checkPermission(c, "analytics:data:cleanup") {
		return
	}

	// 设置租户ID
	req.TenantID = h.getTenantID(c)

	resp, err := h.service.CleanupData(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup data", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HealthCheck 健康检�?
func (h *AnalyticsHandlers) HealthCheck(c *gin.Context) {
	resp, err := h.service.HealthCheck(c.Request.Context(), &analytics.HealthCheckRequest{})
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

// GetSystemStats 获取系统统计
func (h *AnalyticsHandlers) GetSystemStats(c *gin.Context) {
	// 验证权限
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

// 辅助方法

func (h *AnalyticsHandlers) parseDataFilter(c *gin.Context) *analytics.DataFilter {
	filter := &analytics.DataFilter{}

	// 解析数据�?
	if sources := c.QueryArray("sources"); len(sources) > 0 {
		filter.Sources = sources
	}

	// 解析数据类型
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, analytics.DataType(t))
		}
	}

	// 解析分类
	if categories := c.QueryArray("categories"); len(categories) > 0 {
		filter.Categories = categories
	}

	// 解析时间范围
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

	// 解析用户ID
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	// 解析标签
	if tags := c.QueryArray("tags"); len(tags) > 0 {
		filter.Tags = tags
	}

	// 解析分页
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

	// 解析数据�?
	if sources := c.QueryArray("sources"); len(sources) > 0 {
		filter.Sources = sources
	}

	// 解析数据类型
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, analytics.DataType(t))
		}
	}

	// 解析聚合类型
	if aggregations := c.QueryArray("aggregations"); len(aggregations) > 0 {
		for _, a := range aggregations {
			filter.Aggregations = append(filter.Aggregations, analytics.AggregationType(a))
		}
	}

	// 解析时间范围
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

	// 解析分组字段
	if groupBy := c.QueryArray("group_by"); len(groupBy) > 0 {
		filter.GroupBy = groupBy
	}

	return filter
}

func (h *AnalyticsHandlers) parseReportFilter(c *gin.Context) *analytics.ReportFilter {
	filter := &analytics.ReportFilter{}

	// 解析报表类型
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, analytics.ReportType(t))
		}
	}

	// 解析状�?
	if statuses := c.QueryArray("statuses"); len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, analytics.ReportStatus(s))
		}
	}

	// 解析用户ID
	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	// 解析时间范围
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

	// 解析分页
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
	// 这里应该集成权限系统进行权限检�?
	// 暂时返回true，实际应用中需要实现具体的权限验证逻辑
	
	// 示例权限检查逻辑�?
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
	// 从JWT token、header或context中获取用户ID
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
	// 从JWT token、header或context中获取租户ID
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
