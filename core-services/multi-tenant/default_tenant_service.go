package multitenant

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultTenantService 默认租户服务实现
type DefaultTenantService struct {
	repository TenantRepository
	cache      TenantCache
	publisher  TenantEventPublisher
	logger     *zap.Logger
	config     TenantServiceConfig
	mutex      sync.RWMutex
}

// TenantServiceConfig 租户服务配置
type TenantServiceConfig struct {
	CacheExpiry        time.Duration `json:"cache_expiry"`
	MaxTenantsPerUser  int           `json:"max_tenants_per_user"`
	DefaultPlan        TenantPlan    `json:"default_plan"`
	EnableAuditLog     bool          `json:"enable_audit_log"`
	EnableCache        bool          `json:"enable_cache"`
	EnableEvents       bool          `json:"enable_events"`
	BackupRetention    time.Duration `json:"backup_retention"`
	UsageAggregation   time.Duration `json:"usage_aggregation"`
	DefaultLimits      TenantLimits  `json:"default_limits"`
	DefaultSettings    TenantSettings `json:"default_settings"`
}

// NewDefaultTenantService 创建默认租户服务
func NewDefaultTenantService(
	repository TenantRepository,
	cache TenantCache,
	publisher TenantEventPublisher,
	config TenantServiceConfig,
	logger *zap.Logger,
) *DefaultTenantService {
	return &DefaultTenantService{
		repository: repository,
		cache:      cache,
		publisher:  publisher,
		config:     config,
		logger:     logger,
	}
}

// CreateTenant 创建租户
func (s *DefaultTenantService) CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	// 验证请求
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid create request: %w", err)
	}

	// 检查域名是否已存在
	existing, err := s.repository.GetByDomain(ctx, req.Domain)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("domain %s already exists", req.Domain)
	}

	// 应用默认设置和限制
	if req.Settings.TimeZone == "" {
		req.Settings = s.config.DefaultSettings
	}
	if req.Limits.MaxUsers == 0 {
		req.Limits = s.config.DefaultLimits
	}

	// 创建租户
	tenant := NewTenant(req)
	
	// 保存到数据库
	if err := s.repository.Create(ctx, tenant); err != nil {
		s.logger.Error("Failed to create tenant",
			zap.String("tenant_id", tenant.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 缓存租户信息
	if s.config.EnableCache {
		s.cacheTenant(ctx, tenant)
	}

	// 发布事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenant.ID, "tenant", "created", map[string]interface{}{
			"tenant_name": tenant.Name,
			"plan":        tenant.Plan,
		}, req.CreatedBy)
	}

	s.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID),
		zap.String("tenant_name", tenant.Name),
		zap.String("domain", tenant.Domain),
		zap.String("plan", string(tenant.Plan)))

	return tenant, nil
}

// GetTenant 获取租户
func (s *DefaultTenantService) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	// 尝试从缓存获取
	if s.config.EnableCache {
		if cached, err := s.getTenantFromCache(ctx, tenantID); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 从数据库获取
	tenant, err := s.repository.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// 缓存结果
	if s.config.EnableCache && tenant != nil {
		s.cacheTenant(ctx, tenant)
	}

	return tenant, nil
}

// GetTenantByDomain 根据域名获取租户
func (s *DefaultTenantService) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	// 尝试从缓存获取
	if s.config.EnableCache {
		if cached, err := s.getTenantFromCacheByDomain(ctx, domain); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 从数据库获取
	tenant, err := s.repository.GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by domain: %w", err)
	}

	// 缓存结果
	if s.config.EnableCache && tenant != nil {
		s.cacheTenant(ctx, tenant)
		s.cacheTenantByDomain(ctx, domain, tenant)
	}

	return tenant, nil
}

// UpdateTenant 更新租户
func (s *DefaultTenantService) UpdateTenant(ctx context.Context, tenantID string, req UpdateTenantRequest) (*Tenant, error) {
	// 获取现有租户
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.DisplayName != nil {
		tenant.DisplayName = *req.DisplayName
	}
	if req.Domain != nil {
		// 检查新域名是否已存在
		existing, err := s.repository.GetByDomain(ctx, *req.Domain)
		if err == nil && existing != nil && existing.ID != tenantID {
			return nil, fmt.Errorf("domain %s already exists", *req.Domain)
		}
		tenant.Domain = *req.Domain
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}
	if req.Plan != nil {
		tenant.Plan = *req.Plan
	}
	if req.Settings != nil {
		tenant.Settings = *req.Settings
	}
	if req.Limits != nil {
		tenant.Limits = *req.Limits
	}
	if req.Metadata != nil {
		tenant.Metadata = *req.Metadata
	}

	tenant.UpdatedAt = time.Now()
	tenant.UpdatedBy = req.UpdatedBy

	// 保存到数据库
	if err := s.repository.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to update tenant",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// 更新缓存
	if s.config.EnableCache {
		s.cacheTenant(ctx, tenant)
	}

	// 发布事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "tenant", "updated", map[string]interface{}{
			"changes": req,
		}, req.UpdatedBy)
	}

	s.logger.Info("Tenant updated successfully",
		zap.String("tenant_id", tenantID),
		zap.String("updated_by", req.UpdatedBy))

	return tenant, nil
}

// DeleteTenant 删除租户
func (s *DefaultTenantService) DeleteTenant(ctx context.Context, tenantID string) error {
	// 获取租户信息
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	// 软删除：更新状态为已删除
	tenant.Status = TenantStatusDeleted
	tenant.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to delete tenant",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// 清除缓存
	if s.config.EnableCache {
		s.clearTenantCache(ctx, tenantID, tenant.Domain)
	}

	// 发布事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "tenant", "deleted", map[string]interface{}{
			"tenant_name": tenant.Name,
		}, "system")
	}

	s.logger.Info("Tenant deleted successfully",
		zap.String("tenant_id", tenantID),
		zap.String("tenant_name", tenant.Name))

	return nil
}

// ListTenants 列出租户
func (s *DefaultTenantService) ListTenants(ctx context.Context, filter TenantFilter, pagination PaginationRequest) (*ListTenantsResponse, error) {
	tenants, total, err := s.repository.List(ctx, filter, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	totalPages := int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))

	return &ListTenantsResponse{
		Tenants: tenants,
		Pagination: PaginationResponse{
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// ActivateTenant 激活租户
func (s *DefaultTenantService) ActivateTenant(ctx context.Context, tenantID string) error {
	return s.updateTenantStatus(ctx, tenantID, TenantStatusActive, "activated")
}

// SuspendTenant 暂停租户
func (s *DefaultTenantService) SuspendTenant(ctx context.Context, tenantID string, reason string) error {
	err := s.updateTenantStatus(ctx, tenantID, TenantStatusSuspended, "suspended")
	if err != nil {
		return err
	}

	// 发布暂停事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "tenant", "suspended", map[string]interface{}{
			"reason": reason,
		}, "system")
	}

	return nil
}

// DeactivateTenant 停用租户
func (s *DefaultTenantService) DeactivateTenant(ctx context.Context, tenantID string) error {
	return s.updateTenantStatus(ctx, tenantID, TenantStatusInactive, "deactivated")
}

// UpdateTenantSettings 更新租户设置
func (s *DefaultTenantService) UpdateTenantSettings(ctx context.Context, tenantID string, settings TenantSettings) error {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	tenant.Settings = settings
	tenant.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, tenant); err != nil {
		return fmt.Errorf("failed to update tenant settings: %w", err)
	}

	// 更新缓存
	if s.config.EnableCache {
		s.cacheTenant(ctx, tenant)
	}

	return nil
}

// GetTenantSettings 获取租户设置
func (s *DefaultTenantService) GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error) {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &tenant.Settings, nil
}

// UpdateTenantLimits 更新租户限制
func (s *DefaultTenantService) UpdateTenantLimits(ctx context.Context, tenantID string, limits TenantLimits) error {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	tenant.Limits = limits
	tenant.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, tenant); err != nil {
		return fmt.Errorf("failed to update tenant limits: %w", err)
	}

	// 更新缓存
	if s.config.EnableCache {
		s.cacheTenant(ctx, tenant)
	}

	return nil
}

// GetTenantLimits 获取租户限制
func (s *DefaultTenantService) GetTenantLimits(ctx context.Context, tenantID string) (*TenantLimits, error) {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &tenant.Limits, nil
}

// CheckTenantLimit 检查租户限制
func (s *DefaultTenantService) CheckTenantLimit(ctx context.Context, tenantID string, limitType string, value int64) (bool, error) {
	limits, err := s.GetTenantLimits(ctx, tenantID)
	if err != nil {
		return false, err
	}

	switch limitType {
	case "users":
		return int64(limits.MaxUsers) >= value, nil
	case "storage":
		return limits.MaxStorage >= value, nil
	case "api_requests":
		return int64(limits.MaxAPIRequests) >= value, nil
	case "databases":
		return int64(limits.MaxDatabases) >= value, nil
	case "connections":
		return int64(limits.MaxConnections) >= value, nil
	case "file_size":
		return limits.MaxFileSize >= value, nil
	case "bandwidth":
		return limits.MaxBandwidth >= value, nil
	default:
		return false, fmt.Errorf("unknown limit type: %s", limitType)
	}
}

// RecordUsage 记录使用情况
func (s *DefaultTenantService) RecordUsage(ctx context.Context, tenantID string, usage TenantUsage) error {
	usage.TenantID = tenantID
	usage.Timestamp = time.Now()

	if err := s.repository.SaveUsage(ctx, &usage); err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	// 缓存当前使用情况
	if s.config.EnableCache {
		cacheKey := GetUsageCacheKey(tenantID, usage.Period)
		s.cache.Set(ctx, cacheKey, &usage, s.config.CacheExpiry)
	}

	return nil
}

// GetUsage 获取使用情况
func (s *DefaultTenantService) GetUsage(ctx context.Context, tenantID string, period string) (*TenantUsage, error) {
	// 尝试从缓存获取
	if s.config.EnableCache {
		cacheKey := GetUsageCacheKey(tenantID, period)
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
			if usage, ok := cached.(*TenantUsage); ok {
				return usage, nil
			}
		}
	}

	// 从数据库获取
	usage, err := s.repository.GetUsage(ctx, tenantID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	// 缓存结果
	if s.config.EnableCache && usage != nil {
		cacheKey := GetUsageCacheKey(tenantID, period)
		s.cache.Set(ctx, cacheKey, usage, s.config.CacheExpiry)
	}

	return usage, nil
}

// GetUsageHistory 获取使用历史
func (s *DefaultTenantService) GetUsageHistory(ctx context.Context, tenantID string, start, end time.Time) ([]TenantUsage, error) {
	return s.repository.GetUsageHistory(ctx, tenantID, start, end)
}

// GetTenantContext 获取租户上下文
func (s *DefaultTenantService) GetTenantContext(ctx context.Context, tenantID, userID string) (*TenantContext, error) {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 获取当前使用情况
	usage, err := s.GetUsage(ctx, tenantID, "monthly")
	if err != nil {
		// 如果获取使用情况失败，创建空的使用情况
		usage = &TenantUsage{
			TenantID:  tenantID,
			Period:    "monthly",
			Timestamp: time.Now(),
		}
	}

	return &TenantContext{
		TenantID:    tenantID,
		UserID:      userID,
		Settings:    tenant.Settings,
		Limits:      tenant.Limits,
		Usage:       *usage,
		Roles:       []string{}, // 需要从用户服务获取
		Permissions: []string{}, // 需要从权限服务获取
	}, nil
}

// ValidateTenantAccess 验证租户访问权限
func (s *DefaultTenantService) ValidateTenantAccess(ctx context.Context, tenantID, userID string) (bool, error) {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// 检查租户状态
	if !tenant.IsActive() {
		return false, fmt.Errorf("tenant is not active")
	}

	// 这里应该检查用户是否属于该租户
	// 简化实现，实际中需要查询用户-租户关系
	return true, nil
}

// GetTenantDatabase 获取租户数据库
func (s *DefaultTenantService) GetTenantDatabase(ctx context.Context, tenantID string) (string, error) {
	// 根据租户ID生成数据库名称
	return fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenantID, "-", "_")), nil
}

// GetTenantSchema 获取租户模式
func (s *DefaultTenantService) GetTenantSchema(ctx context.Context, tenantID string) (string, error) {
	// 根据租户ID生成模式名称
	return fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenantID, "-", "_")), nil
}

// BackupTenant 备份租户
func (s *DefaultTenantService) BackupTenant(ctx context.Context, tenantID string) (string, error) {
	backupID := uuid.New().String()
	
	// 这里应该实现实际的备份逻辑
	// 简化实现
	s.logger.Info("Tenant backup initiated",
		zap.String("tenant_id", tenantID),
		zap.String("backup_id", backupID))

	// 发布备份事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "backup", "initiated", map[string]interface{}{
			"backup_id": backupID,
		}, "system")
	}

	return backupID, nil
}

// RestoreTenant 恢复租户
func (s *DefaultTenantService) RestoreTenant(ctx context.Context, tenantID, backupID string) error {
	// 这里应该实现实际的恢复逻辑
	// 简化实现
	s.logger.Info("Tenant restore initiated",
		zap.String("tenant_id", tenantID),
		zap.String("backup_id", backupID))

	// 发布恢复事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "restore", "initiated", map[string]interface{}{
			"backup_id": backupID,
		}, "system")
	}

	return nil
}

// MigrateTenant 迁移租户
func (s *DefaultTenantService) MigrateTenant(ctx context.Context, tenantID, targetRegion string) error {
	// 这里应该实现实际的迁移逻辑
	// 简化实现
	s.logger.Info("Tenant migration initiated",
		zap.String("tenant_id", tenantID),
		zap.String("target_region", targetRegion))

	// 发布迁移事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "migration", "initiated", map[string]interface{}{
			"target_region": targetRegion,
		}, "system")
	}

	return nil
}

// HealthCheck 健康检查
func (s *DefaultTenantService) HealthCheck(ctx context.Context) error {
	// 检查数据库连接
	// 检查缓存连接
	// 检查事件发布器
	return nil
}

// 私有方法

// validateCreateRequest 验证创建请求
func (s *DefaultTenantService) validateCreateRequest(req CreateTenantRequest) error {
	if req.Name == "" {
		return fmt.Errorf("tenant name is required")
	}
	if req.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}
	if req.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}
	return nil
}

// updateTenantStatus 更新租户状态
func (s *DefaultTenantService) updateTenantStatus(ctx context.Context, tenantID string, status TenantStatus, action string) error {
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	tenant.Status = status
	tenant.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, tenant); err != nil {
		s.logger.Error("Failed to update tenant status",
			zap.String("tenant_id", tenantID),
			zap.String("status", string(status)),
			zap.Error(err))
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	// 更新缓存
	if s.config.EnableCache {
		s.cacheTenant(ctx, tenant)
	}

	// 发布事件
	if s.config.EnableEvents {
		s.publishTenantEvent(ctx, tenantID, "tenant", action, map[string]interface{}{
			"status": status,
		}, "system")
	}

	s.logger.Info("Tenant status updated",
		zap.String("tenant_id", tenantID),
		zap.String("status", string(status)),
		zap.String("action", action))

	return nil
}

// 缓存相关方法

func (s *DefaultTenantService) cacheTenant(ctx context.Context, tenant *Tenant) {
	if s.cache == nil {
		return
	}
	
	cacheKey := GetTenantCacheKey(tenant.ID)
	s.cache.Set(ctx, cacheKey, tenant, s.config.CacheExpiry)
}

func (s *DefaultTenantService) cacheTenantByDomain(ctx context.Context, domain string, tenant *Tenant) {
	if s.cache == nil {
		return
	}
	
	cacheKey := GetDomainCacheKey(domain)
	s.cache.Set(ctx, cacheKey, tenant, s.config.CacheExpiry)
}

func (s *DefaultTenantService) getTenantFromCache(ctx context.Context, tenantID string) (*Tenant, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}
	
	cacheKey := GetTenantCacheKey(tenantID)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	
	if tenant, ok := cached.(*Tenant); ok {
		return tenant, nil
	}
	
	return nil, fmt.Errorf("invalid cached data")
}

func (s *DefaultTenantService) getTenantFromCacheByDomain(ctx context.Context, domain string) (*Tenant, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}
	
	cacheKey := GetDomainCacheKey(domain)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	
	if tenant, ok := cached.(*Tenant); ok {
		return tenant, nil
	}
	
	return nil, fmt.Errorf("invalid cached data")
}

func (s *DefaultTenantService) clearTenantCache(ctx context.Context, tenantID, domain string) {
	if s.cache == nil {
		return
	}
	
	s.cache.Delete(ctx, GetTenantCacheKey(tenantID))
	s.cache.Delete(ctx, GetDomainCacheKey(domain))
}

// 事件发布方法

func (s *DefaultTenantService) publishTenantEvent(ctx context.Context, tenantID, eventType, action string, data map[string]interface{}, userID string) {
	if s.publisher == nil {
		return
	}
	
	event := TenantEvent{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Type:      eventType,
		Action:    action,
		Data:      data,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	
	if err := s.publisher.PublishEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish tenant event",
			zap.String("tenant_id", tenantID),
			zap.String("event_type", eventType),
			zap.String("action", action),
			zap.Error(err))
	}
}