package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taishanglaojun/core-services/multi-tenancy/models"
	"taishanglaojun/core-services/multi-tenancy/repositories"
	"taishanglaojun/core-services/multi-tenancy/utils"
)

// TenantService 租户服务接口
type TenantService interface {
	// 租户管理
	CreateTenant(ctx context.Context, req *models.CreateTenantRequest) (*models.TenantResponse, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.TenantResponse, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*models.TenantResponse, error)
	GetTenantByDomain(ctx context.Context, domain string) (*models.TenantResponse, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantRequest) (*models.TenantResponse, error)
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	ListTenants(ctx context.Context, query *models.TenantQuery) (*models.TenantListResponse, error)
	
	// 租户设置
	UpdateTenantSettings(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantSettingsRequest) error
	GetTenantSettings(ctx context.Context, tenantID uuid.UUID) (*models.TenantSettings, error)
	
	// 租户配额
	UpdateTenantQuota(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantQuotaRequest) error
	GetTenantQuota(ctx context.Context, tenantID uuid.UUID) (*models.TenantQuota, error)
	CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error
	UpdateUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error
	
	// 租户用户管理
	AddTenantUser(ctx context.Context, tenantID uuid.UUID, req *models.AddTenantUserRequest) error
	RemoveTenantUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
	UpdateTenantUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *models.UpdateTenantUserRequest) error
	GetTenantUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (*models.TenantUserResponse, error)
	ListTenantUsers(ctx context.Context, tenantID uuid.UUID, query *models.TenantUserQuery) (*models.TenantUserListResponse, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]models.TenantResponse, error)
	
	// 租户统计
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*models.TenantStatsResponse, error)
	GetTenantHealth(ctx context.Context, tenantID uuid.UUID) (*models.TenantHealthResponse, error)
	
	// 租户状态管�?
	ActivateTenant(ctx context.Context, tenantID uuid.UUID) error
	SuspendTenant(ctx context.Context, tenantID uuid.UUID, reason string) error
	DeactivateTenant(ctx context.Context, tenantID uuid.UUID) error
	
	// 数据隔离
	GetTenantContext(ctx context.Context, tenantID uuid.UUID) (*models.TenantContext, error)
	ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
}

// tenantService 租户服务实现
type tenantService struct {
	tenantRepo     repositories.TenantRepository
	tenantUserRepo repositories.TenantUserRepository
	db             *gorm.DB
	cache          utils.CacheService
	logger         utils.Logger
	config         *models.TenantConfig
}

// NewTenantService 创建租户服务
func NewTenantService(
	tenantRepo repositories.TenantRepository,
	tenantUserRepo repositories.TenantUserRepository,
	db *gorm.DB,
	cache utils.CacheService,
	logger utils.Logger,
	config *models.TenantConfig,
) TenantService {
	return &tenantService{
		tenantRepo:     tenantRepo,
		tenantUserRepo: tenantUserRepo,
		db:             db,
		cache:          cache,
		logger:         logger,
		config:         config,
	}
}

// CreateTenant 创建租户
func (s *tenantService) CreateTenant(ctx context.Context, req *models.CreateTenantRequest) (*models.TenantResponse, error) {
	// 验证子域名唯一�?
	if err := s.validateSubdomain(ctx, req.Subdomain); err != nil {
		return nil, err
	}
	
	// 验证域名唯一性（如果提供�?
	if req.Domain != "" {
		if err := s.validateDomain(ctx, req.Domain); err != nil {
			return nil, err
		}
	}
	
	// 创建租户
	tenant := &models.Tenant{
		ID:                uuid.New(),
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Description:       req.Description,
		Subdomain:         req.Subdomain,
		Domain:            req.Domain,
		Status:            models.TenantStatusActive,
		IsolationStrategy: req.IsolationStrategy,
		Settings:          models.DefaultTenantSettings(),
		Quota:             models.DefaultTenantQuota(),
		Usage:             models.TenantUsage{},
		Metadata:          req.Metadata,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	
	// 应用自定义设置和配额
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	if req.Quota != nil {
		tenant.Quota = *req.Quota
	}
	
	// 开始事�?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 保存租户
	if err := s.tenantRepo.Create(ctx, tx, tenant); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create tenant", "error", err, "tenant_name", req.Name)
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	
	// 创建管理员用户（如果提供�?
	if req.AdminUser != nil {
		if err := s.createAdminUser(ctx, tx, tenant.ID, req.AdminUser); err != nil {
			tx.Rollback()
			s.logger.Error("Failed to create admin user", "error", err, "tenant_id", tenant.ID)
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
	}
	
	// 初始化租户数据隔�?
	if err := s.initializeTenantIsolation(ctx, tx, tenant); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to initialize tenant isolation", "error", err, "tenant_id", tenant.ID)
		return nil, fmt.Errorf("failed to initialize tenant isolation: %w", err)
	}
	
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit tenant creation transaction", "error", err, "tenant_id", tenant.ID)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// 清除缓存
	s.clearTenantCache(tenant.ID)
	
	s.logger.Info("Tenant created successfully", "tenant_id", tenant.ID, "tenant_name", tenant.Name)
	
	return s.tenantToResponse(tenant), nil
}

// GetTenant 获取租户
func (s *tenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.TenantResponse, error) {
	// 尝试从缓存获�?
	cacheKey := fmt.Sprintf("tenant:%s", tenantID.String())
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tenant, ok := cached.(*models.Tenant); ok {
			return s.tenantToResponse(tenant), nil
		}
	}
	
	// 从数据库获取
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 缓存结果
	s.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
	
	return s.tenantToResponse(tenant), nil
}

// GetTenantBySubdomain 通过子域名获取租�?
func (s *tenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*models.TenantResponse, error) {
	// 尝试从缓存获�?
	cacheKey := fmt.Sprintf("tenant:subdomain:%s", subdomain)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tenant, ok := cached.(*models.Tenant); ok {
			return s.tenantToResponse(tenant), nil
		}
	}
	
	// 从数据库获取
	tenant, err := s.tenantRepo.GetBySubdomain(ctx, s.db, subdomain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant by subdomain", "error", err, "subdomain", subdomain)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 缓存结果
	s.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
	
	return s.tenantToResponse(tenant), nil
}

// GetTenantByDomain 通过域名获取租户
func (s *tenantService) GetTenantByDomain(ctx context.Context, domain string) (*models.TenantResponse, error) {
	// 尝试从缓存获�?
	cacheKey := fmt.Sprintf("tenant:domain:%s", domain)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tenant, ok := cached.(*models.Tenant); ok {
			return s.tenantToResponse(tenant), nil
		}
	}
	
	// 从数据库获取
	tenant, err := s.tenantRepo.GetByDomain(ctx, s.db, domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant by domain", "error", err, "domain", domain)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 缓存结果
	s.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
	
	return s.tenantToResponse(tenant), nil
}

// UpdateTenant 更新租户
func (s *tenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantRequest) (*models.TenantResponse, error) {
	// 获取现有租户
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 更新字段
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.DisplayName != nil {
		tenant.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		tenant.Description = *req.Description
	}
	if req.Domain != nil {
		// 验证域名唯一�?
		if *req.Domain != tenant.Domain {
			if err := s.validateDomain(ctx, *req.Domain); err != nil {
				return nil, err
			}
		}
		tenant.Domain = *req.Domain
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}
	
	tenant.UpdatedAt = time.Now()
	
	// 保存更新
	if err := s.tenantRepo.Update(ctx, s.db, tenant); err != nil {
		s.logger.Error("Failed to update tenant", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	
	// 清除缓存
	s.clearTenantCache(tenantID)
	
	s.logger.Info("Tenant updated successfully", "tenant_id", tenantID)
	
	return s.tenantToResponse(tenant), nil
}

// DeleteTenant 删除租户
func (s *tenantService) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	// 检查租户是否存�?
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tenant not found")
		}
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 开始事�?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 删除租户用户关联
	if err := s.tenantUserRepo.DeleteByTenantID(ctx, tx, tenantID); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete tenant users", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete tenant users: %w", err)
	}
	
	// 清理租户数据隔离
	if err := s.cleanupTenantIsolation(ctx, tx, tenant); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to cleanup tenant isolation", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to cleanup tenant isolation: %w", err)
	}
	
	// 删除租户
	if err := s.tenantRepo.Delete(ctx, tx, tenantID); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete tenant", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit tenant deletion transaction", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// 清除缓存
	s.clearTenantCache(tenantID)
	
	s.logger.Info("Tenant deleted successfully", "tenant_id", tenantID)
	
	return nil
}

// ListTenants 列出租户
func (s *tenantService) ListTenants(ctx context.Context, query *models.TenantQuery) (*models.TenantListResponse, error) {
	// 设置默认�?
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.OrderBy == "" {
		query.OrderBy = "created_at"
	}
	if query.Order == "" {
		query.Order = "desc"
	}
	
	// 获取租户列表
	tenants, total, err := s.tenantRepo.List(ctx, s.db, query)
	if err != nil {
		s.logger.Error("Failed to list tenants", "error", err)
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	
	// 转换为响应格�?
	responses := make([]models.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		responses[i] = *s.tenantToResponse(&tenant)
	}
	
	return &models.TenantListResponse{
		Tenants: responses,
		Total:   total,
		Page:    query.Page,
		Size:    query.PageSize,
	}, nil
}

// CheckQuota 检查配�?
func (s *tenantService) CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error {
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	return tenant.CheckQuota(resource, amount)
}

// UpdateUsage 更新使用�?
func (s *tenantService) UpdateUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error {
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	tenant.UpdateUsage(resource, amount)
	
	if err := s.tenantRepo.Update(ctx, s.db, tenant); err != nil {
		return fmt.Errorf("failed to update tenant usage: %w", err)
	}
	
	// 清除缓存
	s.clearTenantCache(tenantID)
	
	return nil
}

// ValidateTenantAccess 验证租户访问权限
func (s *tenantService) ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error {
	tenantUser, err := s.tenantUserRepo.GetByTenantAndUser(ctx, s.db, tenantID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user does not have access to this tenant")
		}
		return fmt.Errorf("failed to check tenant access: %w", err)
	}
	
	if tenantUser.Status != "active" {
		return fmt.Errorf("user access to tenant is not active")
	}
	
	return nil
}

// 辅助方法

// validateSubdomain 验证子域名唯一�?
func (s *tenantService) validateSubdomain(ctx context.Context, subdomain string) error {
	// 检查格�?
	if len(subdomain) < 2 || len(subdomain) > 100 {
		return fmt.Errorf("subdomain must be between 2 and 100 characters")
	}
	
	// 检查字�?
	for _, char := range subdomain {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return fmt.Errorf("subdomain can only contain lowercase letters, numbers, and hyphens")
		}
	}
	
	// 检查唯一�?
	existing, err := s.tenantRepo.GetBySubdomain(ctx, s.db, subdomain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check subdomain uniqueness: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("subdomain already exists")
	}
	
	return nil
}

// validateDomain 验证域名唯一�?
func (s *tenantService) validateDomain(ctx context.Context, domain string) error {
	if domain == "" {
		return nil
	}
	
	// 检查唯一�?
	existing, err := s.tenantRepo.GetByDomain(ctx, s.db, domain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check domain uniqueness: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("domain already exists")
	}
	
	return nil
}

// createAdminUser 创建管理员用�?
func (s *tenantService) createAdminUser(ctx context.Context, tx *gorm.DB, tenantID uuid.UUID, req *models.CreateAdminUserRequest) error {
	// TODO: 集成用户服务创建用户
	// 这里需要调用用户服务来创建用户，然后添加到租户
	
	// 临时实现：直接创建租户用户关�?
	tenantUser := &models.TenantUser{
		ID:          uuid.New(),
		TenantID:    tenantID,
		UserID:      uuid.New(), // 这里应该是实际创建的用户ID
		Role:        "admin",
		Status:      "active",
		Permissions: []string{"*"}, // 管理员拥有所有权�?
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	return s.tenantUserRepo.Create(ctx, tx, tenantUser)
}

// initializeTenantIsolation 初始化租户数据隔�?
func (s *tenantService) initializeTenantIsolation(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	switch tenant.IsolationStrategy {
	case models.IsolationStrategySchema:
		return s.createTenantSchema(ctx, tx, tenant)
	case models.IsolationStrategyDatabase:
		return s.createTenantDatabase(ctx, tx, tenant)
	case models.IsolationStrategyRowLevel:
		// Row Level Security 不需要额外的初始�?
		return nil
	default:
		return fmt.Errorf("unsupported isolation strategy: %s", tenant.IsolationStrategy)
	}
}

// createTenantSchema 创建租户模式
func (s *tenantService) createTenantSchema(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	// 创建模式
	if err := tx.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
		return fmt.Errorf("failed to create tenant schema: %w", err)
	}
	
	// TODO: 在新模式中创建必要的�?
	
	return nil
}

// createTenantDatabase 创建租户数据�?
func (s *tenantService) createTenantDatabase(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	dbName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	// 创建数据�?
	if err := tx.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)).Error; err != nil {
		return fmt.Errorf("failed to create tenant database: %w", err)
	}
	
	// TODO: 在新数据库中创建必要的表
	
	return nil
}

// cleanupTenantIsolation 清理租户数据隔离
func (s *tenantService) cleanupTenantIsolation(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	switch tenant.IsolationStrategy {
	case models.IsolationStrategySchema:
		return s.dropTenantSchema(ctx, tx, tenant)
	case models.IsolationStrategyDatabase:
		return s.dropTenantDatabase(ctx, tx, tenant)
	case models.IsolationStrategyRowLevel:
		// Row Level Security 不需要额外的清理
		return nil
	default:
		return fmt.Errorf("unsupported isolation strategy: %s", tenant.IsolationStrategy)
	}
}

// dropTenantSchema 删除租户模式
func (s *tenantService) dropTenantSchema(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	if err := tx.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error; err != nil {
		return fmt.Errorf("failed to drop tenant schema: %w", err)
	}
	
	return nil
}

// dropTenantDatabase 删除租户数据�?
func (s *tenantService) dropTenantDatabase(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	dbName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	if err := tx.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).Error; err != nil {
		return fmt.Errorf("failed to drop tenant database: %w", err)
	}
	
	return nil
}

// clearTenantCache 清除租户缓存
func (s *tenantService) clearTenantCache(tenantID uuid.UUID) {
	ctx := context.Background()
	
	// 清除各种缓存�?
	cacheKeys := []string{
		fmt.Sprintf("tenant:%s", tenantID.String()),
		fmt.Sprintf("tenant:stats:%s", tenantID.String()),
		fmt.Sprintf("tenant:health:%s", tenantID.String()),
		fmt.Sprintf("tenant:config:%s", tenantID.String()),
	}
	
	for _, key := range cacheKeys {
		s.cache.Delete(ctx, key)
	}
}

// tenantToResponse 转换租户为响应格�?
func (s *tenantService) tenantToResponse(tenant *models.Tenant) *models.TenantResponse {
	return &models.TenantResponse{
		ID:                tenant.ID,
		Name:              tenant.Name,
		DisplayName:       tenant.DisplayName,
		Description:       tenant.Description,
		Subdomain:         tenant.Subdomain,
		Domain:            tenant.Domain,
		Status:            tenant.Status,
		Settings:          tenant.Settings,
		Quota:             tenant.Quota,
		Usage:             tenant.Usage,
		Metadata:          tenant.Metadata,
		IsolationStrategy: tenant.IsolationStrategy,
		CreatedAt:         tenant.CreatedAt,
		UpdatedAt:         tenant.UpdatedAt,
	}
}
