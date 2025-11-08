package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/services/api/internal/service"
	"github.com/codetaoist/services/api/internal/middleware"
	pbVector "github.com/codetaoist/api/proto/vector"
	pbModel "github.com/codetaoist/api/proto/model"
)

// HybridServiceHandler handles HTTP requests for the hybrid service
type HybridServiceHandler struct {
	aiManager *service.HybridAIServiceManager
	logger    *middleware.Logger
}

// NewHybridServiceHandler creates a new HybridServiceHandler
func NewHybridServiceHandler(aiManager *service.HybridAIServiceManager, logger *middleware.Logger) *HybridServiceHandler {
	return &HybridServiceHandler{
		aiManager: aiManager,
		logger:    logger,
	}
}

// HealthCheck handles health check requests
func (h *HybridServiceHandler) HealthCheck(c *gin.Context) {
	// Perform health check on AI services
	err := h.aiManager.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Health check failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

// ListServices handles list services requests
func (h *HybridServiceHandler) ListServices(c *gin.Context) {
	// Get the AI client
	aiClient := h.aiManager.GetAIClient()
	if aiClient == nil {
		h.logger.Error("AI client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI client not initialized",
		})
		return
	}

	// Get the vector client
	vectorClient := aiClient.GetVectorClient()
	if vectorClient == nil {
		h.logger.Error("Vector client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Vector client not initialized",
		})
		return
	}

	// List collections
	req := &pbVector.ListCollectionsRequest{}
	resp, err := vectorClient.ListCollections(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to list collections: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collections": resp.Collections,
	})
}

// VectorSearch handles vector search requests
func (h *HybridServiceHandler) VectorSearch(c *gin.Context) {
	// Parse request
	var req pbVector.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the AI client
	aiClient := h.aiManager.GetAIClient()
	if aiClient == nil {
		h.logger.Error("AI client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI client not initialized",
		})
		return
	}

	// Get the vector client
	vectorClient := aiClient.GetVectorClient()
	if vectorClient == nil {
		h.logger.Error("Vector client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Vector client not initialized",
		})
		return
	}

	// Perform search
	resp, err := vectorClient.Search(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("Failed to perform vector search: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": resp.Results,
	})
}

// CreateCollection handles create collection requests
func (h *HybridServiceHandler) CreateCollection(c *gin.Context) {
	// Parse request
	var req pbVector.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the AI client
	aiClient := h.aiManager.GetAIClient()
	if aiClient == nil {
		h.logger.Error("AI client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI client not initialized",
		})
		return
	}

	// Get the vector client
	vectorClient := aiClient.GetVectorClient()
	if vectorClient == nil {
		h.logger.Error("Vector client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Vector client not initialized",
		})
		return
	}

	// Create collection
	resp, err := vectorClient.CreateCollection(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("Failed to create collection: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collection_id": resp.CollectionId,
	})
}

// GenerateText handles text generation requests
func (h *HybridServiceHandler) GenerateText(c *gin.Context) {
	// Parse request
	var req pbModel.GenerateTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the AI client
	aiClient := h.aiManager.GetAIClient()
	if aiClient == nil {
		h.logger.Error("AI client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI client not initialized",
		})
		return
	}

	// Get the model client
	modelClient := aiClient.GetModelClient()
	if modelClient == nil {
		h.logger.Error("Model client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Model client not initialized",
		})
		return
	}

	// Generate text
	resp, err := modelClient.GenerateText(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("Failed to generate text: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"text": resp.Text,
	})
}

// GenerateEmbedding handles embedding generation requests
func (h *HybridServiceHandler) GenerateEmbedding(c *gin.Context) {
	// Parse request
	var req pbModel.GenerateEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the AI client
	aiClient := h.aiManager.GetAIClient()
	if aiClient == nil {
		h.logger.Error("AI client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "AI client not initialized",
		})
		return
	}

	// Get the model client
	modelClient := aiClient.GetModelClient()
	if modelClient == nil {
		h.logger.Error("Model client not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Model client not initialized",
		})
		return
	}

	// Generate embedding
	resp, err := modelClient.GenerateEmbedding(c.Request.Context(), &req)
	if err != nil {
		h.logger.Errorf("Failed to generate embedding: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"embedding": resp.Embedding,
	})
}

// HybridAPIRouter configures API routes for the hybrid service
type HybridAPIRouter struct {
	handler *HybridServiceHandler
}

// NewHybridAPIRouter creates a new HybridAPIRouter
func NewHybridAPIRouter(handler *HybridServiceHandler) *HybridAPIRouter {
	return &HybridAPIRouter{
		handler: handler,
	}
}

// SetupRoutes configures the routes for the hybrid service API
func (r *HybridAPIRouter) SetupRoutes(router *gin.Engine) {
	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", r.handler.HealthCheck)

		// Service management
		v1.GET("/services", r.handler.ListServices)

		// Vector operations
		vectors := v1.Group("/vectors")
		{
			vectors.POST("/search", r.handler.VectorSearch)
			vectors.POST("/collections", r.handler.CreateCollection)
		}

		// Model operations
		models := v1.Group("/models")
		{
			models.POST("/generate-text", r.handler.GenerateText)
			models.POST("/generate-embedding", r.handler.GenerateEmbedding)
		}
	}
}

// ServiceDiscoveryHandler handles HTTP requests for service discovery
type ServiceDiscoveryHandler struct {
	registry service.ServiceRegistry
	logger   *middleware.Logger
}

// NewServiceDiscoveryHandler creates a new ServiceDiscoveryHandler
func NewServiceDiscoveryHandler(registry service.ServiceRegistry, logger *middleware.Logger) *ServiceDiscoveryHandler {
	return &ServiceDiscoveryHandler{
		registry: registry,
		logger:   logger,
	}
}

// RegisterService handles service registration requests
func (h *ServiceDiscoveryHandler) RegisterService(c *gin.Context) {
	// Parse request
	var serviceInfo service.ServiceInfo
	if err := c.ShouldBindJSON(&serviceInfo); err != nil {
		h.logger.Errorf("Failed to bind request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Register service
	err := h.registry.Register(c.Request.Context(), &serviceInfo)
	if err != nil {
		h.logger.Errorf("Failed to register service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "registered",
	})
}

// UnregisterService handles service unregistration requests
func (h *ServiceDiscoveryHandler) UnregisterService(c *gin.Context) {
	// Get service ID from path
	serviceID := c.Param("id")
	if serviceID == "" {
		h.logger.Error("Service ID is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	// Unregister service
	err := h.registry.Unregister(c.Request.Context(), serviceID)
	if err != nil {
		h.logger.Errorf("Failed to unregister service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "unregistered",
	})
}

// DiscoverServices handles service discovery requests
func (h *ServiceDiscoveryHandler) DiscoverServices(c *gin.Context) {
	// Get query parameters
	tags := c.QueryArray("tag")
	protocol := c.Query("protocol")

	// Create filter function
	filter := func(service *service.ServiceInfo) bool {
		// Filter by tags
		if len(tags) > 0 {
			hasTag := false
			for _, tag := range tags {
				for _, serviceTag := range service.Tags {
					if serviceTag == tag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				return false
			}
		}

		// Filter by protocol
		if protocol != "" && service.Protocol != protocol {
			return false
		}

		return true
	}

	// Discover services
	services, err := h.registry.Discover(c.Request.Context(), filter)
	if err != nil {
		h.logger.Errorf("Failed to discover services: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
	})
}

// GetService handles get service requests
func (h *ServiceDiscoveryHandler) GetService(c *gin.Context) {
	// Get service ID from path
	serviceID := c.Param("id")
	if serviceID == "" {
		h.logger.Error("Service ID is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Service ID is required",
		})
		return
	}

	// Get service
	service, err := h.registry.GetService(c.Request.Context(), serviceID)
	if err != nil {
		h.logger.Errorf("Failed to get service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": service,
	})
}

// ServiceDiscoveryRouter configures API routes for service discovery
type ServiceDiscoveryRouter struct {
	handler *ServiceDiscoveryHandler
}

// NewServiceDiscoveryRouter creates a new ServiceDiscoveryRouter
func NewServiceDiscoveryRouter(handler *ServiceDiscoveryHandler) *ServiceDiscoveryRouter {
	return &ServiceDiscoveryRouter{
		handler: handler,
	}
}

// SetupRoutes configures the routes for the service discovery API
func (r *ServiceDiscoveryRouter) SetupRoutes(router *gin.Engine) {
	// API v1 group
	v1 := router.Group("/api/v1/discovery")
	{
		// Service registration
		v1.POST("/services", r.handler.RegisterService)
		v1.DELETE("/services/:id", r.handler.UnregisterService)

		// Service discovery
		v1.GET("/services", r.handler.DiscoverServices)
		v1.GET("/services/:id", r.handler.GetService)
	}
}