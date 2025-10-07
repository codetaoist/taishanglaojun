package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
)

// KnowledgeGraphRepository 定义知识图谱数据访问接口
type KnowledgeGraphRepository interface {
	// 知识图谱基本操作
	CreateGraph(ctx context.Context, graph *entities.KnowledgeGraph) error
	GetGraph(ctx context.Context, id uuid.UUID) (*entities.KnowledgeGraph, error)
	GetGraphByDomain(ctx context.Context, domain string) (*entities.KnowledgeGraph, error)
	UpdateGraph(ctx context.Context, graph *entities.KnowledgeGraph) error
	DeleteGraph(ctx context.Context, id uuid.UUID) error
	ListGraphs(ctx context.Context, offset, limit int) ([]*entities.KnowledgeGraph, error)

	// 知识节点操作
	AddNode(ctx context.Context, graphID uuid.UUID, node *entities.KnowledgeNode) error
	GetNode(ctx context.Context, graphID, nodeID uuid.UUID) (*entities.KnowledgeNode, error)
	GetNodeByName(ctx context.Context, graphID uuid.UUID, name string) (*entities.KnowledgeNode, error)
	UpdateNode(ctx context.Context, graphID uuid.UUID, node *entities.KnowledgeNode) error
	RemoveNode(ctx context.Context, graphID, nodeID uuid.UUID) error
	ListNodes(ctx context.Context, graphID uuid.UUID, nodeType *entities.NodeType, offset, limit int) ([]*entities.KnowledgeNode, error)
	SearchNodes(ctx context.Context, graphID uuid.UUID, query *NodeSearchQuery) ([]*entities.KnowledgeNode, int, error)

	// 知识关系操作
	AddRelation(ctx context.Context, graphID uuid.UUID, relation *entities.KnowledgeRelation) error
	GetRelation(ctx context.Context, graphID, relationID uuid.UUID) (*entities.KnowledgeRelation, error)
	UpdateRelation(ctx context.Context, graphID uuid.UUID, relation *entities.KnowledgeRelation) error
	RemoveRelation(ctx context.Context, graphID, relationID uuid.UUID) error
	GetNodeRelations(ctx context.Context, graphID, nodeID uuid.UUID, relationType *entities.RelationType) ([]*entities.KnowledgeRelation, error)
	GetRelationsBetween(ctx context.Context, graphID, fromNodeID, toNodeID uuid.UUID) ([]*entities.KnowledgeRelation, error)

	// 图遍历和查询
	GetPrerequisites(ctx context.Context, graphID, nodeID uuid.UUID, depth int) ([]*entities.KnowledgeNode, error)
	GetDependents(ctx context.Context, graphID, nodeID uuid.UUID, depth int) ([]*entities.KnowledgeNode, error)
	FindShortestPath(ctx context.Context, graphID, fromNodeID, toNodeID uuid.UUID) ([]*entities.KnowledgeNode, error)
	FindAllPaths(ctx context.Context, graphID, fromNodeID, toNodeID uuid.UUID, maxDepth int) ([][]*entities.KnowledgeNode, error)
	GetConnectedComponents(ctx context.Context, graphID uuid.UUID) ([][]uuid.UUID, error)
	GetNodeNeighbors(ctx context.Context, graphID, nodeID uuid.UUID, depth int) ([]*entities.KnowledgeNode, error)

	// 学习路径操作
	CreateLearningPath(ctx context.Context, path *entities.LearningPath) error
	GetLearningPath(ctx context.Context, id uuid.UUID) (*entities.LearningPath, error)
	UpdateLearningPath(ctx context.Context, path *entities.LearningPath) error
	DeleteLearningPath(ctx context.Context, id uuid.UUID) error
	ListLearningPaths(ctx context.Context, graphID uuid.UUID, offset, limit int) ([]*entities.LearningPath, error)
	GetLearningPathsByGoal(ctx context.Context, graphID, goalNodeID uuid.UUID) ([]*entities.LearningPath, error)
	GetPersonalizedPaths(ctx context.Context, graphID, learnerID uuid.UUID, targetNodeID uuid.UUID) ([]*entities.LearningPath, error)

	// 概念图操作
	CreateConceptMap(ctx context.Context, conceptMap *entities.ConceptMap) error
	GetConceptMap(ctx context.Context, centerNodeID uuid.UUID, depth, maxNodes int) (*entities.ConceptMap, error)
	GetConceptMapByID(ctx context.Context, id uuid.UUID) (*entities.ConceptMap, error)
	UpdateConceptMap(ctx context.Context, conceptMap *entities.ConceptMap) error
	DeleteConceptMap(ctx context.Context, id uuid.UUID) error
	GetConceptMapsByTopic(ctx context.Context, graphID uuid.UUID, topic string) ([]*entities.ConceptMap, error)

	// 图分析和统计
	GetGraphStatistics(ctx context.Context, graphID uuid.UUID) (*entities.GraphStatistics, error)
	UpdateGraphStatistics(ctx context.Context, graphID uuid.UUID) error
	GetNodeImportance(ctx context.Context, graphID uuid.UUID) (map[uuid.UUID]float64, error)
	GetGraphComplexity(ctx context.Context, graphID uuid.UUID) (*GraphComplexity, error)
	GetLearningPathEffectiveness(ctx context.Context, pathID uuid.UUID) (*PathEffectiveness, error)

	// 推荐和智能分析
	RecommendNextNodes(ctx context.Context, graphID, currentNodeID, learnerID uuid.UUID, limit int) ([]*NodeRecommendation, error)
	RecommendLearningPaths(ctx context.Context, graphID, learnerID uuid.UUID, targetSkills []string, limit int) ([]*PathRecommendation, error)
	AnalyzeLearningGaps(ctx context.Context, graphID, learnerID uuid.UUID) ([]*LearningGap, error)
	PredictLearningDifficulty(ctx context.Context, graphID, nodeID, learnerID uuid.UUID) (*DifficultyPrediction, error)

	// 版本控制和历史
	CreateGraphVersion(ctx context.Context, graphID uuid.UUID, version *GraphVersion) error
	GetGraphVersions(ctx context.Context, graphID uuid.UUID) ([]*GraphVersion, error)
	RestoreGraphVersion(ctx context.Context, graphID uuid.UUID, versionID uuid.UUID) error
	CompareGraphVersions(ctx context.Context, graphID, version1ID, version2ID uuid.UUID) (*GraphComparison, error)

	// 批量操作
	BatchAddNodes(ctx context.Context, graphID uuid.UUID, nodes []*entities.KnowledgeNode) error
	BatchAddRelations(ctx context.Context, graphID uuid.UUID, relations []*entities.KnowledgeRelation) error
	BatchUpdateNodes(ctx context.Context, graphID uuid.UUID, nodes []*entities.KnowledgeNode) error
	BatchRemoveNodes(ctx context.Context, graphID uuid.UUID, nodeIDs []uuid.UUID) error

	// 导入导出
	ExportGraph(ctx context.Context, graphID uuid.UUID, format string) ([]byte, error)
	ImportGraph(ctx context.Context, data []byte, format string) (*entities.KnowledgeGraph, error)
	ValidateGraphStructure(ctx context.Context, graphID uuid.UUID) (*GraphValidation, error)
}

// NodeSearchQuery 节点搜索查询
type NodeSearchQuery struct {
	Query           string                  `json:"query,omitempty"`
	Keywords        []string                `json:"keywords,omitempty"`
	NodeType        *entities.NodeType      `json:"node_type,omitempty"`
	DifficultyLevel *entities.DifficultyLevel `json:"difficulty_level,omitempty"`
	MinDifficulty   *int                    `json:"min_difficulty,omitempty"`
	MaxDifficulty   *int                    `json:"max_difficulty,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
	CreatedAfter    *time.Time              `json:"created_after,omitempty"`
	CreatedBefore   *time.Time              `json:"created_before,omitempty"`
	UpdatedAfter    *time.Time              `json:"updated_after,omitempty"`
	UpdatedBefore   *time.Time              `json:"updated_before,omitempty"`
	SortBy          string                  `json:"sort_by,omitempty"` // "name", "difficulty", "created_at", "updated_at", "importance"
	SortOrder       string                  `json:"sort_order,omitempty"` // "asc", "desc"
	Offset          int                     `json:"offset"`
	Limit           int                     `json:"limit"`
}

// GraphComplexity 图复杂度分析
type GraphComplexity struct {
	NodeCount           int                     `json:"node_count"`
	RelationCount       int                     `json:"relation_count"`
	AverageConnectivity float64                 `json:"average_connectivity"`
	MaxDepth            int                     `json:"max_depth"`
	CyclomaticComplexity int                    `json:"cyclomatic_complexity"`
	ClusteringCoefficient float64              `json:"clustering_coefficient"`
	NodeTypeDistribution map[entities.NodeType]int `json:"node_type_distribution"`
	RelationTypeDistribution map[entities.RelationType]int `json:"relation_type_distribution"`
	DifficultyDistribution map[entities.DifficultyLevel]int `json:"difficulty_distribution"`
	ConnectedComponents int                     `json:"connected_components"`
	LongestPath         int                     `json:"longest_path"`
	AveragePathLength   float64                 `json:"average_path_length"`
}

// PathEffectiveness 学习路径有效性
type PathEffectiveness struct {
	PathID              uuid.UUID   `json:"path_id"`
	CompletionRate      float64     `json:"completion_rate"`
	AverageCompletionTime time.Duration `json:"average_completion_time"`
	LearnerSatisfaction float64     `json:"learner_satisfaction"`
	SkillImprovement    float64     `json:"skill_improvement"`
	RetentionRate       float64     `json:"retention_rate"`
	DifficultyProgression float64   `json:"difficulty_progression"`
	PrerequisiteAlignment float64   `json:"prerequisite_alignment"`
	LearnerCount        int         `json:"learner_count"`
	SuccessRate         float64     `json:"success_rate"`
	RecommendationScore float64     `json:"recommendation_score"`
}

// NodeRecommendation 节点推荐
type NodeRecommendation struct {
	Node            *entities.KnowledgeNode `json:"node"`
	RecommendationScore float64             `json:"recommendation_score"`
	Reasoning       []string                `json:"reasoning"`
	EstimatedTime   time.Duration           `json:"estimated_time"`
	DifficultyMatch float64                 `json:"difficulty_match"`
	PrerequisitesMet bool                   `json:"prerequisites_met"`
	LearningStyle   entities.LearningStyle  `json:"learning_style"`
	Priority        int                     `json:"priority"`
}

// PathRecommendation 路径推荐
type PathRecommendation struct {
	Path                *entities.LearningPath `json:"path"`
	RecommendationScore float64                `json:"recommendation_score"`
	Reasoning           []string               `json:"reasoning"`
	EstimatedDuration   time.Duration          `json:"estimated_duration"`
	DifficultyProgression []float64            `json:"difficulty_progression"`
	SkillCoverage       map[string]float64     `json:"skill_coverage"`
	PersonalizationScore float64               `json:"personalization_score"`
	SuccessProbability  float64                `json:"success_probability"`
}

// LearningGap 学习差距
type LearningGap struct {
	SkillArea           string                  `json:"skill_area"`
	CurrentLevel        int                     `json:"current_level"`
	RequiredLevel       int                     `json:"required_level"`
	Gap                 int                     `json:"gap"`
	RecommendedNodes    []uuid.UUID             `json:"recommended_nodes"`
	EstimatedTime       time.Duration           `json:"estimated_time"`
	Priority            string                  `json:"priority"`
	DependentSkills     []string                `json:"dependent_skills"`
}

// DifficultyPrediction 难度预测
type DifficultyPrediction struct {
	NodeID              uuid.UUID   `json:"node_id"`
	LearnerID           uuid.UUID   `json:"learner_id"`
	PredictedDifficulty float64     `json:"predicted_difficulty"`
	Confidence          float64     `json:"confidence"`
	EstimatedTime       time.Duration `json:"estimated_time"`
	SuccessProbability  float64     `json:"success_probability"`
	RecommendedPrep     []*entities.KnowledgeNode `json:"recommended_prep"`
	RiskFactors         []string    `json:"risk_factors"`
	SupportResources    []string    `json:"support_resources"`
}

// GraphVersion 图版本
type GraphVersion struct {
	ID          uuid.UUID   `json:"id"`
	GraphID     uuid.UUID   `json:"graph_id"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   uuid.UUID   `json:"created_by"`
	Changes     []Change    `json:"changes"`
	Snapshot    []byte      `json:"snapshot"`
}

// Change 变更记录
type Change struct {
	Type        string      `json:"type"` // "add_node", "remove_node", "update_node", "add_relation", "remove_relation", "update_relation"
	EntityID    uuid.UUID   `json:"entity_id"`
	EntityType  string      `json:"entity_type"` // "node", "relation"
	OldValue    interface{} `json:"old_value,omitempty"`
	NewValue    interface{} `json:"new_value,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// GraphComparison 图比较
type GraphComparison struct {
	Version1ID      uuid.UUID   `json:"version1_id"`
	Version2ID      uuid.UUID   `json:"version2_id"`
	AddedNodes      []uuid.UUID `json:"added_nodes"`
	RemovedNodes    []uuid.UUID `json:"removed_nodes"`
	ModifiedNodes   []uuid.UUID `json:"modified_nodes"`
	AddedRelations  []uuid.UUID `json:"added_relations"`
	RemovedRelations []uuid.UUID `json:"removed_relations"`
	ModifiedRelations []uuid.UUID `json:"modified_relations"`
	Summary         ComparisonSummary `json:"summary"`
}

// ComparisonSummary 比较摘要
type ComparisonSummary struct {
	TotalChanges    int     `json:"total_changes"`
	NodeChanges     int     `json:"node_changes"`
	RelationChanges int     `json:"relation_changes"`
	ChangeRate      float64 `json:"change_rate"`
	MajorChanges    []string `json:"major_changes"`
}

// GraphValidation 图验证
type GraphValidation struct {
	IsValid         bool            `json:"is_valid"`
	Errors          []ValidationError `json:"errors"`
	Warnings        []ValidationWarning `json:"warnings"`
	Statistics      ValidationStatistics `json:"statistics"`
	Suggestions     []string        `json:"suggestions"`
}

// ValidationError 验证错误
type ValidationError struct {
	Type        string      `json:"type"`
	Message     string      `json:"message"`
	EntityID    uuid.UUID   `json:"entity_id,omitempty"`
	EntityType  string      `json:"entity_type,omitempty"`
	Severity    string      `json:"severity"` // "critical", "major", "minor"
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	Type        string      `json:"type"`
	Message     string      `json:"message"`
	EntityID    uuid.UUID   `json:"entity_id,omitempty"`
	EntityType  string      `json:"entity_type,omitempty"`
	Suggestion  string      `json:"suggestion,omitempty"`
}

// ValidationStatistics 验证统计
type ValidationStatistics struct {
	TotalNodes          int `json:"total_nodes"`
	TotalRelations      int `json:"total_relations"`
	OrphanedNodes       int `json:"orphaned_nodes"`
	CircularDependencies int `json:"circular_dependencies"`
	MissingPrerequisites int `json:"missing_prerequisites"`
	InconsistentDifficulty int `json:"inconsistent_difficulty"`
	DuplicateRelations  int `json:"duplicate_relations"`
}

// RelationDirection 关系方向枚举
type RelationDirection string

const (
	RelationDirectionIncoming RelationDirection = "incoming"
	RelationDirectionOutgoing RelationDirection = "outgoing"
	RelationDirectionBoth     RelationDirection = "both"
)

// GraphStatistics 图统计信息
type GraphStatistics struct {
	NodeCount       int                                    `json:"node_count"`
	RelationCount   int                                    `json:"relation_count"`
	PathCount       int                                    `json:"path_count"`
	NodesByType     map[entities.NodeType]int              `json:"nodes_by_type"`
	RelationsByType map[entities.RelationType]int          `json:"relations_by_type"`
	AvgDegree       float64                                `json:"avg_degree"`
	Density         float64                                `json:"density"`
	LastUpdated     time.Time                              `json:"last_updated"`
}