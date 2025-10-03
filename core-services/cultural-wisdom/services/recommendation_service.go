package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"go.uber.org/zap"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// RecommendationService 推荐服务
type RecommendationService struct {
	db                  *gorm.DB
	cache               *CacheService
	userBehaviorService *UserBehaviorService
	logger              *zap.Logger
}

// NewRecommendationService 创建推荐服务实例
func NewRecommendationService(db *gorm.DB, cache *CacheService, userBehaviorService *UserBehaviorService, logger *zap.Logger) *RecommendationService {
	return &RecommendationService{
		db:                  db,
		cache:               cache,
		userBehaviorService: userBehaviorService,
		logger:              logger,
	}
}

// RecommendationItem 推荐项
type RecommendationItem struct {
	WisdomID    string  `json:"wisdom_id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Category    string  `json:"category"`
	School      string  `json:"school"`
	Summary     string  `json:"summary"`
	Score       float64 `json:"score"`
	Reason      string  `json:"reason"`
	ViewCount   int64   `json:"view_count"`
	LikeCount   int64   `json:"like_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// RecommendationRequest 推荐请求
type RecommendationRequest struct {
	WisdomID     string   `json:"wisdom_id"`
	UserID       string   `json:"user_id,omitempty"`
	Categories   []string `json:"categories,omitempty"`
	Schools      []string `json:"schools,omitempty"`
	Authors      []string `json:"authors,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Limit        int      `json:"limit"`
	ExcludeIDs   []string `json:"exclude_ids,omitempty"`
	Algorithm    string   `json:"algorithm,omitempty"` // "content", "collaborative", "hybrid"
}

// GetRecommendations 获取推荐
func (s *RecommendationService) GetRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	// 设置默认算法
	if req.Algorithm == "" {
		req.Algorithm = "hybrid"
	}

	// 检查缓存
	cacheKey := s.buildCacheKey(req)
	if s.cache != nil {
		if cached, err := s.cache.GetRecommendations(ctx, cacheKey); err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	var recommendations []RecommendationItem
	var err error

	switch req.Algorithm {
	case "content":
		recommendations, err = s.getContentBasedRecommendations(ctx, req)
	case "collaborative":
		recommendations, err = s.getCollaborativeRecommendations(ctx, req)
	case "hybrid":
		recommendations, err = s.getHybridRecommendations(ctx, req)
	default:
		recommendations, err = s.getHybridRecommendations(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	// 缓存结果
	if s.cache != nil {
		s.cache.SetRecommendations(ctx, cacheKey, recommendations, 30*time.Minute)
	}

	return recommendations, nil
}

// getContentBasedRecommendations 基于内容的推荐
func (s *RecommendationService) getContentBasedRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	// 获取目标智慧
	var targetWisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", req.WisdomID).First(&targetWisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to get target wisdom: %w", err)
	}

	// 获取候选智慧
	candidates, err := s.getCandidateWisdoms(ctx, req)
	if err != nil {
		return nil, err
	}

	// 计算内容相似度
	var recommendations []RecommendationItem
	for _, candidate := range candidates {
		score := s.calculateContentSimilarity(targetWisdom, candidate)
		if score > 0.1 { // 设置最低相似度阈值
			reason := s.generateContentReason(targetWisdom, candidate)
			recommendations = append(recommendations, RecommendationItem{
				WisdomID:  candidate.ID,
				Title:     candidate.Title,
				Author:    candidate.Author,
				Category:  candidate.Category,
				School:    candidate.School,
				Summary:   candidate.Summary,
				Score:     score,
				Reason:    reason,
				ViewCount: candidate.ViewCount,
				LikeCount: candidate.LikeCount,
				CreatedAt: candidate.CreatedAt,
			})
		}
	}

	// 排序并限制数量
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	return recommendations, nil
}

// getCollaborativeRecommendations 基于协同过滤的推荐
func (s *RecommendationService) getCollaborativeRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	// 获取用户行为数据（浏览、点赞等）
	userBehaviors, err := s.getUserBehaviors(ctx, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get user behaviors, fallback to content-based", zap.Error(err))
		return s.getContentBasedRecommendations(ctx, req)
	}

	// 找到相似用户
	similarUsers, err := s.findSimilarUsers(ctx, req.UserID, userBehaviors)
	if err != nil {
		s.logger.Warn("Failed to find similar users, fallback to content-based", zap.Error(err))
		return s.getContentBasedRecommendations(ctx, req)
	}

	// 基于相似用户的偏好推荐
	recommendations, err := s.generateCollaborativeRecommendations(ctx, req, similarUsers)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

// getHybridRecommendations 混合推荐
func (s *RecommendationService) getHybridRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	// 获取基于内容的推荐
	contentRecs, err := s.getContentBasedRecommendations(ctx, req)
	if err != nil {
		return nil, err
	}

	// 获取热门推荐
	popularRecs, err := s.getPopularRecommendations(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to get popular recommendations", zap.Error(err))
		popularRecs = []RecommendationItem{}
	}

	// 合并和重新评分
	recommendations := s.mergeRecommendations(contentRecs, popularRecs, req.Limit)

	return recommendations, nil
}

// calculateContentSimilarity 计算内容相似度
func (s *RecommendationService) calculateContentSimilarity(target, candidate models.CulturalWisdom) float64 {
	score := 0.0

	// 学派相似度 (权重: 0.3)
	if target.School == candidate.School && target.School != "" {
		score += 0.3
	}

	// 分类相似度 (权重: 0.2)
	if target.Category == candidate.Category && target.Category != "" {
		score += 0.2
	}

	// 作者相似度 (权重: 0.2)
	if target.Author == candidate.Author && target.Author != "" {
		score += 0.2
	}

	// 标签相似度 (权重: 0.15)
	if len(target.Tags) > 0 && len(candidate.Tags) > 0 {
		commonTags := s.countCommonTags(target.Tags, candidate.Tags)
		tagSimilarity := float64(commonTags) / float64(len(target.Tags))
		score += tagSimilarity * 0.15
	}

	// 内容长度相似度 (权重: 0.05)
	targetLen := len(target.Content)
	candidateLen := len(candidate.Content)
	if targetLen > 0 && candidateLen > 0 {
		lengthRatio := math.Min(float64(targetLen), float64(candidateLen)) / math.Max(float64(targetLen), float64(candidateLen))
		score += lengthRatio * 0.05
	}

	// 作者相似度 (权重: 0.05)
	if target.Author == candidate.Author && target.Author != "" {
		score += 0.05
	}

	return score
}

// getCandidateWisdoms 获取候选智慧
func (s *RecommendationService) getCandidateWisdoms(ctx context.Context, req RecommendationRequest) ([]models.CulturalWisdom, error) {
	query := s.db.WithContext(ctx).Where("status = ?", "published")

	// 排除指定的智慧
	excludeIDs := append(req.ExcludeIDs, req.WisdomID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	// 应用过滤条件
	if len(req.Categories) > 0 {
		query = query.Where("category IN ?", req.Categories)
	}
	if len(req.Schools) > 0 {
		query = query.Where("school IN ?", req.Schools)
	}
	if len(req.Authors) > 0 {
		query = query.Where("author IN ?", req.Authors)
	}

	var candidates []models.CulturalWisdom
	if err := query.Limit(req.Limit * 5).Find(&candidates).Error; err != nil {
		return nil, fmt.Errorf("failed to get candidate wisdoms: %w", err)
	}

	return candidates, nil
}

// getPopularRecommendations 获取热门推荐
func (s *RecommendationService) getPopularRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	query := s.db.WithContext(ctx).Where("status = ?", "published")

	// 排除指定的智慧
	excludeIDs := append(req.ExcludeIDs, req.WisdomID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	// 按热度排序（浏览量 + 点赞量）
	var wisdoms []models.CulturalWisdom
	if err := query.Order("(view_count + like_count * 2) DESC").
		Limit(req.Limit).Find(&wisdoms).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular wisdoms: %w", err)
	}

	var recommendations []RecommendationItem
	for _, wisdom := range wisdoms {
		score := s.calculatePopularityScore(wisdom)
		recommendations = append(recommendations, RecommendationItem{
			WisdomID:  wisdom.ID,
			Title:     wisdom.Title,
			Author:    wisdom.Author,
			Category:  wisdom.Category,
			School:    wisdom.School,
			Summary:   wisdom.Summary,
			Score:     score,
			Reason:    "热门推荐",
			ViewCount: wisdom.ViewCount,
			LikeCount: wisdom.LikeCount,
			CreatedAt: wisdom.CreatedAt,
		})
	}

	return recommendations, nil
}

// calculatePopularityScore 计算热度分数
func (s *RecommendationService) calculatePopularityScore(wisdom models.CulturalWisdom) float64 {
	// 基于浏览量和点赞量计算热度分数
	viewScore := math.Log10(float64(wisdom.ViewCount + 1))
	likeScore := math.Log10(float64(wisdom.LikeCount + 1)) * 2

	// 时间衰减因子
	daysSinceCreated := time.Since(wisdom.CreatedAt).Hours() / 24
	timeDecay := math.Exp(-daysSinceCreated / 30) // 30天半衰期

	return (viewScore + likeScore) * timeDecay
}

// mergeRecommendations 合并推荐结果
func (s *RecommendationService) mergeRecommendations(contentRecs, popularRecs []RecommendationItem, limit int) []RecommendationItem {
	// 创建推荐项映射，避免重复
	recMap := make(map[string]RecommendationItem)

	// 添加基于内容的推荐（权重: 0.7）
	for _, rec := range contentRecs {
		rec.Score = rec.Score * 0.7
		rec.Reason = "内容相似：" + rec.Reason
		recMap[rec.WisdomID] = rec
	}

	// 添加热门推荐（权重: 0.3）
	for _, rec := range popularRecs {
		if existing, exists := recMap[rec.WisdomID]; exists {
			// 如果已存在，合并分数
			existing.Score += rec.Score * 0.3
			existing.Reason += "，热门推荐"
			recMap[rec.WisdomID] = existing
		} else {
			rec.Score = rec.Score * 0.3
			recMap[rec.WisdomID] = rec
		}
	}

	// 转换为切片并排序
	var recommendations []RecommendationItem
	for _, rec := range recMap {
		recommendations = append(recommendations, rec)
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}

	return recommendations
}

// generateContentReason 生成内容推荐理由
func (s *RecommendationService) generateContentReason(target, candidate models.CulturalWisdom) string {
	var reasons []string

	if target.School == candidate.School && target.School != "" {
		reasons = append(reasons, fmt.Sprintf("同属%s学派", target.School))
	}

	if target.Category == candidate.Category && target.Category != "" {
		reasons = append(reasons, "同类别智慧")
	}

	if target.Author == candidate.Author && target.Author != "" {
		reasons = append(reasons, fmt.Sprintf("同为%s的作品", target.Author))
	}

	if len(target.Tags) > 0 && len(candidate.Tags) > 0 {
		commonTags := s.countCommonTags(target.Tags, candidate.Tags)
		if commonTags > 0 {
			reasons = append(reasons, fmt.Sprintf("有%d个共同标签", commonTags))
		}
	}

	if target.Category == candidate.Category && target.Category != "" {
		reasons = append(reasons, "同类别智慧")
	}

	if len(reasons) == 0 {
		return "相关智慧推荐"
	}

	return strings.Join(reasons, "，")
}

// countCommonTags 计算共同标签数量
func (s *RecommendationService) countCommonTags(tags1, tags2 []string) int {
	tagSet := make(map[string]bool)
	for _, tag := range tags1 {
		tagSet[tag] = true
	}

	count := 0
	for _, tag := range tags2 {
		if tagSet[tag] {
			count++
		}
	}

	return count
}

// buildCacheKey 构建缓存键
func (s *RecommendationService) buildCacheKey(req RecommendationRequest) string {
	return fmt.Sprintf("recommendations:%s:%s:%d", req.WisdomID, req.Algorithm, req.Limit)
}

// getUserBehaviors 获取用户行为数据
func (s *RecommendationService) getUserBehaviors(ctx context.Context, userID string) (map[string]float64, error) {
	if s.userBehaviorService == nil {
		return make(map[string]float64), nil
	}
	return s.userBehaviorService.GetUserBehaviors(ctx, userID)
}

// findSimilarUsers 找到相似用户
func (s *RecommendationService) findSimilarUsers(ctx context.Context, userID string, behaviors map[string]float64) ([]string, error) {
	if s.userBehaviorService == nil {
		return []string{}, nil
	}
	return s.userBehaviorService.FindSimilarUsers(ctx, userID, 10)
}

// generateCollaborativeRecommendations 生成协同过滤推荐
func (s *RecommendationService) generateCollaborativeRecommendations(ctx context.Context, req RecommendationRequest, similarUsers []string) ([]RecommendationItem, error) {
	if len(similarUsers) == 0 {
		// 如果没有相似用户，回退到基于内容的推荐
		return s.getContentBasedRecommendations(ctx, req)
	}

	// 获取相似用户的行为数据
	userWisdomScores := make(map[string]float64)
	
	for _, similarUserID := range similarUsers {
		behaviors, err := s.userBehaviorService.GetUserBehaviors(ctx, similarUserID)
		if err != nil {
			s.logger.Warn("Failed to get similar user behaviors", 
				zap.String("similar_user_id", similarUserID), 
				zap.Error(err))
			continue
		}
		
		// 累计相似用户对各个智慧的评分
		for wisdomID, score := range behaviors {
			// 排除用户已经交互过的智慧
			if wisdomID != req.WisdomID {
				userWisdomScores[wisdomID] += score
			}
		}
	}

	// 获取当前用户的行为，排除已交互的智慧
	currentUserBehaviors, err := s.userBehaviorService.GetUserBehaviors(ctx, req.UserID)
	if err == nil {
		for wisdomID := range currentUserBehaviors {
			delete(userWisdomScores, wisdomID)
		}
	}

	// 转换为推荐项
	var recommendations []RecommendationItem
	for wisdomID, score := range userWisdomScores {
		// 获取智慧详情
		var wisdom models.CulturalWisdom
		if err := s.db.WithContext(ctx).Where("id = ? AND status = ?", wisdomID, "published").First(&wisdom).Error; err != nil {
			continue
		}

		recommendations = append(recommendations, RecommendationItem{
			WisdomID:  wisdom.ID,
			Title:     wisdom.Title,
			Author:    wisdom.Author,
			Category:  wisdom.Category,
			School:    wisdom.School,
			Summary:   wisdom.Summary,
			Score:     score / float64(len(similarUsers)), // 平均分
			Reason:    "基于相似用户偏好推荐",
			ViewCount: wisdom.ViewCount,
			LikeCount: wisdom.LikeCount,
			CreatedAt: wisdom.CreatedAt,
		})
	}

	// 排序并限制数量
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	return recommendations, nil
}