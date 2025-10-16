package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	Key         string    `json:"key" gorm:"type:varchar(255);uniqueIndex;not null"`
	Value       string    `json:"value" gorm:"type:text"`
	Type        string    `json:"type" gorm:"type:varchar(50);default:'string'"`
	Description string    `json:"description" gorm:"type:text"`
	Category    string    `json:"category" gorm:"type:varchar(100);default:'general'"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(255)"`
	UpdatedBy   string    `json:"updated_by" gorm:"type:varchar(255)"`
}

// BeforeCreate UUID生成
func (sc *SystemConfig) BeforeCreate(tx *gorm.DB) error {
	if sc.ID == uuid.Nil {
		sc.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// Role 角色模型
type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:char(36);primaryKey"`
	Name        string       `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Code        string       `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string       `json:"description" gorm:"type:text"`
	Level       int          `json:"level" gorm:"default:1"`
	Status      string       `json:"status" gorm:"type:varchar(20);default:active"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// BeforeCreate 创建前生成UUID
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}



// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}