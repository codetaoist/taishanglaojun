package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/config"
	"github.com/codetaoist/taishanglaojun/infrastructure/auth-system/internal/models"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数量据库管理器器
type Database struct {
	DB     *gorm.DB
	Redis  *redis.Client
	config *config.Config
	logger *zap.Logger
}

// New 创建数量据库管理器器
func New(cfg *config.Config, log *zap.Logger) (*Database, error) {
	db := &Database{
		config: cfg,
		logger: log,
	}

	// 连接数量据库
	if db.config.Database.Type != "disabled" {
		if err := db.connectDatabase(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	} else {
		db.logger.Info("Database connection disabled for testing")
	}

	// 连接Redis
	if err := db.connectRedis(); err != nil {
		db.logger.Warn("Failed to connect to Redis, continuing without Redis", zap.Error(err))
	}

	// 数量据库迁移
	if db.DB != nil {
		if err := db.migrate(); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	return db, nil
}

// connectDatabase 连接数量据库
func (d *Database) connectDatabase() error {
	dsn := d.config.GetDSN()

	// 配置GORM日志
	var gormLogger logger.Interface
	if d.config.IsDevelopment() {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	var db *gorm.DB
	var err error

	// 根据数量据库类型选择驱动
	switch d.config.Database.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: gormLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: gormLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
	case "sqlserver":
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{
			Logger: gormLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
	default:
		return fmt.Errorf("unsupported database type: %s", d.config.Database.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB对象
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(d.config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(d.config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(d.config.Database.MaxLifetime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.DB = db
	d.logger.Info("Connected to database",
		zap.String("type", d.config.Database.Type),
		zap.String("dsn", dsn))
	return nil
}

// connectRedis 连接Redis
func (d *Database) connectRedis() error {
	rdb := redis.NewClient(&redis.Options{
		Addr:         d.config.GetRedisAddr(),
		Password:     d.config.Redis.Password,
		DB:           d.config.Redis.Database,
		PoolSize:     d.config.Redis.PoolSize,
		MinIdleConns: d.config.Redis.MinIdleConns,
		DialTimeout:  d.config.Redis.DialTimeout,
		ReadTimeout:  d.config.Redis.ReadTimeout,
		WriteTimeout: d.config.Redis.WriteTimeout,
		IdleTimeout:  d.config.Redis.IdleTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	d.Redis = rdb
	d.logger.Info("Connected to Redis")
	return nil
}

// migrate 自动迁移数量据库表
func (d *Database) migrate() error {
	d.logger.Info("Starting database migration")

	// 迁移用户相关�?
	if err := d.DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Token{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserPermission{},
	); err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	// 创建索引
	if err := d.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// 初始化默认数量�?
	if err := d.seedData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	d.logger.Info("Database migration completed")
	return nil
}

// createIndexes 创建数量据库索�?
func (d *Database) createIndexes() error {
	// 检查数量据库类型，MySQL不支�?IF NOT EXISTS 语法
	dbType := d.config.Database.Type

	var indexQueries []string

	if dbType == "mysql" {
		// MySQL 索引创建（不使用户 IF NOT EXISTS�?
		indexQueries = []string{
			// 用户表索引（username、email、status、role已通过GORM自动创建）

			// 会话表索�?
			"CREATE INDEX idx_sessions_user_id ON sessions(user_id)",
			"CREATE INDEX idx_sessions_token ON sessions(token)",
			"CREATE INDEX idx_sessions_status ON sessions(status)",
			"CREATE INDEX idx_sessions_expires_at ON sessions(expires_at)",

			// 令牌表索�?
			"CREATE INDEX idx_tokens_user_id ON tokens(user_id)",
			"CREATE INDEX idx_tokens_token ON tokens(token)",
			"CREATE INDEX idx_tokens_type ON tokens(type)",
			"CREATE INDEX idx_tokens_status ON tokens(status)",
			"CREATE INDEX idx_tokens_expires_at ON tokens(expires_at)",
		}
	} else {
		// PostgreSQL 和其他数量据库索引创建（支�?IF NOT EXISTS�?
		indexQueries = []string{
			// 用户表索引（username、email、status、role已通过GORM自动创建）

			// 会话表索�?
			"CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)",
			"CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)",
			"CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)",

			// 令牌表索�?
			"CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id)",
			"CREATE INDEX IF NOT EXISTS idx_tokens_token ON tokens(token)",
			"CREATE INDEX IF NOT EXISTS idx_tokens_type ON tokens(type)",
			"CREATE INDEX IF NOT EXISTS idx_tokens_status ON tokens(status)",
			"CREATE INDEX IF NOT EXISTS idx_tokens_expires_at ON tokens(expires_at)",
		}
	}

	// 执行索引创建
	for _, query := range indexQueries {
		if err := d.DB.Exec(query).Error; err != nil {
			// 对于MySQL，如果索引已存在会话报错，我们忽略这个错误
			if dbType == "mysql" && strings.Contains(err.Error(), "Duplicate key name") {
				d.logger.Info("Index already exists, skipping", zap.String("query", query))
				continue
			}
			return err
		}
	}

	return nil
}

// seedData 初始化默认数量�?
func (d *Database) seedData() error {
	// 创建默认权限
	permissions := []models.Permission{
		{Name: "user.read", Description: "读取用户信息"},
		{Name: "user.write", Description: "修改用户信息"},
		{Name: "user.delete", Description: "删除用户"},
		{Name: "admin.read", Description: "管理员读取权限"},
		{Name: "admin.write", Description: "管理员写入权限"},
		{Name: "super_admin.all", Description: "超级管理员所有权限"},
	}

	for _, permission := range permissions {
		var existingPermission models.Permission
		if err := d.DB.Where("name = ?", permission.Name).First(&existingPermission).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := d.DB.Create(&permission).Error; err != nil {
					return fmt.Errorf("failed to create permission %s: %w", permission.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check permission %s: %w", permission.Name, err)
			}
		}
	}

	// 为角色分配权限�?
	rolePermissions := map[models.UserRole][]string{
		models.RoleUser: {
			"user.read",
		},
		models.RoleAdmin: {
			"user.read",
			"user.write",
			"admin.read",
			"admin.write",
		},
		models.RoleModerator: {
			"user.read",
			"user.write",
			"user.delete",
			"admin.read",
			"admin.write",
			"super_admin.all",
		},
	}

	for role, permissionNames := range rolePermissions {
		for _, permissionName := range permissionNames {
			var permission models.Permission
			if err := d.DB.Where("name = ?", permissionName).First(&permission).Error; err != nil {
				continue
			}

			var existingRolePermission models.RolePermission
			if err := d.DB.Where("role = ? AND permission_id = ?", role, permission.ID).First(&existingRolePermission).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					rolePermission := models.RolePermission{
						Role:         role,
						PermissionID: permission.ID,
					}
					if err := d.DB.Create(&rolePermission).Error; err != nil {
						return fmt.Errorf("failed to create role permission: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// Close 关闭数量据库连�?
func (d *Database) Close() error {
	var errs []error

	// 关闭PostgreSQL连接
	if d.DB != nil {
		if sqlDB, err := d.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close PostgreSQL: %w", err))
			}
		}
	}

	// 关闭Redis连接
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	d.logger.Info("Database connections closed")
	return nil
}

// Health 检查数量据库健康状态�?
func (d *Database) Health(ctx context.Context) error {
	// 检查PostgreSQL
	if d.DB != nil {
		sqlDB, err := d.DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("PostgreSQL health check failed: %w", err)
		}
	}

	// 检查Redis
	if d.Redis != nil {
		if err := d.Redis.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("Redis health check failed: %w", err)
		}
	}

	return nil
}

// GetDB 获取GORM数量据库实�?
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// GetRedis 获取Redis客户�?
func (d *Database) GetRedis() *redis.Client {
	return d.Redis
}

// Transaction 执行数量据库事�?
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// WithContext 使用户上下�?
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.DB.WithContext(ctx)
}
