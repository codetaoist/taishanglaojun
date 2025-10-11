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
	// 数据库配�?
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

// AssignPermissionToRole 分配权限给角�?
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

// AssignRoleToUser 分配角色给用�?
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

	// 是否激�?
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

	// 搜索关键�?
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

	// 搜索关键�?
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}

// HealthCheck 健康检�?
func (r *PermissionRepositoryImpl) HealthCheck(ctx context.Context) error {
	query := "SELECT 1"
	var result int
	err := r.db.GetContext(ctx, &result, query)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
