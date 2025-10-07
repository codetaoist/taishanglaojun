package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantStatus 租户状态
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"    // 活跃
	TenantStatusSuspended TenantStatus = "suspended" // 暂停
	TenantStatusInactive  TenantStatus = "inactive"  // 非活跃
	TenantStatusDeleted   TenantStatus = "deleted"   // 已删除
)

// IsolationStrategy 数据隔离策略
type IsolationStrategy string

const (
	IsolationRowLevel IsolationStrategy = "row_level" // 行级隔离
	IsolationSchema   IsolationStrategy = "schema"    // Schema隔离
	IsolationDatabase IsolationStrategy = "database"  // 数据库隔离
)

// TenantResolution 租户识别策略
type TenantResolution string

const (
	ResolutionHeader    TenantResolution = "header"    // HTTP头部
	ResolutionSubdomain TenantResolution = "subdomain" // 子域名
	ResolutionPath      TenantResolution = "path"      // URL路径
)

// Tenant 租户模型
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
	
	// 数据隔离配置
	IsolationStrategy IsolationStrategy `gorm:"type:varchar(20);default:'row_level'" json:"isolation_strategy"`
	DatabaseName      string            `gorm:"type:varchar(100)" json:"database_name,omitempty"`
	SchemaName        string            `gorm:"type:varchar(100)" json:"schema_name,omitempty"`
	
	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 关联关系
	Users         []TenantUser         `gorm:"foreignKey:TenantID" json:"users,omitempty"`
	Subscriptions []TenantSubscription `gorm:"foreignKey:TenantID" json:"subscriptions,omitempty"`
}

// TenantSettings 租户设置
type TenantSettings struct {
	// 基础设置
	Timezone     string `json:"timezone"`
	Language     string `json:"language"`
	Currency     string `json:"currency"`
	DateFormat   string `json:"date_format"`
	TimeFormat   string `json:"time_format"`
	
	// 功能开关
	Features TenantFeatures `json:"features"`
	
	// 安全设置
	Security TenantSecurity `json:"security"`
	
	// 通知设置
	Notifications TenantNotifications `json:"notifications"`
	
	// 自定义设置
	Custom map[string]interface{} `json:"custom"`
}

// TenantFeatures 租户功能开关
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

// TenantSecurity 租户安全设置
type TenantSecurity struct {
	PasswordPolicy    PasswordPolicy    `json:"password_policy"`
	SessionPolicy     SessionPolicy     `json:"session_policy"`
	IPWhitelist       []string          `json:"ip_whitelist"`
	IPBlacklist       []string          `json:"ip_blacklist"`
	TwoFactorRequired bool              `json:"two_factor_required"`
	AuditLogging      bool              `json:"audit_logging"`
	DataRetention     DataRetentionPolicy `json:"data_retention"`
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSymbols   bool `json:"require_symbols"`
	ExpirationDays   int  `json:"expiration_days"`
	HistoryCount     int  `json:"history_count"`
}

// SessionPolicy 会话策略
type SessionPolicy struct {
	MaxDuration       int  `json:"max_duration"`        // 最大会话时长（分钟）
	IdleTimeout       int  `json:"idle_timeout"`        // 空闲超时（分钟）
	MaxConcurrentSessions int `json:"max_concurrent_sessions"` // 最大并发会话数
	RequireReauth     bool `json:"require_reauth"`      // 敏感操作需要重新认证
}

// DataRetentionPolicy 数据保留策略
type DataRetentionPolicy struct {
	LogRetentionDays    int `json:"log_retention_days"`
	BackupRetentionDays int `json:"backup_retention_days"`
	DeletedDataRetentionDays int `json:"deleted_data_retention_days"`
}

// TenantNotifications 租户通知设置
type TenantNotifications struct {
	Email    EmailNotificationSettings    `json:"email"`
	SMS      SMSNotificationSettings      `json:"sms"`
	Push     PushNotificationSettings     `json:"push"`
	Webhook  WebhookNotificationSettings  `json:"webhook"`
}

// EmailNotificationSettings 邮件通知设置
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

// SMSNotificationSettings 短信通知设置
type SMSNotificationSettings struct {
	Enabled   bool   `json:"enabled"`
	Provider  string `json:"provider"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	FromNumber string `json:"from_number"`
}

// PushNotificationSettings 推送通知设置
type PushNotificationSettings struct {
	Enabled bool   `json:"enabled"`
	FCMKey  string `json:"fcm_key"`
	APNSKey string `json:"apns_key"`
}

// WebhookNotificationSettings Webhook通知设置
type WebhookNotificationSettings struct {
	Enabled bool     `json:"enabled"`
	URLs    []string `json:"urls"`
	Secret  string   `json:"secret"`
}

// TenantQuota 租户配额
type TenantQuota struct {
	// 用户配额
	MaxUsers int `json:"max_users"`
	
	// 存储配额
	MaxStorageGB int `json:"max_storage_gb"`
	
	// API配额
	MaxAPICallsPerHour   int `json:"max_api_calls_per_hour"`
	MaxAPICallsPerDay    int `json:"max_api_calls_per_day"`
	MaxAPICallsPerMonth  int `json:"max_api_calls_per_month"`
	
	// 并发配额
	MaxConcurrentSessions int `json:"max_concurrent_sessions"`
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
	
	// AI配额
	MaxAIRequestsPerHour  int `json:"max_ai_requests_per_hour"`
	MaxAIRequestsPerDay   int `json:"max_ai_requests_per_day"`
	MaxAIRequestsPerMonth int `json:"max_ai_requests_per_month"`
	
	// 文件上传配额
	MaxFileSize       int `json:"max_file_size"`        // 单个文件最大大小（MB）
	MaxFilesPerUpload int `json:"max_files_per_upload"` // 单次上传最大文件数
	
	// 数据库配额
	MaxDatabaseConnections int `json:"max_database_connections"`
	MaxQueryTimeout        int `json:"max_query_timeout"` // 查询超时时间（秒）
	
	// 自定义配额
	Custom map[string]int `json:"custom"`
}

// TenantUsage 租户使用情况
type TenantUsage struct {
	// 用户使用情况
	CurrentUsers int `json:"current_users"`
	
	// 存储使用情况
	CurrentStorageGB float64 `json:"current_storage_gb"`
	
	// API使用情况
	APICallsThisHour  int `json:"api_calls_this_hour"`
	APICallsThisDay   int `json:"api_calls_this_day"`
	APICallsThisMonth int `json:"api_calls_this_month"`
	
	// 并发使用情况
	CurrentSessions int `json:"current_sessions"`
	CurrentRequests int `json:"current_requests"`
	
	// AI使用情况
	AIRequestsThisHour  int `json:"ai_requests_this_hour"`
	AIRequestsThisDay   int `json:"ai_requests_this_day"`
	AIRequestsThisMonth int `json:"ai_requests_this_month"`
	
	// 数据库使用情况
	CurrentDatabaseConnections int `json:"current_database_connections"`
	
	// 最后更新时间
	LastUpdated time.Time `json:"last_updated"`
	
	// 自定义使用情况
	Custom map[string]interface{} `json:"custom"`
}

// TenantUser 租户用户关联
type TenantUser struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:char(36);not null;index" json:"tenant_id"`
	UserID   uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"`
	Role     string    `gorm:"type:varchar(50);not null" json:"role"`
	Status   string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	
	// 权限
	Permissions []string `gorm:"type:json" json:"permissions"`
	
	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 关联关系
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// TenantSubscription 租户订阅
type TenantSubscription struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:char(36);not null;index" json:"tenant_id"`
	
	// 订阅信息
	PlanID      string    `gorm:"type:varchar(100);not null" json:"plan_id"`
	PlanName    string    `gorm:"type:varchar(255)" json:"plan_name"`
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	TrialEnd    *time.Time `json:"trial_end,omitempty"`
	
	// 计费信息
	BillingCycle string  `gorm:"type:varchar(20)" json:"billing_cycle"` // monthly, yearly
	Amount       float64 `json:"amount"`
	Currency     string  `gorm:"type:varchar(10)" json:"currency"`
	
	// 外部订阅ID（如Stripe订阅ID）
	ExternalSubscriptionID string `gorm:"type:varchar(255)" json:"external_subscription_id"`
	
	// 时间戳
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	
	// 关联关系
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// BeforeCreate 创建前钩子
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	
	// 设置默认配额
	if t.Quota.MaxUsers == 0 {
		t.Quota = GetDefaultQuota()
	}
	
	// 设置默认设置
	if t.Settings.Timezone == "" {
		t.Settings = GetDefaultSettings()
	}
	
	// 初始化使用情况
	t.Usage = TenantUsage{
		LastUpdated: time.Now(),
		Custom:      make(map[string]interface{}),
	}
	
	return nil
}

// BeforeCreate 创建前钩子 - TenantUser
func (tu *TenantUser) BeforeCreate(tx *gorm.DB) error {
	if tu.ID == uuid.Nil {
		tu.ID = uuid.New()
	}
	return nil
}

// BeforeCreate 创建前钩子 - TenantSubscription
func (ts *TenantSubscription) BeforeCreate(tx *gorm.DB) error {
	if ts.ID == uuid.Nil {
		ts.ID = uuid.New()
	}
	return nil
}

// GetDefaultQuota 获取默认配额
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

// GetDefaultSettings 获取默认设置
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
				MaxDuration:           480, // 8小时
				IdleTimeout:           60,  // 1小时
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

// IsActive 检查租户是否活跃
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsSuspended 检查租户是否被暂停
func (t *Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}

// IsDeleted 检查租户是否被删除
func (t *Tenant) IsDeleted() bool {
	return t.Status == TenantStatusDeleted
}

// HasFeature 检查租户是否启用了指定功能
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

// CheckQuota 检查是否超过配额
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

// UpdateUsage 更新使用情况
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

// ToJSON 转换为JSON字符串
func (t *Tenant) ToJSON() (string, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (t *Tenant) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), t)
}