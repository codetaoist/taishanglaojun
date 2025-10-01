package models

import (
	"time"
)

// CulturalWisdom 文化智慧内容模型
type CulturalWisdom struct {
	ID          string    `json:"id" bson:"_id"`
	Title       string    `json:"title" bson:"title"`
	Content     string    `json:"content" bson:"content"`
	Summary     string    `json:"summary" bson:"summary"`
	Category    Category  `json:"category" bson:"category"`
	Tags        []string  `json:"tags" bson:"tags"`
	Source      Source    `json:"source" bson:"source"`
	Difficulty  int       `json:"difficulty" bson:"difficulty"` // 1-9 难度等级
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	ViewCount   int64     `json:"view_count" bson:"view_count"`
	LikeCount   int64     `json:"like_count" bson:"like_count"`
	Status      string    `json:"status" bson:"status"` // draft, published, archived
	AuthorID    string    `json:"author_id" bson:"author_id"`
	Metadata    WisdomMetadata `json:"metadata" bson:"metadata"`
}

// Category 分类模型
type Category struct {
	ID          string `json:"id" bson:"_id"`
	Name        string `json:"name" bson:"name"`
	School      string `json:"school" bson:"school"`         // �?�?�?�?	ParentID    string `json:"parent_id" bson:"parent_id"`
	Level       int    `json:"level" bson:"level"`
	Description string `json:"description" bson:"description"`
	Icon        string `json:"icon" bson:"icon"`
	Color       string `json:"color" bson:"color"`
	SortOrder   int    `json:"sort_order" bson:"sort_order"`
	IsActive    bool   `json:"is_active" bson:"is_active"`
}

// Source 来源信息
type Source struct {
	Type        string `json:"type" bson:"type"`               // book, article, speech, etc.
	Title       string `json:"title" bson:"title"`
	Author      string `json:"author" bson:"author"`
	Dynasty     string `json:"dynasty" bson:"dynasty"`         // 朝代
	Publisher   string `json:"publisher" bson:"publisher"`
	PublishDate string `json:"publish_date" bson:"publish_date"`
	ISBN        string `json:"isbn" bson:"isbn"`
	URL         string `json:"url" bson:"url"`
	Page        string `json:"page" bson:"page"`               // 页码或章�?}

// WisdomMetadata 智慧内容元数�?type WisdomMetadata struct {
	Language     string            `json:"language" bson:"language"`
	Keywords     []string          `json:"keywords" bson:"keywords"`
	ReadingTime  int               `json:"reading_time" bson:"reading_time"` // 预估阅读时间（分钟）
	WordCount    int               `json:"word_count" bson:"word_count"`
	Translations map[string]string `json:"translations" bson:"translations"` // 多语言翻译
	RelatedIDs   []string          `json:"related_ids" bson:"related_ids"`   // 相关内容ID
	CustomFields map[string]interface{} `json:"custom_fields" bson:"custom_fields"`
}

// WisdomSummary 智慧内容摘要（用于列表显示）
type WisdomSummary struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Summary    string   `json:"summary"`
	Category   Category `json:"category"`
	Tags       []string `json:"tags"`
	Difficulty int      `json:"difficulty"`
	ViewCount  int64    `json:"view_count"`
	LikeCount  int64    `json:"like_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// WisdomFilter 智慧内容过滤条件
type WisdomFilter struct {
	CategoryID   string   `json:"category_id"`
	School       string   `json:"school"`
	Tags         []string `json:"tags"`
	Difficulty   []int    `json:"difficulty"`
	AuthorID     string   `json:"author_id"`
	Status       string   `json:"status"`
	Language     string   `json:"language"`
	DateFrom     *time.Time `json:"date_from"`
	DateTo       *time.Time `json:"date_to"`
	SearchQuery  string   `json:"search_query"`
	SortBy       string   `json:"sort_by"`       // created_at, updated_at, view_count, like_count
	SortOrder    string   `json:"sort_order"`    // asc, desc
	Page         int      `json:"page"`
	Size         int      `json:"size"`
}

// WisdomStats 智慧内容统计
type WisdomStats struct {
	TotalCount     int64            `json:"total_count"`
	PublishedCount int64            `json:"published_count"`
	DraftCount     int64            `json:"draft_count"`
	CategoryStats  map[string]int64 `json:"category_stats"`
	SchoolStats    map[string]int64 `json:"school_stats"`
	DifficultyStats map[int]int64   `json:"difficulty_stats"`
	MonthlyStats   []MonthlyCount   `json:"monthly_stats"`
}

// MonthlyCount 月度统计
type MonthlyCount struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Count int64 `json:"count"`
}
