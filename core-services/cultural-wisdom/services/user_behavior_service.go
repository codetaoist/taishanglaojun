﻿package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// UserBehaviorService 
type UserBehaviorService struct {
	db     *gorm.DB
	cache  *CacheService
	logger *zap.Logger
}

// NewUserBehaviorService 
func NewUserBehaviorService(db *gorm.DB, cache *CacheService, logger *zap.Logger) *UserBehaviorService {
	return &UserBehaviorService{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}

// BehaviorRequest 
type BehaviorRequest struct {
	UserID     string            `json:"user_id"`
	WisdomID   string            `json:"wisdom_id"`
	ActionType string            `json:"action_type"`
	Duration   int64             `json:"duration,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	IPAddress  string            `json:"ip_address,omitempty"`
	UserAgent  string            `json:"user_agent,omitempty"`
}

// UserProfile 
type UserProfile struct {
	UserID           string             `json:"user_id"`
	PreferredCategories []CategoryScore `json:"preferred_categories"`
	PreferredSchools    []SchoolScore   `json:"preferred_schools"`
	PreferredAuthors    []AuthorScore   `json:"preferred_authors"`
	PreferredTags       []TagScore      `json:"preferred_tags"`
	ReadingSpeed        float64         `json:"reading_speed"`
	ActiveHours         []int           `json:"active_hours"`
	LastActive          time.Time       `json:"last_active"`
	TotalActions        int64           `json:"total_actions"`
	EngagementScore     float64         `json:"engagement_score"`
}

// CategoryScore 
type CategoryScore struct {
	Category string  `json:"category"`
	Score    float64 `json:"score"`
	Count    int     `json:"count"`
}

// SchoolScore 
type SchoolScore struct {
	School string  `json:"school"`
	Score  float64 `json:"score"`
	Count  int     `json:"count"`
}

// AuthorScore ?
type AuthorScore struct {
	Author string  `json:"author"`
	Score  float64 `json:"score"`
	Count  int     `json:"count"`
}

// TagScore 
type TagScore struct {
	Tag   string  `json:"tag"`
	Score float64 `json:"score"`
	Count int     `json:"count"`
}

// RecordBehavior 
func (s *UserBehaviorService) RecordBehavior(ctx context.Context, req BehaviorRequest) error {
	// ID
	behaviorID := fmt.Sprintf("%s_%s_%s_%d", req.UserID, req.WisdomID, req.ActionType, time.Now().UnixNano())

	// 
	score := s.calculateActionScore(req.ActionType, req.Duration)

	// 
	contextJSON := ""
	if req.Context != nil {
		if data, err := json.Marshal(req.Context); err == nil {
			contextJSON = string(data)
		}
	}

	// 
	behavior := models.UserBehavior{
		ID:         behaviorID,
		UserID:     req.UserID,
		WisdomID:   req.WisdomID,
		ActionType: req.ActionType,
		Duration:   req.Duration,
		Score:      score,
		Context:    contextJSON,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	// 浽
	if err := s.db.WithContext(ctx).Create(&behavior).Error; err != nil {
		return fmt.Errorf("failed to record behavior: %w", err)
	}

	// 
	go s.updateUserPreference(req.UserID, req.WisdomID, req.ActionType, score)

	return nil
}

// GetUserProfile 
func (s *UserBehaviorService) GetUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	// ?
	cacheKey := fmt.Sprintf("user_profile:%s", userID)
	var cachedProfile UserProfile
	if err := s.getCachedData(ctx, cacheKey, &cachedProfile); err == nil {
		return &cachedProfile, nil
	}

	// 
	profile, err := s.buildUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 
	if err := s.setCachedData(ctx, cacheKey, profile, 30*time.Minute); err != nil {
		s.logger.Warn("Failed to cache user profile", zap.Error(err))
	}

	return profile, nil
}

// GetUserBehaviors 
func (s *UserBehaviorService) GetUserBehaviors(ctx context.Context, userID string) (map[string]float64, error) {
	var behaviors []models.UserBehavior
	
	// ?0
	since := time.Now().AddDate(0, 0, -30)
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND created_at > ?", userID, since).
		Find(&behaviors).Error; err != nil {
		return nil, fmt.Errorf("failed to get user behaviors: %w", err)
	}

	// 
	behaviorMap := make(map[string]float64)
	for _, behavior := range behaviors {
		behaviorMap[behavior.WisdomID] += behavior.Score
	}

	return behaviorMap, nil
}

// FindSimilarUsers 
func (s *UserBehaviorService) FindSimilarUsers(ctx context.Context, userID string, limit int) ([]string, error) {
	// 
	userBehaviors, err := s.GetUserBehaviors(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(userBehaviors) == 0 {
		return []string{}, nil
	}

	// ?
	var allUsers []string
	if err := s.db.WithContext(ctx).
		Model(&models.UserBehavior{}).
		Where("user_id != ?", userID).
		Distinct("user_id").
		Pluck("user_id", &allUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	// ?
	type userSimilarity struct {
		UserID     string
		Similarity float64
	}

	var similarities []userSimilarity
	for _, otherUserID := range allUsers {
		otherBehaviors, err := s.GetUserBehaviors(ctx, otherUserID)
		if err != nil {
			continue
		}

		similarity := s.calculateCosineSimilarity(userBehaviors, otherBehaviors)
		if similarity > 0.1 { // ?
			similarities = append(similarities, userSimilarity{
				UserID:     otherUserID,
				Similarity: similarity,
			})
		}
	}

	// ?
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Similarity > similarities[j].Similarity
	})

	var result []string
	for i, sim := range similarities {
		if i >= limit {
			break
		}
		result = append(result, sim.UserID)
	}

	return result, nil
}

// calculateActionScore 
func (s *UserBehaviorService) calculateActionScore(actionType string, duration int64) float64 {
	baseScore := map[string]float64{
		models.ActionTypeView:     1.0,
		models.ActionTypeLike:     3.0,
		models.ActionTypeShare:    5.0,
		models.ActionTypeComment:  4.0,
		models.ActionTypeFavorite: 6.0,
		models.ActionTypeSearch:   0.5,
		models.ActionTypeDownload: 2.0,
	}

	score := baseScore[actionType]
	if score == 0 {
		score = 1.0 // 
	}

	// 
	if actionType == models.ActionTypeView && duration > 0 {
		// ?
		durationFactor := math.Min(float64(duration)/60.0, 5.0) // ??
		score *= (1.0 + durationFactor*0.2)
	}

	return score
}

// buildUserProfile 
func (s *UserBehaviorService) buildUserProfile(ctx context.Context, userID string) (*UserProfile, error) {
	// 
	var behaviors []models.UserBehavior
	since := time.Now().AddDate(0, 0, -90) // ?0?
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND created_at > ?", userID, since).
		Find(&behaviors).Error; err != nil {
		return nil, fmt.Errorf("failed to get user behaviors: %w", err)
	}

	if len(behaviors) == 0 {
		return &UserProfile{
			UserID:      userID,
			LastActive:  time.Now(),
			TotalActions: 0,
		}, nil
	}

	// ?
	wisdomIDs := make([]string, 0, len(behaviors))
	for _, behavior := range behaviors {
		wisdomIDs = append(wisdomIDs, behavior.WisdomID)
	}

	var wisdoms []models.CulturalWisdom
	if err := s.db.WithContext(ctx).
		Where("id IN ?", wisdomIDs).
		Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("failed to get wisdoms: %w", err)
	}

	// 
	wisdomMap := make(map[string]models.CulturalWisdom)
	for _, wisdom := range wisdoms {
		wisdomMap[wisdom.ID] = wisdom
	}

	// 
	categoryScores := make(map[string]float64)
	schoolScores := make(map[string]float64)
	authorScores := make(map[string]float64)
	tagScores := make(map[string]float64)
	
	categoryCounts := make(map[string]int)
	schoolCounts := make(map[string]int)
	authorCounts := make(map[string]int)
	tagCounts := make(map[string]int)

	activeHours := make(map[int]int)
	totalScore := 0.0
	lastActive := time.Time{}

	for _, behavior := range behaviors {
		wisdom, exists := wisdomMap[behavior.WisdomID]
		if !exists {
			continue
		}

		// 
		if wisdom.Category != "" {
			categoryScores[wisdom.Category] += behavior.Score
			categoryCounts[wisdom.Category]++
		}
		if wisdom.School != "" {
			schoolScores[wisdom.School] += behavior.Score
			schoolCounts[wisdom.School]++
		}
		if wisdom.Author != "" {
			authorScores[wisdom.Author] += behavior.Score
			authorCounts[wisdom.Author]++
		}
		for _, tag := range wisdom.Tags {
			tagScores[tag] += behavior.Score
			tagCounts[tag]++
		}

		// 
		hour := behavior.CreatedAt.Hour()
		activeHours[hour]++

		totalScore += behavior.Score
		if behavior.CreatedAt.After(lastActive) {
			lastActive = behavior.CreatedAt
		}
	}

	// 
	profile := &UserProfile{
		UserID:           userID,
		PreferredCategories: s.buildCategoryScores(categoryScores, categoryCounts),
		PreferredSchools:    s.buildSchoolScores(schoolScores, schoolCounts),
		PreferredAuthors:    s.buildAuthorScores(authorScores, authorCounts),
		PreferredTags:       s.buildTagScores(tagScores, tagCounts),
		ReadingSpeed:        s.calculateReadingSpeed(behaviors),
		ActiveHours:         s.getTopActiveHours(activeHours, 5),
		LastActive:          lastActive,
		TotalActions:        int64(len(behaviors)),
		EngagementScore:     totalScore / float64(len(behaviors)),
	}

	return profile, nil
}

// updateUserPreference 
func (s *UserBehaviorService) updateUserPreference(userID, wisdomID, actionType string, score float64) {
	ctx := context.Background()
	
	// 
	var wisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", wisdomID).First(&wisdom).Error; err != nil {
		s.logger.Error("Failed to get wisdom for preference update", zap.Error(err))
		return
	}

	// ?
	var preference models.UserPreference
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&preference).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 
			preference = models.UserPreference{
				ID:     fmt.Sprintf("%s_%d", userID, time.Now().UnixNano()),
				UserID: userID,
			}
		} else {
			s.logger.Error("Failed to get user preference", zap.Error(err))
			return
		}
	}

	// 
	s.updatePreferenceData(&preference, wisdom, score)

	// 
	if err := s.db.WithContext(ctx).Save(&preference).Error; err != nil {
		s.logger.Error("Failed to save user preference", zap.Error(err))
	}

	// 
	cacheKey := fmt.Sprintf("user_profile:%s", userID)
	if err := s.deleteCachedData(ctx, cacheKey); err != nil {
		s.logger.Warn("Failed to delete cached user profile", zap.Error(err))
	}
}

// calculateCosineSimilarity ?
func (s *UserBehaviorService) calculateCosineSimilarity(behaviors1, behaviors2 map[string]float64) float64 {
	// 
	commonItems := make(map[string]bool)
	for item := range behaviors1 {
		if _, exists := behaviors2[item]; exists {
			commonItems[item] = true
		}
	}

	if len(commonItems) == 0 {
		return 0.0
	}

	// ?
	var dotProduct, norm1, norm2 float64
	
	for item := range commonItems {
		score1 := behaviors1[item]
		score2 := behaviors2[item]
		
		dotProduct += score1 * score2
		norm1 += score1 * score1
		norm2 += score2 * score2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// 
func (s *UserBehaviorService) buildCategoryScores(scores map[string]float64, counts map[string]int) []CategoryScore {
	var result []CategoryScore
	for category, score := range scores {
		result = append(result, CategoryScore{
			Category: category,
			Score:    score,
			Count:    counts[category],
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}

func (s *UserBehaviorService) buildSchoolScores(scores map[string]float64, counts map[string]int) []SchoolScore {
	var result []SchoolScore
	for school, score := range scores {
		result = append(result, SchoolScore{
			School: school,
			Score:  score,
			Count:  counts[school],
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}

func (s *UserBehaviorService) buildAuthorScores(scores map[string]float64, counts map[string]int) []AuthorScore {
	var result []AuthorScore
	for author, score := range scores {
		result = append(result, AuthorScore{
			Author: author,
			Score:  score,
			Count:  counts[author],
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}

func (s *UserBehaviorService) buildTagScores(scores map[string]float64, counts map[string]int) []TagScore {
	var result []TagScore
	for tag, score := range scores {
		result = append(result, TagScore{
			Tag:   tag,
			Score: score,
			Count: counts[tag],
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	if len(result) > 20 {
		result = result[:20]
	}
	return result
}

func (s *UserBehaviorService) calculateReadingSpeed(behaviors []models.UserBehavior) float64 {
	var totalDuration int64
	var viewCount int64
	
	for _, behavior := range behaviors {
		if behavior.ActionType == models.ActionTypeView && behavior.Duration > 0 {
			totalDuration += behavior.Duration
			viewCount++
		}
	}
	
	if viewCount == 0 {
		return 1.0
	}
	
	avgDuration := float64(totalDuration) / float64(viewCount)
	// ?0?
	return 60.0 / avgDuration
}

func (s *UserBehaviorService) getTopActiveHours(activeHours map[int]int, limit int) []int {
	type hourCount struct {
		Hour  int
		Count int
	}
	
	var hours []hourCount
	for hour, count := range activeHours {
		hours = append(hours, hourCount{Hour: hour, Count: count})
	}
	
	sort.Slice(hours, func(i, j int) bool {
		return hours[i].Count > hours[j].Count
	})
	
	var result []int
	for i, hc := range hours {
		if i >= limit {
			break
		}
		result = append(result, hc.Hour)
	}
	
	return result
}

func (s *UserBehaviorService) updatePreferenceData(preference *models.UserPreference, wisdom models.CulturalWisdom, score float64) {
	// 
	// ?
	preference.LastActive = time.Now()
}

// 渨
func (s *UserBehaviorService) getCachedData(ctx context.Context, key string, dest interface{}) error {
	if s.cache == nil || s.cache.redis == nil {
		return fmt.Errorf("cache not available")
	}
	
	result, err := s.cache.redis.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(result), dest)
}

func (s *UserBehaviorService) setCachedData(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	if s.cache == nil || s.cache.redis == nil {
		return fmt.Errorf("cache not available")
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	return s.cache.redis.Set(ctx, key, jsonData, expiration).Err()
}

func (s *UserBehaviorService) deleteCachedData(ctx context.Context, key string) error {
	if s.cache == nil || s.cache.redis == nil {
		return fmt.Errorf("cache not available")
	}
	
	return s.cache.redis.Del(ctx, key).Err()
}

