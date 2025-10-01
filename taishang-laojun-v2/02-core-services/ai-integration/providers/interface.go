package providers

import (
	"context"
	"time"
)

// AIProvider AI提供商接口
type AIProvider interface {
	// Chat 发送对话消息
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	
	// GetName 获取提供商名称
	GetName() string
	
	// GetModels 获取支持的模型列表
	GetModels() []string
	
	// ValidateConfig 验证配置
	ValidateConfig() error
}

// ChatRequest 对话请求
type ChatRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
	Stream   bool      `json:"stream,omitempty"`
	
	// 可选参数
	Temperature      *float32 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float32 `json:"top_p,omitempty"`
	FrequencyPenalty *float32 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float32 `json:"presence_penalty,omitempty"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
}

// Choice 选择项
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name    string        `yaml:"name"`
	Enabled bool          `yaml:"enabled"`
	Timeout time.Duration `yaml:"timeout"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	ProviderConfig `yaml:",inline"`
	APIKey         string `yaml:"api_key"`
	BaseURL        string `yaml:"base_url"`
	Model          string `yaml:"model"`
}

// AzureConfig Azure OpenAI配置
type AzureConfig struct {
	ProviderConfig `yaml:",inline"`
	APIKey         string `yaml:"api_key"`
	Endpoint       string `yaml:"endpoint"`
	Deployment     string `yaml:"deployment"`
	APIVersion     string `yaml:"api_version"`
}

// BaiduConfig 百度文心一言配置
type BaiduConfig struct {
	ProviderConfig `yaml:",inline"`
	APIKey         string `yaml:"api_key"`
	SecretKey      string `yaml:"secret_key"`
	Model          string `yaml:"model"`
}

// ProviderError 提供商错误
type ProviderError struct {
	Provider string
	Code     string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Provider + ": " + e.Message + " (" + e.Err.Error() + ")"
	}
	return e.Provider + ": " + e.Message
}

// NewProviderError 创建提供商错误
func NewProviderError(provider, code, message string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Code:     code,
		Message:  message,
		Err:      err,
	}
}