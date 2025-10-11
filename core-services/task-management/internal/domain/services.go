package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ========== 智能分配服务 ==========

// TaskAllocationService 任务智能分配服务
type TaskAllocationService interface {
	// 智能分配任务给最合适的成员
	AllocateTask(ctx context.Context, task *Task, team *Team) (*TaskAssignment, error)
	
	// 批量分配任务
	AllocateTasks(ctx context.Context, tasks []*Task, team *Team) ([]*TaskAssignment, error)
	
	// 重新分配任务
	ReallocateTask(ctx context.Context, taskID uuid.UUID, reason string) (*TaskAssignment, error)
	
	// 获取推荐的分配方�?
	GetAllocationRecommendations(ctx context.Context, task *Task, team *Team) ([]*AllocationRecommendation, error)
	
	// 评估分配质量
	EvaluateAllocation(ctx context.Context, assignment *TaskAssignment) (*AllocationEvaluation, error)
	
	// 优化团队工作负载
	OptimizeWorkload(ctx context.Context, teamID uuid.UUID) (*WorkloadOptimization, error)
}

// AllocationRecommendation 分配推荐
type AllocationRecommendation struct {
	UserID         uuid.UUID              `json:"user_id"`
	Score          float64                `json:"score"`
	Confidence     float64                `json:"confidence"`
	Reasons        []string               `json:"reasons"`
	Factors        map[string]float64     `json:"factors"`
	EstimatedTime  time.Duration          `json:"estimated_time"`
	RiskLevel      string                 `json:"risk_level"`
	Alternatives   []uuid.UUID            `json:"alternatives"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// AllocationEvaluation 分配评估
type AllocationEvaluation struct {
	Score           float64                `json:"score"`
	SkillMatch      float64                `json:"skill_match"`
	WorkloadBalance float64                `json:"workload_balance"`
	AvailabilityFit float64                `json:"availability_fit"`
	HistoryPerformance float64             `json:"history_performance"`
	RiskFactors     []string               `json:"risk_factors"`
	Suggestions     []string               `json:"suggestions"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// WorkloadOptimization 工作负载优化
type WorkloadOptimization struct {
	TeamID          uuid.UUID                    `json:"team_id"`
	CurrentLoad     map[uuid.UUID]float64        `json:"current_load"`
	OptimizedLoad   map[uuid.UUID]float64        `json:"optimized_load"`
	Reassignments   []*TaskReassignment          `json:"reassignments"`
	ImprovementScore float64                     `json:"improvement_score"`
	Recommendations []string                     `json:"recommendations"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// TaskReassignment 任务重新分配
type TaskReassignment struct {
	TaskID      uuid.UUID `json:"task_id"`
	FromUserID  uuid.UUID `json:"from_user_id"`
	ToUserID    uuid.UUID `json:"to_user_id"`
	Reason      string    `json:"reason"`
	Priority    int       `json:"priority"`
	EstimatedImpact float64 `json:"estimated_impact"`
}

// ========== 智能调度服务 ==========

// TaskSchedulingService 任务智能调度服务
type TaskSchedulingService interface {
	// 生成项目调度计划
	GenerateSchedule(ctx context.Context, projectID uuid.UUID) (*ProjectSchedule, error)
	
	// 优化现有调度
	OptimizeSchedule(ctx context.Context, scheduleID uuid.UUID) (*ProjectSchedule, error)
	
	// 检测调度冲�?
	DetectConflicts(ctx context.Context, scheduleID uuid.UUID) ([]*ScheduleConflict, error)
	
	// 解决调度冲突
	ResolveConflicts(ctx context.Context, conflicts []*ScheduleConflict) (*ConflictResolution, error)
	
	// 预测项目完成时间
	PredictCompletion(ctx context.Context, projectID uuid.UUID) (*CompletionPrediction, error)
	
	// 关键路径分析
	AnalyzeCriticalPath(ctx context.Context, projectID uuid.UUID) (*CriticalPathAnalysis, error)
}

// ProjectSchedule 项目调度计划
type ProjectSchedule struct {
	ID              uuid.UUID                  `json:"id"`
	ProjectID       uuid.UUID                  `json:"project_id"`
	Version         int                        `json:"version"`
	TaskSchedules   []*TaskSchedule            `json:"task_schedules"`
	Dependencies    []*ScheduleDependency      `json:"dependencies"`
	Milestones      []*ScheduleMilestone       `json:"milestones"`
	ResourceAllocations []*ResourceAllocation  `json:"resource_allocations"`
	CriticalPath    []uuid.UUID                `json:"critical_path"`
	EstimatedDuration time.Duration            `json:"estimated_duration"`
	Confidence      float64                    `json:"confidence"`
	RiskFactors     []string                   `json:"risk_factors"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
	Metadata        map[string]interface{}     `json:"metadata"`
}

// TaskSchedule 任务调度
type TaskSchedule struct {
	TaskID          uuid.UUID     `json:"task_id"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	AssignedUserID  uuid.UUID     `json:"assigned_user_id"`
	Priority        int           `json:"priority"`
	BufferTime      time.Duration `json:"buffer_time"`
	Flexibility     float64       `json:"flexibility"`
	Dependencies    []uuid.UUID   `json:"dependencies"`
}

// ScheduleDependency 调度依赖
type ScheduleDependency struct {
	FromTaskID   uuid.UUID `json:"from_task_id"`
	ToTaskID     uuid.UUID `json:"to_task_id"`
	Type         string    `json:"type"` // finish-to-start, start-to-start, etc.
	Lag          time.Duration `json:"lag"`
	IsHard       bool      `json:"is_hard"`
}

// ScheduleMilestone 调度里程�?
type ScheduleMilestone struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	TargetDate  time.Time `json:"target_date"`
	Dependencies []uuid.UUID `json:"dependencies"`
	IsCritical  bool      `json:"is_critical"`
}

// ResourceAllocation 资源分配
type ResourceAllocation struct {
	UserID       uuid.UUID     `json:"user_id"`
	TaskID       uuid.UUID     `json:"task_id"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Allocation   float64       `json:"allocation"` // 0.0-1.0
	Skills       []string      `json:"skills"`
	Availability float64       `json:"availability"`
}

// ScheduleConflict 调度冲突
type ScheduleConflict struct {
	ID          uuid.UUID   `json:"id"`
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	TaskIDs     []uuid.UUID `json:"task_ids"`
	UserIDs     []uuid.UUID `json:"user_ids"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     time.Time   `json:"end_time"`
	Suggestions []string    `json:"suggestions"`
}

// ConflictResolution 冲突解决方案
type ConflictResolution struct {
	ConflictID   uuid.UUID                  `json:"conflict_id"`
	Strategy     string                     `json:"strategy"`
	Actions      []*ResolutionAction        `json:"actions"`
	Impact       *ResolutionImpact          `json:"impact"`
	Confidence   float64                    `json:"confidence"`
	Alternatives []*AlternativeResolution   `json:"alternatives"`
}

// ResolutionAction 解决行动
type ResolutionAction struct {
	Type        string                 `json:"type"`
	TaskID      *uuid.UUID             `json:"task_id,omitempty"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	NewStartTime *time.Time            `json:"new_start_time,omitempty"`
	NewEndTime   *time.Time            `json:"new_end_time,omitempty"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// ResolutionImpact 解决影响
type ResolutionImpact struct {
	DelayDays       int     `json:"delay_days"`
	CostIncrease    float64 `json:"cost_increase"`
	QualityImpact   float64 `json:"quality_impact"`
	ResourceImpact  float64 `json:"resource_impact"`
	RiskIncrease    float64 `json:"risk_increase"`
}

// AlternativeResolution 替代解决方案
type AlternativeResolution struct {
	Strategy    string            `json:"strategy"`
	Actions     []*ResolutionAction `json:"actions"`
	Impact      *ResolutionImpact `json:"impact"`
	Confidence  float64           `json:"confidence"`
	Pros        []string          `json:"pros"`
	Cons        []string          `json:"cons"`
}

// CompletionPrediction 完成预测
type CompletionPrediction struct {
	ProjectID         uuid.UUID     `json:"project_id"`
	EstimatedDate     time.Time     `json:"estimated_date"`
	Confidence        float64       `json:"confidence"`
	ConfidenceInterval struct {
		EarliestDate time.Time `json:"earliest_date"`
		LatestDate   time.Time `json:"latest_date"`
	} `json:"confidence_interval"`
	RiskFactors       []string      `json:"risk_factors"`
	Assumptions       []string      `json:"assumptions"`
	Scenarios         []*Scenario   `json:"scenarios"`
}

// Scenario 场景
type Scenario struct {
	Name          string    `json:"name"`
	Probability   float64   `json:"probability"`
	EstimatedDate time.Time `json:"estimated_date"`
	Description   string    `json:"description"`
	Factors       []string  `json:"factors"`
}

// CriticalPathAnalysis 关键路径分析
type CriticalPathAnalysis struct {
	ProjectID       uuid.UUID              `json:"project_id"`
	CriticalPath    []uuid.UUID            `json:"critical_path"`
	TotalDuration   time.Duration          `json:"total_duration"`
	FloatAnalysis   map[uuid.UUID]float64  `json:"float_analysis"`
	BottleneckTasks []uuid.UUID            `json:"bottleneck_tasks"`
	Recommendations []string               `json:"recommendations"`
	RiskAssessment  *PathRiskAssessment    `json:"risk_assessment"`
}

// PathRiskAssessment 路径风险评估
type PathRiskAssessment struct {
	OverallRisk     float64            `json:"overall_risk"`
	TaskRisks       map[uuid.UUID]float64 `json:"task_risks"`
	RiskFactors     []string           `json:"risk_factors"`
	MitigationPlans []string           `json:"mitigation_plans"`
}

// ========== 性能分析服务 ==========

// PerformanceAnalysisService 性能分析服务
type PerformanceAnalysisService interface {
	// 分析用户性能
	AnalyzeUserPerformance(ctx context.Context, userID uuid.UUID, period string) (*UserPerformanceAnalysis, error)
	
	// 分析团队性能
	AnalyzeTeamPerformance(ctx context.Context, teamID uuid.UUID, period string) (*TeamPerformanceAnalysis, error)
	
	// 分析项目性能
	AnalyzeProjectPerformance(ctx context.Context, projectID uuid.UUID) (*ProjectPerformanceAnalysis, error)
	
	// 生成性能报告
	GeneratePerformanceReport(ctx context.Context, scope string, targetID uuid.UUID, period string) (*PerformanceReport, error)
	
	// 预测性能趋势
	PredictPerformanceTrend(ctx context.Context, userID uuid.UUID, days int) (*PerformanceTrendPrediction, error)
	
	// 识别性能瓶颈
	IdentifyBottlenecks(ctx context.Context, projectID uuid.UUID) ([]*PerformanceBottleneck, error)
}

// UserPerformanceAnalysis 用户性能分析
type UserPerformanceAnalysis struct {
	UserID              uuid.UUID                  `json:"user_id"`
	Period              string                     `json:"period"`
	TasksCompleted      int                        `json:"tasks_completed"`
	AverageCompletionTime time.Duration            `json:"average_completion_time"`
	QualityScore        float64                    `json:"quality_score"`
	ProductivityScore   float64                    `json:"productivity_score"`
	EfficiencyTrend     []EfficiencyPoint          `json:"efficiency_trend"`
	SkillUtilization    map[string]float64         `json:"skill_utilization"`
	WorkloadBalance     float64                    `json:"workload_balance"`
	CollaborationScore  float64                    `json:"collaboration_score"`
	Strengths           []string                   `json:"strengths"`
	ImprovementAreas    []string                   `json:"improvement_areas"`
	Recommendations     []string                   `json:"recommendations"`
	Metadata            map[string]interface{}     `json:"metadata"`
}

// EfficiencyPoint 效率�?
type EfficiencyPoint struct {
	Date       time.Time `json:"date"`
	Efficiency float64   `json:"efficiency"`
	Quality    float64   `json:"quality"`
	Velocity   float64   `json:"velocity"`
}

// TeamPerformanceAnalysis 团队性能分析
type TeamPerformanceAnalysis struct {
	TeamID              uuid.UUID                     `json:"team_id"`
	Period              string                        `json:"period"`
	TeamProductivity    float64                       `json:"team_productivity"`
	CollaborationIndex  float64                       `json:"collaboration_index"`
	CommunicationScore  float64                       `json:"communication_score"`
	MemberPerformances  map[uuid.UUID]*UserPerformanceAnalysis `json:"member_performances"`
	TeamDynamics        *TeamDynamicsAnalysis         `json:"team_dynamics"`
	SkillGaps           []string                      `json:"skill_gaps"`
	SuccessFactors      []string                      `json:"success_factors"`
	RiskFactors         []string                      `json:"risk_factors"`
	Recommendations     []string                      `json:"recommendations"`
	Metadata            map[string]interface{}        `json:"metadata"`
}

// TeamDynamicsAnalysis 团队动态分�?
type TeamDynamicsAnalysis struct {
	CohesionScore       float64            `json:"cohesion_score"`
	TrustLevel          float64            `json:"trust_level"`
	ConflictLevel       float64            `json:"conflict_level"`
	LeadershipEffectiveness float64        `json:"leadership_effectiveness"`
	DecisionMakingSpeed float64            `json:"decision_making_speed"`
	KnowledgeSharing    float64            `json:"knowledge_sharing"`
	Adaptability        float64            `json:"adaptability"`
	Factors             map[string]float64 `json:"factors"`
}

// ProjectPerformanceAnalysis 项目性能分析
type ProjectPerformanceAnalysis struct {
	ProjectID           uuid.UUID                  `json:"project_id"`
	SchedulePerformance float64                    `json:"schedule_performance"`
	BudgetPerformance   float64                    `json:"budget_performance"`
	QualityPerformance  float64                    `json:"quality_performance"`
	ScopePerformance    float64                    `json:"scope_performance"`
	RiskLevel           float64                    `json:"risk_level"`
	TeamEffectiveness   float64                    `json:"team_effectiveness"`
	StakeholderSatisfaction float64                `json:"stakeholder_satisfaction"`
	DeliveryPrediction  *DeliveryPrediction        `json:"delivery_prediction"`
	CriticalIssues      []string                   `json:"critical_issues"`
	SuccessIndicators   []string                   `json:"success_indicators"`
	Recommendations     []string                   `json:"recommendations"`
	Metadata            map[string]interface{}     `json:"metadata"`
}

// DeliveryPrediction 交付预测
type DeliveryPrediction struct {
	OnTimeDeliveryProbability float64   `json:"on_time_delivery_probability"`
	EstimatedDeliveryDate     time.Time `json:"estimated_delivery_date"`
	ConfidenceLevel           float64   `json:"confidence_level"`
	RiskFactors               []string  `json:"risk_factors"`
	MitigationStrategies      []string  `json:"mitigation_strategies"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	ID              uuid.UUID                  `json:"id"`
	Type            string                     `json:"type"`
	Scope           string                     `json:"scope"`
	TargetID        uuid.UUID                  `json:"target_id"`
	Period          string                     `json:"period"`
	GeneratedAt     time.Time                  `json:"generated_at"`
	Summary         *ReportSummary             `json:"summary"`
	Sections        []*ReportSection           `json:"sections"`
	Charts          []*ChartData               `json:"charts"`
	Recommendations []string                   `json:"recommendations"`
	Metadata        map[string]interface{}     `json:"metadata"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	OverallScore    float64            `json:"overall_score"`
	KeyMetrics      map[string]float64 `json:"key_metrics"`
	Highlights      []string           `json:"highlights"`
	ConcernAreas    []string           `json:"concern_areas"`
	TrendDirection  string             `json:"trend_direction"`
}

// ReportSection 报告章节
type ReportSection struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Data        map[string]interface{} `json:"data"`
	Charts      []*ChartData           `json:"charts"`
	Insights    []string               `json:"insights"`
}

// ChartData 图表数据
type ChartData struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Data        map[string]interface{} `json:"data"`
	Options     map[string]interface{} `json:"options"`
}

// PerformanceTrendPrediction 性能趋势预测
type PerformanceTrendPrediction struct {
	UserID          uuid.UUID              `json:"user_id"`
	PredictionDays  int                    `json:"prediction_days"`
	TrendPoints     []TrendPoint           `json:"trend_points"`
	Confidence      float64                `json:"confidence"`
	InfluencingFactors []string            `json:"influencing_factors"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TrendPoint 趋势�?
type TrendPoint struct {
	Date            time.Time `json:"date"`
	PredictedScore  float64   `json:"predicted_score"`
	Confidence      float64   `json:"confidence"`
	InfluencingFactors []string `json:"influencing_factors"`
}

// PerformanceBottleneck 性能瓶颈
type PerformanceBottleneck struct {
	ID              uuid.UUID              `json:"id"`
	Type            string                 `json:"type"`
	Severity        string                 `json:"severity"`
	Description     string                 `json:"description"`
	AffectedTasks   []uuid.UUID            `json:"affected_tasks"`
	AffectedUsers   []uuid.UUID            `json:"affected_users"`
	Impact          float64                `json:"impact"`
	RootCauses      []string               `json:"root_causes"`
	Solutions       []string               `json:"solutions"`
	Priority        int                    `json:"priority"`
	EstimatedEffort time.Duration          `json:"estimated_effort"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ========== 通知服务 ==========

// NotificationService 通知服务
type NotificationService interface {
	// 发送任务通知
	SendTaskNotification(ctx context.Context, notification *TaskNotification) error
	
	// 发送项目通知
	SendProjectNotification(ctx context.Context, notification *ProjectNotification) error
	
	// 发送团队通知
	SendTeamNotification(ctx context.Context, notification *TeamNotification) error
	
	// 批量发送通知
	SendBatchNotifications(ctx context.Context, notifications []Notification) error
	
	// 获取用户通知偏好
	GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
	
	// 更新用户通知偏好
	UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, preferences *NotificationPreferences) error
}

// Notification 通知接口
type Notification interface {
	GetID() uuid.UUID
	GetType() string
	GetRecipients() []uuid.UUID
	GetTitle() string
	GetContent() string
	GetPriority() string
	GetChannels() []string
	GetMetadata() map[string]interface{}
}

// TaskNotification 任务通知
type TaskNotification struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	TaskID      uuid.UUID              `json:"task_id"`
	Recipients  []uuid.UUID            `json:"recipients"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Priority    string                 `json:"priority"`
	Channels    []string               `json:"channels"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ProjectNotification 项目通知
type ProjectNotification struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	ProjectID   uuid.UUID              `json:"project_id"`
	Recipients  []uuid.UUID            `json:"recipients"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Priority    string                 `json:"priority"`
	Channels    []string               `json:"channels"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TeamNotification 团队通知
type TeamNotification struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	TeamID      uuid.UUID              `json:"team_id"`
	Recipients  []uuid.UUID            `json:"recipients"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Priority    string                 `json:"priority"`
	Channels    []string               `json:"channels"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NotificationPreferences 通知偏好
type NotificationPreferences struct {
	UserID              uuid.UUID                      `json:"user_id"`
	EmailEnabled        bool                           `json:"email_enabled"`
	SMSEnabled          bool                           `json:"sms_enabled"`
	PushEnabled         bool                           `json:"push_enabled"`
	InAppEnabled        bool                           `json:"in_app_enabled"`
	TaskNotifications   *TaskNotificationPreferences   `json:"task_notifications"`
	ProjectNotifications *ProjectNotificationPreferences `json:"project_notifications"`
	TeamNotifications   *TeamNotificationPreferences   `json:"team_notifications"`
	QuietHours          *QuietHoursSettings            `json:"quiet_hours"`
	Frequency           string                         `json:"frequency"` // immediate, daily, weekly
	UpdatedAt           time.Time                      `json:"updated_at"`
}

// TaskNotificationPreferences 任务通知偏好
type TaskNotificationPreferences struct {
	TaskAssigned    bool `json:"task_assigned"`
	TaskCompleted   bool `json:"task_completed"`
	TaskOverdue     bool `json:"task_overdue"`
	TaskCommented   bool `json:"task_commented"`
	TaskUpdated     bool `json:"task_updated"`
	DueDateReminder bool `json:"due_date_reminder"`
}

// ProjectNotificationPreferences 项目通知偏好
type ProjectNotificationPreferences struct {
	ProjectCreated    bool `json:"project_created"`
	ProjectCompleted  bool `json:"project_completed"`
	ProjectOverdue    bool `json:"project_overdue"`
	MilestoneReached  bool `json:"milestone_reached"`
	MemberAdded       bool `json:"member_added"`
	StatusChanged     bool `json:"status_changed"`
}

// TeamNotificationPreferences 团队通知偏好
type TeamNotificationPreferences struct {
	TeamCreated     bool `json:"team_created"`
	MemberJoined    bool `json:"member_joined"`
	MemberLeft      bool `json:"member_left"`
	RoleChanged     bool `json:"role_changed"`
	TeamUpdated     bool `json:"team_updated"`
	MeetingScheduled bool `json:"meeting_scheduled"`
}

// QuietHoursSettings 免打扰时间设�?
type QuietHoursSettings struct {
	Enabled   bool      `json:"enabled"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	TimeZone  string    `json:"time_zone"`
	Weekdays  []int     `json:"weekdays"` // 0=Sunday, 1=Monday, etc.
}

// ========== 领域服务工厂 ==========

// DomainServiceFactory 领域服务工厂
type DomainServiceFactory interface {
	TaskAllocationService() TaskAllocationService
	TaskSchedulingService() TaskSchedulingService
	PerformanceAnalysisService() PerformanceAnalysisService
	NotificationService() NotificationService
}

// 实现通知接口的方�?
func (tn *TaskNotification) GetID() uuid.UUID { return tn.ID }
func (tn *TaskNotification) GetType() string { return tn.Type }
func (tn *TaskNotification) GetRecipients() []uuid.UUID { return tn.Recipients }
func (tn *TaskNotification) GetTitle() string { return tn.Title }
func (tn *TaskNotification) GetContent() string { return tn.Content }
func (tn *TaskNotification) GetPriority() string { return tn.Priority }
func (tn *TaskNotification) GetChannels() []string { return tn.Channels }
func (tn *TaskNotification) GetMetadata() map[string]interface{} { return tn.Metadata }

func (pn *ProjectNotification) GetID() uuid.UUID { return pn.ID }
func (pn *ProjectNotification) GetType() string { return pn.Type }
func (pn *ProjectNotification) GetRecipients() []uuid.UUID { return pn.Recipients }
func (pn *ProjectNotification) GetTitle() string { return pn.Title }
func (pn *ProjectNotification) GetContent() string { return pn.Content }
func (pn *ProjectNotification) GetPriority() string { return pn.Priority }
func (pn *ProjectNotification) GetChannels() []string { return pn.Channels }
func (pn *ProjectNotification) GetMetadata() map[string]interface{} { return pn.Metadata }

func (tn *TeamNotification) GetID() uuid.UUID { return tn.ID }
func (tn *TeamNotification) GetType() string { return tn.Type }
func (tn *TeamNotification) GetRecipients() []uuid.UUID { return tn.Recipients }
func (tn *TeamNotification) GetTitle() string { return tn.Title }
func (tn *TeamNotification) GetContent() string { return tn.Content }
func (tn *TeamNotification) GetPriority() string { return tn.Priority }
func (tn *TeamNotification) GetChannels() []string { return tn.Channels }
func (tn *TeamNotification) GetMetadata() map[string]interface{} { return tn.Metadata }
