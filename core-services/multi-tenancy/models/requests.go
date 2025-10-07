package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name        string            `json:"name" validate:"required,min=2,max=255"`
	DisplayName string            `json:"display_name" validate:"max=255"`
	Description string            `json:"description" validate:"max=1000"`
	Subdomain   string            `json:"subdomain" validate:"required,min=2,max=100,alphanum"`
	Domain      string            `json:"domain" validate:"omitempty,fqdn"`
	Settings    *TenantSettings   `json:"settings,omitempty"`
	Quota       *TenantQuota      `json:"quota,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	
	// 数据隔离策略
	IsolationStrategy IsolationStrategy `json:"isolation_strategy,omitempty" validate:"omitempty,oneof=row_level schema database"`
	
	// 管理员用户信息
	AdminUser *CreateAdminUserRequest `json:"admin_user,omitempty"`
}

// CreateAdminUserRequest 创建管理员用户请求
type CreateAdminUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"max=255"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name        *string           `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	DisplayName *string           `json:"display_name,omitempty" validate:"omitempty,max=255"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=1000"`
	Domain      *string           `json:"domain,omitempty" validate:"omitempty,fqdn"`
	Status      *TenantStatus     `json:"status,omitempty" validate:"omitempty,oneof=active suspended inactive"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateTenantSettingsRequest 更新租户设置请求
type UpdateTenantSettingsRequest struct {
	Timezone     *string                `json:"timezone,omitempty"`
	Language     *string                `json:"language,omitempty"`
	Currency     *string                `json:"currency,omitempty"`
	DateFormat   *string                `json:"date_format,omitempty"`
	TimeFormat   *string                `json:"time_format,omitempty"`
	Features     *TenantFeatures        `json:"features,omitempty"`
	Security     *TenantSecurity        `json:"security,omitempty"`
	Notifications *TenantNotifications  `json:"notifications,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// UpdateTenantQuotaRequest 更新租户配额请求
type UpdateTenantQuotaRequest struct {
	MaxUsers                   *int           `json:"max_users,omitempty" validate:"omitempty,min=1"`
	MaxStorageGB              *int           `json:"max_storage_gb,omitempty" validate:"omitempty,min=1"`
	MaxAPICallsPerHour        *int           `json:"max_api_calls_per_hour,omitempty" validate:"omitempty,min=1"`
	MaxAPICallsPerDay         *int           `json:"max_api_calls_per_day,omitempty" validate:"omitempty,min=1"`
	MaxAPICallsPerMonth       *int           `json:"max_api_calls_per_month,omitempty" validate:"omitempty,min=1"`
	MaxConcurrentSessions     *int           `json:"max_concurrent_sessions,omitempty" validate:"omitempty,min=1"`
	MaxConcurrentRequests     *int           `json:"max_concurrent_requests,omitempty" validate:"omitempty,min=1"`
	MaxAIRequestsPerHour      *int           `json:"max_ai_requests_per_hour,omitempty" validate:"omitempty,min=1"`
	MaxAIRequestsPerDay       *int           `json:"max_ai_requests_per_day,omitempty" validate:"omitempty,min=1"`
	MaxAIRequestsPerMonth     *int           `json:"max_ai_requests_per_month,omitempty" validate:"omitempty,min=1"`
	MaxFileSize               *int           `json:"max_file_size,omitempty" validate:"omitempty,min=1"`
	MaxFilesPerUpload         *int           `json:"max_files_per_upload,omitempty" validate:"omitempty,min=1"`
	MaxDatabaseConnections    *int           `json:"max_database_connections,omitempty" validate:"omitempty,min=1"`
	MaxQueryTimeout           *int           `json:"max_query_timeout,omitempty" validate:"omitempty,min=1"`
	Custom                    map[string]int `json:"custom,omitempty"`
}

// AddTenantUserRequest 添加租户用户请求
type AddTenantUserRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Role        string    `json:"role" validate:"required,min=1,max=50"`
	Permissions []string  `json:"permissions,omitempty"`
}

// UpdateTenantUserRequest 更新租户用户请求
type UpdateTenantUserRequest struct {
	Role        *string  `json:"role,omitempty" validate:"omitempty,min=1,max=50"`
	Status      *string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Permissions []string `json:"permissions,omitempty"`
}

// TenantQuery 租户查询参数
type TenantQuery struct {
	Name     string       `form:"name"`
	Status   TenantStatus `form:"status"`
	Search   string       `form:"search"`   // 搜索租户名称、显示名称、描述
	Page     int          `form:"page" validate:"min=1"`
	PageSize int          `form:"page_size" validate:"min=1,max=100"`
	OrderBy  string       `form:"order_by" validate:"oneof=created_at updated_at name display_name"`
	Order    string       `form:"order" validate:"oneof=asc desc"`
}

// TenantUserQuery 租户用户查询参数
type TenantUserQuery struct {
	Role     string `form:"role"`
	Status   string `form:"status"`
	Search   string `form:"search"`   // 搜索用户名、邮箱
	Page     int    `form:"page" validate:"min=1"`
	PageSize int    `form:"page_size" validate:"min=1,max=100"`
	OrderBy  string `form:"order_by" validate:"oneof=created_at updated_at role"`
	Order    string `form:"order" validate:"oneof=asc desc"`
}

// TenantResponse 租户响应
type TenantResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Subdomain   string            `json:"subdomain"`
	Domain      string            `json:"domain"`
	Status      TenantStatus      `json:"status"`
	Settings    TenantSettings    `json:"settings"`
	Quota       TenantQuota       `json:"quota"`
	Usage       TenantUsage       `json:"usage"`
	Metadata    map[string]string `json:"metadata"`
	
	// 数据隔离配置
	IsolationStrategy IsolationStrategy `json:"isolation_strategy"`
	
	// 统计信息
	UserCount         int `json:"user_count"`
	ActiveUserCount   int `json:"active_user_count"`
	SubscriptionCount int `json:"subscription_count"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantListResponse 租户列表响应
type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
}

// TenantUserResponse 租户用户响应
type TenantUserResponse struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	Permissions []string  `json:"permissions"`
	
	// 用户信息
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantUserListResponse 租户用户列表响应
type TenantUserListResponse struct {
	Users []TenantUserResponse `json:"users"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// TenantStatsResponse 租户统计响应
type TenantStatsResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	
	// 用户统计
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	NewUsersToday    int `json:"new_users_today"`
	NewUsersThisWeek int `json:"new_users_this_week"`
	
	// 使用统计
	APICallsToday     int     `json:"api_calls_today"`
	APICallsThisWeek  int     `json:"api_calls_this_week"`
	APICallsThisMonth int     `json:"api_calls_this_month"`
	StorageUsedGB     float64 `json:"storage_used_gb"`
	
	// AI使用统计
	AIRequestsToday     int `json:"ai_requests_today"`
	AIRequestsThisWeek  int `json:"ai_requests_this_week"`
	AIRequestsThisMonth int `json:"ai_requests_this_month"`
	
	// 会话统计
	ActiveSessions    int `json:"active_sessions"`
	TotalSessions     int `json:"total_sessions"`
	AvgSessionDuration float64 `json:"avg_session_duration"` // 分钟
	
	// 配额使用率
	QuotaUsage TenantQuotaUsage `json:"quota_usage"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantQuotaUsage 租户配额使用率
type TenantQuotaUsage struct {
	UsersUsage                float64 `json:"users_usage"`                  // 用户使用率 (0-1)
	StorageUsage              float64 `json:"storage_usage"`                // 存储使用率 (0-1)
	APICallsHourUsage         float64 `json:"api_calls_hour_usage"`         // 小时API调用使用率 (0-1)
	APICallsDayUsage          float64 `json:"api_calls_day_usage"`          // 日API调用使用率 (0-1)
	APICallsMonthUsage        float64 `json:"api_calls_month_usage"`        // 月API调用使用率 (0-1)
	ConcurrentSessionsUsage   float64 `json:"concurrent_sessions_usage"`    // 并发会话使用率 (0-1)
	ConcurrentRequestsUsage   float64 `json:"concurrent_requests_usage"`    // 并发请求使用率 (0-1)
	AIRequestsHourUsage       float64 `json:"ai_requests_hour_usage"`       // 小时AI请求使用率 (0-1)
	AIRequestsDayUsage        float64 `json:"ai_requests_day_usage"`        // 日AI请求使用率 (0-1)
	AIRequestsMonthUsage      float64 `json:"ai_requests_month_usage"`      // 月AI请求使用率 (0-1)
	DatabaseConnectionsUsage  float64 `json:"database_connections_usage"`   // 数据库连接使用率 (0-1)
}

// TenantHealthResponse 租户健康状态响应
type TenantHealthResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Status   string    `json:"status"` // healthy, warning, critical
	
	// 健康检查项
	Checks []HealthCheck `json:"checks"`
	
	// 总体评分 (0-100)
	Score int `json:"score"`
	
	// 检查时间
	CheckedAt time.Time `json:"checked_at"`
}

// HealthCheck 健康检查项
type HealthCheck struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // pass, warn, fail
	Message     string    `json:"message"`
	Value       string    `json:"value,omitempty"`
	Threshold   string    `json:"threshold,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// TenantConfigResponse 租户配置响应
type TenantConfigResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	
	// 功能配置
	Features map[string]bool `json:"features"`
	
	// 限制配置
	Limits map[string]interface{} `json:"limits"`
	
	// 安全配置
	Security map[string]interface{} `json:"security"`
	
	// 通知配置
	Notifications map[string]interface{} `json:"notifications"`
	
	// 自定义配置
	Custom map[string]interface{} `json:"custom"`
	
	// 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantAuditLogResponse 租户审计日志响应
type TenantAuditLogResponse struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	UserID   uuid.UUID `json:"user_id"`
	
	// 操作信息
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	
	// 详细信息
	Details map[string]interface{} `json:"details"`
	
	// 请求信息
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	
	// 结果
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	
	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// TenantAuditLogListResponse 租户审计日志列表响应
type TenantAuditLogListResponse struct {
	Logs  []TenantAuditLogResponse `json:"logs"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}

// TenantAuditLogQuery 租户审计日志查询参数
type TenantAuditLogQuery struct {
	UserID     uuid.UUID `form:"user_id"`
	Action     string    `form:"action"`
	Resource   string    `form:"resource"`
	Success    *bool     `form:"success"`
	StartDate  time.Time `form:"start_date"`
	EndDate    time.Time `form:"end_date"`
	Page       int       `form:"page" validate:"min=1"`
	PageSize   int       `form:"page_size" validate:"min=1,max=100"`
	OrderBy    string    `form:"order_by" validate:"oneof=created_at action resource"`
	Order      string    `form:"order" validate:"oneof=asc desc"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}