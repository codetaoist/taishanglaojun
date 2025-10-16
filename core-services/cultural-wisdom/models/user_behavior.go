package models

import (
	"time"
)

// UserBehavior 
type UserBehavior struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID     string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	WisdomID   string    `json:"wisdom_id" gorm:"type:varchar(255);not null;index"`
	ActionType string    `json:"action_type" gorm:"type:varchar(50);not null"` // view, like, share, comment, favorite, search
	Duration   int64     `json:"duration" gorm:"default:0"`                    // 
	Score      float64   `json:"score" gorm:"default:0"`                       // 
	Context    string    `json:"context" gorm:"type:text"`                     // JSON
	IPAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserPreference 
type UserPreference struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID       string    `json:"user_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	Categories   string    `json:"categories" gorm:"type:text"` // JSON洢
	Schools      string    `json:"schools" gorm:"type:text"`    // JSON洢
	Authors      string    `json:"authors" gorm:"type:text"`    // JSON洢
	Tags         string    `json:"tags" gorm:"type:text"`       // JSON洢
	Difficulty   string    `json:"difficulty" gorm:"type:varchar(50);default:'medium'"`
	ReadingSpeed float64   `json:"reading_speed" gorm:"default:1.0"` // 
	LastActive   time.Time `json:"last_active" gorm:"autoUpdateTime"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserInteraction 
type UserInteraction struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID      string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	TargetID    string    `json:"target_id" gorm:"type:varchar(255);not null"`  // ID
	TargetType  string    `json:"target_type" gorm:"type:varchar(50);not null"` // 
	ActionType  string    `json:"action_type" gorm:"type:varchar(50);not null"` // 
	ActionValue float64   `json:"action_value" gorm:"default:1.0"`              // 
	IsPositive  bool      `json:"is_positive" gorm:"default:true"`              // 
	SessionID   string    `json:"session_id" gorm:"type:varchar(255)"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RecommendationLog 
type RecommendationLog struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID        string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	Algorithm     string    `json:"algorithm" gorm:"type:varchar(50);not null"` // 㷨
	RequestParams string    `json:"request_params" gorm:"type:text"`            // JSON
	Results       string    `json:"results" gorm:"type:text"`                   // JSON
	ResultCount   int       `json:"result_count" gorm:"default:0"`              // 
	ClickCount    int       `json:"click_count" gorm:"default:0"`               // 
	CTR           float64   `json:"ctr" gorm:"default:0"`                       // 
	SessionID     string    `json:"session_id" gorm:"type:varchar(255)"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserSimilarity 
type UserSimilarity struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID1     string    `json:"user_id1" gorm:"type:varchar(255);not null;index"`
	UserID2     string    `json:"user_id2" gorm:"type:varchar(255);not null;index"`
	Similarity  float64   `json:"similarity" gorm:"not null"`                 // 0-1
	Algorithm   string    `json:"algorithm" gorm:"type:varchar(50);not null"` // 㷨
	CommonItems int       `json:"common_items" gorm:"default:0"`              // 
	LastUpdated time.Time `json:"last_updated" gorm:"autoUpdateTime"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ActionType 
const (
	ActionTypeView     = "view"
	ActionTypeLike     = "like"
	ActionTypeShare    = "share"
	ActionTypeComment  = "comment"
	ActionTypeFavorite = "favorite"
	ActionTypeSearch   = "search"
	ActionTypeDownload = "download"
)

// TargetType 
const (
	TargetTypeWisdom   = "wisdom"
	TargetTypeUser     = "user"
	TargetTypeCategory = "category"
	TargetTypeTag      = "tag"
)

// Algorithm 
const (
	AlgorithmContent       = "content"
	AlgorithmCollaborative = "collaborative"
	AlgorithmHybrid        = "hybrid"
	AlgorithmPopular       = "popular"
	AlgorithmTrending      = "trending"
)

