package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"math"
	"strings"

	"go.uber.org/zap"
)

// OpenAIProvider OpenAI提供者
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
	maxRetries int
	retryDelay time.Duration
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey     string `yaml:"api_key" json:"api_key"`
	BaseURL    string `yaml:"base_url" json:"base_url"`
	Timeout    int    `yaml:"timeout" json:"timeout"`
	MaxRetries int    `yaml:"max_retries" json:"max_retries"`
	RetryDelay int    `yaml:"retry_delay" json:"retry_delay"` // 秒
}

// OpenAIError OpenAI API错误
type OpenAIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewOpenAIProvider 创建OpenAI提供者
func NewOpenAIProvider(config OpenAIConfig, logger *zap.Logger) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
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

	return &OpenAIProvider{
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
func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 验证请求参数
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

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

	startTime := time.Now()
	resp, err := p.sendRequestWithRetry(ctx, "/chat/completions", openaiReq)
	if err != nil {
		p.logger.Error("Chat request failed", 
			zap.Error(err), 
			zap.String("user_id", req.UserID),
			zap.String("session_id", req.SessionID))
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

	duration := time.Since(startTime)
	p.logger.Info("Chat request completed",
		zap.String("user_id", req.UserID),
		zap.String("session_id", req.SessionID),
		zap.Duration("duration", duration),
		zap.Int("total_tokens", openaiResp.Usage.TotalTokens))

	return &ChatResponse{
		Message: Message{
			Role:    openaiResp.Choices[0].Message.Role,
			Content: openaiResp.Choices[0].Message.Content,
		},
		Usage: Usage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
			Duration:         duration,
		},
		SessionID: req.SessionID,
	}, nil
}

// Generate 文本生成
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
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
func (p *OpenAIProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// 根据分析类型构建提示
	var prompt string
	switch req.Type {
	case "sentiment":
		prompt = fmt.Sprintf("请分析以下文本的情感倾向，返回positive、negative或neutral：\n\n%s", req.Content)
	case "keywords":
		prompt = fmt.Sprintf("请提取以下文本的关键词，用逗号分隔：\n\n%s", req.Content)
	case "classification":
		prompt = fmt.Sprintf("请对以下文本进行分类：\n\n%s", req.Content)
	default:
		prompt = fmt.Sprintf("请分析以下文本：\n\n%s", req.Content)
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
	confidence := float32(0.8) // 默认置信度
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

// Embed 
func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// 
	embedReq := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": text,
	}

	resp, err := p.sendRequest(ctx, "/embeddings", embedReq)
	if err != nil {
		return nil, err
	}

	// 
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

// sendRequestWithRetry 带重试机制的HTTP请求
func (p *OpenAIProvider) sendRequestWithRetry(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	var lastErr error
	
	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * p.retryDelay
			p.logger.Warn("Retrying request", 
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

	return nil, fmt.Errorf("request failed after %d attempts: %w", p.maxRetries+1, lastErr)
}

// shouldRetry 判断是否应该重试
func (p *OpenAIProvider) shouldRetry(err error) bool {
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
		// 尝试解析OpenAI错误格式
		var openaiErr OpenAIError
		if json.Unmarshal(body, &openaiErr) == nil && openaiErr.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, openaiErr.Error.Message)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetName 
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModels 
func (p *OpenAIProvider) GetModels() []string {
	return []string{
		"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo",
		"text-davinci-003", "text-curie-001",
		"text-embedding-ada-002",
		"dall-e-3", "dall-e-2", "gpt-4-vision-preview",
	}
}

func (p *OpenAIProvider) GetCapabilities() []string {
	return []string{
		"chat", "completion", "embedding", "image-generation",
		"image-analysis", "image-edit", "vision",
	}
}

// IntentRecognition 
func (p *OpenAIProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
	// 
	prompt := fmt.Sprintf(`JSON

"%s"

巵
{
  "intent": "",
  "confidence": 0.95,
  "entities": [
    {
      "name": "",
      "value": "",
      "type": "",
      "confidence": 0.9
    }
  ]
}

`, req.Text)

	// Generate
	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3, // 
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	// JSON
	var intentData struct {
		Intent     string   `json:"intent"`
		Confidence float32  `json:"confidence"`
		Entities   []Entity `json:"entities"`
	}

	if err := json.Unmarshal([]byte(generateResp.Content), &intentData); err != nil {
		// JSON
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

// SentimentAnalysis 
// GenerateImage 
func (p *OpenAIProvider) GenerateImage(ctx context.Context, req ImageGenerateRequest) (*ImageGenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = "dall-e-3"
	}

	size := req.Size
	if size == "" {
		size = "1024x1024"
	}

	quality := req.Quality
	if quality == "" {
		quality = "standard"
	}

	count := req.Count
	if count == 0 {
		count = 1
	}

	requestData := map[string]interface{}{
		"model":   model,
		"prompt":  req.Prompt,
		"size":    size,
		"quality": quality,
		"n":       count,
	}

	if req.Style != "" {
		requestData["style"] = req.Style
	}

	respData, err := p.sendRequest(ctx, "/images/generations", requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	var openaiResp struct {
		Data []struct {
			URL           string `json:"url"`
			B64JSON       string `json:"b64_json"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respData, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	images := make([]GeneratedImage, len(openaiResp.Data))
	for i, img := range openaiResp.Data {
		images[i] = GeneratedImage{
			URL:           img.URL,
			Base64:        img.B64JSON,
			Width:         1024, // size
			Height:        1024,
			Format:        "png",
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	return &ImageGenerateResponse{
		Images: images,
		Usage: Usage{
			TotalTokens: len(req.Prompt) / 4, // 
		},
		RequestID: fmt.Sprintf("img_%d", time.Now().Unix()),
		Model:     model,
	}, nil
}

// AnalyzeImage 
func (p *OpenAIProvider) AnalyzeImage(ctx context.Context, req ImageAnalyzeRequest) (*ImageAnalyzeResponse, error) {
	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": req.Prompt,
				},
			},
		},
	}

	// 
	imageContent := map[string]interface{}{
		"type": "image_url",
	}

	if req.ImageURL != "" {
		imageContent["image_url"] = map[string]interface{}{
			"url":    req.ImageURL,
			"detail": req.Detail,
		}
	} else if req.ImageBase64 != "" {
		imageContent["image_url"] = map[string]interface{}{
			"url":    fmt.Sprintf("data:image/jpeg;base64,%s", req.ImageBase64),
			"detail": req.Detail,
		}
	}

	messages[0]["content"] = append(messages[0]["content"].([]map[string]interface{}), imageContent)

	requestData := map[string]interface{}{
		"model":      "gpt-4-vision-preview",
		"messages":   messages,
		"max_tokens": 1000,
	}

	respData, err := p.sendRequest(ctx, "/chat/completions", requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image: %w", err)
	}

	var openaiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respData, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &ImageAnalyzeResponse{
		Description: openaiResp.Choices[0].Message.Content,
		Usage: Usage{
			TotalTokens: openaiResp.Usage.TotalTokens,
		},
		RequestID: fmt.Sprintf("analyze_%d", time.Now().Unix()),
	}, nil
}

// EditImage 
func (p *OpenAIProvider) EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error) {
	// OpenAIAPImultipart/form-data
	// multipart

	size := req.Size
	if size == "" {
		size = "1024x1024"
	}

	count := req.Count
	if count == 0 {
		count = 1
	}

	requestData := map[string]interface{}{
		"prompt": req.Prompt,
		"size":   size,
		"n":      count,
	}

	respData, err := p.sendRequest(ctx, "/images/edits", requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to edit image: %w", err)
	}

	var openaiResp struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respData, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	images := make([]GeneratedImage, len(openaiResp.Data))
	for i, img := range openaiResp.Data {
		images[i] = GeneratedImage{
			URL:    img.URL,
			Base64: img.B64JSON,
			Width:  1024,
			Height: 1024,
			Format: "png",
		}
	}

	return &ImageEditResponse{
		Images: images,
		Usage: Usage{
			TotalTokens: len(req.Prompt) / 4,
		},
		RequestID: fmt.Sprintf("edit_%d", time.Now().Unix()),
		Model:     "dall-e-2",
	}, nil
}

func (p *OpenAIProvider) SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
	// 
	prompt := fmt.Sprintf(`JSON

"%s"

positive/negative/neutral-1.01.00.01.0
{
  "sentiment": "positive",
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

joyangersadnessfearsurprisedisgust`, req.Text)

	// Generate
	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3, // 
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf(": %w", err)
	}

	// JSON
	var sentimentData struct {
		Sentiment  string    `json:"sentiment"`
		Score      float32   `json:"score"`
		Confidence float32   `json:"confidence"`
		Emotions   []Emotion `json:"emotions"`
	}

	if err := json.Unmarshal([]byte(generateResp.Content), &sentimentData); err != nil {
		// JSON
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

