package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
)

// LearningPathService 
type LearningPathService struct {
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	pathRepo           repositories.LearningPathRepository
	analyticsService   interfaces.LearningAnalyticsService
}

// NewLearningPathService 
func NewLearningPathService(
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	pathRepo repositories.LearningPathRepository,
	analyticsService interfaces.LearningAnalyticsService,
) *LearningPathService {
	return &LearningPathService{
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
		knowledgeGraphRepo: knowledgeGraphRepo,
		pathRepo:           pathRepo,
		analyticsService:   analyticsService,
	}
}

// GeneratePersonalizedPathRequest 
type GeneratePersonalizedPathRequest struct {
	LearnerID       uuid.UUID `json:"learner_id" binding:"required"`
	TargetSkills    []string  `json:"target_skills" binding:"required"`
	TimeConstraint  int       `json:"time_constraint,omitempty"` // 
	DifficultyLevel string    `json:"difficulty_level,omitempty"`
	LearningStyle   string    `json:"learning_style,omitempty"`
	Prerequisites   []string  `json:"prerequisites,omitempty"`
}

// LearningPathResponse 
type LearningPathResponse struct {
	ID                uuid.UUID                    `json:"id"`
	LearnerID         uuid.UUID                    `json:"learner_id"`
	Title             string                       `json:"title"`
	Description       string                       `json:"description"`
	EstimatedDuration int                          `json:"estimated_duration"` // 
	DifficultyLevel   string                       `json:"difficulty_level"`
	Steps             []LearningPathStepResponse   `json:"steps"`
	Milestones        []LearningMilestoneResponse  `json:"milestones"`
	Progress          float64                      `json:"progress"`
	Status            string                       `json:"status"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

// LearningPathStepResponse 
type LearningPathStepResponse struct {
	ID              uuid.UUID `json:"id"`
	Order           int       `json:"order"`
	ContentID       uuid.UUID `json:"content_id"`
	ContentTitle    string    `json:"content_title"`
	ContentType     string    `json:"content_type"`
	EstimatedTime   int       `json:"estimated_time"` // 
	Prerequisites   []string  `json:"prerequisites"`
	LearningGoals   []string  `json:"learning_goals"`
	IsCompleted     bool      `json:"is_completed"`
	CompletionRate  float64   `json:"completion_rate"`
	LastAccessedAt  *time.Time `json:"last_accessed_at,omitempty"`
}

// LearningMilestoneResponse ?
type LearningMilestoneResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TargetStep  int       `json:"target_step"`
	IsAchieved  bool      `json:"is_achieved"`
	AchievedAt  *time.Time `json:"achieved_at,omitempty"`
	Reward      string    `json:"reward,omitempty"`
}

// UpdatePathProgressRequest 
type UpdatePathProgressRequest struct {
	PathID    uuid.UUID `json:"path_id" binding:"required"`
	StepID    uuid.UUID `json:"step_id" binding:"required"`
	Progress  float64   `json:"progress" binding:"min=0,max=1"`
	TimeSpent int       `json:"time_spent"` // 
}

// PathRecommendationRequest 
type PathRecommendationRequest struct {
	LearnerID      uuid.UUID `json:"learner_id" binding:"required"`
	CurrentSkills  []string  `json:"current_skills"`
	InterestAreas  []string  `json:"interest_areas"`
	AvailableTime  int       `json:"available_time,omitempty"` // ?
	LearningGoals  []string  `json:"learning_goals"`
}

// PathRecommendationResponse 
type PathRecommendationResponse struct {
	RecommendedPaths []RecommendedPath `json:"recommended_paths"`
	Reasoning        string            `json:"reasoning"`
	Confidence       float64           `json:"confidence"`
}

// RecommendedPath 
type RecommendedPath struct {
	PathID          uuid.UUID `json:"path_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	MatchScore      float64   `json:"match_score"`
	EstimatedTime   int       `json:"estimated_time"`
	DifficultyLevel string    `json:"difficulty_level"`
	SkillsGained    []string  `json:"skills_gained"`
	Reasons         []string  `json:"reasons"`
}

// GeneratePersonalizedPath 
func (s *LearningPathService) GeneratePersonalizedPath(ctx context.Context, req *GeneratePersonalizedPathRequest) (*LearningPathResponse, error) {
	// ?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// ?
	currentSkills := s.extractCurrentSkills(learner)
	
	// ?
	knowledgePath, err := s.buildKnowledgePath(ctx, req.TargetSkills, currentSkills)
	if err != nil {
		return nil, fmt.Errorf("failed to build knowledge path: %w", err)
	}

	// 
	pathSteps, err := s.selectOptimalContent(ctx, knowledgePath, learner, req)
	if err != nil {
		return nil, fmt.Errorf("failed to select content: %w", err)
	}

	// 
	path := &entities.LearningPath{
		ID:              uuid.New(),
		Name:            s.generatePathTitle(req.TargetSkills),
		Description:     s.generatePathDescription(req.TargetSkills, len(pathSteps)),
		Subject:         s.determineSubject(req.TargetSkills),
		DifficultyLevel: s.determineDifficultyLevel(req.DifficultyLevel, learner),
		EstimatedHours:  s.calculateEstimatedDuration(pathSteps),
		Prerequisites:   []uuid.UUID{},
		LearningGoals:   req.TargetSkills,
		Nodes:           s.convertToPathNodes(pathSteps),
		Milestones:      []entities.Milestone{},
		Tags:            []string{},
		IsPublic:        false,
		CreatedBy:       req.LearnerID,
		EnrollmentCount: 1,
		CompletionRate:  0.0,
		Rating:          0.0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// ?repositories ?
	repoPath := s.convertToRepositoryLearningPath(path)
	repoPath.LearnerID = req.LearnerID

	// 
	if err := s.pathRepo.Create(ctx, repoPath); err != nil {
		return nil, fmt.Errorf("failed to create learning path: %w", err)
	}

	// ?
	milestones := s.generateMilestones(pathSteps)

	return &LearningPathResponse{
		ID:                path.ID,
		LearnerID:         req.LearnerID,
		Title:             path.Name,
		Description:       path.Description,
		EstimatedDuration: int(path.EstimatedHours),
		DifficultyLevel:   s.difficultyLevelToString(path.DifficultyLevel),
		Steps:             pathSteps,
		Milestones:        milestones,
		Progress:          0.0,
		Status:            "active",
		CreatedAt:         path.CreatedAt,
		UpdatedAt:         path.UpdatedAt,
	}, nil
}

// UpdatePathProgress 
func (s *LearningPathService) UpdatePathProgress(ctx context.Context, req *UpdatePathProgressRequest) error {
	// 
	_, err := s.pathRepo.GetByID(ctx, req.PathID)
	if err != nil {
		return fmt.Errorf("failed to get learning path: %w", err)
	}

	// 
	// 
	
	// 
	if s.analyticsService != nil {
		analyticsData := map[string]interface{}{
			"path_id":    req.PathID,
			"step_id":    req.StepID,
			"progress":   req.Progress,
			"time_spent": req.TimeSpent,
			"timestamp":  time.Now(),
		}
		
		// TODO: 
		// 
		go func() {
			// 
			fmt.Printf("Learning activity recorded: %+v\n", analyticsData)
		}()
	}

	return nil
}

// GetRecommendedPaths 
func (s *LearningPathService) GetRecommendedPaths(ctx context.Context, req *PathRecommendationRequest) (*PathRecommendationResponse, error) {
	// ?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// ?
	learnerProfile := s.analyzeLearnerProfile(learner, req)
	
	// ?
	candidatePaths, err := s.getCandidatePaths(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidate paths: %w", err)
	}

	// 
	recommendedPaths := s.calculateMatchScores(candidatePaths, learnerProfile)
	
	// N?
	sort.Slice(recommendedPaths, func(i, j int) bool {
		return recommendedPaths[i].MatchScore > recommendedPaths[j].MatchScore
	})

	if len(recommendedPaths) > 5 {
		recommendedPaths = recommendedPaths[:5]
	}

	return &PathRecommendationResponse{
		RecommendedPaths: recommendedPaths,
		Reasoning:        s.generateRecommendationReasoning(learnerProfile),
		Confidence:       s.calculateRecommendationConfidence(recommendedPaths),
	}, nil
}

// 

func (s *LearningPathService) extractCurrentSkills(learner *entities.Learner) []string {
	var skills []string
	for _, skill := range learner.Skills {
		if skill.Level >= 5 { // 5 
			skills = append(skills, skill.SkillName)
		}
	}
	return skills
}

func (s *LearningPathService) buildKnowledgePath(ctx context.Context, targetSkills, currentSkills []string) ([]string, error) {
	// ?
	// ?
	return targetSkills, nil
}

func (s *LearningPathService) selectOptimalContent(ctx context.Context, knowledgePath []string, learner *entities.Learner, req *GeneratePersonalizedPathRequest) ([]LearningPathStepResponse, error) {
	var steps []LearningPathStepResponse
	
	for i, skill := range knowledgePath {
		// 
		contents, err := s.contentRepo.SearchByKeywords(ctx, []string{skill}, 0, 10)
		if err != nil {
			continue
		}

		// ?
		if len(contents) > 0 {
			content := s.selectBestContent(contents, learner)
			step := LearningPathStepResponse{
				ID:             uuid.New(),
				Order:          i + 1,
				ContentID:      content.ID,
				ContentTitle:   content.Title,
				ContentType:    string(content.Type),
				EstimatedTime:  content.EstimatedDuration,
				Prerequisites:  []string{},
				LearningGoals:  []string{skill},
				IsCompleted:    false,
				CompletionRate: 0.0,
			}
			steps = append(steps, step)
		}
	}

	return steps, nil
}

func (s *LearningPathService) selectBestContent(contents []*entities.LearningContent, learner *entities.Learner) *entities.LearningContent {
	// ?
	// 
	if len(contents) > 0 {
		return contents[0]
	}
	return nil
}

func (s *LearningPathService) generatePathTitle(targetSkills []string) string {
	if len(targetSkills) == 1 {
		return fmt.Sprintf("%s ", targetSkills[0])
	}
	return fmt.Sprintf("?(%d?", len(targetSkills))
}

func (s *LearningPathService) generatePathDescription(targetSkills []string, stepCount int) string {
	return fmt.Sprintf("?d%v", stepCount, targetSkills)
}

func (s *LearningPathService) calculateEstimatedDuration(steps []LearningPathStepResponse) float64 {
	total := 0
	for _, step := range steps {
		total += step.EstimatedTime
	}
	return float64(total) / 60.0 // ?
}

func (s *LearningPathService) determineDifficultyLevel(requested string, learner *entities.Learner) entities.DifficultyLevel {
	if requested != "" {
		// ?DifficultyLevel
		switch strings.ToLower(requested) {
		case "beginner", "1":
			return entities.DifficultyBeginner
		case "elementary", "2":
			return entities.DifficultyElementary
		case "intermediate", "3":
			return entities.DifficultyIntermediate
		case "advanced", "4":
			return entities.DifficultyAdvanced
		case "expert", "5":
			return entities.DifficultyExpert
		default:
			return entities.DifficultyBeginner
		}
	}
	
	// ?
	avgLevel := s.calculateAverageSkillLevel(learner)
	if avgLevel >= 4 {
		return entities.DifficultyAdvanced
	} else if avgLevel >= 3 {
		return entities.DifficultyIntermediate
	}
	return entities.DifficultyBeginner
}

func (s *LearningPathService) calculateAverageSkillLevel(learner *entities.Learner) float64 {
	if len(learner.Skills) == 0 {
		return 1.0
	}
	
	total := 0.0
	for _, skill := range learner.Skills {
		total += float64(skill.Level)
	}
	return total / float64(len(learner.Skills))
}

func (s *LearningPathService) generateMilestones(steps []LearningPathStepResponse) []LearningMilestoneResponse {
	var milestones []LearningMilestoneResponse
	
	// ?5%
	stepCount := len(steps)
	milestonePoints := []int{stepCount / 4, stepCount / 2, stepCount * 3 / 4, stepCount}
	
	for i, point := range milestonePoints {
		if point > 0 && point <= stepCount {
			milestone := LearningMilestoneResponse{
				ID:          uuid.New(),
				Title:       fmt.Sprintf("?%d", i+1),
				Description: fmt.Sprintf("?%d ?, point),
				TargetStep:  point,
				IsAchieved:  false,
			}
			milestones = append(milestones, milestone)
		}
	}
	
	return milestones
}

func (s *LearningPathService) analyzeLearnerProfile(learner *entities.Learner, req *PathRecommendationRequest) map[string]interface{} {
	return map[string]interface{}{
		"learning_style":   learner.LearningStyle,
		"current_skills":   req.CurrentSkills,
		"interest_areas":   req.InterestAreas,
		"available_time":   req.AvailableTime,
		"learning_goals":   req.LearningGoals,
		"skill_level":      s.calculateAverageSkillLevel(learner),
	}
}

func (s *LearningPathService) getCandidatePaths(ctx context.Context, req *PathRecommendationRequest) ([]RecommendedPath, error) {
	// 
	// 
	return []RecommendedPath{}, nil
}

func (s *LearningPathService) calculateMatchScores(paths []RecommendedPath, profile map[string]interface{}) []RecommendedPath {
	// ?
	for i := range paths {
		paths[i].MatchScore = 0.8 // 
	}
	return paths
}

func (s *LearningPathService) generateRecommendationReasoning(profile map[string]interface{}) string {
	return "?
}

// ?
func (s *LearningPathService) determineSubject(targetSkills []string) string {
	if len(targetSkills) == 0 {
		return "General Learning"
	}
	return targetSkills[0] // ?
}

func (s *LearningPathService) convertToPathNodes(pathSteps []LearningPathStepResponse) []entities.PathNode {
	nodes := make([]entities.PathNode, len(pathSteps))
	for i, step := range pathSteps {
		nodes[i] = entities.PathNode{
			ID:           uuid.New(),
			KnowledgeID:  step.ContentID, //  ContentID  KnowledgeID
			Order:        i,
			IsOptional:   false,
			Dependencies: []uuid.UUID{},
			Metadata: map[string]interface{}{
				"is_completed":     step.IsCompleted,
				"last_accessed_at": step.LastAccessedAt,
				"estimated_time":   step.EstimatedTime,
			},
		}
	}
	return nodes
}

func (s *LearningPathService) calculateRecommendationConfidence(paths []RecommendedPath) float64 {
	if len(paths) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, path := range paths {
		totalScore += path.MatchScore
	}
	return totalScore / float64(len(paths))
}

// convertToRepositoryLearningPath ?repositories  LearningPath
func (s *LearningPathService) convertToRepositoryLearningPath(path *entities.LearningPath) *repositories.LearningPath {
	// 
	nodes := make([]repositories.PathNode, len(path.Nodes))
	for i, node := range path.Nodes {
		nodes[i] = repositories.PathNode{
			ID:          node.ID,
			ContentID:   node.KnowledgeID,
			Position:    node.Order,
			IsCompleted: false,
			CompletedAt: nil,
		}
	}

	return &repositories.LearningPath{
		ID:          path.ID,
		Title:       path.Name,
		Description: path.Description,
		LearnerID:   uuid.Nil, // ?
		GraphID:     uuid.Nil, // ID
		Nodes:       nodes,
		Edges:       []repositories.PathEdge{}, // 
		Metadata: map[string]interface{}{
			"subject":           path.Subject,
			"difficulty_level":  s.difficultyLevelToString(path.DifficultyLevel),
			"estimated_hours":   path.EstimatedHours,
			"prerequisites":     path.Prerequisites,
			"learning_goals":    path.LearningGoals,
			"tags":             path.Tags,
			"is_public":        path.IsPublic,
			"created_by":       path.CreatedBy,
			"enrollment_count": path.EnrollmentCount,
			"completion_rate":  path.CompletionRate,
			"rating":           path.Rating,
		},
		IsActive:  true,
		CreatedAt: path.CreatedAt,
		UpdatedAt: path.UpdatedAt,
	}
}

// difficultyLevelToString ?DifficultyLevel 
func (s *LearningPathService) difficultyLevelToString(level entities.DifficultyLevel) string {
	switch level {
	case entities.DifficultyBeginner:
		return "beginner"
	case entities.DifficultyIntermediate:
		return "intermediate"
	case entities.DifficultyAdvanced:
		return "advanced"
	case entities.DifficultyExpert:
		return "expert"
	default:
		return "beginner"
	}
}

