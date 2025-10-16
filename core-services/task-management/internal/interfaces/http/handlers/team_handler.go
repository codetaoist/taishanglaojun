package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/google/uuid"

	"task-management/internal/application"
	"task-management/internal/domain"
)

// TeamHandler 团队HTTP处理?
type TeamHandler struct {
	teamService *application.TeamService
}

// NewTeamHandler 创建团队处理?
func NewTeamHandler(teamService *application.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

// CreateTeam 创建团队
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 转换为应用服务请?
	serviceReq := &application.CreateTeamRequest{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: req.OrganizationID,
		LeaderID:       req.LeaderID,
		Type:           req.Type,
		Status:         req.Status,
		MaxMembers:     req.MaxMembers,
		Tags:           req.Tags,
		Labels:         req.Labels,
	}

	// 调用应用服务
	resp, err := h.teamService.CreateTeam(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TeamResponse{
		ID:             resp.Team.ID,
		Name:           resp.Team.Name,
		Description:    resp.Team.Description,
		OrganizationID: resp.Team.OrganizationID,
		LeaderID:       resp.Team.LeaderID,
		Type:           resp.Team.Type,
		Status:         resp.Team.Status,
		MaxMembers:     resp.Team.MaxMembers,
		Tags:           resp.Team.Tags,
		Labels:         resp.Team.Labels,
		CreatedAt:      resp.Team.CreatedAt,
		UpdatedAt:      resp.Team.UpdatedAt,
	})
}

// GetTeam 获取团队
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.teamService.GetTeam(r.Context(), &application.GetTeamRequest{
		TeamID: teamID,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TeamResponse{
		ID:             resp.Team.ID,
		Name:           resp.Team.Name,
		Description:    resp.Team.Description,
		OrganizationID: resp.Team.OrganizationID,
		LeaderID:       resp.Team.LeaderID,
		Type:           resp.Team.Type,
		Status:         resp.Team.Status,
		MaxMembers:     resp.Team.MaxMembers,
		Tags:           resp.Team.Tags,
		Labels:         resp.Team.Labels,
		CreatedAt:      resp.Team.CreatedAt,
		UpdatedAt:      resp.Team.UpdatedAt,
	})
}

// UpdateTeam 更新团队
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var req UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 转换为应用服务请?
	serviceReq := &application.UpdateTeamRequest{
		TeamID:      teamID,
		Name:        req.Name,
		Description: req.Description,
		LeaderID:    req.LeaderID,
		Type:        req.Type,
		Status:      req.Status,
		MaxMembers:  req.MaxMembers,
		Tags:        req.Tags,
		Labels:      req.Labels,
	}

	// 调用应用服务
	resp, err := h.teamService.UpdateTeam(r.Context(), serviceReq)
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TeamResponse{
		ID:             resp.Team.ID,
		Name:           resp.Team.Name,
		Description:    resp.Team.Description,
		OrganizationID: resp.Team.OrganizationID,
		LeaderID:       resp.Team.LeaderID,
		Type:           resp.Team.Type,
		Status:         resp.Team.Status,
		MaxMembers:     resp.Team.MaxMembers,
		Tags:           resp.Team.Tags,
		Labels:         resp.Team.Labels,
		CreatedAt:      resp.Team.CreatedAt,
		UpdatedAt:      resp.Team.UpdatedAt,
	})
}

// DeleteTeam 删除团队
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.teamService.DeleteTeam(r.Context(), &application.DeleteTeamRequest{
		TeamID: teamID,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusNoContent)
}

// ListTeams 列出团队
func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	query := r.URL.Query()
	
	var organizationID *uuid.UUID
	if orgIDStr := query.Get("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			organizationID = &id
		}
	}

	var leaderID *uuid.UUID
	if leaderIDStr := query.Get("leader_id"); leaderIDStr != "" {
		if id, err := uuid.Parse(leaderIDStr); err == nil {
			leaderID = &id
		}
	}

	var status *domain.TeamStatus
	if statusStr := query.Get("status"); statusStr != "" {
		s := domain.TeamStatus(statusStr)
		status = &s
	}

	limit := 20
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 调用应用服务
	resp, err := h.teamService.ListTeams(r.Context(), &application.ListTeamsRequest{
		OrganizationID: organizationID,
		LeaderID:       leaderID,
		Status:         status,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 转换响应
	teams := make([]TeamResponse, len(resp.Teams))
	for i, team := range resp.Teams {
		teams[i] = TeamResponse{
			ID:             team.ID,
			Name:           team.Name,
			Description:    team.Description,
			OrganizationID: team.OrganizationID,
			LeaderID:       team.LeaderID,
			Type:           team.Type,
			Status:         team.Status,
			MaxMembers:     team.MaxMembers,
			Tags:           team.Tags,
			Labels:         team.Labels,
			CreatedAt:      team.CreatedAt,
			UpdatedAt:      team.UpdatedAt,
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListTeamsResponse{
		Teams:  teams,
		Total:  resp.Total,
		Limit:  limit,
		Offset: offset,
	})
}

// SearchTeams 搜索团队
func (h *TeamHandler) SearchTeams(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	keyword := query.Get("q")
	if keyword == "" {
		http.Error(w, "Search keyword is required", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 调用应用服务
	resp, err := h.teamService.SearchTeams(r.Context(), &application.SearchTeamsRequest{
		Keyword: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 转换响应
	teams := make([]TeamResponse, len(resp.Teams))
	for i, team := range resp.Teams {
		teams[i] = TeamResponse{
			ID:             team.ID,
			Name:           team.Name,
			Description:    team.Description,
			OrganizationID: team.OrganizationID,
			LeaderID:       team.LeaderID,
			Type:           team.Type,
			Status:         team.Status,
			MaxMembers:     team.MaxMembers,
			Tags:           team.Tags,
			Labels:         team.Labels,
			CreatedAt:      team.CreatedAt,
			UpdatedAt:      team.UpdatedAt,
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListTeamsResponse{
		Teams:  teams,
		Total:  resp.Total,
		Limit:  limit,
		Offset: offset,
	})
}

// AddTeamMember 添加团队成员
func (h *TeamHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var req AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.teamService.AddTeamMember(r.Context(), &application.AddTeamMemberRequest{
		TeamID:  teamID,
		UserID:  req.UserID,
		Role:    req.Role,
		AddedBy: req.AddedBy,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Team member added successfully",
	})
}

// RemoveTeamMember 移除团队成员
func (h *TeamHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req RemoveTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.teamService.RemoveTeamMember(r.Context(), &application.RemoveTeamMemberRequest{
		TeamID:    teamID,
		UserID:    userID,
		RemovedBy: req.RemovedBy,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Team member removed successfully",
	})
}

// UpdateTeamMember 更新团队成员
func (h *TeamHandler) UpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.teamService.UpdateTeamMember(r.Context(), &application.UpdateTeamMemberRequest{
		TeamID:    teamID,
		UserID:    userID,
		Role:      req.Role,
		UpdatedBy: req.UpdatedBy,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Team member updated successfully",
	})
}

// GetTeamStatistics 获取团队统计
func (h *TeamHandler) GetTeamStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.teamService.GetTeamStatistics(r.Context(), &application.GetTeamStatisticsRequest{
		TeamID: teamID,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Statistics)
}

// GetTeamPerformance 获取团队绩效分析
func (h *TeamHandler) GetTeamPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.teamService.GetTeamPerformance(r.Context(), &application.GetTeamPerformanceRequest{
		TeamID: teamID,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Performance)
}

// GetTeamWorkload 获取团队工作负载
func (h *TeamHandler) GetTeamWorkload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.teamService.GetTeamWorkload(r.Context(), &application.GetTeamWorkloadRequest{
		TeamID: teamID,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Workload)
}

// OptimizeTeamWorkload 优化团队工作负载
func (h *TeamHandler) OptimizeTeamWorkload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.teamService.OptimizeTeamWorkload(r.Context(), &application.OptimizeTeamWorkloadRequest{
		TeamID: teamID,
	})
	if err != nil {
		if err == domain.ErrTeamNotFound {
			http.Error(w, "Team not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Optimization)
}

