package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent 领域事件接口
type DomainEvent interface {
	GetEventID() uuid.UUID
	GetAggregateID() uuid.UUID
	GetEventType() string
	GetOccurredAt() time.Time
	GetVersion() int
}

// BaseDomainEvent 基础领域事件
type BaseDomainEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	EventType   string    `json:"event_type"`
	OccurredAt  time.Time `json:"occurred_at"`
	Version     int       `json:"version"`
}

func (e BaseDomainEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e BaseDomainEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e BaseDomainEvent) GetEventType() string {
	return e.EventType
}

func (e BaseDomainEvent) GetOccurredAt() time.Time {
	return e.OccurredAt
}

func (e BaseDomainEvent) GetVersion() int {
	return e.Version
}

// ========== 任务相关事件 ==========

// TaskCreatedEvent 任务创建事件
type TaskCreatedEvent struct {
	BaseDomainEvent
	TaskID    uuid.UUID    `json:"task_id"`
	Title     string       `json:"title"`
	Type      TaskType     `json:"type"`
	Priority  TaskPriority `json:"priority"`
	ProjectID uuid.UUID    `json:"project_id"`
	CreatorID uuid.UUID    `json:"creator_id"`
}

// TaskAssignedEvent 任务分配事件
type TaskAssignedEvent struct {
	BaseDomainEvent
	TaskID     uuid.UUID `json:"task_id"`
	AssigneeID uuid.UUID `json:"assignee_id"`
	AssignerID uuid.UUID `json:"assigner_id"`
}

// TaskUnassignedEvent 任务取消分配事件
type TaskUnassignedEvent struct {
	BaseDomainEvent
	TaskID uuid.UUID `json:"task_id"`
}

// TaskStartedEvent 任务开始事?
type TaskStartedEvent struct {
	BaseDomainEvent
	TaskID uuid.UUID `json:"task_id"`
	UserID uuid.UUID `json:"user_id"`
}

// TaskCompletedEvent 任务完成事件
type TaskCompletedEvent struct {
	BaseDomainEvent
	TaskID      uuid.UUID `json:"task_id"`
	UserID      uuid.UUID `json:"user_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// TaskCancelledEvent 任务取消事件
type TaskCancelledEvent struct {
	BaseDomainEvent
	TaskID uuid.UUID `json:"task_id"`
	UserID uuid.UUID `json:"user_id"`
	Reason string    `json:"reason"`
}

// TaskProgressUpdatedEvent 任务进度更新事件
type TaskProgressUpdatedEvent struct {
	BaseDomainEvent
	TaskID      uuid.UUID `json:"task_id"`
	UserID      uuid.UUID `json:"user_id"`
	OldProgress float64   `json:"old_progress"`
	NewProgress float64   `json:"new_progress"`
}

// TaskPriorityUpdatedEvent 任务优先级更新事?
type TaskPriorityUpdatedEvent struct {
	BaseDomainEvent
	TaskID      uuid.UUID    `json:"task_id"`
	UserID      uuid.UUID    `json:"user_id"`
	OldPriority TaskPriority `json:"old_priority"`
	NewPriority TaskPriority `json:"new_priority"`
}

// TaskDueDateUpdatedEvent 任务截止日期更新事件
type TaskDueDateUpdatedEvent struct {
	BaseDomainEvent
	TaskID  uuid.UUID  `json:"task_id"`
	UserID  uuid.UUID  `json:"user_id"`
	DueDate *time.Time `json:"due_date"`
}

// TaskCommentAddedEvent 任务评论添加事件
type TaskCommentAddedEvent struct {
	BaseDomainEvent
	TaskID    uuid.UUID `json:"task_id"`
	CommentID uuid.UUID `json:"comment_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
}

// TaskOverdueEvent 任务逾期事件
type TaskOverdueEvent struct {
	BaseDomainEvent
	TaskID  uuid.UUID `json:"task_id"`
	DueDate time.Time `json:"due_date"`
}

// TaskDependencyAddedEvent 任务依赖添加事件
type TaskDependencyAddedEvent struct {
	BaseDomainEvent
	TaskID         uuid.UUID `json:"task_id"`
	DependsOnID    uuid.UUID `json:"depends_on_id"`
	DependencyType string    `json:"dependency_type"`
}

// TaskDependencyRemovedEvent 任务依赖移除事件
type TaskDependencyRemovedEvent struct {
	BaseDomainEvent
	TaskID         uuid.UUID `json:"task_id"`
	DependsOnID    uuid.UUID `json:"depends_on_id"`
	DependencyType string    `json:"dependency_type"`
}

// ========== 项目相关事件 ==========

// ProjectCreatedEvent 项目创建事件
type ProjectCreatedEvent struct {
	BaseDomainEvent
	ProjectID      uuid.UUID       `json:"project_id"`
	Name           string          `json:"name"`
	Type           ProjectType     `json:"type"`
	Priority       ProjectPriority `json:"priority"`
	OwnerID        uuid.UUID       `json:"owner_id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
}

// ProjectStartedEvent 项目启动事件
type ProjectStartedEvent struct {
	BaseDomainEvent
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
}

// ProjectCompletedEvent 项目完成事件
type ProjectCompletedEvent struct {
	BaseDomainEvent
	ProjectID   uuid.UUID `json:"project_id"`
	UserID      uuid.UUID `json:"user_id"`
	CompletedAt time.Time `json:"completed_at"`
}

// ProjectCancelledEvent 项目取消事件
type ProjectCancelledEvent struct {
	BaseDomainEvent
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Reason    string    `json:"reason"`
}

// ProjectProgressUpdatedEvent 项目进度更新事件
type ProjectProgressUpdatedEvent struct {
	BaseDomainEvent
	ProjectID   uuid.UUID `json:"project_id"`
	UserID      uuid.UUID `json:"user_id"`
	OldProgress float64   `json:"old_progress"`
	NewProgress float64   `json:"new_progress"`
}

// ProjectMemberAddedEvent 项目成员添加事件
type ProjectMemberAddedEvent struct {
	BaseDomainEvent
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	AddedBy   uuid.UUID `json:"added_by"`
}

// ProjectMemberRemovedEvent 项目成员移除事件
type ProjectMemberRemovedEvent struct {
	BaseDomainEvent
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	RemovedBy uuid.UUID `json:"removed_by"`
}

// ProjectMemberRoleUpdatedEvent 项目成员角色更新事件
type ProjectMemberRoleUpdatedEvent struct {
	BaseDomainEvent
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	OldRole   string    `json:"old_role"`
	NewRole   string    `json:"new_role"`
	UpdatedBy uuid.UUID `json:"updated_by"`
}

// ========== 团队相关事件 ==========

// TeamCreatedEvent 团队创建事件
type TeamCreatedEvent struct {
	BaseDomainEvent
	TeamID         uuid.UUID `json:"team_id"`
	Name           string    `json:"name"`
	LeaderID       uuid.UUID `json:"leader_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
}

// TeamMemberAddedEvent 团队成员添加事件
type TeamMemberAddedEvent struct {
	BaseDomainEvent
	TeamID  uuid.UUID `json:"team_id"`
	UserID  uuid.UUID `json:"user_id"`
	Role    string    `json:"role"`
	AddedBy uuid.UUID `json:"added_by"`
}

// TeamMemberRemovedEvent 团队成员移除事件
type TeamMemberRemovedEvent struct {
	BaseDomainEvent
	TeamID    uuid.UUID `json:"team_id"`
	UserID    uuid.UUID `json:"user_id"`
	RemovedBy uuid.UUID `json:"removed_by"`
}

// TeamDisbandedEvent 团队解散事件
type TeamDisbandedEvent struct {
	BaseDomainEvent
	TeamID      uuid.UUID `json:"team_id"`
	DisbandedBy uuid.UUID `json:"disbanded_by"`
	Reason      string    `json:"reason"`
}

// ========== 智能分配相关事件 ==========

// TaskAutoAssignedEvent 任务自动分配事件
type TaskAutoAssignedEvent struct {
	BaseDomainEvent
	TaskID       uuid.UUID `json:"task_id"`
	AssigneeID   uuid.UUID `json:"assignee_id"`
	Algorithm    string    `json:"algorithm"`    // 使用的分配算?
	Confidence   float64   `json:"confidence"`   // 分配置信?
	Factors      []string  `json:"factors"`      // 影响分配的因?
}

// WorkloadBalancedEvent 工作负载平衡事件
type WorkloadBalancedEvent struct {
	BaseDomainEvent
	TeamID      uuid.UUID            `json:"team_id"`
	Adjustments []WorkloadAdjustment `json:"adjustments"`
}

// WorkloadAdjustment 工作负载调整
type WorkloadAdjustment struct {
	UserID       uuid.UUID `json:"user_id"`
	OldWorkload  float64   `json:"old_workload"`
	NewWorkload  float64   `json:"new_workload"`
	TasksChanged []uuid.UUID `json:"tasks_changed"`
}

// SkillMatchFoundEvent 技能匹配发现事?
type SkillMatchFoundEvent struct {
	BaseDomainEvent
	TaskID         uuid.UUID `json:"task_id"`
	UserID         uuid.UUID `json:"user_id"`
	RequiredSkills []string  `json:"required_skills"`
	UserSkills     []string  `json:"user_skills"`
	MatchScore     float64   `json:"match_score"`
}

// PerformanceAnalyzedEvent 性能分析事件
type PerformanceAnalyzedEvent struct {
	BaseDomainEvent
	UserID           uuid.UUID `json:"user_id"`
	Period           string    `json:"period"`           // 分析周期
	TasksCompleted   int       `json:"tasks_completed"`
	AverageTime      float64   `json:"average_time"`     // 平均完成时间
	QualityScore     float64   `json:"quality_score"`    // 质量评分
	EfficiencyScore  float64   `json:"efficiency_score"` // 效率评分
}

// ========== 通知相关事件 ==========

// NotificationTriggeredEvent 通知触发事件
type NotificationTriggeredEvent struct {
	BaseDomainEvent
	NotificationType string                 `json:"notification_type"`
	Recipients       []uuid.UUID            `json:"recipients"`
	Subject          string                 `json:"subject"`
	Content          string                 `json:"content"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ReminderScheduledEvent 提醒计划事件
type ReminderScheduledEvent struct {
	BaseDomainEvent
	ReminderID   uuid.UUID `json:"reminder_id"`
	TaskID       uuid.UUID `json:"task_id"`
	UserID       uuid.UUID `json:"user_id"`
	ReminderTime time.Time `json:"reminder_time"`
	Message      string    `json:"message"`
}

// ========== 报告相关事件 ==========

// ReportGeneratedEvent 报告生成事件
type ReportGeneratedEvent struct {
	BaseDomainEvent
	ReportID     uuid.UUID `json:"report_id"`
	ReportType   string    `json:"report_type"`
	GeneratedBy  uuid.UUID `json:"generated_by"`
	GeneratedFor uuid.UUID `json:"generated_for"` // 项目ID或团队ID
	Period       string    `json:"period"`
	Format       string    `json:"format"`
}

// MetricsCalculatedEvent 指标计算事件
type MetricsCalculatedEvent struct {
	BaseDomainEvent
	EntityID   uuid.UUID              `json:"entity_id"`   // 项目ID或团队ID
	EntityType string                 `json:"entity_type"` // project, team, user
	Metrics    map[string]interface{} `json:"metrics"`
	Period     string                 `json:"period"`
}

// ========== 集成相关事件 ==========

// ExternalSystemSyncEvent 外部系统同步事件
type ExternalSystemSyncEvent struct {
	BaseDomainEvent
	SystemName  string                 `json:"system_name"`
	SyncType    string                 `json:"sync_type"` // import, export, update
	EntityType  string                 `json:"entity_type"`
	EntityID    uuid.UUID              `json:"entity_id"`
	Status      string                 `json:"status"` // success, failed, partial
	Details     map[string]interface{} `json:"details"`
}

// WebhookTriggeredEvent Webhook触发事件
type WebhookTriggeredEvent struct {
	BaseDomainEvent
	WebhookURL  string                 `json:"webhook_url"`
	EventType   string                 `json:"event_type"`
	Payload     map[string]interface{} `json:"payload"`
	Status      string                 `json:"status"`
	RetryCount  int                    `json:"retry_count"`
}

// ========== 审计相关事件 ==========

// AuditLogCreatedEvent 审计日志创建事件
type AuditLogCreatedEvent struct {
	BaseDomainEvent
	UserID     uuid.UUID              `json:"user_id"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID uuid.UUID              `json:"resource_id"`
	Changes    map[string]interface{} `json:"changes"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
}

// SecurityEventDetectedEvent 安全事件检测事?
type SecurityEventDetectedEvent struct {
	BaseDomainEvent
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ========== 系统相关事件 ==========

// SystemHealthCheckEvent 系统健康检查事?
type SystemHealthCheckEvent struct {
	BaseDomainEvent
	Component string                 `json:"component"`
	Status    string                 `json:"status"` // healthy, warning, critical
	Metrics   map[string]interface{} `json:"metrics"`
	Message   string                 `json:"message"`
}

// BackupCompletedEvent 备份完成事件
type BackupCompletedEvent struct {
	BaseDomainEvent
	BackupType string    `json:"backup_type"` // full, incremental
	Size       int64     `json:"size"`
	Location   string    `json:"location"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"` // success, failed
}

// MaintenanceScheduledEvent 维护计划事件
type MaintenanceScheduledEvent struct {
	BaseDomainEvent
	MaintenanceType string    `json:"maintenance_type"`
	ScheduledTime   time.Time `json:"scheduled_time"`
	EstimatedDuration int     `json:"estimated_duration"` // 分钟
	Description     string    `json:"description"`
	AffectedServices []string `json:"affected_services"`
}

