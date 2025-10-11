package tests

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"taishanglaojun/core-services/multi-tenancy/models"
	"taishanglaojun/core-services/multi-tenancy/repositories"
	"taishanglaojun/core-services/multi-tenancy/services"
	"taishanglaojun/pkg/logger"
)

// TenantTestSuite 租户测试套件
type TenantTestSuite struct {
	suite.Suite
	db            *gorm.DB
	redis         *redis.Client
	logger        logger.Logger
	tenantRepo    repositories.TenantRepositoryInterface
	tenantUserRepo repositories.TenantUserRepositoryInterface
	tenantService services.TenantServiceInterface
	ctx           context.Context
}

// SetupSuite 设置测试套件
func (suite *TenantTestSuite) SetupSuite() {
	// 设置内存数据�?
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 设置Redis（使用内存模拟）
	suite.redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用测试数据�?
	})

	// 设置日志
	suite.logger = logger.NewLogger("test", "debug")

	// 迁移数据�?
	err = suite.db.AutoMigrate(
		&models.Tenant{},
		&models.TenantUser{},
		&models.TenantSubscription{},
	)
	suite.Require().NoError(err)

	// 初始化仓�?
	suite.tenantRepo = repositories.NewTenantRepository(suite.db, suite.logger)
	suite.tenantUserRepo = repositories.NewTenantUserRepository(suite.db, suite.logger)

	// 初始化服�?
	suite.tenantService = services.NewTenantService(
		suite.tenantRepo,
		suite.tenantUserRepo,
		suite.redis,
		suite.logger,
		"row_level",
	)

	suite.ctx = context.Background()
}

// TearDownSuite 清理测试套件
func (suite *TenantTestSuite) TearDownSuite() {
	// 清理Redis
	suite.redis.FlushDB(suite.ctx)
	suite.redis.Close()
}

// SetupTest 设置每个测试
func (suite *TenantTestSuite) SetupTest() {
	// 清理数据�?
	suite.db.Exec("DELETE FROM tenant_users")
	suite.db.Exec("DELETE FROM tenant_subscriptions")
	suite.db.Exec("DELETE FROM tenants")
}

// TestCreateTenant 测试创建租户
func (suite *TenantTestSuite) TestCreateTenant() {
	req := &models.CreateTenantRequest{
		Name:        "测试租户",
		Subdomain:   "test",
		Description: "这是一个测试租�?,
		Settings: models.TenantSettings{
			Language:   "zh-CN",
			Timezone:   "Asia/Shanghai",
			DateFormat: "YYYY-MM-DD",
			TimeFormat: "24h",
			Currency:   "CNY",
		},
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)
	suite.NotNil(tenant)
	suite.Equal("测试租户", tenant.Name)
	suite.Equal("test", tenant.Subdomain)
	suite.Equal(models.TenantStatusActive, tenant.Status)
	suite.Equal("zh-CN", tenant.Settings.Language)
}

// TestCreateTenantWithDuplicateSubdomain 测试创建重复子域名的租户
func (suite *TenantTestSuite) TestCreateTenantWithDuplicateSubdomain() {
	// 创建第一个租�?
	req1 := &models.CreateTenantRequest{
		Name:        "租户1",
		Subdomain:   "duplicate",
		Description: "第一个租�?,
		OwnerUserID: 1,
	}

	_, err := suite.tenantService.CreateTenant(suite.ctx, req1)
	suite.NoError(err)

	// 尝试创建相同子域名的租户
	req2 := &models.CreateTenantRequest{
		Name:        "租户2",
		Subdomain:   "duplicate",
		Description: "第二个租�?,
		OwnerUserID: 2,
	}

	_, err = suite.tenantService.CreateTenant(suite.ctx, req2)
	suite.Error(err)
	suite.Contains(err.Error(), "subdomain already exists")
}

// TestGetTenant 测试获取租户
func (suite *TenantTestSuite) TestGetTenant() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "获取测试租户",
		Subdomain:   "gettest",
		Description: "用于测试获取功能",
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 获取租户
	tenant, err := suite.tenantService.GetTenant(suite.ctx, createdTenant.ID)
	suite.NoError(err)
	suite.NotNil(tenant)
	suite.Equal(createdTenant.ID, tenant.ID)
	suite.Equal("获取测试租户", tenant.Name)
}

// TestGetTenantBySubdomain 测试通过子域名获取租�?
func (suite *TenantTestSuite) TestGetTenantBySubdomain() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "子域名测试租�?,
		Subdomain:   "subdomaintest",
		Description: "用于测试子域名获取功�?,
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 通过子域名获取租�?
	tenant, err := suite.tenantService.GetTenantBySubdomain(suite.ctx, "subdomaintest")
	suite.NoError(err)
	suite.NotNil(tenant)
	suite.Equal(createdTenant.ID, tenant.ID)
	suite.Equal("subdomaintest", tenant.Subdomain)
}

// TestUpdateTenant 测试更新租户
func (suite *TenantTestSuite) TestUpdateTenant() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "更新测试租户",
		Subdomain:   "updatetest",
		Description: "用于测试更新功能",
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 更新租户
	updateReq := &models.UpdateTenantRequest{
		Name:        "更新后的租户名称",
		Description: "更新后的描述",
		Settings: &models.TenantSettings{
			Language:   "en-US",
			Timezone:   "UTC",
			DateFormat: "MM/DD/YYYY",
			TimeFormat: "12h",
			Currency:   "USD",
		},
	}

	updatedTenant, err := suite.tenantService.UpdateTenant(suite.ctx, createdTenant.ID, updateReq)
	suite.NoError(err)
	suite.NotNil(updatedTenant)
	suite.Equal("更新后的租户名称", updatedTenant.Name)
	suite.Equal("更新后的描述", updatedTenant.Description)
	suite.Equal("en-US", updatedTenant.Settings.Language)
}

// TestDeleteTenant 测试删除租户
func (suite *TenantTestSuite) TestDeleteTenant() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "删除测试租户",
		Subdomain:   "deletetest",
		Description: "用于测试删除功能",
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 删除租户
	err = suite.tenantService.DeleteTenant(suite.ctx, createdTenant.ID)
	suite.NoError(err)

	// 验证租户已被删除
	_, err = suite.tenantService.GetTenant(suite.ctx, createdTenant.ID)
	suite.Error(err)
}

// TestListTenants 测试列出租户
func (suite *TenantTestSuite) TestListTenants() {
	// 创建多个租户
	for i := 1; i <= 5; i++ {
		req := &models.CreateTenantRequest{
			Name:        fmt.Sprintf("租户%d", i),
			Subdomain:   fmt.Sprintf("tenant%d", i),
			Description: fmt.Sprintf("�?d个租�?, i),
			OwnerUserID: uint(i),
		}

		_, err := suite.tenantService.CreateTenant(suite.ctx, req)
		suite.NoError(err)
	}

	// 列出租户
	query := &models.TenantQuery{
		Page:     1,
		PageSize: 10,
	}

	tenants, total, err := suite.tenantService.ListTenants(suite.ctx, query)
	suite.NoError(err)
	suite.Len(tenants, 5)
	suite.Equal(int64(5), total)
}

// TestAddTenantUser 测试添加租户用户
func (suite *TenantTestSuite) TestAddTenantUser() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "用户测试租户",
		Subdomain:   "usertest",
		Description: "用于测试用户功能",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 添加租户用户
	userReq := &models.AddTenantUserRequest{
		UserID: 2,
		Role:   "admin",
	}

	tenantUser, err := suite.tenantService.AddTenantUser(suite.ctx, tenant.ID, userReq)
	suite.NoError(err)
	suite.NotNil(tenantUser)
	suite.Equal(uint(2), tenantUser.UserID)
	suite.Equal("admin", tenantUser.Role)
}

// TestRemoveTenantUser 测试移除租户用户
func (suite *TenantTestSuite) TestRemoveTenantUser() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "移除用户测试租户",
		Subdomain:   "removeusertest",
		Description: "用于测试移除用户功能",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 添加租户用户
	userReq := &models.AddTenantUserRequest{
		UserID: 2,
		Role:   "member",
	}

	_, err = suite.tenantService.AddTenantUser(suite.ctx, tenant.ID, userReq)
	suite.NoError(err)

	// 移除租户用户
	err = suite.tenantService.RemoveTenantUser(suite.ctx, tenant.ID, 2)
	suite.NoError(err)

	// 验证用户已被移除
	_, err = suite.tenantService.GetTenantUser(suite.ctx, tenant.ID, 2)
	suite.Error(err)
}

// TestUpdateQuota 测试更新配额
func (suite *TenantTestSuite) TestUpdateQuota() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "配额测试租户",
		Subdomain:   "quotatest",
		Description: "用于测试配额功能",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 更新配额
	quotaReq := &models.UpdateTenantQuotaRequest{
		MaxUsers:       200,
		MaxStorage:     20 * 1024 * 1024 * 1024, // 20GB
		MaxAPIRequests: 20000,
		MaxBandwidth:   2 * 1024 * 1024 * 1024, // 2GB
	}

	quota, err := suite.tenantService.UpdateQuota(suite.ctx, tenant.ID, quotaReq)
	suite.NoError(err)
	suite.NotNil(quota)
	suite.Equal(int64(200), quota.MaxUsers)
	suite.Equal(int64(20*1024*1024*1024), quota.MaxStorage)
}

// TestCheckQuota 测试检查配�?
func (suite *TenantTestSuite) TestCheckQuota() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "配额检查测试租�?,
		Subdomain:   "quotachecktest",
		Description: "用于测试配额检查功�?,
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 检查配额（应该通过�?
	canProceed, err := suite.tenantService.CheckQuota(suite.ctx, tenant.ID, "users", 1)
	suite.NoError(err)
	suite.True(canProceed)

	// 检查超出配额的情况
	canProceed, err = suite.tenantService.CheckQuota(suite.ctx, tenant.ID, "users", 1000)
	suite.NoError(err)
	suite.False(canProceed)
}

// TestActivateAndSuspendTenant 测试激活和暂停租户
func (suite *TenantTestSuite) TestActivateAndSuspendTenant() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "状态测试租�?,
		Subdomain:   "statustest",
		Description: "用于测试状态功�?,
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)
	suite.Equal(models.TenantStatusActive, tenant.Status)

	// 暂停租户
	err = suite.tenantService.SuspendTenant(suite.ctx, tenant.ID, "测试暂停")
	suite.NoError(err)

	// 验证状�?
	updatedTenant, err := suite.tenantService.GetTenant(suite.ctx, tenant.ID)
	suite.NoError(err)
	suite.Equal(models.TenantStatusSuspended, updatedTenant.Status)

	// 激活租�?
	err = suite.tenantService.ActivateTenant(suite.ctx, tenant.ID)
	suite.NoError(err)

	// 验证状�?
	activatedTenant, err := suite.tenantService.GetTenant(suite.ctx, tenant.ID)
	suite.NoError(err)
	suite.Equal(models.TenantStatusActive, activatedTenant.Status)
}

// TestValidateUserAccess 测试验证用户访问权限
func (suite *TenantTestSuite) TestValidateUserAccess() {
	// 创建租户
	req := &models.CreateTenantRequest{
		Name:        "访问权限测试租户",
		Subdomain:   "accesstest",
		Description: "用于测试访问权限功能",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 验证所有者访问权�?
	hasAccess, err := suite.tenantService.ValidateUserAccess(suite.ctx, tenant.ID, 1)
	suite.NoError(err)
	suite.True(hasAccess)

	// 验证非成员访问权�?
	hasAccess, err = suite.tenantService.ValidateUserAccess(suite.ctx, tenant.ID, 999)
	suite.NoError(err)
	suite.False(hasAccess)

	// 添加用户并验证访问权�?
	userReq := &models.AddTenantUserRequest{
		UserID: 2,
		Role:   "member",
	}

	_, err = suite.tenantService.AddTenantUser(suite.ctx, tenant.ID, userReq)
	suite.NoError(err)

	hasAccess, err = suite.tenantService.ValidateUserAccess(suite.ctx, tenant.ID, 2)
	suite.NoError(err)
	suite.True(hasAccess)
}

// BenchmarkCreateTenant 基准测试创建租户
func BenchmarkCreateTenant(b *testing.B) {
	// 设置测试环境
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	redis := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 1})
	logger := logger.NewLogger("benchmark", "error")

	db.AutoMigrate(&models.Tenant{}, &models.TenantUser{}, &models.TenantSubscription{})

	tenantRepo := repositories.NewTenantRepository(db, logger)
	tenantUserRepo := repositories.NewTenantUserRepository(db, logger)
	tenantService := services.NewTenantService(tenantRepo, tenantUserRepo, redis, logger, "row_level")

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &models.CreateTenantRequest{
			Name:        fmt.Sprintf("基准测试租户%d", i),
			Subdomain:   fmt.Sprintf("benchmark%d", i),
			Description: "基准测试租户",
			OwnerUserID: uint(i + 1),
		}

		_, err := tenantService.CreateTenant(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetTenant 基准测试获取租户
func BenchmarkGetTenant(b *testing.B) {
	// 设置测试环境
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	redis := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 1})
	logger := logger.NewLogger("benchmark", "error")

	db.AutoMigrate(&models.Tenant{}, &models.TenantUser{}, &models.TenantSubscription{})

	tenantRepo := repositories.NewTenantRepository(db, logger)
	tenantUserRepo := repositories.NewTenantUserRepository(db, logger)
	tenantService := services.NewTenantService(tenantRepo, tenantUserRepo, redis, logger, "row_level")

	ctx := context.Background()

	// 创建测试租户
	req := &models.CreateTenantRequest{
		Name:        "基准测试租户",
		Subdomain:   "benchmarkget",
		Description: "用于基准测试",
		OwnerUserID: 1,
	}

	tenant, _ := tenantService.CreateTenant(ctx, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tenantService.GetTenant(ctx, tenant.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestMain 测试主函�?
func TestMain(m *testing.M) {
	// 运行测试
	m.Run()
}

// 运行测试套件
func TestTenantTestSuite(t *testing.T) {
	suite.Run(t, new(TenantTestSuite))
}
