package routes

import (
	"errors"
	
	"github.com/google/uuid"
	"github.com/taishanglaojun/auth_system/internal/handler"
	"github.com/taishanglaojun/auth_system/internal/handlers"
	"github.com/taishanglaojun/auth_system/internal/middleware"
	"github.com/taishanglaojun/auth_system/internal/models"
	"github.com/taishanglaojun/auth_system/internal/repository"
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
	db *gorm.DB,
	logger *zap.Logger,
) {
	// 全局中间件
	router.Use(gin.Recovery())
	router.Use(authMiddleware.RequestLogger())
	router.Use(authMiddleware.CORS())
	router.Use(authMiddleware.SecurityHeaders())

	// API版本分组
	v1 := router.Group("/api/v1")
	{
		// 健康检查
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

			// 会话管理路由
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
			// 用户管理
			users := admin.Group("/users")
			{
				users.GET("", func(c *gin.Context) {
					// TODO: 实现用户列表
					c.JSON(200, gin.H{"message": "User list endpoint"})
				})
				users.GET("/:userId", func(c *gin.Context) {
					// TODO: 实现获取用户详情
					c.JSON(200, gin.H{"message": "User detail endpoint"})
				})
				users.PUT("/:userId", func(c *gin.Context) {
					// TODO: 实现更新用户
					c.JSON(200, gin.H{"message": "Update user endpoint"})
				})
				users.DELETE("/:userId", func(c *gin.Context) {
					// TODO: 实现删除用户
					c.JSON(200, gin.H{"message": "Delete user endpoint"})
				})
				users.PUT("/:userId/status", func(c *gin.Context) {
					// 实现更新用户状态
					userID := c.Param("userId")
					
					var req struct {
						Status string `json:"status" binding:"required,oneof=active inactive suspended"`
						Reason string `json:"reason"`
					}
					
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(400, gin.H{"error": "Invalid request body", "details": err.Error()})
						return
					}
					
					// 验证UUID格式
					if _, err := uuid.Parse(userID); err != nil {
						c.JSON(400, gin.H{"error": "Invalid user ID format"})
						return
					}
					
					// 获取当前管理员信息
					adminUser, exists := c.Get("user")
					if !exists {
						c.JSON(401, gin.H{"error": "Unauthorized"})
						return
					}
					
					admin := adminUser.(*models.User)
					
					// 调用服务层更新用户状态
					ctx := c.Request.Context()
					userRepo := repository.NewUserRepository(db, logger)
					
					// 检查目标用户是否存在
					targetUser, err := userRepo.GetByID(ctx, uuid.MustParse(userID))
					if err != nil {
						if errors.Is(err, repository.ErrUserNotFound) {
							c.JSON(404, gin.H{"error": "User not found"})
							return
						}
						logger.Error("Failed to get user", zap.Error(err))
						c.JSON(500, gin.H{"error": "Internal server error"})
						return
					}
					
					// 防止管理员修改自己的状态
					if targetUser.ID == admin.ID {
						c.JSON(400, gin.H{"error": "Cannot modify your own status"})
						return
					}
					
					// 更新用户状态
					if err := userRepo.UpdateStatus(ctx, targetUser.ID, models.UserStatus(req.Status)); err != nil {
						logger.Error("Failed to update user status", 
							zap.String("user_id", userID),
							zap.String("status", req.Status),
							zap.Error(err),
						)
						c.JSON(500, gin.H{"error": "Failed to update user status"})
						return
					}
					
					// 记录操作日志
					logger.Info("User status updated by admin",
						zap.String("admin_id", admin.ID.String()),
						zap.String("admin_username", admin.Username),
						zap.String("target_user_id", userID),
						zap.String("target_username", targetUser.Username),
						zap.String("old_status", string(targetUser.Status)),
						zap.String("new_status", req.Status),
						zap.String("reason", req.Reason),
					)
					
					// 返回更新后的用户信息
					updatedUser, _ := userRepo.GetByID(ctx, targetUser.ID)
					c.JSON(200, gin.H{
						"message": "User status updated successfully",
						"user": updatedUser.ToPublic(),
					})
				})
			}

			// 会话管理
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

			// 系统统计
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
			// 权限管理
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

			// 角色权限管理
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

			// 用户权限管理
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

			// 系统管理
			system := superAdmin.Group("/system")
			{
				system.GET("/info", func(c *gin.Context) {
					// TODO: 实现系统信息
					c.JSON(200, gin.H{"message": "System info endpoint"})
				})
				system.POST("/cleanup", func(c *gin.Context) {
					// TODO: 实现系统清理
					c.JSON(200, gin.H{"message": "System cleanup endpoint"})
				})
			}

			// 动态数据库管理
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