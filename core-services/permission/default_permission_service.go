package permission

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultPermissionService 默认权限服务实现
type DefaultPermissionService struct {
	repository PermissionRepository
	cache      PermissionCache
	logger     *zap.Logger
	config     DefaultPermissionServiceConfig
	mu         sync.RWMutex
	
	// 内部状态
	policyEvaluators map[PolicyType]PolicyEvaluator
	permissionCache  map[string]*PermissionCheckResponse
	cacheExpiry      map[string]time.Time
}

// DefaultPermissionServiceConfig 默认权限服务配置
type DefaultPermissionServiceConfig struct {
	// 缓存配置
	CacheEnabled         bool          `json:"cache_enabled"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	UserPermissionsTTL   time.Duration `json:"user_permissions_ttl"`
	UserRolesTTL         time.Duration `json:"user_roles_ttl"`
	PermissionCheckTTL   time.Duration `json:"permission_check_ttl"`
	
	// 权限检查配置
	DefaultCheckMode     CheckMode `json:"default_check_mode"`
	EnableInheritance    bool      `json:"enable_inheritance"`
	EnablePolicyEngine   bool      `json:"enable_policy_engine"`
	MaxInheritanceDepth  int       `json:"max_inheritance_depth"`
	
	// 性能配置
	BatchSize            int           `json:"batch_size"`
	MaxConcurrentChecks  int           `json:"max_concurrent_checks"`
	CheckTimeout         time.Duration `json:"check_timeout"`
	
	// 安全配置
	EnableAuditLog       bool     `json:"enable_audit_log"`
	AuditAllChecks       bool     `json:"audit_all_checks"`
	AuditFailedChecks    bool     `json:"audit_failed_checks"`
	SensitiveResources   []string `json:"sensitive_resources"`
	
	// 策略配置
	PolicyEvaluationMode string        `json:"policy_evaluation_mode"`
	PolicyCacheEnabled   bool          `json:"policy_cache_enabled"`
	PolicyCacheTTL       time.Duration `json:"policy_cache_ttl"`
	
	// 角色配置
	MaxRoleDepth         int  `json:"max_role_depth"`
	EnableRoleHierarchy  bool `json:"enable_role_hierarchy"`
	
	// 监控配置
	EnableMetrics        bool `json:"enable_metrics"`
	MetricsInterval      time.Duration `json:"metrics_interval"`
}

// PolicyEvaluator 策略评估器接口
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error)
}

// NewDefaultPermissionService 创建默认权限服务
func NewDefaultPermissionService(
	repository PermissionRepository,
	cache PermissionCache,
	logger *zap.Logger,
	config DefaultPermissionServiceConfig,
) *DefaultPermissionService {
	// 设置默认配置
	if config.CacheTTL == 0 {
		config.CacheTTL = 15 * time.Minute
	}
	if config.UserPermissionsTTL == 0 {
		config.UserPermissionsTTL = 30 * time.Minute
	}
	if config.UserRolesTTL == 0 {
		config.UserRolesTTL = 30 * time.Minute
	}
	if config.PermissionCheckTTL == 0 {
		config.PermissionCheckTTL = 5 * time.Minute
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.MaxConcurrentChecks == 0 {
		config.MaxConcurrentChecks = 10
	}
	if config.CheckTimeout == 0 {
		config.CheckTimeout = 30 * time.Second
	}
	if config.MaxInheritanceDepth == 0 {
		config.MaxInheritanceDepth = 10
	}
	if config.MaxRoleDepth == 0 {
		config.MaxRoleDepth = 10
	}

	service := &DefaultPermissionService{
		repository:       repository,
		cache:           cache,
		logger:          logger,
		config:          config,
		policyEvaluators: make(map[PolicyType]PolicyEvaluator),
		permissionCache:  make(map[string]*PermissionCheckResponse),
		cacheExpiry:     make(map[string]time.Time),
	}

	// 注册策略评估器
	service.registerPolicyEvaluators()

	return service
}

// CheckPermission 检查权限
func (s *DefaultPermissionService) CheckPermission(ctx context.Context, request *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	// 验证请求
	if err := s.validatePermissionCheckRequest(request); err != nil {
		return &PermissionCheckResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("Invalid request: %v", err),
			Effect:  PermissionEffectDeny,
		}, nil
	}

	// 检查缓存
	if s.config.CacheEnabled {
		cacheKey := CreatePermissionCheckKey(request.UserID, request.TenantID, request.Resource, request.Action, request.ResourceID)
		if cached, err := s.getFromCache(cacheKey); err == nil && cached != nil {
			s.logger.Debug("Permission check cache hit", zap.String("cache_key", cacheKey))
			return cached, nil
		}
	}

	// 执行权限检查
	response, err := s.doPermissionCheck(ctx, request)
	if err != nil {
		s.logger.Error("Permission check failed", zap.Error(err))
		return &PermissionCheckResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("Check failed: %v", err),
			Effect:  PermissionEffectDeny,
		}, nil
	}

	// 缓存结果
	if s.config.CacheEnabled {
		cacheKey := CreatePermissionCheckKey(request.UserID, request.TenantID, request.Resource, request.Action, request.ResourceID)
		s.setToCache(cacheKey, response)
	}

	// 记录审计日志
	if s.config.EnableAuditLog && (s.config.AuditAllChecks || (s.config.AuditFailedChecks && !response.Allowed)) {
		s.logPermissionCheck(ctx, request, response)
	}

	return response, nil
}

// CheckPermissions 批量检查权限
func (s *DefaultPermissionService) CheckPermissions(ctx context.Context, requests []*PermissionCheckRequest) ([]*PermissionCheckResponse, error) {
	if len(requests) == 0 {
		return []*PermissionCheckResponse{}, nil
	}

	responses := make([]*PermissionCheckResponse, len(requests))
	
	// 使用并发检查
	if len(requests) > 1 && s.config.MaxConcurrentChecks > 1 {
		return s.checkPermissionsConcurrent(ctx, requests)
	}

	// 顺序检查
	for i, request := range requests {
		response, err := s.CheckPermission(ctx, request)
		if err != nil {
			responses[i] = &PermissionCheckResponse{
				Allowed: false,
				Reason:  fmt.Sprintf("Check failed: %v", err),
				Effect:  PermissionEffectDeny,
			}
		} else {
			responses[i] = response
		}
	}

	return responses, nil
}

// CreateRole 创建角色
func (s *DefaultPermissionService) CreateRole(ctx context.Context, request *CreateRoleRequest) (*Role, error) {
	// 验证请求
	if err := s.validateCreateRoleRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 检查角色名称是否已存在
	existing, err := s.repository.GetRoleByName(ctx, request.Name, request.TenantID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", request.Name)
	}

	// 创建角色
	role := &Role{
		ID:          GenerateRoleID(),
		Name:        request.Name,
		Code:        request.Code,
		Description: request.Description,
		Type:        request.Type,
		Level:       request.Level,
		ParentID:    request.ParentID,
		IsSystem:    false,
		IsActive:    true,
		Metadata:    request.Metadata,
		TenantID:    request.TenantID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存角色
	if err := s.repository.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// 分配权限
	if len(request.Permissions) > 0 {
		for _, permissionID := range request.Permissions {
			if err := s.repository.AssignPermissionToRole(ctx, role.ID, permissionID); err != nil {
				s.logger.Warn("Failed to assign permission to role",
					zap.String("role_id", role.ID),
					zap.String("permission_id", permissionID),
					zap.Error(err))
			}
		}
	}

	// 清除相关缓存
	if s.config.CacheEnabled {
		s.invalidateRoleCache(ctx, role.ID, request.TenantID)
	}

	s.logger.Info("Role created successfully",
		zap.String("role_id", role.ID),
		zap.String("role_name", role.Name),
		zap.String("tenant_id", request.TenantID))

	return role, nil
}

// GetRole 获取角色
func (s *DefaultPermissionService) GetRole(ctx context.Context, roleID string) (*Role, error) {
	role, err := s.repository.GetRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// 加载权限
	permissions, err := s.repository.GetRolePermissions(ctx, roleID)
	if err != nil {
		s.logger.Warn("Failed to load role permissions", zap.String("role_id", roleID), zap.Error(err))
	} else {
		role.Permissions = permissions
	}

	return role, nil
}

// GetRoleByName 根据名称获取角色
func (s *DefaultPermissionService) GetRoleByName(ctx context.Context, name string, tenantID string) (*Role, error) {
	return s.repository.GetRoleByName(ctx, name, tenantID)
}

// UpdateRole 更新角色
func (s *DefaultPermissionService) UpdateRole(ctx context.Context, roleID string, request *UpdateRoleRequest) (*Role, error) {
	// 获取现有角色
	role, err := s.repository.GetRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// 更新字段
	if request.Name != nil {
		role.Name = *request.Name
	}
	if request.Description != nil {
		role.Description = *request.Description
	}
	if request.Type != nil {
		role.Type = *request.Type
	}
	if request.Level != nil {
		role.Level = *request.Level
	}
	if request.ParentID != nil {
		role.ParentID = request.ParentID
	}
	if request.IsActive != nil {
		role.IsActive = *request.IsActive
	}
	if request.Metadata != nil {
		role.Metadata = request.Metadata
	}

	role.UpdatedAt = time.Now()

	// 保存更新
	if err := s.repository.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// 清除相关缓存
	if s.config.CacheEnabled {
		s.invalidateRoleCache(ctx, roleID, role.TenantID)
	}

	s.logger.Info("Role updated successfully", zap.String("role_id", roleID))

	return role, nil
}

// DeleteRole 删除角色
func (s *DefaultPermissionService) DeleteRole(ctx context.Context, roleID string) error {
	// 获取角色信息
	role, err := s.repository.GetRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查是否为系统角色
	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}

	// 删除角色
	if err := s.repository.DeleteRole(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// 清除相关缓存
	if s.config.CacheEnabled {
		s.invalidateRoleCache(ctx, roleID, role.TenantID)
	}

	s.logger.Info("Role deleted successfully", zap.String("role_id", roleID))

	return nil
}

// ListRoles 列出角色
func (s *DefaultPermissionService) ListRoles(ctx context.Context, filter *RoleFilter) (*ListRolesResponse, error) {
	roles, total, err := s.repository.ListRoles(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// 计算分页信息
	totalPages := int(total) / filter.Pagination.PageSize
	if int(total)%filter.Pagination.PageSize > 0 {
		totalPages++
	}

	return &ListRolesResponse{
		Roles: roles,
		Pagination: PaginationResponse{
			Page:       filter.Pagination.Page,
			PageSize:   filter.Pagination.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// CreatePermission 创建权限
func (s *DefaultPermissionService) CreatePermission(ctx context.Context, request *CreatePermissionRequest) (*Permission, error) {
	// 验证请求
	if err := s.validateCreatePermissionRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 创建权限
	permission := &Permission{
		ID:          GeneratePermissionID(),
		Name:        request.Name,
		Code:        request.Code,
		Description: request.Description,
		Category:    request.Category,
		Resource:    request.Resource,
		Action:      request.Action,
		Effect:      request.Effect,
		Conditions:  request.Conditions,
		Metadata:    request.Metadata,
		TenantID:    request.TenantID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存权限
	if err := s.repository.CreatePermission(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	s.logger.Info("Permission created successfully",
		zap.String("permission_id", permission.ID),
		zap.String("permission_name", permission.Name),
		zap.String("tenant_id", request.TenantID))

	return permission, nil
}

// GetPermission 获取权限
func (s *DefaultPermissionService) GetPermission(ctx context.Context, permissionID string) (*Permission, error) {
	return s.repository.GetPermission(ctx, permissionID)
}

// UpdatePermission 更新权限
func (s *DefaultPermissionService) UpdatePermission(ctx context.Context, permissionID string, request *UpdatePermissionRequest) (*Permission, error) {
	// 获取现有权限
	permission, err := s.repository.GetPermission(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	// 更新字段
	if request.Name != nil {
		permission.Name = *request.Name
	}
	if request.Description != nil {
		permission.Description = *request.Description
	}
	if request.Category != nil {
		permission.Category = *request.Category
	}
	if request.Effect != nil {
		permission.Effect = *request.Effect
	}
	if request.Conditions != nil {
		permission.Conditions = request.Conditions
	}
	if request.Metadata != nil {
		permission.Metadata = request.Metadata
	}

	permission.UpdatedAt = time.Now()

	// 保存更新
	if err := s.repository.UpdatePermission(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	s.logger.Info("Permission updated successfully", zap.String("permission_id", permissionID))

	return permission, nil
}

// DeletePermission 删除权限
func (s *DefaultPermissionService) DeletePermission(ctx context.Context, permissionID string) error {
	if err := s.repository.DeletePermission(ctx, permissionID); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	s.logger.Info("Permission deleted successfully", zap.String("permission_id", permissionID))

	return nil
}

// ListPermissions 列出权限
func (s *DefaultPermissionService) ListPermissions(ctx context.Context, filter *PermissionFilter) (*ListPermissionsResponse, error) {
	permissions, total, err := s.repository.ListPermissions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	// 计算分页信息
	totalPages := int(total) / filter.Pagination.PageSize
	if int(total)%filter.Pagination.PageSize > 0 {
		totalPages++
	}

	return &ListPermissionsResponse{
		Permissions: permissions,
		Pagination: PaginationResponse{
			Page:       filter.Pagination.Page,
			PageSize:   filter.Pagination.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// AssignPermissionToRole 分配权限给角色
func (s *DefaultPermissionService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	if err := s.repository.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	// 清除相关缓存
	if s.config.CacheEnabled {
		role, _ := s.repository.GetRole(ctx, roleID)
		if role != nil {
			s.invalidateRoleCache(ctx, roleID, role.TenantID)
		}
	}

	s.logger.Info("Permission assigned to role successfully",
		zap.String("role_id", roleID),
		zap.String("permission_id", permissionID))

	return nil
}

// RevokePermissionFromRole 从角色撤销权限
func (s *DefaultPermissionService) RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	if err := s.repository.RevokePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to revoke permission from role: %w", err)
	}

	// 清除相关缓存
	if s.config.CacheEnabled {
		role, _ := s.repository.GetRole(ctx, roleID)
		if role != nil {
			s.invalidateRoleCache(ctx, roleID, role.TenantID)
		}
	}

	s.logger.Info("Permission revoked from role successfully",
		zap.String("role_id", roleID),
		zap.String("permission_id", permissionID))

	return nil
}

// GetRolePermissions 获取角色权限
func (s *DefaultPermissionService) GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error) {
	return s.repository.GetRolePermissions(ctx, roleID)
}

// AssignRoleToUser 分配角色给用户
func (s *DefaultPermissionService) AssignRoleToUser(ctx context.Context, userID, roleID string, tenantID string) error {
	if err := s.repository.AssignRoleToUser(ctx, userID, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	// 清除用户缓存
	if s.config.CacheEnabled {
		s.invalidateUserCache(ctx, userID, tenantID)
	}

	s.logger.Info("Role assigned to user successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID),
		zap.String("tenant_id", tenantID))

	return nil
}

// RevokeRoleFromUser 从用户撤销角色
func (s *DefaultPermissionService) RevokeRoleFromUser(ctx context.Context, userID, roleID string, tenantID string) error {
	if err := s.repository.RevokeRoleFromUser(ctx, userID, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to revoke role from user: %w", err)
	}

	// 清除用户缓存
	if s.config.CacheEnabled {
		s.invalidateUserCache(ctx, userID, tenantID)
	}

	s.logger.Info("Role revoked from user successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetUserRoles 获取用户角色
func (s *DefaultPermissionService) GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*Role, error) {
	// 检查缓存
	if s.config.CacheEnabled && s.cache != nil {
		if cached, err := s.cache.GetUserRoles(ctx, userID, tenantID); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 从数据库获取
	roles, err := s.repository.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 缓存结果
	if s.config.CacheEnabled && s.cache != nil {
		s.cache.SetUserRoles(ctx, userID, tenantID, roles, s.config.UserRolesTTL)
	}

	return roles, nil
}

// GetUserPermissions 获取用户权限
func (s *DefaultPermissionService) GetUserPermissions(ctx context.Context, userID string, tenantID string) ([]*Permission, error) {
	// 检查缓存
	if s.config.CacheEnabled && s.cache != nil {
		if cached, err := s.cache.GetUserPermissions(ctx, userID, tenantID); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 获取用户角色
	roles, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 收集所有权限
	permissionMap := make(map[string]*Permission)
	for _, role := range roles {
		rolePermissions, err := s.repository.GetRolePermissions(ctx, role.ID)
		if err != nil {
			s.logger.Warn("Failed to get role permissions",
				zap.String("role_id", role.ID),
				zap.Error(err))
			continue
		}

		for _, permission := range rolePermissions {
			permissionMap[permission.ID] = permission
		}
	}

	// 转换为切片
	permissions := make([]*Permission, 0, len(permissionMap))
	for _, permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	// 缓存结果
	if s.config.CacheEnabled && s.cache != nil {
		s.cache.SetUserPermissions(ctx, userID, tenantID, permissions, s.config.UserPermissionsTTL)
	}

	return permissions, nil
}

// GrantResourcePermission 授予资源权限
func (s *DefaultPermissionService) GrantResourcePermission(ctx context.Context, request *GrantResourcePermissionRequest) error {
	// 验证请求
	if err := s.validateGrantResourcePermissionRequest(request); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// 创建资源权限
	resourcePermission := &ResourcePermission{
		ID:           GenerateResourcePermissionID(),
		ResourceID:   request.ResourceID,
		ResourceType: request.ResourceType,
		SubjectID:    request.SubjectID,
		SubjectType:  request.SubjectType,
		PermissionID: request.PermissionID,
		Effect:       request.Effect,
		Conditions:   request.Conditions,
		ExpiresAt:    request.ExpiresAt,
		TenantID:     request.TenantID,
		CreatedAt:    time.Now(),
	}

	// 保存资源权限
	if err := s.repository.CreateResourcePermission(ctx, resourcePermission); err != nil {
		return fmt.Errorf("failed to grant resource permission: %w", err)
	}

	s.logger.Info("Resource permission granted successfully",
		zap.String("resource_id", request.ResourceID),
		zap.String("subject_id", request.SubjectID),
		zap.String("permission_id", request.PermissionID))

	return nil
}

// RevokeResourcePermission 撤销资源权限
func (s *DefaultPermissionService) RevokeResourcePermission(ctx context.Context, request *RevokeResourcePermissionRequest) error {
	if err := s.repository.DeleteResourcePermission(ctx,
		request.ResourceID,
		request.ResourceType,
		request.SubjectID,
		request.SubjectType,
		request.PermissionID); err != nil {
		return fmt.Errorf("failed to revoke resource permission: %w", err)
	}

	s.logger.Info("Resource permission revoked successfully",
		zap.String("resource_id", request.ResourceID),
		zap.String("subject_id", request.SubjectID),
		zap.String("permission_id", request.PermissionID))

	return nil
}

// GetResourcePermissions 获取资源权限
func (s *DefaultPermissionService) GetResourcePermissions(ctx context.Context, resourceID, resourceType string) ([]*ResourcePermission, error) {
	return s.repository.GetResourcePermissions(ctx, resourceID, resourceType)
}

// SetPermissionInheritance 设置权限继承
func (s *DefaultPermissionService) SetPermissionInheritance(ctx context.Context, request *PermissionInheritanceRequest) error {
	// 验证请求
	if err := s.validatePermissionInheritanceRequest(request); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// 创建权限继承
	inheritance := &PermissionInheritance{
		ID:           GenerateInheritanceID(),
		ResourceID:   request.ResourceID,
		ResourceType: request.ResourceType,
		ParentID:     request.ParentID,
		ParentType:   request.ParentType,
		InheritType:  request.InheritType,
		IsActive:     request.IsActive,
		TenantID:     request.TenantID,
		CreatedAt:    time.Now(),
	}

	// 保存权限继承
	if err := s.repository.CreatePermissionInheritance(ctx, inheritance); err != nil {
		return fmt.Errorf("failed to set permission inheritance: %w", err)
	}

	s.logger.Info("Permission inheritance set successfully",
		zap.String("resource_id", request.ResourceID),
		zap.String("parent_id", request.ParentID))

	return nil
}

// GetPermissionInheritance 获取权限继承
func (s *DefaultPermissionService) GetPermissionInheritance(ctx context.Context, resourceID, resourceType string) (*PermissionInheritance, error) {
	return s.repository.GetPermissionInheritance(ctx, resourceID, resourceType)
}

// CreatePolicy 创建策略
func (s *DefaultPermissionService) CreatePolicy(ctx context.Context, request *CreatePolicyRequest) (*Policy, error) {
	// 验证请求
	if err := s.validateCreatePolicyRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 创建策略
	policy := &Policy{
		ID:          GeneratePolicyID(),
		Name:        request.Name,
		Description: request.Description,
		Type:        request.Type,
		Rules:       request.Rules,
		Effect:      request.Effect,
		Priority:    request.Priority,
		IsActive:    request.IsActive,
		Metadata:    request.Metadata,
		TenantID:    request.TenantID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存策略
	if err := s.repository.CreatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	s.logger.Info("Policy created successfully",
		zap.String("policy_id", policy.ID),
		zap.String("policy_name", policy.Name))

	return policy, nil
}

// GetPolicy 获取策略
func (s *DefaultPermissionService) GetPolicy(ctx context.Context, policyID string) (*Policy, error) {
	return s.repository.GetPolicy(ctx, policyID)
}

// UpdatePolicy 更新策略
func (s *DefaultPermissionService) UpdatePolicy(ctx context.Context, policyID string, request *UpdatePolicyRequest) (*Policy, error) {
	// 获取现有策略
	policy, err := s.repository.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// 更新字段
	if request.Name != nil {
		policy.Name = *request.Name
	}
	if request.Description != nil {
		policy.Description = *request.Description
	}
	if request.Rules != nil {
		policy.Rules = request.Rules
	}
	if request.Effect != nil {
		policy.Effect = *request.Effect
	}
	if request.Priority != nil {
		policy.Priority = *request.Priority
	}
	if request.IsActive != nil {
		policy.IsActive = *request.IsActive
	}
	if request.Metadata != nil {
		policy.Metadata = request.Metadata
	}

	policy.UpdatedAt = time.Now()

	// 保存更新
	if err := s.repository.UpdatePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	s.logger.Info("Policy updated successfully", zap.String("policy_id", policyID))

	return policy, nil
}

// DeletePolicy 删除策略
func (s *DefaultPermissionService) DeletePolicy(ctx context.Context, policyID string) error {
	if err := s.repository.DeletePolicy(ctx, policyID); err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	s.logger.Info("Policy deleted successfully", zap.String("policy_id", policyID))

	return nil
}

// ListPolicies 列出策略
func (s *DefaultPermissionService) ListPolicies(ctx context.Context, filter *PolicyFilter) (*ListPoliciesResponse, error) {
	policies, total, err := s.repository.ListPolicies(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	// 计算分页信息
	totalPages := int(total) / filter.Pagination.PageSize
	if int(total)%filter.Pagination.PageSize > 0 {
		totalPages++
	}

	return &ListPoliciesResponse{
		Policies: policies,
		Pagination: PaginationResponse{
			Page:       filter.Pagination.Page,
			PageSize:   filter.Pagination.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// EvaluatePolicy 评估策略
func (s *DefaultPermissionService) EvaluatePolicy(ctx context.Context, request *PolicyEvaluationRequest) (*PolicyEvaluationResponse, error) {
	// 获取策略
	var policies []*Policy
	if len(request.PolicyIDs) > 0 {
		for _, policyID := range request.PolicyIDs {
			policy, err := s.repository.GetPolicy(ctx, policyID)
			if err != nil {
				s.logger.Warn("Failed to get policy", zap.String("policy_id", policyID), zap.Error(err))
				continue
			}
			if policy.IsActive {
				policies = append(policies, policy)
			}
		}
	} else {
		// 获取所有活跃策略
		filter := &PolicyFilter{
			IsActive: &[]bool{true}[0],
			TenantID: request.TenantID,
			Pagination: PaginationRequest{
				Page:     1,
				PageSize: 1000,
			},
		}
		response, err := s.ListPolicies(ctx, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get policies: %w", err)
		}
		policies = response.Policies
	}

	// 评估策略
	for _, policy := range policies {
		if evaluator, exists := s.policyEvaluators[policy.Type]; exists {
			return evaluator.Evaluate(ctx, request, []*Policy{policy})
		}
	}

	// 默认拒绝
	return &PolicyEvaluationResponse{
		Allowed: false,
		Effect:  PermissionEffectDeny,
		EvaluationLog: []string{"No applicable policy found"},
	}, nil
}

// InvalidateCache 清除缓存
func (s *DefaultPermissionService) InvalidateCache(ctx context.Context, userID, tenantID string) error {
	if s.config.CacheEnabled && s.cache != nil {
		return s.cache.ClearUserCache(ctx, userID, tenantID)
	}
	return nil
}

// InvalidateAllCache 清除所有缓存
func (s *DefaultPermissionService) InvalidateAllCache(ctx context.Context) error {
	if s.config.CacheEnabled && s.cache != nil {
		return s.cache.Clear(ctx)
	}
	return nil
}

// GetPermissionAuditLog 获取权限审计日志
func (s *DefaultPermissionService) GetPermissionAuditLog(ctx context.Context, filter *PermissionAuditFilter) (*PermissionAuditResponse, error) {
	auditLogs, total, err := s.repository.GetPermissionAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission audit logs: %w", err)
	}

	// 计算分页信息
	totalPages := int(total) / filter.Pagination.PageSize
	if int(total)%filter.Pagination.PageSize > 0 {
		totalPages++
	}

	return &PermissionAuditResponse{
		AuditLogs: auditLogs,
		Pagination: PaginationResponse{
			Page:       filter.Pagination.Page,
			PageSize:   filter.Pagination.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// HealthCheck 健康检查
func (s *DefaultPermissionService) HealthCheck(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Healthy:   true,
		Status:    "healthy",
		Checks:    make(map[string]string),
		Timestamp: time.Now(),
	}

	// 检查数据库
	if err := s.repository.HealthCheck(ctx); err != nil {
		status.Healthy = false
		status.Status = "unhealthy"
		status.Checks["database"] = fmt.Sprintf("failed: %v", err)
	} else {
		status.Checks["database"] = "ok"
	}

	// 检查缓存
	if s.cache != nil {
		if err := s.cache.HealthCheck(ctx); err != nil {
			status.Checks["cache"] = fmt.Sprintf("failed: %v", err)
		} else {
			status.Checks["cache"] = "ok"
		}
	}

	return status
}

// 私有方法

// doPermissionCheck 执行权限检查
func (s *DefaultPermissionService) doPermissionCheck(ctx context.Context, request *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	// 获取用户权限
	permissions, err := s.getUserEffectivePermissions(ctx, request.UserID, request.TenantID, request.ResourceID, request.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 检查权限
	for _, permission := range permissions {
		if s.matchesPermission(permission, request) {
			// 检查条件
			if s.evaluateConditions(permission.Conditions, request.Context) {
				return &PermissionCheckResponse{
					Allowed:     permission.Effect == PermissionEffectAllow,
					Reason:      fmt.Sprintf("Permission %s matched", permission.Code),
					Effect:      permission.Effect,
					Permissions: []*Permission{permission},
				}, nil
			}
		}
	}

	// 如果启用策略引擎，评估策略
	if s.config.EnablePolicyEngine {
		policyRequest := &PolicyEvaluationRequest{
			UserID:     request.UserID,
			TenantID:   request.TenantID,
			Resource:   request.Resource,
			Action:     request.Action,
			ResourceID: request.ResourceID,
			Context:    request.Context,
		}

		policyResponse, err := s.EvaluatePolicy(ctx, policyRequest)
		if err == nil && policyResponse.Allowed {
			return &PermissionCheckResponse{
				Allowed:  true,
				Reason:   "Policy evaluation allowed",
				Effect:   policyResponse.Effect,
				Policies: []*Policy{}, // 这里应该返回匹配的策略
			}, nil
		}
	}

	// 默认拒绝
	return &PermissionCheckResponse{
		Allowed: false,
		Reason:  "No matching permission found",
		Effect:  PermissionEffectDeny,
	}, nil
}

// getUserEffectivePermissions 获取用户有效权限
func (s *DefaultPermissionService) getUserEffectivePermissions(ctx context.Context, userID, tenantID string, resourceID *string, resource string) ([]*Permission, error) {
	// 获取用户直接权限
	userPermissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	effectivePermissions := make([]*Permission, 0, len(userPermissions))
	effectivePermissions = append(effectivePermissions, userPermissions...)

	// 如果指定了资源ID，获取资源特定权限
	if resourceID != nil {
		resourcePermissions, err := s.getResourceSpecificPermissions(ctx, *resourceID, resource, userID, tenantID)
		if err != nil {
			s.logger.Warn("Failed to get resource specific permissions", zap.Error(err))
		} else {
			effectivePermissions = append(effectivePermissions, resourcePermissions...)
		}
	}

	// 如果启用继承，获取继承权限
	if s.config.EnableInheritance && resourceID != nil {
		inheritedPermissions, err := s.getInheritedPermissions(ctx, *resourceID, resource, userID, tenantID, 0)
		if err != nil {
			s.logger.Warn("Failed to get inherited permissions", zap.Error(err))
		} else {
			effectivePermissions = append(effectivePermissions, inheritedPermissions...)
		}
	}

	return effectivePermissions, nil
}

// getResourceSpecificPermissions 获取资源特定权限
func (s *DefaultPermissionService) getResourceSpecificPermissions(ctx context.Context, resourceID, resourceType, userID, tenantID string) ([]*Permission, error) {
	resourcePermissions, err := s.repository.GetResourcePermissions(ctx, resourceID, resourceType)
	if err != nil {
		return nil, err
	}

	var permissions []*Permission
	for _, rp := range resourcePermissions {
		// 检查是否适用于当前用户
		if s.isResourcePermissionApplicable(rp, userID, tenantID) {
			if rp.Permission != nil {
				permissions = append(permissions, rp.Permission)
			} else {
				// 加载权限详情
				permission, err := s.repository.GetPermission(ctx, rp.PermissionID)
				if err != nil {
					s.logger.Warn("Failed to load permission", zap.String("permission_id", rp.PermissionID), zap.Error(err))
					continue
				}
				permissions = append(permissions, permission)
			}
		}
	}

	return permissions, nil
}

// getInheritedPermissions 获取继承权限
func (s *DefaultPermissionService) getInheritedPermissions(ctx context.Context, resourceID, resourceType, userID, tenantID string, depth int) ([]*Permission, error) {
	if depth >= s.config.MaxInheritanceDepth {
		return nil, nil
	}

	inheritance, err := s.repository.GetPermissionInheritance(ctx, resourceID, resourceType)
	if err != nil || inheritance == nil || !inheritance.IsActive {
		return nil, nil
	}

	// 获取父资源权限
	parentPermissions, err := s.getResourceSpecificPermissions(ctx, inheritance.ParentID, inheritance.ParentType, userID, tenantID)
	if err != nil {
		return nil, err
	}

	// 递归获取父资源的继承权限
	inheritedPermissions, err := s.getInheritedPermissions(ctx, inheritance.ParentID, inheritance.ParentType, userID, tenantID, depth+1)
	if err != nil {
		s.logger.Warn("Failed to get inherited permissions", zap.Error(err))
	} else {
		parentPermissions = append(parentPermissions, inheritedPermissions...)
	}

	return parentPermissions, nil
}

// isResourcePermissionApplicable 检查资源权限是否适用
func (s *DefaultPermissionService) isResourcePermissionApplicable(rp *ResourcePermission, userID, tenantID string) bool {
	// 检查租户
	if rp.TenantID != tenantID {
		return false
	}

	// 检查过期时间
	if rp.ExpiresAt != nil && rp.ExpiresAt.Before(time.Now()) {
		return false
	}

	// 检查主体
	switch rp.SubjectType {
	case SubjectTypeUser:
		return rp.SubjectID == userID
	case SubjectTypeRole:
		// 检查用户是否有该角色
		userRoles, err := s.repository.GetUserRoles(context.Background(), userID, tenantID)
		if err != nil {
			return false
		}
		for _, role := range userRoles {
			if role.ID == rp.SubjectID {
				return true
			}
		}
		return false
	case SubjectTypeGroup:
		// 这里需要实现组成员检查逻辑
		// 暂时返回false
		return false
	default:
		return false
	}
}

// matchesPermission 检查权限是否匹配
func (s *DefaultPermissionService) matchesPermission(permission *Permission, request *PermissionCheckRequest) bool {
	// 检查资源
	if !s.matchesResource(permission.Resource, request.Resource) {
		return false
	}

	// 检查动作
	if !s.matchesAction(permission.Action, request.Action) {
		return false
	}

	return true
}

// matchesResource 检查资源是否匹配
func (s *DefaultPermissionService) matchesResource(permissionResource, requestResource string) bool {
	// 支持通配符匹配
	if permissionResource == "*" {
		return true
	}

	// 精确匹配
	if permissionResource == requestResource {
		return true
	}

	// 前缀匹配
	if strings.HasSuffix(permissionResource, "*") {
		prefix := strings.TrimSuffix(permissionResource, "*")
		return strings.HasPrefix(requestResource, prefix)
	}

	return false
}

// matchesAction 检查动作是否匹配
func (s *DefaultPermissionService) matchesAction(permissionAction, requestAction string) bool {
	// 支持通配符匹配
	if permissionAction == "*" {
		return true
	}

	// 精确匹配
	if permissionAction == requestAction {
		return true
	}

	// 前缀匹配
	if strings.HasSuffix(permissionAction, "*") {
		prefix := strings.TrimSuffix(permissionAction, "*")
		return strings.HasPrefix(requestAction, prefix)
	}

	return false
}

// evaluateConditions 评估条件
func (s *DefaultPermissionService) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	// 简化的条件评估逻辑
	for key, expectedValue := range conditions {
		if contextValue, exists := context[key]; exists {
			if contextValue != expectedValue {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// checkPermissionsConcurrent 并发检查权限
func (s *DefaultPermissionService) checkPermissionsConcurrent(ctx context.Context, requests []*PermissionCheckRequest) ([]*PermissionCheckResponse, error) {
	responses := make([]*PermissionCheckResponse, len(requests))
	
	// 使用信号量限制并发数
	semaphore := make(chan struct{}, s.config.MaxConcurrentChecks)
	var wg sync.WaitGroup
	
	for i, request := range requests {
		wg.Add(1)
		go func(index int, req *PermissionCheckRequest) {
			defer wg.Done()
			
			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			// 执行权限检查
			response, err := s.CheckPermission(ctx, req)
			if err != nil {
				responses[index] = &PermissionCheckResponse{
					Allowed: false,
					Reason:  fmt.Sprintf("Check failed: %v", err),
					Effect:  PermissionEffectDeny,
				}
			} else {
				responses[index] = response
			}
		}(i, request)
	}
	
	wg.Wait()
	return responses, nil
}

// 缓存相关方法

// getFromCache 从缓存获取
func (s *DefaultPermissionService) getFromCache(key string) (*PermissionCheckResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if response, exists := s.permissionCache[key]; exists {
		if expiry, exists := s.cacheExpiry[key]; exists && time.Now().Before(expiry) {
			return response, nil
		}
		// 过期，删除
		delete(s.permissionCache, key)
		delete(s.cacheExpiry, key)
	}
	
	return nil, fmt.Errorf("not found")
}

// setToCache 设置到缓存
func (s *DefaultPermissionService) setToCache(key string, response *PermissionCheckResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.permissionCache[key] = response
	s.cacheExpiry[key] = time.Now().Add(s.config.PermissionCheckTTL)
}

// invalidateUserCache 清除用户缓存
func (s *DefaultPermissionService) invalidateUserCache(ctx context.Context, userID, tenantID string) {
	if s.cache != nil {
		s.cache.ClearUserCache(ctx, userID, tenantID)
	}
}

// invalidateRoleCache 清除角色缓存
func (s *DefaultPermissionService) invalidateRoleCache(ctx context.Context, roleID, tenantID string) {
	// 这里需要清除所有相关用户的缓存
	// 简化实现，清除所有缓存
	if s.cache != nil {
		s.cache.Clear(ctx)
	}
}

// 审计相关方法

// logPermissionCheck 记录权限检查审计日志
func (s *DefaultPermissionService) logPermissionCheck(ctx context.Context, request *PermissionCheckRequest, response *PermissionCheckResponse) {
	auditLog := &PermissionAuditLog{
		ID:         GenerateAuditLogID(),
		UserID:     request.UserID,
		TenantID:   request.TenantID,
		Resource:   request.Resource,
		Action:     request.Action,
		ResourceID: request.ResourceID,
		Effect:     response.Effect,
		Allowed:    response.Allowed,
		Reason:     response.Reason,
		Context:    request.Context,
		CreatedAt:  time.Now(),
	}

	// 异步记录审计日志
	go func() {
		if err := s.repository.CreatePermissionAuditLog(context.Background(), auditLog); err != nil {
			s.logger.Error("Failed to create permission audit log", zap.Error(err))
		}
	}()
}

// 策略评估器注册

// registerPolicyEvaluators 注册策略评估器
func (s *DefaultPermissionService) registerPolicyEvaluators() {
	// 注册RBAC评估器
	s.policyEvaluators[PolicyTypeRBAC] = &RBACPolicyEvaluator{
		service: s,
		logger:  s.logger,
	}

	// 注册ABAC评估器
	s.policyEvaluators[PolicyTypeABAC] = &ABACPolicyEvaluator{
		service: s,
		logger:  s.logger,
	}

	// 注册ACL评估器
	s.policyEvaluators[PolicyTypeACL] = &ACLPolicyEvaluator{
		service: s,
		logger:  s.logger,
	}
}

// 验证方法

// validatePermissionCheckRequest 验证权限检查请求
func (s *DefaultPermissionService) validatePermissionCheckRequest(request *PermissionCheckRequest) error {
	if request.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if request.Resource == "" {
		return fmt.Errorf("resource is required")
	}
	if request.Action == "" {
		return fmt.Errorf("action is required")
	}
	return nil
}

// validateCreateRoleRequest 验证创建角色请求
func (s *DefaultPermissionService) validateCreateRoleRequest(request *CreateRoleRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if request.Code == "" {
		return fmt.Errorf("code is required")
	}
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !ValidateRoleType(request.Type) {
		return fmt.Errorf("invalid role type")
	}
	return nil
}

// validateCreatePermissionRequest 验证创建权限请求
func (s *DefaultPermissionService) validateCreatePermissionRequest(request *CreatePermissionRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if request.Code == "" {
		return fmt.Errorf("code is required")
	}
	if request.Resource == "" {
		return fmt.Errorf("resource is required")
	}
	if request.Action == "" {
		return fmt.Errorf("action is required")
	}
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !ValidatePermissionEffect(request.Effect) {
		return fmt.Errorf("invalid permission effect")
	}
	return nil
}

// validateGrantResourcePermissionRequest 验证授予资源权限请求
func (s *DefaultPermissionService) validateGrantResourcePermissionRequest(request *GrantResourcePermissionRequest) error {
	if request.ResourceID == "" {
		return fmt.Errorf("resource_id is required")
	}
	if request.ResourceType == "" {
		return fmt.Errorf("resource_type is required")
	}
	if request.SubjectID == "" {
		return fmt.Errorf("subject_id is required")
	}
	if !ValidateSubjectType(request.SubjectType) {
		return fmt.Errorf("invalid subject type")
	}
	if request.PermissionID == "" {
		return fmt.Errorf("permission_id is required")
	}
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !ValidatePermissionEffect(request.Effect) {
		return fmt.Errorf("invalid permission effect")
	}
	return nil
}

// validatePermissionInheritanceRequest 验证权限继承请求
func (s *DefaultPermissionService) validatePermissionInheritanceRequest(request *PermissionInheritanceRequest) error {
	if request.ResourceID == "" {
		return fmt.Errorf("resource_id is required")
	}
	if request.ResourceType == "" {
		return fmt.Errorf("resource_type is required")
	}
	if request.ParentID == "" {
		return fmt.Errorf("parent_id is required")
	}
	if request.ParentType == "" {
		return fmt.Errorf("parent_type is required")
	}
	if request.InheritType == "" {
		return fmt.Errorf("inherit_type is required")
	}
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// validateCreatePolicyRequest 验证创建策略请求
func (s *DefaultPermissionService) validateCreatePolicyRequest(request *CreatePolicyRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !ValidatePolicyType(request.Type) {
		return fmt.Errorf("invalid policy type")
	}
	if len(request.Rules) == 0 {
		return fmt.Errorf("rules are required")
	}
	if request.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	if !ValidatePermissionEffect(request.Effect) {
		return fmt.Errorf("invalid permission effect")
	}
	return nil
}

// 策略评估器实现

// RBACPolicyEvaluator RBAC策略评估器
type RBACPolicyEvaluator struct {
	service *DefaultPermissionService
	logger  *zap.Logger
}

// Evaluate 评估RBAC策略
func (e *RBACPolicyEvaluator) Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error) {
	// 简化的RBAC评估逻辑
	for _, policy := range policies {
		for _, rule := range policy.Rules {
			if e.matchesRule(rule, request) {
				return &PolicyEvaluationResponse{
					Allowed:       rule.Effect == PermissionEffectAllow,
					Effect:        rule.Effect,
					MatchedRules:  []*PolicyRule{rule},
					EvaluationLog: []string{fmt.Sprintf("RBAC rule matched: %s", rule.ID)},
				}, nil
			}
		}
	}

	return &PolicyEvaluationResponse{
		Allowed:       false,
		Effect:        PermissionEffectDeny,
		EvaluationLog: []string{"No RBAC rule matched"},
	}, nil
}

// matchesRule 检查规则是否匹配
func (e *RBACPolicyEvaluator) matchesRule(rule *PolicyRule, request *PolicyEvaluationRequest) bool {
	// 检查资源
	if rule.Resource != "*" && rule.Resource != request.Resource {
		return false
	}

	// 检查动作
	if rule.Action != "*" && rule.Action != request.Action {
		return false
	}

	// 检查条件
	return e.evaluateConditions(rule.Conditions, request.Context)
}

// evaluateConditions 评估条件
func (e *RBACPolicyEvaluator) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	for key, expectedValue := range conditions {
		if contextValue, exists := context[key]; exists {
			if contextValue != expectedValue {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// ABACPolicyEvaluator ABAC策略评估器
type ABACPolicyEvaluator struct {
	service *DefaultPermissionService
	logger  *zap.Logger
}

// Evaluate 评估ABAC策略
func (e *ABACPolicyEvaluator) Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error) {
	// 简化的ABAC评估逻辑
	// 这里应该实现更复杂的属性基础访问控制逻辑
	return &PolicyEvaluationResponse{
		Allowed:       false,
		Effect:        PermissionEffectDeny,
		EvaluationLog: []string{"ABAC evaluation not implemented"},
	}, nil
}

// ACLPolicyEvaluator ACL策略评估器
type ACLPolicyEvaluator struct {
	service *DefaultPermissionService
	logger  *zap.Logger
}

// Evaluate 评估ACL策略
func (e *ACLPolicyEvaluator) Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error) {
	// 简化的ACL评估逻辑
	// 这里应该实现访问控制列表逻辑
	return &PolicyEvaluationResponse{
		Allowed:       false,
		Effect:        PermissionEffectDeny,
		EvaluationLog: []string{"ACL evaluation not implemented"},
	}, nil
}