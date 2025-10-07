package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/alerting"
	"github.com/taishanglaojun/core-services/monitoring/automation"
	"github.com/taishanglaojun/core-services/monitoring/dashboard"
	"github.com/taishanglaojun/core-services/monitoring/logging"
	"github.com/taishanglaojun/core-services/monitoring/performance"
	"github.com/taishanglaojun/core-services/monitoring/tracing"
)

// MonitoringSystem 监控系统集成
type MonitoringSystem struct {
	config           MonitoringConfig
	tracer           *tracing.Tracer
	logManager       *logging.LogManager
	logPipeline      *logging.LogPipeline
	alertManager     *alerting.AlertManager
	dashboardManager *dashboard.DashboardManager
	perfAnalyzer     *performance.PerformanceAnalyzer
	orchestrator     *automation.Orchestrator
	stats            *MonitoringStats
	mutex            sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// MonitoringConfig 监控系统配置
type MonitoringConfig struct {
	ServiceName     string                        `json:"service_name" yaml:"service_name"`
	ServiceVersion  string                        `json:"service_version" yaml:"service_version"`
	Environment     string                        `json:"environment" yaml:"environment"`
	Tracing         tracing.TracerConfig          `json:"tracing" yaml:"tracing"`
	Logging         logging.LogManagerConfig      `json:"logging" yaml:"logging"`
	LogPipeline     logging.LogPipelineConfig     `json:"log_pipeline" yaml:"log_pipeline"`
	Alerting        alerting.AlertManagerConfig   `json:"alerting" yaml:"alerting"`
	Dashboard       dashboard.DashboardConfig     `json:"dashboard" yaml:"dashboard"`
	Performance     performance.AnalyzerConfig    `json:"performance" yaml:"performance"`
	Automation      automation.OrchestratorConfig `json:"automation" yaml:"automation"`
	Integration     IntegrationConfig             `json:"integration" yaml:"integration"`
}

// IntegrationConfig 集成配置
type IntegrationConfig struct {
	EnableTraceLogging    bool          `json:"enable_trace_logging" yaml:"enable_trace_logging"`
	EnableMetricAlerting  bool          `json:"enable_metric_alerting" yaml:"enable_metric_alerting"`
	EnableAutoRemediation bool          `json:"enable_auto_remediation" yaml:"enable_auto_remediation"`
	CorrelationWindow     time.Duration `json:"correlation_window" yaml:"correlation_window"`
	SyncInterval          time.Duration `json:"sync_interval" yaml:"sync_interval"`
	HealthCheckInterval   time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
}

// MonitoringStats 监控系统统计信息
type MonitoringStats struct {
	StartTime         time.Time                      `json:"start_time"`
	Uptime            time.Duration                  `json:"uptime"`
	TracingStats      *tracing.TracerStats           `json:"tracing_stats"`
	LoggingStats      *logging.LogManagerStats       `json:"logging_stats"`
	AlertingStats     *alerting.AlertManagerStats    `json:"alerting_stats"`
	DashboardStats    *dashboard.DashboardStats      `json:"dashboard_stats"`
	PerformanceStats  *performance.AnalyzerStats     `json:"performance_stats"`
	AutomationStats   *automation.OrchestratorStats  `json:"automation_stats"`
	IntegrationStats  *IntegrationStats              `json:"integration_stats"`
}

// IntegrationStats 集成统计信息
type IntegrationStats struct {
	CorrelatedEvents    int64     `json:"correlated_events"`
	AutoRemediations    int64     `json:"auto_remediations"`
	HealthChecks        int64     `json:"health_checks"`
	FailedHealthChecks  int64     `json:"failed_health_checks"`
	LastSync            time.Time `json:"last_sync"`
	LastHealthCheck     time.Time `json:"last_health_check"`
}

// NewMonitoringSystem 创建监控系统
func NewMonitoringSystem(config MonitoringConfig) (*MonitoringSystem, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建追踪器
	tracer, err := tracing.NewTracer(config.Tracing)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}
	
	// 创建日志管理器
	logManager, err := logging.NewLogManager(config.Logging)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create log manager: %w", err)
	}
	
	// 创建日志管道
	logPipeline, err := logging.NewLogPipeline(config.LogPipeline)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create log pipeline: %w", err)
	}
	
	// 创建告警管理器
	alertManager, err := alerting.NewAlertManager(config.Alerting)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create alert manager: %w", err)
	}
	
	// 创建仪表板管理器
	dashboardManager, err := dashboard.NewDashboardManager(config.Dashboard)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create dashboard manager: %w", err)
	}
	
	// 创建性能分析器
	perfAnalyzer, err := performance.NewPerformanceAnalyzer(config.Performance)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create performance analyzer: %w", err)
	}
	
	// 创建自动化编排器
	orchestrator, err := automation.NewOrchestrator(config.Automation)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create orchestrator: %w", err)
	}
	
	system := &MonitoringSystem{
		config:           config,
		tracer:           tracer,
		logManager:       logManager,
		logPipeline:      logPipeline,
		alertManager:     alertManager,
		dashboardManager: dashboardManager,
		perfAnalyzer:     perfAnalyzer,
		orchestrator:     orchestrator,
		stats: &MonitoringStats{
			StartTime: time.Now(),
			IntegrationStats: &IntegrationStats{
				CorrelatedEvents:   0,
				AutoRemediations:   0,
				HealthChecks:       0,
				FailedHealthChecks: 0,
				LastSync:           time.Time{},
				LastHealthCheck:    time.Time{},
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	return system, nil
}

// Start 启动监控系统
func (ms *MonitoringSystem) Start() error {
	// 启动追踪器
	if err := ms.tracer.Start(); err != nil {
		return fmt.Errorf("failed to start tracer: %w", err)
	}
	
	// 启动日志管理器
	if err := ms.logManager.Start(); err != nil {
		return fmt.Errorf("failed to start log manager: %w", err)
	}
	
	// 启动日志管道
	if err := ms.logPipeline.Start(); err != nil {
		return fmt.Errorf("failed to start log pipeline: %w", err)
	}
	
	// 启动告警管理器
	if err := ms.alertManager.Start(); err != nil {
		return fmt.Errorf("failed to start alert manager: %w", err)
	}
	
	// 启动仪表板管理器
	if err := ms.dashboardManager.Start(); err != nil {
		return fmt.Errorf("failed to start dashboard manager: %w", err)
	}
	
	// 启动性能分析器
	if err := ms.perfAnalyzer.Start(); err != nil {
		return fmt.Errorf("failed to start performance analyzer: %w", err)
	}
	
	// 启动自动化编排器
	if err := ms.orchestrator.Start(); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}
	
	// 启动集成服务
	ms.wg.Add(3)
	go ms.correlationLoop()
	go ms.syncLoop()
	go ms.healthCheckLoop()
	
	return nil
}

// Stop 停止监控系统
func (ms *MonitoringSystem) Stop() error {
	ms.cancel()
	ms.wg.Wait()
	
	// 停止各个组件
	if err := ms.orchestrator.Stop(); err != nil {
		return fmt.Errorf("failed to stop orchestrator: %w", err)
	}
	
	if err := ms.perfAnalyzer.Stop(); err != nil {
		return fmt.Errorf("failed to stop performance analyzer: %w", err)
	}
	
	if err := ms.dashboardManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop dashboard manager: %w", err)
	}
	
	if err := ms.alertManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop alert manager: %w", err)
	}
	
	if err := ms.logPipeline.Stop(); err != nil {
		return fmt.Errorf("failed to stop log pipeline: %w", err)
	}
	
	if err := ms.logManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop log manager: %w", err)
	}
	
	if err := ms.tracer.Stop(); err != nil {
		return fmt.Errorf("failed to stop tracer: %w", err)
	}
	
	return nil
}

// GetStats 获取统计信息
func (ms *MonitoringSystem) GetStats() *MonitoringStats {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	
	// 更新运行时间
	ms.stats.Uptime = time.Since(ms.stats.StartTime)
	
	// 获取各组件统计信息
	ms.stats.TracingStats = ms.tracer.GetStats()
	ms.stats.LoggingStats = ms.logManager.GetStats()
	ms.stats.AlertingStats = ms.alertManager.GetStats()
	ms.stats.DashboardStats = ms.dashboardManager.GetStats()
	ms.stats.PerformanceStats = ms.perfAnalyzer.GetStats()
	ms.stats.AutomationStats = ms.orchestrator.GetStats()
	
	// 复制统计信息
	stats := *ms.stats
	integrationStats := *ms.stats.IntegrationStats
	stats.IntegrationStats = &integrationStats
	
	return &stats
}

// HealthCheck 健康检查
func (ms *MonitoringSystem) HealthCheck() error {
	// 检查各个组件
	if err := ms.tracer.HealthCheck(); err != nil {
		return fmt.Errorf("tracer health check failed: %w", err)
	}
	
	if err := ms.logManager.HealthCheck(); err != nil {
		return fmt.Errorf("log manager health check failed: %w", err)
	}
	
	if err := ms.logPipeline.HealthCheck(); err != nil {
		return fmt.Errorf("log pipeline health check failed: %w", err)
	}
	
	if err := ms.alertManager.HealthCheck(); err != nil {
		return fmt.Errorf("alert manager health check failed: %w", err)
	}
	
	if err := ms.dashboardManager.HealthCheck(); err != nil {
		return fmt.Errorf("dashboard manager health check failed: %w", err)
	}
	
	if err := ms.perfAnalyzer.HealthCheck(); err != nil {
		return fmt.Errorf("performance analyzer health check failed: %w", err)
	}
	
	if err := ms.orchestrator.HealthCheck(); err != nil {
		return fmt.Errorf("orchestrator health check failed: %w", err)
	}
	
	return nil
}

// correlationLoop 关联循环
func (ms *MonitoringSystem) correlationLoop() {
	defer ms.wg.Done()
	
	ticker := time.NewTicker(ms.config.Integration.CorrelationWindow)
	defer ticker.Stop()
	
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.correlateEvents()
		}
	}
}

// correlateEvents 关联事件
func (ms *MonitoringSystem) correlateEvents() {
	if !ms.config.Integration.EnableTraceLogging {
		return
	}
	
	// 获取最近的追踪数据
	spans := ms.tracer.GetRecentSpans(ms.config.Integration.CorrelationWindow)
	
	// 为每个span创建日志条目
	for _, span := range spans {
		logEntry := &logging.LogEntry{
			ID:        fmt.Sprintf("trace-%s", span.SpanID),
			Timestamp: span.StartTime,
			Level:     logging.LogLevelInfo,
			Message:   fmt.Sprintf("Trace span: %s", span.OperationName),
			Source:    "tracer",
			Service:   ms.config.ServiceName,
			TraceID:   span.TraceID,
			SpanID:    span.SpanID,
			Fields: map[string]interface{}{
				"operation_name": span.OperationName,
				"duration":       span.Duration,
				"status":         span.Status,
			},
			Tags: span.Tags,
		}
		
		// 发送到日志管道
		ms.logPipeline.Input(logEntry)
	}
	
	ms.mutex.Lock()
	ms.stats.IntegrationStats.CorrelatedEvents += int64(len(spans))
	ms.mutex.Unlock()
}

// syncLoop 同步循环
func (ms *MonitoringSystem) syncLoop() {
	defer ms.wg.Done()
	
	ticker := time.NewTicker(ms.config.Integration.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.syncComponents()
		}
	}
}

// syncComponents 同步组件
func (ms *MonitoringSystem) syncComponents() {
	// 同步性能指标到告警系统
	if ms.config.Integration.EnableMetricAlerting {
		metrics := ms.perfAnalyzer.GetSystemMetrics()
		ms.checkMetricAlerts(metrics)
	}
	
	// 同步告警到自动化系统
	if ms.config.Integration.EnableAutoRemediation {
		alerts := ms.alertManager.GetActiveAlerts()
		ms.triggerAutoRemediation(alerts)
	}
	
	ms.mutex.Lock()
	ms.stats.IntegrationStats.LastSync = time.Now()
	ms.mutex.Unlock()
}

// checkMetricAlerts 检查指标告警
func (ms *MonitoringSystem) checkMetricAlerts(metrics *performance.SystemMetrics) {
	// CPU使用率告警
	if metrics.CPU.Usage > 80.0 {
		alert := &alerting.Alert{
			ID:          fmt.Sprintf("cpu-high-%d", time.Now().Unix()),
			Name:        "High CPU Usage",
			Description: fmt.Sprintf("CPU usage is %.2f%%", metrics.CPU.Usage),
			Severity:    alerting.SeverityWarning,
			Status:      alerting.StatusFiring,
			Source:      "performance_analyzer",
			Service:     ms.config.ServiceName,
			CreatedAt:   time.Now(),
			Labels: map[string]string{
				"metric": "cpu_usage",
				"value":  fmt.Sprintf("%.2f", metrics.CPU.Usage),
			},
		}
		ms.alertManager.CreateAlert(alert)
	}
	
	// 内存使用率告警
	if metrics.Memory.Usage > 85.0 {
		alert := &alerting.Alert{
			ID:          fmt.Sprintf("memory-high-%d", time.Now().Unix()),
			Name:        "High Memory Usage",
			Description: fmt.Sprintf("Memory usage is %.2f%%", metrics.Memory.Usage),
			Severity:    alerting.SeverityWarning,
			Status:      alerting.StatusFiring,
			Source:      "performance_analyzer",
			Service:     ms.config.ServiceName,
			CreatedAt:   time.Now(),
			Labels: map[string]string{
				"metric": "memory_usage",
				"value":  fmt.Sprintf("%.2f", metrics.Memory.Usage),
			},
		}
		ms.alertManager.CreateAlert(alert)
	}
}

// triggerAutoRemediation 触发自动修复
func (ms *MonitoringSystem) triggerAutoRemediation(alerts []*alerting.Alert) {
	for _, alert := range alerts {
		if alert.Severity == alerting.SeverityCritical {
			// 创建自动修复工作流
			workflow := ms.createRemediationWorkflow(alert)
			if workflow != nil {
				ms.orchestrator.CreateWorkflow(workflow)
				
				ms.mutex.Lock()
				ms.stats.IntegrationStats.AutoRemediations++
				ms.mutex.Unlock()
			}
		}
	}
}

// createRemediationWorkflow 创建修复工作流
func (ms *MonitoringSystem) createRemediationWorkflow(alert *alerting.Alert) *automation.Workflow {
	// 根据告警类型创建不同的修复工作流
	switch alert.Labels["metric"] {
	case "cpu_usage":
		return ms.createCPURemediationWorkflow(alert)
	case "memory_usage":
		return ms.createMemoryRemediationWorkflow(alert)
	default:
		return nil
	}
}

// createCPURemediationWorkflow 创建CPU修复工作流
func (ms *MonitoringSystem) createCPURemediationWorkflow(alert *alerting.Alert) *automation.Workflow {
	workflow := &automation.Workflow{
		ID:          fmt.Sprintf("cpu-remediation-%s", alert.ID),
		Name:        "CPU High Usage Remediation",
		Description: "Automatic remediation for high CPU usage",
		Definition: automation.WorkflowDefinition{
			Name:        "CPU Remediation",
			Description: "Scale up resources or restart services",
			Version:     "1.0",
			Tasks: []automation.TaskDefinition{
				{
					Name:        "scale_up",
					Type:        "kubernetes",
					Description: "Scale up the deployment",
					Command:     "kubectl scale deployment myapp --replicas=3",
					Timeout:     5 * time.Minute,
				},
			},
			Timeout: 10 * time.Minute,
		},
		Status:    automation.WorkflowStatusPending,
		Tasks:     make(map[string]*automation.Task),
		Variables: make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
	
	return workflow
}

// createMemoryRemediationWorkflow 创建内存修复工作流
func (ms *MonitoringSystem) createMemoryRemediationWorkflow(alert *alerting.Alert) *automation.Workflow {
	workflow := &automation.Workflow{
		ID:          fmt.Sprintf("memory-remediation-%s", alert.ID),
		Name:        "Memory High Usage Remediation",
		Description: "Automatic remediation for high memory usage",
		Definition: automation.WorkflowDefinition{
			Name:        "Memory Remediation",
			Description: "Restart services or clear cache",
			Version:     "1.0",
			Tasks: []automation.TaskDefinition{
				{
					Name:        "restart_service",
					Type:        "shell",
					Description: "Restart the service",
					Command:     "systemctl restart myapp",
					Timeout:     2 * time.Minute,
				},
			},
			Timeout: 5 * time.Minute,
		},
		Status:    automation.WorkflowStatusPending,
		Tasks:     make(map[string]*automation.Task),
		Variables: make(map[string]interface{}),
		CreatedAt: time.Now(),
	}
	
	return workflow
}

// healthCheckLoop 健康检查循环
func (ms *MonitoringSystem) healthCheckLoop() {
	defer ms.wg.Done()
	
	ticker := time.NewTicker(ms.config.Integration.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			ms.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (ms *MonitoringSystem) performHealthCheck() {
	ms.mutex.Lock()
	ms.stats.IntegrationStats.HealthChecks++
	ms.stats.IntegrationStats.LastHealthCheck = time.Now()
	ms.mutex.Unlock()
	
	if err := ms.HealthCheck(); err != nil {
		ms.mutex.Lock()
		ms.stats.IntegrationStats.FailedHealthChecks++
		ms.mutex.Unlock()
		
		// 创建健康检查失败告警
		alert := &alerting.Alert{
			ID:          fmt.Sprintf("health-check-failed-%d", time.Now().Unix()),
			Name:        "Health Check Failed",
			Description: fmt.Sprintf("Monitoring system health check failed: %v", err),
			Severity:    alerting.SeverityCritical,
			Status:      alerting.StatusFiring,
			Source:      "monitoring_system",
			Service:     ms.config.ServiceName,
			CreatedAt:   time.Now(),
			Labels: map[string]string{
				"component": "monitoring_system",
				"error":     err.Error(),
			},
		}
		ms.alertManager.CreateAlert(alert)
	}
}

// GetTracer 获取追踪器
func (ms *MonitoringSystem) GetTracer() *tracing.Tracer {
	return ms.tracer
}

// GetLogManager 获取日志管理器
func (ms *MonitoringSystem) GetLogManager() *logging.LogManager {
	return ms.logManager
}

// GetAlertManager 获取告警管理器
func (ms *MonitoringSystem) GetAlertManager() *alerting.AlertManager {
	return ms.alertManager
}

// GetDashboardManager 获取仪表板管理器
func (ms *MonitoringSystem) GetDashboardManager() *dashboard.DashboardManager {
	return ms.dashboardManager
}

// GetPerformanceAnalyzer 获取性能分析器
func (ms *MonitoringSystem) GetPerformanceAnalyzer() *performance.PerformanceAnalyzer {
	return ms.perfAnalyzer
}

// GetOrchestrator 获取自动化编排器
func (ms *MonitoringSystem) GetOrchestrator() *automation.Orchestrator {
	return ms.orchestrator
}