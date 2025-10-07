package handlers

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name           string                 `json:"name" validate:"required,min=1,max=200"`
	Description    string                 `json:"description" validate:"max=2000"`
	OrganizationID uuid.UUID              `json:"organization_id" validate:"required"`
	ManagerID      uuid.UUID              `json:"manager_id" validate:"required"`
	Status         domain.ProjectStatus   `json:"status" validate:"required"`
	Priority       domain.ProjectPriority `json:"priority" validate:"required"`
	StartDate      time.Time              `json:"start_date" validate:"required"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	Budget         *float64               `json:"budget,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Labels         map[string]string      `json:"labels,omitempty"`
}

// Validate 验证创建项目请求
func (r *CreateProjectRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 200 {
		return errors.New("name too long")
	}
	if len(r.Description) > 2000 {
		return errors.New("description too long")
	}
	if r.OrganizationID == uuid.Nil {
		return errors.New("organization_id is required")
	}
	if r.ManagerID == uuid.Nil {
		return errors.New("manager_id is required")
	}
	if r.Status == "" {
		return errors.New("status is required")
	}
	if r.Priority == "" {
		return errors.New("priority is required")
	}
	if r.StartDate.IsZero() {
		return errors.New("start_date is required")
	}
	if r.EndDate != nil && r.EndDate.Before(r.StartDate) {
		return errors.New("end_date must be after start_date")
	}
	if r.Budget != nil && *r.Budget < 0 {
		return errors.New("budget must be non-negative")
	}
	return nil
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	ManagerID   *uuid.UUID              `json:"manager_id,omitempty"`
	Status      *domain.ProjectStatus   `json:"status,omitempty"`
	Priority    *domain.ProjectPriority `json:"priority,omitempty"`
	StartDate   *time.Time              `json:"start_date,omitempty"`
	EndDate     *time.Time              `json:"end_date,omitempty"`
	Budget      *float64                `json:"budget,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	Labels      map[string]string       `json:"labels,omitempty"`
}

// AddProjectMemberRequest 添加项目成员请求
type AddProjectMemberRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Role    string    `json:"role" validate:"required"`
	AddedBy uuid.UUID `json:"added_by" validate:"required"`
}

// RemoveProjectMemberRequest 移除项目成员请求
type RemoveProjectMemberRequest struct {
	RemovedBy uuid.UUID `json:"removed_by" validate:"required"`
}

// UpdateProjectMemberRoleRequest 更新项目成员角色请求
type UpdateProjectMemberRoleRequest struct {
	Role      string    `json:"role" validate:"required"`
	UpdatedBy uuid.UUID `json:"updated_by" validate:"required"`
}

// AddProjectMilestoneRequest 添加项目里程碑请求
type AddProjectMilestoneRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=200"`
	Description string     `json:"description,omitempty"`
	DueDate     time.Time  `json:"due_date" validate:"required"`
	CreatedBy   uuid.UUID  `json:"created_by" validate:"required"`
}

// CompleteProjectMilestoneRequest 完成项目里程碑请求
type CompleteProjectMilestoneRequest struct {
	CompletedBy uuid.UUID `json:"completed_by" validate:"required"`
}

// ProjectResponse 项目响应
type ProjectResponse struct {
	ID             uuid.UUID              `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	ManagerID      uuid.UUID              `json:"manager_id"`
	Status         domain.ProjectStatus   `json:"status"`
	Priority       domain.ProjectPriority `json:"priority"`
	StartDate      time.Time              `json:"start_date"`
	EndDate        *time.Time             `json:"end_date,omitempty"`
	Budget         *float64               `json:"budget,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Labels         map[string]string      `json:"labels,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ListProjectsResponse 项目列表响应
type ListProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// ProjectMemberResponse 项目成员响应
type ProjectMemberResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ProjectMilestoneResponse 项目里程碑响应
type ProjectMilestoneResponse struct {
	ID          uuid.UUID  `json:"id"`
	ProjectID   uuid.UUID  `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	DueDate     time.Time  `json:"due_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}