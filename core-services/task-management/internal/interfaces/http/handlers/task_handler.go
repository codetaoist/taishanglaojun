package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"

	"task-management/internal/application"
	"task-management/internal/domain"
)

// TaskHandler 任务HTTP处理器
type TaskHandler struct {
	taskService *application.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *application.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 转换为应用服务请求
	serviceReq := &application.CreateTaskRequest{
		Title:         req.Title,
		Description:   req.Description,
		ProjectID:     req.ProjectID,
		AssigneeID:    req.AssigneeID,
		CreatorID:     req.CreatorID,
		Type:          req.Type,
		Priority:      req.Priority,
		Complexity:    req.Complexity,
		EstimatedHours: req.EstimatedHours,
		DueDate:       req.DueDate,
		Tags:          req.Tags,
		Labels:        req.Labels,
		Dependencies:  req.Dependencies,
	}

	// 调用应用服务
	resp, err := h.taskService.CreateTask(r.Context(), serviceReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TaskResponse{
		ID:            resp.Task.ID,
		Title:         resp.Task.Title,
		Description:   resp.Task.Description,
		ProjectID:     resp.Task.ProjectID,
		AssigneeID:    resp.Task.AssigneeID,
		CreatorID:     resp.Task.CreatorID,
		Type:          resp.Task.Type,
		Status:        resp.Task.Status,
		Priority:      resp.Task.Priority,
		Complexity:    resp.Task.Complexity,
		EstimatedHours: resp.Task.EstimatedHours,
		ActualHours:   resp.Task.ActualHours,
		DueDate:       resp.Task.DueDate,
		CompletedAt:   resp.Task.CompletedAt,
		Tags:          resp.Task.Tags,
		Labels:        resp.Task.Labels,
		Dependencies:  resp.Task.Dependencies,
		CreatedAt:     resp.Task.CreatedAt,
		UpdatedAt:     resp.Task.UpdatedAt,
	})
}

// GetTask 获取任务
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.taskService.GetTask(r.Context(), &application.GetTaskRequest{
		TaskID: taskID,
	})
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:            resp.Task.ID,
		Title:         resp.Task.Title,
		Description:   resp.Task.Description,
		ProjectID:     resp.Task.ProjectID,
		AssigneeID:    resp.Task.AssigneeID,
		CreatorID:     resp.Task.CreatorID,
		Type:          resp.Task.Type,
		Status:        resp.Task.Status,
		Priority:      resp.Task.Priority,
		Complexity:    resp.Task.Complexity,
		EstimatedHours: resp.Task.EstimatedHours,
		ActualHours:   resp.Task.ActualHours,
		DueDate:       resp.Task.DueDate,
		CompletedAt:   resp.Task.CompletedAt,
		Tags:          resp.Task.Tags,
		Labels:        resp.Task.Labels,
		Dependencies:  resp.Task.Dependencies,
		CreatedAt:     resp.Task.CreatedAt,
		UpdatedAt:     resp.Task.UpdatedAt,
	})
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 转换为应用服务请求
	serviceReq := &application.UpdateTaskRequest{
		TaskID:        taskID,
		Title:         req.Title,
		Description:   req.Description,
		AssigneeID:    req.AssigneeID,
		Type:          req.Type,
		Status:        req.Status,
		Priority:      req.Priority,
		Complexity:    req.Complexity,
		EstimatedHours: req.EstimatedHours,
		ActualHours:   req.ActualHours,
		DueDate:       req.DueDate,
		Tags:          req.Tags,
		Labels:        req.Labels,
	}

	// 调用应用服务
	resp, err := h.taskService.UpdateTask(r.Context(), serviceReq)
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:            resp.Task.ID,
		Title:         resp.Task.Title,
		Description:   resp.Task.Description,
		ProjectID:     resp.Task.ProjectID,
		AssigneeID:    resp.Task.AssigneeID,
		CreatorID:     resp.Task.CreatorID,
		Type:          resp.Task.Type,
		Status:        resp.Task.Status,
		Priority:      resp.Task.Priority,
		Complexity:    resp.Task.Complexity,
		EstimatedHours: resp.Task.EstimatedHours,
		ActualHours:   resp.Task.ActualHours,
		DueDate:       resp.Task.DueDate,
		CompletedAt:   resp.Task.CompletedAt,
		Tags:          resp.Task.Tags,
		Labels:        resp.Task.Labels,
		Dependencies:  resp.Task.Dependencies,
		CreatedAt:     resp.Task.CreatedAt,
		UpdatedAt:     resp.Task.UpdatedAt,
	})
}

// DeleteTask 删除任务
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.taskService.DeleteTask(r.Context(), &application.DeleteTaskRequest{
		TaskID: taskID,
	})
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusNoContent)
}

// ListTasks 列出任务
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	query := r.URL.Query()
	
	var projectID *uuid.UUID
	if projectIDStr := query.Get("project_id"); projectIDStr != "" {
		if id, err := uuid.Parse(projectIDStr); err == nil {
			projectID = &id
		}
	}

	var assigneeID *uuid.UUID
	if assigneeIDStr := query.Get("assignee_id"); assigneeIDStr != "" {
		if id, err := uuid.Parse(assigneeIDStr); err == nil {
			assigneeID = &id
		}
	}

	var status *domain.TaskStatus
	if statusStr := query.Get("status"); statusStr != "" {
		s := domain.TaskStatus(statusStr)
		status = &s
	}

	var priority *domain.TaskPriority
	if priorityStr := query.Get("priority"); priorityStr != "" {
		p := domain.TaskPriority(priorityStr)
		priority = &p
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
	resp, err := h.taskService.ListTasks(r.Context(), &application.ListTasksRequest{
		ProjectID:  projectID,
		AssigneeID: assigneeID,
		Status:     status,
		Priority:   priority,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 转换响应
	tasks := make([]TaskResponse, len(resp.Tasks))
	for i, task := range resp.Tasks {
		tasks[i] = TaskResponse{
			ID:            task.ID,
			Title:         task.Title,
			Description:   task.Description,
			ProjectID:     task.ProjectID,
			AssigneeID:    task.AssigneeID,
			CreatorID:     task.CreatorID,
			Type:          task.Type,
			Status:        task.Status,
			Priority:      task.Priority,
			Complexity:    task.Complexity,
			EstimatedHours: task.EstimatedHours,
			ActualHours:   task.ActualHours,
			DueDate:       task.DueDate,
			CompletedAt:   task.CompletedAt,
			Tags:          task.Tags,
			Labels:        task.Labels,
			Dependencies:  task.Dependencies,
			CreatedAt:     task.CreatedAt,
			UpdatedAt:     task.UpdatedAt,
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListTasksResponse{
		Tasks: tasks,
		Total: resp.Total,
		Limit: limit,
		Offset: offset,
	})
}

// SearchTasks 搜索任务
func (h *TaskHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
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
	resp, err := h.taskService.SearchTasks(r.Context(), &application.SearchTasksRequest{
		Keyword: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 转换响应
	tasks := make([]TaskResponse, len(resp.Tasks))
	for i, task := range resp.Tasks {
		tasks[i] = TaskResponse{
			ID:            task.ID,
			Title:         task.Title,
			Description:   task.Description,
			ProjectID:     task.ProjectID,
			AssigneeID:    task.AssigneeID,
			CreatorID:     task.CreatorID,
			Type:          task.Type,
			Status:        task.Status,
			Priority:      task.Priority,
			Complexity:    task.Complexity,
			EstimatedHours: task.EstimatedHours,
			ActualHours:   task.ActualHours,
			DueDate:       task.DueDate,
			CompletedAt:   task.CompletedAt,
			Tags:          task.Tags,
			Labels:        task.Labels,
			Dependencies:  task.Dependencies,
			CreatedAt:     task.CreatedAt,
			UpdatedAt:     task.UpdatedAt,
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListTasksResponse{
		Tasks: tasks,
		Total: resp.Total,
		Limit: limit,
		Offset: offset,
	})
}

// AssignTask 分配任务
func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req AssignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.taskService.AssignTask(r.Context(), &application.AssignTaskRequest{
		TaskID:     taskID,
		AssigneeID: req.AssigneeID,
		AssignedBy: req.AssignedBy,
	})
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Task assigned successfully",
	})
}

// UnassignTask 取消分配任务
func (h *TaskHandler) UnassignTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req UnassignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	err = h.taskService.UnassignTask(r.Context(), &application.UnassignTaskRequest{
		TaskID:       taskID,
		UnassignedBy: req.UnassignedBy,
	})
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Task unassigned successfully",
	})
}

// AutoAssignTasks 自动分配任务
func (h *TaskHandler) AutoAssignTasks(w http.ResponseWriter, r *http.Request) {
	var req AutoAssignTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.taskService.AutoAssignTasks(r.Context(), &application.AutoAssignTasksRequest{
		ProjectID: req.ProjectID,
		TeamID:    req.TeamID,
		Strategy:  req.Strategy,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AutoAssignTasksResponse{
		AssignedTasks: resp.AssignedTasks,
		Summary:       resp.Summary,
	})
}

// GetTaskStatistics 获取任务统计
func (h *TaskHandler) GetTaskStatistics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	var projectID *uuid.UUID
	if projectIDStr := query.Get("project_id"); projectIDStr != "" {
		if id, err := uuid.Parse(projectIDStr); err == nil {
			projectID = &id
		}
	}

	var assigneeID *uuid.UUID
	if assigneeIDStr := query.Get("assignee_id"); assigneeIDStr != "" {
		if id, err := uuid.Parse(assigneeIDStr); err == nil {
			assigneeID = &id
		}
	}

	// 调用应用服务
	resp, err := h.taskService.GetTaskStatistics(r.Context(), &application.GetTaskStatisticsRequest{
		ProjectID:  projectID,
		AssigneeID: assigneeID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Statistics)
}

// AddTaskComment 添加任务评论
func (h *TaskHandler) AddTaskComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req AddTaskCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.taskService.AddTaskComment(r.Context(), &application.AddTaskCommentRequest{
		TaskID:   taskID,
		AuthorID: req.AuthorID,
		Content:  req.Content,
	})
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TaskCommentResponse{
		ID:        resp.Comment.ID,
		TaskID:    resp.Comment.TaskID,
		AuthorID:  resp.Comment.AuthorID,
		Content:   resp.Comment.Content,
		CreatedAt: resp.Comment.CreatedAt,
		UpdatedAt: resp.Comment.UpdatedAt,
	})
}

// AddTimeLog 添加时间记录
func (h *TaskHandler) AddTimeLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req AddTimeLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 调用应用服务
	resp, err := h.taskService.AddTimeLog(r.Context(), &application.AddTimeLogRequest{
		TaskID:      taskID,
		UserID:      req.UserID,
		Duration:    req.Duration,
		Description: req.Description,
		LoggedAt:    req.LoggedAt,
	})
	if err != nil {
		if err == domain.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TimeLogResponse{
		ID:          resp.TimeLog.ID,
		TaskID:      resp.TimeLog.TaskID,
		UserID:      resp.TimeLog.UserID,
		Duration:    resp.TimeLog.Duration,
		Description: resp.TimeLog.Description,
		LoggedAt:    resp.TimeLog.LoggedAt,
		CreatedAt:   resp.TimeLog.CreatedAt,
	})
}