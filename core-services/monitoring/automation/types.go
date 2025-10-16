package automation

import (
	"context"
	"sync"
	"time"
)

// Executor 执行器接?
type Executor interface {
	Start() error
	Stop() error
	ExecuteTask(task *Task, inputs map[string]interface{}) error
	GetStats() *ExecutorStats
	HealthCheck() error
}

// ExecutorConfig 执行器配?
type ExecutorConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// ExecutorStats 执行器统计信?
type ExecutorStats struct {
	ExecutedTasks    int64         `json:"executed_tasks"`
	SuccessfulTasks  int64         `json:"successful_tasks"`
	FailedTasks      int64         `json:"failed_tasks"`
	AverageExecTime  time.Duration `json:"average_exec_time"`
	LastExecution    time.Time     `json:"last_execution"`
}

// Workflow 工作?
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Definition  WorkflowDefinition     `json:"definition"`
	Status      WorkflowStatus         `json:"status"`
	Tasks       map[string]*Task       `json:"tasks"`
	Variables   map[string]interface{} `json:"variables"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Error       string                 `json:"error,omitempty"`
	mutex       sync.RWMutex
}

// WorkflowDefinition 工作流定?
type WorkflowDefinition struct {
	Name        string           `json:"name" yaml:"name"`
	Description string           `json:"description" yaml:"description"`
	Version     string           `json:"version" yaml:"version"`
	Tasks       []TaskDefinition `json:"tasks" yaml:"tasks"`
	Variables   map[string]interface{} `json:"variables" yaml:"variables"`
	Triggers    []TriggerDefinition `json:"triggers" yaml:"triggers"`
	Schedule    *ScheduleDefinition `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	Timeout     time.Duration    `json:"timeout" yaml:"timeout"`
	RetryPolicy *RetryPolicy     `json:"retry_policy,omitempty" yaml:"retry_policy,omitempty"`
	OnSuccess   []ActionDefinition `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnFailure   []ActionDefinition `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
	Tags        []string         `json:"tags" yaml:"tags"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
}

// WorkflowStatus 工作流状?
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusPaused    WorkflowStatus = "paused"
)

// Task 任务
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Definition  TaskDefinition         `json:"definition"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      TaskStatus             `json:"status"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Error       string                 `json:"error,omitempty"`
	Logs        []string               `json:"logs"`
	mutex       sync.RWMutex
}

// TaskDefinition 任务定义
type TaskDefinition struct {
	Name         string                 `json:"name" yaml:"name"`
	Type         string                 `json:"type" yaml:"type"`
	Description  string                 `json:"description" yaml:"description"`
	Command      string                 `json:"command" yaml:"command"`
	Script       string                 `json:"script" yaml:"script"`
	Image        string                 `json:"image" yaml:"image"`
	Environment  map[string]string      `json:"environment" yaml:"environment"`
	WorkingDir   string                 `json:"working_dir" yaml:"working_dir"`
	Timeout      time.Duration          `json:"timeout" yaml:"timeout"`
	RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty" yaml:"retry_policy,omitempty"`
	Dependencies []string               `json:"dependencies" yaml:"dependencies"`
	Conditions   []ConditionDefinition  `json:"conditions" yaml:"conditions"`
	Inputs       map[string]interface{} `json:"inputs" yaml:"inputs"`
	Outputs      map[string]string      `json:"outputs" yaml:"outputs"`
	Resources    *ResourceRequirements  `json:"resources,omitempty" yaml:"resources,omitempty"`
	OnSuccess    []ActionDefinition     `json:"on_success,omitempty" yaml:"on_success,omitempty"`
	OnFailure    []ActionDefinition     `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`
	Tags         []string               `json:"tags" yaml:"tags"`
	Labels       map[string]string      `json:"labels" yaml:"labels"`
}

// TaskStatus 任务状?
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusSkipped   TaskStatus = "skipped"
)

// Trigger 触发?
type Trigger struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Definition TriggerDefinition `json:"definition"`
	Status     TriggerStatus     `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	LastFired  time.Time         `json:"last_fired"`
	FireCount  int64             `json:"fire_count"`
}

// TriggerDefinition 触发器定?
type TriggerDefinition struct {
	Type        string                 `json:"type" yaml:"type"`
	Schedule    string                 `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	Event       string                 `json:"event,omitempty" yaml:"event,omitempty"`
	Webhook     *WebhookTrigger        `json:"webhook,omitempty" yaml:"webhook,omitempty"`
	Metric      *MetricTrigger         `json:"metric,omitempty" yaml:"metric,omitempty"`
	File        *FileTrigger           `json:"file,omitempty" yaml:"file,omitempty"`
	Conditions  []ConditionDefinition  `json:"conditions" yaml:"conditions"`
	Settings    map[string]interface{} `json:"settings" yaml:"settings"`
}

// TriggerStatus 触发器状?
type TriggerStatus string

const (
	TriggerStatusActive   TriggerStatus = "active"
	TriggerStatusInactive TriggerStatus = "inactive"
	TriggerStatusError    TriggerStatus = "error"
)

// WebhookTrigger Webhook触发?
type WebhookTrigger struct {
	URL     string            `json:"url" yaml:"url"`
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `json:"headers" yaml:"headers"`
	Secret  string            `json:"secret" yaml:"secret"`
}

// MetricTrigger 指标触发?
type MetricTrigger struct {
	MetricName string  `json:"metric_name" yaml:"metric_name"`
	Operator   string  `json:"operator" yaml:"operator"` // >, <, >=, <=, ==, !=
	Threshold  float64 `json:"threshold" yaml:"threshold"`
	Duration   time.Duration `json:"duration" yaml:"duration"`
}

// FileTrigger 文件触发?
type FileTrigger struct {
	Path    string `json:"path" yaml:"path"`
	Pattern string `json:"pattern" yaml:"pattern"`
	Event   string `json:"event" yaml:"event"` // create, modify, delete
}

// ScheduleDefinition 调度定义
type ScheduleDefinition struct {
	Cron     string        `json:"cron,omitempty" yaml:"cron,omitempty"`
	Interval time.Duration `json:"interval,omitempty" yaml:"interval,omitempty"`
	StartAt  *time.Time    `json:"start_at,omitempty" yaml:"start_at,omitempty"`
	EndAt    *time.Time    `json:"end_at,omitempty" yaml:"end_at,omitempty"`
	Timezone string        `json:"timezone,omitempty" yaml:"timezone,omitempty"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts" yaml:"max_attempts"`
	Delay       time.Duration `json:"delay" yaml:"delay"`
	BackoffType string        `json:"backoff_type" yaml:"backoff_type"` // fixed, exponential, linear
	MaxDelay    time.Duration `json:"max_delay" yaml:"max_delay"`
}

// ConditionDefinition 条件定义
type ConditionDefinition struct {
	Type     string      `json:"type" yaml:"type"` // expression, script, webhook
	Value    string      `json:"value" yaml:"value"`
	Operator string      `json:"operator,omitempty" yaml:"operator,omitempty"`
	Expected interface{} `json:"expected,omitempty" yaml:"expected,omitempty"`
}

// ActionDefinition 动作定义
type ActionDefinition struct {
	Type     string                 `json:"type" yaml:"type"`
	Target   string                 `json:"target" yaml:"target"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// ResourceRequirements 资源需?
type ResourceRequirements struct {
	CPU    string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
	Disk   string `json:"disk,omitempty" yaml:"disk,omitempty"`
}

// WorkflowFilter 工作流过滤器
type WorkflowFilter struct {
	Status        WorkflowStatus `json:"status,omitempty"`
	Name          string         `json:"name,omitempty"`
	CreatedAfter  time.Time      `json:"created_after,omitempty"`
	CreatedBefore time.Time      `json:"created_before,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
}

// TaskFilter 任务过滤?
type TaskFilter struct {
	Status     TaskStatus `json:"status,omitempty"`
	Type       string     `json:"type,omitempty"`
	WorkflowID string     `json:"workflow_id,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// Scheduler 调度?
type Scheduler struct {
	config    SchedulerConfig
	workflows chan *Workflow
	workers   []*Worker
	stats     *SchedulerStats
	mutex     sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// SchedulerConfig 调度器配?
type SchedulerConfig struct {
	Interval      time.Duration `json:"interval" yaml:"interval"`
	MaxConcurrent int           `json:"max_concurrent" yaml:"max_concurrent"`
	QueueSize     int           `json:"queue_size" yaml:"queue_size"`
}

// SchedulerStats 调度器统计信?
type SchedulerStats struct {
	QueuedWorkflows   int64 `json:"queued_workflows"`
	ProcessedWorkflows int64 `json:"processed_workflows"`
	ActiveWorkers     int   `json:"active_workers"`
}

// Worker 工作?
type Worker struct {
	id        int
	scheduler *Scheduler
	ctx       context.Context
	cancel    context.CancelFunc
}

// AutoScaler 自动扩缩容器
type AutoScaler struct {
	config AutoScalerConfig
	stats  *AutoScalerStats
	mutex  sync.RWMutex
}

// AutoScalerConfig 自动扩缩容配?
type AutoScalerConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	MinReplicas       int           `json:"min_replicas" yaml:"min_replicas"`
	MaxReplicas       int           `json:"max_replicas" yaml:"max_replicas"`
	TargetCPU         float64       `json:"target_cpu" yaml:"target_cpu"`
	TargetMemory      float64       `json:"target_memory" yaml:"target_memory"`
	ScaleUpThreshold  float64       `json:"scale_up_threshold" yaml:"scale_up_threshold"`
	ScaleDownThreshold float64      `json:"scale_down_threshold" yaml:"scale_down_threshold"`
	ScaleUpCooldown   time.Duration `json:"scale_up_cooldown" yaml:"scale_up_cooldown"`
	ScaleDownCooldown time.Duration `json:"scale_down_cooldown" yaml:"scale_down_cooldown"`
}

// AutoScalerStats 自动扩缩容统计信?
type AutoScalerStats struct {
	CurrentReplicas int       `json:"current_replicas"`
	DesiredReplicas int       `json:"desired_replicas"`
	LastScaleUp     time.Time `json:"last_scale_up"`
	LastScaleDown   time.Time `json:"last_scale_down"`
	ScaleUpCount    int64     `json:"scale_up_count"`
	ScaleDownCount  int64     `json:"scale_down_count"`
}

// LoadBalancer 负载均衡?
type LoadBalancer struct {
	config LoadBalancerConfig
	stats  *LoadBalancerStats
	mutex  sync.RWMutex
}

// LoadBalancerConfig 负载均衡配置
type LoadBalancerConfig struct {
	Algorithm string   `json:"algorithm" yaml:"algorithm"` // round_robin, least_connections, weighted
	Backends  []string `json:"backends" yaml:"backends"`
	HealthCheck *HealthCheckConfig `json:"health_check" yaml:"health_check"`
}

// LoadBalancerStats 负载均衡统计信息
type LoadBalancerStats struct {
	TotalRequests    int64            `json:"total_requests"`
	SuccessfulRequests int64          `json:"successful_requests"`
	FailedRequests   int64            `json:"failed_requests"`
	BackendStats     map[string]*BackendStats `json:"backend_stats"`
}

// BackendStats 后端统计信息
type BackendStats struct {
	Requests       int64         `json:"requests"`
	Errors         int64         `json:"errors"`
	ResponseTime   time.Duration `json:"response_time"`
	Healthy        bool          `json:"healthy"`
	LastHealthCheck time.Time    `json:"last_health_check"`
}

// HealthCheckConfig 健康检查配?
type HealthCheckConfig struct {
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Interval time.Duration `json:"interval" yaml:"interval"`
	Timeout  time.Duration `json:"timeout" yaml:"timeout"`
	Path     string        `json:"path" yaml:"path"`
	Method   string        `json:"method" yaml:"method"`
}

// CircuitBreaker 熔断?
type CircuitBreaker struct {
	config CircuitBreakerConfig
	stats  *CircuitBreakerStats
	state  CircuitBreakerState
	mutex  sync.RWMutex
}

// CircuitBreakerConfig 熔断器配?
type CircuitBreakerConfig struct {
	Enabled           bool          `json:"enabled" yaml:"enabled"`
	FailureThreshold  int           `json:"failure_threshold" yaml:"failure_threshold"`
	SuccessThreshold  int           `json:"success_threshold" yaml:"success_threshold"`
	Timeout           time.Duration `json:"timeout" yaml:"timeout"`
	ResetTimeout      time.Duration `json:"reset_timeout" yaml:"reset_timeout"`
}

// CircuitBreakerStats 熔断器统计信?
type CircuitBreakerStats struct {
	State           CircuitBreakerState `json:"state"`
	TotalRequests   int64               `json:"total_requests"`
	SuccessfulRequests int64            `json:"successful_requests"`
	FailedRequests  int64               `json:"failed_requests"`
	LastStateChange time.Time           `json:"last_state_change"`
}

// CircuitBreakerState 熔断器状?
type CircuitBreakerState string

const (
	CircuitBreakerStateClosed   CircuitBreakerState = "closed"
	CircuitBreakerStateOpen     CircuitBreakerState = "open"
	CircuitBreakerStateHalfOpen CircuitBreakerState = "half_open"
)

// RateLimiter 限流?
type RateLimiter struct {
	config RateLimiterConfig
	stats  *RateLimiterStats
	mutex  sync.RWMutex
}

// RateLimiterConfig 限流器配?
type RateLimiterConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	Rate       float64       `json:"rate" yaml:"rate"`         // requests per second
	Burst      int           `json:"burst" yaml:"burst"`       // burst size
	Window     time.Duration `json:"window" yaml:"window"`     // time window
	Algorithm  string        `json:"algorithm" yaml:"algorithm"` // token_bucket, sliding_window
}

// RateLimiterStats 限流器统计信?
type RateLimiterStats struct {
	TotalRequests    int64 `json:"total_requests"`
	AllowedRequests  int64 `json:"allowed_requests"`
	RejectedRequests int64 `json:"rejected_requests"`
	CurrentRate      float64 `json:"current_rate"`
}

// Deployment 部署
type Deployment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Image       string            `json:"image"`
	Replicas    int               `json:"replicas"`
	Status      DeploymentStatus  `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// DeploymentStatus 部署状?
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusCompleted DeploymentStatus = "completed"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusRolling   DeploymentStatus = "rolling"
)

// Service 服务
type Service struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Ports       []ServicePort     `json:"ports"`
	Selector    map[string]string `json:"selector"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ServicePort 服务端口
type ServicePort struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
}

// ConfigMap 配置映射
type ConfigMap struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Data        map[string]string `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Secret 密钥
type Secret struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Type        string            `json:"type"`
	Data        map[string][]byte `json:"data"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

