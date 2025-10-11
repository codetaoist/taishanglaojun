package http

import (
	"github.com/gin-gonic/gin"
	
	"github.com/taishanglaojun/health-management/internal/application"
)

// Router HTTP路由�?
type Router struct {
	healthDataHandler         *HealthDataHandler
	healthProfileHandler      *HealthProfileHandler
	healthAnalysisHandler     *HealthAnalysisHandler
	healthAlertHandler        *HealthAlertHandler
	healthRecommendationHandler *HealthRecommendationHandler
	healthDashboardHandler    *HealthDashboardHandler
}

// NewRouter 创建新的路由�?
func NewRouter(
	healthDataHandler *HealthDataHandler,
	healthProfileHandler *HealthProfileHandler,
	healthAnalysisHandler *HealthAnalysisHandler,
	healthAlertHandler *HealthAlertHandler,
	healthRecommendationHandler *HealthRecommendationHandler,
	healthDashboardHandler *HealthDashboardHandler,
) *Router {
	return &Router{
		healthDataHandler:         healthDataHandler,
		healthProfileHandler:      healthProfileHandler,
		healthAnalysisHandler:     healthAnalysisHandler,
		healthAlertHandler:        healthAlertHandler,
		healthRecommendationHandler: healthRecommendationHandler,
		healthDashboardHandler:    healthDashboardHandler,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// API版本分组
	v1 := engine.Group("/api/v1")
	
	// 健康数据路由
	r.setupHealthDataRoutes(v1)
	
	// 健康档案路由
	r.setupHealthProfileRoutes(v1)
	
	// 健康分析路由
	r.setupHealthAnalysisRoutes(v1)
	
	// 健康预警路由
	r.setupHealthAlertRoutes(v1)
	
	// 健康建议路由
	r.setupHealthRecommendationRoutes(v1)
	
	// 健康仪表板路�?
	r.setupHealthDashboardRoutes(v1)
	
	// 用户相关路由
	r.setupUserRoutes(v1)
}

// setupHealthDataRoutes 设置健康数据路由
func (r *Router) setupHealthDataRoutes(rg *gin.RouterGroup) {
	healthData := rg.Group("/health-data")
	{
		healthData.POST("", r.healthDataHandler.CreateHealthData)
		healthData.GET("/:id", r.healthDataHandler.GetHealthData)
		healthData.PUT("/:id", r.healthDataHandler.UpdateHealthData)
		healthData.DELETE("/:id", r.healthDataHandler.DeleteHealthData)
	}
}

// setupHealthProfileRoutes 设置健康档案路由
func (r *Router) setupHealthProfileRoutes(rg *gin.RouterGroup) {
	healthProfiles := rg.Group("/health-profiles")
	{
		healthProfiles.POST("", r.healthProfileHandler.CreateHealthProfile)
		healthProfiles.GET("", r.healthProfileHandler.ListHealthProfiles)
		healthProfiles.GET("/:id", r.healthProfileHandler.GetHealthProfile)
		healthProfiles.PUT("/:id", r.healthProfileHandler.UpdateHealthProfile)
		healthProfiles.DELETE("/:id", r.healthProfileHandler.DeleteHealthProfile)
		
		// 病史管理
		healthProfiles.POST("/:id/medical-history", r.healthProfileHandler.AddMedicalHistory)
		healthProfiles.DELETE("/:id/medical-history/:condition", r.healthProfileHandler.RemoveMedicalHistory)
		
		// 过敏史管�?
		healthProfiles.POST("/:id/allergies", r.healthProfileHandler.AddAllergy)
		healthProfiles.DELETE("/:id/allergies/:allergen", r.healthProfileHandler.RemoveAllergy)
		
		// 健康目标管理
		healthProfiles.PUT("/:id/health-goals", r.healthProfileHandler.SetHealthGoals)
		
		// BMI计算
		healthProfiles.GET("/:id/bmi", r.healthProfileHandler.CalculateBMI)
	}
}

// setupHealthAnalysisRoutes 设置健康分析路由
func (r *Router) setupHealthAnalysisRoutes(rg *gin.RouterGroup) {
	healthAnalysis := rg.Group("/health-analysis")
	{
		healthAnalysis.POST("/trend", r.healthAnalysisHandler.AnalyzeHealthTrend)
		healthAnalysis.POST("/risk-assessment", r.healthAnalysisHandler.AssessHealthRisk)
		healthAnalysis.POST("/insights", r.healthAnalysisHandler.GenerateHealthInsights)
	}
}

// setupHealthAlertRoutes 设置健康预警路由
func (r *Router) setupHealthAlertRoutes(rg *gin.RouterGroup) {
	healthAlert := rg.Group("/health-alerts")
	{
		healthAlert.POST("/detect-anomalies", r.healthAlertHandler.DetectAnomalies)
		healthAlert.POST("/check-emergency", r.healthAlertHandler.CheckEmergency)
		healthAlert.POST("/alerts", r.healthAlertHandler.GetAlerts)
		healthAlert.PUT("/alerts/mark", r.healthAlertHandler.MarkAlert)
	}
}

// setupHealthRecommendationRoutes 设置健康建议路由
func (r *Router) setupHealthRecommendationRoutes(v1 *gin.RouterGroup) {
	recommendations := v1.Group("/health-recommendations")
	{
		recommendations.POST("/generate", r.healthRecommendationHandler.GenerateRecommendations)
		recommendations.GET("/tips", r.healthRecommendationHandler.GetPersonalizedTips)
		recommendations.GET("/types", r.healthRecommendationHandler.GetRecommendationTypes)
	}
}

// setupHealthDashboardRoutes 设置健康仪表板路�?
func (r *Router) setupHealthDashboardRoutes(v1 *gin.RouterGroup) {
	dashboard := v1.Group("/health-dashboard")
	{
		dashboard.GET("", r.healthDashboardHandler.GetDashboard)
		dashboard.GET("/summary", r.healthDashboardHandler.GetDashboardSummary)
		dashboard.GET("/metrics", r.healthDashboardHandler.GetDashboardMetrics)
		dashboard.GET("/charts", r.healthDashboardHandler.GetDashboardCharts)
	}
}

// setupUserRoutes 设置用户相关路由
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// 用户健康数据
		users.GET("/:user_id/health-data", r.healthDataHandler.GetHealthDataByUser)
		users.GET("/:user_id/health-data/latest/:data_type", r.healthDataHandler.GetLatestHealthData)
		users.GET("/:user_id/health-data/statistics/:data_type", r.healthDataHandler.GetHealthDataStatistics)
		users.GET("/:user_id/health-data/abnormal", r.healthDataHandler.GetAbnormalHealthData)
		
		// 用户健康档案
		users.GET("/:user_id/health-profile", r.healthProfileHandler.GetHealthProfileByUser)
		
		// 用户健康分析
		users.GET("/:user_id/health-analysis/trend", r.healthAnalysisHandler.GetHealthTrendByUser)
		users.GET("/:user_id/health-analysis/risk-assessment", r.healthAnalysisHandler.GetHealthRiskByUser)
		users.GET("/:user_id/health-analysis/insights", r.healthAnalysisHandler.GetHealthInsightsByUser)
		
		// 用户健康预警
		users.GET("/:user_id/health-alerts", r.healthAlertHandler.GetAlertsByUser)
		users.GET("/:user_id/health-alerts/statistics", r.healthAlertHandler.GetAlertStatistics)

		// 健康建议
		users.POST("/:user_id/health-recommendations", r.healthRecommendationHandler.GetRecommendationsByUser)
		users.GET("/:user_id/health-tips", r.healthRecommendationHandler.GetTipsByUser)

		// 健康仪表�?
		users.GET("/:user_id/health-dashboard", r.healthDashboardHandler.GetDashboardByUser)
		users.GET("/:user_id/health-dashboard/summary", r.healthDashboardHandler.GetDashboardSummaryByUser)
	}
}

// SetupMiddlewares 设置中间�?
func SetupMiddlewares(engine *gin.Engine) {
	// CORS中间�?
	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// 请求日志中间�?
	engine.Use(gin.Logger())
	
	// 恢复中间�?
	engine.Use(gin.Recovery())
	
	// 请求ID中间�?
	engine.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	// 这里简化处理，实际应该使用更好的ID生成算法
	return "req-" + randomString(8)
}

// randomString 生成随机字符�?
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)/2] // 简化处�?
	}
	return string(b)
}

// HealthCheckHandler 健康检查处理器
func HealthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "health-management",
		"version": "1.0.0",
	})
}

// SetupHealthCheck 设置健康检查路�?
func SetupHealthCheck(engine *gin.Engine, router *Router) {
	engine.GET("/health", HealthCheckHandler)
	engine.GET("/health/ready", HealthCheckHandler)
	engine.GET("/health/live", HealthCheckHandler)
	engine.GET("/health/analysis", router.healthAnalysisHandler.HealthAnalysisHealthCheckHandler)
	engine.GET("/health/alerts", router.healthAlertHandler.HealthAlertHealthCheckHandler)
	engine.GET("/health/recommendations", router.healthRecommendationHandler.HealthRecommendationHealthCheckHandler)
	engine.GET("/health/dashboard", router.healthDashboardHandler.HealthDashboardHealthCheckHandler)
	engine.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
}