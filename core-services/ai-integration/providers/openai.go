package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider OpenAI提供�?type OpenAIProvider struct {
	config OpenAIConfig
	client *http.Client
}

// NewOpenAIProvider 创建OpenAI提供�?func NewOpenAIProvider(config OpenAIConfig) *OpenAIProvider {
	return &OpenAIProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Chat 发送对话消�?func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 设置默认模型
	if req.Model == "" {
		req.Model = p.config.Model
	}

	// 构建请求�?	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, NewProviderError("openai", "marshal_error", "Failed to marshal request", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, NewProviderError("openai", "request_error", "Failed to create request", err)
	}

	// 设置请求�?	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// 发送请�?	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, NewProviderError("openai", "http_error", "Failed to send request", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewProviderError("openai", "read_error", "Failed to read response", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, NewProviderError("openai", errorResp.Error.Code, errorResp.Error.Message, nil)
		}
		
		return nil, NewProviderError("openai", "http_error", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), nil)
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, NewProviderError("openai", "unmarshal_error", "Failed to unmarshal response", err)
	}

	return &chatResp, nil
}

// GetName 获取提供商名�?func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModels 获取支持的模型列�?func (p *OpenAIProvider) GetModels() []string {
	return []string{
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo-preview",
		"gpt-4-vision-preview",
	}
}

// ValidateConfig 验证配置
func (p *OpenAIProvider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return fmt.Errorf("openai api_key is required")
	}
	
	if p.config.BaseURL == "" {
		return fmt.Errorf("openai base_url is required")
	}
	
	if p.config.Model == "" {
		return fmt.Errorf("openai model is required")
	}
	
	if p.config.Timeout <= 0 {
		p.config.Timeout = 30 * time.Second
	}
	
	return nil
}
