package metalearning

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LearningStrategy 学习策略类型
type LearningStrategy string

const (
	StrategyGradientBased    LearningStrategy = "gradient_based"    // 基于梯度的学�?	StrategyModelAgnostic    LearningStrategy = "model_agnostic"    // 模型无关学习
	StrategyMemoryAugmented  LearningStrategy = "memory_augmented"  // 记忆增强学习
	StrategyFewShot          LearningStrategy = "few_shot"          // 少样本学�?	StrategyTransferLearning LearningStrategy = "transfer_learning" // 迁移学习
	StrategyOnlineAdaptation LearningStrategy = "online_adaptation" // 在线适应
)

// Task 学习任务定义
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Domain      string                 `json:"domain"`
	Type        string                 `json:"type"`
	Data        []DataPoint            `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Difficulty  float64                `json:"difficulty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// DataPoint 数据�?type DataPoint struct {
	Input  interface{} `json:"input"`
	Output interface{} `json:"output"`
	Weight float64     `json:"weight"`
}

// LearningResult 学习结果
type LearningResult struct {
	TaskID           string                 `json:"task_id"`
	Strategy         LearningStrategy       `json:"strategy"`
	Performance      float64                `json:"performance"`
	LearningTime     time.Duration          `json:"learning_time"`
	AdaptationSteps  int                    `json:"adaptation_steps"`
	KnowledgeGained  map[string]interface{} `json:"knowledge_gained"`
	TransferredFrom  []string               `json:"transferred_from"`
	Confidence       float64                `json:"confidence"`
	CreatedAt        time.Time              `json:"created_at"`
}

// MetaKnowledge 元知识结�?type MetaKnowledge struct {
	TaskPatterns     map[string][]string    `json:"task_patterns"`
	StrategyMappings map[string]string      `json:"strategy_mappings"`
	PerformanceHist  map[string][]float64   `json:"performance_history"`
	AdaptationRules  []AdaptationRule       `json:"adaptation_rules"`
	TransferMatrix   map[string]map[string]float64 `json:"transfer_matrix"`
}

// AdaptationRule 适应规则
type AdaptationRule struct {
	Condition string  `json:"condition"`
	Action    string  `json:"action"`
	Priority  int     `json:"priority"`
	Success   float64 `json:"success_rate"`
}

// MetaLearningEngine 元学习引�?type MetaLearningEngine struct {
	metaKnowledge    *MetaKnowledge
	strategies       map[LearningStrategy]LearningStrategyImpl
	taskHistory      []Task
	resultHistory    []LearningResult
	mu               sync.RWMutex
	
	// 配置参数
	maxTaskHistory   int
	adaptationThresh float64
	transferThresh   float64
	
	// 统计信息
	totalTasks       int64
	successfulAdapts int64
	avgPerformance   float64
}

// LearningStrategyImpl 学习策略实现接口
type LearningStrategyImpl interface {
	GetStrategy() LearningStrategy
	Learn(ctx context.Context, task *Task, metaKnowledge *MetaKnowledge) (*LearningResult, error)
	Adapt(ctx context.Context, task *Task, priorKnowledge map[string]interface{}) (*LearningResult, error)
	EstimatePerformance(task *Task, metaKnowledge *MetaKnowledge) float64
}

// NewMetaLearningEngine 创建元学习引�?func NewMetaLearningEngine() *MetaLearningEngine {
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
	
	// 初始化学习策�?	engine.initializeStrategies()
	
	return engine
}

// LearnNewTask 学习新任�?func (mle *MetaLearningEngine) LearnNewTask(ctx context.Context, task *Task) (*LearningResult, error) {
	startTime := time.Now()
	
	// 选择最佳学习策�?	strategy, err := mle.selectBestStrategy(task)
	if err != nil {
		return nil, fmt.Errorf("failed to select learning strategy: %w", err)
	}
	
	// 执行学习
	result, err := strategy.Learn(ctx, task, mle.metaKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to learn task: %w", err)
	}
	
	// 更新学习时间
	result.LearningTime = time.Since(startTime)
	
	// 更新元知�?	mle.updateMetaKnowledge(task, result)
	
	// 记录历史
	mle.recordTaskAndResult(task, result)
	
	return result, nil
}

// AdaptToNewTask 适应新任�?func (mle *MetaLearningEngine) AdaptToNewTask(ctx context.Context, task *Task) (*LearningResult, error) {
	// 寻找相似任务
	similarTasks := mle.findSimilarTasks(task)
	if len(similarTasks) == 0 {
		return mle.LearnNewTask(ctx, task)
	}
	
	// 提取先验知识
	priorKnowledge := mle.extractPriorKnowledge(similarTasks)
	
	// 选择适应策略
	strategy, err := mle.selectAdaptationStrategy(task, similarTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to select adaptation strategy: %w", err)
	}
	
	// 执行适应
	result, err := strategy.Adapt(ctx, task, priorKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt to task: %w", err)
	}
	
	// 更新统计
	mle.updateAdaptationStats(result)
	
	return result, nil
}

// TransferKnowledge 知识迁移
func (mle *MetaLearningEngine) TransferKnowledge(ctx context.Context, sourceTask, targetTask *Task) (*LearningResult, error) {
	// 计算任务相似�?	similarity := mle.calculateTaskSimilarity(sourceTask, targetTask)
	if similarity < mle.transferThresh {
		return nil, fmt.Errorf("tasks are not similar enough for knowledge transfer (similarity: %.2f)", similarity)
	}
	
	// 提取可迁移知�?	transferableKnowledge := mle.extractTransferableKnowledge(sourceTask, targetTask)
	
	// 执行知识迁移
	strategy := mle.strategies[StrategyTransferLearning]
	result, err := strategy.Adapt(ctx, targetTask, transferableKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer knowledge: %w", err)
	}
	
	// 更新迁移矩阵
	mle.updateTransferMatrix(sourceTask.Domain, targetTask.Domain, result.Performance)
	
	return result, nil
}

// FewShotLearning 少样本学�?func (mle *MetaLearningEngine) FewShotLearning(ctx context.Context, task *Task, supportSet []DataPoint) (*LearningResult, error) {
	// 创建少样本任�?	fewShotTask := &Task{
		ID:       uuid.New().String(),
		Name:     task.Name + "_few_shot",
		Domain:   task.Domain,
		Type:     task.Type,
		Data:     supportSet,
		Metadata: task.Metadata,
		CreatedAt: time.Now(),
	}
	
	// 使用少样本学习策�?	strategy := mle.strategies[StrategyFewShot]
	result, err := strategy.Learn(ctx, fewShotTask, mle.metaKnowledge)
	if err != nil {
		return nil, fmt.Errorf("failed to perform few-shot learning: %w", err)
	}
	
	return result, nil
}

// OnlineAdaptation 在线适应
func (mle *MetaLearningEngine) OnlineAdaptation(ctx context.Context, task *Task, newData []DataPoint) (*LearningResult, error) {
	// 增量更新任务数据
	task.Data = append(task.Data, newData...)
	
	// 使用在线适应策略
	strategy := mle.strategies[StrategyOnlineAdaptation]
	result, err := strategy.Adapt(ctx, task, map[string]interface{}{
		"new_data": newData,
		"incremental": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to perform online adaptation: %w", err)
	}
	
	return result, nil
}

// GetMetaKnowledge 获取元知�?func (mle *MetaLearningEngine) GetMetaKnowledge() *MetaKnowledge {
	mle.mu.RLock()
	defer mle.mu.RUnlock()
	
	// 返回元知识的副本
	return &MetaKnowledge{
		TaskPatterns:     copyStringSliceMap(mle.metaKnowledge.TaskPatterns),
		StrategyMappings: copyStringMap(mle.metaKnowledge.StrategyMappings),
		PerformanceHist:  copyFloat64SliceMap(mle.metaKnowledge.PerformanceHist),
		AdaptationRules:  copyAdaptationRules(mle.metaKnowledge.AdaptationRules),
		TransferMatrix:   copyTransferMatrix(mle.metaKnowledge.TransferMatrix),
	}
}

// GetLearningStats 获取学习统计信息
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

// 私有方法

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
	// 简单的相似度计�?	domainSim := 0.0
	if task1.Domain == task2.Domain {
		domainSim = 1.0
	}
	
	typeSim := 0.0
	if task1.Type == task2.Type {
		typeSim = 1.0
	}
	
	// 可以添加更复杂的相似度计算逻辑
	return (domainSim + typeSim) / 2.0
}

func (mle *MetaLearningEngine) extractPriorKnowledge(similarTasks []Task) map[string]interface{} {
	priorKnowledge := make(map[string]interface{})
	
	// 提取共同模式
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
	// 根据相似任务选择最佳适应策略
	if len(similarTasks) < 5 {
		return mle.strategies[StrategyFewShot], nil
	}
	
	return mle.strategies[StrategyTransferLearning], nil
}

func (mle *MetaLearningEngine) extractTransferableKnowledge(sourceTask, targetTask *Task) map[string]interface{} {
	transferableKnowledge := make(map[string]interface{})
	
	// 提取可迁移的知识
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
	return 0.5 // 默认迁移比例
}

func (mle *MetaLearningEngine) updateMetaKnowledge(task *Task, result *LearningResult) {
	mle.mu.Lock()
	defer mle.mu.Unlock()
	
	// 更新任务模式
	patterns := extractTaskPatterns(task)
	mle.metaKnowledge.TaskPatterns[task.ID] = patterns
	
	// 更新策略映射
	mle.metaKnowledge.StrategyMappings[task.Type] = string(result.Strategy)
	
	// 更新性能历史
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
	
	// 更新迁移性能
	mle.metaKnowledge.TransferMatrix[sourceDomain][targetDomain] = performance
}

func (mle *MetaLearningEngine) recordTaskAndResult(task *Task, result *LearningResult) {
	mle.mu.Lock()
	defer mle.mu.Unlock()
	
	// 记录任务历史
	mle.taskHistory = append(mle.taskHistory, *task)
	if len(mle.taskHistory) > mle.maxTaskHistory {
		mle.taskHistory = mle.taskHistory[1:]
	}
	
	// 记录结果历史
	mle.resultHistory = append(mle.resultHistory, *result)
	if len(mle.resultHistory) > mle.maxTaskHistory {
		mle.resultHistory = mle.resultHistory[1:]
	}
	
	// 更新统计
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

// 辅助函数

func extractTaskPatterns(task *Task) []string {
	patterns := []string{
		fmt.Sprintf("domain:%s", task.Domain),
		fmt.Sprintf("type:%s", task.Type),
		fmt.Sprintf("difficulty:%.1f", task.Difficulty),
	}
	
	// 可以添加更复杂的模式提取逻辑
	return patterns
}

// 复制函数
func copyStringSliceMap(original map[string][]string) map[string][]string {
	copy := make(map[string][]string)
	for k, v := range original {
		copy[k] = make([]string, len(v))
		copy(copy[k], v)
	}
	return copy
}

func copyStringMap(original map[string]string) map[string]string {
	copy := make(map[string]string)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func copyFloat64SliceMap(original map[string][]float64) map[string][]float64 {
	copy := make(map[string][]float64)
	for k, v := range original {
		copy[k] = make([]float64, len(v))
		copy(copy[k], v)
	}
	return copy
}

func copyAdaptationRules(original []AdaptationRule) []AdaptationRule {
	copy := make([]AdaptationRule, len(original))
	copy(copy, original)
	return copy
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
