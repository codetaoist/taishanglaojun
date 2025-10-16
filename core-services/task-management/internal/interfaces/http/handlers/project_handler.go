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

// ProjectHandler 项目HTTP处理?
type ProjectHandler struct {
	projectService *application.ProjectService
}

// NewProjectHandler 创建项目处理?
func NewProjectHandler(projectService *application.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateProjectRequest
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
	serviceReq := &application.CreateProjectRequest{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: req.OrganizationID,
		ManagerID:      req.ManagerID,
		Status:         req.Status,
		Priority:       req.Priority,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		Budget:         req.Budget,
		Tags:           req.Tags,
		Labels:         req.Labels,
	}

	// 调用应用服务
	resp, err := h.projectService.CreateProject(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ProjectResponse{
		ID:             resp.Project.ID,
		Name:           resp.Project.Name,
		Description:    resp.Project.Description,
		OrganizationID: resp.Project.OrganizationID,
		ManagerID:      resp.Project.ManagerID,
		Status:         resp.Project.Status,
		Priority:       resp.Project.Priority,
		StartDate:      resp.Project.StartDate,
		EndDate:        resp.Project.EndDate,
		Budget:         resp.Project.Budget,
		Tags:           resp.Project.Tags,
		Labels:         resp.Project.Labels,
		CreatedAt:      resp.Project.CreatedAt,
		UpdatedAt:      resp.Project.UpdatedAt,
	})
}

// GetProject 获取项目
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.projectService.GetProject(r.Context(), &application.GetProjectRequest{
		ProjectID: projectID,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProjectResponse{
		ID:             resp.Project.ID,
		Name:           resp.Project.Name,
		Description:    resp.Project.Description,
		OrganizationID: resp.Project.OrganizationID,
		ManagerID:      resp.Project.ManagerID,
		Status:         resp.Project.Status,
		Priority:       resp.Project.Priority,
		StartDate:      resp.Project.StartDate,
		EndDate:        resp.Project.EndDate,
		Budget:         resp.Project.Budget,
		Tags:           resp.Project.Tags,
		Labels:         resp.Project.Labels,
		CreatedAt:      resp.Project.CreatedAt,
		UpdatedAt:      resp.Project.UpdatedAt,
	})
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 转换为应用服务请?
	serviceReq := &application.UpdateProjectRequest{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerID,
		Status:      req.Status,
		Priority:    req.Priority,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Budget:      req.Budget,
		Tags:        req.Tags,
		Labels:      req.Labels,
	}

	// 调用应用服务
	resp, err := h.projectService.UpdateProject(r.Context(), serviceReq)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProjectResponse{
		ID:             resp.Project.ID,
		Name:           resp.Project.Name,
		Description:    resp.Project.Description,
		OrganizationID: resp.Project.OrganizationID,
		ManagerID:      resp.Project.ManagerID,
		Status:         resp.Project.Status,
		Priority:       resp.Project.Priority,
		StartDate:      resp.Project.StartDate,
		EndDate:        resp.Project.EndDate,
		Budget:         resp.Project.Budget,
		Tags:           resp.Project.Tags,
		Labels:         resp.Project.Labels,
		CreatedAt:      resp.Project.CreatedAt,
		UpdatedAt:      resp.Project.UpdatedAt,
	})
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.projectService.DeleteProject(r.Context(), &application.DeleteProjectRequest{
		ProjectID: projectID,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusNoContent)
}

// ListProjects 列出项目
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	query := r.URL.Query()
	
	var organizationID *uuid.UUID
	if orgIDStr := query.Get("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			organizationID = &id
		}
	}

	var managerID *uuid.UUID
	if managerIDStr := query.Get("manager_id"); managerIDStr != "" {
		if id, err := uuid.Parse(managerIDStr); err == nil {
			managerID = &id
		}
	}

	var status *domain.ProjectStatus
	if statusStr := query.Get("status"); statusStr != "" {
		s := domain.ProjectStatus(statusStr)
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
	resp, err := h.projectService.ListProjects(r.Context(), &application.ListProjectsRequest{
		OrganizationID: organizationID,
		ManagerID:      managerID,
		Status:         status,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 转换响应
	projects := make([]ProjectResponse, len(resp.Projects))
	for i, project := range resp.Projects {
		projects[i] = ProjectResponse{
			ID:             project.ID,
			Name:           project.Name,
			Description:    project.Description,
			OrganizationID: project.OrganizationID,
			ManagerID:      project.ManagerID,
			Status:         project.Status,
			Priority:       project.Priority,
			StartDate:      project.StartDate,
			EndDate:        project.EndDate,
			Budget:         project.Budget,
			Tags:           project.Tags,
			Labels:         project.Labels,
			CreatedAt:      project.CreatedAt,
			UpdatedAt:      project.UpdatedAt,
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListProjectsResponse{
		Projects: projects,
		Total:    resp.Total,
		Limit:    limit,
		Offset:   offset,
	})
}

// SearchProjects 搜索项目
func (h *ProjectHandler) SearchProjects(w http.ResponseWriter, r *http.Request) {
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
	resp, err := h.projectService.SearchProjects(r.Context(), &application.SearchProjectsRequest{
		Keyword: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 转换响应
	projects := make([]ProjectResponse, len(resp.Projects))
	for i, project := range resp.Projects {
		projects[i] = ProjectResponse{
			ID:             project.ID,
			Name:           project.Name,
			Description:    project.Description,
			OrganizationID: project.OrganizationID,
			ManagerID:      project.ManagerID,
			Status:         project.Status,
			Priority:       project.Priority,
			StartDate:      project.StartDate,
			EndDate:        project.EndDate,
			Budget:         project.Budget,
			Tags:           project.Tags,
			Labels:         project.Labels,
			CreatedAt:      project.CreatedAt,
			UpdatedAt:      project.UpdatedAt,
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListProjectsResponse{
		Projects: projects,
		Total:    resp.Total,
		Limit:    limit,
		Offset:   offset,
	})
}

// AddProjectMember 添加项目成员
func (h *ProjectHandler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	var req AddProjectMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.projectService.AddProjectMember(r.Context(), &application.AddProjectMemberRequest{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
		AddedBy:   req.AddedBy,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project member added successfully",
	})
}

// RemoveProjectMember 移除项目成员
func (h *ProjectHandler) RemoveProjectMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req RemoveProjectMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.projectService.RemoveProjectMember(r.Context(), &application.RemoveProjectMemberRequest{
		ProjectID: projectID,
		UserID:    userID,
		RemovedBy: req.RemovedBy,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project member removed successfully",
	})
}

// UpdateProjectMemberRole 更新项目成员角色
func (h *ProjectHandler) UpdateProjectMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdateProjectMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.projectService.UpdateProjectMemberRole(r.Context(), &application.UpdateProjectMemberRoleRequest{
		ProjectID: projectID,
		UserID:    userID,
		Role:      req.Role,
		UpdatedBy: req.UpdatedBy,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Project member role updated successfully",
	})
}

// GetProjectStatistics 获取项目统计
func (h *ProjectHandler) GetProjectStatistics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.projectService.GetProjectStatistics(r.Context(), &application.GetProjectStatisticsRequest{
		ProjectID: projectID,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Statistics)
}

// GenerateProjectSchedule 生成项目进度计划
func (h *ProjectHandler) GenerateProjectSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.projectService.GenerateProjectSchedule(r.Context(), &application.GenerateProjectScheduleRequest{
		ProjectID: projectID,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Schedule)
}

// GetProjectPerformance 获取项目绩效分析
func (h *ProjectHandler) GetProjectPerformance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.projectService.GetProjectPerformance(r.Context(), &application.GetProjectPerformanceRequest{
		ProjectID: projectID,
	})
	if err != nil {
		if err == domain.ErrProjectNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Performance)
}

