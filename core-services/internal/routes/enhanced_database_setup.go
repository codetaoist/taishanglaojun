package routes

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	
	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// SetupEnhancedDatabaseRoutes 设置增强数据库监控路由
func SetupEnhancedDatabaseRoutes(router *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware, db *gorm.DB, logger *zap.Logger) {
	// 创建增强数据库处理器
	enhancedDBHandler := handlers.NewEnhancedDatabaseHandler(db, logger)
	
	// 创建缓存管理器
	cacheManager := middleware.NewCacheManager(logger)
	
	// 增强数据库监控路由组
	enhancedDB := router.Group("/enhanced-database")
	enhancedDB.Use(jwtMiddleware.AuthRequired())
	{
		// 实时指标（短缓存）
		metricsGroup := enhancedDB.Group("/metrics")
		metricsGroup.Use(cacheManager.CacheMiddleware(middleware.CacheConfig{
			TTL:       10 * time.Second,
			KeyPrefix: "db_metrics",
		}))
		{
			metricsGroup.GET("", enhancedDBHandler.GetDatabaseMetrics)
		}
		
		// 历史指标（中等缓存）
		enhancedDB.GET("/metrics/history", 
			cacheManager.CacheMiddleware(middleware.CacheConfig{
				TTL:       2 * time.Minute,
				KeyPrefix: "db_historical",
			}),
			enhancedDBHandler.GetDatabaseMetricsHistory)
		
		// 健康检查（短缓存）
		enhancedDB.GET("/health", 
			cacheManager.CacheMiddleware(middleware.CacheConfig{
				TTL:       5 * time.Second,
				KeyPrefix: "db_health",
			}),
			enhancedDBHandler.GetDatabaseHealth)
		
		// 连接监控
		connectionGroup := enhancedDB.Group("/connections")
		{
			// 活跃连接（短缓存）
			connectionGroup.GET("/active", 
				cacheManager.CacheMiddleware(middleware.CacheConfig{
					TTL:       3 * time.Second,
					KeyPrefix: "db_connections",
				}),
				enhancedDBHandler.GetActiveConnections)
			
			// 连接池统计（中等缓存）
			connectionGroup.GET("/pool-stats", 
				cacheManager.CacheMiddleware(middleware.CacheConfig{
					TTL:       30 * time.Second,
					KeyPrefix: "db_pool_stats",
				}),
				enhancedDBHandler.GetConnectionPoolStats)
			
			// 操作类接口不缓存
			connectionGroup.POST("/test-all", enhancedDBHandler.TestAllConnections)
			connectionGroup.DELETE("/kill/:id", enhancedDBHandler.KillConnection)
			connectionGroup.POST("/kill-multiple", enhancedDBHandler.KillMultipleConnections)
		}
		
		// 数据库优化（不缓存）
		enhancedDB.POST("/optimize", enhancedDBHandler.OptimizeDatabase)
		
		// 备份管理
		enhancedDB.POST("/backups", enhancedDBHandler.CreateBackup)
		enhancedDB.GET("/backups", enhancedDBHandler.GetBackups)
		enhancedDB.POST("/backups/:id/restore", enhancedDBHandler.RestoreBackup)
		enhancedDB.DELETE("/backups/:id", enhancedDBHandler.DeleteBackup)
		enhancedDB.GET("/backups/:id/status", enhancedDBHandler.GetBackupStatus)
		enhancedDB.GET("/backups/:id/progress", enhancedDBHandler.GetBackupProgress)
		
		// WebSocket 实时监控
		enhancedDB.GET("/ws/metrics", enhancedDBHandler.MetricsWebSocket)
		enhancedDB.GET("/ws/health", enhancedDBHandler.HealthWebSocket)
	}
	
	logger.Info("Enhanced database monitoring routes setup completed")
}