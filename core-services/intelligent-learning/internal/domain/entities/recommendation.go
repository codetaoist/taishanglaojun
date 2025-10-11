package entities

import (
	"time"
	"github.com/google/uuid"
)

// RecommendationItem 推荐�?
type RecommendationItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // content, path, skill, etc.
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`       // 推荐分数 0.0-1.0
	Confidence  float64                `json:"confidence"`  // 置信�?0.0-1.0
	Reason      string                 `json:"reason"`      // 推荐原因
	Category    string                 `json:"category"`    // 分类
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	ContentID   *uuid.UUID             `json:"content_id,omitempty"`   // 关联的内容ID
	PathID      *uuid.UUID             `json:"path_id,omitempty"`      // 关联的路径ID
	SkillID     *uuid.UUID             `json:"skill_id,omitempty"`     // 关联的技能ID
	CreatedAt   time.Time              `json:"created_at"`
}

// ContentItem 内容�?
type ContentItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // video, text, audio, interactive, etc.
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Difficulty  string                 `json:"difficulty"`  // beginner, intermediate, advanced, expert
	Duration    int                    `json:"duration"`    // 持续时间（分钟）
	Subject     string                 `json:"subject"`     // 学科
	Category    string                 `json:"category"`    // 分类
	Tags        []string               `json:"tags"`
	Keywords    []string               `json:"keywords"`
	Rating      float64                `json:"rating"`      // 评分 0.0-5.0
	ViewCount   int                    `json:"view_count"`  // 观看次数
	Language    string                 `json:"language"`    // 语言
	AuthorID    string                 `json:"author_id"`   // 作者ID
	AuthorName  string                 `json:"author_name"` // 作者名�?
	Thumbnail   string                 `json:"thumbnail"`   // 缩略图URL
	URL         string                 `json:"url"`         // 内容URL
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UserFeedback 用户反馈
type UserFeedback struct {
	ID        uuid.UUID              `json:"id"`
	UserID    string                 `json:"user_id"`
	ItemID    string                 `json:"item_id"`
	ItemType  string                 `json:"item_type"`  // content, recommendation, path, etc.
	Type      string                 `json:"type"`       // like, dislike, rating, comment, bookmark, share
	Value     interface{}            `json:"value"`      // 反馈值（如评分数值、布尔值等�?
	Rating    *float64               `json:"rating,omitempty"`    // 评分 1.0-5.0
	Comment   string                 `json:"comment"`             // 评论
	Sentiment string                 `json:"sentiment"`           // positive, negative, neutral
	Context   map[string]interface{} `json:"context"`             // 反馈上下�?
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`
}

// RecommendationExplanation 推荐解释
type RecommendationExplanation struct {
	ID          uuid.UUID `json:"id"`
	UserID      string    `json:"user_id"`
	ItemID      string    `json:"item_id"`
	ItemType    string    `json:"item_type"`
	Reasons     []string  `json:"reasons"`     // 推荐原因列表
	Confidence  float64   `json:"confidence"`  // 解释置信�?0.0-1.0
	Evidence    []string  `json:"evidence"`    // 支持证据
	Explanation string    `json:"explanation"` // 详细解释
	Algorithm   string    `json:"algorithm"`   // 使用的算�?
	Factors     map[string]float64 `json:"factors"` // 影响因素及权�?
	CreatedAt   time.Time `json:"created_at"`
}

// RecommendationSession 推荐会话
type RecommendationSession struct {
	ID            uuid.UUID                      `json:"id"`
	UserID        string                         `json:"user_id"`
	SessionType   string                         `json:"session_type"`   // browse, search, learn, etc.
	Context       map[string]interface{}         `json:"context"`        // 会话上下�?
	Recommendations []*RecommendationItem        `json:"recommendations"`
	Interactions  []*RecommendationInteraction   `json:"interactions"`
	StartTime     time.Time                      `json:"start_time"`
	EndTime       *time.Time                     `json:"end_time,omitempty"`
	Duration      time.Duration                  `json:"duration"`
	CreatedAt     time.Time                      `json:"created_at"`
}

// RecommendationInteraction 推荐交互
type RecommendationInteraction struct {
	ID            uuid.UUID              `json:"id"`
	SessionID     uuid.UUID              `json:"session_id"`
	UserID        string                 `json:"user_id"`
	ItemID        string                 `json:"item_id"`
	Action        string                 `json:"action"`        // view, click, like, dislike, bookmark, share, skip
	Position      int                    `json:"position"`      // 在推荐列表中的位�?
	Duration      time.Duration          `json:"duration"`      // 交互持续时间
	Context       map[string]interface{} `json:"context"`
	Timestamp     time.Time              `json:"timestamp"`
}

// UserPreference 用户偏好
type UserPreference struct {
	ID            uuid.UUID              `json:"id"`
	UserID        string                 `json:"user_id"`
	Category      string                 `json:"category"`      // content_type, subject, difficulty, etc.
	Preference    string                 `json:"preference"`    // 偏好�?
	Weight        float64                `json:"weight"`        // 权重 0.0-1.0
	Confidence    float64                `json:"confidence"`    // 置信�?0.0-1.0
	Source        string                 `json:"source"`        // explicit, implicit, inferred
	Context       map[string]interface{} `json:"context"`
	LastUpdated   time.Time              `json:"last_updated"`
	CreatedAt     time.Time              `json:"created_at"`
}

// RecommendationMetrics 推荐指标
type RecommendationMetrics struct {
	ID              uuid.UUID `json:"id"`
	UserID          string    `json:"user_id"`
	Period          string    `json:"period"`          // daily, weekly, monthly
	TotalRecommendations int  `json:"total_recommendations"`
	ClickedRecommendations int `json:"clicked_recommendations"`
	ClickThroughRate float64  `json:"click_through_rate"`
	ConversionRate  float64   `json:"conversion_rate"`
	AverageRating   float64   `json:"average_rating"`
	DiversityScore  float64   `json:"diversity_score"`
	NoveltyScore    float64   `json:"novelty_score"`
	SatisfactionScore float64 `json:"satisfaction_score"`
	Timestamp       time.Time `json:"timestamp"`
}

// NewRecommendationItem 创建新的推荐�?
func NewRecommendationItem(itemType, title, description string, score, confidence float64) *RecommendationItem {
	return &RecommendationItem{
		ID:          uuid.New().String(),
		Type:        itemType,
		Title:       title,
		Description: description,
		Score:       score,
		Confidence:  confidence,
		Tags:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}
}

// NewUserFeedback 创建新的用户反馈
func NewUserFeedback(userID, itemID, itemType, feedbackType string, value interface{}) *UserFeedback {
	now := time.Now()
	return &UserFeedback{
		ID:        uuid.New(),
		UserID:    userID,
		ItemID:    itemID,
		ItemType:  itemType,
		Type:      feedbackType,
		Value:     value,
		Context:   make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		Timestamp: now,
		CreatedAt: now,
	}
}

// NewRecommendationExplanation 创建新的推荐解释
func NewRecommendationExplanation(userID, itemID, itemType string, reasons []string, confidence float64) *RecommendationExplanation {
	return &RecommendationExplanation{
		ID:         uuid.New(),
		UserID:     userID,
		ItemID:     itemID,
		ItemType:   itemType,
		Reasons:    reasons,
		Confidence: confidence,
		Evidence:   make([]string, 0),
		Factors:    make(map[string]float64),
		CreatedAt:  time.Now(),
	}
}
