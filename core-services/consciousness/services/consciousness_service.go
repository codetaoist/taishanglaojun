package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/coordinators"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"go.uber.org/zap"
)

// ConsciousnessService 意识服务主类
type ConsciousnessService struct {
	// 核心组件
	fusionEngine     *engines.FusionEngine
	evolutionTracker *engines.EvolutionTracker
	geneManager      *engines.QuantumGeneManager
	coordinator      *coordinators.ThreeAxisCoordinator

	// 配置和状�?	config *ConsciousnessConfig
	logger *zap.Logger
	
	// 运行状�?	isRunning bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	// 统计信息
	stats *ServiceStats
}

// ConsciousnessConfig 意识服务配置
type ConsciousnessConfig struct {
	// 服务基础配置
	ServiceName    string        `json:"service_name"`
	Version        string        `json:"version"`
	Environment    string        `json:"environment"`
	UpdateInterval time.Duration `json:"update_interval"`

	// 组件配置
	FusionConfig      *engines.FusionEngineConfig           `json:"fusion_config"`
	EvolutionConfig   *engines.EvolutionTrackerConfig       `json:"evolution_config"`
	GeneConfig        *engines.QuantumGeneManagerConfig     `json:"gene_config"`
	CoordinationConfig *coordinators.ThreeAxisCoordinatorConfig `json:"coordination_config"`

	// 性能配置
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	MetricsRetention      time.Duration `json:"metrics_retention"`

	// 安全配置
	EnableAuthentication bool     `json:"enable_authentication"`
	AllowedOrigins      []string `json:"allowed_origins"`
	RateLimitRPS        int      `json:"rate_limit_rps"`
}

// ServiceStats 服务统计信息
type ServiceStats struct {
	// 基础统计
	StartTime         time.Time `json:"start_time"`
	TotalRequests     int64     `json:"total_requests"`
	SuccessfulRequests int64    `json:"successful_requests"`
	FailedRequests    int64     `json:"failed_requests"`
	
	// 会话统计
	ActiveSessions    int64 `json:"active_sessions"`
	TotalSessions     int64 `json:"total_sessions"`
	CompletedSessions int64 `json:"completed_sessions"`
	
	// 组件统计
	FusionSessions      int64 `json:"fusion_sessions"`
	EvolutionTracking   int64 `json:"evolution_tracking"`
	GeneOperations      int64 `json:"gene_operations"`
	CoordinationSessions int64 `json:"coordination_sessions"`
	
	// 性能统计
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastUpdateTime      time.Time     `json:"last_update_time"`
	
	mu sync.RWMutex
}

// NewConsciousnessService 创建意识服务实例
func NewConsciousnessService(config *ConsciousnessConfig, logger *zap.Logger) (*ConsciousnessService, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// 设置默认配置
	if config.ServiceName == "" {
		config.ServiceName = "consciousness-service"
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if config.UpdateInterval == 0 {
		config.UpdateInterval = 30 * time.Second
	}
	if config.MaxConcurrentSessions == 0 {
		config.MaxConcurrentSessions = 100
	}
	if config.SessionTimeout == 0 {
		config.SessionTimeout = 30 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &ConsciousnessService{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		stats: &ServiceStats{
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
		},
	}

	// 初始化核心组�?	if err := service.initializeComponents(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return service, nil
}

// initializeComponents 初始化核心组�?func (s *ConsciousnessService) initializeComponents() error {
	var err error

	// 初始化融合引�?	if s.config.FusionConfig == nil {
		s.config.FusionConfig = &engines.FusionEngineConfig{
			MaxConcurrentSessions: 50,
			DefaultStrategy:       "complementary",
			QualityThreshold:      0.7,
			TimeoutDuration:       5 * time.Minute,
		}
	}
	s.fusionEngine, err = engines.NewFusionEngine(s.config.FusionConfig, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create fusion engine: %w", err)
	}

	// 初始化进化追踪器
	if s.config.EvolutionConfig == nil {
		s.config.EvolutionConfig = &engines.EvolutionTrackerConfig{
			UpdateInterval:    30 * time.Second,
			MetricsRetention:  24 * time.Hour,
			PredictionHorizon: 30,
		}
	}
	s.evolutionTracker, err = engines.NewEvolutionTracker(s.config.EvolutionConfig, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create evolution tracker: %w", err)
	}

	// 初始化量子基因管理器
	if s.config.GeneConfig == nil {
		s.config.GeneConfig = &engines.QuantumGeneManagerConfig{
			MaxGenesPerPool:   1000,
			MutationRate:      0.01,
			ExpressionDecay:   0.1,
			InteractionRadius: 5,
		}
	}
	s.geneManager, err = engines.NewQuantumGeneManager(s.config.GeneConfig, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create gene manager: %w", err)
	}

	// 初始化三轴协调机�?	if s.config.CoordinationConfig == nil {
		s.config.CoordinationConfig = &coordinators.ThreeAxisCoordinatorConfig{
			MaxConcurrentSessions: 30,
			BalanceThreshold:      0.8,
			SynergyThreshold:      0.7,
			OptimizationInterval:  1 * time.Minute,
		}
	}
	s.coordinator, err = coordinators.NewThreeAxisCoordinator(s.config.CoordinationConfig, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create coordinator: %w", err)
	}

	return nil
}

// Start 启动意识服务
func (s *ConsciousnessService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("service is already running")
	}

	s.logger.Info("Starting consciousness service", 
		zap.String("service", s.config.ServiceName),
		zap.String("version", s.config.Version))

	// 启动核心组件
	if err := s.fusionEngine.Start(); err != nil {
		return fmt.Errorf("failed to start fusion engine: %w", err)
	}

	if err := s.evolutionTracker.Start(); err != nil {
		return fmt.Errorf("failed to start evolution tracker: %w", err)
	}

	if err := s.geneManager.Start(); err != nil {
		return fmt.Errorf("failed to start gene manager: %w", err)
	}

	if err := s.coordinator.Start(); err != nil {
		return fmt.Errorf("failed to start coordinator: %w", err)
	}

	// 启动后台任务
	go s.backgroundTasks()

	s.isRunning = true
	s.logger.Info("Consciousness service started successfully")

	return nil
}

// Stop 停止意识服务
func (s *ConsciousnessService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("service is not running")
	}

	s.logger.Info("Stopping consciousness service")

	// 取消上下�?	s.cancel()

	// 停止核心组件
	if err := s.coordinator.Stop(); err != nil {
		s.logger.Error("Failed to stop coordinator", zap.Error(err))
	}

	if err := s.geneManager.Stop(); err != nil {
		s.logger.Error("Failed to stop gene manager", zap.Error(err))
	}

	if err := s.evolutionTracker.Stop(); err != nil {
		s.logger.Error("Failed to stop evolution tracker", zap.Error(err))
	}

	if err := s.fusionEngine.Stop(); err != nil {
		s.logger.Error("Failed to stop fusion engine", zap.Error(err))
	}

	s.isRunning = false
	s.logger.Info("Consciousness service stopped")

	return nil
}

// backgroundTasks 后台任务
func (s *ConsciousnessService) backgroundTasks() {
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateStats()
			s.performMaintenance()
		}
	}
}

// updateStats 更新统计信息
func (s *ConsciousnessService) updateStats() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	s.stats.LastUpdateTime = time.Now()
	
	// 这里可以添加更多的统计信息更新逻辑
	s.logger.Debug("Stats updated", 
		zap.Int64("total_requests", s.stats.TotalRequests),
		zap.Int64("active_sessions", s.stats.ActiveSessions))
}

// performMaintenance 执行维护任务
func (s *ConsciousnessService) performMaintenance() {
	// 清理过期会话
	// 优化内存使用
	// 更新缓存
	s.logger.Debug("Performing maintenance tasks")
}

// GetFusionEngine 获取融合引擎
func (s *ConsciousnessService) GetFusionEngine() *engines.FusionEngine {
	return s.fusionEngine
}

// GetEvolutionTracker 获取进化追踪�?func (s *ConsciousnessService) GetEvolutionTracker() *engines.EvolutionTracker {
	return s.evolutionTracker
}

// GetGeneManager 获取量子基因管理�?func (s *ConsciousnessService) GetGeneManager() *engines.QuantumGeneManager {
	return s.geneManager
}

// GetCoordinator 获取三轴协调机制
func (s *ConsciousnessService) GetCoordinator() *coordinators.ThreeAxisCoordinator {
	return s.coordinator
}

// GetStats 获取服务统计信息
func (s *ConsciousnessService) GetStats() *ServiceStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 返回统计信息的副�?	statsCopy := *s.stats
	return &statsCopy
}

// GetConfig 获取服务配置
func (s *ConsciousnessService) GetConfig() *ConsciousnessConfig {
	return s.config
}

// IsRunning 检查服务是否运行中
func (s *ConsciousnessService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// Health 健康检�?func (s *ConsciousnessService) Health() map[string]interface{} {
	s.mu.RLock()
	isRunning := s.isRunning
	s.mu.RUnlock()

	health := map[string]interface{}{
		"service":   s.config.ServiceName,
		"version":   s.config.Version,
		"status":    "unknown",
		"timestamp": time.Now(),
		"components": map[string]interface{}{
			"fusion_engine":     "unknown",
			"evolution_tracker": "unknown",
			"gene_manager":      "unknown",
			"coordinator":       "unknown",
		},
	}

	if !isRunning {
		health["status"] = "stopped"
		return health
	}

	// 检查各组件状�?	allHealthy := true
	
	if s.fusionEngine != nil {
		health["components"].(map[string]interface{})["fusion_engine"] = "healthy"
	} else {
		health["components"].(map[string]interface{})["fusion_engine"] = "unhealthy"
		allHealthy = false
	}

	if s.evolutionTracker != nil {
		health["components"].(map[string]interface{})["evolution_tracker"] = "healthy"
	} else {
		health["components"].(map[string]interface{})["evolution_tracker"] = "unhealthy"
		allHealthy = false
	}

	if s.geneManager != nil {
		health["components"].(map[string]interface{})["gene_manager"] = "healthy"
	} else {
		health["components"].(map[string]interface{})["gene_manager"] = "unhealthy"
		allHealthy = false
	}

	if s.coordinator != nil {
		health["components"].(map[string]interface{})["coordinator"] = "healthy"
	} else {
		health["components"].(map[string]interface{})["coordinator"] = "unhealthy"
		allHealthy = false
	}

	if allHealthy {
		health["status"] = "healthy"
	} else {
		health["status"] = "degraded"
	}

	return health
}

// ProcessConsciousnessRequest 处理综合意识请求
func (s *ConsciousnessService) ProcessConsciousnessRequest(req *models.ConsciousnessRequest) (*models.ConsciousnessResponse, error) {
	if !s.IsRunning() {
		return nil, fmt.Errorf("service is not running")
	}

	// 增加请求计数
	s.stats.mu.Lock()
	s.stats.TotalRequests++
	s.stats.mu.Unlock()

	startTime := time.Now()
	defer func() {
		// 更新响应时间统计
		duration := time.Since(startTime)
		s.stats.mu.Lock()
		if s.stats.AverageResponseTime == 0 {
			s.stats.AverageResponseTime = duration
		} else {
			s.stats.AverageResponseTime = (s.stats.AverageResponseTime + duration) / 2
		}
		s.stats.mu.Unlock()
	}()

	response := &models.ConsciousnessResponse{
		RequestID: req.RequestID,
		EntityID:  req.EntityID,
		Timestamp: time.Now(),
		Results:   make(map[string]interface{}),
	}

	// 根据请求类型处理不同的意识操�?	if req.FusionRequest != nil {
		fusionResult, err := s.processFusionRequest(req.FusionRequest)
		if err != nil {
			s.incrementFailedRequests()
			return nil, fmt.Errorf("fusion processing failed: %w", err)
		}
		response.Results["fusion"] = fusionResult
	}

	if req.EvolutionRequest != nil {
		evolutionResult, err := s.processEvolutionRequest(req.EvolutionRequest)
		if err != nil {
			s.incrementFailedRequests()
			return nil, fmt.Errorf("evolution processing failed: %w", err)
		}
		response.Results["evolution"] = evolutionResult
	}

	if req.GeneRequest != nil {
		geneResult, err := s.processGeneRequest(req.GeneRequest)
		if err != nil {
			s.incrementFailedRequests()
			return nil, fmt.Errorf("gene processing failed: %w", err)
		}
		response.Results["gene"] = geneResult
	}

	if req.CoordinationRequest != nil {
		coordinationResult, err := s.processCoordinationRequest(req.CoordinationRequest)
		if err != nil {
			s.incrementFailedRequests()
			return nil, fmt.Errorf("coordination processing failed: %w", err)
		}
		response.Results["coordination"] = coordinationResult
	}

	// 增加成功请求计数
	s.stats.mu.Lock()
	s.stats.SuccessfulRequests++
	s.stats.mu.Unlock()

	response.Success = true
	return response, nil
}

// processFusionRequest 处理融合请求
func (s *ConsciousnessService) processFusionRequest(req *models.FusionRequest) (interface{}, error) {
	return s.fusionEngine.StartFusion(req)
}

// processEvolutionRequest 处理进化请求
func (s *ConsciousnessService) processEvolutionRequest(req *models.EvolutionRequest) (interface{}, error) {
	switch req.Type {
	case "track":
		return s.evolutionTracker.TrackEvolution(req.EntityID, req.InitialMetrics)
	case "update":
		return s.evolutionTracker.UpdateEvolution(req.EntityID, req.Metrics)
	case "predict":
		return s.evolutionTracker.PredictEvolution(req.EntityID, req.PredictionHorizon)
	default:
		return nil, fmt.Errorf("unknown evolution request type: %s", req.Type)
	}
}

// processGeneRequest 处理基因请求
func (s *ConsciousnessService) processGeneRequest(req *models.GeneRequest) (interface{}, error) {
	switch req.Type {
	case "create_pool":
		return s.geneManager.CreateGenePool(req.EntityID, req.InitialGenes)
	case "add_gene":
		return s.geneManager.AddGene(req.EntityID, req.Gene)
	case "express":
		return s.geneManager.ExpressGene(req.EntityID, req.GeneID, req.Intensity, req.Duration)
	case "mutate":
		return s.geneManager.MutateGene(req.EntityID, req.GeneID, req.MutationType, req.MutationIntensity)
	default:
		return nil, fmt.Errorf("unknown gene request type: %s", req.Type)
	}
}

// processCoordinationRequest 处理协调请求
func (s *ConsciousnessService) processCoordinationRequest(req *models.CoordinationRequest) (interface{}, error) {
	return s.coordinator.StartCoordination(req)
}

// incrementFailedRequests 增加失败请求计数
func (s *ConsciousnessService) incrementFailedRequests() {
	s.stats.mu.Lock()
	s.stats.FailedRequests++
	s.stats.mu.Unlock()
}
