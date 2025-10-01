package models

import (
	"time"
)

// AIResponse AI服务响应基础模型
type AIResponse struct {
	ID           string            `json:"id" bson:"_id"`
	RequestID    string            `json:"request_id" bson:"request_id"`
	UserID       string            `json:"user_id" bson:"user_id"`
	Type         string            `json:"type" bson:"type"`
	Provider     string            `json:"provider" bson:"provider"`
	ResponseData map[string]interface{} `json:"response_data" bson:"response_data"`
	Usage        UsageStats        `json:"usage" bson:"usage"`
	CreatedAt    time.Time         `json:"created_at" bson:"created_at"`
	ProcessTime  int64             `json:"process_time" bson:"process_time"` // 毫秒
	Status       string            `json:"status" bson:"status"`
	Error        *ErrorInfo        `json:"error,omitempty" bson:"error,omitempty"`
}

// UsageStats 使用统计
type UsageStats struct {
	PromptTokens     int     `json:"prompt_tokens" bson:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens" bson:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens" bson:"total_tokens"`
	Cost             float64 `json:"cost" bson:"cost"` // 成本（美元）
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code    string `json:"code" bson:"code"`
	Message string `json:"message" bson:"message"`
	Details string `json:"details" bson:"details"`
}

// ChatResponseData 对话响应数据
type ChatResponseData struct {
	Message     ChatMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
	Model       string      `json:"model"`
}

// GenerateResponseData 内容生成响应数据
type GenerateResponseData struct {
	Content      string            `json:"content"`
	Type         string            `json:"type"`
	Quality      float32           `json:"quality"`      // 质量评分 0-1
	Confidence   float32           `json:"confidence"`   // 置信�?0-1
	Metadata     map[string]string `json:"metadata"`
}

// AnalyzeResponseData 分析响应数据
type AnalyzeResponseData struct {
	Type       string                 `json:"type"`
	Result     map[string]interface{} `json:"result"`
	Confidence float32                `json:"confidence"`
	Details    map[string]interface{} `json:"details"`
}

// SentimentResult 情感分析结果
type SentimentResult struct {
	Label      string  `json:"label"`      // positive, negative, neutral
	Score      float32 `json:"score"`      // -1 �?1
	Confidence float32 `json:"confidence"` // 0 �?1
}

// KeywordsResult 关键词提取结�?type KeywordsResult struct {
	Keywords []Keyword `json:"keywords"`
	Summary  string    `json:"summary"`
}

// Keyword 关键�?type Keyword struct {
	Text   string  `json:"text"`
	Score  float32 `json:"score"`
	Type   string  `json:"type"`   // person, place, concept, etc.
	Count  int     `json:"count"`  // 出现次数
}

// SimilarityResult 相似度计算结�?type SimilarityResult struct {
	Score      float32           `json:"score"`      // 0 �?1
	Method     string            `json:"method"`     // cosine, jaccard, etc.
	Details    map[string]float32 `json:"details"`
	Breakdown  map[string]interface{} `json:"breakdown"`
}

// EmbedResponseData 向量化响应数�?type EmbedResponseData struct {
	Embedding []float32 `json:"embedding"`
	Model     string    `json:"model"`
	Dimension int       `json:"dimension"`
}
