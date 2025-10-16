package database

import (
	"fmt"
	stdlog "log"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
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

// New 创建数据库管理器
func New(config Config, log *zap.Logger) (*Database, error) {
	var dsn string
	var dialector gorm.Dialector

	switch config.Driver {
	case "postgres", "postgresql":
		// 为 Postgres 连接追加 connect_timeout 与 application_name，提升连接健壮性与可观测性
		timeoutSecs := int(config.ConnectTimeout.Seconds())
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d application_name=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode, timeoutSecs, "taishang-core-services")
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dsn = config.Database
		dialector = sqlite.Open(dsn)
	case "sqlserver", "mssql":
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		dialector = sqlserver.Open(dsn)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	// 设置GORM日志
	gormLogger := logger.New(
		stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 初始化数据库连接
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化数据库连接失败: %w", err)
	}

	// 获取底层sql.DB连接
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("测试数据库连接失败: %w", err)
	}

	// 针对 Postgres 强制设置会话参数并记录编码/排序信息，避免潜在编码问题并优化查询行为
	if config.Driver == "postgres" || config.Driver == "postgresql" {
		// 强制会话编码与字符串模式（多数情况下为默认值，显式设置更安全）
		_ = db.Exec("SET client_encoding TO 'UTF8'").Error
		_ = db.Exec("SET standard_conforming_strings = on").Error
		// 设置应用名称与语句超时，防止长时间运行的查询阻塞
		_ = db.Exec("SET application_name = 'taishang-core-services'").Error
		_ = db.Exec("SET statement_timeout = '30s'").Error

		// 记录当前数据库的编码与排序设置，便于排障
		// 使用 current_setting 查询以便可靠扫描
		type settingRow struct{ V string }
		var srvEnc, cliEnc settingRow
		_ = db.Raw("SELECT current_setting('server_encoding') AS v").Scan(&srvEnc).Error
		_ = db.Raw("SELECT current_setting('client_encoding') AS v").Scan(&cliEnc).Error
		// 读取数据库的 LC_COLLATE / LC_CTYPE
		type dbLocale struct{ Datcollate string; Datctype string }
		var dbl dbLocale
		_ = db.Raw("SELECT datcollate, datctype FROM pg_database WHERE datname = current_database()").Scan(&dbl).Error

		log.Info("Postgres 编码与排序", 
			zap.String("server_encoding", srvEnc.V), 
			zap.String("client_encoding", cliEnc.V), 
			zap.String("lc_collate", dbl.Datcollate), 
			zap.String("lc_ctype", dbl.Datctype))
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

	d.logger.Info("数据库迁移成功")
	return nil
}

// Health 检查数据库连接是否正常
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
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}
