package logging

import (
	"fmt"
	"time"
)

// LogEntry 日志条目
type LogEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Service   string                 `json:"service"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Tags      map[string]string      `json:"tags,omitempty"`
	Raw       []byte                 `json:"raw,omitempty"`
}

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelTrace LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别字符�?
func (l LogLevel) String() string {
	switch l {
	case LogLevelTrace:
		return "TRACE"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogCollector 日志收集器接�?
type LogCollector interface {
	// Start 启动收集�?
	Start() error
	
	// Stop 停止收集�?
	Stop() error
	
	// SetLogHandler 设置日志处理回调
	SetLogHandler(handler func(*LogEntry) error)
	
	// GetStats 获取统计信息
	GetStats() *CollectorStats
	
	// HealthCheck 健康检�?
	HealthCheck() error
}

// LogProcessor 日志处理器接�?
type LogProcessor interface {
	// Start 启动处理�?
	Start() error
	
	// Stop 停止处理�?
	Stop() error
	
	// Process 处理日志条目
	Process(entry *LogEntry) (*LogEntry, error)
	
	// GetStats 获取统计信息
	GetStats() *ProcessorStats
	
	// HealthCheck 健康检�?
	HealthCheck() error
}

// LogOutput 日志输出接口
type LogOutput interface {
	// Start 启动输出
	Start() error
	
	// Stop 停止输出
	Stop() error
	
	// Output 输出日志条目
	Output(entries []*LogEntry) error
	
	// GetStats 获取统计信息
	GetStats() *OutputStats
	
	// HealthCheck 健康检�?
	HealthCheck() error
}

// CollectorStats 收集器统计信�?
type CollectorStats struct {
	CollectedLogs  int64         `json:"collected_logs"`
	ErrorLogs      int64         `json:"error_logs"`
	LastCollected  time.Time     `json:"last_collected"`
	CollectionTime time.Duration `json:"collection_time"`
	SourceInfo     string        `json:"source_info"`
	IsActive       bool          `json:"is_active"`
}

// ProcessorStats 处理器统计信�?
type ProcessorStats struct {
	ProcessedLogs  int64         `json:"processed_logs"`
	FilteredLogs   int64         `json:"filtered_logs"`
	ErrorLogs      int64         `json:"error_logs"`
	LastProcessed  time.Time     `json:"last_processed"`
	ProcessingTime time.Duration `json:"processing_time"`
	IsActive       bool          `json:"is_active"`
}

// OutputStats 输出统计信息
type OutputStats struct {
	OutputLogs     int64         `json:"output_logs"`
	FailedLogs     int64         `json:"failed_logs"`
	LastOutput     time.Time     `json:"last_output"`
	OutputTime     time.Duration `json:"output_time"`
	Destination    string        `json:"destination"`
	IsActive       bool          `json:"is_active"`
}

// CollectorConfig 收集器配�?
type CollectorConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Name     string                 `json:"name" yaml:"name"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// ProcessorConfig 处理器配�?
type ProcessorConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Name     string                 `json:"name" yaml:"name"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// OutputConfig 输出配置
type OutputConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Name     string                 `json:"name" yaml:"name"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// LogFilter 日志过滤�?
type LogFilter struct {
	Level    *LogLevel             `json:"level,omitempty"`
	Source   []string              `json:"source,omitempty"`
	Service  []string              `json:"service,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	Tags     map[string]string     `json:"tags,omitempty"`
	TimeFrom *time.Time            `json:"time_from,omitempty"`
	TimeTo   *time.Time            `json:"time_to,omitempty"`
}

// LogQuery 日志查询
type LogQuery struct {
	Filter    *LogFilter `json:"filter,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
	SortBy    string     `json:"sort_by,omitempty"`
	SortOrder string     `json:"sort_order,omitempty"` // asc, desc
}

// LogQueryResult 日志查询结果
type LogQueryResult struct {
	Entries    []*LogEntry `json:"entries"`
	Total      int64       `json:"total"`
	HasMore    bool        `json:"has_more"`
	NextOffset int         `json:"next_offset,omitempty"`
}

// LogAggregation 日志聚合
type LogAggregation struct {
	GroupBy   []string               `json:"group_by"`
	Metrics   []string               `json:"metrics"` // count, avg, sum, min, max
	Interval  time.Duration          `json:"interval,omitempty"`
	Filter    *LogFilter             `json:"filter,omitempty"`
	TimeRange *TimeRange             `json:"time_range,omitempty"`
}

// TimeRange 时间范围
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// LogAggregationResult 日志聚合结果
type LogAggregationResult struct {
	Groups []LogGroup `json:"groups"`
	Total  int64      `json:"total"`
}

// LogGroup 日志分组
type LogGroup struct {
	Key     map[string]interface{} `json:"key"`
	Count   int64                  `json:"count"`
	Metrics map[string]float64     `json:"metrics"`
}

// CreateCollector 创建收集器工厂函�?
func CreateCollector(config CollectorConfig) (LogCollector, error) {
	switch config.Type {
	case "file":
		return NewFileCollector(config)
	case "syslog":
		return NewSyslogCollector(config)
	case "journald":
		return NewJournaldCollector(config)
	case "docker":
		return NewDockerCollector(config)
	case "kubernetes":
		return NewKubernetesCollector(config)
	case "tcp":
		return NewTCPCollector(config)
	case "udp":
		return NewUDPCollector(config)
	case "http":
		return NewHTTPCollector(config)
	default:
		return nil, fmt.Errorf("unknown collector type: %s", config.Type)
	}
}

// CreateProcessor 创建处理器工厂函�?
func CreateProcessor(config ProcessorConfig) (LogProcessor, error) {
	switch config.Type {
	case "filter":
		return NewFilterProcessor(config)
	case "parser":
		return NewParserProcessor(config)
	case "enricher":
		return NewEnricherProcessor(config)
	case "transformer":
		return NewTransformerProcessor(config)
	case "aggregator":
		return NewAggregatorProcessor(config)
	case "sampler":
		return NewSamplerProcessor(config)
	default:
		return nil, fmt.Errorf("unknown processor type: %s", config.Type)
	}
}

// CreateOutput 创建输出工厂函数
func CreateOutput(config OutputConfig) (LogOutput, error) {
	switch config.Type {
	case "file":
		return NewFileOutput(config)
	case "elasticsearch":
		return NewElasticsearchOutput(config)
	case "kafka":
		return NewKafkaOutput(config)
	case "redis":
		return NewRedisOutput(config)
	case "influxdb":
		return NewInfluxDBOutput(config)
	case "prometheus":
		return NewPrometheusOutput(config)
	case "console":
		return NewConsoleOutput(config)
	case "webhook":
		return NewWebhookOutput(config)
	default:
		return nil, fmt.Errorf("unknown output type: %s", config.Type)
	}
}
