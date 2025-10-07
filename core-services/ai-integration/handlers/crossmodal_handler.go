package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"taishanglaojun/core-services/ai-integration/services"
)

// CrossModalHandler 跨模态推理处理器
type CrossModalHandler struct {
	crossModalService *services.CrossModalService
	upgrader          websocket.Upgrader
}

// NewCrossModalHandler 创建跨模态推理处理器
func NewCrossModalHandler(crossModalService *services.CrossModalService) *CrossModalHandler {
	return &CrossModalHandler{
		crossModalService: crossModalService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该进行适当的来源检查
			},
		},
	}
}

// ProcessCrossModalInference 处理跨模态推理请求
// @Summary 处理跨模态推理请求
// @Description 支持多种模态数据之间的智能推理和关联分析
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param request body services.CrossModalRequest true "跨模态推理请求"
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

	// 设置请求ID和时间戳
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}
	req.UserID = userID.(string)

	// 获取会话ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 处理请求
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

// SemanticSearch 语义搜索
// @Summary 多模态语义搜索
// @Description 在多种模态数据中进行语义搜索
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param query query string true "搜索查询"
// @Param provider query string false "AI提供商" default(openai)
// @Param model query string false "AI模型" default(gpt-4)
// @Param max_results query int false "最大结果数" default(10)
// @Param threshold query number false "相似度阈值" default(0.7)
// @Param request body []services.CrossModalInput true "输入数据"
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

	// 解析查询参数
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4")
	maxResults, _ := strconv.Atoi(c.DefaultQuery("max_results", "10"))
	threshold, _ := strconv.ParseFloat(c.DefaultQuery("threshold", "0.7"), 64)

	// 构建请求
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

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// 获取会话ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 处理请求
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

// ContentMatching 内容匹配
// @Summary 多模态内容匹配
// @Description 在多种模态数据之间进行内容匹配
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param provider query string false "AI提供商" default(openai)
// @Param model query string false "AI模型" default(gpt-4)
// @Param threshold query number false "相似度阈值" default(0.7)
// @Param request body []services.CrossModalInput true "输入数据"
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

	// 解析查询参数
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4")
	threshold, _ := strconv.ParseFloat(c.DefaultQuery("threshold", "0.7"), 64)

	// 构建请求
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

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// 获取会话ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 处理请求
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

// MultiModalQA 多模态问答
// @Summary 多模态问答
// @Description 基于多种模态数据进行智能问答
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param query query string true "问题"
// @Param provider query string false "AI提供商" default(openai)
// @Param model query string false "AI模型" default(gpt-4-vision-preview)
// @Param temperature query number false "温度参数" default(0.7)
// @Param request body []services.CrossModalInput true "上下文数据"
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

	// 解析查询参数
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4-vision-preview")
	temperature, _ := strconv.ParseFloat(c.DefaultQuery("temperature", "0.7"), 32)

	// 构建请求
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

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// 获取会话ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 处理请求
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

// SceneUnderstanding 场景理解
// @Summary 多模态场景理解
// @Description 基于多种模态数据进行场景理解和分析
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param provider query string false "AI提供商" default(openai)
// @Param model query string false "AI模型" default(gpt-4-vision-preview)
// @Param request body []services.CrossModalInput true "场景数据"
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

	// 解析查询参数
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4-vision-preview")

	// 构建请求
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

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// 获取会话ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 处理请求
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

// EmotionAnalysis 情感分析
// @Summary 多模态情感分析
// @Description 基于多种模态数据进行情感分析
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param provider query string false "AI提供商" default(openai)
// @Param model query string false "AI模型" default(gpt-4)
// @Param request body []services.CrossModalInput true "情感数据"
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

	// 解析查询参数
	provider := c.DefaultQuery("provider", "openai")
	model := c.DefaultQuery("model", "gpt-4")

	// 构建请求
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

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if exists {
		req.UserID = userID.(string)
	}

	// 获取会话ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	req.SessionID = sessionID

	// 处理请求
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

// StreamCrossModalInference 流式跨模态推理
// @Summary 流式跨模态推理
// @Description 通过WebSocket进行流式跨模态推理
// @Tags crossmodal
// @Accept json
// @Produce json
// @Router /api/v1/crossmodal/stream [get]
func (h *CrossModalHandler) StreamCrossModalInference(c *gin.Context) {
	// 升级到WebSocket连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "WebSocket upgrade failed",
			Message: err.Error(),
		})
		return
	}
	defer conn.Close()

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		conn.WriteJSON(map[string]interface{}{
			"error": "Unauthorized: User ID not found",
		})
		return
	}

	// 处理WebSocket消息
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

		// 设置用户ID
		req.UserID = userID.(string)
		if req.ID == "" {
			req.ID = uuid.New().String()
		}
		req.Timestamp = time.Now()

		// 处理请求
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

		// 发送响应
		if err := conn.WriteJSON(response); err != nil {
			break
		}
	}
}

// GetInferenceHistory 获取推理历史
// @Summary 获取跨模态推理历史
// @Description 获取用户的跨模态推理历史记录
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param user_id query string false "用户ID"
// @Param session_id query string false "会话ID"
// @Param type query string false "推理类型"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/history [get]
func (h *CrossModalHandler) GetInferenceHistory(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// 解析查询参数
	queryUserID := c.Query("user_id")
	if queryUserID == "" {
		queryUserID = userID.(string)
	}
	sessionID := c.Query("session_id")
	inferenceType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 构建历史查询响应（模拟数据）
	history := map[string]interface{}{
		"user_id":     queryUserID,
		"session_id":  sessionID,
		"type":        inferenceType,
		"total":       0,
		"limit":       limit,
		"offset":      offset,
		"records":     []interface{}{},
		"timestamp":   time.Now(),
	}

	c.JSON(http.StatusOK, history)
}

// GetInferenceStats 获取推理统计
// @Summary 获取跨模态推理统计信息
// @Description 获取用户的跨模态推理统计信息
// @Tags crossmodal
// @Accept json
// @Produce json
// @Param user_id query string false "用户ID"
// @Param period query string false "统计周期" default(7d)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/crossmodal/stats [get]
func (h *CrossModalHandler) GetInferenceStats(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
		})
		return
	}

	// 解析查询参数
	queryUserID := c.Query("user_id")
	if queryUserID == "" {
		queryUserID = userID.(string)
	}
	period := c.DefaultQuery("period", "7d")

	// 构建统计响应（模拟数据）
	stats := map[string]interface{}{
		"user_id": queryUserID,
		"period":  period,
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
		"avg_confidence":    0.0,
		"avg_processing_time": "0ms",
		"success_rate":      0.0,
		"timestamp":         time.Now(),
	}

	c.JSON(http.StatusOK, stats)
}