package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taishanglaojun/core-services/multi-tenancy/models"
)

// TenantRepository 租户数据访问接口
type TenantRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, db *gorm.DB, tenant *models.Tenant) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*models.Tenant, error)
	GetBySubdomain(ctx context.Context, db *gorm.DB, subdomain string) (*models.Tenant, error)
	GetByDomain(ctx context.Context, db *gorm.DB, domain string) (*models.Tenant, error)
	Update(ctx context.Context, db *gorm.DB, tenant *models.Tenant) error
	Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
	
	// 查询操作
	List(ctx context.Context, db *gorm.DB, query *models.TenantQuery) ([]models.Tenant, int64, error)
	GetByStatus(ctx context.Context, db *gorm.DB, status models.TenantStatus) ([]models.Tenant, error)
	Search(ctx context.Context, db *gorm.DB, keyword string, limit int) ([]models.Tenant, error)
	
	// 统计操作
	Count(ctx context.Context, db *gorm.DB) (int64, error)
	CountByStatus(ctx context.Context, db *gorm.DB, status models.TenantStatus) (int64, error)
	
	// 批量操作
	BatchUpdate(ctx context.Context, db *gorm.DB, ids []uuid.UUID, updates map[string]interface{}) error
	BatchDelete(ctx context.Context, db *gorm.DB, ids []uuid.UUID) error
}

// tenantRepository 租户数据访问实现
type tenantRepository struct{}

// NewTenantRepository 创建租户数据访问实例
func NewTenantRepository() TenantRepository {
	return &tenantRepository{}
}

// Create 创建租户
func (r *tenantRepository) Create(ctx context.Context, db *gorm.DB, tenant *models.Tenant) error {
	return db.WithContext(ctx).Create(tenant).Error
}

// GetByID 根据ID获取租户
func (r *tenantRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	err := db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetBySubdomain 根据子域名获取租�?
func (r *tenantRepository) GetBySubdomain(ctx context.Context, db *gorm.DB, subdomain string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := db.WithContext(ctx).Where("subdomain = ?", subdomain).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetByDomain 根据域名获取租户
func (r *tenantRepository) GetByDomain(ctx context.Context, db *gorm.DB, domain string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// Update 更新租户
func (r *tenantRepository) Update(ctx context.Context, db *gorm.DB, tenant *models.Tenant) error {
	return db.WithContext(ctx).Save(tenant).Error
}

// Delete 删除租户
func (r *tenantRepository) Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	return db.WithContext(ctx).Where("id = ?", id).Delete(&models.Tenant{}).Error
}

// List 列出租户
func (r *tenantRepository) List(ctx context.Context, db *gorm.DB, query *models.TenantQuery) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64
	
	// 构建查询
	dbQuery := db.WithContext(ctx).Model(&models.Tenant{})
	
	// 应用过滤条件
	if query.Name != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query.Name+"%")
	}
	
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		dbQuery = dbQuery.Where(
			"name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}
	
	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}
	
	// 应用排序
	orderBy := query.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	
	order := query.Order
	if order == "" {
		order = "desc"
	}
	
	dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", orderBy, order))
	
	// 应用分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		dbQuery = dbQuery.Offset(offset).Limit(query.PageSize)
	}
	
	// 执行查询
	if err := dbQuery.Find(&tenants).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	
	return tenants, total, nil
}

// GetByStatus 根据状态获取租�?
func (r *tenantRepository) GetByStatus(ctx context.Context, db *gorm.DB, status models.TenantStatus) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := db.WithContext(ctx).Where("status = ?", status).Find(&tenants).Error
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

// Search 搜索租户
func (r *tenantRepository) Search(ctx context.Context, db *gorm.DB, keyword string, limit int) ([]models.Tenant, error) {
	var tenants []models.Tenant
	
	searchPattern := "%" + keyword + "%"
	query := db.WithContext(ctx).Where(
		"name ILIKE ? OR display_name ILIKE ? OR description ILIKE ? OR subdomain ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&tenants).Error
	if err != nil {
		return nil, err
	}
	
	return tenants, nil
}

// Count 统计租户总数
func (r *tenantRepository) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.Tenant{}).Count(&count).Error
	return count, err
}

// CountByStatus 根据状态统计租户数�?
func (r *tenantRepository) CountByStatus(ctx context.Context, db *gorm.DB, status models.TenantStatus) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.Tenant{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// BatchUpdate 批量更新租户
func (r *tenantRepository) BatchUpdate(ctx context.Context, db *gorm.DB, ids []uuid.UUID, updates map[string]interface{}) error {
	if len(ids) == 0 {
		return nil
	}
	
	return db.WithContext(ctx).Model(&models.Tenant{}).Where("id IN ?", ids).Updates(updates).Error
}

// BatchDelete 批量删除租户
func (r *tenantRepository) BatchDelete(ctx context.Context, db *gorm.DB, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	
	return db.WithContext(ctx).Where("id IN ?", ids).Delete(&models.Tenant{}).Error
}

// TenantUserRepository 租户用户数据访问接口
type TenantUserRepository interface {
	// 基本CRUD操作
	Create(ctx context.Context, db *gorm.DB, tenantUser *models.TenantUser) error
	GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*models.TenantUser, error)
	GetByTenantAndUser(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) (*models.TenantUser, error)
	Update(ctx context.Context, db *gorm.DB, tenantUser *models.TenantUser) error
	Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error
	DeleteByTenantAndUser(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) error
	DeleteByTenantID(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) error
	DeleteByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) error
	
	// 查询操作
	ListByTenant(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, query *models.TenantUserQuery) ([]models.TenantUser, int64, error)
	ListByUser(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]models.TenantUser, error)
	GetTenantUsers(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, role string) ([]models.TenantUser, error)
	GetUserTenants(ctx context.Context, db *gorm.DB, userID uuid.UUID, status string) ([]models.TenantUser, error)
	
	// 权限检�?
	HasPermission(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) ([]string, error)
	
	// 统计操作
	CountByTenant(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) (int64, error)
	CountByUser(ctx context.Context, db *gorm.DB, userID uuid.UUID) (int64, error)
	CountByRole(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, role string) (int64, error)
}

// tenantUserRepository 租户用户数据访问实现
type tenantUserRepository struct{}

// NewTenantUserRepository 创建租户用户数据访问实例
func NewTenantUserRepository() TenantUserRepository {
	return &tenantUserRepository{}
}

// Create 创建租户用户关联
func (r *tenantUserRepository) Create(ctx context.Context, db *gorm.DB, tenantUser *models.TenantUser) error {
	return db.WithContext(ctx).Create(tenantUser).Error
}

// GetByID 根据ID获取租户用户关联
func (r *tenantUserRepository) GetByID(ctx context.Context, db *gorm.DB, id uuid.UUID) (*models.TenantUser, error) {
	var tenantUser models.TenantUser
	err := db.WithContext(ctx).Where("id = ?", id).First(&tenantUser).Error
	if err != nil {
		return nil, err
	}
	return &tenantUser, nil
}

// GetByTenantAndUser 根据租户和用户获取关�?
func (r *tenantUserRepository) GetByTenantAndUser(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) (*models.TenantUser, error) {
	var tenantUser models.TenantUser
	err := db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&tenantUser).Error
	if err != nil {
		return nil, err
	}
	return &tenantUser, nil
}

// Update 更新租户用户关联
func (r *tenantUserRepository) Update(ctx context.Context, db *gorm.DB, tenantUser *models.TenantUser) error {
	return db.WithContext(ctx).Save(tenantUser).Error
}

// Delete 删除租户用户关联
func (r *tenantUserRepository) Delete(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	return db.WithContext(ctx).Where("id = ?", id).Delete(&models.TenantUser{}).Error
}

// DeleteByTenantAndUser 根据租户和用户删除关�?
func (r *tenantUserRepository) DeleteByTenantAndUser(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) error {
	return db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&models.TenantUser{}).Error
}

// DeleteByTenantID 根据租户ID删除所有关�?
func (r *tenantUserRepository) DeleteByTenantID(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) error {
	return db.WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&models.TenantUser{}).Error
}

// DeleteByUserID 根据用户ID删除所有关�?
func (r *tenantUserRepository) DeleteByUserID(ctx context.Context, db *gorm.DB, userID uuid.UUID) error {
	return db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.TenantUser{}).Error
}

// ListByTenant 列出租户的用�?
func (r *tenantUserRepository) ListByTenant(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, query *models.TenantUserQuery) ([]models.TenantUser, int64, error) {
	var tenantUsers []models.TenantUser
	var total int64
	
	// 构建查询
	dbQuery := db.WithContext(ctx).Model(&models.TenantUser{}).Where("tenant_id = ?", tenantID)
	
	// 应用过滤条件
	if query.Role != "" {
		dbQuery = dbQuery.Where("role = ?", query.Role)
	}
	
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	
	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tenant users: %w", err)
	}
	
	// 应用排序
	orderBy := query.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	
	order := query.Order
	if order == "" {
		order = "desc"
	}
	
	dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", orderBy, order))
	
	// 应用分页
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		dbQuery = dbQuery.Offset(offset).Limit(query.PageSize)
	}
	
	// 执行查询
	if err := dbQuery.Find(&tenantUsers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tenant users: %w", err)
	}
	
	return tenantUsers, total, nil
}

// ListByUser 列出用户的租�?
func (r *tenantUserRepository) ListByUser(ctx context.Context, db *gorm.DB, userID uuid.UUID) ([]models.TenantUser, error) {
	var tenantUsers []models.TenantUser
	err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&tenantUsers).Error
	if err != nil {
		return nil, err
	}
	return tenantUsers, nil
}

// GetTenantUsers 获取租户的特定角色用�?
func (r *tenantUserRepository) GetTenantUsers(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, role string) ([]models.TenantUser, error) {
	var tenantUsers []models.TenantUser
	query := db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	
	if role != "" {
		query = query.Where("role = ?", role)
	}
	
	err := query.Find(&tenantUsers).Error
	if err != nil {
		return nil, err
	}
	return tenantUsers, nil
}

// GetUserTenants 获取用户的特定状态租�?
func (r *tenantUserRepository) GetUserTenants(ctx context.Context, db *gorm.DB, userID uuid.UUID, status string) ([]models.TenantUser, error) {
	var tenantUsers []models.TenantUser
	query := db.WithContext(ctx).Where("user_id = ?", userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Find(&tenantUsers).Error
	if err != nil {
		return nil, err
	}
	return tenantUsers, nil
}

// HasPermission 检查用户是否有特定权限
func (r *tenantUserRepository) HasPermission(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID, permission string) (bool, error) {
	var tenantUser models.TenantUser
	err := db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND status = ?", tenantID, userID, "active").First(&tenantUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	
	// 检查权�?
	for _, perm := range tenantUser.Permissions {
		if perm == "*" || perm == permission {
			return true, nil
		}
		
		// 支持通配符权�?
		if strings.HasSuffix(perm, "*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(permission, prefix) {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// GetUserPermissions 获取用户在租户中的权�?
func (r *tenantUserRepository) GetUserPermissions(ctx context.Context, db *gorm.DB, tenantID, userID uuid.UUID) ([]string, error) {
	var tenantUser models.TenantUser
	err := db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND status = ?", tenantID, userID, "active").First(&tenantUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []string{}, nil
		}
		return nil, err
	}
	
	return tenantUser.Permissions, nil
}

// CountByTenant 统计租户的用户数�?
func (r *tenantUserRepository) CountByTenant(ctx context.Context, db *gorm.DB, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.TenantUser{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// CountByUser 统计用户的租户数�?
func (r *tenantUserRepository) CountByUser(ctx context.Context, db *gorm.DB, userID uuid.UUID) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.TenantUser{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountByRole 统计租户中特定角色的用户数量
func (r *tenantUserRepository) CountByRole(ctx context.Context, db *gorm.DB, tenantID uuid.UUID, role string) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.TenantUser{}).Where("tenant_id = ? AND role = ?", tenantID, role).Count(&count).Error
	return count, err
}
