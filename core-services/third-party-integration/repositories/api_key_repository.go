package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
)

// APIKeyRepository API密钥数据仓库
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository 创建新的API密钥仓库
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{
		db: db,
	}
}

// Create 创建API密钥
func (r *APIKeyRepository) Create(apiKey *models.APIKey) (int64, error) {
	query := `
		INSERT INTO api_keys (user_id, name, key_hash, key_prefix, permissions, rate_limit, 
			expires_at, last_used_at, usage_count, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		apiKey.UserID,
		apiKey.Name,
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		strings.Join(apiKey.Permissions, ","),
		apiKey.RateLimit,
		apiKey.ExpiresAt,
		apiKey.LastUsedAt,
		apiKey.UsageCount,
		apiKey.IsActive,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create API key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetByID 根据ID获取API密钥
func (r *APIKeyRepository) GetByID(id int64) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, permissions, rate_limit,
			expires_at, last_used_at, usage_count, is_active, created_at, updated_at
		FROM api_keys
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	return r.scanAPIKey(row)
}

// GetByKeyHash 根据密钥哈希获取API密钥
func (r *APIKeyRepository) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, permissions, rate_limit,
			expires_at, last_used_at, usage_count, is_active, created_at, updated_at
		FROM api_keys
		WHERE key_hash = ? AND is_active = 1
	`

	row := r.db.QueryRow(query, keyHash)
	return r.scanAPIKey(row)
}

// GetByPrefix 根据前缀获取API密钥
func (r *APIKeyRepository) GetByPrefix(prefix string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, permissions, rate_limit,
			expires_at, last_used_at, usage_count, is_active, created_at, updated_at
		FROM api_keys
		WHERE key_prefix = ? AND is_active = 1
	`

	row := r.db.QueryRow(query, prefix)
	return r.scanAPIKey(row)
}

// ListByUserID 根据用户ID获取API密钥列表
func (r *APIKeyRepository) ListByUserID(userID int64, limit, offset int) ([]*models.APIKey, int64, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM api_keys WHERE user_id = ?`
	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	// 获取列表
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, permissions, rate_limit,
			expires_at, last_used_at, usage_count, is_active, created_at, updated_at
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*models.APIKey
	for rows.Next() {
		apiKey, err := r.scanAPIKeyFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return apiKeys, total, nil
}

// Update 更新API密钥
func (r *APIKeyRepository) Update(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE api_keys SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// Delete 删除API密钥
func (r *APIKeyRepository) Delete(id int64) error {
	query := `DELETE FROM api_keys WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// UpdateUsage 更新使用统计
func (r *APIKeyRepository) UpdateUsage(id int64) error {
	query := `
		UPDATE api_keys 
		SET usage_count = usage_count + 1, last_used_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to update API key usage: %w", err)
	}

	return nil
}

// GetExpiredKeys 获取过期的API密钥
func (r *APIKeyRepository) GetExpiredKeys() ([]*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, key_prefix, permissions, rate_limit,
			expires_at, last_used_at, usage_count, is_active, created_at, updated_at
		FROM api_keys
		WHERE expires_at IS NOT NULL AND expires_at < ? AND is_active = 1
	`

	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query expired API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*models.APIKey
	for rows.Next() {
		apiKey, err := r.scanAPIKeyFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return apiKeys, nil
}

// GetUsageStats 获取使用统计
func (r *APIKeyRepository) GetUsageStats(userID int64, days int) (map[string]interface{}, error) {
	// 获取总的API密钥数量
	totalQuery := `SELECT COUNT(*) FROM api_keys WHERE user_id = ?`
	var totalKeys int64
	err := r.db.QueryRow(totalQuery, userID).Scan(&totalKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to count total keys: %w", err)
	}

	// 获取活跃的API密钥数量
	activeQuery := `SELECT COUNT(*) FROM api_keys WHERE user_id = ? AND is_active = 1`
	var activeKeys int64
	err = r.db.QueryRow(activeQuery, userID).Scan(&activeKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to count active keys: %w", err)
	}

	// 获取总使用次�?	usageQuery := `SELECT COALESCE(SUM(usage_count), 0) FROM api_keys WHERE user_id = ?`
	var totalUsage int64
	err = r.db.QueryRow(usageQuery, userID).Scan(&totalUsage)
	if err != nil {
		return nil, fmt.Errorf("failed to get total usage: %w", err)
	}

	// 获取最近使用的API密钥
	recentQuery := `
		SELECT COUNT(*) FROM api_keys 
		WHERE user_id = ? AND last_used_at > ?
	`
	var recentlyUsed int64
	since := time.Now().AddDate(0, 0, -days)
	err = r.db.QueryRow(recentQuery, userID, since).Scan(&recentlyUsed)
	if err != nil {
		return nil, fmt.Errorf("failed to count recently used keys: %w", err)
	}

	stats := map[string]interface{}{
		"total_keys":     totalKeys,
		"active_keys":    activeKeys,
		"total_usage":    totalUsage,
		"recently_used":  recentlyUsed,
		"days":           days,
	}

	return stats, nil
}

// scanAPIKey 扫描单行数据到APIKey结构
func (r *APIKeyRepository) scanAPIKey(row *sql.Row) (*models.APIKey, error) {
	var apiKey models.APIKey
	var permissionsStr string
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := row.Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeyHash,
		&apiKey.KeyPrefix,
		&permissionsStr,
		&apiKey.RateLimit,
		&expiresAt,
		&lastUsedAt,
		&apiKey.UsageCount,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to scan API key: %w", err)
	}

	// 解析权限
	if permissionsStr != "" {
		apiKey.Permissions = strings.Split(permissionsStr, ",")
	}

	// 处理可空时间字段
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}

	return &apiKey, nil
}

// scanAPIKeyFromRows 从多行查询结果扫描APIKey
func (r *APIKeyRepository) scanAPIKeyFromRows(rows *sql.Rows) (*models.APIKey, error) {
	var apiKey models.APIKey
	var permissionsStr string
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := rows.Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeyHash,
		&apiKey.KeyPrefix,
		&permissionsStr,
		&apiKey.RateLimit,
		&expiresAt,
		&lastUsedAt,
		&apiKey.UsageCount,
		&apiKey.IsActive,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan API key: %w", err)
	}

	// 解析权限
	if permissionsStr != "" {
		apiKey.Permissions = strings.Split(permissionsStr, ",")
	}

	// 处理可空时间字段
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}

	return &apiKey, nil
}
