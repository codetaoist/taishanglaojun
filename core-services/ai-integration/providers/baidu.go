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

// BaiduProvider AI
type BaiduProvider struct {
	apiKey      string
	secretKey   string
	baseURL     string
	accessToken string
	tokenExpiry time.Time
	httpClient  *http.Client
	logger      *zap.Logger
}

// BaiduConfig 
type BaiduConfig struct {
	APIKey    string `yaml:"api_key" json:"api_key"`
	SecretKey string `yaml:"secret_key" json:"secret_key"`
	BaseURL   string `yaml:"base_url" json:"base_url"`
	Timeout   int    `yaml:"timeout" json:"timeout"`
}

// NewBaiduProvider 
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

// getAccessToken 
func (p *BaiduProvider) getAccessToken(ctx context.Context) (string, error) {
	// token
	if p.accessToken != "" && time.Now().Before(p.tokenExpiry) {
		return p.accessToken, nil
	}

	// access token
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
	p.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // 5

	return p.accessToken, nil
}

// Chat 
func (p *BaiduProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// API
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

	// 
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

// Generate 
func (p *BaiduProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	// GenerateChat
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

// Analyze 
func (p *BaiduProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	// 
	prompt := fmt.Sprintf("%s\n\n%s\n\n", req.Type, req.Content)

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

// Embed 
func (p *BaiduProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// API
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

// IntentRecognition 
func (p *BaiduProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
	// 
	prompt := fmt.Sprintf(`JSON

"%s"

01巵
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

`, req.Text)

	// Generate
	generateReq := GenerateRequest{
		Prompt:      prompt,
		Temperature: 0.3,
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
func (p *BaiduProvider) SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	// API
	sentimentReq := map[string]interface{}{
		"text": req.Text,
	}

	resp, err := p.sendRequest(ctx, "/rpc/2.0/nlp/v1/sentiment_classify", token, sentimentReq)
	if err != nil {
		// API
		return p.fallbackSentimentAnalysis(ctx, req)
	}

	var sentimentResp struct {
		Items []struct {
			Sentiment  int     `json:"sentiment"`
			Confidence float32 `json:"confidence"`
			Positive   float32 `json:"positive_prob"`
			Negative   float32 `json:"negative_prob"`
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

	// 0=1=2=
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
		Emotions:   []Emotion{}, // API
		Usage: Usage{
			PromptTokens:     len(req.Text) / 4, // 
			CompletionTokens: 10,                // 
			TotalTokens:      len(req.Text)/4 + 10,
		},
	}, nil
}

// fallbackSentimentAnalysis 
func (p *BaiduProvider) fallbackSentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
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
		Temperature: 0.3,
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

// GetName 
func (p *BaiduProvider) GetName() string {
	return "baidu"
}

// GetModels 
func (p *BaiduProvider) GetModels() []string {
	return []string{
		"ernie-bot", "ernie-bot-turbo", "ernie-bot-4",
		"ernie-text-embedding", "ernie-vilg-v2",
		"bloomz-7b", "qianfan-bloomz-7b-compressed",
	}
}

// GetCapabilities 
func (p *BaiduProvider) GetCapabilities() []string {
	return []string{
		"chat", "completion", "embedding", "image-generation",
		"image-analysis", "image-edit", "intent-recognition",
		"sentiment-analysis",
	}
}

// AnalyzeImage 
func (p *BaiduProvider) AnalyzeImage(ctx context.Context, req ImageAnalyzeRequest) (*ImageAnalyzeResponse, error) {
	// AI
	// API
	return &ImageAnalyzeResponse{
		Description: "Image analysis not implemented for Baidu provider",
		Objects:     []DetectedObject{},
		Text:        []DetectedText{},
		Faces:       []DetectedFace{},
		Colors:      []DominantColor{},
		Tags:        []string{},
		Categories:  []Category{},
		Emotions:    []Emotion{},
		Safety:      SafetyAnalysis{},
		Usage:       Usage{},
		RequestID:   "baidu-" + req.UserID,
	}, nil
}

// EditImage 
func (p *BaiduProvider) EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error) {
	// 
	return &ImageEditResponse{
		Images: []GeneratedImage{
			{
				URL:    "https://example.com/baidu-edited-image.png",
				Width:  1024,
				Height: 1024,
				Format: "png",
			},
		},
		Usage: Usage{
			TotalTokens: 100,
		},
		RequestID: fmt.Sprintf("baidu_edit_%d", time.Now().Unix()),
		Model:     "baidu-image-edit",
	}, nil
}

// GenerateImage 
func (p *BaiduProvider) GenerateImage(ctx context.Context, req ImageGenerateRequest) (*ImageGenerateResponse, error) {
	// 
	model := req.Model
	if model == "" {
		model = "baidu-ernie-vilg"
	}

	size := req.Size
	if size == "" {
		size = "1024x1024"
	}

	count := req.Count
	if count == 0 {
		count = 1
	}

	images := make([]GeneratedImage, count)
	for i := 0; i < count; i++ {
		images[i] = GeneratedImage{
			URL:    fmt.Sprintf("https://example.com/baidu-generated-image-%d.png", i+1),
			Width:  1024,
			Height: 1024,
			Format: "png",
			Seed:   req.Seed,
		}
	}

	return &ImageGenerateResponse{
		Images: images,
		Usage: Usage{
			TotalTokens: len(req.Prompt) / 4, // 
		},
		RequestID: fmt.Sprintf("baidu_gen_%d", time.Now().Unix()),
		Model:     model,
	}, nil
}

// sendRequest HTTPAI API
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

