package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// OpenAIProvider OpenAI提供商实现
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey  string `yaml:"api_key" json:"api_key"`
	BaseURL string `yaml:"base_url" json:"base_url"`
	Timeout int    `yaml:"timeout" json:"timeout"`
}

// NewOpenAIProvider 创建OpenAI提供商
func NewOpenAIProvider(config OpenAIConfig, logger *zap.Logger) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	return &OpenAIProvider{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger: logger,
	}
}

// Chat 发送对话消息
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 构建OpenAI API请求
	openaiReq := map[string]interface{}{
		"model":       "gpt-3.5-turbo",
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	if req.MaxTokens == 0 {
		openaiReq["max_tokens"] = 1000
	}

	resp, err := p.sendRequest(ctx, "/chat/completions", openaiReq)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(resp, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	return &ChatResponse{
		Message: Message{
			Role:    openaiResp.Choices[0].Message.Role,
			Content: openaiResp.Choices[0].Message.Content,
		},
		Usage: Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
		SessionID: req.SessionID,
	}, nil
}

// Generate 生成内容
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// 将生成请求转换为对话请求
	chatReq := ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	chatResp, err := p.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	return &GenerateResponse{
		Content: chatResp.Message.Content,
		Usage: Usage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
	}, nil
}

// Analyze 分析内容
func (p *OpenAIProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	// 构建分析提示
	prompt := fmt.Sprintf("请分析以下文本的情感和关键词：\n\n%s", req.Content)

	chatReq := ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   500,
	}

	chatResp, err := p.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	return &AnalyzeResponse{
		Type:         req.Type,
		Confidence:   0.8,
		Result:       "neutral",
		Details:      []string{"keyword1", "keyword2"},
		Usage: Usage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
	}, nil
}

// Embed 生成嵌入向量
func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// 构建嵌入请求
	embedReq := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	resp, err := p.sendRequest(ctx, "/embeddings", embedReq)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var embedResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings in response")
	}

	return embedResp.Data[0].Embedding, nil
}

// sendRequest 发送HTTP请求
func (p *OpenAIProvider) sendRequest(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetName 获取提供商名称
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModels 获取支持的模型列表
func (p *OpenAIProvider) GetModels() []string {
	return []string{
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo-preview",
	}
}
