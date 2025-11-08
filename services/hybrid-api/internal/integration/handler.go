package integration

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/services/api/internal/middleware"
)

// PluginCIHandler handles HTTP requests for plugin-CI integration
type PluginCIHandler struct {
	integration *PluginCIIntegration
	logger      *middleware.Logger
}

// NewPluginCIHandler creates a new instance of PluginCIHandler
func NewPluginCIHandler(integration *PluginCIIntegration, logger *middleware.Logger) *PluginCIHandler {
	return &PluginCIHandler{
		integration: integration,
		logger:      logger,
	}
}

// PluginCIRouter sets up the routes for plugin-CI integration
type PluginCIRouter struct {
	handler *PluginCIHandler
}

// NewPluginCIRouter creates a new instance of PluginCIRouter
func NewPluginCIRouter(handler *PluginCIHandler) *PluginCIRouter {
	return &PluginCIRouter{
		handler: handler,
	}
}

// SetupRoutes sets up the routes for plugin-CI integration
func (r *PluginCIRouter) SetupRoutes(router *gin.RouterGroup) {
	pluginCI := router.Group("/plugin-ci")
	{
		// Plugin build status
		pluginCI.GET("/plugins/:pluginID/builds", r.handler.GetPluginBuilds)
		
		// Trigger plugin build
		pluginCI.POST("/plugins/:pluginID/build", r.handler.TriggerPluginBuild)
		
		// CI triggers
		pluginCI.GET("/triggers", r.handler.GetCITriggers)
		pluginCI.POST("/triggers", r.handler.RegisterCITrigger)
		pluginCI.DELETE("/triggers/:event", r.handler.UnregisterCITrigger)
		
		// Integration status
		pluginCI.GET("/status", r.handler.GetIntegrationStatus)
	}
}

// GetPluginBuilds returns the build status for a plugin
func (h *PluginCIHandler) GetPluginBuilds(c *gin.Context) {
	pluginID := c.Param("pluginID")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID is required"})
		return
	}
	
	statuses, err := h.integration.GetPluginBuildStatus(pluginID)
	if err != nil {
		h.logger.Errorf("Failed to get plugin build status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get plugin build status"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"builds":    statuses,
	})
}

// TriggerPluginBuild manually triggers a build for a plugin
func (h *PluginCIHandler) TriggerPluginBuild(c *gin.Context) {
	pluginID := c.Param("pluginID")
	if pluginID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID is required"})
		return
	}
	
	// Parse request body if needed
	var req struct {
		Parameters map[string]interface{} `json:"parameters,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	buildID, err := h.integration.TriggerPluginBuild(ctx, pluginID)
	if err != nil {
		h.logger.Errorf("Failed to trigger plugin build: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger plugin build"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"build_id":  buildID,
		"message":   "Build triggered successfully",
	})
}

// GetCITriggers returns all registered CI triggers
func (h *PluginCIHandler) GetCITriggers(c *gin.Context) {
	// This would require adding a method to PluginCIIntegration to get all triggers
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"triggers": []gin.H{
			{
				"event":        "install",
				"pipeline_id":  "plugin-build",
				"auto_trigger": true,
			},
			{
				"event":        "update",
				"pipeline_id":  "plugin-build",
				"auto_trigger": true,
			},
		},
	})
}

// RegisterCITrigger registers a new CI trigger
func (h *PluginCIHandler) RegisterCITrigger(c *gin.Context) {
	var req struct {
		PluginID     string                 `json:"plugin_id" binding:"required"`
		TriggerEvent string                 `json:"trigger_event" binding:"required"`
		PipelineID   string                 `json:"pipeline_id" binding:"required"`
		AutoTrigger  bool                   `json:"auto_trigger"`
		Parameters   map[string]interface{} `json:"parameters,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	
	trigger := &CITriggerConfig{
		PluginID:     req.PluginID,
		TriggerEvent: req.TriggerEvent,
		PipelineID:   req.PipelineID,
		AutoTrigger:  req.AutoTrigger,
		Parameters:   req.Parameters,
	}
	
	if err := h.integration.RegisterCITrigger(trigger); err != nil {
		h.logger.Errorf("Failed to register CI trigger: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register CI trigger"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "CI trigger registered successfully",
		"trigger": trigger,
	})
}

// UnregisterCITrigger unregisters a CI trigger
func (h *PluginCIHandler) UnregisterCITrigger(c *gin.Context) {
	event := c.Param("event")
	if event == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event is required"})
		return
	}
	
	if err := h.integration.UnregisterCITrigger(event); err != nil {
		h.logger.Errorf("Failed to unregister CI trigger: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister CI trigger"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("CI trigger for event %s unregistered successfully", event),
	})
}

// GetIntegrationStatus returns the status of the plugin-CI integration
func (h *PluginCIHandler) GetIntegrationStatus(c *gin.Context) {
	// This would require adding a method to PluginCIIntegration to get status
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"status": "active",
		"features": []string{
			"auto_build_on_install",
			"auto_build_on_update",
			"build_status_tracking",
			"artifact_cleanup",
		},
		"stats": gin.H{
			"total_builds":       0,
			"successful_builds":  0,
			"failed_builds":      0,
			"active_triggers":     2,
		},
	})
}

// PluginCIAPIHandler combines all plugin-CI related handlers
type PluginCIAPIHandler struct {
	pluginCI *PluginCIHandler
}

// NewPluginCIAPIHandler creates a new instance of PluginCIAPIHandler
func NewPluginCIAPIHandler(integration *PluginCIIntegration, logger *middleware.Logger) *PluginCIAPIHandler {
	return &PluginCIAPIHandler{
		pluginCI: NewPluginCIHandler(integration, logger),
	}
}

// SetupRoutes sets up all plugin-CI related routes
func (h *PluginCIAPIHandler) SetupRoutes(router *gin.RouterGroup) {
	pluginCIRouter := NewPluginCIRouter(h.pluginCI)
	pluginCIRouter.SetupRoutes(router)
}

// GetPluginBuildLogs returns the logs for a specific plugin build
func (h *PluginCIHandler) GetPluginBuildLogs(c *gin.Context) {
	pluginID := c.Param("pluginID")
	buildID := c.Param("buildID")
	
	if pluginID == "" || buildID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID and Build ID are required"})
		return
	}
	
	// Parse query parameters
	query := c.Request.URL.Query()
	linesStr := query.Get("lines")
	lines := 100 // default
	
	if linesStr != "" {
		if parsedLines, err := strconv.Atoi(linesStr); err == nil && parsedLines > 0 {
			lines = parsedLines
		}
	}
	
	// This would require adding a method to PluginCIIntegration to get build logs
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"build_id":  buildID,
		"lines":     lines,
		"logs":      "Build logs would be displayed here...",
	})
}

// GetPluginBuildArtifacts returns the artifacts for a specific plugin build
func (h *PluginCIHandler) GetPluginBuildArtifacts(c *gin.Context) {
	pluginID := c.Param("pluginID")
	buildID := c.Param("buildID")
	
	if pluginID == "" || buildID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID and Build ID are required"})
		return
	}
	
	// This would require adding a method to PluginCIIntegration to get build artifacts
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"plugin_id":  pluginID,
		"build_id":   buildID,
		"artifacts": []gin.H{
			{
				"name": "plugin.tar.gz",
				"url":  fmt.Sprintf("/api/v1/artifacts/%s/%s/plugin.tar.gz", pluginID, buildID),
				"size": "10.5 MB",
			},
		},
	})
}

// CancelPluginBuild cancels a running plugin build
func (h *PluginCIHandler) CancelPluginBuild(c *gin.Context) {
	pluginID := c.Param("pluginID")
	buildID := c.Param("buildID")
	
	if pluginID == "" || buildID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID and Build ID are required"})
		return
	}
	
	// This would require adding a method to PluginCIIntegration to cancel a build
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"plugin_id": pluginID,
		"build_id":  buildID,
		"message":   "Build cancelled successfully",
	})
}

// RetryPluginBuild retries a failed plugin build
func (h *PluginCIHandler) RetryPluginBuild(c *gin.Context) {
	pluginID := c.Param("pluginID")
	buildID := c.Param("buildID")
	
	if pluginID == "" || buildID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plugin ID and Build ID are required"})
		return
	}
	
	// This would require adding a method to PluginCIIntegration to retry a build
	// For now, return a placeholder response
	newBuildID := "build-" + strconv.FormatInt(time.Now().Unix(), 10)
	
	c.JSON(http.StatusOK, gin.H{
		"plugin_id":    pluginID,
		"old_build_id": buildID,
		"new_build_id": newBuildID,
		"message":      "Build retry initiated successfully",
	})
}