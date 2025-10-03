package models

import (
	"time"
)

// UserBehavior 用户行为记录模型
type UserBehavior struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID     string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	WisdomID   string    `json:"wisdom_id" gorm:"type:varchar(255);not null;index"`
	ActionType string    `json:"action_type" gorm:"type:varchar(50);not null"` // view, like, share, comment, favorite, search
	Duration   int64     `json:"duration" gorm:"default:0"`                    // 浏览时长（秒）
	Score      float64   `json:"score" gorm:"default:0"`                       // 行为评分
	Context    string    `json:"context" gorm:"type:text"`                     // 上下文信息（JSON格式）
	IPAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserPreference 用户偏好模型
type UserPreference struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID       string    `json:"user_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	Categories   string    `json:"categories" gorm:"type:text"`   // JSON存储偏好分类
	Schools      string    `json:"schools" gorm:"type:text"`      // JSON存储偏好学派
	Authors      string    `json:"authors" gorm:"type:text"`      // JSON存储偏好作者
	Tags         string    `json:"tags" gorm:"type:text"`         // JSON存储偏好标签
	Difficulty   string    `json:"difficulty" gorm:"type:varchar(50);default:'medium'"`
	ReadingSpeed float64   `json:"reading_speed" gorm:"default:1.0"` // 阅读速度倍数
	LastActive   time.Time `json:"last_active" gorm:"autoUpdateTime"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserInteraction 用户交互记录
type UserInteraction struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID       string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	TargetID     string    `json:"target_id" gorm:"type:varchar(255);not null"`     // 目标ID（智慧、用户等）
	TargetType   string    `json:"target_type" gorm:"type:varchar(50);not null"`    // 目标类型
	ActionType   string    `json:"action_type" gorm:"type:varchar(50);not null"`    // 交互类型
	ActionValue  float64   `json:"action_value" gorm:"default:1.0"`                 // 交互权重
	IsPositive   bool      `json:"is_positive" gorm:"default:true"`                 // 是否为正向交互
	SessionID    string    `json:"session_id" gorm:"type:varchar(255)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// RecommendationLog 推荐日志
type RecommendationLog struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID        string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	Algorithm     string    `json:"algorithm" gorm:"type:varchar(50);not null"`      // 推荐算法
	RequestParams string    `json:"request_params" gorm:"type:text"`                 // 请求参数（JSON）
	Results       string    `json:"results" gorm:"type:text"`                        // 推荐结果（JSON）
	ResultCount   int       `json:"result_count" gorm:"default:0"`                   // 结果数量
	ClickCount    int       `json:"click_count" gorm:"default:0"`                    // 点击数量
	CTR           float64   `json:"ctr" gorm:"default:0"`                            // 点击率
	SessionID     string    `json:"session_id" gorm:"type:varchar(255)"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserSimilarity 用户相似度
type UserSimilarity struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID1      string    `json:"user_id1" gorm:"type:varchar(255);not null;index"`
	UserID2      string    `json:"user_id2" gorm:"type:varchar(255);not null;index"`
	Similarity   float64   `json:"similarity" gorm:"not null"`                      // 相似度分数
	Algorithm    string    `json:"algorithm" gorm:"type:varchar(50);not null"`      // 计算算法
	CommonItems  int       `json:"common_items" gorm:"default:0"`                   // 共同项目数
	LastUpdated  time.Time `json:"last_updated" gorm:"autoUpdateTime"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ActionType 常量定义
const (
	ActionTypeView     = "view"
	ActionTypeLike     = "like"
	ActionTypeShare    = "share"
	ActionTypeComment  = "comment"
	ActionTypeFavorite = "favorite"
	ActionTypeSearch   = "search"
	ActionTypeDownload = "download"
)

// TargetType 常量定义
const (
	TargetTypeWisdom   = "wisdom"
	TargetTypeUser     = "user"
	TargetTypeCategory = "category"
	TargetTypeTag      = "tag"
)

// Algorithm 常量定义
const (
	AlgorithmContent       = "content"
	AlgorithmCollaborative = "collaborative"
	AlgorithmHybrid        = "hybrid"
	AlgorithmPopular       = "popular"
	AlgorithmTrending      = "trending"
)