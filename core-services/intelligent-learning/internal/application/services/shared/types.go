package shared

import (
	"time"

	"github.com/google/uuid"
)

// RecommendationSettings 推荐设置
type RecommendationSettings struct {
	RecommendationTypes      []RecommendationType              `json:"recommendation_types"`
	RecommendationAlgorithms []AdaptiveRecommendationAlgorithm `json:"recommendation_algorithms"`
	RecommendationFilters    []*RecommendationFilter           `json:"recommendation_filters"`
	RecommendationRanking    *RecommendationRankingConfig      `json:"recommendation_ranking"`
	RecommendationDiversity  *RecommendationDiversityConfig    `json:"recommendation_diversity"`
	MaxRecommendations       int                               `json:"max_recommendations"`
	MinConfidence            float64                           `json:"min_confidence"`
	MinConfidenceScore       float64                           `json:"min_confidence_score"`
	DiversityWeight          float64                           `json:"diversity_weight"`
	NoveltyWeight            float64                           `json:"novelty_weight"`
	RelevanceWeight          float64                           `json:"relevance_weight"`
	PopularityWeight         float64                           `json:"popularity_weight"`
	EnabledStrategies        []RecommendationStrategy          `json:"enabled_strategies"`
	RefreshInterval          time.Duration                     `json:"refresh_interval"`
	Metadata                 map[string]interface{}            `json:"metadata"`
}

// OptimizationSettings 优化设置
type OptimizationSettings struct {
	OptimizationMethod string                 `json:"optimization_method"`
	Parameters         map[string]interface{} `json:"parameters"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// AdaptiveCachedLearnerProfile 自适应缓存的学习者档�?
type AdaptiveCachedLearnerProfile struct {
	LearnerID     uuid.UUID              `json:"learner_id"`
	Profile       *LearnerProfile        `json:"profile"`
	Timestamp     time.Time              `json:"timestamp"`
	ExpiresAt     time.Time              `json:"expires_at"`
	AccessCount   int                    `json:"access_count"`
	LastAccessed  time.Time              `json:"last_accessed"`
	LastUpdated   time.Time              `json:"last_updated"`
	CacheExpiry   time.Time              `json:"cache_expiry"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CachedLearningStrategy 缓存的学习策�?
type CachedLearningStrategy struct {
	StrategyID   string                 `json:"strategy_id"`
	Strategy     *LearningStrategy      `json:"strategy"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedAdaptationResult 缓存的适配结果
type CachedAdaptationResult struct {
	AdaptationID  string                 `json:"adaptation_id"`
	RequestID     uuid.UUID              `json:"request_id"`
	Result        *AdaptationResponse    `json:"result"`
	Timestamp     time.Time              `json:"timestamp"`
	ExpiresAt     time.Time              `json:"expires_at"`
	AccessCount   int                    `json:"access_count"`
	LastAccessed  time.Time              `json:"last_accessed"`
	LastUpdated   time.Time              `json:"last_updated"`
	CacheExpiry   time.Time              `json:"cache_expiry"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CachedPersonalizationData 缓存的个性化数据
type CachedPersonalizationData struct {
	LearnerID    uuid.UUID              `json:"learner_id"`
	Data         *PersonalizationData   `json:"data"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedLearningPath 缓存的学习路�?
type CachedLearningPath struct {
	PathID       uuid.UUID              `json:"path_id"`
	Path         *LearningPath          `json:"path"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedAssessmentResult 缓存的评估结�?
type CachedAssessmentResult struct {
	AssessmentID uuid.UUID              `json:"assessment_id"`
	Result       *AssessmentResult      `json:"result"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedRecommendationResult 缓存的推荐结�?
type CachedRecommendationResult struct {
	RequestID    uuid.UUID              `json:"request_id"`
	Result       *RecommendationResult  `json:"result"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PersonalizationRule 个性化规则
type PersonalizationRule struct {
	RuleID      string                 `json:"rule_id"`
	RuleType    string                 `json:"rule_type"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     map[string]interface{} `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// 支持类型定义
type RecommendationType string
type AdaptiveRecommendationAlgorithm string
type RecommendationStrategy string

type RecommendationFilter struct {
	FilterType string                 `json:"filter_type"`
	Criteria   map[string]interface{} `json:"criteria"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type RecommendationRankingConfig struct {
	RankingMethod string                 `json:"ranking_method"`
	Weights       map[string]float64     `json:"weights"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type RecommendationDiversityConfig struct {
	DiversityMethod string                 `json:"diversity_method"`
	DiversityWeight float64                `json:"diversity_weight"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type LearnerProfile struct {
	ProfileID   uuid.UUID              `json:"profile_id"`
	LearnerID   uuid.UUID              `json:"learner_id"`
	ProfileData map[string]interface{} `json:"profile_data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type LearningStrategy struct {
	StrategyID   string                 `json:"strategy_id"`
	StrategyType string                 `json:"strategy_type"`
	Parameters   map[string]interface{} `json:"parameters"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type AdaptationResponse struct {
	ResponseID string                 `json:"response_id"`
	Data       map[string]interface{} `json:"data"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type PersonalizationData struct {
	DataID   string                 `json:"data_id"`
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

type LearningPath struct {
	PathID   uuid.UUID              `json:"path_id"`
	PathData map[string]interface{} `json:"path_data"`
	Metadata map[string]interface{} `json:"metadata"`
}

type AssessmentResult struct {
	ResultID uuid.UUID              `json:"result_id"`
	Data     map[string]interface{} `json:"data"`
}

type RecommendationResult struct {
	ResultID uuid.UUID              `json:"result_id"`
	Data     map[string]interface{} `json:"data"`
}
