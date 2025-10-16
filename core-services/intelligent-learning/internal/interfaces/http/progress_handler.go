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

// ProgressHandler ?
type ProgressHandler struct {
	progressService analyticsServices.ProgressTrackingService
}

// NewProgressHandler ?
func NewProgressHandler(progressService analyticsServices.ProgressTrackingService) *ProgressHandler {
	return &ProgressHandler{
		progressService: progressService,
	}
}

// UpdateProgress 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body analyticsServices.ProgressUpdateRequest true ""
// @Success 200 {object} analyticsServices.ProgressResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/progress/update [post]
func (h *ProgressHandler) UpdateProgress(c *gin.Context) {
	var req analyticsServices.ProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	if err := h.validateProgressUpdateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	response, err := h.progressService.UpdateProgress(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPDATE_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetLearningReport 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Param period query string false "" Enums(daily,weekly,monthly,custom) default(weekly)
// @Param start_date query string false "?(YYYY-MM-DD)"
// @Param end_date query string false " (YYYY-MM-DD)"
// @Success 200 {object} analyticsServices.LearningReport ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/progress/report/{learner_id} [get]
func (h *ProgressHandler) GetLearningReport(c *gin.Context) {
	// ID
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_LEARNER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	period := c.DefaultQuery("period", "weekly")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	reportPeriod, err := h.parseReportPeriod(period, startDateStr, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_PERIOD",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	report, err := h.progressService.GetLearningReport(c.Request.Context(), learnerID, *reportPeriod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "REPORT_GENERATION_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetProgressSummary 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Success 200 {object} ProgressSummaryResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/progress/summary/{learner_id} [get]
func (h *ProgressHandler) GetProgressSummary(c *gin.Context) {
	// ID
	learnerIDStr := c.Param("learner_id")
	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_LEARNER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	summary, err := h.getProgressSummary(c.Request.Context(), learnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SUMMARY_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetContentProgress 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param learner_id path string true "ID"
// @Param content_id path string true "ID"
// @Success 200 {object} ContentProgressResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 404 {object} ErrorResponse "?
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/progress/{learner_id}/content/{content_id} [get]
func (h *ProgressHandler) GetContentProgress(c *gin.Context) {
	// 
	learnerIDStr := c.Param("learner_id")
	contentIDStr := c.Param("content_id")

	learnerID, err := uuid.Parse(learnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_LEARNER_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	contentID, err := uuid.Parse(contentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_CONTENT_ID",
			Message: "ID",
			Details: err.Error(),
		})
		return
	}

	// 
	progress, err := h.getContentProgress(c.Request.Context(), learnerID, contentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "PROGRESS_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// BatchUpdateProgress 
// @Summary 
// @Description 
// @Tags 
// @Accept json
// @Produce json
// @Param request body BatchProgressUpdateRequest true ""
// @Success 200 {object} BatchProgressUpdateResponse ""
// @Failure 400 {object} ErrorResponse ""
// @Failure 500 {object} ErrorResponse "?
// @Router /api/v1/progress/batch-update [post]
func (h *ProgressHandler) BatchUpdateProgress(c *gin.Context) {
	var req BatchProgressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	// 
	response, err := h.batchUpdateProgress(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "BATCH_UPDATE_FAILED",
			Message: "",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// 

// validateProgressUpdateRequest 
func (h *ProgressHandler) validateProgressUpdateRequest(req *analyticsServices.ProgressUpdateRequest) error {
	if req.LearnerID == uuid.Nil {
		return fmt.Errorf("ID")
	}
	if req.ContentID == uuid.Nil {
		return fmt.Errorf("ID")
	}
	if req.Progress < 0 || req.Progress > 1 {
		return fmt.Errorf("0-1")
	}
	if req.TimeSpent < 0 {
		return fmt.Errorf("䲻")
	}
	return nil
}

// parseReportPeriod 
func (h *ProgressHandler) parseReportPeriod(period, startDateStr, endDateStr string) (*analyticsServices.ReportPeriod, error) {
	now := time.Now()
	var startDate, endDate time.Time

	switch period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
	case "weekly":
		// 
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 7
		}
		startDate = now.AddDate(0, 0, -(weekday - 1))
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDate = startDate.Add(7 * 24 * time.Hour)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	case "custom":
		if startDateStr == "" || endDateStr == "" {
			return nil, fmt.Errorf("")
		}
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("? %v", err)
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf(": %v", err)
		}
		// 
		endDate = endDate.Add(24 * time.Hour)
	default:
		return nil, fmt.Errorf(": %s", period)
	}

	return &analyticsServices.ReportPeriod{
		StartDate: startDate,
		EndDate:   endDate,
		Type:      period,
	}, nil
}

// 

// ErrorResponse 
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ProgressSummaryResponse 
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

// ActiveContentInfo 
type ActiveContentInfo struct {
	ContentID    uuid.UUID     `json:"content_id"`
	Title        string        `json:"title"`
	Type         string        `json:"type"`
	Progress     float64       `json:"progress"`
	LastAccessed time.Time     `json:"last_accessed"`
	TimeSpent    time.Duration `json:"time_spent"`
}

// ContentProgressResponse 
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

// BatchProgressUpdateRequest 
type BatchProgressUpdateRequest struct {
	Updates []analyticsServices.ProgressUpdateRequest `json:"updates" validate:"required,min=1,max=50"`
}

// BatchProgressUpdateResponse 
type BatchProgressUpdateResponse struct {
	SuccessCount int                 `json:"success_count"`
	FailureCount int                 `json:"failure_count"`
	Results      []BatchUpdateResult `json:"results"`
	ProcessedAt  time.Time           `json:"processed_at"`
}

// BatchUpdateResult 
type BatchUpdateResult struct {
	LearnerID uuid.UUID                   `json:"learner_id"`
	ContentID uuid.UUID                   `json:"content_id"`
	Success   bool                        `json:"success"`
	Response  *analyticsServices.ProgressResponse `json:"response,omitempty"`
	Error     string                      `json:"error,omitempty"`
}

// 
func (h *ProgressHandler) getProgressSummary(ctx context.Context, learnerID uuid.UUID) (*ProgressSummaryResponse, error) {
	// TODO: 
	return &ProgressSummaryResponse{
		LearnerID: learnerID,
		UpdatedAt: time.Now(),
	}, nil
}

func (h *ProgressHandler) getContentProgress(ctx context.Context, learnerID, contentID uuid.UUID) (*ContentProgressResponse, error) {
	// TODO: 
	return &ContentProgressResponse{
		LearnerID: learnerID,
		ContentID: contentID,
		UpdatedAt: time.Now(),
	}, nil
}

func (h *ProgressHandler) batchUpdateProgress(ctx context.Context, req *BatchProgressUpdateRequest) (*BatchProgressUpdateResponse, error) {
	// TODO: 
	return &BatchProgressUpdateResponse{
		SuccessCount: len(req.Updates),
		FailureCount: 0,
		ProcessedAt:  time.Now(),
	}, nil
}

