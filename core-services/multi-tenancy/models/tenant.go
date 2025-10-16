package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantStatus ?
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"    // 
	TenantStatusSuspended TenantStatus = "suspended" // 
	TenantStatusInactive  TenantStatus = "inactive"  // ?
	TenantStatusDeleted   TenantStatus = "deleted"   // ?
)

// IsolationStrategy 
type IsolationStrategy string

const (
	IsolationRowLevel IsolationStrategy = "row_level" // 
	IsolationSchema   IsolationStrategy = "schema"    // Schema
	IsolationDatabase IsolationStrategy = "database"  // ?
)

// TenantResolution 
type TenantResolution string

const (
	ResolutionHeader    TenantResolution = "header"    // HTTP
	ResolutionSubdomain TenantResolution = "subdomain" // ?
	ResolutionPath      TenantResolution = "path"      // URL
)

// Tenant 
type Tenant struct {
	ID          uuid.UUID         `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string            `gorm:"type:varchar(255);not null" json:"name"`
	DisplayName string            `gorm:"type:varchar(255)" json:"display_name"`
	Description string            `gorm:"type:text" json:"description"`
	Subdomain   string            `gorm:"type:varchar(100);uniqueIndex" json:"subdomain"`
	Domain      string            `gorm:"type:varchar(255)" json:"domain"`
	Status      TenantStatus      `gorm:"type:varchar(20);default:'active'" json:"status"`
	Settings    TenantSettings    `gorm:"type:json" json:"settings"`
	Quota       TenantQuota       `gorm:"type:json" json:"quota"`
	Usage       TenantUsage       `gorm:"type:json" json:"usage"`
	Metadata    map[string]string `gorm:"type:json" json:"metadata"`
	
	// 
	IsolationStrategy IsolationStrategy `gorm:"type:varchar(20);default:'row_level'" json:"isolation_strategy"`
	DatabaseName      string            `gorm:"type:varchar(100)" json:"database_name,omitempty"`
	SchemaName        string            `gorm:"type:varchar(100)" json:"schema_name,omitempty"`
	
	// ?
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 
	Users         []TenantUser         `gorm:"foreignKey:TenantID" json:"users,omitempty"`
	Subscriptions []TenantSubscription `gorm:"foreignKey:TenantID" json:"subscriptions,omitempty"`
}

// TenantSettings 
type TenantSettings struct {
	// 
	Timezone     string `json:"timezone"`
	Language     string `json:"language"`
	Currency     string `json:"currency"`
	DateFormat   string `json:"date_format"`
	TimeFormat   string `json:"time_format"`
	
	// ?
	Features TenantFeatures `json:"features"`
	
	// 
	Security TenantSecurity `json:"security"`
	
	// 
	Notifications TenantNotifications `json:"notifications"`
	
	// ?
	Custom map[string]interface{} `json:"custom"`
}

// TenantFeatures ?
type TenantFeatures struct {
	AIIntegration     bool `json:"ai_integration"`
	MultimodalAI      bool `json:"multimodal_ai"`
	CulturalWisdom    bool `json:"cultural_wisdom"`
	Community         bool `json:"community"`
	LocationTracking  bool `json:"location_tracking"`
	Consciousness     bool `json:"consciousness"`
	Analytics         bool `json:"analytics"`
	API               bool `json:"api"`
	Webhooks          bool `json:"webhooks"`
	CustomBranding    bool `json:"custom_branding"`
	SSO               bool `json:"sso"`
	LDAP              bool `json:"ldap"`
}

// TenantSecurity 
type TenantSecurity struct {
	PasswordPolicy    PasswordPolicy    `json:"password_policy"`
	SessionPolicy     SessionPolicy     `json:"session_policy"`
	IPWhitelist       []string          `json:"ip_whitelist"`
	IPBlacklist       []string          `json:"ip_blacklist"`
	TwoFactorRequired bool              `json:"two_factor_required"`
	AuditLogging      bool              `json:"audit_logging"`
	DataRetention     DataRetentionPolicy `json:"data_retention"`
}

// PasswordPolicy 
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	ExpirationDays   int  `json:"expiration_days"`
	HistoryCount     int  `json:"history_count"`
}

// SessionPolicy 
type SessionPolicy struct {
	MaxDuration       int  `json:"max_duration"`        // ?
	IdleTimeout       int  `json:"idle_timeout"`        // 
	MaxConcurrentSessions int `json:"max_concurrent_sessions"` // 
	RequireReauth     bool `json:"require_reauth"`      // ?
}

// DataRetentionPolicy 
type DataRetentionPolicy struct {
	LogRetentionDays    int `json:"log_retention_days"`
	BackupRetentionDays int `json:"backup_retention_days"`
	DeletedDataRetentionDays int `json:"deleted_data_retention_days"`
}

// TenantNotifications 
type TenantNotifications struct {
	Email    EmailNotificationSettings    `json:"email"`
	SMS      SMSNotificationSettings      `json:"sms"`
	Push     PushNotificationSettings     `json:"push"`
	Webhook  WebhookNotificationSettings  `json:"webhook"`
}

// EmailNotificationSettings 
type EmailNotificationSettings struct {
	Enabled    bool   `json:"enabled"`
	SMTPHost   string `json:"smtp_host"`
	SMTPPort   int    `json:"smtp_port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	FromEmail  string `json:"from_email"`
	FromName   string `json:"from_name"`
	UseSSL     bool   `json:"use_ssl"`
	UseTLS     bool   `json:"use_tls"`
}

// SMSNotificationSettings 
type SMSNotificationSettings struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	FromNumber string `json:"from_number"`
}

// PushNotificationSettings 
type PushNotificationSettings struct {
	Enabled bool   `json:"enabled"`
	FCMKey  string `json:"fcm_key"`
	APNSKey string `json:"apns_key"`
}

// WebhookNotificationSettings Webhook
type WebhookNotificationSettings struct {
	Enabled bool     `json:"enabled"`
	URLs    []string `json:"urls"`
	Secret  string   `json:"secret"`
}

// TenantQuota 
type TenantQuota struct {
	// 
	MaxUsers int `json:"max_users"`
	
	// 洢
	MaxStorageGB int `json:"max_storage_gb"`
	
	// API
	MaxAPICallsPerHour   int `json:"max_api_calls_per_hour"`
	MaxAPICallsPerDay    int `json:"max_api_calls_per_day"`
	MaxAPICallsPerMonth  int `json:"max_api_calls_per_month"`
	
	// 
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
	
	// AI
	MaxAIRequestsPerHour  int `json:"max_ai_requests_per_hour"`
	MaxAIRequestsPerDay   int `json:"max_ai_requests_per_day"`
	MaxAIRequestsPerMonth int `json:"max_ai_requests_per_month"`
	
	// 
	MaxFileSize       int `json:"max_file_size"`        // MB?
	MaxFilesPerUpload int `json:"max_files_per_upload"` // 
	
	// ?
	MaxDatabaseConnections int `json:"max_database_connections"`
	MaxQueryTimeout        int `json:"max_query_timeout"` // ?
	
	// ?
	Custom map[string]int `json:"custom"`
}

// TenantUsage 
type TenantUsage struct {
	// 
	CurrentUsers int `json:"current_users"`
	
	// 洢
	CurrentStorageGB float64 `json:"current_storage_gb"`
	
	// API
	APICallsThisHour  int `json:"api_calls_this_hour"`
	APICallsThisDay   int `json:"api_calls_this_day"`
	APICallsThisMonth int `json:"api_calls_this_month"`
	
	// 
	CurrentSessions int `json:"current_sessions"`
	CurrentRequests int `json:"current_requests"`
	
	// AI
	AIRequestsThisHour  int `json:"ai_requests_this_hour"`
	AIRequestsThisDay   int `json:"ai_requests_this_day"`
	AIRequestsThisMonth int `json:"ai_requests_this_month"`
	
	// ?
	CurrentDatabaseConnections int `json:"current_database_connections"`
	
	// ?
	LastUpdated time.Time `json:"last_updated"`
	
	// ?
	Custom map[string]interface{} `json:"custom"`
}

// TenantUser 
type TenantUser struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:char(36);not null;index" json:"tenant_id"`
	UserID   uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	Role     string    `gorm:"type:varchar(50);not null" json:"role"`
	Status   string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	
	// 
	Permissions []string `gorm:"type:json" json:"permissions"`
	
	// ?
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// TenantSubscription 
type TenantSubscription struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:char(36);not null;index" json:"tenant_id"`
	
	// 
	PlanID      string    `gorm:"type:varchar(100);not null" json:"plan_id"`
	PlanName    string    `gorm:"type:varchar(255)" json:"plan_name"`
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	TrialEnd    *time.Time `json:"trial_end,omitempty"`
	
	// 
	BillingCycle string  `gorm:"type:varchar(20)" json:"billing_cycle"` // monthly, yearly
	Amount       float64 `json:"amount"`
	Currency     string  `gorm:"type:varchar(10)" json:"currency"`
	
	// IDStripeID?
	ExternalSubscriptionID string `gorm:"type:varchar(255)" json:"external_subscription_id"`
	
	// ?
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// BeforeCreate ?
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	
	// 
	if t.Quota.MaxUsers == 0 {
		t.Quota = GetDefaultQuota()
	}
	
	// 
	if t.Settings.Timezone == "" {
		t.Settings = GetDefaultSettings()
	}
	
	// ?
	t.Usage = TenantUsage{
		LastUpdated: time.Now(),
		Custom:      make(map[string]interface{}),
	}
	
	return nil
}

// BeforeCreate ?- TenantUser
func (tu *TenantUser) BeforeCreate(tx *gorm.DB) error {
	if tu.ID == uuid.Nil {
		tu.ID = uuid.New()
	}
	return nil
}

// BeforeCreate ?- TenantSubscription
func (ts *TenantSubscription) BeforeCreate(tx *gorm.DB) error {
	if ts.ID == uuid.Nil {
		ts.ID = uuid.New()
	}
	return nil
}

// GetDefaultQuota 
func GetDefaultQuota() TenantQuota {
	return TenantQuota{
		MaxUsers:                   100,
		MaxStorageGB:              10,
		MaxAPICallsPerHour:        1000,
		MaxAPICallsPerDay:         10000,
		MaxAPICallsPerMonth:       300000,
		MaxConcurrentSessions:     50,
		MaxConcurrentRequests:     100,
		MaxAIRequestsPerHour:      100,
		MaxAIRequestsPerDay:       1000,
		MaxAIRequestsPerMonth:     30000,
		MaxFileSize:               50,
		MaxFilesPerUpload:         10,
		MaxDatabaseConnections:    20,
		MaxQueryTimeout:           30,
		Custom:                    make(map[string]int),
	}
}

// GetDefaultSettings 
func GetDefaultSettings() TenantSettings {
	return TenantSettings{
		Timezone:   "UTC",
		Language:   "en",
		Currency:   "USD",
		DateFormat: "YYYY-MM-DD",
		TimeFormat: "24h",
		Features: TenantFeatures{
			AIIntegration:    true,
			MultimodalAI:     false,
			CulturalWisdom:   true,
			Community:        true,
			LocationTracking: false,
			Consciousness:    false,
			Analytics:        true,
			API:              true,
			Webhooks:         false,
			CustomBranding:   false,
			SSO:              false,
			LDAP:             false,
		},
		Security: TenantSecurity{
			PasswordPolicy: PasswordPolicy{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
				ExpirationDays:   90,
				HistoryCount:     5,
			},
			SessionPolicy: SessionPolicy{
				MaxDuration:           480, // 8
				IdleTimeout:           60,  // 1
				MaxConcurrentSessions: 5,
				RequireReauth:         false,
			},
			IPWhitelist:       []string{},
			IPBlacklist:       []string{},
			TwoFactorRequired: false,
			AuditLogging:      true,
			DataRetention: DataRetentionPolicy{
				LogRetentionDays:         90,
				BackupRetentionDays:      365,
				DeletedDataRetentionDays: 30,
			},
		},
		Notifications: TenantNotifications{
			Email: EmailNotificationSettings{
				Enabled: false,
			},
			SMS: SMSNotificationSettings{
				Enabled: false,
			},
			Push: PushNotificationSettings{
				Enabled: false,
			},
			Webhook: WebhookNotificationSettings{
				Enabled: false,
				URLs:    []string{},
			},
		},
		Custom: make(map[string]interface{}),
	}
}

// IsActive ?
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsSuspended 
func (t *Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}

// IsDeleted 
func (t *Tenant) IsDeleted() bool {
	return t.Status == TenantStatusDeleted
}

// HasFeature 
func (t *Tenant) HasFeature(feature string) bool {
	switch feature {
	case "ai_integration":
		return t.Settings.Features.AIIntegration
	case "multimodal_ai":
		return t.Settings.Features.MultimodalAI
	case "cultural_wisdom":
		return t.Settings.Features.CulturalWisdom
	case "community":
		return t.Settings.Features.Community
	case "location_tracking":
		return t.Settings.Features.LocationTracking
	case "consciousness":
		return t.Settings.Features.Consciousness
	case "analytics":
		return t.Settings.Features.Analytics
	case "api":
		return t.Settings.Features.API
	case "webhooks":
		return t.Settings.Features.Webhooks
	case "custom_branding":
		return t.Settings.Features.CustomBranding
	case "sso":
		return t.Settings.Features.SSO
	case "ldap":
		return t.Settings.Features.LDAP
	default:
		return false
	}
}

// CheckQuota ?
func (t *Tenant) CheckQuota(resource string, current int) bool {
	switch resource {
	case "users":
		return current < t.Quota.MaxUsers
	case "api_calls_hour":
		return current < t.Quota.MaxAPICallsPerHour
	case "api_calls_day":
		return current < t.Quota.MaxAPICallsPerDay
	case "api_calls_month":
		return current < t.Quota.MaxAPICallsPerMonth
	case "concurrent_sessions":
		return current < t.Quota.MaxConcurrentSessions
	case "concurrent_requests":
		return current < t.Quota.MaxConcurrentRequests
	case "ai_requests_hour":
		return current < t.Quota.MaxAIRequestsPerHour
	case "ai_requests_day":
		return current < t.Quota.MaxAIRequestsPerDay
	case "ai_requests_month":
		return current < t.Quota.MaxAIRequestsPerMonth
	case "database_connections":
		return current < t.Quota.MaxDatabaseConnections
	default:
		if customQuota, exists := t.Quota.Custom[resource]; exists {
			return current < customQuota
		}
		return true
	}
}

// UpdateUsage 
func (t *Tenant) UpdateUsage(resource string, value interface{}) {
	switch resource {
	case "current_users":
		if v, ok := value.(int); ok {
			t.Usage.CurrentUsers = v
		}
	case "current_storage_gb":
		if v, ok := value.(float64); ok {
			t.Usage.CurrentStorageGB = v
		}
	case "api_calls_this_hour":
		if v, ok := value.(int); ok {
			t.Usage.APICallsThisHour = v
		}
	case "api_calls_this_day":
		if v, ok := value.(int); ok {
			t.Usage.APICallsThisDay = v
		}
	case "api_calls_this_month":
		if v, ok := value.(int); ok {
			t.Usage.APICallsThisMonth = v
		}
	case "current_sessions":
		if v, ok := value.(int); ok {
			t.Usage.CurrentSessions = v
		}
	case "current_requests":
		if v, ok := value.(int); ok {
			t.Usage.CurrentRequests = v
		}
	case "ai_requests_this_hour":
		if v, ok := value.(int); ok {
			t.Usage.AIRequestsThisHour = v
		}
	case "ai_requests_this_day":
		if v, ok := value.(int); ok {
			t.Usage.AIRequestsThisDay = v
		}
	case "ai_requests_this_month":
		if v, ok := value.(int); ok {
			t.Usage.AIRequestsThisMonth = v
		}
	case "current_database_connections":
		if v, ok := value.(int); ok {
			t.Usage.CurrentDatabaseConnections = v
		}
	default:
		if t.Usage.Custom == nil {
			t.Usage.Custom = make(map[string]interface{})
		}
		t.Usage.Custom[resource] = value
	}
	
	t.Usage.LastUpdated = time.Now()
}

// ToJSON JSON?
func (t *Tenant) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON JSON?
func (t *Tenant) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), t)
}

// TenantContext 租户上下文信息
type TenantContext struct {
	TenantID     uuid.UUID              `json:"tenant_id"`
	TenantName   string                 `json:"tenant_name"`
	UserID       uuid.UUID              `json:"user_id"`
	UserRole     string                 `json:"user_role"`
	Permissions  []string               `json:"permissions"`
	Settings     TenantSettings         `json:"settings"`
	Quota        TenantQuota            `json:"quota"`
	Usage        TenantUsage            `json:"usage"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// TenantConfig 租户配置
type TenantConfig struct {
	// 数据库配置
	DatabaseConfig DatabaseConfig `json:"database_config"`
	
	// 缓存配置
	CacheConfig CacheConfig `json:"cache_config"`
	
	// 安全配置
	SecurityConfig SecurityConfig `json:"security_config"`
	
	// 功能配置
	FeatureConfig FeatureConfig `json:"feature_config"`
	
	// 通知配置
	NotificationConfig NotificationConfig `json:"notification_config"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret           string `json:"jwt_secret"`
	TokenExpiration     int    `json:"token_expiration"`
	RefreshTokenExpiration int `json:"refresh_token_expiration"`
	PasswordMinLength   int    `json:"password_min_length"`
	MaxLoginAttempts    int    `json:"max_login_attempts"`
}

// FeatureConfig 功能配置
type FeatureConfig struct {
	EnabledFeatures []string `json:"enabled_features"`
	FeatureLimits   map[string]int `json:"feature_limits"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	EmailEnabled  bool   `json:"email_enabled"`
	SMSEnabled    bool   `json:"sms_enabled"`
	PushEnabled   bool   `json:"push_enabled"`
	WebhookEnabled bool  `json:"webhook_enabled"`
}

// DefaultTenantSettings 返回默认的租户设置
func DefaultTenantSettings() *TenantSettings {
	return &TenantSettings{
		Language:     "zh-CN",
		Timezone:     "Asia/Shanghai",
		DateFormat:   "YYYY-MM-DD",
		TimeFormat:   "HH:mm:ss",
		Currency:     "CNY",
		Theme:        "default",
		CustomFields: make(map[string]interface{}),
	}
}

// DefaultTenantQuota 返回默认的租户配额
func DefaultTenantQuota() *TenantQuota {
	return &TenantQuota{
		MaxUsers:        100,
		MaxStorage:      1024 * 1024 * 1024, // 1GB
		MaxAPIRequests:  10000,
		MaxDatabases:    5,
		MaxConnections:  50,
		MaxBandwidth:    100 * 1024 * 1024, // 100MB
		CustomLimits:    make(map[string]interface{}),
	}
}

