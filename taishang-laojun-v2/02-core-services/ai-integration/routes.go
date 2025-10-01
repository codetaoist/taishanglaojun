package ai_integration

import (
	"github.com/gin-gonic/gin"
	"github.com/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/taishanglaojun/core-services/ai-integration/providers"
	"github.com/taishanglaojun/core-services/ai-integration/services"
	"gorm.io/gorm"
)

// SetupRoutes 设置AI集成服务路由
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) {
	// 创建服务
	chatService := services.NewChatService(db, logger, providerManager)
	
	// 创建处理器
	chatHandler := handlers.NewChatHandler(chatService, logger)
	
	// AI相关路由组
	aiGroup := router.Group("/ai")
	{
		// 对话相关路由
		aiGroup.POST("/chat", chatHandler.Chat)
		aiGroup.GET("/sessions", chatHandler.GetSessions)
		aiGroup.GET("/sessions/:session_id/messages", chatHandler.GetMessages)
		aiGroup.DELETE("/sessions/:session_id", chatHandler.DeleteSession)
		
		// 提供商信息路由
		aiGroup.GET("/providers", getProviders(providerManager))
		aiGroup.GET("/models", getModels(providerManager))
		aiGroup.GET("/health", healthCheck(providerManager))
	}
}

// getProviders 获取可用的AI提供商列表
func getProviders(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		providers := manager.ListProviders()
		c.JSON(200, gin.H{
			"code": "SUCCESS",
			"data": gin.H{
				"providers": providers,
			},
		})
	}
}

// getModels 获取所有提供商的模型列表
func getModels(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		models := manager.GetAllModels()
		c.JSON(200, gin.H{
			"code": "SUCCESS",
			"data": gin.H{
				"models": models,
			},
		})
	}
}

// healthCheck AI提供商健康检查
func healthCheck(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		results := manager.HealthCheck(c.Request.Context())
		
		// 检查是否有任何提供商可用
		hasHealthy := false
		for _, err := range results {
			if err == nil {
				hasHealthy = true
				break
			}
		}
		
		status := 200
		if !hasHealthy {
			status = 503 // Service Unavailable
		}
		
		// 转换错误为字符串
		healthStatus := make(map[string]string)
		for provider, err := range results {
			if err != nil {
				healthStatus[provider] = err.Error()
			} else {
				healthStatus[provider] = "healthy"
			}
		}
		
		c.JSON(status, gin.H{
			"code": "SUCCESS",
			"data": gin.H{
				"status":    healthStatus,
				"healthy":   hasHealthy,
				"timestamp": c.Request.Context().Value("request_time"),
			},
		})
	}
}