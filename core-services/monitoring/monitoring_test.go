package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taishanglaojun/core-services/monitoring/alerting"
	"github.com/taishanglaojun/core-services/monitoring/automation"
	"github.com/taishanglaojun/core-services/monitoring/dashboard"
	"github.com/taishanglaojun/core-services/monitoring/integration"
	"github.com/taishanglaojun/core-services/monitoring/logging"
	"github.com/taishanglaojun/core-services/monitoring/performance"
	"github.com/taishanglaojun/core-services/monitoring/tracing"
)

func TestTracingSystem(t *testing.T) {
	config := tracing.TracerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Sampling: tracing.SamplingConfig{
			Type: "always",
		},
		Batching: tracing.BatchingConfig{
			MaxBatchSize: 100,
			FlushTimeout: time.Second,
		},
		Exporters: []tracing.ExporterConfig{
			{
				Type:    "console",
				Enabled: true,
			},
		},
	}

	tracer, err := tracing.NewTracer(config)
	require.NoError(t, err)
	require.NotNil(t, tracer)

	err = tracer.Start()
	require.NoError(t, err)

	// 测试创建span
	span := tracer.StartSpan("test-operation")
	require.NotNil(t, span)
	assert.Equal(t, "test-operation", span.OperationName)
	assert.NotEmpty(t, span.SpanID)
	assert.NotEmpty(t, span.TraceID)

	// 测试添加标签和日志
	span.SetTag("test.key", "test.value")
	span.LogFields(map[string]interface{}{
		"event": "test-event",
		"level": "info",
	})

	// 测试完成span
	tracer.FinishSpan(span)
	assert.Equal(t, tracing.SpanStatusOK, span.Status)
	assert.True(t, span.FinishTime.After(span.StartTime))

	// 测试统计信息
	stats := tracer.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.SpansCreated)
	assert.Equal(t, int64(1), stats.SpansFinished)

	err = tracer.Stop()
	require.NoError(t, err)
}

func TestLoggingSystem(t *testing.T) {
	config := logging.LogManagerConfig{
		BufferSize:   1000,
		FlushTimeout: time.Second,
		Workers:      2,
		BatchSize:    10,
	}

	logManager, err := logging.NewLogManager(config)
	require.NoError(t, err)
	require.NotNil(t, logManager)

	err = logManager.Start()
	require.NoError(t, err)

	// 测试添加收集器
	collectorConfig := logging.CollectorConfig{
		Type:    "console",
		Enabled: true,
		Settings: map[string]interface{}{
			"format": "json",
		},
	}

	collector, err := logging.CreateCollector(collectorConfig)
	require.NoError(t, err)

	err = logManager.AddCollector("test-collector", collector)
	require.NoError(t, err)

	// 测试添加处理器
	processorConfig := logging.ProcessorConfig{
		Type:    "filter",
		Enabled: true,
		Settings: map[string]interface{}{
			"level": "info",
		},
	}

	processor, err := logging.CreateProcessor(processorConfig)
	require.NoError(t, err)

	err = logManager.AddProcessor("test-processor", processor)
	require.NoError(t, err)

	// 测试添加输出
	outputConfig := logging.OutputConfig{
		Type:    "console",
		Enabled: true,
		Settings: map[string]interface{}{
			"format": "text",
		},
	}

	output, err := logging.CreateOutput(outputConfig)
	require.NoError(t, err)

	err = logManager.AddOutput("test-output", output)
	require.NoError(t, err)

	// 测试统计信息
	stats := logManager.GetStats()
	assert.NotNil(t, stats)
	assert.Len(t, stats.Collectors, 1)
	assert.Len(t, stats.Processors, 1)
	assert.Len(t, stats.Outputs, 1)

	err = logManager.Stop()
	require.NoError(t, err)
}

func TestAlertingSystem(t *testing.T) {
	config := alerting.AlertManagerConfig{
		BufferSize:      1000,
		FlushTimeout:    time.Second,
		Workers:         2,
		BatchSize:       10,
		RetentionPeriod: 24 * time.Hour,
	}

	alertManager, err := alerting.NewAlertManager(config)
	require.NoError(t, err)
	require.NotNil(t, alertManager)

	err = alertManager.Start()
	require.NoError(t, err)

	// 测试创建告警
	alert := &alerting.Alert{
		ID:          "test-alert-1",
		Name:        "Test Alert",
		Description: "This is a test alert",
		Severity:    alerting.SeverityWarning,
		Status:      alerting.StatusFiring,
		Source:      "test",
		Service:     "test-service",
		CreatedAt:   time.Now(),
		Labels: map[string]string{
			"component": "test",
		},
	}

	err = alertManager.CreateAlert(alert)
	require.NoError(t, err)

	// 测试获取告警
	retrievedAlert, err := alertManager.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.Equal(t, alert.ID, retrievedAlert.ID)
	assert.Equal(t, alert.Name, retrievedAlert.Name)

	// 测试更新告警
	alert.Status = alerting.StatusResolved
	err = alertManager.UpdateAlert(alert)
	require.NoError(t, err)

	// 测试列出告警
	alerts, err := alertManager.ListAlerts(alerting.AlertFilter{})
	require.NoError(t, err)
	assert.Len(t, alerts, 1)

	// 测试统计信息
	stats := alertManager.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.TotalAlerts)

	err = alertManager.Stop()
	require.NoError(t, err)
}

func TestPerformanceAnalyzer(t *testing.T) {
	config := performance.AnalyzerConfig{
		CollectionInterval: time.Second,
		AnalysisInterval:   5 * time.Second,
		StorageRetention:   time.Hour,
		AlertThresholds: map[string]float64{
			"cpu_usage":    80.0,
			"memory_usage": 85.0,
		},
	}

	analyzer, err := performance.NewPerformanceAnalyzer(config)
	require.NoError(t, err)
	require.NotNil(t, analyzer)

	err = analyzer.Start()
	require.NoError(t, err)

	// 等待一些数据收集
	time.Sleep(2 * time.Second)

	// 测试获取系统指标
	metrics := analyzer.GetSystemMetrics()
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.CPU)
	assert.NotNil(t, metrics.Memory)
	assert.NotNil(t, metrics.Disk)
	assert.NotNil(t, metrics.Network)

	// 测试性能分析
	result := analyzer.AnalyzePerformance(time.Now().Add(-time.Minute), time.Now())
	assert.NotNil(t, result)

	// 测试统计信息
	stats := analyzer.GetStats()
	assert.NotNil(t, stats)
	assert.True(t, stats.CollectedMetrics > 0)

	err = analyzer.Stop()
	require.NoError(t, err)
}

func TestAutomationOrchestrator(t *testing.T) {
	config := automation.OrchestratorConfig{
		MaxConcurrentWorkflows: 10,
		MaxConcurrentTasks:     20,
		WorkflowTimeout:        10 * time.Minute,
		TaskTimeout:            5 * time.Minute,
		RetryPolicy: automation.RetryPolicy{
			MaxAttempts: 3,
			Delay:       time.Second,
			BackoffType: "exponential",
		},
	}

	orchestrator, err := automation.NewOrchestrator(config)
	require.NoError(t, err)
	require.NotNil(t, orchestrator)

	err = orchestrator.Start()
	require.NoError(t, err)

	// 测试创建工作流
	workflow := &automation.Workflow{
		ID:          "test-workflow-1",
		Name:        "Test Workflow",
		Description: "This is a test workflow",
		Definition: automation.WorkflowDefinition{
			Name:        "Test Workflow",
			Description: "Test workflow definition",
			Version:     "1.0",
			Tasks: []automation.TaskDefinition{
				{
					Name:        "test-task",
					Type:        "shell",
					Description: "Test task",
					Command:     "echo 'Hello World'",
					Timeout:     time.Minute,
				},
			},
			Timeout: 5 * time.Minute,
		},
		Status:    automation.WorkflowStatusPending,
		Tasks:     make(map[string]*automation.Task),
		Variables: make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	err = orchestrator.CreateWorkflow(workflow)
	require.NoError(t, err)

	// 测试获取工作流
	retrievedWorkflow, err := orchestrator.GetWorkflow(workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, workflow.ID, retrievedWorkflow.ID)
	assert.Equal(t, workflow.Name, retrievedWorkflow.Name)

	// 测试列出工作流
	workflows, err := orchestrator.ListWorkflows(automation.WorkflowFilter{})
	require.NoError(t, err)
	assert.Len(t, workflows, 1)

	// 测试统计信息
	stats := orchestrator.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.TotalWorkflows)

	err = orchestrator.Stop()
	require.NoError(t, err)
}

func TestDashboardSystem(t *testing.T) {
	config := dashboard.DashboardConfig{
		Port:            8080,
		UpdateInterval:  time.Second,
		CacheSize:       1000,
		CacheTTL:        5 * time.Minute,
		MaxConnections:  100,
		RequestTimeout:  30 * time.Second,
		EnableMetrics:   true,
		EnableProfiling: false,
	}

	dashboardManager, err := dashboard.NewDashboardManager(config)
	require.NoError(t, err)
	require.NotNil(t, dashboardManager)

	err = dashboardManager.Start()
	require.NoError(t, err)

	// 测试创建仪表板
	dashboardDef := &dashboard.Dashboard{
		ID:          "test-dashboard-1",
		Name:        "Test Dashboard",
		Description: "This is a test dashboard",
		Layout: dashboard.Layout{
			Rows: []dashboard.Row{
				{
					Height: 300,
					Panels: []dashboard.Panel{
						{
							ID:     "panel-1",
							Type:   "metric",
							Title:  "CPU Usage",
							Width:  6,
							Height: 300,
							Config: map[string]interface{}{
								"metric": "cpu_usage",
							},
						},
					},
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = dashboardManager.CreateDashboard(dashboardDef)
	require.NoError(t, err)

	// 测试获取仪表板
	retrievedDashboard, err := dashboardManager.GetDashboard(dashboardDef.ID)
	require.NoError(t, err)
	assert.Equal(t, dashboardDef.ID, retrievedDashboard.ID)
	assert.Equal(t, dashboardDef.Name, retrievedDashboard.Name)

	// 测试列出仪表板
	dashboards, err := dashboardManager.ListDashboards()
	require.NoError(t, err)
	assert.Len(t, dashboards, 1)

	// 测试统计信息
	stats := dashboardManager.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.TotalDashboards)

	err = dashboardManager.Stop()
	require.NoError(t, err)
}

func TestMonitoringSystemIntegration(t *testing.T) {
	config := integration.MonitoringConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Tracing: tracing.TracerConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Environment:    "test",
			Sampling: tracing.SamplingConfig{
				Type: "always",
			},
			Batching: tracing.BatchingConfig{
				MaxBatchSize: 100,
				FlushTimeout: time.Second,
			},
			Exporters: []tracing.ExporterConfig{
				{
					Type:    "console",
					Enabled: true,
				},
			},
		},
		Logging: logging.LogManagerConfig{
			BufferSize:   1000,
			FlushTimeout: time.Second,
			Workers:      2,
			BatchSize:    10,
		},
		LogPipeline: logging.LogPipelineConfig{
			BufferSize:     1000,
			Workers:        2,
			BatchSize:      10,
			FlushInterval:  time.Second,
			ProcessTimeout: 30 * time.Second,
		},
		Alerting: alerting.AlertManagerConfig{
			BufferSize:      1000,
			FlushTimeout:    time.Second,
			Workers:         2,
			BatchSize:       10,
			RetentionPeriod: 24 * time.Hour,
		},
		Dashboard: dashboard.DashboardConfig{
			Port:            8080,
			UpdateInterval:  time.Second,
			CacheSize:       1000,
			CacheTTL:        5 * time.Minute,
			MaxConnections:  100,
			RequestTimeout:  30 * time.Second,
			EnableMetrics:   true,
			EnableProfiling: false,
		},
		Performance: performance.AnalyzerConfig{
			CollectionInterval: time.Second,
			AnalysisInterval:   5 * time.Second,
			StorageRetention:   time.Hour,
			AlertThresholds: map[string]float64{
				"cpu_usage":    80.0,
				"memory_usage": 85.0,
			},
		},
		Automation: automation.OrchestratorConfig{
			MaxConcurrentWorkflows: 10,
			MaxConcurrentTasks:     20,
			WorkflowTimeout:        10 * time.Minute,
			TaskTimeout:            5 * time.Minute,
			RetryPolicy: automation.RetryPolicy{
				MaxAttempts: 3,
				Delay:       time.Second,
				BackoffType: "exponential",
			},
		},
		Integration: integration.IntegrationConfig{
			EnableTraceLogging:    true,
			EnableMetricAlerting:  true,
			EnableAutoRemediation: true,
			CorrelationWindow:     time.Minute,
			SyncInterval:          30 * time.Second,
			HealthCheckInterval:   time.Minute,
		},
	}

	monitoringSystem, err := integration.NewMonitoringSystem(config)
	require.NoError(t, err)
	require.NotNil(t, monitoringSystem)

	err = monitoringSystem.Start()
	require.NoError(t, err)

	// 等待系统初始化
	time.Sleep(2 * time.Second)

	// 测试健康检查
	err = monitoringSystem.HealthCheck()
	assert.NoError(t, err)

	// 测试统计信息
	stats := monitoringSystem.GetStats()
	assert.NotNil(t, stats)
	assert.NotNil(t, stats.TracingStats)
	assert.NotNil(t, stats.LoggingStats)
	assert.NotNil(t, stats.AlertingStats)
	assert.NotNil(t, stats.DashboardStats)
	assert.NotNil(t, stats.PerformanceStats)
	assert.NotNil(t, stats.AutomationStats)
	assert.NotNil(t, stats.IntegrationStats)

	// 测试组件访问
	tracer := monitoringSystem.GetTracer()
	assert.NotNil(t, tracer)

	logManager := monitoringSystem.GetLogManager()
	assert.NotNil(t, logManager)

	alertManager := monitoringSystem.GetAlertManager()
	assert.NotNil(t, alertManager)

	dashboardManager := monitoringSystem.GetDashboardManager()
	assert.NotNil(t, dashboardManager)

	perfAnalyzer := monitoringSystem.GetPerformanceAnalyzer()
	assert.NotNil(t, perfAnalyzer)

	orchestrator := monitoringSystem.GetOrchestrator()
	assert.NotNil(t, orchestrator)

	err = monitoringSystem.Stop()
	require.NoError(t, err)
}

func TestExecutors(t *testing.T) {
	// 测试Shell执行器
	t.Run("ShellExecutor", func(t *testing.T) {
		config := automation.ExecutorConfig{
			Type:    "shell",
			Enabled: true,
		}

		executor, err := automation.CreateExecutor(config)
		require.NoError(t, err)
		require.NotNil(t, executor)

		err = executor.Start()
		require.NoError(t, err)

		task := &automation.Task{
			ID:   "test-task-1",
			Name: "Test Shell Task",
			Type: "shell",
			Definition: automation.TaskDefinition{
				Command: "echo 'Hello World'",
				Timeout: time.Minute,
			},
			Inputs: make(map[string]interface{}),
			Logs:   make([]string, 0),
		}

		err = executor.ExecuteTask(task, task.Inputs)
		assert.NoError(t, err)
		assert.NotEmpty(t, task.Logs)

		stats := executor.GetStats()
		assert.NotNil(t, stats)
		assert.Equal(t, int64(1), stats.ExecutedTasks)

		err = executor.Stop()
		require.NoError(t, err)
	})

	// 测试HTTP执行器
	t.Run("HTTPExecutor", func(t *testing.T) {
		config := automation.ExecutorConfig{
			Type:    "http",
			Enabled: true,
		}

		executor, err := automation.CreateExecutor(config)
		require.NoError(t, err)
		require.NotNil(t, executor)

		err = executor.Start()
		require.NoError(t, err)

		task := &automation.Task{
			ID:   "test-task-2",
			Name: "Test HTTP Task",
			Type: "http",
			Definition: automation.TaskDefinition{
				Command: "https://httpbin.org/get",
				Environment: map[string]string{
					"METHOD": "GET",
				},
				Timeout: time.Minute,
			},
			Inputs: make(map[string]interface{}),
			Logs:   make([]string, 0),
		}

		err = executor.ExecuteTask(task, task.Inputs)
		// HTTP请求可能失败，但不应该panic
		assert.NotPanics(t, func() {
			executor.ExecuteTask(task, task.Inputs)
		})

		stats := executor.GetStats()
		assert.NotNil(t, stats)

		err = executor.Stop()
		require.NoError(t, err)
	})
}

func BenchmarkTracingSpanCreation(b *testing.B) {
	config := tracing.TracerConfig{
		ServiceName:    "benchmark-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Sampling: tracing.SamplingConfig{
			Type: "never", // 不导出，只测试创建性能
		},
	}

	tracer, err := tracing.NewTracer(config)
	require.NoError(b, err)

	err = tracer.Start()
	require.NoError(b, err)
	defer tracer.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			span := tracer.StartSpan("benchmark-operation")
			span.SetTag("benchmark", true)
			tracer.FinishSpan(span)
		}
	})
}

func BenchmarkLoggingThroughput(b *testing.B) {
	config := logging.LogPipelineConfig{
		BufferSize:     10000,
		Workers:        4,
		BatchSize:      100,
		FlushInterval:  time.Second,
		ProcessTimeout: 30 * time.Second,
	}

	pipeline, err := logging.NewLogPipeline(config)
	require.NoError(b, err)

	err = pipeline.Start()
	require.NoError(b, err)
	defer pipeline.Stop()

	// 添加一个简单的输出
	outputConfig := logging.OutputConfig{
		Type:    "console",
		Enabled: true,
	}
	output, err := logging.CreateOutput(outputConfig)
	require.NoError(b, err)
	pipeline.AddOutput("benchmark-output", output)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logEntry := &logging.LogEntry{
				ID:        "benchmark-log",
				Timestamp: time.Now(),
				Level:     logging.LogLevelInfo,
				Message:   "Benchmark log message",
				Source:    "benchmark",
				Service:   "benchmark-service",
				Fields: map[string]interface{}{
					"benchmark": true,
					"iteration": b.N,
				},
			}
			pipeline.Input(logEntry)
		}
	})
}