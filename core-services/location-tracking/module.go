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

// Module 位置追踪模块
type Module struct {
	db           *gorm.DB
	redisClient  *redis.Client
	logger       *zap.Logger
	config       *ModuleConfig

	// 服务层
	locationService    services.LocationServiceInterface
	batchUploadService services.BatchUploadServiceInterface

	// 处理器层
	locationHandler handlers.LocationHandlerInterface

	// gRPC 服务器
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

// ModuleConfig 模块配置
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

// LocationServiceConfig 位置服务配置
type LocationServiceConfig struct {
	ServiceName           string
	Version              string
	Environment          string
	MaxPointsPerRequest  int
	MaxConcurrentReqs    int
	RequestTimeout       time.Duration
	MetricsRetention     time.Duration
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Enabled   bool
	Algorithm string
	Key       string
}

// StorageConfig 存储配置
type StorageConfig struct {
	RetentionDays   int
	CleanupInterval time.Duration
	MaxTrajectories int
	MaxPointsPerTrajectory int
}

// NewModule 创建新的位置追踪模块实例
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

	// 初始化服务层
	if err := module.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// 初始化处理器层
	if err := module.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	// 初始化gRPC服务器
	if config.GRPCEnabled {
		if err := module.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}

	return module, nil
}

// initServices 初始化服务层
func (m *Module) initServices() error {
	m.logger.Info("Initializing location tracking services")

	// 创建位置服务
	m.locationService = services.NewLocationService(m.db, m.logger)

	// 创建批量上传服务
	config := services.BatchUploadConfig{
		BatchSize:  100,
		Workers:    5,
		RetryLimit: 3,
		QueueSize:  1000,
	}
	m.batchUploadService = services.NewBatchUploadService(m.db, m.logger, config)

	return nil
}

// initHandlers 初始化处理器层
func (m *Module) initHandlers() error {
	m.logger.Info("Initializing location tracking handlers")

	// 创建处理器实例
	m.locationHandler = handlers.NewLocationHandler(m.locationService.(*services.LocationService), m.logger)

	return nil
}

// initGRPCServer 初始化gRPC服务器
func (m *Module) initGRPCServer() error {
	m.logger.Info("Initializing location tracking gRPC server")

	// 创建监听器
	address := fmt.Sprintf("%s:%d", m.config.GRPCHost, m.config.GRPCPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	m.grpcListener = listener

	// 创建gRPC服务器
	m.grpcServer = grpc.NewServer()

	// 注册服务
	// TODO: 注册位置追踪gRPC服务

	return nil
}

// SetupRoutes 设置HTTP路由
func (m *Module) SetupRoutes(router *gin.RouterGroup) error {
	if !m.config.HTTPEnabled {
		return nil
	}

	m.logger.Info("Setting up location tracking routes")

	// 创建JWT中间件
	jwtConfig := middleware.JWTConfig{
		Secret:     "taishang-secret-key", // 应该从配置中获取
		Issuer:     "taishang-core-services",
		Expiration: 24 * 60 * 60 * 1000000000, // 24小时，单位纳秒
	}
	jwtMiddleware := middleware.NewJWTMiddleware(jwtConfig, m.logger)

	// 位置追踪路由组
	locationGroup := router.Group(m.config.HTTPPrefix)
	{
		// 位置点相关路由
		points := locationGroup.Group("/points")
		points.Use(jwtMiddleware.AuthRequired())
		{
			points.POST("", m.locationHandler.UploadLocationPoints)      // 上传位置点
			points.POST("/batch", m.locationHandler.BatchUploadPoints)   // 批量上传位置点
			points.GET("", m.locationHandler.GetLocationPoints)         // 查询位置点
			points.DELETE("/:id", m.locationHandler.DeleteLocationPoint) // 删除位置点
		}

		// 轨迹相关路由
		trajectories := locationGroup.Group("/trajectories")
		trajectories.Use(jwtMiddleware.AuthRequired())
		{
			trajectories.POST("", m.locationHandler.CreateTrajectory)           // 创建轨迹
			trajectories.GET("", m.locationHandler.GetTrajectories)             // 获取轨迹列表
			trajectories.GET("/:id", m.locationHandler.GetTrajectory)           // 获取轨迹详情
			trajectories.PUT("/:id", m.locationHandler.UpdateTrajectory)        // 更新轨迹
			trajectories.DELETE("/:id", m.locationHandler.DeleteTrajectory)     // 删除轨迹
			trajectories.POST("/:id/finish", m.locationHandler.FinishTrajectory) // 结束轨迹
			trajectories.GET("/:id/points", m.locationHandler.GetTrajectoryPoints) // 获取轨迹点
			trajectories.GET("/:id/stats", m.locationHandler.GetTrajectoryStats)   // 获取轨迹统计
		}

		// 数据同步相关路由
		sync := locationGroup.Group("/sync")
		sync.Use(jwtMiddleware.AuthRequired())
		{
			sync.POST("", m.locationHandler.SyncData)              // 数据同步
			sync.GET("/status", m.locationHandler.GetSyncStatus)   // 同步状态查询
		}

		// 公开路由（不需要认证）
		locationGroup.GET("/health", m.locationHandler.HealthCheck) // 健康检查
	}

	m.logger.Info("Location tracking routes setup completed")
	return nil
}

// Start 启动模块
func (m *Module) Start() error {
	m.logger.Info("Starting location tracking module")

	// 自动迁移数据库
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 启动gRPC服务器
	if m.config.GRPCEnabled && m.grpcServer != nil {
		go func() {
			m.logger.Info("Starting location tracking gRPC server",
				zap.String("address", m.grpcListener.Addr().String()))
			if err := m.grpcServer.Serve(m.grpcListener); err != nil {
				m.logger.Error("gRPC server error", zap.Error(err))
			}
		}()
	}

	// 启动后台任务
	go m.startBackgroundTasks()

	m.logger.Info("Location tracking module started successfully")
	return nil
}

// Stop 停止模块
func (m *Module) Stop() error {
	m.logger.Info("Stopping location tracking module")

	// 停止gRPC服务器
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		if m.grpcListener != nil {
			m.grpcListener.Close()
		}
	}

	m.logger.Info("Location tracking module stopped successfully")
	return nil
}

// Health 健康检查
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

	// 检查数据库连接
	if sqlDB, err := m.db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["database"] = "healthy"
		}
	}

	// 检查Redis连接
	if err := m.redisClient.Ping(context.Background()).Err(); err != nil {
		health["redis"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["redis"] = "healthy"
	}

	return health
}

// migrateDatabase 迁移数据库
func (m *Module) migrateDatabase() error {
	m.logger.Info("Migrating location tracking database")

	// 自动迁移模型
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

// startBackgroundTasks 启动后台任务
func (m *Module) startBackgroundTasks() {
	m.logger.Info("Starting location tracking background tasks")

	// 定期清理过期数据
	go m.cleanupExpiredDataPeriodically()

	// 定期计算轨迹统计
	go m.calculateTrajectoryStatsPeriodically()
}

// cleanupExpiredDataPeriodically 定期清理过期数据
func (m *Module) cleanupExpiredDataPeriodically() {
	ticker := time.NewTicker(m.config.StorageConfig.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.locationService.CleanupExpiredData(m.config.StorageConfig.RetentionDays); err != nil {
			m.logger.Error("Failed to cleanup expired data", zap.Error(err))
		}
	}
}

// calculateTrajectoryStatsPeriodically 定期计算轨迹统计
func (m *Module) calculateTrajectoryStatsPeriodically() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		if err := m.locationService.UpdateTrajectoryStats(""); err != nil {
			m.logger.Error("Failed to update trajectory stats", zap.Error(err))
		}
	}
}

// getDefaultConfig 获取默认配置
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
			Key:       "", // 应该从环境变量获取
		},
		StorageConfig: &StorageConfig{
			RetentionDays:          365,
			CleanupInterval:        24 * time.Hour,
			MaxTrajectories:        1000,
			MaxPointsPerTrajectory: 10000,
		},
	}
}

// SetupRoutes 设置位置跟踪模块路由（向后兼容）
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, jwtMiddleware *middleware.JWTMiddleware) {
	// 自动迁移数据库表
	if err := db.AutoMigrate(&models.Trajectory{}, &models.LocationPoint{}); err != nil {
		logger.Error("Failed to migrate location tracking tables", zap.Error(err))
		return
	}

	// 创建服务实例
	locationService := services.NewLocationService(db, logger)

	// 创建处理器实例
	locationHandler := handlers.NewLocationHandler(locationService, logger)

	// 创建位置跟踪路由组，直接使用传入的apiV1路由组
	locationGroup := router.Group("/location-tracking")
	locationGroup.Use(jwtMiddleware.AuthRequired())

	// 设置基本路由
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