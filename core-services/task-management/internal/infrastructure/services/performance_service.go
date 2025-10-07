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

// PerformanceAnalysisService 性能分析服务实现
type PerformanceAnalysisService struct {
	taskRepo    domain.TaskRepository
	projectRepo domain.ProjectRepository
	teamRepo    domain.TeamRepository
}

// NewPerformanceAnalysisService 创建性能分析服务
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

// AnalyzeUserPerformance 分析用户性能
func (s *PerformanceAnalysisService) AnalyzeUserPerformance(ctx context.Context, req *domain.UserPerformanceRequest) (*domain.UserPerformanceReport, error) {
	// 获取用户的任务
	tasks, err := s.taskRepo.FindByAssignee(ctx, req.UserID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	// 过滤时间范围内的任务
	filteredTasks := s.filterTasksByDateRange(tasks, req.StartDate, req.EndDate)

	// 计算基础指标
	metrics := s.calculateUserMetrics(filteredTasks)

	// 分析任务完成趋势
	trends := s.analyzeUserTrends(filteredTasks, req.StartDate, req.EndDate)

	// 技能分析
	skillAnalysis := s.analyzeUserSkills(filteredTasks)

	// 生成改进建议
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

// AnalyzeTeamPerformance 分析团队性能
func (s *PerformanceAnalysisService) AnalyzeTeamPerformance(ctx context.Context, req *domain.TeamPerformanceRequest) (*domain.TeamPerformanceReport, error) {
	// 获取团队信息
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 获取团队成员的任务
	allTasks := make([]*domain.Task, 0)
	memberPerformances := make([]*domain.MemberPerformance, 0)

	for _, member := range team.Members {
		tasks, err := s.taskRepo.FindByAssignee(ctx, member.UserID, 1000, 0)
		if err != nil {
			continue // 跳过错误，继续处理其他成员
		}

		filteredTasks := s.filterTasksByDateRange(tasks, req.StartDate, req.EndDate)
		allTasks = append(allTasks, filteredTasks...)

		// 计算成员性能
		memberMetrics := s.calculateUserMetrics(filteredTasks)
		memberPerformances = append(memberPerformances, &domain.MemberPerformance{
			UserID:  member.UserID,
			Metrics: memberMetrics,
		})
	}

	// 计算团队整体指标
	teamMetrics := s.calculateTeamMetrics(allTasks, memberPerformances)

	// 分析团队协作
	collaboration := s.analyzeTeamCollaboration(allTasks, team)

	// 分析团队趋势
	trends := s.analyzeTeamTrends(allTasks, req.StartDate, req.EndDate)

	// 生成团队改进建议
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

// AnalyzeProjectPerformance 分析项目性能
func (s *PerformanceAnalysisService) AnalyzeProjectPerformance(ctx context.Context, req *domain.ProjectPerformanceRequest) (*domain.ProjectPerformanceReport, error) {
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

	// 计算项目指标
	metrics := s.calculateProjectMetrics(tasks, project)

	// 分析项目进度
	progress := s.analyzeProjectProgress(tasks, project)

	// 分析资源利用率
	resourceUtilization := s.analyzeProjectResourceUtilization(tasks)

	// 分析风险
	risks := s.analyzeProjectRisks(tasks, project)

	// 预测项目完成
	prediction := s.predictProjectCompletion(tasks, project)

	// 生成项目改进建议
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

// GenerateReport 生成性能报告
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

// PredictTrends 预测趋势
func (s *PerformanceAnalysisService) PredictTrends(ctx context.Context, req *domain.TrendPredictionRequest) (*domain.TrendPredictionResult, error) {
	var historicalData []*domain.DataPoint
	var err error

	// 根据预测类型获取历史数据
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

	// 应用预测算法
	predictions := s.applyPredictionAlgorithm(historicalData, req.PredictionPeriod)

	// 计算置信度
	confidence := s.calculatePredictionConfidence(historicalData, predictions)

	return &domain.TrendPredictionResult{
		Type:         req.Type,
		Predictions:  predictions,
		Confidence:   confidence,
		HistoricalData: historicalData,
		GeneratedAt:  time.Now(),
	}, nil
}

// IdentifyBottlenecks 识别瓶颈
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

	// 按影响程度排序
	sort.Slice(bottlenecks, func(i, j int) bool {
		return bottlenecks[i].Impact > bottlenecks[j].Impact
	})

	// 生成解决建议
	solutions := s.generateBottleneckSolutions(bottlenecks)

	return &domain.BottleneckAnalysisResult{
		Scope:       req.Scope,
		Bottlenecks: bottlenecks,
		Solutions:   solutions,
		GeneratedAt: time.Now(),
	}, nil
}

// ========== 私有辅助方法 ==========

// filterTasksByDateRange 按日期范围过滤任务
func (s *PerformanceAnalysisService) filterTasksByDateRange(tasks []*domain.Task, startDate, endDate time.Time) []*domain.Task {
	filtered := make([]*domain.Task, 0)
	for _, task := range tasks {
		if task.CreatedAt.After(startDate) && task.CreatedAt.Before(endDate) {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// calculateUserMetrics 计算用户指标
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
			
			// 检查是否按时完成
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

// calculateProductivityScore 计算生产力分数
func (s *PerformanceAnalysisService) calculateProductivityScore(metrics *domain.UserMetrics) float64 {
	// 基于完成率和按时交付率计算生产力分数
	score := (metrics.CompletionRate*0.6 + metrics.OnTimeDeliveryRate*0.4) * 100
	return math.Min(100, score)
}

// calculateQualityScore 计算质量分数
func (s *PerformanceAnalysisService) calculateQualityScore(tasks []*domain.Task) float64 {
	// 简化处理：基于任务复杂度和完成情况计算质量分数
	totalScore := 0.0
	count := 0

	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted {
			score := 80.0 // 基础分数
			
			// 根据复杂度调整分数
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

// analyzeUserTrends 分析用户趋势
func (s *PerformanceAnalysisService) analyzeUserTrends(tasks []*domain.Task, startDate, endDate time.Time) *domain.UserTrends {
	trends := &domain.UserTrends{}

	// 按周分组统计
	weeklyData := s.groupTasksByWeek(tasks, startDate, endDate)
	
	trends.ProductivityTrend = s.calculateProductivityTrend(weeklyData)
	trends.QualityTrend = s.calculateQualityTrend(weeklyData)
	trends.VelocityTrend = s.calculateVelocityTrend(weeklyData)

	return trends
}

// groupTasksByWeek 按周分组任务
func (s *PerformanceAnalysisService) groupTasksByWeek(tasks []*domain.Task, startDate, endDate time.Time) map[string][]*domain.Task {
	weeklyData := make(map[string][]*domain.Task)

	for _, task := range tasks {
		// 计算任务所在的周
		year, week := task.CreatedAt.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)
		
		weeklyData[weekKey] = append(weeklyData[weekKey], task)
	}

	return weeklyData
}

// calculateProductivityTrend 计算生产力趋势
func (s *PerformanceAnalysisService) calculateProductivityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		metrics := s.calculateUserMetrics(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  metrics.ProductivityScore,
		})
	}

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateQualityTrend 计算质量趋势
func (s *PerformanceAnalysisService) calculateQualityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		qualityScore := s.calculateQualityScore(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  qualityScore,
		})
	}

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateVelocityTrend 计算速度趋势
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

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// analyzeUserSkills 分析用户技能
func (s *PerformanceAnalysisService) analyzeUserSkills(tasks []*domain.Task) *domain.UserSkillAnalysis {
	analysis := &domain.UserSkillAnalysis{
		SkillPerformance: make(map[string]*domain.SkillMetrics),
	}

	// 按任务类型分组分析
	typeGroups := make(map[domain.TaskType][]*domain.Task)
	for _, task := range tasks {
		typeGroups[task.Type] = append(typeGroups[task.Type], task)
	}

	// 计算每种技能的表现
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

	// 识别强项和弱项
	analysis.Strengths = s.identifyStrengths(analysis.SkillPerformance)
	analysis.Weaknesses = s.identifyWeaknesses(analysis.SkillPerformance)

	return analysis
}

// getSkillNameByTaskType 根据任务类型获取技能名称
func (s *PerformanceAnalysisService) getSkillNameByTaskType(taskType domain.TaskType) string {
	switch taskType {
	case domain.TaskTypeDevelopment:
		return "开发"
	case domain.TaskTypeDesign:
		return "设计"
	case domain.TaskTypeTesting:
		return "测试"
	case domain.TaskTypeDocumentation:
		return "文档"
	case domain.TaskTypeResearch:
		return "研究"
	case domain.TaskTypeMeeting:
		return "会议"
	case domain.TaskTypeReview:
		return "评审"
	default:
		return "其他"
	}
}

// identifyStrengths 识别强项
func (s *PerformanceAnalysisService) identifyStrengths(skillPerformance map[string]*domain.SkillMetrics) []string {
	strengths := make([]string, 0)
	
	for skill, metrics := range skillPerformance {
		if metrics.CompletionRate > 0.8 && metrics.QualityScore > 80 {
			strengths = append(strengths, skill)
		}
	}
	
	return strengths
}

// identifyWeaknesses 识别弱项
func (s *PerformanceAnalysisService) identifyWeaknesses(skillPerformance map[string]*domain.SkillMetrics) []string {
	weaknesses := make([]string, 0)
	
	for skill, metrics := range skillPerformance {
		if metrics.CompletionRate < 0.6 || metrics.QualityScore < 60 {
			weaknesses = append(weaknesses, skill)
		}
	}
	
	return weaknesses
}

// generateUserRecommendations 生成用户改进建议
func (s *PerformanceAnalysisService) generateUserRecommendations(metrics *domain.UserMetrics, trends *domain.UserTrends, skillAnalysis *domain.UserSkillAnalysis) []*domain.Recommendation {
	recommendations := make([]*domain.Recommendation, 0)

	// 基于完成率的建议
	if metrics.CompletionRate < 0.7 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "完成率改进",
			Description: "任务完成率较低，建议优化时间管理和任务优先级排序",
			Priority:    "高",
			Impact:      "提高整体工作效率",
		})
	}

	// 基于按时交付率的建议
	if metrics.OnTimeDeliveryRate < 0.8 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "交付时间改进",
			Description: "按时交付率需要提升，建议改进任务估算和进度跟踪",
			Priority:    "中",
			Impact:      "提高项目可预测性",
		})
	}

	// 基于技能分析的建议
	if len(skillAnalysis.Weaknesses) > 0 {
		for _, weakness := range skillAnalysis.Weaknesses {
			recommendations = append(recommendations, &domain.Recommendation{
				Type:        "技能提升",
				Description: fmt.Sprintf("在%s方面需要加强，建议参加相关培训或寻求指导", weakness),
				Priority:    "中",
				Impact:      "提升专业能力",
			})
		}
	}

	return recommendations
}

// calculateTeamMetrics 计算团队指标
func (s *PerformanceAnalysisService) calculateTeamMetrics(allTasks []*domain.Task, memberPerformances []*domain.MemberPerformance) *domain.TeamMetrics {
	metrics := &domain.TeamMetrics{}

	if len(memberPerformances) == 0 {
		return metrics
	}

	// 计算团队平均指标
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

	// 计算团队协作指标
	metrics.CollaborationScore = s.calculateCollaborationScore(allTasks)
	metrics.TeamVelocity = s.calculateTeamVelocity(allTasks)

	return metrics
}

// calculateCollaborationScore 计算协作分数
func (s *PerformanceAnalysisService) calculateCollaborationScore(tasks []*domain.Task) float64 {
	// 简化处理：基于任务依赖关系和评论数量计算协作分数
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

// calculateTeamVelocity 计算团队速度
func (s *PerformanceAnalysisService) calculateTeamVelocity(tasks []*domain.Task) float64 {
	// 计算最近一周完成的任务数
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	recentCompletedTasks := 0

	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted && task.CompletedAt != nil && task.CompletedAt.After(oneWeekAgo) {
			recentCompletedTasks++
		}
	}

	return float64(recentCompletedTasks)
}

// analyzeTeamCollaboration 分析团队协作
func (s *PerformanceAnalysisService) analyzeTeamCollaboration(tasks []*domain.Task, team *domain.Team) *domain.TeamCollaboration {
	collaboration := &domain.TeamCollaboration{}

	// 分析任务分配分布
	collaboration.TaskDistribution = s.analyzeTaskDistribution(tasks, team)

	// 分析沟通频率
	collaboration.CommunicationFrequency = s.analyzeCommunicationFrequency(tasks)

	// 分析知识共享
	collaboration.KnowledgeSharing = s.analyzeKnowledgeSharing(tasks)

	return collaboration
}

// analyzeTaskDistribution 分析任务分配分布
func (s *PerformanceAnalysisService) analyzeTaskDistribution(tasks []*domain.Task, team *domain.Team) map[string]int {
	distribution := make(map[string]int)

	// 初始化成员计数
	for _, member := range team.Members {
		distribution[member.UserID.String()] = 0
	}

	// 统计任务分配
	for _, task := range tasks {
		if task.AssigneeID != nil {
			distribution[task.AssigneeID.String()]++
		}
	}

	return distribution
}

// analyzeCommunicationFrequency 分析沟通频率
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

// analyzeKnowledgeSharing 分析知识共享
func (s *PerformanceAnalysisService) analyzeKnowledgeSharing(tasks []*domain.Task) float64 {
	// 简化处理：基于文档类型任务和评审任务的比例
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

// analyzeTeamTrends 分析团队趋势
func (s *PerformanceAnalysisService) analyzeTeamTrends(tasks []*domain.Task, startDate, endDate time.Time) *domain.TeamTrends {
	trends := &domain.TeamTrends{}

	// 按周分组统计
	weeklyData := s.groupTasksByWeek(tasks, startDate, endDate)

	trends.VelocityTrend = s.calculateTeamVelocityTrend(weeklyData)
	trends.QualityTrend = s.calculateTeamQualityTrend(weeklyData)
	trends.CollaborationTrend = s.calculateTeamCollaborationTrend(weeklyData)

	return trends
}

// calculateTeamVelocityTrend 计算团队速度趋势
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

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateTeamQualityTrend 计算团队质量趋势
func (s *PerformanceAnalysisService) calculateTeamQualityTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		qualityScore := s.calculateQualityScore(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  qualityScore,
		})
	}

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// calculateTeamCollaborationTrend 计算团队协作趋势
func (s *PerformanceAnalysisService) calculateTeamCollaborationTrend(weeklyData map[string][]*domain.Task) []*domain.TrendPoint {
	points := make([]*domain.TrendPoint, 0)

	for week, tasks := range weeklyData {
		collaborationScore := s.calculateCollaborationScore(tasks)
		points = append(points, &domain.TrendPoint{
			Period: week,
			Value:  collaborationScore,
		})
	}

	// 按时间排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Period < points[j].Period
	})

	return points
}

// generateTeamRecommendations 生成团队改进建议
func (s *PerformanceAnalysisService) generateTeamRecommendations(metrics *domain.TeamMetrics, collaboration *domain.TeamCollaboration, trends *domain.TeamTrends) []*domain.Recommendation {
	recommendations := make([]*domain.Recommendation, 0)

	// 基于平均完成率的建议
	if metrics.AverageCompletionRate < 0.7 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "团队效率提升",
			Description: "团队整体完成率较低，建议加强项目管理和资源配置",
			Priority:    "高",
			Impact:      "提高团队整体效率",
		})
	}

	// 基于协作分数的建议
	if metrics.CollaborationScore < 60 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "协作改进",
			Description: "团队协作程度较低，建议增加团队沟通和知识分享活动",
			Priority:    "中",
			Impact:      "提升团队协作效果",
		})
	}

	// 基于任务分配的建议
	if s.isTaskDistributionUnbalanced(collaboration.TaskDistribution) {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "任务分配优化",
			Description: "任务分配不均衡，建议重新评估成员工作负载",
			Priority:    "中",
			Impact:      "平衡团队工作负载",
		})
	}

	return recommendations
}

// isTaskDistributionUnbalanced 检查任务分配是否不均衡
func (s *PerformanceAnalysisService) isTaskDistributionUnbalanced(distribution map[string]int) bool {
	if len(distribution) < 2 {
		return false
	}

	values := make([]int, 0, len(distribution))
	for _, count := range distribution {
		values = append(values, count)
	}

	// 计算方差
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

	// 如果方差大于平均值的50%，认为分配不均衡
	return variance > mean*0.5
}

// calculateProjectMetrics 计算项目指标
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

// analyzeProjectProgress 分析项目进度
func (s *PerformanceAnalysisService) analyzeProjectProgress(tasks []*domain.Task, project *domain.Project) *domain.ProjectProgress {
	progress := &domain.ProjectProgress{}

	// 计算整体进度
	if len(tasks) > 0 {
		completedTasks := 0
		for _, task := range tasks {
			if task.Status == domain.TaskStatusCompleted {
				completedTasks++
			}
		}
		progress.OverallProgress = float64(completedTasks) / float64(len(tasks))
	}

	// 计算里程碑进度
	progress.MilestoneProgress = s.calculateMilestoneProgress(project)

	// 预测完成日期
	progress.PredictedCompletion = s.predictProjectCompletionDate(tasks, project)

	// 计算进度状态
	progress.Status = s.determineProgressStatus(progress, project)

	return progress
}

// calculateMilestoneProgress 计算里程碑进度
func (s *PerformanceAnalysisService) calculateMilestoneProgress(project *domain.Project) map[string]float64 {
	milestoneProgress := make(map[string]float64)

	// 简化处理：假设里程碑信息在项目中
	for _, milestone := range project.Milestones {
		// 这里应该根据里程碑相关的任务计算进度
		// 简化处理，返回模拟数据
		milestoneProgress[milestone.Name] = 0.75 // 75%完成
	}

	return milestoneProgress
}

// predictProjectCompletionDate 预测项目完成日期
func (s *PerformanceAnalysisService) predictProjectCompletionDate(tasks []*domain.Task, project *domain.Project) time.Time {
	// 简化处理：基于当前进度和剩余任务估算
	remainingTasks := 0
	for _, task := range tasks {
		if task.Status != domain.TaskStatusCompleted {
			remainingTasks++
		}
	}

	// 假设平均每天完成1个任务
	daysToComplete := remainingTasks
	return time.Now().AddDate(0, 0, daysToComplete)
}

// determineProgressStatus 确定进度状态
func (s *PerformanceAnalysisService) determineProgressStatus(progress *domain.ProjectProgress, project *domain.Project) string {
	if progress.OverallProgress >= 0.9 {
		return "即将完成"
	} else if progress.OverallProgress >= 0.7 {
		return "进展良好"
	} else if progress.OverallProgress >= 0.5 {
		return "正常进行"
	} else if progress.OverallProgress >= 0.3 {
		return "进度缓慢"
	} else {
		return "需要关注"
	}
}

// analyzeProjectResourceUtilization 分析项目资源利用率
func (s *PerformanceAnalysisService) analyzeProjectResourceUtilization(tasks []*domain.Task) *domain.ResourceUtilization {
	utilization := &domain.ResourceUtilization{
		ResourceUsage: make(map[string]float64),
	}

	// 统计资源使用情况
	resourceHours := make(map[string]float64)
	totalHours := 0.0

	for _, task := range tasks {
		if task.AssigneeID != nil && task.ActualHours != nil {
			resourceID := task.AssigneeID.String()
			resourceHours[resourceID] += *task.ActualHours
			totalHours += *task.ActualHours
		}
	}

	// 计算利用率
	for resource, hours := range resourceHours {
		if totalHours > 0 {
			utilization.ResourceUsage[resource] = hours / totalHours
		}
	}

	// 计算平均利用率
	if len(utilization.ResourceUsage) > 0 {
		total := 0.0
		for _, usage := range utilization.ResourceUsage {
			total += usage
		}
		utilization.AverageUtilization = total / float64(len(utilization.ResourceUsage))
	}

	return utilization
}

// analyzeProjectRisks 分析项目风险
func (s *PerformanceAnalysisService) analyzeProjectRisks(tasks []*domain.Task, project *domain.Project) []*domain.ProjectRisk {
	risks := make([]*domain.ProjectRisk, 0)

	// 检查逾期任务风险
	overdueCount := 0
	now := time.Now()
	for _, task := range tasks {
		if task.DueDate != nil && task.DueDate.Before(now) && task.Status != domain.TaskStatusCompleted {
			overdueCount++
		}
	}

	if overdueCount > 0 {
		risks = append(risks, &domain.ProjectRisk{
			Type:        "进度风险",
			Description: fmt.Sprintf("有%d个任务已逾期", overdueCount),
			Probability: 0.8,
			Impact:      "高",
			Mitigation:  "重新评估任务优先级，增加资源投入",
		})
	}

	// 检查资源风险
	if s.hasResourceBottleneck(tasks) {
		risks = append(risks, &domain.ProjectRisk{
			Type:        "资源风险",
			Description: "存在资源瓶颈，可能影响项目进度",
			Probability: 0.6,
			Impact:      "中",
			Mitigation:  "重新分配资源或增加人员",
		})
	}

	return risks
}

// hasResourceBottleneck 检查是否存在资源瓶颈
func (s *PerformanceAnalysisService) hasResourceBottleneck(tasks []*domain.Task) bool {
	// 简化处理：检查是否有资源分配了过多任务
	resourceTaskCount := make(map[string]int)
	
	for _, task := range tasks {
		if task.AssigneeID != nil && task.Status != domain.TaskStatusCompleted {
			resourceID := task.AssigneeID.String()
			resourceTaskCount[resourceID]++
		}
	}

	// 如果任何资源分配了超过10个活跃任务，认为存在瓶颈
	for _, count := range resourceTaskCount {
		if count > 10 {
			return true
		}
	}

	return false
}

// predictProjectCompletion 预测项目完成
func (s *PerformanceAnalysisService) predictProjectCompletion(tasks []*domain.Task, project *domain.Project) *domain.ProjectCompletionPrediction {
	prediction := &domain.ProjectCompletionPrediction{}

	// 计算预测完成日期
	prediction.PredictedCompletion = s.predictProjectCompletionDate(tasks, project)

	// 计算置信度
	prediction.Confidence = s.calculateProjectPredictionConfidence(tasks, project)

	// 识别关键任务
	prediction.CriticalTasks = s.identifyCriticalTasks(tasks)

	return prediction
}

// calculateProjectPredictionConfidence 计算项目预测置信度
func (s *PerformanceAnalysisService) calculateProjectPredictionConfidence(tasks []*domain.Task, project *domain.Project) float64 {
	// 基于历史数据和当前进度计算置信度
	// 简化处理，返回基于完成率的置信度
	completedTasks := 0
	for _, task := range tasks {
		if task.Status == domain.TaskStatusCompleted {
			completedTasks++
		}
	}

	if len(tasks) > 0 {
		completionRate := float64(completedTasks) / float64(len(tasks))
		return 0.5 + completionRate*0.4 // 基础置信度50%，根据完成率调整
	}

	return 0.5
}

// identifyCriticalTasks 识别关键任务
func (s *PerformanceAnalysisService) identifyCriticalTasks(tasks []*domain.Task) []uuid.UUID {
	criticalTasks := make([]uuid.UUID, 0)

	// 识别高优先级且未完成的任务
	for _, task := range tasks {
		if task.Priority == domain.TaskPriorityUrgent && task.Status != domain.TaskStatusCompleted {
			criticalTasks = append(criticalTasks, task.ID)
		}
	}

	return criticalTasks
}

// generateProjectRecommendations 生成项目改进建议
func (s *PerformanceAnalysisService) generateProjectRecommendations(metrics *domain.ProjectMetrics, progress *domain.ProjectProgress, risks []*domain.ProjectRisk) []*domain.Recommendation {
	recommendations := make([]*domain.Recommendation, 0)

	// 基于完成率的建议
	if metrics.CompletionRate < 0.5 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "进度加速",
			Description: "项目完成率较低，建议增加资源投入或重新评估范围",
			Priority:    "高",
			Impact:      "确保项目按时完成",
		})
	}

	// 基于逾期任务的建议
	if metrics.OverdueTasks > 0 {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "逾期处理",
			Description: "存在逾期任务，需要立即处理并调整后续计划",
			Priority:    "紧急",
			Impact:      "避免项目延期",
		})
	}

	// 基于风险的建议
	for _, risk := range risks {
		recommendations = append(recommendations, &domain.Recommendation{
			Type:        "风险缓解",
			Description: risk.Mitigation,
			Priority:    "高",
			Impact:      "降低项目风险",
		})
	}

	return recommendations
}

// generateOrganizationReport 生成组织报告
func (s *PerformanceAnalysisService) generateOrganizationReport(ctx context.Context, req *domain.ReportGenerationRequest) (*domain.OrganizationPerformanceReport, error) {
	// 简化处理：返回组织级别的汇总报告
	report := &domain.OrganizationPerformanceReport{
		OrganizationID: req.OrganizationID,
		Period:         fmt.Sprintf("%s - %s", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")),
		GeneratedAt:    time.Now(),
	}

	// 这里应该聚合所有团队和项目的数据
	// 简化处理，返回模拟数据
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

// getProductivityData 获取生产力数据
func (s *PerformanceAnalysisService) getProductivityData(ctx context.Context, req *domain.TrendPredictionRequest) ([]*domain.DataPoint, error) {
	// 简化处理：返回模拟的历史生产力数据
	points := make([]*domain.DataPoint, 0)
	
	// 生成过去12周的数据
	for i := 12; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -7*i)
		value := 70.0 + float64(i)*2 + math.Sin(float64(i))*5 // 模拟趋势
		
		points = append(points, &domain.DataPoint{
			Timestamp: date,
			Value:     value,
		})
	}
	
	return points, nil
}

// getQualityData 获取质量数据
func (s *PerformanceAnalysisService) getQualityData(ctx context.Context, req *domain.TrendPredictionRequest) ([]*domain.DataPoint, error) {
	// 简化处理：返回模拟的历史质量数据
	points := make([]*domain.DataPoint, 0)
	
	// 生成过去12周的数据
	for i := 12; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -7*i)
		value := 80.0 + float64(i)*1.5 + math.Cos(float64(i))*3 // 模拟趋势
		
		points = append(points, &domain.DataPoint{
			Timestamp: date,
			Value:     value,
		})
	}
	
	return points, nil
}

// getDeliveryData 获取交付数据
func (s *PerformanceAnalysisService) getDeliveryData(ctx context.Context, req *domain.TrendPredictionRequest) ([]*domain.DataPoint, error) {
	// 简化处理：返回模拟的历史交付数据
	points := make([]*domain.DataPoint, 0)
	
	// 生成过去12周的数据
	for i := 12; i >= 1; i-- {
		date := time.Now().AddDate(0, 0, -7*i)
		value := 85.0 + float64(i)*1 + math.Sin(float64(i)*0.5)*4 // 模拟趋势
		
		points = append(points, &domain.DataPoint{
			Timestamp: date,
			Value:     value,
		})
	}
	
	return points, nil
}

// applyPredictionAlgorithm 应用预测算法
func (s *PerformanceAnalysisService) applyPredictionAlgorithm(historicalData []*domain.DataPoint, predictionPeriod int) []*domain.DataPoint {
	predictions := make([]*domain.DataPoint, 0)
	
	if len(historicalData) < 2 {
		return predictions
	}
	
	// 简单线性回归预测
	// 计算趋势
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
	
	// 计算斜率和截距
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / float64(n)
	
	// 生成预测数据
	lastTimestamp := historicalData[n-1].Timestamp
	for i := 1; i <= predictionPeriod; i++ {
		predictedValue := intercept + slope*float64(n+i-1)
		predictions = append(predictions, &domain.DataPoint{
			Timestamp: lastTimestamp.AddDate(0, 0, 7*i), // 每周一个数据点
			Value:     predictedValue,
		})
	}
	
	return predictions
}

// calculatePredictionConfidence 计算预测置信度
func (s *PerformanceAnalysisService) calculatePredictionConfidence(historicalData, predictions []*domain.DataPoint) float64 {
	// 基于历史数据的方差计算置信度
	if len(historicalData) < 2 {
		return 0.5
	}
	
	// 计算历史数据的方差
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
	
	// 方差越小，置信度越高
	confidence := 1.0 / (1.0 + variance/100.0)
	return math.Min(0.95, math.Max(0.1, confidence))
}

// identifyUserBottlenecks 识别用户瓶颈
func (s *PerformanceAnalysisService) identifyUserBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	if req.UserID == nil {
		return bottlenecks, nil
	}
	
	// 获取用户任务
	tasks, err := s.taskRepo.FindByAssignee(ctx, *req.UserID, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	// 分析任务积压
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
	
	// 如果待处理任务过多，识别为瓶颈
	if pendingTasks > 10 {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "任务积压",
			Description: fmt.Sprintf("用户有%d个待处理任务", pendingTasks),
			Impact:      float64(pendingTasks) / 10.0, // 影响程度
			Suggestions: []string{"重新分配部分任务", "优化任务优先级"},
		})
	}
	
	// 如果逾期任务过多，识别为瓶颈
	if overdueTasks > 3 {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "逾期任务",
			Description: fmt.Sprintf("用户有%d个逾期任务", overdueTasks),
			Impact:      float64(overdueTasks) / 3.0,
			Suggestions: []string{"立即处理逾期任务", "改进时间管理"},
		})
	}
	
	return bottlenecks, nil
}

// identifyTeamBottlenecks 识别团队瓶颈
func (s *PerformanceAnalysisService) identifyTeamBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	if req.TeamID == nil {
		return bottlenecks, nil
	}
	
	// 获取团队信息
	team, err := s.teamRepo.FindByID(ctx, *req.TeamID)
	if err != nil {
		return nil, err
	}
	
	// 分析团队工作负载分布
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
	
	// 检查工作负载不均衡
	if s.isWorkloadUnbalanced(memberWorkloads) {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "工作负载不均",
			Description: "团队成员工作负载分配不均衡",
			Impact:      0.7,
			Suggestions: []string{"重新分配任务", "平衡团队工作负载"},
		})
	}
	
	return bottlenecks, nil
}

// isWorkloadUnbalanced 检查工作负载是否不均衡
func (s *PerformanceAnalysisService) isWorkloadUnbalanced(workloads map[uuid.UUID]int) bool {
	if len(workloads) < 2 {
		return false
	}
	
	values := make([]int, 0, len(workloads))
	for _, workload := range workloads {
		values = append(values, workload)
	}
	
	// 计算方差
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
	
	// 如果方差大于平均值，认为不均衡
	return variance > mean
}

// identifyProjectBottlenecks 识别项目瓶颈
func (s *PerformanceAnalysisService) identifyProjectBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	if req.ProjectID == nil {
		return bottlenecks, nil
	}
	
	// 获取项目任务
	tasks, err := s.taskRepo.FindByProject(ctx, *req.ProjectID, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	// 分析任务依赖关系，找出阻塞任务
	blockedTasks := s.findBlockedTasks(tasks)
	if len(blockedTasks) > 0 {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "任务阻塞",
			Description: fmt.Sprintf("有%d个任务被依赖关系阻塞", len(blockedTasks)),
			Impact:      float64(len(blockedTasks)) / float64(len(tasks)),
			Suggestions: []string{"解决依赖关系", "并行化任务执行"},
		})
	}
	
	// 分析资源瓶颈
	if s.hasResourceBottleneck(tasks) {
		bottlenecks = append(bottlenecks, &domain.Bottleneck{
			Type:        "资源瓶颈",
			Description: "项目存在资源分配瓶颈",
			Impact:      0.8,
			Suggestions: []string{"增加资源投入", "优化资源分配"},
		})
	}
	
	return bottlenecks, nil
}

// findBlockedTasks 找出被阻塞的任务
func (s *PerformanceAnalysisService) findBlockedTasks(tasks []*domain.Task) []*domain.Task {
	blockedTasks := make([]*domain.Task, 0)
	taskMap := make(map[uuid.UUID]*domain.Task)
	
	// 构建任务映射
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	
	// 检查每个任务的依赖
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

// identifyOrganizationBottlenecks 识别组织瓶颈
func (s *PerformanceAnalysisService) identifyOrganizationBottlenecks(ctx context.Context, req *domain.BottleneckAnalysisRequest) ([]*domain.Bottleneck, error) {
	bottlenecks := make([]*domain.Bottleneck, 0)
	
	// 简化处理：分析组织级别的瓶颈
	// 这里应该聚合所有团队和项目的数据进行分析
	
	// 模拟组织级别瓶颈
	bottlenecks = append(bottlenecks, &domain.Bottleneck{
		Type:        "沟通效率",
		Description: "跨团队沟通效率有待提升",
		Impact:      0.6,
		Suggestions: []string{"建立更好的沟通机制", "使用协作工具"},
	})
	
	return bottlenecks, nil
}

// generateBottleneckSolutions 生成瓶颈解决方案
func (s *PerformanceAnalysisService) generateBottleneckSolutions(bottlenecks []*domain.Bottleneck) []*domain.BottleneckSolution {
	solutions := make([]*domain.BottleneckSolution, 0)
	
	for _, bottleneck := range bottlenecks {
		solution := &domain.BottleneckSolution{
			BottleneckType: bottleneck.Type,
			Solutions:      bottleneck.Suggestions,
			Priority:       s.calculateSolutionPriority(bottleneck.Impact),
			EstimatedEffort: s.estimateSolutionEffort(bottleneck.Type),
			ExpectedImpact:  bottleneck.Impact * 0.8, // 假设解决方案能减少80%的影响
		}
		solutions = append(solutions, solution)
	}
	
	return solutions
}

// calculateSolutionPriority 计算解决方案优先级
func (s *PerformanceAnalysisService) calculateSolutionPriority(impact float64) string {
	if impact >= 0.8 {
		return "紧急"
	} else if impact >= 0.6 {
		return "高"
	} else if impact >= 0.4 {
		return "中"
	} else {
		return "低"
	}
}

// estimateSolutionEffort 估算解决方案工作量
func (s *PerformanceAnalysisService) estimateSolutionEffort(bottleneckType string) string {
	switch bottleneckType {
	case "任务积压":
		return "中等"
	case "逾期任务":
		return "高"
	case "工作负载不均":
		return "中等"
	case "任务阻塞":
		return "低"
	case "资源瓶颈":
		return "高"
	case "沟通效率":
		return "中等"
	default:
		return "中等"
	}
}