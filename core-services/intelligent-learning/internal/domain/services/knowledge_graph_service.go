package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// KnowledgeGraphService 知识图谱推理服务
type KnowledgeGraphService struct {
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
}

// NewKnowledgeGraphService 创建知识图谱服务
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

// GraphAnalysisRequest 图谱分析请求
type GraphAnalysisRequest struct {
	GraphID     uuid.UUID `json:"graph_id"`
	AnalysisType string   `json:"analysis_type"` // "structure", "learning_gaps", "optimization", "recommendations"
	LearnerID   *uuid.UUID `json:"learner_id,omitempty"`
	Depth       int       `json:"depth"`
	IncludeMetrics bool   `json:"include_metrics"`
}

// GraphAnalysisResult 图谱分析结果
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

// StructuralMetrics 结构化指标
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

// LearningGap 学习缺口
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

// GraphRecommendation 图谱推荐
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

// RecommendationAction 推荐动作
type RecommendationAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OptimizationSuggestion 优化建议
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

// ConceptRecommendationRequest 概念推荐请求
type ConceptRecommendationRequest struct {
	GraphID       uuid.UUID   `json:"graph_id"`
	LearnerID     uuid.UUID   `json:"learner_id"`
	CurrentNodeID *uuid.UUID  `json:"current_node_id,omitempty"`
	TargetSkills  []string    `json:"target_skills,omitempty"`
	MaxRecommendations int    `json:"max_recommendations"`
	IncludeReasoning bool     `json:"include_reasoning"`
}

// ConceptRecommendation 概念推荐
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

// AnalyzeGraph 分析知识图谱
func (s *KnowledgeGraphService) AnalyzeGraph(ctx context.Context, req *GraphAnalysisRequest) (*GraphAnalysisResult, error) {
	// 获取知识图谱
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

	// 根据分析类型执行不同的分析
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
		// 综合分析
		result.StructuralMetrics = s.analyzeStructure(ctx, graph)
		result.LearningGaps = s.identifyGeneralGaps(ctx, graph)
		result.OptimizationSuggestions = s.generateOptimizationSuggestions(ctx, graph)
		result.Recommendations = s.generateGraphRecommendations(ctx, graph)
		result.QualityScore = s.calculateOverallQuality(result)
	}

	// 生成洞察和警告
	result.Insights = s.generateInsights(result)
	result.Warnings = s.generateWarnings(result)

	return result, nil
}

// RecommendConcepts 推荐学习概念
func (s *KnowledgeGraphService) RecommendConcepts(ctx context.Context, req *ConceptRecommendationRequest) ([]*ConceptRecommendation, error) {
	// 获取知识图谱
	graph, err := s.knowledgeGraphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 获取学习者技能
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner skills: %w", err)
	}

	// 分析学习者状态
	learnerState := s.analyzeLearnerKnowledgeState(skills, graph)

	// 生成推荐
	recommendations := s.generateConceptRecommendations(ctx, graph, learner, learnerState, req)

	// 排序和过滤
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.MaxRecommendations {
		recommendations = recommendations[:req.MaxRecommendations]
	}

	return recommendations, nil
}

// analyzeStructure 分析图谱结构
func (s *KnowledgeGraphService) analyzeStructure(ctx context.Context, graph *entities.KnowledgeGraph) *StructuralMetrics {
	metrics := &StructuralMetrics{
		NodeCount:     len(graph.Nodes),
		RelationCount: len(graph.Relations),
		CentralityMeasures: make(map[uuid.UUID]float64),
		DifficultyDistribution: make(map[entities.DifficultyLevel]int),
		TypeDistribution: make(map[entities.NodeType]int),
	}

	// 计算密度
	if metrics.NodeCount > 1 {
		maxPossibleEdges := metrics.NodeCount * (metrics.NodeCount - 1)
		metrics.Density = float64(metrics.RelationCount) / float64(maxPossibleEdges)
	}

	// 计算平均路径长度
	metrics.AveragePathLength = s.calculateAveragePathLength(graph)

	// 计算聚类系数
	metrics.ClusteringCoefficient = s.calculateClusteringCoefficient(graph)

	// 计算中心性度量
	metrics.CentralityMeasures = s.calculateCentralityMeasures(graph)

	// 识别连通分量
	metrics.ConnectedComponents = s.countConnectedComponents(graph)

	// 识别关键路径
	metrics.CriticalPaths = s.identifyCriticalPaths(graph)

	// 识别瓶颈节点
	metrics.Bottlenecks = s.identifyBottlenecks(graph)

	// 统计难度分布
	for _, node := range graph.Nodes {
		metrics.DifficultyDistribution[node.DifficultyLevel]++
		metrics.TypeDistribution[node.Type]++
	}

	return metrics
}

// identifyLearningGaps 识别学习缺口
func (s *KnowledgeGraphService) identifyLearningGaps(ctx context.Context, graph *entities.KnowledgeGraph, learner *entities.Learner) []*LearningGap {
	var gaps []*LearningGap

	// 获取学习者技能
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learner.ID)
	if err != nil {
		return gaps
	}

	// 识别缺失的前置条件
	missingPrereqs := s.identifyMissingPrerequisites(graph, skills)
	for _, gap := range missingPrereqs {
		gaps = append(gaps, gap)
	}

	// 识别技能缺口
	skillGaps := s.identifySkillGaps(graph, skills, learner)
	for _, gap := range skillGaps {
		gaps = append(gaps, gap)
	}

	// 识别难度跳跃
	difficultyJumps := s.identifyDifficultyJumps(graph, learner)
	for _, gap := range difficultyJumps {
		gaps = append(gaps, gap)
	}

	// 按严重程度排序
	sort.Slice(gaps, func(i, j int) bool {
		return gaps[i].Severity > gaps[j].Severity
	})

	return gaps
}

// generateOptimizationSuggestions 生成优化建议
func (s *KnowledgeGraphService) generateOptimizationSuggestions(ctx context.Context, graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// 结构优化建议
	structuralSuggestions := s.generateStructuralOptimizations(graph)
	suggestions = append(suggestions, structuralSuggestions...)

	// 内容优化建议
	contentSuggestions := s.generateContentOptimizations(graph)
	suggestions = append(suggestions, contentSuggestions...)

	// 难度优化建议
	difficultySuggestions := s.generateDifficultyOptimizations(graph)
	suggestions = append(suggestions, difficultySuggestions...)

	// 可访问性优化建议
	accessibilitySuggestions := s.generateAccessibilityOptimizations(graph)
	suggestions = append(suggestions, accessibilitySuggestions...)

	// 按ROI排序
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].ROI > suggestions[j].ROI
	})

	return suggestions
}

// generateGraphRecommendations 生成图谱推荐
func (s *KnowledgeGraphService) generateGraphRecommendations(ctx context.Context, graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	// 添加节点推荐
	nodeRecommendations := s.generateNodeAdditionRecommendations(graph)
	recommendations = append(recommendations, nodeRecommendations...)

	// 添加关系推荐
	relationRecommendations := s.generateRelationRecommendations(graph)
	recommendations = append(recommendations, relationRecommendations...)

	// 修改难度推荐
	difficultyRecommendations := s.generateDifficultyModificationRecommendations(graph)
	recommendations = append(recommendations, difficultyRecommendations...)

	// 重构推荐
	restructureRecommendations := s.generateRestructureRecommendations(graph)
	recommendations = append(recommendations, restructureRecommendations...)

	// 按优先级和影响排序
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Priority != recommendations[j].Priority {
			return recommendations[i].Priority > recommendations[j].Priority
		}
		return recommendations[i].Impact > recommendations[j].Impact
	})

	return recommendations
}

// 辅助方法实现

func (s *KnowledgeGraphService) calculateAveragePathLength(graph *entities.KnowledgeGraph) float64 {
	if len(graph.Nodes) < 2 {
		return 0
	}

	totalLength := 0.0
	pathCount := 0

	// 简化实现：计算所有节点对之间的最短路径
	for i, nodeA := range graph.Nodes {
		for j, nodeB := range graph.Nodes {
			if i != j {
				// 这里应该使用实际的最短路径算法
				// 简化为直接连接的距离
				if s.areDirectlyConnected(nodeA.ID, nodeB.ID, graph) {
					totalLength += 1.0
				} else {
					totalLength += 2.0 // 假设平均需要2跳
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

		// 计算邻居之间的连接数
		connections := 0
		for i, neighborA := range neighbors {
			for j, neighborB := range neighbors {
				if i != j && s.areDirectlyConnected(neighborA, neighborB, graph) {
					connections++
				}
			}
		}

		// 计算聚类系数
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

	// 计算度中心性
	for _, node := range graph.Nodes {
		degree := 0
		for _, relation := range graph.Relations {
			if relation.FromNodeID == node.ID || relation.ToNodeID == node.ID {
				degree++
			}
		}
		centrality[node.ID] = float64(degree)
	}

	// 归一化
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

	// 识别入度为0的节点（起始节点）
	inDegree := make(map[uuid.UUID]int)
	for _, node := range graph.Nodes {
		inDegree[node.ID] = 0
	}

	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypePrerequisite {
			inDegree[relation.ToNodeID]++
		}
	}

	// 从每个起始节点找到最长路径
	for _, node := range graph.Nodes {
		if inDegree[node.ID] == 0 {
			path := s.findLongestPath(node.ID, graph, make(map[uuid.UUID]bool))
			if len(path) > 3 { // 只考虑较长的路径
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

	// 找到所有后继节点
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

	// 计算每个节点的介数中心性
	betweennessCentrality := s.calculateBetweennessCentrality(graph)

	// 找到介数中心性最高的节点
	threshold := 0.7 // 阈值
	for nodeID, centrality := range betweennessCentrality {
		if centrality > threshold {
			bottlenecks = append(bottlenecks, nodeID)
		}
	}

	return bottlenecks
}

func (s *KnowledgeGraphService) calculateBetweennessCentrality(graph *entities.KnowledgeGraph) map[uuid.UUID]float64 {
	centrality := make(map[uuid.UUID]float64)

	// 简化实现：基于度数计算
	for _, node := range graph.Nodes {
		degree := len(s.getNeighbors(node.ID, graph))
		centrality[node.ID] = float64(degree) / float64(len(graph.Nodes)-1)
	}

	return centrality
}

func (s *KnowledgeGraphService) identifyGeneralGaps(ctx context.Context, graph *entities.KnowledgeGraph) []*LearningGap {
	var gaps []*LearningGap

	// 识别孤立节点
	isolatedNodes := s.findIsolatedNodes(graph)
	for _, nodeID := range isolatedNodes {
		gap := &LearningGap{
			ID:          uuid.New(),
			Type:        "isolated_node",
			Description: "发现孤立的学习节点，缺乏与其他概念的连接",
			Severity:    0.8,
			AffectedNodes: []uuid.UUID{nodeID},
			Impact:      "学习者可能无法理解该概念与其他知识的关系",
			Priority:    3,
			EstimatedEffort: time.Hour * 2,
		}
		gaps = append(gaps, gap)
	}

	// 识别难度跳跃
	difficultyJumps := s.findDifficultyJumps(graph)
	for _, jump := range difficultyJumps {
		gap := &LearningGap{
			ID:          uuid.New(),
			Type:        "difficulty_jump",
			Description: "发现难度跳跃过大的学习路径",
			Severity:    0.6,
			AffectedNodes: jump,
			Impact:      "学习者可能在难度跳跃处遇到困难",
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

	// 查找相邻节点间难度差异过大的情况
	for _, relation := range graph.Relations {
		if relation.Type == entities.RelationTypePrerequisite {
			fromNode := s.findNodeByID(relation.FromNodeID, graph)
			toNode := s.findNodeByID(relation.ToNodeID, graph)

			if fromNode != nil && toNode != nil {
				difficultyDiff := int(toNode.DifficultyLevel) - int(fromNode.DifficultyLevel)
				if difficultyDiff > 2 { // 难度跳跃超过2级
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
		// 检查节点的前置条件是否满足
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
				Description: fmt.Sprintf("节点 %s 缺少必要的前置条件", node.Name),
				Severity:    0.9,
				AffectedNodes: []uuid.UUID{node.ID},
				SuggestedNodes: missingPrereqs,
				Impact:      "学习者可能无法理解该概念",
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

	// 分析学习者目标与当前技能的差距
	for _, goal := range learner.LearningGoals {
		if !goal.IsActive {
			continue
		}

		// 找到与目标相关的节点
		relatedNodes := s.findNodesForSkill(graph, goal.TargetSkill)
		for _, node := range relatedNodes {
			if !s.hasRequiredSkills(node, skills) {
				gap := &LearningGap{
					ID:          uuid.New(),
					Type:        "skill_gap",
					Description: fmt.Sprintf("缺少技能: %s", goal.TargetSkill),
					Severity:    0.7,
					AffectedNodes: []uuid.UUID{node.ID},
					Impact:      "无法达成学习目标",
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

	// 基于学习者经验水平识别难度跳跃
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
						Description: "难度跳跃过大，可能导致学习困难",
						Severity:    0.6,
						AffectedNodes: []uuid.UUID{fromNode.ID, toNode.ID},
						Impact:      "学习者可能在此处遇到挫折",
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

// 质量评分方法

func (s *KnowledgeGraphService) calculateStructuralQuality(metrics *StructuralMetrics) float64 {
	score := 0.0

	// 密度评分 (0.3权重)
	densityScore := math.Min(metrics.Density * 2, 1.0) // 理想密度约0.5
	score += densityScore * 0.3

	// 连通性评分 (0.3权重)
	connectivityScore := 1.0 / float64(metrics.ConnectedComponents)
	score += connectivityScore * 0.3

	// 平衡性评分 (0.2权重)
	balanceScore := 1.0 - math.Abs(metrics.ClusteringCoefficient - 0.5) * 2
	score += math.Max(0, balanceScore) * 0.2

	// 复杂度评分 (0.2权重)
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
	return math.Min(avgROI / 10.0, 1.0) // 假设最大ROI为10
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

// 洞察和警告生成

func (s *KnowledgeGraphService) generateInsights(result *GraphAnalysisResult) []string {
	var insights []string

	if result.StructuralMetrics != nil {
		if result.StructuralMetrics.Density > 0.7 {
			insights = append(insights, "知识图谱连接密度较高，概念之间关联性强")
		}
		if result.StructuralMetrics.ConnectedComponents == 1 {
			insights = append(insights, "知识图谱结构完整，所有概念都有学习路径")
		}
		if len(result.StructuralMetrics.Bottlenecks) > 0 {
			insights = append(insights, fmt.Sprintf("发现%d个关键节点，是学习路径的重要枢纽", len(result.StructuralMetrics.Bottlenecks)))
		}
	}

	if result.QualityScore > 0.8 {
		insights = append(insights, "知识图谱整体质量优秀，适合学习使用")
	} else if result.QualityScore > 0.6 {
		insights = append(insights, "知识图谱质量良好，有一定优化空间")
	}

	return insights
}

func (s *KnowledgeGraphService) generateWarnings(result *GraphAnalysisResult) []string {
	var warnings []string

	if result.StructuralMetrics != nil {
		if result.StructuralMetrics.ConnectedComponents > 1 {
			warnings = append(warnings, "知识图谱存在孤立的概念群，可能影响学习连贯性")
		}
		if result.StructuralMetrics.Density < 0.2 {
			warnings = append(warnings, "知识图谱连接稀疏，概念之间缺乏足够的关联")
		}
		if result.StructuralMetrics.AveragePathLength > 8 {
			warnings = append(warnings, "平均学习路径过长，可能影响学习效率")
		}
	}

	if len(result.LearningGaps) > 5 {
		warnings = append(warnings, "发现较多学习缺口，建议优先解决高优先级问题")
	}

	if result.QualityScore < 0.4 {
		warnings = append(warnings, "知识图谱质量较低，建议进行全面优化")
	}

	return warnings
}

// 优化建议生成方法

func (s *KnowledgeGraphService) generateStructuralOptimizations(graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// 检查连通性
	components := s.countConnectedComponents(graph)
	if components > 1 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "structure",
			Title:       "改善图谱连通性",
			Description: "添加连接以减少孤立的概念群",
			Benefits:    []string{"提高学习路径的连贯性", "减少学习死角"},
			Risks:       []string{"可能增加复杂度"},
			Complexity:  "medium",
			ROI:         8.0,
		})
	}

	// 检查密度
	density := s.calculateDensity(graph)
	if density < 0.2 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "structure",
			Title:       "增加概念关联",
			Description: "添加更多概念之间的关系连接",
			Benefits:    []string{"增强知识网络", "提供更多学习路径"},
			Risks:       []string{"可能造成信息过载"},
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

	// 检查内容覆盖度
	suggestions = append(suggestions, &OptimizationSuggestion{
		ID:          uuid.New(),
		Category:    "content",
		Title:       "丰富学习内容",
		Description: "为关键概念添加多样化的学习材料",
		Benefits:    []string{"适应不同学习风格", "提高学习效果"},
		Risks:       []string{"增加维护成本"},
		Complexity:  "medium",
		ROI:         7.0,
	})

	return suggestions
}

func (s *KnowledgeGraphService) generateDifficultyOptimizations(graph *entities.KnowledgeGraph) []*OptimizationSuggestion {
	var suggestions []*OptimizationSuggestion

	// 检查难度分布
	difficultyJumps := s.findDifficultyJumps(graph)
	if len(difficultyJumps) > 0 {
		suggestions = append(suggestions, &OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "difficulty",
			Title:       "平滑难度过渡",
			Description: "在难度跳跃处添加过渡性概念",
			Benefits:    []string{"降低学习难度", "提高完成率"},
			Risks:       []string{"可能延长学习时间"},
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
		Title:       "提高可访问性",
		Description: "添加多语言支持和无障碍功能",
		Benefits:    []string{"扩大用户群体", "提高包容性"},
		Risks:       []string{"增加开发成本"},
		Complexity:  "high",
		ROI:         5.0,
	})

	return suggestions
}

// 推荐生成方法

func (s *KnowledgeGraphService) generateNodeAdditionRecommendations(graph *entities.KnowledgeGraph) []*GraphRecommendation {
	var recommendations []*GraphRecommendation

	// 识别缺失的中间概念
	recommendations = append(recommendations, &GraphRecommendation{
		ID:          uuid.New(),
		Type:        "add_node",
		Title:       "添加过渡概念",
		Description: "在难度跳跃处添加中间概念",
		Rationale:   "帮助学习者更好地理解复杂概念",
		Priority:    3,
		Confidence:  0.8,
		Impact:      0.7,
		Effort:      0.6,
		Actions: []RecommendationAction{
			{
				Type:        "create_node",
				Description: "创建新的知识节点",
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
		Title:       "添加概念关联",
		Description: "连接相关但未关联的概念",
		Rationale:   "增强知识网络的完整性",
		Priority:    2,
		Confidence:  0.7,
		Impact:      0.6,
		Effort:      0.3,
		Actions: []RecommendationAction{
			{
				Type:        "create_relation",
				Description: "创建新的概念关系",
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
		Title:       "调整难度级别",
		Description: "重新评估和调整概念的难度级别",
		Rationale:   "确保难度级别准确反映学习要求",
		Priority:    2,
		Confidence:  0.6,
		Impact:      0.5,
		Effort:      0.4,
		Actions: []RecommendationAction{
			{
				Type:        "update_difficulty",
				Description: "更新节点难度级别",
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
		Title:       "优化图谱结构",
		Description: "重新组织概念层次和关系",
		Rationale:   "提高学习路径的效率和清晰度",
		Priority:    1,
		Confidence:  0.5,
		Impact:      0.9,
		Effort:      0.8,
		Actions: []RecommendationAction{
			{
				Type:        "reorganize_hierarchy",
				Description: "重新组织概念层次结构",
				Parameters: map[string]interface{}{
					"strategy": "bottom_up",
				},
			},
		},
	})

	return recommendations
}

// 概念推荐相关方法

func (s *KnowledgeGraphService) analyzeLearnerKnowledgeState(skills map[string]*entities.SkillLevel, graph *entities.KnowledgeGraph) map[uuid.UUID]float64 {
	nodeReadiness := make(map[uuid.UUID]float64)

	for _, node := range graph.Nodes {
		readiness := s.calculateNodeReadiness(&node, skills, graph)
		nodeReadiness[node.ID] = readiness
	}

	return nodeReadiness
}

func (s *KnowledgeGraphService) calculateNodeReadiness(node *entities.KnowledgeNode, skills map[string]*entities.SkillLevel, graph *entities.KnowledgeGraph) float64 {
	// 检查前置条件满足度
	prerequisites := s.getPrerequisites(node.ID, graph)
	if len(prerequisites) == 0 {
		return 1.0 // 没有前置条件，完全准备好
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

	// 生成不同类型的推荐
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

	// 找到准备度最高的节点
	for nodeID, readiness := range learnerState {
		if readiness > 0.8 { // 高准备度
			node := s.findNodeByID(nodeID, graph)
			if node != nil {
				recommendation := &ConceptRecommendation{
					NodeID:             nodeID,
					RecommendationType: "next",
					Score:              readiness * 0.9,
					Confidence:         0.8,
					Reasoning:          []string{"前置条件已满足", "适合当前学习水平"},
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

	// 找到缺失的前置条件
	for nodeID, readiness := range learnerState {
		if readiness < 0.5 && readiness > 0 { // 部分准备但不足
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
							Reasoning:          []string{"需要掌握此前置概念", "有助于理解后续内容"},
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

	// 找到相关概念
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
					Reasoning:          []string{"与当前学习内容相关", "有助于拓展知识面"},
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

	// 找到高级概念
	for nodeID, readiness := range learnerState {
		node := s.findNodeByID(nodeID, graph)
		if node != nil && node.DifficultyLevel > entities.DifficultyLevel(learner.ExperienceLevel) {
			if readiness > 0.7 {
				recommendation := &ConceptRecommendation{
					NodeID:             nodeID,
					RecommendationType: "advanced",
					Score:              readiness * 0.6,
					Confidence:         0.5,
					Reasoning:          []string{"挑战性内容", "有助于提升技能水平"},
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