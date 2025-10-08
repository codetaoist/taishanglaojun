package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)



// ProgressTrackingService 学习进度追踪服务
type ProgressTrackingService struct {
	learnerRepo         repositories.LearnerRepository
	contentRepo         repositories.LearningContentRepository
	knowledgeGraphRepo  repositories.KnowledgeGraphRepository
	analyticsService    LearningAnalyticsService
}

// NewProgressTrackingService 创建新的进度追踪服务
func NewProgressTrackingService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	analyticsService LearningAnalyticsService,
) *ProgressTrackingService {
	return &ProgressTrackingService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		analyticsService:   analyticsService,
	}
}

// ProgressUpdateRequest 进度更新请求
type ProgressUpdateRequest struct {
	LearnerID       uuid.UUID              `json:"learner_id" validate:"required"`
	ContentID       uuid.UUID              `json:"content_id" validate:"required"`
	Progress        float64                `json:"progress" validate:"min=0,max=1"`
	TimeSpent       int                    `json:"time_spent"` // 秒
	LastPosition    int                    `json:"last_position"`
	InteractionData map[string]interface{} `json:"interaction_data"`
	QuizResults     []QuizResult           `json:"quiz_results,omitempty"`
	Notes           []NoteData             `json:"notes,omitempty"`
	Bookmarks       []BookmarkData         `json:"bookmarks,omitempty"`
}

// QuizResult 测验结果
type QuizResult struct {
	QuestionID    uuid.UUID `json:"question_id"`
	Answer        interface{} `json:"answer"`
	IsCorrect     bool      `json:"is_correct"`
	Score         float64   `json:"score"`
	TimeSpent     int       `json:"time_spent"`
	AttemptCount  int       `json:"attempt_count"`
}

// NoteData 笔记数据
type NoteData struct {
	Content   string   `json:"content"`
	Position  int      `json:"position"`
	Tags      []string `json:"tags"`
	IsPublic  bool     `json:"is_public"`
}

// BookmarkData 书签数据
type BookmarkData struct {
	Title    string `json:"title"`
	Position int    `json:"position"`
	Note     string `json:"note"`
}

// ProgressResponse 进度响应
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

// NextStepRecommendation 下一步推荐
type NextStepRecommendation struct {
	Type        string    `json:"type"` // "content", "review", "practice", "assessment"
	ContentID   uuid.UUID `json:"content_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	Reason      string    `json:"reason"`
}

// Achievement 成就
type ProgressAchievement struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Points      int       `json:"points"`
	UnlockedAt  time.Time `json:"unlocked_at"`
}

// LearningReport 学习报告
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

// ReportPeriod 报告周期
type ReportPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Type      string    `json:"type"` // "daily", "weekly", "monthly", "custom"
}

// OverallProgress 总体进度
type OverallProgress struct {
	CompletionRate      float64       `json:"completion_rate"`
	TotalTimeSpent      time.Duration `json:"total_time_spent"`
	ContentCompleted    int           `json:"content_completed"`
	SkillsAcquired      int           `json:"skills_acquired"`
	CurrentStreak       int           `json:"current_streak"`
	WeeklyGoalProgress  float64       `json:"weekly_goal_progress"`
	MonthlyGoalProgress float64       `json:"monthly_goal_progress"`
}

// ContentProgressSummary 内容进度摘要
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

// SkillProgress 技能进度
type SkillProgress struct {
	SkillName       string    `json:"skill_name"`
	PreviousLevel   float64   `json:"previous_level"`
	CurrentLevel    float64   `json:"current_level"`
	Improvement     float64   `json:"improvement"`
	LastUpdated     time.Time `json:"last_updated"`
	RelatedContent  []uuid.UUID `json:"related_content"`
}

// LearningPatternAnalysis 学习模式分析
type LearningPatternAnalysis struct {
	OptimalStudyTime    []TimeSlotAnalysis `json:"optimal_study_time"`
	PreferredContentTypes map[string]float64 `json:"preferred_content_types"`
	LearningVelocity    float64            `json:"learning_velocity"`
	RetentionRate       float64            `json:"retention_rate"`
	EngagementPatterns  []EngagementPattern `json:"engagement_patterns"`
	DropoffPoints       []DropoffAnalysis   `json:"dropoff_points"`
}

// TimeSlotAnalysis 时间段分析
type TimeSlotAnalysis struct {
	Hour            int     `json:"hour"`
	PerformanceScore float64 `json:"performance_score"`
	EngagementLevel float64 `json:"engagement_level"`
	CompletionRate  float64 `json:"completion_rate"`
}

// EngagementPattern 参与模式
type EngagementPattern struct {
	Pattern     string  `json:"pattern"`
	Frequency   float64 `json:"frequency"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
}

// DropoffAnalysis 流失分析
type DropoffAnalysis struct {
	ContentType string  `json:"content_type"`
	Position    int     `json:"position"` // 百分比
	Frequency   float64 `json:"frequency"`
	Reasons     []string `json:"reasons"`
}

// PerformanceMetrics 性能指标
type ProgressPerformanceMetrics struct {
	AverageScore        float64 `json:"average_score"`
	ImprovementRate     float64 `json:"improvement_rate"`
	ConsistencyScore    float64 `json:"consistency_score"`
	EfficiencyScore     float64 `json:"efficiency_score"`
	EngagementScore     float64 `json:"engagement_score"`
	RetentionScore      float64 `json:"retention_score"`
}

// RecommendationItem 推荐项
type RecommendationItem struct {
	Type        string    `json:"type"`
	Priority    int       `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ActionItems []string  `json:"action_items"`
	ExpectedImpact string `json:"expected_impact"`
}

// GoalProgress 目标进度
type GoalProgress struct {
	GoalID          uuid.UUID `json:"goal_id"`
	Description     string    `json:"description"`
	TargetDate      time.Time `json:"target_date"`
	CurrentProgress float64   `json:"current_progress"`
	IsOnTrack       bool      `json:"is_on_track"`
	DaysRemaining   int       `json:"days_remaining"`
	Recommendations []string  `json:"recommendations"`
}

// UpdateProgress 更新学习进度
func (s *ProgressTrackingService) UpdateProgress(ctx context.Context, req *ProgressUpdateRequest) (*ProgressResponse, error) {
	// 获取或创建内容进度记录
	progress, err := s.getOrCreateContentProgress(ctx, req.LearnerID, req.ContentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content progress: %w", err)
	}

	// 更新进度数据
	progress.UpdateProgress(req.Progress, req.LastPosition, req.TimeSpent)

	// 处理测验结果
	if len(req.QuizResults) > 0 {
		s.processQuizResults(progress, req.QuizResults)
	}

	// 处理笔记
	for _, noteData := range req.Notes {
		progress.AddNote(noteData.Content, noteData.Position, noteData.Tags, noteData.IsPublic)
	}

	// 处理书签
	for _, bookmarkData := range req.Bookmarks {
		progress.AddBookmark(bookmarkData.Title, bookmarkData.Position, bookmarkData.Note)
	}

	// 记录交互数据
	if len(req.InteractionData) > 0 {
		progress.RecordInteraction("update", "progress", req.LastPosition, req.InteractionData)
	}

	// 保存进度到学习历史
	err = s.saveProgressToLearningHistory(ctx, req, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to save progress: %w", err)
	}

	// 更新学习者统计信息
	err = s.updateLearnerStatistics(ctx, req.LearnerID, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to update learner statistics: %w", err)
	}

	// 生成响应
	response, err := s.generateProgressResponse(ctx, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

// GetLearningReport 生成学习报告
func (s *ProgressTrackingService) GetLearningReport(ctx context.Context, learnerID uuid.UUID, period ReportPeriod) (*LearningReport, error) {
	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 过滤学习历史记录
	var filteredHistory []entities.LearningHistory
	for _, h := range learner.LearningHistory {
		if h.Timestamp.After(period.StartDate) && h.Timestamp.Before(period.EndDate) {
			filteredHistory = append(filteredHistory, h)
		}
	}

	// 转换为指针切片
	var historyPointers []*entities.LearningHistory
	for i := range filteredHistory {
		historyPointers = append(historyPointers, &filteredHistory[i])
	}

	// 基于学习历史生成内容进度摘要
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

	// 分析各项指标
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

// getOrCreateContentProgress 获取或创建内容进度
func (s *ProgressTrackingService) getOrCreateContentProgress(ctx context.Context, learnerID, contentID uuid.UUID) (*entities.ContentProgress, error) {
	progress, err := s.contentRepo.GetProgress(ctx, learnerID, contentID)
	if err != nil {
		// 如果不存在，创建新的进度记录
		progress = entities.NewContentProgress(learnerID, contentID)
	}
	return progress, nil
}

// processQuizResults 处理测验结果
func (s *ProgressTrackingService) processQuizResults(progress *entities.ContentProgress, results []QuizResult) {
	for _, result := range results {
		progress.QuizScores[result.QuestionID] = result.Score
		
		// 记录交互
		interactionData := map[string]interface{}{
			"answer":        result.Answer,
			"is_correct":    result.IsCorrect,
			"score":         result.Score,
			"attempt_count": result.AttemptCount,
		}
		progress.RecordInteraction("quiz_answer", fmt.Sprintf("question_%s", result.QuestionID), 0, interactionData)
	}
}

// updateLearnerStatistics 更新学习者统计信息
func (s *ProgressTrackingService) updateLearnerStatistics(ctx context.Context, learnerID uuid.UUID, progress *entities.ContentProgress) error {
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 更新总学习时间（转换为小时并四舍五入）
	additionalHours := int(math.Round(float64(progress.TimeSpent) / 3600.0))
	learner.TotalStudyHours += additionalHours

	// 更新学习连续性
	if progress.IsCompleted {
		s.updateLearningStreak(learner, time.Now())
	}

	// 保存更新
	return s.learnerRepo.Update(ctx, learner)
}

// saveProgressToLearningHistory 保存进度到学习历史
func (s *ProgressTrackingService) saveProgressToLearningHistory(ctx context.Context, req *ProgressUpdateRequest, progress *entities.ContentProgress) error {
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return err
	}

	// 查找现有的学习历史记录
	var existingHistory *entities.LearningHistory
	for i := range learner.LearningHistory {
		if learner.LearningHistory[i].ContentID == req.ContentID {
			existingHistory = &learner.LearningHistory[i]
			break
		}
	}

	if existingHistory != nil {
		// 更新现有记录
		existingHistory.Progress = progress.Progress
		existingHistory.Duration += time.Duration(req.TimeSpent) * time.Second
		existingHistory.Completed = progress.IsCompleted
		existingHistory.Timestamp = time.Now()
		if progress.IsCompleted && existingHistory.EndTime == nil {
			now := time.Now()
			existingHistory.EndTime = &now
		}
	} else {
		// 创建新的学习历史记录
		history := entities.LearningHistory{
			ID:          uuid.New(),
			LearnerID:   req.LearnerID,
			ContentID:   req.ContentID,
			ContentType: "unknown", // 可以从content repository获取
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

// updateLearningStreak 更新学习连续性
func (s *ProgressTrackingService) updateLearningStreak(learner *entities.Learner, studyDate time.Time) {
	today := studyDate.Truncate(24 * time.Hour)
	lastStudyDate := learner.Streak.LastStudyDate.Truncate(24 * time.Hour)

	if today.Equal(lastStudyDate) {
		// 同一天，不更新
		return
	}

	if today.Equal(lastStudyDate.Add(24 * time.Hour)) {
		// 连续的下一天
		learner.Streak.CurrentStreak++
		if learner.Streak.CurrentStreak > learner.Streak.LongestStreak {
			learner.Streak.LongestStreak = learner.Streak.CurrentStreak
		}
	} else if today.After(lastStudyDate.Add(24 * time.Hour)) {
		// 中断了连续性
		learner.Streak.CurrentStreak = 1
	}

	learner.Streak.LastStudyDate = studyDate
	learner.Streak.TotalDays++
}

// generateProgressResponse 生成进度响应
func (s *ProgressTrackingService) generateProgressResponse(ctx context.Context, progress *entities.ContentProgress) (*ProgressResponse, error) {
	// 获取内容信息
	content, err := s.contentRepo.GetByID(ctx, progress.ContentID)
	if err != nil {
		return nil, err
	}

	// 计算预估剩余时间
	estimatedRemaining := s.calculateEstimatedRemainingTime(progress, content)

	// 计算性能分数
	performanceScore := s.calculatePerformanceScore(progress)

	// 评估参与度
	engagementLevel := s.assessEngagementLevel(progress)

	// 生成推荐
	recommendations := s.generateProgressRecommendations(progress, content)

	// 生成下一步推荐
	nextSteps := s.generateNextSteps(ctx, progress, content)

	// 检查成就
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

// calculateEstimatedRemainingTime 计算预估剩余时间
func (s *ProgressTrackingService) calculateEstimatedRemainingTime(progress *entities.ContentProgress, content *entities.LearningContent) time.Duration {
	if progress.Progress >= 1.0 {
		return 0
	}

	// 基于当前进度和预估总时间计算
	estimatedTotal := time.Duration(content.EstimatedDuration) * time.Minute
	if progress.Progress > 0 {
		// 基于实际学习速度调整
		actualRate := float64(progress.TimeSpent) / progress.Progress
		remaining := (1.0 - progress.Progress) * actualRate
		return time.Duration(remaining) * time.Second
	}

	return estimatedTotal
}

// calculatePerformanceScore 计算性能分数
func (s *ProgressTrackingService) calculatePerformanceScore(progress *entities.ContentProgress) float64 {
	score := 0.0
	factors := 0

	// 进度因子
	score += progress.Progress * 30
	factors++

	// 测验分数因子
	if len(progress.QuizScores) > 0 {
		totalScore := 0.0
		for _, quizScore := range progress.QuizScores {
			totalScore += quizScore
		}
		avgQuizScore := totalScore / float64(len(progress.QuizScores))
		score += avgQuizScore * 40
		factors++
	}

	// 参与度因子（基于交互记录）
	if len(progress.InteractionLog) > 0 {
		engagementScore := math.Min(float64(len(progress.InteractionLog))/10.0, 1.0) * 20
		score += engagementScore
		factors++
	}

	// 笔记和书签因子
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

// assessEngagementLevel 评估参与度
func (s *ProgressTrackingService) assessEngagementLevel(progress *entities.ContentProgress) string {
	score := 0

	// 基于交互频率
	if len(progress.InteractionLog) > 20 {
		score += 3
	} else if len(progress.InteractionLog) > 10 {
		score += 2
	} else if len(progress.InteractionLog) > 5 {
		score += 1
	}

	// 基于笔记和书签
	if len(progress.Notes) > 3 || len(progress.Bookmarks) > 2 {
		score += 2
	} else if len(progress.Notes) > 0 || len(progress.Bookmarks) > 0 {
		score += 1
	}

	// 基于学习时间
	if progress.TimeSpent > 3600 { // 超过1小时
		score += 2
	} else if progress.TimeSpent > 1800 { // 超过30分钟
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

// generateProgressRecommendations 生成进度推荐
func (s *ProgressTrackingService) generateProgressRecommendations(progress *entities.ContentProgress, content *entities.LearningContent) []string {
	var recommendations []string

	// 基于进度的推荐
	if progress.Progress < 0.3 {
		recommendations = append(recommendations, "建议制定学习计划，设定每日学习目标")
	} else if progress.Progress < 0.7 {
		recommendations = append(recommendations, "学习进度良好，建议保持当前节奏")
	} else if progress.Progress < 1.0 {
		recommendations = append(recommendations, "即将完成，建议加强复习巩固知识点")
	}

	// 基于测验表现的推荐
	if len(progress.QuizScores) > 0 {
		totalScore := 0.0
		for _, score := range progress.QuizScores {
			totalScore += score
		}
		avgScore := totalScore / float64(len(progress.QuizScores))
		
		if avgScore < 0.6 {
			recommendations = append(recommendations, "测验成绩偏低，建议重新学习相关概念")
		} else if avgScore > 0.8 {
			recommendations = append(recommendations, "测验表现优秀，可以尝试更高难度的内容")
		}
	}

	// 基于参与度的推荐
	if len(progress.InteractionLog) < 5 {
		recommendations = append(recommendations, "建议增加互动，多做练习和思考")
	}

	if len(progress.Notes) == 0 {
		recommendations = append(recommendations, "建议记录学习笔记，有助于知识巩固")
	}

	return recommendations
}

// generateNextSteps 生成下一步推荐
func (s *ProgressTrackingService) generateNextSteps(ctx context.Context, progress *entities.ContentProgress, content *entities.LearningContent) []NextStepRecommendation {
	var nextSteps []NextStepRecommendation

	if progress.Progress < 1.0 {
		// 继续当前内容
		nextSteps = append(nextSteps, NextStepRecommendation{
			Type:        "content",
			ContentID:   content.ID,
			Title:       "继续学习",
			Description: fmt.Sprintf("继续学习《%s》，当前进度 %.1f%%", content.Title, progress.Progress*100),
			Priority:    1,
			Reason:      "完成当前内容是学习路径的重要一步",
		})
	} else {
		// 推荐相关内容或下一步内容
		nextSteps = append(nextSteps, NextStepRecommendation{
			Type:        "review",
			Title:       "复习巩固",
			Description: "复习已学内容，加深理解",
			Priority:    2,
			Reason:      "巩固已学知识有助于长期记忆",
		})

		nextSteps = append(nextSteps, NextStepRecommendation{
			Type:        "practice",
			Title:       "实践练习",
			Description: "通过练习应用所学知识",
			Priority:    1,
			Reason:      "实践是检验学习效果的最佳方式",
		})
	}

	return nextSteps
}

// checkAchievements 检查成就
func (s *ProgressTrackingService) checkAchievements(ctx context.Context, progress *entities.ContentProgress) []domainServices.Achievement {
	var achievements []domainServices.Achievement

	// 完成成就
	if progress.IsCompleted {
		achievements = append(achievements, domainServices.Achievement{
			ID:          uuid.New(),
			Type:        "completion",
			Name:        "内容完成者",
			Description: "成功完成一个学习内容",
			Points:      100,
			UnlockedAt:  time.Now(),
		})
	}

	// 笔记达人
	if len(progress.Notes) >= 5 {
		achievements = append(achievements, domainServices.Achievement{
			ID:          uuid.New(),
			Type:        "note_taker",
			Name:        "笔记达人",
			Description: "在单个内容中记录了5条以上笔记",
			Points:      50,
			UnlockedAt:  time.Now(),
		})
	}

	// 测验高手
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
				Name:        "测验高手",
				Description: "测验平均分达到90%以上",
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

	// 计算总学习时间和完成内容数
	for _, h := range history {
		totalTimeSpent += h.Duration
	}

	for _, cp := range contentProgress {
		if cp.IsCompleted {
			completedContent++
		}
	}

	// 计算技能获得数（基于完成的内容）
	skillsAcquired = completedContent / 3 // 假设每3个内容获得1个技能

	// 计算学习连续性
	currentStreak = s.calculateLearningStreak(history)

	// 计算周目标和月目标进度
	weeklyGoalProgress = s.calculateWeeklyGoalProgress(history)
	monthlyGoalProgress = s.calculateMonthlyGoalProgress(history)

	// 计算完成率
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
		// 获取内容信息（这里简化处理）
		title := fmt.Sprintf("Content %s", cp.ContentID.String()[:8])
		contentType := "unknown"
		difficulty := "medium"

		// 计算性能分数
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

	// 分析学习历史中的技能发展
	for _, h := range history {
		// 根据内容类型推断技能
		skills := s.inferSkillsFromContent(h.ContentID)
		
		for _, skill := range skills {
			if progress, exists := skillProgressMap[skill]; exists {
				// 更新现有技能进度
				progress.CurrentLevel += 0.1 // 简化的技能提升计算
				progress.LastUpdated = h.Timestamp
				progress.RelatedContent = append(progress.RelatedContent, h.ContentID)
			} else {
				// 创建新的技能进度记录
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

	// 转换为切片
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

	// 分析最佳学习时间
	optimalStudyTime := s.analyzeOptimalStudyTime(history)

	// 分析偏好的内容类型
	preferredContentTypes := s.analyzePreferredContentTypes(history)

	// 计算学习速度
	learningVelocity := s.calculateLearningVelocity(history)

	// 计算保持率
	retentionRate := s.calculateRetentionRate(history, contentProgress)

	// 分析参与模式
	engagementPatterns := s.analyzeEngagementPatterns(history)

	// 分析流失点
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

	// 计算平均分数
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

	// 计算一致性分数
	consistencyScore := s.calculateConsistencyScore(history)

	// 计算效率分数
	efficiencyScore := s.calculateEfficiencyScore(history, contentProgress)

	// 计算保持分数
	retentionScore := s.calculateRetentionScore(history, contentProgress)

	return domainServices.PerformanceMetrics{
		Accuracy:       averageScore,
		Speed:          efficiencyScore,
		Efficiency:     efficiencyScore,
		CompletionRate: retentionScore,
		ErrorRate:      1.0 - averageScore, // 错误率 = 1 - 准确率
		Consistency:    consistencyScore,
		Timeline:       "recent",
		ExpectedOutcome: "improved_performance",
	}
}

func (s *ProgressTrackingService) generateRecommendations(ctx context.Context, learner *entities.Learner, progress OverallProgress, metrics domainServices.PerformanceMetrics) []RecommendationItem {
	recommendations := make([]RecommendationItem, 0)

	// 基于完成率的推荐
	if progress.CompletionRate < 0.3 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "motivation",
			Priority:    1,
			Title:       "提高学习完成率",
			Description: "您的学习完成率较低，建议制定更具体的学习计划",
			ActionItems: []string{
				"设置每日学习目标",
				"选择感兴趣的内容开始",
				"使用番茄工作法进行学习",
			},
			ExpectedImpact: "提高学习动力和完成率",
		})
	}

	// 基于学习连续性的推荐
	if progress.CurrentStreak < 3 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "habit",
			Priority:    2,
			Title:       "建立学习习惯",
			Description: "保持学习连续性有助于知识巩固",
			ActionItems: []string{
				"设置固定的学习时间",
				"从短时间学习开始",
				"设置学习提醒",
			},
			ExpectedImpact: "建立稳定的学习习惯",
		})
	}

	// 基于性能指标的推荐
	if metrics.Accuracy < 0.7 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "performance",
			Priority:    1,
			Title:       "提高学习效果",
			Description: "您的学习成绩有提升空间",
			ActionItems: []string{
				"复习之前学过的内容",
				"寻求帮助或指导",
				"调整学习方法",
			},
			ExpectedImpact: "提高学习成绩和理解深度",
		})
	}

	// 基于效率的推荐
	if metrics.Efficiency < 0.6 {
		recommendations = append(recommendations, RecommendationItem{
			Type:        "efficiency",
			Priority:    3,
			Title:       "优化学习效率",
			Description: "可以通过调整学习方法提高效率",
			ActionItems: []string{
				"尝试不同的学习技巧",
				"在最佳时间段学习",
				"减少学习时的干扰",
			},
			ExpectedImpact: "在更短时间内获得更好的学习效果",
		})
	}

	return recommendations
}

func (s *ProgressTrackingService) analyzeGoalProgress(learner *entities.Learner, progress OverallProgress) []GoalProgress {
	goalProgresses := make([]GoalProgress, 0)

	for _, goal := range learner.LearningGoals {
		// 计算目标进度
		currentProgress := s.calculateGoalProgress(goal, progress)
		
		// 判断是否按计划进行
		isOnTrack := s.isGoalOnTrack(goal, currentProgress)
		
		// 计算剩余天数
		daysRemaining := int(goal.TargetDate.Sub(time.Now()).Hours() / 24)
		
		// 生成推荐
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

	// 这里可以从数据库获取成就，现在先返回一些示例成就
	achievements = append(achievements, domainServices.Achievement{
		ID:          uuid.New(),
		Type:        "completion",
		Name:        "学习新手",
		Description: "完成第一个学习内容",
		Points:      10,
		UnlockedAt:  time.Now(),
	})

	return achievements
}

// 辅助方法实现

func (s *ProgressTrackingService) calculateLearningStreak(history []*entities.LearningHistory) int {
	if len(history) == 0 {
		return 0
	}

	// 按时间排序
	sortedHistory := make([]*entities.LearningHistory, len(history))
	copy(sortedHistory, history)

	// 简化的连续性计算
	streak := 1
	for i := len(sortedHistory) - 1; i > 0; i-- {
		current := sortedHistory[i].Timestamp
		previous := sortedHistory[i-1].Timestamp
		
		// 如果两次学习间隔超过2天，则中断连续性
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

	// 假设周目标是10小时
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

	// 假设月目标是40小时
	monthlyGoal := 40 * time.Hour
	progress := float64(monthlyTime) / float64(monthlyGoal)
	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}

func (s *ProgressTrackingService) calculateContentPerformanceScore(cp *entities.ContentProgress) float64 {
	// 基于进度、时间效率和交互质量计算性能分数
	progressScore := cp.Progress
	
	// 时间效率分数（简化计算）
	timeEfficiencyScore := 1.0
	if cp.TimeSpent > 0 {
		expectedTime := 3600 // 假设期望时间为1小时
		timeEfficiencyScore = math.Min(1.0, float64(expectedTime)/float64(cp.TimeSpent))
	}

	// 交互质量分数
	interactionScore := math.Min(1.0, float64(len(cp.InteractionLog))/10.0)

	// 综合分数
	return (progressScore*0.5 + timeEfficiencyScore*0.3 + interactionScore*0.2)
}

func (s *ProgressTrackingService) inferSkillsFromContent(contentID uuid.UUID) []string {
	// 这里应该根据内容ID查询相关技能，现在返回示例技能
	return []string{"problem_solving", "critical_thinking", "communication"}
}

func (s *ProgressTrackingService) analyzeOptimalStudyTime(history []*entities.LearningHistory) []TimeSlotAnalysis {
	hourlyStats := make(map[int]*TimeSlotAnalysis)

	// 初始化24小时的统计
	for i := 0; i < 24; i++ {
		hourlyStats[i] = &TimeSlotAnalysis{
			Hour:            i,
			PerformanceScore: 0,
			EngagementLevel: 0,
			CompletionRate:  0,
		}
	}

	// 分析每小时的学习数据
	for _, h := range history {
		hour := h.Timestamp.Hour()
		stats := hourlyStats[hour]
		
		if h.Score != nil {
			stats.PerformanceScore += *h.Score
		}
		stats.EngagementLevel += float64(h.Duration.Minutes()) / 60.0 // 转换为小时
		if h.Progress >= 1.0 {
			stats.CompletionRate += 1.0
		}
	}

	// 计算平均值
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

	// 转换为切片
	result := make([]TimeSlotAnalysis, 0, 24)
	for i := 0; i < 24; i++ {
		result = append(result, *hourlyStats[i])
	}

	return result
}

func (s *ProgressTrackingService) analyzePreferredContentTypes(history []*entities.LearningHistory) map[string]float64 {
	// 简化的内容类型偏好分析
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

	// 计算平均学习速度（内容完成数/时间）
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
	// 简化的保持率计算
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
			Description: "每日持续学习模式",
		},
		{
			Pattern:     "weekend_intensive",
			Frequency:   0.3,
			Impact:      0.6,
			Description: "周末集中学习模式",
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
			Reasons:     []string{"内容过长", "注意力分散"},
		},
		{
			ContentType: "quiz",
			Position:    50,
			Frequency:   0.3,
			Reasons:     []string{"难度过高", "缺乏准备"},
		},
	}

	return dropoffs
}

func (s *ProgressTrackingService) calculateImprovementRate(history []*entities.LearningHistory) float64 {
	if len(history) < 2 {
		return 0
	}

	// 计算分数改进率
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

	// 计算学习时间的一致性
	var intervals []float64
	for i := 1; i < len(history); i++ {
		interval := history[i].Timestamp.Sub(history[i-1].Timestamp).Hours()
		intervals = append(intervals, interval)
	}

	// 计算标准差
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

	// 一致性分数（标准差越小，一致性越高）
	return math.Max(0, 1.0-stdDev/24.0) // 标准化到0-1范围
}

func (s *ProgressTrackingService) calculateEfficiencyScore(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) float64 {
	if len(history) == 0 {
		return 0
	}

	// 计算学习效率（进度/时间）
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
	return math.Min(1.0, efficiency) // 标准化到0-1范围
}

func (s *ProgressTrackingService) calculateEngagementScore(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 0
	}

	// 基于学习频率和时长计算参与度
	totalSessions := len(history)
	totalTime := time.Duration(0)

	for _, h := range history {
		totalTime += h.Duration
	}

	avgSessionTime := totalTime / time.Duration(totalSessions)
	
	// 理想的会话时间是30-60分钟
	idealTime := 45 * time.Minute
	timeDiff := math.Abs(avgSessionTime.Minutes() - idealTime.Minutes())
	timeScore := math.Max(0, 1.0-timeDiff/60.0)

	// 频率分数（基于最近的学习活动）
	now := time.Now()
	recentSessions := 0
	for _, h := range history {
		if now.Sub(h.Timestamp).Hours() < 168 { // 一周内
			recentSessions++
		}
	}
	frequencyScore := math.Min(1.0, float64(recentSessions)/7.0) // 每天一次为满分

	return (timeScore + frequencyScore) / 2.0
}

func (s *ProgressTrackingService) calculateRetentionScore(history []*entities.LearningHistory, contentProgress []*entities.ContentProgress) float64 {
	// 基于复习行为和长期保持计算保持分数
	reviewCount := 0
	for _, h := range history {
		// 如果同一内容被多次学习，认为是复习
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
	// 根据目标技能类型计算进度
	if goal.TargetSkill == "completion" {
		return progress.CompletionRate
	} else if goal.TargetSkill == "time" {
		// 假设目标是总学习时间
		return math.Min(1.0, progress.TotalTimeSpent.Hours()/100.0) // 假设目标是100小时
	} else {
		// 默认按技能进度计算
		return float64(progress.SkillsAcquired) / 10.0 // 假设目标是10个技能
	}
}

func (s *ProgressTrackingService) isGoalOnTrack(goal entities.LearningGoal, currentProgress float64) bool {
	now := time.Now()
	totalDuration := goal.TargetDate.Sub(goal.CreatedAt)
	elapsed := now.Sub(goal.CreatedAt)
	
	expectedProgress := float64(elapsed) / float64(totalDuration)
	
	// 如果当前进度超过期望进度的80%，认为按计划进行
	return currentProgress >= expectedProgress*0.8
}

func (s *ProgressTrackingService) generateGoalRecommendations(goal entities.LearningGoal, currentProgress float64, isOnTrack bool) []string {
	recommendations := make([]string, 0)

	if !isOnTrack {
		recommendations = append(recommendations, "增加学习时间以追赶进度")
		recommendations = append(recommendations, "调整学习计划和优先级")
	}

	if currentProgress < 0.5 {
		recommendations = append(recommendations, "专注于核心学习内容")
		recommendations = append(recommendations, "寻求额外的学习资源")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "保持当前的学习节奏")
	}

	return recommendations
}

// ... existing code ...