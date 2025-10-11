package handlers

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// CreateTeamRequest 创建团队请求
type CreateTeamRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=200"`
	Description    string            `json:"description" validate:"max=2000"`
	OrganizationID uuid.UUID         `json:"organization_id" validate:"required"`
	LeaderID       uuid.UUID         `json:"leader_id" validate:"required"`
	Type           domain.TeamType   `json:"type" validate:"required"`
	Status         domain.TeamStatus `json:"status" validate:"required"`
	MaxMembers     int               `json:"max_members" validate:"min=1"`
	Tags           []string          `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// Validate 验证创建团队请求
func (r *CreateTeamRequest) Validate() error {
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
	if r.LeaderID == uuid.Nil {
		return errors.New("leader_id is required")
	}
	if r.Type == "" {
		return errors.New("type is required")
	}
	if r.Status == "" {
		return errors.New("status is required")
	}
	if r.MaxMembers < 1 {
		return errors.New("max_members must be at least 1")
	}
	return nil
}

// UpdateTeamRequest 更新团队请求
type UpdateTeamRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	LeaderID    *uuid.UUID         `json:"leader_id,omitempty"`
	Type        *domain.TeamType   `json:"type,omitempty"`
	Status      *domain.TeamStatus `json:"status,omitempty"`
	MaxMembers  *int               `json:"max_members,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
}

// AddTeamMemberRequest 添加团队成员请求
type AddTeamMemberRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Role    string    `json:"role" validate:"required"`
	AddedBy uuid.UUID `json:"added_by" validate:"required"`
}

// RemoveTeamMemberRequest 移除团队成员请求
type RemoveTeamMemberRequest struct {
	RemovedBy uuid.UUID `json:"removed_by" validate:"required"`
}

// UpdateTeamMemberRequest 更新团队成员请求
type UpdateTeamMemberRequest struct {
	Role      *string   `json:"role,omitempty"`
	UpdatedBy uuid.UUID `json:"updated_by" validate:"required"`
}

// AddTeamSkillRequest 添加团队技能请�?
type AddTeamSkillRequest struct {
	SkillName   string    `json:"skill_name" validate:"required,min=1,max=100"`
	Level       int       `json:"level" validate:"min=1,max=5"`
	Description string    `json:"description,omitempty"`
	AddedBy     uuid.UUID `json:"added_by" validate:"required"`
}

// UpdateTeamSkillRequest 更新团队技能请�?
type UpdateTeamSkillRequest struct {
	Level       *int      `json:"level,omitempty"`
	Description *string   `json:"description,omitempty"`
	UpdatedBy   uuid.UUID `json:"updated_by" validate:"required"`
}

// RemoveTeamSkillRequest 移除团队技能请�?
type RemoveTeamSkillRequest struct {
	RemovedBy uuid.UUID `json:"removed_by" validate:"required"`
}

// TeamResponse 团队响应
type TeamResponse struct {
	ID             uuid.UUID         `json:"id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	OrganizationID uuid.UUID         `json:"organization_id"`
	LeaderID       uuid.UUID         `json:"leader_id"`
	Type           domain.TeamType   `json:"type"`
	Status         domain.TeamStatus `json:"status"`
	MaxMembers     int               `json:"max_members"`
	Tags           []string          `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// ListTeamsResponse 团队列表响应
type ListTeamsResponse struct {
	Teams  []TeamResponse `json:"teams"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// TeamMemberResponse 团队成员响应
type TeamMemberResponse struct {
	ID       uuid.UUID `json:"id"`
	TeamID   uuid.UUID `json:"team_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// TeamSkillResponse 团队技能响�?
type TeamSkillResponse struct {
	ID          uuid.UUID `json:"id"`
	TeamID      uuid.UUID `json:"team_id"`
	SkillName   string    `json:"skill_name"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
