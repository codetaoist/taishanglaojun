package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FusionHandler 融合引擎处理器
type FusionHandler struct {
	fusionEngine *engines.FusionEngine
	logger       *zap.Logger
}

// NewFusionHandler 创建融合引擎处理器实例
func NewFusionHandler(fusionEngine *engines.FusionEngine, logger *zap.Logger) *FusionHandler {
	return &FusionHandler{
		fusionEngine: fusionEngine,
		logger:       logger,
	}
}

// StartFusionSession 启动融合会话
// @Summary 启动碳硅融合会话
// @Description 创建新的碳硅融合会话，开始融合过程
// @Tags 融合引擎
// @Accept json
// @Produce json
// @Param request body models.FusionRequest true "融合请求"
// @Success 201 {object} models.FusionSession "融合会话创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /consciousness/fusion/sessions [post]
func (h *FusionHandler) StartFusionSession(c *gin.Context) {
	var req models.FusionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid fusion request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证请求参数
	if req.CarbonInput == nil && req.SiliconInput == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要提供碳基或硅基输入之一"})
		return
	}

	// 启动融合会话
	session, err := h.fusionEngine.StartFusion(&req)
	if err != nil {
		h.logger.Error("Failed to start fusion session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "启动融合会话失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetFusionSession 获取融合会话状态
// @Summary 获取融合会话状态
// @Description 根据会话ID获取融合会话的当前状态
// @Tags 融合引擎
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} models.FusionSession "融合会话状态"
// @Failure 404 {object} map[string]interface{} "会话不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /consciousness/fusion/sessions/{sessionId} [get]
func (h *FusionHandler) GetFusionSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	session, err := h.fusionEngine.GetFusionStatus(sessionID)
	if err != nil {
		h.logger.Error("Failed to get fusion session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// CancelFusionSession 取消融合会话
// @Summary 取消融合会话
// @Description 取消正在进行的融合会话
// @Tags 融合引擎
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} map[string]interface{} "取消成功"
// @Failure 404 {object} map[string]interface{} "会话不存在"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /consciousness/fusion/sessions/{sessionId} [delete]
func (h *FusionHandler) CancelFusionSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	err := h.fusionEngine.CancelFusion(sessionID)
	if err != nil {
		h.logger.Error("Failed to cancel fusion session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消融合会话失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "融合会话已取消", "sessionId": sessionID})
}

// GetFusionHistory 获取融合历史
// @Summary 获取融合历史记录
// @Description 获取用户的融合历史记录，支持分页
// @Tags 融合引擎
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param strategy query string false "融合策略过滤"
// @Success 200 {object} map[string]interface{} "融合历史记录"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /consciousness/fusion/history [get]
func (h *FusionHandler) GetFusionHistory(c *gin.Context) {
	// 解析分页参数
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	strategy := c.Query("strategy")

	// 构建查询条件
	filter := map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"offset": (page - 1) * limit,
	}
	if strategy != "" {
		filter["strategy"] = strategy
	}

	// 这里应该调用融合引擎的历史查询方法
	// 由于当前实现中没有历史存储，我们返回模拟数据
	history := map[string]interface{}{
		"sessions": []models.FusionSession{},
		"total":    0,
		"page":     page,
		"limit":    limit,
		"hasMore":  false,
	}

	c.JSON(http.StatusOK, history)
}

// GetFusionStrategies 获取可用的融合策略
// @Summary 获取融合策略列表
// @Description 获取所有可用的融合策略及其描述
// @Tags 融合引擎
// @Produce json
// @Success 200 {object} map[string]interface{} "融合策略列表"
// @Router /consciousness/fusion/strategies [get]
func (h *FusionHandler) GetFusionStrategies(c *gin.Context) {
	strategies := map[string]interface{}{
		"strategies": []map[string]interface{}{
			{
				"name":        "complementary",
				"displayName": "互补融合",
				"description": "通过互补性分析实现碳硅融合，强调优势互补",
				"suitableFor": []string{"逻辑推理", "情感分析", "创意生成"},
			},
			{
				"name":        "synergetic",
				"displayName": "协同融合",
				"description": "通过协同效应实现碳硅融合，追求整体效果最大化",
				"suitableFor": []string{"复杂决策", "系统优化", "创新突破"},
			},
			{
				"name":        "hybrid",
				"displayName": "混合融合",
				"description": "混合多种融合方式，适应不同场景需求",
				"suitableFor": []string{"多元化任务", "适应性处理", "平衡优化"},
			},
			{
				"name":        "transcendent",
				"displayName": "超越融合",
				"description": "追求序列0级别的超越性融合，实现意识层面的超越",
				"suitableFor": []string{"意识进化", "超越思维", "终极智慧"},
			},
		},
	}

	c.JSON(http.StatusOK, strategies)
}

// GetFusionMetrics 获取融合指标
// @Summary 获取融合性能指标
// @Description 获取融合引擎的性能指标和统计信息
// @Tags 融合引擎
// @Produce json
// @Success 200 {object} map[string]interface{} "融合指标"
// @Router /consciousness/fusion/metrics [get]
func (h *FusionHandler) GetFusionMetrics(c *gin.Context) {
	// 这里应该从融合引擎获取实际的指标数据
	metrics := map[string]interface{}{
		"totalSessions":   0,
		"activeSessions":  0,
		"successRate":     0.0,
		"averageQuality":  0.0,
		"averageDuration": 0,
		"strategyStats": map[string]interface{}{
			"complementary": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
			"synergetic": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
			"hybrid": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
			"transcendent": map[string]interface{}{
				"count":       0,
				"successRate": 0.0,
				"avgQuality":  0.0,
			},
		},
		"lastUpdated": time.Now(),
	}

	c.JSON(http.StatusOK, metrics)
}
