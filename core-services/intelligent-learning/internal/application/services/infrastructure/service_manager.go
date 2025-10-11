package infrastructure

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ServiceManager 服务管理�?
type ServiceManager struct {
	initializer  *ServiceInitializer
	errorHandler *ErrorHandler
	logger       *Logger
	config       *ServiceManagerConfig
	running      bool
	mu           sync.RWMutex
}

// ServiceManagerConfig 服务管理器配�?
type ServiceManagerConfig struct {
	ConfigPath          string        `json:"config_path"`
	LogLevel            LogLevel      `json:"log_level"`
	ShutdownTimeout     time.Duration `json:"shutdown_timeout"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	EnableHealthCheck   bool          `json:"enable_health_check"`
	EnableMetrics       bool          `json:"enable_metrics"`
	MetricsPort         int           `json:"metrics_port"`
	EnableProfiling     bool          `json:"enable_profiling"`
	ProfilingPort       int           `json:"profiling_port"`
}

// NewServiceManager 创建服务管理�?
func NewServiceManager(config *ServiceManagerConfig) *ServiceManager {
	if config == nil {
		config = &ServiceManagerConfig{
			ConfigPath:          "./config/services.json",
			LogLevel:            LogLevelInfo,
			ShutdownTimeout:     30 * time.Second,
			HealthCheckInterval: 30 * time.Second,
			EnableHealthCheck:   true,
			EnableMetrics:       true,
			MetricsPort:         9090,
			EnableProfiling:     false,
			ProfilingPort:       6060,
		}
	}
	
	logger := NewLogger(config.LogLevel)
	errorHandler := NewErrorHandler(nil, logger)
	initializer := NewServiceInitializer(config.ConfigPath)
	
	return &ServiceManager{
		initializer:  initializer,
		errorHandler: errorHandler,
		logger:       logger,
		config:       config,
		running:      false,
	}
}

// Start 启动服务管理�?
func (sm *ServiceManager) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.running {
		return fmt.Errorf("service manager is already running")
	}
	
	sm.logger.Info("Starting Intelligent Learning Service Manager...")
	
	// 初始化所有服�?
	results, err := sm.initializer.Initialize(ctx)
	if err != nil {
		serviceErr := sm.errorHandler.HandleError(ctx, err, "service_manager", "initialize", ErrorTypeService, SeverityCritical)
		return serviceErr
	}
	
	// 检查初始化结果
	failedServices := make([]string, 0)
	for _, result := range results {
		if !result.Success {
			failedServices = append(failedServices, result.ServiceName)
			sm.logger.Error(fmt.Sprintf("Service %s failed to initialize: %v", result.ServiceName, result.Error))
		} else {
			sm.logger.Info(fmt.Sprintf("Service %s initialized successfully in %v", result.ServiceName, result.InitTime))
		}
	}
	
	if len(failedServices) > 0 {
		return fmt.Errorf("failed to initialize services: %v", failedServices)
	}
	
	sm.running = true
	
	// 启动健康检�?
	if sm.config.EnableHealthCheck {
		go sm.startHealthCheck(ctx)
	}
	
	// 启动指标收集
	if sm.config.EnableMetrics {
		go sm.startMetricsServer(ctx)
	}
	
	// 启动性能分析
	if sm.config.EnableProfiling {
		go sm.startProfilingServer(ctx)
	}
	
	sm.logger.Info("Intelligent Learning Service Manager started successfully")
	
	return nil
}

// Stop 停止服务管理�?
func (sm *ServiceManager) Stop(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.running {
		return fmt.Errorf("service manager is not running")
	}
	
	sm.logger.Info("Stopping Intelligent Learning Service Manager...")
	
	// 创建带超时的上下�?
	shutdownCtx, cancel := context.WithTimeout(ctx, sm.config.ShutdownTimeout)
	defer cancel()
	
	// 关闭所有服�?
	if err := sm.initializer.Shutdown(shutdownCtx); err != nil {
		serviceErr := sm.errorHandler.HandleError(shutdownCtx, err, "service_manager", "shutdown", ErrorTypeService, SeverityHigh)
		sm.logger.Error(fmt.Sprintf("Error during shutdown: %v", serviceErr))
		return serviceErr
	}
	
	sm.running = false
	sm.logger.Info("Intelligent Learning Service Manager stopped successfully")
	
	return nil
}

// IsRunning 检查是否正在运�?
func (sm *ServiceManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// GetStatus 获取服务状�?
func (sm *ServiceManager) GetStatus() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":     sm.running,
		"initialized": sm.initializer.IsInitialized(),
		"services":    sm.initializer.GetInitializationStatus(),
		"timestamp":   time.Now(),
	}
	
	// 添加错误统计
	errorStats := sm.errorHandler.GetErrorStatistics(time.Hour)
	status["error_statistics"] = errorStats
	
	return status
}

// GetHealthStatus 获取健康状�?
func (sm *ServiceManager) GetHealthStatus() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services":  make(map[string]interface{}),
	}
	
	if !sm.running {
		health["status"] = "unhealthy"
		health["reason"] = "service manager not running"
		return health
	}
	
	// 检查各个服务的健康状�?
	integration, err := sm.initializer.GetIntegration()
	if err != nil {
		health["status"] = "unhealthy"
		health["reason"] = "service integration not available"
		return health
	}
	
	// 获取服务健康状�?
	systemHealth, err := integration.GetSystemHealth(context.Background())
	if err != nil {
		health["status"] = "unhealthy"
		health["reason"] = "failed to get system health"
		return health
	}
	
	health["services"] = systemHealth.Services
	
	// 检查是否有不健康的服务
	for serviceName, serviceStatus := range systemHealth.Services {
		if serviceStatus != "healthy" {
			health["status"] = "degraded"
			if health["unhealthy_services"] == nil {
				health["unhealthy_services"] = []string{}
			}
			health["unhealthy_services"] = append(
				health["unhealthy_services"].([]string),
				serviceName,
			)
		}
	}
	
	return health
}

// GetMetrics 获取指标
func (sm *ServiceManager) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now()), // 这里应该记录实际的启动时�?
	}
	
	if sm.running {
		integration, err := sm.initializer.GetIntegration()
		if err == nil {
			serviceMetrics, err := integration.GetMetrics(context.Background())
			if err == nil {
				metrics["services"] = serviceMetrics
			}
		}
	}
	
	// 添加错误指标
	errorStats := sm.errorHandler.GetErrorStatistics(time.Hour)
	metrics["errors"] = errorStats
	
	return metrics
}

// ProcessRequest 处理学习请求
func (sm *ServiceManager) ProcessRequest(ctx context.Context, request interface{}) (interface{}, error) {
	if !sm.running {
		return nil, sm.errorHandler.CreateError("service_manager", "process_request", "service manager not running", ErrorTypeService, SeverityHigh)
	}
	
	integration, err := sm.initializer.GetIntegration()
	if err != nil {
		return nil, sm.errorHandler.WrapError(err, "service_manager", "process_request", ErrorTypeIntegration, SeverityHigh, nil)
	}
	
	// 类型断言�?LearningRequest
	learningRequest, ok := request.(*LearningRequest)
	if !ok {
		return nil, sm.errorHandler.CreateError("service_manager", "process_request", "invalid request type", ErrorTypeValidation, SeverityHigh)
	}
	
	// 使用错误处理器进行重�?
	var result interface{}
	retryErr := sm.errorHandler.RetryOperation(ctx, func() error {
		var processErr error
		result, processErr = integration.ProcessLearningRequest(ctx, learningRequest)
		return processErr
	}, "service_integration", "process_learning_request", 3)
	
	if retryErr != nil {
		return nil, retryErr
	}
	
	return result, nil
}

// startHealthCheck 启动健康检�?
func (sm *ServiceManager) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(sm.config.HealthCheckInterval)
	defer ticker.Stop()
	
	sm.logger.Info("Health check started")
	
	for {
		select {
		case <-ctx.Done():
			sm.logger.Info("Health check stopped")
			return
		case <-ticker.C:
			sm.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck 执行健康检�?
func (sm *ServiceManager) performHealthCheck(ctx context.Context) {
	health := sm.GetHealthStatus()
	
	if status, ok := health["status"].(string); ok && status != "healthy" {
		sm.logger.Warn("Health check failed", map[string]interface{}{
			"status": status,
			"health": health,
		})
		
		// 这里可以添加自动恢复逻辑
		sm.attemptAutoRecovery(ctx, health)
	} else {
		sm.logger.Debug("Health check passed")
	}
}

// attemptAutoRecovery 尝试自动恢复
func (sm *ServiceManager) attemptAutoRecovery(ctx context.Context, health map[string]interface{}) {
	sm.logger.Info("Attempting auto recovery...")
	
	// 这里可以实现具体的自动恢复逻辑
	// 例如重启失败的服务、清理缓存等
	
	if unhealthyServices, ok := health["unhealthy_services"].([]string); ok {
		for _, serviceName := range unhealthyServices {
			sm.logger.Info(fmt.Sprintf("Attempting to recover service: %s", serviceName))
			// 实现服务恢复逻辑
		}
	}
}

// startMetricsServer 启动指标服务�?
func (sm *ServiceManager) startMetricsServer(ctx context.Context) {
	sm.logger.Info(fmt.Sprintf("Metrics server starting on port %d", sm.config.MetricsPort))
	
	// 这里应该实现实际的HTTP指标服务�?
	// 例如使用Prometheus metrics
	
	select {
	case <-ctx.Done():
		sm.logger.Info("Metrics server stopped")
	}
}

// startProfilingServer 启动性能分析服务�?
func (sm *ServiceManager) startProfilingServer(ctx context.Context) {
	sm.logger.Info(fmt.Sprintf("Profiling server starting on port %d", sm.config.ProfilingPort))
	
	// 这里应该实现实际的pprof服务�?
	
	select {
	case <-ctx.Done():
		sm.logger.Info("Profiling server stopped")
	}
}

// RunWithGracefulShutdown 运行服务管理器并支持优雅关闭
func (sm *ServiceManager) RunWithGracefulShutdown() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 启动服务
	if err := sm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start service manager: %w", err)
	}
	
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sm.logger.Info("Service manager is running. Press Ctrl+C to stop.")
	
	// 等待信号
	<-sigChan
	sm.logger.Info("Received shutdown signal, starting graceful shutdown...")
	
	// 取消上下文以停止所有goroutine
	cancel()
	
	// 停止服务
	if err := sm.Stop(context.Background()); err != nil {
		return fmt.Errorf("failed to stop service manager: %w", err)
	}
	
	sm.logger.Info("Service manager shutdown completed")
	return nil
}

// GetLogger 获取日志记录�?
func (sm *ServiceManager) GetLogger() *Logger {
	return sm.logger
}

// GetErrorHandler 获取错误处理�?
func (sm *ServiceManager) GetErrorHandler() *ErrorHandler {
	return sm.errorHandler
}

// GetInitializer 获取服务初始化器
func (sm *ServiceManager) GetInitializer() *ServiceInitializer {
	return sm.initializer
}
