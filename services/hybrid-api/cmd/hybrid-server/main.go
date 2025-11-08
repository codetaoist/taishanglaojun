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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/codetaoist/taishanglaojun/internal/ci"
	"github.com/codetaoist/taishanglaojun/internal/config"
	"github.com/codetaoist/taishanglaojun/internal/integration"
	"github.com/codetaoist/taishanglaojun/internal/middleware"
	"github.com/codetaoist/taishanglaojun/internal/model"
	"github.com/codetaoist/taishanglaojun/internal/plugin"
	"github.com/codetaoist/taishanglaojun/internal/service"
	"github.com/codetaoist/taishanglaojun/internal/service/handler"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize model manager
	modelManager := model.NewModelManager(cfg, logger)
	if err := modelManager.Start(); err != nil {
		logger.Fatal("Failed to start model manager", zap.Error(err))
	}

	// Initialize plugin manager
	pluginConfig := &plugin.PluginConfig{
		Directory:     cfg.Plugin.Directory,
		AutoLoad:      cfg.Plugin.AutoLoad,
		HealthCheck:   cfg.Plugin.HealthCheck,
		CheckInterval: cfg.Plugin.CheckInterval,
	}
	pluginManager := plugin.NewPluginManager(pluginConfig, logger)
	if err := pluginManager.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start plugin manager", zap.Error(err))
	}

	// Initialize CI/CD pipeline
	ciConfig := &ci.CIConfig{
		Enabled: cfg.CI.Enabled,
		Build: ci.BuildConfig{
			BuildDir:     cfg.CI.BuildDir,
			ArtifactDir:  cfg.CI.ArtifactDir,
			Dockerfile:   cfg.CI.Dockerfile,
			ImageName:    cfg.CI.ImageName,
			ImageTag:     cfg.CI.ImageTag,
			Registry:     cfg.CI.RegistryURL,
			BuildArgs:    cfg.CI.BuildArgs,
			Environment:  cfg.CI.Environment,
			Timeout:      cfg.CI.Timeout,
			CacheEnabled: cfg.CI.CacheEnabled,
			CacheDir:     cfg.CI.CacheDir,
		},
		Test: ci.TestConfig{
			TestDir:      cfg.CI.TestDir,
			TestPattern:  cfg.CI.TestPattern,
			Coverage:     cfg.CI.Coverage,
			CoverageFile: cfg.CI.CoverageFile,
			Environment:  cfg.CI.TestEnvironment,
			Timeout:      cfg.CI.TestTimeout,
			Parallel:     cfg.CI.Parallel,
			Verbose:      cfg.CI.Verbose,
		},
		Deploy: ci.DeployConfig{
			Environment:  cfg.CI.DeployEnvironment,
			Namespace:    cfg.CI.DeployNamespace,
			KubeConfig:   cfg.CI.KubeConfig,
			Manifests:    cfg.CI.Manifests,
			HelmChart:    cfg.CI.HelmChart,
			HelmValues:   cfg.CI.HelmValues,
			HelmRelease:  cfg.CI.HelmRelease,
			Wait:         cfg.CI.DeployWait,
			Timeout:      cfg.CI.DeployTimeout,
			Variables:    cfg.CI.Variables,
		},
	}
	ciPipeline := ci.NewCIPipeline(ciConfig, logger)

	// Initialize plugin-CI integration
	integrationConfig := &integration.IntegrationConfig{
		AutoBuildPlugins:        cfg.Plugin.AutoBuild,
		BuildTimeout:           int(cfg.Plugin.Timeout.Seconds()),
		ArtifactRetention:      int(cfg.CI.ArtifactRetention.Hours() / 24),
		EnableVersioning:       cfg.Plugin.Versioning,
		DefaultPipelineTemplate: cfg.CI.DefaultPipelineTemplate,
		RegistryURL:           cfg.CI.RegistryURL,
	}
	pluginCIIntegration := integration.NewPluginCIIntegration(pluginManager, ciPipeline, logger, integrationConfig)
	if err := pluginCIIntegration.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start plugin-CI integration", zap.Error(err))
	}
	defer pluginCIIntegration.Stop(context.Background())

	// Initialize service manager
	serviceManager := service.NewHybridAIServiceManager(cfg, logger)
	
	// Register built-in services
	if err := registerBuiltinServices(serviceManager, cfg, logger, pluginManager, ciPipeline); err != nil {
		logger.Fatal("Failed to register built-in services", zap.Error(err))
	}
	
	// Start service manager
	if err := serviceManager.Start(); err != nil {
		logger.Fatal("Failed to start service manager", zap.Error(err))
	}

	// Initialize handlers
	hybridHandler := handler.NewHybridServiceHandler(serviceManager, logger)
	serviceDiscoveryHandler := handler.NewServiceDiscoveryHandler(serviceManager, logger)
	modelHandler := model.NewModelHandler(modelManager, logger)
	pluginHandler := plugin.NewPluginAPIHandler(pluginManager, logger)
	ciHandler := ci.NewCIAPIHandler(ciPipeline, logger)
	integrationHandler := integration.NewPluginCIAPIHandler(pluginCIIntegration, logger)

	// Setup routes
	router := gin.New()
	
	// Add global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Tracing())
	router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Logger: logger,
		UTC:    true,
	}))
	router.Use(middleware.Metrics())
	router.Use(middleware.RateLimit(cfg.Hybrid.RateLimit.RequestsPerMinute))
	router.Use(middleware.CircuitBreaker())
	router.Use(middleware.Timeout(time.Duration(cfg.Hybrid.CircuitBreaker.Timeout) * time.Second))
	router.Use(middleware.CORS())
	
	// Setup API routes
	api := router.Group("/api/v1")
	hybridRouter := handler.NewHybridAPIRouter(hybridHandler)
	hybridRouter.SetupRoutes(api)
	
	serviceDiscoveryRouter := handler.NewServiceDiscoveryRouter(serviceDiscoveryHandler)
	serviceDiscoveryRouter.SetupRoutes(api)
	
	modelRouter := model.NewModelRouter(modelHandler)
	modelRouter.SetupRoutes(api)
	
	pluginHandler.SetupRoutes(api)
	ciHandler.SetupRoutes(api)
	integrationHandler.SetupRoutes(api)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown service manager
	if err := serviceManager.Stop(); err != nil {
		logger.Error("Failed to stop service manager", zap.Error(err))
	}

	// Shutdown model manager
	if err := modelManager.Stop(); err != nil {
		logger.Error("Failed to stop model manager", zap.Error(err))
	}

	// Shutdown plugin manager
	if err := pluginManager.Stop(context.Background()); err != nil {
		logger.Error("Failed to stop plugin manager", zap.Error(err))
	}

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// registerBuiltinServices registers built-in services
func registerBuiltinServices(serviceManager *service.HybridAIServiceManager, cfg *config.HybridConfig, logger *zap.Logger, pluginManager *plugin.PluginManager, ciPipeline *ci.CIPipeline) error {
	// Register vector service
	vectorService := &service.ServiceInfo{
		ID:          "vector-service",
		Name:        "Vector Database Service",
		Type:        "vector",
		Address:     cfg.AI.VectorService.Address,
		Port:        cfg.AI.VectorService.Port,
		HealthCheck: "/health",
		Metadata: map[string]string{
			"provider": "codetaoist",
			"version":  "1.0.0",
		},
	}
	
	if err := serviceManager.RegisterService(context.Background(), vectorService); err != nil {
		return fmt.Errorf("failed to register vector service: %v", err)
	}
	
	// Register model service
	modelService := &service.ServiceInfo{
		ID:          "model-service",
		Name:        "Model Service",
		Type:        "model",
		Address:     cfg.AI.ModelService.Address,
		Port:        cfg.AI.ModelService.Port,
		HealthCheck: "/health",
		Metadata: map[string]string{
			"provider": "codetaoist",
			"version":  "1.0.0",
		},
	}
	
	if err := serviceManager.RegisterService(context.Background(), modelService); err != nil {
		return fmt.Errorf("failed to register model service: %v", err)
	}
	
	// Register plugin service
	pluginService := &service.ServiceInfo{
		ID:          "plugin-service",
		Name:        "Plugin Service",
		Type:        "plugin",
		HealthCheck: "/health",
		Metadata: map[string]string{
			"provider": "codetaoist",
			"version":  "1.0.0",
		},
	}
	
	if err := serviceManager.RegisterService(context.Background(), pluginService); err != nil {
		return fmt.Errorf("failed to register plugin service: %v", err)
	}
	
	// Register CI/CD service
	ciService := &service.ServiceInfo{
		ID:          "ci-service",
		Name:        "CI/CD Service",
		Type:        "ci",
		HealthCheck: "/health",
		Metadata: map[string]string{
			"provider": "codetaoist",
			"version":  "1.0.0",
		},
	}
	
	if err := serviceManager.RegisterService(context.Background(), ciService); err != nil {
		return fmt.Errorf("failed to register CI/CD service: %v", err)
	}
	
	logger.Info("Built-in services registered successfully")
	return nil
}