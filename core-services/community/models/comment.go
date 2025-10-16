package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment 
type Comment struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	PostID    string         `json:"post_id" gorm:"type:varchar(36);not null;index"`
	AuthorID  string         `json:"author_id" gorm:"type:varchar(36);not null;index"`
	ParentID  *string        `json:"parent_id" gorm:"type:varchar(36);index"` // ID
	Content   string         `json:"content" gorm:"type:text;not null"`
	LikeCount int            `json:"like_count" gorm:"default:0"`
	Status    CommentStatus  `json:"status" gorm:"type:varchar(20);default:'published';index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	//  - 
	Post    *Post        `json:"post,omitempty" gorm:"-"`
	Author  *UserProfile `json:"author,omitempty" gorm:"foreignKey:AuthorID;references:UserID"`
	Parent  *Comment     `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies []Comment    `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
	Likes   []Like       `json:"likes,omitempty" gorm:"foreignKey:CommentID"`
}

// CommentStatus 
type CommentStatus string

const (
	CommentStatusPending   CommentStatus = "pending"   // 
	CommentStatusPublished CommentStatus = "published" // 
	CommentStatusRejected  CommentStatus = "rejected"  // 
	CommentStatusHidden    CommentStatus = "hidden"    // 
	CommentStatusDeleted   CommentStatus = "deleted"   // 
)

// CommentCreateRequest 
type CommentCreateRequest struct {
	PostID   string  `json:"post_id" binding:"required"`
	ParentID *string `json:"parent_id,omitempty"`
	Content  string  `json:"content" binding:"required,min=1,max=1000"`
}

// CommentUpdateRequest 
type CommentUpdateRequest struct {
	Content *string `json:"content,omitempty" binding:"omitempty,min=1,max=1000"`
	Status  *string `json:"status,omitempty"`
}

// CommentListRequest 
type CommentListRequest struct {
	PostID   string `form:"post_id" binding:"required"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	SortBy   string `form:"sort_by"` // latest, oldest, likes
}

// CommentResponse 
type CommentResponse struct {
	ID        string            `json:"id"`
	PostID    string            `json:"post_id"`
	AuthorID  string            `json:"author_id"`
	Author    *UserProfileBrief `json:"author,omitempty"`
	ParentID  *string           `json:"parent_id"`
	Content   string            `json:"content"`
	LikeCount int               `json:"like_count"`
	Status    string            `json:"status"`
	IsLiked   bool              `json:"is_liked,omitempty"` // 
	Replies   []CommentResponse `json:"replies,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CommentListResponse 
type CommentListResponse struct {
	Comments   []CommentResponse `json:"comments"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// CommentStatsResponse 
type CommentStatsResponse struct {
	TotalComments   int64 `json:"total_comments"`
	TodayComments   int64 `json:"today_comments"`
	WeeklyComments  int64 `json:"weekly_comments"`
	MonthlyComments int64 `json:"monthly_comments"`
	ActiveUsers     int64 `json:"active_users"`
}

// TableName 
func (Comment) TableName() string {
	return "community_comments"
}

// BeforeCreate UUID
func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		// UUIDUUID
		// c.ID = uuid.New().String()
	}
	return nil
}

// ToResponse 
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

// IsReply 
func (c *Comment) IsReply() bool {
	return c.ParentID != nil
}

// GetDepth 㼶
func (c *Comment) GetDepth() int {
	if c.ParentID == nil {
		return 0
	}
	// 
	// 1
	return 1
}

