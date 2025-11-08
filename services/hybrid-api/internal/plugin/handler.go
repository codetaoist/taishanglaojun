package plugin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/services/api/internal/middleware"
)

// PluginHandler handles plugin-related HTTP requests
type PluginHandler struct {
	pluginManager *PluginManager
	logger        *middleware.Logger
}

// NewPluginHandler creates a new plugin handler
func NewPluginHandler(pluginManager *PluginManager, logger *middleware.Logger) *PluginHandler {
	return &PluginHandler{
		pluginManager: pluginManager,
		logger:        logger,
	}
}

// ListPlugins handles the GET /plugins endpoint
func (h *PluginHandler) ListPlugins(c *gin.Context) {
	plugins := h.pluginManager.ListPlugins()
	c.JSON(http.StatusOK, gin.H{
		"plugins": plugins,
		"count":   len(plugins),
	})
}

// GetPlugin handles the GET /plugins/:id endpoint
func (h *PluginHandler) GetPlugin(c *gin.Context) {
	id := c.Param("id")
	plugin, exists := h.pluginManager.GetPlugin(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Plugin with ID %s not found", id),
		})
		return
	}

	info := plugin.GetInfo()
	c.JSON(http.StatusOK, info)
}

// LoadPlugin handles the POST /plugins/load endpoint
func (h *PluginHandler) LoadPlugin(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.pluginManager.LoadPlugin(c.Request.Context(), req.Path); err != nil {
		h.logger.Errorf("Failed to load plugin from %s: %v", req.Path, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to load plugin: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin loaded successfully",
	})
}

// UnloadPlugin handles the DELETE /plugins/:id endpoint
func (h *PluginHandler) UnloadPlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.pluginManager.UnloadPlugin(c.Request.Context(), id); err != nil {
		h.logger.Errorf("Failed to unload plugin %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to unload plugin: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin unloaded successfully",
	})
}

// EnablePlugin handles the PUT /plugins/:id/enable endpoint
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.pluginManager.EnablePlugin(c.Request.Context(), id); err != nil {
		h.logger.Errorf("Failed to enable plugin %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to enable plugin: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin enabled successfully",
	})
}

// DisablePlugin handles the PUT /plugins/:id/disable endpoint
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	id := c.Param("id")
	if err := h.pluginManager.DisablePlugin(c.Request.Context(), id); err != nil {
		h.logger.Errorf("Failed to disable plugin %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to disable plugin: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin disabled successfully",
	})
}

// HealthCheck handles the GET /plugins/:id/health endpoint
func (h *PluginHandler) HealthCheck(c *gin.Context) {
	id := c.Param("id")
	plugin, exists := h.pluginManager.GetPlugin(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Plugin with ID %s not found", id),
		})
		return
	}

	if err := plugin.HealthCheck(c.Request.Context()); err != nil {
		h.logger.Errorf("Health check failed for plugin %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Health check failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// GetPluginsByType handles the GET /plugins/type/:type endpoint
func (h *PluginHandler) GetPluginsByType(c *gin.Context) {
	typeStr := c.Param("type")
	pluginType := PluginType(typeStr)
	
	plugins := h.pluginManager.GetPluginsByType(pluginType)
	
	var pluginInfos []*PluginInfo
	for _, plugin := range plugins {
		info := plugin.GetInfo()
		pluginInfos = append(pluginInfos, info)
	}

	c.JSON(http.StatusOK, gin.H{
		"plugins": pluginInfos,
		"count":   len(pluginInfos),
	})
}

// PluginRouter sets up the plugin API routes
type PluginRouter struct {
	handler *PluginHandler
}

// NewPluginRouter creates a new plugin router
func NewPluginRouter(handler *PluginHandler) *PluginRouter {
	return &PluginRouter{
		handler: handler,
	}
}

// SetupRoutes sets up the plugin API routes
func (r *PluginRouter) SetupRoutes(router gin.IRouter) {
	plugins := router.Group("/plugins")
	{
		plugins.GET("", r.handler.ListPlugins)
		plugins.GET("/:id", r.handler.GetPlugin)
		plugins.GET("/:id/health", r.handler.HealthCheck)
		plugins.GET("/type/:type", r.handler.GetPluginsByType)
		
		plugins.POST("/load", r.handler.LoadPlugin)
		plugins.DELETE("/:id", r.handler.UnloadPlugin)
		plugins.PUT("/:id/enable", r.handler.EnablePlugin)
		plugins.PUT("/:id/disable", r.handler.DisablePlugin)
	}
}

// PluginAPIHandler combines all plugin-related handlers
type PluginAPIHandler struct {
	router *PluginRouter
}

// NewPluginAPIHandler creates a new plugin API handler
func NewPluginAPIHandler(pluginManager *PluginManager, logger *middleware.Logger) *PluginAPIHandler {
	handler := NewPluginHandler(pluginManager, logger)
	router := NewPluginRouter(handler)
	
	return &PluginAPIHandler{
		router: router,
	}
}

// SetupRoutes sets up all plugin API routes
func (h *PluginAPIHandler) SetupRoutes(router gin.IRouter) {
	h.router.SetupRoutes(router)
}