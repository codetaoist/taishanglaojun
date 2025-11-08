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

// OllamaService Ollama本地模型服务实现
type OllamaService struct {
	config     *models.ModelConfig
	httpClient *http.Client
	baseURL    string
}

// NewOllamaService 创建Ollama服务
func NewOllamaService(config *models.ModelConfig) *OllamaService {
	baseURL := "http://localhost:11434"
	if config.BaseURL() != "" {
		baseURL = config.BaseURL()
	}

	return &OllamaService{
		config:  config,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 默认超时时间
		},
	}
}

// Connect 连接到Ollama服务
func (s *OllamaService) Connect(ctx context.Context, config *models.ModelConfig) error {
	s.config = config
	s.baseURL = "http://localhost:11434"
	if config.BaseURL() != "" {
		s.baseURL = config.BaseURL()
	}
	s.httpClient.Timeout = 30 * time.Second // 默认超时时间

	// 验证连接
	return s.Health(ctx)
}

// Disconnect 断开连接
func (s *OllamaService) Disconnect(ctx context.Context) error {
	// Ollama是无状态连接，无需特殊处理
	return nil
}

// Health 检查服务健康状态
func (s *OllamaService) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
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
func (s *OllamaService) ListModels(ctx context.Context) ([]*models.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list models request: %w", err)
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

	var response struct {
		Models []struct {
			Name       string `json:"name"`
			Model      string `json:"model"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
			Digest     string `json:"digest"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode list models response: %w", err)
	}

	var modelInfos []*models.ModelInfo
	for _, model := range response.Models {
		// 确定模型类型
		modelType := "text"
		if strings.Contains(model.Name, "embed") {
			modelType = "embedding"
		} else if strings.Contains(model.Name, "image") {
			modelType = "image"
		} else if strings.Contains(model.Name, "audio") {
			modelType = "audio"
		}

		// 解析修改时间
		modifiedAt, _ := time.Parse(time.RFC3339, model.ModifiedAt)

		modelInfos = append(modelInfos, &models.ModelInfo{
			ID:          model.Name,
			Name:        model.Name,
			Type:        modelType,
			Provider:    "ollama",
			Description: fmt.Sprintf("Ollama model: %s", model.Name),
			CreatedAt:   modifiedAt,
		})
	}

	return modelInfos, nil
}

// GetModel 获取模型信息
func (s *OllamaService) GetModel(ctx context.Context, modelID string) (*models.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/show", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get model request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 请求体
	reqBody := map[string]string{
		"name": modelID,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

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
		License    string `json:"license"`
		Modelfile  string `json:"modelfile"`
		Parameters string `json:"parameters"`
		Template   string `json:"template"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode get model response: %w", err)
	}

	// 确定模型类型
	modelType := "text"
	if strings.Contains(modelID, "embed") {
		modelType = "embedding"
	} else if strings.Contains(modelID, "image") {
		modelType = "image"
	} else if strings.Contains(modelID, "audio") {
		modelType = "audio"
	}

	return &models.ModelInfo{
		ID:          modelID,
		Name:        modelID,
		Type:        modelType,
		Provider:    "ollama",
		Description: fmt.Sprintf("Ollama model: %s", modelID),
		CreatedAt:   time.Now(), // Ollama API不提供创建时间
	}, nil
}

// LoadModel 加载模型
func (s *OllamaService) LoadModel(ctx context.Context, modelID string) error {
	// Ollama模型是按需加载的，无需预加载
	return nil
}

// UnloadModel 卸载模型
func (s *OllamaService) UnloadModel(ctx context.Context, modelID string) error {
	// Ollama模型是按需加载的，无需卸载
	return nil
}

// GenerateText 生成文本
func (s *OllamaService) GenerateText(ctx context.Context, request *models.TextGenerationRequest) (*models.TextGenerationResponse, error) {
	// 转换请求格式
	ollamaReq := map[string]interface{}{
		"model": request.Model,
	}

	// 处理不同类型的输入
	if request.Messages != nil && len(request.Messages) > 0 {
		// 对话格式
		var messages []map[string]interface{}
		for _, msg := range request.Messages {
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
		ollamaReq["messages"] = messages
	} else if request.Prompt != "" {
		// 提示词格式
		ollamaReq["prompt"] = request.Prompt
	}

	// 添加可选参数
	if request.MaxTokens > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"num_predict": request.MaxTokens,
		}
	}
	if request.Temperature > 0 {
		if options, ok := ollamaReq["options"].(map[string]interface{}); ok {
			options["temperature"] = request.Temperature
		} else {
			ollamaReq["options"] = map[string]interface{}{
				"temperature": request.Temperature,
			}
		}
	}
	if request.TopP > 0 {
		if options, ok := ollamaReq["options"].(map[string]interface{}); ok {
			options["top_p"] = request.TopP
		} else {
			ollamaReq["options"] = map[string]interface{}{
				"top_p": request.TopP,
			}
		}
	}

	// 发送请求
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var endpoint string
	if request.Messages != nil && len(request.Messages) > 0 {
		endpoint = "/api/chat"
	} else {
		endpoint = "/api/generate"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
	var ollamaResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取生成的内容
	var content string
	if request.Messages != nil && len(request.Messages) > 0 {
		// 对话格式响应
		if message, ok := ollamaResp["message"].(map[string]interface{}); ok {
			if contentStr, ok := message["content"].(string); ok {
				content = contentStr
			}
		}
	} else {
		// 提示词格式响应
		if response, ok := ollamaResp["response"].(string); ok {
			content = response
		}
	}

	// 提取使用情况
	var usage models.Usage
	if promptEvalCount, ok := ollamaResp["prompt_eval_count"].(float64); ok {
		if evalCount, ok := ollamaResp["eval_count"].(float64); ok {
			usage = models.Usage{
				PromptTokens:     int(promptEvalCount),
				CompletionTokens: int(evalCount),
				TotalTokens:      int(promptEvalCount + evalCount),
			}
		}
	}

	return &models.TextGenerationResponse{
		Content: content,
		Usage:   usage,
		Model:   request.Model,
	}, nil
}

// GenerateTextStream 流式生成文本
func (s *OllamaService) GenerateTextStream(ctx context.Context, request *models.TextGenerationRequest) (<-chan *models.TextGenerationChunk, error) {
	// 转换请求格式
	ollamaReq := map[string]interface{}{
		"model":  request.Model,
		"stream": true,
	}

	// 处理不同类型的输入
	if request.Messages != nil && len(request.Messages) > 0 {
		// 对话格式
		var messages []map[string]interface{}
		for _, msg := range request.Messages {
			messages = append(messages, map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
		ollamaReq["messages"] = messages
	} else if request.Prompt != "" {
		// 提示词格式
		ollamaReq["prompt"] = request.Prompt
	}

	// 添加可选参数
	if request.MaxTokens > 0 {
		ollamaReq["options"] = map[string]interface{}{
			"num_predict": request.MaxTokens,
		}
	}
	if request.Temperature > 0 {
		if options, ok := ollamaReq["options"].(map[string]interface{}); ok {
			options["temperature"] = request.Temperature
		} else {
			ollamaReq["options"] = map[string]interface{}{
				"temperature": request.Temperature,
			}
		}
	}
	if request.TopP > 0 {
		if options, ok := ollamaReq["options"].(map[string]interface{}); ok {
			options["top_p"] = request.TopP
		} else {
			ollamaReq["options"] = map[string]interface{}{
				"top_p": request.TopP,
			}
		}
	}

	// 发送请求
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var endpoint string
	if request.Messages != nil && len(request.Messages) > 0 {
		endpoint = "/api/chat"
	} else {
		endpoint = "/api/generate"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
			if line == "" {
				continue
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				continue
			}

			// 检查是否完成
			if done, ok := chunk["done"].(bool); ok && done {
				break
			}

			// 提取内容
			var content string
			if request.Messages != nil && len(request.Messages) > 0 {
				// 对话格式响应
				if message, ok := chunk["message"].(map[string]interface{}); ok {
					if contentStr, ok := message["content"].(string); ok {
						content = contentStr
					}
				}
			} else {
				// 提示词格式响应
				if response, ok := chunk["response"].(string); ok {
					content = response
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
func (s *OllamaService) GenerateEmbedding(ctx context.Context, request *models.EmbeddingRequest) (*models.EmbeddingResponse, error) {
	// 转换请求格式
	ollamaReq := map[string]interface{}{
		"model":  request.Model,
		"prompt": request.Text,
	}

	// 发送请求
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
	var ollamaResp struct {
		Embedding []float64 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &models.EmbeddingResponse{
		Embedding: ollamaResp.Embedding,
		Model:     request.Model,
	}, nil
}

// GenerateEmbeddings 批量生成嵌入
func (s *OllamaService) GenerateEmbeddings(ctx context.Context, request *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
	// Ollama API不直接支持批量嵌入，需要逐个处理
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
func (s *OllamaService) CreateConversation(ctx context.Context, request *models.CreateConversationRequest) (*models.Conversation, error) {
	// 将map[string]interface{}转换为datatypes.JSON
	modelConfigJSON, err := json.Marshal(request.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model config: %w", err)
	}
	
	// Ollama不直接提供对话管理，这里返回一个模拟实现
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
func (s *OllamaService) GetConversation(ctx context.Context, conversationID string) (*models.Conversation, error) {
	// Ollama不直接提供对话管理，这里返回一个模拟实现
	return &models.Conversation{
		ID:        conversationID,
		Title:     "Sample Conversation",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// ListConversations 列出对话
func (s *OllamaService) ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	// Ollama不直接提供对话管理，这里返回一个模拟实现
	return []*models.Conversation{}, nil
}

// UpdateConversation 更新对话
func (s *OllamaService) UpdateConversation(ctx context.Context, conversationID string, request *models.UpdateConversationRequest) (*models.Conversation, error) {
	// Ollama不直接提供对话管理，这里返回一个模拟实现
	return &models.Conversation{
		ID:        conversationID,
		Title:     request.Title,
		UpdatedAt: time.Now(),
	}, nil
}

// DeleteConversation 删除对话
func (s *OllamaService) DeleteConversation(ctx context.Context, conversationID string) error {
	// Ollama不直接提供对话管理，这里返回一个模拟实现
	return nil
}

// AddMessage 添加消息
func (s *OllamaService) AddMessage(ctx context.Context, request *models.AddMessageRequest) (*models.Message, error) {
	// Ollama不直接提供消息管理，这里返回一个模拟实现
	return &models.Message{
		ID:             "msg_" + generateID(),
		ConversationID: request.ConversationID,
		Role:           request.Role,
		Content:        request.Content,
		CreatedAt:      time.Now(),
	}, nil
}

// GetMessages 获取消息
func (s *OllamaService) GetMessages(ctx context.Context, conversationID string, limit int, offset int) ([]*models.Message, error) {
	// Ollama不直接提供消息管理，这里返回一个模拟实现
	return []*models.Message{}, nil
}

// UpdateMessage 更新消息
func (s *OllamaService) UpdateMessage(ctx context.Context, messageID string, request *models.UpdateMessageRequest) (*models.Message, error) {
	// Ollama不直接提供消息管理，这里返回一个模拟实现
	return &models.Message{
		ID:        messageID,
		Content:   request.Content,
		UpdatedAt: time.Now(),
	}, nil
}

// DeleteMessage 删除消息
func (s *OllamaService) DeleteMessage(ctx context.Context, messageID string) error {
	// Ollama不直接提供消息管理，这里返回一个模拟实现
	return nil
}

// ExecuteTool 执行工具
func (s *OllamaService) ExecuteTool(ctx context.Context, request *models.ToolExecutionRequest) (*models.ToolExecutionResponse, error) {
	// TODO: 实现工具调用功能
	return &models.ToolExecutionResponse{
		Result: "Tool execution not implemented",
	}, nil
}

// CreateFineTuningJob 创建微调作业
func (s *OllamaService) CreateFineTuningJob(ctx context.Context, request *models.CreateFineTuningJobRequest) (*models.FineTuningJob, error) {
	// TODO: 实现微调作业创建
	return &models.FineTuningJob{
		ID:        "ft_" + generateID(),
		Model:     request.Model,
		Status:    "created",
		CreatedAt: time.Now(),
	}, nil
}

// GetFineTuningJob 获取微调作业
func (s *OllamaService) GetFineTuningJob(ctx context.Context, jobID string) (*models.FineTuningJob, error) {
	// TODO: 实现微调作业获取
	return &models.FineTuningJob{
		ID:     jobID,
		Status: "running",
	}, nil
}

// ListFineTuningJobs 列出微调作业
func (s *OllamaService) ListFineTuningJobs(ctx context.Context) ([]*models.FineTuningJob, error) {
	// TODO: 实现微调作业列表
	return []*models.FineTuningJob{}, nil
}

// CancelFineTuningJob 取消微调作业
func (s *OllamaService) CancelFineTuningJob(ctx context.Context, jobID string) error {
	// TODO: 实现微调作业取消
	return nil
}

// GetServiceInfo 获取服务信息
func (s *OllamaService) GetServiceInfo(ctx context.Context) (*models.ModelServiceInfo, error) {
	return &models.ModelServiceInfo{
		Provider:    "ollama",
		Version:     "v1",
		Description: "Ollama Local Model Service",
		Features: []string{
			"text-generation",
			"embedding",
			"chat",
			"streaming",
			"local-models",
		},
	}, nil
}