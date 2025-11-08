package models

import (
	"time"
)

// ConversationSearchResult 对话搜索结果
type ConversationSearchResult struct {
	ConversationID   string            `json:"conversationId"`
	Title            string            `json:"title"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	Score            float64           `json:"score"`
	MatchedMessages  []MatchedMessage  `json:"matchedMessages"`
}

// MatchedMessage 匹配的消息
type MatchedMessage struct {
	ID      string  `json:"id"`
	Role    string  `json:"role"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// MessageSearchResult 消息搜索结果
type MessageSearchResult struct {
	MessageID string  `json:"messageId"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Score     float64 `json:"score"`
}

// SearchConversationsRequest 搜索对话请求
type SearchConversationsRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"topK,omitempty"`
}

// SearchConversationsResponse 搜索对话响应
type SearchConversationsResponse struct {
	Results []ConversationSearchResult `json:"results"`
	Total   int                        `json:"total"`
}

// SearchInConversationRequest 在对话中搜索请求
type SearchInConversationRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
	Query          string `json:"query" binding:"required"`
	TopK           int    `json:"topK,omitempty"`
}

// SearchInConversationResponse 在对话中搜索响应
type SearchInConversationResponse struct {
	Results []MessageSearchResult `json:"results"`
	Total   int                   `json:"total"`
}

// IndexConversationRequest 索引对话请求
type IndexConversationRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
}

// ReindexConversationRequest 重新索引对话请求
type ReindexConversationRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
}

// GetSimilarMessagesRequest 获取相似消息请求
type GetSimilarMessagesRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
	MessageID      string `json:"messageId" binding:"required"`
	TopK           int    `json:"topK,omitempty"`
}

// GetSimilarMessagesResponse 获取相似消息响应
type GetSimilarMessagesResponse struct {
	Results []MessageSearchResult `json:"results"`
	Total   int                   `json:"total"`
}