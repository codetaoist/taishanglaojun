package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ChatService 对话服务
type ChatService struct {
	db              *gorm.DB
	logger          *zap.Logger
	providerManager *providers.Manager
	contextManager  *ContextManager
}

// NewChatService 创建对话服务
func NewChatService(db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) *ChatService {
	return &ChatService{
		db:              db,
		logger:          logger,
		providerManager: providerManager,
		contextManager:  NewContextManager(db, logger),
	}
}

// Chat 发送对话消�?
func (s *ChatService) Chat(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 获取或创建会�?
	session, err := s.getOrCreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create session: %w", err)
	}

	// 获取或创建对话上下文
	conversationContext, err := s.contextManager.GetOrCreateContext(ctx, session.ID, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get conversation context", zap.Error(err))
	}

	// 获取对话历史用于意图分析
	messages, err := s.getSessionMessages(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	// 分析用户意图
	var intentAnalysis *IntentAnalysisResult
	if conversationContext != nil {
		intentAnalysis, err = s.contextManager.AnalyzeIntent(ctx, req.Message, messages)
		if err != nil {
			s.logger.Warn("Failed to analyze intent", zap.Error(err))
		}
	}

	// 获取上下文化的提示词
	contextualMessage := req.Message
	if conversationContext != nil {
		contextualPrompt, err := s.contextManager.GetContextualPrompt(ctx, session.ID, req.Message)
		if err != nil {
			s.logger.Warn("Failed to get contextual prompt", zap.Error(err))
		} else {
			contextualMessage = contextualPrompt
		}
	}

	// 保存用户消息
	userMessage := &models.ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Message, // 保存原始消息
	}

	if err := s.db.Create(userMessage).Error; err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 构建AI请求（使用上下文化的消息�?
	providerReq := &providers.ChatRequest{
		Messages:    s.convertToProviderMessages(messages),
		UserID:      fmt.Sprintf("%d", req.UserID),
		SessionID:   session.ID,
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// 如果有上下文化的消息，替换最后一条用户消�?
	if contextualMessage != req.Message && len(providerReq.Messages) > 0 {
		providerReq.Messages[len(providerReq.Messages)-1].Content = contextualMessage
	}

	// 调用AI提供�?
	providerResp, err := s.providerManager.Chat(ctx, session.Provider, *providerReq)
	if err != nil {
		s.logger.Error("AI provider error",
			zap.Error(err),
			zap.String("provider", session.Provider),
			zap.String("model", session.Model),
		)
		return nil, fmt.Errorf("AI provider error: %w", err)
	}

	// 提取AI回复
	var aiContent string
	var tokenUsed int

	if providerResp.Message.Content != "" {
		aiContent = providerResp.Message.Content
	}

	if providerResp.Usage.TotalTokens > 0 {
		tokenUsed = providerResp.Usage.TotalTokens
	}

	// 保存AI回复
	aiMessage := &models.ChatMessage{
		SessionID: session.ID,
		Role:      "assistant",
		Content:   aiContent,
		TokenUsed: tokenUsed,
	}

	if err := s.db.Create(aiMessage).Error; err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
	}

	// 更新对话上下�?
	if conversationContext != nil && intentAnalysis != nil {
		err = s.contextManager.UpdateContext(ctx, session.ID, req.Message, aiContent, intentAnalysis)
		if err != nil {
			s.logger.Warn("Failed to update conversation context", zap.Error(err))
		}
	}

	// 更新会话信息
	if err := s.updateSession(ctx, session, aiContent); err != nil {
		s.logger.Warn("Failed to update session", zap.Error(err))
	}

	return &models.ChatResponse{
		SessionID: session.ID,
		MessageID: aiMessage.ID,
		Content:   aiContent,
		TokenUsed: tokenUsed,
		Provider:  session.Provider,
		Model:     session.Model,
	}, nil
}

// GetSessions 获取用户会话列表
func (s *ChatService) GetSessions(ctx context.Context, userID uint, req *models.SessionListRequest) (*models.SessionListResponse, error) {
	var sessions []models.ChatSession
	var total int64

	query := s.db.Model(&models.ChatSession{}).Where("user_id = ?", userID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("updated_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return &models.SessionListResponse{
		Sessions: sessions,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetMessages 获取会话消息列表
func (s *ChatService) GetMessages(ctx context.Context, userID uint, req *models.MessageListRequest) (*models.MessageListResponse, error) {
	// 验证会话所有权
	var session models.ChatSession
	if err := s.db.Where("id = ? AND user_id = ?", req.SessionID, userID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var messages []models.ChatMessage
	var total int64

	query := s.db.Model(&models.ChatMessage{}).Where("session_id = ?", req.SessionID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Order("created_at ASC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return &models.MessageListResponse{
		Messages: messages,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeleteSession 删除会话
func (s *ChatService) DeleteSession(ctx context.Context, userID, sessionID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", sessionID, userID).Delete(&models.ChatSession{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// getOrCreateSession 获取或创建会�?
func (s *ChatService) getOrCreateSession(ctx context.Context, req *models.ChatRequest) (*models.ChatSession, error) {
	if req.SessionID != nil {
		// 获取现有会话
		var session models.ChatSession
		if err := s.db.Where("id = ? AND user_id = ?", *req.SessionID, req.UserID).First(&session).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("session not found")
			}
			return nil, err
		}
		return &session, nil
	}

	// 创建新会�?
	provider := req.Provider
	if provider == "" {
		// 使用默认提供�?
		defaultProvider, err := s.providerManager.GetDefaultProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to get default provider: %w", err)
		}
		provider = defaultProvider.GetName()
	}

	model := req.Model
	if model == "" {
		// 使用提供商的默认模型
		providerInstance, err := s.providerManager.GetProvider(provider)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider: %w", err)
		}
		models := providerInstance.GetModels()
		if len(models) > 0 {
			model = models[0]
		}
	}

	session := &models.ChatSession{
		UserID:   req.UserID,
		Title:    s.generateSessionTitle(req.Message),
		Provider: provider,
		Model:    model,
		Status:   "active",
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// getSessionMessages 获取会话消息
func (s *ChatService) getSessionMessages(ctx context.Context, sessionID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	if err := s.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// convertToProviderMessages 转换为提供商消息格式
func (s *ChatService) convertToProviderMessages(messages []models.ChatMessage) []providers.Message {
	providerMessages := make([]providers.Message, len(messages))
	for i, msg := range messages {
		providerMessages[i] = providers.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return providerMessages
}

// updateSession 更新会话信息
func (s *ChatService) updateSession(ctx context.Context, session *models.ChatSession, lastMessage string) error {
	updates := map[string]interface{}{
		"message_count": gorm.Expr("message_count + 2"), // 用户消息 + AI回复
		"updated_at":    time.Now(),
	}

	// 如果会话标题为空或是默认标题，尝试生成新标题
	if session.Title == "" || strings.HasPrefix(session.Title, "新对�?) {
		updates["title"] = s.generateSessionTitle(lastMessage)
	}

	return s.db.Model(session).Updates(updates).Error
}

// generateSessionTitle 生成会话标题
func (s *ChatService) generateSessionTitle(message string) string {
	// 简单的标题生成逻辑
	title := strings.TrimSpace(message)
	if len(title) > 30 {
		title = title[:30] + "..."
	}
	if title == "" {
		title = "新对�?- " + time.Now().Format("01-02 15:04")
	}
	return title
}


// ClearSession 清空会话消息
func (s *ChatService) ClearSession(ctx context.Context, sessionID, userID string) error {
	// 验证会话是否存在且属于该用户
	var session models.ChatSession
	if err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return fmt.Errorf("会话不存在或无权�? %w", err)
	}

	// 删除该会话的所有消�?
	if err := s.db.Where("session_id = ?", sessionID).Delete(&models.ChatMessage{}).Error; err != nil {
		return fmt.Errorf("清空会话消息失败: %w", err)
	}

	// 重置会话的消息计�?
	if err := s.db.Model(&session).Update("message_count", 0).Error; err != nil {
		s.logger.Warn("Failed to reset message count", zap.Error(err))
	}

	s.logger.Info("Session cleared", zap.String("session_id", sessionID), zap.String("user_id", userID))
	return nil
}

