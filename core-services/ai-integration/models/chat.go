﻿package models

import (
	"time"
)

// ChatSession 
type ChatSession struct {
	ID           string       `json:"id" gorm:"primaryKey"`
	UserID       uint         `json:"user_id" gorm:"index"`
	Title        string       `json:"title"`
	Provider     string       `json:"provider"`
	Model        string       `json:"model"`
	Settings     ChatSettings `json:"settings" gorm:"embedded"`
	MessageCount int          `json:"message_count" gorm:"default:0"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Status       string       `json:"status" gorm:"default:active"` // active, archived, deleted
}

// ChatSettings 
type ChatSettings struct {
	Temperature  float32 `json:"temperature" bson:"temperature"`
	MaxTokens    int     `json:"max_tokens" bson:"max_tokens"`
	TopP         float32 `json:"top_p" bson:"top_p"`
	TopK         int     `json:"top_k" bson:"top_k"`
	SystemPrompt string  `json:"system_prompt" bson:"system_prompt"`
}

// ChatRequest 
type ChatRequest struct {
	SessionID   *string       `json:"session_id"`
	Message     string        `json:"message"`
	Messages    []ChatMessage `json:"messages"`
	Provider    string        `json:"provider,omitempty"`
	Model       string        `json:"model,omitempty"`
	Temperature float32       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	UserID      uint          `json:"-"`
}

// ChatResponse 
type ChatResponse struct {
	SessionID string     `json:"session_id"`
	MessageID uint       `json:"message_id"`
	Content   string     `json:"content"`
	TokenUsed int        `json:"token_used"`
	Provider  string     `json:"provider"`
	Model     string     `json:"model"`
	CreatedAt time.Time  `json:"created_at"`
	Usage     TokenUsage `json:"usage"`
}

// StreamChatResponse 
type StreamChatResponse struct {
	SessionID string      `json:"session_id"`
	MessageID string      `json:"message_id"`
	Delta     string      `json:"delta"`
	Done      bool        `json:"done"`
	Usage     *TokenUsage `json:"usage,omitempty"`
}

// ChatHistory 
type ChatHistory struct {
	SessionID string        `json:"session_id"`
	Messages  []ChatMessage `json:"messages"`
	Total     int           `json:"total"`
	Page      int           `json:"page"`
	PageSize  int           `json:"page_size"`
}

// CreateSessionRequest 
type CreateSessionRequest struct {
	Title        string       `json:"title"`
	Provider     string       `json:"provider,omitempty"`
	Model        string       `json:"model,omitempty"`
	Settings     ChatSettings `json:"settings,omitempty"`
	SystemPrompt string       `json:"system_prompt,omitempty"`
}

// UpdateSessionRequest 
type UpdateSessionRequest struct {
	Title    string       `json:"title,omitempty"`
	Settings ChatSettings `json:"settings,omitempty"`
	Status   string       `json:"status,omitempty"`
}

// SessionListRequest 
type SessionListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
}

// SessionListResponse 
type SessionListResponse struct {
	Sessions []ChatSession `json:"sessions"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// MessageListRequest 
type MessageListRequest struct {
	SessionID uint   `json:"session_id"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Order     string `json:"order"` // asc, desc
}

// MessageListResponse 
type MessageListResponse struct {
	Messages []ChatMessage `json:"messages"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// DeleteSessionRequest 
type DeleteSessionRequest struct {
	SessionID string `json:"session_id"`
	Hard      bool   `json:"hard"` // 
}

