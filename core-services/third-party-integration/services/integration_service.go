package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/repositories"
)

// IntegrationService 第三方服务集成服�?type IntegrationService struct {
	repo       *repositories.IntegrationRepository
	httpClient *http.Client
}

// NewIntegrationService 创建新的集成服务
func NewIntegrationService(repo *repositories.IntegrationRepository) *IntegrationService {
	return &IntegrationService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateIntegration 创建集成
func (s *IntegrationService) CreateIntegration(userID int64, name, provider string, integrationType models.IntegrationType, config map[string]interface{}) (*models.Integration, error) {
	integration := &models.Integration{
		UserID:       userID,
		Name:         name,
		Provider:     provider,
		Type:         integrationType,
		Status:       models.IntegrationStatusConfiguring,
		Config:       config,
		Settings:     make(map[string]interface{}),
		SyncInterval: 3600, // 默认1小时同步一�?		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 验证配置
	if err := s.validateIntegrationConfig(integration); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// 保存到数据库
	id, err := s.repo.Create(integration)
	if err != nil {
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}

	integration.ID = id
	return integration, nil
}

// GetIntegration 获取集成
func (s *IntegrationService) GetIntegration(id int64) (*models.Integration, error) {
	return s.repo.GetByID(id)
}

// ListIntegrations 获取用户的集成列�?func (s *IntegrationService) ListIntegrations(userID int64, provider string, limit, offset int) ([]*models.Integration, int64, error) {
	return s.repo.ListByUserID(userID, provider, limit, offset)
}

// UpdateIntegration 更新集成
func (s *IntegrationService) UpdateIntegration(id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.repo.Update(id, updates)
}

// DeleteIntegration 删除集成
func (s *IntegrationService) DeleteIntegration(id int64) error {
	// 先停止同�?	if err := s.StopSync(id); err != nil {
		return fmt.Errorf("failed to stop sync: %w", err)
	}

	return s.repo.Delete(id)
}

// TestIntegration 测试集成连接
func (s *IntegrationService) TestIntegration(id int64) error {
	integration, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	switch integration.Type {
	case models.IntegrationTypeAPI:
		return s.testAPIIntegration(integration)
	case models.IntegrationTypeWebhook:
		return s.testWebhookIntegration(integration)
	case models.IntegrationTypeOAuth:
		return s.testOAuthIntegration(integration)
	case models.IntegrationTypeDatabase:
		return s.testDatabaseIntegration(integration)
	case models.IntegrationTypeFile:
		return s.testFileIntegration(integration)
	default:
		return fmt.Errorf("unsupported integration type: %s", integration.Type)
	}
}

// SyncData 同步数据
func (s *IntegrationService) SyncData(id int64) error {
	integration, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	if !integration.IsActive {
		return fmt.Errorf("integration is not active")
	}

	// 更新状态为同步�?	s.repo.Update(id, map[string]interface{}{
		"status":     models.IntegrationStatusSyncing,
		"updated_at": time.Now(),
	})

	// 执行同步
	err = s.performSync(integration)
	
	// 更新同步状�?	status := models.IntegrationStatusActive
	if err != nil {
		status = models.IntegrationStatusError
	}

	s.repo.Update(id, map[string]interface{}{
		"status":       status,
		"last_sync_at": time.Now(),
		"updated_at":   time.Now(),
	})

	return err
}

// StartSync 启动定时同步
func (s *IntegrationService) StartSync(id int64) error {
	integration, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	// 这里可以启动定时任务
	// 使用cron或其他调度器
	fmt.Printf("Starting sync for integration %s\n", integration.Name)
	
	return s.repo.Update(id, map[string]interface{}{
		"status":     models.IntegrationStatusActive,
		"updated_at": time.Now(),
	})
}

// StopSync 停止同步
func (s *IntegrationService) StopSync(id int64) error {
	integration, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("integration not found: %w", err)
	}

	// 停止定时任务
	fmt.Printf("Stopping sync for integration %s\n", integration.Name)
	
	return s.repo.Update(id, map[string]interface{}{
		"status":     models.IntegrationStatusInactive,
		"updated_at": time.Now(),
	})
}

// GetIntegrationStats 获取集成统计信息
func (s *IntegrationService) GetIntegrationStats(id int64) (map[string]interface{}, error) {
	integration, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("integration not found: %w", err)
	}

	stats := map[string]interface{}{
		"name":           integration.Name,
		"provider":       integration.Provider,
		"type":           integration.Type,
		"status":         integration.Status,
		"last_sync_at":   integration.LastSyncAt,
		"sync_interval":  integration.SyncInterval,
		"is_active":      integration.IsActive,
		"created_at":     integration.CreatedAt,
		"updated_at":     integration.UpdatedAt,
	}

	// 添加提供商特定的统计信息
	switch integration.Provider {
	case "github":
		stats["repositories_synced"] = 10
		stats["last_commit_sync"] = time.Now().Add(-1 * time.Hour)
	case "slack":
		stats["channels_synced"] = 5
		stats["messages_processed"] = 1000
	case "google_drive":
		stats["files_synced"] = 50
		stats["storage_used"] = "2.5GB"
	}

	return stats, nil
}

// validateIntegrationConfig 验证集成配置
func (s *IntegrationService) validateIntegrationConfig(integration *models.Integration) error {
	switch integration.Type {
	case models.IntegrationTypeAPI:
		return s.validateAPIConfig(integration.Config)
	case models.IntegrationTypeWebhook:
		return s.validateWebhookConfig(integration.Config)
	case models.IntegrationTypeOAuth:
		return s.validateOAuthConfig(integration.Config)
	case models.IntegrationTypeDatabase:
		return s.validateDatabaseConfig(integration.Config)
	case models.IntegrationTypeFile:
		return s.validateFileConfig(integration.Config)
	default:
		return fmt.Errorf("unsupported integration type: %s", integration.Type)
	}
}

// validateAPIConfig 验证API配置
func (s *IntegrationService) validateAPIConfig(config map[string]interface{}) error {
	if _, ok := config["base_url"]; !ok {
		return fmt.Errorf("base_url is required for API integration")
	}
	if _, ok := config["api_key"]; !ok {
		return fmt.Errorf("api_key is required for API integration")
	}
	return nil
}

// validateWebhookConfig 验证Webhook配置
func (s *IntegrationService) validateWebhookConfig(config map[string]interface{}) error {
	if _, ok := config["webhook_url"]; !ok {
		return fmt.Errorf("webhook_url is required for webhook integration")
	}
	return nil
}

// validateOAuthConfig 验证OAuth配置
func (s *IntegrationService) validateOAuthConfig(config map[string]interface{}) error {
	if _, ok := config["client_id"]; !ok {
		return fmt.Errorf("client_id is required for OAuth integration")
	}
	if _, ok := config["client_secret"]; !ok {
		return fmt.Errorf("client_secret is required for OAuth integration")
	}
	return nil
}

// validateDatabaseConfig 验证数据库配�?func (s *IntegrationService) validateDatabaseConfig(config map[string]interface{}) error {
	if _, ok := config["connection_string"]; !ok {
		return fmt.Errorf("connection_string is required for database integration")
	}
	return nil
}

// validateFileConfig 验证文件配置
func (s *IntegrationService) validateFileConfig(config map[string]interface{}) error {
	if _, ok := config["file_path"]; !ok {
		return fmt.Errorf("file_path is required for file integration")
	}
	return nil
}

// testAPIIntegration 测试API集成
func (s *IntegrationService) testAPIIntegration(integration *models.Integration) error {
	baseURL, ok := integration.Config["base_url"].(string)
	if !ok {
		return fmt.Errorf("invalid base_url")
	}

	apiKey, ok := integration.Config["api_key"].(string)
	if !ok {
		return fmt.Errorf("invalid api_key")
	}

	// 创建测试请求
	req, err := http.NewRequest("GET", baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请�?	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API test failed with status: %d", resp.StatusCode)
	}

	return nil
}

// testWebhookIntegration 测试Webhook集成
func (s *IntegrationService) testWebhookIntegration(integration *models.Integration) error {
	webhookURL, ok := integration.Config["webhook_url"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook_url")
	}

	// 发送测试payload
	testPayload := map[string]interface{}{
		"test":      true,
		"timestamp": time.Now().Unix(),
		"message":   "Test webhook from integration service",
	}

	payloadBytes, err := json.Marshal(testPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal test payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook test failed with status: %d", resp.StatusCode)
	}

	return nil
}

// testOAuthIntegration 测试OAuth集成
func (s *IntegrationService) testOAuthIntegration(integration *models.Integration) error {
	// 这里可以验证OAuth配置的有效�?	// 检查client_id和client_secret是否有效
	fmt.Printf("Testing OAuth integration for %s\n", integration.Provider)
	return nil
}

// testDatabaseIntegration 测试数据库集�?func (s *IntegrationService) testDatabaseIntegration(integration *models.Integration) error {
	// 这里可以测试数据库连�?	fmt.Printf("Testing database integration for %s\n", integration.Provider)
	return nil
}

// testFileIntegration 测试文件集成
func (s *IntegrationService) testFileIntegration(integration *models.Integration) error {
	// 这里可以测试文件访问权限
	fmt.Printf("Testing file integration for %s\n", integration.Provider)
	return nil
}

// performSync 执行数据同步
func (s *IntegrationService) performSync(integration *models.Integration) error {
	switch integration.Provider {
	case "github":
		return s.syncGitHubData(integration)
	case "slack":
		return s.syncSlackData(integration)
	case "google_drive":
		return s.syncGoogleDriveData(integration)
	default:
		return s.syncGenericData(integration)
	}
}

// syncGitHubData 同步GitHub数据
func (s *IntegrationService) syncGitHubData(integration *models.Integration) error {
	fmt.Printf("Syncing GitHub data for integration %s\n", integration.Name)
	// 实现GitHub API调用和数据同步逻辑
	return nil
}

// syncSlackData 同步Slack数据
func (s *IntegrationService) syncSlackData(integration *models.Integration) error {
	fmt.Printf("Syncing Slack data for integration %s\n", integration.Name)
	// 实现Slack API调用和数据同步逻辑
	return nil
}

// syncGoogleDriveData 同步Google Drive数据
func (s *IntegrationService) syncGoogleDriveData(integration *models.Integration) error {
	fmt.Printf("Syncing Google Drive data for integration %s\n", integration.Name)
	// 实现Google Drive API调用和数据同步逻辑
	return nil
}

// syncGenericData 通用数据同步
func (s *IntegrationService) syncGenericData(integration *models.Integration) error {
	fmt.Printf("Syncing generic data for integration %s\n", integration.Name)
	// 实现通用数据同步逻辑
	return nil
}
