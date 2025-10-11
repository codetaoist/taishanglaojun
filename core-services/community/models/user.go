package models

import (
	"time"

	"gorm.io/gorm"
)

// UserProfile 用户资料模型
type UserProfile struct {
	UserID      string         `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	Username    string         `json:"username" gorm:"type:varchar(50);not null;uniqueIndex"`
	Nickname    string         `json:"nickname" gorm:"type:varchar(100);not null"`
	Avatar      string         `json:"avatar" gorm:"type:varchar(500)"`
	Bio         string         `json:"bio" gorm:"type:text"`
	Location    string         `json:"location" gorm:"type:varchar(100)"`
	Website     string         `json:"website" gorm:"type:varchar(200)"`
	PostCount   int            `json:"post_count" gorm:"default:0"`
	FollowerCount int          `json:"follower_count" gorm:"default:0"`
	FollowingCount int         `json:"following_count" gorm:"default:0"`
	LikeCount   int            `json:"like_count" gorm:"default:0"`
	Level       int            `json:"level" gorm:"default:1"`
	Experience  int            `json:"experience" gorm:"default:0"`
	Status      UserStatus     `json:"status" gorm:"type:varchar(20);default:'active'"`
	LastActiveAt *time.Time    `json:"last_active_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Posts     []Post   `json:"posts,omitempty" gorm:"foreignKey:AuthorID;references:UserID"`
	Comments  []Comment `json:"comments,omitempty" gorm:"foreignKey:AuthorID;references:UserID"`
	Likes     []Like   `json:"likes,omitempty" gorm:"foreignKey:UserID"`
	Followers []Follow `json:"followers,omitempty" gorm:"foreignKey:FollowingID;references:UserID"`
	Following []Follow `json:"following,omitempty" gorm:"foreignKey:FollowerID;references:UserID"`
}

// UserStatus 用户状�?
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 活跃
	UserStatusInactive UserStatus = "inactive" // 不活�?
	UserStatusBanned   UserStatus = "banned"   // 被封�?
	UserStatusDeleted  UserStatus = "deleted"  // 已删�?
)

// UserProfileUpdateRequest 更新用户资料请求
type UserProfileUpdateRequest struct {
	Nickname *string `json:"nickname,omitempty" binding:"omitempty,min=1,max=100"`
	Avatar   *string `json:"avatar,omitempty"`
	Bio      *string `json:"bio,omitempty" binding:"omitempty,max=500"`
	Location *string `json:"location,omitempty" binding:"omitempty,max=100"`
	Website  *string `json:"website,omitempty" binding:"omitempty,url,max=200"`
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	Nickname       string    `json:"nickname"`
	Avatar         string    `json:"avatar"`
	Bio            string    `json:"bio"`
	Location       string    `json:"location"`
	Website        string    `json:"website"`
	PostCount      int       `json:"post_count"`
	FollowerCount  int       `json:"follower_count"`
	FollowingCount int       `json:"following_count"`
	LikeCount      int       `json:"like_count"`
	Level          int       `json:"level"`
	Experience     int       `json:"experience"`
	Status         string    `json:"status"`
	IsFollowing    bool      `json:"is_following,omitempty"`    // 当前用户是否关注此用�?
	IsFollowedBy   bool      `json:"is_followed_by,omitempty"`  // 此用户是否关注当前用�?
	LastActiveAt   *time.Time `json:"last_active_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// UserProfileBrief 用户资料简要信�?
type UserProfileBrief struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Level    int    `json:"level"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Keyword  string `form:"keyword"`
	SortBy   string `form:"sort_by"` // latest, posts, followers, level
	Status   string `form:"status"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users      []UserProfileResponse `json:"users"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// UserStatsResponse 用户统计响应
type UserStatsResponse struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	NewUsers      int64 `json:"new_users"`      // 今日新用�?
	WeeklyUsers   int64 `json:"weekly_users"`   // 本周新用�?
	MonthlyUsers  int64 `json:"monthly_users"`  // 本月新用�?
	OnlineUsers   int64 `json:"online_users"`   // 在线用户
	TopUsers      []UserProfileBrief `json:"top_users"` // 活跃用户
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "community_user_profiles"
}

// ToResponse 转换为响应格�?
func (u *UserProfile) ToResponse() UserProfileResponse {
	return UserProfileResponse{
		UserID:         u.UserID,
		Username:       u.Username,
		Nickname:       u.Nickname,
		Avatar:         u.Avatar,
		Bio:            u.Bio,
		Location:       u.Location,
		Website:        u.Website,
		PostCount:      u.PostCount,
		FollowerCount:  u.FollowerCount,
		FollowingCount: u.FollowingCount,
		LikeCount:      u.LikeCount,
		Level:          u.Level,
		Experience:     u.Experience,
		Status:         string(u.Status),
		LastActiveAt:   u.LastActiveAt,
		CreatedAt:      u.CreatedAt,
	}
}

// ToBrief 转换为简要信息格�?
func (u *UserProfile) ToBrief() *UserProfileBrief {
	return &UserProfileBrief{
		UserID:   u.UserID,
		Username: u.Username,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Level:    u.Level,
	}
}

// IsActive 判断用户是否活跃
func (u *UserProfile) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanPost 判断用户是否可以发帖
func (u *UserProfile) CanPost() bool {
	return u.Status == UserStatusActive
}

// CanComment 判断用户是否可以评论
func (u *UserProfile) CanComment() bool {
	return u.Status == UserStatusActive
}

// AddExperience 增加经验�?
func (u *UserProfile) AddExperience(exp int) {
	u.Experience += exp
	// 简单的等级计算逻辑
	newLevel := u.Experience/1000 + 1
	if newLevel > u.Level {
		u.Level = newLevel
	}
}

// UpdateLastActive 更新最后活跃时�?
func (u *UserProfile) UpdateLastActive() {
	now := time.Now()
	u.LastActiveAt = &now
}
