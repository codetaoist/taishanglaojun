package providers

import (
	"context"
	"time"
)

// AIProvider 定义AI服务提供者接口
type AIProvider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error)
	Embed(ctx context.Context, text string) ([]float32, error)
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// ChatRequest 对话请求
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	Temperature float32   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	Message     Message   `json:"message"`
	Usage       Usage     `json:"usage"`
	SessionID   string    `json:"session_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// GenerateRequest 内容生成请求
type GenerateRequest struct {
	Type        string            `json:"type"`        // summary, explanation, translation
	Content     string            `json:"content"`
	Parameters  map[string]string `json:"parameters"`
	UserID      string            `json:"user_id"`
	Temperature float32           `json:"temperature"`
}

// GenerateResponse 内容生成响应
type GenerateResponse struct {
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Usage     Usage     `json:"usage"`
	Timestamp time.Time `json:"timestamp"`
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	Type    string `json:"type"`    // sentiment, keywords, similarity
	Content string `json:"content"`
	Target  string `json:"target"`  // 用于相似度比较的目标文本
	UserID  string `json:"user_id"`
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	Type      string                 `json:"type"`
	Result    map[string]interface{} `json:"result"`
	Usage     Usage                  `json:"usage"`
	Timestamp time.Time              `json:"timestamp"`
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}