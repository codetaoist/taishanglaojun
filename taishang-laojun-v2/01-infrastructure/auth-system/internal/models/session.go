package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionStatus 会话状态枚举
type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

// Session 会话模型
type Session struct {
	ID        uuid.UUID     `json:"id" gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID     `json:"user_id" gorm:"type:char(36);not null;index"`
	Token     string        `json:"token" gorm:"size:255"`
	Status    SessionStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	UserAgent string        `json:"user_agent" gorm:"size:500"`
	IPAddress string        `json:"ip_address" gorm:"size:45"`
	ExpiresAt time.Time     `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`

	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// BeforeCreate GORM钩子：创建前
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	// 如果Token为空，生成一个随机Token
	if s.Token == "" {
		s.Token = uuid.New().String()
	}
	return nil
}

// IsExpired 检查会话是否过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive 检查会话是否激活
func (s *Session) IsActive() bool {
	return s.Status == SessionStatusActive && !s.IsExpired()
}

// Revoke 撤销会话
func (s *Session) Revoke() {
	s.Status = SessionStatusRevoked
}

// Expire 使会话过期
func (s *Session) Expire() {
	s.Status = SessionStatusExpired
}

// Refresh 刷新会话过期时间
func (s *Session) Refresh(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
	s.Status = SessionStatusActive
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	ExpiresIn int64     `json:"expires_in" validate:"min=1"` // 过期时间（秒）
}

// SessionQuery 会话查询参数
type SessionQuery struct {
	UserID    uuid.UUID     `form:"user_id"`
	Status    SessionStatus `form:"status"`
	IPAddress string        `form:"ip_address"`
	Page      int           `form:"page" validate:"min=1"`
	PageSize  int           `form:"page_size" validate:"min=1,max=100"`
	OrderBy   string        `form:"order_by" validate:"oneof=created_at updated_at expires_at"`
	Order     string        `form:"order" validate:"oneof=asc desc"`
}