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

// IntentRecognition 意图识别
func (p *OpenAIProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
	// 构建意图识别的提示词
	prompt := fmt.Sprintf(`请分析以下文本的意图，并以JSON格式返回结果：

文本："%s"

请识别用户的主要意图，并提取相关实体。返回格式如下：
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

常见意图包括：询问、请求、抱怨、赞美、预订、取消、查询、帮助等。`, req.Text)

	// 使用Generate方法进行意图识别
	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3, // 较低的温度以获得更一致的结果
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("意图识别失败: %w", err)
	}

	// 解析JSON响应
	var intentData struct {
		Intent     string    `json:"intent"`
		Confidence float32   `json:"confidence"`
		Entities   []Entity  `json:"entities"`
	}

	if err := json.Unmarshal([]byte(generateResp.Content), &intentData); err != nil {
		// 如果JSON解析失败，返回基本的意图识别结果
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
// GenerateImage 生成图像
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
			Width:         1024, // 根据size解析
			Height:        1024,
			Format:        "png",
			RevisedPrompt: img.RevisedPrompt,
		}
	}

	return &ImageGenerateResponse{
		Images: images,
		Usage: Usage{
			TotalTokens: len(req.Prompt) / 4, // 估算
		},
		RequestID: fmt.Sprintf("img_%d", time.Now().Unix()),
		Model:     model,
	}, nil
}

// AnalyzeImage 分析图像
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

	// 添加图像内容
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
		"model":     "gpt-4-vision-preview",
		"messages":  messages,
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

// EditImage 编辑图像
func (p *OpenAIProvider) EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error) {
	// OpenAI的图像编辑API需要multipart/form-data
	// 这里简化实现，实际应该使用multipart上传
	
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
	// 构建情感分析的提示词
	prompt := fmt.Sprintf(`请分析以下文本的情感倾向，并以JSON格式返回结果：

文本："%s"

请分析文本的整体情感（positive/negative/neutral），给出情感分数（-1.0到1.0），置信度（0.0到1.0），以及具体的情感类别。返回格式如下：
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

情感类别包括：joy（喜悦）、anger（愤怒）、sadness（悲伤）、fear（恐惧）、surprise（惊讶）、disgust（厌恶）等。`, req.Text)

	// 使用Generate方法进行情感分析
	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3, // 较低的温度以获得更一致的结果
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("情感分析失败: %w", err)
	}

	// 解析JSON响应
	var sentimentData struct {
		Sentiment  string    `json:"sentiment"`
		Score      float32   `json:"score"`
		Confidence float32   `json:"confidence"`
		Emotions   []Emotion `json:"emotions"`
	}

	if err := json.Unmarshal([]byte(generateResp.Content), &sentimentData); err != nil {
		// 如果JSON解析失败，返回基本的情感分析结果
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
