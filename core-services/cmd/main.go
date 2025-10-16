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
	"github.com/codetaoist/taishanglaojun/core-services/community"
	cultural_wisdom "github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom"
	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/database"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/internal/routes"
	"github.com/codetaoist/taishanglaojun/core-services/internal/services"
	location_tracking "github.com/codetaoist/taishanglaojun/core-services/location-tracking"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// 
	cfg, err := config.Load("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 
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

    // 使用配置文件中的数据库设置，避免硬编码导致认证错误
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
        ConnectTimeout:  30 * time.Second,
    }

	// 
	db, err := database.New(dbConfig, log)
	if err != nil {
		log.Fatal("Failed to initialize database", zap.Error(err))
	}
	// defer db.Close()

	// Redis
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

	// AI
	providerManager := providers.NewManager(log)

	// 
	mockProvider := providers.NewMockProvider(log)
	if err := providerManager.RegisterProvider("mock", mockProvider); err != nil {
		log.Error("Failed to register mock provider", zap.Error(err))
	} else {
		log.Info("Mock provider registered successfully")
		// 
		if err := providerManager.SetDefaultProvider("mock"); err != nil {
			log.Error("Failed to set mock as default provider", zap.Error(err))
		} else {
			log.Info("Mock provider set as default")
		}
	}

	// OpenAI
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
				// OpenAI
				if err := providerManager.SetDefaultProvider("openai"); err != nil {
					log.Error("Failed to set OpenAI as default provider", zap.Error(err))
				} else {
					log.Info("OpenAI set as default provider")
				}
			}
		}
	}

	// Azure
	if azureConfig, exists := cfg.AI.Providers["azure"]; exists && azureConfig.Enabled {
		if apiKey, ok := azureConfig.Config["api_key"].(string); ok && apiKey != "" {
			if endpoint, ok := azureConfig.Config["endpoint"].(string); ok && endpoint != "" {
				if deploymentName, ok := azureConfig.Config["deployment_name"].(string); ok && deploymentName != "" {
					apiVersion := "2023-05-15"
					if version, ok := azureConfig.Config["api_version"].(string); ok && version != "" {
						apiVersion = version
					}
					timeout := 30
					if t, ok := azureConfig.Config["timeout"].(int); ok {
						timeout = t
					}

					azureProviderConfig := providers.AzureConfig{
						APIKey:         apiKey,
						Endpoint:       endpoint,
						DeploymentName: deploymentName,
						APIVersion:     apiVersion,
						Timeout:        timeout,
					}
					azureProvider := providers.NewAzureProvider(azureProviderConfig, log)
					if err := providerManager.RegisterProvider("azure", azureProvider); err != nil {
						log.Error("Failed to register Azure provider", zap.Error(err))
					} else {
						log.Info("Azure provider registered successfully")
					}
				}
			}
		}
	}

	// 
	if baiduConfig, exists := cfg.AI.Providers["baidu"]; exists && baiduConfig.Enabled {
		if apiKey, ok := baiduConfig.Config["api_key"].(string); ok && apiKey != "" {
			if secretKey, ok := baiduConfig.Config["secret_key"].(string); ok && secretKey != "" {
				baseURL := "https://aip.baidubce.com"
				if url, ok := baiduConfig.Config["base_url"].(string); ok && url != "" {
					baseURL = url
				}
				timeout := 30
				if t, ok := baiduConfig.Config["timeout"].(int); ok {
					timeout = t
				}

				baiduProviderConfig := providers.BaiduConfig{
					APIKey:    apiKey,
					SecretKey: secretKey,
					BaseURL:   baseURL,
					Timeout:   timeout,
				}
				baiduProvider := providers.NewBaiduProvider(baiduProviderConfig, log)
				if err := providerManager.RegisterProvider("baidu", baiduProvider); err != nil {
					log.Error("Failed to register Baidu provider", zap.Error(err))
				} else {
					log.Info("Baidu provider registered successfully")
				}
			}
		}
	}

	// 数据库迁移 - 默认启用
	migrationService := database.NewMigrationService(db.GetDB(), log)
	if err := migrationService.RunMigration(); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}
	log.Info("Database migration completed successfully")

	// JWT
	expiration, err := time.ParseDuration(cfg.JWT.Expiration)
	if err != nil {
		expiration = 24 * time.Hour // 默认24小时
	}
	jwtConfig := middleware.JWTConfig{
		Secret:     cfg.JWT.Secret,
		Issuer:     cfg.JWT.Issuer,
		Expiration: expiration,
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig, log)

	// 
	authService := middleware.NewAuthService(db.GetDB(), jwtMiddleware, log)
	// 
	// if err := authService.AutoMigrate(); err != nil {
	// 	log.Fatal("Failed to migrate auth tables", zap.Error(err))
	// }

	// 
	authHandler := middleware.NewAuthHandler(authService, log, db.GetDB())

	// 
	appModuleService := services.NewAppModuleService(db.GetDB(), log)
	
	// 
	if err := appModuleService.InitializeDefaultModules(); err != nil {
		log.Error("Failed to initialize default modules", zap.Error(err))
	} else {
		log.Info("Default modules initialized successfully")
	}

	// 初始化用户服务
	userService := services.NewUserService(db.GetDB(), log)
	
	// 初始化菜单服务
	menuService := services.NewMenuService(db.GetDB(), log)



	// Gin
	gin.SetMode(cfg.Server.Mode)

	// 
	router := gin.New()

	// 
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "core-services",
		})
	})

	// API
	apiV1 := router.Group("/api/v1")

	// 
	middleware.SetupAuthRoutes(apiV1, authHandler, jwtMiddleware)

	// 设置管理员路由
	routes.SetupAdminRoutes(apiV1, authService, jwtMiddleware, db.GetDB(), log)

	// 设置仪表板路由
	routes.SetupDashboardRoutes(apiV1, authService, jwtMiddleware, userService, menuService, db.GetDB(), log)

	// 设置菜单路由
	routes.SetupMenuRoutes(apiV1, jwtMiddleware, db.GetDB(), log)

	// 
	var redisClientPtr *redis.Client
	if redisClient != nil {
		redisClientPtr = redisClient.GetClient()
	}

	// 设置权限路由
	routes.SetupPermissionRoutes(apiV1, jwtMiddleware, db.GetDB(), redisClientPtr, log)

	// 设置系统设置与数据库管理路由
	routes.SetupSystemRoutes(apiV1, jwtMiddleware, db.GetDB(), log)

	// 应用模块路由
	routes.SetupAppModuleRoutes(router, db.GetDB(), log, jwtMiddleware)

	// AI
	ai_integration.SetupRoutes(apiV1, db.GetDB(), log, providerManager)

	log.Info("=== Starting cultural wisdom routes setup ===")
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC in cultural wisdom routes setup", zap.Any("error", r))
		}
	}()

	// 
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

	// 
	log.Info("=== Starting location tracking routes setup ===")
	location_tracking.SetupRoutes(apiV1, db.GetDB(), log, jwtMiddleware)

	// 
	log.Info("=== Starting community routes setup ===")
	community.SetupRoutes(apiV1, db.GetDB(), log)
	log.Info("=== Community routes setup completed ===")

	log.Info("=== Cultural wisdom routes setup completed ===")

	// HTTP
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 
	go func() {
		log.Info("Server starting",
			zap.String("host", cfg.Server.Host),
			zap.Int("port", cfg.Server.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}

