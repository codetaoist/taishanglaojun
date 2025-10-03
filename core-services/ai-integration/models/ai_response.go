package models

import (
	"time"
)

// AIResponse AI服务响应基础模型
type AIResponse struct {
	ID          string            `json:"id" bson:"_id"`
	RequestID   string            `json:"request_id" bson:"request_id"`
	UserID      string            `json:"user_id" bson:"user_id"`
	Type        string            `json:"type" bson:"type"`
	Provider    string            `json:"provider" bson:"provider"`
	ResponseData map[string]interface{} `json:"response_data" bson:"response_data"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	Status      string            `json:"status" bson:"status"`
	Error       string            `json:"error,omitempty" bson:"error,omitempty"`
	Metadata    ResponseMetadata  `json:"metadata" bson:"metadata"`
}

// ResponseMetadata 响应元数据
type ResponseMetadata struct {
	ProcessingTime time.Duration     `json:"processing_time" bson:"processing_time"`
	TokensUsed     int               `json:"tokens_used" bson:"tokens_used"`
	Cost           float64           `json:"cost" bson:"cost"`
	Model          string            `json:"model" bson:"model"`
	Quality        QualityMetrics    `json:"quality" bson:"quality"`
	CustomData     map[string]string `json:"custom_data" bson:"custom_data"`
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	Relevance   float32 `json:"relevance"`   // 相关性 0-1
	Coherence   float32 `json:"coherence"`   // 连贯性 0-1
	Accuracy    float32 `json:"accuracy"`    // 准确性 0-1
	Completeness float32 `json:"completeness"` // 完整性 0-1
}

// ChatResponseData 对话响应数据
type ChatResponseData struct {
	Message     ChatMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
	Usage       TokenUsage  `json:"usage"`
}

// TokenUsage 令牌使用情况
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateResponseData 生成响应数据
type GenerateResponseData struct {
	Content      string     `json:"content"`
	Type         string     `json:"type"`
	Language     string     `json:"language"`
	Quality      float32    `json:"quality"`
	Alternatives []string   `json:"alternatives"`
	Usage        TokenUsage `json:"usage"`
}

// AnalyzeResponseData 分析响应数据
type AnalyzeResponseData struct {
	Type    string      `json:"type"`
	Results interface{} `json:"results"`
	Summary string      `json:"summary"`
}

// SentimentResult 情感分析结果
type SentimentResult struct {
	Label      string  `json:"label"`      // positive, negative, neutral
	Score      float32 `json:"score"`      // -1 到 1
	Confidence float32 `json:"confidence"` // 0 到 1
}

// KeywordsResult 关键词提取结果
type KeywordsResult struct {
	Keywords []Keyword `json:"keywords"`
	Summary  string    `json:"summary"`
}

// Keyword 关键词
type Keyword struct {
	Text   string  `json:"text"`
	Score  float32 `json:"score"`
	Type   string  `json:"type"`
}

// ClassificationResult 分类结果
type ClassificationResult struct {
	Category   string  `json:"category"`
	Confidence float32 `json:"confidence"`
	Categories []CategoryScore `json:"categories"`
}

// CategoryScore 分类得分
type CategoryScore struct {
	Name  string  `json:"name"`
	Score float32 `json:"score"`
}

// EmbedResponseData 向量化响应数据
type EmbedResponseData struct {
	Embeddings [][]float32 `json:"embeddings"`
	Model      string      `json:"model"`
	Usage      TokenUsage  `json:"usage"`
}

