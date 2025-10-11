package engines

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/google/uuid"
)

// FusionEngine 碳基-硅基融合引擎
type FusionEngine struct {
	mu                sync.RWMutex
	sessions          map[string]*models.FusionState
	config            *FusionEngineConfig
	carbonProcessors  []CarbonProcessor
	siliconProcessors []SiliconProcessor
	fusionStrategies  map[string]FusionStrategy
	qualityEvaluator  QualityEvaluator
	emergenceDetector EmergenceDetector
	isInitialized     bool
}

// FusionEngineConfig 融合引擎配置
type FusionEngineConfig struct {
	MaxConcurrentSessions    int           `json:"max_concurrent_sessions"`
	SessionTimeout           time.Duration `json:"session_timeout"`
	QualityThreshold         float64       `json:"quality_threshold"`
	SynergyThreshold         float64       `json:"synergy_threshold"`
	EnableEmergenceDetection bool          `json:"enable_emergence_detection"`
	CarbonWeight             float64       `json:"carbon_weight"`
	SiliconWeight            float64       `json:"silicon_weight"`
	FusionStrategies         []string      `json:"fusion_strategies"`
}

// CarbonProcessor 碳基处理器接�?
type CarbonProcessor interface {
	ProcessEmotion(ctx context.Context, emotion *models.EmotionalState) (*models.CarbonInput, error)
	ProcessCulture(ctx context.Context, culture *models.CulturalContext) (*models.CarbonInput, error)
	ProcessIntuition(ctx context.Context, intuition *models.IntuitionData) (*models.CarbonInput, error)
	ProcessCreativity(ctx context.Context, creativity *models.CreativityData) (*models.CarbonInput, error)
	ProcessWisdom(ctx context.Context, wisdom *models.WisdomData) (*models.CarbonInput, error)
	GetProcessorType() models.CarbonInputType
}

// SiliconProcessor 硅基处理器接�?
type SiliconProcessor interface {
	ProcessComputation(ctx context.Context, data interface{}) (*models.SiliconInput, error)
	ProcessLogic(ctx context.Context, reasoning *models.LogicalReasoning) (*models.SiliconInput, error)
	ProcessData(ctx context.Context, processing *models.DataProcessing) (*models.SiliconInput, error)
	ProcessAlgorithm(ctx context.Context, analysis *models.AlgorithmicAnalysis) (*models.SiliconInput, error)
	GetProcessorType() models.SiliconInputType
}

// FusionStrategy 融合策略接口
type FusionStrategy interface {
	Fuse(ctx context.Context, carbon *models.CarbonInput, silicon *models.SiliconInput) (*models.FusionResult, error)
	GetStrategyName() string
	GetCompatibility(carbonType models.CarbonInputType, siliconType models.SiliconInputType) float64
}

// QualityEvaluator 质量评估器接�?
type QualityEvaluator interface {
	EvaluateQuality(ctx context.Context, result *models.FusionResult) (*models.QualityMetrics, error)
}

// EmergenceDetector 涌现检测器接口
type EmergenceDetector interface {
	DetectEmergence(ctx context.Context, carbon *models.CarbonInput, silicon *models.SiliconInput, result *models.FusionResult) ([]models.EmergentProperty, error)
}

// NewFusionEngine 创建新的融合引擎
func NewFusionEngine(config *FusionEngineConfig) *FusionEngine {
	return &FusionEngine{
		sessions:         make(map[string]*models.FusionState),
		config:           config,
		fusionStrategies: make(map[string]FusionStrategy),
		isInitialized:    false,
	}
}

// Initialize 初始化融合引�?
func (fe *FusionEngine) Initialize(ctx context.Context) error {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if fe.isInitialized {
		return nil
	}

	// 初始化默认处理器
	if err := fe.initializeProcessors(); err != nil {
		return fmt.Errorf("failed to initialize processors: %w", err)
	}

	// 初始化融合策�?
	if err := fe.initializeFusionStrategies(); err != nil {
		return fmt.Errorf("failed to initialize fusion strategies: %w", err)
	}

	// 初始化质量评估器
	if err := fe.initializeQualityEvaluator(); err != nil {
		return fmt.Errorf("failed to initialize quality evaluator: %w", err)
	}

	// 初始化涌现检测器
	if fe.config.EnableEmergenceDetection {
		if err := fe.initializeEmergenceDetector(); err != nil {
			return fmt.Errorf("failed to initialize emergence detector: %w", err)
		}
	}

	fe.isInitialized = true
	return nil
}

// StartFusion 开始融合过�?
func (fe *FusionEngine) StartFusion(ctx context.Context, carbon *models.CarbonInput, silicon *models.SiliconInput) (string, error) {
	if !fe.isInitialized {
		return "", models.ErrFusionEngineNotReady
	}

	fe.mu.Lock()
	defer fe.mu.Unlock()

	// 检查并发会话限�?
	if len(fe.sessions) >= fe.config.MaxConcurrentSessions {
		return "", models.ErrFusionEngineBusy
	}

	// 创建新的融合会话
	sessionID := uuid.New().String()
	fusionState := &models.FusionState{
		SessionID:    sessionID,
		CarbonInput:  carbon,
		SiliconInput: silicon,
		Status:       models.FusionStatusPending,
		Progress:     0.0,
		StartTime:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	fe.sessions[sessionID] = fusionState

	// 异步执行融合过程
	go fe.executeFusion(ctx, sessionID)

	return sessionID, nil
}

// GetFusionStatus 获取融合状�?
func (fe *FusionEngine) GetFusionStatus(sessionID string) (*models.FusionState, error) {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	state, exists := fe.sessions[sessionID]
	if !exists {
		return nil, models.ErrFusionSessionNotFound
	}

	return state, nil
}

// CancelFusion 取消融合过程
func (fe *FusionEngine) CancelFusion(sessionID string) error {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	state, exists := fe.sessions[sessionID]
	if !exists {
		return models.ErrFusionSessionNotFound
	}

	if state.Status == models.FusionStatusCompleted || state.Status == models.FusionStatusFailed {
		return models.ErrFusionSessionAlreadyFinished
	}

	state.Status = models.FusionStatusCancelled
	now := time.Now()
	state.EndTime = &now
	state.ProcessDuration = now.Sub(state.StartTime)

	return nil
}

// executeFusion 执行融合过程
func (fe *FusionEngine) executeFusion(ctx context.Context, sessionID string) {
	fe.mu.Lock()
	state := fe.sessions[sessionID]
	fe.mu.Unlock()

	if state == nil {
		return
	}

	// 更新状态为处理�?
	fe.updateFusionStatus(sessionID, models.FusionStatusProcessing, 0.1)

	// 选择最佳融合策�?
	strategy, err := fe.selectBestStrategy(state.CarbonInput, state.SiliconInput)
	if err != nil {
		fe.finalizeFusion(sessionID, nil, err)
		return
	}

	fe.updateFusionStatus(sessionID, models.FusionStatusProcessing, 0.3)

	// 执行融合
	result, err := strategy.Fuse(ctx, state.CarbonInput, state.SiliconInput)
	if err != nil {
		fe.finalizeFusion(sessionID, nil, err)
		return
	}

	fe.updateFusionStatus(sessionID, models.FusionStatusProcessing, 0.6)

	// 质量评估
	if fe.qualityEvaluator != nil {
		qualityMetrics, err := fe.qualityEvaluator.EvaluateQuality(ctx, result)
		if err == nil {
			result.QualityMetrics = qualityMetrics
		}
	}

	fe.updateFusionStatus(sessionID, models.FusionStatusProcessing, 0.8)

	// 涌现检�?
	if fe.emergenceDetector != nil && fe.config.EnableEmergenceDetection {
		emergentProperties, err := fe.emergenceDetector.DetectEmergence(ctx, state.CarbonInput, state.SiliconInput, result)
		if err == nil {
			result.EmergentProperties = emergentProperties
		}
	}

	fe.updateFusionStatus(sessionID, models.FusionStatusProcessing, 0.9)

	// 完成融合
	fe.finalizeFusion(sessionID, result, nil)
}

// selectBestStrategy 选择最佳融合策�?
func (fe *FusionEngine) selectBestStrategy(carbon *models.CarbonInput, silicon *models.SiliconInput) (FusionStrategy, error) {
	var bestStrategy FusionStrategy
	var bestCompatibility float64 = -1

	for _, strategy := range fe.fusionStrategies {
		compatibility := strategy.GetCompatibility(carbon.Type, silicon.Type)
		if compatibility > bestCompatibility {
			bestCompatibility = compatibility
			bestStrategy = strategy
		}
	}

	if bestStrategy == nil {
		return nil, models.ErrNoCompatibleFusionStrategy
	}

	return bestStrategy, nil
}

// updateFusionStatus 更新融合状�?
func (fe *FusionEngine) updateFusionStatus(sessionID string, status models.FusionStatus, progress float64) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	if state, exists := fe.sessions[sessionID]; exists {
		state.Status = status
		state.Progress = progress
	}
}

// finalizeFusion 完成融合过程
func (fe *FusionEngine) finalizeFusion(sessionID string, result *models.FusionResult, err error) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	state, exists := fe.sessions[sessionID]
	if !exists {
		return
	}

	now := time.Now()
	state.EndTime = &now
	state.ProcessDuration = now.Sub(state.StartTime)
	state.Progress = 1.0

	if err != nil {
		state.Status = models.FusionStatusFailed
		state.Metadata["error"] = err.Error()
	} else {
		state.Status = models.FusionStatusCompleted
		state.FusionResult = result
	}
}

// CleanupExpiredSessions 清理过期会话
func (fe *FusionEngine) CleanupExpiredSessions() {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	now := time.Now()
	for sessionID, state := range fe.sessions {
		if now.Sub(state.StartTime) > fe.config.SessionTimeout {
			delete(fe.sessions, sessionID)
		}
	}
}

// GetStatistics 获取引擎统计信息
func (fe *FusionEngine) GetStatistics() map[string]interface{} {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["total_sessions"] = len(fe.sessions)
	stats["is_initialized"] = fe.isInitialized

	statusCounts := make(map[models.FusionStatus]int)
	for _, state := range fe.sessions {
		statusCounts[state.Status]++
	}
	stats["status_counts"] = statusCounts

	return stats
}

// initializeProcessors 初始化处理器
func (fe *FusionEngine) initializeProcessors() error {
	// 初始化碳基处理器
	fe.carbonProcessors = []CarbonProcessor{
		NewEmotionProcessor(),
		NewCultureProcessor(),
		NewIntuitionProcessor(),
		NewCreativityProcessor(),
		NewWisdomProcessor(),
	}

	// 初始化硅基处理器
	fe.siliconProcessors = []SiliconProcessor{
		NewComputationProcessor(),
		NewLogicProcessor(),
		NewDataProcessor(),
		NewAlgorithmProcessor(),
	}

	return nil
}

func (fe *FusionEngine) initializeFusionStrategies() error {
	// 注册融合策略
	strategies := []FusionStrategy{
		NewComplementaryFusionStrategy(),
		NewSynergeticFusionStrategy(),
		NewHybridFusionStrategy(),
		NewTranscendentFusionStrategy(),
	}

	for _, strategy := range strategies {
		fe.fusionStrategies[strategy.GetStrategyName()] = strategy
	}

	return nil
}

func (fe *FusionEngine) initializeQualityEvaluator() error {
	fe.qualityEvaluator = NewDefaultQualityEvaluator()
	return nil
}

func (fe *FusionEngine) initializeEmergenceDetector() error {
	fe.emergenceDetector = NewDefaultEmergenceDetector()
	return nil
}

// 默认实现（简化版本）
type DefaultQualityEvaluator struct{}

func NewDefaultQualityEvaluator() *DefaultQualityEvaluator {
	return &DefaultQualityEvaluator{}
}

func (dqe *DefaultQualityEvaluator) EvaluateQuality(ctx context.Context, result *models.FusionResult) (*models.QualityMetrics, error) {
	// 基于融合结果计算质量指标
	metrics := &models.QualityMetrics{
		Accuracy:     calculateAccuracy(result),
		Relevance:    calculateRelevance(result),
		Completeness: calculateCompleteness(result),
		Coherence:    calculateCoherence(result),
		Creativity:   calculateCreativity(result),
		Practicality: calculatePracticality(result),
	}

	// 计算整体质量
	metrics.Overall = (metrics.Accuracy + metrics.Relevance + metrics.Completeness +
		metrics.Coherence + metrics.Creativity + metrics.Practicality) / 6.0

	return metrics, nil
}

type DefaultEmergenceDetector struct{}

func NewDefaultEmergenceDetector() *DefaultEmergenceDetector {
	return &DefaultEmergenceDetector{}
}

func (ded *DefaultEmergenceDetector) DetectEmergence(ctx context.Context, carbon *models.CarbonInput, silicon *models.SiliconInput, result *models.FusionResult) ([]models.EmergentProperty, error) {
	var properties []models.EmergentProperty

	// 检测认知涌�?
	if cognitiveStrength := detectCognitiveEmergence(carbon, silicon, result); cognitiveStrength > 0.5 {
		properties = append(properties, models.EmergentProperty{
			Name:        "Enhanced Cognitive Processing",
			Description: "Fusion resulted in enhanced cognitive capabilities beyond individual components",
			Strength:    cognitiveStrength,
			Type:        models.EmergentPropertyTypeCognitive,
			Evidence:    []string{"Improved reasoning patterns", "Novel insight generation"},
			Impact:      cognitiveStrength * 0.8,
		})
	}

	// 检测创造性涌�?
	if creativeStrength := detectCreativeEmergence(carbon, silicon, result); creativeStrength > 0.5 {
		properties = append(properties, models.EmergentProperty{
			Name:        "Creative Synthesis",
			Description: "Novel creative solutions emerged from the fusion process",
			Strength:    creativeStrength,
			Type:        models.EmergentPropertyTypeCreative,
			Evidence:    []string{"Original idea generation", "Unexpected connections"},
			Impact:      creativeStrength * 0.9,
		})
	}

	return properties, nil
}

// 辅助函数
func calculateAccuracy(result *models.FusionResult) float64 {
	// 基于协同效应和贡献度计算准确�?
	return math.Min(1.0, result.SynergyScore*(result.CarbonContribution+result.SiliconContribution)/2.0)
}

func calculateRelevance(result *models.FusionResult) float64 {
	// 基于输出内容和洞察的相关�?
	return math.Min(1.0, float64(len(result.Insights))*0.2+result.SynergyScore*0.8)
}

func calculateCompleteness(result *models.FusionResult) float64 {
	// 基于输出的完整�?
	completeness := 0.0
	if len(result.SynthesizedOutput) > 0 {
		completeness += 0.4
	}
	if len(result.Insights) > 0 {
		completeness += 0.3
	}
	if len(result.Recommendations) > 0 {
		completeness += 0.3
	}
	return completeness
}

func calculateCoherence(result *models.FusionResult) float64 {
	// 基于协同效应评估连贯�?
	coherence := 0.0
	if result.SynergyScore > 0.5 {
		coherence += 0.6
	}
	if result.CarbonContribution+result.SiliconContribution > 0.7 {
		coherence += 0.4
	}
	return coherence
}

func calculateCreativity(result *models.FusionResult) float64 {
	// 基于涌现特性和新颖性评估创造�?
	creativity := 0.0
	for _, prop := range result.EmergentProperties {
		if prop.Type == models.EmergentPropertyTypeCreative {
			creativity = math.Max(creativity, prop.Strength)
		}
	}
	return creativity
}

func calculatePracticality(result *models.FusionResult) float64 {
	// 基于建议的实用�?
	practicality := 0.0
	if len(result.Recommendations) > 0 {
		practicality += 0.25
	}
	if result.SynergyScore > 0.5 {
		practicality += 0.25
	}
	return math.Min(1.0, practicality)
}

func detectCognitiveEmergence(carbon *models.CarbonInput, silicon *models.SiliconInput, result *models.FusionResult) float64 {
	// 检测认知涌现的强度
	strength := 0.0

	// 基于协同效应
	strength += result.SynergyScore * 0.4

	// 基于洞察数量和质量评�?
	strength += math.Min(0.6, float64(len(result.Insights))*0.1)

	return math.Min(1.0, strength)
}

func detectCreativeEmergence(carbon *models.CarbonInput, silicon *models.SiliconInput, result *models.FusionResult) float64 {
	// 检测创造性涌现的强度
	strength := 0.0

	// 如果碳基输入包含创造力数据
	if carbon.Creativity != nil {
		strength += carbon.Creativity.Originality * 0.3
		strength += carbon.Creativity.Flexibility * 0.2
	}

	// 基于协同效应
	strength += result.SynergyScore * 0.5

	return math.Min(1.0, strength)
}
