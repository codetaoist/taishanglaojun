package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/interfaces"
)

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	config     AnalyzerConfig
	collectors map[string]MetricCollector
	analyzers  map[string]Analyzer
	storage    interfaces.MetricStorage
	stats      *AnalyzerStats
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// AnalyzerConfig 分析器配置
type AnalyzerConfig struct {
	// 收集配置
	CollectionInterval time.Duration `json:"collection_interval" yaml:"collection_interval"`
	MetricsRetention   time.Duration `json:"metrics_retention" yaml:"metrics_retention"`
	
	// 分析配置
	AnalysisInterval   time.Duration `json:"analysis_interval" yaml:"analysis_interval"`
	AnomalyThreshold   float64       `json:"anomaly_threshold" yaml:"anomaly_threshold"`
	TrendWindow        time.Duration `json:"trend_window" yaml:"trend_window"`
	
	// 存储配置
	StorageEnabled     bool   `json:"storage_enabled" yaml:"storage_enabled"`
	StorageBackend     string `json:"storage_backend" yaml:"storage_backend"`
	
	// 告警配置
	AlertEnabled       bool          `json:"alert_enabled" yaml:"alert_enabled"`
	AlertThresholds    map[string]float64 `json:"alert_thresholds" yaml:"alert_thresholds"`
	
	// 收集器配置
	Collectors map[string]CollectorConfig `json:"collectors" yaml:"collectors"`
	
	// 分析器配置
	Analyzers map[string]AnalyzerSettings `json:"analyzers" yaml:"analyzers"`
}

// AnalyzerStats 分析器统计信息
type AnalyzerStats struct {
	CollectedMetrics   int64         `json:"collected_metrics"`
	AnalyzedMetrics    int64         `json:"analyzed_metrics"`
	DetectedAnomalies  int64         `json:"detected_anomalies"`
	GeneratedAlerts    int64         `json:"generated_alerts"`
	LastCollection     time.Time     `json:"last_collection"`
	LastAnalysis       time.Time     `json:"last_analysis"`
	CollectionTime     time.Duration `json:"collection_time"`
	AnalysisTime       time.Duration `json:"analysis_time"`
	
	// 收集器统计
	CollectorStats map[string]*CollectorStats `json:"collector_stats"`
	
	// 分析器统计
	AnalyzerStats map[string]*AnalysisStats `json:"analyzer_stats"`
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer(config AnalyzerConfig, storage interfaces.MetricStorage) (*PerformanceAnalyzer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 设置默认值
	if config.CollectionInterval == 0 {
		config.CollectionInterval = 10 * time.Second
	}
	if config.AnalysisInterval == 0 {
		config.AnalysisInterval = time.Minute
	}
	if config.AnomalyThreshold == 0 {
		config.AnomalyThreshold = 2.0 // 2个标准差
	}
	if config.TrendWindow == 0 {
		config.TrendWindow = time.Hour
	}
	if config.MetricsRetention == 0 {
		config.MetricsRetention = 24 * time.Hour
	}
	
	analyzer := &PerformanceAnalyzer{
		config:     config,
		collectors: make(map[string]MetricCollector),
		analyzers:  make(map[string]Analyzer),
		storage:    storage,
		stats: &AnalyzerStats{
			CollectorStats: make(map[string]*CollectorStats),
			AnalyzerStats:  make(map[string]*AnalysisStats),
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// 初始化收集器
	if err := analyzer.initCollectors(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize collectors: %w", err)
	}
	
	// 初始化分析器
	if err := analyzer.initAnalyzers(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize analyzers: %w", err)
	}
	
	return analyzer, nil
}

// Start 启动性能分析器
func (pa *PerformanceAnalyzer) Start() error {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	// 启动收集器
	for name, collector := range pa.collectors {
		if err := collector.Start(); err != nil {
			return fmt.Errorf("failed to start collector %s: %w", name, err)
		}
	}
	
	// 启动分析器
	for name, analyzer := range pa.analyzers {
		if err := analyzer.Start(); err != nil {
			return fmt.Errorf("failed to start analyzer %s: %w", name, err)
		}
	}
	
	// 启动收集循环
	pa.wg.Add(1)
	go pa.collectionLoop()
	
	// 启动分析循环
	pa.wg.Add(1)
	go pa.analysisLoop()
	
	// 启动清理循环
	pa.wg.Add(1)
	go pa.cleanupLoop()
	
	return nil
}

// Stop 停止性能分析器
func (pa *PerformanceAnalyzer) Stop() error {
	pa.cancel()
	pa.wg.Wait()
	
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	// 停止收集器
	for _, collector := range pa.collectors {
		collector.Stop()
	}
	
	// 停止分析器
	for _, analyzer := range pa.analyzers {
		analyzer.Stop()
	}
	
	return nil
}

// GetStats 获取统计信息
func (pa *PerformanceAnalyzer) GetStats() *AnalyzerStats {
	pa.mutex.RLock()
	defer pa.mutex.RUnlock()
	
	// 复制统计信息
	stats := *pa.stats
	
	// 复制收集器统计
	stats.CollectorStats = make(map[string]*CollectorStats)
	for name, stat := range pa.stats.CollectorStats {
		statCopy := *stat
		stats.CollectorStats[name] = &statCopy
	}
	
	// 复制分析器统计
	stats.AnalyzerStats = make(map[string]*AnalysisStats)
	for name, stat := range pa.stats.AnalyzerStats {
		statCopy := *stat
		stats.AnalyzerStats[name] = &statCopy
	}
	
	return &stats
}

// GetSystemMetrics 获取系统指标
func (pa *PerformanceAnalyzer) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}
	
	// CPU指标
	if collector, exists := pa.collectors["cpu"]; exists {
		if cpuCollector, ok := collector.(*CPUCollector); ok {
			metrics.CPU = cpuCollector.GetMetrics()
		}
	}
	
	// 内存指标
	if collector, exists := pa.collectors["memory"]; exists {
		if memCollector, ok := collector.(*MemoryCollector); ok {
			metrics.Memory = memCollector.GetMetrics()
		}
	}
	
	// 磁盘指标
	if collector, exists := pa.collectors["disk"]; exists {
		if diskCollector, ok := collector.(*DiskCollector); ok {
			metrics.Disk = diskCollector.GetMetrics()
		}
	}
	
	// 网络指标
	if collector, exists := pa.collectors["network"]; exists {
		if netCollector, ok := collector.(*NetworkCollector); ok {
			metrics.Network = netCollector.GetMetrics()
		}
	}
	
	// 进程指标
	if collector, exists := pa.collectors["process"]; exists {
		if procCollector, ok := collector.(*ProcessCollector); ok {
			metrics.Process = procCollector.GetMetrics()
		}
	}
	
	return metrics, nil
}

// AnalyzePerformance 分析性能
func (pa *PerformanceAnalyzer) AnalyzePerformance(timeRange TimeRange) (*PerformanceReport, error) {
	report := &PerformanceReport{
		TimeRange: timeRange,
		Timestamp: time.Now(),
		Anomalies: make([]*Anomaly, 0),
		Trends:    make([]*Trend, 0),
		Alerts:    make([]*Alert, 0),
	}
	
	// 运行所有分析器
	pa.mutex.RLock()
	analyzers := make(map[string]Analyzer)
	for name, analyzer := range pa.analyzers {
		analyzers[name] = analyzer
	}
	pa.mutex.RUnlock()
	
	for name, analyzer := range analyzers {
		result, err := analyzer.Analyze(timeRange)
		if err != nil {
			continue
		}
		
		// 合并结果
		report.Anomalies = append(report.Anomalies, result.Anomalies...)
		report.Trends = append(report.Trends, result.Trends...)
		report.Alerts = append(report.Alerts, result.Alerts...)
		
		// 更新统计
		pa.updateAnalyzerStats(name, result)
	}
	
	// 计算总体评分
	report.OverallScore = pa.calculateOverallScore(report)
	
	return report, nil
}

// DetectAnomalies 检测异常
func (pa *PerformanceAnalyzer) DetectAnomalies(metricName string, timeRange TimeRange) ([]*Anomaly, error) {
	// 获取历史数据
	query := interfaces.MetricQuery{
		MetricName: metricName,
		TimeRange:  interfaces.TimeRange{From: timeRange.From, To: timeRange.To},
	}
	
	result, err := pa.storage.QueryMetrics(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	
	// 使用统计方法检测异常
	anomalies := make([]*Anomaly, 0)
	
	if len(result.Values) < 10 {
		return anomalies, nil // 数据不足
	}
	
	// 计算统计信息
	mean, stddev := pa.calculateStats(result.Values)
	threshold := pa.config.AnomalyThreshold * stddev
	
	for i, value := range result.Values {
		if abs(value.Value-mean) > threshold {
			anomaly := &Anomaly{
				MetricName:  metricName,
				Timestamp:   value.Timestamp,
				Value:       value.Value,
				Expected:    mean,
				Deviation:   abs(value.Value - mean),
				Severity:    pa.calculateSeverity(abs(value.Value-mean), threshold),
				Description: fmt.Sprintf("Metric %s deviated by %.2f from expected %.2f", metricName, abs(value.Value-mean), mean),
			}
			anomalies = append(anomalies, anomaly)
		}
	}
	
	return anomalies, nil
}

// GetTrends 获取趋势
func (pa *PerformanceAnalyzer) GetTrends(metricName string, timeRange TimeRange) ([]*Trend, error) {
	// 获取历史数据
	query := interfaces.MetricQuery{
		MetricName: metricName,
		TimeRange:  interfaces.TimeRange{From: timeRange.From, To: timeRange.To},
	}
	
	result, err := pa.storage.QueryMetrics(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	
	trends := make([]*Trend, 0)
	
	if len(result.Values) < 5 {
		return trends, nil // 数据不足
	}
	
	// 计算趋势
	slope := pa.calculateSlope(result.Values)
	direction := TrendDirectionStable
	
	if slope > 0.1 {
		direction = TrendDirectionIncreasing
	} else if slope < -0.1 {
		direction = TrendDirectionDecreasing
	}
	
	trend := &Trend{
		MetricName:  metricName,
		TimeRange:   timeRange,
		Direction:   direction,
		Slope:       slope,
		Confidence:  pa.calculateTrendConfidence(result.Values),
		Description: fmt.Sprintf("Metric %s shows %s trend with slope %.4f", metricName, direction, slope),
	}
	
	trends = append(trends, trend)
	
	return trends, nil
}

// initCollectors 初始化收集器
func (pa *PerformanceAnalyzer) initCollectors() error {
	for name, config := range pa.config.Collectors {
		collector, err := CreateMetricCollector(config)
		if err != nil {
			return fmt.Errorf("failed to create collector %s: %w", name, err)
		}
		
		pa.collectors[name] = collector
		pa.stats.CollectorStats[name] = &CollectorStats{}
	}
	
	return nil
}

// initAnalyzers 初始化分析器
func (pa *PerformanceAnalyzer) initAnalyzers() error {
	for name, settings := range pa.config.Analyzers {
		analyzer, err := CreateAnalyzer(settings)
		if err != nil {
			return fmt.Errorf("failed to create analyzer %s: %w", name, err)
		}
		
		pa.analyzers[name] = analyzer
		pa.stats.AnalyzerStats[name] = &AnalysisStats{}
	}
	
	return nil
}

// collectionLoop 收集循环
func (pa *PerformanceAnalyzer) collectionLoop() {
	defer pa.wg.Done()
	
	ticker := time.NewTicker(pa.config.CollectionInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-pa.ctx.Done():
			return
		case <-ticker.C:
			pa.collectMetrics()
		}
	}
}

// analysisLoop 分析循环
func (pa *PerformanceAnalyzer) analysisLoop() {
	defer pa.wg.Done()
	
	ticker := time.NewTicker(pa.config.AnalysisInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-pa.ctx.Done():
			return
		case <-ticker.C:
			pa.runAnalysis()
		}
	}
}

// cleanupLoop 清理循环
func (pa *PerformanceAnalyzer) cleanupLoop() {
	defer pa.wg.Done()
	
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-pa.ctx.Done():
			return
		case <-ticker.C:
			pa.cleanup()
		}
	}
}

// collectMetrics 收集指标
func (pa *PerformanceAnalyzer) collectMetrics() {
	start := time.Now()
	
	pa.mutex.RLock()
	collectors := make(map[string]MetricCollector)
	for name, collector := range pa.collectors {
		collectors[name] = collector
	}
	pa.mutex.RUnlock()
	
	var totalMetrics int64
	
	for name, collector := range collectors {
		metrics, err := collector.Collect()
		if err != nil {
			continue
		}
		
		// 存储指标
		if pa.config.StorageEnabled && pa.storage != nil {
			for _, metric := range metrics {
				pa.storage.StoreMetric(metric)
			}
		}
		
		totalMetrics += int64(len(metrics))
		
		// 更新收集器统计
		pa.updateCollectorStats(name, int64(len(metrics)))
	}
	
	// 更新总体统计
	pa.mutex.Lock()
	pa.stats.CollectedMetrics += totalMetrics
	pa.stats.LastCollection = time.Now()
	pa.stats.CollectionTime = time.Since(start)
	pa.mutex.Unlock()
}

// runAnalysis 运行分析
func (pa *PerformanceAnalyzer) runAnalysis() {
	start := time.Now()
	
	timeRange := TimeRange{
		From: time.Now().Add(-pa.config.TrendWindow),
		To:   time.Now(),
	}
	
	report, err := pa.AnalyzePerformance(timeRange)
	if err != nil {
		return
	}
	
	// 更新统计
	pa.mutex.Lock()
	pa.stats.AnalyzedMetrics++
	pa.stats.DetectedAnomalies += int64(len(report.Anomalies))
	pa.stats.GeneratedAlerts += int64(len(report.Alerts))
	pa.stats.LastAnalysis = time.Now()
	pa.stats.AnalysisTime = time.Since(start)
	pa.mutex.Unlock()
}

// cleanup 清理过期数据
func (pa *PerformanceAnalyzer) cleanup() {
	if !pa.config.StorageEnabled || pa.storage == nil {
		return
	}
	
	// 删除过期指标
	cutoff := time.Now().Add(-pa.config.MetricsRetention)
	// 这里需要存储接口支持删除操作
	// pa.storage.DeleteMetricsBefore(cutoff)
	_ = cutoff
}

// updateCollectorStats 更新收集器统计
func (pa *PerformanceAnalyzer) updateCollectorStats(name string, count int64) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if stats, exists := pa.stats.CollectorStats[name]; exists {
		stats.CollectedMetrics += count
		stats.LastCollection = time.Now()
	}
}

// updateAnalyzerStats 更新分析器统计
func (pa *PerformanceAnalyzer) updateAnalyzerStats(name string, result *AnalysisResult) {
	pa.mutex.Lock()
	defer pa.mutex.Unlock()
	
	if stats, exists := pa.stats.AnalyzerStats[name]; exists {
		stats.AnalyzedMetrics++
		stats.DetectedAnomalies += int64(len(result.Anomalies))
		stats.GeneratedAlerts += int64(len(result.Alerts))
		stats.LastAnalysis = time.Now()
	}
}

// calculateStats 计算统计信息
func (pa *PerformanceAnalyzer) calculateStats(values []interfaces.MetricValue) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	// 计算平均值
	var sum float64
	for _, value := range values {
		sum += value.Value
	}
	mean = sum / float64(len(values))
	
	// 计算标准差
	var variance float64
	for _, value := range values {
		variance += (value.Value - mean) * (value.Value - mean)
	}
	variance /= float64(len(values))
	stddev = sqrt(variance)
	
	return mean, stddev
}

// calculateSlope 计算斜率
func (pa *PerformanceAnalyzer) calculateSlope(values []interfaces.MetricValue) float64 {
	if len(values) < 2 {
		return 0
	}
	
	n := float64(len(values))
	var sumX, sumY, sumXY, sumX2 float64
	
	for i, value := range values {
		x := float64(i)
		y := value.Value
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// 线性回归斜率
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

// calculateSeverity 计算严重程度
func (pa *PerformanceAnalyzer) calculateSeverity(deviation, threshold float64) AnomalySeverity {
	ratio := deviation / threshold
	
	if ratio >= 3.0 {
		return AnomalySeverityCritical
	} else if ratio >= 2.0 {
		return AnomalySeverityHigh
	} else if ratio >= 1.5 {
		return AnomalySeverityMedium
	} else {
		return AnomalySeverityLow
	}
}

// calculateTrendConfidence 计算趋势置信度
func (pa *PerformanceAnalyzer) calculateTrendConfidence(values []interfaces.MetricValue) float64 {
	if len(values) < 3 {
		return 0.0
	}
	
	// 简单的置信度计算，基于数据点数量和变化一致性
	consistency := pa.calculateConsistency(values)
	dataPoints := float64(len(values))
	
	confidence := (consistency * 0.7) + (min(dataPoints/100.0, 1.0) * 0.3)
	return min(confidence, 1.0)
}

// calculateConsistency 计算一致性
func (pa *PerformanceAnalyzer) calculateConsistency(values []interfaces.MetricValue) float64 {
	if len(values) < 3 {
		return 0.0
	}
	
	// 计算相邻点的变化方向一致性
	var positiveChanges, negativeChanges int
	
	for i := 1; i < len(values); i++ {
		change := values[i].Value - values[i-1].Value
		if change > 0 {
			positiveChanges++
		} else if change < 0 {
			negativeChanges++
		}
	}
	
	total := positiveChanges + negativeChanges
	if total == 0 {
		return 1.0 // 完全稳定
	}
	
	// 返回主导方向的比例
	return float64(max(positiveChanges, negativeChanges)) / float64(total)
}

// calculateOverallScore 计算总体评分
func (pa *PerformanceAnalyzer) calculateOverallScore(report *PerformanceReport) float64 {
	score := 100.0
	
	// 根据异常数量扣分
	for _, anomaly := range report.Anomalies {
		switch anomaly.Severity {
		case AnomalySeverityCritical:
			score -= 20
		case AnomalySeverityHigh:
			score -= 10
		case AnomalySeverityMedium:
			score -= 5
		case AnomalySeverityLow:
			score -= 2
		}
	}
	
	// 根据告警数量扣分
	score -= float64(len(report.Alerts)) * 5
	
	// 确保分数在0-100范围内
	if score < 0 {
		score = 0
	}
	
	return score
}

// 辅助函数
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func sqrt(x float64) float64 {
	// 简单的平方根实现，实际项目中应使用math.Sqrt
	if x == 0 {
		return 0
	}
	
	guess := x / 2
	for i := 0; i < 10; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}