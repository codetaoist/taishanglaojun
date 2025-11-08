package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/dao"
	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// ConversationImportExportService 对话导入导出服务
type ConversationImportExportService struct {
	conversationDAO *dao.ConversationDAO
	messageDAO      *dao.MessageDAO
}

// NewConversationImportExportService 创建对话导入导出服务
func NewConversationImportExportService(conversationDAO *dao.ConversationDAO, messageDAO *dao.MessageDAO) *ConversationImportExportService {
	return &ConversationImportExportService{
		conversationDAO: conversationDAO,
		messageDAO:      messageDAO,
	}
}

// ExportConversation 导出单个对话
func (s *ConversationImportExportService) ExportConversation(ctx context.Context, conversationID string) (*models.ConversationExport, error) {
	// 获取对话信息
	conversation, err := s.conversationDAO.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// 获取对话中的所有消息
	messagePtrs, err := s.messageDAO.GetMessagesByConversationID(ctx, conversationID, 0, 0) // 0表示不限制数量
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// 转换指针切片为值切片
	messages := make([]models.Message, len(messagePtrs))
	for i, msg := range messagePtrs {
		messages[i] = *msg
	}

	// 构建导出数据
	export := &models.ConversationExport{
		Conversation: *conversation,
		Messages:     messages,
		ExportedAt:   time.Now(),
	}

	return export, nil
}

// ExportConversations 导出用户的所有对话
func (s *ConversationImportExportService) ExportConversations(ctx context.Context, userID string) (*models.ConversationsExport, error) {
	// 获取用户的所有对话
	conversationPtrs, err := s.conversationDAO.GetConversationsByUserID(ctx, userID, 0, 0) // 0表示不限制数量
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// 转换指针切片为值切片
	conversations := make([]models.Conversation, len(conversationPtrs))
	for i, conv := range conversationPtrs {
		conversations[i] = *conv
	}

	// 为每个对话获取消息
	conversationExports := make([]models.ConversationExport, 0, len(conversations))
	for _, conversation := range conversations {
		messagePtrs, err := s.messageDAO.GetMessagesByConversationID(ctx, conversation.ID, 0, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get messages for conversation %s: %w", conversation.ID, err)
		}

		// 转换指针切片为值切片
		messages := make([]models.Message, len(messagePtrs))
		for i, msg := range messagePtrs {
			messages[i] = *msg
		}

		conversationExports = append(conversationExports, models.ConversationExport{
			Conversation: conversation,
			Messages:     messages,
			ExportedAt:   time.Now(),
		})
	}

	// 构建导出数据
	export := &models.ConversationsExport{
		UserID:        userID,
		Conversations: conversationExports,
		ExportedAt:    time.Now(),
	}

	return export, nil
}

// ImportConversation 导入单个对话
func (s *ConversationImportExportService) ImportConversation(ctx context.Context, export *models.ConversationExport, userID string) (*models.Conversation, error) {
	// 将map[string]interface{}转换为datatypes.JSON
	modelConfigJSON, err := json.Marshal(export.Conversation.ModelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal model config: %w", err)
	}
	
	// 创建新对话
	conversation := &models.Conversation{
		ID:          export.Conversation.ID,
		Title:       export.Conversation.Title,
		ModelConfig: modelConfigJSON,
		CreatedAt:   export.Conversation.CreatedAt,
		UpdatedAt:   export.Conversation.UpdatedAt,
	}

	// 保存对话
	if err := s.conversationDAO.CreateConversation(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// 导入消息
	for _, message := range export.Messages {
		// 创建消息的副本以避免修改原始数据
		msgCopy := message
		// 确保消息的对话ID与新创建的对话ID一致
		msgCopy.ConversationID = conversation.ID

		// 保存消息
		if err := s.messageDAO.CreateMessage(ctx, &msgCopy); err != nil {
			return nil, fmt.Errorf("failed to create message: %w", err)
		}
	}

	return conversation, nil
}

// ImportConversations 导入多个对话
func (s *ConversationImportExportService) ImportConversations(ctx context.Context, export *models.ConversationsExport, userID string) ([]*models.Conversation, error) {
	importedConversations := make([]*models.Conversation, 0, len(export.Conversations))

	for _, convExport := range export.Conversations {
		// 导入单个对话
		conversation, err := s.ImportConversation(ctx, &convExport, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to import conversation %s: %w", convExport.Conversation.ID, err)
		}

		importedConversations = append(importedConversations, conversation)
	}

	return importedConversations, nil
}

// ExportToJSON 将导出数据转换为JSON格式
func (s *ConversationImportExportService) ExportToJSON(export interface{}) ([]byte, error) {
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export to JSON: %w", err)
	}
	return jsonData, nil
}

// ImportFromJSON 从JSON格式导入数据
func (s *ConversationImportExportService) ImportFromJSON(jsonData []byte, isMultiple bool) (interface{}, error) {
	if isMultiple {
		var export models.ConversationsExport
		if err := json.Unmarshal(jsonData, &export); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON to conversations export: %w", err)
		}
		return &export, nil
	} else {
		var export models.ConversationExport
		if err := json.Unmarshal(jsonData, &export); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON to conversation export: %w", err)
		}
		return &export, nil
	}
}
