package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// SetupSystemRoutes 设置系统设置相关路由
func SetupSystemRoutes(apiV1 *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware, db *gorm.DB, logger *zap.Logger) {
	// 创建系统设置处理器
	systemHandler := handlers.NewSystemHandler(db, logger)

	// 系统设置路由组
	systemGroup := apiV1.Group("/system")
	
	// 添加测试路由（不需要认证）
	systemGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "System routes working"})
	})
	
	systemGroup.Use(jwtMiddleware.AuthRequired()) // 需要认证
	{
		// 系统配置管理
		systemGroup.GET("/config", systemHandler.GetSystemConfig)
		systemGroup.PUT("/config", systemHandler.UpdateSystemConfig)
		
		// 系统信息
		systemGroup.GET("/info", systemHandler.GetSystemInfo)
		
		// 系统状态
		systemGroup.GET("/status", systemHandler.GetSystemStatus)
		
		// 系统日志
		systemGroup.GET("/logs", systemHandler.GetSystemLogs)
		systemGroup.GET("/logs/stats", systemHandler.GetSystemLogStats)
		// 前端/客户端日志写入
		systemGroup.POST("/logs", systemHandler.CreateSystemLog)
		
		// 系统备份
		systemGroup.POST("/backup", systemHandler.CreateBackup)
		systemGroup.GET("/backups", systemHandler.GetBackups)
		systemGroup.POST("/restore/:id", systemHandler.RestoreBackup)
		
		// 系统维护
		systemGroup.POST("/maintenance/start", systemHandler.StartMaintenance)
		systemGroup.POST("/maintenance/stop", systemHandler.StopMaintenance)
		systemGroup.GET("/maintenance/status", systemHandler.GetMaintenanceStatus)
		
		// 缓存管理
		systemGroup.DELETE("/cache", systemHandler.ClearCache)
		systemGroup.GET("/cache/stats", systemHandler.GetCacheStats)
		
		// 数据库管理
		systemGroup.GET("/database/stats", systemHandler.GetDatabaseStats)
		systemGroup.POST("/database/optimize", systemHandler.OptimizeDatabase)
		systemGroup.GET("/database/schemas", systemHandler.ListSchemas)
		systemGroup.GET("/database/tables", systemHandler.ListDBTables)
		systemGroup.GET("/database/tables/:name/columns", systemHandler.GetTableColumns)
		systemGroup.POST("/database/query", systemHandler.RunReadOnlyQuery)

		// 问题检测与告警
		systemGroup.GET("/issues/detect", systemHandler.DetectIssues)
		systemGroup.POST("/issues/alert", systemHandler.TriggerIssueAlert)
	}

	logger.Info("System routes registered successfully")
}