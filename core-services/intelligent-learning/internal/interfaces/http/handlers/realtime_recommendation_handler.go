package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/recommendation"
)

// RealtimeRecommendationHandler 
// @Summary 
// @Description HTTP
// @Tags realtime-recommendations
// @Accept json
// @Produce json
// @Param request body recommendation.RealtimeEvent true ""
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/events [post]
type RealtimeRecommendationHandler struct {
	realtimeService *recommendation.RealtimeRecommendationService
	upgrader        websocket.Upgrader
}

// NewRealtimeRecommendationHandler 
// @Summary 
// @Description 
// @Tags realtime-recommendations
// @Param realtimeService body recommendation.RealtimeRecommendationService true ""
// @Success 200 {object} RealtimeRecommendationHandler ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/events [post]
func NewRealtimeRecommendationHandler(realtimeService *recommendation.RealtimeRecommendationService) *RealtimeRecommendationHandler {
	return &RealtimeRecommendationHandler{
		realtimeService: realtimeService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // ?
			},
		},
	}
}

// ProcessRealtimeEvent 
// @Summary 
// @Description ?
// @Tags realtime-recommendations
// @Accept json
// @Produce json
// @Param request body recommendation.RealtimeEvent true ""
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/events [post]
func (h *RealtimeRecommendationHandler) ProcessRealtimeEvent(c *gin.Context) {
	var event recommendation.RealtimeEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	err := h.realtimeService.ProcessEvent(c.Request.Context(), &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "event_processing_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "",
		"event_id":  event.EventID,
		"timestamp": event.Timestamp,
	})
}

// GetRealtimeRecommendations 
// @Summary 
// @Description ?
// @Tags realtime-recommendations
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} RealtimeRecommendationResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/{user_id} [get]
func (h *RealtimeRecommendationHandler) GetRealtimeRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	startTime := time.Now()

	// 
	recommendations, err := h.realtimeService.GetRealtimeRecommendations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: ": " + err.Error(),
		})
		return
	}

	// 
	session, _ := h.realtimeService.GetUserSession(userID)

	processingTime := time.Since(startTime).Milliseconds()

	response := RealtimeRecommendationResponse{
		UserID:          userID,
		Recommendations: convertToPersonalizedRecommendations(recommendations),
		Session:         convertToSessionInfo(session),
		Metadata: RealtimeMetadata{
			Strategy:       "realtime",
			Timestamp:      time.Now(),
			ProcessingTime: processingTime,
			IsRealtime:     true,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetUserSession 
// @Summary 
// @Description ?
// @Tags realtime-recommendations
// @Produce json
// @Param user_id path string true "ID"
// @Success 200 {object} recommendation.RealtimeUserSession ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/sessions/{user_id} [get]
func (h *RealtimeRecommendationHandler) GetUserSession(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	session, err := h.realtimeService.GetUserSession(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "session_not_found",
			Message: "? " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// SubscribeToRecommendationUpdates WebSocket
// @Summary WebSocket
// @Description WebSocket
// @Tags realtime-recommendations
// @Param user_id query string true "ID"
// @Success 101 {string} string "WebSocket"
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/subscribe [get]
func (h *RealtimeRecommendationHandler) SubscribeToRecommendationUpdates(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "ID",
		})
		return
	}

	// WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "websocket_upgrade_failed",
			Message: "WebSocket: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	// 
	updateChannel := h.realtimeService.SubscribeToUpdates(userID)
	defer h.realtimeService.UnsubscribeFromUpdates(userID)

	// ?
	recommendations, err := h.realtimeService.GetRealtimeRecommendations(c.Request.Context(), userID)
	if err == nil {
		initialUpdate := &recommendation.RecommendationUpdate{
			UserID:          userID,
			UpdateType:      "initial",
			Recommendations: recommendations,
			Reason:          "Initial recommendations",
			Timestamp:       time.Now(),
		}
		conn.WriteJSON(initialUpdate)
	}

	// WebSocket
	go func() {
		for {
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				break
			}
			// ?
			if message["type"] == "heartbeat" {
				continue
			}
		}
	}()

	// 
	for {
		select {
		case update, ok := <-updateChannel:
			if !ok {
				return
			}

			err := conn.WriteJSON(update)
			if err != nil {
				return
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}

// BatchProcessEvents 
// @Summary 
// @Description 
// @Tags realtime-recommendations
// @Accept json
// @Produce json
// @Param request body BatchEventRequest true ""
// @Success 200 {object} BatchEventResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/events/batch [post]
func (h *RealtimeRecommendationHandler) BatchProcessEvents(c *gin.Context) {
	var req BatchEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: ": " + err.Error(),
		})
		return
	}

	if len(req.Events) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_events",
			Message: "",
		})
		return
	}

	if len(req.Events) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_events",
			Message: "100?,
		})
		return
	}

	startTime := time.Now()
	results := make([]EventProcessResult, len(req.Events))

	// 
	for i, event := range req.Events {
		err := h.realtimeService.ProcessEvent(c.Request.Context(), &event)
		results[i] = EventProcessResult{
			EventID: event.EventID,
			Success: err == nil,
		}
		if err != nil {
			results[i].Error = err.Error()
		}
	}

	processingTime := time.Since(startTime).Milliseconds()

	// 
	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	response := BatchEventResponse{
		TotalEvents:    len(req.Events),
		SuccessCount:   successCount,
		FailureCount:   len(req.Events) - successCount,
		ProcessingTime: processingTime,
		Results:        results,
	}

	c.JSON(http.StatusOK, response)
}

// GetRecommendationMetrics 
// @Summary 
// @Description 
// @Tags realtime-recommendations
// @Produce json
// @Param user_id query string false "ID"
// @Param hours query int false "" default(24)
// @Success 200 {object} RecommendationMetrics ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/realtime-recommendations/metrics [get]
func (h *RealtimeRecommendationHandler) GetRecommendationMetrics(c *gin.Context) {
	userID := c.Query("user_id")
	hoursStr := c.DefaultQuery("hours", "24")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	// ?
	// 
	metrics := RecommendationMetrics{
		UserID:               userID,
		TimeRange:            hours,
		TotalRecommendations: 1250,
		ClickThroughRate:     0.15,
		ConversionRate:       0.08,
		AvgResponseTime:      45,
		CacheHitRate:         0.85,
		UpdateFrequency:      12,
		Timestamp:            time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}

// 嶨?
// RealtimeRecommendationResponse 
type RealtimeRecommendationResponse struct {
	UserID          string                       `json:"user_id"`
	Recommendations []PersonalizedRecommendation `json:"recommendations"`
	Session         *SessionInfo                 `json:"session,omitempty"`
	Metadata        RealtimeMetadata             `json:"metadata"`
}

// SessionInfo 
type SessionInfo struct {
	SessionID    string                 `json:"session_id"`
	StartTime    time.Time              `json:"start_time"`
	LastActivity time.Time              `json:"last_activity"`
	Duration     int64                  `json:"duration_minutes"`
	EventCount   int                    `json:"event_count"`
	CurrentState *LearningStateInfo     `json:"current_state,omitempty"`
	Context      map[string]interface{} `json:"context"`
}

// LearningStateInfo ?
type LearningStateInfo struct {
	CurrentContent    string  `json:"current_content"`
	Progress          float64 `json:"progress"`
	Engagement        float64 `json:"engagement"`
	Difficulty        string  `json:"difficulty"`
	LearningStyle     string  `json:"learning_style"`
	FocusLevel        float64 `json:"focus_level"`
	ComprehensionRate float64 `json:"comprehension_rate"`
}

// RealtimeMetadata ?
type RealtimeMetadata struct {
	Strategy       string    `json:"strategy"`
	Timestamp      time.Time `json:"timestamp"`
	ProcessingTime int64     `json:"processing_time_ms"`
	IsRealtime     bool      `json:"is_realtime"`
}

// BatchEventRequest 
type BatchEventRequest struct {
	Events []recommendation.RealtimeEvent `json:"events" binding:"required"`
}

// BatchEventResponse 
type BatchEventResponse struct {
	TotalEvents    int                  `json:"total_events"`
	SuccessCount   int                  `json:"success_count"`
	FailureCount   int                  `json:"failure_count"`
	ProcessingTime int64                `json:"processing_time_ms"`
	Results        []EventProcessResult `json:"results"`
}

// EventProcessResult 
type EventProcessResult struct {
	EventID string `json:"event_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RecommendationMetrics 
type RecommendationMetrics struct {
	UserID               string    `json:"user_id,omitempty"`
	TimeRange            int       `json:"time_range_hours"`
	TotalRecommendations int       `json:"total_recommendations"`
	ClickThroughRate     float64   `json:"click_through_rate"`
	ConversionRate       float64   `json:"conversion_rate"`
	AvgResponseTime      int64     `json:"avg_response_time_ms"`
	CacheHitRate         float64   `json:"cache_hit_rate"`
	UpdateFrequency      int       `json:"update_frequency"`
	Timestamp            time.Time `json:"timestamp"`
}

// 

func convertToPersonalizedRecommendations(recommendations []*domainrecommendation.PersonalizedRecommendation) []PersonalizedRecommendation {
	result := make([]PersonalizedRecommendation, len(recommendations))
	for i, rec := range recommendations {
		contentID := ""
		if rec.ContentID != nil {
			contentID = rec.ContentID.String()
		}

		reasoning := ""
		if len(rec.Reasoning) > 0 {
			reasoning = strings.Join(rec.Reasoning, "; ")
		}

		estimatedTime := int(rec.EstimatedTime.Minutes())

		result[i] = PersonalizedRecommendation{
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
	return result
}

func convertToSessionInfo(session *recommendation.RealtimeUserSession) *SessionInfo {
	if session == nil {
		return nil
	}

	info := &SessionInfo{
		SessionID:    session.SessionID,
		StartTime:    session.StartTime,
		LastActivity: session.LastActivity,
		Duration:     int64(time.Since(session.StartTime).Minutes()),
		EventCount:   len(session.Events),
		Context:      session.Context,
	}

	if session.CurrentState != nil {
		info.CurrentState = &LearningStateInfo{
			CurrentContent:    session.CurrentState.CurrentContent,
			Progress:          session.CurrentState.Progress,
			Engagement:        session.CurrentState.Engagement,
			Difficulty:        session.CurrentState.Difficulty,
			LearningStyle:     session.CurrentState.LearningStyle,
			FocusLevel:        session.CurrentState.FocusLevel,
			ComprehensionRate: session.CurrentState.ComprehensionRate,
		}
	}

	return info
}

