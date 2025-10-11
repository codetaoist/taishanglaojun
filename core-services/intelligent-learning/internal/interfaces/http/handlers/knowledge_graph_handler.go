package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// KnowledgeGraphHandler 知识图谱处理�?
type KnowledgeGraphHandler struct {
	kgService *knowledge.KnowledgeGraphAppService
}

// NewKnowledgeGraphHandler 创建新的知识图谱处理�?
func NewKnowledgeGraphHandler(kgService *knowledge.KnowledgeGraphAppService) *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{
		kgService: kgService,
	}
}

// CreateNode 创建知识节点
// @Summary 创建知识节点
// @Description 在知识图谱中创建新的知识节点
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param node body knowledge.CreateNodeRequest true "节点信息"
// @Success 201 {object} knowledge.NodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/nodes [post]
func (h *KnowledgeGraphHandler) CreateNode(c *gin.Context) {
	var req knowledge.CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	node, err := h.kgService.CreateNode(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create node",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, node)
}

// GetNode 获取知识节点
// @Summary 获取知识节点
// @Description 根据ID获取知识节点详细信息
// @Tags knowledge-graph
// @Produce json
// @Param id path string true "节点ID"
// @Success 200 {object} knowledge.NodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/nodes/{id} [get]
func (h *KnowledgeGraphHandler) GetNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid node ID",
			Message: err.Error(),
		})
		return
	}

	node, err := h.kgService.GetNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Node not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, node)
}

// UpdateNode 更新知识节点
// @Summary 更新知识节点
// @Description 更新知识节点信息
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param id path string true "节点ID"
// @Param node body knowledge.UpdateNodeRequest true "更新信息"
// @Success 200 {object} knowledge.NodeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/nodes/{id} [put]
func (h *KnowledgeGraphHandler) UpdateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid node ID",
			Message: err.Error(),
		})
		return
	}

	var req knowledge.UpdateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 将UpdateNodeRequest转换为map[string]interface{}
	updates := make(map[string]interface{})
	if req.Title != nil && *req.Title != "" {
		updates["title"] = *req.Title
	}
	if req.Content != nil && *req.Content != "" {
		updates["content"] = *req.Content
	}
	if req.Type != nil && *req.Type != "" {
		updates["type"] = *req.Type
	}
	if req.Difficulty != nil && *req.Difficulty != "" {
		updates["difficulty"] = *req.Difficulty
	}
	if len(req.Tags) > 0 {
		updates["tags"] = req.Tags
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}

	node, err := h.kgService.UpdateNode(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update node",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, node)
}

// DeleteNode 删除知识节点
// @Summary 删除知识节点
// @Description 删除知识节点及其相关关系
// @Tags knowledge-graph
// @Param id path string true "节点ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/nodes/{id} [delete]
func (h *KnowledgeGraphHandler) DeleteNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid node ID",
			Message: err.Error(),
		})
		return
	}

	err = h.kgService.DeleteNode(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete node",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateRelation 创建知识关系
// @Summary 创建知识关系
// @Description 在知识节点之间创建关�?
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param relation body knowledge.CreateRelationRequest true "关系信息"
// @Success 201 {object} knowledge.RelationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/relations [post]
func (h *KnowledgeGraphHandler) CreateRelation(c *gin.Context) {
	var req knowledge.CreateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	relation, err := h.kgService.CreateRelation(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create relation",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, relation)
}

// GetRelation 获取知识关系
// @Summary 获取知识关系
// @Description 根据ID获取知识关系详细信息
// @Tags knowledge-graph
// @Produce json
// @Param id path string true "关系ID"
// @Success 200 {object} knowledge.RelationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/relations/{id} [get]
func (h *KnowledgeGraphHandler) GetRelation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid relation ID",
			Message: err.Error(),
		})
		return
	}

	relation, err := h.kgService.GetRelation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Relation not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, relation)
}

// UpdateRelation 更新知识关系
// @Summary 更新知识关系
// @Description 更新知识关系信息
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param id path string true "关系ID"
// @Param relation body knowledge.UpdateRelationRequest true "更新信息"
// @Success 200 {object} knowledge.RelationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/relations/{id} [put]
func (h *KnowledgeGraphHandler) UpdateRelation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid relation ID",
			Message: err.Error(),
		})
		return
	}

	var req knowledge.UpdateRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	relation, err := h.kgService.UpdateRelation(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update relation",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, relation)
}

// DeleteRelation 删除知识关系
// @Summary 删除知识关系
// @Description 删除知识关系
// @Tags knowledge-graph
// @Param id path string true "关系ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/relations/{id} [delete]
func (h *KnowledgeGraphHandler) DeleteRelation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid relation ID",
			Message: err.Error(),
		})
		return
	}

	err = h.kgService.DeleteRelation(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete relation",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// SearchNodes 搜索知识节点
// @Summary 搜索知识节点
// @Description 根据条件搜索知识节点
// @Tags knowledge-graph
// @Produce json
// @Param q query string false "搜索关键�?
// @Param type query string false "节点类型"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移�? default(0)
// @Success 200 {object} NodeSearchResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/nodes/search [get]
func (h *KnowledgeGraphHandler) SearchNodes(c *gin.Context) {
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	req := &knowledge.GraphSearchRequest{
		Query:     c.Query("q"),
		NodeTypes: []string{c.Query("type")},
		Limit:     limit,
		Offset:    offset,
	}

	results, err := h.kgService.SearchNodes(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to search nodes",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, NodeSearchResponse{
		Nodes:  results.Nodes,
		Total:  results.Total,
		Limit:  limit,
		Offset: offset,
		Query:  req.Query,
	})
}

// GetNodeNeighbors 获取节点邻居
// @Summary 获取节点邻居
// @Description 获取知识节点的相邻节�?
// @Tags knowledge-graph
// @Produce json
// @Param id path string true "节点ID"
// @Param direction query string false "方向" Enums(incoming,outgoing,both) default(both)
// @Param limit query int false "限制数量" default(20)
// @Success 200 {object} NodeNeighborsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/nodes/{id}/neighbors [get]
func (h *KnowledgeGraphHandler) GetNodeNeighbors(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid node ID",
			Message: err.Error(),
		})
		return
	}

	direction := c.DefaultQuery("direction", "both")
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// GetNodeNeighbors service方法返回nodes和relations，但handler只需要nodes
	nodes, _, err := h.kgService.GetNodeNeighbors(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get node neighbors",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, NodeNeighborsResponse{
		NodeID:    id,
		Neighbors: nodes,
		Direction: direction,
		Limit:     limit,
	})
}

// FindShortestPath 查找最短路�?
// @Summary 查找最短路�?
// @Description 查找两个知识节点之间的最短路�?
// @Tags knowledge-graph
// @Produce json
// @Param fromId query string true "起始节点ID"
// @Param toId query string true "目标节点ID"
// @Param maxDepth query int false "最大深�? default(5)
// @Success 200 {object} ShortestPathResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/shortest-path [get]
func (h *KnowledgeGraphHandler) FindShortestPath(c *gin.Context) {
	fromIdStr := c.Query("fromId")
	if fromIdStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing fromId parameter",
			Message: "fromId is required",
		})
		return
	}

	toIdStr := c.Query("toId")
	if toIdStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing toId parameter",
			Message: "toId is required",
		})
		return
	}

	fromId, err := uuid.Parse(fromIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid fromId",
			Message: err.Error(),
		})
		return
	}

	toId, err := uuid.Parse(toIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid toId",
			Message: err.Error(),
		})
		return
	}

	maxDepth := 5
	if maxDepthStr := c.Query("maxDepth"); maxDepthStr != "" {
		if d, err := strconv.Atoi(maxDepthStr); err == nil && d > 0 {
			maxDepth = d
		}
	}

	// FindShortestPath service方法返回nodes和relations，但handler只需要nodes
	pathNodes, _, err := h.kgService.FindShortestPath(c.Request.Context(), fromId, toId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to find shortest path",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ShortestPathResponse{
		FromID:   fromId,
		ToID:     toId,
		Path:     pathNodes,
		MaxDepth: maxDepth,
	})
}

// GenerateLearningPath 生成学习路径
// @Summary 生成学习路径
// @Description 为学习者生成个性化学习路径
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param request body knowledge.LearningPathRequest true "学习路径请求"
// @Success 200 {object} knowledge.LearningPathResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/learning-path [post]
func (h *KnowledgeGraphHandler) GenerateLearningPath(c *gin.Context) {
	var req knowledge.LearningPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	path, err := h.kgService.GenerateLearningPath(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate learning path",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, path)
}

// GenerateConceptMap 生成概念�?
// @Summary 生成概念�?
// @Description 为特定主题生成概念图
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param request body knowledge.ConceptMapRequest true "概念图请�?
// @Success 200 {object} knowledge.ConceptMapResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/concept-map [post]
func (h *KnowledgeGraphHandler) GenerateConceptMap(c *gin.Context) {
	var req knowledge.ConceptMapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	conceptMap, err := h.kgService.GenerateConceptMap(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate concept map",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, conceptMap)
}

// AnalyzeGraph 分析图谱
// @Summary 分析知识图谱
// @Description 分析知识图谱的结构和特征
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param request body knowledge.GraphAnalysisRequest true "分析请求"
// @Success 200 {object} object
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/analyze [post]
func (h *KnowledgeGraphHandler) AnalyzeGraph(c *gin.Context) {
	var req domainknowledge.GraphAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	analysis, err := h.kgService.AnalyzeGraph(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to analyze graph",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetGraphStatistics 获取图谱统计
// @Summary 获取图谱统计
// @Description 获取知识图谱的统计信�?
// @Tags knowledge-graph
// @Produce json
// @Success 200 {object} object
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/statistics [get]
func (h *KnowledgeGraphHandler) GetGraphStatistics(c *gin.Context) {
	stats, err := h.kgService.GetGraphStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get graph statistics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ValidateGraph 验证图谱
// @Summary 验证知识图谱
// @Description 验证知识图谱的完整性和一致�?
// @Tags knowledge-graph
// @Produce json
// @Success 200 {object} object
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/knowledge-graph/validate [post]
func (h *KnowledgeGraphHandler) ValidateGraph(c *gin.Context) {
	validation, err := h.kgService.ValidateGraph(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to validate graph",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, validation)
}

// 响应结构
// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NodeSearchResponse 节点搜索响应
type NodeSearchResponse struct {
	Nodes  []*domainknowledge.NodeResponse `json:"nodes"`
	Total  int                             `json:"total"`
	Limit  int                             `json:"limit"`
	Offset int                             `json:"offset"`
	Query  string                          `json:"query"`
}

// NodeNeighborsResponse 节点邻居响应
type NodeNeighborsResponse struct {
	NodeID    uuid.UUID                 `json:"node_id"`
	Neighbors []*knowledge.NodeResponse `json:"neighbors"`
	Direction string                    `json:"direction"`
	Limit     int                       `json:"limit"`
}

// ShortestPathResponse 最短路径响�?
type ShortestPathResponse struct {
	FromID   uuid.UUID                       `json:"from_id"`
	ToID     uuid.UUID                       `json:"to_id"`
	Path     []*domainknowledge.NodeResponse `json:"path"`
	MaxDepth int                             `json:"max_depth"`
}
