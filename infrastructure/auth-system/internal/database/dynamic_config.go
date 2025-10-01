package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/config"
	"go.uber.org/zap"
)

// DynamicDatabase 动态数据库管理�?
type DynamicDatabase struct {
	databases map[string]*Database // 数据库连接池
	current   string               // 当前使用的数据库
	mutex     sync.RWMutex         // 读写�?
	logger    *zap.Logger
}

// DatabaseSwitchConfig 数据库切换配�?
type DatabaseSwitchConfig struct {
	Name     string                `json:"name"`     // 数据库配置名�?
	Database config.DatabaseConfig `json:"database"` // 数据库配�?
	Redis    config.RedisConfig    `json:"redis"`    // Redis配置
}

// NewDynamicDatabase 创建动态数据库管理�?
func NewDynamicDatabase(logger *zap.Logger) *DynamicDatabase {
	return &DynamicDatabase{
		databases: make(map[string]*Database),
		logger:    logger,
	}
}

// AddDatabase 添加数据库配�?
func (dd *DynamicDatabase) AddDatabase(name string, cfg *config.Config) error {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	// 检查是否已存在
	if _, exists := dd.databases[name]; exists {
		return fmt.Errorf("database configuration '%s' already exists", name)
	}

	// 创建数据库连�?
	db, err := New(cfg, dd.logger)
	if err != nil {
		return fmt.Errorf("failed to create database connection for '%s': %w", name, err)
	}

	dd.databases[name] = db
	
	// 如果是第一个数据库，设为当前数据库
	if dd.current == "" {
		dd.current = name
	}

	dd.logger.Info("Added database configuration", 
		zap.String("name", name),
		zap.String("type", cfg.Database.Type),
		zap.String("host", cfg.Database.Host))
	
	return nil
}

// SwitchDatabase 切换数据�?
func (dd *DynamicDatabase) SwitchDatabase(name string) error {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	// 检查数据库是否存在
	if _, exists := dd.databases[name]; !exists {
		return fmt.Errorf("database configuration '%s' not found", name)
	}

	oldDB := dd.current
	dd.current = name
	
	dd.logger.Info("Switched database", 
		zap.String("from", oldDB),
		zap.String("to", name))
	
	return nil
}

// GetCurrentDatabase 获取当前数据库连�?
func (dd *DynamicDatabase) GetCurrentDatabase() (*Database, error) {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	if dd.current == "" {
		return nil, fmt.Errorf("no database configuration is currently active")
	}

	db, exists := dd.databases[dd.current]
	if !exists {
		return nil, fmt.Errorf("current database configuration '%s' not found", dd.current)
	}

	return db, nil
}

// GetDatabase 获取指定名称的数据库连接
func (dd *DynamicDatabase) GetDatabase(name string) (*Database, error) {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	db, exists := dd.databases[name]
	if !exists {
		return nil, fmt.Errorf("database configuration '%s' not found", name)
	}

	return db, nil
}

// ListDatabases 列出所有数据库配置
func (dd *DynamicDatabase) ListDatabases() []string {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	names := make([]string, 0, len(dd.databases))
	for name := range dd.databases {
		names = append(names, name)
	}
	return names
}

// GetCurrentDatabaseName 获取当前数据库名�?
func (dd *DynamicDatabase) GetCurrentDatabaseName() string {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()
	return dd.current
}

// RemoveDatabase 移除数据库配�?
func (dd *DynamicDatabase) RemoveDatabase(name string) error {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	// 检查是否存�?
	db, exists := dd.databases[name]
	if !exists {
		return fmt.Errorf("database configuration '%s' not found", name)
	}

	// 不能删除当前正在使用的数据库
	if dd.current == name {
		return fmt.Errorf("cannot remove currently active database '%s'", name)
	}

	// 关闭数据库连�?
	if err := dd.closeDatabase(db); err != nil {
		dd.logger.Warn("Failed to close database connection", 
			zap.String("name", name),
			zap.Error(err))
	}

	delete(dd.databases, name)
	
	dd.logger.Info("Removed database configuration", zap.String("name", name))
	return nil
}

// HealthCheck 检查所有数据库连接健康状�?
func (dd *DynamicDatabase) HealthCheck(ctx context.Context) map[string]error {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	results := make(map[string]error)
	
	for name, db := range dd.databases {
		if db.DB != nil {
			sqlDB, err := db.DB.DB()
			if err != nil {
				results[name] = fmt.Errorf("failed to get sql.DB: %w", err)
				continue
			}
			
			if err := sqlDB.PingContext(ctx); err != nil {
				results[name] = fmt.Errorf("ping failed: %w", err)
			} else {
				results[name] = nil // 健康
			}
		} else {
			results[name] = fmt.Errorf("database connection is nil")
		}
	}
	
	return results
}

// Close 关闭所有数据库连接
func (dd *DynamicDatabase) Close() error {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	var lastErr error
	for name, db := range dd.databases {
		if err := dd.closeDatabase(db); err != nil {
			dd.logger.Error("Failed to close database", 
				zap.String("name", name),
				zap.Error(err))
			lastErr = err
		}
	}

	dd.databases = make(map[string]*Database)
	dd.current = ""
	
	return lastErr
}

// closeDatabase 关闭单个数据库连�?
func (dd *DynamicDatabase) closeDatabase(db *Database) error {
	if db.DB != nil {
		sqlDB, err := db.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	
	if db.Redis != nil {
		if err := db.Redis.Close(); err != nil {
			return fmt.Errorf("failed to close Redis: %w", err)
		}
	}
	
	return nil
}

// GetDatabaseStats 获取数据库统计信�?
func (dd *DynamicDatabase) GetDatabaseStats() map[string]interface{} {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["current_database"] = dd.current
	stats["total_databases"] = len(dd.databases)
	stats["database_names"] = dd.ListDatabases()
	
	// 获取当前数据库的连接池统�?
	if dd.current != "" {
		if db, exists := dd.databases[dd.current]; exists && db.DB != nil {
			if sqlDB, err := db.DB.DB(); err == nil {
				dbStats := sqlDB.Stats()
				stats["current_db_stats"] = map[string]interface{}{
					"open_connections":     dbStats.OpenConnections,
					"in_use":              dbStats.InUse,
					"idle":                dbStats.Idle,
					"wait_count":          dbStats.WaitCount,
					"wait_duration":       dbStats.WaitDuration,
					"max_idle_closed":     dbStats.MaxIdleClosed,
					"max_idle_time_closed": dbStats.MaxIdleTimeClosed,
					"max_lifetime_closed": dbStats.MaxLifetimeClosed,
				}
			}
		}
	}
	
	return stats
}
