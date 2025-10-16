package database

import (
    "fmt"

    "go.uber.org/zap"
    "gorm.io/gorm"
    "golang.org/x/crypto/bcrypt"

    "github.com/codetaoist/taishanglaojun/core-services/internal/models"
    "github.com/codetaoist/taishanglaojun/core-services/internal/handlers"
    "github.com/codetaoist/taishanglaojun/core-services/internal/services"
)

// MigrationService
type MigrationService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMigrationService
func NewMigrationService(db *gorm.DB, logger *zap.Logger) *MigrationService {
	return &MigrationService{
		db:     db,
		logger: logger,
	}
}

// AutoMigrate
func (m *MigrationService) AutoMigrate() error {
	m.logger.Info("Starting database migration...")

	// 如果已存在超级管理员，跳过创建
	if err := m.db.AutoMigrate(&models.User{}); err != nil {
		m.logger.Error("Failed to migrate User table", zap.Error(err))
		return fmt.Errorf("迁移用户表失败: %v", err)
	}

	// 迁移应用模块表
	if err := m.db.AutoMigrate(&models.AppModule{}); err != nil {
		m.logger.Error("Failed to migrate AppModule table", zap.Error(err))
		return fmt.Errorf("迁移应用模块表失败: %v", err)
	}

	// 迁移用户模块权限表
	if err := m.db.AutoMigrate(&models.UserModulePermission{}); err != nil {
		m.logger.Error("Failed to migrate UserModulePermission table", zap.Error(err))
		return fmt.Errorf("迁移用户模块权限表失败: %v", err)
	}

	// 迁移用户偏好表
	if err := m.db.AutoMigrate(&models.UserPreference{}); err != nil {
		m.logger.Error("Failed to migrate UserPreference table", zap.Error(err))
		return fmt.Errorf("迁移用户偏好表失败: %v", err)
	}

	// 迁移模块依赖表
	if err := m.db.AutoMigrate(&models.ModuleDependency{}); err != nil {
		m.logger.Error("Failed to migrate ModuleDependency table", zap.Error(err))
		return fmt.Errorf("迁移模块依赖表失败: %v", err)
	}

	// 迁移用户会话表
	if err := m.db.AutoMigrate(&models.UserSession{}); err != nil {
		m.logger.Error("Failed to migrate UserSession table", zap.Error(err))
		return fmt.Errorf("迁移用户会话表失败: %v", err)
	}

	// 迁移菜单表
	if err := m.db.AutoMigrate(&models.Menu{}); err != nil {
		m.logger.Error("Failed to migrate Menu table", zap.Error(err))
		return fmt.Errorf("迁移菜单表失败: %v", err)
	}

	// 先删除可能存在的外键约束
	m.dropForeignKeyConstraints()

	// 迁移权限表（使用handlers包中的结构体）
	if err := m.db.AutoMigrate(&handlers.Permission{}); err != nil {
		m.logger.Error("Failed to migrate Permission table", zap.Error(err))
		return fmt.Errorf("迁移权限表失败: %v", err)
	}

	// 迁移角色表（使用handlers包中的结构体）
	if err := m.db.AutoMigrate(&handlers.Role{}); err != nil {
		m.logger.Error("Failed to migrate Role table", zap.Error(err))
		return fmt.Errorf("迁移角色表失败: %v", err)
	}

	// 迁移用户角色关联表（使用handlers包中的结构体）
	if err := m.db.AutoMigrate(&handlers.UserRole{}); err != nil {
		m.logger.Error("Failed to migrate UserRole table", zap.Error(err))
		return fmt.Errorf("迁移用户角色关联表失败: %v", err)
	}

	// 迁移系统配置表
	if err := m.db.AutoMigrate(&models.SystemConfig{}); err != nil {
		m.logger.Error("Failed to migrate SystemConfig table", zap.Error(err))
		return fmt.Errorf("迁移系统配置表失败: %v", err)
	}

	// 迁移数据库连接配置表
	if err := m.db.AutoMigrate(&models.DatabaseConnection{}); err != nil {
		m.logger.Error("Failed to migrate DatabaseConnection table", zap.Error(err))
		return fmt.Errorf("迁移数据库连接配置表失败: %v", err)
	}

	// 迁移数据库连接状态表
	if err := m.db.AutoMigrate(&models.DatabaseConnectionStatus{}); err != nil {
		m.logger.Error("Failed to migrate DatabaseConnectionStatus table", zap.Error(err))
		return fmt.Errorf("迁移数据库连接状态表失败: %v", err)
	}

	// 迁移数据库连接事件表
	if err := m.db.AutoMigrate(&models.DatabaseConnectionEvent{}); err != nil {
		m.logger.Error("Failed to migrate DatabaseConnectionEvent table", zap.Error(err))
		return fmt.Errorf("迁移数据库连接事件表失败: %v", err)
	}

	// 迁移API文档相关表
	if err := m.db.AutoMigrate(&models.APICategory{}); err != nil {
		m.logger.Error("Failed to migrate APICategory table", zap.Error(err))
		return fmt.Errorf("迁移API分类表失败: %v", err)
	}

	if err := m.db.AutoMigrate(&models.APIEndpoint{}); err != nil {
		m.logger.Error("Failed to migrate APIEndpoint table", zap.Error(err))
		return fmt.Errorf("迁移API接口表失败: %v", err)
	}

	if err := m.db.AutoMigrate(&models.APIDocumentationSource{}); err != nil {
		m.logger.Error("Failed to migrate APIDocumentationSource table", zap.Error(err))
		return fmt.Errorf("迁移API文档来源表失败: %v", err)
	}

	if err := m.db.AutoMigrate(&models.APITestRecord{}); err != nil {
		m.logger.Error("Failed to migrate APITestRecord table", zap.Error(err))
		return fmt.Errorf("迁移API测试记录表失败: %v", err)
	}

	if err := m.db.AutoMigrate(&models.APIChangeLog{}); err != nil {
		m.logger.Error("Failed to migrate APIChangeLog table", zap.Error(err))
		return fmt.Errorf("迁移API变更日志表失败: %v", err)
	}

	m.logger.Info("Database migration completed successfully")
	return nil
}

// dropForeignKeyConstraints 删除外键约束 (PostgreSQL)
func (m *MigrationService) dropForeignKeyConstraints() {
	// 检查并删除所有引用permissions表的外键约束
	var count int64
	
	// 检查role_permissions表的外键约束 (MySQL语法)
	m.db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = 'laojun' AND table_name = 'role_permissions' AND constraint_name = 'fk_role_permissions_permission'").Scan(&count)
	if count > 0 {
		m.db.Exec("ALTER TABLE role_permissions DROP FOREIGN KEY fk_role_permissions_permission")
		m.logger.Info("Dropped foreign key constraint: fk_role_permissions_permission")
	}
	
	m.db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = 'laojun' AND table_name = 'role_permissions' AND constraint_name = 'fk_role_permissions_role'").Scan(&count)
	if count > 0 {
		m.db.Exec("ALTER TABLE role_permissions DROP FOREIGN KEY fk_role_permissions_role")
		m.logger.Info("Dropped foreign key constraint: fk_role_permissions_role")
	}
	
	// 检查user_permissions表的外键约束 (MySQL语法)
	m.db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = 'laojun' AND table_name = 'user_permissions' AND constraint_name = 'fk_user_permissions_permission'").Scan(&count)
	if count > 0 {
		m.db.Exec("ALTER TABLE user_permissions DROP FOREIGN KEY fk_user_permissions_permission")
		m.logger.Info("Dropped foreign key constraint: fk_user_permissions_permission")
	}
	
	// 检查其他可能的外键约束 (MySQL语法)
	m.db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = 'laojun' AND table_name = 'user_permissions' AND constraint_name = 'fk_user_permissions_user'").Scan(&count)
	if count > 0 {
		m.db.Exec("ALTER TABLE user_permissions DROP FOREIGN KEY fk_user_permissions_user")
		m.logger.Info("Dropped foreign key constraint: fk_user_permissions_user")
	}
	
	m.logger.Info("Foreign key constraints check completed")
}

// CreateIndexes 创建数据库索引
func (m *MigrationService) CreateIndexes() error {
	m.logger.Info("Creating database indexes...")

	// 获取数据库方言
	dialect := m.db.Dialector.Name()

	// 定义索引创建函数
	createIndex := func(tableName, indexName, columns string) error {
		var indexSQL string
		if dialect == "sqlite" {
			indexSQL = fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(%s)", indexName, tableName, columns)
		} else {
			// 对于MySQL/PostgreSQL，先检查索引是否存在
			var count int64
			if dialect == "mysql" {
				m.db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?", tableName, indexName).Scan(&count)
			} else if dialect == "postgres" {
				m.db.Raw("SELECT COUNT(*) FROM pg_indexes WHERE tablename = ? AND indexname = ?", tableName, indexName).Scan(&count)
			}
			if count == 0 {
				indexSQL = fmt.Sprintf("CREATE INDEX %s ON %s(%s)", indexName, tableName, columns)
			}
		}
		
		if indexSQL != "" {
			if err := m.db.Exec(indexSQL).Error; err != nil {
				m.logger.Error("Failed to create index", zap.String("index", indexName), zap.Error(err))
				return fmt.Errorf("创建索引 %s 失败: %v", indexName, err)
			}
		}
		return nil
	}

	// 创建各种索引
	indexes := []struct {
		table   string
		name    string
		columns string
	}{
		{"user_module_permissions", "idx_user_module_permission", "user_id, module_id"},
		{"module_dependencies", "idx_module_dependency", "module_id, dependency_id"},
		{"user_sessions", "idx_user_session_token", "token"},
		{"user_sessions", "idx_user_session_expires", "expires_at"},
		{"app_modules", "idx_app_module_category", "category"},
		{"app_modules", "idx_app_module_role", "required_role"},
	}

	// 创建所有索引
	for _, idx := range indexes {
		if err := createIndex(idx.table, idx.name, idx.columns); err != nil {
			return err
		}
	}

	m.logger.Info("Database indexes created successfully")
	return nil
}

// SeedData 初始化数据库
func (m *MigrationService) SeedData() error {
    m.logger.Info("Seeding initial data...")

    // 创建默认超级管理员
    if err := m.createDefaultSuperAdmin(); err != nil {
        return fmt.Errorf("创建默认超级管理员失败: %v", err)
    }

    // 初始化默认菜单（仅在菜单表为空时执行）
    menuService := services.NewMenuService(m.db, m.logger)
    if err := menuService.SeedDefaultMenus(); err != nil {
        return fmt.Errorf("初始化默认菜单失败: %v", err)
    }

    m.logger.Info("Initial data seeded successfully")
    return nil
}

// createDefaultSuperAdmin 创建默认超级管理员
// 如果已存在超级管理员，跳过创建
func (m *MigrationService) createDefaultSuperAdmin() error {
    var count int64
    if err := m.db.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count).Error; err != nil {
        return fmt.Errorf("检查超级管理员失败: %v", err)
    }

	// 如果已存在超级管理员，跳过创建
	if count > 0 {
		m.logger.Info("Super admin user already exists, skipping creation")
		return nil
	}

    // 如果存在用户名为admin的用户，则优先提升其角色为SUPER_ADMIN，避免唯一索引冲突
    var existingAdmin models.User
    if err := m.db.Where("username = ?", "admin").First(&existingAdmin).Error; err == nil {
        if existingAdmin.Role != models.RoleSuperAdmin {
            existingAdmin.Role = models.RoleSuperAdmin
            existingAdmin.Level = 100
            if err := m.db.Save(&existingAdmin).Error; err != nil {
                return fmt.Errorf("提升现有admin为超级管理员失败: %v", err)
            }
            m.logger.Info("Existing admin promoted to super admin", zap.String("username", existingAdmin.Username))
        }
        return nil
    }

    // 创建默认超级管理员（使用bcrypt加密密码）
    hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("加密默认超级管理员密码失败: %v", err)
    }

    superAdmin := &models.User{
        Username:    "admin",
        Email:       "admin@example.com",
        Password:    string(hashed),
        DisplayName: "",
        Role:        models.RoleSuperAdmin,
        Level:       100,
        IsActive:    true,
    }

	if err := m.db.Create(superAdmin).Error; err != nil {
		return fmt.Errorf("创建默认超级管理员失败: %v", err)
	}

	m.logger.Info("Default super admin user created", zap.String("username", superAdmin.Username))
	return nil
}

// RunMigration 运行数据库迁移
func (m *MigrationService) RunMigration() error {
	// 1. 自动迁移数据库表
	if err := m.AutoMigrate(); err != nil {
		return err
	}

	// 2. 创建数据库索引
	if err := m.CreateIndexes(); err != nil {
		return err
	}

	// 3. 初始化数据库种子数据
	if err := m.SeedData(); err != nil {
		return err
	}

	m.logger.Info("Complete database migration finished successfully")
	return nil
}
