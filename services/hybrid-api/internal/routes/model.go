package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/handlers"
)

// SetupModelRoutes 设置模型路由
func SetupModelRoutes(router *gin.Engine, modelHandler *handlers.ModelHandler) {
	// 模型配置路由
	configGroup := router.Group("/api/v1/models/configs")
	{
		configGroup.POST("", modelHandler.CreateModelConfig)
		configGroup.GET("/:id", modelHandler.GetModelConfig)
		configGroup.GET("", modelHandler.ListModelConfigs)
		configGroup.PUT("/:id", modelHandler.UpdateModelConfig)
		configGroup.DELETE("/:id", modelHandler.DeleteModelConfig)
	}

	// 模型服务路由
	serviceGroup := router.Group("/api/v1/models/services")
	{
		serviceGroup.GET("", modelHandler.ListServices)
		serviceGroup.GET("/health", modelHandler.HealthCheck)
		serviceGroup.POST("/:service/generate-text", modelHandler.GenerateText)
		serviceGroup.POST("/:service/generate-embeddings", modelHandler.GenerateEmbeddings)
		serviceGroup.POST("/generate-text", modelHandler.GenerateText)
		serviceGroup.POST("/generate-embeddings", modelHandler.GenerateEmbeddings)
	}

	// 对话路由
	conversationGroup := router.Group("/api/v1/models/conversations")
	{
		conversationGroup.POST("", modelHandler.CreateConversation)
		conversationGroup.GET("/:id", modelHandler.GetConversation)
		conversationGroup.GET("", modelHandler.ListConversations)
		conversationGroup.DELETE("/:id", modelHandler.DeleteConversation)
	}

	// 消息路由
	messageGroup := router.Group("/api/v1/models/messages")
	{
		messageGroup.POST("", modelHandler.CreateMessage)
		messageGroup.GET("/conversation/:conversation_id", modelHandler.GetMessages)
	}

	// 微调作业路由
	fineTuningGroup := router.Group("/api/v1/models/fine-tuning")
	{
		fineTuningGroup.POST("/jobs", modelHandler.CreateFineTuningJob)
		fineTuningGroup.GET("/jobs/:id", modelHandler.GetFineTuningJob)
		fineTuningGroup.GET("/jobs", modelHandler.ListFineTuningJobs)
	}
}