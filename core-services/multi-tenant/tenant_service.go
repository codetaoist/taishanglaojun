package multitenant

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TenantStatus 租户状态
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantPlan 租户计划
type TenantPlan string

const (
	TenantPlanBasic      TenantPlan = "basic"
	TenantPlanProfessional TenantPlan = "professional"
	TenantPlanEnterprise TenantPlan = "enterprise"
	TenantPlanCustom     TenantPlan = "custom"
)

// Tenant 租户信息
type Tenant struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	DisplayName string                 `json:"display_name" db:"display_name"`
	Domain      string                 `json:"domain" db:"domain"`
	Status      TenantStatus           `json:"status" db:"status"`
	Plan        TenantPlan             `json:"plan" db:"plan"`
	Settings    TenantSettings         `json:"settings" db:"settings"`
	Limits      TenantLimits           `json:"limits" db:"limits"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	UpdatedBy   string                 `json:"updated_by" db:"updated_by"`
}

// TenantSettings 租户设置
type TenantSettings struct {
	TimeZone           string            `json:"time_zone"`
	Language           string            `json:"language"`
	Currency           string            `json:"currency"`
	DateFormat         string            `json:"date_format"`
	NumberFormat       string            `json:"number_format"`
	Theme              string            `json:"theme"`
	Logo               string            `json:"logo"`
	CustomDomain       string            `json:"custom_domain"`
	SSLEnabled         bool              `json:"ssl_enabled"`
	TwoFactorRequired  bool              `json:"two_factor_required"`
	PasswordPolicy     PasswordPolicy    `json:"password_policy"`
	SessionTimeout     time.Duration     `json:"session_timeout"`
	AllowedIPs         []string          `json:"allowed_ips"`
	BlockedIPs         []string          `json:"blocked_ips"`
	Features           map[string]bool   `json:"features"`
	Integrations       map[string]string `json:"integrations"`
	CustomFields       map[string]string `json:"custom_fields"`
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	MaxAge           int  `json:"max_age"` // 密码最大使用天数
	HistoryCount     int  `json:"history_count"` // 记住的历史密码数量
}

// TenantLimits 租户限制
type TenantLimits struct {
	MaxUsers           int           `json:"max_users"`
	MaxStorage         int64         `json:"max_storage"` // 字节
	MaxAPIRequests     int           `json:"max_api_requests"` // 每小时
	MaxDatabases       int           `json:"max_databases"`
	MaxConnections     int           `json:"max_connections"`
	MaxFileSize        int64         `json:"max_file_size"` // 字节
	MaxBandwidth       int64         `json:"max_bandwidth"` // 每月字节
	MaxProjects        int           `json:"max_projects"`
	MaxWorkspaces      int           `json:"max_workspaces"`
	MaxIntegrations    int           `json:"max_integrations"`
	RateLimits         RateLimits    `json:"rate_limits"`
	FeatureLimits      FeatureLimits `json:"feature_limits"`
}

// RateLimits 速率限制
type RateLimits struct {
	RequestsPerSecond  int `json:"requests_per_second"`
	RequestsPerMinute  int `json:"requests_per_minute"`
	RequestsPerHour    int `json:"requests_per_hour"`
	RequestsPerDay     int `json:"requests_per_day"`
	ConcurrentRequests int `json:"concurrent_requests"`
}

// FeatureLimits 功能限制
type FeatureLimits struct {
	AIRequests        int `json:"ai_requests"` // 每月AI请求数
	VoiceMinutes      int `json:"voice_minutes"` // 每月语音处理分钟数
	ImageProcessing   int `json:"image_processing"` // 每月图像处理数
	DataExport        int `json:"data_export"` // 每月数据导出次数
	CustomReports     int `json:"custom_reports"` // 自定义报告数量
	Webhooks          int `json:"webhooks"` // Webhook数量
	BackupRetention   int `json:"backup_retention"` // 备份保留天数
}

// TenantUsage 租户使用情况
type TenantUsage struct {
	TenantID        string            `json:"tenant_id"`
	Period          string            `json:"period"` // monthly, daily, hourly
	Users           int               `json:"users"`
	Storage         int64             `json:"storage"`
	APIRequests     int               `json:"api_requests"`
	Bandwidth       int64             `json:"bandwidth"`
	AIRequests      int               `json:"ai_requests"`
	VoiceMinutes    int               `json:"voice_minutes"`
	ImageProcessing int               `json:"image_processing"`
	DataExport      int               `json:"data_export"`
	CustomMetrics   map[string]int64  `json:"custom_metrics"`
	Timestamp       time.Time         `json:"timestamp"`
}

// TenantContext 租户上下文
type TenantContext struct {
	TenantID   string
	UserID     string
	Roles      []string
	Permissions []string
	Settings   TenantSettings
	Limits     TenantLimits
	Usage      TenantUsage
}

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name        string                 `json:"name" validate:"required,min=2,max=100"`
	DisplayName string                 `json:"display_name" validate:"required,min=2,max=200"`
	Domain      string                 `json:"domain" validate:"required,hostname"`
	Plan        TenantPlan             `json:"plan" validate:"required"`
	Settings    TenantSettings         `json:"settings"`
	Limits      TenantLimits           `json:"limits"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedBy   string                 `json:"created_by" validate:"required"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	DisplayName *string                 `json:"display_name,omitempty"`
	Domain      *string                 `json:"domain,omitempty"`
	Status      *TenantStatus           `json:"status,omitempty"`
	Plan        *TenantPlan             `json:"plan,omitempty"`
	Settings    *TenantSettings         `json:"settings,omitempty"`
	Limits      *TenantLimits           `json:"limits,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
	UpdatedBy   string                  `json:"updated_by" validate:"required"`
}

// TenantFilter 租户过滤器
type TenantFilter struct {
	Status    []TenantStatus `json:"status,omitempty"`
	Plan      []TenantPlan   `json:"plan,omitempty"`
	Domain    string         `json:"domain,omitempty"`
	CreatedBy string         `json:"created_by,omitempty"`
	CreatedAt *TimeRange     `json:"created_at,omitempty"`
	Search    string         `json:"search,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PaginationRequest 分页请求
type PaginationRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	OrderBy  string `json:"order_by"`
	Order    string `json:"order" validate:"oneof=asc desc"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListTenantsResponse 列出租户响应
type ListTenantsResponse struct {
	Tenants    []Tenant           `json:"tenants"`
	Pagination PaginationResponse `json:"pagination"`
}

// TenantService 租户服务接口
type TenantService interface {
	// 租户管理
	CreateTenant(ctx context.Context, req CreateTenantRequest) (*Tenant, error)
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	UpdateTenant(ctx context.Context, tenantID string, req UpdateTenantRequest) (*Tenant, error)
	DeleteTenant(ctx context.Context, tenantID string) error
	ListTenants(ctx context.Context, filter TenantFilter, pagination PaginationRequest) (*ListTenantsResponse, error)
	
	// 租户状态管理
	ActivateTenant(ctx context.Context, tenantID string) error
	SuspendTenant(ctx context.Context, tenantID string, reason string) error
	DeactivateTenant(ctx context.Context, tenantID string) error
	
	// 租户设置管理
	UpdateTenantSettings(ctx context.Context, tenantID string, settings TenantSettings) error
	GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error)
	
	// 租户限制管理
	UpdateTenantLimits(ctx context.Context, tenantID string, limits TenantLimits) error
	GetTenantLimits(ctx context.Context, tenantID string) (*TenantLimits, error)
	CheckTenantLimit(ctx context.Context, tenantID string, limitType string, value int64) (bool, error)
	
	// 租户使用情况
	RecordUsage(ctx context.Context, tenantID string, usage TenantUsage) error
	GetUsage(ctx context.Context, tenantID string, period string) (*TenantUsage, error)
	GetUsageHistory(ctx context.Context, tenantID string, start, end time.Time) ([]TenantUsage, error)
	
	// 租户上下文
	GetTenantContext(ctx context.Context, tenantID, userID string) (*TenantContext, error)
	ValidateTenantAccess(ctx context.Context, tenantID, userID string) (bool, error)
	
	// 租户数据隔离
	GetTenantDatabase(ctx context.Context, tenantID string) (string, error)
	GetTenantSchema(ctx context.Context, tenantID string) (string, error)
	
	// 租户备份和恢复
	BackupTenant(ctx context.Context, tenantID string) (string, error)
	RestoreTenant(ctx context.Context, tenantID, backupID string) error
	
	// 租户迁移
	MigrateTenant(ctx context.Context, tenantID, targetRegion string) error
	
	// 健康检查
	HealthCheck(ctx context.Context) error
}

// TenantRepository 租户存储接口
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter TenantFilter, pagination PaginationRequest) ([]Tenant, int64, error)
	
	// 使用情况存储
	SaveUsage(ctx context.Context, usage *TenantUsage) error
	GetUsage(ctx context.Context, tenantID string, period string) (*TenantUsage, error)
	GetUsageHistory(ctx context.Context, tenantID string, start, end time.Time) ([]TenantUsage, error)
}

// TenantCache 租户缓存接口
type TenantCache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
}

// TenantEvent 租户事件
type TenantEvent struct {
	ID        string                 `json:"id"`
	TenantID  string                 `json:"tenant_id"`
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data"`
	UserID    string                 `json:"user_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// TenantEventPublisher 租户事件发布器
type TenantEventPublisher interface {
	PublishEvent(ctx context.Context, event TenantEvent) error
}

// 辅助函数

// NewTenant 创建新租户
func NewTenant(req CreateTenantRequest) *Tenant {
	now := time.Now()
	return &Tenant{
		ID:          uuid.New().String(),
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Domain:      req.Domain,
		Status:      TenantStatusActive,
		Plan:        req.Plan,
		Settings:    req.Settings,
		Limits:      req.Limits,
		Metadata:    req.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   req.CreatedBy,
		UpdatedBy:   req.CreatedBy,
	}
}

// IsActive 检查租户是否活跃
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// CanAccess 检查是否可以访问功能
func (t *Tenant) CanAccess(feature string) bool {
	if enabled, exists := t.Settings.Features[feature]; exists {
		return enabled
	}
	return false
}

// GetCacheKey 获取缓存键
func GetTenantCacheKey(tenantID string) string {
	return fmt.Sprintf("tenant:%s", tenantID)
}

// GetDomainCacheKey 获取域名缓存键
func GetDomainCacheKey(domain string) string {
	return fmt.Sprintf("tenant:domain:%s", domain)
}

// GetUsageCacheKey 获取使用情况缓存键
func GetUsageCacheKey(tenantID, period string) string {
	return fmt.Sprintf("tenant:usage:%s:%s", tenantID, period)
}