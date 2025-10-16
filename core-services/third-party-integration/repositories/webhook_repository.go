package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
)

// WebhookRepository Webhook数据仓库
type WebhookRepository struct {
	db *sql.DB
}

// NewWebhookRepository 创建新的Webhook仓库
func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{
		db: db,
	}
}

// Create 创建Webhook
func (r *WebhookRepository) Create(webhook *models.Webhook) (int64, error) {
	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal events: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		INSERT INTO webhooks (user_id, name, url, secret, events, headers, status,
			is_active, retry_count, timeout, last_triggered_at, last_success_at,
			last_error, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		webhook.UserID,
		webhook.Name,
		webhook.URL,
		webhook.Secret,
		eventsJSON,
		headersJSON,
		webhook.Status,
		webhook.IsActive,
		webhook.RetryCount,
		webhook.Timeout,
		webhook.LastTriggeredAt,
		webhook.LastSuccessAt,
		webhook.LastError,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create webhook: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetByID 根据ID获取Webhook
func (r *WebhookRepository) GetByID(id int64) (*models.Webhook, error) {
	query := `
		SELECT id, user_id, name, url, secret, events, headers, status,
			is_active, retry_count, timeout, last_triggered_at, last_success_at,
			last_error, created_at, updated_at
		FROM webhooks
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	return r.scanWebhook(row)
}

// ListByUserID 根据用户ID获取Webhook列表
func (r *WebhookRepository) ListByUserID(userID int64, limit, offset int) ([]*models.Webhook, int64, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM webhooks WHERE user_id = ?`
	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count webhooks: %w", err)
	}

	// 获取列表
	query := `
		SELECT id, user_id, name, url, secret, events, headers, status,
			is_active, retry_count, timeout, last_triggered_at, last_success_at,
			last_error, created_at, updated_at
		FROM webhooks
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*models.Webhook
	for rows.Next() {
		webhook, err := r.scanWebhookFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return webhooks, total, nil
}

// Update 更新Webhook
func (r *WebhookRepository) Update(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		// 特殊处理JSON字段
		if field == "events" || field == "headers" {
			if data, ok := value.([]string); ok && field == "events" {
				jsonData, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal events: %w", err)
				}
				value = jsonData
			} else if data, ok := value.(map[string]string); ok && field == "headers" {
				jsonData, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal headers: %w", err)
				}
				value = jsonData
			}
		}
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE webhooks SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	return nil
}

// Delete 删除Webhook
func (r *WebhookRepository) Delete(id int64) error {
	query := `DELETE FROM webhooks WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// scanWebhook 扫描单行数据到Webhook结构
func (r *WebhookRepository) scanWebhook(row *sql.Row) (*models.Webhook, error) {
	var webhook models.Webhook
	var eventsJSON, headersJSON []byte
	var lastTriggeredAt, lastSuccessAt sql.NullTime
	var lastError sql.NullString

	err := row.Scan(
		&webhook.ID,
		&webhook.UserID,
		&webhook.Name,
		&webhook.URL,
		&webhook.Secret,
		&eventsJSON,
		&headersJSON,
		&webhook.Status,
		&webhook.IsActive,
		&webhook.RetryCount,
		&webhook.Timeout,
		&lastTriggeredAt,
		&lastSuccessAt,
		&lastError,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("webhook not found")
		}
		return nil, fmt.Errorf("failed to scan webhook: %w", err)
	}

	// 解析JSON字段
	if len(eventsJSON) > 0 {
		if err := json.Unmarshal(eventsJSON, &webhook.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}
	}

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	} else {
		webhook.Headers = make(map[string]string)
	}

	// 处理可空字段
	if lastTriggeredAt.Valid {
		webhook.LastTriggeredAt = &lastTriggeredAt.Time
	}
	if lastSuccessAt.Valid {
		webhook.LastSuccessAt = &lastSuccessAt.Time
	}
	if lastError.Valid {
		webhook.LastError = lastError.String
	}

	return &webhook, nil
}

// scanWebhookFromRows 从多行查询结果扫描Webhook
func (r *WebhookRepository) scanWebhookFromRows(rows *sql.Rows) (*models.Webhook, error) {
	var webhook models.Webhook
	var eventsJSON, headersJSON []byte
	var lastTriggeredAt, lastSuccessAt sql.NullTime
	var lastError sql.NullString

	err := rows.Scan(
		&webhook.ID,
		&webhook.UserID,
		&webhook.Name,
		&webhook.URL,
		&webhook.Secret,
		&eventsJSON,
		&headersJSON,
		&webhook.Status,
		&webhook.IsActive,
		&webhook.RetryCount,
		&webhook.Timeout,
		&lastTriggeredAt,
		&lastSuccessAt,
		&lastError,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan webhook: %w", err)
	}

	// 解析JSON字段
	if len(eventsJSON) > 0 {
		if err := json.Unmarshal(eventsJSON, &webhook.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}
	}

	if len(headersJSON) > 0 {
		if err := json.Unmarshal(headersJSON, &webhook.Headers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
		}
	} else {
		webhook.Headers = make(map[string]string)
	}

	// 处理可空字段
	if lastTriggeredAt.Valid {
		webhook.LastTriggeredAt = &lastTriggeredAt.Time
	}
	if lastSuccessAt.Valid {
		webhook.LastSuccessAt = &lastSuccessAt.Time
	}
	if lastError.Valid {
		webhook.LastError = lastError.String
	}

	return &webhook, nil
}

