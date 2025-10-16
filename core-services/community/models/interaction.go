package models

import (
	"time"

	"gorm.io/gorm"
)

// Like 
type Like struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	PostID    *string        `json:"post_id" gorm:"type:varchar(36);index"`
	CommentID *string        `json:"comment_id" gorm:"type:varchar(36);index"`
	Type      LikeType       `json:"type" gorm:"type:varchar(20);not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	//  - 
	User    *UserProfile `json:"user,omitempty" gorm:"-"`
	Post    *Post        `json:"post,omitempty" gorm:"-"`
	Comment *Comment     `json:"comment,omitempty" gorm:"-"`
}

// LikeType 
type LikeType string

const (
	LikeTypePost    LikeType = "post"    // 
	LikeTypeComment LikeType = "comment" // 
)

// Follow 
type Follow struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	FollowerID  string         `json:"follower_id" gorm:"type:varchar(36);not null;index"`  // ID
	FollowingID string         `json:"following_id" gorm:"type:varchar(36);not null;index"` // ID
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	//  - 
	Follower  *UserProfile `json:"follower,omitempty" gorm:"foreignKey:FollowerID;references:UserID"`
	Following *UserProfile `json:"following,omitempty" gorm:"foreignKey:FollowingID;references:UserID"`
}

// Bookmark 
type Bookmark struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
	PostID    string         `json:"post_id" gorm:"type:varchar(36);not null;index"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	//  - 
	User *UserProfile `json:"user,omitempty" gorm:"-"`
	Post *Post        `json:"post,omitempty" gorm:"-"`
}

// Report 
type Report struct {
	ID         string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ReporterID string         `json:"reporter_id" gorm:"type:varchar(36);not null;index"`
	PostID     *string        `json:"post_id" gorm:"type:varchar(36);index"`
	CommentID  *string        `json:"comment_id" gorm:"type:varchar(36);index"`
	UserID     *string        `json:"user_id" gorm:"type:varchar(36);index"` // 
	Type       ReportType     `json:"type" gorm:"type:varchar(20);not null;index"`
	Reason     string         `json:"reason" gorm:"type:varchar(100);not null"`
	Content    string         `json:"content" gorm:"type:text"`
	Status     ReportStatus   `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	HandlerID  *string        `json:"handler_id" gorm:"type:varchar(36);index"`
	HandledAt  *time.Time     `json:"handled_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	//  - 
	Reporter *UserProfile `json:"reporter,omitempty" gorm:"-"`
	Post     *Post        `json:"post,omitempty" gorm:"-"`
	Comment  *Comment     `json:"comment,omitempty" gorm:"-"`
	User     *UserProfile `json:"user,omitempty" gorm:"-"`
	Handler  *UserProfile `json:"handler,omitempty" gorm:"-"`
}

// ReportType 
type ReportType string

const (
	ReportTypePost    ReportType = "post"    // 
	ReportTypeComment ReportType = "comment" // 
	ReportTypeUser    ReportType = "user"    // 
)

// ReportStatus 
type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"  // 
	ReportStatusApproved ReportStatus = "approved" // 
	ReportStatusRejected ReportStatus = "rejected" // 
)

// LikeRequest 
type LikeRequest struct {
	PostID    *string `json:"post_id,omitempty"`
	CommentID *string `json:"comment_id,omitempty"`
}

// FollowRequest 
type FollowRequest struct {
	FollowingID string `json:"following_id" binding:"required"`
}

// BookmarkRequest 
type BookmarkRequest struct {
	PostID string `json:"post_id" binding:"required"`
}

// ReportRequest 
type ReportRequest struct {
	PostID    *string `json:"post_id,omitempty"`
	CommentID *string `json:"comment_id,omitempty"`
	UserID    *string `json:"user_id,omitempty"`
	Reason    string  `json:"reason" binding:"required,max=100"`
	Content   string  `json:"content" binding:"max=1000"`
}

// LikeResponse 
type LikeResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PostID    *string   `json:"post_id"`
	CommentID *string   `json:"comment_id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// FollowResponse 
type FollowResponse struct {
	ID          string            `json:"id"`
	FollowerID  string            `json:"follower_id"`
	FollowingID string            `json:"following_id"`
	Following   *UserProfileBrief `json:"following,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// BookmarkResponse 
type BookmarkResponse struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"`
	PostID    string        `json:"post_id"`
	Post      *PostResponse `json:"post,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// ReportResponse 
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

// InteractionStatsResponse 
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

// TableName 
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

// BeforeCreate UUID
func (l *Like) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		// UUIDUUID
		// l.ID = uuid.New().String()
	}
	return nil
}

func (f *Follow) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		// UUIDUUID
		// f.ID = uuid.New().String()
	}
	return nil
}

func (b *Bookmark) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		// UUIDUUID
		// b.ID = uuid.New().String()
	}
	return nil
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		// UUIDUUID
		// r.ID = uuid.New().String()
	}
	return nil
}

// ToResponse 
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

