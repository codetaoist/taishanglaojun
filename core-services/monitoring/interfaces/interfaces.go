package interfaces

import (
	"context"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// MetricCollector 指标收集器接�?
type MetricCollector interface {
	// Collect 收集指标
	Collect(ctx context.Context) ([]models.Metric, error)
	
	// GetName 获取收集器名�?
	GetName() string
	
	// GetCategory 获取收集器分�?
	GetCategory() models.MetricCategory
	
	// GetInterval 获取收集间隔
	GetInterval() time.Duration
	
	// IsEnabled 检查是否启�?
	IsEnabled() bool
	
	// Start 启动收集�?
	Start(ctx context.Context) error
	
	// Stop 停止收集�?
	Stop() error
	
	// Health 健康检�?
	Health() error
}

// MetricStorage 指标存储接口
type MetricStorage interface {
	// Store 存储指标
	Store(ctx context.Context, metrics []models.Metric) error
	
	// Query 查询指标
	Query(ctx context.Context, query models.MetricQuery) (*models.MetricQueryResult, error)
	
	// QueryRange 范围查询
	QueryRange(ctx context.Context, query models.MetricQuery) ([]*models.MetricQueryResult, error)
	
	// GetMetricNames 获取指标名称列表
	GetMetricNames(ctx context.Context) ([]string, error)
	
	// GetLabelValues 获取标签值列�?
	GetLabelValues(ctx context.Context, labelName string) ([]string, error)
	
	// GetSeries 获取时间序列
	GetSeries(ctx context.Context, matchers []string) ([]models.MetricSeries, error)
	
	// Delete 删除指标
	Delete(ctx context.Context, matchers []string) error
	
	// Cleanup 清理过期数据
	Cleanup(ctx context.Context, retention time.Duration) error
	
	// Health 健康检�?
	Health() error
	
	// Close 关闭存储
	Close() error
}

// MetricProcessor 指标处理器接�?
type MetricProcessor interface {
	// Process 处理指标
	Process(ctx context.Context, metrics []models.Metric) ([]models.Metric, error)
	
	// Aggregate 聚合指标
	Aggregate(ctx context.Context, metrics []models.Metric, aggregation string) (*models.AggregatedMetric, error)
	
	// Transform 转换指标
	Transform(ctx context.Context, metric models.Metric, rules []TransformRule) (*models.Metric, error)
	
	// Filter 过滤指标
	Filter(ctx context.Context, metrics []models.Metric, filters []FilterRule) ([]models.Metric, error)
	
	// Enrich 丰富指标
	Enrich(ctx context.Context, metric models.Metric) (*models.Metric, error)
}

// TransformRule 转换规则
type TransformRule struct {
	Type   string                 `json:"type"`   // rename, scale, unit_convert
	Config map[string]interface{} `json:"config"`
}

// FilterRule 过滤规则
type FilterRule struct {
	Type      string            `json:"type"`      // include, exclude
	MetricName string           `json:"metric_name"`
	Labels    map[string]string `json:"labels"`
	Condition string            `json:"condition"` // and, or
}

// AlertManager 告警管理器接�?
type AlertManager interface {
	// EvaluateRules 评估告警规则
	EvaluateRules(ctx context.Context) error
	
	// CreateRule 创建告警规则
	CreateRule(ctx context.Context, rule *models.AlertRule) error
	
	// UpdateRule 更新告警规则
	UpdateRule(ctx context.Context, rule *models.AlertRule) error
	
	// DeleteRule 删除告警规则
	DeleteRule(ctx context.Context, ruleID string) error
	
	// GetRule 获取告警规则
	GetRule(ctx context.Context, ruleID string) (*models.AlertRule, error)
	
	// ListRules 列出告警规则
	ListRules(ctx context.Context, filters map[string]interface{}) ([]*models.AlertRule, error)
	
	// FireAlert 触发告警
	FireAlert(ctx context.Context, alert *models.Alert) error
	
	// ResolveAlert 解决告警
	ResolveAlert(ctx context.Context, alertID string) error
	
	// AcknowledgeAlert 确认告警
	AcknowledgeAlert(ctx context.Context, alertID, acknowledgedBy string) error
	
	// GetAlert 获取告警
	GetAlert(ctx context.Context, alertID string) (*models.Alert, error)
	
	// ListAlerts 列出告警
	ListAlerts(ctx context.Context, filters map[string]interface{}) ([]*models.Alert, error)
	
	// CreateSilence 创建静默
	CreateSilence(ctx context.Context, silence *models.Silence) error
	
	// DeleteSilence 删除静默
	DeleteSilence(ctx context.Context, silenceID string) error
	
	// ListSilences 列出静默
	ListSilences(ctx context.Context) ([]*models.Silence, error)
}

// NotificationManager 通知管理器接�?
type NotificationManager interface {
	// SendNotification 发送通知
	SendNotification(ctx context.Context, notification *models.Notification) error
	
	// CreateChannel 创建通知渠道
	CreateChannel(ctx context.Context, channel *models.NotificationChannel) error
	
	// UpdateChannel 更新通知渠道
	UpdateChannel(ctx context.Context, channel *models.NotificationChannel) error
	
	// DeleteChannel 删除通知渠道
	DeleteChannel(ctx context.Context, channelID string) error
	
	// GetChannel 获取通知渠道
	GetChannel(ctx context.Context, channelID string) (*models.NotificationChannel, error)
	
	// ListChannels 列出通知渠道
	ListChannels(ctx context.Context) ([]*models.NotificationChannel, error)
	
	// TestChannel 测试通知渠道
	TestChannel(ctx context.Context, channelID string) error
	
	// GetNotificationHistory 获取通知历史
	GetNotificationHistory(ctx context.Context, filters map[string]interface{}) ([]*models.Notification, error)
	
	// RetryFailedNotifications 重试失败的通知
	RetryFailedNotifications(ctx context.Context) error
}

// DashboardManager 仪表板管理器接口
type DashboardManager interface {
	// CreateDashboard 创建仪表�?
	CreateDashboard(ctx context.Context, dashboard *Dashboard) error
	
	// UpdateDashboard 更新仪表�?
	UpdateDashboard(ctx context.Context, dashboard *Dashboard) error
	
	// DeleteDashboard 删除仪表�?
	DeleteDashboard(ctx context.Context, dashboardID string) error
	
	// GetDashboard 获取仪表�?
	GetDashboard(ctx context.Context, dashboardID string) (*Dashboard, error)
	
	// ListDashboards 列出仪表�?
	ListDashboards(ctx context.Context, filters map[string]interface{}) ([]*Dashboard, error)
	
	// RenderDashboard 渲染仪表�?
	RenderDashboard(ctx context.Context, dashboardID string, timeRange TimeRange) (*DashboardData, error)
	
	// ExportDashboard 导出仪表�?
	ExportDashboard(ctx context.Context, dashboardID string, format string) ([]byte, error)
	
	// ImportDashboard 导入仪表�?
	ImportDashboard(ctx context.Context, data []byte, format string) (*Dashboard, error)
}

// Dashboard 仪表�?
type Dashboard struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	Panels      []Panel     `json:"panels"`
	Variables   []Variable  `json:"variables"`
	TimeRange   TimeRange   `json:"time_range"`
	Refresh     string      `json:"refresh"`
	CreatedBy   string      `json:"created_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Panel 面板
type Panel struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"` // graph, stat, table, heatmap
	GridPos     GridPos                `json:"grid_pos"`
	Targets     []Target               `json:"targets"`
	Options     map[string]interface{} `json:"options"`
	FieldConfig FieldConfig            `json:"field_config"`
}

// GridPos 网格位置
type GridPos struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// Target 查询目标
type Target struct {
	RefID      string `json:"ref_id"`
	Expression string `json:"expression"`
	Legend     string `json:"legend"`
	Interval   string `json:"interval"`
	Format     string `json:"format"`
}

// FieldConfig 字段配置
type FieldConfig struct {
	Defaults  FieldDefaults `json:"defaults"`
	Overrides []Override    `json:"overrides"`
}

// FieldDefaults 字段默认�?
type FieldDefaults struct {
	Unit        string      `json:"unit"`
	Min         *float64    `json:"min,omitempty"`
	Max         *float64    `json:"max,omitempty"`
	Decimals    *int        `json:"decimals,omitempty"`
	DisplayName string      `json:"display_name"`
	Color       ColorConfig `json:"color"`
}

// ColorConfig 颜色配置
type ColorConfig struct {
	Mode   string `json:"mode"`
	Scheme string `json:"scheme"`
}

// Override 覆盖配置
type Override struct {
	Matcher    Matcher                `json:"matcher"`
	Properties map[string]interface{} `json:"properties"`
}

// Matcher 匹配�?
type Matcher struct {
	ID      string `json:"id"`
	Options string `json:"options"`
}

// Variable 变量
type Variable struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Label       string   `json:"label"`
	Query       string   `json:"query"`
	Options     []Option `json:"options"`
	Current     Option   `json:"current"`
	Multi       bool     `json:"multi"`
	IncludeAll  bool     `json:"include_all"`
	AllValue    string   `json:"all_value"`
	Refresh     string   `json:"refresh"`
}

// Option 选项
type Option struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected"`
}

// TimeRange 时间范围
type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// DashboardData 仪表板数�?
type DashboardData struct {
	Dashboard *Dashboard  `json:"dashboard"`
	Panels    []PanelData `json:"panels"`
}

// PanelData 面板数据
type PanelData struct {
	ID     string      `json:"id"`
	Series []DataSeries `json:"series"`
}

// DataSeries 数据序列
type DataSeries struct {
	Name   string      `json:"name"`
	Fields []DataField `json:"fields"`
}

// DataField 数据字段
type DataField struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"`
	Values []interface{} `json:"values"`
}

// TracingManager 分布式追踪管理器接口
type TracingManager interface {
	// CreateSpan 创建跨度
	CreateSpan(ctx context.Context, span *Span) error
	
	// GetTrace 获取追踪
	GetTrace(ctx context.Context, traceID string) (*Trace, error)
	
	// SearchTraces 搜索追踪
	SearchTraces(ctx context.Context, query TraceQuery) ([]*Trace, error)
	
	// GetServices 获取服务列表
	GetServices(ctx context.Context) ([]string, error)
	
	// GetOperations 获取操作列表
	GetOperations(ctx context.Context, service string) ([]string, error)
	
	// GetDependencies 获取依赖关系
	GetDependencies(ctx context.Context, endTs time.Time, lookback time.Duration) ([]*Dependency, error)
}

// Span 跨度
type Span struct {
	TraceID       string            `json:"trace_id"`
	SpanID        string            `json:"span_id"`
	ParentSpanID  string            `json:"parent_span_id"`
	OperationName string            `json:"operation_name"`
	Service       string            `json:"service"`
	StartTime     time.Time         `json:"start_time"`
	Duration      time.Duration     `json:"duration"`
	Tags          map[string]string `json:"tags"`
	Logs          []Log             `json:"logs"`
	References    []Reference       `json:"references"`
}

// Log 日志
type Log struct {
	Timestamp time.Time         `json:"timestamp"`
	Fields    map[string]string `json:"fields"`
}

// Reference 引用
type Reference struct {
	Type    string `json:"type"`
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

// Trace 追踪
type Trace struct {
	TraceID   string  `json:"trace_id"`
	Spans     []Span  `json:"spans"`
	Processes []Process `json:"processes"`
}

// Process 进程
type Process struct {
	ServiceName string            `json:"service_name"`
	Tags        map[string]string `json:"tags"`
}

// TraceQuery 追踪查询
type TraceQuery struct {
	Service     string        `json:"service"`
	Operation   string        `json:"operation"`
	Tags        map[string]string `json:"tags"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	MinDuration time.Duration `json:"min_duration"`
	MaxDuration time.Duration `json:"max_duration"`
	Limit       int           `json:"limit"`
}

// Dependency 依赖关系
type Dependency struct {
	Parent    string `json:"parent"`
	Child     string `json:"child"`
	CallCount int64  `json:"call_count"`
}

// LogManager 日志管理器接�?
type LogManager interface {
	// IngestLogs 摄取日志
	IngestLogs(ctx context.Context, logs []LogEntry) error
	
	// SearchLogs 搜索日志
	SearchLogs(ctx context.Context, query LogQuery) (*LogSearchResult, error)
	
	// GetLogStream 获取日志�?
	GetLogStream(ctx context.Context, query LogQuery) (<-chan LogEntry, error)
	
	// CreateIndex 创建索引
	CreateIndex(ctx context.Context, index LogIndex) error
	
	// DeleteIndex 删除索引
	DeleteIndex(ctx context.Context, indexName string) error
	
	// GetIndices 获取索引列表
	GetIndices(ctx context.Context) ([]LogIndex, error)
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Service   string            `json:"service"`
	Host      string            `json:"host"`
	Labels    map[string]string `json:"labels"`
	Fields    map[string]interface{} `json:"fields"`
	TraceID   string            `json:"trace_id,omitempty"`
	SpanID    string            `json:"span_id,omitempty"`
}

// LogQuery 日志查询
type LogQuery struct {
	Query     string            `json:"query"`
	Labels    map[string]string `json:"labels"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Limit     int               `json:"limit"`
	Direction string            `json:"direction"` // forward, backward
}

// LogSearchResult 日志搜索结果
type LogSearchResult struct {
	Entries []LogEntry `json:"entries"`
	Total   int64      `json:"total"`
	Took    int64      `json:"took"`
}

// LogIndex 日志索引
type LogIndex struct {
	Name     string            `json:"name"`
	Pattern  string            `json:"pattern"`
	Settings map[string]interface{} `json:"settings"`
	Mappings map[string]interface{} `json:"mappings"`
}

// AutomationManager 自动化管理器接口
type AutomationManager interface {
	// CreateRule 创建自动化规�?
	CreateRule(ctx context.Context, rule *AutomationRule) error
	
	// UpdateRule 更新自动化规�?
	UpdateRule(ctx context.Context, rule *AutomationRule) error
	
	// DeleteRule 删除自动化规�?
	DeleteRule(ctx context.Context, ruleID string) error
	
	// GetRule 获取自动化规�?
	GetRule(ctx context.Context, ruleID string) (*AutomationRule, error)
	
	// ListRules 列出自动化规�?
	ListRules(ctx context.Context) ([]*AutomationRule, error)
	
	// ExecuteRule 执行自动化规�?
	ExecuteRule(ctx context.Context, ruleID string, context map[string]interface{}) error
	
	// GetExecutionHistory 获取执行历史
	GetExecutionHistory(ctx context.Context, ruleID string) ([]*AutomationExecution, error)
}

// AutomationRule 自动化规�?
type AutomationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Trigger     AutomationTrigger      `json:"trigger"`
	Conditions  []AutomationCondition  `json:"conditions"`
	Actions     []AutomationAction     `json:"actions"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AutomationTrigger 自动化触发器
type AutomationTrigger struct {
	Type   string                 `json:"type"` // alert, metric, schedule, webhook
	Config map[string]interface{} `json:"config"`
}

// AutomationCondition 自动化条�?
type AutomationCondition struct {
	Type     string                 `json:"type"` // metric, time, service_status
	Operator string                 `json:"operator"`
	Value    interface{}            `json:"value"`
	Config   map[string]interface{} `json:"config"`
}

// AutomationAction 自动化动�?
type AutomationAction struct {
	Type   string                 `json:"type"` // scale, restart, notify, webhook
	Config map[string]interface{} `json:"config"`
	Retry  RetryConfig            `json:"retry"`
}

// AutomationExecution 自动化执行记�?
type AutomationExecution struct {
	ID        string                 `json:"id"`
	RuleID    string                 `json:"rule_id"`
	Status    string                 `json:"status"` // success, failed, running
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Context   map[string]interface{} `json:"context"`
	Results   []ActionResult         `json:"results"`
}

// ActionResult 动作执行结果
type ActionResult struct {
	ActionType string                 `json:"action_type"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	Duration   time.Duration          `json:"duration"`
}

// HealthChecker 健康检查接�?
type HealthChecker interface {
	// Check 执行健康检�?
	Check(ctx context.Context) HealthStatus
	
	// GetName 获取检查器名称
	GetName() string
	
	// GetDependencies 获取依赖�?
	GetDependencies() []string
}

// HealthStatus 健康状�?
type HealthStatus struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"` // healthy, unhealthy, degraded
	Message      string            `json:"message"`
	Details      map[string]interface{} `json:"details"`
	Timestamp    time.Time         `json:"timestamp"`
	Duration     time.Duration     `json:"duration"`
	Dependencies []HealthStatus    `json:"dependencies"`
}

// ConfigManager 配置管理器接�?
type ConfigManager interface {
	// GetConfig 获取配置
	GetConfig(key string) (interface{}, error)
	
	// SetConfig 设置配置
	SetConfig(key string, value interface{}) error
	
	// DeleteConfig 删除配置
	DeleteConfig(key string) error
	
	// ListConfigs 列出配置
	ListConfigs(prefix string) (map[string]interface{}, error)
	
	// WatchConfig 监听配置变化
	WatchConfig(key string) (<-chan ConfigChange, error)
	
	// Reload 重新加载配置
	Reload() error
}

// ConfigChange 配置变化
type ConfigChange struct {
	Key       string      `json:"key"`
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Operation string      `json:"operation"` // create, update, delete
	Timestamp time.Time   `json:"timestamp"`
}
