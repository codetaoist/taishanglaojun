package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Remember  bool   `json:"remember"` // 是否记住登录状态
	UserAgent string `json:"-"`       // 从请求头获取
	IPAddress string `json:"-"`       // 从请求获取
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         *PublicUser    `json:"user"`
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token,omitempty"`
	TokenType    string         `json:"token_type"`
	ExpiresIn    int64          `json:"expires_in"`
	ExpiresAt    time.Time      `json:"expires_at"`
	SessionID    uuid.UUID      `json:"session_id"`
	Permissions  []string       `json:"permissions,omitempty"`
	Message      string         `json:"message"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=50"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	FirstName       string `json:"first_name" validate:"max=50"`
	LastName        string `json:"last_name" validate:"max=50"`
	Phone           string `json:"phone" validate:"max=20"`
	InviteCode      string `json:"invite_code"` // 邀请码（可选）
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User    *PublicUser `json:"user"`
	Message string      `json:"message"`
	Token   string      `json:"verification_token,omitempty"` // 邮箱验证令牌
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	SessionID    uuid.UUID `json:"session_id"`
	RefreshToken string    `json:"refresh_token"`
	LogoutAll    bool      `json:"logout_all"` // 是否登出所有设备
}

// LogoutResponse 登出响应
type LogoutResponse struct {
	Message string `json:"message"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPasswordResponse 忘记密码响应
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Token   string `json:"reset_token,omitempty"` // 重置令牌（开发环境可返回）
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// VerifyEmailRequest 验证邮箱请求
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// VerifyEmailResponse 验证邮箱响应
type VerifyEmailResponse struct {
	Message string `json:"message"`
}

// ResendVerificationRequest 重发验证邮件请求
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResendVerificationResponse 重发验证邮件响应
type ResendVerificationResponse struct {
	Message string `json:"message"`
	Token   string `json:"verification_token,omitempty"` // 验证令牌（开发环境可返回）
}

// ValidateTokenRequest 验证令牌请求
type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

// ValidateTokenResponse 验证令牌响应
type ValidateTokenResponse struct {
	Valid   bool        `json:"valid"`
	User    *PublicUser `json:"user,omitempty"`
	Claims  interface{} `json:"claims,omitempty"`
	Message string      `json:"message,omitempty"`
}

// AuthStatus 认证状态
type AuthStatus struct {
	Authenticated bool        `json:"authenticated"`
	User          *PublicUser `json:"user,omitempty"`
	SessionID     uuid.UUID   `json:"session_id,omitempty"`
	ExpiresAt     time.Time   `json:"expires_at,omitempty"`
	Permissions   []string    `json:"permissions,omitempty"`
}

// Permission 权限定义
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Description string    `json:"description" gorm:"size:255"`
	Resource    string    `json:"resource" gorm:"not null;size:100"` // 资源名称
	Action      string    `json:"action" gorm:"not null;size:50"`    // 操作名称
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate 在创建前生成UUID
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	Role         UserRole   `json:"role" gorm:"type:varchar(20);not null"`
	PermissionID uuid.UUID  `json:"permission_id" gorm:"type:char(36);not null"`
	CreatedAt    time.Time  `json:"created_at"`
	
	// 关联
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

// BeforeCreate 在创建前生成UUID
func (rp *RolePermission) BeforeCreate(tx *gorm.DB) error {
	if rp.ID == uuid.Nil {
		rp.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserPermission 用户权限关联表（用于特殊权限分配）
type UserPermission struct {
	ID           uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:char(36);not null"`
	PermissionID uuid.UUID  `json:"permission_id" gorm:"type:char(36);not null"`
	Granted      bool       `json:"granted" gorm:"default:true"` // true=授予，false=撤销
	CreatedAt    time.Time  `json:"created_at"`
	
	// 关联
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Permission Permission `json:"permission,omitempty" gorm:"foreignKey:PermissionID"`
}

// BeforeCreate 在创建前生成UUID
func (up *UserPermission) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (UserPermission) TableName() string {
	return "user_permissions"
}