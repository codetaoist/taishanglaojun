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

	// 策略评估器和缓存
	policyEvaluators map[PolicyType]PolicyEvaluator
	permissionCache  map[string]*PermissionCheckResponse
	cacheExpiry      map[string]time.Time
}

// DefaultPermissionServiceConfig 默认权限服务配置
type DefaultPermissionServiceConfig struct {
	// 缓存配置
	CacheEnabled       bool          `json:"cache_enabled"`
	CacheTTL           time.Duration `json:"cache_ttl"`
	UserPermissionsTTL time.Duration `json:"user_permissions_ttl"`
	UserRolesTTL       time.Duration `json:"user_roles_ttl"`
	PermissionCheckTTL time.Duration `json:"permission_check_ttl"`

	// 权限检查配置
	DefaultCheckMode    CheckMode `json:"default_check_mode"`
	EnableInheritance   bool      `json:"enable_inheritance"`
	EnablePolicyEngine  bool      `json:"enable_policy_engine"`
	MaxInheritanceDepth int       `json:"max_inheritance_depth"`

	// 性能配置
	BatchSize           int           `json:"batch_size"`
	MaxConcurrentChecks int           `json:"max_concurrent_checks"`
	CheckTimeout        time.Duration `json:"check_timeout"`

	// 审计配置
	EnableAuditLog     bool     `json:"enable_audit_log"`
	AuditAllChecks     bool     `json:"audit_all_checks"`
	AuditFailedChecks  bool     `json:"audit_failed_checks"`
	SensitiveResources []string `json:"sensitive_resources"`

	// 策略评估模式
	PolicyEvaluationMode string        `json:"policy_evaluation_mode"`
	PolicyCacheEnabled   bool          `json:"policy_cache_enabled"`
	PolicyCacheTTL       time.Duration `json:"policy_cache_ttl"`

	// 角色继承配置
	MaxRoleDepth        int  `json:"max_role_depth"`
	EnableRoleHierarchy bool `json:"enable_role_hierarchy"`

	// 指标配置
	EnableMetrics   bool          `json:"enable_metrics"`
	MetricsInterval time.Duration `json:"metrics_interval"`
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
	//
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
		cache:            cache,
		logger:           logger,
		config:           config,
		policyEvaluators: make(map[PolicyType]PolicyEvaluator),
		permissionCache:  make(map[string]*PermissionCheckResponse),
		cacheExpiry:      make(map[string]time.Time),
	}

	// ?
	service.registerPolicyEvaluators()

	return service
}

// CheckPermission 检查权限
func (s *DefaultPermissionService) CheckPermission(ctx context.Context, request *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	//
	if err := s.validatePermissionCheckRequest(request); err != nil {
		return &PermissionCheckResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("Invalid request: %v", err),
			Effect:  PermissionEffectDeny,
		}, nil
	}

	//
	if s.config.CacheEnabled {
		cacheKey := CreatePermissionCheckKey(request.UserID, request.TenantID, request.Resource, request.Action, request.ResourceID)
		if cached, err := s.getFromCache(cacheKey); err == nil && cached != nil {
			s.logger.Debug("Permission check cache hit", zap.String("cache_key", cacheKey))
			return cached, nil
		}
	}

	// ?
	response, err := s.doPermissionCheck(ctx, request)
	if err != nil {
		s.logger.Error("Permission check failed", zap.Error(err))
		return &PermissionCheckResponse{
			Allowed: false,
			Reason:  fmt.Sprintf("Check failed: %v", err),
			Effect:  PermissionEffectDeny,
		}, nil
	}

	//
	if s.config.CacheEnabled {
		cacheKey := CreatePermissionCheckKey(request.UserID, request.TenantID, request.Resource, request.Action, request.ResourceID)
		s.setToCache(cacheKey, response)
	}

	//
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

	// ?
	if len(requests) > 1 && s.config.MaxConcurrentChecks > 1 {
		return s.checkPermissionsConcurrent(ctx, requests)
	}

	// ?
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
	//
	if err := s.validateCreateRoleRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	//
	existing, err := s.repository.GetRoleByName(ctx, request.Name, request.TenantID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", request.Name)
	}

	//
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

	//
	if err := s.repository.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	//
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

	//
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

	//
	permissions, err := s.repository.GetRolePermissions(ctx, roleID)
	if err != nil {
		s.logger.Warn("Failed to load role permissions", zap.String("role_id", roleID), zap.Error(err))
	} else {
		role.Permissions = permissions
	}

	return role, nil
}

// GetRoleByName 获取角色ByName
func (s *DefaultPermissionService) GetRoleByName(ctx context.Context, name string, tenantID string) (*Role, error) {
	return s.repository.GetRoleByName(ctx, name, tenantID)
}

// UpdateRole 更新角色
func (s *DefaultPermissionService) UpdateRole(ctx context.Context, roleID string, request *UpdateRoleRequest) (*Role, error) {
	//
	role, err := s.repository.GetRole(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	//
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

	//
	if err := s.repository.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	//
	if s.config.CacheEnabled {
		s.invalidateRoleCache(ctx, roleID, role.TenantID)
	}

	s.logger.Info("Role updated successfully", zap.String("role_id", roleID))

	return role, nil
}

// DeleteRole 删除角色
func (s *DefaultPermissionService) DeleteRole(ctx context.Context, roleID string) error {
	//
	role, err := s.repository.GetRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	//
	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}

	//
	if err := s.repository.DeleteRole(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	//
	if s.config.CacheEnabled {
		s.invalidateRoleCache(ctx, roleID, role.TenantID)
	}

	s.logger.Info("Role deleted successfully", zap.String("role_id", roleID))

	return nil
}

// ListRoles 获取角色列表
func (s *DefaultPermissionService) ListRoles(ctx context.Context, filter *RoleFilter) (*ListRolesResponse, error) {
	roles, total, err := s.repository.ListRoles(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	//
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
	//
	if err := s.validateCreatePermissionRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	//
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

	//
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
	//
	permission, err := s.repository.GetPermission(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	//
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

	//
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

// ListPermissions 获取权限列表
func (s *DefaultPermissionService) ListPermissions(ctx context.Context, filter *PermissionFilter) (*ListPermissionsResponse, error) {
	permissions, total, err := s.repository.ListPermissions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	//
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

// AssignPermissionToRole 为角色分配权限
func (s *DefaultPermissionService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	if err := s.repository.AssignPermissionToRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	//
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

// RevokePermissionFromRole 从角色中撤销权限
func (s *DefaultPermissionService) RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	if err := s.repository.RevokePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return fmt.Errorf("failed to revoke permission from role: %w", err)
	}

	//
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

// GetRolePermissions 获取角色的权限列表
func (s *DefaultPermissionService) GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error) {
	return s.repository.GetRolePermissions(ctx, roleID)
}

// AssignRoleToUser 为用户分配角色
func (s *DefaultPermissionService) AssignRoleToUser(ctx context.Context, userID, roleID string, tenantID string) error {
	if err := s.repository.AssignRoleToUser(ctx, userID, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	//
	if s.config.CacheEnabled {
		s.invalidateUserCache(ctx, userID, tenantID)
	}

	s.logger.Info("Role assigned to user successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID),
		zap.String("tenant_id", tenantID))

	return nil
}

// RevokeRoleFromUser 从用户中撤销角色
func (s *DefaultPermissionService) RevokeRoleFromUser(ctx context.Context, userID, roleID string, tenantID string) error {
	if err := s.repository.RevokeRoleFromUser(ctx, userID, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to revoke role from user: %w", err)
	}

	//
	if s.config.CacheEnabled {
		s.invalidateUserCache(ctx, userID, tenantID)
	}

	s.logger.Info("Role revoked from user successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetUserRoles 获取用户的角色列表
func (s *DefaultPermissionService) GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*Role, error) {
	// 黺?
	if s.config.CacheEnabled && s.cache != nil {
		if cached, err := s.cache.GetUserRoles(ctx, userID, tenantID); err == nil && cached != nil {
			return cached, nil
		}
	}

	//
	roles, err := s.repository.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	//
	if s.config.CacheEnabled && s.cache != nil {
		s.cache.SetUserRoles(ctx, userID, tenantID, roles, s.config.UserRolesTTL)
	}

	return roles, nil
}

// GetUserPermissions 获取用户的权限列表
func (s *DefaultPermissionService) GetUserPermissions(ctx context.Context, userID string, tenantID string) ([]*Permission, error) {
	// 黺?
	if s.config.CacheEnabled && s.cache != nil {
		if cached, err := s.cache.GetUserPermissions(ctx, userID, tenantID); err == nil && cached != nil {
			return cached, nil
		}
	}

	//
	roles, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// ?
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

	// ?
	permissions := make([]*Permission, 0, len(permissionMap))
	for _, permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	//
	if s.config.CacheEnabled && s.cache != nil {
		s.cache.SetUserPermissions(ctx, userID, tenantID, permissions, s.config.UserPermissionsTTL)
	}

	return permissions, nil
}

// GrantResourcePermission 为资源授权权限
func (s *DefaultPermissionService) GrantResourcePermission(ctx context.Context, request *GrantResourcePermissionRequest) error {
	//
	if err := s.validateGrantResourcePermissionRequest(request); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	//
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

	//
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

// GetResourcePermissions 获取资源的权限列表
func (s *DefaultPermissionService) GetResourcePermissions(ctx context.Context, resourceID, resourceType string) ([]*ResourcePermission, error) {
	return s.repository.GetResourcePermissions(ctx, resourceID, resourceType)
}

// SetPermissionInheritance 设置权限继承关系
func (s *DefaultPermissionService) SetPermissionInheritance(ctx context.Context, request *PermissionInheritanceRequest) error {
	//
	if err := s.validatePermissionInheritanceRequest(request); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	//
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

	//
	if err := s.repository.CreatePermissionInheritance(ctx, inheritance); err != nil {
		return fmt.Errorf("failed to set permission inheritance: %w", err)
	}

	s.logger.Info("Permission inheritance set successfully",
		zap.String("resource_id", request.ResourceID),
		zap.String("parent_id", request.ParentID))

	return nil
}

// GetPermissionInheritance 获取权限继承关系
func (s *DefaultPermissionService) GetPermissionInheritance(ctx context.Context, resourceID, resourceType string) (*PermissionInheritance, error) {
	return s.repository.GetPermissionInheritance(ctx, resourceID, resourceType)
}

// CreatePolicy 创建策略
func (s *DefaultPermissionService) CreatePolicy(ctx context.Context, request *CreatePolicyRequest) (*Policy, error) {
	//
	if err := s.validateCreatePolicyRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	//
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

	//
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
	//
	policy, err := s.repository.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	//
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

	//
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

// ListPolicies 获取策略列表
func (s *DefaultPermissionService) ListPolicies(ctx context.Context, filter *PolicyFilter) (*ListPoliciesResponse, error) {
	policies, total, err := s.repository.ListPolicies(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	//
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
	//
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
		// ?
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

	//
	for _, policy := range policies {
		if evaluator, exists := s.policyEvaluators[policy.Type]; exists {
			return evaluator.Evaluate(ctx, request, []*Policy{policy})
		}
	}

	//
	return &PolicyEvaluationResponse{
		Allowed:       false,
		Effect:        PermissionEffectDeny,
		EvaluationLog: []string{"No applicable policy found"},
	}, nil
}

// InvalidateCache 失效用户缓存
func (s *DefaultPermissionService) InvalidateCache(ctx context.Context, userID, tenantID string) error {
	if s.config.CacheEnabled && s.cache != nil {
		return s.cache.ClearUserCache(ctx, userID, tenantID)
	}
	return nil
}

// InvalidateAllCache 失效所有缓存
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

	//
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

	//
	if err := s.repository.HealthCheck(ctx); err != nil {
		status.Healthy = false
		status.Status = "unhealthy"
		status.Checks["database"] = fmt.Sprintf("failed: %v", err)
	} else {
		status.Checks["database"] = "ok"
	}

	// 黺?
	if s.cache != nil {
		if err := s.cache.HealthCheck(ctx); err != nil {
			status.Checks["cache"] = fmt.Sprintf("failed: %v", err)
		} else {
			status.Checks["cache"] = "ok"
		}
	}

	return status
}

// doPermissionCheck 执行权限检查
func (s *DefaultPermissionService) doPermissionCheck(ctx context.Context, request *PermissionCheckRequest) (*PermissionCheckResponse, error) {
	//
	permissions, err := s.getUserEffectivePermissions(ctx, request.UserID, request.TenantID, request.ResourceID, request.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// ?
	for _, permission := range permissions {
		if s.matchesPermission(permission, request) {
			// ?
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

	// ?
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
				Policies: []*Policy{}, // ?
			}, nil
		}
	}

	//
	return &PermissionCheckResponse{
		Allowed: false,
		Reason:  "No matching permission found",
		Effect:  PermissionEffectDeny,
	}, nil
}

// getUserEffectivePermissions 获取用户有效权限
func (s *DefaultPermissionService) getUserEffectivePermissions(ctx context.Context, userID, tenantID string, resourceID *string, resource string) ([]*Permission, error) {
	//
	userPermissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	effectivePermissions := make([]*Permission, 0, len(userPermissions))
	effectivePermissions = append(effectivePermissions, userPermissions...)

	// ID?
	if resourceID != nil {
		resourcePermissions, err := s.getResourceSpecificPermissions(ctx, *resourceID, resource, userID, tenantID)
		if err != nil {
			s.logger.Warn("Failed to get resource specific permissions", zap.Error(err))
		} else {
			effectivePermissions = append(effectivePermissions, resourcePermissions...)
		}
	}

	// ?
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
		// ?
		if s.isResourcePermissionApplicable(rp, userID, tenantID) {
			if rp.Permission != nil {
				permissions = append(permissions, rp.Permission)
			} else {
				//
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

	// ?
	parentPermissions, err := s.getResourceSpecificPermissions(ctx, inheritance.ParentID, inheritance.ParentType, userID, tenantID)
	if err != nil {
		return nil, err
	}

	//
	inheritedPermissions, err := s.getInheritedPermissions(ctx, inheritance.ParentID, inheritance.ParentType, userID, tenantID, depth+1)
	if err != nil {
		s.logger.Warn("Failed to get inherited permissions", zap.Error(err))
	} else {
		parentPermissions = append(parentPermissions, inheritedPermissions...)
	}

	return parentPermissions, nil
}

// isResourcePermissionApplicable 判断资源权限是否适用
func (s *DefaultPermissionService) isResourcePermissionApplicable(rp *ResourcePermission, userID, tenantID string) bool {
	// ?
	if rp.TenantID != tenantID {
		return false
	}

	// ?
	if rp.ExpiresAt != nil && rp.ExpiresAt.Before(time.Now()) {
		return false
	}

	// ?
	switch rp.SubjectType {
	case SubjectTypeUser:
		return rp.SubjectID == userID
	case SubjectTypeRole:
		// ?
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
		//
		// false
		return false
	default:
		return false
	}
}

// matchesPermission 判断权限是否匹配
func (s *DefaultPermissionService) matchesPermission(permission *Permission, request *PermissionCheckRequest) bool {
	// ?
	if !s.matchesResource(permission.Resource, request.Resource) {
		return false
	}

	//
	if !s.matchesAction(permission.Action, request.Action) {
		return false
	}

	return true
}

// matchesResource 判断资源是否匹配
func (s *DefaultPermissionService) matchesResource(permissionResource, requestResource string) bool {
	// ?
	if permissionResource == "*" {
		return true
	}

	//
	if permissionResource == requestResource {
		return true
	}

	//
	if strings.HasSuffix(permissionResource, "*") {
		prefix := strings.TrimSuffix(permissionResource, "*")
		return strings.HasPrefix(requestResource, prefix)
	}

	return false
}

// matchesAction 判断操作是否匹配
func (s *DefaultPermissionService) matchesAction(permissionAction, requestAction string) bool {
	// ?
	if permissionAction == "*" {
		return true
	}

	//
	if permissionAction == requestAction {
		return true
	}

	//
	if strings.HasSuffix(permissionAction, "*") {
		prefix := strings.TrimSuffix(permissionAction, "*")
		return strings.HasPrefix(requestAction, prefix)
	}

	return false
}

// evaluateConditions 判断条件是否满足
func (s *DefaultPermissionService) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	//
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

	//
	semaphore := make(chan struct{}, s.config.MaxConcurrentChecks)
	var wg sync.WaitGroup

	for i, request := range requests {
		wg.Add(1)
		go func(index int, req *PermissionCheckRequest) {
			defer wg.Done()

			// ?
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// ?
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

//

// getFromCache 从缓存中获取权限检查结果
func (s *DefaultPermissionService) getFromCache(key string) (*PermissionCheckResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if response, exists := s.permissionCache[key]; exists {
		if expiry, exists := s.cacheExpiry[key]; exists && time.Now().Before(expiry) {
			return response, nil
		}
		// ?
		delete(s.permissionCache, key)
		delete(s.cacheExpiry, key)
	}

	return nil, fmt.Errorf("not found")
}

// setToCache 缓存权限检查结果
func (s *DefaultPermissionService) setToCache(key string, response *PermissionCheckResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.permissionCache[key] = response
	s.cacheExpiry[key] = time.Now().Add(s.config.PermissionCheckTTL)
}

// invalidateUserCache 失效用户缓存
func (s *DefaultPermissionService) invalidateUserCache(ctx context.Context, userID, tenantID string) {
	if s.cache != nil {
		s.cache.ClearUserCache(ctx, userID, tenantID)
	}
}

// invalidateRoleCache 失效角色缓存
func (s *DefaultPermissionService) invalidateRoleCache(ctx context.Context, roleID, tenantID string) {
	//
	// ?
	if s.cache != nil {
		s.cache.Clear(ctx)
	}
}

//

// logPermissionCheck 记录权限检查日志
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

	//
	go func() {
		if err := s.repository.CreatePermissionAuditLog(context.Background(), auditLog); err != nil {
			s.logger.Error("Failed to create permission audit log", zap.Error(err))
		}
	}()
}

// ?

// registerPolicyEvaluators 注册策略评估器
func (s *DefaultPermissionService) registerPolicyEvaluators() {
	// RBAC?
	s.policyEvaluators[PolicyTypeRBAC] = &RBACPolicyEvaluator{
		service: s,
		logger:  s.logger,
	}

	// ABAC?
	s.policyEvaluators[PolicyTypeABAC] = &ABACPolicyEvaluator{
		service: s,
		logger:  s.logger,
	}

	// ACL?
	s.policyEvaluators[PolicyTypeACL] = &ACLPolicyEvaluator{
		service: s,
		logger:  s.logger,
	}
}

//

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

// ?

// RBACPolicyEvaluator RBAC 策略评估器
type RBACPolicyEvaluator struct {
	service *DefaultPermissionService
	logger  *zap.Logger
}

// Evaluate RBAC 策略评估
func (e *RBACPolicyEvaluator) Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error) {
	// RBAC
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

// matchesRule 匹配规则
func (e *RBACPolicyEvaluator) matchesRule(rule *PolicyRule, request *PolicyEvaluationRequest) bool {
	// ?
	if rule.Resource != "*" && rule.Resource != request.Resource {
		return false
	}

	// 匹配操作
	if rule.Action != "*" && rule.Action != request.Action {
		return false
	}

	// ?
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

// ABACPolicyEvaluator ABAC 策略评估器
type ABACPolicyEvaluator struct {
	service *DefaultPermissionService
	logger  *zap.Logger
}

// Evaluate ABAC 策略评估
func (e *ABACPolicyEvaluator) Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error) {
	// ABAC
	//
	return &PolicyEvaluationResponse{
		Allowed:       false,
		Effect:        PermissionEffectDeny,
		EvaluationLog: []string{"ABAC evaluation not implemented"},
	}, nil
}

// ACLPolicyEvaluator ACL 策略评估器
type ACLPolicyEvaluator struct {
	service *DefaultPermissionService
	logger  *zap.Logger
}

// Evaluate ACL 策略评估
func (e *ACLPolicyEvaluator) Evaluate(ctx context.Context, request *PolicyEvaluationRequest, policies []*Policy) (*PolicyEvaluationResponse, error) {
	// ACL
	//
	return &PolicyEvaluationResponse{
		Allowed:       false,
		Effect:        PermissionEffectDeny,
		EvaluationLog: []string{"ACL evaluation not implemented"},
	}, nil
}
