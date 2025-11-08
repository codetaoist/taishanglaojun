package router

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/middleware"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/response"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/vector"
	"github.com/gin-gonic/gin"
)

// SetupVectorRoutes sets up the vector database routes
func SetupVectorRoutes(cfg *config.Config, r *gin.Engine) {
	// Initialize DAOs
	// Convert sql.DB to gorm.DB
	gormDB, err := dao.NewGormDB(cfg.DB)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize GORM DB: %v", err))
	}
	
	// Convert port string to int
	vectorPort, err := strconv.Atoi(cfg.VectorDBPort)
	if err != nil {
		panic(fmt.Sprintf("Invalid VectorDBPort: %v", err))
	}
	
	// Initialize vector database service
	vectorConfig := &vector.DatabaseConfig{
		Type: vector.DatabaseType(cfg.VectorDBType),
		Milvus: &vector.MilvusConfig{
			Address:  cfg.VectorDBHost,
			Port:     vectorPort,
			Username: cfg.VectorDBUser,
			Password: cfg.VectorDBPassword,
			Database: cfg.VectorDBDatabase,
		},
	}
	
	vectorService, err := vector.NewService(vectorConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize vector service: %v", err))
	}
	
	vectorDAO := dao.NewVectorDAO(gormDB, vectorService)

	// Apply authentication middleware
	authMiddleware := middleware.Auth(*cfg)

	vectorGroup := r.Group("/api/v1/vector")
	vectorGroup.Use(authMiddleware)

	// Vector database management
	{
		vectorGroup.GET("/status", getVectorDatabaseStatusV1(vectorDAO))
		vectorGroup.POST("/connect", connectVectorDatabaseV1(vectorDAO))
		vectorGroup.GET("/collections", listVectorCollectionsV1(vectorDAO))
		vectorGroup.POST("/collections", createVectorCollectionV1(vectorDAO))
		vectorGroup.GET("/collections/:name", getVectorCollectionV1(vectorDAO))
		vectorGroup.DELETE("/collections/:name", deleteVectorCollectionV1(vectorDAO))
		vectorGroup.GET("/collections/:name/stats", getVectorCollectionStatsV1(vectorDAO))
		vectorGroup.POST("/collections/:name/index", createVectorIndexV1(vectorDAO))
		vectorGroup.DELETE("/collections/:name/index", deleteVectorIndexV1(vectorDAO))
	}

	// Vector operations
	{
		vectorGroup.POST("/collections/:name/vectors", upsertVectorsV1(vectorDAO))
		vectorGroup.POST("/collections/:name/search", searchVectors(vectorDAO))
		vectorGroup.GET("/collections/:name/vectors/:id", getVector(vectorDAO))
		vectorGroup.DELETE("/collections/:name/vectors/:id", deleteVector(vectorDAO))
		vectorGroup.POST("/collections/:name/vectors/batch", batchDeleteVectors(vectorDAO))
	}

	// Vector database operations
	{
		vectorGroup.GET("/health", healthCheck(vectorDAO))
		vectorGroup.GET("/info", getVectorDatabaseInfoV1(vectorDAO))
	}
}

// Vector database handlers

func getVectorDatabaseStatusV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Check vector database health status
		err := vectorDAO.VectorHealthCheck(c.Request.Context())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Vector database is not healthy", err.Error())
			return
		}

		// Get vector database status
		status, err := vectorDAO.GetVectorDatabaseStatus(c.Request.Context())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to get vector database status", err.Error())
			return
		}

		response.Success(c, status)
	}
}

func connectVectorDatabaseV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
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

		// Convert models.VectorDatabaseConfig to models.VectorDatabaseConfig
		vectorConfig := &models.VectorDatabaseConfig{
			Type:     req.Type,
			Host:     req.Host,
			Port:     req.Port,
			Username: req.Username,
			Password: req.Password,
			Database: req.Database,
		}

		// Test connection with the new configuration
		err := vectorDAO.TestConnection(c.Request.Context(), vectorConfig)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to connect to vector database", err.Error())
			return
		}

		resp := gin.H{
			"connected": true,
			"type":      req.Type,
			"host":      req.Host,
			"port":      req.Port,
		}

		response.Success(c, resp)
	}
}

func listVectorCollectionsV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		// Get collections from vector database
		collections, err := vectorDAO.ListCollections(c.Request.Context(), tenantID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to list collections", err.Error())
			return
		}

		response.Success(c, gin.H{
			"collections": collections,
			"total":       len(collections),
		})
	}
}

func createVectorCollectionV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
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

		// Create collection in vector database
		collection := &models.Collection{
			Name:        req.CollectionName,
			Description: "", // VectorCollectionConfig doesn't have Description field
			Dimension:   req.Dimension,
			MetricType:  req.MetricType,
		}
		
		createdCollection, err := vectorDAO.CreateCollection(c.Request.Context(), tenantID, collection)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to create collection", err.Error())
			return
		}

		response.Success(c, createdCollection)
	}
}

func getVectorCollectionV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Get collection from vector database
		collection, err := vectorDAO.GetCollection(c.Request.Context(), tenantID, collectionName)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to get collection", err.Error())
			return
		}

		response.Success(c, collection)
	}
}

func deleteVectorCollectionV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Delete collection from vector database
		err := vectorDAO.DeleteCollection(c.Request.Context(), tenantID, collectionName)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to delete collection", err.Error())
			return
		}

		resp := gin.H{
			"deleted": true,
			"name":    collectionName,
		}

		response.Success(c, resp)
	}
}

func getVectorCollectionStatsV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Get collection stats from vector database
		// First, get the collection by name to get its ID
		collection, err := vectorDAO.GetCollection(c.Request.Context(), tenantID, collectionName)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to get collection", err.Error())
			return
		}

		// Get collection stats
		stats, err := vectorDAO.GetCollectionStats(c.Request.Context(), tenantID, int(collection.ID))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to get collection stats", err.Error())
			return
		}

		response.Success(c, stats)
	}
}

func createVectorIndexV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
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

		// Create index in vector database
		indexRequest := &models.CreateIndexRequest{
			CollectionName: collectionName,
			FieldName:      req.FieldName,
			IndexType:      req.IndexType,
			MetricType:     req.MetricType,
			Params:         req.Params,
		}
		
		err := vectorDAO.CreateVectorIndex(c.Request.Context(), collectionName, indexRequest)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to create index", err.Error())
			return
		}

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

func deleteVectorIndexV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		// Delete index from vector database
		err := vectorDAO.DeleteVectorIndex(c.Request.Context(), collectionName, "vector")
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to delete index", err.Error())
			return
		}

		resp := gin.H{
			"deleted": true,
			"collectionName": collectionName,
		}

		response.Success(c, resp)
	}
}

// Vector operation handlers

func upsertVectorsV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		var req models.UpsertVectorsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}

		// Set collection name from URL parameter
		req.CollectionName = collectionName

		// Upsert vectors in vector database
		result, err := vectorDAO.UpsertVectors(c.Request.Context(), &req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to upsert vectors", err.Error())
			return
		}

		response.Success(c, result)
	}
}

func searchVectors(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		var req models.SearchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}

		// Set collection name from URL parameter
		req.CollectionName = collectionName

		// Search vectors in vector database
		result, err := vectorDAO.QueryVectors(c.Request.Context(), &req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to search vectors", err.Error())
			return
		}

		response.Success(c, result)
	}
}

func getVector(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")
		vectorID := c.Param("id")
		_ = c.DefaultQuery("include_vector", "false") == "true" // includeVector parameter is currently not used

		// Get vector from vector database
		vector, err := vectorDAO.GetVector(c.Request.Context(), collectionName, vectorID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to get vector", err.Error())
			return
		}

		response.Success(c, vector)
	}
}

func deleteVector(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")
		vectorID := c.Param("id")

		// Delete vector from vector database
		deleteReq := &models.DeleteVectorsRequest{
			CollectionName: collectionName,
			Ids:            []string{vectorID},
		}
		_, err := vectorDAO.BatchDeleteVectorData(c.Request.Context(), collectionName, deleteReq)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to delete vector", err.Error())
			return
		}

		response.Success(c, gin.H{
			"message": "Vector deleted successfully",
			"collection_name": collectionName,
			"vector_id": vectorID,
		})
	}
}

func batchDeleteVectors(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}

		collectionName := c.Param("name")

		var req models.DeleteVectorsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request body")
			return
		}

		// Delete vectors from vector database
		deleteReq := &models.DeleteVectorsRequest{
			CollectionName: collectionName,
			Ids:            req.Ids,
		}
		_, err := vectorDAO.BatchDeleteVectorData(c.Request.Context(), collectionName, deleteReq)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to delete vectors", err.Error())
			return
		}

		response.Success(c, gin.H{
			"message": "Vectors deleted successfully",
			"collection_name": collectionName,
			"deleted_count": len(req.Ids),
		})
	}
}

// Vector database operation handlers

func healthCheck(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := vectorDAO.VectorHealthCheck(c.Request.Context())
		if err != nil {
			response.InternalServerError(c, "Health check failed")
			return
		}

		response.Success(c, gin.H{"status": "healthy"})
	}
}

func getVectorDatabaseInfoV1(vectorDAO *dao.VectorDAO) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get vector database info
		info, err := vectorDAO.GetVectorDatabaseInfo(c.Request.Context())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, 500, "Failed to get vector database info", err.Error())
			return
		}

		response.Success(c, info)
	}
}