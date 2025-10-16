package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/models"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/utils"
)

// TenantService 
type TenantService interface {
	// 
	CreateTenant(ctx context.Context, req *models.CreateTenantRequest) (*models.TenantResponse, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.TenantResponse, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*models.TenantResponse, error)
	GetTenantByDomain(ctx context.Context, domain string) (*models.TenantResponse, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantRequest) (*models.TenantResponse, error)
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	ListTenants(ctx context.Context, query *models.TenantQuery) (*models.TenantListResponse, error)
	
	// 
	UpdateTenantSettings(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantSettingsRequest) error
	GetTenantSettings(ctx context.Context, tenantID uuid.UUID) (*models.TenantSettings, error)
	
	// 
	UpdateTenantQuota(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantQuotaRequest) error
	GetTenantQuota(ctx context.Context, tenantID uuid.UUID) (*models.TenantQuota, error)
	CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error
	UpdateUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error
	
	// 
	AddTenantUser(ctx context.Context, tenantID uuid.UUID, req *models.AddTenantUserRequest) error
	RemoveTenantUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
	UpdateTenantUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *models.UpdateTenantUserRequest) error
	GetTenantUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) (*models.TenantUserResponse, error)
	ListTenantUsers(ctx context.Context, tenantID uuid.UUID, query *models.TenantUserQuery) (*models.TenantUserListResponse, error)
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]models.TenantResponse, error)
	
	// 
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*models.TenantStatsResponse, error)
	GetTenantHealth(ctx context.Context, tenantID uuid.UUID) (*models.TenantHealthResponse, error)
	
	// ?
	ActivateTenant(ctx context.Context, tenantID uuid.UUID) error
	SuspendTenant(ctx context.Context, tenantID uuid.UUID, reason string) error
	DeactivateTenant(ctx context.Context, tenantID uuid.UUID) error
	
	// 
	GetTenantContext(ctx context.Context, tenantID uuid.UUID) (*models.TenantContext, error)
	ValidateTenantAccess(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) error
}

// tenantService 
type tenantService struct {
	tenantRepo     repositories.TenantRepository
	tenantUserRepo repositories.TenantUserRepository
	db             *gorm.DB
	cache          utils.CacheService
	logger         utils.Logger
	config         *models.TenantConfig
}

// NewTenantService 
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

// CreateTenant 
func (s *tenantService) CreateTenant(ctx context.Context, req *models.CreateTenantRequest) (*models.TenantResponse, error) {
	// ?
	if err := s.validateSubdomain(ctx, req.Subdomain); err != nil {
		return nil, err
	}
	
	// ?
	if req.Domain != "" {
		if err := s.validateDomain(ctx, req.Domain); err != nil {
			return nil, err
		}
	}
	
	// 
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
	
	// 
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	if req.Quota != nil {
		tenant.Quota = *req.Quota
	}
	
	// ?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 
	if err := s.tenantRepo.Create(ctx, tx, tenant); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create tenant", "error", err, "tenant_name", req.Name)
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	
	// ?
	if req.AdminUser != nil {
		if err := s.createAdminUser(ctx, tx, tenant.ID, req.AdminUser); err != nil {
			tx.Rollback()
			s.logger.Error("Failed to create admin user", "error", err, "tenant_id", tenant.ID)
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
	}
	
	// ?
	if err := s.initializeTenantIsolation(ctx, tx, tenant); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to initialize tenant isolation", "error", err, "tenant_id", tenant.ID)
		return nil, fmt.Errorf("failed to initialize tenant isolation: %w", err)
	}
	
	// 
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit tenant creation transaction", "error", err, "tenant_id", tenant.ID)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// 
	s.clearTenantCache(tenant.ID)
	
	s.logger.Info("Tenant created successfully", "tenant_id", tenant.ID, "tenant_name", tenant.Name)
	
	return s.tenantToResponse(tenant), nil
}

// GetTenant 
func (s *tenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*models.TenantResponse, error) {
	// ?
	cacheKey := fmt.Sprintf("tenant:%s", tenantID.String())
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tenant, ok := cached.(*models.Tenant); ok {
			return s.tenantToResponse(tenant), nil
		}
	}
	
	// 
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 
	s.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
	
	return s.tenantToResponse(tenant), nil
}

// GetTenantBySubdomain ?
func (s *tenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*models.TenantResponse, error) {
	// ?
	cacheKey := fmt.Sprintf("tenant:subdomain:%s", subdomain)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tenant, ok := cached.(*models.Tenant); ok {
			return s.tenantToResponse(tenant), nil
		}
	}
	
	// 
	tenant, err := s.tenantRepo.GetBySubdomain(ctx, s.db, subdomain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant by subdomain", "error", err, "subdomain", subdomain)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 
	s.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
	
	return s.tenantToResponse(tenant), nil
}

// GetTenantByDomain 
func (s *tenantService) GetTenantByDomain(ctx context.Context, domain string) (*models.TenantResponse, error) {
	// ?
	cacheKey := fmt.Sprintf("tenant:domain:%s", domain)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		if tenant, ok := cached.(*models.Tenant); ok {
			return s.tenantToResponse(tenant), nil
		}
	}
	
	// 
	tenant, err := s.tenantRepo.GetByDomain(ctx, s.db, domain)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		s.logger.Error("Failed to get tenant by domain", "error", err, "domain", domain)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 
	s.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
	
	return s.tenantToResponse(tenant), nil
}

// UpdateTenant 
func (s *tenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *models.UpdateTenantRequest) (*models.TenantResponse, error) {
	// 
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// 
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
		// ?
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
	
	// 
	if err := s.tenantRepo.Update(ctx, s.db, tenant); err != nil {
		s.logger.Error("Failed to update tenant", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	
	// 
	s.clearTenantCache(tenantID)
	
	s.logger.Info("Tenant updated successfully", "tenant_id", tenantID)
	
	return s.tenantToResponse(tenant), nil
}

// DeleteTenant 
func (s *tenantService) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	// ?
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("tenant not found")
		}
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	// ?
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 
	if err := s.tenantUserRepo.DeleteByTenantID(ctx, tx, tenantID); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete tenant users", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete tenant users: %w", err)
	}
	
	// 
	if err := s.cleanupTenantIsolation(ctx, tx, tenant); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to cleanup tenant isolation", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to cleanup tenant isolation: %w", err)
	}
	
	// 
	if err := s.tenantRepo.Delete(ctx, tx, tenantID); err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete tenant", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	
	// 
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit tenant deletion transaction", "error", err, "tenant_id", tenantID)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// 
	s.clearTenantCache(tenantID)
	
	s.logger.Info("Tenant deleted successfully", "tenant_id", tenantID)
	
	return nil
}

// ListTenants 
func (s *tenantService) ListTenants(ctx context.Context, query *models.TenantQuery) (*models.TenantListResponse, error) {
	// ?
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
	
	// 
	tenants, total, err := s.tenantRepo.List(ctx, s.db, query)
	if err != nil {
		s.logger.Error("Failed to list tenants", "error", err)
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	
	// ?
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

// CheckQuota ?
func (s *tenantService) CheckQuota(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error {
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	return tenant.CheckQuota(resource, amount)
}

// UpdateUsage ?
func (s *tenantService) UpdateUsage(ctx context.Context, tenantID uuid.UUID, resource string, amount int) error {
	tenant, err := s.tenantRepo.GetByID(ctx, s.db, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	
	tenant.UpdateUsage(resource, amount)
	
	if err := s.tenantRepo.Update(ctx, s.db, tenant); err != nil {
		return fmt.Errorf("failed to update tenant usage: %w", err)
	}
	
	// 
	s.clearTenantCache(tenantID)
	
	return nil
}

// ValidateTenantAccess 
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

// 

// validateSubdomain ?
func (s *tenantService) validateSubdomain(ctx context.Context, subdomain string) error {
	// ?
	if len(subdomain) < 2 || len(subdomain) > 100 {
		return fmt.Errorf("subdomain must be between 2 and 100 characters")
	}
	
	// ?
	for _, char := range subdomain {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return fmt.Errorf("subdomain can only contain lowercase letters, numbers, and hyphens")
		}
	}
	
	// ?
	existing, err := s.tenantRepo.GetBySubdomain(ctx, s.db, subdomain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check subdomain uniqueness: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("subdomain already exists")
	}
	
	return nil
}

// validateDomain ?
func (s *tenantService) validateDomain(ctx context.Context, domain string) error {
	if domain == "" {
		return nil
	}
	
	// ?
	existing, err := s.tenantRepo.GetByDomain(ctx, s.db, domain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check domain uniqueness: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("domain already exists")
	}
	
	return nil
}

// createAdminUser ?
func (s *tenantService) createAdminUser(ctx context.Context, tx *gorm.DB, tenantID uuid.UUID, req *models.CreateAdminUserRequest) error {
	// TODO: 
	// 
	
	// ?
	tenantUser := &models.TenantUser{
		ID:          uuid.New(),
		TenantID:    tenantID,
		UserID:      uuid.New(), // ID
		Role:        "admin",
		Status:      "active",
		Permissions: []string{"*"}, // ?
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	return s.tenantUserRepo.Create(ctx, tx, tenantUser)
}

// initializeTenantIsolation ?
func (s *tenantService) initializeTenantIsolation(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	switch tenant.IsolationStrategy {
	case models.IsolationStrategySchema:
		return s.createTenantSchema(ctx, tx, tenant)
	case models.IsolationStrategyDatabase:
		return s.createTenantDatabase(ctx, tx, tenant)
	case models.IsolationStrategyRowLevel:
		// Row Level Security ?
		return nil
	default:
		return fmt.Errorf("unsupported isolation strategy: %s", tenant.IsolationStrategy)
	}
}

// createTenantSchema 
func (s *tenantService) createTenantSchema(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	// 
	if err := tx.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
		return fmt.Errorf("failed to create tenant schema: %w", err)
	}
	
	// TODO: ?
	
	return nil
}

// createTenantDatabase ?
func (s *tenantService) createTenantDatabase(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	dbName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	// ?
	if err := tx.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName)).Error; err != nil {
		return fmt.Errorf("failed to create tenant database: %w", err)
	}
	
	// TODO: 
	
	return nil
}

// cleanupTenantIsolation 
func (s *tenantService) cleanupTenantIsolation(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	switch tenant.IsolationStrategy {
	case models.IsolationStrategySchema:
		return s.dropTenantSchema(ctx, tx, tenant)
	case models.IsolationStrategyDatabase:
		return s.dropTenantDatabase(ctx, tx, tenant)
	case models.IsolationStrategyRowLevel:
		// Row Level Security 
		return nil
	default:
		return fmt.Errorf("unsupported isolation strategy: %s", tenant.IsolationStrategy)
	}
}

// dropTenantSchema 
func (s *tenantService) dropTenantSchema(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	if err := tx.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error; err != nil {
		return fmt.Errorf("failed to drop tenant schema: %w", err)
	}
	
	return nil
}

// dropTenantDatabase ?
func (s *tenantService) dropTenantDatabase(ctx context.Context, tx *gorm.DB, tenant *models.Tenant) error {
	dbName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenant.ID.String(), "-", "_"))
	
	if err := tx.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)).Error; err != nil {
		return fmt.Errorf("failed to drop tenant database: %w", err)
	}
	
	return nil
}

// clearTenantCache 
func (s *tenantService) clearTenantCache(tenantID uuid.UUID) {
	ctx := context.Background()
	
	// ?
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

// tenantToResponse ?
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

