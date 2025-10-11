package interfaces

import (
	"context"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
)

// RecommendationItem 推荐�?
type RecommendationItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`
	Confidence  float64                `json:"confidence"`
	Reason      string                 `json:"reason"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ContentItem 内容�?
type ContentItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Difficulty  string                 `json:"difficulty"`
	Duration    int                    `json:"duration"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UserFeedback 用户反馈
type UserFeedback struct {
	UserID      string                 `json:"user_id"`
	ItemID      string                 `json:"item_id"`
	Type        string                 `json:"type"` // like, dislike, rating, comment
	Value       interface{}            `json:"value"`
	Comment     string                 `json:"comment"`
	Timestamp   int64                  `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RecommendationExplanation 推荐解释
type RecommendationExplanation struct {
	ItemID      string   `json:"item_id"`
	Reasons     []string `json:"reasons"`
	Confidence  float64  `json:"confidence"`
	Evidence    []string `json:"evidence"`
	Explanation string   `json:"explanation"`
}

// IntelligentContentRecommendationService 智能内容推荐服务接口
type IntelligentContentRecommendationService interface {
	// GenerateRecommendations 生成推荐
	GenerateRecommendations(ctx context.Context, userID string, context map[string]interface{}) ([]*entities.RecommendationItem, error)
	
	// GetPersonalizedContent 获取个性化内容
	GetPersonalizedContent(ctx context.Context, userID string, contentType string) ([]*entities.ContentItem, error)
	
	// UpdateUserFeedback 更新用户反馈
	UpdateUserFeedback(ctx context.Context, userID string, feedback *entities.UserFeedback) error
	
	// GetRecommendationExplanation 获取推荐解释
	GetRecommendationExplanation(ctx context.Context, userID string, itemID string) (*entities.RecommendationExplanation, error)
}
