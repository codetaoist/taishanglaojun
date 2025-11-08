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
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/config"
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/handlers"
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/middleware"
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/router"
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/routes"
"github.com/codetaoist/taishanglaojun/hybrid-api/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()
	defer cfg.Close()

	// Initialize model manager
	modelManager := service.NewModelManager()
	defer modelManager.Close()

	// Convert sql.DB to gorm.DB for DAOs
	gormDB, err := dao.NewGormDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize GORM DB: %v", err)
	}

	// Initialize DAOs
	modelDAO := dao.NewModelDAO(gormDB)

	// Load model configurations from database and initialize services
	ctx := context.Background()
	modelConfigs, err := modelDAO.GetEnabledModelConfigs(ctx)
	if err != nil {
		log.Printf("Warning: Failed to load model configurations: %v", err)
	} else {
		for _, config := range modelConfigs {
			if _, err := modelManager.CreateService(config); err != nil {
				log.Printf("Warning: Failed to create model service %s: %v", config.Name, err)
			} else {
				log.Printf("Successfully initialized model service: %s", config.Name)
			}
		}
	}

	// Initialize plugin system service
	pluginSystemService := service.NewPluginSystemService(cfg)
	if err := pluginSystemService.Start(); err != nil {
		log.Printf("Warning: Failed to start plugin system service: %v", err)
	} else {
		log.Println("Plugin system service started successfully")
	}
	defer pluginSystemService.Stop()

	// Initialize handlers
	modelHandler := handlers.NewModelHandler(modelDAO, modelManager)

	// Set Gin mode
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	r := gin.New()

	// Add middleware
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORS(cfg))

	// Setup health check endpoint
	r.GET("/health", func(c *gin.Context) {
c.JSON(http.StatusOK, gin.H{
"status":    "ok",
"timestamp": time.Now().Unix(),
		})
	})

	// Setup API routes
	router.SetupLaojun(&cfg, r)
	router.SetupTaishangRoutes(&cfg, r)
	router.SetupTaishangModelRoutes(&cfg, r, modelManager)
	router.SetupProducts(&cfg, r)
	router.SetupVectorRoutes(&cfg, r) // 启用向量路由
	router.SetupAIRoutes(&cfg, r)    // 启用AI服务路由
	routes.SetupModelRoutes(r, modelHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
