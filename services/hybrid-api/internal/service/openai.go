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

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// OpenAIService OpenAI模型服务实现
type OpenAIService struct {
	config          *models.ModelConfig
	httpClient      *http.Client
	baseURL         string
	apiKey          string
	conversationDAO *dao.ConversationDAO
	messageDAO      *dao.MessageDAO
}

// NewOpenAIService 创建OpenAI服务
func NewOpenAIService(config *models.ModelConfig) *OpenAIService {
	baseURL := "https://api.openai.com/v1"
	if config.BaseURL() != "" {
		baseURL = config.BaseURL()
	}

	return &OpenAIService{
		config:  config,
		baseURL: baseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second, // 默认超时时间
		},
	}
}

// SetDAOs 设置对话和消息DAO
func (s *OpenAIService) SetDAOs(conversationDAO *dao.ConversationDAO, messageDAO *dao.MessageDAO) {
	s.conversationDAO = conversationDAO
	s.messageDAO = messageDAO
}

// Connect 连接到OpenAI服务
func (s *OpenAIService) Connect(ctx context.Context, config *models.ModelConfig) error {
	s.config = config
	s.baseURL = "https://api.openai.com/v1"
	if config.BaseURL() != "" {
		s.baseURL = config.BaseURL()
	}
	s.apiKey = config.APIKey
	s.httpClient.Timeout = 30 * time.Second // 默认超时时间

	// 验证连接
	return s.Health(ctx)
}

// Disconnect 断开连接
func (s *OpenAIService) Disconnect(ctx context.Context) error {
	// OpenAI是无状态连接，无需特殊处理
	return nil
}

// Health 检查服务健康状态
func (s *OpenAIService) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

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
func (s *OpenAIService) ListModels(ctx context.Context) ([]*models.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list models request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

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
		Object string `json:"object"`
		Data   []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode list models response: %w", err)
	}

	var modelInfos []*models.ModelInfo
	for _, model := range response.Data {
		modelType := "text"
		if strings.Contains(model.ID, "embedding") {
			modelType = "embedding"
		} else if strings.Contains(model.ID, "image") {
			modelType = "image"
		} else if strings.Contains(model.ID, "audio") {
			modelType = "audio"
		}

		modelInfos = append(modelInfos, &models.ModelInfo{
			ID:          model.ID,
			Name:        model.ID,
			Type:        modelType,
			Provider:    "openai",
			Description: fmt.Sprintf("OpenAI model: %s", model.ID),
			CreatedAt:   time.Unix(model.Created, 0),
		})
	}

	return modelInfos, nil
}

// GetModel 获取模型信息
func (s *OpenAIService) GetModel(ctx context.Context, modelID string) (*models.ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/models/"+modelID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get model request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)

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
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode get model response: %w", err)
	}

	modelType := "text"
	if strings.Contains(response.ID, "embedding") {
		modelType = "embedding"
	} else if strings.Contains(response.ID, "image") {
		modelType = "image"
	} else if strings.Contains(response.ID, "audio") {
		modelType = "audio"
	}

	return &models.ModelInfo{
		ID:          response.ID,
		Name:        response.ID,
		Type:        modelType,
		Provider:    "openai",
		Description: fmt.Sprintf("OpenAI model: %s", response.ID),
		CreatedAt:   time.Unix(response.Created, 0),
	}, nil
}

// LoadModel 加载模型
func (s *OpenAIService) LoadModel(ctx context.Context, modelID string) error {
	// OpenAI模型是按需加载的，无需预加载
	return nil
}

// UnloadModel 卸载模型
func (s *OpenAIService) UnloadModel(ctx context.Context, modelID string) error {
	// OpenAI模型是按需加载的，无需卸载
	return nil
}

// GenerateText 生成文本
func (s *OpenAIService) GenerateText(ctx context.Context, request *models.TextGenerationRequest) (*models.TextGenerationResponse, error) {
	// 转换请求格式
	openaiReq := map[string]interface{}{
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
		openaiReq["messages"] = messages
	} else if request.Prompt != "" {
		// 提示词格式
		openaiReq["prompt"] = request.Prompt
	}

	// 添加可选参数
	if request.MaxTokens > 0 {
		openaiReq["max_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		openaiReq["temperature"] = request.Temperature
	}
	if request.TopP > 0 {
		openaiReq["top_p"] = request.TopP
	}

	// 发送请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var endpoint string
	if request.Messages != nil && len(request.Messages) > 0 {
		endpoint = "/chat/completions"
	} else {
		endpoint = "/completions"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

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
	var openaiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取生成的内容
	var content string
	if request.Messages != nil && len(request.Messages) > 0 {
		// 对话格式响应
		choices, ok := openaiResp["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			return nil, fmt.Errorf("no choices in response")
		}
		
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid choice format")
		}
		
		message, ok := choice["message"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid message format")
		}
		
		content, ok = message["content"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid content format")
		}
	} else {
		// 提示词格式响应
		choices, ok := openaiResp["choices"].([]interface{})
		if !ok || len(choices) == 0 {
			return nil, fmt.Errorf("no choices in response")
		}
		
		choice, ok := choices[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid choice format")
		}
		
		content, ok = choice["text"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid text format")
		}
	}

	// 提取使用情况
	var usage models.Usage
	if usageData, ok := openaiResp["usage"].(map[string]interface{}); ok {
		usage = models.Usage{}
		if promptTokens, ok := usageData["prompt_tokens"].(float64); ok {
			usage.PromptTokens = int(promptTokens)
		}
		if completionTokens, ok := usageData["completion_tokens"].(float64); ok {
			usage.CompletionTokens = int(completionTokens)
		}
		if totalTokens, ok := usageData["total_tokens"].(float64); ok {
			usage.TotalTokens = int(totalTokens)
		}
	}

	return &models.TextGenerationResponse{
		Content: content,
		Usage:   usage,
		Model:   request.Model,
	}, nil
}

// GenerateTextStream 流式生成文本
func (s *OpenAIService) GenerateTextStream(ctx context.Context, request *models.TextGenerationRequest) (<-chan *models.TextGenerationChunk, error) {
	// 转换请求格式
	openaiReq := map[string]interface{}{
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
		openaiReq["messages"] = messages
	} else if request.Prompt != "" {
		// 提示词格式
		openaiReq["prompt"] = request.Prompt
	}

	// 添加可选参数
	if request.MaxTokens > 0 {
		openaiReq["max_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		openaiReq["temperature"] = request.Temperature
	}
	if request.TopP > 0 {
		openaiReq["top_p"] = request.TopP
	}

	// 发送请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var endpoint string
	if request.Messages != nil && len(request.Messages) > 0 {
		endpoint = "/chat/completions"
	} else {
		endpoint = "/completions"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Accept", "text/event-stream")

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
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			// 提取内容
			var content string
			var delta map[string]interface{}
			
			if request.Messages != nil && len(request.Messages) > 0 {
				// 对话格式响应
				choices, ok := chunk["choices"].([]interface{})
				if !ok || len(choices) == 0 {
					continue
				}
				
				choice, ok := choices[0].(map[string]interface{})
				if !ok {
					continue
				}
				
				delta, ok = choice["delta"].(map[string]interface{})
				if !ok {
					continue
				}
				
				if contentStr, ok := delta["content"].(string); ok {
					content = contentStr
				}
			} else {
				// 提示词格式响应
				choices, ok := chunk["choices"].([]interface{})
				if !ok || len(choices) == 0 {
					continue
				}
				
				choice, ok := choices[0].(map[string]interface{})
				if !ok {
					continue
				}
				
				if textStr, ok := choice["text"].(string); ok {
					content = textStr
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
func (s *OpenAIService) GenerateEmbedding(ctx context.Context, request *models.EmbeddingRequest) (*models.EmbeddingResponse, error) {
	// 转换请求格式
	openaiReq := map[string]interface{}{
		"model": request.Model,
		"input": request.Text,
	}

	if request.User != "" {
		openaiReq["user"] = request.User
	}

	// 发送请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

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
	var openaiResp struct {
		Object string `json:"object"`
		Data   []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openaiResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return &models.EmbeddingResponse{
		Embedding: openaiResp.Data[0].Embedding,
		Model:     openaiResp.Model,
		Usage: models.Usage{
			PromptTokens: openaiResp.Usage.PromptTokens,
			TotalTokens:  openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// GenerateEmbeddings 批量生成嵌入
func (s *OpenAIService) GenerateEmbeddings(ctx context.Context, request *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error) {
	// 转换请求格式
	openaiReq := map[string]interface{}{
		"model": request.Model,
		"input": request.Texts,
	}

	if request.User != "" {
		openaiReq["user"] = request.User
	}

	// 发送请求
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("generate embeddings failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var openaiResp struct {
		Object string `json:"object"`
		Data   []struct {
			Object    string    `json:"object"`
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openaiResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	// 按索引排序嵌入
	embeddings := make([][]float64, len(openaiResp.Data))
	for _, item := range openaiResp.Data {
		embeddings[item.Index] = item.Embedding
	}

	return &models.EmbeddingsResponse{
		Embeddings: embeddings,
		Model:      openaiResp.Model,
		Usage: models.UsageInfo{
			PromptTokens: openaiResp.Usage.PromptTokens,
			TotalTokens:  openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// CreateConversation 创建对话
func (s *OpenAIService) CreateConversation(ctx context.Context, request *models.CreateConversationRequest) (*models.Conversation, error) {
	if s.conversationDAO == nil {
		return nil, fmt.Errorf("conversation DAO not initialized")
	}

	// 将map[string]interface{}转换为datatypes.JSON
	modelConfigJSON, err := json.Marshal(request.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model config: %w", err)
	}

	// 创建对话对象
	conversation := &models.Conversation{
		ID:          "conv_" + generateID(),
		UserID:      request.UserID,
		Title:       request.Title,
		ModelConfig: modelConfigJSON,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	if err := s.conversationDAO.CreateConversation(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// 如果请求中包含初始消息，添加到对话
	if len(request.Messages) > 0 && s.messageDAO != nil {
		for _, msg := range request.Messages {
			msg.ConversationID = conversation.ID
			msg.CreatedAt = time.Now()
			msg.UpdatedAt = time.Now()
			
			if err := s.messageDAO.CreateMessage(ctx, &msg); err != nil {
				return nil, fmt.Errorf("failed to create message: %w", err)
			}
		}
	}

	return conversation, nil
}

// GetConversation 获取对话
func (s *OpenAIService) GetConversation(ctx context.Context, conversationID string) (*models.Conversation, error) {
	if s.conversationDAO == nil {
		return nil, fmt.Errorf("conversation DAO not initialized")
	}

	// 从数据库获取对话
	conversation, err := s.conversationDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// 获取对话的消息
	if s.messageDAO != nil {
		messages, err := s.messageDAO.GetMessages(ctx, conversationID, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get messages: %w", err)
		}
		conversation.Messages = make([]models.Message, len(messages))
		for i, msg := range messages {
			conversation.Messages[i] = *msg
		}
	}

	return conversation, nil
}

// ListConversations 列出对话
func (s *OpenAIService) ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error) {
	if s.conversationDAO == nil {
		return nil, fmt.Errorf("conversation DAO not initialized")
	}

	// 从数据库获取对话列表
	conversations, err := s.conversationDAO.ListConversations(ctx, userID, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	return conversations, nil
}

// UpdateConversation 更新对话
func (s *OpenAIService) UpdateConversation(ctx context.Context, conversationID string, request *models.UpdateConversationRequest) (*models.Conversation, error) {
	if s.conversationDAO == nil {
		return nil, fmt.Errorf("conversation DAO not initialized")
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if request.Title != "" {
		updates["title"] = request.Title
	}
	
	if request.ModelConfig != nil {
		modelConfigJSON, err := json.Marshal(request.ModelConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal model config: %w", err)
		}
		updates["model_config"] = modelConfigJSON
	}
	
	updates["updated_at"] = time.Now()

	// 更新对话
	if err := s.conversationDAO.UpdateConversation(ctx, conversationID, updates); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	// 获取更新后的对话
	return s.GetConversation(ctx, conversationID)
}

// DeleteConversation 删除对话
func (s *OpenAIService) DeleteConversation(ctx context.Context, conversationID string) error {
	if s.conversationDAO == nil {
		return fmt.Errorf("conversation DAO not initialized")
	}

	// 删除对话（会自动删除关联的消息）
	if err := s.conversationDAO.DeleteConversation(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

// AddMessage 添加消息
func (s *OpenAIService) AddMessage(ctx context.Context, request *models.AddMessageRequest) (*models.Message, error) {
	if s.messageDAO == nil {
		return nil, fmt.Errorf("message DAO not initialized")
	}

	// 创建消息对象
	message := &models.Message{
		ID:             "msg_" + generateID(),
		ConversationID: request.ConversationID,
		Role:           request.Role,
		Content:        request.Content,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存到数据库
	if err := s.messageDAO.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GetMessages 获取消息
func (s *OpenAIService) GetMessages(ctx context.Context, conversationID string, limit int, offset int) ([]*models.Message, error) {
	if s.messageDAO == nil {
		return nil, fmt.Errorf("message DAO not initialized")
	}

	// 从数据库获取消息列表
	messages, err := s.messageDAO.GetMessages(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// UpdateMessage 更新消息
func (s *OpenAIService) UpdateMessage(ctx context.Context, messageID string, request *models.UpdateMessageRequest) (*models.Message, error) {
	if s.messageDAO == nil {
		return nil, fmt.Errorf("message DAO not initialized")
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if request.Role != "" {
		updates["role"] = request.Role
	}
	if request.Content != "" {
		updates["content"] = request.Content
	}
	updates["updated_at"] = time.Now()

	// 更新消息
	if err := s.messageDAO.UpdateMessage(ctx, messageID, updates); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// 获取更新后的消息
	return s.messageDAO.GetMessage(ctx, messageID)
}

// DeleteMessage 删除消息
func (s *OpenAIService) DeleteMessage(ctx context.Context, messageID string) error {
	if s.messageDAO == nil {
		return fmt.Errorf("message DAO not initialized")
	}

	// 删除消息
	if err := s.messageDAO.DeleteMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// ExecuteTool 执行工具
func (s *OpenAIService) ExecuteTool(ctx context.Context, request *models.ToolExecutionRequest) (*models.ToolExecutionResponse, error) {
	// TODO: 实现工具调用功能
	return &models.ToolExecutionResponse{
		Result: "Tool execution not implemented",
	}, nil
}

// CreateFineTuningJob 创建微调作业
func (s *OpenAIService) CreateFineTuningJob(ctx context.Context, request *models.CreateFineTuningJobRequest) (*models.FineTuningJob, error) {
	// TODO: 实现微调作业创建
	return &models.FineTuningJob{
		ID:        "ft_" + generateID(),
		Model:     request.Model,
		Status:    "created",
		CreatedAt: time.Now(),
	}, nil
}

// GetFineTuningJob 获取微调作业
func (s *OpenAIService) GetFineTuningJob(ctx context.Context, jobID string) (*models.FineTuningJob, error) {
	// TODO: 实现微调作业获取
	return &models.FineTuningJob{
		ID:     jobID,
		Status: "running",
	}, nil
}

// ListFineTuningJobs 列出微调作业
func (s *OpenAIService) ListFineTuningJobs(ctx context.Context) ([]*models.FineTuningJob, error) {
	// TODO: 实现微调作业列表
	return []*models.FineTuningJob{}, nil
}

// CancelFineTuningJob 取消微调作业
func (s *OpenAIService) CancelFineTuningJob(ctx context.Context, jobID string) error {
	// TODO: 实现微调作业取消
	return nil
}

// GetServiceInfo 获取服务信息
func (s *OpenAIService) GetServiceInfo(ctx context.Context) (*models.ModelServiceInfo, error) {
	return &models.ModelServiceInfo{
		Provider:    "openai",
		Version:     "v1",
		Description: "OpenAI API Service",
		Features: []string{
			"text-generation",
			"embedding",
			"chat",
			"streaming",
		},
	}, nil
}

// generateID 生成随机ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}