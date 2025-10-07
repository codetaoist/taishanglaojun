package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task-management/internal/domain"
)

// TeamService 团队应用服务
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

// NewTeamService 创建团队服务实例
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

// ========== 团队CRUD操作 ==========

// CreateTeamRequest 创建团队请求
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

// CreateTeamResponse 创建团队响应
type CreateTeamResponse struct {
	Team *domain.Team `json:"team"`
}

// CreateTeam 创建团队
func (s *TeamService) CreateTeam(ctx context.Context, req *CreateTeamRequest) (*CreateTeamResponse, error) {
	// 创建团队
	team, err := domain.NewTeam(
		req.Name,
		req.Description,
		req.OrganizationID,
		req.LeaderID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// 设置可选字段
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

	// 自动添加团队领导为成员
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

	// 保存团队
	if err := s.teamRepo.Save(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to save team: %w", err)
	}

	// 发布领域事件
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 发送通知
	notification := &domain.TeamNotification{
		Type:     domain.NotificationTypeTeamCreated,
		TeamID:   team.ID,
		UserID:   req.LeaderID,
		Title:    "团队创建成功",
		Message:  fmt.Sprintf("团队 %s 已成功创建，您被设为团队领导", team.Name),
		Metadata: map[string]interface{}{"team_id": team.ID.String()},
	}
	if err := s.notificationSvc.SendTeamNotification(ctx, notification); err != nil {
		fmt.Printf("failed to send notification: %v\n", err)
	}

	team.ClearDomainEvents()

	return &CreateTeamResponse{Team: team}, nil
}

// GetTeamRequest 获取团队请求
type GetTeamRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
}

// GetTeamResponse 获取团队响应
type GetTeamResponse struct {
	Team *domain.Team `json:"team"`
}

// GetTeam 获取团队
func (s *TeamService) GetTeam(ctx context.Context, req *GetTeamRequest) (*GetTeamResponse, error) {
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	return &GetTeamResponse{Team: team}, nil
}

// UpdateTeamRequest 更新团队请求
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

// UpdateTeamResponse 更新团队响应
type UpdateTeamResponse struct {
	Team *domain.Team `json:"team"`
}

// UpdateTeam 更新团队
func (s *TeamService) UpdateTeam(ctx context.Context, req *UpdateTeamRequest) (*UpdateTeamResponse, error) {
	// 获取现有团队
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 更新字段
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

	// 保存更新
	if err := s.teamRepo.Update(ctx, team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	// 发布领域事件
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	team.ClearDomainEvents()

	return &UpdateTeamResponse{Team: team}, nil
}

// DeleteTeamRequest 删除团队请求
type DeleteTeamRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
}

// DeleteTeam 删除团队
func (s *TeamService) DeleteTeam(ctx context.Context, req *DeleteTeamRequest) error {
	// 检查团队是否存在
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// 检查团队是否可以删除
	if team.Status == domain.TeamStatusActive {
		return fmt.Errorf("cannot delete active team")
	}

	// 删除团队
	if err := s.teamRepo.Delete(ctx, req.TeamID); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	return nil
}

// ========== 团队查询操作 ==========

// ListTeamsRequest 列表团队请求
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

// ListTeamsResponse 列表团队响应
type ListTeamsResponse struct {
	Teams   []*domain.Team `json:"teams"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}

// ListTeams 列表团队
func (s *TeamService) ListTeams(ctx context.Context, req *ListTeamsRequest) (*ListTeamsResponse, error) {
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

	// 构建过滤器
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

	// 根据不同条件选择查询方法
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

	// 获取总数
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

// SearchTeamsRequest 搜索团队请求
type SearchTeamsRequest struct {
	Query   string                 `json:"query" validate:"required"`
	Filters map[string]interface{} `json:"filters,omitempty"`
	Limit   int                    `json:"limit" validate:"min=1,max=100"`
	Offset  int                    `json:"offset" validate:"min=0"`
}

// SearchTeamsResponse 搜索团队响应
type SearchTeamsResponse struct {
	Teams   []*domain.Team `json:"teams"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
}

// SearchTeams 搜索团队
func (s *TeamService) SearchTeams(ctx context.Context, req *SearchTeamsRequest) (*SearchTeamsResponse, error) {
	teams, err := s.teamRepo.SearchTeams(ctx, req.Query, req.Filters, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search teams: %w", err)
	}

	// 获取总数（这里简化处理，实际应该根据搜索条件计算）
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

// ========== 团队成员管理 ==========

// AddTeamMemberRequest 添加团队成员请求
type AddTeamMemberRequest struct {
	TeamID       uuid.UUID                `json:"team_id" validate:"required"`
	UserID       uuid.UUID                `json:"user_id" validate:"required"`
	Role         domain.TeamMemberRole    `json:"role" validate:"required"`
	Availability *float64                 `json:"availability,omitempty"`
	Skills       []string                 `json:"skills,omitempty"`
}

// AddTeamMemberResponse 添加团队成员响应
type AddTeamMemberResponse struct {
	Member *domain.TeamMember `json:"member"`
}

// AddTeamMember 添加团队成员
func (s *TeamService) AddTeamMember(ctx context.Context, req *AddTeamMemberRequest) (*AddTeamMemberResponse, error) {
	// 获取团队
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 创建成员
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

	// 添加成员
	if err := team.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// 保存成员
	if err := s.teamRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to save member: %w", err)
	}

	// 添加技能
	if len(req.Skills) > 0 {
		for _, skillName := range req.Skills {
			skill := &domain.TeamSkill{
				ID:       uuid.New(),
				TeamID:   req.TeamID,
				UserID:   req.UserID,
				Skill:    skillName,
				Level:    domain.SkillLevelIntermediate, // 默认中级
				Verified: false,
			}
			if err := s.teamRepo.AddSkill(ctx, skill); err != nil {
				fmt.Printf("failed to add skill %s: %v\n", skillName, err)
			}
		}
	}

	// 发布领域事件
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	// 发送通知
	notification := &domain.TeamNotification{
		Type:     domain.NotificationTypeTeamMemberAdded,
		TeamID:   team.ID,
		UserID:   req.UserID,
		Title:    "团队邀请",
		Message:  fmt.Sprintf("您已被邀请加入团队: %s", team.Name),
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

// RemoveTeamMemberRequest 移除团队成员请求
type RemoveTeamMemberRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// RemoveTeamMember 移除团队成员
func (s *TeamService) RemoveTeamMember(ctx context.Context, req *RemoveTeamMemberRequest) error {
	// 获取团队
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// 移除成员
	if err := team.RemoveMember(req.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// 保存更改
	if err := s.teamRepo.RemoveMember(ctx, req.TeamID, req.UserID); err != nil {
		return fmt.Errorf("failed to save member removal: %w", err)
	}

	// 发布领域事件
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	team.ClearDomainEvents()

	return nil
}

// UpdateTeamMemberRequest 更新团队成员请求
type UpdateTeamMemberRequest struct {
	TeamID       uuid.UUID             `json:"team_id" validate:"required"`
	UserID       uuid.UUID             `json:"user_id" validate:"required"`
	Role         *domain.TeamMemberRole `json:"role,omitempty"`
	IsAvailable  *bool                 `json:"is_available,omitempty"`
	Availability *float64              `json:"availability,omitempty"`
}

// UpdateTeamMember 更新团队成员
func (s *TeamService) UpdateTeamMember(ctx context.Context, req *UpdateTeamMemberRequest) error {
	// 获取团队
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return fmt.Errorf("failed to find team: %w", err)
	}

	// 更新成员信息
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

	// 保存更改
	if err := s.teamRepo.UpdateMember(ctx, req.TeamID, req.UserID, req.Role, req.IsAvailable, req.Availability); err != nil {
		return fmt.Errorf("failed to save member update: %w", err)
	}

	// 发布领域事件
	for _, event := range team.GetDomainEvents() {
		if err := s.eventRepo.Save(ctx, event); err != nil {
			fmt.Printf("failed to save event: %v\n", err)
		}
	}

	team.ClearDomainEvents()

	return nil
}

// ========== 团队技能管理 ==========

// AddTeamSkillRequest 添加团队技能请求
type AddTeamSkillRequest struct {
	TeamID   uuid.UUID           `json:"team_id" validate:"required"`
	UserID   uuid.UUID           `json:"user_id" validate:"required"`
	Skill    string              `json:"skill" validate:"required"`
	Level    domain.SkillLevel   `json:"level" validate:"required"`
	Verified bool                `json:"verified"`
}

// AddTeamSkillResponse 添加团队技能响应
type AddTeamSkillResponse struct {
	Skill *domain.TeamSkill `json:"skill"`
}

// AddTeamSkill 添加团队技能
func (s *TeamService) AddTeamSkill(ctx context.Context, req *AddTeamSkillRequest) (*AddTeamSkillResponse, error) {
	// 验证团队和成员是否存在
	team, err := s.teamRepo.FindByID(ctx, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to find team: %w", err)
	}

	// 检查用户是否是团队成员
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

	// 创建技能
	skill := &domain.TeamSkill{
		ID:       uuid.New(),
		TeamID:   req.TeamID,
		UserID:   req.UserID,
		Skill:    req.Skill,
		Level:    req.Level,
		Verified: req.Verified,
	}

	// 保存技能
	if err := s.teamRepo.AddSkill(ctx, skill); err != nil {
		return nil, fmt.Errorf("failed to save skill: %w", err)
	}

	return &AddTeamSkillResponse{Skill: skill}, nil
}

// UpdateTeamSkillRequest 更新团队技能请求
type UpdateTeamSkillRequest struct {
	SkillID  uuid.UUID         `json:"skill_id" validate:"required"`
	Level    *domain.SkillLevel `json:"level,omitempty"`
	Verified *bool             `json:"verified,omitempty"`
}

// UpdateTeamSkill 更新团队技能
func (s *TeamService) UpdateTeamSkill(ctx context.Context, req *UpdateTeamSkillRequest) error {
	// 更新技能
	if err := s.teamRepo.UpdateSkill(ctx, req.SkillID, req.Level, req.Verified); err != nil {
		return fmt.Errorf("failed to update skill: %w", err)
	}

	return nil
}

// RemoveTeamSkillRequest 移除团队技能请求
type RemoveTeamSkillRequest struct {
	SkillID uuid.UUID `json:"skill_id" validate:"required"`
}

// RemoveTeamSkill 移除团队技能
func (s *TeamService) RemoveTeamSkill(ctx context.Context, req *RemoveTeamSkillRequest) error {
	// 移除技能
	if err := s.teamRepo.RemoveSkill(ctx, req.SkillID); err != nil {
		return fmt.Errorf("failed to remove skill: %w", err)
	}

	return nil
}

// ========== 团队统计操作 ==========

// GetTeamStatisticsRequest 获取团队统计请求
type GetTeamStatisticsRequest struct {
	TeamID         *uuid.UUID `json:"team_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
}

// GetTeamStatisticsResponse 获取团队统计响应
type GetTeamStatisticsResponse struct {
	Statistics *domain.TeamStatistics `json:"statistics"`
}

// GetTeamStatistics 获取团队统计
func (s *TeamService) GetTeamStatistics(ctx context.Context, req *GetTeamStatisticsRequest) (*GetTeamStatisticsResponse, error) {
	stats, err := s.teamRepo.GetTeamStatistics(ctx, req.TeamID, req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team statistics: %w", err)
	}

	return &GetTeamStatisticsResponse{Statistics: stats}, nil
}

// ========== 团队性能分析 ==========

// GetTeamPerformanceRequest 获取团队性能请求
type GetTeamPerformanceRequest struct {
	TeamID    uuid.UUID  `json:"team_id" validate:"required"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}

// GetTeamPerformanceResponse 获取团队性能响应
type GetTeamPerformanceResponse struct {
	Performance *domain.TeamPerformanceReport `json:"performance"`
}

// GetTeamPerformance 获取团队性能
func (s *TeamService) GetTeamPerformance(ctx context.Context, req *GetTeamPerformanceRequest) (*GetTeamPerformanceResponse, error) {
	// 构建性能分析请求
	perfReq := &domain.TeamPerformanceRequest{
		TeamID:    req.TeamID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	// 分析团队性能
	performance, err := s.performanceSvc.AnalyzeTeamPerformance(ctx, perfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze team performance: %w", err)
	}

	return &GetTeamPerformanceResponse{Performance: performance}, nil
}

// ========== 团队任务分配 ==========

// GetTeamWorkloadRequest 获取团队工作负载请求
type GetTeamWorkloadRequest struct {
	TeamID uuid.UUID `json:"team_id" validate:"required"`
}

// GetTeamWorkloadResponse 获取团队工作负载响应
type GetTeamWorkloadResponse struct {
	Workload *domain.TeamWorkloadReport `json:"workload"`
}

// GetTeamWorkload 获取团队工作负载
func (s *TeamService) GetTeamWorkload(ctx context.Context, req *GetTeamWorkloadRequest) (*GetTeamWorkloadResponse, error) {
	// 构建工作负载分析请求
	workloadReq := &domain.TeamWorkloadRequest{
		TeamID: req.TeamID,
	}

	// 分析团队工作负载
	workload, err := s.allocationSvc.AnalyzeTeamWorkload(ctx, workloadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze team workload: %w", err)
	}

	return &GetTeamWorkloadResponse{Workload: workload}, nil
}

// OptimizeTeamWorkloadRequest 优化团队工作负载请求
type OptimizeTeamWorkloadRequest struct {
	TeamID   uuid.UUID                      `json:"team_id" validate:"required"`
	Strategy domain.WorkloadOptimizationStrategy `json:"strategy" validate:"required"`
}

// OptimizeTeamWorkloadResponse 优化团队工作负载响应
type OptimizeTeamWorkloadResponse struct {
	Optimization *domain.WorkloadOptimizationResult `json:"optimization"`
}

// OptimizeTeamWorkload 优化团队工作负载
func (s *TeamService) OptimizeTeamWorkload(ctx context.Context, req *OptimizeTeamWorkloadRequest) (*OptimizeTeamWorkloadResponse, error) {
	// 构建工作负载优化请求
	optimizationReq := &domain.WorkloadOptimizationRequest{
		TeamID:   req.TeamID,
		Strategy: req.Strategy,
	}

	// 优化团队工作负载
	optimization, err := s.allocationSvc.OptimizeWorkload(ctx, optimizationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize team workload: %w", err)
	}

	return &OptimizeTeamWorkloadResponse{Optimization: optimization}, nil
}