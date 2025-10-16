package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/repositories"
)

// APIKeyService API密钥服务
type APIKeyService struct {
	repo *repositories.APIKeyRepository
}

// NewAPIKeyService 创建新的API密钥服务
func NewAPIKeyService(repo *repositories.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		repo: repo,
	}
}

// CreateAPIKey 创建API密钥
func (s *APIKeyService) CreateAPIKey(userID int64, name string, permissions []string, rateLimit int, expiresAt *time.Time) (*models.APIKey, error) {
	// 生成API密钥
	key, secret, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// 创建密钥哈希
	secretHash := s.hashSecret(secret)

	apiKey := &models.APIKey{
		UserID:      userID,
		Name:        name,
		Key:         key,
		SecretHash:  secretHash,
		Permissions: permissions,
		RateLimit:   rateLimit,
		IsActive:    true,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	id, err := s.repo.Create(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	apiKey.ID = id
	return apiKey, nil
}

// GetAPIKey 获取API密钥
func (s *APIKeyService) GetAPIKey(id int64) (*models.APIKey, error) {
	return s.repo.GetByID(id)
}

// GetAPIKeyByKey 通过密钥获取API密钥信息
func (s *APIKeyService) GetAPIKeyByKey(key string) (*models.APIKey, error) {
	return s.repo.GetByKey(key)
}

// ListAPIKeys 获取用户的API密钥列表
func (s *APIKeyService) ListAPIKeys(userID int64, limit, offset int) ([]*models.APIKey, int64, error) {
	return s.repo.ListByUserID(userID, limit, offset)
}

// UpdateAPIKey 更新API密钥
func (s *APIKeyService) UpdateAPIKey(id int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return s.repo.Update(id, updates)
}

// DeleteAPIKey 删除API密钥
func (s *APIKeyService) DeleteAPIKey(id int64) error {
	return s.repo.Delete(id)
}

// RegenerateAPIKey 重新生成API密钥
func (s *APIKeyService) RegenerateAPIKey(id int64) (*models.APIKey, error) {
	// 获取现有密钥
	apiKey, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// 生成新的密钥
	newKey, newSecret, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new API key: %w", err)
	}

	// 更新密钥
	updates := map[string]interface{}{
		"key":         newKey,
		"secret_hash": s.hashSecret(newSecret),
		"updated_at":  time.Now(),
	}

	if err := s.repo.Update(id, updates); err != nil {
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	// 返回更新后的密钥
	apiKey.Key = newKey
	apiKey.SecretHash = s.hashSecret(newSecret)
	apiKey.UpdatedAt = time.Now()

	return apiKey, nil
}

// ValidateAPIKey 验证API密钥
func (s *APIKeyService) ValidateAPIKey(key, secret string) (*models.APIKey, error) {
	// 获取API密钥
	apiKey, err := s.repo.GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	// 检查密钥是否激?	if !apiKey.IsActive {
		return nil, fmt.Errorf("API key is inactive")
	}

	// 检查密钥是否过?	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// 验证密钥
	if !s.verifySecret(secret, apiKey.SecretHash) {
		return nil, fmt.Errorf("invalid API secret")
	}

	// 更新最后使用时?	s.repo.Update(apiKey.ID, map[string]interface{}{
		"last_used_at": time.Now(),
	})

	return apiKey, nil
}

// CheckPermission 检查API密钥权限
func (s *APIKeyService) CheckPermission(apiKey *models.APIKey, permission string) bool {
	// 检查是否有通配符权?	for _, perm := range apiKey.Permissions {
		if perm == "*" || perm == permission {
			return true
		}
		// 支持通配符匹?		if strings.HasSuffix(perm, "*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(permission, prefix) {
				return true
			}
		}
	}
	return false
}

// RevokeAPIKey 撤销API密钥
func (s *APIKeyService) RevokeAPIKey(id int64) error {
	return s.repo.Update(id, map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	})
}

// GetAPIKeyUsageStats 获取API密钥使用统计
func (s *APIKeyService) GetAPIKeyUsageStats(keyID int64, startTime, endTime time.Time) (map[string]interface{}, error) {
	// 这里可以集成监控服务获取使用统计
	// 暂时返回模拟数据
	stats := map[string]interface{}{
		"total_requests":    1000,
		"successful_requests": 950,
		"failed_requests":   50,
		"rate_limit_hits":   5,
		"last_request_at":   time.Now().Add(-1 * time.Hour),
	}
	return stats, nil
}

// generateAPIKey 生成API密钥和密?func (s *APIKeyService) generateAPIKey() (string, string, error) {
	// 生成API密钥 (32字节)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", "", err
	}
	key := "tslj_" + hex.EncodeToString(keyBytes)

	// 生成密钥 (64字节)
	secretBytes := make([]byte, 64)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}
	secret := hex.EncodeToString(secretBytes)

	return key, secret, nil
}

// hashSecret 对密钥进行哈?func (s *APIKeyService) hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// verifySecret 验证密钥
func (s *APIKeyService) verifySecret(secret, hash string) bool {
	return s.hashSecret(secret) == hash
}

// CleanupExpiredKeys 清理过期的API密钥
func (s *APIKeyService) CleanupExpiredKeys() error {
	return s.repo.DeleteExpired()
}

