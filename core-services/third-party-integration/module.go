package thirdpartyintegration

import (
	"context"
	"database/sql"
	"log"

	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/controllers"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/third-party-integration/services"
	"github.com/gin-gonic/gin"
)

// Module 第三方集成模块
type Module struct {
	db                 *sql.DB
	apiKeyService      *services.APIKeyService
	pluginService      *services.PluginService
	integrationService *services.IntegrationService
	webhookService     *services.WebhookService
	oauthService       *services.OAuthService
}

// NewModule 创建第三方集成模块
func NewModule(db *sql.DB) *Module {
	//
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	pluginRepo := repositories.NewPluginRepository(db)
	integrationRepo := repositories.NewIntegrationRepository(db)
	webhookRepo := repositories.NewWebhookRepository(db)
	oauthRepo := repositories.NewOAuthRepository(db)

	//
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)
	pluginService := services.NewPluginService(pluginRepo)
	integrationService := services.NewIntegrationService(integrationRepo)
	webhookService := services.NewWebhookService(webhookRepo)
	oauthService := services.NewOAuthService(oauthRepo)

	return &Module{
		db:                 db,
		apiKeyService:      apiKeyService,
		pluginService:      pluginService,
		integrationService: integrationService,
		webhookService:     webhookService,
		oauthService:       oauthService,
	}
}

// RegisterRoutes
func (m *Module) RegisterRoutes(router *gin.Engine) {
	// API
	apiKeyHandler := controllers.NewAPIKeyHandler(m.apiKeyService)
	apiGroup := router.Group("/api/v1/developer")
	{
		apiGroup.POST("/keys", apiKeyHandler.CreateAPIKey)
		apiGroup.GET("/keys", apiKeyHandler.ListAPIKeys)
		apiGroup.GET("/keys/:id", apiKeyHandler.GetAPIKey)
		apiGroup.PUT("/keys/:id", apiKeyHandler.UpdateAPIKey)
		apiGroup.DELETE("/keys/:id", apiKeyHandler.DeleteAPIKey)
		apiGroup.POST("/keys/:id/regenerate", apiKeyHandler.RegenerateAPIKey)
	}

	//
	pluginHandler := controllers.NewPluginHandler(m.pluginService)
	pluginGroup := router.Group("/api/v1/plugins")
	{
		pluginGroup.POST("", pluginHandler.InstallPlugin)
		pluginGroup.GET("", pluginHandler.ListPlugins)
		pluginGroup.GET("/:id", pluginHandler.GetPlugin)
		pluginGroup.PUT("/:id", pluginHandler.UpdatePlugin)
		pluginGroup.DELETE("/:id", pluginHandler.UninstallPlugin)
		pluginGroup.POST("/:id/enable", pluginHandler.EnablePlugin)
		pluginGroup.POST("/:id/disable", pluginHandler.DisablePlugin)
		pluginGroup.GET("/:id/config", pluginHandler.GetPluginConfig)
		pluginGroup.PUT("/:id/config", pluginHandler.UpdatePluginConfig)
	}

	// 集成处理器
	integrationHandler := controllers.NewIntegrationHandler(m.integrationService, m.oauthService)
	integrationGroup := router.Group("/api/v1/integrations")
	{
		integrationGroup.POST("", integrationHandler.CreateIntegration)
		integrationGroup.GET("", integrationHandler.ListIntegrations)
		integrationGroup.GET("/:id", integrationHandler.GetIntegration)
		integrationGroup.PUT("/:id", integrationHandler.UpdateIntegration)
		integrationGroup.DELETE("/:id", integrationHandler.DeleteIntegration)
		integrationGroup.POST("/:id/test", integrationHandler.TestIntegration)
		integrationGroup.POST("/:id/sync", integrationHandler.SyncData)
	}

	// OAuth
	oauthHandler := controllers.NewOAuthHandler(m.oauthService)
	oauthGroup := router.Group("/api/v1/oauth")
	{
		oauthGroup.GET("/:provider/authorize", oauthHandler.Authorize)
		oauthGroup.POST("/:provider/callback", oauthHandler.Callback)
		oauthGroup.POST("/:provider/refresh", oauthHandler.RefreshToken)
		oauthGroup.DELETE("/:provider/revoke", oauthHandler.RevokeToken)
	}

	// Webhook
	webhookHandler := controllers.NewWebhookHandler(m.webhookService)
	webhookGroup := router.Group("/api/v1/webhooks")
	{
		webhookGroup.POST("", webhookHandler.CreateWebhook)
		webhookGroup.GET("", webhookHandler.ListWebhooks)
		webhookGroup.GET("/:id", webhookHandler.GetWebhook)
		webhookGroup.PUT("/:id", webhookHandler.UpdateWebhook)
		webhookGroup.DELETE("/:id", webhookHandler.DeleteWebhook)
		webhookGroup.POST("/:id/test", webhookHandler.TestWebhook)
		webhookGroup.POST("/receive/:token", webhookHandler.ReceiveWebhook)
	}
}

// Initialize 初始化第三方集成模块
func (m *Module) Initialize(ctx context.Context) error {
	log.Println("Initializing Third-Party Integration module...")

	//
	if err := m.createTables(); err != nil {
		return err
	}

	// 初始化默认值
	if err := m.initializeDefaults(); err != nil {
		return err
	}

	log.Println("Third-Party Integration module initialized successfully")
	return nil
}

// createTables
func (m *Module) createTables() error {
	tables := []string{
		models.APIKeyTableSQL,
		models.PluginTableSQL,
		models.IntegrationTableSQL,
		models.WebhookTableSQL,
		models.OAuthTokenTableSQL,
	}

	for _, tableSQL := range tables {
		if _, err := m.db.Exec(tableSQL); err != nil {
			return err
		}
	}

	return nil
}

// initializeDefaults 初始化默认值
func (m *Module) initializeDefaults() error {
	//
	return nil
}

// Cleanup
func (m *Module) Cleanup() error {
	log.Println("Cleaning up Third-Party Integration module...")
	return nil
}
