package models

import (
	"time"

	"github.com/google/uuid"
)

// CreateTenantRequest 
type CreateTenantRequest struct {
	Name        string            `json:"name" validate:"required,min=2,max=255"`
	DisplayName string            `json:"display_name" validate:"max=255"`
	Description string            `json:"description" validate:"max=1000"`
	Subdomain   string            `json:"subdomain" validate:"required,min=2,max=100,alphanum"`
	Domain      string            `json:"domain" validate:"omitempty,fqdn"`
	Settings    *TenantSettings   `json:"settings,omitempty"`
	Quota       *TenantQuota      `json:"quota,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	
	// 
	IsolationStrategy IsolationStrategy `json:"isolation_strategy,omitempty" validate:"omitempty,oneof=row_level schema database"`
	
	// ?
	AdminUser *CreateAdminUserRequest `json:"admin_user,omitempty"`
}

// CreateAdminUserRequest ?
type CreateAdminUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"max=255"`
}

// UpdateTenantRequest 
type UpdateTenantRequest struct {
	Name        *string           `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	DisplayName *string           `json:"display_name,omitempty" validate:"omitempty,max=255"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=1000"`
	Domain      *string           `json:"domain,omitempty" validate:"omitempty,fqdn"`
	Status      *TenantStatus     `json:"status,omitempty" validate:"omitempty,oneof=active suspended inactive"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateTenantSettingsRequest 
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

// UpdateTenantQuotaRequest 
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

// AddTenantUserRequest 
type AddTenantUserRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Role        string    `json:"role" validate:"required,min=1,max=50"`
	Permissions []string  `json:"permissions,omitempty"`
}

// UpdateTenantUserRequest 
type UpdateTenantUserRequest struct {
	Role        *string  `json:"role,omitempty" validate:"omitempty,min=1,max=50"`
	Status      *string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	Permissions []string `json:"permissions,omitempty"`
}

// TenantQuery 
type TenantQuery struct {
	Name     string       `form:"name"`
	Status   TenantStatus `form:"status"`
	Search   string       `form:"search"`   // ?
	Page     int          `form:"page" validate:"min=1"`
	PageSize int          `form:"page_size" validate:"min=1,max=100"`
	OrderBy  string       `form:"order_by" validate:"oneof=created_at updated_at name display_name"`
	Order    string       `form:"order" validate:"oneof=asc desc"`
}

// TenantUserQuery 
type TenantUserQuery struct {
	Role     string `form:"role"`
	Status   string `form:"status"`
	Search   string `form:"search"`   // ?
	Page     int    `form:"page" validate:"min=1"`
	PageSize int    `form:"page_size" validate:"min=1,max=100"`
	OrderBy  string `form:"order_by" validate:"oneof=created_at updated_at role"`
	Order    string `form:"order" validate:"oneof=asc desc"`
}

// TenantResponse 
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
	
	// 
	IsolationStrategy IsolationStrategy `json:"isolation_strategy"`
	
	// 
	UserCount         int `json:"user_count"`
	ActiveUserCount   int `json:"active_user_count"`
	SubscriptionCount int `json:"subscription_count"`
	
	// ?
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantListResponse 
type TenantListResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Size    int              `json:"size"`
}

// TenantUserResponse 
type TenantUserResponse struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	UserID      uuid.UUID `json:"user_id"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	Permissions []string  `json:"permissions"`
	
	// 
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	
	// ?
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantUserListResponse 
type TenantUserListResponse struct {
	Users []TenantUserResponse `json:"users"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// TenantStatsResponse 
type TenantStatsResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	
	// 
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	NewUsersToday    int `json:"new_users_today"`
	NewUsersThisWeek int `json:"new_users_this_week"`
	
	// 
	APICallsToday     int     `json:"api_calls_today"`
	APICallsThisWeek  int     `json:"api_calls_this_week"`
	APICallsThisMonth int     `json:"api_calls_this_month"`
	StorageUsedGB     float64 `json:"storage_used_gb"`
	
	// AI
	AIRequestsToday     int `json:"ai_requests_today"`
	AIRequestsThisWeek  int `json:"ai_requests_this_week"`
	AIRequestsThisMonth int `json:"ai_requests_this_month"`
	
	// 
	ActiveSessions    int `json:"active_sessions"`
	TotalSessions     int `json:"total_sessions"`
	AvgSessionDuration float64 `json:"avg_session_duration"` // 
	
	// ?
	QuotaUsage TenantQuotaUsage `json:"quota_usage"`
	
	// 
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantQuotaUsage ?
type TenantQuotaUsage struct {
	UsersUsage                float64 `json:"users_usage"`                  // ?(0-1)
	StorageUsage              float64 `json:"storage_usage"`                // 洢?(0-1)
	APICallsHourUsage         float64 `json:"api_calls_hour_usage"`         // API?(0-1)
	APICallsDayUsage          float64 `json:"api_calls_day_usage"`          // API?(0-1)
	APICallsMonthUsage        float64 `json:"api_calls_month_usage"`        // API?(0-1)
	ConcurrentSessionsUsage   float64 `json:"concurrent_sessions_usage"`    // ?(0-1)
	ConcurrentRequestsUsage   float64 `json:"concurrent_requests_usage"`    // ?(0-1)
	AIRequestsHourUsage       float64 `json:"ai_requests_hour_usage"`       // AI?(0-1)
	AIRequestsDayUsage        float64 `json:"ai_requests_day_usage"`        // AI?(0-1)
	AIRequestsMonthUsage      float64 `json:"ai_requests_month_usage"`      // AI?(0-1)
	DatabaseConnectionsUsage  float64 `json:"database_connections_usage"`   //  (0-1)
}

// TenantHealthResponse ?
type TenantHealthResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Status   string    `json:"status"` // healthy, warning, critical
	
	// 
	Checks []HealthCheck `json:"checks"`
	
	//  (0-100)
	Score int `json:"score"`
	
	// ?
	CheckedAt time.Time `json:"checked_at"`
}

// HealthCheck 
type HealthCheck struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // pass, warn, fail
	Message     string    `json:"message"`
	Value       string    `json:"value,omitempty"`
	Threshold   string    `json:"threshold,omitempty"`
	CheckedAt   time.Time `json:"checked_at"`
}

// TenantConfigResponse 
type TenantConfigResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	
	// 
	Features map[string]bool `json:"features"`
	
	// 
	Limits map[string]interface{} `json:"limits"`
	
	// 
	Security map[string]interface{} `json:"security"`
	
	// 
	Notifications map[string]interface{} `json:"notifications"`
	
	// ?
	Custom map[string]interface{} `json:"custom"`
	
	// 
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantAuditLogResponse 
type TenantAuditLogResponse struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	UserID   uuid.UUID `json:"user_id"`
	
	// 
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	
	// 
	Details map[string]interface{} `json:"details"`
	
	// 
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	
	// 
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	
	// ?
	CreatedAt time.Time `json:"created_at"`
}

// TenantAuditLogListResponse 
type TenantAuditLogListResponse struct {
	Logs  []TenantAuditLogResponse `json:"logs"`
	Total int64                    `json:"total"`
	Page  int                      `json:"page"`
	Size  int                      `json:"size"`
}

// TenantAuditLogQuery 
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

// ErrorResponse 
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse 
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

