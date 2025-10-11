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

// RouterConfig 路由配置
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

// SetupRoutes 设置路由
func SetupRoutes(config *RouterConfig) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.New()

	// 添加中间�?
	setupMiddleware(r)

	// 创建处理�?
	learnerHandler := httphandlers.NewLearnerHandler(config.LearnerService)
	contentHandler := httphandlers.NewContentHandler(config.ContentService)
	kgHandler := httphandlers.NewKnowledgeGraphHandler(config.KnowledgeGraphService)
	progressHandler := httpinterfaces.NewProgressHandler(config.ProgressTrackingService)
	adaptiveHandler := httphandlers.NewAdaptiveLearningHandler(config.AdaptiveLearningService)
	knowledgeAnalysisHandler := handlers.NewKnowledgeAnalysisHandler(config.KnowledgeAnalysisService)
	progressWebSocketHandler := websockethandlers.NewProgressWebSocketHandler(config.ProgressTrackingService)
	realtimeAnalyticsHandler := httphandlers.NewRealtimeAnalyticsHandler(config.RealtimeAnalyticsService)
	
	// 创建学习路径处理器（需要添加到RouterConfig中）
	// learningPathHandler := httphandlers.NewLearningPathHandler(config.LearningPathService)
	
	// 创建推荐处理�?
	recommendationHandler := httphandlers.NewRecommendationHandler(
		config.PersonalizationEngine,
		config.UserBehaviorTracker,
		config.PreferenceAnalyzer,
		config.ContextAnalyzer,
	)

	// 创建实时推荐处理�?
	realtimeRecommendationHandler := httphandlers.NewRealtimeRecommendationHandler(
		config.RealtimeRecommendationService,
	)

	// 创建推荐集成处理�?
	recommendationIntegrationHandler := httphandlers.NewRecommendationIntegrationHandler(
		config.RecommendationIntegrationService,
	)

	// 创建API路由�?
	api := r.Group("/api/v1")
	
	// 设置API路由
	setupAPIRoutes(api, learnerHandler, contentHandler, kgHandler, progressHandler, adaptiveHandler, knowledgeAnalysisHandler, progressWebSocketHandler, realtimeAnalyticsHandler, recommendationHandler, realtimeRecommendationHandler, recommendationIntegrationHandler)

	// 设置Swagger文档
	setupSwagger(r)

	// 健康检�?
	setupHealthCheck(r)

	return r
}

// setupMiddleware 设置中间�?
func setupMiddleware(r *gin.Engine) {
	// 恢复中间�?
	r.Use(gin.Recovery())

	// 请求ID中间�?
	r.Use(requestid.New())

	// 日志中间�?
	r.Use(middleware.Logger())

	// CORS中间�?
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 限流中间�?
	r.Use(middleware.RateLimit())

	// 认证中间件（可选）
	// r.Use(middleware.Auth())
}

// setupAPIRoutes 设置API路由
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

	// API v1 路由�?
	v1 := r.Group("/api/v1")
	{
		// 学习者路�?
		learners := v1.Group("/learners")
		{
			learners.POST("", learnerHandler.CreateLearner)
			learners.GET("/:id", learnerHandler.GetLearner)
			learners.PUT("/:id", learnerHandler.UpdateLearner)
			learners.DELETE("/:id", learnerHandler.DeleteLearner)
			
			// 学习目标
			learners.POST("/:id/goals", learnerHandler.AddLearningGoal)
			learners.PUT("/:id/goals/:goalId", learnerHandler.UpdateLearningGoal)
			
			// 技�?
			learners.PUT("/:id/skills", learnerHandler.UpdateSkill)
			
			// 学习活动
			learners.POST("/:id/activities", learnerHandler.RecordLearningActivity)
			learners.GET("/:id/history", learnerHandler.GetLearningHistory)
			
			// 分析和推�?
			learners.GET("/:id/analytics", learnerHandler.GetLearningAnalytics)
			learners.GET("/:id/recommendations", learnerHandler.GetPersonalizedRecommendations)
		}

		// 内容路由
		content := v1.Group("/content")
		{
			content.POST("", contentHandler.CreateContent)
			content.GET("", contentHandler.ListContent)
			content.GET("/:id", contentHandler.GetContent)
			content.PUT("/:id", contentHandler.UpdateContent)
			content.DELETE("/:id", contentHandler.DeleteContent)
			
			// 内容状态管�?
			content.POST("/:id/publish", contentHandler.PublishContent)
			content.POST("/:id/archive", contentHandler.ArchiveContent)
			
			// 搜索和推�?
			content.GET("/search", contentHandler.SearchContent)
			content.GET("/personalized", contentHandler.GetPersonalizedContent)
			
			// 交互和分�?
			content.POST("/:id/interactions", contentHandler.RecordContentInteraction)
			content.GET("/:id/analytics", contentHandler.GetContentAnalytics)
			
			// 按知识节点获取内�?
			content.GET("/knowledge-node/:nodeId", contentHandler.GetContentByKnowledgeNode)
		}

		// 知识图谱路由
		kg := v1.Group("/knowledge-graph")
		{
			// 节点管理
			nodes := kg.Group("/nodes")
			{
				nodes.POST("", kgHandler.CreateNode)
				nodes.GET("/search", kgHandler.SearchNodes)
				nodes.GET("/:id", kgHandler.GetNode)
				nodes.PUT("/:id", kgHandler.UpdateNode)
				nodes.DELETE("/:id", kgHandler.DeleteNode)
				nodes.GET("/:id/neighbors", kgHandler.GetNodeNeighbors)
			}

			// 关系管理
			relations := kg.Group("/relations")
			{
				relations.POST("", kgHandler.CreateRelation)
				relations.GET("/:id", kgHandler.GetRelation)
				relations.PUT("/:id", kgHandler.UpdateRelation)
				relations.DELETE("/:id", kgHandler.DeleteRelation)
			}

			// 图谱分析
			kg.GET("/shortest-path", kgHandler.FindShortestPath)
			kg.POST("/learning-path", kgHandler.GenerateLearningPath)
			kg.POST("/concept-map", kgHandler.GenerateConceptMap)
			kg.POST("/analyze", kgHandler.AnalyzeGraph)
			kg.GET("/statistics", kgHandler.GetGraphStatistics)
			kg.POST("/validate", kgHandler.ValidateGraph)
		}

		// 进度追踪路由
		progress := v1.Group("/progress")
		{
			// 进度更新
			progress.POST("/update", progressHandler.UpdateProgress)
			progress.POST("/batch-update", progressHandler.BatchUpdateProgress)
			
			// 进度查询
			progress.GET("/learner/:learnerId", progressHandler.GetProgressSummary)
			progress.GET("/content/:contentId/learner/:learnerId", progressHandler.GetContentProgress)
			
			// 学习报告
			progress.GET("/report/:learnerId", progressHandler.GetLearningReport)
			
			// WebSocket连接
			progress.GET("/ws/:learnerId", progressWebSocketHandler.HandleWebSocket)
		}

		// 自适应学习路由
		adaptive := v1.Group("/adaptive")
		{
			// 路径适配
			adaptive.POST("/adapt-path", adaptiveHandler.AdaptLearningPath)
			
			// 适配推荐
			adaptive.GET("/recommendations/:learner_id", adaptiveHandler.GetAdaptationRecommendations)
			
			// 适配历史
			adaptive.GET("/history/:learner_id", adaptiveHandler.GetLearnerAdaptationHistory)
			
			// 效果分析
			adaptive.POST("/analyze-effectiveness", adaptiveHandler.AnalyzeLearningEffectiveness)
			
			// 结果预测
			adaptive.POST("/predict-outcome", adaptiveHandler.PredictLearningOutcome)
		}

		// 知识分析路由
		analysis := v1.Group("/knowledge-analysis")
		{
			// 概念关系分析
			analysis.POST("/concept-relationships", knowledgeAnalysisHandler.AnalyzeConceptRelationships)
			
			// 依赖图构�?
			analysis.POST("/dependency-graph", knowledgeAnalysisHandler.BuildDependencyGraph)
			
			// 内容推荐
			analysis.POST("/recommend-content", knowledgeAnalysisHandler.RecommendContent)
			
			// 概念聚类
			analysis.GET("/concept-clusters", knowledgeAnalysisHandler.GetConceptClusters)
			
			// 学习路径
			analysis.GET("/learning-path", knowledgeAnalysisHandler.GetLearningPath)
			
			// 个性化推荐
			analysis.GET("/personalized-recommendations", knowledgeAnalysisHandler.GetPersonalizedRecommendations)
			
			// 知识差距分析
			analysis.POST("/knowledge-gaps", knowledgeAnalysisHandler.AnalyzeKnowledgeGaps)
		}

		// 实时分析路由
		realtime := v1.Group("/realtime-analytics")
		{
			// 事件处理
			realtime.POST("/events", realtimeAnalyticsHandler.ProcessEvent)
			
			// 实时数据获取
			realtime.GET("/:learnerId/data", realtimeAnalyticsHandler.GetRealtimeData)
			realtime.GET("/:learnerId/metrics", realtimeAnalyticsHandler.GetAnalyticsMetrics)
			realtime.GET("/:learnerId/insights", realtimeAnalyticsHandler.GetLearningInsights)
			realtime.GET("/:learnerId/session", realtimeAnalyticsHandler.GetSessionAnalytics)
			realtime.GET("/:learnerId/performance", realtimeAnalyticsHandler.GetPerformanceTrends)
			realtime.GET("/:learnerId/alerts", realtimeAnalyticsHandler.GetAlerts)
			realtime.GET("/:learnerId/recommendations", realtimeAnalyticsHandler.GetRecommendations)
			
			// 分析器管�?
			realtime.POST("/analyzers", realtimeAnalyticsHandler.CreateAnalyzer)
			
			// WebSocket订阅
			realtime.GET("/subscribe", realtimeAnalyticsHandler.SubscribeToUpdates)
		}

		// 推荐系统路由
		recommendations := v1.Group("/recommendations")
		{
			// 个性化推荐
			recommendations.POST("/personalized", recommendationHandler.GetPersonalizedRecommendations)
			
			// 策略推荐
			recommendations.GET("/strategy/:strategy", recommendationHandler.GetRecommendationsByStrategy)
			
			// 批量推荐
			recommendations.POST("/batch", recommendationHandler.BatchRecommendations)
			
			// 用户行为记录
			recommendations.POST("/behavior", recommendationHandler.RecordUserBehavior)
			
			// 用户偏好分析
			recommendations.GET("/preferences/:user_id", recommendationHandler.GetUserPreferences)
			
			// 学习上下文分�?
			recommendations.GET("/context/:user_id", recommendationHandler.GetLearningContext)
			
			// 行为洞察
		recommendations.GET("/insights/:user_id", recommendationHandler.GetBehaviorInsights)
	}

	// 实时推荐路由
	realtimeRec := v1.Group("/realtime-recommendations")
	{
		// 实时事件处理
		realtimeRec.POST("/events", realtimeRecommendationHandler.ProcessRealtimeEvent)
		realtimeRec.POST("/events/batch", realtimeRecommendationHandler.BatchProcessEvents)
		
		// 实时推荐获取
		realtimeRec.GET("/:user_id", realtimeRecommendationHandler.GetRealtimeRecommendations)
		
		// 用户会话管理
		realtimeRec.GET("/sessions/:user_id", realtimeRecommendationHandler.GetUserSession)
		
		// WebSocket订阅
		realtimeRec.GET("/subscribe", realtimeRecommendationHandler.SubscribeToRecommendationUpdates)
		
		// 推荐指标
		realtimeRec.GET("/metrics", realtimeRecommendationHandler.GetRecommendationMetrics)
	}

	// 推荐集成路由
	integratedRec := v1.Group("/integrated-recommendations")
	{
		// 获取集成推荐
		integratedRec.GET("/:user_id", recommendationIntegrationHandler.GetIntegratedRecommendations)
		
		// 批量获取推荐
		integratedRec.POST("/batch", recommendationIntegrationHandler.BatchGetRecommendations)
		
		// 推荐指标
		integratedRec.GET("/metrics", recommendationIntegrationHandler.GetRecommendationMetrics)
		
		// 缓存管理
		integratedRec.DELETE("/cache", recommendationIntegrationHandler.ClearRecommendationCache)
	}
	}
}

// setupSwagger 设置Swagger文档
func setupSwagger(r *gin.Engine) {
	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// 重定向根路径到Swagger文档
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

// setupHealthCheck 设置健康检�?
func setupHealthCheck(r *gin.Engine) {
	// 健康检查端�?
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "intelligent-learning",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 就绪检查端�?
	r.GET("/ready", func(c *gin.Context) {
		// 这里可以添加数据库连接检查等
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// 存活检查端�?
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})
}
