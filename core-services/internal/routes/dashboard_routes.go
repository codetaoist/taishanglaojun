package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// SetupDashboardRoutes 设置仪表板相关路由
func SetupDashboardRoutes(apiV1 *gin.RouterGroup, authService *middleware.AuthService, jwtMiddleware *middleware.JWTMiddleware, userService *services.UserService, menuService *services.MenuService, db *gorm.DB, logger *zap.Logger) {
	// 创建仪表板处理器
	dashboardHandler := handlers.NewDashboardHandler(authService, userService, menuService, db, logger)
	
	// 仪表板路由组
	dashboardGroup := apiV1.Group("/dashboards")
	
	// 添加测试路由（不需要认证）
	dashboardGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Dashboard routes working"})
	})
	
	dashboardGroup.Use(jwtMiddleware.AuthRequired()) // 需要认证
	{
		// 获取仪表板统计数据
		dashboardGroup.GET("/stats", dashboardHandler.GetDashboardStats)
		
		// 获取用户统计数据
		dashboardGroup.GET("/user-stats", dashboardHandler.GetUserStats)
		
		// 获取系统指标
		dashboardGroup.GET("/metrics", dashboardHandler.GetSystemMetrics)
		
		// 获取最近活动
		dashboardGroup.GET("/activities", dashboardHandler.GetRecentActivities)
		
		// 获取趋势数据
		dashboardGroup.GET("/trends", dashboardHandler.GetTrendData)
		
		// 获取快捷操作
		dashboardGroup.GET("/quick-actions", dashboardHandler.GetQuickActions)
	}

	logger.Info("Dashboard routes registered successfully")
}