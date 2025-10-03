package location_tracking

import (
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/models"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/routes"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SetupRoutes 设置位置跟踪模块路由
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, jwtMiddleware *middleware.JWTMiddleware) {
	// 自动迁移数据库表
	if err := db.AutoMigrate(&models.Trajectory{}, &models.LocationPoint{}); err != nil {
		logger.Error("Failed to migrate location tracking tables", zap.Error(err))
		return
	}

	// 创建服务实例
	locationService := services.NewLocationService(db, logger)

	// 创建处理器实例
	locationHandler := handlers.NewLocationHandler(locationService, logger)

	// 创建位置跟踪路由组，直接使用传入的apiV1路由组
	locationGroup := router.Group("/location-tracking")
	locationGroup.Use(jwtMiddleware.AuthRequired())

	// 设置路由
	routes.SetupLocationRoutes(locationGroup, locationHandler)

	logger.Info("Location tracking routes setup completed")
}