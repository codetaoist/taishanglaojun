package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/alerting"
	"github.com/taishanglaojun/core-services/monitoring/automation"
	"github.com/taishanglaojun/core-services/monitoring/dashboard"
	"github.com/taishanglaojun/core-services/monitoring/integration"
	"github.com/taishanglaojun/core-services/monitoring/logging"
	"github.com/taishanglaojun/core-services/monitoring/performance"
	"github.com/taishanglaojun/core-services/monitoring/tracing"
	"gopkg.in/yaml.v2"
)

var (
	configFile = flag.String("config", "config/monitoring.yaml", "配置文件路径")
	logLevel   = flag.String("log-level", "info", "日志级别 (trace, debug, info, warn, error, fatal)")
	version    = flag.Bool("version", false, "显示版本信息")
)

const (
	Version     = "1.0.0"
	ServiceName = "taishanglaojun-monitoring"
)

// Config 监控系统配置
type Config struct {
	Service     ServiceConfig                    `yaml:"service"`
	Tracing     tracing.TracerConfig             `yaml:"tracing"`
	Logging     logging.LogManagerConfig         `yaml:"logging"`
	LogPipeline logging.LogPipelineConfig        `yaml:"log_pipeline"`
	Alerting    alerting.AlertManagerConfig      `yaml:"alerting"`
	Dashboard   dashboard.DashboardConfig        `yaml:"dashboard"`
	Performance performance.AnalyzerConfig      `yaml:"performance"`
	Automation  automation.OrchestratorConfig    `yaml:"automation"`
	Integration integration.IntegrationConfig   `yaml:"integration"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Port        int    `yaml:"port"`
	LogLevel    string `yaml:"log_level"`
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("配置文件 %s 不存在，创建默认配置", configPath)
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置失败: %v", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	setDefaultConfig(&config)

	return &config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	config := getDefaultConfig()

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	log.Printf("默认配置已创建: %s", configPath)
	return nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() Config {
	return Config{
		Service: ServiceConfig{
			Name:        ServiceName,
			Version:     Version,
			Environment: "development",
			Port:        8080,
			LogLevel:    "info",
		},
		Tracing: tracing.TracerConfig{
			ServiceName:    ServiceName,
			ServiceVersion: Version,
			Environment:    "development",
			Sampling: tracing.SamplingConfig{
				Type: "probabilistic",
				Rate: 0.1,
			},
			Batching: tracing.BatchingConfig{
				MaxBatchSize: 512,
				FlushTimeout: 5 * time.Second,
			},
			Exporters: []tracing.ExporterConfig{
				{
					Type:    "console",
					Enabled: true,
				},
				{
					Type:    "jaeger",
					Enabled: false,
					Settings: map[string]interface{}{
						"endpoint": "http://localhost:14268/api/traces",
					},
				},
			},
		},
		Logging: logging.LogManagerConfig{
			BufferSize:   10000,
			FlushTimeout: 5 * time.Second,
			Workers:      4,
			BatchSize:    100,
		},
		LogPipeline: logging.LogPipelineConfig{
			BufferSize:     10000,
			Workers:        4,
			BatchSize:      100,
			FlushInterval:  5 * time.Second,
			ProcessTimeout: 30 * time.Second,
			RetryPolicy: logging.RetryPolicy{
				MaxAttempts: 3,
				Delay:       time.Second,
				BackoffType: "exponential",
			},
		},
		Alerting: alerting.AlertManagerConfig{
			BufferSize:      1000,
			FlushTimeout:    5 * time.Second,
			Workers:         2,
			BatchSize:       50,
			RetentionPeriod: 7 * 24 * time.Hour,
			NotificationChannels: []alerting.NotificationChannelConfig{
				{
					Type:    "console",
					Enabled: true,
				},
				{
					Type:    "webhook",
					Enabled: false,
					Settings: map[string]interface{}{
						"url": "http://localhost:8081/alerts",
					},
				},
			},
		},
		Dashboard: dashboard.DashboardConfig{
			Port:            8080,
			UpdateInterval:  10 * time.Second,
			CacheSize:       1000,
			CacheTTL:        5 * time.Minute,
			MaxConnections:  100,
			RequestTimeout:  30 * time.Second,
			EnableMetrics:   true,
			EnableProfiling: false,
		},
		Performance: performance.AnalyzerConfig{
			CollectionInterval: 10 * time.Second,
			AnalysisInterval:   time.Minute,
			StorageRetention:   24 * time.Hour,
			AlertThresholds: map[string]float64{
				"cpu_usage":    80.0,
				"memory_usage": 85.0,
				"disk_usage":   90.0,
			},
			Collectors: []performance.CollectorConfig{
				{Type: "cpu", Enabled: true},
				{Type: "memory", Enabled: true},
				{Type: "disk", Enabled: true},
				{Type: "network", Enabled: true},
				{Type: "process", Enabled: true},
			},
		},
		Automation: automation.OrchestratorConfig{
			MaxConcurrentWorkflows: 10,
			MaxConcurrentTasks:     50,
			WorkflowTimeout:        30 * time.Minute,
			TaskTimeout:            10 * time.Minute,
			SchedulingInterval:     time.Minute,
			RetryPolicy: automation.RetryPolicy{
				MaxAttempts: 3,
				Delay:       5 * time.Second,
				BackoffType: "exponential",
			},
			Storage: automation.StorageConfig{
				Type: "memory",
			},
			HistoryRetention: 7 * 24 * time.Hour,
			Executors: []automation.ExecutorConfig{
				{Type: "shell", Enabled: true},
				{Type: "http", Enabled: true},
				{Type: "script", Enabled: true},
			},
		},
		Integration: integration.IntegrationConfig{
			EnableTraceLogging:    true,
			EnableMetricAlerting:  true,
			EnableAutoRemediation: true,
			CorrelationWindow:     5 * time.Minute,
			SyncInterval:          30 * time.Second,
			HealthCheckInterval:   time.Minute,
		},
	}
}

// setDefaultConfig 设置默认配置值
func setDefaultConfig(config *Config) {
	if config.Service.Name == "" {
		config.Service.Name = ServiceName
	}
	if config.Service.Version == "" {
		config.Service.Version = Version
	}
	if config.Service.Environment == "" {
		config.Service.Environment = "development"
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8080
	}
	if config.Service.LogLevel == "" {
		config.Service.LogLevel = "info"
	}

	// 设置追踪配置默认值
	config.Tracing.ServiceName = config.Service.Name
	config.Tracing.ServiceVersion = config.Service.Version
	config.Tracing.Environment = config.Service.Environment
}

// createMonitoringConfig 创建监控系统配置
func createMonitoringConfig(config *Config) integration.MonitoringConfig {
	return integration.MonitoringConfig{
		ServiceName:    config.Service.Name,
		ServiceVersion: config.Service.Version,
		Environment:    config.Service.Environment,
		Tracing:        config.Tracing,
		Logging:        config.Logging,
		LogPipeline:    config.LogPipeline,
		Alerting:       config.Alerting,
		Dashboard:      config.Dashboard,
		Performance:    config.Performance,
		Automation:     config.Automation,
		Integration:    config.Integration,
	}
}

// setupSignalHandling 设置信号处理
func setupSignalHandling(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			log.Printf("收到信号 %v，开始优雅关闭...", sig)
			cancel()
		case <-ctx.Done():
			return
		}
	}()
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Printf("%s version %s\n", ServiceName, Version)
	fmt.Printf("Build time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("Go version: %s\n", "go1.21")
}

func main() {
	flag.Parse()

	if *version {
		printVersion()
		return
	}

	// 加载配置
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 覆盖日志级别
	if *logLevel != "" {
		config.Service.LogLevel = *logLevel
	}

	log.Printf("启动 %s v%s (环境: %s)", config.Service.Name, config.Service.Version, config.Service.Environment)
	log.Printf("配置文件: %s", *configFile)
	log.Printf("日志级别: %s", config.Service.LogLevel)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 设置信号处理
	setupSignalHandling(ctx, cancel)

	// 创建监控系统
	monitoringConfig := createMonitoringConfig(config)
	monitoringSystem, err := integration.NewMonitoringSystem(monitoringConfig)
	if err != nil {
		log.Fatalf("创建监控系统失败: %v", err)
	}

	// 启动监控系统
	log.Println("启动监控系统...")
	if err := monitoringSystem.Start(); err != nil {
		log.Fatalf("启动监控系统失败: %v", err)
	}

	log.Printf("监控系统已启动，仪表板地址: http://localhost:%d", config.Dashboard.Port)
	log.Println("按 Ctrl+C 停止服务")

	// 定期打印统计信息
	statsTicker := time.NewTicker(30 * time.Second)
	defer statsTicker.Stop()

	// 定期健康检查
	healthTicker := time.NewTicker(time.Minute)
	defer healthTicker.Stop()

	// 主循环
	for {
		select {
		case <-ctx.Done():
			log.Println("收到停止信号，开始关闭监控系统...")
			
			// 优雅关闭
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()

			done := make(chan error, 1)
			go func() {
				done <- monitoringSystem.Stop()
			}()

			select {
			case err := <-done:
				if err != nil {
					log.Printf("关闭监控系统时出错: %v", err)
				} else {
					log.Println("监控系统已成功关闭")
				}
			case <-shutdownCtx.Done():
				log.Println("关闭超时，强制退出")
			}

			return

		case <-statsTicker.C:
			// 打印统计信息
			stats := monitoringSystem.GetStats()
			log.Printf("系统统计 - 追踪: %d spans, 日志: %d entries, 告警: %d alerts, 工作流: %d workflows",
				stats.TracingStats.SpansCreated,
				stats.LoggingStats.ProcessedLogs,
				stats.AlertingStats.TotalAlerts,
				stats.AutomationStats.TotalWorkflows)

		case <-healthTicker.C:
			// 健康检查
			if err := monitoringSystem.HealthCheck(); err != nil {
				log.Printf("健康检查失败: %v", err)
			} else {
				log.Println("健康检查通过")
			}
		}
	}
}