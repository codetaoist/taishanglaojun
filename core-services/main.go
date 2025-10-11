package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ai_integration "github.com/codetaoist/taishanglaojun/core-services/ai-integration"
	"github.com/codetaoist/taishanglaojun/core-services/community"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness"
	cultural_wisdom "github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom"
	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	location_tracking "github.com/codetaoist/taishanglaojun/core-services/location-tracking"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CoreServices 核心服务管理�?
type CoreServices struct {
	config      *config.Config
	logger      *zap.Logger
	db          *database.Database
	redisClient *redis.Client
	httpServer  *http.Server

	// 服务模块
	consciousnessModule  *consciousness.Module
	culturalWisdomModule *cultural_wisdom.Module
	aiIntegrationModule  *ai_integration.Module
	communityModule      *community.Module
	locationTrackingModule *location_tracking.Module
}

// NewCoreServices 创建核心服务管理�?
func NewCoreServices() (*CoreServices, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 初始化日�?
	zapLogger, err := logger.NewLogger(cfg.LogLevel, cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// 初始化数据库
	db, err := database.NewDatabase(cfg.Database, zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试Redis连接
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		zapLogger.Warn("Redis connection failed", zap.Error(err))
	}

	return &CoreServices{
		config:      cfg,
		logger:      zapLogger,
		db:          db,
		redisClient: redisClient,
	}, nil
}

// Initialize 初始化所有服务模�?
func (cs *CoreServices) Initialize() error {
	cs.logger.Info("Initializing core services")

	// 初始化意识服务模�?
	consciousnessModule, err := consciousness.NewModule(cs.db.DB, cs.redisClient, cs.logger, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize consciousness module: %w", err)
	}
	cs.consciousnessModule = consciousnessModule

	// 初始化文化智慧模�?
	culturalWisdomModule, err := cultural_wisdom.NewModule(cs.db.DB, cs.redisClient, cs.logger, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize cultural wisdom module: %w", err)
	}
	cs.culturalWisdomModule = culturalWisdomModule

	// 初始化AI集成模块
	aiIntegrationModule, err := ai_integration.NewModule(nil, cs.db.DB, cs.redisClient, cs.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize AI integration module: %w", err)
	}
	cs.aiIntegrationModule = aiIntegrationModule

	// 初始化社区模�?
	communityModule, err := community.NewModule(cs.db.DB, cs.redisClient, cs.logger, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize community module: %w", err)
	}
	cs.communityModule = communityModule

	// 初始化位置追踪模�?
	locationTrackingModule, err := location_tracking.NewModule(cs.db.DB, cs.redisClient, cs.logger, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize location tracking module: %w", err)
	}
	cs.locationTrackingModule = locationTrackingModule

	cs.logger.Info("All core services initialized successfully")
	return nil
}

// SetupHTTPServer 设置HTTP服务�?
func (cs *CoreServices) SetupHTTPServer() error {
	cs.logger.Info("Setting up HTTP server")

	// 设置Gin模式
	if cs.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	router := gin.New()

	// 添加中间�?
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS配置
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// 添加全局中间�?
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(cs.logger))

	// 健康检查端�?
	router.GET("/health", cs.healthCheck)
	router.GET("/health/detailed", cs.detailedHealthCheck)

	// API版本路由�?
	apiV1 := router.Group("/api/v1")

	// 设置各模块路�?
	if err := cs.consciousnessModule.SetupRoutes(apiV1); err != nil {
		return fmt.Errorf("failed to setup consciousness routes: %w", err)
	}

	if err := cs.culturalWisdomModule.SetupRoutes(apiV1); err != nil {
		return fmt.Errorf("failed to setup cultural wisdom routes: %w", err)
	}

	if err := cs.aiIntegrationModule.SetupRoutes(apiV1); err != nil {
		return fmt.Errorf("failed to setup AI integration routes: %w", err)
	}

	if err := cs.communityModule.SetupRoutes(apiV1); err != nil {
		return fmt.Errorf("failed to setup community routes: %w", err)
	}

	if err := cs.locationTrackingModule.SetupRoutes(apiV1); err != nil {
		return fmt.Errorf("failed to setup location tracking routes: %w", err)
	}

	// 创建HTTP服务�?
	cs.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cs.config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cs.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cs.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cs.config.Server.IdleTimeout) * time.Second,
	}

	cs.logger.Info("HTTP server setup completed", zap.Int("port", cs.config.Server.Port))
	return nil
}

// Start 启动所有服�?
func (cs *CoreServices) Start() error {
	cs.logger.Info("Starting core services")

	// 启动各个模块
	if err := cs.consciousnessModule.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness module: %w", err)
	}

	if err := cs.culturalWisdomModule.Start(); err != nil {
		return fmt.Errorf("failed to start cultural wisdom module: %w", err)
	}

	if err := cs.aiIntegrationModule.Start(); err != nil {
		return fmt.Errorf("failed to start AI integration module: %w", err)
	}

	if err := cs.communityModule.Start(); err != nil {
		return fmt.Errorf("failed to start community module: %w", err)
	}

	if err := cs.locationTrackingModule.Start(); err != nil {
		return fmt.Errorf("failed to start location tracking module: %w", err)
	}

	// 启动HTTP服务�?
	go func() {
		cs.logger.Info("Starting HTTP server", zap.String("address", cs.httpServer.Addr))
		if err := cs.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cs.logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	cs.logger.Info("All core services started successfully")
	return nil
}

// Stop 停止所有服�?
func (cs *CoreServices) Stop() error {
	cs.logger.Info("Stopping core services")

	// 停止HTTP服务�?
	if cs.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := cs.httpServer.Shutdown(ctx); err != nil {
			cs.logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}

	// 停止各个模块
	if cs.locationTrackingModule != nil {
		if err := cs.locationTrackingModule.Stop(); err != nil {
			cs.logger.Error("Failed to stop location tracking module", zap.Error(err))
		}
	}

	if cs.communityModule != nil {
		if err := cs.communityModule.Stop(); err != nil {
			cs.logger.Error("Failed to stop community module", zap.Error(err))
		}
	}

	if cs.aiIntegrationModule != nil {
		if err := cs.aiIntegrationModule.Stop(); err != nil {
			cs.logger.Error("Failed to stop AI integration module", zap.Error(err))
		}
	}

	if cs.culturalWisdomModule != nil {
		if err := cs.culturalWisdomModule.Stop(); err != nil {
			cs.logger.Error("Failed to stop cultural wisdom module", zap.Error(err))
		}
	}

	if cs.consciousnessModule != nil {
		if err := cs.consciousnessModule.Stop(); err != nil {
			cs.logger.Error("Failed to stop consciousness module", zap.Error(err))
		}
	}

	// 关闭数据库连�?
	if cs.db != nil {
		if err := cs.db.Close(); err != nil {
			cs.logger.Error("Failed to close database", zap.Error(err))
		}
	}

	// 关闭Redis连接
	if cs.redisClient != nil {
		if err := cs.redisClient.Close(); err != nil {
			cs.logger.Error("Failed to close Redis client", zap.Error(err))
		}
	}

	cs.logger.Info("All core services stopped successfully")
	return nil
}

// healthCheck 简单健康检�?
func (cs *CoreServices) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "taishang-core-services",
		"version":   "1.0.0",
	})
}

// detailedHealthCheck 详细健康检�?
func (cs *CoreServices) detailedHealthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "taishang-core-services",
		"version":   "1.0.0",
		"modules":   gin.H{},
	}

	// 检查各个模块健康状�?
	if cs.consciousnessModule != nil {
		health["modules"].(gin.H)["consciousness"] = cs.consciousnessModule.Health()
	}

	if cs.culturalWisdomModule != nil {
		health["modules"].(gin.H)["cultural_wisdom"] = cs.culturalWisdomModule.Health()
	}

	if cs.aiIntegrationModule != nil {
		health["modules"].(gin.H)["ai_integration"] = cs.aiIntegrationModule.Health()
	}

	if cs.communityModule != nil {
		health["modules"].(gin.H)["community"] = cs.communityModule.Health()
	}

	if cs.locationTrackingModule != nil {
		health["modules"].(gin.H)["location_tracking"] = cs.locationTrackingModule.Health()
	}

	// 检查数据库连接
	if cs.db != nil {
		if sqlDB, err := cs.db.DB.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				health["database"] = "unhealthy"
				health["status"] = "degraded"
			} else {
				health["database"] = "healthy"
			}
		}
	}

	// 检查Redis连接
	if cs.redisClient != nil {
		if err := cs.redisClient.Ping(context.Background()).Err(); err != nil {
			health["redis"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["redis"] = "healthy"
		}
	}

	c.JSON(http.StatusOK, health)
}

func main() {
	// 创建核心服务管理�?
	coreServices, err := NewCoreServices()
	if err != nil {
		log.Fatalf("Failed to create core services: %v", err)
	}

	// 初始化服�?
	if err := coreServices.Initialize(); err != nil {
		log.Fatalf("Failed to initialize core services: %v", err)
	}

	// 设置HTTP服务�?
	if err := coreServices.SetupHTTPServer(); err != nil {
		log.Fatalf("Failed to setup HTTP server: %v", err)
	}

	// 启动服务
	if err := coreServices.Start(); err != nil {
		log.Fatalf("Failed to start core services: %v", err)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	coreServices.logger.Info("Shutting down server...")

	// 优雅关闭
	if err := coreServices.Stop(); err != nil {
		log.Fatalf("Failed to stop core services: %v", err)
	}

	coreServices.logger.Info("Server exited")
}
