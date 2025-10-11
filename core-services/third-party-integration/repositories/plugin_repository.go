package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
)

// PluginRepository 插件数据仓库
type PluginRepository struct {
	db *sql.DB
}

// NewPluginRepository 创建新的插件仓库
func NewPluginRepository(db *sql.DB) *PluginRepository {
	return &PluginRepository{
		db: db,
	}
}

// Create 创建插件
func (r *PluginRepository) Create(plugin *models.Plugin) (int64, error) {
	configJSON, err := json.Marshal(plugin.Config)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO plugins (user_id, name, version, description, author, file_path, 
			config, status, is_enabled, installed_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		plugin.UserID,
		plugin.Name,
		plugin.Version,
		plugin.Description,
		plugin.Author,
		plugin.FilePath,
		configJSON,
		plugin.Status,
		plugin.IsEnabled,
		plugin.InstalledAt,
		plugin.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create plugin: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetByID 根据ID获取插件
func (r *PluginRepository) GetByID(id int64) (*models.Plugin, error) {
	query := `
		SELECT id, user_id, name, version, description, author, file_path,
			config, status, is_enabled, installed_at, updated_at
		FROM plugins
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	return r.scanPlugin(row)
}

// GetByName 根据名称获取插件
func (r *PluginRepository) GetByName(userID int64, name string) (*models.Plugin, error) {
	query := `
		SELECT id, user_id, name, version, description, author, file_path,
			config, status, is_enabled, installed_at, updated_at
		FROM plugins
		WHERE user_id = ? AND name = ?
	`

	row := r.db.QueryRow(query, userID, name)
	return r.scanPlugin(row)
}

// ListByUserID 根据用户ID获取插件列表
func (r *PluginRepository) ListByUserID(userID int64, status models.PluginStatus, limit, offset int) ([]*models.Plugin, int64, error) {
	// 构建查询条件
	whereClause := "WHERE user_id = ?"
	args := []interface{}{userID}

	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM plugins %s", whereClause)
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count plugins: %w", err)
	}

	// 获取列表
	query := fmt.Sprintf(`
		SELECT id, user_id, name, version, description, author, file_path,
			config, status, is_enabled, installed_at, updated_at
		FROM plugins
		%s
		ORDER BY installed_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.Plugin
	for rows.Next() {
		plugin, err := r.scanPluginFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return plugins, total, nil
}

// ListEnabled 获取启用的插件列�?func (r *PluginRepository) ListEnabled(userID int64) ([]*models.Plugin, error) {
	query := `
		SELECT id, user_id, name, version, description, author, file_path,
			config, status, is_enabled, installed_at, updated_at
		FROM plugins
		WHERE user_id = ? AND is_enabled = 1 AND status = ?
		ORDER BY name
	`

	rows, err := r.db.Query(query, userID, models.PluginStatusInstalled)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.Plugin
	for rows.Next() {
		plugin, err := r.scanPluginFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return plugins, nil
}

// Update 更新插件
func (r *PluginRepository) Update(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		// 特殊处理config字段
		if field == "config" {
			if configMap, ok := value.(map[string]interface{}); ok {
				configJSON, err := json.Marshal(configMap)
				if err != nil {
					return fmt.Errorf("failed to marshal config: %w", err)
				}
				value = configJSON
			}
		}
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE plugins SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}

	return nil
}

// Delete 删除插件
func (r *PluginRepository) Delete(id int64) error {
	query := `DELETE FROM plugins WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

// GetPluginStats 获取插件统计信息
func (r *PluginRepository) GetPluginStats(userID int64) (map[string]interface{}, error) {
	// 获取总插件数
	totalQuery := `SELECT COUNT(*) FROM plugins WHERE user_id = ?`
	var totalPlugins int64
	err := r.db.QueryRow(totalQuery, userID).Scan(&totalPlugins)
	if err != nil {
		return nil, fmt.Errorf("failed to count total plugins: %w", err)
	}

	// 获取启用的插件数
	enabledQuery := `SELECT COUNT(*) FROM plugins WHERE user_id = ? AND is_enabled = 1`
	var enabledPlugins int64
	err = r.db.QueryRow(enabledQuery, userID).Scan(&enabledPlugins)
	if err != nil {
		return nil, fmt.Errorf("failed to count enabled plugins: %w", err)
	}

	// 获取已安装的插件�?	installedQuery := `SELECT COUNT(*) FROM plugins WHERE user_id = ? AND status = ?`
	var installedPlugins int64
	err = r.db.QueryRow(installedQuery, userID, models.PluginStatusInstalled).Scan(&installedPlugins)
	if err != nil {
		return nil, fmt.Errorf("failed to count installed plugins: %w", err)
	}

	// 获取最近安装的插件数（7天内�?	recentQuery := `
		SELECT COUNT(*) FROM plugins 
		WHERE user_id = ? AND installed_at > ?
	`
	var recentPlugins int64
	since := time.Now().AddDate(0, 0, -7)
	err = r.db.QueryRow(recentQuery, userID, since).Scan(&recentPlugins)
	if err != nil {
		return nil, fmt.Errorf("failed to count recent plugins: %w", err)
	}

	// 按状态统�?	statusStats := make(map[string]int64)
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM plugins 
		WHERE user_id = ? 
		GROUP BY status
	`
	rows, err := r.db.Query(statusQuery, userID)
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
		"total_plugins":     totalPlugins,
		"enabled_plugins":   enabledPlugins,
		"installed_plugins": installedPlugins,
		"recent_plugins":    recentPlugins,
		"status_stats":      statusStats,
	}

	return stats, nil
}

// SearchPlugins 搜索插件
func (r *PluginRepository) SearchPlugins(userID int64, keyword string, limit, offset int) ([]*models.Plugin, int64, error) {
	searchPattern := "%" + keyword + "%"

	// 获取总数
	countQuery := `
		SELECT COUNT(*) FROM plugins 
		WHERE user_id = ? AND (name LIKE ? OR description LIKE ? OR author LIKE ?)
	`
	var total int64
	err := r.db.QueryRow(countQuery, userID, searchPattern, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// 获取搜索结果
	query := `
		SELECT id, user_id, name, version, description, author, file_path,
			config, status, is_enabled, installed_at, updated_at
		FROM plugins
		WHERE user_id = ? AND (name LIKE ? OR description LIKE ? OR author LIKE ?)
		ORDER BY name
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, searchPattern, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.Plugin
	for rows.Next() {
		plugin, err := r.scanPluginFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return plugins, total, nil
}

// scanPlugin 扫描单行数据到Plugin结构
func (r *PluginRepository) scanPlugin(row *sql.Row) (*models.Plugin, error) {
	var plugin models.Plugin
	var configJSON []byte

	err := row.Scan(
		&plugin.ID,
		&plugin.UserID,
		&plugin.Name,
		&plugin.Version,
		&plugin.Description,
		&plugin.Author,
		&plugin.FilePath,
		&configJSON,
		&plugin.Status,
		&plugin.IsEnabled,
		&plugin.InstalledAt,
		&plugin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found")
		}
		return nil, fmt.Errorf("failed to scan plugin: %w", err)
	}

	// 解析配置JSON
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &plugin.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		plugin.Config = make(map[string]interface{})
	}

	return &plugin, nil
}

// scanPluginFromRows 从多行查询结果扫描Plugin
func (r *PluginRepository) scanPluginFromRows(rows *sql.Rows) (*models.Plugin, error) {
	var plugin models.Plugin
	var configJSON []byte

	err := rows.Scan(
		&plugin.ID,
		&plugin.UserID,
		&plugin.Name,
		&plugin.Version,
		&plugin.Description,
		&plugin.Author,
		&plugin.FilePath,
		&configJSON,
		&plugin.Status,
		&plugin.IsEnabled,
		&plugin.InstalledAt,
		&plugin.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan plugin: %w", err)
	}

	// 解析配置JSON
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &plugin.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		plugin.Config = make(map[string]interface{})
	}

	return &plugin, nil
}