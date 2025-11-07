package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/api/internal/config"
	"github.com/codetaoist/taishanglaojun/api/internal/dao"
	"github.com/codetaoist/taishanglaojun/api/internal/middleware"
	"github.com/codetaoist/taishanglaojun/api/internal/models"
	"github.com/codetaoist/taishanglaojun/api/internal/response"
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

func SetupTaishang(cfg *config.Config, r *gin.Engine) {
	// Initialize DAOs
	modelDAO := dao.NewModelDAO(cfg.DB)
	collectionDAO := dao.NewVectorCollectionDAO(cfg.DB)
	taskDAO := dao.NewTaskDAO(cfg.DB)

	// Apply authentication middleware
	authMiddleware := middleware.Auth(cfg.JWT.Secret)

	taishang := r.Group("/api/taishang")
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
		taishang.POST("/collections/:id/upsert", upsertVectors(collectionDAO))
		taishang.POST("/collections/:id/query", queryVectors(collectionDAO))
		taishang.DELETE("/collections/:id/vectors", deleteVectors(collectionDAO))
	}
	
	// Task management
	{
		taishang.GET("/tasks/list", listTasks(taskDAO))
		taishang.POST("/tasks", createTask(taskDAO))
		taishang.GET("/tasks/:id", getTask(taskDAO))
		taishang.PUT("/tasks/:id", updateTask(taskDAO))
		taishang.DELETE("/tasks/:id", deleteTask(taskDAO))
	}
}

// --- Handlers for Taishang Domain ---

func listModels(modelDAO *dao.ModelDAO) gin.HandlerFunc {
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

		models, total, err := modelDAO.List(c.Request.Context(), tenantID, status, page, pageSize)
		if err != nil {
			response.InternalServerError(c, "Failed to list models")
			return
		}

		data := ModelListData{
			Total:    total,
			Page:     page,
			PageSize: pageSize,
			Items:    models,
		}

		response.Success(c, data)
	}
}

func registerModel(modelDAO *dao.ModelDAO) gin.HandlerFunc {
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

func getModel(modelDAO *dao.ModelDAO) gin.HandlerFunc {
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

func updateModel(modelDAO *dao.ModelDAO) gin.HandlerFunc {
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
		model.Params = req.Params

		// Update model
		if err := modelDAO.Update(c.Request.Context(), model); err != nil {
			response.InternalServerError(c, "Failed to update model")
			return
		}

		response.Success(c, model)
	}
}

func deleteModel(modelDAO *dao.ModelDAO) gin.HandlerFunc {
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
			Items:    collections,
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

func upsertVectors(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
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

		var req UpsertVectorsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid upsert vectors request")
			return
		}

		// Check if collection exists
		_, err = collectionDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Collection not found")
			return
		}

		// TODO: Implement vector upsert logic
		resp := gin.H{
			"upserted": len(req.Vectors),
			"collectionId": id,
			"namespace": req.Namespace,
		}

		response.Success(c, resp)
	}
}

func queryVectors(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
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

		var req QueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid query vectors request")
			return
		}

		// Check if collection exists
		_, err = collectionDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Collection not found")
			return
		}

		// TODO: Implement vector query logic
		matches := []Match{
			{
				ID:    "vec-1",
				Score: 0.95,
				Metadata: map[string]interface{}{
					"title": "Document 1",
				},
			},
			{
				ID:    "vec-2",
				Score: 0.85,
				Metadata: map[string]interface{}{
					"title": "Document 2",
				},
			},
		}

		resp := gin.H{
			"matches": matches,
			"count":   len(matches),
			"collectionId": id,
			"namespace": req.Namespace,
		}

		response.Success(c, resp)
	}
}

func deleteVectors(collectionDAO *dao.VectorCollectionDAO) gin.HandlerFunc {
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

		var req DeleteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid delete vectors request")
			return
		}

		// Check if collection exists
		_, err = collectionDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Collection not found")
			return
		}

		// TODO: Implement vector delete logic
		resp := gin.H{
			"deleted": len(req.IDs),
			"collectionId": id,
			"namespace": req.Namespace,
		}

		response.Success(c, resp)
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

		tasks, total, err := taskDAO.List(c.Request.Context(), tenantID, status, page, pageSize)
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
			req.Status = string(TaskStatusPending)
		}
		if req.Priority == "" {
			req.Priority = string(PriorityNormal)
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
		task, err := taskDAO.GetByID(c.Request.Context(), tenantID, id)
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
		
		// Check if task exists
		task, err := taskDAO.GetByID(c.Request.Context(), tenantID, id)
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
		
		// Check if task exists
		_, err := taskDAO.GetByID(c.Request.Context(), tenantID, id)
		if err != nil {
			badRequest(c, "Task not found")
			return
		}

		// Delete task
		if err := taskDAO.Delete(c.Request.Context(), tenantID, id); err != nil {
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