package routes

import (
	"github.com/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/taishanglaojun/core-services/ai-integration/middleware"
	"github.com/gin-gonic/gin"
)

// SetupMultimodalRoutes 设置多模态AI路由
func SetupMultimodalRoutes(router *gin.Engine, handler *handlers.MultimodalHandler) {
	// API版本分组
	v1 := router.Group("/api/v1")
	{
		// 多模态AI路由组
		multimodal := v1.Group("/multimodal")
		{
			// 应用认证中间件
			multimodal.Use(middleware.AuthMiddleware())
			
			// 应用限流中间件
			multimodal.Use(middleware.RateLimitMiddleware())
			
			// 核心处理端点
			multimodal.POST("/process", handler.ProcessMultimodal)
			multimodal.POST("/upload", handler.UploadFile)
			multimodal.GET("/stream", handler.StreamMultimodal)
			
			// 会话管理端点
			sessions := multimodal.Group("/sessions")
			{
				sessions.GET("", handler.GetSessions)
				sessions.POST("", handler.CreateSession)
				sessions.GET("/:id", handler.GetSession)
				sessions.PUT("/:id", handler.UpdateSession)
				sessions.DELETE("/:id", handler.DeleteSession)
				sessions.GET("/:id/messages", handler.GetSessionMessages)
			}
			
			// 图像处理端点
			image := multimodal.Group("/image")
			{
				// 图像生成
				image.POST("/generate", handler.GenerateImage)
				
				// 图像分析
				image.POST("/analyze", handler.AnalyzeImage)
				image.POST("/upload-analyze", handler.UploadImageForAnalysis)
				
				// 图像编辑
				image.POST("/edit", handler.EditImage)
			}
		}
	}
	
	// 健康检查端点（无需认证）
	router.GET("/health/multimodal", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "multimodal-ai",
			"version": "1.0.0",
		})
	})
}