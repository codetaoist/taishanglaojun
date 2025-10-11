package models

import (
	"time"

	"gorm.io/gorm"
)

// Post 帖子模型
type Post struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Title       string         `json:"title" gorm:"type:varchar(200);not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	AuthorID    string         `json:"author_id" gorm:"type:varchar(36);not null;index"`
	Category    string         `json:"category" gorm:"type:varchar(50);not null;index"`
	Tags        string         `json:"tags" gorm:"type:text"` // JSON数组字符�?
	Status      PostStatus     `json:"status" gorm:"type:varchar(20);default:'published';index"`
	ViewCount   int            `json:"view_count" gorm:"default:0"`
	LikeCount   int            `json:"like_count" gorm:"default:0"`
	CommentCount int           `json:"comment_count" gorm:"default:0"`
	IsSticky    bool           `json:"is_sticky" gorm:"default:false;index"`
	IsHot       bool           `json:"is_hot" gorm:"default:false;index"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 - 暂时移除外键约束，但保留关联查询
	Author   *UserProfile `json:"author,omitempty" gorm:"foreignKey:AuthorID;references:UserID"`
	Comments []Comment    `json:"comments,omitempty" gorm:"foreignKey:PostID"`
	Likes    []Like       `json:"likes,omitempty" gorm:"foreignKey:PostID"`
}

// PostStatus 帖子状�?
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"     // 草稿
	PostStatusPending   PostStatus = "pending"   // 待审�?
	PostStatusPublished PostStatus = "published" // 已发�?
	PostStatusRejected  PostStatus = "rejected"  // 已拒�?
	PostStatusHidden    PostStatus = "hidden"    // 隐藏
	PostStatusDeleted   PostStatus = "deleted"   // 已删�?
)

// PostCreateRequest 创建帖子请求
type PostCreateRequest struct {
	Title    string   `json:"title" binding:"required,min=5,max=200"`
	Content  string   `json:"content" binding:"required,min=20,max=10000"`
	Category string   `json:"category" binding:"required"`
	Tags     []string `json:"tags"`
}

// PostUpdateRequest 更新帖子请求
type PostUpdateRequest struct {
	Title    *string  `json:"title,omitempty" binding:"omitempty,min=5,max=200"`
	Content  *string  `json:"content,omitempty" binding:"omitempty,min=20,max=10000"`
	Category *string  `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Status   *string  `json:"status,omitempty"`
}

// PostListRequest 帖子列表请求
type PostListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Category string `form:"category"`
	Tag      string `form:"tag"`
	AuthorID string `form:"author_id"`
	Status   string `form:"status"`
	SortBy   string `form:"sort_by"` // latest, hot, likes, views
	Keyword  string `form:"keyword"`
}

// PostResponse 帖子响应
type PostResponse struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Content      string            `json:"content"`
	AuthorID     string            `json:"author_id"`
	Author       *UserProfileBrief `json:"author,omitempty"`
	Category     string            `json:"category"`
	Tags         []string          `json:"tags"`
	Status       string            `json:"status"`
	ViewCount    int               `json:"view_count"`
	LikeCount    int               `json:"like_count"`
	CommentCount int               `json:"comment_count"`
	IsSticky     bool              `json:"is_sticky"`
	IsHot        bool              `json:"is_hot"`
	IsLiked      bool              `json:"is_liked,omitempty"`      // 当前用户是否点赞
	IsBookmarked bool              `json:"is_bookmarked,omitempty"` // 当前用户是否收藏
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// PostDetailResponse 帖子详情响应
type PostDetailResponse struct {
	PostResponse
	Comments []CommentResponse `json:"comments,omitempty"`
}

// PostListResponse 帖子列表响应
type PostListResponse struct {
	Posts      []PostResponse `json:"posts"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// PostStatsResponse 帖子统计响应
type PostStatsResponse struct {
	TotalPosts     int64 `json:"total_posts"`
	TodayPosts     int64 `json:"today_posts"`
	WeeklyPosts    int64 `json:"weekly_posts"`
	MonthlyPosts   int64 `json:"monthly_posts"`
	TotalViews     int64 `json:"total_views"`
	TotalLikes     int64 `json:"total_likes"`
	TotalComments  int64 `json:"total_comments"`
	ActiveUsers    int64 `json:"active_users"`
	PopularTags    []TagStats `json:"popular_tags"`
	TopCategories  []CategoryStats `json:"top_categories"`
}

// TagStats 标签统计
type TagStats struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
}

// CategoryStats 分类统计
type CategoryStats struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// TableName 指定表名
func (Post) TableName() string {
	return "community_posts"
}

// BeforeCreate 创建前钩�?
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		// 这里可以使用UUID生成�?
		// p.ID = uuid.New().String()
	}
	return nil
}

// ToResponse 转换为响应格�?
func (p *Post) ToResponse() PostResponse {
	var tags []string
	if p.Tags != "" {
		// 这里需要解析JSON字符串为数组
		// 简化处理，实际应该使用JSON解析
		tags = []string{} // 临时空数�?
	}

	response := PostResponse{
		ID:           p.ID,
		Title:        p.Title,
		Content:      p.Content,
		AuthorID:     p.AuthorID,
		Category:     p.Category,
		Tags:         tags,
		Status:       string(p.Status),
		ViewCount:    p.ViewCount,
		LikeCount:    p.LikeCount,
		CommentCount: p.CommentCount,
		IsSticky:     p.IsSticky,
		IsHot:        p.IsHot,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}

	if p.Author != nil {
		response.Author = p.Author.ToBrief()
	}

	return response
}

// ToDetailResponse 转换为详情响应格�?
func (p *Post) ToDetailResponse() PostDetailResponse {
	response := PostDetailResponse{
		PostResponse: p.ToResponse(),
	}

	if len(p.Comments) > 0 {
		response.Comments = make([]CommentResponse, len(p.Comments))
		for i, comment := range p.Comments {
			response.Comments[i] = comment.ToResponse()
		}
	}

	return response
}
