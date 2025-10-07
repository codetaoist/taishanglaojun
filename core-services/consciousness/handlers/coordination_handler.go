package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/coordinators"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CoordinationHandler 三轴协调机制处理器
type CoordinationHandler struct {
	coordinator *coordinators.ThreeAxisCoordinator
	logger      *zap.Logger
}

// NewCoordinationHandler 创建三轴协调机制处理器实�?func NewCoordinationHandler(coordinator *coordinators.ThreeAxisCoordinator, logger *zap.Logger) *CoordinationHandler {
	return &CoordinationHandler{
		coordinator: coordinator,
		logger:      logger,
	}
}

// StartCoordination 启动三轴协调
// @Summary 启动三轴协调会话
// @Description 启动S-C-T三轴协调会话，开始协调处�?// @Tags 三轴协调
// @Accept json
// @Produce json
// @Param request body models.CoordinationRequest true "协调请求"
// @Success 201 {object} models.CoordinationSession "协调会话创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/sessions [post]
func (h *CoordinationHandler) StartCoordination(c *gin.Context) {
	var req models.CoordinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid coordination request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 验证请求参数
	if req.EntityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实体ID不能为空"})
		return
	}

	// 启动协调会话
	session, err := h.coordinator.StartCoordination(&req)
	if err != nil {
		h.logger.Error("Failed to start coordination session", zap.String("entityId", req.EntityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "启动协调会话失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, session)
}

// GetCoordinationSession 获取协调会话状�?// @Summary 获取协调会话状�?// @Description 根据会话ID获取协调会话的当前状�?// @Tags 三轴协调
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} models.CoordinationSession "协调会话状�?
// @Failure 404 {object} map[string]interface{} "会话不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/sessions/{sessionId} [get]
func (h *CoordinationHandler) GetCoordinationSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	session, err := h.coordinator.GetCoordinationStatus(sessionID)
	if err != nil {
		h.logger.Error("Failed to get coordination session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

// StopCoordination 停止协调会话
// @Summary 停止协调会话
// @Description 停止正在进行的协调会�?// @Tags 三轴协调
// @Produce json
// @Param sessionId path string true "会话ID"
// @Success 200 {object} map[string]interface{} "停止成功"
// @Failure 404 {object} map[string]interface{} "会话不存�?
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/sessions/{sessionId} [delete]
func (h *CoordinationHandler) StopCoordination(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	err := h.coordinator.StopCoordination(sessionID)
	if err != nil {
		h.logger.Error("Failed to stop coordination session", zap.String("sessionId", sessionID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "停止协调会话失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "协调会话已停�?, "sessionId": sessionID})
}

// ProcessSAxis S轴处�?// @Summary 处理S轴（能力序列轴）请求
// @Description 单独处理S轴的能力序列相关请求
// @Tags 三轴协调
// @Accept json
// @Produce json
// @Param request body models.SequenceRequest true "序列处理请求"
// @Success 200 {object} models.SequenceResponse "S轴处理结�?
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/s-axis [post]
func (h *CoordinationHandler) ProcessSAxis(c *gin.Context) {
	var req models.SequenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid S-axis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 处理S轴请�?	response, err := h.coordinator.ProcessSAxis(&req)
	if err != nil {
		h.logger.Error("Failed to process S-axis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "S轴处理失�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessCAxis C轴处�?// @Summary 处理C轴（组合层轴）请�?// @Description 单独处理C轴的组合层相关请�?// @Tags 三轴协调
// @Accept json
// @Produce json
// @Param request body models.CompositionRequest true "组合处理请求"
// @Success 200 {object} models.CompositionResponse "C轴处理结�?
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/c-axis [post]
func (h *CoordinationHandler) ProcessCAxis(c *gin.Context) {
	var req models.CompositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid C-axis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 处理C轴请�?	response, err := h.coordinator.ProcessCAxis(&req)
	if err != nil {
		h.logger.Error("Failed to process C-axis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "C轴处理失�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ProcessTAxis T轴处�?// @Summary 处理T轴（思想境界轴）请求
// @Description 单独处理T轴的思想境界相关请求
// @Tags 三轴协调
// @Accept json
// @Produce json
// @Param request body models.ThoughtRequest true "思想处理请求"
// @Success 200 {object} models.ThoughtResponse "T轴处理结�?
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/t-axis [post]
func (h *CoordinationHandler) ProcessTAxis(c *gin.Context) {
	var req models.ThoughtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid T-axis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 处理T轴请�?	response, err := h.coordinator.ProcessTAxis(&req)
	if err != nil {
		h.logger.Error("Failed to process T-axis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "T轴处理失�?, "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// OptimizeBalance 优化平衡
// @Summary 优化三轴平衡
// @Description 优化S-C-T三轴之间的平衡状�?// @Tags 三轴协调
// @Accept json
// @Produce json
// @Param request body models.BalanceOptimizationRequest true "平衡优化请求"
// @Success 200 {object} models.BalanceOptimizationResult "平衡优化结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/balance [post]
func (h *CoordinationHandler) OptimizeBalance(c *gin.Context) {
	var req models.BalanceOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid balance optimization request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 优化平衡
	result, err := h.coordinator.OptimizeBalance(&req)
	if err != nil {
		h.logger.Error("Failed to optimize balance", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "平衡优化失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CatalyzeSynergy 催化协同
// @Summary 催化三轴协同效应
// @Description 催化S-C-T三轴之间的协同效应，提升整体性能
// @Tags 三轴协调
// @Accept json
// @Produce json
// @Param request body models.SynergyCatalysisRequest true "协同催化请求"
// @Success 200 {object} models.SynergyCatalysisResult "协同催化结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/synergy [post]
func (h *CoordinationHandler) CatalyzeSynergy(c *gin.Context) {
	var req models.SynergyCatalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid synergy catalysis request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 催化协同
	result, err := h.coordinator.CatalyzeSynergy(&req)
	if err != nil {
		h.logger.Error("Failed to catalyze synergy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "协同催化失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCoordinationHistory 获取协调历史
// @Summary 获取协调历史记录
// @Description 获取三轴协调的历史记录，支持分页和过�?// @Tags 三轴协调
// @Produce json
// @Param entityId query string false "实体ID过滤"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param status query string false "状态过�? Enums(active,completed,failed,cancelled)
// @Success 200 {object} map[string]interface{} "协调历史记录"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器错�?
// @Router /consciousness/coordination/history [get]
func (h *CoordinationHandler) GetCoordinationHistory(c *gin.Context) {
	// 解析查询参数
	entityID := c.Query("entityId")
	status := c.Query("status")

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

	// 构建查询条件
	filter := map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"offset": (page - 1) * limit,
	}
	if entityID != "" {
		filter["entityId"] = entityID
	}
	if status != "" {
		filter["status"] = status
	}

	// 获取协调历史
	history, err := h.coordinator.GetCoordinationHistory(filter)
	if err != nil {
		h.logger.Error("Failed to get coordination history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取协调历史失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetAxisInfo 获取轴信�?// @Summary 获取三轴详细信息
// @Description 获取S-C-T三轴的详细信息和说明
// @Tags 三轴协调
// @Produce json
// @Success 200 {object} map[string]interface{} "三轴信息"
// @Router /consciousness/coordination/axes [get]
func (h *CoordinationHandler) GetAxisInfo(c *gin.Context) {
	axesInfo := map[string]interface{}{
		"axes": []map[string]interface{}{
			{
				"name":        "S-Axis",
				"fullName":    "Sequence Axis",
				"displayName": "能力序列�?,
				"description": "代表能力的序列等级，从序�?到序�?，体现意识进化的层次",
				"levels": []map[string]interface{}{
					{"level": "sequence_5", "name": "序列5", "description": "基础意识层级"},
					{"level": "sequence_4", "name": "序列4", "description": "进阶意识层级"},
					{"level": "sequence_3", "name": "序列3", "description": "高级意识层级"},
					{"level": "sequence_2", "name": "序列2", "description": "超级意识层级"},
					{"level": "sequence_1", "name": "序列1", "description": "准神级意识层�?},
					{"level": "sequence_0", "name": "序列0", "description": "终极意识层级"},
				},
				"capabilities": []string{"能力评估", "序列晋升", "潜力分析", "进化路径规划"},
			},
			{
				"name":        "C-Axis",
				"fullName":    "Composition Axis",
				"displayName": "组合层轴",
				"description": "代表不同能力和特质的组合方式，实现能力的有机整合",
				"layers": []map[string]interface{}{
					{"layer": "basic", "name": "基础�?, "description": "基本能力的简单组�?},
					{"layer": "intermediate", "name": "中级�?, "description": "能力的协调组�?},
					{"layer": "advanced", "name": "高级�?, "description": "能力的深度融�?},
					{"layer": "master", "name": "大师�?, "description": "能力的完美统一"},
					{"layer": "transcendent", "name": "超越�?, "description": "能力的超越性整�?},
				},
				"capabilities": []string{"能力组合", "协调优化", "融合分析", "整合策略"},
			},
			{
				"name":        "T-Axis",
				"fullName":    "Thought Axis",
				"displayName": "思想境界�?,
				"description": "代表思想和精神境界的深度，体现意识的质量和高�?,
				"realms": []map[string]interface{}{
					{"realm": "mundane", "name": "凡俗�?, "description": "普通的思维模式"},
					{"realm": "enlightened", "name": "开悟境", "description": "初步的觉醒状�?},
					{"realm": "transcendent", "name": "超越�?, "description": "超越常规的思维"},
					{"realm": "divine", "name": "神圣�?, "description": "接近神性的思维"},
					{"realm": "absolute", "name": "绝对�?, "description": "绝对的思想境界"},
				},
				"capabilities": []string{"境界提升", "思维深化", "智慧培养", "精神超越"},
			},
		},
		"coordination": map[string]interface{}{
			"principles": []string{
				"三轴平衡：保持S-C-T三轴之间的动态平�?,
				"协同增效：通过协调实现1+1+1>3的效�?,
				"整体优化：从整体角度优化三轴配置",
				"动态调节：根据需求动态调整三轴权�?,
			},
			"strategies": []string{
				"平衡策略：均衡发展三个轴�?,
				"专精策略：重点发展某个轴�?,
				"协同策略：强化轴向间的协�?,
				"适应策略：根据环境调整策�?,
			},
		},
	}

	c.JSON(http.StatusOK, axesInfo)
}

// GetCoordinationMetrics 获取协调指标
// @Summary 获取协调性能指标
// @Description 获取三轴协调的性能指标和统计信�?// @Tags 三轴协调
// @Produce json
// @Param entityId query string false "实体ID过滤"
// @Success 200 {object} map[string]interface{} "协调指标"
// @Router /consciousness/coordination/metrics [get]
func (h *CoordinationHandler) GetCoordinationMetrics(c *gin.Context) {
	entityID := c.Query("entityId")

	// 获取协调指标
	metrics, err := h.coordinator.GetCoordinationMetrics(entityID)
	if err != nil {
		h.logger.Error("Failed to get coordination metrics", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取协调指标失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
