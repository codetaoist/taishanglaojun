package routes

import (
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/handler"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/handlers"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/jwt"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/middleware"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/service"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes 设置路由
func SetupRoutes(
	router *gin.Engine,
	authHandler *handler.AuthHandler,
	databaseHandler *handlers.DatabaseHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimiter *middleware.RateLimiter,
	db *gorm.DB,
	logger *zap.Logger,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	tokenRepo repository.TokenRepository,
	jwtManager *jwt.Manager,
	authService service.AuthService,
) {
	// 全局中间件
	router.Use(gin.Recovery())
	router.Use(authMiddleware.RequestLogger())
	// CORS由API网关统一处理，这里不再设置
	router.Use(authMiddleware.SecurityHeaders())
	router.Use(rateLimiter.Middleware())

	// 静态文件服务
	router.Static("/uploads", "./uploads")

	// API版本分组
	v1 := router.Group("/api/v1")
	{
		// 健康检查路由
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "auth-system",
				"version": "1.0.0",
			})
		})

		// 公开路由（不需要认证）
		public := v1.Group("/auth")
		{
			public.POST("/register", authHandler.Register)
			public.POST("/login", authHandler.Login)
			public.POST("/refresh", authHandler.RefreshToken)
			public.POST("/forgot-password", authHandler.ForgotPassword)
			public.POST("/reset-password", authHandler.ResetPassword)
			public.POST("/verify-email", authHandler.VerifyEmail)
			public.POST("/resend-verification", authHandler.ResendVerification)
			public.POST("/validate-token", authHandler.ValidateToken)
		}

		// 需要认证的路由
		protected := v1.Group("/")
		protected.Use(authMiddleware.RequireAuth())
		{
			// 用户相关路由
			user := protected.Group("/user")
			{
				user.GET("/me", authHandler.Me)
				user.GET("/profile", authHandler.GetProfile)
				user.PUT("/profile", authHandler.UpdateProfile)
				user.POST("/change-password", authHandler.ChangePassword)
				user.POST("/logout", authHandler.Logout)
			}

			// 文件上传路由
			uploadHandler := handlers.NewUploadHandler(logger, "./uploads", "http://localhost:8082")
			upload := protected.Group("/upload")
			{
				upload.POST("/avatar", uploadHandler.UploadAvatar)
				upload.POST("/file", uploadHandler.UploadFile)
			}

			// 会话管理器路由
			session := protected.Group("/sessions")
			{
				session.GET("/", authHandler.GetSessions)
				session.DELETE("/:sessionId", authHandler.RevokeSession)
				session.DELETE("/", authHandler.RevokeAllSessions)
			}
		}

		// 管理员路由
		admin := v1.Group("/admin")
		admin.Use(authMiddleware.RequireAuth())
		admin.Use(authMiddleware.RequireRole(models.RoleAdmin))
		{
			// 创建用户管理处理器
			userHandler := handlers.NewUserHandler(userRepo, authService, logger)
			
			// 用户管理器
			users := admin.Group("/users")
			{
				users.GET("", userHandler.ListUsers)
				users.POST("", userHandler.CreateUser)
				users.GET("/stats", userHandler.GetUserStats)
				users.POST("/batch-delete", userHandler.BatchDeleteUsers)
				users.GET("/:userId", userHandler.GetUser)
				users.PUT("/:userId", userHandler.UpdateUser)
				users.DELETE("/:userId", userHandler.DeleteUser)
				users.PUT("/:userId/status", userHandler.UpdateUserStatus)
				users.PUT("/:userId/role", userHandler.UpdateUserRole)
			}

			// 会话管理器
			sessions := admin.Group("/sessions")
			{
				sessions.GET("/", func(c *gin.Context) {
					// TODO: 实现会话列表
					c.JSON(200, gin.H{"message": "Session list endpoint"})
				})
				sessions.DELETE("/:sessionId", func(c *gin.Context) {
					// TODO: 实现撤销会话
					c.JSON(200, gin.H{"message": "Revoke session endpoint"})
				})
			}

			// 系统计统计
			stats := admin.Group("/stats")
			{
				stats.GET("/users", func(c *gin.Context) {
					// TODO: 实现用户统计
					c.JSON(200, gin.H{"message": "User stats endpoint"})
				})
				stats.GET("/sessions", func(c *gin.Context) {
					// TODO: 实现会话统计
					c.JSON(200, gin.H{"message": "Session stats endpoint"})
				})
			}
		}

		// 超级管理员路由
		superAdmin := v1.Group("/super-admin")
		superAdmin.Use(authMiddleware.RequireAuth())
		superAdmin.Use(authMiddleware.SuperAdminOnly())
		{
			// 权限管理器
			permissions := superAdmin.Group("/permissions")
			{
				permissions.GET("/", func(c *gin.Context) {
					// TODO: 实现权限列表
					c.JSON(200, gin.H{"message": "Permission list endpoint"})
				})
				permissions.POST("/", func(c *gin.Context) {
					// TODO: 实现创建权限
					c.JSON(200, gin.H{"message": "Create permission endpoint"})
				})
				permissions.PUT("/:permissionId", func(c *gin.Context) {
					// TODO: 实现更新权限
					c.JSON(200, gin.H{"message": "Update permission endpoint"})
				})
				permissions.DELETE("/:permissionId", func(c *gin.Context) {
					// TODO: 实现删除权限
					c.JSON(200, gin.H{"message": "Delete permission endpoint"})
				})
			}

			// 角色权限管理器
			rolePermissions := superAdmin.Group("/role-permissions")
			{
				rolePermissions.GET("/:role", func(c *gin.Context) {
					// TODO: 实现获取角色权限
					c.JSON(200, gin.H{"message": "Role permissions endpoint"})
				})
				rolePermissions.POST("/:role", func(c *gin.Context) {
					// TODO: 实现分配角色权限
					c.JSON(200, gin.H{"message": "Assign role permission endpoint"})
				})
				rolePermissions.DELETE("/:role/:permissionId", func(c *gin.Context) {
					// TODO: 实现撤销角色权限
					c.JSON(200, gin.H{"message": "Revoke role permission endpoint"})
				})
			}

			// 用户权限管理器
			userPermissions := superAdmin.Group("/user-permissions")
			{
				userPermissions.GET("/:userId", func(c *gin.Context) {
					// TODO: 实现获取用户权限
					c.JSON(200, gin.H{"message": "User permissions endpoint"})
				})
				userPermissions.POST("/:userId", func(c *gin.Context) {
					// TODO: 实现分配用户权限
					c.JSON(200, gin.H{"message": "Assign user permission endpoint"})
				})
				userPermissions.DELETE("/:userId/:permissionId", func(c *gin.Context) {
					// TODO: 实现撤销用户权限
					c.JSON(200, gin.H{"message": "Revoke user permission endpoint"})
				})
			}

			// 系统计管理器
			system := superAdmin.Group("/system")
			{
				system.GET("/info", func(c *gin.Context) {
					// TODO: 实现系统计信息
					c.JSON(200, gin.H{"message": "System info endpoint"})
				})
				system.POST("/cleanup", func(c *gin.Context) {
					// TODO: 实现系统计清理
					c.JSON(200, gin.H{"message": "System cleanup endpoint"})
				})
			}

			// 动态数量据库管理器
			database := superAdmin.Group("/database")
			{
				database.GET("/list", databaseHandler.ListDatabases)
				database.GET("/current", databaseHandler.GetCurrentDatabase)
				database.POST("/add", databaseHandler.AddDatabase)
				database.POST("/switch/:name", databaseHandler.SwitchDatabase)
				database.DELETE("/:name", databaseHandler.RemoveDatabase)
				database.GET("/health", databaseHandler.HealthCheck)
				database.GET("/stats", databaseHandler.GetDatabaseStats)
			}
		}
	}

	// 开发环境路由
	if gin.Mode() == gin.DebugMode {
		debug := router.Group("/debug")
		{
			debug.GET("/routes", func(c *gin.Context) {
				routes := router.Routes()
				c.JSON(200, gin.H{
					"total":  len(routes),
					"routes": routes,
				})
			})
		}
	}

	logger.Info("Routes setup completed")
}
