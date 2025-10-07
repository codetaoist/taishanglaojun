package consciousness

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/coordinators"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	grpcserver "github.com/codetaoist/taishanglaojun/core-services/consciousness/grpc"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/handlers"
	pb "github.com/codetaoist/taishanglaojun/core-services/consciousness/proto"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/services"
)

// Module 意识服务模块
type Module struct {
	// 核心服务
	consciousnessService *services.ConsciousnessService
	
	// HTTP处理�?	fusionHandler      *handlers.FusionHandler
	evolutionHandler   *handlers.EvolutionHandler
	geneHandler        *handlers.QuantumGeneHandler
	coordinationHandler *handlers.CoordinationHandler
	
	// gRPC服务�?	grpcServer           *grpc.Server
	consciousnessGRPCServer *grpcserver.ConsciousnessServer
	
	// 配置和依�?	config *ModuleConfig
	db     *gorm.DB
	logger *zap.Logger
	
	// 运行状�?	httpStarted bool
	grpcStarted bool
}

// ModuleConfig 模块配置
type ModuleConfig struct {
	// HTTP配置
	HTTPEnabled bool   `json:"http_enabled"`
	HTTPPrefix  string `json:"http_prefix"`
	
	// gRPC配置
	GRPCEnabled bool   `json:"grpc_enabled"`
	GRPCPort    int    `json:"grpc_port"`
	GRPCHost    string `json:"grpc_host"`
	
	// 服务配置
	ServiceConfig *services.ConsciousnessConfig `json:"service_config"`
	
	// 组件配置
	FusionConfig      *engines.FusionEngineConfig           `json:"fusion_config"`
	EvolutionConfig   *engines.EvolutionTrackerConfig       `json:"evolution_config"`
	GeneConfig        *engines.QuantumGeneManagerConfig     `json:"gene_config"`
	CoordinationConfig *coordinators.ThreeAxisCoordinatorConfig `json:"coordination_config"`
}

// NewModule 创建意识服务模块
func NewModule(config *ModuleConfig, db *gorm.DB, logger *zap.Logger) (*Module, error) {
	if config == nil {
		config = getDefaultConfig()
	}
	
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// 设置默认配置
	if config.ServiceConfig == nil {
		config.ServiceConfig = &services.ConsciousnessConfig{
			ServiceName:    "consciousness-service",
			Version:        "1.0.0",
			Environment:    "development",
			UpdateInterval: 30 * time.Second,
			FusionConfig:   config.FusionConfig,
			EvolutionConfig: config.EvolutionConfig,
			GeneConfig:     config.GeneConfig,
			CoordinationConfig: config.CoordinationConfig,
		}
	}

	// 创建核心服务
	consciousnessService, err := services.NewConsciousnessService(config.ServiceConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create consciousness service: %w", err)
	}

	module := &Module{
		consciousnessService: consciousnessService,
		config:              config,
		db:                  db,
		logger:              logger,
	}

	// 初始化HTTP处理�?	if config.HTTPEnabled {
		if err := module.initHTTPHandlers(); err != nil {
			return nil, fmt.Errorf("failed to initialize HTTP handlers: %w", err)
		}
	}

	// 初始化gRPC服务�?	if config.GRPCEnabled {
		if err := module.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}

	return module, nil
}

// initHTTPHandlers 初始化HTTP处理�?func (m *Module) initHTTPHandlers() error {
	m.fusionHandler = handlers.NewFusionHandler(m.consciousnessService.GetFusionEngine(), m.logger)
	m.evolutionHandler = handlers.NewEvolutionHandler(m.consciousnessService.GetEvolutionTracker(), m.logger)
	m.geneHandler = handlers.NewQuantumGeneHandler(m.consciousnessService.GetGeneManager(), m.logger)
	m.coordinationHandler = handlers.NewCoordinationHandler(m.consciousnessService.GetCoordinator(), m.logger)
	
	return nil
}

// initGRPCServer 初始化gRPC服务�?func (m *Module) initGRPCServer() error {
	m.grpcServer = grpc.NewServer()
	m.consciousnessGRPCServer = grpcserver.NewConsciousnessServer(m.consciousnessService, m.logger)
	
	// 注册gRPC服务
	pb.RegisterConsciousnessServiceServer(m.grpcServer, m.consciousnessGRPCServer)
	
	return nil
}

// SetupRoutes 设置HTTP路由
func (m *Module) SetupRoutes(router *gin.RouterGroup, jwtMiddleware gin.HandlerFunc) error {
	if !m.config.HTTPEnabled {
		m.logger.Info("HTTP routes disabled for consciousness service")
		return nil
	}

	// 创建意识服务路由�?	prefix := m.config.HTTPPrefix
	if prefix == "" {
		prefix = "/consciousness"
	}
	
	consciousnessGroup := router.Group(prefix)
	
	// 应用JWT中间件（如果提供�?	if jwtMiddleware != nil {
		consciousnessGroup.Use(jwtMiddleware)
	}

	// 健康检查和统计（无需认证�?	consciousnessGroup.GET("/health", m.healthHandler)
	consciousnessGroup.GET("/stats", m.statsHandler)

	// 融合引擎路由
	fusionGroup := consciousnessGroup.Group("/fusion")
	{
		fusionGroup.POST("/start", m.fusionHandler.StartFusion)
		fusionGroup.GET("/status/:sessionId", m.fusionHandler.GetFusionStatus)
		fusionGroup.DELETE("/cancel/:sessionId", m.fusionHandler.CancelFusion)
		fusionGroup.GET("/history/:entityId", m.fusionHandler.GetFusionHistory)
		fusionGroup.GET("/strategies", m.fusionHandler.GetFusionStrategies)
		fusionGroup.GET("/metrics/:entityId", m.fusionHandler.GetFusionMetrics)
	}

	// 进化追踪路由
	evolutionGroup := consciousnessGroup.Group("/evolution")
	{
		evolutionGroup.GET("/state/:entityId", m.evolutionHandler.GetEvolutionState)
		evolutionGroup.PUT("/state/:entityId", m.evolutionHandler.UpdateEvolutionState)
		evolutionGroup.POST("/track", m.evolutionHandler.TrackEvolution)
		evolutionGroup.GET("/prediction/:entityId", m.evolutionHandler.GetEvolutionPrediction)
		evolutionGroup.GET("/path/:entityId", m.evolutionHandler.GetEvolutionPath)
		evolutionGroup.GET("/milestones/:entityId", m.evolutionHandler.GetEvolutionMilestones)
		evolutionGroup.GET("/sequence/:level", m.evolutionHandler.GetSequenceLevel)
		evolutionGroup.GET("/stats/:entityId", m.evolutionHandler.GetEvolutionStats)
	}

	// 量子基因路由
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

	// 三轴协调路由
	coordinationGroup := consciousnessGroup.Group("/coordination")
	{
		coordinationGroup.POST("/start", m.coordinationHandler.StartCoordination)
		coordinationGroup.GET("/status/:sessionId", m.coordinationHandler.GetCoordinationStatus)
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

// StartGRPCServer 启动gRPC服务�?func (m *Module) StartGRPCServer() error {
	if !m.config.GRPCEnabled {
		m.logger.Info("gRPC server disabled for consciousness service")
		return nil
	}

	if m.grpcStarted {
		return fmt.Errorf("gRPC server is already running")
	}

	host := m.config.GRPCHost
	if host == "" {
		host = "localhost"
	}
	
	port := m.config.GRPCPort
	if port == 0 {
		port = 50051
	}

	address := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	// 启动gRPC服务�?	go func() {
		m.logger.Info("Starting consciousness gRPC server", zap.String("address", address))
		if err := m.grpcServer.Serve(listener); err != nil {
			m.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	m.grpcStarted = true
	m.logger.Info("Consciousness gRPC server started", zap.String("address", address))

	return nil
}

// Start 启动模块
func (m *Module) Start() error {
	// 启动核心服务
	if err := m.consciousnessService.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness service: %w", err)
	}

	// 启动gRPC服务�?	if m.config.GRPCEnabled {
		if err := m.StartGRPCServer(); err != nil {
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}

	m.logger.Info("Consciousness module started successfully")
	return nil
}

// Stop 停止模块
func (m *Module) Stop() error {
	// 停止gRPC服务�?	if m.grpcStarted && m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		m.grpcStarted = false
		m.logger.Info("Consciousness gRPC server stopped")
	}

	// 停止核心服务
	if err := m.consciousnessService.Stop(); err != nil {
		m.logger.Error("Failed to stop consciousness service", zap.Error(err))
		return err
	}

	m.logger.Info("Consciousness module stopped successfully")
	return nil
}

// Health 健康检�?func (m *Module) Health() map[string]interface{} {
	health := m.consciousnessService.Health()
	health["module"] = map[string]interface{}{
		"http_enabled": m.config.HTTPEnabled,
		"http_started": m.httpStarted,
		"grpc_enabled": m.config.GRPCEnabled,
		"grpc_started": m.grpcStarted,
	}
	return health
}

// healthHandler HTTP健康检查处理器
func (m *Module) healthHandler(c *gin.Context) {
	health := m.Health()
	c.JSON(200, health)
}

// statsHandler HTTP统计信息处理�?func (m *Module) statsHandler(c *gin.Context) {
	stats := m.consciousnessService.GetStats()
	c.JSON(200, stats)
}

// GetService 获取核心服务实例
func (m *Module) GetService() *services.ConsciousnessService {
	return m.consciousnessService
}

// GetConfig 获取模块配置
func (m *Module) GetConfig() *ModuleConfig {
	return m.config
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *ModuleConfig {
	return &ModuleConfig{
		HTTPEnabled: true,
		HTTPPrefix:  "/consciousness",
		GRPCEnabled: true,
		GRPCPort:    50051,
		GRPCHost:    "localhost",
		ServiceConfig: &services.ConsciousnessConfig{
			ServiceName:    "consciousness-service",
			Version:        "1.0.0",
			Environment:    "development",
			UpdateInterval: 30 * time.Second,
			MaxConcurrentSessions: 100,
			SessionTimeout:        30 * time.Minute,
			MetricsRetention:      24 * time.Hour,
			EnableAuthentication:  true,
			AllowedOrigins:       []string{"*"},
			RateLimitRPS:         100,
		},
		FusionConfig: &engines.FusionEngineConfig{
			MaxConcurrentSessions: 50,
			DefaultStrategy:       "complementary",
			QualityThreshold:      0.7,
			TimeoutDuration:       5 * time.Minute,
		},
		EvolutionConfig: &engines.EvolutionTrackerConfig{
			UpdateInterval:    30 * time.Second,
			MetricsRetention:  24 * time.Hour,
			PredictionHorizon: 30,
		},
		GeneConfig: &engines.QuantumGeneManagerConfig{
			MaxGenesPerPool:   1000,
			MutationRate:      0.01,
			ExpressionDecay:   0.1,
			InteractionRadius: 5,
		},
		CoordinationConfig: &coordinators.ThreeAxisCoordinatorConfig{
			MaxConcurrentSessions: 30,
			BalanceThreshold:      0.8,
			SynergyThreshold:      0.7,
			OptimizationInterval:  1 * time.Minute,
		},
	}
}

// SetupRoutes 全局路由设置函数（用于与其他模块保持一致的接口�?func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, logger *zap.Logger, jwtMiddleware gin.HandlerFunc) error {
	// 创建默认配置的模�?	module, err := NewModule(nil, db, logger)
	if err != nil {
		return fmt.Errorf("failed to create consciousness module: %w", err)
	}

	// 启动模块
	if err := module.Start(); err != nil {
		return fmt.Errorf("failed to start consciousness module: %w", err)
	}

	// 设置路由
	if err := module.SetupRoutes(router, jwtMiddleware); err != nil {
		return fmt.Errorf("failed to setup consciousness routes: %w", err)
	}

	logger.Info("Consciousness service routes setup completed")
	return nil
}
