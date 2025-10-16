package migrations

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migrator 数据库迁移器
type Migrator struct {
	db           *gorm.DB
	logger       *zap.Logger
	migrationDir string
}

// NewMigrator 创建新的迁移器
func NewMigrator(db *gorm.DB, logger *zap.Logger, migrationDir string) *Migrator {
	return &Migrator{
		db:           db,
		logger:       logger,
		migrationDir: migrationDir,
	}
}

// Up 执行向上迁移
func (m *Migrator) Up() error {
	migrate, err := m.getMigrate()
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Up(); err != nil && err != migrate.ErrNoChange {
		m.logger.Error("Failed to run up migrations", zap.Error(err))
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	m.logger.Info("Up migrations completed successfully")
	return nil
}

// Down 执行向下迁移
func (m *Migrator) Down() error {
	migrate, err := m.getMigrate()
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Down(); err != nil && err != migrate.ErrNoChange {
		m.logger.Error("Failed to run down migrations", zap.Error(err))
		return fmt.Errorf("failed to run down migrations: %w", err)
	}

	m.logger.Info("Down migrations completed successfully")
	return nil
}

// Steps 执行指定步数的迁�?
func (m *Migrator) Steps(n int) error {
	migrate, err := m.getMigrate()
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		m.logger.Error("Failed to run step migrations",
			zap.Int("steps", n),
			zap.Error(err))
		return fmt.Errorf("failed to run step migrations: %w", err)
	}

	m.logger.Info("Step migrations completed successfully", zap.Int("steps", n))
	return nil
}

// Goto 迁移到指定版�?
func (m *Migrator) Goto(version uint) error {
	migrate, err := m.getMigrate()
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Migrate(version); err != nil && err != migrate.ErrNoChange {
		m.logger.Error("Failed to migrate to version",
			zap.Uint("version", version),
			zap.Error(err))
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	m.logger.Info("Migration to version completed successfully", zap.Uint("version", version))
	return nil
}

// Version 获取值当前迁移版本
func (m *Migrator) Version() (uint, bool, error) {
	migrate, err := m.getMigrate()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	version, dirty, err := migrate.Version()
	if err != nil {
		m.logger.Error("Failed to get migration version", zap.Error(err))
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}

// Force 强制设置迁移版本（用于修复脏状态）
func (m *Migrator) Force(version int) error {
	migrate, err := m.getMigrate()
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Force(version); err != nil {
		m.logger.Error("Failed to force migration version",
			zap.Int("version", version),
			zap.Error(err))
		return fmt.Errorf("failed to force migration version: %w", err)
	}

	m.logger.Info("Migration version forced successfully", zap.Int("version", version))
	return nil
}

// Drop 删除键所有表（危险操作）
func (m *Migrator) Drop() error {
	migrate, err := m.getMigrate()
	if err != nil {
		return fmt.Errorf("failed to get migrate instance: %w", err)
	}
	defer migrate.Close()

	if err := migrate.Drop(); err != nil {
		m.logger.Error("Failed to drop database", zap.Error(err))
		return fmt.Errorf("failed to drop database: %w", err)
	}

	m.logger.Warn("Database dropped successfully")
	return nil
}

// getMigrate 获取值migrate实例
func (m *Migrator) getMigrate() (*migrate.Migrate, error) {
	// 获取值底层的sql.DB
	sqlDB, err := m.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 创建postgres驱动
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 构建迁移文件路径
	sourceURL := fmt.Sprintf("file://%s", filepath.ToSlash(m.migrationDir))

	// 创建migrate实例
	migrate, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return migrate, nil
}

// AutoMigrate 自动迁移（使用GORM自动迁移）
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	if err := m.db.AutoMigrate(models...); err != nil {
		m.logger.Error("Failed to auto migrate", zap.Error(err))
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	m.logger.Info("Auto migration completed successfully")
	return nil
}

// CreateMigration 创建新的迁移文件
func (m *Migrator) CreateMigration(name string) error {
	// 这里可以实现创建迁移文件的逻辑
	// 由于golang-migrate没有直接的API，这里只是一个占位符
	m.logger.Info("Migration file creation requested", zap.String("name", name))
	return fmt.Errorf("migration file creation not implemented")
}

// Status 获取值迁移状态
func (m *Migrator) Status() (MigrationStatus, error) {
	version, dirty, err := m.Version()
	if err != nil {
		return MigrationStatus{}, err
	}

	return MigrationStatus{
		Version: version,
		Dirty:   dirty,
	}, nil
}

// MigrationStatus 迁移状态
type MigrationStatus struct {
	Version uint `json:"version"`
	Dirty   bool `json:"dirty"`
}

// Validate 验证数据库连接和迁移状态
func (m *Migrator) Validate() error {
	// 检查数据库连接
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// 检查迁移状态
	_, dirty, err := m.Version()
	if err != nil {
		// 如果是第一次运行，可能没有迁移状态
		if err == migrate.ErrNilVersion {
			m.logger.Info("No migrations have been run yet")
			return nil
		}
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state, please fix manually")
	}

	m.logger.Info("Database validation passed")
	return nil
}

// Reset 重置数据库（删除键所有数据并重新迁移）
func (m *Migrator) Reset() error {
	m.logger.Warn("Resetting database - this will delete all data")

	// 先执行down到最初状态
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to down migrations: %w", err)
	}

	// 再执行up到最新状态
	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to up migrations after reset: %w", err)
	}

	m.logger.Info("Database reset completed successfully")
	return nil
}

