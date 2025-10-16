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

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearningContentRepositoryImpl 
type LearningContentRepositoryImpl struct {
	db *sql.DB
	es *elasticsearch.Client
}

// NewLearningContentRepository 
func NewLearningContentRepository(db *sql.DB, es *elasticsearch.Client) repositories.LearningContentRepository {
	return &LearningContentRepositoryImpl{
		db: db,
		es: es,
	}
}

// Create 
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

	// Elasticsearch
	if r.es != nil {
		go r.indexToElasticsearch(content)
	}

	return nil
}

// UpdateProgress 
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

// UpdateContentAnalytics 
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

// SearchNotes 
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

		// tags JSON
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

// RestoreContentVersion 汾
func (r *LearningContentRepositoryImpl) RestoreContentVersion(ctx context.Context, contentID, versionID uuid.UUID) error {
	// ?
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 汾
	var versionData []byte
	query := `SELECT content_data FROM content_versions WHERE id = $1 AND content_id = $2`
	err = tx.QueryRowContext(ctx, query, versionID, contentID).Scan(&versionData)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("content version not found")
		}
		return fmt.Errorf("failed to get version data: %w", err)
	}

	// 汾
	var content entities.LearningContent
	if err := json.Unmarshal(versionData, &content); err != nil {
		return fmt.Errorf("failed to unmarshal version data: %w", err)
	}

	// ?
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

	// 
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ValidateContent 
func (r *LearningContentRepositoryImpl) ValidateContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentValidation, error) {
	// 
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

	// 
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

	// 
	qualityScore := r.calculateQualityScore(content, validation)
	validation.QualityScore = qualityScore

	// 
	validation.Recommendations = r.generateRecommendations(content, validation)

	return validation, nil
}

// calculateQualityScore 
func (r *LearningContentRepositoryImpl) calculateQualityScore(content *entities.LearningContent, validation *repositories.ContentValidation) float64 {
	score := 100.0

	// ?
	score -= float64(len(validation.Errors)) * 20.0  // ?0?
	score -= float64(len(validation.Warnings)) * 5.0 // ??

	// ?
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

	// ?-100?
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// generateRecommendations 
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

// SearchByKeywords ?
func (r *LearningContentRepositoryImpl) SearchByKeywords(ctx context.Context, keywords []string, offset, limit int) ([]*entities.LearningContent, error) {
	if len(keywords) == 0 {
		return []*entities.LearningContent{}, nil
	}

	// 
	whereConditions := []string{"status = 'published'"}
	args := []interface{}{}
	argIndex := 1

	// 
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

		// JSON
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

// RestoreContent ?
func (r *LearningContentRepositoryImpl) RestoreContent(ctx context.Context, backupID uuid.UUID) error {
	// 
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
	
	// 
	var contents []*entities.LearningContent
	if err := json.Unmarshal(backupData, &contents); err != nil {
		return fmt.Errorf("failed to unmarshal backup data: %w", err)
	}
	
	// ?
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// 
	for _, content := range contents {
		// 
		var existingID uuid.UUID
		checkQuery := `SELECT id FROM learning_contents WHERE id = $1`
		err := tx.QueryRowContext(ctx, checkQuery, content.ID).Scan(&existingID)
		
		if err == sql.ErrNoRows {
			// ?
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
			// 
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
	
	// 
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// ImportContent 
func (r *LearningContentRepositoryImpl) ImportContent(ctx context.Context, data []byte, format string) ([]*entities.LearningContent, error) {
	var contents []*entities.LearningContent
	
	switch strings.ToLower(format) {
	case "json":
		if err := json.Unmarshal(data, &contents); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}
	case "csv":
		// CSV
		reader := csv.NewReader(strings.NewReader(string(data)))
		records, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV data: %w", err)
		}
		
		// ?
		if len(records) > 1 {
			for i := 1; i < len(records); i++ {
				record := records[i]
				if len(record) >= 4 { // title, description, type, difficulty
					content := &entities.LearningContent{
						ID:          uuid.New(),
						Title:       record[0],
						Description: record[1],
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					}
					
					// 
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
					
					// 
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
	
	// 浼?
	for _, content := range contents {
		if err := r.Create(ctx, content); err != nil {
			return nil, fmt.Errorf("failed to save imported content %s: %w", content.ID, err)
		}
	}
	
	return contents, nil
}

// OptimizeContent 
func (r *LearningContentRepositoryImpl) OptimizeContent(ctx context.Context, contentID uuid.UUID) (*repositories.ContentOptimization, error) {
	// 
	content, err := r.GetByID(ctx, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	
	// ?
	var suggestions []repositories.OptimizationSuggestion
	optimizationScore := 0.8 // 
	
	// ?
	if len(content.Title) < 10 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "title_length",
			Description:    "10?,
			Priority:       2,
			ExpectedImpact: 0.1,
			Implementation: "",
			Resources:      []string{"SEO", "?},
		})
		optimizationScore -= 0.1
	}
	
	// ?
	if len(content.Description) < 50 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "description_length",
			Description:    "50?,
			Priority:       3,
			ExpectedImpact: 0.15,
			Implementation: "",
			Resources:      []string{""},
		})
		optimizationScore -= 0.15
	}
	
	// ?
	if len(content.LearningObjectives) == 0 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "learning_objectives",
			Description:    "?,
			Priority:       1,
			ExpectedImpact: 0.2,
			Implementation: "3-5?,
			Resources:      []string{"", "SMART"},
		})
		optimizationScore -= 0.2
	}
	
	// ?
	if len(content.Tags) < 3 {
		suggestions = append(suggestions, repositories.OptimizationSuggestion{
			Type:           "tags",
			Description:    "㽨?,
			Priority:       3,
			ExpectedImpact: 0.1,
			Implementation: "3-8",
			Resources:      []string{""},
		})
		optimizationScore -= 0.1
	}
	
	// 
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

// PublishContent 
func (r *LearningContentRepositoryImpl) PublishContent(ctx context.Context, contentID uuid.UUID) error {
	// 
	content, err := r.GetByID(ctx, contentID)
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}
	
	// ?
	if content.Status == entities.ContentStatusPublished {
		return fmt.Errorf("content is already published")
	}
	
	// 㷢
	if content.Title == "" {
		return fmt.Errorf("content title is required for publishing")
	}
	if content.Description == "" {
		return fmt.Errorf("content description is required for publishing")
	}
	if len(content.LearningObjectives) == 0 {
		return fmt.Errorf("learning objectives are required for publishing")
	}
	
	// ?
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

// GetPersonalizedContent 
func (r *LearningContentRepositoryImpl) GetPersonalizedContent(ctx context.Context, learnerID uuid.UUID, preferences *repositories.ContentPreferences, limit int) ([]*entities.LearningContent, error) {
	// 
	query := `
		SELECT id, title, description, type, content, difficulty, estimated_duration,
		       tags, created_by, status, created_at, updated_at
		FROM learning_contents
		WHERE status = 'published'
	`
	
	args := []interface{}{}
	argIndex := 1

	// 
	if preferences != nil {
		// 
		if len(preferences.PreferredTypes) > 0 {
			placeholders := make([]string, len(preferences.PreferredTypes))
			for i, contentType := range preferences.PreferredTypes {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, string(contentType))
				argIndex++
			}
			query += fmt.Sprintf(" AND type IN (%s)", strings.Join(placeholders, ","))
		}

		// 
		if preferences.PreferredDifficulty > 0 {
			query += fmt.Sprintf(" AND difficulty = $%d", argIndex)
			args = append(args, int(preferences.PreferredDifficulty))
			argIndex++
		}

		// ?
		if len(preferences.InterestAreas) > 0 {
			for _, area := range preferences.InterestAreas {
				query += fmt.Sprintf(" AND tags::text LIKE $%d", argIndex)
				args = append(args, "%"+area+"%")
				argIndex++
			}
		}

		// ?
		if len(preferences.AvoidTopics) > 0 {
			for _, topic := range preferences.AvoidTopics {
				query += fmt.Sprintf(" AND tags::text NOT LIKE $%d", argIndex)
				args = append(args, "%"+topic+"%")
				argIndex++
			}
		}
	}

	// ?
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

		// 
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &content.Tags); err != nil {
				// ?
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

// GetPopularContent 
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

		// 
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

// GetProgress 
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
			return nil, nil // 
		}
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	return progress, nil
}

// GetRecommendedContent 
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

		// 
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &content.Tags); err != nil {
				content.Tags = []string{}
			}
		}

		recommendation.Content = content
		recommendation.Reasoning = []string{
			"",
			"",
			"",
		}
		recommendation.PersonalizationFactors = []string{
			"",
			"",
			"",
		}
		recommendation.Priority = 1

		recommendations = append(recommendations, recommendation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return recommendations, nil
}

// GetLearningNotes 
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

		// tags JSON
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

// GetInteractionRecords 
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
		
		// data
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

// GetInteractionsByType 
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
		
		// data
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

// GetPrerequisites 
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

// GetFollowUpContent 
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

// GetRelatedContent 
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



// scanContent ?
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

	// JSON
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

// GetFeedbackSummary 
func (r *LearningContentRepositoryImpl) GetFeedbackSummary(ctx context.Context, contentID uuid.UUID) (*entities.FeedbackSummary, error) {
	// ?
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

	//  (-1.0 ?1.0)
	sentimentScore := 0.0
	if avgRating.Valid {
		// ?-5-1??
		sentimentScore = (avgRating.Float64 - 3.0) / 2.0
	}

	// ?
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

		// 
		if comments.Valid && comments.String != "" {
			// NLP?
			if strings.Contains(strings.ToLower(comments.String), "difficult") ||
			   strings.Contains(strings.ToLower(comments.String), "hard") {
				issueMap[""]++
			}
			if strings.Contains(strings.ToLower(comments.String), "confusing") ||
			   strings.Contains(strings.ToLower(comments.String), "unclear") {
				issueMap[""]++
			}
			if strings.Contains(strings.ToLower(comments.String), "boring") ||
			   strings.Contains(strings.ToLower(comments.String), "uninteresting") {
				issueMap["?]++
			}
		}

		// 
		if suggestionsText.Valid && suggestionsText.String != "" {
			if strings.Contains(strings.ToLower(suggestionsText.String), "example") {
				suggestionMap[""]++
			}
			if strings.Contains(strings.ToLower(suggestionsText.String), "practice") {
				suggestionMap[""]++
			}
			if strings.Contains(strings.ToLower(suggestionsText.String), "visual") {
				suggestionMap[""]++
			}
		}

		// ?
		if reportedIssuesJSON.Valid && reportedIssuesJSON.String != "" && reportedIssuesJSON.String != "[]" {
			var reportedIssues []string
			if err := json.Unmarshal([]byte(reportedIssuesJSON.String), &reportedIssues); err == nil {
				for _, issue := range reportedIssues {
					issueMap[issue]++
				}
			}
		}
	}

	// 5
	for issue, count := range issueMap {
		if count >= 2 { // ??
			commonIssues = append(commonIssues, issue)
		}
	}
	for suggestion, count := range suggestionMap {
		if count >= 2 { // ??
			suggestions = append(suggestions, suggestion)
		}
	}

	// 
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

// GetContentVersions ?
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

		// changes JSON
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

// GetContentSequence ID
func (r *LearningContentRepositoryImpl) GetContentSequence(ctx context.Context, sequenceID uuid.UUID) ([]*entities.LearningContent, error) {
	// ID
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

	// ID
	var contentIDs []uuid.UUID
	if err := json.Unmarshal(contentIDsJSON, &contentIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content IDs: %w", err)
	}

	if len(contentIDs) == 0 {
		return []*entities.LearningContent{}, nil
	}

	// 
	placeholders := make([]string, len(contentIDs))
	args := make([]interface{}, len(contentIDs))
	for i, id := range contentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// 
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

	// 
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

		// JSON
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

	// ?
	var orderedContents []*entities.LearningContent
	for _, contentID := range contentIDs {
		if content, exists := contentMap[contentID]; exists {
			orderedContents = append(orderedContents, content)
		}
	}

	return orderedContents, nil
}

// GetContentProgressStats 
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

	// 
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

// GetContentRating 
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

	// 
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

// GetContentReviews 
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

		//  pros JSON 
		if prosJSON.Valid && prosJSON.String != "" {
			if err := json.Unmarshal([]byte(prosJSON.String), &review.Pros); err != nil {
				review.Pros = []string{} // ?
			}
		} else {
			review.Pros = []string{}
		}

		//  cons JSON 
		if consJSON.Valid && consJSON.String != "" {
			if err := json.Unmarshal([]byte(consJSON.String), &review.Cons); err != nil {
				review.Cons = []string{} // ?
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

// GetContentEffectiveness ?
func (r *LearningContentRepositoryImpl) GetContentEffectiveness(ctx context.Context, contentID uuid.UUID) (*repositories.ContentEffectiveness, error) {
	// 
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

	// ?
	learningEffectiveness := 0.0
	if viewCount > 0 {
		learningEffectiveness = float64(completionCount) / float64(viewCount)
	}

	// 㼼?
	skillImprovement := learningEffectiveness * (averageRating / 5.0)

	// 
	knowledgeRetention := averageRating / 5.0 * 0.8

	// ?
	engagementScore := learningEffectiveness * 0.7 + (averageRating / 5.0) * 0.3

	// 
	completionQuality := averageRating / 5.0

	// 
	learnerSatisfaction := averageRating / 5.0

	// ?
	recommendationRate := 0.0
	if averageRating >= 4.0 {
		recommendationRate = 0.8
	} else if averageRating >= 3.0 {
		recommendationRate = 0.5
	} else {
		recommendationRate = 0.2
	}

	// 
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

// GetLearningOutcomes 
func (r *LearningContentRepositoryImpl) GetLearningOutcomes(ctx context.Context, contentID uuid.UUID) (*repositories.LearningOutcomes, error) {
	// 
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

	// ?
	skillsAcquired := map[string]float64{
		"problem_solving": averageRating / 5.0 * 0.8,
		"critical_thinking": averageRating / 5.0 * 0.7,
		"communication": averageRating / 5.0 * 0.6,
	}

	// 
	knowledgeGained := map[string]float64{
		"theoretical_knowledge": averageRating / 5.0 * 0.9,
		"practical_knowledge": averageRating / 5.0 * 0.8,
		"domain_expertise": averageRating / 5.0 * 0.7,
	}

	// 
	competenciesImproved := map[string]float64{
		"analytical_skills": averageRating / 5.0 * 0.8,
		"technical_skills": averageRating / 5.0 * 0.9,
		"soft_skills": averageRating / 5.0 * 0.6,
	}

	// 
	learningObjectivesMet := map[uuid.UUID]float64{
		uuid.New(): averageRating / 5.0 * 0.9,
		uuid.New(): averageRating / 5.0 * 0.8,
	}

	// 㳤?
	longTermRetention := averageRating / 5.0 * 0.7

	// ?
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

// GetByType 
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

		// JSON
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

// CreateContentVersion 汾
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

// CreateProgress 
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

// GetLearnerProgress 
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

		// JSON
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

// CompareContentVersions 汾
func (r *LearningContentRepositoryImpl) CompareContentVersions(ctx context.Context, contentID, version1ID, version2ID uuid.UUID) (*repositories.ContentComparison, error) {
	// 汾?
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

		// JSON
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

	// 汾
	comparison := &repositories.ContentComparison{
		ContentID:  contentID,
		Version1ID: version1ID,
		Version2ID: version2ID,
	}

	var differences []repositories.ContentDifference

	// 
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

	// 
	if len(differences) == 0 {
		comparison.Summary = "No differences found between versions"
	} else {
		comparison.Summary = fmt.Sprintf("Found %d differences between versions", len(differences))
	}

	// 
	var majorChanges []string
	for _, diff := range differences {
		if diff.Impact == "high" {
			majorChanges = append(majorChanges, fmt.Sprintf("%s changed", diff.Field))
		}
	}
	comparison.MajorChanges = majorChanges

	return comparison, nil
}

// BatchUpdateStatus ?
func (r *LearningContentRepositoryImpl) BatchUpdateStatus(ctx context.Context, contentIDs []uuid.UUID, status entities.ContentStatus) error {
	if len(contentIDs) == 0 {
		return nil
	}

	// ?
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

// BatchUpdate 
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

// GetContentInteractions ?
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

		// JSON
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

// GetContentAnalytics 
func (r *LearningContentRepositoryImpl) GetContentAnalytics(ctx context.Context, contentID uuid.UUID) (*entities.ContentAnalytics, error) {
	// 
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

	// ?
	completionRate := 0.0
	if viewCount > 0 {
		completionRate = float64(completionCount) / float64(viewCount)
	}

	// 
	var interactionData map[string]interface{}
	if analyticsDataJSON.Valid && analyticsDataJSON.String != "" {
		if err := json.Unmarshal([]byte(analyticsDataJSON.String), &interactionData); err != nil {
			interactionData = make(map[string]interface{})
		}
	} else {
		interactionData = make(map[string]interface{})
	}

	// 
	dropoffPoints := []int{10, 25, 50, 75, 90} // 

	// 
	feedbackSummary := entities.FeedbackSummary{
		TotalFeedback:   0,
		PositiveCount:   0,
		NegativeCount:   0,
		CommonIssues:    []string{},
		Suggestions:     []string{},
		SentimentScore:  0.0,
	}

	// 
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
		
		// 淴
		if avgFeedbackRating.Valid {
			rating := avgFeedbackRating.Float64
			feedbackSummary.SentimentScore = (rating - 3.0) / 2.0 // ?1??
			
			if rating >= 4.0 {
				feedbackSummary.PositiveCount = int(float64(feedbackCount) * 0.7) // 
			} else if rating <= 2.0 {
				feedbackSummary.NegativeCount = int(float64(feedbackCount) * 0.7) // 
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

// BatchDelete 
func (r *LearningContentRepositoryImpl) BatchDelete(ctx context.Context, contentIDs []uuid.UUID) error {
	if len(contentIDs) == 0 {
		return nil
	}

	// ?
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

// ArchiveContent 鵵
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

// GetByID ID
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

	// JSON
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

// Update 
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

	// Elasticsearch
	if r.es != nil {
		go r.indexToElasticsearch(content)
	}

	return nil
}

// Delete 
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

	// Elasticsearch
	if r.es != nil {
		go r.deleteFromElasticsearch(id)
	}

	return nil
}

// List 
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

		// JSON
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

// GetByKnowledgeNode 
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

		// JSON
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

// GetByTags 
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

		// JSON
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

// GetByAuthor ID
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

		// JSON
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

// GetByDifficulty 
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

		// JSON
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

// GetByStatus ?
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

		// JSON
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

// Search 
func (r *LearningContentRepositoryImpl) Search(ctx context.Context, query *repositories.ContentSearchQuery) ([]*entities.LearningContent, int, error) {
	// ElasticsearchES
	if r.es != nil && len(query.Keywords) > 0 {
		return r.searchWithElasticsearch(ctx, query)
	}

	// PostgreSQL
	return r.searchWithPostgreSQL(ctx, query)
}

// searchWithPostgreSQL PostgreSQL
func (r *LearningContentRepositoryImpl) searchWithPostgreSQL(ctx context.Context, query *repositories.ContentSearchQuery) ([]*entities.LearningContent, int, error) {
	// WHERE
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

	// 
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

	// 
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

		// JSON
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

// searchWithElasticsearch Elasticsearch
func (r *LearningContentRepositoryImpl) searchWithElasticsearch(ctx context.Context, query *repositories.ContentSearchQuery) ([]*entities.LearningContent, int, error) {
	// ES
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

	// 
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

// GetContentStatistics 
func (r *LearningContentRepositoryImpl) GetContentStatistics(ctx context.Context) (*repositories.ContentStatistics, error) {
	// 
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

	// 
	if totalViews.Valid {
		stats.TotalViews = totalViews.Int64
	}
	if totalCompletions.Valid {
		stats.TotalCompletions = totalCompletions.Int64
	}
	if averageRating.Valid {
		stats.AverageRating = averageRating.Float64
	}

	// 
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

	// 
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

// GetPopularContents 
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

		// JSON
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

// indexToElasticsearch Elasticsearch
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

// deleteFromElasticsearch Elasticsearch
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

// AddBookmark 
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

// GetBookmarks 
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

// GetContentBookmarks ?
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

// UpdateBookmark 
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

// DeleteBookmark 
func (r *LearningContentRepositoryImpl) DeleteBookmark(ctx context.Context, bookmarkID uuid.UUID) error {
	query := `DELETE FROM bookmarks WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, bookmarkID)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}
	
	return nil
}

// AddContentFeedback 
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

// AddInteractionRecord 
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

// AddLearningNote 
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

// UpdateLearningNote 
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

// DeleteLearningNote 
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

// ExportContent 
func (r *LearningContentRepositoryImpl) ExportContent(ctx context.Context, contentIDs []uuid.UUID, format string) ([]byte, error) {
	if len(contentIDs) == 0 {
		return []byte("[]"), nil
	}

	// 
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

		// JSON
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

		// ?
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

	// 
	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(contents, "", "  ")
	case "csv":
		// CSV
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
		// JSON
		return json.MarshalIndent(contents, "", "  ")
	}
}

// BackupContent 
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

	// 
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

		// JSON
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

		// map
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

	// ?
	backupData, err := json.Marshal(contents)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal backup data: %w", err)
	}

	// ?
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

// BatchCreate 
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

