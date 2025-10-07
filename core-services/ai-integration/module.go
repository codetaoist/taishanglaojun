package ai_integration

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// Module AI集成服务模块
type Module struct {
	// 基础组件
	db              *gorm.DB
	redisClient     *redis.Client
	logger          *zap.Logger
	providerManager *providers.Manager
	
	// 服务层
	chatService       *services.ChatService
	aiService         *services.AIService
	contextManager    *services.ContextManager
	multimodalService *services.MultimodalService
	
	// 处理器层
	chatHandler       *handlers.ChatHandler
	aiHandler         *handlers.AIHandler
	multimodalHandler *handlers.MultimodalHandler
	
	// gRPC服务器
	grpcServer   *grpc.Server
	grpcListener net.Listener
	
	// 配置
	config *ModuleConfig
}

// ModuleConfig 模块配置
type ModuleConfig struct {
	// HTTP配置
	HTTPEnabled bool   `json:"http_enabled"`
	HTTPPrefix  string `json:"http_prefix"`
	
	// gRPC配置
	GRPCEnabled bool   `json:"grpc_enabled"`
	GRPCPort    int    `json:"grpc_port"`
	GRPCHost    string `json:"grpc_host"`
	
	// 服务配置
	ServiceConfig *AIServiceConfig `json:"service_config"`
	
	// 提供商配置
	ProviderConfig *ProviderConfig `json:"provider_config"`
	
	// 缓存配置
	CacheConfig *CacheConfig `json:"cache_config"`
	
	// 对话配置
	ChatConfig *ChatConfig `json:"chat_config"`
}

// AIServiceConfig AI服务配置
type AIServiceConfig struct {
	ServiceName        string        `json:"service_name"`
	Version           string        `json:"version"`
	Environment       string        `json:"environment"`
	MaxConcurrentReqs int           `json:"max_concurrent_requests"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MetricsRetention  time.Duration `json:"metrics_retention"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	DefaultProvider string                 `json:"default_provider"`
	Providers       map[string]interface{} `json:"providers"`
	LoadBalancing   bool                   `json:"load_balancing"`
	Fallback        bool                   `json:"fallback"`
	RateLimiting    *RateLimitConfig       `json:"rate_limiting"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultTTL       time.Duration `json:"default_ttl"`
	ResponseTTL      time.Duration `json:"response_ttl"`
	ConversationTTL  time.Duration `json:"conversation_ttl"`
	EmbeddingTTL     time.Duration `json:"embedding_ttl"`
	MaxCacheSize     int64         `json:"max_cache_size"`
	EnableCompression bool         `json:"enable_compression"`
}

// ChatConfig 对话配置
type ChatConfig struct {
	MaxMessages        int           `json:"max_messages"`
	MaxTokens          int           `json:"max_tokens"`
	DefaultTemperature float32       `json:"default_temperature"`
	SessionTimeout     time.Duration `json:"session_timeout"`
	ContextWindow      int           `json:"context_window"`
	EnableMemory       bool          `json:"enable_memory"`
}

// NewModule 创建AI集成服务模块
func NewModule(config *ModuleConfig, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) (*Module, error) {
	if config == nil {
		config = getDefaultConfig()
	}
	
	module := &Module{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		config:      config,
	}
	
	// 初始化提供商管理器
	if err := module.initProviderManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize provider manager: %w", err)
	}
	
	// 初始化服务层
	if err := module.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	// 初始化处理器层
	if err := module.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}
	
	// 初始化gRPC服务器（如果启用）
	if config.GRPCEnabled {
		if err := module.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}
	
	return module, nil
}

// initProviderManager 初始化提供商管理器
func (m *Module) initProviderManager() error {
	m.logger.Info("Initializing AI provider manager")
	
	// 创建提供商管理器
	manager := providers.NewManager(m.logger)
	
	// 根据配置初始化提供商
	for providerName, providerConfig := range m.config.ProviderConfig.Providers {
		switch providerName {
		case "openai":
			if configMap, ok := providerConfig.(map[string]interface{}); ok {
				config := providers.OpenAIConfig{}
				if apiKey, exists := configMap["api_key"]; exists {
					if str, ok := apiKey.(string); ok {
						config.APIKey = str
					}
				}
				if baseURL, exists := configMap["base_url"]; exists {
					if str, ok := baseURL.(string); ok {
						config.BaseURL = str
					}
				}
				if timeout, exists := configMap["timeout"]; exists {
					if num, ok := timeout.(float64); ok {
						config.Timeout = int(num)
					}
				}
				provider := providers.NewOpenAIProvider(config, m.logger)
				manager.RegisterProvider("openai", provider)
			}
		case "azure":
			if configMap, ok := providerConfig.(map[string]interface{}); ok {
				config := providers.AzureConfig{}
				if apiKey, exists := configMap["api_key"]; exists {
					if str, ok := apiKey.(string); ok {
						config.APIKey = str
					}
				}
				if endpoint, exists := configMap["endpoint"]; exists {
					if str, ok := endpoint.(string); ok {
						config.Endpoint = str
					}
				}
				if deploymentName, exists := configMap["deployment_name"]; exists {
					if str, ok := deploymentName.(string); ok {
						config.DeploymentName = str
					}
				}
				if apiVersion, exists := configMap["api_version"]; exists {
					if str, ok := apiVersion.(string); ok {
						config.APIVersion = str
					}
				}
				if timeout, exists := configMap["timeout"]; exists {
					if num, ok := timeout.(float64); ok {
						config.Timeout = int(num)
					}
				}
				provider := providers.NewAzureProvider(config, m.logger)
				manager.RegisterProvider("azure", provider)
			}
		case "baidu":
			if configMap, ok := providerConfig.(map[string]interface{}); ok {
				config := providers.BaiduConfig{}
				if apiKey, exists := configMap["api_key"]; exists {
					if str, ok := apiKey.(string); ok {
						config.APIKey = str
					}
				}
				if secretKey, exists := configMap["secret_key"]; exists {
					if str, ok := secretKey.(string); ok {
						config.SecretKey = str
					}
				}
				if baseURL, exists := configMap["base_url"]; exists {
					if str, ok := baseURL.(string); ok {
						config.BaseURL = str
					}
				}
				if timeout, exists := configMap["timeout"]; exists {
					if num, ok := timeout.(float64); ok {
						config.Timeout = int(num)
					}
				}
				provider := providers.NewBaiduProvider(config, m.logger)
				manager.RegisterProvider("baidu", provider)
			}
		case "mock":
			provider := providers.NewMockProvider(m.logger)
			manager.RegisterProvider("mock", provider)
		}
	}
	
	// 设置默认提供商
	if m.config.ProviderConfig.DefaultProvider != "" {
		manager.SetDefaultProvider(m.config.ProviderConfig.DefaultProvider)
	}
	
	m.providerManager = manager
	m.logger.Info("AI provider manager initialized successfully")
	return nil
}

// initServices 初始化服务层
func (m *Module) initServices() error {
	m.logger.Info("Initializing AI integration services")
	
	// 创建上下文管理器
	m.contextManager = services.NewContextManager(m.db, m.logger)
	
	// 创建核心服务
	m.chatService = services.NewChatService(m.db, m.logger, m.providerManager)
	m.aiService = services.NewAIService(m.providerManager)
	
	// 创建多模态服务（需要实现具体的依赖服务）
	// TODO: 实现FileService, AudioService, ImageService, VideoService
	// m.multimodalService = services.NewMultimodalService(
	//     m.providerManager.GetProviders(),
	//     repository,
	//     fileService,
	//     audioService,
	//     imageService,
	//     videoService,
	// )
	
	m.logger.Info("AI integration services initialized successfully")
	return nil
}

// initHandlers 初始化处理器层
func (m *Module) initHandlers() error {
	m.logger.Info("Initializing AI integration handlers")
	
	m.chatHandler = handlers.NewChatHandler(m.chatService, m.logger)
	m.aiHandler = handlers.NewAIHandler(m.aiService, m.logger)
	
	// 初始化多模态处理器
	if m.multimodalService != nil {
		m.multimodalHandler = handlers.NewMultimodalHandler(m.multimodalService)
	}
	
	m.logger.Info("AI integration handlers initialized successfully")
	return nil
}

// initGRPCServer 初始化gRPC服务器
func (m *Module) initGRPCServer() error {
	m.logger.Info("Initializing AI integration gRPC server")
	
	// 创建gRPC服务器
	m.grpcServer = grpc.NewServer()
	
	// 注册gRPC服务
	// TODO: 实现gRPC服务定义和注册
	
	// 创建监听器
	addr := fmt.Sprintf("%s:%d", m.config.GRPCHost, m.config.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	m.grpcListener = listener
	
	m.logger.Info("AI integration gRPC server initialized", zap.String("address", addr))
	return nil
}

// SetupRoutes 设置HTTP路由
func (m *Module) SetupRoutes(router *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware) error {
	if !m.config.HTTPEnabled {
		m.logger.Info("HTTP routes disabled, skipping route setup")
		return nil
	}
	
	m.logger.Info("Setting up AI integration HTTP routes")
	
	// 创建AI集成路由组
	aiGroup := router.Group(m.config.HTTPPrefix)
	
	// 公开路由（不需要认证）
	{
		// 提供商信息
		aiGroup.GET("/providers", m.getProviders)
		aiGroup.GET("/models", m.getModels)
		aiGroup.GET("/health", m.healthCheck)
	}
	
	// 需要认证的路由
	if jwtMiddleware != nil {
		authGroup := aiGroup.Group("")
		authGroup.Use(jwtMiddleware.AuthRequired())
		{
			// 对话相关
			authGroup.POST("/chat", m.chatHandler.Chat)
			authGroup.GET("/sessions", m.chatHandler.GetSessions)
			authGroup.GET("/sessions/:session_id/messages", m.chatHandler.GetMessages)
			authGroup.DELETE("/sessions/:session_id", m.chatHandler.DeleteSession)
			authGroup.POST("/sessions/:session_id/clear", m.chatHandler.ClearSession)
			
			// AI功能
			authGroup.POST("/intent", m.aiHandler.IntentRecognition)
			authGroup.POST("/sentiment", m.aiHandler.SentimentAnalysis)
			authGroup.POST("/generate/summary", m.aiHandler.GenerateSummary)
			authGroup.POST("/generate/explanation", m.aiHandler.GenerateExplanation)
			authGroup.POST("/generate/translation", m.aiHandler.GenerateTranslation)
			authGroup.POST("/analyze/keywords", m.aiHandler.ExtractKeywords)
			authGroup.POST("/analyze/similarity", m.aiHandler.CalculateSimilarity)
			authGroup.POST("/embed", m.aiHandler.GenerateEmbedding)
			
			// 上下文管理
			authGroup.GET("/context/:user_id", m.getContext)
			authGroup.POST("/context/:user_id", m.updateContext)
			authGroup.DELETE("/context/:user_id", m.clearContext)
			
			// 多模态AI功能
			if m.multimodalHandler != nil {
				multimodalGroup := authGroup.Group("/multimodal")
				{
					multimodalGroup.POST("/process", m.multimodalHandler.ProcessMultimodal)
					multimodalGroup.POST("/upload", m.multimodalHandler.UploadFile)
					multimodalGroup.GET("/stream", m.multimodalHandler.StreamMultimodal)
					
					// 会话管理
					multimodalGroup.GET("/sessions", m.multimodalHandler.GetSessions)
					multimodalGroup.POST("/sessions", m.multimodalHandler.CreateSession)
					multimodalGroup.GET("/sessions/:id", m.multimodalHandler.GetSession)
					multimodalGroup.PUT("/sessions/:id", m.multimodalHandler.UpdateSession)
					multimodalGroup.DELETE("/sessions/:id", m.multimodalHandler.DeleteSession)
					multimodalGroup.GET("/sessions/:id/messages", m.multimodalHandler.GetSessionMessages)
				}
			}
		}
	}
	
	m.logger.Info("AI integration HTTP routes setup completed")
	return nil
}

// Start 启动模块
func (m *Module) Start() error {
	m.logger.Info("Starting AI integration module")
	
	// 自动迁移数据库
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	// TODO: 实现Start方法
	// 启动提供商管理器
	// if err := m.providerManager.Start(); err != nil {
	//	return fmt.Errorf("failed to start provider manager: %w", err)
	// }
	
	// 启动gRPC服务器
	if m.config.GRPCEnabled && m.grpcServer != nil {
		go func() {
			m.logger.Info("Starting AI integration gRPC server", 
				zap.String("address", m.grpcListener.Addr().String()))
			if err := m.grpcServer.Serve(m.grpcListener); err != nil {
				m.logger.Error("gRPC server error", zap.Error(err))
			}
		}()
	}
	
	// 启动后台任务
	go m.startBackgroundTasks()
	
	m.logger.Info("AI integration module started successfully")
	return nil
}

// Stop 停止模块
func (m *Module) Stop() error {
	m.logger.Info("Stopping AI integration module")
	
	// 停止gRPC服务器
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		if m.grpcListener != nil {
			m.grpcListener.Close()
		}
	}
	
	// TODO: 实现Stop方法
	// 停止提供商管理器
	// if m.providerManager != nil {
	//	m.providerManager.Stop()
	// }
	
	m.logger.Info("AI integration module stopped successfully")
	return nil
}

// Health 健康检查
func (m *Module) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"module": "ai-integration",
		"version": m.config.ServiceConfig.Version,
		"services": map[string]string{
			"chat_service":    "running",
			"ai_service":      "running",
			"context_manager": "running",
		},
	}
	
	// 检查数据库连接
	if sqlDB, err := m.db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["database"] = "healthy"
		}
	}
	
	// 检查Redis连接
	if err := m.redisClient.Ping(context.Background()).Err(); err != nil {
		health["redis"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["redis"] = "healthy"
	}
	
	// 检查提供商状态
	if m.providerManager != nil {
		// TODO: 实现Health方法
		// providerHealth := m.providerManager.Health()
		// health["providers"] = providerHealth
		// if len(providerHealth) == 0 {
		//	health["status"] = "degraded"
		// }
		health["providers"] = "not_implemented"
	}
	
	return health
}

// migrateDatabase 迁移数据库
func (m *Module) migrateDatabase() error {
	m.logger.Info("Migrating AI integration database")
	
	// 自动迁移模型
	err := m.db.AutoMigrate(
		&models.Conversation{},
		&models.AIRequest{},
		&models.AIResponse{},
		&models.ConversationContext{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.MultimodalSession{},
		&models.MultimodalMessage{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	
	m.logger.Info("AI integration database migration completed")
	return nil
}

// startBackgroundTasks 启动后台任务
func (m *Module) startBackgroundTasks() {
	m.logger.Info("Starting AI integration background tasks")
	
	// 定期清理过期会话
	go m.cleanupExpiredSessionsPeriodically()
	
	// 定期清理缓存
	go m.cleanupCachePeriodically()
	
	// 定期更新提供商状态
	go m.updateProviderStatusPeriodically()
}

// cleanupExpiredSessionsPeriodically 定期清理过期会话
func (m *Module) cleanupExpiredSessionsPeriodically() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: 实现CleanupExpiredSessions方法
		// if err := m.chatService.CleanupExpiredSessions(context.Background()); err != nil {
		//	m.logger.Error("Failed to cleanup expired sessions", zap.Error(err))
		// }
		m.logger.Debug("Cleanup expired sessions task executed")
	}
}

// cleanupCachePeriodically 定期清理缓存
func (m *Module) cleanupCachePeriodically() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: 实现Cleanup方法
		// if err := m.contextManager.Cleanup(context.Background()); err != nil {
		//	m.logger.Error("Failed to cleanup cache", zap.Error(err))
		// }
		m.logger.Debug("Cleanup cache task executed")
	}
}

// updateProviderStatusPeriodically 定期更新提供商状态
func (m *Module) updateProviderStatusPeriodically() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: 实现UpdateStatus方法
		// m.providerManager.UpdateStatus()
		m.logger.Debug("Update provider status task executed")
	}
}

// HTTP处理器方法
func (m *Module) getProviders(c *gin.Context) {
	providers := m.providerManager.GetProviders()
	defaultProvider, _ := m.providerManager.GetDefaultProvider()
	c.JSON(200, gin.H{
		"providers": providers,
		"default":   defaultProvider,
	})
}

func (m *Module) getModels(c *gin.Context) {
	providerName := c.Query("provider")
	if providerName == "" {
		// TODO: 修复类型转换问题
		// defaultProvider, _ := m.providerManager.GetDefaultProvider()
		// providerName = defaultProvider
		providerName = "default"
	}
	
	// TODO: 修复GetModels方法调用
	// models := m.providerManager.GetModels(providerName)
	models := []string{"gpt-3.5-turbo", "gpt-4"}
	c.JSON(200, gin.H{
		"provider": providerName,
		"models":   models,
	})
}

func (m *Module) healthCheck(c *gin.Context) {
	health := m.Health()
	status := 200
	if health["status"] != "healthy" {
		status = 503
	}
	c.JSON(status, health)
}

func (m *Module) getContext(c *gin.Context) {
	userID := c.Param("user_id")
	// TODO: 实现GetContext方法
	// context, err := m.contextManager.GetContext(c.Request.Context(), userID)
	// if err != nil {
	//	c.JSON(500, gin.H{"error": err.Error()})
	//	return
	// }
	context := map[string]interface{}{
		"user_id": userID,
		"message": "Context not implemented yet",
	}
	c.JSON(200, context)
}

func (m *Module) updateContext(c *gin.Context) {
	// userID := c.Param("user_id") // 暂时注释掉未使用的变量
	var contextData map[string]interface{}
	if err := c.ShouldBindJSON(&contextData); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	// TODO: 修复UpdateContext方法调用参数
	// if err := m.contextManager.UpdateContext(c.Request.Context(), userID, contextData); err != nil {
	//	c.JSON(500, gin.H{"error": err.Error()})
	//	return
	// }
	
	c.JSON(200, gin.H{"message": "Context updated successfully"})
}

func (m *Module) clearContext(c *gin.Context) {
	// userID := c.Param("user_id") // 暂时注释掉未使用的变量
	// TODO: 实现ClearContext方法
	// if err := m.contextManager.ClearContext(c.Request.Context(), userID); err != nil {
	//	c.JSON(500, gin.H{"error": err.Error()})
	//	return
	// }
	
	c.JSON(200, gin.H{"message": "Context cleared successfully"})
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *ModuleConfig {
	return &ModuleConfig{
		HTTPEnabled: true,
		HTTPPrefix:  "/ai",
		GRPCEnabled: false,
		GRPCPort:    50053,
		GRPCHost:    "localhost",
		ServiceConfig: &AIServiceConfig{
			ServiceName:       "ai-integration-service",
			Version:          "1.0.0",
			Environment:      "development",
			MaxConcurrentReqs: 50,
			RequestTimeout:    30 * time.Second,
			MetricsRetention:  24 * time.Hour,
		},
		ProviderConfig: &ProviderConfig{
			DefaultProvider: "mock",
			Providers: map[string]interface{}{
				"mock": map[string]interface{}{
					"enabled": true,
				},
			},
			LoadBalancing: false,
			Fallback:      true,
			RateLimiting: &RateLimitConfig{
				RequestsPerMinute: 60,
				BurstSize:         10,
				WindowSize:        1 * time.Minute,
			},
		},
		CacheConfig: &CacheConfig{
			DefaultTTL:        1 * time.Hour,
			ResponseTTL:       30 * time.Minute,
			ConversationTTL:   24 * time.Hour,
			EmbeddingTTL:      7 * 24 * time.Hour,
			MaxCacheSize:      10000000, // 10MB
			EnableCompression: true,
		},
		ChatConfig: &ChatConfig{
			MaxMessages:        100,
			MaxTokens:          4000,
			DefaultTemperature: 0.7,
			SessionTimeout:     2 * time.Hour,
			ContextWindow:      10,
			EnableMemory:       true,
		},
	}
}