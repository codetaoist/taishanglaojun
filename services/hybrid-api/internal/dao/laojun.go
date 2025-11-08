package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// Laojun Domain DAOs

type ConfigDAO struct {
	db *sql.DB
}

func NewConfigDAO(db *sql.DB) *ConfigDAO {
	return &ConfigDAO{db: db}
}

func (d *ConfigDAO) Get(ctx context.Context, tenantID, key string) (*models.Config, error) {
	query := `SELECT key, value, scope, tenant_id, updated_at FROM lao_configs WHERE tenant_id = $1 AND key = $2`
	row := d.db.QueryRowContext(ctx, query, tenantID, key)
	
	var config models.Config
	err := row.Scan(&config.Key, &config.Value, &config.Scope, &config.TenantID, &config.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *ConfigDAO) Set(ctx context.Context, config *models.Config) error {
	query := `
		INSERT INTO lao_configs (key, value, scope, tenant_id, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key, tenant_id)
		DO UPDATE SET value = $2, scope = $3, updated_at = $5
	`
	_, err := d.db.ExecContext(ctx, query, config.Key, config.Value, config.Scope, config.TenantID, time.Now())
	return err
}

type PluginDAO struct {
	db *sql.DB
}

func NewPluginDAO(db *sql.DB) *PluginDAO {
	return &PluginDAO{db: db}
}

func (d *PluginDAO) List(ctx context.Context, tenantID, status, name string, page, pageSize int) ([]*models.Plugin, int, error) {
	offset := (page - 1) * pageSize
	
	query := `
		SELECT id, tenant_id, name, description, status, created_at, updated_at
		FROM lao_plugins
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIndex := 2
	
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}
	
	if name != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR id::text ILIKE $%d)", argIndex, argIndex+1)
		args = append(args, "%"+name+"%", "%"+name+"%")
		argIndex += 2
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM lao_plugins WHERE " + query[len("FROM lao_plugins WHERE "):len(query)]
	var total int
	err := d.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var plugins []*models.Plugin
	for rows.Next() {
		var plugin models.Plugin
		err := rows.Scan(
			&plugin.ID, &plugin.TenantID, &plugin.Name, &plugin.Description,
			&plugin.Status, &plugin.CreatedAt, &plugin.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		plugins = append(plugins, &plugin)
	}
	
	return plugins, total, nil
}

func (d *PluginDAO) GetByID(ctx context.Context, tenantID, id string) (*models.Plugin, error) {
	query := `
		SELECT id, tenant_id, name, description, status, created_at, updated_at
		FROM lao_plugins
		WHERE tenant_id = $1 AND id = $2
	`
	row := d.db.QueryRowContext(ctx, query, tenantID, id)
	
	var plugin models.Plugin
	err := row.Scan(
		&plugin.ID, &plugin.TenantID, &plugin.Name, &plugin.Description,
		&plugin.Status, &plugin.CreatedAt, &plugin.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

func (d *PluginDAO) Create(ctx context.Context, plugin *models.Plugin) error {
	query := `
		INSERT INTO lao_plugins (id, tenant_id, name, description, version, source, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := d.db.ExecContext(ctx, query,
		plugin.ID, plugin.TenantID, plugin.Name, plugin.Description,
		plugin.Version, plugin.Source, plugin.Status, plugin.Config,
		time.Now(), time.Now(),
	)
	return err
}

func (d *PluginDAO) Update(ctx context.Context, plugin *models.Plugin) error {
	query := `
		UPDATE lao_plugins
		SET name = $3, description = $4, status = $5, updated_at = $6
		WHERE tenant_id = $1 AND id = $2
	`
	_, err := d.db.ExecContext(ctx, query,
		plugin.TenantID, plugin.ID, plugin.Name, plugin.Description,
		plugin.Status, time.Now(),
	)
	return err
}

func (d *PluginDAO) SetStatus(ctx context.Context, tenantID, id, status string) error {
	query := `
		UPDATE lao_plugins
		SET status = $3, updated_at = $4
		WHERE tenant_id = $1 AND id = $2
	`
	_, err := d.db.ExecContext(ctx, query, tenantID, id, status, time.Now())
	return err
}

func (d *PluginDAO) Upgrade(ctx context.Context, tenantID, id, version string) error {
	query := `
		UPDATE lao_plugins
		SET version = $3, status = 'installed', updated_at = $4
		WHERE tenant_id = $1 AND id = $2
	`
	_, err := d.db.ExecContext(ctx, query, tenantID, id, version, time.Now())
	return err
}

func (d *PluginDAO) Delete(ctx context.Context, tenantID, id string) error {
	query := `DELETE FROM lao_plugins WHERE tenant_id = $1 AND id = $2`
	_, err := d.db.ExecContext(ctx, query, tenantID, id)
	return err
}

type AuditLogDAO struct {
	db *sql.DB
}

func NewAuditLogDAO(db *sql.DB) *AuditLogDAO {
	return &AuditLogDAO{db: db}
}

func (d *AuditLogDAO) Create(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO lao_audit_logs (tenant_id, actor, actor_type, action, target_type, target_id, payload, result, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := d.db.ExecContext(ctx, query,
		log.TenantID, log.Actor, log.ActorType, log.Action, log.TargetType,
		log.TargetID, log.Payload, log.Result, log.IPAddress, log.UserAgent, time.Now(),
	)
	return err
}

func (d *AuditLogDAO) List(ctx context.Context, tenantID, actor, action, targetType string, page, pageSize int) ([]*models.AuditLog, int, error) {
	offset := (page - 1) * pageSize
	
	query := `
		SELECT id, tenant_id, actor, actor_type, action, target_type, target_id, payload, result, ip_address, user_agent, created_at
		FROM lao_audit_logs
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIndex := 2
	
	if actor != "" {
		query += fmt.Sprintf(" AND actor = $%d", argIndex)
		args = append(args, actor)
		argIndex++
	}
	
	if action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}
	
	if targetType != "" {
		query += fmt.Sprintf(" AND target_type = $%d", argIndex)
		args = append(args, targetType)
		argIndex++
	}
	
	// Get total count
	countQuery := "SELECT COUNT(*) FROM lao_audit_logs WHERE " + query[len("FROM lao_audit_logs WHERE "):len(query)]
	var total int
	err := d.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)
	
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var logs []*models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.Actor, &log.ActorType, &log.Action,
			&log.TargetType, &log.TargetID, &log.Payload, &log.Result,
			&log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}
	
	return logs, total, nil
}