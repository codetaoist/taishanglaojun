package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ServiceInitializer 服务初始化器
type ServiceInitializer struct {
	configManager *ServiceConfigManager
	services      map[string]interface{}
	integration   *IntelligentLearningServiceIntegration
	initialized   bool
	mu            sync.RWMutex
}

// InitializationResult 初始化结果
type InitializationResult struct {
	ServiceName   string        `json:"service_name"`
	Success       bool          `json:"success"`
	Error         error         `json:"error,omitempty"`
	InitTime      time.Duration `json:"init_time"`
	Dependencies  []string      `json:"dependencies"`
	Status        string        `json:"status"`
}

// ServiceDependency 服务依赖
type ServiceDependency struct {
	ServiceName  string   `json:"service_name"`
	Dependencies []string `json:"dependencies"`
	Priority     int      `json:"priority"`
}

// NewServiceInitializer 创建服务初始化器
func NewServiceInitializer(configPath string) *ServiceInitializer {
	return &ServiceInitializer{
		configManager: NewServiceConfigManager(configPath),
		services:      make(map[string]interface{}),
		initialized:   false,
	}
}

// Initialize 初始化所有服务
func (si *ServiceInitializer) Initialize(ctx context.Context) ([]*InitializationResult, error) {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if si.initialized {
		return nil, fmt.Errorf("services already initialized")
	}
	
	log.Println("Starting intelligent learning services initialization...")
	
	// 加载配置
	if err := si.configManager.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	// 验证配置
	if err := si.configManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	config := si.configManager.GetConfig()
	
	// 获取服务依赖关系
	dependencies := si.getServiceDependencies()
	
	// 按依赖顺序初始化服务
	results := make([]*InitializationResult, 0)
	
	for _, dep := range dependencies {
		if si.configManager.IsServiceEnabled(dep.ServiceName) {
			result := si.initializeService(ctx, dep.ServiceName, config)
			results = append(results, result)
			
			if !result.Success {
				log.Printf("Failed to initialize service %s: %v", dep.ServiceName, result.Error)
				// 根据配置决定是否继续初始化其他服务
				if si.isServiceCritical(dep.ServiceName) {
					return results, fmt.Errorf("critical service %s failed to initialize", dep.ServiceName)
				}
			} else {
				log.Printf("Successfully initialized service: %s", dep.ServiceName)
			}
		} else {
			log.Printf("Service %s is disabled, skipping initialization", dep.ServiceName)
		}
	}
	
	// 初始化服务集成
	if err := si.initializeIntegration(ctx, config); err != nil {
		return results, fmt.Errorf("failed to initialize service integration: %w", err)
	}
	
	si.initialized = true
	log.Println("All services initialized successfully")
	
	return results, nil
}

// getServiceDependencies 获取服务依赖关系
func (si *ServiceInitializer) getServiceDependencies() []*ServiceDependency {
	return []*ServiceDependency{
		{
			ServiceName:  "cross_modal",
			Dependencies: []string{},
			Priority:     1,
		},
		{
			ServiceName:  "relation_inference",
			Dependencies: []string{"cross_modal"},
			Priority:     2,
		},
		{
			ServiceName:  "knowledge_graph",
			Dependencies: []string{"cross_modal", "relation_inference"},
			Priority:     3,
		},
		{
			ServiceName:  "realtime_analytics",
			Dependencies: []string{"cross_modal"},
			Priority:     2,
		},
		{
			ServiceName:  "adaptive_learning",
			Dependencies: []string{"realtime_analytics", "knowledge_graph"},
			Priority:     4,
		},
		{
			ServiceName:  "content_recommendation",
			Dependencies: []string{"adaptive_learning", "relation_inference"},
			Priority:     5,
		},
		{
			ServiceName:  "analytics_reporting",
			Dependencies: []string{"realtime_analytics", "adaptive_learning"},
			Priority:     5,
		},
	}
}

// initializeService 初始化单个服务
func (si *ServiceInitializer) initializeService(ctx context.Context, serviceName string, config *GlobalServiceConfig) *InitializationResult {
	startTime := time.Now()
	
	result := &InitializationResult{
		ServiceName: serviceName,
		Success:     false,
		InitTime:    0,
		Status:      "initializing",
	}
	
	defer func() {
		result.InitTime = time.Since(startTime)
		if result.Success {
			result.Status = "running"
		} else {
			result.Status = "failed"
		}
	}()
	
	switch serviceName {
	case "cross_modal":
		service, err := si.initializeCrossModalService(ctx, config.CrossModalService)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	case "relation_inference":
		service, err := si.initializeRelationInferenceEngine(ctx, config.RelationInferenceEngine)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	case "adaptive_learning":
		service, err := si.initializeAdaptiveLearningEngine(ctx, config.AdaptiveLearningEngine)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	case "realtime_analytics":
		service, err := si.initializeRealtimeLearningAnalyticsService(ctx, config.RealtimeLearningAnalyticsService)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	case "knowledge_graph":
		service, err := si.initializeAutomatedKnowledgeGraphService(ctx, config.AutomatedKnowledgeGraphService)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	case "analytics_reporting":
		service, err := si.initializeLearningAnalyticsReportingService(ctx, config.LearningAnalyticsReportingService)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	case "content_recommendation":
		service, err := si.initializeIntelligentContentRecommendationService(ctx, config.IntelligentContentRecommendationService)
		if err != nil {
			result.Error = err
			return result
		}
		si.services[serviceName] = service
		
	default:
		result.Error = fmt.Errorf("unknown service: %s", serviceName)
		return result
	}
	
	result.Success = true
	return result
}

// initializeCrossModalService 初始化跨模态服务
func (si *ServiceInitializer) initializeCrossModalService(ctx context.Context, config *CrossModalServiceConfig) (CrossModalServiceInterface, error) {
	log.Printf("Initializing Cross Modal Service with config: %+v", config)
	
	// 这里应该根据实际的跨模态服务实现来初始化
	// 目前返回一个模拟的服务实例
	service := &MockCrossModalService{
		config:      config,
		initialized: true,
	}
	
	return service, nil
}

// initializeRelationInferenceEngine 初始化关系推理引擎
func (si *ServiceInitializer) initializeRelationInferenceEngine(ctx context.Context, config *RelationInferenceEngineConfig) (*IntelligentRelationInferenceEngine, error) {
	log.Printf("Initializing Relation Inference Engine with config: %+v", config)
	
	// 获取依赖的跨模态服务
	crossModalService, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// 创建关系推理引擎配置
	engineConfig := &RelationInferenceConfig{
		InferenceAlgorithms: config.InferenceAlgorithms,
		ConfidenceThreshold: config.ConfidenceThreshold,
		MaxInferenceDepth:   config.MaxInferenceDepth,
		CacheSize:           config.CacheSize,
		AlgorithmWeights:    config.AlgorithmWeights,
	}
	
	// 初始化关系推理引擎
	engine := NewIntelligentRelationInferenceEngine(
		crossModalService.(CrossModalServiceInterface),
		engineConfig,
	)
	
	return engine, nil
}

// initializeAdaptiveLearningEngine 初始化自适应学习引擎
func (si *ServiceInitializer) initializeAdaptiveLearningEngine(ctx context.Context, config *AdaptiveLearningEngineConfig) (*AdaptiveLearningEngine, error) {
	log.Printf("Initializing Adaptive Learning Engine with config: %+v", config)
	
	// 获取依赖服务
	realtimeAnalytics, exists := si.services["realtime_analytics"]
	if !exists {
		return nil, fmt.Errorf("realtime analytics service dependency not found")
	}
	
	knowledgeGraph, exists := si.services["knowledge_graph"]
	if !exists {
		return nil, fmt.Errorf("knowledge graph service dependency not found")
	}
	
	// 创建自适应学习引擎配置
	engineConfig := &AdaptiveLearningConfig{
		AdaptationStrategies:   config.AdaptationStrategies,
		LearningPathAlgorithms: config.LearningPathAlgorithms,
		PersonalizationLevel:   config.PersonalizationLevel,
		UpdateFrequency:        config.UpdateFrequency,
		ModelParameters:        config.ModelParameters,
		QualityMetrics:         config.QualityMetrics,
	}
	
	// 初始化自适应学习引擎
	engine := NewAdaptiveLearningEngine(
		realtimeAnalytics.(*RealtimeLearningAnalyticsService),
		knowledgeGraph.(*AutomatedKnowledgeGraphService),
		engineConfig,
	)
	
	return engine, nil
}

// initializeRealtimeLearningAnalyticsService 初始化实时学习分析服务
func (si *ServiceInitializer) initializeRealtimeLearningAnalyticsService(ctx context.Context, config *RealtimeLearningAnalyticsServiceConfig) (*RealtimeLearningAnalyticsService, error) {
	log.Printf("Initializing Realtime Learning Analytics Service with config: %+v", config)
	
	// 获取依赖的跨模态服务
	crossModalService, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// 创建实时学习分析服务配置
	serviceConfig := &RealtimeAnalyticsConfig{
		DataStreamSources:   config.DataStreamSources,
		ProcessingInterval:  config.ProcessingInterval,
		AnalyticsAlgorithms: config.AnalyticsAlgorithms,
		AlertThresholds:     config.AlertThresholds,
		DataRetentionPeriod: config.DataRetentionPeriod,
		StreamingConfig:     config.StreamingConfig,
	}
	
	// 初始化实时学习分析服务
	service := NewRealtimeLearningAnalyticsService(
		crossModalService.(CrossModalServiceInterface),
		serviceConfig,
	)
	
	return service, nil
}

// initializeAutomatedKnowledgeGraphService 初始化自动化知识图谱服务
func (si *ServiceInitializer) initializeAutomatedKnowledgeGraphService(ctx context.Context, config *AutomatedKnowledgeGraphServiceConfig) (*AutomatedKnowledgeGraphService, error) {
	log.Printf("Initializing Automated Knowledge Graph Service with config: %+v", config)
	
	// 获取依赖服务
	crossModalService, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	relationInference, exists := si.services["relation_inference"]
	if !exists {
		return nil, fmt.Errorf("relation inference engine dependency not found")
	}
	
	// 创建知识图谱服务配置
	serviceConfig := &KnowledgeGraphConfig{
		GraphDatabases:           config.GraphDatabases,
		EntityExtractionModels:   config.EntityExtractionModels,
		RelationExtractionModels: config.RelationExtractionModels,
		UpdateStrategies:         config.UpdateStrategies,
		ValidationRules:          config.ValidationRules,
		GraphOptimization:        config.GraphOptimization,
	}
	
	// 初始化知识图谱服务
	service := NewAutomatedKnowledgeGraphService(
		crossModalService.(CrossModalServiceInterface),
		relationInference.(*IntelligentRelationInferenceEngine),
		serviceConfig,
	)
	
	return service, nil
}

// initializeLearningAnalyticsReportingService 初始化学习分析报告服务
func (si *ServiceInitializer) initializeLearningAnalyticsReportingService(ctx context.Context, config *LearningAnalyticsReportingServiceConfig) (*LearningAnalyticsReportingService, error) {
	log.Printf("Initializing Learning Analytics Reporting Service with config: %+v", config)
	
	// 获取依赖服务
	crossModalService, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	relationInference, exists := si.services["relation_inference"]
	if !exists {
		return nil, fmt.Errorf("relation inference engine dependency not found")
	}
	
	realtimeAnalytics, exists := si.services["realtime_analytics"]
	if !exists {
		return nil, fmt.Errorf("realtime analytics service dependency not found")
	}
	
	adaptiveLearning, exists := si.services["adaptive_learning"]
	if !exists {
		return nil, fmt.Errorf("adaptive learning engine dependency not found")
	}
	
	knowledgeGraph, exists := si.services["knowledge_graph"]
	if !exists {
		return nil, fmt.Errorf("knowledge graph service dependency not found")
	}
	
	// 创建学习分析报告服务配置
	serviceConfig := &AnalyticsReportingConfig{
		ReportSettings: &ReportSettings{
			ReportTypes:        config.ReportTypes,
			GenerationSchedule: config.GenerationSchedule,
			ExportFormats:      config.ExportFormats,
			QualityStandards:   config.QualityStandards,
		},
		VisualizationSettings: &VisualizationSettings{
			VisualizationEngines: config.VisualizationEngines,
		},
		ExportSettings: &ExportSettings{
			ExportFormats: config.ExportFormats,
		},
		PerformanceSettings: &PerformanceSettings{
			CacheExpiration: config.CacheSettings.TTL,
		},
	}
	
	// 初始化学习分析报告服务
	service := NewLearningAnalyticsReportingService(
		crossModalService.(CrossModalServiceInterface),
		relationInference.(*IntelligentRelationInferenceEngine),
		realtimeAnalytics.(*RealtimeLearningAnalyticsService),
		adaptiveLearning.(*AdaptiveLearningEngine),
		knowledgeGraph.(*AutomatedKnowledgeGraphService),
		serviceConfig,
	)
	
	return service, nil
}

// initializeIntelligentContentRecommendationService 初始化智能内容推荐服务
func (si *ServiceInitializer) initializeIntelligentContentRecommendationService(ctx context.Context, config *IntelligentContentRecommendationServiceConfig) (*IntelligentContentRecommendationService, error) {
	log.Printf("Initializing Intelligent Content Recommendation Service with config: %+v", config)
	
	// 获取依赖服务
	adaptiveLearning, exists := si.services["adaptive_learning"]
	if !exists {
		return nil, fmt.Errorf("adaptive learning engine dependency not found")
	}
	
	relationInference, exists := si.services["relation_inference"]
	if !exists {
		return nil, fmt.Errorf("relation inference engine dependency not found")
	}
	
	// 创建智能内容推荐服务配置
	serviceConfig := &ContentRecommendationConfig{
		RecommendationSettings: &RecommendationSettings{
			RecommendationAlgorithms:  config.RecommendationAlgorithms,
			PersonalizationStrategies: config.PersonalizationStrategies,
			RecommendationLimits:      config.RecommendationLimits,
			QualityFilters:            config.QualityFilters,
		},
		ContentAnalysisSettings: &ContentAnalysisSettings{
			ContentAnalysisModels: config.ContentAnalysisModels,
		},
		LearnerProfilingSettings: &LearnerProfilingSettings{
			LearnerProfilingConfig: config.LearnerProfilingConfig,
		},
	}
	
	// 初始化智能内容推荐服务
	service := NewIntelligentContentRecommendationService(
		adaptiveLearning.(*AdaptiveLearningEngine),
		relationInference.(*IntelligentRelationInferenceEngine),
		serviceConfig,
	)
	
	return service, nil
}

// initializeIntegration 初始化服务集成
func (si *ServiceInitializer) initializeIntegration(ctx context.Context, config *GlobalServiceConfig) error {
	log.Println("Initializing service integration...")
	
	// 获取所有已初始化的服务
	crossModalService, _ := si.services["cross_modal"]
	relationInference, _ := si.services["relation_inference"]
	adaptiveLearning, _ := si.services["adaptive_learning"]
	realtimeAnalytics, _ := si.services["realtime_analytics"]
	knowledgeGraph, _ := si.services["knowledge_graph"]
	analyticsReporting, _ := si.services["analytics_reporting"]
	contentRecommendation, _ := si.services["content_recommendation"]
	
	// 创建服务集成
	si.integration = NewIntelligentLearningServiceIntegration(
		crossModalService.(CrossModalServiceInterface),
		relationInference.(*IntelligentRelationInferenceEngine),
		adaptiveLearning.(*AdaptiveLearningEngine),
		realtimeAnalytics.(*RealtimeLearningAnalyticsService),
		knowledgeGraph.(*AutomatedKnowledgeGraphService),
		analyticsReporting.(*LearningAnalyticsReportingService),
		contentRecommendation.(*IntelligentContentRecommendationService),
		config.IntegrationConfig,
	)
	
	log.Println("Service integration initialized successfully")
	return nil
}

// isServiceCritical 检查服务是否为关键服务
func (si *ServiceInitializer) isServiceCritical(serviceName string) bool {
	criticalServices := []string{
		"cross_modal",
		"realtime_analytics",
	}
	
	for _, critical := range criticalServices {
		if serviceName == critical {
			return true
		}
	}
	
	return false
}

// GetService 获取服务实例
func (si *ServiceInitializer) GetService(serviceName string) (interface{}, error) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.initialized {
		return nil, fmt.Errorf("services not initialized")
	}
	
	service, exists := si.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	
	return service, nil
}

// GetIntegration 获取服务集成实例
func (si *ServiceInitializer) GetIntegration() (*IntelligentLearningServiceIntegration, error) {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	if !si.initialized {
		return nil, fmt.Errorf("services not initialized")
	}
	
	if si.integration == nil {
		return nil, fmt.Errorf("service integration not found")
	}
	
	return si.integration, nil
}

// Shutdown 关闭所有服务
func (si *ServiceInitializer) Shutdown(ctx context.Context) error {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if !si.initialized {
		return fmt.Errorf("services not initialized")
	}
	
	log.Println("Shutting down intelligent learning services...")
	
	// 关闭服务集成
	if si.integration != nil {
		if err := si.integration.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down service integration: %v", err)
		}
	}
	
	// 关闭各个服务（按相反顺序）
	shutdownOrder := []string{
		"analytics_reporting",
		"content_recommendation",
		"adaptive_learning",
		"knowledge_graph",
		"realtime_analytics",
		"relation_inference",
		"cross_modal",
	}
	
	for _, serviceName := range shutdownOrder {
		if service, exists := si.services[serviceName]; exists {
			if shutdownable, ok := service.(interface{ Shutdown(context.Context) error }); ok {
				if err := shutdownable.Shutdown(ctx); err != nil {
					log.Printf("Error shutting down service %s: %v", serviceName, err)
				} else {
					log.Printf("Service %s shut down successfully", serviceName)
				}
			}
		}
	}
	
	si.initialized = false
	si.services = make(map[string]interface{})
	si.integration = nil
	
	log.Println("All services shut down successfully")
	return nil
}

// IsInitialized 检查是否已初始化
func (si *ServiceInitializer) IsInitialized() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.initialized
}

// GetInitializationStatus 获取初始化状态
func (si *ServiceInitializer) GetInitializationStatus() map[string]string {
	si.mu.RLock()
	defer si.mu.RUnlock()
	
	status := make(map[string]string)
	
	if !si.initialized {
		status["overall"] = "not_initialized"
		return status
	}
	
	status["overall"] = "initialized"
	
	for serviceName := range si.services {
		status[serviceName] = "running"
	}
	
	if si.integration != nil {
		status["integration"] = "running"
	}
	
	return status
}

// MockCrossModalService 模拟跨模态服务（用于测试）
type MockCrossModalService struct {
	config      *CrossModalServiceConfig
	initialized bool
}

// ProcessMultiModalContent 处理多模态内容
func (m *MockCrossModalService) ProcessMultiModalContent(ctx context.Context, content interface{}) (interface{}, error) {
	// 模拟处理逻辑
	return map[string]interface{}{
		"processed": true,
		"content":   content,
		"timestamp": time.Now(),
	}, nil
}

// AnalyzeContent 分析内容
func (m *MockCrossModalService) AnalyzeContent(ctx context.Context, content interface{}) (interface{}, error) {
	// 模拟分析逻辑
	return map[string]interface{}{
		"analyzed": true,
		"content":  content,
		"features": []string{"feature1", "feature2", "feature3"},
	}, nil
}

// Shutdown 关闭服务
func (m *MockCrossModalService) Shutdown(ctx context.Context) error {
	m.initialized = false
	return nil
}