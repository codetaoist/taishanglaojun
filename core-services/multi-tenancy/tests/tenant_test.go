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

	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/models"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/services"
	"taishanglaojun/pkg/logger"
)

// TenantTestSuite 
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

// SetupSuite 
func (suite *TenantTestSuite) SetupSuite() {
	// ?
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// Redis
	suite.redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // ?
	})

	// 
	suite.logger = logger.NewLogger("test", "debug")

	// ?
	err = suite.db.AutoMigrate(
		&models.Tenant{},
		&models.TenantUser{},
		&models.TenantSubscription{},
	)
	suite.Require().NoError(err)

	// ?
	suite.tenantRepo = repositories.NewTenantRepository(suite.db, suite.logger)
	suite.tenantUserRepo = repositories.NewTenantUserRepository(suite.db, suite.logger)

	// ?
	suite.tenantService = services.NewTenantService(
		suite.tenantRepo,
		suite.tenantUserRepo,
		suite.redis,
		suite.logger,
		"row_level",
	)

	suite.ctx = context.Background()
}

// TearDownSuite 
func (suite *TenantTestSuite) TearDownSuite() {
	// Redis
	suite.redis.FlushDB(suite.ctx)
	suite.redis.Close()
}

// SetupTest 
func (suite *TenantTestSuite) SetupTest() {
	// ?
	suite.db.Exec("DELETE FROM tenant_users")
	suite.db.Exec("DELETE FROM tenant_subscriptions")
	suite.db.Exec("DELETE FROM tenants")
}

// TestCreateTenant 
func (suite *TenantTestSuite) TestCreateTenant() {
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "test",
		Description: "?,
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
	suite.Equal("", tenant.Name)
	suite.Equal("test", tenant.Subdomain)
	suite.Equal(models.TenantStatusActive, tenant.Status)
	suite.Equal("zh-CN", tenant.Settings.Language)
}

// TestCreateTenantWithDuplicateSubdomain 
func (suite *TenantTestSuite) TestCreateTenantWithDuplicateSubdomain() {
	// ?
	req1 := &models.CreateTenantRequest{
		Name:        "1",
		Subdomain:   "duplicate",
		Description: "?,
		OwnerUserID: 1,
	}

	_, err := suite.tenantService.CreateTenant(suite.ctx, req1)
	suite.NoError(err)

	// 
	req2 := &models.CreateTenantRequest{
		Name:        "2",
		Subdomain:   "duplicate",
		Description: "?,
		OwnerUserID: 2,
	}

	_, err = suite.tenantService.CreateTenant(suite.ctx, req2)
	suite.Error(err)
	suite.Contains(err.Error(), "subdomain already exists")
}

// TestGetTenant 
func (suite *TenantTestSuite) TestGetTenant() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "gettest",
		Description: "",
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 
	tenant, err := suite.tenantService.GetTenant(suite.ctx, createdTenant.ID)
	suite.NoError(err)
	suite.NotNil(tenant)
	suite.Equal(createdTenant.ID, tenant.ID)
	suite.Equal("", tenant.Name)
}

// TestGetTenantBySubdomain ?
func (suite *TenantTestSuite) TestGetTenantBySubdomain() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "?,
		Subdomain:   "subdomaintest",
		Description: "?,
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// ?
	tenant, err := suite.tenantService.GetTenantBySubdomain(suite.ctx, "subdomaintest")
	suite.NoError(err)
	suite.NotNil(tenant)
	suite.Equal(createdTenant.ID, tenant.ID)
	suite.Equal("subdomaintest", tenant.Subdomain)
}

// TestUpdateTenant 
func (suite *TenantTestSuite) TestUpdateTenant() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "updatetest",
		Description: "",
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 
	updateReq := &models.UpdateTenantRequest{
		Name:        "",
		Description: "",
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
	suite.Equal("", updatedTenant.Name)
	suite.Equal("", updatedTenant.Description)
	suite.Equal("en-US", updatedTenant.Settings.Language)
}

// TestDeleteTenant 
func (suite *TenantTestSuite) TestDeleteTenant() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "deletetest",
		Description: "",
		OwnerUserID: 1,
	}

	createdTenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 
	err = suite.tenantService.DeleteTenant(suite.ctx, createdTenant.ID)
	suite.NoError(err)

	// 
	_, err = suite.tenantService.GetTenant(suite.ctx, createdTenant.ID)
	suite.Error(err)
}

// TestListTenants 
func (suite *TenantTestSuite) TestListTenants() {
	// 
	for i := 1; i <= 5; i++ {
		req := &models.CreateTenantRequest{
			Name:        fmt.Sprintf("%d", i),
			Subdomain:   fmt.Sprintf("tenant%d", i),
			Description: fmt.Sprintf("?d?, i),
			OwnerUserID: uint(i),
		}

		_, err := suite.tenantService.CreateTenant(suite.ctx, req)
		suite.NoError(err)
	}

	// 
	query := &models.TenantQuery{
		Page:     1,
		PageSize: 10,
	}

	tenants, total, err := suite.tenantService.ListTenants(suite.ctx, query)
	suite.NoError(err)
	suite.Len(tenants, 5)
	suite.Equal(int64(5), total)
}

// TestAddTenantUser 
func (suite *TenantTestSuite) TestAddTenantUser() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "usertest",
		Description: "",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 
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

// TestRemoveTenantUser 
func (suite *TenantTestSuite) TestRemoveTenantUser() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "removeusertest",
		Description: "",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 
	userReq := &models.AddTenantUserRequest{
		UserID: 2,
		Role:   "member",
	}

	_, err = suite.tenantService.AddTenantUser(suite.ctx, tenant.ID, userReq)
	suite.NoError(err)

	// 
	err = suite.tenantService.RemoveTenantUser(suite.ctx, tenant.ID, 2)
	suite.NoError(err)

	// 
	_, err = suite.tenantService.GetTenantUser(suite.ctx, tenant.ID, 2)
	suite.Error(err)
}

// TestUpdateQuota 
func (suite *TenantTestSuite) TestUpdateQuota() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "quotatest",
		Description: "",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// 
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

// TestCheckQuota ?
func (suite *TenantTestSuite) TestCheckQuota() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "?,
		Subdomain:   "quotachecktest",
		Description: "鹦?,
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// ?
	canProceed, err := suite.tenantService.CheckQuota(suite.ctx, tenant.ID, "users", 1)
	suite.NoError(err)
	suite.True(canProceed)

	// 鳬
	canProceed, err = suite.tenantService.CheckQuota(suite.ctx, tenant.ID, "users", 1000)
	suite.NoError(err)
	suite.False(canProceed)
}

// TestActivateAndSuspendTenant 
func (suite *TenantTestSuite) TestActivateAndSuspendTenant() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "?,
		Subdomain:   "statustest",
		Description: "?,
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)
	suite.Equal(models.TenantStatusActive, tenant.Status)

	// 
	err = suite.tenantService.SuspendTenant(suite.ctx, tenant.ID, "")
	suite.NoError(err)

	// ?
	updatedTenant, err := suite.tenantService.GetTenant(suite.ctx, tenant.ID)
	suite.NoError(err)
	suite.Equal(models.TenantStatusSuspended, updatedTenant.Status)

	// ?
	err = suite.tenantService.ActivateTenant(suite.ctx, tenant.ID)
	suite.NoError(err)

	// ?
	activatedTenant, err := suite.tenantService.GetTenant(suite.ctx, tenant.ID)
	suite.NoError(err)
	suite.Equal(models.TenantStatusActive, activatedTenant.Status)
}

// TestValidateUserAccess 
func (suite *TenantTestSuite) TestValidateUserAccess() {
	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "accesstest",
		Description: "",
		OwnerUserID: 1,
	}

	tenant, err := suite.tenantService.CreateTenant(suite.ctx, req)
	suite.NoError(err)

	// ?
	hasAccess, err := suite.tenantService.ValidateUserAccess(suite.ctx, tenant.ID, 1)
	suite.NoError(err)
	suite.True(hasAccess)

	// ?
	hasAccess, err = suite.tenantService.ValidateUserAccess(suite.ctx, tenant.ID, 999)
	suite.NoError(err)
	suite.False(hasAccess)

	// ?
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

// BenchmarkCreateTenant 
func BenchmarkCreateTenant(b *testing.B) {
	// 
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
			Name:        fmt.Sprintf("%d", i),
			Subdomain:   fmt.Sprintf("benchmark%d", i),
			Description: "",
			OwnerUserID: uint(i + 1),
		}

		_, err := tenantService.CreateTenant(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetTenant 
func BenchmarkGetTenant(b *testing.B) {
	// 
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	redis := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 1})
	logger := logger.NewLogger("benchmark", "error")

	db.AutoMigrate(&models.Tenant{}, &models.TenantUser{}, &models.TenantSubscription{})

	tenantRepo := repositories.NewTenantRepository(db, logger)
	tenantUserRepo := repositories.NewTenantUserRepository(db, logger)
	tenantService := services.NewTenantService(tenantRepo, tenantUserRepo, redis, logger, "row_level")

	ctx := context.Background()

	// 
	req := &models.CreateTenantRequest{
		Name:        "",
		Subdomain:   "benchmarkget",
		Description: "",
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

// TestMain ?
func TestMain(m *testing.M) {
	// 
	m.Run()
}

// 
func TestTenantTestSuite(t *testing.T) {
	suite.Run(t, new(TenantTestSuite))
}

