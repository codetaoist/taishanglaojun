package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	ai_integration "github.com/taishanglaojun/core-services/ai-integration"
	"github.com/taishanglaojun/core-services/ai-integration/providers"
	"github.com/taishanglaojun/core-services/internal/config"
	"github.com/taishanglaojun/core-services/internal/database"
	"github.com/taishanglaojun/core-services/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	log, err := logger.New(cfg.Logging)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting core services", 
		zap.String("version", "v1.0.0"),
		zap.String("mode", cfg.Server.Mode))

	// 初始化数据库
	db, err := database.New(cfg.Database, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// 初始化Redis
	redisClient, err := database.NewRedis(cfg.Redis, log)
	if err != nil {
		log.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 初始化AI提供商管理器
	providerManager := providers.NewManager()
	
	// 注册OpenAI提供商
	if cfg.AI.Providers.OpenAI.APIKey != "" {
		openaiProvider := providers.NewOpenAIProvider(cfg.AI.Providers.OpenAI)
		if err := providerManager.RegisterProvider("openai", openaiProvider); err != nil {
			log.Error("Failed to register OpenAI provider", zap.Error(err))
		} else {
			log.Info("OpenAI provider registered successfully")
		}
	}

	// 数据库迁移
	if err := autoMigrate(db.GetDB()); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由器
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(cfg.Server.CORS))
	router.Use(requestIDMiddleware())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "core-services",
		})
	})

	// API路由组
	apiV1 := router.Group("/api/v1")

	// 设置AI集成服务路由
	ai_integration.SetupRoutes(apiV1, db.GetDB(), log, providerManager)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Info("Server starting", 
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port))
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}