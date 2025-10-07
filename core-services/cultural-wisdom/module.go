package cultural_wisdom

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

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	aiServices "github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// Module 文化智慧服务模块
type Module struct {
	// 基础组件
	db              *gorm.DB
	redisClient     *redis.Client
	logger          *zap.Logger
	providerManager *providers.Manager
	
	// 服务层
	wisdomService         *services.WisdomService
	aiService            *services.AIService
	searchService        *services.SearchService
	recommendationService *services.RecommendationService
	categoryService      *services.CategoryService
	tagService           *services.TagService
	userBehaviorService  *services.UserBehaviorService
	favoritesService     *services.FavoritesService
	cacheService         *services.CacheService
	
	// 处理器层
	wisdomHandler         *handlers.WisdomHandler
	aiHandler            *handlers.AIHandler
	searchHandler        *handlers.SearchHandler
	recommendationHandler *handlers.RecommendationHandler
	categoryHandler      *handlers.CategoryHandler
	tagHandler           *handlers.TagHandler
	userBehaviorHandler  *handlers.UserBehaviorHandler
	favoritesHandler     *handlers.FavoritesHandler
	
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
	ServiceConfig *CulturalWisdomConfig `json:"service_config"`
	
	// 缓存配置
	CacheConfig *CacheConfig `json:"cache_config"`
	
	// 搜索配置
	SearchConfig *SearchConfig `json:"search_config"`
	
	// 推荐配置
	RecommendationConfig *RecommendationConfig `json:"recommendation_config"`
}

// CulturalWisdomConfig 文化智慧服务配置
type CulturalWisdomConfig struct {
	ServiceName        string        `json:"service_name"`
	Version           string        `json:"version"`
	Environment       string        `json:"environment"`
	MaxConcurrentReqs int           `json:"max_concurrent_requests"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MetricsRetention  time.Duration `json:"metrics_retention"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	DefaultTTL    time.Duration `json:"default_ttl"`
	WisdomTTL     time.Duration `json:"wisdom_ttl"`
	SearchTTL     time.Duration `json:"search_ttl"`
	CategoryTTL   time.Duration `json:"category_ttl"`
	TagTTL        time.Duration `json:"tag_ttl"`
	MaxCacheSize  int64         `json:"max_cache_size"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
	MaxResults        int     `json:"max_results"`
	MinScore          float64 `json:"min_score"`
	VectorDimension   int     `json:"vector_dimension"`
	IndexUpdateInterval time.Duration `json:"index_update_interval"`
}

// RecommendationConfig 推荐配置
type RecommendationConfig struct {
	MaxRecommendations int     `json:"max_recommendations"`
	MinSimilarity      float64 `json:"min_similarity"`
	UserBehaviorWeight float64 `json:"user_behavior_weight"`
	ContentWeight      float64 `json:"content_weight"`
	PopularityWeight   float64 `json:"popularity_weight"`
}

// NewModule 创建文化智慧服务模块
func NewModule(config *ModuleConfig, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger, providerManager *providers.Manager) (*Module, error) {
	if config == nil {
		config = getDefaultConfig()
	}
	
	module := &Module{
		db:              db,
		redisClient:     redisClient,
		logger:          logger,
		providerManager: providerManager,
		config:          config,
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

// initServices 初始化服务层
func (m *Module) initServices() error {
	m.logger.Info("Initializing cultural wisdom services")
	
	// 创建缓存服务
	m.cacheService = services.NewCacheService(m.redisClient, m.logger)
	
	// 创建核心服务
	m.wisdomService = services.NewWisdomService(m.db, m.cacheService)
	m.aiService = services.NewAIService(m.db, m.logger, m.providerManager)
	
	// 创建AI集成服务
	aiIntegrationService := aiServices.NewAIService(m.providerManager)
	
	// 创建搜索服务
	m.searchService = services.NewSearchService(m.db, m.cacheService, aiIntegrationService, m.logger)
	
	// 创建用户行为服务
	m.userBehaviorService = services.NewUserBehaviorService(m.db, m.cacheService, m.logger)
	
	// 创建推荐服务
	m.recommendationService = services.NewRecommendationService(m.db, m.cacheService, m.userBehaviorService, m.aiService, m.logger)
	
	// 创建分类和标签服务
	m.categoryService = services.NewCategoryService(m.db, m.cacheService)
	m.tagService = services.NewTagService(m.db, m.cacheService)
	
	// 创建收藏服务
	m.favoritesService = services.NewFavoritesService(m.db, m.logger)
	
	m.logger.Info("Cultural wisdom services initialized successfully")
	return nil
}

// initHandlers 初始化处理器层
func (m *Module) initHandlers() error {
	m.logger.Info("Initializing cultural wisdom handlers")
	
	m.wisdomHandler = handlers.NewWisdomHandler(m.wisdomService)
	m.aiHandler = handlers.NewAIHandler(m.aiService, m.logger)
	m.searchHandler = handlers.NewSearchHandler(m.searchService)
	m.recommendationHandler = handlers.NewRecommendationHandler(m.recommendationService, m.logger)
	m.categoryHandler = handlers.NewCategoryHandler(m.categoryService)
	m.tagHandler = handlers.NewTagHandler(m.tagService)
	m.userBehaviorHandler = handlers.NewUserBehaviorHandler(m.userBehaviorService, m.logger)
	m.favoritesHandler = handlers.NewFavoritesHandler(m.favoritesService, m.logger)
	
	m.logger.Info("Cultural wisdom handlers initialized successfully")
	return nil
}

// initGRPCServer 初始化gRPC服务器
func (m *Module) initGRPCServer() error {
	m.logger.Info("Initializing cultural wisdom gRPC server")
	
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
	
	m.logger.Info("Cultural wisdom gRPC server initialized", zap.String("address", addr))
	return nil
}

// SetupRoutes 设置HTTP路由
func (m *Module) SetupRoutes(router *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware) error {
	if !m.config.HTTPEnabled {
		m.logger.Info("HTTP routes disabled, skipping route setup")
		return nil
	}
	
	m.logger.Info("Setting up cultural wisdom HTTP routes")
	
	// 创建文化智慧路由组
	wisdomGroup := router.Group(m.config.HTTPPrefix)
	
	// 公开路由（不需要认证）
	{
		// 智慧内容相关
		wisdomGroup.GET("/list", m.wisdomHandler.GetWisdomList)
		wisdomGroup.GET("/stats", m.wisdomHandler.GetWisdomStats)
		wisdomGroup.GET("/:id", m.wisdomHandler.GetWisdomDetail) // 修正方法名
		// wisdomGroup.GET("/:id/related", m.wisdomHandler.GetRelatedWisdom) // TODO: 实现GetRelatedWisdom方法
		
		// 搜索相关
		// wisdomGroup.GET("/search", m.searchHandler.SearchWisdom) // TODO: 实现SearchWisdom方法
		wisdomGroup.GET("/search/suggestions", m.searchHandler.GetSearchSuggestions)
		
		// 分类和标签
		wisdomGroup.GET("/categories", m.categoryHandler.GetCategories)
		wisdomGroup.GET("/categories/:id", m.categoryHandler.GetCategoryByID)
		wisdomGroup.GET("/tags", m.tagHandler.GetTags)
		wisdomGroup.GET("/tags/popular", m.tagHandler.GetPopularTags)
		
		// 推荐相关
		// wisdomGroup.GET("/recommend/popular", m.recommendationHandler.GetPopularWisdom) // TODO: 实现GetPopularWisdom方法
		// wisdomGroup.GET("/recommend/featured", m.recommendationHandler.GetFeaturedWisdom) // TODO: 实现GetFeaturedWisdom方法
	}
	
	// 需要认证的路由
	if jwtMiddleware != nil {
		authGroup := wisdomGroup.Group("")
		authGroup.Use(jwtMiddleware.AuthRequired())
		{
			// 智慧内容管理
			authGroup.POST("", m.wisdomHandler.CreateWisdom)
			authGroup.PUT("/:id", m.wisdomHandler.UpdateWisdom)
			authGroup.DELETE("/:id", m.wisdomHandler.DeleteWisdom)
			// authGroup.POST("/:id/like", m.wisdomHandler.LikeWisdom) // TODO: 实现LikeWisdom方法
			// authGroup.POST("/:id/view", m.wisdomHandler.RecordView) // TODO: 实现RecordView方法
			
			// AI功能
			// authGroup.POST("/:id/interpret", m.aiHandler.InterpretWisdom) // TODO: 实现InterpretWisdom方法
			// authGroup.POST("/:id/analyze", m.aiHandler.AnalyzeWisdom) // TODO: 实现AnalyzeWisdom方法
			// authGroup.POST("/generate", m.aiHandler.GenerateWisdom) // TODO: 实现GenerateWisdom方法
			
			// 高级搜索
			authGroup.POST("/search/semantic", m.searchHandler.SemanticSearch)
			// authGroup.POST("/search/vector", m.searchHandler.VectorSearch) // TODO: 实现VectorSearch方法
			
			// 个性化推荐
			// authGroup.GET("/recommend/personal", m.recommendationHandler.GetPersonalRecommendations) // TODO: 实现GetPersonalRecommendations方法
			// authGroup.GET("/recommend/similar/:id", m.recommendationHandler.GetSimilarWisdom) // TODO: 实现GetSimilarWisdom方法
			
			// 用户行为
			authGroup.POST("/behavior", m.userBehaviorHandler.RecordBehavior)
			// authGroup.GET("/behavior/history", m.userBehaviorHandler.GetUserBehaviorHistory) // TODO: 实现GetUserBehaviorHistory方法
			// authGroup.GET("/behavior/stats", m.userBehaviorHandler.GetUserBehaviorStats) // TODO: 实现GetUserBehaviorStats方法
			
			// 收藏功能
			// authGroup.POST("/favorites", m.favoritesHandler.AddToFavorites) // TODO: 实现AddToFavorites方法
			// authGroup.DELETE("/favorites/:id", m.favoritesHandler.RemoveFromFavorites) // TODO: 实现RemoveFromFavorites方法
			authGroup.GET("/favorites", m.favoritesHandler.GetUserFavorites)
			
			// 分类和标签管理
			authGroup.POST("/categories", m.categoryHandler.CreateCategory)
			authGroup.PUT("/categories/:id", m.categoryHandler.UpdateCategory)
			authGroup.DELETE("/categories/:id", m.categoryHandler.DeleteCategory)
			authGroup.POST("/tags", m.tagHandler.CreateTag)
			authGroup.PUT("/tags/:id", m.tagHandler.UpdateTag)
			authGroup.DELETE("/tags/:id", m.tagHandler.DeleteTag)
		}
	}
	
	m.logger.Info("Cultural wisdom HTTP routes setup completed")
	return nil
}

// Start 启动模块
func (m *Module) Start() error {
	m.logger.Info("Starting cultural wisdom module")
	
	// 自动迁移数据库
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	// 启动gRPC服务器
	if m.config.GRPCEnabled && m.grpcServer != nil {
		go func() {
			m.logger.Info("Starting cultural wisdom gRPC server", 
				zap.String("address", m.grpcListener.Addr().String()))
			if err := m.grpcServer.Serve(m.grpcListener); err != nil {
				m.logger.Error("gRPC server error", zap.Error(err))
			}
		}()
	}
	
	// 启动后台任务
	go m.startBackgroundTasks()
	
	m.logger.Info("Cultural wisdom module started successfully")
	return nil
}

// Stop 停止模块
func (m *Module) Stop() error {
	m.logger.Info("Stopping cultural wisdom module")
	
	// 停止gRPC服务器
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		if m.grpcListener != nil {
			m.grpcListener.Close()
		}
	}
	
	m.logger.Info("Cultural wisdom module stopped successfully")
	return nil
}

// Health 健康检查
func (m *Module) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"module": "cultural-wisdom",
		"version": m.config.ServiceConfig.Version,
		"services": map[string]string{
			"wisdom_service":         "running",
			"ai_service":            "running",
			"search_service":        "running",
			"recommendation_service": "running",
			"cache_service":         "running",
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
	
	return health
}

// migrateDatabase 迁移数据库
func (m *Module) migrateDatabase() error {
	m.logger.Info("Migrating cultural wisdom database")
	
	// 自动迁移模型
	err := m.db.AutoMigrate(
		&models.CulturalWisdom{},
		&models.UserBehavior{},
		&models.WisdomFavorite{}, // 修复：使用正确的模型名称
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	
	m.logger.Info("Cultural wisdom database migration completed")
	return nil
}

// startBackgroundTasks 启动后台任务
func (m *Module) startBackgroundTasks() {
	m.logger.Info("Starting cultural wisdom background tasks")
	
	// 定期更新搜索索引
	go m.updateSearchIndexPeriodically()
	
	// 定期清理缓存
	go m.cleanupCachePeriodically()
	
	// 定期更新推荐模型
	go m.updateRecommendationModelPeriodically()
}

// updateSearchIndexPeriodically 定期更新搜索索引
func (m *Module) updateSearchIndexPeriodically() {
	ticker := time.NewTicker(m.config.SearchConfig.IndexUpdateInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: 实现UpdateIndex方法
		// if err := m.searchService.UpdateIndex(context.Background()); err != nil {
		//	m.logger.Error("Failed to update search index", zap.Error(err))
		// }
		m.logger.Info("Search index update not implemented yet")
	}
}

// cleanupCachePeriodically 定期清理缓存
func (m *Module) cleanupCachePeriodically() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: 实现Cleanup方法
		// if err := m.cacheService.Cleanup(context.Background()); err != nil {
		//	m.logger.Error("Failed to cleanup cache", zap.Error(err))
		// }
		m.logger.Info("Cache cleanup not implemented yet")
	}
}

// updateRecommendationModelPeriodically 定期更新推荐模型
func (m *Module) updateRecommendationModelPeriodically() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: 实现UpdateModel方法
		// if err := m.recommendationService.UpdateModel(context.Background()); err != nil {
		//	m.logger.Error("Failed to update recommendation model", zap.Error(err))
		// }
		m.logger.Info("Recommendation model update not implemented yet")
	}
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *ModuleConfig {
	return &ModuleConfig{
		HTTPEnabled: true,
		HTTPPrefix:  "/cultural-wisdom",
		GRPCEnabled: false,
		GRPCPort:    50052,
		GRPCHost:    "localhost",
		ServiceConfig: &CulturalWisdomConfig{
			ServiceName:       "cultural-wisdom-service",
			Version:          "1.0.0",
			Environment:      "development",
			MaxConcurrentReqs: 100,
			RequestTimeout:    30 * time.Second,
			MetricsRetention:  24 * time.Hour,
		},
		CacheConfig: &CacheConfig{
			DefaultTTL:   1 * time.Hour,
			WisdomTTL:    2 * time.Hour,
			SearchTTL:    30 * time.Minute,
			CategoryTTL:  4 * time.Hour,
			TagTTL:       2 * time.Hour,
			MaxCacheSize: 1000000, // 1MB
		},
		SearchConfig: &SearchConfig{
			MaxResults:          100,
			MinScore:           0.1,
			VectorDimension:    768,
			IndexUpdateInterval: 30 * time.Minute,
		},
		RecommendationConfig: &RecommendationConfig{
			MaxRecommendations: 20,
			MinSimilarity:      0.3,
			UserBehaviorWeight: 0.4,
			ContentWeight:      0.4,
			PopularityWeight:   0.2,
		},
	}
}