package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConnectionMonitor 连接监控�?
type ConnectionMonitor struct {
	logger               *zap.Logger
	config               *ConnectionMonitorConfig
	
	// 监控状�?
	isRunning            bool
	stopChan             chan struct{}
	wg                   sync.WaitGroup
	mutex                sync.RWMutex
	
	// 连接池监�?
	connectionPools      map[string]ConnectionPool
	
	// 泄漏检�?
	leakDetectionTicker  *time.Ticker
	
	// 优雅关闭
	shutdownTimeout      time.Duration
	shutdownCallbacks    []ShutdownCallback
}

// ConnectionMonitorConfig 连接监控配置
type ConnectionMonitorConfig struct {
	LeakDetectionInterval    time.Duration
	LeakWarningThreshold     float64  // 连接使用率警告阈�?
	LeakCriticalThreshold    float64  // 连接使用率严重阈�?
	ShutdownTimeout          time.Duration
	MetricsCollectionInterval time.Duration
	EnableDetailedLogging    bool
}

// ConnectionPool 连接池接�?
type ConnectionPool interface {
	GetStats() ConnectionStats
	GetName() string
	IsHealthy() bool
	Close() error
}

// ConnectionStats 连接统计信息
type ConnectionStats struct {
	TotalConnections    int32
	ActiveConnections   int32
	IdleConnections     int32
	MaxConnections      int32
	ConnectionsCreated  int64
	ConnectionsClosed   int64
	ConnectionErrors    int64
	AverageWaitTime     time.Duration
	MaxWaitTime         time.Duration
}

// ShutdownCallback 关闭回调函数
type ShutdownCallback func(ctx context.Context) error

// NewConnectionMonitor 创建新的连接监控�?
func NewConnectionMonitor(config *ConnectionMonitorConfig, logger *zap.Logger) *ConnectionMonitor {
	if config == nil {
		config = &ConnectionMonitorConfig{
			LeakDetectionInterval:     1 * time.Minute,
			LeakWarningThreshold:      0.8,  // 80%
			LeakCriticalThreshold:     0.95, // 95%
			ShutdownTimeout:           30 * time.Second,
			MetricsCollectionInterval: 30 * time.Second,
			EnableDetailedLogging:     false,
		}
	}

	return &ConnectionMonitor{
		logger:            logger,
		config:            config,
		stopChan:          make(chan struct{}),
		connectionPools:   make(map[string]ConnectionPool),
		shutdownTimeout:   config.ShutdownTimeout,
		shutdownCallbacks: make([]ShutdownCallback, 0),
	}
}

// RegisterConnectionPool 注册连接�?
func (cm *ConnectionMonitor) RegisterConnectionPool(name string, pool ConnectionPool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.connectionPools[name] = pool
	cm.logger.Info("Connection pool registered",
		zap.String("pool_name", name),
		zap.Bool("healthy", pool.IsHealthy()),
	)
}

// UnregisterConnectionPool 注销连接�?
func (cm *ConnectionMonitor) UnregisterConnectionPool(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	delete(cm.connectionPools, name)
	cm.logger.Info("Connection pool unregistered", zap.String("pool_name", name))
}

// RegisterShutdownCallback 注册关闭回调
func (cm *ConnectionMonitor) RegisterShutdownCallback(callback ShutdownCallback) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.shutdownCallbacks = append(cm.shutdownCallbacks, callback)
}

// Start 启动连接监控
func (cm *ConnectionMonitor) Start() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if cm.isRunning {
		return fmt.Errorf("connection monitor is already running")
	}
	
	cm.isRunning = true
	
	// 启动泄漏检�?
	cm.startLeakDetection()
	
	// 启动指标收集
	cm.startMetricsCollection()
	
	cm.logger.Info("Connection monitor started",
		zap.Duration("leak_detection_interval", cm.config.LeakDetectionInterval),
		zap.Float64("warning_threshold", cm.config.LeakWarningThreshold),
		zap.Float64("critical_threshold", cm.config.LeakCriticalThreshold),
	)
	
	return nil
}

// Stop 停止连接监控
func (cm *ConnectionMonitor) Stop() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	if !cm.isRunning {
		return nil
	}
	
	cm.logger.Info("Stopping connection monitor")
	
	// 发送停止信�?
	close(cm.stopChan)
	
	// 等待所有goroutine结束
	cm.wg.Wait()
	
	// 停止泄漏检�?
	if cm.leakDetectionTicker != nil {
		cm.leakDetectionTicker.Stop()
	}
	
	cm.isRunning = false
	cm.logger.Info("Connection monitor stopped")
	
	return nil
}

// startLeakDetection 启动泄漏检�?
func (cm *ConnectionMonitor) startLeakDetection() {
	cm.leakDetectionTicker = time.NewTicker(cm.config.LeakDetectionInterval)
	
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		
		for {
			select {
			case <-cm.leakDetectionTicker.C:
				cm.performLeakDetection()
			case <-cm.stopChan:
				return
			}
		}
	}()
}

// startMetricsCollection 启动指标收集
func (cm *ConnectionMonitor) startMetricsCollection() {
	ticker := time.NewTicker(cm.config.MetricsCollectionInterval)
	
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				cm.collectMetrics()
			case <-cm.stopChan:
				return
			}
		}
	}()
}

// performLeakDetection 执行泄漏检�?
func (cm *ConnectionMonitor) performLeakDetection() {
	cm.mutex.RLock()
	pools := make(map[string]ConnectionPool)
	for name, pool := range cm.connectionPools {
		pools[name] = pool
	}
	cm.mutex.RUnlock()
	
	for name, pool := range pools {
		stats := pool.GetStats()
		
		if stats.MaxConnections == 0 {
			continue
		}
		
		usageRatio := float64(stats.ActiveConnections) / float64(stats.MaxConnections)
		
		// 检查警告阈�?
		if usageRatio >= cm.config.LeakWarningThreshold {
			level := "warning"
			if usageRatio >= cm.config.LeakCriticalThreshold {
				level = "critical"
			}
			
			cm.logger.Warn("High connection usage detected",
				zap.String("pool_name", name),
				zap.String("level", level),
				zap.Float64("usage_ratio", usageRatio),
				zap.Int32("active_connections", stats.ActiveConnections),
				zap.Int32("max_connections", stats.MaxConnections),
				zap.Int32("idle_connections", stats.IdleConnections),
				zap.Int64("connection_errors", stats.ConnectionErrors),
				zap.Duration("average_wait_time", stats.AverageWaitTime),
			)
			
			// 如果达到严重阈值，记录详细信息
			if usageRatio >= cm.config.LeakCriticalThreshold {
				cm.logDetailedConnectionInfo(name, stats)
			}
		}
	}
}

// collectMetrics 收集指标
func (cm *ConnectionMonitor) collectMetrics() {
	if !cm.config.EnableDetailedLogging {
		return
	}
	
	cm.mutex.RLock()
	pools := make(map[string]ConnectionPool)
	for name, pool := range cm.connectionPools {
		pools[name] = pool
	}
	cm.mutex.RUnlock()
	
	for name, pool := range pools {
		stats := pool.GetStats()
		
		cm.logger.Debug("Connection pool metrics",
			zap.String("pool_name", name),
			zap.Bool("healthy", pool.IsHealthy()),
			zap.Int32("total_connections", stats.TotalConnections),
			zap.Int32("active_connections", stats.ActiveConnections),
			zap.Int32("idle_connections", stats.IdleConnections),
			zap.Int32("max_connections", stats.MaxConnections),
			zap.Int64("connections_created", stats.ConnectionsCreated),
			zap.Int64("connections_closed", stats.ConnectionsClosed),
			zap.Int64("connection_errors", stats.ConnectionErrors),
			zap.Duration("average_wait_time", stats.AverageWaitTime),
			zap.Duration("max_wait_time", stats.MaxWaitTime),
		)
	}
}

// logDetailedConnectionInfo 记录详细连接信息
func (cm *ConnectionMonitor) logDetailedConnectionInfo(poolName string, stats ConnectionStats) {
	cm.logger.Error("Critical connection usage - detailed analysis",
		zap.String("pool_name", poolName),
		zap.Int32("total_connections", stats.TotalConnections),
		zap.Int32("active_connections", stats.ActiveConnections),
		zap.Int32("idle_connections", stats.IdleConnections),
		zap.Int32("max_connections", stats.MaxConnections),
		zap.Int64("connections_created", stats.ConnectionsCreated),
		zap.Int64("connections_closed", stats.ConnectionsClosed),
		zap.Int64("connection_errors", stats.ConnectionErrors),
		zap.Float64("error_rate", float64(stats.ConnectionErrors)/float64(stats.ConnectionsCreated)),
		zap.Duration("average_wait_time", stats.AverageWaitTime),
		zap.Duration("max_wait_time", stats.MaxWaitTime),
	)
}

// GracefulShutdown 优雅关闭
func (cm *ConnectionMonitor) GracefulShutdown(ctx context.Context) error {
	cm.logger.Info("Starting graceful shutdown of database connections")
	
	// 创建带超时的上下�?
	shutdownCtx, cancel := context.WithTimeout(ctx, cm.shutdownTimeout)
	defer cancel()
	
	// 执行关闭回调
	var errors []error
	for i, callback := range cm.shutdownCallbacks {
		cm.logger.Debug("Executing shutdown callback", zap.Int("callback_index", i))
		
		if err := callback(shutdownCtx); err != nil {
			errors = append(errors, fmt.Errorf("shutdown callback %d failed: %w", i, err))
			cm.logger.Error("Shutdown callback failed", 
				zap.Int("callback_index", i), 
				zap.Error(err),
			)
		}
	}
	
	// 关闭所有连接池
	cm.mutex.RLock()
	pools := make(map[string]ConnectionPool)
	for name, pool := range cm.connectionPools {
		pools[name] = pool
	}
	cm.mutex.RUnlock()
	
	for name, pool := range pools {
		cm.logger.Debug("Closing connection pool", zap.String("pool_name", name))
		
		if err := pool.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close pool %s: %w", name, err))
			cm.logger.Error("Failed to close connection pool", 
				zap.String("pool_name", name), 
				zap.Error(err),
			)
		} else {
			cm.logger.Info("Connection pool closed successfully", zap.String("pool_name", name))
		}
	}
	
	// 停止监控�?
	if err := cm.Stop(); err != nil {
		errors = append(errors, fmt.Errorf("failed to stop connection monitor: %w", err))
	}
	
	if len(errors) > 0 {
		cm.logger.Error("Graceful shutdown completed with errors", zap.Int("error_count", len(errors)))
		return fmt.Errorf("graceful shutdown errors: %v", errors)
	}
	
	cm.logger.Info("Graceful shutdown completed successfully")
	return nil
}

// GetConnectionStats 获取所有连接池统计信息
func (cm *ConnectionMonitor) GetConnectionStats() map[string]ConnectionStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	stats := make(map[string]ConnectionStats)
	for name, pool := range cm.connectionPools {
		stats[name] = pool.GetStats()
	}
	
	return stats
}

// GetHealthStatus 获取所有连接池健康状�?
func (cm *ConnectionMonitor) GetHealthStatus() map[string]bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	status := make(map[string]bool)
	for name, pool := range cm.connectionPools {
		status[name] = pool.IsHealthy()
	}
	
	return status
}

// IsRunning 检查监控器是否运行�?
func (cm *ConnectionMonitor) IsRunning() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.isRunning
}
