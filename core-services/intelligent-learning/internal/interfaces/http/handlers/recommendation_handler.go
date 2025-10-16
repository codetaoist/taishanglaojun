package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RecommendationHandler ?
type RecommendationHandler struct {
	personalizationEngine *domainServices.PersonalizationEngine
	userBehaviorTracker   *domainServices.UserBehaviorTracker
	preferenceAnalyzer    *domainServices.PreferenceAnalyzer
	contextAnalyzer       *domainServices.ContextAnalyzer
}

// NewRecommendationHandler ?
func NewRecommendationHandler(
	personalizationEngine *domainServices.PersonalizationEngine,
	userBehaviorTracker *domainServices.UserBehaviorTracker,
	preferenceAnalyzer *domainServices.PreferenceAnalyzer,
	contextAnalyzer *domainServices.ContextAnalyzer,
) *RecommendationHandler {
	return &RecommendationHandler{
		personalizationEngine: personalizationEngine,
		userBehaviorTracker:   userBehaviorTracker,
		preferenceAnalyzer:    preferenceAnalyzer,
		contextAnalyzer:       contextAnalyzer,
	}
}

// PersonalizedRecommendationRequest 
type PersonalizedRecommendationRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// RecommendationResponse 
type RecommendationResponse struct {
	Recommendations []PersonalizedRecommendation `json:"recommendations"`
	Metadata        RecommendationMetadata       `json:"metadata"`
}

// PersonalizedRecommendation ?
type PersonalizedRecommendation struct {
	ContentID    string                 `json:"content_id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	ContentType  string                 `json:"content_type"`
	Score        float64                `json:"score"`
	Confidence   float64                `json:"confidence"`
	Reasoning    string                 `json:"reasoning"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	EstimatedTime int                   `json:"estimated_time"`
	Difficulty   string                 `json:"difficulty"`
}

// RecommendationMetadata ?
type RecommendationMetadata struct {
	Strategy      string                 `json:"strategy"`
	Timestamp     time.Time              `json:"timestamp"`
	UserProfile   map[string]interface{} `json:"user_profile"`
	Context       map[string]interface{} `json:"context"`
	TotalCount    int                    `json:"total_count"`
	ProcessingTime int64                 `json:"processing_time_ms"`
}

// GetPersonalizedRecommendations 
// @Summary 
// @Description 
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body PersonalizedRecommendationRequest true ""
// @Success 200 {object} RecommendationResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/personalized [post]
func (h *RecommendationHandler) GetPersonalizedRecommendations(c *gin.Context) {
	var req PersonalizedRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: ": " + err.Error(),
		})
		return
	}

	// ?
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	startTime := time.Now()

	// 
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "ID",
		})
		return
	}

	personalizationReq := &domainServices.PersonalizationRequest{
		LearnerID:           userID,
		RecommendationType:  req.ContentType,
		MaxRecommendations:  req.Limit,
		IncludeExplanations: true,
		Filters:             req.Preferences,
		PersonalizationLevel: "advanced",
	}

	// 
	response, err := h.personalizationEngine.GeneratePersonalizedRecommendations(c.Request.Context(), personalizationReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	recommendations := make([]PersonalizedRecommendation, len(response.Recommendations))
	for i, rec := range response.Recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}
		
		reasoning := ""
		if len(rec.Reasoning) > 0 {
			reasoning = strings.Join(rec.Reasoning, "; ")
		}
		
		estimatedTime := int(rec.EstimatedTime.Minutes())
		
		recommendations[i] = PersonalizedRecommendation{
			ContentID:     contentID,
			Title:         rec.Title,
			Description:   rec.Description,
			ContentType:   rec.Type,
			Score:         rec.Score,
			Confidence:    rec.Confidence,
			Reasoning:     reasoning,
			Tags:          rec.Tags,
			Metadata:      rec.Metadata,
			EstimatedTime: estimatedTime,
			Difficulty:    rec.Difficulty,
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	result := RecommendationResponse{
		Recommendations: recommendations,
		Metadata: RecommendationMetadata{
			Strategy:       "personalized",
			Timestamp:      response.GeneratedAt,
			UserProfile:    make(map[string]interface{}),
			Context:        make(map[string]interface{}),
			TotalCount:     len(recommendations),
			ProcessingTime: processingTime,
		},
	}

	c.JSON(http.StatusOK, result)
}

// GetRecommendationsByStrategy 
// @Summary 
// @Description 
// @Tags recommendations
// @Accept json
// @Produce json
// @Param strategy path string true "" Enums(collaborative,content_based,hybrid,popular,trending)
// @Param user_id query string true "ID"
// @Param limit query int false "" default(10)
// @Success 200 {object} RecommendationResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/strategy/{strategy} [get]
func (h *RecommendationHandler) GetRecommendationsByStrategy(c *gin.Context) {
	strategy := c.Param("strategy")
	userIDStr := c.Query("user_id")
	limitStr := c.DefaultQuery("limit", "10")

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_parameter",
			Message: "ID",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	startTime := time.Now()

	// 
	req := &domainServices.PersonalizationRequest{
		LearnerID:           userID,
		MaxRecommendations:  limit,
		PersonalizationLevel: strategy,
		IncludeExplanations: true,
	}

	// 
	response, err := h.personalizationEngine.GeneratePersonalizedRecommendations(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	recommendations := make([]PersonalizedRecommendation, len(response.Recommendations))
	for i, rec := range response.Recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}
		
		reasoning := ""
		if len(rec.Reasoning) > 0 {
			reasoning = strings.Join(rec.Reasoning, "; ")
		}
		
		estimatedTime := int(rec.EstimatedTime.Minutes())
		
		recommendations[i] = PersonalizedRecommendation{
			ContentID:     contentID,
			Title:         rec.Title,
			Description:   rec.Description,
			ContentType:   rec.Type,
			Score:         rec.Score,
			Confidence:    rec.Confidence,
			Reasoning:     reasoning,
			Tags:          rec.Tags,
			Metadata:      rec.Metadata,
			EstimatedTime: estimatedTime,
			Difficulty:    rec.Difficulty,
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	result := RecommendationResponse{
		Recommendations: recommendations,
		Metadata: RecommendationMetadata{
			Strategy:       "personalized",
			Timestamp:      response.GeneratedAt,
			UserProfile:    make(map[string]interface{}),
			Context:        make(map[string]interface{}),
			TotalCount:     len(recommendations),
			ProcessingTime: processingTime,
		},
	}

	c.JSON(http.StatusOK, result)
}

// RecordUserBehavior 
// @Summary 
// @Description ?
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body domainServices.BehaviorEvent true ""
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/behavior [post]
func (h *RecommendationHandler) RecordUserBehavior(c *gin.Context) {
	var event domainServices.BehaviorEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: ": " + err.Error(),
		})
		return
	}

	// ?
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 
	err := h.userBehaviorTracker.TrackBehaviorEvent(c.Request.Context(), &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "tracking_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "",
		"timestamp": event.Timestamp,
	})
}

// GetUserPreferences 
// @Summary 
// @Description 
// @Tags recommendations
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} services.UserPreferences ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/preferences/{user_id} [get]
func (h *RecommendationHandler) GetUserPreferences(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	preferences, err := h.preferenceAnalyzer.AnalyzeUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// GetLearningContext ?
// @Summary ?
// @Description 龳
// @Tags recommendations
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} services.LearningContext "?
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/context/{user_id} [get]
func (h *RecommendationHandler) GetLearningContext(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	context, err := h.contextAnalyzer.AnalyzeLearningContext(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: "? " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, context)
}

// GetBehaviorInsights 
// @Summary 
// @Description 
// @Tags recommendations
// @Produce json
// @Param user_id path string true "ID"
// @Param days query int false "" default(30)
// @Success 200 {object} services.LearningInsights ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/insights/{user_id} [get]
func (h *RecommendationHandler) GetBehaviorInsights(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	// ID
	learnerID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "ID",
		})
		return
	}

	// 
	timeRange := domainServices.BehaviorTimeRange{
		Start: time.Now().AddDate(0, 0, -days),
		End:   time.Now(),
	}

	insights, err := h.userBehaviorTracker.GetLearningInsights(c.Request.Context(), learnerID, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "analysis_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// BatchRecommendations 
// @Summary 
// @Description ?
// @Tags recommendations
// @Accept json
// @Produce json
// @Param request body BatchRecommendationRequest true ""
// @Success 200 {object} BatchRecommendationResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/recommendations/batch [post]
func (h *RecommendationHandler) BatchRecommendations(c *gin.Context) {
	var req BatchRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: ": " + err.Error(),
		})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	if len(req.UserIDs) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_users",
			Message: "100?,
		})
		return
	}

	startTime := time.Now()
	results := make(map[string]RecommendationResponse)
	errors := make(map[string]string)

	// ?
	for _, userID := range req.UserIDs {
		// ID
		learnerID, err := uuid.Parse(userID)
		if err != nil {
			errors[userID] = "ID"
			continue
		}

		personalizationReq := &domainServices.PersonalizationRequest{
			LearnerID:            learnerID,
			RecommendationType:   req.ContentType,
			MaxRecommendations:   req.Limit,
			IncludeExplanations:  true,
			Filters:              make(map[string]interface{}),
			PersonalizationLevel: "advanced",
		}

		response, err := h.personalizationEngine.GeneratePersonalizedRecommendations(c.Request.Context(), personalizationReq)
		if err != nil {
			errors[userID] = err.Error()
			continue
		}

		// 
		recommendations := make([]PersonalizedRecommendation, len(response.Recommendations))
		for i, rec := range response.Recommendations {
			contentID := ""
			if rec.ContentID != nil {
				contentID = rec.ContentID.String()
			}
			
			reasoning := ""
			if len(rec.Reasoning) > 0 {
				reasoning = strings.Join(rec.Reasoning, "; ")
			}
			
			estimatedTime := int(rec.EstimatedTime.Minutes())
			
			recommendations[i] = PersonalizedRecommendation{
				ContentID:     contentID,
				Title:         rec.Title,
				Description:   rec.Description,
				ContentType:   rec.Type,
				Score:         rec.Score,
				Confidence:    rec.Confidence,
				Reasoning:     reasoning,
				Tags:          rec.Tags,
				Metadata:      rec.Metadata,
				EstimatedTime: estimatedTime,
				Difficulty:    rec.Difficulty,
			}
		}

		results[userID] = RecommendationResponse{
			Recommendations: recommendations,
			Metadata: RecommendationMetadata{
				Strategy:    "personalized",
				Timestamp:   response.GeneratedAt,
				UserProfile: make(map[string]interface{}),
				Context:     make(map[string]interface{}),
				TotalCount:  len(recommendations),
			},
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	batchResponse := BatchRecommendationResponse{
		Results:        results,
		Errors:         errors,
		TotalUsers:     len(req.UserIDs),
		SuccessCount:   len(results),
		ErrorCount:     len(errors),
		ProcessingTime: processingTime,
	}

	c.JSON(http.StatusOK, batchResponse)
}

// BatchRecommendationRequest 
type BatchRecommendationRequest struct {
	UserIDs     []string               `json:"user_ids" binding:"required"`
	ContentType string                 `json:"content_type,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// BatchRecommendationResponse 
type BatchRecommendationResponse struct {
	Results        map[string]RecommendationResponse `json:"results"`
	Errors         map[string]string                 `json:"errors"`
	TotalUsers     int                               `json:"total_users"`
	SuccessCount   int                               `json:"success_count"`
	ErrorCount     int                               `json:"error_count"`
	ProcessingTime int64                             `json:"processing_time_ms"`
}

// ErrorResponse 

