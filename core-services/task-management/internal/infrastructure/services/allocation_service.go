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

// TaskAllocationService 任务分配服务实现
type TaskAllocationService struct {
	taskRepo    domain.TaskRepository
	teamRepo    domain.TeamRepository
	projectRepo domain.ProjectRepository
}

// NewTaskAllocationService 创建任务分配服务
func NewTaskAllocationService(
	taskRepo domain.TaskRepository,
	teamRepo domain.TeamRepository,
	projectRepo domain.ProjectRepository,
) domain.TaskAllocationService {
	return &TaskAllocationService{
		taskRepo:    taskRepo,
		teamRepo:    teamRepo,
		projectRepo: projectRepo,
	}
}

// AllocateTasks 分配任务
func (s *TaskAllocationService) AllocateTasks(ctx context.Context, req *domain.TaskAllocationRequest) (*domain.TaskAllocationResult, error) {
	// 获取待分配的任务
	tasks, err := s.getUnassignedTasks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get unassigned tasks: %w", err)
	}

	// 获取可用的团队成�?
	members, err := s.getAvailableMembers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get available members: %w", err)
	}

	// 根据策略分配任务
	assignments, err := s.allocateByStrategy(ctx, tasks, members, req.Strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate tasks: %w", err)
	}

	// 生成分配摘要
	summary := s.generateAllocationSummary(assignments, tasks, members)

	return &domain.TaskAllocationResult{
		Assignments: assignments,
		Summary:     summary,
		Strategy:    req.Strategy,
		Timestamp:   time.Now(),
	}, nil
}

// RecommendAssignee 推荐任务分配�?
func (s *TaskAllocationService) RecommendAssignee(ctx context.Context, req *domain.TaskAssigneeRecommendationRequest) (*domain.TaskAssigneeRecommendationResult, error) {
	// 获取任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 获取候选人
	candidates, err := s.getCandidates(ctx, task, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// 计算推荐分数
	recommendations := make([]*domain.AssigneeRecommendation, 0, len(candidates))
	for _, candidate := range candidates {
		score, factors := s.calculateRecommendationScore(ctx, task, candidate)
		recommendations = append(recommendations, &domain.AssigneeRecommendation{
			UserID:     candidate.UserID,
			Score:      score,
			Factors:    factors,
			Confidence: s.calculateConfidence(score, factors),
		})
	}

	// 按分数排�?
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return &domain.TaskAssigneeRecommendationResult{
		TaskID:          req.TaskID,
		Recommendations: recommendations,
		Timestamp:       time.Now(),
	}, nil
}

// ReallocateTasks 重新分配任务
func (s *TaskAllocationService) ReallocateTasks(ctx context.Context, req *domain.TaskReallocationRequest) (*domain.TaskReallocationResult, error) {
	// 获取需要重新分配的任务
	tasks, err := s.getTasksForReallocation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for reallocation: %w", err)
	}

	// 获取可用成员
	members, err := s.getAvailableMembers(ctx, &domain.TaskAllocationRequest{
		OrganizationID: req.OrganizationID,
		TeamID:         req.TeamID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get available members: %w", err)
	}

	// 执行重新分配
	reallocations := make([]*domain.TaskReallocation, 0)
	for _, task := range tasks {
		if newAssignee := s.findBetterAssignee(ctx, task, members); newAssignee != nil {
			reallocations = append(reallocations, &domain.TaskReallocation{
				TaskID:        task.ID,
				OldAssigneeID: task.AssigneeID,
				NewAssigneeID: &newAssignee.UserID,
				Reason:        s.getReallocationReason(task, newAssignee),
				Score:         s.calculateReallocationScore(task, newAssignee),
			})
		}
	}

	return &domain.TaskReallocationResult{
		Reallocations: reallocations,
		Timestamp:     time.Now(),
	}, nil
}

// OptimizeWorkload 优化工作负载
func (s *TaskAllocationService) OptimizeWorkload(ctx context.Context, req *domain.WorkloadOptimizationRequest) (*domain.WorkloadOptimizationResult, error) {
	// 获取团队成员
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 分析当前工作负载
	workloadAnalysis, err := s.analyzeCurrentWorkload(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze workload: %w", err)
	}

	// 生成优化建议
	optimizations := s.generateOptimizations(workloadAnalysis, req.Strategy)

	return &domain.WorkloadOptimizationResult{
		TeamID:        req.TeamID,
		Strategy:      req.Strategy,
		Optimizations: optimizations,
		Timestamp:     time.Now(),
	}, nil
}

// AnalyzeTeamWorkload 分析团队工作负载
func (s *TaskAllocationService) AnalyzeTeamWorkload(ctx context.Context, req *domain.TeamWorkloadRequest) (*domain.TeamWorkloadReport, error) {
	// 获取团队
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 分析每个成员的工作负�?
	memberWorkloads := make([]*domain.MemberWorkload, 0, len(team.Members))
	for _, member := range team.Members {
		workload, err := s.analyzeMemberWorkload(ctx, member.UserID)
		if err != nil {
			continue // 跳过错误，继续处理其他成�?
		}
		memberWorkloads = append(memberWorkloads, workload)
	}

	// 计算团队整体指标
	teamMetrics := s.calculateTeamWorkloadMetrics(memberWorkloads)

	return &domain.TeamWorkloadReport{
		TeamID:          req.TeamID,
		MemberWorkloads: memberWorkloads,
		TeamMetrics:     teamMetrics,
		Timestamp:       time.Now(),
	}, nil
}

// ========== 私有辅助方法 ==========

// getUnassignedTasks 获取未分配的任务
func (s *TaskAllocationService) getUnassignedTasks(ctx context.Context, req *domain.TaskAllocationRequest) ([]*domain.Task, error) {
	filters := map[string]interface{}{
		"organization_id": req.OrganizationID,
		"assignee_id":     nil, // 未分�?
		"status":          string(domain.TaskStatusPending),
	}

	if req.ProjectID != nil {
		filters["project_id"] = *req.ProjectID
	}

	tasks, err := s.taskRepo.SearchTasks(ctx, "", filters, req.MaxTasks, 0)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// getAvailableMembers 获取可用的团队成�?
func (s *TaskAllocationService) getAvailableMembers(ctx context.Context, req *domain.TaskAllocationRequest) ([]*domain.TeamMember, error) {
	var members []*domain.TeamMember

	if req.TeamID != nil {
		// 获取指定团队的成�?
		team, err := s.teamRepo.FindByID(ctx, *req.TeamID)
		if err != nil {
			return nil, err
		}
		
		for _, member := range team.Members {
			if member.IsAvailable {
				members = append(members, member)
			}
		}
	} else {
		// 获取组织内所有可用成�?
		teams, err := s.teamRepo.FindByOrganization(ctx, req.OrganizationID, 100, 0)
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			for _, member := range team.Members {
				if member.IsAvailable {
					members = append(members, member)
				}
			}
		}
	}

	return members, nil
}

// allocateByStrategy 根据策略分配任务
func (s *TaskAllocationService) allocateByStrategy(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember, strategy domain.TaskAllocationStrategy) ([]*domain.TaskAllocationAssignment, error) {
	switch strategy {
	case domain.AllocationStrategyBalanced:
		return s.allocateBalanced(ctx, tasks, members)
	case domain.AllocationStrategySkillBased:
		return s.allocateSkillBased(ctx, tasks, members)
	case domain.AllocationStrategyPriorityBased:
		return s.allocatePriorityBased(ctx, tasks, members)
	case domain.AllocationStrategyWorkloadBased:
		return s.allocateWorkloadBased(ctx, tasks, members)
	default:
		return s.allocateBalanced(ctx, tasks, members)
	}
}

// allocateBalanced 平衡分配
func (s *TaskAllocationService) allocateBalanced(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember) ([]*domain.TaskAllocationAssignment, error) {
	assignments := make([]*domain.TaskAllocationAssignment, 0)
	memberTaskCount := make(map[uuid.UUID]int)

	// 初始化成员任务计�?
	for _, member := range members {
		memberTaskCount[member.UserID] = 0
	}

	// 为每个任务找到任务最少的成员
	for _, task := range tasks {
		var bestMember *domain.TeamMember
		minTasks := math.MaxInt32

		for _, member := range members {
			if memberTaskCount[member.UserID] < minTasks {
				minTasks = memberTaskCount[member.UserID]
				bestMember = member
			}
		}

		if bestMember != nil {
			assignments = append(assignments, &domain.TaskAllocationAssignment{
				TaskID:     task.ID,
				AssigneeID: bestMember.UserID,
				Score:      s.calculateAssignmentScore(task, bestMember),
				Reason:     "平衡工作负载",
			})
			memberTaskCount[bestMember.UserID]++
		}
	}

	return assignments, nil
}

// allocateSkillBased 基于技能分�?
func (s *TaskAllocationService) allocateSkillBased(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember) ([]*domain.TaskAllocationAssignment, error) {
	assignments := make([]*domain.TaskAllocationAssignment, 0)

	for _, task := range tasks {
		var bestMember *domain.TeamMember
		var bestScore float64

		for _, member := range members {
			score := s.calculateSkillMatchScore(task, member)
			if score > bestScore {
				bestScore = score
				bestMember = member
			}
		}

		if bestMember != nil {
			assignments = append(assignments, &domain.TaskAllocationAssignment{
				TaskID:     task.ID,
				AssigneeID: bestMember.UserID,
				Score:      bestScore,
				Reason:     "技能匹�?,
			})
		}
	}

	return assignments, nil
}

// allocatePriorityBased 基于优先级分�?
func (s *TaskAllocationService) allocatePriorityBased(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember) ([]*domain.TaskAllocationAssignment, error) {
	// 按优先级排序任务
	sort.Slice(tasks, func(i, j int) bool {
		return s.getPriorityWeight(tasks[i].Priority) > s.getPriorityWeight(tasks[j].Priority)
	})

	assignments := make([]*domain.TaskAllocationAssignment, 0)

	for _, task := range tasks {
		var bestMember *domain.TeamMember
		var bestScore float64

		for _, member := range members {
			score := s.calculatePriorityAssignmentScore(task, member)
			if score > bestScore {
				bestScore = score
				bestMember = member
			}
		}

		if bestMember != nil {
			assignments = append(assignments, &domain.TaskAllocationAssignment{
				TaskID:     task.ID,
				AssigneeID: bestMember.UserID,
				Score:      bestScore,
				Reason:     "优先级匹�?,
			})
		}
	}

	return assignments, nil
}

// allocateWorkloadBased 基于工作负载分配
func (s *TaskAllocationService) allocateWorkloadBased(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember) ([]*domain.TaskAllocationAssignment, error) {
	assignments := make([]*domain.TaskAllocationAssignment, 0)

	for _, task := range tasks {
		var bestMember *domain.TeamMember
		var bestScore float64

		for _, member := range members {
			workload, _ := s.analyzeMemberWorkload(ctx, member.UserID)
			score := s.calculateWorkloadScore(task, member, workload)
			if score > bestScore {
				bestScore = score
				bestMember = member
			}
		}

		if bestMember != nil {
			assignments = append(assignments, &domain.TaskAllocationAssignment{
				TaskID:     task.ID,
				AssigneeID: bestMember.UserID,
				Score:      bestScore,
				Reason:     "工作负载优化",
			})
		}
	}

	return assignments, nil
}

// calculateAssignmentScore 计算分配分数
func (s *TaskAllocationService) calculateAssignmentScore(task *domain.Task, member *domain.TeamMember) float64 {
	score := 0.0

	// 可用性权�?
	score += member.Availability * 0.3

	// 优先级匹配权�?
	score += s.getPriorityWeight(task.Priority) * 0.2

	// 复杂度匹配权�?
	score += s.getComplexityWeight(task.Complexity) * 0.2

	// 类型匹配权重
	score += s.getTypeWeight(task.Type) * 0.3

	return score
}

// calculateSkillMatchScore 计算技能匹配分�?
func (s *TaskAllocationService) calculateSkillMatchScore(task *domain.Task, member *domain.TeamMember) float64 {
	// 这里简化处理，实际应该根据任务需要的技能和成员的技能进行匹�?
	baseScore := s.calculateAssignmentScore(task, member)
	
	// 技能匹配加权（这里简化为根据任务类型匹配�?
	skillBonus := 0.0
	switch task.Type {
	case domain.TaskTypeDevelopment:
		skillBonus = 0.8
	case domain.TaskTypeDesign:
		skillBonus = 0.6
	case domain.TaskTypeTesting:
		skillBonus = 0.7
	case domain.TaskTypeDocumentation:
		skillBonus = 0.5
	}

	return baseScore + skillBonus
}

// calculatePriorityAssignmentScore 计算优先级分配分�?
func (s *TaskAllocationService) calculatePriorityAssignmentScore(task *domain.Task, member *domain.TeamMember) float64 {
	baseScore := s.calculateAssignmentScore(task, member)
	priorityBonus := s.getPriorityWeight(task.Priority) * 0.5
	return baseScore + priorityBonus
}

// calculateWorkloadScore 计算工作负载分数
func (s *TaskAllocationService) calculateWorkloadScore(task *domain.Task, member *domain.TeamMember, workload *domain.MemberWorkload) float64 {
	baseScore := s.calculateAssignmentScore(task, member)
	
	// 工作负载越低，分数越�?
	workloadPenalty := 0.0
	if workload != nil {
		workloadPenalty = workload.CurrentLoad * 0.5
	}
	
	return baseScore - workloadPenalty
}

// getPriorityWeight 获取优先级权�?
func (s *TaskAllocationService) getPriorityWeight(priority domain.TaskPriority) float64 {
	switch priority {
	case domain.TaskPriorityUrgent:
		return 1.0
	case domain.TaskPriorityHigh:
		return 0.8
	case domain.TaskPriorityMedium:
		return 0.6
	case domain.TaskPriorityLow:
		return 0.4
	default:
		return 0.5
	}
}

// getComplexityWeight 获取复杂度权�?
func (s *TaskAllocationService) getComplexityWeight(complexity domain.TaskComplexity) float64 {
	switch complexity {
	case domain.TaskComplexityVeryHigh:
		return 1.0
	case domain.TaskComplexityHigh:
		return 0.8
	case domain.TaskComplexityMedium:
		return 0.6
	case domain.TaskComplexityLow:
		return 0.4
	case domain.TaskComplexityVeryLow:
		return 0.2
	default:
		return 0.5
	}
}

// getTypeWeight 获取类型权重
func (s *TaskAllocationService) getTypeWeight(taskType domain.TaskType) float64 {
	switch taskType {
	case domain.TaskTypeDevelopment:
		return 0.9
	case domain.TaskTypeDesign:
		return 0.7
	case domain.TaskTypeTesting:
		return 0.8
	case domain.TaskTypeDocumentation:
		return 0.5
	case domain.TaskTypeResearch:
		return 0.6
	case domain.TaskTypeMeeting:
		return 0.3
	case domain.TaskTypeReview:
		return 0.6
	default:
		return 0.5
	}
}

// generateAllocationSummary 生成分配摘要
func (s *TaskAllocationService) generateAllocationSummary(assignments []*domain.TaskAllocationAssignment, tasks []*domain.Task, members []*domain.TeamMember) *domain.TaskAllocationSummary {
	summary := &domain.TaskAllocationSummary{
		TotalTasks:     len(tasks),
		AssignedTasks:  len(assignments),
		UnassignedTasks: len(tasks) - len(assignments),
		TotalMembers:   len(members),
	}

	// 计算平均分数
	if len(assignments) > 0 {
		totalScore := 0.0
		for _, assignment := range assignments {
			totalScore += assignment.Score
		}
		summary.AverageScore = totalScore / float64(len(assignments))
	}

	// 计算成员分配统计
	memberAssignments := make(map[uuid.UUID]int)
	for _, assignment := range assignments {
		memberAssignments[assignment.AssigneeID]++
	}

	summary.MemberAssignments = memberAssignments

	return summary
}

// getCandidates 获取候选人
func (s *TaskAllocationService) getCandidates(ctx context.Context, task *domain.Task, teamID *uuid.UUID) ([]*domain.TeamMember, error) {
	if teamID != nil {
		team, err := s.teamRepo.FindByID(ctx, *teamID)
		if err != nil {
			return nil, err
		}
		return team.Members, nil
	}

	// 如果没有指定团队，获取项目相关的团队成员
	if task.ProjectID != nil {
		// 这里简化处理，实际应该根据项目获取相关团队
		teams, err := s.teamRepo.FindByOrganization(ctx, task.OrganizationID, 10, 0)
		if err != nil {
			return nil, err
		}

		var candidates []*domain.TeamMember
		for _, team := range teams {
			candidates = append(candidates, team.Members...)
		}
		return candidates, nil
	}

	return nil, fmt.Errorf("no candidates found")
}

// calculateRecommendationScore 计算推荐分数
func (s *TaskAllocationService) calculateRecommendationScore(ctx context.Context, task *domain.Task, member *domain.TeamMember) (float64, map[string]float64) {
	factors := make(map[string]float64)

	// 可用性因�?
	factors["availability"] = member.Availability * 0.25

	// 技能匹配因�?
	factors["skill_match"] = s.calculateSkillMatchScore(task, member) * 0.3

	// 工作负载因子
	workload, _ := s.analyzeMemberWorkload(ctx, member.UserID)
	workloadScore := 1.0
	if workload != nil {
		workloadScore = math.Max(0, 1.0-workload.CurrentLoad)
	}
	factors["workload"] = workloadScore * 0.25

	// 历史表现因子
	factors["performance"] = 0.8 * 0.2 // 简化处理，实际应该从历史数据计�?

	// 计算总分
	totalScore := 0.0
	for _, score := range factors {
		totalScore += score
	}

	return totalScore, factors
}

// calculateConfidence 计算置信�?
func (s *TaskAllocationService) calculateConfidence(score float64, factors map[string]float64) float64 {
	// 基于分数和因子的方差计算置信�?
	variance := 0.0
	mean := score
	
	for _, factor := range factors {
		variance += math.Pow(factor-mean, 2)
	}
	variance /= float64(len(factors))
	
	// 置信度与方差成反�?
	confidence := 1.0 / (1.0 + variance)
	return math.Min(1.0, confidence)
}

// getTasksForReallocation 获取需要重新分配的任务
func (s *TaskAllocationService) getTasksForReallocation(ctx context.Context, req *domain.TaskReallocationRequest) ([]*domain.Task, error) {
	filters := map[string]interface{}{
		"organization_id": req.OrganizationID,
		"status":          string(domain.TaskStatusInProgress),
	}

	if req.TeamID != nil {
		// 获取团队成员的任�?
		team, err := s.teamRepo.FindByID(ctx, *req.TeamID)
		if err != nil {
			return nil, err
		}

		var memberIDs []uuid.UUID
		for _, member := range team.Members {
			memberIDs = append(memberIDs, member.UserID)
		}
		filters["assignee_ids"] = memberIDs
	}

	tasks, err := s.taskRepo.SearchTasks(ctx, "", filters, 100, 0)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// findBetterAssignee 找到更好的分配�?
func (s *TaskAllocationService) findBetterAssignee(ctx context.Context, task *domain.Task, members []*domain.TeamMember) *domain.TeamMember {
	if task.AssigneeID == nil {
		return nil
	}

	currentScore := 0.0
	var currentMember *domain.TeamMember

	// 找到当前分配�?
	for _, member := range members {
		if member.UserID == *task.AssigneeID {
			currentMember = member
			currentScore = s.calculateAssignmentScore(task, member)
			break
		}
	}

	if currentMember == nil {
		return nil
	}

	// 找到更好的候选�?
	var bestMember *domain.TeamMember
	bestScore := currentScore

	for _, member := range members {
		if member.UserID == *task.AssigneeID {
			continue // 跳过当前分配�?
		}

		score := s.calculateAssignmentScore(task, member)
		if score > bestScore+0.1 { // 需要显著更好才重新分配
			bestScore = score
			bestMember = member
		}
	}

	return bestMember
}

// getReallocationReason 获取重新分配原因
func (s *TaskAllocationService) getReallocationReason(task *domain.Task, newMember *domain.TeamMember) string {
	return fmt.Sprintf("找到更适合的分配者，可用�? %.1f%%", newMember.Availability)
}

// calculateReallocationScore 计算重新分配分数
func (s *TaskAllocationService) calculateReallocationScore(task *domain.Task, newMember *domain.TeamMember) float64 {
	return s.calculateAssignmentScore(task, newMember)
}

// analyzeCurrentWorkload 分析当前工作负载
func (s *TaskAllocationService) analyzeCurrentWorkload(ctx context.Context, team *domain.Team) (*domain.TeamWorkloadAnalysis, error) {
	analysis := &domain.TeamWorkloadAnalysis{
		TeamID:    team.ID,
		Timestamp: time.Now(),
	}

	// 分析每个成员的工作负�?
	for _, member := range team.Members {
		workload, err := s.analyzeMemberWorkload(ctx, member.UserID)
		if err != nil {
			continue
		}
		analysis.MemberWorkloads = append(analysis.MemberWorkloads, workload)
	}

	// 计算团队指标
	analysis.TeamMetrics = s.calculateTeamWorkloadMetrics(analysis.MemberWorkloads)

	return analysis, nil
}

// generateOptimizations 生成优化建议
func (s *TaskAllocationService) generateOptimizations(analysis *domain.TeamWorkloadAnalysis, strategy domain.WorkloadOptimizationStrategy) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	switch strategy {
	case domain.OptimizationStrategyBalance:
		optimizations = append(optimizations, s.generateBalanceOptimizations(analysis)...)
	case domain.OptimizationStrategyEfficiency:
		optimizations = append(optimizations, s.generateEfficiencyOptimizations(analysis)...)
	case domain.OptimizationStrategyCapacity:
		optimizations = append(optimizations, s.generateCapacityOptimizations(analysis)...)
	}

	return optimizations
}

// generateBalanceOptimizations 生成平衡优化建议
func (s *TaskAllocationService) generateBalanceOptimizations(analysis *domain.TeamWorkloadAnalysis) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	// 找到工作负载最高和最低的成员
	var maxMember, minMember *domain.MemberWorkload
	for _, workload := range analysis.MemberWorkloads {
		if maxMember == nil || workload.CurrentLoad > maxMember.CurrentLoad {
			maxMember = workload
		}
		if minMember == nil || workload.CurrentLoad < minMember.CurrentLoad {
			minMember = workload
		}
	}

	// 如果差异较大，建议重新分�?
	if maxMember != nil && minMember != nil && maxMember.CurrentLoad-minMember.CurrentLoad > 0.3 {
		optimizations = append(optimizations, &domain.WorkloadOptimization{
			Type:        domain.OptimizationTypeRebalance,
			Description: fmt.Sprintf("将部分任务从 %s 重新分配�?%s", maxMember.UserID, minMember.UserID),
			Impact:      "平衡团队工作负载",
			Priority:    domain.OptimizationPriorityMedium,
		})
	}

	return optimizations
}

// generateEfficiencyOptimizations 生成效率优化建议
func (s *TaskAllocationService) generateEfficiencyOptimizations(analysis *domain.TeamWorkloadAnalysis) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	// 检查是否有成员工作负载过高
	for _, workload := range analysis.MemberWorkloads {
		if workload.CurrentLoad > 0.9 {
			optimizations = append(optimizations, &domain.WorkloadOptimization{
				Type:        domain.OptimizationTypeReduce,
				Description: fmt.Sprintf("减少成员 %s 的工作负�?, workload.UserID),
				Impact:      "提高工作效率，避免过度劳�?,
				Priority:    domain.OptimizationPriorityHigh,
			})
		}
	}

	return optimizations
}

// generateCapacityOptimizations 生成容量优化建议
func (s *TaskAllocationService) generateCapacityOptimizations(analysis *domain.TeamWorkloadAnalysis) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	// 检查团队整体容�?
	if analysis.TeamMetrics.AverageLoad < 0.6 {
		optimizations = append(optimizations, &domain.WorkloadOptimization{
			Type:        domain.OptimizationTypeIncrease,
			Description: "团队整体工作负载较低，可以承担更多任�?,
			Impact:      "提高团队产能利用�?,
			Priority:    domain.OptimizationPriorityLow,
		})
	}

	return optimizations
}

// analyzeMemberWorkload 分析成员工作负载
func (s *TaskAllocationService) analyzeMemberWorkload(ctx context.Context, userID uuid.UUID) (*domain.MemberWorkload, error) {
	// 获取成员的活跃任�?
	tasks, err := s.taskRepo.FindByAssignee(ctx, userID, 100, 0)
	if err != nil {
		return nil, err
	}

	// 计算工作负载
	totalHours := 0.0
	activeTasks := 0
	for _, task := range tasks {
		if task.Status == domain.TaskStatusInProgress || task.Status == domain.TaskStatusPending {
			activeTasks++
			if task.EstimatedHours != nil {
				totalHours += *task.EstimatedHours
			}
		}
	}

	// 计算负载百分比（假设每周40小时工作时间�?
	weeklyCapacity := 40.0
	currentLoad := math.Min(1.0, totalHours/weeklyCapacity)

	return &domain.MemberWorkload{
		UserID:      userID,
		ActiveTasks: activeTasks,
		TotalHours:  totalHours,
		CurrentLoad: currentLoad,
		Capacity:    weeklyCapacity,
	}, nil
}

// calculateTeamWorkloadMetrics 计算团队工作负载指标
func (s *TaskAllocationService) calculateTeamWorkloadMetrics(workloads []*domain.MemberWorkload) *domain.TeamWorkloadMetrics {
	if len(workloads) == 0 {
		return &domain.TeamWorkloadMetrics{}
	}

	totalLoad := 0.0
	maxLoad := 0.0
	minLoad := 1.0
	totalTasks := 0

	for _, workload := range workloads {
		totalLoad += workload.CurrentLoad
		totalTasks += workload.ActiveTasks
		if workload.CurrentLoad > maxLoad {
			maxLoad = workload.CurrentLoad
		}
		if workload.CurrentLoad < minLoad {
			minLoad = workload.CurrentLoad
		}
	}

	averageLoad := totalLoad / float64(len(workloads))
	
	// 计算负载方差
	variance := 0.0
	for _, workload := range workloads {
		variance += math.Pow(workload.CurrentLoad-averageLoad, 2)
	}
	variance /= float64(len(workloads))

	return &domain.TeamWorkloadMetrics{
		AverageLoad:    averageLoad,
		MaxLoad:        maxLoad,
		MinLoad:        minLoad,
		LoadVariance:   variance,
		TotalTasks:     totalTasks,
		BalanceScore:   1.0 - variance, // 方差越小，平衡性越�?
	}
}
