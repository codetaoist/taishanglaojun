package postgres

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"task-management/internal/domain"
)

// TaskRepositoryImpl 任务仓储PostgreSQL实现
type TaskRepositoryImpl struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓储实例
func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &TaskRepositoryImpl{db: db}
}

// ========== 数据库模�?==========

// TaskModel 任务数据库模�?
type TaskModel struct {
	ID                uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title             string                   `gorm:"not null;size:255" json:"title"`
	Description       string                   `gorm:"type:text" json:"description"`
	Status            string                   `gorm:"not null;size:50;index" json:"status"`
	Priority          string                   `gorm:"not null;size:50;index" json:"priority"`
	Type              string                   `gorm:"not null;size:50;index" json:"type"`
	Complexity        string                   `gorm:"size:50;index" json:"complexity"`
	
	// 项目和组织信�?
	ProjectID         *uuid.UUID               `gorm:"type:uuid;index" json:"project_id"`
	OrganizationID    *uuid.UUID               `gorm:"type:uuid;index" json:"organization_id"`
	TeamID            *uuid.UUID               `gorm:"type:uuid;index" json:"team_id"`
	
	// 用户信息
	CreatorID         uuid.UUID                `gorm:"type:uuid;not null;index" json:"creator_id"`
	AssigneeID        *uuid.UUID               `gorm:"type:uuid;index" json:"assignee_id"`
	
	// 时间信息
	StartDate         *time.Time               `gorm:"index" json:"start_date"`
	DueDate           *time.Time               `gorm:"index" json:"due_date"`
	CompletedAt       *time.Time               `gorm:"index" json:"completed_at"`
	EstimatedHours    *float64                 `json:"estimated_hours"`
	ActualHours       *float64                 `json:"actual_hours"`
	
	// 标签和元数据
	Tags              StringSlice              `gorm:"type:text[]" json:"tags"`
	Labels            JSONMap                  `gorm:"type:jsonb" json:"labels"`
	Metadata          JSONMap                  `gorm:"type:jsonb" json:"metadata"`
	
	// 进度和质�?
	Progress          float64                  `gorm:"default:0" json:"progress"`
	QualityScore      *float64                 `json:"quality_score"`
	
	// 审计字段
	CreatedAt         time.Time                `gorm:"not null;index" json:"created_at"`
	UpdatedAt         time.Time                `gorm:"not null;index" json:"updated_at"`
	DeletedAt         *time.Time               `gorm:"index" json:"deleted_at"`
	Version           int                      `gorm:"default:1" json:"version"`
	
	// 关联关系
	Dependencies      []TaskDependencyModel    `gorm:"foreignKey:TaskID" json:"dependencies"`
	Dependents        []TaskDependencyModel    `gorm:"foreignKey:DependsOnID" json:"dependents"`
	Assignments       []TaskAssignmentModel    `gorm:"foreignKey:TaskID" json:"assignments"`
	Comments          []TaskCommentModel       `gorm:"foreignKey:TaskID" json:"comments"`
	Attachments       []TaskAttachmentModel    `gorm:"foreignKey:TaskID" json:"attachments"`
	TimeLogs          []TaskTimeLogModel       `gorm:"foreignKey:TaskID" json:"time_logs"`
}

// TaskDependencyModel 任务依赖数据库模�?
type TaskDependencyModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	DependsOnID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"depends_on_id"`
	Type         string     `gorm:"not null;size:50" json:"type"`
	IsBlocking   bool       `gorm:"default:true" json:"is_blocking"`
	CreatedAt    time.Time  `gorm:"not null" json:"created_at"`
	CreatedBy    uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	
	// 关联
	Task         TaskModel  `gorm:"foreignKey:TaskID" json:"task"`
	DependsOn    TaskModel  `gorm:"foreignKey:DependsOnID" json:"depends_on"`
}

// TaskAssignmentModel 任务分配数据库模�?
type TaskAssignmentModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Role         string     `gorm:"not null;size:50" json:"role"`
	AssignedAt   time.Time  `gorm:"not null" json:"assigned_at"`
	AssignedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"assigned_by"`
	UnassignedAt *time.Time `json:"unassigned_at"`
	UnassignedBy *uuid.UUID `gorm:"type:uuid" json:"unassigned_by"`
	IsActive     bool       `gorm:"default:true;index" json:"is_active"`
	
	// 关联
	Task         TaskModel  `gorm:"foreignKey:TaskID" json:"task"`
}

// TaskCommentModel 任务评论数据库模�?
type TaskCommentModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	ParentID     *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	IsInternal   bool       `gorm:"default:false" json:"is_internal"`
	CreatedAt    time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at"`
	
	// 关联
	Task         TaskModel  `gorm:"foreignKey:TaskID" json:"task"`
	Parent       *TaskCommentModel `gorm:"foreignKey:ParentID" json:"parent"`
	Replies      []TaskCommentModel `gorm:"foreignKey:ParentID" json:"replies"`
}

// TaskAttachmentModel 任务附件数据库模�?
type TaskAttachmentModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	FileName     string     `gorm:"not null;size:255" json:"file_name"`
	FileSize     int64      `gorm:"not null" json:"file_size"`
	FileType     string     `gorm:"not null;size:100" json:"file_type"`
	FilePath     string     `gorm:"not null;size:500" json:"file_path"`
	UploadedBy   uuid.UUID  `gorm:"type:uuid;not null" json:"uploaded_by"`
	UploadedAt   time.Time  `gorm:"not null;index" json:"uploaded_at"`
	Description  string     `gorm:"type:text" json:"description"`
	
	// 关联
	Task         TaskModel  `gorm:"foreignKey:TaskID" json:"task"`
}

// TaskTimeLogModel 任务时间记录数据库模�?
type TaskTimeLogModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TaskID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"task_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StartTime    time.Time  `gorm:"not null;index" json:"start_time"`
	EndTime      *time.Time `gorm:"index" json:"end_time"`
	Duration     *int64     `json:"duration"` // �?
	Description  string     `gorm:"type:text" json:"description"`
	IsActive     bool       `gorm:"default:false;index" json:"is_active"`
	CreatedAt    time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updated_at"`
	
	// 关联
	Task         TaskModel  `gorm:"foreignKey:TaskID" json:"task"`
}

// ========== 自定义类�?==========

// StringSlice 字符串切片类�?
type StringSlice []string

// Scan 实现 sql.Scanner 接口
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}
}

// Value 实现 driver.Valuer 接口
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// JSONMap JSON映射类型
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
}

// Value 实现 driver.Valuer 接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// ========== 表名定义 ==========

func (TaskModel) TableName() string { return "tasks" }
func (TaskDependencyModel) TableName() string { return "task_dependencies" }
func (TaskAssignmentModel) TableName() string { return "task_assignments" }
func (TaskCommentModel) TableName() string { return "task_comments" }
func (TaskAttachmentModel) TableName() string { return "task_attachments" }
func (TaskTimeLogModel) TableName() string { return "task_time_logs" }

// ========== 基本CRUD操作 ==========

// Save 保存任务
func (r *TaskRepositoryImpl) Save(ctx context.Context, task *domain.Task) error {
	model := r.domainToModel(task)
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存主任�?
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to save task: %w", err)
		}
		
		// 保存依赖关系
		if len(task.Dependencies) > 0 {
			for _, dep := range task.Dependencies {
				depModel := &TaskDependencyModel{
					TaskID:      task.ID,
					DependsOnID: dep.DependsOnID,
					Type:        dep.Type,
					IsBlocking:  dep.IsBlocking,
					CreatedAt:   dep.CreatedAt,
					CreatedBy:   dep.CreatedBy,
				}
				if err := tx.Create(depModel).Error; err != nil {
					return fmt.Errorf("failed to save task dependency: %w", err)
				}
			}
		}
		
		// 保存分配关系
		if len(task.Assignments) > 0 {
			for _, assignment := range task.Assignments {
				assignModel := &TaskAssignmentModel{
					TaskID:       task.ID,
					UserID:       assignment.UserID,
					Role:         assignment.Role,
					AssignedAt:   assignment.AssignedAt,
					AssignedBy:   assignment.AssignedBy,
					UnassignedAt: assignment.UnassignedAt,
					UnassignedBy: assignment.UnassignedBy,
					IsActive:     assignment.IsActive,
				}
				if err := tx.Create(assignModel).Error; err != nil {
					return fmt.Errorf("failed to save task assignment: %w", err)
				}
			}
		}
		
		return nil
	})
}

// FindByID 根据ID查找任务
func (r *TaskRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	var model TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Dependents").
		Preload("Assignments").
		Preload("Comments").
		Preload("Attachments").
		Preload("TimeLogs").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to find task: %w", err)
	}
	
	return r.modelToDomain(&model), nil
}

// Update 更新任务
func (r *TaskRepositoryImpl) Update(ctx context.Context, task *domain.Task) error {
	model := r.domainToModel(task)
	model.UpdatedAt = time.Now()
	model.Version++
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用乐观锁更�?
		result := tx.Model(&TaskModel{}).
			Where("id = ? AND version = ? AND deleted_at IS NULL", task.ID, task.Version-1).
			Updates(model)
		
		if result.Error != nil {
			return fmt.Errorf("failed to update task: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return domain.ErrTaskVersionConflict
		}
		
		return nil
	})
}

// Delete 删除任务（软删除�?
func (r *TaskRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 软删除任�?
		result := tx.Model(&TaskModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Update("deleted_at", now)
		
		if result.Error != nil {
			return fmt.Errorf("failed to delete task: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return domain.ErrTaskNotFound
		}
		
		// 软删除相关数�?
		tx.Model(&TaskCommentModel{}).Where("task_id = ?", id).Update("deleted_at", now)
		
		return nil
	})
}

// ========== 查询操作 ==========

// FindByProjectID 根据项目ID查找任务
func (r *TaskRepositoryImpl) FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by project: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByAssigneeID 根据分配者ID查找任务
func (r *TaskRepositoryImpl) FindByAssigneeID(ctx context.Context, assigneeID uuid.UUID, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("assignee_id = ? AND deleted_at IS NULL", assigneeID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by assignee: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByCreatorID 根据创建者ID查找任务
func (r *TaskRepositoryImpl) FindByCreatorID(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("creator_id = ? AND deleted_at IS NULL", creatorID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by creator: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByStatus 根据状态查找任�?
func (r *TaskRepositoryImpl) FindByStatus(ctx context.Context, status domain.TaskStatus, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by status: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByPriority 根据优先级查找任�?
func (r *TaskRepositoryImpl) FindByPriority(ctx context.Context, priority domain.TaskPriority, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("priority = ? AND deleted_at IS NULL", string(priority)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by priority: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByType 根据类型查找任务
func (r *TaskRepositoryImpl) FindByType(ctx context.Context, taskType domain.TaskType, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("type = ? AND deleted_at IS NULL", string(taskType)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by type: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// ========== 复合查询 ==========

// FindByProjectAndStatus 根据项目和状态查找任�?
func (r *TaskRepositoryImpl) FindByProjectAndStatus(ctx context.Context, projectID uuid.UUID, status domain.TaskStatus, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("project_id = ? AND status = ? AND deleted_at IS NULL", projectID, string(status)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by project and status: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByAssigneeAndStatus 根据分配者和状态查找任�?
func (r *TaskRepositoryImpl) FindByAssigneeAndStatus(ctx context.Context, assigneeID uuid.UUID, status domain.TaskStatus, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("assignee_id = ? AND status = ? AND deleted_at IS NULL", assigneeID, string(status)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by assignee and status: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByDateRange 根据日期范围查找任务
func (r *TaskRepositoryImpl) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by date range: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindOverdueTasks 查找过期任务
func (r *TaskRepositoryImpl) FindOverdueTasks(ctx context.Context, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	now := time.Now()
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("due_date < ? AND status NOT IN (?, ?) AND deleted_at IS NULL", 
			now, string(domain.TaskStatusCompleted), string(domain.TaskStatusCancelled)).
		Order("due_date ASC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find overdue tasks: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindTasksWithDependencies 查找有依赖关系的任务
func (r *TaskRepositoryImpl) FindTasksWithDependencies(ctx context.Context, taskID uuid.UUID) ([]*domain.Task, error) {
	var models []TaskModel
	
	// 查找依赖于指定任务的任务
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Joins("JOIN task_dependencies ON tasks.id = task_dependencies.task_id").
		Where("task_dependencies.depends_on_id = ? AND tasks.deleted_at IS NULL", taskID).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks with dependencies: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// ========== 搜索操作 ==========

// SearchTasks 搜索任务
func (r *TaskRepositoryImpl) SearchTasks(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	db := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments")
	
	// 文本搜索
	if query != "" {
		searchQuery := "%" + strings.ToLower(query) + "%"
		db = db.Where("(LOWER(title) LIKE ? OR LOWER(description) LIKE ?)", searchQuery, searchQuery)
	}
	
	// 应用过滤�?
	db = r.applyFilters(db, filters)
	
	err := db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to search tasks: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByTags 根据标签查找任务
func (r *TaskRepositoryImpl) FindByTags(ctx context.Context, tags []string, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	err := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments").
		Where("tags && ? AND deleted_at IS NULL", tags).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by tags: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// FindByLabels 根据标签查找任务
func (r *TaskRepositoryImpl) FindByLabels(ctx context.Context, labels map[string]string, limit, offset int) ([]*domain.Task, error) {
	var models []TaskModel
	
	db := r.db.WithContext(ctx).
		Preload("Dependencies").
		Preload("Assignments")
	
	// 构建标签查询条件
	for key, value := range labels {
		db = db.Where("labels ->> ? = ?", key, value)
	}
	
	err := db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by labels: %w", err)
	}
	
	tasks := make([]*domain.Task, len(models))
	for i, model := range models {
		tasks[i] = r.modelToDomain(&model)
	}
	
	return tasks, nil
}

// ========== 统计操作 ==========

// Count 统计任务总数
func (r *TaskRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TaskModel{}).
		Where("deleted_at IS NULL").
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}
	
	return count, nil
}

// CountByProject 根据项目统计任务�?
func (r *TaskRepositoryImpl) CountByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TaskModel{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks by project: %w", err)
	}
	
	return count, nil
}

// CountByAssignee 根据分配者统计任务数
func (r *TaskRepositoryImpl) CountByAssignee(ctx context.Context, assigneeID uuid.UUID) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TaskModel{}).
		Where("assignee_id = ? AND deleted_at IS NULL", assigneeID).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks by assignee: %w", err)
	}
	
	return count, nil
}

// CountByStatus 根据状态统计任务数
func (r *TaskRepositoryImpl) CountByStatus(ctx context.Context, status domain.TaskStatus) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TaskModel{}).
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks by status: %w", err)
	}
	
	return count, nil
}

// GetTaskStatistics 获取任务统计信息
func (r *TaskRepositoryImpl) GetTaskStatistics(ctx context.Context, projectID *uuid.UUID, teamID *uuid.UUID, userID *uuid.UUID) (*domain.TaskStatistics, error) {
	stats := &domain.TaskStatistics{
		TasksByType:       make(map[domain.TaskType]int),
		TasksByPriority:   make(map[domain.TaskPriority]int),
		TasksByComplexity: make(map[domain.TaskComplexity]int),
	}
	
	db := r.db.WithContext(ctx).Model(&TaskModel{}).Where("deleted_at IS NULL")
	
	// 应用过滤条件
	if projectID != nil {
		db = db.Where("project_id = ?", *projectID)
	}
	if teamID != nil {
		db = db.Where("team_id = ?", *teamID)
	}
	if userID != nil {
		db = db.Where("assignee_id = ?", *userID)
	}
	
	// 总任务数
	if err := db.Count(&stats.TotalTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to count total tasks: %w", err)
	}
	
	// 按状态统�?
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := db.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count tasks by status: %w", err)
	}
	
	for _, sc := range statusCounts {
		switch domain.TaskStatus(sc.Status) {
		case domain.TaskStatusCompleted:
			stats.CompletedTasks = sc.Count
		case domain.TaskStatusInProgress:
			stats.InProgressTasks = sc.Count
		case domain.TaskStatusPending:
			stats.PendingTasks = sc.Count
		case domain.TaskStatusCancelled:
			stats.CancelledTasks = sc.Count
		}
	}
	
	// 过期任务�?
	now := time.Now()
	if err := db.Where("due_date < ? AND status NOT IN (?, ?)", 
		now, string(domain.TaskStatusCompleted), string(domain.TaskStatusCancelled)).
		Count(&stats.OverdueTasks).Error; err != nil {
		return nil, fmt.Errorf("failed to count overdue tasks: %w", err)
	}
	
	// 计算完成�?
	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}
	
	// 平均任务时间
	var avgTime sql.NullFloat64
	if err := db.Where("actual_hours IS NOT NULL").
		Select("AVG(actual_hours)").Scan(&avgTime).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average task time: %w", err)
	}
	if avgTime.Valid {
		stats.AverageTaskTime = avgTime.Float64
	}
	
	return stats, nil
}

// ========== 批量操作 ==========

// SaveBatch 批量保存任务
func (r *TaskRepositoryImpl) SaveBatch(ctx context.Context, tasks []*domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	
	models := make([]TaskModel, len(tasks))
	for i, task := range tasks {
		models[i] = *r.domainToModel(task)
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量插入任务
		if err := tx.CreateInBatches(models, 100).Error; err != nil {
			return fmt.Errorf("failed to batch save tasks: %w", err)
		}
		
		// 批量插入依赖关系
		var dependencies []TaskDependencyModel
		for _, task := range tasks {
			for _, dep := range task.Dependencies {
				dependencies = append(dependencies, TaskDependencyModel{
					TaskID:      task.ID,
					DependsOnID: dep.DependsOnID,
					Type:        dep.Type,
					IsBlocking:  dep.IsBlocking,
					CreatedAt:   dep.CreatedAt,
					CreatedBy:   dep.CreatedBy,
				})
			}
		}
		
		if len(dependencies) > 0 {
			if err := tx.CreateInBatches(dependencies, 100).Error; err != nil {
				return fmt.Errorf("failed to batch save task dependencies: %w", err)
			}
		}
		
		return nil
	})
}

// UpdateBatch 批量更新任务
func (r *TaskRepositoryImpl) UpdateBatch(ctx context.Context, tasks []*domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, task := range tasks {
			model := r.domainToModel(task)
			model.UpdatedAt = time.Now()
			model.Version++
			
			result := tx.Model(&TaskModel{}).
				Where("id = ? AND version = ? AND deleted_at IS NULL", task.ID, task.Version-1).
				Updates(model)
			
			if result.Error != nil {
				return fmt.Errorf("failed to update task %s: %w", task.ID, result.Error)
			}
			
			if result.RowsAffected == 0 {
				return fmt.Errorf("task %s version conflict or not found", task.ID)
			}
		}
		
		return nil
	})
}

// DeleteBatch 批量删除任务
func (r *TaskRepositoryImpl) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	
	now := time.Now()
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量软删除任�?
		result := tx.Model(&TaskModel{}).
			Where("id IN ? AND deleted_at IS NULL", ids).
			Update("deleted_at", now)
		
		if result.Error != nil {
			return fmt.Errorf("failed to batch delete tasks: %w", result.Error)
		}
		
		// 软删除相关评�?
		tx.Model(&TaskCommentModel{}).
			Where("task_id IN ?", ids).
			Update("deleted_at", now)
		
		return nil
	})
}

// ========== 辅助方法 ==========

// applyFilters 应用查询过滤�?
func (r *TaskRepositoryImpl) applyFilters(db *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		switch key {
		case "project_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("project_id = ?", v)
			}
		case "assignee_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("assignee_id = ?", v)
			}
		case "creator_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("creator_id = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok {
				db = db.Where("status = ?", v)
			}
		case "priority":
			if v, ok := value.(string); ok {
				db = db.Where("priority = ?", v)
			}
		case "type":
			if v, ok := value.(string); ok {
				db = db.Where("type = ?", v)
			}
		case "start_date":
			if v, ok := value.(time.Time); ok {
				db = db.Where("start_date >= ?", v)
			}
		case "end_date":
			if v, ok := value.(time.Time); ok {
				db = db.Where("due_date <= ?", v)
			}
		case "is_overdue":
			if v, ok := value.(bool); ok && v {
				now := time.Now()
				db = db.Where("due_date < ? AND status NOT IN (?, ?)", 
					now, string(domain.TaskStatusCompleted), string(domain.TaskStatusCancelled))
			}
		}
	}
	
	return db
}

// domainToModel 将领域模型转换为数据库模�?
func (r *TaskRepositoryImpl) domainToModel(task *domain.Task) *TaskModel {
	model := &TaskModel{
		ID:             task.ID,
		Title:          task.Title,
		Description:    task.Description,
		Status:         string(task.Status),
		Priority:       string(task.Priority),
		Type:           string(task.Type),
		Complexity:     string(task.Complexity),
		ProjectID:      task.ProjectID,
		OrganizationID: task.OrganizationID,
		TeamID:         task.TeamID,
		CreatorID:      task.CreatorID,
		AssigneeID:     task.AssigneeID,
		StartDate:      task.StartDate,
		DueDate:        task.DueDate,
		CompletedAt:    task.CompletedAt,
		EstimatedHours: task.EstimatedHours,
		ActualHours:    task.ActualHours,
		Tags:           StringSlice(task.Tags),
		Labels:         JSONMap(task.Labels),
		Metadata:       JSONMap(task.Metadata),
		Progress:       task.Progress,
		QualityScore:   task.QualityScore,
		CreatedAt:      task.CreatedAt,
		UpdatedAt:      task.UpdatedAt,
		Version:        task.Version,
	}
	
	return model
}

// modelToDomain 将数据库模型转换为领域模�?
func (r *TaskRepositoryImpl) modelToDomain(model *TaskModel) *domain.Task {
	task := &domain.Task{
		ID:             model.ID,
		Title:          model.Title,
		Description:    model.Description,
		Status:         domain.TaskStatus(model.Status),
		Priority:       domain.TaskPriority(model.Priority),
		Type:           domain.TaskType(model.Type),
		Complexity:     domain.TaskComplexity(model.Complexity),
		ProjectID:      model.ProjectID,
		OrganizationID: model.OrganizationID,
		TeamID:         model.TeamID,
		CreatorID:      model.CreatorID,
		AssigneeID:     model.AssigneeID,
		StartDate:      model.StartDate,
		DueDate:        model.DueDate,
		CompletedAt:    model.CompletedAt,
		EstimatedHours: model.EstimatedHours,
		ActualHours:    model.ActualHours,
		Tags:           []string(model.Tags),
		Labels:         map[string]string(model.Labels),
		Metadata:       map[string]interface{}(model.Metadata),
		Progress:       model.Progress,
		QualityScore:   model.QualityScore,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		Version:        model.Version,
	}
	
	// 转换依赖关系
	for _, dep := range model.Dependencies {
		task.Dependencies = append(task.Dependencies, &domain.TaskDependency{
			ID:          dep.ID,
			TaskID:      dep.TaskID,
			DependsOnID: dep.DependsOnID,
			Type:        dep.Type,
			IsBlocking:  dep.IsBlocking,
			CreatedAt:   dep.CreatedAt,
			CreatedBy:   dep.CreatedBy,
		})
	}
	
	// 转换分配关系
	for _, assignment := range model.Assignments {
		task.Assignments = append(task.Assignments, &domain.TaskAssignment{
			ID:           assignment.ID,
			TaskID:       assignment.TaskID,
			UserID:       assignment.UserID,
			Role:         assignment.Role,
			AssignedAt:   assignment.AssignedAt,
			AssignedBy:   assignment.AssignedBy,
			UnassignedAt: assignment.UnassignedAt,
			UnassignedBy: assignment.UnassignedBy,
			IsActive:     assignment.IsActive,
		})
	}
	
	// 转换评论
	for _, comment := range model.Comments {
		task.Comments = append(task.Comments, &domain.TaskComment{
			ID:         comment.ID,
			TaskID:     comment.TaskID,
			UserID:     comment.UserID,
			Content:    comment.Content,
			ParentID:   comment.ParentID,
			IsInternal: comment.IsInternal,
			CreatedAt:  comment.CreatedAt,
			UpdatedAt:  comment.UpdatedAt,
		})
	}
	
	// 转换附件
	for _, attachment := range model.Attachments {
		task.Attachments = append(task.Attachments, &domain.TaskAttachment{
			ID:          attachment.ID,
			TaskID:      attachment.TaskID,
			FileName:    attachment.FileName,
			FileSize:    attachment.FileSize,
			FileType:    attachment.FileType,
			FilePath:    attachment.FilePath,
			UploadedBy:  attachment.UploadedBy,
			UploadedAt:  attachment.UploadedAt,
			Description: attachment.Description,
		})
	}
	
	// 转换时间记录
	for _, timeLog := range model.TimeLogs {
		task.TimeLogs = append(task.TimeLogs, &domain.TaskTimeLog{
			ID:          timeLog.ID,
			TaskID:      timeLog.TaskID,
			UserID:      timeLog.UserID,
			StartTime:   timeLog.StartTime,
			EndTime:     timeLog.EndTime,
			Duration:    timeLog.Duration,
			Description: timeLog.Description,
			IsActive:    timeLog.IsActive,
			CreatedAt:   timeLog.CreatedAt,
			UpdatedAt:   timeLog.UpdatedAt,
		})
	}
	
	return task
}

// ========== 依赖关系操作 ==========

// FindDependencies 查找任务依赖
func (r *TaskRepositoryImpl) FindDependencies(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskDependency, error) {
	var models []TaskDependencyModel
	
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find task dependencies: %w", err)
	}
	
	dependencies := make([]*domain.TaskDependency, len(models))
	for i, model := range models {
		dependencies[i] = &domain.TaskDependency{
			ID:          model.ID,
			TaskID:      model.TaskID,
			DependsOnID: model.DependsOnID,
			Type:        model.Type,
			IsBlocking:  model.IsBlocking,
			CreatedAt:   model.CreatedAt,
			CreatedBy:   model.CreatedBy,
		}
	}
	
	return dependencies, nil
}

// FindDependents 查找依赖于指定任务的任务
func (r *TaskRepositoryImpl) FindDependents(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskDependency, error) {
	var models []TaskDependencyModel
	
	err := r.db.WithContext(ctx).
		Where("depends_on_id = ?", taskID).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find task dependents: %w", err)
	}
	
	dependents := make([]*domain.TaskDependency, len(models))
	for i, model := range models {
		dependents[i] = &domain.TaskDependency{
			ID:          model.ID,
			TaskID:      model.TaskID,
			DependsOnID: model.DependsOnID,
			Type:        model.Type,
			IsBlocking:  model.IsBlocking,
			CreatedAt:   model.CreatedAt,
			CreatedBy:   model.CreatedBy,
		}
	}
	
	return dependents, nil
}

// AddDependency 添加任务依赖
func (r *TaskRepositoryImpl) AddDependency(ctx context.Context, dependency *domain.TaskDependency) error {
	model := &TaskDependencyModel{
		TaskID:      dependency.TaskID,
		DependsOnID: dependency.DependsOnID,
		Type:        dependency.Type,
		IsBlocking:  dependency.IsBlocking,
		CreatedAt:   dependency.CreatedAt,
		CreatedBy:   dependency.CreatedBy,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add task dependency: %w", err)
	}
	
	dependency.ID = model.ID
	return nil
}

// RemoveDependency 移除任务依赖
func (r *TaskRepositoryImpl) RemoveDependency(ctx context.Context, taskID, dependsOnID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("task_id = ? AND depends_on_id = ?", taskID, dependsOnID).
		Delete(&TaskDependencyModel{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to remove task dependency: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTaskDependencyNotFound
	}
	
	return nil
}

// ========== 评论操作 ==========

// FindComments 查找任务评论
func (r *TaskRepositoryImpl) FindComments(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]*domain.TaskComment, error) {
	var models []TaskCommentModel
	
	err := r.db.WithContext(ctx).
		Where("task_id = ? AND deleted_at IS NULL", taskID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find task comments: %w", err)
	}
	
	comments := make([]*domain.TaskComment, len(models))
	for i, model := range models {
		comments[i] = &domain.TaskComment{
			ID:         model.ID,
			TaskID:     model.TaskID,
			UserID:     model.UserID,
			Content:    model.Content,
			ParentID:   model.ParentID,
			IsInternal: model.IsInternal,
			CreatedAt:  model.CreatedAt,
			UpdatedAt:  model.UpdatedAt,
		}
	}
	
	return comments, nil
}

// AddComment 添加任务评论
func (r *TaskRepositoryImpl) AddComment(ctx context.Context, comment *domain.TaskComment) error {
	model := &TaskCommentModel{
		TaskID:     comment.TaskID,
		UserID:     comment.UserID,
		Content:    comment.Content,
		ParentID:   comment.ParentID,
		IsInternal: comment.IsInternal,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add task comment: %w", err)
	}
	
	comment.ID = model.ID
	return nil
}

// UpdateComment 更新任务评论
func (r *TaskRepositoryImpl) UpdateComment(ctx context.Context, comment *domain.TaskComment) error {
	result := r.db.WithContext(ctx).
		Model(&TaskCommentModel{}).
		Where("id = ? AND deleted_at IS NULL", comment.ID).
		Updates(map[string]interface{}{
			"content":    comment.Content,
			"updated_at": time.Now(),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update task comment: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTaskCommentNotFound
	}
	
	return nil
}

// DeleteComment 删除任务评论
func (r *TaskRepositoryImpl) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	now := time.Now()
	
	result := r.db.WithContext(ctx).
		Model(&TaskCommentModel{}).
		Where("id = ? AND deleted_at IS NULL", commentID).
		Update("deleted_at", now)
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete task comment: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTaskCommentNotFound
	}
	
	return nil
}

// ========== 附件操作 ==========

// FindAttachments 查找任务附件
func (r *TaskRepositoryImpl) FindAttachments(ctx context.Context, taskID uuid.UUID) ([]*domain.TaskAttachment, error) {
	var models []TaskAttachmentModel
	
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("uploaded_at DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find task attachments: %w", err)
	}
	
	attachments := make([]*domain.TaskAttachment, len(models))
	for i, model := range models {
		attachments[i] = &domain.TaskAttachment{
			ID:          model.ID,
			TaskID:      model.TaskID,
			FileName:    model.FileName,
			FileSize:    model.FileSize,
			FileType:    model.FileType,
			FilePath:    model.FilePath,
			UploadedBy:  model.UploadedBy,
			UploadedAt:  model.UploadedAt,
			Description: model.Description,
		}
	}
	
	return attachments, nil
}

// AddAttachment 添加任务附件
func (r *TaskRepositoryImpl) AddAttachment(ctx context.Context, attachment *domain.TaskAttachment) error {
	model := &TaskAttachmentModel{
		TaskID:      attachment.TaskID,
		FileName:    attachment.FileName,
		FileSize:    attachment.FileSize,
		FileType:    attachment.FileType,
		FilePath:    attachment.FilePath,
		UploadedBy:  attachment.UploadedBy,
		UploadedAt:  attachment.UploadedAt,
		Description: attachment.Description,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add task attachment: %w", err)
	}
	
	attachment.ID = model.ID
	return nil
}

// DeleteAttachment 删除任务附件
func (r *TaskRepositoryImpl) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", attachmentID).
		Delete(&TaskAttachmentModel{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete task attachment: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTaskAttachmentNotFound
	}
	
	return nil
}

// ========== 时间记录操作 ==========

// FindTimeLogs 查找任务时间记录
func (r *TaskRepositoryImpl) FindTimeLogs(ctx context.Context, taskID uuid.UUID, limit, offset int) ([]*domain.TaskTimeLog, error) {
	var models []TaskTimeLogModel
	
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("start_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find task time logs: %w", err)
	}
	
	timeLogs := make([]*domain.TaskTimeLog, len(models))
	for i, model := range models {
		timeLogs[i] = &domain.TaskTimeLog{
			ID:          model.ID,
			TaskID:      model.TaskID,
			UserID:      model.UserID,
			StartTime:   model.StartTime,
			EndTime:     model.EndTime,
			Duration:    model.Duration,
			Description: model.Description,
			IsActive:    model.IsActive,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
		}
	}
	
	return timeLogs, nil
}

// AddTimeLog 添加任务时间记录
func (r *TaskRepositoryImpl) AddTimeLog(ctx context.Context, timeLog *domain.TaskTimeLog) error {
	model := &TaskTimeLogModel{
		TaskID:      timeLog.TaskID,
		UserID:      timeLog.UserID,
		StartTime:   timeLog.StartTime,
		EndTime:     timeLog.EndTime,
		Duration:    timeLog.Duration,
		Description: timeLog.Description,
		IsActive:    timeLog.IsActive,
		CreatedAt:   timeLog.CreatedAt,
		UpdatedAt:   timeLog.UpdatedAt,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add task time log: %w", err)
	}
	
	timeLog.ID = model.ID
	return nil
}

// UpdateTimeLog 更新任务时间记录
func (r *TaskRepositoryImpl) UpdateTimeLog(ctx context.Context, timeLog *domain.TaskTimeLog) error {
	result := r.db.WithContext(ctx).
		Model(&TaskTimeLogModel{}).
		Where("id = ?", timeLog.ID).
		Updates(map[string]interface{}{
			"end_time":    timeLog.EndTime,
			"duration":    timeLog.Duration,
			"description": timeLog.Description,
			"is_active":   timeLog.IsActive,
			"updated_at":  time.Now(),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update task time log: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTaskTimeLogNotFound
	}
	
	return nil
}

// DeleteTimeLog 删除任务时间记录
func (r *TaskRepositoryImpl) DeleteTimeLog(ctx context.Context, timeLogID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", timeLogID).
		Delete(&TaskTimeLogModel{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete task time log: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTaskTimeLogNotFound
	}
	
	return nil
}

// GetTimeLogStatistics 获取时间记录统计
func (r *TaskRepositoryImpl) GetTimeLogStatistics(ctx context.Context, taskID uuid.UUID) (*domain.TimeLogStatistics, error) {
	stats := &domain.TimeLogStatistics{}
	
	// 总时长和会话�?
	var result struct {
		TotalDuration int64
		TotalSessions int64
	}
	
	err := r.db.WithContext(ctx).
		Model(&TaskTimeLogModel{}).
		Where("task_id = ? AND duration IS NOT NULL", taskID).
		Select("COALESCE(SUM(duration), 0) as total_duration, COUNT(*) as total_sessions").
		Scan(&result).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get time log statistics: %w", err)
	}
	
	stats.TotalDuration = result.TotalDuration
	stats.TotalSessions = int(result.TotalSessions)
	
	// 平均会话时长
	if stats.TotalSessions > 0 {
		stats.AverageSessionDuration = float64(stats.TotalDuration) / float64(stats.TotalSessions)
	}
	
	// 最长和最短会�?
	var minMax struct {
		MinDuration sql.NullInt64
		MaxDuration sql.NullInt64
	}
	
	err = r.db.WithContext(ctx).
		Model(&TaskTimeLogModel{}).
		Where("task_id = ? AND duration IS NOT NULL", taskID).
		Select("MIN(duration) as min_duration, MAX(duration) as max_duration").
		Scan(&minMax).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get min/max duration: %w", err)
	}
	
	if minMax.MinDuration.Valid {
		stats.MinSessionDuration = minMax.MinDuration.Int64
	}
	if minMax.MaxDuration.Valid {
		stats.MaxSessionDuration = minMax.MaxDuration.Int64
	}
	
	return stats, nil
}
