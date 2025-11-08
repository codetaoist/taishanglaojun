package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
	"gorm.io/gorm"
)

// ModelDAO 模型数据访问对象
type ModelDAO struct {
	db *gorm.DB
}

// NewModelDAO 创建模型DAO
func NewModelDAO(db *gorm.DB) *ModelDAO {
	return &ModelDAO{
		db: db,
	}
}

// CreateModelConfig 创建模型配置
func (d *ModelDAO) CreateModelConfig(ctx context.Context, config *models.ModelConfig) error {
	if err := d.db.WithContext(ctx).Create(config).Error; err != nil {
		return fmt.Errorf("failed to create model config: %w", err)
	}
	return nil
}

// GetModelConfig 获取模型配置
func (d *ModelDAO) GetModelConfig(ctx context.Context, id uint) (*models.ModelConfig, error) {
	var config models.ModelConfig
	if err := d.db.WithContext(ctx).First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model config not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}
	return &config, nil
}

// GetModelConfigByName 根据名称获取模型配置
func (d *ModelDAO) GetModelConfigByName(ctx context.Context, name string) (*models.ModelConfig, error) {
	var config models.ModelConfig
	if err := d.db.WithContext(ctx).Where("name = ?", name).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("model config not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get model config: %w", err)
	}
	return &config, nil
}

// ListModelConfigs 列出模型配置
func (d *ModelDAO) ListModelConfigs(ctx context.Context, limit, offset int) ([]*models.ModelConfig, error) {
	var configs []*models.ModelConfig
	query := d.db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list model configs: %w", err)
	}
	return configs, nil
}

// GetEnabledModelConfigs 获取所有启用的模型配置
func (d *ModelDAO) GetEnabledModelConfigs(ctx context.Context) ([]*models.ModelConfig, error) {
	var configs []*models.ModelConfig
	if err := d.db.WithContext(ctx).Where("is_active = ?", true).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get enabled model configs: %w", err)
	}
	return configs, nil
}

// UpdateModelConfig 更新模型配置
func (d *ModelDAO) UpdateModelConfig(ctx context.Context, id uint, config *models.ModelConfig) error {
	if err := d.db.WithContext(ctx).Model(&models.ModelConfig{}).Where("id = ?", id).Updates(config).Error; err != nil {
		return fmt.Errorf("failed to update model config: %w", err)
	}
	return nil
}

// DeleteModelConfig 删除模型配置
func (d *ModelDAO) DeleteModelConfig(ctx context.Context, id uint) error {
	if err := d.db.WithContext(ctx).Delete(&models.ModelConfig{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete model config: %w", err)
	}
	return nil
}

// CreateConversation 创建对话
func (d *ModelDAO) CreateConversation(ctx context.Context, conversation *models.Conversation) error {
	if err := d.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}
	return nil
}

// GetConversation 获取对话
func (d *ModelDAO) GetConversation(ctx context.Context, id string) (*models.Conversation, error) {
	var conversation models.Conversation
	if err := d.db.WithContext(ctx).First(&conversation, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	return &conversation, nil
}

// ListConversations 列出对话
func (d *ModelDAO) ListConversations(ctx context.Context, userID string, limit, offset int) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	query := d.db.WithContext(ctx).Where("user_id = ?", userID)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("updated_at DESC").Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	return conversations, nil
}

// UpdateConversation 更新对话
func (d *ModelDAO) UpdateConversation(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := d.db.WithContext(ctx).Model(&models.Conversation{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}
	return nil
}

// DeleteConversation 删除对话
func (d *ModelDAO) DeleteConversation(ctx context.Context, id string) error {
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

// CreateMessage 创建消息
func (d *ModelDAO) CreateMessage(ctx context.Context, message *models.Message) error {
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
func (d *ModelDAO) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	if err := d.db.WithContext(ctx).First(&message, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("message not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	return &message, nil
}

// GetMessages 获取对话的消息列表
func (d *ModelDAO) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.Message, error) {
	var messages []*models.Message
	query := d.db.WithContext(ctx).Where("conversation_id = ?", conversationID)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}

// UpdateMessage 更新消息
func (d *ModelDAO) UpdateMessage(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := d.db.WithContext(ctx).Model(&models.Message{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}
	return nil
}

// DeleteMessage 删除消息
func (d *ModelDAO) DeleteMessage(ctx context.Context, id string) error {
	if err := d.db.WithContext(ctx).Delete(&models.Message{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// CreateFineTuningJob 创建微调作业
func (d *ModelDAO) CreateFineTuningJob(ctx context.Context, job *models.FineTuningJob) error {
	if err := d.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("failed to create fine-tuning job: %w", err)
	}
	return nil
}

// GetFineTuningJob 获取微调作业
func (d *ModelDAO) GetFineTuningJob(ctx context.Context, id string) (*models.FineTuningJob, error) {
	var job models.FineTuningJob
	if err := d.db.WithContext(ctx).First(&job, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("fine-tuning job not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get fine-tuning job: %w", err)
	}
	return &job, nil
}

// ListFineTuningJobs 列出微调作业
func (d *ModelDAO) ListFineTuningJobs(ctx context.Context, limit, offset int) ([]*models.FineTuningJob, error) {
	var jobs []*models.FineTuningJob
	query := d.db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("failed to list fine-tuning jobs: %w", err)
	}
	return jobs, nil
}

// UpdateFineTuningJob 更新微调作业
func (d *ModelDAO) UpdateFineTuningJob(ctx context.Context, id string, updates map[string]interface{}) error {
	if err := d.db.WithContext(ctx).Model(&models.FineTuningJob{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update fine-tuning job: %w", err)
	}
	return nil
}

// DeleteFineTuningJob 删除微调作业
func (d *ModelDAO) DeleteFineTuningJob(ctx context.Context, id string) error {
	if err := d.db.WithContext(ctx).Delete(&models.FineTuningJob{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete fine-tuning job: %w", err)
	}
	return nil
}

// GetFineTuningJobStatus 获取微调作业状态
func (d *ModelDAO) GetFineTuningJobStatus(ctx context.Context, jobID string) (string, error) {
	// 实现获取微调作业状态的逻辑
	return "running", nil
}

// GetConversationDAO 获取对话DAO
func (d *ModelDAO) GetConversationDAO() *ConversationDAO {
	return NewConversationDAO(d.db)
}

// GetMessageDAO 获取消息DAO
func (d *ModelDAO) GetMessageDAO() *MessageDAO {
	return NewMessageDAO(d.db)
}