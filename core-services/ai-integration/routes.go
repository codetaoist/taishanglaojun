package ai_integration

import (
	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutes 设置AI集成服务路由
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) {
	// 创建服务
	chatService := services.NewChatService(db, logger, providerManager)
	
	// 创建处理器
	chatHandler := handlers.NewChatHandler(chatService, logger)

	// 创建JWT中间件（从配置中获取）
	jwtConfig := middleware.JWTConfig{
		Secret:     "taishang-secret-key", // 应该从配置中获取
		Issuer:     "taishang-core-services",
		Expiration: 24 * 60 * 60 * 1000000000, // 24小时，单位纳秒
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig, logger)
	
	// AI相关路由组
	aiGroup := router.Group("/ai")
	{
		// 需要认证的对话相关路由
		aiGroup.POST("/chat", jwtMiddleware.AuthRequired(), chatHandler.Chat)
		aiGroup.GET("/sessions", jwtMiddleware.AuthRequired(), chatHandler.GetSessions)
		aiGroup.GET("/sessions/:session_id/messages", jwtMiddleware.AuthRequired(), chatHandler.GetMessages)
		aiGroup.DELETE("/sessions/:session_id", jwtMiddleware.AuthRequired(), chatHandler.DeleteSession)
		
		// 公开的提供商信息路由（不需要认证）
		aiGroup.GET("/providers", getProviders(providerManager))
		aiGroup.GET("/models", getModels(providerManager))
		aiGroup.GET("/health", healthCheck(providerManager))
	}
}

// getProviders 获取可用的AI提供商列表
func getProviders(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := manager.GetProviderInfo()
		c.JSON(200, gin.H{
			"code": "SUCCESS",
			"data": info,
		})
	}
}

// getModels 获取所有提供商的模型列表
func getModels(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := manager.GetProviderInfo()
		c.JSON(200, gin.H{
			"code": "SUCCESS",
			"data": gin.H{
				"message": "模型信息需要通过具体提供商获取",
				"providers": info["providers"],
			},
		})
	}
}

// healthCheck AI提供商健康检查
func healthCheck(manager *providers.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		info := manager.GetProviderInfo()
		providers, ok := info["providers"].([]string)
		if !ok {
			providers = []string{}
		}
		
		// 检查每个提供商的健康状态
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

