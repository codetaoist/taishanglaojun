package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
)

// LearnerRepositoryImpl 学习者仓储实�?
type LearnerRepositoryImpl struct {
	db *sql.DB
}

// NewLearnerRepository 创建新的学习者仓�?
func NewLearnerRepository(db *sql.DB) repositories.LearnerRepository {
	return &LearnerRepositoryImpl{
		db: db,
	}
}

// Create 创建学习�?
func (r *LearnerRepositoryImpl) Create(ctx context.Context, learner *entities.Learner) error {
	query := `
		INSERT INTO learners (
			id, name, email, avatar_url, bio, timezone, language,
			preferences, skills, learning_goals, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	preferencesJSON, _ := json.Marshal(learner.Preferences)
	skillsJSON, _ := json.Marshal(learner.Skills)
	goalsJSON, _ := json.Marshal(learner.LearningGoals)

	_, err := r.db.ExecContext(ctx, query,
		learner.ID, learner.Name, learner.Email, learner.AvatarURL,
		learner.Bio, learner.Timezone, learner.Language,
		preferencesJSON, skillsJSON, goalsJSON,
		learner.CreatedAt, learner.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create learner: %w", err)
	}

	return nil
}

// GetLearningHistoryByContent 根据内容ID获取学习历史
func (r *LearnerRepositoryImpl) GetLearningHistoryByContent(ctx context.Context, learnerID, contentID uuid.UUID) ([]*entities.LearningHistory, error) {
	query := `
		SELECT id, learner_id, content_id, content_type, skill_name, start_time, end_time,
			   duration, progress, score, completed, interactions, timestamp, created_at
		FROM learning_history
		WHERE learner_id = $1 AND content_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning history by content: %w", err)
	}
	defer rows.Close()

	var history []*entities.LearningHistory

	for rows.Next() {
		var h entities.LearningHistory
		var interactionsJSON []byte

		err := rows.Scan(
			&h.ID, &h.LearnerID, &h.ContentID, &h.ContentType, &h.SkillName,
			&h.StartTime, &h.EndTime, &h.Duration, &h.Progress, &h.Score,
			&h.Completed, &interactionsJSON, &h.Timestamp, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan learning history row: %w", err)
		}

		// 解析JSON字段
		if len(interactionsJSON) > 0 {
			if err := json.Unmarshal(interactionsJSON, &h.Interactions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal interactions: %w", err)
			}
		}

		history = append(history, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating learning history rows: %w", err)
	}

	return history, nil
}

// GetSimilarLearners 获取相似学习�?
func (r *LearnerRepositoryImpl) GetSimilarLearners(ctx context.Context, learnerID uuid.UUID, limit int) ([]*entities.Learner, error) {
	// 基于学习偏好、技能水平和学习历史找到相似的学习�?
	query := `
		WITH target_learner AS (
			SELECT level, experience_level, weekly_goal_hours
			FROM learners
			WHERE id = $1
		),
		similar_learners AS (
			SELECT l.*, 
				   ABS(l.level - tl.level) + 
				   ABS(l.weekly_goal_hours - tl.weekly_goal_hours) as similarity_score
			FROM learners l, target_learner tl
			WHERE l.id != $1
			  AND l.experience_level = tl.experience_level
			ORDER BY similarity_score ASC
			LIMIT $2
		)
		SELECT id, user_id, name, email, avatar_url, bio, timezone, language,
			   level, experience, experience_level, weekly_goal_hours, total_study_hours,
			   created_at, updated_at
		FROM similar_learners
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar learners: %w", err)
	}
	defer rows.Close()

	var learners []*entities.Learner

	for rows.Next() {
		var learner entities.Learner

		err := rows.Scan(
			&learner.ID, &learner.UserID, &learner.Name, &learner.Email,
			&learner.AvatarURL, &learner.Bio, &learner.Timezone, &learner.Language,
			&learner.Level, &learner.Experience, &learner.ExperienceLevel,
			&learner.WeeklyGoalHours, &learner.TotalStudyHours,
			&learner.CreatedAt, &learner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan learner row: %w", err)
		}

		learners = append(learners, &learner)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating learner rows: %w", err)
	}

	return learners, nil
}

// GetSkillsByLevel 根据技能等级范围获取技�?
func (r *LearnerRepositoryImpl) GetSkillsByLevel(ctx context.Context, learnerID uuid.UUID, minLevel, maxLevel int) (map[string]*entities.SkillLevel, error) {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	skillsMap := make(map[string]*entities.SkillLevel)
	
	for _, skill := range learner.Skills {
		if skill.Level >= minLevel && skill.Level <= maxLevel {
			skillsMap[skill.SkillName] = &entities.SkillLevel{
				SkillID:     skill.SkillID,
				SkillName:   skill.SkillName,
				Level:       skill.Level,
				Experience:  skill.Experience,
				Confidence:  skill.Confidence,
				LastUpdated: skill.LastUpdated,
			}
		}
	}

	return skillsMap, nil
}

// GetLearningStreaks 获取学习连击记录
func (r *LearnerRepositoryImpl) GetLearningStreaks(ctx context.Context, learnerID uuid.UUID) ([]*entities.LearningStreak, error) {
	query := `
		SELECT current_streak, longest_streak, last_study_date, total_days
		FROM learning_streaks
		WHERE learner_id = $1
		ORDER BY last_study_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning streaks: %w", err)
	}
	defer rows.Close()

	var streaks []*entities.LearningStreak

	for rows.Next() {
		var streak entities.LearningStreak

		err := rows.Scan(
			&streak.CurrentStreak, &streak.LongestStreak, &streak.LastStudyDate, &streak.TotalDays,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan learning streak row: %w", err)
		}

		streaks = append(streaks, &streak)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating learning streak rows: %w", err)
	}

	return streaks, nil
}

// RecordLearningActivity 记录学习活动
func (r *LearnerRepositoryImpl) RecordLearningActivity(ctx context.Context, learnerID uuid.UUID, history *entities.LearningHistory) error {
	// 使用AddLearningHistory的实�?
	return r.AddLearningHistory(ctx, learnerID, history)
}

// GetLearningHistoryByDateRange 根据日期范围获取学习历史
func (r *LearnerRepositoryImpl) GetLearningHistoryByDateRange(ctx context.Context, learnerID uuid.UUID, startDate, endDate time.Time) ([]*entities.LearningHistory, error) {
	query := `
		SELECT id, learner_id, content_id, content_type, skill_name, start_time, end_time,
			   duration, progress, score, completed, interactions, timestamp, created_at
		FROM learning_history
		WHERE learner_id = $1 AND created_at >= $2 AND created_at <= $3
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning history by date range: %w", err)
	}
	defer rows.Close()

	var history []*entities.LearningHistory

	for rows.Next() {
		var h entities.LearningHistory
		var interactionsJSON []byte

		err := rows.Scan(
			&h.ID, &h.LearnerID, &h.ContentID, &h.ContentType, &h.SkillName,
			&h.StartTime, &h.EndTime, &h.Duration, &h.Progress, &h.Score,
			&h.Completed, &interactionsJSON, &h.Timestamp, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan learning history row: %w", err)
		}

		// 解析JSON字段
		if len(interactionsJSON) > 0 {
			if err := json.Unmarshal(interactionsJSON, &h.Interactions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal interactions: %w", err)
			}
		}

		history = append(history, &h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating learning history rows: %w", err)
	}

	return history, nil
}

// GetLearnerSkills 获取学习者的技能水�?
func (r *LearnerRepositoryImpl) GetLearnerSkills(ctx context.Context, learnerID uuid.UUID) (map[string]*entities.SkillLevel, error) {
	query := `
		SELECT skills
		FROM learners
		WHERE id = $1
	`

	var skillsJSON []byte
	err := r.db.QueryRowContext(ctx, query, learnerID).Scan(&skillsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return make(map[string]*entities.SkillLevel), nil
		}
		return nil, fmt.Errorf("failed to get learner skills: %w", err)
	}

	var skills map[string]*entities.SkillLevel
	if len(skillsJSON) > 0 {
		if err := json.Unmarshal(skillsJSON, &skills); err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
		}
	}

	if skills == nil {
		skills = make(map[string]*entities.SkillLevel)
	}

	return skills, nil
}

// GetCurrentStreak 获取当前学习连续�?
func (r *LearnerRepositoryImpl) GetCurrentStreak(ctx context.Context, learnerID uuid.UUID, streakType string) (*entities.LearningStreak, error) {
	query := `
		SELECT 
			current_streak,
			longest_streak,
			last_study_date,
			total_days
		FROM learning_streaks
		WHERE learner_id = $1 AND streak_type = $2
	`

	streak := &entities.LearningStreak{}
	err := r.db.QueryRowContext(ctx, query, learnerID, streakType).Scan(
		&streak.CurrentStreak,
		&streak.LongestStreak,
		&streak.LastStudyDate,
		&streak.TotalDays,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有记录，返回默认的空连续性记�?
			return &entities.LearningStreak{
				CurrentStreak: 0,
				LongestStreak: 0,
				LastStudyDate: time.Time{},
				TotalDays:     0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get current streak: %w", err)
	}

	return streak, nil
}

// UpdateLearningStreak 更新学习连续�?
func (r *LearnerRepositoryImpl) UpdateLearningStreak(ctx context.Context, learnerID uuid.UUID, streak *entities.LearningStreak) error {
	// 首先检查是否已存在记录
	checkQuery := `
		SELECT COUNT(*) FROM learning_streaks 
		WHERE learner_id = $1
	`
	
	var count int
	err := r.db.QueryRowContext(ctx, checkQuery, learnerID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing streak: %w", err)
	}

	if count > 0 {
		// 更新现有记录
		updateQuery := `
			UPDATE learning_streaks 
			SET current_streak = $2, longest_streak = $3, last_study_date = $4, total_days = $5, updated_at = $6
			WHERE learner_id = $1
		`
		
		_, err = r.db.ExecContext(ctx, updateQuery,
			learnerID, streak.CurrentStreak, streak.LongestStreak, 
			streak.LastStudyDate, streak.TotalDays, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to update learning streak: %w", err)
		}
	} else {
		// 插入新记�?
		insertQuery := `
			INSERT INTO learning_streaks (learner_id, current_streak, longest_streak, last_study_date, total_days, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		
		now := time.Now()
		_, err = r.db.ExecContext(ctx, insertQuery,
			learnerID, streak.CurrentStreak, streak.LongestStreak,
			streak.LastStudyDate, streak.TotalDays, now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert learning streak: %w", err)
		}
	}

	return nil
}

// GetByID 根据ID获取学习�?
func (r *LearnerRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Learner, error) {
	query := `
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	learner := &entities.Learner{}
	var preferencesJSON, skillsJSON, goalsJSON []byte

	err := row.Scan(
		&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
		&learner.Bio, &learner.Timezone, &learner.Language,
		&preferencesJSON, &skillsJSON, &goalsJSON,
		&learner.CreatedAt, &learner.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("learner not found")
		}
		return nil, fmt.Errorf("failed to get learner: %w", err)
	}

	// 解析JSON字段
	if len(preferencesJSON) > 0 {
		json.Unmarshal(preferencesJSON, &learner.Preferences)
	}
	if len(skillsJSON) > 0 {
		json.Unmarshal(skillsJSON, &learner.Skills)
	}
	if len(goalsJSON) > 0 {
		json.Unmarshal(goalsJSON, &learner.LearningGoals)
	}

	return learner, nil
}

// GetByEmail 根据邮箱获取学习�?
func (r *LearnerRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entities.Learner, error) {
	query := `
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email)

	learner := &entities.Learner{}
	var preferencesJSON, skillsJSON, goalsJSON []byte

	err := row.Scan(
		&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
		&learner.Bio, &learner.Timezone, &learner.Language,
		&preferencesJSON, &skillsJSON, &goalsJSON,
		&learner.CreatedAt, &learner.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("learner not found")
		}
		return nil, fmt.Errorf("failed to get learner by email: %w", err)
	}

	// 解析JSON字段
	if len(preferencesJSON) > 0 {
		json.Unmarshal(preferencesJSON, &learner.Preferences)
	}
	if len(skillsJSON) > 0 {
		json.Unmarshal(skillsJSON, &learner.Skills)
	}
	if len(goalsJSON) > 0 {
		json.Unmarshal(goalsJSON, &learner.LearningGoals)
	}

	return learner, nil
}

// GetByUserID 根据用户ID获取学习�?
func (r *LearnerRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Learner, error) {
	query := `
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners WHERE user_id = $1
	`

	row := r.db.QueryRowContext(ctx, query, userID)

	learner := &entities.Learner{}
	var preferencesJSON, skillsJSON, goalsJSON []byte

	err := row.Scan(
		&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
		&learner.Bio, &learner.Timezone, &learner.Language,
		&preferencesJSON, &skillsJSON, &goalsJSON,
		&learner.CreatedAt, &learner.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("learner not found")
		}
		return nil, fmt.Errorf("failed to get learner by user ID: %w", err)
	}

	// 解析JSON字段
	if len(preferencesJSON) > 0 {
		json.Unmarshal(preferencesJSON, &learner.Preferences)
	}
	if len(skillsJSON) > 0 {
		json.Unmarshal(skillsJSON, &learner.Skills)
	}
	if len(goalsJSON) > 0 {
		json.Unmarshal(goalsJSON, &learner.LearningGoals)
	}

	return learner, nil
}

// Update 更新学习�?
func (r *LearnerRepositoryImpl) Update(ctx context.Context, learner *entities.Learner) error {
	query := `
		UPDATE learners SET
			name = $2, email = $3, avatar_url = $4, bio = $5,
			timezone = $6, language = $7, preferences = $8,
			skills = $9, learning_goals = $10, updated_at = $11
		WHERE id = $1
	`

	preferencesJSON, _ := json.Marshal(learner.Preferences)
	skillsJSON, _ := json.Marshal(learner.Skills)
	goalsJSON, _ := json.Marshal(learner.LearningGoals)

	learner.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		learner.ID, learner.Name, learner.Email, learner.AvatarURL,
		learner.Bio, learner.Timezone, learner.Language,
		preferencesJSON, skillsJSON, goalsJSON, learner.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update learner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("learner not found")
	}

	return nil
}

// Delete 删除学习�?
func (r *LearnerRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM learners WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete learner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("learner not found")
	}

	return nil
}

// List 获取学习者列�?
func (r *LearnerRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Learner, error) {
	query := `
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list learners: %w", err)
	}
	defer rows.Close()

	var learners []*entities.Learner

	for rows.Next() {
		learner := &entities.Learner{}
		var preferencesJSON, skillsJSON, goalsJSON []byte

		err := rows.Scan(
			&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
			&learner.Bio, &learner.Timezone, &learner.Language,
			&preferencesJSON, &skillsJSON, &goalsJSON,
			&learner.CreatedAt, &learner.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learner: %w", err)
		}

		// 解析JSON字段
		if len(preferencesJSON) > 0 {
			json.Unmarshal(preferencesJSON, &learner.Preferences)
		}
		if len(skillsJSON) > 0 {
			json.Unmarshal(skillsJSON, &learner.Skills)
		}
		if len(goalsJSON) > 0 {
			json.Unmarshal(goalsJSON, &learner.LearningGoals)
		}

		learners = append(learners, learner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learners: %w", err)
	}

	return learners, nil
}

// ListByLearningStyle 根据学习风格列出学习�?
func (r *LearnerRepositoryImpl) ListByLearningStyle(ctx context.Context, style entities.LearningStyle, offset, limit int) ([]*entities.Learner, error) {
	query := `
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners
		WHERE JSON_EXTRACT(preferences, '$.style') = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(style), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list learners by learning style: %w", err)
	}
	defer rows.Close()

	var learners []*entities.Learner

	for rows.Next() {
		learner := &entities.Learner{}
		var preferencesJSON, skillsJSON, goalsJSON []byte

		err := rows.Scan(
			&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
			&learner.Bio, &learner.Timezone, &learner.Language,
			&preferencesJSON, &skillsJSON, &goalsJSON,
			&learner.CreatedAt, &learner.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learner: %w", err)
		}

		// 解析JSON字段
		if len(preferencesJSON) > 0 {
			json.Unmarshal(preferencesJSON, &learner.Preferences)
		}
		if len(skillsJSON) > 0 {
			json.Unmarshal(skillsJSON, &learner.Skills)
		}
		if len(goalsJSON) > 0 {
			json.Unmarshal(goalsJSON, &learner.LearningGoals)
		}

		learners = append(learners, learner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learners: %w", err)
	}

	return learners, nil
}

// ListByLearningPace 根据学习节奏列出学习�?
func (r *LearnerRepositoryImpl) ListByLearningPace(ctx context.Context, pace entities.LearningPace, offset, limit int) ([]*entities.Learner, error) {
	query := `
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners
		WHERE JSON_EXTRACT(preferences, '$.pace') = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(pace), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list learners by learning pace: %w", err)
	}
	defer rows.Close()

	var learners []*entities.Learner

	for rows.Next() {
		learner := &entities.Learner{}
		var preferencesJSON, skillsJSON, goalsJSON []byte

		err := rows.Scan(
			&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
			&learner.Bio, &learner.Timezone, &learner.Language,
			&preferencesJSON, &skillsJSON, &goalsJSON,
			&learner.CreatedAt, &learner.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learner: %w", err)
		}

		// 解析JSON字段
		if len(preferencesJSON) > 0 {
			json.Unmarshal(preferencesJSON, &learner.Preferences)
		}
		if len(skillsJSON) > 0 {
			json.Unmarshal(skillsJSON, &learner.Skills)
		}
		if len(goalsJSON) > 0 {
			json.Unmarshal(goalsJSON, &learner.LearningGoals)
		}

		learners = append(learners, learner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learners: %w", err)
	}

	return learners, nil
}

// AddLearnerGoal 添加学习目标
func (r *LearnerRepositoryImpl) AddLearnerGoal(ctx context.Context, learnerID uuid.UUID, goal *entities.LearningGoal) error {
	// 首先获取当前学习�?
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 添加新目�?
	learner.LearningGoals = append(learner.LearningGoals, *goal)

	// 更新学习�?
	return r.Update(ctx, learner)
}

// UpdateLearnerGoal 更新学习目标
func (r *LearnerRepositoryImpl) UpdateLearnerGoal(ctx context.Context, learnerID uuid.UUID, goal *entities.LearningGoal) error {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 查找并更新目�?
	for i, g := range learner.LearningGoals {
		if g.ID == goal.ID {
			learner.LearningGoals[i] = *goal
			break
		}
	}

	return r.Update(ctx, learner)
}

// RemoveLearnerGoal 移除学习目标
func (r *LearnerRepositoryImpl) RemoveLearnerGoal(ctx context.Context, learnerID, goalID uuid.UUID) error {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 移除目标
	for i, g := range learner.LearningGoals {
		if g.ID == goalID {
			learner.LearningGoals = append(learner.LearningGoals[:i], learner.LearningGoals[i+1:]...)
			break
		}
	}

	return r.Update(ctx, learner)
}

// GetLearnerGoals 获取学习目标
func (r *LearnerRepositoryImpl) GetLearnerGoals(ctx context.Context, learnerID uuid.UUID) ([]*entities.LearningGoal, error) {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	// 转换为指针切�?
	goals := make([]*entities.LearningGoal, len(learner.LearningGoals))
	for i := range learner.LearningGoals {
		goals[i] = &learner.LearningGoals[i]
	}

	return goals, nil
}

// UpdateSkill 更新技�?
func (r *LearnerRepositoryImpl) UpdateSkill(ctx context.Context, learnerID uuid.UUID, skill *entities.Skill) error {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 查找并更新技�?
	found := false
	for i, s := range learner.Skills {
		if s.SkillID == skill.ID {
			// 将Skill转换为SkillLevel
			skillLevel := entities.SkillLevel{
				SkillID:     skill.ID,
				SkillName:   skill.Name,
				Level:       skill.Level,
				Experience:  0, // 默认�?
				Confidence:  0.5, // 默认�?
				LastUpdated: time.Now(),
			}
			learner.Skills[i] = skillLevel
			found = true
			break
		}
	}

	// 如果没找到，添加新技�?
	if !found {
		skillLevel := entities.SkillLevel{
			SkillID:     skill.ID,
			SkillName:   skill.Name,
			Level:       skill.Level,
			Experience:  0, // 默认�?
			Confidence:  0.5, // 默认�?
			LastUpdated: time.Now(),
		}
		learner.Skills = append(learner.Skills, skillLevel)
	}

	return r.Update(ctx, learner)
}

// GetSkills 获取技能列�?
func (r *LearnerRepositoryImpl) GetSkills(ctx context.Context, learnerID uuid.UUID) ([]*entities.Skill, error) {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return nil, err
	}

	// 转换SkillLevel为Skill
	skills := make([]*entities.Skill, len(learner.Skills))
	for i, skillLevel := range learner.Skills {
		skills[i] = &entities.Skill{
			ID:          skillLevel.SkillID,
			Name:        skillLevel.SkillName,
			Level:       skillLevel.Level,
			Category:    "", // 默认�?
			Description: "", // 默认�?
			AcquiredAt:  skillLevel.LastUpdated,
			UpdatedAt:   skillLevel.LastUpdated,
		}
	}

	return skills, nil
}

// AddOrUpdateSkill 添加或更新技�?
func (r *LearnerRepositoryImpl) AddOrUpdateSkill(ctx context.Context, learnerID uuid.UUID, skill string, level *entities.SkillLevel) error {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 查找并更新技�?
	found := false
	for i, s := range learner.Skills {
		if s.SkillName == skill {
			learner.Skills[i] = *level
			found = true
			break
		}
	}

	// 如果没找到，添加新技�?
	if !found {
		learner.Skills = append(learner.Skills, *level)
	}

	return r.Update(ctx, learner)
}

// RemoveSkill 移除技�?
func (r *LearnerRepositoryImpl) RemoveSkill(ctx context.Context, learnerID uuid.UUID, skill string) error {
	learner, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return err
	}

	// 查找并移除技�?
	for i, s := range learner.Skills {
		if s.SkillName == skill {
			learner.Skills = append(learner.Skills[:i], learner.Skills[i+1:]...)
			break
		}
	}

	return r.Update(ctx, learner)
}



// AddLearningHistory 添加学习历史记录
func (r *LearnerRepositoryImpl) AddLearningHistory(ctx context.Context, learnerID uuid.UUID, history *entities.LearningHistory) error {
	query := `
		INSERT INTO learning_history (
			id, learner_id, content_id, content_type, start_time, end_time,
			duration, progress, score, completed, interactions, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.LearnerID, history.ContentID, history.ContentType,
		history.StartTime, history.EndTime, history.Duration, history.Progress,
		history.Score, history.Completed, history.Interactions, history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add learning history: %w", err)
	}

	return nil
}

// GetLearningHistory 获取学习历史
func (r *LearnerRepositoryImpl) GetLearningHistory(ctx context.Context, learnerID uuid.UUID, limit int) ([]*entities.LearningHistory, error) {
	query := `
		SELECT id, learner_id, content_id, content_type, start_time, end_time,
			   duration, progress, score, completed, interactions, created_at
		FROM learning_history
		WHERE learner_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning history: %w", err)
	}
	defer rows.Close()

	var history []*entities.LearningHistory

	for rows.Next() {
		h := &entities.LearningHistory{}

		err := rows.Scan(
			&h.ID, &h.LearnerID, &h.ContentID, &h.ContentType,
			&h.StartTime, &h.EndTime, &h.Duration, &h.Progress,
			&h.Score, &h.Completed, &h.Interactions, &h.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan learning history: %w", err)
		}

		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate learning history: %w", err)
	}

	return history, nil
}

// GetLearningStreak 获取学习连续天数
func (r *LearnerRepositoryImpl) GetLearningStreak(ctx context.Context, learnerID uuid.UUID) (int, error) {
	query := `
		WITH daily_activity AS (
			SELECT DISTINCT DATE(timestamp) as activity_date
			FROM learning_history
			WHERE learner_id = $1
			ORDER BY activity_date DESC
		),
		streak_calc AS (
			SELECT activity_date,
				   ROW_NUMBER() OVER (ORDER BY activity_date DESC) as rn,
				   activity_date - INTERVAL '1 day' * (ROW_NUMBER() OVER (ORDER BY activity_date DESC) - 1) as streak_date
			FROM daily_activity
		)
		SELECT COUNT(*) as streak
		FROM streak_calc
		WHERE streak_date = (SELECT MAX(streak_date) FROM streak_calc)
	`

	var streak int
	err := r.db.QueryRowContext(ctx, query, learnerID).Scan(&streak)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get learning streak: %w", err)
	}

	return streak, nil
}

// GetLearnerStatistics 获取学习者统�?
func (r *LearnerRepositoryImpl) GetLearnerStatistics(ctx context.Context, learnerID uuid.UUID) (*repositories.LearnerStatistics, error) {
	// 获取基本统计
	query := `
		SELECT 
			COALESCE(SUM(duration), 0) as total_learning_time,
			COUNT(DISTINCT content_id) as completed_content,
			0 as active_goals,
			0 as completed_goals,
			0 as skill_count,
			0.0 as average_skill_level,
			0 as longest_streak,
			0.0 as learning_frequency
		FROM learning_history
		WHERE learner_id = $1 AND completed = true
	`

	stats := &repositories.LearnerStatistics{}
	var totalLearningTimeSeconds int64
	err := r.db.QueryRowContext(ctx, query, learnerID).Scan(
		&totalLearningTimeSeconds,
		&stats.CompletedContent,
		&stats.ActiveGoals,
		&stats.CompletedGoals,
		&stats.SkillCount,
		&stats.AverageSkillLevel,
		&stats.LongestStreak,
		&stats.LearningFrequency,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get learner statistics: %w", err)
	}

	// 转换秒为时间�?
	stats.TotalLearningTime = time.Duration(totalLearningTimeSeconds) * time.Second

	// 初始化映射字�?
	stats.CurrentStreaks = make(map[string]int)
	stats.SkillDistribution = make(map[string]int)
	stats.ContentTypeProgress = make(map[string]float64)
	stats.PreferredTimeSlots = []entities.TimeSlot{}

	return stats, nil
}

// GetWeeklyProgress 获取周进�?
func (r *LearnerRepositoryImpl) GetWeeklyProgress(ctx context.Context, learnerID uuid.UUID, weekStart time.Time) (*repositories.WeeklyProgress, error) {
	// 计算周结束时�?
	weekEnd := weekStart.AddDate(0, 0, 6)
	
	query := `
		SELECT 
			COALESCE(SUM(duration), 0) as total_learning_time,
			COUNT(DISTINCT DATE(created_at)) as days_active,
			COUNT(CASE WHEN completed = true THEN 1 END) as content_completed
		FROM learning_history
		WHERE learner_id = $1
		  AND created_at >= $2
		  AND created_at <= $3
	`

	row := r.db.QueryRowContext(ctx, query, learnerID, weekStart, weekEnd)
	
	progress := &repositories.WeeklyProgress{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
	}
	
	var totalLearningTimeSeconds int64

	err := row.Scan(
		&totalLearningTimeSeconds,
		&progress.DaysActive,
		&progress.ContentCompleted,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan weekly progress: %w", err)
	}

	// Convert seconds to time.Duration
	progress.TotalLearningTime = time.Duration(totalLearningTimeSeconds) * time.Second

	// Initialize slices and maps
	progress.SkillsImproved = make([]string, 0)
	progress.GoalsProgress = make(map[uuid.UUID]float64)
	progress.DailyProgress = make(map[time.Time]repositories.DayProgress)

	return progress, nil
}

// GetLearningTrends 获取学习趋势
func (r *LearnerRepositoryImpl) GetLearningTrends(ctx context.Context, learnerID uuid.UUID, days int) (*repositories.LearningTrends, error) {
	// 使用传入的天数作为周�?
	period := days
	if period <= 0 {
		period = 30 // 默认30�?
	}

	// 创建基本的学习趋势对�?
	trends := &repositories.LearningTrends{
		Period: period,
		LearningTimetrend: repositories.TimeTrend{
			DailyAverage:    0,
			WeeklyAverage:   0,
			Trend:           "stable",
			TrendPercentage: 0,
			DailyData:       make([]time.Duration, 0),
		},
		ContentCompletionTrend: repositories.CompletionTrend{
			DailyAverage:    0,
			WeeklyAverage:   0,
			Trend:           "stable",
			TrendPercentage: 0,
			DailyData:       make([]float64, 0),
		},
		SkillProgressTrend: make(map[string]repositories.SkillTrend),
		EngagementTrend: repositories.EngagementTrend{
			DailyEngagement:     make([]float64, 0),
			AverageEngagement:   0,
			Trend:               "stable",
			TrendPercentage:     0,
			PeakEngagementTime:  time.Now(),
		},
		Predictions: repositories.TrendPredictions{
			NextWeekLearningTime:    0,
			NextWeekCompletion:      0,
			GoalCompletionDates:     make(map[uuid.UUID]time.Time),
			SkillLevelUpDates:       make(map[string]time.Time),
			RiskFactors:             make([]string, 0),
			Recommendations:         make([]string, 0),
		},
	}

	return trends, nil
}

// SearchLearners 搜索学习�?
func (r *LearnerRepositoryImpl) SearchLearners(ctx context.Context, query *repositories.LearnerSearchQuery) ([]*entities.Learner, int, error) {
	// 构建WHERE条件
	whereConditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if query.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+query.Query+"%")
		argIndex++
	}

	if query.Timezone != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("timezone = $%d", argIndex))
		args = append(args, query.Timezone)
		argIndex++
	}

	if query.Language != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("language = $%d", argIndex))
		args = append(args, query.Language)
		argIndex++
	}

	whereClause := fmt.Sprintf("WHERE %s", fmt.Sprintf("%s", whereConditions[0]))
	for i := 1; i < len(whereConditions); i++ {
		whereClause += " AND " + whereConditions[i]
	}

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM learners
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count learners: %w", err)
	}

	// 获取数据
	dataQuery := fmt.Sprintf(`
		SELECT id, name, email, avatar_url, bio, timezone, language,
			   preferences, skills, learning_goals, created_at, updated_at
		FROM learners
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, query.SortBy, query.SortOrder, argIndex, argIndex+1)

	args = append(args, query.Limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search learners: %w", err)
	}
	defer rows.Close()

	var learners []*entities.Learner

	for rows.Next() {
		learner := &entities.Learner{}
		var preferencesJSON, skillsJSON, goalsJSON []byte

		err := rows.Scan(
			&learner.ID, &learner.Name, &learner.Email, &learner.AvatarURL,
			&learner.Bio, &learner.Timezone, &learner.Language,
			&preferencesJSON, &skillsJSON, &goalsJSON,
			&learner.CreatedAt, &learner.UpdatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan learner: %w", err)
		}

		// 解析JSON字段
		if len(preferencesJSON) > 0 {
			json.Unmarshal(preferencesJSON, &learner.Preferences)
		}
		if len(skillsJSON) > 0 {
			json.Unmarshal(skillsJSON, &learner.Skills)
		}
		if len(goalsJSON) > 0 {
			json.Unmarshal(goalsJSON, &learner.LearningGoals)
		}

		learners = append(learners, learner)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate learners: %w", err)
	}

	return learners, total, nil
}

// BatchCreate 批量创建学习�?
func (r *LearnerRepositoryImpl) BatchCreate(ctx context.Context, learners []*entities.Learner) error {
	if len(learners) == 0 {
		return nil
	}

	query := `
		INSERT INTO learners (
			id, name, email, avatar_url, bio, timezone, language,
			preferences, skills, learning_goals, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, learner := range learners {
		preferencesJSON, _ := json.Marshal(learner.Preferences)
		skillsJSON, _ := json.Marshal(learner.Skills)
		goalsJSON, _ := json.Marshal(learner.LearningGoals)

		_, err := stmt.ExecContext(ctx,
			learner.ID, learner.Name, learner.Email, learner.AvatarURL,
			learner.Bio, learner.Timezone, learner.Language,
			preferencesJSON, skillsJSON, goalsJSON,
			learner.CreatedAt, learner.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to insert learner %s: %w", learner.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLearnerPreferences 获取学习者偏好设�?
func (r *LearnerRepositoryImpl) GetLearnerPreferences(ctx context.Context, learnerID uuid.UUID) (*entities.LearningPreference, error) {
	query := `
		SELECT preferences
		FROM learners 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, learnerID)

	var preferencesJSON []byte
	err := row.Scan(&preferencesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("learner not found")
		}
		return nil, fmt.Errorf("failed to get learner preferences: %w", err)
	}

	if len(preferencesJSON) == 0 {
		// 返回默认偏好设置
		return &entities.LearningPreference{
			Style:     entities.LearningStyleVisual,
			Pace:      entities.LearningPaceMedium,
			SessionDuration:     45,
			BreakDuration:       15,
			DifficultyTolerance: 0.7,
			InteractiveContent:  true,
			MultimediaContent:   true,
		}, nil
	}

	var preferences entities.LearningPreference
	if err := json.Unmarshal(preferencesJSON, &preferences); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences: %w", err)
	}

	return &preferences, nil
}

// UpdateLearnerPreferences 更新学习者偏好设�?
func (r *LearnerRepositoryImpl) UpdateLearnerPreferences(ctx context.Context, learnerID uuid.UUID, preferences *entities.LearningPreference) error {
	// 首先检查学习者是否存�?
	_, err := r.GetByID(ctx, learnerID)
	if err != nil {
		return fmt.Errorf("learner not found: %w", err)
	}

	// 序列化偏好设�?
	preferencesJSON, err := json.Marshal(preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	// 更新偏好设置
	query := `
		UPDATE learners 
		SET preferences = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, learnerID, preferencesJSON, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update learner preferences: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("learner not found")
	}

	return nil
}

// GetActiveGoals 获取学习者的活跃目标
func (r *LearnerRepositoryImpl) GetActiveGoals(ctx context.Context, learnerID uuid.UUID) ([]*entities.LearningGoal, error) {
	query := `
		SELECT id, title, description, target_date, priority, achieved, created_at, updated_at
		FROM learning_goals
		WHERE learner_id = $1 AND achieved = false
		ORDER BY priority DESC, target_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, learnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active goals: %w", err)
	}
	defer rows.Close()

	var goals []*entities.LearningGoal
	for rows.Next() {
		goal := &entities.LearningGoal{}
		err := rows.Scan(
			&goal.ID, &goal.Title, &goal.Description, &goal.TargetDate,
			&goal.Priority, &goal.Achieved, &goal.CreatedAt, &goal.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		goals = append(goals, goal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate goals: %w", err)
	}

	return goals, nil
}

// BatchUpdate 批量更新学习�?
func (r *LearnerRepositoryImpl) BatchUpdate(ctx context.Context, learners []*entities.Learner) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE learners SET
			name = $2, email = $3, avatar_url = $4, bio = $5,
			timezone = $6, language = $7, preferences = $8,
			skills = $9, learning_goals = $10, updated_at = $11
		WHERE id = $1
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, learner := range learners {
		preferencesJSON, _ := json.Marshal(learner.Preferences)
		skillsJSON, _ := json.Marshal(learner.Skills)
		goalsJSON, _ := json.Marshal(learner.LearningGoals)

		_, err := stmt.ExecContext(ctx,
			learner.ID, learner.Name, learner.Email, learner.AvatarURL,
			learner.Bio, learner.Timezone, learner.Language,
			preferencesJSON, skillsJSON, goalsJSON, learner.UpdatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to update learner %s: %w", learner.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
