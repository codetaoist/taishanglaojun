package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
)

// IntegrationRepository 
type IntegrationRepository struct {
	db *sql.DB
}

// NewIntegrationRepository 
func NewIntegrationRepository(db *sql.DB) *IntegrationRepository {
	return &IntegrationRepository{
		db: db,
	}
}

// Create 
func (r *IntegrationRepository) Create(integration *models.Integration) (int64, error) {
	configJSON, err := json.Marshal(integration.Config)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config: %w", err)
	}

	settingsJSON, err := json.Marshal(integration.Settings)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO integrations (user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		integration.UserID,
		integration.Name,
		integration.Provider,
		integration.Type,
		configJSON,
		settingsJSON,
		integration.Status,
		integration.SyncInterval,
		integration.LastSyncAt,
		integration.IsActive,
		integration.CreatedAt,
		integration.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create integration: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetByID ID
func (r *IntegrationRepository) GetByID(id int64) (*models.Integration, error) {
	query := `
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	return r.scanIntegration(row)
}

// GetByName 
func (r *IntegrationRepository) GetByName(userID int64, name string) (*models.Integration, error) {
	query := `
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		WHERE user_id = ? AND name = ?
	`

	row := r.db.QueryRow(query, userID, name)
	return r.scanIntegration(row)
}

// ListByUserID ID
func (r *IntegrationRepository) ListByUserID(userID int64, provider string, limit, offset int) ([]*models.Integration, int64, error) {
	// 
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if provider != "" {
		whereClause += " AND provider = ?"
		args = append(args, provider)
	}

	// 
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM integrations %s", whereClause)
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count integrations: %w", err)
	}

	// 
	query := fmt.Sprintf(`
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*models.Integration
	for rows.Next() {
		integration, err := r.scanIntegrationFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan integration: %w", err)
		}
		integrations = append(integrations, integration)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return integrations, total, nil
}

// ListByProvider ?func (r *IntegrationRepository) ListByProvider(provider string, limit, offset int) ([]*models.Integration, int64, error) {
	// 
	countQuery := `SELECT COUNT(*) FROM integrations WHERE provider = ?`
	var total int64
	err := r.db.QueryRow(countQuery, provider).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count integrations: %w", err)
	}

	// 
	query := `
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		WHERE provider = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, provider, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*models.Integration
	for rows.Next() {
		integration, err := r.scanIntegrationFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan integration: %w", err)
		}
		integrations = append(integrations, integration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return integrations, total, nil
}

// ListActive ?func (r *IntegrationRepository) ListActive(userID int64) ([]*models.Integration, error) {
	query := `
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		WHERE user_id = ? AND is_active = 1 AND status IN (?, ?)
		ORDER BY name
	`

	rows, err := r.db.Query(query, userID, models.IntegrationStatusActive, models.IntegrationStatusSyncing)
	if err != nil {
		return nil, fmt.Errorf("failed to query active integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*models.Integration
	for rows.Next() {
		integration, err := r.scanIntegrationFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan integration: %w", err)
		}
		integrations = append(integrations, integration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return integrations, nil
}

// Update 
func (r *IntegrationRepository) Update(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		// JSON
		if field == "config" || field == "settings" {
			if configMap, ok := value.(map[string]interface{}); ok {
				configJSON, err := json.Marshal(configMap)
				if err != nil {
					return fmt.Errorf("failed to marshal %s: %w", field, err)
				}
				value = configJSON
			}
		}
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE integrations SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update integration: %w", err)
	}

	return nil
}

// Delete 
func (r *IntegrationRepository) Delete(id int64) error {
	query := `DELETE FROM integrations WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("integration not found")
	}

	return nil
}

// GetIntegrationsForSync 
func (r *IntegrationRepository) GetIntegrationsForSync() ([]*models.Integration, error) {
	query := `
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		WHERE is_active = 1 
		AND status = ?
		AND (last_sync_at IS NULL OR last_sync_at < datetime('now', '-' || sync_interval || ' seconds'))
	`

	rows, err := r.db.Query(query, models.IntegrationStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to query integrations for sync: %w", err)
	}
	defer rows.Close()

	var integrations []*models.Integration
	for rows.Next() {
		integration, err := r.scanIntegrationFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan integration: %w", err)
		}
		integrations = append(integrations, integration)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return integrations, nil
}

// GetIntegrationStats 
func (r *IntegrationRepository) GetIntegrationStats(userID int64) (map[string]interface{}, error) {
	// 
	totalQuery := `SELECT COUNT(*) FROM integrations WHERE user_id = ?`
	var totalIntegrations int64
	err := r.db.QueryRow(totalQuery, userID).Scan(&totalIntegrations)
	if err != nil {
		return nil, fmt.Errorf("failed to count total integrations: %w", err)
	}

	// 
	activeQuery := `SELECT COUNT(*) FROM integrations WHERE user_id = ? AND is_active = 1`
	var activeIntegrations int64
	err = r.db.QueryRow(activeQuery, userID).Scan(&activeIntegrations)
	if err != nil {
		return nil, fmt.Errorf("failed to count active integrations: %w", err)
	}

	// 24
	recentSyncQuery := `
		SELECT COUNT(*) FROM integrations 
		WHERE user_id = ? AND last_sync_at > ?
	`
	var recentSynced int64
	since := time.Now().Add(-24 * time.Hour)
	err = r.db.QueryRow(recentSyncQuery, userID, since).Scan(&recentSynced)
	if err != nil {
		return nil, fmt.Errorf("failed to count recently synced integrations: %w", err)
	}

	// 
	providerStats := make(map[string]int64)
	providerQuery := `
		SELECT provider, COUNT(*) 
		FROM integrations 
		WHERE user_id = ? 
		GROUP BY provider
	`
	rows, err := r.db.Query(providerQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var provider string
		var count int64
		if err := rows.Scan(&provider, &count); err != nil {
			return nil, fmt.Errorf("failed to scan provider stats: %w", err)
		}
		providerStats[provider] = count
	}

	// ?	statusStats := make(map[string]int64)
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM integrations 
		WHERE user_id = ? 
		GROUP BY status
	`
	rows, err = r.db.Query(statusQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query status stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status stats: %w", err)
		}
		statusStats[status] = count
	}

	stats := map[string]interface{}{
		"total_integrations":  totalIntegrations,
		"active_integrations": activeIntegrations,
		"recent_synced":       recentSynced,
		"provider_stats":      providerStats,
		"status_stats":        statusStats,
	}

	return stats, nil
}

// SearchIntegrations 
func (r *IntegrationRepository) SearchIntegrations(userID int64, keyword string, limit, offset int) ([]*models.Integration, int64, error) {
	searchPattern := "%" + keyword + "%"

	// 
	countQuery := `
		SELECT COUNT(*) FROM integrations 
		WHERE user_id = ? AND (name LIKE ? OR provider LIKE ?)
	`
	var total int64
	err := r.db.QueryRow(countQuery, userID, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// 
	query := `
		SELECT id, user_id, name, provider, type, config, settings,
			status, sync_interval, last_sync_at, is_active, created_at, updated_at
		FROM integrations
		WHERE user_id = ? AND (name LIKE ? OR provider LIKE ?)
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search integrations: %w", err)
	}
	defer rows.Close()

	var integrations []*models.Integration
	for rows.Next() {
		integration, err := r.scanIntegrationFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan integration: %w", err)
		}
		integrations = append(integrations, integration)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return integrations, total, nil
}

// scanIntegration 赥Integration
func (r *IntegrationRepository) scanIntegration(row *sql.Row) (*models.Integration, error) {
	var integration models.Integration
	var configJSON, settingsJSON []byte
	var lastSyncAt sql.NullTime

	err := row.Scan(
		&integration.ID,
		&integration.UserID,
		&integration.Name,
		&integration.Provider,
		&integration.Type,
		&configJSON,
		&settingsJSON,
		&integration.Status,
		&integration.SyncInterval,
		&lastSyncAt,
		&integration.IsActive,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("integration not found")
		}
		return nil, fmt.Errorf("failed to scan integration: %w", err)
	}

	// JSON
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		integration.Config = make(map[string]interface{})
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &integration.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	} else {
		integration.Settings = make(map[string]interface{})
	}

	// 
	if lastSyncAt.Valid {
		integration.LastSyncAt = &lastSyncAt.Time
	}

	return &integration, nil
}

// scanIntegrationFromRows Integration
func (r *IntegrationRepository) scanIntegrationFromRows(rows *sql.Rows) (*models.Integration, error) {
	var integration models.Integration
	var configJSON, settingsJSON []byte
	var lastSyncAt sql.NullTime

	err := rows.Scan(
		&integration.ID,
		&integration.UserID,
		&integration.Name,
		&integration.Provider,
		&integration.Type,
		&configJSON,
		&settingsJSON,
		&integration.Status,
		&integration.SyncInterval,
		&lastSyncAt,
		&integration.IsActive,
		&integration.CreatedAt,
		&integration.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan integration: %w", err)
	}

	// JSON
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &integration.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		integration.Config = make(map[string]interface{})
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &integration.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	} else {
		integration.Settings = make(map[string]interface{})
	}

	// 
	if lastSyncAt.Valid {
		integration.LastSyncAt = &lastSyncAt.Time
	}

	return &integration, nil
}

