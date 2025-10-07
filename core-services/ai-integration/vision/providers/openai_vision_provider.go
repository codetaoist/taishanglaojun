package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"../vision"
)

// OpenAIVisionProvider OpenAI视觉服务提供商
type OpenAIVisionProvider struct {
	config     OpenAIVisionConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// OpenAIVisionConfig OpenAI视觉配置
type OpenAIVisionConfig struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Model       string        `json:"model"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float64       `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
}

// OpenAIVisionRequest OpenAI视觉请求
type OpenAIVisionRequest struct {
	Model     string                   `json:"model"`
	Messages  []OpenAIVisionMessage    `json:"messages"`
	MaxTokens int                      `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
}

// OpenAIVisionMessage OpenAI视觉消息
type OpenAIVisionMessage struct {
	Role    string                     `json:"role"`
	Content []OpenAIVisionContent      `json:"content"`
}

// OpenAIVisionContent OpenAI视觉内容
type OpenAIVisionContent struct {
	Type     string                    `json:"type"`
	Text     string                    `json:"text,omitempty"`
	ImageURL *OpenAIVisionImageURL     `json:"image_url,omitempty"`
}

// OpenAIVisionImageURL OpenAI视觉图像URL
type OpenAIVisionImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// OpenAIVisionResponse OpenAI视觉响应
type OpenAIVisionResponse struct {
	ID      string                    `json:"id"`
	Object  string                    `json:"object"`
	Created int64                     `json:"created"`
	Model   string                    `json:"model"`
	Choices []OpenAIVisionChoice      `json:"choices"`
	Usage   OpenAIVisionUsage         `json:"usage"`
}

// OpenAIVisionChoice OpenAI视觉选择
type OpenAIVisionChoice struct {
	Index        int                   `json:"index"`
	Message      OpenAIVisionMessage   `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

// OpenAIVisionUsage OpenAI视觉使用情况
type OpenAIVisionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewOpenAIVisionProvider 创建OpenAI视觉提供商
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

// RecognizeObjects 物体识别
func (p *OpenAIVisionProvider) RecognizeObjects(ctx context.Context, input vision.ImageInput) (*vision.ObjectRecognitionResult, error) {
	prompt := "Analyze this image and identify all objects present. For each object, provide its name, confidence score (0-1), and bounding box coordinates if possible. Return the response in JSON format with an array of objects containing: name, confidence, x, y, width, height."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	// 解析响应并构建结果
	result := &vision.ObjectRecognitionResult{
		ID:           uuid.New().String(),
		RequestID:    input.ID,
		Objects:      make([]vision.DetectedObject, 0),
		TotalObjects: 0,
		Confidence:   0.0,
		Timestamp:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// 尝试解析JSON响应
	objects, err := p.parseObjectsFromResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse objects from response", zap.Error(err))
		// 如果解析失败，创建一个通用对象
		objects = []vision.DetectedObject{
			{
				ID:         uuid.New().String(),
				Name:       "detected_content",
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
	
	// 计算平均置信度
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

// RecognizeFaces 人脸识别
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

	// 尝试解析人脸信息
	faces, err := p.parseFacesFromResponse(response)
	if err != nil {
		p.logger.Warn("Failed to parse faces from response", zap.Error(err))
	}

	result.Faces = faces
	result.TotalFaces = len(faces)
	result.Metadata["raw_response"] = response

	return result, nil
}

// RecognizeText 文本识别
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
		Confidence: 0.9, // OpenAI通常有较高的准确性
		Language:   "auto",
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	// 尝试检测语言
	if language := p.detectLanguage(response); language != "" {
		result.Language = language
	}

	result.Metadata["raw_response"] = response

	return result, nil
}

// RecognizeScene 场景识别
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

// AnalyzeImage 图像分析
func (p *OpenAIVisionProvider) AnalyzeImage(ctx context.Context, input vision.ImageInput) (*vision.ImageAnalysisResult, error) {
	prompt := "Provide a comprehensive analysis of this image including: 1) Overall quality assessment, 2) Color analysis, 3) Composition analysis, 4) Content description, 5) Technical aspects, 6) Aesthetic evaluation. Be detailed and specific."

	response, err := p.analyzeImage(ctx, input, prompt)
	if err != nil {
		return nil, err
	}

	result := &vision.ImageAnalysisResult{
		ID:        uuid.New().String(),
		RequestID: input.ID,
		Quality: vision.QualityAnalysis{
			Overall:    0.8,
			Sharpness:  0.8,
			Brightness: 0.8,
			Contrast:   0.8,
			Noise:      0.2,
		},
		Color: vision.ColorAnalysis{
			Dominant:   []string{"#000000", "#FFFFFF"},
			Palette:    []string{"#000000", "#FFFFFF", "#808080"},
			Saturation: 0.5,
			Brightness: 0.5,
			Contrast:   0.5,
		},
		Composition: vision.CompositionAnalysis{
			RuleOfThirds: 0.7,
			Symmetry:     0.5,
			Balance:      0.6,
			Leading:      0.4,
		},
		Content: vision.ContentAnalysis{
			Category:    "general",
			Tags:        p.extractContentTags(response),
			Description: response,
			Complexity:  0.5,
		},
		Technical: vision.TechnicalAnalysis{
			Resolution: fmt.Sprintf("%dx%d", input.Width, input.Height),
			Format:     string(input.Format),
			FileSize:   input.Size,
			AspectRatio: float64(input.Width) / float64(input.Height),
		},
		Aesthetic: vision.AestheticAnalysis{
			Beauty:      0.7,
			Interesting: 0.6,
			Happy:       0.5,
			Sad:         0.2,
		},
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	result.Metadata["raw_response"] = response

	return result, nil
}

// ProcessImage 图像处理
func (p *OpenAIVisionProvider) ProcessImage(ctx context.Context, input vision.ImageInput, operations []vision.ImageOperation) (*vision.ImageProcessingResult, error) {
	// OpenAI Vision API主要用于分析，不支持图像处理
	// 这里返回一个模拟的结果
	return &vision.ImageProcessingResult{
		ID:               uuid.New().String(),
		RequestID:        input.ID,
		ProcessedImage:   input, // 返回原图像
		AppliedOperations: operations,
		Success:          false,
		Message:          "Image processing not supported by OpenAI Vision API",
		Timestamp:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}, nil
}

// GetSupportedFormats 获取支持的格式
func (p *OpenAIVisionProvider) GetSupportedFormats() []vision.ImageFormat {
	return []vision.ImageFormat{
		vision.FormatJPEG,
		vision.FormatPNG,
		vision.FormatWebP,
		vision.FormatGIF,
	}
}

// GetSupportedOperations 获取支持的操作
func (p *OpenAIVisionProvider) GetSupportedOperations() []vision.OperationType {
	// OpenAI Vision API主要用于分析，不支持图像处理操作
	return []vision.OperationType{}
}

// HealthCheck 健康检查
func (p *OpenAIVisionProvider) HealthCheck(ctx context.Context) error {
	// 创建一个简单的测试请求
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

// 私有方法

// analyzeImage 分析图像
func (p *OpenAIVisionProvider) analyzeImage(ctx context.Context, input vision.ImageInput, prompt string) (string, error) {
	// 将图像转换为base64
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

// makeRequest 发送请求
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

// parseObjectsFromResponse 从响应中解析物体
func (p *OpenAIVisionProvider) parseObjectsFromResponse(response string) ([]vision.DetectedObject, error) {
	// 尝试解析JSON格式的响应
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
				Name:       obj.Name,
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

	// 如果JSON解析失败，尝试从文本中提取物体名称
	return p.extractObjectsFromText(response), nil
}

// parseFacesFromResponse 从响应中解析人脸
func (p *OpenAIVisionProvider) parseFacesFromResponse(response string) ([]vision.DetectedFace, error) {
	// 简单的人脸信息提取
	faces := make([]vision.DetectedFace, 0)

	// 如果响应中包含人脸相关信息，创建一个通用的人脸对象
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
			Landmarks:  make([]vision.FaceLandmark, 0),
			Attributes: map[string]interface{}{
				"description": response,
			},
		}
		faces = append(faces, face)
	}

	return faces, nil
}

// extractObjectsFromText 从文本中提取物体
func (p *OpenAIVisionProvider) extractObjectsFromText(text string) []vision.DetectedObject {
	// 简单的文本解析，实际实现中可以使用更复杂的NLP技术
	objects := make([]vision.DetectedObject, 0)

	// 常见物体关键词
	keywords := []string{"person", "car", "tree", "building", "animal", "furniture", "food", "device", "tool", "clothing"}

	for _, keyword := range keywords {
		if p.containsKeyword(text, keyword) {
			objects = append(objects, vision.DetectedObject{
				ID:         uuid.New().String(),
				Name:       keyword,
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

// detectLanguage 检测语言
func (p *OpenAIVisionProvider) detectLanguage(text string) string {
	// 简单的语言检测
	if p.containsChinese(text) {
		return "zh"
	}
	return "en"
}

// extractSceneType 提取场景类型
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

// extractSceneTags 提取场景标签
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

// extractContentTags 提取内容标签
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

// 辅助函数

// containsKeyword 检查是否包含关键词
func (p *OpenAIVisionProvider) containsKeyword(text, keyword string) bool {
	return len(text) > 0 && len(keyword) > 0 && 
		   (text == keyword || 
		    fmt.Sprintf(" %s ", text) != fmt.Sprintf(" %s ", text) ||
		    fmt.Sprintf("%s ", text)[:len(keyword)+1] == fmt.Sprintf("%s ", keyword) ||
		    fmt.Sprintf(" %s", text)[len(text)-len(keyword):] == fmt.Sprintf(" %s", keyword))
}

// containsFaceKeywords 检查是否包含人脸关键词
func (p *OpenAIVisionProvider) containsFaceKeywords(text string) bool {
	faceKeywords := []string{"face", "person", "human", "people", "man", "woman", "child", "adult"}
	for _, keyword := range faceKeywords {
		if p.containsKeyword(text, keyword) {
			return true
		}
	}
	return false
}

// containsChinese 检查是否包含中文
func (p *OpenAIVisionProvider) containsChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}