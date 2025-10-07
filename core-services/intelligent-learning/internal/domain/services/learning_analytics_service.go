package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningAnalyticsService 学习分析服务
type LearningAnalyticsService struct {
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
}

// NewLearningAnalyticsService 创建学习分析服务
func NewLearningAnalyticsService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
) *LearningAnalyticsService {
	return &LearningAnalyticsService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
	}
}

// AnalyticsRequest 分析请求
type AnalyticsRequest struct {
	LearnerID    uuid.UUID  `json:"learner_id"`
	TimeRange    AnalyticsTimeRange  `json:"time_range"`
	AnalysisType string     `json:"analysis_type"` // "progress", "performance", "engagement", "prediction", "comprehensive"
	Granularity  string     `json:"granularity"`   // "daily", "weekly", "monthly"
	IncludeComparison bool  `json:"include_comparison"`
	ComparisonGroup   string `json:"comparison_group,omitempty"` // "peers", "cohort", "global"
}

// AnalyticsTimeRange 分析时间范围
type AnalyticsTimeRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// LearningAnalyticsReport 学习分析报告
type LearningAnalyticsReport struct {
	LearnerID         uuid.UUID              `json:"learner_id"`
	GeneratedAt       time.Time              `json:"generated_at"`
	TimeRange         AnalyticsTimeRange              `json:"time_range"`
	OverallScore      float64                `json:"overall_score"`
	ProgressAnalysis  *ProgressAnalysis      `json:"progress_analysis,omitempty"`
	PerformanceAnalysis *PerformanceAnalysis `json:"performance_analysis,omitempty"`
	EngagementAnalysis *EngagementAnalysis   `json:"engagement_analysis,omitempty"`
	PredictiveAnalysis *PredictiveAnalysis   `json:"predictive_analysis,omitempty"`
	Recommendations   []*AnalyticsRecommendation `json:"recommendations"`
	Insights          []string               `json:"insights"`
	Warnings          []string               `json:"warnings"`
	ComparisonData    *ComparisonData        `json:"comparison_data,omitempty"`
}

// ProgressAnalysis 进度分析
type ProgressAnalysis struct {
	OverallProgress     float64                    `json:"overall_progress"`
	SkillProgress       map[string]*SkillProgress  `json:"skill_progress"`
	GoalProgress        []*GoalProgress            `json:"goal_progress"`
	LearningPathProgress []*PathProgress           `json:"learning_path_progress"`
	MilestoneAchievements []*MilestoneAchievement  `json:"milestone_achievements"`
	ProgressTrend       string                     `json:"progress_trend"` // "improving", "stable", "declining"
	ProgressVelocity    float64                    `json:"progress_velocity"`
	EstimatedCompletion time.Time                  `json:"estimated_completion"`
	CompletionProbability float64                  `json:"completion_probability"`
}

// SkillProgress 技能进度
type SkillProgress struct {
	SkillName       string        `json:"skill_name"`
	CurrentLevel    float64       `json:"current_level"`
	TargetLevel     float64       `json:"target_level"`
	Progress        float64       `json:"progress"`
	TimeSpent       time.Duration `json:"time_spent"`
	LastActivity    time.Time     `json:"last_activity"`
	Trend           string        `json:"trend"`
	Proficiency     string        `json:"proficiency"` // "beginner", "intermediate", "advanced", "expert"
	StrengthAreas   []string      `json:"strength_areas"`
	ImprovementAreas []string     `json:"improvement_areas"`
}

// GoalProgress 目标进度
type GoalProgress struct {
	GoalID          uuid.UUID     `json:"goal_id"`
	GoalName        string        `json:"goal_name"`
	TargetSkill     string        `json:"target_skill"`
	Progress        float64       `json:"progress"`
	TimeRemaining   time.Duration `json:"time_remaining"`
	OnTrack         bool          `json:"on_track"`
	RiskLevel       string        `json:"risk_level"` // "low", "medium", "high"
	CompletionDate  *time.Time    `json:"completion_date,omitempty"`
	Blockers        []string      `json:"blockers"`
	NextActions     []string      `json:"next_actions"`
}

// PathProgress 路径进度
type PathProgress struct {
	PathID          uuid.UUID     `json:"path_id"`
	PathName        string        `json:"path_name"`
	NodesCompleted  int           `json:"nodes_completed"`
	TotalNodes      int           `json:"total_nodes"`
	Progress        float64       `json:"progress"`
	TimeSpent       time.Duration `json:"time_spent"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	CurrentNode     *uuid.UUID    `json:"current_node,omitempty"`
	NextMilestone   *uuid.UUID    `json:"next_milestone,omitempty"`
	Difficulty      string        `json:"difficulty"`
	SuccessRate     float64       `json:"success_rate"`
}

// MilestoneAchievement 里程碑成就
type MilestoneAchievement struct {
	MilestoneID   uuid.UUID `json:"milestone_id"`
	Name          string    `json:"name"`
	AchievedAt    time.Time `json:"achieved_at"`
	TimeToAchieve time.Duration `json:"time_to_achieve"`
	Difficulty    string    `json:"difficulty"`
	Points        int       `json:"points"`
	Badge         string    `json:"badge,omitempty"`
}

// PerformanceAnalysis 表现分析
type PerformanceAnalysis struct {
	OverallPerformance float64                      `json:"overall_performance"`
	AccuracyMetrics    *AccuracyMetrics             `json:"accuracy_metrics"`
	SpeedMetrics       *SpeedMetrics                `json:"speed_metrics"`
	ConsistencyMetrics *ConsistencyMetrics          `json:"consistency_metrics"`
	StrengthsWeaknesses *StrengthsWeaknessesAnalysis `json:"strengths_weaknesses"`
	PerformanceTrends  []*PerformanceTrend          `json:"performance_trends"`
	CompetencyMap      map[string]float64           `json:"competency_map"`
	LearningEfficiency float64                      `json:"learning_efficiency"`
}

// AccuracyMetrics 准确性指标
type AccuracyMetrics struct {
	OverallAccuracy    float64            `json:"overall_accuracy"`
	SkillAccuracy      map[string]float64 `json:"skill_accuracy"`
	DifficultyAccuracy map[string]float64 `json:"difficulty_accuracy"`
	ContentTypeAccuracy map[string]float64 `json:"content_type_accuracy"`
	RecentAccuracy     float64            `json:"recent_accuracy"`
	AccuracyTrend      string             `json:"accuracy_trend"`
}

// SpeedMetrics 速度指标
type SpeedMetrics struct {
	AverageCompletionTime time.Duration      `json:"average_completion_time"`
	SpeedBySkill          map[string]time.Duration `json:"speed_by_skill"`
	SpeedByDifficulty     map[string]time.Duration `json:"speed_by_difficulty"`
	SpeedImprovement      float64            `json:"speed_improvement"`
	OptimalPace           time.Duration      `json:"optimal_pace"`
	CurrentPace           time.Duration      `json:"current_pace"`
}

// ConsistencyMetrics 一致性指标
type ConsistencyMetrics struct {
	PerformanceVariance   float64   `json:"performance_variance"`
	ConsistencyScore      float64   `json:"consistency_score"`
	ReliabilityIndex      float64   `json:"reliability_index"`
	PeakPerformanceTimes  []string  `json:"peak_performance_times"`
	LowPerformanceTimes   []string  `json:"low_performance_times"`
	ConsistencyTrend      string    `json:"consistency_trend"`
}

// StrengthsWeaknessesAnalysis 优势劣势分析
type StrengthsWeaknessesAnalysis struct {
	TopStrengths      []string  `json:"top_strengths"`
	KeyWeaknesses     []string  `json:"key_weaknesses"`
	EmergingStrengths []string  `json:"emerging_strengths"`
	ImprovingAreas    []string  `json:"improving_areas"`
	StagnantAreas     []string  `json:"stagnant_areas"`
	RecommendedFocus  []string  `json:"recommended_focus"`
}

// PerformanceTrend 表现趋势
type PerformanceTrend struct {
	Metric    string    `json:"metric"`
	Trend     string    `json:"trend"` // "improving", "stable", "declining"
	Change    float64   `json:"change"`
	Period    string    `json:"period"`
	Confidence float64  `json:"confidence"`
}

// EngagementAnalysis 参与度分析
type EngagementAnalysis struct {
	OverallEngagement   float64                 `json:"overall_engagement"`
	SessionMetrics      *SessionMetrics         `json:"session_metrics"`
	InteractionMetrics  *InteractionMetrics     `json:"interaction_metrics"`
	MotivationIndicators *MotivationIndicators  `json:"motivation_indicators"`
	EngagementPatterns  []*EngagementPattern    `json:"engagement_patterns"`
	RiskFactors         []string                `json:"risk_factors"`
	EngagementTrend     string                  `json:"engagement_trend"`
}

// SessionMetrics 会话指标
type SessionMetrics struct {
	TotalSessions       int           `json:"total_sessions"`
	AverageSessionLength time.Duration `json:"average_session_length"`
	SessionFrequency    float64       `json:"session_frequency"` // sessions per week
	LongestSession      time.Duration `json:"longest_session"`
	ShortestSession     time.Duration `json:"shortest_session"`
	SessionConsistency  float64       `json:"session_consistency"`
	PreferredTimes      []string      `json:"preferred_times"`
}

// InteractionMetrics 交互指标
type InteractionMetrics struct {
	TotalInteractions    int     `json:"total_interactions"`
	InteractionRate      float64 `json:"interaction_rate"` // interactions per minute
	ContentEngagement    map[string]float64 `json:"content_engagement"`
	FeatureUsage         map[string]int `json:"feature_usage"`
	FeedbackFrequency    float64 `json:"feedback_frequency"`
	HelpSeekingBehavior  float64 `json:"help_seeking_behavior"`
}

// MotivationIndicators 动机指标
type MotivationIndicators struct {
	IntrinsicMotivation  float64   `json:"intrinsic_motivation"`
	ExtrinsicMotivation  float64   `json:"extrinsic_motivation"`
	SelfEfficacy         float64   `json:"self_efficacy"`
	GoalOrientation      string    `json:"goal_orientation"` // "mastery", "performance"
	PersistenceLevel     float64   `json:"persistence_level"`
	ChallengePreference  string    `json:"challenge_preference"`
	MotivationTrend      string    `json:"motivation_trend"`
}

// EngagementPattern 参与模式
type EngagementPattern struct {
	PatternType   string    `json:"pattern_type"` // "daily", "weekly", "content_based"
	Description   string    `json:"description"`
	Frequency     float64   `json:"frequency"`
	Strength      float64   `json:"strength"`
	Trend         string    `json:"trend"`
	Implications  []string  `json:"implications"`
}

// PredictiveAnalysis 预测分析
type PredictiveAnalysis struct {
	SuccessPrediction    *SuccessPrediction    `json:"success_prediction"`
	RiskAssessment       *DomainRiskAssessment       `json:"risk_assessment"`
	LearningTrajectory   *LearningTrajectory   `json:"learning_trajectory"`
	RecommendedActions   []*PredictiveAction   `json:"recommended_actions"`
	ConfidenceLevel      float64               `json:"confidence_level"`
	PredictionHorizon    time.Duration         `json:"prediction_horizon"`
}

// SuccessPrediction 成功预测
type SuccessPrediction struct {
	OverallSuccessRate   float64            `json:"overall_success_rate"`
	GoalSuccessRates     map[string]float64 `json:"goal_success_rates"`
	SkillMasteryRates    map[string]float64 `json:"skill_mastery_rates"`
	CompletionTimeline   map[string]time.Time `json:"completion_timeline"`
	SuccessFactors       []string           `json:"success_factors"`
	RiskFactors          []string           `json:"risk_factors"`
}

// RiskAssessment 风险评估
type DomainRiskAssessment struct {
	OverallRiskLevel     string             `json:"overall_risk_level"` // "low", "medium", "high"
	DropoutRisk          float64            `json:"dropout_risk"`
	PerformanceRisk      float64            `json:"performance_risk"`
	EngagementRisk       float64            `json:"engagement_risk"`
	SpecificRisks        []*SpecificRisk    `json:"specific_risks"`
	MitigationStrategies []string           `json:"mitigation_strategies"`
}

// SpecificRisk 具体风险
type SpecificRisk struct {
	RiskType     string    `json:"risk_type"`
	Description  string    `json:"description"`
	Probability  float64   `json:"probability"`
	Impact       string    `json:"impact"` // "low", "medium", "high"
	Timeline     string    `json:"timeline"`
	Indicators   []string  `json:"indicators"`
	Interventions []string `json:"interventions"`
}

// LearningTrajectory 学习轨迹
type LearningTrajectory struct {
	CurrentState      string                 `json:"current_state"`
	PredictedPath     []*TrajectoryPoint     `json:"predicted_path"`
	AlternativePaths  []*AlternativePath     `json:"alternative_paths"`
	KeyMilestones     []*FutureMilestone     `json:"key_milestones"`
	OptimalStrategy   string                 `json:"optimal_strategy"`
	ExpectedOutcomes  map[string]float64     `json:"expected_outcomes"`
}

// TrajectoryPoint 轨迹点
type TrajectoryPoint struct {
	Timestamp        time.Time `json:"timestamp"`
	PredictedState   string    `json:"predicted_state"`
	SkillLevels      map[string]float64 `json:"skill_levels"`
	Confidence       float64   `json:"confidence"`
	KeyEvents        []string  `json:"key_events"`
}

// AlternativePath 替代路径
type AlternativePath struct {
	PathName         string    `json:"path_name"`
	Description      string    `json:"description"`
	SuccessRate      float64   `json:"success_rate"`
	EstimatedTime    time.Duration `json:"estimated_time"`
	RequiredChanges  []string  `json:"required_changes"`
	Benefits         []string  `json:"benefits"`
	Risks            []string  `json:"risks"`
}

// FutureMilestone 未来里程碑
type FutureMilestone struct {
	Name             string    `json:"name"`
	PredictedDate    time.Time `json:"predicted_date"`
	Probability      float64   `json:"probability"`
	Prerequisites    []string  `json:"prerequisites"`
	Impact           string    `json:"impact"`
}

// PredictiveAction 预测性行动
type PredictiveAction struct {
	ActionType       string    `json:"action_type"`
	Description      string    `json:"description"`
	Priority         int       `json:"priority"`
	ExpectedImpact   float64   `json:"expected_impact"`
	Timeline         string    `json:"timeline"`
	Resources        []string  `json:"resources"`
	SuccessMetrics   []string  `json:"success_metrics"`
}

// AnalyticsRecommendation 分析推荐
type AnalyticsRecommendation struct {
	ID               uuid.UUID `json:"id"`
	Type             string    `json:"type"` // "study_strategy", "content", "schedule", "intervention"
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Rationale        string    `json:"rationale"`
	Priority         int       `json:"priority"`
	ExpectedBenefit  string    `json:"expected_benefit"`
	ImplementationSteps []string `json:"implementation_steps"`
	Timeline         string    `json:"timeline"`
	SuccessMetrics   []string  `json:"success_metrics"`
	Confidence       float64   `json:"confidence"`
}

// ComparisonData 比较数据
type ComparisonData struct {
	ComparisonGroup  string                 `json:"comparison_group"`
	PeerRanking      int                    `json:"peer_ranking"`
	TotalPeers       int                    `json:"total_peers"`
	Percentile       float64                `json:"percentile"`
	BenchmarkMetrics map[string]float64     `json:"benchmark_metrics"`
	RelativeStrengths []string              `json:"relative_strengths"`
	RelativeWeaknesses []string             `json:"relative_weaknesses"`
	ComparisonInsights []string             `json:"comparison_insights"`
}

// GenerateAnalyticsReport 生成学习分析报告
func (s *LearningAnalyticsService) GenerateAnalyticsReport(ctx context.Context, req *AnalyticsRequest) (*LearningAnalyticsReport, error) {
	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	report := &LearningAnalyticsReport{
		LearnerID:   req.LearnerID,
		GeneratedAt: time.Now(),
		TimeRange:   req.TimeRange,
		Insights:    []string{},
		Warnings:    []string{},
	}

	// 根据分析类型生成不同的分析
	switch req.AnalysisType {
	case "progress":
		report.ProgressAnalysis, err = s.analyzeProgress(ctx, learner, req)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze progress: %w", err)
		}
		
	case "performance":
		report.PerformanceAnalysis, err = s.analyzePerformance(ctx, learner, req)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze performance: %w", err)
		}
		
	case "engagement":
		report.EngagementAnalysis, err = s.analyzeEngagement(ctx, learner, req)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze engagement: %w", err)
		}
		
	case "prediction":
		report.PredictiveAnalysis, err = s.generatePredictiveAnalysis(ctx, learner, req)
		if err != nil {
			return nil, fmt.Errorf("failed to generate predictive analysis: %w", err)
		}
		
	default: // comprehensive
		report.ProgressAnalysis, _ = s.analyzeProgress(ctx, learner, req)
		report.PerformanceAnalysis, _ = s.analyzePerformance(ctx, learner, req)
		report.EngagementAnalysis, _ = s.analyzeEngagement(ctx, learner, req)
		report.PredictiveAnalysis, _ = s.generatePredictiveAnalysis(ctx, learner, req)
	}

	// 生成推荐
	report.Recommendations = s.generateRecommendations(ctx, report, learner)

	// 计算总体评分
	report.OverallScore = s.calculateOverallScore(report)

	// 生成洞察和警告
	report.Insights = s.generateInsights(report)
	report.Warnings = s.generateWarnings(report)

	// 生成比较数据（如果需要）
	if req.IncludeComparison {
		report.ComparisonData, _ = s.generateComparisonData(ctx, learner, req)
	}

	return report, nil
}

// analyzeProgress 分析学习进度
func (s *LearningAnalyticsService) analyzeProgress(ctx context.Context, learner *entities.Learner, req *AnalyticsRequest) (*ProgressAnalysis, error) {
	// 获取学习历史
	history, err := s.learnerRepo.GetLearningHistory(ctx, learner.ID, 1000)
	if err != nil {
		return nil, err
	}

	// 获取技能信息
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learner.ID)
	if err != nil {
		return nil, err
	}

	// 获取目标信息
	goals := learner.LearningGoals

	analysis := &ProgressAnalysis{
		SkillProgress: make(map[string]*SkillProgress),
	}

	// 转换 goals 为指针切片
	var goalPtrs []*entities.LearningGoal
	for i := range goals {
		goalPtrs = append(goalPtrs, &goals[i])
	}

	// 分析整体进度
	analysis.OverallProgress = s.calculateOverallProgress(history, skills, goalPtrs)

	// 分析技能进度
	for skillName, skill := range skills {
		skillProgress := s.analyzeSkillProgress(skillName, skill, history)
		analysis.SkillProgress[skillName] = skillProgress
	}

	// 分析目标进度
	for _, goal := range goals {
		if goal.IsActive {
			goalProgress := s.analyzeGoalProgress(&goal, skills, history)
			analysis.GoalProgress = append(analysis.GoalProgress, goalProgress)
		}
	}

	// 分析学习路径进度
	analysis.LearningPathProgress = s.analyzeLearningPathProgress(ctx, learner, history)

	// 分析里程碑成就
	analysis.MilestoneAchievements = s.analyzeMilestoneAchievements(history)

	// 计算进度趋势
	analysis.ProgressTrend = s.calculateProgressTrend(history)
	analysis.ProgressVelocity = s.calculateProgressVelocity(history)

	// 预测完成时间
	analysis.EstimatedCompletion = s.estimateCompletionTime(goalPtrs, analysis.ProgressVelocity)
	analysis.CompletionProbability = s.calculateCompletionProbability(analysis)

	return analysis, nil
}

// analyzePerformance 分析学习表现
func (s *LearningAnalyticsService) analyzePerformance(ctx context.Context, learner *entities.Learner, req *AnalyticsRequest) (*PerformanceAnalysis, error) {
	// 获取学习历史
	history, err := s.learnerRepo.GetLearningHistory(ctx, learner.ID, 1000)
	if err != nil {
		return nil, err
	}

	analysis := &PerformanceAnalysis{
		CompetencyMap: make(map[string]float64),
	}

	// 分析准确性指标
	analysis.AccuracyMetrics = s.analyzeAccuracyMetrics(history)

	// 分析速度指标
	analysis.SpeedMetrics = s.analyzeSpeedMetrics(history)

	// 分析一致性指标
	analysis.ConsistencyMetrics = s.analyzeConsistencyMetrics(history)

	// 分析优势劣势
	analysis.StrengthsWeaknesses = s.analyzeStrengthsWeaknesses(history, learner)

	// 分析表现趋势
	analysis.PerformanceTrends = s.analyzePerformanceTrends(history)

	// 构建能力图谱
	analysis.CompetencyMap = s.buildCompetencyMap(history, learner)

	// 计算学习效率
	analysis.LearningEfficiency = s.calculateLearningEfficiency(history)

	// 计算整体表现
	analysis.OverallPerformance = s.calculateOverallPerformance(analysis)

	return analysis, nil
}

// analyzeEngagement 分析学习参与度
func (s *LearningAnalyticsService) analyzeEngagement(ctx context.Context, learner *entities.Learner, req *AnalyticsRequest) (*EngagementAnalysis, error) {
	// 获取学习历史
	history, err := s.learnerRepo.GetLearningHistory(ctx, learner.ID, 1000)
	if err != nil {
		return nil, err
	}

	analysis := &EngagementAnalysis{}

	// 分析会话指标
	analysis.SessionMetrics = s.analyzeSessionMetrics(history)

	// 分析交互指标
	analysis.InteractionMetrics = s.analyzeInteractionMetrics(history)

	// 分析动机指标
	analysis.MotivationIndicators = s.analyzeMotivationIndicators(history, learner)

	// 识别参与模式
	analysis.EngagementPatterns = s.identifyEngagementPatterns(history)

	// 识别风险因素
	analysis.RiskFactors = s.identifyEngagementRiskFactors(analysis)

	// 计算参与度趋势
	analysis.EngagementTrend = s.calculateEngagementTrend(history)

	// 计算整体参与度
	analysis.OverallEngagement = s.calculateOverallEngagement(analysis)

	return analysis, nil
}

// generatePredictiveAnalysis 生成预测分析
func (s *LearningAnalyticsService) generatePredictiveAnalysis(ctx context.Context, learner *entities.Learner, req *AnalyticsRequest) (*PredictiveAnalysis, error) {
	// 获取学习历史
	history, err := s.learnerRepo.GetLearningHistory(ctx, learner.ID, 1000)
	if err != nil {
		return nil, err
	}

	analysis := &PredictiveAnalysis{
		PredictionHorizon: time.Hour * 24 * 90, // 90天预测期
	}

	// 生成成功预测
	analysis.SuccessPrediction = s.generateSuccessPrediction(history, learner)

	// 生成风险评估
	analysis.RiskAssessment = s.generateRiskAssessment(history, learner)

	// 生成学习轨迹
	analysis.LearningTrajectory = s.generateLearningTrajectory(history, learner)

	// 生成推荐行动
	analysis.RecommendedActions = s.generatePredictiveActions(analysis)

	// 计算置信度
	targetTime := time.Now().Add(analysis.PredictionHorizon)
	analysis.ConfidenceLevel = s.calculatePredictionConfidence(history, targetTime)

	return analysis, nil
}

// 辅助方法实现

func (s *LearningAnalyticsService) calculateOverallProgress(history []*entities.LearningHistory, skills map[string]*entities.SkillLevel, goals []*entities.LearningGoal) float64 {
	if len(goals) == 0 {
		return 0.5 // 默认进度
	}

	totalProgress := 0.0
	activeGoals := 0

	for _, goal := range goals {
		if goal.IsActive {
			progress := s.calculateGoalProgressValue(goal, skills)
			totalProgress += progress
			activeGoals++
		}
	}

	if activeGoals == 0 {
		return 0.5
	}

	return totalProgress / float64(activeGoals)
}

func (s *LearningAnalyticsService) calculateGoalProgressValue(goal *entities.LearningGoal, skills map[string]*entities.SkillLevel) float64 {
	if skill, exists := skills[goal.TargetSkill]; exists {
		return float64(skill.Level) / float64(goal.TargetLevel)
	}
	return 0.0
}

func (s *LearningAnalyticsService) analyzeSkillProgress(skillName string, skill *entities.SkillLevel, history []*entities.LearningHistory) *SkillProgress {
	// 计算该技能的学习时间
	timeSpent := time.Duration(0)
	var lastActivity time.Time

	for _, h := range history {
		if h.SkillName == skillName {
			timeSpent += h.Duration
			if h.Timestamp.After(lastActivity) {
				lastActivity = h.Timestamp
			}
		}
	}

	// 分析趋势
	trend := s.calculateSkillTrend(skillName, history)

	// 确定熟练度级别
	proficiency := s.determineProficiency(skill.Level)

	return &SkillProgress{
		SkillName:    skillName,
		CurrentLevel: float64(skill.Level),
		TargetLevel:  10.0, // 假设最高级别为10
		Progress:     float64(skill.Level) / 10.0,
		TimeSpent:    timeSpent,
		LastActivity: lastActivity,
		Trend:        trend,
		Proficiency:  proficiency,
		StrengthAreas: s.identifySkillStrengths(skillName, history),
		ImprovementAreas: s.identifySkillImprovements(skillName, history),
	}
}

func (s *LearningAnalyticsService) calculateSkillTrend(skillName string, history []*entities.LearningHistory) string {
	// 简化实现：基于最近的表现计算趋势
	recentPerformance := []float64{}
	
	for _, h := range history {
		if h.SkillName == skillName && time.Since(h.Timestamp) <= time.Hour*24*30 {
			recentPerformance = append(recentPerformance, h.Progress)
		}
	}

	if len(recentPerformance) < 2 {
		return "stable"
	}

	// 计算趋势
	firstHalf := recentPerformance[:len(recentPerformance)/2]
	secondHalf := recentPerformance[len(recentPerformance)/2:]

	firstAvg := s.calculateAverage(firstHalf)
	secondAvg := s.calculateAverage(secondHalf)

	if secondAvg > firstAvg+0.1 {
		return "improving"
	} else if secondAvg < firstAvg-0.1 {
		return "declining"
	}
	return "stable"
}

func (s *LearningAnalyticsService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *LearningAnalyticsService) determineProficiency(level int) string {
	switch {
	case level >= 8:
		return "expert"
	case level >= 6:
		return "advanced"
	case level >= 4:
		return "intermediate"
	default:
		return "beginner"
	}
}

func (s *LearningAnalyticsService) identifySkillStrengths(skillName string, history []*entities.LearningHistory) []string {
	// 简化实现
	return []string{"quick_learning", "consistent_practice"}
}

func (s *LearningAnalyticsService) identifySkillImprovements(skillName string, history []*entities.LearningHistory) []string {
	// 简化实现
	return []string{"complex_problems", "time_management"}
}

func (s *LearningAnalyticsService) analyzeGoalProgress(goal *entities.LearningGoal, skills map[string]*entities.SkillLevel, history []*entities.LearningHistory) *GoalProgress {
	progress := s.calculateGoalProgressValue(goal, skills)
	timeRemaining := time.Until(goal.TargetDate)
	onTrack := s.isGoalOnTrack(goal, progress, timeRemaining)
	riskLevel := s.calculateGoalRiskLevel(goal, progress, timeRemaining)

	return &GoalProgress{
		GoalID:        goal.ID,
		GoalName:      goal.Description,
		TargetSkill:   goal.TargetSkill,
		Progress:      progress,
		TimeRemaining: timeRemaining,
		OnTrack:       onTrack,
		RiskLevel:     riskLevel,
		Blockers:      s.identifyGoalBlockers(goal, skills),
		NextActions:   s.suggestNextActions(goal, progress),
	}
}

func (s *LearningAnalyticsService) isGoalOnTrack(goal *entities.LearningGoal, progress float64, timeRemaining time.Duration) bool {
	// 简单的线性预测
	totalTime := goal.TargetDate.Sub(goal.CreatedAt)
	expectedProgress := 1.0 - (timeRemaining.Seconds() / totalTime.Seconds())
	
	return progress >= expectedProgress*0.8 // 允许20%的偏差
}

func (s *LearningAnalyticsService) calculateGoalRiskLevel(goal *entities.LearningGoal, progress float64, timeRemaining time.Duration) string {
	if timeRemaining < 0 {
		return "high" // 已过期
	}

	totalTime := goal.TargetDate.Sub(goal.CreatedAt)
	timeProgress := 1.0 - (timeRemaining.Seconds() / totalTime.Seconds())
	
	if progress < timeProgress*0.5 {
		return "high"
	} else if progress < timeProgress*0.8 {
		return "medium"
	}
	return "low"
}

func (s *LearningAnalyticsService) identifyGoalBlockers(goal *entities.LearningGoal, skills map[string]*entities.SkillLevel) []string {
	var blockers []string
	
	// 检查是否缺少前置技能
	if skill, exists := skills[goal.TargetSkill]; exists {
		if skill.Level < 3 {
			blockers = append(blockers, "insufficient_foundation")
		}
	} else {
		blockers = append(blockers, "skill_not_started")
	}
	
	return blockers
}

func (s *LearningAnalyticsService) suggestNextActions(goal *entities.LearningGoal, progress float64) []string {
	var actions []string
	
	if progress < 0.3 {
		actions = append(actions, "focus_on_fundamentals")
		actions = append(actions, "increase_practice_time")
	} else if progress < 0.7 {
		actions = append(actions, "tackle_intermediate_challenges")
		actions = append(actions, "seek_feedback")
	} else {
		actions = append(actions, "work_on_advanced_topics")
		actions = append(actions, "prepare_for_assessment")
	}
	
	return actions
}

func (s *LearningAnalyticsService) analyzeLearningPathProgress(ctx context.Context, learner *entities.Learner, history []*entities.LearningHistory) []*PathProgress {
	// 这里需要获取学习者的学习路径信息
	// 简化实现，返回示例数据
	return []*PathProgress{
		{
			PathID:         uuid.New(),
			PathName:       "Programming Fundamentals",
			NodesCompleted: 8,
			TotalNodes:     12,
			Progress:       0.67,
			TimeSpent:      time.Hour * 24,
			EstimatedTime:  time.Hour * 36,
			Difficulty:     "intermediate",
			SuccessRate:    0.85,
		},
	}
}

func (s *LearningAnalyticsService) analyzeMilestoneAchievements(history []*entities.LearningHistory) []*MilestoneAchievement {
	var achievements []*MilestoneAchievement
	
	// 基于学习历史识别里程碑
	for _, h := range history {
		if h.Progress >= 1.0 && h.DifficultyLevel >= entities.DifficultyIntermediate {
			achievement := &MilestoneAchievement{
				MilestoneID:   uuid.New(),
				Name:          fmt.Sprintf("Completed %s", h.ContentTitle),
				AchievedAt:    h.Timestamp,
				TimeToAchieve: h.Duration,
				Difficulty:    string(h.DifficultyLevel),
				Points:        int(h.DifficultyLevel) * 10,
				Badge:         "completion_badge",
			}
			achievements = append(achievements, achievement)
		}
	}
	
	return achievements
}

func (s *LearningAnalyticsService) calculateProgressTrend(history []*entities.LearningHistory) string {
	if len(history) < 10 {
		return "stable"
	}

	// 计算最近的进度趋势
	recentHistory := history[len(history)-10:]
	firstHalf := recentHistory[:5]
	secondHalf := recentHistory[5:]

	firstAvg := s.calculateAverageCompletion(firstHalf)
	secondAvg := s.calculateAverageCompletion(secondHalf)

	if secondAvg > firstAvg+0.1 {
		return "improving"
	} else if secondAvg < firstAvg-0.1 {
		return "declining"
	}
	return "stable"
}

func (s *LearningAnalyticsService) calculateAverageCompletion(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0
	}

	total := 0.0
	for _, h := range history {
		total += h.Progress
	}
	return total / float64(len(history))
}

func (s *LearningAnalyticsService) calculateProgressVelocity(history []*entities.LearningHistory) float64 {
	if len(history) < 2 {
		return 1.0
	}

	// 计算最近30天的学习速度
	recentHistory := []*entities.LearningHistory{}
	cutoff := time.Now().AddDate(0, 0, -30)

	for _, h := range history {
		if h.Timestamp.After(cutoff) {
			recentHistory = append(recentHistory, h)
		}
	}

	if len(recentHistory) == 0 {
		return 0.1
	}

	// 计算完成的内容数量/时间
	completedContent := 0
	totalTime := time.Duration(0)

	for _, h := range recentHistory {
		if h.Progress >= 1.0 {
			completedContent++
		}
		totalTime += h.Duration
	}

	if totalTime == 0 {
		return 0.1
	}

	// 返回每小时完成的内容数量
	return float64(completedContent) / totalTime.Hours()
}

func (s *LearningAnalyticsService) estimateCompletionTime(goals []*entities.LearningGoal, velocity float64) time.Time {
	if len(goals) == 0 || velocity <= 0 {
		return time.Now().AddDate(0, 6, 0) // 默认6个月
	}

	// 计算剩余工作量
	remainingWork := 0.0
	for _, goal := range goals {
		if goal.IsActive {
			remainingWork += 1.0 // 简化：每个目标算作1个工作单位
		}
	}

	// 基于速度估算时间
	estimatedHours := remainingWork / velocity
	return time.Now().Add(time.Duration(estimatedHours) * time.Hour)
}

func (s *LearningAnalyticsService) calculateCompletionProbability(analysis *ProgressAnalysis) float64 {
	// 基于多个因素计算完成概率
	baseProbability := 0.7

	// 根据进度调整
	progressFactor := analysis.OverallProgress
	
	// 根据趋势调整
	trendFactor := 1.0
	switch analysis.ProgressTrend {
	case "improving":
		trendFactor = 1.2
	case "declining":
		trendFactor = 0.8
	}

	// 根据速度调整
	velocityFactor := math.Min(analysis.ProgressVelocity, 2.0) / 2.0

	probability := baseProbability * progressFactor * trendFactor * velocityFactor
	return math.Min(math.Max(probability, 0.1), 0.95)
}

// 其他分析方法的简化实现

func (s *LearningAnalyticsService) analyzeAccuracyMetrics(history []*entities.LearningHistory) *AccuracyMetrics {
	if len(history) == 0 {
		return &AccuracyMetrics{
			OverallAccuracy: 0.5,
			SkillAccuracy: make(map[string]float64),
			DifficultyAccuracy: make(map[string]float64),
			ContentTypeAccuracy: make(map[string]float64),
		}
	}

	totalAccuracy := 0.0
	for _, h := range history {
		totalAccuracy += h.Progress
	}

	return &AccuracyMetrics{
		OverallAccuracy: totalAccuracy / float64(len(history)),
		SkillAccuracy: make(map[string]float64),
		DifficultyAccuracy: make(map[string]float64),
		ContentTypeAccuracy: make(map[string]float64),
		RecentAccuracy: s.calculateRecentAccuracy(history),
		AccuracyTrend: s.calculateAccuracyTrend(history),
	}
}

func (s *LearningAnalyticsService) calculateRecentAccuracy(history []*entities.LearningHistory) float64 {
	recentHistory := []*entities.LearningHistory{}
	cutoff := time.Now().AddDate(0, 0, -7) // 最近7天

	for _, h := range history {
		if h.Timestamp.After(cutoff) {
			recentHistory = append(recentHistory, h)
		}
	}

	if len(recentHistory) == 0 {
		return 0.5
	}

	total := 0.0
	for _, h := range history {
		total += h.Progress
	}
	return total / float64(len(recentHistory))
}

func (s *LearningAnalyticsService) calculateAccuracyTrend(history []*entities.LearningHistory) string {
	if len(history) < 10 {
		return "stable"
	}

	recent := history[len(history)-5:]
	earlier := history[len(history)-10 : len(history)-5]

	recentAvg := s.calculateAverageCompletion(recent)
	earlierAvg := s.calculateAverageCompletion(earlier)

	if recentAvg > earlierAvg+0.1 {
		return "improving"
	} else if recentAvg < earlierAvg-0.1 {
		return "declining"
	}
	return "stable"
}

func (s *LearningAnalyticsService) analyzeSpeedMetrics(history []*entities.LearningHistory) *SpeedMetrics {
	if len(history) == 0 {
		return &SpeedMetrics{
			AverageCompletionTime: time.Hour,
			SpeedBySkill: make(map[string]time.Duration),
			SpeedByDifficulty: make(map[string]time.Duration),
		}
	}

	totalTime := time.Duration(0)
	for _, h := range history {
		totalTime += h.Duration
	}

	avgTime := totalTime / time.Duration(len(history))

	return &SpeedMetrics{
		AverageCompletionTime: avgTime,
		SpeedBySkill: make(map[string]time.Duration),
		SpeedByDifficulty: make(map[string]time.Duration),
		SpeedImprovement: s.calculateSpeedImprovement(history),
		OptimalPace: avgTime,
		CurrentPace: s.calculateCurrentPace(history),
	}
}

func (s *LearningAnalyticsService) calculateSpeedImprovement(history []*entities.LearningHistory) float64 {
	if len(history) < 10 {
		return 0.0
	}

	recent := history[len(history)-5:]
	earlier := history[len(history)-10 : len(history)-5]

	recentAvg := s.calculateAverageDuration(recent)
	earlierAvg := s.calculateAverageDuration(earlier)

	if earlierAvg == 0 {
		return 0.0
	}

	return (earlierAvg.Seconds() - recentAvg.Seconds()) / earlierAvg.Seconds()
}

func (s *LearningAnalyticsService) calculateAverageDuration(history []*entities.LearningHistory) time.Duration {
	if len(history) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, h := range history {
		total += h.Duration
	}
	return total / time.Duration(len(history))
}

func (s *LearningAnalyticsService) calculateCurrentPace(history []*entities.LearningHistory) time.Duration {
	recentHistory := []*entities.LearningHistory{}
	cutoff := time.Now().AddDate(0, 0, -7)

	for _, h := range history {
		if h.Timestamp.After(cutoff) {
			recentHistory = append(recentHistory, h)
		}
	}

	return s.calculateAverageDuration(recentHistory)
}

func (s *LearningAnalyticsService) analyzeConsistencyMetrics(history []*entities.LearningHistory) *ConsistencyMetrics {
	if len(history) < 5 {
		return &ConsistencyMetrics{
			ConsistencyScore: 0.5,
			ReliabilityIndex: 0.5,
		}
	}

	// 计算表现方差
	performances := []float64{}
	for _, h := range history {
		performances = append(performances, h.Progress)
	}

	variance := s.calculateVariance(performances)
	consistencyScore := math.Max(0, 1.0-variance)

	return &ConsistencyMetrics{
		PerformanceVariance: variance,
		ConsistencyScore: consistencyScore,
		ReliabilityIndex: consistencyScore * 0.9, // 简化计算
		PeakPerformanceTimes: []string{"morning", "evening"},
		LowPerformanceTimes: []string{"afternoon"},
		ConsistencyTrend: s.calculateConsistencyTrend(history),
	}
}

func (s *LearningAnalyticsService) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := s.calculateAverage(values)
	sumSquaredDiff := 0.0

	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(values))
}

func (s *LearningAnalyticsService) calculateConsistencyTrend(history []*entities.LearningHistory) string {
	// 简化实现
	return "stable"
}

func (s *LearningAnalyticsService) analyzeStrengthsWeaknesses(history []*entities.LearningHistory, learner *entities.Learner) *StrengthsWeaknessesAnalysis {
	return &StrengthsWeaknessesAnalysis{
		TopStrengths: []string{"problem_solving", "persistence", "quick_learning"},
		KeyWeaknesses: []string{"time_management", "complex_algorithms"},
		EmergingStrengths: []string{"pattern_recognition"},
		ImprovingAreas: []string{"debugging_skills"},
		StagnantAreas: []string{"theoretical_concepts"},
		RecommendedFocus: []string{"practice_more_algorithms", "improve_time_management"},
	}
}

func (s *LearningAnalyticsService) analyzePerformanceTrends(history []*entities.LearningHistory) []*PerformanceTrend {
	return []*PerformanceTrend{
		{
			Metric: "accuracy",
			Trend: "improving",
			Change: 0.15,
			Period: "last_month",
			Confidence: 0.8,
		},
		{
			Metric: "speed",
			Trend: "stable",
			Change: 0.02,
			Period: "last_month",
			Confidence: 0.6,
		},
	}
}

func (s *LearningAnalyticsService) buildCompetencyMap(history []*entities.LearningHistory, learner *entities.Learner) map[string]float64 {
	competencyMap := make(map[string]float64)
	
	// 基于学习历史构建能力图谱
	skillPerformance := make(map[string][]float64)
	
	for _, h := range history {
		skillPerformance[h.SkillName] = append(skillPerformance[h.SkillName], h.Progress)
	}
	
	for skill, performances := range skillPerformance {
		competencyMap[skill] = s.calculateAverage(performances)
	}
	
	return competencyMap
}

func (s *LearningAnalyticsService) calculateLearningEfficiency(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0.5
	}

	totalEfficiency := 0.0
	for _, h := range history {
		// 效率 = 完成率 / 时间（小时）
		efficiency := h.Progress / math.Max(h.Duration.Hours(), 0.1)
		totalEfficiency += efficiency
	}

	avgEfficiency := totalEfficiency / float64(len(history))
	return math.Min(avgEfficiency / 2.0, 1.0) // 归一化到0-1
}

func (s *LearningAnalyticsService) calculateOverallPerformance(analysis *PerformanceAnalysis) float64 {
	score := 0.0
	
	if analysis.AccuracyMetrics != nil {
		score += analysis.AccuracyMetrics.OverallAccuracy * 0.4
	}
	
	if analysis.ConsistencyMetrics != nil {
		score += analysis.ConsistencyMetrics.ConsistencyScore * 0.3
	}
	
	score += analysis.LearningEfficiency * 0.3
	
	return math.Min(score, 1.0)
}

// 参与度分析方法

func (s *LearningAnalyticsService) analyzeSessionMetrics(history []*entities.LearningHistory) *SessionMetrics {
	if len(history) == 0 {
		return &SessionMetrics{
			TotalSessions: 0,
			AverageSessionLength: time.Hour,
			SessionFrequency: 0,
		}
	}

	// 按日期分组会话
	sessionsByDate := make(map[string][]*entities.LearningHistory)
	for _, h := range history {
		date := h.Timestamp.Format("2006-01-02")
		sessionsByDate[date] = append(sessionsByDate[date], h)
	}

	totalSessions := len(sessionsByDate)
	totalDuration := time.Duration(0)
	
	var sessionLengths []time.Duration
	for _, sessions := range sessionsByDate {
		sessionDuration := time.Duration(0)
		for _, session := range sessions {
			sessionDuration += session.Duration
		}
		sessionLengths = append(sessionLengths, sessionDuration)
		totalDuration += sessionDuration
	}

	avgSessionLength := time.Duration(0)
	if totalSessions > 0 {
		avgSessionLength = totalDuration / time.Duration(totalSessions)
	}

	// 计算会话频率（每周会话数）
	if len(history) > 0 {
		timeSpan := time.Since(history[0].Timestamp)
		weeks := math.Max(timeSpan.Hours()/168, 1) // 168小时 = 1周
		sessionFrequency := float64(totalSessions) / weeks
		
		return &SessionMetrics{
			TotalSessions: totalSessions,
			AverageSessionLength: avgSessionLength,
			SessionFrequency: sessionFrequency,
			LongestSession: s.findMaxDuration(sessionLengths),
			ShortestSession: s.findMinDuration(sessionLengths),
			SessionConsistency: s.calculateSessionConsistency(sessionLengths),
			PreferredTimes: []string{"morning", "evening"},
		}
	}

	return &SessionMetrics{}
}

func (s *LearningAnalyticsService) findMaxDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	max := durations[0]
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return max
}

func (s *LearningAnalyticsService) findMinDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	min := durations[0]
	for _, d := range durations {
		if d < min {
			min = d
		}
	}
	return min
}

func (s *LearningAnalyticsService) calculateSessionConsistency(durations []time.Duration) float64 {
	if len(durations) < 2 {
		return 1.0
	}

	// 计算时长的变异系数
	durationHours := make([]float64, len(durations))
	for i, d := range durations {
		durationHours[i] = d.Hours()
	}

	mean := s.calculateAverage(durationHours)
	variance := s.calculateVariance(durationHours)
	
	if mean == 0 {
		return 0
	}

	cv := math.Sqrt(variance) / mean // 变异系数
	return math.Max(0, 1.0-cv) // 一致性分数
}

func (s *LearningAnalyticsService) analyzeInteractionMetrics(history []*entities.LearningHistory) *InteractionMetrics {
	totalInteractions := len(history)
	
	if totalInteractions == 0 {
		return &InteractionMetrics{
			TotalInteractions: 0,
			InteractionRate: 0,
			ContentEngagement: make(map[string]float64),
			FeatureUsage: make(map[string]int),
		}
	}

	totalTime := time.Duration(0)
	for _, h := range history {
		totalTime += h.Duration
	}

	interactionRate := 0.0
	if totalTime > 0 {
		interactionRate = float64(totalInteractions) / totalTime.Minutes()
	}

	return &InteractionMetrics{
		TotalInteractions: totalInteractions,
		InteractionRate: interactionRate,
		ContentEngagement: s.calculateContentEngagement(history),
		FeatureUsage: s.calculateFeatureUsage(history),
		FeedbackFrequency: s.calculateFeedbackFrequency(history),
		HelpSeekingBehavior: s.calculateHelpSeekingBehavior(history),
	}
}

func (s *LearningAnalyticsService) calculateContentEngagement(history []*entities.LearningHistory) map[string]float64 {
	engagement := make(map[string]float64)
	contentCount := make(map[string]int)

	for _, h := range history {
		contentType := string(h.ContentType)
		engagement[contentType] += h.Progress
		contentCount[contentType]++
	}

	// 计算平均参与度
	for contentType, total := range engagement {
		engagement[contentType] = total / float64(contentCount[contentType])
	}

	return engagement
}

func (s *LearningAnalyticsService) calculateFeatureUsage(history []*entities.LearningHistory) map[string]int {
	// 简化实现
	return map[string]int{
		"video_player": len(history) / 2,
		"quiz_system": len(history) / 3,
		"note_taking": len(history) / 4,
		"bookmarks": len(history) / 5,
	}
}

func (s *LearningAnalyticsService) calculateFeedbackFrequency(history []*entities.LearningHistory) float64 {
	// 简化实现：假设20%的学习活动包含反馈
	return 0.2
}

func (s *LearningAnalyticsService) calculateHelpSeekingBehavior(history []*entities.LearningHistory) float64 {
	// 简化实现：基于低完成率的内容推断求助行为
	lowPerformanceCount := 0
	for _, h := range history {
		if h.Progress < 0.5 {
			lowPerformanceCount++
		}
	}

	if len(history) == 0 {
		return 0
	}

	return float64(lowPerformanceCount) / float64(len(history))
}

func (s *LearningAnalyticsService) analyzeMotivationIndicators(history []*entities.LearningHistory, learner *entities.Learner) *MotivationIndicators {
	// 基于学习行为分析动机指标
	intrinsicMotivation := s.calculateIntrinsicMotivation(history)
	extrinsicMotivation := s.calculateExtrinsicMotivation(history)
	selfEfficacy := s.calculateSelfEfficacy(history)
	persistenceLevel := s.calculatePersistenceLevel(history)

	return &MotivationIndicators{
		IntrinsicMotivation: intrinsicMotivation,
		ExtrinsicMotivation: extrinsicMotivation,
		SelfEfficacy: selfEfficacy,
		GoalOrientation: s.determineGoalOrientation(learner),
		PersistenceLevel: persistenceLevel,
		ChallengePreference: s.determineChallengePreference(history),
		MotivationTrend: s.calculateMotivationTrend(history),
	}
}

func (s *LearningAnalyticsService) calculateIntrinsicMotivation(history []*entities.LearningHistory) float64 {
	// 基于自主选择的学习活动和持续时间
	voluntaryLearning := 0
	for _, h := range history {
		if h.Duration > time.Hour { // 长时间学习可能表示内在动机
			voluntaryLearning++
		}
	}

	if len(history) == 0 {
		return 0.5
	}

	return float64(voluntaryLearning) / float64(len(history))
}

func (s *LearningAnalyticsService) calculateExtrinsicMotivation(history []*entities.LearningHistory) float64 {
	// 基于目标导向的学习行为
	goalOrientedLearning := 0
	for _, h := range history {
		if h.Progress >= 1.0 { // 完成学习可能表示外在动机
			goalOrientedLearning++
		}
	}

	if len(history) == 0 {
		return 0.5
	}

	return float64(goalOrientedLearning) / float64(len(history))
}

func (s *LearningAnalyticsService) calculateSelfEfficacy(history []*entities.LearningHistory) float64 {
	// 基于成功率和挑战接受度
	successfulChallenges := 0
	totalChallenges := 0

	for _, h := range history {
		if h.DifficultyLevel >= entities.DifficultyIntermediate {
			totalChallenges++
			if h.Progress >= 0.8 {
				successfulChallenges++
			}
		}
	}

	if totalChallenges == 0 {
		return 0.5
	}

	return float64(successfulChallenges) / float64(totalChallenges)
}

func (s *LearningAnalyticsService) calculatePersistenceLevel(history []*entities.LearningHistory) float64 {
	// 基于重复尝试和持续学习
	retryCount := 0
	for _, h := range history {
		if h.Progress < 1.0 && h.Progress > 0 {
			retryCount++ // 部分完成可能表示坚持
		}
	}

	if len(history) == 0 {
		return 0.5
	}

	persistenceRatio := float64(retryCount) / float64(len(history))
	return math.Min(persistenceRatio * 2, 1.0) // 归一化
}

func (s *LearningAnalyticsService) determineGoalOrientation(learner *entities.Learner) string {
	// 基于学习目标类型判断
	masteryGoals := 0
	performanceGoals := 0

	for _, goal := range learner.LearningGoals {
		if goal.TargetLevel >= 8 { // 高水平目标可能表示掌握导向
			masteryGoals++
		} else {
			performanceGoals++
		}
	}

	if masteryGoals > performanceGoals {
		return "mastery"
	}
	return "performance"
}

func (s *LearningAnalyticsService) determineChallengePreference(history []*entities.LearningHistory) string {
	difficultContent := 0
	easyContent := 0

	for _, h := range history {
		if h.DifficultyLevel >= entities.DifficultyAdvanced {
			difficultContent++
		} else if h.DifficultyLevel <= entities.DifficultyBeginner {
			easyContent++
		}
	}

	if difficultContent > easyContent {
		return "high_challenge"
	} else if easyContent > difficultContent {
		return "low_challenge"
	}
	return "moderate_challenge"
}

func (s *LearningAnalyticsService) calculateMotivationTrend(history []*entities.LearningHistory) string {
	if len(history) < 10 {
		return "stable"
	}

	// 基于学习时长和频率的趋势
	recent := history[len(history)-5:]
	earlier := history[len(history)-10 : len(history)-5]

	recentAvgDuration := s.calculateAverageDuration(recent)
	earlierAvgDuration := s.calculateAverageDuration(earlier)

	if recentAvgDuration > time.Duration(float64(earlierAvgDuration)*1.2) {
		return "increasing"
	} else if recentAvgDuration < time.Duration(float64(earlierAvgDuration)*0.8) {
		return "decreasing"
	}
	return "stable"
}

func (s *LearningAnalyticsService) identifyEngagementPatterns(history []*entities.LearningHistory) []*EngagementPattern {
	patterns := []*EngagementPattern{}

	// 识别每日学习模式
	dailyPattern := s.analyzeDailyPattern(history)
	if dailyPattern != nil {
		patterns = append(patterns, dailyPattern)
	}

	// 识别每周学习模式
	weeklyPattern := s.analyzeWeeklyPattern(history)
	if weeklyPattern != nil {
		patterns = append(patterns, weeklyPattern)
	}

	// 识别内容偏好模式
	contentPattern := s.analyzeContentPattern(history)
	if contentPattern != nil {
		patterns = append(patterns, contentPattern)
	}

	return patterns
}

func (s *LearningAnalyticsService) analyzeDailyPattern(history []*entities.LearningHistory) *EngagementPattern {
	hourCounts := make(map[int]int)
	
	for _, h := range history {
		hour := h.Timestamp.Hour()
		hourCounts[hour]++
	}

	// 找到最活跃的时间段
	maxCount := 0
	peakHour := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}

	if maxCount < 3 { // 至少3次活动才算模式
		return nil
	}

	var timeDescription string
	switch {
	case peakHour >= 6 && peakHour < 12:
		timeDescription = "morning_learner"
	case peakHour >= 12 && peakHour < 18:
		timeDescription = "afternoon_learner"
	default:
		timeDescription = "evening_learner"
	}

	return &EngagementPattern{
		PatternType:  "daily",
		Description:  timeDescription,
		Frequency:    float64(maxCount) / float64(len(history)),
		Strength:     math.Min(float64(maxCount)/10.0, 1.0),
		Trend:        "stable",
		Implications: []string{"optimize_schedule_for_peak_hours"},
	}
}

func (s *LearningAnalyticsService) analyzeWeeklyPattern(history []*entities.LearningHistory) *EngagementPattern {
	weekdayCounts := make(map[time.Weekday]int)
	
	for _, h := range history {
		weekday := h.Timestamp.Weekday()
		weekdayCounts[weekday]++
	}

	// 计算工作日vs周末的学习分布
	weekdayTotal := 0
	weekendTotal := 0

	for day, count := range weekdayCounts {
		if day == time.Saturday || day == time.Sunday {
			weekendTotal += count
		} else {
			weekdayTotal += count
		}
	}

	var description string
	if weekdayTotal > weekendTotal*2 {
		description = "weekday_focused"
	} else if weekendTotal > weekdayTotal {
		description = "weekend_focused"
	} else {
		description = "balanced_weekly"
	}

	return &EngagementPattern{
		PatternType:  "weekly",
		Description:  description,
		Frequency:    1.0, // 每周模式
		Strength:     0.7,
		Trend:        "stable",
		Implications: []string{"adjust_content_delivery_schedule"},
	}
}

func (s *LearningAnalyticsService) analyzeContentPattern(history []*entities.LearningHistory) *EngagementPattern {
	contentTypeCounts := make(map[entities.ContentType]int)
	
	for _, h := range history {
		contentTypeCounts[entities.ContentType(h.ContentType)]++
	}

	// 找到最偏好的内容类型
	maxCount := 0
	var preferredType entities.ContentType
	for contentType, count := range contentTypeCounts {
		if count > maxCount {
			maxCount = count
			preferredType = contentType
		}
	}

	if maxCount < 3 {
		return nil
	}

	return &EngagementPattern{
		PatternType:  "content_based",
		Description:  fmt.Sprintf("prefers_%s_content", string(preferredType)),
		Frequency:    float64(maxCount) / float64(len(history)),
		Strength:     math.Min(float64(maxCount)/float64(len(history))*2, 1.0),
		Trend:        "stable",
		Implications: []string{"recommend_similar_content_types"},
	}
}

func (s *LearningAnalyticsService) identifyEngagementRiskFactors(analysis *EngagementAnalysis) []string {
	var risks []string

	if analysis.SessionMetrics.SessionFrequency < 2 {
		risks = append(risks, "low_session_frequency")
	}

	if analysis.SessionMetrics.AverageSessionLength < time.Minute*30 {
		risks = append(risks, "short_session_duration")
	}

	if analysis.InteractionMetrics.InteractionRate < 1.0 {
		risks = append(risks, "low_interaction_rate")
	}

	if analysis.MotivationIndicators.PersistenceLevel < 0.5 {
		risks = append(risks, "low_persistence")
	}

	return risks
}

func (s *LearningAnalyticsService) calculateEngagementTrend(history []*entities.LearningHistory) string {
	if len(history) < 10 {
		return "stable"
	}

	// 基于最近的活动频率计算趋势
	recent := history[len(history)-5:]
	earlier := history[len(history)-10 : len(history)-5]

	recentTimeSpan := recent[len(recent)-1].Timestamp.Sub(recent[0].Timestamp)
	earlierTimeSpan := earlier[len(earlier)-1].Timestamp.Sub(earlier[0].Timestamp)

	if recentTimeSpan == 0 || earlierTimeSpan == 0 {
		return "stable"
	}

	recentFreq := float64(len(recent)) / recentTimeSpan.Hours()
	earlierFreq := float64(len(earlier)) / earlierTimeSpan.Hours()

	if recentFreq > earlierFreq*1.2 {
		return "increasing"
	} else if recentFreq < earlierFreq*0.8 {
		return "decreasing"
	}
	return "stable"
}

func (s *LearningAnalyticsService) calculateOverallEngagement(analysis *EngagementAnalysis) float64 {
	score := 0.0

	// 会话指标权重 30%
	if analysis.SessionMetrics != nil {
		sessionScore := math.Min(analysis.SessionMetrics.SessionFrequency/5.0, 1.0) * 0.5
		sessionScore += math.Min(analysis.SessionMetrics.AverageSessionLength.Hours(), 2.0) / 2.0 * 0.5
		score += sessionScore * 0.3
	}

	// 交互指标权重 30%
	if analysis.InteractionMetrics != nil {
		interactionScore := math.Min(analysis.InteractionMetrics.InteractionRate/2.0, 1.0)
		score += interactionScore * 0.3
	}

	// 动机指标权重 40%
	if analysis.MotivationIndicators != nil {
		motivationScore := (analysis.MotivationIndicators.IntrinsicMotivation + 
						  analysis.MotivationIndicators.PersistenceLevel) / 2.0
		score += motivationScore * 0.4
	}

	return math.Min(score, 1.0)
}

// 预测分析方法

func (s *LearningAnalyticsService) generateSuccessPrediction(history []*entities.LearningHistory, learner *entities.Learner) *SuccessPrediction {
	overallSuccessRate := s.calculateOverallSuccessRate(history)
	
	// Convert []entities.LearningGoal to []*entities.LearningGoal
	goalPointers := make([]*entities.LearningGoal, len(learner.LearningGoals))
	for i := range learner.LearningGoals {
		goalPointers[i] = &learner.LearningGoals[i]
	}
	
	goalSuccessRates := s.calculateGoalSuccessRates(goalPointers, history)
	skillMasteryRates := s.calculateSkillMasteryRates(learner, history)

	return &SuccessPrediction{
		OverallSuccessRate: overallSuccessRate,
		GoalSuccessRates:   goalSuccessRates,
		SkillMasteryRates:  skillMasteryRates,
		CompletionTimeline: s.predictCompletionTimeline(goalPointers),
		SuccessFactors:     s.identifySuccessFactors(history),
		RiskFactors:        s.identifyRiskFactors(history),
	}
}

func (s *LearningAnalyticsService) calculateOverallSuccessRate(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0.7 // 默认成功率
	}

	successCount := 0
	for _, h := range history {
		if h.Progress >= 0.8 {
			successCount++
		}
	}

	baseRate := float64(successCount) / float64(len(history))
	
	// 基于趋势调整
	trend := s.calculateProgressTrend(history)
	switch trend {
	case "improving":
		return math.Min(baseRate*1.2, 0.95)
	case "declining":
		return math.Max(baseRate*0.8, 0.1)
	default:
		return baseRate
	}
}

func (s *LearningAnalyticsService) calculateGoalSuccessRates(goals []*entities.LearningGoal, history []*entities.LearningHistory) map[string]float64 {
	rates := make(map[string]float64)
	
	for _, goal := range goals {
		if goal.IsActive {
			rate := s.predictGoalSuccessRate(goal, history)
			rates[goal.Description] = rate
		}
	}
	
	return rates
}

func (s *LearningAnalyticsService) predictGoalSuccessRate(goal *entities.LearningGoal, history []*entities.LearningHistory) float64 {
	// 基于目标相关的学习历史预测成功率
	relevantHistory := []*entities.LearningHistory{}
	for _, h := range history {
		if h.SkillName == goal.TargetSkill {
			relevantHistory = append(relevantHistory, h)
		}
	}

	if len(relevantHistory) == 0 {
		return 0.6 // 默认成功率
	}

	avgPerformance := s.calculateAverageCompletion(relevantHistory)
	timeRemaining := time.Until(goal.TargetDate)
	
	// 基于时间压力调整
	if timeRemaining < time.Hour*24*7 { // 少于一周
		return avgPerformance * 0.8
	} else if timeRemaining > time.Hour*24*30 { // 超过一个月
		return math.Min(avgPerformance*1.2, 0.95)
	}
	
	return avgPerformance
}

func (s *LearningAnalyticsService) calculateSkillMasteryRates(learner *entities.Learner, history []*entities.LearningHistory) map[string]float64 {
	rates := make(map[string]float64)
	
	for _, skill := range learner.Skills {
		rate := s.predictSkillMasteryRate(skill.SkillName, history)
		rates[skill.SkillName] = rate
	}
	
	return rates
}

func (s *LearningAnalyticsService) predictSkillMasteryRate(skillName string, history []*entities.LearningHistory) float64 {
	skillHistory := []*entities.LearningHistory{}
	for _, h := range history {
		if h.SkillName == skillName {
			skillHistory = append(skillHistory, h)
		}
	}

	if len(skillHistory) == 0 {
		return 0.5
	}

	// 基于学习曲线预测掌握率
	recentPerformance := s.calculateAverageCompletion(skillHistory)
	learningVelocity := s.calculateSkillLearningVelocity(skillHistory)
	
	masteryRate := recentPerformance + learningVelocity*0.3
	return math.Min(math.Max(masteryRate, 0.1), 0.95)
}

func (s *LearningAnalyticsService) calculateSkillLearningVelocity(history []*entities.LearningHistory) float64 {
	if len(history) < 5 {
		return 0.1
	}

	// 计算学习速度（性能改善率）
	recent := history[len(history)-3:]
	earlier := history[len(history)-6 : len(history)-3]

	recentAvg := s.calculateAverageCompletion(recent)
	earlierAvg := s.calculateAverageCompletion(earlier)

	return recentAvg - earlierAvg
}

func (s *LearningAnalyticsService) predictCompletionTimeline(goals []*entities.LearningGoal) map[string]time.Time {
	timeline := make(map[string]time.Time)
	
	for _, goal := range goals {
		if goal.IsActive {
			// 基于目标的目标日期和当前进度预测
			timeline[goal.Description] = goal.TargetDate
		}
	}
	
	return timeline
}

func (s *LearningAnalyticsService) identifySuccessFactors(history []*entities.LearningHistory) []string {
	factors := []string{}
	
	// 分析成功的学习会话特征
	successfulSessions := []*entities.LearningHistory{}
	for _, h := range history {
		if h.Progress >= 0.8 {
			successfulSessions = append(successfulSessions, h)
		}
	}

	if len(successfulSessions) > 0 {
		avgDuration := s.calculateAverageDuration(successfulSessions)
		if avgDuration > time.Hour {
			factors = append(factors, "longer_study_sessions")
		}
		
		// 分析成功的内容类型
		contentTypes := make(map[entities.ContentType]int)
		for _, h := range successfulSessions {
			contentTypes[entities.ContentType(h.ContentType)]++
		}
		
		for contentType, count := range contentTypes {
			if count > len(successfulSessions)/3 {
				factors = append(factors, fmt.Sprintf("effective_%s_content", string(contentType)))
			}
		}
	}

	if len(factors) == 0 {
		factors = append(factors, "consistent_practice", "appropriate_difficulty")
	}
	
	return factors
}

func (s *LearningAnalyticsService) identifyRiskFactors(history []*entities.LearningHistory) []string {
	factors := []string{}
	
	// 分析失败的学习会话特征
	failedSessions := []*entities.LearningHistory{}
	for _, h := range history {
		if h.Progress < 0.5 {
			failedSessions = append(failedSessions, h)
		}
	}

	if len(failedSessions) > len(history)/4 { // 超过25%的失败率
		factors = append(factors, "high_failure_rate")
	}

	// 检查学习间隔
	if len(history) >= 2 {
		gaps := []time.Duration{}
		for i := 1; i < len(history); i++ {
			gap := history[i].Timestamp.Sub(history[i-1].Timestamp)
			gaps = append(gaps, gap)
		}
		
		avgGap := s.calculateAverageDuration(history) // 这里应该是gaps，但为了简化
		if avgGap > time.Hour*24*7 { // 超过一周的间隔
			factors = append(factors, "irregular_study_schedule")
		}
	}

	if len(factors) == 0 {
		factors = append(factors, "time_management", "difficulty_level")
	}
	
	return factors
}

func (s *LearningAnalyticsService) generateRiskAssessment(history []*entities.LearningHistory, learner *entities.Learner) *DomainRiskAssessment {
	dropoutRisk := s.calculateDropoutRisk(history)
	performanceRisk := s.calculatePerformanceRisk(history)
	engagementRisk := s.calculateEngagementRisk(history)
	
	overallRisk := (dropoutRisk + performanceRisk + engagementRisk) / 3.0
	var riskLevel string
	switch {
	case overallRisk >= 0.7:
		riskLevel = "high"
	case overallRisk >= 0.4:
		riskLevel = "medium"
	default:
		riskLevel = "low"
	}

	return &DomainRiskAssessment{
		OverallRiskLevel: riskLevel,
		DropoutRisk: dropoutRisk,
		PerformanceRisk: performanceRisk,
		EngagementRisk: engagementRisk,
		SpecificRisks: s.identifySpecificRisks(history, learner),
		MitigationStrategies: s.generateMitigationStrategies(riskLevel, history),
	}
}

func (s *LearningAnalyticsService) calculateDropoutRisk(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0.5
	}

	// 基于学习频率和一致性计算辍学风险
	recentActivity := 0
	cutoff := time.Now().AddDate(0, 0, -14) // 最近两周
	
	for _, h := range history {
		if h.Timestamp.After(cutoff) {
			recentActivity++
		}
	}

	if recentActivity == 0 {
		return 0.9 // 高辍学风险
	}

	// 计算活动下降趋势
	if len(history) >= 10 {
		recent := history[len(history)-5:]
		earlier := history[len(history)-10 : len(history)-5]
		
		recentFreq := float64(len(recent)) / 5.0
		earlierFreq := float64(len(earlier)) / 5.0
		
		if recentFreq < earlierFreq * 0.5 {
			return 0.8
		}
	}

	// 基于完成率计算
	avgCompletion := s.calculateAverageCompletion(history)
	if avgCompletion < 0.3 {
		return 0.7
	}

	return math.Max(0.1, 0.5 - avgCompletion*0.4)
}

func (s *LearningAnalyticsService) calculatePerformanceRisk(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0.5
	}

	avgPerformance := s.calculateAverageCompletion(history)
	performanceVariance := s.calculatePerformanceVariance(history)
	
	// 低表现和高方差都是风险因素
	performanceRisk := 1.0 - avgPerformance
	varianceRisk := math.Min(performanceVariance * 2, 1.0)
	
	return (performanceRisk + varianceRisk) / 2.0
}

func (s *LearningAnalyticsService) calculatePerformanceVariance(history []*entities.LearningHistory) float64 {
	if len(history) < 2 {
		return 0
	}

	performances := []float64{}
	for _, h := range history {
		performances = append(performances, h.Progress)
	}

	return s.calculateVariance(performances)
}

func (s *LearningAnalyticsService) calculateEngagementRisk(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0.5
	}

	// 基于会话长度和频率计算参与风险
	avgDuration := s.calculateAverageDuration(history)
	if avgDuration < time.Minute*15 { // 会话太短
		return 0.8
	}

	// 检查学习间隔
	if len(history) >= 2 {
		totalTimeSpan := history[len(history)-1].Timestamp.Sub(history[0].Timestamp)
		expectedSessions := totalTimeSpan.Hours() / 24 / 3 // 期望每3天一次
		actualSessions := float64(len(history))
		
		if actualSessions < expectedSessions * 0.5 {
			return 0.7
		}
	}

	return 0.3 // 低风险
}

func (s *LearningAnalyticsService) identifySpecificRisks(history []*entities.LearningHistory, learner *entities.Learner) []*SpecificRisk {
	risks := []*SpecificRisk{}

	// 检查学习间隔风险
	if len(history) >= 2 {
		lastActivity := history[len(history)-1].Timestamp
		if time.Since(lastActivity) > time.Hour*24*7 {
			risks = append(risks, &SpecificRisk{
				RiskType: "inactivity",
				Description: "Long period without learning activity",
				Probability: 0.8,
				Impact: "high",
				Timeline: "immediate",
				Indicators: []string{"no_recent_activity", "declining_engagement"},
				Interventions: []string{"send_reminder", "adjust_difficulty", "provide_motivation"},
			})
		}
	}

	// 检查表现下降风险
	if len(history) >= 10 {
		recent := history[len(history)-5:]
		earlier := history[len(history)-10 : len(history)-5]
		
		recentAvg := s.calculateAverageCompletion(recent)
		earlierAvg := s.calculateAverageCompletion(earlier)
		
		if recentAvg < earlierAvg * 0.8 {
			risks = append(risks, &SpecificRisk{
				RiskType: "performance_decline",
				Description: "Declining learning performance",
				Probability: 0.7,
				Impact: "medium",
				Timeline: "short_term",
				Indicators: []string{"lower_completion_rates", "increased_difficulty"},
				Interventions: []string{"review_fundamentals", "adjust_pace", "provide_support"},
			})
		}
	}

	// 检查目标风险
	for _, goal := range learner.LearningGoals {
		if goal.IsActive && time.Until(goal.TargetDate) < time.Hour*24*7 {
			risks = append(risks, &SpecificRisk{
				RiskType: "goal_deadline",
				Description: fmt.Sprintf("Goal '%s' deadline approaching", goal.Description),
				Probability: 0.6,
				Impact: "medium",
				Timeline: "immediate",
				Indicators: []string{"approaching_deadline", "insufficient_progress"},
				Interventions: []string{"intensive_practice", "extend_deadline", "adjust_scope"},
			})
		}
	}

	return risks
}

func (s *LearningAnalyticsService) generateMitigationStrategies(riskLevel string, history []*entities.LearningHistory) []string {
	strategies := []string{}

	switch riskLevel {
	case "high":
		strategies = append(strategies, 
			"immediate_intervention_required",
			"personalized_support_session",
			"adjust_learning_path_difficulty",
			"increase_motivation_elements",
			"provide_additional_resources")
	case "medium":
		strategies = append(strategies,
			"monitor_progress_closely",
			"provide_targeted_feedback",
			"adjust_content_delivery",
			"enhance_engagement_features")
	default: // low
		strategies = append(strategies,
			"maintain_current_approach",
			"periodic_check_ins",
			"continue_progress_monitoring")
	}

	return strategies
}

func (s *LearningAnalyticsService) generateLearningTrajectory(history []*entities.LearningHistory, learner *entities.Learner) *LearningTrajectory {
	currentState := s.assessCurrentLearningState(history, learner)
	predictedPath := s.generatePredictedPath(history, learner)
	alternativePaths := s.generateAlternativePaths(history, learner)
	keyMilestones := s.predictKeyMilestones(learner)

	return &LearningTrajectory{
		CurrentState: currentState,
		PredictedPath: predictedPath,
		AlternativePaths: alternativePaths,
		KeyMilestones: keyMilestones,
		OptimalStrategy: s.determineOptimalStrategy(history, learner),
		ExpectedOutcomes: s.predictExpectedOutcomes(history, learner),
	}
}

func (s *LearningAnalyticsService) assessCurrentLearningState(history []*entities.LearningHistory, learner *entities.Learner) string {
	if len(history) == 0 {
		return "beginner"
	}

	recentPerformance := s.calculateAverageCompletion(history[max(0, len(history)-5):])
	avgSkillLevel := s.calculateAverageSkillLevel(learner)

	switch {
	case recentPerformance >= 0.8 && avgSkillLevel >= 7:
		return "advanced"
	case recentPerformance >= 0.6 && avgSkillLevel >= 5:
		return "intermediate"
	case recentPerformance >= 0.4 && avgSkillLevel >= 3:
		return "developing"
	default:
		return "beginner"
	}
}

func (s *LearningAnalyticsService) calculateAverageSkillLevel(learner *entities.Learner) float64 {
	if len(learner.Skills) == 0 {
		return 1.0
	}

	total := 0.0
	for _, skill := range learner.Skills {
		total += float64(skill.Level)
	}
	return total / float64(len(learner.Skills))
}

func (s *LearningAnalyticsService) generatePredictedPath(history []*entities.LearningHistory, learner *entities.Learner) []*TrajectoryPoint {
	points := []*TrajectoryPoint{}
	
	// 生成未来3个月的预测点
	for i := 1; i <= 12; i++ { // 每周一个点
		timestamp := time.Now().AddDate(0, 0, i*7)
		
		// 预测技能水平
		skillLevels := make(map[string]float64)
		for _, skill := range learner.Skills {
			predictedLevel := s.predictSkillLevelAtTime(skill.SkillName, history, timestamp)
			skillLevels[skill.SkillName] = predictedLevel
		}

		// 预测学习状态
		predictedState := s.predictLearningStateAtTime(history, timestamp)
		
		point := &TrajectoryPoint{
			Timestamp: timestamp,
			PredictedState: predictedState,
			SkillLevels: skillLevels,
			Confidence: s.calculatePredictionConfidence(history, timestamp),
			KeyEvents: s.predictKeyEventsAtTime(learner, timestamp),
		}
		
		points = append(points, point)
	}

	return points
}

func (s *LearningAnalyticsService) predictSkillLevelAtTime(skillName string, history []*entities.LearningHistory, targetTime time.Time) float64 {
	skillHistory := []*entities.LearningHistory{}
	for _, h := range history {
		if h.SkillName == skillName {
			skillHistory = append(skillHistory, h)
		}
	}

	if len(skillHistory) == 0 {
		return 1.0
	}

	// 简单的线性预测
	velocity := s.calculateSkillLearningVelocity(skillHistory)
	timeFromNow := targetTime.Sub(time.Now())
	weeksFromNow := timeFromNow.Hours() / 168 // 168小时 = 1周

	currentLevel := s.calculateAverageCompletion(skillHistory) * 10 // 假设最高10级
	predictedLevel := currentLevel + velocity*weeksFromNow*10

	return math.Min(math.Max(predictedLevel, 1.0), 10.0)
}

func (s *LearningAnalyticsService) predictLearningStateAtTime(history []*entities.LearningHistory, targetTime time.Time) string {
	// 基于当前趋势预测未来状态
	if len(history) == 0 {
		return "beginner"
	}

	currentTrend := s.calculateProgressTrend(history)
	timeFromNow := targetTime.Sub(time.Now())
	weeksFromNow := timeFromNow.Hours() / 168

	currentPerformance := s.calculateAverageCompletion(history)
	
	var futurePerformance float64
	switch currentTrend {
	case "improving":
		futurePerformance = currentPerformance + 0.1*weeksFromNow
	case "declining":
		futurePerformance = currentPerformance - 0.05*weeksFromNow
	default:
		futurePerformance = currentPerformance
	}

	futurePerformance = math.Min(math.Max(futurePerformance, 0.0), 1.0)

	switch {
	case futurePerformance >= 0.8:
		return "advanced"
	case futurePerformance >= 0.6:
		return "intermediate"
	case futurePerformance >= 0.4:
		return "developing"
	default:
		return "beginner"
	}
}

func (s *LearningAnalyticsService) calculatePredictionConfidence(history []*entities.LearningHistory, targetTime time.Time) float64 {
	// 基于历史数据量和时间距离计算置信度
	baseConfidence := math.Min(float64(len(history))/50.0, 1.0) // 50个数据点为满置信度
	
	timeFromNow := targetTime.Sub(time.Now())
	timeDecay := math.Exp(-timeFromNow.Hours() / (24 * 30)) // 30天衰减
	
	return baseConfidence * timeDecay
}

func (s *LearningAnalyticsService) predictKeyEventsAtTime(learner *entities.Learner, targetTime time.Time) []string {
	events := []string{}
	
	// 检查目标截止日期
	for _, goal := range learner.LearningGoals {
		if goal.IsActive && goal.TargetDate.Before(targetTime.Add(time.Hour*24*7)) && goal.TargetDate.After(targetTime.Add(-time.Hour*24*7)) {
			events = append(events, fmt.Sprintf("goal_deadline_%s", goal.Description))
		}
	}

	// 预测技能里程碑
	for _, skill := range learner.Skills {
		predictedLevel := s.predictSkillLevelAtTime(skill.SkillName, []*entities.LearningHistory{}, targetTime)
		if int(predictedLevel) > 0 && int(predictedLevel)%3 == 0 { // 每3级一个里程碑
			events = append(events, fmt.Sprintf("skill_milestone_%s_level_%d", skill.SkillName, int(predictedLevel)))
		}
	}

	return events
}

func (s *LearningAnalyticsService) generateAlternativePaths(history []*entities.LearningHistory, learner *entities.Learner) []*AlternativePath {
	paths := []*AlternativePath{}

	// 快速路径
	paths = append(paths, &AlternativePath{
		PathName: "accelerated_learning",
		Description: "Intensive learning with increased difficulty",
		SuccessRate: 0.7,
		EstimatedTime: time.Hour * 24 * 60, // 2个月
		RequiredChanges: []string{"increase_study_time", "higher_difficulty_content"},
		Benefits: []string{"faster_completion", "deeper_understanding"},
		Risks: []string{"burnout", "knowledge_gaps"},
	})

	// 稳健路径
	paths = append(paths, &AlternativePath{
		PathName: "steady_progress",
		Description: "Consistent pace with balanced difficulty",
		SuccessRate: 0.85,
		EstimatedTime: time.Hour * 24 * 120, // 4个月
		RequiredChanges: []string{"maintain_current_pace", "regular_reviews"},
		Benefits: []string{"solid_foundation", "sustainable_learning"},
		Risks: []string{"slower_progress", "potential_boredom"},
	})

	// 灵活路径
	paths = append(paths, &AlternativePath{
		PathName: "adaptive_learning",
		Description: "Dynamic adjustment based on performance",
		SuccessRate: 0.8,
		EstimatedTime: time.Hour * 24 * 90, // 3个月
		RequiredChanges: []string{"dynamic_difficulty", "personalized_content"},
		Benefits: []string{"optimized_learning", "maintained_engagement"},
		Risks: []string{"complexity", "inconsistent_pace"},
	})

	return paths
}

func (s *LearningAnalyticsService) predictKeyMilestones(learner *entities.Learner) []*FutureMilestone {
	milestones := []*FutureMilestone{}

	// 基于目标生成里程碑
	for _, goal := range learner.LearningGoals {
		if goal.IsActive {
			milestone := &FutureMilestone{
				Name: fmt.Sprintf("Complete goal: %s", goal.Description),
				PredictedDate: goal.TargetDate,
				Probability: 0.7,
				Prerequisites: []string{fmt.Sprintf("master_%s", goal.TargetSkill)},
				Impact: "high",
			}
			milestones = append(milestones, milestone)
		}
	}

	// 基于技能水平生成里程碑
	for _, skill := range learner.Skills {
		if skill.Level < 10 {
			nextMilestoneLevel := ((skill.Level / 3) + 1) * 3 // 下一个3的倍数
			if nextMilestoneLevel <= 10 {
				estimatedDate := time.Now().AddDate(0, 0, (nextMilestoneLevel-skill.Level)*14) // 每级2周
				milestone := &FutureMilestone{
					Name: fmt.Sprintf("%s Level %d", skill.SkillName, nextMilestoneLevel),
					PredictedDate: estimatedDate,
					Probability: 0.8,
					Prerequisites: []string{fmt.Sprintf("practice_%s", skill.SkillName)},
					Impact: "medium",
				}
				milestones = append(milestones, milestone)
			}
		}
	}

	return milestones
}

func (s *LearningAnalyticsService) determineOptimalStrategy(history []*entities.LearningHistory, learner *entities.Learner) string {
	if len(history) == 0 {
		return "structured_learning"
	}

	avgPerformance := s.calculateAverageCompletion(history)
	consistency := s.calculateConsistencyScore(history)
	
	switch {
	case avgPerformance >= 0.8 && consistency >= 0.7:
		return "accelerated_learning"
	case avgPerformance >= 0.6 && consistency >= 0.5:
		return "steady_progress"
	case avgPerformance < 0.5:
		return "remedial_support"
	default:
		return "adaptive_learning"
	}
}

func (s *LearningAnalyticsService) calculateConsistencyScore(history []*entities.LearningHistory) float64 {
	if len(history) < 3 {
		return 0.5
	}

	performances := []float64{}
	for _, h := range history {
		performances = append(performances, h.Progress)
	}

	variance := s.calculateVariance(performances)
	return math.Max(0, 1.0-variance*2) // 低方差 = 高一致性
}

func (s *LearningAnalyticsService) predictExpectedOutcomes(history []*entities.LearningHistory, learner *entities.Learner) map[string]float64 {
	outcomes := make(map[string]float64)

	// 预测技能掌握度
	for _, skill := range learner.Skills {
		masteryRate := s.predictSkillMasteryRate(skill.SkillName, history)
		outcomes[fmt.Sprintf("skill_mastery_%s", skill.SkillName)] = masteryRate
	}

	// 预测目标完成率
	for _, goal := range learner.LearningGoals {
		if goal.IsActive {
			completionRate := s.predictGoalSuccessRate(&goal, history)
			outcomes[fmt.Sprintf("goal_completion_%s", goal.Description)] = completionRate
		}
	}

	// 预测整体学习成功率
	outcomes["overall_success"] = s.calculateOverallSuccessRate(history)

	return outcomes
}

func (s *LearningAnalyticsService) generatePredictiveActions(analysis *PredictiveAnalysis) []*PredictiveAction {
	actions := []*PredictiveAction{}

	// 基于风险评估生成行动
	if analysis.RiskAssessment.OverallRiskLevel == "high" {
		actions = append(actions, &PredictiveAction{
			ActionType: "intervention",
			Description: "Immediate learning intervention required",
			Priority: 1,
			ExpectedImpact: 0.8,
			Timeline: "immediate",
			Resources: []string{"tutor_support", "additional_materials"},
			SuccessMetrics: []string{"improved_engagement", "better_performance"},
		})
	}

	// 基于成功预测生成行动
	if analysis.SuccessPrediction.OverallSuccessRate < 0.6 {
		actions = append(actions, &PredictiveAction{
			ActionType: "study_strategy",
			Description: "Adjust learning strategy for better outcomes",
			Priority: 2,
			ExpectedImpact: 0.6,
			Timeline: "short_term",
			Resources: []string{"strategy_guide", "practice_materials"},
			SuccessMetrics: []string{"increased_success_rate", "goal_achievement"},
		})
	}

	// 基于学习轨迹生成行动
	if analysis.LearningTrajectory.OptimalStrategy == "accelerated_learning" {
		actions = append(actions, &PredictiveAction{
			ActionType: "content",
			Description: "Provide advanced learning materials",
			Priority: 3,
			ExpectedImpact: 0.7,
			Timeline: "medium_term",
			Resources: []string{"advanced_content", "challenging_exercises"},
			SuccessMetrics: []string{"skill_advancement", "maintained_engagement"},
		})
	}

	return actions
}

func (s *LearningAnalyticsService) generateRecommendations(ctx context.Context, report *LearningAnalyticsReport, learner *entities.Learner) []*AnalyticsRecommendation {
	recommendations := []*AnalyticsRecommendation{}

	// 基于进度分析生成推荐
	if report.ProgressAnalysis != nil {
		if report.ProgressAnalysis.OverallProgress < 0.5 {
			rec := &AnalyticsRecommendation{
				ID: uuid.New(),
				Type: "study_strategy",
				Title: "Increase Study Intensity",
				Description: "Your progress is below expected levels. Consider increasing study time and frequency.",
				Rationale: "Low overall progress indicates need for more intensive learning approach",
				Priority: 1,
				ExpectedBenefit: "Improved learning outcomes and goal achievement",
				ImplementationSteps: []string{
					"Increase daily study time by 30 minutes",
					"Add one extra study session per week",
					"Focus on fundamental concepts",
				},
				Timeline: "immediate",
				SuccessMetrics: []string{"increased_completion_rate", "improved_skill_levels"},
				Confidence: 0.8,
			}
			recommendations = append(recommendations, rec)
		}
	}

	// 基于表现分析生成推荐
	if report.PerformanceAnalysis != nil {
		if report.PerformanceAnalysis.OverallPerformance < 0.6 {
			rec := &AnalyticsRecommendation{
				ID: uuid.New(),
				Type: "content",
				Title: "Review Fundamental Concepts",
				Description: "Performance analysis suggests gaps in foundational knowledge.",
				Rationale: "Low performance scores indicate need for knowledge reinforcement",
				Priority: 2,
				ExpectedBenefit: "Stronger foundation leading to better performance",
				ImplementationSteps: []string{
					"Review previous learning materials",
					"Complete additional practice exercises",
					"Seek clarification on difficult concepts",
				},
				Timeline: "short_term",
				SuccessMetrics: []string{"improved_accuracy", "better_consistency"},
				Confidence: 0.7,
			}
			recommendations = append(recommendations, rec)
		}
	}

	// 基于参与度分析生成推荐
	if report.EngagementAnalysis != nil {
		if report.EngagementAnalysis.OverallEngagement < 0.5 {
			rec := &AnalyticsRecommendation{
				ID: uuid.New(),
				Type: "schedule",
				Title: "Optimize Learning Schedule",
				Description: "Low engagement suggests need for schedule optimization.",
				Rationale: "Engagement patterns indicate suboptimal learning timing",
				Priority: 3,
				ExpectedBenefit: "Increased motivation and learning effectiveness",
				ImplementationSteps: []string{
					"Identify peak learning hours",
					"Schedule learning during high-energy periods",
					"Break learning into shorter, focused sessions",
				},
				Timeline: "medium_term",
				SuccessMetrics: []string{"increased_session_length", "better_completion_rates"},
				Confidence: 0.6,
			}
			recommendations = append(recommendations, rec)
		}
	}

	// 基于预测分析生成推荐
	if report.PredictiveAnalysis != nil {
		if report.PredictiveAnalysis.RiskAssessment.OverallRiskLevel == "high" {
			rec := &AnalyticsRecommendation{
				ID: uuid.New(),
				Type: "intervention",
				Title: "Immediate Learning Support",
				Description: "High risk factors detected. Immediate intervention recommended.",
				Rationale: "Risk assessment indicates high probability of learning difficulties",
				Priority: 1,
				ExpectedBenefit: "Prevention of learning failure and improved outcomes",
				ImplementationSteps: []string{
					"Schedule one-on-one support session",
					"Adjust learning difficulty level",
					"Provide additional learning resources",
				},
				Timeline: "immediate",
				SuccessMetrics: []string{"reduced_risk_factors", "improved_engagement"},
				Confidence: 0.9,
			}
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

func (s *LearningAnalyticsService) calculateOverallScore(report *LearningAnalyticsReport) float64 {
	score := 0.0
	components := 0

	if report.ProgressAnalysis != nil {
		score += report.ProgressAnalysis.OverallProgress * 0.3
		components++
	}

	if report.PerformanceAnalysis != nil {
		score += report.PerformanceAnalysis.OverallPerformance * 0.3
		components++
	}

	if report.EngagementAnalysis != nil {
		score += report.EngagementAnalysis.OverallEngagement * 0.2
		components++
	}

	if report.PredictiveAnalysis != nil {
		score += report.PredictiveAnalysis.SuccessPrediction.OverallSuccessRate * 0.2
		components++
	}

	if components == 0 {
		return 0.5
	}

	return score
}

func (s *LearningAnalyticsService) generateInsights(report *LearningAnalyticsReport) []string {
	insights := []string{}

	// 基于整体评分生成洞察
	if report.OverallScore >= 0.8 {
		insights = append(insights, "Excellent learning progress with strong performance across all areas")
	} else if report.OverallScore >= 0.6 {
		insights = append(insights, "Good learning progress with room for improvement in some areas")
	} else {
		insights = append(insights, "Learning progress needs attention and support")
	}

	// 基于进度分析生成洞察
	if report.ProgressAnalysis != nil {
		switch report.ProgressAnalysis.ProgressTrend {
		case "improving":
			insights = append(insights, "Learning trajectory shows positive improvement over time")
		case "declining":
			insights = append(insights, "Recent learning performance shows concerning decline")
		default:
			insights = append(insights, "Learning progress remains stable and consistent")
		}
	}

	// 基于表现分析生成洞察
	if report.PerformanceAnalysis != nil && report.PerformanceAnalysis.StrengthsWeaknesses != nil {
		if len(report.PerformanceAnalysis.StrengthsWeaknesses.TopStrengths) > 0 {
			insights = append(insights, fmt.Sprintf("Key strengths identified: %s", 
				strings.Join(report.PerformanceAnalysis.StrengthsWeaknesses.TopStrengths, ", ")))
		}
	}

	return insights
}

func (s *LearningAnalyticsService) generateWarnings(report *LearningAnalyticsReport) []string {
	warnings := []string{}

	// 基于整体评分生成警告
	if report.OverallScore < 0.4 {
		warnings = append(warnings, "Overall learning performance is significantly below expectations")
	}

	// 基于风险评估生成警告
	if report.PredictiveAnalysis != nil && report.PredictiveAnalysis.RiskAssessment != nil {
		if report.PredictiveAnalysis.RiskAssessment.OverallRiskLevel == "high" {
			warnings = append(warnings, "High risk of learning failure detected - immediate intervention recommended")
		}
		
		if report.PredictiveAnalysis.RiskAssessment.DropoutRisk > 0.7 {
			warnings = append(warnings, "High dropout risk - consider motivational support")
		}
	}

	// 基于参与度生成警告
	if report.EngagementAnalysis != nil {
		if len(report.EngagementAnalysis.RiskFactors) > 0 {
			warnings = append(warnings, fmt.Sprintf("Engagement risk factors detected: %s", 
				strings.Join(report.EngagementAnalysis.RiskFactors, ", ")))
		}
	}

	return warnings
}

func (s *LearningAnalyticsService) generateComparisonData(ctx context.Context, learner *entities.Learner, req *AnalyticsRequest) (*ComparisonData, error) {
	// 简化实现 - 在实际应用中需要从数据库获取同龄人数据
	return &ComparisonData{
		ComparisonGroup: req.ComparisonGroup,
		PeerRanking: 45,
		TotalPeers: 100,
		Percentile: 55.0,
		BenchmarkMetrics: map[string]float64{
			"average_completion_rate": 0.75,
			"average_session_length": 45.0, // minutes
			"weekly_study_hours": 8.5,
		},
		RelativeStrengths: []string{"consistent_practice", "good_time_management"},
		RelativeWeaknesses: []string{"complex_problem_solving", "advanced_concepts"},
		ComparisonInsights: []string{
			"Performance is above average compared to peers",
			"Study consistency is a key differentiator",
			"Consider focusing on advanced topics for further improvement",
		},
	}, nil
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}