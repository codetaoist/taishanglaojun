package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// BaiduProvider 百度AI提供商实现
type BaiduProvider struct {
	apiKey      string
	secretKey   string
	baseURL     string
	accessToken string
	tokenExpiry time.Time
	httpClient  *http.Client
	logger      *zap.Logger
}

// BaiduConfig 百度配置
type BaiduConfig struct {
	APIKey    string `yaml:"api_key" json:"api_key"`
	SecretKey string `yaml:"secret_key" json:"secret_key"`
	BaseURL   string `yaml:"base_url" json:"base_url"`
	Timeout   int    `yaml:"timeout" json:"timeout"`
}

// NewBaiduProvider 创建百度提供商
func NewBaiduProvider(config BaiduConfig, logger *zap.Logger) *BaiduProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://aip.baidubce.com"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	return &BaiduProvider{
		apiKey:    config.APIKey,
		secretKey: config.SecretKey,
		baseURL:   config.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
		logger: logger,
	}
}

// getAccessToken 获取访问令牌
func (p *BaiduProvider) getAccessToken(ctx context.Context) (string, error) {
	// 如果token还有效，直接返回
	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		return p.accessToken, nil
	}

	// 获取新的access token
	tokenURL := fmt.Sprintf("%s/oauth/2.0/token", p.baseURL)
	
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", p.apiKey)
	data.Set("client_secret", p.secretKey)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	p.accessToken = tokenResp.AccessToken
	p.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // 提前5分钟过期

	return p.accessToken, nil
}

// Chat 发送对话消息
func (p *BaiduProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 构建文心一言API请求
	baiduReq := map[string]interface{}{
		"messages": req.Messages,
	}

	if req.Temperature > 0 {
		baiduReq["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		baiduReq["max_output_tokens"] = req.MaxTokens
	}

	resp, err := p.sendRequest(ctx, "/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions", token, baiduReq)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var baiduResp struct {
		Result string `json:"result"`
		Usage  struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}

	if err := json.Unmarshal(resp, &baiduResp); err != nil {
		return nil, fmt.Errorf("failed to parse Baidu response: %w", err)
	}

	if baiduResp.ErrorCode != 0 {
		return nil, fmt.Errorf("Baidu API error (code %d): %s", baiduResp.ErrorCode, baiduResp.ErrorMsg)
	}

	return &ChatResponse{
		Message: Message{
			Role:    "assistant",
			Content: baiduResp.Result,
		},
		Usage: Usage{
			PromptTokens:     baiduResp.Usage.PromptTokens,
			CompletionTokens: baiduResp.Usage.CompletionTokens,
			TotalTokens:      baiduResp.Usage.TotalTokens,
		},
		SessionID: req.SessionID,
	}, nil
}

// Generate 生成文本
func (p *BaiduProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
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
func (p *BaiduProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
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
		Type:    req.Type,
		Details: []string{generateResp.Content},
		Result:  generateResp.Content,
		Usage:   generateResp.Usage,
	}, nil
}

// Embed 获取文本嵌入向量
func (p *BaiduProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 百度文本向量化API
	embedReq := map[string]interface{}{
		"input": []string{text},
	}

	resp, err := p.sendRequest(ctx, "/rpc/2.0/ai_custom/v1/wenxinworkshop/embeddings/embedding-v1", token, embedReq)
	if err != nil {
		return nil, err
	}

	var embedResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}

	if err := json.Unmarshal(resp, &embedResp); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if embedResp.ErrorCode != 0 {
		return nil, fmt.Errorf("Baidu embedding error (code %d): %s", embedResp.ErrorCode, embedResp.ErrorMsg)
	}

	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embedResp.Data[0].Embedding, nil
}

// IntentRecognition 意图识别
func (p *BaiduProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
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
func (p *BaiduProvider) SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// 百度情感倾向分析API
	sentimentReq := map[string]interface{}{
		"text": req.Text,
	}

	resp, err := p.sendRequest(ctx, "/rpc/2.0/nlp/v1/sentiment_classify", token, sentimentReq)
	if err != nil {
		// 如果专用API失败，使用通用生成方法
		return p.fallbackSentimentAnalysis(ctx, req)
	}

	var sentimentResp struct {
		Items []struct {
			Sentiment   int     `json:"sentiment"`
			Confidence  float32 `json:"confidence"`
			Positive    float32 `json:"positive_prob"`
			Negative    float32 `json:"negative_prob"`
		} `json:"items"`
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	}

	if err := json.Unmarshal(resp, &sentimentResp); err != nil {
		return p.fallbackSentimentAnalysis(ctx, req)
	}

	if sentimentResp.ErrorCode != 0 {
		return p.fallbackSentimentAnalysis(ctx, req)
	}

	if len(sentimentResp.Items) == 0 {
		return p.fallbackSentimentAnalysis(ctx, req)
	}

	item := sentimentResp.Items[0]
	
	// 转换百度的情感分类（0=负面，1=中性，2=正面）
	var sentiment string
	var score float32
	
	switch item.Sentiment {
	case 0:
		sentiment = "negative"
		score = -item.Negative
	case 1:
		sentiment = "neutral"
		score = 0.0
	case 2:
		sentiment = "positive"
		score = item.Positive
	default:
		sentiment = "neutral"
		score = 0.0
	}

	return &SentimentResponse{
		Sentiment:  sentiment,
		Score:      score,
		Confidence: item.Confidence,
		Emotions:   []Emotion{}, // 百度API不提供详细情感分类
		Usage: Usage{
			PromptTokens:     len(req.Text) / 4, // 估算
			CompletionTokens: 10,                // 估算
			TotalTokens:      len(req.Text)/4 + 10,
		},
	}, nil
}

// fallbackSentimentAnalysis 备用情感分析方法
func (p *BaiduProvider) fallbackSentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
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
func (p *BaiduProvider) GetName() string {
	return "baidu"
}

// GetModels 获取支持的模型列表
func (p *BaiduProvider) GetModels() []string {
	return []string{
		"ERNIE-Bot",
		"ERNIE-Bot-turbo",
		"ERNIE-Bot-4",
		"BLOOMZ-7B",
		"Llama-2-7b-chat",
		"Llama-2-13b-chat",
		"Llama-2-70b-chat",
	}
}

// sendRequest 发送HTTP请求到百度AI API
func (p *BaiduProvider) sendRequest(ctx context.Context, endpoint string, accessToken string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s?access_token=%s", p.baseURL, endpoint, accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("Baidu API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}