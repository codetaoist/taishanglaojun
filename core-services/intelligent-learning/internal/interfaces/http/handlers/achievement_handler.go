package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
)

// AchievementHandler 成就处理
// @Summary 成就处理
// @Description 处理与学习成就相关的HTTP请求
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.CheckAchievementsRequest true "检查成就请包含学习者ID和学习事件"
// @Success 200 {object} analytics.CheckAchievementsResponse "成就检查结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/achievements/check [post]
type AchievementHandler struct {
	achievementService *analytics.LearningAchievementService
}

// NewAchievementHandler 创建新的成就处理
// @Summary 创建新的成就处理
// @Description 创建一个新的成就处理
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.CheckAchievementsRequest true "检查成就请包含学习者ID和学习事件"
// @Success 200 {object} analytics.CheckAchievementsResponse "成就检查结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/achievements/check [post]
func NewAchievementHandler(achievementService *analytics.LearningAchievementService) *AchievementHandler {
	return &AchievementHandler{
		achievementService: achievementService,
	}
}

// CheckAchievements 检查并更新学习者成就
// @Summary 检查并更新学习者成就
// @Description 根据学习事件检查并更新学习者的成就状态
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.CheckAchievementsRequest true "检查成就请包含学习者ID和学习事件"
// @Success 200 {object} analytics.CheckAchievementsResponse "成就检查结果"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/achievements/check [post]
func (h *AchievementHandler) CheckAchievements(c *gin.Context) {
	var req analytics.CheckAchievementsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	response, err := h.achievementService.CheckAchievements(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to check achievements",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLearnerAchievements 获取学习者成就列表
// @Summary 获取学习者成就列表
// @Description 获取指定学习者的成就列表，支持筛选和分页
// @Tags achievements
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param type query string false "成就类型" Enums(progress,streak,skill,milestone,time,quality,social,challenge)
// @Param level query string false "成就等级" Enums(bronze,silver,gold,platinum,diamond)
// @Param status query string false "成就状态" Enums(unlocked,in_progress,locked)
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} analytics.GetAchievementsResponse "成就列表"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/learners/{learner_id}/achievements [get]
func (h *AchievementHandler) GetLearnerAchievements(c *gin.Context) {
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: "Learner ID must be a valid UUID",
		})
		return
	}

	req := analytics.GetAchievementsRequest{
		LearnerID: learnerID,
	}

	// 解析查询参数
	if typeStr := c.Query("type"); typeStr != "" {
		achievementType := analytics.AchievementType(typeStr)
		req.Type = &achievementType
	}

	if levelStr := c.Query("level"); levelStr != "" {
		achievementLevel := analytics.AchievementLevel(levelStr)
		req.Level = &achievementLevel
	}

	if status := c.Query("status"); status != "" {
		req.Status = status
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	response, err := h.achievementService.GetLearnerAchievements(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get learner achievements",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAchievementSummary 获取学习者成就摘要
// @Summary 获取学习者成就摘要
// @Description 获取学习者的成就统计摘要信息
// @Tags achievements
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Success 200 {object} analytics.AchievementSummary "成就摘要"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/learners/{learner_id}/achievements/summary [get]
func (h *AchievementHandler) GetAchievementSummary(c *gin.Context) {
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: "Learner ID must be a valid UUID",
		})
		return
	}

	// 获取成就列表（只需要摘要信息）
	req := analytics.GetAchievementsRequest{
		LearnerID: learnerID,
		Limit:     1, // 只需要摘要，不需要具体成就列表
	}

	response, err := h.achievementService.GetLearnerAchievements(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get achievement summary",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Summary)
}

// CreateAchievement 创建新成就
// @Summary 创建新成就
// @Description 创建新的成就定义（管理员功能）
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.Achievement true "成就信息"
// @Success 201 {object} SuccessResponse "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/achievements [post]
func (h *AchievementHandler) CreateAchievement(c *gin.Context) {
	var achievement analytics.Achievement
	if err := c.ShouldBindJSON(&achievement); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	if err := h.achievementService.CreateAchievement(c.Request.Context(), &achievement); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create achievement",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Achievement created successfully",
		Data: map[string]interface{}{
			"achievement_id": achievement.ID,
		},
	})
}

// GetAchievementLeaderboard 获取成就排行榜
// @Summary 获取成就排行榜
// @Description 获取学习者成就积分排行榜
// @Tags achievements
// @Accept json
// @Produce json
// @Param type query string false "排行榜类型" Enums(points,achievements,recent) default(points)
// @Param period query string false "时间周期" Enums(daily,weekly,monthly,all) default(all)
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} LeaderboardResponse "排行榜数据"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/achievements/leaderboard [get]
func (h *AchievementHandler) GetAchievementLeaderboard(c *gin.Context) {
	leaderboardType := c.DefaultQuery("type", "points")
	period := c.DefaultQuery("period", "all")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 这里应该调用服务层方法获取排行榜数据
	// 为了演示，返回模拟数据
	leaderboard := LeaderboardResponse{
		Type:   leaderboardType,
		Period: period,
		Rankings: []LeaderboardEntry{
			{
				Rank:         1,
				LearnerID:    uuid.New(),
				LearnerName:  "张三",
				Score:        1500,
				Achievements: 25,
			},
			{
				Rank:         2,
				LearnerID:    uuid.New(),
				LearnerName:  "李四",
				Score:        1200,
				Achievements: 20,
			},
		},
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetAchievementProgress 获取成就进度详情
// @Summary 获取成就进度详情
// @Description 获取学习者特定成就的详细进度信息
// @Tags achievements
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param achievement_id path string true "成就ID"
// @Success 200 {object} analytics.LearnerAchievement "成就进度详情"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "成就不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/v1/learners/{learner_id}/achievements/{achievement_id} [get]
func (h *AchievementHandler) GetAchievementProgress(c *gin.Context) {
	learnerIDStr := c.Param("learner_id")
	achievementIDStr := c.Param("achievement_id")

	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid learner ID",
			Message: "Learner ID must be a valid UUID",
		})
		return
	}

	achievementID, err := uuid.Parse(achievementIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid achievement ID",
			Message: "Achievement ID must be a valid UUID",
		})
		return
	}

	// 获取学习者的所有成就进度
	req := analytics.GetAchievementsRequest{
		LearnerID: learnerID,
	}

	response, err := h.achievementService.GetLearnerAchievements(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get achievement progress",
			Message: err.Error(),
		})
		return
	}

	// 查找特定成就
	for _, achievement := range response.Achievements {
		if achievement.AchievementID == achievementID {
			c.JSON(http.StatusOK, achievement)
			return
		}
	}

	c.JSON(http.StatusNotFound, ErrorResponse{
		Error:   "Achievement not found",
		Message: "The specified achievement was not found for this learner",
	})
}

// 响应结构
// LeaderboardResponse 排行榜响应
type LeaderboardResponse struct {
	Type      string             `json:"type"`
	Period    string             `json:"period"`
	Rankings  []LeaderboardEntry `json:"rankings"`
	UpdatedAt string             `json:"updated_at"`
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	LearnerID    uuid.UUID `json:"learner_id"`
	LearnerName  string    `json:"learner_name"`
	Score        int       `json:"score"`
	Achievements int       `json:"achievements"`
	Avatar       string    `json:"avatar,omitempty"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
