package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host                string
	Port                int
	Username            string
	Password            string
	Database            string
	SSLMode             string
	MaxOpenConns        int
	MaxIdleConns        int
	MaxLifetime         time.Duration
	MaxIdleTime         time.Duration        // 连接最大空闲时�?
	ConnMaxIdleTime     time.Duration        // 连接空闲超时
	HealthCheckInterval time.Duration        // 健康检查间�?
	ReconnectInterval   time.Duration        // 重连间隔
	MaxReconnectAttempts int                 // 最大重连尝试次�?
}

// PostgresDB PostgreSQL数据库管理器
type PostgresDB struct {
	db                   *gorm.DB
	config               *PostgresConfig
	logger               *zap.Logger
	healthCheckTicker    *time.Ticker
	healthCheckStop      chan bool
	reconnectMutex       sync.RWMutex
	isHealthy            bool
	lastHealthCheck      time.Time
	connectionLeakDetector *ConnectionLeakDetector
}

// ConnectionLeakDetector 连接泄漏检测器
type ConnectionLeakDetector struct {
	maxConnections     int
	warningThreshold   float64
	checkInterval      time.Duration
	logger            *zap.Logger
	ticker            *time.Ticker
	stop              chan bool
}

// NewPostgresDB 创建新的PostgreSQL数据库连�?
func NewPostgresDB(config *PostgresConfig, log *zap.Logger) (*PostgresDB, error) {
	// 设置默认�?
	if config.MaxIdleTime == 0 {
		config.MaxIdleTime = 30 * time.Minute
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 10 * time.Minute
	}
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 30 * time.Second
	}
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 5 * time.Second
	}
	if config.MaxReconnectAttempts == 0 {
		config.MaxReconnectAttempts = 3
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	// 配置GORM日志
	gormLogger := logger.New(
		&gormLogWriter{logger: log},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接数据�?
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参�?
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	postgresDB := &PostgresDB{
		db:              db,
		config:          config,
		logger:          log,
		healthCheckStop: make(chan bool),
		isHealthy:       true,
		lastHealthCheck: time.Now(),
	}

	// 初始化连接泄漏检测器
	postgresDB.connectionLeakDetector = &ConnectionLeakDetector{
		maxConnections:   config.MaxOpenConns,
		warningThreshold: 0.8, // 80%阈�?
		checkInterval:    1 * time.Minute,
		logger:          log,
		stop:            make(chan bool),
	}

	// 启动健康检�?
	postgresDB.startHealthCheck()
	
	// 启动连接泄漏检�?
	postgresDB.startConnectionLeakDetection()

	log.Info("PostgreSQL connected successfully",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
		zap.Duration("max_lifetime", config.MaxLifetime),
	)

	return postgresDB, nil
}

// startHealthCheck 启动健康检�?
func (p *PostgresDB) startHealthCheck() {
	p.healthCheckTicker = time.NewTicker(p.config.HealthCheckInterval)
	
	go func() {
		for {
			select {
			case <-p.healthCheckTicker.C:
				p.performHealthCheck()
			case <-p.healthCheckStop:
				p.healthCheckTicker.Stop()
				return
			}
		}
	}()
}

// performHealthCheck 执行健康检�?
func (p *PostgresDB) performHealthCheck() {
	p.reconnectMutex.Lock()
	defer p.reconnectMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := p.db.DB()
	if err != nil {
		p.logger.Error("Failed to get sql.DB for health check", zap.Error(err))
		p.isHealthy = false
		p.attemptReconnect()
		return
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		p.logger.Error("Database health check failed", zap.Error(err))
		p.isHealthy = false
		p.attemptReconnect()
		return
	}

	if !p.isHealthy {
		p.logger.Info("Database connection restored")
	}
	
	p.isHealthy = true
	p.lastHealthCheck = time.Now()
	
	// 记录连接池统计信�?
	stats := sqlDB.Stats()
	p.logger.Debug("Database connection pool stats",
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
	)
}

// attemptReconnect 尝试重连
func (p *PostgresDB) attemptReconnect() {
	for attempt := 1; attempt <= p.config.MaxReconnectAttempts; attempt++ {
		p.logger.Info("Attempting to reconnect to database",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", p.config.MaxReconnectAttempts),
		)

		time.Sleep(p.config.ReconnectInterval)

		sqlDB, err := p.db.DB()
		if err != nil {
			p.logger.Error("Failed to get sql.DB for reconnection", zap.Error(err))
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := sqlDB.PingContext(ctx); err != nil {
			cancel()
			p.logger.Error("Reconnection attempt failed", 
				zap.Int("attempt", attempt), 
				zap.Error(err),
			)
			continue
		}
		cancel()

		p.logger.Info("Successfully reconnected to database", zap.Int("attempt", attempt))
		p.isHealthy = true
		return
	}

	p.logger.Error("Failed to reconnect after maximum attempts",
		zap.Int("max_attempts", p.config.MaxReconnectAttempts),
	)
}

// startConnectionLeakDetection 启动连接泄漏检�?
func (p *PostgresDB) startConnectionLeakDetection() {
	detector := p.connectionLeakDetector
	detector.ticker = time.NewTicker(detector.checkInterval)
	
	go func() {
		for {
			select {
			case <-detector.ticker.C:
				p.checkConnectionLeak()
			case <-detector.stop:
				detector.ticker.Stop()
				return
			}
		}
	}()
}

// checkConnectionLeak 检查连接泄�?
func (p *PostgresDB) checkConnectionLeak() {
	sqlDB, err := p.db.DB()
	if err != nil {
		return
	}

	stats := sqlDB.Stats()
	detector := p.connectionLeakDetector
	
	usageRatio := float64(stats.OpenConnections) / float64(detector.maxConnections)
	
	if usageRatio >= detector.warningThreshold {
		detector.logger.Warn("High database connection usage detected",
			zap.Int("open_connections", stats.OpenConnections),
			zap.Int("max_connections", detector.maxConnections),
			zap.Float64("usage_ratio", usageRatio),
			zap.Int("in_use", stats.InUse),
			zap.Int("idle", stats.Idle),
			zap.Int64("wait_count", stats.WaitCount),
		)
	}
}

// GetDB 获取GORM数据库实�?
func (p *PostgresDB) GetDB() *gorm.DB {
	return p.db
}

// Close 关闭数据库连�?
func (p *PostgresDB) Close() error {
	// 停止健康检�?
	if p.healthCheckTicker != nil {
		close(p.healthCheckStop)
	}
	
	// 停止连接泄漏检�?
	if p.connectionLeakDetector != nil && p.connectionLeakDetector.ticker != nil {
		close(p.connectionLeakDetector.stop)
	}

	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	
	p.logger.Info("Closing PostgreSQL database connections")
	return sqlDB.Close()
}

// Health 检查数据库健康状�?
func (p *PostgresDB) Health() error {
	p.reconnectMutex.RLock()
	defer p.reconnectMutex.RUnlock()
	
	if !p.isHealthy {
		return fmt.Errorf("database is unhealthy, last check: %v", p.lastHealthCheck)
	}
	
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	return sqlDB.PingContext(ctx)
}

// IsHealthy 返回数据库健康状�?
func (p *PostgresDB) IsHealthy() bool {
	p.reconnectMutex.RLock()
	defer p.reconnectMutex.RUnlock()
	return p.isHealthy
}

// GetStats 获取连接池统计信�?
func (p *PostgresDB) GetStats() map[string]interface{} {
	sqlDB, err := p.db.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
		"is_healthy":              p.isHealthy,
		"last_health_check":       p.lastHealthCheck,
		"connection_usage_ratio":  float64(stats.OpenConnections) / float64(stats.MaxOpenConnections),
	}
}

// AutoMigrate 自动迁移数据库表结构
func (p *PostgresDB) AutoMigrate(models ...interface{}) error {
	p.logger.Info("Starting database migration")
	
	for _, model := range models {
		if err := p.db.AutoMigrate(model); err != nil {
			p.logger.Error("Failed to migrate model", zap.Error(err))
			return fmt.Errorf("failed to migrate model: %w", err)
		}
	}
	
	p.logger.Info("Database migration completed successfully")
	return nil
}

// Transaction 执行事务
func (p *PostgresDB) Transaction(fn func(*gorm.DB) error) error {
	return p.db.Transaction(fn)
}

// gormLogWriter GORM日志写入�?
type gormLogWriter struct {
	logger *zap.Logger
}

// Printf 实现GORM日志接口
func (w *gormLogWriter) Printf(format string, args ...interface{}) {
	w.logger.Info(fmt.Sprintf(format, args...))
}
