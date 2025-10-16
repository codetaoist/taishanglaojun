﻿package adaptive

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
)

// TimeSlot ?
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	DayOfWeek int       `json:"day_of_week"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RecommendedAction 
type RecommendedAction struct {
	ActionID        string                 `json:"action_id"`
	ActionType      string                 `json:"action_type"`
	Description     string                 `json:"description"`
	Priority        string                 `json:"priority"`
	Timeline        string                 `json:"timeline"`
	ExpectedOutcome string                 `json:"expected_outcome"`
	Parameters      map[string]interface{} `json:"parameters"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// MonitoringPlan 
type MonitoringPlan struct {
	PlanID              string                 `json:"plan_id"`
	Metrics             []string               `json:"metrics"`
	Frequency           time.Duration          `json:"frequency"`
	MonitoringFrequency string                 `json:"monitoring_frequency"`
	KeyMetrics          []MonitoringMetric     `json:"key_metrics"`
	Thresholds          map[string]float64     `json:"thresholds"`
	Actions             []RecommendedAction    `json:"actions"`
	Metadata            map[string]interface{} `json:"metadata"`
}



// RiskAssessment 
type RiskAssessment struct {
	AssessmentID string                 `json:"assessment_id"`
	Risks        []Risk                 `json:"risks"`
	OverallRisk  float64                `json:"overall_risk"`
	Mitigation   []RecommendedAction    `json:"mitigation"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Risk 
type Risk struct {
	RiskID      string                 `json:"risk_id"`
	RiskType    string                 `json:"risk_type"`
	Description string                 `json:"description"`
	Probability float64                `json:"probability"`
	Impact      float64                `json:"impact"`
	Severity    string                 `json:"severity"`
	Metadata    map[string]interface{} `json:"metadata"`
}



// AdaptiveLearningService 
type AdaptiveLearningService struct {
	learnerRepo         repositories.LearnerRepository
	contentRepo         repositories.LearningContentRepository
	knowledgeGraphRepo  repositories.KnowledgeGraphRepository
	analyticsService    interfaces.LearningAnalyticsService
	pathService         interfaces.LearningPathService
}

// NewAdaptiveLearningService 
func NewAdaptiveLearningService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	analyticsService interfaces.LearningAnalyticsService,
	pathService interfaces.LearningPathService,
) *AdaptiveLearningService {
	return &AdaptiveLearningService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		analyticsService:   analyticsService,
		pathService:        pathService,
	}
}

// PathAdaptationRequest 
type PathAdaptationRequest struct {
	LearnerID       string                    `json:"learner_id" binding:"required"`
	CurrentPathID   string                    `json:"current_path_id" binding:"required"`
	RecentProgress  []ProgressDataPoint       `json:"recent_progress"`
	PerformanceData PerformanceAnalysisData   `json:"performance_data"`
	ContextData     LearningContextData       `json:"context_data"`
	AdaptationGoals []AdaptationGoal          `json:"adaptation_goals"`
}

// ProgressDataPoint ?
type ProgressDataPoint struct {
	ContentID           string    `json:"content_id"`
	CompletionRate      float64   `json:"completion_rate"`
	TimeSpent           int       `json:"time_spent"`
	DifficultyLevel     string    `json:"difficulty_level"`
	PerformanceScore    float64   `json:"performance_score"`
	EngagementLevel     float64   `json:"engagement_level"`
	Timestamp           time.Time `json:"timestamp"`
	InteractionPatterns []string  `json:"interaction_patterns"`
}

// PerformanceAnalysisData 
type PerformanceAnalysisData struct {
	OverallAccuracy     float64            `json:"overall_accuracy"`
	SpeedMetrics        SpeedAnalysis      `json:"speed_metrics"`
	ConsistencyScore    float64            `json:"consistency_score"`
	StrengthAreas       []SkillStrength    `json:"strength_areas"`
	WeaknessAreas       []SkillWeakness    `json:"weakness_areas"`
	LearningTrends      []TrendAnalysis    `json:"learning_trends"`
	MotivationIndicators MotivationMetrics `json:"motivation_indicators"`
}

// SpeedAnalysis 
type SpeedAnalysis struct {
	AverageCompletionTime float64 `json:"average_completion_time"`
	SpeedTrend            string  `json:"speed_trend"` // "improving", "declining", "stable"
	OptimalPaceRange      struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"optimal_pace_range"`
}

// SkillStrength ?
type SkillStrength struct {
	SkillName       string  `json:"skill_name"`
	ProficiencyLevel float64 `json:"proficiency_level"`
	ConfidenceScore float64 `json:"confidence_score"`
	RecentImprovement float64 `json:"recent_improvement"`
}

// SkillWeakness ?
type SkillWeakness struct {
	SkillName       string  `json:"skill_name"`
	DeficiencyLevel float64 `json:"deficiency_level"`
	ImpactScore     float64 `json:"impact_score"`
	RecommendedFocus float64 `json:"recommended_focus"`
}

// TrendAnalysis 
type TrendAnalysis struct {
	MetricName  string  `json:"metric_name"`
	TrendType   string  `json:"trend_type"` // "improving", "declining", "stable", "volatile"
	ChangeRate  float64 `json:"change_rate"`
	Confidence  float64 `json:"confidence"`
	Prediction  float64 `json:"prediction"`
}

// MotivationMetrics 
type MotivationMetrics struct {
	EngagementScore     float64   `json:"engagement_score"`
	PersistenceLevel    float64   `json:"persistence_level"`
	ChallengePreference string    `json:"challenge_preference"`
	FeedbackResponsiveness float64 `json:"feedback_responsiveness"`
	GoalAlignment       float64   `json:"goal_alignment"`
}

// LearningContextData ?
type LearningContextData struct {
	TimeConstraints     TimeConstraints     `json:"time_constraints"`
	EnvironmentalFactors EnvironmentalFactors `json:"environmental_factors"`
	ResourceAvailability ResourceAvailability `json:"resource_availability"`
	SocialContext       SocialContext       `json:"social_context"`
}

// TimeConstraints 
type TimeConstraints struct {
	AvailableHoursPerWeek int       `json:"available_hours_per_week"`
	PreferredSessionLength int      `json:"preferred_session_length"`
	DeadlineConstraints    []Deadline `json:"deadline_constraints"`
	OptimalLearningTimes   []TimeSlot `json:"optimal_learning_times"`
}

// Deadline 
type Deadline struct {
	GoalID      string    `json:"goal_id"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Priority    string    `json:"priority"`
}

// TimeSlot ?
type AdaptiveTimeSlot struct {
	DayOfWeek string `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Efficiency float64 `json:"efficiency"`
}

// EnvironmentalFactors 
type EnvironmentalFactors struct {
	DeviceType        string   `json:"device_type"`
	NetworkQuality    string   `json:"network_quality"`
	DistractionLevel  float64  `json:"distraction_level"`
	SupportAvailability string `json:"support_availability"`
}

// ResourceAvailability ?
type ResourceAvailability struct {
	ContentTypes     []string `json:"content_types"`
	LanguageOptions  []string `json:"language_options"`
	AccessibilityNeeds []string `json:"accessibility_needs"`
	TechnicalLimitations []string `json:"technical_limitations"`
}

// SocialContext 罻?
type SocialContext struct {
	StudyGroupParticipation bool     `json:"study_group_participation"`
	MentorAvailability     bool     `json:"mentor_availability"`
	PeerInteractionLevel   float64  `json:"peer_interaction_level"`
	CompetitiveElements    bool     `json:"competitive_elements"`
}

// AdaptationGoal 
type AdaptationGoal struct {
	Type        string  `json:"type"` // "performance", "engagement", "efficiency", "retention"
	Priority    float64 `json:"priority"`
	TargetValue float64 `json:"target_value"`
	TimeFrame   string  `json:"time_frame"`
}

// PathAdaptationResponse 
type PathAdaptationResponse struct {
	AdaptedPath         *AdaptedLearningPath    `json:"adapted_path"`
	AdaptationReasoning AdaptationReasoning     `json:"adaptation_reasoning"`
	ImpactPrediction    ImpactPrediction        `json:"impact_prediction"`
	RecommendedActions  []RecommendedAction     `json:"recommended_actions"`
	MonitoringPlan      MonitoringPlan          `json:"monitoring_plan"`
}

// AdaptedLearningPath 
type AdaptedLearningPath struct {
	PathID              string                  `json:"path_id"`
	OriginalPathID      string                  `json:"original_path_id"`
	AdaptationLevel     string                  `json:"adaptation_level"` // "minor", "moderate", "major"
	ModifiedNodes       []ModifiedPathNode      `json:"modified_nodes"`
	AddedNodes          []PathNode              `json:"added_nodes"`
	RemovedNodes        []string                `json:"removed_nodes"`
	ReorderedSequence   []string                `json:"reordered_sequence"`
	DifficultyAdjustments []DifficultyAdjustment `json:"difficulty_adjustments"`
	PacingAdjustments   []PacingAdjustment      `json:"pacing_adjustments"`
	ContentSubstitutions []ContentSubstitution  `json:"content_substitutions"`
	EstimatedDuration   time.Duration           `json:"estimated_duration"`
	SuccessProbability  float64                 `json:"success_probability"`
}

// ModifiedPathNode ?
type ModifiedPathNode struct {
	NodeID              string                 `json:"node_id"`
	OriginalProperties  map[string]interface{} `json:"original_properties"`
	ModifiedProperties  map[string]interface{} `json:"modified_properties"`
	ModificationReason  string                 `json:"modification_reason"`
}

// PathNode 
type PathNode struct {
	NodeID              string                 `json:"node_id"`
	Title               string                 `json:"title"`
	Description         string                 `json:"description"`
	DifficultyLevel     string                 `json:"difficulty_level"`
	EstimatedDuration   time.Duration          `json:"estimated_duration"`
	Prerequisites       []string               `json:"prerequisites"`
	LearningObjectives  []string               `json:"learning_objectives"`
	ContentReferences   []string               `json:"content_references"`
	AssessmentCriteria  []string               `json:"assessment_criteria"`
	AdaptationMetadata  map[string]interface{} `json:"adaptation_metadata"`
}

// DifficultyAdjustment 
type DifficultyAdjustment struct {
	NodeID          string  `json:"node_id"`
	OriginalLevel   string  `json:"original_level"`
	AdjustedLevel   string  `json:"adjusted_level"`
	AdjustmentRatio float64 `json:"adjustment_ratio"`
	Reason          string  `json:"reason"`
}

// PacingAdjustment 
type PacingAdjustment struct {
	NodeID              string        `json:"node_id"`
	OriginalDuration    time.Duration `json:"original_duration"`
	AdjustedDuration    time.Duration `json:"adjusted_duration"`
	SpeedMultiplier     float64       `json:"speed_multiplier"`
	BreakRecommendations []BreakRecommendation `json:"break_recommendations"`
}

// BreakRecommendation 
type BreakRecommendation struct {
	AfterDuration time.Duration `json:"after_duration"`
	BreakLength   time.Duration `json:"break_length"`
	ActivityType  string        `json:"activity_type"`
}

// ContentSubstitution 滻
type ContentSubstitution struct {
	OriginalContentID string `json:"original_content_id"`
	SubstituteContentID string `json:"substitute_content_id"`
	SubstitutionReason string `json:"substitution_reason"`
	CompatibilityScore float64 `json:"compatibility_score"`
}

// AdaptationReasoning 
type AdaptationReasoning struct {
	PrimaryFactors      []ReasoningFactor   `json:"primary_factors"`
	SecondaryFactors    []ReasoningFactor   `json:"secondary_factors"`
	DecisionMatrix      DecisionMatrix      `json:"decision_matrix"`
	AlternativesConsidered []AdaptiveAlternative    `json:"alternatives_considered"`
	ConfidenceLevel     float64             `json:"confidence_level"`
}

// ReasoningFactor 
type ReasoningFactor struct {
	FactorType    string  `json:"factor_type"`
	Description   string  `json:"description"`
	Weight        float64 `json:"weight"`
	Evidence      []string `json:"evidence"`
	ImpactLevel   string  `json:"impact_level"`
}

// DecisionMatrix 
type DecisionMatrix struct {
	Criteria    []string              `json:"criteria"`
	Weights     []float64             `json:"weights"`
	Scores      map[string][]float64  `json:"scores"`
	FinalScores map[string]float64    `json:"final_scores"`
}

// AdaptiveAlternative ?
type AdaptiveAlternative struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Pros        []string `json:"pros"`
	Cons        []string `json:"cons"`
}

// ImpactPrediction 
type ImpactPrediction struct {
	PerformanceImpact   PerformanceImpact   `json:"performance_impact"`
	EngagementImpact    EngagementImpact    `json:"engagement_impact"`
	EfficiencyImpact    EfficiencyImpact    `json:"efficiency_impact"`
	RetentionImpact     RetentionImpact     `json:"retention_impact"`
	OverallImpact       OverallImpact       `json:"overall_impact"`
	RiskAssessment      RiskAssessment      `json:"risk_assessment"`
}

// PerformanceImpact 
type PerformanceImpact struct {
	ExpectedImprovement float64 `json:"expected_improvement"`
	ConfidenceInterval  struct {
		Lower float64 `json:"lower"`
		Upper float64 `json:"upper"`
	} `json:"confidence_interval"`
	TimeToImpact time.Duration `json:"time_to_impact"`
}

// EngagementImpact ?
type EngagementImpact struct {
	ExpectedChange     float64 `json:"expected_change"`
	MotivationFactors  []string `json:"motivation_factors"`
	RiskFactors        []string `json:"risk_factors"`
}

// EfficiencyImpact 
type EfficiencyImpact struct {
	TimeEfficiencyGain float64 `json:"time_efficiency_gain"`
	ResourceOptimization float64 `json:"resource_optimization"`
	CognitiveLoadReduction float64 `json:"cognitive_load_reduction"`
}

// RetentionImpact 
type RetentionImpact struct {
	ExpectedRetentionRate float64 `json:"expected_retention_rate"`
	LongTermBenefits     []string `json:"long_term_benefits"`
	ReinforcementNeeds   []string `json:"reinforcement_needs"`
}

// OverallImpact 
type OverallImpact struct {
	SuccessProbability float64 `json:"success_probability"`
	ExpectedROI        float64 `json:"expected_roi"`
	QualityScore       float64 `json:"quality_score"`
}

// RiskAssessment 
type AdaptiveRiskAssessment struct {
	OverallRiskLevel string      `json:"overall_risk_level"`
	SpecificRisks    []Risk      `json:"specific_risks"`
	MitigationStrategies []string `json:"mitigation_strategies"`
}

// Risk 
type AdaptiveRisk struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Probability float64 `json:"probability"`
	Impact      string  `json:"impact"`
	Severity    float64 `json:"severity"`
}

// RecommendedAction 
type AdaptiveRecommendedAction struct {
	ActionType    string    `json:"action_type"`
	Description   string    `json:"description"`
	Priority      string    `json:"priority"`
	Timeline      string    `json:"timeline"`
	ExpectedOutcome string  `json:"expected_outcome"`
	Resources     []string  `json:"resources"`
}

// MonitoringPlan 
type AdaptiveMonitoringPlan struct {
	MonitoringFrequency string              `json:"monitoring_frequency"`
	KeyMetrics          []MonitoringMetric  `json:"key_metrics"`
	TriggerConditions   []TriggerCondition  `json:"trigger_conditions"`
	ReviewSchedule      []ReviewPoint       `json:"review_schedule"`
}

// MonitoringMetric 
type MonitoringMetric struct {
	MetricName    string  `json:"metric_name"`
	TargetValue   float64 `json:"target_value"`
	ThresholdMin  float64 `json:"threshold_min"`
	ThresholdMax  float64 `json:"threshold_max"`
	MeasurementMethod string `json:"measurement_method"`
}

// TriggerCondition 
type TriggerCondition struct {
	ConditionType string  `json:"condition_type"`
	Threshold     float64 `json:"threshold"`
	Action        string  `json:"action"`
	Urgency       string  `json:"urgency"`
}

// ReviewPoint ?
type ReviewPoint struct {
	ScheduledTime time.Time `json:"scheduled_time"`
	ReviewType    string    `json:"review_type"`
	Objectives    []string  `json:"objectives"`
}

// AdaptLearningPath 
func (s *AdaptiveLearningService) AdaptLearningPath(ctx context.Context, req *PathAdaptationRequest) (*PathAdaptationResponse, error) {
	// ?
	learnerID, err := uuid.Parse(req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("invalid learner ID: %w", err)
	}
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 
	currentPath, err := s.getCurrentLearningPath(ctx, req.CurrentPathID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current path: %w", err)
	}

	// ?
	learnerState, err := s.analyzeLearnerCurrentState(ctx, learner, req)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze learner state: %w", err)
	}

	// ?
	adaptationNeeds, err := s.identifyAdaptationNeeds(ctx, learnerState, req)
	if err != nil {
		return nil, fmt.Errorf("failed to identify adaptation needs: %w", err)
	}

	// 
	adaptationStrategy, err := s.generateAdaptationStrategy(ctx, adaptationNeeds, currentPath, learner)
	if err != nil {
		return nil, fmt.Errorf("failed to generate adaptation strategy: %w", err)
	}

	// 
	adaptedPath, err := s.applyAdaptationStrategy(ctx, currentPath, adaptationStrategy, learner)
	if err != nil {
		return nil, fmt.Errorf("failed to apply adaptation strategy: %w", err)
	}

	// 
	impactPrediction, err := s.predictAdaptationImpact(ctx, adaptedPath, learner, req)
	if err != nil {
		return nil, fmt.Errorf("failed to predict impact: %w", err)
	}

	// 
	recommendedActions := s.generateRecommendedActions(adaptationStrategy, adaptedPath)

	// 
	monitoringPlan := s.createMonitoringPlan(adaptedPath, adaptationNeeds)

	// 
	reasoning := s.generateAdaptationReasoning(adaptationNeeds, adaptationStrategy, impactPrediction)

	return &PathAdaptationResponse{
		AdaptedPath:         adaptedPath,
		AdaptationReasoning: reasoning,
		ImpactPrediction:    impactPrediction,
		RecommendedActions:  recommendedActions,
		MonitoringPlan:      monitoringPlan,
	}, nil
}

// 汾

func (s *AdaptiveLearningService) getCurrentLearningPath(ctx context.Context, pathID string) (*entities.LearningPath, error) {
	// 
	// 
	return &entities.LearningPath{
		ID:          uuid.MustParse(pathID),
		Name:        "",
		Description: "",
	}, nil
}

func (s *AdaptiveLearningService) analyzeLearnerCurrentState(ctx context.Context, learner *entities.Learner, req *PathAdaptationRequest) (*LearnerCurrentState, error) {
	// 
	return &LearnerCurrentState{
		LearnerID:           learner.ID.String(),
		CurrentPerformance:  s.calculateCurrentPerformance(req.PerformanceData),
		LearningVelocity:    s.calculateLearningVelocity(req.RecentProgress),
		EngagementLevel:     s.calculateEngagementLevel(req.RecentProgress),
		MotivationLevel:     req.PerformanceData.MotivationIndicators.EngagementScore,
		CognitiveLoad:       s.estimateCognitiveLoad(req.RecentProgress),
		StressLevel:         s.estimateStressLevel(req.PerformanceData),
		AdaptationReadiness: s.assessAdaptationReadiness(learner, req),
	}, nil
}

// LearnerCurrentState ?
type LearnerCurrentState struct {
	LearnerID           string  `json:"learner_id"`
	CurrentPerformance  float64 `json:"current_performance"`
	LearningVelocity    float64 `json:"learning_velocity"`
	EngagementLevel     float64 `json:"engagement_level"`
	MotivationLevel     float64 `json:"motivation_level"`
	CognitiveLoad       float64 `json:"cognitive_load"`
	StressLevel         float64 `json:"stress_level"`
	AdaptationReadiness float64 `json:"adaptation_readiness"`
}

func (s *AdaptiveLearningService) calculateCurrentPerformance(data PerformanceAnalysisData) float64 {
	// 㵱
	return (data.OverallAccuracy + data.ConsistencyScore + data.MotivationIndicators.EngagementScore) / 3.0
}

func (s *AdaptiveLearningService) calculateLearningVelocity(progress []ProgressDataPoint) float64 {
	if len(progress) == 0 {
		return 1.0
	}

	// 
	var totalCompletion float64
	var totalTime float64

	for _, p := range progress {
		totalCompletion += p.CompletionRate
		totalTime += float64(p.TimeSpent)
	}

	if totalTime == 0 {
		return 1.0
	}

	return totalCompletion / (totalTime / 3600) // 
}

func (s *AdaptiveLearningService) calculateEngagementLevel(progress []ProgressDataPoint) float64 {
	if len(progress) == 0 {
		return 0.5
	}

	var totalEngagement float64
	for _, p := range progress {
		totalEngagement += p.EngagementLevel
	}

	return totalEngagement / float64(len(progress))
}

func (s *AdaptiveLearningService) estimateCognitiveLoad(progress []ProgressDataPoint) float64 {
	// ?
	if len(progress) == 0 {
		return 0.5
	}

	var totalLoad float64
	for _, p := range progress {
		difficultyWeight := s.getDifficultyWeight(p.DifficultyLevel)
		performanceAdjustment := 1.0 - p.PerformanceScore
		totalLoad += difficultyWeight * performanceAdjustment
	}

	return math.Min(totalLoad/float64(len(progress)), 1.0)
}

func (s *AdaptiveLearningService) getDifficultyWeight(level string) float64 {
	switch level {
	case "beginner":
		return 0.3
	case "intermediate":
		return 0.6
	case "advanced":
		return 0.9
	case "expert":
		return 1.0
	default:
		return 0.5
	}
}

func (s *AdaptiveLearningService) estimateStressLevel(data PerformanceAnalysisData) float64 {
	// 
	stressIndicators := []float64{
		1.0 - data.ConsistencyScore,
		1.0 - data.MotivationIndicators.PersistenceLevel,
		1.0 - data.MotivationIndicators.FeedbackResponsiveness,
	}

	var totalStress float64
	for _, indicator := range stressIndicators {
		totalStress += indicator
	}

	return totalStress / float64(len(stressIndicators))
}

func (s *AdaptiveLearningService) assessAdaptationReadiness(learner *entities.Learner, req *PathAdaptationRequest) float64 {
	// ?
	readinessFactors := []float64{
		s.calculateExperienceReadiness(learner),
		s.calculatePerformanceReadiness(req.PerformanceData),
		s.calculateMotivationReadiness(req.PerformanceData.MotivationIndicators),
		s.calculateContextReadiness(req.ContextData),
	}

	var totalReadiness float64
	for _, factor := range readinessFactors {
		totalReadiness += factor
	}

	return totalReadiness / float64(len(readinessFactors))
}

func (s *AdaptiveLearningService) calculateExperienceReadiness(learner *entities.Learner) float64 {
	// 
	if learner.Experience > 1000 {
		return 0.9
	} else if learner.Experience > 500 {
		return 0.7
	} else if learner.Experience > 100 {
		return 0.5
	}
	return 0.3
}

func (s *AdaptiveLearningService) calculatePerformanceReadiness(data PerformanceAnalysisData) float64 {
	// ?
	return (data.OverallAccuracy + data.ConsistencyScore) / 2.0
}

func (s *AdaptiveLearningService) calculateMotivationReadiness(motivation MotivationMetrics) float64 {
	// ?
	return (motivation.EngagementScore + motivation.PersistenceLevel + motivation.GoalAlignment) / 3.0
}

func (s *AdaptiveLearningService) calculateContextReadiness(context LearningContextData) float64 {
	// 
	timeReadiness := float64(context.TimeConstraints.AvailableHoursPerWeek) / 40.0 // 40?
	if timeReadiness > 1.0 {
		timeReadiness = 1.0
	}

	environmentReadiness := 1.0 - context.EnvironmentalFactors.DistractionLevel
	
	return (timeReadiness + environmentReadiness) / 2.0
}

// ?..

func (s *AdaptiveLearningService) identifyAdaptationNeeds(ctx context.Context, state *LearnerCurrentState, req *PathAdaptationRequest) ([]AdaptationNeed, error) {
	var needs []AdaptationNeed

	// ?
	if state.CurrentPerformance < 0.6 {
		needs = append(needs, AdaptationNeed{
			Type:        "difficulty_reduction",
			Urgency:     "high",
			Description: "?,
			Evidence:    []string{"60%"},
		})
	}

	// ?
	if state.EngagementLevel < 0.5 {
		needs = append(needs, AdaptationNeed{
			Type:        "engagement_enhancement",
			Urgency:     "medium",
			Description: "?,
			Evidence:    []string{"?0%"},
		})
	}

	// ?
	if state.CognitiveLoad > 0.8 {
		needs = append(needs, AdaptationNeed{
			Type:        "cognitive_load_reduction",
			Urgency:     "high",
			Description: "",
			Evidence:    []string{"80%"},
		})
	}

	return needs, nil
}

// AdaptationNeed ?
type AdaptationNeed struct {
	Type        string   `json:"type"`
	Urgency     string   `json:"urgency"`
	Description string   `json:"description"`
	Evidence    []string `json:"evidence"`
	Priority    float64  `json:"priority"`
}

func (s *AdaptiveLearningService) generateAdaptationStrategy(ctx context.Context, needs []AdaptationNeed, currentPath *entities.LearningPath, learner *entities.Learner) (*AdaptationStrategy, error) {
	// ?
	return &AdaptationStrategy{
		StrategyID:   uuid.New().String(),
		StrategyType: "comprehensive",
		Adaptations:  s.createAdaptationsFromNeeds(needs),
		Priority:     s.calculateStrategyPriority(needs),
		Confidence:   s.calculateStrategyConfidence(needs, learner),
	}, nil
}

// AdaptationStrategy 
type AdaptationStrategy struct {
	StrategyID   string       `json:"strategy_id"`
	StrategyType string       `json:"strategy_type"`
	Adaptations  []Adaptation `json:"adaptations"`
	Priority     float64      `json:"priority"`
	Confidence   float64      `json:"confidence"`
}

// Adaptation 
type Adaptation struct {
	Type        string                 `json:"type"`
	Target      string                 `json:"target"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Rationale   string                 `json:"rationale"`
}

func (s *AdaptiveLearningService) createAdaptationsFromNeeds(needs []AdaptationNeed) []Adaptation {
	var adaptations []Adaptation

	for _, need := range needs {
		switch need.Type {
		case "difficulty_reduction":
			adaptations = append(adaptations, Adaptation{
				Type:   "difficulty_adjustment",
				Target: "content_difficulty",
				Action: "reduce",
				Parameters: map[string]interface{}{
					"reduction_factor": 0.2,
					"gradual":         true,
				},
				Rationale: need.Description,
			})
		case "engagement_enhancement":
			adaptations = append(adaptations, Adaptation{
				Type:   "content_substitution",
				Target: "content_type",
				Action: "increase_interactivity",
				Parameters: map[string]interface{}{
					"interactive_ratio": 0.6,
					"gamification":     true,
				},
				Rationale: need.Description,
			})
		case "cognitive_load_reduction":
			adaptations = append(adaptations, Adaptation{
				Type:   "pacing_adjustment",
				Target: "session_length",
				Action: "reduce_and_add_breaks",
				Parameters: map[string]interface{}{
					"session_reduction": 0.3,
					"break_frequency":   15, // ?5
				},
				Rationale: need.Description,
			})
		}
	}

	return adaptations
}

func (s *AdaptiveLearningService) calculateStrategyPriority(needs []AdaptationNeed) float64 {
	var totalPriority float64
	for _, need := range needs {
		switch need.Urgency {
		case "high":
			totalPriority += 1.0
		case "medium":
			totalPriority += 0.6
		case "low":
			totalPriority += 0.3
		}
	}
	return math.Min(totalPriority/float64(len(needs)), 1.0)
}

func (s *AdaptiveLearningService) calculateStrategyConfidence(needs []AdaptationNeed, learner *entities.Learner) float64 {
	// 
	baseConfidence := 0.7
	
	// ?
	if learner.Experience > 500 {
		baseConfidence += 0.2
	}
	
	// ?
	if len(needs) > 3 {
		baseConfidence -= 0.1
	}
	
	return math.Max(math.Min(baseConfidence, 1.0), 0.0)
}

// ...

func (s *AdaptiveLearningService) applyAdaptationStrategy(ctx context.Context, currentPath *entities.LearningPath, strategy *AdaptationStrategy, learner *entities.Learner) (*AdaptedLearningPath, error) {
	// ?
	return &AdaptedLearningPath{
		PathID:              uuid.New().String(),
		OriginalPathID:      currentPath.ID.String(),
		AdaptationLevel:     "moderate",
		EstimatedDuration:   time.Hour * 20, // ?
		SuccessProbability:  0.85,           // ?
	}, nil
}

func (s *AdaptiveLearningService) predictAdaptationImpact(ctx context.Context, adaptedPath *AdaptedLearningPath, learner *entities.Learner, req *PathAdaptationRequest) (ImpactPrediction, error) {
	// ?
	return ImpactPrediction{
		PerformanceImpact: PerformanceImpact{
			ExpectedImprovement: 0.15,
			TimeToImpact:       time.Hour * 24,
		},
		EngagementImpact: EngagementImpact{
			ExpectedChange: 0.2,
		},
		OverallImpact: OverallImpact{
			SuccessProbability: adaptedPath.SuccessProbability,
			ExpectedROI:        1.3,
			QualityScore:       0.8,
		},
	}, nil
}

func (s *AdaptiveLearningService) generateRecommendedActions(strategy *AdaptationStrategy, adaptedPath *AdaptedLearningPath) []RecommendedAction {
	// ?
	return []RecommendedAction{
		{
			ActionType:      "immediate",
			Description:     "",
			Priority:        "high",
			Timeline:        "",
			ExpectedOutcome: "",
		},
	}
}

func (s *AdaptiveLearningService) createMonitoringPlan(adaptedPath *AdaptedLearningPath, needs []AdaptationNeed) MonitoringPlan {
	// ?
	return MonitoringPlan{
		MonitoringFrequency: "daily",
		KeyMetrics: []MonitoringMetric{
			{
				MetricName:   "performance_score",
				TargetValue:  0.8,
				ThresholdMin: 0.6,
				ThresholdMax: 1.0,
			},
		},
	}
}

func (s *AdaptiveLearningService) generateAdaptationReasoning(needs []AdaptationNeed, strategy *AdaptationStrategy, impact ImpactPrediction) AdaptationReasoning {
	// ?
	return AdaptationReasoning{
		PrimaryFactors: []ReasoningFactor{
			{
				FactorType:  "performance",
				Description: "?,
				Weight:      0.8,
				ImpactLevel: "high",
			},
		},
		ConfidenceLevel: strategy.Confidence,
	}
}

