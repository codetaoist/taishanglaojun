package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// CrossModalHandler 
type CrossModalHandler struct {
	crossModalService *services.CrossModalService
	upgrader          websocket.Upgrader
}

// NewCrossModalHandler 
func NewCrossModalHandler(crossModalService *services.CrossModalService) *CrossModalHandler {
	return &CrossModalHandler{
		crossModalService: crossModalService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 
			},
		},
	}
}

// ProcessCrossModalInference 
// @Summary 
// @Description ?
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param request body services.CrossModalRequest true ""
// @Success 200 {object} services.CrossModalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/inference [post]
func (h *CrossModalHandler) ProcessCrossModalInference(c *gin.Context) {
	var req services.CrossModalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// ID
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}
	req.UserID = userID.(string)

	// ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 
	ctx := context.Background()
	response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Cross-modal inference failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SemanticSearch 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param query query string true ""
// @Param provider query string false "AI" default(openai)
// @Param model query string false "AI" default(gpt-4)
// @Param max_results query int false "" default(10)
// @Param threshold query number false "" default(0.7)
// @Param request body []services.CrossModalInput true ""
// @Success 200 {object} services.CrossModalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/search [post]
func (h *CrossModalHandler) SemanticSearch(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing query parameter",
			Message: "Query parameter is required for semantic search",
		})
		return
	}

	var inputs []services.CrossModalInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input format",
			Message: err.Error(),
		})
		return
	}

	// 
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4")
	maxResults, _ := strconv.Atoi(c.DefaultQuery("max_results", "10"))
	threshold, _ := strconv.ParseFloat(c.DefaultQuery("threshold", "0.7"), 64)

	// 
	req := services.CrossModalRequest{
		ID:     uuid.New().String(),
		Type:   services.InferenceTypeSemanticSearch,
		Inputs: inputs,
		Query:  query,
		Config: services.CrossModalInferenceConfig{
			Provider:            provider,
			Model:               model,
			MaxResults:          maxResults,
			SimilarityThreshold: threshold,
			EnableExplanation:   true,
		},
		Timestamp: time.Now(),
	}

	// ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 
	ctx := context.Background()
	response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Semantic search failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ContentMatching 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param provider query string false "AI" default(openai)
// @Param model query string false "AI" default(gpt-4)
// @Param threshold query number false "" default(0.7)
// @Param request body []services.CrossModalInput true ""
// @Success 200 {object} services.CrossModalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/match [post]
func (h *CrossModalHandler) ContentMatching(c *gin.Context) {
	var inputs []services.CrossModalInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input format",
			Message: err.Error(),
		})
		return
	}

	if len(inputs) < 2 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Insufficient inputs",
			Message: "Content matching requires at least 2 inputs",
		})
		return
	}

	// 
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4")
	threshold, _ := strconv.ParseFloat(c.DefaultQuery("threshold", "0.7"), 64)

	// 
	req := services.CrossModalRequest{
		ID:     uuid.New().String(),
		Type:   services.InferenceTypeContentMatching,
		Inputs: inputs,
		Config: services.CrossModalInferenceConfig{
			Provider:            provider,
			Model:               model,
			SimilarityThreshold: threshold,
			EnableExplanation:   true,
		},
		Timestamp: time.Now(),
	}

	// ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 
	ctx := context.Background()
	response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Content matching failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MultiModalQA 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param query query string true ""
// @Param provider query string false "AI" default(openai)
// @Param model query string false "AI" default(gpt-4-vision-preview)
// @Param temperature query number false "" default(0.7)
// @Param request body []services.CrossModalInput true ""
// @Success 200 {object} services.CrossModalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/qa [post]
func (h *CrossModalHandler) MultiModalQA(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing query parameter",
			Message: "Query parameter is required for multimodal QA",
		})
		return
	}

	var inputs []services.CrossModalInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input format",
			Message: err.Error(),
		})
		return
	}

	// 
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4-vision-preview")
	temperature, _ := strconv.ParseFloat(c.DefaultQuery("temperature", "0.7"), 32)

	// 
	req := services.CrossModalRequest{
		ID:     uuid.New().String(),
		Type:   services.InferenceTypeMultiModalQA,
		Inputs: inputs,
		Query:  query,
		Config: services.CrossModalInferenceConfig{
			Provider:          provider,
			Model:             model,
			Temperature:       float32(temperature),
			EnableExplanation: true,
		},
		Timestamp: time.Now(),
	}

	// ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 
	ctx := context.Background()
	response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Multimodal QA failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SceneUnderstanding 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param provider query string false "AI" default(openai)
// @Param model query string false "AI" default(gpt-4-vision-preview)
// @Param request body []services.CrossModalInput true ""
// @Success 200 {object} services.CrossModalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/scene [post]
func (h *CrossModalHandler) SceneUnderstanding(c *gin.Context) {
	var inputs []services.CrossModalInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input format",
			Message: err.Error(),
		})
		return
	}

	// 
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4-vision-preview")

	// 
	req := services.CrossModalRequest{
		ID:     uuid.New().String(),
		Type:   services.InferenceTypeSceneUnderstand,
		Inputs: inputs,
		Config: services.CrossModalInferenceConfig{
			Provider:          provider,
			Model:             model,
			EnableExplanation: true,
		},
		Timestamp: time.Now(),
	}

	// ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 
	ctx := context.Background()
	response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Scene understanding failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// EmotionAnalysis 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param provider query string false "AI" default(openai)
// @Param model query string false "AI" default(gpt-4)
// @Param request body []services.CrossModalInput true ""
// @Success 200 {object} services.CrossModalResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/emotion [post]
func (h *CrossModalHandler) EmotionAnalysis(c *gin.Context) {
	var inputs []services.CrossModalInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid input format",
			Message: err.Error(),
		})
		return
	}

	// 
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4")

	// 
	req := services.CrossModalRequest{
		ID:     uuid.New().String(),
		Type:   services.InferenceTypeEmotionAnalysis,
		Inputs: inputs,
		Config: services.CrossModalInferenceConfig{
			Provider:          provider,
			Model:             model,
			EnableExplanation: true,
		},
		Timestamp: time.Now(),
	}

	// ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 
	ctx := context.Background()
	response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Emotion analysis failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// StreamCrossModalInference ?
// @Summary 
// @Description WebSocket
// @Tags crossmodal
// @Accept json
// @Produce json
// @Router /api/v1/crossmodal/stream [get]
func (h *CrossModalHandler) StreamCrossModalInference(c *gin.Context) {
	// WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "WebSocket upgrade failed",
			Message: err.Error(),
		})
		return
	}
	defer conn.Close()

	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		conn.WriteJSON(map[string]interface{}{
			"error": "Unauthorized: User ID not found",
		})
		return
	}

	// WebSocket
	for {
		var req services.CrossModalRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				break
			}
			conn.WriteJSON(map[string]interface{}{
				"error": "Invalid request format: " + err.Error(),
			})
			continue
		}

		// ID
		req.UserID = userID.(string)
		if req.ID == "" {
			req.ID = uuid.New().String()
		}
		req.Timestamp = time.Now()

		// 
		ctx := context.Background()
		response, err := h.crossModalService.ProcessCrossModalInference(ctx, &req)
		if err != nil {
			conn.WriteJSON(map[string]interface{}{
				"error":      "Cross-modal inference failed",
				"message":    err.Error(),
				"request_id": req.ID,
			})
			continue
		}

		// 
		if err := conn.WriteJSON(response); err != nil {
			break
		}
	}
}

// GetInferenceHistory 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param user_id query string false "ID"
// @Param session_id query string false "ID"
// @Param type query string false ""
// @Param limit query int false "" default(20)
// @Param offset query int false "" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/history [get]
func (h *CrossModalHandler) GetInferenceHistory(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// 
	queryUserID := c.Query("user_id")
	if queryUserID == "" {
		queryUserID = userID.(string)
	}
	sessionID := c.Query("session_id")
	inferenceType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 
	history := map[string]interface{}{
		"user_id":    queryUserID,
		"session_id": sessionID,
		"type":       inferenceType,
		"total":      0,
		"limit":      limit,
		"offset":     offset,
		"records":    []interface{}{},
		"timestamp":  time.Now(),
	}

	c.JSON(http.StatusOK, history)
}

// GetInferenceStats 
// @Summary 
// @Description 
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param user_id query string false "ID"
// @Param period query string false "" default(7d)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/stats [get]
func (h *CrossModalHandler) GetInferenceStats(c *gin.Context) {
	// ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// 
	queryUserID := c.Query("user_id")
	if queryUserID == "" {
		queryUserID = userID.(string)
	}
	period := c.DefaultQuery("period", "7d")

	// 
	stats := map[string]interface{}{
		"user_id":          queryUserID,
		"period":           period,
		"total_inferences": 0,
		"inference_types": map[string]int{
			"semantic_search":   0,
			"content_matching":  0,
			"cross_modal_align": 0,
			"multimodal_qa":     0,
			"scene_understand":  0,
			"emotion_analysis":  0,
			"content_generate":  0,
		},
		"avg_confidence":      0.0,
		"avg_processing_time": "0ms",
		"success_rate":        0.0,
		"timestamp":           time.Now(),
	}

	c.JSON(http.StatusOK, stats)
}

