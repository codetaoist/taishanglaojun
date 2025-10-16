package agi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AGICapability AGI	能力
type AGICapability string

const (
	CapabilityReasoning     AGICapability = "reasoning"     // 推理
	CapabilityPlanning      AGICapability = "planning"      // 计划
	CapabilityLearning      AGICapability = "learning"      // 学习
	CapabilityCreativity    AGICapability = "creativity"    // 创造力
	CapabilityMultimodal    AGICapability = "multimodal"    // 多模态
	CapabilityMetaCognition AGICapability = "metacognition" // 元认知
)

// AGITask AGI	任务
type AGITask struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Description  string                 `json:"description"`
	Input        interface{}            `json:"input"`
	Requirements []string               `json:"requirements"`
	Context      map[string]interface{} `json:"context"`
	Priority     int                    `json:"priority"`
	Deadline     *time.Time             `json:"deadline,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// AGIResponse AGI	响应
type AGIResponse struct {
	TaskID           string                 `json:"task_id"`
	Result           interface{}            `json:"result"`
	Output           interface{}            `json:"output"`
	UsedCapabilities []AGICapability        `json:"used_capabilities"`
	Confidence       float64                `json:"confidence"`
	Reasoning        []string               `json:"reasoning"`
	Metadata         map[string]interface{} `json:"metadata"`
	ProcessTime      time.Duration          `json:"process_time"`
	CreatedAt        time.Time              `json:"created_at"`
}

// CapabilityModule 能力模块
type CapabilityModule interface {
	GetCapability() AGICapability
	Process(ctx context.Context, task *AGITask) (*AGIResponse, error)
	GetConfidence(task *AGITask) float64
	IsApplicable(task *AGITask) bool
}

// AGIService AGI	服务
type AGIService struct {
	capabilities map[AGICapability]CapabilityModule
	taskQueue    chan *AGITask
	resultCache  map[string]*AGIResponse
	mu           sync.RWMutex

	//
	maxConcurrency int
	cacheSize      int

	//
	processedTasks int64
	totalTime      time.Duration
}

// NewAGIService AGI	服务
func NewAGIService() *AGIService {
	service := &AGIService{
		capabilities:   make(map[AGICapability]CapabilityModule),
		taskQueue:      make(chan *AGITask, 1000),
		resultCache:    make(map[string]*AGIResponse),
		maxConcurrency: 10,
		cacheSize:      1000,
	}

	//
	service.initializeCapabilities()

	return service
}

// RegisterCapability 注册能力模块
func (s *AGIService) RegisterCapability(module CapabilityModule) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.capabilities[module.GetCapability()] = module
}

// ProcessTask AGI	处理任务
func (s *AGIService) ProcessTask(ctx context.Context, task *AGITask) (*AGIResponse, error) {
	startTime := time.Now()

	// 检查缓存
	if cached := s.getCachedResult(task.ID); cached != nil {
		return cached, nil
	}

	// 选择能力模块
	module, err := s.selectBestModule(task)
	if err != nil {
		return nil, fmt.Errorf("failed to select capability module: %w", err)
	}

	// 处理任务
	response, err := module.Process(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to process task: %w", err)
	}

	// 计算处理时间
	response.ProcessTime = time.Since(startTime)

	// 缓存结果
	s.cacheResult(task.ID, response)

	// 更新统计信息
	s.updateStats(response.ProcessTime)

	return response, nil
}

// ProcessMultiModalTask AGI	处理多模态任务
func (s *AGIService) ProcessMultiModalTask(ctx context.Context, task *AGITask) (*AGIResponse, error) {
	// 分解多模态任务
	subTasks, err := s.decomposeMultiModalTask(task)
	if err != nil {
		return nil, fmt.Errorf("failed to decompose multimodal task: %w", err)
	}

	// 并行处理子任务
	results := make([]*AGIResponse, len(subTasks))
	var wg sync.WaitGroup
	errChan := make(chan error, len(subTasks))

	for i, subTask := range subTasks {
		wg.Add(1)
		go func(idx int, task *AGITask) {
			defer wg.Done()

			result, err := s.ProcessTask(ctx, task)
			if err != nil {
				errChan <- err
				return
			}
			results[idx] = result
		}(i, subTask)
	}

	wg.Wait()
	close(errChan)

	// 合并子任务结果
	if len(errChan) > 0 {
		return nil, <-errChan
	}

	// 合并子任务结果
	return s.fuseMultiModalResults(task, results)
}

// ReasoningChain AGI	推理链
func (s *AGIService) ReasoningChain(ctx context.Context, problem string, steps []string) (*AGIResponse, error) {
	task := &AGITask{
		ID:          uuid.New().String(),
		Type:        "reasoning_chain",
		Description: problem,
		Context: map[string]interface{}{
			"steps": steps,
		},
		Priority:  1,
		CreatedAt: time.Now(),
	}

	// 选择推理模块
	reasoningModule, exists := s.capabilities[CapabilityReasoning]
	if !exists {
		return nil, fmt.Errorf("reasoning capability not available")
	}

	return reasoningModule.Process(ctx, task)
}

// PlanGeneration AGI	计划生成
func (s *AGIService) PlanGeneration(ctx context.Context, goal string, constraints []string) (*AGIResponse, error) {
	task := &AGITask{
		ID:          uuid.New().String(),
		Type:        "plan_generation",
		Description: goal,
		Context: map[string]interface{}{
			"constraints": constraints,
		},
		Priority:  1,
		CreatedAt: time.Now(),
	}

	// 选择计划模块
	planningModule, exists := s.capabilities[CapabilityPlanning]
	if !exists {
		return nil, fmt.Errorf("planning capability not available")
	}

	return planningModule.Process(ctx, task)
}

// CreativeGeneration AGI	创意生成
func (s *AGIService) CreativeGeneration(ctx context.Context, prompt string, style string) (*AGIResponse, error) {
	task := &AGITask{
		ID:          uuid.New().String(),
		Type:        "creative_generation",
		Description: prompt,
		Context: map[string]interface{}{
			"style": style,
		},
		Priority:  1,
		CreatedAt: time.Now(),
	}

	// 选择创意模块
	creativityModule, exists := s.capabilities[CapabilityCreativity]
	if !exists {
		return nil, fmt.Errorf("creativity capability not available")
	}

	return creativityModule.Process(ctx, task)
}

// GetCapabilities AGI	获取能力
func (s *AGIService) GetCapabilities() []AGICapability {
	s.mu.RLock()
	defer s.mu.RUnlock()

	capabilities := make([]AGICapability, 0, len(s.capabilities))
	for cap := range s.capabilities {
		capabilities = append(capabilities, cap)
	}

	return capabilities
}

// GetStats AGI	获取统计信息
func (s *AGIService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	avgTime := time.Duration(0)
	if s.processedTasks > 0 {
		avgTime = time.Duration(int64(s.totalTime) / s.processedTasks)
	}

	return map[string]interface{}{
		"processed_tasks":        s.processedTasks,
		"total_time":             s.totalTime.String(),
		"average_time":           avgTime.String(),
		"cache_size":             len(s.resultCache),
		"available_capabilities": len(s.capabilities),
	}
}



// InitializeCapabilities AGI	初始化能力模块
func (s *AGIService) initializeCapabilities() {
	//
	s.RegisterCapability(NewReasoningModule())
	s.RegisterCapability(NewPlanningModule())
	s.RegisterCapability(NewLearningModule())
	s.RegisterCapability(NewCreativityModule())
	s.RegisterCapability(NewMultiModalModule())
	s.RegisterCapability(NewMetaCognitionModule())
}

// selectBestModule AGI	选择最佳能力模块
func (s *AGIService) selectBestModule(task *AGITask) (CapabilityModule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestModule CapabilityModule
	var bestConfidence float64

	for _, module := range s.capabilities {
		if module.IsApplicable(task) {
			confidence := module.GetConfidence(task)
			if confidence > bestConfidence {
				bestConfidence = confidence
				bestModule = module
			}
		}
	}

	if bestModule == nil {
		return nil, fmt.Errorf("no applicable capability module found for task type: %s", task.Type)
	}

	return bestModule, nil
}

// getCachedResult AGI	获取缓存结果
func (s *AGIService) getCachedResult(taskID string) *AGIResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.resultCache[taskID]
}

// cacheResult AGI	缓存结果
func (s *AGIService) cacheResult(taskID string, response *AGIResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// LRU
	if len(s.resultCache) >= s.cacheSize {
		//
		for id := range s.resultCache {
			delete(s.resultCache, id)
			break
		}
	}

	s.resultCache[taskID] = response
}

// updateStats AGI	更新统计信息
func (s *AGIService) updateStats(processTime time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.processedTasks++
	s.totalTime += processTime
}

// decomposeMultiModalTask AGI	分解多模态任务
func (s *AGIService) decomposeMultiModalTask(task *AGITask) ([]*AGITask, error) {
	//
	subTasks := []*AGITask{}

	//
	if task.Type == "text_image_analysis" {
		//
		textTask := &AGITask{
			ID:          uuid.New().String(),
			Type:        "text_analysis",
			Description: task.Description,
			Context:     task.Context,
			Priority:    task.Priority,
			CreatedAt:   time.Now(),
		}
		subTasks = append(subTasks, textTask)

		//
		imageTask := &AGITask{
			ID:          uuid.New().String(),
			Type:        "image_analysis",
			Description: task.Description,
			Context:     task.Context,
			Priority:    task.Priority,
			CreatedAt:   time.Now(),
		}
		subTasks = append(subTasks, imageTask)
	}

	return subTasks, nil
}

// fuseMultiModalResults AGI	融合多模态结果
func (s *AGIService) fuseMultiModalResults(originalTask *AGITask, results []*AGIResponse) (*AGIResponse, error) {
	//
	fusedResult := &AGIResponse{
		TaskID:     originalTask.ID,
		Confidence: 0.0,
		Reasoning:  []string{},
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	// 融合结果
	totalConfidence := 0.0
	for _, result := range results {
		totalConfidence += result.Confidence
		fusedResult.Reasoning = append(fusedResult.Reasoning, result.Reasoning...)
	}
	fusedResult.Confidence = totalConfidence / float64(len(results))

	// 融合元数据
	for key, value := range originalTask.Context {
		fusedResult.Metadata[key] = value
	}

	// 融合结果
	fusedResult.Result = map[string]interface{}{
		"sub_results": results,
		"fusion_type": "weighted_average",
	}

	return fusedResult, nil
}

// Initialize 初始化AGI服务
func (s *AGIService) Initialize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 重新初始化能力模块
	s.initializeCapabilities()

	// 清空缓存
	s.resultCache = make(map[string]*AGIResponse)

	// 重置统计信息
	s.processedTasks = 0
	s.totalTime = 0

	return nil
}

// GetStatus 获取AGI服务状态
func (s *AGIService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"status":                 "running",
		"capabilities_count":     len(s.capabilities),
		"cache_size":             len(s.resultCache),
		"max_cache_size":         s.cacheSize,
		"max_concurrency":        s.maxConcurrency,
		"processed_tasks":        s.processedTasks,
		"total_processing_time":  s.totalTime.String(),
		"available_capabilities": s.GetCapabilities(),
	}

	// 计算平均处理时间
	if s.processedTasks > 0 {
		avgTime := time.Duration(int64(s.totalTime) / s.processedTasks)
		status["average_processing_time"] = avgTime.String()
	} else {
		status["average_processing_time"] = "0s"
	}

	return status
}

// Shutdown 关闭AGI服务
func (s *AGIService) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 清空任务队列
	close(s.taskQueue)
	for range s.taskQueue {
		// 清空剩余任务
	}

	// 清空缓存
	s.resultCache = make(map[string]*AGIResponse)

	// 清空能力模块
	s.capabilities = make(map[AGICapability]CapabilityModule)

	return nil
}
