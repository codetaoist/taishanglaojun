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

// Module 
type Module struct {
	// 
	db              *gorm.DB
	redisClient     *redis.Client
	logger          *zap.Logger
	providerManager *providers.Manager
	
	// ?
	wisdomService         *services.WisdomService
	aiService            *services.AIService
	searchService        *services.SearchService
	recommendationService *services.RecommendationService
	categoryService      *services.CategoryService
	tagService           *services.TagService
	userBehaviorService  *services.UserBehaviorService
	favoritesService     *services.FavoritesService
	cacheService         *services.CacheService
	
	// 
	wisdomHandler         *handlers.WisdomHandler
	aiHandler            *handlers.AIHandler
	searchHandler        *handlers.SearchHandler
	recommendationHandler *handlers.RecommendationHandler
	categoryHandler      *handlers.CategoryHandler
	tagHandler           *handlers.TagHandler
	userBehaviorHandler  *handlers.UserBehaviorHandler
	favoritesHandler     *handlers.FavoritesHandler
	
	// gRPC?
	grpcServer   *grpc.Server
	grpcListener net.Listener
	
	// 
	config *ModuleConfig
}

// ModuleConfig 
type ModuleConfig struct {
	// HTTP
	HTTPEnabled bool   `json:"http_enabled"`
	HTTPPrefix  string `json:"http_prefix"`
	
	// gRPC
	GRPCEnabled bool   `json:"grpc_enabled"`
	GRPCPort    int    `json:"grpc_port"`
	GRPCHost    string `json:"grpc_host"`
	
	// 
	ServiceConfig *CulturalWisdomConfig `json:"service_config"`
	
	// 
	CacheConfig *CacheConfig `json:"cache_config"`
	
	// 
	SearchConfig *SearchConfig `json:"search_config"`
	
	// 
	RecommendationConfig *RecommendationConfig `json:"recommendation_config"`
}

// CulturalWisdomConfig 
type CulturalWisdomConfig struct {
	ServiceName        string        `json:"service_name"`
	Version           string        `json:"version"`
	Environment       string        `json:"environment"`
	MaxConcurrentReqs int           `json:"max_concurrent_requests"`
	RequestTimeout    time.Duration `json:"request_timeout"`
	MetricsRetention  time.Duration `json:"metrics_retention"`
}

// CacheConfig 
type CacheConfig struct {
	DefaultTTL    time.Duration `json:"default_ttl"`
	WisdomTTL     time.Duration `json:"wisdom_ttl"`
	SearchTTL     time.Duration `json:"search_ttl"`
	CategoryTTL   time.Duration `json:"category_ttl"`
	TagTTL        time.Duration `json:"tag_ttl"`
	MaxCacheSize  int64         `json:"max_cache_size"`
}

// SearchConfig 
type SearchConfig struct {
	MaxResults        int     `json:"max_results"`
	MinScore          float64 `json:"min_score"`
	VectorDimension   int     `json:"vector_dimension"`
	IndexUpdateInterval time.Duration `json:"index_update_interval"`
}

// RecommendationConfig 
type RecommendationConfig struct {
	MaxRecommendations int     `json:"max_recommendations"`
	MinSimilarity      float64 `json:"min_similarity"`
	UserBehaviorWeight float64 `json:"user_behavior_weight"`
	ContentWeight      float64 `json:"content_weight"`
	PopularityWeight   float64 `json:"popularity_weight"`
}

// NewModule 
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
	
	// 
	if err := module.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	// ?
	if err := module.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}
	
	// gRPC?
	if config.GRPCEnabled {
		if err := module.initGRPCServer(); err != nil {
			return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
		}
	}
	
	return module, nil
}

// initServices 
func (m *Module) initServices() error {
	m.logger.Info("Initializing cultural wisdom services")
	
	// 
	m.cacheService = services.NewCacheService(m.redisClient, m.logger)
	
	// 
	m.wisdomService = services.NewWisdomService(m.db, m.cacheService)
	m.aiService = services.NewAIService(m.db, m.logger, m.providerManager)
	
	// AI
	aiIntegrationService := aiServices.NewAIService(m.providerManager)
	
	// 
	m.searchService = services.NewSearchService(m.db, m.cacheService, aiIntegrationService, m.logger)
	
	// 
	m.userBehaviorService = services.NewUserBehaviorService(m.db, m.cacheService, m.logger)
	
	// 
	m.recommendationService = services.NewRecommendationService(m.db, m.cacheService, m.userBehaviorService, m.aiService, m.logger)
	
	// ?
	m.categoryService = services.NewCategoryService(m.db, m.cacheService)
	m.tagService = services.NewTagService(m.db, m.cacheService)
	
	// 
	m.favoritesService = services.NewFavoritesService(m.db, m.logger)
	
	m.logger.Info("Cultural wisdom services initialized successfully")
	return nil
}

// initHandlers ?
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

// initGRPCServer gRPC?
func (m *Module) initGRPCServer() error {
	m.logger.Info("Initializing cultural wisdom gRPC server")
	
	// gRPC?
	m.grpcServer = grpc.NewServer()
	
	// gRPC
	// TODO: gRPC?
	
	// ?
	addr := fmt.Sprintf("%s:%d", m.config.GRPCHost, m.config.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	m.grpcListener = listener
	
	m.logger.Info("Cultural wisdom gRPC server initialized", zap.String("address", addr))
	return nil
}

// SetupRoutes HTTP
func (m *Module) SetupRoutes(router *gin.RouterGroup, jwtMiddleware *middleware.JWTMiddleware) error {
	if !m.config.HTTPEnabled {
		m.logger.Info("HTTP routes disabled, skipping route setup")
		return nil
	}
	
	m.logger.Info("Setting up cultural wisdom HTTP routes")
	
	// ?
	wisdomGroup := router.Group(m.config.HTTPPrefix)
	
	// 
	{
		// 
		wisdomGroup.GET("/list", m.wisdomHandler.GetWisdomList)
		wisdomGroup.GET("/stats", m.wisdomHandler.GetWisdomStats)
		wisdomGroup.GET("/:id", m.wisdomHandler.GetWisdomDetail) // ?
		// wisdomGroup.GET("/:id/related", m.wisdomHandler.GetRelatedWisdom) // TODO: GetRelatedWisdom
		
		// 
		// wisdomGroup.GET("/search", m.searchHandler.SearchWisdom) // TODO: SearchWisdom
		wisdomGroup.GET("/search/suggestions", m.searchHandler.GetSearchSuggestions)
		
		// ?
		wisdomGroup.GET("/categories", m.categoryHandler.GetCategories)
		wisdomGroup.GET("/categories/:id", m.categoryHandler.GetCategoryByID)
		wisdomGroup.GET("/tags", m.tagHandler.GetTags)
		wisdomGroup.GET("/tags/popular", m.tagHandler.GetPopularTags)
		
		// 
		// wisdomGroup.GET("/recommend/popular", m.recommendationHandler.GetPopularWisdom) // TODO: GetPopularWisdom
		// wisdomGroup.GET("/recommend/featured", m.recommendationHandler.GetFeaturedWisdom) // TODO: GetFeaturedWisdom
	}
	
	// 
	if jwtMiddleware != nil {
		authGroup := wisdomGroup.Group("")
		authGroup.Use(jwtMiddleware.AuthRequired())
		{
			// 
			authGroup.POST("", m.wisdomHandler.CreateWisdom)
			authGroup.PUT("/:id", m.wisdomHandler.UpdateWisdom)
			authGroup.DELETE("/:id", m.wisdomHandler.DeleteWisdom)
			// authGroup.POST("/:id/like", m.wisdomHandler.LikeWisdom) // TODO: LikeWisdom
			// authGroup.POST("/:id/view", m.wisdomHandler.RecordView) // TODO: RecordView
			
			// AI
			// authGroup.POST("/:id/interpret", m.aiHandler.InterpretWisdom) // TODO: InterpretWisdom
			// authGroup.POST("/:id/analyze", m.aiHandler.AnalyzeWisdom) // TODO: AnalyzeWisdom
			// authGroup.POST("/generate", m.aiHandler.GenerateWisdom) // TODO: GenerateWisdom
			
			// 
			authGroup.POST("/search/semantic", m.searchHandler.SemanticSearch)
			// authGroup.POST("/search/vector", m.searchHandler.VectorSearch) // TODO: VectorSearch
			
			// 
			// authGroup.GET("/recommend/personal", m.recommendationHandler.GetPersonalRecommendations) // TODO: GetPersonalRecommendations
			// authGroup.GET("/recommend/similar/:id", m.recommendationHandler.GetSimilarWisdom) // TODO: GetSimilarWisdom
			
			// 
			authGroup.POST("/behavior", m.userBehaviorHandler.RecordBehavior)
			// authGroup.GET("/behavior/history", m.userBehaviorHandler.GetUserBehaviorHistory) // TODO: GetUserBehaviorHistory
			// authGroup.GET("/behavior/stats", m.userBehaviorHandler.GetUserBehaviorStats) // TODO: GetUserBehaviorStats
			
			// 
			// authGroup.POST("/favorites", m.favoritesHandler.AddToFavorites) // TODO: AddToFavorites
			// authGroup.DELETE("/favorites/:id", m.favoritesHandler.RemoveFromFavorites) // TODO: RemoveFromFavorites
			authGroup.GET("/favorites", m.favoritesHandler.GetUserFavorites)
			
			// ?
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

// Start 
func (m *Module) Start() error {
	m.logger.Info("Starting cultural wisdom module")
	
	// ?
	if err := m.migrateDatabase(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	// gRPC?
	if m.config.GRPCEnabled && m.grpcServer != nil {
		go func() {
			m.logger.Info("Starting cultural wisdom gRPC server", 
				zap.String("address", m.grpcListener.Addr().String()))
			if err := m.grpcServer.Serve(m.grpcListener); err != nil {
				m.logger.Error("gRPC server error", zap.Error(err))
			}
		}()
	}
	
	// 
	go m.startBackgroundTasks()
	
	m.logger.Info("Cultural wisdom module started successfully")
	return nil
}

// Stop 
func (m *Module) Stop() error {
	m.logger.Info("Stopping cultural wisdom module")
	
	// gRPC?
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
		if m.grpcListener != nil {
			m.grpcListener.Close()
		}
	}
	
	m.logger.Info("Cultural wisdom module stopped successfully")
	return nil
}

// Health ?
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
	
	// 
	if sqlDB, err := m.db.DB(); err == nil {
		if err := sqlDB.Ping(); err != nil {
			health["database"] = "unhealthy"
			health["status"] = "degraded"
		} else {
			health["database"] = "healthy"
		}
	}
	
	// Redis
	if err := m.redisClient.Ping(context.Background()).Err(); err != nil {
		health["redis"] = "unhealthy"
		health["status"] = "degraded"
	} else {
		health["redis"] = "healthy"
	}
	
	return health
}

// migrateDatabase ?
func (m *Module) migrateDatabase() error {
	m.logger.Info("Migrating cultural wisdom database")
	
	// 
	err := m.db.AutoMigrate(
		&models.CulturalWisdom{},
		&models.UserBehavior{},
		&models.WisdomFavorite{}, // 
	)
	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}
	
	m.logger.Info("Cultural wisdom database migration completed")
	return nil
}

// startBackgroundTasks 
func (m *Module) startBackgroundTasks() {
	m.logger.Info("Starting cultural wisdom background tasks")
	
	// 
	go m.updateSearchIndexPeriodically()
	
	// 
	go m.cleanupCachePeriodically()
	
	// 
	go m.updateRecommendationModelPeriodically()
}

// updateSearchIndexPeriodically 
func (m *Module) updateSearchIndexPeriodically() {
	ticker := time.NewTicker(m.config.SearchConfig.IndexUpdateInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: UpdateIndex
		// if err := m.searchService.UpdateIndex(context.Background()); err != nil {
		//	m.logger.Error("Failed to update search index", zap.Error(err))
		// }
		m.logger.Info("Search index update not implemented yet")
	}
}

// cleanupCachePeriodically 
func (m *Module) cleanupCachePeriodically() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: Cleanup
		// if err := m.cacheService.Cleanup(context.Background()); err != nil {
		//	m.logger.Error("Failed to cleanup cache", zap.Error(err))
		// }
		m.logger.Info("Cache cleanup not implemented yet")
	}
}

// updateRecommendationModelPeriodically 
func (m *Module) updateRecommendationModelPeriodically() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		// TODO: UpdateModel
		// if err := m.recommendationService.UpdateModel(context.Background()); err != nil {
		//	m.logger.Error("Failed to update recommendation model", zap.Error(err))
		// }
		m.logger.Info("Recommendation model update not implemented yet")
	}
}

// getDefaultConfig 
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

