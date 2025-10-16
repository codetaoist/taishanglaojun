package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Code        string    `json:"code" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Resource    string    `json:"resource" gorm:"type:varchar(100)"`
	Action      string    `json:"action" gorm:"type:varchar(50)"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BeforeCreate 创建前生成UUID
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

// UserRoleAssociation 用户角色关联模型
type UserRoleAssociation struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);not null"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:char(36);not null"`
	Role      Role      `json:"role" gorm:"foreignKey:RoleID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate 创建前生成UUID
func (ura *UserRoleAssociation) BeforeCreate(tx *gorm.DB) error {
	if ura.ID == uuid.Nil {
		ura.ID = uuid.New()
	}
	return nil
}

// TableName 指定表名
func (UserRoleAssociation) TableName() string {
	return "user_roles"
}