package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/project-scaffold/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/project-scaffold/internal/logger"
	"github.com/codetaoist/taishanglaojun/infrastructure/project-scaffold/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Server HTTP服务器配置
type Server struct {
	config *config.Config
	logger logger.Logger
	router *gin.Engine
	server *http.Server
}

// New 创建新的服务器配置实例
func New(cfg *config.Config, log logger.Logger) *Server {
	// 设置Gin模式
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由实例
	router := gin.New()

	// 添加中间件
	router.Use(middleware.Logger(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS())

	return &Server{
		config: cfg,
		logger: log,
		router: router,
	}
}

// Start 启动服务器器配置实例
func (s *Server) Start() error {
	// 设置路由
	s.setupRoutes()

	// 创建HTTP服务实例
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:      s.router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器器配置实例
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// 等待中断信号
	s.waitForShutdown()

	return nil
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查路由
	s.router.GET("/health", s.healthCheck)

	// API版本路由
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/ping", s.ping)
		v1.GET("/version", s.version)
	}
}

// healthCheck 健康检查处理器
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   s.config.App.Name,
		"version":   s.config.App.Version,
	})
}

// ping Ping处理
func (s *Server) ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// version 版本处理
func (s *Server) version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        s.config.App.Name,
		"version":     s.config.App.Version,
		"environment": s.config.App.Environment,
	})
}

// waitForShutdown 等待关闭信号
func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器配置实例
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
	}

	s.logger.Info("Server exited")
}

