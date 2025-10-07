package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContentBasedModel 基于内容的推荐模型
type ContentBasedModel struct {
	contentFeatures map[uuid.UUID]*ContentProfile
	userProfiles    map[uuid.UUID]*UserProfile
	modelInfo       ModelInfo
	weights         map[string]float64
}

// ContentProfile 内容画像
type ContentProfile struct {
	ContentID    uuid.UUID              `json:"content_id"`
	FeatureVector map[string]float64    `json:"feature_vector"`
	Categories   []string               `json:"categories"`
	Tags         []string               `json:"tags"`
	Keywords     []string               `json:"keywords"`
	Difficulty   float64                `json:"difficulty"`
	Duration     int64                  `json:"duration"`
	Quality      float64                `json:"quality"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// UserProfile 用户画像
type UserProfile struct {
	UserID        uuid.UUID           `json:"user_id"`
	Preferences   map[string]float64  `json:"preferences"`
	Categories    map[string]float64  `json:"categories"`
	Tags          map[string]float64  `json:"tags"`
	Keywords      map[string]float64  `json:"keywords"`
	SkillLevels   map[string]float64  `json:"skill_levels"`
	LearningStyle string              `json:"learning_style"`
	Difficulty    float64             `json:"preferred_difficulty"`
	Duration      int64               `json:"preferred_duration"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// NewContentBasedModel 创建基于内容的推荐模型
func NewContentBasedModel() *ContentBasedModel {
	return &ContentBasedModel{
		contentFeatures: make(map[uuid.UUID]*ContentProfile),
		userProfiles:    make(map[uuid.UUID]*UserProfile),
		weights: map[string]float64{
			"category":   0.3,
			"tag":        0.2,
			"keyword":    0.2,
			"difficulty": 0.15,
			"duration":   0.1,
			"quality":    0.05,
		},
		modelInfo: ModelInfo{
			Name:    "Content-Based Filtering",
			Version: "1.0",
			Type:    "content_based",
		},
	}
}

// Train 训练基于内容的模型
func (m *ContentBasedModel) Train(ctx context.Context, data *TrainingData) error {
	// 构建内容画像
	m.buildContentProfiles(data)
	
	// 构建用户画像
	m.buildUserProfiles(data)
	
	// 优化特征权重
	m.optimizeWeights(data)
	
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = m.evaluateModel(data)
	
	return nil
}

// Predict 预测推荐
func (m *ContentBasedModel) Predict(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	userProfile := m.userProfiles[userID]
	if userProfile == nil {
		return nil, fmt.Errorf("user profile not found for user %s", userID)
	}
	
	predictions := make([]Prediction, 0, len(candidates))
	
	for _, contentID := range candidates {
		contentProfile := m.contentFeatures[contentID]
		if contentProfile == nil {
			continue
		}
		
		score := m.calculateContentSimilarity(userProfile, contentProfile)
		confidence := m.calculateContentConfidence(userProfile, contentProfile)
		
		prediction := Prediction{
			ContentID:   contentID,
			Score:       score,
			Confidence:  confidence,
			Explanation: m.generateContentExplanation(userProfile, contentProfile),
			Factors:     m.getContentFactors(userProfile, contentProfile),
		}
		
		predictions = append(predictions, prediction)
	}
	
	// 按分数排序
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	return predictions, nil
}

// GetModelInfo 获取模型信息
func (m *ContentBasedModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// buildContentProfiles 构建内容画像
func (m *ContentBasedModel) buildContentProfiles(data *TrainingData) {
	for _, feature := range data.ContentFeatures {
		profile := &ContentProfile{
			ContentID:     feature.ContentID,
			FeatureVector: feature.Features,
			Categories:    feature.Categories,
			Tags:          feature.Tags,
			Keywords:      m.extractKeywords(feature),
			Difficulty:    feature.Difficulty,
			Duration:      feature.Duration,
			Quality:       m.calculateContentQuality(feature, data),
			Metadata:      feature.Metadata,
		}
		
		m.contentFeatures[feature.ContentID] = profile
	}
}

// buildUserProfiles 构建用户画像
func (m *ContentBasedModel) buildUserProfiles(data *TrainingData) {
	for _, userFeature := range data.UserFeatures {
		profile := &UserProfile{
			UserID:        userFeature.UserID,
			Preferences:   userFeature.Preferences,
			Categories:    make(map[string]float64),
			Tags:          make(map[string]float64),
			Keywords:      make(map[string]float64),
			SkillLevels:   userFeature.Skills,
			LearningStyle: userFeature.LearningStyle,
			Difficulty:    m.calculatePreferredDifficulty(userFeature.UserID, data),
			Duration:      m.calculatePreferredDuration(userFeature.UserID, data),
			UpdatedAt:     time.Now(),
		}
		
		// 基于用户交互构建偏好
		m.buildUserPreferencesFromInteractions(profile, userFeature.UserID, data)
		
		m.userProfiles[userFeature.UserID] = profile
	}
}

// buildUserPreferencesFromInteractions 基于交互构建用户偏好
func (m *ContentBasedModel) buildUserPreferencesFromInteractions(profile *UserProfile, userID uuid.UUID, data *TrainingData) {
	categoryWeights := make(map[string]float64)
	tagWeights := make(map[string]float64)
	keywordWeights := make(map[string]float64)
	
	totalWeight := 0.0
	
	for _, interaction := range data.UserInteractions {
		if interaction.UserID != userID {
			continue
		}
		
		contentProfile := m.contentFeatures[interaction.ContentID]
		if contentProfile == nil {
			continue
		}
		
		weight := m.calculateInteractionWeight(interaction)
		totalWeight += weight
		
		// 累积类别权重
		for _, category := range contentProfile.Categories {
			categoryWeights[category] += weight
		}
		
		// 累积标签权重
		for _, tag := range contentProfile.Tags {
			tagWeights[tag] += weight
		}
		
		// 累积关键词权重
		for _, keyword := range contentProfile.Keywords {
			keywordWeights[keyword] += weight
		}
	}
	
	// 归一化权重
	if totalWeight > 0 {
		for category, weight := range categoryWeights {
			profile.Categories[category] = weight / totalWeight
		}
		
		for tag, weight := range tagWeights {
			profile.Tags[tag] = weight / totalWeight
		}
		
		for keyword, weight := range keywordWeights {
			profile.Keywords[keyword] = weight / totalWeight
		}
	}
}

// calculateInteractionWeight 计算交互权重
func (m *ContentBasedModel) calculateInteractionWeight(interaction UserInteraction) float64 {
	baseWeight := 1.0
	
	switch interaction.Interaction {
	case "view":
		baseWeight = 1.0
	case "like":
		baseWeight = 3.0
	case "complete":
		baseWeight = 5.0
	case "share":
		baseWeight = 4.0
	}
	
	// 时间衰减
	daysSince := time.Since(interaction.Timestamp).Hours() / 24
	timeDecay := math.Exp(-daysSince / 30) // 30天半衰期
	
	// 时长加成
	durationBonus := 1.0
	if interaction.Duration > 0 {
		durationBonus = math.Min(1.0+float64(interaction.Duration)/3600.0, 2.0)
	}
	
	return baseWeight * timeDecay * durationBonus
}

// calculateContentSimilarity 计算内容相似度
func (m *ContentBasedModel) calculateContentSimilarity(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	var totalScore float64
	
	// 类别匹配
	categoryScore := m.calculateCategoryMatch(userProfile, contentProfile)
	totalScore += categoryScore * m.weights["category"]
	
	// 标签匹配
	tagScore := m.calculateTagMatch(userProfile, contentProfile)
	totalScore += tagScore * m.weights["tag"]
	
	// 关键词匹配
	keywordScore := m.calculateKeywordMatch(userProfile, contentProfile)
	totalScore += keywordScore * m.weights["keyword"]
	
	// 难度匹配
	difficultyScore := m.calculateDifficultyMatch(userProfile, contentProfile)
	totalScore += difficultyScore * m.weights["difficulty"]
	
	// 时长匹配
	durationScore := m.calculateDurationMatch(userProfile, contentProfile)
	totalScore += durationScore * m.weights["duration"]
	
	// 质量分数
	qualityScore := contentProfile.Quality
	totalScore += qualityScore * m.weights["quality"]
	
	return math.Min(totalScore, 1.0)
}

// calculateCategoryMatch 计算类别匹配度
func (m *ContentBasedModel) calculateCategoryMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	var maxMatch float64
	
	for _, category := range contentProfile.Categories {
		if weight, exists := userProfile.Categories[category]; exists {
			if weight > maxMatch {
				maxMatch = weight
			}
		}
	}
	
	return maxMatch
}

// calculateTagMatch 计算标签匹配度
func (m *ContentBasedModel) calculateTagMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	var totalMatch float64
	matchCount := 0
	
	for _, tag := range contentProfile.Tags {
		if weight, exists := userProfile.Tags[tag]; exists {
			totalMatch += weight
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0
	}
	
	return totalMatch / float64(matchCount)
}

// calculateKeywordMatch 计算关键词匹配度
func (m *ContentBasedModel) calculateKeywordMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	var totalMatch float64
	matchCount := 0
	
	for _, keyword := range contentProfile.Keywords {
		if weight, exists := userProfile.Keywords[keyword]; exists {
			totalMatch += weight
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0
	}
	
	return totalMatch / float64(matchCount)
}

// calculateDifficultyMatch 计算难度匹配度
func (m *ContentBasedModel) calculateDifficultyMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	diff := math.Abs(userProfile.Difficulty - contentProfile.Difficulty)
	return math.Exp(-diff) // 指数衰减
}

// calculateDurationMatch 计算时长匹配度
func (m *ContentBasedModel) calculateDurationMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	if userProfile.Duration == 0 {
		return 0.5 // 默认中等匹配
	}
	
	ratio := float64(contentProfile.Duration) / float64(userProfile.Duration)
	if ratio > 1 {
		ratio = 1 / ratio
	}
	
	return ratio
}

// calculateContentConfidence 计算内容推荐置信度
func (m *ContentBasedModel) calculateContentConfidence(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	// 基于特征覆盖度
	featureCoverage := 0.0
	totalFeatures := 0.0
	
	// 类别覆盖
	if len(contentProfile.Categories) > 0 {
		totalFeatures++
		for _, category := range contentProfile.Categories {
			if _, exists := userProfile.Categories[category]; exists {
				featureCoverage++
				break
			}
		}
	}
	
	// 标签覆盖
	if len(contentProfile.Tags) > 0 {
		totalFeatures++
		for _, tag := range contentProfile.Tags {
			if _, exists := userProfile.Tags[tag]; exists {
				featureCoverage++
				break
			}
		}
	}
	
	// 关键词覆盖
	if len(contentProfile.Keywords) > 0 {
		totalFeatures++
		for _, keyword := range contentProfile.Keywords {
			if _, exists := userProfile.Keywords[keyword]; exists {
				featureCoverage++
				break
			}
		}
	}
	
	if totalFeatures == 0 {
		return 0.1
	}
	
	return featureCoverage / totalFeatures
}

// generateContentExplanation 生成内容推荐解释
func (m *ContentBasedModel) generateContentExplanation(userProfile *UserProfile, contentProfile *ContentProfile) string {
	reasons := []string{}
	
	// 检查类别匹配
	for _, category := range contentProfile.Categories {
		if weight, exists := userProfile.Categories[category]; exists && weight > 0.3 {
			reasons = append(reasons, fmt.Sprintf("您对%s类内容感兴趣", category))
			break
		}
	}
	
	// 检查标签匹配
	matchedTags := []string{}
	for _, tag := range contentProfile.Tags {
		if weight, exists := userProfile.Tags[tag]; exists && weight > 0.2 {
			matchedTags = append(matchedTags, tag)
		}
	}
	if len(matchedTags) > 0 {
		reasons = append(reasons, fmt.Sprintf("包含您关注的标签: %s", strings.Join(matchedTags[:min(3, len(matchedTags))], ", ")))
	}
	
	// 检查难度匹配
	diffMatch := m.calculateDifficultyMatch(userProfile, contentProfile)
	if diffMatch > 0.8 {
		reasons = append(reasons, "难度适合您的水平")
	}
	
	if len(reasons) == 0 {
		return "基于您的学习偏好推荐"
	}
	
	return strings.Join(reasons, "; ")
}

// getContentFactors 获取内容推荐因子
func (m *ContentBasedModel) getContentFactors(userProfile *UserProfile, contentProfile *ContentProfile) map[string]float64 {
	factors := make(map[string]float64)
	
	factors["category_match"] = m.calculateCategoryMatch(userProfile, contentProfile)
	factors["tag_match"] = m.calculateTagMatch(userProfile, contentProfile)
	factors["keyword_match"] = m.calculateKeywordMatch(userProfile, contentProfile)
	factors["difficulty_match"] = m.calculateDifficultyMatch(userProfile, contentProfile)
	factors["duration_match"] = m.calculateDurationMatch(userProfile, contentProfile)
	factors["quality"] = contentProfile.Quality
	
	return factors
}

// extractKeywords 提取关键词
func (m *ContentBasedModel) extractKeywords(feature ContentFeature) []string {
	keywords := []string{}
	
	// 从标签中提取
	keywords = append(keywords, feature.Tags...)
	
	// 从元数据中提取
	if title, exists := feature.Metadata["title"].(string); exists {
		titleKeywords := m.extractFromText(title)
		keywords = append(keywords, titleKeywords...)
	}
	
	if description, exists := feature.Metadata["description"].(string); exists {
		descKeywords := m.extractFromText(description)
		keywords = append(keywords, descKeywords...)
	}
	
	return m.deduplicateKeywords(keywords)
}

// extractFromText 从文本中提取关键词
func (m *ContentBasedModel) extractFromText(text string) []string {
	// 简单的关键词提取（实际应用中可以使用更复杂的NLP技术）
	words := strings.Fields(strings.ToLower(text))
	keywords := []string{}
	
	for _, word := range words {
		if len(word) > 3 && !m.isStopWord(word) {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// isStopWord 检查是否为停用词
func (m *ContentBasedModel) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
	}
	
	return stopWords[word]
}

// deduplicateKeywords 去重关键词
func (m *ContentBasedModel) deduplicateKeywords(keywords []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, keyword := range keywords {
		if !seen[keyword] {
			seen[keyword] = true
			result = append(result, keyword)
		}
	}
	
	return result
}

// calculateContentQuality 计算内容质量
func (m *ContentBasedModel) calculateContentQuality(feature ContentFeature, data *TrainingData) float64 {
	// 基于用户评分和交互计算质量
	var totalRating float64
	var ratingCount int
	var interactionCount int
	
	for _, rating := range data.Ratings {
		if rating.ContentID == feature.ContentID {
			totalRating += rating.Rating
			ratingCount++
		}
	}
	
	for _, interaction := range data.UserInteractions {
		if interaction.ContentID == feature.ContentID {
			interactionCount++
		}
	}
	
	avgRating := 0.5 // 默认质量
	if ratingCount > 0 {
		avgRating = totalRating / float64(ratingCount) / 5.0 // 归一化到0-1
	}
	
	popularity := math.Log(float64(interactionCount)+1) / 10.0 // 流行度因子
	
	return math.Min(avgRating+popularity*0.1, 1.0)
}

// calculatePreferredDifficulty 计算用户偏好难度
func (m *ContentBasedModel) calculatePreferredDifficulty(userID uuid.UUID, data *TrainingData) float64 {
	var totalDifficulty float64
	var count int
	
	for _, interaction := range data.UserInteractions {
		if interaction.UserID == userID {
			if contentFeature := m.findContentFeature(interaction.ContentID, data); contentFeature != nil {
				weight := m.calculateInteractionWeight(interaction)
				totalDifficulty += contentFeature.Difficulty * weight
				count++
			}
		}
	}
	
	if count == 0 {
		return 0.5 // 默认中等难度
	}
	
	return totalDifficulty / float64(count)
}

// calculatePreferredDuration 计算用户偏好时长
func (m *ContentBasedModel) calculatePreferredDuration(userID uuid.UUID, data *TrainingData) int64 {
	var totalDuration int64
	var count int
	
	for _, interaction := range data.UserInteractions {
		if interaction.UserID == userID && interaction.Duration > 0 {
			totalDuration += interaction.Duration
			count++
		}
	}
	
	if count == 0 {
		return 1800 // 默认30分钟
	}
	
	return totalDuration / int64(count)
}

// findContentFeature 查找内容特征
func (m *ContentBasedModel) findContentFeature(contentID uuid.UUID, data *TrainingData) *ContentFeature {
	for _, feature := range data.ContentFeatures {
		if feature.ContentID == contentID {
			return &feature
		}
	}
	return nil
}

// optimizeWeights 优化特征权重
func (m *ContentBasedModel) optimizeWeights(data *TrainingData) {
	// 简单的网格搜索优化（实际应用中可以使用更复杂的优化算法）
	bestAccuracy := 0.0
	bestWeights := make(map[string]float64)
	
	// 复制当前权重
	for k, v := range m.weights {
		bestWeights[k] = v
	}
	
	// 尝试不同的权重组合
	steps := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	
	for _, categoryWeight := range steps {
		for _, tagWeight := range steps {
			for _, keywordWeight := range steps {
				remaining := 1.0 - categoryWeight - tagWeight - keywordWeight
				if remaining < 0 {
					continue
				}
				
				testWeights := map[string]float64{
					"category":   categoryWeight,
					"tag":        tagWeight,
					"keyword":    keywordWeight,
					"difficulty": remaining * 0.5,
					"duration":   remaining * 0.3,
					"quality":    remaining * 0.2,
				}
				
				// 临时设置权重
				originalWeights := m.weights
				m.weights = testWeights
				
				// 评估准确率
				accuracy := m.evaluateModel(data)
				
				if accuracy > bestAccuracy {
					bestAccuracy = accuracy
					for k, v := range testWeights {
						bestWeights[k] = v
					}
				}
				
				// 恢复原权重
				m.weights = originalWeights
			}
		}
	}
	
	// 设置最佳权重
	m.weights = bestWeights
}

// evaluateModel 评估模型
func (m *ContentBasedModel) evaluateModel(data *TrainingData) float64 {
	// 使用交叉验证评估
	correct := 0
	total := 0
	
	for _, rating := range data.Ratings {
		userProfile := m.userProfiles[rating.UserID]
		contentProfile := m.contentFeatures[rating.ContentID]
		
		if userProfile != nil && contentProfile != nil {
			predicted := m.calculateContentSimilarity(userProfile, contentProfile) * 5.0
			actual := rating.Rating
			
			// 如果预测值与实际值差距在1分以内，认为预测正确
			if math.Abs(predicted-actual) <= 1.0 {
				correct++
			}
			total++
		}
	}
	
	if total == 0 {
		return 0
	}
	
	return float64(correct) / float64(total)
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}