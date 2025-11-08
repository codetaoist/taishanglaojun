package ci

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/services/api/internal/middleware"
)

// CIHandler handles CI/CD-related HTTP requests
type CIHandler struct {
	pipeline *CIPipeline
	logger   *middleware.Logger
}

// NewCIHandler creates a new CI/CD handler
func NewCIHandler(pipeline *CIPipeline, logger *middleware.Logger) *CIHandler {
	return &CIHandler{
		pipeline: pipeline,
		logger:   logger,
	}
}

// ListPipelines handles the GET /pipelines endpoint
func (h *CIHandler) ListPipelines(c *gin.Context) {
	pipelines := h.pipeline.ListPipelines()
	c.JSON(http.StatusOK, gin.H{
		"pipelines": pipelines,
		"count":     len(pipelines),
	})
}

// GetPipeline handles the GET /pipelines/:id endpoint
func (h *CIHandler) GetPipeline(c *gin.Context) {
	id := c.Param("id")
	pipeline, exists := h.pipeline.GetPipeline(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pipeline with ID %s not found", id),
		})
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

// CreatePipeline handles the POST /pipelines endpoint
func (h *CIHandler) CreatePipeline(c *gin.Context) {
	var req struct {
		Name       string   `json:"name" binding:"required"`
		Repository string   `json:"repository" binding:"required"`
		Branch     string   `json:"branch" binding:"required"`
		Commit     string   `json:"commit" binding:"required"`
		Config     CIConfig `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	pipeline, err := h.pipeline.CreatePipeline(
		c.Request.Context(),
		req.Name,
		req.Repository,
		req.Branch,
		req.Commit,
		req.Config,
	)
	if err != nil {
		h.logger.Errorf("Failed to create pipeline: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to create pipeline: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, pipeline)
}

// DeletePipeline handles the DELETE /pipelines/:id endpoint
func (h *CIHandler) DeletePipeline(c *gin.Context) {
	id := c.Param("id")
	if err := h.pipeline.DeletePipeline(c.Request.Context(), id); err != nil {
		h.logger.Errorf("Failed to delete pipeline %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to delete pipeline: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pipeline deleted successfully",
	})
}

// RunBuild handles the POST /pipelines/:id/builds endpoint
func (h *CIHandler) RunBuild(c *gin.Context) {
	id := c.Param("id")
	
	// Get triggered by from query parameter or use default
	triggeredBy := c.DefaultQuery("triggered_by", "api")
	
	build, err := h.pipeline.RunBuild(c.Request.Context(), id, triggeredBy)
	if err != nil {
		h.logger.Errorf("Failed to run build for pipeline %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to run build: %v", err),
		})
		return
	}

	c.JSON(http.StatusAccepted, build)
}

// GetBuilds handles the GET /pipelines/:id/builds endpoint
func (h *CIHandler) GetBuilds(c *gin.Context) {
	id := c.Param("id")
	pipeline, exists := h.pipeline.GetPipeline(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pipeline with ID %s not found", id),
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Apply pagination
	builds := pipeline.Builds
	total := len(builds)
	
	start := (page - 1) * limit
	end := start + limit
	
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	
	paginatedBuilds := builds[start:end]

	c.JSON(http.StatusOK, gin.H{
		"builds": paginatedBuilds,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetBuild handles the GET /pipelines/:id/builds/:build_id endpoint
func (h *CIHandler) GetBuild(c *gin.Context) {
	id := c.Param("id")
	buildID := c.Param("build_id")
	
	pipeline, exists := h.pipeline.GetPipeline(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pipeline with ID %s not found", id),
		})
		return
	}

	// Find the build
	for _, build := range pipeline.Builds {
		if build.ID == buildID {
			c.JSON(http.StatusOK, build)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": fmt.Sprintf("Build with ID %s not found", buildID),
	})
}

// GetBuildLogs handles the GET /pipelines/:id/builds/:build_id/logs endpoint
func (h *CIHandler) GetBuildLogs(c *gin.Context) {
	id := c.Param("id")
	buildID := c.Param("build_id")
	
	pipeline, exists := h.pipeline.GetPipeline(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pipeline with ID %s not found", id),
		})
		return
	}

	// Find the build
	for _, build := range pipeline.Builds {
		if build.ID == buildID {
			c.JSON(http.StatusOK, gin.H{
				"build_id": buildID,
				"logs":     build.Log,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": fmt.Sprintf("Build with ID %s not found", buildID),
	})
}

// CancelBuild handles the DELETE /pipelines/:id/builds/:build_id endpoint
func (h *CIHandler) CancelBuild(c *gin.Context) {
	id := c.Param("id")
	buildID := c.Param("build_id")
	
	pipeline, exists := h.pipeline.GetPipeline(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Pipeline with ID %s not found", id),
		})
		return
	}

	// Find the build
	for _, build := range pipeline.Builds {
		if build.ID == buildID {
			// Check if build is running
			if build.Status != BuildStatusRunning {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Build is not running",
				})
				return
			}
			
			// Update status to cancelled
			build.Status = BuildStatusCancelled
			build.EndTime = &time.Time{}
			*build.EndTime = time.Now()
			build.Duration = build.EndTime.Sub(build.StartTime)
			
			c.JSON(http.StatusOK, gin.H{
				"message": "Build cancelled successfully",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": fmt.Sprintf("Build with ID %s not found", buildID),
	})
}

// CIRouter sets up the CI/CD API routes
type CIRouter struct {
	handler *CIHandler
}

// NewCIRouter creates a new CI/CD router
func NewCIRouter(handler *CIHandler) *CIRouter {
	return &CIRouter{
		handler: handler,
	}
}

// SetupRoutes sets up the CI/CD API routes
func (r *CIRouter) SetupRoutes(router gin.IRouter) {
	pipelines := router.Group("/pipelines")
	{
		pipelines.GET("", r.handler.ListPipelines)
		pipelines.POST("", r.handler.CreatePipeline)
		pipelines.GET("/:id", r.handler.GetPipeline)
		pipelines.DELETE("/:id", r.handler.DeletePipeline)
		
		// Build endpoints
		pipelines.GET("/:id/builds", r.handler.GetBuilds)
		pipelines.POST("/:id/builds", r.handler.RunBuild)
		pipelines.GET("/:id/builds/:build_id", r.handler.GetBuild)
		pipelines.GET("/:id/builds/:build_id/logs", r.handler.GetBuildLogs)
		pipelines.DELETE("/:id/builds/:build_id", r.handler.CancelBuild)
	}
}

// CIAPIHandler combines all CI/CD-related handlers
type CIAPIHandler struct {
	router *CIRouter
}

// NewCIAPIHandler creates a new CI/CD API handler
func NewCIAPIHandler(pipeline *CIPipeline, logger *middleware.Logger) *CIAPIHandler {
	handler := NewCIHandler(pipeline, logger)
	router := NewCIRouter(handler)
	
	return &CIAPIHandler{
		router: router,
	}
}

// SetupRoutes sets up all CI/CD API routes
func (h *CIAPIHandler) SetupRoutes(router gin.IRouter) {
	h.router.SetupRoutes(router)
}