package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"gorm.io/gorm"
)

// ConversationDAO 对话数据访问对象
type ConversationDAO struct {
	db *gorm.DB
}

// NewConversationDAO 创建对话DAO
func NewConversationDAO(db *gorm.DB) *ConversationDAO {
	return &ConversationDAO{
		db: db,
	}
}

// CreateConversation 创建对话
func (d *ConversationDAO) CreateConversation(ctx context.Context, conversation *models.Conversation) error {
	if err := d.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

// GetConversation 获取对话
func (d *ConversationDAO) GetConversation(ctx context.Context, id string) (*models.Conversation, error) {
	var conversation models.Conversation
	if err := d.db.WithContext(ctx).First(&conversation, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	return &conversation, nil
}

// GetConversationsByUserID 根据用户ID获取对话列表
func (d *ConversationDAO) GetConversationsByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	query := d.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("updated_at DESC").Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	return conversations, nil
}

// ListConversations 列出对话
func (d *ConversationDAO) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, error) {
	return d.GetConversationsByUserID(ctx, userID, limit, offset)
}

// UpdateConversation 更新对话
func (d *ConversationDAO) UpdateConversation(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := d.db.WithContext(ctx).Model(&models.Conversation{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

// DeleteConversation 删除对话
func (d *ConversationDAO) DeleteConversation(ctx context.Context, id string) error {
	// 先删除关联的消息
	if err := d.db.WithContext(ctx).Where("conversation_id = ?", id).Delete(&models.Message{}).Error; err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	
	// 再删除对话
	if err := d.db.WithContext(ctx).Delete(&models.Conversation{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}
	return nil
}

// UpdateConversationTimestamp 更新对话时间戳
func (d *ConversationDAO) UpdateConversationTimestamp(ctx context.Context, id string) error {
	if err := d.db.WithContext(ctx).Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("updated_at", time.Now()).Error; err != nil {
		return fmt.Errorf("failed to update conversation timestamp: %w", err)
	}
	return nil
}

// GetConversationsByIDs 根据ID列表获取对话
func (d *ConversationDAO) GetConversationsByIDs(ctx context.Context, ids []string) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	if err := d.db.WithContext(ctx).Where("id IN ?", ids).Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("failed to get conversations by IDs: %w", err)
	}
	return conversations, nil
}

// ConversationExists 检查对话是否存在
func (d *ConversationDAO) ConversationExists(ctx context.Context, id string) (bool, error) {
	var count int64
	if err := d.db.WithContext(ctx).Model(&models.Conversation{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check conversation existence: %w", err)
	}
	return count > 0, nil
}

// GetConversationCount 获取对话总数
func (d *ConversationDAO) GetConversationCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := d.db.WithContext(ctx).Model(&models.Conversation{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get conversation count: %w", err)
	}
	return count, nil
}