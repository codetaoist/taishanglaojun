package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// TaskService 任务应用服务
type TaskService struct {
	taskRepo         domain.TaskRepository
	projectRepo      domain.ProjectRepository
	teamRepo         domain.TeamRepository
	eventRepo        domain.EventRepository
	allocationSvc    domain.TaskAllocationService
	schedulingSvc    domain.TaskSchedulingService
	performanceSvc   domain.PerformanceAnalysisService
	notificationSvc  domain.NotificationService
	unitOfWork       domain.UnitOfWork
}

// NewTaskService 创建任务服务实例
func NewTaskService(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	teamRepo domain.TeamRepository,
	eventRepo domain.EventRepository,
	allocationSvc domain.TaskAllocationService,
	schedulingSvc domain.TaskSchedulingService,
	performanceSvc domain.PerformanceAnalysisService,
	notificationSvc domain.NotificationService,
	unitOfWork domain.UnitOfWork,
) *TaskService {
	return &TaskService{
		taskRepo:        taskRepo,
		projectRepo:     projectRepo,
		teamRepo:        teamRepo,
		eventRepo:       eventRepo,
		allocationSvc:   allocationSvc,
		schedulingSvc:   schedulingSvc,
		performanceSvc:  performanceSvc,
		notificationSvc: notificationSvc,
		unitOfWork:      unitOfWork,
	}
}

// ========== 任务CRUD操作 ==========

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title           string                 `json:"title" validate:"required,min=1,max=255"`
	Description     string                 `json:"description"`
	Priority        domain.TaskPriority    `json:"priority" validate:"required"`
	Type            domain.TaskType        `json:"type" validate:"required"`
	Complexity      domain.TaskComplexity  `json:"complexity" validate:"required"`
	ProjectID       *uuid.UUID             `json:"project_id"`
	OrganizationID  uuid.UUID              `json:"organization_id" validate:"required"`
	CreatorID       uuid.UUID              `json:"creator_id" validate:"required"`
	AssigneeID      *uuid.UUID             `json:"assignee_id"`
	DueDate         *time.Time             `json:"due_date"`
	EstimatedHours  *float64               `json:"estimated_hours"`
	Tags            []string               `json:"tags"`
	Labels          map[string]string      `json:"labels"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// CreateTaskResponse 创建任务响应
type CreateTaskResponse struct {
	Task *domain.Task `json:"task"`
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*CreateTaskResponse, error) {
	// 验证项目是否存在
	if req.ProjectID != nil {
		project, err := s.projectRepo.FindByID(ctx, *req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to find project: %w", err)
		}
		if project == nil {
			return nil, domain.ErrProjectNotFound
		}
	}

	// 创建任务
	task, err := domain.NewTask(
		req.Title,
		req.Description,
		req.Priority,
		req.Type,
		req.Complexity,
		req.ProjectID,
		req.OrganizationID,
		req.CreatorID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 设置可选字�?
	if req.AssigneeID != nil {
		if err := task.AssignTo(*req.AssigneeID); err != nil {
			return nil, fmt.Errorf("failed to assign task: %w", err)
		}
	}

	if req.DueDate != nil {
		task.SetDueDate(*req.DueDate)
	}

	if req.EstimatedHours != nil {
		task.SetEstimatedHours(*req.EstimatedHours)
	}

	if len(req.Tags) > 0 {
		task.Tags = req.Tags
	}

	if len(req.Labels) > 0 {
		task.Labels = req.Labels
	}

	if len(req.Metadata) > 0 {
		task.Metadata = req.Metadata
	}

	// 保存任务
	if err := s.taskRepo.Save(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	// 发布领域事件
	for _, event := range task.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 发送通知
	if req.AssigneeID != nil {
		notification := &domain.TaskNotification{
			Type:     domain.NotificationTypeTaskAssigned,
			TaskID:   task.ID,
			UserID:   *req.AssigneeID,
			Title:    "新任务分�?,
			Message:  fmt.Sprintf("您被分配了新任务: %s", task.Title),
			Metadata: map[string]interface{}{"task_id": task.ID.String()},
		}
		if err := s.notificationSvc.SendTaskNotification(ctx, notification); err != nil {
			// 记录错误但不中断流程
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}

	task.ClearDomainEvents()

	return &CreateTaskResponse{Task: task}, nil
}

// GetTaskRequest 获取任务请求
type GetTaskRequest struct {
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// GetTaskResponse 获取任务响应
type GetTaskResponse struct {
	Task *domain.Task `json:"task"`
}

// GetTask 获取任务
func (s *TaskService) GetTask(ctx context.Context, req *GetTaskRequest) (*GetTaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &GetTaskResponse{Task: task}, nil
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	TaskID          uuid.UUID              `json:"task_id" validate:"required"`
	Title           *string                `json:"title,omitempty"`
	Description     *string                `json:"description,omitempty"`
	Priority        *domain.TaskPriority   `json:"priority,omitempty"`
	Status          *domain.TaskStatus     `json:"status,omitempty"`
	AssigneeID      *uuid.UUID             `json:"assignee_id,omitempty"`
	DueDate         *time.Time             `json:"due_date,omitempty"`
	EstimatedHours  *float64               `json:"estimated_hours,omitempty"`
	Progress        *float64               `json:"progress,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTaskResponse 更新任务响应
type UpdateTaskResponse struct {
	Task *domain.Task `json:"task"`
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(ctx context.Context, req *UpdateTaskRequest) (*UpdateTaskResponse, error) {
	// 获取现有任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 更新字段
	if req.Title != nil {
		task.Title = *req.Title
	}

	if req.Description != nil {
		task.Description = *req.Description
	}

	if req.Priority != nil {
		if err := task.SetPriority(*req.Priority); err != nil {
			return nil, fmt.Errorf("failed to set priority: %w", err)
		}
	}

	if req.Status != nil {
		switch *req.Status {
		case domain.TaskStatusInProgress:
			if err := task.Start(); err != nil {
				return nil, fmt.Errorf("failed to start task: %w", err)
			}
		case domain.TaskStatusCompleted:
			if err := task.Complete(); err != nil {
				return nil, fmt.Errorf("failed to complete task: %w", err)
			}
		case domain.TaskStatusCancelled:
			if err := task.Cancel(); err != nil {
				return nil, fmt.Errorf("failed to cancel task: %w", err)
			}
		}
	}

	if req.AssigneeID != nil {
		if task.AssigneeID != nil && *task.AssigneeID != *req.AssigneeID {
			// 重新分配任务
			task.Unassign()
		}
		if err := task.AssignTo(*req.AssigneeID); err != nil {
			return nil, fmt.Errorf("failed to assign task: %w", err)
		}
	}

	if req.DueDate != nil {
		task.SetDueDate(*req.DueDate)
	}

	if req.EstimatedHours != nil {
		task.SetEstimatedHours(*req.EstimatedHours)
	}

	if req.Progress != nil {
		if err := task.UpdateProgress(*req.Progress); err != nil {
			return nil, fmt.Errorf("failed to update progress: %w", err)
		}
	}

	if req.Tags != nil {
		task.Tags = req.Tags
	}

	if req.Labels != nil {
		task.Labels = req.Labels
	}

	if req.Metadata != nil {
		task.Metadata = req.Metadata
	}

	// 保存更新
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// 发布领域事件
	for _, event := range task.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	task.ClearDomainEvents()

	return &UpdateTaskResponse{Task: task}, nil
}

// DeleteTaskRequest 删除任务请求
type DeleteTaskRequest struct {
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(ctx context.Context, req *DeleteTaskRequest) error {
	// 检查任务是否存�?
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return fmt.Errorf("failed to find task: %w", err)
	}

	// 检查任务是否可以删�?
	if task.Status == domain.TaskStatusInProgress {
		return fmt.Errorf("cannot delete task in progress")
	}

	// 删除任务
	if err := s.taskRepo.Delete(ctx, req.TaskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// ========== 任务查询操作 ==========

// ListTasksRequest 列表任务请求
type ListTasksRequest struct {
	ProjectID      *uuid.UUID            `json:"project_id,omitempty"`
	AssigneeID     *uuid.UUID            `json:"assignee_id,omitempty"`
	CreatorID      *uuid.UUID            `json:"creator_id,omitempty"`
	Status         *domain.TaskStatus    `json:"status,omitempty"`
	Priority       *domain.TaskPriority  `json:"priority,omitempty"`
	Type           *domain.TaskType      `json:"type,omitempty"`
	OrganizationID *uuid.UUID            `json:"organization_id,omitempty"`
	Tags           []string              `json:"tags,omitempty"`
	Labels         map[string]string     `json:"labels,omitempty"`
	IsOverdue      *bool                 `json:"is_overdue,omitempty"`
	StartDate      *time.Time            `json:"start_date,omitempty"`
	EndDate        *time.Time            `json:"end_date,omitempty"`
	Limit          int                   `json:"limit" validate:"min=1,max=100"`
	Offset         int                   `json:"offset" validate:"min=0"`
	SortBy         string                `json:"sort_by"`
	SortOrder      string                `json:"sort_order"`
}

// ListTasksResponse 列表任务响应
type ListTasksResponse struct {
	Tasks      []*domain.Task `json:"tasks"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	HasMore    bool           `json:"has_more"`
}

// ListTasks 列表任务
func (s *TaskService) ListTasks(ctx context.Context, req *ListTasksRequest) (*ListTasksResponse, error) {
	// 构建查询选项
	options := &domain.QueryOptions{
		Pagination: &domain.PaginationOption{
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	}

	if req.SortBy != "" {
		options.Sort = &domain.SortOption{
			Field: req.SortBy,
			Order: req.SortOrder,
		}
	}

	// 构建过滤�?
	filters := make(map[string]interface{})
	if req.ProjectID != nil {
		filters["project_id"] = *req.ProjectID
	}
	if req.AssigneeID != nil {
		filters["assignee_id"] = *req.AssigneeID
	}
	if req.CreatorID != nil {
		filters["creator_id"] = *req.CreatorID
	}
	if req.Status != nil {
		filters["status"] = string(*req.Status)
	}
	if req.Priority != nil {
		filters["priority"] = string(*req.Priority)
	}
	if req.Type != nil {
		filters["type"] = string(*req.Type)
	}
	if req.OrganizationID != nil {
		filters["organization_id"] = *req.OrganizationID
	}
	if req.IsOverdue != nil {
		filters["is_overdue"] = *req.IsOverdue
	}
	if req.StartDate != nil {
		filters["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		filters["end_date"] = *req.EndDate
	}

	var tasks []*domain.Task
	var err error

	// 根据不同条件选择查询方法
	if len(req.Tags) > 0 {
		tasks, err = s.taskRepo.FindByTags(ctx, req.Tags, req.Limit, req.Offset)
	} else if len(req.Labels) > 0 {
		tasks, err = s.taskRepo.FindByLabels(ctx, req.Labels, req.Limit, req.Offset)
	} else {
		tasks, err = s.taskRepo.SearchTasks(ctx, "", filters, req.Limit, req.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// 获取总数
	total, err := s.taskRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	hasMore := int64(req.Offset+len(tasks)) < total

	return &ListTasksResponse{
		Tasks:   tasks,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: hasMore,
	}, nil
}

// SearchTasksRequest 搜索任务请求
type SearchTasksRequest struct {
	Query          string                `json:"query" validate:"required"`
	Filters        map[string]interface{} `json:"filters,omitempty"`
	Limit          int                   `json:"limit" validate:"min=1,max=100"`
	Offset         int                   `json:"offset" validate:"min=0"`
}

// SearchTasksResponse 搜索任务响应
type SearchTasksResponse struct {
	Tasks   []*domain.Task `json:"tasks"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}

// SearchTasks 搜索任务
func (s *TaskService) SearchTasks(ctx context.Context, req *SearchTasksRequest) (*SearchTasksResponse, error) {
	tasks, err := s.taskRepo.SearchTasks(ctx, req.Query, req.Filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}

	// 获取总数（这里简化处理，实际应该根据搜索条件计算�?
	total, err := s.taskRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	hasMore := int64(req.Offset+len(tasks)) < total

	return &SearchTasksResponse{
		Tasks:   tasks,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: hasMore,
	}, nil
}

// ========== 任务分配操作 ==========

// AssignTaskRequest 分配任务请求
type AssignTaskRequest struct {
	TaskID     uuid.UUID `json:"task_id" validate:"required"`
	AssigneeID uuid.UUID `json:"assignee_id" validate:"required"`
}

// AssignTaskResponse 分配任务响应
type AssignTaskResponse struct {
	Task *domain.Task `json:"task"`
}

// AssignTask 分配任务
func (s *TaskService) AssignTask(ctx context.Context, req *AssignTaskRequest) (*AssignTaskResponse, error) {
	// 获取任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 分配任务
	if err := task.AssignTo(req.AssigneeID); err != nil {
		return nil, fmt.Errorf("failed to assign task: %w", err)
	}

	// 保存更新
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// 发布领域事件
	for _, event := range task.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 发送通知
	notification := &domain.TaskNotification{
		Type:     domain.NotificationTypeTaskAssigned,
		TaskID:   task.ID,
		UserID:   req.AssigneeID,
		Title:    "任务分配",
		Message:  fmt.Sprintf("您被分配了任�? %s", task.Title),
		Metadata: map[string]interface{}{"task_id": task.ID.String()},
	}
	if err := s.notificationSvc.SendTaskNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	task.ClearDomainEvents()

	return &AssignTaskResponse{Task: task}, nil
}

// UnassignTaskRequest 取消分配任务请求
type UnassignTaskRequest struct {
	TaskID uuid.UUID `json:"task_id" validate:"required"`
}

// UnassignTaskResponse 取消分配任务响应
type UnassignTaskResponse struct {
	Task *domain.Task `json:"task"`
}

// UnassignTask 取消分配任务
func (s *TaskService) UnassignTask(ctx context.Context, req *UnassignTaskRequest) (*UnassignTaskResponse, error) {
	// 获取任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 取消分配
	task.Unassign()

	// 保存更新
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	// 发布领域事件
	for _, event := range task.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	task.ClearDomainEvents()

	return &UnassignTaskResponse{Task: task}, nil
}

// ========== 智能分配操作 ==========

// AutoAssignTasksRequest 自动分配任务请求
type AutoAssignTasksRequest struct {
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	TeamID         *uuid.UUID `json:"team_id,omitempty"`
	MaxTasks       int        `json:"max_tasks" validate:"min=1,max=100"`
}

// AutoAssignTasksResponse 自动分配任务响应
type AutoAssignTasksResponse struct {
	AssignedTasks []*domain.Task                      `json:"assigned_tasks"`
	Assignments   []*domain.TaskAllocationResult      `json:"assignments"`
	Summary       *domain.TaskAllocationSummary       `json:"summary"`
}

// AutoAssignTasks 自动分配任务
func (s *TaskService) AutoAssignTasks(ctx context.Context, req *AutoAssignTasksRequest) (*AutoAssignTasksResponse, error) {
	// 构建分配请求
	allocReq := &domain.TaskAllocationRequest{
		OrganizationID: req.OrganizationID,
		ProjectID:      req.ProjectID,
		TeamID:         req.TeamID,
		MaxTasks:       req.MaxTasks,
		Strategy:       domain.AllocationStrategyBalanced, // 默认使用平衡策略
	}

	// 执行智能分配
	result, err := s.allocationSvc.AllocateTasks(ctx, allocReq)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate tasks: %w", err)
	}

	// 应用分配结果
	var assignedTasks []*domain.Task
	for _, assignment := range result.Assignments {
		task, err := s.taskRepo.FindByID(ctx, assignment.TaskID)
		if err != nil {
			fmt.Printf("failed to find task %s: %v\n", assignment.TaskID, err)
			continue
		}

		if err := task.AssignTo(assignment.AssigneeID); err != nil {
			fmt.Printf("failed to assign task %s: %v\n", assignment.TaskID, err)
			continue
		}

		if err := s.taskRepo.Update(ctx, task); err != nil {
			fmt.Printf("failed to update task %s: %v\n", assignment.TaskID, err)
			continue
		}

		assignedTasks = append(assignedTasks, task)

		// 发送通知
		notification := &domain.TaskNotification{
			Type:     domain.NotificationTypeTaskAssigned,
			TaskID:   task.ID,
			UserID:   assignment.AssigneeID,
			Title:    "自动任务分配",
			Message:  fmt.Sprintf("系统为您自动分配了任�? %s", task.Title),
			Metadata: map[string]interface{}{
				"task_id":     task.ID.String(),
				"auto_assign": true,
			},
		}
		if err := s.notificationSvc.SendTaskNotification(ctx, notification); err != nil {
			fmt.Printf("failed to send notification: %v\n", err)
		}
	}

	return &AutoAssignTasksResponse{
		AssignedTasks: assignedTasks,
		Assignments:   result.Assignments,
		Summary:       result.Summary,
	}, nil
}

// ========== 任务统计操作 ==========

// GetTaskStatisticsRequest 获取任务统计请求
type GetTaskStatisticsRequest struct {
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	AssigneeID     *uuid.UUID `json:"assignee_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// GetTaskStatisticsResponse 获取任务统计响应
type GetTaskStatisticsResponse struct {
	Statistics *domain.TaskStatistics `json:"statistics"`
}

// GetTaskStatistics 获取任务统计
func (s *TaskService) GetTaskStatistics(ctx context.Context, req *GetTaskStatisticsRequest) (*GetTaskStatisticsResponse, error) {
	stats, err := s.taskRepo.GetTaskStatistics(ctx, req.ProjectID, req.AssigneeID, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task statistics: %w", err)
	}

	return &GetTaskStatisticsResponse{Statistics: stats}, nil
}

// ========== 任务依赖操作 ==========

// AddTaskDependencyRequest 添加任务依赖请求
type AddTaskDependencyRequest struct {
	TaskID           uuid.UUID                   `json:"task_id" validate:"required"`
	DependsOnTaskID  uuid.UUID                   `json:"depends_on_task_id" validate:"required"`
	DependencyType   domain.TaskDependencyType   `json:"dependency_type" validate:"required"`
}

// AddTaskDependencyResponse 添加任务依赖响应
type AddTaskDependencyResponse struct {
	Task *domain.Task `json:"task"`
}

// AddTaskDependency 添加任务依赖
func (s *TaskService) AddTaskDependency(ctx context.Context, req *AddTaskDependencyRequest) (*AddTaskDependencyResponse, error) {
	// 获取任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 验证依赖任务是否存在
	dependsOnTask, err := s.taskRepo.FindByID(ctx, req.DependsOnTaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find dependency task: %w", err)
	}
	if dependsOnTask == nil {
		return nil, fmt.Errorf("dependency task not found")
	}

	// 添加依赖
	dependency := &domain.TaskDependency{
		ID:              uuid.New(),
		TaskID:          req.TaskID,
		DependsOnTaskID: req.DependsOnTaskID,
		DependencyType:  req.DependencyType,
		CreatedAt:       time.Now(),
	}

	if err := task.AddDependency(dependency); err != nil {
		return nil, fmt.Errorf("failed to add dependency: %w", err)
	}

	// 保存依赖
	if err := s.taskRepo.AddDependency(ctx, dependency); err != nil {
		return nil, fmt.Errorf("failed to save dependency: %w", err)
	}

	return &AddTaskDependencyResponse{Task: task}, nil
}

// RemoveTaskDependencyRequest 移除任务依赖请求
type RemoveTaskDependencyRequest struct {
	TaskID       uuid.UUID `json:"task_id" validate:"required"`
	DependencyID uuid.UUID `json:"dependency_id" validate:"required"`
}

// RemoveTaskDependencyResponse 移除任务依赖响应
type RemoveTaskDependencyResponse struct {
	Task *domain.Task `json:"task"`
}

// RemoveTaskDependency 移除任务依赖
func (s *TaskService) RemoveTaskDependency(ctx context.Context, req *RemoveTaskDependencyRequest) (*RemoveTaskDependencyResponse, error) {
	// 获取任务
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// 移除依赖
	if err := s.taskRepo.RemoveDependency(ctx, req.DependencyID); err != nil {
		return nil, fmt.Errorf("failed to remove dependency: %w", err)
	}

	// 重新加载任务以获取最新的依赖关系
	task, err = s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload task: %w", err)
	}

	return &RemoveTaskDependencyResponse{Task: task}, nil
}

// ========== 任务评论操作 ==========

// AddTaskCommentRequest 添加任务评论请求
type AddTaskCommentRequest struct {
	TaskID   uuid.UUID `json:"task_id" validate:"required"`
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	Content  string    `json:"content" validate:"required,min=1"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// AddTaskCommentResponse 添加任务评论响应
type AddTaskCommentResponse struct {
	Comment *domain.TaskComment `json:"comment"`
}

// AddTaskComment 添加任务评论
func (s *TaskService) AddTaskComment(ctx context.Context, req *AddTaskCommentRequest) (*AddTaskCommentResponse, error) {
	// 验证任务是否存在
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	// 创建评论
	comment := &domain.TaskComment{
		ID:        uuid.New(),
		TaskID:    req.TaskID,
		UserID:    req.UserID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存评论
	if err := s.taskRepo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return &AddTaskCommentResponse{Comment: comment}, nil
}

// ========== 任务时间记录操作 ==========

// AddTimeLogRequest 添加时间记录请求
type AddTimeLogRequest struct {
	TaskID      uuid.UUID `json:"task_id" validate:"required"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Duration    float64   `json:"duration" validate:"required,min=0.1"`
	Description string    `json:"description"`
	LogDate     time.Time `json:"log_date" validate:"required"`
}

// AddTimeLogResponse 添加时间记录响应
type AddTimeLogResponse struct {
	TimeLog *domain.TaskTimeLog `json:"time_log"`
}

// AddTimeLog 添加时间记录
func (s *TaskService) AddTimeLog(ctx context.Context, req *AddTimeLogRequest) (*AddTimeLogResponse, error) {
	// 验证任务是否存在
	task, err := s.taskRepo.FindByID(ctx, req.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	if task == nil {
		return nil, domain.ErrTaskNotFound
	}

	// 创建时间记录
	timeLog := &domain.TaskTimeLog{
		ID:          uuid.New(),
		TaskID:      req.TaskID,
		UserID:      req.UserID,
		Duration:    req.Duration,
		Description: req.Description,
		LogDate:     req.LogDate,
		CreatedAt:   time.Now(),
	}

	// 保存时间记录
	if err := s.taskRepo.AddTimeLog(ctx, timeLog); err != nil {
		return nil, fmt.Errorf("failed to add time log: %w", err)
	}

	return &AddTimeLogResponse{TimeLog: timeLog}, nil
}
