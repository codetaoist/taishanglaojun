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
	IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error)
	SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error)
	
	// 图像相关功能
	GenerateImage(ctx context.Context, req ImageGenerateRequest) (*ImageGenerateResponse, error)
	AnalyzeImage(ctx context.Context, req ImageAnalyzeRequest) (*ImageAnalyzeResponse, error)
	EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error)
	
	GetName() string
	GetModels() []string
	GetCapabilities() []string
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

// IntentRequest 意图识别请求
type IntentRequest struct {
	Text     string            `json:"text"`
	Context  map[string]string `json:"context,omitempty"`
	UserID   string            `json:"user_id,omitempty"`
	Language string            `json:"language,omitempty"`
}

// IntentResponse 意图识别响应
type IntentResponse struct {
	Intent     string            `json:"intent"`
	Confidence float32           `json:"confidence"`
	Entities   []Entity          `json:"entities,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
	Usage      Usage             `json:"usage"`
}

// Entity 实体信息
type Entity struct {
	Name       string  `json:"name"`
	Value      string  `json:"value"`
	Type       string  `json:"type"`
	Confidence float32 `json:"confidence"`
	StartPos   int     `json:"start_pos,omitempty"`
	EndPos     int     `json:"end_pos,omitempty"`
}

// SentimentRequest 情感分析请求
type SentimentRequest struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// SentimentResponse 情感分析响应
type SentimentResponse struct {
	Sentiment  string  `json:"sentiment"`  // positive, negative, neutral
	Score      float32 `json:"score"`      // -1.0 to 1.0
	Confidence float32 `json:"confidence"` // 0.0 to 1.0
	Emotions   []Emotion `json:"emotions,omitempty"`
	Usage      Usage   `json:"usage"`
}

// Emotion 情感详情
type Emotion struct {
	Name       string  `json:"name"`       // joy, anger, sadness, fear, etc.
	Score      float32 `json:"score"`      // 0.0 to 1.0
	Confidence float32 `json:"confidence"` // 0.0 to 1.0
}

// 图像生成请求
type ImageGenerateRequest struct {
	Prompt         string            `json:"prompt"`
	NegativePrompt string            `json:"negative_prompt,omitempty"`
	Style          string            `json:"style,omitempty"`          // realistic, artistic, cartoon, etc.
	Size           string            `json:"size,omitempty"`           // 1024x1024, 512x512, etc.
	Quality        string            `json:"quality,omitempty"`        // standard, hd
	Count          int               `json:"count,omitempty"`          // 生成图片数量
	Seed           int64             `json:"seed,omitempty"`           // 随机种子
	Steps          int               `json:"steps,omitempty"`          // 生成步数
	CfgScale       float32           `json:"cfg_scale,omitempty"`      // 提示词相关性
	Model          string            `json:"model,omitempty"`          // 使用的模型
	UserID         string            `json:"user_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// 图像生成响应
type ImageGenerateResponse struct {
	Images    []GeneratedImage `json:"images"`
	Usage     Usage            `json:"usage"`
	RequestID string           `json:"request_id"`
	Model     string           `json:"model"`
}

// 生成的图像
type GeneratedImage struct {
	URL         string            `json:"url,omitempty"`
	Base64      string            `json:"base64,omitempty"`
	Width       int               `json:"width"`
	Height      int               `json:"height"`
	Format      string            `json:"format"`
	Seed        int64             `json:"seed,omitempty"`
	RevisedPrompt string          `json:"revised_prompt,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// 图像分析请求
type ImageAnalyzeRequest struct {
	ImageURL    string            `json:"image_url,omitempty"`
	ImageBase64 string            `json:"image_base64,omitempty"`
	ImageData   []byte            `json:"image_data,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`         // 分析指令
	Features    []string          `json:"features,omitempty"`       // 要分析的特征: objects, text, faces, colors, etc.
	Language    string            `json:"language,omitempty"`       // 返回语言
	Detail      string            `json:"detail,omitempty"`         // low, high
	UserID      string            `json:"user_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// 图像分析响应
type ImageAnalyzeResponse struct {
	Description string                 `json:"description"`           // 图像描述
	Objects     []DetectedObject       `json:"objects,omitempty"`     // 检测到的对象
	Text        []DetectedText         `json:"text,omitempty"`        // 检测到的文字
	Faces       []DetectedFace         `json:"faces,omitempty"`       // 检测到的人脸
	Colors      []DominantColor        `json:"colors,omitempty"`      // 主要颜色
	Tags        []string               `json:"tags,omitempty"`        // 标签
	Categories  []Category             `json:"categories,omitempty"`  // 分类
	Emotions    []Emotion              `json:"emotions,omitempty"`    // 情感分析
	Safety      SafetyAnalysis         `json:"safety,omitempty"`      // 安全性分析
	Usage       Usage                  `json:"usage"`
	RequestID   string                 `json:"request_id"`
}

// 检测到的对象
type DetectedObject struct {
	Name       string    `json:"name"`
	Confidence float32   `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box,omitempty"`
}

// 检测到的文字
type DetectedText struct {
	Text        string      `json:"text"`
	Confidence  float32     `json:"confidence"`
	Language    string      `json:"language,omitempty"`
	BoundingBox BoundingBox `json:"bounding_box,omitempty"`
}

// 检测到的人脸
type DetectedFace struct {
	Age         int         `json:"age,omitempty"`
	Gender      string      `json:"gender,omitempty"`
	Emotions    []Emotion   `json:"emotions,omitempty"`
	Confidence  float32     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box,omitempty"`
}

// 边界框
type BoundingBox struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
}

// 主要颜色
type DominantColor struct {
	Color      string  `json:"color"`      // hex color code
	Percentage float32 `json:"percentage"` // 占比
	Name       string  `json:"name"`       // 颜色名称
}

// 分类
type Category struct {
	Name       string  `json:"name"`
	Confidence float32 `json:"confidence"`
}

// 安全性分析
type SafetyAnalysis struct {
	IsAdult    bool    `json:"is_adult"`
	IsViolent  bool    `json:"is_violent"`
	IsRacy     bool    `json:"is_racy"`
	AdultScore float32 `json:"adult_score"`
	ViolenceScore float32 `json:"violence_score"`
	RacyScore  float32 `json:"racy_score"`
}

// 图像编辑请求
type ImageEditRequest struct {
	ImageURL       string            `json:"image_url,omitempty"`
	ImageBase64    string            `json:"image_base64,omitempty"`
	ImageData      []byte            `json:"image_data,omitempty"`
	MaskURL        string            `json:"mask_url,omitempty"`
	MaskBase64     string            `json:"mask_base64,omitempty"`
	MaskData       []byte            `json:"mask_data,omitempty"`
	Prompt         string            `json:"prompt"`               // 编辑指令
	Size           string            `json:"size,omitempty"`
	Count          int               `json:"count,omitempty"`
	UserID         string            `json:"user_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// 图像编辑响应
type ImageEditResponse struct {
	Images    []GeneratedImage `json:"images"`
	Usage     Usage            `json:"usage"`
	RequestID string           `json:"request_id"`
	Model     string           `json:"model"`
}
