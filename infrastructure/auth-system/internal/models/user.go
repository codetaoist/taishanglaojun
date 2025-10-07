package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserStatus 用户状态枚举举
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// UserRole 用户角色枚举
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleModerator  UserRole = "moderator"
	RoleUser       UserRole = "user"
	RoleGuest      UserRole = "guest"
)

// User 用户模型
type User struct {
	ID          uuid.UUID      `json:"id" gorm:"type:char(36);primary_key"`
	Username    string         `json:"username" gorm:"uniqueIndex;not null;size:50" validate:"required,min=3,max=50"`
	Email       string         `json:"email" gorm:"uniqueIndex;not null;size:100" validate:"required,email"`
	Password    string         `json:"-" gorm:"not null;size:255" validate:"required,min=8"`
	FirstName   string         `json:"first_name" gorm:"size:50" validate:"max=50"`
	LastName    string         `json:"last_name" gorm:"size:50" validate:"max=50"`
	Avatar      string         `json:"avatar" gorm:"size:255"`
	Phone       string         `json:"phone" gorm:"size:20" validate:"max=20"`
	Status      UserStatus     `json:"status" gorm:"type:varchar(20);default:'active';index" validate:"oneof=active inactive suspended deleted"`
	Role        UserRole       `json:"role" gorm:"type:varchar(20);default:'user';index" validate:"oneof=admin moderator user guest"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联
	Sessions []Session `json:"sessions,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Tokens   []Token   `json:"tokens,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM钩子：创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return u.HashPassword()
}

// BeforeUpdate GORM钩子：更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// 如果密码被修改，重新哈希
	if tx.Statement.Changed("Password") {
		return u.HashPassword()
	}
	return nil
}

// HashPassword 哈希密码
func (u *User) HashPassword() error {
	if u.Password == "" {
		return nil
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsActive 检查用户是否激活�?
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// HasRole 检查用户是否具有效指定角�?
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// IsAdmin 检查是否为管理器�?
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsModerator 检查是否为版主
func (u *User) IsModerator() bool {
	return u.Role == RoleModerator || u.Role == RoleAdmin
}

// GetFullName 获取全名
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

// UpdateLastLogin 更新最后登录时�?
func (u *User) UpdateLastLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
}

// ToPublic 转换为公开信息（不包含敏感数量据�?
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Avatar:      u.Avatar,
		Status:      u.Status,
		Role:        u.Role,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// PublicUser 公开用户信息（不包含敏感数量据�?
type PublicUser struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Avatar      string     `json:"avatar"`
	Status      UserStatus `json:"status"`
	Role        UserRole   `json:"role"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username  string   `json:"username" validate:"required,min=3,max=50"`
	Email     string   `json:"email" validate:"required,email"`
	Password  string   `json:"password" validate:"required,min=8"`
	FirstName string   `json:"first_name" validate:"max=50"`
	LastName  string   `json:"last_name" validate:"max=50"`
	Phone     string   `json:"phone" validate:"max=20"`
	Role      UserRole `json:"role" validate:"omitempty,oneof=admin moderator user guest"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	FirstName *string     `json:"first_name" validate:"omitempty,max=50"`
	LastName  *string     `json:"last_name" validate:"omitempty,max=50"`
	Avatar    *string     `json:"avatar" validate:"omitempty,max=255"`
	Phone     *string     `json:"phone" validate:"omitempty,max=20"`
	Status    *UserStatus `json:"status" validate:"omitempty,oneof=active inactive suspended deleted"`
	Role      *UserRole   `json:"role" validate:"omitempty,oneof=admin moderator user guest"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UserQuery 用户查询参数量
type UserQuery struct {
	Username string     `form:"username"`
	Email    string     `form:"email"`
	Status   UserStatus `form:"status"`
	Role     UserRole   `form:"role"`
	Search   string     `form:"search"` // 搜索用户名、邮箱件箱、姓�?
	Page     int        `form:"page" validate:"min=1"`
	PageSize int        `form:"page_size" validate:"min=1,max=100"`
	OrderBy  string     `form:"order_by" validate:"oneof=created_at updated_at username email"`
	Order    string     `form:"order" validate:"oneof=asc desc"`
}
