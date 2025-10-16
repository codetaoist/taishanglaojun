package models

import (
	"time"
)

// APIKey API密钥模型
type APIKey struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Key         string    `json:"key" db:"key"`
	SecretHash  string    `json:"-" db:"secret_hash"`
	Permissions []string  `json:"permissions" db:"permissions"`
	RateLimit   int       `json:"rate_limit" db:"rate_limit"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Plugin 插件模型
type Plugin struct {
	ID          int64                  `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Version     string                 `json:"version" db:"version"`
	Description string                 `json:"description" db:"description"`
	Author      string                 `json:"author" db:"author"`
	Homepage    string                 `json:"homepage" db:"homepage"`
	Repository  string                 `json:"repository" db:"repository"`
	License     string                 `json:"license" db:"license"`
	Tags        []string               `json:"tags" db:"tags"`
	Category    string                 `json:"category" db:"category"`
	Status      PluginStatus           `json:"status" db:"status"`
	Config      map[string]interface{} `json:"config" db:"config"`
	Manifest    PluginManifest         `json:"manifest" db:"manifest"`
	InstallPath string                 `json:"install_path" db:"install_path"`
	IsEnabled   bool                   `json:"is_enabled" db:"is_enabled"`
	InstallDate time.Time              `json:"install_date" db:"install_date"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// PluginStatus 插件状?type PluginStatus string

const (
	PluginStatusInstalled   PluginStatus = "installed"
	PluginStatusEnabled     PluginStatus = "enabled"
	PluginStatusDisabled    PluginStatus = "disabled"
	PluginStatusError       PluginStatus = "error"
	PluginStatusUpdating    PluginStatus = "updating"
	PluginStatusUninstalling PluginStatus = "uninstalling"
)

// PluginManifest 插件清单
type PluginManifest struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Main         string                 `json:"main"`
	Dependencies map[string]string      `json:"dependencies"`
	Permissions  []string               `json:"permissions"`
	Hooks        []string               `json:"hooks"`
	Config       map[string]interface{} `json:"config"`
	MinVersion   string                 `json:"min_version"`
	MaxVersion   string                 `json:"max_version"`
}

// Integration 第三方服务集成模?type Integration struct {
	ID           int64                  `json:"id" db:"id"`
	UserID       int64                  `json:"user_id" db:"user_id"`
	Name         string                 `json:"name" db:"name"`
	Provider     string                 `json:"provider" db:"provider"`
	Type         IntegrationType        `json:"type" db:"type"`
	Status       IntegrationStatus      `json:"status" db:"status"`
	Config       map[string]interface{} `json:"config" db:"config"`
	Credentials  map[string]interface{} `json:"-" db:"credentials"`
	Settings     map[string]interface{} `json:"settings" db:"settings"`
	LastSyncAt   *time.Time             `json:"last_sync_at" db:"last_sync_at"`
	SyncInterval int                    `json:"sync_interval" db:"sync_interval"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// IntegrationType 集成类型
type IntegrationType string

const (
	IntegrationTypeAPI      IntegrationType = "api"
	IntegrationTypeWebhook  IntegrationType = "webhook"
	IntegrationTypeOAuth    IntegrationType = "oauth"
	IntegrationTypeDatabase IntegrationType = "database"
	IntegrationTypeFile     IntegrationType = "file"
)

// IntegrationStatus 集成状?type IntegrationStatus string

const (
	IntegrationStatusActive      IntegrationStatus = "active"
	IntegrationStatusInactive    IntegrationStatus = "inactive"
	IntegrationStatusError       IntegrationStatus = "error"
	IntegrationStatusSyncing     IntegrationStatus = "syncing"
	IntegrationStatusConfiguring IntegrationStatus = "configuring"
)

// Webhook Webhook模型
type Webhook struct {
	ID          int64                  `json:"id" db:"id"`
	UserID      int64                  `json:"user_id" db:"user_id"`
	Name        string                 `json:"name" db:"name"`
	URL         string                 `json:"url" db:"url"`
	Token       string                 `json:"token" db:"token"`
	Secret      string                 `json:"-" db:"secret"`
	Events      []string               `json:"events" db:"events"`
	Headers     map[string]string      `json:"headers" db:"headers"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	RetryCount  int                    `json:"retry_count" db:"retry_count"`
	MaxRetries  int                    `json:"max_retries" db:"max_retries"`
	LastTrigger *time.Time             `json:"last_trigger" db:"last_trigger"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// OAuthToken OAuth令牌模型
type OAuthToken struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Provider     string    `json:"provider" db:"provider"`
	AccessToken  string    `json:"-" db:"access_token"`
	RefreshToken string    `json:"-" db:"refresh_token"`
	TokenType    string    `json:"token_type" db:"token_type"`
	Scope        string    `json:"scope" db:"scope"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// WebhookEvent Webhook事件模型
type WebhookEvent struct {
	ID          int64                  `json:"id" db:"id"`
	WebhookID   int64                  `json:"webhook_id" db:"webhook_id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Status      WebhookEventStatus     `json:"status" db:"status"`
	Response    string                 `json:"response" db:"response"`
	RetryCount  int                    `json:"retry_count" db:"retry_count"`
	ProcessedAt *time.Time             `json:"processed_at" db:"processed_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// WebhookEventStatus Webhook事件状?type WebhookEventStatus string

const (
	WebhookEventStatusPending    WebhookEventStatus = "pending"
	WebhookEventStatusProcessing WebhookEventStatus = "processing"
	WebhookEventStatusSuccess    WebhookEventStatus = "success"
	WebhookEventStatusFailed     WebhookEventStatus = "failed"
	WebhookEventStatusRetrying   WebhookEventStatus = "retrying"
)

// 数据库表创建SQL
const (
	APIKeyTableSQL = `
	CREATE TABLE IF NOT EXISTS api_keys (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL,
		key VARCHAR(255) UNIQUE NOT NULL,
		secret_hash VARCHAR(255) NOT NULL,
		permissions TEXT[],
		rate_limit INTEGER DEFAULT 1000,
		is_active BOOLEAN DEFAULT true,
		last_used_at TIMESTAMP,
		expires_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
	CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key);
	`

	PluginTableSQL = `
	CREATE TABLE IF NOT EXISTS plugins (
		id BIGSERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		version VARCHAR(50) NOT NULL,
		description TEXT,
		author VARCHAR(255),
		homepage VARCHAR(500),
		repository VARCHAR(500),
		license VARCHAR(100),
		tags TEXT[],
		category VARCHAR(100),
		status VARCHAR(50) DEFAULT 'installed',
		config JSONB,
		manifest JSONB,
		install_path VARCHAR(500),
		is_enabled BOOLEAN DEFAULT false,
		install_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_plugins_name ON plugins(name);
	CREATE INDEX IF NOT EXISTS idx_plugins_category ON plugins(category);
	`

	IntegrationTableSQL = `
	CREATE TABLE IF NOT EXISTS integrations (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL,
		provider VARCHAR(100) NOT NULL,
		type VARCHAR(50) NOT NULL,
		status VARCHAR(50) DEFAULT 'inactive',
		config JSONB,
		credentials JSONB,
		settings JSONB,
		last_sync_at TIMESTAMP,
		sync_interval INTEGER DEFAULT 3600,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_integrations_user_id ON integrations(user_id);
	CREATE INDEX IF NOT EXISTS idx_integrations_provider ON integrations(provider);
	`

	WebhookTableSQL = `
	CREATE TABLE IF NOT EXISTS webhooks (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(500) NOT NULL,
		token VARCHAR(255) UNIQUE NOT NULL,
		secret VARCHAR(255),
		events TEXT[],
		headers JSONB,
		payload JSONB,
		is_active BOOLEAN DEFAULT true,
		retry_count INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 3,
		last_trigger TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_webhooks_user_id ON webhooks(user_id);
	CREATE INDEX IF NOT EXISTS idx_webhooks_token ON webhooks(token);
	`

	OAuthTokenTableSQL = `
	CREATE TABLE IF NOT EXISTS oauth_tokens (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		provider VARCHAR(100) NOT NULL,
		access_token TEXT NOT NULL,
		refresh_token TEXT,
		token_type VARCHAR(50) DEFAULT 'Bearer',
		scope TEXT,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, provider)
	);
	CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_id ON oauth_tokens(user_id);
	CREATE INDEX IF NOT EXISTS idx_oauth_tokens_provider ON oauth_tokens(provider);
	`

	WebhookEventTableSQL = `
	CREATE TABLE IF NOT EXISTS webhook_events (
		id BIGSERIAL PRIMARY KEY,
		webhook_id BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
		event_type VARCHAR(100) NOT NULL,
		payload JSONB,
		status VARCHAR(50) DEFAULT 'pending',
		response TEXT,
		retry_count INTEGER DEFAULT 0,
		processed_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_webhook_events_webhook_id ON webhook_events(webhook_id);
	CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON webhook_events(status);
	`
)

