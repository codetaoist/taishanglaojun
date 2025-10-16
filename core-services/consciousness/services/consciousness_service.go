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

// LoggerAdapter 适配 zap.Logger 到 engines.Logger 接口
type LoggerAdapter struct {
	logger *zap.Logger
}

func (l *LoggerAdapter) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, zap.Any("fields", fields))
}

func (l *LoggerAdapter) Error(msg string, err error, fields ...interface{}) {
	l.logger.Error(msg, zap.Error(err), zap.Any("fields", fields))
}

func (l *LoggerAdapter) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, zap.Any("fields", fields))
}

func (l *LoggerAdapter) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, zap.Any("fields", fields))
}

// ModelsLoggerAdapter 适配器，将 zap.Logger 转换为 models.Logger
type ModelsLoggerAdapter struct {
	logger *zap.Logger
}

func (l *ModelsLoggerAdapter) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, zap.Any("fields", fields))
}

func (l *ModelsLoggerAdapter) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, zap.Any("fields", fields))
}

func (l *ModelsLoggerAdapter) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, zap.Any("fields", fields))
}

func (l *ModelsLoggerAdapter) Error(msg string, err error, fields ...interface{}) {
	l.logger.Error(msg, zap.Error(err), zap.Any("fields", fields))
}

func (l *ModelsLoggerAdapter) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal(msg, zap.Any("fields", fields))
}

func (l *ModelsLoggerAdapter) WithField(key string, value interface{}) models.Logger {
	return &ModelsLoggerAdapter{logger: l.logger.With(zap.Any(key, value))}
}

func (l *ModelsLoggerAdapter) WithFields(fields map[string]interface{}) models.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &ModelsLoggerAdapter{logger: l.logger.With(zapFields...)}
}

func (l *ModelsLoggerAdapter) WithError(err error) models.Logger {
	return &ModelsLoggerAdapter{logger: l.logger.With(zap.Error(err))}
}

// ConsciousnessService 
type ConsciousnessService struct {
	// 
	fusionEngine     *engines.FusionEngine
	evolutionTracker *engines.EvolutionTracker
	geneManager      *engines.QuantumGeneManager
	coordinator      *coordinators.ThreeAxisCoordinator

	// 
	config *ConsciousnessConfig
	logger *zap.Logger

	// 
	isRunning bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc

	// 
	stats *ServiceStats
}

// ConsciousnessConfig 
type ConsciousnessConfig struct {
	// 
	ServiceName    string        `json:"service_name"`
	Version        string        `json:"version"`
	Environment    string        `json:"environment"`
	UpdateInterval time.Duration `json:"update_interval"`

	// 
	FusionConfig       *engines.FusionEngineConfig              `json:"fusion_config"`
	EvolutionConfig    *engines.EvolutionTrackerConfig          `json:"evolution_config"`
	GeneConfig         *engines.QuantumGeneManagerConfig        `json:"gene_config"`
	CoordinationConfig *coordinators.ThreeAxisCoordinatorConfig `json:"coordination_config"`

	// 
	MaxConcurrentSessions int           `json:"max_concurrent_sessions"`
	SessionTimeout        time.Duration `json:"session_timeout"`
	MetricsRetention      time.Duration `json:"metrics_retention"`

	// 
	EnableAuthentication bool     `json:"enable_authentication"`
	AllowedOrigins       []string `json:"allowed_origins"`
	RateLimitRPS         int      `json:"rate_limit_rps"`
}

// ServiceStats 
type ServiceStats struct {
	// 
	StartTime          time.Time `json:"start_time"`
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`

	// 
	ActiveSessions    int64 `json:"active_sessions"`
	TotalSessions     int64 `json:"total_sessions"`
	CompletedSessions int64 `json:"completed_sessions"`

	// 
	FusionSessions       int64 `json:"fusion_sessions"`
	EvolutionTracking    int64 `json:"evolution_tracking"`
	GeneOperations       int64 `json:"gene_operations"`
	CoordinationSessions int64 `json:"coordination_sessions"`

	// 
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastUpdateTime      time.Time     `json:"last_update_time"`

	mu sync.RWMutex
}

// NewConsciousnessService 
func NewConsciousnessService(config *ConsciousnessConfig, logger *zap.Logger) (*ConsciousnessService, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// 
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

	// 
	if err := service.initializeComponents(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return service, nil
}

// initializeComponents 
func (s *ConsciousnessService) initializeComponents() error {
	var err error

	// 
	if s.config.FusionConfig == nil {
		s.config.FusionConfig = &engines.FusionEngineConfig{
			MaxConcurrentSessions:    50,
			SessionTimeout:           5 * time.Minute,
			QualityThreshold:         0.7,
			SynergyThreshold:         0.6,
			EnableEmergenceDetection: true,
			CarbonWeight:             0.5,
			SiliconWeight:            0.5,
			FusionStrategies:         []string{"complementary", "synergistic"},
		}
	}
	s.fusionEngine = engines.NewFusionEngine(s.config.FusionConfig)

	// 
	if s.config.EvolutionConfig == nil {
		s.config.EvolutionConfig = &engines.EvolutionTrackerConfig{
			UpdateInterval:          30 * time.Second,
			MaxConcurrentTracking:   10,
			PredictionHorizon:       24 * time.Hour,
			MetricsRetentionPeriod:  24 * time.Hour,
			EnableRealTimeTracking:  true,
			EnablePredictiveAnalysis: true,
			EnablePathOptimization:  true,
			MinConfidenceThreshold:  0.7,
			MaxEvolutionDuration:    30 * 24 * time.Hour,
		}
	}
	// 创建 Logger 适配器
	loggerAdapter := &LoggerAdapter{logger: s.logger}
	s.evolutionTracker = engines.NewEvolutionTracker(s.config.EvolutionConfig, loggerAdapter)
	if err != nil {
		return fmt.Errorf("failed to create evolution tracker: %w", err)
	}

	// 
	if s.config.GeneConfig == nil {
		s.config.GeneConfig = &engines.QuantumGeneManagerConfig{
			MaxGenePools:                100,
			MaxGenesPerPool:             1000,
			MutationRate:                0.01,
			ExpressionUpdateInterval:    time.Minute * 5,
			EvolutionSimulationInterval: time.Hour,
			EnableAutoMutation:          true,
			EnableExpressionControl:     true,
			EnableInteractionAnalysis:   true,
			EnableEvolutionSimulation:   true,
			GeneStabilityThreshold:      0.8,
			ExpressionThreshold:         0.5,
		}
	}
	s.geneManager = engines.NewQuantumGeneManager(s.config.GeneConfig, loggerAdapter)

	// 暂时移除 CoordinationConfig，因为 ThreeAxisCoordinator 尚未实现
	// if s.config.CoordinationConfig == nil {
	// 	s.config.CoordinationConfig = &coordinators.ThreeAxisCoordinatorConfig{
	// 		MaxConcurrentSessions: 30,
	// 		BalanceThreshold:      0.8,
	// 		SynergyThreshold:      0.7,
	// 		OptimizationInterval:  1 * time.Minute,
	// 	}
	// }
	// 创建 Logger 适配器
	// loggerAdapter := &LoggerAdapter{logger: s.logger}
	// s.coordinator = coordinators.NewThreeAxisCoordinator(s.config.CoordinationConfig, loggerAdapter)

	return nil
}

// Start 
func (s *ConsciousnessService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("service is already running")
	}

	s.logger.Info("Starting consciousness service",
		zap.String("service", s.config.ServiceName),
		zap.String("version", s.config.Version))

	// 初始化融合引擎
	if err := s.fusionEngine.Initialize(s.ctx); err != nil {
		return fmt.Errorf("failed to initialize fusion engine: %w", err)
	}

	if err := s.evolutionTracker.Start(s.ctx); err != nil {
		return fmt.Errorf("failed to start evolution tracker: %w", err)
	}

	if err := s.geneManager.Start(s.ctx); err != nil {
		return fmt.Errorf("failed to start gene manager: %w", err)
	}

	// 暂时注释掉coordinator的启动，因为它没有被初始化
	// if err := s.coordinator.Start(s.ctx); err != nil {
	// 	return fmt.Errorf("failed to start coordinator: %w", err)
	// }

	// 
	go s.backgroundTasks()

	s.isRunning = true
	s.logger.Info("Consciousness service started successfully")

	return nil
}

// Stop 
func (s *ConsciousnessService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("service is not running")
	}

	s.logger.Info("Stopping consciousness service")

	// 
	s.cancel()

	// 
	if err := s.coordinator.Stop(); err != nil {
		s.logger.Error("Failed to stop coordinator", zap.Error(err))
	}

	if err := s.geneManager.Stop(); err != nil {
		s.logger.Error("Failed to stop gene manager", zap.Error(err))
	}

	if err := s.evolutionTracker.Stop(); err != nil {
		s.logger.Error("Failed to stop evolution tracker", zap.Error(err))
	}

	// FusionEngine 没有 Stop 方法，无需调用

	s.isRunning = false
	s.logger.Info("Consciousness service stopped")

	return nil
}

// backgroundTasks 
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

// updateStats 
func (s *ConsciousnessService) updateStats() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()

	s.stats.LastUpdateTime = time.Now()

	// 
	s.logger.Debug("Stats updated",
		zap.Int64("total_requests", s.stats.TotalRequests),
		zap.Int64("active_sessions", s.stats.ActiveSessions))
}

// performMaintenance 
func (s *ConsciousnessService) performMaintenance() {
	// 
	// 
	// 
	s.logger.Debug("Performing maintenance tasks")
}

// GetFusionEngine 
func (s *ConsciousnessService) GetFusionEngine() *engines.FusionEngine {
	return s.fusionEngine
}

// GetEvolutionTracker 
func (s *ConsciousnessService) GetEvolutionTracker() *engines.EvolutionTracker {
	return s.evolutionTracker
}

// GetGeneManager 
func (s *ConsciousnessService) GetGeneManager() *engines.QuantumGeneManager {
	return s.geneManager
}

// GetCoordinator 
func (s *ConsciousnessService) GetCoordinator() *coordinators.ThreeAxisCoordinator {
	return s.coordinator
}

// GetStats 
func (s *ConsciousnessService) GetStats() *ServiceStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 
	statsCopy := *s.stats
	return &statsCopy
}

// GetConfig 
func (s *ConsciousnessService) GetConfig() *ConsciousnessConfig {
	return s.config
}

// IsRunning 
func (s *ConsciousnessService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// Health 
func (s *ConsciousnessService) Health() map[string]interface{} {
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

	// 
	allHealthy := true

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

// ProcessConsciousnessRequest 
func (s *ConsciousnessService) ProcessConsciousnessRequest(req *models.ConsciousnessRequest) (*models.ConsciousnessResponse, error) {
	if !s.IsRunning() {
		return nil, fmt.Errorf("service is not running")
	}

	// 
	s.stats.mu.Lock()
	s.stats.TotalRequests++
	s.stats.mu.Unlock()

	startTime := time.Now()
	defer func() {
		// 
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
		RequestID:   req.ID,
		ProcessedAt: time.Now(),
		Success:     false,
		Metrics:     make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	// 根据请求类型处理
	var result interface{}
	var err error
	
	switch req.Type {
	case "fusion":
		// 处理融合请求
		// 创建 FusionRequest 对象
		fusionReq := &models.FusionRequest{
			ID:          req.ID,
			EntityID:    req.EntityID,
			CarbonData:  req.Parameters["carbon_data"].(*models.CarbonInput),
			SiliconData: req.Parameters["silicon_data"].(*models.SiliconInput),
		}
		result, err = s.processFusionRequest(context.Background(), fusionReq)
		if err != nil {
			s.incrementFailedRequests()
			return nil, fmt.Errorf("fusion request failed: %w", err)
		}
	case "evolution":
		// 处理进化请求
		result = map[string]interface{}{
			"type":      "evolution",
			"status":    "processed",
			"entity_id": req.EntityID,
		}
	case "gene":
		// 处理基因请求
		result = map[string]interface{}{
			"type": "gene",
			"status": "processed", 
			"entity_id": req.EntityID,
		}
	case "coordination":
		// 处理协调请求
		result = map[string]interface{}{
			"type": "coordination",
			"status": "processed",
			"entity_id": req.EntityID,
		}
	default:
		s.incrementFailedRequests()
		return nil, fmt.Errorf("unknown request type: %s", req.Type)
	}
	
	if err != nil {
		s.incrementFailedRequests()
		return nil, fmt.Errorf("processing failed: %w", err)
	}
	
	// 设置结果
	response.Result = result

	// 
	s.stats.mu.Lock()
	s.stats.SuccessfulRequests++
	s.stats.mu.Unlock()

	response.Success = true
	return response, nil
}

// processFusionRequest 处理融合请求
func (s *ConsciousnessService) processFusionRequest(ctx context.Context, req *models.FusionRequest) (interface{}, error) {
	return s.fusionEngine.StartFusion(ctx, req.CarbonData, req.SiliconData)
}



// processGeneRequest 处理基因请求
func (s *ConsciousnessService) processGeneRequest(req *models.GeneRequest) (interface{}, error) {
	// 简化的基因请求处理
	return map[string]interface{}{
		"type":      "gene",
		"status":    "processed",
		"entity_id": req.EntityID,
		"action":    req.Action,
	}, nil
}

// processCoordinationRequest 处理协调请求
func (s *ConsciousnessService) processCoordinationRequest(req interface{}) (interface{}, error) {
	// 简化的协调请求处理
	return map[string]interface{}{
		"type":   "coordination",
		"status": "processed",
	}, nil
}

// incrementFailedRequests 
func (s *ConsciousnessService) incrementFailedRequests() {
	s.stats.mu.Lock()
	s.stats.FailedRequests++
	s.stats.mu.Unlock()
}

