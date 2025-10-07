package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
)

// LearnerHandler 学习者处理器
type LearnerHandler struct {
	learnerService *services.LearnerService
}

// NewLearnerHandler 创建新的学习者处理器
func NewLearnerHandler(learnerService *services.LearnerService) *LearnerHandler {
	return &LearnerHandler{
		learnerService: learnerService,
	}
}

// CreateLearner 创建学习者
// @Summary 创建学习者
// @Description 创建新的学习者账户
// @Tags learners
// @Accept json
// @Produce json
// @Param learner body services.CreateLearnerRequest true "学习者信息"
// @Success 201 {object} services.LearnerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/learners [post]
func (h *LearnerHandler) CreateLearner(c *gin.Context) {
	var req services.CreateLearnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	learner, err := h.learnerService.CreateLearner(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create learner",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, learner)
}

// GetLearner 获取学习者
// @Summary 获取学习者信息
// @Description 根据ID获取学习者详细信息
// @Tags learners
// @Produce json
// @Param id path string true "学习者ID"
// @Success 200 {object} services.LearnerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id} [get]
func (h *LearnerHandler) GetLearner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	learner, err := h.learnerService.GetLearner(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Learner not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, learner)
}

// UpdateLearner 更新学习者
// @Summary 更新学习者信息
// @Description 更新学习者的基本信息和偏好设置
// @Tags learners
// @Accept json
// @Produce json
// @Param id path string true "学习者ID"
// @Param learner body services.UpdateLearnerRequest true "更新信息"
// @Success 200 {object} services.LearnerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id} [put]
func (h *LearnerHandler) UpdateLearner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	var req services.UpdateLearnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	learner, err := h.learnerService.UpdateLearner(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update learner",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, learner)
}

// DeleteLearner 删除学习者
// @Summary 删除学习者
// @Description 删除学习者账户及相关数据
// @Tags learners
// @Param id path string true "学习者ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id} [delete]
func (h *LearnerHandler) DeleteLearner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	err = h.learnerService.DeleteLearner(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete learner",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddLearningGoal 添加学习目标
// @Summary 添加学习目标
// @Description 为学习者添加新的学习目标
// @Tags learners
// @Accept json
// @Produce json
// @Param id path string true "学习者ID"
// @Param goal body services.LearningGoalRequest true "学习目标"
// @Success 200 {object} services.LearnerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/goals [post]
func (h *LearnerHandler) AddLearningGoal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	var req services.LearningGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	learner, err := h.learnerService.AddLearningGoal(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to add learning goal",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, learner)
}

// UpdateLearningGoal 更新学习目标
// @Summary 更新学习目标
// @Description 更新学习者的学习目标
// @Tags learners
// @Accept json
// @Produce json
// @Param id path string true "学习者ID"
// @Param goalId path string true "目标ID"
// @Param goal body services.LearningGoalRequest true "学习目标"
// @Success 200 {object} services.LearnerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/goals/{goalId} [put]
func (h *LearnerHandler) UpdateLearningGoal(c *gin.Context) {
	goalIdStr := c.Param("goalId")
	goalId, err := uuid.Parse(goalIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid goal ID",
			Message: err.Error(),
		})
		return
	}

	var req services.LearningGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 将请求转换为更新映射
	updates := map[string]interface{}{
		"description":   req.Description,
		"target_skill":  req.TargetSkill,
		"target_level":  req.TargetLevel,
		"target_date":   req.TargetDate,
		"priority":      req.Priority,
	}

	goal, err := h.learnerService.UpdateLearningGoal(c.Request.Context(), goalId, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update learning goal",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, goal)
}

// UpdateSkill 更新技能
// @Summary 更新技能
// @Description 更新学习者的技能水平
// @Tags learners
// @Accept json
// @Produce json
// @Param id path string true "学习者ID"
// @Param skill body services.SkillRequest true "技能信息"
// @Success 200 {object} services.LearnerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/skills [put]
func (h *LearnerHandler) UpdateSkill(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	var req services.SkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	learner, err := h.learnerService.UpdateSkill(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update skill",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, learner)
}

// RecordLearningActivity 记录学习活动
// @Summary 记录学习活动
// @Description 记录学习者的学习活动和进度
// @Tags learners
// @Accept json
// @Produce json
// @Param id path string true "学习者ID"
// @Param activity body RecordActivityRequest true "学习活动"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/activities [post]
func (h *LearnerHandler) RecordLearningActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	var req RecordActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 构建学习历史记录
	activity := &entities.LearningHistory{
		ID:             uuid.New(),
		LearnerID:      id,
		ContentID:      *req.ContentID, // ContentID is required in LearningHistory
		ContentType:    req.ActivityType,
		Progress:       req.CompletionRate,
		Duration:       time.Duration(req.TimeSpent) * time.Minute,
		Score:          req.Score,
		StartTime:      time.Now(),
		Timestamp:      time.Now(),
	}

	err = h.learnerService.RecordLearningActivity(c.Request.Context(), activity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to record learning activity",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Learning activity recorded successfully",
	})
}

// GetLearningHistory 获取学习历史
// @Summary 获取学习历史
// @Description 获取学习者的学习历史记录
// @Tags learners
// @Produce json
// @Param id path string true "学习者ID"
// @Param limit query int false "限制数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} LearningHistoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/history [get]
func (h *LearnerHandler) GetLearningHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
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

	history, err := h.learnerService.GetLearningHistory(c.Request.Context(), id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get learning history",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, LearningHistoryResponse{
		History: history,
		Limit:   limit,
		Offset:  offset,
	})
}

// GetLearningAnalytics 获取学习分析报告
// @Summary 获取学习分析报告
// @Description 获取学习者的详细学习分析报告
// @Tags learners
// @Produce json
// @Param id path string true "学习者ID"
// @Param timeRange query string false "时间范围" Enums(week,month,quarter,year) default(month)
// @Success 200 {object} object
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/analytics [get]
func (h *LearnerHandler) GetLearningAnalytics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	timeRange := c.DefaultQuery("timeRange", "month")

	report, err := h.learnerService.GetLearningAnalytics(c.Request.Context(), id, timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get learning analytics",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetPersonalizedRecommendations 获取个性化推荐
// @Summary 获取个性化推荐
// @Description 获取基于学习者状态的个性化推荐
// @Tags learners
// @Produce json
// @Param id path string true "学习者ID"
// @Success 200 {object} services.PersonalizedRecommendations
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/learners/{id}/recommendations [get]
func (h *LearnerHandler) GetPersonalizedRecommendations(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: err.Error(),
		})
		return
	}

	recommendations, err := h.learnerService.GetPersonalizedRecommendations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get recommendations",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recommendations)
}

// 请求和响应结构体

// RecordActivityRequest 记录活动请求
type RecordActivityRequest struct {
	ContentID         *uuid.UUID `json:"content_id,omitempty"`
	KnowledgeNodeID   *uuid.UUID `json:"knowledge_node_id,omitempty"`
	ActivityType      string     `json:"activity_type" binding:"required"`
	CompletionRate    float64    `json:"completion_rate"`
	TimeSpent         int        `json:"time_spent"`
	Score             *float64   `json:"score,omitempty"`
	Notes             string     `json:"notes,omitempty"`
}

// LearningHistoryResponse 学习历史响应
type LearningHistoryResponse struct {
	History []*entities.LearningHistory `json:"history"`
	Limit   int                         `json:"limit"`
	Offset  int                         `json:"offset"`
}

// ErrorResponse 错误响应
// SuccessResponse 成功响应