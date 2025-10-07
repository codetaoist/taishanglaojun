package cultural_wisdom

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/services"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	aiServices "github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/codetaoist/taishanglaojun/core-services/internal/middleware"
)

// SetupRoutes 设置文化智慧路由
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger, jwtMiddleware *middleware.JWTMiddleware, providerManager *providers.Manager) {
	logger.Info("Starting cultural wisdom routes setup")
	
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC in SetupRoutes", zap.Any("error", r))
			panic(r) // 重新抛出 panic 以便上层捕获
		}
	}()
	
	// 创建缓存服务
	logger.Info("Creating cache service")
	cacheService := services.NewCacheService(redisClient, logger)
	
	// 创建服务实例
	logger.Info("Creating wisdom service")
	wisdomService := services.NewWisdomService(db, cacheService)
	logger.Info("Creating AI service")
	aiService := services.NewAIService(db, logger, providerManager)
	logger.Info("Creating AI integration service")
	aiIntegrationService := aiServices.NewAIService(providerManager)
	logger.Info("Creating search service")
	searchService := services.NewSearchService(db, cacheService, aiIntegrationService, logger)
	logger.Info("Creating user behavior service")
	userBehaviorService := services.NewUserBehaviorService(db, cacheService, logger)
	logger.Info("Creating recommendation service")
	recommendationService := services.NewRecommendationService(db, cacheService, userBehaviorService, aiService, logger)
	logger.Info("Creating category service")
	categoryService := services.NewCategoryService(db, cacheService)
	logger.Info("Creating tag service")
	tagService := services.NewTagService(db, cacheService)
	logger.Info("Creating favorites service")
	favoritesService := services.NewFavoritesService(db, logger)

	// 创建处理器实例
	logger.Info("Creating handlers")
	wisdomHandler := handlers.NewWisdomHandler(wisdomService)
	searchHandler := handlers.NewSearchHandler(searchService)
	aiHandler := handlers.NewAIHandler(aiService, logger)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationService, logger)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	tagHandler := handlers.NewTagHandler(tagService)
	userBehaviorHandler := handlers.NewUserBehaviorHandler(userBehaviorService, logger)
	favoritesHandler := handlers.NewFavoritesHandler(favoritesService, logger)
	
	logger.Info("Setting up routes")

	// 文化智慧路由组
	logger.Info("Creating cultural wisdom route group", zap.String("prefix", "/cultural-wisdom"))
	wisdomGroup := router.Group("/cultural-wisdom")
	logger.Info("Route group created successfully", zap.String("base_path", wisdomGroup.BasePath()))
	
	// 强制启用GIN调试模式以显示路由注册
	gin.SetMode(gin.DebugMode)
	{
		// 公开路由（不需要认证）
		logger.Info("Registering wisdom list route")
		wisdomGroup.GET("/list", wisdomHandler.GetWisdomList)
		logger.Info("Wisdom list route registered")
		
		wisdomGroup.GET("/stats", wisdomHandler.GetWisdomStats)
		logger.Info("Wisdom stats route registered")
		
		// 详情路由 - 放在前面，使用更具体的路径模式
		logger.Info("Registering wisdom detail route", zap.String("pattern", "/detail/:id"))
		wisdomGroup.GET("/detail/:id", wisdomHandler.GetWisdomDetail)
		logger.Info("Wisdom detail route registered successfully")
		
		// 搜索相关路由（公开）
		wisdomGroup.GET("/search", searchHandler.FullTextSearch)
		wisdomGroup.GET("/search/semantic", searchHandler.SemanticSearch)
		wisdomGroup.GET("/search/enhanced-semantic", searchHandler.EnhancedSemanticSearch)
		// wisdomGroup.GET("/search/vector", searchHandler.VectorSearch) // TODO: 实现VectorSearch方法
		wisdomGroup.GET("/search/analytics", searchHandler.GetSearchAnalytics)
		wisdomGroup.POST("/search/advanced", searchHandler.AdvancedSearch)
		wisdomGroup.GET("/search/facets", searchHandler.SearchWithFacets)
		wisdomGroup.GET("/search/filters", searchHandler.GetSearchFilters)
		wisdomGroup.GET("/search/suggestions", searchHandler.GetSearchSuggestions)
		wisdomGroup.GET("/search/popular", searchHandler.GetPopularSearches)
		wisdomGroup.GET("/categories", searchHandler.GetCategories)
		wisdomGroup.GET("/categories/:category/wisdoms", searchHandler.SearchByCategory)
		
		// 分类管理路由（公开）- 修复路径冲突问题
		// wisdomGroup.GET("/categories", categoryHandler.GetCategories)
		wisdomGroup.GET("/category/:id", categoryHandler.GetCategoryByID)
		wisdomGroup.GET("/category/:id/stats", categoryHandler.GetCategoryStats)
		
		// 标签管理路由（公开）
		wisdomGroup.GET("/tags", tagHandler.GetTags)
		wisdomGroup.GET("/tags/:id", tagHandler.GetTagByID)
		wisdomGroup.GET("/tags/popular", tagHandler.GetPopularTags)
		wisdomGroup.GET("/tags/:id/stats", tagHandler.GetTagStats)
		
		// 推荐系统路由
		// 个性化推荐（不依赖wisdom_id的路由）
		wisdomGroup.GET("/recommendations/personalized", recommendationHandler.GetPersonalizedRecommendations)
		wisdomGroup.POST("/recommendations/batch", recommendationHandler.BatchRecommendations)
		
		// AI功能路由（公开）- 使用ai前缀避免冲突
		logger.Info("Registering AI routes")
		wisdomGroup.POST("/ai/:wisdom_id/interpret", aiHandler.InterpretWisdom)
		logger.Info("AI interpret route registered", zap.String("path", "/ai/:wisdom_id/interpret"))
		wisdomGroup.GET("/ai/:wisdom_id/recommend", aiHandler.RecommendWisdom)
		logger.Info("AI recommend route registered", zap.String("path", "/ai/:wisdom_id/recommend"))
		wisdomGroup.GET("/ai/:wisdom_id/analysis", aiHandler.GetAIAnalysis)
		logger.Info("AI analysis route registered", zap.String("path", "/ai/:wisdom_id/analysis"))
		wisdomGroup.POST("/ai/:wisdom_id/depth-analysis", aiHandler.AnalyzeWisdomInDepth)
		logger.Info("AI depth analysis route registered", zap.String("path", "/ai/:wisdom_id/depth-analysis"))
		wisdomGroup.POST("/ai/batch-recommend", aiHandler.BatchRecommend)
		logger.Info("AI batch recommend route registered", zap.String("path", "/ai/batch-recommend"))
		wisdomGroup.POST("/ai/qa", aiHandler.IntelligentQA) // 智能问答
		logger.Info("AI intelligent QA route registered", zap.String("path", "/ai/qa"))
		
		// 用户行为路由
		behaviorGroup := wisdomGroup.Group("/user-behavior")
		{
			behaviorGroup.POST("/record", userBehaviorHandler.RecordBehavior)
			behaviorGroup.GET("/profile", userBehaviorHandler.GetUserProfile)
			behaviorGroup.GET("/similar-users", userBehaviorHandler.GetSimilarUsers)
		}
		
		// 智慧推荐（依赖wisdom_id的路由）
		wisdomGroup.GET("/:wisdom_id/recommendations", recommendationHandler.GetRecommendations)
		wisdomGroup.GET("/:wisdom_id/similar", recommendationHandler.GetSimilarWisdoms)
	}

	// 需要认证的路由
	logger.Info("Creating protected cultural wisdom route group")
	protectedWisdomGroup := router.Group("/cultural-wisdom")
	protectedWisdomGroup.Use(jwtMiddleware.AuthRequired())
	logger.Info("Protected route group created with JWT middleware")
	{
		protectedWisdomGroup.POST("/", wisdomHandler.CreateWisdom)
		protectedWisdomGroup.PUT("/:id", wisdomHandler.UpdateWisdom)
		protectedWisdomGroup.DELETE("/:id", wisdomHandler.DeleteWisdom)
		
		// 批量操作路由
		protectedWisdomGroup.POST("/batch-delete", wisdomHandler.BatchDeleteWisdom)
		
		// 高级搜索路由
		protectedWisdomGroup.GET("/advanced-search", wisdomHandler.AdvancedSearchWisdom)
		
		// 需要认证的分类管理路由
		protectedWisdomGroup.POST("/categories", categoryHandler.CreateCategory)
		protectedWisdomGroup.PUT("/categories/:id", categoryHandler.UpdateCategory)
		protectedWisdomGroup.DELETE("/categories/:id", categoryHandler.DeleteCategory)
		
		// 需要认证的标签管理路由
		protectedWisdomGroup.POST("/tags", tagHandler.CreateTag)
		protectedWisdomGroup.PUT("/tags/:id", tagHandler.UpdateTag)
		protectedWisdomGroup.DELETE("/tags/:id", tagHandler.DeleteTag)
		
		// 收藏功能路由
		favoritesGroup := protectedWisdomGroup.Group("/favorites")
		{
			favoritesGroup.POST("", favoritesHandler.AddFavorite)                    // 添加收藏
			favoritesGroup.DELETE("/:wisdom_id", favoritesHandler.RemoveFavorite)    // 移除收藏
			favoritesGroup.GET("", favoritesHandler.GetUserFavorites)               // 获取用户收藏列表
			favoritesGroup.GET("/:wisdom_id/status", favoritesHandler.CheckFavoriteStatus) // 检查收藏状态
		}
		
		// 笔记功能路由
		notesGroup := protectedWisdomGroup.Group("/notes")
		{
			notesGroup.POST("", favoritesHandler.CreateNote)                        // 创建笔记
			notesGroup.PUT("/:wisdom_id", favoritesHandler.UpdateNote)              // 更新笔记
			notesGroup.GET("/:wisdom_id", favoritesHandler.GetNote)                 // 获取笔记
			notesGroup.GET("", favoritesHandler.GetUserNotes)                       // 获取用户笔记列表
			notesGroup.DELETE("/:wisdom_id", favoritesHandler.DeleteNote)           // 删除笔记
		}
	}
	
	logger.Info("Cultural wisdom routes setup completed successfully")
}