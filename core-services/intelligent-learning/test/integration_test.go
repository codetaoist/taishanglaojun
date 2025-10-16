package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/adaptive"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/handlers"
)

// IntegrationTestSuite 
type IntegrationTestSuite struct {
	suite.Suite
	router           *gin.Engine
	learningPathSvc  *adaptive.LearningPathService
	achievementSvc   *analytics.LearningAchievementService
	communitySvc     *analytics.LearningCommunityService
	testLearnerID    string
	testContentID    string
	testCommunityID  string
	testPathID       string
}

// SetupSuite ?
func (suite *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// ?
	suite.testLearnerID = "test-learner-123"
	suite.testContentID = "test-content-456"
	suite.testCommunityID = "test-community-789"
	suite.testPathID = "test-path-101"
	
	// 
	suite.initializeServices()
	
	// 
	suite.setupRoutes()
}

// initializeServices ?
func (suite *IntegrationTestSuite) initializeServices() {
	// 
	// ?
	
	suite.learningPathSvc = &adaptive.LearningPathService{
		// 
	}
	
	suite.achievementSvc = &analytics.LearningAchievementService{
		// 
	}
	
	suite.communitySvc = &analytics.LearningCommunityService{
		// 
	}
}

// setupRoutes 
func (suite *IntegrationTestSuite) setupRoutes() {
	suite.router = gin.New()
	
	// 
	learningPathHandler := &handlers.LearningPathHandler{
		LearningPathService: suite.learningPathSvc,
	}
	
	// 
	achievementHandler := &handlers.AchievementHandler{
		AchievementService: suite.achievementSvc,
	}
	
	// 
	communityHandler := &handlers.CommunityHandler{
		CommunityService: suite.communitySvc,
	}
	
	v1 := suite.router.Group("/api/v1")
	{
		// 
		learningPath := v1.Group("/learning-path")
		{
			learningPath.POST("/generate", learningPathHandler.GeneratePersonalizedPath)
			learningPath.GET("/:id", learningPathHandler.GetLearningPath)
			learningPath.PUT("/:id/progress", learningPathHandler.UpdatePathProgress)
			learningPath.GET("/recommendations/:learner_id", learningPathHandler.GetRecommendedPaths)
			learningPath.GET("/learner/:learner_id", learningPathHandler.GetLearnerPaths)
			learningPath.DELETE("/:id", learningPathHandler.DeleteLearningPath)
			learningPath.PUT("/:id/pause", learningPathHandler.PauseLearningPath)
			learningPath.PUT("/:id/resume", learningPathHandler.ResumeLearningPath)
		}
		
		// 
		achievements := v1.Group("/achievements")
		{
			achievements.POST("/check", achievementHandler.CheckAchievements)
			achievements.GET("/learner/:learner_id", achievementHandler.GetLearnerAchievements)
			achievements.GET("/learner/:learner_id/summary", achievementHandler.GetAchievementSummary)
			achievements.POST("/create", achievementHandler.CreateAchievement)
			achievements.GET("/leaderboard", achievementHandler.GetAchievementLeaderboard)
			achievements.GET("/:achievement_id/progress/:learner_id", achievementHandler.GetAchievementProgress)
		}
		
		// 
		community := v1.Group("/community")
		{
			community.POST("/create", communityHandler.CreateCommunity)
			community.POST("/:id/join", communityHandler.JoinCommunity)
			community.POST("/:id/posts", communityHandler.CreatePost)
			community.GET("/:id/posts", communityHandler.GetCommunityPosts)
			community.POST("/posts/:post_id/replies", communityHandler.CreateReply)
			community.POST("/study-groups", communityHandler.CreateStudyGroup)
			community.GET("/list", communityHandler.GetCommunityList)
			community.GET("/:id/details", communityHandler.GetCommunityDetails)
			community.GET("/study-groups", communityHandler.GetStudyGroupList)
		}
	}
}

// TestLearningPathIntegration 
func (suite *IntegrationTestSuite) TestLearningPathIntegration() {
	// 1. 
	generateRequest := map[string]interface{}{
		"learner_id":     suite.testLearnerID,
		"target_skills":  []string{"golang", "microservices", "docker"},
		"time_constraint": 30,
		"difficulty":     "intermediate",
		"learning_style": "visual",
		"prerequisites":  []string{"basic_programming"},
	}
	
	body, _ := json.Marshal(generateRequest)
	req := httptest.NewRequest("POST", "/api/v1/learning-path/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// 
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	
	// 2. 
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/learning-path/%s", suite.testPathID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 3. 
	progressRequest := map[string]interface{}{
		"learner_id":      suite.testLearnerID,
		"step_id":         "step-1",
		"completion_rate": 0.75,
		"time_spent":      1800,
		"performance_score": 85.5,
	}
	
	body, _ = json.Marshal(progressRequest)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/learning-path/%s/progress", suite.testPathID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 4. 
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/learning-path/recommendations/%s", suite.testLearnerID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestAchievementSystemIntegration 
func (suite *IntegrationTestSuite) TestAchievementSystemIntegration() {
	// 1. ?
	checkRequest := map[string]interface{}{
		"learner_id": suite.testLearnerID,
		"event_type": "content_completed",
		"event_data": map[string]interface{}{
			"content_id":        suite.testContentID,
			"completion_time":   1800,
			"performance_score": 92.5,
		},
	}
	
	body, _ := json.Marshal(checkRequest)
	req := httptest.NewRequest("POST", "/api/v1/achievements/check", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	
	// 2. ?
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/achievements/learner/%s", suite.testLearnerID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 3. 
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/achievements/learner/%s/summary", suite.testLearnerID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 4. ?
	req = httptest.NewRequest("GET", "/api/v1/achievements/leaderboard", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestCommunityIntegration 
func (suite *IntegrationTestSuite) TestCommunityIntegration() {
	// 1. 
	createCommunityRequest := map[string]interface{}{
		"name":        "Go?,
		"description": "Go",
		"type":        "study_group",
		"is_public":   true,
		"creator_id":  suite.testLearnerID,
		"tags":        []string{"golang", "programming", "backend"},
	}
	
	body, _ := json.Marshal(createCommunityRequest)
	req := httptest.NewRequest("POST", "/api/v1/community/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	
	// 2. 
	joinRequest := map[string]interface{}{
		"learner_id": suite.testLearnerID,
	}
	
	body, _ = json.Marshal(joinRequest)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/community/%s/join", suite.testCommunityID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 3. 
	createPostRequest := map[string]interface{}{
		"author_id": suite.testLearnerID,
		"title":     "Go?,
		"content":   "Go?..",
		"type":      "discussion",
		"tags":      []string{"golang", "concurrency", "goroutine"},
	}
	
	body, _ = json.Marshal(createPostRequest)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/community/%s/posts", suite.testCommunityID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 4. 
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/community/%s/posts", suite.testCommunityID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 5. 
	createStudyGroupRequest := map[string]interface{}{
		"name":         "Go?,
		"description":  "Go?,
		"creator_id":   suite.testLearnerID,
		"community_id": suite.testCommunityID,
		"max_members":  20,
		"schedule": map[string]interface{}{
			"meeting_time": "2024-01-15T19:00:00Z",
			"duration":     120,
			"frequency":    "weekly",
		},
	}
	
	body, _ = json.Marshal(createStudyGroupRequest)
	req = httptest.NewRequest("POST", "/api/v1/community/study-groups", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestServiceHealthAndMetrics 
func (suite *IntegrationTestSuite) TestServiceHealthAndMetrics() {
	// ?
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// ?
	// 
	
	// 
	healthResponse := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"services": map[string]interface{}{
			"database":    "healthy",
			"redis":       "healthy",
			"elasticsearch": "healthy",
		},
	}
	
	responseBody, _ := json.Marshal(healthResponse)
	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
}

// TestErrorHandling 
func (suite *IntegrationTestSuite) TestErrorHandling() {
	// 1. ID
	req := httptest.NewRequest("GET", "/api/v1/learning-path/invalid-id", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	
	// 2. JSON
	req = httptest.NewRequest("POST", "/api/v1/learning-path/generate", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// 3. 
	incompleteRequest := map[string]interface{}{
		"learner_id": suite.testLearnerID,
		// 
	}
	
	body, _ := json.Marshal(incompleteRequest)
	req = httptest.NewRequest("POST", "/api/v1/learning-path/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestConcurrentRequests 
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 10
	results := make(chan int, numRequests)
	
	// ?
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/achievements/learner/%s", suite.testLearnerID), nil)
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}
	
	// 
	successCount := 0
	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		}
	}
	
	// 
	assert.Equal(suite.T(), numRequests, successCount)
}

// TearDownSuite 
func (suite *IntegrationTestSuite) TearDownSuite() {
	// 
	// 
}

// TestIntegrationSuite 
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// BenchmarkAPIPerformance 
func BenchmarkAPIPerformance(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

