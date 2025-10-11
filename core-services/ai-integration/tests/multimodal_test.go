package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/handlers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
)

func setupMultimodalTest(t *testing.T) (*gin.Engine, *gorm.DB, *handlers.MultimodalHandler) {
	// 设置测试数据�?
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 迁移�?
	err = db.AutoMigrate(
		&models.MultimodalSession{},
		&models.MultimodalMessage{},
	)
	require.NoError(t, err)

	// 创建服务和处理器
	multimodalService := services.NewMultimodalService(db, nil, nil)
	multimodalHandler := handlers.NewMultimodalHandler(multimodalService, nil)

	// 设置路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	v1 := router.Group("/api/v1")
	multimodal := v1.Group("/multimodal")
	{
		multimodal.POST("/process", multimodalHandler.ProcessMultimodal)
		multimodal.POST("/upload", multimodalHandler.UploadFile)
		multimodal.GET("/sessions", multimodalHandler.ListSessions)
		multimodal.POST("/sessions", multimodalHandler.CreateSession)
		multimodal.GET("/sessions/:id", multimodalHandler.GetSession)
		multimodal.PUT("/sessions/:id", multimodalHandler.UpdateSession)
		multimodal.DELETE("/sessions/:id", multimodalHandler.DeleteSession)
		multimodal.GET("/sessions/:id/messages", multimodalHandler.GetSessionMessages)
	}

	return router, db, multimodalHandler
}

func TestMultimodalProcess(t *testing.T) {
	router, db, _ := setupMultimodalTest(t)

	tests := []struct {
		name           string
		request        models.MultimodalRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "文本输入处理",
			request: models.MultimodalRequest{
				SessionID: "test-session-1",
				Inputs: []models.MultimodalInput{
					{
						Type: "text",
						Content: models.MultimodalContent{
							Text: "Hello, how are you?",
						},
					},
				},
				Config: models.MultimodalConfig{
					Provider:    "openai",
					Model:       "gpt-4",
					MaxTokens:   1000,
					Temperature: 0.7,
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "多模态输入处�?,
			request: models.MultimodalRequest{
				SessionID: "test-session-2",
				Inputs: []models.MultimodalInput{
					{
						Type: "text",
						Content: models.MultimodalContent{
							Text: "Describe this image:",
						},
					},
					{
						Type: "image",
						Content: models.MultimodalContent{
							URL: "https://example.com/image.jpg",
						},
					},
				},
				Config: models.MultimodalConfig{
					Provider:    "openai",
					Model:       "gpt-4-vision-preview",
					MaxTokens:   1000,
					Temperature: 0.7,
				},
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "无效输入类型",
			request: models.MultimodalRequest{
				SessionID: "test-session-3",
				Inputs: []models.MultimodalInput{
					{
						Type: "invalid",
						Content: models.MultimodalContent{
							Text: "Invalid input type",
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建会话
			session := &models.MultimodalSession{
				ID:        tt.request.SessionID,
				UserID:    "test-user",
				Title:     "Test Session",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			db.Create(session)

			// 准备请求
			jsonData, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/v1/multimodal/process", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				var response models.MultimodalResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.RequestID)
			}
		})
	}
}

func TestFileUpload(t *testing.T) {
	router, _, _ := setupMultimodalTest(t)

	// 创建临时测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// 准备multipart请求
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加文件
	file, err := os.Open(testFile)
	require.NoError(t, err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = io.Copy(part, file)
	require.NoError(t, err)

	// 添加其他字段
	writer.WriteField("type", "text")
	writer.WriteField("session_id", "test-session")
	writer.Close()

	// 创建请求
	req, err := http.NewRequest("POST", "/api/v1/multimodal/upload", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "file_id")
	assert.Contains(t, response, "url")
}

func TestSessionManagement(t *testing.T) {
	router, db, _ := setupMultimodalTest(t)

	// 测试创建会话
	t.Run("创建会话", func(t *testing.T) {
		createReq := map[string]interface{}{
			"user_id": "test-user",
			"title":   "Test Session",
			"config": map[string]interface{}{
				"provider": "openai",
				"model":    "gpt-4",
			},
		}

		jsonData, err := json.Marshal(createReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/v1/multimodal/sessions", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.MultimodalSession
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "test-user", response.UserID)
		assert.Equal(t, "Test Session", response.Title)
	})

	// 测试获取会话列表
	t.Run("获取会话列表", func(t *testing.T) {
		// 先创建一些测试会�?
		sessions := []models.MultimodalSession{
			{
				ID:        "session-1",
				UserID:    "test-user",
				Title:     "Session 1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "session-2",
				UserID:    "test-user",
				Title:     "Session 2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		for _, session := range sessions {
			db.Create(&session)
		}

		req, err := http.NewRequest("GET", "/api/v1/multimodal/sessions?user_id=test-user", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []models.MultimodalSession
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2)
	})

	// 测试获取单个会话
	t.Run("获取单个会话", func(t *testing.T) {
		session := &models.MultimodalSession{
			ID:        "session-detail",
			UserID:    "test-user",
			Title:     "Detail Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(session)

		req, err := http.NewRequest("GET", "/api/v1/multimodal/sessions/session-detail", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.MultimodalSession
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "session-detail", response.ID)
		assert.Equal(t, "Detail Session", response.Title)
	})

	// 测试更新会话
	t.Run("更新会话", func(t *testing.T) {
		session := &models.MultimodalSession{
			ID:        "session-update",
			UserID:    "test-user",
			Title:     "Original Title",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(session)

		updateReq := map[string]interface{}{
			"title": "Updated Title",
		}

		jsonData, err := json.Marshal(updateReq)
		require.NoError(t, err)

		req, err := http.NewRequest("PUT", "/api/v1/multimodal/sessions/session-update", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.MultimodalSession
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", response.Title)
	})

	// 测试删除会话
	t.Run("删除会话", func(t *testing.T) {
		session := &models.MultimodalSession{
			ID:        "session-delete",
			UserID:    "test-user",
			Title:     "Delete Session",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(session)

		req, err := http.NewRequest("DELETE", "/api/v1/multimodal/sessions/session-delete", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证会话已被删除
		var count int64
		db.Model(&models.MultimodalSession{}).Where("id = ?", "session-delete").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestSessionMessages(t *testing.T) {
	router, db, _ := setupMultimodalTest(t)

	// 创建测试会话
	session := &models.MultimodalSession{
		ID:        "session-messages",
		UserID:    "test-user",
		Title:     "Messages Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(session)

	// 创建测试消息
	messages := []models.MultimodalMessage{
		{
			ID:        "msg-1",
			SessionID: "session-messages",
			Role:      "user",
			Content: models.MultimodalContent{
				Text: "Hello",
			},
			CreatedAt: time.Now(),
		},
		{
			ID:        "msg-2",
			SessionID: "session-messages",
			Role:      "assistant",
			Content: models.MultimodalContent{
				Text: "Hi there!",
			},
			CreatedAt: time.Now(),
		},
	}

	for _, msg := range messages {
		db.Create(&msg)
	}

	// 测试获取会话消息
	req, err := http.NewRequest("GET", "/api/v1/multimodal/sessions/session-messages/messages", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.MultimodalMessage
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "user", response[0].Role)
	assert.Equal(t, "assistant", response[1].Role)
}

func TestInputValidation(t *testing.T) {
	router, _, _ := setupMultimodalTest(t)

	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
	}{
		{
			name:           "空请求体",
			request:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "缺少会话ID",
			request: map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"type": "text",
						"content": map[string]interface{}{
							"text": "Hello",
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "空输入数�?,
			request: map[string]interface{}{
				"session_id": "test-session",
				"inputs":     []interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "无效输入类型",
			request: map[string]interface{}{
				"session_id": "test-session",
				"inputs": []map[string]interface{}{
					{
						"type": "invalid_type",
						"content": map[string]interface{}{
							"text": "Hello",
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonData []byte
			var err error

			if tt.request != nil {
				jsonData, err = json.Marshal(tt.request)
				require.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/v1/multimodal/process", bytes.NewBuffer(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// 基准测试
func BenchmarkMultimodalProcess(b *testing.B) {
	router, db, _ := setupMultimodalTest(&testing.T{})

	// 创建测试会话
	session := &models.MultimodalSession{
		ID:        "bench-session",
		UserID:    "bench-user",
		Title:     "Benchmark Session",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(session)

	request := models.MultimodalRequest{
		SessionID: "bench-session",
		Inputs: []models.MultimodalInput{
			{
				Type: "text",
				Content: models.MultimodalContent{
					Text: "Benchmark test message",
				},
			},
		},
		Config: models.MultimodalConfig{
			Provider:    "openai",
			Model:       "gpt-4",
			MaxTokens:   100,
			Temperature: 0.7,
		},
	}

	jsonData, _ := json.Marshal(request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/multimodal/process", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
