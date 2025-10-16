package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// KnowledgeGraphHandler ?
type KnowledgeGraphHandler struct {
	kgService *knowledge.KnowledgeGraphAppService
}

// NewKnowledgeGraphHandler ?
func NewKnowledgeGraphHandler(kgService *knowledge.KnowledgeGraphAppService) *KnowledgeGraphHandler {
	return &KnowledgeGraphHandler{
		kgService: kgService,
	}
}

// CreateNode 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param node body knowledge.CreateNodeRequest true ""
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

// GetNode 
// @Summary 
// @Description ID
// @Tags knowledge-graph
// @Produce json
// @Param id path string true "ID"
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

// UpdateNode 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param node body knowledge.UpdateNodeRequest true ""
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

	// UpdateNodeRequestmap[string]interface{}
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

// DeleteNode 
// @Summary 
// @Description 㼰
// @Tags knowledge-graph
// @Param id path string true "ID"
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

// CreateRelation 
// @Summary 
// @Description 䴴?
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param relation body knowledge.CreateRelationRequest true ""
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

// GetRelation 
// @Summary 
// @Description ID
// @Tags knowledge-graph
// @Produce json
// @Param id path string true "ID"
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

// UpdateRelation 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param relation body knowledge.UpdateRelationRequest true ""
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

// DeleteRelation 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Param id path string true "ID"
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

// SearchNodes 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Produce json
// @Param q query string false "?
// @Param type query string false ""
// @Param limit query int false "" default(20)
// @Param offset query int false "? default(0)
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

// GetNodeNeighbors 
// @Summary 
// @Description ?
// @Tags knowledge-graph
// @Produce json
// @Param id path string true "ID"
// @Param direction query string false "" Enums(incoming,outgoing,both) default(both)
// @Param limit query int false "" default(20)
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

	// GetNodeNeighbors servicenodesrelationshandlernodes
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

// FindShortestPath ?
// @Summary ?
// @Description ?
// @Tags knowledge-graph
// @Produce json
// @Param fromId query string true "ID"
// @Param toId query string true "ID"
// @Param maxDepth query int false "? default(5)
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

	// FindShortestPath servicenodesrelationshandlernodes
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

// GenerateLearningPath 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param request body knowledge.LearningPathRequest true ""
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

// GenerateConceptMap ?
// @Summary ?
// @Description 
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param request body knowledge.ConceptMapRequest true "?
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

// AnalyzeGraph 
// @Summary 
// @Description 
// @Tags knowledge-graph
// @Accept json
// @Produce json
// @Param request body knowledge.GraphAnalysisRequest true ""
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

// GetGraphStatistics 
// @Summary 
// @Description ?
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

// ValidateGraph 
// @Summary 
// @Description ?
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

// 
// ErrorResponse 
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NodeSearchResponse 
type NodeSearchResponse struct {
	Nodes  []*domainknowledge.NodeResponse `json:"nodes"`
	Total  int                             `json:"total"`
	Limit  int                             `json:"limit"`
	Offset int                             `json:"offset"`
	Query  string                          `json:"query"`
}

// NodeNeighborsResponse 
type NodeNeighborsResponse struct {
	NodeID    uuid.UUID                 `json:"node_id"`
	Neighbors []*knowledge.NodeResponse `json:"neighbors"`
	Direction string                    `json:"direction"`
	Limit     int                       `json:"limit"`
}

// ShortestPathResponse ?
type ShortestPathResponse struct {
	FromID   uuid.UUID                       `json:"from_id"`
	ToID     uuid.UUID                       `json:"to_id"`
	Path     []*domainknowledge.NodeResponse `json:"path"`
	MaxDepth int                             `json:"max_depth"`
}

