package models

import (
	"time"
)

// AIRequest AI服务请求基础模型
type AIRequest struct {
	ID          string            `json:"id" bson:"_id"`
	UserID      string            `json:"user_id" bson:"user_id"`
	Type        string            `json:"type" bson:"type"` // chat, generate, analyze, embed
	Provider    string            `json:"provider" bson:"provider"`
	RequestData map[string]interface{} `json:"request_data" bson:"request_data"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	Status      string            `json:"status" bson:"status"` // pending, processing, completed, failed
	Metadata    RequestMetadata   `json:"metadata" bson:"metadata"`
}

// RequestMetadata 请求元数�?
type RequestMetadata struct {
	SessionID   string            `json:"session_id" bson:"session_id"`
	Source      string            `json:"source" bson:"source"`
	Priority    int               `json:"priority" bson:"priority"` // 1-10
	RetryCount  int               `json:"retry_count" bson:"retry_count"`
	Tags        []string          `json:"tags" bson:"tags"`
	CustomData  map[string]string `json:"custom_data" bson:"custom_data"`
}

// ChatRequestData 对话请求数据
type ChatRequestData struct {
	Messages    []ChatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	Model       string        `json:"model"`
	Stream      bool          `json:"stream"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	SessionID string    `json:"session_id" gorm:"index"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	TokenUsed int       `json:"token_used"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GenerateRequestData 内容生成请求数据
type GenerateRequestData struct {
	Type        string            `json:"type"` // summary, explanation, translation
	Content     string            `json:"content"`
	Template    string            `json:"template"`
	Parameters  map[string]string `json:"parameters"`
	Language    string            `json:"language"`
	MaxLength   int               `json:"max_length"`
}

// AnalyzeRequestData 分析请求数据
type AnalyzeRequestData struct {
	Type     string `json:"type"` // sentiment, keyword, classification
	Content  string `json:"content"`
	Language string `json:"language"`
	Options  map[string]interface{} `json:"options"`
}

// EmbedRequestData 向量化请求数�?
type EmbedRequestData struct {
	Texts []string `json:"texts"`
	Model string   `json:"model"`
}

