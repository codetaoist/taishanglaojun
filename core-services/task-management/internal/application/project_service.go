package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// ProjectService 项目应用服务
type ProjectService struct {
	projectRepo     domain.ProjectRepository
	taskRepo        domain.TaskRepository
	teamRepo        domain.TeamRepository
	eventRepo       domain.EventRepository
	schedulingSvc   domain.TaskSchedulingService
	performanceSvc  domain.PerformanceAnalysisService
	notificationSvc domain.NotificationService
	unitOfWork      domain.UnitOfWork
}

// NewProjectService 创建项目服务实例
func NewProjectService(
	projectRepo domain.ProjectRepository,
	taskRepo domain.TaskRepository,
	teamRepo domain.TeamRepository,
	eventRepo domain.EventRepository,
	schedulingSvc domain.TaskSchedulingService,
	performanceSvc domain.PerformanceAnalysisService,
	notificationSvc domain.NotificationService,
	unitOfWork domain.UnitOfWork,
) *ProjectService {
	return &ProjectService{
		projectRepo:     projectRepo,
		taskRepo:        taskRepo,
		teamRepo:        teamRepo,
		eventRepo:       eventRepo,
		schedulingSvc:   schedulingSvc,
		performanceSvc:  performanceSvc,
		notificationSvc: notificationSvc,
		unitOfWork:      unitOfWork,
	}
}

// ========== 项目CRUD操作 ==========

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Description    string                 `json:"description"`
	Priority       domain.ProjectPriority `json:"priority" validate:"required"`
	Type           domain.ProjectType     `json:"type" validate:"required"`
	OrganizationID uuid.UUID              `json:"organization_id" validate:"required"`
	OwnerID        uuid.UUID              `json:"owner_id" validate:"required"`
	ManagerID      *uuid.UUID             `json:"manager_id"`
	StartDate      *time.Time             `json:"start_date"`
	EndDate        *time.Time             `json:"end_date"`
	Budget         *float64               `json:"budget"`
	Tags           []string               `json:"tags"`
	Labels         map[string]string      `json:"labels"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CreateProjectResponse 创建项目响应
type CreateProjectResponse struct {
	Project *domain.Project `json:"project"`
}

// CreateProject 创建项目
func (s *ProjectService) CreateProject(ctx context.Context, req *CreateProjectRequest) (*CreateProjectResponse, error) {
	// 创建项目
	project, err := domain.NewProject(
		req.Name,
		req.Description,
		req.Priority,
		req.Type,
		req.OrganizationID,
		req.OwnerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// 设置可选字�?
	if req.ManagerID != nil {
		project.ManagerID = req.ManagerID
	}

	if req.StartDate != nil {
		project.StartDate = req.StartDate
	}

	if req.EndDate != nil {
		project.SetEndDate(*req.EndDate)
	}

	if req.Budget != nil {
		project.SetBudget(*req.Budget)
	}

	if len(req.Tags) > 0 {
		project.Tags = req.Tags
	}

	if len(req.Labels) > 0 {
		project.Labels = req.Labels
	}

	if len(req.Metadata) > 0 {
		project.Metadata = req.Metadata
	}

	// 保存项目
	if err := s.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	// 发布领域事件
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 发送通知
	notification := &domain.ProjectNotification{
		Type:      domain.NotificationTypeProjectCreated,
		ProjectID: project.ID,
		UserID:    req.OwnerID,
		Title:     "项目创建成功",
		Message:   fmt.Sprintf("项目 %s 已成功创�?, project.Name),
		Metadata:  map[string]interface{}{"project_id": project.ID.String()},
	}
	if err := s.notificationSvc.SendProjectNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	project.ClearDomainEvents()

	return &CreateProjectResponse{Project: project}, nil
}

// GetProjectRequest 获取项目请求
type GetProjectRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// GetProjectResponse 获取项目响应
type GetProjectResponse struct {
	Project *domain.Project `json:"project"`
}

// GetProject 获取项目
func (s *ProjectService) GetProject(ctx context.Context, req *GetProjectRequest) (*GetProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	return &GetProjectResponse{Project: project}, nil
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	ProjectID   uuid.UUID              `json:"project_id" validate:"required"`
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Priority    *domain.ProjectPriority `json:"priority,omitempty"`
	Status      *domain.ProjectStatus  `json:"status,omitempty"`
	ManagerID   *uuid.UUID             `json:"manager_id,omitempty"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	EndDate     *time.Time             `json:"end_date,omitempty"`
	Budget      *float64               `json:"budget,omitempty"`
	Progress    *float64               `json:"progress,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateProjectResponse 更新项目响应
type UpdateProjectResponse struct {
	Project *domain.Project `json:"project"`
}

// UpdateProject 更新项目
func (s *ProjectService) UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*UpdateProjectResponse, error) {
	// 获取现有项目
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 更新字段
	if req.Name != nil {
		project.Name = *req.Name
	}

	if req.Description != nil {
		project.Description = *req.Description
	}

	if req.Priority != nil {
		project.Priority = *req.Priority
	}

	if req.Status != nil {
		switch *req.Status {
		case domain.ProjectStatusActive:
			if err := project.Start(); err != nil {
				return nil, fmt.Errorf("failed to start project: %w", err)
			}
		case domain.ProjectStatusCompleted:
			if err := project.Complete(); err != nil {
				return nil, fmt.Errorf("failed to complete project: %w", err)
			}
		case domain.ProjectStatusCancelled:
			if err := project.Cancel(); err != nil {
				return nil, fmt.Errorf("failed to cancel project: %w", err)
			}
		case domain.ProjectStatusPaused:
			if err := project.Pause(); err != nil {
				return nil, fmt.Errorf("failed to pause project: %w", err)
			}
		}
	}

	if req.ManagerID != nil {
		project.ManagerID = req.ManagerID
	}

	if req.StartDate != nil {
		project.StartDate = req.StartDate
	}

	if req.EndDate != nil {
		project.SetEndDate(*req.EndDate)
	}

	if req.Budget != nil {
		project.SetBudget(*req.Budget)
	}

	if req.Progress != nil {
		if err := project.UpdateProgress(*req.Progress); err != nil {
			return nil, fmt.Errorf("failed to update progress: %w", err)
		}
	}

	if req.Tags != nil {
		project.Tags = req.Tags
	}

	if req.Labels != nil {
		project.Labels = req.Labels
	}

	if req.Metadata != nil {
		project.Metadata = req.Metadata
	}

	// 保存更新
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// 发布领域事件
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return &UpdateProjectResponse{Project: project}, nil
}

// DeleteProjectRequest 删除项目请求
type DeleteProjectRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// DeleteProject 删除项目
func (s *ProjectService) DeleteProject(ctx context.Context, req *DeleteProjectRequest) error {
	// 检查项目是否存�?
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// 检查项目是否可以删�?
	if project.Status == domain.ProjectStatusActive {
		return fmt.Errorf("cannot delete active project")
	}

	// 检查是否有关联的任�?
	taskCount, err := s.taskRepo.CountByProject(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to count project tasks: %w", err)
	}
	if taskCount > 0 {
		return fmt.Errorf("cannot delete project with existing tasks")
	}

	// 删除项目
	if err := s.projectRepo.Delete(ctx, req.ProjectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ========== 项目查询操作 ==========

// ListProjectsRequest 列表项目请求
type ListProjectsRequest struct {
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty"`
	OwnerID        *uuid.UUID             `json:"owner_id,omitempty"`
	ManagerID      *uuid.UUID             `json:"manager_id,omitempty"`
	Status         *domain.ProjectStatus  `json:"status,omitempty"`
	Priority       *domain.ProjectPriority `json:"priority,omitempty"`
	Type           *domain.ProjectType    `json:"type,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Labels         map[string]string      `json:"labels,omitempty"`
	IsOverdue      *bool                  `json:"is_overdue,omitempty"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	Limit          int                    `json:"limit" validate:"min=1,max=100"`
	Offset         int                    `json:"offset" validate:"min=0"`
	SortBy         string                 `json:"sort_by"`
	SortOrder      string                 `json:"sort_order"`
}

// ListProjectsResponse 列表项目响应
type ListProjectsResponse struct {
	Projects []*domain.Project `json:"projects"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasMore  bool              `json:"has_more"`
}

// ListProjects 列表项目
func (s *ProjectService) ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
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
	if req.OrganizationID != nil {
		filters["organization_id"] = *req.OrganizationID
	}
	if req.OwnerID != nil {
		filters["owner_id"] = *req.OwnerID
	}
	if req.ManagerID != nil {
		filters["manager_id"] = *req.ManagerID
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
	if req.IsOverdue != nil {
		filters["is_overdue"] = *req.IsOverdue
	}
	if req.StartDate != nil {
		filters["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		filters["end_date"] = *req.EndDate
	}

	var projects []*domain.Project
	var err error

	// 根据不同条件选择查询方法
	if len(req.Tags) > 0 {
		projects, err = s.projectRepo.FindByTags(ctx, req.Tags, req.Limit, req.Offset)
	} else if len(req.Labels) > 0 {
		projects, err = s.projectRepo.FindByLabels(ctx, req.Labels, req.Limit, req.Offset)
	} else {
		projects, err = s.projectRepo.SearchProjects(ctx, "", filters, req.Limit, req.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// 获取总数
	total, err := s.projectRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	hasMore := int64(req.Offset+len(projects)) < total

	return &ListProjectsResponse{
		Projects: projects,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
		HasMore:  hasMore,
	}, nil
}

// SearchProjectsRequest 搜索项目请求
type SearchProjectsRequest struct {
	Query   string                 `json:"query" validate:"required"`
	Filters map[string]interface{} `json:"filters,omitempty"`
	Limit   int                    `json:"limit" validate:"min=1,max=100"`
	Offset  int                    `json:"offset" validate:"min=0"`
}

// SearchProjectsResponse 搜索项目响应
type SearchProjectsResponse struct {
	Projects []*domain.Project `json:"projects"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasMore  bool              `json:"has_more"`
}

// SearchProjects 搜索项目
func (s *ProjectService) SearchProjects(ctx context.Context, req *SearchProjectsRequest) (*SearchProjectsResponse, error) {
	projects, err := s.projectRepo.SearchProjects(ctx, req.Query, req.Filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	// 获取总数（这里简化处理，实际应该根据搜索条件计算�?
	total, err := s.projectRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	hasMore := int64(req.Offset+len(projects)) < total

	return &SearchProjectsResponse{
		Projects: projects,
		Total:    total,
		Limit:    req.Limit,
		Offset:   req.Offset,
		HasMore:  hasMore,
	}, nil
}

// ========== 项目成员管理 ==========

// AddProjectMemberRequest 添加项目成员请求
type AddProjectMemberRequest struct {
	ProjectID uuid.UUID                  `json:"project_id" validate:"required"`
	UserID    uuid.UUID                  `json:"user_id" validate:"required"`
	Role      domain.ProjectMemberRole   `json:"role" validate:"required"`
}

// AddProjectMemberResponse 添加项目成员响应
type AddProjectMemberResponse struct {
	Member *domain.ProjectMember `json:"member"`
}

// AddProjectMember 添加项目成员
func (s *ProjectService) AddProjectMember(ctx context.Context, req *AddProjectMemberRequest) (*AddProjectMemberResponse, error) {
	// 获取项目
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 创建成员
	member := &domain.ProjectMember{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		UserID:    req.UserID,
		Role:      req.Role,
		JoinedAt:  time.Now(),
	}

	// 添加成员
	if err := project.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// 保存成员
	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to save member: %w", err)
	}

	// 发布领域事件
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 发送通知
	notification := &domain.ProjectNotification{
		Type:      domain.NotificationTypeProjectMemberAdded,
		ProjectID: project.ID,
		UserID:    req.UserID,
		Title:     "项目成员邀�?,
		Message:   fmt.Sprintf("您已被邀请加入项�? %s", project.Name),
		Metadata: map[string]interface{}{
			"project_id": project.ID.String(),
			"role":       string(req.Role),
		},
	}
	if err := s.notificationSvc.SendProjectNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	project.ClearDomainEvents()

	return &AddProjectMemberResponse{Member: member}, nil
}

// RemoveProjectMemberRequest 移除项目成员请求
type RemoveProjectMemberRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
}

// RemoveProjectMember 移除项目成员
func (s *ProjectService) RemoveProjectMember(ctx context.Context, req *RemoveProjectMemberRequest) error {
	// 获取项目
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// 移除成员
	if err := project.RemoveMember(req.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// 保存更改
	if err := s.projectRepo.RemoveMember(ctx, req.ProjectID, req.UserID); err != nil {
		return fmt.Errorf("failed to save member removal: %w", err)
	}

	// 发布领域事件
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return nil
}

// UpdateProjectMemberRoleRequest 更新项目成员角色请求
type UpdateProjectMemberRoleRequest struct {
	ProjectID uuid.UUID                `json:"project_id" validate:"required"`
	UserID    uuid.UUID                `json:"user_id" validate:"required"`
	Role      domain.ProjectMemberRole `json:"role" validate:"required"`
}

// UpdateProjectMemberRole 更新项目成员角色
func (s *ProjectService) UpdateProjectMemberRole(ctx context.Context, req *UpdateProjectMemberRoleRequest) error {
	// 获取项目
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// 更新成员角色
	if err := project.UpdateMemberRole(req.UserID, req.Role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// 保存更改
	if err := s.projectRepo.UpdateMemberRole(ctx, req.ProjectID, req.UserID, req.Role); err != nil {
		return fmt.Errorf("failed to save member role update: %w", err)
	}

	// 发布领域事件
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return nil
}

// ========== 项目里程碑管�?==========

// AddProjectMilestoneRequest 添加项目里程碑请�?
type AddProjectMilestoneRequest struct {
	ProjectID   uuid.UUID `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=1,max=255"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date" validate:"required"`
}

// AddProjectMilestoneResponse 添加项目里程碑响�?
type AddProjectMilestoneResponse struct {
	Milestone *domain.ProjectMilestone `json:"milestone"`
}

// AddProjectMilestone 添加项目里程�?
func (s *ProjectService) AddProjectMilestone(ctx context.Context, req *AddProjectMilestoneRequest) (*AddProjectMilestoneResponse, error) {
	// 获取项目
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 创建里程�?
	milestone := &domain.ProjectMilestone{
		ID:          uuid.New(),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		DueDate:     req.DueDate,
		Status:      domain.MilestoneStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 添加里程�?
	project.AddMilestone(milestone)

	// 保存里程�?
	if err := s.projectRepo.AddMilestone(ctx, milestone); err != nil {
		return nil, fmt.Errorf("failed to save milestone: %w", err)
	}

	return &AddProjectMilestoneResponse{Milestone: milestone}, nil
}

// CompleteProjectMilestoneRequest 完成项目里程碑请�?
type CompleteProjectMilestoneRequest struct {
	ProjectID   uuid.UUID `json:"project_id" validate:"required"`
	MilestoneID uuid.UUID `json:"milestone_id" validate:"required"`
}

// CompleteProjectMilestone 完成项目里程�?
func (s *ProjectService) CompleteProjectMilestone(ctx context.Context, req *CompleteProjectMilestoneRequest) error {
	// 获取项目
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// 完成里程�?
	if err := project.CompleteMilestone(req.MilestoneID); err != nil {
		return fmt.Errorf("failed to complete milestone: %w", err)
	}

	// 保存更改
	if err := s.projectRepo.UpdateMilestone(ctx, req.MilestoneID, domain.MilestoneStatusCompleted); err != nil {
		return fmt.Errorf("failed to save milestone completion: %w", err)
	}

	// 发布领域事件
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return nil
}

// ========== 项目统计操作 ==========

// GetProjectStatisticsRequest 获取项目统计请求
type GetProjectStatisticsRequest struct {
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	OwnerID        *uuid.UUID `json:"owner_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// GetProjectStatisticsResponse 获取项目统计响应
type GetProjectStatisticsResponse struct {
	Statistics *domain.ProjectStatistics `json:"statistics"`
}

// GetProjectStatistics 获取项目统计
func (s *ProjectService) GetProjectStatistics(ctx context.Context, req *GetProjectStatisticsRequest) (*GetProjectStatisticsResponse, error) {
	stats, err := s.projectRepo.GetProjectStatistics(ctx, req.ProjectID, req.OrganizationID, req.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project statistics: %w", err)
	}

	return &GetProjectStatisticsResponse{Statistics: stats}, nil
}

// ========== 项目调度操作 ==========

// GenerateProjectScheduleRequest 生成项目调度请求
type GenerateProjectScheduleRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// GenerateProjectScheduleResponse 生成项目调度响应
type GenerateProjectScheduleResponse struct {
	Schedule *domain.TaskSchedule `json:"schedule"`
}

// GenerateProjectSchedule 生成项目调度
func (s *ProjectService) GenerateProjectSchedule(ctx context.Context, req *GenerateProjectScheduleRequest) (*GenerateProjectScheduleResponse, error) {
	// 获取项目任务
	tasks, err := s.taskRepo.FindByProject(ctx, req.ProjectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to find project tasks: %w", err)
	}

	// 生成调度
	scheduleReq := &domain.TaskSchedulingRequest{
		ProjectID: &req.ProjectID,
		Tasks:     tasks,
		Strategy:  domain.SchedulingStrategyCriticalPath,
	}

	schedule, err := s.schedulingSvc.GenerateSchedule(ctx, scheduleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schedule: %w", err)
	}

	return &GenerateProjectScheduleResponse{Schedule: schedule}, nil
}

// ========== 项目性能分析 ==========

// GetProjectPerformanceRequest 获取项目性能请求
type GetProjectPerformanceRequest struct {
	ProjectID uuid.UUID  `json:"project_id" validate:"required"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// GetProjectPerformanceResponse 获取项目性能响应
type GetProjectPerformanceResponse struct {
	Performance *domain.ProjectPerformanceReport `json:"performance"`
}

// GetProjectPerformance 获取项目性能
func (s *ProjectService) GetProjectPerformance(ctx context.Context, req *GetProjectPerformanceRequest) (*GetProjectPerformanceResponse, error) {
	// 构建性能分析请求
	perfReq := &domain.ProjectPerformanceRequest{
		ProjectID: req.ProjectID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	// 分析项目性能
	performance, err := s.performanceSvc.AnalyzeProjectPerformance(ctx, perfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project performance: %w", err)
	}

	return &GetProjectPerformanceResponse{Performance: performance}, nil
}
