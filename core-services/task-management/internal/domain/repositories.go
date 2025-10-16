package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TaskRepository 任务仓储接口
type TaskRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id uuid.UUID) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 查询操作
	FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*Task, error)
	FindByAssigneeID(ctx context.Context, assigneeID uuid.UUID, limit, offset int) ([]*Task, error)
	FindByCreatorID(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*Task, error)
	FindByStatus(ctx context.Context, status TaskStatus, limit, offset int) ([]*Task, error)
	FindByPriority(ctx context.Context, priority TaskPriority, limit, offset int) ([]*Task, error)
	FindByType(ctx context.Context, taskType TaskType, limit, offset int) ([]*Task, error)

	// 复合查询
	FindByProjectAndStatus(ctx context.Context, projectID uuid.UUID, status TaskStatus, limit, offset int) ([]*Task, error)
	FindByAssigneeAndStatus(ctx context.Context, assigneeID uuid.UUID, status TaskStatus, limit, offset int) ([]*Task, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*Task, error)
	FindOverdueTasks(ctx context.Context, limit, offset int) ([]*Task, error)
	FindTasksWithDependencies(ctx context.Context, taskID uuid.UUID) ([]*Task, error)

	// 搜索操作
	SearchTasks(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*Task, error)
	FindByTags(ctx context.Context, tags []string, limit, offset int) ([]*Task, error)
	FindByLabels(ctx context.Context, labels map[string]string, limit, offset int) ([]*Task, error)

	// 统计操作
	Count(ctx context.Context) (int64, error)
	CountByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
	CountByAssignee(ctx context.Context, assigneeID uuid.UUID) (int64, error)
	CountByStatus(ctx context.Context, status TaskStatus) (int64, error)
	GetTaskStatistics(ctx context.Context, projectID *uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID) (*TaskStatistics, error)

	// 批量操作
	SaveBatch(ctx context.Context, tasks []*Task) error
	UpdateBatch(ctx context.Context, tasks []*Task) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// 依赖关系操作
	FindDependencies(ctx context.Context, taskID uuid.UUID) ([]*TaskDependency, error)
	FindDependents(ctx context.Context, taskID uuid.UUID) ([]*TaskDependency, error)
	AddDependency(ctx context.Context, dependency *TaskDependency) error
	RemoveDependency(ctx context.Context, taskID, dependsOnID uuid.UUID) error

	// 评论操作
	FindComments(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]*TaskComment, error)
	AddComment(ctx context.Context, comment *TaskComment) error
	UpdateComment(ctx context.Context, comment *TaskComment) error
	DeleteComment(ctx context.Context, commentID uuid.UUID) error

	// 附件操作
	FindAttachments(ctx context.Context, taskID uuid.UUID) ([]*TaskAttachment, error)
	AddAttachment(ctx context.Context, attachment *TaskAttachment) error
	DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error

	// 时间记录操作
	FindTimeLogs(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]*TaskTimeLog, error)
	AddTimeLog(ctx context.Context, timeLog *TaskTimeLog) error
	UpdateTimeLog(ctx context.Context, timeLog *TaskTimeLog) error
	DeleteTimeLog(ctx context.Context, timeLogID uuid.UUID) error
	GetTimeLogStatistics(ctx context.Context, taskID uuid.UUID) (*TimeLogStatistics, error)
}

// ProjectRepository 项目仓储接口
type ProjectRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 查询操作
	FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Project, error)
	FindByOrganizationID(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]*Project, error)
	FindByStatus(ctx context.Context, status ProjectStatus, limit, offset int) ([]*Project, error)
	FindByType(ctx context.Context, projectType ProjectType, limit, offset int) ([]*Project, error)
	FindByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]*Project, error)

	// 复合查询
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Project, error) // 用户参与的项?
	FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*Project, error)
	FindOverdueProjects(ctx context.Context, limit, offset int) ([]*Project, error)

	// 搜索操作
	SearchProjects(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*Project, error)
	FindByTags(ctx context.Context, tags []string, limit, offset int) ([]*Project, error)

	// 统计操作
	Count(ctx context.Context) (int64, error)
	CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error)
	CountByOrganization(ctx context.Context, organizationID uuid.UUID) (int64, error)
	CountByStatus(ctx context.Context, status ProjectStatus) (int64, error)
	GetProjectStatistics(ctx context.Context, projectID uuid.UUID) (*ProjectStatistics, error)

	// 成员操作
	FindMembers(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)
	AddMember(ctx context.Context, member *ProjectMember) error
	UpdateMember(ctx context.Context, member *ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	FindMembersByRole(ctx context.Context, projectID uuid.UUID, role string) ([]*ProjectMember, error)

	// 里程碑操?
	FindMilestones(ctx context.Context, projectID uuid.UUID) ([]*ProjectMilestone, error)
	AddMilestone(ctx context.Context, milestone *ProjectMilestone) error
	UpdateMilestone(ctx context.Context, milestone *ProjectMilestone) error
	DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error
	FindUpcomingMilestones(ctx context.Context, projectID uuid.UUID, days int) ([]*ProjectMilestone, error)
}

// TeamRepository 团队仓储接口
type TeamRepository interface {
	// 基本CRUD操作
	Save(ctx context.Context, team *Team) error
	FindByID(ctx context.Context, id uuid.UUID) (*Team, error)
	Update(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 查询操作
	FindByLeaderID(ctx context.Context, leaderID uuid.UUID, limit, offset int) ([]*Team, error)
	FindByOrganizationID(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]*Team, error)
	FindByStatus(ctx context.Context, status TeamStatus, limit, offset int) ([]*Team, error)
	FindByParentTeamID(ctx context.Context, parentTeamID uuid.UUID, limit, offset int) ([]*Team, error)

	// 复合查询
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Team, error) // 用户参与的团?
	FindTeamsWithSkill(ctx context.Context, skill string, limit, offset int) ([]*Team, error)

	// 搜索操作
	SearchTeams(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*Team, error)
	FindByTags(ctx context.Context, tags []string, limit, offset int) ([]*Team, error)

	// 统计操作
	Count(ctx context.Context) (int64, error)
	CountByLeader(ctx context.Context, leaderID uuid.UUID) (int64, error)
	CountByOrganization(ctx context.Context, organizationID uuid.UUID) (int64, error)
	GetTeamStatistics(ctx context.Context, teamID uuid.UUID) (*TeamStatistics, error)

	// 成员操作
	FindMembers(ctx context.Context, teamID uuid.UUID) ([]*TeamMember, error)
	AddMember(ctx context.Context, member *TeamMember) error
	UpdateMember(ctx context.Context, member *TeamMember) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	FindMembersByRole(ctx context.Context, teamID uuid.UUID, role TeamMemberRole) ([]*TeamMember, error)
	FindAvailableMembers(ctx context.Context, teamID uuid.UUID, timeSlot string) ([]*TeamMember, error)

	// 技能操?
	FindTeamSkills(ctx context.Context, teamID uuid.UUID) ([]*TeamSkill, error)
	UpdateTeamSkills(ctx context.Context, teamID uuid.UUID, skills []*TeamSkill) error
	FindTeamsBySkills(ctx context.Context, skills []string, minLevel string) ([]*Team, error)

	// 指标操作
	FindTeamMetrics(ctx context.Context, teamID uuid.UUID, period string, startDate, endDate time.Time) ([]*TeamMetrics, error)
	SaveTeamMetrics(ctx context.Context, metrics *TeamMetrics) error
	GetLatestMetrics(ctx context.Context, teamID uuid.UUID) (*TeamMetrics, error)
}

// EventRepository 事件仓储接口
type EventRepository interface {
	// 事件存储
	SaveEvent(ctx context.Context, event DomainEvent) error
	SaveEvents(ctx context.Context, events []DomainEvent) error

	// 事件查询
	FindEventsByAggregateID(ctx context.Context, aggregateID uuid.UUID, limit, offset int) ([]DomainEvent, error)
	FindEventsByType(ctx context.Context, eventType string, limit, offset int) ([]DomainEvent, error)
	FindEventsByTimeRange(ctx context.Context, startTime, endTime time.Time, limit, offset int) ([]DomainEvent, error)

	// 事件?
	GetEventStream(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]DomainEvent, error)
	GetLastEventVersion(ctx context.Context, aggregateID uuid.UUID) (int, error)

	// 事件快照
	SaveSnapshot(ctx context.Context, aggregateID uuid.UUID, version int, data []byte) error
	FindLatestSnapshot(ctx context.Context, aggregateID uuid.UUID) (*EventSnapshot, error)
}

// ========== 统计和分析相关结构体 ==========

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TotalTasks         int                    `json:"total_tasks"`
	CompletedTasks     int                    `json:"completed_tasks"`
	InProgressTasks    int                    `json:"in_progress_tasks"`
	PendingTasks       int                    `json:"pending_tasks"`
	OverdueTasks       int                    `json:"overdue_tasks"`
	CancelledTasks     int                    `json:"cancelled_tasks"`
	CompletionRate     float64                `json:"completion_rate"`
	AverageTaskTime    float64                `json:"average_task_time"`
	TasksByType        map[TaskType]int       `json:"tasks_by_type"`
	TasksByPriority    map[TaskPriority]int   `json:"tasks_by_priority"`
	TasksByComplexity  map[TaskComplexity]int `json:"tasks_by_complexity"`
	ProductivityTrend  []ProductivityPoint    `json:"productivity_trend"`
}

// ProductivityPoint 生产力趋势点
type ProductivityPoint struct {
	Date           time.Time `json:"date"`
	TasksCompleted int       `json:"tasks_completed"`
	AverageTime    float64   `json:"average_time"`
	QualityScore   float64   `json:"quality_score"`
}

// TimeLogStatistics 时间记录统计
type TimeLogStatistics struct {
	TotalHours     float64 `json:"total_hours"`
	AverageSession float64 `json:"average_session"`
	LongestSession float64 `json:"longest_session"`
	ShortestSession float64 `json:"shortest_session"`
	SessionCount   int     `json:"session_count"`
	DailyAverage   float64 `json:"daily_average"`
}

// TeamStatistics 团队统计信息
type TeamStatistics struct {
	TotalMembers       int                        `json:"total_members"`
	ActiveMembers      int                        `json:"active_members"`
	MembersByRole      map[TeamMemberRole]int     `json:"members_by_role"`
	SkillCoverage      map[string]int             `json:"skill_coverage"`
	TeamProductivity   float64                    `json:"team_productivity"`
	CollaborationScore float64                    `json:"collaboration_score"`
	TaskDistribution   map[uuid.UUID]int          `json:"task_distribution"` // 成员ID -> 任务?
	PerformanceMetrics map[string]interface{}     `json:"performance_metrics"`
}

// EventSnapshot 事件快照
type EventSnapshot struct {
	ID          uuid.UUID `json:"id"`
	AggregateID uuid.UUID `json:"aggregate_id"`
	Version     int       `json:"version"`
	Data        []byte    `json:"data"`
	CreatedAt   time.Time `json:"created_at"`
}

// ========== 查询过滤?==========

// TaskFilter 任务查询过滤?
type TaskFilter struct {
	ProjectID    *uuid.UUID     `json:"project_id,omitempty"`
	AssigneeID   *uuid.UUID     `json:"assignee_id,omitempty"`
	CreatorID    *uuid.UUID     `json:"creator_id,omitempty"`
	Status       *TaskStatus    `json:"status,omitempty"`
	Priority     *TaskPriority  `json:"priority,omitempty"`
	Type         *TaskType      `json:"type,omitempty"`
	Complexity   *TaskComplexity `json:"complexity,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	StartDate    *time.Time     `json:"start_date,omitempty"`
	EndDate      *time.Time     `json:"end_date,omitempty"`
	DueDateFrom  *time.Time     `json:"due_date_from,omitempty"`
	DueDateTo    *time.Time     `json:"due_date_to,omitempty"`
	IsOverdue    *bool          `json:"is_overdue,omitempty"`
	HasAttachments *bool        `json:"has_attachments,omitempty"`
	HasComments  *bool          `json:"has_comments,omitempty"`
	ProgressMin  *float64       `json:"progress_min,omitempty"`
	ProgressMax  *float64       `json:"progress_max,omitempty"`
}

// ProjectFilter 项目查询过滤?
type ProjectFilter struct {
	OwnerID        *uuid.UUID       `json:"owner_id,omitempty"`
	OrganizationID *uuid.UUID       `json:"organization_id,omitempty"`
	TeamID         *uuid.UUID       `json:"team_id,omitempty"`
	Status         *ProjectStatus   `json:"status,omitempty"`
	Priority       *ProjectPriority `json:"priority,omitempty"`
	Type           *ProjectType     `json:"type,omitempty"`
	Tags           []string         `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	StartDate      *time.Time       `json:"start_date,omitempty"`
	EndDate        *time.Time       `json:"end_date,omitempty"`
	DueDateFrom    *time.Time       `json:"due_date_from,omitempty"`
	DueDateTo      *time.Time       `json:"due_date_to,omitempty"`
	IsOverdue      *bool            `json:"is_overdue,omitempty"`
	BudgetMin      *float64         `json:"budget_min,omitempty"`
	BudgetMax      *float64         `json:"budget_max,omitempty"`
	ProgressMin    *float64         `json:"progress_min,omitempty"`
	ProgressMax    *float64         `json:"progress_max,omitempty"`
}

// TeamFilter 团队查询过滤?
type TeamFilter struct {
	LeaderID       *uuid.UUID    `json:"leader_id,omitempty"`
	OrganizationID *uuid.UUID    `json:"organization_id,omitempty"`
	ParentTeamID   *uuid.UUID    `json:"parent_team_id,omitempty"`
	Status         *TeamStatus   `json:"status,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Skills         []string      `json:"skills,omitempty"`
	MinMembers     *int          `json:"min_members,omitempty"`
	MaxMembers     *int          `json:"max_members,omitempty"`
	TimeZone       *string       `json:"time_zone,omitempty"`
}

// SortOption 排序选项
type SortOption struct {
	Field string `json:"field"`
	Order string `json:"order"` // asc, desc
}

// PaginationOption 分页选项
type PaginationOption struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// QueryOptions 查询选项
type QueryOptions struct {
	Pagination *PaginationOption `json:"pagination,omitempty"`
	Sort       []SortOption      `json:"sort,omitempty"`
	Include    []string          `json:"include,omitempty"` // 包含的关联数?
}

// ========== 仓储工厂接口 ==========

// RepositoryFactory 仓储工厂接口
type RepositoryFactory interface {
	TaskRepository() TaskRepository
	ProjectRepository() ProjectRepository
	TeamRepository() TeamRepository
	EventRepository() EventRepository
}

// UnitOfWork 工作单元接口
type UnitOfWork interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	TaskRepository() TaskRepository
	ProjectRepository() ProjectRepository
	TeamRepository() TeamRepository
	EventRepository() EventRepository
}

