package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Menu 菜单模型
type Menu struct {
	ID          uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Title       string     `gorm:"type:varchar(100);not null" json:"title"`
	Path        string     `gorm:"type:varchar(200);not null" json:"path"`
	Icon        string     `gorm:"type:varchar(100)" json:"icon"`
	ParentID    *uuid.UUID `gorm:"type:char(36);index" json:"parent_id"`
	Sort        int        `gorm:"type:int;default:0" json:"sort"`
	Level       int        `gorm:"type:int;default:1" json:"level"`
	IsVisible   bool       `gorm:"type:boolean;default:true" json:"is_visible"`
	IsEnabled   bool       `gorm:"type:boolean;default:true" json:"is_enabled"`
	RequiredRole UserRole  `gorm:"type:varchar(50);default:'USER'" json:"required_role"`
	Component   string     `gorm:"type:varchar(200)" json:"component"`
	Redirect    string     `gorm:"type:varchar(200)" json:"redirect"`
	Meta        string     `gorm:"type:json" json:"meta"` // JSON格式的元数据
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// 关联关系
	Parent   *Menu  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Menu `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// BeforeCreate 创建前生成UUID
func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	// 如果Meta为空，设置默认JSON值
	if m.Meta == "" {
		m.Meta = "{}"
	}
	return nil
}

// MenuResponse 菜单响应结构
type MenuResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Path        string          `json:"path"`
	Icon        string          `json:"icon"`
	ParentID    *uuid.UUID      `json:"parent_id"`
	Sort        int             `json:"sort"`
	Level       int             `json:"level"`
	IsVisible   bool            `json:"is_visible"`
	IsEnabled   bool            `json:"is_enabled"`
	RequiredRole UserRole       `json:"required_role"`
	Component   string          `json:"component"`
	Redirect    string          `json:"redirect"`
	Meta        interface{}     `json:"meta"`
	Children    []MenuResponse  `json:"children,omitempty"`
}

// ToResponse 转换为响应结构
func (m *Menu) ToResponse() MenuResponse {
	response := MenuResponse{
		ID:          m.ID,
		Name:        m.Name,
		Title:       m.Title,
		Path:        m.Path,
		Icon:        m.Icon,
		ParentID:    m.ParentID,
		Sort:        m.Sort,
		Level:       m.Level,
		IsVisible:   m.IsVisible,
		IsEnabled:   m.IsEnabled,
		RequiredRole: m.RequiredRole,
		Component:   m.Component,
		Redirect:    m.Redirect,
	}

	// 处理JSON元数据
	if m.Meta != "" {
		// 这里可以解析JSON字符串为interface{}
		response.Meta = m.Meta
	}

	// 处理子菜单
	if len(m.Children) > 0 {
		response.Children = make([]MenuResponse, len(m.Children))
		for i, child := range m.Children {
			response.Children[i] = child.ToResponse()
		}
	}

	return response
}