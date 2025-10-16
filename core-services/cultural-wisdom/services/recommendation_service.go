package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/codetaoist/taishanglaojun/core-services/cultural-wisdom/models"
)

// RecommendationService 推荐服务
type RecommendationService struct {
	db                  *gorm.DB
	cache               *CacheService
	userBehaviorService *UserBehaviorService
	aiService           *AIService
	logger              *zap.Logger
}

// NewRecommendationService 创建推荐服务
func NewRecommendationService(db *gorm.DB, cache *CacheService, userBehaviorService *UserBehaviorService, aiService *AIService, logger *zap.Logger) *RecommendationService {
	return &RecommendationService{
		db:                  db,
		cache:               cache,
		userBehaviorService: userBehaviorService,
		aiService:           aiService,
		logger:              logger,
	}
}

// RecommendationItem 推荐项
type RecommendationItem struct {
	WisdomID  string    `json:"wisdom_id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Category  string    `json:"category"`
	School    string    `json:"school"`
	Summary   string    `json:"summary"`
	Score     float64   `json:"score"`
	Reason    string    `json:"reason"`
	ViewCount int64     `json:"view_count"`
	LikeCount int64     `json:"like_count"`
	CreatedAt time.Time `json:"created_at"`
}

// RecommendationRequest 推荐请求
type RecommendationRequest struct {
	WisdomID   string   `json:"wisdom_id"`
	UserID     string   `json:"user_id,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Schools    []string `json:"schools,omitempty"`
	Authors    []string `json:"authors,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Limit      int      `json:"limit"`
	ExcludeIDs []string `json:"exclude_ids,omitempty"`
	Algorithm  string   `json:"algorithm,omitempty"` // "content", "collaborative", "hybrid"
}

// GetRecommendations 获取推荐项
func (s *RecommendationService) GetRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	// 3. 算法选择
	if req.Algorithm == "" {
		req.Algorithm = "hybrid"
	}

	// 4. 缓存检查
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

	// 5. 缓存存储
	if s.cache != nil {
		s.cache.SetRecommendations(ctx, cacheKey, recommendations, 30*time.Minute)
	}

	return recommendations, nil
}

// getContentBasedRecommendations 获取基于内容的推荐项
func (s *RecommendationService) getContentBasedRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	// 1. 目标智慧项
	var targetWisdom models.CulturalWisdom
	if err := s.db.WithContext(ctx).Where("id = ?", req.WisdomID).First(&targetWisdom).Error; err != nil {
		return nil, fmt.Errorf("failed to get target wisdom: %w", err)
	}

	// 2. 候选智慧项
	candidates, err := s.getCandidateWisdoms(ctx, req)
	if err != nil {
		return nil, err
	}

	// 3. 计算相似度
	var recommendations []RecommendationItem
	for _, candidate := range candidates {
		score := s.calculateContentSimilarity(targetWisdom, candidate)
		if score > 0.1 { // ?
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

	// 4. 排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	return recommendations, nil
}

// calculateVectorSimilarity 计算向量相似度
func (s *RecommendationService) calculateVectorSimilarity(target, candidate models.CulturalWisdom) float64 {
	// 1. 向量存在性检查
	if len(target.Vector) > 0 && len(candidate.Vector) > 0 {
		return s.cosineSimilarity(target.Vector, candidate.Vector)
	}

	// 2. 向量生成
	ctx := context.Background()

	// 3. 目标向量生成
	if len(target.Vector) == 0 {
		targetEmbedding, err := s.aiService.GetEmbedding(ctx, target.Content)
		if err != nil {
			s.logger.Warn("Failed to generate target embedding",
				zap.String("wisdom_id", target.ID),
				zap.Error(err))
			return 0
		}
		target.Vector = targetEmbedding

		// 4. 目标向量存储
		s.db.Model(&target).Update("vector", target.Vector)
	}

	// 5. 候选向量生成
	if len(candidate.Vector) == 0 {
		candidateEmbedding, err := s.aiService.GetEmbedding(ctx, candidate.Content)
		if err != nil {
			s.logger.Warn("Failed to generate candidate embedding",
				zap.String("wisdom_id", candidate.ID),
				zap.Error(err))
			return 0
		}
		candidate.Vector = candidateEmbedding

		// 6. 候选向量存储
		s.db.Model(&candidate).Update("vector", candidate.Vector)
	}

	return s.cosineSimilarity(target.Vector, candidate.Vector)
}

// applyDiversityOptimization 多样性优化
func (s *RecommendationService) applyDiversityOptimization(recommendations []RecommendationItem, limit int) []RecommendationItem {
	if len(recommendations) <= limit {
		return recommendations
	}

	// 7. 多样性筛选
	diversified := make([]RecommendationItem, 0, limit)
	categoryCount := make(map[string]int)
	schoolCount := make(map[string]int)
	authorCount := make(map[string]int)

	// 8. 多样性参数
	maxPerCategory := max(1, limit/3)
	maxPerSchool := max(1, limit/4)
	maxPerAuthor := max(1, limit/5)

	for _, rec := range recommendations {
		// 9. 多样性检查
		if categoryCount[rec.Category] >= maxPerCategory ||
			schoolCount[rec.School] >= maxPerSchool ||
			authorCount[rec.Author] >= maxPerAuthor {
			continue
		}

		diversified = append(diversified, rec)
		categoryCount[rec.Category]++
		schoolCount[rec.School]++
		authorCount[rec.Author]++

		if len(diversified) >= limit {
			break
		}
	}

	// 10. 填充不足项
	if len(diversified) < limit {
		for _, rec := range recommendations {
			if len(diversified) >= limit {
				break
			}

			// 11. 检查是否已存在
			exists := false
			for _, existing := range diversified {
				if existing.WisdomID == rec.WisdomID {
					exists = true
					break
				}
			}

			if !exists {
				diversified = append(diversified, rec)
			}
		}
	}

	return diversified
}

// max 取较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// cosineSimilarity 余弦相似度
func (s *RecommendationService) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// getCollaborativeRecommendations 协同过滤推荐
func (s *RecommendationService) getCollaborativeRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	// 12. 获取用户行为
	userBehaviors, err := s.getUserBehaviors(ctx, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get user behaviors, fallback to content-based", zap.Error(err))
		return s.getContentBasedRecommendations(ctx, req)
	}

	// 13. 找到相似用户
	similarUsers, err := s.findSimilarUsers(ctx, req.UserID, userBehaviors)
	if err != nil {
		s.logger.Warn("Failed to find similar users, fallback to content-based", zap.Error(err))
		return s.getContentBasedRecommendations(ctx, req)
	}

	// 14. 生成协同过滤推荐
	recommendations, err := s.generateCollaborativeRecommendations(ctx, req, similarUsers)
	if err != nil {
		return nil, err
	}

	return recommendations, nil
}

// getHybridRecommendations 混合推荐
func (s *RecommendationService) getHybridRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	// 15. 内容推荐
	contentRecs, err := s.getContentBasedRecommendations(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to get content-based recommendations", zap.Error(err))
		contentRecs = []RecommendationItem{}
	}

	// 16. 协同过滤推荐
	collaborativeRecs, err := s.getCollaborativeRecommendations(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to get collaborative recommendations", zap.Error(err))
		collaborativeRecs = []RecommendationItem{}
	}

	// 17. 热门推荐
	popularRecs, err := s.getPopularRecommendations(ctx, req)
	if err != nil {
		s.logger.Warn("Failed to get popular recommendations", zap.Error(err))
		popularRecs = []RecommendationItem{}
	}

	// 18. 合并推荐
	userProfile, err := s.userBehaviorService.GetUserProfile(ctx, req.UserID)
	if err != nil {
		s.logger.Warn("Failed to get user profile", zap.Error(err))
	}

	// 19. 合并推荐权重
	contentWeight := 0.4
	collaborativeWeight := 0.3
	popularWeight := 0.3

	if userProfile != nil {
		// 20. 新用户权重
		if userProfile.TotalActions < 10 {
			contentWeight = 0.2
			collaborativeWeight = 0.1
			popularWeight = 0.7
		} else if userProfile.TotalActions > 100 {
			// 21. 活跃用户权重
			contentWeight = 0.5
			collaborativeWeight = 0.4
			popularWeight = 0.1
		}
	}

	// 22. 合并推荐
	merged := s.mergeRecommendations(contentRecs, collaborativeRecs, popularRecs,
		contentWeight, collaborativeWeight, popularWeight)

	// 23. 多样性优化
	diversified := s.applyDiversityOptimization(merged, req.Limit)

	return diversified, nil
}

// calculateContentSimilarity ?
func (s *RecommendationService) calculateContentSimilarity(target, candidate models.CulturalWisdom) float64 {
	score := 0.0

	// 24. AI 内容相似度(: 0.4)
	if s.aiService != nil {
		vectorSimilarity := s.calculateVectorSimilarity(target, candidate)
		if vectorSimilarity > 0 {
			score += vectorSimilarity * 0.4
		}
	}

	// 25. 学校相似度(: 0.25)
	if target.School == candidate.School && target.School != "" {
		score += 0.25
	}

	// 26. 分类相似度(: 0.15)
	if target.Category == candidate.Category && target.Category != "" {
		score += 0.15
	}

	// 27. 作者相似度(: 0.1)
	if target.Author == candidate.Author && target.Author != "" {
		score += 0.1
	}

	// 28. 标签相似度(: 0.08)
	if len(target.Tags) > 0 && len(candidate.Tags) > 0 {
		commonTags := s.countCommonTags(target.Tags, candidate.Tags)
		tagSimilarity := float64(commonTags) / float64(len(target.Tags))
		score += tagSimilarity * 0.08
	}

	// 29. 内容长度相似度(: 0.02)
	targetLen := len(target.Content)
	candidateLen := len(candidate.Content)
	if targetLen > 0 && candidateLen > 0 {
		lengthRatio := math.Min(float64(targetLen), float64(candidateLen)) / math.Max(float64(targetLen), float64(candidateLen))
		score += lengthRatio * 0.02
	}

	return score
}

// getCandidateWisdoms 31. 候选内容
func (s *RecommendationService) getCandidateWisdoms(ctx context.Context, req RecommendationRequest) ([]models.CulturalWisdom, error) {
	query := s.db.WithContext(ctx).Where("status = ?", "published")

	// 30. 排除已推荐
	excludeIDs := append(req.ExcludeIDs, req.WisdomID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	// 31. 分类筛选
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

// getPopularRecommendations 32. 热门推荐
func (s *RecommendationService) getPopularRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	query := s.db.WithContext(ctx).Where("status = ?", "published")

	// 30. 排除已推荐
	excludeIDs := append(req.ExcludeIDs, req.WisdomID)
	if len(excludeIDs) > 0 {
		query = query.Where("id NOT IN ?", excludeIDs)
	}

	// 33. 热门排序
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
			Reason:    "",
			ViewCount: wisdom.ViewCount,
			LikeCount: wisdom.LikeCount,
			CreatedAt: wisdom.CreatedAt,
		})
	}

	return recommendations, nil
}

// calculatePopularityScore
func (s *RecommendationService) calculatePopularityScore(wisdom models.CulturalWisdom) float64 {
	// ?
	viewScore := math.Log10(float64(wisdom.ViewCount + 1))
	likeScore := math.Log10(float64(wisdom.LikeCount+1)) * 2

	//
	daysSinceCreated := time.Since(wisdom.CreatedAt).Hours() / 24
	timeDecay := math.Exp(-daysSinceCreated / 30) // 30

	return (viewScore + likeScore) * timeDecay
}

// mergeRecommendations 34. 合并推荐
func (s *RecommendationService) mergeRecommendations(contentRecs, collaborativeRecs, popularRecs []RecommendationItem,
	contentWeight, collaborativeWeight, popularWeight float64) []RecommendationItem {
	//
	recMap := make(map[string]RecommendationItem)

	// 34.1 内容推荐
	for _, rec := range contentRecs {
		rec.Score = rec.Score * contentWeight
		rec.Reason = "内容推荐：" + rec.Reason
		recMap[rec.WisdomID] = rec
	}

	// 34.2 协同推荐
	for _, rec := range collaborativeRecs {
		if existing, exists := recMap[rec.WisdomID]; exists {
			//
			existing.Score += rec.Score * collaborativeWeight
			existing.Reason += "协同推荐"
			recMap[rec.WisdomID] = existing
		} else {
			rec.Score = rec.Score * collaborativeWeight
			rec.Reason = "热门推荐"
			recMap[rec.WisdomID] = rec
		}
	}

	// 34.3 热门推荐
	for _, rec := range popularRecs {
		if existing, exists := recMap[rec.WisdomID]; exists {
			//
			existing.Score += rec.Score * popularWeight
			existing.Reason += "；热门推荐"
			recMap[rec.WisdomID] = existing
		} else {
			rec.Score = rec.Score * popularWeight
			rec.Reason = ""
			recMap[rec.WisdomID] = rec
		}
	}

	// 34.4 合并排序
	var recommendations []RecommendationItem
	for _, rec := range recMap {
		recommendations = append(recommendations, rec)
	}

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	return recommendations
}

// generateContentReason 34.1 内容推荐原因
func (s *RecommendationService) generateContentReason(target, candidate models.CulturalWisdom) string {
	var reasons []string

	if target.School == candidate.School && target.School != "" {
		reasons = append(reasons, fmt.Sprintf("%s", target.School))
	}

	if target.Category == candidate.Category && target.Category != "" {
		reasons = append(reasons, "相同类别")
	}

	if target.Author == candidate.Author && target.Author != "" {
		reasons = append(reasons, fmt.Sprintf("同作者：%s", target.Author))
	}

	if len(target.Tags) > 0 && len(candidate.Tags) > 0 {
		commonTags := s.countCommonTags(target.Tags, candidate.Tags)
		if commonTags > 0 {
			reasons = append(reasons, fmt.Sprintf("共同标签：%d个", commonTags))
		}
	}

	if target.Category == candidate.Category && target.Category != "" {
		reasons = append(reasons, "相同类别")
	}

	if len(reasons) == 0 {
		return ""
	}

	return strings.Join(reasons, "；")
}

// countCommonTags 34.1 共同标签数量
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

// buildCacheKey 34.4 缓存键构建
func (s *RecommendationService) buildCacheKey(req RecommendationRequest) string {
	return fmt.Sprintf("recommendations:%s:%s:%d", req.WisdomID, req.Algorithm, req.Limit)
}

// getUserBehaviors 34.2 用户行为获取
func (s *RecommendationService) getUserBehaviors(ctx context.Context, userID string) (map[string]float64, error) {
	if s.userBehaviorService == nil {
		return make(map[string]float64), nil
	}
	return s.userBehaviorService.GetUserBehaviors(ctx, userID)
}

// findSimilarUsers 34.2 相似用户查找
func (s *RecommendationService) findSimilarUsers(ctx context.Context, userID string, behaviors map[string]float64) ([]string, error) {
	if s.userBehaviorService == nil {
		return []string{}, nil
	}
	return s.userBehaviorService.FindSimilarUsers(ctx, userID, 10)
}

// generateCollaborativeRecommendations 34.3 协同推荐生成
func (s *RecommendationService) generateCollaborativeRecommendations(ctx context.Context, req RecommendationRequest, similarUsers []string) ([]RecommendationItem, error) {
	if len(similarUsers) == 0 {
		// 34.3.1 无相似用户时，返回内容推荐
		return s.getContentBasedRecommendations(ctx, req)
	}

	// 34.3.2 相似用户行为聚合
	userWisdomScores := make(map[string]float64)

	for _, similarUserID := range similarUsers {
		behaviors, err := s.userBehaviorService.GetUserBehaviors(ctx, similarUserID)
		if err != nil {
			s.logger.Warn("Failed to get similar user behaviors",
				zap.String("similar_user_id", similarUserID),
				zap.Error(err))
			continue
		}

		// 34.3.3 相似用户行为加权
		for wisdomID, score := range behaviors {
			//
			if wisdomID != req.WisdomID {
				userWisdomScores[wisdomID] += score
			}
		}
	}

	// 34.3.4 排除已交互内容
	currentUserBehaviors, err := s.userBehaviorService.GetUserBehaviors(ctx, req.UserID)
	if err == nil {
		for wisdomID := range currentUserBehaviors {
			delete(userWisdomScores, wisdomID)
		}
	}

	// 34.3.5 排序
	var recommendations []RecommendationItem
	for wisdomID, score := range userWisdomScores {
		//
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
			Score:     score / float64(len(similarUsers)), // ?
			Reason:    "",
			ViewCount: wisdom.ViewCount,
			LikeCount: wisdom.LikeCount,
			CreatedAt: wisdom.CreatedAt,
		})
	}

	// 34.3.6 排序
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	return recommendations, nil
}
