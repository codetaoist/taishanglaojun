package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// TeamService 
type TeamService struct {
	teamRepo        domain.TeamRepository
	taskRepo        domain.TaskRepository
	projectRepo     domain.ProjectRepository
	eventRepo       domain.EventRepository
	allocationSvc   domain.TaskAllocationService
	performanceSvc  domain.PerformanceAnalysisService
	notificationSvc domain.NotificationService
	unitOfWork      domain.UnitOfWork
}

// NewTeamService 
func NewTeamService(
	teamRepo domain.TeamRepository,
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	eventRepo domain.EventRepository,
	allocationSvc domain.TaskAllocationService,
	performanceSvc domain.PerformanceAnalysisService,
	notificationSvc domain.NotificationService,
	unitOfWork domain.UnitOfWork,
) *TeamService {
	return &TeamService{
		teamRepo:        teamRepo,
		taskRepo:        taskRepo,
		projectRepo:     projectRepo,
		eventRepo:       eventRepo,
		allocationSvc:   allocationSvc,
		performanceSvc:  performanceSvc,
		notificationSvc: notificationSvc,
		unitOfWork:      unitOfWork,
	}
}

// ========== CRUD ==========

// CreateTeamRequest 
type CreateTeamRequest struct {
	Name           string                 `json:"name" validate:"required,min=1,max=255"`
	Description    string                 `json:"description"`
	OrganizationID uuid.UUID              `json:"organization_id" validate:"required"`
	LeaderID       uuid.UUID              `json:"leader_id" validate:"required"`
	MaxMembers     *int                   `json:"max_members,omitempty"`
	WorkingHours   *float64               `json:"working_hours,omitempty"`
	Tags           []string               `json:"tags"`
	Labels         map[string]string      `json:"labels"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// CreateTeamResponse 
type CreateTeamResponse struct {
	Team *domain.Team `json:"team"`
}

// CreateTeam 
func (s *TeamService) CreateTeam(ctx context.Context, req *CreateTeamRequest) (*CreateTeamResponse, error) {
	// 
	team, err := domain.NewTeam(
		req.Name,
		req.Description,
		req.OrganizationID,
		req.LeaderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// ?
	if req.MaxMembers != nil {
		team.SetMaxMembers(*req.MaxMembers)
	}

	if req.WorkingHours != nil {
		team.SetWorkingHours(*req.WorkingHours)
	}

	if len(req.Tags) > 0 {
		team.Tags = req.Tags
	}

	if len(req.Labels) > 0 {
		team.Labels = req.Labels
	}

	if len(req.Metadata) > 0 {
		team.Metadata = req.Metadata
	}

	// ?
	leader := &domain.TeamMember{
		ID:           uuid.New(),
		TeamID:       team.ID,
		UserID:       req.LeaderID,
		Role:         domain.TeamMemberRoleLeader,
		JoinedAt:     time.Now(),
		IsAvailable:  true,
		Availability: 100.0,
	}
	team.AddMember(leader)

	// 
	if err := s.teamRepo.Save(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to save team: %w", err)
	}

	// 
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 
	notification := &domain.TeamNotification{
		Type:     domain.NotificationTypeTeamCreated,
		TeamID:   team.ID,
		UserID:   req.LeaderID,
		Title:    "",
		Message:  fmt.Sprintf(" %s ", team.Name),
		Metadata: map[string]interface{}{"team_id": team.ID.String()},
	}
	if err := s.notificationSvc.SendTeamNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	team.ClearDomainEvents()

	return &CreateTeamResponse{Team: team}, nil
}

// GetTeamRequest 
type GetTeamRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
}

// GetTeamResponse 
type GetTeamResponse struct {
	Team *domain.Team `json:"team"`
}

// GetTeam 
func (s *TeamService) GetTeam(ctx context.Context, req *GetTeamRequest) (*GetTeamResponse, error) {
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	return &GetTeamResponse{Team: team}, nil
}

// UpdateTeamRequest 
type UpdateTeamRequest struct {
	TeamID       uuid.UUID              `json:"team_id" validate:"required"`
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Status       *domain.TeamStatus     `json:"status,omitempty"`
	LeaderID     *uuid.UUID             `json:"leader_id,omitempty"`
	MaxMembers   *int                   `json:"max_members,omitempty"`
	WorkingHours *float64               `json:"working_hours,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Labels       map[string]string      `json:"labels,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTeamResponse 
type UpdateTeamResponse struct {
	Team *domain.Team `json:"team"`
}

// UpdateTeam 
func (s *TeamService) UpdateTeam(ctx context.Context, req *UpdateTeamRequest) (*UpdateTeamResponse, error) {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 
	if req.Name != nil {
		team.Name = *req.Name
	}

	if req.Description != nil {
		team.Description = *req.Description
	}

	if req.Status != nil {
		switch *req.Status {
		case domain.TeamStatusDisbanded:
			if err := team.Disband(); err != nil {
				return nil, fmt.Errorf("failed to disband team: %w", err)
			}
		}
	}

	if req.LeaderID != nil {
		if err := team.ChangeLeader(*req.LeaderID); err != nil {
			return nil, fmt.Errorf("failed to change leader: %w", err)
		}
	}

	if req.MaxMembers != nil {
		team.SetMaxMembers(*req.MaxMembers)
	}

	if req.WorkingHours != nil {
		team.SetWorkingHours(*req.WorkingHours)
	}

	if req.Tags != nil {
		team.Tags = req.Tags
	}

	if req.Labels != nil {
		team.Labels = req.Labels
	}

	if req.Metadata != nil {
		team.Metadata = req.Metadata
	}

	// 
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	// 
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	team.ClearDomainEvents()

	return &UpdateTeamResponse{Team: team}, nil
}

// DeleteTeamRequest 
type DeleteTeamRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
}

// DeleteTeam 
func (s *TeamService) DeleteTeam(ctx context.Context, req *DeleteTeamRequest) error {
	// ?
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// ?
	if team.Status == domain.TeamStatusActive {
		return fmt.Errorf("cannot delete active team")
	}

	// 
	if err := s.teamRepo.Delete(ctx, req.TeamID); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

// ==========  ==========

// ListTeamsRequest 
type ListTeamsRequest struct {
	OrganizationID *uuid.UUID         `json:"organization_id,omitempty"`
	LeaderID       *uuid.UUID         `json:"leader_id,omitempty"`
	MemberID       *uuid.UUID         `json:"member_id,omitempty"`
	Status         *domain.TeamStatus `json:"status,omitempty"`
	Tags           []string           `json:"tags,omitempty"`
	Labels         map[string]string  `json:"labels,omitempty"`
	Skills         []string           `json:"skills,omitempty"`
	Limit          int                `json:"limit" validate:"min=1,max=100"`
	Offset         int                `json:"offset" validate:"min=0"`
	SortBy         string             `json:"sort_by"`
	SortOrder      string             `json:"sort_order"`
}

// ListTeamsResponse 
type ListTeamsResponse struct {
	Teams   []*domain.Team `json:"teams"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}

// ListTeams 
func (s *TeamService) ListTeams(ctx context.Context, req *ListTeamsRequest) (*ListTeamsResponse, error) {
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
	if req.LeaderID != nil {
		filters["leader_id"] = *req.LeaderID
	}
	if req.MemberID != nil {
		filters["member_id"] = *req.MemberID
	}
	if req.Status != nil {
		filters["status"] = string(*req.Status)
	}

	var teams []*domain.Team
	var err error

	// 
	if len(req.Tags) > 0 {
		teams, err = s.teamRepo.FindByTags(ctx, req.Tags, req.Limit, req.Offset)
	} else if len(req.Labels) > 0 {
		teams, err = s.teamRepo.FindByLabels(ctx, req.Labels, req.Limit, req.Offset)
	} else if len(req.Skills) > 0 {
		teams, err = s.teamRepo.FindBySkills(ctx, req.Skills, req.Limit, req.Offset)
	} else {
		teams, err = s.teamRepo.SearchTeams(ctx, "", filters, req.Limit, req.Offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	// 
	total, err := s.teamRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count teams: %w", err)
	}

	hasMore := int64(req.Offset+len(teams)) < total

	return &ListTeamsResponse{
		Teams:   teams,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: hasMore,
	}, nil
}

// SearchTeamsRequest 
type SearchTeamsRequest struct {
	Query   string                 `json:"query" validate:"required"`
	Filters map[string]interface{} `json:"filters,omitempty"`
	Limit   int                    `json:"limit" validate:"min=1,max=100"`
	Offset  int                    `json:"offset" validate:"min=0"`
}

// SearchTeamsResponse 
type SearchTeamsResponse struct {
	Teams   []*domain.Team `json:"teams"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}

// SearchTeams 
func (s *TeamService) SearchTeams(ctx context.Context, req *SearchTeamsRequest) (*SearchTeamsResponse, error) {
	teams, err := s.teamRepo.SearchTeams(ctx, req.Query, req.Filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search teams: %w", err)
	}

	// ?
	total, err := s.teamRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count teams: %w", err)
	}

	hasMore := int64(req.Offset+len(teams)) < total

	return &SearchTeamsResponse{
		Teams:   teams,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: hasMore,
	}, nil
}

// ==========  ==========

// AddTeamMemberRequest 
type AddTeamMemberRequest struct {
	TeamID       uuid.UUID                `json:"team_id" validate:"required"`
	UserID       uuid.UUID                `json:"user_id" validate:"required"`
	Role         domain.TeamMemberRole    `json:"role" validate:"required"`
	Availability *float64                 `json:"availability,omitempty"`
	Skills       []string                 `json:"skills,omitempty"`
}

// AddTeamMemberResponse 
type AddTeamMemberResponse struct {
	Member *domain.TeamMember `json:"member"`
}

// AddTeamMember 
func (s *TeamService) AddTeamMember(ctx context.Context, req *AddTeamMemberRequest) (*AddTeamMemberResponse, error) {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 
	member := &domain.TeamMember{
		ID:           uuid.New(),
		TeamID:       req.TeamID,
		UserID:       req.UserID,
		Role:         req.Role,
		JoinedAt:     time.Now(),
		IsAvailable:  true,
		Availability: 100.0,
	}

	if req.Availability != nil {
		member.Availability = *req.Availability
	}

	// 
	if err := team.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// 
	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to save member: %w", err)
	}

	// ?
	if len(req.Skills) > 0 {
		for _, skillName := range req.Skills {
			skill := &domain.TeamSkill{
				ID:       uuid.New(),
				TeamID:   req.TeamID,
				UserID:   req.UserID,
				Skill:    skillName,
				Level:    domain.SkillLevelIntermediate, // 
				Verified: false,
			}
			if err := s.teamRepo.AddSkill(ctx, skill); err != nil {
				fmt.Printf("failed to add skill %s: %v\n", skillName, err)
			}
		}
	}

	// 
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 
	notification := &domain.TeamNotification{
		Type:     domain.NotificationTypeTeamMemberAdded,
		TeamID:   team.ID,
		UserID:   req.UserID,
		Title:    "?,
		Message:  fmt.Sprintf("? %s", team.Name),
		Metadata: map[string]interface{}{
			"team_id": team.ID.String(),
			"role":    string(req.Role),
		},
	}
	if err := s.notificationSvc.SendTeamNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	team.ClearDomainEvents()

	return &AddTeamMemberResponse{Member: member}, nil
}

// RemoveTeamMemberRequest 
type RemoveTeamMemberRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// RemoveTeamMember 
func (s *TeamService) RemoveTeamMember(ctx context.Context, req *RemoveTeamMemberRequest) error {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// 
	if err := team.RemoveMember(req.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// 
	if err := s.teamRepo.RemoveMember(ctx, req.TeamID, req.UserID); err != nil {
		return fmt.Errorf("failed to save member removal: %w", err)
	}

	// 
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	team.ClearDomainEvents()

	return nil
}

// UpdateTeamMemberRequest 
type UpdateTeamMemberRequest struct {
	TeamID       uuid.UUID             `json:"team_id" validate:"required"`
	UserID       uuid.UUID             `json:"user_id" validate:"required"`
	Role         *domain.TeamMemberRole `json:"role,omitempty"`
	IsAvailable  *bool                 `json:"is_available,omitempty"`
	Availability *float64              `json:"availability,omitempty"`
}

// UpdateTeamMember 
func (s *TeamService) UpdateTeamMember(ctx context.Context, req *UpdateTeamMemberRequest) error {
	// 
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// 
	if req.Role != nil {
		if err := team.UpdateMemberRole(req.UserID, *req.Role); err != nil {
			return fmt.Errorf("failed to update member role: %w", err)
		}
	}

	if req.IsAvailable != nil || req.Availability != nil {
		availability := 100.0
		if req.Availability != nil {
			availability = *req.Availability
		}
		isAvailable := true
		if req.IsAvailable != nil {
			isAvailable = *req.IsAvailable
		}
		if err := team.UpdateMemberAvailability(req.UserID, isAvailable, availability); err != nil {
			return fmt.Errorf("failed to update member availability: %w", err)
		}
	}

	// 
	if err := s.teamRepo.UpdateMember(ctx, req.TeamID, req.UserID, req.Role, req.IsAvailable, req.Availability); err != nil {
		return fmt.Errorf("failed to save member update: %w", err)
	}

	// 
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	team.ClearDomainEvents()

	return nil
}

// ========== ?==========

// AddTeamSkillRequest ?
type AddTeamSkillRequest struct {
	TeamID   uuid.UUID           `json:"team_id" validate:"required"`
	UserID   uuid.UUID           `json:"user_id" validate:"required"`
	Skill    string              `json:"skill" validate:"required"`
	Level    domain.SkillLevel   `json:"level" validate:"required"`
	Verified bool                `json:"verified"`
}

// AddTeamSkillResponse ?
type AddTeamSkillResponse struct {
	Skill *domain.TeamSkill `json:"skill"`
}

// AddTeamSkill ?
func (s *TeamService) AddTeamSkill(ctx context.Context, req *AddTeamSkillRequest) (*AddTeamSkillResponse, error) {
	// ?
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 
	isMember := false
	for _, member := range team.Members {
		if member.UserID == req.UserID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, fmt.Errorf("user is not a team member")
	}

	// ?
	skill := &domain.TeamSkill{
		ID:       uuid.New(),
		TeamID:   req.TeamID,
		UserID:   req.UserID,
		Skill:    req.Skill,
		Level:    req.Level,
		Verified: req.Verified,
	}

	// 漼?
	if err := s.teamRepo.AddSkill(ctx, skill); err != nil {
		return nil, fmt.Errorf("failed to save skill: %w", err)
	}

	return &AddTeamSkillResponse{Skill: skill}, nil
}

// UpdateTeamSkillRequest ?
type UpdateTeamSkillRequest struct {
	SkillID  uuid.UUID         `json:"skill_id" validate:"required"`
	Level    *domain.SkillLevel `json:"level,omitempty"`
	Verified *bool             `json:"verified,omitempty"`
}

// UpdateTeamSkill ?
func (s *TeamService) UpdateTeamSkill(ctx context.Context, req *UpdateTeamSkillRequest) error {
	// ?
	if err := s.teamRepo.UpdateSkill(ctx, req.SkillID, req.Level, req.Verified); err != nil {
		return fmt.Errorf("failed to update skill: %w", err)
	}

	return nil
}

// RemoveTeamSkillRequest ?
type RemoveTeamSkillRequest struct {
	SkillID uuid.UUID `json:"skill_id" validate:"required"`
}

// RemoveTeamSkill ?
func (s *TeamService) RemoveTeamSkill(ctx context.Context, req *RemoveTeamSkillRequest) error {
	// ?
	if err := s.teamRepo.RemoveSkill(ctx, req.SkillID); err != nil {
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	return nil
}

// ==========  ==========

// GetTeamStatisticsRequest 
type GetTeamStatisticsRequest struct {
	TeamID         *uuid.UUID `json:"team_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// GetTeamStatisticsResponse 
type GetTeamStatisticsResponse struct {
	Statistics *domain.TeamStatistics `json:"statistics"`
}

// GetTeamStatistics 
func (s *TeamService) GetTeamStatistics(ctx context.Context, req *GetTeamStatisticsRequest) (*GetTeamStatisticsResponse, error) {
	stats, err := s.teamRepo.GetTeamStatistics(ctx, req.TeamID, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team statistics: %w", err)
	}

	return &GetTeamStatisticsResponse{Statistics: stats}, nil
}

// ==========  ==========

// GetTeamPerformanceRequest 
type GetTeamPerformanceRequest struct {
	TeamID    uuid.UUID  `json:"team_id" validate:"required"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// GetTeamPerformanceResponse 
type GetTeamPerformanceResponse struct {
	Performance *domain.TeamPerformanceReport `json:"performance"`
}

// GetTeamPerformance 
func (s *TeamService) GetTeamPerformance(ctx context.Context, req *GetTeamPerformanceRequest) (*GetTeamPerformanceResponse, error) {
	// 
	perfReq := &domain.TeamPerformanceRequest{
		TeamID:    req.TeamID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	// 
	performance, err := s.performanceSvc.AnalyzeTeamPerformance(ctx, perfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze team performance: %w", err)
	}

	return &GetTeamPerformanceResponse{Performance: performance}, nil
}

// ==========  ==========

// GetTeamWorkloadRequest 
type GetTeamWorkloadRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
}

// GetTeamWorkloadResponse 
type GetTeamWorkloadResponse struct {
	Workload *domain.TeamWorkloadReport `json:"workload"`
}

// GetTeamWorkload 
func (s *TeamService) GetTeamWorkload(ctx context.Context, req *GetTeamWorkloadRequest) (*GetTeamWorkloadResponse, error) {
	// 
	workloadReq := &domain.TeamWorkloadRequest{
		TeamID: req.TeamID,
	}

	// 
	workload, err := s.allocationSvc.AnalyzeTeamWorkload(ctx, workloadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze team workload: %w", err)
	}

	return &GetTeamWorkloadResponse{Workload: workload}, nil
}

// OptimizeTeamWorkloadRequest 
type OptimizeTeamWorkloadRequest struct {
	TeamID   uuid.UUID                      `json:"team_id" validate:"required"`
	Strategy domain.WorkloadOptimizationStrategy `json:"strategy" validate:"required"`
}

// OptimizeTeamWorkloadResponse 
type OptimizeTeamWorkloadResponse struct {
	Optimization *domain.WorkloadOptimizationResult `json:"optimization"`
}

// OptimizeTeamWorkload 
func (s *TeamService) OptimizeTeamWorkload(ctx context.Context, req *OptimizeTeamWorkloadRequest) (*OptimizeTeamWorkloadResponse, error) {
	// 
	optimizationReq := &domain.WorkloadOptimizationRequest{
		TeamID:   req.TeamID,
		Strategy: req.Strategy,
	}

	// 
	optimization, err := s.allocationSvc.OptimizeWorkload(ctx, optimizationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize team workload: %w", err)
	}

	return &OptimizeTeamWorkloadResponse{Optimization: optimization}, nil
}

