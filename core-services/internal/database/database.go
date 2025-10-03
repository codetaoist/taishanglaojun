package database

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	db     *gorm.DB
	logger *zap.Logger
}

// Config 数据库配置
type Config struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout"`
}

// New 创建新的数据库连接
func New(config Config, log *zap.Logger) (*Database, error) {
	var dsn string
	var dialector gorm.Dialector

	switch config.Driver {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		dialector = mysql.Open(dsn)
	case "sqlserver", "mssql":
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		dialector = sqlserver.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	// 配置GORM日志
	gormLogger := logger.New(
		stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接数据库
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层的sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Info("数据库连接成功", zap.String("driver", config.Driver), zap.String("host", config.Host))

	return &Database{
		db:     db,
		logger: log,
	}, nil
}

// GetDB 获取GORM数据库实例
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Migrate 执行数据库迁移
func (d *Database) Migrate(models ...interface{}) error {
	d.logger.Info("开始数据库迁移")
	
	for _, model := range models {
		if err := d.db.AutoMigrate(model); err != nil {
			d.logger.Error("数据库迁移失败", zap.Error(err))
			return fmt.Errorf("数据库迁移失败: %w", err)
		}
	}
	
	d.logger.Info("数据库迁移完成")
	return nil
}

// Health 检查数据库健康状态
func (d *Database) Health() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Stats 获取数据库连接统计信息
func (d *Database) Stats() map[string]interface{} {
	sqlDB, err := d.db.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

