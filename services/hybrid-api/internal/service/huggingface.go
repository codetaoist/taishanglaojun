package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// HuggingFaceService Hugging Face模型服务实现
type HuggingFaceService struct {
	config     *models.ModelConfig
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewHuggingFaceService 创建Hugging Face服务
func NewHuggingFaceService(config *models.ModelConfig) *HuggingFaceService {
	baseURL := "https://api-inference.huggingface.co"
	if config.BaseURL() != "" {
		baseURL = config.BaseURL()
	}

	return &HuggingFaceService{
		config:  config,
		baseURL: baseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 默认超时时间
		},
	}
}

// Connect 连接到Hugging Face服务
func (s *HuggingFaceService) Connect(ctx context.Context, config *models.ModelConfig) error {
	s.config = config
	s.baseURL = "https://api-inference.huggingface.co"
	if config.BaseURL() != "" {
		s.baseURL = config.BaseURL()
	}
	s.apiKey = config.APIKey
	s.httpClient.Timeout = 30 * time.Second // 默认超时时间

	// 验证连接
	return s.Health(ctx)
}

// Disconnect 断开连接
func (s *HuggingFaceService) Disconnect(ctx context.Context) error {
	// Hugging Face是无状态连接，无需特殊处理
	return nil
}

// Health 检查服务健康状态
func (s *HuggingFaceService) Health(ctx context.Context) error {
	// 检查API是否可访问
	req, err := http.NewRequestWithContext(ctx, "GET", "https://huggingface.co/api/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	return nil
}

// ListModels 列出可用模型
func (s *HuggingFaceService) ListModels(ctx context.Context) ([]*models.ModelInfo, error) {
	// 获取热门模型列表
	req, err := http.NewRequestWithContext(ctx, "GET", "https://huggingface.co/api/models?limit=100&sort=downloads&direction=-1", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list models request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list models failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response []struct {
		ID           string   `json:"id"`
		ModelId      string   `json:"modelId"`
		Author       string   `json:"author"`
		LastModified string   `json:"lastModified"`
		Tags         []string `json:"tags"`
		PipelineTag  string   `json:"pipeline_tag"`
		LibraryName  string   `json:"library_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode list models response: %w", err)
	}

	var modelInfos []*models.ModelInfo
	for _, model := range response {
		// 确定模型类型
		modelType := "text"
		if model.PipelineTag == "feature-extraction" || model.PipelineTag == "sentence-similarity" {
			modelType = "embedding"
		} else if model.PipelineTag == "image-classification" || model.PipelineTag == "text-to-image" {
			modelType = "image"
		} else if model.PipelineTag == "automatic-speech-recognition" || model.PipelineTag == "text-to-speech" {
			modelType = "audio"
		}

		modelID := model.ID
		if modelID == "" {
			modelID = model.ModelId
		}

		modelInfos = append(modelInfos, &models.ModelInfo{
			ID:          modelID,
			Name:        modelID,
			Type:        modelType,
			Provider:    "huggingface",
			Description: fmt.Sprintf("Hugging Face model: %s by %s", modelID, model.Author),
			CreatedAt:   time.Now(), // HF API不提供创建时间
		})
	}

	return modelInfos, nil
}

// GetModel 获取模型信息
func (s *HuggingFaceService) GetModel(ctx context.Context, modelID string) (*models.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://huggingface.co/api/models/"+modelID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get model request: %w", err)
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get model failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		ID           string   `json:"id"`
		ModelId      string   `json:"modelId"`
		Author       string   `json:"author"`
		LastModified string   `json:"lastModified"`
		Tags         []string `json:"tags"`
		PipelineTag  string   `json:"pipeline_tag"`
		LibraryName  string   `json:"library_name"`
		CardData     struct {
			Language []string `json:"language"`
			Tags     []string `json:"tags"`
		} `json:"cardData"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode get model response: %w", err)
	}

	// 确定模型类型
	modelType := "text"
	if response.PipelineTag == "feature-extraction" || response.PipelineTag == "sentence-similarity" {
		modelType = "embedding"
	} else if response.PipelineTag == "image-classification" || response.PipelineTag == "text-to-image" {
		modelType = "image"
	} else if response.PipelineTag == "automatic-speech-recognition" || response.PipelineTag == "text-to-speech" {
		modelType = "audio"
	}

	modelIDResult := response.ID
	if modelIDResult == "" {
		modelIDResult = response.ModelId
	}

	return &models.ModelInfo{
		ID:          modelIDResult,
		Name:        modelIDResult,
		Type:        modelType,
		Provider:    "huggingface",
		Description: fmt.Sprintf("Hugging Face model: %s by %s", modelIDResult, response.Author),
		CreatedAt:   time.Now(), // HF API不提供创建时间
	}, nil
}

// LoadModel 加载模型
func (s *HuggingFaceService) LoadModel(ctx context.Context, modelID string) error {
	// Hugging Face模型是按需加载的，无需预加载
	return nil
}

// UnloadModel 卸载模型
func (s *HuggingFaceService) UnloadModel(ctx context.Context, modelID string) error {
	// Hugging Face模型是按需加载的，无需卸载
	return nil
}

// GenerateText 生成文本
func (s *HuggingFaceService) GenerateText(ctx context.Context, request *models.TextGenerationRequest) (*models.TextGenerationResponse, error) {
	// 转换请求格式
	hfReq := map[string]interface{}{
		"inputs": request.Prompt,
	}

	// 添加可选参数
	options := make(map[string]interface{})
	if request.MaxTokens > 0 {
		options["max_new_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		options["temperature"] = request.Temperature
	}
	if request.TopP > 0 {
		options["top_p"] = request.TopP
	}

	if len(options) > 0 {
		hfReq["parameters"] = options
	}

	// 发送请求
	reqBody, err := json.Marshal(hfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/models/"+request.Model, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate text failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var hfResp []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(hfResp) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	// 提取生成的内容
	var content string
	if generatedText, ok := hfResp[0]["generated_text"].(string); ok {
		// 移除原始提示词
		if strings.HasPrefix(generatedText, request.Prompt) {
			content = strings.TrimPrefix(generatedText, request.Prompt)
		} else {
			content = generatedText
		}
	} else {
		return nil, fmt.Errorf("invalid response format")
	}

	return &models.TextGenerationResponse{
		Content: content,
		Model:   request.Model,
	}, nil
}

// GenerateTextStream 流式生成文本
func (s *HuggingFaceService) GenerateTextStream(ctx context.Context, request *models.TextGenerationRequest) (<-chan *models.TextGenerationChunk, error) {
	// 转换请求格式
	hfReq := map[string]interface{}{
		"inputs": request.Prompt,
		"stream": true,
	}

	// 添加可选参数
	options := make(map[string]interface{})
	if request.MaxTokens > 0 {
		options["max_new_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		options["temperature"] = request.Temperature
	}
	if request.TopP > 0 {
		options["top_p"] = request.TopP
	}

	if len(options) > 0 {
		hfReq["parameters"] = options
	}

	// 发送请求
	reqBody, err := json.Marshal(hfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/models/"+request.Model, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("generate text stream failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 创建输出通道
	ch := make(chan *models.TextGenerationChunk)

	// 启动goroutine处理流式响应
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}

			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			if data == "[DONE]" {
				break
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			// 提取内容
			var content string
			if token, ok := chunk["token"].(map[string]interface{}); ok {
				if text, ok := token["text"].(string); ok {
					content = text
				}
			}

			ch <- &models.TextGenerationChunk{
				Content: content,
				Done:    false,
			}
		}

		// 发送结束标记
		ch <- &models.TextGenerationChunk{
			Content: "",
			Done:    true,
		}
	}()

	return ch, nil
}

// GenerateEmbedding 生成嵌入
func (s *HuggingFaceService) GenerateEmbedding(ctx context.Context, request *models.EmbeddingRequest) (*models.EmbeddingResponse, error) {
	// 转换请求格式
	hfReq := map[string]interface{}{
		"inputs": request.Text,
	}

	// 发送请求
	reqBody, err := json.Marshal(hfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/models/"+request.Model, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate embedding failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var hfResp []float64
	if err := json.NewDecoder(resp.Body).Decode(&hfResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(hfResp) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	return &models.EmbeddingResponse{
		Embedding: hfResp,
		Model:     request.Model,
	}, nil
}

// GenerateEmbeddings 批量生成嵌入
func (s *HuggingFaceService) GenerateEmbeddings(ctx context.Context, request *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
	// Hugging Face API不直接支持批量嵌入，需要逐个处理
	embeddings := make([][]float64, len(request.Texts))
	
	for i, text := range request.Texts {
		singleRequest := &models.EmbeddingRequest{
			Text:  text,
			Model: request.Model,
			User:  request.User,
		}
		
		response, err := s.GenerateEmbedding(ctx, singleRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		
		embeddings[i] = response.Embedding
	}

	return &models.EmbeddingsResponse{
		Embeddings: embeddings,
		Model:      request.Model,
	}, nil
}

// CreateConversation 创建对话
func (s *HuggingFaceService) CreateConversation(ctx context.Context, request *models.CreateConversationRequest) (*models.Conversation, error) {
	// 将map[string]interface{}转换为datatypes.JSON
	modelConfigJSON, err := json.Marshal(request.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model config: %w", err)
	}
	
	// Hugging Face不直接提供对话管理，这里返回一个模拟实现
	return &models.Conversation{
		ID:          "conv_" + generateID(),
		Title:       request.Title,
		ModelConfig: modelConfigJSON,
		Messages:    request.Messages,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetConversation 获取对话
func (s *HuggingFaceService) GetConversation(ctx context.Context, conversationID string) (*models.Conversation, error) {
	// Hugging Face不直接提供对话管理，这里返回一个模拟实现
	return &models.Conversation{
		ID:        conversationID,
		Title:     "Sample Conversation",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// ListConversations 列出对话
func (s *HuggingFaceService) ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	// Hugging Face不直接提供对话管理，这里返回一个模拟实现
	return []*models.Conversation{}, nil
}

// UpdateConversation 更新对话
func (s *HuggingFaceService) UpdateConversation(ctx context.Context, conversationID string, request *models.UpdateConversationRequest) (*models.Conversation, error) {
	// Hugging Face不直接提供对话管理，这里返回一个模拟实现
	return &models.Conversation{
		ID:        conversationID,
		Title:     request.Title,
		UpdatedAt: time.Now(),
	}, nil
}

// DeleteConversation 删除对话
func (s *HuggingFaceService) DeleteConversation(ctx context.Context, conversationID string) error {
	// Hugging Face不直接提供对话管理，这里返回一个模拟实现
	return nil
}

// AddMessage 添加消息
func (s *HuggingFaceService) AddMessage(ctx context.Context, request *models.AddMessageRequest) (*models.Message, error) {
	// Hugging Face不直接提供消息管理，这里返回一个模拟实现
	return &models.Message{
		ID:             "msg_" + generateID(),
		ConversationID: request.ConversationID,
		Role:           request.Role,
		Content:        request.Content,
		CreatedAt:      time.Now(),
	}, nil
}

// GetMessages 获取消息
func (s *HuggingFaceService) GetMessages(ctx context.Context, conversationID string, limit int, offset int) ([]*models.Message, error) {
	// Hugging Face不直接提供消息管理，这里返回一个模拟实现
	return []*models.Message{}, nil
}

// UpdateMessage 更新消息
func (s *HuggingFaceService) UpdateMessage(ctx context.Context, messageID string, request *models.UpdateMessageRequest) (*models.Message, error) {
	// Hugging Face不直接提供消息管理，这里返回一个模拟实现
	return &models.Message{
		ID:        messageID,
		Content:   request.Content,
		UpdatedAt: time.Now(),
	}, nil
}

// DeleteMessage 删除消息
func (s *HuggingFaceService) DeleteMessage(ctx context.Context, messageID string) error {
	// Hugging Face不直接提供消息管理，这里返回一个模拟实现
	return nil
}

// ExecuteTool 执行工具
func (s *HuggingFaceService) ExecuteTool(ctx context.Context, request *models.ToolExecutionRequest) (*models.ToolExecutionResponse, error) {
	// TODO: 实现工具调用功能
	return &models.ToolExecutionResponse{
		Result: "Tool execution not implemented",
	}, nil
}

// CreateFineTuningJob 创建微调作业
func (s *HuggingFaceService) CreateFineTuningJob(ctx context.Context, request *models.CreateFineTuningJobRequest) (*models.FineTuningJob, error) {
	// TODO: 实现微调作业创建
	return &models.FineTuningJob{
		ID:        "ft_" + generateID(),
		Model:     request.Model,
		Status:    "created",
		CreatedAt: time.Now(),
	}, nil
}

// GetFineTuningJob 获取微调作业
func (s *HuggingFaceService) GetFineTuningJob(ctx context.Context, jobID string) (*models.FineTuningJob, error) {
	// TODO: 实现微调作业获取
	return &models.FineTuningJob{
		ID:     jobID,
		Status: "running",
	}, nil
}

// ListFineTuningJobs 列出微调作业
func (s *HuggingFaceService) ListFineTuningJobs(ctx context.Context) ([]*models.FineTuningJob, error) {
	// TODO: 实现微调作业列表
	return []*models.FineTuningJob{}, nil
}

// CancelFineTuningJob 取消微调作业
func (s *HuggingFaceService) CancelFineTuningJob(ctx context.Context, jobID string) error {
	// TODO: 实现微调作业取消
	return nil
}

// GetServiceInfo 获取服务信息
func (s *HuggingFaceService) GetServiceInfo(ctx context.Context) (*models.ModelServiceInfo, error) {
	return &models.ModelServiceInfo{
		Provider:    "huggingface",
		Version:     "v1",
		Description: "Hugging Face Inference API Service",
		Features: []string{
			"text-generation",
			"embedding",
			"image-generation",
			"audio-processing",
			"streaming",
		},
	}, nil
}