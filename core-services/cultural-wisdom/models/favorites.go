package models

import (
	"time"
)

// WisdomFavorite 智慧收藏模型
type WisdomFavorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	WisdomID  string    `json:"wisdom_id" gorm:"type:varchar(255);not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (WisdomFavorite) TableName() string {
	return "wisdom_favorites"
}

// WisdomNote 智慧笔记模型
type WisdomNote struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	WisdomID  string    `json:"wisdom_id" gorm:"type:varchar(255);not null;index"`
	Title     string    `json:"title" gorm:"type:varchar(255)"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	IsPrivate bool      `json:"is_private" gorm:"default:true"`
	Tags      string    `json:"tags" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (WisdomNote) TableName() string {
	return "wisdom_notes"
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
	WisdomID string `json:"wisdom_id" binding:"required"`
}

// NoteRequest 笔记请求
type NoteRequest struct {
	WisdomID  string   `json:"wisdom_id" binding:"required"`
	Title     string   `json:"title"`
	Content   string   `json:"content" binding:"required"`
	IsPrivate bool     `json:"is_private"`
	Tags      []string `json:"tags"`
}

// NoteUpdateRequest 笔记更新请求
type NoteUpdateRequest struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	IsPrivate bool     `json:"is_private"`
	Tags      []string `json:"tags"`
}

// FavoriteResponse 收藏响应
type FavoriteResponse struct {
	ID        uint      `json:"id"`
	UserID    string    `json:"user_id"`
	WisdomID  string    `json:"wisdom_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NoteResponse 笔记响应
type NoteResponse struct {
	ID        uint      `json:"id"`
	UserID    string    `json:"user_id"`
	WisdomID  string    `json:"wisdom_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsPrivate bool      `json:"is_private"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

