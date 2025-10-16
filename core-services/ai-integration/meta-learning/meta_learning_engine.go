package metalearning

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LearningStrategy 
type LearningStrategy string

const (
	StrategyGradientBased    LearningStrategy = "gradient_based"    // 
	StrategyModelAgnostic    LearningStrategy = "model_agnostic"    // 
	StrategyMemoryAugmented  LearningStrategy = "memory_augmented"  // 
	StrategyFewShot          LearningStrategy = "few_shot"          // 
	StrategyTransferLearning LearningStrategy = "transfer_learning" // 
	StrategyOnlineAdaptation LearningStrategy = "online_adaptation" // 
)

// Task 
type Task struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Domain     string                 `json:"domain"`
	Type       string                 `json:"type"`
	Data       []DataPoint            `json:"data"`
	Metadata   map[string]interface{} `json:"metadata"`
	Difficulty float64                `json:"difficulty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// DataPoint 
type DataPoint struct {
	Input  interface{} `json:"input"`
	Output interface{} `json:"output"`
	Weight float64     `json:"weight"`
}

// LearningResult 
type LearningResult struct {
	TaskID          string                 `json:"task_id"`
	Strategy        LearningStrategy       `json:"strategy"`
	Performance     float64                `json:"performance"`
	LearningTime    time.Duration          `json:"learning_time"`
	AdaptationSteps int                    `json:"adaptation_steps"`
	KnowledgeGained map[string]interface{} `json:"knowledge_gained"`
	TransferredFrom []string               `json:"transferred_from"`
	Confidence      float64                `json:"confidence"`
	CreatedAt       time.Time              `json:"created_at"`
}

// MetaKnowledge 
type MetaKnowledge struct {
	TaskPatterns     map[string][]string           `json:"task_patterns"`
	StrategyMappings map[string]string             `json:"strategy_mappings"`
	PerformanceHist  map[string][]float64          `json:"performance_history"`
	AdaptationRules  []AdaptationRule              `json:"adaptation_rules"`
	TransferMatrix   map[string]map[string]float64 `json:"transfer_matrix"`
}

// AdaptationRule 
type AdaptationRule struct {
	Condition string  `json:"condition"`
	Action    string  `json:"action"`
	Priority  int     `json:"priority"`
	Success   float64 `json:"success_rate"`
}

// MetaLearningEngine 
type MetaLearningEngine struct {
	metaKnowledge *MetaKnowledge
	strategies    map[LearningStrategy]LearningStrategyImpl
	taskHistory   []Task
	resultHistory []LearningResult
	mu            sync.RWMutex

	// 
	maxTaskHistory   int
	adaptationThresh float64
	transferThresh   float64

	// 
	totalTasks       int64
	successfulAdapts int64
	avgPerformance   float64
}

// LearningStrategyImpl 
type LearningStrategyImpl interface {
	GetStrategy() LearningStrategy
	Learn(ctx context.Context, task *Task, metaKnowledge *MetaKnowledge) (*LearningResult, error)
	Adapt(ctx context.Context, task *Task, priorKnowledge map[string]interface{}) (*LearningResult, error)
	EstimatePerformance(task *Task, metaKnowledge *MetaKnowledge) float64
}

// NewMetaLearningEngine 
func NewMetaLearningEngine() *MetaLearningEngine {
	engine := &MetaLearningEngine{
		metaKnowledge: &MetaKnowledge{
			TaskPatterns:     make(map[string][]string),
			StrategyMappings: make(map[string]string),
			PerformanceHist:  make(map[string][]float64),
			AdaptationRules:  []AdaptationRule{},
			TransferMatrix:   make(map[string]map[string]float64),
		},
		strategies:       make(map[LearningStrategy]LearningStrategyImpl),
		taskHistory:      []Task{},
		resultHistory:    []LearningResult{},
		maxTaskHistory:   1000,
		adaptationThresh: 0.7,
		transferThresh:   0.6,
	}

	// 
	engine.initializeStrategies()

	return engine
}

// LearnNewTask 
func (mle *MetaLearningEngine) LearnNewTask(ctx context.Context, task *Task) (*LearningResult, error) {
	startTime := time.Now()

	// 
	strategy, err := mle.selectBestStrategy(task)
	if err != nil {
		return nil, fmt.Errorf("failed to select learning strategy: %w", err)
	}

	// 
	result, err := strategy.Learn(ctx, task, mle.metaKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to learn task: %w", err)
	}

	// 
	result.LearningTime = time.Since(startTime)

	// 
	mle.updateMetaKnowledge(task, result)

	// 
	mle.recordTaskAndResult(task, result)

	return result, nil
}

// AdaptToNewTask 
func (mle *MetaLearningEngine) AdaptToNewTask(ctx context.Context, task *Task) (*LearningResult, error) {
	// 
	similarTasks := mle.findSimilarTasks(task)
	if len(similarTasks) == 0 {
		return mle.LearnNewTask(ctx, task)
	}

	// 
	priorKnowledge := mle.extractPriorKnowledge(similarTasks)

	// 
	strategy, err := mle.selectAdaptationStrategy(task, similarTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to select adaptation strategy: %w", err)
	}

	// 
	result, err := strategy.Adapt(ctx, task, priorKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt to task: %w", err)
	}

	// 
	mle.updateAdaptationStats(result)

	return result, nil
}

// TransferKnowledge 
func (mle *MetaLearningEngine) TransferKnowledge(ctx context.Context, sourceTask, targetTask *Task) (*LearningResult, error) {
	// 
	similarity := mle.calculateTaskSimilarity(sourceTask, targetTask)
	if similarity < mle.transferThresh {
		return nil, fmt.Errorf("tasks are not similar enough for knowledge transfer (similarity: %.2f)", similarity)
	}

	// 
	transferableKnowledge := mle.extractTransferableKnowledge(sourceTask, targetTask)

	// 
	strategy := mle.strategies[StrategyTransferLearning]
	result, err := strategy.Adapt(ctx, targetTask, transferableKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer knowledge: %w", err)
	}

	// 
	mle.updateTransferMatrix(sourceTask.Domain, targetTask.Domain, result.Performance)

	return result, nil
}

// FewShotLearning 
func (mle *MetaLearningEngine) FewShotLearning(ctx context.Context, task *Task, supportSet []DataPoint) (*LearningResult, error) {
	// 
	fewShotTask := &Task{
		ID:        uuid.New().String(),
		Name:      task.Name + "_few_shot",
		Domain:    task.Domain,
		Type:      task.Type,
		Data:      supportSet,
		Metadata:  task.Metadata,
		CreatedAt: time.Now(),
	}

	// 
	strategy := mle.strategies[StrategyFewShot]
	result, err := strategy.Learn(ctx, fewShotTask, mle.metaKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to perform few-shot learning: %w", err)
	}

	return result, nil
}

// OnlineAdaptation 
func (mle *MetaLearningEngine) OnlineAdaptation(ctx context.Context, task *Task, newData []DataPoint) (*LearningResult, error) {
	// 
	task.Data = append(task.Data, newData...)

	// 
	strategy := mle.strategies[StrategyOnlineAdaptation]
	result, err := strategy.Adapt(ctx, task, map[string]interface{}{
		"new_data":    newData,
		"incremental": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform online adaptation: %w", err)
	}

	return result, nil
}

// GetMetaKnowledge 
func (mle *MetaLearningEngine) GetMetaKnowledge() *MetaKnowledge {
	mle.mu.RLock()
	defer mle.mu.RUnlock()

	// 
	return &MetaKnowledge{
		TaskPatterns:     copyStringSliceMap(mle.metaKnowledge.TaskPatterns),
		StrategyMappings: copyStringMap(mle.metaKnowledge.StrategyMappings),
		PerformanceHist:  copyFloat64SliceMap(mle.metaKnowledge.PerformanceHist),
		AdaptationRules:  copyAdaptationRules(mle.metaKnowledge.AdaptationRules),
		TransferMatrix:   copyTransferMatrix(mle.metaKnowledge.TransferMatrix),
	}
}

// GetLearningStats 
func (mle *MetaLearningEngine) GetLearningStats() map[string]interface{} {
	mle.mu.RLock()
	defer mle.mu.RUnlock()

	adaptationRate := float64(0)
	if mle.totalTasks > 0 {
		adaptationRate = float64(mle.successfulAdapts) / float64(mle.totalTasks)
	}

	return map[string]interface{}{
		"total_tasks":       mle.totalTasks,
		"successful_adapts": mle.successfulAdapts,
		"adaptation_rate":   adaptationRate,
		"avg_performance":   mle.avgPerformance,
		"task_history_size": len(mle.taskHistory),
		"strategies_count":  len(mle.strategies),
	}
}

// 

func (mle *MetaLearningEngine) initializeStrategies() {
	mle.strategies[StrategyGradientBased] = NewGradientBasedStrategy()
	mle.strategies[StrategyModelAgnostic] = NewModelAgnosticStrategy()
	mle.strategies[StrategyMemoryAugmented] = NewMemoryAugmentedStrategy()
	mle.strategies[StrategyFewShot] = NewFewShotStrategy()
	mle.strategies[StrategyTransferLearning] = NewTransferLearningStrategy()
	mle.strategies[StrategyOnlineAdaptation] = NewOnlineAdaptationStrategy()
}

func (mle *MetaLearningEngine) selectBestStrategy(task *Task) (LearningStrategyImpl, error) {
	mle.mu.RLock()
	defer mle.mu.RUnlock()

	var bestStrategy LearningStrategyImpl
	var bestScore float64

	for _, strategy := range mle.strategies {
		score := strategy.EstimatePerformance(task, mle.metaKnowledge)
		if score > bestScore {
			bestScore = score
			bestStrategy = strategy
		}
	}

	if bestStrategy == nil {
		return nil, fmt.Errorf("no suitable strategy found for task")
	}

	return bestStrategy, nil
}

func (mle *MetaLearningEngine) findSimilarTasks(task *Task) []Task {
	mle.mu.RLock()
	defer mle.mu.RUnlock()

	var similarTasks []Task
	for _, histTask := range mle.taskHistory {
		similarity := mle.calculateTaskSimilarity(task, &histTask)
		if similarity > mle.transferThresh {
			similarTasks = append(similarTasks, histTask)
		}
	}

	return similarTasks
}

func (mle *MetaLearningEngine) calculateTaskSimilarity(task1, task2 *Task) float64 {
	// 
	domainSim := 0.0
	if task1.Domain == task2.Domain {
		domainSim = 1.0
	}

	typeSim := 0.0
	if task1.Type == task2.Type {
		typeSim = 1.0
	}

	// 
	return (domainSim + typeSim) / 2.0
}

func (mle *MetaLearningEngine) extractPriorKnowledge(similarTasks []Task) map[string]interface{} {
	priorKnowledge := make(map[string]interface{})

	// 
	patterns := []string{}
	for _, task := range similarTasks {
		if taskPatterns, exists := mle.metaKnowledge.TaskPatterns[task.ID]; exists {
			patterns = append(patterns, taskPatterns...)
		}
	}

	priorKnowledge["patterns"] = patterns
	priorKnowledge["similar_tasks"] = similarTasks

	return priorKnowledge
}

func (mle *MetaLearningEngine) selectAdaptationStrategy(task *Task, similarTasks []Task) (LearningStrategyImpl, error) {
	// 
	if len(similarTasks) < 5 {
		return mle.strategies[StrategyFewShot], nil
	}

	return mle.strategies[StrategyTransferLearning], nil
}

func (mle *MetaLearningEngine) extractTransferableKnowledge(sourceTask, targetTask *Task) map[string]interface{} {
	transferableKnowledge := make(map[string]interface{})

	// 
	transferableKnowledge["source_domain"] = sourceTask.Domain
	transferableKnowledge["target_domain"] = targetTask.Domain
	transferableKnowledge["transfer_ratio"] = mle.getTransferRatio(sourceTask.Domain, targetTask.Domain)

	return transferableKnowledge
}

func (mle *MetaLearningEngine) getTransferRatio(sourceDomain, targetDomain string) float64 {
	if transferMap, exists := mle.metaKnowledge.TransferMatrix[sourceDomain]; exists {
		if ratio, exists := transferMap[targetDomain]; exists {
			return ratio
		}
	}
	return 0.5 // 
}

func (mle *MetaLearningEngine) updateMetaKnowledge(task *Task, result *LearningResult) {
	mle.mu.Lock()
	defer mle.mu.Unlock()

	// 
	patterns := extractTaskPatterns(task)
	mle.metaKnowledge.TaskPatterns[task.ID] = patterns

	// 
	mle.metaKnowledge.StrategyMappings[task.Type] = string(result.Strategy)

	// 
	if _, exists := mle.metaKnowledge.PerformanceHist[task.Domain]; !exists {
		mle.metaKnowledge.PerformanceHist[task.Domain] = []float64{}
	}
	mle.metaKnowledge.PerformanceHist[task.Domain] = append(
		mle.metaKnowledge.PerformanceHist[task.Domain],
		result.Performance,
	)
}

func (mle *MetaLearningEngine) updateTransferMatrix(sourceDomain, targetDomain string, performance float64) {
	mle.mu.Lock()
	defer mle.mu.Unlock()

	if _, exists := mle.metaKnowledge.TransferMatrix[sourceDomain]; !exists {
		mle.metaKnowledge.TransferMatrix[sourceDomain] = make(map[string]float64)
	}

	// 
	mle.metaKnowledge.TransferMatrix[sourceDomain][targetDomain] = performance
}

func (mle *MetaLearningEngine) recordTaskAndResult(task *Task, result *LearningResult) {
	mle.mu.Lock()
	defer mle.mu.Unlock()

	// 
	mle.taskHistory = append(mle.taskHistory, *task)
	if len(mle.taskHistory) > mle.maxTaskHistory {
		mle.taskHistory = mle.taskHistory[1:]
	}

	// 
	mle.resultHistory = append(mle.resultHistory, *result)
	if len(mle.resultHistory) > mle.maxTaskHistory {
		mle.resultHistory = mle.resultHistory[1:]
	}

	// 
	mle.totalTasks++
	mle.avgPerformance = (mle.avgPerformance*float64(mle.totalTasks-1) + result.Performance) / float64(mle.totalTasks)
}

func (mle *MetaLearningEngine) updateAdaptationStats(result *LearningResult) {
	mle.mu.Lock()
	defer mle.mu.Unlock()

	if result.Performance > mle.adaptationThresh {
		mle.successfulAdapts++
	}
}

// 

func extractTaskPatterns(task *Task) []string {
	patterns := []string{
		fmt.Sprintf("domain:%s", task.Domain),
		fmt.Sprintf("type:%s", task.Type),
		fmt.Sprintf("difficulty:%.1f", task.Difficulty),
	}

	// 
	return patterns
}

// 
func copyStringSliceMap(original map[string][]string) map[string][]string {
	result := make(map[string][]string)
	for k, v := range original {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}
	return result
}

func copyStringMap(original map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range original {
		result[k] = v
	}
	return result
}

func copyFloat64SliceMap(original map[string][]float64) map[string][]float64 {
	result := make(map[string][]float64)
	for k, v := range original {
		result[k] = make([]float64, len(v))
		copy(result[k], v)
	}
	return result
}

func copyAdaptationRules(original []AdaptationRule) []AdaptationRule {
	result := make([]AdaptationRule, len(original))
	copy(result, original)
	return result
}

func copyTransferMatrix(original map[string]map[string]float64) map[string]map[string]float64 {
	copy := make(map[string]map[string]float64)
	for k, v := range original {
		copy[k] = make(map[string]float64)
		for k2, v2 := range v {
			copy[k][k2] = v2
		}
	}
	return copy
}

// Initialize 初始化元学习引擎
func (mle *MetaLearningEngine) Initialize() error {
	mle.mu.Lock()
	defer mle.mu.Unlock()

	// 重新初始化元知识
	mle.metaKnowledge = &MetaKnowledge{
		TaskPatterns:     make(map[string][]string),
		StrategyMappings: make(map[string]string),
		PerformanceHist:  make(map[string][]float64),
		AdaptationRules:  []AdaptationRule{},
		TransferMatrix:   make(map[string]map[string]float64),
	}

	// 重新初始化策略
	mle.initializeStrategies()

	// 清空历史记录
	mle.taskHistory = []Task{}
	mle.resultHistory = []LearningResult{}

	// 重置统计信息
	mle.totalTasks = 0
	mle.successfulAdapts = 0
	mle.avgPerformance = 0.0

	return nil
}

// GetStatus 获取元学习引擎状态
func (mle *MetaLearningEngine) GetStatus() map[string]interface{} {
	mle.mu.RLock()
	defer mle.mu.RUnlock()

	status := map[string]interface{}{
		"status":                "running",
		"total_tasks":           mle.totalTasks,
		"successful_adaptations": mle.successfulAdapts,
		"average_performance":   mle.avgPerformance,
		"task_history_size":     len(mle.taskHistory),
		"result_history_size":   len(mle.resultHistory),
		"max_task_history":      mle.maxTaskHistory,
		"adaptation_threshold":  mle.adaptationThresh,
		"transfer_threshold":    mle.transferThresh,
		"available_strategies":  len(mle.strategies),
	}

	// 计算成功率
	if mle.totalTasks > 0 {
		successRate := float64(mle.successfulAdapts) / float64(mle.totalTasks)
		status["adaptation_success_rate"] = successRate
	} else {
		status["adaptation_success_rate"] = 0.0
	}

	// 添加策略信息
	strategies := make([]string, 0, len(mle.strategies))
	for strategy := range mle.strategies {
		strategies = append(strategies, string(strategy))
	}
	status["strategies"] = strategies

	// 添加元知识统计
	status["meta_knowledge"] = map[string]interface{}{
		"task_patterns_count":     len(mle.metaKnowledge.TaskPatterns),
		"strategy_mappings_count": len(mle.metaKnowledge.StrategyMappings),
		"performance_history_count": len(mle.metaKnowledge.PerformanceHist),
		"adaptation_rules_count":  len(mle.metaKnowledge.AdaptationRules),
		"transfer_matrix_size":    len(mle.metaKnowledge.TransferMatrix),
	}

	return status
}

// Shutdown 关闭元学习引擎
func (mle *MetaLearningEngine) Shutdown() error {
	mle.mu.Lock()
	defer mle.mu.Unlock()

	// 清空所有数据
	mle.metaKnowledge = &MetaKnowledge{
		TaskPatterns:     make(map[string][]string),
		StrategyMappings: make(map[string]string),
		PerformanceHist:  make(map[string][]float64),
		AdaptationRules:  []AdaptationRule{},
		TransferMatrix:   make(map[string]map[string]float64),
	}

	// 清空策略
	mle.strategies = make(map[LearningStrategy]LearningStrategyImpl)

	// 清空历史记录
	mle.taskHistory = []Task{}
	mle.resultHistory = []LearningResult{}

	// 重置统计信息
	mle.totalTasks = 0
	mle.successfulAdapts = 0
	mle.avgPerformance = 0.0

	return nil
}

