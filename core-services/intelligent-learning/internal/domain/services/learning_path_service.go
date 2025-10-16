package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningPathService 
type LearningPathService struct {
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
}

// NewLearningPathService 
func NewLearningPathService(
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
) *LearningPathService {
	return &LearningPathService{
		knowledgeGraphRepo: knowledgeGraphRepo,
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
	}
}

// PathRecommendationRequest 
type PathRecommendationRequest struct {
	LearnerID       uuid.UUID   `json:"learner_id"`
	GraphID         uuid.UUID   `json:"graph_id"`
	TargetNodeID    uuid.UUID   `json:"target_node_id"`
	TargetSkills    []string    `json:"target_skills,omitempty"`
	TimeConstraint  *time.Duration `json:"time_constraint,omitempty"`
	DifficultyLimit *entities.DifficultyLevel `json:"difficulty_limit,omitempty"`
	LearningStyle   *entities.LearningStyle `json:"learning_style,omitempty"`
	Preferences     *PathPreferences `json:"preferences,omitempty"`
	MaxPaths        int         `json:"max_paths"`
}

// PathPreferences 
type PathPreferences struct {
	PreferredContentTypes []entities.ContentType `json:"preferred_content_types"`
	AvoidContentTypes     []entities.ContentType `json:"avoid_content_types"`
	MaxPathLength         int                    `json:"max_path_length"`
	MinPathLength         int                    `json:"min_path_length"`
	PreferShortPaths      bool                   `json:"prefer_short_paths"`
	AllowSkipping         bool                   `json:"allow_skipping"`
	IncludeReviews        bool                   `json:"include_reviews"`
	AdaptiveDifficulty    bool                   `json:"adaptive_difficulty"`
}

// PersonalizedPath 
type PersonalizedPath struct {
	Path                *entities.LearningPath  `json:"path"`
	PersonalizationScore float64                `json:"personalization_score"`
	EstimatedDuration   time.Duration          `json:"estimated_duration"`
	DifficultyProgression []float64            `json:"difficulty_progression"`
	SkillProgression    map[string][]float64   `json:"skill_progression"`
	SuccessProbability  float64                `json:"success_probability"`
	EngagementScore     float64                `json:"engagement_score"`
	Reasoning           []string               `json:"reasoning"`
	Adaptations         []PathAdaptation       `json:"adaptations"`
	Milestones          []PathMilestone        `json:"milestones"`
}

// PathAdaptation 
type PathAdaptation struct {
	Type        string      `json:"type"` // "difficulty", "content_type", "pacing", "style"
	Description string      `json:"description"`
	Impact      float64     `json:"impact"`
	Confidence  float64     `json:"confidence"`
}

// PathMilestone ?
type PathMilestone struct {
	NodeID          uuid.UUID     `json:"node_id"`
	Position        int           `json:"position"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	SkillsAcquired  []string      `json:"skills_acquired"`
	Prerequisites   []uuid.UUID   `json:"prerequisites"`
	Assessments     []uuid.UUID   `json:"assessments"`
	Rewards         []string      `json:"rewards"`
}

// RecommendPersonalizedPaths 
func (s *LearningPathService) RecommendPersonalizedPaths(ctx context.Context, req *PathRecommendationRequest) ([]*PersonalizedPath, error) {
	// ?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 
	graph, err := s.knowledgeGraphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	// 
	targetNode, err := s.knowledgeGraphRepo.GetNode(ctx, req.GraphID, req.TargetNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target node: %w", err)
	}

	// ?
	learnerState, err := s.analyzeLearnerState(ctx, learner, graph)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze learner state: %w", err)
	}

	// ?
	candidatePaths, err := s.generateCandidatePaths(ctx, graph, learnerState, targetNode, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidate paths: %w", err)
	}

	// ?
	personalizedPaths, err := s.personalizeAndRankPaths(ctx, candidatePaths, learner, req)
	if err != nil {
		return nil, fmt.Errorf("failed to personalize paths: %w", err)
	}

	// 
	if len(personalizedPaths) > req.MaxPaths {
		personalizedPaths = personalizedPaths[:req.MaxPaths]
	}

	return personalizedPaths, nil
}

// LearnerState ?
type LearnerState struct {
	CurrentSkills       map[string]*entities.SkillLevel `json:"current_skills"`
	MasteredNodes       []uuid.UUID                     `json:"mastered_nodes"`
	InProgressNodes     []uuid.UUID                     `json:"in_progress_nodes"`
	AvailableNodes      []uuid.UUID                     `json:"available_nodes"`
	LearningVelocity    float64                         `json:"learning_velocity"`
	PreferredDifficulty entities.DifficultyLevel        `json:"preferred_difficulty"`
	StrengthAreas       []string                        `json:"strength_areas"`
	WeaknessAreas       []string                        `json:"weakness_areas"`
	LearningPatterns    LearningPatterns                `json:"learning_patterns"`
}

// LearningPatterns 
type LearningPatterns struct {
	OptimalSessionLength time.Duration               `json:"optimal_session_length"`
	PreferredTimeSlots   []entities.TimeSlot         `json:"preferred_time_slots"`
	ContentTypePreference map[entities.ContentType]float64 `json:"content_type_preference"`
	DifficultyProgression string                     `json:"difficulty_progression"` // "gradual", "steep", "mixed"
	RetentionRate        float64                     `json:"retention_rate"`
	EngagementFactors    []string                    `json:"engagement_factors"`
}

// analyzeLearnerState ?
func (s *LearningPathService) analyzeLearnerState(ctx context.Context, learner *entities.Learner, graph *entities.KnowledgeGraph) (*LearnerState, error) {
	// ?
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learner.ID)
	if err != nil {
		return nil, err
	}

	// 
	history, err := s.learnerRepo.GetLearningHistory(ctx, learner.ID, 100)
	if err != nil {
		return nil, err
	}

	// 
	masteredNodes := s.identifyMasteredNodes(skills, graph)
	
	// ?
	availableNodes := s.identifyAvailableNodes(ctx, masteredNodes, graph)

	// 
	velocity := s.calculateLearningVelocity(history)

	// 
	patterns := s.analyzeLearningPatterns(history, learner)

	// ?
	strengthAreas, weaknessAreas := s.identifyStrengthsAndWeaknesses(skills)

	return &LearnerState{
		CurrentSkills:       skills,
		MasteredNodes:       masteredNodes,
		AvailableNodes:      availableNodes,
		LearningVelocity:    velocity,
		PreferredDifficulty: s.calculatePreferredDifficulty(history),
		StrengthAreas:       strengthAreas,
		WeaknessAreas:       weaknessAreas,
		LearningPatterns:    patterns,
	}, nil
}

// generateCandidatePaths ?
func (s *LearningPathService) generateCandidatePaths(ctx context.Context, graph *entities.KnowledgeGraph, learnerState *LearnerState, targetNode *entities.KnowledgeNode, req *PathRecommendationRequest) ([]*entities.LearningPath, error) {
	var candidatePaths []*entities.LearningPath

	// 1: 
	shortestPaths, err := s.generateShortestPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, shortestPaths...)
	}

	// 2: 
	skillBasedPaths, err := s.generateSkillBasedPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, skillBasedPaths...)
	}

	// 3: ?
	difficultyBasedPaths, err := s.generateDifficultyBasedPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, difficultyBasedPaths...)
	}

	// 4: ?
	interestBasedPaths, err := s.generateInterestBasedPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, interestBasedPaths...)
	}

	// ?
	candidatePaths = s.deduplicateAndFilterPaths(candidatePaths, req)

	return candidatePaths, nil
}

// personalizeAndRankPaths ?
func (s *LearningPathService) personalizeAndRankPaths(ctx context.Context, candidatePaths []*entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) ([]*PersonalizedPath, error) {
	var personalizedPaths []*PersonalizedPath

	for _, path := range candidatePaths {
		personalizedPath, err := s.personalizePath(ctx, path, learner, req)
		if err != nil {
			continue // ?
		}
		personalizedPaths = append(personalizedPaths, personalizedPath)
	}

	// 
	sort.Slice(personalizedPaths, func(i, j int) bool {
		return personalizedPaths[i].PersonalizationScore > personalizedPaths[j].PersonalizationScore
	})

	return personalizedPaths, nil
}

// personalizePath 
func (s *LearningPathService) personalizePath(ctx context.Context, path *entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) (*PersonalizedPath, error) {
	// 
	personalizationScore := s.calculatePersonalizationScore(path, learner, req)
	
	// 
	estimatedDuration := s.estimateLearningDuration(path, learner)
	
	// 
	difficultyProgression := s.analyzeDifficultyProgression(path)
	
	// ?
	skillProgression := s.analyzeSkillProgression(path, learner)
	
	// 
	successProbability := s.predictSuccessProbability(path, learner)
	
	// ?
	engagementScore := s.calculateEngagementScore(path, learner)
	
	// 
	reasoning := s.generateReasoning(path, learner, req)
	
	// 
	adaptations := s.identifyAdaptations(path, learner)
	
	// ?
	milestones := s.generateMilestones(path, learner)

	return &PersonalizedPath{
		Path:                path,
		PersonalizationScore: personalizationScore,
		EstimatedDuration:   estimatedDuration,
		DifficultyProgression: difficultyProgression,
		SkillProgression:    skillProgression,
		SuccessProbability:  successProbability,
		EngagementScore:     engagementScore,
		Reasoning:           reasoning,
		Adaptations:         adaptations,
		Milestones:          milestones,
	}, nil
}

// 

func (s *LearningPathService) identifyMasteredNodes(skills map[string]*entities.SkillLevel, graph *entities.KnowledgeGraph) []uuid.UUID {
	var masteredNodes []uuid.UUID
	
	for _, node := range graph.Nodes {
		if s.isNodeMastered(&node, skills) {
			masteredNodes = append(masteredNodes, node.ID)
		}
	}
	
	return masteredNodes
}

func (s *LearningPathService) isNodeMastered(node *entities.KnowledgeNode, skills map[string]*entities.SkillLevel) bool {
	// ?
	for _, skill := range node.Skills {
		if skillLevel, exists := skills[skill]; exists {
			if skillLevel.Level >= 7 { // 7
				return true
			}
		}
	}
	return false
}

func (s *LearningPathService) identifyAvailableNodes(ctx context.Context, masteredNodes []uuid.UUID, graph *entities.KnowledgeGraph) []uuid.UUID {
	var availableNodes []uuid.UUID
	masteredSet := make(map[uuid.UUID]bool)
	
	for _, nodeID := range masteredNodes {
		masteredSet[nodeID] = true
	}
	
	for _, node := range graph.Nodes {
		if masteredSet[node.ID] {
			continue // 
		}
		
		// ?
		if s.arePrerequisitesMet(&node, masteredSet, graph) {
			availableNodes = append(availableNodes, node.ID)
		}
	}
	
	return availableNodes
}

func (s *LearningPathService) arePrerequisitesMet(node *entities.KnowledgeNode, masteredSet map[uuid.UUID]bool, graph *entities.KnowledgeGraph) bool {
	// 
	for _, relation := range graph.Relations {
		if relation.ToNodeID == node.ID && relation.Type == entities.RelationTypePrerequisite {
			if !masteredSet[relation.FromNodeID] {
				return false
			}
		}
	}
	return true
}

func (s *LearningPathService) calculateLearningVelocity(history []*entities.LearningHistory) float64 {
	if len(history) == 0 {
		return 1.0 // 
	}
	
	// /?
	var totalTime time.Duration
	var completedContent int
	
	for _, h := range history {
		totalTime += h.Duration
		if h.Progress >= 1.0 {
			completedContent++
		}
	}
	
	if totalTime == 0 {
		return 1.0
	}
	
	// 
	return float64(completedContent) / totalTime.Hours()
}

func (s *LearningPathService) calculatePreferredDifficulty(history []*entities.LearningHistory) entities.DifficultyLevel {
	if len(history) == 0 {
		return entities.DifficultyBeginner
	}
	
	// 
	difficultyPerformance := make(map[entities.DifficultyLevel]float64)
	difficultyCount := make(map[entities.DifficultyLevel]int)
	
	for _, h := range history {
		difficultyPerformance[h.DifficultyLevel] += h.Progress
		difficultyCount[h.DifficultyLevel]++
	}
	
	var bestDifficulty entities.DifficultyLevel
	var bestPerformance float64
	
	for difficulty, totalPerformance := range difficultyPerformance {
		avgPerformance := totalPerformance / float64(difficultyCount[difficulty])
		if avgPerformance > bestPerformance {
			bestPerformance = avgPerformance
			bestDifficulty = difficulty
		}
	}
	
	return bestDifficulty
}

func (s *LearningPathService) analyzeLearningPatterns(history []*entities.LearningHistory, learner *entities.Learner) LearningPatterns {
	// ?
	var totalDuration time.Duration
	var sessionCount int
	
	for _, h := range history {
		totalDuration += h.Duration
		sessionCount++
	}
	
	optimalSessionLength := time.Hour // 1
	if sessionCount > 0 {
		optimalSessionLength = totalDuration / time.Duration(sessionCount)
	}
	
	// 
	contentTypePreference := make(map[entities.ContentType]float64)
	for _, h := range history {
		contentTypePreference[entities.ContentType(h.ContentType)] += h.Progress
	}
	
	// ?
	var maxPreference float64
	for _, preference := range contentTypePreference {
		if preference > maxPreference {
			maxPreference = preference
		}
	}
	
	if maxPreference > 0 {
		for contentType := range contentTypePreference {
			contentTypePreference[contentType] /= maxPreference
		}
	}
	
	return LearningPatterns{
		OptimalSessionLength:  optimalSessionLength,
		PreferredTimeSlots:    learner.Preferences.PreferredTimeSlots,
		ContentTypePreference: contentTypePreference,
		DifficultyProgression: "gradual", // 
		RetentionRate:         0.8,       // 
		EngagementFactors:     []string{"interactive", "visual", "practical"},
	}
}

func (s *LearningPathService) identifyStrengthsAndWeaknesses(skills map[string]*entities.SkillLevel) ([]string, []string) {
	var strengths, weaknesses []string
	
	for skill, level := range skills {
		if level.Level >= 7 {
			strengths = append(strengths, skill)
		} else if level.Level <= 3 {
			weaknesses = append(weaknesses, skill)
		}
	}
	
	return strengths, weaknesses
}

func (s *LearningPathService) generateShortestPaths(ctx context.Context, graph *entities.KnowledgeGraph, learnerState *LearnerState, targetNode *entities.KnowledgeNode, req *PathRecommendationRequest) ([]*entities.LearningPath, error) {
	var paths []*entities.LearningPath
	
	// ?
	for _, startNodeID := range learnerState.AvailableNodes {
		pathNodes, err := s.knowledgeGraphRepo.FindShortestPath(ctx, req.GraphID, startNodeID, targetNode.ID)
		if err != nil {
			continue
		}
		
		if len(pathNodes) > 0 {
			path := s.createLearningPath(pathNodes, "shortest_path")
			paths = append(paths, path)
		}
	}
	
	return paths, nil
}

func (s *LearningPathService) generateSkillBasedPaths(ctx context.Context, graph *entities.KnowledgeGraph, learnerState *LearnerState, targetNode *entities.KnowledgeNode, req *PathRecommendationRequest) ([]*entities.LearningPath, error) {
	// ?
	var paths []*entities.LearningPath
	
	if len(req.TargetSkills) == 0 {
		return paths, nil
	}
	
	// ?
	for _, skill := range req.TargetSkills {
		skillNodes := s.findNodesForSkill(graph, skill)
		for _, node := range skillNodes {
			if node.ID == targetNode.ID {
				continue
			}
			
			pathNodes, err := s.knowledgeGraphRepo.FindShortestPath(ctx, req.GraphID, node.ID, targetNode.ID)
			if err != nil {
				continue
			}
			
			if len(pathNodes) > 0 {
				path := s.createLearningPath(pathNodes, "skill_based")
				paths = append(paths, path)
			}
		}
	}
	
	return paths, nil
}

func (s *LearningPathService) generateDifficultyBasedPaths(ctx context.Context, graph *entities.KnowledgeGraph, learnerState *LearnerState, targetNode *entities.KnowledgeNode, req *PathRecommendationRequest) ([]*entities.LearningPath, error) {
	// 
	var paths []*entities.LearningPath
	
	// ?
	nodesByDifficulty := make(map[entities.DifficultyLevel][]*entities.KnowledgeNode)
	for _, node := range graph.Nodes {
		nodesByDifficulty[node.DifficultyLevel] = append(nodesByDifficulty[node.DifficultyLevel], &node)
	}
	
	// ?
	currentDifficulty := learnerState.PreferredDifficulty
	var pathNodes []*entities.KnowledgeNode
	
	for currentDifficulty <= targetNode.DifficultyLevel {
		if nodes, exists := nodesByDifficulty[currentDifficulty]; exists {
			// ?
			for _, node := range nodes {
				if s.isNodeRelevant(node, targetNode) {
					pathNodes = append(pathNodes, node)
					break
				}
			}
		}
		currentDifficulty++
	}
	
	if len(pathNodes) > 0 {
		path := s.createLearningPath(pathNodes, "difficulty_based")
		paths = append(paths, path)
	}
	
	return paths, nil
}

func (s *LearningPathService) generateInterestBasedPaths(ctx context.Context, graph *entities.KnowledgeGraph, learnerState *LearnerState, targetNode *entities.KnowledgeNode, req *PathRecommendationRequest) ([]*entities.LearningPath, error) {
	// ?
	var paths []*entities.LearningPath
	
	// 
	for _, strengthArea := range learnerState.StrengthAreas {
		strengthNodes := s.findNodesForSkill(graph, strengthArea)
		for _, node := range strengthNodes {
			pathNodes, err := s.knowledgeGraphRepo.FindShortestPath(ctx, req.GraphID, node.ID, targetNode.ID)
			if err != nil {
				continue
			}
			
			if len(pathNodes) > 0 {
				path := s.createLearningPath(pathNodes, "interest_based")
				paths = append(paths, path)
			}
		}
	}
	
	return paths, nil
}

func (s *LearningPathService) createLearningPath(nodes []*entities.KnowledgeNode, pathType string) *entities.LearningPath {
	path := &entities.LearningPath{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("%s_path_%s", pathType, uuid.New().String()[:8]),
		Description: fmt.Sprintf("Generated %s learning path", pathType),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	for i, node := range nodes {
		pathNode := &entities.PathNode{
			ID:          uuid.New(),
			KnowledgeID: node.ID,
			Order:       i,
			IsOptional:  false,
		}
		path.Nodes = append(path.Nodes, *pathNode)
	}
	
	return path
}

func (s *LearningPathService) findNodesForSkill(graph *entities.KnowledgeGraph, skill string) []*entities.KnowledgeNode {
	var nodes []*entities.KnowledgeNode
	
	for _, node := range graph.Nodes {
		for _, nodeSkill := range node.Skills {
			if nodeSkill == skill {
				nodes = append(nodes, &node)
				break
			}
		}
	}
	
	return nodes
}

func (s *LearningPathService) isNodeRelevant(node, targetNode *entities.KnowledgeNode) bool {
	// 
	for _, skill := range node.Skills {
		for _, targetSkill := range targetNode.Skills {
			if skill == targetSkill {
				return true
			}
		}
	}
	return false
}

func (s *LearningPathService) deduplicateAndFilterPaths(paths []*entities.LearningPath, req *PathRecommendationRequest) []*entities.LearningPath {
	// ?
	seen := make(map[string]bool)
	var uniquePaths []*entities.LearningPath
	
	for _, path := range paths {
		signature := s.getPathSignature(path)
		if !seen[signature] {
			seen[signature] = true
			
			// 
			if s.passesFilters(path, req) {
				uniquePaths = append(uniquePaths, path)
			}
		}
	}
	
	return uniquePaths
}

func (s *LearningPathService) getPathSignature(path *entities.LearningPath) string {
	var nodeIDs []string
	for _, node := range path.Nodes {
		nodeIDs = append(nodeIDs, node.KnowledgeID.String())
	}
	return fmt.Sprintf("%v", nodeIDs)
}

func (s *LearningPathService) passesFilters(path *entities.LearningPath, req *PathRecommendationRequest) bool {
	// ?
	if req.Preferences != nil {
		if req.Preferences.MaxPathLength > 0 && len(path.Nodes) > req.Preferences.MaxPathLength {
			return false
		}
		if req.Preferences.MinPathLength > 0 && len(path.Nodes) < req.Preferences.MinPathLength {
			return false
		}
	}
	
	return true
}

func (s *LearningPathService) calculatePersonalizationScore(path *entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) float64 {
	score := 0.0
	
	// ?(: 0.3)
	styleScore := s.calculateStyleMatchScore(path, learner.Preferences.Style)
	score += styleScore * 0.3
	
	// ?(: 0.25)
	difficultyScore := s.calculateDifficultyMatchScore(path, learner)
	score += difficultyScore * 0.25
	
	//  (: 0.25)
	skillScore := s.calculateSkillRelevanceScore(path, learner)
	score += skillScore * 0.25
	
	// ?(: 0.2)
	timeScore := s.calculateTimeConstraintScore(path, req.TimeConstraint)
	score += timeScore * 0.2
	
	return math.Min(score, 1.0) // ?.0
}

func (s *LearningPathService) calculateStyleMatchScore(path *entities.LearningPath, style entities.LearningStyle) float64 {
	// ?
	switch style {
	case entities.LearningStyleVisual:
		return 0.8 // ?
	case entities.LearningStyleAuditory:
		return 0.7
	case entities.LearningStyleKinesthetic:
		return 0.6
	case entities.LearningStyleReading:
		return 0.9
	default:
		return 0.5
	}
}

func (s *LearningPathService) calculateDifficultyMatchScore(path *entities.LearningPath, learner *entities.Learner) float64 {
	// 
	avgDifficulty := s.calculateAveragePathDifficulty(path)
	experienceLevel := float64(learner.ExperienceLevel)
	
	// 
	difficultyDiff := math.Abs(avgDifficulty - experienceLevel)
	
	// ?(?
	return math.Max(0, 1.0 - difficultyDiff/10.0)
}

func (s *LearningPathService) calculateAveragePathDifficulty(path *entities.LearningPath) float64 {
	// ?
	// 
	return 5.0
}

func (s *LearningPathService) calculateSkillRelevanceScore(path *entities.LearningPath, learner *entities.Learner) float64 {
	// ?
	// ?
	return 0.8
}

func (s *LearningPathService) calculateTimeConstraintScore(path *entities.LearningPath, timeConstraint *time.Duration) float64 {
	if timeConstraint == nil {
		return 1.0 // ?
	}
	
	// 
	estimatedTime := time.Duration(len(path.Nodes)) * time.Hour * 2 // ?
	
	if estimatedTime <= *timeConstraint {
		return 1.0
	}
	
	// 
	ratio := float64(*timeConstraint) / float64(estimatedTime)
	return math.Max(0, ratio)
}

func (s *LearningPathService) estimateLearningDuration(path *entities.LearningPath, learner *entities.Learner) time.Duration {
	// 
	baseTime := time.Duration(len(path.Nodes)) * time.Hour * 2
	
	// ?
	experienceFactor := 1.0 - (float64(learner.ExperienceLevel) / 20.0) // ?
	adjustedTime := time.Duration(float64(baseTime) * (0.5 + experienceFactor))
	
	return adjustedTime
}

func (s *LearningPathService) analyzeDifficultyProgression(path *entities.LearningPath) []float64 {
	progression := make([]float64, len(path.Nodes))
	
	for i := range path.Nodes {
		// 
		progression[i] = float64(i+1) / float64(len(path.Nodes)) * 10.0
	}
	
	return progression
}

func (s *LearningPathService) analyzeSkillProgression(path *entities.LearningPath, learner *entities.Learner) map[string][]float64 {
	// 
	skillProgression := make(map[string][]float64)
	
	// ?
	skills := []string{"programming", "algorithms", "data_structures"}
	for _, skill := range skills {
		progression := make([]float64, len(path.Nodes))
		for i := range progression {
			progression[i] = float64(i+1) / float64(len(path.Nodes)) * 10.0
		}
		skillProgression[skill] = progression
	}
	
	return skillProgression
}

func (s *LearningPathService) predictSuccessProbability(path *entities.LearningPath, learner *entities.Learner) float64 {
	// 
	baseProbability := 0.7 // ?
	
	// 
	experienceBonus := float64(learner.ExperienceLevel) / 100.0
	
	// 
	lengthPenalty := math.Max(0, float64(len(path.Nodes)-5)) * 0.02
	
	probability := baseProbability + experienceBonus - lengthPenalty
	return math.Max(0.1, math.Min(0.95, probability))
}

func (s *LearningPathService) calculateEngagementScore(path *entities.LearningPath, learner *entities.Learner) float64 {
	// ?
	baseScore := 0.6
	
	// 
	styleBonus := 0.2 // ?
	
	// ?
	diversityBonus := 0.1 // ?
	
	return math.Min(1.0, baseScore + styleBonus + diversityBonus)
}

func (s *LearningPathService) generateReasoning(path *entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) []string {
	var reasoning []string
	
	reasoning = append(reasoning, fmt.Sprintf("%d", len(path.Nodes)))
	reasoning = append(reasoning, "?)
	reasoning = append(reasoning, "?)
	
	if req.TimeConstraint != nil {
		reasoning = append(reasoning, "䰲")
	}
	
	return reasoning
}

func (s *LearningPathService) identifyAdaptations(path *entities.LearningPath, learner *entities.Learner) []PathAdaptation {
	var adaptations []PathAdaptation
	
	// 
	if learner.Preferences.Style == entities.LearningStyleVisual {
		adaptations = append(adaptations, PathAdaptation{
			Type:        "content_type",
			Description: "?,
			Impact:      0.8,
			Confidence:  0.9,
		})
	}
	
	if learner.ExperienceLevel < 3 {
		adaptations = append(adaptations, PathAdaptation{
			Type:        "difficulty",
			Description: "",
			Impact:      0.7,
			Confidence:  0.85,
		})
	}
	
	return adaptations
}

func (s *LearningPathService) generateMilestones(path *entities.LearningPath, learner *entities.Learner) []PathMilestone {
	var milestones []PathMilestone
	
	// 
	for i, node := range path.Nodes {
		if i%3 == 0 || i == len(path.Nodes)-1 { // ?
			milestone := PathMilestone{
				NodeID:        node.ID,
				Position:      i,
				EstimatedTime: time.Duration(i+1) * time.Hour * 2,
				SkillsAcquired: []string{fmt.Sprintf("skill_%d", i+1)},
				Prerequisites: []uuid.UUID{},
				Assessments:   []uuid.UUID{uuid.New()},
				Rewards:       []string{"badge", "certificate"},
			}
			milestones = append(milestones, milestone)
		}
	}
	
	return milestones
}

