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

// AzureProvider Azure OpenAI提供商实现
type AzureProvider struct {
	apiKey         string
	endpoint       string
	deploymentName string
	apiVersion     string
	httpClient     *http.Client
	logger         *zap.Logger
}

// AzureConfig Azure配置
type AzureConfig struct {
	APIKey         string `yaml:"api_key" json:"api_key"`
	Endpoint       string `yaml:"endpoint" json:"endpoint"`
	DeploymentName string `yaml:"deployment_name" json:"deployment_name"`
	APIVersion     string `yaml:"api_version" json:"api_version"`
	Timeout        int    `yaml:"timeout" json:"timeout"`
}

// NewAzureProvider 创建Azure提供商
func NewAzureProvider(config AzureConfig, logger *zap.Logger) *AzureProvider {
	if config.APIVersion == "" {
		config.APIVersion = "2023-05-15"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	return &AzureProvider{
		apiKey:         config.APIKey,
		endpoint:       config.Endpoint,
		deploymentName: config.DeploymentName,
		apiVersion:     config.APIVersion,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger: logger,
	}
}

// Chat 发送对话消息
func (p *AzureProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 构建Azure OpenAI API请求
	azureReq := map[string]interface{}{
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	if req.MaxTokens == 0 {
		azureReq["max_tokens"] = 1000
	}

	resp, err := p.sendRequest(ctx, "/chat/completions", azureReq)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var azureResp struct {
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

	if err := json.Unmarshal(resp, &azureResp); err != nil {
		return nil, fmt.Errorf("failed to parse Azure response: %w", err)
	}

	if len(azureResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in Azure response")
	}

	return &ChatResponse{
		Message: Message{
			Role:    azureResp.Choices[0].Message.Role,
			Content: azureResp.Choices[0].Message.Content,
		},
		Usage: Usage{
			PromptTokens:     azureResp.Usage.PromptTokens,
			CompletionTokens: azureResp.Usage.CompletionTokens,
			TotalTokens:      azureResp.Usage.TotalTokens,
		},
		SessionID: req.SessionID,
	}, nil
}

// Generate 生成文本
func (p *AzureProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// 将Generate请求转换为Chat请求
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
		Usage:   chatResp.Usage,
	}, nil
}

// Analyze 分析文本
func (p *AzureProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	// 构建分析提示
	prompt := fmt.Sprintf("请分析以下%s类型的内容：\n\n%s\n\n请提供详细的分析结果。", req.Type, req.Content)

	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3,
		MaxTokens:   1000,
	}

	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, err
	}

	return &AnalyzeResponse{
		Type:       req.Type,
		Confidence: 0.8,
		Result:     generateResp.Content,
		Details:    []string{generateResp.Content},
		Usage:      generateResp.Usage,
	}, nil
}

// Embed 获取文本嵌入向量
func (p *AzureProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Azure OpenAI的嵌入API
	embedReq := map[string]interface{}{
		"input": text,
	}

	resp, err := p.sendRequest(ctx, "/embeddings", embedReq)
	if err != nil {
		return nil, err
	}

	var embedResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embedResp.Data[0].Embedding, nil
}

// IntentRecognition 意图识别
func (p *AzureProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
	// 构建意图识别的提示词
	prompt := fmt.Sprintf(`请分析以下文本的用户意图，并以JSON格式返回结果：

文本："%s"

请识别用户的主要意图，给出置信度（0.0到1.0），并提取相关实体。返回格式如下：
{
  "intent": "intent_name",
  "confidence": 0.95,
  "entities": [
    {
      "name": "entity_name",
      "value": "entity_value",
      "type": "entity_type"
    }
  ]
}

常见意图包括：查询信息、执行操作、获取帮助、表达情感等。`, req.Text)

	// 使用Generate方法进行意图识别
	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3,
		MaxTokens:   500,
	}

	startTime := time.Now()
	generateResp, err := p.Generate(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("意图识别失败: %w", err)
	}

	// 解析JSON响应
	var intentData struct {
		Intent     string   `json:"intent"`
		Confidence float32  `json:"confidence"`
		Entities   []Entity `json:"entities"`
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
func (p *AzureProvider) SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
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
		Temperature: 0.3,
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

// GetName 获取提供商名称
func (p *AzureProvider) GetName() string {
	return "azure"
}

// GetModels 获取支持的模型列表
func (p *AzureProvider) GetModels() []string {
	return []string{
		"gpt-35-turbo",
		"gpt-35-turbo-16k",
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo",
	}
}

// sendRequest 发送HTTP请求到Azure OpenAI API
func (p *AzureProvider) sendRequest(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/openai/deployments/%s%s?api-version=%s", 
		p.endpoint, p.deploymentName, endpoint, p.apiVersion)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", p.apiKey)

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
		return nil, fmt.Errorf("Azure API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}