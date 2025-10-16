package tests

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
	"github.com/stretchr/testify/require"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/advanced"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/config"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
)

// TestSuite 
type TestSuite struct {
	router  *gin.Engine
	service *advanced.AdvancedAIService
	handler *handlers.AdvancedAIHandler
	config  *config.AdvancedAIConfig
	ctx     context.Context
	cancel  context.CancelFunc
}

// SetupTestSuite 
func SetupTestSuite(t *testing.T) *TestSuite {
	// Gin
	gin.SetMode(gin.TestMode)

	// 
	cfg := config.DefaultAdvancedAIConfig()
	cfg.AGI.MaxConcurrentTasks = 5
	cfg.MetaLearning.MaxConcurrentSessions = 3
	cfg.SelfEvolution.MaxConcurrentEvolutions = 2

	// 
	service := advanced.NewAdvancedAIService(cfg)

	// 
	handler := handlers.NewAdvancedAIHandler(service)

	// 
	router := gin.New()
	router.Use(gin.Recovery())

	// 
	v1 := router.Group("/api/v1/advanced-ai")
	{
		// 
		v1.POST("/process", handler.ProcessRequest)
		v1.POST("/agi/task", handler.ProcessAGITask)
		v1.POST("/meta-learning/learn", handler.ProcessMetaLearning)
		v1.POST("/evolution/optimize", handler.TriggerEvolution)

		// 
		v1.GET("/status", handler.GetStatus)
		v1.GET("/metrics", handler.GetMetrics)
		v1.GET("/health", handler.HealthCheck)

		// 
		v1.GET("/config", handler.GetConfig)
		v1.PUT("/config", handler.UpdateConfig)
		v1.GET("/capabilities", handler.GetCapabilities)
		v1.GET("/history", handler.GetHistory)
		v1.GET("/statistics", handler.GetStatistics)

		// 
		v1.POST("/initialize", handler.Initialize)
		v1.POST("/shutdown", handler.Shutdown)
		v1.POST("/reset", handler.Reset)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TestSuite{
		router:  router,
		service: service,
		handler: handler,
		config:  cfg,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// TearDownTestSuite 
func (ts *TestSuite) TearDownTestSuite() {
	ts.cancel()
	if ts.service != nil {
		ts.service.Shutdown(ts.ctx)
	}
}

// TestBasicFunctionality 
func TestBasicFunctionality(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	t.Run("Health Check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/health", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "version")
	})

	t.Run("Get Status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/status", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "overall_health")
		assert.Contains(t, response, "capabilities")
		assert.Contains(t, response, "resource_usage")
	})

	t.Run("Get Capabilities", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/capabilities", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		capabilities := response["capabilities"].(map[string]interface{})
		assert.Contains(t, capabilities, "agi")
		assert.Contains(t, capabilities, "meta_learning")
		assert.Contains(t, capabilities, "self_evolution")
	})
}

// TestAGIIntegration AGI
func TestAGIIntegration(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	testCases := []struct {
		name     string
		taskType string
		input    map[string]interface{}
		expected int
	}{
		{
			name:     "Reasoning Task",
			taskType: "reasoning",
			input: map[string]interface{}{
				"problem": "",
				"context": map[string]interface{}{
					"city_size":  "large",
					"population": 5000000,
				},
			},
			expected: http.StatusOK,
		},
		{
			name:     "Planning Task",
			taskType: "planning",
			input: map[string]interface{}{
				"goal": "",
				"constraints": map[string]interface{}{
					"budget":   1000000,
					"timeline": "3",
				},
			},
			expected: http.StatusOK,
		},
		{
			name:     "Creative Generation",
			taskType: "creative_generation",
			input: map[string]interface{}{
				"prompt": "",
				"style":  "",
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"type":  tc.taskType,
				"input": tc.input,
				"requirements": map[string]interface{}{
					"detail_level": "high",
				},
				"priority": 1,
				"timeout":  300,
			}

			body, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/agi/task", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "request_id")
				assert.Contains(t, response, "result")
				assert.Contains(t, response, "confidence")
				assert.Contains(t, response, "process_time")
				assert.Contains(t, response, "used_capabilities")

				// 
				confidence := response["confidence"].(float64)
				assert.GreaterOrEqual(t, confidence, 0.0)
				assert.LessOrEqual(t, confidence, 1.0)
			}
		})
	}
}

// TestMetaLearningIntegration 
func TestMetaLearningIntegration(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	testCases := []struct {
		name     string
		taskType string
		domain   string
		strategy string
		expected int
	}{
		{
			name:     "Few-Shot Learning",
			taskType: "few_shot_learning",
			domain:   "image_classification",
			strategy: "few_shot",
			expected: http.StatusOK,
		},
		{
			name:     "Transfer Learning",
			taskType: "transfer_learning",
			domain:   "natural_language_processing",
			strategy: "transfer_learning",
			expected: http.StatusOK,
		},
		{
			name:     "Online Adaptation",
			taskType: "online_adaptation",
			domain:   "recommendation_system",
			strategy: "online_adaptation",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"task_type": tc.taskType,
				"domain":    tc.domain,
				"data": []map[string]interface{}{
					{
						"input": map[string]interface{}{
							"features": []float64{1.0, 2.0, 3.0},
						},
						"label": "positive",
					},
					{
						"input": map[string]interface{}{
							"features": []float64{-1.0, -2.0, -3.0},
						},
						"label": "negative",
					},
				},
				"strategy": tc.strategy,
				"parameters": map[string]interface{}{
					"learning_rate": 0.001,
					"epochs":        10,
				},
			}

			body, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/meta-learning/learn", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "request_id")
				assert.Contains(t, response, "result")

				result := response["result"].(map[string]interface{})
				assert.Contains(t, result, "model_id")
				assert.Contains(t, result, "learning_performance")
				assert.Contains(t, result, "adaptation_capability")
			}
		})
	}
}

// TestSelfEvolutionIntegration 
func TestSelfEvolutionIntegration(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	testCases := []struct {
		name     string
		strategy string
		expected int
	}{
		{
			name:     "Genetic Evolution",
			strategy: "genetic",
			expected: http.StatusOK,
		},
		{
			name:     "Neuro Evolution",
			strategy: "neuro_evolution",
			expected: http.StatusOK,
		},
		{
			name:     "Gradient Free",
			strategy: "gradient_free",
			expected: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"optimization_targets": []map[string]interface{}{
					{
						"metric":        "accuracy",
						"weight":        0.6,
						"target_value":  0.95,
						"current_value": 0.85,
						"direction":     "maximize",
					},
					{
						"metric":        "inference_speed",
						"weight":        0.4,
						"target_value":  50.0,
						"current_value": 100.0,
						"direction":     "minimize",
					},
				},
				"strategy": tc.strategy,
				"parameters": map[string]interface{}{
					"population_size": 20,
					"generations":     10,
					"mutation_rate":   0.1,
				},
			}

			body, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/evolution/optimize", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expected, w.Code)

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.True(t, response["success"].(bool))
				assert.Contains(t, response, "request_id")
				assert.Contains(t, response, "result")

				result := response["result"].(map[string]interface{})
				assert.Contains(t, result, "evolution_id")
				assert.Contains(t, result, "status")
				assert.Contains(t, result, "optimization_progress")
			}
		})
	}
}

// TestHybridMode 
func TestHybridMode(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	t.Run("Hybrid Processing", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"type":       "reasoning",
			"capability": "hybrid",
			"input": map[string]interface{}{
				"problem": "AI?",
				"context": map[string]interface{}{
					"model_type": "transformer",
					"dataset":    "large_scale",
				},
			},
			"requirements": map[string]interface{}{
				"use_meta_learning": true,
				"enable_evolution":  true,
				"detail_level":      "high",
			},
		}

		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/process", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Contains(t, response, "used_capabilities")

		usedCapabilities := response["used_capabilities"].([]interface{})
		assert.GreaterOrEqual(t, len(usedCapabilities), 2) // 
	})
}

// TestConcurrentRequests 
func TestConcurrentRequests(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	t.Run("Concurrent AGI Tasks", func(t *testing.T) {
		concurrency := 5
		results := make(chan bool, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				requestBody := map[string]interface{}{
					"type": "reasoning",
					"input": map[string]interface{}{
						"problem": fmt.Sprintf(" %d: ?", id),
					},
				}

				body, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/agi/task", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				ts.router.ServeHTTP(w, req)

				results <- w.Code == http.StatusOK
			}(i)
		}

		// 
		successCount := 0
		for i := 0; i < concurrency; i++ {
			if <-results {
				successCount++
			}
		}

		// 80%
		assert.GreaterOrEqual(t, successCount, int(float64(concurrency)*0.8))
	})
}

// TestErrorHandling 
func TestErrorHandling(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	testCases := []struct {
		name         string
		endpoint     string
		method       string
		body         interface{}
		expectedCode int
	}{
		{
			name:         "Invalid JSON",
			endpoint:     "/api/v1/advanced-ai/process",
			method:       "POST",
			body:         "invalid json",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:     "Missing Required Field",
			endpoint: "/api/v1/advanced-ai/agi/task",
			method:   "POST",
			body: map[string]interface{}{
				"input": map[string]interface{}{
					"problem": "test",
				},
				//  type 
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid Endpoint",
			endpoint:     "/api/v1/advanced-ai/invalid",
			method:       "GET",
			body:         nil,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error

			if tc.body != nil {
				if bodyStr, ok := tc.body.(string); ok {
					req, err = http.NewRequest(tc.method, tc.endpoint, bytes.NewBufferString(bodyStr))
				} else {
					body, _ := json.Marshal(tc.body)
					req, err = http.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(body))
				}
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tc.method, tc.endpoint, nil)
			}

			require.NoError(t, err)

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// TestPerformanceMetrics 
func TestPerformanceMetrics(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	// 
	for i := 0; i < 5; i++ {
		requestBody := map[string]interface{}{
			"type": "reasoning",
			"input": map[string]interface{}{
				"problem": fmt.Sprintf(" %d", i),
			},
		}

		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/process", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)
	}

	// 
	time.Sleep(100 * time.Millisecond)

	t.Run("Get Metrics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/metrics", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "metrics")
		assert.Contains(t, response, "count")
	})

	t.Run("Get Statistics", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/statistics", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "total_requests")
		assert.Contains(t, response, "success_rate")
		assert.Contains(t, response, "capability_stats")
	})
}

// TestConfigurationManagement 
func TestConfigurationManagement(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	t.Run("Get Config", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/config", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "enable_agi")
		assert.Contains(t, response, "enable_meta_learning")
		assert.Contains(t, response, "enable_self_evolution")
	})

	t.Run("Update Config", func(t *testing.T) {
		updateBody := map[string]interface{}{
			"max_concurrent_requests": 100,
			"default_timeout":         "60s",
			"log_level":               "debug",
		}

		body, _ := json.Marshal(updateBody)
		req, _ := http.NewRequest("PUT", "/api/v1/advanced-ai/config", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.Contains(t, response, "updated_fields")
	})
}

// BenchmarkProcessRequest 
func BenchmarkProcessRequest(b *testing.B) {
	ts := SetupTestSuite(&testing.T{})
	defer ts.TearDownTestSuite()

	requestBody := map[string]interface{}{
		"type": "reasoning",
		"input": map[string]interface{}{
			"problem": "",
		},
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/process", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// TestSystemLifecycle 
func TestSystemLifecycle(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	t.Run("Initialize", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"config": map[string]interface{}{
				"enable_agi":            true,
				"enable_meta_learning":  true,
				"enable_self_evolution": true,
			},
			"force": false,
		}

		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/initialize", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.Contains(t, response, "initialized")
	})

	t.Run("Reset", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"reset_data":    false,
			"reset_config":  false,
			"reset_metrics": true,
		}

		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/reset", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "message")
		assert.True(t, response["reset_metrics"].(bool))
	})
}

// TestDataValidation 
func TestDataValidation(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	testCases := []struct {
		name         string
		requestBody  map[string]interface{}
		expectedCode int
	}{
		{
			name: "Valid Request",
			requestBody: map[string]interface{}{
				"type": "reasoning",
				"input": map[string]interface{}{
					"problem": "?",
				},
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Empty Input",
			requestBody: map[string]interface{}{
				"type":  "reasoning",
				"input": map[string]interface{}{},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid Type",
			requestBody: map[string]interface{}{
				"type": "invalid_type",
				"input": map[string]interface{}{
					"problem": "",
				},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid Priority",
			requestBody: map[string]interface{}{
				"type": "reasoning",
				"input": map[string]interface{}{
					"problem": "",
				},
				"priority": 10, //  (1-5)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/process", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// TestResourceManagement 
func TestResourceManagement(t *testing.T) {
	ts := SetupTestSuite(t)
	defer ts.TearDownTestSuite()

	t.Run("Resource Usage Monitoring", func(t *testing.T) {
		// 
		for i := 0; i < 10; i++ {
			requestBody := map[string]interface{}{
				"type": "planning",
				"input": map[string]interface{}{
					"goal": fmt.Sprintf("滮 %d", i),
					"constraints": map[string]interface{}{
						"complexity": "high",
					},
				},
			}

			body, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/advanced-ai/agi/task", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			ts.router.ServeHTTP(w, req)
		}

		// 
		req, _ := http.NewRequest("GET", "/api/v1/advanced-ai/status", nil)
		w := httptest.NewRecorder()
		ts.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		resourceUsage := response["resource_usage"].(map[string]interface{})
		assert.Contains(t, resourceUsage, "cpu")
		assert.Contains(t, resourceUsage, "memory")

		// 
		cpuUsage := resourceUsage["cpu"].(float64)
		memoryUsage := resourceUsage["memory"].(float64)
		assert.GreaterOrEqual(t, cpuUsage, 0.0)
		assert.LessOrEqual(t, cpuUsage, 1.0)
		assert.GreaterOrEqual(t, memoryUsage, 0.0)
		assert.LessOrEqual(t, memoryUsage, 1.0)
	})
}

