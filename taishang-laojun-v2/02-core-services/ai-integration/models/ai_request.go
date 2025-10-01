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

// RequestMetadata 请求元数据
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
	Role    string `json:"role"`
	Content string `json:"content"`
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
	Type       string   `json:"type"` // sentiment, keywords, similarity, classification
	Content    string   `json:"content"`
	Target     string   `json:"target"`     // 用于相似度比较
	Categories []string `json:"categories"` // 用于分类
	Language   string   `json:"language"`
}

// EmbedRequestData 向量化请求数据
type EmbedRequestData struct {
	Text     string `json:"text"`
	Model    string `json:"model"`
	Language string `json:"language"`
}