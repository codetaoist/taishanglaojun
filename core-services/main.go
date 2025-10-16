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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness"
	cultural_wisdom "github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom"
	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/routes"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// CoreServices 核心服务管理器
type CoreServices struct {
	config      *config.Config
	logger      *zap.Logger
	db          *database.Database
	redisClient *redis.Client
	httpServer  *http.Server

	// 认证相关
	jwtMiddleware *middleware.JWTMiddleware
	authService   *middleware.AuthService
	authHandler   *middleware.AuthHandler
	adminHandler  *handlers.AdminHandler

	// 业务服务
	userService *services.UserService
	menuService *services.MenuService

	// 性能优化服务
	performanceMiddleware *middleware.PerformanceMiddleware
	optimizationService   *services.OptimizationService

	// 服务模块
	consciousnessModule  *consciousness.Module
	culturalWisdomModule *cultural_wisdom.Module
}

// NewCoreServices 创建核心服务管理器
func NewCoreServices() (*CoreServices, error) {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 初始化日志记录器
	zapLogger, err := logger.New(logger.LogConfig{
		Level:      cfg.Logger.Level,
		Format:     cfg.Logger.Format,
		Output:     cfg.Logger.Output,
		Filename:   cfg.Logger.Filename,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// 初始化数据库
	db, err := database.New(database.Config{
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
		ConnectTimeout:  30 * time.Second,
	}, zapLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
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

// Initialize 初始化所有服务模块
func (cs *CoreServices) Initialize() error {
	cs.logger.Info("Initializing core services")

	// 初始化JWT中间件
	expiration, err := time.ParseDuration(cs.config.JWT.Expiration)
	if err != nil {
		expiration = 24 * time.Hour // 默认24小时
	}
	jwtConfig := middleware.JWTConfig{
		Secret:     cs.config.JWT.Secret, // 从配置文件读取
		Issuer:     cs.config.JWT.Issuer,
		Expiration: expiration,
	}
	cs.jwtMiddleware = middleware.NewJWTMiddleware(jwtConfig, cs.logger)

	// 初始化认证服务
	cs.authService = middleware.NewAuthService(cs.db.GetDB(), cs.jwtMiddleware, cs.logger)
	
	// 执行完整的数据库迁移
	migrationService := database.NewMigrationService(cs.db.GetDB(), cs.logger)
	if err := migrationService.RunMigration(); err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}

	// 初始化认证处理器
	cs.authHandler = middleware.NewAuthHandler(cs.authService, cs.logger, cs.db.GetDB())
	
	// 初始化管理员处理器
	cs.adminHandler = handlers.NewAdminHandler(cs.authService, cs.logger, cs.db.GetDB())

	// 初始化业务服务
	cs.userService = services.NewUserService(cs.db.GetDB(), cs.logger)
	cs.menuService = services.NewMenuService(cs.db.GetDB(), cs.logger)

	// 初始化性能优化服务（暂时注释掉，因为参数不匹配）
	// cs.optimizationService = services.NewOptimizationService(cs.db.GetDB(), cs.logger)
	
	// 初始化性能中间件
	performanceConfig := middleware.PerformanceConfig{
		EnableCache:       true,
		EnableRateLimit:   true,
		EnableCompression: true,
		EnableMetrics:     true,
		CacheTTL:          300 * time.Second, // 5分钟
		RateLimit:         100,               // 每分钟100次请求
		RateBurst:         20,                // 突发限制
		CompressionLevel:  6,                 // 压缩级别
	}
	cs.performanceMiddleware = middleware.NewPerformanceMiddleware(cs.redisClient, cs.logger, performanceConfig)

	// 初始化意识服务模块
	consciousnessModule, err := consciousness.NewModule(nil, cs.db.GetDB(), cs.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize consciousness module: %w", err)
	}
	cs.consciousnessModule = consciousnessModule

	// 初始化文化智慧模块 - 暂时跳过，因为需要更多依赖
	// culturalWisdomModule, err := cultural_wisdom.NewModule(nil, cs.db.GetDB(), cs.redisClient, cs.logger, nil)
	// if err != nil {
	// 	return fmt.Errorf("failed to initialize cultural wisdom module: %w", err)
	// }
	// cs.culturalWisdomModule = culturalWisdomModule

	cs.logger.Info("All core services initialized successfully")
	return nil
}

// SetupHTTPServer 设置HTTP服务
func (cs *CoreServices) SetupHTTPServer() error {
	cs.logger.Info("Setting up HTTP server")

	// 设置Gin模式
	if cs.config.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin引擎
	router := gin.New()

	// 添加基础中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 应用性能优化中间件
	router.Use(cs.performanceMiddleware.RateLimitMiddleware())
	router.Use(cs.performanceMiddleware.CompressionMiddleware())
	router.Use(cs.performanceMiddleware.MetricsMiddleware())

	// 设置CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

    // 设置上传相关的限制与静态服务
    // 允许最大 32MB 的 multipart 内存缓冲（大文件仍将落盘）
    router.MaxMultipartMemory = 32 << 20
    // 提供上传文件的静态访问
    router.Static("/uploads", "./uploads")

    // 健康检查端点
    router.GET("/health", cs.healthCheck)
    router.GET("/health/detailed", cs.detailedHealthCheck)
	
	// 性能监控端点
	// 添加性能指标端点
	router.GET("/metrics", func(c *gin.Context) {
		metrics := cs.performanceMiddleware.GetMetrics()
		c.JSON(200, metrics)
	})
	
	// 数据库优化端点
	router.GET("/api/v1/optimization/slow-queries", 
		cs.jwtMiddleware.AuthRequired(),
		func(c *gin.Context) {
			queries := cs.optimizationService.GetSlowQueries()
			c.JSON(http.StatusOK, gin.H{
				"slow_queries": queries,
			})
		})
	
	// 根路径测试路由
	router.GET("/debug-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Root route working"})
	})

    // API版本路由
    apiV1 := router.Group("/api/v1")

	// 添加简单测试路由
	apiV1.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API v1 working"})
	})

    // 设置认证路由
    middleware.SetupAuthRoutes(apiV1, cs.authHandler, cs.jwtMiddleware)

    // 通用上传端点（需认证）
    router.POST("/api/upload", cs.jwtMiddleware.AuthRequired(), handlers.NewUploadHandler(cs.logger).HandleUpload)

	// 设置用户相关路由
	userGroup := apiV1.Group("/user")
	userGroup.Use(cs.jwtMiddleware.AuthRequired())
	{
		userGroup.GET("/me", 
			cs.performanceMiddleware.CacheMiddleware(300), // 缓存5分钟
			cs.authHandler.GetCurrentUser)
	}

	// 添加直接测试路由
	apiV1.GET("/admin/direct-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Direct admin route working"})
	})

	// 设置管理员路由
	admin := apiV1.Group("/admin")
	admin.Use(cs.jwtMiddleware.AuthRequired())
	{
		admin.GET("/users", 
			cs.performanceMiddleware.CacheMiddleware(180*time.Second), // 缓存3分钟
			cs.adminHandler.GetUsers)
		admin.GET("/users/stats",
			cs.performanceMiddleware.CacheMiddleware(60*time.Second), // 缓存1分钟
			cs.adminHandler.GetUserStats)
		admin.POST("/users", cs.adminHandler.CreateUser)
		admin.PUT("/users/:id", cs.adminHandler.UpdateUser)
		admin.DELETE("/users/:id", cs.adminHandler.DeleteUser)
		admin.POST("/users/batch-delete", cs.adminHandler.BatchDeleteUsers)
		admin.PUT("/users/:id/status", cs.adminHandler.UpdateUserStatus)
	}

	// 设置菜单路由
	routes.SetupMenuRoutes(apiV1, cs.jwtMiddleware, cs.db.GetDB(), cs.logger)

	// 设置dashboard路由
	routes.SetupDashboardRoutes(apiV1, cs.authService, cs.jwtMiddleware, cs.userService, cs.menuService, cs.db.GetDB(), cs.logger)

    // 设置权限路由
	routes.SetupPermissionRoutes(apiV1, cs.jwtMiddleware, cs.db.GetDB(), cs.redisClient, cs.logger)

	// 设置系统设置路由
	routes.SetupSystemRoutes(apiV1, cs.jwtMiddleware, cs.db.GetDB(), cs.logger)

	// 设置增强数据库监控路由
	routes.SetupEnhancedDatabaseRoutes(apiV1, cs.jwtMiddleware, cs.db.GetDB(), cs.logger)

	// 设置API文档管理路由
	apiDocService := services.NewAPIDocumentationService(cs.db.GetDB(), cs.logger)
	apiDocHandler := handlers.NewAPIDocumentationHandler(apiDocService, cs.authService, cs.logger)
	routes.SetupAPIDocumentationRoutes(router, apiDocHandler, cs.jwtMiddleware, cs.logger)

	// 设置各模块路由
	if err := cs.consciousnessModule.SetupRoutes(apiV1, gin.Logger()); err != nil {
		return fmt.Errorf("failed to setup consciousness routes: %w", err)
	}

	// 暂时跳过其他模块的路由设置
	// if cs.culturalWisdomModule != nil {
	// 	if err := cs.culturalWisdomModule.SetupRoutes(apiV1); err != nil {
	// 		return fmt.Errorf("failed to setup cultural wisdom routes: %w", err)
	// 	}
	// }

	// 创建HTTP服务器
	cs.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cs.config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cs.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cs.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	cs.logger.Info("HTTP server setup completed", zap.Int("port", cs.config.Server.Port))
	return nil
}

// Start 启动所有服务模块
func (cs *CoreServices) Start() error {
	cs.logger.Info("Starting core services")

	// 启动各个模块
	if err := cs.consciousnessModule.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness module: %w", err)
	}

	// 暂时跳过其他模块的启动
	// if cs.culturalWisdomModule != nil {
	// 	if err := cs.culturalWisdomModule.Start(); err != nil {
	// 		return fmt.Errorf("failed to start cultural wisdom module: %w", err)
	// 	}
	// }

	// if cs.aiIntegrationModule != nil {
	// 	if err := cs.aiIntegrationModule.Start(); err != nil {
	// 		return fmt.Errorf("failed to start AI integration module: %w", err)
	// 	}
	// }

	// if cs.communityModule != nil {
	// 	if err := cs.communityModule.Start(); err != nil {
	// 		return fmt.Errorf("failed to start community module: %w", err)
	// 	}
	// }

	// if cs.locationTrackingModule != nil {
	// 	if err := cs.locationTrackingModule.Start(); err != nil {
	// 		return fmt.Errorf("failed to start location tracking module: %w", err)
	// 	}
	// }

	// 启动HTTP服务
	go func() {
		cs.logger.Info("Starting HTTP server", zap.String("address", cs.httpServer.Addr))
		if err := cs.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cs.logger.Fatal("HTTP server failed to start", zap.Error(err))
		}
	}()

	cs.logger.Info("All core services started successfully")
	return nil
}

// Stop 停止所有服务模块
func (cs *CoreServices) Stop() error {
	cs.logger.Info("Stopping core services")

	// 停止HTTP服务
	if cs.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := cs.httpServer.Shutdown(ctx); err != nil {
			cs.logger.Error("HTTP server shutdown error", zap.Error(err))
		}
	}

	// 停止各个模块
	if cs.consciousnessModule != nil {
		if err := cs.consciousnessModule.Stop(); err != nil {
			cs.logger.Error("Failed to stop consciousness module", zap.Error(err))
		}
	}

	// 暂时跳过其他模块的停止
	// if cs.culturalWisdomModule != nil {
	// 	if err := cs.culturalWisdomModule.Stop(); err != nil {
	// 		cs.logger.Error("Failed to stop cultural wisdom module", zap.Error(err))
	// 	}
	// }

	// 关闭数据库连接
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

// healthCheck 简单健康检查
func (cs *CoreServices) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "taishang-core-services",
		"version":   "1.0.0",
	})
}

// detailedHealthCheck 详细健康检查
func (cs *CoreServices) detailedHealthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "taishang-core-services",
		"version":   "1.0.0",
		"modules":   gin.H{},
	}

	// 检查各个模块健康状态
	if cs.consciousnessModule != nil {
		health["modules"].(gin.H)["consciousness"] = cs.consciousnessModule.Health()
	}

	if cs.culturalWisdomModule != nil {
		health["modules"].(gin.H)["cultural_wisdom"] = cs.culturalWisdomModule.Health()
	}

	// 暂时跳过其他模块的健康检查
	// if cs.aiIntegrationModule != nil {
	// 	health["modules"].(gin.H)["ai_integration"] = cs.aiIntegrationModule.Health()
	// }

	// if cs.communityModule != nil {
	// 	health["modules"].(gin.H)["community"] = cs.communityModule.Health()
	// }

	// if cs.locationTrackingModule != nil {
	// 	health["modules"].(gin.H)["location_tracking"] = cs.locationTrackingModule.Health()
	// }

	// 检查数据库连接
	if cs.db != nil {
		if err := cs.db.Health(); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["database"] = "healthy"
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
	// 创建核心服务管理模块
	coreServices, err := NewCoreServices()
	if err != nil {
		log.Fatalf("Failed to create core services: %v", err)
	}

	// 初始化服务模块
	if err := coreServices.Initialize(); err != nil {
		log.Fatalf("Failed to initialize core services: %v", err)
	}

	// 设置HTTP服务
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
