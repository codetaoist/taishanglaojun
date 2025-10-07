package models

import (
	"time"

	"gorm.io/gorm"
)

// Like 点赞模型
type Like struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	PostID    *string        `json:"post_id" gorm:"type:varchar(36);index"`
	CommentID *string        `json:"comment_id" gorm:"type:varchar(36);index"`
	Type      LikeType       `json:"type" gorm:"type:varchar(20);not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 - 暂时移除外键约束
	User    *UserProfile `json:"user,omitempty" gorm:"-"`
	Post    *Post        `json:"post,omitempty" gorm:"-"`
	Comment *Comment     `json:"comment,omitempty" gorm:"-"`
}

// LikeType 点赞类型
type LikeType string

const (
	LikeTypePost    LikeType = "post"    // 帖子点赞
	LikeTypeComment LikeType = "comment" // 评论点赞
)

// Follow 关注模型
type Follow struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	FollowerID  string         `json:"follower_id" gorm:"type:varchar(36);not null;index"`  // 关注者ID
	FollowingID string         `json:"following_id" gorm:"type:varchar(36);not null;index"` // 被关注者ID
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 - 暂时移除外键约束
	Follower  *UserProfile `json:"follower,omitempty" gorm:"foreignKey:FollowerID;references:UserID"`
	Following *UserProfile `json:"following,omitempty" gorm:"foreignKey:FollowingID;references:UserID"`
}

// Bookmark 收藏模型
type Bookmark struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	PostID    string         `json:"post_id" gorm:"type:varchar(36);not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 - 暂时移除外键约束
	User *UserProfile `json:"user,omitempty" gorm:"-"`
	Post *Post        `json:"post,omitempty" gorm:"-"`
}

// Report 举报模型
type Report struct {
	ID         string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ReporterID string         `json:"reporter_id" gorm:"type:varchar(36);not null;index"`
	PostID     *string        `json:"post_id" gorm:"type:varchar(36);index"`
	CommentID  *string        `json:"comment_id" gorm:"type:varchar(36);index"`
	UserID     *string        `json:"user_id" gorm:"type:varchar(36);index"` // 被举报的用户
	Type       ReportType     `json:"type" gorm:"type:varchar(20);not null;index"`
	Reason     string         `json:"reason" gorm:"type:varchar(100);not null"`
	Content    string         `json:"content" gorm:"type:text"`
	Status     ReportStatus   `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	HandlerID  *string        `json:"handler_id" gorm:"type:varchar(36);index"`
	HandledAt  *time.Time     `json:"handled_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系 - 暂时移除外键约束
	Reporter *UserProfile `json:"reporter,omitempty" gorm:"-"`
	Post     *Post        `json:"post,omitempty" gorm:"-"`
	Comment  *Comment     `json:"comment,omitempty" gorm:"-"`
	User     *UserProfile `json:"user,omitempty" gorm:"-"`
	Handler  *UserProfile `json:"handler,omitempty" gorm:"-"`
}

// ReportType 举报类型
type ReportType string

const (
	ReportTypePost    ReportType = "post"    // 举报帖子
	ReportTypeComment ReportType = "comment" // 举报评论
	ReportTypeUser    ReportType = "user"    // 举报用户
)

// ReportStatus 举报状态
type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"  // 待处理
	ReportStatusApproved ReportStatus = "approved" // 已通过
	ReportStatusRejected ReportStatus = "rejected" // 已拒绝
)

// LikeRequest 点赞请求
type LikeRequest struct {
	PostID    *string `json:"post_id,omitempty"`
	CommentID *string `json:"comment_id,omitempty"`
}

// FollowRequest 关注请求
type FollowRequest struct {
	FollowingID string `json:"following_id" binding:"required"`
}

// BookmarkRequest 收藏请求
type BookmarkRequest struct {
	PostID string `json:"post_id" binding:"required"`
}

// ReportRequest 举报请求
type ReportRequest struct {
	PostID    *string `json:"post_id,omitempty"`
	CommentID *string `json:"comment_id,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	Reason    string  `json:"reason" binding:"required,max=100"`
	Content   string  `json:"content" binding:"max=1000"`
}

// LikeResponse 点赞响应
type LikeResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PostID    *string   `json:"post_id"`
	CommentID *string   `json:"comment_id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// FollowResponse 关注响应
type FollowResponse struct {
	ID          string            `json:"id"`
	FollowerID  string            `json:"follower_id"`
	FollowingID string            `json:"following_id"`
	Following   *UserProfileBrief `json:"following,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// BookmarkResponse 收藏响应
type BookmarkResponse struct {
	ID        string       `json:"id"`
	UserID    string       `json:"user_id"`
	PostID    string       `json:"post_id"`
	Post      *PostResponse `json:"post,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

// ReportResponse 举报响应
type ReportResponse struct {
	ID         string            `json:"id"`
	ReporterID string            `json:"reporter_id"`
	Reporter   *UserProfileBrief `json:"reporter,omitempty"`
	PostID     *string           `json:"post_id"`
	CommentID  *string           `json:"comment_id"`
	UserID     *string           `json:"user_id"`
	Type       string            `json:"type"`
	Reason     string            `json:"reason"`
	Content    string            `json:"content"`
	Status     string            `json:"status"`
	HandlerID  *string           `json:"handler_id"`
	Handler    *UserProfileBrief `json:"handler,omitempty"`
	HandledAt  *time.Time        `json:"handled_at"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// InteractionStatsResponse 互动统计响应
type InteractionStatsResponse struct {
	TotalLikes     int64 `json:"total_likes"`
	TotalFollows   int64 `json:"total_follows"`
	TotalBookmarks int64 `json:"total_bookmarks"`
	TotalReports   int64 `json:"total_reports"`
	TodayLikes     int64 `json:"today_likes"`
	TodayFollows   int64 `json:"today_follows"`
	TodayBookmarks int64 `json:"today_bookmarks"`
	TodayReports   int64 `json:"today_reports"`
}

// TableName 指定表名
func (Like) TableName() string {
	return "community_likes"
}

func (Follow) TableName() string {
	return "community_follows"
}

func (Bookmark) TableName() string {
	return "community_bookmarks"
}

func (Report) TableName() string {
	return "community_reports"
}

// BeforeCreate 创建前钩子
func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		// 这里可以使用UUID生成器
		// l.ID = uuid.New().String()
	}
	return nil
}

func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		// 这里可以使用UUID生成器
		// f.ID = uuid.New().String()
	}
	return nil
}

func (b *Bookmark) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		// 这里可以使用UUID生成器
		// b.ID = uuid.New().String()
	}
	return nil
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		// 这里可以使用UUID生成器
		// r.ID = uuid.New().String()
	}
	return nil
}

// ToResponse 转换为响应格式
func (l *Like) ToResponse() LikeResponse {
	return LikeResponse{
		ID:        l.ID,
		UserID:    l.UserID,
		PostID:    l.PostID,
		CommentID: l.CommentID,
		Type:      string(l.Type),
		CreatedAt: l.CreatedAt,
	}
}

func (f *Follow) ToResponse() FollowResponse {
	response := FollowResponse{
		ID:          f.ID,
		FollowerID:  f.FollowerID,
		FollowingID: f.FollowingID,
		CreatedAt:   f.CreatedAt,
	}

	if f.Following != nil {
		response.Following = f.Following.ToBrief()
	}

	return response
}

func (b *Bookmark) ToResponse() BookmarkResponse {
	response := BookmarkResponse{
		ID:        b.ID,
		UserID:    b.UserID,
		PostID:    b.PostID,
		CreatedAt: b.CreatedAt,
	}

	if b.Post != nil {
		postResponse := b.Post.ToResponse()
		response.Post = &postResponse
	}

	return response
}

func (r *Report) ToResponse() ReportResponse {
	response := ReportResponse{
		ID:         r.ID,
		ReporterID: r.ReporterID,
		PostID:     r.PostID,
		CommentID:  r.CommentID,
		UserID:     r.UserID,
		Type:       string(r.Type),
		Reason:     r.Reason,
		Content:    r.Content,
		Status:     string(r.Status),
		HandlerID:  r.HandlerID,
		HandledAt:  r.HandledAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}

	if r.Reporter != nil {
		response.Reporter = r.Reporter.ToBrief()
	}

	if r.Handler != nil {
		response.Handler = r.Handler.ToBrief()
	}

	return response
}