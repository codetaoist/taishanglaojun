package location_tracking

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/models"
	"github.com/codetaoist/taishanglaojun/core-services/location-tracking/services"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

// Module 
type Module struct {
	db           *gorm.DB
	redisClient  *redis.Client
	logger       *zap.Logger
	config       *ModuleConfig

	// ?
	locationService    services.LocationServiceInterface
	batchUploadService services.BatchUploadServiceInterface

	// 
	locationHandler handlers.LocationHandlerInterface

	// gRPC ?
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

// ModuleConfig 
type ModuleConfig struct {
	HTTPEnabled bool
	HTTPPrefix  string
	GRPCEnabled bool
	GRPCPort    int
	GRPCHost    string

	ServiceConfig   *LocationServiceConfig
	EncryptionConfig *EncryptionConfig
	StorageConfig   *StorageConfig
}

// LocationServiceConfig 
type LocationServiceConfig struct {
	ServiceName           string
	Version              string
	Environment          string
	MaxPointsPerRequest  int
	MaxConcurrentReqs    int
	RequestTimeout       time.Duration
	MetricsRetention     time.Duration
}

// EncryptionConfig 
type EncryptionConfig struct {
	Enabled   bool
	Algorithm string
	Key       string
}

// StorageConfig 洢
type StorageConfig struct {
	RetentionDays   int
	CleanupInterval time.Duration
	MaxTrajectories int
	MaxPointsPerTrajectory int
}

// NewModule 
func NewModule(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger, config *ModuleConfig) (*Module, error) {
	if config == nil {
		config = getDefaultConfig()
	}

	module := &Module{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		config:      config,
	}

	// 
	if err := module.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// ?
	if err := module.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	// gRPC?
	if config.GRPCEnabled {
		if err := module.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}

	return module, nil
}

// initServices 
func (m *Module) initServices() error {
	m.logger.Info("Initializing location tracking services")

	// 
	m.locationService = services.NewLocationService(m.db, m.logger)

	// 
	config := services.BatchUploadConfig{
		BatchSize:  100,
		Workers:    5,
		RetryLimit: 3,
		QueueSize:  1000,
	}
	m.batchUploadService = services.NewBatchUploadService(m.db, m.logger, config)

	return nil
}

// initHandlers ?
func (m *Module) initHandlers() error {
	m.logger.Info("Initializing location tracking handlers")

	// ?
	m.locationHandler = handlers.NewLocationHandler(m.locationService.(*services.LocationService), m.logger)

	return nil
}

// initGRPCServer gRPC?
func (m *Module) initGRPCServer() error {
	m.logger.Info("Initializing location tracking gRPC server")

	// ?
	address := fmt.Sprintf("%s:%d", m.config.GRPCHost, m.config.GRPCPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	m.grpcListener = listener

	// gRPC?
	m.grpcServer = grpc.NewServer()

	// 
	// TODO: gRPC

	return nil
}

// SetupRoutes HTTP
func (m *Module) SetupRoutes(router *gin.RouterGroup) error {
	if !m.config.HTTPEnabled {
		return nil
	}

	m.logger.Info("Setting up location tracking routes")

	// JWT?
	jwtConfig := middleware.JWTConfig{
		Secret:     "taishang-secret-key", // 
		Issuer:     "taishang-core-services",
		Expiration: 24 * 60 * 60 * 1000000000, // 24?
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig, m.logger)

	// ?
	locationGroup := router.Group(m.config.HTTPPrefix)
	{
		// ?
		points := locationGroup.Group("/points")
		points.Use(jwtMiddleware.AuthRequired())
		{
			points.POST("", m.locationHandler.UploadLocationPoints)      // ?
			points.POST("/batch", m.locationHandler.BatchUploadPoints)   // ?
			points.GET("", m.locationHandler.GetLocationPoints)         // ?
			points.DELETE("/:id", m.locationHandler.DeleteLocationPoint) // ?
		}

		// 
		trajectories := locationGroup.Group("/trajectories")
		trajectories.Use(jwtMiddleware.AuthRequired())
		{
			trajectories.POST("", m.locationHandler.CreateTrajectory)           // 
			trajectories.GET("", m.locationHandler.GetTrajectories)             // 
			trajectories.GET("/:id", m.locationHandler.GetTrajectory)           // 
			trajectories.PUT("/:id", m.locationHandler.UpdateTrajectory)        // 
			trajectories.DELETE("/:id", m.locationHandler.DeleteTrajectory)     // 
			trajectories.POST("/:id/finish", m.locationHandler.FinishTrajectory) // 
			trajectories.GET("/:id/points", m.locationHandler.GetTrajectoryPoints) // ?
			trajectories.GET("/:id/stats", m.locationHandler.GetTrajectoryStats)   // 
		}

		// 
		sync := locationGroup.Group("/sync")
		sync.Use(jwtMiddleware.AuthRequired())
		{
			sync.POST("", m.locationHandler.SyncData)              // 
			sync.GET("/status", m.locationHandler.GetSyncStatus)   // ?
		}

		// 
		locationGroup.GET("/health", m.locationHandler.HealthCheck) // ?
	}

	m.logger.Info("Location tracking routes setup completed")
	return nil
}

// Start 
func (m *Module) Start() error {
	m.logger.Info("Starting location tracking module")

	// ?
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// gRPC?
	if m.config.GRPCEnabled && m.grpcServer != nil {
		go func() {
			m.logger.Info("Starting location tracking gRPC server",
				zap.String("address", m.grpcListener.Addr().String()))
			if err := m.grpcServer.Serve(m.grpcListener); err != nil {
				m.logger.Error("gRPC server error", zap.Error(err))
			}
		}()
	}

	// 
	go m.startBackgroundTasks()

	m.logger.Info("Location tracking module started successfully")
	return nil
}

// Stop 
func (m *Module) Stop() error {
	m.logger.Info("Stopping location tracking module")

	// gRPC?
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		if m.grpcListener != nil {
			m.grpcListener.Close()
		}
	}

	m.logger.Info("Location tracking module stopped successfully")
	return nil
}

// Health ?
func (m *Module) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status":  "healthy",
		"module":  "location-tracking",
		"version": m.config.ServiceConfig.Version,
		"services": map[string]string{
			"location_service":     "running",
			"batch_upload_service": "running",
		},
	}

	// 
	if sqlDB, err := m.db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["database"] = "healthy"
		}
	}

	// Redis
	if err := m.redisClient.Ping(context.Background()).Err(); err != nil {
		health["redis"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["redis"] = "healthy"
	}

	return health
}

// migrateDatabase ?
func (m *Module) migrateDatabase() error {
	m.logger.Info("Migrating location tracking database")

	// 
	err := m.db.AutoMigrate(
		&models.Trajectory{},
		&models.LocationPoint{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	m.logger.Info("Location tracking database migration completed")
	return nil
}

// startBackgroundTasks 
func (m *Module) startBackgroundTasks() {
	m.logger.Info("Starting location tracking background tasks")

	// 
	go m.cleanupExpiredDataPeriodically()

	// 
	go m.calculateTrajectoryStatsPeriodically()
}

// cleanupExpiredDataPeriodically 
func (m *Module) cleanupExpiredDataPeriodically() {
	ticker := time.NewTicker(m.config.StorageConfig.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.locationService.CleanupExpiredData(m.config.StorageConfig.RetentionDays); err != nil {
			m.logger.Error("Failed to cleanup expired data", zap.Error(err))
		}
	}
}

// calculateTrajectoryStatsPeriodically 
func (m *Module) calculateTrajectoryStatsPeriodically() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.locationService.UpdateTrajectoryStats(""); err != nil {
			m.logger.Error("Failed to update trajectory stats", zap.Error(err))
		}
	}
}

// getDefaultConfig 
func getDefaultConfig() *ModuleConfig {
	return &ModuleConfig{
		HTTPEnabled: true,
		HTTPPrefix:  "/location-tracking",
		GRPCEnabled: false,
		GRPCPort:    50055,
		GRPCHost:    "localhost",
		ServiceConfig: &LocationServiceConfig{
			ServiceName:          "location-tracking-service",
			Version:             "1.0.0",
			Environment:         "development",
			MaxPointsPerRequest: 1000,
			MaxConcurrentReqs:   100,
			RequestTimeout:      30 * time.Second,
			MetricsRetention:    24 * time.Hour,
		},
		EncryptionConfig: &EncryptionConfig{
			Enabled:   true,
			Algorithm: "AES-256-GCM",
			Key:       "", // ?
		},
		StorageConfig: &StorageConfig{
			RetentionDays:          365,
			CleanupInterval:        24 * time.Hour,
			MaxTrajectories:        1000,
			MaxPointsPerTrajectory: 10000,
		},
	}
}

// SetupRoutes 
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, jwtMiddleware *middleware.JWTMiddleware) {
	// 
	if err := db.AutoMigrate(&models.Trajectory{}, &models.LocationPoint{}); err != nil {
		logger.Error("Failed to migrate location tracking tables", zap.Error(err))
		return
	}

	// 
	locationService := services.NewLocationService(db, logger)

	// ?
	locationHandler := handlers.NewLocationHandler(locationService, logger)

	// apiV1?
	locationGroup := router.Group("/location-tracking")
	locationGroup.Use(jwtMiddleware.AuthRequired())

	// 
	locationGroup.POST("/points", locationHandler.UploadLocationPoints)
	locationGroup.GET("/points", locationHandler.GetLocationPoints)
	locationGroup.DELETE("/points/:id", locationHandler.DeleteLocationPoint)

	locationGroup.POST("/trajectories", locationHandler.CreateTrajectory)
	locationGroup.GET("/trajectories", locationHandler.GetTrajectories)
	locationGroup.GET("/trajectories/:id", locationHandler.GetTrajectory)
	locationGroup.PUT("/trajectories/:id", locationHandler.UpdateTrajectory)
	locationGroup.DELETE("/trajectories/:id", locationHandler.DeleteTrajectory)

	logger.Info("Location tracking routes setup completed")
}

