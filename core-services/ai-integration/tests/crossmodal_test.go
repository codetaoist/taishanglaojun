package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
)

// MockCrossModalService 模拟跨模态推理服�?
type MockCrossModalService struct {
	mock.Mock
}

func (m *MockCrossModalService) ProcessCrossModalInference(ctx context.Context, req *services.CrossModalRequest) (*services.CrossModalResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.CrossModalResponse), args.Error(1)
}

// setupCrossModalTest 设置跨模态推理测试环�?
func setupCrossModalTest(t *testing.T) (*gin.Engine, *gorm.DB, *handlers.CrossModalHandler) {
	// 设置测试数据�?
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 创建测试路由
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建模拟服务
	mockService := &MockCrossModalService{}
	
	// 创建处理�?
	crossModalHandler := handlers.NewCrossModalHandler(mockService)

	// 设置路由
	v1 := router.Group("/api/v1")
	crossModalGroup := v1.Group("/crossmodal")
	{
		crossModalGroup.POST("/inference", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.ProcessCrossModalInference(c)
		})
		crossModalGroup.POST("/search", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.SemanticSearch(c)
		})
		crossModalGroup.POST("/match", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.ContentMatching(c)
		})
		crossModalGroup.POST("/qa", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.MultiModalQA(c)
		})
		crossModalGroup.POST("/scene", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.SceneUnderstanding(c)
		})
		crossModalGroup.POST("/emotion", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.EmotionAnalysis(c)
		})
		crossModalGroup.GET("/history", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.GetInferenceHistory(c)
		})
		crossModalGroup.GET("/stats", func(c *gin.Context) {
			c.Set("user_id", "test-user-123")
			crossModalHandler.GetInferenceStats(c)
		})
	}

	return router, db, crossModalHandler
}

// TestCrossModalInference 测试跨模态推�?
func TestCrossModalInference(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	tests := []struct {
		name           string
		request        services.CrossModalRequest
		expectedStatus int
		setupMock      func(*MockCrossModalService)
	}{
		{
			name: "语义搜索成功",
			request: services.CrossModalRequest{
				Type:  services.InferenceTypeSemanticSearch,
				Query: "寻找与猫相关的图�?,
				Inputs: []services.CrossModalInput{
					{
						Type:    "text",
						Content: "一只可爱的小猫在花园里玩�?,
					},
					{
						Type: "image",
						URL:  "https://example.com/cat.jpg",
					},
				},
				Config: services.CrossModalInferenceConfig{
					Provider:            "openai",
					Model:               "gpt-4",
					SimilarityThreshold: 0.7,
				},
			},
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockCrossModalService) {
				m.On("ProcessCrossModalInference", mock.Anything, mock.AnythingOfType("*services.CrossModalRequest")).Return(
					&services.CrossModalResponse{
						ID:        "test-response-123",
						RequestID: "test-request-123",
						Type:      services.InferenceTypeSemanticSearch,
						Results: []services.CrossModalResult{
							{
								ID:         "result-1",
								Type:       "semantic_match",
								Content:    "找到与猫相关的内容匹�?,
								Confidence: 0.85,
								Metadata: map[string]interface{}{
									"similarity_score": 0.85,
									"matched_content":  "小猫图片",
								},
							},
						},
						Confidence: 0.85,
						Timestamp:  time.Now(),
					}, nil)
			},
		},
		{
			name: "内容匹配成功",
			request: services.CrossModalRequest{
				Type: services.InferenceTypeContentMatching,
				Inputs: []services.CrossModalInput{
					{
						Type:    "text",
						Content: "美丽的日落景�?,
					},
					{
						Type: "image",
						URL:  "https://example.com/sunset.jpg",
					},
				},
				Config: services.CrossModalInferenceConfig{
					Provider:            "openai",
					Model:               "gpt-4-vision-preview",
					SimilarityThreshold: 0.6,
				},
			},
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockCrossModalService) {
				m.On("ProcessCrossModalInference", mock.Anything, mock.AnythingOfType("*services.CrossModalRequest")).Return(
					&services.CrossModalResponse{
						ID:        "test-response-124",
						RequestID: "test-request-124",
						Type:      services.InferenceTypeContentMatching,
						Results: []services.CrossModalResult{
							{
								ID:         "result-2",
								Type:       "content_match",
								Content:    "文本和图片内容高度匹�?,
								Confidence: 0.92,
								Metadata: map[string]interface{}{
									"match_type":       "semantic_visual",
									"similarity_score": 0.92,
								},
							},
						},
						Confidence: 0.92,
						Timestamp:  time.Now(),
					}, nil)
			},
		},
		{
			name: "多模态问答成�?,
			request: services.CrossModalRequest{
				Type:  services.InferenceTypeMultiModalQA,
				Query: "这张图片中的动物是什么？",
				Inputs: []services.CrossModalInput{
					{
						Type: "image",
						URL:  "https://example.com/animal.jpg",
					},
				},
				Config: services.CrossModalInferenceConfig{
					Provider:    "openai",
					Model:       "gpt-4-vision-preview",
					Temperature: 0.7,
				},
			},
			expectedStatus: http.StatusOK,
			setupMock: func(m *MockCrossModalService) {
				m.On("ProcessCrossModalInference", mock.Anything, mock.AnythingOfType("*services.CrossModalRequest")).Return(
					&services.CrossModalResponse{
						ID:        "test-response-125",
						RequestID: "test-request-125",
						Type:      services.InferenceTypeMultiModalQA,
						Results: []services.CrossModalResult{
							{
								ID:         "result-3",
								Type:       "qa_answer",
								Content:    "这是一只金毛犬，看起来很友好和活泼�?,
								Confidence: 0.88,
								Metadata: map[string]interface{}{
									"detected_objects": []string{"dog", "golden_retriever"},
									"confidence_score": 0.88,
								},
							},
						},
						Confidence: 0.88,
						Timestamp:  time.Now(),
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置模拟
			mockService := &MockCrossModalService{}
			tt.setupMock(mockService)

			// 创建请求
			requestBody, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/crossmodal/inference", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Session-ID", "test-session-123")

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response services.CrossModalResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.NotEmpty(t, response.Results)
			}

			// 验证模拟调用
			mockService.AssertExpectations(t)
		})
	}
}

// TestSemanticSearch 测试语义搜索
func TestSemanticSearch(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type:    "text",
			Content: "寻找关于人工智能的内�?,
		},
		{
			Type: "image",
			URL:  "https://example.com/ai.jpg",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/search?query=人工智能&provider=openai&model=gpt-4&max_results=5&threshold=0.8", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestContentMatching 测试内容匹配
func TestContentMatching(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type:    "text",
			Content: "美丽的风�?,
		},
		{
			Type: "image",
			URL:  "https://example.com/landscape.jpg",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/match?provider=openai&threshold=0.7", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestMultiModalQA 测试多模态问�?
func TestMultiModalQA(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type: "image",
			URL:  "https://example.com/question.jpg",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/qa?query=这张图片显示了什么？&provider=openai&model=gpt-4-vision-preview", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSceneUnderstanding 测试场景理解
func TestSceneUnderstanding(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type: "image",
			URL:  "https://example.com/scene.jpg",
		},
		{
			Type:    "text",
			Content: "分析这个场景的环境和氛围",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/scene?provider=openai", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestEmotionAnalysis 测试情感分析
func TestEmotionAnalysis(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type:    "text",
			Content: "今天心情很好，阳光明�?,
		},
		{
			Type: "audio",
			URL:  "https://example.com/happy_voice.mp3",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/emotion?provider=openai", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetInferenceHistory 测试获取推理历史
func TestGetInferenceHistory(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	req, _ := http.NewRequest("GET", "/api/v1/crossmodal/history?limit=10&offset=0", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "user_id")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "records")
}

// TestGetInferenceStats 测试获取推理统计
func TestGetInferenceStats(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	req, _ := http.NewRequest("GET", "/api/v1/crossmodal/stats?period=7d", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "user_id")
	assert.Contains(t, response, "total_inferences")
	assert.Contains(t, response, "inference_types")
	assert.Contains(t, response, "success_rate")
}

// TestInvalidRequests 测试无效请求
func TestInvalidRequests(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		expectedStatus int
	}{
		{
			name:           "无效JSON格式",
			method:         "POST",
			url:            "/api/v1/crossmodal/inference",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "缺少查询参数",
			method:         "POST",
			url:            "/api/v1/crossmodal/search",
			body:           "[]",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "内容匹配输入不足",
			method:         "POST",
			url:            "/api/v1/crossmodal/match",
			body:           `[{"type": "text", "content": "单个输入"}]`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "多模态问答缺少查�?,
			method:         "POST",
			url:            "/api/v1/crossmodal/qa",
			body:           "[]",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// BenchmarkCrossModalInference 跨模态推理性能测试
func BenchmarkCrossModalInference(b *testing.B) {
	router, db, _ := setupCrossModalTest(&testing.T{})
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	request := services.CrossModalRequest{
		Type:  services.InferenceTypeSemanticSearch,
		Query: "测试查询",
		Inputs: []services.CrossModalInput{
			{
				Type:    "text",
				Content: "测试内容",
			},
		},
		Config: services.CrossModalInferenceConfig{
			Provider: "openai",
			Model:    "gpt-4",
		},
	}

	requestBody, _ := json.Marshal(request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/crossmodal/inference", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
