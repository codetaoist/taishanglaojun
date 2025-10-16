package repository

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/permission"
)

// PermissionRepositoryImpl 权限仓储实现
type PermissionRepositoryImpl struct {
	db     *sqlx.DB
	logger *zap.Logger
	config PermissionRepositoryConfig
}

// PermissionRepositoryConfig 权限仓储配置
type PermissionRepositoryConfig struct {
	// 数据库配?
	TablePrefix      string `json:"table_prefix"`
	EnableSharding   bool   `json:"enable_sharding"`
	ShardCount       int    `json:"shard_count"`
	
	// 性能配置
	BatchSize        int           `json:"batch_size"`
	QueryTimeout     time.Duration `json:"query_timeout"`
	MaxConnections   int           `json:"max_connections"`
	
	// 索引配置
	EnableIndexing   bool `json:"enable_indexing"`
	IndexPrefix      string `json:"index_prefix"`
	
	// 审计配置
	EnableAuditLog   bool `json:"enable_audit_log"`
	AuditTableName   string `json:"audit_table_name"`
}

// JSONField JSON字段类型
type JSONField map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONField", value)
	}

	return json.Unmarshal(bytes, j)
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *sqlx.DB, logger *zap.Logger, config PermissionRepositoryConfig) permission.PermissionRepository {
	// 设置默认配置
	if config.TablePrefix == "" {
		config.TablePrefix = "perm_"
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.QueryTimeout == 0 {
		config.QueryTimeout = 30 * time.Second
	}
	if config.AuditTableName == "" {
		config.AuditTableName = config.TablePrefix + "audit_logs"
	}

	return &PermissionRepositoryImpl{
		db:     db,
		logger: logger,
		config: config,
	}
}

// CreateRole 创建角色
func (r *PermissionRepositoryImpl) CreateRole(ctx context.Context, role *permission.Role) error {
	query := fmt.Sprintf(`
		INSERT INTO %sroles (
			id, name, code, description, type, level, parent_id, 
			is_system, is_active, metadata, tenant_id, created_at, updated_at
		) VALUES (
			:id, :name, :code, :description, :type, :level, :parent_id,
			:is_system, :is_active, :metadata, :tenant_id, :created_at, :updated_at
		)`, r.config.TablePrefix)

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"code":        role.Code,
		"description": role.Description,
		"type":        role.Type,
		"level":       role.Level,
		"parent_id":   role.ParentID,
		"is_system":   role.IsSystem,
		"is_active":   role.IsActive,
		"metadata":    JSONField(role.Metadata),
		"tenant_id":   role.TenantID,
		"created_at":  role.CreatedAt,
		"updated_at":  role.UpdatedAt,
	})

	if err != nil {
		r.logger.Error("Failed to create role", zap.Error(err))
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// CreateResourcePermission 创建资源权限
func (r *PermissionRepositoryImpl) CreateResourcePermission(ctx context.Context, resourcePermission *permission.ResourcePermission) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		INSERT INTO resource_permissions (id, resource_id, resource_type, subject_id, subject_type, permission_id, effect, conditions, expires_at, tenant_id, created_at, created_by)
		VALUES (:id, :resource_id, :resource_type, :subject_id, :subject_type, :permission_id, :effect, :conditions, :expires_at, :tenant_id, :created_at, :created_by)
	`

	_, err := r.db.NamedExecContext(ctx, query, resourcePermission)
	if err != nil {
		r.logger.Error("Failed to create resource permission", zap.Error(err), zap.String("resource_permission_id", resourcePermission.ID))
		return fmt.Errorf("failed to create resource permission: %w", err)
	}

	return nil
}

// DeleteResourcePermission 删除资源权限
func (r *PermissionRepositoryImpl) DeleteResourcePermission(ctx context.Context, resourceID, resourceType, subjectID string, subjectType permission.SubjectType, permissionID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		DELETE FROM resource_permissions 
		WHERE resource_id = $1 AND resource_type = $2 AND subject_id = $3 AND subject_type = $4 AND permission_id = $5
	`

	result, err := r.db.ExecContext(ctx, query, resourceID, resourceType, subjectID, subjectType, permissionID)
	if err != nil {
		r.logger.Error("Failed to delete resource permission", zap.Error(err), 
			zap.String("resource_id", resourceID), 
			zap.String("subject_id", subjectID))
		return fmt.Errorf("failed to delete resource permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("resource permission not found")
	}

	return nil
}

// GetResourcePermissions 获取资源权限列表
func (r *PermissionRepositoryImpl) GetResourcePermissions(ctx context.Context, resourceID, resourceType string) ([]*permission.ResourcePermission, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		SELECT rp.id, rp.resource_id, rp.resource_type, rp.subject_id, rp.subject_type, 
		       rp.permission_id, rp.effect, rp.conditions, rp.expires_at, rp.tenant_id, 
		       rp.created_at, rp.created_by,
		       p.id as "permission.id", p.name as "permission.name", p.code as "permission.code",
		       p.description as "permission.description", p.category as "permission.category",
		       p.resource as "permission.resource", p.action as "permission.action",
		       p.effect as "permission.effect", p.conditions as "permission.conditions"
		FROM resource_permissions rp
		LEFT JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.resource_id = $1 AND rp.resource_type = $2
		ORDER BY rp.created_at DESC
	`

	rows, err := r.db.QueryxContext(ctx, query, resourceID, resourceType)
	if err != nil {
		r.logger.Error("Failed to get resource permissions", zap.Error(err), 
			zap.String("resource_id", resourceID), 
			zap.String("resource_type", resourceType))
		return nil, fmt.Errorf("failed to get resource permissions: %w", err)
	}
	defer rows.Close()

	var resourcePermissions []*permission.ResourcePermission
	for rows.Next() {
		var rp permission.ResourcePermission
		var perm permission.Permission
		
		err := rows.Scan(
			&rp.ID, &rp.ResourceID, &rp.ResourceType, &rp.SubjectID, &rp.SubjectType,
			&rp.PermissionID, &rp.Effect, &rp.Conditions, &rp.ExpiresAt, &rp.TenantID,
			&rp.CreatedAt, &rp.CreatedBy,
			&perm.ID, &perm.Name, &perm.Code, &perm.Description, &perm.Category,
			&perm.Resource, &perm.Action, &perm.Effect, &perm.Conditions,
		)
		if err != nil {
			r.logger.Error("Failed to scan resource permission", zap.Error(err))
			return nil, fmt.Errorf("failed to scan resource permission: %w", err)
		}

		if perm.ID != "" {
			rp.Permission = &perm
		}
		
		resourcePermissions = append(resourcePermissions, &rp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate resource permissions: %w", err)
	}

	return resourcePermissions, nil
}

// CreatePolicy 创建策略
func (r *PermissionRepositoryImpl) CreatePolicy(ctx context.Context, policy *permission.Policy) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		INSERT INTO policies (id, name, description, type, rules, effect, priority, is_active, metadata, tenant_id, created_at, updated_at, created_by, updated_by)
		VALUES (:id, :name, :description, :type, :rules, :effect, :priority, :is_active, :metadata, :tenant_id, :created_at, :updated_at, :created_by, :updated_by)
	`

	_, err := r.db.NamedExecContext(ctx, query, policy)
	if err != nil {
		r.logger.Error("Failed to create policy", zap.Error(err), zap.String("policy_id", policy.ID))
		return fmt.Errorf("failed to create policy: %w", err)
	}

	return nil
}

// GetPolicy 获取策略
func (r *PermissionRepositoryImpl) GetPolicy(ctx context.Context, policyID string) (*permission.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		SELECT id, name, description, type, rules, effect, priority, is_active, metadata, tenant_id, created_at, updated_at, created_by, updated_by
		FROM policies
		WHERE id = $1
	`

	var policy permission.Policy
	err := r.db.GetContext(ctx, &policy, query, policyID)
	if err != nil {
		r.logger.Error("Failed to get policy", zap.Error(err), zap.String("policy_id", policyID))
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	return &policy, nil
}

// UpdatePolicy 更新策略
func (r *PermissionRepositoryImpl) UpdatePolicy(ctx context.Context, policy *permission.Policy) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		UPDATE policies
		SET name = :name, description = :description, type = :type, rules = :rules, 
		    effect = :effect, priority = :priority, is_active = :is_active, 
		    metadata = :metadata, updated_at = :updated_at, updated_by = :updated_by
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, policy)
	if err != nil {
		r.logger.Error("Failed to update policy", zap.Error(err), zap.String("policy_id", policy.ID))
		return fmt.Errorf("failed to update policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

// DeletePolicy 删除策略
func (r *PermissionRepositoryImpl) DeletePolicy(ctx context.Context, policyID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `DELETE FROM policies WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, policyID)
	if err != nil {
		r.logger.Error("Failed to delete policy", zap.Error(err), zap.String("policy_id", policyID))
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("policy not found")
	}

	return nil
}

// ListPolicies 获取策略列表
func (r *PermissionRepositoryImpl) ListPolicies(ctx context.Context, filter *permission.PolicyFilter) ([]*permission.Policy, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	whereClause, args := r.buildPolicyWhereClause(filter)
	
	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM policies %s", whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count policies", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count policies: %w", err)
	}

	// 获取数据
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, name, description, type, rules, effect, priority, is_active, metadata, tenant_id, created_at, updated_at, created_by, updated_by
		FROM policies %s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, filter.Pagination.PageSize, offset)

	var policies []*permission.Policy
	err = r.db.SelectContext(ctx, &policies, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to list policies", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list policies: %w", err)
	}

	return policies, total, nil
}

// buildPolicyWhereClause 构建策略查询条件
func (r *PermissionRepositoryImpl) buildPolicyWhereClause(filter *permission.PolicyFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Name+"%")
		argIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Effect != nil {
		conditions = append(conditions, fmt.Sprintf("effect = $%d", argIndex))
		args = append(args, *filter.Effect)
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// GetRole 获取角色
func (r *PermissionRepositoryImpl) GetRole(ctx context.Context, roleID string) (*permission.Role, error) {
	query := fmt.Sprintf(`
		SELECT id, name, code, description, type, level, parent_id,
			   is_system, is_active, metadata, tenant_id, created_at, updated_at
		FROM %sroles 
		WHERE id = $1 AND deleted_at IS NULL`, r.config.TablePrefix)

	var role permission.Role
	var metadata JSONField

	err := r.db.QueryRowxContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Code, &role.Description,
		&role.Type, &role.Level, &role.ParentID,
		&role.IsSystem, &role.IsActive, &metadata,
		&role.TenantID, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("role not found")
		}
		r.logger.Error("Failed to get role", zap.Error(err))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	role.Metadata = map[string]interface{}(metadata)
	return &role, nil
}

// GetRoleByName 根据名称获取角色
func (r *PermissionRepositoryImpl) GetRoleByName(ctx context.Context, name string, tenantID string) (*permission.Role, error) {
	query := fmt.Sprintf(`
		SELECT id, name, code, description, type, level, parent_id,
			   is_system, is_active, metadata, tenant_id, created_at, updated_at
		FROM %sroles 
		WHERE name = $1 AND tenant_id = $2 AND deleted_at IS NULL`, r.config.TablePrefix)

	var role permission.Role
	var metadata JSONField

	err := r.db.QueryRowxContext(ctx, query, name, tenantID).Scan(
		&role.ID, &role.Name, &role.Code, &role.Description,
		&role.Type, &role.Level, &role.ParentID,
		&role.IsSystem, &role.IsActive, &metadata,
		&role.TenantID, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("role not found")
		}
		r.logger.Error("Failed to get role by name", zap.Error(err))
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	role.Metadata = map[string]interface{}(metadata)
	return &role, nil
}

// UpdateRole 更新角色
func (r *PermissionRepositoryImpl) UpdateRole(ctx context.Context, role *permission.Role) error {
	query := fmt.Sprintf(`
		UPDATE %sroles SET 
			name = :name, code = :code, description = :description,
			type = :type, level = :level, parent_id = :parent_id,
			is_active = :is_active, metadata = :metadata, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`, r.config.TablePrefix)

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"code":        role.Code,
		"description": role.Description,
		"type":        role.Type,
		"level":       role.Level,
		"parent_id":   role.ParentID,
		"is_active":   role.IsActive,
		"metadata":    JSONField(role.Metadata),
		"updated_at":  role.UpdatedAt,
	})

	if err != nil {
		r.logger.Error("Failed to update role", zap.Error(err))
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found or already deleted")
	}

	return nil
}

// DeleteRole 删除角色
func (r *PermissionRepositoryImpl) DeleteRole(ctx context.Context, roleID string) error {
	query := fmt.Sprintf(`
		UPDATE %sroles SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, r.config.TablePrefix)

	result, err := r.db.ExecContext(ctx, query, time.Now(), roleID)
	if err != nil {
		r.logger.Error("Failed to delete role", zap.Error(err))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role not found or already deleted")
	}

	return nil
}

// ListRoles 列出角色
func (r *PermissionRepositoryImpl) ListRoles(ctx context.Context, filter *permission.RoleFilter) ([]*permission.Role, int64, error) {
	// 构建查询条件
	whereClause, args := r.buildRoleWhereClause(filter)
	
	// 计算总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %sroles WHERE %s`, r.config.TablePrefix, whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count roles", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// 查询数据
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, name, code, description, type, level, parent_id,
			   is_system, is_active, metadata, tenant_id, created_at, updated_at
		FROM %sroles 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, r.config.TablePrefix, whereClause, len(args)+1, len(args)+2)

	args = append(args, filter.Pagination.PageSize, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to query roles", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []*permission.Role
	for rows.Next() {
		var role permission.Role
		var metadata JSONField

		err := rows.Scan(
			&role.ID, &role.Name, &role.Code, &role.Description,
			&role.Type, &role.Level, &role.ParentID,
			&role.IsSystem, &role.IsActive, &metadata,
			&role.TenantID, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan role", zap.Error(err))
			continue
		}

		role.Metadata = map[string]interface{}(metadata)
		roles = append(roles, &role)
	}

	return roles, total, nil
}

// CreatePermission 创建权限
func (r *PermissionRepositoryImpl) CreatePermission(ctx context.Context, perm *permission.Permission) error {
	query := fmt.Sprintf(`
		INSERT INTO %spermissions (
			id, name, code, description, category, resource, action,
			effect, conditions, metadata, tenant_id, created_at, updated_at
		) VALUES (
			:id, :name, :code, :description, :category, :resource, :action,
			:effect, :conditions, :metadata, :tenant_id, :created_at, :updated_at
		)`, r.config.TablePrefix)

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          perm.ID,
		"name":        perm.Name,
		"code":        perm.Code,
		"description": perm.Description,
		"category":    perm.Category,
		"resource":    perm.Resource,
		"action":      perm.Action,
		"effect":      perm.Effect,
		"conditions":  JSONField(perm.Conditions),
		"metadata":    JSONField(perm.Metadata),
		"tenant_id":   perm.TenantID,
		"created_at":  perm.CreatedAt,
		"updated_at":  perm.UpdatedAt,
	})

	if err != nil {
		r.logger.Error("Failed to create permission", zap.Error(err))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetPermission 获取权限
func (r *PermissionRepositoryImpl) GetPermission(ctx context.Context, permissionID string) (*permission.Permission, error) {
	query := fmt.Sprintf(`
		SELECT id, name, code, description, category, resource, action,
			   effect, conditions, metadata, tenant_id, created_at, updated_at
		FROM %spermissions 
		WHERE id = $1 AND deleted_at IS NULL`, r.config.TablePrefix)

	var perm permission.Permission
	var conditions, metadata JSONField

	err := r.db.QueryRowxContext(ctx, query, permissionID).Scan(
		&perm.ID, &perm.Name, &perm.Code, &perm.Description,
		&perm.Category, &perm.Resource, &perm.Action,
		&perm.Effect, &conditions, &metadata,
		&perm.TenantID, &perm.CreatedAt, &perm.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("permission not found")
		}
		r.logger.Error("Failed to get permission", zap.Error(err))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	perm.Conditions = map[string]interface{}(conditions)
	perm.Metadata = map[string]interface{}(metadata)
	return &perm, nil
}

// UpdatePermission 更新权限
func (r *PermissionRepositoryImpl) UpdatePermission(ctx context.Context, perm *permission.Permission) error {
	query := fmt.Sprintf(`
		UPDATE %spermissions SET 
			name = :name, code = :code, description = :description,
			category = :category, effect = :effect, conditions = :conditions,
			metadata = :metadata, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`, r.config.TablePrefix)

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          perm.ID,
		"name":        perm.Name,
		"code":        perm.Code,
		"description": perm.Description,
		"category":    perm.Category,
		"effect":      perm.Effect,
		"conditions":  JSONField(perm.Conditions),
		"metadata":    JSONField(perm.Metadata),
		"updated_at":  perm.UpdatedAt,
	})

	if err != nil {
		r.logger.Error("Failed to update permission", zap.Error(err))
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found or already deleted")
	}

	return nil
}

// DeletePermission 删除权限
func (r *PermissionRepositoryImpl) DeletePermission(ctx context.Context, permissionID string) error {
	query := fmt.Sprintf(`
		UPDATE %spermissions SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, r.config.TablePrefix)

	result, err := r.db.ExecContext(ctx, query, time.Now(), permissionID)
	if err != nil {
		r.logger.Error("Failed to delete permission", zap.Error(err))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not found or already deleted")
	}

	return nil
}

// ListPermissions 列出权限
func (r *PermissionRepositoryImpl) ListPermissions(ctx context.Context, filter *permission.PermissionFilter) ([]*permission.Permission, int64, error) {
	// 构建查询条件
	whereClause, args := r.buildPermissionWhereClause(filter)
	
	// 计算总数
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %spermissions WHERE %s`, r.config.TablePrefix, whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count permissions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// 查询数据
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, name, code, description, category, resource, action,
			   effect, conditions, metadata, tenant_id, created_at, updated_at
		FROM %spermissions 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, r.config.TablePrefix, whereClause, len(args)+1, len(args)+2)

	args = append(args, filter.Pagination.PageSize, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to query permissions", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*permission.Permission
	for rows.Next() {
		var perm permission.Permission
		var conditions, metadata JSONField

		err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Code, &perm.Description,
			&perm.Category, &perm.Resource, &perm.Action,
			&perm.Effect, &conditions, &metadata,
			&perm.TenantID, &perm.CreatedAt, &perm.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan permission", zap.Error(err))
			continue
		}

		perm.Conditions = map[string]interface{}(conditions)
		perm.Metadata = map[string]interface{}(metadata)
		permissions = append(permissions, &perm)
	}

	return permissions, total, nil
}

// AssignPermissionToRole 分配权限给角?
func (r *PermissionRepositoryImpl) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	query := fmt.Sprintf(`
		INSERT INTO %srole_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (role_id, permission_id) DO NOTHING`, r.config.TablePrefix)

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID, time.Now())
	if err != nil {
		r.logger.Error("Failed to assign permission to role", zap.Error(err))
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RevokePermissionFromRole 从角色撤销权限
func (r *PermissionRepositoryImpl) RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	query := fmt.Sprintf(`
		DELETE FROM %srole_permissions 
		WHERE role_id = $1 AND permission_id = $2`, r.config.TablePrefix)

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		r.logger.Error("Failed to revoke permission from role", zap.Error(err))
		return fmt.Errorf("failed to revoke permission from role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission assignment not found")
	}

	return nil
}

// GetRolePermissions 获取角色权限
func (r *PermissionRepositoryImpl) GetRolePermissions(ctx context.Context, roleID string) ([]*permission.Permission, error) {
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.code, p.description, p.category, p.resource, p.action,
			   p.effect, p.conditions, p.metadata, p.tenant_id, p.created_at, p.updated_at
		FROM %spermissions p
		INNER JOIN %srole_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1 AND p.deleted_at IS NULL`, r.config.TablePrefix, r.config.TablePrefix)

	rows, err := r.db.QueryxContext(ctx, query, roleID)
	if err != nil {
		r.logger.Error("Failed to get role permissions", zap.Error(err))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*permission.Permission
	for rows.Next() {
		var perm permission.Permission
		var conditions, metadata JSONField

		err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Code, &perm.Description,
			&perm.Category, &perm.Resource, &perm.Action,
			&perm.Effect, &conditions, &metadata,
			&perm.TenantID, &perm.CreatedAt, &perm.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan permission", zap.Error(err))
			continue
		}

		perm.Conditions = map[string]interface{}(conditions)
		perm.Metadata = map[string]interface{}(metadata)
		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

// AssignRoleToUser 分配角色给用?
func (r *PermissionRepositoryImpl) AssignRoleToUser(ctx context.Context, userID, roleID string, tenantID string) error {
	query := fmt.Sprintf(`
		INSERT INTO %suser_roles (user_id, role_id, tenant_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, role_id, tenant_id) DO NOTHING`, r.config.TablePrefix)

	_, err := r.db.ExecContext(ctx, query, userID, roleID, tenantID, time.Now())
	if err != nil {
		r.logger.Error("Failed to assign role to user", zap.Error(err))
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

// RevokeRoleFromUser 从用户撤销角色
func (r *PermissionRepositoryImpl) RevokeRoleFromUser(ctx context.Context, userID, roleID string, tenantID string) error {
	query := fmt.Sprintf(`
		DELETE FROM %suser_roles 
		WHERE user_id = $1 AND role_id = $2 AND tenant_id = $3`, r.config.TablePrefix)

	result, err := r.db.ExecContext(ctx, query, userID, roleID, tenantID)
	if err != nil {
		r.logger.Error("Failed to revoke role from user", zap.Error(err))
		return fmt.Errorf("failed to revoke role from user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("role assignment not found")
	}

	return nil
}

// GetUserRoles 获取用户角色
func (r *PermissionRepositoryImpl) GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*permission.Role, error) {
	query := fmt.Sprintf(`
		SELECT r.id, r.name, r.code, r.description, r.type, r.level, r.parent_id,
			   r.is_system, r.is_active, r.metadata, r.tenant_id, r.created_at, r.updated_at
		FROM %sroles r
		INNER JOIN %suser_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND ur.tenant_id = $2 AND r.deleted_at IS NULL AND r.is_active = true`, 
		r.config.TablePrefix, r.config.TablePrefix)

	rows, err := r.db.QueryxContext(ctx, query, userID, tenantID)
	if err != nil {
		r.logger.Error("Failed to get user roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*permission.Role
	for rows.Next() {
		var role permission.Role
		var metadata JSONField

		err := rows.Scan(
			&role.ID, &role.Name, &role.Code, &role.Description,
			&role.Type, &role.Level, &role.ParentID,
			&role.IsSystem, &role.IsActive, &metadata,
			&role.TenantID, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan role", zap.Error(err))
			continue
		}

		role.Metadata = map[string]interface{}(metadata)
		roles = append(roles, &role)
	}

	return roles, nil
}

// 构建角色查询条件
func (r *PermissionRepositoryImpl) buildRoleWhereClause(filter *permission.RoleFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 基本条件
	conditions = append(conditions, "deleted_at IS NULL")

	// 租户ID
	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

	// 角色类型
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	// 是否激?
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	// 是否系统角色
	if filter.IsSystem != nil {
		conditions = append(conditions, fmt.Sprintf("is_system = $%d", argIndex))
		args = append(args, *filter.IsSystem)
		argIndex++
	}

	// 父角色ID
	if filter.ParentID != nil {
		if *filter.ParentID == "" {
			conditions = append(conditions, "parent_id IS NULL")
		} else {
			conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
			args = append(args, *filter.ParentID)
			argIndex++
		}
	}

	// 搜索关键?
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}

// 构建权限查询条件
func (r *PermissionRepositoryImpl) buildPermissionWhereClause(filter *permission.PermissionFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// 基本条件
	conditions = append(conditions, "deleted_at IS NULL")

	// 租户ID
	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

	// 分类
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, filter.Category)
		argIndex++
	}

	// 资源
	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, filter.Resource)
		argIndex++
	}

	// 动作
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, filter.Action)
		argIndex++
	}

	// 效果
	if filter.Effect != nil {
		conditions = append(conditions, fmt.Sprintf("effect = $%d", argIndex))
		args = append(args, *filter.Effect)
		argIndex++
	}

	// 搜索关键?
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}

// CreatePermissionAuditLog 创建权限审计日志
func (r *PermissionRepositoryImpl) CreatePermissionAuditLog(ctx context.Context, auditLog *permission.PermissionAuditLog) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, user_id, tenant_id, resource, action, resource_id, 
			effect, allowed, reason, context, ip_address, user_agent, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.config.AuditTableName)

	contextJSON, _ := json.Marshal(auditLog.Context)
	
	_, err := r.db.ExecContext(ctx, query,
		auditLog.ID, auditLog.UserID, auditLog.TenantID, auditLog.Resource,
		auditLog.Action, auditLog.ResourceID, auditLog.Effect, auditLog.Allowed,
		auditLog.Reason, contextJSON, auditLog.IPAddress, auditLog.UserAgent,
		auditLog.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create permission audit log", zap.Error(err))
		return fmt.Errorf("failed to create permission audit log: %w", err)
	}

	return nil
}

// GetPermissionAuditLogs 获取权限审计日志
func (r *PermissionRepositoryImpl) GetPermissionAuditLogs(ctx context.Context, filter *permission.PermissionAuditFilter) ([]*permission.PermissionAuditLog, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	whereClause, args := r.buildAuditLogWhereClause(filter)
	
	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.config.AuditTableName, whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		r.logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// 获取数据
	offset := (filter.Pagination.Page - 1) * filter.Pagination.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, tenant_id, resource, action, resource_id, effect, allowed, reason, context, ip_address, user_agent, created_at
		FROM %s %s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, r.config.AuditTableName, whereClause, filter.Pagination.PageSize, offset)

	rows, err := r.db.QueryxContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Error("Failed to get audit logs", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var auditLogs []*permission.PermissionAuditLog
	for rows.Next() {
		var auditLog permission.PermissionAuditLog
		var contextJSON []byte
		
		err := rows.Scan(
			&auditLog.ID, &auditLog.UserID, &auditLog.TenantID, &auditLog.Resource,
			&auditLog.Action, &auditLog.ResourceID, &auditLog.Effect, &auditLog.Allowed,
			&auditLog.Reason, &contextJSON, &auditLog.IPAddress, &auditLog.UserAgent,
			&auditLog.CreatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan audit log", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// 解析 context JSON
		if len(contextJSON) > 0 {
			err = json.Unmarshal(contextJSON, &auditLog.Context)
			if err != nil {
				r.logger.Warn("Failed to unmarshal audit log context", zap.Error(err))
				auditLog.Context = make(map[string]interface{})
			}
		}
		
		auditLogs = append(auditLogs, &auditLog)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// buildAuditLogWhereClause 构建审计日志查询条件
func (r *PermissionRepositoryImpl) buildAuditLogWhereClause(filter *permission.PermissionAuditFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.TenantID != "" {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filter.TenantID)
		argIndex++
	}

	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, filter.Resource)
		argIndex++
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, filter.Action)
		argIndex++
	}

	if filter.Effect != nil {
		conditions = append(conditions, fmt.Sprintf("effect = $%d", argIndex))
		args = append(args, *filter.Effect)
		argIndex++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.EndTime)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

// HealthCheck 健康检查
func (r *PermissionRepositoryImpl) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()
	
	return r.db.PingContext(ctx)
}

// CreatePermissionInheritance 创建权限继承
func (r *PermissionRepositoryImpl) CreatePermissionInheritance(ctx context.Context, inheritance *permission.PermissionInheritance) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		INSERT INTO permission_inheritances (id, resource_id, resource_type, parent_id, parent_type, inherit_type, is_active, tenant_id, created_at, created_by)
		VALUES (:id, :resource_id, :resource_type, :parent_id, :parent_type, :inherit_type, :is_active, :tenant_id, :created_at, :created_by)
	`

	_, err := r.db.NamedExecContext(ctx, query, inheritance)
	if err != nil {
		r.logger.Error("Failed to create permission inheritance", zap.Error(err), zap.String("inheritance_id", inheritance.ID))
		return fmt.Errorf("failed to create permission inheritance: %w", err)
	}

	return nil
}

// GetPermissionInheritance 获取权限继承
func (r *PermissionRepositoryImpl) GetPermissionInheritance(ctx context.Context, resourceID, resourceType string) (*permission.PermissionInheritance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		SELECT id, resource_id, resource_type, parent_id, parent_type, inherit_type, is_active, tenant_id, created_at, created_by
		FROM permission_inheritances
		WHERE resource_id = $1 AND resource_type = $2 AND is_active = true
	`

	var inheritance permission.PermissionInheritance
	err := r.db.GetContext(ctx, &inheritance, query, resourceID, resourceType)
	if err != nil {
		r.logger.Error("Failed to get permission inheritance", zap.Error(err), zap.String("resource_id", resourceID), zap.String("resource_type", resourceType))
		return nil, fmt.Errorf("failed to get permission inheritance: %w", err)
	}

	return &inheritance, nil
}

// UpdatePermissionInheritance 更新权限继承
func (r *PermissionRepositoryImpl) UpdatePermissionInheritance(ctx context.Context, inheritance *permission.PermissionInheritance) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `
		UPDATE permission_inheritances
		SET parent_id = :parent_id, parent_type = :parent_type, inherit_type = :inherit_type, is_active = :is_active
		WHERE resource_id = :resource_id AND resource_type = :resource_type
	`

	result, err := r.db.NamedExecContext(ctx, query, inheritance)
	if err != nil {
		r.logger.Error("Failed to update permission inheritance", zap.Error(err), zap.String("resource_id", inheritance.ResourceID))
		return fmt.Errorf("failed to update permission inheritance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission inheritance not found")
	}

	return nil
}

// DeletePermissionInheritance 删除权限继承
func (r *PermissionRepositoryImpl) DeletePermissionInheritance(ctx context.Context, resourceID, resourceType string) error {
	ctx, cancel := context.WithTimeout(ctx, r.config.QueryTimeout)
	defer cancel()

	query := `DELETE FROM permission_inheritances WHERE resource_id = $1 AND resource_type = $2`

	result, err := r.db.ExecContext(ctx, query, resourceID, resourceType)
	if err != nil {
		r.logger.Error("Failed to delete permission inheritance", zap.Error(err), zap.String("resource_id", resourceID))
		return fmt.Errorf("failed to delete permission inheritance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission inheritance not found")
	}

	return nil
}

