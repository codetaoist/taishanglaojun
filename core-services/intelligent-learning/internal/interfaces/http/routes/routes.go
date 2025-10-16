package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	learnerservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/learner"
	contentservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/content"
	knowledgeservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/knowledge"
	adaptiveservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/adaptive"
	analyticsservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
	recommendationservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/recommendation"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/handlers"
	httphandlers "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/http/handlers"
	httpinterfaces "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/http"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/http/middleware"
	websockethandlers "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/websocket"
	realtimeservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RouterConfig 
type RouterConfig struct {
	LearnerService                    *learnerservices.LearnerService
	ContentService                    *contentservices.ContentService
	KnowledgeGraphService             *knowledgeservices.KnowledgeGraphAppService
	ProgressTrackingService           *learnerservices.ProgressTrackingService
	AdaptiveLearningService           *adaptiveservices.AdaptiveLearningService
	KnowledgeAnalysisService          *knowledgeservices.KnowledgeAnalysisService
	RealtimeAnalyticsService          *realtimeservices.RealtimeAnalyticsService
	PersonalizationEngine             *realtimeservices.PersonalizationEngine
	UserBehaviorTracker               *realtimeservices.UserBehaviorTracker
	PreferenceAnalyzer                *realtimeservices.PreferenceAnalyzer
	ContextAnalyzer                   *realtimeservices.ContextAnalyzer
	RealtimeRecommendationService     *recommendationservices.RealtimeRecommendationService
	RecommendationIntegrationService  *recommendationservices.RecommendationIntegrationService
}

// SetupRoutes 
func SetupRoutes(config *RouterConfig) *gin.Engine {
	// Gin
	gin.SetMode(gin.ReleaseMode)

	// Gin
	r := gin.New()

	// ?
	setupMiddleware(r)

	// ?
	learnerHandler := httphandlers.NewLearnerHandler(config.LearnerService)
	contentHandler := httphandlers.NewContentHandler(config.ContentService)
	kgHandler := httphandlers.NewKnowledgeGraphHandler(config.KnowledgeGraphService)
	progressHandler := httpinterfaces.NewProgressHandler(config.ProgressTrackingService)
	adaptiveHandler := httphandlers.NewAdaptiveLearningHandler(config.AdaptiveLearningService)
	knowledgeAnalysisHandler := handlers.NewKnowledgeAnalysisHandler(config.KnowledgeAnalysisService)
	progressWebSocketHandler := websockethandlers.NewProgressWebSocketHandler(config.ProgressTrackingService)
	realtimeAnalyticsHandler := httphandlers.NewRealtimeAnalyticsHandler(config.RealtimeAnalyticsService)
	
	// RouterConfig
	// learningPathHandler := httphandlers.NewLearningPathHandler(config.LearningPathService)
	
	// ?
	recommendationHandler := httphandlers.NewRecommendationHandler(
		config.PersonalizationEngine,
		config.UserBehaviorTracker,
		config.PreferenceAnalyzer,
		config.ContextAnalyzer,
	)

	// ?
	realtimeRecommendationHandler := httphandlers.NewRealtimeRecommendationHandler(
		config.RealtimeRecommendationService,
	)

	// ?
	recommendationIntegrationHandler := httphandlers.NewRecommendationIntegrationHandler(
		config.RecommendationIntegrationService,
	)

	// API?
	api := r.Group("/api/v1")
	
	// API
	setupAPIRoutes(api, learnerHandler, contentHandler, kgHandler, progressHandler, adaptiveHandler, knowledgeAnalysisHandler, progressWebSocketHandler, realtimeAnalyticsHandler, recommendationHandler, realtimeRecommendationHandler, recommendationIntegrationHandler)

	// Swagger
	setupSwagger(r)

	// ?
	setupHealthCheck(r)

	return r
}

// setupMiddleware ?
func setupMiddleware(r *gin.Engine) {
	// ?
	r.Use(gin.Recovery())

	// ID?
	r.Use(requestid.New())

	// ?
	r.Use(middleware.Logger())

	// CORS?
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ?
	r.Use(middleware.RateLimit())

	// 
	// r.Use(middleware.Auth())
}

// setupAPIRoutes API
func setupAPIRoutes(
	r *gin.RouterGroup,
	learnerHandler *httphandlers.LearnerHandler,
	contentHandler *httphandlers.ContentHandler,
	kgHandler *httphandlers.KnowledgeGraphHandler,
	progressHandler *httpinterfaces.ProgressHandler,
	adaptiveHandler *httphandlers.AdaptiveLearningHandler,
	knowledgeAnalysisHandler *handlers.KnowledgeAnalysisHandler,
	progressWebSocketHandler *websockethandlers.ProgressWebSocketHandler,
	realtimeAnalyticsHandler *httphandlers.RealtimeAnalyticsHandler,
	recommendationHandler *httphandlers.RecommendationHandler,
	realtimeRecommendationHandler *httphandlers.RealtimeRecommendationHandler,
	recommendationIntegrationHandler *httphandlers.RecommendationIntegrationHandler,
) {

	// API v1 ?
	v1 := r.Group("/api/v1")
	{
		// ?
		learners := v1.Group("/learners")
		{
			learners.POST("", learnerHandler.CreateLearner)
			learners.GET("/:id", learnerHandler.GetLearner)
			learners.PUT("/:id", learnerHandler.UpdateLearner)
			learners.DELETE("/:id", learnerHandler.DeleteLearner)
			
			// 
			learners.POST("/:id/goals", learnerHandler.AddLearningGoal)
			learners.PUT("/:id/goals/:goalId", learnerHandler.UpdateLearningGoal)
			
			// ?
			learners.PUT("/:id/skills", learnerHandler.UpdateSkill)
			
			// 
			learners.POST("/:id/activities", learnerHandler.RecordLearningActivity)
			learners.GET("/:id/history", learnerHandler.GetLearningHistory)
			
			// ?
			learners.GET("/:id/analytics", learnerHandler.GetLearningAnalytics)
			learners.GET("/:id/recommendations", learnerHandler.GetPersonalizedRecommendations)
		}

		// 
		content := v1.Group("/content")
		{
			content.POST("", contentHandler.CreateContent)
			content.GET("", contentHandler.ListContent)
			content.GET("/:id", contentHandler.GetContent)
			content.PUT("/:id", contentHandler.UpdateContent)
			content.DELETE("/:id", contentHandler.DeleteContent)
			
			// ?
			content.POST("/:id/publish", contentHandler.PublishContent)
			content.POST("/:id/archive", contentHandler.ArchiveContent)
			
			// ?
			content.GET("/search", contentHandler.SearchContent)
			content.GET("/personalized", contentHandler.GetPersonalizedContent)
			
			// ?
			content.POST("/:id/interactions", contentHandler.RecordContentInteraction)
			content.GET("/:id/analytics", contentHandler.GetContentAnalytics)
			
			// ?
			content.GET("/knowledge-node/:nodeId", contentHandler.GetContentByKnowledgeNode)
		}

		// 
		kg := v1.Group("/knowledge-graph")
		{
			// 
			nodes := kg.Group("/nodes")
			{
				nodes.POST("", kgHandler.CreateNode)
				nodes.GET("/search", kgHandler.SearchNodes)
				nodes.GET("/:id", kgHandler.GetNode)
				nodes.PUT("/:id", kgHandler.UpdateNode)
				nodes.DELETE("/:id", kgHandler.DeleteNode)
				nodes.GET("/:id/neighbors", kgHandler.GetNodeNeighbors)
			}

			// 
			relations := kg.Group("/relations")
			{
				relations.POST("", kgHandler.CreateRelation)
				relations.GET("/:id", kgHandler.GetRelation)
				relations.PUT("/:id", kgHandler.UpdateRelation)
				relations.DELETE("/:id", kgHandler.DeleteRelation)
			}

			// 
			kg.GET("/shortest-path", kgHandler.FindShortestPath)
			kg.POST("/learning-path", kgHandler.GenerateLearningPath)
			kg.POST("/concept-map", kgHandler.GenerateConceptMap)
			kg.POST("/analyze", kgHandler.AnalyzeGraph)
			kg.GET("/statistics", kgHandler.GetGraphStatistics)
			kg.POST("/validate", kgHandler.ValidateGraph)
		}

		// 
		progress := v1.Group("/progress")
		{
			// 
			progress.POST("/update", progressHandler.UpdateProgress)
			progress.POST("/batch-update", progressHandler.BatchUpdateProgress)
			
			// 
			progress.GET("/learner/:learnerId", progressHandler.GetProgressSummary)
			progress.GET("/content/:contentId/learner/:learnerId", progressHandler.GetContentProgress)
			
			// 
			progress.GET("/report/:learnerId", progressHandler.GetLearningReport)
			
			// WebSocket
			progress.GET("/ws/:learnerId", progressWebSocketHandler.HandleWebSocket)
		}

		// 
		adaptive := v1.Group("/adaptive")
		{
			// 
			adaptive.POST("/adapt-path", adaptiveHandler.AdaptLearningPath)
			
			// 
			adaptive.GET("/recommendations/:learner_id", adaptiveHandler.GetAdaptationRecommendations)
			
			// 
			adaptive.GET("/history/:learner_id", adaptiveHandler.GetLearnerAdaptationHistory)
			
			// 
			adaptive.POST("/analyze-effectiveness", adaptiveHandler.AnalyzeLearningEffectiveness)
			
			// 
			adaptive.POST("/predict-outcome", adaptiveHandler.PredictLearningOutcome)
		}

		// 
		analysis := v1.Group("/knowledge-analysis")
		{
			// 
			analysis.POST("/concept-relationships", knowledgeAnalysisHandler.AnalyzeConceptRelationships)
			
			// ?
			analysis.POST("/dependency-graph", knowledgeAnalysisHandler.BuildDependencyGraph)
			
			// 
			analysis.POST("/recommend-content", knowledgeAnalysisHandler.RecommendContent)
			
			// 
			analysis.GET("/concept-clusters", knowledgeAnalysisHandler.GetConceptClusters)
			
			// 
			analysis.GET("/learning-path", knowledgeAnalysisHandler.GetLearningPath)
			
			// 
			analysis.GET("/personalized-recommendations", knowledgeAnalysisHandler.GetPersonalizedRecommendations)
			
			// 
			analysis.POST("/knowledge-gaps", knowledgeAnalysisHandler.AnalyzeKnowledgeGaps)
		}

		// 
		realtime := v1.Group("/realtime-analytics")
		{
			// 
			realtime.POST("/events", realtimeAnalyticsHandler.ProcessEvent)
			
			// 
			realtime.GET("/:learnerId/data", realtimeAnalyticsHandler.GetRealtimeData)
			realtime.GET("/:learnerId/metrics", realtimeAnalyticsHandler.GetAnalyticsMetrics)
			realtime.GET("/:learnerId/insights", realtimeAnalyticsHandler.GetLearningInsights)
			realtime.GET("/:learnerId/session", realtimeAnalyticsHandler.GetSessionAnalytics)
			realtime.GET("/:learnerId/performance", realtimeAnalyticsHandler.GetPerformanceTrends)
			realtime.GET("/:learnerId/alerts", realtimeAnalyticsHandler.GetAlerts)
			realtime.GET("/:learnerId/recommendations", realtimeAnalyticsHandler.GetRecommendations)
			
			// ?
			realtime.POST("/analyzers", realtimeAnalyticsHandler.CreateAnalyzer)
			
			// WebSocket
			realtime.GET("/subscribe", realtimeAnalyticsHandler.SubscribeToUpdates)
		}

		// 
		recommendations := v1.Group("/recommendations")
		{
			// 
			recommendations.POST("/personalized", recommendationHandler.GetPersonalizedRecommendations)
			
			// 
			recommendations.GET("/strategy/:strategy", recommendationHandler.GetRecommendationsByStrategy)
			
			// 
			recommendations.POST("/batch", recommendationHandler.BatchRecommendations)
			
			// 
			recommendations.POST("/behavior", recommendationHandler.RecordUserBehavior)
			
			// 
			recommendations.GET("/preferences/:user_id", recommendationHandler.GetUserPreferences)
			
			// ?
			recommendations.GET("/context/:user_id", recommendationHandler.GetLearningContext)
			
			// 
		recommendations.GET("/insights/:user_id", recommendationHandler.GetBehaviorInsights)
	}

	// 
	realtimeRec := v1.Group("/realtime-recommendations")
	{
		// 
		realtimeRec.POST("/events", realtimeRecommendationHandler.ProcessRealtimeEvent)
		realtimeRec.POST("/events/batch", realtimeRecommendationHandler.BatchProcessEvents)
		
		// 
		realtimeRec.GET("/:user_id", realtimeRecommendationHandler.GetRealtimeRecommendations)
		
		// 
		realtimeRec.GET("/sessions/:user_id", realtimeRecommendationHandler.GetUserSession)
		
		// WebSocket
		realtimeRec.GET("/subscribe", realtimeRecommendationHandler.SubscribeToRecommendationUpdates)
		
		// 
		realtimeRec.GET("/metrics", realtimeRecommendationHandler.GetRecommendationMetrics)
	}

	// 
	integratedRec := v1.Group("/integrated-recommendations")
	{
		// 
		integratedRec.GET("/:user_id", recommendationIntegrationHandler.GetIntegratedRecommendations)
		
		// 
		integratedRec.POST("/batch", recommendationIntegrationHandler.BatchGetRecommendations)
		
		// 
		integratedRec.GET("/metrics", recommendationIntegrationHandler.GetRecommendationMetrics)
		
		// 
		integratedRec.DELETE("/cache", recommendationIntegrationHandler.ClearRecommendationCache)
	}
	}
}

// setupSwagger Swagger
func setupSwagger(r *gin.Engine) {
	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// Swagger
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

// setupHealthCheck ?
func setupHealthCheck(r *gin.Engine) {
	// ?
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "intelligent-learning",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// ?
	r.GET("/ready", func(c *gin.Context) {
		// 
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// ?
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})
}

