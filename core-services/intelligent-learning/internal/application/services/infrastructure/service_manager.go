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

// ServiceManager ?
type ServiceManager struct {
	initializer  *ServiceInitializer
	errorHandler *ErrorHandler
	logger       *Logger
	config       *ServiceManagerConfig
	running      bool
	mu           sync.RWMutex
}

// ServiceManagerConfig ?
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

// NewServiceManager ?
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

// Start ?
func (sm *ServiceManager) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if sm.running {
		return fmt.Errorf("service manager is already running")
	}
	
	sm.logger.Info("Starting Intelligent Learning Service Manager...")
	
	// ?
	results, err := sm.initializer.Initialize(ctx)
	if err != nil {
		serviceErr := sm.errorHandler.HandleError(ctx, err, "service_manager", "initialize", ErrorTypeService, SeverityCritical)
		return serviceErr
	}
	
	// 
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
	
	// ?
	if sm.config.EnableHealthCheck {
		go sm.startHealthCheck(ctx)
	}
	
	// 
	if sm.config.EnableMetrics {
		go sm.startMetricsServer(ctx)
	}
	
	// 
	if sm.config.EnableProfiling {
		go sm.startProfilingServer(ctx)
	}
	
	sm.logger.Info("Intelligent Learning Service Manager started successfully")
	
	return nil
}

// Stop ?
func (sm *ServiceManager) Stop(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.running {
		return fmt.Errorf("service manager is not running")
	}
	
	sm.logger.Info("Stopping Intelligent Learning Service Manager...")
	
	// ?
	shutdownCtx, cancel := context.WithTimeout(ctx, sm.config.ShutdownTimeout)
	defer cancel()
	
	// ?
	if err := sm.initializer.Shutdown(shutdownCtx); err != nil {
		serviceErr := sm.errorHandler.HandleError(shutdownCtx, err, "service_manager", "shutdown", ErrorTypeService, SeverityHigh)
		sm.logger.Error(fmt.Sprintf("Error during shutdown: %v", serviceErr))
		return serviceErr
	}
	
	sm.running = false
	sm.logger.Info("Intelligent Learning Service Manager stopped successfully")
	
	return nil
}

// IsRunning ?
func (sm *ServiceManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.running
}

// GetStatus ?
func (sm *ServiceManager) GetStatus() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	status := map[string]interface{}{
		"running":     sm.running,
		"initialized": sm.initializer.IsInitialized(),
		"services":    sm.initializer.GetInitializationStatus(),
		"timestamp":   time.Now(),
	}
	
	// 
	errorStats := sm.errorHandler.GetErrorStatistics(time.Hour)
	status["error_statistics"] = errorStats
	
	return status
}

// GetHealthStatus ?
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
	
	// ?
	integration, err := sm.initializer.GetIntegration()
	if err != nil {
		health["status"] = "unhealthy"
		health["reason"] = "service integration not available"
		return health
	}
	
	// ?
	systemHealth, err := integration.GetSystemHealth(context.Background())
	if err != nil {
		health["status"] = "unhealthy"
		health["reason"] = "failed to get system health"
		return health
	}
	
	health["services"] = systemHealth.Services
	
	// 
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

// GetMetrics 
func (sm *ServiceManager) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"uptime":    time.Since(time.Now()), // ?
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
	
	// 
	errorStats := sm.errorHandler.GetErrorStatistics(time.Hour)
	metrics["errors"] = errorStats
	
	return metrics
}

// ProcessRequest 
func (sm *ServiceManager) ProcessRequest(ctx context.Context, request interface{}) (interface{}, error) {
	if !sm.running {
		return nil, sm.errorHandler.CreateError("service_manager", "process_request", "service manager not running", ErrorTypeService, SeverityHigh)
	}
	
	integration, err := sm.initializer.GetIntegration()
	if err != nil {
		return nil, sm.errorHandler.WrapError(err, "service_manager", "process_request", ErrorTypeIntegration, SeverityHigh, nil)
	}
	
	// ?LearningRequest
	learningRequest, ok := request.(*LearningRequest)
	if !ok {
		return nil, sm.errorHandler.CreateError("service_manager", "process_request", "invalid request type", ErrorTypeValidation, SeverityHigh)
	}
	
	// ?
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

// startHealthCheck ?
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

// performHealthCheck ?
func (sm *ServiceManager) performHealthCheck(ctx context.Context) {
	health := sm.GetHealthStatus()
	
	if status, ok := health["status"].(string); ok && status != "healthy" {
		sm.logger.Warn("Health check failed", map[string]interface{}{
			"status": status,
			"health": health,
		})
		
		// 
		sm.attemptAutoRecovery(ctx, health)
	} else {
		sm.logger.Debug("Health check passed")
	}
}

// attemptAutoRecovery 
func (sm *ServiceManager) attemptAutoRecovery(ctx context.Context, health map[string]interface{}) {
	sm.logger.Info("Attempting auto recovery...")
	
	// 
	// 
	
	if unhealthyServices, ok := health["unhealthy_services"].([]string); ok {
		for _, serviceName := range unhealthyServices {
			sm.logger.Info(fmt.Sprintf("Attempting to recover service: %s", serviceName))
			// 
		}
	}
}

// startMetricsServer ?
func (sm *ServiceManager) startMetricsServer(ctx context.Context) {
	sm.logger.Info(fmt.Sprintf("Metrics server starting on port %d", sm.config.MetricsPort))
	
	// HTTP?
	// Prometheus metrics
	
	select {
	case <-ctx.Done():
		sm.logger.Info("Metrics server stopped")
	}
}

// startProfilingServer ?
func (sm *ServiceManager) startProfilingServer(ctx context.Context) {
	sm.logger.Info(fmt.Sprintf("Profiling server starting on port %d", sm.config.ProfilingPort))
	
	// pprof?
	
	select {
	case <-ctx.Done():
		sm.logger.Info("Profiling server stopped")
	}
}

// RunWithGracefulShutdown 
func (sm *ServiceManager) RunWithGracefulShutdown() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 
	if err := sm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start service manager: %w", err)
	}
	
	// 
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	sm.logger.Info("Service manager is running. Press Ctrl+C to stop.")
	
	// 
	<-sigChan
	sm.logger.Info("Received shutdown signal, starting graceful shutdown...")
	
	// goroutine
	cancel()
	
	// 
	if err := sm.Stop(context.Background()); err != nil {
		return fmt.Errorf("failed to stop service manager: %w", err)
	}
	
	sm.logger.Info("Service manager shutdown completed")
	return nil
}

// GetLogger ?
func (sm *ServiceManager) GetLogger() *Logger {
	return sm.logger
}

// GetErrorHandler ?
func (sm *ServiceManager) GetErrorHandler() *ErrorHandler {
	return sm.errorHandler
}

// GetInitializer 
func (sm *ServiceManager) GetInitializer() *ServiceInitializer {
	return sm.initializer
}

