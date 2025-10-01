package database

import (
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// MetricsCollector 指标收集�?
type MetricsCollector struct {
	logger    *zap.Logger
	mutex     sync.RWMutex
	
	// 查询指标
	totalQueries      int64
	successfulQueries int64
	failedQueries     int64
	
	// 响应时间指标
	totalResponseTime int64  // 纳秒
	minResponseTime   int64  // 纳秒
	maxResponseTime   int64  // 纳秒
	
	// 连接指标
	connectionMetrics map[string]*ConnectionMetrics
	
	// 错误指标
	errorMetrics map[string]*ErrorMetrics
	
	// 时间窗口指标
	windowMetrics *WindowMetrics
}

// ConnectionMetrics 连接指标
type ConnectionMetrics struct {
	PoolName            string
	TotalConnections    int64
	ActiveConnections   int64
	IdleConnections     int64
	ConnectionsCreated  int64
	ConnectionsClosed   int64
	ConnectionErrors    int64
	AverageWaitTime     time.Duration
	MaxWaitTime         time.Duration
	LastUpdated         time.Time
	mutex               sync.RWMutex
}

// ErrorMetrics 错误指标
type ErrorMetrics struct {
	ErrorType    string
	Count        int64
	LastOccurred time.Time
	Details      []ErrorDetail
	mutex        sync.RWMutex
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Timestamp time.Time
	Message   string
	Context   map[string]interface{}
}

// WindowMetrics 时间窗口指标
type WindowMetrics struct {
	WindowSize    time.Duration
	CurrentWindow *TimeWindow
	PreviousWindow *TimeWindow
	mutex         sync.RWMutex
}

// TimeWindow 时间窗口
type TimeWindow struct {
	StartTime     time.Time
	EndTime       time.Time
	QueryCount    int64
	ErrorCount    int64
	TotalLatency  time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
}

// QueryMetrics 查询指标
type QueryMetrics struct {
	QueryType     string
	Duration      time.Duration
	Success       bool
	Error         error
	Timestamp     time.Time
	ConnectionID  string
	Context       map[string]interface{}
}

// NewMetricsCollector 创建新的指标收集�?
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:            logger,
		connectionMetrics: make(map[string]*ConnectionMetrics),
		errorMetrics:      make(map[string]*ErrorMetrics),
		windowMetrics: &WindowMetrics{
			WindowSize: 5 * time.Minute,
			CurrentWindow: &TimeWindow{
				StartTime:  time.Now(),
				MinLatency: time.Duration(^uint64(0) >> 1), // 最大�?
			},
		},
		minResponseTime: int64(^uint64(0) >> 1), // 最大�?
	}
}

// RecordQuery 记录查询指标
func (mc *MetricsCollector) RecordQuery(metrics QueryMetrics) {
	atomic.AddInt64(&mc.totalQueries, 1)
	
	if metrics.Success {
		atomic.AddInt64(&mc.successfulQueries, 1)
	} else {
		atomic.AddInt64(&mc.failedQueries, 1)
		
		// 记录错误
		if metrics.Error != nil {
			mc.recordError("query_error", metrics.Error.Error(), map[string]interface{}{
				"query_type":    metrics.QueryType,
				"connection_id": metrics.ConnectionID,
				"timestamp":     metrics.Timestamp,
			})
		}
	}
	
	// 更新响应时间指标
	durationNanos := metrics.Duration.Nanoseconds()
	atomic.AddInt64(&mc.totalResponseTime, durationNanos)
	
	// 更新最小响应时�?
	for {
		current := atomic.LoadInt64(&mc.minResponseTime)
		if durationNanos >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&mc.minResponseTime, current, durationNanos) {
			break
		}
	}
	
	// 更新最大响应时�?
	for {
		current := atomic.LoadInt64(&mc.maxResponseTime)
		if durationNanos <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&mc.maxResponseTime, current, durationNanos) {
			break
		}
	}
	
	// 更新时间窗口指标
	mc.updateWindowMetrics(metrics)
	
	// 记录详细日志
	if metrics.Success {
		mc.logger.Debug("Query executed successfully",
			zap.String("query_type", metrics.QueryType),
			zap.Duration("duration", metrics.Duration),
			zap.String("connection_id", metrics.ConnectionID),
			zap.Time("timestamp", metrics.Timestamp),
		)
	} else {
		mc.logger.Error("Query execution failed",
			zap.String("query_type", metrics.QueryType),
			zap.Duration("duration", metrics.Duration),
			zap.String("connection_id", metrics.ConnectionID),
			zap.Error(metrics.Error),
			zap.Time("timestamp", metrics.Timestamp),
		)
	}
}

// UpdateConnectionMetrics 更新连接指标
func (mc *MetricsCollector) UpdateConnectionMetrics(poolName string, stats ConnectionStats) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if mc.connectionMetrics[poolName] == nil {
		mc.connectionMetrics[poolName] = &ConnectionMetrics{
			PoolName: poolName,
		}
	}
	
	metrics := mc.connectionMetrics[poolName]
	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()
	
	metrics.TotalConnections = int64(stats.TotalConnections)
	metrics.ActiveConnections = int64(stats.ActiveConnections)
	metrics.IdleConnections = int64(stats.IdleConnections)
	metrics.ConnectionsCreated = stats.ConnectionsCreated
	metrics.ConnectionsClosed = stats.ConnectionsClosed
	metrics.ConnectionErrors = stats.ConnectionErrors
	metrics.AverageWaitTime = stats.AverageWaitTime
	metrics.MaxWaitTime = stats.MaxWaitTime
	metrics.LastUpdated = time.Now()
	
	mc.logger.Debug("Connection metrics updated",
		zap.String("pool_name", poolName),
		zap.Int64("total_connections", metrics.TotalConnections),
		zap.Int64("active_connections", metrics.ActiveConnections),
		zap.Int64("idle_connections", metrics.IdleConnections),
		zap.Int64("connection_errors", metrics.ConnectionErrors),
		zap.Duration("average_wait_time", metrics.AverageWaitTime),
	)
}

// recordError 记录错误
func (mc *MetricsCollector) recordError(errorType, message string, context map[string]interface{}) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	if mc.errorMetrics[errorType] == nil {
		mc.errorMetrics[errorType] = &ErrorMetrics{
			ErrorType: errorType,
			Details:   make([]ErrorDetail, 0),
		}
	}
	
	errorMetric := mc.errorMetrics[errorType]
	errorMetric.mutex.Lock()
	defer errorMetric.mutex.Unlock()
	
	errorMetric.Count++
	errorMetric.LastOccurred = time.Now()
	
	// 添加错误详情（保留最�?00条）
	detail := ErrorDetail{
		Timestamp: time.Now(),
		Message:   message,
		Context:   context,
	}
	
	errorMetric.Details = append(errorMetric.Details, detail)
	if len(errorMetric.Details) > 100 {
		errorMetric.Details = errorMetric.Details[1:]
	}
}

// updateWindowMetrics 更新时间窗口指标
func (mc *MetricsCollector) updateWindowMetrics(metrics QueryMetrics) {
	mc.windowMetrics.mutex.Lock()
	defer mc.windowMetrics.mutex.Unlock()
	
	now := time.Now()
	currentWindow := mc.windowMetrics.CurrentWindow
	
	// 检查是否需要切换窗�?
	if now.Sub(currentWindow.StartTime) >= mc.windowMetrics.WindowSize {
		// 保存当前窗口为上一个窗�?
		mc.windowMetrics.PreviousWindow = currentWindow
		
		// 创建新的当前窗口
		mc.windowMetrics.CurrentWindow = &TimeWindow{
			StartTime:  now,
			MinLatency: time.Duration(^uint64(0) >> 1), // 最大�?
		}
		currentWindow = mc.windowMetrics.CurrentWindow
	}
	
	// 更新当前窗口指标
	currentWindow.EndTime = now
	currentWindow.QueryCount++
	
	if !metrics.Success {
		currentWindow.ErrorCount++
	}
	
	currentWindow.TotalLatency += metrics.Duration
	
	if metrics.Duration < currentWindow.MinLatency {
		currentWindow.MinLatency = metrics.Duration
	}
	
	if metrics.Duration > currentWindow.MaxLatency {
		currentWindow.MaxLatency = metrics.Duration
	}
}

// GetSummaryMetrics 获取汇总指�?
func (mc *MetricsCollector) GetSummaryMetrics() map[string]interface{} {
	totalQueries := atomic.LoadInt64(&mc.totalQueries)
	successfulQueries := atomic.LoadInt64(&mc.successfulQueries)
	failedQueries := atomic.LoadInt64(&mc.failedQueries)
	totalResponseTime := atomic.LoadInt64(&mc.totalResponseTime)
	minResponseTime := atomic.LoadInt64(&mc.minResponseTime)
	maxResponseTime := atomic.LoadInt64(&mc.maxResponseTime)
	
	var averageResponseTime time.Duration
	if totalQueries > 0 {
		averageResponseTime = time.Duration(totalResponseTime / totalQueries)
	}
	
	var successRate float64
	if totalQueries > 0 {
		successRate = float64(successfulQueries) / float64(totalQueries)
	}
	
	return map[string]interface{}{
		"total_queries":        totalQueries,
		"successful_queries":   successfulQueries,
		"failed_queries":       failedQueries,
		"success_rate":         successRate,
		"average_response_time": averageResponseTime,
		"min_response_time":    time.Duration(minResponseTime),
		"max_response_time":    time.Duration(maxResponseTime),
	}
}

// GetConnectionMetrics 获取连接指标
func (mc *MetricsCollector) GetConnectionMetrics() map[string]*ConnectionMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	result := make(map[string]*ConnectionMetrics)
	for name, metrics := range mc.connectionMetrics {
		metrics.mutex.RLock()
		result[name] = &ConnectionMetrics{
			PoolName:            metrics.PoolName,
			TotalConnections:    metrics.TotalConnections,
			ActiveConnections:   metrics.ActiveConnections,
			IdleConnections:     metrics.IdleConnections,
			ConnectionsCreated:  metrics.ConnectionsCreated,
			ConnectionsClosed:   metrics.ConnectionsClosed,
			ConnectionErrors:    metrics.ConnectionErrors,
			AverageWaitTime:     metrics.AverageWaitTime,
			MaxWaitTime:         metrics.MaxWaitTime,
			LastUpdated:         metrics.LastUpdated,
		}
		metrics.mutex.RUnlock()
	}
	
	return result
}

// GetErrorMetrics 获取错误指标
func (mc *MetricsCollector) GetErrorMetrics() map[string]*ErrorMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	result := make(map[string]*ErrorMetrics)
	for errorType, metrics := range mc.errorMetrics {
		metrics.mutex.RLock()
		result[errorType] = &ErrorMetrics{
			ErrorType:    metrics.ErrorType,
			Count:        metrics.Count,
			LastOccurred: metrics.LastOccurred,
			Details:      append([]ErrorDetail{}, metrics.Details...),
		}
		metrics.mutex.RUnlock()
	}
	
	return result
}

// GetWindowMetrics 获取时间窗口指标
func (mc *MetricsCollector) GetWindowMetrics() *WindowMetrics {
	mc.windowMetrics.mutex.RLock()
	defer mc.windowMetrics.mutex.RUnlock()
	
	result := &WindowMetrics{
		WindowSize: mc.windowMetrics.WindowSize,
	}
	
	if mc.windowMetrics.CurrentWindow != nil {
		result.CurrentWindow = &TimeWindow{
			StartTime:    mc.windowMetrics.CurrentWindow.StartTime,
			EndTime:      mc.windowMetrics.CurrentWindow.EndTime,
			QueryCount:   mc.windowMetrics.CurrentWindow.QueryCount,
			ErrorCount:   mc.windowMetrics.CurrentWindow.ErrorCount,
			TotalLatency: mc.windowMetrics.CurrentWindow.TotalLatency,
			MinLatency:   mc.windowMetrics.CurrentWindow.MinLatency,
			MaxLatency:   mc.windowMetrics.CurrentWindow.MaxLatency,
		}
	}
	
	if mc.windowMetrics.PreviousWindow != nil {
		result.PreviousWindow = &TimeWindow{
			StartTime:    mc.windowMetrics.PreviousWindow.StartTime,
			EndTime:      mc.windowMetrics.PreviousWindow.EndTime,
			QueryCount:   mc.windowMetrics.PreviousWindow.QueryCount,
			ErrorCount:   mc.windowMetrics.PreviousWindow.ErrorCount,
			TotalLatency: mc.windowMetrics.PreviousWindow.TotalLatency,
			MinLatency:   mc.windowMetrics.PreviousWindow.MinLatency,
			MaxLatency:   mc.windowMetrics.PreviousWindow.MaxLatency,
		}
	}
	
	return result
}

// LogPerformanceReport 记录性能报告
func (mc *MetricsCollector) LogPerformanceReport() {
	summary := mc.GetSummaryMetrics()
	connectionMetrics := mc.GetConnectionMetrics()
	errorMetrics := mc.GetErrorMetrics()
	windowMetrics := mc.GetWindowMetrics()
	
	mc.logger.Info("Database performance report",
		zap.Any("summary", summary),
		zap.Int("connection_pools", len(connectionMetrics)),
		zap.Int("error_types", len(errorMetrics)),
	)
	
	// 记录连接池详�?
	for poolName, metrics := range connectionMetrics {
		mc.logger.Info("Connection pool performance",
			zap.String("pool_name", poolName),
			zap.Int64("total_connections", metrics.TotalConnections),
			zap.Int64("active_connections", metrics.ActiveConnections),
			zap.Int64("idle_connections", metrics.IdleConnections),
			zap.Int64("connections_created", metrics.ConnectionsCreated),
			zap.Int64("connections_closed", metrics.ConnectionsClosed),
			zap.Int64("connection_errors", metrics.ConnectionErrors),
			zap.Duration("average_wait_time", metrics.AverageWaitTime),
			zap.Duration("max_wait_time", metrics.MaxWaitTime),
		)
	}
	
	// 记录错误统计
	for errorType, metrics := range errorMetrics {
		mc.logger.Warn("Error statistics",
			zap.String("error_type", errorType),
			zap.Int64("count", metrics.Count),
			zap.Time("last_occurred", metrics.LastOccurred),
			zap.Int("recent_details", len(metrics.Details)),
		)
	}
	
	// 记录时间窗口指标
	if windowMetrics.CurrentWindow != nil {
		window := windowMetrics.CurrentWindow
		var avgLatency time.Duration
		if window.QueryCount > 0 {
			avgLatency = window.TotalLatency / time.Duration(window.QueryCount)
		}
		
		mc.logger.Info("Current window performance",
			zap.Duration("window_size", windowMetrics.WindowSize),
			zap.Time("start_time", window.StartTime),
			zap.Time("end_time", window.EndTime),
			zap.Int64("query_count", window.QueryCount),
			zap.Int64("error_count", window.ErrorCount),
			zap.Duration("average_latency", avgLatency),
			zap.Duration("min_latency", window.MinLatency),
			zap.Duration("max_latency", window.MaxLatency),
		)
	}
}

// Reset 重置所有指�?
func (mc *MetricsCollector) Reset() {
	atomic.StoreInt64(&mc.totalQueries, 0)
	atomic.StoreInt64(&mc.successfulQueries, 0)
	atomic.StoreInt64(&mc.failedQueries, 0)
	atomic.StoreInt64(&mc.totalResponseTime, 0)
	atomic.StoreInt64(&mc.minResponseTime, int64(^uint64(0)>>1))
	atomic.StoreInt64(&mc.maxResponseTime, 0)
	
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.connectionMetrics = make(map[string]*ConnectionMetrics)
	mc.errorMetrics = make(map[string]*ErrorMetrics)
	
	mc.windowMetrics.mutex.Lock()
	mc.windowMetrics.CurrentWindow = &TimeWindow{
		StartTime:  time.Now(),
		MinLatency: time.Duration(^uint64(0) >> 1),
	}
	mc.windowMetrics.PreviousWindow = nil
	mc.windowMetrics.mutex.Unlock()
	
	mc.logger.Info("Metrics collector reset")
}
