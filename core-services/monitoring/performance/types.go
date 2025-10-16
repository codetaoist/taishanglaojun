package performance

import (
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
)

// MetricCollector ?
type MetricCollector interface {
	Start() error
	Stop() error
	Collect() ([]interfaces.Metric, error)
	GetStats() *CollectorStats
	HealthCheck() error
}

// Analyzer ?
type Analyzer interface {
	Start() error
	Stop() error
	Analyze(timeRange TimeRange) (*AnalysisResult, error)
	GetStats() *AnalysisStats
}

// CollectorConfig ?
type CollectorConfig struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Interval time.Duration          `json:"interval" yaml:"interval"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// AnalyzerSettings ?
type AnalyzerSettings struct {
	Type     string                 `json:"type" yaml:"type"`
	Enabled  bool                   `json:"enabled" yaml:"enabled"`
	Settings map[string]interface{} `json:"settings" yaml:"settings"`
}

// CollectorStats ?
type CollectorStats struct {
	CollectedMetrics int64     `json:"collected_metrics"`
	Errors           int64     `json:"errors"`
	LastCollection   time.Time `json:"last_collection"`
	CollectionTime   time.Duration `json:"collection_time"`
}

// AnalysisStats 
type AnalysisStats struct {
	AnalyzedMetrics   int64     `json:"analyzed_metrics"`
	DetectedAnomalies int64     `json:"detected_anomalies"`
	GeneratedAlerts   int64     `json:"generated_alerts"`
	LastAnalysis      time.Time `json:"last_analysis"`
	AnalysisTime      time.Duration `json:"analysis_time"`
}

// SystemMetrics 
type SystemMetrics struct {
	Timestamp time.Time      `json:"timestamp"`
	CPU       *CPUMetrics    `json:"cpu,omitempty"`
	Memory    *MemoryMetrics `json:"memory,omitempty"`
	Disk      *DiskMetrics   `json:"disk,omitempty"`
	Network   *NetworkMetrics `json:"network,omitempty"`
	Process   *ProcessMetrics `json:"process,omitempty"`
}

// CPUMetrics CPU
type CPUMetrics struct {
	Usage      float64            `json:"usage"`       // CPU?(%)
	LoadAvg1   float64            `json:"load_avg_1"`  // 1?
	LoadAvg5   float64            `json:"load_avg_5"`  // 5?
	LoadAvg15  float64            `json:"load_avg_15"` // 15?
	Cores      int                `json:"cores"`       // CPU?
	PerCore    map[string]float64 `json:"per_core"`    // 
	Frequency  float64            `json:"frequency"`   // CPU (MHz)
	Temperature float64           `json:"temperature"` // CPU (C)
}

// MemoryMetrics 
type MemoryMetrics struct {
	Total       uint64  `json:"total"`        // ?(bytes)
	Used        uint64  `json:"used"`         // ?(bytes)
	Free        uint64  `json:"free"`         //  (bytes)
	Available   uint64  `json:"available"`    //  (bytes)
	Usage       float64 `json:"usage"`        // ?(%)
	Cached      uint64  `json:"cached"`       //  (bytes)
	Buffers     uint64  `json:"buffers"`      // ?(bytes)
	SwapTotal   uint64  `json:"swap_total"`   // ?(bytes)
	SwapUsed    uint64  `json:"swap_used"`    // ?(bytes)
	SwapFree    uint64  `json:"swap_free"`    //  (bytes)
}

// DiskMetrics 
type DiskMetrics struct {
	Devices map[string]*DiskDeviceMetrics `json:"devices"` // 豸
}

// DiskDeviceMetrics 豸
type DiskDeviceMetrics struct {
	Total       uint64  `json:"total"`        // ?(bytes)
	Used        uint64  `json:"used"`         // ?(bytes)
	Free        uint64  `json:"free"`         //  (bytes)
	Usage       float64 `json:"usage"`        // ?(%)
	ReadBytes   uint64  `json:"read_bytes"`   // ?
	WriteBytes  uint64  `json:"write_bytes"`  // ?
	ReadOps     uint64  `json:"read_ops"`     // ?
	WriteOps    uint64  `json:"write_ops"`    // ?
	ReadTime    uint64  `json:"read_time"`    //  (ms)
	WriteTime   uint64  `json:"write_time"`   //  (ms)
	IOTime      uint64  `json:"io_time"`      // IO (ms)
	IOPS        float64 `json:"iops"`         // IO?
	Throughput  float64 `json:"throughput"`   // ?(bytes/s)
}

// NetworkMetrics 
type NetworkMetrics struct {
	Interfaces map[string]*NetworkInterfaceMetrics `json:"interfaces"` // 
}

// NetworkInterfaceMetrics 
type NetworkInterfaceMetrics struct {
	BytesReceived    uint64  `json:"bytes_received"`    // ?
	BytesSent        uint64  `json:"bytes_sent"`        // 
	PacketsReceived  uint64  `json:"packets_received"`  // 
	PacketsSent      uint64  `json:"packets_sent"`      // ?
	ErrorsReceived   uint64  `json:"errors_received"`   // ?
	ErrorsSent       uint64  `json:"errors_sent"`       // 
	DroppedReceived  uint64  `json:"dropped_received"`  // ?
	DroppedSent      uint64  `json:"dropped_sent"`      // 
	Speed            uint64  `json:"speed"`             //  (bits/s)
	Duplex           string  `json:"duplex"`            // 
	MTU              int     `json:"mtu"`               // 䵥?
	RxRate           float64 `json:"rx_rate"`           //  (bytes/s)
	TxRate           float64 `json:"tx_rate"`           //  (bytes/s)
}

// ProcessMetrics 
type ProcessMetrics struct {
	Count       int                        `json:"count"`       // 
	Running     int                        `json:"running"`     // 
	Sleeping    int                        `json:"sleeping"`    // ?
	Stopped     int                        `json:"stopped"`     // ?
	Zombie      int                        `json:"zombie"`      // ?
	TopCPU      []*ProcessInfo             `json:"top_cpu"`     // CPU
	TopMemory   []*ProcessInfo             `json:"top_memory"`  // 
	Details     map[string]*ProcessInfo    `json:"details"`     // 
}

// ProcessInfo 
type ProcessInfo struct {
	PID         int     `json:"pid"`          // ID
	Name        string  `json:"name"`         // 
	Command     string  `json:"command"`      // ?
	CPUUsage    float64 `json:"cpu_usage"`    // CPU?(%)
	MemoryUsage uint64  `json:"memory_usage"` // ?(bytes)
	MemoryPercent float64 `json:"memory_percent"` // ?(%)
	Status      string  `json:"status"`       // ?
	StartTime   time.Time `json:"start_time"` // 
	User        string  `json:"user"`         // 
	Threads     int     `json:"threads"`      // ?
	FDs         int     `json:"fds"`          // 
}

// TimeRange 
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// PerformanceReport 
type PerformanceReport struct {
	TimeRange    TimeRange  `json:"time_range"`
	Timestamp    time.Time  `json:"timestamp"`
	OverallScore float64    `json:"overall_score"` //  (0-100)
	Anomalies    []*Anomaly `json:"anomalies"`     // ?
	Trends       []*Trend   `json:"trends"`        // 
	Alerts       []*Alert   `json:"alerts"`        // ?
}

// Anomaly 
type Anomaly struct {
	MetricName  string          `json:"metric_name"`
	Timestamp   time.Time       `json:"timestamp"`
	Value       float64         `json:"value"`
	Expected    float64         `json:"expected"`
	Deviation   float64         `json:"deviation"`
	Severity    AnomalySeverity `json:"severity"`
	Description string          `json:"description"`
}

// AnomalySeverity 
type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "low"
	AnomalySeverityMedium   AnomalySeverity = "medium"
	AnomalySeverityHigh     AnomalySeverity = "high"
	AnomalySeverityCritical AnomalySeverity = "critical"
)

// Trend 
type Trend struct {
	MetricName  string         `json:"metric_name"`
	TimeRange   TimeRange      `json:"time_range"`
	Direction   TrendDirection `json:"direction"`
	Slope       float64        `json:"slope"`
	Confidence  float64        `json:"confidence"` // ?(0-1)
	Description string         `json:"description"`
}

// TrendDirection 
type TrendDirection string

const (
	TrendDirectionIncreasing TrendDirection = "increasing"
	TrendDirectionDecreasing TrendDirection = "decreasing"
	TrendDirectionStable     TrendDirection = "stable"
)

// Alert 澯
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

// AlertSeverity 澯
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus 澯?
type AlertStatus string

const (
	AlertStatusActive    AlertStatus = "active"
	AlertStatusSuppressed AlertStatus = "suppressed"
	AlertStatusResolved  AlertStatus = "resolved"
)

// AnalysisResult 
type AnalysisResult struct {
	Anomalies []*Anomaly `json:"anomalies"`
	Trends    []*Trend   `json:"trends"`
	Alerts    []*Alert   `json:"alerts"`
}

// MetricThreshold ?
type MetricThreshold struct {
	MetricName string    `json:"metric_name"`
	Warning    float64   `json:"warning"`
	Critical   float64   `json:"critical"`
	Operator   string    `json:"operator"` // >, <, >=, <=, ==, !=
}

// PerformanceProfile 
type PerformanceProfile struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Thresholds  []*MetricThreshold `json:"thresholds"`
	Rules       []*AnalysisRule    `json:"rules"`
	Enabled     bool               `json:"enabled"`
}

// AnalysisRule 
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

// ResourceUsage 
type ResourceUsage struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`       // CPU?(%)
	Memory    float64   `json:"memory"`    // ?(%)
	Disk      float64   `json:"disk"`      // ?(%)
	Network   float64   `json:"network"`   // ?(%)
	Load      float64   `json:"load"`      // 
}

// PerformanceBaseline 
type PerformanceBaseline struct {
	MetricName string    `json:"metric_name"`
	Mean       float64   `json:"mean"`
	StdDev     float64   `json:"std_dev"`
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	Percentiles map[string]float64 `json:"percentiles"` // P50, P90, P95, P99
	UpdatedAt  time.Time `json:"updated_at"`
}

// CapacityPrediction 
type CapacityPrediction struct {
	MetricName     string    `json:"metric_name"`
	CurrentValue   float64   `json:"current_value"`
	PredictedValue float64   `json:"predicted_value"`
	PredictionTime time.Time `json:"prediction_time"`
	Confidence     float64   `json:"confidence"`
	Method         string    `json:"method"` // linear, exponential, seasonal
	TimeToLimit    *time.Duration `json:"time_to_limit,omitempty"`
}

// OptimizationSuggestion 
type OptimizationSuggestion struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`        // cpu, memory, disk, network
	Priority    string            `json:"priority"`    // low, medium, high, critical
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Impact      string            `json:"impact"`      // 
	Effort      string            `json:"effort"`      // 
	Actions     []string          `json:"actions"`     // ?
	Metrics     []string          `json:"metrics"`     // 
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
}

// HealthScore 
type HealthScore struct {
	Overall    float64            `json:"overall"`    //  (0-100)
	Components map[string]float64 `json:"components"` // ?
	Timestamp  time.Time          `json:"timestamp"`
	Details    map[string]string  `json:"details"`    // 
}

