package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)



// ProgressTrackingService 
type ProgressTrackingService struct {
	learnerRepo         repositories.LearnerRepository
	contentRepo         repositories.LearningContentRepository
	knowledgeGraphRepo  repositories.KnowledgeGraphRepository
	analyticsService    interfaces.LearningAnalyticsService
}

// NewProgressTrackingService 
func NewProgressTrackingService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	analyticsService interfaces.LearningAnalyticsService,
) *ProgressTrackingService {
	return &ProgressTrackingService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		analyticsService:   analyticsService,
	}
}

// ProgressUpdateRequest 
type ProgressUpdateRequest struct {
	LearnerID       uuid.UUID              `json:"learner_id" validate:"required"`
	ContentID       uuid.UUID              `json:"content_id" validate:"required"`
	Progress        float64                `json:"progress" validate:"min=0,max=1"`
	TimeSpent       int                    `json:"time_spent"` // ?
	LastPosition    int                    `json:"last_position"`
	InteractionData map[string]interface{} `json:"interaction_data"`
	QuizResults     []QuizResult           `json:"quiz_results,omitempty"`
	Notes           []NoteData             `json:"notes,omitempty"`
	Bookmarks       []BookmarkData         `json:"bookmarks,omitempty"`
}

// QuizResult 
type QuizResult struct {
	QuestionID    uuid.UUID `json:"question_id"`
	Answer        interface{} `json:"answer"`
	IsCorrect     bool      `json:"is_correct"`
	Score         float64   `json:"score"`
	TimeSpent     int       `json:"time_spent"`
	AttemptCount  int       `json:"attempt_count"`
}

// NoteData 
type NoteData struct {
	Content   string   `json:"content"`
	Position  int      `json:"position"`
	Tags      []string `json:"tags"`
	IsPublic  bool     `json:"is_public"`
}

// BookmarkData 
type BookmarkData struct {
	Title    string `json:"title"`
	Position int    `json:"position"`
	Note     string `json:"note"`
}

// ProgressResponse 
type ProgressResponse struct {
	LearnerID          uuid.UUID                    `json:"learner_id"`
	ContentID          uuid.UUID                    `json:"content_id"`
	Progress           float64                      `json:"progress"`
	TimeSpent          time.Duration                `json:"time_spent"`
	EstimatedRemaining time.Duration                `json:"estimated_remaining"`
	CompletionRate     float64                      `json:"completion_rate"`
	PerformanceScore   float64                      `json:"performance_score"`
	EngagementLevel    string                       `json:"engagement_level"`
	Recommendations    []string                     `json:"recommendations"`
	NextSteps          []NextStepRecommendation     `json:"next_steps"`
	Achievements       []domainServices.Achievement                `json:"achievements"`
	UpdatedAt          time.Time                    `json:"updated_at"`
}

// NextStepRecommendation ?
type NextStepRecommendation struct {
	Type        string    `json:"type"` // "content", "review", "practice", "assessment"
	ContentID   uuid.UUID `json:"content_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Reason      string    `json:"reason"`
}

// Achievement 
type ProgressAchievement struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Points      int       `json:"points"`
	UnlockedAt  time.Time `json:"unlocked_at"`
}

// LearningReport 
type LearningReport struct {
	LearnerID           uuid.UUID                 `json:"learner_id"`
	ReportPeriod        ReportPeriod              `json:"report_period"`
	OverallProgress     OverallProgress           `json:"overall_progress"`
	ContentProgress     []ContentProgressSummary `json:"content_progress"`
	SkillDevelopment    []SkillProgress           `json:"skill_development"`
	LearningPatterns    LearningPatternAnalysis   `json:"learning_patterns"`
	PerformanceMetrics  domainServices.PerformanceMetrics        `json:"performance_metrics"`
	Recommendations     []RecommendationItem      `json:"recommendations"`
	Goals               []GoalProgress            `json:"goals"`
	Achievements        []domainServices.Achievement             `json:"achievements"`
	GeneratedAt         time.Time                 `json:"generated_at"`
}

// ReportPeriod 
type ReportPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Type      string    `json:"type"` // "daily", "weekly", "monthly", "custom"
}

// OverallProgress 
type OverallProgress struct {
	CompletionRate      float64       `json:"completion_rate"`
	TotalTimeSpent      time.Duration `json:"total_time_spent"`
	ContentCompleted    int           `json:"content_completed"`
	SkillsAcquired      int           `json:"skills_acquired"`
	CurrentStreak       int           `json:"current_streak"`
	WeeklyGoalProgress  float64       `json:"weekly_goal_progress"`
	MonthlyGoalProgress float64       `json:"monthly_goal_progress"`
}

// ContentProgressSummary 
type ContentProgressSummary struct {
	ContentID        uuid.UUID     `json:"content_id"`
	Title            string        `json:"title"`
	Type             string        `json:"type"`
	Progress         float64       `json:"progress"`
	TimeSpent        time.Duration `json:"time_spent"`
	CompletedAt      *time.Time    `json:"completed_at"`
	PerformanceScore float64       `json:"performance_score"`
	Difficulty       string        `json:"difficulty"`
}

// SkillProgress ?
type SkillProgress struct {
	SkillName       string    `json:"skill_name"`
	PreviousLevel   float64   `json:"previous_level"`
	CurrentLevel    float64   `json:"current_level"`
	Improvement     float64   `json:"improvement"`
	LastUpdated     time.Time `json:"last_updated"`
	RelatedContent  []uuid.UUID `json:"related_content"`
}

// LearningPatternAnalysis 
type LearningPatternAnalysis struct {
	OptimalStudyTime    []TimeSlotAnalysis `json:"optimal_study_time"`
	PreferredContentTypes map[string]float64 `json:"preferred_content_types"`
	LearningVelocity    float64            `json:"learning_velocity"`
	RetentionRate       float64            `json:"retention_rate"`
	EngagementPatterns  []EngagementPattern `json:"engagement_patterns"`
	DropoffPoints       []DropoffAnalysis   `json:"dropoff_points"`
}

// TimeSlotAnalysis ?
type TimeSlotAnalysis struct {
	Hour            int     `json:"hour"`
	PerformanceScore float64 `json:"performance_score"`
	EngagementLevel float64 `json:"engagement_level"`
	CompletionRate  float64 `json:"completion_rate"`
}

// EngagementPattern 
type EngagementPattern struct {
	Pattern     string  `json:"pattern"`
	Frequency   float64 `json:"frequency"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
}

// DropoffAnalysis 
type DropoffAnalysis struct {
	ContentType string  `json:"content_type"`
	Position    int     `json:"position"` // ?
	Frequency   float64 `json:"frequency"`
	Reasons     []string `json:"reasons"`
}

// PerformanceMetrics 
type ProgressPerformanceMetrics struct {
	AverageScore        float64 `json:"average_score"`
	ImprovementRate     float64 `json:"improvement_rate"`
	ConsistencyScore    float64 `json:"consistency_score"`
	EfficiencyScore     float64 `json:"efficiency_score"`
	EngagementScore     float64 `json:"engagement_score"`
	RetentionScore      float64 `json:"retention_score"`
}

// RecommendationItem ?
type RecommendationItem struct {
	Type        string    `json:"type"`
	Priority    int       `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ActionItems []string  `json:"action_items"`
	ExpectedImpact string `json:"expected_impact"`
}

// GoalProgress 
type GoalProgress struct {
	GoalID          uuid.UUID `json:"goal_id"`
	Description     string    `json:"description"`
	TargetDate      time.Time `json:"target_date"`
	CurrentProgress float64   `json:"current_progress"`
	IsOnTrack       bool      `json:"is_on_track"`
	DaysRemaining   int       `json:"days_remaining"`
	Recommendations []string  `json:"recommendations"`
}

// UpdateProgress 
func (s *ProgressTrackingService) UpdateProgress(ctx context.Context, req *ProgressUpdateRequest) (*ProgressResponse, error) {
	// ?
	progress, err := s.getOrCreateContentProgress(ctx, req.LearnerID, req.ContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content progress: %w", err)
	}

	// 
	progress.UpdateProgress(req.Progress, req.LastPosition, req.TimeSpent)

	// 
	if len(req.QuizResults) > 0 {
		s.processQuizResults(progress, req.QuizResults)
	}

	// 
	for _, noteData := range req.Notes {
		progress.AddNote(noteData.Content, noteData.Position, noteData.Tags, noteData.IsPublic)
	}

	// 
	for _, bookmarkData := range req.Bookmarks {
		progress.AddBookmark(bookmarkData.Title, bookmarkData.Position, bookmarkData.Note)
	}

	// 
	if len(req.InteractionData) > 0 {
		progress.RecordInteraction("update", "progress", req.LastPosition, req.InteractionData)
	}

	// ?
	err = s.saveProgressToLearningHistory(ctx, req, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to save progress: %w", err)
	}

	// ?
	err = s.updateLearnerStatistics(ctx, req.LearnerID, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to update learner statistics: %w", err)
	}

	// 
	response, err := s.generateProgressResponse(ctx, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

// GetLearningReport 
func (s *ProgressTrackingService) GetLearningReport(ctx context.Context, learnerID uuid.UUID, period ReportPeriod) (*LearningReport, error) {
	// ?
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 
	var filteredHistory []entities.LearningHistory
	for _, h := range learner.LearningHistory {
		if h.Timestamp.After(period.StartDate) && h.Timestamp.Before(period.EndDate) {
			filteredHistory = append(filteredHistory, h)
		}
	}

	// ?
	var historyPointers []*entities.LearningHistory
	for i := range filteredHistory {
		historyPointers = append(historyPointers, &filteredHistory[i])
	}

	// 
	var contentProgress []*entities.ContentProgress
	for _, h := range filteredHistory {
		cp := &entities.ContentProgress{
			LearnerID:      learnerID,
			ContentID:      h.ContentID,
			Progress:       h.Progress,
			TimeSpent:      int(h.Duration.Seconds()),
			LastAccessedAt: h.Timestamp,
			IsCompleted:    h.Progress >= 1.0,
		}
		contentProgress = append(contentProgress, cp)
	}

	// 
	overallProgress := s.analyzeOverallProgress(historyPointers, contentProgress)
	contentProgressSummary := s.analyzeContentProgress(contentProgress)
	skillDevelopment := s.analyzeSkillDevelopment(ctx, learner, historyPointers)
	learningPatterns := s.analyzeLearningPatterns(historyPointers, contentProgress)
	performanceMetrics := s.calculatePerformanceMetrics(historyPointers, contentProgress)
	recommendations := s.generateRecommendations(ctx, learner, overallProgress, performanceMetrics)
	goals := s.analyzeGoalProgress(learner, overallProgress)
	achievements := s.getAchievements(ctx, learnerID, period)

	return &LearningReport{
		LearnerID:          learnerID,
		ReportPeriod:       period,
		OverallProgress:    overallProgress,
		ContentProgress:    contentProgressSummary,
		SkillDevelopment:   skillDevelopment,
		LearningPatterns:   learningPatterns,
		PerformanceMetrics: performanceMetrics,
		Recommendations:    recommendations,
		Goals:              goals,
		Achievements:       achievements,
		GeneratedAt:        time.Now(),
	}, nil
}

// getOrCreateContentProgress ?
func (s *ProgressTrackingService) getOrCreateContentProgress(ctx context.Context, learnerID, contentID uuid.UUID) (*entities.ContentProgress, error) {
	progress, err := s.contentRepo.GetProgress(ctx, learnerID, contentID)
	if err != nil {
		// 
		progress = entities.NewContentProgress(learnerID, contentID)
	}
	return progress, nil
}

// processQuizResults 
func (s *ProgressTrackingService) processQuizResults(progress *entities.ContentProgress, results []QuizResult) {
	for _, result := range results {
		progress.QuizScores[result.QuestionID] = result.Score
		
		// 
		interactionData := map[string]interface{}{
			"answer":        result.Answer,
			"is_correct":    result.IsCorrect,
			"score":         result.Score,
			"attempt_count": result.AttemptCount,
		}
		progress.RecordInteraction("quiz_answer", fmt.Sprintf("question_%s", result.QuestionID), 0, interactionData)
	}
}

// updateLearnerStatistics ?
func (s *ProgressTrackingService) updateLearnerStatistics(ctx context.Context, learnerID uuid.UUID, progress *entities.ContentProgress) error {
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// ?
	additionalHours := int(math.Round(float64(progress.TimeSpent) / 3600.0))
	learner.TotalStudyHours += additionalHours

	// ?
	if progress.IsCompleted {
		s.updateLearningStreak(learner, time.Now())
	}

	// 
	return s.learnerRepo.Update(ctx, learner)
}

// saveProgressToLearningHistory ?
func (s *ProgressTrackingService) saveProgressToLearningHistory(ctx context.Context, req *ProgressUpdateRequest, progress *entities.ContentProgress) error {
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return err
	}

	// ?
	var existingHistory *entities.LearningHistory
	for i := range learner.LearningHistory {
		if learner.LearningHistory[i].ContentID == req.ContentID {
			existingHistory = &learner.LearningHistory[i]
			break
		}
	}

	if existingHistory != nil {
		// 
		existingHistory.Progress = progress.Progress
		existingHistory.Duration += time.Duration(req.TimeSpent) * time.Second
		existingHistory.Completed = progress.IsCompleted
		existingHistory.Timestamp = time.Now()
		if progress.IsCompleted && existingHistory.EndTime == nil {
			now := time.Now()
			existingHistory.EndTime = &now
		}
	} else {
		// 
		history := entities.LearningHistory{
			ID:          uuid.New(),
			LearnerID:   req.LearnerID,
			ContentID:   req.ContentID,
			ContentType: "unknown", // content repository
			Progress:    progress.Progress,
			Duration:    time.Duration(req.TimeSpent) * time.Second,
			Completed:   progress.IsCompleted,
			StartTime:   time.Now(),
			Timestamp:   time.Now(),
			CreatedAt:   time.Now(),
		}
		
		if progress.IsCompleted {
			now := time.Now()
			history.EndTime = &now
		}

		learner.LearningHistory = append(learner.LearningHistory, history)
	}

	return s.learnerRepo.Update(ctx, learner)
}

// updateLearningStreak ?
func (s *ProgressTrackingService) updateLearningStreak(learner *entities.Learner, studyDate time.Time) {
	today := studyDate.Truncate(24 * time.Hour)
	lastStudyDate := learner.Streak.LastStudyDate.Truncate(24 * time.Hour)

	if today.Equal(lastStudyDate) {
		// ?
		return
	}

	if today.Equal(lastStudyDate.Add(24 * time.Hour)) {
		// ?
		learner.Streak.CurrentStreak++
		if learner.Streak.CurrentStreak > learner.Streak.LongestStreak {
			learner.Streak.LongestStreak = learner.Streak.CurrentStreak
		}
	} else if today.After(lastStudyDate.Add(24 * time.Hour)) {
		// ?
		learner.Streak.CurrentStreak = 1
	}

	learner.Streak.LastStudyDate = studyDate
	learner.Streak.TotalDays++
}

// generateProgressResponse 
func (s *ProgressTrackingService) generateProgressResponse(ctx context.Context, progress *entities.ContentProgress) (*ProgressResponse, error) {
	// 
	content, err := s.contentRepo.GetByID(ctx, progress.ContentID)
	if err != nil {
		return nil, err
	}

	// 
	estimatedRemaining := s.calculateEstimatedRemainingTime(progress, content)

	// 
	performanceScore := s.calculatePerformanceScore(progress)

	// ?
	engagementLevel := s.assessEngagementLevel(progress)

	// 
	recommendations := s.generateProgressRecommendations(progress, content)

	// ?
	nextSteps := s.generateNextSteps(ctx, progress, content)

	// ?
	achievements := s.checkAchievements(ctx, progress)

	return &ProgressResponse{
		LearnerID:          progress.LearnerID,
		ContentID:          progress.ContentID,
		Progress:           progress.Progress,
		TimeSpent:          time.Duration(progress.TimeSpent) * time.Second,
		EstimatedRemaining: estimatedRemaining,
		CompletionRate:     progress.Progress,
		PerformanceScore:   performanceScore,
		EngagementLevel:    engagementLevel,
		Recommendations:    recommendations,
		NextSteps:          nextSteps,
		Achievements:       achievements,
		UpdatedAt:          time.Now(),
	}, nil
}

// calculateEstimatedRemainingTime 
func (s *ProgressTrackingService) calculateEstimatedRemainingTime(progress *entities.ContentProgress, content *entities.LearningContent) time.Duration {
	if progress.Progress >= 1.0 {
		return 0
	}

	// ?
	estimatedTotal := time.Duration(content.EstimatedDuration) * time.Minute
	if progress.Progress > 0 {
		// 
		actualRate := float64(progress.TimeSpent) / progress.Progress
		remaining := (1.0 - progress.Progress) * actualRate
		return time.Duration(remaining) * time.Second
	}

	return estimatedTotal
}

// calculatePerformanceScore 
func (s *ProgressTrackingService) calculatePerformanceScore(progress *entities.ContentProgress) float64 {
	score := 0.0
	factors := 0

	// 
	score += progress.Progress * 30
	factors++

	// 
	if len(progress.QuizScores) > 0 {
		totalScore := 0.0
		for _, quizScore := range progress.QuizScores {
			totalScore += quizScore
		}
		avgQuizScore := totalScore / float64(len(progress.QuizScores))
		score += avgQuizScore * 40
		factors++
	}

	// ?
	if len(progress.InteractionLog) > 0 {
		engagementScore := math.Min(float64(len(progress.InteractionLog))/10.0, 1.0) * 20
		score += engagementScore
		factors++
	}

	// ?
	if len(progress.Notes) > 0 || len(progress.Bookmarks) > 0 {
		activeScore := math.Min(float64(len(progress.Notes)+len(progress.Bookmarks))/5.0, 1.0) * 10
		score += activeScore
		factors++
	}

	if factors > 0 {
		return score / float64(factors)
	}
	return 0
}

// assessEngagementLevel ?
func (s *ProgressTrackingService) assessEngagementLevel(progress *entities.ContentProgress) string {
	score := 0

	// 
	if len(progress.InteractionLog) > 20 {
		score += 3
	} else if len(progress.InteractionLog) > 10 {
		score += 2
	} else if len(progress.InteractionLog) > 5 {
		score += 1
	}

	// ?
	if len(progress.Notes) > 3 || len(progress.Bookmarks) > 2 {
		score += 2
	} else if len(progress.Notes) > 0 || len(progress.Bookmarks) > 0 {
		score += 1
	}

	// 
	if progress.TimeSpent > 3600 { // 1
		score += 2
	} else if progress.TimeSpent > 1800 { // 30
		score += 1
	}

	switch {
	case score >= 6:
		return "high"
	case score >= 3:
		return "medium"
	default:
		return "low"
	}
}

// generateProgressRecommendations 
func (s *ProgressTrackingService) generateProgressRecommendations(progress *entities.ContentProgress, content *entities.LearningContent) []string {
	var recommendations []string

	// ?
	if progress.Progress < 0.3 {
		recommendations = append(recommendations, "趨?)
	} else if progress.Progress < 0.7 {
		recommendations = append(recommendations, "鱣?)
	} else if progress.Progress < 1.0 {
		recommendations = append(recommendations, "")
	}

	// ?
	if len(progress.QuizScores) > 0 {
		totalScore := 0.0
		for _, score := range progress.QuizScores {
			totalScore += score
		}
		avgScore := totalScore / float64(len(progress.QuizScores))
		
		if avgScore < 0.6 {
			recommendations = append(recommendations, "?)
		} else if avgScore > 0.8 {
			recommendations = append(recommendations, "")
		}
	}

	// 
	if len(progress.InteractionLog) < 5 {
		recommendations = append(recommendations, "?)
	}

	if len(progress.Notes) == 0 {
		recommendations = append(recommendations, "")
	}

	return recommendations
}

// generateNextSteps ?
func (s *ProgressTrackingService) generateNextSteps(ctx context.Context, progress *entities.ContentProgress, content *entities.LearningContent) []NextStepRecommendation {
	var nextSteps []NextStepRecommendation

	if progress.Progress < 1.0 {
		// 
		nextSteps = append(nextSteps, NextStepRecommendation{
			Type:        "content",
			ContentID:   content.ID,
			Title:       "",
			Description: fmt.Sprintf("?s %.1f%%", content.Title, progress.Progress*100),
			Priority:    1,
			Reason:      "?,
		})
	} else {
		// ?
		nextSteps = append(nextSteps, NextStepRecommendation{
			Type:        "review",
			Title:       "",
			Description: "?,
			Priority:    2,
			Reason:      "?,
		})

		nextSteps = append(nextSteps, NextStepRecommendation{
			Type:        "practice",
			Title:       "",
			Description: "?,
			Priority:    1,
			Reason:      "?,
		})
	}

	return nextSteps
}

// checkAchievements ?
func (s *ProgressTrackingService) checkAchievements(ctx context.Context, progress *entities.ContentProgress) []domainServices.Achievement {
	var achievements []domainServices.Achievement

	// 
	if progress.IsCompleted {
		achievements = append(achievements, domainServices.Achievement{
			ID:          uuid.New(),
			Type:        "completion",
			Name:        "?,
			Description: "?,
			Points:      100,
			UnlockedAt:  time.Now(),
		})
	}

	// 
	if len(progress.Notes) >= 5 {
		achievements = append(achievements, domainServices.Achievement{
			ID:          uuid.New(),
			Type:        "note_taker",
			Name:        "",
			Description: "??,
			Points:      50,
			UnlockedAt:  time.Now(),
		})
	}

	// 
	if len(progress.QuizScores) > 0 {
		totalScore := 0.0
		for _, score := range progress.QuizScores {
			totalScore += score
		}
		avgScore := totalScore / float64(len(progress.QuizScores))
		
		if avgScore >= 0.9 {
			achievements = append(achievements, domainServices.Achievement{
				ID:          uuid.New(),
				Type:        "quiz_master",
				Name:        "",
				Description: "?0%",
				Points:      150,
				UnlockedAt:  time.Now(),
			})
		}
	}

	return achievements
}

func (s *ProgressTrackingService) analyzeOverallProgress(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) OverallProgress {
	if len(history) == 0 {
		return OverallProgress{}
	}

	var totalTimeSpent time.Duration
	var completedContent int
	var skillsAcquired int
	var currentStreak int
	var weeklyGoalProgress float64
	var monthlyGoalProgress float64

	// ?
	for _, h := range history {
		totalTimeSpent += h.Duration
	}

	for _, cp := range contentProgress {
		if cp.IsCompleted {
			completedContent++
		}
	}

	// 㼼?
	skillsAcquired = completedContent / 3 // ???

	// ?
	currentStreak = s.calculateLearningStreak(history)

	// ?
	weeklyGoalProgress = s.calculateWeeklyGoalProgress(history)
	monthlyGoalProgress = s.calculateMonthlyGoalProgress(history)

	// ?
	completionRate := float64(completedContent) / float64(len(contentProgress))
	if len(contentProgress) == 0 {
		completionRate = 0
	}

	return OverallProgress{
		CompletionRate:      completionRate,
		TotalTimeSpent:      totalTimeSpent,
		ContentCompleted:    completedContent,
		SkillsAcquired:      skillsAcquired,
		CurrentStreak:       currentStreak,
		WeeklyGoalProgress:  weeklyGoalProgress,
		MonthlyGoalProgress: monthlyGoalProgress,
	}
}

func (s *ProgressTrackingService) analyzeContentProgress(contentProgress []*entities.ContentProgress) []ContentProgressSummary {
	summaries := make([]ContentProgressSummary, 0, len(contentProgress))

	for _, cp := range contentProgress {
		// 
		title := fmt.Sprintf("Content %s", cp.ContentID.String()[:8])
		contentType := "unknown"
		difficulty := "medium"

		// 
		performanceScore := s.calculateContentPerformanceScore(cp)

		var completedAt *time.Time
		if cp.IsCompleted {
			completedAt = &cp.LastAccessedAt
		}

		summary := ContentProgressSummary{
			ContentID:        cp.ContentID,
			Title:            title,
			Type:             contentType,
			Progress:         cp.Progress,
			TimeSpent:        time.Duration(cp.TimeSpent) * time.Second,
			CompletedAt:      completedAt,
			PerformanceScore: performanceScore,
			Difficulty:       difficulty,
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

func (s *ProgressTrackingService) analyzeSkillDevelopment(ctx context.Context, learner *entities.Learner, history []*entities.LearningHistory) []SkillProgress {
	skillProgressMap := make(map[string]*SkillProgress)

	// ?
	for _, h := range history {
		// ?
		skills := s.inferSkillsFromContent(h.ContentID)
		
		for _, skill := range skills {
			if progress, exists := skillProgressMap[skill]; exists {
				// ?
				progress.CurrentLevel += 0.1 // ?
				progress.LastUpdated = h.Timestamp
				progress.RelatedContent = append(progress.RelatedContent, h.ContentID)
			} else {
				// ?
				skillProgressMap[skill] = &SkillProgress{
					SkillName:       skill,
					PreviousLevel:   0.0,
					CurrentLevel:    0.1,
					Improvement:     0.1,
					LastUpdated:     h.Timestamp,
					RelatedContent:  []uuid.UUID{h.ContentID},
				}
			}
		}
	}

	// ?
	skillProgresses := make([]SkillProgress, 0, len(skillProgressMap))
	for _, progress := range skillProgressMap {
		progress.Improvement = progress.CurrentLevel - progress.PreviousLevel
		skillProgresses = append(skillProgresses, *progress)
	}

	return skillProgresses
}

func (s *ProgressTrackingService) analyzeLearningPatterns(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) LearningPatternAnalysis {
	if len(history) == 0 {
		return LearningPatternAnalysis{}
	}

	// ?
	optimalStudyTime := s.analyzeOptimalStudyTime(history)

	// ?
	preferredContentTypes := s.analyzePreferredContentTypes(history)

	// 
	learningVelocity := s.calculateLearningVelocity(history)

	// 㱣?
	retentionRate := s.calculateRetentionRate(history, contentProgress)

	// 
	engagementPatterns := s.analyzeEngagementPatterns(history)

	// ?
	dropoffPoints := s.analyzeDropoffPoints(contentProgress)

	return LearningPatternAnalysis{
		OptimalStudyTime:      optimalStudyTime,
		PreferredContentTypes: preferredContentTypes,
		LearningVelocity:      learningVelocity,
		RetentionRate:         retentionRate,
		EngagementPatterns:    engagementPatterns,
		DropoffPoints:         dropoffPoints,
	}
}

func (s *ProgressTrackingService) calculatePerformanceMetrics(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) domainServices.PerformanceMetrics {
	if len(history) == 0 {
		return domainServices.PerformanceMetrics{}
	}

	// 
	var totalScore float64
	var scoreCount int
	for _, h := range history {
		if h.Score != nil && *h.Score > 0 {
			totalScore += *h.Score
			scoreCount++
		}
	}
	averageScore := totalScore / float64(scoreCount)
	if scoreCount == 0 {
		averageScore = 0
	}

	// ?
	consistencyScore := s.calculateConsistencyScore(history)

	// 
	efficiencyScore := s.calculateEfficiencyScore(history, contentProgress)

	// 㱣
	retentionScore := s.calculateRetentionScore(history, contentProgress)

	return domainServices.PerformanceMetrics{
		Accuracy:       averageScore,
		Speed:          efficiencyScore,
		Efficiency:     efficiencyScore,
		CompletionRate: retentionScore,
		ErrorRate:      1.0 - averageScore, // ?= 1 - ?
		Consistency:    consistencyScore,
		Timeline:       "recent",
		ExpectedOutcome: "improved_performance",
	}
}

func (s *ProgressTrackingService) generateRecommendations(ctx context.Context, learner *entities.Learner, progress OverallProgress, metrics domainServices.PerformanceMetrics) []RecommendationItem {
	recommendations := make([]RecommendationItem, 0)

	// 
	if progress.CompletionRate < 0.3 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "motivation",
			Priority:    1,
			Title:       "?,
			Description: "",
			ActionItems: []string{
				"",
				"?,
				"?,
			},
			ExpectedImpact: "",
		})
	}

	// 
	if progress.CurrentStreak < 3 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "habit",
			Priority:    2,
			Title:       "",
			Description: "",
			ActionItems: []string{
				"?,
				"?,
				"",
			},
			ExpectedImpact: "?,
		})
	}

	// ?
	if metrics.Accuracy < 0.7 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "performance",
			Priority:    1,
			Title:       "",
			Description: "?,
			ActionItems: []string{
				"?,
				"?,
				"",
			},
			ExpectedImpact: "?,
		})
	}

	// ?
	if metrics.Efficiency < 0.6 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "efficiency",
			Priority:    3,
			Title:       "",
			Description: "",
			ActionItems: []string{
				"?,
				"",
				"",
			},
			ExpectedImpact: "?,
		})
	}

	return recommendations
}

func (s *ProgressTrackingService) analyzeGoalProgress(learner *entities.Learner, progress OverallProgress) []GoalProgress {
	goalProgresses := make([]GoalProgress, 0)

	for _, goal := range learner.LearningGoals {
		// 
		currentProgress := s.calculateGoalProgress(goal, progress)
		
		// ?
		isOnTrack := s.isGoalOnTrack(goal, currentProgress)
		
		// 
		daysRemaining := int(goal.TargetDate.Sub(time.Now()).Hours() / 24)
		
		// 
		recommendations := s.generateGoalRecommendations(goal, currentProgress, isOnTrack)

		goalProgress := GoalProgress{
			GoalID:          goal.ID,
			Description:     goal.Description,
			TargetDate:      goal.TargetDate,
			CurrentProgress: currentProgress,
			IsOnTrack:       isOnTrack,
			DaysRemaining:   daysRemaining,
			Recommendations: recommendations,
		}

		goalProgresses = append(goalProgresses, goalProgress)
	}

	return goalProgresses
}

func (s *ProgressTrackingService) getAchievements(ctx context.Context, learnerID uuid.UUID, period ReportPeriod) []domainServices.Achievement {
	achievements := make([]domainServices.Achievement, 0)

	// ?
	achievements = append(achievements, domainServices.Achievement{
		ID:          uuid.New(),
		Type:        "completion",
		Name:        "",
		Description: "?,
		Points:      10,
		UnlockedAt:  time.Now(),
	})

	return achievements
}

// 

func (s *ProgressTrackingService) calculateLearningStreak(history []*entities.LearningHistory) int {
	if len(history) == 0 {
		return 0
	}

	// ?
	sortedHistory := make([]*entities.LearningHistory, len(history))
	copy(sortedHistory, history)

	// ?
	streak := 1
	for i := len(sortedHistory) - 1; i > 0; i-- {
		current := sortedHistory[i].Timestamp
		previous := sortedHistory[i-1].Timestamp
		
		// 2?
		if current.Sub(previous).Hours() > 48 {
			break
		}
		streak++
	}

	return streak
}

func (s *ProgressTrackingService) calculateWeeklyGoalProgress(history []*entities.LearningHistory) float64 {
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	
	var weeklyTime time.Duration
	for _, h := range history {
		if h.Timestamp.After(weekStart) {
			weeklyTime += h.Duration
		}
	}

	// 10
	weeklyGoal := 10 * time.Hour
	progress := float64(weeklyTime) / float64(weeklyGoal)
	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}

func (s *ProgressTrackingService) calculateMonthlyGoalProgress(history []*entities.LearningHistory) float64 {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	var monthlyTime time.Duration
	for _, h := range history {
		if h.Timestamp.After(monthStart) {
			monthlyTime += h.Duration
		}
	}

	// 40
	monthlyGoal := 40 * time.Hour
	progress := float64(monthlyTime) / float64(monthlyGoal)
	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}

func (s *ProgressTrackingService) calculateContentPerformanceScore(cp *entities.ContentProgress) float64 {
	// 
	progressScore := cp.Progress
	
	// 
	timeEfficiencyScore := 1.0
	if cp.TimeSpent > 0 {
		expectedTime := 3600 // ?
		timeEfficiencyScore = math.Min(1.0, float64(expectedTime)/float64(cp.TimeSpent))
	}

	// 
	interactionScore := math.Min(1.0, float64(len(cp.InteractionLog))/10.0)

	// 
	return (progressScore*0.5 + timeEfficiencyScore*0.3 + interactionScore*0.2)
}

func (s *ProgressTrackingService) inferSkillsFromContent(contentID uuid.UUID) []string {
	// ID?
	return []string{"problem_solving", "critical_thinking", "communication"}
}

func (s *ProgressTrackingService) analyzeOptimalStudyTime(history []*entities.LearningHistory) []TimeSlotAnalysis {
	hourlyStats := make(map[int]*TimeSlotAnalysis)

	// ?4?
	for i := 0; i < 24; i++ {
		hourlyStats[i] = &TimeSlotAnalysis{
			Hour:            i,
			PerformanceScore: 0,
			EngagementLevel: 0,
			CompletionRate:  0,
		}
	}

	// 
	for _, h := range history {
		hour := h.Timestamp.Hour()
		stats := hourlyStats[hour]
		
		if h.Score != nil {
			stats.PerformanceScore += *h.Score
		}
		stats.EngagementLevel += float64(h.Duration.Minutes()) / 60.0 // ?
		if h.Progress >= 1.0 {
			stats.CompletionRate += 1.0
		}
	}

	// ?
	for _, stats := range hourlyStats {
		count := 0
		for _, h := range history {
			if h.Timestamp.Hour() == stats.Hour {
				count++
			}
		}
		if count > 0 {
			stats.PerformanceScore /= float64(count)
			stats.EngagementLevel /= float64(count)
			stats.CompletionRate /= float64(count)
		}
	}

	// ?
	result := make([]TimeSlotAnalysis, 0, 24)
	for i := 0; i < 24; i++ {
		result = append(result, *hourlyStats[i])
	}

	return result
}

func (s *ProgressTrackingService) analyzePreferredContentTypes(history []*entities.LearningHistory) map[string]float64 {
	// 
	contentTypes := map[string]float64{
		"video":     0.3,
		"text":      0.4,
		"quiz":      0.2,
		"practice":  0.1,
	}

	return contentTypes
}

func (s *ProgressTrackingService) calculateLearningVelocity(history []*entities.LearningHistory) float64 {
	if len(history) < 2 {
		return 0
	}

	// /?
	totalProgress := 0.0
	for _, h := range history {
		totalProgress += h.Progress
	}

	timeSpan := history[len(history)-1].Timestamp.Sub(history[0].Timestamp)
	if timeSpan.Hours() == 0 {
		return 0
	}

	return totalProgress / timeSpan.Hours()
}

func (s *ProgressTrackingService) calculateRetentionRate(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) float64 {
	// ?
	completedCount := 0
	for _, cp := range contentProgress {
		if cp.IsCompleted {
			completedCount++
		}
	}

	if len(contentProgress) == 0 {
		return 0
	}

	return float64(completedCount) / float64(len(contentProgress))
}

func (s *ProgressTrackingService) analyzeEngagementPatterns(history []*entities.LearningHistory) []EngagementPattern {
	patterns := []EngagementPattern{
		{
			Pattern:     "consistent_daily",
			Frequency:   0.7,
			Impact:      0.8,
			Description: "",
		},
		{
			Pattern:     "weekend_intensive",
			Frequency:   0.3,
			Impact:      0.6,
			Description: "",
		},
	}

	return patterns
}

func (s *ProgressTrackingService) analyzeDropoffPoints(contentProgress []*entities.ContentProgress) []DropoffAnalysis {
	dropoffs := []DropoffAnalysis{
		{
			ContentType: "video",
			Position:    30,
			Frequency:   0.4,
			Reasons:     []string{"", "?},
		},
		{
			ContentType: "quiz",
			Position:    50,
			Frequency:   0.3,
			Reasons:     []string{"", ""},
		},
	}

	return dropoffs
}

func (s *ProgressTrackingService) calculateImprovementRate(history []*entities.LearningHistory) float64 {
	if len(history) < 2 {
		return 0
	}

	// ?
	firstScore := history[0].Score
	lastScore := history[len(history)-1].Score

	if firstScore == nil || *firstScore == 0 {
		return 0
	}

	if lastScore == nil {
		return 0
	}

	return (*lastScore - *firstScore) / *firstScore
}

func (s *ProgressTrackingService) calculateConsistencyScore(history []*entities.LearningHistory) float64 {
	if len(history) < 2 {
		return 0
	}

	// ?
	var intervals []float64
	for i := 1; i < len(history); i++ {
		interval := history[i].Timestamp.Sub(history[i-1].Timestamp).Hours()
		intervals = append(intervals, interval)
	}

	// ?
	mean := 0.0
	for _, interval := range intervals {
		mean += interval
	}
	mean /= float64(len(intervals))

	variance := 0.0
	for _, interval := range intervals {
		variance += math.Pow(interval-mean, 2)
	}
	variance /= float64(len(intervals))

	stdDev := math.Sqrt(variance)

	// 
	return math.Max(0, 1.0-stdDev/24.0) // 0-1
}

func (s *ProgressTrackingService) calculateEfficiencyScore(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) float64 {
	if len(history) == 0 {
		return 0
	}

	// ??
	totalProgress := 0.0
	totalTime := time.Duration(0)

	for _, h := range history {
		totalProgress += h.Progress
		totalTime += h.Duration
	}

	if totalTime.Hours() == 0 {
		return 0
	}

	efficiency := totalProgress / totalTime.Hours()
	return math.Min(1.0, efficiency) // 0-1
}

func (s *ProgressTrackingService) calculateEngagementScore(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0
	}

	// 
	totalSessions := len(history)
	totalTime := time.Duration(0)

	for _, h := range history {
		totalTime += h.Duration
	}

	avgSessionTime := totalTime / time.Duration(totalSessions)
	
	// 30-60
	idealTime := 45 * time.Minute
	timeDiff := math.Abs(avgSessionTime.Minutes() - idealTime.Minutes())
	timeScore := math.Max(0, 1.0-timeDiff/60.0)

	// ?
	now := time.Now()
	recentSessions := 0
	for _, h := range history {
		if now.Sub(h.Timestamp).Hours() < 168 { // 
			recentSessions++
		}
	}
	frequencyScore := math.Min(1.0, float64(recentSessions)/7.0) // 

	return (timeScore + frequencyScore) / 2.0
}

func (s *ProgressTrackingService) calculateRetentionScore(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) float64 {
	// 㱣?
	reviewCount := 0
	for _, h := range history {
		// ?
		contentReviewCount := 0
		for _, h2 := range history {
			if h2.ContentID == h.ContentID && h2.Timestamp.After(h.Timestamp) {
				contentReviewCount++
			}
		}
		if contentReviewCount > 0 {
			reviewCount++
		}
	}

	if len(history) == 0 {
		return 0
	}

	return float64(reviewCount) / float64(len(history))
}

func (s *ProgressTrackingService) calculateGoalProgress(goal entities.LearningGoal, progress OverallProgress) float64 {
	// ?
	if goal.TargetSkill == "completion" {
		return progress.CompletionRate
	} else if goal.TargetSkill == "time" {
		// ?
		return math.Min(1.0, progress.TotalTimeSpent.Hours()/100.0) // ?00
	} else {
		// ?
		return float64(progress.SkillsAcquired) / 10.0 // ?0?
	}
}

func (s *ProgressTrackingService) isGoalOnTrack(goal entities.LearningGoal, currentProgress float64) bool {
	now := time.Now()
	totalDuration := goal.TargetDate.Sub(goal.CreatedAt)
	elapsed := now.Sub(goal.CreatedAt)
	
	expectedProgress := float64(elapsed) / float64(totalDuration)
	
	// ?0%
	return currentProgress >= expectedProgress*0.8
}

func (s *ProgressTrackingService) generateGoalRecommendations(goal entities.LearningGoal, currentProgress float64, isOnTrack bool) []string {
	recommendations := make([]string, 0)

	if !isOnTrack {
		recommendations = append(recommendations, "?)
		recommendations = append(recommendations, "")
	}

	if currentProgress < 0.5 {
		recommendations = append(recommendations, "?)
		recommendations = append(recommendations, "?)
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "?)
	}

	return recommendations
}

// ... existing code ...

