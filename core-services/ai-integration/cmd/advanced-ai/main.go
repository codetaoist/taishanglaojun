package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/advanced"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/agi"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/config"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/evolution"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/meta-learning"
)

const (
	// 服务信息
	ServiceName    = "taishang-laojun-advanced-ai"
	ServiceVersion = "1.0.0"
	DefaultPort    = "8080"
)

func main() {
	// 初始化日�?	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting %s v%s", ServiceName, ServiceVersion)

	// 加载配置
	cfg := config.DefaultAdvancedAIConfig()
	cfg.LoadFromEnv()
	
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	
	log.Printf("Configuration loaded successfully")

	// 创建上下�?	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化服务组�?	services, err := initializeServices(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}
	defer services.Cleanup()

	// 创建HTTP服务�?	server := createHTTPServer(cfg, services)

	// 启动服务�?	go func() {
		port := getPort()
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	waitForShutdown(ctx, server, services)
}

// ServiceContainer 服务容器
type ServiceContainer struct {
	Config           *config.AdvancedAIConfig
	AGIService       *agi.AGIService
	MetaLearning     *meta_learning.MetaLearningEngine
	SelfEvolution    *evolution.SelfEvolutionSystem
	AdvancedAI       *advanced.AdvancedAIService
	Handler          *handlers.AdvancedAIHandler
}

// Cleanup 清理资源
func (sc *ServiceContainer) Cleanup() {
	log.Println("Cleaning up services...")
	
	if sc.AdvancedAI != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := sc.AdvancedAI.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down advanced AI service: %v", err)
		}
	}
	
	if sc.SelfEvolution != nil {
		if err := sc.SelfEvolution.StopEvolution(); err != nil {
			log.Printf("Error stopping evolution system: %v", err)
		}
	}
	
	log.Println("Services cleanup completed")
}

// initializeServices 初始化所有服�?func initializeServices(ctx context.Context, cfg *config.AdvancedAIConfig) (*ServiceContainer, error) {
	log.Println("Initializing services...")

	// 初始化AGI服务
	agiService := agi.NewAGIService()
	if err := agiService.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize AGI service: %w", err)
	}
	log.Println("AGI service initialized")

	// 初始化元学习引擎
	metaLearning := meta_learning.NewMetaLearningEngine()
	if err := metaLearning.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize meta-learning engine: %w", err)
	}
	log.Println("Meta-learning engine initialized")

	// 初始化自我进化系�?	selfEvolution := evolution.NewSelfEvolutionSystem()
	if err := selfEvolution.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize self-evolution system: %w", err)
	}
	log.Println("Self-evolution system initialized")

	// 初始化高级AI服务
	advancedAI := advanced.NewAdvancedAIService(agiService, metaLearning, selfEvolution)
	if err := advancedAI.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize advanced AI service: %w", err)
	}
	log.Println("Advanced AI service initialized")

	// 创建处理�?	handler := handlers.NewAdvancedAIHandler(advancedAI)
	log.Println("HTTP handler created")

	// 启动后台任务
	go startBackgroundTasks(ctx, cfg, advancedAI)

	return &ServiceContainer{
		Config:        cfg,
		AGIService:    agiService,
		MetaLearning:  metaLearning,
		SelfEvolution: selfEvolution,
		AdvancedAI:    advancedAI,
		Handler:       handler,
	}, nil
}

// createHTTPServer 创建HTTP服务�?func createHTTPServer(cfg *config.AdvancedAIConfig, services *ServiceContainer) *http.Server {
	// 设置Gin模式
	if cfg.Logging.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由�?	router := gin.New()

	// 添加中间�?	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(securityMiddleware(cfg))
	
	if cfg.Security.EnableRateLimiting {
		router.Use(rateLimitMiddleware(cfg))
	}

	// 注册路由
	registerRoutes(router, services)

	// 创建服务�?	server := &http.Server{
		Addr:         ":" + getPort(),
		Handler:      router,
		ReadTimeout:  cfg.DefaultTimeout,
		WriteTimeout: cfg.DefaultTimeout,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// registerRoutes 注册路由
func registerRoutes(router *gin.Engine, services *ServiceContainer) {
	// 健康检�?	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   ServiceName,
			"version":   ServiceVersion,
			"timestamp": time.Now(),
		})
	})

	// 服务信息
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     ServiceName,
			"version":     ServiceVersion,
			"description": "Taishang Laojun Advanced AI Platform",
			"capabilities": []string{
				"AGI Integration",
				"Meta-Learning Engine",
				"Self-Evolution System",
				"Hybrid AI Processing",
			},
			"endpoints": gin.H{
				"health":      "/health",
				"info":        "/info",
				"api":         "/api/v1/advanced-ai",
				"docs":        "/docs",
				"metrics":     "/metrics",
			},
		})
	})

	// API文档
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"title":       "Taishang Laojun Advanced AI API",
			"version":     ServiceVersion,
			"description": "Advanced AI capabilities including AGI, Meta-Learning, and Self-Evolution",
			"endpoints": map[string]interface{}{
				"POST /api/v1/advanced-ai/process": gin.H{
					"description": "Process general AI request",
					"parameters": gin.H{
						"type":         "string (required)",
						"capability":   "string (optional)",
						"input":        "object (required)",
						"context":      "object (optional)",
						"requirements": "object (optional)",
					},
				},
				"POST /api/v1/advanced-ai/agi/task": gin.H{
					"description": "Process AGI-specific task",
					"parameters": gin.H{
						"type":         "string (required)",
						"input":        "object (required)",
						"context":      "object (optional)",
						"requirements": "object (optional)",
					},
				},
				"POST /api/v1/advanced-ai/meta-learning/learn": gin.H{
					"description": "Trigger meta-learning process",
					"parameters": gin.H{
						"task_type": "string (required)",
						"domain":    "string (required)",
						"data":      "array (required)",
						"strategy":  "string (optional)",
					},
				},
				"POST /api/v1/advanced-ai/evolution/optimize": gin.H{
					"description": "Trigger evolution optimization",
					"parameters": gin.H{
						"optimization_targets": "array (required)",
						"strategy":            "string (optional)",
						"parameters":          "object (optional)",
					},
				},
				"GET /api/v1/advanced-ai/status": gin.H{
					"description": "Get system status",
				},
				"GET /api/v1/advanced-ai/metrics": gin.H{
					"description": "Get performance metrics",
				},
			},
		})
	})

	// 指标端点
	router.GET("/metrics", func(c *gin.Context) {
		status := services.AdvancedAI.GetSystemStatus()
		c.JSON(http.StatusOK, gin.H{
			"service_metrics": gin.H{
				"total_requests":    status.TotalRequests,
				"success_rate":      status.SuccessRate,
				"avg_response_time": status.AvgResponseTime,
				"active_requests":   status.ActiveRequests,
				"overall_health":    status.OverallHealth,
			},
			"system_metrics": gin.H{
				"uptime":     time.Since(time.Now().Add(-time.Hour)), // 模拟运行时间
				"goroutines": "N/A", // 可以添加实际的goroutine数量
				"memory":     "N/A", // 可以添加实际的内存使用情�?			},
			"timestamp": time.Now(),
		})
	})

	// 注册高级AI API路由
	services.Handler.RegisterRoutes(router)
}

// startBackgroundTasks 启动后台任务
func startBackgroundTasks(ctx context.Context, cfg *config.AdvancedAIConfig, service *advanced.AdvancedAIService) {
	log.Println("Starting background tasks...")

	// 性能监控任务
	if cfg.Monitoring.EnablePerformanceMonitoring {
		go performanceMonitoringTask(ctx, cfg, service)
	}

	// 健康检查任�?	if cfg.Monitoring.EnableHealthChecks {
		go healthCheckTask(ctx, cfg, service)
	}

	// 自动优化任务
	go autoOptimizationTask(ctx, cfg, service)

	log.Println("Background tasks started")
}

// performanceMonitoringTask 性能监控任务
func performanceMonitoringTask(ctx context.Context, cfg *config.AdvancedAIConfig, service *advanced.AdvancedAIService) {
	ticker := time.NewTicker(cfg.Monitoring.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 收集性能指标
			metrics := service.GetPerformanceMetrics(100)
			log.Printf("Collected %d performance metrics", len(metrics))
			
			// 这里可以将指标发送到监控系统
		}
	}
}

// healthCheckTask 健康检查任�?func healthCheckTask(ctx context.Context, cfg *config.AdvancedAIConfig, service *advanced.AdvancedAIService) {
	ticker := time.NewTicker(cfg.Monitoring.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 执行健康检�?			status := service.GetSystemStatus()
			if status.OverallHealth < cfg.Monitoring.HealthThreshold {
				log.Printf("Health check warning: overall health is %.2f (threshold: %.2f)", 
					status.OverallHealth, cfg.Monitoring.HealthThreshold)
			}
		}
	}
}

// autoOptimizationTask 自动优化任务
func autoOptimizationTask(ctx context.Context, cfg *config.AdvancedAIConfig, service *advanced.AdvancedAIService) {
	ticker := time.NewTicker(1 * time.Hour) // 每小时执行一次优�?	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 执行自动优化
			log.Println("Starting auto-optimization...")
			if err := service.AutoOptimize(ctx); err != nil {
				log.Printf("Auto-optimization failed: %v", err)
			} else {
				log.Println("Auto-optimization completed successfully")
			}
		}
	}
}

// 中间�?
// corsMiddleware CORS中间�?func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// securityMiddleware 安全中间�?func securityMiddleware(cfg *config.AdvancedAIConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 添加安全�?		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// 检查请求大�?		if c.Request.ContentLength > cfg.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request entity too large",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// rateLimitMiddleware 限流中间�?func rateLimitMiddleware(cfg *config.AdvancedAIConfig) gin.HandlerFunc {
	// 简化的限流实现，实际应该使用专业的限流�?	return func(c *gin.Context) {
		// 这里应该实现实际的限流逻辑
		c.Next()
	}
}

// 工具函数

// getPort 获取端口�?func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}
	return port
}

// waitForShutdown 等待关闭信号
func waitForShutdown(ctx context.Context, server *http.Server, services *ServiceContainer) {
	// 创建信号通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit
	log.Println("Shutting down server...")

	// 创建关闭上下�?	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务�?	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 清理服务
	services.Cleanup()

	log.Println("Server exited")
}
