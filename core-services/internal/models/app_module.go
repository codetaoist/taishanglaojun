package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID          uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	Username    string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email       string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"type:varchar(255);not null" json:"-"`
	DisplayName string     `gorm:"type:varchar(100)" json:"display_name"`
	Avatar      string     `gorm:"type:varchar(255)" json:"avatar"`
	Role        UserRole   `gorm:"type:varchar(50);default:'USER'" json:"role"`
	Level       int        `gorm:"type:int;default:1" json:"level"` // 1
	IsActive    bool       `gorm:"type:boolean;default:true" json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BeforeCreate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserRole 用户角色
type UserRole string

const (
	RoleGuest      UserRole = "GUEST"
	RoleUser       UserRole = "USER"
	RolePremium    UserRole = "PREMIUM"
	RoleAdmin      UserRole = "ADMIN"
	RoleSuperAdmin UserRole = "SUPER_ADMIN"
)

// ModuleCategory 应用模块分类
type ModuleCategory string

const (
	CategorySystem   ModuleCategory = "SYSTEM"
	CategoryUser     ModuleCategory = "USER"
	CategoryBusiness ModuleCategory = "BUSINESS"
	CategoryCreative ModuleCategory = "CREATIVE"
	CategoryAI       ModuleCategory = "AI_SERVICE"
	CategoryFile     ModuleCategory = "FILE_SERVICE"
)

// AppModule 应用模块
type AppModule struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	DisplayName  string         `gorm:"type:varchar(100);not null" json:"display_name"`
	Description  string         `gorm:"type:text" json:"description"`
	Category     ModuleCategory `gorm:"type:varchar(50);not null" json:"category"`
	Icon         string         `gorm:"type:varchar(100)" json:"icon"`
	Path         string         `gorm:"type:varchar(200);not null" json:"path"`
	RequiredRole UserRole       `gorm:"type:varchar(50);not null;default:'USER'" json:"required_role"`
	IsCore       bool           `gorm:"type:boolean;default:false" json:"is_core"`
	IsEnabled    bool           `gorm:"type:boolean;default:true" json:"is_enabled"`
	AutoStart    bool           `gorm:"type:boolean;default:false" json:"auto_start"`
	Priority     int            `gorm:"type:int;default:0" json:"priority"`
	Version      string         `gorm:"type:varchar(20)" json:"version"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate UUID
func (m *AppModule) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// UserModulePermission 用户模块权限
type UserModulePermission struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	ModuleID  uuid.UUID `gorm:"type:char(36);not null;index" json:"module_id"`
	Enabled   bool      `gorm:"type:boolean;default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	//
	User   User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Module AppModule `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
}

// BeforeCreate UUID
func (p *UserModulePermission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// UserPreference 用户偏好
type UserPreference struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;uniqueIndex" json:"user_id"`
	Theme     string    `gorm:"type:varchar(50);default:'light'" json:"theme"`
	Language  string    `gorm:"type:varchar(10);default:'zh-CN'" json:"language"`
	MenuStyle string    `gorm:"type:varchar(50);default:'sidebar'" json:"menu_style"`
	AutoStart bool      `gorm:"type:boolean;default:false" json:"auto_start"`
	Settings  string    `gorm:"type:json" json:"settings"` // JSON格式
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	//
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate UUID
func (p *UserPreference) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ModuleDependency 模块依赖
type ModuleDependency struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	ModuleID     uuid.UUID `gorm:"type:char(36);not null;index" json:"module_id"`
	DependencyID uuid.UUID `gorm:"type:char(36);not null;index" json:"dependency_id"`
	Required     bool      `gorm:"type:boolean;default:true" json:"required"`
	CreatedAt    time.Time `json:"created_at"`

	//
	Module     AppModule `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Dependency AppModule `gorm:"foreignKey:DependencyID" json:"dependency,omitempty"`
}

// BeforeCreate UUID
func (d *ModuleDependency) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// UserSession 用户会话
type UserSession struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(500);not null;uniqueIndex" json:"token"`
	DeviceID  string    `gorm:"type:varchar(100)" json:"device_id"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsActive  bool      `gorm:"type:boolean;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	//
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// BeforeCreate UUID
func (s *UserSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsExpired 会话是否过期
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
