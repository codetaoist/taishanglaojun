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

// KnowledgeGraphService 
type KnowledgeGraphService struct {
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
}

// NewKnowledgeGraphService 
func NewKnowledgeGraphService(
	knowledgeGraphRepo repositories.KnowledgeGraphRepository,
	learnerRepo repositories.LearnerRepository,
	contentRepo repositories.LearningContentRepository,
) *KnowledgeGraphService {
	return &KnowledgeGraphService{
		knowledgeGraphRepo: knowledgeGraphRepo,
		learnerRepo:        learnerRepo,
		contentRepo:        contentRepo,
	}
}

// GraphAnalysisRequest 
type GraphAnalysisRequest struct {
	GraphID     uuid.UUID `json:"graph_id"`
	AnalysisType string   `json:"analysis_type"` // "structure", "learning_gaps", "optimization", "recommendations"
	LearnerID   *uuid.UUID `json:"learner_id,omitempty"`
	Depth       int       `json:"depth"`
	IncludeMetrics bool   `json:"include_metrics"`
}

// GraphAnalysisResult 
type GraphAnalysisResult struct {
	GraphID           uuid.UUID              `json:"graph_id"`
	AnalysisType      string                 `json:"analysis_type"`
	Timestamp         time.Time              `json:"timestamp"`
	StructuralMetrics *StructuralMetrics     `json:"structural_metrics,omitempty"`
	LearningGaps      []*LearningGap         `json:"learning_gaps,omitempty"`
	Recommendations   []*GraphRecommendation `json:"recommendations,omitempty"`
	OptimizationSuggestions []*OptimizationSuggestion `json:"optimization_suggestions,omitempty"`
	QualityScore      float64                `json:"quality_score"`
	Insights          []string               `json:"insights"`
	Warnings          []string               `json:"warnings"`
}

// StructuralMetrics ?
type StructuralMetrics struct {
	NodeCount           int                    `json:"node_count"`
	RelationCount       int                    `json:"relation_count"`
	Density             float64                `json:"density"`
	AveragePathLength   float64                `json:"average_path_length"`
	ClusteringCoefficient float64              `json:"clustering_coefficient"`
	CentralityMeasures  map[uuid.UUID]float64  `json:"centrality_measures"`
	ConnectedComponents int                    `json:"connected_components"`
	CriticalPaths       [][]uuid.UUID          `json:"critical_paths"`
	Bottlenecks         []uuid.UUID            `json:"bottlenecks"`
	DifficultyDistribution map[entities.DifficultyLevel]int `json:"difficulty_distribution"`
	TypeDistribution    map[entities.NodeType]int `json:"type_distribution"`
}

// LearningGap 
type LearningGap struct {
	ID              uuid.UUID              `json:"id"`
	Type            string                 `json:"type"` // "missing_prerequisite", "skill_gap", "knowledge_gap", "difficulty_jump"
	Description     string                 `json:"description"`
	Severity        float64                `json:"severity"` // 0-1
	AffectedNodes   []uuid.UUID            `json:"affected_nodes"`
	SuggestedNodes  []uuid.UUID            `json:"suggested_nodes"`
	Impact          string                 `json:"impact"`
	Priority        int                    `json:"priority"`
	EstimatedEffort time.Duration          `json:"estimated_effort"`
}

// GraphRecommendation 
type GraphRecommendation struct {
	ID          uuid.UUID   `json:"id"`
	Type        string      `json:"type"` // "add_node", "add_relation", "modify_difficulty", "restructure"
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Rationale   string      `json:"rationale"`
	Priority    int         `json:"priority"`
	Confidence  float64     `json:"confidence"`
	Impact      float64     `json:"impact"`
	Effort      float64     `json:"effort"`
	TargetNodes []uuid.UUID `json:"target_nodes"`
	Actions     []RecommendationAction `json:"actions"`
}

// RecommendationAction 
type RecommendationAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OptimizationSuggestion 
type OptimizationSuggestion struct {
	ID          uuid.UUID `json:"id"`
	Category    string    `json:"category"` // "structure", "content", "difficulty", "accessibility"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Benefits    []string  `json:"benefits"`
	Risks       []string  `json:"risks"`
	Complexity  string    `json:"complexity"` // "low", "medium", "high"
	ROI         float64   `json:"roi"` // Return on Investment
}

// ConceptRecommendationRequest 
type ConceptRecommendationRequest struct {
	GraphID       uuid.UUID   `json:"graph_id"`
	LearnerID     uuid.UUID   `json:"learner_id"`
	CurrentNodeID *uuid.UUID  `json:"current_node_id,omitempty"`
	TargetSkills  []string    `json:"target_skills,omitempty"`
	MaxRecommendations int    `json:"max_recommendations"`
	IncludeReasoning bool     `json:"include_reasoning"`
}

// ConceptRecommendation 
type ConceptRecommendation struct {
	NodeID          uuid.UUID `json:"node_id"`
	RecommendationType string `json:"recommendation_type"` // "next", "prerequisite", "related", "advanced"
	Score           float64   `json:"score"`
	Confidence      float64   `json:"confidence"`
	Reasoning       []string  `json:"reasoning"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	DifficultyMatch float64   `json:"difficulty_match"`
	SkillRelevance  float64   `json:"skill_relevance"`
	Prerequisites   []uuid.UUID `json:"prerequisites"`
	LearningPath    []uuid.UUID `json:"learning_path"`
}

// AnalyzeGraph 
func (s *KnowledgeGraphService) AnalyzeGraph(ctx context.Context, req *GraphAnalysisRequest) (*GraphAnalysisResult, error) {
	// 
	graph, err := s.knowledgeGraphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	result := &GraphAnalysisResult{
		GraphID:      req.GraphID,
		AnalysisType: req.AnalysisType,
		Timestamp:    time.Now(),
		Insights:     []string{},
		Warnings:     []string{},
	}

	// ?
	switch req.AnalysisType {
	case "structure":
		result.StructuralMetrics = s.analyzeStructure(ctx, graph)
		result.QualityScore = s.calculateStructuralQuality(result.StructuralMetrics)
		
	case "learning_gaps":
		if req.LearnerID != nil {
			learner, err := s.learnerRepo.GetByID(ctx, *req.LearnerID)
			if err == nil {
				result.LearningGaps = s.identifyLearningGaps(ctx, graph, learner)
			}
		} else {
			result.LearningGaps = s.identifyGeneralGaps(ctx, graph)
		}
		result.QualityScore = s.calculateGapScore(result.LearningGaps)
		
	case "optimization":
		result.OptimizationSuggestions = s.generateOptimizationSuggestions(ctx, graph)
		result.QualityScore = s.calculateOptimizationPotential(result.OptimizationSuggestions)
		
	case "recommendations":
		result.Recommendations = s.generateGraphRecommendations(ctx, graph)
		result.QualityScore = s.calculateRecommendationValue(result.Recommendations)
		
	default:
		// 
		result.StructuralMetrics = s.analyzeStructure(ctx, graph)
		result.LearningGaps = s.identifyGeneralGaps(ctx, graph)
		result.OptimizationSuggestions = s.generateOptimizationSuggestions(ctx, graph)
		result.Recommendations = s.generateGraphRecommendations(ctx, graph)
		result.QualityScore = s.calculateOverallQuality(result)
	}

	// ?
	result.Insights = s.generateInsights(result)
	result.Warnings = s.generateWarnings(result)

	return result, nil
}

// RecommendConcepts 
func (s *KnowledgeGraphService) RecommendConcepts(ctx context.Context, req *ConceptRecommendationRequest) ([]*ConceptRecommendation, error) {
	// 
	graph, err := s.knowledgeGraphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	// ?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// ?
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner skills: %w", err)
	}

	// ?
	learnerState := s.analyzeLearnerKnowledgeState(skills, graph)

	// 
	recommendations := s.generateConceptRecommendations(ctx, graph, learner, learnerState, req)

	// ?
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.MaxRecommendations {
		recommendations = recommendations[:req.MaxRecommendations]
	}

	return recommendations, nil
}

// analyzeStructure 
func (s *KnowledgeGraphService) analyzeStructure(ctx context.Context, graph *entities.KnowledgeGraph) *StructuralMetrics {
	metrics := &StructuralMetrics{
		NodeCount:     len(graph.Nodes),
		RelationCount: len(graph.Relations),
		CentralityMeasures: make(map[uuid.UUID]float64),
		DifficultyDistribution: make(map[entities.DifficultyLevel]int),
		TypeDistribution: make(map[entities.NodeType]int),
	}

	// 
	if metrics.NodeCount > 1 {
		maxPossibleEdges := metrics.NodeCount * (metrics.NodeCount - 1)
		metrics.Density = float64(metrics.RelationCount) / float64(maxPossibleEdges)
	}

	// 
	metrics.AveragePathLength = s.calculateAveragePathLength(graph)

	// 
	metrics.ClusteringCoefficient = s.calculateClusteringCoefficient(graph)

	// ?
	metrics.CentralityMeasures = s.calculateCentralityMeasures(graph)

	// ?
	metrics.ConnectedComponents = s.countConnectedComponents(graph)

	// 
	metrics.CriticalPaths = s.identifyCriticalPaths(graph)

	// 
	metrics.Bottlenecks = s.identifyBottlenecks(graph)

	// 
	for _, node := range graph.Nodes {
		metrics.DifficultyDistribution[node.DifficultyLevel]++
		metrics.TypeDistribution[node.Type]++
	}

	return metrics
}

// identifyLearningGaps 
func (s *KnowledgeGraphService) identifyLearningGaps(ctx context.Context, graph *entities.KnowledgeGraph, learner *entities.Learner) []*LearningGap {
	var gaps []*LearningGap

	// ?
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learner.ID)
	if err != nil {
		return gaps
	}

	// ?
	missingPrereqs := s.identifyMissingPrerequisites(graph, skills)
	for _, gap := range missingPrereqs {
		gaps = append(gaps, gap)
	}

	// ?
	skillGaps := s.identifySkillGaps(graph, skills, learner)
	for _, gap := range skillGaps {
		gaps = append(gaps, gap)
	}

	// 
	difficultyJumps := s.identifyDifficultyJumps(graph, learner)
	for _, gap := range difficultyJumps {
		gaps = append(gaps, gap)
	}

	// ?
	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].Severity > gaps[j].Severity
	})

	return gaps
}

// generateOptimizationSuggestions 
func (s *KnowledgeGraphService) generateOptimizationSuggestions(ctx context.Context, graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// 
	structuralSuggestions := s.generateStructuralOptimizations(graph)
	suggestions = append(suggestions, structuralSuggestions...)

	// 
	contentSuggestions := s.generateContentOptimizations(graph)
	suggestions = append(suggestions, contentSuggestions...)

	// 
	difficultySuggestions := s.generateDifficultyOptimizations(graph)
	suggestions = append(suggestions, difficultySuggestions...)

	// ?
	accessibilitySuggestions := s.generateAccessibilityOptimizations(graph)
	suggestions = append(suggestions, accessibilitySuggestions...)

	// ROI
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].ROI > suggestions[j].ROI
	})

	return suggestions
}

// generateGraphRecommendations 
func (s *KnowledgeGraphService) generateGraphRecommendations(ctx context.Context, graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	// 
	nodeRecommendations := s.generateNodeAdditionRecommendations(graph)
	recommendations = append(recommendations, nodeRecommendations...)

	// 
	relationRecommendations := s.generateRelationRecommendations(graph)
	recommendations = append(recommendations, relationRecommendations...)

	// 
	difficultyRecommendations := s.generateDifficultyModificationRecommendations(graph)
	recommendations = append(recommendations, difficultyRecommendations...)

	// 
	restructureRecommendations := s.generateRestructureRecommendations(graph)
	recommendations = append(recommendations, restructureRecommendations...)

	// ?
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Priority != recommendations[j].Priority {
			return recommendations[i].Priority > recommendations[j].Priority
		}
		return recommendations[i].Impact > recommendations[j].Impact
	})

	return recommendations
}

// 

func (s *KnowledgeGraphService) calculateAveragePathLength(graph *entities.KnowledgeGraph) float64 {
	if len(graph.Nodes) < 2 {
		return 0
	}

	totalLength := 0.0
	pathCount := 0

	// ?
	for i, nodeA := range graph.Nodes {
		for j, nodeB := range graph.Nodes {
			if i != j {
				// ?
				// ?
				if s.areDirectlyConnected(nodeA.ID, nodeB.ID, graph) {
					totalLength += 1.0
				} else {
					totalLength += 2.0 // ??
				}
				pathCount++
			}
		}
	}

	if pathCount == 0 {
		return 0
	}

	return totalLength / float64(pathCount)
}

func (s *KnowledgeGraphService) areDirectlyConnected(nodeA, nodeB uuid.UUID, graph *entities.KnowledgeGraph) bool {
	for _, relation := range graph.Relations {
		if (relation.FromNodeID == nodeA && relation.ToNodeID == nodeB) ||
		   (relation.FromNodeID == nodeB && relation.ToNodeID == nodeA) {
			return true
		}
	}
	return false
}

func (s *KnowledgeGraphService) calculateClusteringCoefficient(graph *entities.KnowledgeGraph) float64 {
	if len(graph.Nodes) < 3 {
		return 0
	}

	totalCoefficient := 0.0
	nodeCount := 0

	for _, node := range graph.Nodes {
		neighbors := s.getNeighbors(node.ID, graph)
		if len(neighbors) < 2 {
			continue
		}

		// 
		connections := 0
		for i, neighborA := range neighbors {
			for j, neighborB := range neighbors {
				if i != j && s.areDirectlyConnected(neighborA, neighborB, graph) {
					connections++
				}
			}
		}

		// 
		maxConnections := len(neighbors) * (len(neighbors) - 1)
		if maxConnections > 0 {
			coefficient := float64(connections) / float64(maxConnections)
			totalCoefficient += coefficient
			nodeCount++
		}
	}

	if nodeCount == 0 {
		return 0
	}

	return totalCoefficient / float64(nodeCount)
}

func (s *KnowledgeGraphService) getNeighbors(nodeID uuid.UUID, graph *entities.KnowledgeGraph) []uuid.UUID {
	var neighbors []uuid.UUID
	neighborSet := make(map[uuid.UUID]bool)

	for _, relation := range graph.Relations {
		if relation.FromNodeID == nodeID && !neighborSet[relation.ToNodeID] {
			neighbors = append(neighbors, relation.ToNodeID)
			neighborSet[relation.ToNodeID] = true
		} else if relation.ToNodeID == nodeID && !neighborSet[relation.FromNodeID] {
			neighbors = append(neighbors, relation.FromNodeID)
			neighborSet[relation.FromNodeID] = true
		}
	}

	return neighbors
}

func (s *KnowledgeGraphService) calculateCentralityMeasures(graph *entities.KnowledgeGraph) map[uuid.UUID]float64 {
	centrality := make(map[uuid.UUID]float64)

	// ?
	for _, node := range graph.Nodes {
		degree := 0
		for _, relation := range graph.Relations {
			if relation.FromNodeID == node.ID || relation.ToNodeID == node.ID {
				degree++
			}
		}
		centrality[node.ID] = float64(degree)
	}

	// ?
	maxDegree := 0.0
	for _, degree := range centrality {
		if degree > maxDegree {
			maxDegree = degree
		}
	}

	if maxDegree > 0 {
		for nodeID := range centrality {
			centrality[nodeID] /= maxDegree
		}
	}

	return centrality
}

func (s *KnowledgeGraphService) countConnectedComponents(graph *entities.KnowledgeGraph) int {
	visited := make(map[uuid.UUID]bool)
	components := 0

	for _, node := range graph.Nodes {
		if !visited[node.ID] {
			s.dfsVisit(node.ID, graph, visited)
			components++
		}
	}

	return components
}

func (s *KnowledgeGraphService) dfsVisit(nodeID uuid.UUID, graph *entities.KnowledgeGraph, visited map[uuid.UUID]bool) {
	visited[nodeID] = true

	neighbors := s.getNeighbors(nodeID, graph)
	for _, neighbor := range neighbors {
		if !visited[neighbor] {
			s.dfsVisit(neighbor, graph, visited)
		}
	}
}

func (s *KnowledgeGraphService) identifyCriticalPaths(graph *entities.KnowledgeGraph) [][]uuid.UUID {
	var criticalPaths [][]uuid.UUID

	// ??
	inDegree := make(map[uuid.UUID]int)
	for _, node := range graph.Nodes {
		inDegree[node.ID] = 0
	}

	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypePrerequisite {
			inDegree[relation.ToNodeID]++
		}
	}

	// ?
	for _, node := range graph.Nodes {
		if inDegree[node.ID] == 0 {
			path := s.findLongestPath(node.ID, graph, make(map[uuid.UUID]bool))
			if len(path) > 3 { // ?
				criticalPaths = append(criticalPaths, path)
			}
		}
	}

	return criticalPaths
}

func (s *KnowledgeGraphService) findLongestPath(nodeID uuid.UUID, graph *entities.KnowledgeGraph, visited map[uuid.UUID]bool) []uuid.UUID {
	visited[nodeID] = true
	defer func() { visited[nodeID] = false }()

	longestPath := []uuid.UUID{nodeID}

	// ?
	for _, relation := range graph.Relations {
		if relation.FromNodeID == nodeID && relation.Type == entities.RelationTypePrerequisite {
			if !visited[relation.ToNodeID] {
				subPath := s.findLongestPath(relation.ToNodeID, graph, visited)
				if len(subPath) > len(longestPath)-1 {
					longestPath = append([]uuid.UUID{nodeID}, subPath...)
				}
			}
		}
	}

	return longestPath
}

func (s *KnowledgeGraphService) identifyBottlenecks(graph *entities.KnowledgeGraph) []uuid.UUID {
	var bottlenecks []uuid.UUID

	// ?
	betweennessCentrality := s.calculateBetweennessCentrality(graph)

	// 
	threshold := 0.7 // ?
	for nodeID, centrality := range betweennessCentrality {
		if centrality > threshold {
			bottlenecks = append(bottlenecks, nodeID)
		}
	}

	return bottlenecks
}

func (s *KnowledgeGraphService) calculateBetweennessCentrality(graph *entities.KnowledgeGraph) map[uuid.UUID]float64 {
	centrality := make(map[uuid.UUID]float64)

	// 
	for _, node := range graph.Nodes {
		degree := len(s.getNeighbors(node.ID, graph))
		centrality[node.ID] = float64(degree) / float64(len(graph.Nodes)-1)
	}

	return centrality
}

func (s *KnowledgeGraphService) identifyGeneralGaps(ctx context.Context, graph *entities.KnowledgeGraph) []*LearningGap {
	var gaps []*LearningGap

	// 
	isolatedNodes := s.findIsolatedNodes(graph)
	for _, nodeID := range isolatedNodes {
		gap := &LearningGap{
			ID:          uuid.New(),
			Type:        "isolated_node",
			Description: "",
			Severity:    0.8,
			AffectedNodes: []uuid.UUID{nodeID},
			Impact:      "",
			Priority:    3,
			EstimatedEffort: time.Hour * 2,
		}
		gaps = append(gaps, gap)
	}

	// 
	difficultyJumps := s.findDifficultyJumps(graph)
	for _, jump := range difficultyJumps {
		gap := &LearningGap{
			ID:          uuid.New(),
			Type:        "difficulty_jump",
			Description: "?,
			Severity:    0.6,
			AffectedNodes: jump,
			Impact:      "?,
			Priority:    2,
			EstimatedEffort: time.Hour * 4,
		}
		gaps = append(gaps, gap)
	}

	return gaps
}

func (s *KnowledgeGraphService) findIsolatedNodes(graph *entities.KnowledgeGraph) []uuid.UUID {
	var isolated []uuid.UUID

	for _, node := range graph.Nodes {
		neighbors := s.getNeighbors(node.ID, graph)
		if len(neighbors) == 0 {
			isolated = append(isolated, node.ID)
		}
	}

	return isolated
}

func (s *KnowledgeGraphService) findDifficultyJumps(graph *entities.KnowledgeGraph) [][]uuid.UUID {
	var jumps [][]uuid.UUID

	// 
	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypePrerequisite {
			fromNode := s.findNodeByID(relation.FromNodeID, graph)
			toNode := s.findNodeByID(relation.ToNodeID, graph)

			if fromNode != nil && toNode != nil {
				difficultyDiff := int(toNode.DifficultyLevel) - int(fromNode.DifficultyLevel)
				if difficultyDiff > 2 { // 2?
					jumps = append(jumps, []uuid.UUID{fromNode.ID, toNode.ID})
				}
			}
		}
	}

	return jumps
}

func (s *KnowledgeGraphService) findNodeByID(nodeID uuid.UUID, graph *entities.KnowledgeGraph) *entities.KnowledgeNode {
	for _, node := range graph.Nodes {
		if node.ID == nodeID {
			return &node
		}
	}
	return nil
}

func (s *KnowledgeGraphService) identifyMissingPrerequisites(graph *entities.KnowledgeGraph, skills map[string]*entities.SkillLevel) []*LearningGap {
	var gaps []*LearningGap

	for _, node := range graph.Nodes {
		// 
		prerequisites := s.getPrerequisites(node.ID, graph)
		var missingPrereqs []uuid.UUID

		for _, prereqID := range prerequisites {
			prereqNode := s.findNodeByID(prereqID, graph)
			if prereqNode != nil && !s.hasRequiredSkills(prereqNode, skills) {
				missingPrereqs = append(missingPrereqs, prereqID)
			}
		}

		if len(missingPrereqs) > 0 {
			gap := &LearningGap{
				ID:          uuid.New(),
				Type:        "missing_prerequisite",
				Description: fmt.Sprintf(" %s ?, node.Name),
				Severity:    0.9,
				AffectedNodes: []uuid.UUID{node.ID},
				SuggestedNodes: missingPrereqs,
				Impact:      "",
				Priority:    1,
				EstimatedEffort: time.Duration(len(missingPrereqs)) * time.Hour * 3,
			}
			gaps = append(gaps, gap)
		}
	}

	return gaps
}

func (s *KnowledgeGraphService) getPrerequisites(nodeID uuid.UUID, graph *entities.KnowledgeGraph) []uuid.UUID {
	var prerequisites []uuid.UUID

	for _, relation := range graph.Relations {
		if relation.ToNodeID == nodeID && relation.Type == entities.RelationTypePrerequisite {
			prerequisites = append(prerequisites, relation.FromNodeID)
		}
	}

	return prerequisites
}

func (s *KnowledgeGraphService) hasRequiredSkills(node *entities.KnowledgeNode, skills map[string]*entities.SkillLevel) bool {
	for _, skill := range node.Skills {
		if skillLevel, exists := skills[skill]; !exists || skillLevel.Level < 5 {
			return false
		}
	}
	return true
}

func (s *KnowledgeGraphService) identifySkillGaps(graph *entities.KnowledgeGraph, skills map[string]*entities.SkillLevel, learner *entities.Learner) []*LearningGap {
	var gaps []*LearningGap

	// 
	for _, goal := range learner.LearningGoals {
		if !goal.IsActive {
			continue
		}

		// 
		relatedNodes := s.findNodesForSkill(graph, goal.TargetSkill)
		for _, node := range relatedNodes {
			if !s.hasRequiredSkills(node, skills) {
				gap := &LearningGap{
					ID:          uuid.New(),
					Type:        "skill_gap",
					Description: fmt.Sprintf("? %s", goal.TargetSkill),
					Severity:    0.7,
					AffectedNodes: []uuid.UUID{node.ID},
					Impact:      "",
					Priority:    2,
					EstimatedEffort: time.Until(goal.TargetDate),
				}
				gaps = append(gaps, gap)
			}
		}
	}

	return gaps
}

func (s *KnowledgeGraphService) findNodesForSkill(graph *entities.KnowledgeGraph, skill string) []*entities.KnowledgeNode {
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

func (s *KnowledgeGraphService) identifyDifficultyJumps(graph *entities.KnowledgeGraph, learner *entities.Learner) []*LearningGap {
	var gaps []*LearningGap

	// ?
	experienceLevel := entities.DifficultyLevel(learner.ExperienceLevel)

	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypePrerequisite {
			fromNode := s.findNodeByID(relation.FromNodeID, graph)
			toNode := s.findNodeByID(relation.ToNodeID, graph)

			if fromNode != nil && toNode != nil {
				difficultyDiff := int(toNode.DifficultyLevel) - int(fromNode.DifficultyLevel)
				if difficultyDiff > 2 && toNode.DifficultyLevel > experienceLevel+1 {
					gap := &LearningGap{
						ID:          uuid.New(),
						Type:        "difficulty_jump",
						Description: "?,
						Severity:    0.6,
						AffectedNodes: []uuid.UUID{fromNode.ID, toNode.ID},
						Impact:      "",
						Priority:    2,
						EstimatedEffort: time.Hour * 6,
					}
					gaps = append(gaps, gap)
				}
			}
		}
	}

	return gaps
}

// 

func (s *KnowledgeGraphService) calculateStructuralQuality(metrics *StructuralMetrics) float64 {
	score := 0.0

	//  (0.3)
	densityScore := math.Min(metrics.Density * 2, 1.0) // ?.5
	score += densityScore * 0.3

	// ?(0.3)
	connectivityScore := 1.0 / float64(metrics.ConnectedComponents)
	score += connectivityScore * 0.3

	// ?(0.2)
	balanceScore := 1.0 - math.Abs(metrics.ClusteringCoefficient - 0.5) * 2
	score += math.Max(0, balanceScore) * 0.2

	// ?(0.2)
	complexityScore := 1.0 - math.Min(metrics.AveragePathLength / 10.0, 1.0)
	score += complexityScore * 0.2

	return math.Min(score, 1.0)
}

func (s *KnowledgeGraphService) calculateGapScore(gaps []*LearningGap) float64 {
	if len(gaps) == 0 {
		return 1.0
	}

	totalSeverity := 0.0
	for _, gap := range gaps {
		totalSeverity += gap.Severity
	}

	avgSeverity := totalSeverity / float64(len(gaps))
	return math.Max(0, 1.0 - avgSeverity)
}

func (s *KnowledgeGraphService) calculateOptimizationPotential(suggestions []*OptimizationSuggestion) float64 {
	if len(suggestions) == 0 {
		return 0.5
	}

	totalROI := 0.0
	for _, suggestion := range suggestions {
		totalROI += suggestion.ROI
	}

	avgROI := totalROI / float64(len(suggestions))
	return math.Min(avgROI / 10.0, 1.0) // ROI?0
}

func (s *KnowledgeGraphService) calculateRecommendationValue(recommendations []*GraphRecommendation) float64 {
	if len(recommendations) == 0 {
		return 0.5
	}

	totalValue := 0.0
	for _, rec := range recommendations {
		value := rec.Impact * rec.Confidence / rec.Effort
		totalValue += value
	}

	avgValue := totalValue / float64(len(recommendations))
	return math.Min(avgValue, 1.0)
}

func (s *KnowledgeGraphService) calculateOverallQuality(result *GraphAnalysisResult) float64 {
	score := 0.0
	components := 0

	if result.StructuralMetrics != nil {
		score += s.calculateStructuralQuality(result.StructuralMetrics) * 0.4
		components++
	}

	if result.LearningGaps != nil {
		score += s.calculateGapScore(result.LearningGaps) * 0.3
		components++
	}

	if result.OptimizationSuggestions != nil {
		score += s.calculateOptimizationPotential(result.OptimizationSuggestions) * 0.2
		components++
	}

	if result.Recommendations != nil {
		score += s.calculateRecommendationValue(result.Recommendations) * 0.1
		components++
	}

	if components == 0 {
		return 0.5
	}

	return score
}

// ?

func (s *KnowledgeGraphService) generateInsights(result *GraphAnalysisResult) []string {
	var insights []string

	if result.StructuralMetrics != nil {
		if result.StructuralMetrics.Density > 0.7 {
			insights = append(insights, "")
		}
		if result.StructuralMetrics.ConnectedComponents == 1 {
			insights = append(insights, "?)
		}
		if len(result.StructuralMetrics.Bottlenecks) > 0 {
			insights = append(insights, fmt.Sprintf("%d", len(result.StructuralMetrics.Bottlenecks)))
		}
	}

	if result.QualityScore > 0.8 {
		insights = append(insights, "")
	} else if result.QualityScore > 0.6 {
		insights = append(insights, "?)
	}

	return insights
}

func (s *KnowledgeGraphService) generateWarnings(result *GraphAnalysisResult) []string {
	var warnings []string

	if result.StructuralMetrics != nil {
		if result.StructuralMetrics.ConnectedComponents > 1 {
			warnings = append(warnings, "?)
		}
		if result.StructuralMetrics.Density < 0.2 {
			warnings = append(warnings, "㹻?)
		}
		if result.StructuralMetrics.AveragePathLength > 8 {
			warnings = append(warnings, "?)
		}
	}

	if len(result.LearningGaps) > 5 {
		warnings = append(warnings, "?)
	}

	if result.QualityScore < 0.4 {
		warnings = append(warnings, "?)
	}

	return warnings
}

// 

func (s *KnowledgeGraphService) generateStructuralOptimizations(graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// ?
	components := s.countConnectedComponents(graph)
	if components > 1 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "structure",
			Title:       "?,
			Description: "?,
			Benefits:    []string{"?, ""},
			Risks:       []string{"?},
			Complexity:  "medium",
			ROI:         8.0,
		})
	}

	// ?
	density := s.calculateDensity(graph)
	if density < 0.2 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "structure",
			Title:       "",
			Description: "?,
			Benefits:    []string{"", ""},
			Risks:       []string{""},
			Complexity:  "low",
			ROI:         6.0,
		})
	}

	return suggestions
}

func (s *KnowledgeGraphService) calculateDensity(graph *entities.KnowledgeGraph) float64 {
	if len(graph.Nodes) < 2 {
		return 0
	}
	maxEdges := len(graph.Nodes) * (len(graph.Nodes) - 1)
	return float64(len(graph.Relations)) / float64(maxEdges)
}

func (s *KnowledgeGraphService) generateContentOptimizations(graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// 
	suggestions = append(suggestions, &OptimizationSuggestion{
		ID:          uuid.New(),
		Category:    "content",
		Title:       "",
		Description: "?,
		Benefits:    []string{"", ""},
		Risks:       []string{""},
		Complexity:  "medium",
		ROI:         7.0,
	})

	return suggestions
}

func (s *KnowledgeGraphService) generateDifficultyOptimizations(graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// ?
	difficultyJumps := s.findDifficultyJumps(graph)
	if len(difficultyJumps) > 0 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "difficulty",
			Title:       "",
			Description: "?,
			Benefits:    []string{"", "?},
			Risks:       []string{""},
			Complexity:  "medium",
			ROI:         9.0,
		})
	}

	return suggestions
}

func (s *KnowledgeGraphService) generateAccessibilityOptimizations(graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	suggestions = append(suggestions, &OptimizationSuggestion{
		ID:          uuid.New(),
		Category:    "accessibility",
		Title:       "?,
		Description: "",
		Benefits:    []string{"", "?},
		Risks:       []string{"?},
		Complexity:  "high",
		ROI:         5.0,
	})

	return suggestions
}

// 

func (s *KnowledgeGraphService) generateNodeAdditionRecommendations(graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	// ?
	recommendations = append(recommendations, &GraphRecommendation{
		ID:          uuid.New(),
		Type:        "add_node",
		Title:       "",
		Description: "",
		Rationale:   "",
		Priority:    3,
		Confidence:  0.8,
		Impact:      0.7,
		Effort:      0.6,
		Actions: []RecommendationAction{
			{
				Type:        "create_node",
				Description: "",
				Parameters: map[string]interface{}{
					"type": "concept",
					"difficulty": "intermediate",
				},
			},
		},
	})

	return recommendations
}

func (s *KnowledgeGraphService) generateRelationRecommendations(graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	recommendations = append(recommendations, &GraphRecommendation{
		ID:          uuid.New(),
		Type:        "add_relation",
		Title:       "",
		Description: "?,
		Rationale:   "?,
		Priority:    2,
		Confidence:  0.7,
		Impact:      0.6,
		Effort:      0.3,
		Actions: []RecommendationAction{
			{
				Type:        "create_relation",
				Description: "",
				Parameters: map[string]interface{}{
					"type": "related_to",
					"strength": 0.8,
				},
			},
		},
	})

	return recommendations
}

func (s *KnowledgeGraphService) generateDifficultyModificationRecommendations(graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	recommendations = append(recommendations, &GraphRecommendation{
		ID:          uuid.New(),
		Type:        "modify_difficulty",
		Title:       "",
		Description: "",
		Rationale:   "",
		Priority:    2,
		Confidence:  0.6,
		Impact:      0.5,
		Effort:      0.4,
		Actions: []RecommendationAction{
			{
				Type:        "update_difficulty",
				Description: "",
				Parameters: map[string]interface{}{
					"method": "expert_review",
				},
			},
		},
	})

	return recommendations
}

func (s *KnowledgeGraphService) generateRestructureRecommendations(graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	recommendations = append(recommendations, &GraphRecommendation{
		ID:          uuid.New(),
		Type:        "restructure",
		Title:       "",
		Description: "?,
		Rationale:   "?,
		Priority:    1,
		Confidence:  0.5,
		Impact:      0.9,
		Effort:      0.8,
		Actions: []RecommendationAction{
			{
				Type:        "reorganize_hierarchy",
				Description: "",
				Parameters: map[string]interface{}{
					"strategy": "bottom_up",
				},
			},
		},
	})

	return recommendations
}

// 

func (s *KnowledgeGraphService) analyzeLearnerKnowledgeState(skills map[string]*entities.SkillLevel, graph *entities.KnowledgeGraph) map[uuid.UUID]float64 {
	nodeReadiness := make(map[uuid.UUID]float64)

	for _, node := range graph.Nodes {
		readiness := s.calculateNodeReadiness(&node, skills, graph)
		nodeReadiness[node.ID] = readiness
	}

	return nodeReadiness
}

func (s *KnowledgeGraphService) calculateNodeReadiness(node *entities.KnowledgeNode, skills map[string]*entities.SkillLevel, graph *entities.KnowledgeGraph) float64 {
	// 
	prerequisites := s.getPrerequisites(node.ID, graph)
	if len(prerequisites) == 0 {
		return 1.0 // 
	}

	satisfiedPrereqs := 0
	for _, prereqID := range prerequisites {
		prereqNode := s.findNodeByID(prereqID, graph)
		if prereqNode != nil && s.hasRequiredSkills(prereqNode, skills) {
			satisfiedPrereqs++
		}
	}

	return float64(satisfiedPrereqs) / float64(len(prerequisites))
}

func (s *KnowledgeGraphService) generateConceptRecommendations(ctx context.Context, graph *entities.KnowledgeGraph, learner *entities.Learner, learnerState map[uuid.UUID]float64, req *ConceptRecommendationRequest) []*ConceptRecommendation {
	var recommendations []*ConceptRecommendation

	// ?
	nextRecommendations := s.generateNextConceptRecommendations(graph, learner, learnerState, req)
	recommendations = append(recommendations, nextRecommendations...)

	prerequisiteRecommendations := s.generatePrerequisiteRecommendations(graph, learner, learnerState, req)
	recommendations = append(recommendations, prerequisiteRecommendations...)

	relatedRecommendations := s.generateRelatedConceptRecommendations(graph, learner, learnerState, req)
	recommendations = append(recommendations, relatedRecommendations...)

	advancedRecommendations := s.generateAdvancedConceptRecommendations(graph, learner, learnerState, req)
	recommendations = append(recommendations, advancedRecommendations...)

	return recommendations
}

func (s *KnowledgeGraphService) generateNextConceptRecommendations(graph *entities.KnowledgeGraph, learner *entities.Learner, learnerState map[uuid.UUID]float64, req *ConceptRecommendationRequest) []*ConceptRecommendation {
	var recommendations []*ConceptRecommendation

	// 
	for nodeID, readiness := range learnerState {
		if readiness > 0.8 { // 
			node := s.findNodeByID(nodeID, graph)
			if node != nil {
				recommendation := &ConceptRecommendation{
					NodeID:             nodeID,
					RecommendationType: "next",
					Score:              readiness * 0.9,
					Confidence:         0.8,
					Reasoning:          []string{"?, ""},
					EstimatedTime:      time.Hour * 3,
					DifficultyMatch:    s.calculateDifficultyMatch(node, learner),
					SkillRelevance:     s.calculateSkillRelevance(node, learner),
					Prerequisites:      s.getPrerequisites(nodeID, graph),
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations
}

func (s *KnowledgeGraphService) generatePrerequisiteRecommendations(graph *entities.KnowledgeGraph, learner *entities.Learner, learnerState map[uuid.UUID]float64, req *ConceptRecommendationRequest) []*ConceptRecommendation {
	var recommendations []*ConceptRecommendation

	// ?
	for nodeID, readiness := range learnerState {
		if readiness < 0.5 && readiness > 0 { // ?
			prerequisites := s.getPrerequisites(nodeID, graph)
			for _, prereqID := range prerequisites {
				if learnerState[prereqID] < 0.8 {
					node := s.findNodeByID(prereqID, graph)
					if node != nil {
						recommendation := &ConceptRecommendation{
							NodeID:             prereqID,
							RecommendationType: "prerequisite",
							Score:              (1.0 - learnerState[prereqID]) * 0.8,
							Confidence:         0.9,
							Reasoning:          []string{"", "?},
							EstimatedTime:      time.Hour * 2,
							DifficultyMatch:    s.calculateDifficultyMatch(node, learner),
							SkillRelevance:     s.calculateSkillRelevance(node, learner),
						}
						recommendations = append(recommendations, recommendation)
					}
				}
			}
		}
	}

	return recommendations
}

func (s *KnowledgeGraphService) generateRelatedConceptRecommendations(graph *entities.KnowledgeGraph, learner *entities.Learner, learnerState map[uuid.UUID]float64, req *ConceptRecommendationRequest) []*ConceptRecommendation {
	var recommendations []*ConceptRecommendation

	if req.CurrentNodeID == nil {
		return recommendations
	}

	// 
	relatedNodes := s.getRelatedNodes(*req.CurrentNodeID, graph)
	for _, nodeID := range relatedNodes {
		if learnerState[nodeID] > 0.6 {
			node := s.findNodeByID(nodeID, graph)
			if node != nil {
				recommendation := &ConceptRecommendation{
					NodeID:             nodeID,
					RecommendationType: "related",
					Score:              learnerState[nodeID] * 0.7,
					Confidence:         0.6,
					Reasoning:          []string{"?, ""},
					EstimatedTime:      time.Hour * 2,
					DifficultyMatch:    s.calculateDifficultyMatch(node, learner),
					SkillRelevance:     s.calculateSkillRelevance(node, learner),
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations
}

func (s *KnowledgeGraphService) generateAdvancedConceptRecommendations(graph *entities.KnowledgeGraph, learner *entities.Learner, learnerState map[uuid.UUID]float64, req *ConceptRecommendationRequest) []*ConceptRecommendation {
	var recommendations []*ConceptRecommendation

	// 
	for nodeID, readiness := range learnerState {
		node := s.findNodeByID(nodeID, graph)
		if node != nil && node.DifficultyLevel > entities.DifficultyLevel(learner.ExperienceLevel) {
			if readiness > 0.7 {
				recommendation := &ConceptRecommendation{
					NodeID:             nodeID,
					RecommendationType: "advanced",
					Score:              readiness * 0.6,
					Confidence:         0.5,
					Reasoning:          []string{"?, "?},
					EstimatedTime:      time.Hour * 4,
					DifficultyMatch:    s.calculateDifficultyMatch(node, learner),
					SkillRelevance:     s.calculateSkillRelevance(node, learner),
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations
}

func (s *KnowledgeGraphService) getRelatedNodes(nodeID uuid.UUID, graph *entities.KnowledgeGraph) []uuid.UUID {
	var relatedNodes []uuid.UUID

	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypeRelatedTo {
			if relation.FromNodeID == nodeID {
				relatedNodes = append(relatedNodes, relation.ToNodeID)
			} else if relation.ToNodeID == nodeID {
				relatedNodes = append(relatedNodes, relation.FromNodeID)
			}
		}
	}

	return relatedNodes
}

func (s *KnowledgeGraphService) calculateDifficultyMatch(node *entities.KnowledgeNode, learner *entities.Learner) float64 {
	difficultyDiff := math.Abs(float64(node.DifficultyLevel) - float64(learner.ExperienceLevel))
	return math.Max(0, 1.0 - difficultyDiff/10.0)
}

func (s *KnowledgeGraphService) calculateSkillRelevance(node *entities.KnowledgeNode, learner *entities.Learner) float64 {
	relevantSkills := 0
	totalSkills := len(node.Skills)

	if totalSkills == 0 {
		return 0.5
	}

	for _, skill := range node.Skills {
		for _, goal := range learner.LearningGoals {
			if goal.TargetSkill == skill && goal.IsActive {
				relevantSkills++
				break
			}
		}
	}

	return float64(relevantSkills) / float64(totalSkills)
}

