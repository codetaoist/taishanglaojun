package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// HybridRecommendationModel 
type HybridRecommendationModel struct {
	collaborativeModel *CollaborativeFilteringModel
	contentBasedModel  *ContentBasedModel
	deepLearningModel  *DeepLearningModel
	modelInfo          ModelInfo
	weights            HybridWeights
	strategy           string // "weighted", "switching", "cascade", "mixed"
}

// HybridWeights 
type HybridWeights struct {
	Collaborative float64 `json:"collaborative"`
	ContentBased  float64 `json:"content_based"`
	DeepLearning  float64 `json:"deep_learning"`
	Popularity    float64 `json:"popularity"`
	Diversity     float64 `json:"diversity"`
}

// DeepLearningModel 
type DeepLearningModel struct {
	neuralNetwork *NeuralNetwork
	modelInfo     ModelInfo
	isTrained     bool
}

// NeuralNetwork 
type NeuralNetwork struct {
	inputSize    int
	hiddenSizes  []int
	outputSize   int
	weights      [][][]float64
	biases       [][]float64
	activations  []string
	learningRate float64
	epochs       int
}

// NewHybridRecommendationModel 
func NewHybridRecommendationModel(strategy string) *HybridRecommendationModel {
	return &HybridRecommendationModel{
		collaborativeModel: NewCollaborativeFilteringModel(20, 0.1),
		contentBasedModel:  NewContentBasedModel(),
		deepLearningModel:  NewDeepLearningModel(),
		strategy:           strategy,
		weights: HybridWeights{
			Collaborative: 0.4,
			ContentBased:  0.3,
			DeepLearning:  0.2,
			Popularity:    0.05,
			Diversity:     0.05,
		},
		modelInfo: ModelInfo{
			Name:    "Hybrid Recommendation System",
			Version: "1.0",
			Type:    "hybrid",
		},
	}
}

// NewDeepLearningModel 
func NewDeepLearningModel() *DeepLearningModel {
	return &DeepLearningModel{
		neuralNetwork: &NeuralNetwork{
			inputSize:    100, // ?
			hiddenSizes:  []int{64, 32, 16},
			outputSize:   1, // 
			activations:  []string{"relu", "relu", "sigmoid"},
			learningRate: 0.001,
			epochs:       100,
		},
		modelInfo: ModelInfo{
			Name:    "Deep Learning Recommendation",
			Version: "1.0",
			Type:    "deep_learning",
		},
	}
}

// Train 
func (m *HybridRecommendationModel) Train(ctx context.Context, data *TrainingData) error {
	// 
	if err := m.collaborativeModel.Train(ctx, data); err != nil {
		return fmt.Errorf("failed to train collaborative model: %w", err)
	}
	
	// ?
	if err := m.contentBasedModel.Train(ctx, data); err != nil {
		return fmt.Errorf("failed to train content-based model: %w", err)
	}
	
	// 
	if err := m.deepLearningModel.Train(ctx, data); err != nil {
		return fmt.Errorf("failed to train deep learning model: %w", err)
	}
	
	// 
	m.optimizeHybridWeights(data)
	
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = m.evaluateHybridModel(data)
	
	return nil
}

// Predict 
func (m *HybridRecommendationModel) Predict(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	switch m.strategy {
	case "weighted":
		return m.weightedPrediction(ctx, userID, candidates)
	case "switching":
		return m.switchingPrediction(ctx, userID, candidates)
	case "cascade":
		return m.cascadePrediction(ctx, userID, candidates)
	case "mixed":
		return m.mixedPrediction(ctx, userID, candidates)
	default:
		return m.weightedPrediction(ctx, userID, candidates)
	}
}

// GetModelInfo 
func (m *HybridRecommendationModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// weightedPrediction 
func (m *HybridRecommendationModel) weightedPrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	// 
	collaborativePreds, err := m.collaborativeModel.Predict(ctx, userID, candidates)
	if err != nil {
		collaborativePreds = []Prediction{}
	}
	
	contentPreds, err := m.contentBasedModel.Predict(ctx, userID, candidates)
	if err != nil {
		contentPreds = []Prediction{}
	}
	
	deepLearningPreds, err := m.deepLearningModel.Predict(ctx, userID, candidates)
	if err != nil {
		deepLearningPreds = []Prediction{}
	}
	
	// 
	collabMap := make(map[uuid.UUID]Prediction)
	for _, pred := range collaborativePreds {
		collabMap[pred.ContentID] = pred
	}
	
	contentMap := make(map[uuid.UUID]Prediction)
	for _, pred := range contentPreds {
		contentMap[pred.ContentID] = pred
	}
	
	deepMap := make(map[uuid.UUID]Prediction)
	for _, pred := range deepLearningPreds {
		deepMap[pred.ContentID] = pred
	}
	
	// 
	predictions := make([]Prediction, 0, len(candidates))
	
	for _, contentID := range candidates {
		hybridScore := 0.0
		hybridConfidence := 0.0
		factors := make(map[string]float64)
		explanations := []string{}
		
		// 
		if collabPred, exists := collabMap[contentID]; exists {
			hybridScore += collabPred.Score * m.weights.Collaborative
			hybridConfidence += collabPred.Confidence * m.weights.Collaborative
			factors["collaborative"] = collabPred.Score
			if collabPred.Explanation != "" {
				explanations = append(explanations, collabPred.Explanation)
			}
		}
		
		// 
		if contentPred, exists := contentMap[contentID]; exists {
			hybridScore += contentPred.Score * m.weights.ContentBased
			hybridConfidence += contentPred.Confidence * m.weights.ContentBased
			factors["content_based"] = contentPred.Score
			if contentPred.Explanation != "" {
				explanations = append(explanations, contentPred.Explanation)
			}
		}
		
		// 
		if deepPred, exists := deepMap[contentID]; exists {
			hybridScore += deepPred.Score * m.weights.DeepLearning
			hybridConfidence += deepPred.Confidence * m.weights.DeepLearning
			factors["deep_learning"] = deepPred.Score
			if deepPred.Explanation != "" {
				explanations = append(explanations, deepPred.Explanation)
			}
		}
		
		// ?
		popularityScore := m.calculatePopularityScore(contentID)
		hybridScore += popularityScore * m.weights.Popularity
		factors["popularity"] = popularityScore
		
		// ?
		diversityScore := m.calculateDiversityScore(contentID, userID)
		hybridScore += diversityScore * m.weights.Diversity
		factors["diversity"] = diversityScore
		
		prediction := Prediction{
			ContentID:   contentID,
			Score:       math.Min(hybridScore, 1.0),
			Confidence:  math.Min(hybridConfidence, 1.0),
			Explanation: m.combineExplanations(explanations),
			Factors:     factors,
		}
		
		predictions = append(predictions, prediction)
	}
	
	// ?
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	// 
	predictions = m.diversityReranking(predictions, userID)
	
	return predictions, nil
}

// switchingPrediction 
func (m *HybridRecommendationModel) switchingPrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	// 
	userDataDensity := m.calculateUserDataDensity(userID)
	
	if userDataDensity > 0.7 {
		// ?
		return m.collaborativeModel.Predict(ctx, userID, candidates)
	} else if userDataDensity > 0.3 {
		// ?
		return m.deepLearningModel.Predict(ctx, userID, candidates)
	} else {
		// ?
		return m.contentBasedModel.Predict(ctx, userID, candidates)
	}
}

// cascadePrediction ?
func (m *HybridRecommendationModel) cascadePrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	// ?
	contentPreds, err := m.contentBasedModel.Predict(ctx, userID, candidates)
	if err != nil {
		return nil, err
	}
	
	// ?0%?
	topContentCandidates := make([]uuid.UUID, 0, len(contentPreds)/2)
	for i := 0; i < len(contentPreds)/2 && i < len(contentPreds); i++ {
		topContentCandidates = append(topContentCandidates, contentPreds[i].ContentID)
	}
	
	// 
	if len(topContentCandidates) > 0 {
		return m.collaborativeModel.Predict(ctx, userID, topContentCandidates)
	}
	
	return contentPreds, nil
}

// mixedPrediction 
func (m *HybridRecommendationModel) mixedPrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	predictions := make([]Prediction, 0, len(candidates))
	
	// ?
	groupSize := len(candidates) / 3
	
	// ?
	if groupSize > 0 {
		group1 := candidates[:groupSize]
		preds1, err := m.collaborativeModel.Predict(ctx, userID, group1)
		if err == nil {
			predictions = append(predictions, preds1...)
		}
	}
	
	// 
	if groupSize > 0 && len(candidates) > groupSize {
		group2 := candidates[groupSize : 2*groupSize]
		preds2, err := m.contentBasedModel.Predict(ctx, userID, group2)
		if err == nil {
			predictions = append(predictions, preds2...)
		}
	}
	
	// ?
	if len(candidates) > 2*groupSize {
		group3 := candidates[2*groupSize:]
		preds3, err := m.deepLearningModel.Predict(ctx, userID, group3)
		if err == nil {
			predictions = append(predictions, preds3...)
		}
	}
	
	// ?
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	return predictions, nil
}

// calculatePopularityScore ?
func (m *HybridRecommendationModel) calculatePopularityScore(contentID uuid.UUID) float64 {
	// ?
	// ?
	return 0.5
}

// calculateDiversityScore ?
func (m *HybridRecommendationModel) calculateDiversityScore(contentID uuid.UUID, userID uuid.UUID) float64 {
	// ?
	// ?
	return 0.3
}

// calculateUserDataDensity 
func (m *HybridRecommendationModel) calculateUserDataDensity(userID uuid.UUID) float64 {
	// ?
	// ?
	return 0.5
}

// combineExplanations 
func (m *HybridRecommendationModel) combineExplanations(explanations []string) string {
	if len(explanations) == 0 {
		return "㷨?
	}
	
	if len(explanations) == 1 {
		return explanations[0]
	}
	
	return fmt.Sprintf(": %s?, explanations[0])
}

// diversityReranking 
func (m *HybridRecommendationModel) diversityReranking(predictions []Prediction, userID uuid.UUID) []Prediction {
	if len(predictions) <= 1 {
		return predictions
	}
	
	// ?
	reranked := make([]Prediction, 0, len(predictions))
	used := make(map[uuid.UUID]bool)
	
	// ?
	reranked = append(reranked, predictions[0])
	used[predictions[0].ContentID] = true
	
	// ?
	for len(reranked) < len(predictions) {
		bestIdx := -1
		bestScore := -1.0
		
		for i, pred := range predictions {
			if used[pred.ContentID] {
				continue
			}
			
			// ?
			diversityBonus := m.calculateDiversityBonus(pred.ContentID, reranked)
			adjustedScore := pred.Score + diversityBonus*0.1
			
			if adjustedScore > bestScore {
				bestScore = adjustedScore
				bestIdx = i
			}
		}
		
		if bestIdx >= 0 {
			reranked = append(reranked, predictions[bestIdx])
			used[predictions[bestIdx].ContentID] = true
		} else {
			break
		}
	}
	
	return reranked
}

// calculateDiversityBonus ?
func (m *HybridRecommendationModel) calculateDiversityBonus(contentID uuid.UUID, selected []Prediction) float64 {
	// 轱
	return 0.1
}

// optimizeHybridWeights 
func (m *HybridRecommendationModel) optimizeHybridWeights(data *TrainingData) {
	// 
	bestAccuracy := 0.0
	bestWeights := m.weights
	
	steps := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6}
	
	for _, collab := range steps {
		for _, content := range steps {
			for _, deep := range steps {
				if collab+content+deep > 0.9 {
					continue
				}
				
				remaining := 1.0 - collab - content - deep
				
				testWeights := HybridWeights{
					Collaborative: collab,
					ContentBased:  content,
					DeepLearning:  deep,
					Popularity:    remaining * 0.3,
					Diversity:     remaining * 0.7,
				}
				
				// 
				originalWeights := m.weights
				m.weights = testWeights
				
				// ?
				accuracy := m.evaluateHybridModel(data)
				
				if accuracy > bestAccuracy {
					bestAccuracy = accuracy
					bestWeights = testWeights
				}
				
				// ?
				m.weights = originalWeights
			}
		}
	}
	
	// ?
	m.weights = bestWeights
}

// evaluateHybridModel 
func (m *HybridRecommendationModel) evaluateHybridModel(data *TrainingData) float64 {
	// 
	return (m.collaborativeModel.GetModelInfo().Accuracy*m.weights.Collaborative +
		m.contentBasedModel.GetModelInfo().Accuracy*m.weights.ContentBased +
		0.8*m.weights.DeepLearning) // 0.8
}

// Train 
func (m *DeepLearningModel) Train(ctx context.Context, data *TrainingData) error {
	// ?
	m.neuralNetwork.initializeWeights()
	
	// 
	trainX, trainY := m.prepareTrainingData(data)
	
	// 
	for epoch := 0; epoch < m.neuralNetwork.epochs; epoch++ {
		for i := 0; i < len(trainX); i++ {
			// 
			output := m.neuralNetwork.forward(trainX[i])
			
			// 
			m.neuralNetwork.backward(trainX[i], trainY[i], output)
		}
	}
	
	m.isTrained = true
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = 0.8 // ?
	
	return nil
}

// Predict 
func (m *DeepLearningModel) Predict(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	if !m.isTrained {
		return nil, fmt.Errorf("model not trained")
	}
	
	predictions := make([]Prediction, 0, len(candidates))
	
	for _, contentID := range candidates {
		// 
		input := m.buildInputFeatures(userID, contentID)
		
		// 
		output := m.neuralNetwork.forward(input)
		score := output[0]
		
		prediction := Prediction{
			ContentID:   contentID,
			Score:       math.Max(0, math.Min(score, 1.0)),
			Confidence:  0.7, // ?
			Explanation: "?,
			Factors: map[string]float64{
				"neural_network": score,
			},
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
func (m *DeepLearningModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// initializeWeights ?
func (nn *NeuralNetwork) initializeWeights() {
	// ?
	layers := []int{nn.inputSize}
	layers = append(layers, nn.hiddenSizes...)
	layers = append(layers, nn.outputSize)
	
	nn.weights = make([][][]float64, len(layers)-1)
	nn.biases = make([][]float64, len(layers)-1)
	
	for i := 0; i < len(layers)-1; i++ {
		nn.weights[i] = make([][]float64, layers[i])
		nn.biases[i] = make([]float64, layers[i+1])
		
		for j := 0; j < layers[i]; j++ {
			nn.weights[i][j] = make([]float64, layers[i+1])
			for k := 0; k < layers[i+1]; k++ {
				nn.weights[i][j][k] = (math.Mod(float64(i*j*k+1), 1000) - 500) / 1000.0 // ?
			}
		}
		
		for j := 0; j < layers[i+1]; j++ {
			nn.biases[i][j] = (math.Mod(float64(i*j+1), 1000) - 500) / 1000.0
		}
	}
}

// forward 
func (nn *NeuralNetwork) forward(input []float64) []float64 {
	current := input
	
	for layer := 0; layer < len(nn.weights); layer++ {
		next := make([]float64, len(nn.biases[layer]))
		
		// 
		for j := 0; j < len(next); j++ {
			sum := nn.biases[layer][j]
			for i := 0; i < len(current); i++ {
				sum += current[i] * nn.weights[layer][i][j]
			}
			
			// ?
			if layer < len(nn.activations) {
				switch nn.activations[layer] {
				case "relu":
					next[j] = math.Max(0, sum)
				case "sigmoid":
					next[j] = 1.0 / (1.0 + math.Exp(-sum))
				default:
					next[j] = sum
				}
			} else {
				next[j] = sum
			}
		}
		
		current = next
	}
	
	return current
}

// backward 
func (nn *NeuralNetwork) backward(input []float64, target float64, output []float64) {
	// 
	error := output[0] - target
	
	// 
	lastLayer := len(nn.weights) - 1
	for i := 0; i < len(nn.weights[lastLayer]); i++ {
		for j := 0; j < len(nn.weights[lastLayer][i]); j++ {
			gradient := error * input[i%len(input)]
			nn.weights[lastLayer][i][j] -= nn.learningRate * gradient
		}
	}
}

// prepareTrainingData 
func (m *DeepLearningModel) prepareTrainingData(data *TrainingData) ([][]float64, []float64) {
	trainX := make([][]float64, 0, len(data.Ratings))
	trainY := make([]float64, 0, len(data.Ratings))
	
	for _, rating := range data.Ratings {
		input := m.buildInputFeatures(rating.UserID, rating.ContentID)
		trainX = append(trainX, input)
		trainY = append(trainY, rating.Rating/5.0) // 0-1
	}
	
	return trainX, trainY
}

// buildInputFeatures 
func (m *DeepLearningModel) buildInputFeatures(userID, contentID uuid.UUID) []float64 {
	// 
	features := make([]float64, m.neuralNetwork.inputSize)
	
	// ID
	userHash := float64(userID.ID() % 1000) / 1000.0
	features[0] = userHash
	
	// ID
	contentHash := float64(contentID.ID() % 1000) / 1000.0
	features[1] = contentHash
	
	// 
	for i := 2; i < len(features); i++ {
		features[i] = math.Mod(float64(i)*userHash*contentHash, 1.0)
	}
	
	return features
}

