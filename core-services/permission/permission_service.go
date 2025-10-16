package permission

import (
	"context"
	"fmt"
	"time"
)

// PermissionService 权限服务接口
type PermissionService interface {
	// 权限检查
	CheckPermission(ctx context.Context, request *PermissionCheckRequest) (*PermissionCheckResponse, error)
	CheckPermissions(ctx context.Context, requests []*PermissionCheckRequest) ([]*PermissionCheckResponse, error)

	// 角色管理
	CreateRole(ctx context.Context, request *CreateRoleRequest) (*Role, error)
	GetRole(ctx context.Context, roleID string) (*Role, error)
	GetRoleByName(ctx context.Context, name string, tenantID string) (*Role, error)
	UpdateRole(ctx context.Context, request *UpdateRoleRequest) (*Role, error)
	DeleteRole(ctx context.Context, request *DeleteRoleRequest) error
	ListRoles(ctx context.Context, request *ListRolesRequest) (*ListRolesResponse, error)

	// 权限管理
	CreatePermission(ctx context.Context, request *CreatePermissionRequest) (*Permission, error)
	GetPermission(ctx context.Context, request *GetPermissionRequest) (*Permission, error)
	UpdatePermission(ctx context.Context, request *UpdatePermissionRequest) (*Permission, error)
	DeletePermission(ctx context.Context, request *DeletePermissionRequest) error
	ListPermissions(ctx context.Context, request *ListPermissionsRequest) (*ListPermissionsResponse, error)

	// 角色权限关联
	AssignPermissionToRole(ctx context.Context, request *AssignPermissionToRoleRequest) error
	RevokePermissionFromRole(ctx context.Context, request *RevokePermissionFromRoleRequest) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error)

	// 用户角色关联
	AssignRoleToUser(ctx context.Context, userID, roleID string, tenantID string) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID string, tenantID string) error
	GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*Role, error)
	GetUserPermissions(ctx context.Context, userID string, tenantID string) ([]*Permission, error)

	// 资源权限管理
	GrantResourcePermission(ctx context.Context, request *GrantResourcePermissionRequest) error
	RevokeResourcePermission(ctx context.Context, request *RevokeResourcePermissionRequest) error
	GetResourcePermissions(ctx context.Context, resourceID, resourceType string) ([]*ResourcePermission, error)

	// 权限继承
	SetPermissionInheritance(ctx context.Context, request *PermissionInheritanceRequest) error
	GetPermissionInheritance(ctx context.Context, resourceID, resourceType string) (*PermissionInheritance, error)

	// 策略管理
	CreatePolicy(ctx context.Context, request *CreatePolicyRequest) (*Policy, error)
	GetPolicy(ctx context.Context, policyID string) (*Policy, error)
	UpdatePolicy(ctx context.Context, policyID string, request *UpdatePolicyRequest) (*Policy, error)
	DeletePolicy(ctx context.Context, policyID string) error
	ListPolicies(ctx context.Context, filter *PolicyFilter) (*ListPoliciesResponse, error)
	EvaluatePolicy(ctx context.Context, request *PolicyEvaluationRequest) (*PolicyEvaluationResponse, error)

	// 缓存管理
	InvalidateCache(ctx context.Context, userID, tenantID string) error
	InvalidateAllCache(ctx context.Context) error

	// 权限审计日志
	GetPermissionAuditLog(ctx context.Context, filter *PermissionAuditFilter) (*PermissionAuditResponse, error)

	// 健康检查
	HealthCheck(ctx context.Context) *HealthStatus
}



// Permission 权限
type Permission struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Code        string                 `json:"code" db:"code"`
	Description string                 `json:"description" db:"description"`
	Category    string                 `json:"category" db:"category"`
	Resource    string                 `json:"resource" db:"resource"`
	Action      string                 `json:"action" db:"action"`
	Effect      PermissionEffect       `json:"effect" db:"effect"`
	Conditions  map[string]interface{} `json:"conditions" db:"conditions"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	UpdatedBy   string                 `json:"updated_by" db:"updated_by"`
}

// Role 角色
type Role struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Code        string                 `json:"code" db:"code"`
	Description string                 `json:"description" db:"description"`
	Type        RoleType               `json:"type" db:"type"`
	Level       int                    `json:"level" db:"level"`
	ParentID    *string                `json:"parent_id" db:"parent_id"`
	IsSystem    bool                   `json:"is_system" db:"is_system"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Permissions []*Permission          `json:"permissions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	UpdatedBy   string                 `json:"updated_by" db:"updated_by"`
}

// ResourcePermission 资源权限
type ResourcePermission struct {
	ID           string                 `json:"id" db:"id"`
	ResourceID   string                 `json:"resource_id" db:"resource_id"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	SubjectID    string                 `json:"subject_id" db:"subject_id"`
	SubjectType  SubjectType            `json:"subject_type" db:"subject_type"`
	PermissionID string                 `json:"permission_id" db:"permission_id"`
	Permission   *Permission            `json:"permission,omitempty"`
	Effect       PermissionEffect       `json:"effect" db:"effect"`
	Conditions   map[string]interface{} `json:"conditions" db:"conditions"`
	ExpiresAt    *time.Time             `json:"expires_at" db:"expires_at"`
	TenantID     string                 `json:"tenant_id" db:"tenant_id"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	CreatedBy    string                 `json:"created_by" db:"created_by"`
}

// Policy 策略
type Policy struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Type        PolicyType             `json:"type" db:"type"`
	Rules       []*PolicyRule          `json:"rules" db:"rules"`
	Effect      PermissionEffect       `json:"effect" db:"effect"`
	Priority    int                    `json:"priority" db:"priority"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	UpdatedBy   string                 `json:"updated_by" db:"updated_by"`
}

// PolicyRule 策略规则
type PolicyRule struct {
	ID         string                 `json:"id"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	Effect     PermissionEffect       `json:"effect"`
	Conditions map[string]interface{} `json:"conditions"`
}

// PermissionInheritance 权限继承
type PermissionInheritance struct {
	ID           string    `json:"id" db:"id"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ParentID     string    `json:"parent_id" db:"parent_id"`
	ParentType   string    `json:"parent_type" db:"parent_type"`
	InheritType  string    `json:"inherit_type" db:"inherit_type"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
}

//

// PermissionEffect 权限效果
type PermissionEffect string

const (
	PermissionEffectAllow PermissionEffect = "allow"
	PermissionEffectDeny  PermissionEffect = "deny"
)

// RoleType 角色类型
type RoleType string

const (
	RoleTypeSystem     RoleType = "system"
	RoleTypeCustom     RoleType = "custom"
	RoleTypeFunctional RoleType = "functional"
	RoleTypeData       RoleType = "data"
)

// SubjectType 主体类型
type SubjectType string

const (
	SubjectTypeUser  SubjectType = "user"
	SubjectTypeRole  SubjectType = "role"
	SubjectTypeGroup SubjectType = "group"
)

// PolicyType 策略类型
type PolicyType string

const (
	PolicyTypeRBAC PolicyType = "rbac"
	PolicyTypeABAC PolicyType = "abac"
	PolicyTypeACL  PolicyType = "acl"
)

// ?

// PermissionCheckRequest 权限检查请求
type PermissionCheckRequest struct {
	UserID     string                 `json:"user_id"`
	TenantID   string                 `json:"tenant_id"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	ResourceID *string                `json:"resource_id,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	CheckMode  CheckMode              `json:"check_mode,omitempty"`
}

// PermissionCheckResponse 权限检查响应
type PermissionCheckResponse struct {
	Allowed     bool                   `json:"allowed"`
	Reason      string                 `json:"reason"`
	Effect      PermissionEffect       `json:"effect"`
	Permissions []*Permission          `json:"permissions,omitempty"`
	Policies    []*Policy              `json:"policies,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// CheckMode 检查模式
type CheckMode string

const (
	CheckModeStrict CheckMode = "strict"
	CheckModeLoose  CheckMode = "loose"
)

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Code        string                 `json:"code" validate:"required"`
	Description string                 `json:"description"`
	Type        RoleType               `json:"type"`
	Level       int                    `json:"level"`
	ParentID    *string                `json:"parent_id"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
	TenantID    string                 `json:"tenant_id"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	RoleID      string                 `json:"role_id"`
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Type        *RoleType              `json:"type"`
	Level       *int                   `json:"level"`
	ParentID    *string                `json:"parent_id"`
	IsActive    *bool                  `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Code        string                 `json:"code" validate:"required"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Resource    string                 `json:"resource" validate:"required"`
	Action      string                 `json:"action" validate:"required"`
	Effect      PermissionEffect       `json:"effect"`
	Conditions  map[string]interface{} `json:"conditions"`
	Metadata    map[string]interface{} `json:"metadata"`
	TenantID    string                 `json:"tenant_id"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	PermissionID string                 `json:"permission_id"`
	Name         *string                `json:"name"`
	Description  *string                `json:"description"`
	Category     *string                `json:"category"`
	Effect       *PermissionEffect      `json:"effect"`
	Conditions   map[string]interface{} `json:"conditions"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DeleteRoleRequest 删除角色请求
type DeleteRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}

// ListRolesRequest 列出角色请求
type ListRolesRequest struct {
	Filter *RoleFilter `json:"filter"`
}

// GetPermissionRequest 获取权限请求
type GetPermissionRequest struct {
	PermissionID string `json:"permission_id" validate:"required"`
}

// DeletePermissionRequest 删除权限请求
type DeletePermissionRequest struct {
	PermissionID string `json:"permission_id" validate:"required"`
}

// ListPermissionsRequest 列出权限请求
type ListPermissionsRequest struct {
	Filter *PermissionFilter `json:"filter"`
}

// AssignPermissionToRoleRequest 分配权限给角色请求
type AssignPermissionToRoleRequest struct {
	RoleID       string `json:"role_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// RevokePermissionFromRoleRequest 从角色撤销权限请求
type RevokePermissionFromRoleRequest struct {
	RoleID       string `json:"role_id" validate:"required"`
	PermissionID string `json:"permission_id" validate:"required"`
}

// GrantResourcePermissionRequest 授予资源权限请求
type GrantResourcePermissionRequest struct {
	ResourceID   string                 `json:"resource_id" validate:"required"`
	ResourceType string                 `json:"resource_type" validate:"required"`
	SubjectID    string                 `json:"subject_id" validate:"required"`
	SubjectType  SubjectType            `json:"subject_type" validate:"required"`
	PermissionID string                 `json:"permission_id" validate:"required"`
	Effect       PermissionEffect       `json:"effect"`
	Conditions   map[string]interface{} `json:"conditions"`
	ExpiresAt    *time.Time             `json:"expires_at"`
	TenantID     string                 `json:"tenant_id"`
}

// RevokeResourcePermissionRequest 撤销资源权限请求
type RevokeResourcePermissionRequest struct {
	ResourceID   string      `json:"resource_id" validate:"required"`
	ResourceType string      `json:"resource_type" validate:"required"`
	SubjectID    string      `json:"subject_id" validate:"required"`
	SubjectType  SubjectType `json:"subject_type" validate:"required"`
	PermissionID string      `json:"permission_id" validate:"required"`
	TenantID     string      `json:"tenant_id"`
}

// PermissionInheritanceRequest 权限继承请求
type PermissionInheritanceRequest struct {
	ResourceID   string `json:"resource_id" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	ParentID     string `json:"parent_id" validate:"required"`
	ParentType   string `json:"parent_type" validate:"required"`
	InheritType  string `json:"inherit_type" validate:"required"`
	IsActive     bool   `json:"is_active"`
	TenantID     string `json:"tenant_id"`
}

// CreatePolicyRequest 创建策略请求
type CreatePolicyRequest struct {
	Name        string                 `json:"name" validate:"required"`
	Description string                 `json:"description"`
	Type        PolicyType             `json:"type" validate:"required"`
	Rules       []*PolicyRule          `json:"rules" validate:"required"`
	Effect      PermissionEffect       `json:"effect"`
	Priority    int                    `json:"priority"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
	TenantID    string                 `json:"tenant_id"`
}

// UpdatePolicyRequest 更新策略请求
type UpdatePolicyRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Rules       []*PolicyRule          `json:"rules"`
	Effect      *PermissionEffect      `json:"effect"`
	Priority    *int                   `json:"priority"`
	IsActive    *bool                  `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PolicyEvaluationRequest 策略评估请求
type PolicyEvaluationRequest struct {
	UserID     string                 `json:"user_id" validate:"required"`
	TenantID   string                 `json:"tenant_id" validate:"required"`
	Resource   string                 `json:"resource" validate:"required"`
	Action     string                 `json:"action" validate:"required"`
	ResourceID *string                `json:"resource_id"`
	Context    map[string]interface{} `json:"context"`
	PolicyIDs  []string               `json:"policy_ids"`
}

// PolicyEvaluationResponse 策略评估响应
type PolicyEvaluationResponse struct {
	Allowed       bool                   `json:"allowed"`
	Effect        PermissionEffect       `json:"effect"`
	MatchedRules  []*PolicyRule          `json:"matched_rules"`
	FailedRules   []*PolicyRule          `json:"failed_rules"`
	EvaluationLog []string               `json:"evaluation_log"`
	Context       map[string]interface{} `json:"context"`
}



// RoleFilter 角色筛选器
type RoleFilter struct {
	Name       string            `json:"name"`
	Search     string            `json:"search"`
	Type       *RoleType         `json:"type"`
	Level      *int              `json:"level"`
	ParentID   *string           `json:"parent_id"`
	IsSystem   *bool             `json:"is_system"`
	IsActive   *bool             `json:"is_active"`
	TenantID   string            `json:"tenant_id"`
	Pagination PaginationRequest `json:"pagination"`
}

// PermissionFilter 权限筛选器
type PermissionFilter struct {
	Name       string            `json:"name"`
	Search     string            `json:"search"`
	Category   string            `json:"category"`
	Resource   string            `json:"resource"`
	Action     string            `json:"action"`
	Effect     *PermissionEffect `json:"effect"`
	TenantID   string            `json:"tenant_id"`
	Pagination PaginationRequest `json:"pagination"`
}

// PolicyFilter 策略筛选器
type PolicyFilter struct {
	Name       string            `json:"name"`
	Type       *PolicyType       `json:"type"`
	Effect     *PermissionEffect `json:"effect"`
	IsActive   *bool             `json:"is_active"`
	TenantID   string            `json:"tenant_id"`
	Pagination PaginationRequest `json:"pagination"`
}

// PermissionAuditFilter 权限审计筛选器
type PermissionAuditFilter struct {
	UserID     string            `json:"user_id"`
	TenantID   string            `json:"tenant_id"`
	Resource   string            `json:"resource"`
	Action     string            `json:"action"`
	Effect     *PermissionEffect `json:"effect"`
	StartTime  *time.Time        `json:"start_time"`
	EndTime    *time.Time        `json:"end_time"`
	Pagination PaginationRequest `json:"pagination"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

//

// ListRolesResponse 角色列表响应
type ListRolesResponse struct {
	Roles      []*Role            `json:"roles"`
	Pagination PaginationResponse `json:"pagination"`
}

// ListPermissionsResponse 权限列表响应
type ListPermissionsResponse struct {
	Permissions []*Permission      `json:"permissions"`
	Pagination  PaginationResponse `json:"pagination"`
}

// ListPoliciesResponse 策略列表响应
type ListPoliciesResponse struct {
	Policies   []*Policy          `json:"policies"`
	Pagination PaginationResponse `json:"pagination"`
}

// PermissionAuditResponse 权限审计响应
type PermissionAuditResponse struct {
	AuditLogs  []*PermissionAuditLog `json:"audit_logs"`
	Pagination PaginationResponse    `json:"pagination"`
}

// PermissionAuditLog 权限审计日志
type PermissionAuditLog struct {
	ID         string                 `json:"id" db:"id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	TenantID   string                 `json:"tenant_id" db:"tenant_id"`
	Resource   string                 `json:"resource" db:"resource"`
	Action     string                 `json:"action" db:"action"`
	ResourceID *string                `json:"resource_id" db:"resource_id"`
	Effect     PermissionEffect       `json:"effect" db:"effect"`
	Allowed    bool                   `json:"allowed" db:"allowed"`
	Reason     string                 `json:"reason" db:"reason"`
	Context    map[string]interface{} `json:"context" db:"context"`
	IPAddress  string                 `json:"ip_address" db:"ip_address"`
	UserAgent  string                 `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Healthy   bool              `json:"healthy"`
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

//

// PermissionRepository 权限存储库
type PermissionRepository interface {
	// 权限
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, permissionID string) (*Permission, error)
	UpdatePermission(ctx context.Context, permission *Permission) error
	DeletePermission(ctx context.Context, permissionID string) error
	ListPermissions(ctx context.Context, filter *PermissionFilter) ([]*Permission, int64, error)

	// 角色
	CreateRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, roleID string) (*Role, error)
	GetRoleByName(ctx context.Context, name string, tenantID string) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context, filter *RoleFilter) ([]*Role, int64, error)

	// 角色权限
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RevokePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error)

	// 用户角色
	AssignRoleToUser(ctx context.Context, userID, roleID string, tenantID string) error
	RevokeRoleFromUser(ctx context.Context, userID, roleID string, tenantID string) error
	GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*Role, error)

	// 资源权限
	CreateResourcePermission(ctx context.Context, resourcePermission *ResourcePermission) error
	DeleteResourcePermission(ctx context.Context, resourceID, resourceType, subjectID string, subjectType SubjectType, permissionID string) error
	GetResourcePermissions(ctx context.Context, resourceID, resourceType string) ([]*ResourcePermission, error)

	// 权限继承
	CreatePermissionInheritance(ctx context.Context, inheritance *PermissionInheritance) error
	GetPermissionInheritance(ctx context.Context, resourceID, resourceType string) (*PermissionInheritance, error)
	UpdatePermissionInheritance(ctx context.Context, inheritance *PermissionInheritance) error
	DeletePermissionInheritance(ctx context.Context, resourceID, resourceType string) error

	// 策略
	CreatePolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, policyID string) (*Policy, error)
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, policyID string) error
	ListPolicies(ctx context.Context, filter *PolicyFilter) ([]*Policy, int64, error)

	// 审计日志
	CreatePermissionAuditLog(ctx context.Context, auditLog *PermissionAuditLog) error
	GetPermissionAuditLogs(ctx context.Context, filter *PermissionAuditFilter) ([]*PermissionAuditLog, int64, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// PermissionCache 权限缓存
type PermissionCache interface {
	// 用户权限
	SetUserPermissions(ctx context.Context, userID, tenantID string, permissions []*Permission, ttl time.Duration) error
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]*Permission, error)
	DeleteUserPermissions(ctx context.Context, userID, tenantID string) error

	// 用户角色
	SetUserRoles(ctx context.Context, userID, tenantID string, roles []*Role, ttl time.Duration) error
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error)
	DeleteUserRoles(ctx context.Context, userID, tenantID string) error

	// 权限检查结果
	SetPermissionCheck(ctx context.Context, key string, result *PermissionCheckResponse, ttl time.Duration) error
	GetPermissionCheck(ctx context.Context, key string) (*PermissionCheckResponse, error)
	DeletePermissionCheck(ctx context.Context, key string) error

	// 清除缓存
	Clear(ctx context.Context) error
	ClearUserCache(ctx context.Context, userID, tenantID string) error

	// 健康检查
	HealthCheck(ctx context.Context) error
}



// GeneratePermissionID ID
func GeneratePermissionID() string {
	return "perm_" + generateUUID()
}

// GenerateRoleID ID
func GenerateRoleID() string {
	return "role_" + generateUUID()
}

// GeneratePolicyID ID
func GeneratePolicyID() string {
	return "policy_" + generateUUID()
}

// GenerateResourcePermissionID ID
func GenerateResourcePermissionID() string {
	return "resperm_" + generateUUID()
}

// GenerateInheritanceID ID
func GenerateInheritanceID() string {
	return "inherit_" + generateUUID()
}

// GenerateAuditLogID ID
func GenerateAuditLogID() string {
	return "audit_" + generateUUID()
}

// generateUUID UUID
func generateUUID() string {
	// UUID?
	// ?
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// CreatePermissionCheckKey 权限检查缓存键
func CreatePermissionCheckKey(userID, tenantID, resource, action string, resourceID *string) string {
	key := fmt.Sprintf("perm_check:%s:%s:%s:%s", userID, tenantID, resource, action)
	if resourceID != nil {
		key += ":" + *resourceID
	}
	return key
}

// ValidatePermissionEffect 验证权限效果
func ValidatePermissionEffect(effect PermissionEffect) bool {
	return effect == PermissionEffectAllow || effect == PermissionEffectDeny
}

// ValidateRoleType 验证角色类型
func ValidateRoleType(roleType RoleType) bool {
	validTypes := []RoleType{
		RoleTypeSystem,
		RoleTypeCustom,
		RoleTypeFunctional,
		RoleTypeData,
	}

	for _, validType := range validTypes {
		if roleType == validType {
			return true
		}
	}
	return false
}

// ValidateSubjectType 验证主体类型
func ValidateSubjectType(subjectType SubjectType) bool {
	validTypes := []SubjectType{
		SubjectTypeUser,
		SubjectTypeRole,
		SubjectTypeGroup,
	}

	for _, validType := range validTypes {
		if subjectType == validType {
			return true
		}
	}
	return false
}

// ValidatePolicyType 验证策略类型
func ValidatePolicyType(policyType PolicyType) bool {
	validTypes := []PolicyType{
		PolicyTypeRBAC,
		PolicyTypeABAC,
		PolicyTypeACL,
	}

	for _, validType := range validTypes {
		if policyType == validType {
			return true
		}
	}
	return false
}

// 批量权限检查请求
type BatchPermissionCheckRequest struct {
	Requests []*PermissionCheckRequest `json:"requests" validate:"required"`
}

// 批量权限检查响应
type BatchPermissionCheckResponse struct {
	Responses []*PermissionCheckResponse `json:"responses"`
}

// 用户角色分配请求
type AssignRoleToUserRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// 用户角色撤销请求
type RevokeRoleFromUserRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	RoleID   string `json:"role_id" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// 获取用户角色请求
type GetUserRolesRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	TenantID string `json:"tenant_id"`
}

// 获取用户权限请求
type GetUserPermissionsRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	TenantID string `json:"tenant_id"`
}
