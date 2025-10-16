package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenType 令牌类型枚举
type TokenType string

const (
	TokenTypeAccess       TokenType = "access"
	TokenTypeRefresh      TokenType = "refresh"
	TokenTypeReset        TokenType = "reset"
	TokenTypeVerification TokenType = "verification"
	TokenTypeInvitation   TokenType = "invitation"
)

// TokenStatus 令牌状态枚举?
type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusUsed    TokenStatus = "used"
	TokenStatusExpired TokenStatus = "expired"
	TokenStatusRevoked TokenStatus = "revoked"
)

// Token 令牌模型
type Token struct {
	ID        uuid.UUID   `json:"id" gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID   `json:"user_id" gorm:"type:char(36);not null;index"`
	Type      TokenType   `json:"type" gorm:"type:varchar(20);not null;index"`
	Token     string      `json:"token" gorm:"uniqueIndex;not null;size:255"`
	Status    TokenStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	Purpose   string      `json:"purpose" gorm:"size:100"`   // 令牌用户途描
	Metadata  string      `json:"metadata" gorm:"type:text"` // JSON格式的元数量据
	ExpiresAt time.Time   `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time  `json:"used_at"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Token) TableName() string {
	return "tokens"
}

// BeforeCreate GORM钩子：创建前
func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IsExpired 检查令牌是否过期
func (t *Token) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// IsActive 检查令牌是否激活?
func (t *Token) IsActive() bool {
	return t.Status == TokenStatusActive && !t.IsExpired()
}

// IsUsed 检查令牌是否已使用户
func (t *Token) IsUsed() bool {
	return t.Status == TokenStatusUsed
}

// Use 使用户令牌
func (t *Token) Use() {
	t.Status = TokenStatusUsed
	now := time.Now()
	t.UsedAt = &now
}

// Revoke 撤销令牌
func (t *Token) Revoke() {
	t.Status = TokenStatusRevoked
}

// Expire 使令牌过期
func (t *Token) Expire() {
	t.Status = TokenStatusExpired
}

// CanBeUsed 检查令牌是否可以使用?
func (t *Token) CanBeUsed() bool {
	return t.IsActive() && !t.IsUsed()
}

// Refresh 刷新令牌过期时间
func (t *Token) Refresh(duration time.Duration) {
	now := time.Now().UTC()
	t.ExpiresAt = now.Add(duration)
	t.UpdatedAt = now
}

// CreateTokenRequest 创建令牌请求
type CreateTokenRequest struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	Type      TokenType `json:"type" validate:"required,oneof=access refresh reset verification invitation"`
	Purpose   string    `json:"purpose" validate:"max=100"`
	Metadata  string    `json:"metadata"`
	ExpiresIn int64     `json:"expires_in" validate:"min=1"` // 过期时间（秒）
}

// TokenQuery 令牌查询参数量
type TokenQuery struct {
	UserID   uuid.UUID   `form:"user_id"`
	Type     TokenType   `form:"type"`
	Status   TokenStatus `form:"status"`
	Purpose  string      `form:"purpose"`
	Page     int         `form:"page" validate:"min=1"`
	PageSize int         `form:"page_size" validate:"min=1,max=100"`
	OrderBy  string      `form:"order_by" validate:"oneof=created_at updated_at expires_at used_at"`
	Order    string      `form:"order" validate:"oneof=asc desc"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
