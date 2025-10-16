﻿package postgres

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

// TeamRepositoryImpl PostgreSQL
type TeamRepositoryImpl struct {
	db *gorm.DB
}

// NewTeamRepository 
func NewTeamRepository(db *gorm.DB) domain.TeamRepository {
	return &TeamRepositoryImpl{db: db}
}

// ========== ?==========

// TeamModel ?
type TeamModel struct {
	ID             uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string              `gorm:"not null;size:255" json:"name"`
	Description    string              `gorm:"type:text" json:"description"`
	Status         string              `gorm:"not null;size:50;index" json:"status"`
	
	// 
	OrganizationID uuid.UUID           `gorm:"type:uuid;not null;index" json:"organization_id"`
	LeaderID       *uuid.UUID          `gorm:"type:uuid;index" json:"leader_id"`
	
	// 
	MaxMembers     int                 `gorm:"default:10" json:"max_members"`
	WorkingHours   JSONMap             `gorm:"type:jsonb" json:"working_hours"`
	
	// 
	Tags           StringSlice         `gorm:"type:text[]" json:"tags"`
	Labels         JSONMap             `gorm:"type:jsonb" json:"labels"`
	Metadata       JSONMap             `gorm:"type:jsonb" json:"metadata"`
	
	// 
	CreatedAt      time.Time           `gorm:"not null;index" json:"created_at"`
	UpdatedAt      time.Time           `gorm:"not null;index" json:"updated_at"`
	DeletedAt      *time.Time          `gorm:"index" json:"deleted_at"`
	Version        int                 `gorm:"default:1" json:"version"`
	
	// 
	Members        []TeamMemberModel   `gorm:"foreignKey:TeamID" json:"members"`
	Skills         []TeamSkillModel    `gorm:"foreignKey:TeamID" json:"skills"`
	Metrics        *TeamMetricsModel   `gorm:"foreignKey:TeamID" json:"metrics"`
}

// TeamMemberModel ?
type TeamMemberModel struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TeamID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"team_id"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	Role         string       `gorm:"not null;size:50" json:"role"`
	JoinedAt     time.Time    `gorm:"not null" json:"joined_at"`
	LeftAt       *time.Time   `json:"left_at"`
	IsActive     bool         `gorm:"default:true;index" json:"is_active"`
	Availability float64      `gorm:"default:100" json:"availability"` // 
	
	// 
	Team         TeamModel    `gorm:"foreignKey:TeamID" json:"team"`
}

// TeamSkillModel 
type TeamSkillModel struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TeamID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"team_id"`
	UserID      uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	SkillName   string       `gorm:"not null;size:255" json:"skill_name"`
	Level       int          `gorm:"not null" json:"level"` // ?1-10
	Experience  int          `gorm:"default:0" json:"experience"` // 
	Certified   bool         `gorm:"default:false" json:"certified"` // 
	CreatedAt   time.Time    `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null" json:"updated_at"`
	
	// 
	Team        TeamModel    `gorm:"foreignKey:TeamID" json:"team"`
}

// TeamMetricsModel ?
type TeamMetricsModel struct {
	ID                    uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TeamID                uuid.UUID    `gorm:"type:uuid;not null;unique;index" json:"team_id"`
	TotalTasks            int          `gorm:"default:0" json:"total_tasks"`
	CompletedTasks        int          `gorm:"default:0" json:"completed_tasks"`
	InProgressTasks       int          `gorm:"default:0" json:"in_progress_tasks"`
	OverdueTasks          int          `gorm:"default:0" json:"overdue_tasks"`
	AverageTaskDuration   float64      `gorm:"default:0" json:"average_task_duration"` // 
	ProductivityScore     float64      `gorm:"default:0" json:"productivity_score"`
	CollaborationScore    float64      `gorm:"default:0" json:"collaboration_score"`
	QualityScore          float64      `gorm:"default:0" json:"quality_score"`
	LastCalculatedAt      time.Time    `gorm:"not null" json:"last_calculated_at"`
	
	// 
	Team                  TeamModel    `gorm:"foreignKey:TeamID" json:"team"`
}

// ==========  ==========

func (TeamModel) TableName() string { return "teams" }
func (TeamMemberModel) TableName() string { return "team_members" }
func (TeamSkillModel) TableName() string { return "team_skills" }
func (TeamMetricsModel) TableName() string { return "team_metrics" }

// ========== CRUD ==========

// Save 
func (r *TeamRepositoryImpl) Save(ctx context.Context, team *domain.Team) error {
	model := r.domainToModel(team)
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ?
		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to save team: %w", err)
		}
		
		// 
		if len(team.Members) > 0 {
			for _, member := range team.Members {
				memberModel := &TeamMemberModel{
					TeamID:       team.ID,
					UserID:       member.UserID,
					Role:         string(member.Role),
					JoinedAt:     member.JoinedAt,
					LeftAt:       member.LeftAt,
					IsActive:     member.IsActive,
					Availability: member.Availability,
				}
				if err := tx.Create(memberModel).Error; err != nil {
					return fmt.Errorf("failed to save team member: %w", err)
				}
			}
		}
		
		// 漼?
		if len(team.Skills) > 0 {
			for _, skill := range team.Skills {
				skillModel := &TeamSkillModel{
					TeamID:     team.ID,
					UserID:     skill.UserID,
					SkillName:  skill.SkillName,
					Level:      skill.Level,
					Experience: skill.Experience,
					Certified:  skill.Certified,
					CreatedAt:  skill.CreatedAt,
					UpdatedAt:  skill.UpdatedAt,
				}
				if err := tx.Create(skillModel).Error; err != nil {
					return fmt.Errorf("failed to save team skill: %w", err)
				}
			}
		}
		
		// 
		if team.Metrics != nil {
			metricsModel := &TeamMetricsModel{
				TeamID:                team.ID,
				TotalTasks:            team.Metrics.TotalTasks,
				CompletedTasks:        team.Metrics.CompletedTasks,
				InProgressTasks:       team.Metrics.InProgressTasks,
				OverdueTasks:          team.Metrics.OverdueTasks,
				AverageTaskDuration:   team.Metrics.AverageTaskDuration,
				ProductivityScore:     team.Metrics.ProductivityScore,
				CollaborationScore:    team.Metrics.CollaborationScore,
				QualityScore:          team.Metrics.QualityScore,
				LastCalculatedAt:      team.Metrics.LastCalculatedAt,
			}
			if err := tx.Create(metricsModel).Error; err != nil {
				return fmt.Errorf("failed to save team metrics: %w", err)
			}
		}
		
		return nil
	})
}

// FindByID ID
func (r *TeamRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*domain.Team, error) {
	var model TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to find team: %w", err)
	}
	
	return r.modelToDomain(&model), nil
}

// Update 
func (r *TeamRepositoryImpl) Update(ctx context.Context, team *domain.Team) error {
	model := r.domainToModel(team)
	model.UpdatedAt = time.Now()
	model.Version++
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ?
		result := tx.Model(&TeamModel{}).
			Where("id = ? AND version = ? AND deleted_at IS NULL", team.ID, team.Version-1).
			Updates(model)
		
		if result.Error != nil {
			return fmt.Errorf("failed to update team: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return domain.ErrTeamVersionConflict
		}
		
		return nil
	})
}

// Delete ?
func (r *TeamRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ?
		result := tx.Model(&TeamModel{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Update("deleted_at", now)
		
		if result.Error != nil {
			return fmt.Errorf("failed to delete team: %w", result.Error)
		}
		
		if result.RowsAffected == 0 {
			return domain.ErrTeamNotFound
		}
		
		// ?
		if err := tx.Model(&TeamMemberModel{}).
			Where("team_id = ? AND is_active = true", id).
			Updates(map[string]interface{}{
				"left_at":   &now,
				"is_active": false,
			}).Error; err != nil {
			return fmt.Errorf("failed to deactivate team members: %w", err)
		}
		
		return nil
	})
}

// ==========  ==========

// FindByOrganizationID ID
func (r *TeamRepositoryImpl) FindByOrganizationID(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by organization: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// FindByLeaderID ID
func (r *TeamRepositoryImpl) FindByLeaderID(ctx context.Context, leaderID uuid.UUID, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Where("leader_id = ? AND deleted_at IS NULL", leaderID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by leader: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// FindByMemberID ID
func (r *TeamRepositoryImpl) FindByMemberID(ctx context.Context, memberID uuid.UUID, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Joins("JOIN team_members tm ON teams.id = tm.team_id").
		Where("tm.user_id = ? AND tm.is_active = true AND teams.deleted_at IS NULL", memberID).
		Order("teams.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by member: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// FindByStatus ?
func (r *TeamRepositoryImpl) FindByStatus(ctx context.Context, status domain.TeamStatus, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by status: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// ==========  ==========

// SearchTeams 
func (r *TeamRepositoryImpl) SearchTeams(ctx context.Context, query string, filters map[string]interface{}, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	db := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics")
	
	// 
	if query != "" {
		searchQuery := "%" + strings.ToLower(query) + "%"
		db = db.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", searchQuery, searchQuery)
	}
	
	// ?
	db = r.applyFilters(db, filters)
	
	err := db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to search teams: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// FindByTags 
func (r *TeamRepositoryImpl) FindByTags(ctx context.Context, tags []string, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Where("tags && ? AND deleted_at IS NULL", tags).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by tags: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// FindByLabels 
func (r *TeamRepositoryImpl) FindByLabels(ctx context.Context, labels map[string]string, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	db := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics")
	
	// 
	for key, value := range labels {
		db = db.Where("labels ->> ? = ?", key, value)
	}
	
	err := db.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by labels: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// FindBySkill ?
func (r *TeamRepositoryImpl) FindBySkill(ctx context.Context, skillName string, minLevel int, limit, offset int) ([]*domain.Team, error) {
	var models []TeamModel
	
	err := r.db.WithContext(ctx).
		Preload("Members").
		Preload("Skills").
		Preload("Metrics").
		Joins("JOIN team_skills ts ON teams.id = ts.team_id").
		Where("ts.skill_name = ? AND ts.level >= ? AND teams.deleted_at IS NULL", skillName, minLevel).
		Group("teams.id").
		Order("teams.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by skill: %w", err)
	}
	
	teams := make([]*domain.Team, len(models))
	for i, model := range models {
		teams[i] = r.modelToDomain(&model)
	}
	
	return teams, nil
}

// ==========  ==========

// Count 
func (r *TeamRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TeamModel{}).
		Where("deleted_at IS NULL").
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count teams: %w", err)
	}
	
	return count, nil
}

// CountByOrganization ?
func (r *TeamRepositoryImpl) CountByOrganization(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TeamModel{}).
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count teams by organization: %w", err)
	}
	
	return count, nil
}

// CountByStatus 
func (r *TeamRepositoryImpl) CountByStatus(ctx context.Context, status domain.TeamStatus) (int64, error) {
	var count int64
	
	err := r.db.WithContext(ctx).
		Model(&TeamModel{}).
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Count(&count).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to count teams by status: %w", err)
	}
	
	return count, nil
}

// GetTeamStatistics 
func (r *TeamRepositoryImpl) GetTeamStatistics(ctx context.Context, organizationID *uuid.UUID) (*domain.TeamStatistics, error) {
	stats := &domain.TeamStatistics{}
	
	db := r.db.WithContext(ctx).Model(&TeamModel{}).Where("deleted_at IS NULL")
	
	// 
	if organizationID != nil {
		db = db.Where("organization_id = ?", *organizationID)
	}
	
	// 
	if err := db.Count(&stats.TotalTeams).Error; err != nil {
		return nil, fmt.Errorf("failed to count total teams: %w", err)
	}
	
	// ?
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := db.Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to count teams by status: %w", err)
	}
	
	for _, sc := range statusCounts {
		switch domain.TeamStatus(sc.Status) {
		case domain.TeamStatusActive:
			stats.ActiveTeams = sc.Count
		case domain.TeamStatusInactive:
			stats.InactiveTeams = sc.Count
		case domain.TeamStatusDisbanded:
			stats.DisbandedTeams = sc.Count
		}
	}
	
	// 
	var avgSize sql.NullFloat64
	memberQuery := `
		SELECT AVG(member_count) 
		FROM (
			SELECT COUNT(*) as member_count 
			FROM team_members tm 
			JOIN teams t ON tm.team_id = t.id 
			WHERE tm.is_active = true AND t.deleted_at IS NULL
	`
	
	if organizationID != nil {
		memberQuery += " AND t.organization_id = ?"
		if err := r.db.WithContext(ctx).Raw(memberQuery+" GROUP BY tm.team_id) as team_sizes", *organizationID).Scan(&avgSize).Error; err != nil {
			return nil, fmt.Errorf("failed to calculate average team size: %w", err)
		}
	} else {
		memberQuery += " GROUP BY tm.team_id) as team_sizes"
		if err := r.db.WithContext(ctx).Raw(memberQuery).Scan(&avgSize).Error; err != nil {
			return nil, fmt.Errorf("failed to calculate average team size: %w", err)
		}
	}
	
	if avgSize.Valid {
		stats.AverageTeamSize = avgSize.Float64
	}
	
	// 
	memberCountQuery := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Joins("JOIN teams ON team_members.team_id = teams.id").
		Where("team_members.is_active = true AND teams.deleted_at IS NULL")
	
	if organizationID != nil {
		memberCountQuery = memberCountQuery.Where("teams.organization_id = ?", *organizationID)
	}
	
	if err := memberCountQuery.Count(&stats.TotalMembers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total members: %w", err)
	}
	
	// ?
	var avgProductivity sql.NullFloat64
	metricsQuery := r.db.WithContext(ctx).
		Model(&TeamMetricsModel{}).
		Joins("JOIN teams ON team_metrics.team_id = teams.id").
		Where("teams.deleted_at IS NULL")
	
	if organizationID != nil {
		metricsQuery = metricsQuery.Where("teams.organization_id = ?", *organizationID)
	}
	
	if err := metricsQuery.Select("AVG(productivity_score)").Scan(&avgProductivity).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate average productivity: %w", err)
	}
	
	if avgProductivity.Valid {
		stats.AverageProductivity = avgProductivity.Float64
	}
	
	return stats, nil
}

// ==========  ==========

// SaveBatch 
func (r *TeamRepositoryImpl) SaveBatch(ctx context.Context, teams []*domain.Team) error {
	if len(teams) == 0 {
		return nil
	}
	
	models := make([]TeamModel, len(teams))
	for i, team := range teams {
		models[i] = *r.domainToModel(team)
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 
		if err := tx.CreateInBatches(models, 100).Error; err != nil {
			return fmt.Errorf("failed to batch save teams: %w", err)
		}
		
		// 
		var members []TeamMemberModel
		for _, team := range teams {
			for _, member := range team.Members {
				members = append(members, TeamMemberModel{
					TeamID:       team.ID,
					UserID:       member.UserID,
					Role:         string(member.Role),
					JoinedAt:     member.JoinedAt,
					LeftAt:       member.LeftAt,
					IsActive:     member.IsActive,
					Availability: member.Availability,
				})
			}
		}
		
		if len(members) > 0 {
			if err := tx.CreateInBatches(members, 100).Error; err != nil {
				return fmt.Errorf("failed to batch save team members: %w", err)
			}
		}
		
		return nil
	})
}

// UpdateBatch 
func (r *TeamRepositoryImpl) UpdateBatch(ctx context.Context, teams []*domain.Team) error {
	if len(teams) == 0 {
		return nil
	}
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, team := range teams {
			model := r.domainToModel(team)
			model.UpdatedAt = time.Now()
			model.Version++
			
			result := tx.Model(&TeamModel{}).
				Where("id = ? AND version = ? AND deleted_at IS NULL", team.ID, team.Version-1).
				Updates(model)
			
			if result.Error != nil {
				return fmt.Errorf("failed to update team %s: %w", team.ID, result.Error)
			}
			
			if result.RowsAffected == 0 {
				return fmt.Errorf("team %s version conflict or not found", team.ID)
			}
		}
		
		return nil
	})
}

// DeleteBatch 
func (r *TeamRepositoryImpl) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	
	now := time.Now()
	
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ?
		result := tx.Model(&TeamModel{}).
			Where("id IN ? AND deleted_at IS NULL", ids).
			Update("deleted_at", now)
		
		if result.Error != nil {
			return fmt.Errorf("failed to batch delete teams: %w", result.Error)
		}
		
		// ?
		if err := tx.Model(&TeamMemberModel{}).
			Where("team_id IN ? AND is_active = true", ids).
			Updates(map[string]interface{}{
				"left_at":   &now,
				"is_active": false,
			}).Error; err != nil {
			return fmt.Errorf("failed to deactivate team members: %w", err)
		}
		
		return nil
	})
}

// ==========  ==========

// FindMembers 
func (r *TeamRepositoryImpl) FindMembers(ctx context.Context, teamID uuid.UUID) ([]*domain.TeamMember, error) {
	var models []TeamMemberModel
	
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND is_active = true", teamID).
		Order("joined_at ASC").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find team members: %w", err)
	}
	
	members := make([]*domain.TeamMember, len(models))
	for i, model := range models {
		members[i] = &domain.TeamMember{
			ID:           model.ID,
			TeamID:       model.TeamID,
			UserID:       model.UserID,
			Role:         domain.TeamMemberRole(model.Role),
			JoinedAt:     model.JoinedAt,
			LeftAt:       model.LeftAt,
			IsActive:     model.IsActive,
			Availability: model.Availability,
		}
	}
	
	return members, nil
}

// AddMember 
func (r *TeamRepositoryImpl) AddMember(ctx context.Context, member *domain.TeamMember) error {
	model := &TeamMemberModel{
		TeamID:       member.TeamID,
		UserID:       member.UserID,
		Role:         string(member.Role),
		JoinedAt:     member.JoinedAt,
		LeftAt:       member.LeftAt,
		IsActive:     member.IsActive,
		Availability: member.Availability,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}
	
	member.ID = model.ID
	return nil
}

// UpdateMember 
func (r *TeamRepositoryImpl) UpdateMember(ctx context.Context, member *domain.TeamMember) error {
	result := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Where("id = ?", member.ID).
		Updates(map[string]interface{}{
			"role":         string(member.Role),
			"left_at":      member.LeftAt,
			"is_active":    member.IsActive,
			"availability": member.Availability,
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update team member: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTeamMemberNotFound
	}
	
	return nil
}

// RemoveMember 
func (r *TeamRepositoryImpl) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	now := time.Now()
	
	result := r.db.WithContext(ctx).
		Model(&TeamMemberModel{}).
		Where("team_id = ? AND user_id = ? AND is_active = true", teamID, userID).
		Updates(map[string]interface{}{
			"left_at":   &now,
			"is_active": false,
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to remove team member: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTeamMemberNotFound
	}
	
	return nil
}

// ========== ?==========

// FindSkills ?
func (r *TeamRepositoryImpl) FindSkills(ctx context.Context, teamID uuid.UUID) ([]*domain.TeamSkill, error) {
	var models []TeamSkillModel
	
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("skill_name ASC, level DESC").
		Find(&models).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to find team skills: %w", err)
	}
	
	skills := make([]*domain.TeamSkill, len(models))
	for i, model := range models {
		skills[i] = &domain.TeamSkill{
			ID:         model.ID,
			TeamID:     model.TeamID,
			UserID:     model.UserID,
			SkillName:  model.SkillName,
			Level:      model.Level,
			Experience: model.Experience,
			Certified:  model.Certified,
			CreatedAt:  model.CreatedAt,
			UpdatedAt:  model.UpdatedAt,
		}
	}
	
	return skills, nil
}

// AddSkill ?
func (r *TeamRepositoryImpl) AddSkill(ctx context.Context, skill *domain.TeamSkill) error {
	model := &TeamSkillModel{
		TeamID:     skill.TeamID,
		UserID:     skill.UserID,
		SkillName:  skill.SkillName,
		Level:      skill.Level,
		Experience: skill.Experience,
		Certified:  skill.Certified,
		CreatedAt:  skill.CreatedAt,
		UpdatedAt:  skill.UpdatedAt,
	}
	
	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to add team skill: %w", err)
	}
	
	skill.ID = model.ID
	return nil
}

// UpdateSkill ?
func (r *TeamRepositoryImpl) UpdateSkill(ctx context.Context, skill *domain.TeamSkill) error {
	result := r.db.WithContext(ctx).
		Model(&TeamSkillModel{}).
		Where("id = ?", skill.ID).
		Updates(map[string]interface{}{
			"skill_name": skill.SkillName,
			"level":      skill.Level,
			"experience": skill.Experience,
			"certified":  skill.Certified,
			"updated_at": time.Now(),
		})
	
	if result.Error != nil {
		return fmt.Errorf("failed to update team skill: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTeamSkillNotFound
	}
	
	return nil
}

// DeleteSkill ?
func (r *TeamRepositoryImpl) DeleteSkill(ctx context.Context, skillID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", skillID).
		Delete(&TeamSkillModel{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to delete team skill: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return domain.ErrTeamSkillNotFound
	}
	
	return nil
}

// ==========  ==========

// applyFilters ?
func (r *TeamRepositoryImpl) applyFilters(db *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		switch key {
		case "organization_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("organization_id = ?", v)
			}
		case "leader_id":
			if v, ok := value.(uuid.UUID); ok {
				db = db.Where("leader_id = ?", v)
			}
		case "status":
			if v, ok := value.(string); ok {
				db = db.Where("status = ?", v)
			}
		case "max_members_min":
			if v, ok := value.(int); ok {
				db = db.Where("max_members >= ?", v)
			}
		case "max_members_max":
			if v, ok := value.(int); ok {
				db = db.Where("max_members <= ?", v)
			}
		case "has_leader":
			if v, ok := value.(bool); ok {
				if v {
					db = db.Where("leader_id IS NOT NULL")
				} else {
					db = db.Where("leader_id IS NULL")
				}
			}
		}
	}
	
	return db
}

// domainToModel ?
func (r *TeamRepositoryImpl) domainToModel(team *domain.Team) *TeamModel {
	model := &TeamModel{
		ID:             team.ID,
		Name:           team.Name,
		Description:    team.Description,
		Status:         string(team.Status),
		OrganizationID: team.OrganizationID,
		LeaderID:       team.LeaderID,
		MaxMembers:     team.MaxMembers,
		WorkingHours:   JSONMap(team.WorkingHours),
		Tags:           StringSlice(team.Tags),
		Labels:         JSONMap(team.Labels),
		Metadata:       JSONMap(team.Metadata),
		CreatedAt:      team.CreatedAt,
		UpdatedAt:      team.UpdatedAt,
		Version:        team.Version,
	}
	
	return model
}

// modelToDomain ?
func (r *TeamRepositoryImpl) modelToDomain(model *TeamModel) *domain.Team {
	team := &domain.Team{
		ID:             model.ID,
		Name:           model.Name,
		Description:    model.Description,
		Status:         domain.TeamStatus(model.Status),
		OrganizationID: model.OrganizationID,
		LeaderID:       model.LeaderID,
		MaxMembers:     model.MaxMembers,
		WorkingHours:   map[string]interface{}(model.WorkingHours),
		Tags:           []string(model.Tags),
		Labels:         map[string]string(model.Labels),
		Metadata:       map[string]interface{}(model.Metadata),
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		Version:        model.Version,
	}
	
	// 
	for _, member := range model.Members {
		team.Members = append(team.Members, &domain.TeamMember{
			ID:           member.ID,
			TeamID:       member.TeamID,
			UserID:       member.UserID,
			Role:         domain.TeamMemberRole(member.Role),
			JoinedAt:     member.JoinedAt,
			LeftAt:       member.LeftAt,
			IsActive:     member.IsActive,
			Availability: member.Availability,
		})
	}
	
	// ?
	for _, skill := range model.Skills {
		team.Skills = append(team.Skills, &domain.TeamSkill{
			ID:         skill.ID,
			TeamID:     skill.TeamID,
			UserID:     skill.UserID,
			SkillName:  skill.SkillName,
			Level:      skill.Level,
			Experience: skill.Experience,
			Certified:  skill.Certified,
			CreatedAt:  skill.CreatedAt,
			UpdatedAt:  skill.UpdatedAt,
		})
	}
	
	// 
	if model.Metrics != nil {
		team.Metrics = &domain.TeamMetrics{
			TotalTasks:            model.Metrics.TotalTasks,
			CompletedTasks:        model.Metrics.CompletedTasks,
			InProgressTasks:       model.Metrics.InProgressTasks,
			OverdueTasks:          model.Metrics.OverdueTasks,
			AverageTaskDuration:   model.Metrics.AverageTaskDuration,
			ProductivityScore:     model.Metrics.ProductivityScore,
			CollaborationScore:    model.Metrics.CollaborationScore,
			QualityScore:          model.Metrics.QualityScore,
			LastCalculatedAt:      model.Metrics.LastCalculatedAt,
		}
	}
	
	return team
}

