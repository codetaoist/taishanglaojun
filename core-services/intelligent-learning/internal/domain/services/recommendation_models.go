package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// RecommendationModel 
type RecommendationModel interface {
	Train(ctx context.Context, data *TrainingData) error
	Predict(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error)
	GetModelInfo() ModelInfo
}

// TrainingData 
type TrainingData struct {
	UserInteractions []UserInteraction `json:"user_interactions"`
	ContentFeatures  []ContentFeature  `json:"content_features"`
	UserFeatures     []UserFeature     `json:"user_features"`
	Ratings          []Rating          `json:"ratings"`
	ImplicitFeedback []ImplicitFeedback `json:"implicit_feedback"`
}

// UserInteraction 
type UserInteraction struct {
	UserID      uuid.UUID `json:"user_id"`
	ContentID   uuid.UUID `json:"content_id"`
	Interaction string    `json:"interaction"` // "view", "like", "complete", "share"
	Duration    int64     `json:"duration"`    // ?
	Timestamp   time.Time `json:"timestamp"`
	Rating      float64   `json:"rating"`      // 
	Context     map[string]interface{} `json:"context"`
}

// ContentFeature 
type ContentFeature struct {
	ContentID    uuid.UUID              `json:"content_id"`
	Features     map[string]float64     `json:"features"`
	Categories   []string               `json:"categories"`
	Tags         []string               `json:"tags"`
	Difficulty   float64                `json:"difficulty"`
	Duration     int64                  `json:"duration"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// UserFeature 
type UserFeature struct {
	UserID       uuid.UUID              `json:"user_id"`
	Demographics map[string]interface{} `json:"demographics"`
	Preferences  map[string]float64     `json:"preferences"`
	Skills       map[string]float64     `json:"skills"`
	Behavior     map[string]float64     `json:"behavior"`
	LearningStyle string                `json:"learning_style"`
}

// Rating 
type Rating struct {
	UserID    uuid.UUID `json:"user_id"`
	ContentID uuid.UUID `json:"content_id"`
	Rating    float64   `json:"rating"`
	Timestamp time.Time `json:"timestamp"`
}

// ImplicitFeedback 
type ImplicitFeedback struct {
	UserID      uuid.UUID              `json:"user_id"`
	ContentID   uuid.UUID              `json:"content_id"`
	Action      string                 `json:"action"` // "click", "view", "download", "bookmark"
	Confidence  float64                `json:"confidence"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
}

// Prediction 
type Prediction struct {
	ContentID   uuid.UUID `json:"content_id"`
	Score       float64   `json:"score"`
	Confidence  float64   `json:"confidence"`
	Explanation string    `json:"explanation"`
	Factors     map[string]float64 `json:"factors"`
}

// ModelInfo 
type ModelInfo struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Type        string    `json:"type"`
	TrainedAt   time.Time `json:"trained_at"`
	Accuracy    float64   `json:"accuracy"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// CollaborativeFilteringModel 
type CollaborativeFilteringModel struct {
	userItemMatrix map[uuid.UUID]map[uuid.UUID]float64
	userSimilarity map[uuid.UUID]map[uuid.UUID]float64
	itemSimilarity map[uuid.UUID]map[uuid.UUID]float64
	modelInfo      ModelInfo
	k              int     // 
	threshold      float64 // ?
}

// NewCollaborativeFilteringModel 
func NewCollaborativeFilteringModel(k int, threshold float64) *CollaborativeFilteringModel {
	return &CollaborativeFilteringModel{
		userItemMatrix: make(map[uuid.UUID]map[uuid.UUID]float64),
		userSimilarity: make(map[uuid.UUID]map[uuid.UUID]float64),
		itemSimilarity: make(map[uuid.UUID]map[uuid.UUID]float64),
		k:              k,
		threshold:      threshold,
		modelInfo: ModelInfo{
			Name:    "Collaborative Filtering",
			Version: "1.0",
			Type:    "collaborative",
		},
	}
}

// Train 
func (m *CollaborativeFilteringModel) Train(ctx context.Context, data *TrainingData) error {
	// -
	m.buildUserItemMatrix(data)
	
	// ?
	m.calculateUserSimilarity()
	
	// ?
	m.calculateItemSimilarity()
	
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = m.evaluateModel(data)
	
	return nil
}

// Predict 
func (m *CollaborativeFilteringModel) Predict(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	predictions := make([]Prediction, 0, len(candidates))
	
	for _, contentID := range candidates {
		score := m.predictUserItemScore(userID, contentID)
		confidence := m.calculateConfidence(userID, contentID)
		
		prediction := Prediction{
			ContentID:   contentID,
			Score:       score,
			Confidence:  confidence,
			Explanation: m.generateExplanation(userID, contentID),
			Factors:     m.getFactors(userID, contentID),
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
func (m *CollaborativeFilteringModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// buildUserItemMatrix -
func (m *CollaborativeFilteringModel) buildUserItemMatrix(data *TrainingData) {
	for _, interaction := range data.UserInteractions {
		if m.userItemMatrix[interaction.UserID] == nil {
			m.userItemMatrix[interaction.UserID] = make(map[uuid.UUID]float64)
		}
		
		// 
		score := m.calculateImplicitRating(interaction)
		m.userItemMatrix[interaction.UserID][interaction.ContentID] = score
	}
	
	// 
	for _, rating := range data.Ratings {
		if m.userItemMatrix[rating.UserID] == nil {
			m.userItemMatrix[rating.UserID] = make(map[uuid.UUID]float64)
		}
		m.userItemMatrix[rating.UserID][rating.ContentID] = rating.Rating
	}
}

// calculateImplicitRating 
func (m *CollaborativeFilteringModel) calculateImplicitRating(interaction UserInteraction) float64 {
	baseScore := 0.0
	
	switch interaction.Interaction {
	case "view":
		baseScore = 1.0
	case "like":
		baseScore = 3.0
	case "complete":
		baseScore = 4.0
	case "share":
		baseScore = 5.0
	}
	
	// 
	if interaction.Duration > 0 {
		durationFactor := math.Min(float64(interaction.Duration)/3600.0, 2.0) // ??
		baseScore *= (1.0 + durationFactor*0.5)
	}
	
	return math.Min(baseScore, 5.0)
}

// calculateUserSimilarity ?
func (m *CollaborativeFilteringModel) calculateUserSimilarity() {
	users := make([]uuid.UUID, 0, len(m.userItemMatrix))
	for userID := range m.userItemMatrix {
		users = append(users, userID)
	}
	
	for i, userA := range users {
		if m.userSimilarity[userA] == nil {
			m.userSimilarity[userA] = make(map[uuid.UUID]float64)
		}
		
		for j, userB := range users {
			if i != j {
				similarity := m.cosineSimilarity(m.userItemMatrix[userA], m.userItemMatrix[userB])
				if similarity > m.threshold {
					m.userSimilarity[userA][userB] = similarity
				}
			}
		}
	}
}

// calculateItemSimilarity ?
func (m *CollaborativeFilteringModel) calculateItemSimilarity() {
	// ?
	itemUserMatrix := make(map[uuid.UUID]map[uuid.UUID]float64)
	
	for userID, items := range m.userItemMatrix {
		for itemID, rating := range items {
			if itemUserMatrix[itemID] == nil {
				itemUserMatrix[itemID] = make(map[uuid.UUID]float64)
			}
			itemUserMatrix[itemID][userID] = rating
		}
	}
	
	// 
	items := make([]uuid.UUID, 0, len(itemUserMatrix))
	for itemID := range itemUserMatrix {
		items = append(items, itemID)
	}
	
	for i, itemA := range items {
		if m.itemSimilarity[itemA] == nil {
			m.itemSimilarity[itemA] = make(map[uuid.UUID]float64)
		}
		
		for j, itemB := range items {
			if i != j {
				similarity := m.cosineSimilarity(itemUserMatrix[itemA], itemUserMatrix[itemB])
				if similarity > m.threshold {
					m.itemSimilarity[itemA][itemB] = similarity
				}
			}
		}
	}
}

// cosineSimilarity ?
func (m *CollaborativeFilteringModel) cosineSimilarity(vectorA, vectorB map[uuid.UUID]float64) float64 {
	var dotProduct, normA, normB float64
	
	// 
	for key, valueA := range vectorA {
		if valueB, exists := vectorB[key]; exists {
			dotProduct += valueA * valueB
		}
		normA += valueA * valueA
	}
	
	for _, valueB := range vectorB {
		normB += valueB * valueB
	}
	
	if normA == 0 || normB == 0 {
		return 0
	}
	
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// predictUserItemScore 
func (m *CollaborativeFilteringModel) predictUserItemScore(userID, itemID uuid.UUID) float64 {
	// ?
	userBasedScore := m.userBasedPrediction(userID, itemID)
	
	// ?
	itemBasedScore := m.itemBasedPrediction(userID, itemID)
	
	// 
	return (userBasedScore + itemBasedScore) / 2.0
}

// userBasedPrediction ?
func (m *CollaborativeFilteringModel) userBasedPrediction(userID, itemID uuid.UUID) float64 {
	similarUsers := m.getTopKSimilarUsers(userID, m.k)
	
	var weightedSum, similaritySum float64
	
	for _, similarUser := range similarUsers {
		if rating, exists := m.userItemMatrix[similarUser.ID][itemID]; exists {
			weightedSum += similarUser.Similarity * rating
			similaritySum += math.Abs(similarUser.Similarity)
		}
	}
	
	if similaritySum == 0 {
		return 0
	}
	
	return weightedSum / similaritySum
}

// itemBasedPrediction ?
func (m *CollaborativeFilteringModel) itemBasedPrediction(userID, itemID uuid.UUID) float64 {
	userRatings := m.userItemMatrix[userID]
	if userRatings == nil {
		return 0
	}
	
	var weightedSum, similaritySum float64
	
	for ratedItemID, rating := range userRatings {
		if similarity, exists := m.itemSimilarity[itemID][ratedItemID]; exists {
			weightedSum += similarity * rating
			similaritySum += math.Abs(similarity)
		}
	}
	
	if similaritySum == 0 {
		return 0
	}
	
	return weightedSum / similaritySum
}

// SimilarUser 
type SimilarUser struct {
	ID         uuid.UUID
	Similarity float64
}

// getTopKSimilarUsers K?
func (m *CollaborativeFilteringModel) getTopKSimilarUsers(userID uuid.UUID, k int) []SimilarUser {
	similarities := m.userSimilarity[userID]
	if similarities == nil {
		return []SimilarUser{}
	}
	
	users := make([]SimilarUser, 0, len(similarities))
	for similarUserID, similarity := range similarities {
		users = append(users, SimilarUser{
			ID:         similarUserID,
			Similarity: similarity,
		})
	}
	
	// 
	sort.Slice(users, func(i, j int) bool {
		return users[i].Similarity > users[j].Similarity
	})
	
	if len(users) > k {
		users = users[:k]
	}
	
	return users
}

// calculateConfidence ?
func (m *CollaborativeFilteringModel) calculateConfidence(userID, itemID uuid.UUID) float64 {
	// 
	similarUsers := m.getTopKSimilarUsers(userID, m.k)
	
	if len(similarUsers) == 0 {
		return 0.1
	}
	
	var avgSimilarity float64
	ratedCount := 0
	
	for _, user := range similarUsers {
		avgSimilarity += user.Similarity
		if _, exists := m.userItemMatrix[user.ID][itemID]; exists {
			ratedCount++
		}
	}
	
	avgSimilarity /= float64(len(similarUsers))
	coverage := float64(ratedCount) / float64(len(similarUsers))
	
	return avgSimilarity * coverage
}

// generateExplanation 
func (m *CollaborativeFilteringModel) generateExplanation(userID, itemID uuid.UUID) string {
	similarUsers := m.getTopKSimilarUsers(userID, 3)
	
	if len(similarUsers) == 0 {
		return ""
	}
	
	return fmt.Sprintf("?%d ", len(similarUsers))
}

// getFactors 
func (m *CollaborativeFilteringModel) getFactors(userID, itemID uuid.UUID) map[string]float64 {
	factors := make(map[string]float64)
	
	userBasedScore := m.userBasedPrediction(userID, itemID)
	itemBasedScore := m.itemBasedPrediction(userID, itemID)
	
	factors["user_based"] = userBasedScore
	factors["item_based"] = itemBasedScore
	factors["popularity"] = m.calculatePopularity(itemID)
	
	return factors
}

// calculatePopularity ?
func (m *CollaborativeFilteringModel) calculatePopularity(itemID uuid.UUID) float64 {
	count := 0
	var totalRating float64
	
	for _, userRatings := range m.userItemMatrix {
		if rating, exists := userRatings[itemID]; exists {
			count++
			totalRating += rating
		}
	}
	
	if count == 0 {
		return 0
	}
	
	avgRating := totalRating / float64(count)
	popularity := math.Log(float64(count)+1) * avgRating / 5.0
	
	return math.Min(popularity, 1.0)
}

// evaluateModel 
func (m *CollaborativeFilteringModel) evaluateModel(data *TrainingData) float64 {
	// RMSE
	var totalError float64
	count := 0
	
	for _, rating := range data.Ratings {
		predicted := m.predictUserItemScore(rating.UserID, rating.ContentID)
		error := rating.Rating - predicted
		totalError += error * error
		count++
	}
	
	if count == 0 {
		return 0
	}
	
	rmse := math.Sqrt(totalError / float64(count))
	return math.Max(0, 1.0-rmse/5.0) // 
}

