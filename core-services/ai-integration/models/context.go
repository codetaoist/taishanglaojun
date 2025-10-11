package models

import (
	"time"
)

// ConversationContext 对话上下文模型
type ConversationContext struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SessionID    string    `json:"session_id" gorm:"index;not null"`
	UserID       uint      `json:"user_id" gorm:"index;not null"`
	Summary      string    `json:"summary" gorm:"type:text"`
	Keywords     []string  `json:"keywords" gorm:"serializer:json"`
	Topics       []string  `json:"topics" gorm:"serializer:json"`
	Entities     []string  `json:"entities" gorm:"serializer:json"`
	Sentiment    string    `json:"sentiment" gorm:"size:50"`
	Intent       string    `json:"intent" gorm:"size:100"`
	MessageCount int       `json:"message_count" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IntentHistory 意图历史记录
type IntentHistory struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"session_id" gorm:"index;not null"`
	UserID     uint      `json:"user_id" gorm:"index;not null"`
	Intent     string    `json:"intent" gorm:"size:100;not null"`
	Entities   []string  `json:"entities" gorm:"serializer:json"`
	Sentiment  string    `json:"sentiment" gorm:"size:50"`
	Message    string    `json:"message" gorm:"type:text"`
	Confidence float64   `json:"confidence" gorm:"type:decimal(5,4)"`
	CreatedAt  time.Time `json:"created_at"`
}

// ContextKeyword 上下文关键词
type ContextKeyword struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index;not null"`
	Keyword   string    `json:"keyword" gorm:"size:100;not null"`
	Weight    float64   `json:"weight" gorm:"type:decimal(5,4);default:1.0"`
	Count     int       `json:"count" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContextTopic 上下文话题模型
type ContextTopic struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index;not null"`
	Topic     string    `json:"topic" gorm:"size:100;not null"`
	Weight    float64   `json:"weight" gorm:"type:decimal(5,4);default:1.0"`
	Count     int       `json:"count" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (ConversationContext) TableName() string {
	return "conversation_contexts"
}

func (IntentHistory) TableName() string {
	return "intent_histories"
}

func (ContextKeyword) TableName() string {
	return "context_keywords"
}

func (ContextTopic) TableName() string {
	return "context_topics"
}
