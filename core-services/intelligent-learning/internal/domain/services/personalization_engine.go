package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// PersonalizationEngine 个性化引擎
type PersonalizationEngine struct {
	learnerRepo         repositories.LearnerRepository
	contentRepo         repositories.LearningContentRepository
	knowledgeGraphRepo  repositories.KnowledgeGraphRepository
	behaviorTracker     *UserBehaviorTracker
	preferenceAnalyzer  *PreferenceAnalyzer
	contextAnalyzer     *ContextAnalyzer
	recommendationModel RecommendationModel
}

// NewPersonalizationEngine 创建个性化引擎
func NewPersonalizationEngine(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	behaviorTracker *UserBehaviorTracker,
	preferenceAnalyzer *PreferenceAnalyzer,
	contextAnalyzer *ContextAnalyzer,
	recommendationModel RecommendationModel,
) *PersonalizationEngine {
	return &PersonalizationEngine{
		learnerRepo:         learnerRepo,
		contentRepo:         contentRepo,
		knowledgeGraphRepo:  knowledgeGraphRepo,
		behaviorTracker:     behaviorTracker,
		preferenceAnalyzer:  preferenceAnalyzer,
		contextAnalyzer:     contextAnalyzer,
		recommendationModel: recommendationModel,
	}
}

// PersonalizationRequest 个性化请求
type PersonalizationRequest struct {
	LearnerID           uuid.UUID                      `json:"learner_id"`
	Context             *PersonalizationContext        `json:"context"`
	RecommendationType  string                         `json:"recommendation_type"` // content, path, concept, action
	MaxRecommendations  int                            `json:"max_recommendations"`
	IncludeExplanations bool                           `json:"include_explanations"`
	Filters             map[string]interface{}         `json:"filters"`
	PersonalizationLevel string                        `json:"personalization_level"` // basic, advanced, deep
}

// PersonalizationResponse 个性化响应
type PersonalizationResponse struct {
	LearnerID              uuid.UUID                    `json:"learner_id"`
	Recommendations        []PersonalizedRecommendation `json:"recommendations"`
	PersonalizationFactors PersonalizationFactors       `json:"personalization_factors"`
	ContextualInsights     []ContextualInsight          `json:"contextual_insights"`
	AdaptationSuggestions  []AdaptationSuggestion       `json:"adaptation_suggestions"`
	GeneratedAt            time.Time                    `json:"generated_at"`
	ValidUntil             time.Time                    `json:"valid_until"`
	Confidence             float64                      `json:"confidence"`
}

// PersonalizedRecommendation 个性化推荐
type PersonalizedRecommendation struct {
	ID                  uuid.UUID                      `json:"id"`
	Type                string                         `json:"type"`
	ContentID           *uuid.UUID                     `json:"content_id,omitempty"`
	PathID              *uuid.UUID                     `json:"path_id,omitempty"`
	ConceptID           *uuid.UUID                     `json:"concept_id,omitempty"`
	Title               string                         `json:"title"`
	Description         string                         `json:"description"`
	Score               float64                        `json:"score"`
	Confidence          float64                        `json:"confidence"`
	PersonalizationScore float64                       `json:"personalization_score"`
	Reasoning           []string                       `json:"reasoning"`
	ExpectedBenefit     string                         `json:"expected_benefit"`
	EstimatedTime       time.Duration                  `json:"estimated_time"`
	Difficulty          string                         `json:"difficulty"`
	Prerequisites       []uuid.UUID                    `json:"prerequisites"`
	Tags                []string                       `json:"tags"`
	Metadata            map[string]interface{}         `json:"metadata"`
}

// PersonalizationFactors 个性化因子
type PersonalizationFactors struct {
	LearningStyle       LearningStyleProfile    `json:"learning_style"`
	SkillLevel          SkillLevelProfile       `json:"skill_level"`
	Preferences         PreferenceProfile       `json:"preferences"`
	BehaviorPattern     BehaviorPattern         `json:"behavior_pattern"`
	ContextualFactors   ContextualFactors       `json:"contextual_factors"`
	MotivationFactors   MotivationFactors       `json:"motivation_factors"`
	AdaptationHistory   []AdaptationRecord      `json:"adaptation_history"`
}

// LearningStyleProfile 学习风格档案
type LearningStyleProfile struct {
	VisualPreference    float64 `json:"visual_preference"`
	AuditoryPreference  float64 `json:"auditory_preference"`
	KinestheticPreference float64 `json:"kinesthetic_preference"`
	ReadingPreference   float64 `json:"reading_preference"`
	InteractivePreference float64 `json:"interactive_preference"`
	PacePreference      string  `json:"pace_preference"` // slow, medium, fast
	DepthPreference     string  `json:"depth_preference"` // surface, deep, comprehensive
}

// SkillLevelProfile 技能水平档�?
type SkillLevelProfile struct {
	OverallLevel        float64            `json:"overall_level"`
	DomainLevels        map[string]float64 `json:"domain_levels"`
	SkillGaps           []SkillGap         `json:"skill_gaps"`
	StrengthAreas       []string           `json:"strength_areas"`
	ImprovementAreas    []string           `json:"improvement_areas"`
	LearningVelocity    float64            `json:"learning_velocity"`
}

// PreferenceProfile 偏好档案
type PreferenceProfile struct {
	ContentTypes        []string           `json:"content_types"`
	DifficultyTolerance float64            `json:"difficulty_tolerance"`
	SessionDuration     time.Duration      `json:"session_duration"`
	TimeOfDay           []string           `json:"time_of_day"`
	DevicePreferences   []string           `json:"device_preferences"`
	LanguagePreferences []string           `json:"language_preferences"`
	TopicInterests      map[string]float64 `json:"topic_interests"`
}

// BehaviorPattern 行为模式
type BehaviorPattern struct {
	EngagementLevel     float64                `json:"engagement_level"`
	CompletionRate      float64                `json:"completion_rate"`
	SessionFrequency    float64                `json:"session_frequency"`
	AverageSessionTime  time.Duration          `json:"average_session_time"`
	InteractionPatterns map[string]float64     `json:"interaction_patterns"`
	DropoffPoints       []DropoffPoint         `json:"dropoff_points"`
	PeakPerformanceTimes []time.Time           `json:"peak_performance_times"`
}

// ContextualFactors 上下文因�?
type ContextualFactors struct {
	CurrentTime         time.Time              `json:"current_time"`
	Device              string                 `json:"device"`
	Location            string                 `json:"location"`
	NetworkCondition    string                 `json:"network_condition"`
	AvailableTime       time.Duration          `json:"available_time"`
	EnergyLevel         float64                `json:"energy_level"`
	DistractionLevel    float64                `json:"distraction_level"`
	SocialContext       string                 `json:"social_context"`
}

// MotivationFactors 动机因子
type MotivationFactors struct {
	IntrinsicMotivation float64                `json:"intrinsic_motivation"`
	ExtrinsicMotivation float64                `json:"extrinsic_motivation"`
	GoalOrientation     string                 `json:"goal_orientation"`
	AchievementLevel    float64                `json:"achievement_level"`
	ChallengePreference float64                `json:"challenge_preference"`
	FeedbackPreference  string                 `json:"feedback_preference"`
	RewardSensitivity   float64                `json:"reward_sensitivity"`
}

// ContextualInsight 上下文洞�?
type ContextualInsight struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      float64                `json:"impact"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AdaptationSuggestion 适应性建�?
type AdaptationSuggestion struct {
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Priority        int                    `json:"priority"`
	ExpectedImpact  float64                `json:"expected_impact"`
	Implementation  []string               `json:"implementation"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SkillGap 技能差�?
type SkillGap struct {
	Skill       string  `json:"skill"`
	CurrentLevel float64 `json:"current_level"`
	TargetLevel  float64 `json:"target_level"`
	Priority     int     `json:"priority"`
}

// DropoffPoint 流失�?
type DropoffPoint struct {
	ContentID   uuid.UUID `json:"content_id"`
	Position    float64   `json:"position"`
	Frequency   int       `json:"frequency"`
	Reason      string    `json:"reason"`
}

// AdaptationRecord 适应记录
type AdaptationRecord struct {
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Effectiveness float64              `json:"effectiveness"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PersonalizationContext 个性化上下�?
type PersonalizationContext struct {
	SessionID       string                 `json:"session_id"`
	Device          string                 `json:"device"`
	Location        string                 `json:"location"`
	TimeOfDay       string                 `json:"time_of_day"`
	AvailableTime   int                    `json:"available_time"` // 分钟
	EnergyLevel     float64                `json:"energy_level"`   // 0-1
	Goals           []string               `json:"goals"`
	CurrentContent  string                 `json:"current_content"`
	RecentActivity  []string               `json:"recent_activity"`
	SocialContext   string                 `json:"social_context"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ActivityRecord 活动记录
type ActivityRecord struct {
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	ContentID   uuid.UUID              `json:"content_id"`
	Duration    time.Duration          `json:"duration"`
	Performance float64                `json:"performance"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GeneratePersonalizedRecommendations 生成个性化推荐
func (pe *PersonalizationEngine) GeneratePersonalizedRecommendations(ctx context.Context, req *PersonalizationRequest) (*PersonalizationResponse, error) {
	// 获取学习者信�?
	learner, err := pe.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 分析个性化因子
	factors, err := pe.analyzePersonalizationFactors(ctx, learner, req.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze personalization factors: %w", err)
	}

	// 生成候选推�?
	candidates, err := pe.generateCandidateRecommendations(ctx, req, factors)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidate recommendations: %w", err)
	}

	// 个性化评分
	personalizedRecs, err := pe.personalizeRecommendations(ctx, candidates, factors, req)
	if err != nil {
		return nil, fmt.Errorf("failed to personalize recommendations: %w", err)
	}

	// 生成上下文洞�?
	insights := pe.generateContextualInsights(ctx, factors, req.Context)

	// 生成适应性建�?
	adaptations := pe.generateAdaptationSuggestions(ctx, factors, personalizedRecs)

	// 计算整体置信�?
	confidence := pe.calculateOverallConfidence(factors, personalizedRecs)

	response := &PersonalizationResponse{
		LearnerID:              req.LearnerID,
		Recommendations:        personalizedRecs,
		PersonalizationFactors: *factors,
		ContextualInsights:     insights,
		AdaptationSuggestions:  adaptations,
		GeneratedAt:            time.Now(),
		ValidUntil:             time.Now().Add(2 * time.Hour), // 2小时有效�?
		Confidence:             confidence,
	}

	return response, nil
}

// analyzePersonalizationFactors 分析个性化因子
func (pe *PersonalizationEngine) analyzePersonalizationFactors(ctx context.Context, learner *entities.Learner, context *PersonalizationContext) (*PersonalizationFactors, error) {
	// 分析学习风格
	learningStyle := pe.analyzeLearningStyle(ctx, learner)

	// 分析技能水�?
	skillLevel := pe.analyzeSkillLevel(ctx, learner)

	// 分析偏好
	preferences := pe.analyzePreferences(ctx, learner)

	// 分析行为模式
	behaviorPattern, err := pe.behaviorTracker.AnalyzeBehaviorPattern(ctx, learner.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze behavior pattern: %w", err)
	}

	// 分析上下文因�?
	contextualFactors := pe.analyzeContextualFactors(context)

	// 分析动机因子
	motivationFactors := pe.analyzeMotivationFactors(ctx, learner)

	// 获取适应历史
	adaptationHistory := pe.getAdaptationHistory(ctx, learner.ID)

	factors := &PersonalizationFactors{
		LearningStyle:     *learningStyle,
		SkillLevel:        *skillLevel,
		Preferences:       *preferences,
		BehaviorPattern:   *behaviorPattern,
		ContextualFactors: *contextualFactors,
		MotivationFactors: *motivationFactors,
		AdaptationHistory: adaptationHistory,
	}

	return factors, nil
}

// analyzeLearningStyle 分析学习风格
func (pe *PersonalizationEngine) analyzeLearningStyle(ctx context.Context, learner *entities.Learner) *LearningStyleProfile {
	// 基于学习者偏好和历史行为分析学习风格
	profile := &LearningStyleProfile{
		VisualPreference:      0.5,
		AuditoryPreference:    0.3,
		KinestheticPreference: 0.4,
		ReadingPreference:     0.6,
		InteractivePreference: 0.7,
		PacePreference:        "medium",
		DepthPreference:       "deep",
	}

	// 根据学习者偏好调�?
	if learner.Preferences.Style == "visual" {
		profile.VisualPreference = 0.9
	} else if learner.Preferences.Style == "auditory" {
		profile.AuditoryPreference = 0.9
	} else if learner.Preferences.Style == "kinesthetic" {
		profile.KinestheticPreference = 0.9
	}

	return profile
}

// analyzeSkillLevel 分析技能水�?
func (pe *PersonalizationEngine) analyzeSkillLevel(ctx context.Context, learner *entities.Learner) *SkillLevelProfile {
	// 将技能列表转换为map以便处理
	skillMap := make(map[string]float64)
	var totalLevel float64
	
	for _, skill := range learner.Skills {
		// 将技能等级转换为0-1的浮点数
		normalizedLevel := float64(skill.Level) / 10.0
		skillMap[skill.SkillName] = normalizedLevel
		totalLevel += normalizedLevel
	}
	
	var overallLevel float64
	if len(learner.Skills) > 0 {
		overallLevel = totalLevel / float64(len(learner.Skills))
	}

	// 识别技能差�?
	skillGaps := make([]SkillGap, 0)
	for skillName, level := range skillMap {
		if level < 0.7 { // 低于70%认为是技能差�?
			skillGaps = append(skillGaps, SkillGap{
				Skill:        skillName,
				CurrentLevel: level,
				TargetLevel:  0.8,
				Priority:     int((0.8 - level) * 10),
			})
		}
	}

	profile := &SkillLevelProfile{
		OverallLevel:     overallLevel,
		DomainLevels:     skillMap,
		SkillGaps:        skillGaps,
		StrengthAreas:    pe.identifyStrengthAreas(skillMap),
		ImprovementAreas: pe.identifyImprovementAreas(skillMap),
		LearningVelocity: pe.calculateLearningVelocity(ctx, learner.ID),
	}

	return profile
}

// analyzePreferences 分析偏好
func (pe *PersonalizationEngine) analyzePreferences(ctx context.Context, learner *entities.Learner) *PreferenceProfile {
	profile := &PreferenceProfile{
		ContentTypes:        []string{"video", "interactive", "text"},
		DifficultyTolerance: learner.Preferences.DifficultyTolerance,
		SessionDuration:     time.Duration(learner.Preferences.SessionDuration) * time.Minute,
		TimeOfDay:           []string{"morning", "afternoon"},
		DevicePreferences:   []string{"desktop", "mobile"},
		LanguagePreferences: []string{"zh-CN", "en"},
		TopicInterests:      make(map[string]float64),
	}

	// 基于历史行为分析主题兴趣
	// 这里可以集成更复杂的兴趣分析算法
	profile.TopicInterests["programming"] = 0.8
	profile.TopicInterests["mathematics"] = 0.6
	profile.TopicInterests["science"] = 0.7

	return profile
}

// analyzeContextualFactors 分析上下文因�?
func (pe *PersonalizationEngine) analyzeContextualFactors(context *PersonalizationContext) *ContextualFactors {
	if context == nil {
		return &ContextualFactors{
			CurrentTime:      time.Now(),
			Device:           "unknown",
			EnergyLevel:      0.7,
			DistractionLevel: 0.3,
		}
	}

	// 解析时间字符�?
	currentTime := time.Now()
	if context.TimeOfDay != "" {
		if parsedTime, err := time.Parse("15:04", context.TimeOfDay); err == nil {
			currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
				parsedTime.Hour(), parsedTime.Minute(), 0, 0, currentTime.Location())
		}
	}

	factors := &ContextualFactors{
		CurrentTime:      currentTime,
		Device:           context.Device,
		Location:         context.Location,
		AvailableTime:    time.Duration(context.AvailableTime) * time.Minute,
		EnergyLevel:      context.EnergyLevel,
		SocialContext:    context.SocialContext,
	}

	// 根据时间推断能量水平和干扰水�?
	hour := currentTime.Hour()
	if hour >= 9 && hour <= 11 {
		factors.EnergyLevel = 0.9 // 上午精力充沛
		factors.DistractionLevel = 0.2
	} else if hour >= 14 && hour <= 16 {
		factors.EnergyLevel = 0.7 // 下午中等
		factors.DistractionLevel = 0.4
	} else {
		factors.EnergyLevel = 0.5 // 其他时间较低
		factors.DistractionLevel = 0.6
	}

	return factors
}

// analyzeMotivationFactors 分析动机因子
func (pe *PersonalizationEngine) analyzeMotivationFactors(ctx context.Context, learner *entities.Learner) *MotivationFactors {
	factors := &MotivationFactors{
		IntrinsicMotivation: 0.7,
		ExtrinsicMotivation: 0.5,
		GoalOrientation:     "mastery",
		AchievementLevel:    0.6,
		ChallengePreference: 0.7,
		FeedbackPreference:  "immediate",
		RewardSensitivity:   0.6,
	}

	// 基于学习者目标和历史表现调整动机因子
	if len(learner.LearningGoals) > 0 {
		factors.IntrinsicMotivation = 0.8
		factors.GoalOrientation = "achievement"
	}

	return factors
}

// generateCandidateRecommendations 生成候选推�?
func (pe *PersonalizationEngine) generateCandidateRecommendations(ctx context.Context, req *PersonalizationRequest, factors *PersonalizationFactors) ([]PersonalizedRecommendation, error) {
	candidates := make([]PersonalizedRecommendation, 0)

	switch req.RecommendationType {
	case "content":
		contentCandidates, err := pe.generateContentCandidates(ctx, req, factors)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, contentCandidates...)

	case "path":
		pathCandidates, err := pe.generatePathCandidates(ctx, req, factors)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, pathCandidates...)

	case "concept":
		conceptCandidates, err := pe.generateConceptCandidates(ctx, req, factors)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, conceptCandidates...)

	default:
		// 生成混合推荐
		contentCandidates, _ := pe.generateContentCandidates(ctx, req, factors)
		pathCandidates, _ := pe.generatePathCandidates(ctx, req, factors)
		conceptCandidates, _ := pe.generateConceptCandidates(ctx, req, factors)
		
		candidates = append(candidates, contentCandidates...)
		candidates = append(candidates, pathCandidates...)
		candidates = append(candidates, conceptCandidates...)
	}

	return candidates, nil
}

// personalizeRecommendations 个性化推荐
func (pe *PersonalizationEngine) personalizeRecommendations(ctx context.Context, candidates []PersonalizedRecommendation, factors *PersonalizationFactors, req *PersonalizationRequest) ([]PersonalizedRecommendation, error) {
	for i := range candidates {
		// 计算个性化分数
		personalizedScore := pe.calculatePersonalizationScore(&candidates[i], factors)
		candidates[i].PersonalizationScore = personalizedScore

		// 调整总分�?
		candidates[i].Score = candidates[i].Score * personalizedScore

		// 生成推理解释
		if req.IncludeExplanations {
			candidates[i].Reasoning = pe.generateReasoning(&candidates[i], factors)
		}

		// 计算置信�?
		candidates[i].Confidence = pe.calculateRecommendationConfidence(&candidates[i], factors)
	}

	// 排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// 限制数量
	if req.MaxRecommendations > 0 && len(candidates) > req.MaxRecommendations {
		candidates = candidates[:req.MaxRecommendations]
	}

	return candidates, nil
}

// calculatePersonalizationScore 计算个性化分数
func (pe *PersonalizationEngine) calculatePersonalizationScore(rec *PersonalizedRecommendation, factors *PersonalizationFactors) float64 {
	score := 1.0

	// 学习风格匹配
	styleScore := pe.calculateStyleMatchScore(rec, &factors.LearningStyle)
	score *= (0.8 + 0.2*styleScore)

	// 技能水平匹�?
	skillScore := pe.calculateSkillMatchScore(rec, &factors.SkillLevel)
	score *= (0.8 + 0.2*skillScore)

	// 偏好匹配
	preferenceScore := pe.calculatePreferenceMatchScore(rec, &factors.Preferences)
	score *= (0.8 + 0.2*preferenceScore)

	// 上下文匹�?
	contextScore := pe.calculateContextMatchScore(rec, &factors.ContextualFactors)
	score *= (0.9 + 0.1*contextScore)

	// 动机匹配
	motivationScore := pe.calculateMotivationMatchScore(rec, &factors.MotivationFactors)
	score *= (0.9 + 0.1*motivationScore)

	return math.Min(score, 2.0) // 限制最大倍数�?
}

// 辅助方法实现
func (pe *PersonalizationEngine) identifyStrengthAreas(skills map[string]float64) []string {
	strengths := make([]string, 0)
	for skill, level := range skills {
		if level > 0.8 {
			strengths = append(strengths, skill)
		}
	}
	return strengths
}

func (pe *PersonalizationEngine) identifyImprovementAreas(skills map[string]float64) []string {
	improvements := make([]string, 0)
	for skill, level := range skills {
		if level < 0.6 {
			improvements = append(improvements, skill)
		}
	}
	return improvements
}

func (pe *PersonalizationEngine) calculateLearningVelocity(ctx context.Context, learnerID uuid.UUID) float64 {
	// 简化实现，实际应该基于历史学习数据计算
	return 0.7
}

func (pe *PersonalizationEngine) getAdaptationHistory(ctx context.Context, learnerID uuid.UUID) []AdaptationRecord {
	// 简化实现，实际应该从数据库获取
	return []AdaptationRecord{}
}

func (pe *PersonalizationEngine) generateContentCandidates(ctx context.Context, req *PersonalizationRequest, factors *PersonalizationFactors) ([]PersonalizedRecommendation, error) {
	// 简化实现，实际应该调用内容推荐算法
	return []PersonalizedRecommendation{}, nil
}

func (pe *PersonalizationEngine) generatePathCandidates(ctx context.Context, req *PersonalizationRequest, factors *PersonalizationFactors) ([]PersonalizedRecommendation, error) {
	// 简化实现，实际应该调用路径推荐算法
	return []PersonalizedRecommendation{}, nil
}

func (pe *PersonalizationEngine) generateConceptCandidates(ctx context.Context, req *PersonalizationRequest, factors *PersonalizationFactors) ([]PersonalizedRecommendation, error) {
	// 简化实现，实际应该调用概念推荐算法
	return []PersonalizedRecommendation{}, nil
}

func (pe *PersonalizationEngine) generateContextualInsights(ctx context.Context, factors *PersonalizationFactors, context *PersonalizationContext) []ContextualInsight {
	insights := make([]ContextualInsight, 0)

	// 基于能量水平的洞�?
	if factors.ContextualFactors.EnergyLevel > 0.8 {
		insights = append(insights, ContextualInsight{
			Type:        "energy",
			Title:       "高能量状�?,
			Description: "当前精力充沛，适合学习挑战性内�?,
			Impact:      0.8,
			Confidence:  0.9,
		})
	}

	return insights
}

func (pe *PersonalizationEngine) generateAdaptationSuggestions(ctx context.Context, factors *PersonalizationFactors, recs []PersonalizedRecommendation) []AdaptationSuggestion {
	suggestions := make([]AdaptationSuggestion, 0)

	// 基于技能差距的建议
	if len(factors.SkillLevel.SkillGaps) > 0 {
		suggestions = append(suggestions, AdaptationSuggestion{
			Type:           "skill_development",
			Title:          "技能提升建�?,
			Description:    "建议优先学习基础技能以提高整体水平",
			Priority:       1,
			ExpectedImpact: 0.8,
			Implementation: []string{"选择基础课程", "增加练习时间", "寻求导师指导"},
		})
	}

	return suggestions
}

func (pe *PersonalizationEngine) calculateOverallConfidence(factors *PersonalizationFactors, recs []PersonalizedRecommendation) float64 {
	if len(recs) == 0 {
		return 0.0
	}

	var totalConfidence float64
	for _, rec := range recs {
		totalConfidence += rec.Confidence
	}

	return totalConfidence / float64(len(recs))
}

func (pe *PersonalizationEngine) calculateStyleMatchScore(rec *PersonalizedRecommendation, style *LearningStyleProfile) float64 {
	// 简化实现，实际应该基于内容类型和学习风格匹�?
	return 0.7
}

func (pe *PersonalizationEngine) calculateSkillMatchScore(rec *PersonalizedRecommendation, skill *SkillLevelProfile) float64 {
	// 简化实现，实际应该基于内容难度和技能水平匹�?
	return 0.8
}

func (pe *PersonalizationEngine) calculatePreferenceMatchScore(rec *PersonalizedRecommendation, pref *PreferenceProfile) float64 {
	// 简化实现，实际应该基于内容特征和用户偏好匹�?
	return 0.7
}

func (pe *PersonalizationEngine) calculateContextMatchScore(rec *PersonalizedRecommendation, context *ContextualFactors) float64 {
	// 简化实现，实际应该基于上下文因素匹�?
	return 0.6
}

func (pe *PersonalizationEngine) calculateMotivationMatchScore(rec *PersonalizedRecommendation, motivation *MotivationFactors) float64 {
	// 简化实现，实际应该基于动机因素匹配
	return 0.7
}

func (pe *PersonalizationEngine) generateReasoning(rec *PersonalizedRecommendation, factors *PersonalizationFactors) []string {
	reasoning := make([]string, 0)

	reasoning = append(reasoning, "基于您的学习风格和偏好推�?)
	reasoning = append(reasoning, "符合您当前的技能水�?)
	reasoning = append(reasoning, "适合您的学习时间安排")

	return reasoning
}

func (pe *PersonalizationEngine) calculateRecommendationConfidence(rec *PersonalizedRecommendation, factors *PersonalizationFactors) float64 {
	// 基于多个因素计算置信�?
	confidence := 0.5

	// 基于个性化分数调整
	confidence += rec.PersonalizationScore * 0.3

	// 基于数据完整性调�?
	if factors.BehaviorPattern.EngagementLevel > 0 {
		confidence += 0.2
	}

	return math.Min(confidence, 1.0)
}
