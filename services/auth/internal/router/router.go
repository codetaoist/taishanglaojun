package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"github.com/codetaoist/taishanglaojun/auth/internal/config"
	"github.com/codetaoist/taishanglaojun/auth/internal/handler"
	"github.com/codetaoist/taishanglaojun/auth/internal/middleware"
	"github.com/codetaoist/taishanglaojun/auth/internal/repository"
	"github.com/codetaoist/taishanglaojun/auth/internal/service"
)

// Setup sets up the router with all routes and middleware
func Setup(cfg config.Config, r *gin.Engine, db *sql.DB) {
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	blacklistRepo := repository.NewBlacklistRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, sessionRepo, blacklistRepo, cfg.JWTSecret, cfg.JWTExpiration)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "auth",
		})
	})

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// User profile routes
			protected.GET("/profile", authHandler.GetProfile)
			protected.POST("/change-password", authHandler.ChangePassword)
			protected.POST("/revoke-token", authHandler.RevokeToken)

			// Admin routes (admin role required)
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.GET("/users/:id", authHandler.GetUser)
			}
		}

		// Protected auth routes (authentication required)
		protectedAuth := v1.Group("/auth")
		protectedAuth.Use(middleware.AuthMiddleware(authService))
		{
			protectedAuth.POST("/logout", authHandler.Logout)
		}
	}
}