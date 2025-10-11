package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ModelType 模型类型枚举
type ModelType string

const (
	ModelTypeText        ModelType = "text"
	ModelTypeImage       ModelType = "image"
	ModelTypeAudio       ModelType = "audio"
	ModelTypeVideo       ModelType = "video"
	ModelTypeMultiModal  ModelType = "multimodal"
	ModelTypeEmbedding   ModelType = "embedding"
	ModelTypeClassifier  ModelType = "classifier"
	ModelTypeGenerator   ModelType = "generator"
)

// ModelProvider 模型提供�?
type ModelProvider string

const (
	ProviderOpenAI     ModelProvider = "openai"
	ProviderAnthropic  ModelProvider = "anthropic"
	ProviderGoogle     ModelProvider = "google"
	ProviderBaidu      ModelProvider = "baidu"
	ProviderAlibaba    ModelProvider = "alibaba"
	ProviderTencent    ModelProvider = "tencent"
	ProviderCustom     ModelProvider = "custom"
	ProviderLocal      ModelProvider = "local"
)

// ModelStatus 模型状�?
type ModelStatus string

const (
	StatusActive     ModelStatus = "active"
	StatusInactive   ModelStatus = "inactive"
	StatusTraining   ModelStatus = "training"
	StatusDeploying  ModelStatus = "deploying"
	StatusError      ModelStatus = "error"
	StatusMaintenance ModelStatus = "maintenance"
)

// AIModel AI模型接口
type AIModel interface {
	// 基础信息
	GetID() string
	GetName() string
	GetType() ModelType
	GetProvider() ModelProvider
	GetVersion() string
	GetStatus() ModelStatus

	// 模型操作
	Initialize(ctx context.Context, config ModelConfig) error
	Process(ctx context.Context, input ModelInput) (*ModelOutput, error)
	Validate(ctx context.Context, input ModelInput) error
	GetCapabilities() ModelCapabilities
	GetMetrics() ModelMetrics

	// 生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Reload(ctx context.Context, config ModelConfig) error
	HealthCheck(ctx context.Context) error
}

// ModelConfig 模型配置
type ModelConfig struct {
	ID           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name" yaml:"name"`
	Type         ModelType              `json:"type" yaml:"type"`
	Provider     ModelProvider          `json:"provider" yaml:"provider"`
	Version      string                 `json:"version" yaml:"version"`
	Endpoint     string                 `json:"endpoint" yaml:"endpoint"`
	APIKey       string                 `json:"api_key" yaml:"api_key"`
	MaxTokens    int                    `json:"max_tokens" yaml:"max_tokens"`
	Temperature  float64                `json:"temperature" yaml:"temperature"`
	TopP         float64                `json:"top_p" yaml:"top_p"`
	Timeout      time.Duration          `json:"timeout" yaml:"timeout"`
	RetryCount   int                    `json:"retry_count" yaml:"retry_count"`
	RateLimit    RateLimitConfig        `json:"rate_limit" yaml:"rate_limit"`
	CustomParams map[string]interface{} `json:"custom_params" yaml:"custom_params"`
	CreatedAt    time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" yaml:"updated_at"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute" yaml:"requests_per_minute"`
	TokensPerMinute   int           `json:"tokens_per_minute" yaml:"tokens_per_minute"`
	ConcurrentLimit   int           `json:"concurrent_limit" yaml:"concurrent_limit"`
	BurstSize         int           `json:"burst_size" yaml:"burst_size"`
	WindowDuration    time.Duration `json:"window_duration" yaml:"window_duration"`
}

// ModelInput 模型输入
type ModelInput struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Content     interface{}            `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Parameters  map[string]interface{} `json:"parameters"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	RequestID   string                 `json:"request_id"`
}

// ModelOutput 模型输出
type ModelOutput struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Content      interface{}            `json:"content"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata"`
	Metrics      ProcessingMetrics      `json:"metrics"`
	Timestamp    time.Time              `json:"timestamp"`
	ModelInfo    ModelInfo              `json:"model_info"`
	Error        *ModelError            `json:"error,omitempty"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Provider ModelProvider `json:"provider"`
	Type     ModelType     `json:"type"`
}

// ModelError 模型错误
type ModelError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// ProcessingMetrics 处理指标
type ProcessingMetrics struct {
	ProcessingTime   time.Duration `json:"processing_time"`
	TokensUsed       int           `json:"tokens_used"`
	InputTokens      int           `json:"input_tokens"`
	OutputTokens     int           `json:"output_tokens"`
	Cost             float64       `json:"cost"`
	QueueTime        time.Duration `json:"queue_time"`
	ModelLoadTime    time.Duration `json:"model_load_time"`
	InferenceTime    time.Duration `json:"inference_time"`
	PostProcessTime  time.Duration `json:"post_process_time"`
}

// ModelCapabilities 模型能力
type ModelCapabilities struct {
	SupportedInputTypes  []string               `json:"supported_input_types"`
	SupportedOutputTypes []string               `json:"supported_output_types"`
	MaxInputSize         int64                  `json:"max_input_size"`
	MaxOutputSize        int64                  `json:"max_output_size"`
	SupportsBatch        bool                   `json:"supports_batch"`
	SupportsStreaming    bool                   `json:"supports_streaming"`
	SupportsFinetuning   bool                   `json:"supports_finetuning"`
	Languages            []string               `json:"languages"`
	Features             map[string]interface{} `json:"features"`
}

// ModelMetrics 模型指标
type ModelMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	P95Latency        time.Duration `json:"p95_latency"`
	P99Latency        time.Duration `json:"p99_latency"`
	ThroughputRPS     float64       `json:"throughput_rps"`
	ErrorRate         float64       `json:"error_rate"`
	UpTime            time.Duration `json:"uptime"`
	LastHealthCheck   time.Time     `json:"last_health_check"`
	ResourceUsage     ResourceUsage `json:"resource_usage"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	GPUUsage    float64 `json:"gpu_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   int64   `json:"network_io"`
}

// ModelRegistry 模型注册表接�?
type ModelRegistry interface {
	Register(model AIModel) error
	Unregister(modelID string) error
	GetModel(modelID string) (AIModel, error)
	ListModels() []AIModel
	GetModelsByType(modelType ModelType) []AIModel
	GetModelsByProvider(provider ModelProvider) []AIModel
	UpdateModelConfig(modelID string, config ModelConfig) error
	GetModelConfig(modelID string) (*ModelConfig, error)
}

// ModelManager 模型管理器接�?
type ModelManager interface {
	// 模型生命周期管理
	LoadModel(ctx context.Context, config ModelConfig) error
	UnloadModel(ctx context.Context, modelID string) error
	ReloadModel(ctx context.Context, modelID string) error
	
	// 模型调用
	ProcessRequest(ctx context.Context, modelID string, input ModelInput) (*ModelOutput, error)
	BatchProcess(ctx context.Context, modelID string, inputs []ModelInput) ([]*ModelOutput, error)
	StreamProcess(ctx context.Context, modelID string, input ModelInput) (<-chan *ModelOutput, error)
	
	// 模型监控
	GetModelMetrics(modelID string) (*ModelMetrics, error)
	GetAllMetrics() (map[string]*ModelMetrics, error)
	HealthCheck(ctx context.Context, modelID string) error
	
	// 模型配置
	UpdateModelConfig(modelID string, config ModelConfig) error
	GetModelConfig(modelID string) (*ModelConfig, error)
	ListModels() ([]*ModelConfig, error)
}

// ModelFactory 模型工厂接口
type ModelFactory interface {
	CreateModel(config ModelConfig) (AIModel, error)
	GetSupportedTypes() []ModelType
	GetSupportedProviders() []ModelProvider
	ValidateConfig(config ModelConfig) error
}

// BaseModel 基础模型实现
type BaseModel struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         ModelType     `json:"type"`
	Provider     ModelProvider `json:"provider"`
	Version      string        `json:"version"`
	Status       ModelStatus   `json:"status"`
	Config       ModelConfig   `json:"config"`
	Capabilities ModelCapabilities `json:"capabilities"`
	Metrics      ModelMetrics  `json:"metrics"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// GetID 获取模型ID
func (m *BaseModel) GetID() string {
	return m.ID
}

// GetName 获取模型名称
func (m *BaseModel) GetName() string {
	return m.Name
}

// GetType 获取模型类型
func (m *BaseModel) GetType() ModelType {
	return m.Type
}

// GetProvider 获取模型提供�?
func (m *BaseModel) GetProvider() ModelProvider {
	return m.Provider
}

// GetVersion 获取模型版本
func (m *BaseModel) GetVersion() string {
	return m.Version
}

// GetStatus 获取模型状�?
func (m *BaseModel) GetStatus() ModelStatus {
	return m.Status
}

// GetCapabilities 获取模型能力
func (m *BaseModel) GetCapabilities() ModelCapabilities {
	return m.Capabilities
}

// GetMetrics 获取模型指标
func (m *BaseModel) GetMetrics() ModelMetrics {
	return m.Metrics
}

// NewModelInput 创建新的模型输入
func NewModelInput(content interface{}, inputType string) *ModelInput {
	return &ModelInput{
		ID:        uuid.New().String(),
		Type:      inputType,
		Content:   content,
		Metadata:  make(map[string]interface{}),
		Parameters: make(map[string]interface{}),
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
		RequestID: uuid.New().String(),
	}
}

// NewModelOutput 创建新的模型输出
func NewModelOutput(content interface{}, requestID string, modelInfo ModelInfo) *ModelOutput {
	return &ModelOutput{
		ID:        uuid.New().String(),
		RequestID: requestID,
		Content:   content,
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
		ModelInfo: modelInfo,
	}
}

// Validate 验证模型配置
func (c *ModelConfig) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("model ID is required")
	}
	if c.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("model type is required")
	}
	if c.Provider == "" {
		return fmt.Errorf("model provider is required")
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	if c.RetryCount < 0 {
		c.RetryCount = 3
	}
	return nil
}

// ToJSON 转换为JSON
func (c *ModelConfig) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// FromJSON 从JSON解析
func (c *ModelConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}
