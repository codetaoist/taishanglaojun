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

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/alerting"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/automation"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/dashboard"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/integration"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/logging"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/performance"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/tracing"
	"gopkg.in/yaml.v2"
)

var (
	configFile = flag.String("config", "config/monitoring.yaml", "")
	logLevel   = flag.String("log-level", "info", " (trace, debug, info, warn, error, fatal)")
	version    = flag.Bool("version", false, "汾")
)

const (
	Version     = "1.0.0"
	ServiceName = "taishanglaojun-monitoring"
)

// Config 
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

// ServiceConfig 
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Port        int    `yaml:"port"`
	LogLevel    string `yaml:"log_level"`
}

// loadConfig 
func loadConfig(configPath string) (*Config, error) {
	// 
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf(" %s ", configPath)
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf(": %v", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf(": %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf(": %v", err)
	}

	// ?
	setDefaultConfig(&config)

	return &config, nil
}

// createDefaultConfig 
func createDefaultConfig(configPath string) error {
	config := getDefaultConfig()

	// 
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf(": %v", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("? %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf(": %v", err)
	}

	log.Printf("? %s", configPath)
	return nil
}

// getDefaultConfig 
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

// setDefaultConfig ?
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

	// ?
	config.Tracing.ServiceName = config.Service.Name
	config.Tracing.ServiceVersion = config.Service.Version
	config.Tracing.Environment = config.Service.Environment
}

// createMonitoringConfig 
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

// setupSignalHandling 
func setupSignalHandling(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigChan:
			log.Printf(" %v?..", sig)
			cancel()
		case <-ctx.Done():
			return
		}
	}()
}

// printVersion 汾
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

	// 
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf(": %v", err)
	}

	// 
	if *logLevel != "" {
		config.Service.LogLevel = *logLevel
	}

	log.Printf(" %s v%s (: %s)", config.Service.Name, config.Service.Version, config.Service.Environment)
	log.Printf(": %s", *configFile)
	log.Printf(": %s", config.Service.LogLevel)

	// ?
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 
	setupSignalHandling(ctx, cancel)

	// 
	monitoringConfig := createMonitoringConfig(config)
	monitoringSystem, err := integration.NewMonitoringSystem(monitoringConfig)
	if err != nil {
		log.Fatalf(": %v", err)
	}

	// 
	log.Println("...")
	if err := monitoringSystem.Start(); err != nil {
		log.Fatalf(": %v", err)
	}

	log.Printf(": http://localhost:%d", config.Dashboard.Port)
	log.Println("?Ctrl+C ")

	// 
	statsTicker := time.NewTicker(30 * time.Second)
	defer statsTicker.Stop()

	// ?
	healthTicker := time.NewTicker(time.Minute)
	defer healthTicker.Stop()

	// ?
	for {
		select {
		case <-ctx.Done():
			log.Println("?..")
			
			// 
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer shutdownCancel()

			done := make(chan error, 1)
			go func() {
				done <- monitoringSystem.Stop()
			}()

			select {
			case err := <-done:
				if err != nil {
					log.Printf("? %v", err)
				} else {
					log.Println("?)
				}
			case <-shutdownCtx.Done():
				log.Println("?)
			}

			return

		case <-statsTicker.C:
			// 
			stats := monitoringSystem.GetStats()
			log.Printf(" - : %d spans, : %d entries, 澯: %d alerts, ? %d workflows",
				stats.TracingStats.SpansCreated,
				stats.LoggingStats.ProcessedLogs,
				stats.AlertingStats.TotalAlerts,
				stats.AutomationStats.TotalWorkflows)

		case <-healthTicker.C:
			// ?
			if err := monitoringSystem.HealthCheck(); err != nil {
				log.Printf("? %v", err)
			} else {
				log.Println("")
			}
		}
	}
}

