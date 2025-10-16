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

// PerformanceAnalysisService 
type PerformanceAnalysisService struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	teamRepo    domain.TeamRepository
}

// NewPerformanceAnalysisService 
func NewPerformanceAnalysisService(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	teamRepo domain.TeamRepository,
) domain.PerformanceAnalysisService {
	return &PerformanceAnalysisService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		teamRepo:    teamRepo,
	}
}

// AnalyzeUserPerformance 
func (s *PerformanceAnalysisService) AnalyzeUserPerformance(ctx context.Context, req *domain.UserPerformanceRequest) (*domain.UserPerformanceReport, error) {
	// ?
	tasks, err := s.taskRepo.FindByAssignee(ctx, req.UserID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	// 
	filteredTasks := s.filterTasksByDateRange(tasks, req.StartDate, req.EndDate)

	// 
	metrics := s.calculateUserMetrics(filteredTasks)

	// 
	trends := s.analyzeUserTrends(filteredTasks, req.StartDate, req.EndDate)

	// ?
	skillAnalysis := s.analyzeUserSkills(filteredTasks)

	// 
	recommendations := s.generateUserRecommendations(metrics, trends, skillAnalysis)

	return &domain.UserPerformanceReport{
		UserID:          req.UserID,
		Period:          fmt.Sprintf("%s - %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		Metrics:         metrics,
		Trends:          trends,
		SkillAnalysis:   skillAnalysis,
		Recommendations: recommendations,
		GeneratedAt:     time.Now(),
	}, nil
}

// AnalyzeTeamPerformance 
func (s *PerformanceAnalysisService) AnalyzeTeamPerformance(ctx context.Context, req *domain.TeamPerformanceRequest) (*domain.TeamPerformanceReport, error) {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// ?
	allTasks := make([]*domain.Task, 0)
	memberPerformances := make([]*domain.MemberPerformance, 0)

	for _, member := range team.Members {
		tasks, err := s.taskRepo.FindByAssignee(ctx, member.UserID, 1000, 0)
		if err != nil {
			continue // ?
		}

		filteredTasks := s.filterTasksByDateRange(tasks, req.StartDate, req.EndDate)
		allTasks = append(allTasks, filteredTasks...)

		// 
		memberMetrics := s.calculateUserMetrics(filteredTasks)
		memberPerformances = append(memberPerformances, &domain.MemberPerformance{
			UserID:  member.UserID,
			Metrics: memberMetrics,
		})
	}

	// 
	teamMetrics := s.calculateTeamMetrics(allTasks, memberPerformances)

	// 
	collaboration := s.analyzeTeamCollaboration(allTasks, team)

	// 
	trends := s.analyzeTeamTrends(allTasks, req.StartDate, req.EndDate)

	// 
	recommendations := s.generateTeamRecommendations(teamMetrics, collaboration, trends)

	return &domain.TeamPerformanceReport{
		TeamID:             req.TeamID,
		Period:             fmt.Sprintf("%s - %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		TeamMetrics:        teamMetrics,
		MemberPerformances: memberPerformances,
		Collaboration:      collaboration,
		Trends:             trends,
		Recommendations:    recommendations,
		GeneratedAt:        time.Now(),
	}, nil
}

// AnalyzeProjectPerformance 
func (s *PerformanceAnalysisService) AnalyzeProjectPerformance(ctx context.Context, req *domain.ProjectPerformanceRequest) (*domain.ProjectPerformanceReport, error) {
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

	// 
	metrics := s.calculateProjectMetrics(tasks, project)

	// 
	progress := s.analyzeProjectProgress(tasks, project)

	// ?
	resourceUtilization := s.analyzeProjectResourceUtilization(tasks)

	// 
	risks := s.analyzeProjectRisks(tasks, project)

	// 
	prediction := s.predictProjectCompletion(tasks, project)

	// 
	recommendations := s.generateProjectRecommendations(metrics, progress, risks)

	return &domain.ProjectPerformanceReport{
		ProjectID:           req.ProjectID,
		Metrics:             metrics,
		Progress:            progress,
		ResourceUtilization: resourceUtilization,
		Risks:               risks,
		Prediction:          prediction,
		Recommendations:     recommendations,
		GeneratedAt:         time.Now(),
	}, nil
}

// GenerateReport 
func (s *PerformanceAnalysisService) GenerateReport(ctx context.Context, req *domain.ReportGenerationRequest) (*domain.PerformanceReport, error) {
	report := &domain.PerformanceReport{
		Type:        req.Type,
		Period:      fmt.Sprintf("%s - %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
	}

	switch req.Type {
	case domain.ReportTypeUser:
		if req.UserID == nil {
			return nil, fmt.Errorf("user_id is required for user report")
		}
		userReport, err := s.AnalyzeUserPerformance(ctx, &domain.UserPerformanceRequest{
			UserID:    *req.UserID,
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
		})
		if err != nil {
			return nil, err
		}
		report.UserReport = userReport

	case domain.ReportTypeTeam:
		if req.TeamID == nil {
			return nil, fmt.Errorf("team_id is required for team report")
		}
		teamReport, err := s.AnalyzeTeamPerformance(ctx, &domain.TeamPerformanceRequest{
			TeamID:    *req.TeamID,
			StartDate: req.StartDate,
			EndDate:   req.EndDate,
		})
		if err != nil {
			return nil, err
		}
		report.TeamReport = teamReport

	case domain.ReportTypeProject:
		if req.ProjectID == nil {
			return nil, fmt.Errorf("project_id is required for project report")
		}
		projectReport, err := s.AnalyzeProjectPerformance(ctx, &domain.ProjectPerformanceRequest{
			ProjectID: *req.ProjectID,
		})
		if err != nil {
			return nil, err
		}
		report.ProjectReport = projectReport

	case domain.ReportTypeOrganization:
		orgReport, err := s.generateOrganizationReport(ctx, req)
		if err != nil {
			return nil, err
		}
		report.OrganizationReport = orgReport

	default:
		return nil, fmt.Errorf("unsupported report type: %s", req.Type)
	}

	return report, nil
}

// PredictTrends 
func (s *PerformanceAnalysisService) PredictTrends(ctx context.Context, req *domain.TrendPredictionRequest) (*domain.TrendPredictionResult, error) {
	var historicalData []*domain.DataPoint
	var err error

	// 
	switch req.Type {
	case domain.PredictionTypeProductivity:
		historicalData, err = s.getProductivityData(ctx, req)
	case domain.PredictionTypeQuality:
		historicalData, err = s.getQualityData(ctx, req)
	case domain.PredictionTypeDelivery:
		historicalData, err = s.getDeliveryData(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported prediction type: %s", req.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// 㷨
	predictions := s.applyPredictionAlgorithm(historicalData, req.PredictionPeriod)

	// ?
	confidence := s.calculatePredictionConfidence(historicalData, predictions)

	return &domain.TrendPredictionResult{
		Type:         req.Type,
		Predictions:  predictions,
		Confidence:   confidence,
		HistoricalData: historicalData,
		GeneratedAt:  time.Now(),
	}, nil
}

// IdentifyBottlenecks 
func (s *PerformanceAnalysisService) IdentifyBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) (*domain.BottleneckAnalysisResult, error) {
	var bottlenecks []*domain.Bottleneck
	var err error

	switch req.Scope {
	case domain.BottleneckScopeUser:
		bottlenecks, err = s.identifyUserBottlenecks(ctx, req)
	case domain.BottleneckScopeTeam:
		bottlenecks, err = s.identifyTeamBottlenecks(ctx, req)
	case domain.BottleneckScopeProject:
		bottlenecks, err = s.identifyProjectBottlenecks(ctx, req)
	case domain.BottleneckScopeOrganization:
		bottlenecks, err = s.identifyOrganizationBottlenecks(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported bottleneck scope: %s", req.Scope)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to identify bottlenecks: %w", err)
	}

	// ?
	sort.Slice(bottlenecks, func(i, j int) bool {
		return bottlenecks[i].Impact > bottlenecks[j].Impact
	})

	// 
	solutions := s.generateBottleneckSolutions(bottlenecks)

	return &domain.BottleneckAnalysisResult{
		Scope:       req.Scope,
		Bottlenecks: bottlenecks,
		Solutions:   solutions,
		GeneratedAt: time.Now(),
	}, nil
}

// ==========  ==========

// filterTasksByDateRange ?
func (s *PerformanceAnalysisService) filterTasksByDateRange(tasks []*domain.Task, startDate, endDate time.Time) []*domain.Task {
	filtered := make([]*domain.Task, 0)
	for _, task := range tasks {
		if task.CreatedAt.After(startDate) && task.CreatedAt.Before(endDate) {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// calculateUserMetrics 
func (s *PerformanceAnalysisService) calculateUserMetrics(tasks []*domain.Task) *domain.UserMetrics {
	metrics := &domain.UserMetrics{}

	totalTasks := len(tasks)
	completedTasks := 0
	totalEstimatedHours := 0.0
	totalActualHours := 0.0
	onTimeTasks := 0

	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted {
			completedTasks++
			
			// ?
			if task.DueDate != nil && task.CompletedAt != nil && task.CompletedAt.Before(*task.DueDate) {
				onTimeTasks++
			}
		}

		if task.EstimatedHours != nil {
			totalEstimatedHours += *task.EstimatedHours
		}

		if task.ActualHours != nil {
			totalActualHours += *task.ActualHours
		}
	}

	metrics.TotalTasks = totalTasks
	metrics.CompletedTasks = completedTasks
	metrics.CompletionRate = 0
	if totalTasks > 0 {
		metrics.CompletionRate = float64(completedTasks) / float64(totalTasks)
	}

	metrics.OnTimeDeliveryRate = 0
	if completedTasks > 0 {
		metrics.OnTimeDeliveryRate = float64(onTimeTasks) / float64(completedTasks)
	}

	metrics.AverageTaskDuration = 0
	if completedTasks > 0 && totalActualHours > 0 {
		metrics.AverageTaskDuration = totalActualHours / float64(completedTasks)
	}

	metrics.ProductivityScore = s.calculateProductivityScore(metrics)
	metrics.QualityScore = s.calculateQualityScore(tasks)

	return metrics
}

// calculateProductivityScore ?
func (s *PerformanceAnalysisService) calculateProductivityScore(metrics *domain.UserMetrics) float64 {
	// 
	score := (metrics.CompletionRate*0.6 + metrics.OnTimeDeliveryRate*0.4) * 100
	return math.Min(100, score)
}

// calculateQualityScore 
func (s *PerformanceAnalysisService) calculateQualityScore(tasks []*domain.Task) float64 {
	// 
	totalScore := 0.0
	count := 0

	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted {
			score := 80.0 // 
			
			// ?
			switch task.Complexity {
			case domain.TaskComplexityVeryHigh:
				score += 20
			case domain.TaskComplexityHigh:
				score += 15
			case domain.TaskComplexityMedium:
				score += 10
			case domain.TaskComplexityLow:
				score += 5
			}

			totalScore += score
			count++
		}
	}

	if count > 0 {
		return totalScore / float64(count)
	}
	return 0
}

// analyzeUserTrends 
func (s *PerformanceAnalysisService) analyzeUserTrends(tasks []*domain.Task, startDate, endDate time.Time) *domain.UserTrends {
	trends := &domain.UserTrends{}

	// 
	weeklyData := s.groupTasksByWeek(tasks, startDate, endDate)
	
	trends.ProductivityTrend = s.calculateProductivityTrend(weeklyData)
	trends.QualityTrend = s.calculateQualityTrend(weeklyData)
	trends.VelocityTrend = s.calculateVelocityTrend(weeklyData)

	return trends
}

// groupTasksByWeek 
func (s *PerformanceAnalysisService) groupTasksByWeek(tasks []*domain.Task, startDate, endDate time.Time) map[string][]*domain.Task {
	weeklyData := make(map[string][]*domain.Task)

	for _, task := range tasks {
		// ?
		year, week := task.CreatedAt.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)
		
		weeklyData[weekKey] = append(weeklyData[weekKey], task)
	}

	return weeklyData
}

// calculateProductivityTrend ?
func (s *PerformanceAnalysisService) calculateProductivityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		metrics := s.calculateUserMetrics(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  metrics.ProductivityScore,
		})
	}

	// ?
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateQualityTrend 
func (s *PerformanceAnalysisService) calculateQualityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		qualityScore := s.calculateQualityScore(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  qualityScore,
		})
	}

	// ?
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateVelocityTrend 
func (s *PerformanceAnalysisService) calculateVelocityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		completedTasks := 0
		for _, task := range tasks {
			if task.Status == domain.TaskStatusCompleted {
				completedTasks++
			}
		}
		
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  float64(completedTasks),
		})
	}

	// ?
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// analyzeUserSkills ?
func (s *PerformanceAnalysisService) analyzeUserSkills(tasks []*domain.Task) *domain.UserSkillAnalysis {
	analysis := &domain.UserSkillAnalysis{
		SkillPerformance: make(map[string]*domain.SkillMetrics),
	}

	// ?
	typeGroups := make(map[domain.TaskType][]*domain.Task)
	for _, task := range tasks {
		typeGroups[task.Type] = append(typeGroups[task.Type], task)
	}

	// 
	for taskType, typeTasks := range typeGroups {
		metrics := s.calculateUserMetrics(typeTasks)
		skillName := s.getSkillNameByTaskType(taskType)
		
		analysis.SkillPerformance[skillName] = &domain.SkillMetrics{
			CompletionRate:     metrics.CompletionRate,
			QualityScore:       metrics.QualityScore,
			AverageDuration:    metrics.AverageTaskDuration,
			TaskCount:          len(typeTasks),
		}
	}

	// ?
	analysis.Strengths = s.identifyStrengths(analysis.SkillPerformance)
	analysis.Weaknesses = s.identifyWeaknesses(analysis.SkillPerformance)

	return analysis
}

// getSkillNameByTaskType ?
func (s *PerformanceAnalysisService) getSkillNameByTaskType(taskType domain.TaskType) string {
	switch taskType {
	case domain.TaskTypeDevelopment:
		return "?
	case domain.TaskTypeDesign:
		return ""
	case domain.TaskTypeTesting:
		return ""
	case domain.TaskTypeDocumentation:
		return ""
	case domain.TaskTypeResearch:
		return ""
	case domain.TaskTypeMeeting:
		return ""
	case domain.TaskTypeReview:
		return ""
	default:
		return ""
	}
}

// identifyStrengths 
func (s *PerformanceAnalysisService) identifyStrengths(skillPerformance map[string]*domain.SkillMetrics) []string {
	strengths := make([]string, 0)
	
	for skill, metrics := range skillPerformance {
		if metrics.CompletionRate > 0.8 && metrics.QualityScore > 80 {
			strengths = append(strengths, skill)
		}
	}
	
	return strengths
}

// identifyWeaknesses 
func (s *PerformanceAnalysisService) identifyWeaknesses(skillPerformance map[string]*domain.SkillMetrics) []string {
	weaknesses := make([]string, 0)
	
	for skill, metrics := range skillPerformance {
		if metrics.CompletionRate < 0.6 || metrics.QualityScore < 60 {
			weaknesses = append(weaknesses, skill)
		}
	}
	
	return weaknesses
}

// generateUserRecommendations 
func (s *PerformanceAnalysisService) generateUserRecommendations(metrics *domain.UserMetrics, trends *domain.UserTrends, skillAnalysis *domain.UserSkillAnalysis) []*domain.Recommendation {
	recommendations := make([]*domain.Recommendation, 0)

	// 
	if metrics.CompletionRate < 0.7 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "?,
			Description: "",
			Priority:    "?,
			Impact:      "幤",
		})
	}

	// 
	if metrics.OnTimeDeliveryRate < 0.8 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "",
			Description: "?,
			Priority:    "?,
			Impact:      "?,
		})
	}

	// 
	if len(skillAnalysis.Weaknesses) > 0 {
		for _, weakness := range skillAnalysis.Weaknesses {
			recommendations = append(recommendations, &domain.Recommendation{
				Type:        "?,
				Description: fmt.Sprintf("?s?, weakness),
				Priority:    "?,
				Impact:      "",
			})
		}
	}

	return recommendations
}

// calculateTeamMetrics 
func (s *PerformanceAnalysisService) calculateTeamMetrics(allTasks []*domain.Task, memberPerformances []*domain.MemberPerformance) *domain.TeamMetrics {
	metrics := &domain.TeamMetrics{}

	if len(memberPerformances) == 0 {
		return metrics
	}

	// 
	totalCompletionRate := 0.0
	totalProductivityScore := 0.0
	totalQualityScore := 0.0
	totalOnTimeRate := 0.0

	for _, member := range memberPerformances {
		totalCompletionRate += member.Metrics.CompletionRate
		totalProductivityScore += member.Metrics.ProductivityScore
		totalQualityScore += member.Metrics.QualityScore
		totalOnTimeRate += member.Metrics.OnTimeDeliveryRate
	}

	memberCount := float64(len(memberPerformances))
	metrics.AverageCompletionRate = totalCompletionRate / memberCount
	metrics.AverageProductivityScore = totalProductivityScore / memberCount
	metrics.AverageQualityScore = totalQualityScore / memberCount
	metrics.AverageOnTimeRate = totalOnTimeRate / memberCount

	// 
	metrics.CollaborationScore = s.calculateCollaborationScore(allTasks)
	metrics.TeamVelocity = s.calculateTeamVelocity(allTasks)

	return metrics
}

// calculateCollaborationScore 
func (s *PerformanceAnalysisService) calculateCollaborationScore(tasks []*domain.Task) float64 {
	// ?
	collaborativeTasks := 0
	totalTasks := len(tasks)

	for _, task := range tasks {
		if len(task.Dependencies) > 0 || len(task.Comments) > 1 {
			collaborativeTasks++
		}
	}

	if totalTasks > 0 {
		return float64(collaborativeTasks) / float64(totalTasks) * 100
	}
	return 0
}

// calculateTeamVelocity 
func (s *PerformanceAnalysisService) calculateTeamVelocity(tasks []*domain.Task) float64 {
	// ?
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	recentCompletedTasks := 0

	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted && task.CompletedAt != nil && task.CompletedAt.After(oneWeekAgo) {
			recentCompletedTasks++
		}
	}

	return float64(recentCompletedTasks)
}

// analyzeTeamCollaboration 
func (s *PerformanceAnalysisService) analyzeTeamCollaboration(tasks []*domain.Task, team *domain.Team) *domain.TeamCollaboration {
	collaboration := &domain.TeamCollaboration{}

	// 
	collaboration.TaskDistribution = s.analyzeTaskDistribution(tasks, team)

	// ?
	collaboration.CommunicationFrequency = s.analyzeCommunicationFrequency(tasks)

	// 
	collaboration.KnowledgeSharing = s.analyzeKnowledgeSharing(tasks)

	return collaboration
}

// analyzeTaskDistribution 
func (s *PerformanceAnalysisService) analyzeTaskDistribution(tasks []*domain.Task, team *domain.Team) map[string]int {
	distribution := make(map[string]int)

	// ?
	for _, member := range team.Members {
		distribution[member.UserID.String()] = 0
	}

	// 
	for _, task := range tasks {
		if task.AssigneeID != nil {
			distribution[task.AssigneeID.String()]++
		}
	}

	return distribution
}

// analyzeCommunicationFrequency ?
func (s *PerformanceAnalysisService) analyzeCommunicationFrequency(tasks []*domain.Task) float64 {
	totalComments := 0
	totalTasks := len(tasks)

	for _, task := range tasks {
		totalComments += len(task.Comments)
	}

	if totalTasks > 0 {
		return float64(totalComments) / float64(totalTasks)
	}
	return 0
}

// analyzeKnowledgeSharing 
func (s *PerformanceAnalysisService) analyzeKnowledgeSharing(tasks []*domain.Task) float64 {
	// 
	knowledgeTasks := 0
	totalTasks := len(tasks)

	for _, task := range tasks {
		if task.Type == domain.TaskTypeDocumentation || task.Type == domain.TaskTypeReview {
			knowledgeTasks++
		}
	}

	if totalTasks > 0 {
		return float64(knowledgeTasks) / float64(totalTasks) * 100
	}
	return 0
}

// analyzeTeamTrends 
func (s *PerformanceAnalysisService) analyzeTeamTrends(tasks []*domain.Task, startDate, endDate time.Time) *domain.TeamTrends {
	trends := &domain.TeamTrends{}

	// 
	weeklyData := s.groupTasksByWeek(tasks, startDate, endDate)

	trends.VelocityTrend = s.calculateTeamVelocityTrend(weeklyData)
	trends.QualityTrend = s.calculateTeamQualityTrend(weeklyData)
	trends.CollaborationTrend = s.calculateTeamCollaborationTrend(weeklyData)

	return trends
}

// calculateTeamVelocityTrend 
func (s *PerformanceAnalysisService) calculateTeamVelocityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		completedTasks := 0
		for _, task := range tasks {
			if task.Status == domain.TaskStatusCompleted {
				completedTasks++
			}
		}

		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  float64(completedTasks),
		})
	}

	// ?
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateTeamQualityTrend 
func (s *PerformanceAnalysisService) calculateTeamQualityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		qualityScore := s.calculateQualityScore(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  qualityScore,
		})
	}

	// ?
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateTeamCollaborationTrend 
func (s *PerformanceAnalysisService) calculateTeamCollaborationTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		collaborationScore := s.calculateCollaborationScore(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  collaborationScore,
		})
	}

	// ?
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// generateTeamRecommendations 
func (s *PerformanceAnalysisService) generateTeamRecommendations(metrics *domain.TeamMetrics, collaboration *domain.TeamCollaboration, trends *domain.TeamTrends) []*domain.Recommendation {
	recommendations := make([]*domain.Recommendation, 0)

	// 
	if metrics.AverageCompletionRate < 0.7 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "",
			Description: "?,
			Priority:    "?,
			Impact:      "",
		})
	}

	// ?
	if metrics.CollaborationScore < 60 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "",
			Description: "",
			Priority:    "?,
			Impact:      "",
		})
	}

	// ?
	if s.isTaskDistributionUnbalanced(collaboration.TaskDistribution) {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "",
			Description: "䲻",
			Priority:    "?,
			Impact:      "",
		})
	}

	return recommendations
}

// isTaskDistributionUnbalanced 
func (s *PerformanceAnalysisService) isTaskDistributionUnbalanced(distribution map[string]int) bool {
	if len(distribution) < 2 {
		return false
	}

	values := make([]int, 0, len(distribution))
	for _, count := range distribution {
		values = append(values, count)
	}

	// 㷽
	mean := 0.0
	for _, v := range values {
		mean += float64(v)
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += math.Pow(float64(v)-mean, 2)
	}
	variance /= float64(len(values))

	// 50%䲻
	return variance > mean*0.5
}

// calculateProjectMetrics 
func (s *PerformanceAnalysisService) calculateProjectMetrics(tasks []*domain.Task, project *domain.Project) *domain.ProjectMetrics {
	metrics := &domain.ProjectMetrics{}

	totalTasks := len(tasks)
	completedTasks := 0
	inProgressTasks := 0
	overdueTasks := 0
	totalEstimatedHours := 0.0
	totalActualHours := 0.0

	now := time.Now()

	for _, task := range tasks {
		switch task.Status {
		case domain.TaskStatusCompleted:
			completedTasks++
		case domain.TaskStatusInProgress:
			inProgressTasks++
		}

		if task.DueDate != nil && task.DueDate.Before(now) && task.Status != domain.TaskStatusCompleted {
			overdueTasks++
		}

		if task.EstimatedHours != nil {
			totalEstimatedHours += *task.EstimatedHours
		}

		if task.ActualHours != nil {
			totalActualHours += *task.ActualHours
		}
	}

	metrics.TotalTasks = totalTasks
	metrics.CompletedTasks = completedTasks
	metrics.InProgressTasks = inProgressTasks
	metrics.OverdueTasks = overdueTasks

	if totalTasks > 0 {
		metrics.CompletionRate = float64(completedTasks) / float64(totalTasks)
	}

	metrics.EstimatedHours = totalEstimatedHours
	metrics.ActualHours = totalActualHours

	if totalEstimatedHours > 0 {
		metrics.ScheduleVariance = (totalActualHours - totalEstimatedHours) / totalEstimatedHours
	}

	return metrics
}

// analyzeProjectProgress 
func (s *PerformanceAnalysisService) analyzeProjectProgress(tasks []*domain.Task, project *domain.Project) *domain.ProjectProgress {
	progress := &domain.ProjectProgress{}

	// 
	if len(tasks) > 0 {
		completedTasks := 0
		for _, task := range tasks {
			if task.Status == domain.TaskStatusCompleted {
				completedTasks++
			}
		}
		progress.OverallProgress = float64(completedTasks) / float64(len(tasks))
	}

	// ?
	progress.MilestoneProgress = s.calculateMilestoneProgress(project)

	// 
	progress.PredictedCompletion = s.predictProjectCompletionDate(tasks, project)

	// ?
	progress.Status = s.determineProgressStatus(progress, project)

	return progress
}

// calculateMilestoneProgress ?
func (s *PerformanceAnalysisService) calculateMilestoneProgress(project *domain.Project) map[string]float64 {
	milestoneProgress := make(map[string]float64)

	// ?
	for _, milestone := range project.Milestones {
		// 
		// 
		milestoneProgress[milestone.Name] = 0.75 // 75%
	}

	return milestoneProgress
}

// predictProjectCompletionDate 
func (s *PerformanceAnalysisService) predictProjectCompletionDate(tasks []*domain.Task, project *domain.Project) time.Time {
	// ?
	remainingTasks := 0
	for _, task := range tasks {
		if task.Status != domain.TaskStatusCompleted {
			remainingTasks++
		}
	}

	// 1?
	daysToComplete := remainingTasks
	return time.Now().AddDate(0, 0, daysToComplete)
}

// determineProgressStatus ?
func (s *PerformanceAnalysisService) determineProgressStatus(progress *domain.ProjectProgress, project *domain.Project) string {
	if progress.OverallProgress >= 0.9 {
		return ""
	} else if progress.OverallProgress >= 0.7 {
		return ""
	} else if progress.OverallProgress >= 0.5 {
		return ""
	} else if progress.OverallProgress >= 0.3 {
		return ""
	} else {
		return "?
	}
}

// analyzeProjectResourceUtilization ?
func (s *PerformanceAnalysisService) analyzeProjectResourceUtilization(tasks []*domain.Task) *domain.ResourceUtilization {
	utilization := &domain.ResourceUtilization{
		ResourceUsage: make(map[string]float64),
	}

	// 
	resourceHours := make(map[string]float64)
	totalHours := 0.0

	for _, task := range tasks {
		if task.AssigneeID != nil && task.ActualHours != nil {
			resourceID := task.AssigneeID.String()
			resourceHours[resourceID] += *task.ActualHours
			totalHours += *task.ActualHours
		}
	}

	// ?
	for resource, hours := range resourceHours {
		if totalHours > 0 {
			utilization.ResourceUsage[resource] = hours / totalHours
		}
	}

	// ?
	if len(utilization.ResourceUsage) > 0 {
		total := 0.0
		for _, usage := range utilization.ResourceUsage {
			total += usage
		}
		utilization.AverageUtilization = total / float64(len(utilization.ResourceUsage))
	}

	return utilization
}

// analyzeProjectRisks 
func (s *PerformanceAnalysisService) analyzeProjectRisks(tasks []*domain.Task, project *domain.Project) []*domain.ProjectRisk {
	risks := make([]*domain.ProjectRisk, 0)

	// 
	overdueCount := 0
	now := time.Now()
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Before(now) && task.Status != domain.TaskStatusCompleted {
			overdueCount++
		}
	}

	if overdueCount > 0 {
		risks = append(risks, &domain.ProjectRisk{
			Type:        "",
			Description: fmt.Sprintf("?d", overdueCount),
			Probability: 0.8,
			Impact:      "?,
			Mitigation:  "",
		})
	}

	// ?
	if s.hasResourceBottleneck(tasks) {
		risks = append(risks, &domain.ProjectRisk{
			Type:        "",
			Description: "?,
			Probability: 0.6,
			Impact:      "?,
			Mitigation:  "?,
		})
	}

	return risks
}

// hasResourceBottleneck ?
func (s *PerformanceAnalysisService) hasResourceBottleneck(tasks []*domain.Task) bool {
	// ?
	resourceTaskCount := make(map[string]int)
	
	for _, task := range tasks {
		if task.AssigneeID != nil && task.Status != domain.TaskStatusCompleted {
			resourceID := task.AssigneeID.String()
			resourceTaskCount[resourceID]++
		}
	}

	// ?0
	for _, count := range resourceTaskCount {
		if count > 10 {
			return true
		}
	}

	return false
}

// predictProjectCompletion 
func (s *PerformanceAnalysisService) predictProjectCompletion(tasks []*domain.Task, project *domain.Project) *domain.ProjectCompletionPrediction {
	prediction := &domain.ProjectCompletionPrediction{}

	// 
	prediction.PredictedCompletion = s.predictProjectCompletionDate(tasks, project)

	// ?
	prediction.Confidence = s.calculateProjectPredictionConfidence(tasks, project)

	// 
	prediction.CriticalTasks = s.identifyCriticalTasks(tasks)

	return prediction
}

// calculateProjectPredictionConfidence ?
func (s *PerformanceAnalysisService) calculateProjectPredictionConfidence(tasks []*domain.Task, project *domain.Project) float64 {
	// 
	// ?
	completedTasks := 0
	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted {
			completedTasks++
		}
	}

	if len(tasks) > 0 {
		completionRate := float64(completedTasks) / float64(len(tasks))
		return 0.5 + completionRate*0.4 // ?0%
	}

	return 0.5
}

// identifyCriticalTasks 
func (s *PerformanceAnalysisService) identifyCriticalTasks(tasks []*domain.Task) []uuid.UUID {
	criticalTasks := make([]uuid.UUID, 0)

	// ?
	for _, task := range tasks {
		if task.Priority == domain.TaskPriorityUrgent && task.Status != domain.TaskStatusCompleted {
			criticalTasks = append(criticalTasks, task.ID)
		}
	}

	return criticalTasks
}

// generateProjectRecommendations 
func (s *PerformanceAnalysisService) generateProjectRecommendations(metrics *domain.ProjectMetrics, progress *domain.ProjectProgress, risks []*domain.ProjectRisk) []*domain.Recommendation {
	recommendations := make([]*domain.Recommendation, 0)

	// 
	if metrics.CompletionRate < 0.5 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "?,
			Description: "?,
			Priority:    "?,
			Impact:      "",
		})
	}

	// ?
	if metrics.OverdueTasks > 0 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "",
			Description: "",
			Priority:    "?,
			Impact:      "",
		})
	}

	// ?
	for _, risk := range risks {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "",
			Description: risk.Mitigation,
			Priority:    "?,
			Impact:      "",
		})
	}

	return recommendations
}

// generateOrganizationReport 
func (s *PerformanceAnalysisService) generateOrganizationReport(ctx context.Context, req *domain.ReportGenerationRequest) (*domain.OrganizationPerformanceReport, error) {
	// ?
	report := &domain.OrganizationPerformanceReport{
		OrganizationID: req.OrganizationID,
		Period:         fmt.Sprintf("%s - %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		GeneratedAt:    time.Now(),
	}

	// ?
	// 
	report.OverallMetrics = &domain.OrganizationMetrics{
		TotalProjects:     10,
		ActiveProjects:    7,
		CompletedProjects: 3,
		TotalTeams:        5,
		TotalMembers:      25,
		AverageProductivity: 78.5,
	}

	return report, nil
}

// getProductivityData ?
func (s *PerformanceAnalysisService) getProductivityData(ctx context.Context, req *domain.TrendPredictionRequest) ([]*domain.DataPoint, error) {
	// 
	points := make([]*domain.DataPoint, 0)
	
	// 12
	for i := 12; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -7*i)
		value := 70.0 + float64(i)*2 + math.Sin(float64(i))*5 // 
		
		points = append(points, &domain.DataPoint{
			Timestamp: date,
			Value:     value,
		})
	}
	
	return points, nil
}

// getQualityData 
func (s *PerformanceAnalysisService) getQualityData(ctx context.Context, req *domain.TrendPredictionRequest) ([]*domain.DataPoint, error) {
	// ?
	points := make([]*domain.DataPoint, 0)
	
	// 12
	for i := 12; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -7*i)
		value := 80.0 + float64(i)*1.5 + math.Cos(float64(i))*3 // 
		
		points = append(points, &domain.DataPoint{
			Timestamp: date,
			Value:     value,
		})
	}
	
	return points, nil
}

// getDeliveryData 
func (s *PerformanceAnalysisService) getDeliveryData(ctx context.Context, req *domain.TrendPredictionRequest) ([]*domain.DataPoint, error) {
	// ?
	points := make([]*domain.DataPoint, 0)
	
	// 12
	for i := 12; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -7*i)
		value := 85.0 + float64(i)*1 + math.Sin(float64(i)*0.5)*4 // 
		
		points = append(points, &domain.DataPoint{
			Timestamp: date,
			Value:     value,
		})
	}
	
	return points, nil
}

// applyPredictionAlgorithm 㷨
func (s *PerformanceAnalysisService) applyPredictionAlgorithm(historicalData []*domain.DataPoint, predictionPeriod int) []*domain.DataPoint {
	predictions := make([]*domain.DataPoint, 0)
	
	if len(historicalData) < 2 {
		return predictions
	}
	
	// ?
	// 
	n := len(historicalData)
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	
	for i, point := range historicalData {
		x := float64(i)
		y := point.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// ?
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / float64(n)
	
	// 
	lastTimestamp := historicalData[n-1].Timestamp
	for i := 1; i <= predictionPeriod; i++ {
		predictedValue := intercept + slope*float64(n+i-1)
		predictions = append(predictions, &domain.DataPoint{
			Timestamp: lastTimestamp.AddDate(0, 0, 7*i), // 
			Value:     predictedValue,
		})
	}
	
	return predictions
}

// calculatePredictionConfidence ?
func (s *PerformanceAnalysisService) calculatePredictionConfidence(historicalData, predictions []*domain.DataPoint) float64 {
	// 
	if len(historicalData) < 2 {
		return 0.5
	}
	
	// ?
	mean := 0.0
	for _, point := range historicalData {
		mean += point.Value
	}
	mean /= float64(len(historicalData))
	
	variance := 0.0
	for _, point := range historicalData {
		variance += math.Pow(point.Value-mean, 2)
	}
	variance /= float64(len(historicalData))
	
	// 
	confidence := 1.0 / (1.0 + variance/100.0)
	return math.Min(0.95, math.Max(0.1, confidence))
}

// identifyUserBottlenecks 
func (s *PerformanceAnalysisService) identifyUserBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	if req.UserID == nil {
		return bottlenecks, nil
	}
	
	// 
	tasks, err := s.taskRepo.FindByAssignee(ctx, *req.UserID, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	// 
	pendingTasks := 0
	overdueTasks := 0
	now := time.Now()
	
	for _, task := range tasks {
		if task.Status == domain.TaskStatusPending {
			pendingTasks++
		}
		if task.DueDate != nil && task.DueDate.Before(now) && task.Status != domain.TaskStatusCompleted {
			overdueTasks++
		}
	}
	
	// ?
	if pendingTasks > 10 {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "",
			Description: fmt.Sprintf("?d", pendingTasks),
			Impact:      float64(pendingTasks) / 10.0, // 
			Suggestions: []string{"䲿", "?},
		})
	}
	
	// 
	if overdueTasks > 3 {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "",
			Description: fmt.Sprintf("?d", overdueTasks),
			Impact:      float64(overdueTasks) / 3.0,
			Suggestions: []string{"", ""},
		})
	}
	
	return bottlenecks, nil
}

// identifyTeamBottlenecks 
func (s *PerformanceAnalysisService) identifyTeamBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	if req.TeamID == nil {
		return bottlenecks, nil
	}
	
	// 
	team, err := s.teamRepo.FindByID(ctx, *req.TeamID)
	if err != nil {
		return nil, err
	}
	
	// 
	memberWorkloads := make(map[uuid.UUID]int)
	for _, member := range team.Members {
		tasks, err := s.taskRepo.FindByAssignee(ctx, member.UserID, 1000, 0)
		if err != nil {
			continue
		}
		
		activeTasks := 0
		for _, task := range tasks {
			if task.Status == domain.TaskStatusInProgress || task.Status == domain.TaskStatusPending {
				activeTasks++
			}
		}
		memberWorkloads[member.UserID] = activeTasks
	}
	
	// 鹤
	if s.isWorkloadUnbalanced(memberWorkloads) {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "",
			Description: "䲻?,
			Impact:      0.7,
			Suggestions: []string{"", ""},
		})
	}
	
	return bottlenecks, nil
}

// isWorkloadUnbalanced 鹤
func (s *PerformanceAnalysisService) isWorkloadUnbalanced(workloads map[uuid.UUID]int) bool {
	if len(workloads) < 2 {
		return false
	}
	
	values := make([]int, 0, len(workloads))
	for _, workload := range workloads {
		values = append(values, workload)
	}
	
	// 㷽
	mean := 0.0
	for _, v := range values {
		mean += float64(v)
	}
	mean /= float64(len(values))
	
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(float64(v)-mean, 2)
	}
	variance /= float64(len(values))
	
	// ?
	return variance > mean
}

// identifyProjectBottlenecks 
func (s *PerformanceAnalysisService) identifyProjectBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	if req.ProjectID == nil {
		return bottlenecks, nil
	}
	
	// 
	tasks, err := s.taskRepo.FindByProject(ctx, *req.ProjectID, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	// ?
	blockedTasks := s.findBlockedTasks(tasks)
	if len(blockedTasks) > 0 {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "",
			Description: fmt.Sprintf("?d", len(blockedTasks)),
			Impact:      float64(len(blockedTasks)) / float64(len(tasks)),
			Suggestions: []string{"", "?},
		})
	}
	
	// 
	if s.hasResourceBottleneck(tasks) {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "",
			Description: "",
			Impact:      0.8,
			Suggestions: []string{"", ""},
		})
	}
	
	return bottlenecks, nil
}

// findBlockedTasks 
func (s *PerformanceAnalysisService) findBlockedTasks(tasks []*domain.Task) []*domain.Task {
	blockedTasks := make([]*domain.Task, 0)
	taskMap := make(map[uuid.UUID]*domain.Task)
	
	// 
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	
	// 
	for _, task := range tasks {
		if task.Status != domain.TaskStatusCompleted {
			for _, depID := range task.Dependencies {
				if depTask, exists := taskMap[depID]; exists {
					if depTask.Status != domain.TaskStatusCompleted {
						blockedTasks = append(blockedTasks, task)
						break
					}
				}
			}
		}
	}
	
	return blockedTasks
}

// identifyOrganizationBottlenecks 
func (s *PerformanceAnalysisService) identifyOrganizationBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	// ?
	// ?
	
	// 
	bottlenecks = append(bottlenecks, &domain.Bottleneck{
		Type:        "?,
		Description: "?,
		Impact:      0.6,
		Suggestions: []string{"?, ""},
	})
	
	return bottlenecks, nil
}

// generateBottleneckSolutions 
func (s *PerformanceAnalysisService) generateBottleneckSolutions(bottlenecks []*domain.Bottleneck) []*domain.BottleneckSolution {
	solutions := make([]*domain.BottleneckSolution, 0)
	
	for _, bottleneck := range bottlenecks {
		solution := &domain.BottleneckSolution{
			BottleneckType: bottleneck.Type,
			Solutions:      bottleneck.Suggestions,
			Priority:       s.calculateSolutionPriority(bottleneck.Impact),
			EstimatedEffort: s.estimateSolutionEffort(bottleneck.Type),
			ExpectedImpact:  bottleneck.Impact * 0.8, // ?0%?
		}
		solutions = append(solutions, solution)
	}
	
	return solutions
}

// calculateSolutionPriority ?
func (s *PerformanceAnalysisService) calculateSolutionPriority(impact float64) string {
	if impact >= 0.8 {
		return "?
	} else if impact >= 0.6 {
		return "?
	} else if impact >= 0.4 {
		return "?
	} else {
		return "?
	}
}

// estimateSolutionEffort ?
func (s *PerformanceAnalysisService) estimateSolutionEffort(bottleneckType string) string {
	switch bottleneckType {
	case "":
		return ""
	case "":
		return "?
	case "":
		return ""
	case "":
		return "?
	case "":
		return "?
	case "?:
		return ""
	default:
		return ""
	}
}

