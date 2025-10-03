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
	GetName() string
	GetModels() []string
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
	Message   Message `json:"message"`
	Usage     Usage   `json:"usage"`
	SessionID string  `json:"session_id"`
}

// GenerateRequest 生成请求
type GenerateRequest struct {
	Prompt      string  `json:"prompt"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// GenerateResponse 生成响应
type GenerateResponse struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	Content string `json:"content"`
	Type    string `json:"type"` // sentiment, keywords, classification
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	Type         string   `json:"type"`
	Confidence   float32  `json:"confidence"`
	Result       string   `json:"result"`
	Details      []string `json:"details"`
	Usage        Usage    `json:"usage"`
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	Cost             float64   `json:"cost,omitempty"`
	Duration         time.Duration `json:"duration,omitempty"`
}
