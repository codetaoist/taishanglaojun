package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	domainservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
)



// LearnerService ?
type LearnerService struct {
	learnerRepo         repositories.LearnerRepository
	knowledgeGraphRepo  repositories.KnowledgeGraphRepository
	learningContentRepo repositories.LearningContentRepository
	pathService         LearningPathService
	analyticsService    interfaces.LearningAnalyticsService
	knowledgeService    interfaces.KnowledgeGraphService
}

// NewLearnerService ?
func NewLearnerService(
	learnerRepo repositories.LearnerRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	learningContentRepo repositories.LearningContentRepository,
	pathService LearningPathService,
	analyticsService interfaces.LearningAnalyticsService,
	knowledgeService interfaces.KnowledgeGraphService,
) *LearnerService {
	return &LearnerService{
		learnerRepo:         learnerRepo,
		knowledgeGraphRepo:  knowledgeGraphRepo,
		learningContentRepo: learningContentRepo,
		pathService:         pathService,
		analyticsService:    analyticsService,
		knowledgeService:    knowledgeService,
	}
}

// CreateLearnerRequest ?
type CreateLearnerRequest struct {
	Name            string                 `json:"name" validate:"required,min=2,max=100"`
	Email           string                 `json:"email" validate:"required,email"`
	Age             int                    `json:"age" validate:"min=5,max=120"`
	EducationLevel  string                 `json:"education_level" validate:"required"`
	LearningStyle   string                 `json:"learning_style"`
	Goals           []LearningGoalRequest  `json:"goals"`
	Preferences     LearnerLearningPreferences    `json:"preferences"`
	InitialSkills   []SkillRequest         `json:"initial_skills"`
}

// LearningGoalRequest 
type LearningGoalRequest struct {
	Description  string    `json:"description" validate:"required"`
	TargetSkill  string    `json:"target_skill" validate:"required"`
	TargetLevel  int       `json:"target_level" validate:"min=1,max=10"`
	TargetDate   time.Time `json:"target_date" validate:"required"`
	Priority     int       `json:"priority" validate:"min=1,max=5"`
}

// LearningPreferences 
type LearnerLearningPreferences struct {
	PreferredDifficulty   string   `json:"preferred_difficulty"`
	PreferredContentTypes []string `json:"preferred_content_types"`
	StudyTimePreference   string   `json:"study_time_preference"`
	SessionDuration       int      `json:"session_duration"` // minutes
	WeeklyStudyHours      int      `json:"weekly_study_hours"`
	LearningPace          string   `json:"pace"` // entities.LearningPreferencePace
	InteractionStyle      string   `json:"interaction_style"`
}

// SkillRequest ?
type SkillRequest struct {
	Name        string `json:"name" validate:"required"`
	Level       int    `json:"level" validate:"min=1,max=10"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// LearnerResponse ?
type LearnerResponse struct {
	ID                uuid.UUID                    `json:"id"`
	Name              string                       `json:"name"`
	Email             string                       `json:"email"`
	Age               int                          `json:"age"`
	EducationLevel    string                       `json:"education_level"`
	LearningStyle     string                       `json:"learning_style"`
	Goals             []entities.LearningGoal      `json:"goals"`
	Preferences       *entities.LearningPreference `json:"preferences"`
	Skills            []entities.SkillLevel        `json:"skills"`
	Statistics        *LearnerStatistics           `json:"statistics"`
	RecommendedPaths  []*PersonalizedPath `json:"recommended_paths,omitempty"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

// LearnerStatistics ?
type LearnerStatistics struct {
	TotalStudyTime      time.Duration `json:"total_study_time"`
	CompletedContent    int           `json:"completed_content"`
	CurrentStreak       int           `json:"current_streak"`
	LongestStreak       int           `json:"longest_streak"`
	AverageSessionTime  time.Duration `json:"average_session_time"`
	WeeklyProgress      float64       `json:"weekly_progress"`
	OverallProgress     float64       `json:"overall_progress"`
	SkillsAcquired      int           `json:"skills_acquired"`
	GoalsCompleted      int           `json:"goals_completed"`
	EngagementScore     float64       `json:"engagement_score"`
}

// UpdateLearnerRequest ?
type UpdateLearnerRequest struct {
	Name           *string                   `json:"name,omitempty"`
	Age            *int                      `json:"age,omitempty"`
	EducationLevel *string                   `json:"education_level,omitempty"`
	LearningStyle  *string                   `json:"learning_style,omitempty"`
	Preferences    *LearnerLearningPreferences `json:"preferences,omitempty"`
}

// CreateLearner ?
func (s *LearnerService) CreateLearner(ctx context.Context, req *CreateLearnerRequest) (*LearnerResponse, error) {
	// ?
	userID := uuid.New() // ID
	learner := entities.NewLearner(userID, req.Name, req.Email)
	
	// 
	learner.Language = "zh-CN" // 
	learner.Timezone = "Asia/Shanghai" // 

	// 
	learner.Preferences = entities.LearningPreference{
		Style:               entities.LearningStyle(req.LearningStyle),
		Pace:                entities.LearningPaceMedium, // 
		PreferredTimeSlots:  []entities.TimeSlot{},
		SessionDuration:     60, // 60
		BreakDuration:       15, // 15
		DifficultyTolerance: 0.7, // ?
		InteractiveContent:  true,
		MultimediaContent:   true,
	}

	// ?
	for _, skillReq := range req.InitialSkills {
		skillLevel := entities.SkillLevel{
			SkillID:     uuid.New(),
			SkillName:   skillReq.Name,
			Level:       skillReq.Level,
			Experience:  0,
			Confidence:  0.5, // ?
			LastUpdated: time.Now(),
		}
		learner.Skills = append(learner.Skills, skillLevel)
	}

	// 
	for _, goalReq := range req.Goals {
		goal := entities.LearningGoal{
			ID:          uuid.New(),
			Title:       goalReq.Description, // DescriptionTitle
			Description: goalReq.Description,
			TargetSkill: goalReq.TargetSkill,
			TargetLevel: goalReq.TargetLevel,
			TargetDate:  goalReq.TargetDate,
			Priority:    goalReq.Priority,
			IsActive:    true,
			Achieved:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		learner.LearningGoals = append(learner.LearningGoals, goal)
	}

	// ?
	if err := s.learnerRepo.Create(ctx, learner); err != nil {
		return nil, fmt.Errorf("failed to create learner: %w", err)
	}

	// 
	go s.updateLearningPathRecommendations(ctx, learner)

	return s.buildLearnerResponse(learner, nil), nil
}

// GetLearner ?
func (s *LearnerService) GetLearner(ctx context.Context, learnerID uuid.UUID) (*LearnerResponse, error) {
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	return s.buildLearnerResponse(learner, nil), nil
}

// UpdateLearner ?
func (s *LearnerService) UpdateLearner(ctx context.Context, learnerID uuid.UUID, req *UpdateLearnerRequest) (*LearnerResponse, error) {
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 
	if req.Name != nil {
		learner.Name = *req.Name
	}
	if req.Age != nil {
		learner.Age = *req.Age
	}
	if req.EducationLevel != nil {
		learner.EducationLevel = *req.EducationLevel
	}
	if req.LearningStyle != nil {
		learner.LearningStyle = *req.LearningStyle
	}

	// 
	if req.Preferences != nil {
		// 
		if req.Preferences.LearningPace != "" {
			learner.Preferences.Pace = entities.LearningPace(req.Preferences.LearningPace)
		}
		if req.Preferences.SessionDuration > 0 {
			learner.Preferences.SessionDuration = req.Preferences.SessionDuration
		}
		if req.Preferences.PreferredDifficulty != "" {
			// preferred_difficultyDifficultyTolerance
			switch req.Preferences.PreferredDifficulty {
			case "beginner":
				learner.Preferences.DifficultyTolerance = 0.3
			case "intermediate":
				learner.Preferences.DifficultyTolerance = 0.6
			case "advanced":
				learner.Preferences.DifficultyTolerance = 0.9
			default:
				learner.Preferences.DifficultyTolerance = 0.7 // ?
			}
		}
		learner.Preferences.InteractiveContent = true
		learner.Preferences.MultimediaContent = true
	}

	learner.UpdatedAt = time.Now()

	if err := s.learnerRepo.Update(ctx, learner); err != nil {
		return nil, fmt.Errorf("failed to update learner: %w", err)
	}

	return s.buildLearnerResponse(learner, nil), nil
}

// DeleteLearner ?
func (s *LearnerService) DeleteLearner(ctx context.Context, learnerID uuid.UUID) error {
	if err := s.learnerRepo.Delete(ctx, learnerID); err != nil {
		return fmt.Errorf("failed to delete learner: %w", err)
	}
	return nil
}

// AddLearningGoal 
func (s *LearnerService) AddLearningGoal(ctx context.Context, learnerID uuid.UUID, req *LearningGoalRequest) (*entities.LearningGoal, error) {
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	goal := &entities.LearningGoal{
		ID:          uuid.New(),
		Title:       req.Description, // Using description as title
		Description: req.Description,
		TargetSkill: req.TargetSkill,
		TargetLevel: req.TargetLevel,
		TargetDate:  req.TargetDate,
		Priority:    req.Priority,
		IsActive:    true,
		Achieved:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.learnerRepo.AddLearnerGoal(ctx, learnerID, goal); err != nil {
		return nil, fmt.Errorf("failed to add learning goal: %w", err)
	}

	// 
	go s.updateLearningPathRecommendations(context.Background(), learner)

	return goal, nil
}

// UpdateLearningGoal 
func (s *LearnerService) UpdateLearningGoal(ctx context.Context, goalID uuid.UUID, updates map[string]interface{}) (*entities.LearningGoal, error) {
	// learner
	// repositorylearnerID
	// learnerID
	return nil, fmt.Errorf("UpdateLearningGoal method needs to be redesigned to include learnerID parameter")
}

// AddSkill ?
func (s *LearnerService) AddSkill(ctx context.Context, learnerID uuid.UUID, req *SkillRequest) (*entities.Skill, error) {
	skill := &entities.Skill{
		ID:          uuid.New(),
		Name:        req.Name,
		Level:       req.Level,
		Category:    req.Category,
		Description: req.Description,
		AcquiredAt:  time.Now(),
	}

	skillLevel := &entities.SkillLevel{
		SkillID:     skill.ID,
		SkillName:   skill.Name,
		Level:       req.Level,
		Experience:  0,
		Confidence:  0.5, // Default confidence
		LastUpdated: time.Now(),
	}

	if err := s.learnerRepo.AddOrUpdateSkill(ctx, learnerID, req.Name, skillLevel); err != nil {
		return nil, fmt.Errorf("failed to add skill: %w", err)
	}

	return skill, nil
}

// UpdateSkillLevel ?
func (s *LearnerService) UpdateSkillLevel(ctx context.Context, learnerID uuid.UUID, skillName string, newLevel int) error {
	// ?
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learnerID)
	if err != nil {
		return fmt.Errorf("failed to get learner skills: %w", err)
	}

	existingSkillLevel, exists := skills[skillName]
	if !exists {
		return fmt.Errorf("skill %s not found for learner", skillName)
	}

	// ?
	updatedSkillLevel := &entities.SkillLevel{
		SkillID:     existingSkillLevel.SkillID,
		SkillName:   existingSkillLevel.SkillName,
		Level:       newLevel,
		Experience:  existingSkillLevel.Experience,
		Confidence:  existingSkillLevel.Confidence,
		LastUpdated: time.Now(),
	}

	if err := s.learnerRepo.AddOrUpdateSkill(ctx, learnerID, skillName, updatedSkillLevel); err != nil {
		return fmt.Errorf("failed to update skill level: %w", err)
	}

	// 
	go func() {
		learner, err := s.learnerRepo.GetByID(context.Background(), learnerID)
		if err == nil {
			s.updateLearningPathRecommendations(context.Background(), learner)
		}
	}()

	return nil
}

// UpdateSkill ?
func (s *LearnerService) UpdateSkill(ctx context.Context, learnerID uuid.UUID, req *SkillRequest) (*LearnerResponse, error) {
	// ?
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner skills: %w", err)
	}

	// 鼼
	skillLevel := &entities.SkillLevel{
		SkillName:   req.Name,
		Level:       req.Level,
		Experience:  0,
		Confidence:  0.5, // Default confidence
		LastUpdated: time.Now(),
	}

	// SkillIDExperience
	if existingSkill, exists := skills[req.Name]; exists {
		skillLevel.SkillID = existingSkill.SkillID
		skillLevel.Experience = existingSkill.Experience
	} else {
		skillLevel.SkillID = uuid.New()
	}

	if err := s.learnerRepo.AddOrUpdateSkill(ctx, learnerID, req.Name, skillLevel); err != nil {
		return nil, fmt.Errorf("failed to update skill: %w", err)
	}

	// 
	go func() {
		learner, err := s.learnerRepo.GetByID(context.Background(), learnerID)
		if err == nil {
			s.updateLearningPathRecommendations(context.Background(), learner)
		}
	}()

	// 
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated learner: %w", err)
	}

	return s.buildLearnerResponse(learner, nil), nil
}

// GetLearningHistory 
func (s *LearnerService) GetLearningHistory(ctx context.Context, learnerID uuid.UUID, limit int, offset int) ([]*entities.LearningHistory, error) {
	// repositoryoffset?
	totalLimit := limit + offset
	history, err := s.learnerRepo.GetLearningHistory(ctx, learnerID, totalLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning history: %w", err)
	}
	
	// offsetlimit
	if offset >= len(history) {
		return []*entities.LearningHistory{}, nil
	}
	
	end := offset + limit
	if end > len(history) {
		end = len(history)
	}
	
	return history[offset:end], nil
}

// RecordLearningActivity 
func (s *LearnerService) RecordLearningActivity(ctx context.Context, activity *entities.LearningHistory) error {
	if err := s.learnerRepo.RecordLearningActivity(ctx, activity.LearnerID, activity); err != nil {
		return fmt.Errorf("failed to record learning activity: %w", err)
	}

	// 
	go s.updateLearningAnalytics(context.Background(), activity.LearnerID)

	return nil
}

// GetLearningAnalytics 
func (s *LearnerService) GetLearningAnalytics(ctx context.Context, learnerID uuid.UUID, timeRange string) (*domainservices.LearningAnalyticsReport, error) {
	// 
	var startTime, endTime time.Time
	now := time.Now()
	
	switch timeRange {
	case "week":
		startTime = now.AddDate(0, 0, -7)
		endTime = now
	case "month":
		startTime = now.AddDate(0, -1, 0)
		endTime = now
	case "quarter":
		startTime = now.AddDate(0, -3, 0)
		endTime = now
	default:
		startTime = now.AddDate(0, -1, 0) // 
		endTime = now
	}

	req := &domainservices.AnalyticsRequest{
		LearnerID:         learnerID,
		TimeRange:         domainservices.AnalyticsTimeRange{StartDate: startTime, EndDate: endTime},
		AnalysisType:      "comprehensive",
		Granularity:       "daily",
		IncludeComparison: true,
		ComparisonGroup:   "age_group", // 
	}

	report, err := s.analyticsService.GenerateAnalyticsReport(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analytics report: %w", err)
	}

	return report, nil
}

func (s *LearnerService) GetPersonalizedRecommendations(ctx context.Context, learnerID uuid.UUID) (*PersonalizedRecommendations, error) {
	learner, err := s.learnerRepo.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 
	pathRecommendations, err := s.pathService.GetRecommendedPaths(ctx, &PathRecommendationRequest{
		LearnerID:     learnerID,
		CurrentSkills: []string{}, // 
		InterestAreas: []string{}, // 
		AvailableTime: 10,         // 10
		LearningGoals: []string{}, // 
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get path recommendations: %w", err)
	}

	// 
	var targetSkills []string
	for _, goal := range learner.LearningGoals {
		targetSkills = append(targetSkills, goal.TargetSkill)
	}

	conceptReq := &domainservices.ConceptRecommendationRequest{
		GraphID:            uuid.New(), // ID
		LearnerID:          learnerID,
		TargetSkills:       targetSkills,
		MaxRecommendations: 10,
		IncludeReasoning:   true,
	}

	conceptRecommendations, err := s.knowledgeService.RecommendConcepts(ctx, conceptReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept recommendations: %w", err)
	}

	// 
	contentRecommendations, err := s.getContentRecommendations(ctx, learner)
	if err != nil {
		return nil, fmt.Errorf("failed to get content recommendations: %w", err)
	}

	// 
	var convertedConceptRecommendations []*ConceptRecommendation
	for _, rec := range conceptRecommendations {
		convertedConceptRecommendations = append(convertedConceptRecommendations, &ConceptRecommendation{
			ConceptID:       rec.NodeID,
			ConceptName:     rec.RecommendationType,
			Description:     rec.RecommendationType,
			RelevanceScore:  rec.Score,
			Reason:          fmt.Sprintf("Score: %.2f, Confidence: %.2f", rec.Score, rec.Confidence),
			Difficulty:      "intermediate",
			Prerequisites:   []string{},
			RelatedConcepts: []string{},
			EstimatedTime:   time.Hour,
			LearningResources: []LearningResource{},
		})
	}

	// 
	var convertedPathRecommendations []*PersonalizedPath
	for _, path := range pathRecommendations.RecommendedPaths {
		convertedPathRecommendations = append(convertedPathRecommendations, &PersonalizedPath{
			ID:                  path.PathID,
			LearnerID:           learnerID,
			Title:               path.Title,
			Description:         path.Description,
			Difficulty:          path.DifficultyLevel,
			EstimatedTime:       time.Duration(path.EstimatedTime) * time.Hour,
			Prerequisites:       []string{},
			LearningGoals:       []string{},
			Steps:               []*LearningStep{},
			Progress:            0.0,
			Status:              "recommended",
			RecommendationScore: path.MatchScore,
			CreatedAt:           time.Now(),
		})
	}

	return &PersonalizedRecommendations{
		LearnerID:              learnerID,
		PathRecommendations:    convertedPathRecommendations,
		ConceptRecommendations: convertedConceptRecommendations,
		ContentRecommendations: contentRecommendations,
		GeneratedAt:            time.Now(),
	}, nil
}

// PersonalizedRecommendations 
type PersonalizedRecommendations struct {
	LearnerID              uuid.UUID                           `json:"learner_id"`
	PathRecommendations    []*PersonalizedPath                 `json:"path_recommendations"`
	ConceptRecommendations []*ConceptRecommendation            `json:"concept_recommendations"`
	ContentRecommendations []*LearnerContentRecommendation     `json:"content_recommendations"`
	GeneratedAt            time.Time                           `json:"generated_at"`
}

// LearnerContentRecommendation ?
type LearnerContentRecommendation struct {
	ContentID   uuid.UUID `json:"content_id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Difficulty  string    `json:"difficulty"`
	Relevance   float64   `json:"relevance"`
	Reason      string    `json:"reason"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// 

func (s *LearnerService) buildLearnerResponse(learner *entities.Learner, recommendedPaths []*PersonalizedPath) *LearnerResponse {
	return &LearnerResponse{
		ID:               learner.ID,
		Name:             learner.Name,
		Email:            learner.Email,
		Age:              learner.Age,
		EducationLevel:   learner.EducationLevel,
		LearningStyle:    learner.LearningStyle,
		Goals:            learner.LearningGoals,
		Preferences:      &learner.Preferences,
		Skills:           learner.Skills,
		Statistics:       s.calculateLearnerStatistics(learner),
		RecommendedPaths: recommendedPaths,
		CreatedAt:        learner.CreatedAt,
		UpdatedAt:        learner.UpdatedAt,
	}
}

func (s *LearnerService) calculateLearnerStatistics(learner *entities.Learner) *LearnerStatistics {
	// 
	var totalStudyTime time.Duration
	var completedContent int
	for _, history := range learner.LearningHistory {
		if history.Completed {
			totalStudyTime += history.Duration
			completedContent++
		}
	}

	// 
	var averageSessionTime time.Duration
	if len(learner.LearningHistory) > 0 {
		averageSessionTime = totalStudyTime / time.Duration(len(learner.LearningHistory))
	}

	// 
	weeklyProgress := 0.0
	overallProgress := 0.0
	if len(learner.LearningGoals) > 0 {
		completedGoals := s.countCompletedGoals(learner.LearningGoals)
		overallProgress = float64(completedGoals) / float64(len(learner.LearningGoals))
		weeklyProgress = overallProgress * 0.1 // ?
	}

	return &LearnerStatistics{
		TotalStudyTime:      totalStudyTime,
		CompletedContent:    completedContent,
		CurrentStreak:       learner.Streak.CurrentStreak,
		LongestStreak:       learner.Streak.LongestStreak,
		AverageSessionTime:  averageSessionTime,
		WeeklyProgress:      weeklyProgress,
		OverallProgress:     overallProgress,
		SkillsAcquired:      len(learner.Skills),
		GoalsCompleted:      s.countCompletedGoals(learner.LearningGoals),
		EngagementScore:     0.8, // ?
	}
}

func (s *LearnerService) countCompletedGoals(goals []entities.LearningGoal) int {
	count := 0
	for _, goal := range goals {
		if goal.Achieved {
			count++
		}
	}
	return count
}

func (s *LearnerService) updateLearningPathRecommendations(ctx context.Context, learner *entities.Learner) {
	// 
	go func() {
		var targetSkills []string
		for _, goal := range learner.LearningGoals {
			targetSkills = append(targetSkills, goal.TargetSkill)
		}

		_, err := s.pathService.GetRecommendedPaths(ctx, &PathRecommendationRequest{
			LearnerID:     learner.ID,
			CurrentSkills: []string{}, // 
			InterestAreas: []string{}, // 
			AvailableTime: 10,         // 10
			LearningGoals: []string{}, // 
		})
		if err != nil {
			fmt.Printf("Warning: failed to update learning path recommendations: %v\n", err)
		}
	}()
}

func (s *LearnerService) updateLearningAnalytics(ctx context.Context, learnerID uuid.UUID) {
	// 
	req := &domainservices.AnalyticsRequest{
		LearnerID:         learnerID,
		TimeRange:         domainservices.AnalyticsTimeRange{StartDate: time.Now().AddDate(0, -1, 0), EndDate: time.Now()},
		AnalysisType:      "comprehensive",
		Granularity:       "daily",
		IncludeComparison: true,
		ComparisonGroup:   "peer",
	}

	_, err := s.analyticsService.GenerateAnalyticsReport(ctx, req)
	if err != nil {
		fmt.Printf("Warning: failed to update learning analytics: %v\n", err)
	}
}

func (s *LearnerService) getContentRecommendations(ctx context.Context, learner *entities.Learner) ([]*LearnerContentRecommendation, error) {
	// ?- ?
	recommendations := []*LearnerContentRecommendation{
		{
			ContentID:     uuid.New(),
			Title:         "Introduction to Machine Learning",
			Type:          "video",
			Difficulty:    "beginner",
			Relevance:     0.9,
			Reason:        "Matches your learning goals and current skill level",
			EstimatedTime: time.Hour * 2,
		},
		{
			ContentID:     uuid.New(),
			Title:         "Python Programming Basics",
			Type:          "interactive",
			Difficulty:    "beginner",
			Relevance:     0.85,
			Reason:        "Foundation skill for your target goals",
			EstimatedTime: time.Hour * 3,
		},
	}
	
	return recommendations, nil
}

// PersonalizedPath 
type PersonalizedPath struct {
	ID              uuid.UUID                `json:"id"`
	LearnerID       uuid.UUID                `json:"learner_id"`
	Title           string                   `json:"title"`
	Description     string                   `json:"description"`
	Difficulty      string                   `json:"difficulty"`
	EstimatedTime   time.Duration            `json:"estimated_time"`
	Prerequisites   []string                 `json:"prerequisites"`
	LearningGoals   []string                 `json:"learning_goals"`
	Steps           []*LearningStep          `json:"steps"`
	Progress        float64                  `json:"progress"`
	Status          string                   `json:"status"`
	RecommendationScore float64              `json:"recommendation_score"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

// LearningStep 
type LearningStep struct {
	ID            uuid.UUID     `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ContentType   string        `json:"content_type"`
	ContentID     uuid.UUID     `json:"content_id"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Order         int           `json:"order"`
	IsCompleted   bool          `json:"is_completed"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
}

// ConceptRecommendation 
type ConceptRecommendation struct {
	ConceptID       uuid.UUID `json:"concept_id"`
	ConceptName     string    `json:"concept_name"`
	Description     string    `json:"description"`
	Difficulty      string    `json:"difficulty"`
	Prerequisites   []string  `json:"prerequisites"`
	RelatedConcepts []string  `json:"related_concepts"`
	RelevanceScore  float64   `json:"relevance_score"`
	Reason          string    `json:"reason"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	LearningResources []LearningResource `json:"learning_resources"`
}

// LearningResource 
type LearningResource struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Difficulty  string    `json:"difficulty"`
	Duration    time.Duration `json:"duration"`
}

