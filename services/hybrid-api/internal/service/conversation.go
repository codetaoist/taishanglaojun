package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// ConversationService 对话管理服务
type ConversationService struct {
	modelDAO *dao.ModelDAO
}

// NewConversationService 创建对话管理服务
func NewConversationService(modelDAO *dao.ModelDAO) *ConversationService {
	return &ConversationService{
		modelDAO: modelDAO,
	}
}

// CreateConversation 创建对话
func (s *ConversationService) CreateConversation(ctx context.Context, request *models.CreateConversationRequest) (*models.Conversation, error) {
	// 将map[string]interface{}转换为datatypes.JSON
	modelConfigJSON, err := json.Marshal(request.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model config: %w", err)
	}
	
	// 创建对话对象
	conversation := &models.Conversation{
		ID:          generateConversationID(),
		UserID:      request.UserID,
		Title:       request.Title,
		ModelConfig: modelConfigJSON,
		Messages:    request.Messages,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	if err := s.modelDAO.CreateConversation(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// GetConversation 获取对话详情
func (s *ConversationService) GetConversation(ctx context.Context, conversationID string) (*models.Conversation, error) {
	// 从数据库获取对话
	conversation, err := s.modelDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversation, nil
}

// ListConversations 列出用户的对话
func (s *ConversationService) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, error) {
	// 从数据库获取对话列表
	conversations, err := s.modelDAO.ListConversations(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	return conversations, nil
}

// UpdateConversation 更新对话
func (s *ConversationService) UpdateConversation(ctx context.Context, conversationID string, request *models.UpdateConversationRequest) (*models.Conversation, error) {
	// 验证对话是否存在
	_, err := s.modelDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if request.Title != "" {
		updates["title"] = request.Title
	}
	if request.Model != "" {
		updates["model"] = request.Model
	}
	if request.ModelConfig != nil {
		// 将map[string]interface{}转换为datatypes.JSON
		modelConfigJSON, err := json.Marshal(request.ModelConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal model config: %w", err)
		}
		updates["model_config"] = modelConfigJSON
	}
	updates["updated_at"] = time.Now()

	// 更新对话
	if err := s.modelDAO.UpdateConversation(ctx, conversationID, updates); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	// 获取更新后的对话
	return s.modelDAO.GetConversation(ctx, conversationID)
}

// DeleteConversation 删除对话
func (s *ConversationService) DeleteConversation(ctx context.Context, conversationID string) error {
	// 验证对话是否存在
	_, err := s.modelDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("conversation not found: %w", err)
	}

	// 删除对话（包括关联的消息）
	if err := s.modelDAO.DeleteConversation(ctx, conversationID); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

// AddMessage 添加消息到对话
func (s *ConversationService) AddMessage(ctx context.Context, request *models.AddMessageRequest) (*models.Message, error) {
	// 验证对话是否存在
	conversation, err := s.modelDAO.GetConversation(ctx, request.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// 创建消息对象
	metadataJSON, _ := json.Marshal(map[string]interface{}{})
	message := &models.Message{
		ID:             generateMessageID(),
		ConversationID: request.ConversationID,
		UserID:         conversation.UserID, // 使用对话的UserID
		Role:           request.Role,
		Content:        request.Content,
		Metadata:       metadataJSON, // 初始化Metadata字段为空JSON
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存到数据库
	if err := s.modelDAO.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GetMessages 获取对话的消息列表
func (s *ConversationService) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	// 验证对话是否存在
	_, err := s.modelDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// 获取消息列表
	messages, err := s.modelDAO.GetMessages(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// UpdateMessage 更新消息
func (s *ConversationService) UpdateMessage(ctx context.Context, messageID string, request *models.UpdateMessageRequest) (*models.Message, error) {
	// 验证消息是否存在
	_, err := s.modelDAO.GetMessage(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if request.Content != "" {
		updates["content"] = request.Content
	}
	if request.Role != "" {
		updates["role"] = request.Role
	}
	updates["updated_at"] = time.Now()

	// 更新消息
	if err := s.modelDAO.UpdateMessage(ctx, messageID, updates); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// 获取更新后的消息
	return s.modelDAO.GetMessage(ctx, messageID)
}

// DeleteMessage 删除消息
func (s *ConversationService) DeleteMessage(ctx context.Context, messageID string) error {
	// 验证消息是否存在
	_, err := s.modelDAO.GetMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("message not found: %w", err)
	}

	// 删除消息
	if err := s.modelDAO.DeleteMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// SearchConversations 搜索对话（基于标题或消息内容）
func (s *ConversationService) SearchConversations(ctx context.Context, userID, query string, limit, offset int) ([]*models.Conversation, error) {
	// TODO: 实现基于向量数据库的对话搜索功能
	// 这里先实现简单的文本搜索
	
	// 获取用户的所有对话
	conversations, err := s.modelDAO.ListConversations(ctx, userID, 1000, 0) // 获取更多结果用于搜索
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations for search: %w", err)
	}

	// 简单的文本匹配过滤
	var filteredConversations []*models.Conversation
	for _, conv := range conversations {
		// 检查标题是否匹配
		titleMatch := containsIgnoreCase(conv.Title, query)
		
		// 检查消息是否匹配
		messageMatch := false
		for _, msg := range conv.Messages {
			if containsIgnoreCase(msg.Content, query) {
				messageMatch = true
				break
			}
		}
		
		if titleMatch || messageMatch {
			filteredConversations = append(filteredConversations, conv)
		}
	}

	// 应用分页
	total := len(filteredConversations)
	start := offset
	if start > total {
		start = total
	}
	
	end := start + limit
	if end > total {
		end = total
	}
	
	if start >= end {
		return []*models.Conversation{}, nil
	}
	
	return filteredConversations[start:end], nil
}

// containsIgnoreCase 检查字符串是否包含子字符串（不区分大小写）
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

// toLower 转换为小写
func toLower(s string) string {
	result := make([]rune, len([]rune(s)))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	sLen := len(s)
	subLen := len(substr)
	
	if subLen > sLen {
		return false
	}
	
	if subLen == 0 {
		return true
	}
	
	for i := 0; i <= sLen-subLen; i++ {
		if s[i:i+subLen] == substr {
			return true
		}
	}
	
	return false
}

// generateConversationID 生成对话ID
func generateConversationID() string {
	return fmt.Sprintf("conv_%d", time.Now().UnixNano())
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}