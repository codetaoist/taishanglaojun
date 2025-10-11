package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	analyticsServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/analytics"
)

// ProgressHandler 进度追踪处理�?
type ProgressHandler struct {
	progressService analyticsServices.ProgressTrackingService
}

// NewProgressHandler 创建新的进度处理�?
func NewProgressHandler(progressService analyticsServices.ProgressTrackingService) *ProgressHandler {
	return &ProgressHandler{
		progressService: progressService,
	}
}

// UpdateProgress 更新学习进度
// @Summary 更新学习进度
// @Description 实时更新学习者的内容学习进度，包括进度百分比、学习时间、测验结果等
// @Tags 进度追踪
// @Accept json
// @Produce json
// @Param request body analyticsServices.ProgressUpdateRequest true "进度更新请求"
// @Success 200 {object} analyticsServices.ProgressResponse "进度更新响应"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者或内容不存�?
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/progress/update [post]
func (h *ProgressHandler) UpdateProgress(c *gin.Context) {
	var req analyticsServices.ProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数格式错误",
			Details: err.Error(),
		})
		return
	}

	// 验证请求参数
	if err := h.validateProgressUpdateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "请求参数验证失败",
			Details: err.Error(),
		})
		return
	}

	// 更新进度
	response, err := h.progressService.UpdateProgress(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: "进度更新失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLearningReport 获取学习报告
// @Summary 获取学习报告
// @Description 生成指定时间段的学习报告，包括进度分析、性能指标、学习模式等
// @Tags 进度追踪
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param period query string false "报告周期" Enums(daily,weekly,monthly,custom) default(weekly)
// @Param start_date query string false "开始日�?(YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} analyticsServices.LearningReport "学习报告"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/progress/report/{learner_id} [get]
func (h *ProgressHandler) GetLearningReport(c *gin.Context) {
	// 解析学习者ID
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_LEARNER_ID",
			Message: "学习者ID格式错误",
			Details: err.Error(),
		})
		return
	}

	// 解析报告周期
	period := c.DefaultQuery("period", "weekly")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	reportPeriod, err := h.parseReportPeriod(period, startDateStr, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PERIOD",
			Message: "报告周期参数错误",
			Details: err.Error(),
		})
		return
	}

	// 生成学习报告
	report, err := h.progressService.GetLearningReport(c.Request.Context(), learnerID, *reportPeriod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "REPORT_GENERATION_FAILED",
			Message: "学习报告生成失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetProgressSummary 获取进度摘要
// @Summary 获取进度摘要
// @Description 获取学习者的进度摘要信息，包括总体进度、当前学习内容等
// @Tags 进度追踪
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Success 200 {object} ProgressSummaryResponse "进度摘要"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "学习者不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/progress/summary/{learner_id} [get]
func (h *ProgressHandler) GetProgressSummary(c *gin.Context) {
	// 解析学习者ID
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_LEARNER_ID",
			Message: "学习者ID格式错误",
			Details: err.Error(),
		})
		return
	}

	// 获取进度摘要
	summary, err := h.getProgressSummary(c.Request.Context(), learnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SUMMARY_FAILED",
			Message: "进度摘要获取失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetContentProgress 获取内容进度
// @Summary 获取内容进度
// @Description 获取学习者对特定内容的详细进度信息，包括学习状态、完成时间等
// @Tags 进度追踪
// @Accept json
// @Produce json
// @Param learner_id path string true "学习者ID"
// @Param content_id path string true "内容ID"
// @Success 200 {object} ContentProgressResponse "内容进度"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "进度记录不存�?
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/progress/{learner_id}/content/{content_id} [get]
func (h *ProgressHandler) GetContentProgress(c *gin.Context) {
	// 解析参数
	learnerIDStr := c.Param("learner_id")
	contentIDStr := c.Param("content_id")

	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_LEARNER_ID",
			Message: "学习者ID格式错误",
			Details: err.Error(),
		})
		return
	}

	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_CONTENT_ID",
			Message: "内容ID格式错误",
			Details: err.Error(),
		})
		return
	}

	// 获取内容进度
	progress, err := h.getContentProgress(c.Request.Context(), learnerID, contentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "PROGRESS_FAILED",
			Message: "内容进度获取失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// BatchUpdateProgress 批量更新进度
// @Summary 批量更新进度
// @Description 批量更新多个内容的学习进度，包括学习状态、完成时间等
// @Tags 进度追踪
// @Accept json
// @Produce json
// @Param request body BatchProgressUpdateRequest true "批量进度更新请求"
// @Success 200 {object} BatchProgressUpdateResponse "批量更新响应"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 500 {object} ErrorResponse "服务器内部错�?
// @Router /api/v1/progress/batch-update [post]
func (h *ProgressHandler) BatchUpdateProgress(c *gin.Context) {
	var req BatchProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "请求参数格式错误",
			Details: err.Error(),
		})
		return
	}

	// 批量更新进度
	response, err := h.batchUpdateProgress(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "BATCH_UPDATE_FAILED",
			Message: "批量更新失败",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// 辅助方法

// validateProgressUpdateRequest 验证进度更新请求
func (h *ProgressHandler) validateProgressUpdateRequest(req *analyticsServices.ProgressUpdateRequest) error {
	if req.LearnerID == uuid.Nil {
		return fmt.Errorf("学习者ID不能为空")
	}
	if req.ContentID == uuid.Nil {
		return fmt.Errorf("内容ID不能为空")
	}
	if req.Progress < 0 || req.Progress > 1 {
		return fmt.Errorf("进度值必须在0-1之间")
	}
	if req.TimeSpent < 0 {
		return fmt.Errorf("学习时间不能为负")
	}
	return nil
}

// parseReportPeriod 解析报告周期
func (h *ProgressHandler) parseReportPeriod(period, startDateStr, endDateStr string) (*analyticsServices.ReportPeriod, error) {
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		// 获取本周开始时间（周一为第一天）
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 将周日调整为7
		}
		startDate = now.AddDate(0, 0, -(weekday - 1))
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDate = startDate.Add(7 * 24 * time.Hour)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	case "custom":
		if startDateStr == "" || endDateStr == "" {
			return nil, fmt.Errorf("自定义周期需要提供开始和结束日期")
		}
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错�? %v", err)
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误: %v", err)
		}
		// 结束日期加一天，包含整个结束日期
		endDate = endDate.Add(24 * time.Hour)
	default:
		return nil, fmt.Errorf("不支持的报告周期: %s", period)
	}

	return &analyticsServices.ReportPeriod{
		StartDate: startDate,
		EndDate:   endDate,
		Type:      period,
	}, nil
}

// 数据传输对象

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProgressSummaryResponse 进度摘要响应
type ProgressSummaryResponse struct {
	LearnerID           uuid.UUID                          `json:"learner_id"`
	OverallProgress     float64                            `json:"overall_progress"`
	TotalTimeSpent      time.Duration                      `json:"total_time_spent"`
	ContentCompleted    int                                `json:"content_completed"`
	ContentInProgress   int                                `json:"content_in_progress"`
	CurrentStreak       int                                `json:"current_streak"`
	WeeklyGoalProgress  float64                            `json:"weekly_goal_progress"`
	RecentAchievements  []analyticsServices.ProgressAchievement    `json:"recent_achievements"`
	ActiveContent       []ActiveContentInfo                `json:"active_content"`
	NextRecommendations []analyticsServices.NextStepRecommendation `json:"next_recommendations"`
	UpdatedAt           time.Time                          `json:"updated_at"`
}

// ActiveContentInfo 活跃内容信息
type ActiveContentInfo struct {
	ContentID    uuid.UUID     `json:"content_id"`
	Title        string        `json:"title"`
	Type         string        `json:"type"`
	Progress     float64       `json:"progress"`
	LastAccessed time.Time     `json:"last_accessed"`
	TimeSpent    time.Duration `json:"time_spent"`
}

// ContentProgressResponse 内容进度响应
type ContentProgressResponse struct {
	LearnerID        uuid.UUID                `json:"learner_id"`
	ContentID        uuid.UUID                `json:"content_id"`
	ContentTitle     string                   `json:"content_title"`
	Progress         float64                  `json:"progress"`
	TimeSpent        time.Duration            `json:"time_spent"`
	LastPosition     int                      `json:"last_position"`
	IsCompleted      bool                     `json:"is_completed"`
	CompletedAt      *time.Time               `json:"completed_at"`
	QuizScores       map[uuid.UUID]float64    `json:"quiz_scores"`
	Notes            []analyticsServices.NoteData     `json:"notes"`
	Bookmarks        []analyticsServices.BookmarkData `json:"bookmarks"`
	InteractionCount int                      `json:"interaction_count"`
	PerformanceScore float64                  `json:"performance_score"`
	EngagementLevel  string                   `json:"engagement_level"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

// BatchProgressUpdateRequest 批量进度更新请求
type BatchProgressUpdateRequest struct {
	Updates []analyticsServices.ProgressUpdateRequest `json:"updates" validate:"required,min=1,max=50"`
}

// BatchProgressUpdateResponse 批量进度更新响应
type BatchProgressUpdateResponse struct {
	SuccessCount int                 `json:"success_count"`
	FailureCount int                 `json:"failure_count"`
	Results      []BatchUpdateResult `json:"results"`
	ProcessedAt  time.Time           `json:"processed_at"`
}

// BatchUpdateResult 批量更新结果
type BatchUpdateResult struct {
	LearnerID uuid.UUID                   `json:"learner_id"`
	ContentID uuid.UUID                   `json:"content_id"`
	Success   bool                        `json:"success"`
	Response  *analyticsServices.ProgressResponse `json:"response,omitempty"`
	Error     string                      `json:"error,omitempty"`
}

// 实现辅助方法的占位符
func (h *ProgressHandler) getProgressSummary(ctx context.Context, learnerID uuid.UUID) (*ProgressSummaryResponse, error) {
	// TODO: 实现进度摘要获取逻辑
	return &ProgressSummaryResponse{
		LearnerID: learnerID,
		UpdatedAt: time.Now(),
	}, nil
}

func (h *ProgressHandler) getContentProgress(ctx context.Context, learnerID, contentID uuid.UUID) (*ContentProgressResponse, error) {
	// TODO: 实现内容进度获取逻辑
	return &ContentProgressResponse{
		LearnerID: learnerID,
		ContentID: contentID,
		UpdatedAt: time.Now(),
	}, nil
}

func (h *ProgressHandler) batchUpdateProgress(ctx context.Context, req *BatchProgressUpdateRequest) (*BatchProgressUpdateResponse, error) {
	// TODO: 实现批量更新逻辑
	return &BatchProgressUpdateResponse{
		SuccessCount: len(req.Updates),
		FailureCount: 0,
		ProcessedAt:  time.Now(),
	}, nil
}
