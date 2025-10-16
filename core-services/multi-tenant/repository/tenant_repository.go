package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	multitenant "github.com/codetaoist/taishanglaojun/core-services/multi-tenant"
	"go.uber.org/zap"
)

// TenantRepository 租户数据仓库实现
type TenantRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
	config TenantRepositoryConfig
}

// TenantRepositoryConfig 租户数据仓库配置
type TenantRepositoryConfig struct {
	TableName       string `json:"table_name"`
	UsageTableName  string `json:"usage_table_name"`
	EnableSoftDelete bool   `json:"enable_soft_delete"`
	QueryTimeout    time.Duration `json:"query_timeout"`
}

// NewTenantRepository 创建租户数据仓库
func NewTenantRepository(
	db *sqlx.DB,
	config TenantRepositoryConfig,
	logger *zap.Logger,
) *TenantRepository {
	// 设置默认?
	if config.TableName == "" {
		config.TableName = "tenants"
	}
	if config.UsageTableName == "" {
		config.UsageTableName = "tenant_usage"
	}
	if config.QueryTimeout == 0 {
		config.QueryTimeout = 30 * time.Second
	}

	return &TenantRepository{
		db:     db,
		config: config,
		logger: logger,
	}
}

// Create 创建租户
func (r *TenantRepository) Create(ctx context.Context, tenant *multitenant.Tenant) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 序列化JSON字段
	settingsJSON, err := json.Marshal(tenant.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	limitsJSON, err := json.Marshal(tenant.Limits)
	if err != nil {
		return fmt.Errorf("failed to marshal limits: %w", err)
	}

	metadataJSON, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, name, domain, plan, status, owner_id, 
			settings, limits, metadata, 
			created_at, updated_at
		) VALUES (
			:id, :name, :domain, :plan, :status, :owner_id,
			:settings, :limits, :metadata,
			:created_at, :updated_at
		)
	`, r.config.TableName)

	_, err = r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"domain":     tenant.Domain,
		"plan":       tenant.Plan,
		"status":     tenant.Status,
		"owner_id":   tenant.OwnerID,
		"settings":   settingsJSON,
		"limits":     limitsJSON,
		"metadata":   metadataJSON,
		"created_at": tenant.CreatedAt,
		"updated_at": tenant.UpdatedAt,
	})

	if err != nil {
		r.logger.Error("Failed to create tenant",
			zap.String("tenant_id", tenant.ID),
			zap.Error(err))
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	r.logger.Info("Tenant created",
		zap.String("tenant_id", tenant.ID),
		zap.String("name", tenant.Name))

	return nil
}

// GetByID 根据ID获取租户
func (r *TenantRepository) GetByID(ctx context.Context, id string) (*multitenant.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT id, name, domain, plan, status, owner_id,
			   settings, limits, metadata,
			   created_at, updated_at, deleted_at
		FROM %s 
		WHERE id = $1
	`, r.config.TableName)

	if r.config.EnableSoftDelete {
		query += " AND deleted_at IS NULL"
	}

	var row struct {
		ID        string         `db:"id"`
		Name      string         `db:"name"`
		Domain    string         `db:"domain"`
		Plan      string         `db:"plan"`
		Status    string         `db:"status"`
		OwnerID   string         `db:"owner_id"`
		Settings  []byte         `db:"settings"`
		Limits    []byte         `db:"limits"`
		Metadata  []byte         `db:"metadata"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
		DeletedAt sql.NullTime   `db:"deleted_at"`
	}

	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &multitenant.TenantError{
				Code:    "TENANT_NOT_FOUND",
				Message: "Tenant not found",
				Details: map[string]interface{}{"tenant_id": id},
			}
		}
		r.logger.Error("Failed to get tenant by ID",
			zap.String("tenant_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return r.rowToTenant(&row)
}

// GetByDomain 根据域名获取租户
func (r *TenantRepository) GetByDomain(ctx context.Context, domain string) (*multitenant.Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT id, name, domain, plan, status, owner_id,
			   settings, limits, metadata,
			   created_at, updated_at, deleted_at
		FROM %s 
		WHERE domain = $1
	`, r.config.TableName)

	if r.config.EnableSoftDelete {
		query += " AND deleted_at IS NULL"
	}

	var row struct {
		ID        string         `db:"id"`
		Name      string         `db:"name"`
		Domain    string         `db:"domain"`
		Plan      string         `db:"plan"`
		Status    string         `db:"status"`
		OwnerID   string         `db:"owner_id"`
		Settings  []byte         `db:"settings"`
		Limits    []byte         `db:"limits"`
		Metadata  []byte         `db:"metadata"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
		DeletedAt sql.NullTime   `db:"deleted_at"`
	}

	err := r.db.GetContext(ctx, &row, query, domain)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &multitenant.TenantError{
				Code:    "TENANT_NOT_FOUND",
				Message: "Tenant not found",
				Details: map[string]interface{}{"domain": domain},
			}
		}
		r.logger.Error("Failed to get tenant by domain",
			zap.String("domain", domain),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return r.rowToTenant(&row)
}

// Update 更新租户
func (r *TenantRepository) Update(ctx context.Context, tenant *multitenant.Tenant) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 序列化JSON字段
	settingsJSON, err := json.Marshal(tenant.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	limitsJSON, err := json.Marshal(tenant.Limits)
	if err != nil {
		return fmt.Errorf("failed to marshal limits: %w", err)
	}

	metadataJSON, err := json.Marshal(tenant.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE %s SET
			name = :name,
			domain = :domain,
			plan = :plan,
			status = :status,
			owner_id = :owner_id,
			settings = :settings,
			limits = :limits,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id
	`, r.config.TableName)

	if r.config.EnableSoftDelete {
		query += " AND deleted_at IS NULL"
	}

	tenant.UpdatedAt = time.Now()

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"domain":     tenant.Domain,
		"plan":       tenant.Plan,
		"status":     tenant.Status,
		"owner_id":   tenant.OwnerID,
		"settings":   settingsJSON,
		"limits":     limitsJSON,
		"metadata":   metadataJSON,
		"updated_at": tenant.UpdatedAt,
	})

	if err != nil {
		r.logger.Error("Failed to update tenant",
			zap.String("tenant_id", tenant.ID),
			zap.Error(err))
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &multitenant.TenantError{
			Code:    "TENANT_NOT_FOUND",
			Message: "Tenant not found or already deleted",
			Details: map[string]interface{}{"tenant_id": tenant.ID},
		}
	}

	r.logger.Info("Tenant updated",
		zap.String("tenant_id", tenant.ID),
		zap.String("name", tenant.Name))

	return nil
}

// Delete 删除租户
func (r *TenantRepository) Delete(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	var query string
	var args []interface{}

	if r.config.EnableSoftDelete {
		// 软删?
		query = fmt.Sprintf(`
			UPDATE %s SET 
				deleted_at = $1,
				updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL
		`, r.config.TableName)
		args = []interface{}{time.Now(), id}
	} else {
		// 硬删?
		query = fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, r.config.TableName)
		args = []interface{}{id}
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to delete tenant",
			zap.String("tenant_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &multitenant.TenantError{
			Code:    "TENANT_NOT_FOUND",
			Message: "Tenant not found or already deleted",
			Details: map[string]interface{}{"tenant_id": id},
		}
	}

	r.logger.Info("Tenant deleted",
		zap.String("tenant_id", id),
		zap.Bool("soft_delete", r.config.EnableSoftDelete))

	return nil
}

// List 列出租户
func (r *TenantRepository) List(ctx context.Context, filter multitenant.TenantFilter, pagination multitenant.PaginationRequest) (*multitenant.ListTenantsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 构建查询条件
	whereClause, args := r.buildWhereClause(filter)
	
	// 构建排序
	orderClause := r.buildOrderClause(filter.SortBy, filter.SortOrder)
	
	// 计算偏移?
	offset := (pagination.Page - 1) * pagination.PageSize

	// 查询总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s %s
	`, r.config.TableName, whereClause)

	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to count tenants: %w", err)
	}

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT id, name, domain, plan, status, owner_id,
			   settings, limits, metadata,
			   created_at, updated_at, deleted_at
		FROM %s %s %s
		LIMIT $%d OFFSET $%d
	`, r.config.TableName, whereClause, orderClause, len(args)+1, len(args)+2)

	args = append(args, pagination.PageSize, offset)

	var rows []struct {
		ID        string         `db:"id"`
		Name      string         `db:"name"`
		Domain    string         `db:"domain"`
		Plan      string         `db:"plan"`
		Status    string         `db:"status"`
		OwnerID   string         `db:"owner_id"`
		Settings  []byte         `db:"settings"`
		Limits    []byte         `db:"limits"`
		Metadata  []byte         `db:"metadata"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
		DeletedAt sql.NullTime   `db:"deleted_at"`
	}

	err = r.db.SelectContext(ctx, &rows, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// 转换为租户对?
	tenants := make([]*multitenant.Tenant, len(rows))
	for i, row := range rows {
		tenant, err := r.rowToTenant(&row)
		if err != nil {
			r.logger.Error("Failed to convert row to tenant",
				zap.String("tenant_id", row.ID),
				zap.Error(err))
			continue
		}
		tenants[i] = tenant
	}

	// 计算分页信息
	totalPages := (total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)
	hasNext := pagination.Page < int(totalPages)
	hasPrev := pagination.Page > 1

	return &multitenant.ListTenantsResponse{
		Tenants: tenants,
		Pagination: multitenant.PaginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			Total:      total,
			TotalPages: int(totalPages),
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}, nil
}

// RecordUsage 记录使用情况
func (r *TenantRepository) RecordUsage(ctx context.Context, tenantID string, usage *multitenant.TenantUsage) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	// 序列化详细信?
	detailsJSON, err := json.Marshal(usage.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal usage details: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			tenant_id, metric_type, value, unit, details, recorded_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`, r.config.UsageTableName)

	_, err = r.db.ExecContext(ctx, query,
		tenantID,
		usage.MetricType,
		usage.Value,
		usage.Unit,
		detailsJSON,
		usage.RecordedAt,
	)

	if err != nil {
		r.logger.Error("Failed to record usage",
			zap.String("tenant_id", tenantID),
			zap.String("metric_type", usage.MetricType),
			zap.Error(err))
		return fmt.Errorf("failed to record usage: %w", err)
	}

	return nil
}

// GetUsage 获取使用情况
func (r *TenantRepository) GetUsage(ctx context.Context, tenantID string, timeRange multitenant.TimeRange) ([]*multitenant.TenantUsage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		SELECT metric_type, value, unit, details, recorded_at
		FROM %s
		WHERE tenant_id = $1 
		  AND recorded_at >= $2 
		  AND recorded_at <= $3
		ORDER BY recorded_at DESC
	`, r.config.UsageTableName)

	var rows []struct {
		MetricType string    `db:"metric_type"`
		Value      float64   `db:"value"`
		Unit       string    `db:"unit"`
		Details    []byte    `db:"details"`
		RecordedAt time.Time `db:"recorded_at"`
	}

	err := r.db.SelectContext(ctx, &rows, query, tenantID, timeRange.Start, timeRange.End)
	if err != nil {
		r.logger.Error("Failed to get usage",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	usage := make([]*multitenant.TenantUsage, len(rows))
	for i, row := range rows {
		var details map[string]interface{}
		if len(row.Details) > 0 {
			if err := json.Unmarshal(row.Details, &details); err != nil {
				r.logger.Warn("Failed to unmarshal usage details",
					zap.String("tenant_id", tenantID),
					zap.Error(err))
			}
		}

		usage[i] = &multitenant.TenantUsage{
			MetricType: row.MetricType,
			Value:      row.Value,
			Unit:       row.Unit,
			Details:    details,
			RecordedAt: row.RecordedAt,
		}
	}

	return usage, nil
}

// 辅助方法

// rowToTenant 将数据库行转换为租户对象
func (r *TenantRepository) rowToTenant(row interface{}) (*multitenant.Tenant, error) {
	var (
		id, name, domain, plan, status, ownerID string
		settings, limits, metadata              []byte
		createdAt, updatedAt                    time.Time
		deletedAt                               sql.NullTime
	)

	// 使用类型断言获取字段?
	switch r := row.(type) {
	case *struct {
		ID        string         `db:"id"`
		Name      string         `db:"name"`
		Domain    string         `db:"domain"`
		Plan      string         `db:"plan"`
		Status    string         `db:"status"`
		OwnerID   string         `db:"owner_id"`
		Settings  []byte         `db:"settings"`
		Limits    []byte         `db:"limits"`
		Metadata  []byte         `db:"metadata"`
		CreatedAt time.Time      `db:"created_at"`
		UpdatedAt time.Time      `db:"updated_at"`
		DeletedAt sql.NullTime   `db:"deleted_at"`
	}:
		id = r.ID
		name = r.Name
		domain = r.Domain
		plan = r.Plan
		status = r.Status
		ownerID = r.OwnerID
		settings = r.Settings
		limits = r.Limits
		metadata = r.Metadata
		createdAt = r.CreatedAt
		updatedAt = r.UpdatedAt
		deletedAt = r.DeletedAt
	default:
		return nil, fmt.Errorf("unsupported row type")
	}

	// 反序列化JSON字段
	var tenantSettings multitenant.TenantSettings
	if len(settings) > 0 {
		if err := json.Unmarshal(settings, &tenantSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}
	}

	var tenantLimits multitenant.TenantLimits
	if len(limits) > 0 {
		if err := json.Unmarshal(limits, &tenantLimits); err != nil {
			return nil, fmt.Errorf("failed to unmarshal limits: %w", err)
		}
	}

	var tenantMetadata map[string]interface{}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &tenantMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	tenant := &multitenant.Tenant{
		ID:        id,
		Name:      name,
		Domain:    domain,
		Plan:      multitenant.TenantPlan(plan),
		Status:    multitenant.TenantStatus(status),
		OwnerID:   ownerID,
		Settings:  tenantSettings,
		Limits:    tenantLimits,
		Metadata:  tenantMetadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if deletedAt.Valid {
		tenant.DeletedAt = &deletedAt.Time
	}

	return tenant, nil
}

// buildWhereClause 构建WHERE子句
func (r *TenantRepository) buildWhereClause(filter multitenant.TenantFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 软删除过?
	if r.config.EnableSoftDelete {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	// 状态过?
	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	// 计划过滤
	if len(filter.Plans) > 0 {
		placeholders := make([]string, len(filter.Plans))
		for i, plan := range filter.Plans {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, plan)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("plan IN (%s)", strings.Join(placeholders, ",")))
	}

	// 所有者过?
	if filter.OwnerID != "" {
		conditions = append(conditions, fmt.Sprintf("owner_id = $%d", argIndex))
		args = append(args, filter.OwnerID)
		argIndex++
	}

	// 搜索过滤
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR domain ILIKE $%d)", argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// 时间范围过滤
	if filter.CreatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// buildOrderClause 构建ORDER BY子句
func (r *TenantRepository) buildOrderClause(sortBy, sortOrder string) string {
	if sortBy == "" {
		sortBy = "created_at"
	}

	if sortOrder == "" {
		sortOrder = "DESC"
	}

	// 验证排序字段
	validSortFields := map[string]bool{
		"id":         true,
		"name":       true,
		"domain":     true,
		"plan":       true,
		"status":     true,
		"owner_id":   true,
		"created_at": true,
		"updated_at": true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	// 验证排序顺序
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", sortBy, sortOrder)
}

