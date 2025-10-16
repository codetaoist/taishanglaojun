package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/multi-tenancy/middleware"
)

// TenantRoutes 
type TenantRoutes struct {
	tenantHandler *handlers.TenantHandler
}

// NewTenantRoutes 
func NewTenantRoutes(tenantHandler *handlers.TenantHandler) *TenantRoutes {
	return &TenantRoutes{
		tenantHandler: tenantHandler,
	}
}

// SetupRoutes 
func (r *TenantRoutes) SetupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// API汾?
	v1 := router.Group("/api/v1")
	
	// ?- ?
	tenants := v1.Group("/tenants")
	tenants.Use(authMiddleware)
	tenants.Use(middleware.TenantPermissionMiddleware()) // ?
	{
		// 
		tenants.POST("", r.tenantHandler.CreateTenant)           // 
		tenants.GET("", r.tenantHandler.ListTenants)             // 
		tenants.GET("/:id", r.tenantHandler.GetTenant)           // 
		tenants.PUT("/:id", r.tenantHandler.UpdateTenant)        // 
		tenants.DELETE("/:id", r.tenantHandler.DeleteTenant)     // 
		
		// 
		tenants.GET("/subdomain/:subdomain", r.tenantHandler.GetTenantBySubdomain) // ?
		tenants.GET("/domain/:domain", r.tenantHandler.GetTenantByDomain)          // 
		
		// 
		tenants.POST("/:id/users", r.tenantHandler.AddTenantUser)           // 
		tenants.GET("/:id/users", r.tenantHandler.ListTenantUsers)          // 
		tenants.DELETE("/:id/users/:user_id", r.tenantHandler.RemoveTenantUser) // 
		
		// ?
		tenants.GET("/:id/stats", r.tenantHandler.GetTenantStats)   // 
		tenants.GET("/:id/health", r.tenantHandler.GetTenantHealth) // ?
		
		// ?
		tenants.POST("/:id/activate", r.tenantHandler.ActivateTenant) // ?
		tenants.POST("/:id/suspend", r.tenantHandler.SuspendTenant)   // 
	}
	
	// ?- ?
	publicTenants := v1.Group("/public/tenants")
	{
		// ?
		publicTenants.GET("/resolve/subdomain/:subdomain", r.tenantHandler.GetTenantBySubdomain)
		publicTenants.GET("/resolve/domain/:domain", r.tenantHandler.GetTenantByDomain)
	}
	
	//  - 
	adminTenants := v1.Group("/admin/tenants")
	adminTenants.Use(authMiddleware)
	adminTenants.Use(middleware.AdminPermissionMiddleware()) // 
	{
		// ?
		adminTenants.GET("", r.tenantHandler.ListTenants)                    // ?
		adminTenants.GET("/:id", r.tenantHandler.GetTenant)                  // 
		adminTenants.PUT("/:id", r.tenantHandler.UpdateTenant)               // 
		adminTenants.DELETE("/:id", r.tenantHandler.DeleteTenant)            // 
		adminTenants.POST("/:id/activate", r.tenantHandler.ActivateTenant)   // ?
		adminTenants.POST("/:id/suspend", r.tenantHandler.SuspendTenant)     // 
		
		// TODO: ?
		// adminTenants.POST("/batch/activate", r.tenantHandler.BatchActivateTenants)
		// adminTenants.POST("/batch/suspend", r.tenantHandler.BatchSuspendTenants)
		// adminTenants.DELETE("/batch", r.tenantHandler.BatchDeleteTenants)
	}
}

// SetupTenantContextRoutes 
func (r *TenantRoutes) SetupTenantContextRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	//  - ?
	tenantContext := router.Group("/api/v1/tenant/:tenant_id")
	tenantContext.Use(authMiddleware)
	tenantContext.Use(middleware.TenantContextMiddleware()) // 
	tenantContext.Use(middleware.TenantAccessMiddleware())  // ?
	{
		// 
		tenantContext.GET("/info", r.tenantHandler.GetTenant)
		tenantContext.PUT("/info", r.tenantHandler.UpdateTenant)
		
		// 
		tenantContext.GET("/users", r.tenantHandler.ListTenantUsers)
		tenantContext.POST("/users", r.tenantHandler.AddTenantUser)
		tenantContext.DELETE("/users/:user_id", r.tenantHandler.RemoveTenantUser)
		
		// 
		tenantContext.GET("/stats", r.tenantHandler.GetTenantStats)
		tenantContext.GET("/health", r.tenantHandler.GetTenantHealth)
		
		// TODO: ?
		// AI?
	}
}

// SetupWebhookRoutes Webhook
func (r *TenantRoutes) SetupWebhookRoutes(router *gin.Engine) {
	webhooks := router.Group("/webhooks/tenants")
	{
		// Webhook?
		// webhooks.POST("/created", r.tenantHandler.HandleTenantCreatedWebhook)
		// webhooks.POST("/updated", r.tenantHandler.HandleTenantUpdatedWebhook)
		// webhooks.POST("/deleted", r.tenantHandler.HandleTenantDeletedWebhook)
		// webhooks.POST("/suspended", r.tenantHandler.HandleTenantSuspendedWebhook)
		// webhooks.POST("/activated", r.tenantHandler.HandleTenantActivatedWebhook)
	}
}

// SetupHealthRoutes ?
func (r *TenantRoutes) SetupHealthRoutes(router *gin.Engine) {
	health := router.Group("/health/tenants")
	{
		// ?
		health.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"service": "multi-tenancy",
				"version": "1.0.0",
			})
		})
		
		// ?
		// health.GET("/db", r.tenantHandler.CheckDatabaseHealth)
		
		// ?
		// health.GET("/cache", r.tenantHandler.CheckCacheHealth)
		
		// 彡?
		// health.GET("/status", r.tenantHandler.GetServiceHealth)
	}
}

// SetupMetricsRoutes 
func (r *TenantRoutes) SetupMetricsRoutes(router *gin.Engine) {
	metrics := router.Group("/metrics/tenants")
	{
		// Prometheus
		// metrics.GET("/prometheus", r.tenantHandler.PrometheusMetrics)
		
		// 
		// metrics.GET("/usage", r.tenantHandler.GetUsageMetrics)
		
		// 
		// metrics.GET("/performance", r.tenantHandler.GetPerformanceMetrics)
	}
}

// SetupAllRoutes ?
func (r *TenantRoutes) SetupAllRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	// 
	r.SetupRoutes(router, authMiddleware)
	
	// ?
	r.SetupTenantContextRoutes(router, authMiddleware)
	
	// Webhook
	r.SetupWebhookRoutes(router)
	
	// ?
	r.SetupHealthRoutes(router)
	
	// 
	r.SetupMetricsRoutes(router)
}

