package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	domainServices "github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// EvidenceType 证据类型
type EvidenceType string

const (
	EvidenceTypeStatistical EvidenceType = "statistical"
	EvidenceTypeObservation EvidenceType = "observation"
	EvidenceTypeExperimental EvidenceType = "experimental"
	EvidenceTypeCorrelation EvidenceType = "correlation"
)

// IntelligentRelationInferenceEngine 智能关系推理引擎
type IntelligentRelationInferenceEngine struct {
	crossModalService CrossModalServiceInterface
	config           *InferenceEngineConfig
	cache            *InferenceCache
	metrics          *InferenceMetrics
	rules            []*AdvancedInferenceRule
}

// InferenceEngineConfig 推理引擎配置
type InferenceEngineConfig struct {
	MinConfidenceThreshold      float64 `json:"min_confidence_threshold"`      // 最小置信度阈值
	MaxInferenceDepth          int     `json:"max_inference_depth"`           // 最大推理深度
	EnableTransitiveInference  bool    `json:"enable_transitive_inference"`   // 启用传递推理
	EnableContradictionCheck   bool    `json:"enable_contradiction_check"`    // 启用矛盾检查
	EnableUncertaintyReasoning bool    `json:"enable_uncertainty_reasoning"`  // 启用不确定性推理
	ContextWindowSize          int     `json:"context_window_size"`           // 上下文窗口大小
	ParallelProcessing         bool    `json:"parallel_processing"`           // 并行处理
	CacheEnabled               bool    `json:"cache_enabled"`                 // 启用缓存
}

// InferenceCache 推理缓存
type InferenceCache struct {
	RelationProbabilities map[string]float64                    `json:"relation_probabilities"` // 关系概率
	InferenceChains      map[string][]*InferenceStep           `json:"inference_chains"`       // 推理链
	ContextEmbeddings    map[string][]float64                  `json:"context_embeddings"`     // 上下文嵌入
	CachedResults        map[string]*domainServices.CachedInferenceResult     `json:"cached_results"`         // 缓存结果
	LastUpdated          time.Time                             `json:"last_updated"`           // 最后更新时间
}

// InferenceMetrics 推理指标
type InferenceMetrics struct {
	TotalInferences       int64     `json:"total_inferences"`       // 总推理次数
	SuccessfulInferences  int64     `json:"successful_inferences"`  // 成功推理次数
	FailedInferences      int64     `json:"failed_inferences"`      // 失败推理次数
	AverageConfidence     float64   `json:"average_confidence"`     // 平均置信度
	AverageProcessingTime int64     `json:"average_processing_time"` // 平均处理时间
	ContradictionsFound   int64     `json:"contradictions_found"`   // 发现的矛盾数
	TransitiveInferences  int64     `json:"transitive_inferences"`  // 传递推理次数
	LastInferenceTime     time.Time `json:"last_inference_time"`    // 最后推理时间
}

// AdvancedInferenceRule 高级推理规则
type AdvancedInferenceRule struct {
	ID                string                    `json:"id"`                // 规则ID
	Name              string                    `json:"name"`              // 规则名称
	Description       string                    `json:"description"`       // 规则描述
	Type              InferenceRuleType         `json:"type"`              // 规则类型
	Conditions        []*ComplexCondition       `json:"conditions"`        // 复杂条件
	Conclusions       []*InferenceConclusion    `json:"conclusions"`       // 推理结论
	Priority          int                       `json:"priority"`          // 优先级
	Confidence        float64                   `json:"confidence"`        // 规则置信度
	Enabled           bool                      `json:"enabled"`           // 是否启用
	ContextRequired   bool                      `json:"context_required"`  // 是否需要上下文
	MetaReasoning     bool                      `json:"meta_reasoning"`    // 是否为元推理
}

// InferenceRuleType 推理规则类型
type InferenceRuleType string

const (
	RuleTypeDeductive    InferenceRuleType = "deductive"    // 演绎推理
	RuleTypeInductive    InferenceRuleType = "inductive"    // 归纳推理
	RuleTypeAbductive    InferenceRuleType = "abductive"    // 溯因推理
	RuleTypeAnalogical   InferenceRuleType = "analogical"   // 类比推理
	RuleTypeTransitive   InferenceRuleType = "transitive"   // 传递推理
	RuleTypeCausal       InferenceRuleType = "causal"       // 因果推理
	RuleTypeStatistical  InferenceRuleType = "statistical"  // 统计推理
	RuleTypeOntological  InferenceRuleType = "ontological"  // 本体推理
)

// ComplexCondition 复杂条件
type ComplexCondition struct {
	ID           string                 `json:"id"`           // 条件ID
	Type         ConditionType          `json:"type"`         // 条件类型
	Operator     LogicalOperator        `json:"operator"`     // 逻辑操作符
	Predicates   []*Predicate          `json:"predicates"`   // 谓词
	SubConditions []*ComplexCondition   `json:"sub_conditions"` // 子条件
	Weight       float64               `json:"weight"`       // 权重
	Negated      bool                  `json:"negated"`      // 是否否定
}

// ConditionType 条件类型
type ConditionType string

const (
	ConditionTypeAtomic    ConditionType = "atomic"    // 原子条件
	ConditionTypeComposite ConditionType = "composite" // 复合条件
	ConditionTypeModal     ConditionType = "modal"     // 模态条件
	ConditionTypeTemporal  ConditionType = "temporal"  // 时间条件
)

// LogicalOperator 逻辑操作符
type LogicalOperator string

const (
	OperatorAND LogicalOperator = "AND"
	OperatorOR  LogicalOperator = "OR"
	OperatorNOT LogicalOperator = "NOT"
	OperatorXOR LogicalOperator = "XOR"
)

// Predicate 谓词
type Predicate struct {
	Subject   *PredicateElement `json:"subject"`   // 主语
	Predicate string           `json:"predicate"` // 谓语
	Object    *PredicateElement `json:"object"`    // 宾语
	Modality  string           `json:"modality"`  // 模态
	Certainty float64          `json:"certainty"` // 确定性
}

// PredicateElement 谓词元素
type PredicateElement struct {
	Type       string      `json:"type"`       // 类型
	Value      interface{} `json:"value"`      // 值
	Variable   bool        `json:"variable"`   // 是否为变量
	Quantifier string      `json:"quantifier"` // 量词
}

// InferenceConclusion 推理结论
type InferenceConclusion struct {
	Type         ConclusionType         `json:"type"`         // 结论类型
	Relation     *InferredRelation      `json:"relation"`     // 推理关系
	Confidence   float64               `json:"confidence"`   // 置信度
	Evidence     []*domainServices.Evidence           `json:"evidence"`     // 证据
	Explanation  string                `json:"explanation"`  // 解释
	Metadata     map[string]interface{} `json:"metadata"`     // 元数据
}

// ConclusionType 结论类型
type ConclusionType string

const (
	ConclusionTypeRelation     ConclusionType = "relation"     // 关系结论
	ConclusionTypeProperty     ConclusionType = "property"     // 属性结论
	ConclusionTypeClassification ConclusionType = "classification" // 分类结论
	ConclusionTypeContradiction ConclusionType = "contradiction" // 矛盾结论
)

// InferredRelation 推理关系
type InferredRelation struct {
	FromNodeID   uuid.UUID             `json:"from_node_id"`   // 源节点ID
	ToNodeID     uuid.UUID             `json:"to_node_id"`     // 目标节点ID
	RelationType entities.RelationType `json:"relation_type"`  // 关系类型
	Weight       float64               `json:"weight"`         // 权重
	Certainty    float64               `json:"certainty"`      // 确定性
	Confidence   float64               `json:"confidence"`     // 置信度
	Evidence     []string              `json:"evidence"`       // 证据
	Reasoning    []string              `json:"reasoning"`      // 推理过程
	Temporal     *TemporalInfo         `json:"temporal"`       // 时间信息
}

// TemporalInfo 时间信息
type TemporalInfo struct {
	StartTime *time.Time `json:"start_time"` // 开始时间
	EndTime   *time.Time `json:"end_time"`   // 结束时间
	Duration  *int64     `json:"duration"`   // 持续时间
	Frequency string     `json:"frequency"`  // 频率
}

// Evidence 证据
type RelationEvidence struct {
	Type        EvidenceType           `json:"type"`        // 证据类型
	Source      string                 `json:"source"`      // 来源
	Content     interface{}            `json:"content"`     // 内容
	Reliability float64               `json:"reliability"` // 可靠性
	Timestamp   time.Time             `json:"timestamp"`   // 时间戳
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
}

// EvidenceType 证据类型
type InferenceEvidenceType string

const (
	InferenceEvidenceTypeEmpirical   InferenceEvidenceType = "empirical"   // 经验证据
	InferenceEvidenceTypeStatistical InferenceEvidenceType = "statistical" // 统计证据
	InferenceEvidenceTypeLogical     InferenceEvidenceType = "logical"     // 逻辑证据
	InferenceEvidenceTypeExpert      InferenceEvidenceType = "expert"      // 专家证据
	InferenceEvidenceTypeContextual  InferenceEvidenceType = "contextual"  // 上下文证据
)

// InferenceStep 推理步骤
type InferenceStep struct {
	ID          string                 `json:"id"`          // 步骤ID
	RuleID      string                 `json:"rule_id"`     // 规则ID
	Input       interface{}            `json:"input"`       // 输入
	Output      interface{}            `json:"output"`      // 输出
	Confidence  float64               `json:"confidence"`  // 置信度
	Explanation string                `json:"explanation"` // 解释
	Timestamp   time.Time             `json:"timestamp"`   // 时间戳
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
}

// CachedInferenceResult 缓存推理结果
type RelationCachedInferenceResult struct {
	Input       string                 `json:"input"`       // 输入哈希
	Result      *InferenceResult       `json:"result"`      // 结果
	Confidence  float64               `json:"confidence"`  // 置信度
	Timestamp   time.Time             `json:"timestamp"`   // 时间戳
	AccessCount int                   `json:"access_count"` // 访问次数
}

// InferenceRequest 推理请求
type InferenceRequest struct {
	Nodes           []*entities.KnowledgeNode     `json:"nodes"`           // 节点列表
	ExistingRelations []*entities.KnowledgeRelation `json:"existing_relations"` // 现有关系
	Context         *InferenceContext             `json:"context"`         // 推理上下文
	Options         *InferenceOptions             `json:"options"`         // 推理选项
	TargetRelations []entities.RelationType       `json:"target_relations"` // 目标关系类型
}

// InferenceContext 推理上下文
type InferenceContext struct {
	Domain          string                 `json:"domain"`          // 领域
	Subject         string                 `json:"subject"`         // 主题
	LearnerProfile  *LearnerProfile        `json:"learner_profile"` // 学习者档案
	TemporalContext *TemporalContext       `json:"temporal_context"` // 时间上下文
	SpatialContext  *SpatialContext        `json:"spatial_context"` // 空间上下文
	Metadata        map[string]interface{} `json:"metadata"`        // 元数据
}

// LearnerProfile 学习者档案
type RelationLearnerProfile struct {
	LearnerID      uuid.UUID              `json:"learner_id"`      // 学习者ID
	LearningStyle  string                 `json:"learning_style"`  // 学习风格
	KnowledgeLevel string                 `json:"knowledge_level"` // 知识水平
	Preferences    map[string]interface{} `json:"preferences"`     // 偏好
	History        []string               `json:"history"`         // 历史记录
}

// TemporalContext 时间上下文
type TemporalContext struct {
	CurrentTime time.Time `json:"current_time"` // 当前时间
	TimeWindow  int64     `json:"time_window"`  // 时间窗口
	Seasonality string    `json:"seasonality"`  // 季节性
}

// SpatialContext 空间上下文
type SpatialContext struct {
	Location    string                 `json:"location"`    // 位置
	Environment string                 `json:"environment"` // 环境
	Context     map[string]interface{} `json:"context"`     // 上下文
}

// InferenceOptions 推理选项
type InferenceOptions struct {
	MaxDepth            int     `json:"max_depth"`            // 最大深度
	MinConfidence       float64 `json:"min_confidence"`       // 最小置信度
	EnableExplanation   bool    `json:"enable_explanation"`   // 启用解释
	EnableUncertainty   bool    `json:"enable_uncertainty"`   // 启用不确定性
	EnableContradiction bool    `json:"enable_contradiction"` // 启用矛盾检查
	ParallelProcessing  bool    `json:"parallel_processing"`  // 并行处理
}

// InferenceResponse 推理响应
type InferenceResponse struct {
	InferredRelations []*InferredRelation    `json:"inferred_relations"` // 推理关系
	InferenceChain    []*InferenceStep       `json:"inference_chain"`    // 推理链
	Contradictions    []*Contradiction       `json:"contradictions"`     // 矛盾
	Uncertainties     []*Uncertainty         `json:"uncertainties"`      // 不确定性
	QualityMetrics    *InferenceQualityMetrics `json:"quality_metrics"`    // 质量指标
	ProcessingTime    int64                  `json:"processing_time"`    // 处理时间
	Explanations      []*Explanation         `json:"explanations"`       // 解释
}

// Contradiction 矛盾
type Contradiction struct {
	ID          string                 `json:"id"`          // 矛盾ID
	Type        ContradictionType      `json:"type"`        // 矛盾类型
	Relations   []*InferredRelation    `json:"relations"`   // 相关关系
	Severity    float64               `json:"severity"`    // 严重程度
	Resolution  *ResolutionSuggestion  `json:"resolution"`  // 解决建议
	Evidence    []*Evidence           `json:"evidence"`    // 证据
	Explanation string                `json:"explanation"` // 解释
}

// ContradictionType 矛盾类型
type ContradictionType string

const (
	ContradictionTypeLogical   ContradictionType = "logical"   // 逻辑矛盾
	ContradictionTypeTemporal  ContradictionType = "temporal"  // 时间矛盾
	ContradictionTypeOntological ContradictionType = "ontological" // 本体矛盾
)

// ResolutionSuggestion 解决建议
type ResolutionSuggestion struct {
	Type        InferenceResolutionType `json:"type"`        // 解决类型
	Action      string                 `json:"action"`      // 行动
	Priority    int                   `json:"priority"`    // 优先级
	Confidence  float64               `json:"confidence"`  // 置信度
	Explanation string                `json:"explanation"` // 解释
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
}

// InferenceResolutionType 解决类型
type InferenceResolutionType string

const (
	InferenceResolutionTypeRemove   InferenceResolutionType = "remove"   // 移除
	InferenceResolutionTypeModify   InferenceResolutionType = "modify"   // 修改
	InferenceResolutionTypeReweight InferenceResolutionType = "reweight" // 重新加权
	InferenceResolutionTypeIgnore   InferenceResolutionType = "ignore"   // 忽略
)

// Uncertainty 不确定性
type Uncertainty struct {
	ID          string                 `json:"id"`          // 不确定性ID
	Type        UncertaintyType        `json:"type"`        // 不确定性类型
	Source      string                 `json:"source"`      // 来源
	Level       float64               `json:"level"`       // 不确定性水平
	Impact      float64               `json:"impact"`      // 影响
	Mitigation  *MitigationStrategy    `json:"mitigation"`  // 缓解策略
	Explanation string                `json:"explanation"` // 解释
}

// UncertaintyType 不确定性类型
type UncertaintyType string

const (
	UncertaintyTypeEpistemic UncertaintyType = "epistemic" // 认知不确定性
	UncertaintyTypeAleatory  UncertaintyType = "aleatory"  // 随机不确定性
	UncertaintyTypeModel     UncertaintyType = "model"     // 模型不确定性
)

// MitigationStrategy 缓解策略
type MitigationStrategy struct {
	Type        MitigationType         `json:"type"`        // 缓解类型
	Action      string                 `json:"action"`      // 行动
	Confidence  float64               `json:"confidence"`  // 置信度
	Cost        float64               `json:"cost"`        // 成本
	Benefit     float64               `json:"benefit"`     // 收益
	Explanation string                `json:"explanation"` // 解释
}

// MitigationType 缓解类型
type MitigationType string

const (
	MitigationTypeDataCollection MitigationType = "data_collection" // 数据收集
	MitigationTypeModelImprovement MitigationType = "model_improvement" // 模型改进
	MitigationTypeExpertConsultation MitigationType = "expert_consultation" // 专家咨询
)

// InferenceQualityMetrics 推理质量指标
type InferenceQualityMetrics struct {
	Precision    float64 `json:"precision"`    // 精确度
	Recall       float64 `json:"recall"`       // 召回率
	F1Score      float64 `json:"f1_score"`     // F1分数
	Consistency  float64 `json:"consistency"`  // 一致性
	Completeness float64 `json:"completeness"` // 完整性
	Novelty      float64 `json:"novelty"`      // 新颖性
	Utility      float64 `json:"utility"`      // 实用性
}

// Explanation 解释
type Explanation struct {
	ID          string                 `json:"id"`          // 解释ID
	Type        ExplanationType        `json:"type"`        // 解释类型
	Content     string                 `json:"content"`     // 内容
	Confidence  float64               `json:"confidence"`  // 置信度
	Evidence    []*Evidence           `json:"evidence"`    // 证据
	Reasoning   string                `json:"reasoning"`   // 推理过程
	Metadata    map[string]interface{} `json:"metadata"`    // 元数据
}

// ExplanationType 解释类型
type ExplanationType string

const (
	ExplanationTypeDeductive   ExplanationType = "deductive"   // 演绎解释
	ExplanationTypeInductive   ExplanationType = "inductive"   // 归纳解释
	ExplanationTypeAbductive   ExplanationType = "abductive"   // 溯因解释
	ExplanationTypeContrastive ExplanationType = "contrastive" // 对比解释
	ExplanationTypeCounterfactual ExplanationType = "counterfactual" // 反事实解释
)

// NewIntelligentRelationInferenceEngine 创建智能关系推理引擎
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
		CachedResults:        make(map[string]*CachedInferenceResult),
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

	// 初始化默认推理规则
	engine.initializeDefaultRules()

	return engine
}

// ProcessInference 处理推理请求
func (e *IntelligentRelationInferenceEngine) ProcessInference(ctx context.Context, req *InferenceRequest) (*InferenceResponse, error) {
	startTime := time.Now()
	
	response := &InferenceResponse{
		InferredRelations: make([]*InferredRelation, 0),
		InferenceChain:    make([]*InferenceStep, 0),
		Contradictions:    make([]*Contradiction, 0),
		Uncertainties:     make([]*Uncertainty, 0),
		Explanations:      make([]*Explanation, 0),
	}

	// 1. 预处理和验证
	if err := e.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. 生成上下文嵌入
	if err := e.generateContextEmbeddings(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to generate context embeddings: %w", err)
	}

	// 3. 应用推理规则
	inferredRelations, inferenceChain := e.applyInferenceRules(ctx, req)
	response.InferredRelations = inferredRelations
	response.InferenceChain = inferenceChain

	// 4. 检查矛盾
	if e.config.EnableContradictionCheck {
		contradictions := e.detectContradictions(inferredRelations, req.ExistingRelations)
		response.Contradictions = contradictions
	}

	// 5. 分析不确定性
	if e.config.EnableUncertaintyReasoning {
		uncertainties := e.analyzeUncertainties(inferredRelations, inferenceChain)
		response.Uncertainties = uncertainties
	}

	// 6. 生成解释
	explanations := e.generateExplanations(inferredRelations, inferenceChain)
	response.Explanations = explanations

	// 7. 计算质量指标
	response.QualityMetrics = e.calculateQualityMetrics(inferredRelations, req.ExistingRelations)

	// 8. 更新指标和缓存
	e.updateMetrics(len(inferredRelations), time.Since(startTime))
	e.updateCache(req, response)

	response.ProcessingTime = time.Since(startTime).Milliseconds()
	return response, nil
}

// initializeDefaultRules 初始化默认推理规则
func (e *IntelligentRelationInferenceEngine) initializeDefaultRules() {
	// 传递性规则
	transitiveRule := &AdvancedInferenceRule{
		ID:          "transitive_prerequisite",
		Name:        "Transitive Prerequisite Rule",
		Description: "If A is prerequisite to B and B is prerequisite to C, then A is prerequisite to C",
		Type:        RuleTypeTransitive,
		Priority:    10,
		Confidence:  0.9,
		Enabled:     true,
	}

	// 层次结构规则
	hierarchyRule := &AdvancedInferenceRule{
		ID:          "hierarchy_part_of",
		Name:        "Hierarchy Part-Of Rule",
		Description: "If A is part of B and B is part of C, then A is part of C",
		Type:        RuleTypeTransitive,
		Priority:    9,
		Confidence:  0.85,
		Enabled:     true,
	}

	// 相似性规则
	similarityRule := &AdvancedInferenceRule{
		ID:          "similarity_related",
		Name:        "Similarity Related Rule",
		Description: "If A and B have high semantic similarity, they are likely related",
		Type:        RuleTypeInductive,
		Priority:    7,
		Confidence:  0.75,
		Enabled:     true,
	}

	// 因果关系规则
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

// validateRequest 验证请求
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

// generateContextEmbeddings 生成上下文嵌入
func (e *IntelligentRelationInferenceEngine) generateContextEmbeddings(ctx context.Context, req *InferenceRequest) error {
	if req.Context == nil {
		return nil
	}

	// 构建上下文文本
	contextText := fmt.Sprintf("Domain: %s, Subject: %s", req.Context.Domain, req.Context.Subject)
	
	// 使用跨模态AI生成嵌入
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

// applyInferenceRules 应用推理规则
func (e *IntelligentRelationInferenceEngine) applyInferenceRules(ctx context.Context, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	inferredRelations := make([]*InferredRelation, 0)
	inferenceChain := make([]*InferenceStep, 0)

	// 按优先级排序规则
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].Priority > e.rules[j].Priority
	})

	// 应用每个规则
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		relations, steps := e.applyRule(ctx, rule, req)
		inferredRelations = append(inferredRelations, relations...)
		inferenceChain = append(inferenceChain, steps...)
	}

	// 去重和过滤
	inferredRelations = e.deduplicateRelations(inferredRelations)
	inferredRelations = e.filterByConfidence(inferredRelations, e.config.MinConfidenceThreshold)

	return inferredRelations, inferenceChain
}

// applyRule 应用单个规则
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

// applyTransitiveRule 应用传递规则
func (e *IntelligentRelationInferenceEngine) applyTransitiveRule(rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// 构建关系图
	relationMap := make(map[uuid.UUID]map[uuid.UUID]*entities.KnowledgeRelation)
	for _, rel := range req.ExistingRelations {
		if relationMap[rel.FromNodeID] == nil {
			relationMap[rel.FromNodeID] = make(map[uuid.UUID]*entities.KnowledgeRelation)
		}
		relationMap[rel.FromNodeID][rel.ToNodeID] = rel
	}

	// 查找传递关系
	for _, nodeA := range req.Nodes {
		for _, nodeB := range req.Nodes {
			if nodeA.ID == nodeB.ID {
				continue
			}

			// 查找A->B的直接关系
			if relationMap[nodeA.ID] != nil {
				if relAB, exists := relationMap[nodeA.ID][nodeB.ID]; exists {
					// 查找B->C的关系
					if relationMap[nodeB.ID] != nil {
						for nodeC_ID, relBC := range relationMap[nodeB.ID] {
							if nodeC_ID == nodeA.ID {
								continue
							}

							// 检查是否已存在A->C的直接关系
							if relationMap[nodeA.ID][nodeC_ID] == nil {
								// 创建传递关系A->C
								confidence := math.Min(relAB.Confidence, relBC.Confidence) * rule.Confidence
								if confidence > e.config.MinConfidenceThreshold {
									inferredRel := &InferredRelation{
										FromNodeID:   nodeA.ID,
										ToNodeID:     nodeC_ID,
										RelationType: relAB.Type, // 保持相同的关系类型
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

// applyInductiveRule 应用归纳规则
func (e *IntelligentRelationInferenceEngine) applyInductiveRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// 基于相似性的归纳推理
	for i, nodeA := range req.Nodes {
		for j, nodeB := range req.Nodes {
			if i >= j {
				continue
			}

			// 计算节点相似性
			similarity := e.calculateNodeSimilarity(ctx, nodeA, nodeB)
			if similarity > 0.7 { // 高相似性阈值
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

// applyCausalRule 应用因果规则
func (e *IntelligentRelationInferenceEngine) applyCausalRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// 基于难度级别和类型的因果推理
	for _, nodeA := range req.Nodes {
		for _, nodeB := range req.Nodes {
			if nodeA.ID == nodeB.ID {
				continue
			}

			// 检查是否存在因果关系的条件
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

// applyAnalogicalRule 应用类比规则
func (e *IntelligentRelationInferenceEngine) applyAnalogicalRule(ctx context.Context, rule *AdvancedInferenceRule, req *InferenceRequest) ([]*InferredRelation, []*InferenceStep) {
	relations := make([]*InferredRelation, 0)
	steps := make([]*InferenceStep, 0)

	// 类比推理：如果A和B相似，B和C有关系，那么A和C可能也有类似关系
	for _, nodeA := range req.Nodes {
		for _, nodeB := range req.Nodes {
			if nodeA.ID == nodeB.ID {
				continue
			}

			similarity := e.calculateNodeSimilarity(ctx, nodeA, nodeB)
			if similarity > 0.8 { // 高相似性阈值用于类比
				// 查找B的关系，推理A的类似关系
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

// calculateNodeSimilarity 计算节点相似性
func (e *IntelligentRelationInferenceEngine) calculateNodeSimilarity(ctx context.Context, nodeA, nodeB *entities.KnowledgeNode) float64 {
	similarity := 0.0

	// 类型相似性
	if nodeA.Type == nodeB.Type {
		similarity += 0.3
	}

	// 难度相似性
	diffA := int(nodeA.Difficulty)
	diffB := int(nodeB.Difficulty)
	diffSim := 1.0 - math.Abs(float64(diffA-diffB))/4.0
	similarity += 0.2 * diffSim

	// 主题相似性
	if nodeA.Subject == nodeB.Subject {
		similarity += 0.3
	}

	// 标签相似性
	tagSim := e.calculateTagSimilarity(nodeA.Tags, nodeB.Tags)
	similarity += 0.2 * tagSim

	return similarity
}

// calculateTagSimilarity 计算标签相似性
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

// hasCausalRelationship 检查是否存在因果关系
func (e *IntelligentRelationInferenceEngine) hasCausalRelationship(nodeA, nodeB *entities.KnowledgeNode) bool {
	// 基于难度级别的因果关系
	if int(nodeA.Difficulty) < int(nodeB.Difficulty) && nodeA.Subject == nodeB.Subject {
		return true
	}

	// 基于节点类型的因果关系
	if nodeA.Type == entities.NodeTypeSkill && nodeB.Type == entities.NodeTypeSkill {
		return true
	}

	return false
}

// calculateCausalConfidence 计算因果置信度
func (e *IntelligentRelationInferenceEngine) calculateCausalConfidence(nodeA, nodeB *entities.KnowledgeNode) float64 {
	confidence := 0.5

	// 难度差异越小，因果关系越强
	diffA := int(nodeA.Difficulty)
	diffB := int(nodeB.Difficulty)
	if diffB > diffA {
		diffFactor := 1.0 - float64(diffB-diffA)/4.0
		confidence += 0.3 * diffFactor
	}

	// 相同主题增加置信度
	if nodeA.Subject == nodeB.Subject {
		confidence += 0.2
	}

	return math.Min(confidence, 1.0)
}

// deduplicateRelations 去重关系
func (e *IntelligentRelationInferenceEngine) deduplicateRelations(relations []*InferredRelation) []*InferredRelation {
	seen := make(map[string]*InferredRelation)
	result := make([]*InferredRelation, 0)

	for _, rel := range relations {
		key := fmt.Sprintf("%s-%s-%s", rel.FromNodeID, rel.ToNodeID, rel.RelationType)
		if existing, exists := seen[key]; exists {
			// 保留置信度更高的关系
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

// filterByConfidence 按置信度过滤
func (e *IntelligentRelationInferenceEngine) filterByConfidence(relations []*InferredRelation, threshold float64) []*InferredRelation {
	result := make([]*InferredRelation, 0)
	for _, rel := range relations {
		if rel.Confidence >= threshold {
			result = append(result, rel)
		}
	}
	return result
}

// detectContradictions 检测矛盾
func (e *IntelligentRelationInferenceEngine) detectContradictions(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) []*Contradiction {
	contradictions := make([]*Contradiction, 0)

	// 检查推理关系与现有关系的矛盾
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

// areContradictory 检查两个关系是否矛盾
func (e *IntelligentRelationInferenceEngine) areContradictory(inferred *InferredRelation, existing *entities.KnowledgeRelation) bool {
	// 相同节点对但关系类型矛盾
	if inferred.FromNodeID == existing.FromNodeID && inferred.ToNodeID == existing.ToNodeID {
		return e.areRelationTypesContradictory(inferred.RelationType, existing.Type)
	}

	// 反向关系矛盾
	if inferred.FromNodeID == existing.ToNodeID && inferred.ToNodeID == existing.FromNodeID {
		return e.areReverseRelationsContradictory(inferred.RelationType, existing.Type)
	}

	return false
}

// areRelationTypesContradictory 检查关系类型是否矛盾
func (e *IntelligentRelationInferenceEngine) areRelationTypesContradictory(type1, type2 entities.RelationType) bool {
	// 定义矛盾的关系类型对
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

// areReverseRelationsContradictory 检查反向关系是否矛盾
func (e *IntelligentRelationInferenceEngine) areReverseRelationsContradictory(type1, type2 entities.RelationType) bool {
	// 某些关系类型不应该有反向关系
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

// calculateContradictionSeverity 计算矛盾严重程度
func (e *IntelligentRelationInferenceEngine) calculateContradictionSeverity(inferred *InferredRelation, existing *entities.KnowledgeRelation) float64 {
	// 基于置信度差异计算严重程度
	confidenceDiff := math.Abs(inferred.Confidence - existing.Confidence)
	severity := 0.5 + 0.5*confidenceDiff

	// 某些关系类型的矛盾更严重
	if inferred.RelationType == entities.RelationTypeOppositeOf || existing.Type == entities.RelationTypeOppositeOf {
		severity *= 1.5
	}

	return math.Min(severity, 1.0)
}

// analyzeUncertainties 分析不确定性
func (e *IntelligentRelationInferenceEngine) analyzeUncertainties(inferredRelations []*InferredRelation, inferenceChain []*InferenceStep) []*Uncertainty {
	uncertainties := make([]*Uncertainty, 0)

	// 分析低置信度关系的不确定性
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

// calculateUncertaintyImpact 计算不确定性影响
func (e *IntelligentRelationInferenceEngine) calculateUncertaintyImpact(rel *InferredRelation) float64 {
	// 基于关系类型和权重计算影响
	impact := rel.Weight

	// 某些关系类型的不确定性影响更大
	if rel.RelationType == entities.RelationTypePrerequisite {
		impact *= 1.2
	}

	return math.Min(impact, 1.0)
}

// generateExplanations 生成解释
func (e *IntelligentRelationInferenceEngine) generateExplanations(inferredRelations []*InferredRelation, inferenceChain []*InferenceStep) []*Explanation {
	explanations := make([]*Explanation, 0)

	// 为每个推理步骤生成解释
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

// calculateQualityMetrics 计算质量指标
func (e *IntelligentRelationInferenceEngine) calculateQualityMetrics(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) *InferenceQualityMetrics {
	metrics := &InferenceQualityMetrics{}

	if len(inferredRelations) == 0 {
		return metrics
	}

	// 计算平均置信度作为精确度的代理
	totalConfidence := 0.0
	for _, rel := range inferredRelations {
		totalConfidence += rel.Confidence
	}
	metrics.Precision = totalConfidence / float64(len(inferredRelations))

	// 计算召回率（简化版本）
	totalPossibleRelations := len(existingRelations) + len(inferredRelations)
	if totalPossibleRelations > 0 {
		metrics.Recall = float64(len(inferredRelations)) / float64(totalPossibleRelations)
	}

	// 计算F1分数
	if metrics.Precision+metrics.Recall > 0 {
		metrics.F1Score = 2 * (metrics.Precision * metrics.Recall) / (metrics.Precision + metrics.Recall)
	}

	// 计算一致性
	metrics.Consistency = e.calculateConsistencyScore(inferredRelations)

	// 计算完整性
	metrics.Completeness = e.calculateCompletenessScore(inferredRelations, existingRelations)

	// 计算新颖性
	metrics.Novelty = e.calculateNoveltyScore(inferredRelations, existingRelations)

	// 计算实用性
	metrics.Utility = e.calculateUtilityScore(inferredRelations)

	return metrics
}

// calculateConsistencyScore 计算一致性分数
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

// areInferredRelationsContradictory 检查推理关系是否矛盾
func (e *IntelligentRelationInferenceEngine) areInferredRelationsContradictory(relA, relB *InferredRelation) bool {
	// 相同节点对但关系类型矛盾
	if relA.FromNodeID == relB.FromNodeID && relA.ToNodeID == relB.ToNodeID {
		return e.areRelationTypesContradictory(relA.RelationType, relB.RelationType)
	}

	// 反向关系矛盾
	if relA.FromNodeID == relB.ToNodeID && relA.ToNodeID == relB.FromNodeID {
		return e.areReverseRelationsContradictory(relA.RelationType, relB.RelationType)
	}

	return false
}

// calculateCompletenessScore 计算完整性分数
func (e *IntelligentRelationInferenceEngine) calculateCompletenessScore(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) float64 {
	// 简化的完整性计算：推理关系数量与现有关系数量的比例
	if len(existingRelations) == 0 {
		return 1.0
	}

	ratio := float64(len(inferredRelations)) / float64(len(existingRelations))
	return math.Min(ratio, 1.0)
}

// calculateNoveltyScore 计算新颖性分数
func (e *IntelligentRelationInferenceEngine) calculateNoveltyScore(inferredRelations []*InferredRelation, existingRelations []*entities.KnowledgeRelation) float64 {
	if len(inferredRelations) == 0 {
		return 0.0
	}

	novelRelations := 0
	existingRelationSet := make(map[string]bool)

	// 构建现有关系集合
	for _, rel := range existingRelations {
		key := fmt.Sprintf("%s-%s-%s", rel.FromNodeID, rel.ToNodeID, rel.Type)
		existingRelationSet[key] = true
	}

	// 检查推理关系的新颖性
	for _, rel := range inferredRelations {
		key := fmt.Sprintf("%s-%s-%s", rel.FromNodeID, rel.ToNodeID, rel.RelationType)
		if !existingRelationSet[key] {
			novelRelations++
		}
	}

	return float64(novelRelations) / float64(len(inferredRelations))
}

// calculateUtilityScore 计算实用性分数
func (e *IntelligentRelationInferenceEngine) calculateUtilityScore(relations []*InferredRelation) float64 {
	if len(relations) == 0 {
		return 0.0
	}

	totalUtility := 0.0
	for _, rel := range relations {
		// 基于关系类型和置信度计算实用性
		utility := rel.Confidence

		// 某些关系类型更有用
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

// updateMetrics 更新指标
func (e *IntelligentRelationInferenceEngine) updateMetrics(inferenceCount int, processingTime time.Duration) {
	e.metrics.TotalInferences += int64(inferenceCount)
	e.metrics.SuccessfulInferences += int64(inferenceCount)
	e.metrics.LastInferenceTime = time.Now()

	// 更新平均处理时间
	if e.metrics.TotalInferences > 0 {
		totalTime := e.metrics.AverageProcessingTime*int64(e.metrics.TotalInferences-int64(inferenceCount)) + processingTime.Milliseconds()
		e.metrics.AverageProcessingTime = totalTime / e.metrics.TotalInferences
	} else {
		e.metrics.AverageProcessingTime = processingTime.Milliseconds()
	}
}

// updateCache 更新缓存
func (e *IntelligentRelationInferenceEngine) updateCache(req *InferenceRequest, resp *InferenceResponse) {
	if !e.config.CacheEnabled {
		return
	}

	// 缓存推理结果
	inputHash := e.generateInputHash(req)
	cachedResult := &CachedInferenceResult{
		Input:       inputHash,
		Result:      &InferenceResult{}, // 简化版本
		Confidence:  e.calculateAverageConfidence(resp.InferredRelations),
		Timestamp:   time.Now(),
		AccessCount: 1,
	}

	e.cache.CachedResults[inputHash] = cachedResult
	e.cache.LastUpdated = time.Now()
}

// generateInputHash 生成输入哈希
func (e *IntelligentRelationInferenceEngine) generateInputHash(req *InferenceRequest) string {
	// 简化的哈希生成
	nodeIDs := make([]string, len(req.Nodes))
	for i, node := range req.Nodes {
		nodeIDs[i] = node.ID.String()
	}
	sort.Strings(nodeIDs)
	return fmt.Sprintf("%x", nodeIDs)
}

// calculateAverageConfidence 计算平均置信度
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

// GetMetrics 获取推理指标
func (e *IntelligentRelationInferenceEngine) GetMetrics() *InferenceMetrics {
	return e.metrics
}

// UpdateConfig 更新配置
func (e *IntelligentRelationInferenceEngine) UpdateConfig(config *InferenceEngineConfig) {
	e.config = config
}

// AddRule 添加推理规则
func (e *IntelligentRelationInferenceEngine) AddRule(rule *AdvancedInferenceRule) {
	e.rules = append(e.rules, rule)
}

// RemoveRule 移除推理规则
func (e *IntelligentRelationInferenceEngine) RemoveRule(ruleID string) {
	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			break
		}
	}
}

// EnableRule 启用推理规则
func (e *IntelligentRelationInferenceEngine) EnableRule(ruleID string) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			rule.Enabled = true
			break
		}
	}
}

// DisableRule 禁用推理规则
func (e *IntelligentRelationInferenceEngine) DisableRule(ruleID string) {
	for _, rule := range e.rules {
		if rule.ID == ruleID {
			rule.Enabled = false
			break
		}
	}
}

// ClearCache 清空缓存
func (e *IntelligentRelationInferenceEngine) ClearCache() {
	e.cache.RelationProbabilities = make(map[string]float64)
	e.cache.InferenceChains = make(map[string][]*InferenceStep)
	e.cache.ContextEmbeddings = make(map[string][]float64)
	e.cache.CachedResults = make(map[string]*CachedInferenceResult)
	e.cache.LastUpdated = time.Now()
}