package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"math"

	"go.uber.org/zap"
)

// ClaudeProvider Anthropic Claude提供者
type ClaudeProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	maxRetries int
	retryDelay time.Duration
}

// ClaudeConfig Claude配置
type ClaudeConfig struct {
	APIKey     string `yaml:"api_key" json:"api_key"`
	BaseURL    string `yaml:"base_url" json:"base_url"`
	Timeout    int    `yaml:"timeout" json:"timeout"`
	MaxRetries int    `yaml:"max_retries" json:"max_retries"`
	RetryDelay int    `yaml:"retry_delay" json:"retry_delay"` // 秒
}

// ClaudeError Claude API错误
type ClaudeError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewClaudeProvider 创建Claude提供者
func NewClaudeProvider(config ClaudeConfig, logger *zap.Logger) *ClaudeProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com/v1"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1
	}

	return &ClaudeProvider{
		apiKey:  config.APIKey,
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger:     logger,
		maxRetries: config.MaxRetries,
		retryDelay: time.Duration(config.RetryDelay) * time.Second,
	}
}

// Chat 聊天对话
func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 验证请求参数
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

	// 转换消息格式为Claude格式
	claudeMessages := make([]map[string]interface{}, 0, len(req.Messages))
	var systemMessage string

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemMessage = msg.Content
			continue
		}
		claudeMessages = append(claudeMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 构建Claude API请求
	claudeReq := map[string]interface{}{
		"model":      "claude-3-sonnet-20240229",
		"max_tokens": req.MaxTokens,
		"messages":   claudeMessages,
	}

	if req.MaxTokens == 0 {
		claudeReq["max_tokens"] = 1000
	}

	if systemMessage != "" {
		claudeReq["system"] = systemMessage
	}

	if req.Temperature > 0 {
		claudeReq["temperature"] = req.Temperature
	}

	startTime := time.Now()
	resp, err := p.sendRequestWithRetry(ctx, "/messages", claudeReq)
	if err != nil {
		p.logger.Error("Claude chat request failed",
			zap.Error(err),
			zap.String("user_id", req.UserID),
			zap.String("session_id", req.SessionID))
		return nil, err
	}

	// 解析响应
	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(resp, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in Claude response")
	}

	duration := time.Since(startTime)
	totalTokens := claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens

	p.logger.Info("Claude chat request completed",
		zap.String("user_id", req.UserID),
		zap.String("session_id", req.SessionID),
		zap.Duration("duration", duration),
		zap.Int("total_tokens", totalTokens))

	return &ChatResponse{
		Message: Message{
			Role:    "assistant",
			Content: claudeResp.Content[0].Text,
		},
		Usage: Usage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      totalTokens,
			Duration:         duration,
		},
		SessionID: req.SessionID,
	}, nil
}

// Generate 文本生成
func (p *ClaudeProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	// 转换为聊天请求
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
			Duration:         chatResp.Usage.Duration,
		},
	}, nil
}

// Analyze 内容分析
func (p *ClaudeProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// 根据分析类型构建提示
	var prompt string
	switch req.Type {
	case "sentiment":
		prompt = fmt.Sprintf("请分析以下文本的情感倾向，只返回positive、negative或neutral中的一个：\n\n%s", req.Content)
	case "keywords":
		prompt = fmt.Sprintf("请提取以下文本的关键词，用逗号分隔，不要其他解释：\n\n%s", req.Content)
	case "classification":
		prompt = fmt.Sprintf("请对以下文本进行分类，给出最合适的类别：\n\n%s", req.Content)
	default:
		prompt = fmt.Sprintf("请分析以下文本的主要内容和特点：\n\n%s", req.Content)
	}

	chatReq := ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3,
		MaxTokens:   500,
		UserID:      req.UserID,
	}

	chatResp, err := p.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// 解析分析结果
	result := strings.TrimSpace(chatResp.Message.Content)
	confidence := float32(0.85) // Claude通常有较高的准确性
	details := []string{}

	if req.Type == "keywords" {
		details = strings.Split(result, ",")
		for i, detail := range details {
			details[i] = strings.TrimSpace(detail)
		}
	}

	return &AnalyzeResponse{
		Type:       req.Type,
		Confidence: confidence,
		Result:     result,
		Details:    details,
		Usage: Usage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
			Duration:         chatResp.Usage.Duration,
		},
	}, nil
}

// Embed 文本嵌入（Claude不直接支持，返回错误）
func (p *ClaudeProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("Claude does not support text embedding directly")
}

// IntentRecognition 意图识别
func (p *ClaudeProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
	if strings.TrimSpace(req.Text) == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	prompt := fmt.Sprintf(`请分析以下文本的用户意图，并以JSON格式返回结果：

文本："%s"

请返回以下格式的JSON：
{
  "intent": "意图名称",
  "confidence": 0.95,
  "entities": [
    {
      "name": "实体名称",
      "value": "实体值",
      "type": "实体类型",
      "confidence": 0.9
    }
  ]
}

只返回JSON，不要其他解释。`, req.Text)

	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3,
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("intent recognition failed: %w", err)
	}

	// 解析JSON响应
	var intentData struct {
		Intent     string   `json:"intent"`
		Confidence float32  `json:"confidence"`
		Entities   []Entity `json:"entities"`
	}

	content := strings.TrimSpace(generateResp.Content)
	// 尝试提取JSON部分
	if start := strings.Index(content, "{"); start != -1 {
		if end := strings.LastIndex(content, "}"); end != -1 && end > start {
			content = content[start : end+1]
		}
	}

	if err := json.Unmarshal([]byte(content), &intentData); err != nil {
		// JSON解析失败，返回默认结果
		return &IntentResponse{
			Intent:     "unknown",
			Confidence: 0.5,
			Entities:   []Entity{},
			Context:    req.Context,
			Usage: Usage{
				PromptTokens:     generateResp.Usage.PromptTokens,
				CompletionTokens: generateResp.Usage.CompletionTokens,
				TotalTokens:      generateResp.Usage.TotalTokens,
				Duration:         time.Since(startTime),
			},
		}, nil
	}

	return &IntentResponse{
		Intent:     intentData.Intent,
		Confidence: intentData.Confidence,
		Entities:   intentData.Entities,
		Context:    req.Context,
		Usage: Usage{
			PromptTokens:     generateResp.Usage.PromptTokens,
			CompletionTokens: generateResp.Usage.CompletionTokens,
			TotalTokens:      generateResp.Usage.TotalTokens,
			Duration:         time.Since(startTime),
		},
	}, nil
}

// SentimentAnalysis 情感分析
func (p *ClaudeProvider) SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
	if strings.TrimSpace(req.Text) == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	prompt := fmt.Sprintf(`请分析以下文本的情感，并以JSON格式返回结果：

文本："%s"

请返回以下格式的JSON：
{
  "sentiment": "positive/negative/neutral",
  "score": 0.8,
  "confidence": 0.95,
  "emotions": [
    {
      "name": "joy",
      "score": 0.8,
      "confidence": 0.9
    }
  ]
}

情感分数范围：-1.0到1.0，置信度范围：0.0到1.0
可能的情感类型：joy, anger, sadness, fear, surprise, disgust
只返回JSON，不要其他解释。`, req.Text)

	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3,
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("sentiment analysis failed: %w", err)
	}

	// 解析JSON响应
	var sentimentData struct {
		Sentiment  string    `json:"sentiment"`
		Score      float32   `json:"score"`
		Confidence float32   `json:"confidence"`
		Emotions   []Emotion `json:"emotions"`
	}

	content := strings.TrimSpace(generateResp.Content)
	// 尝试提取JSON部分
	if start := strings.Index(content, "{"); start != -1 {
		if end := strings.LastIndex(content, "}"); end != -1 && end > start {
			content = content[start : end+1]
		}
	}

	if err := json.Unmarshal([]byte(content), &sentimentData); err != nil {
		// JSON解析失败，返回默认结果
		return &SentimentResponse{
			Sentiment:  "neutral",
			Score:      0.0,
			Confidence: 0.5,
			Emotions:   []Emotion{},
			Usage: Usage{
				PromptTokens:     generateResp.Usage.PromptTokens,
				CompletionTokens: generateResp.Usage.CompletionTokens,
				TotalTokens:      generateResp.Usage.TotalTokens,
				Duration:         time.Since(startTime),
			},
		}, nil
	}

	return &SentimentResponse{
		Sentiment:  sentimentData.Sentiment,
		Score:      sentimentData.Score,
		Confidence: sentimentData.Confidence,
		Emotions:   sentimentData.Emotions,
		Usage: Usage{
			PromptTokens:     generateResp.Usage.PromptTokens,
			CompletionTokens: generateResp.Usage.CompletionTokens,
			TotalTokens:      generateResp.Usage.TotalTokens,
			Duration:         time.Since(startTime),
		},
	}, nil
}

// GenerateImage 图像生成（Claude不支持）
func (p *ClaudeProvider) GenerateImage(ctx context.Context, req ImageGenerateRequest) (*ImageGenerateResponse, error) {
	return nil, fmt.Errorf("Claude does not support image generation")
}

// AnalyzeImage 图像分析（Claude 3支持视觉功能）
func (p *ClaudeProvider) AnalyzeImage(ctx context.Context, req ImageAnalyzeRequest) (*ImageAnalyzeResponse, error) {
	if req.ImageURL == "" && req.ImageBase64 == "" && len(req.ImageData) == 0 {
		return nil, fmt.Errorf("image data is required")
	}

	// 构建消息内容
	content := []map[string]interface{}{
		{
			"type": "text",
			"text": req.Prompt,
		},
	}

	// 添加图像内容
	if req.ImageBase64 != "" {
		content = append(content, map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": "image/jpeg",
				"data":       req.ImageBase64,
			},
		})
	} else if req.ImageURL != "" {
		// Claude API不直接支持URL，需要先下载图像
		return nil, fmt.Errorf("Claude requires base64 encoded images, URL not supported")
	}

	claudeReq := map[string]interface{}{
		"model":      "claude-3-sonnet-20240229",
		"max_tokens": 1000,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
	}

	respData, err := p.sendRequestWithRetry(ctx, "/messages", claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image: %w", err)
	}

	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respData, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	return &ImageAnalyzeResponse{
		Description: claudeResp.Content[0].Text,
		Usage: Usage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
		RequestID: fmt.Sprintf("claude_analyze_%d", time.Now().Unix()),
	}, nil
}

// EditImage 图像编辑（Claude不支持）
func (p *ClaudeProvider) EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error) {
	return nil, fmt.Errorf("Claude does not support image editing")
}

// GetName 获取提供者名称
func (p *ClaudeProvider) GetName() string {
	return "claude"
}

// GetModels 获取支持的模型列表
func (p *ClaudeProvider) GetModels() []string {
	return []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
		"claude-2.1",
		"claude-2.0",
		"claude-instant-1.2",
	}
}

// GetCapabilities 获取支持的功能列表
func (p *ClaudeProvider) GetCapabilities() []string {
	return []string{
		"chat", "completion", "analysis", "intent-recognition",
		"sentiment-analysis", "image-analysis", "vision",
	}
}

// sendRequestWithRetry 带重试机制的HTTP请求
func (p *ClaudeProvider) sendRequestWithRetry(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * p.retryDelay
			p.logger.Warn("Retrying Claude request",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.String("endpoint", endpoint))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := p.sendRequest(ctx, endpoint, data)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 检查是否应该重试
		if !p.shouldRetry(err) {
			break
		}
	}

	return nil, fmt.Errorf("Claude request failed after %d attempts: %w", p.maxRetries+1, lastErr)
}

// shouldRetry 判断是否应该重试
func (p *ClaudeProvider) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// 网络错误或临时错误应该重试
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "status 429") || // 速率限制
		strings.Contains(errStr, "status 500") || // 服务器错误
		strings.Contains(errStr, "status 502") ||
		strings.Contains(errStr, "status 503") ||
		strings.Contains(errStr, "status 504") {
		return true
	}

	return false
}

// sendRequest HTTP请求
func (p *ClaudeProvider) sendRequest(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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
		// 尝试解析Claude错误格式
		var claudeErr ClaudeError
		if json.Unmarshal(body, &claudeErr) == nil && claudeErr.Message != "" {
			return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, claudeErr.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}