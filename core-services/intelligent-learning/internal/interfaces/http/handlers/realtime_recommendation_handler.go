package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services/recommendation"
)

// RealtimeRecommendationHandler 实时推荐处理
// @Summary 实时推荐处理
// @Description 处理与实时推荐相关的HTTP请求
// @Tags realtime-recommendations
// @Accept json
// @Produce json
// @Param request body recommendation.RealtimeEvent true "实时事件"
// @Success 200 {object} map[string]interface{} "事件处理成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/events [post]
type RealtimeRecommendationHandler struct {
	realtimeService *recommendation.RealtimeRecommendationService
	upgrader        websocket.Upgrader
}

// NewRealtimeRecommendationHandler 创建实时推荐处理
// @Summary 创建实时推荐处理
// @Description 创建实时推荐处理
// @Tags realtime-recommendations
// @Param realtimeService body recommendation.RealtimeRecommendationService true "实时推荐服务"
// @Success 200 {object} RealtimeRecommendationHandler "实时推荐处理成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/events [post]
func NewRealtimeRecommendationHandler(realtimeService *recommendation.RealtimeRecommendationService) *RealtimeRecommendationHandler {
	return &RealtimeRecommendationHandler{
		realtimeService: realtimeService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该进行适当的来源检查
			},
		},
	}
}

// ProcessRealtimeEvent 处理实时事件
// @Summary 处理实时事件
// @Description 处理用户的实时学习事件
// @Tags realtime-recommendations
// @Accept json
// @Produce json
// @Param request body recommendation.RealtimeEvent true "实时事件"
// @Success 200 {object} map[string]interface{} "事件处理成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/events [post]
func (h *RealtimeRecommendationHandler) ProcessRealtimeEvent(c *gin.Context) {
	var event recommendation.RealtimeEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	// 处理事件
	err := h.realtimeService.ProcessEvent(c.Request.Context(), &event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "event_processing_failed",
			Message: "事件处理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "事件处理成功",
		"event_id":  event.EventID,
		"timestamp": event.Timestamp,
	})
}

// GetRealtimeRecommendations 获取实时推荐
// @Summary 获取实时推荐
// @Description 获取基于用户实时行为的推荐
// @Tags realtime-recommendations
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} RealtimeRecommendationResponse "实时推荐成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/{user_id} [get]
func (h *RealtimeRecommendationHandler) GetRealtimeRecommendations(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	startTime := time.Now()

	// 获取实时推荐
	recommendations, err := h.realtimeService.GetRealtimeRecommendations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "recommendation_failed",
			Message: "获取实时推荐失败: " + err.Error(),
		})
		return
	}

	// 获取用户会话信息
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

// GetUserSession 获取用户会话信息
// @Summary 获取用户会话信息
// @Description 获取用户当前的学习会话信息
// @Tags realtime-recommendations
// @Produce json
// @Param user_id path string true "用户ID"
// @Success 200 {object} recommendation.RealtimeUserSession "会话信息"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "会话不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/sessions/{user_id} [get]
func (h *RealtimeRecommendationHandler) GetUserSession(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	session, err := h.realtimeService.GetUserSession(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "session_not_found",
			Message: "用户会话不存在: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, session)
}

// SubscribeToRecommendationUpdates WebSocket订阅推荐更新
// @Summary WebSocket订阅推荐更新
// @Description 通过WebSocket实时接收推荐更新
// @Tags realtime-recommendations
// @Param user_id query string true "用户ID"
// @Success 101 {string} string "WebSocket连接建立成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/subscribe [get]
func (h *RealtimeRecommendationHandler) SubscribeToRecommendationUpdates(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_parameter",
			Message: "用户ID不能为空",
		})
		return
	}

	// 升级到WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "websocket_upgrade_failed",
			Message: "WebSocket连接升级失败: " + err.Error(),
		})
		return
	}
	defer conn.Close()

	// 订阅推荐更新
	updateChannel := h.realtimeService.SubscribeToUpdates(userID)
	defer h.realtimeService.UnsubscribeFromUpdates(userID)

	// 发送初始推荐更新
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

	// 处理WebSocket消息
	go func() {
		for {
			var message map[string]interface{}
			err := conn.ReadJSON(&message)
			if err != nil {
				break
			}
			// 处理客户端消息（如心跳包）
			if message["type"] == "heartbeat" {
				continue
			}
		}
	}()

	// 监听推荐更新
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

// BatchProcessEvents 批量处理事件
// @Summary 批量处理事件
// @Description 批量处理多个实时事件
// @Tags realtime-recommendations
// @Accept json
// @Produce json
// @Param request body BatchEventRequest true "批量事件请求"
// @Success 200 {object} BatchEventResponse "批量处理成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/events/batch [post]
func (h *RealtimeRecommendationHandler) BatchProcessEvents(c *gin.Context) {
	var req BatchEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "请求参数格式错误: " + err.Error(),
		})
		return
	}

	if len(req.Events) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_events",
			Message: "事件列表不能为空",
		})
		return
	}

	if len(req.Events) > 100 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "too_many_events",
			Message: "批量事件数量不能超过100条",
		})
		return
	}

	startTime := time.Now()
	results := make([]EventProcessResult, len(req.Events))

	// 处理每个事件
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

	// 统计结果
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

// GetRecommendationMetrics 获取推荐指标
// @Summary 获取推荐指标
// @Description 获取实时推荐系统的性能指标
// @Tags realtime-recommendations
// @Produce json
// @Param user_id query string false "用户ID（可选）"
// @Param hours query int false "时间范围（小时）" default(24)
// @Success 200 {object} RecommendationMetrics "推荐指标"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/realtime-recommendations/metrics [get]
func (h *RealtimeRecommendationHandler) GetRecommendationMetrics(c *gin.Context) {
	userID := c.Query("user_id")
	hoursStr := c.DefaultQuery("hours", "24")

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}

	// 这里应该从监控系统或数据库获取实际指标
	// 目前返回模拟数据
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

// 响应结构体定义
// RealtimeRecommendationResponse 实时推荐响应
type RealtimeRecommendationResponse struct {
	UserID          string                       `json:"user_id"`
	Recommendations []PersonalizedRecommendation `json:"recommendations"`
	Session         *SessionInfo                 `json:"session,omitempty"`
	Metadata        RealtimeMetadata             `json:"metadata"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID    string                 `json:"session_id"`
	StartTime    time.Time              `json:"start_time"`
	LastActivity time.Time              `json:"last_activity"`
	Duration     int64                  `json:"duration_minutes"`
	EventCount   int                    `json:"event_count"`
	CurrentState *LearningStateInfo     `json:"current_state,omitempty"`
	Context      map[string]interface{} `json:"context"`
}

// LearningStateInfo 学习状态信息
type LearningStateInfo struct {
	CurrentContent    string  `json:"current_content"`
	Progress          float64 `json:"progress"`
	Engagement        float64 `json:"engagement"`
	Difficulty        string  `json:"difficulty"`
	LearningStyle     string  `json:"learning_style"`
	FocusLevel        float64 `json:"focus_level"`
	ComprehensionRate float64 `json:"comprehension_rate"`
}

// RealtimeMetadata 实时元数据
type RealtimeMetadata struct {
	Strategy       string    `json:"strategy"`
	Timestamp      time.Time `json:"timestamp"`
	ProcessingTime int64     `json:"processing_time_ms"`
	IsRealtime     bool      `json:"is_realtime"`
}

// BatchEventRequest 批量事件请求
type BatchEventRequest struct {
	Events []recommendation.RealtimeEvent `json:"events" binding:"required"`
}

// BatchEventResponse 批量事件响应
type BatchEventResponse struct {
	TotalEvents    int                  `json:"total_events"`
	SuccessCount   int                  `json:"success_count"`
	FailureCount   int                  `json:"failure_count"`
	ProcessingTime int64                `json:"processing_time_ms"`
	Results        []EventProcessResult `json:"results"`
}

// EventProcessResult 事件处理结果
type EventProcessResult struct {
	EventID string `json:"event_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RecommendationMetrics 推荐指标
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

// 辅助函数

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
