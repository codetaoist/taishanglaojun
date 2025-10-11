package entities

import (
	"time"

	"github.com/google/uuid"
)

// NodeType 节点类型
type NodeType string

const (
	NodeTypeConcept    NodeType = "concept"    // 概念
	NodeTypeSkill      NodeType = "skill"      // 技�?
	NodeTypeTopic      NodeType = "topic"      // 主题
	NodeTypeSubject    NodeType = "subject"    // 学科
	NodeTypeLearningPath NodeType = "learning_path" // 学习路径
	NodeTypeContent    NodeType = "content"    // 学习内容
)

// RelationType 关系类型
type RelationType string

const (
	RelationTypePrerequisite RelationType = "prerequisite" // 前置条件
	RelationTypePartOf       RelationType = "part_of"      // 属于
	RelationTypeRelatedTo    RelationType = "related_to"   // 相关
	RelationTypeLeadsTo      RelationType = "leads_to"     // 导向
	RelationTypeSimilarTo    RelationType = "similar_to"   // 相似
	RelationTypeOppositeOf   RelationType = "opposite_of"  // 相对
	RelationTypeExampleOf    RelationType = "example_of"   // 示例
	RelationTypeApplicationOf RelationType = "application_of" // 应用
)

// DifficultyLevel 难度等级
type DifficultyLevel int

const (
	DifficultyBeginner     DifficultyLevel = 1 // 初学�?
	DifficultyElementary   DifficultyLevel = 2 // 基础
	DifficultyIntermediate DifficultyLevel = 3 // 中级
	DifficultyAdvanced     DifficultyLevel = 4 // 高级
	DifficultyExpert       DifficultyLevel = 5 // 专家
)

// KnowledgeNode 知识节点
type KnowledgeNode struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Type            NodeType        `json:"type"`
	Subject         string          `json:"subject"`         // 所属学�?
	DifficultyLevel DifficultyLevel `json:"difficulty_level"`
	EstimatedHours  float64         `json:"estimated_hours"` // 预估学习时间
	Prerequisites   []uuid.UUID     `json:"prerequisites"`   // 前置知识点ID
	Skills          []string        `json:"skills"`          // 相关技�?
	Keywords        []string        `json:"keywords"`        // 关键�?
	Tags            []string        `json:"tags"`            // 标签
	Metadata        map[string]interface{} `json:"metadata"` // 元数�?
	LearningObjectives []string     `json:"learning_objectives"` // 学习目标
	AssessmentCriteria []string     `json:"assessment_criteria"` // 评估标准
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// KnowledgeRelation 知识关系
type KnowledgeRelation struct {
	ID          uuid.UUID    `json:"id"`
	FromNodeID  uuid.UUID    `json:"from_node_id"`
	ToNodeID    uuid.UUID    `json:"to_node_id"`
	Type        RelationType `json:"type"`
	Weight      float64      `json:"weight"`      // 关系权重 0.0-1.0
	Confidence  float64      `json:"confidence"`  // 置信�?0.0-1.0
	Description string       `json:"description"` // 关系描述
	Metadata    map[string]interface{} `json:"metadata"` // 元数�?
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// NewKnowledgeRelation 创建新的知识关系
func NewKnowledgeRelation(fromNodeID, toNodeID uuid.UUID, relationType RelationType, weight float64) *KnowledgeRelation {
	now := time.Now()
	return &KnowledgeRelation{
		ID:          uuid.New(),
		FromNodeID:  fromNodeID,
		ToNodeID:    toNodeID,
		Type:        relationType,
		Weight:      weight,
		Confidence:  0.8, // 默认置信�?
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ConceptMap 概念�?
type ConceptMap struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Subject     string              `json:"subject"`
	GraphID     uuid.UUID           `json:"graph_id"`
	Nodes       []KnowledgeNode     `json:"nodes"`
	Relations   []KnowledgeRelation `json:"relations"`
	CreatedBy   uuid.UUID           `json:"created_by"`
	IsPublic    bool                `json:"is_public"`
	Version     int                 `json:"version"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// LearningPath 学习路径
type LearningPath struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Subject         string          `json:"subject"`
	DifficultyLevel DifficultyLevel `json:"difficulty_level"`
	EstimatedHours  float64         `json:"estimated_hours"`
	Prerequisites   []uuid.UUID     `json:"prerequisites"`   // 前置技�?知识�?
	LearningGoals   []string        `json:"learning_goals"`  // 学习目标
	Nodes           []PathNode      `json:"nodes"`           // 路径节点
	Milestones      []Milestone     `json:"milestones"`      // 里程�?
	Tags            []string        `json:"tags"`
	IsPublic        bool            `json:"is_public"`
	CreatedBy       uuid.UUID       `json:"created_by"`
	EnrollmentCount int             `json:"enrollment_count"` // 注册人数
	CompletionRate  float64         `json:"completion_rate"`  // 完成�?
	Rating          float64         `json:"rating"`           // 评分
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PathNode 路径节点
type PathNode struct {
	ID           uuid.UUID   `json:"id"`
	KnowledgeID  uuid.UUID   `json:"knowledge_id"`  // 关联的知识点ID
	Order        int         `json:"order"`         // 顺序
	IsOptional   bool        `json:"is_optional"`   // 是否可�?
	Dependencies []uuid.UUID `json:"dependencies"`  // 依赖的其他节�?
	Metadata     map[string]interface{} `json:"metadata"`
}

// Milestone 里程�?
type Milestone struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	NodeIDs     []uuid.UUID `json:"node_ids"` // 包含的节点ID
	Criteria    []string  `json:"criteria"`  // 完成标准
}

// KnowledgeGraph 知识图谱
type KnowledgeGraph struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Domain      string              `json:"domain"`      // 领域
	Subject     string              `json:"subject"`     // 学科
	Version     string              `json:"version"`
	IsPublic    bool                `json:"is_public"`   // 是否公开
	CreatedBy   uuid.UUID           `json:"created_by"`  // 创建者ID
	Nodes       []KnowledgeNode     `json:"nodes"`
	Relations   []KnowledgeRelation `json:"relations"`
	ConceptMaps []ConceptMap        `json:"concept_maps"`
	Paths       []LearningPath      `json:"paths"`
	Statistics  GraphStatistics     `json:"statistics"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// GraphStatistics 图谱统计信息
type GraphStatistics struct {
	NodeCount     int                    `json:"node_count"`
	RelationCount int                    `json:"relation_count"`
	PathCount     int                    `json:"path_count"`
	NodesByType   map[NodeType]int       `json:"nodes_by_type"`
	RelationsByType map[RelationType]int `json:"relations_by_type"`
	AvgDegree     float64                `json:"avg_degree"`     // 平均度数
	Density       float64                `json:"density"`        // 密度
	LastUpdated   time.Time              `json:"last_updated"`
}

// NewKnowledgeGraph 创建新的知识图谱
func NewKnowledgeGraph(name, description, domain string) *KnowledgeGraph {
	now := time.Now()
	return &KnowledgeGraph{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Domain:      domain,
		Version:     "1.0.0",
		Nodes:       make([]KnowledgeNode, 0),
		Relations:   make([]KnowledgeRelation, 0),
		ConceptMaps: make([]ConceptMap, 0),
		Paths:       make([]LearningPath, 0),
		Statistics: GraphStatistics{
			NodesByType:     make(map[NodeType]int),
			RelationsByType: make(map[RelationType]int),
			LastUpdated:     now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddNode 添加知识节点
func (kg *KnowledgeGraph) AddNode(node KnowledgeNode) {
	kg.Nodes = append(kg.Nodes, node)
	kg.updateStatistics()
	kg.UpdatedAt = time.Now()
}

// AddRelation 添加知识关系
func (kg *KnowledgeGraph) AddRelation(relation KnowledgeRelation) {
	kg.Relations = append(kg.Relations, relation)
	kg.updateStatistics()
	kg.UpdatedAt = time.Now()
}

// GetNode 获取节点
func (kg *KnowledgeGraph) GetNode(nodeID uuid.UUID) *KnowledgeNode {
	for i, node := range kg.Nodes {
		if node.ID == nodeID {
			return &kg.Nodes[i]
		}
	}
	return nil
}

// GetNodesByType 根据类型获取节点
func (kg *KnowledgeGraph) GetNodesByType(nodeType NodeType) []KnowledgeNode {
	var nodes []KnowledgeNode
	for _, node := range kg.Nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetRelations 获取节点的所有关�?
func (kg *KnowledgeGraph) GetRelations(nodeID uuid.UUID) []KnowledgeRelation {
	var relations []KnowledgeRelation
	for _, relation := range kg.Relations {
		if relation.FromNodeID == nodeID || relation.ToNodeID == nodeID {
			relations = append(relations, relation)
		}
	}
	return relations
}

// GetPrerequisites 获取前置条件
func (kg *KnowledgeGraph) GetPrerequisites(nodeID uuid.UUID) []KnowledgeNode {
	var prerequisites []KnowledgeNode
	for _, relation := range kg.Relations {
		if relation.ToNodeID == nodeID && relation.Type == RelationTypePrerequisite {
			if node := kg.GetNode(relation.FromNodeID); node != nil {
				prerequisites = append(prerequisites, *node)
			}
		}
	}
	return prerequisites
}

// GetDependents 获取依赖此节点的节点
func (kg *KnowledgeGraph) GetDependents(nodeID uuid.UUID) []KnowledgeNode {
	var dependents []KnowledgeNode
	for _, relation := range kg.Relations {
		if relation.FromNodeID == nodeID && relation.Type == RelationTypePrerequisite {
			if node := kg.GetNode(relation.ToNodeID); node != nil {
				dependents = append(dependents, *node)
			}
		}
	}
	return dependents
}

// FindShortestPath 查找两个节点间的最短路�?
func (kg *KnowledgeGraph) FindShortestPath(fromID, toID uuid.UUID) []uuid.UUID {
	// 使用BFS算法查找最短路�?
	if fromID == toID {
		return []uuid.UUID{fromID}
	}

	visited := make(map[uuid.UUID]bool)
	queue := [][]uuid.UUID{{fromID}}
	visited[fromID] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		currentNode := path[len(path)-1]

		// 获取当前节点的所有邻�?
		for _, relation := range kg.Relations {
			var nextNode uuid.UUID
			if relation.FromNodeID == currentNode {
				nextNode = relation.ToNodeID
			} else if relation.ToNodeID == currentNode {
				nextNode = relation.FromNodeID
			} else {
				continue
			}

			if nextNode == toID {
				return append(path, nextNode)
			}

			if !visited[nextNode] {
				visited[nextNode] = true
				newPath := make([]uuid.UUID, len(path)+1)
				copy(newPath, path)
				newPath[len(path)] = nextNode
				queue = append(queue, newPath)
			}
		}
	}

	return nil // 没有找到路径
}

// updateStatistics 更新统计信息
func (kg *KnowledgeGraph) updateStatistics() {
	kg.Statistics.NodeCount = len(kg.Nodes)
	kg.Statistics.RelationCount = len(kg.Relations)
	kg.Statistics.PathCount = len(kg.Paths)

	// 重置计数�?
	kg.Statistics.NodesByType = make(map[NodeType]int)
	kg.Statistics.RelationsByType = make(map[RelationType]int)

	// 统计节点类型
	for _, node := range kg.Nodes {
		kg.Statistics.NodesByType[node.Type]++
	}

	// 统计关系类型
	for _, relation := range kg.Relations {
		kg.Statistics.RelationsByType[relation.Type]++
	}

	// 计算平均度数
	if kg.Statistics.NodeCount > 0 {
		kg.Statistics.AvgDegree = float64(kg.Statistics.RelationCount*2) / float64(kg.Statistics.NodeCount)
	}

	// 计算密度
	if kg.Statistics.NodeCount > 1 {
		maxEdges := kg.Statistics.NodeCount * (kg.Statistics.NodeCount - 1) / 2
		kg.Statistics.Density = float64(kg.Statistics.RelationCount) / float64(maxEdges)
	}

	kg.Statistics.LastUpdated = time.Now()
}

// NewLearningPath 创建新的学习路径
func NewLearningPath(name, description, subject string, difficulty DifficultyLevel, createdBy uuid.UUID) *LearningPath {
	now := time.Now()
	return &LearningPath{
		ID:              uuid.New(),
		Name:            name,
		Description:     description,
		Subject:         subject,
		DifficultyLevel: difficulty,
		EstimatedHours:  0,
		Prerequisites:   make([]uuid.UUID, 0),
		LearningGoals:   make([]string, 0),
		Nodes:           make([]PathNode, 0),
		Milestones:      make([]Milestone, 0),
		Tags:            make([]string, 0),
		IsPublic:        false,
		CreatedBy:       createdBy,
		EnrollmentCount: 0,
		CompletionRate:  0,
		Rating:          0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// AddPathNode 添加路径节点
func (lp *LearningPath) AddPathNode(knowledgeID uuid.UUID, order int, isOptional bool, dependencies []uuid.UUID) {
	node := PathNode{
		ID:           uuid.New(),
		KnowledgeID:  knowledgeID,
		Order:        order,
		IsOptional:   isOptional,
		Dependencies: dependencies,
		Metadata:     make(map[string]interface{}),
	}
	lp.Nodes = append(lp.Nodes, node)
	lp.UpdatedAt = time.Now()
}

// AddMilestone 添加里程�?
func (lp *LearningPath) AddMilestone(name, description string, order int, nodeIDs []uuid.UUID, criteria []string) {
	milestone := Milestone{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Order:       order,
		NodeIDs:     nodeIDs,
		Criteria:    criteria,
	}
	lp.Milestones = append(lp.Milestones, milestone)
	lp.UpdatedAt = time.Now()
}

// GetOrderedNodes 获取按顺序排列的节点
func (lp *LearningPath) GetOrderedNodes() []PathNode {
	nodes := make([]PathNode, len(lp.Nodes))
	copy(nodes, lp.Nodes)

	// 按order字段排序
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].Order > nodes[j].Order {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	return nodes
}

// ValidatePath 验证路径的有效�?
func (lp *LearningPath) ValidatePath() []string {
	var errors []string

	// 检查节点顺�?
	orders := make(map[int]bool)
	for _, node := range lp.Nodes {
		if orders[node.Order] {
			errors = append(errors, "重复的节点顺�?)
		}
		orders[node.Order] = true
	}

	// 检查依赖关�?
	nodeIDs := make(map[uuid.UUID]bool)
	for _, node := range lp.Nodes {
		nodeIDs[node.ID] = true
	}

	for _, node := range lp.Nodes {
		for _, dep := range node.Dependencies {
			if !nodeIDs[dep] {
				errors = append(errors, "依赖的节点不存在于路径中")
			}
		}
	}

	return errors
}
