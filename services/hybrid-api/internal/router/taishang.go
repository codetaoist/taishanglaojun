package router

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/middleware"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/response"
	"github.com/gin-gonic/gin"
)

// --- Types aligned with OpenAPI for Taishang Domain ---

type ModelStatus string

const (
	ModelStatusEnabled  ModelStatus = "enabled"
	ModelStatusDisabled ModelStatus = "disabled"
)

type Model struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Status  ModelStatus `json:"status"`
	Params  map[string]string `json:"params,omitempty"`
}

type ModelListData struct {
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
	Items    []Model `json:"items"`
}

type RegisterModelRequest struct {
	Name         string            `json:"name" binding:"required"`
	Version      string            `json:"version" binding:"required"`
	Family       string            `json:"family"`
	Quantization string            `json:"quantization"`
	Params       map[string]string `json:"params"`
}

type IndexType string

const (
	IndexTypeHNSW IndexType = "HNSW"
	IndexTypeIVF  IndexType = "IVF"
	IndexTypeFLAT IndexType = "FLAT"
)

type MetricType string

const (
	MetricTypeCosine MetricType = "cosine"
	MetricTypeL2     MetricType = "l2"
	MetricTypeDot    MetricType = "dot"
)

type VectorCollection struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Dim       int        `json:"dim"`
	IndexType IndexType  `json:"indexType"`
	Metric    MetricType `json:"metric"`
}

type VectorCollectionList struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
	Items    []VectorCollection `json:"items"`
}

type CreateVectorCollectionRequest struct {
	Name       string     `json:"name" binding:"required"`
	Dim        int        `json:"dim" binding:"required"`
	IndexType  IndexType  `json:"indexType" binding:"required"`
	Metric     MetricType `json:"metric" binding:"required"`
	Replication int       `json:"replication"`
}

type UpsertVector struct {
	ID       string                 `json:"id"`
	Values   []float64              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UpsertVectorsRequest struct {
	Vectors   []UpsertVector `json:"vectors" binding:"required"`
	Namespace string         `json:"namespace"`
}

type QueryRequest struct {
	TopK      int                    `json:"topK" binding:"required"`
	Query     map[string]interface{} `json:"query" binding:"required"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
	Namespace string                 `json:"namespace"`
}

type DeleteRequest struct {
	IDs       []string               `json:"ids"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
	Namespace string                 `json:"namespace"`
}

type Match struct {
	ID       string                 `json:"id"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusSucceeded TaskStatus = "SUCCEEDED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusCanceled  TaskStatus = "CANCELED"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
)

type Task struct {
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Status   TaskStatus `json:"status"`
	Priority Priority   `json:"priority"`
	Payload  interface{} `json:"payload,omitempty"`
	Result   interface{} `json:"result,omitempty"`
}

// --- Response helpers for Taishang Domain ---
// Note: Using unified response package instead of custom helpers

// --- Taishang Domain Router Setup ---

func SetupTaishangRoutes(cfg *config.Config, r *gin.Engine) {
	// Initialize DAOs
	// Convert sql.DB to gorm.DB
	gormDB, err := dao.NewGormDB(cfg.DB)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize GORM DB: %v", err))
	}
	
	modelDAO := dao.NewTaishangModelDAO(gormDB)
	collectionDAO := dao.NewVectorCollectionDAO(gormDB)
	taskDAO := dao.NewTaskDAO(gormDB)
	vectorDAO := dao.NewVectorDAO(gormDB, nil) // We'll need to inject the vector service

	// Apply authentication middleware
	authMiddleware := middleware.Auth(*cfg)

	taishang := r.Group("/api/v1/taishang")
	taishang.Use(authMiddleware)
	
	// Model management
	{
		taishang.GET("/models/list", listModels(modelDAO))
		taishang.POST("/models/register", registerModel(modelDAO))
		taishang.GET("/models/:id", getModel(modelDAO))
		taishang.PUT("/models/:id", updateModel(modelDAO))
		taishang.DELETE("/models/:id", deleteModel(modelDAO))
	}
	
	// Vector collection management
	{
		taishang.GET("/collections/list", listCollections(collectionDAO))
		taishang.POST("/collections", createCollection(collectionDAO))
		taishang.GET("/collections/:id", getCollection(collectionDAO))
		taishang.DELETE("/collections/:id", deleteCollection(collectionDAO))
		taishang.POST("/collections/:id/upsert", upsertVectors(collectionDAO, vectorDAO))
		taishang.POST("/collections/:id/query", queryVectors(collectionDAO, vectorDAO))
		taishang.DELETE("/collections/:id/vectors", deleteVectors(collectionDAO, vectorDAO))
	}
	
	// Task management
	{
		taishang.GET("/tasks/list", listTasks(taskDAO))
		taishang.POST("/tasks", createTask(taskDAO))
		taishang.GET("/tasks/:id", getTask(taskDAO))
		taishang.PUT("/tasks/:id", updateTask(taskDAO))
		taishang.DELETE("/tasks/:id", deleteTask(taskDAO))
	}
	
	// Vector database operations
	{
		taishang.GET("/vector/status", getVectorDatabaseStatus())
		taishang.POST("/vector/connect", connectVectorDatabase())
		taishang.GET("/vector/collections", listVectorCollections())
		taishang.POST("/vector/collections", createVectorCollection())
		taishang.GET("/vector/collections/:name", getVectorCollection())
		taishang.DELETE("/vector/collections/:name", deleteVectorCollection())
		taishang.GET("/vector/collections/:name/stats", getVectorCollectionStats())
		taishang.POST("/vector/collections/:name/index", createVectorIndex())
		taishang.DELETE("/vector/collections/:name/index", deleteVectorIndex())
		taishang.POST("/vector/collections/:name/vectors", upsertVectorData())
		taishang.POST("/vector/collections/:name/search", searchVectorData(vectorDAO))
		taishang.GET("/vector/collections/:name/vectors/:id", getVectorData())
		taishang.DELETE("/vector/collections/:name/vectors/:id", deleteVectorData())
		taishang.POST("/vector/collections/:name/vectors/batch", batchDeleteVectorData())
		taishang.GET("/vector/health", vectorHealthCheck())
		taishang.GET("/vector/info", getVectorDatabaseInfo())
	}
}

// Product management
func SetupProducts(cfg *config.Config, r *gin.Engine) {
	// Apply authentication middleware
	authMiddleware := middleware.Auth(*cfg)
	
	// Product routes
	productGroup := r.Group("/api/v1/products")
	productGroup.Use(authMiddleware)
	{
		productGroup.GET("", listProducts)
		productGroup.POST("", createProduct)
		productGroup.GET("/:id", getProduct)
		productGroup.PUT("/:id", updateProduct)
		productGroup.DELETE("/:id", deleteProduct)
	}
}

// --- Product Handlers ---

type Product struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       float64                `json:"price"`
	Category    string                 `json:"category"`
	ImageURL    string                 `json:"imageUrl"`
	Tags        []string               `json:"tags"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

type ProductListResponse struct {
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Items    []Product `json:"items"`
}

func listProducts(c *gin.Context) {
	// Mock data for now
	products := []Product{
		{
			ID:          1,
			Name:        "Sample Product 1",
			Description: "This is a sample product description",
			Price:       99.99,
			Category:    "electronics",
			ImageURL:    "https://example.com/image1.jpg",
			Tags:        []string{"popular", "new"},
			CreatedAt:   "2023-01-01T00:00:00Z",
			UpdatedAt:   "2023-01-01T00:00:00Z",
		},
		{
			ID:          2,
			Name:        "Sample Product 2",
			Description: "This is another sample product description",
			Price:       149.99,
			Category:    "home",
			ImageURL:    "https://example.com/image2.jpg",
			Tags:        []string{"sale", "featured"},
			CreatedAt:   "2023-01-02T00:00:00Z",
			UpdatedAt:   "2023-01-02T00:00:00Z",
		},
	}

	response := ProductListResponse{
		Total:    len(products),
		Page:     1,
		PageSize: 10,
		Items:    products,
	}

	c.JSON(200, gin.H{
		"code":    "SUCCESS",
		"data":    response,
		"message": "Products retrieved successfully",
	})
}

func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid product data",
		})
		return
	}

	// In a real implementation, save to database
	product.ID = 3 // Mock ID
	product.CreatedAt = "2023-01-03T00:00:00Z"
	product.UpdatedAt = "2023-01-03T00:00:00Z"

	c.JSON(201, gin.H{
		"code":    "SUCCESS",
		"data":    product,
		"message": "Product created successfully",
	})
}

func getProduct(c *gin.Context) {
	// id := c.Param("id")
	
	// Mock product
	product := Product{
		ID:          1,
		Name:        "Sample Product 1",
		Description: "This is a sample product description",
		Price:       99.99,
		Category:    "electronics",
		ImageURL:    "https://example.com/image1.jpg",
		Tags:        []string{"popular", "new"},
		CreatedAt:   "2023-01-01T00:00:00Z",
		UpdatedAt:   "2023-01-01T00:00:00Z",
	}

	c.JSON(200, gin.H{
		"code":    "SUCCESS",
		"data":    product,
		"message": "Product retrieved successfully",
	})
}

func updateProduct(c *gin.Context) {
	id := c.Param("id")
	
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "Invalid product data",
		})
		return
	}

	// In a real implementation, update in database
	product.ID, _ = strconv.Atoi(id)
	product.UpdatedAt = "2023-01-04T00:00:00Z"

	c.JSON(200, gin.H{
		"code":    "SUCCESS",
		"data":    product,
		"message": "Product updated successfully",
	})
}

func deleteProduct(c *gin.Context) {
	// id := c.Param("id")
	
	// In a real implementation, delete from database
	
	c.JSON(200, gin.H{
		"code":    "SUCCESS",
		"data":    gin.H{"id": "1"},
		"message": "Product deleted successfully",
	})
}

// --- Handlers for Taishang Domain ---

func listModels(modelDAO *dao.TaishangModelDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default" // Fallback tenant
		}

		// Parse query parameters
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		models, total, err := modelDAO.List(c.Request.Context(), tenantID, status, "", page, pageSize)
		if err != nil {
			response.InternalServerError(c, "Failed to list models")
			return
		}

		data := ModelListData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Items:    convertModels(models),
		}

		response.Success(c, data)
	}
}

func registerModel(modelDAO *dao.TaishangModelDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req models.Model
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid register model request")
			return
		}

		// Set tenant ID and default values
		req.TenantID = tenantID
		if req.Status == "" {
			req.Status = string(ModelStatusEnabled)
		}

		// Create model
		if err := modelDAO.Create(c.Request.Context(), &req); err != nil {
			response.InternalServerError(c, "Failed to register model")
			return
		}

		response.Success(c, req)
	}
}

func getModel(modelDAO *dao.TaishangModelDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id := c.Param("id")
		model, err := modelDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			response.NotFound(c, "Model not found")
			return
		}

		response.Success(c, model)
	}
}

func updateModel(modelDAO *dao.TaishangModelDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id := c.Param("id")
		
		// Check if model exists
		model, err := modelDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Model not found")
			return
		}

		var req models.Model
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid update model request")
			return
		}

		// Update fields
		model.Name = req.Name
		model.Version = req.Version
		model.Status = req.Status
		model.Meta = req.Meta

		// Update model
		if err := modelDAO.Update(c.Request.Context(), model); err != nil {
			response.InternalServerError(c, "Failed to update model")
			return
		}

		response.Success(c, model)
	}
}

func deleteModel(modelDAO *dao.TaishangModelDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id := c.Param("id")
		
		// Check if model exists
		_, err := modelDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Model not found")
			return
		}

		// Delete model
		if err := modelDAO.Delete(c.Request.Context(), tenantID, id); err != nil {
			response.InternalServerError(c, "Failed to delete model")
			return
		}

		resp := gin.H{
			"deleted": true,
			"id":      id,
		}

		response.Success(c, resp)
	}
}

func listCollections(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		collections, total, err := collectionDAO.List(c.Request.Context(), tenantID, "", page, pageSize)
		if err != nil {
			response.InternalServerError(c, "Failed to list collections")
			return
		}

		data := VectorCollectionList{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Items:    convertCollections(collections),
		}

		response.Success(c, data)
	}
}

func createCollection(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req models.VectorCollection
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid create collection request")
			return
		}

		// Set tenant ID
		req.TenantID = tenantID

		// Create collection
		if err := collectionDAO.Create(c.Request.Context(), &req); err != nil {
			response.InternalServerError(c, "Failed to create collection")
			return
		}

		response.Success(c, req)
	}
}

func getCollection(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid collection ID")
			return
		}

		collection, err := collectionDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			response.NotFound(c, "Collection not found")
			return
		}

		response.Success(c, collection)
	}
}

func deleteCollection(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			badRequest(c, "Invalid collection ID")
			return
		}

		// Check if collection exists
		_, err = collectionDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Collection not found")
			return
		}

		// Delete collection
		if err := collectionDAO.Delete(c.Request.Context(), tenantID, id); err != nil {
			response.InternalServerError(c, "Failed to delete collection")
			return
		}

		resp := gin.H{
			"deleted": true,
			"id":      id,
		}

		response.Success(c, resp)
	}
}

func upsertVectors(collectionDAO *dao.VectorCollectionDAO, vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionIDStr := c.Param("collection_id")
		collectionID, err := strconv.Atoi(collectionIDStr)
		if err != nil {
			response.BadRequest(c, "Invalid collection ID")
			return
		}

		// Check if collection exists
		collection, err := collectionDAO.GetByID(c.Request.Context(), tenantID, collectionID)
		if err != nil {
			response.NotFound(c, "Collection not found")
			return
		}

		var req models.UpsertVectorsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}

		// Convert vectors to the expected format
		vectors := make([]models.VectorData, 0, len(req.Vectors))
		for _, v := range req.Vectors {
			vectors = append(vectors, models.VectorData{
				ID:         v.ID,
				Vector:     v.Vector,
				Metadata:   v.Metadata,
				Embedding:  v.Embedding,
				ExternalID: v.ExternalID,
			})
		}

		// Get collection name for vector service
		collectionName := fmt.Sprintf("tai_collection_%d", collectionID)

		// Upsert vectors
		upsertReq := &models.UpsertVectorsRequest{
			CollectionName: collectionName,
			Vectors:        vectors,
		}
		_, err = vectorDAO.UpsertVectors(c.Request.Context(), upsertReq)
		if err != nil {
			response.InternalServerError(c, "Failed to upsert vectors: "+err.Error())
			return
		}

		response.Success(c, gin.H{
			"message":        "Vectors upserted successfully",
			"collection_id":   collectionID,
			"collection_name": collection.Name,
			"vectors_count":   len(req.Vectors),
		})
	}
}

func queryVectors(collectionDAO *dao.VectorCollectionDAO, vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionIDStr := c.Param("collection_id")
		collectionID, err := strconv.Atoi(collectionIDStr)
		if err != nil {
			response.BadRequest(c, "Invalid collection ID")
			return
		}

		// Check if collection exists
		collection, err := collectionDAO.GetByID(c.Request.Context(), tenantID, collectionID)
		if err != nil {
			response.NotFound(c, "Collection not found")
			return
		}

		var req models.SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}

		// Get collection name for vector service
		collectionName := fmt.Sprintf("tai_collection_%d", collectionID)

		// Set collection name in search request
		req.CollectionName = collectionName

		// Query vectors
		searchResponse, err := vectorDAO.SearchVectorData(c.Request.Context(), collectionName, &req)
		if err != nil {
			response.InternalServerError(c, "Failed to query vectors: "+err.Error())
			return
		}

		response.Success(c, gin.H{
			"message":        "Vectors queried successfully",
			"collection_id":   collectionID,
			"collection_name": collection.Name,
			"results":         searchResponse.Results,
			"total":           searchResponse.Total,
		})
	}
}

func deleteVectors(collectionDAO *dao.VectorCollectionDAO, vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionIDStr := c.Param("collection_id")
		collectionID, err := strconv.Atoi(collectionIDStr)
		if err != nil {
			response.BadRequest(c, "Invalid collection ID")
			return
		}

		// Check if collection exists
		collection, err := collectionDAO.GetByID(c.Request.Context(), tenantID, collectionID)
		if err != nil {
			response.NotFound(c, "Collection not found")
			return
		}

		var req models.DeleteVectorsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}

		// Get collection name for vector service
		collectionName := fmt.Sprintf("tai_collection_%d", collectionID)

		// Set collection name in delete request
		req.CollectionName = collectionName

		// Delete vectors
		err = vectorDAO.DeleteVectorsByCollection(c.Request.Context(), tenantID, collectionID, req.Ids)
		if err != nil {
			response.InternalServerError(c, "Failed to delete vectors: "+err.Error())
			return
		}

		response.Success(c, gin.H{
			"message":        "Vectors deleted successfully",
			"collection_id":   collectionID,
			"collection_name": collection.Name,
			"deleted_count":   len(req.Ids),
		})
	}
}

func listTasks(taskDAO *dao.TaskDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Parse query parameters
		status := c.Query("status")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		tasks, total, err := taskDAO.List(c.Request.Context(), tenantID, models.TaskStatus(status), page, pageSize)
		if err != nil {
			response.InternalServerError(c, "Failed to list tasks")
			return
		}

		data := gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"items":    tasks,
		}

		response.Success(c, data)
	}
}

func createTask(taskDAO *dao.TaskDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req models.Task
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid create task request")
			return
		}

		// Set tenant ID and default values
		req.TenantID = tenantID
		if req.Status == "" {
			req.Status = models.TaskStatusPending
		}
		if req.Priority == "" {
			req.Priority = models.TaskPriorityNormal
		}

		// Create task
		if err := taskDAO.Create(c.Request.Context(), &req); err != nil {
			response.InternalServerError(c, "Failed to create task")
			return
		}

		response.Success(c, req)
	}
}

func getTask(taskDAO *dao.TaskDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id := c.Param("id")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid task ID")
			return
		}
		task, err := taskDAO.GetByID(c.Request.Context(), tenantID, idInt)
		if err != nil {
			response.NotFound(c, "Task not found")
			return
		}

		response.Success(c, task)
	}
}

func updateTask(taskDAO *dao.TaskDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id := c.Param("id")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid task ID")
			return
		}
		
		// Check if task exists
		task, err := taskDAO.GetByID(c.Request.Context(), tenantID, idInt)
		if err != nil {
			badRequest(c, "Task not found")
			return
		}

		var req models.Task
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid update task request")
			return
		}

		// Update fields
		task.Type = req.Type
		task.Status = req.Status
		task.Priority = req.Priority
		task.Payload = req.Payload
		task.Result = req.Result

		// Update task
		if err := taskDAO.Update(c.Request.Context(), task); err != nil {
			response.InternalServerError(c, "Failed to update task")
			return
		}

		response.Success(c, task)
	}
}

func deleteTask(taskDAO *dao.TaskDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		id := c.Param("id")
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.BadRequest(c, "Invalid task ID")
			return
		}
		
		// Check if task exists
		_, err = taskDAO.GetByID(c.Request.Context(), tenantID, idInt)
		if err != nil {
			badRequest(c, "Task not found")
			return
		}

		// Delete task
		if err := taskDAO.Delete(c.Request.Context(), tenantID, idInt); err != nil {
			response.InternalServerError(c, "Failed to delete task")
			return
		}

		resp := gin.H{
			"deleted": true,
			"id":      id,
		}

		response.Success(c, resp)
	}
}

// Vector database handlers

func getVectorDatabaseStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Mock status for now
		status := gin.H{
			"connected": true,
			"type":      "milvus",
			"host":      "localhost",
			"port":      19530,
		}

		response.Success(c, status)
	}
}

func connectVectorDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req models.VectorDatabaseConfig
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid vector database configuration")
			return
		}

		// Mock connection for now
		resp := gin.H{
			"connected": true,
			"type":      req.Type,
			"host":      req.Host,
			"port":      req.Port,
		}

		response.Success(c, resp)
	}
}

func listVectorCollections() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Mock collections for now
		collections := []string{
			"documents",
			"images",
			"products",
		}

		response.Success(c, gin.H{
			"collections": collections,
			"total":       len(collections),
		})
	}
}

func createVectorCollection() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req models.VectorCollectionConfig
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid collection configuration")
			return
		}

		// Mock creation for now
		resp := gin.H{
			"created": true,
			"name":    req.CollectionName,
			"dimension": req.Dimension,
			"indexType": req.IndexType,
			"metricType": req.MetricType,
		}

		response.Success(c, resp)
	}
}

func getVectorCollection() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Mock collection for now
		collection := gin.H{
			"name": collectionName,
			"dimension": 1536,
			"indexType": "HNSW",
			"metricType": "cosine",
			"vectorCount": 1000,
		}

		response.Success(c, collection)
	}
}

func deleteVectorCollection() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Mock deletion for now
		resp := gin.H{
			"deleted": true,
			"name":    collectionName,
		}

		response.Success(c, resp)
	}
}

func getVectorCollectionStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Mock stats for now
		stats := models.VectorCollectionStats{
			CollectionName: collectionName,
			VectorCount:    1000,
			IndexSize:      1024000,
			LastUpdated:    time.Now(),
		}

		response.Success(c, stats)
	}
}

func createVectorIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		var req models.VectorIndex
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid index configuration")
			return
		}

		// Mock creation for now
		resp := gin.H{
			"created": true,
			"collectionName": collectionName,
			"indexName": req.IndexName,
			"indexType": req.IndexType,
			"metricType": req.MetricType,
		}

		response.Success(c, resp)
	}
}

func deleteVectorIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Mock deletion for now
		resp := gin.H{
			"deleted": true,
		}

		response.Success(c, resp)
	}
}

func upsertVectorData() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req struct {
			Vectors []models.VectorData `json:"vectors" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid vector data")
			return
		}

		// Mock upsert for now
		resp := gin.H{
			"upserted": len(req.Vectors),
		}

		response.Success(c, resp)
	}
}

func searchVectorData(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		var req models.VectorSearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid search request")
			return
		}

		// Set collection name in search request
		req.CollectionName = collectionName

		// Perform vector search
		searchResponse, err := vectorDAO.SearchVectorData(c.Request.Context(), collectionName, &req)
		if err != nil {
			response.InternalServerError(c, "Failed to search vectors: "+err.Error())
			return
		}

		response.Success(c, searchResponse)
	}
}

func getVectorData() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		vectorID := c.Param("id")

		// Mock vector for now
		vector := models.VectorData{
			ID:     vectorID,
			Vector: make([]float64, 1536), // Mock vector
			Metadata: map[string]interface{}{
				"title": "Document 1",
			},
		}

		response.Success(c, vector)
	}
}

func deleteVectorData() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		vectorID := c.Param("id")

		// Mock deletion for now
		resp := gin.H{
			"deleted": true,
			"id":      vectorID,
		}

		response.Success(c, resp)
	}
}

func batchDeleteVectorData() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		var req struct {
			IDs []string `json:"ids" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request")
			return
		}

		// Mock batch deletion for now
		resp := gin.H{
			"deleted": len(req.IDs),
		}

		response.Success(c, resp)
	}
}

func vectorHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Mock health check for now
		status := gin.H{
			"status": "healthy",
			"timestamp": time.Now(),
		}

		response.Success(c, status)
	}
}

func getVectorDatabaseInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Mock database info for now
		info := models.VectorDatabaseInfo{
			Type:        "milvus",
			Version:     "2.3.0",
			Collections: []string{"documents", "images", "products"},
		}

		response.Success(c, info)
	}
}