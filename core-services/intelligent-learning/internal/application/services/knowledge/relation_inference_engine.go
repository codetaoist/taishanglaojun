package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// EvidenceType 
type EvidenceType string

const (
	EvidenceTypeStatistical EvidenceType = "statistical"
	EvidenceTypeObservation EvidenceType = "observation"
	EvidenceTypeExperimental EvidenceType = "experimental"
	EvidenceTypeCorrelation EvidenceType = "correlation"
)

// Evidence 
type Evidence struct {
	ID          string                 `json:"id"`          // ID
	Type        EvidenceType           `json:"type"`        // 
	Source      string                 `json:"source"`      // 
	Content     interface{}            `json:"content"`     // 
	Reliability float64               `json:"reliability"` // ?
	Confidence  float64               `json:"confidence"`  // ?
	Timestamp   time.Time             `json:"timestamp"`   // ?
	Metadata    map[string]interface{} `json:"metadata"`    // ?
}

// IntelligentRelationInferenceEngine 
type IntelligentRelationInferenceEngine struct {
	crossModalService CrossModalServiceInterface
	config           *InferenceEngineConfig
	cache            *InferenceCache
	metrics          *InferenceMetrics
	rules            []*AdvancedInferenceRule
}

// InferenceEngineConfig 
type InferenceEngineConfig struct {
	MinConfidenceThreshold      float64 `json:"min_confidence_threshold"`      // ?
	MaxInferenceDepth          int     `json:"max_inference_depth"`           // ?
	EnableTransitiveInference  bool    `json:"enable_transitive_inference"`   // ?
	EnableContradictionCheck   bool    `json:"enable_contradiction_check"`    // ?
	EnableUncertaintyReasoning bool    `json:"enable_uncertainty_reasoning"`  // ?
	ContextWindowSize          int     `json:"context_window_size"`           // ?
	ParallelProcessing         bool    `json:"parallel_processing"`           // 
	CacheEnabled               bool    `json:"cache_enabled"`                 // 
}

// InferenceCache 
type InferenceCache struct {
	RelationProbabilities map[string]float64                    `json:"relation_probabilities"` // 
	InferenceChains      map[string][]*InferenceStep           `json:"inference_chains"`       // ?
	ContextEmbeddings    map[string][]float64                  `json:"context_embeddings"`     // ?
	CachedResults        map[string]*domainServices.CachedInferenceResult     `json:"cached_results"`         // 
	LastUpdated          time.Time                             `json:"last_updated"`           // ?
}

// InferenceMetrics 
type InferenceMetrics struct {
	TotalInferences       int64     `json:"total_inferences"`       // ?
	SuccessfulInferences  int64     `json:"successful_inferences"`  // 
	FailedInferences      int64     `json:"failed_inferences"`      // 
	AverageConfidence     float64   `json:"average_confidence"`     // ?
	AverageProcessingTime int64     `json:"average_processing_time"` // 
	ContradictionsFound   int64     `json:"contradictions_found"`   // 
	TransitiveInferences  int64     `json:"transitive_inferences"`  // ?
	LastInferenceTime     time.Time `json:"last_inference_time"`    // ?
}

// AdvancedInferenceRule 
type AdvancedInferenceRule struct {
	ID                string                    `json:"id"`                // ID
	Name              string                    `json:"name"`              // 
	Description       string                    `json:"description"`       // 
	Type              InferenceRuleType         `json:"type"`              // 
	Conditions        []*ComplexCondition       `json:"conditions"`        // 
	Conclusions       []*InferenceConclusion    `json:"conclusions"`       // 
	Priority          int                       `json:"priority"`          // ?
	Confidence        float64                   `json:"confidence"`        // ?
	Enabled           bool                      `json:"enabled"`           // 
	ContextRequired   bool                      `json:"context_required"`  // 
	MetaReasoning     bool                      `json:"meta_reasoning"`    // 
}

// InferenceRuleType 
type InferenceRuleType string

const (
	RuleTypeDeductive    InferenceRuleType = "deductive"    // 
	RuleTypeInductive    InferenceRuleType = "inductive"    // 
	RuleTypeAbductive    InferenceRuleType = "abductive"    // 
	RuleTypeAnalogical   InferenceRuleType = "analogical"   // 
	RuleTypeTransitive   InferenceRuleType = "transitive"   // ?
	RuleTypeCausal       InferenceRuleType = "causal"       // 
	RuleTypeStatistical  InferenceRuleType = "statistical"  // 
	RuleTypeOntological  InferenceRuleType = "ontological"  // 
)

// ComplexCondition 
type ComplexCondition struct {
	ID           string                 `json:"id"`           // ID
	Type         ConditionType          `json:"type"`         // 
	Operator     LogicalOperator        `json:"operator"`     // ?
	Predicates   []*Predicate          `json:"predicates"`   // 
	SubConditions []*ComplexCondition   `json:"sub_conditions"` // ?
	Weight       float64               `json:"weight"`       // 
	Negated      bool                  `json:"negated"`      // 
}

// ConditionType 
type ConditionType string

const (
	ConditionTypeAtomic    ConditionType = "atomic"    // 
	ConditionTypeComposite ConditionType = "composite" // 
	ConditionTypeModal     ConditionType = "modal"     // ?
	ConditionTypeTemporal  ConditionType = "temporal"  // 
)

// LogicalOperator ?
type LogicalOperator string

const (
	OperatorAND LogicalOperator = "AND"
	OperatorOR  LogicalOperator = "OR"
	OperatorNOT LogicalOperator = "NOT"
	OperatorXOR LogicalOperator = "XOR"
)

// Predicate 
type Predicate struct {
	Subject   *PredicateElement `json:"subject"`   // 
	Predicate string           `json:"predicate"` // 
	Object    *PredicateElement `json:"object"`    // 
	Modality  string           `json:"modality"`  // ?
	Certainty float64          `json:"certainty"` // ?
}

// PredicateElement 
type PredicateElement struct {
	Type       string      `json:"type"`       // 
	Value      interface{} `json:"value"`      // ?
	Variable   bool        `json:"variable"`   // ?
	Quantifier string      `json:"quantifier"` // 
}

// InferenceConclusion 
type InferenceConclusion struct {
	Type         ConclusionType         `json:"type"`         // 
	Relation     *InferredRelation      `json:"relation"`     // 
	Confidence   float64               `json:"confidence"`   // ?
	Evidence     []*domainServices.Evidence           `json:"evidence"`     // 
	Explanation  string                `json:"explanation"`  // 
	Metadata     map[string]interface{} `json:"metadata"`     // ?
}

// ConclusionType 
type ConclusionType string

const (
	ConclusionTypeRelation     ConclusionType = "relation"     // 
	ConclusionTypeProperty     ConclusionType = "property"     // ?
	ConclusionTypeClassification ConclusionType = "classification" // 
	ConclusionTypeContradiction ConclusionType = "contradiction" // 
)

// InferredRelation 
type InferredRelation struct {
	FromNodeID   uuid.UUID             `json:"from_node_id"`   // ID
	ToNodeID     uuid.UUID             `json:"to_node_id"`     // ID
	RelationType entities.RelationType `json:"relation_type"`  // 
	Weight       float64               `json:"weight"`         // 
	Certainty    float64               `json:"certainty"`      // ?
	Confidence   float64               `json:"confidence"`     // ?
	Evidence     []string              `json:"evidence"`       // 
	Reasoning    []string              `json:"reasoning"`      // 
	Temporal     *TemporalInfo         `json:"temporal"`       // 
}

// TemporalInfo 
type TemporalInfo struct {
	StartTime *time.Time `json:"start_time"` // ?
	EndTime   *time.Time `json:"end_time"`   // 
	Duration  *int64     `json:"duration"`   // 
	Frequency string     `json:"frequency"`  // 
}

// EvidenceType 
type InferenceEvidenceType string

const (
	InferenceEvidenceTypeEmpirical   InferenceEvidenceType = "empirical"   // 
	InferenceEvidenceTypeStatistical InferenceEvidenceType = "statistical" // 
	InferenceEvidenceTypeLogical     InferenceEvidenceType = "logical"     // 
	InferenceEvidenceTypeExpert      InferenceEvidenceType = "expert"      // 
	InferenceEvidenceTypeContextual  InferenceEvidenceType = "contextual"  // ?
)

// InferenceStep 
type InferenceStep struct {
	ID          string                 `json:"id"`          // ID
	RuleID      string                 `json:"rule_id"`     // ID
	Input       interface{}            `json:"input"`       // 
	Output      interface{}            `json:"output"`      // 
	Confidence  float64               `json:"confidence"`  // ?
	Explanation string                `json:"explanation"` // 
	Timestamp   time.Time             `json:"timestamp"`   // ?
	Metadata    map[string]interface{} `json:"metadata"`    // ?
}

// CachedInferenceResult 
type RelationCachedInferenceResult struct {
	Input       string                 `json:"input"`       // 
	Result      *InferenceResult       `json:"result"`      // 
	Confidence  float64               `json:"confidence"`  // ?
	Timestamp   time.Time             `json:"timestamp"`   // ?
	AccessCount int                   `json:"access_count"` // 
}

// InferenceRequest 
type InferenceRequest struct {
	Nodes           []*entities.KnowledgeNode     `json:"nodes"`           // 
	ExistingRelations []*entities.KnowledgeRelation `json:"existing_relations"` // 
	Context         *InferenceContext             `json:"context"`         // ?
	Options         *InferenceOptions             `json:"options"`         // 
	TargetRelations []entities.RelationType       `json:"target_relations"` // 
}

// InferenceContext ?
type InferenceContext struct {
	Domain          string                 `json:"domain"`          // 
	Subject         string                 `json:"subject"`         // 
	LearnerProfile  *LearnerProfile        `json:"learner_profile"` // ?
	TemporalContext *TemporalContext       `json:"temporal_context"` // ?
	SpatialContext  *SpatialContext        `json:"spatial_context"` // ?
	Metadata        map[string]interface{} `json:"metadata"`        // ?
}

// LearnerProfile ?
type RelationLearnerProfile struct {
	LearnerID      uuid.UUID              `json:"learner_id"`      // ID
	LearningStyle  string                 `json:"learning_style"`  // 
	KnowledgeLevel string                 `json:"knowledge_level"` // 
	Preferences    map[string]interface{} `json:"preferences"`     // 
	History        []string               `json:"history"`         // 
}

// TemporalContext ?
type TemporalContext struct {
	CurrentTime time.Time `json:"current_time"` // 
	TimeWindow  int64     `json:"time_window"`  // 䴰
	Seasonality string    `json:"seasonality"`  // ?
}

// SpatialContext ?
type SpatialContext struct {
	Location    string                 `json:"location"`    // 
	Environment string                 `json:"environment"` // 
	Context     map[string]interface{} `json:"context"`     // ?
}

// InferenceOptions 
type InferenceOptions struct {
	MaxDepth            int     `json:"max_depth"`            // ?
	MinConfidence       float64 `json:"min_confidence"`       // 
	EnableExplanation   bool    `json:"enable_explanation"`   // 
	EnableUncertainty   bool    `json:"enable_uncertainty"`   // ?
	EnableContradiction bool    `json:"enable_contradiction"` // ?
	ParallelProcessing  bool    `json:"parallel_processing"`  // 
}

// InferenceResponse 
type InferenceResponse struct {
	InferredRelations []*InferredRelation    `json:"inferred_relations"` // 
	InferenceChain    []*InferenceStep       `json:"inference_chain"`    // ?
	Contradictions    []*Contradiction       `json:"contradictions"`     // 
	Uncertainties     []*Uncertainty         `json:"uncertainties"`      // ?
	QualityMetrics    *InferenceQualityMetrics `json:"quality_metrics"`    // 
	ProcessingTime    int64                  `json:"processing_time"`    // 
	Explanations      []*Explanation         `json:"explanations"`       // 
}

// Contradiction 
type Contradiction struct {
	ID          string                 `json:"id"`          // ID
	Type        ContradictionType      `json:"type"`        // 
	Relations   []*InferredRelation    `json:"relations"`   // 
	Severity    float64               `json:"severity"`    // 
	Resolution  *ResolutionSuggestion  `json:"resolution"`  // 
	Evidence    []*Evidence           `json:"evidence"`    // 
	Explanation string                `json:"explanation"` // 
}

// ContradictionType 
type ContradictionType string

const (
	ContradictionTypeLogical   ContradictionType = "logical"   // 
	ContradictionTypeTemporal  ContradictionType = "temporal"  // 
	ContradictionTypeOntological ContradictionType = "ontological" // 
)

// ResolutionSuggestion 
type ResolutionSuggestion struct {
	Type        InferenceResolutionType `json:"type"`        // 
	Action      string                 `json:"action"`      // 
	Priority    int                   `json:"priority"`    // ?
	Confidence  float64               `json:"confidence"`  // ?
	Explanation string                `json:"explanation"` // 
	Metadata    map[string]interface{} `json:"metadata"`    // ?
}

// InferenceResolutionType 
type InferenceResolutionType string

const (
	InferenceResolutionTypeRemove   InferenceResolutionType = "remove"   // 
	InferenceResolutionTypeModify   InferenceResolutionType = "modify"   // 
	InferenceResolutionTypeReweight InferenceResolutionType = "reweight" // 
	InferenceResolutionTypeIgnore   InferenceResolutionType = "ignore"   // 
)

// Uncertainty ?
type Uncertainty struct {
	ID          string                 `json:"id"`          // ID
	Type        UncertaintyType        `json:"type"`        // ?
	Source      string                 `json:"source"`      // 
	Level       float64               `json:"level"`       // ?
	Impact      float64               `json:"impact"`      // 
	Mitigation  *MitigationStrategy    `json:"mitigation"`  // 
	Explanation string                `json:"explanation"` // 
}

// UncertaintyType ?
type UncertaintyType string

const (
	UncertaintyTypeEpistemic UncertaintyType = "epistemic" // ?
	UncertaintyTypeAleatory  UncertaintyType = "aleatory"  // ?
	UncertaintyTypeModel     UncertaintyType = "model"     // ?
)

// MitigationStrategy 
type MitigationStrategy struct {
	Type        MitigationType         `json:"type"`        // 
	Action      string                 `json:"action"`      // 
	Confidence  float64               `json:"confidence"`  // ?
	Cost        float64               `json:"cost"`        // 
	Benefit     float64               `json:"benefit"`     // 
	Explanation string                `json:"explanation"` // 
}

// MitigationType 
type MitigationType string

const (
	MitigationTypeDataCollection MitigationType = "data_collection" // 
	MitigationTypeModelImprovement MitigationType = "model_improvement" // 
	MitigationTypeExpertConsultation MitigationType = "expert_consultation" // 
)

// InferenceQualityMetrics 
type InferenceQualityMetrics struct {
	Precision    float64 `json:"precision"`    // ?
	Recall       float64 `json:"recall"`       // ?
	F1Score      float64 `json:"f1_score"`     // F1
	Consistency  float64 `json:"consistency"`  // ?
	Completeness float64 `json:"completeness"` // ?
	Novelty      float64 `json:"novelty"`      // ?
	Utility      float64 `json:"utility"`      // ?
}

// Explanation 
type Explanation struct {
	ID          string                 `json:"id"`          // ID
	Type        ExplanationType        `json:"type"`        // 
	Content     string                 `json:"content"`     // 
	Confidence  float64               `json:"confidence"`  // ?
	Evidence    []*Evidence           `json:"evidence"`    // 
	Reasoning   string                `json:"reasoning"`   // 
	Metadata    map[string]interface{} `json:"metadata"`    // ?
}

// ExplanationType 
type ExplanationType string

const (
	ExplanationTypeDeductive   ExplanationType = "deductive"   // 
	ExplanationTypeInductive   ExplanationType = "inductive"   // 
	ExplanationTypeAbductive   ExplanationType = "abductive"   // 
	ExplanationTypeContrastive ExplanationType = "contrastive" // 
	ExplanationTypeCounterfactual ExplanationType = "counterfactual" // ?
)

// NewIntelligentRelationInferenceEngine 
func NewIntelligentRelationInferenceEngine(crossModalService CrossModalServiceInterface) *IntelligentRelationInferenceEngine {
	config := &InferenceEngineConfig{
		MinConfidenceThreshold:     0.7,
		MaxInferenceDepth:         5,
		EnableTransitiveInference:  true,
		EnableContradictionCheck:   true,
		EnableUncertaintyReasoning: true,
		ContextWindowSize:          100,
		ParallelProcessing:         true,
		CacheEnabled:              true,
	}

	cache := &InferenceCache{
		RelationProbabilities: make(map[string]float64),
		InferenceChains:      make(map[string][]*InferenceStep),
		ContextEmbeddings:    make(map[string][]float64),
		CachedResults:        make(map[string]*domainServices.CachedInferenceResult),
		LastUpdated:          time.Now(),
	}

	metrics := &InferenceMetrics{
		LastInferenceTime: time.Now(),
	}

	engine := &IntelligentRelationInferenceEngine{
		crossModalService: crossModalService,
		config:           config,
		cache:            cache,
		metrics:          metrics,
		rules:            make([]*AdvancedInferenceRule, 0),
	}

	// ?
	engine.initializeDefaultRules()

	return engine
}

// ProcessInference 
func (e *IntelligentRelationInferenceEngine) ProcessInference(ctx context.Context, req *InferenceRequest) (*InferenceResponse, error) {
	startTime := time.Now()
	
	response := &InferenceResponse{
		InferredRelations: make([]*InferredRelation, 0),
		InferenceChain:    make([]*InferenceStep, 0),
		Contradictions:    make([]*Contradiction, 0),
		Uncertainties:     make([]*Uncertainty, 0),
		Explanations:      make([]*Explanation, 0),
	}

	// 1. 
	if err := e.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. ?
	if err := e.generateContextEmbeddings(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to generate context embeddings: %w", err)
	}

	// 3. 
	inferredRelations, inferenceChain := e.applyInferenceRules(ctx, req)
	response.InferredRelations = inferredRelations
	response.InferenceChain = inferenceChain

	// 4. ?
	if e.config.EnableContradictionCheck {
		contradictions := e.detectContradictions(inferredRelations, req.ExistingRelations)
		response.Contradictions = contradictions
	}

	// 5. ?
	if e.config.EnableUncertaintyReasoning {
		uncertainties := e.analyzeUncertainties(inferredRelations, inferenceChain)
		response.Uncertainties = uncertainties
	}

	// 6. 
	explanations := e.generateExplanations(inferredRelations, inferenceChain)
	response.Explanations = explanations

	// 7. 
	response.QualityMetrics = e.calculateQualityMetrics(inferredRelations, req.ExistingRelations)

	// 8. ?
	e.updateMetrics(len(inferredRelations), time.Since(startTime))
	e.updateCache(req, response)

	response.ProcessingTime = time.Since(startTime).Milliseconds()
	return response, nil
}

// initializeDefaultRules ?
func (e *IntelligentRelationInferenceEngine) initializeDefaultRules() {
	// ?
	transitiveRule := &AdvancedInferenceRule{
		ID:          "transitive_prerequisite",
		Name:        "Transitive Prerequisite Rule",
		Description: "If A is prerequisite to B and B is prerequisite to C, then A is prerequisite to C",
		Type:        RuleTypeTransitive,
		Priority:    10,
		Confidence:  0.9,
		Enabled:     true,
	}

	// 
	hierarchyRule := &AdvancedInferenceRule{
		ID:          "hierarchy_part_of",
		Name:        "Hierarchy Part-Of Rule",
		Description: "If A is part of B and B is part of C, then A is part of C",
		Type:        RuleTypeTransitive,
		Priority:    9,
		Confidence:  0.85,
		Enabled:     true,
	}

	// ?
	similarityRule := &AdvancedInferenceRule{
		ID:          "similarity_related",
		Name:        "Similarity Related Rule",
		Description: "If A and B have high semantic similarity, they are likely related",
		Type:        RuleTypeInductive,
		Priority:    7,
		Confidence:  0.75,
		Enabled:     true,
	}

	// 
	causalRule := &AdvancedInferenceRule{
		ID:          "causal_leads_to",
		Name:        "Causal Leads-To Rule",
		Description: "If A causally influences B, then A leads to B",
		Type:        RuleTypeCausal,
		Priority:    8,
		Confidence:  0.8,
		Enabled:     true,
	}

	e.rules = append(e.rules, transitiveRule, hierarchyRule, similarityRule, causalRule)
}

// validateRequest 
func (e *IntelligentRelationInferenceEngine) validateRequest(req *InferenceRequest) error {
	if len(req.Nodes) == 0 {
		return fmt.Errorf("no nodes provided")
	}

	if req.Options != nil {
		if req.Options.MaxDepth > e.config.MaxInferenceDepth {
			return fmt.Errorf("max depth exceeds limit")
		}
		if req.Options.MinConfidence < 0 || req.Options.MinConfidence > 1 {
			return fmt.Errorf("invalid confidence threshold")
		}
	}

	return nil
}

// generateContextEmbeddings ?
func (e *IntelligentRelationInferenceEngine) generateContextEmbeddings(ctx context.Context, req *InferenceRequest) error {
	if req.Context == nil {
		return nil
	}

	// ?
	contextText := fmt.Sprintf("Domain: %s, Subject: %s", req.Context.Domain, req.Context.Subject)
	
	// AI
	embeddingReq := &CrossModalInferenceRequest{
		Type: "text_embedding",
		Data: map[string]interface{}{
			"text": contextText,
		},
	}

	embeddingResp, err := e.crossModalService.ProcessCrossModalInference(ctx, embeddingReq)
	if err != nil {
		return err
	}

	if embedding, ok := embeddingResp.Result["embedding"].([]interface{}); ok {
		vector := make([]float64, len(embedding))
		for i, val := range embedding {
			if floatVal, ok := val.(float64); ok {
				vector[i] = floatVal
			}
		}
		e.cache.ContextEmbeddings[req.Context.Domain] = vector
	}

	return nil
}

// applyInferenceRules 
func (e *IntelligentRelationInferenceEngine) applyInferenceRules(ctx context.Context, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	inferredRelations := make([]*InferredRelation, 0)
	inferenceChain := make([]*InferenceStep, 0)

	// 
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].Priority > e.rules[j].Priority
	})

	// 
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		relations, steps := e.applyRule(ctx, rule, req)
		inferredRelations = append(inferredRelations, relations...)
		inferenceChain = append(inferenceChain, steps...)
	}

	// ?
	inferredRelations = e.deduplicateRelations(inferredRelations)
	inferredRelations = e.filterByConfidence(inferredRelations, e.config.MinConfidenceThreshold)

	return inferredRelations, inferenceChain
}

// applyRule 
func (e *IntelligentRelationInferenceEngine) applyRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	switch rule.Type {
	case RuleTypeTransitive:
		r, s := e.applyTransitiveRule(rule, req)
		relations = append(relations, r...)
		steps = append(steps, s...)
	case RuleTypeInductive:
		r, s := e.applyInductiveRule(ctx, rule, req)
		relations = append(relations, r...)
		steps = append(steps, s...)
	case RuleTypeCausal:
		r, s := e.applyCausalRule(ctx, rule, req)
		relations = append(relations, r...)
		steps = append(steps, s...)
	case RuleTypeAnalogical:
		r, s := e.applyAnalogicalRule(ctx, rule, req)
		relations = append(relations, r...)
		steps = append(steps, s...)
	}

	return relations, steps
}

// applyTransitiveRule ?
func (e *IntelligentRelationInferenceEngine) applyTransitiveRule(rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// ?
	relationMap := make(map[uuid.UUID]map[uuid.UUID]*entities.KnowledgeRelation)
	for _, rel := range req.ExistingRelations {
		if relationMap[rel.FromNodeID] == nil {
			relationMap[rel.FromNodeID] = make(map[uuid.UUID]*entities.KnowledgeRelation)
		}
		relationMap[rel.FromNodeID][rel.ToNodeID] = rel
	}

	// ?
	for _, nodeA := range req.Nodes {
		for _, nodeB := range req.Nodes {
			if nodeA.ID == nodeB.ID {
				continue
			}

			// A->B?
			if relationMap[nodeA.ID] != nil {
				if relAB, exists := relationMap[nodeA.ID][nodeB.ID]; exists {
					// B->C?
					if relationMap[nodeB.ID] != nil {
						for nodeC_ID, relBC := range relationMap[nodeB.ID] {
							if nodeC_ID == nodeA.ID {
								continue
							}

							// A->C?
							if relationMap[nodeA.ID][nodeC_ID] == nil {
								// A->C
								confidence := math.Min(relAB.Confidence, relBC.Confidence) * rule.Confidence
								if confidence > e.config.MinConfidenceThreshold {
									inferredRel := &InferredRelation{
										FromNodeID:   nodeA.ID,
										ToNodeID:     nodeC_ID,
										RelationType: relAB.Type, // ?
										Weight:       confidence,
										Certainty:    confidence,
									}

									relations = append(relations, inferredRel)

									step := &InferenceStep{
										ID:          uuid.New().String(),
										RuleID:      rule.ID,
										Input:       fmt.Sprintf("Relations: %s->%s, %s->%s", nodeA.ID, nodeB.ID, nodeB.ID, nodeC_ID),
										Output:      fmt.Sprintf("Inferred: %s->%s", nodeA.ID, nodeC_ID),
										Confidence:  confidence,
										Explanation: fmt.Sprintf("Transitive inference using rule %s", rule.Name),
										Timestamp:   time.Now(),
									}

									steps = append(steps, step)
								}
							}
						}
					}
				}
			}
		}
	}

	return relations, steps
}

// applyInductiveRule 
func (e *IntelligentRelationInferenceEngine) applyInductiveRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// 
	for i, nodeA := range req.Nodes {
		for j, nodeB := range req.Nodes {
			if i >= j {
				continue
			}

			// ?
			similarity := e.calculateNodeSimilarity(ctx, nodeA, nodeB)
			if similarity > 0.7 { // ?
				confidence := similarity * rule.Confidence
				if confidence > e.config.MinConfidenceThreshold {
					inferredRel := &InferredRelation{
						FromNodeID:   nodeA.ID,
						ToNodeID:     nodeB.ID,
						RelationType: entities.RelationTypeRelatedTo,
						Weight:       confidence,
						Certainty:    confidence,
					}

					relations = append(relations, inferredRel)

					step := &InferenceStep{
						ID:          uuid.New().String(),
						RuleID:      rule.ID,
						Input:       fmt.Sprintf("Nodes: %s, %s", nodeA.Name, nodeB.Name),
						Output:      fmt.Sprintf("Inferred relation: %s related_to %s", nodeA.Name, nodeB.Name),
						Confidence:  confidence,
						Explanation: fmt.Sprintf("Inductive inference based on similarity: %.2f", similarity),
						Timestamp:   time.Now(),
					}

					steps = append(steps, step)
				}
			}
		}
	}

	return relations, steps
}

// applyCausalRule 
func (e *IntelligentRelationInferenceEngine) applyCausalRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// 
	for _, nodeA := range req.Nodes {
		for _, nodeB := range req.Nodes {
			if nodeA.ID == nodeB.ID {
				continue
			}

			// 
			if e.hasCausalRelationship(nodeA, nodeB) {
				confidence := e.calculateCausalConfidence(nodeA, nodeB) * rule.Confidence
				if confidence > e.config.MinConfidenceThreshold {
					inferredRel := &InferredRelation{
						FromNodeID:   nodeA.ID,
						ToNodeID:     nodeB.ID,
						RelationType: entities.RelationTypeLeadsTo,
						Weight:       confidence,
						Certainty:    confidence,
					}

					relations = append(relations, inferredRel)

					step := &InferenceStep{
						ID:          uuid.New().String(),
						RuleID:      rule.ID,
						Input:       fmt.Sprintf("Nodes: %s, %s", nodeA.Name, nodeB.Name),
						Output:      fmt.Sprintf("Inferred causal relation: %s leads_to %s", nodeA.Name, nodeB.Name),
						Confidence:  confidence,
						Explanation: "Causal inference based on difficulty progression and domain knowledge",
						Timestamp:   time.Now(),
					}

					steps = append(steps, step)
				}
			}
		}
	}

	return relations, steps
}

// applyAnalogicalRule 
func (e *IntelligentRelationInferenceEngine) applyAnalogicalRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// ABBCAC
	for _, nodeA := range req.Nodes {
		for _, nodeB := range req.Nodes {
			if nodeA.ID == nodeB.ID {
				continue
			}

			similarity := e.calculateNodeSimilarity(ctx, nodeA, nodeB)
			if similarity > 0.8 { // ?
				// BA?
				for _, rel := range req.ExistingRelations {
					if rel.FromNodeID == nodeB.ID {
						confidence := similarity * rel.Confidence * rule.Confidence
						if confidence > e.config.MinConfidenceThreshold {
							inferredRel := &InferredRelation{
								FromNodeID:   nodeA.ID,
								ToNodeID:     rel.ToNodeID,
								RelationType: rel.Type,
								Weight:       confidence,
								Certainty:    confidence,
							}

							relations = append(relations, inferredRel)

							step := &InferenceStep{
								ID:          uuid.New().String(),
								RuleID:      rule.ID,
								Input:       fmt.Sprintf("Analogy: %s similar to %s, %s has relation to %s", nodeA.Name, nodeB.Name, nodeB.Name, rel.ToNodeID),
								Output:      fmt.Sprintf("Inferred analogical relation: %s %s %s", nodeA.Name, rel.Type, rel.ToNodeID),
								Confidence:  confidence,
								Explanation: fmt.Sprintf("Analogical inference based on similarity: %.2f", similarity),
								Timestamp:   time.Now(),
							}

							steps = append(steps, step)
						}
					}
				}
			}
		}
	}

	return relations, steps
}

// calculateNodeSimilarity ?
func (e *IntelligentRelationInferenceEngine) calculateNodeSimilarity(ctx context.Context, nodeA, nodeB *entities.KnowledgeNode) float64 {
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
	tagSim := e.calculateTagSimilarity(nodeA.Tags, nodeB.Tags)
	similarity += 0.2 * tagSim

	return similarity
}

// calculateTagSimilarity ?
func (e *IntelligentRelationInferenceEngine) calculateTagSimilarity(tagsA, tagsB []string) float64 {
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

// hasCausalRelationship ?
func (e *IntelligentRelationInferenceEngine) hasCausalRelationship(nodeA, nodeB *entities.KnowledgeNode) bool {
	// ?
	if int(nodeA.DifficultyLevel) < int(nodeB.DifficultyLevel) && nodeA.Subject == nodeB.Subject {
		return true
	}

	// ?
	if nodeA.Type == entities.NodeTypeSkill && nodeB.Type == entities.NodeTypeSkill {
		return true
	}

	return false
}

// calculateCausalConfidence ?
func (e *IntelligentRelationInferenceEngine) calculateCausalConfidence(nodeA, nodeB *entities.KnowledgeNode) float64 {
	confidence := 0.5

	// ?
	diffA := int(nodeA.DifficultyLevel)
	diffB := int(nodeB.DifficultyLevel)
	if diffB > diffA {
		diffFactor := 1.0 - float64(diffB-diffA)/4.0
		confidence += 0.3 * diffFactor
	}

	// ?
	if nodeA.Subject == nodeB.Subject {
		confidence += 0.2
	}

	return math.Min(confidence, 1.0)
}

// deduplicateRelations 
func (e *IntelligentRelationInferenceEngine) deduplicateRelations(relations []*InferredRelation) []*InferredRelation {
	seen := make(map[string]*InferredRelation)
	result := make([]*InferredRelation, 0)

	for _, rel := range relations {
		key := fmt.Sprintf("%s-%s-%s", rel.FromNodeID, rel.ToNodeID, rel.RelationType)
		if existing, exists := seen[key]; exists {
			// 
			if rel.Confidence > existing.Confidence {
				seen[key] = rel
			}
		} else {
			seen[key] = rel
		}
	}

	for _, rel := range seen {
		result = append(result, rel)
	}

	return result
}

// filterByConfidence 
func (e *IntelligentRelationInferenceEngine) filterByConfidence(relations []*InferredRelation, threshold float64) []*InferredRelation {
	result := make([]*InferredRelation, 0)
	for _, rel := range relations {
		if rel.Confidence >= threshold {
			result = append(result, rel)
		}
	}
	return result
}

// detectContradictions ?
func (e *IntelligentRelationInferenceEngine) detectContradictions(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) []*Contradiction {
	contradictions := make([]*Contradiction, 0)

	// ?
	for _, inferred := range inferredRelations {
		for _, existing := range existingRelations {
			if e.areContradictory(inferred, existing) {
				contradiction := &Contradiction{
					ID:       uuid.New().String(),
					Type:     ContradictionTypeLogical,
					Severity: e.calculateContradictionSeverity(inferred, existing),
					Relations: []*InferredRelation{inferred},
					Explanation: fmt.Sprintf("Inferred relation %s conflicts with existing relation", inferred.RelationType),
				}
				contradictions = append(contradictions, contradiction)
			}
		}
	}

	return contradictions
}

// areContradictory ?
func (e *IntelligentRelationInferenceEngine) areContradictory(inferred *InferredRelation, existing *entities.KnowledgeRelation) bool {
	// 
	if inferred.FromNodeID == existing.FromNodeID && inferred.ToNodeID == existing.ToNodeID {
		return e.areRelationTypesContradictory(inferred.RelationType, existing.Type)
	}

	// 
	if inferred.FromNodeID == existing.ToNodeID && inferred.ToNodeID == existing.FromNodeID {
		return e.areReverseRelationsContradictory(inferred.RelationType, existing.Type)
	}

	return false
}

// areRelationTypesContradictory ?
func (e *IntelligentRelationInferenceEngine) areRelationTypesContradictory(type1, type2 entities.RelationType) bool {
	// 
	contradictoryPairs := map[entities.RelationType][]entities.RelationType{
		entities.RelationTypePrerequisite: {entities.RelationTypeOppositeOf},
		entities.RelationTypeOppositeOf:   {entities.RelationTypePrerequisite, entities.RelationTypeSimilarTo},
		entities.RelationTypeSimilarTo:    {entities.RelationTypeOppositeOf},
	}

	if contradictory, exists := contradictoryPairs[type1]; exists {
		for _, contradictoryType := range contradictory {
			if type2 == contradictoryType {
				return true
			}
		}
	}

	return false
}

// areReverseRelationsContradictory 鷴?
func (e *IntelligentRelationInferenceEngine) areReverseRelationsContradictory(type1, type2 entities.RelationType) bool {
	// 
	asymmetricRelations := []entities.RelationType{
		entities.RelationTypePrerequisite,
		entities.RelationTypeLeadsTo,
		entities.RelationTypePartOf,
	}

	for _, asymmetric := range asymmetricRelations {
		if type1 == asymmetric && type2 == asymmetric {
			return true
		}
	}

	return false
}

// calculateContradictionSeverity 
func (e *IntelligentRelationInferenceEngine) calculateContradictionSeverity(inferred *InferredRelation, existing *entities.KnowledgeRelation) float64 {
	// ?
	confidenceDiff := math.Abs(inferred.Confidence - existing.Confidence)
	severity := 0.5 + 0.5*confidenceDiff

	// 
	if inferred.RelationType == entities.RelationTypeOppositeOf || existing.Type == entities.RelationTypeOppositeOf {
		severity *= 1.5
	}

	return math.Min(severity, 1.0)
}

// analyzeUncertainties ?
func (e *IntelligentRelationInferenceEngine) analyzeUncertainties(inferredRelations []*InferredRelation, inferenceChain []*InferenceStep) []*Uncertainty {
	uncertainties := make([]*Uncertainty, 0)

	// ?
	for _, rel := range inferredRelations {
		if rel.Confidence < 0.8 {
			uncertainty := &Uncertainty{
				ID:     uuid.New().String(),
				Type:   UncertaintyTypeEpistemic,
				Source: "low_confidence_inference",
				Level:  1.0 - rel.Confidence,
				Impact: e.calculateUncertaintyImpact(rel),
				Explanation: fmt.Sprintf("Low confidence (%.2f) in inferred relation %s", rel.Confidence, rel.RelationType),
			}
			uncertainties = append(uncertainties, uncertainty)
		}
	}

	return uncertainties
}

// calculateUncertaintyImpact 㲻?
func (e *IntelligentRelationInferenceEngine) calculateUncertaintyImpact(rel *InferredRelation) float64 {
	// ?
	impact := rel.Weight

	// ?
	if rel.RelationType == entities.RelationTypePrerequisite {
		impact *= 1.2
	}

	return math.Min(impact, 1.0)
}

// generateExplanations 
func (e *IntelligentRelationInferenceEngine) generateExplanations(inferredRelations []*InferredRelation, inferenceChain []*InferenceStep) []*Explanation {
	explanations := make([]*Explanation, 0)

	// ?
	for _, step := range inferenceChain {
		explanation := &Explanation{
			ID:         uuid.New().String(),
			Type:       ExplanationTypeDeductive,
			Content:    step.Explanation,
			Confidence: step.Confidence,
			Reasoning:  fmt.Sprintf("Applied rule %s: %s", step.RuleID, step.Explanation),
			Metadata: map[string]interface{}{
				"step_id": step.ID,
				"rule_id": step.RuleID,
			},
		}
		explanations = append(explanations, explanation)
	}

	return explanations
}

// calculateQualityMetrics 
func (e *IntelligentRelationInferenceEngine) calculateQualityMetrics(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) *InferenceQualityMetrics {
	metrics := &InferenceQualityMetrics{}

	if len(inferredRelations) == 0 {
		return metrics
	}

	// ?
	totalConfidence := 0.0
	for _, rel := range inferredRelations {
		totalConfidence += rel.Confidence
	}
	metrics.Precision = totalConfidence / float64(len(inferredRelations))

	// 汾
	totalPossibleRelations := len(existingRelations) + len(inferredRelations)
	if totalPossibleRelations > 0 {
		metrics.Recall = float64(len(inferredRelations)) / float64(totalPossibleRelations)
	}

	// F1
	if metrics.Precision+metrics.Recall > 0 {
		metrics.F1Score = 2 * (metrics.Precision * metrics.Recall) / (metrics.Precision + metrics.Recall)
	}

	// ?
	metrics.Consistency = e.calculateConsistencyScore(inferredRelations)

	// ?
	metrics.Completeness = e.calculateCompletenessScore(inferredRelations, existingRelations)

	// ?
	metrics.Novelty = e.calculateNoveltyScore(inferredRelations, existingRelations)

	// ?
	metrics.Utility = e.calculateUtilityScore(inferredRelations)

	return metrics
}

// calculateConsistencyScore ?
func (e *IntelligentRelationInferenceEngine) calculateConsistencyScore(relations []*InferredRelation) float64 {
	if len(relations) <= 1 {
		return 1.0
	}

	consistentPairs := 0
	totalPairs := 0

	for i, relA := range relations {
		for j, relB := range relations {
			if i >= j {
				continue
			}
			totalPairs++
			if !e.areInferredRelationsContradictory(relA, relB) {
				consistentPairs++
			}
		}
	}

	if totalPairs == 0 {
		return 1.0
	}

	return float64(consistentPairs) / float64(totalPairs)
}

// areInferredRelationsContradictory ?
func (e *IntelligentRelationInferenceEngine) areInferredRelationsContradictory(relA, relB *InferredRelation) bool {
	// 
	if relA.FromNodeID == relB.FromNodeID && relA.ToNodeID == relB.ToNodeID {
		return e.areRelationTypesContradictory(relA.RelationType, relB.RelationType)
	}

	// 
	if relA.FromNodeID == relB.ToNodeID && relA.ToNodeID == relB.FromNodeID {
		return e.areReverseRelationsContradictory(relA.RelationType, relB.RelationType)
	}

	return false
}

// calculateCompletenessScore ?
func (e *IntelligentRelationInferenceEngine) calculateCompletenessScore(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) float64 {
	// 
	if len(existingRelations) == 0 {
		return 1.0
	}

	ratio := float64(len(inferredRelations)) / float64(len(existingRelations))
	return math.Min(ratio, 1.0)
}

// calculateNoveltyScore ?
func (e *IntelligentRelationInferenceEngine) calculateNoveltyScore(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) float64 {
	if len(inferredRelations) == 0 {
		return 0.0
	}

	novelRelations := 0
	existingRelationSet := make(map[string]bool)

	// 
	for _, rel := range existingRelations {
		key := fmt.Sprintf("%s-%s-%s", rel.FromNodeID, rel.ToNodeID, rel.Type)
		existingRelationSet[key] = true
	}

	// ?
	for _, rel := range inferredRelations {
		key := fmt.Sprintf("%s-%s-%s", rel.FromNodeID, rel.ToNodeID, rel.RelationType)
		if !existingRelationSet[key] {
			novelRelations++
		}
	}

	return float64(novelRelations) / float64(len(inferredRelations))
}

// calculateUtilityScore ?
func (e *IntelligentRelationInferenceEngine) calculateUtilityScore(relations []*InferredRelation) float64 {
	if len(relations) == 0 {
		return 0.0
	}

	totalUtility := 0.0
	for _, rel := range relations {
		// ?
		utility := rel.Confidence

		// ?
		switch rel.RelationType {
		case entities.RelationTypePrerequisite:
			utility *= 1.2
		case entities.RelationTypeLeadsTo:
			utility *= 1.1
		case entities.RelationTypePartOf:
			utility *= 1.1
		}

		totalUtility += utility
	}

	return totalUtility / float64(len(relations))
}

// updateMetrics 
func (e *IntelligentRelationInferenceEngine) updateMetrics(inferenceCount int, processingTime time.Duration) {
	e.metrics.TotalInferences += int64(inferenceCount)
	e.metrics.SuccessfulInferences += int64(inferenceCount)
	e.metrics.LastInferenceTime = time.Now()

	// 
	if e.metrics.TotalInferences > 0 {
		totalTime := e.metrics.AverageProcessingTime*int64(e.metrics.TotalInferences-int64(inferenceCount)) + processingTime.Milliseconds()
		e.metrics.AverageProcessingTime = totalTime / e.metrics.TotalInferences
	} else {
		e.metrics.AverageProcessingTime = processingTime.Milliseconds()
	}
}

// updateCache 
func (e *IntelligentRelationInferenceEngine) updateCache(req *InferenceRequest, resp *InferenceResponse) {
	if !e.config.CacheEnabled {
		return
	}

	// 
	inputHash := e.generateInputHash(req)
	cachedResult := &domainServices.CachedInferenceResult{
		QueryID:    inputHash,
		Result:     resp, // ?
		Confidence: e.calculateAverageConfidence(resp.InferredRelations),
		CachedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24?
		Metadata:   make(map[string]interface{}),
	}

	e.cache.CachedResults[inputHash] = cachedResult
	e.cache.LastUpdated = time.Now()
}

// generateInputHash 
func (e *IntelligentRelationInferenceEngine) generateInputHash(req *InferenceRequest) string {
	// 
	nodeIDs := make([]string, len(req.Nodes))
	for i, node := range req.Nodes {
		nodeIDs[i] = node.ID.String()
	}
	sort.Strings(nodeIDs)
	return fmt.Sprintf("%x", nodeIDs)
}

// calculateAverageConfidence ?
func (e *IntelligentRelationInferenceEngine) calculateAverageConfidence(relations []*InferredRelation) float64 {
	if len(relations) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, rel := range relations {
		totalConfidence += rel.Confidence
	}

	return totalConfidence / float64(len(relations))
}

// GetMetrics 
func (e *IntelligentRelationInferenceEngine) GetMetrics() *InferenceMetrics {
	return e.metrics
}

// UpdateConfig 
func (e *IntelligentRelationInferenceEngine) UpdateConfig(config *InferenceEngineConfig) {
	e.config = config
}

// AddRule 
func (e *IntelligentRelationInferenceEngine) AddRule(rule *AdvancedInferenceRule) {
	e.rules = append(e.rules, rule)
}

// RemoveRule 
func (e *IntelligentRelationInferenceEngine) RemoveRule(ruleID string) {
	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			break
		}
	}
}

// EnableRule 
func (e *IntelligentRelationInferenceEngine) EnableRule(ruleID string) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			rule.Enabled = true
			break
		}
	}
}

// DisableRule 
func (e *IntelligentRelationInferenceEngine) DisableRule(ruleID string) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			rule.Enabled = false
			break
		}
	}
}

// ClearCache 
func (e *IntelligentRelationInferenceEngine) ClearCache() {
	e.cache.RelationProbabilities = make(map[string]float64)
	e.cache.InferenceChains = make(map[string][]*InferenceStep)
	e.cache.ContextEmbeddings = make(map[string][]float64)
	e.cache.CachedResults = make(map[string]*domainServices.CachedInferenceResult)
	e.cache.LastUpdated = time.Now()
}

