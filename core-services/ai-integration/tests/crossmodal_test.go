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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
)

// MockCrossModalService 
type MockCrossModalService struct {
	mock.Mock
}

func (m *MockCrossModalService) ProcessCrossModalInference(ctx context.Context, req *services.CrossModalRequest) (*services.CrossModalResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*services.CrossModalResponse), args.Error(1)
}

// setupCrossModalTest 
func setupCrossModalTest(t *testing.T) (*gin.Engine, *gorm.DB, *handlers.CrossModalHandler) {
	// 
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 
	mockService := &MockCrossModalService{}

	// 
	crossModalHandler := handlers.NewCrossModalHandler(mockService)

	// 
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

// TestCrossModalInference 
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
			name: "",
			request: services.CrossModalRequest{
				Type:  services.InferenceTypeSemanticSearch,
				Query: "",
				Inputs: []services.CrossModalInput{
					{
						Type:    "text",
						Content: "",
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
								Content:    "",
								Confidence: 0.85,
								Metadata: map[string]interface{}{
									"similarity_score": 0.85,
									"matched_content":  "",
								},
							},
						},
						Confidence: 0.85,
						Timestamp:  time.Now(),
					}, nil)
			},
		},
		{
			name: "",
			request: services.CrossModalRequest{
				Type: services.InferenceTypeContentMatching,
				Inputs: []services.CrossModalInput{
					{
						Type:    "text",
						Content: "䳡",
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
								Content:    "",
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
			name: "",
			request: services.CrossModalRequest{
				Type:  services.InferenceTypeMultiModalQA,
				Query: "",
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
								Content:    "",
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
			// 
			mockService := &MockCrossModalService{}
			tt.setupMock(mockService)

			// 
			requestBody, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/crossmodal/inference", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Session-ID", "test-session-123")

			// 
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 
			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {
				var response services.CrossModalResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, tt.request.Type, response.Type)
				assert.NotEmpty(t, response.Results)
			}

			// 
			mockService.AssertExpectations(t)
		})
	}
}

// TestSemanticSearch 
func TestSemanticSearch(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type:    "text",
			Content: "",
		},
		{
			Type: "image",
			URL:  "https://example.com/ai.jpg",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/search?query=&provider=openai&model=gpt-4&max_results=5&threshold=0.8", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestContentMatching 
func TestContentMatching(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type:    "text",
			Content: "䳡",
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

// TestMultiModalQA 
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
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/qa?query=&provider=openai&model=gpt-4-vision-preview", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSceneUnderstanding 
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
			Content: "",
		},
	}

	requestBody, _ := json.Marshal(inputs)
	req, _ := http.NewRequest("POST", "/api/v1/crossmodal/scene?provider=openai", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestEmotionAnalysis 
func TestEmotionAnalysis(t *testing.T) {
	router, db, _ := setupCrossModalTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	inputs := []services.CrossModalInput{
		{
			Type:    "text",
			Content: "",
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

// TestGetInferenceHistory 
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

// TestGetInferenceStats 
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

// TestInvalidRequests 
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
			name:           "JSON",
			method:         "POST",
			url:            "/api/v1/crossmodal/inference",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "",
			method:         "POST",
			url:            "/api/v1/crossmodal/search",
			body:           "[]",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "",
			method:         "POST",
			url:            "/api/v1/crossmodal/match",
			body:           `[{"type": "text", "content": ""}]`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "",
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

// BenchmarkCrossModalInference 
func BenchmarkCrossModalInference(b *testing.B) {
	router, db, _ := setupCrossModalTest(&testing.T{})
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	request := services.CrossModalRequest{
		Type:  services.InferenceTypeSemanticSearch,
		Query: "",
		Inputs: []services.CrossModalInput{
			{
				Type:    "text",
				Content: "",
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

