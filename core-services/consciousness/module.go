package consciousness

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/coordinators"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	// grpcserver "github.com/codetaoist/taishanglaojun/core-services/consciousness/grpc" // 暂时注释掉
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/handlers"

	// pb "github.com/codetaoist/taishanglaojun/core-services/consciousness/proto" // 暂时注释掉
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/services"
)

// Module -consciousness模块
type Module struct {
	//
	consciousnessService *services.ConsciousnessService

	// HTTP
	fusionHandler       *handlers.FusionHandler
	evolutionHandler    *handlers.EvolutionHandler
	geneHandler         *handlers.QuantumGeneHandler
	coordinationHandler *handlers.CoordinationHandler

	// gRPC - 暂时注释掉
	// grpcServer              *grpc.Server
	// consciousnessGRPCServer *grpcserver.ConsciousnessServer

	//
	config *ModuleConfig
	db     *gorm.DB
	logger *zap.Logger

	//
	httpStarted bool
	grpcStarted bool
}

// ModuleConfig -consciousness模块配置
type ModuleConfig struct {
	// HTTP
	HTTPEnabled bool   `json:"http_enabled"`
	HTTPPrefix  string `json:"http_prefix"`

	// gRPC
	GRPCEnabled bool   `json:"grpc_enabled"`
	GRPCPort    int    `json:"grpc_port"`
	GRPCHost    string `json:"grpc_host"`

	//
	ServiceConfig *services.ConsciousnessConfig `json:"service_config"`

	//
	FusionConfig       *engines.FusionEngineConfig              `json:"fusion_config"`
	EvolutionConfig    *engines.EvolutionTrackerConfig          `json:"evolution_config"`
	GeneConfig         *engines.QuantumGeneManagerConfig        `json:"gene_config"`
	CoordinationConfig *coordinators.ThreeAxisCoordinatorConfig `json:"coordination_config"`
}

// NewModule -consciousness模块
func NewModule(config *ModuleConfig, db *gorm.DB, logger *zap.Logger) (*Module, error) {
	if config == nil {
		config = getDefaultConfig()
	}

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	//
	if config.ServiceConfig == nil {
		config.ServiceConfig = &services.ConsciousnessConfig{
			ServiceName:        "consciousness-service",
			Version:            "1.0.0",
			Environment:        "development",
			UpdateInterval:     30 * time.Second,
			FusionConfig:       config.FusionConfig,
			EvolutionConfig:    config.EvolutionConfig,
			GeneConfig:         config.GeneConfig,
			CoordinationConfig: config.CoordinationConfig,
		}
	}

	//
	consciousnessService, err := services.NewConsciousnessService(config.ServiceConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create consciousness service: %w", err)
	}

	module := &Module{
		consciousnessService: consciousnessService,
		config:               config,
		db:                   db,
		logger:               logger,
	}

	// HTTP
	if config.HTTPEnabled {
		if err := module.initHTTPHandlers(); err != nil {
			return nil, fmt.Errorf("failed to initialize HTTP handlers: %w", err)
		}
	}

	// gRPC - 暂时注释掉
	// if config.GRPCEnabled {
	// 	if err := module.initGRPCServer(); err != nil {
	// 		return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
	// 	}
	// }

	return module, nil
}

// initHTTPHandlers HTTP -consciousness模块
func (m *Module) initHTTPHandlers() error {
	m.fusionHandler = handlers.NewFusionHandler(m.consciousnessService.GetFusionEngine(), m.logger)
	m.evolutionHandler = handlers.NewEvolutionHandler(m.consciousnessService.GetEvolutionTracker(), m.logger)
	m.geneHandler = handlers.NewQuantumGeneHandler(m.consciousnessService.GetGeneManager(), m.logger)
	m.coordinationHandler = handlers.NewCoordinationHandler(m.consciousnessService.GetCoordinator(), m.logger)

	return nil
}

// initGRPCServer gRPC - 暂时注释掉
// func (m *Module) initGRPCServer() error {
// 	m.grpcServer = grpc.NewServer()
// 	m.consciousnessGRPCServer = grpcserver.NewConsciousnessServer(m.consciousnessService, m.logger)

// 	// gRPC - 暂时注释掉，等待protobuf文件更新
// 	// pb.RegisterConsciousnessServiceServer(m.grpcServer, m.consciousnessGRPCServer)

// 	return nil
// }

// SetupRoutes HTTP
func (m *Module) SetupRoutes(router *gin.RouterGroup, jwtMiddleware gin.HandlerFunc) error {
	if !m.config.HTTPEnabled {
		m.logger.Info("HTTP routes disabled for consciousness service")
		return nil
	}

	//
	prefix := m.config.HTTPPrefix
	if prefix == "" {
		prefix = "/consciousness"
	}

	consciousnessGroup := router.Group(prefix)

	// JWT
	if jwtMiddleware != nil {
		consciousnessGroup.Use(jwtMiddleware)
	}

	//
	consciousnessGroup.GET("/health", m.healthHandler)
	consciousnessGroup.GET("/stats", m.statsHandler)

	//
	fusionGroup := consciousnessGroup.Group("/fusion")
	{
		fusionGroup.POST("/start", m.fusionHandler.StartFusionSession)
		fusionGroup.GET("/status/:sessionId", m.fusionHandler.GetFusionSession)
		fusionGroup.DELETE("/cancel/:sessionId", m.fusionHandler.CancelFusionSession)
		fusionGroup.GET("/history/:entityId", m.fusionHandler.GetFusionHistory)
		fusionGroup.GET("/strategies", m.fusionHandler.GetFusionStrategies)
		fusionGroup.GET("/metrics/:entityId", m.fusionHandler.GetFusionMetrics)
	}

	//
	evolutionGroup := consciousnessGroup.Group("/evolution")
	{
		evolutionGroup.GET("/state/:entityId", m.evolutionHandler.GetEvolutionState)
		evolutionGroup.PUT("/state/:entityId", m.evolutionHandler.UpdateEvolutionState)
		evolutionGroup.POST("/track", m.evolutionHandler.TrackEvolution)
		evolutionGroup.GET("/prediction/:entityId", m.evolutionHandler.GetEvolutionPrediction)
		evolutionGroup.GET("/path/:entityId", m.evolutionHandler.GetEvolutionPath)
		evolutionGroup.GET("/milestones/:entityId", m.evolutionHandler.GetEvolutionMilestones)
		evolutionGroup.GET("/sequence/:level", m.evolutionHandler.GetSequenceLevels)
		evolutionGroup.GET("/stats/:entityId", m.evolutionHandler.GetEvolutionStats)
	}

	//
	geneGroup := consciousnessGroup.Group("/genes")
	{
		geneGroup.POST("/pool", m.geneHandler.CreateGenePool)
		geneGroup.GET("/pool/:entityId", m.geneHandler.GetGenePool)
		geneGroup.POST("/add", m.geneHandler.AddGene)
		geneGroup.POST("/express", m.geneHandler.ExpressGene)
		geneGroup.POST("/mutate", m.geneHandler.MutateGene)
		geneGroup.POST("/interactions", m.geneHandler.AnalyzeInteractions)
		geneGroup.POST("/simulate", m.geneHandler.SimulateEvolution)
		geneGroup.GET("/stats/:entityId", m.geneHandler.GetGeneStats)
		geneGroup.GET("/types", m.geneHandler.GetGeneTypes)
		geneGroup.GET("/search/:entityId", m.geneHandler.SearchGenes)
	}

	//
	coordinationGroup := consciousnessGroup.Group("/coordination")
	{
		coordinationGroup.POST("/start", m.coordinationHandler.StartCoordination)
		coordinationGroup.GET("/status/:sessionId", m.coordinationHandler.GetCoordinationSession)
		coordinationGroup.DELETE("/stop/:sessionId", m.coordinationHandler.StopCoordination)
		coordinationGroup.POST("/s-axis", m.coordinationHandler.ProcessSAxis)
		coordinationGroup.POST("/c-axis", m.coordinationHandler.ProcessCAxis)
		coordinationGroup.POST("/t-axis", m.coordinationHandler.ProcessTAxis)
		coordinationGroup.POST("/balance", m.coordinationHandler.OptimizeBalance)
		coordinationGroup.POST("/synergy", m.coordinationHandler.CatalyzeSynergy)
		coordinationGroup.GET("/history/:entityId", m.coordinationHandler.GetCoordinationHistory)
		coordinationGroup.GET("/axis/:axisType", m.coordinationHandler.GetAxisInfo)
		coordinationGroup.GET("/metrics/:entityId", m.coordinationHandler.GetCoordinationMetrics)
	}

	m.httpStarted = true
	m.logger.Info("Consciousness service HTTP routes registered",
		zap.String("prefix", prefix),
		zap.Bool("jwt_enabled", jwtMiddleware != nil))

	return nil
}

// StartGRPCServer gRPC - 暂时注释掉
// func (m *Module) StartGRPCServer() error {
// 	if !m.config.GRPCEnabled {
// 		m.logger.Info("gRPC server disabled for consciousness service")
// 		return nil
// 	}

// 	if m.grpcStarted {
// 		return fmt.Errorf("gRPC server is already running")
// 	}

// 	host := m.config.GRPCHost
// 	if host == "" {
// 		host = "localhost"
// 	}

// 	port := m.config.GRPCPort
// 	if port == 0 {
// 		port = 50051
// 	}

// 	address := fmt.Sprintf("%s:%d", host, port)
// 	listener, err := net.Listen("tcp", address)
// 	if err != nil {
// 		return fmt.Errorf("failed to listen on %s: %w", address, err)
// 	}

// 	// gRPC
// 	go func() {
// 		m.logger.Info("Starting consciousness gRPC server", zap.String("address", address))
// 		if err := m.grpcServer.Serve(listener); err != nil {
// 			m.logger.Error("gRPC server failed", zap.Error(err))
// 		}
// 	}()

// 	m.grpcStarted = true
// 	m.logger.Info("Consciousness gRPC server started", zap.String("address", address))

// 	return nil
// }

// Start
func (m *Module) Start() error {
	//
	if err := m.consciousnessService.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness service: %w", err)
	}

	// gRPC - 暂时注释掉
	// if m.config.GRPCEnabled {
	// 	if err := m.StartGRPCServer(); err != nil {
	// 		return fmt.Errorf("failed to start gRPC server: %w", err)
	// 	}
	// }

	m.logger.Info("Consciousness module started successfully")
	return nil
}

// Stop
func (m *Module) Stop() error {
	// gRPC - 暂时注释掉
	// if m.grpcStarted && m.grpcServer != nil {
	// 	m.grpcServer.GracefulStop()
	// 	m.grpcStarted = false
	// 	m.logger.Info("Consciousness gRPC server stopped")
	// }

	//
	if err := m.consciousnessService.Stop(); err != nil {
		m.logger.Error("Failed to stop consciousness service", zap.Error(err))
		return err
	}

	m.logger.Info("Consciousness module stopped successfully")
	return nil
}

// Health
func (m *Module) Health() map[string]interface{} {
	health := m.consciousnessService.Health()
	health["module"] = map[string]interface{}{
		"http_enabled": m.config.HTTPEnabled,
		"http_started": m.httpStarted,
		// "grpc_enabled": m.config.GRPCEnabled,
		// "grpc_started": m.grpcStarted,
	}
	return health
}

// healthHandler HTTP鴦
func (m *Module) healthHandler(c *gin.Context) {
	health := m.Health()
	c.JSON(200, health)
}

// statsHandler HTTP
func (m *Module) statsHandler(c *gin.Context) {
	stats := m.consciousnessService.GetStats()
	c.JSON(200, stats)
}

// GetService
func (m *Module) GetService() *services.ConsciousnessService {
	return m.consciousnessService
}

// GetConfig
func (m *Module) GetConfig() *ModuleConfig {
	return m.config
}

// getDefaultConfig
func getDefaultConfig() *ModuleConfig {
	return &ModuleConfig{
		HTTPEnabled: true,
		HTTPPrefix:  "/consciousness",
		GRPCEnabled: true,
		GRPCPort:    50051,
		GRPCHost:    "localhost",
		ServiceConfig: &services.ConsciousnessConfig{
			ServiceName:           "consciousness-service",
			Version:               "1.0.0",
			Environment:           "development",
			UpdateInterval:        30 * time.Second,
			MaxConcurrentSessions: 100,
			SessionTimeout:        30 * time.Minute,
			MetricsRetention:      24 * time.Hour,
			EnableAuthentication:  true,
			AllowedOrigins:        []string{"*"},
			RateLimitRPS:          100,
		},
		FusionConfig: &engines.FusionEngineConfig{
			MaxConcurrentSessions:    50,
			SessionTimeout:           5 * time.Minute,
			QualityThreshold:         0.7,
			SynergyThreshold:         0.6,
			EnableEmergenceDetection: true,
			CarbonWeight:             0.5,
			SiliconWeight:            0.5,
			FusionStrategies:         []string{"complementary", "synergistic"},
		},
		EvolutionConfig: &engines.EvolutionTrackerConfig{
			UpdateInterval:           30 * time.Second,
			MaxConcurrentTracking:    20,
			PredictionHorizon:        24 * time.Hour,
			MetricsRetentionPeriod:   7 * 24 * time.Hour,
			EnableRealTimeTracking:   true,
			EnablePredictiveAnalysis: true,
			EnablePathOptimization:   true,
			MinConfidenceThreshold:   0.6,
			MaxEvolutionDuration:     30 * 24 * time.Hour,
		},
		GeneConfig: &engines.QuantumGeneManagerConfig{
			MaxGenePools:                10,
			MaxGenesPerPool:             1000,
			MutationRate:                0.01,
			ExpressionUpdateInterval:    1 * time.Minute,
			EvolutionSimulationInterval: 5 * time.Minute,
			EnableAutoMutation:          true,
			EnableExpressionControl:     true,
			EnableInteractionAnalysis:   true,
			EnableEvolutionSimulation:   true,
			GeneStabilityThreshold:      0.8,
			ExpressionThreshold:         0.5,
			MutationSeverityLimit:       "moderate",
		},
		CoordinationConfig: &coordinators.ThreeAxisCoordinatorConfig{
			MaxConcurrentCoordinations: 30,
			CoordinationTimeout:        5 * time.Minute,
			BalanceThreshold:           0.8,
			SynergyThreshold:           0.7,
			OptimizationInterval:       1 * time.Minute,
			EnableAutoBalance:          true,
			EnableSynergyCatalysis:     true,
			EnableHistoryTracking:      true,
			MaxHistoryRecords:          1000,
			QualityThreshold:           0.8,
			ConvergenceThreshold:       0.95,
			MaxIterations:              100,
		},
	}
}

// SetupRoutes
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, jwtMiddleware gin.HandlerFunc) error {
	//
	module, err := NewModule(nil, db, logger)
	if err != nil {
		return fmt.Errorf("failed to create consciousness module: %w", err)
	}

	//
	if err := module.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness module: %w", err)
	}

	//
	if err := module.SetupRoutes(router, jwtMiddleware); err != nil {
		return fmt.Errorf("failed to setup consciousness routes: %w", err)
	}

	logger.Info("Consciousness service routes setup completed")
	return nil
}
