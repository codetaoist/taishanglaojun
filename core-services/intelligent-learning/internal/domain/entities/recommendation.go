package entities

import (
	"time"
	"github.com/google/uuid"
)

// RecommendationItem ?
type RecommendationItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // content, path, skill, etc.
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`       //  0.0-1.0
	Confidence  float64                `json:"confidence"`  // ?0.0-1.0
	Reason      string                 `json:"reason"`      // 
	Category    string                 `json:"category"`    // 
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	ContentID   *uuid.UUID             `json:"content_id,omitempty"`   // ID
	PathID      *uuid.UUID             `json:"path_id,omitempty"`      // ID
	SkillID     *uuid.UUID             `json:"skill_id,omitempty"`     // ID
	CreatedAt   time.Time              `json:"created_at"`
}

// ContentItem ?
type ContentItem struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // video, text, audio, interactive, etc.
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Difficulty  string                 `json:"difficulty"`  // beginner, intermediate, advanced, expert
	Duration    int                    `json:"duration"`    // 
	Subject     string                 `json:"subject"`     // 
	Category    string                 `json:"category"`    // 
	Tags        []string               `json:"tags"`
	Keywords    []string               `json:"keywords"`
	Rating      float64                `json:"rating"`      //  0.0-5.0
	ViewCount   int                    `json:"view_count"`  // 
	Language    string                 `json:"language"`    // 
	AuthorID    string                 `json:"author_id"`   // ID
	AuthorName  string                 `json:"author_name"` // ?
	Thumbnail   string                 `json:"thumbnail"`   // URL
	URL         string                 `json:"url"`         // URL
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UserFeedback 
type UserFeedback struct {
	ID        uuid.UUID              `json:"id"`
	UserID    string                 `json:"user_id"`
	ItemID    string                 `json:"item_id"`
	ItemType  string                 `json:"item_type"`  // content, recommendation, path, etc.
	Type      string                 `json:"type"`       // like, dislike, rating, comment, bookmark, share
	Value     interface{}            `json:"value"`      // ?
	Rating    *float64               `json:"rating,omitempty"`    //  1.0-5.0
	Comment   string                 `json:"comment"`             // 
	Sentiment string                 `json:"sentiment"`           // positive, negative, neutral
	Context   map[string]interface{} `json:"context"`             // ?
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	CreatedAt time.Time              `json:"created_at"`
}

// RecommendationExplanation 
type RecommendationExplanation struct {
	ID          uuid.UUID `json:"id"`
	UserID      string    `json:"user_id"`
	ItemID      string    `json:"item_id"`
	ItemType    string    `json:"item_type"`
	Reasons     []string  `json:"reasons"`     // 
	Confidence  float64   `json:"confidence"`  // ?0.0-1.0
	Evidence    []string  `json:"evidence"`    // 
	Explanation string    `json:"explanation"` // 
	Algorithm   string    `json:"algorithm"`   // ?
	Factors     map[string]float64 `json:"factors"` // ?
	CreatedAt   time.Time `json:"created_at"`
}

// RecommendationSession 
type RecommendationSession struct {
	ID            uuid.UUID                      `json:"id"`
	UserID        string                         `json:"user_id"`
	SessionType   string                         `json:"session_type"`   // browse, search, learn, etc.
	Context       map[string]interface{}         `json:"context"`        // ?
	Recommendations []*RecommendationItem        `json:"recommendations"`
	Interactions  []*RecommendationInteraction   `json:"interactions"`
	StartTime     time.Time                      `json:"start_time"`
	EndTime       *time.Time                     `json:"end_time,omitempty"`
	Duration      time.Duration                  `json:"duration"`
	CreatedAt     time.Time                      `json:"created_at"`
}

// RecommendationInteraction 
type RecommendationInteraction struct {
	ID            uuid.UUID              `json:"id"`
	SessionID     uuid.UUID              `json:"session_id"`
	UserID        string                 `json:"user_id"`
	ItemID        string                 `json:"item_id"`
	Action        string                 `json:"action"`        // view, click, like, dislike, bookmark, share, skip
	Position      int                    `json:"position"`      // ?
	Duration      time.Duration          `json:"duration"`      // 
	Context       map[string]interface{} `json:"context"`
	Timestamp     time.Time              `json:"timestamp"`
}

// UserPreference 
type UserPreference struct {
	ID            uuid.UUID              `json:"id"`
	UserID        string                 `json:"user_id"`
	Category      string                 `json:"category"`      // content_type, subject, difficulty, etc.
	Preference    string                 `json:"preference"`    // ?
	Weight        float64                `json:"weight"`        //  0.0-1.0
	Confidence    float64                `json:"confidence"`    // ?0.0-1.0
	Source        string                 `json:"source"`        // explicit, implicit, inferred
	Context       map[string]interface{} `json:"context"`
	LastUpdated   time.Time              `json:"last_updated"`
	CreatedAt     time.Time              `json:"created_at"`
}

// RecommendationMetrics 
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

// NewRecommendationItem ?
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

// NewUserFeedback 
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

// NewRecommendationExplanation 
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

