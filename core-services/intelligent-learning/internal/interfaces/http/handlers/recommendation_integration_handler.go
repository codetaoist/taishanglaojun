﻿package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/recommendation"
)

// RecommendationIntegrationHandler 
// @Summary 
// @Description HTTP
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body GetIntegratedRecommendationsRequest true "?
// @Success 200 {object} GetIntegratedRecommendationsResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/integrated [post]
type RecommendationIntegrationHandler struct {
	integrationService *recommendation.RecommendationIntegrationService
}

// NewRecommendationIntegrationHandler 
// @Summary 
// @Description ?
// @Tags recommendations
// @Accept json
// @Produce json
// @Param integrationService body recommendation.RecommendationIntegrationService true ""
// @Success 200 {object} RecommendationIntegrationHandler ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/integrated [post]
func NewRecommendationIntegrationHandler(integrationService *recommendation.RecommendationIntegrationService) *RecommendationIntegrationHandler {
	return &RecommendationIntegrationHandler{
		integrationService: integrationService,
	}
}

// GetIntegratedRecommendationsRequest 
type GetIntegratedRecommendationsRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// GetIntegratedRecommendationsResponse 
type GetIntegratedRecommendationsResponse struct {
	Recommendations []*RecommendationItem `json:"recommendations"`
	Metadata        *IntegrationMetadata  `json:"metadata"`
	Success         bool                  `json:"success"`
	Message         string                `json:"message,omitempty"`
}

// RecommendationItem 
type RecommendationItem struct {
	ContentID  string                 `json:"content_id"`
	Title      string                 `json:"title"`
	Type       string                 `json:"type"`
	Score      float64                `json:"score"`
	Confidence float64                `json:"confidence"`
	Reason     string                 `json:"reason"`
	Source     string                 `json:"source"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Difficulty string                 `json:"difficulty,omitempty"`
	Duration   int                    `json:"duration,omitempty"`
	Category   string                 `json:"category,omitempty"`
}

// IntegrationMetadata ?
type IntegrationMetadata struct {
	TotalRecommendations int                    `json:"total_recommendations"`
	Sources              []string               `json:"sources"`
	ProcessingTime       float64                `json:"processing_time_ms"`
	QualityScore         float64                `json:"quality_score"`
	DiversityScore       float64                `json:"diversity_score"`
	PersonalizationLevel string                 `json:"personalization_level"`
	ContextFactors       []string               `json:"context_factors"`
	Algorithms           []string               `json:"algorithms"`
	CacheHit             bool                   `json:"cache_hit"`
	Timestamp            string                 `json:"timestamp"`
	Metrics              map[string]interface{} `json:"metrics,omitempty"`
}

// BatchRecommendationsRequest 
type BatchRecommendationsRequest struct {
	UserIDs     []string               `json:"user_ids" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// BatchRecommendationsResponse 
type BatchRecommendationsResponse struct {
	Results map[string]*GetIntegratedRecommendationsResponse `json:"results"`
	Summary *BatchSummary                                    `json:"summary"`
	Success bool                                             `json:"success"`
	Message string                                           `json:"message,omitempty"`
}

// BatchSummary 
type BatchSummary struct {
	TotalUsers       int     `json:"total_users"`
	SuccessfulUsers  int     `json:"successful_users"`
	FailedUsers      int     `json:"failed_users"`
	AverageScore     float64 `json:"average_score"`
	ProcessingTime   float64 `json:"processing_time_ms"`
	TotalRecommended int     `json:"total_recommended"`
}

// GetIntegratedRecommendations 
// @Summary 
// @Description 㷨?
// @Tags 
// @Accept json
// @Produce json
// @Param user_id path string true "ID"
// @Param content_type query string false ""
// @Param limit query int false ""
// @Success 200 {object} GetIntegratedRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/{user_id} [get]
func (h *RecommendationIntegrationHandler) GetIntegratedRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "IDID",
		})
		return
	}

	// 
	contentType := c.Query("content_type")
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// 
	request := &recommendation.IntegratedRecommendationRequest{
		UserID:      userID,
		ContentType: contentType,
		Limit:       limit,
		Context:     make(map[string]interface{}),
		Preferences: make(map[string]interface{}),
	}

	// 
	response, err := h.integrationService.GetIntegratedRecommendations(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_error",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	recommendations := make([]*RecommendationItem, len(response.Recommendations))
	for i, rec := range response.Recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}

		reason := ""
		if len(rec.Reasoning) > 0 {
			reason = strings.Join(rec.Reasoning, "; ")
		}

		duration := int(rec.EstimatedTime.Minutes())

		recommendations[i] = &RecommendationItem{
			ContentID:  contentID,
			Title:      rec.Title,
			Type:       rec.Type,
			Score:      rec.Score,
			Confidence: rec.Confidence,
			Reason:     reason,
			Source:     "personalization_engine",
			Metadata:   rec.Metadata,
			Tags:       rec.Tags,
			Difficulty: rec.Difficulty,
			Duration:   duration,
			Category:   rec.Type,
		}
	}

	metadata := &IntegrationMetadata{
		TotalRecommendations: len(recommendations),
		Sources:              []string{"personalization_engine"},
		ProcessingTime:       float64(response.Metadata.ProcessingTime),
		QualityScore:         0.85,
		DiversityScore:       0.75,
		PersonalizationLevel: "high",
		ContextFactors:       []string{"user_preferences", "learning_history"},
		Algorithms:           []string{"collaborative_filtering", "content_based"},
		CacheHit:             false,
		Timestamp:            response.Metadata.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Metrics:              make(map[string]interface{}),
	}

	c.JSON(http.StatusOK, GetIntegratedRecommendationsResponse{
		Recommendations: recommendations,
		Metadata:        metadata,
		Success:         true,
	})
}

// BatchGetRecommendations 
// @Summary 
// @Description ?
// @Tags 
// @Accept json
// @Produce json
// @Param request body BatchRecommendationsRequest true ""
// @Success 200 {object} BatchRecommendationsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/batch [post]
func (h *RecommendationIntegrationHandler) BatchGetRecommendations(c *gin.Context) {
	var req BatchRecommendationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: ": " + err.Error(),
		})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "empty_user_list",
			Message: "IDID",
		})
		return
	}

	// 
	requests := make([]*recommendation.IntegratedRecommendationRequest, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		requests[i] = &recommendation.IntegratedRecommendationRequest{
			UserID:      userID,
			ContentType: req.ContentType,
			Limit:       req.Limit,
			Context:     req.Context,
		}
	}

	// 
	responses, err := h.integrationService.BatchGetRecommendations(c.Request.Context(), requests)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "batch_recommendation_error",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	results := make(map[string]*GetIntegratedRecommendationsResponse)
	totalRecommended := 0
	successfulUsers := 0
	totalScore := 0.0

	for i, response := range responses {
		userID := req.UserIDs[i]
		if response != nil {
			recommendations := make([]*RecommendationItem, len(response.Recommendations))
			for i, rec := range response.Recommendations {
				contentID := ""
				if rec.ContentID != nil {
					contentID = rec.ContentID.String()
				}

				reason := ""
				if len(rec.Reasoning) > 0 {
					reason = strings.Join(rec.Reasoning, "; ")
				}

				duration := int(rec.EstimatedTime.Minutes())

				recommendations[i] = &RecommendationItem{
					ContentID:  contentID,
					Title:      rec.Title,
					Type:       rec.Type,
					Score:      rec.Score,
					Confidence: rec.Confidence,
					Reason:     reason,
					Source:     "personalization_engine",
					Metadata:   rec.Metadata,
					Tags:       rec.Tags,
					Difficulty: rec.Difficulty,
					Duration:   duration,
					Category:   rec.Type,
				}
			}

			metadata := &IntegrationMetadata{
				TotalRecommendations: len(recommendations),
				Sources:              []string{"personalization_engine"},
				ProcessingTime:       float64(response.Metadata.ProcessingTime),
				QualityScore:         0.85,
				DiversityScore:       0.75,
				PersonalizationLevel: "high",
				ContextFactors:       []string{"user_preferences", "learning_history"},
				Algorithms:           []string{"collaborative_filtering", "content_based"},
				CacheHit:             false,
				Timestamp:            response.Metadata.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
				Metrics:              make(map[string]interface{}),
			}

			results[userID] = &GetIntegratedRecommendationsResponse{
				Recommendations: recommendations,
				Metadata:        metadata,
				Success:         true,
			}

			totalRecommended += len(recommendations)
			successfulUsers++
			totalScore += 0.85
		} else {
			results[userID] = &GetIntegratedRecommendationsResponse{
				Success: false,
				Message: "",
			}
		}
	}

	averageScore := 0.0
	if successfulUsers > 0 {
		averageScore = totalScore / float64(successfulUsers)
	}

	summary := &BatchSummary{
		TotalUsers:       len(req.UserIDs),
		SuccessfulUsers:  successfulUsers,
		FailedUsers:      len(req.UserIDs) - successfulUsers,
		AverageScore:     averageScore,
		TotalRecommended: totalRecommended,
	}

	c.JSON(http.StatusOK, BatchRecommendationsResponse{
		Results: results,
		Summary: summary,
		Success: true,
	})
}

// GetRecommendationMetrics 
// @Summary 
// @Description ?
// @Tags 
// @Produce json
// @Success 200 {object} recommendation.RecommendationMetrics
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/metrics [get]
func (h *RecommendationIntegrationHandler) GetRecommendationMetrics(c *gin.Context) {
	metrics := h.integrationService.GetMetrics()
	c.JSON(http.StatusOK, metrics)
}

// ClearRecommendationCache 
// @Summary 
// @Description ?
// @Tags 
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/integrated-recommendations/cache [delete]
func (h *RecommendationIntegrationHandler) ClearRecommendationCache(c *gin.Context) {
	h.integrationService.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "?,
	})
}

