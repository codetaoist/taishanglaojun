package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/codetaoist/taishanglaojun/api/proto"
	"github.com/codetaoist/taishanglaojun/api/internal/service"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	aiService *service.AIService
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// Health handles the health check request
func (h *AIHandler) Health(c *gin.Context) {
	ctx := c.Request.Context()
	vectorHealthy, modelHealthy, err := h.aiService.Health(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vector_healthy": vectorHealthy,
		"model_healthy":  modelHealthy,
	})
}

// CreateCollection handles the create collection request
func (h *AIHandler) CreateCollection(c *gin.Context) {
	var req struct {
		Name   string               `json:"name" binding:"required"`
		Schema *pb.CollectionSchema `json:"schema" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.CreateCollection(ctx, req.Name, req.Schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection created successfully",
	})
}

// DropCollection handles the drop collection request
func (h *AIHandler) DropCollection(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.DropCollection(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection dropped successfully",
	})
}

// ListCollections handles the list collections request
func (h *AIHandler) ListCollections(c *gin.Context) {
	ctx := c.Request.Context()
	collections, err := h.aiService.ListCollections(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collections": collections,
	})
}

// Search handles the vector search request
func (h *AIHandler) Search(c *gin.Context) {
	var req struct {
		CollectionName string    `json:"collection_name" binding:"required"`
		Vector         []float64 `json:"vector" binding:"required"`
		TopK           int64     `json:"top_k"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default top_k if not provided
	if req.TopK <= 0 {
		req.TopK = 10
	}

	ctx := c.Request.Context()
	results, err := h.aiService.Search(ctx, req.CollectionName, req.Vector, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// Insert handles the vector insert request
func (h *AIHandler) Insert(c *gin.Context) {
	var req struct {
		CollectionName string               `json:"collection_name" binding:"required"`
		Data           []*pb.VectorData     `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	ids, err := h.aiService.Insert(ctx, req.CollectionName, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ids": ids,
	})
}

// GenerateText handles the text generation request
func (h *AIHandler) GenerateText(c *gin.Context) {
	var req struct {
		ModelName   string  `json:"model_name" binding:"required"`
		Prompt      string  `json:"prompt" binding:"required"`
		MaxLength   int64   `json:"max_length"`
		Temperature float32 `json:"temperature"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default values if not provided
	if req.MaxLength <= 0 {
		req.MaxLength = 100
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}

	ctx := c.Request.Context()
	texts, err := h.aiService.GenerateText(ctx, req.ModelName, req.Prompt, req.MaxLength, req.Temperature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"texts": texts,
	})
}

// GenerateEmbedding handles the embedding generation request
func (h *AIHandler) GenerateEmbedding(c *gin.Context) {
	var req struct {
		ModelName string   `json:"model_name" binding:"required"`
		Texts     []string `json:"texts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	embeddings, err := h.aiService.GenerateEmbedding(ctx, req.ModelName, req.Texts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"embeddings": embeddings,
	})
}

// ListModels handles the list models request
func (h *AIHandler) ListModels(c *gin.Context) {
	ctx := c.Request.Context()
	models, err := h.aiService.ListModels(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}

// LoadModel handles the load model request
func (h *AIHandler) LoadModel(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Model name is required",
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.LoadModel(ctx, modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model loaded successfully",
	})
}

// GetModelStatus handles the get model status request
func (h *AIHandler) GetModelStatus(c *gin.Context) {
	modelName := c.Param("name")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Model name is required",
		})
		return
	}

	ctx := c.Request.Context()
	isLoaded, loadTime, err := h.aiService.GetModelStatus(ctx, modelName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_loaded": isLoaded,
		"load_time": loadTime,
	})
}

// DescribeCollection handles the describe collection request
func (h *AIHandler) DescribeCollection(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	ctx := c.Request.Context()
	collection, err := h.aiService.DescribeCollection(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"collection": collection,
	})
}

// LoadCollection handles the load collection request
func (h *AIHandler) LoadCollection(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	var req struct {
		ReplicaNumber int64 `json:"replica_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default replica number if not provided
	if req.ReplicaNumber <= 0 {
		req.ReplicaNumber = 1
	}

	ctx := c.Request.Context()
	err := h.aiService.LoadCollection(ctx, name, req.ReplicaNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection loaded successfully",
	})
}

// ReleaseCollection handles the release collection request
func (h *AIHandler) ReleaseCollection(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.ReleaseCollection(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection released successfully",
	})
}

// GetCollectionStatistics handles the get collection statistics request
func (h *AIHandler) GetCollectionStatistics(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	ctx := c.Request.Context()
	stats, err := h.aiService.GetCollectionStatistics(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
	})
}

// CreateIndex handles the create index request
func (h *AIHandler) CreateIndex(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	var req struct {
		IndexName string            `json:"index_name" binding:"required"`
		FieldName string            `json:"field_name" binding:"required"`
		IndexType pb.IndexType      `json:"index_type" binding:"required"`
		MetricType map[string]string `json:"metric_type"`
		Params     map[string]string `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	indexParams := &pb.IndexParams{
		IndexName: req.IndexName,
		FieldName: req.FieldName,
		IndexType: req.IndexType,
		MetricType: req.MetricType,
		Params: req.Params,
	}

	ctx := c.Request.Context()
	err := h.aiService.CreateIndex(ctx, name, indexParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Index created successfully",
	})
}

// DropIndex handles the drop index request
func (h *AIHandler) DropIndex(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	indexName := c.Param("indexName")
	if indexName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Index name is required",
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.DropIndex(ctx, name, indexName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Index dropped successfully",
	})
}

// DescribeIndex handles the describe index request
func (h *AIHandler) DescribeIndex(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	var req struct {
		IndexName string `json:"index_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	indexes, err := h.aiService.DescribeIndex(ctx, name, req.IndexName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"indexes": indexes,
	})
}

// Delete handles the delete data request
func (h *AIHandler) Delete(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	var req struct {
		Expr string `json:"expr" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	ids, err := h.aiService.Delete(ctx, name, req.Expr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ids": ids,
	})
}

// GetById handles the get data by IDs request
func (h *AIHandler) GetById(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	var req struct {
		IDs          []string `json:"ids" binding:"required"`
		OutputFields []string `json:"output_fields"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	results, err := h.aiService.GetById(ctx, name, req.IDs, req.OutputFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// Compact handles the compact collection request
func (h *AIHandler) Compact(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Collection name is required",
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.Compact(ctx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Collection compacted successfully",
	})
}

// RegisterModel handles the register model request
func (h *AIHandler) RegisterModel(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Provider    pb.ModelProvider  `json:"provider" binding:"required"`
		ModelPath   string            `json:"model_path" binding:"required"`
		Description string            `json:"description"`
		Config      map[string]string `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.RegisterModel(ctx, req.Name, req.Provider, req.ModelPath, req.Description, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model registered successfully",
	})
}

// UpdateModel handles the update model request
func (h *AIHandler) UpdateModel(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		ModelPath   string            `json:"model_path"`
		Description string            `json:"description"`
		Config      map[string]string `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.UpdateModel(ctx, req.Name, req.ModelPath, req.Description, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model updated successfully",
	})
}

// UnregisterModel handles the unregister model request
func (h *AIHandler) UnregisterModel(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.UnregisterModel(ctx, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model unregistered successfully",
	})
}

// GetModel handles the get model request
func (h *AIHandler) GetModel(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	model, err := h.aiService.GetModel(ctx, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model": model,
	})
}

// UnloadModel handles the unload model request
func (h *AIHandler) UnloadModel(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	err := h.aiService.UnloadModel(ctx, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Model unloaded successfully",
	})
}