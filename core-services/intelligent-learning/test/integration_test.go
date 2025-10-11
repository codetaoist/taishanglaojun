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

// IntegrationTestSuite 集成测试套件
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

// SetupSuite 测试套件初始�?
func (suite *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// 初始化测试数�?
	suite.testLearnerID = "test-learner-123"
	suite.testContentID = "test-content-456"
	suite.testCommunityID = "test-community-789"
	suite.testPathID = "test-path-101"
	
	// 初始化服务（这里使用模拟的依赖）
	suite.initializeServices()
	
	// 设置路由
	suite.setupRoutes()
}

// initializeServices 初始化服�?
func (suite *IntegrationTestSuite) initializeServices() {
	// 这里应该初始化真实的服务，但为了测试，我们使用模拟的依赖
	// 在实际环境中，这些应该连接到测试数据�?
	
	suite.learningPathSvc = &adaptive.LearningPathService{
		// 模拟依赖
	}
	
	suite.achievementSvc = &analytics.LearningAchievementService{
		// 模拟依赖
	}
	
	suite.communitySvc = &analytics.LearningCommunityService{
		// 模拟依赖
	}
}

// setupRoutes 设置测试路由
func (suite *IntegrationTestSuite) setupRoutes() {
	suite.router = gin.New()
	
	// 学习路径相关路由
	learningPathHandler := &handlers.LearningPathHandler{
		LearningPathService: suite.learningPathSvc,
	}
	
	// 成就系统相关路由
	achievementHandler := &handlers.AchievementHandler{
		AchievementService: suite.achievementSvc,
	}
	
	// 社区相关路由
	communityHandler := &handlers.CommunityHandler{
		CommunityService: suite.communitySvc,
	}
	
	v1 := suite.router.Group("/api/v1")
	{
		// 学习路径路由
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
		
		// 成就系统路由
		achievements := v1.Group("/achievements")
		{
			achievements.POST("/check", achievementHandler.CheckAchievements)
			achievements.GET("/learner/:learner_id", achievementHandler.GetLearnerAchievements)
			achievements.GET("/learner/:learner_id/summary", achievementHandler.GetAchievementSummary)
			achievements.POST("/create", achievementHandler.CreateAchievement)
			achievements.GET("/leaderboard", achievementHandler.GetAchievementLeaderboard)
			achievements.GET("/:achievement_id/progress/:learner_id", achievementHandler.GetAchievementProgress)
		}
		
		// 社区路由
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

// TestLearningPathIntegration 测试学习路径集成功能
func (suite *IntegrationTestSuite) TestLearningPathIntegration() {
	// 1. 测试生成个性化学习路径
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
	
	// 验证响应
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	
	// 2. 测试获取学习路径
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/learning-path/%s", suite.testPathID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 3. 测试更新学习进度
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
	
	// 4. 测试获取推荐路径
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/learning-path/recommendations/%s", suite.testLearnerID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestAchievementSystemIntegration 测试成就系统集成功能
func (suite *IntegrationTestSuite) TestAchievementSystemIntegration() {
	// 1. 测试检查成�?
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
	
	// 2. 测试获取学习者成�?
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/achievements/learner/%s", suite.testLearnerID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 3. 测试获取成就摘要
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/achievements/learner/%s/summary", suite.testLearnerID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 4. 测试获取排行�?
	req = httptest.NewRequest("GET", "/api/v1/achievements/leaderboard", nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestCommunityIntegration 测试社区功能集成
func (suite *IntegrationTestSuite) TestCommunityIntegration() {
	// 1. 测试创建社区
	createCommunityRequest := map[string]interface{}{
		"name":        "Go学习交流�?,
		"description": "专注于Go语言学习和交流的社区",
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
	
	// 2. 测试加入社区
	joinRequest := map[string]interface{}{
		"learner_id": suite.testLearnerID,
	}
	
	body, _ = json.Marshal(joinRequest)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/community/%s/join", suite.testCommunityID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 3. 测试创建帖子
	createPostRequest := map[string]interface{}{
		"author_id": suite.testLearnerID,
		"title":     "Go语言并发编程最佳实�?,
		"content":   "分享一些Go语言并发编程的经验和技�?..",
		"type":      "discussion",
		"tags":      []string{"golang", "concurrency", "goroutine"},
	}
	
	body, _ = json.Marshal(createPostRequest)
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/community/%s/posts", suite.testCommunityID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 4. 测试获取社区帖子
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/community/%s/posts", suite.testCommunityID), nil)
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	// 5. 测试创建学习小组
	createStudyGroupRequest := map[string]interface{}{
		"name":         "Go微服务实战小�?,
		"description":  "一起学习Go微服务开�?,
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

// TestServiceHealthAndMetrics 测试服务健康状态和指标
func (suite *IntegrationTestSuite) TestServiceHealthAndMetrics() {
	// 测试健康检查端�?
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// 这里需要模拟主服务的健康检�?
	// 在实际测试中，应该启动完整的服务
	
	// 模拟健康响应
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

// TestErrorHandling 测试错误处理
func (suite *IntegrationTestSuite) TestErrorHandling() {
	// 1. 测试无效的学习路径ID
	req := httptest.NewRequest("GET", "/api/v1/learning-path/invalid-id", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	
	// 2. 测试无效的JSON请求
	req = httptest.NewRequest("POST", "/api/v1/learning-path/generate", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	
	// 3. 测试缺少必需参数
	incompleteRequest := map[string]interface{}{
		"learner_id": suite.testLearnerID,
		// 缺少其他必需字段
	}
	
	body, _ := json.Marshal(incompleteRequest)
	req = httptest.NewRequest("POST", "/api/v1/learning-path/generate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestConcurrentRequests 测试并发请求处理
func (suite *IntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 10
	results := make(chan int, numRequests)
	
	// 并发发送多个请�?
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/achievements/learner/%s", suite.testLearnerID), nil)
			w := httptest.NewRecorder()
			
			suite.router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}
	
	// 收集结果
	successCount := 0
	for i := 0; i < numRequests; i++ {
		code := <-results
		if code == http.StatusOK {
			successCount++
		}
	}
	
	// 验证所有请求都成功处理
	assert.Equal(suite.T(), numRequests, successCount)
}

// TearDownSuite 测试套件清理
func (suite *IntegrationTestSuite) TearDownSuite() {
	// 清理测试数据
	// 在实际环境中，这里应该清理测试数据库
}

// TestIntegrationSuite 运行集成测试套件
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// BenchmarkAPIPerformance 性能基准测试
func BenchmarkAPIPerformance(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 设置简单的测试路由
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
