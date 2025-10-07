package routes

import (
	"github.com/gin-gonic/gin"

	"taishanglaojun/core-services/multi-tenancy/handlers"
	"taishanglaojun/core-services/multi-tenancy/middleware"
)

// TenantRoutes 租户路由配置
type TenantRoutes struct {
	tenantHandler *handlers.TenantHandler
}

// NewTenantRoutes 创建租户路由配置
func NewTenantRoutes(tenantHandler *handlers.TenantHandler) *TenantRoutes {
	return &TenantRoutes{
		tenantHandler: tenantHandler,
	}
}

// SetupRoutes 设置租户相关路由
func (r *TenantRoutes) SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// API版本组
	v1 := router.Group("/api/v1")
	
	// 租户管理路由组 - 需要认证
	tenants := v1.Group("/tenants")
	tenants.Use(authMiddleware)
	tenants.Use(middleware.TenantPermissionMiddleware()) // 租户权限中间件
	{
		// 租户基本操作
		tenants.POST("", r.tenantHandler.CreateTenant)           // 创建租户
		tenants.GET("", r.tenantHandler.ListTenants)             // 列出租户
		tenants.GET("/:id", r.tenantHandler.GetTenant)           // 获取租户详情
		tenants.PUT("/:id", r.tenantHandler.UpdateTenant)        // 更新租户
		tenants.DELETE("/:id", r.tenantHandler.DeleteTenant)     // 删除租户
		
		// 租户查找
		tenants.GET("/subdomain/:subdomain", r.tenantHandler.GetTenantBySubdomain) // 通过子域名获取租户
		tenants.GET("/domain/:domain", r.tenantHandler.GetTenantByDomain)          // 通过域名获取租户
		
		// 租户用户管理
		tenants.POST("/:id/users", r.tenantHandler.AddTenantUser)           // 添加租户用户
		tenants.GET("/:id/users", r.tenantHandler.ListTenantUsers)          // 列出租户用户
		tenants.DELETE("/:id/users/:user_id", r.tenantHandler.RemoveTenantUser) // 移除租户用户
		
		// 租户统计和监控
		tenants.GET("/:id/stats", r.tenantHandler.GetTenantStats)   // 获取租户统计
		tenants.GET("/:id/health", r.tenantHandler.GetTenantHealth) // 获取租户健康状态
		
		// 租户状态管理
		tenants.POST("/:id/activate", r.tenantHandler.ActivateTenant) // 激活租户
		tenants.POST("/:id/suspend", r.tenantHandler.SuspendTenant)   // 暂停租户
	}
	
	// 公共租户路由组 - 不需要认证（用于租户识别）
	publicTenants := v1.Group("/public/tenants")
	{
		// 租户识别接口（用于前端路由和域名解析）
		publicTenants.GET("/resolve/subdomain/:subdomain", r.tenantHandler.GetTenantBySubdomain)
		publicTenants.GET("/resolve/domain/:domain", r.tenantHandler.GetTenantByDomain)
	}
	
	// 管理员租户路由组 - 需要管理员权限
	adminTenants := v1.Group("/admin/tenants")
	adminTenants.Use(authMiddleware)
	adminTenants.Use(middleware.AdminPermissionMiddleware()) // 管理员权限中间件
	{
		// 管理员专用操作
		adminTenants.GET("", r.tenantHandler.ListTenants)                    // 列出所有租户
		adminTenants.GET("/:id", r.tenantHandler.GetTenant)                  // 获取任意租户详情
		adminTenants.PUT("/:id", r.tenantHandler.UpdateTenant)               // 更新任意租户
		adminTenants.DELETE("/:id", r.tenantHandler.DeleteTenant)            // 删除任意租户
		adminTenants.POST("/:id/activate", r.tenantHandler.ActivateTenant)   // 激活任意租户
		adminTenants.POST("/:id/suspend", r.tenantHandler.SuspendTenant)     // 暂停任意租户
		
		// 批量操作（TODO: 需要实现批量操作处理器）
		// adminTenants.POST("/batch/activate", r.tenantHandler.BatchActivateTenants)
		// adminTenants.POST("/batch/suspend", r.tenantHandler.BatchSuspendTenants)
		// adminTenants.DELETE("/batch", r.tenantHandler.BatchDeleteTenants)
	}
}

// SetupTenantContextRoutes 设置租户上下文路由（在特定租户下的操作）
func (r *TenantRoutes) SetupTenantContextRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// 租户上下文路由组 - 所有操作都在特定租户上下文中
	tenantContext := router.Group("/api/v1/tenant/:tenant_id")
	tenantContext.Use(authMiddleware)
	tenantContext.Use(middleware.TenantContextMiddleware()) // 租户上下文中间件
	tenantContext.Use(middleware.TenantAccessMiddleware())  // 租户访问权限中间件
	{
		// 当前租户信息
		tenantContext.GET("/info", r.tenantHandler.GetTenant)
		tenantContext.PUT("/info", r.tenantHandler.UpdateTenant)
		
		// 当前租户用户管理
		tenantContext.GET("/users", r.tenantHandler.ListTenantUsers)
		tenantContext.POST("/users", r.tenantHandler.AddTenantUser)
		tenantContext.DELETE("/users/:user_id", r.tenantHandler.RemoveTenantUser)
		
		// 当前租户统计
		tenantContext.GET("/stats", r.tenantHandler.GetTenantStats)
		tenantContext.GET("/health", r.tenantHandler.GetTenantHealth)
		
		// TODO: 在这里可以添加其他服务的租户上下文路由
		// 例如：AI服务、聊天服务等在特定租户下的操作
	}
}

// SetupWebhookRoutes 设置租户相关的Webhook路由
func (r *TenantRoutes) SetupWebhookRoutes(router *gin.Engine) {
	webhooks := router.Group("/webhooks/tenants")
	{
		// 租户事件Webhook（用于集成第三方系统）
		// webhooks.POST("/created", r.tenantHandler.HandleTenantCreatedWebhook)
		// webhooks.POST("/updated", r.tenantHandler.HandleTenantUpdatedWebhook)
		// webhooks.POST("/deleted", r.tenantHandler.HandleTenantDeletedWebhook)
		// webhooks.POST("/suspended", r.tenantHandler.HandleTenantSuspendedWebhook)
		// webhooks.POST("/activated", r.tenantHandler.HandleTenantActivatedWebhook)
	}
}

// SetupHealthRoutes 设置租户健康检查路由
func (r *TenantRoutes) SetupHealthRoutes(router *gin.Engine) {
	health := router.Group("/health/tenants")
	{
		// 租户服务健康检查
		health.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "multi-tenancy",
				"version": "1.0.0",
			})
		})
		
		// 租户数据库连接检查
		// health.GET("/db", r.tenantHandler.CheckDatabaseHealth)
		
		// 租户缓存连接检查
		// health.GET("/cache", r.tenantHandler.CheckCacheHealth)
		
		// 租户整体健康状态
		// health.GET("/status", r.tenantHandler.GetServiceHealth)
	}
}

// SetupMetricsRoutes 设置租户指标路由
func (r *TenantRoutes) SetupMetricsRoutes(router *gin.Engine) {
	metrics := router.Group("/metrics/tenants")
	{
		// Prometheus指标端点
		// metrics.GET("/prometheus", r.tenantHandler.PrometheusMetrics)
		
		// 租户使用指标
		// metrics.GET("/usage", r.tenantHandler.GetUsageMetrics)
		
		// 租户性能指标
		// metrics.GET("/performance", r.tenantHandler.GetPerformanceMetrics)
	}
}

// SetupAllRoutes 设置所有租户相关路由
func (r *TenantRoutes) SetupAllRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// 设置主要路由
	r.SetupRoutes(router, authMiddleware)
	
	// 设置租户上下文路由
	r.SetupTenantContextRoutes(router, authMiddleware)
	
	// 设置Webhook路由
	r.SetupWebhookRoutes(router)
	
	// 设置健康检查路由
	r.SetupHealthRoutes(router)
	
	// 设置指标路由
	r.SetupMetricsRoutes(router)
}