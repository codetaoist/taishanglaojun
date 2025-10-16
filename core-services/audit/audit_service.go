package audit

import (
	"context"
	"time"
)

// AuditService 审计服务接口
type AuditService interface {
	// LogEvent 记录单个审计事件
	LogEvent(ctx context.Context, event *AuditEvent) error

	// LogEvents 批量记录审计事件
	LogEvents(ctx context.Context, events []*AuditEvent) error

	// QueryLogs 查询审计日志
	QueryLogs(ctx context.Context, query *AuditQuery) (*AuditLogResponse, error)

	// GetStatistics 获取审计统计信息
	GetStatistics(ctx context.Context, filter *StatisticsFilter) (*AuditStatistics, error)

	// ExportLogs 导出审计日志
	ExportLogs(ctx context.Context, request *ExportRequest) (*ExportResponse, error)

	// CleanupLogs 清理过期的审计日志
	CleanupLogs(ctx context.Context, retentionPolicy *RetentionPolicy) (int64, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
}

// AuditEvent 审计事件
// 记录系统中的关键操作事件，如用户登录、资源变更、系统配置修改等
type AuditEvent struct {
	// ID 事件唯一标识符
	ID        string    `json:"id" db:"id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`

	// EventType 事件类型
	EventType    EventType `json:"event_type" db:"event_type"`
	EventAction  string    `json:"event_action" db:"event_action"`
	EventResult  string    `json:"event_result" db:"event_result"`
	EventMessage string    `json:"event_message" db:"event_message"`

	// UserID 用户ID
	UserID    string `json:"user_id" db:"user_id"`
	UserName  string `json:"user_name" db:"user_name"`
	UserRole  string `json:"user_role" db:"user_role"`
	UserEmail string `json:"user_email" db:"user_email"`

	// TenantID 租户ID
	TenantID   string `json:"tenant_id" db:"tenant_id"`
	TenantName string `json:"tenant_name" db:"tenant_name"`

	// ResourceType 资源类型
	ResourceType string `json:"resource_type" db:"resource_type"`
	ResourceID   string `json:"resource_id" db:"resource_id"`
	ResourceName string `json:"resource_name" db:"resource_name"`

	// RequestID 请求ID
	RequestID     string            `json:"request_id" db:"request_id"`
	SessionID     string            `json:"session_id" db:"session_id"`
	IPAddress     string            `json:"ip_address" db:"ip_address"`
	UserAgent     string            `json:"user_agent" db:"user_agent"`
	RequestMethod string            `json:"request_method" db:"request_method"`
	RequestPath   string            `json:"request_path" db:"request_path"`
	RequestURL    string            `json:"request_url" db:"request_url"`
	RequestParams map[string]string `json:"request_params" db:"request_params"`

	// ResponseCode 响应状态码
	ResponseCode   int           `json:"response_code" db:"response_code"`
	ResponseStatus string        `json:"response_status" db:"response_status"`
	ResponseTime   time.Duration `json:"response_time" db:"response_time"`
	ResponseSize   int64         `json:"response_size" db:"response_size"`

	// Changes 变更字段
	Changes   []FieldChange          `json:"changes" db:"changes"`
	OldValues map[string]interface{} `json:"old_values" db:"old_values"`
	NewValues map[string]interface{} `json:"new_values" db:"new_values"`

	// SecurityLevel 安全等级
	SecurityLevel SecurityLevel `json:"security_level" db:"security_level"`
	RiskScore     float64       `json:"risk_score" db:"risk_score"`
	Anomaly       bool          `json:"anomaly" db:"anomaly"`

	// Metadata 元数据
	Metadata map[string]interface{} `json:"metadata" db:"metadata"`
	Tags     []string               `json:"tags" db:"tags"`

	// ServiceName 服务名称
	ServiceName    string `json:"service_name" db:"service_name"`
	ServiceVersion string `json:"service_version" db:"service_version"`
	Environment    string `json:"environment" db:"environment"`

	// ComplianceFlags 合规性标志
	ComplianceFlags    []string `json:"compliance_flags" db:"compliance_flags"`
	ComplianceTags     []string `json:"compliance_tags" db:"compliance_tags"`
	DataClassification string   `json:"data_classification" db:"data_classification"`

	// EventCategory 事件分类
	EventCategory string `json:"event_category" db:"event_category"`

	// CorrelationID 关联ID
	CorrelationID string `json:"correlation_id" db:"correlation_id"`

	// SourceSystem 源系统
	SourceSystem    string `json:"source_system" db:"source_system"`
	SourceComponent string `json:"source_component" db:"source_component"`
}

// EventType 事件类型
type EventType string

const (
	// 认证事件
	EventTypeAuth EventType = "auth"

	// 授权事件
	EventTypeAuthz EventType = "authz"

	// 数据访问事件
	EventTypeDataAccess EventType = "data_access"

	// 数据修改事件
	EventTypeDataModification EventType = "data_modification"

	// 系统事件
	EventTypeSystem EventType = "system"

	// 配置事件
	EventTypeConfiguration EventType = "configuration"

	// 安全事件
	EventTypeSecurity EventType = "security"

	// 业务事件
	EventTypeBusiness EventType = "business"

	// API调用事件
	EventTypeAPI EventType = "api"

	// 文件事件
	EventTypeFile EventType = "file"

	// 网络事件
	EventTypeNetwork EventType = "network"

	// 错误事件
	EventTypeError EventType = "error"
)

// 具体的事件类型常量
const (
	// 用户认证相关事件
	EventTypeUserLogin  EventType = "user_login"
	EventTypeUserLogout EventType = "user_logout"

	// 用户管理相关事件
	EventTypeUserCreate EventType = "user_create"
	EventTypeUserUpdate EventType = "user_update"
	EventTypeUserDelete EventType = "user_delete"

	// 数据操作相关事件
	EventTypeDataCreate EventType = "data_create"
	EventTypeDataUpdate EventType = "data_update"
	EventTypeDataDelete EventType = "data_delete"

	// API调用事件
	EventTypeAPICall EventType = "api_call"
)

// SecurityLevel 安全等级
type SecurityLevel string

const (
	SecurityLevelLow      SecurityLevel = "low"
	SecurityLevelMedium   SecurityLevel = "medium"
	SecurityLevelHigh     SecurityLevel = "high"
	SecurityLevelCritical SecurityLevel = "critical"
)

// FieldChange 字段变更
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Action   string      `json:"action"` // create, update, delete
}

// AuditQuery 查询参数
type AuditQuery struct {
	// 时间范围
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`

	// 事件类型和操作
	EventTypes   []EventType `json:"event_types"`
	EventActions []string    `json:"event_actions"`
	EventResults []string    `json:"event_results"`

	// 用户相关
	UserIDs   []string `json:"user_ids"`
	UserRoles []string `json:"user_roles"`

	// 租户相关
	TenantIDs []string `json:"tenant_ids"`

	// 资源相关
	ResourceTypes []string `json:"resource_types"`
	ResourceIDs   []string `json:"resource_ids"`

	// 安全相关
	SecurityLevels []SecurityLevel `json:"security_levels"`
	MinRiskScore   *float64        `json:"min_risk_score"`
	MaxRiskScore   *float64        `json:"max_risk_score"`
	AnomalyOnly    bool            `json:"anomaly_only"`

	// IP地址相关
	IPAddresses []string `json:"ip_addresses"`
	IPRanges    []string `json:"ip_ranges"`

	// 响应状态码相关
	ResponseCodes []int `json:"response_codes"`

	// 搜索文本
	SearchText string `json:"search_text"`

	// 标签相关
	Tags []string `json:"tags"`

	// 合规性标志相关
	ComplianceFlags []string `json:"compliance_flags"`

	// 排序相关
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // asc, desc

	// 分页相关
	Page     int `json:"page"`
	PageSize int `json:"page_size"`

	// 分组聚合相关
	GroupBy    []string `json:"group_by"`
	Aggregates []string `json:"aggregates"` // count, sum, avg, min, max
}

// AuditLogResponse 审计日志响应
type AuditLogResponse struct {
	Events     []*AuditEvent          `json:"events"`
	Pagination PaginationResponse     `json:"pagination"`
	Aggregates map[string]interface{} `json:"aggregates,omitempty"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// StatisticsFilter 统计过滤参数
type StatisticsFilter struct {
	StartTime   *time.Time  `json:"start_time"`
	EndTime     *time.Time  `json:"end_time"`
	TenantIDs   []string    `json:"tenant_ids"`
	EventTypes  []EventType `json:"event_types"`
	UserIDs     []string    `json:"user_ids"`
	Granularity string      `json:"granularity"` // hour, day, week, month
}

// AuditStatistics 审计统计
type AuditStatistics struct {
	TotalEvents        int64               `json:"total_events"`
	EventsByType       map[EventType]int64 `json:"events_by_type"`
	EventsByResult     map[string]int64    `json:"events_by_result"`
	EventsByUser       map[string]int64    `json:"events_by_user"`
	EventsByTenant     map[string]int64    `json:"events_by_tenant"`
	EventsByHour       map[string]int64    `json:"events_by_hour"`
	TopUsers           []UserActivity      `json:"top_users"`
	TopResources       []ResourceActivity  `json:"top_resources"`
	SecurityEvents     SecurityStatistics  `json:"security_events"`
	AnomalyCount       int64               `json:"anomaly_count"`
	FailureRate        float64             `json:"failure_rate"`
	AverageRiskScore   float64             `json:"average_risk_score"`
	UserActivities     []UserActivity      `json:"user_activities"`
	ResourceActivities []ResourceActivity  `json:"resource_activities"`
	SecurityStats      SecurityStatistics  `json:"security_stats"`
}

// UserActivity 用户活动
type UserActivity struct {
	UserID   string    `json:"user_id"`
	UserName string    `json:"user_name"`
	Count    int64     `json:"count"`
	LastSeen time.Time `json:"last_seen"`
}

// ResourceActivity 资源活动
type ResourceActivity struct {
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	ResourceName string    `json:"resource_name"`
	Count        int64     `json:"count"`
	LastAccessed time.Time `json:"last_accessed"`
}

// SecurityStatistics 安全统计
type SecurityStatistics struct {
	HighRiskEvents     int64   `json:"high_risk_events"`
	CriticalEvents     int64   `json:"critical_events"`
	FailedLogins       int64   `json:"failed_logins"`
	UnauthorizedAccess int64   `json:"unauthorized_access"`
	SuspiciousActivity int64   `json:"suspicious_activity"`
	AverageRiskScore   float64 `json:"average_risk_score"`
}

// ExportRequest 导出请求
type ExportRequest struct {
	Query  *AuditQuery `json:"query"`
	Format string      `json:"format"` // json, csv, xlsx, pdf
	Fields []string    `json:"fields"`

	// 压缩导出文件
	Compress bool   `json:"compress"`
	Password string `json:"password,omitempty"`

	// 分块导出文件大小（MB）
	SplitSize int64 `json:"split_size,omitempty"` //

	// 导出文件标题和描述
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	RequestedBy string `json:"requested_by"`
}

// ExportResponse 导出响应
type ExportResponse struct {
	ExportID    string     `json:"export_id"`
	Status      string     `json:"status"` // pending, processing, completed, failed
	Format      string     `json:"format"`
	FileSize    int64      `json:"file_size"`
	RecordCount int64      `json:"record_count"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DownloadURL string     `json:"download_url,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// RetentionPolicy 保留策略
type RetentionPolicy struct {
	// 默认保留时间
	DefaultRetention time.Duration `json:"default_retention"`

	// 事件类型保留时间
	EventTypeRetention map[EventType]time.Duration `json:"event_type_retention"`

	// 安全等级保留时间
	SecurityLevelRetention map[SecurityLevel]time.Duration `json:"security_level_retention"`

	// 租户保留时间
	TenantRetention map[string]time.Duration `json:"tenant_retention"`

	// 归档事件
	ArchiveEnabled bool          `json:"archive_enabled"`
	ArchiveAfter   time.Duration `json:"archive_after"`
	ArchiveStorage string        `json:"archive_storage"` // s3, gcs, azure

	// 硬删除事件时间
	HardDeleteAfter time.Duration `json:"hard_delete_after"`

	// 批量删除事件数量
	BatchSize int `json:"batch_size"`
}

// AuditRepository 审计事件存储库
type AuditRepository interface {
	// 保存审计事件
	SaveEvent(ctx context.Context, event *AuditEvent) error

	// 批量保存审计事件
	SaveEvents(ctx context.Context, events []*AuditEvent) error

	// 查询审计事件
	QueryEvents(ctx context.Context, query *AuditQuery) ([]*AuditEvent, int64, error)

	// 获取审计统计
	GetStatistics(ctx context.Context, filter *StatisticsFilter) (*AuditStatistics, error)

	// 删除过期事件
	DeleteExpiredEvents(ctx context.Context, before time.Time, batchSize int) (int64, error)

	// 归档事件
	ArchiveEvents(ctx context.Context, before time.Time, batchSize int) (int64, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// AuditEventPublisher 审计事件发布器
type AuditEventPublisher interface {
	// 发布审计事件
	PublishEvent(ctx context.Context, event *AuditEvent) error

	// 批量发布审计事件
	PublishEvents(ctx context.Context, events []*AuditEvent) error

	// 订阅审计事件
	Subscribe(ctx context.Context, handler AuditEventHandler) error

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// AuditEventHandler 审计事件处理函数
type AuditEventHandler func(ctx context.Context, event *AuditEvent) error

// AuditError 审计错误
type AuditError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AuditError) Error() string {
	return e.Message
}

// AuditEvent 审计事件
type AuditEvent struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	EventType       EventType              `json:"event_type"`
	EventAction     string                 `json:"event_action"`
	EventResult     string                 `json:"event_result"`
	SecurityLevel   SecurityLevel          `json:"security_level"`
	UserID          string                 `json:"user_id,omitempty"`
	UserName        string                 `json:"user_name,omitempty"`
	UserRole        string                 `json:"user_role,omitempty"`
	UserEmail       string                 `json:"user_email,omitempty"`
	TenantID        string                 `json:"tenant_id,omitempty"`
	TenantName      string                 `json:"tenant_name,omitempty"`
	ResourceType    string                 `json:"resource_type,omitempty"`
	ResourceID      string                 `json:"resource_id,omitempty"`
	ResourceName    string                 `json:"resource_name,omitempty"`
	RequestID       string                 `json:"request_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	UserAgent       string                 `json:"user_agent,omitempty"`
	RequestMethod   string                 `json:"request_method,omitempty"`
	RequestPath     string                 `json:"request_path,omitempty"`
	ResponseCode    int                    `json:"response_code,omitempty"`
	ResponseTime    time.Duration          `json:"response_time,omitempty"`
	ResponseSize    int64                  `json:"response_size,omitempty"`
	Changes         []FieldChange          `json:"changes,omitempty"`
	OldValues       map[string]interface{} `json:"old_values,omitempty"`
	NewValues       map[string]interface{} `json:"new_values,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	ComplianceFlags []string               `json:"compliance_flags,omitempty"`
}

// NewAuditEvent 创建新的审计事件
func NewAuditEvent(eventType EventType, action string) *AuditEvent {
	return &AuditEvent{
		ID:              generateEventID(),
		Timestamp:       time.Now(),
		EventType:       eventType,
		EventAction:     action,
		EventResult:     "success",
		SecurityLevel:   SecurityLevelLow,
		Metadata:        make(map[string]interface{}),
		Tags:            make([]string, 0),
		ComplianceFlags: make([]string, 0),
	}
}

// WithUser 设置用户信息
func (e *AuditEvent) WithUser(userID, userName, userRole, userEmail string) *AuditEvent {
	e.UserID = userID
	e.UserName = userName
	e.UserRole = userRole
	e.UserEmail = userEmail
	return e
}

// WithTenant 设置租户信息
func (e *AuditEvent) WithTenant(tenantID, tenantName string) *AuditEvent {
	e.TenantID = tenantID
	e.TenantName = tenantName
	return e
}

// WithResource 设置资源信息
func (e *AuditEvent) WithResource(resourceType, resourceID, resourceName string) *AuditEvent {
	e.ResourceType = resourceType
	e.ResourceID = resourceID
	e.ResourceName = resourceName
	return e
}

// WithRequest 设置请求信息
func (e *AuditEvent) WithRequest(requestID, sessionID, ipAddress, userAgent, method, path string) *AuditEvent {
	e.RequestID = requestID
	e.SessionID = sessionID
	e.IPAddress = ipAddress
	e.UserAgent = userAgent
	e.RequestMethod = method
	e.RequestPath = path
	return e
}

// WithResponse 设置响应信息
func (e *AuditEvent) WithResponse(code int, responseTime time.Duration, size int64) *AuditEvent {
	e.ResponseCode = code
	e.ResponseTime = responseTime
	e.ResponseSize = size
	return e
}

// WithSecurity 设置安全信息
func (e *AuditEvent) WithSecurity(level SecurityLevel, riskScore float64, anomaly bool) *AuditEvent {
	e.SecurityLevel = level
	e.RiskScore = riskScore
	e.Anomaly = anomaly
	return e
}

// WithChanges 设置字段变更信息
func (e *AuditEvent) WithChanges(changes []FieldChange, oldValues, newValues map[string]interface{}) *AuditEvent {
	e.Changes = changes
	e.OldValues = oldValues
	e.NewValues = newValues
	return e
}

// WithMetadata 设置元数据
func (e *AuditEvent) WithMetadata(key string, value interface{}) *AuditEvent {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithTags 设置标签
func (e *AuditEvent) WithTags(tags ...string) *AuditEvent {
	e.Tags = append(e.Tags, tags...)
	return e
}

// WithCompliance 设置合规性标志
func (e *AuditEvent) WithCompliance(flags ...string) *AuditEvent {
	e.ComplianceFlags = append(e.ComplianceFlags, flags...)
	return e
}

// SetResult 设置事件结果
func (e *AuditEvent) SetResult(result string) *AuditEvent {
	e.EventResult = result
	return e
}

// SetMessage 设置事件消息
func (e *AuditEvent) SetMessage(message string) *AuditEvent {
	e.EventMessage = message
	return e
}

// generateEventID 生成事件ID
func generateEventID() string {
	// UUIDID
	//
	return time.Now().Format("20060102150405") + "-" + "random"
}

// GenerateEventID 导出的事件ID生成函数
func GenerateEventID() string {
	return generateEventID()
}
