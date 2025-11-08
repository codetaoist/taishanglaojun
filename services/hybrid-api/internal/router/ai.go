package router

import (
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/middleware"
	grpcClient "github.com/codetaoist/taishang/api/internal/grpc"
	"github.com/codetaoist/taishang/api/internal/handler"
	"github.com/codetaoist/taishang/api/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupAIRoutes sets up the AI service routes
func SetupAIRoutes(cfg *config.Config, r *gin.Engine) {
	// Initialize gRPC client
	aiClient, err := grpcClient.NewAIServiceClient(cfg.AIService.VectorAddr, cfg.AIService.ModelAddr)
	if err != nil {
		panic("Failed to initialize AI service client: " + err.Error())
	}

	// Initialize services
	aiService := service.NewAIService(aiClient)

	// Initialize handlers
	aiHandler := handler.NewAIHandler(aiService)

	// Apply authentication middleware
	authMiddleware := middleware.Auth(*cfg)

	ai := r.Group("/api/v1/ai")
	ai.Use(authMiddleware)

	// Health check
	{
		ai.GET("/health", aiHandler.Health)
	}

	// Vector database operations
	{
		ai.POST("/collections", aiHandler.CreateCollection)
		ai.DELETE("/collections/:name", aiHandler.DropCollection)
		ai.GET("/collections", aiHandler.ListCollections)
		ai.POST("/search", aiHandler.Search)
		ai.POST("/insert", aiHandler.Insert)
	}

	// Model operations
	{
		ai.GET("/models", aiHandler.ListModels)
		ai.POST("/models/:name/load", aiHandler.LoadModel)
		ai.GET("/models/:name/status", aiHandler.GetModelStatus)
	}

	// Inference operations
	{
		ai.POST("/generate/text", aiHandler.GenerateText)
		ai.POST("/generate/embedding", aiHandler.GenerateEmbedding)
	}
}