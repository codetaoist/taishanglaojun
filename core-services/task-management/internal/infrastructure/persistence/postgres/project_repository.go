package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"task-management/internal/domain"
)

// ProjectRepositoryImpl 项目仓储PostgreSQL实现
type ProjectRepositoryImpl struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓储实例
func NewProjectRepository(db *gorm.DB) domain.ProjectRepository {
	return &ProjectRepositoryImpl{db: db}
}

// ========== 数据库模型 ==========

// ProjectModel 项目数据库模型
type ProjectModel struct {
	ID             uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string                   `gorm:"not null;size:255" json:"name"`
	Description    string                   `gorm:"type:text" json:"description"`
	Status         string                   `gorm:"not null;size:50;index" json:"status"`
	Priority       string                   `gorm:"not null;size:50;index" json:"priority"`
	Type           string                   `gorm:"not null;size:50;index" json:"type"`
	
	// 组织信息
	OrganizationID uuid.UUID                `gorm:"type:uuid;not null;index" json:"organization_id"`
	OwnerID        uuid.UUID                `gorm:"type:uuid;not null;index" json:"owner_id"`
	ManagerID      *uuid.UUID               `gorm:"type:uuid;index" json:"manager_id"`
	
	// 时间信息
	StartDate      *time.Time               `gorm:"index" json:"start_date"`
	EndDate        *time.Time               `gorm:"index" json:"end_date"`
	DueDate        *time.Time               `gorm:"index" json:"due_date"`
	CompletedAt    *time.Time               `gorm:"index" json:"completed_at"`
	
	// 预算信息
	Budget         *float64                 `json:"budget"`
	ActualCost     *float64                 `json:"actual_cost"`
	Currency       string                   `gorm:"size:10" json:"currency"`
	
	// 标签和元数据
	Tags           StringSlice              `gorm:"type:text[]" json:"tags"`
	Labels         JSONMap                  `gorm:"type:jsonb" json:"labels"`
	Metadata       JSONMap                  `gorm:"type:jsonb" json:"metadata"`
	
	// 进度信息
	Progress       float64                  `gorm:"default:0" json:"progress"`
	
	// 审计字段
	CreatedAt      time.Time                `gorm:"not null;index" json:"created_at"`
	UpdatedAt      time.Time                `gorm:"not null;index" json:"updated_at"`
	DeletedAt      *time.Time               `gorm:"index" json:"deleted_at"`
	Version        int                      `gorm:"default:1" json:"version"`
	
	// 关联关系
	Members        []ProjectMemberModel     `gorm:"foreignKey:ProjectID" json:"members"`
	Milestones     []ProjectMilestoneModel  `gorm:"foreignKey:ProjectID" json:"milestones"`
}

// ProjectMemberModel 项目成员数据库模型
type ProjectMemberModel struct {
	ID         uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID    `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	Role       string       `gorm:"not null;size:50" json:"role"`
	JoinedAt   time.Time    `gorm:"not null" json:"joined_at"`
	LeftAt     *time.Time   `json:"left_at"`
	IsActive   bool         `gorm:"default:true;index" json:"is_active"`
	
	// 关联
	Project    ProjectModel `gorm:"foreignKey:ProjectID" json:"project"`
}

// ProjectMilestoneModel 项目里程碑数据库模型
type ProjectMilestoneModel struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID    `gorm:"type:uuid;not null;index" json:"project_id"`
	Name        string       `gorm:"not null;size:255" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	DueDate     *time.Time   `gorm:"index" json:"due_date"`
	CompletedAt *time.Time   `gorm:"index" json:"completed_at"`
	IsCompleted bool         `gorm:"default:false;index" json:"is_completed"`
	CreatedAt   time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updated_at"`
	
	// 关联
	Project     ProjectModel `gorm:"foreignKey:ProjectID" json:"project"`
}

// ========== 表名定义 ==========

func (ProjectModel) TableName() string { return "projects" }
func (ProjectMemberModel) TableName() string { return "project_members" }
func (ProjectMilestoneModel) TableName() string { return "project_milestones" }

// ========== 基本CRUD操作 ==========

// Save 保存项目
func (r *ProjectRepositoryImpl) Save(ctx context.Context, project *domain.Project) error {
	model := r.domainToModel(project)
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存主项目
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to save project: %w", err)
		}
		
		// 保存成员关系
		if len(project.Members) > 0 {
			for _, member := range project.Members {
				memberModel := &ProjectMemberModel{
					ProjectID: project.ID,
					UserID:    member.UserID,
					Role:      member.Role,
					JoinedAt:  member.JoinedAt,
					LeftAt:    member.LeftAt,
					IsActive:  member.IsActive,
				}
				if err := tx.Create(memberModel).Error; err != nil {
					return fmt.Errorf("failed to save project member: %w", err)
				}
			}
		}
		
		// 保存里程碑
		if len(project.Milestones) > 0 {
			for _, milestone := range project.Milestones {
				milestoneModel := &ProjectMilestoneModel{
					ProjectID:   project.ID,
					Name:        milestone.Name,
					Description: milestone.Description,
					DueDate:     milestone.DueDate,
					CompletedAt: milestone.CompletedAt,
					IsCompleted: milestone.IsCompleted,
					CreatedAt:   milestone.CreatedAt,
					UpdatedAt:   milestone.UpdatedAt,
				}
				if err := tx.Create(milestoneModel).Error; err != nil {
					return fmt.Errorf("failed to save project milestone: %w", err)
				}
			}
		}
		
		return nil
	})
}

// FindByID 根据ID查找项目
func (r *ProjectRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	var model ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	
	return r.modelToDomain(&model), nil
}

// Update 更新项目
func (r *ProjectRepositoryImpl) Update(ctx context.Context, project *domain.Project) error {
	model := r.domainToModel(project)
	model.UpdatedAt = time.Now()
	model.Version++
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用乐观锁更新
		result := tx.Model(&ProjectModel{}).
			Where("id = ? AND version = ? AND deleted_at IS NULL", project.ID, project.Version-1).
			Updates(model)
		
		if result.Error != nil {
			return fmt.Errorf("failed to update project: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return domain.ErrProjectVersionConflict
		}
		
		return nil
	})
}

// Delete 删除项目（软删除）
func (r *ProjectRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 软删除项目
		result := tx.Model(&ProjectModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Update("deleted_at", now)
		
		if result.Error != nil {
			return fmt.Errorf("failed to delete project: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return domain.ErrProjectNotFound
		}
		
		return nil
	})
}

// ========== 查询操作 ==========

// FindByOrganizationID 根据组织ID查找项目
func (r *ProjectRepositoryImpl) FindByOrganizationID(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by organization: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByOwnerID 根据所有者ID查找项目
func (r *ProjectRepositoryImpl) FindByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by owner: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByManagerID 根据管理者ID查找项目
func (r *ProjectRepositoryImpl) FindByManagerID(ctx context.Context, managerID uuid.UUID, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("manager_id = ? AND deleted_at IS NULL", managerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by manager: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByStatus 根据状态查找项目
func (r *ProjectRepositoryImpl) FindByStatus(ctx context.Context, status domain.ProjectStatus, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by status: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByPriority 根据优先级查找项目
func (r *ProjectRepositoryImpl) FindByPriority(ctx context.Context, priority domain.ProjectPriority, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("priority = ? AND deleted_at IS NULL", string(priority)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by priority: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByType 根据类型查找项目
func (r *ProjectRepositoryImpl) FindByType(ctx context.Context, projectType domain.ProjectType, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("type = ? AND deleted_at IS NULL", string(projectType)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by type: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// ========== 复合查询 ==========

// FindByOrganizationAndStatus 根据组织和状态查找项目
func (r *ProjectRepositoryImpl) FindByOrganizationAndStatus(ctx context.Context, organizationID uuid.UUID, status domain.ProjectStatus, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("organization_id = ? AND status = ? AND deleted_at IS NULL", organizationID, string(status)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by organization and status: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByDateRange 根据日期范围查找项目
func (r *ProjectRepositoryImpl) FindByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("created_at BETWEEN ? AND ? AND deleted_at IS NULL", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by date range: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindOverdueProjects 查找过期项目
func (r *ProjectRepositoryImpl) FindOverdueProjects(ctx context.Context, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	now := time.Now()
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("due_date < ? AND status NOT IN (?, ?) AND deleted_at IS NULL", 
			now, string(domain.ProjectStatusCompleted), string(domain.ProjectStatusCancelled)).
		Order("due_date ASC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find overdue projects: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// ========== 搜索操作 ==========

// SearchProjects 搜索项目
func (r *ProjectRepositoryImpl) SearchProjects(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	db := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones")
	
	// 文本搜索
	if query != "" {
		searchQuery := "%" + strings.ToLower(query) + "%"
		db = db.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", searchQuery, searchQuery)
	}
	
	// 应用过滤器
	db = r.applyFilters(db, filters)
	
	err := db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByTags 根据标签查找项目
func (r *ProjectRepositoryImpl) FindByTags(ctx context.Context, tags []string, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones").
		Where("tags && ? AND deleted_at IS NULL", tags).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by tags: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// FindByLabels 根据标签查找项目
func (r *ProjectRepositoryImpl) FindByLabels(ctx context.Context, labels map[string]string, limit, offset int) ([]*domain.Project, error) {
	var models []ProjectModel
	
	db := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Milestones")
	
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
		return nil, fmt.Errorf("failed to find projects by labels: %w", err)
	}
	
	projects := make([]*domain.Project, len(models))
	for i, model := range models {
		projects[i] = r.modelToDomain(&model)
	}
	
	return projects, nil
}

// ========== 统计操作 ==========

// Count 统计项目总数
func (r *ProjectRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&ProjectModel{}).
		Where("deleted_at IS NULL").
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}
	
	return count, nil
}

// CountByOrganization 根据组织统计项目数
func (r *ProjectRepositoryImpl) CountByOrganization(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&ProjectModel{}).
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count projects by organization: %w", err)
	}
	
	return count, nil
}

// CountByOwner 根据所有者统计项目数
func (r *ProjectRepositoryImpl) CountByOwner(ctx context.Context, ownerID uuid.UUID) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&ProjectModel{}).
		Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count projects by owner: %w", err)
	}
	
	return count, nil
}

// CountByStatus 根据状态统计项目数
func (r *ProjectRepositoryImpl) CountByStatus(ctx context.Context, status domain.ProjectStatus) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&ProjectModel{}).
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count projects by status: %w", err)
	}
	
	return count, nil
}

// GetProjectStatistics 获取项目统计信息
func (r *ProjectRepositoryImpl) GetProjectStatistics(ctx context.Context, organizationID *uuid.UUID, ownerID *uuid.UUID) (*domain.ProjectStatistics, error) {
	stats := &domain.ProjectStatistics{
		ProjectsByType:     make(map[domain.ProjectType]int),
		ProjectsByPriority: make(map[domain.ProjectPriority]int),
	}
	
	db := r.db.WithContext(ctx).Model(&ProjectModel{}).Where("deleted_at IS NULL")
	
	// 应用过滤条件
	if organizationID != nil {
		db = db.Where("organization_id = ?", *organizationID)
	}
	if ownerID != nil {
		db = db.Where("owner_id = ?", *ownerID)
	}
	
	// 总项目数
	if err := db.Count(&stats.TotalProjects).Error; err != nil {
		return nil, fmt.Errorf("failed to count total projects: %w", err)
	}
	
	// 按状态统计
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := db.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count projects by status: %w", err)
	}
	
	for _, sc := range statusCounts {
		switch domain.ProjectStatus(sc.Status) {
		case domain.ProjectStatusCompleted:
			stats.CompletedProjects = sc.Count
		case domain.ProjectStatusInProgress:
			stats.InProgressProjects = sc.Count
		case domain.ProjectStatusPlanning:
			stats.PlanningProjects = sc.Count
		case domain.ProjectStatusOnHold:
			stats.OnHoldProjects = sc.Count
		case domain.ProjectStatusCancelled:
			stats.CancelledProjects = sc.Count
		}
	}
	
	// 过期项目数
	now := time.Now()
	if err := db.Where("due_date < ? AND status NOT IN (?, ?)", 
		now, string(domain.ProjectStatusCompleted), string(domain.ProjectStatusCancelled)).
		Count(&stats.OverdueProjects).Error; err != nil {
		return nil, fmt.Errorf("failed to count overdue projects: %w", err)
	}
	
	// 计算完成率
	if stats.TotalProjects > 0 {
		stats.CompletionRate = float64(stats.CompletedProjects) / float64(stats.TotalProjects) * 100
	}
	
	// 平均项目预算
	var avgBudget sql.NullFloat64
	if err := db.Where("budget IS NOT NULL").
		Select("AVG(budget)").Scan(&avgBudget).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average project budget: %w", err)
	}
	if avgBudget.Valid {
		stats.AverageBudget = avgBudget.Float64
	}
	
	// 总预算和实际成本
	var budgetSum, costSum sql.NullFloat64
	if err := db.Select("SUM(budget) as budget_sum, SUM(actual_cost) as cost_sum").
		Scan(&struct {
			BudgetSum sql.NullFloat64 `json:"budget_sum"`
			CostSum   sql.NullFloat64 `json:"cost_sum"`
		}{BudgetSum: budgetSum, CostSum: costSum}).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total budget and cost: %w", err)
	}
	
	if budgetSum.Valid {
		stats.TotalBudget = budgetSum.Float64
	}
	if costSum.Valid {
		stats.TotalActualCost = costSum.Float64
	}
	
	return stats, nil
}

// ========== 批量操作 ==========

// SaveBatch 批量保存项目
func (r *ProjectRepositoryImpl) SaveBatch(ctx context.Context, projects []*domain.Project) error {
	if len(projects) == 0 {
		return nil
	}
	
	models := make([]ProjectModel, len(projects))
	for i, project := range projects {
		models[i] = *r.domainToModel(project)
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量插入项目
		if err := tx.CreateInBatches(models, 100).Error; err != nil {
			return fmt.Errorf("failed to batch save projects: %w", err)
		}
		
		// 批量插入成员关系
		var members []ProjectMemberModel
		for _, project := range projects {
			for _, member := range project.Members {
				members = append(members, ProjectMemberModel{
					ProjectID: project.ID,
					UserID:    member.UserID,
					Role:      member.Role,
					JoinedAt:  member.JoinedAt,
					LeftAt:    member.LeftAt,
					IsActive:  member.IsActive,
				})
			}
		}
		
		if len(members) > 0 {
			if err := tx.CreateInBatches(members, 100).Error; err != nil {
				return fmt.Errorf("failed to batch save project members: %w", err)
			}
		}
		
		return nil
	})
}

// UpdateBatch 批量更新项目
func (r *ProjectRepositoryImpl) UpdateBatch(ctx context.Context, projects []*domain.Project) error {
	if len(projects) == 0 {
		return nil
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, project := range projects {
			model := r.domainToModel(project)
			model.UpdatedAt = time.Now()
			model.Version++
			
			result := tx.Model(&ProjectModel{}).
				Where("id = ? AND version = ? AND deleted_at IS NULL", project.ID, project.Version-1).
				Updates(model)
			
			if result.Error != nil {
				return fmt.Errorf("failed to update project %s: %w", project.ID, result.Error)
			}
			
			if result.RowsAffected == 0 {
				return fmt.Errorf("project %s version conflict or not found", project.ID)
			}
		}
		
		return nil
	})
}

// DeleteBatch 批量删除项目
func (r *ProjectRepositoryImpl) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	
	now := time.Now()
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量软删除项目
		result := tx.Model(&ProjectModel{}).
			Where("id IN ? AND deleted_at IS NULL", ids).
			Update("deleted_at", now)
		
		if result.Error != nil {
			return fmt.Errorf("failed to batch delete projects: %w", result.Error)
		}
		
		return nil
	})
}

// ========== 成员管理 ==========

// FindMembers 查找项目成员
func (r *ProjectRepositoryImpl) FindMembers(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectMember, error) {
	var models []ProjectMemberModel
	
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND is_active = true", projectID).
		Order("joined_at ASC").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find project members: %w", err)
	}
	
	members := make([]*domain.ProjectMember, len(models))
	for i, model := range models {
		members[i] = &domain.ProjectMember{
			ID:        model.ID,
			ProjectID: model.ProjectID,
			UserID:    model.UserID,
			Role:      model.Role,
			JoinedAt:  model.JoinedAt,
			LeftAt:    model.LeftAt,
			IsActive:  model.IsActive,
		}
	}
	
	return members, nil
}

// AddMember 添加项目成员
func (r *ProjectRepositoryImpl) AddMember(ctx context.Context, member *domain.ProjectMember) error {
	model := &ProjectMemberModel{
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
		LeftAt:    member.LeftAt,
		IsActive:  member.IsActive,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}
	
	member.ID = model.ID
	return nil
}

// UpdateMember 更新项目成员
func (r *ProjectRepositoryImpl) UpdateMember(ctx context.Context, member *domain.ProjectMember) error {
	result := r.db.WithContext(ctx).
		Model(&ProjectMemberModel{}).
		Where("id = ?", member.ID).
		Updates(map[string]interface{}{
			"role":      member.Role,
			"left_at":   member.LeftAt,
			"is_active": member.IsActive,
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update project member: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrProjectMemberNotFound
	}
	
	return nil
}

// RemoveMember 移除项目成员
func (r *ProjectRepositoryImpl) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	now := time.Now()
	
	result := r.db.WithContext(ctx).
		Model(&ProjectMemberModel{}).
		Where("project_id = ? AND user_id = ? AND is_active = true", projectID, userID).
		Updates(map[string]interface{}{
			"left_at":   &now,
			"is_active": false,
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to remove project member: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrProjectMemberNotFound
	}
	
	return nil
}

// ========== 里程碑管理 ==========

// FindMilestones 查找项目里程碑
func (r *ProjectRepositoryImpl) FindMilestones(ctx context.Context, projectID uuid.UUID) ([]*domain.ProjectMilestone, error) {
	var models []ProjectMilestoneModel
	
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("due_date ASC").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find project milestones: %w", err)
	}
	
	milestones := make([]*domain.ProjectMilestone, len(models))
	for i, model := range models {
		milestones[i] = &domain.ProjectMilestone{
			ID:          model.ID,
			ProjectID:   model.ProjectID,
			Name:        model.Name,
			Description: model.Description,
			DueDate:     model.DueDate,
			CompletedAt: model.CompletedAt,
			IsCompleted: model.IsCompleted,
			CreatedAt:   model.CreatedAt,
			UpdatedAt:   model.UpdatedAt,
		}
	}
	
	return milestones, nil
}

// AddMilestone 添加项目里程碑
func (r *ProjectRepositoryImpl) AddMilestone(ctx context.Context, milestone *domain.ProjectMilestone) error {
	model := &ProjectMilestoneModel{
		ProjectID:   milestone.ProjectID,
		Name:        milestone.Name,
		Description: milestone.Description,
		DueDate:     milestone.DueDate,
		CompletedAt: milestone.CompletedAt,
		IsCompleted: milestone.IsCompleted,
		CreatedAt:   milestone.CreatedAt,
		UpdatedAt:   milestone.UpdatedAt,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add project milestone: %w", err)
	}
	
	milestone.ID = model.ID
	return nil
}

// UpdateMilestone 更新项目里程碑
func (r *ProjectRepositoryImpl) UpdateMilestone(ctx context.Context, milestone *domain.ProjectMilestone) error {
	result := r.db.WithContext(ctx).
		Model(&ProjectMilestoneModel{}).
		Where("id = ?", milestone.ID).
		Updates(map[string]interface{}{
			"name":         milestone.Name,
			"description":  milestone.Description,
			"due_date":     milestone.DueDate,
			"completed_at": milestone.CompletedAt,
			"is_completed": milestone.IsCompleted,
			"updated_at":   time.Now(),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update project milestone: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrProjectMilestoneNotFound
	}
	
	return nil
}

// DeleteMilestone 删除项目里程碑
func (r *ProjectRepositoryImpl) DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", milestoneID).
		Delete(&ProjectMilestoneModel{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete project milestone: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrProjectMilestoneNotFound
	}
	
	return nil
}

// ========== 辅助方法 ==========

// applyFilters 应用查询过滤器
func (r *ProjectRepositoryImpl) applyFilters(db *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		switch key {
		case "organization_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("organization_id = ?", v)
			}
		case "owner_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("owner_id = ?", v)
			}
		case "manager_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("manager_id = ?", v)
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
				db = db.Where("end_date <= ?", v)
			}
		case "is_overdue":
			if v, ok := value.(bool); ok && v {
				now := time.Now()
				db = db.Where("due_date < ? AND status NOT IN (?, ?)", 
					now, string(domain.ProjectStatusCompleted), string(domain.ProjectStatusCancelled))
			}
		case "budget_min":
			if v, ok := value.(float64); ok {
				db = db.Where("budget >= ?", v)
			}
		case "budget_max":
			if v, ok := value.(float64); ok {
				db = db.Where("budget <= ?", v)
			}
		}
	}
	
	return db
}

// domainToModel 将领域模型转换为数据库模型
func (r *ProjectRepositoryImpl) domainToModel(project *domain.Project) *ProjectModel {
	model := &ProjectModel{
		ID:             project.ID,
		Name:           project.Name,
		Description:    project.Description,
		Status:         string(project.Status),
		Priority:       string(project.Priority),
		Type:           string(project.Type),
		OrganizationID: project.OrganizationID,
		OwnerID:        project.OwnerID,
		ManagerID:      project.ManagerID,
		StartDate:      project.StartDate,
		EndDate:        project.EndDate,
		DueDate:        project.DueDate,
		CompletedAt:    project.CompletedAt,
		Budget:         project.Budget,
		ActualCost:     project.ActualCost,
		Currency:       project.Currency,
		Tags:           StringSlice(project.Tags),
		Labels:         JSONMap(project.Labels),
		Metadata:       JSONMap(project.Metadata),
		Progress:       project.Progress,
		CreatedAt:      project.CreatedAt,
		UpdatedAt:      project.UpdatedAt,
		Version:        project.Version,
	}
	
	return model
}

// modelToDomain 将数据库模型转换为领域模型
func (r *ProjectRepositoryImpl) modelToDomain(model *ProjectModel) *domain.Project {
	project := &domain.Project{
		ID:             model.ID,
		Name:           model.Name,
		Description:    model.Description,
		Status:         domain.ProjectStatus(model.Status),
		Priority:       domain.ProjectPriority(model.Priority),
		Type:           domain.ProjectType(model.Type),
		OrganizationID: model.OrganizationID,
		OwnerID:        model.OwnerID,
		ManagerID:      model.ManagerID,
		StartDate:      model.StartDate,
		EndDate:        model.EndDate,
		DueDate:        model.DueDate,
		CompletedAt:    model.CompletedAt,
		Budget:         model.Budget,
		ActualCost:     model.ActualCost,
		Currency:       model.Currency,
		Tags:           []string(model.Tags),
		Labels:         map[string]string(model.Labels),
		Metadata:       map[string]interface{}(model.Metadata),
		Progress:       model.Progress,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		Version:        model.Version,
	}
	
	// 转换成员关系
	for _, member := range model.Members {
		project.Members = append(project.Members, &domain.ProjectMember{
			ID:        member.ID,
			ProjectID: member.ProjectID,
			UserID:    member.UserID,
			Role:      member.Role,
			JoinedAt:  member.JoinedAt,
			LeftAt:    member.LeftAt,
			IsActive:  member.IsActive,
		})
	}
	
	// 转换里程碑
	for _, milestone := range model.Milestones {
		project.Milestones = append(project.Milestones, &domain.ProjectMilestone{
			ID:          milestone.ID,
			ProjectID:   milestone.ProjectID,
			Name:        milestone.Name,
			Description: milestone.Description,
			DueDate:     milestone.DueDate,
			CompletedAt: milestone.CompletedAt,
			IsCompleted: milestone.IsCompleted,
			CreatedAt:   milestone.CreatedAt,
			UpdatedAt:   milestone.UpdatedAt,
		})
	}
	
	return project
}