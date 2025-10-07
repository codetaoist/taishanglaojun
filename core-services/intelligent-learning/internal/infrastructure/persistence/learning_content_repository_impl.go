package persistence

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningContentRepositoryImpl 学习内容仓储实现
type LearningContentRepositoryImpl struct {
	db *sql.DB
	es *elasticsearch.Client
}

// NewLearningContentRepository 创建新的学习内容仓储
func NewLearningContentRepository(db *sql.DB, es *elasticsearch.Client) repositories.LearningContentRepository {
	return &LearningContentRepositoryImpl{
		db: db,
		es: es,
	}
}

// Create 创建学习内容
func (r *LearningContentRepositoryImpl) Create(ctx context.Context, content *entities.LearningContent) error {
	query := `
		INSERT INTO learning_contents (
			id, title, description, content_type, difficulty_level,
			estimated_duration, tags, metadata, content_data,
			knowledge_node_ids, prerequisites, learning_objectives,
			author_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	tagsJSON, _ := json.Marshal(content.Tags)
	metadataJSON, _ := json.Marshal(content.Metadata)
	contentDataJSON, _ := json.Marshal(content.Content)
	nodeIdsJSON, _ := json.Marshal(content.KnowledgeNodeIDs)
	prerequisitesJSON, _ := json.Marshal(content.Prerequisites)
	objectivesJSON, _ := json.Marshal(content.LearningObjectives)

	_, err := r.db.ExecContext(ctx, query,
		content.ID, content.Title, content.Description, content.Type,
		content.Difficulty, content.EstimatedDuration, tagsJSON,
		metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON,
		objectivesJSON, content.CreatedBy, content.Status,
		content.CreatedAt, content.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create learning content: %w", err)
	}

	// 同步到Elasticsearch
	if r.es != nil {
		go r.indexToElasticsearch(content)
	}

	return nil
}

// UpdateProgress 更新学习进度
func (r *LearningContentRepositoryImpl) UpdateProgress(ctx context.Context, progress *entities.ContentProgress) error {
	completedSectionsJSON, err := json.Marshal(progress.CompletedSections)
	if err != nil {
		return fmt.Errorf("failed to marshal completed sections: %w", err)
	}

	quizScoresJSON, err := json.Marshal(progress.QuizScores)
	if err != nil {
		return fmt.Errorf("failed to marshal quiz scores: %w", err)
	}

	interactionLogJSON, err := json.Marshal(progress.InteractionLog)
	if err != nil {
		return fmt.Errorf("failed to marshal interaction log: %w", err)
	}

	notesJSON, err := json.Marshal(progress.Notes)
	if err != nil {
		return fmt.Errorf("failed to marshal notes: %w", err)
	}

	bookmarksJSON, err := json.Marshal(progress.Bookmarks)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	query := `
		UPDATE content_progress SET
			progress = $2,
			time_spent = $3,
			last_position = $4,
			completed_sections = $5,
			quiz_scores = $6,
			interaction_log = $7,
			notes = $8,
			bookmarks = $9,
			is_completed = $10,
			completed_at = $11,
			last_accessed_at = $12
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		progress.ID,
		progress.Progress,
		progress.TimeSpent,
		progress.LastPosition,
		completedSectionsJSON,
		quizScoresJSON,
		interactionLogJSON,
		notesJSON,
		bookmarksJSON,
		progress.IsCompleted,
		progress.CompletedAt,
		progress.LastAccessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update content progress: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("content progress not found: %s", progress.ID)
	}

	return nil
}

// UpdateContentAnalytics 更新内容分析数据
func (r *LearningContentRepositoryImpl) UpdateContentAnalytics(ctx context.Context, contentID uuid.UUID, analytics *entities.ContentAnalytics) error {
	analyticsJSON, err := json.Marshal(analytics.InteractionData)
	if err != nil {
		return fmt.Errorf("failed to marshal interaction data: %w", err)
	}

	dropoffJSON, err := json.Marshal(analytics.DropoffPoints)
	if err != nil {
		return fmt.Errorf("failed to marshal dropoff points: %w", err)
	}

	feedbackJSON, err := json.Marshal(analytics.FeedbackSummary)
	if err != nil {
		return fmt.Errorf("failed to marshal feedback summary: %w", err)
	}

	query := `
		UPDATE learning_contents SET
			view_count = $2,
			completion_count = $3,
			average_rating = $4,
			average_time = $5,
			analytics_data = $6,
			dropoff_points = $7,
			feedback_summary = $8,
			analytics_updated_at = $9,
			updated_at = $10
		WHERE id = $1
	`

	completionCount := int(analytics.CompletionRate * float64(analytics.ViewCount))
	
	result, err := r.db.ExecContext(ctx, query,
		contentID,
		analytics.ViewCount,
		completionCount,
		analytics.AverageRating,
		analytics.AverageTime,
		analyticsJSON,
		dropoffJSON,
		feedbackJSON,
		analytics.LastUpdated,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update content analytics: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("content not found: %s", contentID)
	}

	return nil
}

// SearchNotes 搜索学习笔记
func (r *LearningContentRepositoryImpl) SearchNotes(ctx context.Context, learnerID uuid.UUID, query string, limit int) ([]*entities.LearningNote, error) {
	sqlQuery := `
		SELECT id, learner_id, content_id, content, position, tags, is_public, created_at, updated_at
		FROM learning_notes
		WHERE learner_id = $1 AND (content ILIKE $2 OR tags::text ILIKE $2)
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, learnerID, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []*entities.LearningNote

	for rows.Next() {
		note := &entities.LearningNote{}
		var tagsJSON []byte

		err := rows.Scan(
			&note.ID,
			&note.LearnerID,
			&note.ContentID,
			&note.Content,
			&note.Position,
			&tagsJSON,
			&note.IsPublic,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan learning note: %w", err)
		}

		// 解析tags JSON
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &note.Tags); err != nil {
				note.Tags = []string{}
			}
		} else {
			note.Tags = []string{}
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notes rows: %w", err)
	}

	return notes, nil
}

// RestoreContentVersion 恢复内容版本
func (r *LearningContentRepositoryImpl) RestoreContentVersion(ctx context.Context, contentID, versionID uuid.UUID) error {
	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 获取版本数据
	var versionData []byte
	query := `SELECT content_data FROM content_versions WHERE id = $1 AND content_id = $2`
	err = tx.QueryRowContext(ctx, query, versionID, contentID).Scan(&versionData)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("content version not found")
		}
		return fmt.Errorf("failed to get version data: %w", err)
	}

	// 解析版本数据
	var content entities.LearningContent
	if err := json.Unmarshal(versionData, &content); err != nil {
		return fmt.Errorf("failed to unmarshal version data: %w", err)
	}

	// 更新主内容
	updateQuery := `
		UPDATE learning_contents SET
			title = $2, description = $3, type = $4, difficulty = $5,
			estimated_duration = $6, tags = $7, learning_objectives = $8,
			prerequisites = $9, content = $10, metadata = $11,
			updated_at = $12
		WHERE id = $1
	`
	
	_, err = tx.ExecContext(ctx, updateQuery,
		contentID, content.Title, content.Description, content.Type,
		content.Difficulty, content.EstimatedDuration, pq.Array(content.Tags),
		pq.Array(content.LearningObjectives), pq.Array(content.Prerequisites),
		content.Content, content.Metadata, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to restore content: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ValidateContent 验证学习内容
func (r *LearningContentRepositoryImpl) ValidateContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentValidation, error) {
	// 获取内容信息
	content, err := r.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	validation := &repositories.ContentValidation{
		ContentID:     contentID,
		IsValid:       true,
		Errors:        []repositories.ValidationError{},
		Warnings:      []repositories.ValidationWarning{},
		QualityScore:  0.0,
		Recommendations: []string{},
	}

	// 基本验证
	if content.Title == "" {
		validation.Errors = append(validation.Errors, repositories.ValidationError{
			Type:       "REQUIRED_FIELD",
			Message:    "Title is required",
			EntityID:   contentID,
			EntityType: "content",
			Severity:   "critical",
		})
		validation.IsValid = false
	}

	if content.Description == "" {
		validation.Warnings = append(validation.Warnings, repositories.ValidationWarning{
			Type:       "MISSING_DESCRIPTION",
			Message:    "Description is recommended for better content discovery",
			EntityID:   contentID,
			EntityType: "content",
			Suggestion: "Add a detailed description to help learners understand the content",
		})
	}

	if content.Content == "" {
		validation.Errors = append(validation.Errors, repositories.ValidationError{
			Type:       "REQUIRED_FIELD",
			Message:    "Content body is required",
			EntityID:   contentID,
			EntityType: "content",
			Severity:   "critical",
		})
		validation.IsValid = false
	}

	if content.EstimatedDuration <= 0 {
		validation.Warnings = append(validation.Warnings, repositories.ValidationWarning{
			Type:       "MISSING_DURATION",
			Message:    "Estimated duration should be specified for better learning planning",
			EntityID:   contentID,
			EntityType: "content",
			Suggestion: "Specify estimated learning duration for better planning",
		})
	}

	if len(content.Tags) == 0 {
		validation.Warnings = append(validation.Warnings, repositories.ValidationWarning{
			Type:       "MISSING_TAGS",
			Message:    "Tags help with content discovery and organization",
			EntityID:   contentID,
			EntityType: "content",
			Suggestion: "Add relevant tags to improve content discoverability",
		})
	}

	if len(content.LearningObjectives) == 0 {
		validation.Warnings = append(validation.Warnings, repositories.ValidationWarning{
			Type:       "MISSING_OBJECTIVES",
			Message:    "Learning objectives help learners understand what they will achieve",
			EntityID:   contentID,
			EntityType: "content",
			Suggestion: "Define clear learning objectives to set expectations",
		})
	}

	// 计算质量分数
	qualityScore := r.calculateQualityScore(content, validation)
	validation.QualityScore = qualityScore

	// 生成改进建议
	validation.Recommendations = r.generateRecommendations(content, validation)

	return validation, nil
}

// calculateQualityScore 计算内容质量分数
func (r *LearningContentRepositoryImpl) calculateQualityScore(content *entities.LearningContent, validation *repositories.ContentValidation) float64 {
	score := 100.0

	// 扣分项
	score -= float64(len(validation.Errors)) * 20.0  // 每个错误扣20分
	score -= float64(len(validation.Warnings)) * 5.0 // 每个警告扣5分

	// 加分项
	if content.Title != "" && len(content.Title) > 10 {
		score += 5.0
	}
	if content.Description != "" && len(content.Description) > 50 {
		score += 5.0
	}
	if len(content.Tags) > 0 {
		score += 5.0
	}
	if len(content.LearningObjectives) > 0 {
		score += 10.0
	}
	if content.EstimatedDuration > 0 {
		score += 5.0
	}

	// 确保分数在0-100范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateRecommendations 生成改进建议
func (r *LearningContentRepositoryImpl) generateRecommendations(content *entities.LearningContent, validation *repositories.ContentValidation) []string {
	var recommendations []string

	if len(validation.Errors) > 0 {
		recommendations = append(recommendations, "Fix all validation errors before publishing")
	}

	if content.Description == "" {
		recommendations = append(recommendations, "Add a detailed description to help learners understand the content")
	}

	if len(content.Tags) == 0 {
		recommendations = append(recommendations, "Add relevant tags to improve content discoverability")
	}

	if len(content.LearningObjectives) == 0 {
		recommendations = append(recommendations, "Define clear learning objectives to set expectations")
	}

	if content.EstimatedDuration <= 0 {
		recommendations = append(recommendations, "Specify estimated learning duration for better planning")
	}

	if len(content.MediaResources) == 0 {
		recommendations = append(recommendations, "Consider adding multimedia resources to enhance engagement")
	}

	if validation.QualityScore < 70 {
		recommendations = append(recommendations, "Content quality is below recommended threshold. Consider reviewing and improving.")
	}

	return recommendations
}

// SearchByKeywords 根据关键词搜索学习内容
func (r *LearningContentRepositoryImpl) SearchByKeywords(ctx context.Context, keywords []string, offset, limit int) ([]*entities.LearningContent, error) {
	if len(keywords) == 0 {
		return []*entities.LearningContent{}, nil
	}

	// 构建搜索条件
	whereConditions := []string{"status = 'published'"}
	args := []interface{}{}
	argIndex := 1

	// 为每个关键词添加搜索条件
	keywordConditions := []string{}
	for _, keyword := range keywords {
		keywordConditions = append(keywordConditions, 
			fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d OR tags::text ILIKE $%d)", 
				argIndex, argIndex, argIndex))
		args = append(args, "%"+keyword+"%")
		argIndex++
	}
	
	if len(keywordConditions) > 0 {
		whereConditions = append(whereConditions, "("+strings.Join(keywordConditions, " OR ")+")")
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, type, difficulty, estimated_duration,
			   tags, metadata, content, knowledge_node_ids, prerequisites,
			   learning_objectives, created_by, status, created_at, updated_at
		FROM learning_contents
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(whereConditions, " AND "), argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search contents by keywords: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentJSON) > 0 {
			json.Unmarshal(contentJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over search results: %w", err)
	}

	return contents, nil
}

// RestoreContent 从备份恢复内容
func (r *LearningContentRepositoryImpl) RestoreContent(ctx context.Context, backupID uuid.UUID) error {
	// 查询备份信息
	var backupData []byte
	var contentIDs []string
	
	query := `
		SELECT backup_data, content_ids 
		FROM content_backups 
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	err := r.db.QueryRowContext(ctx, query, backupID).Scan(&backupData, pq.Array(&contentIDs))
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("backup not found with id: %s", backupID.String())
		}
		return fmt.Errorf("failed to query backup: %w", err)
	}
	
	// 解析备份数据
	var contents []*entities.LearningContent
	if err := json.Unmarshal(backupData, &contents); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}
	
	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// 恢复每个内容
	for _, content := range contents {
		// 检查内容是否已存在
		var existingID uuid.UUID
		checkQuery := `SELECT id FROM learning_contents WHERE id = $1`
		err := tx.QueryRowContext(ctx, checkQuery, content.ID).Scan(&existingID)
		
		if err == sql.ErrNoRows {
			// 内容不存在，创建新内容
			insertQuery := `
				INSERT INTO learning_contents (
					id, title, description, type, difficulty, estimated_duration, 
					tags, learning_objectives, prerequisites, content, metadata, 
					status, created_by, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			`
			
			_, err = tx.ExecContext(ctx, insertQuery,
				content.ID, content.Title, content.Description, content.Type,
				content.Difficulty, content.EstimatedDuration, pq.Array(content.Tags),
				pq.Array(content.LearningObjectives), pq.Array(content.Prerequisites),
				content.Content, content.Metadata, content.Status,
				content.CreatedBy, content.CreatedAt, content.UpdatedAt,
			)
			if err != nil {
				return fmt.Errorf("failed to insert restored content: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to check existing content: %w", err)
		} else {
			// 内容已存在，更新内容
			updateQuery := `
				UPDATE learning_contents SET
					title = $2, description = $3, type = $4, difficulty = $5,
					estimated_duration = $6, tags = $7, learning_objectives = $8,
					prerequisites = $9, content = $10, metadata = $11, status = $12,
					updated_at = $13
				WHERE id = $1
			`
			
			_, err = tx.ExecContext(ctx, updateQuery,
				content.ID, content.Title, content.Description, content.Type,
				content.Difficulty, content.EstimatedDuration, pq.Array(content.Tags),
				pq.Array(content.LearningObjectives), pq.Array(content.Prerequisites),
				content.Content, content.Metadata, content.Status, time.Now(),
			)
			if err != nil {
				return fmt.Errorf("failed to update restored content: %w", err)
			}
		}
	}
	
	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// ImportContent 导入内容
func (r *LearningContentRepositoryImpl) ImportContent(ctx context.Context, data []byte, format string) ([]*entities.LearningContent, error) {
	var contents []*entities.LearningContent
	
	switch strings.ToLower(format) {
	case "json":
		if err := json.Unmarshal(data, &contents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}
	case "csv":
		// 简单的CSV导入实现
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV data: %w", err)
		}
		
		// 跳过标题行
		if len(records) > 1 {
			for i := 1; i < len(records); i++ {
				record := records[i]
				if len(record) >= 4 { // 至少需要title, description, type, difficulty
					content := &entities.LearningContent{
						ID:          uuid.New(),
						Title:       record[0],
						Description: record[1],
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}
					
					// 解析内容类型
					switch strings.ToLower(record[2]) {
					case "video":
						content.Type = entities.ContentTypeVideo
					case "text":
						content.Type = entities.ContentTypeText
					case "audio":
						content.Type = entities.ContentTypeAudio
					case "interactive":
						content.Type = entities.ContentTypeInteractive
					case "quiz":
						content.Type = entities.ContentTypeQuiz
					default:
						content.Type = entities.ContentTypeText
					}
					
					// 解析难度
					switch strings.ToLower(record[3]) {
					case "beginner", "1":
						content.Difficulty = entities.DifficultyBeginner
					case "elementary", "2":
						content.Difficulty = entities.DifficultyElementary
					case "intermediate", "3":
						content.Difficulty = entities.DifficultyIntermediate
					case "advanced", "4":
						content.Difficulty = entities.DifficultyAdvanced
					case "expert", "5":
						content.Difficulty = entities.DifficultyExpert
					default:
						content.Difficulty = entities.DifficultyBeginner
					}
					
					contents = append(contents, content)
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
	
	// 批量保存导入的内容
	for _, content := range contents {
		if err := r.Create(ctx, content); err != nil {
			return nil, fmt.Errorf("failed to save imported content %s: %w", content.ID, err)
		}
	}
	
	return contents, nil
}

// OptimizeContent 优化内容
func (r *LearningContentRepositoryImpl) OptimizeContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentOptimization, error) {
	// 获取内容信息
	content, err := r.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	
	// 分析内容并生成优化建议
	var suggestions []repositories.OptimizationSuggestion
	optimizationScore := 0.8 // 基础分数
	
	// 检查标题长度
	if len(content.Title) < 10 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "title_length",
			Description:    "标题过短，建议增加到10个字符以上以提高搜索可见性",
			Priority:       2,
			ExpectedImpact: 0.1,
			Implementation: "扩展标题内容，添加关键词",
			Resources:      []string{"SEO指南", "标题优化最佳实践"},
		})
		optimizationScore -= 0.1
	}
	
	// 检查描述长度
	if len(content.Description) < 50 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "description_length",
			Description:    "描述过短，建议增加到50个字符以上以提供更多上下文",
			Priority:       3,
			ExpectedImpact: 0.15,
			Implementation: "扩展描述内容，添加学习目标和预期收益",
			Resources:      []string{"内容描述指南"},
		})
		optimizationScore -= 0.15
	}
	
	// 检查学习目标
	if len(content.LearningObjectives) == 0 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "learning_objectives",
			Description:    "缺少学习目标，建议添加明确的学习目标以提高学习效果",
			Priority:       1,
			ExpectedImpact: 0.2,
			Implementation: "添加3-5个具体、可衡量的学习目标",
			Resources:      []string{"学习目标设计指南", "SMART目标原则"},
		})
		optimizationScore -= 0.2
	}
	
	// 检查标签
	if len(content.Tags) < 3 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "tags",
			Description:    "标签数量不足，建议添加更多相关标签以提高内容发现性",
			Priority:       3,
			ExpectedImpact: 0.1,
			Implementation: "添加3-8个相关标签，包括主题、技能、难度等",
			Resources:      []string{"标签策略指南"},
		})
		optimizationScore -= 0.1
	}
	
	// 确定预计影响和工作量
	predictedImpact := 0.0
	for _, suggestion := range suggestions {
		predictedImpact += suggestion.ExpectedImpact
	}
	
	var estimatedEffort string
	if len(suggestions) <= 2 {
		estimatedEffort = "low"
	} else if len(suggestions) <= 4 {
		estimatedEffort = "medium"
	} else {
		estimatedEffort = "high"
	}
	
	optimization := &repositories.ContentOptimization{
		ContentID:         contentID,
		OptimizationScore: optimizationScore,
		Suggestions:       suggestions,
		PredictedImpact:   predictedImpact,
		EstimatedEffort:   estimatedEffort,
	}
	
	return optimization, nil
}

// PublishContent 发布内容
func (r *LearningContentRepositoryImpl) PublishContent(ctx context.Context, contentID uuid.UUID) error {
	// 首先验证内容是否存在
	content, err := r.GetByID(ctx, contentID)
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}
	
	// 检查内容是否已经发布
	if content.Status == entities.ContentStatusPublished {
		return fmt.Errorf("content is already published")
	}
	
	// 验证内容是否满足发布条件
	if content.Title == "" {
		return fmt.Errorf("content title is required for publishing")
	}
	if content.Description == "" {
		return fmt.Errorf("content description is required for publishing")
	}
	if len(content.LearningObjectives) == 0 {
		return fmt.Errorf("learning objectives are required for publishing")
	}
	
	// 更新内容状态为已发布
	query := `
		UPDATE learning_contents 
		SET status = $1, updated_at = $2 
		WHERE id = $3
	`
	
	_, err = r.db.ExecContext(ctx, query, entities.ContentStatusPublished, time.Now(), contentID)
	if err != nil {
		return fmt.Errorf("failed to publish content: %w", err)
	}
	
	return nil
}

// GetPersonalizedContent 获取个性化内容
func (r *LearningContentRepositoryImpl) GetPersonalizedContent(ctx context.Context, learnerID uuid.UUID, preferences *repositories.ContentPreferences, limit int) ([]*entities.LearningContent, error) {
	// 构建基础查询
	query := `
		SELECT id, title, description, type, content, difficulty, estimated_duration,
		       tags, created_by, status, created_at, updated_at
		FROM learning_contents
		WHERE status = 'published'
	`
	
	args := []interface{}{}
	argIndex := 1

	// 根据偏好添加过滤条件
	if preferences != nil {
		// 过滤内容类型
		if len(preferences.PreferredTypes) > 0 {
			placeholders := make([]string, len(preferences.PreferredTypes))
			for i, contentType := range preferences.PreferredTypes {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, string(contentType))
				argIndex++
			}
			query += fmt.Sprintf(" AND type IN (%s)", strings.Join(placeholders, ","))
		}

		// 过滤难度级别
		if preferences.PreferredDifficulty > 0 {
			query += fmt.Sprintf(" AND difficulty = $%d", argIndex)
			args = append(args, int(preferences.PreferredDifficulty))
			argIndex++
		}

		// 过滤兴趣领域（通过标签）
		if len(preferences.InterestAreas) > 0 {
			for _, area := range preferences.InterestAreas {
				query += fmt.Sprintf(" AND tags::text LIKE $%d", argIndex)
				args = append(args, "%"+area+"%")
				argIndex++
			}
		}

		// 排除不感兴趣的主题
		if len(preferences.AvoidTopics) > 0 {
			for _, topic := range preferences.AvoidTopics {
				query += fmt.Sprintf(" AND tags::text NOT LIKE $%d", argIndex)
				args = append(args, "%"+topic+"%")
				argIndex++
			}
		}
	}

	// 添加排序和限制
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query personalized content: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent
	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Content, &content.Difficulty, &content.EstimatedDuration,
			&tagsJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}

		// 解析标签
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &content.Tags); err != nil {
				// 如果解析失败，设置为空数组
				content.Tags = []string{}
			}
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return contents, nil
}

// GetPopularContent 获取热门内容
func (r *LearningContentRepositoryImpl) GetPopularContent(ctx context.Context, timeRange repositories.TimeRange, limit int) ([]*repositories.PopularContent, error) {
	query := `
		SELECT 
			lc.id, lc.title, lc.description, lc.type, lc.content, lc.difficulty, 
			lc.estimated_duration, lc.tags, lc.created_by, lc.status, lc.created_at, lc.updated_at,
			COALESCE(stats.view_count, 0) as view_count,
			COALESCE(stats.completion_count, 0) as completion_count,
			COALESCE(stats.rating, 0.0) as rating,
			COALESCE(stats.trend_score, 0.0) as trend_score,
			COALESCE(stats.growth_rate, 0.0) as growth_rate
		FROM learning_contents lc
		LEFT JOIN (
			SELECT 
				content_id,
				COUNT(CASE WHEN action_type = 'view' THEN 1 END) as view_count,
				COUNT(CASE WHEN action_type = 'complete' THEN 1 END) as completion_count,
				AVG(CASE WHEN action_type = 'rate' THEN CAST(metadata->>'rating' AS FLOAT) END) as rating,
				COUNT(*) * 1.0 / EXTRACT(EPOCH FROM (NOW() - $1)) as trend_score,
				(COUNT(*) - LAG(COUNT(*)) OVER (ORDER BY content_id)) * 1.0 / NULLIF(LAG(COUNT(*)) OVER (ORDER BY content_id), 0) as growth_rate
			FROM interaction_records 
			WHERE created_at BETWEEN $1 AND $2
			GROUP BY content_id
		) stats ON lc.id = stats.content_id
		WHERE lc.status = 'published'
		ORDER BY 
			COALESCE(stats.view_count, 0) * 0.3 + 
			COALESCE(stats.completion_count, 0) * 0.4 + 
			COALESCE(stats.rating, 0.0) * 0.2 + 
			COALESCE(stats.trend_score, 0.0) * 0.1 DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, timeRange.Start, timeRange.End, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query popular content: %w", err)
	}
	defer rows.Close()

	var popularContents []*repositories.PopularContent
	for rows.Next() {
		content := &entities.LearningContent{}
		popularContent := &repositories.PopularContent{}
		var tagsJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Content, &content.Difficulty, &content.EstimatedDuration,
			&tagsJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
			&popularContent.ViewCount, &popularContent.CompletionCount,
			&popularContent.Rating, &popularContent.TrendScore, &popularContent.GrowthRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan popular content: %w", err)
		}

		// 解析标签
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &content.Tags); err != nil {
				content.Tags = []string{}
			}
		}

		popularContent.Content = content
		popularContents = append(popularContents, popularContent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return popularContents, nil
}

// GetProgress 获取学习进度
func (r *LearningContentRepositoryImpl) GetProgress(ctx context.Context, learnerID, contentID uuid.UUID) (*entities.ContentProgress, error) {
	query := `
		SELECT id, learner_id, content_id, progress, time_spent, 
			   last_position, is_completed, completed_at, started_at, last_accessed_at
		FROM content_progress 
		WHERE learner_id = $1 AND content_id = $2
	`

	row := r.db.QueryRowContext(ctx, query, learnerID, contentID)

	progress := &entities.ContentProgress{}
	err := row.Scan(
		&progress.ID, &progress.LearnerID, &progress.ContentID,
		&progress.Progress, &progress.TimeSpent, &progress.LastPosition,
		&progress.IsCompleted, &progress.CompletedAt,
		&progress.StartedAt, &progress.LastAccessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到进度记录
		}
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	return progress, nil
}

// GetRecommendedContent 获取推荐内容
func (r *LearningContentRepositoryImpl) GetRecommendedContent(ctx context.Context, learnerID uuid.UUID, limit int) ([]*repositories.ContentRecommendation, error) {
	query := `
		SELECT 
			lc.id, lc.title, lc.description, lc.type, lc.content, lc.difficulty, 
			lc.estimated_duration, lc.tags, lc.created_by, lc.status, lc.created_at, lc.updated_at,
			COALESCE(rec.recommendation_score, 0.5) as recommendation_score,
			COALESCE(rec.difficulty_match, 0.5) as difficulty_match,
			COALESCE(rec.style_match, 0.5) as style_match,
			COALESCE(rec.estimated_engagement, 0.5) as estimated_engagement
		FROM learning_contents lc
		LEFT JOIN (
			SELECT 
				content_id,
				AVG(CASE 
					WHEN learner_preferences.difficulty = lc_inner.difficulty THEN 1.0 
					ELSE 0.5 
				END) as recommendation_score,
				AVG(CASE 
					WHEN learner_preferences.difficulty = lc_inner.difficulty THEN 1.0 
					ELSE ABS(learner_preferences.difficulty - lc_inner.difficulty) / 5.0 
				END) as difficulty_match,
				0.7 as style_match,
				0.6 as estimated_engagement
			FROM learning_contents lc_inner
			CROSS JOIN (
				SELECT COALESCE(AVG(difficulty), 3) as difficulty
				FROM content_progress cp
				JOIN learning_contents lc_pref ON cp.content_id = lc_pref.id
				WHERE cp.learner_id = $1 AND cp.is_completed = true
			) learner_preferences
			GROUP BY content_id
		) rec ON lc.id = rec.content_id
		WHERE lc.status = 'published'
		AND lc.id NOT IN (
			SELECT content_id 
			FROM content_progress 
			WHERE learner_id = $1 AND is_completed = true
		)
		ORDER BY 
			COALESCE(rec.recommendation_score, 0.5) * 0.4 + 
			COALESCE(rec.difficulty_match, 0.5) * 0.3 + 
			COALESCE(rec.style_match, 0.5) * 0.2 + 
			COALESCE(rec.estimated_engagement, 0.5) * 0.1 DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recommended content: %w", err)
	}
	defer rows.Close()

	var recommendations []*repositories.ContentRecommendation
	for rows.Next() {
		content := &entities.LearningContent{}
		recommendation := &repositories.ContentRecommendation{}
		var tagsJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Content, &content.Difficulty, &content.EstimatedDuration,
			&tagsJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
			&recommendation.RecommendationScore, &recommendation.DifficultyMatch,
			&recommendation.StyleMatch, &recommendation.EstimatedEngagement,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommended content: %w", err)
		}

		// 解析标签
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &content.Tags); err != nil {
				content.Tags = []string{}
			}
		}

		recommendation.Content = content
		recommendation.Reasoning = []string{
			"基于学习历史匹配",
			"难度适中",
			"学习风格匹配",
		}
		recommendation.PersonalizationFactors = []string{
			"历史偏好",
			"难度偏好",
			"学习进度",
		}
		recommendation.Priority = 1

		recommendations = append(recommendations, recommendation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return recommendations, nil
}

// GetLearningNotes 获取学习笔记
func (r *LearningContentRepositoryImpl) GetLearningNotes(ctx context.Context, learnerID, contentID uuid.UUID) ([]*entities.LearningNote, error) {
	query := `
		SELECT id, learner_id, content_id, content, position, tags, is_public, created_at, updated_at
		FROM learning_notes
		WHERE learner_id = $1 AND content_id = $2
		ORDER BY position, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning notes: %w", err)
	}
	defer rows.Close()

	var notes []*entities.LearningNote

	for rows.Next() {
		note := &entities.LearningNote{}
		var tagsJSON []byte

		err := rows.Scan(
			&note.ID,
			&note.LearnerID,
			&note.ContentID,
			&note.Content,
			&note.Position,
			&tagsJSON,
			&note.IsPublic,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan learning note: %w", err)
		}

		// 解析tags JSON
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &note.Tags); err != nil {
				note.Tags = []string{}
			}
		} else {
			note.Tags = []string{}
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notes rows: %w", err)
	}

	return notes, nil
}

// GetInteractionRecords 获取交互记录
func (r *LearningContentRepositoryImpl) GetInteractionRecords(ctx context.Context, learnerID, contentID uuid.UUID, limit int) ([]*entities.InteractionRecord, error) {
	query := `
		SELECT id, learner_id, content_id, type, element, position, data, timestamp
		FROM interaction_records
		WHERE learner_id = $1 AND content_id = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, learnerID, contentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query interaction records: %w", err)
	}
	defer rows.Close()
	
	var records []*entities.InteractionRecord
	for rows.Next() {
		record := &entities.InteractionRecord{}
		var dataJSON []byte
		
		err := rows.Scan(
			&record.ID, &record.LearnerID, &record.ContentID,
			&record.Type, &record.Element, &record.Position,
			&dataJSON, &record.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interaction record: %w", err)
		}
		
		// 反序列化data字段
		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &record.Data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal interaction data: %w", err)
			}
		}
		
		records = append(records, record)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating interaction records: %w", err)
	}
	
	return records, nil
}

// GetInteractionsByType 根据类型获取交互记录
func (r *LearningContentRepositoryImpl) GetInteractionsByType(ctx context.Context, learnerID uuid.UUID, actionType string, limit int) ([]*entities.InteractionRecord, error) {
	query := `
		SELECT id, learner_id, content_id, type, element, position, data, timestamp
		FROM interaction_records
		WHERE learner_id = $1 AND type = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, learnerID, actionType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query interaction records by type: %w", err)
	}
	defer rows.Close()
	
	var records []*entities.InteractionRecord
	for rows.Next() {
		record := &entities.InteractionRecord{}
		var dataJSON []byte
		
		err := rows.Scan(
			&record.ID, &record.LearnerID, &record.ContentID,
			&record.Type, &record.Element, &record.Position,
			&dataJSON, &record.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interaction record: %w", err)
		}
		
		// 反序列化data字段
		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &record.Data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal interaction data: %w", err)
			}
		}
		
		records = append(records, record)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating interaction records: %w", err)
	}
	
	return records, nil
}

// GetPrerequisites 获取前置内容
func (r *LearningContentRepositoryImpl) GetPrerequisites(ctx context.Context, contentID uuid.UUID) ([]*entities.LearningContent, error) {
	query := `
		SELECT lc.id, lc.title, lc.description, lc.content_type, lc.difficulty_level,
			   lc.estimated_duration, lc.tags, lc.metadata, lc.content_data,
			   lc.knowledge_node_ids, lc.prerequisites, lc.learning_objectives,
			   lc.author_id, lc.status, lc.created_at, lc.updated_at
		FROM learning_contents lc
		JOIN content_relationships cr ON lc.id = cr.prerequisite_id
		WHERE cr.content_id = $1 AND cr.relationship_type = 'prerequisite'
		ORDER BY lc.created_at
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query prerequisites: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// GetFollowUpContent 获取后续内容
func (r *LearningContentRepositoryImpl) GetFollowUpContent(ctx context.Context, contentID uuid.UUID) ([]*entities.LearningContent, error) {
	query := `
		SELECT lc.id, lc.title, lc.description, lc.content_type, lc.difficulty_level,
			   lc.estimated_duration, lc.tags, lc.metadata, lc.content_data,
			   lc.knowledge_node_ids, lc.prerequisites, lc.learning_objectives,
			   lc.author_id, lc.status, lc.created_at, lc.updated_at
		FROM learning_contents lc
		JOIN content_relationships cr ON lc.id = cr.content_id
		WHERE cr.prerequisite_id = $1 AND cr.relationship_type = 'prerequisite'
		ORDER BY lc.created_at
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query follow-up content: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// GetRelatedContent 获取相关内容
func (r *LearningContentRepositoryImpl) GetRelatedContent(ctx context.Context, contentID uuid.UUID, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT lc.id, lc.title, lc.description, lc.content_type, lc.difficulty_level,
			   lc.estimated_duration, lc.tags, lc.metadata, lc.content_data,
			   lc.knowledge_node_ids, lc.prerequisites, lc.learning_objectives,
			   lc.author_id, lc.status, lc.created_at, lc.updated_at
		FROM learning_contents lc
		JOIN content_relationships cr ON (lc.id = cr.content_id OR lc.id = cr.prerequisite_id)
		WHERE (cr.content_id = $1 OR cr.prerequisite_id = $1) 
		  AND lc.id != $1 
		  AND cr.relationship_type = 'related'
		ORDER BY lc.created_at
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, contentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query related content: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	return contents, nil
}



// scanContent 扫描内容行
func (r *LearningContentRepositoryImpl) scanContent(rows *sql.Rows) (*entities.LearningContent, error) {
	var content entities.LearningContent
	var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

	err := rows.Scan(
		&content.ID, &content.Title, &content.Description, &content.Type, &content.Difficulty,
		&content.EstimatedDuration, &tagsJSON, &metadataJSON, &contentDataJSON,
		&nodeIdsJSON, &prerequisitesJSON, &objectivesJSON,
		&content.CreatedBy, &content.Status, &content.CreatedAt, &content.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// 解析JSON字段
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &content.Tags)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &content.Metadata)
	}
	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &content.Content)
	}
	if len(nodeIdsJSON) > 0 {
		json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
	}
	if len(prerequisitesJSON) > 0 {
		json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
	}
	if len(objectivesJSON) > 0 {
		json.Unmarshal(objectivesJSON, &content.LearningObjectives)
	}

	return &content, nil
}

// GetFeedbackSummary 获取内容反馈摘要
func (r *LearningContentRepositoryImpl) GetFeedbackSummary(ctx context.Context, contentID uuid.UUID) (*entities.FeedbackSummary, error) {
	// 获取反馈总数和平均评分
	query := `
		SELECT 
			COUNT(*) as total_feedback,
			AVG(rating) as avg_rating,
			COUNT(CASE WHEN rating >= 4.0 THEN 1 END) as positive_count,
			COUNT(CASE WHEN rating < 3.0 THEN 1 END) as negative_count
		FROM content_feedback 
		WHERE content_id = $1
	`

	var totalFeedback, positiveCount, negativeCount int
	var avgRating sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, contentID).Scan(
		&totalFeedback, &avgRating, &positiveCount, &negativeCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback summary: %w", err)
	}

	// 计算情感分数 (-1.0 到 1.0)
	sentimentScore := 0.0
	if avgRating.Valid {
		// 将1-5的评分转换为-1到1的情感分数
		sentimentScore = (avgRating.Float64 - 3.0) / 2.0
	}

	// 获取常见问题和建议
	issuesQuery := `
		SELECT 
			comments,
			suggestions,
			reported_issues
		FROM content_feedback 
		WHERE content_id = $1 
		AND (comments IS NOT NULL AND comments != '' 
		     OR suggestions IS NOT NULL AND suggestions != ''
		     OR reported_issues IS NOT NULL AND reported_issues != '[]')
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := r.db.QueryContext(ctx, issuesQuery, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback details: %w", err)
	}
	defer rows.Close()

	var commonIssues []string
	var suggestions []string
	issueMap := make(map[string]int)
	suggestionMap := make(map[string]int)

	for rows.Next() {
		var comments, suggestionsText sql.NullString
		var reportedIssuesJSON sql.NullString

		err := rows.Scan(&comments, &suggestionsText, &reportedIssuesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback details: %w", err)
		}

		// 处理评论中的问题
		if comments.Valid && comments.String != "" {
			// 简单的关键词提取（实际应用中可能需要更复杂的NLP处理）
			if strings.Contains(strings.ToLower(comments.String), "difficult") ||
			   strings.Contains(strings.ToLower(comments.String), "hard") {
				issueMap["内容难度过高"]++
			}
			if strings.Contains(strings.ToLower(comments.String), "confusing") ||
			   strings.Contains(strings.ToLower(comments.String), "unclear") {
				issueMap["内容不够清晰"]++
			}
			if strings.Contains(strings.ToLower(comments.String), "boring") ||
			   strings.Contains(strings.ToLower(comments.String), "uninteresting") {
				issueMap["内容缺乏吸引力"]++
			}
		}

		// 处理建议
		if suggestionsText.Valid && suggestionsText.String != "" {
			if strings.Contains(strings.ToLower(suggestionsText.String), "example") {
				suggestionMap["增加更多示例"]++
			}
			if strings.Contains(strings.ToLower(suggestionsText.String), "practice") {
				suggestionMap["增加练习机会"]++
			}
			if strings.Contains(strings.ToLower(suggestionsText.String), "visual") {
				suggestionMap["增加视觉元素"]++
			}
		}

		// 处理报告的问题
		if reportedIssuesJSON.Valid && reportedIssuesJSON.String != "" && reportedIssuesJSON.String != "[]" {
			var reportedIssues []string
			if err := json.Unmarshal([]byte(reportedIssuesJSON.String), &reportedIssues); err == nil {
				for _, issue := range reportedIssues {
					issueMap[issue]++
				}
			}
		}
	}

	// 提取最常见的问题和建议（前5个）
	for issue, count := range issueMap {
		if count >= 2 { // 至少被提及2次
			commonIssues = append(commonIssues, issue)
		}
	}
	for suggestion, count := range suggestionMap {
		if count >= 2 { // 至少被提及2次
			suggestions = append(suggestions, suggestion)
		}
	}

	// 限制数量
	if len(commonIssues) > 5 {
		commonIssues = commonIssues[:5]
	}
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return &entities.FeedbackSummary{
		TotalFeedback:  totalFeedback,
		PositiveCount:  positiveCount,
		NegativeCount:  negativeCount,
		CommonIssues:   commonIssues,
		Suggestions:    suggestions,
		SentimentScore: sentimentScore,
	}, nil
}

// GetContentVersions 获取内容的所有版本
func (r *LearningContentRepositoryImpl) GetContentVersions(ctx context.Context, contentID uuid.UUID) ([]*repositories.ContentVersion, error) {
	query := `
		SELECT id, content_id, version, description, changes, created_at, created_by, snapshot
		FROM content_versions
		WHERE content_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content versions: %w", err)
	}
	defer rows.Close()

	var versions []*repositories.ContentVersion

	for rows.Next() {
		version := &repositories.ContentVersion{}
		var changesJSON []byte

		err := rows.Scan(
			&version.ID,
			&version.ContentID,
			&version.Version,
			&version.Description,
			&changesJSON,
			&version.CreatedAt,
			&version.CreatedBy,
			&version.Snapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content version: %w", err)
		}

		// 解析changes JSON
		if len(changesJSON) > 0 {
			if err := json.Unmarshal(changesJSON, &version.Changes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
			}
		}

		versions = append(versions, version)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating content versions: %w", err)
	}

	return versions, nil
}

// GetContentSequence 根据序列ID获取学习内容序列
func (r *LearningContentRepositoryImpl) GetContentSequence(ctx context.Context, sequenceID uuid.UUID) ([]*entities.LearningContent, error) {
	// 查询内容序列表，获取序列中的内容ID列表
	sequenceQuery := `
		SELECT content_ids, sequence_order
		FROM content_sequences
		WHERE id = $1 AND status = 'active'
	`

	var contentIDsJSON []byte
	var sequenceOrder []byte

	err := r.db.QueryRowContext(ctx, sequenceQuery, sequenceID).Scan(&contentIDsJSON, &sequenceOrder)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content sequence not found: %s", sequenceID)
		}
		return nil, fmt.Errorf("failed to get content sequence: %w", err)
	}

	// 解析内容ID列表
	var contentIDs []uuid.UUID
	if err := json.Unmarshal(contentIDsJSON, &contentIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content IDs: %w", err)
	}

	if len(contentIDs) == 0 {
		return []*entities.LearningContent{}, nil
	}

	// 构建查询参数
	placeholders := make([]string, len(contentIDs))
	args := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// 查询内容详情
	contentQuery := fmt.Sprintf(`
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE id IN (%s) AND status = 'published'
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, contentQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query content sequence: %w", err)
	}
	defer rows.Close()

	// 创建内容映射，用于按顺序排列
	contentMap := make(map[uuid.UUID]*entities.LearningContent)

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contentMap[content.ID] = content
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating content rows: %w", err)
	}

	// 按序列顺序排列内容
	var orderedContents []*entities.LearningContent
	for _, contentID := range contentIDs {
		if content, exists := contentMap[contentID]; exists {
			orderedContents = append(orderedContents, content)
		}
	}

	return orderedContents, nil
}

// GetContentProgressStats 获取内容进度统计信息
func (r *LearningContentRepositoryImpl) GetContentProgressStats(ctx context.Context, contentID uuid.UUID) (*repositories.ContentProgressStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_learners,
			COUNT(CASE WHEN progress >= 100 THEN 1 END) as completed_count,
			AVG(progress) as average_progress,
			AVG(time_spent) as average_time_spent,
			COUNT(CASE WHEN last_accessed >= NOW() - INTERVAL '7 days' THEN 1 END) as active_learners_week,
			COUNT(CASE WHEN last_accessed >= NOW() - INTERVAL '30 days' THEN 1 END) as active_learners_month
		FROM content_progress 
		WHERE content_id = $1
	`

	var stats repositories.ContentProgressStats
	var totalLearners, completedCount, activeLearners7Days, activeLearners30Days int
	var avgProgress, avgTimeSpent sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, contentID).Scan(
		&totalLearners,
		&completedCount,
		&avgProgress,
		&avgTimeSpent,
		&activeLearners7Days,
		&activeLearners30Days,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get content progress stats: %w", err)
	}

	stats.ContentID = contentID
	stats.TotalLearners = totalLearners
	stats.CompletedLearners = completedCount
	stats.CompletionRate = 0.0
	if totalLearners > 0 {
		stats.CompletionRate = float64(completedCount) / float64(totalLearners) * 100
	}

	if avgProgress.Valid {
		stats.AverageProgress = avgProgress.Float64
	}
	if avgTimeSpent.Valid {
		stats.AverageCompletionTime = time.Duration(avgTimeSpent.Float64) * time.Second
	}

	// 初始化其他必需字段
	stats.DropoffPoints = []repositories.DropoffPoint{}
	stats.EngagementMetrics = repositories.EngagementMetrics{
		AverageTimeSpent: stats.AverageCompletionTime,
		InteractionRate:  0.0,
		ReturnRate:       0.0,
		ShareRate:        0.0,
		BookmarkRate:     0.0,
		NotesTakenRate:   0.0,
		QuizAttemptRate:  0.0,
		DiscussionRate:   0.0,
	}
	stats.DifficultyFeedback = repositories.DifficultyFeedback{
		PerceivedDifficulty:    0.0,
		DifficultyDistribution: make(map[string]int),
		TooEasyCount:          0,
		TooHardCount:          0,
		JustRightCount:        0,
		SuggestedAdjustment:   0.0,
	}

	return &stats, nil
}

// GetContentRating 获取内容评分
func (r *LearningContentRepositoryImpl) GetContentRating(ctx context.Context, contentID uuid.UUID) (*repositories.ContentRating, error) {
	query := `
		SELECT 
			AVG(rating) as average_rating,
			COUNT(*) as rating_count
		FROM content_feedback 
		WHERE content_id = $1 AND rating > 0
	`

	var avgRating sql.NullFloat64
	var ratingCount int

	err := r.db.QueryRowContext(ctx, query, contentID).Scan(&avgRating, &ratingCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get content rating: %w", err)
	}

	rating := &repositories.ContentRating{
		ContentID:     contentID,
		AverageRating: 0.0,
		RatingCount:   ratingCount,
		RatingDistribution: make(map[int]int),
		RecentRating:  0.0,
		TrendDirection: "stable",
	}

	if avgRating.Valid {
		rating.AverageRating = avgRating.Float64
		rating.RecentRating = avgRating.Float64
	}

	// 获取评分分布
	distributionQuery := `
		SELECT 
			FLOOR(rating) as rating_level,
			COUNT(*) as count
		FROM content_feedback 
		WHERE content_id = $1 AND rating > 0
		GROUP BY FLOOR(rating)
	`

	rows, err := r.db.QueryContext(ctx, distributionQuery, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ratingLevel int
		var count int
		if err := rows.Scan(&ratingLevel, &count); err != nil {
			return nil, fmt.Errorf("failed to scan rating distribution: %w", err)
		}
		rating.RatingDistribution[ratingLevel] = count
	}

	return rating, nil
}

// GetContentReviews 获取内容评论列表
func (r *LearningContentRepositoryImpl) GetContentReviews(ctx context.Context, contentID uuid.UUID, offset, limit int) ([]*repositories.ContentReview, error) {
	query := `
		SELECT 
			cr.id,
			cr.content_id,
			cr.learner_id,
			l.name as learner_name,
			cr.rating,
			cr.title,
			cr.review,
			cr.pros,
			cr.cons,
			cr.helpful,
			cr.helpful_count,
			cr.created_at,
			cr.updated_at
		FROM content_reviews cr
		LEFT JOIN learners l ON cr.learner_id = l.id
		WHERE cr.content_id = $1
		ORDER BY cr.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, contentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get content reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*repositories.ContentReview
	for rows.Next() {
		var review repositories.ContentReview
		var prosJSON, consJSON sql.NullString
		var learnerName sql.NullString

		err := rows.Scan(
			&review.ID,
			&review.ContentID,
			&review.LearnerID,
			&learnerName,
			&review.Rating,
			&review.Title,
			&review.Review,
			&prosJSON,
			&consJSON,
			&review.Helpful,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content review: %w", err)
		}

		if learnerName.Valid {
			review.LearnerName = learnerName.String
		}

		// 解析 pros JSON 数组
		if prosJSON.Valid && prosJSON.String != "" {
			if err := json.Unmarshal([]byte(prosJSON.String), &review.Pros); err != nil {
				review.Pros = []string{} // 如果解析失败，设置为空数组
			}
		} else {
			review.Pros = []string{}
		}

		// 解析 cons JSON 数组
		if consJSON.Valid && consJSON.String != "" {
			if err := json.Unmarshal([]byte(consJSON.String), &review.Cons); err != nil {
				review.Cons = []string{} // 如果解析失败，设置为空数组
			}
		} else {
			review.Cons = []string{}
		}

		reviews = append(reviews, &review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate content reviews: %w", err)
	}

	return reviews, nil
}

// GetContentEffectiveness 获取内容有效性分析
func (r *LearningContentRepositoryImpl) GetContentEffectiveness(ctx context.Context, contentID uuid.UUID) (*repositories.ContentEffectiveness, error) {
	// 查询基础数据
	query := `
		SELECT 
			view_count,
			completion_count,
			average_rating,
			average_time
		FROM learning_contents 
		WHERE id = $1
	`

	var viewCount, completionCount, averageTime int
	var averageRating float64

	err := r.db.QueryRowContext(ctx, query, contentID).Scan(
		&viewCount,
		&completionCount,
		&averageRating,
		&averageTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content not found: %s", contentID)
		}
		return nil, fmt.Errorf("failed to get content effectiveness data: %w", err)
	}

	// 计算学习有效性指标
	learningEffectiveness := 0.0
	if viewCount > 0 {
		learningEffectiveness = float64(completionCount) / float64(viewCount)
	}

	// 计算技能提升（基于完成率和评分）
	skillImprovement := learningEffectiveness * (averageRating / 5.0)

	// 计算知识保持率（简化计算）
	knowledgeRetention := averageRating / 5.0 * 0.8

	// 计算参与度分数
	engagementScore := learningEffectiveness * 0.7 + (averageRating / 5.0) * 0.3

	// 计算完成质量
	completionQuality := averageRating / 5.0

	// 计算学习者满意度
	learnerSatisfaction := averageRating / 5.0

	// 计算推荐率（基于评分）
	recommendationRate := 0.0
	if averageRating >= 4.0 {
		recommendationRate = 0.8
	} else if averageRating >= 3.0 {
		recommendationRate = 0.5
	} else {
		recommendationRate = 0.2
	}

	// 计算总体分数
	overallScore := (learningEffectiveness*0.3 + skillImprovement*0.2 + 
		knowledgeRetention*0.15 + engagementScore*0.15 + 
		completionQuality*0.1 + learnerSatisfaction*0.1)

	return &repositories.ContentEffectiveness{
		ContentID:           contentID,
		LearningEffectiveness: learningEffectiveness,
		SkillImprovement:    skillImprovement,
		KnowledgeRetention:  knowledgeRetention,
		EngagementScore:     engagementScore,
		CompletionQuality:   completionQuality,
		LearnerSatisfaction: learnerSatisfaction,
		RecommendationRate:  recommendationRate,
		OverallScore:        overallScore,
	}, nil
}

// GetLearningOutcomes 获取学习成果
func (r *LearningContentRepositoryImpl) GetLearningOutcomes(ctx context.Context, contentID uuid.UUID) (*repositories.LearningOutcomes, error) {
	// 查询学习成果数据
	query := `
		SELECT 
			view_count,
			completion_count,
			average_rating
		FROM learning_contents 
		WHERE id = $1
	`

	var viewCount, completionCount int
	var averageRating float64

	err := r.db.QueryRowContext(ctx, query, contentID).Scan(
		&viewCount,
		&completionCount,
		&averageRating,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content not found: %s", contentID)
		}
		return nil, fmt.Errorf("failed to get learning outcomes data: %w", err)
	}

	// 模拟技能获得数据
	skillsAcquired := map[string]float64{
		"problem_solving": averageRating / 5.0 * 0.8,
		"critical_thinking": averageRating / 5.0 * 0.7,
		"communication": averageRating / 5.0 * 0.6,
	}

	// 模拟知识获得数据
	knowledgeGained := map[string]float64{
		"theoretical_knowledge": averageRating / 5.0 * 0.9,
		"practical_knowledge": averageRating / 5.0 * 0.8,
		"domain_expertise": averageRating / 5.0 * 0.7,
	}

	// 模拟能力提升数据
	competenciesImproved := map[string]float64{
		"analytical_skills": averageRating / 5.0 * 0.8,
		"technical_skills": averageRating / 5.0 * 0.9,
		"soft_skills": averageRating / 5.0 * 0.6,
	}

	// 模拟学习目标达成数据
	learningObjectivesMet := map[uuid.UUID]float64{
		uuid.New(): averageRating / 5.0 * 0.9,
		uuid.New(): averageRating / 5.0 * 0.8,
	}

	// 计算长期保持率
	longTermRetention := averageRating / 5.0 * 0.7

	// 计算应用率
	applicationRate := float64(completionCount) / float64(viewCount) * 0.8

	return &repositories.LearningOutcomes{
		ContentID:             contentID,
		SkillsAcquired:        skillsAcquired,
		KnowledgeGained:       knowledgeGained,
		CompetenciesImproved:  competenciesImproved,
		LearningObjectivesMet: learningObjectivesMet,
		AssessmentResults:     repositories.AssessmentResults{
			AverageScore:        averageRating,
			PassRate:            float64(completionCount) / float64(viewCount),
			ScoreDistribution:   make(map[string]int),
			CommonMistakes:      []string{},
			ImprovementAreas:    []string{},
			StrengthAreas:       []string{},
		},
		LongTermRetention: longTermRetention,
		ApplicationRate:   applicationRate,
	}, nil
}

// GetByType 根据内容类型获取学习内容
func (r *LearningContentRepositoryImpl) GetByType(ctx context.Context, contentType entities.ContentType, offset, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE content_type = $1 AND status = 'published'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, contentType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by type: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating content rows: %w", err)
	}

	return contents, nil
}

// CreateContentVersion 创建内容版本
func (r *LearningContentRepositoryImpl) CreateContentVersion(ctx context.Context, contentID uuid.UUID, version *repositories.ContentVersion) error {
	changesJSON, err := json.Marshal(version.Changes)
	if err != nil {
		return fmt.Errorf("failed to marshal changes: %w", err)
	}

	query := `
		INSERT INTO content_versions (id, content_id, version, description, changes, created_at, created_by, snapshot)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.ExecContext(ctx, query,
		version.ID, contentID, version.Version, version.Description,
		changesJSON, version.CreatedAt, version.CreatedBy, version.Snapshot,
	)

	if err != nil {
		return fmt.Errorf("failed to create content version: %w", err)
	}

	return nil
}

// CreateProgress 创建内容学习进度
func (r *LearningContentRepositoryImpl) CreateProgress(ctx context.Context, progress *entities.ContentProgress) error {
	completedSectionsJSON, err := json.Marshal(progress.CompletedSections)
	if err != nil {
		return fmt.Errorf("failed to marshal completed sections: %w", err)
	}

	quizScoresJSON, err := json.Marshal(progress.QuizScores)
	if err != nil {
		return fmt.Errorf("failed to marshal quiz scores: %w", err)
	}

	interactionLogJSON, err := json.Marshal(progress.InteractionLog)
	if err != nil {
		return fmt.Errorf("failed to marshal interaction log: %w", err)
	}

	notesJSON, err := json.Marshal(progress.Notes)
	if err != nil {
		return fmt.Errorf("failed to marshal notes: %w", err)
	}

	bookmarksJSON, err := json.Marshal(progress.Bookmarks)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmarks: %w", err)
	}

	query := `
		INSERT INTO content_progress (
			id, learner_id, content_id, progress, time_spent, last_position,
			completed_sections, quiz_scores, interaction_log, notes, bookmarks,
			is_completed, completed_at, started_at, last_accessed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = r.db.ExecContext(ctx, query,
		progress.ID, progress.LearnerID, progress.ContentID, progress.Progress,
		progress.TimeSpent, progress.LastPosition, completedSectionsJSON,
		quizScoresJSON, interactionLogJSON, notesJSON, bookmarksJSON,
		progress.IsCompleted, progress.CompletedAt, progress.StartedAt,
		progress.LastAccessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create content progress: %w", err)
	}

	return nil
}

// GetLearnerProgress 获取学习者的内容进度
func (r *LearningContentRepositoryImpl) GetLearnerProgress(ctx context.Context, learnerID uuid.UUID, offset, limit int) ([]*entities.ContentProgress, error) {
	query := `
		SELECT id, learner_id, content_id, progress, time_spent, last_position,
			   completed_sections, quiz_scores, is_completed, completed_at,
			   started_at, last_accessed_at, interaction_log, notes, bookmarks
		FROM content_progress
		WHERE learner_id = $1
		ORDER BY last_accessed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query learner progress: %w", err)
	}
	defer rows.Close()

	var progressList []*entities.ContentProgress
	for rows.Next() {
		var progress entities.ContentProgress
		var completedSectionsJSON, quizScoresJSON, interactionLogJSON, notesJSON, bookmarksJSON []byte

		err := rows.Scan(
			&progress.ID,
			&progress.LearnerID,
			&progress.ContentID,
			&progress.Progress,
			&progress.TimeSpent,
			&progress.LastPosition,
			&completedSectionsJSON,
			&quizScoresJSON,
			&progress.IsCompleted,
			&progress.CompletedAt,
			&progress.StartedAt,
			&progress.LastAccessedAt,
			&interactionLogJSON,
			&notesJSON,
			&bookmarksJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress row: %w", err)
		}

		// 解析JSON字段
		if len(completedSectionsJSON) > 0 {
			json.Unmarshal(completedSectionsJSON, &progress.CompletedSections)
		}
		if len(quizScoresJSON) > 0 {
			json.Unmarshal(quizScoresJSON, &progress.QuizScores)
		}
		if len(interactionLogJSON) > 0 {
			json.Unmarshal(interactionLogJSON, &progress.InteractionLog)
		}
		if len(notesJSON) > 0 {
			json.Unmarshal(notesJSON, &progress.Notes)
		}
		if len(bookmarksJSON) > 0 {
			json.Unmarshal(bookmarksJSON, &progress.Bookmarks)
		}

		progressList = append(progressList, &progress)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating progress rows: %w", err)
	}

	return progressList, nil
}

// CompareContentVersions 比较内容版本
func (r *LearningContentRepositoryImpl) CompareContentVersions(ctx context.Context, contentID, version1ID, version2ID uuid.UUID) (*repositories.ContentComparison, error) {
	// 获取两个版本的内容
	query := `
		SELECT id, title, description, type, content, difficulty, 
			   estimated_duration, tags, created_at
		FROM learning_contents 
		WHERE id = $1 
		ORDER BY created_at
		LIMIT 2
	`

	rows, err := r.db.QueryContext(ctx, query, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query content versions: %w", err)
	}
	defer rows.Close()

	var versions []*entities.LearningContent
	for rows.Next() {
		var content entities.LearningContent
		var tagsJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Content, &content.Difficulty, &content.EstimatedDuration,
			&tagsJSON, &content.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content row: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &content.Tags); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
			}
		}

		versions = append(versions, &content)
	}

	if len(versions) != 2 {
		return nil, fmt.Errorf("expected 2 versions, got %d", len(versions))
	}

	// 比较两个版本
	comparison := &repositories.ContentComparison{
		ContentID:  contentID,
		Version1ID: version1ID,
		Version2ID: version2ID,
	}

	var differences []repositories.ContentDifference

	// 比较各个字段
	if versions[0].Title != versions[1].Title {
		differences = append(differences, repositories.ContentDifference{
			Field:    "title",
			Type:     "modified",
			OldValue: versions[0].Title,
			NewValue: versions[1].Title,
			Impact:   "medium",
		})
	}

	if versions[0].Description != versions[1].Description {
		differences = append(differences, repositories.ContentDifference{
			Field:    "description",
			Type:     "modified",
			OldValue: versions[0].Description,
			NewValue: versions[1].Description,
			Impact:   "low",
		})
	}

	if versions[0].Type != versions[1].Type {
		differences = append(differences, repositories.ContentDifference{
			Field:    "type",
			Type:     "modified",
			OldValue: versions[0].Type,
			NewValue: versions[1].Type,
			Impact:   "high",
		})
	}

	if versions[0].Difficulty != versions[1].Difficulty {
		differences = append(differences, repositories.ContentDifference{
			Field:    "difficulty",
			Type:     "modified",
			OldValue: versions[0].Difficulty,
			NewValue: versions[1].Difficulty,
			Impact:   "medium",
		})
	}

	comparison.Differences = differences
	comparison.ChangeCount = len(differences)

	// 生成摘要
	if len(differences) == 0 {
		comparison.Summary = "No differences found between versions"
	} else {
		comparison.Summary = fmt.Sprintf("Found %d differences between versions", len(differences))
	}

	// 识别主要变更
	var majorChanges []string
	for _, diff := range differences {
		if diff.Impact == "high" {
			majorChanges = append(majorChanges, fmt.Sprintf("%s changed", diff.Field))
		}
	}
	comparison.MajorChanges = majorChanges

	return comparison, nil
}

// BatchUpdateStatus 批量更新学习内容状态
func (r *LearningContentRepositoryImpl) BatchUpdateStatus(ctx context.Context, contentIDs []uuid.UUID, status entities.ContentStatus) error {
	if len(contentIDs) == 0 {
		return nil
	}

	// 构建占位符
	placeholders := make([]string, len(contentIDs))
	args := make([]interface{}, len(contentIDs)+2)
	
	for i, id := range contentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	
	args[len(contentIDs)] = status
	args[len(contentIDs)+1] = time.Now()

	query := fmt.Sprintf(`
		UPDATE learning_contents 
		SET status = $%d, updated_at = $%d
		WHERE id IN (%s)
	`, len(contentIDs)+1, len(contentIDs)+2, strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch update content status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no content found with provided IDs")
	}

	return nil
}

// BatchUpdate 批量更新学习内容
func (r *LearningContentRepositoryImpl) BatchUpdate(ctx context.Context, contents []*entities.LearningContent) error {
	if len(contents) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE learning_contents SET
			title = ?, description = ?, type = ?, content = ?, 
			difficulty = ?, estimated_duration = ?, tags = ?, 
			status = ?, updated_at = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, content := range contents {
		content.UpdatedAt = time.Now()
		tagsJSON, _ := json.Marshal(content.Tags)

		_, err := stmt.ExecContext(ctx,
			content.Title, content.Description, content.Type, content.Content,
			content.Difficulty, content.EstimatedDuration, tagsJSON,
			content.Status, content.UpdatedAt, content.ID,
		)

		if err != nil {
			return fmt.Errorf("failed to update content %s: %w", content.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetContentInteractions 获取内容的交互记录
func (r *LearningContentRepositoryImpl) GetContentInteractions(ctx context.Context, contentID uuid.UUID, limit int) ([]*entities.InteractionRecord, error) {
	query := `
		SELECT id, learner_id, content_id, type, element, position, data, timestamp
		FROM interaction_records
		WHERE content_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, contentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get content interactions: %w", err)
	}
	defer rows.Close()

	var interactions []*entities.InteractionRecord

	for rows.Next() {
		interaction := &entities.InteractionRecord{}
		var dataJSON []byte

		err := rows.Scan(
			&interaction.ID,
			&interaction.LearnerID,
			&interaction.ContentID,
			&interaction.Type,
			&interaction.Element,
			&interaction.Position,
			&dataJSON,
			&interaction.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan interaction record: %w", err)
		}

		// 解析JSON数据
		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &interaction.Data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal interaction data: %w", err)
			}
		}

		interactions = append(interactions, interaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating interaction records: %w", err)
	}

	return interactions, nil
}

// GetContentAnalytics 获取内容分析数据
func (r *LearningContentRepositoryImpl) GetContentAnalytics(ctx context.Context, contentID uuid.UUID) (*entities.ContentAnalytics, error) {
	// 查询基本分析数据
	query := `
		SELECT 
			view_count,
			completion_count,
			average_rating,
			average_time,
			analytics_data,
			created_at,
			updated_at
		FROM learning_contents 
		WHERE id = $1
	`

	var viewCount, completionCount, averageTime int
	var averageRating float64
	var analyticsDataJSON sql.NullString
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, contentID).Scan(
		&viewCount,
		&completionCount,
		&averageRating,
		&averageTime,
		&analyticsDataJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content not found: %s", contentID)
		}
		return nil, fmt.Errorf("failed to get content analytics: %w", err)
	}

	// 计算完成率
	completionRate := 0.0
	if viewCount > 0 {
		completionRate = float64(completionCount) / float64(viewCount)
	}

	// 解析交互数据
	var interactionData map[string]interface{}
	if analyticsDataJSON.Valid && analyticsDataJSON.String != "" {
		if err := json.Unmarshal([]byte(analyticsDataJSON.String), &interactionData); err != nil {
			interactionData = make(map[string]interface{})
		}
	} else {
		interactionData = make(map[string]interface{})
	}

	// 获取流失点数据（简化实现）
	dropoffPoints := []int{10, 25, 50, 75, 90} // 默认流失点百分比

	// 获取反馈摘要
	feedbackSummary := entities.FeedbackSummary{
		TotalFeedback:   0,
		PositiveCount:   0,
		NegativeCount:   0,
		CommonIssues:    []string{},
		Suggestions:     []string{},
		SentimentScore:  0.0,
	}

	// 查询反馈数据
	feedbackQuery := `
		SELECT COUNT(*), AVG(rating)
		FROM content_feedback 
		WHERE content_id = $1
	`
	
	var feedbackCount int
	var avgFeedbackRating sql.NullFloat64
	
	err = r.db.QueryRowContext(ctx, feedbackQuery, contentID).Scan(&feedbackCount, &avgFeedbackRating)
	if err == nil {
		feedbackSummary.TotalFeedback = feedbackCount
		
		// 计算情感分数和正负面反馈
		if avgFeedbackRating.Valid {
			rating := avgFeedbackRating.Float64
			feedbackSummary.SentimentScore = (rating - 3.0) / 2.0 // 转换为-1到1的范围
			
			if rating >= 4.0 {
				feedbackSummary.PositiveCount = int(float64(feedbackCount) * 0.7) // 估算
			} else if rating <= 2.0 {
				feedbackSummary.NegativeCount = int(float64(feedbackCount) * 0.7) // 估算
			}
		}
	}

	analytics := &entities.ContentAnalytics{
		ViewCount:       viewCount,
		CompletionRate:  completionRate,
		AverageRating:   averageRating,
		AverageTime:     averageTime,
		DropoffPoints:   dropoffPoints,
		InteractionData: interactionData,
		FeedbackSummary: feedbackSummary,
		LastUpdated:     time.Now(),
	}

	return analytics, nil
}

// BatchDelete 批量删除学习内容
func (r *LearningContentRepositoryImpl) BatchDelete(ctx context.Context, contentIDs []uuid.UUID) error {
	if len(contentIDs) == 0 {
		return nil
	}

	// 构建查询占位符
	placeholders := make([]string, len(contentIDs))
	args := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		DELETE FROM learning_contents 
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to batch delete content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no content found to delete")
	}

	return nil
}

// ArchiveContent 归档学习内容
func (r *LearningContentRepositoryImpl) ArchiveContent(ctx context.Context, contentID uuid.UUID) error {
	query := `
		UPDATE learning_contents 
		SET status = 'archived', updated_at = $2
		WHERE id = $1
	`
	
	result, err := r.db.ExecContext(ctx, query, contentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to archive content: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("content not found")
	}
	
	return nil
}

// GetByID 根据ID获取学习内容
func (r *LearningContentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	content := &entities.LearningContent{}
	var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

	err := row.Scan(
		&content.ID, &content.Title, &content.Description, &content.Type,
		&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
		&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
		&objectivesJSON, &content.CreatedBy, &content.Status,
		&content.CreatedAt, &content.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("learning content not found")
		}
		return nil, fmt.Errorf("failed to get learning content: %w", err)
	}

	// 解析JSON字段
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &content.Tags)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &content.Metadata)
	}
	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &content.Content)
	}
	if len(nodeIdsJSON) > 0 {
		json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
	}
	if len(prerequisitesJSON) > 0 {
		json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
	}
	if len(objectivesJSON) > 0 {
		json.Unmarshal(objectivesJSON, &content.LearningObjectives)
	}

	return content, nil
}

// Update 更新学习内容
func (r *LearningContentRepositoryImpl) Update(ctx context.Context, content *entities.LearningContent) error {
	query := `
		UPDATE learning_contents SET
			title = $2, description = $3, content_type = $4,
			difficulty_level = $5, estimated_duration = $6, tags = $7,
			metadata = $8, content_data = $9, knowledge_node_ids = $10,
			prerequisites = $11, learning_objectives = $12, author_id = $13,
			status = $14, updated_at = $15
		WHERE id = $1
	`

	tagsJSON, _ := json.Marshal(content.Tags)
	metadataJSON, _ := json.Marshal(content.Metadata)
	contentDataJSON, _ := json.Marshal(content.Content)
	nodeIdsJSON, _ := json.Marshal(content.KnowledgeNodeIDs)
	prerequisitesJSON, _ := json.Marshal(content.Prerequisites)
	objectivesJSON, _ := json.Marshal(content.LearningObjectives)

	content.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		content.ID, content.Title, content.Description, content.Type,
		content.Difficulty, content.EstimatedDuration, tagsJSON,
		metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON,
		objectivesJSON, content.CreatedBy, content.Status, content.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update learning content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("learning content not found")
	}

	// 同步到Elasticsearch
	if r.es != nil {
		go r.indexToElasticsearch(content)
	}

	return nil
}

// Delete 删除学习内容
func (r *LearningContentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM learning_contents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete learning content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("learning content not found")
	}

	// 从Elasticsearch删除
	if r.es != nil {
		go r.deleteFromElasticsearch(id)
	}

	return nil
}

// List 获取学习内容列表
func (r *LearningContentRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE status = 'published'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list learning contents: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, nil
}

// GetByKnowledgeNode 根据知识节点获取学习内容
func (r *LearningContentRepositoryImpl) GetByKnowledgeNode(ctx context.Context, nodeID uuid.UUID, offset, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE knowledge_node_ids @> $1 AND status = 'published'
		ORDER BY difficulty_level, created_at
		LIMIT $2 OFFSET $3
	`

	nodeIDJSON, _ := json.Marshal([]uuid.UUID{nodeID})

	rows, err := r.db.QueryContext(ctx, query, nodeIDJSON, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by knowledge node: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, nil
}

// GetByTags 根据标签获取学习内容
func (r *LearningContentRepositoryImpl) GetByTags(ctx context.Context, tags []string, offset, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE tags ?| $1 AND status = 'published'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(tags), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by tags: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, nil
}

// GetByAuthor 根据作者ID获取学习内容
func (r *LearningContentRepositoryImpl) GetByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE author_id = $1 AND status = 'published'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by author: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, nil
}

// GetByDifficulty 根据难度级别范围获取学习内容
func (r *LearningContentRepositoryImpl) GetByDifficulty(ctx context.Context, minLevel, maxLevel entities.DifficultyLevel, offset, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE difficulty_level >= $1 AND difficulty_level <= $2 AND status = 'published'
		ORDER BY difficulty_level, created_at
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, minLevel, maxLevel, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by difficulty: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, nil
}

// GetByStatus 根据状态获取学习内容
func (r *LearningContentRepositoryImpl) GetByStatus(ctx context.Context, status entities.ContentStatus, offset, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get contents by status: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, nil
}

// Search 搜索学习内容
func (r *LearningContentRepositoryImpl) Search(ctx context.Context, query *repositories.ContentSearchQuery) ([]*entities.LearningContent, int, error) {
	// 如果有Elasticsearch，优先使用ES搜索
	if r.es != nil && len(query.Keywords) > 0 {
		return r.searchWithElasticsearch(ctx, query)
	}

	// 否则使用PostgreSQL搜索
	return r.searchWithPostgreSQL(ctx, query)
}

// searchWithPostgreSQL 使用PostgreSQL搜索
func (r *LearningContentRepositoryImpl) searchWithPostgreSQL(ctx context.Context, query *repositories.ContentSearchQuery) ([]*entities.LearningContent, int, error) {
	// 构建WHERE条件
	whereConditions := []string{"status = 'published'"}
	args := []interface{}{}
	argIndex := 1

	if len(query.Keywords) > 0 {
		keywordConditions := []string{}
		for _, keyword := range query.Keywords {
			keywordConditions = append(keywordConditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
			args = append(args, "%"+keyword+"%")
			argIndex++
		}
		whereConditions = append(whereConditions, "("+strings.Join(keywordConditions, " OR ")+")")
	}

	if query.ContentType != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("content_type = $%d", argIndex))
		args = append(args, *query.ContentType)
		argIndex++
	}

	if query.DifficultyLevel != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("difficulty_level = $%d", argIndex))
		args = append(args, *query.DifficultyLevel)
		argIndex++
	}

	if len(query.Tags) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("tags ?| $%d", argIndex))
		args = append(args, pq.Array(query.Tags))
		argIndex++
	}

	if query.AuthorID != nil && *query.AuthorID != uuid.Nil {
		whereConditions = append(whereConditions, fmt.Sprintf("author_id = $%d", argIndex))
		args = append(args, *query.AuthorID)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(whereConditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM learning_contents
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count learning contents: %w", err)
	}

	// 获取数据
	dataQuery := fmt.Sprintf(`
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, query.SortBy, query.SortOrder, argIndex, argIndex+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search learning contents: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan learning content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate learning contents: %w", err)
	}

	return contents, total, nil
}

// searchWithElasticsearch 使用Elasticsearch搜索
func (r *LearningContentRepositoryImpl) searchWithElasticsearch(ctx context.Context, query *repositories.ContentSearchQuery) ([]*entities.LearningContent, int, error) {
	// 构建ES查询
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"status": "published",
						},
					},
				},
			},
		},
		"from": query.Offset,
		"size": query.Limit,
		"sort": []map[string]interface{}{
			{
				query.SortBy: map[string]interface{}{
					"order": strings.ToLower(query.SortOrder),
				},
			},
		},
	}

	mustQueries := esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	if len(query.Keywords) > 0 {
		for _, keyword := range query.Keywords {
			mustQueries = append(mustQueries, map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  keyword,
					"fields": []string{"title^2", "description", "tags"},
				},
			})
		}
	}

	if query.ContentType != nil {
		mustQueries = append(mustQueries, map[string]interface{}{
			"term": map[string]interface{}{
				"content_type": *query.ContentType,
			},
		})
	}

	if query.DifficultyLevel != nil {
		mustQueries = append(mustQueries, map[string]interface{}{
			"term": map[string]interface{}{
				"difficulty_level": *query.DifficultyLevel,
			},
		})
	}

	if len(query.Tags) > 0 {
		mustQueries = append(mustQueries, map[string]interface{}{
			"terms": map[string]interface{}{
				"tags": query.Tags,
			},
		})
	}

	if query.AuthorID != nil && *query.AuthorID != uuid.Nil {
		mustQueries = append(mustQueries, map[string]interface{}{
			"term": map[string]interface{}{
				"author_id": query.AuthorID.String(),
			},
		})
	}

	esQuery["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries

	// 执行搜索
	queryJSON, _ := json.Marshal(esQuery)
	
	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex("learning_contents"),
		r.es.Search.WithBody(strings.NewReader(string(queryJSON))),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search with elasticsearch: %w", err)
	}
	defer res.Body.Close()

	var esResponse struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source entities.LearningContent `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to decode elasticsearch response: %w", err)
	}

	var contents []*entities.LearningContent
	for _, hit := range esResponse.Hits.Hits {
		content := hit.Source
		contents = append(contents, &content)
	}

	return contents, esResponse.Hits.Total.Value, nil
}

// GetContentStatistics 获取内容统计
func (r *LearningContentRepositoryImpl) GetContentStatistics(ctx context.Context) (*repositories.ContentStatistics, error) {
	// 首先获取基本统计信息
	query := `
		SELECT 
			COUNT(*) as total_content,
			SUM(view_count) as total_views,
			SUM(completion_count) as total_completions,
			AVG(average_rating) as average_rating
		FROM learning_contents
	`

	stats := &repositories.ContentStatistics{}
	var totalViews, totalCompletions sql.NullInt64
	var averageRating sql.NullFloat64
	
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalContent,
		&totalViews,
		&totalCompletions,
		&averageRating,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get content statistics: %w", err)
	}

	// 设置可空字段
	if totalViews.Valid {
		stats.TotalViews = totalViews.Int64
	}
	if totalCompletions.Valid {
		stats.TotalCompletions = totalCompletions.Int64
	}
	if averageRating.Valid {
		stats.AverageRating = averageRating.Float64
	}

	// 获取按类型分组的统计
	stats.ContentByType = make(map[entities.ContentType]int)
	typeQuery := `SELECT content_type, COUNT(*) FROM learning_contents GROUP BY content_type`
	rows, err := r.db.QueryContext(ctx, typeQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get content by type: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var contentType entities.ContentType
		var count int
		if err := rows.Scan(&contentType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan content type: %w", err)
		}
		stats.ContentByType[contentType] = count
	}

	// 获取按状态分组的统计
	stats.ContentByStatus = make(map[entities.ContentStatus]int)
	statusQuery := `SELECT status, COUNT(*) FROM learning_contents GROUP BY status`
	rows, err = r.db.QueryContext(ctx, statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get content by status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status entities.ContentStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan content status: %w", err)
		}
		stats.ContentByStatus[status] = count
	}

	stats.LastUpdated = time.Now()

	return stats, nil
}

// GetPopularContents 获取热门内容
func (r *LearningContentRepositoryImpl) GetPopularContents(ctx context.Context, limit int) ([]*entities.LearningContent, error) {
	query := `
		SELECT lc.id, lc.title, lc.description, lc.content_type, lc.difficulty_level,
			   lc.estimated_duration, lc.tags, lc.metadata, lc.content_data,
			   lc.knowledge_node_ids, lc.prerequisites, lc.learning_objectives,
			   lc.author_id, lc.status, lc.created_at, lc.updated_at,
			   COUNT(lh.id) as activity_count
		FROM learning_contents lc
		LEFT JOIN learning_history lh ON lc.id = lh.content_id
		WHERE lc.status = 'published'
		GROUP BY lc.id
		ORDER BY activity_count DESC, lc.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular contents: %w", err)
	}
	defer rows.Close()

	var contents []*entities.LearningContent

	for rows.Next() {
		content := &entities.LearningContent{}
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte
		var activityCount int

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt, &activityCount,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan popular content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		contents = append(contents, content)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate popular contents: %w", err)
	}

	return contents, nil
}

// indexToElasticsearch 索引到Elasticsearch
func (r *LearningContentRepositoryImpl) indexToElasticsearch(content *entities.LearningContent) {
	if r.es == nil {
		return
	}

	contentJSON, err := json.Marshal(content)
	if err != nil {
		return
	}

	r.es.Index(
		"learning_contents",
		strings.NewReader(string(contentJSON)),
		r.es.Index.WithDocumentID(content.ID.String()),
		r.es.Index.WithRefresh("true"),
	)
}

// deleteFromElasticsearch 从Elasticsearch删除
func (r *LearningContentRepositoryImpl) deleteFromElasticsearch(id uuid.UUID) {
	if r.es == nil {
		return
	}

	r.es.Delete(
		"learning_contents",
		id.String(),
		r.es.Delete.WithRefresh("true"),
	)
}

// AddBookmark 添加书签
func (r *LearningContentRepositoryImpl) AddBookmark(ctx context.Context, bookmark *entities.Bookmark) error {
	query := `
		INSERT INTO bookmarks (id, learner_id, content_id, title, position, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		bookmark.ID, bookmark.LearnerID, bookmark.ContentID,
		bookmark.Title, bookmark.Position, bookmark.Note, bookmark.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	
	return nil
}

// GetBookmarks 获取学习者的书签
func (r *LearningContentRepositoryImpl) GetBookmarks(ctx context.Context, learnerID uuid.UUID, offset, limit int) ([]*entities.Bookmark, error) {
	query := `
		SELECT id, learner_id, content_id, title, position, note, created_at
		FROM bookmarks
		WHERE learner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, learnerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks: %w", err)
	}
	defer rows.Close()
	
	var bookmarks []*entities.Bookmark
	for rows.Next() {
		bookmark := &entities.Bookmark{}
		err := rows.Scan(
			&bookmark.ID, &bookmark.LearnerID, &bookmark.ContentID,
			&bookmark.Title, &bookmark.Position, &bookmark.Note, &bookmark.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, bookmark)
	}
	
	return bookmarks, nil
}

// GetContentBookmarks 获取特定内容的书签
func (r *LearningContentRepositoryImpl) GetContentBookmarks(ctx context.Context, learnerID, contentID uuid.UUID) ([]*entities.Bookmark, error) {
	query := `
		SELECT id, learner_id, content_id, title, position, note, created_at
		FROM bookmarks
		WHERE learner_id = $1 AND content_id = $2
		ORDER BY position ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, learnerID, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content bookmarks: %w", err)
	}
	defer rows.Close()
	
	var bookmarks []*entities.Bookmark
	for rows.Next() {
		bookmark := &entities.Bookmark{}
		err := rows.Scan(
			&bookmark.ID, &bookmark.LearnerID, &bookmark.ContentID,
			&bookmark.Title, &bookmark.Position, &bookmark.Note, &bookmark.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, bookmark)
	}
	
	return bookmarks, nil
}

// UpdateBookmark 更新书签
func (r *LearningContentRepositoryImpl) UpdateBookmark(ctx context.Context, bookmark *entities.Bookmark) error {
	query := `
		UPDATE bookmarks
		SET title = $2, position = $3, note = $4
		WHERE id = $1
	`
	
	_, err := r.db.ExecContext(ctx, query,
		bookmark.ID, bookmark.Title, bookmark.Position, bookmark.Note,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}
	
	return nil
}

// DeleteBookmark 删除书签
func (r *LearningContentRepositoryImpl) DeleteBookmark(ctx context.Context, bookmarkID uuid.UUID) error {
	query := `DELETE FROM bookmarks WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, bookmarkID)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	
	return nil
}

// AddContentFeedback 添加内容反馈
func (r *LearningContentRepositoryImpl) AddContentFeedback(ctx context.Context, feedback *repositories.ContentFeedback) error {
	query := `
		INSERT INTO content_feedback (
			id, content_id, learner_id, rating, difficulty, usefulness,
			clarity, engagement, comments, suggestions, reported_issues, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	reportedIssuesJSON, _ := json.Marshal(feedback.ReportedIssues)
	
	_, err := r.db.ExecContext(ctx, query,
		feedback.ID, feedback.ContentID, feedback.LearnerID,
		feedback.Rating, feedback.Difficulty, feedback.Usefulness,
		feedback.Clarity, feedback.Engagement, feedback.Comments,
		feedback.Suggestions, reportedIssuesJSON, feedback.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to add content feedback: %w", err)
	}
	
	return nil
}

// AddInteractionRecord 添加交互记录
func (r *LearningContentRepositoryImpl) AddInteractionRecord(ctx context.Context, record *entities.InteractionRecord) error {
	query := `
		INSERT INTO interaction_records (
			id, learner_id, content_id, type, element, position, data, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	dataJSON, _ := json.Marshal(record.Data)
	
	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.LearnerID, record.ContentID,
		record.Type, record.Element, record.Position,
		dataJSON, record.Timestamp,
	)
	
	if err != nil {
		return fmt.Errorf("failed to add interaction record: %w", err)
	}
	
	return nil
}

// AddLearningNote 添加学习笔记
func (r *LearningContentRepositoryImpl) AddLearningNote(ctx context.Context, note *entities.LearningNote) error {
	query := `
		INSERT INTO learning_notes (
			id, learner_id, content_id, content, position, tags, is_public, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	tagsJSON, _ := json.Marshal(note.Tags)
	
	_, err := r.db.ExecContext(ctx, query,
		note.ID, note.LearnerID, note.ContentID, note.Content,
		note.Position, tagsJSON, note.IsPublic, note.CreatedAt, note.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to add learning note: %w", err)
	}
	
	return nil
}

// UpdateLearningNote 更新学习笔记
func (r *LearningContentRepositoryImpl) UpdateLearningNote(ctx context.Context, note *entities.LearningNote) error {
	query := `
		UPDATE learning_notes 
		SET content = $2, position = $3, tags = $4, is_public = $5, updated_at = $6
		WHERE id = $1
	`
	
	tagsJSON, _ := json.Marshal(note.Tags)
	note.UpdatedAt = time.Now()
	
	result, err := r.db.ExecContext(ctx, query,
		note.ID, note.Content, note.Position, tagsJSON, note.IsPublic, note.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update learning note: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("learning note not found: %s", note.ID)
	}
	
	return nil
}

// DeleteLearningNote 删除学习笔记
func (r *LearningContentRepositoryImpl) DeleteLearningNote(ctx context.Context, noteID uuid.UUID) error {
	query := `DELETE FROM learning_notes WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, noteID)
	if err != nil {
		return fmt.Errorf("failed to delete learning note: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("learning note with ID %s not found", noteID)
	}
	
	return nil
}

// ExportContent 导出内容
func (r *LearningContentRepositoryImpl) ExportContent(ctx context.Context, contentIDs []uuid.UUID, format string) ([]byte, error) {
	if len(contentIDs) == 0 {
		return []byte("[]"), nil
	}

	// 构建查询条件
	placeholders := make([]string, len(contentIDs))
	args := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents 
		WHERE id IN (%s)
		ORDER BY created_at DESC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query content for export: %w", err)
	}
	defer rows.Close()

	var contents []map[string]interface{}
	for rows.Next() {
		var content entities.LearningContent
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		// 转换为导出格式
		contentMap := map[string]interface{}{
			"id":                  content.ID,
			"title":               content.Title,
			"description":         content.Description,
			"content_type":        content.Type,
			"content_data":        content.Content,
			"difficulty_level":    content.Difficulty,
			"estimated_duration":  content.EstimatedDuration,
			"tags":                content.Tags,
			"metadata":            content.Metadata,
			"knowledge_node_ids":  content.KnowledgeNodeIDs,
			"prerequisites":       content.Prerequisites,
			"learning_objectives": content.LearningObjectives,
			"author_id":           content.CreatedBy,
			"status":              content.Status,
			"created_at":          content.CreatedAt,
			"updated_at":          content.UpdatedAt,
		}
		contents = append(contents, contentMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over content rows: %w", err)
	}

	// 根据格式导出
	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(contents, "", "  ")
	case "csv":
		// 简化的CSV导出
		var csvData []string
		csvData = append(csvData, "id,title,description,content_type,difficulty_level,status,created_at")
		for _, content := range contents {
			line := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v",
				content["id"], content["title"], content["description"],
				content["content_type"], content["difficulty_level"],
				content["status"], content["created_at"])
			csvData = append(csvData, line)
		}
		return []byte(strings.Join(csvData, "\n")), nil
	default:
		// 默认返回JSON格式
		return json.MarshalIndent(contents, "", "  ")
	}
}

// BackupContent 备份内容
func (r *LearningContentRepositoryImpl) BackupContent(ctx context.Context, contentIDs []uuid.UUID) (*repositories.ContentBackup, error) {
	if len(contentIDs) == 0 {
		return &repositories.ContentBackup{
			ID:          uuid.New(),
			ContentIDs:  []uuid.UUID{},
			BackupData:  []byte("[]"),
			CreatedAt:   time.Now(),
			CreatedBy:   uuid.Nil,
			Description: "Empty backup",
			Size:        2,
			Checksum:    "",
		}, nil
	}

	// 构建查询条件
	placeholders := make([]string, len(contentIDs))
	args := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, content_type, difficulty_level,
			   estimated_duration, tags, metadata, content_data,
			   knowledge_node_ids, prerequisites, learning_objectives,
			   author_id, status, created_at, updated_at
		FROM learning_contents 
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query content for backup: %w", err)
	}
	defer rows.Close()

	var contents []map[string]interface{}
	for rows.Next() {
		var content entities.LearningContent
		var tagsJSON, metadataJSON, contentDataJSON, nodeIdsJSON, prerequisitesJSON, objectivesJSON []byte

		err := rows.Scan(
			&content.ID, &content.Title, &content.Description, &content.Type,
			&content.Difficulty, &content.EstimatedDuration, &tagsJSON,
			&metadataJSON, &contentDataJSON, &nodeIdsJSON, &prerequisitesJSON,
			&objectivesJSON, &content.CreatedBy, &content.Status,
			&content.CreatedAt, &content.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}

		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &content.Content)
		}
		if len(nodeIdsJSON) > 0 {
			json.Unmarshal(nodeIdsJSON, &content.KnowledgeNodeIDs)
		}
		if len(prerequisitesJSON) > 0 {
			json.Unmarshal(prerequisitesJSON, &content.Prerequisites)
		}
		if len(objectivesJSON) > 0 {
			json.Unmarshal(objectivesJSON, &content.LearningObjectives)
		}

		// 转换为map用于备份
		contentMap := map[string]interface{}{
			"id":                  content.ID,
			"title":               content.Title,
			"description":         content.Description,
			"content_type":        content.Type,
			"content_data":        content.Content,
			"difficulty_level":    content.Difficulty,
			"estimated_duration":  content.EstimatedDuration,
			"tags":                content.Tags,
			"metadata":            content.Metadata,
			"knowledge_node_ids":  content.KnowledgeNodeIDs,
			"prerequisites":       content.Prerequisites,
			"learning_objectives": content.LearningObjectives,
			"author_id":           content.CreatedBy,
			"status":              content.Status,
			"created_at":          content.CreatedAt,
			"updated_at":          content.UpdatedAt,
		}
		contents = append(contents, contentMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over content rows: %w", err)
	}

	// 序列化备份数据
	backupData, err := json.Marshal(contents)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// 计算校验和
	hasher := sha256.New()
	hasher.Write(backupData)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	return &repositories.ContentBackup{
		ID:          uuid.New(),
		ContentIDs:  contentIDs,
		BackupData:  backupData,
		CreatedAt:   time.Now(),
		CreatedBy:   uuid.Nil,
		Description: fmt.Sprintf("Backup of %d content items", len(contents)),
		Size:        int64(len(backupData)),
		Checksum:    checksum,
	}, nil
}

// BatchCreate 批量创建学习内容
func (r *LearningContentRepositoryImpl) BatchCreate(ctx context.Context, contents []*entities.LearningContent) error {
	if len(contents) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO learning_contents (
			id, title, description, type, content, difficulty,
			estimated_duration, tags, created_by, status,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, content := range contents {
		if content.ID == uuid.Nil {
			content.ID = uuid.New()
		}
		if content.CreatedAt.IsZero() {
			content.CreatedAt = time.Now()
		}
		if content.UpdatedAt.IsZero() {
			content.UpdatedAt = time.Now()
		}

		tagsJSON, _ := json.Marshal(content.Tags)

		_, err := stmt.ExecContext(ctx,
			content.ID, content.Title, content.Description, content.Type,
			content.Content, content.Difficulty, content.EstimatedDuration,
			tagsJSON, content.CreatedBy, content.Status,
			content.CreatedAt, content.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to insert content %s: %w", content.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}