package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// CrossModalServiceInterface 跨模态服务接口
type CrossModalServiceInterface interface {
	ProcessCrossModalInference(ctx context.Context, request *CrossModalInferenceRequest) (*CrossModalInferenceResponse, error)
}

// CrossModalInferenceRequest 跨模态推理请求
type CrossModalInferenceRequest struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Options   map[string]interface{} `json:"options"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

// CrossModalInferenceResponse 跨模态推理响应
type CrossModalInferenceResponse struct {
	Success bool                   `json:"success"`
	Result  map[string]interface{} `json:"result"`
	Error   string                 `json:"error,omitempty"`
}

// LearnerProfile 学习者档案
type LearnerProfile struct {
	UserID             string                 `json:"user_id"`
	LearningStyle      string                 `json:"learning_style"`
	PreferredDifficulty string                `json:"preferred_difficulty"`
	Interests          []string               `json:"interests"`
	Goals              []string               `json:"goals"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// AutomatedKnowledgeGraphService 自动化知识图谱构建服务
type AutomatedKnowledgeGraphService struct {
	graphRepo           repositories.KnowledgeGraphRepository
	contentRepo         repositories.LearningContentRepository
	crossModalService   CrossModalServiceInterface
	inferenceEngine     *IntelligentRelationInferenceEngine
	config              *AutomatedGraphConfig
	cache               *AutomatedGraphCache
	metrics             *AutomatedGraphMetrics
}

// AutomatedGraphConfig 自动化图谱配置
type AutomatedGraphConfig struct {
	MinConfidenceThreshold    float64 `json:"min_confidence_threshold"`    // 最小置信度阈值
	MaxRelationsPerNode       int     `json:"max_relations_per_node"`      // 每个节点最大关系数
	AutoInferenceEnabled      bool    `json:"auto_inference_enabled"`      // 是否启用自动推理
	SemanticSimilarityWeight  float64 `json:"semantic_similarity_weight"`  // 语义相似度权重
	StructuralSimilarityWeight float64 `json:"structural_similarity_weight"` // 结构相似度权重
	ContentAnalysisDepth      int     `json:"content_analysis_depth"`      // 内容分析深度
	BatchProcessingSize       int     `json:"batch_processing_size"`       // 批处理大小
	UpdateInterval            int     `json:"update_interval"`             // 更新间隔(分钟)
}

// AutomatedGraphCache 自动化图谱缓存
type AutomatedGraphCache struct {
	NodeEmbeddings     map[uuid.UUID][]float64           `json:"node_embeddings"`     // 节点嵌入向量
	RelationScores     map[string]float64                `json:"relation_scores"`     // 关系评分
	InferenceResults   map[string]*InferenceResult       `json:"inference_results"`   // 推理结果
	SemanticClusters   map[string][]uuid.UUID            `json:"semantic_clusters"`   // 语义聚类
	LastUpdated        time.Time                         `json:"last_updated"`        // 最后更新时间
	ProcessingQueue    []uuid.UUID                       `json:"processing_queue"`    // 处理队列
}

// AutomatedGraphMetrics 自动化图谱指标
type AutomatedGraphMetrics struct {
	NodesProcessed        int64     `json:"nodes_processed"`        // 已处理节点数
	RelationsInferred     int64     `json:"relations_inferred"`     // 推理出的关系数
	SuccessfulInferences  int64     `json:"successful_inferences"`  // 成功推理数
	FailedInferences      int64     `json:"failed_inferences"`      // 失败推理数
	AverageConfidence     float64   `json:"average_confidence"`     // 平均置信度
	ProcessingTime        int64     `json:"processing_time"`        // 处理时间(毫秒)
	LastProcessingTime    time.Time `json:"last_processing_time"`   // 最后处理时间
	QualityScore          float64   `json:"quality_score"`          // 质量评分
}

// InferenceResult 推理结果
type InferenceResult struct {
	FromNodeID    uuid.UUID                `json:"from_node_id"`    // 源节点ID
	ToNodeID      uuid.UUID                `json:"to_node_id"`      // 目标节点ID
	RelationType  entities.RelationType    `json:"relation_type"`   // 关系类型
	Confidence    float64                  `json:"confidence"`      // 置信度
	Evidence      []string                 `json:"evidence"`        // 证据
	Reasoning     string                   `json:"reasoning"`       // 推理过程
	Metadata      map[string]interface{}   `json:"metadata"`        // 元数据
	CreatedAt     time.Time                `json:"created_at"`      // 创建时间
}

// AutoBuildRequest 自动构建请求
type AutoBuildRequest struct {
	ContentIDs        []uuid.UUID            `json:"content_ids"`        // 内容ID列表
	AnalysisDepth     int                    `json:"analysis_depth"`     // 分析深度
	EnableInference   bool                   `json:"enable_inference"`   // 启用推理
	CustomRules       []InferenceRule        `json:"custom_rules"`       // 自定义规则
	Options           map[string]interface{} `json:"options"`            // 选项
}

// InferenceRule 推理规则
type InferenceRule struct {
	ID          string                 `json:"id"`          // 规则ID
	Name        string                 `json:"name"`        // 规则名称
	Description string                 `json:"description"` // 规则描述
	Conditions  []RuleCondition        `json:"conditions"`  // 条件
	Actions     []RuleAction           `json:"actions"`     // 动作
	Priority    int                    `json:"priority"`    // 优先级
	Enabled     bool                   `json:"enabled"`     // 是否启用
}

// RuleCondition 规则条件
type RuleCondition struct {
	Type      string      `json:"type"`      // 条件类型
	Field     string      `json:"field"`     // 字段
	Operator  string      `json:"operator"`  // 操作符
	Value     interface{} `json:"value"`     // 值
	Weight    float64     `json:"weight"`    // 权重
}

// RuleAction 规则动作
type RuleAction struct {
	Type       string                 `json:"type"`       // 动作类型
	Parameters map[string]interface{} `json:"parameters"` // 参数
}

// AutoBuildResponse 自动构建响应
type AutoBuildResponse struct {
	NodesCreated      int                            `json:"nodes_created"`      // 创建的节点数
	RelationsCreated  int                            `json:"relations_created"`  // 创建的关系数
	InferenceResults  []*InferenceResult             `json:"inference_results"`  // 推理结果
	QualityMetrics    *KnowledgeGraphQualityMetrics  `json:"quality_metrics"`    // 质量指标
	ProcessingTime    int64                          `json:"processing_time"`    // 处理时间
	Warnings          []string                       `json:"warnings"`           // 警告信息
}

// KnowledgeGraphQualityMetrics 知识图谱质量指标
type KnowledgeGraphQualityMetrics struct {
	Completeness      float64 `json:"completeness"`      // 完整性
	Consistency       float64 `json:"consistency"`       // 一致性
	Accuracy          float64 `json:"accuracy"`          // 准确性
	Relevance         float64 `json:"relevance"`         // 相关性
	Coverage          float64 `json:"coverage"`          // 覆盖率
	Redundancy        float64 `json:"redundancy"`        // 冗余度
	OverallScore      float64 `json:"overall_score"`     // 总体评分
}

// NewAutomatedKnowledgeGraphService 创建自动化知识图谱服务
func NewAutomatedKnowledgeGraphService(
	graphRepo repositories.KnowledgeGraphRepository,
	contentRepo repositories.LearningContentRepository,
	crossModalService CrossModalServiceInterface,
) *AutomatedKnowledgeGraphService {
	// 创建智能关系推理引擎
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

// AutoBuildFromContent 从内容自动构建知识图谱
func (s *AutomatedKnowledgeGraphService) AutoBuildFromContent(ctx context.Context, req *AutoBuildRequest) (*AutoBuildResponse, error) {
	startTime := time.Now()
	response := &AutoBuildResponse{
		InferenceResults: make([]*InferenceResult, 0),
		Warnings:         make([]string, 0),
	}

	// 1. 分析内容并提取知识点
	nodes, err := s.extractKnowledgeNodes(ctx, req.ContentIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to extract knowledge nodes: %w", err)
	}
	response.NodesCreated = len(nodes)

	// 2. 生成节点嵌入向量
	if err := s.generateNodeEmbeddings(ctx, nodes); err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("Failed to generate embeddings: %v", err))
	}

	// 3. 推理节点间关系
	relations, inferenceResults := s.inferRelations(ctx, nodes, req.CustomRules)
	response.RelationsCreated = len(relations)
	response.InferenceResults = inferenceResults

	// 4. 验证和优化图谱结构
	if err := s.validateAndOptimizeGraph(ctx, nodes, relations); err != nil {
		response.Warnings = append(response.Warnings, fmt.Sprintf("Graph optimization warning: %v", err))
	}

	// 5. 计算质量指标
	response.QualityMetrics = s.calculateQualityMetrics(nodes, relations, inferenceResults)

	// 6. 更新指标
	s.updateMetrics(len(nodes), len(relations), len(inferenceResults))

	response.ProcessingTime = time.Since(startTime).Milliseconds()
	return response, nil
}

// extractKnowledgeNodes 从内容中提取知识节点
func (s *AutomatedKnowledgeGraphService) extractKnowledgeNodes(ctx context.Context, contentIDs []uuid.UUID) ([]*entities.KnowledgeNode, error) {
	nodes := make([]*entities.KnowledgeNode, 0)

	for _, contentID := range contentIDs {
		content, err := s.contentRepo.GetByID(ctx, contentID)
		if err != nil {
			continue
		}

		// 使用跨模态AI分析内容
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

		// 解析分析结果并创建节点
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

// createNodeFromConcept 从概念创建知识节点
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

// generateNodeEmbeddings 生成节点嵌入向量
func (s *AutomatedKnowledgeGraphService) generateNodeEmbeddings(ctx context.Context, nodes []*entities.KnowledgeNode) error {
	for _, node := range nodes {
		// 构建节点文本表示
		text := fmt.Sprintf("%s %s %s", node.Name, node.Description, strings.Join(node.Keywords, " "))
		
		// 使用跨模态AI生成嵌入向量
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

// inferRelations 推理节点间关系
func (s *AutomatedKnowledgeGraphService) inferRelations(ctx context.Context, nodes []*entities.KnowledgeNode, customRules []InferenceRule) ([]*entities.KnowledgeRelation, []*InferenceResult) {
	relations := make([]*entities.KnowledgeRelation, 0)
	inferenceResults := make([]*InferenceResult, 0)

	// 计算节点间的相似度矩阵
	similarityMatrix := s.calculateSimilarityMatrix(nodes)

	// 应用推理规则
	for i, nodeA := range nodes {
		for j, nodeB := range nodes {
			if i >= j {
				continue
			}

			// 基于相似度推理关系
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

			// 应用自定义规则
			for _, rule := range customRules {
				if rule.Enabled && s.evaluateRule(rule, nodeA, nodeB) {
					// 执行规则动作
					s.executeRuleActions(rule.Actions, nodeA, nodeB, &relations, &inferenceResults)
				}
			}
		}
	}

	return relations, inferenceResults
}

// InferRelations 推理关系 - 使用智能推理引擎
func (s *AutomatedKnowledgeGraphService) InferRelations(ctx context.Context, nodes []*entities.KnowledgeNode, existingRelations []*entities.KnowledgeRelation) ([]*InferenceResult, error) {
	// 构建推理请求
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

	// 使用智能推理引擎处理
	inferenceResp, err := s.inferenceEngine.ProcessInference(ctx, inferenceReq)
	if err != nil {
		return nil, fmt.Errorf("inference engine failed: %w", err)
	}

	// 转换结果格式
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

// InferRelationsWithContext 带上下文的关系推理
func (s *AutomatedKnowledgeGraphService) InferRelationsWithContext(ctx context.Context, nodes []*entities.KnowledgeNode, existingRelations []*entities.KnowledgeRelation, learnerProfile *LearnerProfile) ([]*InferenceResult, error) {
	// 构建带学习者上下文的推理请求
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

	// 使用智能推理引擎处理
	inferenceResp, err := s.inferenceEngine.ProcessInference(ctx, inferenceReq)
	if err != nil {
		return nil, fmt.Errorf("inference engine failed: %w", err)
	}

	// 转换结果格式
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

// calculateSimilarityMatrix 计算相似度矩阵
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

			// 计算语义相似度
			semanticSim := s.calculateSemanticSimilarity(nodeA, nodeB)
			
			// 计算结构相似度
			structuralSim := s.calculateStructuralSimilarity(nodeA, nodeB)
			
			// 加权组合
			similarity := s.config.SemanticSimilarityWeight*semanticSim + 
						 s.config.StructuralSimilarityWeight*structuralSim

			matrix[i][j] = similarity
		}
	}

	return matrix
}

// calculateSemanticSimilarity 计算语义相似度
func (s *AutomatedKnowledgeGraphService) calculateSemanticSimilarity(nodeA, nodeB *entities.KnowledgeNode) float64 {
	embeddingA, okA := s.cache.NodeEmbeddings[nodeA.ID]
	embeddingB, okB := s.cache.NodeEmbeddings[nodeB.ID]

	if !okA || !okB {
		// 回退到基于文本的相似度计算
		return s.calculateTextSimilarity(nodeA, nodeB)
	}

	// 计算余弦相似度
	return s.cosineSimilarity(embeddingA, embeddingB)
}

// calculateStructuralSimilarity 计算结构相似度
func (s *AutomatedKnowledgeGraphService) calculateStructuralSimilarity(nodeA, nodeB *entities.KnowledgeNode) float64 {
	similarity := 0.0

	// 类型相似度
	if nodeA.Type == nodeB.Type {
		similarity += 0.3
	}

	// 难度相似度
	diffA := int(nodeA.DifficultyLevel)
	diffB := int(nodeB.DifficultyLevel)
	diffSim := 1.0 - math.Abs(float64(diffA-diffB))/4.0
	similarity += 0.2 * diffSim

	// 主题相似度
	if nodeA.Subject == nodeB.Subject {
		similarity += 0.3
	}

	// 标签相似度
	tagSim := s.calculateTagSimilarity(nodeA.Tags, nodeB.Tags)
	similarity += 0.2 * tagSim

	return similarity
}

// calculateTextSimilarity 计算文本相似度
func (s *AutomatedKnowledgeGraphService) calculateTextSimilarity(nodeA, nodeB *entities.KnowledgeNode) float64 {
	textA := strings.ToLower(nodeA.Name + " " + nodeA.Description)
	textB := strings.ToLower(nodeB.Name + " " + nodeB.Description)

	wordsA := strings.Fields(textA)
	wordsB := strings.Fields(textB)

	// 计算Jaccard相似度
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

// calculateTagSimilarity 计算标签相似度
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

// cosineSimilarity 计算余弦相似度
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

// inferRelationType 推理关系类型
func (s *AutomatedKnowledgeGraphService) inferRelationType(nodeA, nodeB *entities.KnowledgeNode) (entities.RelationType, float64) {
	// 基于节点类型和属性推理关系类型
	if nodeA.Type == entities.NodeTypeConcept && nodeB.Type == entities.NodeTypeConcept {
		return entities.RelationTypeRelatedTo, 0.8
	}

	if nodeA.Type == entities.NodeTypeSkill && nodeB.Type == entities.NodeTypeSkill {
		// 检查难度级别来判断是否为前置关系
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

// evaluateRule 评估规则
func (s *AutomatedKnowledgeGraphService) evaluateRule(rule InferenceRule, nodeA, nodeB *entities.KnowledgeNode) bool {
	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(condition, nodeA, nodeB) {
			return false
		}
	}
	return true
}

// evaluateCondition 评估条件
func (s *AutomatedKnowledgeGraphService) evaluateCondition(condition RuleCondition, nodeA, nodeB *entities.KnowledgeNode) bool {
	// 简化的条件评估逻辑
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

// evaluateNodeTypeCondition 评估节点类型条件
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

// evaluateDifficultyCondition 评估难度条件
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

// evaluateSubjectCondition 评估主题条件
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

// executeRuleActions 执行规则动作
func (s *AutomatedKnowledgeGraphService) executeRuleActions(actions []RuleAction, nodeA, nodeB *entities.KnowledgeNode, relations *[]*entities.KnowledgeRelation, inferenceResults *[]*InferenceResult) {
	for _, action := range actions {
		switch action.Type {
		case "create_relation":
			s.executeCreateRelationAction(action, nodeA, nodeB, relations, inferenceResults)
		}
	}
}

// executeCreateRelationAction 执行创建关系动作
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

// validateAndOptimizeGraph 验证和优化图谱结构
func (s *AutomatedKnowledgeGraphService) validateAndOptimizeGraph(ctx context.Context, nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) error {
	// 检查循环依赖
	if s.hasCyclicDependencies(nodes, relations) {
		return fmt.Errorf("cyclic dependencies detected in graph")
	}

	// 移除冗余关系
	s.removeRedundantRelations(relations)

	// 优化关系权重
	s.optimizeRelationWeights(relations)

	return nil
}

// hasCyclicDependencies 检查循环依赖
func (s *AutomatedKnowledgeGraphService) hasCyclicDependencies(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) bool {
	// 构建邻接表
	graph := make(map[uuid.UUID][]uuid.UUID)
	for _, relation := range relations {
		if relation.Type == entities.RelationTypePrerequisite {
			graph[relation.FromNodeID] = append(graph[relation.FromNodeID], relation.ToNodeID)
		}
	}

	// DFS检查循环
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

// dfsHasCycle DFS检查循环
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

// removeRedundantRelations 移除冗余关系
func (s *AutomatedKnowledgeGraphService) removeRedundantRelations(relations []*entities.KnowledgeRelation) {
	// 按置信度排序
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Confidence > relations[j].Confidence
	})

	// 移除重复关系
	seen := make(map[string]bool)
	filtered := make([]*entities.KnowledgeRelation, 0)

	for _, relation := range relations {
		key := fmt.Sprintf("%s-%s-%s", relation.FromNodeID, relation.ToNodeID, relation.Type)
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, relation)
		}
	}

	// 更新relations切片
	copy(relations, filtered)
}

// optimizeRelationWeights 优化关系权重
func (s *AutomatedKnowledgeGraphService) optimizeRelationWeights(relations []*entities.KnowledgeRelation) {
	for _, relation := range relations {
		// 基于多个因素调整权重
		adjustedWeight := relation.Weight

		// 根据关系类型调整
		switch relation.Type {
		case entities.RelationTypePrerequisite:
			adjustedWeight *= 1.1 // 前置关系更重要
		case entities.RelationTypePartOf:
			adjustedWeight *= 1.05
		}

		// 确保权重在有效范围内
		if adjustedWeight > 1.0 {
			adjustedWeight = 1.0
		}
		if adjustedWeight < 0.0 {
			adjustedWeight = 0.0
		}

		relation.Weight = adjustedWeight
	}
}

// calculateQualityMetrics 计算质量指标
func (s *AutomatedKnowledgeGraphService) calculateQualityMetrics(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation, inferenceResults []*InferenceResult) *KnowledgeGraphQualityMetrics {
	metrics := &KnowledgeGraphQualityMetrics{}

	// 计算完整性
	metrics.Completeness = s.calculateCompleteness(nodes, relations)

	// 计算一致性
	metrics.Consistency = s.calculateConsistency(relations)

	// 计算准确性
	metrics.Accuracy = s.calculateAccuracy(inferenceResults)

	// 计算相关性
	metrics.Relevance = s.calculateRelevance(nodes, relations)

	// 计算覆盖率
	metrics.Coverage = s.calculateCoverage(nodes)

	// 计算冗余度
	metrics.Redundancy = s.calculateRedundancy(relations)

	// 计算总体评分
	metrics.OverallScore = (metrics.Completeness + metrics.Consistency + metrics.Accuracy + 
						   metrics.Relevance + metrics.Coverage + (1.0-metrics.Redundancy)) / 6.0

	return metrics
}

// calculateCompleteness 计算完整性
func (s *AutomatedKnowledgeGraphService) calculateCompleteness(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) float64 {
	if len(nodes) == 0 {
		return 0.0
	}

	// 计算节点连接度
	connectedNodes := make(map[uuid.UUID]bool)
	for _, relation := range relations {
		connectedNodes[relation.FromNodeID] = true
		connectedNodes[relation.ToNodeID] = true
	}

	return float64(len(connectedNodes)) / float64(len(nodes))
}

// calculateConsistency 计算一致性
func (s *AutomatedKnowledgeGraphService) calculateConsistency(relations []*entities.KnowledgeRelation) float64 {
	if len(relations) == 0 {
		return 1.0
	}

	// 检查关系的一致性（例如，没有矛盾的关系）
	consistentRelations := 0
	for _, relation := range relations {
		if relation.Confidence > 0.5 {
			consistentRelations++
		}
	}

	return float64(consistentRelations) / float64(len(relations))
}

// calculateAccuracy 计算准确性
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

// calculateRelevance 计算相关性
func (s *AutomatedKnowledgeGraphService) calculateRelevance(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) float64 {
	// 基于节点类型和关系类型的相关性评估
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

// calculateCoverage 计算覆盖率
func (s *AutomatedKnowledgeGraphService) calculateCoverage(nodes []*entities.KnowledgeNode) float64 {
	// 计算不同类型节点的覆盖率
	typeCount := make(map[entities.NodeType]int)
	for _, node := range nodes {
		typeCount[node.Type]++
	}

	// 期望的类型数量
	expectedTypes := 4 // concept, skill, topic, subject
	actualTypes := len(typeCount)

	return float64(actualTypes) / float64(expectedTypes)
}

// calculateRedundancy 计算冗余度
func (s *AutomatedKnowledgeGraphService) calculateRedundancy(relations []*entities.KnowledgeRelation) float64 {
	if len(relations) == 0 {
		return 0.0
	}

	// 检查重复关系
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

// updateMetrics 更新指标
func (s *AutomatedKnowledgeGraphService) updateMetrics(nodesProcessed, relationsCreated, inferencesCount int) {
	s.metrics.NodesProcessed += int64(nodesProcessed)
	s.metrics.RelationsInferred += int64(relationsCreated)
	s.metrics.SuccessfulInferences += int64(inferencesCount)
	s.metrics.LastProcessingTime = time.Now()
}

// GetMetrics 获取指标
func (s *AutomatedKnowledgeGraphService) GetMetrics() *AutomatedGraphMetrics {
	return s.metrics
}

// UpdateConfig 更新配置
func (s *AutomatedKnowledgeGraphService) UpdateConfig(config *AutomatedGraphConfig) {
	s.config = config
}