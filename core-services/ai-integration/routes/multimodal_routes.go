package routes

import (
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/gin-gonic/gin"
)

// SetupMultimodalRoutes AI
func SetupMultimodalRoutes(router *gin.Engine, handler *handlers.MultimodalHandler) {
	// API汾
	v1 := router.Group("/api/v1")
	{
		// AI
		multimodal := v1.Group("/multimodal")
		{
			// 
			multimodal.POST("/process", handler.ProcessMultimodal)
			multimodal.POST("/upload", handler.UploadFile)
			multimodal.GET("/stream", handler.StreamMultimodal)

			// 
			sessions := multimodal.Group("/sessions")
			{
				sessions.GET("", handler.GetSessions)
				sessions.POST("", handler.CreateSession)
				sessions.GET("/:id", handler.GetSession)
				sessions.PUT("/:id", handler.UpdateSession)
				sessions.DELETE("/:id", handler.DeleteSession)
				sessions.GET("/:id/messages", handler.GetSessionMessages)
			}

			// 
			image := multimodal.Group("/image")
			{
				// 
				image.POST("/generate", handler.GenerateImage)

				// 
				image.POST("/analyze", handler.AnalyzeImage)
				image.POST("/upload-analyze", handler.UploadImageForAnalysis)

				// 
				image.POST("/edit", handler.EditImage)
			}
		}
	}

	// 
	router.GET("/health/multimodal", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "multimodal-ai",
			"version": "1.0.0",
		})
	})
}

