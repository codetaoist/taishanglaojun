package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
)

// LearningContentRepository 定义学习内容数据访问接口
type LearningContentRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, content *entities.LearningContent) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.LearningContent, error)
	Update(ctx context.Context, content *entities.LearningContent) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entities.LearningContent, error)

	// 查询操作
	GetByType(ctx context.Context, contentType entities.ContentType, offset, limit int) ([]*entities.LearningContent, error)
	GetByStatus(ctx context.Context, status entities.ContentStatus, offset, limit int) ([]*entities.LearningContent, error)
	GetByDifficulty(ctx context.Context, minLevel, maxLevel entities.DifficultyLevel, offset, limit int) ([]*entities.LearningContent, error)
	GetByTags(ctx context.Context, tags []string, offset, limit int) ([]*entities.LearningContent, error)
	GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*entities.LearningContent, error)
	GetByKnowledgeNode(ctx context.Context, nodeID uuid.UUID, offset, limit int) ([]*entities.LearningContent, error)

	// 搜索操作
	Search(ctx context.Context, query *ContentSearchQuery) ([]*entities.LearningContent, int, error)
	SearchByKeywords(ctx context.Context, keywords []string, offset, limit int) ([]*entities.LearningContent, error)
	GetRecommendedContent(ctx context.Context, learnerID uuid.UUID, limit int) ([]*ContentRecommendation, error)
	GetPersonalizedContent(ctx context.Context, learnerID uuid.UUID, preferences *ContentPreferences, limit int) ([]*entities.LearningContent, error)

	// 内容关系操作
	GetPrerequisites(ctx context.Context, contentID uuid.UUID) ([]*entities.LearningContent, error)
	GetFollowUpContent(ctx context.Context, contentID uuid.UUID) ([]*entities.LearningContent, error)
	GetRelatedContent(ctx context.Context, contentID uuid.UUID, limit int) ([]*entities.LearningContent, error)
	GetContentSequence(ctx context.Context, sequenceID uuid.UUID) ([]*entities.LearningContent, error)

	// 内容进度操作
	CreateProgress(ctx context.Context, progress *entities.ContentProgress) error
	GetProgress(ctx context.Context, learnerID, contentID uuid.UUID) (*entities.ContentProgress, error)
	UpdateProgress(ctx context.Context, progress *entities.ContentProgress) error
	GetLearnerProgress(ctx context.Context, learnerID uuid.UUID, offset, limit int) ([]*entities.ContentProgress, error)
	GetContentProgressStats(ctx context.Context, contentID uuid.UUID) (*ContentProgressStats, error)

	// 交互记录操作
	AddInteractionRecord(ctx context.Context, record *entities.InteractionRecord) error
	GetInteractionRecords(ctx context.Context, learnerID, contentID uuid.UUID, limit int) ([]*entities.InteractionRecord, error)
	GetInteractionsByType(ctx context.Context, learnerID uuid.UUID, actionType string, limit int) ([]*entities.InteractionRecord, error)
	GetContentInteractions(ctx context.Context, contentID uuid.UUID, limit int) ([]*entities.InteractionRecord, error)

	// 学习笔记操作
	AddLearningNote(ctx context.Context, note *entities.LearningNote) error
	GetLearningNotes(ctx context.Context, learnerID, contentID uuid.UUID) ([]*entities.LearningNote, error)
	UpdateLearningNote(ctx context.Context, note *entities.LearningNote) error
	DeleteLearningNote(ctx context.Context, noteID uuid.UUID) error
	SearchNotes(ctx context.Context, learnerID uuid.UUID, query string, limit int) ([]*entities.LearningNote, error)

	// 书签操作
	AddBookmark(ctx context.Context, bookmark *entities.Bookmark) error
	GetBookmarks(ctx context.Context, learnerID uuid.UUID, offset, limit int) ([]*entities.Bookmark, error)
	GetContentBookmarks(ctx context.Context, learnerID, contentID uuid.UUID) ([]*entities.Bookmark, error)
	UpdateBookmark(ctx context.Context, bookmark *entities.Bookmark) error
	DeleteBookmark(ctx context.Context, bookmarkID uuid.UUID) error

	// 内容分析和统计
	GetContentAnalytics(ctx context.Context, contentID uuid.UUID) (*entities.ContentAnalytics, error)
	UpdateContentAnalytics(ctx context.Context, contentID uuid.UUID, analytics *entities.ContentAnalytics) error
	GetPopularContent(ctx context.Context, timeRange TimeRange, limit int) ([]*PopularContent, error)
	GetContentEffectiveness(ctx context.Context, contentID uuid.UUID) (*ContentEffectiveness, error)
	GetLearningOutcomes(ctx context.Context, contentID uuid.UUID) (*LearningOutcomes, error)

	// 内容质量和反馈
	GetFeedbackSummary(ctx context.Context, contentID uuid.UUID) (*entities.FeedbackSummary, error)
	AddContentFeedback(ctx context.Context, feedback *ContentFeedback) error
	GetContentRating(ctx context.Context, contentID uuid.UUID) (*ContentRating, error)
	GetContentReviews(ctx context.Context, contentID uuid.UUID, offset, limit int) ([]*ContentReview, error)

	// 内容版本控制
	CreateContentVersion(ctx context.Context, contentID uuid.UUID, version *ContentVersion) error
	GetContentVersions(ctx context.Context, contentID uuid.UUID) ([]*ContentVersion, error)
	RestoreContentVersion(ctx context.Context, contentID, versionID uuid.UUID) error
	CompareContentVersions(ctx context.Context, contentID, version1ID, version2ID uuid.UUID) (*ContentComparison, error)

	// 批量操作
	BatchCreate(ctx context.Context, contents []*entities.LearningContent) error
	BatchUpdate(ctx context.Context, contents []*entities.LearningContent) error
	BatchUpdateStatus(ctx context.Context, contentIDs []uuid.UUID, status entities.ContentStatus) error
	BatchDelete(ctx context.Context, contentIDs []uuid.UUID) error

	// 内容管理
	PublishContent(ctx context.Context, contentID uuid.UUID) error
	ArchiveContent(ctx context.Context, contentID uuid.UUID) error
	ValidateContent(ctx context.Context, contentID uuid.UUID) (*ContentValidation, error)
	OptimizeContent(ctx context.Context, contentID uuid.UUID) (*ContentOptimization, error)

	// 导入导出
	ExportContent(ctx context.Context, contentIDs []uuid.UUID, format string) ([]byte, error)
	ImportContent(ctx context.Context, data []byte, format string) ([]*entities.LearningContent, error)
	BackupContent(ctx context.Context, contentIDs []uuid.UUID) (*ContentBackup, error)
	RestoreContent(ctx context.Context, backupID uuid.UUID) error
}

// ContentSearchQuery 内容搜索查询
type ContentSearchQuery struct {
	// 基本搜索
	Keywords    []string `json:"keywords,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`

	// 类型和状态过滤
	ContentType *entities.ContentType   `json:"content_type,omitempty"`
	Status      *entities.ContentStatus `json:"status,omitempty"`

	// 难度过滤
	DifficultyLevel *entities.DifficultyLevel `json:"difficulty_level,omitempty"`
	MinDifficulty   *int                      `json:"min_difficulty,omitempty"`
	MaxDifficulty   *int                      `json:"max_difficulty,omitempty"`

	// 时间过滤
	MinDuration   *time.Duration `json:"min_duration,omitempty"`
	MaxDuration   *time.Duration `json:"max_duration,omitempty"`
	CreatedAfter  *time.Time     `json:"created_after,omitempty"`
	CreatedBefore *time.Time     `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time     `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time     `json:"updated_before,omitempty"`

	// 标签和分类
	Tags       []string `json:"tags,omitempty"`
	Categories []string `json:"categories,omitempty"`

	// 作者和来源
	AuthorID *uuid.UUID `json:"author_id,omitempty"`
	Source   string     `json:"source,omitempty"`

	// 知识图谱相关
	KnowledgeNodeID *uuid.UUID `json:"knowledge_node_id,omitempty"`
	SkillAreas      []string   `json:"skill_areas,omitempty"`

	// 质量过滤
	MinRating         *float64 `json:"min_rating,omitempty"`
	MinCompletionRate *float64 `json:"min_completion_rate,omitempty"`

	// 个性化过滤
	LearnerID     *uuid.UUID              `json:"learner_id,omitempty"`
	LearningStyle *entities.LearningStyle `json:"learning_style,omitempty"`

	// 排序
	SortBy    string `json:"sort_by,omitempty"`    // "title", "created_at", "updated_at", "rating", "popularity", "difficulty"
	SortOrder string `json:"sort_order,omitempty"` // "asc", "desc"

	// 分页
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// ContentPreferences 内容偏好
type ContentPreferences struct {
	PreferredTypes      []entities.ContentType   `json:"preferred_types"`
	PreferredDifficulty entities.DifficultyLevel `json:"preferred_difficulty"`
	PreferredDuration   time.Duration            `json:"preferred_duration"`
	LearningStyle       entities.LearningStyle   `json:"learning_style"`
	InterestAreas       []string                 `json:"interest_areas"`
	AvoidTopics         []string                 `json:"avoid_topics"`
	PreferredLanguages  []string                 `json:"preferred_languages"`
	AccessibilityNeeds  []string                 `json:"accessibility_needs"`
}

// ContentRecommendation 内容推荐
type ContentRecommendation struct {
	Content                *entities.LearningContent `json:"content"`
	RecommendationScore    float64                   `json:"recommendation_score"`
	Reasoning              []string                  `json:"reasoning"`
	PersonalizationFactors []string                  `json:"personalization_factors"`
	EstimatedEngagement    float64                   `json:"estimated_engagement"`
	DifficultyMatch        float64                   `json:"difficulty_match"`
	StyleMatch             float64                   `json:"style_match"`
	Priority               int                       `json:"priority"`
}

// ContentProgressStats 内容进度统计
type ContentProgressStats struct {
	ContentID             uuid.UUID          `json:"content_id"`
	TotalLearners         int                `json:"total_learners"`
	CompletedLearners     int                `json:"completed_learners"`
	CompletionRate        float64            `json:"completion_rate"`
	AverageProgress       float64            `json:"average_progress"`
	AverageCompletionTime time.Duration      `json:"average_completion_time"`
	DropoffPoints         []DropoffPoint     `json:"dropoff_points"`
	EngagementMetrics     EngagementMetrics  `json:"engagement_metrics"`
	DifficultyFeedback    DifficultyFeedback `json:"difficulty_feedback"`
}

// DropoffPoint 流失点
type DropoffPoint struct {
	Position      float64  `json:"position"` // 0.0 to 1.0
	DropoffRate   float64  `json:"dropoff_rate"`
	LearnerCount  int      `json:"learner_count"`
	CommonReasons []string `json:"common_reasons"`
}

// EngagementMetrics 参与度指标
type EngagementMetrics struct {
	AverageTimeSpent time.Duration `json:"average_time_spent"`
	InteractionRate  float64       `json:"interaction_rate"`
	ReturnRate       float64       `json:"return_rate"`
	ShareRate        float64       `json:"share_rate"`
	BookmarkRate     float64       `json:"bookmark_rate"`
	NotesTakenRate   float64       `json:"notes_taken_rate"`
	QuizAttemptRate  float64       `json:"quiz_attempt_rate"`
	DiscussionRate   float64       `json:"discussion_rate"`
}

// DifficultyFeedback 难度反馈
type DifficultyFeedback struct {
	PerceivedDifficulty    float64        `json:"perceived_difficulty"`
	DifficultyDistribution map[string]int `json:"difficulty_distribution"`
	TooEasyCount           int            `json:"too_easy_count"`
	TooHardCount           int            `json:"too_hard_count"`
	JustRightCount         int            `json:"just_right_count"`
	SuggestedAdjustment    float64        `json:"suggested_adjustment"`
}

// PopularContent 热门内容
type PopularContent struct {
	Content         *entities.LearningContent `json:"content"`
	ViewCount       int                       `json:"view_count"`
	CompletionCount int                       `json:"completion_count"`
	Rating          float64                   `json:"rating"`
	TrendScore      float64                   `json:"trend_score"`
	GrowthRate      float64                   `json:"growth_rate"`
}

// ContentEffectiveness 内容有效性
type ContentEffectiveness struct {
	ContentID             uuid.UUID `json:"content_id"`
	LearningEffectiveness float64   `json:"learning_effectiveness"`
	SkillImprovement      float64   `json:"skill_improvement"`
	KnowledgeRetention    float64   `json:"knowledge_retention"`
	EngagementScore       float64   `json:"engagement_score"`
	CompletionQuality     float64   `json:"completion_quality"`
	LearnerSatisfaction   float64   `json:"learner_satisfaction"`
	RecommendationRate    float64   `json:"recommendation_rate"`
	OverallScore          float64   `json:"overall_score"`
}

// LearningOutcomes 学习成果
type LearningOutcomes struct {
	ContentID             uuid.UUID             `json:"content_id"`
	SkillsAcquired        map[string]float64    `json:"skills_acquired"`
	KnowledgeGained       map[string]float64    `json:"knowledge_gained"`
	CompetenciesImproved  map[string]float64    `json:"competencies_improved"`
	LearningObjectivesMet map[uuid.UUID]float64 `json:"learning_objectives_met"`
	AssessmentResults     AssessmentResults     `json:"assessment_results"`
	LongTermRetention     float64               `json:"long_term_retention"`
	ApplicationRate       float64               `json:"application_rate"`
}

// AssessmentResults 评估结果
type AssessmentResults struct {
	AverageScore      float64        `json:"average_score"`
	PassRate          float64        `json:"pass_rate"`
	ScoreDistribution map[string]int `json:"score_distribution"`
	CommonMistakes    []string       `json:"common_mistakes"`
	ImprovementAreas  []string       `json:"improvement_areas"`
	StrengthAreas     []string       `json:"strength_areas"`
}

// ContentFeedback 内容反馈
type ContentFeedback struct {
	ID             uuid.UUID `json:"id"`
	ContentID      uuid.UUID `json:"content_id"`
	LearnerID      uuid.UUID `json:"learner_id"`
	Rating         float64   `json:"rating"`
	Difficulty     float64   `json:"difficulty"`
	Usefulness     float64   `json:"usefulness"`
	Clarity        float64   `json:"clarity"`
	Engagement     float64   `json:"engagement"`
	Comments       string    `json:"comments"`
	Suggestions    string    `json:"suggestions"`
	ReportedIssues []string  `json:"reported_issues"`
	CreatedAt      time.Time `json:"created_at"`
}

// ContentRating 内容评分
type ContentRating struct {
	ContentID          uuid.UUID   `json:"content_id"`
	AverageRating      float64     `json:"average_rating"`
	RatingCount        int         `json:"rating_count"`
	RatingDistribution map[int]int `json:"rating_distribution"`
	RecentRating       float64     `json:"recent_rating"`
	TrendDirection     string      `json:"trend_direction"` // "up", "down", "stable"
}

// ContentReview 内容评论
type ContentReview struct {
	ID           uuid.UUID `json:"id"`
	ContentID    uuid.UUID `json:"content_id"`
	LearnerID    uuid.UUID `json:"learner_id"`
	LearnerName  string    `json:"learner_name"`
	Rating       float64   `json:"rating"`
	Title        string    `json:"title"`
	Review       string    `json:"review"`
	Pros         []string  `json:"pros"`
	Cons         []string  `json:"cons"`
	Helpful      bool      `json:"helpful"`
	HelpfulCount int       `json:"helpful_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ContentVersion 内容版本
type ContentVersion struct {
	ID          uuid.UUID       `json:"id"`
	ContentID   uuid.UUID       `json:"content_id"`
	Version     string          `json:"version"`
	Description string          `json:"description"`
	Changes     []ContentChange `json:"changes"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   uuid.UUID       `json:"created_by"`
	Snapshot    []byte          `json:"snapshot"`
}

// ContentChange 内容变更
type ContentChange struct {
	Field      string      `json:"field"`
	OldValue   interface{} `json:"old_value"`
	NewValue   interface{} `json:"new_value"`
	ChangeType string      `json:"change_type"` // "add", "update", "remove"
	Timestamp  time.Time   `json:"timestamp"`
}

// ContentComparison 内容比较
type ContentComparison struct {
	ContentID    uuid.UUID           `json:"content_id"`
	Version1ID   uuid.UUID           `json:"version1_id"`
	Version2ID   uuid.UUID           `json:"version2_id"`
	Differences  []ContentDifference `json:"differences"`
	Summary      string              `json:"summary"`
	ChangeCount  int                 `json:"change_count"`
	MajorChanges []string            `json:"major_changes"`
}

// ContentDifference 内容差异
type ContentDifference struct {
	Field    string      `json:"field"`
	Type     string      `json:"type"` // "added", "removed", "modified"
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	Impact   string      `json:"impact"` // "low", "medium", "high"
}

// ContentValidation 内容验证
type ContentValidation struct {
	ContentID       uuid.UUID           `json:"content_id"`
	IsValid         bool                `json:"is_valid"`
	Errors          []ValidationError   `json:"errors"`
	Warnings        []ValidationWarning `json:"warnings"`
	QualityScore    float64             `json:"quality_score"`
	Recommendations []string            `json:"recommendations"`
}

// ContentOptimization 内容优化
type ContentOptimization struct {
	ContentID         uuid.UUID                `json:"content_id"`
	OptimizationScore float64                  `json:"optimization_score"`
	Suggestions       []OptimizationSuggestion `json:"suggestions"`
	PredictedImpact   float64                  `json:"predicted_impact"`
	EstimatedEffort   string                   `json:"estimated_effort"` // "low", "medium", "high"
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	Priority       int      `json:"priority"`
	ExpectedImpact float64  `json:"expected_impact"`
	Implementation string   `json:"implementation"`
	Resources      []string `json:"resources"`
}

// ContentBackup 内容备份
type ContentBackup struct {
	ID          uuid.UUID   `json:"id"`
	ContentIDs  []uuid.UUID `json:"content_ids"`
	BackupData  []byte      `json:"backup_data"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   uuid.UUID   `json:"created_by"`
	Description string      `json:"description"`
	Size        int64       `json:"size"`
	Checksum    string      `json:"checksum"`
}

// ContentStatistics 内容统计信息
type ContentStatistics struct {
	TotalContent     int                            `json:"total_content"`
	ContentByType    map[entities.ContentType]int   `json:"content_by_type"`
	ContentByStatus  map[entities.ContentStatus]int `json:"content_by_status"`
	TotalViews       int64                          `json:"total_views"`
	TotalCompletions int64                          `json:"total_completions"`
	AverageRating    float64                        `json:"average_rating"`
	AverageDuration  time.Duration                  `json:"average_duration"`
	PopularContent   []*entities.LearningContent    `json:"popular_content"`
	LastUpdated      time.Time                      `json:"last_updated"`
}
