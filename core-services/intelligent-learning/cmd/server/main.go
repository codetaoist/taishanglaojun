package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/adaptive"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/content"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/knowledge"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/learner"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/recommendation"
	domainservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/infrastructure/config"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/infrastructure/external"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/infrastructure/persistence"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/http/routes"
)

// @title 智能学习系统 API
// @version 1.0
// @description 个性化学习路径和知识图谱系统的RESTful API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 设置日志格式
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库管理
	dbManager, err := persistence.NewDatabaseManager(&cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize database manager: %v", err)
	}
	defer dbManager.Close()

	// 健康检查数据库连接
	healthResults := dbManager.Health(context.Background())
	for name, err := range healthResults {
		if err != nil {
			log.Printf("Database %s health check failed: %v", name, err)
		} else {
			log.Printf("Database %s is healthy", name)
		}
	}

	// 初始化仓储层
	learnerRepo := persistence.NewLearnerRepository(dbManager.GetPostgreSQL())
	contentRepo := persistence.NewLearningContentRepository(dbManager.GetPostgreSQL(), dbManager.GetElasticsearch())
	knowledgeGraphRepo := persistence.NewKnowledgeGraphRepository(dbManager.GetNeo4j())
	recommendationRepo := persistence.NewRecommendationRepository(dbManager.GetPostgreSQL())

	// 初始化领域服?
	kgDomainService := domainservices.NewKnowledgeGraphService(knowledgeGraphRepo, learnerRepo, contentRepo)
	learningPathService := domainservices.NewLearningPathService(
		knowledgeGraphRepo,
		learnerRepo,
		contentRepo, // 添加缺少的contentRepo参数
	)
	analyticsService := domainservices.NewLearningAnalyticsService(
		learnerRepo,
		contentRepo,
		knowledgeGraphRepo,
	)

	// 创建学习路径服务适配?
	learningPathServiceAdapter := adaptive.NewLearningPathServiceAdapter(learningPathService)

	adaptiveLearningService := adaptive.NewAdaptiveLearningService(
		learnerRepo,
		contentRepo,
		knowledgeGraphRepo,
		analyticsService,           // 正确的参数类?
		learningPathServiceAdapter, // 使用适配?
	)

	// 初始化应用服?
	learnerService := learner.NewLearnerService(
		learnerRepo,
		knowledgeGraphRepo,
		contentRepo,
		learningPathServiceAdapter,
		analyticsService,
		kgDomainService,
	)

	contentService := content.NewContentService(
		contentRepo,
		learnerRepo,
		knowledgeGraphRepo,
		analyticsService,
	)

	// 使用默认的知识图谱ID
	defaultGraphID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	kgAppService := knowledge.NewKnowledgeGraphAppService(
		knowledgeGraphRepo,
		contentRepo,
		learnerRepo,
		kgDomainService,
		learningPathService,
		defaultGraphID,
	)

	// 初始化进度追踪服?
	progressService := analytics.NewProgressTrackingService(
		learnerRepo,
		contentRepo,
		knowledgeGraphRepo,
		analyticsService,
	)

	// 初始化知识分析服?
	knowledgeAnalysisService := knowledge.NewKnowledgeAnalysisService(
		knowledgeGraphRepo,
		contentRepo,
		learnerRepo,
		kgDomainService,
		analyticsService,
	)

	// 初始化外部服?
	locationService := external.NewMockLocationService()
	weatherService := external.NewMockWeatherService()

	// 创建适配?
	behaviorRepoAdapter := recommendation.NewBehaviorRepositoryAdapter(recommendationRepo)
	eventStoreAdapter := recommendation.NewEventStoreAdapter(analyticsService)
	environmentRepoAdapter := recommendation.NewEnvironmentRepositoryAdapter(recommendationRepo)
	locationServiceAdapter := recommendation.NewLocationServiceAdapter(locationService)
	weatherServiceAdapter := recommendation.NewWeatherServiceAdapter(weatherService)
	contentRepoAdapter := recommendation.NewContentRepositoryAdapter(contentRepo)
	userRepoAdapter := recommendation.NewUserRepositoryAdapter(learnerRepo)

	// 初始化用户行为跟踪器
	userBehaviorTracker := domainservices.NewUserBehaviorTracker(
		behaviorRepoAdapter,
		eventStoreAdapter,
	)

	preferenceAnalyzer := domainservices.NewPreferenceAnalyzer(
		userBehaviorTracker,
		contentRepoAdapter,
		userRepoAdapter,
	)

	contextAnalyzer := domainservices.NewContextAnalyzer(
		userBehaviorTracker,
		preferenceAnalyzer,
		environmentRepoAdapter,
		locationServiceAdapter,
		weatherServiceAdapter,
	)

	// 创建一个简单的推荐模型实现
	recommendationModel := &recommendation.SimpleRecommendationModel{}

	// 初始化个性化引擎
	personalizationEngine := domainservices.NewPersonalizationEngine(
		learnerRepo,
		contentRepo,
		knowledgeGraphRepo,
		userBehaviorTracker,
		preferenceAnalyzer,
		contextAnalyzer,
		recommendationModel,
	)

	// 初始化实时推荐服?
	realtimeRecommendationService := recommendation.NewRealtimeRecommendationService(
		personalizationEngine,
		userBehaviorTracker,
		preferenceAnalyzer,
		contextAnalyzer,
	)

	// 初始化推荐集成服?
	recommendationIntegrationService := recommendation.NewRecommendationIntegrationService(
		personalizationEngine,
		userBehaviorTracker,
		preferenceAnalyzer,
		contextAnalyzer,
		realtimeRecommendationService,
		nil, // 使用默认配置
	)

	// 设置路由
	router := routes.SetupRoutes(&routes.RouterConfig{
		LearnerService:                   learnerService,
		ContentService:                   contentService,
		KnowledgeGraphService:            kgAppService,
		ProgressTrackingService:          progressService,
		AdaptiveLearningService:          adaptiveLearningService,
		KnowledgeAnalysisService:         knowledgeAnalysisService,
		PersonalizationEngine:            personalizationEngine,
		UserBehaviorTracker:              userBehaviorTracker,
		PreferenceAnalyzer:               preferenceAnalyzer,
		ContextAnalyzer:                  contextAnalyzer,
		RealtimeRecommendationService:    realtimeRecommendationService,
		RecommendationIntegrationService: recommendationIntegrationService,
	})

	// 创建HTTP服务
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务
	go func() {
		logrus.WithFields(logrus.Fields{
			"port":    cfg.Server.Port,
			"env":     cfg.App.Environment,
			"version": cfg.App.Version,
		}).Info("Starting intelligent learning system server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

