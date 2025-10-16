package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// SetupAppModuleRoutes 设置应用模块相关路由
func SetupAppModuleRoutes(router *gin.Engine, db *gorm.DB, logger *zap.Logger, jwtMiddleware *middleware.JWTMiddleware) {
	// 创建服务和处理器
	appModuleService := services.NewAppModuleService(db, logger)
	appModuleHandler := handlers.NewAppModuleHandler(appModuleService, logger)

	// API版本组
	v1 := router.Group("/api/v1")
	
	// 应用模块路由组
	moduleGroup := v1.Group("/modules")
	moduleGroup.Use(jwtMiddleware.AuthRequired()) // 需认证
	{
		// 获取用户模块
		moduleGroup.GET("/user", appModuleHandler.GetUserModules)
		
		// 获取模块详情
		moduleGroup.GET("/:id", appModuleHandler.GetModule)
		
		// 设置用户模块权限
		moduleGroup.POST("/permission", appModuleHandler.SetUserModulePermission)
		
		// 管理员模块路由组
		adminGroup := moduleGroup.Group("/admin")
		{
			// 创建模块
			adminGroup.POST("", appModuleHandler.CreateModule)
			
			// 更新模块
			adminGroup.PUT("/:id", appModuleHandler.UpdateModule)
			
			// 删除模块	
			adminGroup.DELETE("/:id", appModuleHandler.DeleteModule)
			
			// 初始化默认模块	- 需管理员权限
			adminGroup.POST("/initialize", appModuleHandler.InitializeDefaultModules)
		}
	}

	// 用户偏好路由组
	preferenceGroup := v1.Group("/preferences")
	preferenceGroup.Use(jwtMiddleware.AuthRequired()) // 需认证
	{
		// 获取用户偏好
		preferenceGroup.GET("", appModuleHandler.GetUserPreference)
		
		// 更新用户偏好
		preferenceGroup.PUT("", appModuleHandler.UpdateUserPreference)
	}

	logger.Info("App module routes registered successfully")
}

