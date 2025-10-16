﻿package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RealtimeAnalyticsHandler ?
type RealtimeAnalyticsHandler struct {
	realtimeService *services.RealtimeAnalyticsService
	upgrader        websocket.Upgrader
}

// NewRealtimeAnalyticsHandler ?
func NewRealtimeAnalyticsHandler(realtimeService *services.RealtimeAnalyticsService) *RealtimeAnalyticsHandler {
	return &RealtimeAnalyticsHandler{
		realtimeService: realtimeService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // ?
			},
		},
	}
}

// ProcessEventRequest 
type ProcessEventRequest struct {
	Type      string                 `json:"type" binding:"required"`
	LearnerID uuid.UUID              `json:"learner_id" binding:"required"`
	ContentID uuid.UUID              `json:"content_id,omitempty"`
	SessionID uuid.UUID              `json:"session_id" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CreateAnalyzerRequest ?
type CreateAnalyzerRequest struct {
	Type      string                 `json:"type" binding:"required"`
	LearnerID uuid.UUID              `json:"learner_id" binding:"required"`
	Config    map[string]interface{} `json:"config"`
}

// ProcessEvent 
func (h *RealtimeAnalyticsHandler) ProcessEvent(c *gin.Context) {
	var req ProcessEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := &services.LearningEvent{
		ID:        uuid.New(),
		Type:      req.Type,
		LearnerID: req.LearnerID,
		ContentID: req.ContentID,
		SessionID: req.SessionID,
		Timestamp: time.Now(),
		Data:      req.Data,
		Metadata:  req.Metadata,
	}

	if err := h.realtimeService.ProcessEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Event processed successfully",
		"event_id": event.ID,
	})
}

// GetRealtimeData 
func (h *RealtimeAnalyticsHandler) GetRealtimeData(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

// CreateAnalyzer ?
func (h *RealtimeAnalyticsHandler) CreateAnalyzer(c *gin.Context) {
	var req CreateAnalyzerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	analyzer, err := h.realtimeService.CreateAnalyzer(req.Type, req.LearnerID, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"analyzer": analyzer,
	})
}

// SubscribeToUpdates  (WebSocket)
func (h *RealtimeAnalyticsHandler) SubscribeToUpdates(c *gin.Context) {
	subscriberID := c.Query("subscriber_id")
	if subscriberID == "" {
		subscriberID = uuid.New().String()
	}

	updateChan, err := h.realtimeService.Subscribe(subscriberID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()
	defer h.realtimeService.Unsubscribe(subscriberID)

	// WebSocket
	for {
		select {
		case update, ok := <-updateChan:
			if !ok {
				return
			}
			
			if err := conn.WriteJSON(update); err != nil {
				return
			}
		}
	}
}

// GetAnalyticsMetrics 
func (h *RealtimeAnalyticsHandler) GetAnalyticsMetrics(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 
	metrics := gin.H{
		"learning_velocity":    data.RealtimeMetrics.LearningVelocity,
		"engagement_score":     data.EngagementState.Score,
		"performance_score":    data.PerformanceState.Score,
		"focus_score":          data.RealtimeMetrics.FocusScore,
		"efficiency_score":     data.RealtimeMetrics.EfficiencyScore,
		"motivation_level":     data.RealtimeMetrics.MotivationLevel,
		"cognitive_load":       data.RealtimeMetrics.CognitiveLoad,
		"engagement_trend":     data.RealtimeMetrics.EngagementTrend,
		"completion_probability": data.PredictiveInsights.CompletionProbability,
		"dropout_risk":         data.PredictiveInsights.RiskOfDropout,
		"last_updated":         data.RealtimeMetrics.LastUpdated,
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetLearningInsights 
func (h *RealtimeAnalyticsHandler) GetLearningInsights(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	insights := gin.H{
		"behavior_patterns": gin.H{
			"learning_rhythm":       data.BehaviorPatterns.LearningRhythm,
			"attention_span":        data.BehaviorPatterns.AttentionSpan,
			"interaction_frequency": data.BehaviorPatterns.InteractionFrequency,
			"preferred_content_types": data.BehaviorPatterns.PreferredContentTypes,
		},
		"engagement_state": gin.H{
			"level":               data.EngagementState.Level,
			"score":               data.EngagementState.Score,
			"trend":               data.EngagementState.Trend,
			"interaction_quality": data.EngagementState.InteractionQuality,
			"risk_factors":        data.EngagementState.RiskFactors,
		},
		"performance_state": gin.H{
			"current_level":        data.PerformanceState.CurrentLevel,
			"score":                data.PerformanceState.Score,
			"trend":                data.PerformanceState.Trend,
			"strength_areas":       data.PerformanceState.StrengthAreas,
			"improvement_areas":    data.PerformanceState.ImprovementAreas,
			"recent_achievements":  data.PerformanceState.RecentAchievements,
		},
		"predictive_insights": gin.H{
			"completion_probability":   data.PredictiveInsights.CompletionProbability,
			"estimated_completion_time": data.PredictiveInsights.EstimatedCompletionTime,
			"risk_of_dropout":          data.PredictiveInsights.RiskOfDropout,
			"recommended_actions":      data.PredictiveInsights.RecommendedActions,
			"predicted_challenges":     data.PredictiveInsights.PredictedChallenges,
		},
		"alerts": data.Alerts,
	}

	c.JSON(http.StatusOK, gin.H{
		"insights": insights,
	})
}

// GetSessionAnalytics 
func (h *RealtimeAnalyticsHandler) GetSessionAnalytics(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if data.CurrentSession == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active session found"})
		return
	}

	session := gin.H{
		"session_id":        data.CurrentSession.ID,
		"start_time":        data.CurrentSession.StartTime,
		"duration":          data.CurrentSession.Duration,
		"interaction_count": data.CurrentSession.InteractionCount,
		"progress_made":     data.CurrentSession.ProgressMade,
		"engagement_score":  data.CurrentSession.EngagementScore,
		"focus_level":       data.CurrentSession.FocusLevel,
		"content_items":     data.CurrentSession.ContentItems,
		"last_activity":     data.CurrentSession.LastActivity,
	}

	c.JSON(http.StatusOK, gin.H{
		"session": session,
	})
}

// GetPerformanceTrends 
func (h *RealtimeAnalyticsHandler) GetPerformanceTrends(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	// 
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hours parameter"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// ?
	trends := gin.H{
		"time_range": gin.H{
			"start": time.Now().Add(-time.Duration(hours) * time.Hour),
			"end":   time.Now(),
		},
		"performance_trend": data.PerformanceState.Trend,
		"engagement_trend":  data.RealtimeMetrics.EngagementTrend,
		"current_metrics": gin.H{
			"performance_score": data.PerformanceState.Score,
			"engagement_score":  data.EngagementState.Score,
			"learning_velocity": data.RealtimeMetrics.LearningVelocity,
			"efficiency_score":  data.RealtimeMetrics.EfficiencyScore,
		},
		"predictions": gin.H{
			"completion_probability": data.PredictiveInsights.CompletionProbability,
			"dropout_risk":          data.PredictiveInsights.RiskOfDropout,
			"estimated_completion":  data.PredictiveInsights.EstimatedCompletionTime,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
	})
}

// GetAlerts 
func (h *RealtimeAnalyticsHandler) GetAlerts(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 龯?
	alertsByLevel := make(map[string][]services.Alert)
	for _, alert := range data.Alerts {
		alertsByLevel[alert.Level] = append(alertsByLevel[alert.Level], alert)
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": gin.H{
			"total":     len(data.Alerts),
			"by_level":  alertsByLevel,
			"all_alerts": data.Alerts,
		},
	})
}

// GetRecommendations 
func (h *RealtimeAnalyticsHandler) GetRecommendations(c *gin.Context) {
	learnerIDStr := c.Param("learnerId")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid learner ID"})
		return
	}

	data, err := h.realtimeService.GetRealtimeData(learnerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	recommendations := gin.H{
		"immediate_actions":    data.PredictiveInsights.RecommendedActions,
		"next_best_content":    data.PredictiveInsights.NextBestContent,
		"optimal_study_time":   data.PredictiveInsights.OptimalStudyTime,
		"predicted_challenges": data.PredictiveInsights.PredictedChallenges,
		"success_factors":      data.PredictiveInsights.SuccessFactors,
		"learning_strategy": gin.H{
			"based_on_rhythm":     data.BehaviorPatterns.LearningRhythm,
			"attention_span":      data.BehaviorPatterns.AttentionSpan,
			"preferred_content":   data.BehaviorPatterns.PreferredContentTypes,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
	})
}

