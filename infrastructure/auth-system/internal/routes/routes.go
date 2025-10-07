package routes

import (
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/handler"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/handlers"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/jwt"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/middleware"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/monitoring"
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
	metrics *monitoring.Metrics,
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
		userHandler := handlers.NewUserHandler(userRepo, sessionRepo, authService, logger)
			
			// 创建会话管理处理器
			sessionHandler := handlers.NewSessionHandler(sessionRepo, authService, logger)
			
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
				sessions.GET("/", sessionHandler.ListSessions)
				sessions.DELETE("/:sessionId", sessionHandler.RevokeSession)
				sessions.POST("/cleanup", sessionHandler.CleanupExpiredSessions)
			}

			// 系统计统计
			stats := admin.Group("/stats")
			{
				stats.GET("/users", func(c *gin.Context) {
					// TODO: 实现用户统计
					c.JSON(200, gin.H{"message": "User stats endpoint"})
				})
				stats.GET("/sessions", sessionHandler.GetSessionStats)
			}
		}

		// 超级管理员路由
		superAdmin := v1.Group("/super-admin")
		superAdmin.Use(authMiddleware.RequireAuth())
		superAdmin.Use(authMiddleware.SuperAdminOnly())
		{
			// 创建权限管理相关的处理器
			permissionRepo := repository.NewPermissionRepository(db)
			rolePermissionRepo := repository.NewRolePermissionRepository(db)
			userPermissionRepo := repository.NewUserPermissionRepository(db)
			permissionService := service.NewPermissionService(db, permissionRepo, rolePermissionRepo, userPermissionRepo, userRepo, logger)
			permissionHandler := handlers.NewPermissionHandler(permissionService, logger)

			// 创建系统监控处理器
			systemHandler := handlers.NewSystemHandler(db, userRepo, sessionRepo, metrics, logger)
			// 权限管理器
			permissions := superAdmin.Group("/permissions")
			{
				permissions.GET("/", permissionHandler.ListPermissions)
				permissions.POST("/", permissionHandler.CreatePermission)
				permissions.GET("/:id", permissionHandler.GetPermission)
				permissions.PUT("/:id", permissionHandler.UpdatePermission)
				permissions.DELETE("/:id", permissionHandler.DeletePermission)
			}

			// 角色权限管理器
			rolePermissions := superAdmin.Group("/role-permissions")
			{
				rolePermissions.GET("/:role", permissionHandler.GetRolePermissions)
				rolePermissions.POST("/:role", permissionHandler.AssignRolePermission)
				rolePermissions.DELETE("/:role/:permissionId", permissionHandler.RevokeRolePermission)
			}

			// 用户权限管理器
			userPermissions := superAdmin.Group("/user-permissions")
			{
				userPermissions.GET("/:userId", permissionHandler.GetUserPermissions)
				userPermissions.POST("/:userId", permissionHandler.AssignUserPermission)
				userPermissions.DELETE("/:userId/:permissionId", permissionHandler.RevokeUserPermission)
			}

			// 系统监控管理器
			system := superAdmin.Group("/system")
			{
				system.GET("/health", systemHandler.HealthCheck)
				system.GET("/info", systemHandler.GetSystemInfo)
				system.GET("/metrics", systemHandler.GetMetrics)
				system.POST("/cleanup", systemHandler.CleanupSystem)
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
