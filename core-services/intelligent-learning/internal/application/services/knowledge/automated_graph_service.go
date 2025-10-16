package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// CrossModalServiceInterface ?
type CrossModalServiceInterface interface {
	ProcessCrossModalInference(ctx context.Context, request *CrossModalInferenceRequest) (*CrossModalInferenceResponse, error)
}

// CrossModalInferenceRequest ?
type CrossModalInferenceRequest struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Options   map[string]interface{} `json:"options"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

// CrossModalInferenceResponse ?
type CrossModalInferenceResponse struct {
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result"`
	Error   string                 `json:"error,omitempty"`
}

// LearnerProfile ?
type LearnerProfile struct {
	UserID             string                 `json:"user_id"`
	LearningStyle      string                 `json:"learning_style"`
	PreferredDifficulty string                `json:"preferred_difficulty"`
	Interests          []string               `json:"interests"`
	Goals              []string               `json:"goals"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// AutomatedKnowledgeGraphService ?
type AutomatedKnowledgeGraphService struct {
	graphRepo           repositories.KnowledgeGraphRepository
	contentRepo         repositories.LearningContentRepository
	crossModalService   CrossModalServiceInterface
	inferenceEngine     *IntelligentRelationInferenceEngine
	config              *AutomatedGraphConfig
	cache               *AutomatedGraphCache
	metrics             *AutomatedGraphMetrics
}

// AutomatedGraphConfig ?
type AutomatedGraphConfig struct {
	MinConfidenceThreshold    float64 `json:"min_confidence_threshold"`    // ?
	MaxRelationsPerNode       int     `json:"max_relations_per_node"`      // 
	AutoInferenceEnabled      bool    `json:"auto_inference_enabled"`      // 
	SemanticSimilarityWeight  float64 `json:"semantic_similarity_weight"`  // ?
	StructuralSimilarityWeight float64 `json:"structural_similarity_weight"` // ?
	ContentAnalysisDepth      int     `json:"content_analysis_depth"`      // 
	BatchProcessingSize       int     `json:"batch_processing_size"`       // ?
	UpdateInterval            int     `json:"update_interval"`             // ()
}

// AutomatedGraphCache ?
type AutomatedGraphCache struct {
	NodeEmbeddings     map[uuid.UUID][]float64           `json:"node_embeddings"`     // 
	RelationScores     map[string]float64                `json:"relation_scores"`     // 
	InferenceResults   map[string]*InferenceResult       `json:"inference_results"`   // 
	SemanticClusters   map[string][]uuid.UUID            `json:"semantic_clusters"`   // 
	LastUpdated        time.Time                         `json:"last_updated"`        // ?
	ProcessingQueue    []uuid.UUID                       `json:"processing_queue"`    // 
}

// AutomatedGraphMetrics ?
type AutomatedGraphMetrics struct {
	NodesProcessed        int64     `json:"nodes_processed"`        // 
	RelationsInferred     int64     `json:"relations_inferred"`     // ?
	SuccessfulInferences  int64     `json:"successful_inferences"`  // ?
	FailedInferences      int64     `json:"failed_inferences"`      // ?
	AverageConfidence     float64   `json:"average_confidence"`     // ?
	ProcessingTime        int64     `json:"processing_time"`        // ()
	LastProcessingTime    time.Time `json:"last_processing_time"`   // ?
	QualityScore          float64   `json:"quality_score"`          // 
}

// InferenceResult 
type InferenceResult struct {
	FromNodeID    uuid.UUID                `json:"from_node_id"`    // ID
	ToNodeID      uuid.UUID                `json:"to_node_id"`      // ID
	RelationType  entities.RelationType    `json:"relation_type"`   // 
	Confidence    float64                  `json:"confidence"`      // ?
	Evidence      []string                 `json:"evidence"`        // 
	Reasoning     string                   `json:"reasoning"`       // 
	Metadata      map[string]interface{}   `json:"metadata"`        // ?
	CreatedAt     time.Time                `json:"created_at"`      // 
}

// AutoBuildRequest 
type AutoBuildRequest struct {
	ContentIDs        []uuid.UUID            `json:"content_ids"`        // ID
	AnalysisDepth     int                    `json:"analysis_depth"`     // 
	EnableInference   bool                   `json:"enable_inference"`   // 
	CustomRules       []InferenceRule        `json:"custom_rules"`       // ?
	Options           map[string]interface{} `json:"options"`            // 
}

// InferenceRule 
type InferenceRule struct {
	ID          string                 `json:"id"`          // ID
	Name        string                 `json:"name"`        // 
	Description string                 `json:"description"` // 
	Conditions  []RuleCondition        `json:"conditions"`  // 
	Actions     []RuleAction           `json:"actions"`     // 
	Priority    int                    `json:"priority"`    // ?
	Enabled     bool                   `json:"enabled"`     // 
}

// RuleCondition 
type RuleCondition struct {
	Type      string      `json:"type"`      // 
	Field     string      `json:"field"`     // 
	Operator  string      `json:"operator"`  // ?
	Value     interface{} `json:"value"`     // ?
	Weight    float64     `json:"weight"`    // 
}

// RuleAction 
type RuleAction struct {
	Type       string                 `json:"type"`       // 
	Parameters map[string]interface{} `json:"parameters"` // 
}

// AutoBuildResponse 
type AutoBuildResponse struct {
	NodesCreated      int                            `json:"nodes_created"`      // 
	RelationsCreated  int                            `json:"relations_created"`  // 
	InferenceResults  []*InferenceResult             `json:"inference_results"`  // 
	QualityMetrics    *KnowledgeGraphQualityMetrics  `json:"quality_metrics"`    // 
	ProcessingTime    int64                          `json:"processing_time"`    // 
	Warnings          []string                       `json:"warnings"`           // 
}

// KnowledgeGraphQualityMetrics 
type KnowledgeGraphQualityMetrics struct {
	Completeness      float64 `json:"completeness"`      // ?
	Consistency       float64 `json:"consistency"`       // ?
	Accuracy          float64 `json:"accuracy"`          // ?
	Relevance         float64 `json:"relevance"`         // ?
	Coverage          float64 `json:"coverage"`          // ?
	Redundancy        float64 `json:"redundancy"`        // ?
	OverallScore      float64 `json:"overall_score"`     // 
}

// NewAutomatedKnowledgeGraphService ?
func NewAutomatedKnowledgeGraphService(
	graphRepo repositories.KnowledgeGraphRepository,
	contentRepo repositories.LearningContentRepository,
	crossModalService CrossModalServiceInterface,
) *AutomatedKnowledgeGraphService {
	// 
	inferenceEngine := NewIntelligentRelationInferenceEngine(crossModalService)
	config := &AutomatedGraphConfig{
		MinConfidenceThreshold:     0.7,
		MaxRelationsPerNode:        20,
		AutoInferenceEnabled:       true,
		SemanticSimilarityWeight:   0.6,
		StructuralSimilarityWeight: 0.4,
		ContentAnalysisDepth:       3,
		BatchProcessingSize:        50,
		UpdateInterval:             30,
	}

	cache := &AutomatedGraphCache{
		NodeEmbeddings:   make(map[uuid.UUID][]float64),
		RelationScores:   make(map[string]float64),
		InferenceResults: make(map[string]*InferenceResult),
		SemanticClusters: make(map[string][]uuid.UUID),
		ProcessingQueue:  make([]uuid.UUID, 0),
		LastUpdated:      time.Now(),
	}

	metrics := &AutomatedGraphMetrics{
		LastProcessingTime: time.Now(),
	}

	return &AutomatedKnowledgeGraphService{
		graphRepo:         graphRepo,
		contentRepo:       contentRepo,
		crossModalService: crossModalService,
		inferenceEngine:   inferenceEngine,
		config:            config,
		cache:             cache,
		metrics:           metrics,
	}
}

// AutoBuildFromContent ?
func (s *AutomatedKnowledgeGraphService) AutoBuildFromContent(ctx context.Context, req *AutoBuildRequest) (*AutoBuildResponse, error) {
	startTime := time.Now()
	response := &AutoBuildResponse{
		InferenceResults: make([]*InferenceResult, 0),
		Warnings:         make([]string, 0),
	}

	// 1. 
	nodes, err := s.extractKnowledgeNodes(ctx, req.ContentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to extract knowledge nodes: %w", err)
	}
	response.NodesCreated = len(nodes)

	// 2. 
	if err := s.generateNodeEmbeddings(ctx, nodes); err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("Failed to generate embeddings: %v", err))
	}

	// 3. ?
	relations, inferenceResults := s.inferRelations(ctx, nodes, req.CustomRules)
	response.RelationsCreated = len(relations)
	response.InferenceResults = inferenceResults

	// 4. ?
	if err := s.validateAndOptimizeGraph(ctx, nodes, relations); err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("Graph optimization warning: %v", err))
	}

	// 5. 
	response.QualityMetrics = s.calculateQualityMetrics(nodes, relations, inferenceResults)

	// 6. 
	s.updateMetrics(len(nodes), len(relations), len(inferenceResults))

	response.ProcessingTime = time.Since(startTime).Milliseconds()
	return response, nil
}

// extractKnowledgeNodes 
func (s *AutomatedKnowledgeGraphService) extractKnowledgeNodes(ctx context.Context, contentIDs []uuid.UUID) ([]*entities.KnowledgeNode, error) {
	nodes := make([]*entities.KnowledgeNode, 0)

	for _, contentID := range contentIDs {
		content, err := s.contentRepo.GetByID(ctx, contentID)
		if err != nil {
			continue
		}

		// AI
		analysisReq := &CrossModalInferenceRequest{
			Type: "content_analysis",
			Data: map[string]interface{}{
				"content":     content.Content,
				"title":       content.Title,
				"description": content.Description,
				"type":        content.Type,
			},
			Options: map[string]interface{}{
				"extract_concepts": true,
				"identify_skills":  true,
				"analyze_topics":   true,
				"depth":           s.config.ContentAnalysisDepth,
			},
		}

		analysisResp, err := s.crossModalService.ProcessCrossModalInference(ctx, analysisReq)
		if err != nil {
			continue
		}

		// ?
		if concepts, ok := analysisResp.Result["concepts"].([]interface{}); ok {
			for _, concept := range concepts {
				if conceptMap, ok := concept.(map[string]interface{}); ok {
					node := s.createNodeFromConcept(contentID, conceptMap)
					if node != nil {
						nodes = append(nodes, node)
					}
				}
			}
		}
	}

	return nodes, nil
}

// createNodeFromConcept ?
func (s *AutomatedKnowledgeGraphService) createNodeFromConcept(contentID uuid.UUID, concept map[string]interface{}) *entities.KnowledgeNode {
	name, _ := concept["name"].(string)
	if name == "" {
		return nil
	}

	nodeType := entities.NodeTypeConcept
	if typeStr, ok := concept["type"].(string); ok {
		switch typeStr {
		case "skill":
			nodeType = entities.NodeTypeSkill
		case "topic":
			nodeType = entities.NodeTypeTopic
		case "subject":
			nodeType = entities.NodeTypeSubject
		}
	}

	difficulty := entities.DifficultyBeginner
	if diffStr, ok := concept["difficulty"].(string); ok {
		switch diffStr {
		case "intermediate":
			difficulty = entities.DifficultyIntermediate
		case "advanced":
			difficulty = entities.DifficultyAdvanced
		case "expert":
			difficulty = entities.DifficultyExpert
		}
	}

	node := &entities.KnowledgeNode{
		ID:              uuid.New(),
		Name:            name,
		Type:            nodeType,
		DifficultyLevel: difficulty,
		Subject:         concept["subject"].(string),
		Description:     concept["description"].(string),
		Metadata: map[string]interface{}{
			"source_content_id": contentID,
			"confidence":        concept["confidence"],
			"keywords":          concept["keywords"],
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if keywords, ok := concept["keywords"].([]interface{}); ok {
		for _, keyword := range keywords {
			if keywordStr, ok := keyword.(string); ok {
				node.Keywords = append(node.Keywords, keywordStr)
			}
		}
	}

	if tags, ok := concept["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				node.Tags = append(node.Tags, tagStr)
			}
		}
	}

	return node
}

// generateNodeEmbeddings 
func (s *AutomatedKnowledgeGraphService) generateNodeEmbeddings(ctx context.Context, nodes []*entities.KnowledgeNode) error {
	for _, node := range nodes {
		// 
		text := fmt.Sprintf("%s %s %s", node.Name, node.Description, strings.Join(node.Keywords, " "))
		
		// AI
		embeddingReq := &CrossModalInferenceRequest{
			Type: "text_embedding",
			Data: map[string]interface{}{
				"text": text,
			},
		}

		embeddingResp, err := s.crossModalService.ProcessCrossModalInference(ctx, embeddingReq)
		if err != nil {
			continue
		}

		if embedding, ok := embeddingResp.Result["embedding"].([]interface{}); ok {
			vector := make([]float64, len(embedding))
			for i, val := range embedding {
				if floatVal, ok := val.(float64); ok {
					vector[i] = floatVal
				}
			}
			s.cache.NodeEmbeddings[node.ID] = vector
		}
	}

	return nil
}

// inferRelations ?
func (s *AutomatedKnowledgeGraphService) inferRelations(ctx context.Context, nodes []*entities.KnowledgeNode, customRules []InferenceRule) ([]*entities.KnowledgeRelation, []*InferenceResult) {
	relations := make([]*entities.KnowledgeRelation, 0)
	inferenceResults := make([]*InferenceResult, 0)

	// ?
	similarityMatrix := s.calculateSimilarityMatrix(nodes)

	// 
	for i, nodeA := range nodes {
		for j, nodeB := range nodes {
			if i >= j {
				continue
			}

			// ?
			similarity := similarityMatrix[i][j]
			if similarity > s.config.MinConfidenceThreshold {
				relationType, confidence := s.inferRelationType(nodeA, nodeB)
				if confidence > s.config.MinConfidenceThreshold {
					relation := &entities.KnowledgeRelation{
						ID:          uuid.New(),
						FromNodeID:  nodeA.ID,
						ToNodeID:    nodeB.ID,
						Type:        relationType,
						Weight:      confidence,
						Confidence:  confidence,
						Description: fmt.Sprintf("Inferred relation based on similarity: %.2f", similarity),
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}

					relations = append(relations, relation)

					inferenceResult := &InferenceResult{
						FromNodeID:   nodeA.ID,
						ToNodeID:     nodeB.ID,
						RelationType: relationType,
						Confidence:   confidence,
						Evidence:     []string{fmt.Sprintf("Semantic similarity: %.2f", similarity)},
						Reasoning:    "Based on semantic and structural similarity analysis",
						Metadata: map[string]interface{}{
							"similarity_score": similarity,
							"method":          "automated_inference",
						},
						CreatedAt: time.Now(),
					}

					inferenceResults = append(inferenceResults, inferenceResult)
				}
			}

			// ?
			for _, rule := range customRules {
				if rule.Enabled && s.evaluateRule(rule, nodeA, nodeB) {
					// 
					s.executeRuleActions(rule.Actions, nodeA, nodeB, &relations, &inferenceResults)
				}
			}
		}
	}

	return relations, inferenceResults
}

// InferRelations  - 
func (s *AutomatedKnowledgeGraphService) InferRelations(ctx context.Context, nodes []*entities.KnowledgeNode, existingRelations []*entities.KnowledgeRelation) ([]*InferenceResult, error) {
	// 
	inferenceReq := &InferenceRequest{
		Nodes:             nodes,
		ExistingRelations: existingRelations,
		Context: &InferenceContext{
			Domain:  "learning",
			Subject: "knowledge_graph",
		},
		Options: &InferenceOptions{
			MaxDepth:            3,
			MinConfidence:       s.config.MinConfidenceThreshold,
			EnableExplanation:   true,
			EnableUncertainty:   true,
			EnableContradiction: true,
			ParallelProcessing:  true,
		},
	}

	// 洦
	inferenceResp, err := s.inferenceEngine.ProcessInference(ctx, inferenceReq)
	if err != nil {
		return nil, fmt.Errorf("inference engine failed: %w", err)
	}

	// 
	results := make([]*InferenceResult, 0)
	for _, inferredRel := range inferenceResp.InferredRelations {
		result := &InferenceResult{
			FromNodeID:   inferredRel.FromNodeID,
			ToNodeID:     inferredRel.ToNodeID,
			RelationType: inferredRel.RelationType,
			Confidence:   inferredRel.Confidence,
			Evidence:     inferredRel.Evidence,
			Reasoning:    strings.Join(inferredRel.Reasoning, "; "),
			Metadata: map[string]interface{}{
				"inference_chain":  inferenceResp.InferenceChain,
				"quality_metrics":  inferenceResp.QualityMetrics,
				"explanations":     inferenceResp.Explanations,
				"processing_time":  inferenceResp.ProcessingTime,
			},
			CreatedAt: time.Now(),
		}
		results = append(results, result)
	}

	return results, nil
}

// InferRelationsWithContext ?
func (s *AutomatedKnowledgeGraphService) InferRelationsWithContext(ctx context.Context, nodes []*entities.KnowledgeNode, existingRelations []*entities.KnowledgeRelation, learnerProfile *LearnerProfile) ([]*InferenceResult, error) {
	// ?
	inferenceReq := &InferenceRequest{
		Nodes:             nodes,
		ExistingRelations: existingRelations,
		Context: &InferenceContext{
			Domain:  "learning",
			Subject: "knowledge_graph",
			LearnerProfile: learnerProfile,
		},
		Options: &InferenceOptions{
			MaxDepth:            3,
			MinConfidence:       s.config.MinConfidenceThreshold,
			EnableExplanation:   true,
			EnableUncertainty:   true,
			EnableContradiction: true,
			ParallelProcessing:  true,
		},
	}

	// 洦
	inferenceResp, err := s.inferenceEngine.ProcessInference(ctx, inferenceReq)
	if err != nil {
		return nil, fmt.Errorf("inference engine failed: %w", err)
	}

	// 
	results := make([]*InferenceResult, 0)
	for _, inferredRel := range inferenceResp.InferredRelations {
		result := &InferenceResult{
			FromNodeID:   inferredRel.FromNodeID,
			ToNodeID:     inferredRel.ToNodeID,
			RelationType: inferredRel.RelationType,
			Confidence:   inferredRel.Confidence,
			Evidence:     inferredRel.Evidence,
			Reasoning:    strings.Join(inferredRel.Reasoning, "; "),
			Metadata: map[string]interface{}{
				"inference_chain":  inferenceResp.InferenceChain,
				"quality_metrics":  inferenceResp.QualityMetrics,
				"explanations":     inferenceResp.Explanations,
				"processing_time":  inferenceResp.ProcessingTime,
				"personalized":     true,
				"learner_profile":  learnerProfile,
			},
			CreatedAt: time.Now(),
		}
		results = append(results, result)
	}

	return results, nil
}

// calculateSimilarityMatrix ?
func (s *AutomatedKnowledgeGraphService) calculateSimilarityMatrix(nodes []*entities.KnowledgeNode) [][]float64 {
	n := len(nodes)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
	}

	for i, nodeA := range nodes {
		for j, nodeB := range nodes {
			if i == j {
				matrix[i][j] = 1.0
				continue
			}

			// ?
			semanticSim := s.calculateSemanticSimilarity(nodeA, nodeB)
			
			// ?
			structuralSim := s.calculateStructuralSimilarity(nodeA, nodeB)
			
			// 
			similarity := s.config.SemanticSimilarityWeight*semanticSim + 
						 s.config.StructuralSimilarityWeight*structuralSim

			matrix[i][j] = similarity
		}
	}

	return matrix
}

// calculateSemanticSimilarity ?
func (s *AutomatedKnowledgeGraphService) calculateSemanticSimilarity(nodeA, nodeB *entities.KnowledgeNode) float64 {
	embeddingA, okA := s.cache.NodeEmbeddings[nodeA.ID]
	embeddingB, okB := s.cache.NodeEmbeddings[nodeB.ID]

	if !okA || !okB {
		// ?
		return s.calculateTextSimilarity(nodeA, nodeB)
	}

	// ?
	return s.cosineSimilarity(embeddingA, embeddingB)
}

// calculateStructuralSimilarity ?
func (s *AutomatedKnowledgeGraphService) calculateStructuralSimilarity(nodeA, nodeB *entities.KnowledgeNode) float64 {
	similarity := 0.0

	// ?
	if nodeA.Type == nodeB.Type {
		similarity += 0.3
	}

	// ?
	diffA := int(nodeA.DifficultyLevel)
	diffB := int(nodeB.DifficultyLevel)
	diffSim := 1.0 - math.Abs(float64(diffA-diffB))/4.0
	similarity += 0.2 * diffSim

	// ?
	if nodeA.Subject == nodeB.Subject {
		similarity += 0.3
	}

	// ?
	tagSim := s.calculateTagSimilarity(nodeA.Tags, nodeB.Tags)
	similarity += 0.2 * tagSim

	return similarity
}

// calculateTextSimilarity ?
func (s *AutomatedKnowledgeGraphService) calculateTextSimilarity(nodeA, nodeB *entities.KnowledgeNode) float64 {
	textA := strings.ToLower(nodeA.Name + " " + nodeA.Description)
	textB := strings.ToLower(nodeB.Name + " " + nodeB.Description)

	wordsA := strings.Fields(textA)
	wordsB := strings.Fields(textB)

	// Jaccard?
	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, word := range wordsA {
		setA[word] = true
	}
	for _, word := range wordsB {
		setB[word] = true
	}

	intersection := 0
	union := len(setA)

	for word := range setB {
		if setA[word] {
			intersection++
		} else {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// calculateTagSimilarity ?
func (s *AutomatedKnowledgeGraphService) calculateTagSimilarity(tagsA, tagsB []string) float64 {
	if len(tagsA) == 0 && len(tagsB) == 0 {
		return 1.0
	}

	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, tag := range tagsA {
		setA[tag] = true
	}
	for _, tag := range tagsB {
		setB[tag] = true
	}

	intersection := 0
	union := len(setA)

	for tag := range setB {
		if setA[tag] {
			intersection++
		} else {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// cosineSimilarity ?
func (s *AutomatedKnowledgeGraphService) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// inferRelationType 
func (s *AutomatedKnowledgeGraphService) inferRelationType(nodeA, nodeB *entities.KnowledgeNode) (entities.RelationType, float64) {
	// ?
	if nodeA.Type == entities.NodeTypeConcept && nodeB.Type == entities.NodeTypeConcept {
		return entities.RelationTypeRelatedTo, 0.8
	}

	if nodeA.Type == entities.NodeTypeSkill && nodeB.Type == entities.NodeTypeSkill {
		// ?
		if int(nodeA.DifficultyLevel) < int(nodeB.DifficultyLevel) {
			return entities.RelationTypePrerequisite, 0.9
		}
		return entities.RelationTypeRelatedTo, 0.7
	}

	if nodeA.Type == entities.NodeTypeTopic && nodeB.Type == entities.NodeTypeConcept {
		return entities.RelationTypePartOf, 0.85
	}

	if nodeA.Type == entities.NodeTypeSubject && nodeB.Type == entities.NodeTypeTopic {
		return entities.RelationTypePartOf, 0.9
	}

	return entities.RelationTypeRelatedTo, 0.6
}

// evaluateRule 
func (s *AutomatedKnowledgeGraphService) evaluateRule(rule InferenceRule, nodeA, nodeB *entities.KnowledgeNode) bool {
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(condition, nodeA, nodeB) {
			return false
		}
	}
	return true
}

// evaluateCondition 
func (s *AutomatedKnowledgeGraphService) evaluateCondition(condition RuleCondition, nodeA, nodeB *entities.KnowledgeNode) bool {
	// 
	switch condition.Type {
	case "node_type":
		return s.evaluateNodeTypeCondition(condition, nodeA, nodeB)
	case "difficulty":
		return s.evaluateDifficultyCondition(condition, nodeA, nodeB)
	case "subject":
		return s.evaluateSubjectCondition(condition, nodeA, nodeB)
	default:
		return false
	}
}

// evaluateNodeTypeCondition 
func (s *AutomatedKnowledgeGraphService) evaluateNodeTypeCondition(condition RuleCondition, nodeA, nodeB *entities.KnowledgeNode) bool {
	expectedType := condition.Value.(string)
	switch condition.Field {
	case "nodeA_type":
		return string(nodeA.Type) == expectedType
	case "nodeB_type":
		return string(nodeB.Type) == expectedType
	default:
		return false
	}
}

// evaluateDifficultyCondition 
func (s *AutomatedKnowledgeGraphService) evaluateDifficultyCondition(condition RuleCondition, nodeA, nodeB *entities.KnowledgeNode) bool {
	switch condition.Operator {
	case "less_than":
		return int(nodeA.DifficultyLevel) < int(nodeB.DifficultyLevel)
	case "greater_than":
		return int(nodeA.DifficultyLevel) > int(nodeB.DifficultyLevel)
	case "equal":
		return nodeA.DifficultyLevel == nodeB.DifficultyLevel
	default:
		return false
	}
}

// evaluateSubjectCondition 
func (s *AutomatedKnowledgeGraphService) evaluateSubjectCondition(condition RuleCondition, nodeA, nodeB *entities.KnowledgeNode) bool {
	switch condition.Operator {
	case "equal":
		return nodeA.Subject == nodeB.Subject
	case "not_equal":
		return nodeA.Subject != nodeB.Subject
	default:
		return false
	}
}

// executeRuleActions 
func (s *AutomatedKnowledgeGraphService) executeRuleActions(actions []RuleAction, nodeA, nodeB *entities.KnowledgeNode, relations *[]*entities.KnowledgeRelation, inferenceResults *[]*InferenceResult) {
	for _, action := range actions {
		switch action.Type {
		case "create_relation":
			s.executeCreateRelationAction(action, nodeA, nodeB, relations, inferenceResults)
		}
	}
}

// executeCreateRelationAction 
func (s *AutomatedKnowledgeGraphService) executeCreateRelationAction(action RuleAction, nodeA, nodeB *entities.KnowledgeNode, relations *[]*entities.KnowledgeRelation, inferenceResults *[]*InferenceResult) {
	relationType := entities.RelationTypeRelatedTo
	if typeStr, ok := action.Parameters["relation_type"].(string); ok {
		switch typeStr {
		case "prerequisite":
			relationType = entities.RelationTypePrerequisite
		case "part_of":
			relationType = entities.RelationTypePartOf
		case "leads_to":
			relationType = entities.RelationTypeLeadsTo
		}
	}

	confidence := 0.8
	if conf, ok := action.Parameters["confidence"].(float64); ok {
		confidence = conf
	}

	relation := &entities.KnowledgeRelation{
		ID:          uuid.New(),
		FromNodeID:  nodeA.ID,
		ToNodeID:    nodeB.ID,
		Type:        relationType,
		Weight:      confidence,
		Confidence:  confidence,
		Description: "Created by inference rule",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	*relations = append(*relations, relation)

	inferenceResult := &InferenceResult{
		FromNodeID:   nodeA.ID,
		ToNodeID:     nodeB.ID,
		RelationType: relationType,
		Confidence:   confidence,
		Evidence:     []string{"Rule-based inference"},
		Reasoning:    "Applied custom inference rule",
		Metadata: map[string]interface{}{
			"method": "rule_based",
			"rule":   action.Parameters["rule_name"],
		},
		CreatedAt: time.Now(),
	}

	*inferenceResults = append(*inferenceResults, inferenceResult)
}

// validateAndOptimizeGraph ?
func (s *AutomatedKnowledgeGraphService) validateAndOptimizeGraph(ctx context.Context, nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) error {
	// ?
	if s.hasCyclicDependencies(nodes, relations) {
		return fmt.Errorf("cyclic dependencies detected in graph")
	}

	// 
	s.removeRedundantRelations(relations)

	// 
	s.optimizeRelationWeights(relations)

	return nil
}

// hasCyclicDependencies ?
func (s *AutomatedKnowledgeGraphService) hasCyclicDependencies(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) bool {
	// ?
	graph := make(map[uuid.UUID][]uuid.UUID)
	for _, relation := range relations {
		if relation.Type == entities.RelationTypePrerequisite {
			graph[relation.FromNodeID] = append(graph[relation.FromNodeID], relation.ToNodeID)
		}
	}

	// DFS?
	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	for _, node := range nodes {
		if !visited[node.ID] {
			if s.dfsHasCycle(node.ID, graph, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// dfsHasCycle DFS?
func (s *AutomatedKnowledgeGraphService) dfsHasCycle(nodeID uuid.UUID, graph map[uuid.UUID][]uuid.UUID, visited, recStack map[uuid.UUID]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	for _, neighbor := range graph[nodeID] {
		if !visited[neighbor] {
			if s.dfsHasCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// removeRedundantRelations 
func (s *AutomatedKnowledgeGraphService) removeRedundantRelations(relations []*entities.KnowledgeRelation) {
	// 
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Confidence > relations[j].Confidence
	})

	// 
	seen := make(map[string]bool)
	filtered := make([]*entities.KnowledgeRelation, 0)

	for _, relation := range relations {
		key := fmt.Sprintf("%s-%s-%s", relation.FromNodeID, relation.ToNodeID, relation.Type)
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, relation)
		}
	}

	// relations
	copy(relations, filtered)
}

// optimizeRelationWeights 
func (s *AutomatedKnowledgeGraphService) optimizeRelationWeights(relations []*entities.KnowledgeRelation) {
	for _, relation := range relations {
		// 
		adjustedWeight := relation.Weight

		// 
		switch relation.Type {
		case entities.RelationTypePrerequisite:
			adjustedWeight *= 1.1 // ?
		case entities.RelationTypePartOf:
			adjustedWeight *= 1.05
		}

		// 
		if adjustedWeight > 1.0 {
			adjustedWeight = 1.0
		}
		if adjustedWeight < 0.0 {
			adjustedWeight = 0.0
		}

		relation.Weight = adjustedWeight
	}
}

// calculateQualityMetrics 
func (s *AutomatedKnowledgeGraphService) calculateQualityMetrics(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation, inferenceResults []*InferenceResult) *KnowledgeGraphQualityMetrics {
	metrics := &KnowledgeGraphQualityMetrics{}

	// ?
	metrics.Completeness = s.calculateCompleteness(nodes, relations)

	// ?
	metrics.Consistency = s.calculateConsistency(relations)

	// ?
	metrics.Accuracy = s.calculateAccuracy(inferenceResults)

	// ?
	metrics.Relevance = s.calculateRelevance(nodes, relations)

	// 㸲?
	metrics.Coverage = s.calculateCoverage(nodes)

	// ?
	metrics.Redundancy = s.calculateRedundancy(relations)

	// 
	metrics.OverallScore = (metrics.Completeness + metrics.Consistency + metrics.Accuracy + 
						   metrics.Relevance + metrics.Coverage + (1.0-metrics.Redundancy)) / 6.0

	return metrics
}

// calculateCompleteness ?
func (s *AutomatedKnowledgeGraphService) calculateCompleteness(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) float64 {
	if len(nodes) == 0 {
		return 0.0
	}

	// ?
	connectedNodes := make(map[uuid.UUID]bool)
	for _, relation := range relations {
		connectedNodes[relation.FromNodeID] = true
		connectedNodes[relation.ToNodeID] = true
	}

	return float64(len(connectedNodes)) / float64(len(nodes))
}

// calculateConsistency ?
func (s *AutomatedKnowledgeGraphService) calculateConsistency(relations []*entities.KnowledgeRelation) float64 {
	if len(relations) == 0 {
		return 1.0
	}

	// ?
	consistentRelations := 0
	for _, relation := range relations {
		if relation.Confidence > 0.5 {
			consistentRelations++
		}
	}

	return float64(consistentRelations) / float64(len(relations))
}

// calculateAccuracy ?
func (s *AutomatedKnowledgeGraphService) calculateAccuracy(inferenceResults []*InferenceResult) float64 {
	if len(inferenceResults) == 0 {
		return 1.0
	}

	totalConfidence := 0.0
	for _, result := range inferenceResults {
		totalConfidence += result.Confidence
	}

	return totalConfidence / float64(len(inferenceResults))
}

// calculateRelevance ?
func (s *AutomatedKnowledgeGraphService) calculateRelevance(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) float64 {
	// ?
	relevantRelations := 0
	for _, relation := range relations {
		if relation.Type == entities.RelationTypePrerequisite || 
		   relation.Type == entities.RelationTypePartOf ||
		   relation.Type == entities.RelationTypeLeadsTo {
			relevantRelations++
		}
	}

	if len(relations) == 0 {
		return 1.0
	}

	return float64(relevantRelations) / float64(len(relations))
}

// calculateCoverage 㸲?
func (s *AutomatedKnowledgeGraphService) calculateCoverage(nodes []*entities.KnowledgeNode) float64 {
	// 㲻
	typeCount := make(map[entities.NodeType]int)
	for _, node := range nodes {
		typeCount[node.Type]++
	}

	// ?
	expectedTypes := 4 // concept, skill, topic, subject
	actualTypes := len(typeCount)

	return float64(actualTypes) / float64(expectedTypes)
}

// calculateRedundancy ?
func (s *AutomatedKnowledgeGraphService) calculateRedundancy(relations []*entities.KnowledgeRelation) float64 {
	if len(relations) == 0 {
		return 0.0
	}

	// ?
	uniqueRelations := make(map[string]bool)
	duplicates := 0

	for _, relation := range relations {
		key := fmt.Sprintf("%s-%s-%s", relation.FromNodeID, relation.ToNodeID, relation.Type)
		if uniqueRelations[key] {
			duplicates++
		} else {
			uniqueRelations[key] = true
		}
	}

	return float64(duplicates) / float64(len(relations))
}

// updateMetrics 
func (s *AutomatedKnowledgeGraphService) updateMetrics(nodesProcessed, relationsCreated, inferencesCount int) {
	s.metrics.NodesProcessed += int64(nodesProcessed)
	s.metrics.RelationsInferred += int64(relationsCreated)
	s.metrics.SuccessfulInferences += int64(inferencesCount)
	s.metrics.LastProcessingTime = time.Now()
}

// GetMetrics 
func (s *AutomatedKnowledgeGraphService) GetMetrics() *AutomatedGraphMetrics {
	return s.metrics
}

// UpdateConfig 
func (s *AutomatedKnowledgeGraphService) UpdateConfig(config *AutomatedGraphConfig) {
	s.config = config
}

