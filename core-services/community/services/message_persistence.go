package services

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MessagePersistenceService 消息持久化服务
type MessagePersistenceService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMessagePersistenceService 创建消息持久化服务
func NewMessagePersistenceService(db *gorm.DB, logger *zap.Logger) *MessagePersistenceService {
	return &MessagePersistenceService{
		db:     db,
		logger: logger,
	}
}

// SaveMessage 保存消息到数据库
func (s *MessagePersistenceService) SaveMessage(ctx context.Context, message *models.ChatMessage) error {
	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		s.logger.Error("Failed to save message", 
			zap.Error(err),
			zap.Uint("room_id", message.RoomID),
			zap.Uint("sender_id", message.SenderID))
		return fmt.Errorf("failed to save message: %w", err)
	}

	s.logger.Debug("Message saved successfully", 
		zap.Uint("message_id", message.ID),
		zap.Uint("room_id", message.RoomID),
		zap.Uint("sender_id", message.SenderID))
	
	return nil
}

// GetMessagesByRoom 获取房间的消息历史
func (s *MessagePersistenceService) GetMessagesByRoom(ctx context.Context, roomID uint, limit, offset int) ([]*models.ChatMessage, error) {
	var messages []*models.ChatMessage
	
	query := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&messages).Error; err != nil {
		s.logger.Error("Failed to get messages by room", 
			zap.Error(err),
			zap.Uint("room_id", roomID))
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	
	return messages, nil
}

// GetMessagesByUser 获取用户发送的消息
func (s *MessagePersistenceService) GetMessagesByUser(ctx context.Context, userID uint, limit, offset int) ([]*models.ChatMessage, error) {
	var messages []*models.ChatMessage
	
	query := s.db.WithContext(ctx).
		Where("sender_id = ?", userID).
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&messages).Error; err != nil {
		s.logger.Error("Failed to get messages by user", 
			zap.Error(err),
			zap.Uint("user_id", userID))
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	
	return messages, nil
}

// GetMessageByID 根据ID获取消息
func (s *MessagePersistenceService) GetMessageByID(ctx context.Context, messageID uint) (*models.ChatMessage, error) {
	var message models.ChatMessage
	
	if err := s.db.WithContext(ctx).First(&message, messageID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("message not found")
		}
		s.logger.Error("Failed to get message by ID", 
			zap.Error(err),
			zap.Uint("message_id", messageID))
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	
	return &message, nil
}

// UpdateMessage 更新消息
func (s *MessagePersistenceService) UpdateMessage(ctx context.Context, messageID uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	
	if err := s.db.WithContext(ctx).Model(&models.ChatMessage{}).Where("id = ?", messageID).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update message", 
			zap.Error(err),
			zap.Uint("message_id", messageID))
		return fmt.Errorf("failed to update message: %w", err)
	}
	
	s.logger.Debug("Message updated successfully", 
		zap.Uint("message_id", messageID))
	
	return nil
}

// DeleteMessage 删除消息（软删除）
func (s *MessagePersistenceService) DeleteMessage(ctx context.Context, messageID uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.ChatMessage{}, messageID).Error; err != nil {
		s.logger.Error("Failed to delete message", 
			zap.Error(err),
			zap.Uint("message_id", messageID))
		return fmt.Errorf("failed to delete message: %w", err)
	}
	
	s.logger.Debug("Message deleted successfully", 
		zap.Uint("message_id", messageID))
	
	return nil
}

// GetMessageCount 获取房间消息总数
func (s *MessagePersistenceService) GetMessageCount(ctx context.Context, roomID uint) (int64, error) {
	var count int64
	
	if err := s.db.WithContext(ctx).Model(&models.ChatMessage{}).Where("room_id = ?", roomID).Count(&count).Error; err != nil {
		s.logger.Error("Failed to count messages", 
			zap.Error(err),
			zap.Uint("room_id", roomID))
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}
	
	return count, nil
}

// SearchMessages 搜索消息
func (s *MessagePersistenceService) SearchMessages(ctx context.Context, roomID uint, keyword string, limit, offset int) ([]*models.ChatMessage, error) {
	var messages []*models.ChatMessage
	
	query := s.db.WithContext(ctx).
		Where("room_id = ? AND content LIKE ?", roomID, "%"+keyword+"%").
		Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&messages).Error; err != nil {
		s.logger.Error("Failed to search messages", 
			zap.Error(err),
			zap.Uint("room_id", roomID),
			zap.String("keyword", keyword))
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}
	
	return messages, nil
}

// MarkMessageAsRead 标记消息为已读
func (s *MessagePersistenceService) MarkMessageAsRead(ctx context.Context, messageID, userID uint) error {
	readRecord := &models.ChatMessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}
	
	// 使用 ON CONFLICT 或类似机制避免重复插入
	if err := s.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		FirstOrCreate(readRecord).Error; err != nil {
		s.logger.Error("Failed to mark message as read", 
			zap.Error(err),
			zap.Uint("message_id", messageID),
			zap.Uint("user_id", userID))
		return fmt.Errorf("failed to mark message as read: %w", err)
	}
	
	return nil
}

// GetUnreadMessageCount 获取用户未读消息数量
func (s *MessagePersistenceService) GetUnreadMessageCount(ctx context.Context, userID, roomID uint) (int64, error) {
	var count int64
	
	// 查询房间中用户未读的消息数量
	subQuery := s.db.WithContext(ctx).
		Select("message_id").
		Table("chat_message_reads").
		Where("user_id = ?", userID)
	
	if err := s.db.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("room_id = ? AND sender_id != ? AND id NOT IN (?)", roomID, userID, subQuery).
		Count(&count).Error; err != nil {
		s.logger.Error("Failed to count unread messages", 
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.Uint("room_id", roomID))
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}
	
	return count, nil
}