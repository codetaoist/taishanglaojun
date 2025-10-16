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

// ChatService 聊天服务
type ChatService struct {
	db              *gorm.DB
	logger          *zap.Logger
	providerManager *providers.Manager
	contextManager  *ContextManager
}

// NewChatService 创建新的聊天服务
func NewChatService(db *gorm.DB, logger *zap.Logger, providerManager *providers.Manager) *ChatService {
	return &ChatService{
		db:              db,
		logger:          logger,
		providerManager: providerManager,
		contextManager:  NewContextManager(db, logger),
	}
}

// Chat 处理聊天请求
func (s *ChatService) Chat(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 获取或创建会话
	session, err := s.getOrCreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create session: %w", err)
	}

	// 
	conversationContext, err := s.contextManager.GetOrCreateContext(ctx, session.ID, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get conversation context", zap.Error(err))
	}

	// 
	messages, err := s.getSessionMessages(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	// 
	var intentAnalysis *IntentAnalysisResult
	if conversationContext != nil {
		intentAnalysis, err = s.contextManager.AnalyzeIntent(ctx, req.Message, messages)
		if err != nil {
			s.logger.Warn("Failed to analyze intent", zap.Error(err))
		}
	}

	// 
	contextualMessage := req.Message
	if conversationContext != nil {
		contextualPrompt, err := s.contextManager.GetContextualPrompt(ctx, session.ID, req.Message)
		if err != nil {
			s.logger.Warn("Failed to get contextual prompt", zap.Error(err))
		} else {
			contextualMessage = contextualPrompt
		}
	}

	// 
	userMessage := &models.ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Message, // 
	}

	if err := s.db.Create(userMessage).Error; err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// AI
	providerReq := &providers.ChatRequest{
		Messages:    s.convertToProviderMessages(messages),
		UserID:      fmt.Sprintf("%d", req.UserID),
		SessionID:   session.ID,
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	// 滻
	if contextualMessage != req.Message && len(providerReq.Messages) > 0 {
		providerReq.Messages[len(providerReq.Messages)-1].Content = contextualMessage
	}

	// AI
	providerResp, err := s.providerManager.Chat(ctx, session.Provider, *providerReq)
	if err != nil {
		s.logger.Error("AI provider error",
			zap.Error(err),
			zap.String("provider", session.Provider),
			zap.String("model", session.Model),
		)
		return nil, fmt.Errorf("AI provider error: %w", err)
	}

	// AI
	var aiContent string
	var tokenUsed int

	if providerResp.Message.Content != "" {
		aiContent = providerResp.Message.Content
	}

	if providerResp.Usage.TotalTokens > 0 {
		tokenUsed = providerResp.Usage.TotalTokens
	}

	// AI
	aiMessage := &models.ChatMessage{
		SessionID: session.ID,
		Role:      "assistant",
		Content:   aiContent,
		TokenUsed: tokenUsed,
	}

	if err := s.db.Create(aiMessage).Error; err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
	}

	// 
	if conversationContext != nil && intentAnalysis != nil {
		err = s.contextManager.UpdateContext(ctx, session.ID, req.Message, aiContent, intentAnalysis)
		if err != nil {
			s.logger.Warn("Failed to update conversation context", zap.Error(err))
		}
	}

	// 
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

// GetSessions 
func (s *ChatService) GetSessions(ctx context.Context, userID uint, req *models.SessionListRequest) (*models.SessionListResponse, error) {
	var sessions []models.ChatSession
	var total int64

	query := s.db.Model(&models.ChatSession{}).Where("user_id = ?", userID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// 
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	// 
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

// GetMessages 
func (s *ChatService) GetMessages(ctx context.Context, userID uint, req *models.MessageListRequest) (*models.MessageListResponse, error) {
	// 
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

	// 
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	// 
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

// DeleteSession 
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

// getOrCreateSession 
func (s *ChatService) getOrCreateSession(ctx context.Context, req *models.ChatRequest) (*models.ChatSession, error) {
	if req.SessionID != nil {
		// 
		var session models.ChatSession
		if err := s.db.Where("id = ? AND user_id = ?", *req.SessionID, req.UserID).First(&session).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("session not found")
			}
			return nil, err
		}
		return &session, nil
	}

	// 
	provider := req.Provider
	if provider == "" {
		// 
		defaultProvider, err := s.providerManager.GetDefaultProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to get default provider: %w", err)
		}
		provider = defaultProvider.GetName()
	}

	model := req.Model
	if model == "" {
		// 
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

// getSessionMessages 
func (s *ChatService) getSessionMessages(ctx context.Context, sessionID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	if err := s.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// convertToProviderMessages 
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

// updateSession 
func (s *ChatService) updateSession(ctx context.Context, session *models.ChatSession, lastMessage string) error {
	updates := map[string]interface{}{
		"message_count": gorm.Expr("message_count + 2"), //  + AI
		"updated_at":    time.Now(),
	}

	// 
	if session.Title == "" || strings.HasPrefix(session.Title, "") {
		updates["title"] = s.generateSessionTitle(lastMessage)
	}

	return s.db.Model(session).Updates(updates).Error
}

// generateSessionTitle 
func (s *ChatService) generateSessionTitle(message string) string {
	// 
	title := strings.TrimSpace(message)
	if len(title) > 30 {
		title = title[:30] + "..."
	}
	if title == "" {
		title = " - " + time.Now().Format("01-02 15:04")
	}
	return title
}

// ClearSession 
func (s *ChatService) ClearSession(ctx context.Context, sessionID, userID string) error {
	// 
	var session models.ChatSession
	if err := s.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return fmt.Errorf(": %w", err)
	}

	// 
	if err := s.db.Where("session_id = ?", sessionID).Delete(&models.ChatMessage{}).Error; err != nil {
		return fmt.Errorf(": %w", err)
	}

	// 
	if err := s.db.Model(&session).Update("message_count", 0).Error; err != nil {
		s.logger.Warn("Failed to reset message count", zap.Error(err))
	}

	s.logger.Info("Session cleared", zap.String("session_id", sessionID), zap.String("user_id", userID))
	return nil
}

