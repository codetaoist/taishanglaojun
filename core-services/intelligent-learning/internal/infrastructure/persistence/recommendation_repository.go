package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	domainServices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// RecommendationRepository 推荐系统数据访问�?
type RecommendationRepository struct {
	db *sql.DB
}

// NewRecommendationRepository 创建推荐系统仓储
func NewRecommendationRepository(db *sql.DB) *RecommendationRepository {
	return &RecommendationRepository{
		db: db,
	}
}

// 实现 PreferenceRepository 接口

func (r *RecommendationRepository) GetUserPreferences(ctx context.Context, userID string) (*domainServices.UserPreferences, error) {
	query := `
		SELECT user_id, content_preferences, learning_preferences, interaction_patterns, 
			   difficulty_preference, time_preferences, device_preferences, updated_at
		FROM user_preferences 
		WHERE user_id = $1`
	
	var prefs domainServices.UserPreferences
	var contentPrefsJSON, learningPrefsJSON, interactionPatternsJSON []byte
	var difficultyPrefJSON, timePrefsJSON, devicePrefsJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID,
		&contentPrefsJSON,
		&learningPrefsJSON,
		&interactionPatternsJSON,
		&difficultyPrefJSON,
		&timePrefsJSON,
		&devicePrefsJSON,
		&prefs.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 用户偏好不存�?
		}
		return nil, fmt.Errorf("获取用户偏好失败: %w", err)
	}
	
	// 解析JSON字段
	if err := json.Unmarshal(contentPrefsJSON, &prefs.ContentPreferences); err != nil {
		return nil, fmt.Errorf("解析内容偏好失败: %w", err)
	}
	if err := json.Unmarshal(learningPrefsJSON, &prefs.LearningPreferences); err != nil {
		return nil, fmt.Errorf("解析学习偏好失败: %w", err)
	}
	if err := json.Unmarshal(interactionPatternsJSON, &prefs.InteractionPatterns); err != nil {
		return nil, fmt.Errorf("解析交互模式失败: %w", err)
	}
	if err := json.Unmarshal(difficultyPrefJSON, &prefs.DifficultyPreference); err != nil {
		return nil, fmt.Errorf("解析难度偏好失败: %w", err)
	}
	if err := json.Unmarshal(timePrefsJSON, &prefs.TimePreferences); err != nil {
		return nil, fmt.Errorf("解析时间偏好失败: %w", err)
	}
	if err := json.Unmarshal(devicePrefsJSON, &prefs.DevicePreferences); err != nil {
		return nil, fmt.Errorf("解析设备偏好失败: %w", err)
	}
	
	return &prefs, nil
}

func (r *RecommendationRepository) SaveUserPreferences(ctx context.Context, prefs *domainServices.UserPreferences) error {
	// 序列化JSON字段
	contentPrefsJSON, _ := json.Marshal(prefs.ContentPreferences)
	learningPrefsJSON, _ := json.Marshal(prefs.LearningPreferences)
	interactionPatternsJSON, _ := json.Marshal(prefs.InteractionPatterns)
	difficultyPrefJSON, _ := json.Marshal(prefs.DifficultyPreference)
	timePrefsJSON, _ := json.Marshal(prefs.TimePreferences)
	devicePrefsJSON, _ := json.Marshal(prefs.DevicePreferences)
	
	query := `
		INSERT INTO user_preferences (
			user_id, content_preferences, learning_preferences, interaction_patterns,
			difficulty_preference, time_preferences, device_preferences, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			content_preferences = EXCLUDED.content_preferences,
			learning_preferences = EXCLUDED.learning_preferences,
			interaction_patterns = EXCLUDED.interaction_patterns,
			difficulty_preference = EXCLUDED.difficulty_preference,
			time_preferences = EXCLUDED.time_preferences,
			device_preferences = EXCLUDED.device_preferences,
			updated_at = EXCLUDED.updated_at`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		prefs.UserID,
		contentPrefsJSON,
		learningPrefsJSON,
		interactionPatternsJSON,
		difficultyPrefJSON,
		timePrefsJSON,
		devicePrefsJSON,
		now,
		now,
	)
	
	if err != nil {
		return fmt.Errorf("保存用户偏好失败: %w", err)
	}
	
	return nil
}

func (r *RecommendationRepository) GetPreferenceHistory(ctx context.Context, userID string, limit int) ([]*domainServices.PreferenceHistory, error) {
	query := `
		SELECT user_id, timestamp, action, content_id, context, preferences, confidence
		FROM preference_history 
		WHERE user_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取偏好历史失败: %w", err)
	}
	defer rows.Close()
	
	var history []*domainServices.PreferenceHistory
	for rows.Next() {
		var h domainServices.PreferenceHistory
		var contextJSON, preferencesJSON []byte
		
		err := rows.Scan(
			&h.UserID,
			&h.Timestamp,
			&h.Action,
			&h.ContentID,
			&contextJSON,
			&preferencesJSON,
			&h.Confidence,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描偏好历史失败: %w", err)
		}
		
		// 解析JSON字段
		if len(contextJSON) > 0 {
			json.Unmarshal(contextJSON, &h.Context)
		}
		if len(preferencesJSON) > 0 {
			json.Unmarshal(preferencesJSON, &h.Preferences)
		}
		
		history = append(history, &h)
	}
	
	return history, nil
}

// 实现 ContentRepository 接口

func (r *RecommendationRepository) GetContentByIDs(ctx context.Context, contentIDs []string) ([]*domainServices.Content, error) {
	if len(contentIDs) == 0 {
		return []*domainServices.Content{}, nil
	}
	
	query := `
		SELECT id, title, description, content_type, tags, metadata, difficulty_level, 
			   estimated_duration, created_at, updated_at
		FROM learning_content 
		WHERE id = ANY($1)`
	
	rows, err := r.db.QueryContext(ctx, query, pq.Array(contentIDs))
	if err != nil {
		return nil, fmt.Errorf("获取内容失败: %w", err)
	}
	defer rows.Close()
	
	var contents []*domainServices.Content
	for rows.Next() {
		var content domainServices.Content
		var tagsJSON, metadataJSON []byte
		
		err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.Category,
			&tagsJSON,
			&metadataJSON,
			&content.Difficulty,
			&content.Duration,
			&content.Format,
			&content.Language,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描内容失败: %w", err)
		}
		
		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		
		contents = append(contents, &content)
	}
	
	return contents, nil
}

func (r *RecommendationRepository) SearchContent(ctx context.Context, criteria *domainServices.ContentSearchCriteria) ([]*domainServices.Content, error) {
	query := `
		SELECT id, title, description, category, tags, metadata, difficulty, 
			   duration, created_at, updated_at
		FROM content 
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	// 构建动态查询条�?
	if criteria.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, criteria.Category)
		argIndex++
	}
	
	if criteria.Difficulty != "" {
		query += fmt.Sprintf(" AND difficulty = $%d", argIndex)
		args = append(args, criteria.Difficulty)
		argIndex++
	}
	
	if len(criteria.Tags) > 0 {
		query += fmt.Sprintf(" AND tags ?| $%d", argIndex)
		args = append(args, pq.Array(criteria.Tags))
		argIndex++
	}
	
	if criteria.MinDuration > 0 {
		query += fmt.Sprintf(" AND duration >= $%d", argIndex)
		args = append(args, criteria.MinDuration)
		argIndex++
	}
	
	if criteria.MaxDuration > 0 {
		query += fmt.Sprintf(" AND duration <= $%d", argIndex)
		args = append(args, criteria.MaxDuration)
		argIndex++
	}
	
	// 添加排序和限�?
	query += " ORDER BY created_at DESC"
	if criteria.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, criteria.Limit)
	}
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("搜索内容失败: %w", err)
	}
	defer rows.Close()
	
	var contents []*domainServices.Content
	for rows.Next() {
		var content domainServices.Content
		var tagsJSON, metadataJSON []byte
		
		err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.Category,
			&tagsJSON,
			&content.Difficulty,
			&content.Duration,
			&content.Format,
			&content.Language,
			&content.Quality,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描内容失败: %w", err)
		}
		
		// 解析JSON字段
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &content.Tags)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &content.Metadata)
		}
		
		contents = append(contents, &content)
	}
	
	return contents, nil
}

// 实现 UserRepository 接口

func (r *RecommendationRepository) GetLearningRecords(ctx context.Context, userID string, limit int) ([]*domainServices.LearningRecord, error) {
	query := `
		SELECT id, user_id, content_id, action_type, duration, score, completion_rate, 
			   metadata, created_at
		FROM learning_records 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取学习记录失败: %w", err)
	}
	defer rows.Close()
	
	var records []*domainServices.LearningRecord
	for rows.Next() {
		var record domainServices.LearningRecord
		var metadataJSON []byte
		
		err := rows.Scan(
			&record.UserID,
			&record.ContentID,
			&record.StartTime,
			&record.EndTime,
			&record.Progress,
			&record.Score,
			&record.Completed,
			&record.Interactions,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描学习记录失败: %w", err)
		}
		
		// 解析JSON字段
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &record.Context)
		}
		
		records = append(records, &record)
	}
	
	return records, nil
}

func (r *RecommendationRepository) GetUserInteractions(ctx context.Context, userID string, limit int) ([]*domainServices.UserInteraction, error) {
	query := `
		SELECT id, user_id, content_id, interaction_type, value, context, created_at
		FROM user_interactions 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取用户交互失败: %w", err)
	}
	defer rows.Close()
	
	var interactions []*domainServices.UserInteraction
	for rows.Next() {
		var interaction domainServices.UserInteraction
		var contextJSON []byte
		
		err = rows.Scan(
			&interaction.UserID,
			&interaction.ContentID,
			&interaction.Interaction,
			&interaction.Duration,
			&interaction.Timestamp,
			&interaction.Rating,
			&contextJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描用户交互失败: %w", err)
		}
		
		// 解析JSON字段
		if len(contextJSON) > 0 {
			json.Unmarshal(contextJSON, &interaction.Context)
		}
		
		interactions = append(interactions, &interaction)
	}
	
	return interactions, nil
}

// 实现 EnvironmentRepository 接口

func (r *RecommendationRepository) GetEnvironmentData(ctx context.Context, userID string) (*domainServices.EnvironmentData, error) {
	query := `
		SELECT user_id, location_data, device_info, network_info, ambient_data, created_at, updated_at
		FROM environment_data 
		WHERE user_id = $1 
		ORDER BY updated_at DESC 
		LIMIT 1`
	
	var envData domainServices.EnvironmentData
	var locationJSON, weatherJSON, deviceJSON, environmentJSON []byte
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&envData.UserID,
		&locationJSON,
		&weatherJSON,
		&deviceJSON,
		&environmentJSON,
		&envData.Timestamp,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 环境数据不存�?
		}
		return nil, fmt.Errorf("获取环境数据失败: %w", err)
	}
	
	// 解析JSON字段
	if len(locationJSON) > 0 {
		json.Unmarshal(locationJSON, &envData.Location)
	}
	if len(weatherJSON) > 0 {
		json.Unmarshal(weatherJSON, &envData.Weather)
	}
	if len(deviceJSON) > 0 {
		json.Unmarshal(deviceJSON, &envData.DeviceInfo)
	}
	if len(environmentJSON) > 0 {
		json.Unmarshal(environmentJSON, &envData.Environment)
	}
	
	return &envData, nil
}

func (r *RecommendationRepository) SaveEnvironmentData(ctx context.Context, envData *domainServices.EnvironmentData) error {
	// 序列化JSON字段
	locationJSON, _ := json.Marshal(envData.Location)
	weatherJSON, _ := json.Marshal(envData.Weather)
	deviceJSON, _ := json.Marshal(envData.DeviceInfo)
	environmentJSON, _ := json.Marshal(envData.Environment)
	
	query := `
		INSERT INTO environment_data (
			user_id, location, weather, device_info, environment, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			location = EXCLUDED.location,
			weather = EXCLUDED.weather,
			device_info = EXCLUDED.device_info,
			environment = EXCLUDED.environment,
			timestamp = EXCLUDED.timestamp`
	
	_, err := r.db.ExecContext(ctx, query,
		envData.UserID,
		locationJSON,
		weatherJSON,
		deviceJSON,
		environmentJSON,
		envData.Timestamp,
	)
	
	if err != nil {
		return fmt.Errorf("保存环境数据失败: %w", err)
	}
	
	return nil
}

func (r *RecommendationRepository) GetContextRecords(ctx context.Context, userID string, limit int) ([]*domainServices.ContextRecord, error) {
	query := `
		SELECT id, user_id, context_type, context_data, created_at
		FROM context_records 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取上下文记录失�? %w", err)
	}
	defer rows.Close()
	
	var records []*domainServices.ContextRecord
	for rows.Next() {
		var record domainServices.ContextRecord
		var contextDataJSON []byte
		
		err := rows.Scan(
			&record.UserID,
			&record.Timestamp,
			&contextDataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描上下文记录失�? %w", err)
		}
		
		// 解析JSON字段
		if len(contextDataJSON) > 0 {
			json.Unmarshal(contextDataJSON, &record.Context)
		}
		
		records = append(records, &record)
	}
	
	return records, nil
}

// 行为追踪相关方法

func (r *RecommendationRepository) SaveBehaviorEvent(ctx context.Context, event *domainServices.BehaviorEvent) error {
	// 序列化JSON字段
	contextJSON, _ := json.Marshal(event.Context)
	propertiesJSON, _ := json.Marshal(event.Properties)
	
	query := `
		INSERT INTO behavior_events (
			id, learner_id, session_id, event_type, content_id, timestamp, duration, context, properties
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.LearnerID,
		event.SessionID,
		event.EventType,
		event.ContentID,
		event.Timestamp,
		event.Duration,
		contextJSON,
		propertiesJSON,
	)
	
	if err != nil {
		return fmt.Errorf("保存行为事件失败: %w", err)
	}
	
	return nil
}

func (r *RecommendationRepository) GetBehaviorEvents(ctx context.Context, userID string, limit int) ([]*domainServices.BehaviorEvent, error) {
	query := `
		SELECT id, user_id, event_type, content_id, action, value, context, metadata, created_at
		FROM behavior_events 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("获取行为事件失败: %w", err)
	}
	defer rows.Close()
	
	var events []*domainServices.BehaviorEvent
	for rows.Next() {
		var event domainServices.BehaviorEvent
		var contextJSON, propertiesJSON []byte
		
		err := rows.Scan(
			&event.ID,
			&event.LearnerID,
			&event.SessionID,
			&event.EventType,
			&event.ContentID,
			&event.Timestamp,
			&event.Duration,
			&contextJSON,
			&propertiesJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描行为事件失败: %w", err)
		}
		
		// 解析JSON字段
		if len(contextJSON) > 0 {
			json.Unmarshal(contextJSON, &event.Context)
		}
		if len(propertiesJSON) > 0 {
			json.Unmarshal(propertiesJSON, &event.Properties)
		}
		
		events = append(events, &event)
	}
	
	return events, nil
}
