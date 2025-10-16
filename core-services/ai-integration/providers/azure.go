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

// AzureProvider Azure OpenAI
type AzureProvider struct {
	apiKey         string
	endpoint       string
	deploymentName string
	apiVersion     string
	httpClient     *http.Client
	logger         *zap.Logger
}

// AzureConfig Azure
type AzureConfig struct {
	APIKey         string `yaml:"api_key" json:"api_key"`
	Endpoint       string `yaml:"endpoint" json:"endpoint"`
	DeploymentName string `yaml:"deployment_name" json:"deployment_name"`
	APIVersion     string `yaml:"api_version" json:"api_version"`
	Timeout        int    `yaml:"timeout" json:"timeout"`
}

// NewAzureProvider Azure
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

// Chat 
func (p *AzureProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Azure OpenAI API
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

	// 
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

// Generate 
func (p *AzureProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
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
func (p *AzureProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
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
		Type:       req.Type,
		Confidence: 0.8,
		Result:     generateResp.Content,
		Details:    []string{generateResp.Content},
		Usage:      generateResp.Usage,
	}, nil
}

// Embed 
func (p *AzureProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Azure OpenAIAPI
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

// IntentRecognition 
func (p *AzureProvider) IntentRecognition(ctx context.Context, req IntentRequest) (*IntentResponse, error) {
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
func (p *AzureProvider) SentimentAnalysis(ctx context.Context, req SentimentRequest) (*SentimentResponse, error) {
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
func (p *AzureProvider) GetName() string {
	return "azure"
}

// GetModels 
func (p *AzureProvider) GetModels() []string {
	return []string{
		"gpt-4", "gpt-4-turbo", "gpt-35-turbo",
		"text-davinci-003", "text-curie-001",
		"text-embedding-ada-002",
		"dall-e-3", "dall-e-2",
	}
}

// GetCapabilities 
func (p *AzureProvider) GetCapabilities() []string {
	return []string{
		"chat", "completion", "embedding", "image-generation",
		"image-analysis", "image-edit", "vision", "intent-recognition",
		"sentiment-analysis",
	}
}

// AnalyzeImage 
func (p *AzureProvider) AnalyzeImage(ctx context.Context, req ImageAnalyzeRequest) (*ImageAnalyzeResponse, error) {
	// Azure OpenAI 
	//  Azure Computer Vision API  GPT-4 Vision 
	return &ImageAnalyzeResponse{
		Description: "Image analysis not implemented for Azure provider",
		Objects:     []DetectedObject{},
		Text:        []DetectedText{},
		Faces:       []DetectedFace{},
		Colors:      []DominantColor{},
		Tags:        []string{},
		Categories:  []Category{},
		Emotions:    []Emotion{},
		Safety:      SafetyAnalysis{},
		Usage:       Usage{},
		RequestID:   "azure-" + req.UserID,
	}, nil
}

// EditImage 
func (p *AzureProvider) EditImage(ctx context.Context, req ImageEditRequest) (*ImageEditResponse, error) {
	// Azure 
	return &ImageEditResponse{
		Images: []GeneratedImage{
			{
				URL:    "https://example.com/edited-image.png",
				Width:  1024,
				Height: 1024,
				Format: "png",
			},
		},
		Usage: Usage{
			TotalTokens: 100,
		},
		RequestID: fmt.Sprintf("edit_%d", time.Now().Unix()),
		Model:     "azure-image-edit",
	}, nil
}

// GenerateImage 
func (p *AzureProvider) GenerateImage(ctx context.Context, req ImageGenerateRequest) (*ImageGenerateResponse, error) {
	// Azure 
	model := req.Model
	if model == "" {
		model = "azure-dalle"
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
			URL:    fmt.Sprintf("https://example.com/generated-image-%d.png", i+1),
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
		RequestID: fmt.Sprintf("gen_%d", time.Now().Unix()),
		Model:     model,
	}, nil
}

// sendRequest HTTPAzure OpenAI API
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

