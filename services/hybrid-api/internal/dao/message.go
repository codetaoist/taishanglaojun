package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"gorm.io/gorm"
)

// MessageDAO 消息数据访问对象
type MessageDAO struct {
	db *gorm.DB
}

// NewMessageDAO 创建消息DAO
func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{
		db: db,
	}
}

// CreateMessage 创建消息
func (d *MessageDAO) CreateMessage(ctx context.Context, message *models.Message) error {
	if err := d.db.WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}
	
	// 更新对话的更新时间
	if err := d.db.WithContext(ctx).Model(&models.Conversation{}).
		Where("id = ?", message.ConversationID).
		Update("updated_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to update conversation timestamp: %w", err)
	}
	
	return nil
}

// GetMessage 获取消息
func (d *MessageDAO) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	if err := d.db.WithContext(ctx).First(&message, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("message not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return &message, nil
}

// GetMessagesByConversationID 根据对话ID获取消息列表
func (d *MessageDAO) GetMessagesByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message
	query := d.db.WithContext(ctx).Where("conversation_id = ?", conversationID)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages by conversation ID: %w", err)
	}
	return messages, nil
}

// GetMessages 获取对话的消息列表
func (d *MessageDAO) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	return d.GetMessagesByConversationID(ctx, conversationID, limit, offset)
}

// UpdateMessage 更新消息
func (d *MessageDAO) UpdateMessage(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := d.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

// DeleteMessage 删除消息
func (d *MessageDAO) DeleteMessage(ctx context.Context, id string) error {
	if err := d.db.WithContext(ctx).Delete(&models.Message{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// DeleteMessagesByConversationID 根据对话ID删除所有消息
func (d *MessageDAO) DeleteMessagesByConversationID(ctx context.Context, conversationID string) error {
	if err := d.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Delete(&models.Message{}).Error; err != nil {
		return fmt.Errorf("failed to delete messages by conversation ID: %w", err)
	}
	return nil
}

// GetMessagesByIDs 根据ID列表获取消息
func (d *MessageDAO) GetMessagesByIDs(ctx context.Context, ids []string) ([]*models.Message, error) {
	var messages []*models.Message
	if err := d.db.WithContext(ctx).Where("id IN ?", ids).Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages by IDs: %w", err)
	}
	return messages, nil
}

// MessageExists 检查消息是否存在
func (d *MessageDAO) MessageExists(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := d.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check message existence: %w", err)
	}
	return count > 0, nil
}

// GetMessageCount 获取对话消息总数
func (d *MessageDAO) GetMessageCount(ctx context.Context, conversationID string) (int64, error) {
	var count int64
	if err := d.db.WithContext(ctx).Model(&models.Message{}).Where("conversation_id = ?", conversationID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}
	return count, nil
}

// GetLastMessage 获取对话的最后一条消息
func (d *MessageDAO) GetLastMessage(ctx context.Context, conversationID string) (*models.Message, error) {
	var message models.Message
	if err := d.db.WithContext(ctx).Where("conversation_id = ?", conversationID).
		Order("created_at DESC").First(&message).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no messages found for conversation: %s", conversationID)
		}
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}
	return &message, nil
}