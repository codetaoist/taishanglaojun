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

// TaskAllocationService 
type TaskAllocationService struct {
	taskRepo    domain.TaskRepository
	teamRepo    domain.TeamRepository
	projectRepo domain.ProjectRepository
}

// NewTaskAllocationService 
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

// AllocateTasks 
func (s *TaskAllocationService) AllocateTasks(ctx context.Context, req *domain.TaskAllocationRequest) (*domain.TaskAllocationResult, error) {
	// 
	tasks, err := s.getUnassignedTasks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get unassigned tasks: %w", err)
	}

	// ?
	members, err := s.getAvailableMembers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get available members: %w", err)
	}

	// 
	assignments, err := s.allocateByStrategy(ctx, tasks, members, req.Strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate tasks: %w", err)
	}

	// 
	summary := s.generateAllocationSummary(assignments, tasks, members)

	return &domain.TaskAllocationResult{
		Assignments: assignments,
		Summary:     summary,
		Strategy:    req.Strategy,
		Timestamp:   time.Now(),
	}, nil
}

// RecommendAssignee ?
func (s *TaskAllocationService) RecommendAssignee(ctx context.Context, req *domain.TaskAssigneeRecommendationRequest) (*domain.TaskAssigneeRecommendationResult, error) {
	// 
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 
	candidates, err := s.getCandidates(ctx, task, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	// 
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

	// ?
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return &domain.TaskAssigneeRecommendationResult{
		TaskID:          req.TaskID,
		Recommendations: recommendations,
		Timestamp:       time.Now(),
	}, nil
}

// ReallocateTasks 
func (s *TaskAllocationService) ReallocateTasks(ctx context.Context, req *domain.TaskReallocationRequest) (*domain.TaskReallocationResult, error) {
	// 
	tasks, err := s.getTasksForReallocation(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for reallocation: %w", err)
	}

	// 
	members, err := s.getAvailableMembers(ctx, &domain.TaskAllocationRequest{
		OrganizationID: req.OrganizationID,
		TeamID:         req.TeamID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get available members: %w", err)
	}

	// 
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

// OptimizeWorkload 
func (s *TaskAllocationService) OptimizeWorkload(ctx context.Context, req *domain.WorkloadOptimizationRequest) (*domain.WorkloadOptimizationResult, error) {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 
	workloadAnalysis, err := s.analyzeCurrentWorkload(ctx, team)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze workload: %w", err)
	}

	// 
	optimizations := s.generateOptimizations(workloadAnalysis, req.Strategy)

	return &domain.WorkloadOptimizationResult{
		TeamID:        req.TeamID,
		Strategy:      req.Strategy,
		Optimizations: optimizations,
		Timestamp:     time.Now(),
	}, nil
}

// AnalyzeTeamWorkload 
func (s *TaskAllocationService) AnalyzeTeamWorkload(ctx context.Context, req *domain.TeamWorkloadRequest) (*domain.TeamWorkloadReport, error) {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// ?
	memberWorkloads := make([]*domain.MemberWorkload, 0, len(team.Members))
	for _, member := range team.Members {
		workload, err := s.analyzeMemberWorkload(ctx, member.UserID)
		if err != nil {
			continue // ?
		}
		memberWorkloads = append(memberWorkloads, workload)
	}

	// 
	teamMetrics := s.calculateTeamWorkloadMetrics(memberWorkloads)

	return &domain.TeamWorkloadReport{
		TeamID:          req.TeamID,
		MemberWorkloads: memberWorkloads,
		TeamMetrics:     teamMetrics,
		Timestamp:       time.Now(),
	}, nil
}

// ==========  ==========

// getUnassignedTasks 
func (s *TaskAllocationService) getUnassignedTasks(ctx context.Context, req *domain.TaskAllocationRequest) ([]*domain.Task, error) {
	filters := map[string]interface{}{
		"organization_id": req.OrganizationID,
		"assignee_id":     nil, // ?
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

// getAvailableMembers ?
func (s *TaskAllocationService) getAvailableMembers(ctx context.Context, req *domain.TaskAllocationRequest) ([]*domain.TeamMember, error) {
	var members []*domain.TeamMember

	if req.TeamID != nil {
		// ?
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
		// ?
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

// allocateByStrategy 
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

// allocateBalanced 
func (s *TaskAllocationService) allocateBalanced(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember) ([]*domain.TaskAllocationAssignment, error) {
	assignments := make([]*domain.TaskAllocationAssignment, 0)
	memberTaskCount := make(map[uuid.UUID]int)

	// ?
	for _, member := range members {
		memberTaskCount[member.UserID] = 0
	}

	// 
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
				Reason:     "",
			})
			memberTaskCount[bestMember.UserID]++
		}
	}

	return assignments, nil
}

// allocateSkillBased ?
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
				Reason:     "?,
			})
		}
	}

	return assignments, nil
}

// allocatePriorityBased ?
func (s *TaskAllocationService) allocatePriorityBased(ctx context.Context, tasks []*domain.Task, members []*domain.TeamMember) ([]*domain.TaskAllocationAssignment, error) {
	// 
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
				Reason:     "?,
			})
		}
	}

	return assignments, nil
}

// allocateWorkloadBased 
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
				Reason:     "",
			})
		}
	}

	return assignments, nil
}

// calculateAssignmentScore 
func (s *TaskAllocationService) calculateAssignmentScore(task *domain.Task, member *domain.TeamMember) float64 {
	score := 0.0

	// ?
	score += member.Availability * 0.3

	// ?
	score += s.getPriorityWeight(task.Priority) * 0.2

	// ?
	score += s.getComplexityWeight(task.Complexity) * 0.2

	// 
	score += s.getTypeWeight(task.Type) * 0.3

	return score
}

// calculateSkillMatchScore 㼼?
func (s *TaskAllocationService) calculateSkillMatchScore(task *domain.Task, member *domain.TeamMember) float64 {
	// ?
	baseScore := s.calculateAssignmentScore(task, member)
	
	// ?
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

// calculatePriorityAssignmentScore ?
func (s *TaskAllocationService) calculatePriorityAssignmentScore(task *domain.Task, member *domain.TeamMember) float64 {
	baseScore := s.calculateAssignmentScore(task, member)
	priorityBonus := s.getPriorityWeight(task.Priority) * 0.5
	return baseScore + priorityBonus
}

// calculateWorkloadScore 㹤
func (s *TaskAllocationService) calculateWorkloadScore(task *domain.Task, member *domain.TeamMember, workload *domain.MemberWorkload) float64 {
	baseScore := s.calculateAssignmentScore(task, member)
	
	// ?
	workloadPenalty := 0.0
	if workload != nil {
		workloadPenalty = workload.CurrentLoad * 0.5
	}
	
	return baseScore - workloadPenalty
}

// getPriorityWeight ?
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

// getComplexityWeight ?
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

// getTypeWeight 
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

// generateAllocationSummary 
func (s *TaskAllocationService) generateAllocationSummary(assignments []*domain.TaskAllocationAssignment, tasks []*domain.Task, members []*domain.TeamMember) *domain.TaskAllocationSummary {
	summary := &domain.TaskAllocationSummary{
		TotalTasks:     len(tasks),
		AssignedTasks:  len(assignments),
		UnassignedTasks: len(tasks) - len(assignments),
		TotalMembers:   len(members),
	}

	// 
	if len(assignments) > 0 {
		totalScore := 0.0
		for _, assignment := range assignments {
			totalScore += assignment.Score
		}
		summary.AverageScore = totalScore / float64(len(assignments))
	}

	// 
	memberAssignments := make(map[uuid.UUID]int)
	for _, assignment := range assignments {
		memberAssignments[assignment.AssigneeID]++
	}

	summary.MemberAssignments = memberAssignments

	return summary
}

// getCandidates 
func (s *TaskAllocationService) getCandidates(ctx context.Context, task *domain.Task, teamID *uuid.UUID) ([]*domain.TeamMember, error) {
	if teamID != nil {
		team, err := s.teamRepo.FindByID(ctx, *teamID)
		if err != nil {
			return nil, err
		}
		return team.Members, nil
	}

	// 
	if task.ProjectID != nil {
		// 
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

// calculateRecommendationScore 
func (s *TaskAllocationService) calculateRecommendationScore(ctx context.Context, task *domain.Task, member *domain.TeamMember) (float64, map[string]float64) {
	factors := make(map[string]float64)

	// ?
	factors["availability"] = member.Availability * 0.25

	// ?
	factors["skill_match"] = s.calculateSkillMatchScore(task, member) * 0.3

	// 
	workload, _ := s.analyzeMemberWorkload(ctx, member.UserID)
	workloadScore := 1.0
	if workload != nil {
		workloadScore = math.Max(0, 1.0-workload.CurrentLoad)
	}
	factors["workload"] = workloadScore * 0.25

	// 
	factors["performance"] = 0.8 * 0.2 // ?

	// 
	totalScore := 0.0
	for _, score := range factors {
		totalScore += score
	}

	return totalScore, factors
}

// calculateConfidence ?
func (s *TaskAllocationService) calculateConfidence(score float64, factors map[string]float64) float64 {
	// ?
	variance := 0.0
	mean := score
	
	for _, factor := range factors {
		variance += math.Pow(factor-mean, 2)
	}
	variance /= float64(len(factors))
	
	// ?
	confidence := 1.0 / (1.0 + variance)
	return math.Min(1.0, confidence)
}

// getTasksForReallocation 
func (s *TaskAllocationService) getTasksForReallocation(ctx context.Context, req *domain.TaskReallocationRequest) ([]*domain.Task, error) {
	filters := map[string]interface{}{
		"organization_id": req.OrganizationID,
		"status":          string(domain.TaskStatusInProgress),
	}

	if req.TeamID != nil {
		// ?
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

// findBetterAssignee ?
func (s *TaskAllocationService) findBetterAssignee(ctx context.Context, task *domain.Task, members []*domain.TeamMember) *domain.TeamMember {
	if task.AssigneeID == nil {
		return nil
	}

	currentScore := 0.0
	var currentMember *domain.TeamMember

	// ?
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

	// ?
	var bestMember *domain.TeamMember
	bestScore := currentScore

	for _, member := range members {
		if member.UserID == *task.AssigneeID {
			continue // ?
		}

		score := s.calculateAssignmentScore(task, member)
		if score > bestScore+0.1 { // 
			bestScore = score
			bestMember = member
		}
	}

	return bestMember
}

// getReallocationReason 
func (s *TaskAllocationService) getReallocationReason(task *domain.Task, newMember *domain.TeamMember) string {
	return fmt.Sprintf("? %.1f%%", newMember.Availability)
}

// calculateReallocationScore 
func (s *TaskAllocationService) calculateReallocationScore(task *domain.Task, newMember *domain.TeamMember) float64 {
	return s.calculateAssignmentScore(task, newMember)
}

// analyzeCurrentWorkload 
func (s *TaskAllocationService) analyzeCurrentWorkload(ctx context.Context, team *domain.Team) (*domain.TeamWorkloadAnalysis, error) {
	analysis := &domain.TeamWorkloadAnalysis{
		TeamID:    team.ID,
		Timestamp: time.Now(),
	}

	// ?
	for _, member := range team.Members {
		workload, err := s.analyzeMemberWorkload(ctx, member.UserID)
		if err != nil {
			continue
		}
		analysis.MemberWorkloads = append(analysis.MemberWorkloads, workload)
	}

	// 
	analysis.TeamMetrics = s.calculateTeamWorkloadMetrics(analysis.MemberWorkloads)

	return analysis, nil
}

// generateOptimizations 
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

// generateBalanceOptimizations 
func (s *TaskAllocationService) generateBalanceOptimizations(analysis *domain.TeamWorkloadAnalysis) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	// 
	var maxMember, minMember *domain.MemberWorkload
	for _, workload := range analysis.MemberWorkloads {
		if maxMember == nil || workload.CurrentLoad > maxMember.CurrentLoad {
			maxMember = workload
		}
		if minMember == nil || workload.CurrentLoad < minMember.CurrentLoad {
			minMember = workload
		}
	}

	// ?
	if maxMember != nil && minMember != nil && maxMember.CurrentLoad-minMember.CurrentLoad > 0.3 {
		optimizations = append(optimizations, &domain.WorkloadOptimization{
			Type:        domain.OptimizationTypeRebalance,
			Description: fmt.Sprintf(" %s ?%s", maxMember.UserID, minMember.UserID),
			Impact:      "",
			Priority:    domain.OptimizationPriorityMedium,
		})
	}

	return optimizations
}

// generateEfficiencyOptimizations 
func (s *TaskAllocationService) generateEfficiencyOptimizations(analysis *domain.TeamWorkloadAnalysis) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	// 
	for _, workload := range analysis.MemberWorkloads {
		if workload.CurrentLoad > 0.9 {
			optimizations = append(optimizations, &domain.WorkloadOptimization{
				Type:        domain.OptimizationTypeReduce,
				Description: fmt.Sprintf(" %s ?, workload.UserID),
				Impact:      "?,
				Priority:    domain.OptimizationPriorityHigh,
			})
		}
	}

	return optimizations
}

// generateCapacityOptimizations 
func (s *TaskAllocationService) generateCapacityOptimizations(analysis *domain.TeamWorkloadAnalysis) []*domain.WorkloadOptimization {
	optimizations := make([]*domain.WorkloadOptimization, 0)

	// ?
	if analysis.TeamMetrics.AverageLoad < 0.6 {
		optimizations = append(optimizations, &domain.WorkloadOptimization{
			Type:        domain.OptimizationTypeIncrease,
			Description: "幤?,
			Impact:      "?,
			Priority:    domain.OptimizationPriorityLow,
		})
	}

	return optimizations
}

// analyzeMemberWorkload 
func (s *TaskAllocationService) analyzeMemberWorkload(ctx context.Context, userID uuid.UUID) (*domain.MemberWorkload, error) {
	// ?
	tasks, err := s.taskRepo.FindByAssignee(ctx, userID, 100, 0)
	if err != nil {
		return nil, err
	}

	// 㹤
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

	// 㸺40?
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

// calculateTeamWorkloadMetrics 
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
	
	// 㸺
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
		BalanceScore:   1.0 - variance, // ?
	}
}

