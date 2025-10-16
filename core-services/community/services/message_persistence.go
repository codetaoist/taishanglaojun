package services

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/community/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MessagePersistenceService 
type MessagePersistenceService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMessagePersistenceService 
func NewMessagePersistenceService(db *gorm.DB, logger *zap.Logger) *MessagePersistenceService {
	return &MessagePersistenceService{
		db:     db,
		logger: logger,
	}
}

// SaveMessage 
func (s *MessagePersistenceService) SaveMessage(ctx context.Context, message *models.ChatMessage) error {
	if err := s.db.WithContext(ctx).Create(message).Error; err != nil {
		s.logger.Error("Failed to save message",
			zap.Error(err),
			zap.Uint("room_id", message.RoomID),
			zap.String("sender_id", message.SenderID.String()))
		return fmt.Errorf("failed to save message: %w", err)
	}

	s.logger.Debug("Message saved successfully",
		zap.Uint("message_id", message.ID),
		zap.Uint("room_id", message.RoomID),
		zap.String("sender_id", message.SenderID.String()))

	return nil
}

// GetMessagesByRoom 
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

// GetMessagesByUser 
func (s *MessagePersistenceService) GetMessagesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.ChatMessage, error) {
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
			zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// GetMessageByID ID
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

// UpdateMessage 
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

// DeleteMessage 
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

// GetMessageCount 
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

// SearchMessages 
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

// MarkMessageAsRead 
func (s *MessagePersistenceService) MarkMessageAsRead(ctx context.Context, messageID uint, userID uuid.UUID) error {
	readRecord := &models.ChatMessageRead{
		MessageID: messageID,
		UserID:    userID,
		ReadAt:    time.Now(),
	}

	//  ON CONFLICT 
	if err := s.db.WithContext(ctx).
		Where("message_id = ? AND user_id = ?", messageID, userID).
		FirstOrCreate(readRecord).Error; err != nil {
		s.logger.Error("Failed to mark message as read",
			zap.Error(err),
			zap.Uint("message_id", messageID),
			zap.String("user_id", userID.String()))
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// GetUnreadMessageCount 
func (s *MessagePersistenceService) GetUnreadMessageCount(ctx context.Context, userID uuid.UUID, roomID uint) (int64, error) {
	var count int64

	// 
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
			zap.String("user_id", userID.String()),
			zap.Uint("room_id", roomID))
		return 0, fmt.Errorf("failed to count unread messages: %w", err)
	}

	return count, nil
}

