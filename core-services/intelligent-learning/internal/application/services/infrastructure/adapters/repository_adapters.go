package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/entities"
	"github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/repositories"
	domainservices "github.com/codetaoist/taishanglaojun/core-services/intelligent-learning/internal/domain/services"
)

// ContentRepositoryAdapter 内容仓库适配�?
type ContentRepositoryAdapter struct {
	contentRepo repositories.LearningContentRepository
}

// NewContentRepositoryAdapter 创建内容仓库适配�?
func NewContentRepositoryAdapter(contentRepo repositories.LearningContentRepository) domainservices.ContentRepository {
	return &ContentRepositoryAdapter{
		contentRepo: contentRepo,
	}
}

// GetContentByID 根据ID获取内容
func (a *ContentRepositoryAdapter) GetContentByID(ctx context.Context, contentID string) (*domainservices.Content, error) {
	id, err := uuid.Parse(contentID)
	if err != nil {
		return nil, fmt.Errorf("invalid content ID: %w", err)
	}

	content, err := a.contentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domainservices.Content{
		ID:          content.ID.String(),
		Title:       content.Title,
		Description: content.Description,
		Category:    string(content.Type),
		Difficulty:  string(content.Difficulty),
		Tags:        content.Tags,
	}, nil
}

// GetContentsByCategory 根据分类获取内容
func (a *ContentRepositoryAdapter) GetContentsByCategory(ctx context.Context, category string) ([]*domainservices.Content, error) {
	contents, err := a.contentRepo.GetByType(ctx, entities.ContentType(category), 0, 100)
	if err != nil {
		return nil, err
	}

	var result []*domainservices.Content
	for _, content := range contents {
		result = append(result, &domainservices.Content{
			ID:          content.ID.String(),
			Title:       content.Title,
			Description: content.Description,
			Category:    string(content.Type),
			Difficulty:  string(content.Difficulty),
			Tags:        content.Tags,
		})
	}

	return result, nil
}

// GetContentsByDifficulty 根据难度获取内容
func (a *ContentRepositoryAdapter) GetContentsByDifficulty(ctx context.Context, difficulty string) ([]*domainservices.Content, error) {
	// 将字符串转换为DifficultyLevel
	var diffLevel entities.DifficultyLevel
	switch difficulty {
	case "1", "beginner":
		diffLevel = entities.DifficultyBeginner
	case "2", "elementary":
		diffLevel = entities.DifficultyElementary
	case "3", "intermediate":
		diffLevel = entities.DifficultyIntermediate
	case "4", "advanced":
		diffLevel = entities.DifficultyAdvanced
	case "5", "expert":
		diffLevel = entities.DifficultyExpert
	default:
		diffLevel = entities.DifficultyBeginner
	}
	contents, err := a.contentRepo.GetByDifficulty(ctx, diffLevel, diffLevel, 0, 100)
	if err != nil {
		return nil, err
	}

	var result []*domainservices.Content
	for _, content := range contents {
		result = append(result, &domainservices.Content{
			ID:          content.ID.String(),
			Title:       content.Title,
			Description: content.Description,
			Category:    string(content.Type),
			Difficulty:  string(content.Difficulty),
			Tags:        content.Tags,
		})
	}

	return result, nil
}

// GetContentsByTags 根据标签获取内容
func (a *ContentRepositoryAdapter) GetContentsByTags(ctx context.Context, tags []string) ([]*domainservices.Content, error) {
	contents, err := a.contentRepo.GetByTags(ctx, tags, 0, 100)
	if err != nil {
		return nil, err
	}

	var result []*domainservices.Content
	for _, content := range contents {
		result = append(result, &domainservices.Content{
			ID:          content.ID.String(),
			Title:       content.Title,
			Description: content.Description,
			Category:    string(content.Type),
			Difficulty:  string(content.Difficulty),
			Tags:        content.Tags,
		})
	}

	return result, nil
}

// UserRepositoryAdapter 用户仓库适配�?
type UserRepositoryAdapter struct {
	learnerRepo repositories.LearnerRepository
}

// NewUserRepositoryAdapter 创建用户仓库适配�?
func NewUserRepositoryAdapter(learnerRepo repositories.LearnerRepository) domainservices.UserRepository {
	return &UserRepositoryAdapter{
		learnerRepo: learnerRepo,
	}
}

// GetUserProfile 获取用户画像
func (a *UserRepositoryAdapter) GetUserProfile(ctx context.Context, userID string) (*domainservices.UserProfile, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	learner, err := a.learnerRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// 从技能等级构建技能映�?
	skillLevels := make(map[string]float64)
	for _, skill := range learner.Skills {
		skillLevels[skill.SkillName] = float64(skill.Level)
	}

	return &domainservices.UserProfile{
		UserID:        learner.UserID,
		Preferences:   make(map[string]float64),
		Categories:    make(map[string]float64),
		Tags:          make(map[string]float64),
		Keywords:      make(map[string]float64),
		SkillLevels:   skillLevels,
		LearningStyle: learner.LearningStyle,
		Difficulty:    0.5, // 默认�?
		Duration:      int64(learner.WeeklyGoalHours * 3600),
		UpdatedAt:     learner.UpdatedAt,
	}, nil
}

// GetUserLearningHistory 获取用户学习历史
func (a *UserRepositoryAdapter) GetUserLearningHistory(ctx context.Context, userID string) ([]*domainservices.LearningRecord, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	learner, err := a.learnerRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var records []*domainservices.LearningRecord
	for _, history := range learner.LearningHistory {
		score := 0.0
		if history.Score != nil {
			score = *history.Score
		}
		endTime := time.Time{}
		if history.EndTime != nil {
			endTime = *history.EndTime
		}
		records = append(records, &domainservices.LearningRecord{
			UserID:    userID,
			ContentID: history.ContentID.String(),
			StartTime: history.StartTime,
			EndTime:   endTime,
			Progress:  history.Progress,
			Score:     score,
			Completed: history.Completed,
		})
	}

	return records, nil
}

// GetUserInteractions 获取用户交互记录
func (a *UserRepositoryAdapter) GetUserInteractions(ctx context.Context, userID string, limit int) ([]*domainservices.UserInteraction, error) {
	// 这里可以从学习历史中提取交互信息
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	learner, err := a.learnerRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var interactions []*domainservices.UserInteraction
	count := 0
	for _, history := range learner.LearningHistory {
		if limit > 0 && count >= limit {
			break
		}
		interactions = append(interactions, &domainservices.UserInteraction{
			UserID:      learner.UserID,
			ContentID:   history.ContentID,
			Interaction: "view",
			Duration:    int64(history.Duration.Seconds()),
			Timestamp:   history.StartTime,
			Rating:      0.0, // 默认评分
		})
		count++
	}

	return interactions, nil
}
