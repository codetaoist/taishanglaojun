package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"
)

var (
	ErrPermissionNotFound = errors.New("permission not found")
)

// PermissionRepository 权限仓库接口
type PermissionRepository interface {
	Create(ctx context.Context, permission *models.Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	GetByName(ctx context.Context, name string) (*models.Permission, error)
	List(ctx context.Context, opts *ListPermissionsOptions) ([]*models.Permission, int64, error)
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ListPermissionsOptions 权限列表查询选项
type ListPermissionsOptions struct {
	Page     int
	PageSize int
	Resource string
	Action   string
	Search   string
}

// permissionRepository 权限仓库实现
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓库实例
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// Create 创建权限
func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetByID 根据ID获取权限
func (r *permissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// GetByName 根据名称获取权限
func (r *permissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// List 获取权限列表
func (r *permissionRepository) List(ctx context.Context, opts *ListPermissionsOptions) ([]*models.Permission, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Permission{})

	// 应用过滤条件
	if opts.Resource != "" {
		query = query.Where("resource = ?", opts.Resource)
	}
	if opts.Action != "" {
		query = query.Where("action = ?", opts.Action)
	}
	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var permissions []*models.Permission
	offset := (opts.Page - 1) * opts.PageSize
	err := query.Offset(offset).Limit(opts.PageSize).Order("created_at DESC").Find(&permissions).Error
	if err != nil {
		return nil, 0, err
	}

	return permissions, total, nil
}

// Update 更新权限
func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// Delete 删除权限
func (r *permissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Permission{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPermissionNotFound
	}
	return nil
}

// RolePermissionRepository 角色权限仓库接口
type RolePermissionRepository interface {
	Create(ctx context.Context, rolePermission *models.RolePermission) error
	GetPermissionsByRole(ctx context.Context, role models.UserRole) ([]*models.Permission, error)
	Exists(ctx context.Context, role models.UserRole, permissionID uuid.UUID) (bool, error)
	Delete(ctx context.Context, role models.UserRole, permissionID uuid.UUID) error
	DeleteByPermissionID(ctx context.Context, permissionID uuid.UUID) error
	HasPermission(ctx context.Context, role models.UserRole, resource, action string) (bool, error)
}

// rolePermissionRepository 角色权限仓库实现
type rolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository 创建角色权限仓库实例
func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

// Create 创建角色权限
func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *models.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

// GetPermissionsByRole 获取角色的所有权限
func (r *rolePermissionRepository) GetPermissionsByRole(ctx context.Context, role models.UserRole) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role = ?", role).
		Find(&permissions).Error
	return permissions, err
}

// Exists 检查角色权限是否存在
func (r *rolePermissionRepository) Exists(ctx context.Context, role models.UserRole, permissionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RolePermission{}).
		Where("role = ? AND permission_id = ?", role, permissionID).
		Count(&count).Error
	return count > 0, err
}

// Delete 删除角色权限
func (r *rolePermissionRepository) Delete(ctx context.Context, role models.UserRole, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("role = ? AND permission_id = ?", role, permissionID).
		Delete(&models.RolePermission{}).Error
}

// DeleteByPermissionID 根据权限ID删除所有相关的角色权限
func (r *rolePermissionRepository) DeleteByPermissionID(ctx context.Context, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("permission_id = ?", permissionID).
		Delete(&models.RolePermission{}).Error
}

// HasPermission 检查角色是否有指定权限
func (r *rolePermissionRepository) HasPermission(ctx context.Context, role models.UserRole, resource, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("role_permissions").
		Joins("INNER JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role = ? AND permissions.resource = ? AND permissions.action = ?", role, resource, action).
		Count(&count).Error
	return count > 0, err
}

// UserPermissionRepository 用户权限仓库接口
type UserPermissionRepository interface {
	Create(ctx context.Context, userPermission *models.UserPermission) error
	GetPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error)
	Exists(ctx context.Context, userID, permissionID uuid.UUID) (bool, error)
	UpdateGranted(ctx context.Context, userID, permissionID uuid.UUID, granted bool) error
	Delete(ctx context.Context, userID, permissionID uuid.UUID) error
	DeleteByPermissionID(ctx context.Context, permissionID uuid.UUID) error
	HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

// userPermissionRepository 用户权限仓库实现
type userPermissionRepository struct {
	db *gorm.DB
}

// NewUserPermissionRepository 创建用户权限仓库实例
func NewUserPermissionRepository(db *gorm.DB) UserPermissionRepository {
	return &userPermissionRepository{db: db}
}

// Create 创建用户权限
func (r *userPermissionRepository) Create(ctx context.Context, userPermission *models.UserPermission) error {
	return r.db.WithContext(ctx).Create(userPermission).Error
}

// GetPermissionsByUserID 获取用户的所有直接权限
func (r *userPermissionRepository) GetPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ? AND user_permissions.granted = ?", userID, true).
		Find(&permissions).Error
	return permissions, err
}

// Exists 检查用户权限是否存在
func (r *userPermissionRepository) Exists(ctx context.Context, userID, permissionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserPermission{}).
		Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Count(&count).Error
	return count > 0, err
}

// UpdateGranted 更新用户权限的授予状态
func (r *userPermissionRepository) UpdateGranted(ctx context.Context, userID, permissionID uuid.UUID, granted bool) error {
	return r.db.WithContext(ctx).
		Model(&models.UserPermission{}).
		Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Update("granted", granted).Error
}

// Delete 删除用户权限
func (r *userPermissionRepository) Delete(ctx context.Context, userID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Delete(&models.UserPermission{}).Error
}

// DeleteByPermissionID 根据权限ID删除所有相关的用户权限
func (r *userPermissionRepository) DeleteByPermissionID(ctx context.Context, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("permission_id = ?", permissionID).
		Delete(&models.UserPermission{}).Error
}

// HasPermission 检查用户是否有指定权限
func (r *userPermissionRepository) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_permissions").
		Joins("INNER JOIN permissions ON user_permissions.permission_id = permissions.id").
		Where("user_permissions.user_id = ? AND user_permissions.granted = ? AND permissions.resource = ? AND permissions.action = ?", 
			userID, true, resource, action).
		Count(&count).Error
	return count > 0, err
}