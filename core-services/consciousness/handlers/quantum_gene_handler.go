package handlers

import (
	"net/http"
	"strconv"

	"github.com/codetaoist/taishanglaojun/core-services/consciousness/engines"
	"github.com/codetaoist/taishanglaojun/core-services/consciousness/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// QuantumGeneHandler 
type QuantumGeneHandler struct {
	geneManager *engines.QuantumGeneManager
	logger      *zap.Logger
}

// NewQuantumGeneHandler 
func NewQuantumGeneHandler(geneManager *engines.QuantumGeneManager, logger *zap.Logger) *QuantumGeneHandler {
	return &QuantumGeneHandler{
		geneManager: geneManager,
		logger:      logger,
	}
}

// CreateGenePool 
// @Summary 
// @Description 崴
// @Tags 
// @Accept json
// @Produce json
// @Param request body models.GenePoolCreateRequest true ""
// @Success 201 {object} models.GenePool ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools [post]
func (h *QuantumGeneHandler) CreateGenePool(c *gin.Context) {
	var req models.GenePoolCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid gene pool create request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	if req.EntityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	pool, err := h.geneManager.CreateGenePool(c.Request.Context(), req.EntityID, req.Name, req.Description)
	if err != nil {
		h.logger.Error("Failed to create gene pool", zap.String("entityId", req.EntityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pool)
}

// GetGenePool 
// @Summary 
// @Description ID
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Success 200 {object} models.GenePool ""
// @Failure 404 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools/{entityId} [get]
func (h *QuantumGeneHandler) GetGenePool(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	pool, err := h.geneManager.GetGenePool(entityID)
	if err != nil {
		h.logger.Error("Failed to get gene pool", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pool)
}

// AddGene 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param entityId path string true "ID"
// @Param request body models.QuantumGene true ""
// @Success 201 {object} models.QuantumGene ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools/{entityId}/genes [post]
func (h *QuantumGeneHandler) AddGene(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var gene models.QuantumGene
	if err := c.ShouldBindJSON(&gene); err != nil {
		h.logger.Error("Invalid gene data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	if gene.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	// 
	err := h.geneManager.AddGeneToPool(c.Request.Context(), entityID, &gene)
	if err != nil {
		h.logger.Error("Failed to add gene", zap.String("entityId", entityID), zap.String("geneName", gene.Name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Gene added successfully", "gene": gene})
}

// ExpressGene 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param entityId path string true "ID"
// @Param geneId path string true "ID"
// @Param request body models.GeneExpressionRequest true ""
// @Success 200 {object} models.GeneExpression ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools/{entityId}/genes/{geneId}/express [post]
func (h *QuantumGeneHandler) ExpressGene(c *gin.Context) {
	entityID := c.Param("entityId")
	geneID := c.Param("geneId")
	
	if entityID == "" || geneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDID"})
		return
	}

	var req models.GeneExpressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid gene expression request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	expression, err := h.geneManager.ExpressGene(c.Request.Context(), geneID, entityID, req.Duration)
	if err != nil {
		h.logger.Error("Failed to express gene", zap.String("entityId", entityID), zap.String("geneId", geneID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expression)
}

// MutateGene 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param entityId path string true "ID"
// @Param geneId path string true "ID"
// @Param request body models.GeneMutationRequest true ""
// @Success 200 {object} models.GeneMutation ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools/{entityId}/genes/{geneId}/mutate [post]
func (h *QuantumGeneHandler) MutateGene(c *gin.Context) {
	entityID := c.Param("entityId")
	geneID := c.Param("geneId")
	
	if entityID == "" || geneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDID"})
		return
	}

	var req models.GeneMutationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid gene mutation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	mutation, err := h.geneManager.MutateGene(c.Request.Context(), geneID)
	if err != nil {
		h.logger.Error("Failed to mutate gene", zap.String("entityId", entityID), zap.String("geneId", geneID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mutation)
}

// AnalyzeInteractions 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Param geneIds query string false "ID"
// @Success 200 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools/{entityId}/interactions [get]
func (h *QuantumGeneHandler) AnalyzeInteractions(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	interactions, err := h.geneManager.AnalyzeInteractions(c.Request.Context(), entityID)
	if err != nil {
		h.logger.Error("Failed to analyze gene interactions", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	result := map[string]interface{}{
		"entityId":     entityID,
		"interactions": interactions,
		"total":        len(interactions),
	}

	c.JSON(http.StatusOK, result)
}

// SimulateEvolution 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param entityId path string true "ID"
// @Param request body models.EvolutionSimulationRequest true ""
// @Success 200 {object} map[string]interface{} ""
// @Failure 400 {object} map[string]interface{} ""
// @Failure 500 {object} map[string]interface{} ""
// @Router /consciousness/genes/pools/{entityId}/simulate [post]
func (h *QuantumGeneHandler) SimulateEvolution(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	var req models.EvolutionSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid evolution simulation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	result, err := h.geneManager.SimulateEvolution(c.Request.Context(), entityID, req.Generations)
	if err != nil {
		h.logger.Error("Failed to simulate evolution", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetGeneStats 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Param entityId path string true "ID"
// @Success 200 {object} models.GenePoolStats "
// @Failure 500 {object} map[string]interface{} "
// @Router /consciousness/genes/pools/{entityId}/stats [get]
func (h *QuantumGeneHandler) GetGeneStats(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	stats := h.geneManager.GetStats()
	if stats == nil {
		h.logger.Error("Failed to get gene stats", zap.String("entityId", entityID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get gene stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetGeneTypes 
// @Summary 
// @Description 
// @Tags 
// @Produce json
// @Success 200 {object} map[string]interface{} ""
// @Router /consciousness/genes/types [get]
func (h *QuantumGeneHandler) GetGeneTypes(c *gin.Context) {
	types := map[string]interface{}{
		"geneTypes": []map[string]interface{}{
			{
				"type":        "consciousness",
				"name":        "意识基因",
				"description": "控制感知和认知能力的基因",
				"categories":  []string{"perception", "cognition", "awareness"},
			},
			{
				"type":        "intelligence",
				"name":        "智能基因",
				"description": "提升学习和推理能力的基因",
				"categories":  []string{"learning", "reasoning", "problem_solving"},
			},
			{
				"type":        "emotion",
				"name":        "情感基因",
				"description": "控制情感表达和共情能力的基因",
				"categories":  []string{"empathy", "expression", "regulation"},
			},
			{
				"type":        "creativity",
				"name":        "创造基因",
				"description": "激发创新和想象力的基因",
				"categories":  []string{"innovation", "imagination", "artistic"},
			},
			{
				"type":        "adaptation",
				"name":        "适应基因",
				"description": "提升适应性和进化能力的基因",
				"categories":  []string{"flexibility", "resilience", "evolution"},
			},
			{
				"type":        "transcendence",
				"name":        "超越基因",
				"description": "实现突破和超越的基因",
				"categories":  []string{"breakthrough", "transcendence", "enlightenment"},
			},
		},
		"categories": map[string]interface{}{
			"core":        "核心基因类别",
			"enhancement": "增强基因类别",
			"special":     "特殊基因类别",
			"rare":        "稀有基因类别",
			"legendary":   "传奇基因类别",
		},
	}

	c.JSON(http.StatusOK, types)
}

// SearchGenes 搜索基因
// @Summary 搜索基因
// @Description 根据条件搜索基因
// @Tags 量子基因
// @Produce json
// @Param entityId path string true "实体ID"
// @Param type query string false "基因类型"
// @Param category query string false "基因类别"
// @Param active query bool false "是否激活" default(false)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "搜索结果"
// @Failure 500 {object} map[string]interface{} "错误信息"
// @Router /consciousness/genes/pools/{entityId}/search [get]
func (h *QuantumGeneHandler) SearchGenes(c *gin.Context) {
	entityID := c.Param("entityId")
	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID"})
		return
	}

	// 
	geneType := c.Query("type")
	category := c.Query("category")
	activeOnly := c.Query("active") == "true"

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

	// 
	pool, err := h.geneManager.GetGenePool(entityID)
	if err != nil {
		h.logger.Error("Failed to get gene pool for search", zap.String("entityId", entityID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "", "details": err.Error()})
		return
	}

	// 
	filteredGenes := []models.QuantumGene{}
	for _, gene := range pool.Genes {
		// 
		if geneType != "" && string(gene.Type) != geneType {
			continue
		}
		// 
		if category != "" && string(gene.Category) != category {
			continue
		}
		// 
		if activeOnly && !gene.IsActive() {
			continue
		}
		filteredGenes = append(filteredGenes, gene)
	}

	// 
	total := len(filteredGenes)
	start := (page - 1) * limit
	end := start + limit
	if start >= total {
		filteredGenes = []models.QuantumGene{}
	} else {
		if end > total {
			end = total
		}
		filteredGenes = filteredGenes[start:end]
	}

	result := map[string]interface{}{
		"entityId": entityID,
		"genes":    filteredGenes,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"hasMore":  (page * limit) < total,
		"filters": map[string]interface{}{
			"type":       geneType,
			"category":   category,
			"activeOnly": activeOnly,
		},
	}

	c.JSON(http.StatusOK, result)
}

