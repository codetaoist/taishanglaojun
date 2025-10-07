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

// LearningPathService 学习路径推荐服务
type LearningPathService struct {
	knowledgeGraphRepo repositories.KnowledgeGraphRepository
	learnerRepo        repositories.LearnerRepository
	contentRepo        repositories.LearningContentRepository
}

// NewLearningPathService 创建学习路径服务
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

// PathRecommendationRequest 路径推荐请求
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

// PathPreferences 路径偏好
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

// PersonalizedPath 个性化学习路径
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

// PathAdaptation 路径适配
type PathAdaptation struct {
	Type        string      `json:"type"` // "difficulty", "content_type", "pacing", "style"
	Description string      `json:"description"`
	Impact      float64     `json:"impact"`
	Confidence  float64     `json:"confidence"`
}

// PathMilestone 路径里程碑
type PathMilestone struct {
	NodeID          uuid.UUID     `json:"node_id"`
	Position        int           `json:"position"`
	EstimatedTime   time.Duration `json:"estimated_time"`
	SkillsAcquired  []string      `json:"skills_acquired"`
	Prerequisites   []uuid.UUID   `json:"prerequisites"`
	Assessments     []uuid.UUID   `json:"assessments"`
	Rewards         []string      `json:"rewards"`
}

// RecommendPersonalizedPaths 推荐个性化学习路径
func (s *LearningPathService) RecommendPersonalizedPaths(ctx context.Context, req *PathRecommendationRequest) ([]*PersonalizedPath, error) {
	// 获取学习者信息
	learner, err := s.learnerRepo.GetByID(ctx, req.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 获取知识图谱
	graph, err := s.knowledgeGraphRepo.GetGraph(ctx, req.GraphID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge graph: %w", err)
	}

	// 获取目标节点
	targetNode, err := s.knowledgeGraphRepo.GetNode(ctx, req.GraphID, req.TargetNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target node: %w", err)
	}

	// 分析学习者当前状态
	learnerState, err := s.analyzeLearnerState(ctx, learner, graph)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze learner state: %w", err)
	}

	// 生成候选路径
	candidatePaths, err := s.generateCandidatePaths(ctx, graph, learnerState, targetNode, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidate paths: %w", err)
	}

	// 个性化评分和排序
	personalizedPaths, err := s.personalizeAndRankPaths(ctx, candidatePaths, learner, req)
	if err != nil {
		return nil, fmt.Errorf("failed to personalize paths: %w", err)
	}

	// 限制返回数量
	if len(personalizedPaths) > req.MaxPaths {
		personalizedPaths = personalizedPaths[:req.MaxPaths]
	}

	return personalizedPaths, nil
}

// LearnerState 学习者状态
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

// LearningPatterns 学习模式
type LearningPatterns struct {
	OptimalSessionLength time.Duration               `json:"optimal_session_length"`
	PreferredTimeSlots   []entities.TimeSlot         `json:"preferred_time_slots"`
	ContentTypePreference map[entities.ContentType]float64 `json:"content_type_preference"`
	DifficultyProgression string                     `json:"difficulty_progression"` // "gradual", "steep", "mixed"
	RetentionRate        float64                     `json:"retention_rate"`
	EngagementFactors    []string                    `json:"engagement_factors"`
}

// analyzeLearnerState 分析学习者状态
func (s *LearningPathService) analyzeLearnerState(ctx context.Context, learner *entities.Learner, graph *entities.KnowledgeGraph) (*LearnerState, error) {
	// 获取学习者技能
	skills, err := s.learnerRepo.GetLearnerSkills(ctx, learner.ID)
	if err != nil {
		return nil, err
	}

	// 获取学习历史
	history, err := s.learnerRepo.GetLearningHistory(ctx, learner.ID, 100)
	if err != nil {
		return nil, err
	}

	// 分析已掌握的节点
	masteredNodes := s.identifyMasteredNodes(skills, graph)
	
	// 分析可用节点（满足前置条件的节点）
	availableNodes := s.identifyAvailableNodes(ctx, masteredNodes, graph)

	// 计算学习速度
	velocity := s.calculateLearningVelocity(history)

	// 分析学习模式
	patterns := s.analyzeLearningPatterns(history, learner)

	// 识别强项和弱项
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

// generateCandidatePaths 生成候选路径
func (s *LearningPathService) generateCandidatePaths(ctx context.Context, graph *entities.KnowledgeGraph, learnerState *LearnerState, targetNode *entities.KnowledgeNode, req *PathRecommendationRequest) ([]*entities.LearningPath, error) {
	var candidatePaths []*entities.LearningPath

	// 方法1: 基于最短路径的生成
	shortestPaths, err := s.generateShortestPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, shortestPaths...)
	}

	// 方法2: 基于技能导向的路径生成
	skillBasedPaths, err := s.generateSkillBasedPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, skillBasedPaths...)
	}

	// 方法3: 基于难度渐进的路径生成
	difficultyBasedPaths, err := s.generateDifficultyBasedPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, difficultyBasedPaths...)
	}

	// 方法4: 基于兴趣的路径生成
	interestBasedPaths, err := s.generateInterestBasedPaths(ctx, graph, learnerState, targetNode, req)
	if err == nil {
		candidatePaths = append(candidatePaths, interestBasedPaths...)
	}

	// 去重和过滤
	candidatePaths = s.deduplicateAndFilterPaths(candidatePaths, req)

	return candidatePaths, nil
}

// personalizeAndRankPaths 个性化评分和排序路径
func (s *LearningPathService) personalizeAndRankPaths(ctx context.Context, candidatePaths []*entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) ([]*PersonalizedPath, error) {
	var personalizedPaths []*PersonalizedPath

	for _, path := range candidatePaths {
		personalizedPath, err := s.personalizePath(ctx, path, learner, req)
		if err != nil {
			continue // 跳过无法个性化的路径
		}
		personalizedPaths = append(personalizedPaths, personalizedPath)
	}

	// 按个性化评分排序
	sort.Slice(personalizedPaths, func(i, j int) bool {
		return personalizedPaths[i].PersonalizationScore > personalizedPaths[j].PersonalizationScore
	})

	return personalizedPaths, nil
}

// personalizePath 个性化单个路径
func (s *LearningPathService) personalizePath(ctx context.Context, path *entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) (*PersonalizedPath, error) {
	// 计算个性化评分
	personalizationScore := s.calculatePersonalizationScore(path, learner, req)
	
	// 估算学习时长
	estimatedDuration := s.estimateLearningDuration(path, learner)
	
	// 分析难度进展
	difficultyProgression := s.analyzeDifficultyProgression(path)
	
	// 分析技能进展
	skillProgression := s.analyzeSkillProgression(path, learner)
	
	// 预测成功概率
	successProbability := s.predictSuccessProbability(path, learner)
	
	// 计算参与度评分
	engagementScore := s.calculateEngagementScore(path, learner)
	
	// 生成推理说明
	reasoning := s.generateReasoning(path, learner, req)
	
	// 识别适配调整
	adaptations := s.identifyAdaptations(path, learner)
	
	// 生成里程碑
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

// 辅助方法实现

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
	// 检查节点相关技能是否达到掌握水平
	for _, skill := range node.Skills {
		if skillLevel, exists := skills[skill]; exists {
			if skillLevel.Level >= 7 { // 假设7级以上为掌握
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
			continue // 已掌握的节点跳过
		}
		
		// 检查前置条件是否满足
		if s.arePrerequisitesMet(&node, masteredSet, graph) {
			availableNodes = append(availableNodes, node.ID)
		}
	}
	
	return availableNodes
}

func (s *LearningPathService) arePrerequisitesMet(node *entities.KnowledgeNode, masteredSet map[uuid.UUID]bool, graph *entities.KnowledgeGraph) bool {
	// 获取前置条件关系
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
		return 1.0 // 默认速度
	}
	
	// 计算最近的学习速度（内容完成率/时间）
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
	
	// 返回每小时完成的内容数量
	return float64(completedContent) / totalTime.Hours()
}

func (s *LearningPathService) calculatePreferredDifficulty(history []*entities.LearningHistory) entities.DifficultyLevel {
	if len(history) == 0 {
		return entities.DifficultyBeginner
	}
	
	// 分析历史中表现最好的难度级别
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
	// 分析最优学习时长
	var totalDuration time.Duration
	var sessionCount int
	
	for _, h := range history {
		totalDuration += h.Duration
		sessionCount++
	}
	
	optimalSessionLength := time.Hour // 默认1小时
	if sessionCount > 0 {
		optimalSessionLength = totalDuration / time.Duration(sessionCount)
	}
	
	// 分析内容类型偏好
	contentTypePreference := make(map[entities.ContentType]float64)
	for _, h := range history {
		contentTypePreference[entities.ContentType(h.ContentType)] += h.Progress
	}
	
	// 归一化偏好分数
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
		DifficultyProgression: "gradual", // 可以基于历史数据分析
		RetentionRate:         0.8,       // 可以基于重复学习数据计算
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
	
	// 从每个可用节点找到目标节点的最短路径
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
	// 基于目标技能生成路径
	var paths []*entities.LearningPath
	
	if len(req.TargetSkills) == 0 {
		return paths, nil
	}
	
	// 为每个目标技能找到相关节点
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
	// 基于难度渐进生成路径
	var paths []*entities.LearningPath
	
	// 按难度级别组织节点
	nodesByDifficulty := make(map[entities.DifficultyLevel][]*entities.KnowledgeNode)
	for _, node := range graph.Nodes {
		nodesByDifficulty[node.DifficultyLevel] = append(nodesByDifficulty[node.DifficultyLevel], &node)
	}
	
	// 生成渐进式路径
	currentDifficulty := learnerState.PreferredDifficulty
	var pathNodes []*entities.KnowledgeNode
	
	for currentDifficulty <= targetNode.DifficultyLevel {
		if nodes, exists := nodesByDifficulty[currentDifficulty]; exists {
			// 选择最相关的节点
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
	// 基于学习者兴趣生成路径
	var paths []*entities.LearningPath
	
	// 找到与学习者强项相关的节点
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
	// 检查节点是否与目标节点相关
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
	// 简单的去重逻辑，基于路径节点序列
	seen := make(map[string]bool)
	var uniquePaths []*entities.LearningPath
	
	for _, path := range paths {
		signature := s.getPathSignature(path)
		if !seen[signature] {
			seen[signature] = true
			
			// 应用过滤条件
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
	// 检查路径长度限制
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
	
	// 基于学习风格的评分 (权重: 0.3)
	styleScore := s.calculateStyleMatchScore(path, learner.Preferences.Style)
	score += styleScore * 0.3
	
	// 基于难度适配的评分 (权重: 0.25)
	difficultyScore := s.calculateDifficultyMatchScore(path, learner)
	score += difficultyScore * 0.25
	
	// 基于技能相关性的评分 (权重: 0.25)
	skillScore := s.calculateSkillRelevanceScore(path, learner)
	score += skillScore * 0.25
	
	// 基于时间约束的评分 (权重: 0.2)
	timeScore := s.calculateTimeConstraintScore(path, req.TimeConstraint)
	score += timeScore * 0.2
	
	return math.Min(score, 1.0) // 确保分数不超过1.0
}

func (s *LearningPathService) calculateStyleMatchScore(path *entities.LearningPath, style entities.LearningStyle) float64 {
	// 根据学习风格计算匹配度
	switch style {
	case entities.LearningStyleVisual:
		return 0.8 // 假设路径适合视觉学习者
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
	// 基于学习者经验水平计算难度匹配度
	avgDifficulty := s.calculateAveragePathDifficulty(path)
	experienceLevel := float64(learner.ExperienceLevel)
	
	// 计算难度差异
	difficultyDiff := math.Abs(avgDifficulty - experienceLevel)
	
	// 转换为匹配分数 (差异越小，分数越高)
	return math.Max(0, 1.0 - difficultyDiff/10.0)
}

func (s *LearningPathService) calculateAveragePathDifficulty(path *entities.LearningPath) float64 {
	// 这里需要从知识图谱中获取节点难度信息
	// 简化实现，返回中等难度
	return 5.0
}

func (s *LearningPathService) calculateSkillRelevanceScore(path *entities.LearningPath, learner *entities.Learner) float64 {
	// 计算路径与学习者技能的相关性
	// 简化实现
	return 0.8
}

func (s *LearningPathService) calculateTimeConstraintScore(path *entities.LearningPath, timeConstraint *time.Duration) float64 {
	if timeConstraint == nil {
		return 1.0 // 没有时间约束，满分
	}
	
	// 估算路径完成时间
	estimatedTime := time.Duration(len(path.Nodes)) * time.Hour * 2 // 假设每个节点需要2小时
	
	if estimatedTime <= *timeConstraint {
		return 1.0
	}
	
	// 超出时间约束，按比例扣分
	ratio := float64(*timeConstraint) / float64(estimatedTime)
	return math.Max(0, ratio)
}

func (s *LearningPathService) estimateLearningDuration(path *entities.LearningPath, learner *entities.Learner) time.Duration {
	// 基于路径长度和学习者速度估算时间
	baseTime := time.Duration(len(path.Nodes)) * time.Hour * 2
	
	// 根据学习者经验调整
	experienceFactor := 1.0 - (float64(learner.ExperienceLevel) / 20.0) // 经验越高，时间越短
	adjustedTime := time.Duration(float64(baseTime) * (0.5 + experienceFactor))
	
	return adjustedTime
}

func (s *LearningPathService) analyzeDifficultyProgression(path *entities.LearningPath) []float64 {
	progression := make([]float64, len(path.Nodes))
	
	for i := range path.Nodes {
		// 简化实现，假设难度逐渐增加
		progression[i] = float64(i+1) / float64(len(path.Nodes)) * 10.0
	}
	
	return progression
}

func (s *LearningPathService) analyzeSkillProgression(path *entities.LearningPath, learner *entities.Learner) map[string][]float64 {
	// 分析每个技能在路径中的进展
	skillProgression := make(map[string][]float64)
	
	// 简化实现
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
	// 基于学习者历史表现和路径特征预测成功概率
	baseProbability := 0.7 // 基础成功率
	
	// 根据经验水平调整
	experienceBonus := float64(learner.ExperienceLevel) / 100.0
	
	// 根据路径长度调整（路径越长，成功率可能越低）
	lengthPenalty := math.Max(0, float64(len(path.Nodes)-5)) * 0.02
	
	probability := baseProbability + experienceBonus - lengthPenalty
	return math.Max(0.1, math.Min(0.95, probability))
}

func (s *LearningPathService) calculateEngagementScore(path *entities.LearningPath, learner *entities.Learner) float64 {
	// 基于学习者偏好和路径特征计算参与度分数
	baseScore := 0.6
	
	// 根据学习风格调整
	styleBonus := 0.2 // 假设路径适合学习者风格
	
	// 根据路径多样性调整
	diversityBonus := 0.1 // 假设路径有良好的多样性
	
	return math.Min(1.0, baseScore + styleBonus + diversityBonus)
}

func (s *LearningPathService) generateReasoning(path *entities.LearningPath, learner *entities.Learner, req *PathRecommendationRequest) []string {
	var reasoning []string
	
	reasoning = append(reasoning, fmt.Sprintf("路径包含%d个学习节点，适合您的经验水平", len(path.Nodes)))
	reasoning = append(reasoning, "路径设计考虑了您的学习风格偏好")
	reasoning = append(reasoning, "难度进展合理，有助于循序渐进地掌握知识")
	
	if req.TimeConstraint != nil {
		reasoning = append(reasoning, "路径时间安排符合您的时间约束")
	}
	
	return reasoning
}

func (s *LearningPathService) identifyAdaptations(path *entities.LearningPath, learner *entities.Learner) []PathAdaptation {
	var adaptations []PathAdaptation
	
	// 基于学习者特征识别需要的适配
	if learner.Preferences.Style == entities.LearningStyleVisual {
		adaptations = append(adaptations, PathAdaptation{
			Type:        "content_type",
			Description: "增加视觉化学习材料",
			Impact:      0.8,
			Confidence:  0.9,
		})
	}
	
	if learner.ExperienceLevel < 3 {
		adaptations = append(adaptations, PathAdaptation{
			Type:        "difficulty",
			Description: "降低初始难度，增加基础概念讲解",
			Impact:      0.7,
			Confidence:  0.85,
		})
	}
	
	return adaptations
}

func (s *LearningPathService) generateMilestones(path *entities.LearningPath, learner *entities.Learner) []PathMilestone {
	var milestones []PathMilestone
	
	// 为路径中的关键节点生成里程碑
	for i, node := range path.Nodes {
		if i%3 == 0 || i == len(path.Nodes)-1 { // 每3个节点或最后一个节点设置里程碑
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