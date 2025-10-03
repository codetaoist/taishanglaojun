package models

import (
	"time"
)

// Conversation 对话会话模型
type Conversation struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	Title     string    `json:"title" bson:"title"`
	Messages  []Message `json:"messages" bson:"messages"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	IsActive  bool      `json:"is_active" bson:"is_active"`
	Metadata  Metadata  `json:"metadata" bson:"metadata"`
}

// Message 对话消息
type Message struct {
	ID        string            `json:"id" bson:"_id"`
	Role      string            `json:"role" bson:"role"` // system, user, assistant
	Content   string            `json:"content" bson:"content"`
	Timestamp time.Time         `json:"timestamp" bson:"timestamp"`
	Metadata  map[string]string `json:"metadata" bson:"metadata"`
}

// Metadata 会话元数据
type Metadata struct {
	Source      string            `json:"source" bson:"source"`           // web, mobile, api
	UserAgent   string            `json:"user_agent" bson:"user_agent"`
	IPAddress   string            `json:"ip_address" bson:"ip_address"`
	Tags        []string          `json:"tags" bson:"tags"`
	CustomData  map[string]string `json:"custom_data" bson:"custom_data"`
	TokenUsage  ConversationTokenUsage `json:"token_usage" bson:"token_usage"`
}

// ConversationTokenUsage 对话令牌使用统计
type ConversationTokenUsage struct {
	TotalPromptTokens     int `json:"total_prompt_tokens" bson:"total_prompt_tokens"`
	TotalCompletionTokens int `json:"total_completion_tokens" bson:"total_completion_tokens"`
	TotalTokens           int `json:"total_tokens" bson:"total_tokens"`
}

// ConversationSummary 对话摘要
type ConversationSummary struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	MessageCount int      `json:"message_count"`
	LastMessage string    `json:"last_message"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

// ConversationListRequest 对话列表请求
type ConversationListRequest struct {
	UserID   string `json:"user_id"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// ConversationListResponse 对话列表响应
type ConversationListResponse struct {
	Conversations []ConversationSummary `json:"conversations"`
	Total         int                   `json:"total"`
	Page          int                   `json:"page"`
	PageSize      int                   `json:"page_size"`
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	Title    string `json:"title"`
	UserID   string `json:"user_id"`
	Metadata Metadata `json:"metadata,omitempty"`
}

// UpdateConversationRequest 更新对话请求
type UpdateConversationRequest struct {
	Title    string `json:"title,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// AddMessageRequest 添加消息请求
type AddMessageRequest struct {
	Role     string            `json:"role"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

