package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"task-management/internal/application"
	"task-management/internal/interfaces/http/handlers"
	"task-management/internal/interfaces/http/middleware"
)

// Router HTTP路由?
type Router struct {
	taskHandler    *handlers.TaskHandler
	projectHandler *handlers.ProjectHandler
	teamHandler    *handlers.TeamHandler
}

// NewRouter 创建新的路由?
func NewRouter(
	taskService *application.TaskService,
	projectService *application.ProjectService,
	teamService *application.TeamService,
) *Router {
	return &Router{
		taskHandler:    handlers.NewTaskHandler(taskService),
		projectHandler: handlers.NewProjectHandler(projectService),
		teamHandler:    handlers.NewTeamHandler(teamService),
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes() http.Handler {
	router := mux.NewRouter()

	// 添加中间?
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)

	// API版本前缀
	api := router.PathPrefix("/api/v1").Subrouter()

	// 任务路由
	r.setupTaskRoutes(api)

	// 项目路由
	r.setupProjectRoutes(api)

	// 团队路由
	r.setupTeamRoutes(api)

	// 健康检?
	router.HandleFunc("/health", r.healthCheck).Methods("GET")

	// 设置CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	return c.Handler(router)
}

// setupTaskRoutes 设置任务路由
func (r *Router) setupTaskRoutes(api *mux.Router) {
	tasks := api.PathPrefix("/tasks").Subrouter()

	// 基本CRUD操作
	tasks.HandleFunc("", r.taskHandler.CreateTask).Methods("POST")
	tasks.HandleFunc("", r.taskHandler.ListTasks).Methods("GET")
	tasks.HandleFunc("/search", r.taskHandler.SearchTasks).Methods("GET")
	tasks.HandleFunc("/{id}", r.taskHandler.GetTask).Methods("GET")
	tasks.HandleFunc("/{id}", r.taskHandler.UpdateTask).Methods("PUT")
	tasks.HandleFunc("/{id}", r.taskHandler.DeleteTask).Methods("DELETE")

	// 任务分配
	tasks.HandleFunc("/{id}/assign", r.taskHandler.AssignTask).Methods("POST")
	tasks.HandleFunc("/{id}/unassign", r.taskHandler.UnassignTask).Methods("POST")
	tasks.HandleFunc("/auto-assign", r.taskHandler.AutoAssignTasks).Methods("POST")

	// 任务统计
	tasks.HandleFunc("/statistics", r.taskHandler.GetTaskStatistics).Methods("GET")

	// 任务评论和时间记?
	tasks.HandleFunc("/{id}/comments", r.taskHandler.AddTaskComment).Methods("POST")
	tasks.HandleFunc("/{id}/time-logs", r.taskHandler.AddTimeLog).Methods("POST")
}

// setupProjectRoutes 设置项目路由
func (r *Router) setupProjectRoutes(api *mux.Router) {
	projects := api.PathPrefix("/projects").Subrouter()

	// 基本CRUD操作
	projects.HandleFunc("", r.projectHandler.CreateProject).Methods("POST")
	projects.HandleFunc("", r.projectHandler.ListProjects).Methods("GET")
	projects.HandleFunc("/search", r.projectHandler.SearchProjects).Methods("GET")
	projects.HandleFunc("/{id}", r.projectHandler.GetProject).Methods("GET")
	projects.HandleFunc("/{id}", r.projectHandler.UpdateProject).Methods("PUT")
	projects.HandleFunc("/{id}", r.projectHandler.DeleteProject).Methods("DELETE")

	// 项目成员管理
	projects.HandleFunc("/{id}/members", r.projectHandler.AddProjectMember).Methods("POST")
	projects.HandleFunc("/{id}/members/{user_id}", r.projectHandler.RemoveProjectMember).Methods("DELETE")
	projects.HandleFunc("/{id}/members/{user_id}/role", r.projectHandler.UpdateProjectMemberRole).Methods("PUT")

	// 项目统计和分?
	projects.HandleFunc("/{id}/statistics", r.projectHandler.GetProjectStatistics).Methods("GET")
	projects.HandleFunc("/{id}/schedule", r.projectHandler.GenerateProjectSchedule).Methods("GET")
	projects.HandleFunc("/{id}/performance", r.projectHandler.GetProjectPerformance).Methods("GET")
}

// setupTeamRoutes 设置团队路由
func (r *Router) setupTeamRoutes(api *mux.Router) {
	teams := api.PathPrefix("/teams").Subrouter()

	// 基本CRUD操作
	teams.HandleFunc("", r.teamHandler.CreateTeam).Methods("POST")
	teams.HandleFunc("", r.teamHandler.ListTeams).Methods("GET")
	teams.HandleFunc("/search", r.teamHandler.SearchTeams).Methods("GET")
	teams.HandleFunc("/{id}", r.teamHandler.GetTeam).Methods("GET")
	teams.HandleFunc("/{id}", r.teamHandler.UpdateTeam).Methods("PUT")
	teams.HandleFunc("/{id}", r.teamHandler.DeleteTeam).Methods("DELETE")

	// 团队成员管理
	teams.HandleFunc("/{id}/members", r.teamHandler.AddTeamMember).Methods("POST")
	teams.HandleFunc("/{id}/members/{user_id}", r.teamHandler.RemoveTeamMember).Methods("DELETE")
	teams.HandleFunc("/{id}/members/{user_id}", r.teamHandler.UpdateTeamMember).Methods("PUT")

	// 团队统计和分?
	teams.HandleFunc("/{id}/statistics", r.teamHandler.GetTeamStatistics).Methods("GET")
	teams.HandleFunc("/{id}/performance", r.teamHandler.GetTeamPerformance).Methods("GET")
	teams.HandleFunc("/{id}/workload", r.teamHandler.GetTeamWorkload).Methods("GET")
	teams.HandleFunc("/{id}/workload/optimize", r.teamHandler.OptimizeTeamWorkload).Methods("POST")
}

// healthCheck 健康检?
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "task-management"}`))
}

