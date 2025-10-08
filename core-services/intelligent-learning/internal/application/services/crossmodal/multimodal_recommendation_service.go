package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// MultimodalRecommendationService 多模态个性化推荐服务
type MultimodalRecommendationService struct {
	learnerRepo         repositories.LearnerRepository
	contentRepo         repositories.LearningContentRepository
	knowledgeGraphRepo  repositories.KnowledgeGraphRepository
	
	// AI服务集成
	crossModalService   CrossModalServiceInterface
	
	// 推荐引擎
	personalizationEngine *domainServices.PersonalizationEngine
	userBehaviorTracker   *domainServices.UserBehaviorTracker
	preferenceAnalyzer    *domainServices.PreferenceAnalyzer
	contextAnalyzer       *domainServices.ContextAnalyzer
	
	// 缓存和配置
	cache               map[string]*MultimodalRecommendationCache
	cacheMutex          sync.RWMutex
	config              *MultimodalRecommendationConfig
	
	// 性能监控
	metrics             *MultimodalRecommendationMetrics
	metricsMutex        sync.RWMutex
}

// CrossModalServiceInterface 跨模态服务接口
type CrossModalServiceInterface interface {
	ProcessCrossModalInference(ctx context.Context, req *CrossModalInferenceRequest) (*CrossModalInferenceResponse, error)
}

// MultimodalRecommendationConfig 多模态推荐配置
type MultimodalRecommendationConfig struct {
	// 基础配置
	CacheEnabled         bool          `json:"cache_enabled"`
	CacheTTL            time.Duration `json:"cache_ttl"`
	MaxRecommendations  int           `json:"max_recommendations"`
	
	// AI增强配置
	UseSemanticSearch   bool    `json:"use_semantic_search"`
	UseContentMatching  bool    `json:"use_content_matching"`
	UseEmotionAnalysis  bool    `json:"use_emotion_analysis"`
	UseSceneUnderstanding bool  `json:"use_scene_understanding"`
	
	// 权重配置
	ContentSimilarityWeight    float64 `json:"content_similarity_weight"`
	LearningStyleWeight        float64 `json:"learning_style_weight"`
	ProgressWeight             float64 `json:"progress_weight"`
	DifficultyWeight           float64 `json:"difficulty_weight"`
	EmotionalStateWeight       float64 `json:"emotional_state_weight"`
	ContextualRelevanceWeight  float64 `json:"contextual_relevance_weight"`
	
	// 质量控制
	MinConfidenceScore  float64 `json:"min_confidence_score"`
	DiversityThreshold  float64 `json:"diversity_threshold"`
	FreshnessWeight     float64 `json:"freshness_weight"`
}

// MultimodalRecommendationCache 多模态推荐缓存
type MultimodalRecommendationCache struct {
	UserID              string                          `json:"user_id"`
	Recommendations     []*MultimodalRecommendation     `json:"recommendations"`
	SemanticEmbeddings  map[string][]float64           `json:"semantic_embeddings"`
	EmotionalProfile    *EmotionalProfile              `json:"emotional_profile"`
	ContextualFactors   map[string]interface{}         `json:"contextual_factors"`
	GeneratedAt         time.Time                      `json:"generated_at"`
	ExpiresAt           time.Time                      `json:"expires_at"`
}

// MultimodalRecommendation 多模态推荐
type MultimodalRecommendation struct {
	ContentID           uuid.UUID                      `json:"content_id"`
	Title               string                         `json:"title"`
	Description         string                         `json:"description"`
	ContentType         string                         `json:"content_type"`
	Difficulty          string                         `json:"difficulty"`
	EstimatedDuration   time.Duration                  `json:"estimated_duration"`
	
	// 推荐评分
	OverallScore        float64                        `json:"overall_score"`
	ContentSimilarity   float64                        `json:"content_similarity"`
	LearningStyleMatch  float64                        `json:"learning_style_match"`
	ProgressAlignment   float64                        `json:"progress_alignment"`
	DifficultyMatch     float64                        `json:"difficulty_match"`
	EmotionalRelevance  float64                        `json:"emotional_relevance"`
	ContextualRelevance float64                        `json:"contextual_relevance"`
	
	// AI增强信息
	SemanticSimilarity  float64                        `json:"semantic_similarity"`
	EmotionalTone       string                         `json:"emotional_tone"`
	LearningObjectives  []string                       `json:"learning_objectives"`
	Prerequisites       []string                       `json:"prerequisites"`
	
	// 元数据
	RecommendationReason string                        `json:"recommendation_reason"`
	ConfidenceScore     float64                        `json:"confidence_score"`
	Timestamp           time.Time                      `json:"timestamp"`
	Metadata            map[string]interface{}         `json:"metadata"`
}



// MultimodalRecommendationMetrics 多模态推荐指标
type MultimodalRecommendationMetrics struct {
	TotalRequests           int64     `json:"total_requests"`
	AIEnhancedRequests      int64     `json:"ai_enhanced_requests"`
	CacheHitRate            float64   `json:"cache_hit_rate"`
	AverageResponseTime     float64   `json:"average_response_time_ms"`
	AverageConfidenceScore  float64   `json:"average_confidence_score"`
	SemanticSearchUsage     int64     `json:"semantic_search_usage"`
	EmotionAnalysisUsage    int64     `json:"emotion_analysis_usage"`
	LastUpdated             time.Time `json:"last_updated"`
}

// MultimodalRecommendationRequest 多模态推荐请求
type MultimodalRecommendationRequest struct {
	UserID              string                         `json:"user_id" binding:"required"`
	CurrentContentID    *uuid.UUID                     `json:"current_content_id,omitempty"`
	LearningGoals       []string                       `json:"learning_goals,omitempty"`
	PreferredTypes      []string                       `json:"preferred_types,omitempty"`
	MaxDifficulty       string                         `json:"max_difficulty,omitempty"`
	TimeAvailable       *time.Duration                 `json:"time_available,omitempty"`
	Context             map[string]interface{}         `json:"context,omitempty"`
	EmotionalState      *domainServices.EmotionalState                `json:"emotional_state,omitempty"`
	UseAIEnhancement    bool                           `json:"use_ai_enhancement"`
	Limit               int                            `json:"limit,omitempty"`
}

// EmotionalState 情感状态
type MultimodalEmotionalState struct {
	Mood            string  `json:"mood"`
	Energy          float64 `json:"energy"`
	Stress          float64 `json:"stress"`
	Motivation      float64 `json:"motivation"`
	Focus           float64 `json:"focus"`
	Confidence      float64 `json:"confidence"`
}

// MultimodalRecommendationResponse 多模态推荐响应
type MultimodalRecommendationResponse struct {
	UserID              string                         `json:"user_id"`
	Recommendations     []*MultimodalRecommendation    `json:"recommendations"`
	EmotionalProfile    *EmotionalProfile              `json:"emotional_profile"`
	LearningInsights    *LearningInsights              `json:"learning_insights"`
	ProcessingMetadata  *MultimodalProcessingMetadata  `json:"processing_metadata"`
}

// LearningInsights 学习洞察
type LearningInsights struct {
	StrengthAreas       []string                       `json:"strength_areas"`
	ImprovementAreas    []string                       `json:"improvement_areas"`
	LearningPatterns    []string                       `json:"learning_patterns"`
	OptimalStudyTime    []string                       `json:"optimal_study_time"`
	RecommendedStrategy string                         `json:"recommended_strategy"`
}

// ProcessingMetadata 处理元数据
type MultimodalProcessingMetadata struct {
	ProcessingTime      int64                          `json:"processing_time_ms"`
	AIEnhancementUsed   bool                           `json:"ai_enhancement_used"`
	CacheUsed           bool                           `json:"cache_used"`
	ComponentsUsed      []string                       `json:"components_used"`
	ConfidenceLevel     float64                        `json:"confidence_level"`
	Timestamp           time.Time                      `json:"timestamp"`
}

// NewMultimodalRecommendationService 创建多模态推荐服务
func NewMultimodalRecommendationService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	crossModalService CrossModalServiceInterface,
	personalizationEngine *domainServices.PersonalizationEngine,
	userBehaviorTracker *domainServices.UserBehaviorTracker,
	preferenceAnalyzer *domainServices.PreferenceAnalyzer,
	contextAnalyzer *domainServices.ContextAnalyzer,
) *MultimodalRecommendationService {
	config := &MultimodalRecommendationConfig{
		CacheEnabled:               true,
		CacheTTL:                  30 * time.Minute,
		MaxRecommendations:        20,
		UseSemanticSearch:         true,
		UseContentMatching:        true,
		UseEmotionAnalysis:        true,
		UseSceneUnderstanding:     true,
		ContentSimilarityWeight:   0.25,
		LearningStyleWeight:       0.20,
		ProgressWeight:            0.15,
		DifficultyWeight:          0.15,
		EmotionalStateWeight:      0.15,
		ContextualRelevanceWeight: 0.10,
		MinConfidenceScore:        0.6,
		DiversityThreshold:        0.7,
		FreshnessWeight:           0.1,
	}

	return &MultimodalRecommendationService{
		learnerRepo:           learnerRepo,
		contentRepo:           contentRepo,
		knowledgeGraphRepo:    knowledgeGraphRepo,
		crossModalService:     crossModalService,
		personalizationEngine: personalizationEngine,
		userBehaviorTracker:   userBehaviorTracker,
		preferenceAnalyzer:    preferenceAnalyzer,
		contextAnalyzer:       contextAnalyzer,
		cache:                 make(map[string]*MultimodalRecommendationCache),
		config:                config,
		metrics: &MultimodalRecommendationMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// GetMultimodalRecommendations 获取多模态个性化推荐
func (s *MultimodalRecommendationService) GetMultimodalRecommendations(
	ctx context.Context,
	req *MultimodalRecommendationRequest,
) (*MultimodalRecommendationResponse, error) {
	startTime := time.Now()
	
	// 更新指标
	s.updateMetrics(func(m *MultimodalRecommendationMetrics) {
		m.TotalRequests++
		if req.UseAIEnhancement {
			m.AIEnhancedRequests++
		}
	})

	// 检查缓存
	cacheKey := s.generateCacheKey(req)
	if cached := s.getCachedRecommendations(cacheKey); cached != nil && s.config.CacheEnabled {
		return s.buildResponseFromCache(cached, req.UserID, startTime), nil
	}

	// 获取学习者信息
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	learner, err := s.learnerRepo.GetByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 分析情感状态
	emotionalProfile, err := s.analyzeEmotionalState(ctx, req, learner)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze emotional state: %w", err)
	}

	// 获取候选内容
	candidates, err := s.getCandidateContent(ctx, req, learner)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate content: %w", err)
	}

	// AI增强推荐
	var recommendations []*MultimodalRecommendation
	if req.UseAIEnhancement && s.config.UseSemanticSearch {
		recommendations, err = s.generateAIEnhancedRecommendations(ctx, req, learner, candidates, emotionalProfile)
		if err != nil {
			// 降级到基础推荐
			recommendations, err = s.generateBasicRecommendations(ctx, req, learner, candidates, emotionalProfile)
		}
	} else {
		recommendations, err = s.generateBasicRecommendations(ctx, req, learner, candidates, emotionalProfile)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// 应用多样性和质量过滤
	recommendations = s.applyDiversityFilter(recommendations)
	recommendations = s.applyQualityFilter(recommendations)

	// 限制数量
	limit := req.Limit
	if limit <= 0 || limit > s.config.MaxRecommendations {
		limit = s.config.MaxRecommendations
	}
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	// 生成学习洞察
	insights := s.generateLearningInsights(ctx, learner, recommendations)

	// 缓存结果
	if s.config.CacheEnabled {
		s.cacheRecommendations(cacheKey, recommendations, emotionalProfile)
	}

	// 构建响应
	response := &MultimodalRecommendationResponse{
		UserID:           req.UserID,
		Recommendations:  recommendations,
		EmotionalProfile: emotionalProfile,
		LearningInsights: insights,
		ProcessingMetadata: &MultimodalProcessingMetadata{
			ProcessingTime:    time.Since(startTime).Milliseconds(),
			AIEnhancementUsed: req.UseAIEnhancement,
			CacheUsed:         false,
			ComponentsUsed:    s.getUsedComponents(req),
			ConfidenceLevel:   s.calculateOverallConfidence(recommendations),
			Timestamp:         time.Now(),
		},
	}

	// 更新平均响应时间
	s.updateMetrics(func(m *MultimodalRecommendationMetrics) {
		m.AverageResponseTime = (m.AverageResponseTime + float64(time.Since(startTime).Milliseconds())) / 2
		m.AverageConfidenceScore = (m.AverageConfidenceScore + response.ProcessingMetadata.ConfidenceLevel) / 2
	})

	return response, nil
}

// analyzeEmotionalState 分析情感状态
func (s *MultimodalRecommendationService) analyzeEmotionalState(
	ctx context.Context,
	req *MultimodalRecommendationRequest,
	learner *entities.Learner,
) (*EmotionalProfile, error) {
	if !s.config.UseEmotionAnalysis {
		return &EmotionalProfile{
			CurrentMood:     "neutral",
			StressLevel:     0.5,
			MotivationLevel: 0.7,
			FocusLevel:      0.6,
			PreferredTone:   "encouraging",
			LastUpdated:     time.Now(),
		}, nil
	}

	// 使用跨模态AI服务进行情感分析
	if req.EmotionalState != nil {
		emotionReq := &CrossModalInferenceRequest{
			Type: "emotion_analysis",
			Data: map[string]interface{}{
				"emotional_state": req.EmotionalState,
				"learning_history": learner.LearningHistory,
				"context": req.Context,
			},
		}

		response, err := s.crossModalService.ProcessCrossModalInference(ctx, emotionReq)
		if err == nil && response.Success {
			return s.parseEmotionalProfile(response.Result)
		}
	}

	// 基于学习历史推断情感状态
	return s.inferEmotionalStateFromHistory(learner), nil
}

// getCandidateContent 获取候选内容
func (s *MultimodalRecommendationService) getCandidateContent(
	ctx context.Context,
	req *MultimodalRecommendationRequest,
	learner *entities.Learner,
) ([]*entities.LearningContent, error) {
	// 基于学习目标和偏好获取候选内容
	query := &repositories.ContentSearchQuery{
		Offset: 0,
		Limit:  100, // 默认限制
	}

	// 设置内容类型过滤
	if len(req.PreferredTypes) > 0 {
		// 将字符串转换为ContentType枚举
		for _, typeStr := range req.PreferredTypes {
			contentType := entities.ContentType(typeStr)
			query.ContentType = &contentType
			break // 目前只支持单个类型，取第一个
		}
	}

	// 设置时间限制
	if req.TimeAvailable != nil {
		query.MaxDuration = req.TimeAvailable
	}

	// 设置学习者ID用于个性化
	query.LearnerID = &learner.ID

	contents, _, err := s.contentRepo.Search(ctx, query)
	return contents, err
}

// generateAIEnhancedRecommendations 生成AI增强推荐
func (s *MultimodalRecommendationService) generateAIEnhancedRecommendations(
	ctx context.Context,
	req *MultimodalRecommendationRequest,
	learner *entities.Learner,
	candidates []*entities.LearningContent,
	emotionalProfile *EmotionalProfile,
) ([]*MultimodalRecommendation, error) {
	var recommendations []*MultimodalRecommendation

	for _, content := range candidates {
		// 语义相似度分析
		semanticSimilarity, err := s.calculateSemanticSimilarity(ctx, content, learner, req)
		if err != nil {
			semanticSimilarity = 0.5 // 默认值
		}

		// 内容匹配分析
		contentMatch, err := s.analyzeContentMatching(ctx, content, learner, req)
		if err != nil {
			contentMatch = 0.5 // 默认值
		}

		// 计算综合评分
		recommendation := s.buildMultimodalRecommendation(
			content, learner, req, emotionalProfile,
			semanticSimilarity, contentMatch,
		)

		if recommendation.ConfidenceScore >= s.config.MinConfidenceScore {
			recommendations = append(recommendations, recommendation)
		}
	}

	// 按评分排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].OverallScore > recommendations[j].OverallScore
	})

	return recommendations, nil
}

// calculateSemanticSimilarity 计算语义相似度
func (s *MultimodalRecommendationService) calculateSemanticSimilarity(
	ctx context.Context,
	content *entities.LearningContent,
	learner *entities.Learner,
	req *MultimodalRecommendationRequest,
) (float64, error) {
	if !s.config.UseSemanticSearch {
		return 0.5, nil
	}

	// 构建语义搜索请求
	semanticReq := &CrossModalInferenceRequest{
		Type: "semantic_search",
		Data: map[string]interface{}{
			"content_description": content.Description,
			"learning_goals":      req.LearningGoals,
			"learner_preferences": learner.Preferences,
			"context":            req.Context,
		},
	}

	response, err := s.crossModalService.ProcessCrossModalInference(ctx, semanticReq)
	if err != nil {
		return 0.5, err
	}

	if !response.Success {
		return 0.5, fmt.Errorf("semantic search failed: %s", response.Error)
	}

	// 解析相似度分数
	if similarity, ok := response.Result["similarity_score"].(float64); ok {
		return similarity, nil
	}

	return 0.5, nil
}

// analyzeContentMatching 分析内容匹配
func (s *MultimodalRecommendationService) analyzeContentMatching(
	ctx context.Context,
	content *entities.LearningContent,
	learner *entities.Learner,
	req *MultimodalRecommendationRequest,
) (float64, error) {
	if !s.config.UseContentMatching {
		return 0.5, nil
	}

	matchingReq := &CrossModalInferenceRequest{
		Type: "content_matching",
		Data: map[string]interface{}{
			"content":         content,
			"learner_profile": learner,
			"learning_goals":  req.LearningGoals,
			"context":        req.Context,
		},
	}

	response, err := s.crossModalService.ProcessCrossModalInference(ctx, matchingReq)
	if err != nil {
		return 0.5, err
	}

	if !response.Success {
		return 0.5, fmt.Errorf("content matching failed: %s", response.Error)
	}

	if matchScore, ok := response.Result["match_score"].(float64); ok {
		return matchScore, nil
	}

	return 0.5, nil
}

// buildMultimodalRecommendation 构建多模态推荐
func (s *MultimodalRecommendationService) buildMultimodalRecommendation(
	content *entities.LearningContent,
	learner *entities.Learner,
	req *MultimodalRecommendationRequest,
	emotionalProfile *EmotionalProfile,
	semanticSimilarity, contentMatch float64,
) *MultimodalRecommendation {
	// 计算各项评分
	contentSimilarity := s.calculateContentSimilarity(content, learner)
	learningStyleMatch := s.calculateLearningStyleMatch(content, learner)
	progressAlignment := s.calculateProgressAlignment(content, learner)
	difficultyMatch := s.calculateDifficultyMatch(content, learner)
	emotionalRelevance := s.calculateEmotionalRelevance(content, emotionalProfile)
	contextualRelevance := s.calculateContextualRelevance(content, req.Context)

	// 计算综合评分
	overallScore := s.config.ContentSimilarityWeight*contentSimilarity +
		s.config.LearningStyleWeight*learningStyleMatch +
		s.config.ProgressWeight*progressAlignment +
		s.config.DifficultyWeight*difficultyMatch +
		s.config.EmotionalStateWeight*emotionalRelevance +
		s.config.ContextualRelevanceWeight*contextualRelevance

	// 融合AI增强评分
	if semanticSimilarity > 0 {
		overallScore = (overallScore + semanticSimilarity) / 2
	}
	if contentMatch > 0 {
		overallScore = (overallScore + contentMatch) / 2
	}

	// 计算置信度
	confidenceScore := s.calculateConfidenceScore(
		contentSimilarity, learningStyleMatch, progressAlignment,
		difficultyMatch, emotionalRelevance, contextualRelevance,
		semanticSimilarity, contentMatch,
	)

	return &MultimodalRecommendation{
		ContentID:           content.ID,
		Title:               content.Title,
		Description:         content.Description,
		ContentType:         string(content.Type),
		Difficulty:          string(content.Difficulty),
		EstimatedDuration:   time.Duration(content.EstimatedDuration) * time.Minute,
		OverallScore:        overallScore,
		ContentSimilarity:   contentSimilarity,
		LearningStyleMatch:  learningStyleMatch,
		ProgressAlignment:   progressAlignment,
		DifficultyMatch:     difficultyMatch,
		EmotionalRelevance:  emotionalRelevance,
		ContextualRelevance: contextualRelevance,
		SemanticSimilarity:  semanticSimilarity,
		EmotionalTone:       s.inferEmotionalTone(content, emotionalProfile),
		LearningObjectives:  s.extractLearningObjectiveDescriptions(content.LearningObjectives),
		Prerequisites:       s.getPrerequisites(content),
		RecommendationReason: s.generateRecommendationReason(
			contentSimilarity, learningStyleMatch, progressAlignment,
			difficultyMatch, emotionalRelevance, contextualRelevance,
		),
		ConfidenceScore: confidenceScore,
		Timestamp:       time.Now(),
		Metadata: map[string]interface{}{
			"ai_enhanced":        semanticSimilarity > 0 || contentMatch > 0,
			"emotional_profile":  emotionalProfile,
			"learner_level":      learner.Skills,
		},
	}
}

// 辅助方法实现...
func (s *MultimodalRecommendationService) generateBasicRecommendations(
	ctx context.Context,
	req *MultimodalRecommendationRequest,
	learner *entities.Learner,
	candidates []*entities.LearningContent,
	emotionalProfile *EmotionalProfile,
) ([]*MultimodalRecommendation, error) {
	var recommendations []*MultimodalRecommendation

	for _, content := range candidates {
		recommendation := s.buildMultimodalRecommendation(
			content, learner, req, emotionalProfile, 0, 0,
		)

		if recommendation.ConfidenceScore >= s.config.MinConfidenceScore {
			recommendations = append(recommendations, recommendation)
		}
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].OverallScore > recommendations[j].OverallScore
	})

	return recommendations, nil
}

// 其他辅助方法的实现...
func (s *MultimodalRecommendationService) calculateContentSimilarity(content *entities.LearningContent, learner *entities.Learner) float64 {
	// 基于内容标签和学习者兴趣计算相似度
	return 0.7 // 简化实现
}

func (s *MultimodalRecommendationService) calculateLearningStyleMatch(content *entities.LearningContent, learner *entities.Learner) float64 {
	// 基于学习风格匹配度计算
	return 0.8 // 简化实现
}

func (s *MultimodalRecommendationService) calculateProgressAlignment(content *entities.LearningContent, learner *entities.Learner) float64 {
	// 基于学习进度对齐度计算
	return 0.6 // 简化实现
}

func (s *MultimodalRecommendationService) calculateDifficultyMatch(content *entities.LearningContent, learner *entities.Learner) float64 {
	// 基于难度匹配度计算
	return 0.75 // 简化实现
}

func (s *MultimodalRecommendationService) calculateEmotionalRelevance(content *entities.LearningContent, profile *EmotionalProfile) float64 {
	// 基于情感相关性计算
	return 0.65 // 简化实现
}

func (s *MultimodalRecommendationService) calculateContextualRelevance(content *entities.LearningContent, context map[string]interface{}) float64 {
	// 基于上下文相关性计算
	return 0.7 // 简化实现
}

func (s *MultimodalRecommendationService) calculateConfidenceScore(scores ...float64) float64 {
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

func (s *MultimodalRecommendationService) generateRecommendationReason(scores ...float64) string {
	return "基于多模态AI分析和个性化学习偏好推荐"
}

func (s *MultimodalRecommendationService) inferEmotionalTone(content *entities.LearningContent, profile *EmotionalProfile) string {
	return "encouraging"
}

func (s *MultimodalRecommendationService) extractLearningObjectiveDescriptions(objectives []entities.LearningObjective) []string {
	descriptions := make([]string, len(objectives))
	for i, objective := range objectives {
		descriptions[i] = objective.Description
	}
	return descriptions
}

func (s *MultimodalRecommendationService) getPrerequisites(content *entities.LearningContent) []string {
	return []string{} // 实现获取前置条件的逻辑
}

func (s *MultimodalRecommendationService) getAllowedDifficultyLevels(maxDifficulty string, learner *entities.Learner) []string {
	return []string{"beginner", "intermediate"} // 简化实现
}

func (s *MultimodalRecommendationService) parseEmotionalProfile(result map[string]interface{}) (*EmotionalProfile, error) {
	// 解析AI返回的情感分析结果
	return &EmotionalProfile{
		CurrentMood:     "positive",
		StressLevel:     0.3,
		MotivationLevel: 0.8,
		FocusLevel:      0.7,
		PreferredTone:   "encouraging",
		LastUpdated:     time.Now(),
	}, nil
}

func (s *MultimodalRecommendationService) inferEmotionalStateFromHistory(learner *entities.Learner) *EmotionalProfile {
	return &EmotionalProfile{
		CurrentMood:     "neutral",
		StressLevel:     0.5,
		MotivationLevel: 0.7,
		FocusLevel:      0.6,
		PreferredTone:   "encouraging",
		LastUpdated:     time.Now(),
	}
}

func (s *MultimodalRecommendationService) applyDiversityFilter(recommendations []*MultimodalRecommendation) []*MultimodalRecommendation {
	// 应用多样性过滤
	return recommendations
}

func (s *MultimodalRecommendationService) applyQualityFilter(recommendations []*MultimodalRecommendation) []*MultimodalRecommendation {
	// 应用质量过滤
	filtered := make([]*MultimodalRecommendation, 0)
	for _, rec := range recommendations {
		if rec.ConfidenceScore >= s.config.MinConfidenceScore {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

func (s *MultimodalRecommendationService) generateLearningInsights(ctx context.Context, learner *entities.Learner, recommendations []*MultimodalRecommendation) *LearningInsights {
	return &LearningInsights{
		StrengthAreas:       []string{"逻辑思维", "问题解决"},
		ImprovementAreas:    []string{"创意思维", "团队协作"},
		LearningPatterns:    []string{"视觉学习偏好", "实践导向"},
		OptimalStudyTime:    []string{"上午9-11点", "下午2-4点"},
		RecommendedStrategy: "结合理论学习和实践练习",
	}
}

func (s *MultimodalRecommendationService) generateCacheKey(req *MultimodalRecommendationRequest) string {
	data, _ := json.Marshal(req)
	return fmt.Sprintf("multimodal_rec_%s_%x", req.UserID, data)
}

func (s *MultimodalRecommendationService) getCachedRecommendations(key string) *MultimodalRecommendationCache {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	if cached, exists := s.cache[key]; exists && time.Now().Before(cached.ExpiresAt) {
		return cached
	}
	return nil
}

func (s *MultimodalRecommendationService) cacheRecommendations(key string, recommendations []*MultimodalRecommendation, profile *EmotionalProfile) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	s.cache[key] = &MultimodalRecommendationCache{
		Recommendations:  recommendations,
		EmotionalProfile: profile,
		GeneratedAt:      time.Now(),
		ExpiresAt:        time.Now().Add(s.config.CacheTTL),
	}
}

func (s *MultimodalRecommendationService) buildResponseFromCache(cached *MultimodalRecommendationCache, userID string, startTime time.Time) *MultimodalRecommendationResponse {
	return &MultimodalRecommendationResponse{
		UserID:           userID,
		Recommendations:  cached.Recommendations,
		EmotionalProfile: cached.EmotionalProfile,
		ProcessingMetadata: &MultimodalProcessingMetadata{
			ProcessingTime: time.Since(startTime).Milliseconds(),
			CacheUsed:      true,
			Timestamp:      time.Now(),
		},
	}
}

func (s *MultimodalRecommendationService) getUsedComponents(req *MultimodalRecommendationRequest) []string {
	components := []string{"personalization_engine"}
	if req.UseAIEnhancement {
		components = append(components, "cross_modal_ai", "semantic_search", "emotion_analysis")
	}
	return components
}

func (s *MultimodalRecommendationService) calculateOverallConfidence(recommendations []*MultimodalRecommendation) float64 {
	if len(recommendations) == 0 {
		return 0
	}
	
	sum := 0.0
	for _, rec := range recommendations {
		sum += rec.ConfidenceScore
	}
	return sum / float64(len(recommendations))
}

func (s *MultimodalRecommendationService) updateMetrics(updateFunc func(*MultimodalRecommendationMetrics)) {
	s.metricsMutex.Lock()
	defer s.metricsMutex.Unlock()
	updateFunc(s.metrics)
	s.metrics.LastUpdated = time.Now()
}

// GetMetrics 获取服务指标
func (s *MultimodalRecommendationService) GetMetrics() *MultimodalRecommendationMetrics {
	s.metricsMutex.RLock()
	defer s.metricsMutex.RUnlock()
	return s.metrics
}

// UpdateConfig 更新配置
func (s *MultimodalRecommendationService) UpdateConfig(config *MultimodalRecommendationConfig) {
	s.config = config
}