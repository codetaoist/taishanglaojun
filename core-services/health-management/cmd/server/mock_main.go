package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HealthData 
type HealthData struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	DataType   string    `json:"data_type"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Source     string    `json:"source"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// HealthProfile 
type HealthProfile struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	Name            string                 `json:"name"`
	Age             int                    `json:"age"`
	Gender          string                 `json:"gender"`
	Height          float64                `json:"height"`
	Weight          float64                `json:"weight"`
	BloodType       string                 `json:"blood_type"`
	MedicalHistory  []string               `json:"medical_history"`
	Allergies       []string               `json:"allergies"`
	HealthGoals     map[string]interface{} `json:"health_goals"`
	EmergencyContact map[string]string     `json:"emergency_contact"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// HealthAnalysis 
type HealthAnalysis struct {
	UserID      string                 `json:"user_id"`
	AnalysisType string                `json:"analysis_type"`
	Period      string                 `json:"period"`
	Summary     map[string]interface{} `json:"summary"`
	Trends      []map[string]interface{} `json:"trends"`
	Insights    []string               `json:"insights"`
	Recommendations []string           `json:"recommendations"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// 洢
var (
	healthDataStore    = make(map[string]HealthData)
	healthProfileStore = make(map[string]HealthProfile)
)

func main() {
	// ?
	initMockData()

	// Gin
	r := gin.Default()

	// CORS?
	r.Use(func(c *gin.Context) {
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
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "health-management",
			"timestamp": time.Now(),
		})
	})

	// API
	api := r.Group("/api/v1")
	{
		// 
		api.POST("/health-data", createHealthData)
		api.GET("/health-data", listHealthData)
		api.GET("/health-data/:id", getHealthData)
		api.PUT("/health-data/:id", updateHealthData)
		api.DELETE("/health-data/:id", deleteHealthData)

		// 
		api.POST("/health-profiles", createHealthProfile)
		api.GET("/health-profiles", listHealthProfiles)
		api.GET("/health-profiles/:id", getHealthProfile)
		api.PUT("/health-profiles/:id", updateHealthProfile)
		api.DELETE("/health-profiles/:id", deleteHealthProfile)

		// 
		api.GET("/health-analysis/:user_id", getHealthAnalysis)
		api.GET("/health-analysis/:user_id/trends", getHealthTrends)
		api.GET("/health-analysis/:user_id/insights", getHealthInsights)

		// 
		api.GET("/health-recommendations/:user_id", getHealthRecommendations)

		// ?
		api.GET("/health-dashboard/:user_id", getHealthDashboard)

		// 
		api.GET("/health-alerts/:user_id", getHealthAlerts)
	}

	log.Println(" (Mock汾) ?8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

// ?
func initMockData() {
	userID := "user-123"
	now := time.Now()

	// 
	profile := HealthProfile{
		ID:       uuid.New().String(),
		UserID:   userID,
		Name:     "",
		Age:      30,
		Gender:   "male",
		Height:   175.0,
		Weight:   70.0,
		BloodType: "A+",
		MedicalHistory: []string{"?, "?},
		Allergies:      []string{"", ""},
		HealthGoals: map[string]interface{}{
			"target_weight": 65.0,
			"daily_steps":   10000,
			"sleep_hours":   8,
		},
		EmergencyContact: map[string]string{
			"name":  "",
			"phone": "13800138000",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	healthProfileStore[profile.ID] = profile

	// 
	healthDataTypes := []struct {
		dataType string
		value    float64
		unit     string
	}{
		{"heart_rate", 72, "bpm"},
		{"blood_pressure", 120, "mmHg"},
		{"weight", 70.5, "kg"},
		{"steps", 8500, "steps"},
		{"sleep_duration", 7.5, "hours"},
	}

	for _, data := range healthDataTypes {
		healthData := HealthData{
			ID:         uuid.New().String(),
			UserID:     userID,
			DataType:   data.dataType,
			Value:      data.value,
			Unit:       data.unit,
			Source:     "smart_watch",
			RecordedAt: now,
			CreatedAt:  now,
		}
		healthDataStore[healthData.ID] = healthData
	}
}

// ?
func createHealthData(c *gin.Context) {
	var req HealthData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.RecordedAt = time.Now()

	healthDataStore[req.ID] = req
	c.JSON(201, req)
}

func listHealthData(c *gin.Context) {
	userID := c.Query("user_id")
	dataType := c.Query("data_type")

	var result []HealthData
	for _, data := range healthDataStore {
		if userID != "" && data.UserID != userID {
			continue
		}
		if dataType != "" && data.DataType != dataType {
			continue
		}
		result = append(result, data)
	}

	c.JSON(200, gin.H{
		"data":  result,
		"total": len(result),
	})
}

func getHealthData(c *gin.Context) {
	id := c.Param("id")
	data, exists := healthDataStore[id]
	if !exists {
		c.JSON(404, gin.H{"error": "?})
		return
	}
	c.JSON(200, data)
}

func updateHealthData(c *gin.Context) {
	id := c.Param("id")
	data, exists := healthDataStore[id]
	if !exists {
		c.JSON(404, gin.H{"error": "?})
		return
	}

	var req HealthData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.ID = data.ID
	req.CreatedAt = data.CreatedAt
	healthDataStore[id] = req
	c.JSON(200, req)
}

func deleteHealthData(c *gin.Context) {
	id := c.Param("id")
	if _, exists := healthDataStore[id]; !exists {
		c.JSON(404, gin.H{"error": "?})
		return
	}

	delete(healthDataStore, id)
	c.JSON(200, gin.H{"message": ""})
}

// ?
func createHealthProfile(c *gin.Context) {
	var req HealthProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	healthProfileStore[req.ID] = req
	c.JSON(201, req)
}

func listHealthProfiles(c *gin.Context) {
	userID := c.Query("user_id")

	var result []HealthProfile
	for _, profile := range healthProfileStore {
		if userID != "" && profile.UserID != userID {
			continue
		}
		result = append(result, profile)
	}

	c.JSON(200, gin.H{
		"data":  result,
		"total": len(result),
	})
}

func getHealthProfile(c *gin.Context) {
	id := c.Param("id")
	profile, exists := healthProfileStore[id]
	if !exists {
		c.JSON(404, gin.H{"error": "?})
		return
	}
	c.JSON(200, profile)
}

func updateHealthProfile(c *gin.Context) {
	id := c.Param("id")
	profile, exists := healthProfileStore[id]
	if !exists {
		c.JSON(404, gin.H{"error": "?})
		return
	}

	var req HealthProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.ID = profile.ID
	req.CreatedAt = profile.CreatedAt
	req.UpdatedAt = time.Now()
	healthProfileStore[id] = req
	c.JSON(200, req)
}

func deleteHealthProfile(c *gin.Context) {
	id := c.Param("id")
	if _, exists := healthProfileStore[id]; !exists {
		c.JSON(404, gin.H{"error": "?})
		return
	}

	delete(healthProfileStore, id)
	c.JSON(200, gin.H{"message": ""})
}

// ?
func getHealthAnalysis(c *gin.Context) {
	userID := c.Param("user_id")
	analysisType := c.DefaultQuery("type", "comprehensive")
	period := c.DefaultQuery("period", "week")

	analysis := HealthAnalysis{
		UserID:       userID,
		AnalysisType: analysisType,
		Period:       period,
		Summary: map[string]interface{}{
			"avg_heart_rate":    72,
			"avg_blood_pressure": 120,
			"total_steps":       59500,
			"avg_sleep_hours":   7.5,
			"weight_change":     -0.5,
		},
		Trends: []map[string]interface{}{
			{
				"metric": "heart_rate",
				"trend":  "stable",
				"change": 0.02,
			},
			{
				"metric": "weight",
				"trend":  "decreasing",
				"change": -0.5,
			},
		},
		Insights: []string{
			"",
			"",
			"鱣?,
		},
		Recommendations: []string{
			"8000?,
			"?,
			"?,
		},
		GeneratedAt: time.Now(),
	}

	c.JSON(200, analysis)
}

func getHealthTrends(c *gin.Context) {
	userID := c.Param("user_id")
	metric := c.DefaultQuery("metric", "heart_rate")
	period := c.DefaultQuery("period", "week")

	// 
	trends := []map[string]interface{}{
		{"date": "2024-01-01", "value": 70},
		{"date": "2024-01-02", "value": 72},
		{"date": "2024-01-03", "value": 71},
		{"date": "2024-01-04", "value": 73},
		{"date": "2024-01-05", "value": 72},
		{"date": "2024-01-06", "value": 74},
		{"date": "2024-01-07", "value": 72},
	}

	c.JSON(200, gin.H{
		"user_id": userID,
		"metric":  metric,
		"period":  period,
		"trends":  trends,
	})
}

func getHealthInsights(c *gin.Context) {
	userID := c.Param("user_id")

	insights := []map[string]interface{}{
		{
			"type":        "positive",
			"title":       "",
			"description": "9000?,
			"priority":    "medium",
		},
		{
			"type":        "warning",
			"title":       "",
			"description": "䲻??,
			"priority":    "high",
		},
		{
			"type":        "info",
			"title":       "",
			"description": "?,
			"priority":    "low",
		},
	}

	c.JSON(200, gin.H{
		"user_id":  userID,
		"insights": insights,
	})
}

func getHealthRecommendations(c *gin.Context) {
	userID := c.Param("user_id")

	recommendations := []map[string]interface{}{
		{
			"category":    "exercise",
			"title":       "",
			"description": "3-430",
			"priority":    "high",
		},
		{
			"category":    "diet",
			"title":       "",
			"description": "?,
			"priority":    "medium",
		},
		{
			"category":    "sleep",
			"title":       "",
			"description": "豸",
			"priority":    "high",
		},
	}

	c.JSON(200, gin.H{
		"user_id":         userID,
		"recommendations": recommendations,
	})
}

func getHealthDashboard(c *gin.Context) {
	userID := c.Param("user_id")

	dashboard := map[string]interface{}{
		"user_id": userID,
		"overview": map[string]interface{}{
			"health_score":      85,
			"risk_level":        "low",
			"last_checkup":      "2024-01-01",
			"next_checkup":      "2024-07-01",
		},
		"vital_signs": map[string]interface{}{
			"heart_rate":      72,
			"blood_pressure":  "120/80",
			"body_temperature": 36.5,
			"weight":          70.5,
		},
		"activity": map[string]interface{}{
			"daily_steps":     8500,
			"calories_burned": 2200,
			"active_minutes":  45,
			"sleep_hours":     7.5,
		},
		"alerts": []map[string]interface{}{
			{
				"type":    "warning",
				"message": "",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			},
		},
	}

	c.JSON(200, dashboard)
}

func getHealthAlerts(c *gin.Context) {
	userID := c.Param("user_id")

	alerts := []map[string]interface{}{
		{
			"id":          uuid.New().String(),
			"type":        "warning",
			"title":       "?,
			"description": "140/90?,
			"severity":    "high",
			"created_at":  time.Now().Add(-2 * time.Hour),
			"read":        false,
		},
		{
			"id":          uuid.New().String(),
			"type":        "info",
			"title":       "",
			"description": "?0000",
			"severity":    "low",
			"created_at":  time.Now().Add(-4 * time.Hour),
			"read":        true,
		},
	}

	c.JSON(200, gin.H{
		"user_id": userID,
		"alerts":  alerts,
	})
}

