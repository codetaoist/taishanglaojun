package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/adaptive"
)

// LearningPathHandler 学习路径处理
// @Summary 学习路径处理
// @Description 处理与学习路径相关的HTTP请求
// @Tags learning-paths
// @Accept json
// @Produce json
// @Param request body adaptive.GeneratePersonalizedPathRequest true "路径生成请求"
// @Success 201 {object} adaptive.LearningPathResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/learning-paths/generate [post]
type LearningPathHandler struct {
	pathService *adaptive.LearningPathService
}

// NewLearningPathHandler 创建新的学习路径处理
// @Summary 创建新的学习路径处理
// @Description 创建一个新的学习路径处理实?
// @Tags learning-paths
// @Accept json
// @Produce json
// @Param pathService body adaptive.LearningPathService true "学习路径服务"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
func NewLearningPathHandler(pathService *adaptive.LearningPathService) *LearningPathHandler {
	return &LearningPathHandler{
		pathService: pathService,
	}
}

// GeneratePersonalizedPath 生成个性化学习路径
// @Summary 生成个性化学习路径
// @Description 基于学习者特征和目标生成个性化学习路径
// @Tags learning-paths
// @Accept json
// @Produce json
// @Param request body adaptive.GeneratePersonalizedPathRequest true "路径生成请求"
// @Success 201 {object} adaptive.LearningPathResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/learning-paths/generate [post]
func (h *LearningPathHandler) GeneratePersonalizedPath(c *gin.Context) {
	var req adaptive.GeneratePersonalizedPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	path, err := h.pathService.GeneratePersonalizedPath(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to generate learning path",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, path)
}

// GetLearningPath 获取学习路径
// @Summary 获取学习路径详情
// @Description 根据ID获取学习路径的详细信?
// @Tags learning-paths
// @Produce json
// @Param id path string true "学习路径ID"
// @Success 200 {object} adaptive.LearningPathResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learning-paths/{id} [get]
func (h *LearningPathHandler) GetLearningPath(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid path ID",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现获取路径的逻辑
	// path, err := h.pathService.GetLearningPath(c.Request.Context(), id)
	// 暂时返回模拟响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Learning path retrieved",
		"path_id": id,
	})
}

// UpdatePathProgress 更新学习路径进度
// @Summary 更新学习路径进度
// @Description 更新学习者在特定路径步骤的学习进?
// @Tags learning-paths
// @Accept json
// @Produce json
// @Param request body adaptive.UpdatePathProgressRequest true "进度更新请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/learning-paths/progress [put]
func (h *LearningPathHandler) UpdatePathProgress(c *gin.Context) {
	var req adaptive.UpdatePathProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	err := h.pathService.UpdatePathProgress(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update progress",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Progress updated successfully",
		"path_id":  req.PathID,
		"step_id":  req.StepID,
		"progress": req.Progress,
	})
}

// GetRecommendedPaths 获取推荐学习路径
// @Summary 获取推荐学习路径
// @Description 基于学习者特征推荐合适的学习路径
// @Tags learning-paths
// @Accept json
// @Produce json
// @Param request body adaptive.PathRecommendationRequest true "路径推荐请求"
// @Success 200 {object} adaptive.PathRecommendationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/learning-paths/recommendations [post]
func (h *LearningPathHandler) GetRecommendedPaths(c *gin.Context) {
	var req adaptive.PathRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	recommendations, err := h.pathService.GetRecommendedPaths(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get recommendations",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recommendations)
}

// GetLearnerPaths 获取学习者的所有学习路?
// @Summary 获取学习者的学习路径列表
// @Description 获取指定学习者的所有学习路径列?
// @Tags learning-paths
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param status query string false "路径状态过? Enums(active, completed, paused)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/learning-paths/learner/{learner_id} [get]
func (h *LearningPathHandler) GetLearnerPaths(c *gin.Context) {
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 这里需要实现获取学习者路径列表的逻辑
	// paths, total, err := h.pathService.GetLearnerPaths(c.Request.Context(), learnerID, status, page, limit)

	c.JSON(http.StatusOK, gin.H{
		"learner_id": learnerID,
		"status":     status,
		"page":       page,
		"limit":      limit,
		"paths":      []interface{}{},
		"total":      0,
		"message":    "Learner paths retrieved",
	})
}

// DeleteLearningPath 删除学习路径
// @Summary 删除学习路径
// @Description 删除指定的学习路?
// @Tags learning-paths
// @Param id path string true "学习路径ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learning-paths/{id} [delete]
func (h *LearningPathHandler) DeleteLearningPath(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid path ID",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现删除路径的逻辑
	// err = h.pathService.DeleteLearningPath(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Learning path deleted",
		"path_id": id,
	})
}

// PauseLearningPath 暂停学习路径
// @Summary 暂停学习路径
// @Description 暂停学习者的学习路径
// @Tags learning-paths
// @Param id path string true "学习路径ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learning-paths/{id}/pause [post]
func (h *LearningPathHandler) PauseLearningPath(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid path ID",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现暂停路径的逻辑
	// err = h.pathService.PauseLearningPath(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Learning path paused",
		"path_id": id,
		"status":  "paused",
	})
}

// ResumeLearningPath 恢复学习路径
// @Summary 恢复学习路径
// @Description 恢复学习者的学习路径
// @Tags learning-paths
// @Param id path string true "学习路径ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learning-paths/{id}/resume [post]
func (h *LearningPathHandler) ResumeLearningPath(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid path ID",
			Message: err.Error(),
		})
		return
	}

	// 这里需要实现恢复路径的逻辑
	// err = h.pathService.ResumeLearningPath(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Learning path resumed",
		"path_id": id,
		"status":  "active",
	})
}

