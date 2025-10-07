package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

type KnowledgeGraphRepositoryImpl struct {
	driver neo4j.DriverWithContext
}

func NewKnowledgeGraphRepository(driver neo4j.DriverWithContext) repositories.KnowledgeGraphRepository {
	return &KnowledgeGraphRepositoryImpl{
		driver: driver,
	}
}

// CreateGraph 创建知识图谱
func (r *KnowledgeGraphRepositoryImpl) CreateGraph(ctx context.Context, graph *entities.KnowledgeGraph) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (g:KnowledgeGraph {
				id: $id,
				name: $name,
				description: $description,
				domain: $domain,
				subject: $subject,
				version: $version,
				is_public: $isPublic,
				created_by: $createdBy,
				created_at: $createdAt,
				updated_at: $updatedAt
			})
			RETURN g
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          graph.ID.String(),
			"name":        graph.Name,
			"description": graph.Description,
			"domain":      graph.Domain,
			"subject":     graph.Subject,
			"version":     graph.Version,
			"isPublic":    graph.IsPublic,
			"createdBy":   graph.CreatedBy.String(),
			"createdAt":   graph.CreatedAt,
			"updatedAt":   graph.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// UpdateGraph 更新知识图谱
func (r *KnowledgeGraphRepositoryImpl) UpdateGraph(ctx context.Context, graph *entities.KnowledgeGraph) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $id})
			SET g.name = $name,
				g.description = $description,
				g.domain = $domain,
				g.subject = $subject,
				g.version = $version,
				g.is_public = $isPublic,
				g.updated_at = $updatedAt
			RETURN g
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          graph.ID.String(),
			"name":        graph.Name,
			"description": graph.Description,
			"domain":      graph.Domain,
			"subject":     graph.Subject,
			"version":     graph.Version,
			"isPublic":    graph.IsPublic,
			"updatedAt":   graph.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// ExportGraph 导出知识图谱
func (r *KnowledgeGraphRepositoryImpl) ExportGraph(ctx context.Context, graphID uuid.UUID, format string) ([]byte, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 获取图谱基本信息
		graphQuery := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			RETURN g
		`

		graphResult, err := tx.Run(ctx, graphQuery, map[string]interface{}{
			"graphID": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		if !graphResult.Next(ctx) {
			return nil, fmt.Errorf("knowledge graph not found")
		}

		graphRecord := graphResult.Record()
		graphValue, ok := graphRecord.Get("g")
		if !ok {
			return nil, fmt.Errorf("graph not found in result")
		}

		graphNode := graphValue.(neo4j.Node)
		graph, err := r.mapGraphFromNeo4j(graphNode)
		if err != nil {
			return nil, err
		}

		// 获取所有节点
		nodesQuery := `
			MATCH (n:KnowledgeNode)
			WHERE n.graph_id = $graphID
			RETURN n
		`

		nodesResult, err := tx.Run(ctx, nodesQuery, map[string]interface{}{
			"graphID": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		var nodes []entities.KnowledgeNode
		for nodesResult.Next(ctx) {
			nodeRecord := nodesResult.Record()
			nodeValue, ok := nodeRecord.Get("n")
			if !ok {
				continue
			}

			node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, *node)
		}

		// 直接赋值，nodes已经是值类型数组
		graph.Nodes = nodes

		// 获取所有关系
		relationshipsQuery := `
			MATCH (source:KnowledgeNode)-[r:PREREQUISITE]->(target:KnowledgeNode)
			WHERE source.graph_id = $graphID AND target.graph_id = $graphID
			RETURN r, source.id as sourceID, target.id as targetID
		`

		relationshipsResult, err := tx.Run(ctx, relationshipsQuery, map[string]interface{}{
			"graphID": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		var relationships []entities.KnowledgeRelation
		for relationshipsResult.Next(ctx) {
			relRecord := relationshipsResult.Record()
			relValue, ok := relRecord.Get("r")
			if !ok {
				continue
			}

			relationship, err := r.mapRelationshipFromNeo4j(relValue.(neo4j.Relationship))
			if err != nil {
				return nil, err
			}
			relationships = append(relationships, *relationship)
		}

		// 直接赋值，relationships已经是值类型数组
		graph.Relations = relationships

		return graph, nil
	})

	if err != nil {
		return nil, err
	}

	graph := result.(*entities.KnowledgeGraph)
	
	// 根据格式序列化图谱数据
	switch format {
	case "json":
		data, err := json.Marshal(graph)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal graph to JSON: %w", err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetLearningPath 获取学习路径
func (r *KnowledgeGraphRepositoryImpl) GetLearningPath(ctx context.Context, id uuid.UUID) (*entities.LearningPath, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (lp:LearningPath {id: $id})
			RETURN lp
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id": id.String(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			node, ok := record.Get("lp")
			if !ok {
				return nil, fmt.Errorf("learning path not found")
			}

			return r.mapLearningPathFromNeo4j(node.(neo4j.Node))
		}

		return nil, fmt.Errorf("learning path not found")
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.LearningPath), nil
}

// GetLearningPathsByGoal 根据目标获取学习路径
func (r *KnowledgeGraphRepositoryImpl) GetLearningPathsByGoal(ctx context.Context, graphID, goalNodeID uuid.UUID) ([]*entities.LearningPath, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (goal:KnowledgeNode {id: $goalNodeId})<-[:LEADS_TO]-(lp:LearningPath)
			WHERE lp.graph_id = $graphId
			RETURN lp
			ORDER BY lp.rating DESC, lp.enrollment_count DESC
			LIMIT 20
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphId":    graphID.String(),
			"goalNodeId": goalNodeID.String(),
		})
		if err != nil {
			return nil, err
		}

		var paths []*entities.LearningPath
		for result.Next(ctx) {
			record := result.Record()
			node, ok := record.Get("lp")
			if !ok {
				continue
			}

			path, err := r.mapLearningPathFromNeo4j(node.(neo4j.Node))
			if err != nil {
				continue
			}

			paths = append(paths, path)
		}

		return paths, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*entities.LearningPath), nil
}

// DeleteLearningPath 删除学习路径
func (r *KnowledgeGraphRepositoryImpl) DeleteLearningPath(ctx context.Context, pathID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (p:LearningPath {id: $pathID})
			DETACH DELETE p
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"pathID": pathID.String(),
		})

		return nil, err
	})

	return err
}

// DeleteGraph 删除知识图谱
func (r *KnowledgeGraphRepositoryImpl) DeleteGraph(ctx context.Context, graphID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 删除图谱及其所有相关节点和关系
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			OPTIONAL MATCH (n:KnowledgeNode {graph_id: $graphID})
			DETACH DELETE g, n
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
		})

		return nil, err
	})

	return err
}

// AddNode 添加知识节点
func (r *KnowledgeGraphRepositoryImpl) AddNode(ctx context.Context, graphID uuid.UUID, node *entities.KnowledgeNode) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 序列化各种数组字段
		tagsJSON, _ := json.Marshal(node.Tags)
		metadataJSON, _ := json.Marshal(node.Metadata)
		prerequisitesJSON, _ := json.Marshal(node.Prerequisites)
		skillsJSON, _ := json.Marshal(node.Skills)
		keywordsJSON, _ := json.Marshal(node.Keywords)
		learningObjectivesJSON, _ := json.Marshal(node.LearningObjectives)
		assessmentCriteriaJSON, _ := json.Marshal(node.AssessmentCriteria)

		query := `
			CREATE (n:KnowledgeNode {
				id: $id,
				graph_id: $graphID,
				name: $name,
				description: $description,
				node_type: $nodeType,
				subject: $subject,
				difficulty_level: $difficultyLevel,
				estimated_learning_time: $estimatedHours,
				prerequisites: $prerequisites,
				skills: $skills,
				keywords: $keywords,
				tags: $tags,
				metadata: $metadata,
				learning_objectives: $learningObjectives,
				assessment_criteria: $assessmentCriteria,
				created_at: $createdAt,
				updated_at: $updatedAt
			})
			RETURN n
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":                     node.ID.String(),
			"graphID":                graphID.String(),
			"name":                   node.Name,
			"description":            node.Description,
			"nodeType":               string(node.Type),
			"subject":                node.Subject,
			"difficultyLevel":        int(node.DifficultyLevel),
			"estimatedHours":         node.EstimatedHours,
			"prerequisites":          string(prerequisitesJSON),
			"skills":                 string(skillsJSON),
			"keywords":               string(keywordsJSON),
			"tags":                   string(tagsJSON),
			"metadata":               string(metadataJSON),
			"learningObjectives":     string(learningObjectivesJSON),
			"assessmentCriteria":     string(assessmentCriteriaJSON),
			"createdAt":              node.CreatedAt,
			"updatedAt":              node.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// UpdateNode 更新知识节点
func (r *KnowledgeGraphRepositoryImpl) UpdateNode(ctx context.Context, graphID uuid.UUID, node *entities.KnowledgeNode) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 序列化各种数组字段
		tagsJSON, _ := json.Marshal(node.Tags)
		metadataJSON, _ := json.Marshal(node.Metadata)
		prerequisitesJSON, _ := json.Marshal(node.Prerequisites)
		skillsJSON, _ := json.Marshal(node.Skills)
		keywordsJSON, _ := json.Marshal(node.Keywords)
		learningObjectivesJSON, _ := json.Marshal(node.LearningObjectives)
		assessmentCriteriaJSON, _ := json.Marshal(node.AssessmentCriteria)

		query := `
			MATCH (n:KnowledgeNode {id: $id, graph_id: $graphID})
			SET n.name = $name,
				n.description = $description,
				n.node_type = $nodeType,
				n.subject = $subject,
				n.difficulty_level = $difficultyLevel,
				n.estimated_learning_time = $estimatedHours,
				n.prerequisites = $prerequisites,
				n.skills = $skills,
				n.keywords = $keywords,
				n.tags = $tags,
				n.metadata = $metadata,
				n.learning_objectives = $learningObjectives,
				n.assessment_criteria = $assessmentCriteria,
				n.updated_at = $updatedAt
			RETURN n
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":                     node.ID.String(),
			"graphID":                graphID.String(),
			"name":                   node.Name,
			"description":            node.Description,
			"nodeType":               string(node.Type),
			"subject":                node.Subject,
			"difficultyLevel":        int(node.DifficultyLevel),
			"estimatedHours":         node.EstimatedHours,
			"prerequisites":          string(prerequisitesJSON),
			"skills":                 string(skillsJSON),
			"keywords":               string(keywordsJSON),
			"tags":                   string(tagsJSON),
			"metadata":               string(metadataJSON),
			"learningObjectives":     string(learningObjectivesJSON),
			"assessmentCriteria":     string(assessmentCriteriaJSON),
			"updatedAt":              node.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// GetNode 根据ID获取知识节点
func (r *KnowledgeGraphRepositoryImpl) GetNode(ctx context.Context, graphID, nodeID uuid.UUID) (*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (n:KnowledgeNode {id: $nodeID, graph_id: $graphID})
			RETURN n
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"nodeID":  nodeID.String(),
			"graphID": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, fmt.Errorf("node not found")
		}

		record := result.Record()
		nodeValue, ok := record.Get("n")
		if !ok {
			return nil, fmt.Errorf("node not found in result")
		}

		node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
		if err != nil {
			return nil, err
		}

		return node, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.KnowledgeNode), nil
}

// ValidateGraphStructure 验证图结构
func (r *KnowledgeGraphRepositoryImpl) ValidateGraphStructure(ctx context.Context, graphID uuid.UUID) (*repositories.GraphValidation, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		validation := &repositories.GraphValidation{
			IsValid:     true,
			Errors:      []repositories.ValidationError{},
			Warnings:    []repositories.ValidationWarning{},
			Statistics:  repositories.ValidationStatistics{},
			Suggestions: []string{},
		}

		// 获取基本统计信息
		statsQuery := `
			MATCH (g:KnowledgeGraph {id: $graphId})
			OPTIONAL MATCH (g)-[:CONTAINS]->(n:KnowledgeNode)
			OPTIONAL MATCH (n)-[r:RELATES_TO]->()
			RETURN count(DISTINCT n) as nodeCount, count(DISTINCT r) as relationCount
		`
		
		statsResult, err := tx.Run(ctx, statsQuery, map[string]interface{}{
			"graphId": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		if statsResult.Next(ctx) {
			record := statsResult.Record()
			validation.Statistics.TotalNodes = int(record.Values[0].(int64))
			validation.Statistics.TotalRelations = int(record.Values[1].(int64))
		}

		// 检查孤立节点
		orphanQuery := `
			MATCH (g:KnowledgeGraph {id: $graphId})-[:CONTAINS]->(n:KnowledgeNode)
			WHERE NOT (n)-[:RELATES_TO]-() AND NOT ()-[:RELATES_TO]->(n)
			RETURN count(n) as orphanCount
		`
		
		orphanResult, err := tx.Run(ctx, orphanQuery, map[string]interface{}{
			"graphId": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		if orphanResult.Next(ctx) {
			orphanCount := int(orphanResult.Record().Values[0].(int64))
			validation.Statistics.OrphanedNodes = orphanCount
			
			if orphanCount > 0 {
				validation.Warnings = append(validation.Warnings, repositories.ValidationWarning{
					Type:    "orphaned_nodes",
					Message: fmt.Sprintf("Found %d orphaned nodes", orphanCount),
				})
			}
		}

		// 检查循环依赖
		circularQuery := `
			MATCH (g:KnowledgeGraph {id: $graphId})-[:CONTAINS]->(n:KnowledgeNode)
			MATCH path = (n)-[:RELATES_TO*1..10]->(n)
			RETURN count(DISTINCT n) as circularCount
		`
		
		circularResult, err := tx.Run(ctx, circularQuery, map[string]interface{}{
			"graphId": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		if circularResult.Next(ctx) {
			circularCount := int(circularResult.Record().Values[0].(int64))
			validation.Statistics.CircularDependencies = circularCount
			
			if circularCount > 0 {
				validation.Errors = append(validation.Errors, repositories.ValidationError{
					Type:     "circular_dependency",
					Message:  fmt.Sprintf("Found %d nodes with circular dependencies", circularCount),
					Severity: "major",
				})
				validation.IsValid = false
			}
		}

		// 添加建议
		if validation.Statistics.OrphanedNodes > 0 {
			validation.Suggestions = append(validation.Suggestions, "Consider connecting orphaned nodes to the main graph structure")
		}
		
		if validation.Statistics.TotalNodes > 0 && validation.Statistics.TotalRelations == 0 {
			validation.Suggestions = append(validation.Suggestions, "Add relationships between nodes to create meaningful learning paths")
		}

		return validation, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate graph structure: %w", err)
	}

	return result.(*repositories.GraphValidation), nil
}

// SearchNodes 搜索知识节点
func (r *KnowledgeGraphRepositoryImpl) SearchNodes(ctx context.Context, graphID uuid.UUID, query *repositories.NodeSearchQuery) ([]*entities.KnowledgeNode, int, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// 构建搜索查询
	var whereConditions []string
	params := map[string]interface{}{
		"graphID": graphID.String(),
	}

	// 基本条件
	whereConditions = append(whereConditions, "n.graph_id = $graphID")

	// 文本搜索
	if query.Query != "" {
		whereConditions = append(whereConditions, "(n.name CONTAINS $query OR n.description CONTAINS $query)")
		params["query"] = query.Query
	}

	// 节点类型过滤
	if query.NodeType != nil {
		whereConditions = append(whereConditions, "n.node_type = $nodeType")
		params["nodeType"] = string(*query.NodeType)
	}

	// 难度级别过滤
	if query.DifficultyLevel != nil {
		whereConditions = append(whereConditions, "n.difficulty_level = $difficultyLevel")
		params["difficultyLevel"] = int64(*query.DifficultyLevel)
	}

	// 难度范围过滤
	if query.MinDifficulty != nil {
		whereConditions = append(whereConditions, "n.difficulty_level >= $minDifficulty")
		params["minDifficulty"] = int64(*query.MinDifficulty)
	}
	if query.MaxDifficulty != nil {
		whereConditions = append(whereConditions, "n.difficulty_level <= $maxDifficulty")
		params["maxDifficulty"] = int64(*query.MaxDifficulty)
	}

	// 标签过滤
	if len(query.Tags) > 0 {
		whereConditions = append(whereConditions, "ANY(tag IN $tags WHERE tag IN n.tags)")
		params["tags"] = query.Tags
	}

	// 关键词过滤
	if len(query.Keywords) > 0 {
		whereConditions = append(whereConditions, "ANY(keyword IN $keywords WHERE keyword IN n.keywords)")
		params["keywords"] = query.Keywords
	}

	// 时间范围过滤
	if query.CreatedAfter != nil {
		whereConditions = append(whereConditions, "n.created_at >= $createdAfter")
		params["createdAfter"] = *query.CreatedAfter
	}
	if query.CreatedBefore != nil {
		whereConditions = append(whereConditions, "n.created_at <= $createdBefore")
		params["createdBefore"] = *query.CreatedBefore
	}
	if query.UpdatedAfter != nil {
		whereConditions = append(whereConditions, "n.updated_at >= $updatedAfter")
		params["updatedAfter"] = *query.UpdatedAfter
	}
	if query.UpdatedBefore != nil {
		whereConditions = append(whereConditions, "n.updated_at <= $updatedBefore")
		params["updatedBefore"] = *query.UpdatedBefore
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 排序
	orderClause := "ORDER BY n.created_at DESC"
	if query.SortBy != "" {
		sortField := "n.created_at"
		switch query.SortBy {
		case "name":
			sortField = "n.name"
		case "difficulty":
			sortField = "n.difficulty_level"
		case "updated_at":
			sortField = "n.updated_at"
		case "importance":
			sortField = "n.importance"
		}
		
		sortOrder := "DESC"
		if query.SortOrder == "asc" {
			sortOrder = "ASC"
		}
		
		orderClause = fmt.Sprintf("ORDER BY %s %s", sortField, sortOrder)
	}

	// 分页参数
	params["offset"] = query.Offset
	params["limit"] = query.Limit

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 先获取总数
		countQuery := fmt.Sprintf(`
			MATCH (n:KnowledgeNode)
			%s
			RETURN count(n) as total
		`, whereClause)

		countResult, err := tx.Run(ctx, countQuery, params)
		if err != nil {
			return nil, err
		}

		var total int64
		if countResult.Next(ctx) {
			record := countResult.Record()
			if totalValue, ok := record.Get("total"); ok {
				total = totalValue.(int64)
			}
		}

		// 获取节点数据
		dataQuery := fmt.Sprintf(`
			MATCH (n:KnowledgeNode)
			%s
			RETURN n
			%s
			SKIP $offset
			LIMIT $limit
		`, whereClause, orderClause)

		dataResult, err := tx.Run(ctx, dataQuery, params)
		if err != nil {
			return nil, err
		}

		var nodes []*entities.KnowledgeNode
		for dataResult.Next(ctx) {
			record := dataResult.Record()
			nodeValue, found := record.Get("n")
			if !found {
				continue
			}

			nodeRecord, ok := nodeValue.(neo4j.Node)
			if !ok {
				continue
			}

			node, err := r.mapNodeFromNeo4j(nodeRecord)
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, node)
		}

		return map[string]interface{}{
			"nodes": nodes,
			"total": int(total),
		}, nil
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to search nodes: %w", err)
	}

	resultMap := result.(map[string]interface{})
	nodes := resultMap["nodes"].([]*entities.KnowledgeNode)
	total := resultMap["total"].(int)

	return nodes, total, nil
}

// RecommendNextNodes 推荐下一个学习节点
func (r *KnowledgeGraphRepositoryImpl) RecommendNextNodes(ctx context.Context, graphID, currentNodeID, learnerID uuid.UUID, limit int) ([]*repositories.NodeRecommendation, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (g:KnowledgeGraph {id: $graphID})
		MATCH (current:KnowledgeNode {id: $currentNodeID})-[:BELONGS_TO]->(g)
		OPTIONAL MATCH (current)-[r:PREREQUISITE_OF|RELATED_TO|PART_OF]->(next:KnowledgeNode)-[:BELONGS_TO]->(g)
		WHERE next.id <> current.id
		OPTIONAL MATCH (learner:Learner {id: $learnerID})-[progress:LEARNED]->(completed:KnowledgeNode)-[:BELONGS_TO]->(g)
		WITH current, next, r, collect(DISTINCT completed.id) as completedNodes
		WHERE next IS NOT NULL AND NOT next.id IN completedNodes
		WITH next, r, current, completedNodes,
			 CASE 
				WHEN type(r) = 'PREREQUISITE_OF' THEN 3.0
				WHEN type(r) = 'RELATED_TO' THEN 2.0
				WHEN type(r) = 'PART_OF' THEN 2.5
				ELSE 1.0
			 END as relationScore
		OPTIONAL MATCH (prereq:KnowledgeNode)-[:PREREQUISITE_OF]->(next)
		WITH next, relationScore, current, completedNodes, collect(DISTINCT prereq.id) as prerequisites
		WITH next, relationScore, current, completedNodes, prerequisites,
			 CASE 
				WHEN size(prerequisites) = 0 THEN true
				WHEN all(p IN prerequisites WHERE p IN completedNodes) THEN true
				ELSE false
			 END as prerequisitesMet
		WITH next, relationScore, prerequisitesMet,
			 CASE 
				WHEN abs(next.difficulty_level - current.difficulty_level) <= 1 THEN 1.0
				WHEN abs(next.difficulty_level - current.difficulty_level) = 2 THEN 0.7
				ELSE 0.4
			 END as difficultyMatch
		WITH next, relationScore, prerequisitesMet, difficultyMatch,
			 (relationScore * 0.4 + 
			  (CASE WHEN prerequisitesMet THEN 1.0 ELSE 0.2 END) * 0.3 + 
			  difficultyMatch * 0.3) as finalScore
		RETURN next, finalScore, prerequisitesMet, difficultyMatch
		ORDER BY finalScore DESC
		LIMIT $limit
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"graphID":       graphID.String(),
		"currentNodeID": currentNodeID.String(),
		"learnerID":     learnerID.String(),
		"limit":         limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query next node recommendations: %w", err)
	}

	var recommendations []*repositories.NodeRecommendation
	for result.Next(ctx) {
		record := result.Record()
		
		nodeValue, ok := record.Get("next")
		if !ok {
			continue
		}
		
		node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
		if err != nil {
			continue
		}

		finalScore, _ := record.Get("finalScore")
		prerequisitesMet, _ := record.Get("prerequisitesMet")
		difficultyMatch, _ := record.Get("difficultyMatch")

		reasoning := []string{"基于学习路径推荐"}
		if prerequisitesMet.(bool) {
			reasoning = append(reasoning, "前置条件已满足")
		}

		estimatedTime := 30 * time.Minute
		switch node.DifficultyLevel {
		case entities.DifficultyIntermediate:
			estimatedTime = 60 * time.Minute
		case entities.DifficultyAdvanced:
			estimatedTime = 90 * time.Minute
		case entities.DifficultyExpert:
			estimatedTime = 120 * time.Minute
		}

		priority := 3
		if finalScore.(float64) >= 2.5 {
			priority = 1
		} else if finalScore.(float64) >= 1.5 {
			priority = 2
		}

		recommendation := &repositories.NodeRecommendation{
			Node:                node,
			RecommendationScore: finalScore.(float64),
			Reasoning:           reasoning,
			EstimatedTime:       estimatedTime,
			DifficultyMatch:     difficultyMatch.(float64),
			PrerequisitesMet:    prerequisitesMet.(bool),
			LearningStyle:       entities.LearningStyleVisual,
			Priority:            priority,
		}

		recommendations = append(recommendations, recommendation)
	}

	if err = result.Err(); err != nil {
		return nil, fmt.Errorf("failed to process next node recommendations: %w", err)
	}

	return recommendations, nil
}

// GetNodeImportance 获取节点重要性
func (r *KnowledgeGraphRepositoryImpl) GetNodeImportance(ctx context.Context, graphID uuid.UUID) (map[uuid.UUID]float64, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode)
			OPTIONAL MATCH (n)-[r:RELATES_TO]-()
			WITH n, count(r) as degree
			RETURN n.id as nodeID, 
				   CASE 
					   WHEN degree = 0 THEN 0.1
					   ELSE toFloat(degree) / 10.0
				   END as importance
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		importance := make(map[uuid.UUID]float64)
		for result.Next(ctx) {
			record := result.Record()
			nodeIDStr, _ := record.Get("nodeID")
			importanceValue, _ := record.Get("importance")

			nodeID, err := uuid.Parse(nodeIDStr.(string))
			if err != nil {
				continue
			}

			importance[nodeID] = importanceValue.(float64)
		}

		return importance, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node importance: %w", err)
	}

	return result.(map[uuid.UUID]float64), nil
}

// GetNodeNeighbors 获取节点邻居
func (r *KnowledgeGraphRepositoryImpl) GetNodeNeighbors(ctx context.Context, graphID, nodeID uuid.UUID, depth int) ([]*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(start:KnowledgeNode {id: $nodeID})
			MATCH (start)-[:RELATES_TO*1..$depth]-(neighbor:KnowledgeNode)
			WHERE (g)-[:CONTAINS]->(neighbor)
			RETURN DISTINCT neighbor
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
			"nodeID":  nodeID.String(),
			"depth":   depth,
		})
		if err != nil {
			return nil, err
		}

		var neighbors []*entities.KnowledgeNode
		for result.Next(ctx) {
			record := result.Record()
			nodeValue, found := record.Get("neighbor")
			if !found {
				continue
			}

			node, ok := nodeValue.(neo4j.Node)
			if !ok {
				continue
			}

			neighbor, err := r.mapNodeFromNeo4j(node)
			if err != nil {
				continue
			}

			neighbors = append(neighbors, neighbor)
		}

		return neighbors, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node neighbors: %w", err)
	}

	return result.([]*entities.KnowledgeNode), nil
}

// GetNodeRelations 获取节点关系
func (r *KnowledgeGraphRepositoryImpl) GetNodeRelations(ctx context.Context, graphID, nodeID uuid.UUID, relationType *entities.RelationType) ([]*entities.KnowledgeRelation, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		var query string
		params := map[string]interface{}{
			"graphID": graphID.String(),
			"nodeID":  nodeID.String(),
		}

		if relationType != nil {
			query = `
				MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode {id: $nodeID})
				MATCH (n)-[r:RELATES_TO {type: $relationType}]-(other:KnowledgeNode)
				WHERE (g)-[:CONTAINS]->(other)
				RETURN r, n, other
			`
			params["relationType"] = string(*relationType)
		} else {
			query = `
				MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode {id: $nodeID})
				MATCH (n)-[r:RELATES_TO]-(other:KnowledgeNode)
				WHERE (g)-[:CONTAINS]->(other)
				RETURN r, n, other
			`
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var relations []*entities.KnowledgeRelation
		for result.Next(ctx) {
			record := result.Record()
			relValue, found := record.Get("r")
			if !found {
				continue
			}

			rel, ok := relValue.(neo4j.Relationship)
			if !ok {
				continue
			}

			relation, err := r.mapRelationshipFromNeo4j(rel)
			if err != nil {
				continue
			}

			relations = append(relations, relation)
		}

		return relations, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node relations: %w", err)
	}

	return result.([]*entities.KnowledgeRelation), nil
}

// GetPersonalizedPaths 获取个性化学习路径
func (r *KnowledgeGraphRepositoryImpl) GetPersonalizedPaths(ctx context.Context, graphID, learnerID uuid.UUID, targetNodeID uuid.UUID) ([]*entities.LearningPath, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(target:KnowledgeNode {id: $targetNodeID})
			MATCH (learner:Learner {id: $learnerID})
			OPTIONAL MATCH (learner)-[:HAS_MASTERED]->(mastered:KnowledgeNode)
			WHERE (g)-[:CONTAINS]->(mastered)
			
			// 查找从已掌握的知识点到目标知识点的路径
			MATCH path = shortestPath((mastered)-[:RELATES_TO*1..10]-(target))
			WHERE ALL(n IN nodes(path) WHERE (g)-[:CONTAINS]->(n))
			
			WITH collect(DISTINCT path) as paths, learner, target
			
			// 如果没有从已掌握知识点的路径，则查找从基础知识点的路径
			OPTIONAL MATCH (basic:KnowledgeNode {difficulty_level: 'beginner'})
			WHERE (g)-[:CONTAINS]->(basic) AND NOT (learner)-[:HAS_MASTERED]->(basic)
			OPTIONAL MATCH basicPath = shortestPath((basic)-[:RELATES_TO*1..10]-(target))
			WHERE ALL(n IN nodes(basicPath) WHERE (g)-[:CONTAINS]->(n))
			
			WITH CASE 
				WHEN size(paths) > 0 THEN paths 
				ELSE collect(DISTINCT basicPath) 
			END as finalPaths
			
			UNWIND finalPaths as path
			RETURN path
			LIMIT 5
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":      graphID.String(),
			"learnerID":    learnerID.String(),
			"targetNodeID": targetNodeID.String(),
		})
		if err != nil {
			return nil, err
		}

		var paths []*entities.LearningPath
		pathIndex := 0
		for result.Next(ctx) {
			record := result.Record()
			pathValue, found := record.Get("path")
			if !found {
				continue
			}

			path, ok := pathValue.(neo4j.Path)
			if !ok {
				continue
			}

			// 创建学习路径
			pathNodes := path.Nodes
			learningPath := &entities.LearningPath{
				ID:              uuid.New(),
				Name:            fmt.Sprintf("个性化路径 %d", pathIndex+1),
				Description:     "基于学习者当前掌握情况生成的个性化学习路径",
				Subject:         "个性化学习",
				DifficultyLevel: entities.DifficultyIntermediate,
				EstimatedHours:  float64(len(pathNodes)) * 2.0, // 每个节点估计2小时
				Prerequisites:   []uuid.UUID{},
				LearningGoals:   []string{"达到目标知识点"},
				Nodes:           []entities.PathNode{},
				Milestones:      []entities.Milestone{},
				Tags:            []string{"个性化", "推荐"},
				IsPublic:        false,
				CreatedBy:       learnerID,
				EnrollmentCount: 0,
				CompletionRate:  0.0,
				Rating:          0.0,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			// 添加路径中的节点
			for i, node := range pathNodes {
				if nodeID, ok := node.Props["id"].(string); ok {
					if id, err := uuid.Parse(nodeID); err == nil {
						pathNode := entities.PathNode{
							ID:          uuid.New(),
							KnowledgeID: id,
							Order:       i + 1,
							IsOptional:  false,
							Dependencies: []uuid.UUID{},
							Metadata:    make(map[string]interface{}),
						}
						learningPath.Nodes = append(learningPath.Nodes, pathNode)
					}
				}
			}

			paths = append(paths, learningPath)
			pathIndex++
		}

		return paths, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get personalized paths: %w", err)
	}

	return result.([]*entities.LearningPath), nil
}

// GetPrerequisites 获取前置条件
func (r *KnowledgeGraphRepositoryImpl) GetPrerequisites(ctx context.Context, graphID, nodeID uuid.UUID, depth int) ([]*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(target:KnowledgeNode {id: $nodeID})
			MATCH (target)<-[:RELATES_TO*1..$depth]-(prerequisite:KnowledgeNode)
			WHERE (g)-[:CONTAINS]->(prerequisite) 
			  AND prerequisite.difficulty_level <= target.difficulty_level
			RETURN DISTINCT prerequisite
			ORDER BY prerequisite.difficulty_level ASC
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
			"nodeID":  nodeID.String(),
			"depth":   depth,
		})
		if err != nil {
			return nil, err
		}

		var prerequisites []*entities.KnowledgeNode
		for result.Next(ctx) {
			record := result.Record()
			nodeValue, found := record.Get("prerequisite")
			if !found {
				continue
			}

			node, ok := nodeValue.(neo4j.Node)
			if !ok {
				continue
			}

			prerequisite, err := r.mapNodeFromNeo4j(node)
			if err != nil {
				continue
			}

			prerequisites = append(prerequisites, prerequisite)
		}

		return prerequisites, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get prerequisites: %w", err)
	}

	return result.([]*entities.KnowledgeNode), nil
}

// GetRelation 获取知识关系
func (r *KnowledgeGraphRepositoryImpl) GetRelation(ctx context.Context, graphID, relationID uuid.UUID) (*entities.KnowledgeRelation, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			MATCH (from:KnowledgeNode)-[r:RELATES_TO {id: $relationID}]->(to:KnowledgeNode)
			WHERE (g)-[:CONTAINS]->(from) AND (g)-[:CONTAINS]->(to)
			RETURN r
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":    graphID.String(),
			"relationID": relationID.String(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			relValue, found := record.Get("r")
			if !found {
				return nil, fmt.Errorf("relation not found in result")
			}

			rel, ok := relValue.(neo4j.Relationship)
			if !ok {
				return nil, fmt.Errorf("invalid relation type")
			}

			return r.mapRelationshipFromNeo4j(rel)
		}

		return nil, fmt.Errorf("relation with ID '%s' not found", relationID.String())
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get relation: %w", err)
	}

	return result.(*entities.KnowledgeRelation), nil
}

// GetRelationsBetween 获取两个节点之间的关系
func (r *KnowledgeGraphRepositoryImpl) GetRelationsBetween(ctx context.Context, graphID, fromNodeID, toNodeID uuid.UUID) ([]*entities.KnowledgeRelation, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			MATCH (from:KnowledgeNode {id: $fromNodeID})-[r:RELATES_TO]->(to:KnowledgeNode {id: $toNodeID})
			WHERE (g)-[:CONTAINS]->(from) AND (g)-[:CONTAINS]->(to)
			RETURN r
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":    graphID.String(),
			"fromNodeID": fromNodeID.String(),
			"toNodeID":   toNodeID.String(),
		})
		if err != nil {
			return nil, err
		}

		var relations []*entities.KnowledgeRelation
		for result.Next(ctx) {
			record := result.Record()
			relValue, found := record.Get("r")
			if !found {
				continue
			}

			rel, ok := relValue.(neo4j.Relationship)
			if !ok {
				continue
			}

			relation, err := r.mapRelationshipFromNeo4j(rel)
			if err != nil {
				return nil, err
			}

			relations = append(relations, relation)
		}

		return relations, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get relations between nodes: %w", err)
	}

	return result.([]*entities.KnowledgeRelation), nil
}

// ImportGraph 导入图谱
func (r *KnowledgeGraphRepositoryImpl) ImportGraph(ctx context.Context, data []byte, format string) (*entities.KnowledgeGraph, error) {
	// 创建新的知识图谱
	graph := &entities.KnowledgeGraph{
		ID:          uuid.New(),
		Name:        "Imported Graph",
		Description: "Graph imported from " + format + " format",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Nodes:       []entities.KnowledgeNode{},
		Relations:   []entities.KnowledgeRelation{},
	}

	// 根据格式解析数据
	switch format {
	case "json":
		if err := r.importFromJSON(ctx, graph, data); err != nil {
			return nil, fmt.Errorf("failed to import from JSON: %w", err)
		}
	case "csv":
		if err := r.importFromCSV(ctx, graph, data); err != nil {
			return nil, fmt.Errorf("failed to import from CSV: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// 保存图谱到数据库
	if err := r.CreateGraph(ctx, graph); err != nil {
		return nil, fmt.Errorf("failed to save imported graph: %w", err)
	}

	return graph, nil
}

// importFromJSON 从JSON格式导入
func (r *KnowledgeGraphRepositoryImpl) importFromJSON(ctx context.Context, graph *entities.KnowledgeGraph, data []byte) error {
	var importData struct {
		Nodes []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Type        string `json:"type"`
		} `json:"nodes"`
		Relations []struct {
			ID       string `json:"id"`
			FromID   string `json:"from_id"`
			ToID     string `json:"to_id"`
			Type     string `json:"type"`
			Strength float64 `json:"strength"`
		} `json:"relations"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// 导入节点
	for _, nodeData := range importData.Nodes {
		nodeID, err := uuid.Parse(nodeData.ID)
		if err != nil {
			nodeID = uuid.New()
		}

		node := entities.KnowledgeNode{
			ID:          nodeID,
			Name:        nodeData.Title,
			Description: nodeData.Description,
			Type:        entities.NodeType(nodeData.Type),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		graph.Nodes = append(graph.Nodes, node)
	}

	// 导入关系
	for _, relData := range importData.Relations {
		relID, err := uuid.Parse(relData.ID)
		if err != nil {
			relID = uuid.New()
		}

		fromID, err := uuid.Parse(relData.FromID)
		if err != nil {
			continue
		}

		toID, err := uuid.Parse(relData.ToID)
		if err != nil {
			continue
		}

		relation := entities.KnowledgeRelation{
			ID:         relID,
			FromNodeID: fromID,
			ToNodeID:   toID,
			Type:       entities.RelationType(relData.Type),
			Weight:     relData.Strength,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		graph.Relations = append(graph.Relations, relation)
	}

	return nil
}

// importFromCSV 从CSV格式导入
func (r *KnowledgeGraphRepositoryImpl) importFromCSV(ctx context.Context, graph *entities.KnowledgeGraph, data []byte) error {
	// 简单的CSV解析实现
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("invalid CSV format")
	}

	// 假设第一行是标题行，跳过
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			continue
		}

		// 创建节点
		node := entities.KnowledgeNode{
			ID:          uuid.New(),
			Name:        strings.TrimSpace(fields[0]),
			Description: strings.TrimSpace(fields[1]),
			Type:        entities.NodeType(strings.TrimSpace(fields[2])),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		graph.Nodes = append(graph.Nodes, node)
	}

	return nil
}

// ListGraphs 列出知识图谱
func (r *KnowledgeGraphRepositoryImpl) ListGraphs(ctx context.Context, offset, limit int) ([]*entities.KnowledgeGraph, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph)
			RETURN g
			ORDER BY g.created_at DESC
			SKIP $offset
			LIMIT $limit
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"offset": offset,
			"limit":  limit,
		})
		if err != nil {
			return nil, err
		}

		var graphs []*entities.KnowledgeGraph
		for result.Next(ctx) {
			record := result.Record()
			graphValue, found := record.Get("g")
			if !found {
				continue
			}

			graphNode, ok := graphValue.(neo4j.Node)
			if !ok {
				continue
			}

			graph, err := r.mapGraphFromNeo4j(graphNode)
			if err != nil {
				return nil, err
			}

			graphs = append(graphs, graph)
		}

		return graphs, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list graphs: %w", err)
	}

	return result.([]*entities.KnowledgeGraph), nil
}

// ListLearningPaths 列出学习路径
func (r *KnowledgeGraphRepositoryImpl) ListLearningPaths(ctx context.Context, graphID uuid.UUID, offset, limit int) ([]*entities.LearningPath, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(p:LearningPath)
			RETURN p
			ORDER BY p.created_at DESC
			SKIP $offset
			LIMIT $limit
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
			"offset":  offset,
			"limit":   limit,
		})
		if err != nil {
			return nil, err
		}

		var paths []*entities.LearningPath
		for result.Next(ctx) {
			record := result.Record()
			pathValue, found := record.Get("p")
			if !found {
				continue
			}

			pathNode, ok := pathValue.(neo4j.Node)
			if !ok {
				continue
			}

			path, err := r.mapLearningPathFromNeo4j(pathNode)
			if err != nil {
				return nil, err
			}

			paths = append(paths, path)
		}

		return paths, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list learning paths: %w", err)
	}

	return result.([]*entities.LearningPath), nil
}

// ListNodes 列出知识节点
func (r *KnowledgeGraphRepositoryImpl) ListNodes(ctx context.Context, graphID uuid.UUID, nodeType *entities.NodeType, offset, limit int) ([]*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		var query string
		params := map[string]interface{}{
			"graphID": graphID.String(),
			"offset":  offset,
			"limit":   limit,
		}

		if nodeType != nil {
			query = `
				MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode {type: $nodeType})
				RETURN n
				ORDER BY n.created_at DESC
				SKIP $offset
				LIMIT $limit
			`
			params["nodeType"] = string(*nodeType)
		} else {
			query = `
				MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode)
				RETURN n
				ORDER BY n.created_at DESC
				SKIP $offset
				LIMIT $limit
			`
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var nodes []*entities.KnowledgeNode
		for result.Next(ctx) {
			record := result.Record()
			nodeValue, found := record.Get("n")
			if !found {
				continue
			}

			nodeRecord, ok := nodeValue.(neo4j.Node)
			if !ok {
				continue
			}

			node, err := r.mapNodeFromNeo4j(nodeRecord)
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, node)
		}

		return nodes, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return result.([]*entities.KnowledgeNode), nil
}

// PredictLearningDifficulty 预测学习难度
func (r *KnowledgeGraphRepositoryImpl) PredictLearningDifficulty(ctx context.Context, graphID, nodeID, learnerID uuid.UUID) (*repositories.DifficultyPrediction, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 获取节点信息和学习者历史数据
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode {id: $nodeID})
			OPTIONAL MATCH (l:Learner {id: $learnerID})-[h:LEARNED]->(learned:KnowledgeNode)
			OPTIONAL MATCH (n)<-[:PREREQUISITE]-(prereq:KnowledgeNode)
			OPTIONAL MATCH (l)-[a:ATTEMPTED]->(n)
			RETURN n, 
				   collect(DISTINCT learned) as learnedNodes,
				   collect(DISTINCT prereq) as prerequisites,
				   collect(DISTINCT a) as attempts
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":   graphID.String(),
			"nodeID":    nodeID.String(),
			"learnerID": learnerID.String(),
		})
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, fmt.Errorf("node not found")
		}

		record := result.Record()
		nodeValue, _ := record.Get("n")
		learnedNodes, _ := record.Get("learnedNodes")
		prerequisites, _ := record.Get("prerequisites")
		attempts, _ := record.Get("attempts")

		node, ok := nodeValue.(neo4j.Node)
		if !ok {
			return nil, fmt.Errorf("invalid node data")
		}

		// 基础难度评估
		baseDifficulty := 0.5 // 默认中等难度
		if diffProp, exists := node.Props["difficulty"]; exists {
			if diff, ok := diffProp.(float64); ok {
				baseDifficulty = diff / 10.0 // 假设难度范围是0-10
			}
		}

		// 计算先决条件满足度
		prereqList, _ := prerequisites.([]interface{})
		learnedList, _ := learnedNodes.([]interface{})
		prereqSatisfaction := r.calculatePrerequisiteSatisfaction(prereqList, learnedList)

		// 计算历史表现
		attemptsList, _ := attempts.([]interface{})
		historicalPerformance := r.calculateHistoricalPerformance(attemptsList)

		// 综合计算预测难度
		predictedDifficulty := baseDifficulty * (1.0 + (1.0-prereqSatisfaction)*0.5) * (1.0 + (1.0-historicalPerformance)*0.3)
		if predictedDifficulty > 1.0 {
			predictedDifficulty = 1.0
		}

		// 计算置信度
		confidence := 0.7 // 基础置信度
		if len(attemptsList) > 0 {
			confidence += 0.2 // 有历史数据增加置信度
		}
		if len(prereqList) > 0 {
			confidence += 0.1 // 有先决条件信息增加置信度
		}

		// 估算学习时间
		estimatedMinutes := int(60 * (1 + predictedDifficulty)) // 基础1小时，根据难度调整
		estimatedTime := time.Duration(estimatedMinutes) * time.Minute

		// 计算成功概率
		successProbability := 1.0 - predictedDifficulty*0.8

		prediction := &repositories.DifficultyPrediction{
			NodeID:              nodeID,
			LearnerID:           learnerID,
			PredictedDifficulty: predictedDifficulty,
			Confidence:          confidence,
			EstimatedTime:       estimatedTime,
			SuccessProbability:  successProbability,
			RecommendedPrep:     []*entities.KnowledgeNode{},
			RiskFactors:         r.identifyRiskFactors(prereqSatisfaction, historicalPerformance),
			SupportResources:    []string{"在线教程", "练习题库", "专家指导"},
		}

		return prediction, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to predict learning difficulty: %w", err)
	}

	return result.(*repositories.DifficultyPrediction), nil
}

// calculatePrerequisiteSatisfaction 计算先决条件满足度
func (r *KnowledgeGraphRepositoryImpl) calculatePrerequisiteSatisfaction(prerequisites, learnedNodes []interface{}) float64 {
	if len(prerequisites) == 0 {
		return 1.0 // 没有先决条件，满足度为100%
	}

	learnedSet := make(map[string]bool)
	for _, learned := range learnedNodes {
		if node, ok := learned.(neo4j.Node); ok {
			if id, exists := node.Props["id"]; exists {
				learnedSet[id.(string)] = true
			}
		}
	}

	satisfied := 0
	for _, prereq := range prerequisites {
		if node, ok := prereq.(neo4j.Node); ok {
			if id, exists := node.Props["id"]; exists {
				if learnedSet[id.(string)] {
					satisfied++
				}
			}
		}
	}

	return float64(satisfied) / float64(len(prerequisites))
}

// calculateHistoricalPerformance 计算历史表现
func (r *KnowledgeGraphRepositoryImpl) calculateHistoricalPerformance(attempts []interface{}) float64 {
	if len(attempts) == 0 {
		return 0.5 // 没有历史数据，返回中性值
	}

	totalScore := 0.0
	for _, attempt := range attempts {
		if rel, ok := attempt.(neo4j.Relationship); ok {
			if score, exists := rel.Props["score"]; exists {
				if s, ok := score.(float64); ok {
					totalScore += s
				}
			}
		}
	}

	return totalScore / float64(len(attempts)) / 100.0 // 假设分数范围是0-100
}

// identifyRiskFactors 识别风险因素
func (r *KnowledgeGraphRepositoryImpl) identifyRiskFactors(prereqSatisfaction, historicalPerformance float64) []string {
	var risks []string

	if prereqSatisfaction < 0.7 {
		risks = append(risks, "先决条件掌握不足")
	}

	if historicalPerformance < 0.6 {
		risks = append(risks, "历史学习表现较差")
	}

	if prereqSatisfaction < 0.5 && historicalPerformance < 0.5 {
		risks = append(risks, "基础薄弱，建议加强预习")
	}

	return risks
}

// RecommendLearningPaths 推荐学习路径
func (r *KnowledgeGraphRepositoryImpl) RecommendLearningPaths(ctx context.Context, graphID, learnerID uuid.UUID, targetSkills []string, limit int) ([]*repositories.PathRecommendation, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 查找与目标技能相关的学习路径
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(p:LearningPath)
			OPTIONAL MATCH (l:Learner {id: $learnerID})-[h:COMPLETED]->(completedPath:LearningPath)
			OPTIONAL MATCH (p)-[:TARGETS]->(skill:KnowledgeNode)
			WHERE skill.name IN $targetSkills OR skill.tags IN $targetSkills
			WITH p, collect(DISTINCT skill) as targetNodes, collect(DISTINCT completedPath) as completedPaths
			WHERE NOT p IN completedPaths
			RETURN p, targetNodes
			ORDER BY size(targetNodes) DESC
			LIMIT $limit
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":      graphID.String(),
			"learnerID":    learnerID.String(),
			"targetSkills": targetSkills,
			"limit":        limit,
		})
		if err != nil {
			return nil, err
		}

		var recommendations []*repositories.PathRecommendation
		for result.Next(ctx) {
			record := result.Record()
			pathValue, found := record.Get("p")
			if !found {
				continue
			}

			pathNode, ok := pathValue.(neo4j.Node)
			if !ok {
				continue
			}

			path, err := r.mapLearningPathFromNeo4j(pathNode)
			if err != nil {
				continue
			}

			targetNodes, _ := record.Get("targetNodes")
			targetNodesList, _ := targetNodes.([]interface{})

			// 计算推荐分数
			recommendationScore := r.calculatePathRecommendationScore(path, targetNodesList, targetSkills)

			// 估算学习时长
			estimatedDuration := r.estimatePathDuration(path)

			// 计算技能覆盖度
			skillCoverage := r.calculateSkillCoverage(targetNodesList, targetSkills)

			// 生成推荐理由
			reasoning := r.generatePathRecommendationReasoning(path, targetNodesList, recommendationScore)

			recommendation := &repositories.PathRecommendation{
				Path:                path,
				RecommendationScore: recommendationScore,
				Reasoning:           reasoning,
				EstimatedDuration:   estimatedDuration,
				DifficultyProgression: []float64{0.3, 0.5, 0.7, 0.9}, // 示例难度递进
				SkillCoverage:       skillCoverage,
				PersonalizationScore: 0.8, // 个性化分数
				SuccessProbability:  0.75, // 成功概率
			}

			recommendations = append(recommendations, recommendation)
		}

		// 按推荐分数排序
		for i := 0; i < len(recommendations)-1; i++ {
			for j := i + 1; j < len(recommendations); j++ {
				if recommendations[i].RecommendationScore < recommendations[j].RecommendationScore {
					recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
				}
			}
		}

		return recommendations, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to recommend learning paths: %w", err)
	}

	return result.([]*repositories.PathRecommendation), nil
}

// calculatePathRecommendationScore 计算路径推荐分数
func (r *KnowledgeGraphRepositoryImpl) calculatePathRecommendationScore(path *entities.LearningPath, targetNodes []interface{}, targetSkills []string) float64 {
	baseScore := 0.5

	// 根据目标节点数量调整分数
	if len(targetNodes) > 0 {
		baseScore += float64(len(targetNodes)) * 0.1
	}

	// 根据路径难度调整分数
	switch path.DifficultyLevel {
	case entities.DifficultyBeginner:
		baseScore += 0.2
	case entities.DifficultyIntermediate:
		baseScore += 0.1
	case entities.DifficultyAdvanced:
		baseScore -= 0.1
	}

	// 确保分数在0-1范围内
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0.0 {
		baseScore = 0.0
	}

	return baseScore
}

// estimatePathDuration 估算路径学习时长
func (r *KnowledgeGraphRepositoryImpl) estimatePathDuration(path *entities.LearningPath) time.Duration {
	// 基础时长：根据路径中的节点数量估算
	baseHours := len(path.Nodes) * 2 // 每个节点平均2小时

	// 根据难度调整
	switch path.DifficultyLevel {
	case entities.DifficultyBeginner:
		baseHours = int(float64(baseHours) * 0.8)
	case entities.DifficultyAdvanced:
		baseHours = int(float64(baseHours) * 1.5)
	}

	return time.Duration(baseHours) * time.Hour
}

// calculateSkillCoverage 计算技能覆盖度
func (r *KnowledgeGraphRepositoryImpl) calculateSkillCoverage(targetNodes []interface{}, targetSkills []string) map[string]float64 {
	coverage := make(map[string]float64)

	for _, skill := range targetSkills {
		coverage[skill] = 0.0
	}

	// 计算每个技能的覆盖度
	for _, nodeInterface := range targetNodes {
		if node, ok := nodeInterface.(neo4j.Node); ok {
			if name, exists := node.Props["name"]; exists {
				if skillName, ok := name.(string); ok {
					for _, targetSkill := range targetSkills {
						if skillName == targetSkill {
							coverage[targetSkill] = 1.0
						}
					}
				}
			}
		}
	}

	return coverage
}

// generatePathRecommendationReasoning 生成路径推荐理由
func (r *KnowledgeGraphRepositoryImpl) generatePathRecommendationReasoning(path *entities.LearningPath, targetNodes []interface{}, score float64) []string {
	var reasoning []string

	if score > 0.8 {
		reasoning = append(reasoning, "高度匹配您的学习目标")
	} else if score > 0.6 {
		reasoning = append(reasoning, "较好匹配您的学习需求")
	} else {
		reasoning = append(reasoning, "部分匹配您的学习目标")
	}

	if len(targetNodes) > 0 {
		reasoning = append(reasoning, fmt.Sprintf("涵盖%d个相关技能点", len(targetNodes)))
	}

	switch path.DifficultyLevel {
	case entities.DifficultyBeginner:
		reasoning = append(reasoning, "适合初学者，循序渐进")
	case entities.DifficultyIntermediate:
		reasoning = append(reasoning, "中等难度，适合有基础的学习者")
	case entities.DifficultyAdvanced:
		reasoning = append(reasoning, "高级内容，适合深入学习")
	}

	return reasoning
}

// CreateConceptMap 创建概念图
func (r *KnowledgeGraphRepositoryImpl) CreateConceptMap(ctx context.Context, conceptMap *entities.ConceptMap) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (cm:ConceptMap {
				id: $id,
				name: $name,
				description: $description,
				graph_id: $graphID,
				created_by: $createdBy,
				created_at: $createdAt,
				updated_at: $updatedAt
			})
			RETURN cm
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          conceptMap.ID.String(),
			"name":        conceptMap.Name,
			"description": conceptMap.Description,
			"graphID":     conceptMap.GraphID.String(),
			"createdBy":   conceptMap.CreatedBy.String(),
			"createdAt":   conceptMap.CreatedAt,
			"updatedAt":   conceptMap.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// UpdateConceptMap 更新概念图
func (r *KnowledgeGraphRepositoryImpl) UpdateConceptMap(ctx context.Context, conceptMap *entities.ConceptMap) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (cm:ConceptMap {id: $id})
			SET cm.name = $name,
				cm.description = $description,
				cm.graph_id = $graphID,
				cm.updated_at = $updatedAt
			RETURN cm
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          conceptMap.ID.String(),
			"name":        conceptMap.Name,
			"description": conceptMap.Description,
			"graphID":     conceptMap.GraphID.String(),
			"updatedAt":   conceptMap.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// DeleteConceptMap 删除概念图
func (r *KnowledgeGraphRepositoryImpl) DeleteConceptMap(ctx context.Context, conceptMapID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (cm:ConceptMap {id: $conceptMapID})
			DETACH DELETE cm
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"conceptMapID": conceptMapID.String(),
		})

		return nil, err
	})

	return err
}

// CreateGraphVersion 创建图谱版本
func (r *KnowledgeGraphRepositoryImpl) CreateGraphVersion(ctx context.Context, graphID uuid.UUID, version *repositories.GraphVersion) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (gv:GraphVersion {
				id: $id,
				graph_id: $graphID,
				version_number: $versionNumber,
				description: $description,
				created_by: $createdBy,
				created_at: $createdAt
			})
			RETURN gv
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":            version.ID.String(),
			"graphID":       version.GraphID.String(),
			"version": version.Version,
			"description":   version.Description,
			"createdBy":     version.CreatedBy.String(),
			"createdAt":     version.CreatedAt,
		})

		return nil, err
	})

	return err
}

// RestoreGraphVersion 恢复图谱版本
func (r *KnowledgeGraphRepositoryImpl) RestoreGraphVersion(ctx context.Context, graphID, versionID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 获取版本快照数据
		versionQuery := `
			MATCH (gv:GraphVersion {id: $versionID, graph_id: $graphID})
			RETURN gv.snapshot as snapshot
		`

		versionResult, err := tx.Run(ctx, versionQuery, map[string]interface{}{
			"versionID": versionID.String(),
			"graphID":   graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		if !versionResult.Next(ctx) {
			return nil, fmt.Errorf("graph version not found")
		}

		record := versionResult.Record()
		snapshotValue, found := record.Get("snapshot")
		if !found {
			return nil, fmt.Errorf("version snapshot not found")
		}

		// 这里应该解析快照数据并恢复图谱状态
		// 为了简化实现，我们只是记录恢复操作
		_ = snapshotValue

		// 更新图谱的当前版本标记
		updateQuery := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			SET g.current_version = $versionID, g.updated_at = datetime()
			RETURN g
		`

		_, err = tx.Run(ctx, updateQuery, map[string]interface{}{
			"graphID":   graphID.String(),
			"versionID": versionID.String(),
		})

		return nil, err
	})

	return err
}

// CompareGraphVersions 比较图谱版本
func (r *KnowledgeGraphRepositoryImpl) CompareGraphVersions(ctx context.Context, graphID, version1ID, version2ID uuid.UUID) (*repositories.GraphComparison, error) {
	// 这里应该实现版本比较逻辑
	// 为了简化，返回一个空的比较结果
	return &repositories.GraphComparison{
		Version1ID: version1ID,
		Version2ID: version2ID,
	}, nil
}

// getIntFromProps 从属性中安全获取整数值
func getIntFromProps(props map[string]interface{}, key string) int {
	if val, ok := props[key]; ok {
		switch v := val.(type) {
		case int64:
			return int(v)
		case int:
			return v
		case float64:
			return int(v)
		default:
			return 0
		}
	}
	return 0
}

// getFloatFromProps 从属性中安全获取浮点数值
func getFloatFromProps(props map[string]interface{}, key string) float64 {
	if val, ok := props[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int64:
			return float64(v)
		case int:
			return float64(v)
		default:
			return 0.0
		}
	}
	return 0.0
}

// getDifficultyLevelFromProps 从属性中安全获取难度级别
func getDifficultyLevelFromProps(props map[string]interface{}, key string) entities.DifficultyLevel {
	if val, ok := props[key]; ok {
		if str, ok := val.(string); ok {
			switch str {
			case "beginner":
				return entities.DifficultyBeginner
			case "elementary":
				return entities.DifficultyElementary
			case "intermediate":
				return entities.DifficultyIntermediate
			case "advanced":
				return entities.DifficultyAdvanced
			case "expert":
				return entities.DifficultyExpert
			default:
				return entities.DifficultyBeginner
			}
		}
	}
	return entities.DifficultyBeginner
}

// BatchUpdateNodes 批量更新节点
func (r *KnowledgeGraphRepositoryImpl) BatchUpdateNodes(ctx context.Context, graphID uuid.UUID, nodes []*entities.KnowledgeNode) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, node := range nodes {
			tagsJSON, _ := json.Marshal(node.Tags)
			metadataJSON, _ := json.Marshal(node.Metadata)

			query := `
				MATCH (n:KnowledgeNode {id: $id, graph_id: $graphID})
				SET n.name = $name,
					n.description = $description,
					n.node_type = $nodeType,
					n.difficulty_level = $difficultyLevel,
					n.estimated_learning_time = $estimatedLearningTime,
					n.tags = $tags,
					n.metadata = $metadata,
					n.updated_at = $updatedAt
				RETURN n
			`

			_, err := tx.Run(ctx, query, map[string]interface{}{
				"id":                     node.ID.String(),
				"graphID":                graphID.String(),
				"name":                   node.Name,
				"description":            node.Description,
				"nodeType":               string(node.Type),
				"difficultyLevel":        node.DifficultyLevel,
				"estimatedLearningTime":  node.EstimatedHours,
				"tags":                   string(tagsJSON),
				"metadata":               string(metadataJSON),
				"updatedAt":              node.UpdatedAt,
			})
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

// RemoveNode 删除单个节点
func (r *KnowledgeGraphRepositoryImpl) RemoveNode(ctx context.Context, graphID, nodeID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (n:KnowledgeNode)
			WHERE n.graph_id = $graphID AND n.id = $nodeID
			DETACH DELETE n
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
			"nodeID":  nodeID.String(),
		})

		return nil, err
	})

	return err
}

// BatchRemoveNodes 批量删除节点
func (r *KnowledgeGraphRepositoryImpl) BatchRemoveNodes(ctx context.Context, graphID uuid.UUID, nodeIDs []uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		nodeIDStrings := make([]string, len(nodeIDs))
		for i, id := range nodeIDs {
			nodeIDStrings[i] = id.String()
		}

		query := `
			MATCH (n:KnowledgeNode)
			WHERE n.graph_id = $graphID AND n.id IN $nodeIDs
			DETACH DELETE n
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
			"nodeIDs": nodeIDStrings,
		})

		return nil, err
	})

	return err
}

// RemoveRelation 删除知识关系
func (r *KnowledgeGraphRepositoryImpl) RemoveRelation(ctx context.Context, graphID, relationID uuid.UUID) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			MATCH (from:KnowledgeNode)-[r:RELATES_TO {id: $relationID}]->(to:KnowledgeNode)
			WHERE (g)-[:CONTAINS]->(from) AND (g)-[:CONTAINS]->(to)
			DELETE r
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":    graphID.String(),
			"relationID": relationID.String(),
		})

		return nil, err
	})

	return err
}

// AddRelation 添加知识关系
func (r *KnowledgeGraphRepositoryImpl) AddRelation(ctx context.Context, graphID uuid.UUID, relation *entities.KnowledgeRelation) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		metadataJSON, _ := json.Marshal(relation.Metadata)

		query := `
			MATCH (from:KnowledgeNode {id: $fromID})
			MATCH (to:KnowledgeNode {id: $toID})
			CREATE (from)-[r:PREREQUISITE {
				id: $id,
				relation_type: $relationType,
				weight: $weight,
				confidence: $confidence,
				description: $description,
				metadata: $metadata,
				created_at: $createdAt,
				updated_at: $updatedAt
			}]->(to)
			RETURN r
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"id":           relation.ID.String(),
			"fromID":       relation.FromNodeID.String(),
			"toID":         relation.ToNodeID.String(),
			"relationType": string(relation.Type),
			"weight":       relation.Weight,
			"confidence":   relation.Confidence,
			"description":  relation.Description,
			"metadata":     string(metadataJSON),
			"createdAt":    relation.CreatedAt,
			"updatedAt":    relation.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// UpdateRelation 更新知识关系
func (r *KnowledgeGraphRepositoryImpl) UpdateRelation(ctx context.Context, graphID uuid.UUID, relation *entities.KnowledgeRelation) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		metadataJSON, _ := json.Marshal(relation.Metadata)

		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			MATCH (from:KnowledgeNode)-[r:RELATES_TO {id: $id}]->(to:KnowledgeNode)
			WHERE (g)-[:CONTAINS]->(from) AND (g)-[:CONTAINS]->(to)
			SET r.relation_type = $relationType,
				r.weight = $weight,
				r.confidence = $confidence,
				r.description = $description,
				r.metadata = $metadata,
				r.updated_at = $updatedAt
			RETURN r
		`

		_, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":      graphID.String(),
			"id":           relation.ID.String(),
			"relationType": string(relation.Type),
			"weight":       relation.Weight,
			"confidence":   relation.Confidence,
			"description":  relation.Description,
			"metadata":     string(metadataJSON),
			"updatedAt":    relation.UpdatedAt,
		})

		return nil, err
	})

	return err
}

// BatchAddNodes 批量添加节点
func (r *KnowledgeGraphRepositoryImpl) BatchAddNodes(ctx context.Context, graphID uuid.UUID, nodes []*entities.KnowledgeNode) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, node := range nodes {
			// 序列化各种数组字段
			tagsJSON, _ := json.Marshal(node.Tags)
			metadataJSON, _ := json.Marshal(node.Metadata)
			prerequisitesJSON, _ := json.Marshal(node.Prerequisites)
			skillsJSON, _ := json.Marshal(node.Skills)
			keywordsJSON, _ := json.Marshal(node.Keywords)
			learningObjectivesJSON, _ := json.Marshal(node.LearningObjectives)
			assessmentCriteriaJSON, _ := json.Marshal(node.AssessmentCriteria)

			query := `
				CREATE (n:KnowledgeNode {
					id: $id,
					graph_id: $graphID,
					name: $name,
					description: $description,
					node_type: $nodeType,
					subject: $subject,
					difficulty_level: $difficultyLevel,
					estimated_learning_time: $estimatedHours,
					prerequisites: $prerequisites,
					skills: $skills,
					keywords: $keywords,
					tags: $tags,
					metadata: $metadata,
					learning_objectives: $learningObjectives,
					assessment_criteria: $assessmentCriteria,
					created_at: $createdAt,
					updated_at: $updatedAt
				})
				RETURN n
			`

			_, err := tx.Run(ctx, query, map[string]interface{}{
				"id":                     node.ID.String(),
				"graphID":                graphID.String(),
				"name":                   node.Name,
				"description":            node.Description,
				"nodeType":               string(node.Type),
				"subject":                node.Subject,
				"difficultyLevel":        int(node.DifficultyLevel),
				"estimatedHours":         node.EstimatedHours,
				"prerequisites":          string(prerequisitesJSON),
				"skills":                 string(skillsJSON),
				"keywords":               string(keywordsJSON),
				"tags":                   string(tagsJSON),
				"metadata":               string(metadataJSON),
				"learningObjectives":     string(learningObjectivesJSON),
				"assessmentCriteria":     string(assessmentCriteriaJSON),
				"createdAt":              node.CreatedAt,
				"updatedAt":              node.UpdatedAt,
			})
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

// BatchAddRelations 批量添加关系
func (r *KnowledgeGraphRepositoryImpl) BatchAddRelations(ctx context.Context, graphID uuid.UUID, relations []*entities.KnowledgeRelation) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, relation := range relations {
			metadataJSON, _ := json.Marshal(relation.Metadata)

			query := `
				MATCH (from:KnowledgeNode {id: $fromID})
				MATCH (to:KnowledgeNode {id: $toID})
				CREATE (from)-[r:PREREQUISITE {
					id: $id,
					relation_type: $relationType,
					weight: $weight,
					confidence: $confidence,
					description: $description,
					metadata: $metadata,
					created_at: $createdAt,
					updated_at: $updatedAt
				}]->(to)
				RETURN r
			`

			_, err := tx.Run(ctx, query, map[string]interface{}{
				"id":           relation.ID.String(),
				"fromID":       relation.FromNodeID.String(),
				"toID":         relation.ToNodeID.String(),
				"relationType": string(relation.Type),
				"weight":       relation.Weight,
				"confidence":   relation.Confidence,
				"description":  relation.Description,
				"metadata":     string(metadataJSON),
				"createdAt":    relation.CreatedAt,
				"updatedAt":    relation.UpdatedAt,
			})
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}

// AnalyzeLearningGaps 分析学习差距
func (r *KnowledgeGraphRepositoryImpl) AnalyzeLearningGaps(ctx context.Context, graphID, learnerID uuid.UUID) ([]*repositories.LearningGap, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 查找学习者在整个图中未掌握但需要掌握的前置知识
		query := `
			MATCH (node:KnowledgeNode)
			WHERE node.graph_id = $graphID
			AND NOT EXISTS {
				MATCH (learner:Learner {id: $learnerID})-[:MASTERED]->(node)
			}
			RETURN DISTINCT node
			ORDER BY node.difficulty_level
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID":   graphID.String(),
			"learnerID": learnerID.String(),
		})
		if err != nil {
			return nil, err
		}

		var gaps []*repositories.LearningGap
		for result.Next(ctx) {
			record := result.Record()
			nodeValue, ok := record.Get("node")
			if !ok {
				continue
			}

			node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
			if err != nil {
				return nil, err
			}

			gap := &repositories.LearningGap{
			SkillArea:        node.Name,
			CurrentLevel:     0,
			RequiredLevel:    int(node.DifficultyLevel),
			Gap:              int(node.DifficultyLevel),
			RecommendedNodes: []uuid.UUID{node.ID},
			EstimatedTime:    time.Duration(node.EstimatedHours) * time.Hour,
			Priority:         calculateGapPriorityString(node.DifficultyLevel),
			DependentSkills:  node.Skills,
		}

			gaps = append(gaps, gap)
		}

		return gaps, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*repositories.LearningGap), nil
}

// calculateGapPriority 计算学习差距优先级
func calculateGapPriority(difficultyLevel int) int {
	// 简单的优先级计算：难度越低，优先级越高
	return 10 - difficultyLevel
}

// calculateGapPriorityString 计算学习差距优先级（字符串类型）
func calculateGapPriorityString(difficultyLevel entities.DifficultyLevel) string {
	switch difficultyLevel {
	case entities.DifficultyBeginner:
		return "high"
	case entities.DifficultyElementary:
		return "high"
	case entities.DifficultyIntermediate:
		return "medium"
	case entities.DifficultyAdvanced:
		return "low"
	case entities.DifficultyExpert:
		return "low"
	default:
		return "medium"
	}
}

// mapGraphFromNeo4j 将Neo4j节点映射为KnowledgeGraph实体
func (r *KnowledgeGraphRepositoryImpl) mapGraphFromNeo4j(node neo4j.Node) (*entities.KnowledgeGraph, error) {
	graph := &entities.KnowledgeGraph{}

	if id, ok := node.Props["id"].(string); ok {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		graph.ID = parsedID
	}

	if name, ok := node.Props["name"].(string); ok {
		graph.Name = name
	}

	if description, ok := node.Props["description"].(string); ok {
		graph.Description = description
	}

	if domain, ok := node.Props["domain"].(string); ok {
		graph.Domain = domain
	}

	if subject, ok := node.Props["subject"].(string); ok {
		graph.Subject = subject
	}

	if version, ok := node.Props["version"].(string); ok {
		graph.Version = version
	}

	if isPublic, ok := node.Props["is_public"].(bool); ok {
		graph.IsPublic = isPublic
	}

	if createdBy, ok := node.Props["created_by"].(string); ok {
		parsedCreatedBy, err := uuid.Parse(createdBy)
		if err != nil {
			return nil, err
		}
		graph.CreatedBy = parsedCreatedBy
	}

	if createdAt, ok := node.Props["created_at"].(time.Time); ok {
		graph.CreatedAt = createdAt
	}

	if updatedAt, ok := node.Props["updated_at"].(time.Time); ok {
		graph.UpdatedAt = updatedAt
	}

	return graph, nil
}

// mapRelationshipFromNeo4j 将Neo4j关系映射为KnowledgeRelation实体
func (r *KnowledgeGraphRepositoryImpl) mapRelationshipFromNeo4j(rel neo4j.Relationship) (*entities.KnowledgeRelation, error) {
	relationship := &entities.KnowledgeRelation{}

	if id, ok := rel.Props["id"].(string); ok {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		relationship.ID = parsedID
	}

	// 从关系的起始和结束节点ID获取FromNodeID和ToNodeID
	fromNodeID, err := uuid.Parse(fmt.Sprintf("%v", rel.StartId))
	if err != nil {
		return nil, err
	}
	relationship.FromNodeID = fromNodeID

	toNodeID, err := uuid.Parse(fmt.Sprintf("%v", rel.EndId))
	if err != nil {
		return nil, err
	}
	relationship.ToNodeID = toNodeID

	if relationType, ok := rel.Props["relation_type"].(string); ok {
		relationship.Type = entities.RelationType(relationType)
	}

	if weight, ok := rel.Props["strength"].(float64); ok {
		relationship.Weight = weight
	}

	if confidence, ok := rel.Props["confidence"].(float64); ok {
		relationship.Confidence = confidence
	}

	if description, ok := rel.Props["description"].(string); ok {
		relationship.Description = description
	}

	if metadata, ok := rel.Props["metadata"].(string); ok {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err == nil {
			relationship.Metadata = metadataMap
		}
	}

	if createdAt, ok := rel.Props["created_at"].(time.Time); ok {
		relationship.CreatedAt = createdAt
	}

	if updatedAt, ok := rel.Props["updated_at"].(time.Time); ok {
		relationship.UpdatedAt = updatedAt
	}

	return relationship, nil
}

// GetGraph 根据ID获取知识图谱
func (r *KnowledgeGraphRepositoryImpl) GetGraph(ctx context.Context, id uuid.UUID) (*entities.KnowledgeGraph, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})
			RETURN g
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": id.String(),
		})
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, nil
		}

		record := result.Record()
		graphValue, ok := record.Get("g")
		if !ok {
			return nil, fmt.Errorf("graph not found in result")
		}

		graphNode := graphValue.(neo4j.Node)
		graph, err := r.mapGraphFromNeo4j(graphNode)
		if err != nil {
			return nil, err
		}

		// 获取图谱的所有节点
		nodesQuery := `
			MATCH (n:KnowledgeNode)
			WHERE n.graph_id = $graphID
			RETURN n
			ORDER BY n.name
		`

		nodesResult, err := tx.Run(ctx, nodesQuery, map[string]interface{}{
			"graphID": graph.ID.String(),
		})
		if err != nil {
			return nil, err
		}

		var nodes []*entities.KnowledgeNode
		for nodesResult.Next(ctx) {
			nodeRecord := nodesResult.Record()
			if nodeValue, ok := nodeRecord.Get("n"); ok {
				node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, node)
			}
		}

		if err = nodesResult.Err(); err != nil {
			return nil, err
		}

		// 转换指针数组为值数组
		nodeValues := make([]entities.KnowledgeNode, len(nodes))
		for i, node := range nodes {
			nodeValues[i] = *node
		}
		graph.Nodes = nodeValues

		// 获取图谱的所有关系
		relationshipsQuery := `
			MATCH (source:KnowledgeNode)-[r:PREREQUISITE]->(target:KnowledgeNode)
			WHERE source.graph_id = $graphID AND target.graph_id = $graphID
			RETURN r, source.id as from_id, target.id as to_id
		`

		relationshipsResult, err := tx.Run(ctx, relationshipsQuery, map[string]interface{}{
			"graphID": graph.ID.String(),
		})
		if err != nil {
			return nil, err
		}

		var relationships []*entities.KnowledgeRelation
		for relationshipsResult.Next(ctx) {
			relRecord := relationshipsResult.Record()
			if relValue, ok := relRecord.Get("r"); ok {
				relationship, err := r.mapRelationshipFromNeo4j(relValue.(neo4j.Relationship))
				if err != nil {
					return nil, err
				}

				// 设置正确的FromNodeID和ToNodeID
				if fromIDValue, ok := relRecord.Get("from_id"); ok {
					if fromIDStr, ok := fromIDValue.(string); ok {
						if fromID, err := uuid.Parse(fromIDStr); err == nil {
							relationship.FromNodeID = fromID
						}
					}
				}

				if toIDValue, ok := relRecord.Get("to_id"); ok {
					if toIDStr, ok := toIDValue.(string); ok {
						if toID, err := uuid.Parse(toIDStr); err == nil {
							relationship.ToNodeID = toID
						}
					}
				}

				relationships = append(relationships, relationship)
			}
		}

		if err = relationshipsResult.Err(); err != nil {
			return nil, err
		}

		// 转换指针数组为值数组
		relationValues := make([]entities.KnowledgeRelation, len(relationships))
		for i, relation := range relationships {
			relationValues[i] = *relation
		}
		graph.Relations = relationValues

		return graph, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get graph by ID: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	return result.(*entities.KnowledgeGraph), nil
}

// GetGraphByDomain 根据领域获取知识图谱
func (r *KnowledgeGraphRepositoryImpl) GetGraphByDomain(ctx context.Context, domain string) (*entities.KnowledgeGraph, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 查询指定领域的知识图谱
		query := `
			MATCH (g:KnowledgeGraph {domain: $domain})
			RETURN g
			LIMIT 1
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"domain": domain,
		})
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, nil // 未找到指定领域的知识图谱
		}

		record := result.Record()
		graphValue, ok := record.Get("g")
		if !ok {
			return nil, fmt.Errorf("graph not found in result")
		}

		graphNode := graphValue.(neo4j.Node)
		graph, err := r.mapGraphFromNeo4j(graphNode)
		if err != nil {
			return nil, err
		}

		// 获取图谱的所有节点
		nodesQuery := `
			MATCH (n:KnowledgeNode)
			WHERE n.graph_id = $graphID
			RETURN n
			ORDER BY n.name
		`

		nodesResult, err := tx.Run(ctx, nodesQuery, map[string]interface{}{
			"graphID": graph.ID.String(),
		})
		if err != nil {
			return nil, err
		}

		var nodes []*entities.KnowledgeNode
		for nodesResult.Next(ctx) {
			nodeRecord := nodesResult.Record()
			if nodeValue, ok := nodeRecord.Get("n"); ok {
				node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, node)
			}
		}

		if err = nodesResult.Err(); err != nil {
			return nil, err
		}

		// 转换指针数组为值数组
		nodeValues := make([]entities.KnowledgeNode, len(nodes))
		for i, node := range nodes {
			nodeValues[i] = *node
		}
		graph.Nodes = nodeValues

		// 获取图谱的所有关系
		relationshipsQuery := `
			MATCH (source:KnowledgeNode)-[r:PREREQUISITE]->(target:KnowledgeNode)
			WHERE source.graph_id = $graphID AND target.graph_id = $graphID
			RETURN r, source.id as from_id, target.id as to_id
		`

		relationshipsResult, err := tx.Run(ctx, relationshipsQuery, map[string]interface{}{
			"graphID": graph.ID.String(),
		})
		if err != nil {
			return nil, err
		}

		var relationships []*entities.KnowledgeRelation
		for relationshipsResult.Next(ctx) {
			relRecord := relationshipsResult.Record()
			if relValue, ok := relRecord.Get("r"); ok {
				relationship, err := r.mapRelationshipFromNeo4j(relValue.(neo4j.Relationship))
				if err != nil {
					return nil, err
				}

				// 设置正确的FromNodeID和ToNodeID
				if fromIDValue, ok := relRecord.Get("from_id"); ok {
					if fromIDStr, ok := fromIDValue.(string); ok {
						if fromID, err := uuid.Parse(fromIDStr); err == nil {
							relationship.FromNodeID = fromID
						}
					}
				}

				if toIDValue, ok := relRecord.Get("to_id"); ok {
					if toIDStr, ok := toIDValue.(string); ok {
						if toID, err := uuid.Parse(toIDStr); err == nil {
							relationship.ToNodeID = toID
						}
					}
				}

				relationships = append(relationships, relationship)
			}
		}

		if err = relationshipsResult.Err(); err != nil {
			return nil, err
		}

		// 转换指针数组为值数组
		relationValues := make([]entities.KnowledgeRelation, len(relationships))
		for i, relation := range relationships {
			relationValues[i] = *relation
		}
		graph.Relations = relationValues

		return graph, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get graph by domain: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	return result.(*entities.KnowledgeGraph), nil
}

// GetDependents 获取依赖于指定节点的所有节点
func (r *KnowledgeGraphRepositoryImpl) GetDependents(ctx context.Context, graphID, nodeID uuid.UUID, depth int) ([]*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		var query string
		params := map[string]interface{}{
			"nodeID":  nodeID.String(),
			"graphID": graphID.String(),
		}

		if depth > 0 {
			query = `
				MATCH (n:KnowledgeNode {id: $nodeID, graph_id: $graphID})<-[:PREREQUISITE*1..` + fmt.Sprintf("%d", depth) + `]-(dependent:KnowledgeNode)
				WHERE dependent.graph_id = $graphID
				RETURN DISTINCT dependent
			`
		} else {
			query = `
				MATCH (n:KnowledgeNode {id: $nodeID, graph_id: $graphID})<-[:PREREQUISITE*]-(dependent:KnowledgeNode)
				WHERE dependent.graph_id = $graphID
				RETURN DISTINCT dependent
			`
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var dependents []*entities.KnowledgeNode
		for result.Next(ctx) {
			record := result.Record()
			if nodeValue, ok := record.Get("dependent"); ok {
				node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
				if err != nil {
					return nil, err
				}
				dependents = append(dependents, node)
			}
		}

		return dependents, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get dependents: %w", err)
	}

	return result.([]*entities.KnowledgeNode), nil
}

// mapNodeFromNeo4j 将Neo4j节点映射为KnowledgeNode实体
func (r *KnowledgeGraphRepositoryImpl) mapNodeFromNeo4j(node neo4j.Node) (*entities.KnowledgeNode, error) {
	knowledgeNode := &entities.KnowledgeNode{}

	if id, ok := node.Props["id"].(string); ok {
		parsedID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		knowledgeNode.ID = parsedID
	}

	if name, ok := node.Props["name"].(string); ok {
		knowledgeNode.Name = name
	}

	if description, ok := node.Props["description"].(string); ok {
		knowledgeNode.Description = description
	}

	if nodeType, ok := node.Props["node_type"].(string); ok {
		knowledgeNode.Type = entities.NodeType(nodeType)
	}

	if subject, ok := node.Props["subject"].(string); ok {
		knowledgeNode.Subject = subject
	}

	if difficultyLevel, ok := node.Props["difficulty_level"].(int64); ok {
		knowledgeNode.DifficultyLevel = entities.DifficultyLevel(difficultyLevel)
	}

	if estimatedHours, ok := node.Props["estimated_learning_time"].(float64); ok {
		knowledgeNode.EstimatedHours = estimatedHours
	}

	if prerequisites, ok := node.Props["prerequisites"].(string); ok {
		var prereqList []uuid.UUID
		if err := json.Unmarshal([]byte(prerequisites), &prereqList); err == nil {
			knowledgeNode.Prerequisites = prereqList
		}
	}

	if skills, ok := node.Props["skills"].(string); ok {
		var skillsList []string
		if err := json.Unmarshal([]byte(skills), &skillsList); err == nil {
			knowledgeNode.Skills = skillsList
		}
	}

	if keywords, ok := node.Props["keywords"].(string); ok {
		var keywordsList []string
		if err := json.Unmarshal([]byte(keywords), &keywordsList); err == nil {
			knowledgeNode.Keywords = keywordsList
		}
	}

	if tags, ok := node.Props["tags"].(string); ok {
		var tagsList []string
		if err := json.Unmarshal([]byte(tags), &tagsList); err == nil {
			knowledgeNode.Tags = tagsList
		}
	}

	if metadata, ok := node.Props["metadata"].(string); ok {
		var metadataMap map[string]interface{}
		if err := json.Unmarshal([]byte(metadata), &metadataMap); err == nil {
			knowledgeNode.Metadata = metadataMap
		}
	}

	if learningObjectives, ok := node.Props["learning_objectives"].(string); ok {
		var objectivesList []string
		if err := json.Unmarshal([]byte(learningObjectives), &objectivesList); err == nil {
			knowledgeNode.LearningObjectives = objectivesList
		}
	}

	if assessmentCriteria, ok := node.Props["assessment_criteria"].(string); ok {
		var criteriaList []string
		if err := json.Unmarshal([]byte(assessmentCriteria), &criteriaList); err == nil {
			knowledgeNode.AssessmentCriteria = criteriaList
		}
	}

	if createdAt, ok := node.Props["created_at"].(time.Time); ok {
		knowledgeNode.CreatedAt = createdAt
	}

	if updatedAt, ok := node.Props["updated_at"].(time.Time); ok {
		knowledgeNode.UpdatedAt = updatedAt
	}

	return knowledgeNode, nil
}

// CreateLearningPath 创建学习路径
func (r *KnowledgeGraphRepositoryImpl) CreateLearningPath(ctx context.Context, path *entities.LearningPath) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// 创建学习路径节点
	query := `
		CREATE (lp:LearningPath {
			id: $id,
			name: $name,
			description: $description,
			subject: $subject,
			difficulty_level: $difficulty_level,
			estimated_hours: $estimated_hours,
			prerequisites: $prerequisites,
			learning_goals: $learning_goals,
			nodes: $nodes,
			milestones: $milestones,
			tags: $tags,
			is_public: $is_public,
			created_by: $created_by,
			enrollment_count: $enrollment_count,
			completion_rate: $completion_rate,
			rating: $rating,
			created_at: $created_at,
			updated_at: $updated_at
		})
		RETURN lp
	`

	// 序列化复杂字段
	prerequisitesJSON, _ := json.Marshal(path.Prerequisites)
	learningGoalsJSON, _ := json.Marshal(path.LearningGoals)
	nodesJSON, _ := json.Marshal(path.Nodes)
	milestonesJSON, _ := json.Marshal(path.Milestones)
	tagsJSON, _ := json.Marshal(path.Tags)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":               path.ID.String(),
			"name":             path.Name,
			"description":      path.Description,
			"subject":          path.Subject,
			"difficulty_level": int(path.DifficultyLevel),
			"estimated_hours":  path.EstimatedHours,
			"prerequisites":    string(prerequisitesJSON),
			"learning_goals":   string(learningGoalsJSON),
			"nodes":            string(nodesJSON),
			"milestones":       string(milestonesJSON),
			"tags":             string(tagsJSON),
			"is_public":        path.IsPublic,
			"created_by":       path.CreatedBy.String(),
			"enrollment_count": path.EnrollmentCount,
			"completion_rate":  path.CompletionRate,
			"rating":           path.Rating,
			"created_at":       path.CreatedAt,
			"updated_at":       path.UpdatedAt,
		})
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, fmt.Errorf("failed to create learning path")
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to create learning path: %w", err)
	}

	return nil
}

func (r *KnowledgeGraphRepositoryImpl) UpdateLearningPath(ctx context.Context, path *entities.LearningPath) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// 更新学习路径节点
	query := `
		MATCH (lp:LearningPath {id: $id})
		SET lp.name = $name,
			lp.description = $description,
			lp.subject = $subject,
			lp.difficulty_level = $difficulty_level,
			lp.estimated_hours = $estimated_hours,
			lp.prerequisites = $prerequisites,
			lp.learning_goals = $learning_goals,
			lp.nodes = $nodes,
			lp.milestones = $milestones,
			lp.tags = $tags,
			lp.is_public = $is_public,
			lp.enrollment_count = $enrollment_count,
			lp.completion_rate = $completion_rate,
			lp.rating = $rating,
			lp.updated_at = $updated_at
		RETURN lp
	`

	// 序列化复杂字段
	prerequisitesJSON, _ := json.Marshal(path.Prerequisites)
	learningGoalsJSON, _ := json.Marshal(path.LearningGoals)
	nodesJSON, _ := json.Marshal(path.Nodes)
	milestonesJSON, _ := json.Marshal(path.Milestones)
	tagsJSON, _ := json.Marshal(path.Tags)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":               path.ID.String(),
			"name":             path.Name,
			"description":      path.Description,
			"subject":          path.Subject,
			"difficulty_level": int(path.DifficultyLevel),
			"estimated_hours":  path.EstimatedHours,
			"prerequisites":    string(prerequisitesJSON),
			"learning_goals":   string(learningGoalsJSON),
			"nodes":            string(nodesJSON),
			"milestones":       string(milestonesJSON),
			"tags":             string(tagsJSON),
			"is_public":        path.IsPublic,
			"enrollment_count": path.EnrollmentCount,
			"completion_rate":  path.CompletionRate,
			"rating":           path.Rating,
			"updated_at":       path.UpdatedAt,
		})
		if err != nil {
			return nil, err
		}

		if !result.Next(ctx) {
			return nil, fmt.Errorf("learning path not found or failed to update")
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to update learning path: %w", err)
	}

	return nil
}

// FindAllPaths 查找两个节点之间的所有路径
func (r *KnowledgeGraphRepositoryImpl) FindAllPaths(ctx context.Context, graphID, fromNodeID, toNodeID uuid.UUID, maxDepth int) ([][]*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH path = (start:KnowledgeNode {id: $fromNodeID, graph_id: $graphID})-[:PREREQUISITE*1..%d]->(end:KnowledgeNode {id: $toNodeID, graph_id: $graphID})
			RETURN [node in nodes(path) | node] as path_nodes
			LIMIT 100
		`
		
		formattedQuery := fmt.Sprintf(query, maxDepth)
		
		queryResult, err := tx.Run(ctx, formattedQuery, map[string]interface{}{
			"graphID":    graphID.String(),
			"fromNodeID": fromNodeID.String(),
			"toNodeID":   toNodeID.String(),
		})
		if err != nil {
			return nil, err
		}

		var allPaths [][]*entities.KnowledgeNode
		for queryResult.Next(ctx) {
			record := queryResult.Record()
			pathNodesValue, ok := record.Get("path_nodes")
			if !ok {
				continue
			}

			pathNodesList := pathNodesValue.([]interface{})
			var pathNodes []*entities.KnowledgeNode
			
			for _, nodeValue := range pathNodesList {
				node, err := r.mapNodeFromNeo4j(nodeValue.(neo4j.Node))
				if err != nil {
					return nil, err
				}
				pathNodes = append(pathNodes, node)
			}
			
			allPaths = append(allPaths, pathNodes)
		}

		return allPaths, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to find paths: %w", err)
	}

	return result.([][]*entities.KnowledgeNode), nil
}

// FindShortestPath 查找最短路径
func (r *KnowledgeGraphRepositoryImpl) FindShortestPath(ctx context.Context, graphID, fromNodeID, toNodeID uuid.UUID) ([]*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (start:KnowledgeNode {id: $fromNodeID, graph_id: $graphID}),
			  (end:KnowledgeNode {id: $toNodeID, graph_id: $graphID})
		MATCH path = shortestPath((start)-[*]-(end))
		RETURN nodes(path) as nodes
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"graphID":    graphID.String(),
		"fromNodeID": fromNodeID.String(),
		"toNodeID":   toNodeID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find shortest path: %w", err)
	}

	var nodes []*entities.KnowledgeNode
	for result.Next(ctx) {
		record := result.Record()
		nodeList, ok := record.Get("nodes")
		if !ok {
			continue
		}

		nodeSlice, ok := nodeList.([]interface{})
		if !ok {
			continue
		}

		for _, nodeInterface := range nodeSlice {
			nodeData, ok := nodeInterface.(neo4j.Node)
			if !ok {
				continue
			}

			props := nodeData.Props
			node := &entities.KnowledgeNode{}

			if id, ok := props["id"].(string); ok {
				if parsedID, err := uuid.Parse(id); err == nil {
					node.ID = parsedID
				}
			}
			// GraphID不是KnowledgeNode的字段，跳过此设置
			if name, ok := props["name"].(string); ok {
				node.Name = name
			}
			if description, ok := props["description"].(string); ok {
				node.Description = description
			}
			if nodeType, ok := props["type"].(string); ok {
				node.Type = entities.NodeType(nodeType)
			}
			if subject, ok := props["subject"].(string); ok {
				node.Subject = subject
			}
			if difficulty, ok := props["difficulty_level"].(string); ok {
				switch difficulty {
				case "beginner", "1":
					node.DifficultyLevel = entities.DifficultyBeginner
				case "elementary", "2":
					node.DifficultyLevel = entities.DifficultyElementary
				case "intermediate", "3":
					node.DifficultyLevel = entities.DifficultyIntermediate
				case "advanced", "4":
					node.DifficultyLevel = entities.DifficultyAdvanced
				case "expert", "5":
					node.DifficultyLevel = entities.DifficultyExpert
				default:
					node.DifficultyLevel = entities.DifficultyBeginner
				}
			}
			if estimatedHours, ok := props["estimated_hours"].(float64); ok {
				node.EstimatedHours = estimatedHours
			}

			nodes = append(nodes, node)
		}
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("error processing shortest path result: %w", err)
	}

	return nodes, nil
}

// GetConceptMap 获取概念图
func (r *KnowledgeGraphRepositoryImpl) GetConceptMap(ctx context.Context, centerNodeID uuid.UUID, depth, maxNodes int) (*entities.ConceptMap, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// 查询以centerNodeID为中心的概念图
	query := `
		MATCH (center:KnowledgeNode {id: $centerNodeID})
		CALL apoc.path.subgraphNodes(center, {
			relationshipFilter: "PREREQUISITE|RELATED_TO|PART_OF",
			minLevel: 0,
			maxLevel: $depth,
			limit: $maxNodes
		}) YIELD node
		OPTIONAL MATCH (node)-[r:PREREQUISITE|RELATED_TO|PART_OF]-(connected)
		WHERE connected IN apoc.path.subgraphNodes(center, {
			relationshipFilter: "PREREQUISITE|RELATED_TO|PART_OF",
			minLevel: 0,
			maxLevel: $depth,
			limit: $maxNodes
		})
		RETURN DISTINCT node, collect(DISTINCT {
			id: id(r),
			type: type(r),
			source: startNode(r).id,
			target: endNode(r).id,
			properties: properties(r)
		}) as relations
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"centerNodeID": centerNodeID.String(),
		"depth":        depth,
		"maxNodes":     maxNodes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute concept map query: %w", err)
	}

	var nodes []entities.KnowledgeNode
	var relations []entities.KnowledgeRelation
	relationMap := make(map[string]bool) // 用于去重

	for result.Next(ctx) {
		record := result.Record()
		
		// 处理节点
		if nodeValue, ok := record.Get("node"); ok {
			if nodeData, ok := nodeValue.(neo4j.Node); ok {
				node, err := r.mapNodeFromNeo4j(nodeData)
				if err != nil {
					continue
				}
				nodes = append(nodes, *node)
			}
		}

		// 处理关系
		if relationsValue, ok := record.Get("relations"); ok {
			if relationsList, ok := relationsValue.([]interface{}); ok {
				for _, relItem := range relationsList {
					if relMap, ok := relItem.(map[string]interface{}); ok {
						if relType, ok := relMap["type"].(string); ok && relType != "" {
							relationKey := fmt.Sprintf("%v-%v-%s", relMap["source"], relMap["target"], relType)
							if !relationMap[relationKey] {
								relationMap[relationKey] = true
								
								relation := &entities.KnowledgeRelation{
									Type: entities.RelationType(relType),
								}
								
								// 解析源节点和目标节点ID
								if sourceID, ok := relMap["source"].(string); ok {
									if parsedID, err := uuid.Parse(sourceID); err == nil {
										relation.FromNodeID = parsedID
									}
								}
								if targetID, ok := relMap["target"].(string); ok {
									if parsedID, err := uuid.Parse(targetID); err == nil {
										relation.ToNodeID = parsedID
									}
								}
								
								relations = append(relations, *relation)
							}
						}
					}
				}
			}
		}
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("error processing concept map result: %w", err)
	}

	// 创建概念图
	conceptMap := &entities.ConceptMap{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Concept Map - Center: %s", centerNodeID.String()),
		Description: fmt.Sprintf("Concept map with center node %s, depth %d, max nodes %d", centerNodeID.String(), depth, maxNodes),
		Nodes:       nodes,
		Relations:   relations,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return conceptMap, nil
}

// GetConnectedComponents 获取知识图谱的连通分量
func (r *KnowledgeGraphRepositoryImpl) GetConnectedComponents(ctx context.Context, graphID uuid.UUID) ([][]uuid.UUID, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 使用Neo4j的图算法来查找连通分量
		query := `
			MATCH (n:KnowledgeNode)
			WHERE n.graph_id = $graphID
			WITH collect(n) as nodes
			CALL gds.wcc.stream({
				nodeQuery: 'MATCH (n:KnowledgeNode) WHERE n.graph_id = "' + $graphID + '" RETURN id(n) as id',
				relationshipQuery: 'MATCH (n:KnowledgeNode)-[r:PREREQUISITE|RELATED_TO|PART_OF]-(m:KnowledgeNode) WHERE n.graph_id = "' + $graphID + '" AND m.graph_id = "' + $graphID + '" RETURN id(n) as source, id(m) as target'
			})
			YIELD nodeId, componentId
			RETURN componentId, collect(nodeId) as nodeIds
			ORDER BY componentId
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
		})
		if err != nil {
			// 如果图算法不可用，使用简单的DFS实现
			return r.getConnectedComponentsWithDFS(ctx, tx, graphID)
		}

		var components [][]uuid.UUID
		for result.Next(ctx) {
			record := result.Record()
			nodeIds, ok := record.Get("nodeIds")
			if !ok {
				continue
			}

			var component []uuid.UUID
			for _, nodeId := range nodeIds.([]interface{}) {
				if nodeIdStr, ok := nodeId.(string); ok {
					if id, err := uuid.Parse(nodeIdStr); err == nil {
						component = append(component, id)
					}
				}
			}
			if len(component) > 0 {
				components = append(components, component)
			}
		}

		return components, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get connected components: %w", err)
	}

	return result.([][]uuid.UUID), nil
}

// getConnectedComponentsWithDFS 使用DFS算法获取连通分量（备用方案）
func (r *KnowledgeGraphRepositoryImpl) getConnectedComponentsWithDFS(ctx context.Context, tx neo4j.ManagedTransaction, graphID uuid.UUID) ([][]uuid.UUID, error) {
	// 获取所有节点
	nodesQuery := `
		MATCH (n:KnowledgeNode)
		WHERE n.graph_id = $graphID
		RETURN n.id as nodeId
	`

	nodesResult, err := tx.Run(ctx, nodesQuery, map[string]interface{}{
		"graphID": graphID.String(),
	})
	if err != nil {
		return nil, err
	}

	var allNodes []uuid.UUID
	for nodesResult.Next(ctx) {
		record := nodesResult.Record()
		if nodeIdValue, ok := record.Get("nodeId"); ok {
			if nodeIdStr, ok := nodeIdValue.(string); ok {
				if nodeId, err := uuid.Parse(nodeIdStr); err == nil {
					allNodes = append(allNodes, nodeId)
				}
			}
		}
	}

	// 获取所有关系
	relationsQuery := `
		MATCH (n:KnowledgeNode)-[r:PREREQUISITE|RELATED_TO|PART_OF]-(m:KnowledgeNode)
		WHERE n.graph_id = $graphID AND m.graph_id = $graphID
		RETURN n.id as fromId, m.id as toId
	`

	relationsResult, err := tx.Run(ctx, relationsQuery, map[string]interface{}{
		"graphID": graphID.String(),
	})
	if err != nil {
		return nil, err
	}

	// 构建邻接表
	adjacencyList := make(map[uuid.UUID][]uuid.UUID)
	for _, nodeId := range allNodes {
		adjacencyList[nodeId] = []uuid.UUID{}
	}

	for relationsResult.Next(ctx) {
		record := relationsResult.Record()
		fromIdValue, _ := record.Get("fromId")
		toIdValue, _ := record.Get("toId")

		if fromIdStr, ok := fromIdValue.(string); ok {
			if toIdStr, ok := toIdValue.(string); ok {
				if fromId, err := uuid.Parse(fromIdStr); err == nil {
					if toId, err := uuid.Parse(toIdStr); err == nil {
						adjacencyList[fromId] = append(adjacencyList[fromId], toId)
						adjacencyList[toId] = append(adjacencyList[toId], fromId)
					}
				}
			}
		}
	}

	// 使用DFS查找连通分量
	visited := make(map[uuid.UUID]bool)
	var components [][]uuid.UUID

	for _, nodeId := range allNodes {
		if !visited[nodeId] {
			component := r.dfsComponent(nodeId, adjacencyList, visited)
			components = append(components, component)
		}
	}

	return components, nil
}

// dfsComponent 使用DFS遍历一个连通分量
func (r *KnowledgeGraphRepositoryImpl) dfsComponent(nodeId uuid.UUID, adjacencyList map[uuid.UUID][]uuid.UUID, visited map[uuid.UUID]bool) []uuid.UUID {
	visited[nodeId] = true
	component := []uuid.UUID{nodeId}

	for _, neighbor := range adjacencyList[nodeId] {
		if !visited[neighbor] {
			component = append(component, r.dfsComponent(neighbor, adjacencyList, visited)...)
		}
	}

	return component
}

// GetConceptMapsByTopic 根据主题获取概念图列表
func (r *KnowledgeGraphRepositoryImpl) GetConceptMapsByTopic(ctx context.Context, graphID uuid.UUID, topic string) ([]*entities.ConceptMap, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// 查询指定主题的概念图
	query := `
		MATCH (g:KnowledgeGraph {id: $graphId})-[:CONTAINS]->(cm:ConceptMap)
		WHERE cm.subject = $topic OR cm.name CONTAINS $topic
		RETURN cm.id as id, cm.name as name, cm.description as description,
		       cm.subject as subject, cm.created_at as created_at, cm.updated_at as updated_at
		ORDER BY cm.created_at DESC
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"graphId": graphID.String(),
		"topic":   topic,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query concept maps by topic: %w", err)
	}

	var conceptMaps []*entities.ConceptMap
	for result.Next(ctx) {
		record := result.Record()
		
		id, _ := uuid.Parse(record.Values[0].(string))
		conceptMap := &entities.ConceptMap{
			ID:          id,
			Name:        record.Values[1].(string),
			Description: record.Values[2].(string),
			Subject:     record.Values[3].(string),
		}
		
		// 处理时间字段
		if createdAt, ok := record.Values[4].(time.Time); ok {
			conceptMap.CreatedAt = createdAt
		}
		if updatedAt, ok := record.Values[5].(time.Time); ok {
			conceptMap.UpdatedAt = updatedAt
		}
		
		conceptMaps = append(conceptMaps, conceptMap)
	}

	return conceptMaps, nil
}

// GetConceptMapByID 根据ID获取概念图
func (r *KnowledgeGraphRepositoryImpl) GetConceptMapByID(ctx context.Context, id uuid.UUID) (*entities.ConceptMap, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// 查询概念图基本信息
	query := `
		MATCH (cm:ConceptMap {id: $id})
		RETURN cm.id as id, cm.name as name, cm.description as description,
		       cm.created_at as created_at, cm.updated_at as updated_at
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"id": id.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query concept map: %w", err)
	}

	var conceptMap *entities.ConceptMap
	if result.Next(ctx) {
		record := result.Record()
		
		conceptMap = &entities.ConceptMap{
			ID:          id,
			Name:        record.Values[1].(string),
			Description: record.Values[2].(string),
		}
		
		// 处理时间字段
		if createdAt, ok := record.Values[3].(time.Time); ok {
			conceptMap.CreatedAt = createdAt
		}
		if updatedAt, ok := record.Values[4].(time.Time); ok {
			conceptMap.UpdatedAt = updatedAt
		}
	} else {
		return nil, fmt.Errorf("concept map not found with id: %s", id.String())
	}

	// 查询关联的节点和关系
	nodesQuery := `
		MATCH (cm:ConceptMap {id: $id})-[:CONTAINS]->(n:KnowledgeNode)
		RETURN n.id as id, n.title as title, n.description as description,
		       n.node_type as node_type, n.difficulty_level as difficulty_level
	`

	nodesResult, err := session.Run(ctx, nodesQuery, map[string]interface{}{
		"id": id.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query concept map nodes: %w", err)
	}

	var nodes []*entities.KnowledgeNode
	for nodesResult.Next(ctx) {
		record := nodesResult.Record()
		
		nodeID, _ := uuid.Parse(record.Values[0].(string))
		node := &entities.KnowledgeNode{
			ID:          nodeID,
			Name:        record.Values[1].(string),
			Description: record.Values[2].(string),
			Type:        entities.NodeType(record.Values[3].(string)),
		}
		
		if diffLevel, ok := record.Values[4].(int64); ok {
			node.DifficultyLevel = entities.DifficultyLevel(diffLevel)
		}
		
		nodes = append(nodes, node)
	}

	// 查询关系
	relationsQuery := `
		MATCH (cm:ConceptMap {id: $id})-[:CONTAINS]->(n1:KnowledgeNode)
		MATCH (n1)-[r:RELATES_TO]->(n2:KnowledgeNode)
		WHERE (cm)-[:CONTAINS]->(n2)
		RETURN n1.id as from_id, n2.id as to_id, r.type as rel_type,
		       r.weight as weight, r.confidence as confidence
	`

	relationsResult, err := session.Run(ctx, relationsQuery, map[string]interface{}{
		"id": id.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query concept map relations: %w", err)
	}

	var relations []*entities.KnowledgeRelation
	for relationsResult.Next(ctx) {
		record := relationsResult.Record()
		
		fromID, _ := uuid.Parse(record.Values[0].(string))
		toID, _ := uuid.Parse(record.Values[1].(string))
		
		relation := &entities.KnowledgeRelation{
			FromNodeID: fromID,
			ToNodeID:   toID,
			Type:       entities.RelationType(record.Values[2].(string)),
		}
		
		if weight, ok := record.Values[3].(float64); ok {
			relation.Weight = weight
		}
		if confidence, ok := record.Values[4].(float64); ok {
			relation.Confidence = confidence
		}
		
		relations = append(relations, relation)
	}

	// 转换切片类型
	conceptMapNodes := make([]entities.KnowledgeNode, len(nodes))
	for i, node := range nodes {
		conceptMapNodes[i] = *node
	}
	
	conceptMapRelations := make([]entities.KnowledgeRelation, len(relations))
	for i, relation := range relations {
		conceptMapRelations[i] = *relation
	}
	
	conceptMap.Nodes = conceptMapNodes
	conceptMap.Relations = conceptMapRelations

	return conceptMap, nil
}

// GetGraphComplexity 获取图的复杂度分析
func (r *KnowledgeGraphRepositoryImpl) GetGraphComplexity(ctx context.Context, graphID uuid.UUID) (*repositories.GraphComplexity, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// 统计节点数量
	nodeCountResult, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, 
			"MATCH (n:KnowledgeNode {graph_id: $graphId}) RETURN count(n) as count",
			map[string]interface{}{"graphId": graphID.String()})
		if err != nil {
			return 0, err
		}
		
		if result.Next(ctx) {
			return result.Record().Values[0].(int64), nil
		}
		return 0, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count nodes: %w", err)
	}
	nodeCount := int(nodeCountResult.(int64))

	// 统计关系数量
	relationCountResult, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, 
			"MATCH ()-[r:RELATION {graph_id: $graphId}]->() RETURN count(r) as count",
			map[string]interface{}{"graphId": graphID.String()})
		if err != nil {
			return 0, err
		}
		
		if result.Next(ctx) {
			return result.Record().Values[0].(int64), nil
		}
		return 0, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count relations: %w", err)
	}
	relationCount := int(relationCountResult.(int64))

	// 计算平均连接度
	averageConnectivity := 0.0
	if nodeCount > 0 {
		averageConnectivity = float64(relationCount*2) / float64(nodeCount)
	}

	// 构建复杂度结果（简化实现）
	complexity := &repositories.GraphComplexity{
		NodeCount:               nodeCount,
		RelationCount:           relationCount,
		AverageConnectivity:     averageConnectivity,
		MaxDepth:                calculateMaxDepth(ctx, graphID),
		CyclomaticComplexity:    int(calculateCyclomaticComplexity(relationCount, nodeCount)),
		ClusteringCoefficient:   calculateClusteringCoefficient(ctx, graphID),
		NodeTypeDistribution:    make(map[entities.NodeType]int),
		RelationTypeDistribution: make(map[entities.RelationType]int),
		DifficultyDistribution:  make(map[entities.DifficultyLevel]int),
		ConnectedComponents:     calculateConnectedComponents(ctx, graphID),
		LongestPath:            calculateLongestPath(ctx, graphID),
		AveragePathLength:      calculateAveragePathLength(ctx, graphID),
	}
	
	return complexity, nil
}

// GetLearningPathEffectiveness 获取学习路径有效性
func (r *KnowledgeGraphRepositoryImpl) GetLearningPathEffectiveness(ctx context.Context, pathID uuid.UUID) (*repositories.PathEffectiveness, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 查询学习路径的统计数据
		query := `
			MATCH (lp:LearningPath {id: $pathId})
			OPTIONAL MATCH (lp)<-[:ENROLLED_IN]-(learner:Learner)
			OPTIONAL MATCH (learner)-[:COMPLETED]->(lp)
			OPTIONAL MATCH (learner)-[:RATED]->(rating:Rating)-[:FOR]->(lp)
			RETURN 
				lp.id as path_id,
				count(DISTINCT learner) as learner_count,
				count(DISTINCT CASE WHEN exists((learner)-[:COMPLETED]->(lp)) THEN learner END) as completed_count,
				avg(rating.score) as avg_rating,
				avg(rating.satisfaction) as avg_satisfaction,
				lp.estimated_hours as estimated_hours
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"pathId": pathID.String(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			
			learnerCount, _ := record.Get("learner_count")
			completedCount, _ := record.Get("completed_count")
			avgRating, _ := record.Get("avg_rating")
			avgSatisfaction, _ := record.Get("avg_satisfaction")
			estimatedHours, _ := record.Get("estimated_hours")

			// 计算完成率
			var completionRate float64
			if learnerCountInt, ok := learnerCount.(int64); ok && learnerCountInt > 0 {
				if completedCountInt, ok := completedCount.(int64); ok {
					completionRate = float64(completedCountInt) / float64(learnerCountInt)
				}
			}

			// 计算成功率（假设完成即为成功）
			successRate := completionRate

			// 计算平均完成时间（简化实现）
			var avgCompletionTime time.Duration
			if estimatedHoursFloat, ok := estimatedHours.(float64); ok {
				avgCompletionTime = time.Duration(estimatedHoursFloat) * time.Hour
			}

			// 计算学习者满意度
			var learnerSatisfaction float64
			if avgSatisfactionFloat, ok := avgSatisfaction.(float64); ok {
				learnerSatisfaction = avgSatisfactionFloat
			}

			// 计算技能提升（简化实现，基于完成率）
			skillImprovement := completionRate * 0.8

			// 计算保留率（简化实现）
			retentionRate := completionRate * 0.9

			// 计算难度进展（简化实现）
			difficultyProgression := 0.75

			// 计算前置条件对齐度（简化实现）
			prerequisiteAlignment := 0.85

			// 计算推荐分数
			var recommendationScore float64
			if avgRatingFloat, ok := avgRating.(float64); ok {
				recommendationScore = avgRatingFloat / 5.0 // 假设评分是1-5分
			}

			effectiveness := &repositories.PathEffectiveness{
				PathID:                pathID,
				CompletionRate:        completionRate,
				AverageCompletionTime: avgCompletionTime,
				LearnerSatisfaction:   learnerSatisfaction,
				SkillImprovement:      skillImprovement,
				RetentionRate:         retentionRate,
				DifficultyProgression: difficultyProgression,
				PrerequisiteAlignment: prerequisiteAlignment,
				LearnerCount:          int(learnerCount.(int64)),
				SuccessRate:           successRate,
				RecommendationScore:   recommendationScore,
			}

			return effectiveness, nil
		}

		return nil, fmt.Errorf("learning path not found")
	})

	if err != nil {
		return nil, err
	}

	return result.(*repositories.PathEffectiveness), nil
}

// GetGraphStatistics 获取图统计信息
func (r *KnowledgeGraphRepositoryImpl) GetGraphStatistics(ctx context.Context, graphID uuid.UUID) (*entities.GraphStatistics, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// 统计节点数量
	nodeCountResult, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, 
			"MATCH (n:KnowledgeNode {graph_id: $graphId}) RETURN count(n) as count",
			map[string]interface{}{"graphId": graphID.String()})
		if err != nil {
			return 0, err
		}
		
		if result.Next(ctx) {
			return result.Record().Values[0].(int64), nil
		}
		return 0, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count nodes: %w", err)
	}
	nodeCount := int(nodeCountResult.(int64))

	// 统计关系数量
	relationCountResult, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, 
			"MATCH ()-[r:RELATION {graph_id: $graphId}]->() RETURN count(r) as count",
			map[string]interface{}{"graphId": graphID.String()})
		if err != nil {
			return 0, err
		}
		
		if result.Next(ctx) {
			return result.Record().Values[0].(int64), nil
		}
		return 0, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count relations: %w", err)
	}
	relationCount := int(relationCountResult.(int64))

	// 统计路径数量（简化实现）
	pathCount := 0

	// 计算平均度数
	avgDegree := 0.0
	if nodeCount > 0 {
		avgDegree = float64(relationCount*2) / float64(nodeCount)
	}

	// 计算密度
	density := 0.0
	if nodeCount > 1 {
		maxEdges := nodeCount * (nodeCount - 1) / 2
		density = float64(relationCount) / float64(maxEdges)
	}

	// 构建统计结果
	stats := &entities.GraphStatistics{
		NodeCount:       nodeCount,
		RelationCount:   relationCount,
		PathCount:       pathCount,
		NodesByType:     make(map[entities.NodeType]int),
		RelationsByType: make(map[entities.RelationType]int),
		AvgDegree:       avgDegree,
		Density:         density,
		LastUpdated:     time.Now(),
	}

	return stats, nil
}

// UpdateGraphStatistics 更新图谱统计信息
func (r *KnowledgeGraphRepositoryImpl) UpdateGraphStatistics(ctx context.Context, graphID uuid.UUID) error {
	// 简化实现：实际应该更新图谱的统计信息
	return nil
}

// 辅助函数：计算最大深度
func calculateMaxDepth(ctx context.Context, graphID uuid.UUID) int {
	// 简化实现：返回固定值，实际应该进行图遍历
	return 10
}

// 辅助函数：计算圈复杂度
func calculateCyclomaticComplexity(edges, nodes int) float64 {
	if nodes <= 1 {
		return 0
	}
	// 圈复杂度 = E - N + 2P (E=边数, N=节点数, P=连通分量数)
	// 简化为 E - N + 2 (假设为连通图)
	return float64(edges - nodes + 2)
}

// 辅助函数：计算聚类系数
func calculateClusteringCoefficient(ctx context.Context, graphID uuid.UUID) float64 {
	// 简化实现：返回固定值，实际应该计算三角形数量
	return 0.3
}

// 辅助函数：计算连通分量数
func calculateConnectedComponents(ctx context.Context, graphID uuid.UUID) int {
	// 简化实现：返回固定值，实际应该进行图遍历
	return 1
}

// 辅助函数：计算最长路径
func calculateLongestPath(ctx context.Context, graphID uuid.UUID) int {
	// 简化实现：返回固定值，实际应该进行图遍历
	return 15
}

// 辅助函数：计算平均路径长度
func calculateAveragePathLength(ctx context.Context, graphID uuid.UUID) float64 {
	// 简化实现：返回固定值，实际应该计算所有节点对之间的最短路径
	return 4.5
}

// GetGraphVersions 获取图谱版本列表
func (r *KnowledgeGraphRepositoryImpl) GetGraphVersions(ctx context.Context, graphID uuid.UUID) ([]*repositories.GraphVersion, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (gv:GraphVersion {graph_id: $graphID})
			RETURN gv
			ORDER BY gv.created_at DESC
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
		})
		if err != nil {
			return nil, err
		}

		var versions []*repositories.GraphVersion
		for result.Next(ctx) {
			record := result.Record()
			versionValue, ok := record.Get("gv")
			if !ok {
				continue
			}

			versionNode := versionValue.(neo4j.Node)
			version, err := r.mapGraphVersionFromNeo4j(versionNode)
			if err != nil {
				return nil, err
			}
			versions = append(versions, version)
		}

		return versions, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get graph versions: %w", err)
	}

	return result.([]*repositories.GraphVersion), nil
}

// mapGraphVersionFromNeo4j 将Neo4j节点映射为GraphVersion
func (r *KnowledgeGraphRepositoryImpl) mapGraphVersionFromNeo4j(node neo4j.Node) (*repositories.GraphVersion, error) {
	props := node.Props
	
	id, err := uuid.Parse(props["id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid version ID: %w", err)
	}

	graphID, err := uuid.Parse(props["graph_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid graph ID: %w", err)
	}

	createdBy, err := uuid.Parse(props["created_by"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid created_by ID: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, props["created_at"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	return &repositories.GraphVersion{
		ID:          id,
		GraphID:     graphID,
		Version:     props["version"].(string),
		Description: props["description"].(string),
		CreatedAt:   createdAt,
		CreatedBy:   createdBy,
		Changes:     []repositories.Change{}, // 简化实现
		Snapshot:    []byte{},                // 简化实现
	}, nil
}

// mapLearningPathFromNeo4j 将Neo4j节点映射为LearningPath
func (r *KnowledgeGraphRepositoryImpl) mapLearningPathFromNeo4j(node neo4j.Node) (*entities.LearningPath, error) {
	props := node.Props
	
	id, err := uuid.Parse(props["id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid learning path ID: %w", err)
	}

	var createdBy uuid.UUID
	if createdByStr, ok := props["created_by"].(string); ok && createdByStr != "" {
		createdByUUID, err := uuid.Parse(createdByStr)
		if err == nil {
			createdBy = createdByUUID
		}
	}

	createdAt, _ := time.Parse(time.RFC3339, props["created_at"].(string))
	updatedAt, _ := time.Parse(time.RFC3339, props["updated_at"].(string))

	// 解析复杂字段
	var prerequisites []uuid.UUID
	if prereqData, ok := props["prerequisites"].(string); ok {
		json.Unmarshal([]byte(prereqData), &prerequisites)
	}

	var learningGoals []string
	if goalsData, ok := props["learning_goals"].(string); ok {
		json.Unmarshal([]byte(goalsData), &learningGoals)
	}

	var nodes []entities.PathNode
	if nodesData, ok := props["nodes"].(string); ok {
		json.Unmarshal([]byte(nodesData), &nodes)
	}

	var milestones []entities.Milestone
	if milestonesData, ok := props["milestones"].(string); ok {
		json.Unmarshal([]byte(milestonesData), &milestones)
	}

	var tags []string
	if tagsData, ok := props["tags"].(string); ok {
		json.Unmarshal([]byte(tagsData), &tags)
	}

	return &entities.LearningPath{
		ID:               id,
		Name:             props["name"].(string),
		Description:      props["description"].(string),
		Subject:          props["subject"].(string),
		DifficultyLevel:  getDifficultyLevelFromProps(props, "difficulty_level"),
		EstimatedHours:   getFloatFromProps(props, "estimated_hours"),
		Prerequisites:    prerequisites,
		LearningGoals:    learningGoals,
		Nodes:            nodes,
		Milestones:       milestones,
		Tags:             tags,
		IsPublic:         props["is_public"].(bool),
		CreatedBy:        createdBy,
		EnrollmentCount:  getIntFromProps(props, "enrollment_count"),
		CompletionRate:   getFloatFromProps(props, "completion_rate"),
		Rating:           getFloatFromProps(props, "rating"),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}, nil
}

// GetNodeByName 根据名称获取节点
func (r *KnowledgeGraphRepositoryImpl) GetNodeByName(ctx context.Context, graphID uuid.UUID, name string) (*entities.KnowledgeNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (g:KnowledgeGraph {id: $graphID})-[:CONTAINS]->(n:KnowledgeNode {name: $name})
			RETURN n
		`

		result, err := tx.Run(ctx, query, map[string]interface{}{
			"graphID": graphID.String(),
			"name":    name,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			nodeValue, found := record.Get("n")
			if !found {
				return nil, fmt.Errorf("node not found in result")
			}

			node, ok := nodeValue.(neo4j.Node)
			if !ok {
				return nil, fmt.Errorf("invalid node type")
			}

			return r.mapNodeFromNeo4j(node)
		}

		return nil, fmt.Errorf("node with name '%s' not found", name)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get node by name: %w", err)
	}

	return result.(*entities.KnowledgeNode), nil
}