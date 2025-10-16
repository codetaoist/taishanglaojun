package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// SetupAPIDocumentationRoutes 设置API文档管理路由
func SetupAPIDocumentationRoutes(
	router *gin.Engine,
	apiDocHandler *handlers.APIDocumentationHandler,
	jwtMiddleware *middleware.JWTMiddleware,
	logger *zap.Logger,
) {
	logger.Info("Setting up API documentation routes")

	// API文档管理路由组
	apiDocGroup := router.Group("/api/v1/api-docs")
	
	// 公开路由（不需要认证）
	{
		// 获取分类列表
		apiDocGroup.GET("/categories", apiDocHandler.GetCategories)
		
		// 根据ID获取分类
		apiDocGroup.GET("/categories/:id", apiDocHandler.GetCategoryByID)
		
		// 获取接口列表
		apiDocGroup.GET("/endpoints", apiDocHandler.GetEndpoints)
		
		// 根据ID获取接口详情
		apiDocGroup.GET("/endpoints/:id", apiDocHandler.GetEndpointByID)
		
		// 根据分类获取接口列表
		apiDocGroup.GET("/categories/:id/endpoints", apiDocHandler.GetEndpointsByCategory)
		
		// 搜索接口
		apiDocGroup.GET("/search", apiDocHandler.SearchEndpoints)
		
		// 获取统计信息
		apiDocGroup.GET("/statistics", apiDocHandler.GetStatistics)
	}

	// 需要认证的路由
	authGroup := apiDocGroup.Group("")
	authGroup.Use(jwtMiddleware.AuthRequired())
	{
		// 测试API接口
		authGroup.POST("/test", apiDocHandler.TestAPI)
		
		// 获取API测试历史
		authGroup.GET("/endpoints/:id/test-history", apiDocHandler.GetAPITestHistory)
	}

	logger.Info("API documentation routes setup completed")
}