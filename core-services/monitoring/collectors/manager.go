package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// CollectorManager 收集器管理器
type CollectorManager struct {
	collectors map[string]interfaces.MetricCollector
	config     *CollectorManagerConfig
	
	// 数据库连�?
	db *sql.DB
	
	// Redis客户�?
	redisClient redis.UniversalClient
	
	// 上下文和取消函数
	ctx    context.Context
	cancel context.CancelFunc
	
	// 同步�?
	mutex sync.RWMutex
	
	// 运行状�?
	running bool
	
	// 指标通道
	metricsChan chan []models.Metric
	
	// 错误通道
	errorChan chan error
}

// CollectorManagerConfig 收集器管理器配置
type CollectorManagerConfig struct {
	// 全局配置
	GlobalInterval time.Duration     `yaml:"global_interval"`
	GlobalLabels   map[string]string `yaml:"global_labels"`
	
	// 收集器配�?
	System      SystemCollectorConfig      `yaml:"system"`
	Application ApplicationCollectorConfig `yaml:"application"`
	Database    DatabaseCollectorConfig    `yaml:"database"`
	Business    BusinessCollectorConfig    `yaml:"business"`
	Redis       RedisCollectorConfig       `yaml:"redis"`
	
	// 输出配置
	MetricsBufferSize int           `yaml:"metrics_buffer_size"`
	ErrorBufferSize   int           `yaml:"error_buffer_size"`
	FlushInterval     time.Duration `yaml:"flush_interval"`
}

// CollectorStats 收集器统计信�?
type CollectorStats struct {
	Name            string        `json:"name"`
	Category        string        `json:"category"`
	Enabled         bool          `json:"enabled"`
	Interval        time.Duration `json:"interval"`
	LastCollectTime time.Time     `json:"last_collect_time"`
	CollectCount    uint64        `json:"collect_count"`
	ErrorCount      uint64        `json:"error_count"`
	LastError       string        `json:"last_error"`
	HealthStatus    string        `json:"health_status"`
	MetricsCount    uint64        `json:"metrics_count"`
}

// ManagerStats 管理器统计信�?
type ManagerStats struct {
	Running          bool                       `json:"running"`
	CollectorCount   int                        `json:"collector_count"`
	EnabledCount     int                        `json:"enabled_count"`
	TotalMetrics     uint64                     `json:"total_metrics"`
	TotalErrors      uint64                     `json:"total_errors"`
	CollectorStats   map[string]CollectorStats  `json:"collector_stats"`
	LastUpdateTime   time.Time                  `json:"last_update_time"`
}

// NewCollectorManager 创建收集器管理器
func NewCollectorManager(config *CollectorManagerConfig, db *sql.DB, redisClient redis.UniversalClient) *CollectorManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &CollectorManager{
		collectors:  make(map[string]interfaces.MetricCollector),
		config:      config,
		db:          db,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
		metricsChan: make(chan []models.Metric, config.MetricsBufferSize),
		errorChan:   make(chan error, config.ErrorBufferSize),
	}
}

// Initialize 初始化收集器
func (m *CollectorManager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 合并全局标签到各个收集器配置
	m.mergeGlobalLabels()
	
	// 初始化系统收集器
	if m.config.System.Enabled {
		systemCollector := NewSystemCollector(m.config.System)
		m.collectors["system"] = systemCollector
	}
	
	// 初始化应用收集器
	if m.config.Application.Enabled {
		appCollector := NewApplicationCollector(m.config.Application)
		m.collectors["application"] = appCollector
	}
	
	// 初始化数据库收集�?
	if m.config.Database.Enabled && m.db != nil {
		dbCollector := NewDatabaseCollector(m.config.Database, m.db)
		m.collectors["database"] = dbCollector
	}
	
	// 初始化业务收集器
	if m.config.Business.Enabled && m.db != nil {
		businessCollector := NewBusinessCollector(m.config.Business, m.db)
		m.collectors["business"] = businessCollector
	}
	
	// 初始化Redis收集�?
	if m.config.Redis.Enabled && m.redisClient != nil {
		redisCollector := NewRedisCollector(m.config.Redis, m.redisClient)
		m.collectors["redis"] = redisCollector
	}
	
	return nil
}

// mergeGlobalLabels 合并全局标签
func (m *CollectorManager) mergeGlobalLabels() {
	if len(m.config.GlobalLabels) == 0 {
		return
	}
	
	// 合并到系统收集器
	if m.config.System.Labels == nil {
		m.config.System.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.System.Labels[k]; !exists {
			m.config.System.Labels[k] = v
		}
	}
	
	// 合并到应用收集器
	if m.config.Application.Labels == nil {
		m.config.Application.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Application.Labels[k]; !exists {
			m.config.Application.Labels[k] = v
		}
	}
	
	// 合并到数据库收集�?
	if m.config.Database.Labels == nil {
		m.config.Database.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Database.Labels[k]; !exists {
			m.config.Database.Labels[k] = v
		}
	}
	
	// 合并到业务收集器
	if m.config.Business.Labels == nil {
		m.config.Business.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Business.Labels[k]; !exists {
			m.config.Business.Labels[k] = v
		}
	}
	
	// 合并到Redis收集�?
	if m.config.Redis.Labels == nil {
		m.config.Redis.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Redis.Labels[k]; !exists {
			m.config.Redis.Labels[k] = v
		}
	}
}

// Start 启动所有收集器
func (m *CollectorManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("collector manager is already running")
	}
	
	// 启动所有收集器
	for name, collector := range m.collectors {
		if collector.IsEnabled() {
			go func(name string, collector interfaces.MetricCollector) {
				if err := collector.Start(m.ctx); err != nil {
					select {
					case m.errorChan <- fmt.Errorf("collector %s error: %w", name, err):
					default:
						// 错误通道已满，丢弃错�?
					}
				}
			}(name, collector)
		}
	}
	
	// 启动指标收集协程
	go m.collectMetrics()
	
	// 启动错误处理协程
	go m.handleErrors()
	
	m.running = true
	return nil
}

// Stop 停止所有收集器
func (m *CollectorManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.running {
		return nil
	}
	
	// 取消上下文，停止所有收集器
	m.cancel()
	
	// 停止所有收集器
	for _, collector := range m.collectors {
		if err := collector.Stop(); err != nil {
			fmt.Printf("Error stopping collector: %v\n", err)
		}
	}
	
	// 关闭通道
	close(m.metricsChan)
	close(m.errorChan)
	
	m.running = false
	return nil
}

// collectMetrics 收集指标协程
func (m *CollectorManager) collectMetrics() {
	ticker := time.NewTicker(m.config.FlushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.flushMetrics()
		}
	}
}

// flushMetrics 刷新指标
func (m *CollectorManager) flushMetrics() {
	m.mutex.RLock()
	collectors := make(map[string]interfaces.MetricCollector)
	for k, v := range m.collectors {
		collectors[k] = v
	}
	m.mutex.RUnlock()
	
	for name, collector := range collectors {
		if !collector.IsEnabled() {
			continue
		}
		
		metrics, err := collector.Collect(m.ctx)
		if err != nil {
			select {
			case m.errorChan <- fmt.Errorf("collector %s collect error: %w", name, err):
			default:
				// 错误通道已满，丢弃错�?
			}
			continue
		}
		
		if len(metrics) > 0 {
			select {
			case m.metricsChan <- metrics:
			default:
				// 指标通道已满，丢弃指�?
				fmt.Printf("Metrics channel full, dropping %d metrics from %s\n", len(metrics), name)
			}
		}
	}
}

// handleErrors 处理错误协程
func (m *CollectorManager) handleErrors() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case err := <-m.errorChan:
			if err != nil {
				fmt.Printf("Collector error: %v\n", err)
				// 这里可以添加错误处理逻辑，如发送告警等
			}
		}
	}
}

// GetMetricsChannel 获取指标通道
func (m *CollectorManager) GetMetricsChannel() <-chan []models.Metric {
	return m.metricsChan
}

// GetErrorChannel 获取错误通道
func (m *CollectorManager) GetErrorChannel() <-chan error {
	return m.errorChan
}

// AddCollector 添加收集�?
func (m *CollectorManager) AddCollector(name string, collector interfaces.MetricCollector) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.collectors[name]; exists {
		return fmt.Errorf("collector %s already exists", name)
	}
	
	m.collectors[name] = collector
	
	// 如果管理器正在运行，启动新收集器
	if m.running && collector.IsEnabled() {
		go func() {
			if err := collector.Start(m.ctx); err != nil {
				select {
				case m.errorChan <- fmt.Errorf("collector %s error: %w", name, err):
				default:
					// 错误通道已满，丢弃错�?
				}
			}
		}()
	}
	
	return nil
}

// RemoveCollector 移除收集�?
func (m *CollectorManager) RemoveCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// 停止收集�?
	if err := collector.Stop(); err != nil {
		return fmt.Errorf("failed to stop collector %s: %w", name, err)
	}
	
	delete(m.collectors, name)
	return nil
}

// GetCollector 获取收集�?
func (m *CollectorManager) GetCollector(name string) (interfaces.MetricCollector, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector %s not found", name)
	}
	
	return collector, nil
}

// ListCollectors 列出所有收集器
func (m *CollectorManager) ListCollectors() map[string]interfaces.MetricCollector {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]interfaces.MetricCollector)
	for k, v := range m.collectors {
		result[k] = v
	}
	
	return result
}

// EnableCollector 启用收集�?
func (m *CollectorManager) EnableCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// 这里需要根据具体的收集器实现来启用
	// 由于接口中没有Enable方法，这里只是示�?
	fmt.Printf("Enabling collector %s\n", name)
	
	// 如果管理器正在运行，启动收集�?
	if m.running {
		go func() {
			if err := collector.Start(m.ctx); err != nil {
				select {
				case m.errorChan <- fmt.Errorf("collector %s error: %w", name, err):
				default:
					// 错误通道已满，丢弃错�?
				}
			}
		}()
	}
	
	return nil
}

// DisableCollector 禁用收集�?
func (m *CollectorManager) DisableCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// 停止收集�?
	if err := collector.Stop(); err != nil {
		return fmt.Errorf("failed to stop collector %s: %w", name, err)
	}
	
	return nil
}

// Health 健康检�?
func (m *CollectorManager) Health() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var errors []string
	
	for name, collector := range m.collectors {
		if collector.IsEnabled() {
			if err := collector.Health(); err != nil {
				errors = append(errors, fmt.Sprintf("collector %s: %v", name, err))
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("health check failed: %v", errors)
	}
	
	return nil
}

// GetStats 获取统计信息
func (m *CollectorManager) GetStats() *ManagerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	stats := &ManagerStats{
		Running:        m.running,
		CollectorCount: len(m.collectors),
		EnabledCount:   0,
		TotalMetrics:   0,
		TotalErrors:    0,
		CollectorStats: make(map[string]CollectorStats),
		LastUpdateTime: time.Now(),
	}
	
	for name, collector := range m.collectors {
		collectorStats := CollectorStats{
			Name:         name,
			Category:     string(collector.GetCategory()),
			Enabled:      collector.IsEnabled(),
			Interval:     collector.GetInterval(),
			HealthStatus: "unknown",
		}
		
		if collector.IsEnabled() {
			stats.EnabledCount++
			
			// 检查健康状�?
			if err := collector.Health(); err != nil {
				collectorStats.HealthStatus = "unhealthy"
				collectorStats.LastError = err.Error()
			} else {
				collectorStats.HealthStatus = "healthy"
			}
		} else {
			collectorStats.HealthStatus = "disabled"
		}
		
		stats.CollectorStats[name] = collectorStats
	}
	
	return stats
}

// GetCollectorStats 获取特定收集器统计信�?
func (m *CollectorManager) GetCollectorStats(name string) (*CollectorStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector %s not found", name)
	}
	
	stats := &CollectorStats{
		Name:         name,
		Category:     string(collector.GetCategory()),
		Enabled:      collector.IsEnabled(),
		Interval:     collector.GetInterval(),
		HealthStatus: "unknown",
	}
	
	if collector.IsEnabled() {
		// 检查健康状�?
		if err := collector.Health(); err != nil {
			stats.HealthStatus = "unhealthy"
			stats.LastError = err.Error()
		} else {
			stats.HealthStatus = "healthy"
		}
	} else {
		stats.HealthStatus = "disabled"
	}
	
	return stats, nil
}

// IsRunning 检查是否运行中
func (m *CollectorManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// GetConfig 获取配置
func (m *CollectorManager) GetConfig() *CollectorManagerConfig {
	return m.config
}

// UpdateConfig 更新配置
func (m *CollectorManager) UpdateConfig(config *CollectorManagerConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("cannot update config while running")
	}
	
	m.config = config
	return nil
}

// Restart 重启管理�?
func (m *CollectorManager) Restart() error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}
	
	// 重新初始�?
	if err := m.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	
	if err := m.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}
	
	return nil
}

// CollectOnce 执行一次收�?
func (m *CollectorManager) CollectOnce() (map[string][]models.Metric, error) {
	m.mutex.RLock()
	collectors := make(map[string]interfaces.MetricCollector)
	for k, v := range m.collectors {
		collectors[k] = v
	}
	m.mutex.RUnlock()
	
	result := make(map[string][]models.Metric)
	var errors []string
	
	for name, collector := range collectors {
		if !collector.IsEnabled() {
			continue
		}
		
		metrics, err := collector.Collect(m.ctx)
		if err != nil {
			errors = append(errors, fmt.Sprintf("collector %s: %v", name, err))
			continue
		}
		
		result[name] = metrics
	}
	
	if len(errors) > 0 {
		return result, fmt.Errorf("collect errors: %v", errors)
	}
	
	return result, nil
}
