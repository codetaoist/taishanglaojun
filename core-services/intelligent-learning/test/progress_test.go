﻿package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/interfaces/http/handlers"
)

// MockProgressService 
type MockProgressService struct {
	mock.Mock
}

func (m *MockProgressService) UpdateProgress(req *analytics.ProgressUpdateRequest) (*analytics.ProgressResponse, error) {
	args := m.Called(req)
	return args.Get(0).(*analytics.ProgressResponse), args.Error(1)
}

func (m *MockProgressService) GetLearningReport(learnerID string, period analytics.ReportPeriod) (*analytics.LearningReport, error) {
	args := m.Called(learnerID, period)
	return args.Get(0).(*analytics.LearningReport), args.Error(1)
}

func TestProgressHandler_UpdateProgress(t *testing.T) {
	// Gin?	gin.SetMode(gin.TestMode)

	// 
	mockService := new(MockProgressService)
	handler := handlers.NewProgressHandler(mockService)

	// 
	router := gin.New()
	router.POST("/progress/update", handler.UpdateProgress)

	// 
	updateReq := analytics.ProgressUpdateRequest{
		LearnerID: "learner-123",
		ContentID: "content-456",
		SessionID: "session-789",
		Progress: analytics.ContentProgress{
			CompletionPercentage: 75.5,
			TimeSpent:           1800, // 30
			LastAccessedAt:      time.Now(),
			Status:              "in_progress",
		},
		InteractionType: "video_watch",
		Timestamp:       time.Now(),
	}

	expectedResponse := &analytics.ProgressResponse{
		Success: true,
		Message: "",
		Data: map[string]interface{}{
			"completion_percentage": 75.5,
			"time_spent":           1800,
			"level_up":             false,
		},
		NextSteps: []analytics.NextStepRecommendation{
			{
				Type:        "continue_content",
				ContentID:   "content-456",
				Title:       "",
				Description: "75.5%?,
				Priority:    "high",
			},
		},
		Achievements: []analytics.Achievement{},
	}

	// 
	mockService.On("UpdateProgress", mock.AnythingOfType("*analytics.ProgressUpdateRequest")).Return(expectedResponse, nil)

	// 
	reqBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("POST", "/progress/update", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 
	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.ProgressResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "", response.Message)
	assert.Equal(t, 75.5, response.Data["completion_percentage"])

	// 
	mockService.AssertExpectations(t)
}

func TestProgressHandler_GetLearningReport(t *testing.T) {
	// Gin?	gin.SetMode(gin.TestMode)

	// 
	mockService := new(MockProgressService)
	handler := handlers.NewProgressHandler(mockService)

	// 
	router := gin.New()
	router.GET("/progress/report/:learnerId", handler.GetLearningReport)

	// 
	expectedReport := &analytics.LearningReport{
		LearnerID:     "learner-123",
		Period:        analytics.ReportPeriodWeek,
		GeneratedAt:   time.Now(),
		OverallProgress: analytics.OverallProgress{
			TotalTimeSpent:        7200, // 2
			CompletedContents:     5,
			InProgressContents:    3,
			AverageCompletionRate: 68.5,
			StreakDays:           7,
		},
		ContentProgress: []analytics.ContentProgressSummary{
			{
				ContentID:            "content-456",
				Title:               "Go",
				CompletionPercentage: 75.5,
				TimeSpent:           1800,
				LastAccessed:        time.Now().Add(-2 * time.Hour),
				Status:              "in_progress",
			},
		},
		SkillProgress: []analytics.SkillProgress{
			{
				SkillName:     "Go",
				CurrentLevel:  3,
				Experience:    750,
				NextLevelExp:  1000,
				Improvement:   25.5,
			},
		},
		LearningPatterns: analytics.LearningPatternAnalysis{
			PreferredTimeSlots: []analytics.TimeSlotAnalysis{
				{
					TimeSlot:    "morning",
					Frequency:   5,
					AvgDuration: 45,
					Efficiency:  85.2,
				},
			},
			EngagementPatterns: []analytics.EngagementPattern{
				{
					ContentType:   "video",
					AvgEngagement: 78.5,
					CompletionRate: 82.3,
				},
			},
		},
		Recommendations: []analytics.RecommendationItem{
			{
				Type:        "content",
				ContentID:   "content-789",
				Title:       "Go?,
				Reason:      "Go",
				Priority:    "high",
				EstimatedTime: 60,
			},
		},
	}

	// 
	mockService.On("GetLearningReport", "learner-123", analytics.ReportPeriodWeek).Return(expectedReport, nil)

	// 
	req, _ := http.NewRequest("GET", "/progress/report/learner-123?period=week", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 
	assert.Equal(t, http.StatusOK, w.Code)

	var response analytics.LearningReport
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "learner-123", response.LearnerID)
	assert.Equal(t, analytics.ReportPeriodWeek, response.Period)
	assert.Equal(t, 7200, response.OverallProgress.TotalTimeSpent)
	assert.Equal(t, 5, response.OverallProgress.CompletedContents)
	assert.Len(t, response.ContentProgress, 1)
	assert.Len(t, response.SkillProgress, 1)
	assert.Len(t, response.Recommendations, 1)

	// 
	mockService.AssertExpectations(t)
}

func TestProgressHandler_ValidationErrors(t *testing.T) {
	// Gin?	gin.SetMode(gin.TestMode)

	// 
	mockService := new(MockProgressService)
	handler := handlers.NewProgressHandler(mockService)

	// 
	router := gin.New()
	router.POST("/progress/update", handler.UpdateProgress)

	// JSON
	req, _ := http.NewRequest("POST", "/progress/update", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 
	invalidReq := map[string]interface{}{
		"learner_id": "", // learner_id
		"content_id": "content-456",
	}

	reqBody, _ := json.Marshal(invalidReq)
	req, _ = http.NewRequest("POST", "/progress/update", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

