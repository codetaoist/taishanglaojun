package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Laojun Domain Models

type Config struct {
	Key       string     `json:"key" db:"key"`
	Value     JSONB      `json:"value" db:"value"`
	Scope     string     `json:"scope" db:"scope"`
	TenantID  string     `json:"tenantId" db:"tenant_id"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

type Plugin struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenantId" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Version     string    `json:"version" db:"version"`
	Source      string    `json:"source" db:"source"`
	Status      string    `json:"status" db:"status"`
	Config      JSONB     `json:"config" db:"config"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type PluginVersion struct {
	ID        int       `json:"id" db:"id"`
	PluginID  string    `json:"pluginId" db:"plugin_id"`
	Version   string    `json:"version" db:"version"`
	Manifest  JSONB     `json:"manifest" db:"manifest"`
	Signature string    `json:"signature" db:"signature"`
	PackageURL string   `json:"packageUrl" db:"package_url"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type AuditLog struct {
	ID         int       `json:"id" db:"id"`
	TenantID   string    `json:"tenantId" db:"tenant_id"`
	Actor      string    `json:"actor" db:"actor"`
	ActorType  string    `json:"actorType" db:"actor_type"`
	Action     string    `json:"action" db:"action"`
	TargetType string    `json:"targetType" db:"target_type"`
	TargetID   string    `json:"targetId" db:"target_id"`
	Resource   string    `json:"resource" db:"resource"`
	ResourceID string    `json:"resourceId" db:"resource_id"`
	Details    JSONB     `json:"details" db:"details"`
	Payload    JSONB     `json:"payload" db:"payload"`
	Result     string    `json:"result" db:"result"`
	IPAddress  string    `json:"ipAddress" db:"ip_address"`
	UserAgent  string    `json:"userAgent" db:"user_agent"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

// Taishang Domain Models

type Model struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenantId" db:"tenant_id"`
	Name      string    `json:"name" db:"name"`
	Provider  string    `json:"provider" db:"provider"`
	Version   string    `json:"version" db:"version"`
	Status    string    `json:"status" db:"status"`
	Meta      JSONB     `json:"meta" db:"meta"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type VectorCollection struct {
	ID               int       `json:"id" db:"id"`
	TenantID         string    `json:"tenantId" db:"tenant_id"`
	Name             string    `json:"name" db:"name"`
	ModelID          string    `json:"modelId" db:"model_id"`
	Dims             int       `json:"dims" db:"dims"`
	IndexType        string    `json:"indexType" db:"index_type"`
	MetricType       string    `json:"metricType" db:"metric_type"`
	ExtraIndexArgs   JSONB     `json:"extraIndexArgs" db:"extra_index_args"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at"`
}

// TableName returns the table name for VectorCollection
func (VectorCollection) TableName() string {
	return "tai_vector_collections"
}

type Vector struct {
	ID           int       `json:"id" db:"id"`
	TenantID     string    `json:"tenantId" db:"tenant_id"`
	CollectionID int       `json:"collectionId" db:"collection_id"`
	ExternalID   string    `json:"externalId" db:"external_id"`
	Embedding    []float64 `json:"embedding" db:"embedding"`
	Metadata     JSONB     `json:"metadata" db:"metadata"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSuccess   TaskStatus = "success"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityNormal TaskPriority = "normal"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent  TaskPriority = "urgent"
)

type Task struct {
	ID        int64      `json:"id" db:"id"`
	TenantID  string     `json:"tenantId" db:"tenant_id"`
	Type      string     `json:"type" db:"type"`
	Status    TaskStatus `json:"status" db:"status"`
	Priority  TaskPriority `json:"priority" db:"priority"`
	Payload   JSONB      `json:"payload" db:"payload"`
	Result    JSONB      `json:"result" db:"result"`
	WorkerID  string     `json:"workerId" db:"worker_id"`
	StartedAt *time.Time `json:"startedAt" db:"started_at"`
	FinishedAt *time.Time `json:"finishedAt" db:"finished_at"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// JSONB is a custom type for handling PostgreSQL JSONB data
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, j)
}