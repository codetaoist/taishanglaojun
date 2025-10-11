package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/task-management/internal/application"
	"github.com/codetaoist/taishanglaojun/core-services/task-management/internal/infrastructure/persistence"
	"github.com/codetaoist/taishanglaojun/core-services/task-management/internal/infrastructure/services"
	"github.com/codetaoist/taishanglaojun/core-services/task-management/internal/interfaces/http/middleware"
)

// Server HTTP服务�?
type Server struct {
	httpServer *http.Server
	router     *Router
}

// Config 服务器配�?
type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// NewServer 创建新的HTTP服务�?
func NewServer(config *Config) *Server {
	// 初始化仓储层
	taskRepo := persistence.NewInMemoryTaskRepository()
	projectRepo := persistence.NewInMemoryProjectRepository()
	teamRepo := persistence.NewInMemoryTeamRepository()

	// 初始化领域服务工�?
	domainServiceFactory := services.NewDomainServiceFactory(taskRepo, projectRepo, teamRepo)

	// 初始化应用服�?
	taskService := application.NewTaskService(taskRepo, projectRepo, teamRepo, domainServiceFactory)
	projectService := application.NewProjectService(projectRepo, taskRepo, teamRepo, domainServiceFactory)
	teamService := application.NewTeamService(teamRepo, taskRepo, projectRepo, domainServiceFactory)

	// 初始化路�?
	router := NewRouter(taskService, projectService, teamService)

	// 创建HTTP服务�?
	httpServer := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router.SetupRoutes(),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		router:     router,
	}
}

// Start 启动服务�?
func (s *Server) Start() error {
	log.Printf("Starting HTTP server on port %s", s.httpServer.Addr)

	// 启动服务�?
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop 停止服务�?
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	log.Println("HTTP server stopped")
	return nil
}

// StartWithGracefulShutdown 启动服务器并支持优雅关闭
func (s *Server) StartWithGracefulShutdown() error {
	// 创建错误通道
	errChan := make(chan error, 1)

	// 在goroutine中启动服务器
	go func() {
		if err := s.Start(); err != nil {
			errChan <- err
		}
	}()

	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待错误或信�?
	select {
	case err := <-errChan:
		return err
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)

		// 创建关闭上下�?
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 优雅关闭
		return s.Stop(ctx)
	}
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// ConfigFromEnv 从环境变量创建配�?
func ConfigFromEnv() *Config {
	config := DefaultConfig()

	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}

	if readTimeout := os.Getenv("READ_TIMEOUT"); readTimeout != "" {
		if duration, err := time.ParseDuration(readTimeout); err == nil {
			config.ReadTimeout = duration
		}
	}

	if writeTimeout := os.Getenv("WRITE_TIMEOUT"); writeTimeout != "" {
		if duration, err := time.ParseDuration(writeTimeout); err == nil {
			config.WriteTimeout = duration
		}
	}

	if idleTimeout := os.Getenv("IDLE_TIMEOUT"); idleTimeout != "" {
		if duration, err := time.ParseDuration(idleTimeout); err == nil {
			config.IdleTimeout = duration
		}
	}

	return config
}

// HealthCheck 健康检查处理器
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
}

// ApplyMiddleware 应用中间�?
func ApplyMiddleware(handler http.Handler) http.Handler {
	// 按顺序应用中间件
	handler = middleware.CORSMiddleware(handler)
	handler = middleware.ValidationMiddleware(handler)
	handler = middleware.RateLimitMiddleware(100)(handler) // 每分�?00个请�?
	handler = middleware.RequestIDMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.RecoveryMiddleware(handler)

	return handler
}
