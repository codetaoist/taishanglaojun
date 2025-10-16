﻿package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContentBasedModel ?
type ContentBasedModel struct {
	contentFeatures map[uuid.UUID]*ContentProfile
	userProfiles    map[uuid.UUID]*UserProfile
	modelInfo       ModelInfo
	weights         map[string]float64
}

// ContentProfile 
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

// UserProfile 
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

// NewContentBasedModel ?
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

// Train ?
func (m *ContentBasedModel) Train(ctx context.Context, data *TrainingData) error {
	// 
	m.buildContentProfiles(data)
	
	// 
	m.buildUserProfiles(data)
	
	// 
	m.optimizeWeights(data)
	
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = m.evaluateModel(data)
	
	return nil
}

// Predict 
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
	
	// ?
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	return predictions, nil
}

// GetModelInfo 
func (m *ContentBasedModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// buildContentProfiles 
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

// buildUserProfiles 
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
		
		// 
		m.buildUserPreferencesFromInteractions(profile, userFeature.UserID, data)
		
		m.userProfiles[userFeature.UserID] = profile
	}
}

// buildUserPreferencesFromInteractions 
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
		
		// 
		for _, category := range contentProfile.Categories {
			categoryWeights[category] += weight
		}
		
		// 
		for _, tag := range contentProfile.Tags {
			tagWeights[tag] += weight
		}
		
		// ?
		for _, keyword := range contentProfile.Keywords {
			keywordWeights[keyword] += weight
		}
	}
	
	// ?
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

// calculateInteractionWeight 㽻
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
	
	// 
	daysSince := time.Since(interaction.Timestamp).Hours() / 24
	timeDecay := math.Exp(-daysSince / 30) // 30
	
	// 
	durationBonus := 1.0
	if interaction.Duration > 0 {
		durationBonus = math.Min(1.0+float64(interaction.Duration)/3600.0, 2.0)
	}
	
	return baseWeight * timeDecay * durationBonus
}

// calculateContentSimilarity ?
func (m *ContentBasedModel) calculateContentSimilarity(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	var totalScore float64
	
	// 
	categoryScore := m.calculateCategoryMatch(userProfile, contentProfile)
	totalScore += categoryScore * m.weights["category"]
	
	// 
	tagScore := m.calculateTagMatch(userProfile, contentProfile)
	totalScore += tagScore * m.weights["tag"]
	
	// ?
	keywordScore := m.calculateKeywordMatch(userProfile, contentProfile)
	totalScore += keywordScore * m.weights["keyword"]
	
	// 
	difficultyScore := m.calculateDifficultyMatch(userProfile, contentProfile)
	totalScore += difficultyScore * m.weights["difficulty"]
	
	// 
	durationScore := m.calculateDurationMatch(userProfile, contentProfile)
	totalScore += durationScore * m.weights["duration"]
	
	// 
	qualityScore := contentProfile.Quality
	totalScore += qualityScore * m.weights["quality"]
	
	return math.Min(totalScore, 1.0)
}

// calculateCategoryMatch ?
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

// calculateTagMatch ?
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

// calculateKeywordMatch 
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

// calculateDifficultyMatch ?
func (m *ContentBasedModel) calculateDifficultyMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	diff := math.Abs(userProfile.Difficulty - contentProfile.Difficulty)
	return math.Exp(-diff) // 
}

// calculateDurationMatch ?
func (m *ContentBasedModel) calculateDurationMatch(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	if userProfile.Duration == 0 {
		return 0.5 // 
	}
	
	ratio := float64(contentProfile.Duration) / float64(userProfile.Duration)
	if ratio > 1 {
		ratio = 1 / ratio
	}
	
	return ratio
}

// calculateContentConfidence ?
func (m *ContentBasedModel) calculateContentConfidence(userProfile *UserProfile, contentProfile *ContentProfile) float64 {
	// ?
	featureCoverage := 0.0
	totalFeatures := 0.0
	
	// 
	if len(contentProfile.Categories) > 0 {
		totalFeatures++
		for _, category := range contentProfile.Categories {
			if _, exists := userProfile.Categories[category]; exists {
				featureCoverage++
				break
			}
		}
	}
	
	// 
	if len(contentProfile.Tags) > 0 {
		totalFeatures++
		for _, tag := range contentProfile.Tags {
			if _, exists := userProfile.Tags[tag]; exists {
				featureCoverage++
				break
			}
		}
	}
	
	// ?
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

// generateContentExplanation 
func (m *ContentBasedModel) generateContentExplanation(userProfile *UserProfile, contentProfile *ContentProfile) string {
	reasons := []string{}
	
	// ?
	for _, category := range contentProfile.Categories {
		if weight, exists := userProfile.Categories[category]; exists && weight > 0.3 {
			reasons = append(reasons, fmt.Sprintf("%s", category))
			break
		}
	}
	
	// ?
	matchedTags := []string{}
	for _, tag := range contentProfile.Tags {
		if weight, exists := userProfile.Tags[tag]; exists && weight > 0.2 {
			matchedTags = append(matchedTags, tag)
		}
	}
	if len(matchedTags) > 0 {
		reasons = append(reasons, fmt.Sprintf(": %s", strings.Join(matchedTags[:min(3, len(matchedTags))], ", ")))
	}
	
	// ?
	diffMatch := m.calculateDifficultyMatch(userProfile, contentProfile)
	if diffMatch > 0.8 {
		reasons = append(reasons, "")
	}
	
	if len(reasons) == 0 {
		return ""
	}
	
	return strings.Join(reasons, "; ")
}

// getContentFactors 
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

// extractKeywords ?
func (m *ContentBasedModel) extractKeywords(feature ContentFeature) []string {
	keywords := []string{}
	
	// 
	keywords = append(keywords, feature.Tags...)
	
	// ?
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

// extractFromText ?
func (m *ContentBasedModel) extractFromText(text string) []string {
	// NLP
	words := strings.Fields(strings.ToLower(text))
	keywords := []string{}
	
	for _, word := range words {
		if len(word) > 3 && !m.isStopWord(word) {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// isStopWord ?
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

// deduplicateKeywords ?
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

// calculateContentQuality 
func (m *ContentBasedModel) calculateContentQuality(feature ContentFeature, data *TrainingData) float64 {
	// ?
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
	
	avgRating := 0.5 // 
	if ratingCount > 0 {
		avgRating = totalRating / float64(ratingCount) / 5.0 // 0-1
	}
	
	popularity := math.Log(float64(interactionCount)+1) / 10.0 // ?
	
	return math.Min(avgRating+popularity*0.1, 1.0)
}

// calculatePreferredDifficulty 
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
		return 0.5 // 
	}
	
	return totalDifficulty / float64(count)
}

// calculatePreferredDuration 
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
		return 1800 // 30
	}
	
	return totalDuration / int64(count)
}

// findContentFeature 
func (m *ContentBasedModel) findContentFeature(contentID uuid.UUID, data *TrainingData) *ContentFeature {
	for _, feature := range data.ContentFeatures {
		if feature.ContentID == contentID {
			return &feature
		}
	}
	return nil
}

// optimizeWeights 
func (m *ContentBasedModel) optimizeWeights(data *TrainingData) {
	// 㷨?
	bestAccuracy := 0.0
	bestWeights := make(map[string]float64)
	
	// 
	for k, v := range m.weights {
		bestWeights[k] = v
	}
	
	// ?
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
				
				// 
				originalWeights := m.weights
				m.weights = testWeights
				
				// ?
				accuracy := m.evaluateModel(data)
				
				if accuracy > bestAccuracy {
					bestAccuracy = accuracy
					for k, v := range testWeights {
						bestWeights[k] = v
					}
				}
				
				// ?
				m.weights = originalWeights
			}
		}
	}
	
	// ?
	m.weights = bestWeights
}

// evaluateModel 
func (m *ContentBasedModel) evaluateModel(data *TrainingData) float64 {
	// 
	correct := 0
	total := 0
	
	for _, rating := range data.Ratings {
		userProfile := m.userProfiles[rating.UserID]
		contentProfile := m.contentFeatures[rating.ContentID]
		
		if userProfile != nil && contentProfile != nil {
			predicted := m.calculateContentSimilarity(userProfile, contentProfile) * 5.0
			actual := rating.Rating
			
			// 1
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

// min 
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

