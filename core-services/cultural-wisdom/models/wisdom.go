package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice 自定义字符串切片类型，用于JSON序列�?
type StringSlice []string

// Value 实现 driver.Valuer 接口
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan 实现 sql.Scanner 接口
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}
	
	if len(data) == 0 || string(data) == "null" {
		*s = StringSlice{}
		return nil
	}
	
	return json.Unmarshal(data, s)
}

// CulturalWisdom 文化智慧内容模型
type CulturalWisdom struct {
	ID          string      `json:"id" gorm:"primaryKey;type:varchar(255)" bson:"_id"`
	Title       string      `json:"title" gorm:"type:varchar(500);not null" bson:"title"`
	Content     string      `json:"content" gorm:"type:text;not null" bson:"content"`
	Summary     string      `json:"summary" gorm:"type:text" bson:"summary"`
	Author      string      `json:"author" gorm:"type:varchar(255);not null" bson:"author"`
	AuthorID    string      `json:"author_id" gorm:"type:varchar(255)" bson:"author_id"`
	Category    string      `json:"category" gorm:"type:varchar(100)" bson:"category"`
	School      string      `json:"school" gorm:"type:varchar(100)" bson:"school"`
	Tags        StringSlice `json:"tags" gorm:"type:text" bson:"tags"` // JSON存储
	Vector      []float32   `json:"vector" gorm:"type:text" bson:"vector"` // 向量表示
	Difficulty  string      `json:"difficulty" gorm:"type:varchar(50);default:'medium'" bson:"difficulty"`
	Status      string      `json:"status" gorm:"type:varchar(50);default:'published'" bson:"status"`
	ViewCount   int64       `json:"view_count" gorm:"default:0" bson:"view_count"`
	LikeCount   int64       `json:"like_count" gorm:"default:0" bson:"like_count"`
	ShareCount  int64       `json:"share_count" gorm:"default:0" bson:"share_count"`
	CommentCount int64      `json:"comment_count" gorm:"default:0" bson:"comment_count"`
	IsFeatured  bool        `json:"is_featured" gorm:"default:false" bson:"is_featured"`
	IsRecommended bool      `json:"is_recommended" gorm:"default:false" bson:"is_recommended"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime" bson:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime" bson:"updated_at"`
	PublishedAt *time.Time  `json:"published_at" gorm:"type:timestamp" bson:"published_at"`
	DeletedAt   *time.Time  `json:"deleted_at" gorm:"type:timestamp" bson:"deleted_at"`
	Metadata    string      `json:"metadata" gorm:"type:text" bson:"metadata"` // JSON存储
}

// Category 分类模型
type Category struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement" bson:"_id"`
	Name        string `json:"name" gorm:"type:varchar(100);uniqueIndex;not null" bson:"name"`
	Description string `json:"description" gorm:"type:text" bson:"description"`
	ParentID    *int   `json:"parent_id" gorm:"type:int" bson:"parent_id"`
	SortOrder   int    `json:"sort_order" gorm:"default:0" bson:"sort_order"`
	IsActive    bool   `json:"is_active" gorm:"default:true" bson:"is_active"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" bson:"updated_at"`
}

// WisdomTag 智慧标签模型
type WisdomTag struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Color       string    `json:"color" gorm:"type:varchar(20);default:'#007bff'"`
	UsageCount  int       `json:"usage_count" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// WisdomSchool 智慧学派
type WisdomSchool struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string    `json:"description" gorm:"type:text"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (WisdomSchool) TableName() string {
	return "wisdom_schools"
}

// WisdomTagRelation 智慧内容与标签关联模�?
type WisdomTagRelation struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	WisdomID string `json:"wisdom_id" gorm:"type:varchar(255);not null"`
	TagID    int    `json:"tag_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
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
	Page        string `json:"page" bson:"page"`               // 页码或章�?
}

// WisdomMetadata 智慧内容元数�?
type WisdomMetadata struct {
	ID           int               `json:"id" gorm:"primaryKey;autoIncrement"`
	WisdomID     string            `json:"wisdom_id" gorm:"type:varchar(255);uniqueIndex"`
	Language     string            `json:"language" gorm:"type:varchar(10);default:'zh'"`
	Keywords     string            `json:"keywords" gorm:"type:text"` // JSON存储
	ReadingTime  int               `json:"reading_time" gorm:"default:0"` // 预估阅读时间（分钟）
	WordCount    int               `json:"word_count" gorm:"default:0"`
	Translations string            `json:"translations" gorm:"type:text"` // JSON存储多语言翻译
	RelatedIDs   string            `json:"related_ids" gorm:"type:text"`   // JSON存储相关内容ID
	CustomFields string            `json:"custom_fields" gorm:"type:text"` // JSON存储自定义字�?
	CreatedAt    time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
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

// Wisdom 智慧内容简化模型（用于搜索结果�?
type Wisdom struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Author   string   `json:"author"`
	Source   string   `json:"source"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Wisdoms  []Wisdom `json:"wisdoms"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

// SearchResultWithFacets 带分面的搜索结果
type SearchResultWithFacets struct {
	SearchResult
	Facets map[string]interface{} `json:"facets"`
}

// WisdomContent 智慧内容详细信息（用于详情显示和创建/更新�?
type WisdomContent struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Summary  string   `json:"summary"`
	Category Category `json:"category"`
	Tags     []string `json:"tags"`
	Source   Source   `json:"source"`
	Difficulty int    `json:"difficulty"`
	Status   string   `json:"status"`
	AuthorID string   `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ViewCount int64    `json:"view_count"`
	LikeCount int64    `json:"like_count"`
	Metadata WisdomMetadata `json:"metadata"`
}

// CreateWisdomRequest 创建智慧内容请求
type CreateWisdomRequest struct {
	Title      string   `json:"title" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Summary    string   `json:"summary"`
	CategoryID string   `json:"category_id"`
	Tags       []string `json:"tags"`
	Source     Source   `json:"source"`
	Difficulty int      `json:"difficulty"`
	Status     string   `json:"status"`
	Metadata   WisdomMetadata `json:"metadata"`
}

// UpdateWisdomRequest 更新智慧内容请求
type UpdateWisdomRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Summary    string   `json:"summary"`
	CategoryID string   `json:"category_id"`
	Tags       []string `json:"tags"`
	Source     Source   `json:"source"`
	Difficulty int      `json:"difficulty"`
	Status     string   `json:"status"`
	Metadata   WisdomMetadata `json:"metadata"`
}


// CategoryStats 分类统计
type CategoryStats struct {
	CategoryID     int   `json:"category_id"`
	TotalCount     int64 `json:"total_count"`
	PublishedCount int64 `json:"published_count"`
	DraftCount     int64 `json:"draft_count"`
	TotalViews     int64 `json:"total_views"`
	TotalLikes     int64 `json:"total_likes"`
}

// TagStats 标签统计信息
type TagStats struct {
	TagID       int    `json:"tag_id"`
	TagName     string `json:"tag_name"`
	UsageCount  int    `json:"usage_count"`
	WisdomCount int64  `json:"wisdom_count"`
	TotalViews  int64  `json:"total_views"`
	TotalLikes  int64  `json:"total_likes"`
}

// CategoryNode 分类树节�?
type CategoryNode struct {
	Category
	Children []CategoryNode `json:"children"`
}

