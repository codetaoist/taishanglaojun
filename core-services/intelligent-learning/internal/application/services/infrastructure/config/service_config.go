package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ServiceConfigManager 服务配置管理器
type ServiceConfigManager struct {
	configPath string
	config     *GlobalServiceConfig
}

// GlobalServiceConfig 全局服务配置
type GlobalServiceConfig struct {
	// 服务配置
	CrossModalService                    *CrossModalServiceConfig                    `json:"cross_modal_service"`
	RelationInferenceEngine              *RelationInferenceEngineConfig              `json:"relation_inference_engine"`
	AdaptiveLearningEngine               *AdaptiveLearningEngineConfig               `json:"adaptive_learning_engine"`
	RealtimeLearningAnalyticsService     *RealtimeLearningAnalyticsServiceConfig     `json:"realtime_learning_analytics_service"`
	AutomatedKnowledgeGraphService       *AutomatedKnowledgeGraphServiceConfig       `json:"automated_knowledge_graph_service"`
	LearningAnalyticsReportingService    *LearningAnalyticsReportingServiceConfig    `json:"learning_analytics_reporting_service"`
	IntelligentContentRecommendationService *IntelligentContentRecommendationServiceConfig `json:"intelligent_content_recommendation_service"`
	
	// 集成配置
	IntegrationConfig *IntegrationConfig `json:"integration_config"`
	
	// 全局设置
	GlobalSettings *GlobalSettings `json:"global_settings"`
	
	// 环境配置
	Environment string `json:"environment"`
	Version     string `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CrossModalServiceConfig 跨模态服务配置
type CrossModalServiceConfig struct {
	Enabled           bool                   `json:"enabled"`
	MaxConcurrency    int                    `json:"max_concurrency"`
	Timeout           time.Duration          `json:"timeout"`
	ModelConfigs      map[string]interface{} `json:"model_configs"`
	ProcessingLimits  *ProcessingLimits      `json:"processing_limits"`
	QualityThresholds map[string]float64     `json:"quality_thresholds"`
}

// RelationInferenceEngineConfig 关系推理引擎配置
type RelationInferenceEngineConfig struct {
	Enabled              bool                   `json:"enabled"`
	InferenceAlgorithms  []string               `json:"inference_algorithms"`
	ConfidenceThreshold  float64                `json:"confidence_threshold"`
	MaxInferenceDepth    int                    `json:"max_inference_depth"`
	CacheSize            int                    `json:"cache_size"`
	PerformanceSettings  *PerformanceSettings   `json:"performance_settings"`
	AlgorithmWeights     map[string]float64     `json:"algorithm_weights"`
}

// AdaptiveLearningEngineConfig 自适应学习引擎配置
type AdaptiveLearningEngineConfig struct {
	Enabled                bool                   `json:"enabled"`
	AdaptationStrategies   []string               `json:"adaptation_strategies"`
	LearningPathAlgorithms []string               `json:"learning_path_algorithms"`
	PersonalizationLevel   string                 `json:"personalization_level"`
	UpdateFrequency        time.Duration          `json:"update_frequency"`
	ModelParameters        map[string]interface{} `json:"model_parameters"`
	QualityMetrics         map[string]float64     `json:"quality_metrics"`
}

// RealtimeLearningAnalyticsServiceConfig 实时学习分析服务配置
type RealtimeLearningAnalyticsServiceConfig struct {
	Enabled              bool                   `json:"enabled"`
	DataStreamSources    []string               `json:"data_stream_sources"`
	ProcessingInterval   time.Duration          `json:"processing_interval"`
	AnalyticsAlgorithms  []string               `json:"analytics_algorithms"`
	AlertThresholds      map[string]float64     `json:"alert_thresholds"`
	DataRetentionPeriod  time.Duration          `json:"data_retention_period"`
	StreamingConfig      *StreamingConfig       `json:"streaming_config"`
}

// AutomatedKnowledgeGraphServiceConfig 自动化知识图谱服务配置
type AutomatedKnowledgeGraphServiceConfig struct {
	Enabled                bool                   `json:"enabled"`
	GraphDatabases         []string               `json:"graph_databases"`
	EntityExtractionModels []string               `json:"entity_extraction_models"`
	RelationExtractionModels []string             `json:"relation_extraction_models"`
	UpdateStrategies       []string               `json:"update_strategies"`
	ValidationRules        map[string]interface{} `json:"validation_rules"`
	GraphOptimization      *GraphOptimizationConfig `json:"graph_optimization"`
}

// LearningAnalyticsReportingServiceConfig 学习分析报告服务配置
type LearningAnalyticsReportingServiceConfig struct {
	Enabled              bool                   `json:"enabled"`
	ReportTypes          []string               `json:"report_types"`
	GenerationSchedule   map[string]string      `json:"generation_schedule"`
	ExportFormats        []string               `json:"export_formats"`
	VisualizationEngines []string               `json:"visualization_engines"`
	QualityStandards     map[string]float64     `json:"quality_standards"`
	CacheSettings        *CacheSettings         `json:"cache_settings"`
}

// IntelligentContentRecommendationServiceConfig 智能内容推荐服务配置
type IntelligentContentRecommendationServiceConfig struct {
	Enabled                  bool                   `json:"enabled"`
	RecommendationAlgorithms []string               `json:"recommendation_algorithms"`
	PersonalizationStrategies []string              `json:"personalization_strategies"`
	ContentAnalysisModels    []string               `json:"content_analysis_models"`
	RecommendationLimits     *RecommendationLimits  `json:"recommendation_limits"`
	QualityFilters           map[string]interface{} `json:"quality_filters"`
	LearnerProfilingConfig   *LearnerProfilingConfig `json:"learner_profiling_config"`
}

// ProcessingLimits 处理限制
type ProcessingLimits struct {
	MaxFileSize       int64         `json:"max_file_size"`
	MaxProcessingTime time.Duration `json:"max_processing_time"`
	MaxMemoryUsage    int64         `json:"max_memory_usage"`
	ConcurrentJobs    int           `json:"concurrent_jobs"`
}

// StreamingConfig 流处理配置
type StreamingConfig struct {
	BufferSize       int           `json:"buffer_size"`
	FlushInterval    time.Duration `json:"flush_interval"`
	MaxBatchSize     int           `json:"max_batch_size"`
	CompressionLevel int           `json:"compression_level"`
}

// GraphOptimizationConfig 图优化配置
type GraphOptimizationConfig struct {
	EnableIndexing       bool          `json:"enable_indexing"`
	IndexUpdateInterval  time.Duration `json:"index_update_interval"`
	CompressionEnabled   bool          `json:"compression_enabled"`
	PartitioningStrategy string        `json:"partitioning_strategy"`
}

// CacheSettings 缓存设置
type CacheSettings struct {
	Enabled         bool          `json:"enabled"`
	MaxSize         int           `json:"max_size"`
	TTL             time.Duration `json:"ttl"`
	EvictionPolicy  string        `json:"eviction_policy"`
	// 新增字段以支持intelligent_content_recommendation_service.go的使用
	EnableCaching   bool          `json:"enable_caching"`
	CacheTTL        time.Duration `json:"cache_ttl"`
	MaxCacheSize    int           `json:"max_cache_size"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// RecommendationLimits 推荐限制
type RecommendationLimits struct {
	MaxRecommendations     int           `json:"max_recommendations"`
	MinConfidenceScore     float64       `json:"min_confidence_score"`
	RecommendationTimeout  time.Duration `json:"recommendation_timeout"`
	MaxContentAnalysisTime time.Duration `json:"max_content_analysis_time"`
}

// LearnerProfilingConfig 学习者画像配置
type LearnerProfilingConfig struct {
	ProfilingStrategies    []string               `json:"profiling_strategies"`
	BehaviorAnalysisModels []string               `json:"behavior_analysis_models"`
	PreferenceWeights      map[string]float64     `json:"preference_weights"`
	UpdateFrequency        time.Duration          `json:"update_frequency"`
	PrivacySettings        *PrivacySettings       `json:"privacy_settings"`
}

// PrivacySettings 隐私设置
type PrivacySettings struct {
	DataAnonymization bool          `json:"data_anonymization"`
	DataRetention     time.Duration `json:"data_retention"`
	ConsentRequired   bool          `json:"consent_required"`
	EncryptionLevel   string        `json:"encryption_level"`
}

// GlobalSettings 全局设置
type GlobalSettings struct {
	LogLevel            string                 `json:"log_level"`
	MetricsEnabled      bool                   `json:"metrics_enabled"`
	TracingEnabled      bool                   `json:"tracing_enabled"`
	HealthCheckInterval time.Duration          `json:"health_check_interval"`
	DatabaseConfig      *DatabaseConfig        `json:"database_config"`
	SecurityConfig      *SecurityConfiguration `json:"security_config"`
	MonitoringConfig    *MonitoringConfiguration `json:"monitoring_config"`
}

// IntegrationConfig 集成配置
type IntegrationConfig struct {
	ServiceTimeouts    map[string]time.Duration `json:"service_timeouts"`
	ServiceRetries     map[string]int           `json:"service_retries"`
	IntegrationSettings *LearningIntegrationSettings `json:"integration_settings"`
	PerformanceConfig  *PerformanceConfiguration    `json:"performance_config"`
}

// LearningIntegrationSettings 学习集成设置
type LearningIntegrationSettings struct {
	EnableServiceOrchestration bool     `json:"enable_service_orchestration"`
	EnableDataSynchronization  bool     `json:"enable_data_synchronization"`
	EnableCrossServiceCaching  bool     `json:"enable_cross_service_caching"`
	EnableEventDrivenUpdates   bool     `json:"enable_event_driven_updates"`
	DataFlowPriority          []string `json:"data_flow_priority"`
}

// PerformanceConfiguration 性能配置
type PerformanceConfiguration struct {
	MaxConcurrentRequests int                          `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration                `json:"request_timeout"`
	CacheExpiration       time.Duration                `json:"cache_expiration"`
	BatchProcessingSize   int                          `json:"batch_processing_size"`
	LoadBalancing         *LearningLoadBalancingConfig `json:"load_balancing"`
	CircuitBreaker        *CircuitBreakerConfig        `json:"circuit_breaker"`
}

// LearningLoadBalancingConfig 学习负载均衡配置
type LearningLoadBalancingConfig struct {
	Strategy            string        `json:"strategy"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MaxRetries          int           `json:"max_retries"`
}

// CircuitBreakerConfig 断路器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenRequests int           `json:"half_open_requests"`
}

// SecurityConfiguration 安全配置
type SecurityConfiguration struct {
	EnableAuthentication bool          `json:"enable_authentication"`
	EnableAuthorization  bool          `json:"enable_authorization"`
	TokenExpiration      time.Duration `json:"token_expiration"`
	EncryptionEnabled    bool          `json:"encryption_enabled"`
	AuditLogging         bool          `json:"audit_logging"`
}

// MonitoringConfiguration 监控配置
type MonitoringConfiguration struct {
	EnableMetrics     bool          `json:"enable_metrics"`
	EnableTracing     bool          `json:"enable_tracing"`
	MetricsInterval   time.Duration `json:"metrics_interval"`
	AlertingEnabled   bool          `json:"alerting_enabled"`
	LogLevel          string        `json:"log_level"`
}

// PerformanceSettings 性能设置
type PerformanceSettings struct {
	MaxConcurrency    int           `json:"max_concurrency"`
	Timeout           time.Duration `json:"timeout"`
	MemoryLimit       int64         `json:"memory_limit"`
	CPULimit          float64       `json:"cpu_limit"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type             string        `json:"type"`
	ConnectionString string        `json:"connection_string"`
	MaxConnections   int           `json:"max_connections"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	QueryTimeout     time.Duration `json:"query_timeout"`
	RetryAttempts    int           `json:"retry_attempts"`
}

// NewServiceConfigManager 创建服务配置管理器
func NewServiceConfigManager(configPath string) *ServiceConfigManager {
	return &ServiceConfigManager{
		configPath: configPath,
		config:     getDefaultConfig(),
	}
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *GlobalServiceConfig {
	return &GlobalServiceConfig{
		CrossModalService: &CrossModalServiceConfig{
			Enabled:        true,
			MaxConcurrency: 10,
			Timeout:        30 * time.Second,
			ModelConfigs: map[string]interface{}{
				"text_model":  "bert-base-uncased",
				"image_model": "resnet50",
				"audio_model": "wav2vec2",
			},
			ProcessingLimits: &ProcessingLimits{
				MaxFileSize:       100 * 1024 * 1024, // 100MB
				MaxProcessingTime: 5 * time.Minute,
				MaxMemoryUsage:    1024 * 1024 * 1024, // 1GB
				ConcurrentJobs:    5,
			},
			QualityThresholds: map[string]float64{
				"confidence": 0.8,
				"accuracy":   0.85,
			},
		},
		
		RelationInferenceEngine: &RelationInferenceEngineConfig{
			Enabled:             true,
			InferenceAlgorithms: []string{"rule_based", "ml_based", "hybrid"},
			ConfidenceThreshold: 0.7,
			MaxInferenceDepth:   5,
			CacheSize:           1000,
			PerformanceSettings: &PerformanceSettings{
				MaxConcurrency: 20,
				Timeout:        15 * time.Second,
				MemoryLimit:    1024 * 1024 * 1024, // 1GB
				CPULimit:       0.8,
			},
			AlgorithmWeights: map[string]float64{
				"rule_based": 0.3,
				"ml_based":   0.5,
				"hybrid":     0.2,
			},
		},
		
		AdaptiveLearningEngine: &AdaptiveLearningEngineConfig{
			Enabled:                true,
			AdaptationStrategies:   []string{"content_based", "collaborative", "knowledge_based"},
			LearningPathAlgorithms: []string{"shortest_path", "optimal_learning", "personalized"},
			PersonalizationLevel:   "high",
			UpdateFrequency:        10 * time.Minute,
			ModelParameters: map[string]interface{}{
				"learning_rate":     0.01,
				"adaptation_factor": 0.8,
				"forgetting_curve":  0.9,
			},
			QualityMetrics: map[string]float64{
				"effectiveness": 0.8,
				"efficiency":    0.75,
				"engagement":    0.85,
			},
		},
		
		RealtimeLearningAnalyticsService: &RealtimeLearningAnalyticsServiceConfig{
			Enabled:             true,
			DataStreamSources:   []string{"user_interactions", "learning_events", "system_metrics"},
			ProcessingInterval:  5 * time.Second,
			AnalyticsAlgorithms: []string{"statistical", "ml_based", "pattern_recognition"},
			AlertThresholds: map[string]float64{
				"performance_drop": 0.2,
				"engagement_low":   0.3,
				"error_rate_high":  0.1,
			},
			DataRetentionPeriod: 30 * 24 * time.Hour, // 30 days
			StreamingConfig: &StreamingConfig{
				BufferSize:       1000,
				FlushInterval:    1 * time.Second,
				MaxBatchSize:     100,
				CompressionLevel: 6,
			},
		},
		
		AutomatedKnowledgeGraphService: &AutomatedKnowledgeGraphServiceConfig{
			Enabled:                  true,
			GraphDatabases:           []string{"neo4j", "arangodb"},
			EntityExtractionModels:   []string{"spacy", "bert_ner"},
			RelationExtractionModels: []string{"openie", "bert_relation"},
			UpdateStrategies:         []string{"incremental", "batch", "real_time"},
			ValidationRules: map[string]interface{}{
				"min_confidence":     0.6,
				"max_entity_length":  100,
				"allowed_relations":  []string{"is_a", "part_of", "related_to"},
			},
			GraphOptimization: &GraphOptimizationConfig{
				EnableIndexing:       true,
				IndexUpdateInterval:  1 * time.Hour,
				CompressionEnabled:   true,
				PartitioningStrategy: "hash_based",
			},
		},
		
		LearningAnalyticsReportingService: &LearningAnalyticsReportingServiceConfig{
			Enabled:     true,
			ReportTypes: []string{"progress", "performance", "engagement", "competency"},
			GenerationSchedule: map[string]string{
				"daily":   "0 8 * * *",
				"weekly":  "0 8 * * 1",
				"monthly": "0 8 1 * *",
			},
			ExportFormats:        []string{"pdf", "html", "json", "csv"},
			VisualizationEngines: []string{"d3js", "plotly", "chartjs"},
			QualityStandards: map[string]float64{
				"data_completeness": 0.9,
				"accuracy":          0.95,
				"timeliness":        0.8,
			},
			CacheSettings: &CacheSettings{
				Enabled:        true,
				MaxSize:        500,
				TTL:            2 * time.Hour,
				EvictionPolicy: "lru",
			},
		},
		
		IntelligentContentRecommendationService: &IntelligentContentRecommendationServiceConfig{
			Enabled:                   true,
			RecommendationAlgorithms:  []string{"collaborative_filtering", "content_based", "hybrid"},
			PersonalizationStrategies: []string{"learning_style", "performance_based", "interest_based"},
			ContentAnalysisModels:     []string{"bert", "word2vec", "doc2vec"},
			RecommendationLimits: &RecommendationLimits{
				MaxRecommendations:     20,
				MinConfidenceScore:     0.6,
				RecommendationTimeout:  10 * time.Second,
				MaxContentAnalysisTime: 30 * time.Second,
			},
			QualityFilters: map[string]interface{}{
				"min_rating":        4.0,
				"content_freshness": 30, // days
				"difficulty_match":  0.8,
			},
			LearnerProfilingConfig: &LearnerProfilingConfig{
				ProfilingStrategies:    []string{"behavior_based", "preference_based", "performance_based"},
				BehaviorAnalysisModels: []string{"clustering", "classification", "sequence_analysis"},
				PreferenceWeights: map[string]float64{
					"learning_style": 0.3,
					"content_type":   0.25,
					"difficulty":     0.2,
					"topic":          0.25,
				},
				UpdateFrequency: 1 * time.Hour,
				PrivacySettings: &PrivacySettings{
					DataAnonymization: true,
					DataRetention:     90 * 24 * time.Hour, // 90 days
					ConsentRequired:   true,
					EncryptionLevel:   "AES256",
				},
			},
		},
		
		IntegrationConfig: &IntegrationConfig{
			ServiceTimeouts: map[string]time.Duration{
				"cross_modal":            30 * time.Second,
				"relation_inference":     15 * time.Second,
				"adaptive_learning":      20 * time.Second,
				"realtime_analytics":     10 * time.Second,
				"knowledge_graph":        25 * time.Second,
				"analytics_reporting":    60 * time.Second,
				"content_recommendation": 15 * time.Second,
			},
			ServiceRetries: map[string]int{
				"cross_modal":            3,
				"relation_inference":     2,
				"adaptive_learning":      3,
				"realtime_analytics":     2,
				"knowledge_graph":        3,
				"analytics_reporting":    2,
				"content_recommendation": 3,
			},
			IntegrationSettings: &LearningIntegrationSettings{
				EnableServiceOrchestration: true,
				EnableDataSynchronization:  true,
				EnableCrossServiceCaching:  true,
				EnableEventDrivenUpdates:   true,
				DataFlowPriority: []string{
					"realtime_analytics",
					"adaptive_learning",
					"content_recommendation",
					"analytics_reporting",
				},
			},
			PerformanceConfig: &PerformanceConfiguration{
				MaxConcurrentRequests: 100,
				RequestTimeout:        60 * time.Second,
				CacheExpiration:       30 * time.Minute,
				BatchProcessingSize:   50,
				LoadBalancing: &LearningLoadBalancingConfig{
					Strategy:            "round_robin",
					HealthCheckInterval: 30 * time.Second,
					MaxRetries:          3,
				},
				CircuitBreaker: &CircuitBreakerConfig{
					FailureThreshold: 5,
					RecoveryTimeout:  30 * time.Second,
					HalfOpenRequests: 3,
				},
			},
		},
		
		GlobalSettings: &GlobalSettings{
			LogLevel:            "info",
			MetricsEnabled:      true,
			TracingEnabled:      true,
			HealthCheckInterval: 30 * time.Second,
			DatabaseConfig: &DatabaseConfig{
				Type:              "postgresql",
				ConnectionString:  "postgres://user:password@localhost:5432/intelligent_learning",
				MaxConnections:    20,
				ConnectionTimeout: 10 * time.Second,
				QueryTimeout:      30 * time.Second,
				RetryAttempts:     3,
			},
		},
		
		Environment: "development",
		Version:     "1.0.0",
		UpdatedAt:   time.Now(),
	}
}

// LoadConfig 加载配置
func (scm *ServiceConfigManager) LoadConfig() error {
	if _, err := os.Stat(scm.configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		return scm.SaveConfig()
	}
	
	data, err := os.ReadFile(scm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config GlobalServiceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	scm.config = &config
	return nil
}

// SaveConfig 保存配置
func (scm *ServiceConfigManager) SaveConfig() error {
	// 确保目录存在
	dir := filepath.Dir(scm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	scm.config.UpdatedAt = time.Now()
	
	data, err := json.MarshalIndent(scm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(scm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// GetConfig 获取配置
func (scm *ServiceConfigManager) GetConfig() *GlobalServiceConfig {
	return scm.config
}

// UpdateConfig 更新配置
func (scm *ServiceConfigManager) UpdateConfig(config *GlobalServiceConfig) error {
	scm.config = config
	return scm.SaveConfig()
}

// ValidateConfig 验证配置
func (scm *ServiceConfigManager) ValidateConfig() error {
	if scm.config == nil {
		return fmt.Errorf("config is nil")
	}
	
	// 验证必要的配置项
	if scm.config.GlobalSettings == nil {
		return fmt.Errorf("global settings is required")
	}
	
	if scm.config.IntegrationConfig == nil {
		return fmt.Errorf("integration config is required")
	}
	
	// 验证数据库配置
	if scm.config.GlobalSettings.DatabaseConfig == nil {
		return fmt.Errorf("database config is required")
	}
	
	if scm.config.GlobalSettings.DatabaseConfig.ConnectionString == "" {
		return fmt.Errorf("database connection string is required")
	}
	
	// 验证集成配置
	if scm.config.IntegrationConfig == nil {
		return fmt.Errorf("integration config is required")
	}
	
	return nil
}

// GetServiceConfig 获取特定服务的配置
func (scm *ServiceConfigManager) GetServiceConfig(serviceName string) (interface{}, error) {
	switch serviceName {
	case "cross_modal":
		return scm.config.CrossModalService, nil
	case "relation_inference":
		return scm.config.RelationInferenceEngine, nil
	case "adaptive_learning":
		return scm.config.AdaptiveLearningEngine, nil
	case "realtime_analytics":
		return scm.config.RealtimeLearningAnalyticsService, nil
	case "knowledge_graph":
		return scm.config.AutomatedKnowledgeGraphService, nil
	case "analytics_reporting":
		return scm.config.LearningAnalyticsReportingService, nil
	case "content_recommendation":
		return scm.config.IntelligentContentRecommendationService, nil
	default:
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}
}

// IsServiceEnabled 检查服务是否启用
func (scm *ServiceConfigManager) IsServiceEnabled(serviceName string) bool {
	config, err := scm.GetServiceConfig(serviceName)
	if err != nil {
		return false
	}
	
	switch c := config.(type) {
	case *CrossModalServiceConfig:
		return c.Enabled
	case *RelationInferenceEngineConfig:
		return c.Enabled
	case *AdaptiveLearningEngineConfig:
		return c.Enabled
	case *RealtimeLearningAnalyticsServiceConfig:
		return c.Enabled
	case *AutomatedKnowledgeGraphServiceConfig:
		return c.Enabled
	case *LearningAnalyticsReportingServiceConfig:
		return c.Enabled
	case *IntelligentContentRecommendationServiceConfig:
		return c.Enabled
	default:
		return false
	}
}

// GetEnvironment 获取环境
func (scm *ServiceConfigManager) GetEnvironment() string {
	return scm.config.Environment
}

// SetEnvironment 设置环境
func (scm *ServiceConfigManager) SetEnvironment(env string) {
	scm.config.Environment = env
	scm.config.UpdatedAt = time.Now()
}

// GetVersion 获取版本
func (scm *ServiceConfigManager) GetVersion() string {
	return scm.config.Version
}

// SetVersion 设置版本
func (scm *ServiceConfigManager) SetVersion(version string) {
	scm.config.Version = version
	scm.config.UpdatedAt = time.Now()
}