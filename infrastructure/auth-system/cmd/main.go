package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/database"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/handler"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/handlers"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/jwt"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/middleware"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/monitoring"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/repository"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/routes"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.New(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync(log)

	// 设置全局日志
	zap.ReplaceGlobals(log)

	log.Info("Starting auth-system service",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode),
		zap.String("address", cfg.GetServerAddr()),
	)

	// 初始化数量据库
	db, err := database.New(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database", zap.Error(err))
		}
	}()

	// 初始化动态数量据库管理器
	dynamicDB := database.NewDynamicDatabase(log)

	// 添加默认数量据库配置
	if err := dynamicDB.AddDatabase("default", cfg); err != nil {
		log.Fatal("Failed to add default database configuration", zap.Error(err))
	}

	// 初始化JWT管理器
	jwtConfig := &jwt.Config{
		SecretKey:        cfg.JWT.SecretKey,
		AccessTokenTTL:   cfg.JWT.AccessTokenTTL,
		RefreshTokenTTL:  cfg.JWT.RefreshTokenTTL,
		Issuer:           cfg.JWT.Issuer,
		RefreshThreshold: cfg.JWT.RefreshThreshold,
		RequireAudience:  cfg.JWT.RequireAudience,
		AllowedAudiences: cfg.JWT.AllowedAudiences,
		MaxTokenAge:      cfg.JWT.MaxTokenAge,
	}
	jwtManager := jwt.NewManager(jwtConfig, log)

	// 初始化仓库层
	userRepo := repository.NewUserRepository(db.GetDB(), log)
	sessionRepo := repository.NewSessionRepository(db.GetDB(), log)
	tokenRepo := repository.NewTokenRepository(db.GetDB(), log)

	// 初始化邮件服务
	emailService, err := service.NewEmailService(cfg.Email, log)
	if err != nil {
		log.Fatal("Failed to initialize email service", zap.Error(err))
	}

	// 初始化服务层
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		tokenRepo,
		emailService,
		jwtManager,
		log,
	)

	// 初始化中间件件
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, authService, log)

	// 初始化限流中间件
	rateLimiterConfig := &middleware.RateLimiterConfig{
		RequestsPerWindow: cfg.Security.RateLimitRequests,
		WindowDuration:    time.Minute,
		BurstSize:         10,
		CleanupInterval:   5 * time.Minute,
		LoginLimit:        cfg.Security.MaxLoginAttempts,
		RegisterLimit:     3, // 每分钟最多3次注册尝试
	}
	rateLimiter := middleware.NewRateLimiter(rateLimiterConfig, log)

	// 初始化Prometheus指标收集
	metrics := monitoring.NewMetrics(log)

	// 初始化处理器器
	authHandler := handler.NewAuthHandler(authService, log)

	// 初始化动态数量据库处理器
	databaseHandler := handlers.NewDatabaseHandler(dynamicDB, log)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	router := gin.New()

	// 添加Prometheus指标中间件
	router.Use(metrics.GinMiddleware())

	// 设置路由
	routes.SetupRoutes(router, authHandler, databaseHandler, authMiddleware, rateLimiter, db.GetDB(), log, userRepo, sessionRepo, tokenRepo, jwtManager, authService, metrics)

	// 创建HTTP服务
	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 启动服务
	go func() {
		log.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 启动后台任务
	go startBackgroundTasks(cfg, db, tokenRepo, sessionRepo, metrics, log)

	// 等待中间件断信息号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	} else {
		log.Info("Server shutdown completed")
	}
}

// startBackgroundTasks 启动后台任务
func startBackgroundTasks(
	cfg *config.Config,
	db *database.Database,
	tokenRepo repository.TokenRepository,
	sessionRepo repository.SessionRepository,
	metrics *monitoring.Metrics,
	log *zap.Logger,
) {
	ticker := time.NewTicker(cfg.Security.TokenCleanupInterval)
	defer ticker.Stop()

	log.Info("Starting background tasks",
		zap.Duration("cleanup_interval", cfg.Security.TokenCleanupInterval),
	)

	for {
		select {
		case <-ticker.C:
			cleanupExpiredTokensAndSessions(tokenRepo, sessionRepo, metrics, log)
		}
	}
}

// cleanupExpiredTokensAndSessions 清理过期的令牌和会话
func cleanupExpiredTokensAndSessions(
	tokenRepo repository.TokenRepository,
	sessionRepo repository.SessionRepository,
	metrics *monitoring.Metrics,
	log *zap.Logger,
) {
	ctx := context.Background()

	// 清理过期令牌
	if deletedTokens, err := tokenRepo.CleanupExpiredTokens(ctx); err != nil {
		log.Error("Failed to cleanup expired tokens", zap.Error(err))
		metrics.RecordSystemCleanup("tokens", "expired", "error")
	} else if deletedTokens > 0 {
		log.Info("Cleaned up expired tokens", zap.Int64("count", deletedTokens))
		metrics.RecordSystemCleanup("tokens", "expired", "success")
	}

	// 清理已使用户的令牌
	if deletedTokens, err := tokenRepo.CleanupUsedTokens(ctx, 24*time.Hour); err != nil {
		log.Error("Failed to cleanup used tokens", zap.Error(err))
		metrics.RecordSystemCleanup("tokens", "used", "error")
	} else if deletedTokens > 0 {
		log.Info("Cleaned up used tokens", zap.Int64("count", deletedTokens))
		metrics.RecordSystemCleanup("tokens", "used", "success")
	}

	// 清理已撤销的令牌
	if deletedTokens, err := tokenRepo.CleanupRevokedTokens(ctx, 24*time.Hour); err != nil {
		log.Error("Failed to cleanup revoked tokens", zap.Error(err))
		metrics.RecordSystemCleanup("tokens", "revoked", "error")
	} else if deletedTokens > 0 {
		log.Info("Cleaned up revoked tokens", zap.Int64("count", deletedTokens))
		metrics.RecordSystemCleanup("tokens", "revoked", "success")
	}

	// 清理过期会话
	if deletedSessions, err := sessionRepo.CleanupExpiredSessions(ctx); err != nil {
		log.Error("Failed to cleanup expired sessions", zap.Error(err))
		metrics.RecordSystemCleanup("sessions", "expired", "error")
	} else if deletedSessions > 0 {
		log.Info("Cleaned up expired sessions", zap.Int64("count", deletedSessions))
		metrics.RecordSystemCleanup("sessions", "expired", "success")
	}

	// 清理已撤销的会话
	if deletedSessions, err := sessionRepo.CleanupRevokedSessions(ctx); err != nil {
		log.Error("Failed to cleanup revoked sessions", zap.Error(err))
		metrics.RecordSystemCleanup("sessions", "revoked", "error")
	} else if deletedSessions > 0 {
		log.Info("Cleaned up revoked sessions", zap.Int64("count", deletedSessions))
		metrics.RecordSystemCleanup("sessions", "revoked", "success")
	}
}
