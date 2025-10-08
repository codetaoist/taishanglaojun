package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services/content"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
)

// ContentHandler 内容处理器
type ContentHandler struct {
	contentService *content.ContentService
}

// NewContentHandler 创建新的内容处理器
func NewContentHandler(contentService *content.ContentService) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
	}
}

// CreateContent 创建内容
// @Summary 创建学习内容
// @Description 创建新的学习内容
// @Tags content
// @Accept json
// @Produce json
// @Param content body content.CreateContentRequest true "内容信息"
// @Success 201 {object} content.ContentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/content [post]
func (h *ContentHandler) CreateContent(c *gin.Context) {
	var req content.CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	content, err := h.contentService.CreateContent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, content)
}

// GetContent 获取内容
// @Summary 获取学习内容
// @Description 根据ID获取学习内容详细信息
// @Tags content
// @Produce json
// @Param id path string true "内容ID"
// @Success 200 {object} content.ContentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id} [get]
func (h *ContentHandler) GetContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	content, err := h.contentService.GetContent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Content not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, content)
}

// UpdateContent 更新内容
// @Summary 更新学习内容
// @Description 更新学习内容信息
// @Tags content
// @Accept json
// @Produce json
// @Param id path string true "内容ID"
// @Param content body content.UpdateContentRequest true "更新信息"
// @Success 200 {object} content.ContentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id} [put]
func (h *ContentHandler) UpdateContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	var req content.UpdateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	content, err := h.contentService.UpdateContent(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, content)
}

// DeleteContent 删除内容
// @Summary 删除学习内容
// @Description 删除学习内容
// @Tags content
// @Param id path string true "内容ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id} [delete]
func (h *ContentHandler) DeleteContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	err = h.contentService.DeleteContent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete content",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListContent 列出内容
// @Summary 列出学习内容
// @Description 分页列出学习内容
// @Tags content
// @Produce json
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Param type query string false "内容类型"
// @Param difficulty query string false "难度级别"
// @Param status query string false "状态"
// @Success 200 {object} ContentListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/content [get]
func (h *ContentHandler) ListContent(c *gin.Context) {
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

	contentType := c.Query("type")
	difficulty := c.Query("difficulty")
	status := c.Query("status")

	contents, err := h.contentService.ListContent(c.Request.Context(), limit, offset, contentType, difficulty, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ContentListResponse{
		Contents: contents,
		Limit:    limit,
		Offset:   offset,
	})
}

// PublishContent 发布内容
// @Summary 发布学习内容
// @Description 发布学习内容使其可用
// @Tags content
// @Param id path string true "内容ID"
// @Success 200 {object} services.ContentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/publish [post]
func (h *ContentHandler) PublishContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	content, err := h.contentService.PublishContent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to publish content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, content)
}

// ArchiveContent 归档内容
// @Summary 归档学习内容
// @Description 归档学习内容使其不可用
// @Tags content
// @Param id path string true "内容ID"
// @Success 200 {object} services.ContentResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/archive [post]
func (h *ContentHandler) ArchiveContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	content, err := h.contentService.ArchiveContent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to archive content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, content)
}

// SearchContent 搜索内容
// @Summary 搜索学习内容
// @Description 全文搜索学习内容
// @Tags content
// @Produce json
// @Param q query string true "搜索关键词"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Param type query string false "内容类型"
// @Param difficulty query string false "难度级别"
// @Success 200 {object} ContentSearchResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/content/search [get]
func (h *ContentHandler) SearchContent(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing search query",
			Message: "Query parameter 'q' is required",
		})
		return
	}

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

	req := &content.ContentSearchRequest{
		Query:      query,
		Limit:      limit,
		Offset:     offset,
		Type:       c.Query("type"),
		Difficulty: c.Query("difficulty"),
	}

	results, err := h.contentService.SearchContent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to search content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ContentSearchResponse{
		Results: results.Contents,
		Total:   results.Total,
		Limit:   limit,
		Offset:  offset,
		Query:   query,
	})
}

// GetPersonalizedContent 获取个性化内容推荐
// @Summary 获取个性化内容推荐
// @Description 根据学习者偏好获取个性化内容推荐
// @Tags content
// @Produce json
// @Param learnerId query string true "学习者ID"
// @Param limit query int false "限制数量" default(10)
// @Success 200 {object} PersonalizedContentResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/content/personalized [get]
func (h *ContentHandler) GetPersonalizedContent(c *gin.Context) {
	learnerIdStr := c.Query("learnerId")
	if learnerIdStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing learner ID",
			Message: "Query parameter 'learnerId' is required",
		})
		return
	}

	learnerId, err := uuid.Parse(learnerIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	req := &content.PersonalizedContentRequest{
		LearnerID:          learnerId,
		MaxRecommendations: limit,
	}

	recommendations, err := h.contentService.GetPersonalizedContent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get personalized content",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PersonalizedContentResponse{
		Recommendations: recommendations,
		LearnerID:       learnerId,
		Limit:           limit,
	})
}

// RecordContentInteraction 记录内容交互
// @Summary 记录内容交互
// @Description 记录学习者与内容的交互行为
// @Tags content
// @Accept json
// @Produce json
// @Param id path string true "内容ID"
// @Param interaction body ContentInteractionRequest true "交互信息"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/content/{id}/interactions [post]
func (h *ContentHandler) RecordContentInteraction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	var req ContentInteractionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	interaction := &entities.InteractionRecord{
		ID:        uuid.New(),
		LearnerID: req.LearnerID,
		ContentID: id,
		Type:      req.InteractionType,
		Position:  req.Duration,
		Timestamp: time.Now(),
	}

	if err := h.contentService.RecordContentInteraction(c.Request.Context(), interaction); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to record interaction",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Content interaction recorded successfully",
	})
}

// GetContentByKnowledgeNode 根据知识节点获取内容
// @Summary 根据知识节点获取内容
// @Description 获取与特定知识节点相关的学习内容
// @Tags content
// @Produce json
// @Param nodeId path string true "知识节点ID"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} ContentListResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/content/knowledge-node/{nodeId} [get]
func (h *ContentHandler) GetContentByKnowledgeNode(c *gin.Context) {
	nodeIdStr := c.Param("nodeId")
	nodeId, err := uuid.Parse(nodeIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid knowledge node ID",
			Message: err.Error(),
		})
		return
	}

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

	contents, err := h.contentService.GetContentsByKnowledgeNode(c.Request.Context(), nodeId, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get content by knowledge node",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ContentListResponse{
		Contents: contents,
		Limit:    limit,
		Offset:   offset,
	})
}

// GetContentAnalytics 获取内容分析
// @Summary 获取内容分析
// @Description 获取内容的使用统计和分析数据
// @Tags content
// @Produce json
// @Param id path string true "内容ID"
// @Success 200 {object} object
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/content/{id}/analytics [get]
func (h *ContentHandler) GetContentAnalytics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid content ID",
			Message: err.Error(),
		})
		return
	}

	analytics, err := h.contentService.GetContentAnalytics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get content analytics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// 请求和响应结构体

// ContentListResponse 内容列表响应
type ContentListResponse struct {
	Contents []*content.ContentResponse `json:"contents"`
	Limit    int                        `json:"limit"`
	Offset   int                        `json:"offset"`
}

// ContentSearchResponse 内容搜索响应
type ContentSearchResponse struct {
	Results []*content.ContentResponse `json:"results"`
	Total   int                        `json:"total"`
	Limit   int                        `json:"limit"`
	Offset  int                        `json:"offset"`
	Query   string                     `json:"query"`
}

// PersonalizedContentResponse 个性化内容响应
type PersonalizedContentResponse struct {
	Recommendations []*content.SimpleContentRecommendation `json:"recommendations"`
	LearnerID       uuid.UUID                              `json:"learner_id"`
	Limit           int                                    `json:"limit"`
}

// ContentInteractionRequest 内容交互请求
type ContentInteractionRequest struct {
	LearnerID       uuid.UUID `json:"learner_id" binding:"required"`
	InteractionType string    `json:"interaction_type" binding:"required"`
	Duration        int       `json:"duration"`
}