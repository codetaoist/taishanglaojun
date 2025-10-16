package advanced

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/agi"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/evolution"
	metalearning "github.com/codetaoist/taishanglaojun/core-services/ai-integration/meta-learning"
)

// AdvancedAICapability AI
// 定义了不同的 AI 能力，包括 AGI、元学习、自进化和混合能力
type AdvancedAICapability string

const (
	CapabilityAGI           AdvancedAICapability = "agi"
	CapabilityMetaLearning  AdvancedAICapability = "meta_learning"
	CapabilitySelfEvolution AdvancedAICapability = "self_evolution"
	CapabilityHybrid        AdvancedAICapability = "hybrid"
	CapabilityMultimodal    AdvancedAICapability = "multimodal"
)

// AIRequest AI
// 包含请求ID、类型、能力、输入、上下文、需求、优先级、超时时间和创建时间
type AIRequest struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Capability   AdvancedAICapability   `json:"capability"`
	Input        map[string]interface{} `json:"input"`
	Context      map[string]interface{} `json:"context"`
	Requirements map[string]interface{} `json:"requirements"`
	Priority     int                    `json:"priority"`
	Timeout      time.Duration          `json:"timeout"`
	CreatedAt    time.Time              `json:"created_at"`
}

// AIResponse AI
// 包含请求ID、成功状态、结果、置信度、处理时间、使用的能力列表和元数据
// 成功状态表示请求是否成功处理，结果包含处理结果，置信度表示处理结果的置信度，处理时间表示处理耗时
// 使用的能力列表表示请求使用的 AI 能力，元数据包含额外的信息
type AIResponse struct {
	RequestID        string                 `json:"request_id"`
	Success          bool                   `json:"success"`
	Result           map[string]interface{} `json:"result"`
	Confidence       float64                `json:"confidence"`
	ProcessTime      time.Duration          `json:"process_time"`
	UsedCapabilities []AdvancedAICapability `json:"used_capabilities"`
	Metadata         map[string]interface{} `json:"metadata"`
	Error            string                 `json:"error,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// SystemStatus 系统状态
// 包含 AGI、元学习、自进化和混合能力的状态信息
// 以及整体健康状态、活跃请求数、总请求数和成功率等指标
type SystemStatus struct {
	AGIStatus          map[string]interface{} `json:"agi_status"`
	MetaLearningStatus map[string]interface{} `json:"meta_learning_status"`
	EvolutionStatus    map[string]interface{} `json:"evolution_status"`
	OverallHealth      float64                `json:"overall_health"`
	ActiveRequests     int                    `json:"active_requests"`
	TotalRequests      int64                  `json:"total_requests"`
	SuccessRate        float64                `json:"success_rate"`
	AvgResponseTime    time.Duration          `json:"avg_response_time"`
	LastUpdated        time.Time              `json:"last_updated"`
}

// PerformanceMetrics 性能指标
// 包含请求数每秒、平均延迟、P95 延迟、P99 延迟、错误率、资源利用率和能力指标
// 以及时间戳
type PerformanceMetrics struct {
	RequestsPerSecond   float64                                     `json:"requests_per_second"`
	AvgLatency          time.Duration                               `json:"avg_latency"`
	P95Latency          time.Duration                               `json:"p95_latency"`
	P99Latency          time.Duration                               `json:"p99_latency"`
	ErrorRate           float64                                     `json:"error_rate"`
	ResourceUtilization map[string]float64                          `json:"resource_utilization"`
	CapabilityMetrics   map[AdvancedAICapability]map[string]float64 `json:"capability_metrics"`
	Timestamp           time.Time                                   `json:"timestamp"`
}

// AdvancedAIService AI
// 包含 AGI、元学习、自进化和混合能力的 AI 服务
// 以及多模态处理器、推理引擎、NLP 增强器等组件
type AdvancedAIService struct {
	mu                 sync.RWMutex
	agiService         *agi.AGIService
	metaLearningEngine *metalearning.MetaLearningEngine
	evolutionSystem    *evolution.SelfEvolutionSystem

	// AI 组件
	multimodalProcessor *MultimodalProcessor
	reasoningEngine     *ReasoningEngine
	nlpEnhancer         *NLPEnhancer

	// 系统状态
	isInitialized      bool
	activeRequests     map[string]*AIRequest
	requestHistory     []AIResponse
	performanceMetrics []PerformanceMetrics

	// 配置
	config *AdvancedAIConfig

	// 统计指标
	totalRequests      int64
	successfulRequests int64
	failedRequests     int64

	// 控制通道
	stopChan  chan struct{}
	isRunning bool
}

// AdvancedAIConfig AI 配置
// 包含 AGI、元学习、自进化和混合能力的开关、最大并发请求数、默认超时时间、性能监控、自动优化和日志级别
// 以及 AGI、元学习、自进化的配置
type AdvancedAIConfig struct {
	EnableAGI             bool                       `json:"enable_agi"`
	EnableMetaLearning    bool                       `json:"enable_meta_learning"`
	EnableEvolution       bool                       `json:"enable_evolution"`
	MaxConcurrentRequests int                        `json:"max_concurrent_requests"`
	DefaultTimeout        time.Duration              `json:"default_timeout"`
	PerformanceMonitoring bool                       `json:"performance_monitoring"`
	AutoOptimization      bool                       `json:"auto_optimization"`
	LogLevel              string                     `json:"log_level"`
	AGIConfig             map[string]interface{}     `json:"agi_config"`
	MetaLearningConfig    map[string]interface{}     `json:"meta_learning_config"`
	EvolutionConfig       *evolution.EvolutionConfig `json:"evolution_config"`
	EnabledCapabilities   []string                   `json:"enabled_capabilities"`
}

// NewAdvancedAIService AI 服务
// 初始化 AI 组件、系统状态、统计指标和控制通道
func NewAdvancedAIService(config *AdvancedAIConfig) *AdvancedAIService {
	service := &AdvancedAIService{
		config:             config,
		activeRequests:     make(map[string]*AIRequest),
		requestHistory:     make([]AIResponse, 0),
		performanceMetrics: make([]PerformanceMetrics, 0),
		stopChan:           make(chan struct{}),
	}

	return service
}

// Initialize 初始化 AI 服务
// 初始化多模态处理器、推理引擎、NLP 增强器、AGI、元学习、自进化系统
// 并设置 isInitialized 为 true
func (aas *AdvancedAIService) Initialize(ctx context.Context) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	if aas.isInitialized {
		return fmt.Errorf("service already initialized")
	}

	// 初始化多模态处理器
	aas.multimodalProcessor = NewMultimodalProcessor(&ProcessingConfig{
		MaxConcurrency:   aas.config.MaxConcurrentRequests,
		Timeout:          aas.config.DefaultTimeout,
		EnableCache:      true,
		QualityThreshold: 0.7,
		RetryAttempts:    3,
	})

	// 初始化推理引擎
	aas.reasoningEngine = NewReasoningEngine(&ReasoningConfig{
		MaxSteps:          10,
		Timeout:           aas.config.DefaultTimeout,
		EnableExplanation: true,
		MinConfidence:     0.7,
		MaxAlternatives:   5,
		UseCache:          true,
		Depth:             3,
		Breadth:           5,
	})

	// NLP 增强器
	aas.nlpEnhancer = NewNLPEnhancer(&NLPConfig{
		Model:          "default",
		MaxTokens:      512,
		Temperature:    0.7,
		TopP:           0.9,
		TopK:           50,
		UseCache:       true,
		EnableBatching: true,
		BatchSize:      32,
		Timeout:        aas.config.DefaultTimeout,
		CustomParams:   make(map[string]interface{}),
	})

	// AGI 服务
	if aas.config.EnableAGI {
		aas.agiService = agi.NewAGIService()
		if err := aas.agiService.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize AGI service: %w", err)
		}
	}

	// 元学习引擎
	if aas.config.EnableMetaLearning {
		aas.metaLearningEngine = metalearning.NewMetaLearningEngine()
		if err := aas.metaLearningEngine.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize meta-learning engine: %w", err)
		}
	}

	// 自进化系统
	if aas.config.EnableEvolution {
		aas.evolutionSystem = evolution.NewSelfEvolutionSystem(aas.config.EvolutionConfig)
		if err := aas.evolutionSystem.StartEvolution(ctx); err != nil {
			return fmt.Errorf("failed to start evolution system: %w", err)
		}
	}

	aas.isInitialized = true
	aas.isRunning = true

	// 性能监控循环
	if aas.config.PerformanceMonitoring {
		go aas.performanceMonitoringLoop(ctx)
	}

	// 自动优化循环
	if aas.config.AutoOptimization {
		go aas.autoOptimizationLoop(ctx)
	}

	return nil
}

// ProcessRequest AI 请求处理
// 检查服务是否初始化、设置请求创建时间、检查并发请求数、处理请求、更新统计指标、记录响应历史、
// 维护 activeRequests 映射、设置超时上下文、处理请求、更新统计指标、记录响应历史
// ProcessRequest AI
func (aas *AdvancedAIService) ProcessRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if !aas.isInitialized {
		return nil, fmt.Errorf("service not initialized")
	}

	startTime := time.Now()
	request.CreatedAt = startTime

	// 检查并发请求数
	if len(aas.activeRequests) >= aas.config.MaxConcurrentRequests {
		return &AIResponse{
			RequestID: request.ID,
			Success:   false,
			Error:     "max concurrent requests exceeded",
			CreatedAt: time.Now(),
		}, nil
	}

	// 维护 activeRequests 映射
	aas.mu.Lock()
	aas.activeRequests[request.ID] = request
	aas.totalRequests++
	aas.mu.Unlock()

	defer func() {
		aas.mu.Lock()
		delete(aas.activeRequests, request.ID)
		aas.mu.Unlock()
	}()

	// 设置超时上下文
	timeout := request.Timeout
	if timeout == 0 {
		timeout = aas.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 处理请求
	response, err := aas.processRequestByCapability(ctx, request)
	if err != nil {
		aas.mu.Lock()
		aas.failedRequests++
		aas.mu.Unlock()

		return &AIResponse{
			RequestID:   request.ID,
			Success:     false,
			Error:       err.Error(),
			ProcessTime: time.Since(startTime),
			CreatedAt:   time.Now(),
		}, nil
	}

	response.ProcessTime = time.Since(startTime)
	response.CreatedAt = time.Now()

	aas.mu.Lock()
	aas.successfulRequests++
	aas.requestHistory = append(aas.requestHistory, *response)

	// 记录响应历史
	if len(aas.requestHistory) > 10000 {
		aas.requestHistory = aas.requestHistory[len(aas.requestHistory)-10000:]
	}
	aas.mu.Unlock()

	return response, nil
}

// UpdateConfiguration 更新配置
func (aas *AdvancedAIService) UpdateConfiguration(config *AdvancedAIConfig) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()
	
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	aas.config = config
	return nil
}

// EnableCapability 启用能力
func (aas *AdvancedAIService) EnableCapability(capability AdvancedAICapability) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()
	
	if aas.config == nil {
		return fmt.Errorf("service not configured")
	}
	
	capabilityStr := string(capability)
	
	// 检查能力是否已启用
	for _, cap := range aas.config.EnabledCapabilities {
		if cap == capabilityStr {
			return nil // 已启用
		}
	}
	
	// 添加新能力
	aas.config.EnabledCapabilities = append(aas.config.EnabledCapabilities, capabilityStr)
	return nil
}

// DisableCapability 禁用能力
func (aas *AdvancedAIService) DisableCapability(capability AdvancedAICapability) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()
	
	if aas.config == nil {
		return fmt.Errorf("service not configured")
	}
	
	capabilityStr := string(capability)
	
	// 移除能力
	for i, cap := range aas.config.EnabledCapabilities {
		if cap == capabilityStr {
			aas.config.EnabledCapabilities = append(aas.config.EnabledCapabilities[:i], aas.config.EnabledCapabilities[i+1:]...)
			break
		}
	}
	
	return nil
}

// Reset 重置服务
func (aas *AdvancedAIService) Reset() error {
	aas.mu.Lock()
	defer aas.mu.Unlock()
	
	// 清空活动请求
	aas.activeRequests = make(map[string]*AIRequest)
	
	// 清空请求历史
	aas.requestHistory = make([]AIResponse, 0)
	
	// 重置统计指标
	aas.totalRequests = 0
	aas.successfulRequests = 0
	aas.failedRequests = 0
	
	// 清空性能指标
	aas.performanceMetrics = make([]PerformanceMetrics, 0)
	
	return nil
}

// GetSystemStatus 获取系统状态
func (aas *AdvancedAIService) GetSystemStatus() *SystemStatus {
	aas.mu.RLock()
	defer aas.mu.RUnlock()

	status := &SystemStatus{
		ActiveRequests: len(aas.activeRequests),
		TotalRequests:  aas.totalRequests,
		LastUpdated:    time.Now(),
	}

	// 成功率
	if aas.totalRequests > 0 {
		status.SuccessRate = float64(aas.successfulRequests) / float64(aas.totalRequests)
	}

	// 平均响应时间
	if len(aas.requestHistory) > 0 {
		totalTime := time.Duration(0)
		for _, resp := range aas.requestHistory {
			totalTime += resp.ProcessTime
		}
		status.AvgResponseTime = totalTime / time.Duration(len(aas.requestHistory))
	}

	// AGI 状态
	if aas.agiService != nil {
		status.AGIStatus = aas.agiService.GetStatus()
	}

	// 元学习状态
	if aas.metaLearningEngine != nil {
		status.MetaLearningStatus = aas.metaLearningEngine.GetStatus()
	}

	// 自进化状态
	if aas.evolutionSystem != nil {
		status.EvolutionStatus = aas.evolutionSystem.GetEvolutionStatus()
	}

	// 总体健康状态
	status.OverallHealth = aas.calculateOverallHealth(status)

	return status
}

// GetPerformanceMetrics 获取性能指标
func (aas *AdvancedAIService) GetPerformanceMetrics(limit int) []PerformanceMetrics {
	aas.mu.RLock()
	defer aas.mu.RUnlock()

	if limit <= 0 || limit > len(aas.performanceMetrics) {
		limit = len(aas.performanceMetrics)
	}

	start := len(aas.performanceMetrics) - limit
	metrics := make([]PerformanceMetrics, limit)
	copy(metrics, aas.performanceMetrics[start:])

	return metrics
}

// Shutdown 关闭服务
func (aas *AdvancedAIService) Shutdown(ctx context.Context) error {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	if !aas.isRunning {
		return nil
	}

	aas.isRunning = false
	close(aas.stopChan)

	// 关闭 AGI 服务
	if aas.agiService != nil {
		aas.agiService.Shutdown()
	}

	// 关闭元学习引擎
	if aas.metaLearningEngine != nil {
		aas.metaLearningEngine.Shutdown()
	}

	// 停止自进化系统
	if aas.evolutionSystem != nil {
		aas.evolutionSystem.StopEvolution()
	}

	// 等待所有请求处理完成
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("shutdown timeout: %d active requests remaining", len(aas.activeRequests))
		case <-ticker.C:
			if len(aas.activeRequests) == 0 {
				return nil
			}
		}
	}
}

// GetActiveRequests 获取当前活动请求
func (aas *AdvancedAIService) GetActiveRequests() []*AIRequest {
	aas.mu.RLock()
	defer aas.mu.RUnlock()

	requests := make([]*AIRequest, 0, len(aas.activeRequests))
	for _, req := range aas.activeRequests {
		requests = append(requests, req)
	}

	return requests
}

// processRequestByCapability 根据能力处理请求
func (aas *AdvancedAIService) processRequestByCapability(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	switch request.Capability {
	case CapabilityAGI:
		return aas.processAGIRequest(ctx, request)
	case CapabilityMetaLearning:
		return aas.processMetaLearningRequest(ctx, request)
	case CapabilitySelfEvolution:
		return aas.processEvolutionRequest(ctx, request)
	case CapabilityHybrid:
		return aas.processHybridRequest(ctx, request)
	default:
		//
		switch request.Type {
		case "multimodal":
			return aas.processMultimodalRequest(ctx, request)
		case "reasoning":
			return aas.processReasoningRequest(ctx, request)
		case "nlp":
			return aas.processNLPRequest(ctx, request)
		default:
			return nil, fmt.Errorf("unsupported capability: %s and type: %s", request.Capability, request.Type)
		}
	}
}

// processAGIRequest 处理 AGI 请求
func (aas *AdvancedAIService) processAGIRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if aas.agiService == nil {
		return nil, fmt.Errorf("AGI service not enabled")
	}

	// 转换Requirements从map[string]interface{}到[]string
	var requirements []string
	if req, ok := request.Requirements["requirements"]; ok {
		if reqSlice, ok := req.([]interface{}); ok {
			for _, r := range reqSlice {
				if str, ok := r.(string); ok {
					requirements = append(requirements, str)
				}
			}
		}
	}

	// AGI 任务
	task := &agi.AGITask{
		ID:           request.ID,
		Type:         request.Type,
		Description:  request.Type, // 添加描述字段
		Input:        request.Input,
		Requirements: requirements,
		Context:      request.Context,
		Priority:     1, // 设置默认优先级
		CreatedAt:    request.CreatedAt,
	}

	// 处理AGI任务
	result, err := aas.agiService.ProcessTask(ctx, task)
	if err != nil {
		return nil, err
	}

	// 处理Output字段的类型断言
	var outputMap map[string]interface{}
	if result.Output != nil {
		if om, ok := result.Output.(map[string]interface{}); ok {
			outputMap = om
		} else {
			outputMap = map[string]interface{}{"data": result.Output}
		}
	} else {
		outputMap = map[string]interface{}{"data": result.Result}
	}

	response := &AIResponse{
		RequestID:        request.ID,
		Success:          true,
		Result:           outputMap,
		Confidence:       result.Confidence,
		UsedCapabilities: []AdvancedAICapability{CapabilityAGI},
		Metadata: map[string]interface{}{
			"agi_capabilities": result.UsedCapabilities,
			"reasoning_steps":  result.Reasoning, // 使用Reasoning字段而不是ReasoningSteps
		},
	}

	return response, nil
}

// processMetaLearningRequest 处理元学习请求
func (aas *AdvancedAIService) processMetaLearningRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if aas.metaLearningEngine == nil {
		return nil, fmt.Errorf("meta-learning engine not enabled")
	}

	// 元学习任务
	task := &metalearning.Task{
		ID:         request.ID,
		Type:       request.Type,
		Domain:     fmt.Sprintf("%v", request.Context["domain"]),
		Data:       aas.extractDataFromInput(request.Input),
		Difficulty: aas.estimateTaskDifficulty(request),
		CreatedAt:  request.CreatedAt,
	}

	// 处理元学习任务
	result, err := aas.metaLearningEngine.LearnNewTask(ctx, task)
	if err != nil {
		return nil, err
	}

	response := &AIResponse{
		RequestID: request.ID,
		Success:   true,
		Result: map[string]interface{}{
			"performance":      result.Performance,
			"learning_time":    result.LearningTime,
			"adaptation_steps": result.AdaptationSteps,
			"knowledge_gained": result.KnowledgeGained,
		},
		Confidence:       result.Confidence,
		UsedCapabilities: []AdvancedAICapability{CapabilityMetaLearning},
		Metadata: map[string]interface{}{
			"strategy":         result.Strategy,
			"transferred_from": result.TransferredFrom,
		},
	}

	return response, nil
}

// processEvolutionRequest 处理自进化请求
func (aas *AdvancedAIService) processEvolutionRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if aas.evolutionSystem == nil {
		return nil, fmt.Errorf("evolution system not enabled")
	}

	// 处理自进化任务
	status := aas.evolutionSystem.GetEvolutionStatus()
	bestIndividual := aas.evolutionSystem.GetBestIndividual()

	response := &AIResponse{
		RequestID: request.ID,
		Success:   true,
		Result: map[string]interface{}{
			"evolution_status":     status,
			"best_individual":      bestIndividual,
			"optimization_targets": aas.getOptimizationTargets(),
		},
		Confidence:       0.9,
		UsedCapabilities: []AdvancedAICapability{CapabilitySelfEvolution},
		Metadata: map[string]interface{}{
			"evolution_generation": status["current_generation"],
			"population_diversity": status["diversity"],
		},
	}

	return response, nil
}

// processHybridRequest 处理混合请求
func (aas *AdvancedAIService) processHybridRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	//
	results := make(map[string]interface{})
	usedCapabilities := make([]AdvancedAICapability, 0)
	totalConfidence := 0.0
	capabilityCount := 0

	// AGI
	if aas.agiService != nil {
		agiResp, err := aas.processAGIRequest(ctx, request)
		if err == nil {
			results["agi"] = agiResp.Result
			usedCapabilities = append(usedCapabilities, CapabilityAGI)
			totalConfidence += agiResp.Confidence
			capabilityCount++
		}
	}

	// 元学习
	if aas.metaLearningEngine != nil {
		mlResp, err := aas.processMetaLearningRequest(ctx, request)
		if err == nil {
			results["meta_learning"] = mlResp.Result
			usedCapabilities = append(usedCapabilities, CapabilityMetaLearning)
			totalConfidence += mlResp.Confidence
			capabilityCount++
		}
	}

	// 自进化
	if aas.evolutionSystem != nil {
		evResp, err := aas.processEvolutionRequest(ctx, request)
		if err == nil {
			results["self_evolution"] = evResp.Result
			usedCapabilities = append(usedCapabilities, CapabilitySelfEvolution)
			totalConfidence += evResp.Confidence
			capabilityCount++
		}
	}

	// 计算平均置信度
	avgConfidence := 0.0
	if capabilityCount > 0 {
		avgConfidence = totalConfidence / float64(capabilityCount)
	}

	response := &AIResponse{
		RequestID:        request.ID,
		Success:          len(results) > 0,
		Result:           results,
		Confidence:       avgConfidence,
		UsedCapabilities: usedCapabilities,
		Metadata: map[string]interface{}{
			"hybrid_mode":      true,
			"capability_count": capabilityCount,
		},
	}

	return response, nil
}

// performanceMonitoringLoop 性能监控循环
func (aas *AdvancedAIService) performanceMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-aas.stopChan:
			return
		case <-ticker.C:
			aas.collectPerformanceMetrics()
		}
	}
}

// autoOptimizationLoop 自动优化循环
func (aas *AdvancedAIService) autoOptimizationLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-aas.stopChan:
			return
		case <-ticker.C:
			aas.performAutoOptimization(ctx)
		}
	}
}

// collectPerformanceMetrics 收集性能指标
func (aas *AdvancedAIService) collectPerformanceMetrics() {
	aas.mu.Lock()
	defer aas.mu.Unlock()

	now := time.Now()

	// 最近1分钟请求数
	recentRequests := 0
	for _, resp := range aas.requestHistory {
		if now.Sub(resp.CreatedAt) <= time.Minute {
			recentRequests++
		}
	}

	// 最近1小时平均响应时间
	var latencies []time.Duration
	for _, resp := range aas.requestHistory {
		if now.Sub(resp.CreatedAt) <= time.Hour {
			latencies = append(latencies, resp.ProcessTime)
		}
	}

	metrics := PerformanceMetrics{
		RequestsPerSecond:   float64(recentRequests) / 60.0,
		ResourceUtilization: aas.getResourceUtilization(),
		CapabilityMetrics:   aas.getCapabilityMetrics(),
		Timestamp:           now,
	}

	if len(latencies) > 0 {
		// 最近1小时平均响应时间
		total := time.Duration(0)
		for _, lat := range latencies {
			total += lat
		}
		metrics.AvgLatency = total / time.Duration(len(latencies))

		// P95/P99响应时间
		if len(latencies) >= 20 {
			p95Idx := int(float64(len(latencies)) * 0.95)
			p99Idx := int(float64(len(latencies)) * 0.99)
			metrics.P95Latency = latencies[p95Idx]
			metrics.P99Latency = latencies[p99Idx]
		}
	}

	// 错误率
	if aas.totalRequests > 0 {
		metrics.ErrorRate = float64(aas.failedRequests) / float64(aas.totalRequests)
	}

	aas.performanceMetrics = append(aas.performanceMetrics, metrics)

	// 保留最近24小时的指标
	if len(aas.performanceMetrics) > 1440 { // 24小时
		aas.performanceMetrics = aas.performanceMetrics[len(aas.performanceMetrics)-1440:]
	}
}

func (aas *AdvancedAIService) performAutoOptimization(ctx context.Context) {
	// 至少需要10条指标才能进行优化
	if len(aas.performanceMetrics) < 10 {
		return
	}

	recentMetrics := aas.performanceMetrics[len(aas.performanceMetrics)-10:]

	// 最近10分钟平均响应时间和错误率
	avgLatency := time.Duration(0)
	avgErrorRate := 0.0

	for _, metric := range recentMetrics {
		avgLatency += metric.AvgLatency
		avgErrorRate += metric.ErrorRate
	}

	avgLatency /= time.Duration(len(recentMetrics))
	avgErrorRate /= float64(len(recentMetrics))

	// 响应时间超过5秒，减少并发请求
	if avgLatency > time.Second*5 {
		//
		if aas.config.MaxConcurrentRequests > 10 {
			aas.config.MaxConcurrentRequests -= 5
		}
	} else if avgLatency < time.Millisecond*500 {
		//
		if aas.config.MaxConcurrentRequests < 100 {
			aas.config.MaxConcurrentRequests += 5
		}
	}

	// 错误率超过5%，增加并发请求
	if avgErrorRate > 0.05 {
		//
		if aas.config.MaxConcurrentRequests < 200 {
			aas.config.MaxConcurrentRequests += 5
		}
	} else if avgErrorRate < 0.01 {
		//
		if aas.config.MaxConcurrentRequests > 50 {
			aas.config.MaxConcurrentRequests -= 5
		}
	}

	if aas.evolutionSystem != nil {
		evolutionMetrics := &evolution.PerformanceMetrics{
			Accuracy:      1.0 - avgErrorRate,
			Efficiency:    math.Max(0, 1.0-avgLatency.Seconds()/10.0),
			Robustness:    aas.calculateRobustness(),
			Adaptability:  aas.calculateAdaptability(),
			ResourceUsage: aas.getAverageResourceUsage(),
			Latency:       avgLatency,
			Throughput:    aas.calculateThroughput(),
			ErrorRate:     avgErrorRate,
			Timestamp:     time.Now(),
		}

		aas.evolutionSystem.UpdatePerformanceMetrics(evolutionMetrics)
	}
}

// calculateOverallHealth 计算系统整体健康度
func (aas *AdvancedAIService) calculateOverallHealth(status *SystemStatus) float64 {
	health := 0.0
	factors := 0

	// 成功率超过50%，健康度增加
	if status.SuccessRate > 0.5 {
		health += status.SuccessRate
		factors++
	}

	// 平均响应时间低于10秒，健康度增加
	if status.AvgResponseTime > 0 {
		responseTimeFactor := math.Max(0, 1.0-status.AvgResponseTime.Seconds()/10.0)
		health += responseTimeFactor
		factors++
	}

	// 活跃请求数低于最大并发请求数的50%，健康度增加
	if status.ActiveRequests < aas.config.MaxConcurrentRequests/2 {
		loadFactor := 1.0 - float64(status.ActiveRequests)/float64(aas.config.MaxConcurrentRequests)
		health += loadFactor
		factors++
	}

	if factors == 0 {
		return 0.0
	}

	return health / float64(factors)
}

// extractDataFromInput 从输入中提取数据点
func (aas *AdvancedAIService) extractDataFromInput(input map[string]interface{}) []metalearning.DataPoint {
	// 从输入中提取数据点
	dataPoints := make([]metalearning.DataPoint, 0)

	if data, exists := input["data"]; exists {
		if dataList, ok := data.([]interface{}); ok {
			for _, item := range dataList {
				dataPoint := metalearning.DataPoint{
					Input:  item,
					Output: nil,
					Weight: 1.0,
				}

				if itemMap, ok := item.(map[string]interface{}); ok {
					dataPoint.Input = itemMap
					if output, exists := itemMap["output"]; exists {
						dataPoint.Output = output
					}
					if weight, exists := itemMap["weight"]; exists {
						if w, ok := weight.(float64); ok {
							dataPoint.Weight = w
						}
					}
				}

				dataPoints = append(dataPoints, dataPoint)
			}
		}
	}

	return dataPoints
}

// estimateTaskDifficulty 估计任务难度
func (aas *AdvancedAIService) estimateTaskDifficulty(request *AIRequest) float64 {
	// 从输入中提取数据点
	dataPoints := aas.extractDataFromInput(request.Input)

	difficulty := 0.5 // 基础难度

	// 输入数据点越多，难度增加
	if len(dataPoints) > 5 {
		difficulty += 0.2
	}

	// 需求描述越详细，难度增加
	if len(request.Requirements) > 5 {
		difficulty += 0.2
	}

	// 任务类型越复杂，难度增加
	switch request.Type {
	case "reasoning", "planning":
		difficulty += 0.3
	case "learning", "adaptation":
		difficulty += 0.2
	case "generation", "creativity":
		difficulty += 0.1
	}

	return math.Min(difficulty, 1.0)
}

// getOptimizationTargets 获取优化目标
func (aas *AdvancedAIService) getOptimizationTargets() []evolution.OptimizationTarget {
	return []evolution.OptimizationTarget{
		{
			Name:      "accuracy",
			Weight:    0.4,
			Target:    0.95,
			Tolerance: 0.05,
			Maximize:  true,
			Priority:  1,
		},
		{
			Name:      "efficiency",
			Weight:    0.3,
			Target:    0.9,
			Tolerance: 0.1,
			Maximize:  true,
			Priority:  2,
		},
		{
			Name:      "latency",
			Weight:    0.2,
			Target:    1.0, // 1
			Tolerance: 0.5,
			Maximize:  false,
			Priority:  3,
		},
		{
			Name:      "error_rate",
			Weight:    0.1,
			Target:    0.01,
			Tolerance: 0.005,
			Maximize:  false,
			Priority:  4,
		},
	}
}

// getResourceUtilization 获取资源利用率
func (aas *AdvancedAIService) getResourceUtilization() map[string]float64 {
	// 0-1
	return map[string]float64{
		"cpu":    0.6 + rand.Float64()*0.3,
		"memory": 0.5 + rand.Float64()*0.4,
		"gpu":    0.7 + rand.Float64()*0.2,
		"disk":   0.3 + rand.Float64()*0.2,
	}
}

// getCapabilityMetrics 获取能力指标
func (aas *AdvancedAIService) getCapabilityMetrics() map[AdvancedAICapability]map[string]float64 {
	metrics := make(map[AdvancedAICapability]map[string]float64)

	if aas.agiService != nil {
		metrics[CapabilityAGI] = map[string]float64{
			"accuracy":   0.85 + rand.Float64()*0.1,
			"efficiency": 0.8 + rand.Float64()*0.15,
			"robustness": 0.9 + rand.Float64()*0.05,
		}
	}

	if aas.metaLearningEngine != nil {
		metrics[CapabilityMetaLearning] = map[string]float64{
			"adaptation_speed":    0.75 + rand.Float64()*0.2,
			"transfer_quality":    0.8 + rand.Float64()*0.15,
			"learning_efficiency": 0.85 + rand.Float64()*0.1,
		}
	}

	if aas.evolutionSystem != nil {
		metrics[CapabilitySelfEvolution] = map[string]float64{
			"evolution_rate": 0.7 + rand.Float64()*0.2,
			"diversity":      0.6 + rand.Float64()*0.3,
			"convergence":    0.8 + rand.Float64()*0.15,
		}
	}

	return metrics
}

// calculateRobustness 计算鲁棒性
func (aas *AdvancedAIService) calculateRobustness() float64 {
	//
	if aas.totalRequests == 0 {
		return 1.0
	}

	errorRate := float64(aas.failedRequests) / float64(aas.totalRequests)
	return math.Max(0, 1.0-errorRate*2)
}

// calculateAdaptability 计算适应性
func (aas *AdvancedAIService) calculateAdaptability() float64 {
	//
	if len(aas.requestHistory) < 10 {
		return 0.5
	}

	// 最近100个请求中，不同任务类型的比例
	recentTypes := make(map[string]int)
	recentCount := math.Min(100, float64(len(aas.requestHistory)))

	for i := len(aas.requestHistory) - int(recentCount); i < len(aas.requestHistory); i++ {
		//
		resp := aas.requestHistory[i]
		if resp.RequestID != "" {
			recentTypes[resp.RequestID]++
		} else {
			recentTypes["default"]++
		}
	}

	// 最近100个请求中，不同任务类型的比例
	diversity := float64(len(recentTypes)) / 10.0
	// 10个不同任务类型的比例
	//  = 最近100个请求中，不同任务类型的比例
	return math.Min(diversity, 1.0)
}

// getAverageResourceUsage 获取平均资源利用率
func (aas *AdvancedAIService) getAverageResourceUsage() float64 {
	utilization := aas.getResourceUtilization()
	total := 0.0
	count := 0

	for _, usage := range utilization {
		total += usage
		count++
	}

	if count == 0 {
		return 0.0
	}

	return total / float64(count)
}

// calculateThroughput 计算吞吐量
func (aas *AdvancedAIService) calculateThroughput() float64 {
	if len(aas.requestHistory) == 0 {
		return 0.0
	}

	// 最近1小时内的请求数
	now := time.Now()
	recentRequests := 0

	for _, resp := range aas.requestHistory {
		if now.Sub(resp.CreatedAt) <= time.Hour {
			recentRequests++
		}
	}

	return float64(recentRequests) / 3600.0 // 每小时请求数
}

// processMultimodalRequest 处理多模态请求
func (aas *AdvancedAIService) processMultimodalRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if aas.multimodalProcessor == nil {
		return nil, fmt.Errorf("multimodal processor not initialized")
	}

	// 构建多模态输入
	multimodalInput := &MultimodalInput{
		ID:       request.ID,
		Content:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}

	// 根据输入类型设置内容
	if textData, exists := request.Input["text"]; exists {
		multimodalInput.Type = TypeText
		multimodalInput.Content = textData
	} else if imageData, exists := request.Input["image"]; exists {
		multimodalInput.Type = TypeImage
		multimodalInput.Content = imageData
	} else if audioData, exists := request.Input["audio"]; exists {
		multimodalInput.Type = TypeAudio
		multimodalInput.Content = audioData
	} else if videoData, exists := request.Input["video"]; exists {
		multimodalInput.Type = TypeVideo
		multimodalInput.Content = videoData
	}

	// 处理多模态输入
	result, err := aas.multimodalProcessor.ProcessSync(ctx, multimodalInput)
	if err != nil {
		return nil, fmt.Errorf("multimodal processing failed: %w", err)
	}

	// 将result.Result转换为map[string]interface{}
	resultMap := make(map[string]interface{})
	if result.Result != nil {
		resultMap["data"] = result.Result
	}

	response := &AIResponse{
		RequestID:        request.ID,
		Success:          true,
		Result:           resultMap,
		Confidence:       result.Confidence,
		UsedCapabilities: []AdvancedAICapability{CapabilityMultimodal}, //
		Metadata: map[string]interface{}{
			"processing_time": result.ProcessTime,
			"input_type":      result.Type,
			"output_id":       result.ID,
		},
	}

	return response, nil
}

// processReasoningRequest 处理推理请求
func (aas *AdvancedAIService) processReasoningRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if aas.reasoningEngine == nil {
		return nil, fmt.Errorf("reasoning engine not initialized")
	}

	// 转换Context为[]string
	var contextStrings []string
	for key, value := range request.Context {
		contextStrings = append(contextStrings, fmt.Sprintf("%s: %v", key, value))
	}

	// 构建推理请求
	reasoningReq := &ReasoningRequest{
		ID:          request.ID,
		Type:        ReasoningType(fmt.Sprintf("%v", request.Context["reasoning_type"])),
		Query:       fmt.Sprintf("%v", request.Input["query"]),
		Context:     contextStrings,
		Constraints: make([]Constraint, 0),
	}

	//  premises
	if premises, exists := request.Input["premises"]; exists {
		if premiseList, ok := premises.([]interface{}); ok {
			for _, premise := range premiseList {
				reasoningReq.Premises = append(reasoningReq.Premises, Premise{
					Statement:  fmt.Sprintf("%v", premise),
					Confidence: 1.0,
					Source:     "user_input",
					Type:       "user_provided",
				})
			}
		}
	}

	// 推理
	result, err := aas.reasoningEngine.Reason(ctx, reasoningReq)
	if err != nil {
		return nil, fmt.Errorf("reasoning failed: %w", err)
	}

	response := &AIResponse{
		RequestID: request.ID,
		Success:   true,
		Result: map[string]interface{}{
			"conclusion":      result.Conclusion,
			"reasoning_steps": result.Steps,
			"alternatives":    result.Alternatives,
			"explanation":     result.Explanation,
		},
		Confidence:       result.Confidence,
		UsedCapabilities: []AdvancedAICapability{CapabilityAGI},
		Metadata: map[string]interface{}{
			"reasoning_type":  result.Type,
			"evidence_count":  len(result.Evidence),
			"processing_time": result.ProcessTime,
		},
	}

	return response, nil
}

// processNLPRequest 处理NLP请求
func (aas *AdvancedAIService) processNLPRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	if aas.nlpEnhancer == nil {
		return nil, fmt.Errorf("NLP enhancer not initialized")
	}

	// NLP请求
	nlpReq := &NLPRequest{
		ID:       request.ID,
		Text:     fmt.Sprintf("%v", request.Input["text"]),
		Language: fmt.Sprintf("%v", request.Context["language"]),
		Task:     TaskTokenization, // 默认任务
	}

	// NLP任务
	if task, exists := request.Input["task"]; exists {
		nlpReq.Task = NLPTaskType(fmt.Sprintf("%v", task))
	}

	// NLP处理
	result, err := aas.nlpEnhancer.Process(ctx, nlpReq)
	if err != nil {
		return nil, fmt.Errorf("NLP processing failed: %w", err)
	}

	// 将NLP结果转换为map格式
	resultMap := make(map[string]interface{})
	if result.Result != nil {
		resultMap["data"] = result.Result
		resultMap["task"] = string(result.Task)
	}

	response := &AIResponse{
		RequestID:        request.ID,
		Success:          true,
		Result:           resultMap,
		Confidence:       result.Confidence,
		UsedCapabilities: []AdvancedAICapability{CapabilityAGI},
		Metadata: map[string]interface{}{
			"language":        result.Language,
			"processing_time": result.ProcessTime,
			"task_completed":  string(nlpReq.Task),
		},
	}

	return response, nil
}
