package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/api/internal/handler"
	"github.com/codetaoist/taishanglaojun/api/internal/service"
	"github.com/codetaoist/taishanglaojun/api/internal/grpc"
	"github.com/codetaoist/taishanglaojun/api/internal/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/codetaoist/taishanglaojun/api/docs"
)

// @title Taishang Laojun AI API
// @version 1.0
// @description REST API for AI services including vector database and model operations
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	// Create a new Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Get service addresses from environment or use defaults
	vectorAddr := os.Getenv("VECTOR_SERVICE_ADDR")
	if vectorAddr == "" {
		vectorAddr = "localhost:50051"
	}

	modelAddr := os.Getenv("MODEL_SERVICE_ADDR")
	if modelAddr == "" {
		modelAddr = "localhost:50052"
	}

	// Initialize gRPC clients
	grpcClient, err := grpc.NewAIServiceClient(vectorAddr, modelAddr)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcClient.Close()

	// Initialize services
	aiService := service.NewAIService(grpcClient)

	// Initialize handlers
	aiHandler := handler.NewAIHandler(aiService)

	// Setup routes
	setupRoutes(router, aiHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

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

func setupRoutes(router *gin.Engine, aiHandler *handler.AIHandler) {
	// Health check endpoint
	router.GET("/health", aiHandler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication middleware for protected routes
		v1.Use(middleware.Auth())

		// Vector database operations
		vector := v1.Group("/vector")
		{
			vector.POST("/collections", aiHandler.CreateCollection)
			vector.DELETE("/collections/:name", aiHandler.DropCollection)
			vector.GET("/collections", aiHandler.ListCollections)
			vector.GET("/collections/:name", aiHandler.DescribeCollection)
			vector.POST("/collections/:name/load", aiHandler.LoadCollection)
			vector.POST("/collections/:name/release", aiHandler.ReleaseCollection)
			vector.GET("/collections/:name/stats", aiHandler.GetCollectionStatistics)
			
			vector.POST("/collections/:name/index", aiHandler.CreateIndex)
			vector.DELETE("/collections/:name/index/:indexName", aiHandler.DropIndex)
			vector.GET("/collections/:name/index", aiHandler.DescribeIndex)
			
			vector.POST("/collections/:name/search", aiHandler.Search)
			vector.POST("/collections/:name/insert", aiHandler.Insert)
			vector.DELETE("/collections/:name/data", aiHandler.Delete)
			vector.GET("/collections/:name/data", aiHandler.GetById)
			vector.POST("/collections/:name/compact", aiHandler.Compact)
		}

		// Model operations
		model := v1.Group("/model")
		{
			model.POST("/register", aiHandler.RegisterModel)
			model.PUT("/update", aiHandler.UpdateModel)
			model.DELETE("/unregister", aiHandler.UnregisterModel)
			model.GET("/list", aiHandler.ListModels)
			model.GET("/get", aiHandler.GetModel)
			model.POST("/load", aiHandler.LoadModel)
			model.POST("/unload", aiHandler.UnloadModel)
			model.GET("/status", aiHandler.GetModelStatus)
			
			model.POST("/generate/text", aiHandler.GenerateText)
			model.POST("/generate/embedding", aiHandler.GenerateEmbedding)
		}

		// Authentication endpoints
		auth := v1.Group("/auth")
		{
			auth.POST("/login", middleware.Login)
			auth.POST("/refresh", middleware.RefreshToken)
		}
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}