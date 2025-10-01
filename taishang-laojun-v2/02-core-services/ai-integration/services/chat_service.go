package services

import (
	"context"
	"fmt"
	"time"

	"github.com/taishanglaojun/core-services/ai-integration/models"
	"github.com/taishanglaojun/core-services/ai-integration/providers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ChatService 对话服务
type ChatService struct {
	provider providers.AIProvider
	cache    CacheService
	repo     ConversationRepository
}

// NewChatService 创建对话服务实例
func NewChatService(provider providers.AIProvider, cache CacheService, repo ConversationRepository) *ChatService {
	return &ChatService{
		provider: provider,
		cache:    cache,
		repo:     repo,
	}
}

// Chat 处理对话请求
func (s *ChatService) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 验证请求
	if err := s.validateChatRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 获取或创建会话
	conversation, err := s.getOrCreateConversation(ctx, req.UserID, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// 构建提供者请求
	providerReq := s.buildProviderRequest(req, conversation)

	// 调用AI提供者
	providerResp, err := s.provider.Chat(ctx, providerReq)
	if err != nil {
		return nil, fmt.Errorf("provider error: %w", err)
	}

	// 保存消息到会话
	if err := s.saveMessages(ctx, conversation.ID, req.Message, providerResp.Message); err != nil {
		return nil, fmt.Errorf("failed to save messages: %w", err)
	}

	// 构建响应
	response := &ChatResponse{
		SessionID: conversation.ID,
		Message:   providerResp.Message,
		Usage:     providerResp.Usage,
		Timestamp: time.Now(),
	}

	return response, nil
}

// GetConversationHistory 获取对话历史
func (s *ChatService) GetConversationHistory(ctx context.Context, userID, sessionID string, limit int) (*models.Conversation, error) {
	return s.repo.GetConversation(ctx, userID, sessionID)
}

// ListConversations 获取用户的对话列表
func (s *ChatService) ListConversations(ctx context.Context, userID string, page, size int) ([]*models.ConversationSummary, error) {
	return s.repo.ListConversations(ctx, userID, page, size)
}

// DeleteConversation 删除对话
func (s *ChatService) DeleteConversation(ctx context.Context, userID, sessionID string) error {
	return s.repo.DeleteConversation(ctx, userID, sessionID)
}

// validateChatRequest 验证对话请求
func (s *ChatService) validateChatRequest(req *ChatRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Message.Content == "" {
		return fmt.Errorf("message content is required")
	}
	if req.Message.Role == "" {
		req.Message.Role = "user"
	}
	return nil
}

// getOrCreateConversation 获取或创建会话
func (s *ChatService) getOrCreateConversation(ctx context.Context, userID, sessionID string) (*models.Conversation, error) {
	if sessionID != "" {
		conv, err := s.repo.GetConversation(ctx, userID, sessionID)
		if err == nil {
			return conv, nil
		}
	}

	// 创建新会话
	conversation := &models.Conversation{
		ID:        generateSessionID(),
		UserID:    userID,
		Title:     "新对话",
		Messages:  []models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		Metadata: models.Metadata{
			Source:     "api",
			Tags:       []string{},
			CustomData: make(map[string]string),
		},
	}

	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

// buildProviderRequest 构建提供者请求
func (s *ChatService) buildProviderRequest(req *ChatRequest, conv *models.Conversation) providers.ChatRequest {
	messages := make([]providers.Message, 0, len(conv.Messages)+1)
	
	// 添加历史消息
	for _, msg := range conv.Messages {
		messages = append(messages, providers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	
	// 添加当前消息
	messages = append(messages, providers.Message{
		Role:    req.Message.Role,
		Content: req.Message.Content,
	})

	return providers.ChatRequest{
		Messages:    messages,
		UserID:      req.UserID,
		SessionID:   conv.ID,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
}

// saveMessages 保存消息
func (s *ChatService) saveMessages(ctx context.Context, sessionID string, userMsg, assistantMsg providers.Message) error {
	messages := []models.Message{
		{
			ID:        generateMessageID(),
			Role:      userMsg.Role,
			Content:   userMsg.Content,
			Timestamp: time.Now(),
			Metadata:  make(map[string]string),
		},
		{
			ID:        generateMessageID(),
			Role:      assistantMsg.Role,
			Content:   assistantMsg.Content,
			Timestamp: time.Now(),
			Metadata:  make(map[string]string),
		},
	}

	return s.repo.AddMessages(ctx, sessionID, messages)
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// ChatRequest 对话请求
type ChatRequest struct {
	UserID      string           `json:"user_id"`
	SessionID   string           `json:"session_id,omitempty"`
	Message     providers.Message `json:"message"`
	Temperature float32          `json:"temperature"`
	MaxTokens   int              `json:"max_tokens"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	SessionID string           `json:"session_id"`
	Message   providers.Message `json:"message"`
	Usage     providers.Usage   `json:"usage"`
	Timestamp time.Time        `json:"timestamp"`
}

// CacheService 缓存服务接口
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// ConversationRepository 对话仓储接口
type ConversationRepository interface {
	CreateConversation(ctx context.Context, conv *models.Conversation) error
	GetConversation(ctx context.Context, userID, sessionID string) (*models.Conversation, error)
	UpdateConversation(ctx context.Context, conv *models.Conversation) error
	DeleteConversation(ctx context.Context, userID, sessionID string) error
	ListConversations(ctx context.Context, userID string, page, size int) ([]*models.ConversationSummary, error)
	AddMessages(ctx context.Context, sessionID string, messages []models.Message) error
}