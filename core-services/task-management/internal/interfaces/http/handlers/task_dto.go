package handlers

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title          string                `json:"title" validate:"required,min=1,max=200"`
	Description    string                `json:"description" validate:"max=2000"`
	ProjectID      uuid.UUID             `json:"project_id" validate:"required"`
	AssigneeID     *uuid.UUID            `json:"assignee_id,omitempty"`
	CreatorID      uuid.UUID             `json:"creator_id" validate:"required"`
	Type           domain.TaskType       `json:"type" validate:"required"`
	Priority       domain.TaskPriority   `json:"priority" validate:"required"`
	Complexity     domain.TaskComplexity `json:"complexity" validate:"required"`
	EstimatedHours float64               `json:"estimated_hours" validate:"min=0"`
	DueDate        *time.Time            `json:"due_date,omitempty"`
	Tags           []string              `json:"tags,omitempty"`
	Labels         map[string]string     `json:"labels,omitempty"`
	Dependencies   []uuid.UUID           `json:"dependencies,omitempty"`
}

// Validate 验证创建任务请求
func (r *CreateTaskRequest) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if len(r.Title) > 200 {
		return errors.New("title too long")
	}
	if len(r.Description) > 2000 {
		return errors.New("description too long")
	}
	if r.ProjectID == uuid.Nil {
		return errors.New("project_id is required")
	}
	if r.CreatorID == uuid.Nil {
		return errors.New("creator_id is required")
	}
	if r.Type == "" {
		return errors.New("type is required")
	}
	if r.Priority == "" {
		return errors.New("priority is required")
	}
	if r.Complexity == "" {
		return errors.New("complexity is required")
	}
	if r.EstimatedHours < 0 {
		return errors.New("estimated_hours must be non-negative")
	}
	return nil
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Title          *string                `json:"title,omitempty"`
	Description    *string                `json:"description,omitempty"`
	AssigneeID     *uuid.UUID             `json:"assignee_id,omitempty"`
	Type           *domain.TaskType       `json:"type,omitempty"`
	Status         *domain.TaskStatus     `json:"status,omitempty"`
	Priority       *domain.TaskPriority   `json:"priority,omitempty"`
	Complexity     *domain.TaskComplexity `json:"complexity,omitempty"`
	EstimatedHours *float64               `json:"estimated_hours,omitempty"`
	ActualHours    *float64               `json:"actual_hours,omitempty"`
	DueDate        *time.Time             `json:"due_date,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Labels         map[string]string      `json:"labels,omitempty"`
}

// AssignTaskRequest 分配任务请求
type AssignTaskRequest struct {
	AssigneeID uuid.UUID `json:"assignee_id" validate:"required"`
	AssignedBy uuid.UUID `json:"assigned_by" validate:"required"`
}

// UnassignTaskRequest 取消分配任务请求
type UnassignTaskRequest struct {
	UnassignedBy uuid.UUID `json:"unassigned_by" validate:"required"`
}

// AutoAssignTasksRequest 自动分配任务请求
type AutoAssignTasksRequest struct {
	ProjectID *uuid.UUID `json:"project_id,omitempty"`
	TeamID    *uuid.UUID `json:"team_id,omitempty"`
	Strategy  string     `json:"strategy" validate:"required"`
}

// AddTaskCommentRequest 添加任务评论请求
type AddTaskCommentRequest struct {
	AuthorID uuid.UUID `json:"author_id" validate:"required"`
	Content  string    `json:"content" validate:"required,min=1,max=2000"`
}

// AddTimeLogRequest 添加时间记录请求
type AddTimeLogRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Duration    float64   `json:"duration" validate:"required,min=0"`
	Description string    `json:"description,omitempty"`
	LoggedAt    time.Time `json:"logged_at" validate:"required"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID             uuid.UUID             `json:"id"`
	Title          string                `json:"title"`
	Description    string                `json:"description"`
	ProjectID      uuid.UUID             `json:"project_id"`
	AssigneeID     *uuid.UUID            `json:"assignee_id,omitempty"`
	CreatorID      uuid.UUID             `json:"creator_id"`
	Type           domain.TaskType       `json:"type"`
	Status         domain.TaskStatus     `json:"status"`
	Priority       domain.TaskPriority   `json:"priority"`
	Complexity     domain.TaskComplexity `json:"complexity"`
	EstimatedHours float64               `json:"estimated_hours"`
	ActualHours    float64               `json:"actual_hours"`
	DueDate        *time.Time            `json:"due_date,omitempty"`
	CompletedAt    *time.Time            `json:"completed_at,omitempty"`
	Tags           []string              `json:"tags,omitempty"`
	Labels         map[string]string     `json:"labels,omitempty"`
	Dependencies   []uuid.UUID           `json:"dependencies,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// ListTasksResponse 任务列表响应
type ListTasksResponse struct {
	Tasks  []TaskResponse `json:"tasks"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// AutoAssignTasksResponse 自动分配任务响应
type AutoAssignTasksResponse struct {
	AssignedTasks int    `json:"assigned_tasks"`
	Summary       string `json:"summary"`
}

// TaskCommentResponse 任务评论响应
type TaskCommentResponse struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"task_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TimeLogResponse 时间记录响应
type TimeLogResponse struct {
	ID          uuid.UUID `json:"id"`
	TaskID      uuid.UUID `json:"task_id"`
	UserID      uuid.UUID `json:"user_id"`
	Duration    float64   `json:"duration"`
	Description string    `json:"description"`
	LoggedAt    time.Time `json:"logged_at"`
	CreatedAt   time.Time `json:"created_at"`
}