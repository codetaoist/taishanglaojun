package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// SetupAdminRoutes 设置管理员相关路由
func SetupAdminRoutes(apiV1 *gin.RouterGroup, authService *middleware.AuthService, jwtMiddleware *middleware.JWTMiddleware, db *gorm.DB, logger *zap.Logger) {
	// 创建管理员处理器
	adminHandler := handlers.NewAdminHandler(authService, logger, db)
	
	// 创建数据库连接管理服务和处理器
	// 这里使用一个默认的加密密钥，在生产环境中应该从配置文件或环境变量中获取
	encryptKey := "your-32-byte-encryption-key-here!!" // 32字节密钥
	dbConnectionService := services.NewDatabaseConnectionService(db, encryptKey)
	dbConnectionHandler := handlers.NewDatabaseConnectionHandler(dbConnectionService)
	
	// 管理员路由组
	adminGroup := apiV1.Group("/admin")
	
	// 添加测试路由（不需要认证）
	adminGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Admin routes working"})
	})
	
	adminGroup.Use(jwtMiddleware.AuthRequired()) // 需要认证
	{
		// 用户管理路由
		userGroup := adminGroup.Group("/users")
		{
			// 获取用户列表
			userGroup.GET("", adminHandler.GetUsers)
			
			// 获取用户统计
			userGroup.GET("/stats", adminHandler.GetUserStats)
			
			// 创建用户
			userGroup.POST("", adminHandler.CreateUser)
			
			// 更新用户
			userGroup.PUT("/:id", adminHandler.UpdateUser)
			
			// 删除用户
			userGroup.DELETE("/:id", adminHandler.DeleteUser)
			
			// 批量删除用户
			userGroup.POST("/batch-delete", adminHandler.BatchDeleteUsers)
			
			// 更新用户状态
			userGroup.PUT("/:id/status", adminHandler.UpdateUserStatus)
		}
		
		// 数据库连接管理路由
		databaseGroup := adminGroup.Group("/database")
		{
			// 获取支持的数据库类型
			databaseGroup.GET("/types", dbConnectionHandler.GetDatabaseTypes)
			
			// 连接管理路由
			connectionGroup := databaseGroup.Group("/connections")
			{
				// 获取连接列表
				connectionGroup.GET("", dbConnectionHandler.GetConnections)
				
				// 创建连接
				connectionGroup.POST("", dbConnectionHandler.CreateConnection)
				
				// 测试连接（不保存）
				connectionGroup.POST("/test", dbConnectionHandler.TestConnection)
				
				// 获取连接统计
				connectionGroup.GET("/stats", dbConnectionHandler.GetConnectionStats)
				
				// 获取所有连接状态
				connectionGroup.GET("/status", dbConnectionHandler.GetConnectionsStatus)
				
				// 单个连接操作
				connectionGroup.GET("/:id", dbConnectionHandler.GetConnection)
				connectionGroup.PUT("/:id", dbConnectionHandler.UpdateConnection)
				connectionGroup.DELETE("/:id", dbConnectionHandler.DeleteConnection)
				
				// 测试已保存的连接
				connectionGroup.POST("/:id/test", dbConnectionHandler.TestSavedConnection)
				
				// 刷新连接状态
				connectionGroup.POST("/:id/status", dbConnectionHandler.RefreshConnectionStatus)
			}
		}
	}

	logger.Info("Admin routes registered successfully")
}