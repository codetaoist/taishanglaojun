package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/api/internal/config"
	"github.com/codetaoist/taishanglaojun/api/internal/dao"
	"github.com/codetaoist/taishanglaojun/api/internal/middleware"
)

// --- Types aligned with OpenAPI ---

type PluginStatus string

const (
	StatusInstalled PluginStatus = "installed"
	StatusRunning   PluginStatus = "running"
	StatusStopped   PluginStatus = "stopped"
	StatusDisabled  PluginStatus = "disabled"
)

type InstallRequest struct {
	PluginID string `json:"pluginId" binding:"required"`
	Version  string `json:"version" binding:"required"`
	SourceURL string `json:"sourceUrl"`
}

type StartStopRequest struct {
	PluginID string `json:"pluginId" binding:"required"`
}

type UpgradeRequest struct {
	PluginID string `json:"pluginId" binding:"required"`
	Version  string `json:"version" binding:"required"`
}

type UninstallRequest struct {
	PluginID string `json:"pluginId" binding:"required"`
}

type Plugin struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Version  string       `json:"version"`
	Status   PluginStatus `json:"status"`
	Checksum string       `json:"checksum,omitempty"`
}

type PluginListData struct {
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Items    []Plugin  `json:"items"`
}

// --- Response helpers ---

type ErrorCode string

const (
	CodeOK               ErrorCode = "OK"
	CodeInvalidArgument  ErrorCode = "INVALID_ARGUMENT"
	CodeUnauthenticated  ErrorCode = "UNAUTHENTICATED"
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeConflict         ErrorCode = "CONFLICT"
	CodeFailedPrecond    ErrorCode = "FAILED_PRECONDITION"
	CodeInternal         ErrorCode = "INTERNAL"
	CodeUnavailable      ErrorCode = "UNAVAILABLE"
)

func traceIDFromCtx(c *gin.Context) string {
	v, exists := c.Get("traceID")
	if !exists {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func ok(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    CodeOK,
		"message": message,
		"traceId": traceIDFromCtx(c),
		"data":    data,
	})
}

func errorJSON(c *gin.Context, httpStatus int, code ErrorCode, msg string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": msg,
		"traceId": traceIDFromCtx(c),
	})
}

func badRequest(c *gin.Context, msg string) {
	errorJSON(c, http.StatusBadRequest, CodeInvalidArgument, msg)
}

// --- Router setup ---

var pluginStore = store.New()

func Setup(cfg config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestID(cfg.TraceHeader))
	r.Use(middleware.CORS(cfg))

	r.GET("/health", health)

	// Setup Laojun domain routes
	SetupLaojun(cfg, r)

	// Setup Taishang domain routes
	SetupTaishang(cfg, r)

	return r
}

func requestID(header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(header)
		if id == "" {
			id = strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		c.Set("traceID", id)
		c.Next()
	}
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// --- Laojun Domain Handlers ---

func listPlugins(pluginDAO *dao.PluginDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		plugins, err := pluginDAO.List(c.Request.Context(), tenantID)
		if err != nil {
			response.InternalServerError(c, "Failed to list plugins")
			return
		}

		response.Success(c, plugins)
	}
}

func installPlugin(pluginDAO *dao.PluginDAO, auditLogDAO *dao.AuditLogDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req struct {
			Name        string            `json:"name" binding:"required"`
			Version     string            `json:"version" binding:"required"`
			Source      string            `json:"source" binding:"required"`
			Config      map[string]string `json:"config"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request parameters")
			return
		}

		// Create plugin
		plugin := &models.Plugin{
			ID:       "plugin-" + time.Now().Format("20060102150405"),
			TenantID: tenantID,
			Name:     req.Name,
			Version:  req.Version,
			Source:   req.Source,
			Status:   StatusInstalled,
			Config:   req.Config,
		}

		if err := pluginDAO.Create(c.Request.Context(), plugin); err != nil {
			response.InternalServerError(c, "Failed to install plugin")
			return
		}

		// Log audit event
		auditLog := &models.AuditLog{
			TenantID: tenantID,
			Action:   "install",
			Resource: "plugin",
			ResourceID: plugin.ID,
			Details: map[string]interface{}{
				"name":    req.Name,
				"version": req.Version,
			},
		}
		auditLogDAO.Create(c.Request.Context(), auditLog)

		response.Success(c, plugin)
	}
}

func startPlugin(pluginDAO *dao.PluginDAO, auditLogDAO *dao.AuditLogDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req StartStopRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid start request")
			return
		}

		// Check if plugin exists
		plugin, err := pluginDAO.GetByID(c.Request.Context(), tenantID, req.PluginID)
		if err != nil {
			response.NotFound(c, "plugin not found")
			return
		}

		// Update plugin status
		if err := pluginDAO.SetStatus(c.Request.Context(), tenantID, req.PluginID, string(StatusRunning)); err != nil {
			response.InternalServerError(c, "Failed to start plugin")
			return
		}

		// Log audit event
		auditLog := &models.AuditLog{
			TenantID: tenantID,
			Action:   "start",
			Resource: "plugin",
			ResourceID: req.PluginID,
			Details: map[string]interface{}{
				"name": plugin.Name,
			},
		}
		auditLogDAO.Create(c.Request.Context(), auditLog)

		resp := gin.H{
			"running":  true,
			"pluginId": req.PluginID,
		}
		response.Success(c, resp)
	}
}

func stopPlugin(pluginDAO *dao.PluginDAO, auditLogDAO *dao.AuditLogDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req StartStopRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid stop request")
			return
		}

		// Check if plugin exists
		plugin, err := pluginDAO.GetByID(c.Request.Context(), tenantID, req.PluginID)
		if err != nil {
			response.NotFound(c, "plugin not found")
			return
		}

		// Update plugin status
		if err := pluginDAO.SetStatus(c.Request.Context(), tenantID, req.PluginID, string(StatusStopped)); err != nil {
			response.InternalServerError(c, "Failed to stop plugin")
			return
		}

		// Log audit event
		auditLog := &models.AuditLog{
			TenantID: tenantID,
			Action:   "stop",
			Resource: "plugin",
			ResourceID: req.PluginID,
			Details: map[string]interface{}{
				"name": plugin.Name,
			},
		}
		auditLogDAO.Create(c.Request.Context(), auditLog)

		resp := gin.H{
			"running":  false,
			"pluginId": req.PluginID,
		}
		response.Success(c, resp)
	}
}

func upgradePlugin(pluginDAO *dao.PluginDAO, auditLogDAO *dao.AuditLogDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req UpgradeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid upgrade request")
			return
		}

		// Check if plugin exists
		plugin, err := pluginDAO.GetByID(c.Request.Context(), tenantID, req.PluginID)
		if err != nil {
			response.NotFound(c, "plugin not found")
			return
		}

		// Update plugin version
		if err := pluginDAO.Upgrade(c.Request.Context(), tenantID, req.PluginID, req.Version); err != nil {
			response.InternalServerError(c, "Failed to upgrade plugin")
			return
		}

		// Log audit event
		auditLog := &models.AuditLog{
			TenantID: tenantID,
			Action:   "upgrade",
			Resource: "plugin",
			ResourceID: req.PluginID,
			Details: map[string]interface{}{
				"name":    plugin.Name,
				"version": req.Version,
			},
		}
		auditLogDAO.Create(c.Request.Context(), auditLog)

		resp := gin.H{
			"upgraded": true,
			"pluginId": req.PluginID,
			"version":  req.Version,
		}
		response.Success(c, resp)
	}
}

func uninstallPlugin(pluginDAO *dao.PluginDAO, auditLogDAO *dao.AuditLogDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req UninstallRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid uninstall request")
			return
		}

		// Check if plugin exists
		plugin, err := pluginDAO.GetByID(c.Request.Context(), tenantID, req.PluginID)
		if err != nil {
			response.NotFound(c, "plugin not found")
			return
		}

		// Delete plugin
		if err := pluginDAO.Delete(c.Request.Context(), tenantID, req.PluginID); err != nil {
			response.InternalServerError(c, "Failed to uninstall plugin")
			return
		}

		// Log audit event
		auditLog := &models.AuditLog{
			TenantID: tenantID,
			Action:   "uninstall",
			Resource: "plugin",
			ResourceID: req.PluginID,
			Details: map[string]interface{}{
				"name": plugin.Name,
			},
		}
		auditLogDAO.Create(c.Request.Context(), auditLog)

		resp := gin.H{
			"uninstalled": true,
			"pluginId":    req.PluginID,
		}
		response.Success(c, resp)
	}
}

// --- utils ---

// SetupLaojun configures routes for the Laojun domain
func SetupLaojun(cfg config.Config, r *gin.Engine) {
	// Apply authentication middleware
	authMiddleware := middleware.Auth(cfg)

	// Plugin management routes
	pluginGroup := r.Group("/api/laojun/plugins")
	pluginGroup.Use(authMiddleware)
	{
		pluginGroup.GET("/list", listPlugins)
		pluginGroup.POST("/install", installPlugin)
		pluginGroup.POST("/start", startPlugin)
		pluginGroup.POST("/stop", stopPlugin)
		pluginGroup.POST("/upgrade", upgradePlugin)
		pluginGroup.DELETE("/uninstall", uninstallPlugin)
	}
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}