package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// StorageManager 洢?
type StorageManager struct {
	storages map[string]interfaces.MetricStorage
	config   *StorageManagerConfig
	
	// 洢
	primaryStorage interfaces.MetricStorage
	
	// ?
	mutex sync.RWMutex
	
	// ?
	running bool
	
	// ?
	ctx    context.Context
	cancel context.CancelFunc
}

// StorageManagerConfig 洢?
type StorageManagerConfig struct {
	// 洢?
	Primary string `yaml:"primary"`
	
	// 洢
	Prometheus *PrometheusConfig `yaml:"prometheus"`
	InfluxDB   *InfluxDBConfig   `yaml:"influxdb"`
	
	// 
	Replication *ReplicationConfig `yaml:"replication"`
	
	// 
	Sharding *ShardingConfig `yaml:"sharding"`
	
	// 
	Cache *CacheConfig `yaml:"cache"`
	
	// 
	Backup *BackupConfig `yaml:"backup"`
}

// ReplicationConfig 
type ReplicationConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Targets   []string `yaml:"targets"`
	Async     bool     `yaml:"async"`
	BatchSize int      `yaml:"batch_size"`
}

// ShardingConfig 
type ShardingConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Strategy   string            `yaml:"strategy"` // hash, range, time
	ShardCount int               `yaml:"shard_count"`
	ShardKey   string            `yaml:"shard_key"`
	Shards     map[string]string `yaml:"shards"`
}

// CacheConfig 
type CacheConfig struct {
	Enabled    bool          `yaml:"enabled"`
	TTL        time.Duration `yaml:"ttl"`
	MaxSize    int           `yaml:"max_size"`
	Strategy   string        `yaml:"strategy"` // lru, lfu, fifo
}

// BackupConfig 
type BackupConfig struct {
	Enabled   bool          `yaml:"enabled"`
	Interval  time.Duration `yaml:"interval"`
	Retention time.Duration `yaml:"retention"`
	Location  string        `yaml:"location"`
	Format    string        `yaml:"format"` // json, csv, parquet
}

// StorageStats 洢
type StorageStats struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Status         string            `json:"status"`
	MetricsCount   uint64            `json:"metrics_count"`
	QueriesCount   uint64            `json:"queries_count"`
	ErrorsCount    uint64            `json:"errors_count"`
	LastWriteTime  time.Time         `json:"last_write_time"`
	LastQueryTime  time.Time         `json:"last_query_time"`
	LastError      string            `json:"last_error"`
	CustomStats    map[string]interface{} `json:"custom_stats"`
}

// ManagerStats ?
type ManagerStats struct {
	Running        bool                     `json:"running"`
	PrimaryStorage string                   `json:"primary_storage"`
	StorageCount   int                      `json:"storage_count"`
	TotalMetrics   uint64                   `json:"total_metrics"`
	TotalQueries   uint64                   `json:"total_queries"`
	TotalErrors    uint64                   `json:"total_errors"`
	StorageStats   map[string]StorageStats  `json:"storage_stats"`
	LastUpdateTime time.Time                `json:"last_update_time"`
}

// NewStorageManager 洢?
func NewStorageManager(config *StorageManagerConfig) (*StorageManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &StorageManager{
		storages: make(map[string]interfaces.MetricStorage),
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		running:  false,
	}
	
	return manager, nil
}

// Initialize ?
func (m *StorageManager) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Prometheus洢
	if m.config.Prometheus != nil {
		prometheus, err := NewPrometheusStorage(m.config.Prometheus)
		if err != nil {
			return fmt.Errorf("failed to initialize prometheus storage: %w", err)
		}
		m.storages["prometheus"] = prometheus
	}
	
	// InfluxDB洢
	if m.config.InfluxDB != nil {
		influxdb, err := NewInfluxDBStorage(m.config.InfluxDB)
		if err != nil {
			return fmt.Errorf("failed to initialize influxdb storage: %w", err)
		}
		m.storages["influxdb"] = influxdb
	}
	
	// ?
	if m.config.Primary != "" {
		if storage, exists := m.storages[m.config.Primary]; exists {
			m.primaryStorage = storage
		} else {
			return fmt.Errorf("primary storage %s not found", m.config.Primary)
		}
	} else if len(m.storages) > 0 {
		// 洢洢
		for _, storage := range m.storages {
			m.primaryStorage = storage
			break
		}
	}
	
	return nil
}

// Start 洢?
func (m *StorageManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("storage manager is already running")
	}
	
	// ?
	for name, storage := range m.storages {
		if err := storage.Health(m.ctx); err != nil {
			fmt.Printf("Storage %s health check failed: %v\n", name, err)
		}
	}
	
	m.running = true
	return nil
}

// Stop 洢?
func (m *StorageManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.running {
		return nil
	}
	
	// ?
	m.cancel()
	
	// ?
	for name, storage := range m.storages {
		if closer, ok := storage.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				fmt.Printf("Error closing storage %s: %v\n", name, err)
			}
		}
	}
	
	m.running = false
	return nil
}

// Store 洢
func (m *StorageManager) Store(ctx context.Context, metrics []models.Metric) error {
	if m.primaryStorage == nil {
		return fmt.Errorf("no primary storage configured")
	}
	
	// 洢洢
	if err := m.primaryStorage.Store(ctx, metrics); err != nil {
		return fmt.Errorf("failed to store to primary storage: %w", err)
	}
	
	// ?
	if m.config.Replication != nil && m.config.Replication.Enabled {
		m.replicateMetrics(ctx, metrics)
	}
	
	return nil
}

// replicateMetrics ?
func (m *StorageManager) replicateMetrics(ctx context.Context, metrics []models.Metric) {
	if m.config.Replication.Async {
		// 
		go m.doReplication(ctx, metrics)
	} else {
		// 
		m.doReplication(ctx, metrics)
	}
}

// doReplication 
func (m *StorageManager) doReplication(ctx context.Context, metrics []models.Metric) {
	for _, target := range m.config.Replication.Targets {
		if storage, exists := m.storages[target]; exists && storage != m.primaryStorage {
			if err := storage.Store(ctx, metrics); err != nil {
				fmt.Printf("Failed to replicate to %s: %v\n", target, err)
			}
		}
	}
}

// Query 
func (m *StorageManager) Query(ctx context.Context, query *models.MetricQuery) (*models.MetricQueryResult, error) {
	// ?
	storage := m.primaryStorage
	if query.Storage != "" {
		if s, exists := m.storages[query.Storage]; exists {
			storage = s
		}
	}
	
	if storage == nil {
		return nil, fmt.Errorf("no storage available for query")
	}
	
	return storage.Query(ctx, query)
}

// QueryMultiple 洢?
func (m *StorageManager) QueryMultiple(ctx context.Context, query *models.MetricQuery, storageNames []string) (map[string]*models.MetricQueryResult, error) {
	results := make(map[string]*models.MetricQueryResult)
	errors := make(map[string]error)
	
	// 洢
	var wg sync.WaitGroup
	var resultMutex sync.Mutex
	
	for _, name := range storageNames {
		if storage, exists := m.storages[name]; exists {
			wg.Add(1)
			go func(name string, storage interfaces.MetricStorage) {
				defer wg.Done()
				
				result, err := storage.Query(ctx, query)
				
				resultMutex.Lock()
				if err != nil {
					errors[name] = err
				} else {
					results[name] = result
				}
				resultMutex.Unlock()
			}(name, storage)
		}
	}
	
	wg.Wait()
	
	// 
	if len(errors) > 0 {
		return results, fmt.Errorf("query errors: %v", errors)
	}
	
	return results, nil
}

// AddStorage 洢
func (m *StorageManager) AddStorage(name string, storage interfaces.MetricStorage) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.storages[name]; exists {
		return fmt.Errorf("storage %s already exists", name)
	}
	
	// ?
	if err := storage.Health(m.ctx); err != nil {
		return fmt.Errorf("storage %s health check failed: %w", name, err)
	}
	
	m.storages[name] = storage
	
	// 洢洢
	if m.primaryStorage == nil {
		m.primaryStorage = storage
	}
	
	return nil
}

// RemoveStorage 洢
func (m *StorageManager) RemoveStorage(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	storage, exists := m.storages[name]
	if !exists {
		return fmt.Errorf("storage %s not found", name)
	}
	
	// ?
	if storage == m.primaryStorage {
		return fmt.Errorf("cannot remove primary storage %s", name)
	}
	
	// 洢
	if closer, ok := storage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close storage %s: %w", name, err)
		}
	}
	
	delete(m.storages, name)
	return nil
}

// GetStorage 洢
func (m *StorageManager) GetStorage(name string) (interfaces.MetricStorage, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	storage, exists := m.storages[name]
	if !exists {
		return nil, fmt.Errorf("storage %s not found", name)
	}
	
	return storage, nil
}

// ListStorages ?
func (m *StorageManager) ListStorages() map[string]interfaces.MetricStorage {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]interfaces.MetricStorage)
	for k, v := range m.storages {
		result[k] = v
	}
	
	return result
}

// SetPrimaryStorage ?
func (m *StorageManager) SetPrimaryStorage(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	storage, exists := m.storages[name]
	if !exists {
		return fmt.Errorf("storage %s not found", name)
	}
	
	m.primaryStorage = storage
	m.config.Primary = name
	
	return nil
}

// GetPrimaryStorage ?
func (m *StorageManager) GetPrimaryStorage() interfaces.MetricStorage {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.primaryStorage
}

// Health ?
func (m *StorageManager) Health() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var errors []string
	
	for name, storage := range m.storages {
		if err := storage.Health(m.ctx); err != nil {
			errors = append(errors, fmt.Sprintf("storage %s: %v", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("health check failed: %v", errors)
	}
	
	return nil
}

// GetStats 
func (m *StorageManager) GetStats() *ManagerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	stats := &ManagerStats{
		Running:        m.running,
		StorageCount:   len(m.storages),
		StorageStats:   make(map[string]StorageStats),
		LastUpdateTime: time.Now(),
	}
	
	if m.primaryStorage != nil {
		for name, storage := range m.storages {
			if storage == m.primaryStorage {
				stats.PrimaryStorage = name
				break
			}
		}
	}
	
	// 洢?
	for name, storage := range m.storages {
		storageStats := StorageStats{
			Name:   name,
			Status: "unknown",
		}
		
		// ?
		if err := storage.Health(m.ctx); err != nil {
			storageStats.Status = "unhealthy"
			storageStats.LastError = err.Error()
		} else {
			storageStats.Status = "healthy"
		}
		
		// ?
		if statsProvider, ok := storage.(interface {
			GetStats(context.Context) (map[string]interface{}, error)
		}); ok {
			if customStats, err := statsProvider.GetStats(m.ctx); err == nil {
				storageStats.CustomStats = customStats
			}
		}
		
		stats.StorageStats[name] = storageStats
	}
	
	return stats
}

// GetStorageStats 洢
func (m *StorageManager) GetStorageStats(name string) (*StorageStats, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	storage, exists := m.storages[name]
	if !exists {
		return nil, fmt.Errorf("storage %s not found", name)
	}
	
	stats := &StorageStats{
		Name:   name,
		Status: "unknown",
	}
	
	// ?
	if err := storage.Health(m.ctx); err != nil {
		stats.Status = "unhealthy"
		stats.LastError = err.Error()
	} else {
		stats.Status = "healthy"
	}
	
	// ?
	if statsProvider, ok := storage.(interface {
		GetStats(context.Context) (map[string]interface{}, error)
	}); ok {
		if customStats, err := statsProvider.GetStats(m.ctx); err == nil {
			stats.CustomStats = customStats
		}
	}
	
	return stats, nil
}

// IsRunning 
func (m *StorageManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.running
}

// GetConfig 
func (m *StorageManager) GetConfig() *StorageManagerConfig {
	return m.config
}

// UpdateConfig 
func (m *StorageManager) UpdateConfig(config *StorageManagerConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("cannot update config while running")
	}
	
	m.config = config
	return nil
}

// Backup 
func (m *StorageManager) Backup(ctx context.Context, location string) error {
	if m.config.Backup == nil || !m.config.Backup.Enabled {
		return fmt.Errorf("backup not enabled")
	}
	
	// 
	// 絼洢
	
	return fmt.Errorf("backup not implemented")
}

// Restore 
func (m *StorageManager) Restore(ctx context.Context, location string) error {
	if m.config.Backup == nil || !m.config.Backup.Enabled {
		return fmt.Errorf("backup not enabled")
	}
	
	// 
	// 洢
	
	return fmt.Errorf("restore not implemented")
}

// Migrate 
func (m *StorageManager) Migrate(ctx context.Context, from, to string) error {
	fromStorage, exists := m.storages[from]
	if !exists {
		return fmt.Errorf("source storage %s not found", from)
	}
	
	toStorage, exists := m.storages[to]
	if !exists {
		return fmt.Errorf("target storage %s not found", to)
	}
	
	// 
	// 洢?
	
	_ = fromStorage
	_ = toStorage
	
	return fmt.Errorf("migration not implemented")
}

// Compact 
func (m *StorageManager) Compact(ctx context.Context, storage string) error {
	s, exists := m.storages[storage]
	if !exists {
		return fmt.Errorf("storage %s not found", storage)
	}
	
	// 洢?
	if compactor, ok := s.(interface{ Compact(context.Context) error }); ok {
		return compactor.Compact(ctx)
	}
	
	return fmt.Errorf("storage %s does not support compaction", storage)
}

// Vacuum 
func (m *StorageManager) Vacuum(ctx context.Context, storage string) error {
	s, exists := m.storages[storage]
	if !exists {
		return fmt.Errorf("storage %s not found", storage)
	}
	
	// 洢?
	if vacuumer, ok := s.(interface{ Vacuum(context.Context) error }); ok {
		return vacuumer.Vacuum(ctx)
	}
	
	return fmt.Errorf("storage %s does not support vacuum", storage)
}

