package infrastructure

import (
	"context"
	"time"

	"github.com/google/uuid"
	learnerservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/learner"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/shared"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
)

// 共享服务接口定义

// 类型别名，引用interfaces包中的接�?
type LearningAnalyticsService = interfaces.LearningAnalyticsService
type KnowledgeGraphService = interfaces.KnowledgeGraphService

// PathRecommendationRequest 路径推荐请求
type PathRecommendationRequest struct {
	LearnerID      uuid.UUID `json:"learner_id" binding:"required"`
	CurrentSkills  []string  `json:"current_skills"`
	InterestAreas  []string  `json:"interest_areas"`
	AvailableTime  int       `json:"available_time,omitempty"` // 可用学习时间（小�?周）
	LearningGoals  []string  `json:"learning_goals"`
}

// PathRecommendationResponse 路径推荐响应
type PathRecommendationResponse struct {
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
	Reasoning        string            `json:"reasoning"`
	Confidence       float64           `json:"confidence"`
}

// RecommendedPath 推荐路径
type RecommendedPath struct {
	PathID          uuid.UUID `json:"path_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	MatchScore      float64   `json:"match_score"`
	EstimatedTime   int       `json:"estimated_time"`
	DifficultyLevel string    `json:"difficulty_level"`
	SkillsGained    []string  `json:"skills_gained"`
	Reasons         []string  `json:"reasons"`
}

// LearningPathServiceInterface 学习路径服务接口
type LearningPathServiceInterface interface {
	GetRecommendedPaths(ctx context.Context, req *learnerservices.PathRecommendationRequest) (*learnerservices.PathRecommendationResponse, error)
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
    // 现有字段（保持向后兼容）
    LearnerID   uuid.UUID          `json:"learner_id"`
    TimeRange   AnalyticsTimeRange `json:"time_range"`
    GeneratedAt time.Time          `json:"generated_at"`

    // 新增字段（用于学习分析报告服务的使用场景�?
    ID               uuid.UUID              `json:"id"`
    Type             AnalyticsReportType    `json:"type"`
    Title            string                 `json:"title"`
    Description      string                 `json:"description"`
    GeneratedFor     *ReportTarget          `json:"generated_for"`
    ReportingTimeRange *ReportingTimeRange  `json:"reporting_time_range"`
    DataSources      []*DataSource          `json:"data_sources"`
    Sections         []*ReportSection       `json:"sections"`
    Visualizations   []*Visualization       `json:"visualizations"`
    Insights         []*Insight             `json:"insights"`
    Recommendations  []*Recommendation      `json:"recommendations"`
    Summary          *ReportSummary         `json:"summary"`
    Metadata         *ReportMetadata        `json:"metadata"`
    QualityScore     float64                `json:"quality_score"`
    GenerationTime   time.Duration          `json:"generation_time"`
    CreatedAt        time.Time              `json:"created_at"`
    ExpiresAt        time.Time              `json:"expires_at"`
    Version          string                 `json:"version"`
    Tags             []string               `json:"tags"`
    AccessLevel      AccessLevel            `json:"access_level"`
    ExportFormats    []ExportFormat         `json:"export_formats"`
    CustomData       map[string]interface{} `json:"custom_data"`
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
	DifficultyLevel    string        `json:"difficulty_level"`
	LearningStyle      string        `json:"learning_style"`
	TimeConstraint     time.Duration `json:"time_constraint"`
	MaxPathLength      int           `json:"max_path_length"`
	PreferShortPaths   bool          `json:"prefer_short_paths"`
	AdaptiveDifficulty bool          `json:"adaptive_difficulty"`
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

// AnalyticsReportType 分析报告类型
type AnalyticsReportType string

const (
	ReportTypePerformance AnalyticsReportType = "performance"
	ReportTypeProgress    AnalyticsReportType = "progress"
	ReportTypeBehavior    AnalyticsReportType = "behavior"
	ReportTypeEngagement  AnalyticsReportType = "engagement"
	ReportTypeComparative AnalyticsReportType = "comparative"
)

// ReportTarget 报告目标
type ReportTarget struct {
	Type       string    `json:"type"`        // learner, group, course, etc.
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ReportingTimeRange 报告时间范围
type ReportingTimeRange struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Granularity string    `json:"granularity"` // daily, weekly, monthly
	Timezone    string    `json:"timezone"`
}

// DataSource 数据�?
type DataSource struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ReportSection 报告章节
type ReportSection struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content"`
	Order       int                    `json:"order"`
	Type        string                 `json:"type"`
}

// Visualization 可视�?
type Visualization struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Data        map[string]interface{} `json:"data"`
	Config      map[string]interface{} `json:"config"`
	Order       int                    `json:"order"`
}

// Insight 洞察
type Insight struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Confidence  float64                `json:"confidence"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Recommendation 推荐
type Recommendation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	Confidence  float64                `json:"confidence"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	TotalDataPoints    int                    `json:"total_data_points"`
	KeyFindings        []string               `json:"key_findings"`
	MainInsights       []string               `json:"main_insights"`
	RecommendationCount int                   `json:"recommendation_count"`
	QualityScore       float64                `json:"quality_score"`
	CompletionRate     float64                `json:"completion_rate"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// ReportMetadata 报告元数�?
type ReportMetadata struct {
	GeneratedBy     string                 `json:"generated_by"`
	GenerationTime  time.Duration          `json:"generation_time"`
	DataSources     []string               `json:"data_sources"`
	ProcessingSteps []string               `json:"processing_steps"`
	Version         string                 `json:"version"`
	Checksum        string                 `json:"checksum"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AccessLevel 访问级别
type AccessLevel string

const (
	AccessLevelPublic     AccessLevel = "public"
	AccessLevelPrivate    AccessLevel = "private"
	AccessLevelRestricted AccessLevel = "restricted"
	AccessLevelInternal   AccessLevel = "internal"
)

// ExportFormat 导出格式
type ExportFormat string

const (
	ExportFormatPDF   ExportFormat = "pdf"
	ExportFormatExcel ExportFormat = "excel"
	ExportFormatCSV   ExportFormat = "csv"
	ExportFormatJSON  ExportFormat = "json"
	ExportFormatHTML  ExportFormat = "html"
)

// VisualizationConfig 可视化配�?
type VisualizationConfig struct {
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	XAxis      string                 `json:"x_axis"`
	YAxis      string                 `json:"y_axis"`
	Colors     []string               `json:"colors"`
	Options    map[string]interface{} `json:"options"`
}



// AnalyticsDataCollection 分析数据集合
type AnalyticsDataCollection struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	DataType    string                 `json:"data_type"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
	Schema      map[string]interface{} `json:"schema"`
	CollectedAt time.Time              `json:"collected_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ProcessedAnalyticsData 处理后的分析数据
type ProcessedAnalyticsData struct {
    ProcessingID  uuid.UUID                  `json:"processing_id"`
    SourceData    *AnalyticsDataCollection   `json:"source_data"`
    ProcessedData map[string]interface{}     `json:"processed_data"`
    Aggregations  map[string]*DataAggregation `json:"aggregations"`
    Statistics    map[string]*StatisticalSummary `json:"statistics"`
    QualityScore  float64                    `json:"quality_score"`
    ProcessedAt   time.Time                  `json:"processed_at"`
    Metadata      map[string]interface{}     `json:"metadata"`
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


// 使用shared包中的类型定�?
type RecommendationSettings = shared.RecommendationSettings
type OptimizationSettings = shared.OptimizationSettings
type AdaptiveCachedLearnerProfile = shared.AdaptiveCachedLearnerProfile
type CachedLearningStrategy = shared.CachedLearningStrategy
type CachedAdaptationResult = shared.CachedAdaptationResult
type CachedPersonalizationData = shared.CachedPersonalizationData
type CachedLearningPath = shared.CachedLearningPath
type CachedAssessmentResult = shared.CachedAssessmentResult
type CachedRecommendationResult = shared.CachedRecommendationResult
type PersonalizationRule = shared.PersonalizationRule
type RecommendationFilter = shared.RecommendationFilter
type RecommendationRankingConfig = shared.RecommendationRankingConfig
type RecommendationDiversityConfig = shared.RecommendationDiversityConfig
type AdaptiveRecommendationAlgorithm = shared.AdaptiveRecommendationAlgorithm
type RecommendationType = shared.RecommendationType
type RecommendationStrategy = shared.RecommendationStrategy
type LearnerProfile = shared.LearnerProfile
type LearningStrategy = shared.LearningStrategy
type AdaptationResponse = shared.AdaptationResponse
type PersonalizationData = shared.PersonalizationData
type LearningPath = shared.LearningPath

// 使用shared包中的类型定�?
type AssessmentResult = shared.AssessmentResult
type RecommendationResult = shared.RecommendationResult

// PersonalizedFeedback 个性化反馈
type PersonalizedFeedback struct {
	FeedbackID      uuid.UUID              `json:"feedback_id"`
	LearnerID       uuid.UUID              `json:"learner_id"`
	ContentID       uuid.UUID              `json:"content_id"`
	FeedbackType    string                 `json:"feedback_type"`
	Message         string                 `json:"message"`
	Tone            string                 `json:"tone"`
	DetailLevel     string                 `json:"detail_level"`
	Timing          string                 `json:"timing"`
	Effectiveness   float64                `json:"effectiveness"`
	PersonalizationLevel float64           `json:"personalization_level"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// LearningProgress 学习进度
type LearningProgress struct {
	ProgressID      uuid.UUID              `json:"progress_id"`
	LearnerID       uuid.UUID              `json:"learner_id"`
	ContentID       uuid.UUID              `json:"content_id"`
	CompletionRate  float64                `json:"completion_rate"`
	TimeSpent       time.Duration          `json:"time_spent"`
	PerformanceScore float64               `json:"performance_score"`
	MasteryLevel    float64                `json:"mastery_level"`
	EngagementLevel float64                `json:"engagement_level"`
	LastActivity    time.Time              `json:"last_activity"`
	Milestones      []ProgressMilestone    `json:"milestones"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ProgressMilestone 进度里程�?
type ProgressMilestone struct {
	MilestoneID   uuid.UUID `json:"milestone_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	AchievedAt    time.Time `json:"achieved_at"`
	Score         float64   `json:"score"`
	IsCompleted   bool      `json:"is_completed"`
}

// ContentItem 内容�?
type ContentItem struct {
	ContentID     uuid.UUID              `json:"content_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	ContentType   string                 `json:"content_type"`
	DifficultyLevel string               `json:"difficulty_level"`
	Duration      time.Duration          `json:"duration"`
	Tags          []string               `json:"tags"`
	Skills        []string               `json:"skills"`
	Prerequisites []uuid.UUID            `json:"prerequisites"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Skill 技�?
type Skill struct {
	SkillID     uuid.UUID `json:"skill_id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	AcquiredAt  time.Time `json:"acquired_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LearningResource 学习资源
type LearningResource struct {
	ResourceID   uuid.UUID              `json:"resource_id"`
	Title        string                 `json:"title"`
	Type         string                 `json:"type"`
	URL          string                 `json:"url"`
	Description  string                 `json:"description"`
	Tags         []string               `json:"tags"`
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// LearningAnalytics 学习分析
type LearningAnalytics struct {
	AnalyticsID   uuid.UUID              `json:"analytics_id"`
	LearnerID     uuid.UUID              `json:"learner_id"`
	ContentID     uuid.UUID              `json:"content_id"`
	SessionID     uuid.UUID              `json:"session_id"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	Interactions  int                    `json:"interactions"`
	Score         float64                `json:"score"`
	Completion    float64                `json:"completion"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// OutcomePrediction 结果预测
type OutcomePrediction struct {
	PredictionID    uuid.UUID              `json:"prediction_id"`
	LearnerID       uuid.UUID              `json:"learner_id"`
	ContentID       uuid.UUID              `json:"content_id"`
	PredictedScore  float64                `json:"predicted_score"`
	Confidence      float64                `json:"confidence"`
	Factors         []string               `json:"factors"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// UserInterface 用户界面
type UserInterface struct {
	InterfaceID   uuid.UUID              `json:"interface_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Layout        string                 `json:"layout"`
	Theme         string                 `json:"theme"`
	Preferences   map[string]interface{} `json:"preferences"`
	LastUpdated   time.Time              `json:"last_updated"`
}

// UsageAnalytics 使用分析
type UsageAnalytics struct {
	AnalyticsID   uuid.UUID              `json:"analytics_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Feature       string                 `json:"feature"`
	Action        string                 `json:"action"`
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// MotivationAnalysis 动机分析
type MotivationAnalysis struct {
	AnalysisID      uuid.UUID              `json:"analysis_id"`
	LearnerID       uuid.UUID              `json:"learner_id"`
	MotivationType  string                 `json:"motivation_type"`
	Level           float64                `json:"level"`
	Factors         []string               `json:"factors"`
	Recommendations []string               `json:"recommendations"`
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// MotivationalContent 激励内�?
type MotivationalContent struct {
	ContentID     uuid.UUID              `json:"content_id"`
	Title         string                 `json:"title"`
	Type          string                 `json:"type"`
	Message       string                 `json:"message"`
	TargetGroup   string                 `json:"target_group"`
	Effectiveness float64                `json:"effectiveness"`
	CreatedAt     time.Time              `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// LearningPace 学习节奏
type LearningPace struct {
	PaceID        uuid.UUID `json:"pace_id"`
	LearnerID     uuid.UUID `json:"learner_id"`
	Speed         float64   `json:"speed"`
	Consistency   float64   `json:"consistency"`
	OptimalPace   float64   `json:"optimal_pace"`
	LastUpdated   time.Time `json:"last_updated"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	MetricsID     uuid.UUID              `json:"metrics_id"`
	LearnerID     uuid.UUID              `json:"learner_id"`
	Accuracy      float64                `json:"accuracy"`
	Speed         float64                `json:"speed"`
	Retention     float64                `json:"retention"`
	Engagement    float64                `json:"engagement"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// LearningObjective 学习目标
type LearningObjective struct {
	ObjectiveID   uuid.UUID `json:"objective_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Priority      int       `json:"priority"`
	DueDate       time.Time `json:"due_date"`
	IsCompleted   bool      `json:"is_completed"`
	CompletedAt   time.Time `json:"completed_at"`
}

// ContentVariation 内容变化
type ContentVariation struct {
	VariationID   uuid.UUID              `json:"variation_id"`
	OriginalID    uuid.UUID              `json:"original_id"`
	Type          string                 `json:"type"`
	Difficulty    string                 `json:"difficulty"`
	Format        string                 `json:"format"`
	Adaptations   []string               `json:"adaptations"`
	CreatedAt     time.Time              `json:"created_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}
