package models

import (
	"time"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"   // 计数器
	MetricTypeGauge     MetricType = "gauge"     // 仪表盘
	MetricTypeHistogram MetricType = "histogram" // 直方图
	MetricTypeSummary   MetricType = "summary"   // 摘要
)

// MetricCategory 指标分类
type MetricCategory string

const (
	CategorySystem      MetricCategory = "system"      // 系统指标
	CategoryApplication MetricCategory = "application" // 应用指标
	CategoryBusiness    MetricCategory = "business"    // 业务指标
	CategoryDatabase    MetricCategory = "database"    // 数据库指标
	CategoryNetwork     MetricCategory = "network"     // 网络指标
	CategorySecurity    MetricCategory = "security"    // 安全指标
	CategoryCustom      MetricCategory = "custom"      // 自定义指标
)

// Metric 指标基础结构
type Metric struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"not null;index"`
	Type        MetricType        `json:"type" gorm:"not null"`
	Category    MetricCategory    `json:"category" gorm:"not null;index"`
	Description string            `json:"description"`
	Unit        string            `json:"unit"`
	Labels      map[string]string `json:"labels" gorm:"type:jsonb"`
	Value       float64           `json:"value"`
	Timestamp   time.Time         `json:"timestamp" gorm:"not null;index"`
	Source      string            `json:"source" gorm:"index"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// MetricSample 指标样本
type MetricSample struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Value      float64           `json:"value"`
	Timestamp  time.Time         `json:"timestamp"`
}

// MetricSeries 指标时间序列
type MetricSeries struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Samples    []MetricSample    `json:"samples"`
}

// CounterMetric 计数器指标
type CounterMetric struct {
	Metric
	Total float64 `json:"total"`
	Rate  float64 `json:"rate"` // 每秒增长率
}

// GaugeMetric 仪表盘指标
type GaugeMetric struct {
	Metric
	Current float64 `json:"current"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
}

// HistogramMetric 直方图指标
type HistogramMetric struct {
	Metric
	Buckets    []HistogramBucket `json:"buckets"`
	Count      uint64            `json:"count"`
	Sum        float64           `json:"sum"`
	Quantiles  map[float64]float64 `json:"quantiles"` // 分位数
}

// HistogramBucket 直方图桶
type HistogramBucket struct {
	UpperBound float64 `json:"upper_bound"`
	Count      uint64  `json:"count"`
}

// SummaryMetric 摘要指标
type SummaryMetric struct {
	Metric
	Count     uint64              `json:"count"`
	Sum       float64             `json:"sum"`
	Quantiles map[float64]float64 `json:"quantiles"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	DiskUsage      float64 `json:"disk_usage"`
	NetworkIO      NetworkIOMetrics `json:"network_io"`
	LoadAverage    LoadAverageMetrics `json:"load_average"`
	ProcessCount   int64   `json:"process_count"`
	FileDescriptors int64  `json:"file_descriptors"`
	Timestamp      time.Time `json:"timestamp"`
}

// NetworkIOMetrics 网络IO指标
type NetworkIOMetrics struct {
	BytesReceived    uint64 `json:"bytes_received"`
	BytesSent        uint64 `json:"bytes_sent"`
	PacketsReceived  uint64 `json:"packets_received"`
	PacketsSent      uint64 `json:"packets_sent"`
	ErrorsReceived   uint64 `json:"errors_received"`
	ErrorsSent       uint64 `json:"errors_sent"`
	DroppedReceived  uint64 `json:"dropped_received"`
	DroppedSent      uint64 `json:"dropped_sent"`
}

// LoadAverageMetrics 负载平均值指标
type LoadAverageMetrics struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// ApplicationMetrics 应用指标
type ApplicationMetrics struct {
	HTTPRequests    HTTPMetrics    `json:"http_requests"`
	GRPCRequests    GRPCMetrics    `json:"grpc_requests"`
	DatabaseQueries DatabaseMetrics `json:"database_queries"`
	CacheOperations CacheMetrics   `json:"cache_operations"`
	Goroutines      int64          `json:"goroutines"`
	GCMetrics       GCMetrics      `json:"gc_metrics"`
	Timestamp       time.Time      `json:"timestamp"`
}

// HTTPMetrics HTTP指标
type HTTPMetrics struct {
	RequestsTotal    uint64            `json:"requests_total"`
	RequestsPerSecond float64          `json:"requests_per_second"`
	ResponseTime     ResponseTimeMetrics `json:"response_time"`
	StatusCodes      map[string]uint64 `json:"status_codes"`
	ErrorRate        float64           `json:"error_rate"`
}

// ResponseTimeMetrics 响应时间指标
type ResponseTimeMetrics struct {
	P50  float64 `json:"p50"`
	P90  float64 `json:"p90"`
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
	Mean float64 `json:"mean"`
	Max  float64 `json:"max"`
}

// GRPCMetrics GRPC指标
type GRPCMetrics struct {
	RequestsTotal     uint64            `json:"requests_total"`
	RequestsPerSecond float64           `json:"requests_per_second"`
	ResponseTime      ResponseTimeMetrics `json:"response_time"`
	StatusCodes       map[string]uint64 `json:"status_codes"`
	ErrorRate         float64           `json:"error_rate"`
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	ConnectionsActive   int64   `json:"connections_active"`
	ConnectionsIdle     int64   `json:"connections_idle"`
	ConnectionsMax      int64   `json:"connections_max"`
	QueriesTotal        uint64  `json:"queries_total"`
	QueriesPerSecond    float64 `json:"queries_per_second"`
	SlowQueries         uint64  `json:"slow_queries"`
	QueryDuration       ResponseTimeMetrics `json:"query_duration"`
	LockWaitTime        float64 `json:"lock_wait_time"`
	DeadlockCount       uint64  `json:"deadlock_count"`
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	HitRate         float64 `json:"hit_rate"`
	MissRate        float64 `json:"miss_rate"`
	OperationsTotal uint64  `json:"operations_total"`
	KeysTotal       uint64  `json:"keys_total"`
	MemoryUsage     uint64  `json:"memory_usage"`
	EvictionsTotal  uint64  `json:"evictions_total"`
}

// GCMetrics 垃圾回收指标
type GCMetrics struct {
	GCCount       uint64  `json:"gc_count"`
	GCDuration    float64 `json:"gc_duration"`
	HeapSize      uint64  `json:"heap_size"`
	HeapUsed      uint64  `json:"heap_used"`
	HeapObjects   uint64  `json:"heap_objects"`
	NextGC        uint64  `json:"next_gc"`
	LastGC        time.Time `json:"last_gc"`
}

// BusinessMetrics 业务指标
type BusinessMetrics struct {
	UserMetrics         UserMetrics         `json:"user_metrics"`
	TransactionMetrics  TransactionMetrics  `json:"transaction_metrics"`
	RevenueMetrics      RevenueMetrics      `json:"revenue_metrics"`
	ConversionMetrics   ConversionMetrics   `json:"conversion_metrics"`
	FeatureUsageMetrics FeatureUsageMetrics `json:"feature_usage_metrics"`
	Timestamp           time.Time           `json:"timestamp"`
}

// UserMetrics 用户指标
type UserMetrics struct {
	ActiveUsers     int64 `json:"active_users"`
	NewUsers        int64 `json:"new_users"`
	ReturningUsers  int64 `json:"returning_users"`
	SessionDuration float64 `json:"session_duration"`
	PageViews       int64 `json:"page_views"`
	BounceRate      float64 `json:"bounce_rate"`
}

// TransactionMetrics 交易指标
type TransactionMetrics struct {
	TransactionsTotal   uint64  `json:"transactions_total"`
	TransactionsSuccess uint64  `json:"transactions_success"`
	TransactionsFailed  uint64  `json:"transactions_failed"`
	SuccessRate         float64 `json:"success_rate"`
	FailureRate         float64 `json:"failure_rate"`
	AverageValue        float64 `json:"average_value"`
	TotalValue          float64 `json:"total_value"`
}

// RevenueMetrics 收入指标
type RevenueMetrics struct {
	TotalRevenue    float64 `json:"total_revenue"`
	RevenuePerUser  float64 `json:"revenue_per_user"`
	RevenueGrowth   float64 `json:"revenue_growth"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`
	DailyRevenue    float64 `json:"daily_revenue"`
	HourlyRevenue   float64 `json:"hourly_revenue"`
}

// ConversionMetrics 转化指标
type ConversionMetrics struct {
	ConversionRate      float64 `json:"conversion_rate"`
	FunnelConversions   map[string]float64 `json:"funnel_conversions"`
	AbandonmentRate     float64 `json:"abandonment_rate"`
	TimeToConversion    float64 `json:"time_to_conversion"`
}

// FeatureUsageMetrics 功能使用指标
type FeatureUsageMetrics struct {
	FeatureUsage    map[string]uint64 `json:"feature_usage"`
	PopularFeatures []string          `json:"popular_features"`
	UnusedFeatures  []string          `json:"unused_features"`
	FeatureAdoption map[string]float64 `json:"feature_adoption"`
}

// MetricQuery 指标查询
type MetricQuery struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Step       time.Duration     `json:"step"`
	Aggregation string           `json:"aggregation"` // sum, avg, min, max, count
}

// MetricQueryResult 指标查询结果
type MetricQueryResult struct {
	MetricName string         `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	Values     []MetricValue  `json:"values"`
}

// MetricValue 指标值
type MetricValue struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// AggregatedMetric 聚合指标
type AggregatedMetric struct {
	MetricName  string            `json:"metric_name"`
	Labels      map[string]string `json:"labels"`
	Aggregation string            `json:"aggregation"`
	Value       float64           `json:"value"`
	Count       int64             `json:"count"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
}

// MetricThreshold 指标阈值
type MetricThreshold struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	MetricName  string            `json:"metric_name" gorm:"not null;index"`
	Labels      map[string]string `json:"labels" gorm:"type:jsonb"`
	Operator    string            `json:"operator"` // >, <, >=, <=, ==, !=
	Value       float64           `json:"value"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"` // info, warning, critical, emergency
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// MetricAnnotation 指标注释
type MetricAnnotation struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	MetricName  string            `json:"metric_name" gorm:"not null;index"`
	Labels      map[string]string `json:"labels" gorm:"type:jsonb"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags" gorm:"type:jsonb"`
	StartTime   time.Time         `json:"start_time" gorm:"index"`
	EndTime     *time.Time        `json:"end_time,omitempty" gorm:"index"`
	CreatedBy   string            `json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// MetricMetadata 指标元数据
type MetricMetadata struct {
	MetricName  string            `json:"metric_name" gorm:"primaryKey"`
	Type        MetricType        `json:"type" gorm:"not null"`
	Category    MetricCategory    `json:"category" gorm:"not null"`
	Description string            `json:"description"`
	Unit        string            `json:"unit"`
	Help        string            `json:"help"`
	Labels      []string          `json:"labels" gorm:"type:jsonb"`
	Retention   time.Duration     `json:"retention"`
	SampleRate  float64           `json:"sample_rate"`
	Enabled     bool              `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// MetricExport 指标导出配置
type MetricExport struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	Name        string            `json:"name" gorm:"not null"`
	MetricNames []string          `json:"metric_names" gorm:"type:jsonb"`
	Labels      map[string]string `json:"labels" gorm:"type:jsonb"`
	Format      string            `json:"format"` // prometheus, json, csv
	Destination string            `json:"destination"` // file, http, s3
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	Schedule    string            `json:"schedule"` // cron expression
	Enabled     bool              `json:"enabled" gorm:"default:true"`
	LastExport  *time.Time        `json:"last_export,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// Helper functions

// NewMetric 创建新指标
func NewMetric(name string, metricType MetricType, category MetricCategory) *Metric {
	return &Metric{
		ID:        generateID(),
		Name:      name,
		Type:      metricType,
		Category:  category,
		Labels:    make(map[string]string),
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// WithLabels 添加标签
func (m *Metric) WithLabels(labels map[string]string) *Metric {
	if m.Labels == nil {
		m.Labels = make(map[string]string)
	}
	for k, v := range labels {
		m.Labels[k] = v
	}
	return m
}

// WithValue 设置值
func (m *Metric) WithValue(value float64) *Metric {
	m.Value = value
	m.UpdatedAt = time.Now()
	return m
}

// WithSource 设置来源
func (m *Metric) WithSource(source string) *Metric {
	m.Source = source
	return m
}

// IsExpired 检查指标是否过期
func (m *Metric) IsExpired(retention time.Duration) bool {
	return time.Since(m.Timestamp) > retention
}

// GetLabelValue 获取标签值
func (m *Metric) GetLabelValue(key string) (string, bool) {
	if m.Labels == nil {
		return "", false
	}
	value, exists := m.Labels[key]
	return value, exists
}

// MatchesLabels 检查标签是否匹配
func (m *Metric) MatchesLabels(labels map[string]string) bool {
	if m.Labels == nil && len(labels) == 0 {
		return true
	}
	if m.Labels == nil || len(labels) == 0 {
		return false
	}
	
	for k, v := range labels {
		if labelValue, exists := m.Labels[k]; !exists || labelValue != v {
			return false
		}
	}
	return true
}

// generateID 生成唯一ID
func generateID() string {
	// 实现ID生成逻辑
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	// 实现随机字符串生成逻辑
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}