package providers

import (
	"context"
	"time"
)

// AIProvider AI 提供器接口
type AIProvider interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error)
	Embed(ctx context.Context, text string) ([]float32, error)
	IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error)
	SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error)

	//
	GenerateImage(ctx context.Context, req ImageGenerateRequest) (*ImageGenerateResponse, error)
	AnalyzeImage(ctx context.Context, req ImageAnalyzeRequest) (*ImageAnalyzeResponse, error)
	EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error)

	GetName() string
	GetModels() []string
	GetCapabilities() []string
}

// Message 消息
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	UserID      string    `json:"user_id"`
	SessionID   string    `json:"session_id"`
	Temperature float32   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// ChatResponse 聊天响应
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
	UserID  string `json:"user_id"`
}

// AnalyzeResponse 分析响应
type AnalyzeResponse struct {
	Type       string   `json:"type"`
	Confidence float32  `json:"confidence"`
	Result     string   `json:"result"`
	Details    []string `json:"details"`
	Usage      Usage    `json:"usage"`
}

// Usage 用量
type Usage struct {
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	TotalTokens      int           `json:"total_tokens"`
	Cost             float64       `json:"cost,omitempty"`
	Duration         time.Duration `json:"duration,omitempty"`
}

// IntentRequest 意图请求
type IntentRequest struct {
	Text     string            `json:"text"`
	Context  map[string]string `json:"context,omitempty"`
	UserID   string            `json:"user_id,omitempty"`
	Language string            `json:"language,omitempty"`
}

// IntentResponse 意图响应
type IntentResponse struct {
	Intent     string            `json:"intent"`
	Confidence float32           `json:"confidence"`
	Entities   []Entity          `json:"entities,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
	Usage      Usage             `json:"usage"`
}

// Entity 实体
type Entity struct {
	Name       string  `json:"name"`
	Value      string  `json:"value"`
	Type       string  `json:"type"`
	Confidence float32 `json:"confidence"`
	StartPos   int     `json:"start_pos,omitempty"`
	EndPos     int     `json:"end_pos,omitempty"`
}

// SentimentRequest 情感请求
type SentimentRequest struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// SentimentResponse 情感响应
type SentimentResponse struct {
	Sentiment  string    `json:"sentiment"`  // positive, negative, neutral
	Score      float32   `json:"score"`      // -1.0 to 1.0
	Confidence float32   `json:"confidence"` // 0.0 to 1.0
	Emotions   []Emotion `json:"emotions,omitempty"`
	Usage      Usage     `json:"usage"`
}

// Emotion 情感
type Emotion struct {
	Name       string  `json:"name"`       // joy, anger, sadness, fear, etc.
	Score      float32 `json:"score"`      // 0.0 to 1.0
	Confidence float32 `json:"confidence"` // 0.0 to 1.0
}

// ImageGenerateRequest 图片生成请求
type ImageGenerateRequest struct {
	Prompt         string            `json:"prompt"`
	NegativePrompt string            `json:"negative_prompt,omitempty"`
	Style          string            `json:"style,omitempty"`     // realistic, artistic, cartoon, etc.
	Size           string            `json:"size,omitempty"`      // 1024x1024, 512x512, etc.
	Quality        string            `json:"quality,omitempty"`   // standard, hd
	Count          int               `json:"count,omitempty"`     //
	Seed           int64             `json:"seed,omitempty"`      //
	Steps          int               `json:"steps,omitempty"`     //
	CfgScale       float32           `json:"cfg_scale,omitempty"` // 0.010.0
	Model          string            `json:"model,omitempty"`     //
	UserID         string            `json:"user_id,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// ImageGenerateResponse 图片生成响应
type ImageGenerateResponse struct {
	Images    []GeneratedImage `json:"images"`
	Usage     Usage            `json:"usage"`
	RequestID string           `json:"request_id"`
	Model     string           `json:"model"`
}

// GeneratedImage 生成的图片
type GeneratedImage struct {
	URL           string            `json:"url,omitempty"`
	Base64        string            `json:"base64,omitempty"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	Format        string            `json:"format"`
	Seed          int64             `json:"seed,omitempty"`
	RevisedPrompt string            `json:"revised_prompt,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// ImageAnalyzeRequest 图片分析请求
type ImageAnalyzeRequest struct {
	ImageURL    string            `json:"image_url,omitempty"`
	ImageBase64 string            `json:"image_base64,omitempty"`
	ImageData   []byte            `json:"image_data,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`   // 分析提示
	Features    []string          `json:"features,omitempty"` // : objects, text, faces, colors, etc.
	Language    string            `json:"language,omitempty"` // 分析语言
	Detail      string            `json:"detail,omitempty"`   // low, high
	UserID      string            `json:"user_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ImageAnalyzeResponse 图片分析响应
type ImageAnalyzeResponse struct {
	Description string           `json:"description"`          // 图片描述
	Objects     []DetectedObject `json:"objects,omitempty"`    // 检测到的物体
	Text        []DetectedText   `json:"text,omitempty"`       // 检测到的文本
	Faces       []DetectedFace   `json:"faces,omitempty"`      // 检测到的人脸
	Colors      []DominantColor  `json:"colors,omitempty"`     // 主要颜色
	Tags        []string         `json:"tags,omitempty"`       // 图片标签
	Categories  []Category       `json:"categories,omitempty"` // 图片分类
	Emotions    []Emotion        `json:"emotions,omitempty"`   // 图片情感
	Safety      SafetyAnalysis   `json:"safety,omitempty"`     // 图片安全分析
	Usage       Usage            `json:"usage"`
	RequestID   string           `json:"request_id"`
}

// DetectedObject 检测到的物体
type DetectedObject struct {
	Name        string      `json:"name"`
	Confidence  float32     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box,omitempty"`
}

// DetectedText 检测到的文本
type DetectedText struct {
	Text        string      `json:"text"`
	Confidence  float32     `json:"confidence"`
	Language    string      `json:"language,omitempty"` // 文本语言
	BoundingBox BoundingBox `json:"bounding_box,omitempty"`
}

// DetectedFace 检测到的人脸
type DetectedFace struct {
	Age         int         `json:"age,omitempty"`
	Gender      string      `json:"gender,omitempty"`
	Emotions    []Emotion   `json:"emotions,omitempty"`
	Confidence  float32     `json:"confidence"`
	BoundingBox BoundingBox `json:"bounding_box,omitempty"`
}

// BoundingBox 边界框
type BoundingBox struct {
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
}

// DominantColor 主要颜色
type DominantColor struct {
	Color      string  `json:"color"`      // hex color code
	Percentage float32 `json:"percentage"` // 颜色占比
	Name       string  `json:"name"`       // 颜色名称
}

// Category 图片分类
type Category struct {
	Name       string  `json:"name"`       // 分类名称
	Confidence float32 `json:"confidence"` // 分类置信度
}

// SafetyAnalysis 图片安全分析
type SafetyAnalysis struct {
	IsAdult       bool    `json:"is_adult"`       // 是否成人内容
	IsViolent     bool    `json:"is_violent"`     // 是否暴力内容
	IsRacy        bool    `json:"is_racy"`        // 是否racy内容
	AdultScore    float32 `json:"adult_score"`    // 成人内容得分
	ViolenceScore float32 `json:"violence_score"` // 暴力内容得分
	RacyScore     float32 `json:"racy_score"`     // racy内容得分
}

// ImageEditRequest 图片编辑请求
type ImageEditRequest struct {
	ImageURL    string            `json:"image_url,omitempty"`
	ImageBase64 string            `json:"image_base64,omitempty"`
	ImageData   []byte            `json:"image_data,omitempty"`
	MaskURL     string            `json:"mask_url,omitempty"`
	MaskBase64  string            `json:"mask_base64,omitempty"`
	MaskData    []byte            `json:"mask_data,omitempty"`
	Prompt      string            `json:"prompt"` //
	Size        string            `json:"size,omitempty"`
	Count       int               `json:"count,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ImageEditResponse 图片编辑响应
type ImageEditResponse struct {
	Images    []GeneratedImage `json:"images"`     // 编辑后的图片
	Usage     Usage            `json:"usage"`      // 图片编辑使用量
	RequestID string           `json:"request_id"` // 请求ID
	Model     string           `json:"model"`      // 模型名称
}
