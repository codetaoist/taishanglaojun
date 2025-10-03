package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ai_integration "github.com/codetaoist/taishanglaojun/core-services/ai-integration"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	cultural_wisdom "github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom"
	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	location_tracking "github.com/codetaoist/taishanglaojun/core-services/location-tracking"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	logConfig := logger.LogConfig{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		Filename:   cfg.Logger.Filename,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	}
	log, err := logger.New(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer log.Sync()

	log.Info("Starting core services",
		zap.String("version", "v1.0.0"),
		zap.String("mode", cfg.Server.Mode))

	// 初始化数据库
	dbConfig := database.Config{
		Driver:          cfg.Database.Type,
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Database,
		Username:        cfg.Database.Username,
		Password:        cfg.Database.Password,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.Database.MaxLifetime) * time.Second,
		SSLMode:         cfg.Database.SSLMode,
		ConnectTimeout:  30 * time.Second, // 默认连接超时
	}

	// 连接数据库
	db, err := database.New(dbConfig, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	// defer db.Close()

	// 初始化Redis（如果启用）
	var redisClient *database.Redis
	if cfg.Redis.Enabled {
		redisConfig := database.RedisConfig{
			Host:         cfg.Redis.Host,
			Port:         cfg.Redis.Port,
			Password:     cfg.Redis.Password,
			Database:     cfg.Redis.Database,
			PoolSize:     cfg.Redis.PoolSize,
			MinIdleConns: cfg.Redis.MinIdleConns,
		}

		var err error
		redisClient, err = database.NewRedis(redisConfig, log)
		if err != nil {
			log.Fatal("Failed to initialize Redis", zap.Error(err))
		}
		defer redisClient.Close()
		log.Info("Redis connected successfully")
	} else {
		log.Info("Redis is disabled in configuration")
	}

	// 初始化AI提供商管理器
	providerManager := providers.NewManager(log)

	// 注册模拟提供商（用于开发和测试）
	mockProvider := providers.NewMockProvider(log)
	if err := providerManager.RegisterProvider("mock", mockProvider); err != nil {
		log.Error("Failed to register mock provider", zap.Error(err))
	} else {
		log.Info("Mock provider registered successfully")
		// 设置模拟提供者为默认提供者
		if err := providerManager.SetDefaultProvider("mock"); err != nil {
			log.Error("Failed to set mock as default provider", zap.Error(err))
		} else {
			log.Info("Mock provider set as default")
		}
	}

	// 注册OpenAI提供商
	if openaiConfig, exists := cfg.AI.Providers["openai"]; exists && openaiConfig.Enabled {
		if apiKey, ok := openaiConfig.Config["api_key"].(string); ok && apiKey != "" {
			baseURL := "https://api.openai.com/v1"
			if url, ok := openaiConfig.Config["base_url"].(string); ok && url != "" {
				baseURL = url
			}
			timeout := 30
			if t, ok := openaiConfig.Config["timeout"].(int); ok {
				timeout = t
			}

			openaiProviderConfig := providers.OpenAIConfig{
				APIKey:  apiKey,
				BaseURL: baseURL,
				Timeout: timeout,
			}
			openaiProvider := providers.NewOpenAIProvider(openaiProviderConfig, log)
			if err := providerManager.RegisterProvider("openai", openaiProvider); err != nil {
				log.Error("Failed to register OpenAI provider", zap.Error(err))
			} else {
				log.Info("OpenAI provider registered successfully")
				// 如果OpenAI配置有效，则设置为默认提供者
				if err := providerManager.SetDefaultProvider("openai"); err != nil {
					log.Error("Failed to set OpenAI as default provider", zap.Error(err))
				} else {
					log.Info("OpenAI set as default provider")
				}
			}
		}
	}

	// 数据库迁移
	if err := autoMigrate(db.GetDB()); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 初始化JWT中间件
	jwtConfig := middleware.JWTConfig{
		Secret:     cfg.JWT.Secret,
		Issuer:     cfg.JWT.Issuer,
		Expiration: cfg.JWT.Expiration,
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig, log)

	// 初始化认证服务
	authService := middleware.NewAuthService(db.GetDB(), jwtMiddleware, log)
	// 迁移认证表（注释掉以避免外键约束问题）
	// if err := authService.AutoMigrate(); err != nil {
	// 	log.Fatal("Failed to migrate auth tables", zap.Error(err))
	// }

	// 初始化认证处理器
	authHandler := middleware.NewAuthHandler(authService, log)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(cfg.CORS))
	router.Use(requestIDMiddleware())

	// 健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "core-services",
		})
	})

	// API路由组
	apiV1 := router.Group("/api/v1")

	// 设置认证路由
	middleware.SetupAuthRoutes(apiV1, authHandler, jwtMiddleware)

	// 设置AI集成服务路由
	ai_integration.SetupRoutes(apiV1, db.GetDB(), log, providerManager)

	// 设置文化智慧路由
	var redisClientPtr *redis.Client
	if redisClient != nil {
		redisClientPtr = redisClient.GetClient()
	}
	
	log.Info("=== Starting cultural wisdom routes setup ===")
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC in cultural wisdom routes setup", zap.Any("error", r))
		}
	}()
	
	// 尝试设置文化智慧路由
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("PANIC during cultural_wisdom.SetupRoutes", zap.Any("error", r))
				return
			}
		}()
		log.Info("Calling cultural_wisdom.SetupRoutes")
		cultural_wisdom.SetupRoutes(apiV1, db.GetDB(), redisClientPtr, log, jwtMiddleware, providerManager)
		log.Info("cultural_wisdom.SetupRoutes completed successfully")
	}()

	// 设置位置跟踪路由
	log.Info("=== Starting location tracking routes setup ===")
	location_tracking.SetupRoutes(apiV1, db.GetDB(), log, jwtMiddleware)
	
	log.Info("=== Cultural wisdom routes setup completed ===")

	// 创建HTTP服务
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 启动服务
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
