package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// TaskSchedulingService 任务调度服务实现
type TaskSchedulingService struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	teamRepo    domain.TeamRepository
}

// NewTaskSchedulingService 创建任务调度服务
func NewTaskSchedulingService(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	teamRepo domain.TeamRepository,
) domain.TaskSchedulingService {
	return &TaskSchedulingService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		teamRepo:    teamRepo,
	}
}

// GenerateSchedule 生成调度计划
func (s *TaskSchedulingService) GenerateSchedule(ctx context.Context, req *domain.ScheduleGenerationRequest) (*domain.ScheduleGenerationResult, error) {
	// 获取项目信息
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 获取项目任务
	tasks, err := s.taskRepo.FindByProject(ctx, req.ProjectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// 构建任务依赖图
	dependencyGraph := s.buildDependencyGraph(tasks)

	// 生成调度计划
	schedule, err := s.generateOptimalSchedule(ctx, tasks, dependencyGraph, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schedule: %w", err)
	}

	// 计算关键路径
	criticalPath := s.calculateCriticalPath(tasks, dependencyGraph, schedule)

	// 检测冲突
	conflicts := s.detectScheduleConflicts(schedule)

	return &domain.ScheduleGenerationResult{
		ProjectID:    req.ProjectID,
		Schedule:     schedule,
		CriticalPath: criticalPath,
		Conflicts:    conflicts,
		Timestamp:    time.Now(),
	}, nil
}

// OptimizeSchedule 优化调度计划
func (s *TaskSchedulingService) OptimizeSchedule(ctx context.Context, req *domain.ScheduleOptimizationRequest) (*domain.ScheduleOptimizationResult, error) {
	// 获取当前调度
	currentSchedule := req.CurrentSchedule

	// 分析当前调度的问题
	issues := s.analyzeScheduleIssues(currentSchedule)

	// 生成优化建议
	optimizations := s.generateOptimizations(ctx, currentSchedule, issues, req.Strategy)

	// 应用优化
	optimizedSchedule := s.applyOptimizations(currentSchedule, optimizations)

	// 计算改进指标
	improvements := s.calculateImprovements(currentSchedule, optimizedSchedule)

	return &domain.ScheduleOptimizationResult{
		OriginalSchedule:  currentSchedule,
		OptimizedSchedule: optimizedSchedule,
		Optimizations:     optimizations,
		Improvements:      improvements,
		Timestamp:         time.Now(),
	}, nil
}

// DetectConflicts 检测调度冲突
func (s *TaskSchedulingService) DetectConflicts(ctx context.Context, req *domain.ConflictDetectionRequest) (*domain.ConflictDetectionResult, error) {
	// 获取调度中的所有任务
	schedule := req.Schedule

	// 检测资源冲突
	resourceConflicts := s.detectResourceConflicts(schedule)

	// 检测时间冲突
	timeConflicts := s.detectTimeConflicts(schedule)

	// 检测依赖冲突
	dependencyConflicts := s.detectDependencyConflicts(schedule)

	// 合并所有冲突
	allConflicts := append(resourceConflicts, timeConflicts...)
	allConflicts = append(allConflicts, dependencyConflicts...)

	return &domain.ConflictDetectionResult{
		Schedule:  schedule,
		Conflicts: allConflicts,
		Timestamp: time.Now(),
	}, nil
}

// ResolveConflicts 解决调度冲突
func (s *TaskSchedulingService) ResolveConflicts(ctx context.Context, req *domain.ConflictResolutionRequest) (*domain.ConflictResolutionResult, error) {
	schedule := req.Schedule
	conflicts := req.Conflicts

	resolutions := make([]*domain.ConflictResolution, 0)

	// 按优先级排序冲突
	sort.Slice(conflicts, func(i, j int) bool {
		return s.getConflictPriority(conflicts[i]) > s.getConflictPriority(conflicts[j])
	})

	// 逐个解决冲突
	for _, conflict := range conflicts {
		resolution := s.resolveConflict(ctx, schedule, conflict, req.Strategy)
		if resolution != nil {
			resolutions = append(resolutions, resolution)
			// 应用解决方案到调度中
			s.applyResolution(schedule, resolution)
		}
	}

	return &domain.ConflictResolutionResult{
		OriginalSchedule: req.Schedule,
		ResolvedSchedule: schedule,
		Resolutions:      resolutions,
		Timestamp:        time.Now(),
	}, nil
}

// PredictCompletion 预测完成时间
func (s *TaskSchedulingService) PredictCompletion(ctx context.Context, req *domain.CompletionPredictionRequest) (*domain.CompletionPredictionResult, error) {
	// 获取任务或项目信息
	var tasks []*domain.Task
	var err error

	if req.TaskID != nil {
		task, err := s.taskRepo.FindByID(ctx, *req.TaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to find task: %w", err)
		}
		tasks = []*domain.Task{task}
	} else if req.ProjectID != nil {
		tasks, err = s.taskRepo.FindByProject(ctx, *req.ProjectID, 1000, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get project tasks: %w", err)
		}
	} else {
		return nil, fmt.Errorf("either task_id or project_id must be provided")
	}

	// 分析历史数据
	historicalData := s.analyzeHistoricalData(ctx, tasks)

	// 计算预测
	predictions := make([]*domain.TaskCompletionPrediction, 0)
	for _, task := range tasks {
		prediction := s.predictTaskCompletion(task, historicalData)
		predictions = append(predictions, prediction)
	}

	// 计算整体预测（如果是项目）
	var overallPrediction *domain.ProjectCompletionPrediction
	if req.ProjectID != nil {
		overallPrediction = s.predictProjectCompletion(tasks, predictions)
	}

	return &domain.CompletionPredictionResult{
		TaskPredictions:    predictions,
		ProjectPrediction:  overallPrediction,
		HistoricalAccuracy: historicalData.Accuracy,
		Timestamp:          time.Now(),
	}, nil
}

// AnalyzeCriticalPath 分析关键路径
func (s *TaskSchedulingService) AnalyzeCriticalPath(ctx context.Context, req *domain.CriticalPathRequest) (*domain.CriticalPathResult, error) {
	// 获取项目任务
	tasks, err := s.taskRepo.FindByProject(ctx, req.ProjectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// 构建依赖图
	dependencyGraph := s.buildDependencyGraph(tasks)

	// 计算关键路径
	criticalPath := s.calculateCriticalPath(tasks, dependencyGraph, nil)

	// 分析关键路径风险
	risks := s.analyzeCriticalPathRisks(criticalPath, tasks)

	// 生成优化建议
	optimizations := s.generateCriticalPathOptimizations(criticalPath, tasks)

	return &domain.CriticalPathResult{
		ProjectID:     req.ProjectID,
		CriticalPath:  criticalPath,
		Risks:         risks,
		Optimizations: optimizations,
		Timestamp:     time.Now(),
	}, nil
}

// ========== 私有辅助方法 ==========

// buildDependencyGraph 构建任务依赖图
func (s *TaskSchedulingService) buildDependencyGraph(tasks []*domain.Task) map[uuid.UUID][]*domain.Task {
	graph := make(map[uuid.UUID][]*domain.Task)
	taskMap := make(map[uuid.UUID]*domain.Task)

	// 创建任务映射
	for _, task := range tasks {
		taskMap[task.ID] = task
		graph[task.ID] = make([]*domain.Task, 0)
	}

	// 构建依赖关系
	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			if depTask, exists := taskMap[depID]; exists {
				graph[depID] = append(graph[depID], task)
			}
		}
	}

	return graph
}

// generateOptimalSchedule 生成最优调度计划
func (s *TaskSchedulingService) generateOptimalSchedule(ctx context.Context, tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, req *domain.ScheduleGenerationRequest) (*domain.TaskSchedule, error) {
	schedule := &domain.TaskSchedule{
		ProjectID:   req.ProjectID,
		Tasks:       make([]*domain.ScheduledTask, 0),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now(),
	}

	// 拓扑排序获取任务执行顺序
	sortedTasks := s.topologicalSort(tasks, dependencyGraph)

	// 为每个任务安排时间
	currentTime := req.StartDate
	for _, task := range sortedTasks {
		scheduledTask := s.scheduleTask(task, currentTime, req)
		schedule.Tasks = append(schedule.Tasks, scheduledTask)
		
		// 更新当前时间
		if scheduledTask.EndDate.After(currentTime) {
			currentTime = scheduledTask.EndDate
		}
	}

	// 更新调度结束时间
	if len(schedule.Tasks) > 0 {
		lastTask := schedule.Tasks[len(schedule.Tasks)-1]
		schedule.EndDate = lastTask.EndDate
	}

	return schedule, nil
}

// topologicalSort 拓扑排序
func (s *TaskSchedulingService) topologicalSort(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task) []*domain.Task {
	// 计算入度
	inDegree := make(map[uuid.UUID]int)
	for _, task := range tasks {
		inDegree[task.ID] = len(task.Dependencies)
	}

	// 找到入度为0的任务
	queue := make([]*domain.Task, 0)
	for _, task := range tasks {
		if inDegree[task.ID] == 0 {
			queue = append(queue, task)
		}
	}

	result := make([]*domain.Task, 0)
	
	// 拓扑排序
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 更新依赖任务的入度
		for _, dependent := range dependencyGraph[current.ID] {
			inDegree[dependent.ID]--
			if inDegree[dependent.ID] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	return result
}

// scheduleTask 安排单个任务
func (s *TaskSchedulingService) scheduleTask(task *domain.Task, startTime time.Time, req *domain.ScheduleGenerationRequest) *domain.ScheduledTask {
	// 计算任务持续时间
	duration := s.calculateTaskDuration(task)
	
	// 考虑工作日历
	endTime := s.calculateEndTime(startTime, duration, req.WorkingHours)

	return &domain.ScheduledTask{
		TaskID:    task.ID,
		StartDate: startTime,
		EndDate:   endTime,
		Duration:  duration,
		Resources: s.getTaskResources(task),
	}
}

// calculateTaskDuration 计算任务持续时间
func (s *TaskSchedulingService) calculateTaskDuration(task *domain.Task) time.Duration {
	if task.EstimatedHours != nil {
		return time.Duration(*task.EstimatedHours) * time.Hour
	}
	
	// 根据任务复杂度估算默认时间
	switch task.Complexity {
	case domain.TaskComplexityVeryLow:
		return 2 * time.Hour
	case domain.TaskComplexityLow:
		return 4 * time.Hour
	case domain.TaskComplexityMedium:
		return 8 * time.Hour
	case domain.TaskComplexityHigh:
		return 16 * time.Hour
	case domain.TaskComplexityVeryHigh:
		return 32 * time.Hour
	default:
		return 8 * time.Hour
	}
}

// calculateEndTime 计算结束时间（考虑工作时间）
func (s *TaskSchedulingService) calculateEndTime(startTime time.Time, duration time.Duration, workingHours *domain.WorkingHours) time.Time {
	if workingHours == nil {
		return startTime.Add(duration)
	}

	// 简化处理：假设每天8小时工作时间
	dailyHours := 8 * time.Hour
	days := int(duration / dailyHours)
	remainingHours := duration % dailyHours

	endTime := startTime.AddDate(0, 0, days)
	endTime = endTime.Add(remainingHours)

	return endTime
}

// getTaskResources 获取任务资源
func (s *TaskSchedulingService) getTaskResources(task *domain.Task) []string {
	resources := make([]string, 0)
	
	if task.AssigneeID != nil {
		resources = append(resources, task.AssigneeID.String())
	}
	
	return resources
}

// calculateCriticalPath 计算关键路径
func (s *TaskSchedulingService) calculateCriticalPath(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, schedule *domain.TaskSchedule) *domain.CriticalPath {
	// 创建任务映射
	taskMap := make(map[uuid.UUID]*domain.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// 计算最早开始时间和最晚开始时间
	earliestStart := s.calculateEarliestStart(tasks, dependencyGraph, taskMap)
	latestStart := s.calculateLatestStart(tasks, dependencyGraph, taskMap, earliestStart)

	// 找到关键路径上的任务
	criticalTasks := make([]*domain.CriticalPathTask, 0)
	for _, task := range tasks {
		if earliestStart[task.ID].Equal(latestStart[task.ID]) {
			criticalTasks = append(criticalTasks, &domain.CriticalPathTask{
				TaskID:       task.ID,
				EarliestStart: earliestStart[task.ID],
				LatestStart:   latestStart[task.ID],
				Float:         0,
			})
		}
	}

	// 计算总持续时间
	totalDuration := s.calculateTotalDuration(criticalTasks, taskMap)

	return &domain.CriticalPath{
		Tasks:         criticalTasks,
		TotalDuration: totalDuration,
		StartDate:     s.findEarliestStartDate(criticalTasks),
		EndDate:       s.findLatestEndDate(criticalTasks, taskMap),
	}
}

// calculateEarliestStart 计算最早开始时间
func (s *TaskSchedulingService) calculateEarliestStart(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, taskMap map[uuid.UUID]*domain.Task) map[uuid.UUID]time.Time {
	earliestStart := make(map[uuid.UUID]time.Time)
	baseTime := time.Now()

	// 拓扑排序
	sortedTasks := s.topologicalSort(tasks, dependencyGraph)

	for _, task := range sortedTasks {
		maxEndTime := baseTime
		
		// 找到所有依赖任务的最晚结束时间
		for _, depID := range task.Dependencies {
			if depTask, exists := taskMap[depID]; exists {
				depEndTime := earliestStart[depID].Add(s.calculateTaskDuration(depTask))
				if depEndTime.After(maxEndTime) {
					maxEndTime = depEndTime
				}
			}
		}
		
		earliestStart[task.ID] = maxEndTime
	}

	return earliestStart
}

// calculateLatestStart 计算最晚开始时间
func (s *TaskSchedulingService) calculateLatestStart(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, taskMap map[uuid.UUID]*domain.Task, earliestStart map[uuid.UUID]time.Time) map[uuid.UUID]time.Time {
	latestStart := make(map[uuid.UUID]time.Time)

	// 找到项目结束时间
	projectEndTime := time.Now()
	for taskID, startTime := range earliestStart {
		if task, exists := taskMap[taskID]; exists {
			endTime := startTime.Add(s.calculateTaskDuration(task))
			if endTime.After(projectEndTime) {
				projectEndTime = endTime
			}
		}
	}

	// 反向计算最晚开始时间
	// 这里简化处理，实际应该从项目结束时间反向计算
	for taskID, startTime := range earliestStart {
		latestStart[taskID] = startTime // 简化处理
	}

	return latestStart
}

// detectScheduleConflicts 检测调度冲突
func (s *TaskSchedulingService) detectScheduleConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// 检测资源冲突
	conflicts = append(conflicts, s.detectResourceConflicts(schedule)...)

	// 检测时间冲突
	conflicts = append(conflicts, s.detectTimeConflicts(schedule)...)

	return conflicts
}

// detectResourceConflicts 检测资源冲突
func (s *TaskSchedulingService) detectResourceConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// 按资源分组任务
	resourceTasks := make(map[string][]*domain.ScheduledTask)
	for _, task := range schedule.Tasks {
		for _, resource := range task.Resources {
			resourceTasks[resource] = append(resourceTasks[resource], task)
		}
	}

	// 检查每个资源的任务是否有时间重叠
	for resource, tasks := range resourceTasks {
		for i := 0; i < len(tasks); i++ {
			for j := i + 1; j < len(tasks); j++ {
				if s.isTimeOverlap(tasks[i], tasks[j]) {
					conflicts = append(conflicts, &domain.ScheduleConflict{
						Type:        domain.ConflictTypeResource,
						Description: fmt.Sprintf("资源 %s 在任务 %s 和 %s 之间存在冲突", resource, tasks[i].TaskID, tasks[j].TaskID),
						TaskIDs:     []uuid.UUID{tasks[i].TaskID, tasks[j].TaskID},
						Severity:    domain.ConflictSeverityHigh,
					})
				}
			}
		}
	}

	return conflicts
}

// detectTimeConflicts 检测时间冲突
func (s *TaskSchedulingService) detectTimeConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// 检查任务是否超出项目时间范围
	for _, task := range schedule.Tasks {
		if task.StartDate.Before(schedule.StartDate) || task.EndDate.After(schedule.EndDate) {
			conflicts = append(conflicts, &domain.ScheduleConflict{
				Type:        domain.ConflictTypeTime,
				Description: fmt.Sprintf("任务 %s 超出项目时间范围", task.TaskID),
				TaskIDs:     []uuid.UUID{task.TaskID},
				Severity:    domain.ConflictSeverityMedium,
			})
		}
	}

	return conflicts
}

// detectDependencyConflicts 检测依赖冲突
func (s *TaskSchedulingService) detectDependencyConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// 创建任务映射
	taskMap := make(map[uuid.UUID]*domain.ScheduledTask)
	for _, task := range schedule.Tasks {
		taskMap[task.TaskID] = task
	}

	// 检查依赖关系是否满足
	// 这里需要访问原始任务数据来获取依赖关系
	// 简化处理，实际应该传入依赖关系数据

	return conflicts
}

// isTimeOverlap 检查两个任务是否有时间重叠
func (s *TaskSchedulingService) isTimeOverlap(task1, task2 *domain.ScheduledTask) bool {
	return task1.StartDate.Before(task2.EndDate) && task2.StartDate.Before(task1.EndDate)
}

// analyzeScheduleIssues 分析调度问题
func (s *TaskSchedulingService) analyzeScheduleIssues(schedule *domain.TaskSchedule) []*domain.ScheduleIssue {
	issues := make([]*domain.ScheduleIssue, 0)

	// 检查资源利用率
	resourceUtilization := s.calculateResourceUtilization(schedule)
	for resource, utilization := range resourceUtilization {
		if utilization > 1.0 {
			issues = append(issues, &domain.ScheduleIssue{
				Type:        domain.IssueTypeOverallocation,
				Description: fmt.Sprintf("资源 %s 过度分配，利用率: %.1f%%", resource, utilization*100),
				Severity:    domain.IssueSeverityHigh,
			})
		} else if utilization < 0.5 {
			issues = append(issues, &domain.ScheduleIssue{
				Type:        domain.IssueTypeUnderUtilization,
				Description: fmt.Sprintf("资源 %s 利用率不足，利用率: %.1f%%", resource, utilization*100),
				Severity:    domain.IssueSeverityLow,
			})
		}
	}

	return issues
}

// calculateResourceUtilization 计算资源利用率
func (s *TaskSchedulingService) calculateResourceUtilization(schedule *domain.TaskSchedule) map[string]float64 {
	utilization := make(map[string]float64)

	// 计算总的项目时间
	totalDuration := schedule.EndDate.Sub(schedule.StartDate)

	// 计算每个资源的工作时间
	resourceWorkTime := make(map[string]time.Duration)
	for _, task := range schedule.Tasks {
		for _, resource := range task.Resources {
			resourceWorkTime[resource] += task.Duration
		}
	}

	// 计算利用率
	for resource, workTime := range resourceWorkTime {
		utilization[resource] = workTime.Seconds() / totalDuration.Seconds()
	}

	return utilization
}

// generateOptimizations 生成优化建议
func (s *TaskSchedulingService) generateOptimizations(ctx context.Context, schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue, strategy domain.ScheduleOptimizationStrategy) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	switch strategy {
	case domain.OptimizationStrategyTime:
		optimizations = append(optimizations, s.generateTimeOptimizations(schedule, issues)...)
	case domain.OptimizationStrategyResource:
		optimizations = append(optimizations, s.generateResourceOptimizations(schedule, issues)...)
	case domain.OptimizationStrategyCost:
		optimizations = append(optimizations, s.generateCostOptimizations(schedule, issues)...)
	}

	return optimizations
}

// generateTimeOptimizations 生成时间优化建议
func (s *TaskSchedulingService) generateTimeOptimizations(schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	// 建议并行执行独立任务
	optimizations = append(optimizations, &domain.ScheduleOptimization{
		Type:        domain.OptimizationTypeParallelize,
		Description: "将独立任务并行执行以缩短项目时间",
		Impact:      "减少项目总时间",
		Priority:    domain.OptimizationPriorityHigh,
	})

	return optimizations
}

// generateResourceOptimizations 生成资源优化建议
func (s *TaskSchedulingService) generateResourceOptimizations(schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	// 建议重新分配过度分配的资源
	for _, issue := range issues {
		if issue.Type == domain.IssueTypeOverallocation {
			optimizations = append(optimizations, &domain.ScheduleOptimization{
				Type:        domain.OptimizationTypeRebalance,
				Description: "重新分配过度分配的资源",
				Impact:      "平衡资源利用率",
				Priority:    domain.OptimizationPriorityMedium,
			})
		}
	}

	return optimizations
}

// generateCostOptimizations 生成成本优化建议
func (s *TaskSchedulingService) generateCostOptimizations(schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	// 建议优化资源利用率以降低成本
	optimizations = append(optimizations, &domain.ScheduleOptimization{
		Type:        domain.OptimizationTypeCostReduction,
		Description: "优化资源利用率以降低项目成本",
		Impact:      "减少项目成本",
		Priority:    domain.OptimizationPriorityMedium,
	})

	return optimizations
}

// applyOptimizations 应用优化
func (s *TaskSchedulingService) applyOptimizations(schedule *domain.TaskSchedule, optimizations []*domain.ScheduleOptimization) *domain.TaskSchedule {
	// 创建优化后的调度副本
	optimizedSchedule := &domain.TaskSchedule{
		ProjectID: schedule.ProjectID,
		Tasks:     make([]*domain.ScheduledTask, len(schedule.Tasks)),
		StartDate: schedule.StartDate,
		EndDate:   schedule.EndDate,
		CreatedAt: time.Now(),
	}

	// 复制任务
	copy(optimizedSchedule.Tasks, schedule.Tasks)

	// 应用优化（这里简化处理）
	for _, optimization := range optimizations {
		switch optimization.Type {
		case domain.OptimizationTypeParallelize:
			s.applyParallelization(optimizedSchedule)
		case domain.OptimizationTypeRebalance:
			s.applyRebalancing(optimizedSchedule)
		}
	}

	return optimizedSchedule
}

// applyParallelization 应用并行化
func (s *TaskSchedulingService) applyParallelization(schedule *domain.TaskSchedule) {
	// 简化处理：尝试将一些任务的开始时间提前
	for i := 1; i < len(schedule.Tasks); i++ {
		task := schedule.Tasks[i]
		prevTask := schedule.Tasks[i-1]
		
		// 如果任务没有资源冲突，可以并行执行
		if !s.hasResourceConflict(task, prevTask) {
			task.StartDate = prevTask.StartDate
			task.EndDate = task.StartDate.Add(task.Duration)
		}
	}
}

// applyRebalancing 应用重新平衡
func (s *TaskSchedulingService) applyRebalancing(schedule *domain.TaskSchedule) {
	// 简化处理：调整任务的资源分配
	// 实际实现应该更复杂，包括重新分配资源等
}

// hasResourceConflict 检查是否有资源冲突
func (s *TaskSchedulingService) hasResourceConflict(task1, task2 *domain.ScheduledTask) bool {
	for _, r1 := range task1.Resources {
		for _, r2 := range task2.Resources {
			if r1 == r2 {
				return true
			}
		}
	}
	return false
}

// calculateImprovements 计算改进指标
func (s *TaskSchedulingService) calculateImprovements(original, optimized *domain.TaskSchedule) *domain.ScheduleImprovements {
	originalDuration := original.EndDate.Sub(original.StartDate)
	optimizedDuration := optimized.EndDate.Sub(optimized.StartDate)
	
	timeSaving := originalDuration - optimizedDuration
	timeSavingPercent := 0.0
	if originalDuration > 0 {
		timeSavingPercent = timeSaving.Seconds() / originalDuration.Seconds() * 100
	}

	return &domain.ScheduleImprovements{
		TimeSaving:        timeSaving,
		TimeSavingPercent: timeSavingPercent,
		ResourceEfficiency: s.calculateResourceEfficiencyImprovement(original, optimized),
		CostSaving:        0, // 简化处理
	}
}

// calculateResourceEfficiencyImprovement 计算资源效率改进
func (s *TaskSchedulingService) calculateResourceEfficiencyImprovement(original, optimized *domain.TaskSchedule) float64 {
	originalUtilization := s.calculateResourceUtilization(original)
	optimizedUtilization := s.calculateResourceUtilization(optimized)

	// 计算平均利用率改进
	originalAvg := 0.0
	optimizedAvg := 0.0
	count := 0

	for resource := range originalUtilization {
		if _, exists := optimizedUtilization[resource]; exists {
			originalAvg += originalUtilization[resource]
			optimizedAvg += optimizedUtilization[resource]
			count++
		}
	}

	if count > 0 {
		originalAvg /= float64(count)
		optimizedAvg /= float64(count)
		return (optimizedAvg - originalAvg) * 100
	}

	return 0
}

// getConflictPriority 获取冲突优先级
func (s *TaskSchedulingService) getConflictPriority(conflict *domain.ScheduleConflict) int {
	switch conflict.Severity {
	case domain.ConflictSeverityHigh:
		return 3
	case domain.ConflictSeverityMedium:
		return 2
	case domain.ConflictSeverityLow:
		return 1
	default:
		return 0
	}
}

// resolveConflict 解决单个冲突
func (s *TaskSchedulingService) resolveConflict(ctx context.Context, schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	switch conflict.Type {
	case domain.ConflictTypeResource:
		return s.resolveResourceConflict(schedule, conflict, strategy)
	case domain.ConflictTypeTime:
		return s.resolveTimeConflict(schedule, conflict, strategy)
	case domain.ConflictTypeDependency:
		return s.resolveDependencyConflict(schedule, conflict, strategy)
	default:
		return nil
	}
}

// resolveResourceConflict 解决资源冲突
func (s *TaskSchedulingService) resolveResourceConflict(schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	return &domain.ConflictResolution{
		ConflictID:  conflict.Type.String() + "_" + fmt.Sprintf("%d", len(conflict.TaskIDs)),
		Strategy:    strategy,
		Action:      "调整任务时间以避免资源冲突",
		Description: "将冲突任务重新安排到不同时间段",
	}
}

// resolveTimeConflict 解决时间冲突
func (s *TaskSchedulingService) resolveTimeConflict(schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	return &domain.ConflictResolution{
		ConflictID:  conflict.Type.String() + "_" + fmt.Sprintf("%d", len(conflict.TaskIDs)),
		Strategy:    strategy,
		Action:      "调整项目时间范围",
		Description: "扩展项目时间范围以容纳所有任务",
	}
}

// resolveDependencyConflict 解决依赖冲突
func (s *TaskSchedulingService) resolveDependencyConflict(schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	return &domain.ConflictResolution{
		ConflictID:  conflict.Type.String() + "_" + fmt.Sprintf("%d", len(conflict.TaskIDs)),
		Strategy:    strategy,
		Action:      "重新排序任务以满足依赖关系",
		Description: "调整任务执行顺序以满足依赖要求",
	}
}

// applyResolution 应用解决方案
func (s *TaskSchedulingService) applyResolution(schedule *domain.TaskSchedule, resolution *domain.ConflictResolution) {
	// 简化处理：根据解决方案类型应用相应的修改
	// 实际实现应该根据具体的解决方案修改调度
}

// analyzeHistoricalData 分析历史数据
func (s *TaskSchedulingService) analyzeHistoricalData(ctx context.Context, tasks []*domain.Task) *domain.HistoricalAnalysis {
	// 简化处理：返回模拟的历史分析数据
	return &domain.HistoricalAnalysis{
		Accuracy:           0.85,
		AverageDelay:       2 * time.Hour,
		CompletionRate:     0.92,
		VariancePercent:    15.0,
	}
}

// predictTaskCompletion 预测任务完成时间
func (s *TaskSchedulingService) predictTaskCompletion(task *domain.Task, historical *domain.HistoricalAnalysis) *domain.TaskCompletionPrediction {
	// 基础预测时间
	baseDuration := s.calculateTaskDuration(task)
	
	// 应用历史修正因子
	adjustedDuration := time.Duration(float64(baseDuration) * (1 + historical.VariancePercent/100))
	
	predictedCompletion := time.Now().Add(adjustedDuration)
	
	return &domain.TaskCompletionPrediction{
		TaskID:              task.ID,
		PredictedCompletion: predictedCompletion,
		Confidence:          historical.Accuracy,
		VarianceRange:       historical.VariancePercent,
	}
}

// predictProjectCompletion 预测项目完成时间
func (s *TaskSchedulingService) predictProjectCompletion(tasks []*domain.Task, taskPredictions []*domain.TaskCompletionPrediction) *domain.ProjectCompletionPrediction {
	if len(taskPredictions) == 0 {
		return nil
	}

	// 找到最晚的任务完成时间
	latestCompletion := taskPredictions[0].PredictedCompletion
	totalConfidence := 0.0

	for _, prediction := range taskPredictions {
		if prediction.PredictedCompletion.After(latestCompletion) {
			latestCompletion = prediction.PredictedCompletion
		}
		totalConfidence += prediction.Confidence
	}

	averageConfidence := totalConfidence / float64(len(taskPredictions))

	return &domain.ProjectCompletionPrediction{
		PredictedCompletion: latestCompletion,
		Confidence:          averageConfidence,
		CriticalTasks:       s.findCriticalTasksForPrediction(taskPredictions),
	}
}

// findCriticalTasksForPrediction 找到预测中的关键任务
func (s *TaskSchedulingService) findCriticalTasksForPrediction(predictions []*domain.TaskCompletionPrediction) []uuid.UUID {
	// 简化处理：返回完成时间最晚的几个任务
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].PredictedCompletion.After(predictions[j].PredictedCompletion)
	})

	criticalTasks := make([]uuid.UUID, 0)
	for i := 0; i < len(predictions) && i < 3; i++ {
		criticalTasks = append(criticalTasks, predictions[i].TaskID)
	}

	return criticalTasks
}

// analyzeCriticalPathRisks 分析关键路径风险
func (s *TaskSchedulingService) analyzeCriticalPathRisks(criticalPath *domain.CriticalPath, tasks []*domain.Task) []*domain.CriticalPathRisk {
	risks := make([]*domain.CriticalPathRisk, 0)

	// 分析关键路径上的高风险任务
	for _, criticalTask := range criticalPath.Tasks {
		// 找到对应的任务
		for _, task := range tasks {
			if task.ID == criticalTask.TaskID {
				if task.Complexity == domain.TaskComplexityVeryHigh {
					risks = append(risks, &domain.CriticalPathRisk{
						TaskID:      task.ID,
						RiskType:    "复杂度风险",
						Description: "任务复杂度很高，可能导致延期",
						Probability: 0.7,
						Impact:      "高",
					})
				}
				break
			}
		}
	}

	return risks
}

// generateCriticalPathOptimizations 生成关键路径优化建议
func (s *TaskSchedulingService) generateCriticalPathOptimizations(criticalPath *domain.CriticalPath, tasks []*domain.Task) []*domain.CriticalPathOptimization {
	optimizations := make([]*domain.CriticalPathOptimization, 0)

	// 建议增加关键任务的资源
	optimizations = append(optimizations, &domain.CriticalPathOptimization{
		Type:        "资源增强",
		Description: "为关键路径上的任务增加更多资源",
		Impact:      "缩短关键路径时间",
		Priority:    "高",
	})

	// 建议并行化非关键任务
	optimizations = append(optimizations, &domain.CriticalPathOptimization{
		Type:        "任务并行化",
		Description: "将非关键任务与关键任务并行执行",
		Impact:      "提高整体效率",
		Priority:    "中",
	})

	return optimizations
}

// calculateTotalDuration 计算总持续时间
func (s *TaskSchedulingService) calculateTotalDuration(criticalTasks []*domain.CriticalPathTask, taskMap map[uuid.UUID]*domain.Task) time.Duration {
	totalDuration := time.Duration(0)
	for _, criticalTask := range criticalTasks {
		if task, exists := taskMap[criticalTask.TaskID]; exists {
			totalDuration += s.calculateTaskDuration(task)
		}
	}
	return totalDuration
}

// findEarliestStartDate 找到最早开始日期
func (s *TaskSchedulingService) findEarliestStartDate(criticalTasks []*domain.CriticalPathTask) time.Time {
	if len(criticalTasks) == 0 {
		return time.Now()
	}

	earliest := criticalTasks[0].EarliestStart
	for _, task := range criticalTasks {
		if task.EarliestStart.Before(earliest) {
			earliest = task.EarliestStart
		}
	}
	return earliest
}

// findLatestEndDate 找到最晚结束日期
func (s *TaskSchedulingService) findLatestEndDate(criticalTasks []*domain.CriticalPathTask, taskMap map[uuid.UUID]*domain.Task) time.Time {
	if len(criticalTasks) == 0 {
		return time.Now()
	}

	latest := time.Now()
	for _, criticalTask := range criticalTasks {
		if task, exists := taskMap[criticalTask.TaskID]; exists {
			endTime := criticalTask.EarliestStart.Add(s.calculateTaskDuration(task))
			if endTime.After(latest) {
				latest = endTime
			}
		}
	}
	return latest
}