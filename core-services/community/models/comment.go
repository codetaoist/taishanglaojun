package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment 评论模型
type Comment struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PostID    string         `json:"post_id" gorm:"type:varchar(36);not null;index"`
	AuthorID  string         `json:"author_id" gorm:"type:varchar(36);not null;index"`
	ParentID  *string        `json:"parent_id" gorm:"type:varchar(36);index"` // 父评论ID，用于回�?
	Content   string         `json:"content" gorm:"type:text;not null"`
	LikeCount int            `json:"like_count" gorm:"default:0"`
	Status    CommentStatus  `json:"status" gorm:"type:varchar(20);default:'published';index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 - 暂时移除外键约束
	Post     *Post        `json:"post,omitempty" gorm:"-"`
	Author   *UserProfile `json:"author,omitempty" gorm:"foreignKey:AuthorID;references:UserID"`
	Parent   *Comment     `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies  []Comment    `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
	Likes    []Like       `json:"likes,omitempty" gorm:"foreignKey:CommentID"`
}

// CommentStatus 评论状�?
type CommentStatus string

const (
	CommentStatusPending   CommentStatus = "pending"   // 待审�?
	CommentStatusPublished CommentStatus = "published" // 已发�?
	CommentStatusRejected  CommentStatus = "rejected"  // 已拒�?
	CommentStatusHidden    CommentStatus = "hidden"    // 隐藏
	CommentStatusDeleted   CommentStatus = "deleted"   // 已删�?
)

// CommentCreateRequest 创建评论请求
type CommentCreateRequest struct {
	PostID   string  `json:"post_id" binding:"required"`
	ParentID *string `json:"parent_id,omitempty"`
	Content  string  `json:"content" binding:"required,min=1,max=1000"`
}

// CommentUpdateRequest 更新评论请求
type CommentUpdateRequest struct {
	Content *string `json:"content,omitempty" binding:"omitempty,min=1,max=1000"`
	Status  *string `json:"status,omitempty"`
}

// CommentListRequest 评论列表请求
type CommentListRequest struct {
	PostID   string `form:"post_id" binding:"required"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	SortBy   string `form:"sort_by"` // latest, oldest, likes
}

// CommentResponse 评论响应
type CommentResponse struct {
	ID        string            `json:"id"`
	PostID    string            `json:"post_id"`
	AuthorID  string            `json:"author_id"`
	Author    *UserProfileBrief `json:"author,omitempty"`
	ParentID  *string           `json:"parent_id"`
	Content   string            `json:"content"`
	LikeCount int               `json:"like_count"`
	Status    string            `json:"status"`
	IsLiked   bool              `json:"is_liked,omitempty"` // 当前用户是否点赞
	Replies   []CommentResponse `json:"replies,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// CommentStatsResponse 评论统计响应
type CommentStatsResponse struct {
	TotalComments   int64 `json:"total_comments"`
	TodayComments   int64 `json:"today_comments"`
	WeeklyComments  int64 `json:"weekly_comments"`
	MonthlyComments int64 `json:"monthly_comments"`
	ActiveUsers     int64 `json:"active_users"`
}

// TableName 指定表名
func (Comment) TableName() string {
	return "community_comments"
}

// BeforeCreate 创建前钩�?
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		// 这里可以使用UUID生成�?
		// c.ID = uuid.New().String()
	}
	return nil
}

// ToResponse 转换为响应格�?
func (c *Comment) ToResponse() CommentResponse {
	response := CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		AuthorID:  c.AuthorID,
		ParentID:  c.ParentID,
		Content:   c.Content,
		LikeCount: c.LikeCount,
		Status:    string(c.Status),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	if c.Author != nil {
		response.Author = c.Author.ToBrief()
	}

	if len(c.Replies) > 0 {
		response.Replies = make([]CommentResponse, len(c.Replies))
		for i, reply := range c.Replies {
			response.Replies[i] = reply.ToResponse()
		}
	}

	return response
}

// IsReply 判断是否为回复评�?
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// GetDepth 获取评论层级深度
func (c *Comment) GetDepth() int {
	if c.ParentID == nil {
		return 0
	}
	// 这里需要递归查询父评论来计算深度
	// 简化处理，返回1表示是回�?
	return 1
}
