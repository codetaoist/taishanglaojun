package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
)

// OAuthRepository OAuth数据仓库
type OAuthRepository struct {
	db *sql.DB
}

// NewOAuthRepository 创建新的OAuth仓库
func NewOAuthRepository(db *sql.DB) *OAuthRepository {
	return &OAuthRepository{
		db: db,
	}
}

// CreateApp 创建OAuth应用
func (r *OAuthRepository) CreateApp(app *models.OAuthApp) (int64, error) {
	scopesJSON, err := json.Marshal(app.Scopes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scopes: %w", err)
	}

	redirectURIsJSON, err := json.Marshal(app.RedirectURIs)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal redirect URIs: %w", err)
	}

	query := `
		INSERT INTO oauth_apps (user_id, name, description, client_id, client_secret,
			scopes, redirect_uris, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		app.UserID,
		app.Name,
		app.Description,
		app.ClientID,
		app.ClientSecret,
		scopesJSON,
		redirectURIsJSON,
		app.IsActive,
		app.CreatedAt,
		app.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create oauth app: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetAppByID 根据ID获取OAuth应用
func (r *OAuthRepository) GetAppByID(id int64) (*models.OAuthApp, error) {
	query := `
		SELECT id, user_id, name, description, client_id, client_secret,
			scopes, redirect_uris, is_active, created_at, updated_at
		FROM oauth_apps
		WHERE id = ?
	`

	row := r.db.QueryRow(query, id)
	return r.scanOAuthApp(row)
}

// GetAppByClientID 根据ClientID获取OAuth应用
func (r *OAuthRepository) GetAppByClientID(clientID string) (*models.OAuthApp, error) {
	query := `
		SELECT id, user_id, name, description, client_id, client_secret,
			scopes, redirect_uris, is_active, created_at, updated_at
		FROM oauth_apps
		WHERE client_id = ?
	`

	row := r.db.QueryRow(query, clientID)
	return r.scanOAuthApp(row)
}

// ListAppsByUserID 根据用户ID获取OAuth应用列表
func (r *OAuthRepository) ListAppsByUserID(userID int64, limit, offset int) ([]*models.OAuthApp, int64, error) {
	// 获取总数
	countQuery := `SELECT COUNT(*) FROM oauth_apps WHERE user_id = ?`
	var total int64
	err := r.db.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count oauth apps: %w", err)
	}

	// 获取列表
	query := `
		SELECT id, user_id, name, description, client_id, client_secret,
			scopes, redirect_uris, is_active, created_at, updated_at
		FROM oauth_apps
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query oauth apps: %w", err)
	}
	defer rows.Close()

	var apps []*models.OAuthApp
	for rows.Next() {
		app, err := r.scanOAuthAppFromRows(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan oauth app: %w", err)
		}
		apps = append(apps, app)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return apps, total, nil
}

// UpdateApp 更新OAuth应用
func (r *OAuthRepository) UpdateApp(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		// 特殊处理JSON字段
		if field == "scopes" || field == "redirect_uris" {
			if data, ok := value.([]string); ok {
				jsonData, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal %s: %w", field, err)
				}
				value = jsonData
			}
		}
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE oauth_apps SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update oauth app: %w", err)
	}

	return nil
}

// DeleteApp 删除OAuth应用
func (r *OAuthRepository) DeleteApp(id int64) error {
	query := `DELETE FROM oauth_apps WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oauth app: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oauth app not found")
	}

	return nil
}

// CreateToken 创建OAuth令牌
func (r *OAuthRepository) CreateToken(token *models.OAuthToken) (int64, error) {
	scopesJSON, err := json.Marshal(token.Scopes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scopes: %w", err)
	}

	query := `
		INSERT INTO oauth_tokens (app_id, user_id, access_token, refresh_token,
			token_type, scopes, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		token.AppID,
		token.UserID,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		scopesJSON,
		token.ExpiresAt,
		token.CreatedAt,
		token.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create oauth token: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return id, nil
}

// GetTokenByAccessToken 根据访问令牌获取OAuth令牌
func (r *OAuthRepository) GetTokenByAccessToken(accessToken string) (*models.OAuthToken, error) {
	query := `
		SELECT id, app_id, user_id, access_token, refresh_token,
			token_type, scopes, expires_at, created_at, updated_at
		FROM oauth_tokens
		WHERE access_token = ?
	`

	row := r.db.QueryRow(query, accessToken)
	return r.scanOAuthToken(row)
}

// GetTokenByRefreshToken 根据刷新令牌获取OAuth令牌
func (r *OAuthRepository) GetTokenByRefreshToken(refreshToken string) (*models.OAuthToken, error) {
	query := `
		SELECT id, app_id, user_id, access_token, refresh_token,
			token_type, scopes, expires_at, created_at, updated_at
		FROM oauth_tokens
		WHERE refresh_token = ?
	`

	row := r.db.QueryRow(query, refreshToken)
	return r.scanOAuthToken(row)
}

// UpdateToken 更新OAuth令牌
func (r *OAuthRepository) UpdateToken(id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)

	for field, value := range updates {
		// 特殊处理JSON字段
		if field == "scopes" {
			if data, ok := value.([]string); ok {
				jsonData, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal scopes: %w", err)
				}
				value = jsonData
			}
		}
		setParts = append(setParts, field+" = ?")
		args = append(args, value)
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE oauth_tokens SET %s WHERE id = ?", strings.Join(setParts, ", "))

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update oauth token: %w", err)
	}

	return nil
}

// DeleteToken 删除OAuth令牌
func (r *OAuthRepository) DeleteToken(id int64) error {
	query := `DELETE FROM oauth_tokens WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oauth token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oauth token not found")
	}

	return nil
}

// DeleteExpiredTokens 删除过期的令?
func (r *OAuthRepository) DeleteExpiredTokens() error {
	query := `DELETE FROM oauth_tokens WHERE expires_at < ?`

	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}

// scanOAuthApp 扫描单行数据到OAuthApp结构
func (r *OAuthRepository) scanOAuthApp(row *sql.Row) (*models.OAuthApp, error) {
	var app models.OAuthApp
	var scopesJSON, redirectURIsJSON []byte

	err := row.Scan(
		&app.ID,
		&app.UserID,
		&app.Name,
		&app.Description,
		&app.ClientID,
		&app.ClientSecret,
		&scopesJSON,
		&redirectURIsJSON,
		&app.IsActive,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oauth app not found")
		}
		return nil, fmt.Errorf("failed to scan oauth app: %w", err)
	}

	// 解析JSON字段
	if len(scopesJSON) > 0 {
		if err := json.Unmarshal(scopesJSON, &app.Scopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
		}
	}

	if len(redirectURIsJSON) > 0 {
		if err := json.Unmarshal(redirectURIsJSON, &app.RedirectURIs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redirect URIs: %w", err)
		}
	}

	return &app, nil
}

// scanOAuthAppFromRows 从多行查询结果扫描OAuthApp
func (r *OAuthRepository) scanOAuthAppFromRows(rows *sql.Rows) (*models.OAuthApp, error) {
	var app models.OAuthApp
	var scopesJSON, redirectURIsJSON []byte

	err := rows.Scan(
		&app.ID,
		&app.UserID,
		&app.Name,
		&app.Description,
		&app.ClientID,
		&app.ClientSecret,
		&scopesJSON,
		&redirectURIsJSON,
		&app.IsActive,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan oauth app: %w", err)
	}

	// 解析JSON字段
	if len(scopesJSON) > 0 {
		if err := json.Unmarshal(scopesJSON, &app.Scopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
		}
	}

	if len(redirectURIsJSON) > 0 {
		if err := json.Unmarshal(redirectURIsJSON, &app.RedirectURIs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal redirect URIs: %w", err)
		}
	}

	return &app, nil
}

// scanOAuthToken 扫描单行数据到OAuthToken结构
func (r *OAuthRepository) scanOAuthToken(row *sql.Row) (*models.OAuthToken, error) {
	var token models.OAuthToken
	var scopesJSON []byte
	var refreshToken sql.NullString

	err := row.Scan(
		&token.ID,
		&token.AppID,
		&token.UserID,
		&token.AccessToken,
		&refreshToken,
		&token.TokenType,
		&scopesJSON,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oauth token not found")
		}
		return nil, fmt.Errorf("failed to scan oauth token: %w", err)
	}

	// 解析JSON字段
	if len(scopesJSON) > 0 {
		if err := json.Unmarshal(scopesJSON, &token.Scopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
		}
	}

	// 处理可空字段
	if refreshToken.Valid {
		token.RefreshToken = refreshToken.String
	}

	return &token, nil
}

