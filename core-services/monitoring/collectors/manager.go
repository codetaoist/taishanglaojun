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

// CollectorManager 
type CollectorManager struct {
	collectors map[string]interfaces.MetricCollector
	config     *CollectorManagerConfig
	
	// ?
	db *sql.DB
	
	// Redis?
	redisClient redis.UniversalClient
	
	// 
	ctx    context.Context
	cancel context.CancelFunc
	
	// ?
	mutex sync.RWMutex
	
	// ?
	running bool
	
	// 
	metricsChan chan []models.Metric
	
	// 
	errorChan chan error
}

// CollectorManagerConfig 
type CollectorManagerConfig struct {
	// 
	GlobalInterval time.Duration     `yaml:"global_interval"`
	GlobalLabels   map[string]string `yaml:"global_labels"`
	
	// ?
	System      SystemCollectorConfig      `yaml:"system"`
	Application ApplicationCollectorConfig `yaml:"application"`
	Database    DatabaseCollectorConfig    `yaml:"database"`
	Business    BusinessCollectorConfig    `yaml:"business"`
	Redis       RedisCollectorConfig       `yaml:"redis"`
	
	// 
	MetricsBufferSize int           `yaml:"metrics_buffer_size"`
	ErrorBufferSize   int           `yaml:"error_buffer_size"`
	FlushInterval     time.Duration `yaml:"flush_interval"`
}

// CollectorStats ?
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

// ManagerStats ?
type ManagerStats struct {
	Running          bool                       `json:"running"`
	CollectorCount   int                        `json:"collector_count"`
	EnabledCount     int                        `json:"enabled_count"`
	TotalMetrics     uint64                     `json:"total_metrics"`
	TotalErrors      uint64                     `json:"total_errors"`
	CollectorStats   map[string]CollectorStats  `json:"collector_stats"`
	LastUpdateTime   time.Time                  `json:"last_update_time"`
}

// NewCollectorManager 
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

// Initialize 
func (m *CollectorManager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 
	m.mergeGlobalLabels()
	
	// 
	if m.config.System.Enabled {
		systemCollector := NewSystemCollector(m.config.System)
		m.collectors["system"] = systemCollector
	}
	
	// 
	if m.config.Application.Enabled {
		appCollector := NewApplicationCollector(m.config.Application)
		m.collectors["application"] = appCollector
	}
	
	// ?
	if m.config.Database.Enabled && m.db != nil {
		dbCollector := NewDatabaseCollector(m.config.Database, m.db)
		m.collectors["database"] = dbCollector
	}
	
	// 
	if m.config.Business.Enabled && m.db != nil {
		businessCollector := NewBusinessCollector(m.config.Business, m.db)
		m.collectors["business"] = businessCollector
	}
	
	// Redis?
	if m.config.Redis.Enabled && m.redisClient != nil {
		redisCollector := NewRedisCollector(m.config.Redis, m.redisClient)
		m.collectors["redis"] = redisCollector
	}
	
	return nil
}

// mergeGlobalLabels 
func (m *CollectorManager) mergeGlobalLabels() {
	if len(m.config.GlobalLabels) == 0 {
		return
	}
	
	// 
	if m.config.System.Labels == nil {
		m.config.System.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.System.Labels[k]; !exists {
			m.config.System.Labels[k] = v
		}
	}
	
	// 
	if m.config.Application.Labels == nil {
		m.config.Application.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Application.Labels[k]; !exists {
			m.config.Application.Labels[k] = v
		}
	}
	
	// ?
	if m.config.Database.Labels == nil {
		m.config.Database.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Database.Labels[k]; !exists {
			m.config.Database.Labels[k] = v
		}
	}
	
	// 
	if m.config.Business.Labels == nil {
		m.config.Business.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Business.Labels[k]; !exists {
			m.config.Business.Labels[k] = v
		}
	}
	
	// Redis?
	if m.config.Redis.Labels == nil {
		m.config.Redis.Labels = make(map[string]string)
	}
	for k, v := range m.config.GlobalLabels {
		if _, exists := m.config.Redis.Labels[k]; !exists {
			m.config.Redis.Labels[k] = v
		}
	}
}

// Start 
func (m *CollectorManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("collector manager is already running")
	}
	
	// 
	for name, collector := range m.collectors {
		if collector.IsEnabled() {
			go func(name string, collector interfaces.MetricCollector) {
				if err := collector.Start(m.ctx); err != nil {
					select {
					case m.errorChan <- fmt.Errorf("collector %s error: %w", name, err):
					default:
						// ?
					}
				}
			}(name, collector)
		}
	}
	
	// 
	go m.collectMetrics()
	
	// 
	go m.handleErrors()
	
	m.running = true
	return nil
}

// Stop 
func (m *CollectorManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.running {
		return nil
	}
	
	// 
	m.cancel()
	
	// 
	for _, collector := range m.collectors {
		if err := collector.Stop(); err != nil {
			fmt.Printf("Error stopping collector: %v\n", err)
		}
	}
	
	// 
	close(m.metricsChan)
	close(m.errorChan)
	
	m.running = false
	return nil
}

// collectMetrics 
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

// flushMetrics 
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
				// ?
			}
			continue
		}
		
		if len(metrics) > 0 {
			select {
			case m.metricsChan <- metrics:
			default:
				// ?
				fmt.Printf("Metrics channel full, dropping %d metrics from %s\n", len(metrics), name)
			}
		}
	}
}

// handleErrors 
func (m *CollectorManager) handleErrors() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case err := <-m.errorChan:
			if err != nil {
				fmt.Printf("Collector error: %v\n", err)
				// 緢澯
			}
		}
	}
}

// GetMetricsChannel 
func (m *CollectorManager) GetMetricsChannel() <-chan []models.Metric {
	return m.metricsChan
}

// GetErrorChannel 
func (m *CollectorManager) GetErrorChannel() <-chan error {
	return m.errorChan
}

// AddCollector ?
func (m *CollectorManager) AddCollector(name string, collector interfaces.MetricCollector) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.collectors[name]; exists {
		return fmt.Errorf("collector %s already exists", name)
	}
	
	m.collectors[name] = collector
	
	// 
	if m.running && collector.IsEnabled() {
		go func() {
			if err := collector.Start(m.ctx); err != nil {
				select {
				case m.errorChan <- fmt.Errorf("collector %s error: %w", name, err):
				default:
					// ?
				}
			}
		}()
	}
	
	return nil
}

// RemoveCollector ?
func (m *CollectorManager) RemoveCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// ?
	if err := collector.Stop(); err != nil {
		return fmt.Errorf("failed to stop collector %s: %w", name, err)
	}
	
	delete(m.collectors, name)
	return nil
}

// GetCollector ?
func (m *CollectorManager) GetCollector(name string) (interfaces.MetricCollector, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return nil, fmt.Errorf("collector %s not found", name)
	}
	
	return collector, nil
}

// ListCollectors 
func (m *CollectorManager) ListCollectors() map[string]interfaces.MetricCollector {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]interfaces.MetricCollector)
	for k, v := range m.collectors {
		result[k] = v
	}
	
	return result
}

// EnableCollector ?
func (m *CollectorManager) EnableCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// 
	// Enable?
	fmt.Printf("Enabling collector %s\n", name)
	
	// ?
	if m.running {
		go func() {
			if err := collector.Start(m.ctx); err != nil {
				select {
				case m.errorChan <- fmt.Errorf("collector %s error: %w", name, err):
				default:
					// ?
				}
			}
		}()
	}
	
	return nil
}

// DisableCollector ?
func (m *CollectorManager) DisableCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}
	
	// ?
	if err := collector.Stop(); err != nil {
		return fmt.Errorf("failed to stop collector %s: %w", name, err)
	}
	
	return nil
}

// Health ?
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

// GetStats 
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
			
			// 齡?
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

// GetCollectorStats ?
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
		// 齡?
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

// IsRunning 
func (m *CollectorManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// GetConfig 
func (m *CollectorManager) GetConfig() *CollectorManagerConfig {
	return m.config
}

// UpdateConfig 
func (m *CollectorManager) UpdateConfig(config *CollectorManagerConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("cannot update config while running")
	}
	
	m.config = config
	return nil
}

// Restart ?
func (m *CollectorManager) Restart() error {
	if err := m.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}
	
	// ?
	if err := m.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	
	if err := m.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}
	
	return nil
}

// CollectOnce ?
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

