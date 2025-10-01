package models

import (
	"time"

	"gorm.io/gorm"
)

// ChatSession 对话会话
type ChatSession struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"size:255"`
	Provider    string         `json:"provider" gorm:"size:50;not null"` // openai, azure, baidu
	Model       string         `json:"model" gorm:"size:100;not null"`   // gpt-3.5-turbo, gpt-4
	Status      string         `json:"status" gorm:"size:20;default:active"` // active, archived, deleted
	MessageCount int           `json:"message_count" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Messages []ChatMessage `json:"messages,omitempty" gorm:"foreignKey:SessionID"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	SessionID uint           `json:"session_id" gorm:"not null;index"`
	Role      string         `json:"role" gorm:"size:20;not null"` // user, assistant, system
	Content   string         `json:"content" gorm:"type:text;not null"`
	TokenUsed int            `json:"token_used" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 关联
	Session ChatSession `json:"-" gorm:"foreignKey:SessionID"`
}

// ChatRequest 对话请求
type ChatRequest struct {
	SessionID *uint  `json:"session_id,omitempty"` // 可选，用于继续对话
	Message   string `json:"message" binding:"required"`
	Provider  string `json:"provider,omitempty"` // 可选，默认使用配置中的提供商
	Model     string `json:"model,omitempty"`    // 可选，默认使用配置中的模型
	UserID    uint   `json:"-"`                  // 从JWT中获取
}

// ChatResponse 对话响应
type ChatResponse struct {
	SessionID uint   `json:"session_id"`
	MessageID uint   `json:"message_id"`
	Content   string `json:"content"`
	TokenUsed int    `json:"token_used"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
}

// SessionListRequest 会话列表请求
type SessionListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Status   string `form:"status,omitempty"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []ChatSession `json:"sessions"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// MessageListRequest 消息列表请求
type MessageListRequest struct {
	SessionID uint `uri:"session_id" binding:"required"`
	Page      int  `form:"page,default=1"`
	PageSize  int  `form:"page_size,default=50"`
}

// MessageListResponse 消息列表响应
type MessageListResponse struct {
	Messages []ChatMessage `json:"messages"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// TableName 指定表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}