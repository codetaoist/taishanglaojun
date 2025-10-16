package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/application/services/interfaces"
)



// KnowledgeAnalysisService 
type KnowledgeAnalysisService struct {
	kgRepo              repositories.KnowledgeGraphRepository
	contentRepo         repositories.LearningContentRepository
	learnerRepo         repositories.LearnerRepository
	kgService           interfaces.KnowledgeGraphService
	analyticsService    interfaces.LearningAnalyticsService
}

// NewKnowledgeAnalysisService 
func NewKnowledgeAnalysisService(
	kgRepo repositories.KnowledgeGraphRepository,
	contentRepo repositories.LearningContentRepository,
	learnerRepo repositories.LearnerRepository,
	kgService interfaces.KnowledgeGraphService,
	analyticsService interfaces.LearningAnalyticsService,
) *KnowledgeAnalysisService {
	return &KnowledgeAnalysisService{
		kgRepo:           kgRepo,
		contentRepo:      contentRepo,
		learnerRepo:      learnerRepo,
		kgService:        kgService,
		analyticsService: analyticsService,
	}
}

// ConceptRelationshipAnalysisRequest 
type ConceptRelationshipAnalysisRequest struct {
	GraphID         uuid.UUID   `json:"graph_id" binding:"required"`
	ConceptIDs      []uuid.UUID `json:"concept_ids,omitempty"`
	AnalysisDepth   int         `json:"analysis_depth"`
	IncludeMetrics  bool        `json:"include_metrics"`
	RelationTypes   []string    `json:"relation_types,omitempty"`
	MinStrength     float64     `json:"min_strength"`
}

// ConceptRelationshipAnalysisResponse 
type ConceptRelationshipAnalysisResponse struct {
	GraphID           uuid.UUID                `json:"graph_id"`
	AnalysisTimestamp time.Time                `json:"analysis_timestamp"`
	ConceptClusters   []ConceptCluster         `json:"concept_clusters"`
	RelationshipMap   map[string][]Relationship `json:"relationship_map"`
	CentralConcepts   []ConceptImportance      `json:"central_concepts"`
	WeakConnections   []WeakConnection         `json:"weak_connections"`
	RecommendedLinks  []RecommendedLink        `json:"recommended_links"`
	AnalysisMetrics   ConceptAnalysisMetrics   `json:"analysis_metrics"`
}

// ConceptCluster 
type ConceptCluster struct {
	ClusterID     string      `json:"cluster_id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	ConceptIDs    []uuid.UUID `json:"concept_ids"`
	Cohesion      float64     `json:"cohesion"`
	Centrality    float64     `json:"centrality"`
	DifficultyRange struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
		Avg float64 `json:"avg"`
	} `json:"difficulty_range"`
	LearningSequence []uuid.UUID `json:"learning_sequence"`
}

// Relationship 
type Relationship struct {
	SourceID      uuid.UUID `json:"source_id"`
	TargetID      uuid.UUID `json:"target_id"`
	RelationType  string    `json:"relation_type"`
	Strength      float64   `json:"strength"`
	Confidence    float64   `json:"confidence"`
	Direction     string    `json:"direction"`
	Properties    map[string]interface{} `json:"properties"`
}

// ConceptImportance ?
type ConceptImportance struct {
	ConceptID         uuid.UUID `json:"concept_id"`
	ImportanceScore   float64   `json:"importance_score"`
	CentralityScore   float64   `json:"centrality_score"`
	ConnectivityScore float64   `json:"connectivity_score"`
	InfluenceScore    float64   `json:"influence_score"`
	PrerequisiteCount int       `json:"prerequisite_count"`
	DependentCount    int       `json:"dependent_count"`
	LearningImpact    float64   `json:"learning_impact"`
}

// WeakConnection ?
type WeakConnection struct {
	ConceptID1    uuid.UUID `json:"concept_id1"`
	ConceptID2    uuid.UUID `json:"concept_id2"`
	CurrentStrength float64 `json:"current_strength"`
	PotentialStrength float64 `json:"potential_strength"`
	ReasonForWeakness string `json:"reason_for_weakness"`
	ImprovementSuggestion string `json:"improvement_suggestion"`
}

// RecommendedLink 
type RecommendedLink struct {
	SourceID      uuid.UUID `json:"source_id"`
	TargetID      uuid.UUID `json:"target_id"`
	RelationType  string    `json:"relation_type"`
	Confidence    float64   `json:"confidence"`
	Reasoning     []string  `json:"reasoning"`
	ExpectedBenefit float64 `json:"expected_benefit"`
	Priority      string    `json:"priority"`
}

// ConceptAnalysisMetrics 
type ConceptAnalysisMetrics struct {
	TotalConcepts       int     `json:"total_concepts"`
	TotalRelationships  int     `json:"total_relationships"`
	AverageConnectivity float64 `json:"average_connectivity"`
	GraphDensity        float64 `json:"graph_density"`
	ClusteringCoefficient float64 `json:"clustering_coefficient"`
	SmallWorldIndex     float64 `json:"small_world_index"`
	ModularityScore     float64 `json:"modularity_score"`
}

// DependencyGraphRequest ?
type DependencyGraphRequest struct {
	GraphID       uuid.UUID   `json:"graph_id" binding:"required"`
	RootConceptID *uuid.UUID  `json:"root_concept_id,omitempty"`
	TargetSkills  []string    `json:"target_skills,omitempty"`
	MaxDepth      int         `json:"max_depth"`
	IncludeOptional bool      `json:"include_optional"`
	DifficultyFilter struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"difficulty_filter"`
}

// DependencyGraphResponse ?
type DependencyGraphResponse struct {
	GraphID           uuid.UUID            `json:"graph_id"`
	RootConceptID     *uuid.UUID           `json:"root_concept_id"`
	DependencyLayers  []DependencyLayer    `json:"dependency_layers"`
	CriticalPath      []uuid.UUID          `json:"critical_path"`
	OptionalPaths     [][]uuid.UUID        `json:"optional_paths"`
	Prerequisites     map[uuid.UUID][]uuid.UUID `json:"prerequisites"`
	Dependents        map[uuid.UUID][]uuid.UUID `json:"dependents"`
	LearningSequence  []LearningStep       `json:"learning_sequence"`
	EstimatedDuration time.Duration        `json:"estimated_duration"`
	DifficultyProgression []DifficultyPoint `json:"difficulty_progression"`
}

// DependencyLayer ?
type DependencyLayer struct {
	Level       int         `json:"level"`
	ConceptIDs  []uuid.UUID `json:"concept_ids"`
	LayerType   string      `json:"layer_type"` // "foundation", "intermediate", "advanced", "specialization"
	Parallelizable bool     `json:"parallelizable"`
	EstimatedTime time.Duration `json:"estimated_time"`
}

// LearningStep 
type LearningStep struct {
	StepID        string      `json:"step_id"`
	ConceptID     uuid.UUID   `json:"concept_id"`
	Order         int         `json:"order"`
	IsRequired    bool        `json:"is_required"`
	Prerequisites []uuid.UUID `json:"prerequisites"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Difficulty    float64     `json:"difficulty"`
	LearningGoals []string    `json:"learning_goals"`
}

// DifficultyPoint ?
type DifficultyPoint struct {
	ConceptID   uuid.UUID `json:"concept_id"`
	Position    int       `json:"position"`
	Difficulty  float64   `json:"difficulty"`
	DifficultyJump float64 `json:"difficulty_jump"`
	IsBottleneck bool     `json:"is_bottleneck"`
}

// ContentRecommendationRequest 
type KnowledgeContentRecommendationRequest struct {
	LearnerID         uuid.UUID   `json:"learner_id" binding:"required"`
	ConceptID         *uuid.UUID  `json:"concept_id,omitempty"`
	LearningGoals     []string    `json:"learning_goals,omitempty"`
	PreferredTypes    []string    `json:"preferred_types,omitempty"`
	DifficultyRange   struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"difficulty_range"`
	MaxRecommendations int    `json:"max_recommendations"`
	IncludeReasoning   bool   `json:"include_reasoning"`
	PersonalizationLevel string `json:"personalization_level"` // "basic", "moderate", "high"
}

// ContentRecommendationResponse 
type KnowledgeContentRecommendationResponse struct {
	LearnerID       uuid.UUID              `json:"learner_id"`
	Recommendations []KnowledgeContentRecommendation `json:"recommendations"`
	PersonalizationFactors []KnowledgePersonalizationFactor `json:"personalization_factors"`
	LearningPath    []uuid.UUID            `json:"learning_path"`
	GeneratedAt     time.Time              `json:"generated_at"`
	ValidUntil      time.Time              `json:"valid_until"`
}

// KnowledgeContentRecommendation 
type KnowledgeContentRecommendation struct {
	ContentID       uuid.UUID `json:"content_id"`
	RecommendationType string `json:"recommendation_type"` // "next_step", "review", "challenge", "alternative"
	Score           float64   `json:"score"`
	Confidence      float64   `json:"confidence"`
	Reasoning       []string  `json:"reasoning"`
	PersonalizationScore float64 `json:"personalization_score"`
	DifficultyMatch float64   `json:"difficulty_match"`
	StyleMatch      float64   `json:"style_match"`
	GoalAlignment   float64   `json:"goal_alignment"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	Prerequisites   []uuid.UUID `json:"prerequisites"`
	LearningOutcomes []string  `json:"learning_outcomes"`
	Tags            []string  `json:"tags"`
}

// PersonalizationFactor 
type KnowledgePersonalizationFactor struct {
	FactorType  string  `json:"factor_type"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Impact      string  `json:"impact"`
}

// AnalyzeConceptRelationships 
func (s *KnowledgeAnalysisService) AnalyzeConceptRelationships(ctx context.Context, req *ConceptRelationshipAnalysisRequest) (*ConceptRelationshipAnalysisResponse, error) {
	// 
	graph, err := s.kgRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	response := &ConceptRelationshipAnalysisResponse{
		GraphID:           req.GraphID,
		AnalysisTimestamp: time.Now(),
		RelationshipMap:   make(map[string][]Relationship),
	}

	// 
	response.ConceptClusters = s.identifyConceptClusters(ctx, graph, req)

	// 
	response.RelationshipMap = s.buildRelationshipMap(ctx, graph, req)

	// 
	response.CentralConcepts = s.identifyCentralConcepts(ctx, graph, req)

	// ?
	response.WeakConnections = s.identifyWeakConnections(ctx, graph, req)

	// 
	response.RecommendedLinks = s.generateRecommendedLinks(ctx, graph, req)

	// 
	response.AnalysisMetrics = s.calculateAnalysisMetrics(ctx, graph)

	return response, nil
}

// BuildDependencyGraph ?
func (s *KnowledgeAnalysisService) BuildDependencyGraph(ctx context.Context, req *DependencyGraphRequest) (*DependencyGraphResponse, error) {
	// 
	graph, err := s.kgRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	response := &DependencyGraphResponse{
		GraphID:       req.GraphID,
		RootConceptID: req.RootConceptID,
		Prerequisites: make(map[uuid.UUID][]uuid.UUID),
		Dependents:    make(map[uuid.UUID][]uuid.UUID),
	}

	// 
	response.DependencyLayers = s.buildDependencyLayers(ctx, graph, req)

	// 
	response.CriticalPath = s.identifyCriticalPath(ctx, graph, req)

	// ?
	response.OptionalPaths = s.identifyOptionalPaths(ctx, graph, req)

	// ?
	response.Prerequisites, response.Dependents = s.buildDependencyMaps(ctx, graph, req)

	// 
	response.LearningSequence = s.generateLearningSequence(ctx, graph, req)

	// 
	response.EstimatedDuration = s.estimateLearningDuration(response.LearningSequence)

	// 
	response.DifficultyProgression = s.analyzeDifficultyProgression(response.LearningSequence)

	return response, nil
}

// RecommendContent 
func (s *KnowledgeAnalysisService) RecommendContent(ctx context.Context, req *KnowledgeContentRecommendationRequest) (*KnowledgeContentRecommendationResponse, error) {
	// ?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	response := &KnowledgeContentRecommendationResponse{
		LearnerID:   req.LearnerID,
		GeneratedAt: time.Now(),
		ValidUntil:  time.Now().Add(24 * time.Hour), // ?4
	}

	// ?
	learnerProfile := s.analyzeLearnerProfile(ctx, learner)

	// 
	response.PersonalizationFactors = s.generatePersonalizationFactors(learnerProfile, req)

	// ?
	candidateContent := s.getCandidateContent(ctx, req, learnerProfile)

	// 
	response.Recommendations = s.calculateRecommendationScores(ctx, candidateContent, learnerProfile, req)

	// 
	sort.Slice(response.Recommendations, func(i, j int) bool {
		return response.Recommendations[i].Score > response.Recommendations[j].Score
	})

	// 
	if req.MaxRecommendations > 0 && len(response.Recommendations) > req.MaxRecommendations {
		response.Recommendations = response.Recommendations[:req.MaxRecommendations]
	}

	// 
	response.LearningPath = s.generateContentLearningPath(response.Recommendations)

	return response, nil
}

// 

func (s *KnowledgeAnalysisService) identifyConceptClusters(ctx context.Context, graph *entities.KnowledgeGraph, req *ConceptRelationshipAnalysisRequest) []ConceptCluster {
	// 㷨
	clusters := []ConceptCluster{}
	
	// 㷨Louvain㷨黯
	// ?
	
	clusterID := 1
	for _, node := range graph.Nodes {
		if len(clusters) == 0 || !s.belongsToExistingCluster(node.ID, clusters) {
			cluster := ConceptCluster{
				ClusterID:   fmt.Sprintf("cluster-%d", clusterID),
				Name:        fmt.Sprintf(" %d", clusterID),
				Description: "",
				ConceptIDs:  []uuid.UUID{node.ID},
				Cohesion:    0.8, // ?
				Centrality:  0.6, // ?
			}
			
			// 
			cluster.DifficultyRange.Min = float64(node.DifficultyLevel)
			cluster.DifficultyRange.Max = float64(node.DifficultyLevel)
			cluster.DifficultyRange.Avg = float64(node.DifficultyLevel)
			
			clusters = append(clusters, cluster)
			clusterID++
		}
	}
	
	return clusters
}

func (s *KnowledgeAnalysisService) belongsToExistingCluster(nodeID uuid.UUID, clusters []ConceptCluster) bool {
	for _, cluster := range clusters {
		for _, conceptID := range cluster.ConceptIDs {
			if conceptID == nodeID {
				return true
			}
		}
	}
	return false
}

func (s *KnowledgeAnalysisService) buildRelationshipMap(ctx context.Context, graph *entities.KnowledgeGraph, req *ConceptRelationshipAnalysisRequest) map[string][]Relationship {
	relationshipMap := make(map[string][]Relationship)
	
	for _, relation := range graph.Relations {
		if req.MinStrength > 0 && relation.Weight < req.MinStrength {
			continue
		}
		
		relType := string(relation.Type)
		if len(req.RelationTypes) > 0 && !contains(req.RelationTypes, relType) {
			continue
		}
		
		relationship := Relationship{
			SourceID:     relation.FromNodeID,
			TargetID:     relation.ToNodeID,
			RelationType: relType,
			Strength:     relation.Weight,
			Confidence:   0.9, // ?
			Direction:    "directed",
			Properties:   relation.Metadata,
		}
		
		relationshipMap[relType] = append(relationshipMap[relType], relationship)
	}
	
	return relationshipMap
}

func (s *KnowledgeAnalysisService) identifyCentralConcepts(ctx context.Context, graph *entities.KnowledgeGraph, req *ConceptRelationshipAnalysisRequest) []ConceptImportance {
	concepts := []ConceptImportance{}
	
	for _, node := range graph.Nodes {
		// ?
		inDegree := s.calculateInDegree(node.ID, graph.Relations)
		outDegree := s.calculateOutDegree(node.ID, graph.Relations)
		
		importance := ConceptImportance{
			ConceptID:         node.ID,
			ImportanceScore:   float64(inDegree + outDegree) / float64(len(graph.Nodes)),
			CentralityScore:   s.calculateBetweennessCentrality(node.ID, graph),
			ConnectivityScore: float64(inDegree + outDegree),
			InfluenceScore:    float64(outDegree) / math.Max(1, float64(inDegree)),
			PrerequisiteCount: inDegree,
			DependentCount:    outDegree,
			LearningImpact:    0.8, // ?
		}
		
		concepts = append(concepts, importance)
	}
	
	// ?
	sort.Slice(concepts, func(i, j int) bool {
		return concepts[i].ImportanceScore > concepts[j].ImportanceScore
	})
	
	return concepts
}

func (s *KnowledgeAnalysisService) identifyWeakConnections(ctx context.Context, graph *entities.KnowledgeGraph, req *ConceptRelationshipAnalysisRequest) []WeakConnection {
	weakConnections := []WeakConnection{}
	
	// 
	for i, node1 := range graph.Nodes {
		for j, node2 := range graph.Nodes {
			if i >= j {
				continue
			}
			
			currentStrength := s.getRelationStrength(node1.ID, node2.ID, graph.Relations)
			potentialStrength := s.calculatePotentialStrength(node1, node2)
			
			if potentialStrength > currentStrength+0.3 { // ?
				weakConnection := WeakConnection{
					ConceptID1:            node1.ID,
					ConceptID2:            node2.ID,
					CurrentStrength:       currentStrength,
					PotentialStrength:     potentialStrength,
					ReasonForWeakness:     "?,
					ImprovementSuggestion: "?,
				}
				weakConnections = append(weakConnections, weakConnection)
			}
		}
	}
	
	return weakConnections
}

func (s *KnowledgeAnalysisService) generateRecommendedLinks(ctx context.Context, graph *entities.KnowledgeGraph, req *ConceptRelationshipAnalysisRequest) []RecommendedLink {
	recommendedLinks := []RecommendedLink{}
	
	// ?
	weakConnections := s.identifyWeakConnections(ctx, graph, req)
	
	for _, weak := range weakConnections {
		link := RecommendedLink{
			SourceID:        weak.ConceptID1,
			TargetID:        weak.ConceptID2,
			RelationType:    "related_to",
			Confidence:      weak.PotentialStrength,
			Reasoning:       []string{"", "?},
			ExpectedBenefit: weak.PotentialStrength - weak.CurrentStrength,
			Priority:        s.calculateLinkPriority(weak.PotentialStrength - weak.CurrentStrength),
		}
		recommendedLinks = append(recommendedLinks, link)
	}
	
	return recommendedLinks
}

func (s *KnowledgeAnalysisService) calculateAnalysisMetrics(ctx context.Context, graph *entities.KnowledgeGraph) ConceptAnalysisMetrics {
	nodeCount := len(graph.Nodes)
	relationCount := len(graph.Relations)
	
	// ?
	totalConnections := 0
	for _, node := range graph.Nodes {
		totalConnections += s.calculateInDegree(node.ID, graph.Relations) + s.calculateOutDegree(node.ID, graph.Relations)
	}
	avgConnectivity := float64(totalConnections) / float64(nodeCount) / 2 // 2?
	
	// ?
	maxPossibleEdges := nodeCount * (nodeCount - 1)
	density := float64(relationCount) / float64(maxPossibleEdges)
	
	return ConceptAnalysisMetrics{
		TotalConcepts:         nodeCount,
		TotalRelationships:    relationCount,
		AverageConnectivity:   avgConnectivity,
		GraphDensity:          density,
		ClusteringCoefficient: s.calculateClusteringCoefficient(graph),
		SmallWorldIndex:       s.calculateSmallWorldIndex(graph),
		ModularityScore:       s.calculateModularityScore(graph),
	}
}

// ...

func (s *KnowledgeAnalysisService) buildDependencyLayers(ctx context.Context, graph *entities.KnowledgeGraph, req *DependencyGraphRequest) []DependencyLayer {
	layers := []DependencyLayer{}
	
	// 
	visited := make(map[uuid.UUID]bool)
	level := 0
	
	for len(visited) < len(graph.Nodes) {
		currentLayer := DependencyLayer{
			Level:          level,
			ConceptIDs:     []uuid.UUID{},
			LayerType:      s.determineLayerType(level),
			Parallelizable: true,
			EstimatedTime:  0,
		}
		
		for _, node := range graph.Nodes {
			if visited[node.ID] {
				continue
			}
			
			// ?
			if s.allPrerequisitesSatisfied(node.ID, graph.Relations, visited) {
				currentLayer.ConceptIDs = append(currentLayer.ConceptIDs, node.ID)
				visited[node.ID] = true
				currentLayer.EstimatedTime += time.Duration(node.EstimatedHours) * time.Hour
			}
		}
		
		if len(currentLayer.ConceptIDs) > 0 {
			layers = append(layers, currentLayer)
		}
		level++
		
		// 
		if level > len(graph.Nodes) {
			break
		}
	}
	
	return layers
}

func (s *KnowledgeAnalysisService) identifyCriticalPath(ctx context.Context, graph *entities.KnowledgeGraph, req *DependencyGraphRequest) []uuid.UUID {
	// 
	criticalPath := []uuid.UUID{}
	
	// ?
	// 
	if req.RootConceptID != nil {
		criticalPath = append(criticalPath, *req.RootConceptID)
	}
	
	return criticalPath
}

func (s *KnowledgeAnalysisService) identifyOptionalPaths(ctx context.Context, graph *entities.KnowledgeGraph, req *DependencyGraphRequest) [][]uuid.UUID {
	optionalPaths := [][]uuid.UUID{}
	
	// 
	// ?
	
	return optionalPaths
}

func (s *KnowledgeAnalysisService) buildDependencyMaps(ctx context.Context, graph *entities.KnowledgeGraph, req *DependencyGraphRequest) (map[uuid.UUID][]uuid.UUID, map[uuid.UUID][]uuid.UUID) {
	prerequisites := make(map[uuid.UUID][]uuid.UUID)
	dependents := make(map[uuid.UUID][]uuid.UUID)
	
	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypePrerequisite {
			prerequisites[relation.ToNodeID] = append(prerequisites[relation.ToNodeID], relation.FromNodeID)
			dependents[relation.FromNodeID] = append(dependents[relation.FromNodeID], relation.ToNodeID)
		}
	}
	
	return prerequisites, dependents
}

func (s *KnowledgeAnalysisService) generateLearningSequence(ctx context.Context, graph *entities.KnowledgeGraph, req *DependencyGraphRequest) []LearningStep {
	sequence := []LearningStep{}
	
	// 
	layers := s.buildDependencyLayers(ctx, graph, req)
	
	stepOrder := 1
	for _, layer := range layers {
		for _, conceptID := range layer.ConceptIDs {
			node := s.findNodeByID(conceptID, graph.Nodes)
			if node != nil {
				step := LearningStep{
					StepID:        fmt.Sprintf("step-%d", stepOrder),
					ConceptID:     conceptID,
					Order:         stepOrder,
					IsRequired:    true,
					Prerequisites: s.getPrerequisites(conceptID, graph.Relations),
					EstimatedTime: time.Duration(node.EstimatedHours) * time.Hour,
					Difficulty:    float64(node.DifficultyLevel),
					LearningGoals: []string{node.Name}, // ?
				}
				sequence = append(sequence, step)
				stepOrder++
			}
		}
	}
	
	return sequence
}

func (s *KnowledgeAnalysisService) estimateLearningDuration(sequence []LearningStep) time.Duration {
	totalDuration := time.Duration(0)
	for _, step := range sequence {
		totalDuration += step.EstimatedTime
	}
	return totalDuration
}

func (s *KnowledgeAnalysisService) analyzeDifficultyProgression(sequence []LearningStep) []DifficultyPoint {
	points := []DifficultyPoint{}
	
	for i, step := range sequence {
		point := DifficultyPoint{
			ConceptID:  step.ConceptID,
			Position:   i + 1,
			Difficulty: step.Difficulty,
		}
		
		if i > 0 {
			point.DifficultyJump = step.Difficulty - sequence[i-1].Difficulty
			point.IsBottleneck = point.DifficultyJump > 0.3 // ?
		}
		
		points = append(points, point)
	}
	
	return points
}

// 

func (s *KnowledgeAnalysisService) analyzeLearnerProfile(ctx context.Context, learner *entities.Learner) map[string]interface{} {
	profile := make(map[string]interface{})
	
	// ?
	profile["learning_style"] = learner.Preferences.Style
	profile["difficulty_tolerance"] = learner.Preferences.DifficultyTolerance
	profile["session_duration"] = learner.Preferences.SessionDuration
	profile["skill_levels"] = learner.Skills
	
	return profile
}

func (s *KnowledgeAnalysisService) generatePersonalizationFactors(learnerProfile map[string]interface{}, req *KnowledgeContentRecommendationRequest) []KnowledgePersonalizationFactor {
	factors := []KnowledgePersonalizationFactor{
		{
			FactorType:  "learning_style",
			Description: "",
			Weight:      0.3,
			Impact:      "",
		},
		{
			FactorType:  "skill_level",
			Description: "?,
			Weight:      0.4,
			Impact:      "",
		},
		{
			FactorType:  "learning_goals",
			Description: "",
			Weight:      0.3,
			Impact:      "",
		},
	}
	
	return factors
}

func (s *KnowledgeAnalysisService) getCandidateContent(ctx context.Context, req *KnowledgeContentRecommendationRequest, learnerProfile map[string]interface{}) []*entities.LearningContent {
	// ?
	// 
	return []*entities.LearningContent{}
}

func (s *KnowledgeAnalysisService) calculateRecommendationScores(ctx context.Context, candidates []*entities.LearningContent, learnerProfile map[string]interface{}, req *KnowledgeContentRecommendationRequest) []KnowledgeContentRecommendation {
	recommendations := []KnowledgeContentRecommendation{}

	for _, content := range candidates {
		// 
		learningOutcomes := make([]string, len(content.LearningObjectives))
		for i, objective := range content.LearningObjectives {
			learningOutcomes[i] = objective.Description
		}

		recommendation := KnowledgeContentRecommendation{
			ContentID:            content.ID,
			RecommendationType:   "next_step",
			Score:                s.calculateContentScore(content, learnerProfile, req),
			Confidence:           0.8,
			Reasoning:            []string{"?, ""},
			PersonalizationScore: 0.7,
			DifficultyMatch:      0.8,
			StyleMatch:           0.6,
			GoalAlignment:        0.9,
			EstimatedTime:        time.Duration(content.EstimatedDuration) * time.Minute,
			Prerequisites:        []uuid.UUID{},
			LearningOutcomes:     learningOutcomes,
			Tags:                 content.Tags,
		}
		
		recommendations = append(recommendations, recommendation)
	}
	
	return recommendations
}

func (s *KnowledgeAnalysisService) generateContentLearningPath(recommendations []KnowledgeContentRecommendation) []uuid.UUID {
	path := []uuid.UUID{}
	for _, rec := range recommendations {
		path = append(path, rec.ContentID)
	}
	return path
}

// 

func (s *KnowledgeAnalysisService) calculateInDegree(nodeID uuid.UUID, relations []entities.KnowledgeRelation) int {
	count := 0
	for _, relation := range relations {
		if relation.ToNodeID == nodeID {
			count++
		}
	}
	return count
}

func (s *KnowledgeAnalysisService) calculateOutDegree(nodeID uuid.UUID, relations []entities.KnowledgeRelation) int {
	count := 0
	for _, relation := range relations {
		if relation.FromNodeID == nodeID {
			count++
		}
	}
	return count
}

func (s *KnowledgeAnalysisService) calculateBetweennessCentrality(nodeID uuid.UUID, graph *entities.KnowledgeGraph) float64 {
	// ?
	return 0.5 // ?
}

func (s *KnowledgeAnalysisService) getRelationStrength(nodeID1, nodeID2 uuid.UUID, relations []entities.KnowledgeRelation) float64 {
	for _, relation := range relations {
		if (relation.FromNodeID == nodeID1 && relation.ToNodeID == nodeID2) ||
			(relation.FromNodeID == nodeID2 && relation.ToNodeID == nodeID1) {
			return relation.Weight
		}
	}
	return 0.0
}

func (s *KnowledgeAnalysisService) calculatePotentialStrength(node1, node2 entities.KnowledgeNode) float64 {
	// ?
	// ?
	return 0.7 // ?
}

func (s *KnowledgeAnalysisService) calculateLinkPriority(benefit float64) string {
	if benefit > 0.7 {
		return "high"
	} else if benefit > 0.4 {
		return "medium"
	}
	return "low"
}

func (s *KnowledgeAnalysisService) calculateClusteringCoefficient(graph *entities.KnowledgeGraph) float64 {
	// 
	return 0.6 // ?
}

func (s *KnowledgeAnalysisService) calculateSmallWorldIndex(graph *entities.KnowledgeGraph) float64 {
	// ?
	return 0.8 // ?
}

func (s *KnowledgeAnalysisService) calculateModularityScore(graph *entities.KnowledgeGraph) float64 {
	// 黯?
	return 0.7 // ?
}

func (s *KnowledgeAnalysisService) determineLayerType(level int) string {
	switch {
	case level == 0:
		return "foundation"
	case level <= 2:
		return "intermediate"
	case level <= 4:
		return "advanced"
	default:
		return "specialization"
	}
}

func (s *KnowledgeAnalysisService) allPrerequisitesSatisfied(nodeID uuid.UUID, relations []entities.KnowledgeRelation, visited map[uuid.UUID]bool) bool {
	for _, relation := range relations {
		if relation.ToNodeID == nodeID && relation.Type == entities.RelationTypePrerequisite {
			if !visited[relation.FromNodeID] {
				return false
			}
		}
	}
	return true
}

func (s *KnowledgeAnalysisService) findNodeByID(nodeID uuid.UUID, nodes []entities.KnowledgeNode) *entities.KnowledgeNode {
	for i, node := range nodes {
		if node.ID == nodeID {
			return &nodes[i]
		}
	}
	return nil
}

func (s *KnowledgeAnalysisService) getPrerequisites(nodeID uuid.UUID, relations []entities.KnowledgeRelation) []uuid.UUID {
	prerequisites := []uuid.UUID{}
	for _, relation := range relations {
		if relation.ToNodeID == nodeID && relation.Type == entities.RelationTypePrerequisite {
			prerequisites = append(prerequisites, relation.FromNodeID)
		}
	}
	return prerequisites
}

func (s *KnowledgeAnalysisService) calculateContentScore(content *entities.LearningContent, learnerProfile map[string]interface{}, req *KnowledgeContentRecommendationRequest) float64 {
	// 
	score := 0.0
	
	// 
	difficultyScore := 1.0 - math.Abs(float64(content.Difficulty)-req.DifficultyRange.Min)
	score += difficultyScore * 0.4
	
	// 
	typeScore := 0.8 // ?
	score += typeScore * 0.3
	
	// 
	goalScore := 0.9 // ?
	score += goalScore * 0.3
	
	return math.Min(score, 1.0)
}

// 
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

