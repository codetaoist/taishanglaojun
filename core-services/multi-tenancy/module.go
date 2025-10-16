package multitenancy

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/internal/config"
	"github.com/codetaoist/taishanglaojun/core-services/internal/logger"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/middleware"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/models"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/routes"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/services"
)

// Module 多租户模块
type Module struct {
	// 数据库连接
	db     *gorm.DB
	redis  *redis.Client
	logger logger.Logger
	config *config.Config

	// 租户仓库
	tenantRepo     repositories.TenantRepositoryInterface
	tenantUserRepo repositories.TenantUserRepositoryInterface

	// 租户服务
	tenantService services.TenantServiceInterface

	// 租户处理程序
	tenantHandler *handlers.TenantHandler

	// 租户中间件
	tenantMiddleware *middleware.TenantMiddleware

	// 租户路由
	tenantRoutes *routes.TenantRoutes

	// 多租户配置
	multiTenancyConfig *MultiTenancyConfig
}

// MultiTenancyConfig 多租户配置
type MultiTenancyConfig struct {
	// 隔离策略	row_level, schema, database
	IsolationStrategy string `yaml:"isolation_strategy" json:"isolation_strategy"` // row_level, schema, database

	// 默认配额
	DefaultQuota models.TenantQuota `yaml:"default_quota" json:"default_quota"`

	// 缓存配置
	Cache struct {
		TTL                time.Duration `yaml:"ttl" json:"ttl"`
		TenantCachePrefix  string        `yaml:"tenant_cache_prefix" json:"tenant_cache_prefix"`
		SessionCachePrefix string        `yaml:"session_cache_prefix" json:"session_cache_prefix"`
	} `yaml:"cache" json:"cache"`

	// 安全配置
	Security struct {
		EnableSubdomainValidation bool     `yaml:"enable_subdomain_validation" json:"enable_subdomain_validation"`
		AllowedDomains            []string `yaml:"allowed_domains" json:"allowed_domains"`
		RequireHTTPS              bool     `yaml:"require_https" json:"require_https"`
	} `yaml:"security" json:"security"`

	// 监控配置
	Monitoring struct {
		EnableMetrics     bool          `yaml:"enable_metrics" json:"enable_metrics"`
		MetricsInterval   time.Duration `yaml:"metrics_interval" json:"metrics_interval"`
		EnableHealthCheck bool          `yaml:"enable_health_check" json:"enable_health_check"`
	} `yaml:"monitoring" json:"monitoring"`

	// 数据库配置
	Database struct {
		MaxConnections    int           `yaml:"max_connections" json:"max_connections"`
		ConnectionTimeout time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
		EnableAutoMigrate bool          `yaml:"enable_auto_migrate" json:"enable_auto_migrate"`
	} `yaml:"database" json:"database"`
}

// NewModule 创建多租户模块
func NewModule(db *gorm.DB, redis *redis.Client, logger logger.Logger, config *config.Config) *Module {
	return &Module{
		db:     db,
		redis:  redis,
		logger: logger,
		config: config,
	}
}

// Initialize 初始化多租户模块
func (m *Module) Initialize() error {
	m.logger.Info("Initializing multi-tenancy module...")

	// 加载多租户配置
	if err := m.loadConfig(); err != nil {
		return fmt.Errorf("failed to load multi-tenancy config: %w", err)
	}

	// 初始化数据库
	if err := m.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// 初始化仓库
	m.initRepositories()

	// 初始化服务
	m.initServices()

	// 初始化处理程序
	m.initHandlers()

	// 初始化中间件
	m.initMiddleware()

	// 初始化路由
	m.initRoutes()

	m.logger.Info("Multi-tenancy module initialized successfully")
	return nil
}

// loadConfig 加载多租户配置
func (m *Module) loadConfig() error {
	//
	m.multiTenancyConfig = &MultiTenancyConfig{
		IsolationStrategy: "row_level", //
		DefaultQuota: models.TenantQuota{
			MaxUsers:       100,
			MaxStorage:     10 * 1024 * 1024 * 1024, // 10GB
			MaxAPIRequests: 10000,
			MaxBandwidth:   1024 * 1024 * 1024, // 1GB
			MaxDatabases:   5,
			MaxConnections: 50,
			MaxFileSize:    100 * 1024 * 1024, // 100MB
			MaxFiles:       1000,
			MaxSessions:    100,
			MaxProjects:    10,
		},
	}

	// 缓存配置
	m.multiTenancyConfig.Cache.TTL = 30 * time.Minute
	m.multiTenancyConfig.Cache.TenantCachePrefix = "tenant:"
	m.multiTenancyConfig.Cache.SessionCachePrefix = "tenant_session:"

	// 安全配置
	m.multiTenancyConfig.Security.EnableSubdomainValidation = true
	m.multiTenancyConfig.Security.AllowedDomains = []string{"localhost", "taishanglaojun.com"}
	m.multiTenancyConfig.Security.RequireHTTPS = false // false

	// 监控配置
	m.multiTenancyConfig.Monitoring.EnableMetrics = true
	m.multiTenancyConfig.Monitoring.MetricsInterval = 5 * time.Minute
	m.multiTenancyConfig.Monitoring.EnableHealthCheck = true

	// 数据库配置
	m.multiTenancyConfig.Database.MaxConnections = 100
	m.multiTenancyConfig.Database.ConnectionTimeout = 30 * time.Second
	m.multiTenancyConfig.Database.EnableAutoMigrate = true

	// TODO: 从配置文件加载多租户配置
	// if err := m.config.UnmarshalKey("multi_tenancy", m.multiTenancyConfig); err != nil {
	//     m.logger.Warn("Failed to load multi-tenancy config from file, using defaults", "error", err)
	// }

	return nil
}

// initDatabase 初始化数据库
// 启用自动迁移时，会执行数据库迁移操作
func (m *Module) initDatabase() error {
	if !m.multiTenancyConfig.Database.EnableAutoMigrate {
		return nil
	}

	m.logger.Info("Running database migrations for multi-tenancy...")

	// 执行数据库迁移操作
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	m.logger.Info("Database migrations completed successfully")
	return nil
}

// migrateDatabase 执行数据库迁移操作
func (m *Module) migrateDatabase() error {
	// 自动迁移数据库表
	return m.db.AutoMigrate(
		&models.Tenant{},
		&models.TenantUser{},
		&models.TenantSubscription{},
	)
}

// initRepositories 初始化仓库
func (m *Module) initRepositories() {
	m.logger.Info("Initializing repositories...")

	m.tenantRepo = repositories.NewTenantRepository(m.db, m.logger)
	m.tenantUserRepo = repositories.NewTenantUserRepository(m.db, m.logger)

	m.logger.Info("Repositories initialized successfully")
}

// initServices 初始化服务
func (m *Module) initServices() {
	m.logger.Info("Initializing services...")

	m.tenantService = services.NewTenantService(
		m.tenantRepo,
		m.tenantUserRepo,
		m.db,
		m.redis,
		m.logger,
		m.multiTenancyConfig.IsolationStrategy,
	)

	m.logger.Info("Services initialized successfully")
}

// initHandlers 初始化处理程序
func (m *Module) initHandlers() {
	m.logger.Info("Initializing handlers...")

	m.tenantHandler = handlers.NewTenantHandler(m.tenantService, m.logger, m.redis)

	m.logger.Info("Handlers initialized successfully")
}

// initMiddleware 初始化中间件
func (m *Module) initMiddleware() {
	m.logger.Info("Initializing middleware...")

	m.tenantMiddleware = middleware.NewTenantMiddleware(m.tenantService, m.logger, m.redis)

	m.logger.Info("Middleware initialized successfully")
}

// initRoutes 初始化路由
func (m *Module) initRoutes() {
	m.logger.Info("Initializing routes...")

	m.tenantRoutes = routes.NewTenantRoutes(m.tenantHandler)

	m.logger.Info("Routes initialized successfully")
}

// SetupRoutes 设置多租户路由
func (m *Module) SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	m.logger.Info("Setting up multi-tenancy routes...")

	// 挂载租户路由
	m.tenantRoutes.SetupAllRoutes(router, authMiddleware)

	m.logger.Info("Multi-tenancy routes setup completed")
}

// GetTenantMiddleware 获取租户中间件
func (m *Module) GetTenantMiddleware() *middleware.TenantMiddleware {
	return m.tenantMiddleware
}

// GetTenantService 获取租户服务
func (m *Module) GetTenantService() services.TenantServiceInterface {
	return m.tenantService
}

// GetTenantHandler 获取租户处理程序
func (m *Module) GetTenantHandler() *handlers.TenantHandler {
	return m.tenantHandler
}

// Health 检查多租户模块健康状态
func (m *Module) Health(ctx context.Context) error {
	// 数据库健康检查
	sqlDB, err := m.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Redis健康检查
	if err := m.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// 服务健康检查
	if m.tenantService == nil {
		return fmt.Errorf("tenant service not initialized")
	}

	return nil
}

// Shutdown 关闭多租户模块
func (m *Module) Shutdown(ctx context.Context) error {
	m.logger.Info("Shutting down multi-tenancy module...")

	// 关闭数据库连接
	if err := sqlDB.Close(); err != nil {
		m.logger.Error("Failed to close database connection", "error", err)
	}

	// 关闭Redis连接
	if err := m.redis.Close(); err != nil {
		m.logger.Error("Failed to close Redis connection", "error", err)
	}

	m.logger.Info("Multi-tenancy module shutdown completed")
	return nil
}

// GetConfig 获取多租户配置
func (m *Module) GetConfig() *MultiTenancyConfig {
	return m.multiTenancyConfig
}

// CreateDefaultTenant 创建默认租户
// 初始化时创建默认租户，用于处理单租户场景
func (m *Module) CreateDefaultTenant(ctx context.Context, name, subdomain string, ownerUserID uint) (*models.Tenant, error) {
	m.logger.Info("Creating default tenant", "name", name, "subdomain", subdomain)

	// 创建默认租户请求
	req := &models.CreateTenantRequest{
		Name:        name,
		Subdomain:   subdomain,
		Description: "Default tenant created during initialization",
		Settings: &models.TenantSettings{
			Language:       "zh-CN",
			Timezone:       "Asia/Shanghai",
			DateFormat:     "YYYY-MM-DD",
			TimeFormat:     "24h",
			Currency:       "CNY",
			EnableFeatureA: true,
			EnableFeatureB: true,
			EnableFeatureC: false,
		},
		OwnerUserID: ownerUserID,
	}

	// 创建默认租户
	tenant, err := m.tenantService.CreateTenant(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create default tenant: %w", err)
	}

	m.logger.Info("Default tenant created successfully", "tenant_id", tenant.ID, "name", tenant.Name)
	return tenant, nil
}

// ValidateConfiguration 验证多租户配置
// 检查配置是否符合要求，如隔离策略、默认配额等
func (m *Module) ValidateConfiguration() error {
	if m.multiTenancyConfig == nil {
		return fmt.Errorf("multi-tenancy configuration not loaded")
	}

	// 验证隔离策略
	validStrategies := []string{"row_level", "schema", "database"}
	isValid := false
	for _, strategy := range validStrategies {
		if m.multiTenancyConfig.IsolationStrategy == strategy {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid isolation strategy: %s", m.multiTenancyConfig.IsolationStrategy)
	}

	// 验证默认配额
	if m.multiTenancyConfig.DefaultQuota.MaxUsers <= 0 {
		return fmt.Errorf("default quota max_users must be greater than 0")
	}

	if m.multiTenancyConfig.DefaultQuota.MaxStorage <= 0 {
		return fmt.Errorf("default quota max_storage must be greater than 0")
	}

	return nil
}

// GetMetrics 获取多租户指标
// 包括租户总数、活跃租户数、暂停租户数等
func (m *Module) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 查询租户总数
	var totalTenants int64
	if err := m.db.WithContext(ctx).Model(&models.Tenant{}).Count(&totalTenants).Error; err != nil {
		return nil, fmt.Errorf("failed to count tenants: %w", err)
	}
	metrics["total_tenants"] = totalTenants

	// 查询活跃租户数
	var activeTenants int64
	if err := m.db.WithContext(ctx).Model(&models.Tenant{}).Where("status = ?", models.TenantStatusActive).Count(&activeTenants).Error; err != nil {
		return nil, fmt.Errorf("failed to count active tenants: %w", err)
	}
	metrics["active_tenants"] = activeTenants

	// 查询暂停租户数
	var suspendedTenants int64
	if err := m.db.WithContext(ctx).Model(&models.Tenant{}).Where("status = ?", models.TenantStatusSuspended).Count(&suspendedTenants).Error; err != nil {
		return nil, fmt.Errorf("failed to count suspended tenants: %w", err)
	}
	metrics["suspended_tenants"] = suspendedTenants

	return metrics, nil
}
