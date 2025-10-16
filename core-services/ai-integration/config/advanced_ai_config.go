package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AdvancedAIConfig AI
type AdvancedAIConfig struct {
	//
	EnableAGI           bool `json:"enable_agi" yaml:"enable_agi"`
	EnableMetaLearning  bool `json:"enable_meta_learning" yaml:"enable_meta_learning"`
	EnableSelfEvolution bool `json:"enable_self_evolution" yaml:"enable_self_evolution"`
	EnableHybridMode    bool `json:"enable_hybrid_mode" yaml:"enable_hybrid_mode"`

	// 启用的能力列表
	EnabledCapabilities []string `json:"enabled_capabilities" yaml:"enabled_capabilities"`

	//
	MaxConcurrentRequests int           `json:"max_concurrent_requests" yaml:"max_concurrent_requests"`
	DefaultTimeout        time.Duration `json:"default_timeout" yaml:"default_timeout"`
	MaxRequestSize        int64         `json:"max_request_size" yaml:"max_request_size"`

	// AGI 配置
	AGI AGIConfig `json:"agi" yaml:"agi"`

	// 元学习配置
	MetaLearning MetaLearningConfig `json:"meta_learning" yaml:"meta_learning"`

	// 自进化配置
	SelfEvolution SelfEvolutionConfig `json:"self_evolution" yaml:"self_evolution"`

	// 监控配置
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`

	// 安全配置
	Security SecurityConfig `json:"security" yaml:"security"`

	// 日志配置
	Logging LoggingConfig `json:"logging" yaml:"logging"`
}

// AGIConfig AGI
type AGIConfig struct {
	EnableReasoning     bool `json:"enable_reasoning" yaml:"enable_reasoning"`
	EnablePlanning      bool `json:"enable_planning" yaml:"enable_planning"`
	EnableLearning      bool `json:"enable_learning" yaml:"enable_learning"`
	EnableCreativity    bool `json:"enable_creativity" yaml:"enable_creativity"`
	EnableMultimodal    bool `json:"enable_multimodal" yaml:"enable_multimodal"`
	EnableMetacognition bool `json:"enable_metacognition" yaml:"enable_metacognition"`

	// 推理配置
	ReasoningDepth   int           `json:"reasoning_depth" yaml:"reasoning_depth"`
	ReasoningTimeout time.Duration `json:"reasoning_timeout" yaml:"reasoning_timeout"`

	// 规划配置
	PlanningHorizon    int    `json:"planning_horizon" yaml:"planning_horizon"`
	PlanningComplexity string `json:"planning_complexity" yaml:"planning_complexity"`

	// 学习配置
	LearningRate        float64 `json:"learning_rate" yaml:"learning_rate"`
	AdaptationThreshold float64 `json:"adaptation_threshold" yaml:"adaptation_threshold"`

	// 创造力配置
	CreativityLevel  string  `json:"creativity_level" yaml:"creativity_level"`
	NoveltyThreshold float64 `json:"novelty_threshold" yaml:"novelty_threshold"`

	// 多模态配置
	SupportedModalities []string           `json:"supported_modalities" yaml:"supported_modalities"`
	ModalityWeights     map[string]float64 `json:"modality_weights" yaml:"modality_weights"`

	// 元认知配置
	SelfAwarenessLevel  string  `json:"self_awareness_level" yaml:"self_awareness_level"`
	ConfidenceThreshold float64 `json:"confidence_threshold" yaml:"confidence_threshold"`

	// 自定义参数
	CustomParameters map[string]interface{} `json:"custom_parameters" yaml:"custom_parameters"`
}

// MetaLearningConfig 元学习
type MetaLearningConfig struct {
	EnableGradientBased    bool `json:"enable_gradient_based" yaml:"enable_gradient_based"`
	EnableModelAgnostic    bool `json:"enable_model_agnostic" yaml:"enable_model_agnostic"`
	EnableMemoryAugmented  bool `json:"enable_memory_augmented" yaml:"enable_memory_augmented"`
	EnableFewShot          bool `json:"enable_few_shot" yaml:"enable_few_shot"`
	EnableTransferLearning bool `json:"enable_transfer_learning" yaml:"enable_transfer_learning"`
	EnableOnlineAdaptation bool `json:"enable_online_adaptation" yaml:"enable_online_adaptation"`

	// 策略配置
	DefaultStrategy   string `json:"default_strategy" yaml:"default_strategy"`
	StrategySelection string `json:"strategy_selection" yaml:"strategy_selection"`

	// 学习配置
	MaxLearningSteps     int           `json:"max_learning_steps" yaml:"max_learning_steps"`
	LearningTimeout      time.Duration `json:"learning_timeout" yaml:"learning_timeout"`
	ConvergenceThreshold float64       `json:"convergence_threshold" yaml:"convergence_threshold"`

	// 记忆配置
	MemoryCapacity      int     `json:"memory_capacity" yaml:"memory_capacity"`
	MemoryRetentionRate float64 `json:"memory_retention_rate" yaml:"memory_retention_rate"`

	// 迁移学习配置
	TransferThreshold      float64 `json:"transfer_threshold" yaml:"transfer_threshold"`
	DomainSimilarityWeight float64 `json:"domain_similarity_weight" yaml:"domain_similarity_weight"`

	// 自适应配置
	AdaptationRate float64 `json:"adaptation_rate" yaml:"adaptation_rate"`
	ForgettingRate float64 `json:"forgetting_rate" yaml:"forgetting_rate"`

	// 自定义参数
	CustomParameters map[string]interface{} `json:"custom_parameters" yaml:"custom_parameters"`
}

// SelfEvolutionConfig 自进化
type SelfEvolutionConfig struct {
	EnableGenetic           bool `json:"enable_genetic" yaml:"enable_genetic"`
	EnableNeuroEvolution    bool `json:"enable_neuro_evolution" yaml:"enable_neuro_evolution"`
	EnableGradientFree      bool `json:"enable_gradient_free" yaml:"enable_gradient_free"`
	EnableHybrid            bool `json:"enable_hybrid" yaml:"enable_hybrid"`
	EnableReinforcement     bool `json:"enable_reinforcement" yaml:"enable_reinforcement"`
	EnableSwarmIntelligence bool `json:"enable_swarm_intelligence" yaml:"enable_swarm_intelligence"`

	// 策略配置
	DefaultStrategy   string `json:"default_strategy" yaml:"default_strategy"`
	StrategySelection string `json:"strategy_selection" yaml:"strategy_selection"`

	// 遗传算法配置
	PopulationSize int     `json:"population_size" yaml:"population_size"`
	MaxGenerations int     `json:"max_generations" yaml:"max_generations"`
	EliteRatio     float64 `json:"elite_ratio" yaml:"elite_ratio"`

	// NeuroEvolution 配置
	MutationRate      float64 `json:"mutation_rate" yaml:"mutation_rate"`
	CrossoverRate     float64 `json:"crossover_rate" yaml:"crossover_rate"`
	SelectionPressure float64 `json:"selection_pressure" yaml:"selection_pressure"`

	// 神经网络配置
	NetworkArchitecture    string  `json:"network_architecture" yaml:"network_architecture"`
	NetworkComplexity      string  `json:"network_complexity" yaml:"network_complexity"`
	StructuralMutationRate float64 `json:"structural_mutation_rate" yaml:"structural_mutation_rate"`

	// 多目标优化配置
	OptimizationTargets   []string           `json:"optimization_targets" yaml:"optimization_targets"`
	MultiObjectiveWeights map[string]float64 `json:"multi_objective_weights" yaml:"multi_objective_weights"`

	// 进化配置
	EvolutionTimeout     time.Duration `json:"evolution_timeout" yaml:"evolution_timeout"`
	EvaluationTimeout    time.Duration `json:"evaluation_timeout" yaml:"evaluation_timeout"`
	ConvergenceThreshold float64       `json:"convergence_threshold" yaml:"convergence_threshold"`

	// 自定义参数
	CustomParameters map[string]interface{} `json:"custom_parameters" yaml:"custom_parameters"`
}

// MonitoringConfig 监控
type MonitoringConfig struct {
	EnablePerformanceMonitoring bool `json:"enable_performance_monitoring" yaml:"enable_performance_monitoring"`
	EnableHealthChecks          bool `json:"enable_health_checks" yaml:"enable_health_checks"`
	EnableMetricsCollection     bool `json:"enable_metrics_collection" yaml:"enable_metrics_collection"`
	EnableAlerting              bool `json:"enable_alerting" yaml:"enable_alerting"`

	// 监控配置
	MetricsRetentionPeriod    time.Duration `json:"metrics_retention_period" yaml:"metrics_retention_period"`
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval" yaml:"metrics_collection_interval"`

	// 健康检查配置
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	HealthThreshold     float64       `json:"health_threshold" yaml:"health_threshold"`

	// 告警配置
	AlertThresholds     map[string]float64 `json:"alert_thresholds" yaml:"alert_thresholds"`
	AlertCooldownPeriod time.Duration      `json:"alert_cooldown_period" yaml:"alert_cooldown_period"`
}

// SecurityConfig 安全
type SecurityConfig struct {
	EnableAuthentication  bool `json:"enable_authentication" yaml:"enable_authentication"`
	EnableAuthorization   bool `json:"enable_authorization" yaml:"enable_authorization"`
	EnableRateLimiting    bool `json:"enable_rate_limiting" yaml:"enable_rate_limiting"`
	EnableInputValidation bool `json:"enable_input_validation" yaml:"enable_input_validation"`
	EnableOutputFiltering bool `json:"enable_output_filtering" yaml:"enable_output_filtering"`

	// 认证配置
	AuthenticationMethod string        `json:"authentication_method" yaml:"authentication_method"`
	TokenExpirationTime  time.Duration `json:"token_expiration_time" yaml:"token_expiration_time"`

	// 授权配置
	AuthorizationModel  string   `json:"authorization_model" yaml:"authorization_model"`
	RequiredPermissions []string `json:"required_permissions" yaml:"required_permissions"`

	// 速率限制配置
	RateLimitRequests int           `json:"rate_limit_requests" yaml:"rate_limit_requests"`
	RateLimitWindow   time.Duration `json:"rate_limit_window" yaml:"rate_limit_window"`

	// 输入验证配置
	MaxInputSize      int64    `json:"max_input_size" yaml:"max_input_size"`
	AllowedInputTypes []string `json:"allowed_input_types" yaml:"allowed_input_types"`

	// 输出过滤配置
	FilterSensitiveData   bool     `json:"filter_sensitive_data" yaml:"filter_sensitive_data"`
	SensitiveDataPatterns []string `json:"sensitive_data_patterns" yaml:"sensitive_data_patterns"`
}

// LoggingConfig 日志
type LoggingConfig struct {
	LogLevel             string `json:"log_level" yaml:"log_level"`
	LogFormat            string `json:"log_format" yaml:"log_format"`
	LogOutput            string `json:"log_output" yaml:"log_output"`
	EnableRequestLogging bool   `json:"enable_request_logging" yaml:"enable_request_logging"`
	EnableErrorLogging   bool   `json:"enable_error_logging" yaml:"enable_error_logging"`
	EnableDebugLogging   bool   `json:"enable_debug_logging" yaml:"enable_debug_logging"`

	// 日志旋转配置
	LogRotationSize    int64         `json:"log_rotation_size" yaml:"log_rotation_size"`
	LogRetentionPeriod time.Duration `json:"log_retention_period" yaml:"log_retention_period"`
	MaxLogFiles        int           `json:"max_log_files" yaml:"max_log_files"`
}

// DefaultAdvancedAIConfig 默认配置
func DefaultAdvancedAIConfig() *AdvancedAIConfig {
	return &AdvancedAIConfig{
		//
		EnableAGI:           true,
		EnableMetaLearning:  true,
		EnableSelfEvolution: true,
		EnableHybridMode:    true,

		//
		MaxConcurrentRequests: 50,
		DefaultTimeout:        30 * time.Second,
		MaxRequestSize:        10 * 1024 * 1024, // 10MB

		// AGI
		AGI: AGIConfig{
			EnableReasoning:     true,
			EnablePlanning:      true,
			EnableLearning:      true,
			EnableCreativity:    true,
			EnableMultimodal:    true,
			EnableMetacognition: true,

			ReasoningDepth:      5,
			ReasoningTimeout:    10 * time.Second,
			PlanningHorizon:     10,
			PlanningComplexity:  "medium",
			LearningRate:        0.01,
			AdaptationThreshold: 0.8,
			CreativityLevel:     "medium",
			NoveltyThreshold:    0.7,
			SupportedModalities: []string{"text", "image", "audio", "video"},
			ModalityWeights: map[string]float64{
				"text":  1.0,
				"image": 0.8,
				"audio": 0.6,
				"video": 0.7,
			},
			SelfAwarenessLevel:  "medium",
			ConfidenceThreshold: 0.75,
			CustomParameters:    make(map[string]interface{}),
		},

		// 元学习配置
		MetaLearning: MetaLearningConfig{
			EnableGradientBased:    true,
			EnableModelAgnostic:    true,
			EnableMemoryAugmented:  true,
			EnableFewShot:          true,
			EnableTransferLearning: true,
			EnableOnlineAdaptation: true,

			DefaultStrategy:        "model_agnostic",
			StrategySelection:      "adaptive",
			MaxLearningSteps:       1000,
			LearningTimeout:        5 * time.Minute,
			ConvergenceThreshold:   0.01,
			MemoryCapacity:         10000,
			MemoryRetentionRate:    0.95,
			TransferThreshold:      0.8,
			DomainSimilarityWeight: 0.7,
			AdaptationRate:         0.1,
			ForgettingRate:         0.01,
			CustomParameters:       make(map[string]interface{}),
		},

		// 自进化配置
		SelfEvolution: SelfEvolutionConfig{
			EnableGenetic:           true,
			EnableNeuroEvolution:    true,
			EnableGradientFree:      true,
			EnableHybrid:            true,
			EnableReinforcement:     true,
			EnableSwarmIntelligence: true,

			DefaultStrategy:        "genetic",
			StrategySelection:      "adaptive",
			PopulationSize:         50,
			MaxGenerations:         100,
			EliteRatio:             0.1,
			MutationRate:           0.1,
			CrossoverRate:          0.8,
			SelectionPressure:      2.0,
			NetworkComplexity:      "medium",
			StructuralMutationRate: 0.05,
			OptimizationTargets:    []string{"accuracy", "efficiency", "robustness"},
			MultiObjectiveWeights: map[string]float64{
				"accuracy":   0.4,
				"efficiency": 0.3,
				"robustness": 0.3,
			},
			EvolutionTimeout:     30 * time.Minute,
			EvaluationTimeout:    5 * time.Minute,
			ConvergenceThreshold: 0.001,
			CustomParameters:     make(map[string]interface{}),
		},

		// 监控配置
		Monitoring: MonitoringConfig{
			EnablePerformanceMonitoring: true,
			EnableHealthChecks:          true,
			EnableMetricsCollection:     true,
			EnableAlerting:              true,

			MetricsRetentionPeriod:    24 * time.Hour,
			MetricsCollectionInterval: 1 * time.Minute,
			HealthCheckInterval:       30 * time.Second,
			HealthThreshold:           0.8,
			AlertThresholds: map[string]float64{
				"error_rate":    0.05,
				"response_time": 5.0,
				"cpu_usage":     0.8,
				"memory_usage":  0.8,
			},
			AlertCooldownPeriod: 5 * time.Minute,
		},

		// 安全配置
		Security: SecurityConfig{
			EnableAuthentication:  true,
			EnableAuthorization:   true,
			EnableRateLimiting:    true,
			EnableInputValidation: true,
			EnableOutputFiltering: true,

			AuthenticationMethod: "jwt",
			TokenExpirationTime:  24 * time.Hour,
			AuthorizationModel:   "rbac",
			RequiredPermissions:  []string{"ai:read", "ai:write"},
			RateLimitRequests:    100,
			RateLimitWindow:      1 * time.Minute,
			MaxInputSize:         1024 * 1024, // 1MB
			AllowedInputTypes:    []string{"text", "json", "image", "audio"},
			FilterSensitiveData:  true,
			SensitiveDataPatterns: []string{
				`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,          //
				`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, //
				`\b\d{3}-\d{2}-\d{4}\b`,                               // SSN
			},
		},

		// 日志配置
		Logging: LoggingConfig{
			LogLevel:             "info",
			LogFormat:            "json",
			LogOutput:            "file",
			EnableRequestLogging: true,
			EnableErrorLogging:   true,
			EnableDebugLogging:   false,

			LogRotationSize:    100 * 1024 * 1024,  // 100MB
			LogRetentionPeriod: 7 * 24 * time.Hour, // 7
			MaxLogFiles:        10,
		},
	}
}

// LoadFromEnv 从环境变量加载配置
func (c *AdvancedAIConfig) LoadFromEnv() {
	//
	if val := os.Getenv("ADVANCED_AI_ENABLE_AGI"); val != "" {
		c.EnableAGI = val == "true"
	}
	if val := os.Getenv("ADVANCED_AI_ENABLE_META_LEARNING"); val != "" {
		c.EnableMetaLearning = val == "true"
	}
	if val := os.Getenv("ADVANCED_AI_ENABLE_SELF_EVOLUTION"); val != "" {
		c.EnableSelfEvolution = val == "true"
	}
	if val := os.Getenv("ADVANCED_AI_ENABLE_HYBRID_MODE"); val != "" {
		c.EnableHybridMode = val == "true"
	}

	// 元学习配置
	if val := os.Getenv("ADVANCED_AI_MAX_CONCURRENT_REQUESTS"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			c.MaxConcurrentRequests = intVal
		}
	}
	if val := os.Getenv("ADVANCED_AI_DEFAULT_TIMEOUT"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			c.DefaultTimeout = duration
		}
	}
	if val := os.Getenv("ADVANCED_AI_MAX_REQUEST_SIZE"); val != "" {
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			c.MaxRequestSize = intVal
		}
	}

	// 日志配置
	if val := os.Getenv("ADVANCED_AI_LOG_LEVEL"); val != "" {
		c.Logging.LogLevel = val
	}
	if val := os.Getenv("ADVANCED_AI_LOG_FORMAT"); val != "" {
		c.Logging.LogFormat = val
	}
	if val := os.Getenv("ADVANCED_AI_LOG_OUTPUT"); val != "" {
		c.Logging.LogOutput = val
	}
}

// Validate 验证配置
func (c *AdvancedAIConfig) Validate() error {
	//
	if c.MaxConcurrentRequests <= 0 {
		return fmt.Errorf("max_concurrent_requests must be positive")
	}
	if c.DefaultTimeout <= 0 {
		return fmt.Errorf("default_timeout must be positive")
	}
	if c.MaxRequestSize <= 0 {
		return fmt.Errorf("max_request_size must be positive")
	}

	// AGI 配置
	if c.AGI.ReasoningDepth <= 0 {
		return fmt.Errorf("agi.reasoning_depth must be positive")
	}
	if c.AGI.PlanningHorizon <= 0 {
		return fmt.Errorf("agi.planning_horizon must be positive")
	}
	if c.AGI.LearningRate <= 0 || c.AGI.LearningRate >= 1 {
		return fmt.Errorf("agi.learning_rate must be between 0 and 1")
	}

	// 元学习配置
	if c.MetaLearning.MaxLearningSteps <= 0 {
		return fmt.Errorf("meta_learning.max_learning_steps must be positive")
	}
	if c.MetaLearning.MemoryCapacity <= 0 {
		return fmt.Errorf("meta_learning.memory_capacity must be positive")
	}

	// 自进化配置
	if c.SelfEvolution.PopulationSize <= 0 {
		return fmt.Errorf("self_evolution.population_size must be positive")
	}
	if c.SelfEvolution.MaxGenerations <= 0 {
		return fmt.Errorf("self_evolution.max_generations must be positive")
	}
	if c.SelfEvolution.MutationRate < 0 || c.SelfEvolution.MutationRate > 1 {
		return fmt.Errorf("self_evolution.mutation_rate must be between 0 and 1")
	}

	return nil
}

// GetAGIConfig AGI 配置
func (c *AdvancedAIConfig) GetAGIConfig() *AGIConfig {
	return &c.AGI
}

// GetMetaLearningConfig 元学习配置
func (c *AdvancedAIConfig) GetMetaLearningConfig() *MetaLearningConfig {
	return &c.MetaLearning
}

// GetSelfEvolutionConfig 自进化配置
func (c *AdvancedAIConfig) GetSelfEvolutionConfig() *SelfEvolutionConfig {
	return &c.SelfEvolution
}

// GetMonitoringConfig 监控配置
func (c *AdvancedAIConfig) GetMonitoringConfig() *MonitoringConfig {
	return &c.Monitoring
}

// GetSecurityConfig 安全配置
func (c *AdvancedAIConfig) GetSecurityConfig() *SecurityConfig {
	return &c.Security
}

// GetLoggingConfig 日志配置
func (c *AdvancedAIConfig) GetLoggingConfig() *LoggingConfig {
	return &c.Logging
}

// UpdateConfig 更新配置
func (c *AdvancedAIConfig) UpdateConfig(updates map[string]interface{}) error {
	//
	//
	for key, value := range updates {
		switch key {
		case "enable_agi":
			if boolVal, ok := value.(bool); ok {
				c.EnableAGI = boolVal
			}
		case "enable_meta_learning":
			if boolVal, ok := value.(bool); ok {
				c.EnableMetaLearning = boolVal
			}
		case "enable_self_evolution":
			if boolVal, ok := value.(bool); ok {
				c.EnableSelfEvolution = boolVal
			}
		case "max_concurrent_requests":
			if intVal, ok := value.(int); ok {
				c.MaxConcurrentRequests = intVal
			}
		case "log_level":
			if strVal, ok := value.(string); ok {
				c.Logging.LogLevel = strVal
			}
		}
	}

	return c.Validate()
}
