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

// SetupRoutes 
func SetupRoutes(router *gin.RouterGroup, db *gorm.DB, redisClient *redis.Client, logger *zap.Logger, jwtMiddleware *middleware.JWTMiddleware, providerManager *providers.Manager) {
	logger.Info("Starting cultural wisdom routes setup")
	
	defer func() {
		if r := recover(); r != nil {
			logger.Error("PANIC in SetupRoutes", zap.Any("error", r))
			panic(r) //  panic 㲶
		}
	}()
	
	// 
	logger.Info("Creating cache service")
	cacheService := services.NewCacheService(redisClient, logger)
	
	// 
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

	// ?
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

	// ?
	logger.Info("Creating cultural wisdom route group", zap.String("prefix", "/cultural-wisdom"))
	wisdomGroup := router.Group("/cultural-wisdom")
	logger.Info("Route group created successfully", zap.String("base_path", wisdomGroup.BasePath()))
	
	// GIN?
	gin.SetMode(gin.DebugMode)
	{
		// 
		logger.Info("Registering wisdom list route")
		wisdomGroup.GET("/list", wisdomHandler.GetWisdomList)
		logger.Info("Wisdom list route registered")
		
		wisdomGroup.GET("/stats", wisdomHandler.GetWisdomStats)
		logger.Info("Wisdom stats route registered")
		
		//  - ?
		logger.Info("Registering wisdom detail route", zap.String("pattern", "/detail/:id"))
		wisdomGroup.GET("/detail/:id", wisdomHandler.GetWisdomDetail)
		logger.Info("Wisdom detail route registered successfully")
		
		// ?
		wisdomGroup.GET("/search", searchHandler.FullTextSearch)
		wisdomGroup.GET("/search/semantic", searchHandler.SemanticSearch)
		wisdomGroup.GET("/search/enhanced-semantic", searchHandler.EnhancedSemanticSearch)
		// wisdomGroup.GET("/search/vector", searchHandler.VectorSearch) // TODO: VectorSearch
		wisdomGroup.GET("/search/analytics", searchHandler.GetSearchAnalytics)
		wisdomGroup.POST("/search/advanced", searchHandler.AdvancedSearch)
		wisdomGroup.GET("/search/facets", searchHandler.SearchWithFacets)
		wisdomGroup.GET("/search/filters", searchHandler.GetSearchFilters)
		wisdomGroup.GET("/search/suggestions", searchHandler.GetSearchSuggestions)
		wisdomGroup.GET("/search/popular", searchHandler.GetPopularSearches)
		wisdomGroup.GET("/categories", searchHandler.GetCategories)
		wisdomGroup.GET("/categories/:category/wisdoms", searchHandler.SearchByCategory)
		
		// ? 
		// wisdomGroup.GET("/categories", categoryHandler.GetCategories)
		wisdomGroup.GET("/category/:id", categoryHandler.GetCategoryByID)
		wisdomGroup.GET("/category/:id/stats", categoryHandler.GetCategoryStats)
		
		// ?
		wisdomGroup.GET("/tags", tagHandler.GetTags)
		wisdomGroup.GET("/tags/:id", tagHandler.GetTagByID)
		wisdomGroup.GET("/tags/popular", tagHandler.GetPopularTags)
		wisdomGroup.GET("/tags/:id/stats", tagHandler.GetTagStats)
		
		// 
		// wisdom_id
		wisdomGroup.GET("/recommendations/personalized", recommendationHandler.GetPersonalizedRecommendations)
		wisdomGroup.POST("/recommendations/batch", recommendationHandler.BatchRecommendations)
		
		// AI? ai
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
		wisdomGroup.POST("/ai/qa", aiHandler.IntelligentQA) // 
		logger.Info("AI intelligent QA route registered", zap.String("path", "/ai/qa"))
		
		// 
		behaviorGroup := wisdomGroup.Group("/user-behavior")
		{
			behaviorGroup.POST("/record", userBehaviorHandler.RecordBehavior)
			behaviorGroup.GET("/profile", userBehaviorHandler.GetUserProfile)
			behaviorGroup.GET("/similar-users", userBehaviorHandler.GetSimilarUsers)
		}
		
		// wisdom_id
		wisdomGroup.GET("/:wisdom_id/recommendations", recommendationHandler.GetRecommendations)
		wisdomGroup.GET("/:wisdom_id/similar", recommendationHandler.GetSimilarWisdoms)
	}

	// 
	logger.Info("Creating protected cultural wisdom route group")
	protectedWisdomGroup := router.Group("/cultural-wisdom")
	protectedWisdomGroup.Use(jwtMiddleware.AuthRequired())
	logger.Info("Protected route group created with JWT middleware")
	{
		protectedWisdomGroup.POST("/", wisdomHandler.CreateWisdom)
		protectedWisdomGroup.PUT("/:id", wisdomHandler.UpdateWisdom)
		protectedWisdomGroup.DELETE("/:id", wisdomHandler.DeleteWisdom)
		
		// 
		protectedWisdomGroup.POST("/batch-delete", wisdomHandler.BatchDeleteWisdom)
		
		// 
		protectedWisdomGroup.GET("/advanced-search", wisdomHandler.AdvancedSearchWisdom)
		
		// 
		protectedWisdomGroup.POST("/categories", categoryHandler.CreateCategory)
		protectedWisdomGroup.PUT("/categories/:id", categoryHandler.UpdateCategory)
		protectedWisdomGroup.DELETE("/categories/:id", categoryHandler.DeleteCategory)
		
		// 
		protectedWisdomGroup.POST("/tags", tagHandler.CreateTag)
		protectedWisdomGroup.PUT("/tags/:id", tagHandler.UpdateTag)
		protectedWisdomGroup.DELETE("/tags/:id", tagHandler.DeleteTag)
		
		// 
		favoritesGroup := protectedWisdomGroup.Group("/favorites")
		{
			favoritesGroup.POST("", favoritesHandler.AddFavorite)                    // 
			favoritesGroup.DELETE("/:wisdom_id", favoritesHandler.RemoveFavorite)    // 
			favoritesGroup.GET("", favoritesHandler.GetUserFavorites)               // 
			favoritesGroup.GET("/:wisdom_id/status", favoritesHandler.CheckFavoriteStatus) // ?
		}
		
		// 
		notesGroup := protectedWisdomGroup.Group("/notes")
		{
			notesGroup.POST("", favoritesHandler.CreateNote)                        // 
			notesGroup.PUT("/:wisdom_id", favoritesHandler.UpdateNote)              // 
			notesGroup.GET("/:wisdom_id", favoritesHandler.GetNote)                 // 
			notesGroup.GET("", favoritesHandler.GetUserNotes)                       // 
			notesGroup.DELETE("/:wisdom_id", favoritesHandler.DeleteNote)           // 
		}
	}
	
	logger.Info("Cultural wisdom routes setup completed successfully")
}

