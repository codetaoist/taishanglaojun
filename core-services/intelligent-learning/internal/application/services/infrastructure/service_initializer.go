package infrastructure

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/crossmodal"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/knowledge"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics/realtime"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/adaptive"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/recommendation/content"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/infrastructure/config"
)



// ServiceConfigManager ?
type ServiceConfigManager struct {
	configPath   string                 `json:"config_path"`
	configs      map[string]interface{} `json:"configs"`
	lastModified time.Time              `json:"last_modified"`
	mu           sync.RWMutex           `json:"-"`
}

// NewServiceConfigManager ?
func NewServiceConfigManager(configPath string) *ServiceConfigManager {
	return &ServiceConfigManager{
		configPath: configPath,
		configs:    make(map[string]interface{}),
	}
}

// LoadConfig 
func (scm *ServiceConfigManager) LoadConfig() error {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	
	// ?
	// nil
	return nil
}

// ValidateConfig 
func (scm *ServiceConfigManager) ValidateConfig() error {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	
	// ?
	// nil
	return nil
}

// GetConfig 
func (scm *ServiceConfigManager) GetConfig() *config.GlobalServiceConfig {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	
	// 
	return &config.GlobalServiceConfig{}
}

// IsServiceEnabled ?
func (scm *ServiceConfigManager) IsServiceEnabled(serviceName string) bool {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	
	// true
	return true
}

// ServiceInitializer 
type ServiceInitializer struct {
	configManager *ServiceConfigManager
	services      map[string]interface{}
	integration   *IntelligentLearningServiceIntegration
	initialized   bool
	mu            sync.RWMutex
}

// InitializationResult ?
type InitializationResult struct {
	ServiceName   string        `json:"service_name"`
	Success       bool          `json:"success"`
	Error         error         `json:"error,omitempty"`
	InitTime      time.Duration `json:"init_time"`
	Dependencies  []string      `json:"dependencies"`
	Status        string        `json:"status"`
}

// ServiceDependency 
type ServiceDependency struct {
	ServiceName  string   `json:"service_name"`
	Dependencies []string `json:"dependencies"`
	Priority     int      `json:"priority"`
}

// NewServiceInitializer 
func NewServiceInitializer(configPath string) *ServiceInitializer {
	return &ServiceInitializer{
		configManager: NewServiceConfigManager(configPath),
		services:      make(map[string]interface{}),
		initialized:   false,
	}
}

// Initialize ?
func (si *ServiceInitializer) Initialize(ctx context.Context) ([]*InitializationResult, error) {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if si.initialized {
		return nil, fmt.Errorf("services already initialized")
	}
	
	log.Println("Starting intelligent learning services initialization...")
	
	// 
	if err := si.configManager.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	// 
	if err := si.configManager.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	config := si.configManager.GetConfig()
	
	// 
	dependencies := si.getServiceDependencies()
	
	// 
	results := make([]*InitializationResult, 0)
	
	for _, dep := range dependencies {
		if si.configManager.IsServiceEnabled(dep.ServiceName) {
			result := si.initializeService(ctx, dep.ServiceName, config)
			results = append(results, result)
			
			if !result.Success {
				log.Printf("Failed to initialize service %s: %v", dep.ServiceName, result.Error)
				// ?
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
	
	// ?
	if err := si.initializeIntegration(ctx, config); err != nil {
		return results, fmt.Errorf("failed to initialize service integration: %w", err)
	}
	
	si.initialized = true
	log.Println("All services initialized successfully")
	
	return results, nil
}

// getServiceDependencies 
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

// initializeService ?
func (si *ServiceInitializer) initializeService(ctx context.Context, serviceName string, config *config.GlobalServiceConfig) *InitializationResult {
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

// initializeCrossModalService ?
func (si *ServiceInitializer) initializeCrossModalService(ctx context.Context, config *config.CrossModalServiceConfig) (crossmodal.CrossModalServiceInterface, error) {
	log.Println("Initializing Cross Modal Service...")
	
	// 
	crossModalConfig := &crossmodal.CrossModalServiceConfig{
		APIEndpoint:  "http://localhost:8080/api/crossmodal",
		APIKey:       "mock-api-key",
		Timeout:      config.Timeout,
		MaxRetries:   3,
		EnableCache:  true,
		CacheExpiry:  1 * time.Hour,
		ModelVersion: "v1.0",
		BatchSize:    10,
	}
	
	// ?
	service := crossmodal.NewCrossModalServiceImpl(crossModalConfig)
	
	// ?
	if service == nil {
		return nil, fmt.Errorf("failed to create cross modal service instance")
	}
	
	log.Println("Cross Modal Service initialized successfully")
	return service, nil
}

// initializeRelationInferenceEngine ?
func (si *ServiceInitializer) initializeRelationInferenceEngine(ctx context.Context, config *config.RelationInferenceEngineConfig) (*knowledge.IntelligentRelationInferenceEngine, error) {
	log.Printf("Initializing Relation Inference Engine with config: %+v", config)
	
	// ?
	crossModalService, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// ?
	crossModalAdapter := NewCrossModalServiceAdapter(crossModalService.(crossmodal.CrossModalServiceInterface))
	
	// ?
	engine := knowledge.NewIntelligentRelationInferenceEngine(
		crossModalAdapter,
	)
	
	return engine, nil
}

// initializeAdaptiveLearningEngine 
func (si *ServiceInitializer) initializeAdaptiveLearningEngine(ctx context.Context, config *config.AdaptiveLearningEngineConfig) (*adaptive.AdaptiveLearningEngine, error) {
	log.Printf("Initializing Adaptive Learning Engine with config: %+v", config)
	
	// ?
	crossModalService, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// ?
	inferenceEngine, exists := si.services["relation_inference"]
	if !exists {
		return nil, fmt.Errorf("relation inference engine dependency not found")
	}
	
	// 
	realtimeAnalytics, exists := si.services["realtime_analytics"]
	if !exists {
		return nil, fmt.Errorf("realtime analytics service dependency not found")
	}
	
	knowledgeGraph, exists := si.services["knowledge_graph"]
	if !exists {
		return nil, fmt.Errorf("knowledge graph service dependency not found")
	}
	
	// ?
	crossModalAdapter := NewCrossModalServiceAdapter(crossModalService.(crossmodal.CrossModalServiceInterface))
	
	// 
	engine := adaptive.NewAdaptiveLearningEngine(
		crossModalAdapter,
		inferenceEngine.(*knowledge.IntelligentRelationInferenceEngine),
		realtimeAnalytics.(*realtime.RealtimeLearningAnalyticsService),
		knowledgeGraph.(*knowledge.AutomatedKnowledgeGraphService),
	)
	
	return engine, nil
}

// initializeRealtimeLearningAnalyticsService ?
func (si *ServiceInitializer) initializeRealtimeLearningAnalyticsService(ctx context.Context, config *config.RealtimeLearningAnalyticsServiceConfig) (*realtime.RealtimeLearningAnalyticsServiceImpl, error) {
	log.Printf("Initializing Realtime Learning Analytics Service with config: %+v", config)
	
	// ?
	_, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// ?
	service := realtime.NewRealtimeLearningAnalyticsServiceImpl(config)
	
	// ?
	if service == nil {
		return nil, fmt.Errorf("failed to create realtime learning analytics service instance")
	}
	
	log.Println("Realtime Learning Analytics Service initialized successfully")
	return service, nil
}

// initializeAutomatedKnowledgeGraphService 
func (si *ServiceInitializer) initializeAutomatedKnowledgeGraphService(ctx context.Context, config *config.AutomatedKnowledgeGraphServiceConfig) (*knowledge.AutomatedKnowledgeGraphServiceImpl, error) {
	log.Printf("Initializing Automated Knowledge Graph Service with config: %+v", config)
	
	// 
	_, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// 
	knowledgeConfig := &knowledge.AutomatedKnowledgeGraphServiceConfig{
		MaxNodes:                1000,
		MaxEdges:                5000,
		ExtractionTimeout:       30 * time.Second,
		InferenceTimeout:        60 * time.Second,
		CacheSize:               100,
		CacheTTL:                time.Hour,
		EnableParallelProcessing: true,
		MaxConcurrency:          4,
		Metadata:                make(map[string]interface{}),
	}
	
	// 
	service := knowledge.NewAutomatedKnowledgeGraphServiceImpl(knowledgeConfig)
	
	// ?
	if service == nil {
		return nil, fmt.Errorf("failed to create automated knowledge graph service instance")
	}
	
	log.Println("Automated Knowledge Graph Service initialized successfully")
	return service, nil
}

// initializeLearningAnalyticsReportingService ?
func (si *ServiceInitializer) initializeLearningAnalyticsReportingService(ctx context.Context, config *config.LearningAnalyticsReportingServiceConfig) (*analytics.LearningAnalyticsReportingService, error) {
	log.Printf("Initializing Learning Analytics Reporting Service with config: %+v", config)
	
	// 
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
	
	knowledgeGraph, exists := si.services["knowledge_graph"]
	if !exists {
		return nil, fmt.Errorf("knowledge graph service dependency not found")
	}
	

	
	// 䲻
	adapter := NewCrossModalServiceAdapter(crossModalService.(crossmodal.CrossModalServiceInterface))
	
	// ?
	service := analytics.NewLearningAnalyticsReportingService(
		adapter,
		relationInference.(*knowledge.IntelligentRelationInferenceEngine),
		realtimeAnalytics.(*realtime.RealtimeLearningAnalyticsService),
		knowledgeGraph.(*knowledge.AutomatedKnowledgeGraphService),
	)
	
	return service, nil
}

// initializeIntelligentContentRecommendationService ?
func (si *ServiceInitializer) initializeIntelligentContentRecommendationService(ctx context.Context, config *config.IntelligentContentRecommendationServiceConfig) (*content.IntelligentContentRecommendationServiceImpl, error) {
	log.Printf("Initializing Intelligent Content Recommendation Service with config: %+v", config)
	
	// ?
	_, exists := si.services["cross_modal"]
	if !exists {
		return nil, fmt.Errorf("cross modal service dependency not found")
	}
	
	// 
	maxRecommendations := 10
	minScore := 0.6
	if config.RecommendationLimits != nil {
		maxRecommendations = config.RecommendationLimits.MaxRecommendations
		minScore = config.RecommendationLimits.MinConfidenceScore
	}
	
	recommendationConfig := &content.RecommendationConfig{
		MaxRecommendations: maxRecommendations,
		MinScore:          minScore,
		DiversityWeight:   0.3,
		NoveltyWeight:     0.2,
		RealtimeEnabled:   true,
		CacheEnabled:      true,
		ExplanationEnabled: true,
		Config:            make(map[string]interface{}),
	}
	
	// ?
	service := content.NewIntelligentContentRecommendationServiceImpl(recommendationConfig)
	
	// ?
	if service == nil {
		return nil, fmt.Errorf("failed to create intelligent content recommendation service instance")
	}
	
	log.Println("Intelligent Content Recommendation Service initialized successfully")
	return service, nil
}

// initializeIntegration ?
func (si *ServiceInitializer) initializeIntegration(ctx context.Context, config *config.GlobalServiceConfig) error {
	log.Println("Initializing service integration...")
	
	// 
	crossModalService, _ := si.services["cross_modal"]
	relationInference, _ := si.services["relation_inference"]
	adaptiveLearning, _ := si.services["adaptive_learning"]
	realtimeAnalytics, _ := si.services["realtime_analytics"]
	knowledgeGraph, _ := si.services["knowledge_graph"]
	analyticsReporting, _ := si.services["analytics_reporting"]
	contentRecommendation, _ := si.services["content_recommendation"]
	
	// 
	integrationConfig := &IntegrationConfig{
		ServiceConfig: &ServiceConfiguration{
			EnableCrossModalService:                    true,
			EnableRelationInferenceEngine:              true,
			EnableAdaptiveLearningEngine:               true,
			EnableRealtimeLearningAnalyticsService:     true,
			EnableAutomatedKnowledgeGraphService:       true,
			EnableLearningAnalyticsReportingService:    true,
			EnableIntelligentContentRecommendationService: true,
		},
		IntegrationSettings: &LearningIntegrationSettings{
			EnableServiceOrchestration: true,
			EnableDataSynchronization:  true,
			EnableEventDrivenUpdates: true,
		},
		PerformanceConfig: &PerformanceConfiguration{
			MaxConcurrentRequests: 10,
			RequestTimeout: 30 * time.Second,
			CacheExpiration:   time.Hour,
			BatchProcessingSize: 100,
		},
		SecurityConfig: &SecurityConfiguration{
			EnableAuthentication: true,
			EnableAuthorization:  true,
			EnableEncryption:     true,
		},
		MonitoringConfig: &MonitoringConfiguration{
			EnableMetrics: true,
			EnableTracing: true,
			EnableLogging: true,
		},
	}
	
	// 
	si.integration = NewIntelligentLearningServiceIntegration(
		crossModalService.(crossmodal.CrossModalServiceInterface),
		relationInference.(*knowledge.IntelligentRelationInferenceEngine),
		adaptiveLearning.(*adaptive.AdaptiveLearningEngine),
		realtimeAnalytics.(*realtime.RealtimeLearningAnalyticsService),
		knowledgeGraph.(*knowledge.AutomatedKnowledgeGraphService),
		analyticsReporting.(*analytics.LearningAnalyticsReportingService),
		contentRecommendation.(*content.IntelligentContentRecommendationService),
		integrationConfig,
	)
	
	log.Println("Service integration initialized successfully")
	return nil
}

// isServiceCritical 
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

// GetService 
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

// GetIntegration 
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

// Shutdown ?
func (si *ServiceInitializer) Shutdown(ctx context.Context) error {
	si.mu.Lock()
	defer si.mu.Unlock()
	
	if !si.initialized {
		return fmt.Errorf("services not initialized")
	}
	
	log.Println("Shutting down intelligent learning services...")
	
	// 
	if si.integration != nil {
		if err := si.integration.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down service integration: %v", err)
		}
	}
	
	// ?
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

// IsInitialized ?
func (si *ServiceInitializer) IsInitialized() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.initialized
}

// GetInitializationStatus ?
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

// MockCrossModalService ?
type MockCrossModalService struct {
	config      *config.CrossModalServiceConfig
	initialized bool
}

// ProcessCrossModalInference ?(crossmodal?
func (m *MockCrossModalService) ProcessCrossModalInference(ctx context.Context, req *crossmodal.CrossModalInferenceRequest) (*crossmodal.CrossModalInferenceResponse, error) {
	// 
	return &crossmodal.CrossModalInferenceResponse{
		Success: true,
		Result: map[string]interface{}{
			"inference_type": req.Type,
			"processed":      true,
			"data":          req.Data,
			"timestamp":     time.Now(),
		},
		Confidence: 0.85,
		Metadata: map[string]interface{}{
			"mock_service": true,
		},
	}, nil
}

// ProcessCrossModalInferenceKnowledge ?(knowledge?
func (m *MockCrossModalService) ProcessCrossModalInferenceKnowledge(ctx context.Context, req *knowledge.CrossModalInferenceRequest) (*knowledge.CrossModalInferenceResponse, error) {
	// 
	return &knowledge.CrossModalInferenceResponse{
		Success: true,
		Result: map[string]interface{}{
			"inference_type": req.Type,
			"processed":      true,
			"data":          req.Data,
			"timestamp":     time.Now(),
		},
	}, nil
}

// ProcessMultiModalContent ?
func (m *MockCrossModalService) ProcessMultiModalContent(ctx context.Context, content interface{}) (interface{}, error) {
	// 
	return map[string]interface{}{
		"processed": true,
		"content":   content,
		"timestamp": time.Now(),
	}, nil
}

// AnalyzeContent 
func (m *MockCrossModalService) AnalyzeContent(ctx context.Context, content interface{}) (interface{}, error) {
	// 
	return map[string]interface{}{
		"analyzed": true,
		"content":  content,
		"features": []string{"feature1", "feature2", "feature3"},
	}, nil
}

// Shutdown 
func (m *MockCrossModalService) Shutdown(ctx context.Context) error {
	m.initialized = false
	return nil
}

// CrossModalServiceAdapter 䲻
type CrossModalServiceAdapter struct {
	crossModalService crossmodal.CrossModalServiceInterface
}

// NewCrossModalServiceAdapter ?
func NewCrossModalServiceAdapter(service crossmodal.CrossModalServiceInterface) *CrossModalServiceAdapter {
	return &CrossModalServiceAdapter{
		crossModalService: service,
	}
}

// ProcessCrossModalInference knowledgeCrossModalServiceInterface
func (a *CrossModalServiceAdapter) ProcessCrossModalInference(ctx context.Context, req *knowledge.CrossModalInferenceRequest) (*knowledge.CrossModalInferenceResponse, error) {
	// knowledgecrossmodal
	crossModalReq := &crossmodal.CrossModalInferenceRequest{
		Type:      req.Type,
		Data:      req.Data,
		Options:   req.Options,
		Context:   req.Context,
		Timestamp: req.Timestamp,
	}
	
	// crossmodal
	crossModalResp, err := a.crossModalService.ProcessCrossModalInference(ctx, crossModalReq)
	if err != nil {
		return nil, err
	}
	
	// crossmodalknowledge
	return &knowledge.CrossModalInferenceResponse{
		Success: crossModalResp.Success,
		Result:  crossModalResp.Result,
	}, nil
}

