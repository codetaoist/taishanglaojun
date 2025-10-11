package models

import (
	"time"
)

// MultimodalRequest 多模态AI请求
type MultimodalRequest struct {
	ID          string                 `json:"id" bson:"_id"`
	UserID      string                 `json:"user_id" bson:"user_id"`
	SessionID   string                 `json:"session_id" bson:"session_id"`
	Type        MultimodalType         `json:"type" bson:"type"`
	Inputs      []MultimodalInput      `json:"inputs" bson:"inputs"`
	Outputs     []MultimodalOutputType `json:"outputs" bson:"outputs"`
	Config      MultimodalConfig       `json:"config" bson:"config"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	Status      string                 `json:"status" bson:"status"`
	Metadata    RequestMetadata        `json:"metadata" bson:"metadata"`
}

// MultimodalType 多模态交互类�?
type MultimodalType string

const (
	MultimodalTypeChat        MultimodalType = "chat"        // 多模态对�?
	MultimodalTypeAnalysis    MultimodalType = "analysis"    // 多模态分�?
	MultimodalTypeGeneration  MultimodalType = "generation"  // 多模态生�?
	MultimodalTypeTranslation MultimodalType = "translation" // 多模态翻�?
	MultimodalTypeSearch      MultimodalType = "search"      // 多模态搜�?
)

// MultimodalInput 多模态输�?
type MultimodalInput struct {
	Type     InputType   `json:"type"`
	Content  interface{} `json:"content"`
	Metadata InputMetadata `json:"metadata"`
}

// InputType 输入类型
type InputType string

const (
	InputTypeText  InputType = "text"
	InputTypeAudio InputType = "audio"
	InputTypeImage InputType = "image"
	InputTypeVideo InputType = "video"
	InputTypeFile  InputType = "file"
)

// InputMetadata 输入元数�?
type InputMetadata struct {
	MimeType   string            `json:"mime_type"`
	Size       int64             `json:"size"`
	Duration   float64           `json:"duration,omitempty"` // 音频/视频时长(�?
	Dimensions ImageDimensions   `json:"dimensions,omitempty"` // 图像/视频尺寸
	Language   string            `json:"language,omitempty"`
	Encoding   string            `json:"encoding,omitempty"`
	Quality    string            `json:"quality,omitempty"`
	Custom     map[string]string `json:"custom,omitempty"`
}

// ImageDimensions 图像尺寸
type ImageDimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TextInput 文本输入
type TextInput struct {
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
	Format   string `json:"format,omitempty"` // plain, markdown, html
}

// AudioInput 音频输入
type AudioInput struct {
	Data     []byte  `json:"data,omitempty"`     // 音频数据
	URL      string  `json:"url,omitempty"`      // 音频URL
	Format   string  `json:"format"`             // wav, mp3, flac, etc.
	Duration float64 `json:"duration"`           // 时长(�?
	Language string  `json:"language,omitempty"` // 语言代码
}

// ImageInput 图像输入
type ImageInput struct {
	Data       []byte          `json:"data,omitempty"` // 图像数据
	URL        string          `json:"url,omitempty"`  // 图像URL
	Format     string          `json:"format"`         // jpg, png, gif, etc.
	Dimensions ImageDimensions `json:"dimensions"`
	Description string         `json:"description,omitempty"` // 图像描述
}

// VideoInput 视频输入
type VideoInput struct {
	Data       []byte          `json:"data,omitempty"` // 视频数据
	URL        string          `json:"url,omitempty"`  // 视频URL
	Format     string          `json:"format"`         // mp4, avi, mov, etc.
	Duration   float64         `json:"duration"`       // 时长(�?
	Dimensions ImageDimensions `json:"dimensions"`
	FrameRate  float64         `json:"frame_rate"`
}

// MultimodalOutputType 多模态输出类�?
type MultimodalOutputType string

const (
	OutputTypeText  MultimodalOutputType = "text"
	OutputTypeAudio MultimodalOutputType = "audio"
	OutputTypeImage MultimodalOutputType = "image"
	OutputTypeVideo MultimodalOutputType = "video"
)

// MultimodalConfig 多模态配�?
type MultimodalConfig struct {
	Provider     string                 `json:"provider"`     // openai, anthropic, google, etc.
	Model        string                 `json:"model"`        // gpt-4-vision, claude-3, etc.
	Temperature  float32                `json:"temperature"`
	MaxTokens    int                    `json:"max_tokens"`
	Stream       bool                   `json:"stream"`
	AudioConfig  AudioConfig            `json:"audio_config,omitempty"`
	ImageConfig  ImageConfig            `json:"image_config,omitempty"`
	VideoConfig  VideoConfig            `json:"video_config,omitempty"`
	CustomConfig map[string]interface{} `json:"custom_config,omitempty"`
}

// AudioConfig 音频配置
type AudioConfig struct {
	Voice      string  `json:"voice"`       // 语音类型
	Speed      float32 `json:"speed"`       // 语�?
	Pitch      float32 `json:"pitch"`       // 音调
	Volume     float32 `json:"volume"`      // 音量
	Format     string  `json:"format"`      // 输出格式
	SampleRate int     `json:"sample_rate"` // 采样�?
	Language   string  `json:"language"`    // 语言
}

// ImageConfig 图像配置
type ImageConfig struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`    // jpg, png, webp
	Quality   int    `json:"quality"`   // 1-100
	Style     string `json:"style"`     // realistic, artistic, cartoon
	Aspect    string `json:"aspect"`    // 16:9, 4:3, 1:1
}

// VideoConfig 视频配置
type VideoConfig struct {
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Duration  float64 `json:"duration"`
	FrameRate float64 `json:"frame_rate"`
	Format    string  `json:"format"` // mp4, avi, mov
	Quality   string  `json:"quality"` // low, medium, high
}

// MultimodalResponse 多模态响�?
type MultimodalResponse struct {
	ID          string                `json:"id" bson:"_id"`
	RequestID   string                `json:"request_id" bson:"request_id"`
	UserID      string                `json:"user_id" bson:"user_id"`
	SessionID   string                `json:"session_id" bson:"session_id"`
	Type        MultimodalType        `json:"type" bson:"type"`
	Outputs     []MultimodalOutput    `json:"outputs" bson:"outputs"`
	CreatedAt   time.Time             `json:"created_at" bson:"created_at"`
	Status      string                `json:"status" bson:"status"`
	Error       string                `json:"error,omitempty" bson:"error,omitempty"`
	Metadata    ResponseMetadata      `json:"metadata" bson:"metadata"`
}

// MultimodalOutput 多模态输�?
type MultimodalOutput struct {
	Type     MultimodalOutputType `json:"type"`
	Content  interface{}          `json:"content"`
	Metadata OutputMetadata       `json:"metadata"`
}

// OutputMetadata 输出元数�?
type OutputMetadata struct {
	MimeType   string            `json:"mime_type"`
	Size       int64             `json:"size"`
	Duration   float64           `json:"duration,omitempty"`
	Dimensions ImageDimensions   `json:"dimensions,omitempty"`
	Quality    QualityMetrics    `json:"quality"`
	Custom     map[string]string `json:"custom,omitempty"`
}

// TextOutput 文本输出
type TextOutput struct {
	Content  string `json:"content"`
	Language string `json:"language"`
	Format   string `json:"format"`
}

// AudioOutput 音频输出
type AudioOutput struct {
	Data       []byte  `json:"data,omitempty"`
	URL        string  `json:"url,omitempty"`
	Format     string  `json:"format"`
	Duration   float64 `json:"duration"`
	SampleRate int     `json:"sample_rate"`
	Language   string  `json:"language,omitempty"`
}

// ImageOutput 图像输出
type ImageOutput struct {
	Data        []byte          `json:"data,omitempty"`
	URL         string          `json:"url,omitempty"`
	Format      string          `json:"format"`
	Dimensions  ImageDimensions `json:"dimensions"`
	Description string          `json:"description,omitempty"`
}

// VideoOutput 视频输出
type VideoOutput struct {
	Data       []byte          `json:"data,omitempty"`
	URL        string          `json:"url,omitempty"`
	Format     string          `json:"format"`
	Duration   float64         `json:"duration"`
	Dimensions ImageDimensions `json:"dimensions"`
	FrameRate  float64         `json:"frame_rate"`
}

// MultimodalSession 多模态会�?
type MultimodalSession struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	UserID      string                 `json:"user_id" gorm:"index"`
	Title       string                 `json:"title"`
	Type        MultimodalType         `json:"type"`
	Config      MultimodalConfig       `json:"config" gorm:"type:json"`
	Messages    []MultimodalMessage    `json:"messages" gorm:"foreignKey:SessionID"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastActiveAt time.Time             `json:"last_active_at"`
	Status      string                 `json:"status"` // active, archived, deleted
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:json"`
}

// MultimodalMessage 多模态消�?
type MultimodalMessage struct {
	ID        string             `json:"id" gorm:"primaryKey"`
	SessionID string             `json:"session_id" gorm:"index"`
	Role      string             `json:"role"` // user, assistant, system
	Inputs    []MultimodalInput  `json:"inputs" gorm:"type:json"`
	Outputs   []MultimodalOutput `json:"outputs" gorm:"type:json"`
	CreatedAt time.Time          `json:"created_at"`
	TokenUsed int                `json:"token_used"`
	Cost      float64            `json:"cost"`
}
