package models

import (
	"time"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusPending   AlertStatus = "pending"   // 待处理
	AlertStatusFiring    AlertStatus = "firing"    // 触发中
	AlertStatusResolved  AlertStatus = "resolved"  // 已解决
	AlertStatusSuppressed AlertStatus = "suppressed" // 已抑制
	AlertStatusAcknowledged AlertStatus = "acknowledged" // 已确认
)

// AlertSeverity 告警严重程度
type AlertSeverity string

const (
	SeverityInfo      AlertSeverity = "info"      // 信息
	SeverityWarning   AlertSeverity = "warning"   // 警告
	SeverityCritical  AlertSeverity = "critical"  // 严重
	SeverityEmergency AlertSeverity = "emergency" // 紧急
)

// AlertRule 告警规则
type AlertRule struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"not null;uniqueIndex"`
	Description string            `json:"description"`
	Expression  string            `json:"expression" gorm:"not null"` // PromQL表达式
	Labels      map[string]string `json:"labels" gorm:"type:jsonb"`
	Annotations map[string]string `json:"annotations" gorm:"type:jsonb"`
	Severity    AlertSeverity     `json:"severity" gorm:"not null;index"`
	Duration    time.Duration     `json:"duration"` // 持续时间阈值
	Interval    time.Duration     `json:"interval"` // 评估间隔
	Enabled     bool              `json:"enabled" gorm:"default:true;index"`
	GroupBy     []string          `json:"group_by" gorm:"type:jsonb"`
	Conditions  []AlertCondition  `json:"conditions" gorm:"type:jsonb"`
	Actions     []AlertAction     `json:"actions" gorm:"type:jsonb"`
	Runbook     string            `json:"runbook"`
	Dashboard   string            `json:"dashboard"`
	CreatedBy   string            `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AlertCondition 告警条件
type AlertCondition struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Operator   string            `json:"operator"` // >, <, >=, <=, ==, !=
	Threshold  float64           `json:"threshold"`
	Duration   time.Duration     `json:"duration"`
	Function   string            `json:"function"` // avg, sum, min, max, count
}

// AlertAction 告警动作
type AlertAction struct {
	Type       string                 `json:"type"` // notification, webhook, script, escalation
	Config     map[string]interface{} `json:"config"`
	Conditions map[string]string      `json:"conditions"` // 执行条件
	Delay      time.Duration          `json:"delay"`
	Retry      int                    `json:"retry"`
	Timeout    time.Duration          `json:"timeout"`
}

// Alert 告警实例
type Alert struct {
	ID           string            `json:"id" gorm:"primaryKey"`
	RuleID       string            `json:"rule_id" gorm:"not null;index"`
	RuleName     string            `json:"rule_name" gorm:"not null;index"`
	Status       AlertStatus       `json:"status" gorm:"not null;index"`
	Severity     AlertSeverity     `json:"severity" gorm:"not null;index"`
	Labels       map[string]string `json:"labels" gorm:"type:jsonb"`
	Annotations  map[string]string `json:"annotations" gorm:"type:jsonb"`
	Value        float64           `json:"value"`
	Threshold    float64           `json:"threshold"`
	StartsAt     time.Time         `json:"starts_at" gorm:"not null;index"`
	EndsAt       *time.Time        `json:"ends_at,omitempty" gorm:"index"`
	UpdatedAt    time.Time         `json:"updated_at"`
	ResolvedAt   *time.Time        `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time      `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string          `json:"acknowledged_by"`
	Fingerprint  string            `json:"fingerprint" gorm:"uniqueIndex"`
	GeneratorURL string            `json:"generator_url"`
	SilenceID    string            `json:"silence_id,omitempty"`
	InhibitedBy  []string          `json:"inhibited_by" gorm:"type:jsonb"`
	Count        int64             `json:"count" gorm:"default:1"`
	LastSeen     time.Time         `json:"last_seen"`
}

// AlertGroup 告警分组
type AlertGroup struct {
	ID        string            `json:"id" gorm:"primaryKey"`
	GroupKey  string            `json:"group_key" gorm:"not null;uniqueIndex"`
	Labels    map[string]string `json:"labels" gorm:"type:jsonb"`
	Alerts    []Alert           `json:"alerts" gorm:"foreignKey:GroupID"`
	Status    AlertStatus       `json:"status" gorm:"not null;index"`
	Severity  AlertSeverity     `json:"severity" gorm:"not null;index"`
	Count     int               `json:"count"`
	StartsAt  time.Time         `json:"starts_at" gorm:"not null;index"`
	EndsAt    *time.Time        `json:"ends_at,omitempty"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// AlertHistory 告警历史
type AlertHistory struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	AlertID     string            `json:"alert_id" gorm:"not null;index"`
	RuleID      string            `json:"rule_id" gorm:"not null;index"`
	Status      AlertStatus       `json:"status" gorm:"not null;index"`
	Severity    AlertSeverity     `json:"severity" gorm:"not null;index"`
	Labels      map[string]string `json:"labels" gorm:"type:jsonb"`
	Annotations map[string]string `json:"annotations" gorm:"type:jsonb"`
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	Message     string            `json:"message"`
	Timestamp   time.Time         `json:"timestamp" gorm:"not null;index"`
	Duration    time.Duration     `json:"duration"`
	CreatedAt   time.Time         `json:"created_at"`
}

// Silence 告警静默
type Silence struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Matchers    []SilenceMatcher  `json:"matchers" gorm:"type:jsonb"`
	StartsAt    time.Time         `json:"starts_at" gorm:"not null;index"`
	EndsAt      time.Time         `json:"ends_at" gorm:"not null;index"`
	Comment     string            `json:"comment"`
	CreatedBy   string            `json:"created_by"`
	UpdatedBy   string            `json:"updated_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Status      string            `json:"status" gorm:"index"` // active, expired, pending
}

// SilenceMatcher 静默匹配器
type SilenceMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"is_regex"`
	IsEqual bool   `json:"is_equal"`
}

// Inhibition 告警抑制
type Inhibition struct {
	ID               string            `json:"id" gorm:"primaryKey"`
	SourceMatchers   []InhibitMatcher  `json:"source_matchers" gorm:"type:jsonb"`
	TargetMatchers   []InhibitMatcher  `json:"target_matchers" gorm:"type:jsonb"`
	Equal            []string          `json:"equal" gorm:"type:jsonb"`
	Description      string            `json:"description"`
	Enabled          bool              `json:"enabled" gorm:"default:true"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// InhibitMatcher 抑制匹配器
type InhibitMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"is_regex"`
	IsEqual bool   `json:"is_equal"`
}

// NotificationChannel 通知渠道
type NotificationChannel struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"not null;uniqueIndex"`
	Type        string                 `json:"type" gorm:"not null"` // email, sms, webhook, slack, dingtalk, wechat
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	Enabled     bool                   `json:"enabled" gorm:"default:true"`
	RateLimit   RateLimit              `json:"rate_limit" gorm:"embedded"`
	Retry       RetryConfig            `json:"retry" gorm:"embedded"`
	Template    string                 `json:"template"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RateLimit 速率限制
type RateLimit struct {
	Enabled   bool          `json:"enabled"`
	Requests  int           `json:"requests"`
	Duration  time.Duration `json:"duration"`
	BurstSize int           `json:"burst_size"`
}

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled     bool          `json:"enabled"`
	MaxRetries  int           `json:"max_retries"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay    time.Duration `json:"max_delay"`
	Multiplier  float64       `json:"multiplier"`
}

// Notification 通知记录
type Notification struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	AlertID     string            `json:"alert_id" gorm:"not null;index"`
	ChannelID   string            `json:"channel_id" gorm:"not null;index"`
	ChannelType string            `json:"channel_type" gorm:"not null;index"`
	Status      string            `json:"status" gorm:"not null;index"` // pending, sent, failed, retry
	Message     string            `json:"message"`
	Recipients  []string          `json:"recipients" gorm:"type:jsonb"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	SentAt      *time.Time        `json:"sent_at,omitempty"`
	FailedAt    *time.Time        `json:"failed_at,omitempty"`
	Error       string            `json:"error"`
	RetryCount  int               `json:"retry_count" gorm:"default:0"`
	NextRetry   *time.Time        `json:"next_retry,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// EscalationPolicy 升级策略
type EscalationPolicy struct {
	ID          string              `json:"id" gorm:"primaryKey"`
	Name        string              `json:"name" gorm:"not null;uniqueIndex"`
	Description string              `json:"description"`
	Rules       []EscalationRule    `json:"rules" gorm:"type:jsonb"`
	Enabled     bool                `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// EscalationRule 升级规则
type EscalationRule struct {
	Level       int           `json:"level"`
	Delay       time.Duration `json:"delay"`
	Channels    []string      `json:"channels"`
	Recipients  []string      `json:"recipients"`
	Conditions  map[string]string `json:"conditions"`
	StopOnAck   bool          `json:"stop_on_ack"`
	StopOnResolve bool        `json:"stop_on_resolve"`
}

// OnCallSchedule 值班计划
type OnCallSchedule struct {
	ID          string              `json:"id" gorm:"primaryKey"`
	Name        string              `json:"name" gorm:"not null;uniqueIndex"`
	Description string              `json:"description"`
	TimeZone    string              `json:"time_zone"`
	Rotations   []OnCallRotation    `json:"rotations" gorm:"type:jsonb"`
	Overrides   []OnCallOverride    `json:"overrides" gorm:"type:jsonb"`
	Enabled     bool                `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// OnCallRotation 值班轮换
type OnCallRotation struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	Users       []string      `json:"users"`
	Type        string        `json:"type"` // daily, weekly, monthly
	Handoff     time.Time     `json:"handoff"`
}

// OnCallOverride 值班覆盖
type OnCallOverride struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason"`
}

// AlertTemplate 告警模板
type AlertTemplate struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"not null;uniqueIndex"`
	Type        string            `json:"type" gorm:"not null"` // email, sms, webhook, etc.
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	Variables   []string          `json:"variables" gorm:"type:jsonb"`
	Format      string            `json:"format"` // text, html, markdown, json
	Language    string            `json:"language" gorm:"default:'en'"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// AlertMetrics 告警指标统计
type AlertMetrics struct {
	ID              string        `json:"id" gorm:"primaryKey"`
	Date            time.Time     `json:"date" gorm:"not null;index"`
	TotalAlerts     int64         `json:"total_alerts"`
	FiringAlerts    int64         `json:"firing_alerts"`
	ResolvedAlerts  int64         `json:"resolved_alerts"`
	CriticalAlerts  int64         `json:"critical_alerts"`
	WarningAlerts   int64         `json:"warning_alerts"`
	InfoAlerts      int64         `json:"info_alerts"`
	MTTR            time.Duration `json:"mttr"` // Mean Time To Resolution
	MTTD            time.Duration `json:"mttd"` // Mean Time To Detection
	FalsePositives  int64         `json:"false_positives"`
	MissedAlerts    int64         `json:"missed_alerts"`
	NotificationsSent int64       `json:"notifications_sent"`
	NotificationsFailed int64     `json:"notifications_failed"`
	CreatedAt       time.Time     `json:"created_at"`
}

// AlertConfiguration 告警配置
type AlertConfiguration struct {
	ID                    string        `json:"id" gorm:"primaryKey"`
	GlobalEnabled         bool          `json:"global_enabled" gorm:"default:true"`
	EvaluationInterval    time.Duration `json:"evaluation_interval"`
	GroupWait             time.Duration `json:"group_wait"`
	GroupInterval         time.Duration `json:"group_interval"`
	RepeatInterval        time.Duration `json:"repeat_interval"`
	ResolveTimeout        time.Duration `json:"resolve_timeout"`
	MaxConcurrentAlerts   int           `json:"max_concurrent_alerts"`
	MaxNotificationsPerHour int         `json:"max_notifications_per_hour"`
	DefaultSeverity       AlertSeverity `json:"default_severity"`
	AutoResolveEnabled    bool          `json:"auto_resolve_enabled"`
	AutoResolveTimeout    time.Duration `json:"auto_resolve_timeout"`
	DeduplicationEnabled  bool          `json:"deduplication_enabled"`
	DeduplicationWindow   time.Duration `json:"deduplication_window"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

// Helper functions

// NewAlertRule 创建新告警规则
func NewAlertRule(name, expression string, severity AlertSeverity) *AlertRule {
	return &AlertRule{
		ID:          generateID(),
		Name:        name,
		Expression:  expression,
		Severity:    severity,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Enabled:     true,
		Duration:    5 * time.Minute,
		Interval:    1 * time.Minute,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewAlert 创建新告警
func NewAlert(ruleID, ruleName string, severity AlertSeverity) *Alert {
	now := time.Now()
	return &Alert{
		ID:          generateID(),
		RuleID:      ruleID,
		RuleName:    ruleName,
		Status:      AlertStatusPending,
		Severity:    severity,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		StartsAt:    now,
		UpdatedAt:   now,
		LastSeen:    now,
		Count:       1,
		Fingerprint: generateFingerprint(ruleID, make(map[string]string)),
	}
}

// Fire 触发告警
func (a *Alert) Fire() {
	a.Status = AlertStatusFiring
	a.UpdatedAt = time.Now()
	a.LastSeen = time.Now()
	a.Count++
}

// Resolve 解决告警
func (a *Alert) Resolve() {
	now := time.Now()
	a.Status = AlertStatusResolved
	a.UpdatedAt = now
	a.ResolvedAt = &now
	a.EndsAt = &now
}

// Acknowledge 确认告警
func (a *Alert) Acknowledge(by string) {
	now := time.Now()
	a.Status = AlertStatusAcknowledged
	a.UpdatedAt = now
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = by
}

// Suppress 抑制告警
func (a *Alert) Suppress(silenceID string) {
	a.Status = AlertStatusSuppressed
	a.UpdatedAt = time.Now()
	a.SilenceID = silenceID
}

// IsActive 检查告警是否活跃
func (a *Alert) IsActive() bool {
	return a.Status == AlertStatusFiring || a.Status == AlertStatusPending
}

// IsResolved 检查告警是否已解决
func (a *Alert) IsResolved() bool {
	return a.Status == AlertStatusResolved
}

// Duration 获取告警持续时间
func (a *Alert) Duration() time.Duration {
	if a.EndsAt != nil {
		return a.EndsAt.Sub(a.StartsAt)
	}
	return time.Since(a.StartsAt)
}

// MatchesLabels 检查告警是否匹配标签
func (a *Alert) MatchesLabels(labels map[string]string) bool {
	if a.Labels == nil && len(labels) == 0 {
		return true
	}
	if a.Labels == nil || len(labels) == 0 {
		return false
	}
	
	for k, v := range labels {
		if labelValue, exists := a.Labels[k]; !exists || labelValue != v {
			return false
		}
	}
	return true
}

// generateFingerprint 生成告警指纹
func generateFingerprint(ruleID string, labels map[string]string) string {
	// 实现指纹生成逻辑
	return ruleID + "-" + hashLabels(labels)
}

// hashLabels 计算标签哈希
func hashLabels(labels map[string]string) string {
	// 实现标签哈希计算逻辑
	return "hash-" + randomString(8)
}

// NewSilence 创建新静默
func NewSilence(matchers []SilenceMatcher, startsAt, endsAt time.Time, comment, createdBy string) *Silence {
	return &Silence{
		ID:        generateID(),
		Matchers:  matchers,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Comment:   comment,
		CreatedBy: createdBy,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsActive 检查静默是否活跃
func (s *Silence) IsActive() bool {
	now := time.Now()
	return s.Status == "active" && now.After(s.StartsAt) && now.Before(s.EndsAt)
}

// IsExpired 检查静默是否过期
func (s *Silence) IsExpired() bool {
	return time.Now().After(s.EndsAt)
}

// Matches 检查静默是否匹配告警
func (s *Silence) Matches(alert *Alert) bool {
	for _, matcher := range s.Matchers {
		if !s.matchLabel(alert.Labels, matcher) {
			return false
		}
	}
	return true
}

// matchLabel 匹配标签
func (s *Silence) matchLabel(labels map[string]string, matcher SilenceMatcher) bool {
	value, exists := labels[matcher.Name]
	if !exists {
		return !matcher.IsEqual
	}
	
	if matcher.IsRegex {
		// 实现正则表达式匹配
		return true // 简化实现
	}
	
	if matcher.IsEqual {
		return value == matcher.Value
	}
	return value != matcher.Value
}