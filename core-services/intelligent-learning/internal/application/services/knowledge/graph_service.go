package knowledge

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// KnowledgeGraphService 知识图谱服务接口
type KnowledgeGraphService interface {
	AnalyzeGraph(ctx context.Context, req *domainServices.GraphAnalysisRequest) (*domainServices.GraphAnalysisResult, error)
	RecommendConcepts(ctx context.Context, req *domainServices.ConceptRecommendationRequest) ([]*domainServices.ConceptRecommendation, error)
}






// KnowledgeGraphAppService 知识图谱应用服务
type KnowledgeGraphAppService struct {
	graphRepo           repositories.KnowledgeGraphRepository
	contentRepo         repositories.LearningContentRepository
	learnerRepo         repositories.LearnerRepository
	graphService        KnowledgeGraphService
	learningPathService *domainServices.LearningPathService
	defaultGraphID      uuid.UUID // 默认图谱ID
}

// NewKnowledgeGraphAppService 创建新的知识图谱应用服务
func NewKnowledgeGraphAppService(
	graphRepo repositories.KnowledgeGraphRepository,
	contentRepo repositories.LearningContentRepository,
	learnerRepo repositories.LearnerRepository,
	graphService KnowledgeGraphService,
	learningPathService *domainServices.LearningPathService,
	defaultGraphID uuid.UUID,
) *KnowledgeGraphAppService {
	return &KnowledgeGraphAppService{
		graphRepo:           graphRepo,
		contentRepo:         contentRepo,
		learnerRepo:         learnerRepo,
		graphService:        graphService,
		learningPathService: learningPathService,
		defaultGraphID:      defaultGraphID,
	}
}

// CreateNodeRequest 创建节点请求
type CreateNodeRequest struct {
	Name        string            `json:"name" validate:"required,min=2,max=100"`
	Type        string            `json:"type" validate:"required"`
	Description string            `json:"description" validate:"max=500"`
	Difficulty  string            `json:"difficulty" validate:"required"`
	Properties  map[string]string `json:"properties"`
	Tags        []string          `json:"tags"`
}

// CreateRelationRequest 创建关系请求
type CreateRelationRequest struct {
	FromNodeID  uuid.UUID         `json:"from_node_id" validate:"required"`
	ToNodeID    uuid.UUID         `json:"to_node_id" validate:"required"`
	Type        string            `json:"type" validate:"required"`
	Weight      float64           `json:"weight" validate:"min=0,max=1"`
	Properties  map[string]string `json:"properties"`
	Description string            `json:"description"`
}

// UpdateNodeRequest 更新节点请求
type UpdateNodeRequest struct {
	Title      *string           `json:"title,omitempty"`
	Content    *string           `json:"content,omitempty"`
	Type       *string           `json:"type,omitempty"`
	Difficulty *string           `json:"difficulty,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// UpdateRelationRequest 更新关系请求
type UpdateRelationRequest struct {
	Type        *string           `json:"type,omitempty"`
	Weight      *float64          `json:"weight,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
	Description *string           `json:"description,omitempty"`
}

// NodeResponse 节点响应
type NodeResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Difficulty  string            `json:"difficulty"`
	Properties  map[string]string `json:"properties"`
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// RelationResponse 关系响应
type RelationResponse struct {
	ID          uuid.UUID         `json:"id"`
	FromNodeID  uuid.UUID         `json:"from_node_id"`
	ToNodeID    uuid.UUID         `json:"to_node_id"`
	Type        string            `json:"type"`
	Weight      float64           `json:"weight"`
	Properties  map[string]string `json:"properties"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// GraphSearchRequest 图谱搜索请求
type GraphSearchRequest struct {
	Query      string   `json:"query"`
	NodeTypes  []string `json:"node_types"`
	Tags       []string `json:"tags"`
	Difficulty string   `json:"difficulty"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

// GraphSearchResponse 图谱搜索响应
type GraphSearchResponse struct {
	Nodes   []*NodeResponse `json:"nodes"`
	Total   int             `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
}

// LearningPathRequest 学习路径请求
type LearningPathRequest struct {
	LearnerID      uuid.UUID `json:"learner_id" validate:"required"`
	StartNodeID    uuid.UUID `json:"start_node_id" validate:"required"`
	TargetNodeID   uuid.UUID `json:"target_node_id" validate:"required"`
	MaxDepth       int       `json:"max_depth"`
	PathType       string    `json:"path_type"` // shortest, comprehensive, adaptive
	Constraints    []string  `json:"constraints"`
	PreferredTypes []string  `json:"preferred_types"`
}

// LearningPathResponse 学习路径响应
type LearningPathResponse struct {
	ID                uuid.UUID                     `json:"id"`
	LearnerID         uuid.UUID                     `json:"learner_id"`
	Title             string                        `json:"title"`
	Description       string                        `json:"description"`
	EstimatedDuration int                           `json:"estimated_duration"`
	DifficultyLevel   string                        `json:"difficulty_level"`
	Progress          float64                       `json:"progress"`
	Status            string                        `json:"status"`
	Steps             []LearningPathStepResponse    `json:"steps"`
	Milestones        []LearningMilestoneResponse   `json:"milestones"`
	CreatedAt         time.Time                     `json:"created_at"`
	UpdatedAt         time.Time                     `json:"updated_at"`
}

// LearningPathStepResponse 学习路径步骤响应
type LearningPathStepResponse struct {
	ID              uuid.UUID `json:"id"`
	Order           int       `json:"order"`
	ContentID       uuid.UUID `json:"content_id"`
	ContentTitle    string    `json:"content_title"`
	ContentType     string    `json:"content_type"`
	EstimatedTime   int       `json:"estimated_time"`
	Prerequisites   []string  `json:"prerequisites"`
	LearningGoals   []string  `json:"learning_goals"`
	IsCompleted     bool      `json:"is_completed"`
	CompletionRate  float64   `json:"completion_rate"`
}

// LearningMilestoneResponse 学习里程碑响�?
type LearningMilestoneResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	TargetStep  int        `json:"target_step"`
	IsAchieved  bool       `json:"is_achieved"`
	AchievedAt  *time.Time `json:"achieved_at,omitempty"`
	Reward      string     `json:"reward"`
}

// PathMilestone 路径里程�?
type PathMilestone struct {
	ID          uuid.UUID `json:"id"`
	NodeID      uuid.UUID `json:"node_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	IsRequired  bool      `json:"is_required"`
}

// PathAdaptation 路径适应
type PathAdaptation struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Reason     string    `json:"reason"`
	Suggestion string    `json:"suggestion"`
	Priority   int       `json:"priority"`
}

// ConceptMapRequest 概念图请�?
type ConceptMapRequest struct {
	CenterNodeID uuid.UUID `json:"center_node_id" validate:"required"`
	Depth        int       `json:"depth"`
	MaxNodes     int       `json:"max_nodes"`
	IncludeTypes []string  `json:"include_types"`
	ExcludeTypes []string  `json:"exclude_types"`
}

// ConceptMapResponse 概念图响�?
type ConceptMapResponse struct {
	CenterNode *NodeResponse       `json:"center_node"`
	Nodes      []*NodeResponse     `json:"nodes"`
	Relations  []*RelationResponse `json:"relations"`
	Layout     *GraphLayout        `json:"layout"`
}

// GraphLayout 图布局
type GraphLayout struct {
	Positions map[uuid.UUID]*domainServices.Position `json:"positions"`
	Width     float64                 `json:"width"`
	Height    float64                 `json:"height"`
}

// Position 位置
type GraphPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}



// CreateNode 创建知识节点
func (s *KnowledgeGraphAppService) CreateNode(ctx context.Context, req *CreateNodeRequest) (*NodeResponse, error) {
	// 将字符串难度转换为DifficultyLevel
	var difficultyLevel entities.DifficultyLevel
	switch req.Difficulty {
	case "beginner":
		difficultyLevel = entities.DifficultyBeginner
	case "elementary":
		difficultyLevel = entities.DifficultyElementary
	case "intermediate":
		difficultyLevel = entities.DifficultyIntermediate
	case "advanced":
		difficultyLevel = entities.DifficultyAdvanced
	case "expert":
		difficultyLevel = entities.DifficultyExpert
	default:
		difficultyLevel = entities.DifficultyBeginner
	}

	// 将字符串类型转换为NodeType
	var nodeType entities.NodeType
	switch req.Type {
	case "concept":
		nodeType = entities.NodeTypeConcept
	case "skill":
		nodeType = entities.NodeTypeSkill
	case "topic":
		nodeType = entities.NodeTypeTopic
	case "subject":
		nodeType = entities.NodeTypeSubject
	case "learning_path":
		nodeType = entities.NodeTypeLearningPath
	case "content":
		nodeType = entities.NodeTypeContent
	default:
		nodeType = entities.NodeTypeConcept
	}

	node := &entities.KnowledgeNode{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		Type:            nodeType,
		DifficultyLevel: difficultyLevel,
		Tags:            req.Tags,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.graphRepo.AddNode(ctx, s.defaultGraphID, node); err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	return s.buildNodeResponse(node), nil
}

// GetNode 获取知识节点
func (s *KnowledgeGraphAppService) GetNode(ctx context.Context, nodeID uuid.UUID) (*NodeResponse, error) {
	node, err := s.graphRepo.GetNode(ctx, s.defaultGraphID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return s.buildNodeResponse(node), nil
}

// UpdateNode 更新知识节点
func (s *KnowledgeGraphAppService) UpdateNode(ctx context.Context, nodeID uuid.UUID, updates map[string]interface{}) (*NodeResponse, error) {
	// 首先获取现有节点以进行更�?
	existingNode, err := s.graphRepo.GetNode(ctx, s.defaultGraphID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing node: %w", err)
	}

	// �?map 转换�?UpdateNodeRequest
	updateReq := s.mapToUpdateNodeRequest(updates)

	// 应用更新到节�?
	updatedNode := s.applyUpdatesToNode(existingNode, updateReq)

	if err := s.graphRepo.UpdateNode(ctx, s.defaultGraphID, updatedNode); err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	node, err := s.graphRepo.GetNode(ctx, s.defaultGraphID, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated node: %w", err)
	}

	return s.buildNodeResponse(node), nil
}

// DeleteNode 删除知识节点
func (s *KnowledgeGraphAppService) DeleteNode(ctx context.Context, nodeID uuid.UUID) error {
	if err := s.graphRepo.RemoveNode(ctx, s.defaultGraphID, nodeID); err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}
	return nil
}

// CreateRelation 创建知识关系
func (s *KnowledgeGraphAppService) CreateRelation(ctx context.Context, req *CreateRelationRequest) (*RelationResponse, error) {
	relation := entities.NewKnowledgeRelation(req.FromNodeID, req.ToNodeID, entities.RelationType(req.Type), req.Weight)
	// �?Properties 转换�?Metadata
	if req.Properties != nil {
		metadata := make(map[string]interface{})
		for k, v := range req.Properties {
			metadata[k] = v
		}
		relation.Metadata = metadata
	}
	relation.Description = req.Description

	if err := s.graphRepo.AddRelation(ctx, s.defaultGraphID, relation); err != nil {
		return nil, fmt.Errorf("failed to create relation: %w", err)
	}

	return s.buildRelationResponse(relation), nil
}

// GetRelation 获取知识关系
func (s *KnowledgeGraphAppService) GetRelation(ctx context.Context, relationID uuid.UUID) (*RelationResponse, error) {
	relation, err := s.graphRepo.GetRelation(ctx, s.defaultGraphID, relationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation: %w", err)
	}

	return s.buildRelationResponse(relation), nil
}

// DeleteRelation 删除知识关系
func (s *KnowledgeGraphAppService) DeleteRelation(ctx context.Context, relationID uuid.UUID) error {
	if err := s.graphRepo.RemoveRelation(ctx, s.defaultGraphID, relationID); err != nil {
		return fmt.Errorf("failed to delete relation: %w", err)
	}
	return nil
}

// UpdateRelation 更新知识关系
func (s *KnowledgeGraphAppService) UpdateRelation(ctx context.Context, relationID uuid.UUID, req *UpdateRelationRequest) (*RelationResponse, error) {
	// 首先获取现有关系
	existingRelation, err := s.graphRepo.GetRelation(ctx, s.defaultGraphID, relationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing relation: %w", err)
	}

	// 应用更新
	updatedRelation := s.applyUpdatesToRelation(existingRelation, req)

	// 更新关系
	if err := s.graphRepo.UpdateRelation(ctx, s.defaultGraphID, updatedRelation); err != nil {
		return nil, fmt.Errorf("failed to update relation: %w", err)
	}

	// 获取更新后的关系
	relation, err := s.graphRepo.GetRelation(ctx, s.defaultGraphID, relationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated relation: %w", err)
	}

	return s.buildRelationResponse(relation), nil
}

// SearchNodes 搜索知识节点
func (s *KnowledgeGraphAppService) SearchNodes(ctx context.Context, req *GraphSearchRequest) (*GraphSearchResponse, error) {
	var nodeType *entities.NodeType
	if len(req.NodeTypes) > 0 {
		// 取第一个节点类�?
		nt := entities.NodeType(req.NodeTypes[0])
		nodeType = &nt
	}

	var difficultyLevel *entities.DifficultyLevel
	if req.Difficulty != "" {
		dl := s.stringToDifficultyLevel(req.Difficulty)
		difficultyLevel = &dl
	}

	query := &repositories.NodeSearchQuery{
		Query:           req.Query,
		NodeType:        nodeType,
		Tags:            req.Tags,
		DifficultyLevel: difficultyLevel,
		Limit:           req.Limit,
		Offset:          req.Offset,
	}

	nodes, total, err := s.graphRepo.SearchNodes(ctx, s.defaultGraphID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search nodes: %w", err)
	}

	nodeResponses := make([]*NodeResponse, len(nodes))
	for i, node := range nodes {
		nodeResponses[i] = s.buildNodeResponse(node)
	}

	return &GraphSearchResponse{
		Nodes:   nodeResponses,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: req.Offset+len(nodes) < total,
	}, nil
}

// GetNodeNeighbors 获取节点邻居
func (s *KnowledgeGraphAppService) GetNodeNeighbors(ctx context.Context, nodeID uuid.UUID, depth int) ([]*NodeResponse, []*RelationResponse, error) {
	neighbors, err := s.graphRepo.GetNodeNeighbors(ctx, s.defaultGraphID, nodeID, depth)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get node neighbors: %w", err)
	}

	nodeResponses := make([]*NodeResponse, len(neighbors))
	for i, node := range neighbors {
		nodeResponses[i] = s.buildNodeResponse(node)
	}

	// 获取节点的关�?
	relations, err := s.graphRepo.GetNodeRelations(ctx, s.defaultGraphID, nodeID, nil)
	if err != nil {
		return nodeResponses, nil, fmt.Errorf("failed to get node relations: %w", err)
	}

	relationResponses := make([]*RelationResponse, len(relations))
	for i, relation := range relations {
		relationResponses[i] = s.buildRelationResponse(relation)
	}

	return nodeResponses, relationResponses, nil
}

// FindShortestPath 查找最短路�?
func (s *KnowledgeGraphAppService) FindShortestPath(ctx context.Context, fromNodeID, toNodeID uuid.UUID) ([]*NodeResponse, []*RelationResponse, error) {
	path, err := s.graphRepo.FindShortestPath(ctx, s.defaultGraphID, fromNodeID, toNodeID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find shortest path: %w", err)
	}

	nodeResponses := make([]*NodeResponse, len(path))
	for i, node := range path {
		nodeResponses[i] = s.buildNodeResponse(node)
	}

	// 获取路径中相邻节点之间的关系
	var relationResponses []*RelationResponse
	for i := 0; i < len(path)-1; i++ {
		relations, err := s.graphRepo.GetRelationsBetween(ctx, s.defaultGraphID, path[i].ID, path[i+1].ID)
		if err != nil {
			continue // 跳过无法获取的关�?
		}
		for _, relation := range relations {
			relationResponses = append(relationResponses, s.buildRelationResponse(relation))
		}
	}

	return nodeResponses, relationResponses, nil
}

// GenerateLearningPath 生成学习路径
func (s *KnowledgeGraphAppService) GenerateLearningPath(ctx context.Context, req *LearningPathRequest) (*LearningPathResponse, error) {
	// 获取学习者信�?
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 构建路径推荐请求
	pathReq := &domainServices.PathRecommendationRequest{
		LearnerID:    req.LearnerID,
		GraphID:      s.defaultGraphID,
		TargetNodeID: req.TargetNodeID,
		MaxPaths:     5, // 默认生成5条路�?
		Preferences: &domainServices.PathPreferences{
			MaxPathLength:      req.MaxDepth,
			PreferShortPaths:   req.PathType == "shortest",
			AdaptiveDifficulty: req.PathType == "adaptive",
		},
	}

	// 根据学习者偏好调整路径偏�?
	if learner.Preferences.DifficultyTolerance > 0 {
		// 根据学习者的难度容忍度调整适应性难�?
		pathReq.Preferences.AdaptiveDifficulty = learner.Preferences.DifficultyTolerance > 0.5
	}

	// 生成个性化路径
	personalizedPaths, err := s.learningPathService.RecommendPersonalizedPaths(ctx, pathReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate learning path: %w", err)
	}

	if len(personalizedPaths) == 0 {
		return nil, fmt.Errorf("no learning paths generated")
	}

	// 使用第一个推荐路�?
	personalizedPath := personalizedPaths[0]

	// 构建响应
	response := &LearningPathResponse{
		ID:                personalizedPath.Path.ID,
		LearnerID:         req.LearnerID,
		Title:             personalizedPath.Path.Name,
		Description:       personalizedPath.Path.Description,
		EstimatedDuration: int(personalizedPath.EstimatedDuration.Hours()),
		DifficultyLevel:   s.difficultyLevelToString(personalizedPath.Path.DifficultyLevel),
		Progress:          0.0, // 新路径进度为0
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// 构建学习步骤
	steps := make([]LearningPathStepResponse, len(personalizedPath.Path.Nodes))
	for i, pathNode := range personalizedPath.Path.Nodes {
		steps[i] = LearningPathStepResponse{
			ID:              pathNode.ID,
			Order:           pathNode.Order,
			ContentID:       pathNode.KnowledgeID,
			ContentTitle:    fmt.Sprintf("学习节点 %d", pathNode.Order+1),
			ContentType:     "knowledge_node",
			EstimatedTime:   60, // 默认60分钟
			Prerequisites:   []string{},
			LearningGoals:   []string{},
			IsCompleted:     false,
			CompletionRate:  0.0,
		}
	}
	response.Steps = steps

	// 构建里程�?
	milestones := make([]LearningMilestoneResponse, len(personalizedPath.Milestones))
	for i, milestone := range personalizedPath.Milestones {
		milestones[i] = LearningMilestoneResponse{
			ID:          uuid.New(),
			Title:       fmt.Sprintf("里程�?%d", milestone.Position+1),
			Description: fmt.Sprintf("学习里程碑位�?%d", milestone.Position),
			TargetStep:  milestone.Position,
			IsAchieved:  false,
			AchievedAt:  nil,
			Reward:      "",
		}
	}
	response.Milestones = milestones

	return response, nil
}

// GenerateConceptMap 生成概念�?
func (s *KnowledgeGraphAppService) GenerateConceptMap(ctx context.Context, req *ConceptMapRequest) (*ConceptMapResponse, error) {
	// 获取中心节点
	centerNode, err := s.graphRepo.GetNode(ctx, s.defaultGraphID, req.CenterNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get center node: %w", err)
	}

	// 获取概念图数�?
	conceptMap, err := s.graphRepo.GetConceptMap(ctx, req.CenterNodeID, req.Depth, req.MaxNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept map: %w", err)
	}

	// 构建节点响应
	nodeResponses := make([]*NodeResponse, len(conceptMap.Nodes))
	for i, node := range conceptMap.Nodes {
		nodeResponses[i] = s.buildNodeResponse(&node)
	}

	// 构建关系响应
	relationResponses := make([]*RelationResponse, len(conceptMap.Relations))
	for i, relation := range conceptMap.Relations {
		relationResponses[i] = s.buildRelationResponse(&relation)
	}

	// 转换为指针切片用于布局生成
	nodePointers := make([]*entities.KnowledgeNode, len(conceptMap.Nodes))
	for i := range conceptMap.Nodes {
		nodePointers[i] = &conceptMap.Nodes[i]
	}
	
	relationPointers := make([]*entities.KnowledgeRelation, len(conceptMap.Relations))
	for i := range conceptMap.Relations {
		relationPointers[i] = &conceptMap.Relations[i]
	}

	// 生成布局
	layout := s.generateGraphLayout(nodePointers, relationPointers)

	return &ConceptMapResponse{
		CenterNode: s.buildNodeResponse(centerNode),
		Nodes:      nodeResponses,
		Relations:  relationResponses,
		Layout:     layout,
	}, nil
}

// AnalyzeGraph 分析图谱
func (s *KnowledgeGraphAppService) AnalyzeGraph(ctx context.Context, req *domainServices.GraphAnalysisRequest) (interface{}, error) {
	switch req.AnalysisType {
	case "structure":
		return s.analyzeGraphStructure(ctx, req)
	case "gaps":
		return s.analyzeLearningGaps(ctx, req)
	case "optimization":
		return s.analyzeOptimization(ctx, req)
	case "recommendations":
		return s.generateRecommendations(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported analysis type: %s", req.AnalysisType)
	}
}

// GetGraphStatistics 获取图谱统计信息
func (s *KnowledgeGraphAppService) GetGraphStatistics(ctx context.Context) (*repositories.GraphStatistics, error) {
	entityStats, err := s.graphRepo.GetGraphStatistics(ctx, s.defaultGraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph statistics: %w", err)
	}
	
	// 转换�?repositories.GraphStatistics 类型
	stats := &repositories.GraphStatistics{
		NodeCount:       entityStats.NodeCount,
		RelationCount:   entityStats.RelationCount,
		PathCount:       entityStats.PathCount,
		NodesByType:     entityStats.NodesByType,
		RelationsByType: entityStats.RelationsByType,
		AvgDegree:       entityStats.AvgDegree,
		Density:         entityStats.Density,
		LastUpdated:     entityStats.LastUpdated,
	}
	
	return stats, nil
}

// ValidateGraph 验证图谱
func (s *KnowledgeGraphAppService) ValidateGraph(ctx context.Context) (*repositories.GraphValidation, error) {
	validation, err := s.graphRepo.ValidateGraphStructure(ctx, s.defaultGraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate graph: %w", err)
	}
	return validation, nil
}

// ExportGraph 导出图谱
func (s *KnowledgeGraphAppService) ExportGraph(ctx context.Context, format string) ([]byte, error) {
	data, err := s.graphRepo.ExportGraph(ctx, s.defaultGraphID, format)
	if err != nil {
		return nil, fmt.Errorf("failed to export graph: %w", err)
	}
	return data, nil
}

// ImportGraph 导入图谱
func (s *KnowledgeGraphAppService) ImportGraph(ctx context.Context, data []byte, format string) error {
	_, err := s.graphRepo.ImportGraph(ctx, data, format)
	if err != nil {
		return fmt.Errorf("failed to import graph: %w", err)
	}
	return nil
}

// 私有辅助方法

func (s *KnowledgeGraphAppService) buildNodeResponse(node *entities.KnowledgeNode) *NodeResponse {
	return &NodeResponse{
		ID:          node.ID,
		Name:        node.Name,
		Type:        string(node.Type),
		Description: node.Description,
		Difficulty:  string(node.DifficultyLevel),
		Properties:  make(map[string]string), // 临时空映射，因为实体使用 Metadata
		Tags:        node.Tags,
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
	}
}

func (s *KnowledgeGraphAppService) buildRelationResponse(relation *entities.KnowledgeRelation) *RelationResponse {
	// �?Metadata 转换�?Properties
	properties := make(map[string]string)
	for k, v := range relation.Metadata {
		if str, ok := v.(string); ok {
			properties[k] = str
		} else {
			properties[k] = fmt.Sprintf("%v", v)
		}
	}
	
	return &RelationResponse{
		ID:          relation.ID,
		FromNodeID:  relation.FromNodeID,
		ToNodeID:    relation.ToNodeID,
		Type:        string(relation.Type),
		Weight:      relation.Weight,
		Properties:  properties,
		Description: relation.Description,
		CreatedAt:   relation.CreatedAt,
		UpdatedAt:   relation.UpdatedAt,
	}
}

func (s *KnowledgeGraphAppService) generateGraphLayout(nodes []*entities.KnowledgeNode, relations []*entities.KnowledgeRelation) *GraphLayout {
	positions := make(map[uuid.UUID]*domainServices.Position)

	// 简单的圆形布局算法
	nodeCount := len(nodes)
	if nodeCount == 0 {
		return &GraphLayout{
			Positions: positions,
			Width:     400,
			Height:    400,
		}
	}

	centerX, centerY := 200.0, 200.0
	radius := 150.0

	for i, node := range nodes {
		x := centerX + radius*float64(i%2+1)*0.5*(1.0+0.5*float64(i)/float64(nodeCount))*
			(1.0+0.3*float64(len(node.Name))/10.0)
		y := centerY + radius*float64((i+1)%3+1)*0.5*(1.0+0.5*float64(i)/float64(nodeCount))

		if i == 0 {
			x, y = centerX, centerY // 中心节点
		} else {
			x = centerX + radius*0.8*float64(1+i%3)*0.7
			y = centerY + radius*0.8*float64(1+(i+1)%3)*0.7
		}

		positions[node.ID] = &domainServices.Position{X: x, Y: y}
	}

	return &GraphLayout{
		Positions: positions,
		Width:     400,
		Height:    400,
	}
}

func (s *KnowledgeGraphAppService) analyzeGraphStructure(ctx context.Context, req *domainServices.GraphAnalysisRequest) (*domainServices.GraphAnalysisResult, error) {
	analysisReq := &domainServices.GraphAnalysisRequest{
		GraphID:      s.defaultGraphID,
		AnalysisType: "structure",
	}

	if req.LearnerID != nil {
		analysisReq.LearnerID = req.LearnerID
	}

	result, err := s.graphService.AnalyzeGraph(ctx, analysisReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze graph structure: %w", err)
	}

	return result, nil
}

func (s *KnowledgeGraphAppService) analyzeLearningGaps(ctx context.Context, req *domainServices.GraphAnalysisRequest) ([]*domainServices.LearningGap, error) {
	if req.LearnerID == nil {
		return nil, fmt.Errorf("learner ID is required for gap analysis")
	}

	// 获取知识图谱
	_, err := s.graphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	// 获取学习者信�?
	_, err = s.learnerRepo.GetByID(ctx, *req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 调用领域服务的方法来识别学习缺口
	// 创建领域服务的分析请�?
	domainReq := &domainServices.GraphAnalysisRequest{
		GraphID:      req.GraphID,
		AnalysisType: "learning_gaps",
		LearnerID:    req.LearnerID,
	}
	
	domainResult, err := s.graphService.AnalyzeGraph(ctx, domainReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze learning gaps: %w", err)
	}

	// 转换为应用服务的类型
	var gaps []*domainServices.LearningGap
	if domainResult != nil {
		// 临时实现：创建示例差�?
		for i := 0; i < 3; i++ {
			gap := &domainServices.LearningGap{
				ID:          uuid.New(),
				Type:        "skill_gap",
				Description: fmt.Sprintf("Learning gap %d identified", i+1),
				Severity:    0.6, // 中等严重程度
				AffectedNodes: []uuid.UUID{uuid.New()},
				SuggestedNodes: []uuid.UUID{uuid.New()},
				Impact:      "Learners may struggle with this concept",
				Priority:    2,
				EstimatedEffort: time.Hour * 2,
			}
			gaps = append(gaps, gap)
		}
	}

	return gaps, nil
}

func (s *KnowledgeGraphAppService) analyzeOptimization(ctx context.Context, req *domainServices.GraphAnalysisRequest) ([]*domainServices.OptimizationSuggestion, error) {
	// 获取知识图谱
	_, err := s.graphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	// 临时实现：创建示例优化建�?
	var suggestions []*domainServices.OptimizationSuggestion
	for i := 0; i < 3; i++ {
		suggestion := &domainServices.OptimizationSuggestion{
			ID:          uuid.New(),
			Category:    "structure",
			Title:       fmt.Sprintf("Optimization suggestion %d", i+1),
			Description: fmt.Sprintf("Optimization suggestion %d", i+1),
			Benefits:    []string{"Improved structure", "Better learning flow"},
			Risks:       []string{"Potential complexity"},
			Complexity:  "medium",
			ROI:         0.7,
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

func (s *KnowledgeGraphAppService) generateRecommendations(ctx context.Context, req *domainServices.GraphAnalysisRequest) ([]*domainServices.ConceptRecommendation, error) {
	if req.LearnerID == nil {
		return nil, fmt.Errorf("learner ID is required for recommendations")
	}

	// 临时实现：创建示例推�?
	var recommendations []*domainServices.ConceptRecommendation
	for i := 0; i < 5; i++ {
		recommendation := &domainServices.ConceptRecommendation{
			NodeID:          uuid.New(),
			RecommendationType: "next",
			Score:           0.8,
			Confidence:      0.9,
			Reasoning:       []string{"Based on learning progress", "Matches skill level"},
			EstimatedTime:   time.Hour * 2,
			DifficultyMatch: 0.85,
			SkillRelevance:  0.9,
			Prerequisites:   []uuid.UUID{},
			LearningPath:    []uuid.UUID{},
		}
		recommendations = append(recommendations, recommendation)
	}

	return recommendations, nil
}

// mapToUpdateNodeRequest �?map[string]interface{} 转换�?UpdateNodeRequest
func (s *KnowledgeGraphAppService) mapToUpdateNodeRequest(updates map[string]interface{}) *UpdateNodeRequest {
	req := &UpdateNodeRequest{}
	
	if title, ok := updates["title"].(string); ok {
		req.Title = &title
	}
	if content, ok := updates["content"].(string); ok {
		req.Content = &content
	}
	if nodeType, ok := updates["type"].(string); ok {
		req.Type = &nodeType
	}
	if difficulty, ok := updates["difficulty"].(string); ok {
		req.Difficulty = &difficulty
	}
	if tags, ok := updates["tags"].([]string); ok {
		req.Tags = tags
	}
	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		// �?map[string]interface{} 转换�?map[string]string
		stringMetadata := make(map[string]string)
		for k, v := range metadata {
			if str, ok := v.(string); ok {
				stringMetadata[k] = str
			}
		}
		req.Metadata = stringMetadata
	}
	
	return req
}

// applyUpdatesToNode 将更新应用到现有节点
func (s *KnowledgeGraphAppService) applyUpdatesToNode(existingNode *entities.KnowledgeNode, updates *UpdateNodeRequest) *entities.KnowledgeNode {
	updatedNode := *existingNode // 创建副本

	if updates.Title != nil {
		updatedNode.Name = *updates.Title // Title 映射�?Name
	}
	if updates.Content != nil {
		updatedNode.Description = *updates.Content // Content 映射�?Description
	}
	if updates.Type != nil {
		updatedNode.Type = entities.NodeType(*updates.Type) // 转换�?NodeType
	}
	if updates.Difficulty != nil {
		// 将字符串难度转换�?DifficultyLevel
		switch *updates.Difficulty {
		case "beginner":
			updatedNode.DifficultyLevel = entities.DifficultyBeginner
		case "elementary":
			updatedNode.DifficultyLevel = entities.DifficultyElementary
		case "intermediate":
			updatedNode.DifficultyLevel = entities.DifficultyIntermediate
		case "advanced":
			updatedNode.DifficultyLevel = entities.DifficultyAdvanced
		case "expert":
			updatedNode.DifficultyLevel = entities.DifficultyExpert
		default:
			updatedNode.DifficultyLevel = entities.DifficultyBeginner
		}
	}
	if updates.Tags != nil {
		updatedNode.Tags = updates.Tags
	}
	if updates.Metadata != nil {
		// �?map[string]string 转换�?map[string]interface{}
		metadata := make(map[string]interface{})
		for k, v := range updates.Metadata {
			metadata[k] = v
		}
		updatedNode.Metadata = metadata
	}

	updatedNode.UpdatedAt = time.Now()

	return &updatedNode
}

// extractNodeIDs 从路径节点中提取节点ID列表
func extractNodeIDs(pathNodes []entities.PathNode) []uuid.UUID {
	nodeIDs := make([]uuid.UUID, len(pathNodes))
	for i, pathNode := range pathNodes {
		nodeIDs[i] = pathNode.KnowledgeID
	}
	return nodeIDs
}

// stringToDifficultyLevel 将字符串转换为DifficultyLevel
func (s *KnowledgeGraphAppService) stringToDifficultyLevel(str string) entities.DifficultyLevel {
	str = strings.ToLower(strings.TrimSpace(str))
	
	// 尝试按数字解�?
	if num, err := strconv.Atoi(str); err == nil {
		switch num {
		case 1:
			return entities.DifficultyBeginner
		case 2:
			return entities.DifficultyElementary
		case 3:
			return entities.DifficultyIntermediate
		case 4:
			return entities.DifficultyAdvanced
		case 5:
			return entities.DifficultyExpert
		}
	}
	
	// 按字符串解析
	switch str {
	case "beginner", "初学�?:
		return entities.DifficultyBeginner
	case "elementary", "基础":
		return entities.DifficultyElementary
	case "intermediate", "中级":
		return entities.DifficultyIntermediate
	case "advanced", "高级":
		return entities.DifficultyAdvanced
	case "expert", "专家":
		return entities.DifficultyExpert
	default:
		return entities.DifficultyBeginner // 默认为初学�?
	}
}

// difficultyLevelToString 将DifficultyLevel转换为字符串
func (s *KnowledgeGraphAppService) difficultyLevelToString(level entities.DifficultyLevel) string {
	switch level {
	case entities.DifficultyBeginner:
		return "beginner"
	case entities.DifficultyElementary:
		return "elementary"
	case entities.DifficultyIntermediate:
		return "intermediate"
	case entities.DifficultyAdvanced:
		return "advanced"
	case entities.DifficultyExpert:
		return "expert"
	default:
		return "intermediate"
	}
}

// applyUpdatesToRelation 将更新应用到现有关系
func (s *KnowledgeGraphAppService) applyUpdatesToRelation(existingRelation *entities.KnowledgeRelation, updates *UpdateRelationRequest) *entities.KnowledgeRelation {
	updatedRelation := *existingRelation // 创建副本

	if updates.Type != nil {
		updatedRelation.Type = entities.RelationType(*updates.Type)
	}
	if updates.Weight != nil {
		updatedRelation.Weight = *updates.Weight
	}
	if updates.Description != nil {
		updatedRelation.Description = *updates.Description
	}
	if updates.Properties != nil {
		// �?map[string]string 转换�?map[string]interface{}
		metadata := make(map[string]interface{})
		for k, v := range updates.Properties {
			metadata[k] = v
		}
		updatedRelation.Metadata = metadata
	}

	updatedRelation.UpdatedAt = time.Now()

	return &updatedRelation
}
