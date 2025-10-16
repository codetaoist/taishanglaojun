package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
)

// AchievementHandler 
// @Summary 
// @Description HTTP
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.CheckAchievementsRequest true "ID?
// @Success 200 {object} analytics.CheckAchievementsResponse "?
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/achievements/check [post]
type AchievementHandler struct {
	achievementService *analytics.LearningAchievementService
}

// NewAchievementHandler 
// @Summary 
// @Description ?
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.CheckAchievementsRequest true "ID?
// @Success 200 {object} analytics.CheckAchievementsResponse "?
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/achievements/check [post]
func NewAchievementHandler(achievementService *analytics.LearningAchievementService) *AchievementHandler {
	return &AchievementHandler{
		achievementService: achievementService,
	}
}

// CheckAchievements 鲢?
// @Summary 鲢?
// @Description 鲢?
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.CheckAchievementsRequest true "ID?
// @Success 200 {object} analytics.CheckAchievementsResponse "?
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

// GetLearnerAchievements ?
// @Summary ?
// @Description 
// @Tags achievements
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Param type query string false "" Enums(progress,streak,skill,milestone,time,quality,social,challenge)
// @Param level query string false "" Enums(bronze,silver,gold,platinum,diamond)
// @Param status query string false "? Enums(unlocked,in_progress,locked)
// @Param page query int false "" default(1)
// @Param limit query int false "" default(20)
// @Success 200 {object} analytics.GetAchievementsResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

	// 
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

// GetAchievementSummary ?
// @Summary ?
// @Description 
// @Tags achievements
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Success 200 {object} analytics.AchievementSummary ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

	// 
	req := analytics.GetAchievementsRequest{
		LearnerID: learnerID,
		Limit:     1, // ?
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

// CreateAchievement ?
// @Summary ?
// @Description ?
// @Tags achievements
// @Accept json
// @Produce json
// @Param request body analytics.Achievement true ""
// @Success 201 {object} SuccessResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
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

// GetAchievementLeaderboard ?
// @Summary ?
// @Description 
// @Tags achievements
// @Accept json
// @Produce json
// @Param type query string false "? Enums(points,achievements,recent) default(points)
// @Param period query string false "" Enums(daily,weekly,monthly,all) default(all)
// @Param limit query int false "" default(10)
// @Success 200 {object} LeaderboardResponse "?
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/achievements/leaderboard [get]
func (h *AchievementHandler) GetAchievementLeaderboard(c *gin.Context) {
	leaderboardType := c.DefaultQuery("type", "points")
	period := c.DefaultQuery("period", "all")
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 㷽
	// ?
	leaderboard := LeaderboardResponse{
		Type:   leaderboardType,
		Period: period,
		Rankings: []LeaderboardEntry{
			{
				Rank:         1,
				LearnerID:    uuid.New(),
				LearnerName:  "",
				Score:        1500,
				Achievements: 25,
			},
			{
				Rank:         2,
				LearnerID:    uuid.New(),
				LearnerName:  "",
				Score:        1200,
				Achievements: 20,
			},
		},
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetAchievementProgress 
// @Summary 
// @Description 
// @Tags achievements
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Param achievement_id path string true "ID"
// @Success 200 {object} analytics.LearnerAchievement ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
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

	// ?
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

	// 
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

// 
// LeaderboardResponse ?
type LeaderboardResponse struct {
	Type      string             `json:"type"`
	Period    string             `json:"period"`
	Rankings  []LeaderboardEntry `json:"rankings"`
	UpdatedAt string             `json:"updated_at"`
}

// LeaderboardEntry ?
type LeaderboardEntry struct {
	Rank         int       `json:"rank"`
	LearnerID    uuid.UUID `json:"learner_id"`
	LearnerName  string    `json:"learner_name"`
	Score        int       `json:"score"`
	Achievements int       `json:"achievements"`
	Avatar       string    `json:"avatar,omitempty"`
}

// SuccessResponse 
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse 
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

