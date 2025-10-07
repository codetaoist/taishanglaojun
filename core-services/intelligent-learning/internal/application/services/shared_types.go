package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// 共享服务接口定义

// LearningAnalyticsService 学习分析服务接口
type LearningAnalyticsService interface {
	GenerateAnalyticsReport(ctx context.Context, req *AnalyticsRequest) (*LearningAnalyticsReport, error)
}

// LearningPathService 学习路径服务接口
type LearningPathService interface {
	RecommendPersonalizedPaths(ctx context.Context, req *PathRecommendationRequest) ([]*PersonalizedPath, error)
}

// KnowledgeGraphService 知识图谱服务接口
type KnowledgeGraphService interface {
	RecommendConcepts(ctx context.Context, req *ConceptRecommendationRequest) ([]*ConceptRecommendation, error)
	AnalyzeGraph(ctx context.Context, req interface{}) (interface{}, error)
}

// 共享数据类型定义

// AnalyticsRequest 分析请求
type AnalyticsRequest struct {
	LearnerID         uuid.UUID         `json:"learner_id"`
	TimeRange         AnalyticsTimeRange `json:"time_range"`
	AnalysisType      string            `json:"analysis_type"`
	Granularity       string            `json:"granularity"`
	IncludeComparison bool              `json:"include_comparison"`
	ComparisonGroup   string            `json:"comparison_group"`
}

// AnalyticsTimeRange 分析时间范围
type AnalyticsTimeRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// LearningAnalyticsReport 学习分析报告
type LearningAnalyticsReport struct {
	LearnerID   uuid.UUID `json:"learner_id"`
	TimeRange   AnalyticsTimeRange `json:"time_range"`
	GeneratedAt time.Time `json:"generated_at"`
}

// PersonalizedPath 个性化路径
type PersonalizedPath struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Difficulty  string    `json:"difficulty"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// ConceptRecommendation 概念推荐
type ConceptRecommendation struct {
	ConceptID   uuid.UUID `json:"concept_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Relevance   float64   `json:"relevance"`
	Reason      string    `json:"reason"`
}

// PathRecommendationRequest 路径推荐请求
type PathRecommendationRequest struct {
	LearnerID    uuid.UUID `json:"learner_id"`
	GraphID      uuid.UUID `json:"graph_id"`
	TargetNodeID uuid.UUID `json:"target_node_id"`
	MaxPaths     int       `json:"max_paths"`
}

// ConceptRecommendationRequest 概念推荐请求
type ConceptRecommendationRequest struct {
	GraphID            uuid.UUID `json:"graph_id"`
	LearnerID          uuid.UUID `json:"learner_id"`
	TargetSkills       []string  `json:"target_skills"`
	MaxRecommendations int       `json:"max_recommendations"`
	IncludeReasoning   bool      `json:"include_reasoning"`
}

// PathPreferences 路径偏好
type PathPreferences struct {
	DifficultyLevel string        `json:"difficulty_level"`
	LearningStyle   string        `json:"learning_style"`
	TimeConstraint  time.Duration `json:"time_constraint"`
}

// PathRecommendationResponse 路径推荐响应
type PathRecommendationResponse struct {
	Paths       []*PersonalizedPath `json:"paths"`
	GeneratedAt time.Time          `json:"generated_at"`
}

// GraphAnalysisResult 图谱分析结果
type GraphAnalysisResult struct {
	GraphID     uuid.UUID `json:"graph_id"`
	Metrics     map[string]interface{} `json:"metrics"`
	GeneratedAt time.Time `json:"generated_at"`
}

// LearningGap 学习差距
type LearningGap struct {
	ConceptID   uuid.UUID `json:"concept_id"`
	GapLevel    float64   `json:"gap_level"`
	Description string    `json:"description"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
}

// GraphAnalysisRequest 图谱分析请求
type GraphAnalysisRequest struct {
	GraphID     uuid.UUID `json:"graph_id"`
	AnalysisType string   `json:"analysis_type"`
}

// LearningPreferences 学习偏好
type LearningPreferences struct {
	LearningStyle    string            `json:"learning_style"`
	DifficultyLevel  string            `json:"difficulty_level"`
	PreferredTopics  []string          `json:"preferred_topics"`
	StudySchedule    map[string]string `json:"study_schedule"`
	NotificationSettings map[string]bool `json:"notification_settings"`
}

// DataAggregation 数据聚合
type DataAggregation struct {
	Type        string                 `json:"type"`
	Field       string                 `json:"field"`
	Value       interface{}            `json:"value"`
	Count       int                    `json:"count"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// StatisticalSummary 统计摘要
type StatisticalSummary struct {
	Mean           float64                    `json:"mean"`
	Median         float64                    `json:"median"`
	StandardDev    float64                    `json:"standard_deviation"`
	Variance       float64                    `json:"variance"`
	Min            float64                    `json:"min"`
	Max            float64                    `json:"max"`
	Percentiles    map[string]float64         `json:"percentiles"`
	SampleSize     int                        `json:"sample_size"`
}

// VisualizationConfig 可视化配置
type VisualizationConfig struct {
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	XAxis      string                 `json:"x_axis"`
	YAxis      string                 `json:"y_axis"`
	Colors     []string               `json:"colors"`
	Options    map[string]interface{} `json:"options"`
}

// PriorityLevel 优先级级别
type PriorityLevel string

const (
	PriorityLow    PriorityLevel = "low"
	PriorityMedium PriorityLevel = "medium"
	PriorityHigh   PriorityLevel = "high"
	PriorityCritical PriorityLevel = "critical"
)

// ProcessedAnalyticsData 处理后的分析数据
type ProcessedAnalyticsData struct {
	SourceData    interface{}                `json:"source_data"`
	ProcessedData map[string]interface{}     `json:"processed_data"`
	Aggregations  map[string]*DataAggregation `json:"aggregations"`
	Statistics    map[string]*StatisticalSummary `json:"statistics"`
	QualityScore  float64                    `json:"quality_score"`
	ProcessedAt   time.Time                  `json:"processed_at"`
}

// ImplementationPlan 实施计划
type ImplementationPlan struct {
	Steps       []string  `json:"steps"`
	Timeline    string    `json:"timeline"`
	Resources   []string  `json:"resources"`
	Milestones  []string  `json:"milestones"`
}

// TrendDirection 趋势方向
type TrendDirection string

const (
	TrendUp    TrendDirection = "up"
	TrendDown  TrendDirection = "down"
	TrendFlat  TrendDirection = "flat"
)

// ConfidenceInterval 置信区间
type ConfidenceInterval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"`
}

// RecommendationCategory 推荐类别
type RecommendationCategory string

const (
	CategoryLearning     RecommendationCategory = "learning"
	CategoryContent      RecommendationCategory = "content"
	CategoryPath         RecommendationCategory = "path"
	CategoryPerformance  RecommendationCategory = "performance"
)

// UserSession 用户会话
type UserSession struct {
	SessionID   uuid.UUID `json:"session_id"`
	UserID      uuid.UUID `json:"user_id"`
	StartTime   time.Time `json:"start_time"`
	LastActive  time.Time `json:"last_active"`
	IsActive    bool      `json:"is_active"`
	DeviceInfo  string    `json:"device_info"`
	IPAddress   string    `json:"ip_address"`
}

// Trend 趋势
type Trend struct {
	Direction   TrendDirection `json:"direction"`
	Strength    float64        `json:"strength"`
	Confidence  float64        `json:"confidence"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Description string         `json:"description"`
}


// RecommendationSettings 推荐设置
type RecommendationSettings struct {
	RecommendationTypes      []RecommendationType              `json:"recommendation_types"`
	RecommendationAlgorithms []AdaptiveRecommendationAlgorithm `json:"recommendation_algorithms"`
	RecommendationFilters    []*RecommendationFilter           `json:"recommendation_filters"`
	RecommendationRanking    *RecommendationRankingConfig      `json:"recommendation_ranking"`
	RecommendationDiversity  *RecommendationDiversityConfig    `json:"recommendation_diversity"`
	MaxRecommendations       int                               `json:"max_recommendations"`
	MinConfidence            float64                           `json:"min_confidence"`
	Metadata                 map[string]interface{}            `json:"metadata"`
}

// RecommendationType 推荐类型
type RecommendationType string

const (
	RecommendationTypeContent  RecommendationType = "content"
	RecommendationTypeStrategy RecommendationType = "strategy"
	RecommendationTypePath     RecommendationType = "path"
	RecommendationTypeResource RecommendationType = "resource"
	RecommendationTypePeer     RecommendationType = "peer"
	RecommendationTypeActivity RecommendationType = "activity"
)

// AdaptiveCachedLearnerProfile 自适应缓存的学习者档案
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

// CachedLearningStrategy 缓存的学习策略
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

// CachedLearningPath 缓存的学习路径
type CachedLearningPath struct {
	PathID       uuid.UUID              `json:"path_id"`
	Path         *LearningPath          `json:"path"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedAssessmentResult 缓存的评估结果
type CachedAssessmentResult struct {
	AssessmentID uuid.UUID              `json:"assessment_id"`
	Result       *AssessmentResult      `json:"result"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CachedRecommendationResult 缓存的推荐结果
type CachedRecommendationResult struct {
	RequestID    uuid.UUID              `json:"request_id"`
	Result       *RecommendationResult  `json:"result"`
	LastUpdated  time.Time              `json:"last_updated"`
	CacheExpiry  time.Time              `json:"cache_expiry"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	OverallScore          float64                `json:"overall_score"`
	OverallQuality        float64                `json:"overall_quality"`
	ContentQuality        float64                `json:"content_quality"`
	DeliveryQuality       float64                `json:"delivery_quality"`
	EngagementQuality     float64                `json:"engagement_quality"`
	LearningEffectiveness float64                `json:"learning_effectiveness"`
	Confidence            float64                `json:"confidence"`
	Metadata              map[string]interface{} `json:"metadata"`
}


// AdaptiveRecommendationAlgorithm 自适应推荐算法类型
type AdaptiveRecommendationAlgorithm string

const (
	AlgorithmCollaborativeFiltering AdaptiveRecommendationAlgorithm = "collaborative_filtering"
	AlgorithmContentBased          AdaptiveRecommendationAlgorithm = "content_based"
	AlgorithmKnowledgeBased        AdaptiveRecommendationAlgorithm = "knowledge_based"
	AlgorithmHybrid                AdaptiveRecommendationAlgorithm = "hybrid"
)

// RecommendationFilter 推荐过滤器
type RecommendationFilter struct {
	FilterType string                 `json:"filter_type"`
	Criteria   map[string]interface{} `json:"criteria"`
	Weight     float64                `json:"weight"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// RecommendationRankingConfig 推荐排序配置
type RecommendationRankingConfig struct {
	RankingMethod string                 `json:"ranking_method"`
	Weights       map[string]float64     `json:"weights"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// RecommendationDiversityConfig 推荐多样性配置
type RecommendationDiversityConfig struct {
	DiversityMethod string                 `json:"diversity_method"`
	DiversityWeight float64                `json:"diversity_weight"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// OptimizationSettings 优化设置
type OptimizationSettings struct {
	OptimizationMethod string                 `json:"optimization_method"`
	Parameters         map[string]interface{} `json:"parameters"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// PersonalizationEngine 个性化引擎
type PersonalizationEngine struct {
	EngineID               string                           `json:"engine_id"`
	EngineType             string                           `json:"engine_type"`
	PersonalizationModels  map[string]*PersonalizationModel `json:"personalization_models"`
	LearnerModels          map[string]*LearnerModel         `json:"learner_models"`
	PersonalizationRules   []*PersonalizationRule           `json:"personalization_rules"`
	PersonalizationHistory []*PersonalizationRecord         `json:"personalization_history"`
	PersonalizationMetrics *PersonalizationMetrics          `json:"personalization_metrics"`
	Config                 map[string]interface{}           `json:"config"`
	Metadata               map[string]interface{}           `json:"metadata"`
}


// PersonalizationModel 个性化模型
type PersonalizationModel struct {
	ModelID     string                 `json:"model_id"`
	ModelType   string                 `json:"model_type"`
	ModelData   map[string]interface{} `json:"model_data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// LearnerModel 学习者模型
type LearnerModel struct {
	ModelID     string                 `json:"model_id"`
	LearnerID   string                 `json:"learner_id"`
	ModelData   map[string]interface{} `json:"model_data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PersonalizationRule 个性化规则
type PersonalizationRule struct {
	RuleID      string                 `json:"rule_id"`
	RuleType    string                 `json:"rule_type"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     map[string]interface{} `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PersonalizationRecord 个性化记录
type PersonalizationRecord struct {
	RecordID    string                 `json:"record_id"`
	LearnerID   string                 `json:"learner_id"`
	Action      string                 `json:"action"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PersonalizationMetrics 个性化指标
type PersonalizationMetrics struct {
	TotalPersonalizations int                    `json:"total_personalizations"`
	SuccessRate           float64                `json:"success_rate"`
	Metadata              map[string]interface{} `json:"metadata"`
}