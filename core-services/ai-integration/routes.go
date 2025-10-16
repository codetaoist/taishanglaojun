package ai_integration

import (
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutes AI
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) {
	// 
	chatService := services.NewChatService(db, logger, providerManager)
	aiService := services.NewAIService(providerManager)
	crossModalService := services.NewCrossModalService(providerManager, logger)

	// 
	chatHandler := handlers.NewChatHandler(chatService, logger)
	aiHandler := handlers.NewAIHandler(aiService, logger)
	crossModalHandler := handlers.NewCrossModalHandler(crossModalService)

	// JWT
	jwtConfig := middleware.JWTConfig{
		Secret:     "taishang-secret-key", // 
		Issuer:     "taishang-core-services",
		Expiration: 24 * 60 * 60 * 1000000000, // 24
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig, logger)

	// AI
	aiGroup := router.Group("/ai")
	{
		// 
		aiGroup.POST("/chat", jwtMiddleware.AuthRequired(), chatHandler.Chat)
		aiGroup.GET("/sessions", jwtMiddleware.AuthRequired(), chatHandler.GetSessions)
		aiGroup.GET("/sessions/:session_id/messages", jwtMiddleware.AuthRequired(), chatHandler.GetMessages)
		aiGroup.DELETE("/sessions/:session_id", jwtMiddleware.AuthRequired(), chatHandler.DeleteSession)

		// AI
		aiGroup.POST("/intent", jwtMiddleware.AuthRequired(), aiHandler.IntentRecognition)
		aiGroup.POST("/sentiment", jwtMiddleware.AuthRequired(), aiHandler.SentimentAnalysis)

		// 
		aiGroup.GET("/providers", getProviders(providerManager))
		aiGroup.GET("/models", getModels(providerManager))
		aiGroup.GET("/health", healthCheck(providerManager))
	}

	// 
	crossModalGroup := router.Group("/crossmodal")
	{
		// 
		crossModalGroup.POST("/inference", jwtMiddleware.AuthRequired(), crossModalHandler.ProcessCrossModalInference)
		crossModalGroup.POST("/search", jwtMiddleware.AuthRequired(), crossModalHandler.SemanticSearch)
		crossModalGroup.POST("/match", jwtMiddleware.AuthRequired(), crossModalHandler.ContentMatching)
		crossModalGroup.POST("/qa", jwtMiddleware.AuthRequired(), crossModalHandler.MultiModalQA)
		crossModalGroup.POST("/scene", jwtMiddleware.AuthRequired(), crossModalHandler.SceneUnderstanding)
		crossModalGroup.POST("/emotion", jwtMiddleware.AuthRequired(), crossModalHandler.EmotionAnalysis)
		crossModalGroup.GET("/stream", jwtMiddleware.AuthRequired(), crossModalHandler.StreamCrossModalInference)
		crossModalGroup.GET("/history", jwtMiddleware.AuthRequired(), crossModalHandler.GetInferenceHistory)
		crossModalGroup.GET("/stats", jwtMiddleware.AuthRequired(), crossModalHandler.GetInferenceStats)
	}
}

// getProviders AI
func getProviders(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := manager.GetProviderInfo()
		c.JSON(200, gin.H{
			"code": "SUCCESS",
			"data": info,
		})
	}
}

// getModels 
func getModels(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := manager.GetProviderInfo()
		c.JSON(200, gin.H{
			"code": "SUCCESS",
			"data": gin.H{
				"message":   "",
				"providers": info["providers"],
			},
		})
	}
}

// healthCheck AI
func healthCheck(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := manager.GetProviderInfo()
		providers, ok := info["providers"].([]string)
		if !ok {
			providers = []string{}
		}

		// 
		healthStatus := make(map[string]string)
		hasHealthy := false

		for _, providerName := range providers {
			if manager.IsHealthy(providerName) {
				healthStatus[providerName] = "healthy"
				hasHealthy = true
			} else {
				healthStatus[providerName] = "unhealthy"
			}
		}

		status := 200
		if !hasHealthy {
			status = 503 // Service Unavailable
		}

		c.JSON(status, gin.H{
			"code": "SUCCESS",
			"data": gin.H{
				"status":    healthStatus,
				"healthy":   hasHealthy,
				"providers": len(providers),
			},
		})
	}
}

