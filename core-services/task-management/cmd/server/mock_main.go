package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskStatus ?
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusOnHold     TaskStatus = "on_hold"
	TaskStatusOverdue    TaskStatus = "overdue"
)

// TaskPriority ?
type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"
)

// TaskType 
type TaskType string

const (
	TaskTypeDevelopment   TaskType = "development"
	TaskTypeBug          TaskType = "bug"
	TaskTypeFeature      TaskType = "feature"
	TaskTypeResearch     TaskType = "research"
	TaskTypeMaintenance  TaskType = "maintenance"
	TaskTypeReview       TaskType = "review"
	TaskTypeTesting      TaskType = "testing"
	TaskTypeDocumentation TaskType = "documentation"
)

// Task 
type Task struct {
	ID             string       `json:"id"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Status         TaskStatus   `json:"status"`
	Priority       TaskPriority `json:"priority"`
	Type           TaskType     `json:"type"`
	ProjectID      string       `json:"project_id"`
	TeamID         *string      `json:"team_id,omitempty"`
	CreatorID      string       `json:"creator_id"`
	AssigneeID     *string      `json:"assignee_id,omitempty"`
	ReviewerID     *string      `json:"reviewer_id,omitempty"`
	EstimatedHours *float64     `json:"estimated_hours,omitempty"`
	ActualHours    *float64     `json:"actual_hours,omitempty"`
	StartDate      *time.Time   `json:"start_date,omitempty"`
	DueDate        *time.Time   `json:"due_date,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	Progress       float64      `json:"progress"`
	Tags           []string     `json:"tags"`
	Labels         map[string]string `json:"labels"`
}

// Project 
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	OwnerID     string    `json:"owner_id"`
	TeamID      string    `json:"team_id"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Progress    float64   `json:"progress"`
	TaskCount   int       `json:"task_count"`
}

// Team 
type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LeaderID    string    `json:"leader_id"`
	MemberIDs   []string  `json:"member_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

// TaskComment 
type TaskComment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AuthorID  string    `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskTimeLog 
type TaskTimeLog struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	UserID      string     `json:"user_id"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Duration    int64      `json:"duration"` // ?
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateTaskRequest 
type CreateTaskRequest struct {
	Title          string       `json:"title" binding:"required"`
	Description    string       `json:"description"`
	Priority       TaskPriority `json:"priority"`
	Type           TaskType     `json:"type" binding:"required"`
	ProjectID      string       `json:"project_id" binding:"required"`
	TeamID         *string      `json:"team_id,omitempty"`
	AssigneeID     *string      `json:"assignee_id,omitempty"`
	EstimatedHours *float64     `json:"estimated_hours,omitempty"`
	StartDate      *time.Time   `json:"start_date,omitempty"`
	DueDate        *time.Time   `json:"due_date,omitempty"`
	Tags           []string     `json:"tags"`
	Labels         map[string]string `json:"labels"`
}

// UpdateTaskRequest 
type UpdateTaskRequest struct {
	Title          *string       `json:"title,omitempty"`
	Description    *string       `json:"description,omitempty"`
	Status         *TaskStatus   `json:"status,omitempty"`
	Priority       *TaskPriority `json:"priority,omitempty"`
	AssigneeID     *string       `json:"assignee_id,omitempty"`
	EstimatedHours *float64      `json:"estimated_hours,omitempty"`
	StartDate      *time.Time    `json:"start_date,omitempty"`
	DueDate        *time.Time    `json:"due_date,omitempty"`
	Progress       *float64      `json:"progress,omitempty"`
	Tags           []string      `json:"tags,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
}

// 
var tasks []Task
var projects []Project
var teams []Team
var comments []TaskComment
var timeLogs []TaskTimeLog

func initMockData() {
	// ?
	teamID1 := uuid.New().String()
	teamID2 := uuid.New().String()
	
	teams = []Team{
		{
			ID:          teamID1,
			Name:        "?,
			Description: "濪",
			LeaderID:    "user123",
			MemberIDs:   []string{"user123", "user124", "user125"},
			CreatedAt:   time.Now().Add(-time.Hour * 24 * 30),
			UpdatedAt:   time.Now().Add(-time.Hour * 24),
			IsActive:    true,
		},
		{
			ID:          teamID2,
			Name:        "?,
			Description: "",
			LeaderID:    "user126",
			MemberIDs:   []string{"user126", "user127", "user128"},
			CreatedAt:   time.Now().Add(-time.Hour * 24 * 25),
			UpdatedAt:   time.Now().Add(-time.Hour * 12),
			IsActive:    true,
		},
	}

	// ?
	projectID1 := uuid.New().String()
	projectID2 := uuid.New().String()
	
	projects = []Project{
		{
			ID:          projectID1,
			Name:        "",
			Description: "",
			Status:      "active",
			OwnerID:     "user123",
			TeamID:      teamID1,
			StartDate:   time.Now().Add(-time.Hour * 24 * 30),
			CreatedAt:   time.Now().Add(-time.Hour * 24 * 30),
			UpdatedAt:   time.Now().Add(-time.Hour * 2),
			Progress:    65.5,
			TaskCount:   12,
		},
		{
			ID:          projectID2,
			Name:        "?,
			Description: "?,
			Status:      "planning",
			OwnerID:     "user126",
			TeamID:      teamID2,
			StartDate:   time.Now().Add(time.Hour * 24 * 7),
			CreatedAt:   time.Now().Add(-time.Hour * 24 * 7),
			UpdatedAt:   time.Now().Add(-time.Hour * 1),
			Progress:    15.0,
			TaskCount:   8,
		},
	}

	// ?
	taskID1 := uuid.New().String()
	taskID2 := uuid.New().String()
	taskID3 := uuid.New().String()
	
	estimatedHours1 := 8.0
	estimatedHours2 := 16.0
	actualHours1 := 6.5
	assigneeID1 := "user124"
	assigneeID2 := "user127"
	
	dueDate1 := time.Now().Add(time.Hour * 24 * 3)
	dueDate2 := time.Now().Add(time.Hour * 24 * 7)
	startDate1 := time.Now().Add(-time.Hour * 24 * 2)
	
	tasks = []Task{
		{
			ID:             taskID1,
			Title:          "",
			Description:    "",
			Status:         TaskStatusInProgress,
			Priority:       TaskPriorityHigh,
			Type:           TaskTypeDevelopment,
			ProjectID:      projectID1,
			TeamID:         &teamID1,
			CreatorID:      "user123",
			AssigneeID:     &assigneeID1,
			EstimatedHours: &estimatedHours1,
			ActualHours:    &actualHours1,
			StartDate:      &startDate1,
			DueDate:        &dueDate1,
			CreatedAt:      time.Now().Add(-time.Hour * 24 * 3),
			UpdatedAt:      time.Now().Add(-time.Hour * 2),
			Progress:       75.0,
			Tags:           []string{"frontend", "ui", "authentication"},
			Labels:         map[string]string{"sprint": "sprint-1", "component": "auth"},
		},
		{
			ID:             taskID2,
			Title:          "API?,
			Description:    "RESTful API",
			Status:         TaskStatusAssigned,
			Priority:       TaskPriorityMedium,
			Type:           TaskTypeDevelopment,
			ProjectID:      projectID1,
			TeamID:         &teamID2,
			CreatorID:      "user126",
			AssigneeID:     &assigneeID2,
			EstimatedHours: &estimatedHours2,
			DueDate:        &dueDate2,
			CreatedAt:      time.Now().Add(-time.Hour * 24 * 2),
			UpdatedAt:      time.Now().Add(-time.Hour * 1),
			Progress:       25.0,
			Tags:           []string{"backend", "api", "user-management"},
			Labels:         map[string]string{"sprint": "sprint-1", "priority": "medium"},
		},
		{
			ID:          taskID3,
			Title:       "?,
			Description: "",
			Status:      TaskStatusPending,
			Priority:    TaskPriorityCritical,
			Type:        TaskTypeMaintenance,
			ProjectID:   projectID2,
			TeamID:      &teamID2,
			CreatorID:   "user126",
			CreatedAt:   time.Now().Add(-time.Hour * 24),
			UpdatedAt:   time.Now().Add(-time.Hour * 24),
			Progress:    0.0,
			Tags:        []string{"database", "performance", "optimization"},
			Labels:      map[string]string{"priority": "critical", "type": "maintenance"},
		},
	}

	// ?
	comments = []TaskComment{
		{
			ID:        uuid.New().String(),
			TaskID:    taskID1,
			AuthorID:  "user123",
			Content:   "",
			CreatedAt: time.Now().Add(-time.Hour * 4),
			UpdatedAt: time.Now().Add(-time.Hour * 4),
		},
		{
			ID:        uuid.New().String(),
			TaskID:    taskID1,
			AuthorID:  "user124",
			Content:   "?,
			CreatedAt: time.Now().Add(-time.Hour * 2),
			UpdatedAt: time.Now().Add(-time.Hour * 2),
		},
	}

	// ?
	endTime := time.Now().Add(-time.Hour * 2)
	timeLogs = []TaskTimeLog{
		{
			ID:          uuid.New().String(),
			TaskID:      taskID1,
			UserID:      "user124",
			StartTime:   time.Now().Add(-time.Hour * 4),
			EndTime:     &endTime,
			Duration:    7200, // 2
			Description: "?,
			CreatedAt:   time.Now().Add(-time.Hour * 2),
		},
	}
}

// 
func getTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	priority := c.Query("priority")
	projectID := c.Query("project_id")
	assigneeID := c.Query("assignee_id")
	search := c.Query("search")

	filteredTasks := make([]Task, 0)
	for _, task := range tasks {
		// ?
		if status != "" && string(task.Status) != status {
			continue
		}
		// ?
		if priority != "" && string(task.Priority) != priority {
			continue
		}
		// 
		if projectID != "" && task.ProjectID != projectID {
			continue
		}
		// ?
		if assigneeID != "" && (task.AssigneeID == nil || *task.AssigneeID != assigneeID) {
			continue
		}
		// 
		if search != "" && !strings.Contains(strings.ToLower(task.Title), strings.ToLower(search)) &&
			!strings.Contains(strings.ToLower(task.Description), strings.ToLower(search)) {
			continue
		}
		filteredTasks = append(filteredTasks, task)
	}

	total := len(filteredTasks)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data": gin.H{
			"tasks":       filteredTasks[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 
func getTask(c *gin.Context) {
	taskID := c.Param("id")
	
	for _, task := range tasks {
		if task.ID == taskID {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "",
				"data":    task,
			})
			return
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "?,
	})
}

// 
func createTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "",
			"error":   err.Error(),
		})
		return
	}

	task := Task{
		ID:             uuid.New().String(),
		Title:          req.Title,
		Description:    req.Description,
		Status:         TaskStatusPending,
		Priority:       req.Priority,
		Type:           req.Type,
		ProjectID:      req.ProjectID,
		TeamID:         req.TeamID,
		CreatorID:      "user123", // 
		AssigneeID:     req.AssigneeID,
		EstimatedHours: req.EstimatedHours,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Progress:       0.0,
		Tags:           req.Tags,
		Labels:         req.Labels,
	}

	if task.Tags == nil {
		task.Tags = make([]string, 0)
	}
	if task.Labels == nil {
		task.Labels = make(map[string]string)
	}

	tasks = append(tasks, task)

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "",
		"data":    task,
	})
}

// 
func updateTask(c *gin.Context) {
	taskID := c.Param("id")
	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "",
			"error":   err.Error(),
		})
		return
	}

	for i, task := range tasks {
		if task.ID == taskID {
			// 
			if req.Title != nil {
				tasks[i].Title = *req.Title
			}
			if req.Description != nil {
				tasks[i].Description = *req.Description
			}
			if req.Status != nil {
				tasks[i].Status = *req.Status
				if *req.Status == TaskStatusCompleted {
					now := time.Now()
					tasks[i].CompletedAt = &now
				}
			}
			if req.Priority != nil {
				tasks[i].Priority = *req.Priority
			}
			if req.AssigneeID != nil {
				tasks[i].AssigneeID = req.AssigneeID
			}
			if req.EstimatedHours != nil {
				tasks[i].EstimatedHours = req.EstimatedHours
			}
			if req.StartDate != nil {
				tasks[i].StartDate = req.StartDate
			}
			if req.DueDate != nil {
				tasks[i].DueDate = req.DueDate
			}
			if req.Progress != nil {
				tasks[i].Progress = *req.Progress
			}
			if req.Tags != nil {
				tasks[i].Tags = req.Tags
			}
			if req.Labels != nil {
				tasks[i].Labels = req.Labels
			}
			tasks[i].UpdatedAt = time.Now()

			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "",
				"data":    tasks[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "?,
	})
}

// 
func deleteTask(c *gin.Context) {
	taskID := c.Param("id")
	
	for i, task := range tasks {
		if task.ID == taskID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "",
			})
			return
		}
	}
	
	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "?,
	})
}

// 
func getProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	total := len(projects)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data": gin.H{
			"projects":    projects[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 
func getTeams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	total := len(teams)
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data": gin.H{
			"teams":       teams[start:end],
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// 
func getTaskComments(c *gin.Context) {
	taskID := c.Param("id")
	
	taskComments := make([]TaskComment, 0)
	for _, comment := range comments {
		if comment.TaskID == taskID {
			taskComments = append(taskComments, comment)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    taskComments,
	})
}

// 
func getTaskTimeLogs(c *gin.Context) {
	taskID := c.Param("id")
	
	taskTimeLogs := make([]TaskTimeLog, 0)
	for _, timeLog := range timeLogs {
		if timeLog.TaskID == taskID {
			taskTimeLogs = append(taskTimeLogs, timeLog)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    taskTimeLogs,
	})
}

// 
func getTaskStats(c *gin.Context) {
	projectID := c.Query("project_id")
	teamID := c.Query("team_id")

	stats := map[string]interface{}{
		"total_tasks": 0,
		"by_status": map[string]int{
			"pending":     0,
			"assigned":    0,
			"in_progress": 0,
			"completed":   0,
			"cancelled":   0,
			"on_hold":     0,
			"overdue":     0,
		},
		"by_priority": map[string]int{
			"low":      0,
			"medium":   0,
			"high":     0,
			"critical": 0,
		},
		"by_type": map[string]int{
			"development":   0,
			"bug":          0,
			"feature":      0,
			"research":     0,
			"maintenance":  0,
			"review":       0,
			"testing":      0,
			"documentation": 0,
		},
	}

	for _, task := range tasks {
		// 
		if projectID != "" && task.ProjectID != projectID {
			continue
		}
		// 
		if teamID != "" && (task.TeamID == nil || *task.TeamID != teamID) {
			continue
		}

		stats["total_tasks"] = stats["total_tasks"].(int) + 1
		stats["by_status"].(map[string]int)[string(task.Status)]++
		stats["by_priority"].(map[string]int)[string(task.Priority)]++
		stats["by_type"].(map[string]int)[string(task.Type)]++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    stats,
	})
}

// ?
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "task-management",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
		"stats": gin.H{
			"total_tasks":    len(tasks),
			"total_projects": len(projects),
			"total_teams":    len(teams),
		},
	})
}

func main() {
	// ?
	initMockData()

	// Gin?
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API?
	api := r.Group("/api/v1")
	{
		// 
		tasks := api.Group("/tasks")
		{
			tasks.GET("", getTasks)
			tasks.POST("", createTask)
			tasks.GET("/:id", getTask)
			tasks.PUT("/:id", updateTask)
			tasks.DELETE("/:id", deleteTask)
			tasks.GET("/:id/comments", getTaskComments)
			tasks.GET("/:id/time-logs", getTaskTimeLogs)
		}

		// 
		projects := api.Group("/projects")
		{
			projects.GET("", getProjects)
		}

		// 
		teams := api.Group("/teams")
		{
			teams.GET("", getTeams)
		}

		// 
		api.GET("/stats/tasks", getTaskStats)
	}

	// ?
	r.GET("/health", healthCheck)

	log.Println(" (Mock汾) ?8084")
	log.Fatal(r.Run(":8084"))
}

