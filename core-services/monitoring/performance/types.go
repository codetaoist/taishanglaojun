package performance

import (
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
)

// MetricCollector 指标收集器接�?
type MetricCollector interface {
	Start() error
	Stop() error
	Collect() ([]interfaces.Metric, error)
	GetStats() *CollectorStats
	HealthCheck() error
}

// Analyzer 分析器接�?
type Analyzer interface {
	Start() error
	Stop() error
	Analyze(timeRange TimeRange) (*AnalysisResult, error)
	GetStats() *AnalysisStats
}

// CollectorConfig 收集器配�?
type CollectorConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Interval time.Duration          `json:"interval" yaml:"interval"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// AnalyzerSettings 分析器设�?
type AnalyzerSettings struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// CollectorStats 收集器统计信�?
type CollectorStats struct {
	CollectedMetrics int64     `json:"collected_metrics"`
	Errors           int64     `json:"errors"`
	LastCollection   time.Time `json:"last_collection"`
	CollectionTime   time.Duration `json:"collection_time"`
}

// AnalysisStats 分析统计信息
type AnalysisStats struct {
	AnalyzedMetrics   int64     `json:"analyzed_metrics"`
	DetectedAnomalies int64     `json:"detected_anomalies"`
	GeneratedAlerts   int64     `json:"generated_alerts"`
	LastAnalysis      time.Time `json:"last_analysis"`
	AnalysisTime      time.Duration `json:"analysis_time"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp time.Time      `json:"timestamp"`
	CPU       *CPUMetrics    `json:"cpu,omitempty"`
	Memory    *MemoryMetrics `json:"memory,omitempty"`
	Disk      *DiskMetrics   `json:"disk,omitempty"`
	Network   *NetworkMetrics `json:"network,omitempty"`
	Process   *ProcessMetrics `json:"process,omitempty"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	Usage      float64            `json:"usage"`       // CPU使用�?(%)
	LoadAvg1   float64            `json:"load_avg_1"`  // 1分钟负载平均�?
	LoadAvg5   float64            `json:"load_avg_5"`  // 5分钟负载平均�?
	LoadAvg15  float64            `json:"load_avg_15"` // 15分钟负载平均�?
	Cores      int                `json:"cores"`       // CPU核心�?
	PerCore    map[string]float64 `json:"per_core"`    // 每个核心的使用率
	Frequency  float64            `json:"frequency"`   // CPU频率 (MHz)
	Temperature float64           `json:"temperature"` // CPU温度 (°C)
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Total       uint64  `json:"total"`        // 总内�?(bytes)
	Used        uint64  `json:"used"`         // 已使用内�?(bytes)
	Free        uint64  `json:"free"`         // 空闲内存 (bytes)
	Available   uint64  `json:"available"`    // 可用内存 (bytes)
	Usage       float64 `json:"usage"`        // 内存使用�?(%)
	Cached      uint64  `json:"cached"`       // 缓存内存 (bytes)
	Buffers     uint64  `json:"buffers"`      // 缓冲区内�?(bytes)
	SwapTotal   uint64  `json:"swap_total"`   // 总交换空�?(bytes)
	SwapUsed    uint64  `json:"swap_used"`    // 已使用交换空�?(bytes)
	SwapFree    uint64  `json:"swap_free"`    // 空闲交换空间 (bytes)
}

// DiskMetrics 磁盘指标
type DiskMetrics struct {
	Devices map[string]*DiskDeviceMetrics `json:"devices"` // 按设备分组的磁盘指标
}

// DiskDeviceMetrics 磁盘设备指标
type DiskDeviceMetrics struct {
	Total       uint64  `json:"total"`        // 总空�?(bytes)
	Used        uint64  `json:"used"`         // 已使用空�?(bytes)
	Free        uint64  `json:"free"`         // 空闲空间 (bytes)
	Usage       float64 `json:"usage"`        // 使用�?(%)
	ReadBytes   uint64  `json:"read_bytes"`   // 读取字节�?
	WriteBytes  uint64  `json:"write_bytes"`  // 写入字节�?
	ReadOps     uint64  `json:"read_ops"`     // 读取操作�?
	WriteOps    uint64  `json:"write_ops"`    // 写入操作�?
	ReadTime    uint64  `json:"read_time"`    // 读取时间 (ms)
	WriteTime   uint64  `json:"write_time"`   // 写入时间 (ms)
	IOTime      uint64  `json:"io_time"`      // IO时间 (ms)
	IOPS        float64 `json:"iops"`         // 每秒IO操作�?
	Throughput  float64 `json:"throughput"`   // 吞吐�?(bytes/s)
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	Interfaces map[string]*NetworkInterfaceMetrics `json:"interfaces"` // 按接口分组的网络指标
}

// NetworkInterfaceMetrics 网络接口指标
type NetworkInterfaceMetrics struct {
	BytesReceived    uint64  `json:"bytes_received"`    // 接收字节�?
	BytesSent        uint64  `json:"bytes_sent"`        // 发送字节数
	PacketsReceived  uint64  `json:"packets_received"`  // 接收包数
	PacketsSent      uint64  `json:"packets_sent"`      // 发送包�?
	ErrorsReceived   uint64  `json:"errors_received"`   // 接收错误�?
	ErrorsSent       uint64  `json:"errors_sent"`       // 发送错误数
	DroppedReceived  uint64  `json:"dropped_received"`  // 接收丢包�?
	DroppedSent      uint64  `json:"dropped_sent"`      // 发送丢包数
	Speed            uint64  `json:"speed"`             // 接口速度 (bits/s)
	Duplex           string  `json:"duplex"`            // 双工模式
	MTU              int     `json:"mtu"`               // 最大传输单�?
	RxRate           float64 `json:"rx_rate"`           // 接收速率 (bytes/s)
	TxRate           float64 `json:"tx_rate"`           // 发送速率 (bytes/s)
}

// ProcessMetrics 进程指标
type ProcessMetrics struct {
	Count       int                        `json:"count"`       // 进程总数
	Running     int                        `json:"running"`     // 运行中进程数
	Sleeping    int                        `json:"sleeping"`    // 睡眠进程�?
	Stopped     int                        `json:"stopped"`     // 停止进程�?
	Zombie      int                        `json:"zombie"`      // 僵尸进程�?
	TopCPU      []*ProcessInfo             `json:"top_cpu"`     // CPU使用率最高的进程
	TopMemory   []*ProcessInfo             `json:"top_memory"`  // 内存使用最多的进程
	Details     map[string]*ProcessInfo    `json:"details"`     // 特定进程详情
}

// ProcessInfo 进程信息
type ProcessInfo struct {
	PID         int     `json:"pid"`          // 进程ID
	Name        string  `json:"name"`         // 进程名称
	Command     string  `json:"command"`      // 命令�?
	CPUUsage    float64 `json:"cpu_usage"`    // CPU使用�?(%)
	MemoryUsage uint64  `json:"memory_usage"` // 内存使用�?(bytes)
	MemoryPercent float64 `json:"memory_percent"` // 内存使用�?(%)
	Status      string  `json:"status"`       // 进程状�?
	StartTime   time.Time `json:"start_time"` // 启动时间
	User        string  `json:"user"`         // 用户
	Threads     int     `json:"threads"`      // 线程�?
	FDs         int     `json:"fds"`          // 文件描述符数
}

// TimeRange 时间范围
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	TimeRange    TimeRange  `json:"time_range"`
	Timestamp    time.Time  `json:"timestamp"`
	OverallScore float64    `json:"overall_score"` // 总体评分 (0-100)
	Anomalies    []*Anomaly `json:"anomalies"`     // 检测到的异�?
	Trends       []*Trend   `json:"trends"`        // 趋势分析
	Alerts       []*Alert   `json:"alerts"`        // 生成的告�?
}

// Anomaly 异常
type Anomaly struct {
	MetricName  string          `json:"metric_name"`
	Timestamp   time.Time       `json:"timestamp"`
	Value       float64         `json:"value"`
	Expected    float64         `json:"expected"`
	Deviation   float64         `json:"deviation"`
	Severity    AnomalySeverity `json:"severity"`
	Description string          `json:"description"`
}

// AnomalySeverity 异常严重程度
type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "low"
	AnomalySeverityMedium   AnomalySeverity = "medium"
	AnomalySeverityHigh     AnomalySeverity = "high"
	AnomalySeverityCritical AnomalySeverity = "critical"
)

// Trend 趋势
type Trend struct {
	MetricName  string         `json:"metric_name"`
	TimeRange   TimeRange      `json:"time_range"`
	Direction   TrendDirection `json:"direction"`
	Slope       float64        `json:"slope"`
	Confidence  float64        `json:"confidence"` // 置信�?(0-1)
	Description string         `json:"description"`
}

// TrendDirection 趋势方向
type TrendDirection string

const (
	TrendDirectionIncreasing TrendDirection = "increasing"
	TrendDirectionDecreasing TrendDirection = "decreasing"
	TrendDirectionStable     TrendDirection = "stable"
)

// Alert 告警
type Alert struct {
	ID          string      `json:"id"`
	MetricName  string      `json:"metric_name"`
	Timestamp   time.Time   `json:"timestamp"`
	Severity    AlertSeverity `json:"severity"`
	Message     string      `json:"message"`
	Value       float64     `json:"value"`
	Threshold   float64     `json:"threshold"`
	Status      AlertStatus `json:"status"`
	Resolved    bool        `json:"resolved"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
}

// AlertSeverity 告警严重程度
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus 告警状�?
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusSuppressed AlertStatus = "suppressed"
	AlertStatusResolved  AlertStatus = "resolved"
)

// AnalysisResult 分析结果
type AnalysisResult struct {
	Anomalies []*Anomaly `json:"anomalies"`
	Trends    []*Trend   `json:"trends"`
	Alerts    []*Alert   `json:"alerts"`
}

// MetricThreshold 指标阈�?
type MetricThreshold struct {
	MetricName string    `json:"metric_name"`
	Warning    float64   `json:"warning"`
	Critical   float64   `json:"critical"`
	Operator   string    `json:"operator"` // >, <, >=, <=, ==, !=
}

// PerformanceProfile 性能配置文件
type PerformanceProfile struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Thresholds  []*MetricThreshold `json:"thresholds"`
	Rules       []*AnalysisRule    `json:"rules"`
	Enabled     bool               `json:"enabled"`
}

// AnalysisRule 分析规则
type AnalysisRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	MetricName  string            `json:"metric_name"`
	Condition   string            `json:"condition"`
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Severity    AlertSeverity     `json:"severity"`
	Actions     []string          `json:"actions"`
	Labels      map[string]string `json:"labels"`
	Enabled     bool              `json:"enabled"`
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`       // CPU使用�?(%)
	Memory    float64   `json:"memory"`    // 内存使用�?(%)
	Disk      float64   `json:"disk"`      // 磁盘使用�?(%)
	Network   float64   `json:"network"`   // 网络使用�?(%)
	Load      float64   `json:"load"`      // 系统负载
}

// PerformanceBaseline 性能基线
type PerformanceBaseline struct {
	MetricName string    `json:"metric_name"`
	Mean       float64   `json:"mean"`
	StdDev     float64   `json:"std_dev"`
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	Percentiles map[string]float64 `json:"percentiles"` // P50, P90, P95, P99
	UpdatedAt  time.Time `json:"updated_at"`
}

// CapacityPrediction 容量预测
type CapacityPrediction struct {
	MetricName     string    `json:"metric_name"`
	CurrentValue   float64   `json:"current_value"`
	PredictedValue float64   `json:"predicted_value"`
	PredictionTime time.Time `json:"prediction_time"`
	Confidence     float64   `json:"confidence"`
	Method         string    `json:"method"` // linear, exponential, seasonal
	TimeToLimit    *time.Duration `json:"time_to_limit,omitempty"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`        // cpu, memory, disk, network
	Priority    string            `json:"priority"`    // low, medium, high, critical
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Impact      string            `json:"impact"`      // 预期影响
	Effort      string            `json:"effort"`      // 实施难度
	Actions     []string          `json:"actions"`     // 具体行动�?
	Metrics     []string          `json:"metrics"`     // 相关指标
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
}

// HealthScore 健康评分
type HealthScore struct {
	Overall    float64            `json:"overall"`    // 总体评分 (0-100)
	Components map[string]float64 `json:"components"` // 各组件评�?
	Timestamp  time.Time          `json:"timestamp"`
	Details    map[string]string  `json:"details"`    // 评分详情
}
