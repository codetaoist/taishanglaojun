package http

import (
	"github.com/gin-gonic/gin"
	
	"github.com/taishanglaojun/health-management/internal/application"
)

// Router HTTP?
type Router struct {
	healthDataHandler         *HealthDataHandler
	healthProfileHandler      *HealthProfileHandler
	healthAnalysisHandler     *HealthAnalysisHandler
	healthAlertHandler        *HealthAlertHandler
	healthRecommendationHandler *HealthRecommendationHandler
	healthDashboardHandler    *HealthDashboardHandler
}

// NewRouter ?
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

// SetupRoutes 
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// API汾
	v1 := engine.Group("/api/v1")
	
	// 
	r.setupHealthDataRoutes(v1)
	
	// 
	r.setupHealthProfileRoutes(v1)
	
	// 
	r.setupHealthAnalysisRoutes(v1)
	
	// 
	r.setupHealthAlertRoutes(v1)
	
	// 
	r.setupHealthRecommendationRoutes(v1)
	
	// ?
	r.setupHealthDashboardRoutes(v1)
	
	// 
	r.setupUserRoutes(v1)
}

// setupHealthDataRoutes 
func (r *Router) setupHealthDataRoutes(rg *gin.RouterGroup) {
	healthData := rg.Group("/health-data")
	{
		healthData.POST("", r.healthDataHandler.CreateHealthData)
		healthData.GET("/:id", r.healthDataHandler.GetHealthData)
		healthData.PUT("/:id", r.healthDataHandler.UpdateHealthData)
		healthData.DELETE("/:id", r.healthDataHandler.DeleteHealthData)
	}
}

// setupHealthProfileRoutes 
func (r *Router) setupHealthProfileRoutes(rg *gin.RouterGroup) {
	healthProfiles := rg.Group("/health-profiles")
	{
		healthProfiles.POST("", r.healthProfileHandler.CreateHealthProfile)
		healthProfiles.GET("", r.healthProfileHandler.ListHealthProfiles)
		healthProfiles.GET("/:id", r.healthProfileHandler.GetHealthProfile)
		healthProfiles.PUT("/:id", r.healthProfileHandler.UpdateHealthProfile)
		healthProfiles.DELETE("/:id", r.healthProfileHandler.DeleteHealthProfile)
		
		// 
		healthProfiles.POST("/:id/medical-history", r.healthProfileHandler.AddMedicalHistory)
		healthProfiles.DELETE("/:id/medical-history/:condition", r.healthProfileHandler.RemoveMedicalHistory)
		
		// ?
		healthProfiles.POST("/:id/allergies", r.healthProfileHandler.AddAllergy)
		healthProfiles.DELETE("/:id/allergies/:allergen", r.healthProfileHandler.RemoveAllergy)
		
		// 
		healthProfiles.PUT("/:id/health-goals", r.healthProfileHandler.SetHealthGoals)
		
		// BMI
		healthProfiles.GET("/:id/bmi", r.healthProfileHandler.CalculateBMI)
	}
}

// setupHealthAnalysisRoutes 
func (r *Router) setupHealthAnalysisRoutes(rg *gin.RouterGroup) {
	healthAnalysis := rg.Group("/health-analysis")
	{
		healthAnalysis.POST("/trend", r.healthAnalysisHandler.AnalyzeHealthTrend)
		healthAnalysis.POST("/risk-assessment", r.healthAnalysisHandler.AssessHealthRisk)
		healthAnalysis.POST("/insights", r.healthAnalysisHandler.GenerateHealthInsights)
	}
}

// setupHealthAlertRoutes 
func (r *Router) setupHealthAlertRoutes(rg *gin.RouterGroup) {
	healthAlert := rg.Group("/health-alerts")
	{
		healthAlert.POST("/detect-anomalies", r.healthAlertHandler.DetectAnomalies)
		healthAlert.POST("/check-emergency", r.healthAlertHandler.CheckEmergency)
		healthAlert.POST("/alerts", r.healthAlertHandler.GetAlerts)
		healthAlert.PUT("/alerts/mark", r.healthAlertHandler.MarkAlert)
	}
}

// setupHealthRecommendationRoutes 
func (r *Router) setupHealthRecommendationRoutes(v1 *gin.RouterGroup) {
	recommendations := v1.Group("/health-recommendations")
	{
		recommendations.POST("/generate", r.healthRecommendationHandler.GenerateRecommendations)
		recommendations.GET("/tips", r.healthRecommendationHandler.GetPersonalizedTips)
		recommendations.GET("/types", r.healthRecommendationHandler.GetRecommendationTypes)
	}
}

// setupHealthDashboardRoutes ?
func (r *Router) setupHealthDashboardRoutes(v1 *gin.RouterGroup) {
	dashboard := v1.Group("/health-dashboard")
	{
		dashboard.GET("", r.healthDashboardHandler.GetDashboard)
		dashboard.GET("/summary", r.healthDashboardHandler.GetDashboardSummary)
		dashboard.GET("/metrics", r.healthDashboardHandler.GetDashboardMetrics)
		dashboard.GET("/charts", r.healthDashboardHandler.GetDashboardCharts)
	}
}

// setupUserRoutes 
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// 
		users.GET("/:user_id/health-data", r.healthDataHandler.GetHealthDataByUser)
		users.GET("/:user_id/health-data/latest/:data_type", r.healthDataHandler.GetLatestHealthData)
		users.GET("/:user_id/health-data/statistics/:data_type", r.healthDataHandler.GetHealthDataStatistics)
		users.GET("/:user_id/health-data/abnormal", r.healthDataHandler.GetAbnormalHealthData)
		
		// 
		users.GET("/:user_id/health-profile", r.healthProfileHandler.GetHealthProfileByUser)
		
		// 
		users.GET("/:user_id/health-analysis/trend", r.healthAnalysisHandler.GetHealthTrendByUser)
		users.GET("/:user_id/health-analysis/risk-assessment", r.healthAnalysisHandler.GetHealthRiskByUser)
		users.GET("/:user_id/health-analysis/insights", r.healthAnalysisHandler.GetHealthInsightsByUser)
		
		// 
		users.GET("/:user_id/health-alerts", r.healthAlertHandler.GetAlertsByUser)
		users.GET("/:user_id/health-alerts/statistics", r.healthAlertHandler.GetAlertStatistics)

		// 
		users.POST("/:user_id/health-recommendations", r.healthRecommendationHandler.GetRecommendationsByUser)
		users.GET("/:user_id/health-tips", r.healthRecommendationHandler.GetTipsByUser)

		// ?
		users.GET("/:user_id/health-dashboard", r.healthDashboardHandler.GetDashboardByUser)
		users.GET("/:user_id/health-dashboard/summary", r.healthDashboardHandler.GetDashboardSummaryByUser)
	}
}

// SetupMiddlewares ?
func SetupMiddlewares(engine *gin.Engine) {
	// CORS?
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
	
	// ?
	engine.Use(gin.Logger())
	
	// ?
	engine.Use(gin.Recovery())
	
	// ID?
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

// generateRequestID ID
func generateRequestID() string {
	// ID㷨
	return "req-" + randomString(8)
}

// randomString ?
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)/2] // ?
	}
	return string(b)
}

// HealthCheckHandler 鴦
func HealthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "health-management",
		"version": "1.0.0",
	})
}

// SetupHealthCheck ?
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

