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

// TaskSchedulingService 
type TaskSchedulingService struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	teamRepo    domain.TeamRepository
}

// NewTaskSchedulingService 
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

// GenerateSchedule 
func (s *TaskSchedulingService) GenerateSchedule(ctx context.Context, req *domain.ScheduleGenerationRequest) (*domain.ScheduleGenerationResult, error) {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 
	tasks, err := s.taskRepo.FindByProject(ctx, req.ProjectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// ?
	dependencyGraph := s.buildDependencyGraph(tasks)

	// 
	schedule, err := s.generateOptimalSchedule(ctx, tasks, dependencyGraph, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schedule: %w", err)
	}

	// 
	criticalPath := s.calculateCriticalPath(tasks, dependencyGraph, schedule)

	// ?
	conflicts := s.detectScheduleConflicts(schedule)

	return &domain.ScheduleGenerationResult{
		ProjectID:    req.ProjectID,
		Schedule:     schedule,
		CriticalPath: criticalPath,
		Conflicts:    conflicts,
		Timestamp:    time.Now(),
	}, nil
}

// OptimizeSchedule 
func (s *TaskSchedulingService) OptimizeSchedule(ctx context.Context, req *domain.ScheduleOptimizationRequest) (*domain.ScheduleOptimizationResult, error) {
	// 
	currentSchedule := req.CurrentSchedule

	// ?
	issues := s.analyzeScheduleIssues(currentSchedule)

	// 
	optimizations := s.generateOptimizations(ctx, currentSchedule, issues, req.Strategy)

	// 
	optimizedSchedule := s.applyOptimizations(currentSchedule, optimizations)

	// 
	improvements := s.calculateImprovements(currentSchedule, optimizedSchedule)

	return &domain.ScheduleOptimizationResult{
		OriginalSchedule:  currentSchedule,
		OptimizedSchedule: optimizedSchedule,
		Optimizations:     optimizations,
		Improvements:      improvements,
		Timestamp:         time.Now(),
	}, nil
}

// DetectConflicts ?
func (s *TaskSchedulingService) DetectConflicts(ctx context.Context, req *domain.ConflictDetectionRequest) (*domain.ConflictDetectionResult, error) {
	// ?
	schedule := req.Schedule

	// ?
	resourceConflicts := s.detectResourceConflicts(schedule)

	// ?
	timeConflicts := s.detectTimeConflicts(schedule)

	// ?
	dependencyConflicts := s.detectDependencyConflicts(schedule)

	// ?
	allConflicts := append(resourceConflicts, timeConflicts...)
	allConflicts = append(allConflicts, dependencyConflicts...)

	return &domain.ConflictDetectionResult{
		Schedule:  schedule,
		Conflicts: allConflicts,
		Timestamp: time.Now(),
	}, nil
}

// ResolveConflicts 
func (s *TaskSchedulingService) ResolveConflicts(ctx context.Context, req *domain.ConflictResolutionRequest) (*domain.ConflictResolutionResult, error) {
	schedule := req.Schedule
	conflicts := req.Conflicts

	resolutions := make([]*domain.ConflictResolution, 0)

	// 
	sort.Slice(conflicts, func(i, j int) bool {
		return s.getConflictPriority(conflicts[i]) > s.getConflictPriority(conflicts[j])
	})

	// 
	for _, conflict := range conflicts {
		resolution := s.resolveConflict(ctx, schedule, conflict, req.Strategy)
		if resolution != nil {
			resolutions = append(resolutions, resolution)
			// 
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

// PredictCompletion 
func (s *TaskSchedulingService) PredictCompletion(ctx context.Context, req *domain.CompletionPredictionRequest) (*domain.CompletionPredictionResult, error) {
	// ?
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

	// 
	historicalData := s.analyzeHistoricalData(ctx, tasks)

	// 
	predictions := make([]*domain.TaskCompletionPrediction, 0)
	for _, task := range tasks {
		prediction := s.predictTaskCompletion(task, historicalData)
		predictions = append(predictions, prediction)
	}

	// ?
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

// AnalyzeCriticalPath 
func (s *TaskSchedulingService) AnalyzeCriticalPath(ctx context.Context, req *domain.CriticalPathRequest) (*domain.CriticalPathResult, error) {
	// 
	tasks, err := s.taskRepo.FindByProject(ctx, req.ProjectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}

	// ?
	dependencyGraph := s.buildDependencyGraph(tasks)

	// 
	criticalPath := s.calculateCriticalPath(tasks, dependencyGraph, nil)

	// 
	risks := s.analyzeCriticalPathRisks(criticalPath, tasks)

	// 
	optimizations := s.generateCriticalPathOptimizations(criticalPath, tasks)

	return &domain.CriticalPathResult{
		ProjectID:     req.ProjectID,
		CriticalPath:  criticalPath,
		Risks:         risks,
		Optimizations: optimizations,
		Timestamp:     time.Now(),
	}, nil
}

// ==========  ==========

// buildDependencyGraph ?
func (s *TaskSchedulingService) buildDependencyGraph(tasks []*domain.Task) map[uuid.UUID][]*domain.Task {
	graph := make(map[uuid.UUID][]*domain.Task)
	taskMap := make(map[uuid.UUID]*domain.Task)

	// 
	for _, task := range tasks {
		taskMap[task.ID] = task
		graph[task.ID] = make([]*domain.Task, 0)
	}

	// 
	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			if depTask, exists := taskMap[depID]; exists {
				graph[depID] = append(graph[depID], task)
			}
		}
	}

	return graph
}

// generateOptimalSchedule ?
func (s *TaskSchedulingService) generateOptimalSchedule(ctx context.Context, tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, req *domain.ScheduleGenerationRequest) (*domain.TaskSchedule, error) {
	schedule := &domain.TaskSchedule{
		ProjectID:   req.ProjectID,
		Tasks:       make([]*domain.ScheduledTask, 0),
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now(),
	}

	// 
	sortedTasks := s.topologicalSort(tasks, dependencyGraph)

	// ?
	currentTime := req.StartDate
	for _, task := range sortedTasks {
		scheduledTask := s.scheduleTask(task, currentTime, req)
		schedule.Tasks = append(schedule.Tasks, scheduledTask)
		
		// 
		if scheduledTask.EndDate.After(currentTime) {
			currentTime = scheduledTask.EndDate
		}
	}

	// 
	if len(schedule.Tasks) > 0 {
		lastTask := schedule.Tasks[len(schedule.Tasks)-1]
		schedule.EndDate = lastTask.EndDate
	}

	return schedule, nil
}

// topologicalSort 
func (s *TaskSchedulingService) topologicalSort(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task) []*domain.Task {
	// 
	inDegree := make(map[uuid.UUID]int)
	for _, task := range tasks {
		inDegree[task.ID] = len(task.Dependencies)
	}

	// ??
	queue := make([]*domain.Task, 0)
	for _, task := range tasks {
		if inDegree[task.ID] == 0 {
			queue = append(queue, task)
		}
	}

	result := make([]*domain.Task, 0)
	
	// 
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// ?
		for _, dependent := range dependencyGraph[current.ID] {
			inDegree[dependent.ID]--
			if inDegree[dependent.ID] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	return result
}

// scheduleTask 
func (s *TaskSchedulingService) scheduleTask(task *domain.Task, startTime time.Time, req *domain.ScheduleGenerationRequest) *domain.ScheduledTask {
	// 
	duration := s.calculateTaskDuration(task)
	
	// 
	endTime := s.calculateEndTime(startTime, duration, req.WorkingHours)

	return &domain.ScheduledTask{
		TaskID:    task.ID,
		StartDate: startTime,
		EndDate:   endTime,
		Duration:  duration,
		Resources: s.getTaskResources(task),
	}
}

// calculateTaskDuration 
func (s *TaskSchedulingService) calculateTaskDuration(task *domain.Task) time.Duration {
	if task.EstimatedHours != nil {
		return time.Duration(*task.EstimatedHours) * time.Hour
	}
	
	// ?
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

// calculateEndTime 俼?
func (s *TaskSchedulingService) calculateEndTime(startTime time.Time, duration time.Duration, workingHours *domain.WorkingHours) time.Time {
	if workingHours == nil {
		return startTime.Add(duration)
	}

	// 8
	dailyHours := 8 * time.Hour
	days := int(duration / dailyHours)
	remainingHours := duration % dailyHours

	endTime := startTime.AddDate(0, 0, days)
	endTime = endTime.Add(remainingHours)

	return endTime
}

// getTaskResources 
func (s *TaskSchedulingService) getTaskResources(task *domain.Task) []string {
	resources := make([]string, 0)
	
	if task.AssigneeID != nil {
		resources = append(resources, task.AssigneeID.String())
	}
	
	return resources
}

// calculateCriticalPath 
func (s *TaskSchedulingService) calculateCriticalPath(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, schedule *domain.TaskSchedule) *domain.CriticalPath {
	// 
	taskMap := make(map[uuid.UUID]*domain.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// 翪?
	earliestStart := s.calculateEarliestStart(tasks, dependencyGraph, taskMap)
	latestStart := s.calculateLatestStart(tasks, dependencyGraph, taskMap, earliestStart)

	// 
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

	// ?
	totalDuration := s.calculateTotalDuration(criticalTasks, taskMap)

	return &domain.CriticalPath{
		Tasks:         criticalTasks,
		TotalDuration: totalDuration,
		StartDate:     s.findEarliestStartDate(criticalTasks),
		EndDate:       s.findLatestEndDate(criticalTasks, taskMap),
	}
}

// calculateEarliestStart 翪?
func (s *TaskSchedulingService) calculateEarliestStart(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, taskMap map[uuid.UUID]*domain.Task) map[uuid.UUID]time.Time {
	earliestStart := make(map[uuid.UUID]time.Time)
	baseTime := time.Now()

	// 
	sortedTasks := s.topologicalSort(tasks, dependencyGraph)

	for _, task := range sortedTasks {
		maxEndTime := baseTime
		
		// ?
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

// calculateLatestStart ?
func (s *TaskSchedulingService) calculateLatestStart(tasks []*domain.Task, dependencyGraph map[uuid.UUID][]*domain.Task, taskMap map[uuid.UUID]*domain.Task, earliestStart map[uuid.UUID]time.Time) map[uuid.UUID]time.Time {
	latestStart := make(map[uuid.UUID]time.Time)

	// 
	projectEndTime := time.Now()
	for taskID, startTime := range earliestStart {
		if task, exists := taskMap[taskID]; exists {
			endTime := startTime.Add(s.calculateTaskDuration(task))
			if endTime.After(projectEndTime) {
				projectEndTime = endTime
			}
		}
	}

	// ?
	// ?
	for taskID, startTime := range earliestStart {
		latestStart[taskID] = startTime // ?
	}

	return latestStart
}

// detectScheduleConflicts ?
func (s *TaskSchedulingService) detectScheduleConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// ?
	conflicts = append(conflicts, s.detectResourceConflicts(schedule)...)

	// ?
	conflicts = append(conflicts, s.detectTimeConflicts(schedule)...)

	return conflicts
}

// detectResourceConflicts ?
func (s *TaskSchedulingService) detectResourceConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// ?
	resourceTasks := make(map[string][]*domain.ScheduledTask)
	for _, task := range schedule.Tasks {
		for _, resource := range task.Resources {
			resourceTasks[resource] = append(resourceTasks[resource], task)
		}
	}

	// ?
	for resource, tasks := range resourceTasks {
		for i := 0; i < len(tasks); i++ {
			for j := i + 1; j < len(tasks); j++ {
				if s.isTimeOverlap(tasks[i], tasks[j]) {
					conflicts = append(conflicts, &domain.ScheduleConflict{
						Type:        domain.ConflictTypeResource,
						Description: fmt.Sprintf(" %s ?%s ?%s ", resource, tasks[i].TaskID, tasks[j].TaskID),
						TaskIDs:     []uuid.UUID{tasks[i].TaskID, tasks[j].TaskID},
						Severity:    domain.ConflictSeverityHigh,
					})
				}
			}
		}
	}

	return conflicts
}

// detectTimeConflicts ?
func (s *TaskSchedulingService) detectTimeConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// ?
	for _, task := range schedule.Tasks {
		if task.StartDate.Before(schedule.StartDate) || task.EndDate.After(schedule.EndDate) {
			conflicts = append(conflicts, &domain.ScheduleConflict{
				Type:        domain.ConflictTypeTime,
				Description: fmt.Sprintf(" %s ", task.TaskID),
				TaskIDs:     []uuid.UUID{task.TaskID},
				Severity:    domain.ConflictSeverityMedium,
			})
		}
	}

	return conflicts
}

// detectDependencyConflicts ?
func (s *TaskSchedulingService) detectDependencyConflicts(schedule *domain.TaskSchedule) []*domain.ScheduleConflict {
	conflicts := make([]*domain.ScheduleConflict, 0)

	// 
	taskMap := make(map[uuid.UUID]*domain.ScheduledTask)
	for _, task := range schedule.Tasks {
		taskMap[task.TaskID] = task
	}

	// ?
	// 
	// 

	return conflicts
}

// isTimeOverlap 
func (s *TaskSchedulingService) isTimeOverlap(task1, task2 *domain.ScheduledTask) bool {
	return task1.StartDate.Before(task2.EndDate) && task2.StartDate.Before(task1.EndDate)
}

// analyzeScheduleIssues 
func (s *TaskSchedulingService) analyzeScheduleIssues(schedule *domain.TaskSchedule) []*domain.ScheduleIssue {
	issues := make([]*domain.ScheduleIssue, 0)

	// 
	resourceUtilization := s.calculateResourceUtilization(schedule)
	for resource, utilization := range resourceUtilization {
		if utilization > 1.0 {
			issues = append(issues, &domain.ScheduleIssue{
				Type:        domain.IssueTypeOverallocation,
				Description: fmt.Sprintf(" %s : %.1f%%", resource, utilization*100),
				Severity:    domain.IssueSeverityHigh,
			})
		} else if utilization < 0.5 {
			issues = append(issues, &domain.ScheduleIssue{
				Type:        domain.IssueTypeUnderUtilization,
				Description: fmt.Sprintf(" %s ? %.1f%%", resource, utilization*100),
				Severity:    domain.IssueSeverityLow,
			})
		}
	}

	return issues
}

// calculateResourceUtilization ?
func (s *TaskSchedulingService) calculateResourceUtilization(schedule *domain.TaskSchedule) map[string]float64 {
	utilization := make(map[string]float64)

	// 
	totalDuration := schedule.EndDate.Sub(schedule.StartDate)

	// ?
	resourceWorkTime := make(map[string]time.Duration)
	for _, task := range schedule.Tasks {
		for _, resource := range task.Resources {
			resourceWorkTime[resource] += task.Duration
		}
	}

	// ?
	for resource, workTime := range resourceWorkTime {
		utilization[resource] = workTime.Seconds() / totalDuration.Seconds()
	}

	return utilization
}

// generateOptimizations 
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

// generateTimeOptimizations 
func (s *TaskSchedulingService) generateTimeOptimizations(schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	// 鲢
	optimizations = append(optimizations, &domain.ScheduleOptimization{
		Type:        domain.OptimizationTypeParallelize,
		Description: "",
		Impact:      "?,
		Priority:    domain.OptimizationPriorityHigh,
	})

	return optimizations
}

// generateResourceOptimizations 
func (s *TaskSchedulingService) generateResourceOptimizations(schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	// ?
	for _, issue := range issues {
		if issue.Type == domain.IssueTypeOverallocation {
			optimizations = append(optimizations, &domain.ScheduleOptimization{
				Type:        domain.OptimizationTypeRebalance,
				Description: "?,
				Impact:      "?,
				Priority:    domain.OptimizationPriorityMedium,
			})
		}
	}

	return optimizations
}

// generateCostOptimizations 
func (s *TaskSchedulingService) generateCostOptimizations(schedule *domain.TaskSchedule, issues []*domain.ScheduleIssue) []*domain.ScheduleOptimization {
	optimizations := make([]*domain.ScheduleOptimization, 0)

	// 
	optimizations = append(optimizations, &domain.ScheduleOptimization{
		Type:        domain.OptimizationTypeCostReduction,
		Description: "",
		Impact:      "",
		Priority:    domain.OptimizationPriorityMedium,
	})

	return optimizations
}

// applyOptimizations 
func (s *TaskSchedulingService) applyOptimizations(schedule *domain.TaskSchedule, optimizations []*domain.ScheduleOptimization) *domain.TaskSchedule {
	// 
	optimizedSchedule := &domain.TaskSchedule{
		ProjectID: schedule.ProjectID,
		Tasks:     make([]*domain.ScheduledTask, len(schedule.Tasks)),
		StartDate: schedule.StartDate,
		EndDate:   schedule.EndDate,
		CreatedAt: time.Now(),
	}

	// 
	copy(optimizedSchedule.Tasks, schedule.Tasks)

	// 
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

// applyParallelization ?
func (s *TaskSchedulingService) applyParallelization(schedule *domain.TaskSchedule) {
	// ?
	for i := 1; i < len(schedule.Tasks); i++ {
		task := schedule.Tasks[i]
		prevTask := schedule.Tasks[i-1]
		
		// ?
		if !s.hasResourceConflict(task, prevTask) {
			task.StartDate = prevTask.StartDate
			task.EndDate = task.StartDate.Add(task.Duration)
		}
	}
}

// applyRebalancing 
func (s *TaskSchedulingService) applyRebalancing(schedule *domain.TaskSchedule) {
	// ?
	// ?
}

// hasResourceConflict 
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

// calculateImprovements 
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
		CostSaving:        0, // ?
	}
}

// calculateResourceEfficiencyImprovement 
func (s *TaskSchedulingService) calculateResourceEfficiencyImprovement(original, optimized *domain.TaskSchedule) float64 {
	originalUtilization := s.calculateResourceUtilization(original)
	optimizedUtilization := s.calculateResourceUtilization(optimized)

	// ?
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

// getConflictPriority ?
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

// resolveConflict 
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

// resolveResourceConflict 
func (s *TaskSchedulingService) resolveResourceConflict(schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	return &domain.ConflictResolution{
		ConflictID:  conflict.Type.String() + "_" + fmt.Sprintf("%d", len(conflict.TaskIDs)),
		Strategy:    strategy,
		Action:      "?,
		Description: "?,
	}
}

// resolveTimeConflict 
func (s *TaskSchedulingService) resolveTimeConflict(schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	return &domain.ConflictResolution{
		ConflictID:  conflict.Type.String() + "_" + fmt.Sprintf("%d", len(conflict.TaskIDs)),
		Strategy:    strategy,
		Action:      "",
		Description: "?,
	}
}

// resolveDependencyConflict 
func (s *TaskSchedulingService) resolveDependencyConflict(schedule *domain.TaskSchedule, conflict *domain.ScheduleConflict, strategy domain.ConflictResolutionStrategy) *domain.ConflictResolution {
	return &domain.ConflictResolution{
		ConflictID:  conflict.Type.String() + "_" + fmt.Sprintf("%d", len(conflict.TaskIDs)),
		Strategy:    strategy,
		Action:      "?,
		Description: "?,
	}
}

// applyResolution 
func (s *TaskSchedulingService) applyResolution(schedule *domain.TaskSchedule, resolution *domain.ConflictResolution) {
	// ?
	// ?
}

// analyzeHistoricalData 
func (s *TaskSchedulingService) analyzeHistoricalData(ctx context.Context, tasks []*domain.Task) *domain.HistoricalAnalysis {
	// ?
	return &domain.HistoricalAnalysis{
		Accuracy:           0.85,
		AverageDelay:       2 * time.Hour,
		CompletionRate:     0.92,
		VariancePercent:    15.0,
	}
}

// predictTaskCompletion 
func (s *TaskSchedulingService) predictTaskCompletion(task *domain.Task, historical *domain.HistoricalAnalysis) *domain.TaskCompletionPrediction {
	// 
	baseDuration := s.calculateTaskDuration(task)
	
	// 
	adjustedDuration := time.Duration(float64(baseDuration) * (1 + historical.VariancePercent/100))
	
	predictedCompletion := time.Now().Add(adjustedDuration)
	
	return &domain.TaskCompletionPrediction{
		TaskID:              task.ID,
		PredictedCompletion: predictedCompletion,
		Confidence:          historical.Accuracy,
		VarianceRange:       historical.VariancePercent,
	}
}

// predictProjectCompletion 
func (s *TaskSchedulingService) predictProjectCompletion(tasks []*domain.Task, taskPredictions []*domain.TaskCompletionPrediction) *domain.ProjectCompletionPrediction {
	if len(taskPredictions) == 0 {
		return nil
	}

	// 
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

// findCriticalTasksForPrediction 
func (s *TaskSchedulingService) findCriticalTasksForPrediction(predictions []*domain.TaskCompletionPrediction) []uuid.UUID {
	// 
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].PredictedCompletion.After(predictions[j].PredictedCompletion)
	})

	criticalTasks := make([]uuid.UUID, 0)
	for i := 0; i < len(predictions) && i < 3; i++ {
		criticalTasks = append(criticalTasks, predictions[i].TaskID)
	}

	return criticalTasks
}

// analyzeCriticalPathRisks 
func (s *TaskSchedulingService) analyzeCriticalPathRisks(criticalPath *domain.CriticalPath, tasks []*domain.Task) []*domain.CriticalPathRisk {
	risks := make([]*domain.CriticalPathRisk, 0)

	// ?
	for _, criticalTask := range criticalPath.Tasks {
		// ?
		for _, task := range tasks {
			if task.ID == criticalTask.TaskID {
				if task.Complexity == domain.TaskComplexityVeryHigh {
					risks = append(risks, &domain.CriticalPathRisk{
						TaskID:      task.ID,
						RiskType:    "?,
						Description: "",
						Probability: 0.7,
						Impact:      "?,
					})
				}
				break
			}
		}
	}

	return risks
}

// generateCriticalPathOptimizations 
func (s *TaskSchedulingService) generateCriticalPathOptimizations(criticalPath *domain.CriticalPath, tasks []*domain.Task) []*domain.CriticalPathOptimization {
	optimizations := make([]*domain.CriticalPathOptimization, 0)

	// ?
	optimizations = append(optimizations, &domain.CriticalPathOptimization{
		Type:        "",
		Description: "?,
		Impact:      "",
		Priority:    "?,
	})

	// 鲢
	optimizations = append(optimizations, &domain.CriticalPathOptimization{
		Type:        "?,
		Description: "?,
		Impact:      "",
		Priority:    "?,
	})

	return optimizations
}

// calculateTotalDuration ?
func (s *TaskSchedulingService) calculateTotalDuration(criticalTasks []*domain.CriticalPathTask, taskMap map[uuid.UUID]*domain.Task) time.Duration {
	totalDuration := time.Duration(0)
	for _, criticalTask := range criticalTasks {
		if task, exists := taskMap[criticalTask.TaskID]; exists {
			totalDuration += s.calculateTaskDuration(task)
		}
	}
	return totalDuration
}

// findEarliestStartDate 翪?
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

// findLatestEndDate ?
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

