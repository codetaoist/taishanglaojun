package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// ProjectService 
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

// NewProjectService 
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

// ========== CRUD ==========

// CreateProjectRequest 
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

// CreateProjectResponse 
type CreateProjectResponse struct {
	Project *domain.Project `json:"project"`
}

// CreateProject 
func (s *ProjectService) CreateProject(ctx context.Context, req *CreateProjectRequest) (*CreateProjectResponse, error) {
	// 
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

	// ?
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

	// 
	if err := s.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	// 
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 
	notification := &domain.ProjectNotification{
		Type:      domain.NotificationTypeProjectCreated,
		ProjectID: project.ID,
		UserID:    req.OwnerID,
		Title:     "",
		Message:   fmt.Sprintf(" %s ?, project.Name),
		Metadata:  map[string]interface{}{"project_id": project.ID.String()},
	}
	if err := s.notificationSvc.SendProjectNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	project.ClearDomainEvents()

	return &CreateProjectResponse{Project: project}, nil
}

// GetProjectRequest 
type GetProjectRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// GetProjectResponse 
type GetProjectResponse struct {
	Project *domain.Project `json:"project"`
}

// GetProject 
func (s *ProjectService) GetProject(ctx context.Context, req *GetProjectRequest) (*GetProjectResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	return &GetProjectResponse{Project: project}, nil
}

// UpdateProjectRequest 
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

// UpdateProjectResponse 
type UpdateProjectResponse struct {
	Project *domain.Project `json:"project"`
}

// UpdateProject 
func (s *ProjectService) UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*UpdateProjectResponse, error) {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 
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

	// 
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// 
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return &UpdateProjectResponse{Project: project}, nil
}

// DeleteProjectRequest 
type DeleteProjectRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// DeleteProject 
func (s *ProjectService) DeleteProject(ctx context.Context, req *DeleteProjectRequest) error {
	// ?
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// ?
	if project.Status == domain.ProjectStatusActive {
		return fmt.Errorf("cannot delete active project")
	}

	// ?
	taskCount, err := s.taskRepo.CountByProject(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to count project tasks: %w", err)
	}
	if taskCount > 0 {
		return fmt.Errorf("cannot delete project with existing tasks")
	}

	// 
	if err := s.projectRepo.Delete(ctx, req.ProjectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ==========  ==========

// ListProjectsRequest 
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

// ListProjectsResponse 
type ListProjectsResponse struct {
	Projects []*domain.Project `json:"projects"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasMore  bool              `json:"has_more"`
}

// ListProjects 
func (s *ProjectService) ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
	// 
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

	// ?
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

	// 
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

	// 
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

// SearchProjectsRequest 
type SearchProjectsRequest struct {
	Query   string                 `json:"query" validate:"required"`
	Filters map[string]interface{} `json:"filters,omitempty"`
	Limit   int                    `json:"limit" validate:"min=1,max=100"`
	Offset  int                    `json:"offset" validate:"min=0"`
}

// SearchProjectsResponse 
type SearchProjectsResponse struct {
	Projects []*domain.Project `json:"projects"`
	Total    int64             `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasMore  bool              `json:"has_more"`
}

// SearchProjects 
func (s *ProjectService) SearchProjects(ctx context.Context, req *SearchProjectsRequest) (*SearchProjectsResponse, error) {
	projects, err := s.projectRepo.SearchProjects(ctx, req.Query, req.Filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	// ?
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

// ==========  ==========

// AddProjectMemberRequest 
type AddProjectMemberRequest struct {
	ProjectID uuid.UUID                  `json:"project_id" validate:"required"`
	UserID    uuid.UUID                  `json:"user_id" validate:"required"`
	Role      domain.ProjectMemberRole   `json:"role" validate:"required"`
}

// AddProjectMemberResponse 
type AddProjectMemberResponse struct {
	Member *domain.ProjectMember `json:"member"`
}

// AddProjectMember 
func (s *ProjectService) AddProjectMember(ctx context.Context, req *AddProjectMemberRequest) (*AddProjectMemberResponse, error) {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// 
	member := &domain.ProjectMember{
		ID:        uuid.New(),
		ProjectID: req.ProjectID,
		UserID:    req.UserID,
		Role:      req.Role,
		JoinedAt:  time.Now(),
	}

	// 
	if err := project.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// 
	if err := s.projectRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to save member: %w", err)
	}

	// 
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 
	notification := &domain.ProjectNotification{
		Type:      domain.NotificationTypeProjectMemberAdded,
		ProjectID: project.ID,
		UserID:    req.UserID,
		Title:     "?,
		Message:   fmt.Sprintf("? %s", project.Name),
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

// RemoveProjectMemberRequest 
type RemoveProjectMemberRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
}

// RemoveProjectMember 
func (s *ProjectService) RemoveProjectMember(ctx context.Context, req *RemoveProjectMemberRequest) error {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// 
	if err := project.RemoveMember(req.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// 
	if err := s.projectRepo.RemoveMember(ctx, req.ProjectID, req.UserID); err != nil {
		return fmt.Errorf("failed to save member removal: %w", err)
	}

	// 
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return nil
}

// UpdateProjectMemberRoleRequest 
type UpdateProjectMemberRoleRequest struct {
	ProjectID uuid.UUID                `json:"project_id" validate:"required"`
	UserID    uuid.UUID                `json:"user_id" validate:"required"`
	Role      domain.ProjectMemberRole `json:"role" validate:"required"`
}

// UpdateProjectMemberRole 
func (s *ProjectService) UpdateProjectMemberRole(ctx context.Context, req *UpdateProjectMemberRoleRequest) error {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// 
	if err := project.UpdateMemberRole(req.UserID, req.Role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	// 
	if err := s.projectRepo.UpdateMemberRole(ctx, req.ProjectID, req.UserID, req.Role); err != nil {
		return fmt.Errorf("failed to save member role update: %w", err)
	}

	// 
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return nil
}

// ========== ?==========

// AddProjectMilestoneRequest ?
type AddProjectMilestoneRequest struct {
	ProjectID   uuid.UUID `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required,min=1,max=255"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date" validate:"required"`
}

// AddProjectMilestoneResponse ?
type AddProjectMilestoneResponse struct {
	Milestone *domain.ProjectMilestone `json:"milestone"`
}

// AddProjectMilestone ?
func (s *ProjectService) AddProjectMilestone(ctx context.Context, req *AddProjectMilestoneRequest) (*AddProjectMilestoneResponse, error) {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// ?
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

	// ?
	project.AddMilestone(milestone)

	// ?
	if err := s.projectRepo.AddMilestone(ctx, milestone); err != nil {
		return nil, fmt.Errorf("failed to save milestone: %w", err)
	}

	return &AddProjectMilestoneResponse{Milestone: milestone}, nil
}

// CompleteProjectMilestoneRequest ?
type CompleteProjectMilestoneRequest struct {
	ProjectID   uuid.UUID `json:"project_id" validate:"required"`
	MilestoneID uuid.UUID `json:"milestone_id" validate:"required"`
}

// CompleteProjectMilestone ?
func (s *ProjectService) CompleteProjectMilestone(ctx context.Context, req *CompleteProjectMilestoneRequest) error {
	// 
	project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}

	// ?
	if err := project.CompleteMilestone(req.MilestoneID); err != nil {
		return fmt.Errorf("failed to complete milestone: %w", err)
	}

	// 
	if err := s.projectRepo.UpdateMilestone(ctx, req.MilestoneID, domain.MilestoneStatusCompleted); err != nil {
		return fmt.Errorf("failed to save milestone completion: %w", err)
	}

	// 
	for _, event := range project.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	project.ClearDomainEvents()

	return nil
}

// ==========  ==========

// GetProjectStatisticsRequest 
type GetProjectStatisticsRequest struct {
	ProjectID      *uuid.UUID `json:"project_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	OwnerID        *uuid.UUID `json:"owner_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// GetProjectStatisticsResponse 
type GetProjectStatisticsResponse struct {
	Statistics *domain.ProjectStatistics `json:"statistics"`
}

// GetProjectStatistics 
func (s *ProjectService) GetProjectStatistics(ctx context.Context, req *GetProjectStatisticsRequest) (*GetProjectStatisticsResponse, error) {
	stats, err := s.projectRepo.GetProjectStatistics(ctx, req.ProjectID, req.OrganizationID, req.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project statistics: %w", err)
	}

	return &GetProjectStatisticsResponse{Statistics: stats}, nil
}

// ==========  ==========

// GenerateProjectScheduleRequest 
type GenerateProjectScheduleRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// GenerateProjectScheduleResponse 
type GenerateProjectScheduleResponse struct {
	Schedule *domain.TaskSchedule `json:"schedule"`
}

// GenerateProjectSchedule 
func (s *ProjectService) GenerateProjectSchedule(ctx context.Context, req *GenerateProjectScheduleRequest) (*GenerateProjectScheduleResponse, error) {
	// 
	tasks, err := s.taskRepo.FindByProject(ctx, req.ProjectID, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to find project tasks: %w", err)
	}

	// 
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

// ==========  ==========

// GetProjectPerformanceRequest 
type GetProjectPerformanceRequest struct {
	ProjectID uuid.UUID  `json:"project_id" validate:"required"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// GetProjectPerformanceResponse 
type GetProjectPerformanceResponse struct {
	Performance *domain.ProjectPerformanceReport `json:"performance"`
}

// GetProjectPerformance 
func (s *ProjectService) GetProjectPerformance(ctx context.Context, req *GetProjectPerformanceRequest) (*GetProjectPerformanceResponse, error) {
	// 
	perfReq := &domain.ProjectPerformanceRequest{
		ProjectID: req.ProjectID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	// 
	performance, err := s.performanceSvc.AnalyzeProjectPerformance(ctx, perfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project performance: %w", err)
	}

	return &GetProjectPerformanceResponse{Performance: performance}, nil
}

