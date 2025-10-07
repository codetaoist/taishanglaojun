package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// HybridRecommendationModel 混合推荐模型
type HybridRecommendationModel struct {
	collaborativeModel *CollaborativeFilteringModel
	contentBasedModel  *ContentBasedModel
	deepLearningModel  *DeepLearningModel
	modelInfo          ModelInfo
	weights            HybridWeights
	strategy           string // "weighted", "switching", "cascade", "mixed"
}

// HybridWeights 混合权重
type HybridWeights struct {
	Collaborative float64 `json:"collaborative"`
	ContentBased  float64 `json:"content_based"`
	DeepLearning  float64 `json:"deep_learning"`
	Popularity    float64 `json:"popularity"`
	Diversity     float64 `json:"diversity"`
}

// DeepLearningModel 深度学习模型接口
type DeepLearningModel struct {
	neuralNetwork *NeuralNetwork
	modelInfo     ModelInfo
	isTrained     bool
}

// NeuralNetwork 神经网络结构
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

// NewHybridRecommendationModel 创建混合推荐模型
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

// NewDeepLearningModel 创建深度学习模型
func NewDeepLearningModel() *DeepLearningModel {
	return &DeepLearningModel{
		neuralNetwork: &NeuralNetwork{
			inputSize:    100, // 用户和物品特征维度
			hiddenSizes:  []int{64, 32, 16},
			outputSize:   1, // 预测评分
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

// Train 训练混合模型
func (m *HybridRecommendationModel) Train(ctx context.Context, data *TrainingData) error {
	// 训练协同过滤模型
	if err := m.collaborativeModel.Train(ctx, data); err != nil {
		return fmt.Errorf("failed to train collaborative model: %w", err)
	}
	
	// 训练基于内容的模型
	if err := m.contentBasedModel.Train(ctx, data); err != nil {
		return fmt.Errorf("failed to train content-based model: %w", err)
	}
	
	// 训练深度学习模型
	if err := m.deepLearningModel.Train(ctx, data); err != nil {
		return fmt.Errorf("failed to train deep learning model: %w", err)
	}
	
	// 优化混合权重
	m.optimizeHybridWeights(data)
	
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = m.evaluateHybridModel(data)
	
	return nil
}

// Predict 混合预测
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

// GetModelInfo 获取模型信息
func (m *HybridRecommendationModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// weightedPrediction 加权预测
func (m *HybridRecommendationModel) weightedPrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	// 获取各模型的预测结果
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
	
	// 创建预测映射
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
	
	// 混合预测
	predictions := make([]Prediction, 0, len(candidates))
	
	for _, contentID := range candidates {
		hybridScore := 0.0
		hybridConfidence := 0.0
		factors := make(map[string]float64)
		explanations := []string{}
		
		// 协同过滤分数
		if collabPred, exists := collabMap[contentID]; exists {
			hybridScore += collabPred.Score * m.weights.Collaborative
			hybridConfidence += collabPred.Confidence * m.weights.Collaborative
			factors["collaborative"] = collabPred.Score
			if collabPred.Explanation != "" {
				explanations = append(explanations, collabPred.Explanation)
			}
		}
		
		// 基于内容分数
		if contentPred, exists := contentMap[contentID]; exists {
			hybridScore += contentPred.Score * m.weights.ContentBased
			hybridConfidence += contentPred.Confidence * m.weights.ContentBased
			factors["content_based"] = contentPred.Score
			if contentPred.Explanation != "" {
				explanations = append(explanations, contentPred.Explanation)
			}
		}
		
		// 深度学习分数
		if deepPred, exists := deepMap[contentID]; exists {
			hybridScore += deepPred.Score * m.weights.DeepLearning
			hybridConfidence += deepPred.Confidence * m.weights.DeepLearning
			factors["deep_learning"] = deepPred.Score
			if deepPred.Explanation != "" {
				explanations = append(explanations, deepPred.Explanation)
			}
		}
		
		// 流行度分数
		popularityScore := m.calculatePopularityScore(contentID)
		hybridScore += popularityScore * m.weights.Popularity
		factors["popularity"] = popularityScore
		
		// 多样性分数
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
	
	// 按分数排序
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	// 应用多样性重排序
	predictions = m.diversityReranking(predictions, userID)
	
	return predictions, nil
}

// switchingPrediction 切换预测（根据情况选择最佳模型）
func (m *HybridRecommendationModel) switchingPrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	// 根据用户特征和数据稀疏性选择模型
	userDataDensity := m.calculateUserDataDensity(userID)
	
	if userDataDensity > 0.7 {
		// 数据充足，使用协同过滤
		return m.collaborativeModel.Predict(ctx, userID, candidates)
	} else if userDataDensity > 0.3 {
		// 数据中等，使用深度学习
		return m.deepLearningModel.Predict(ctx, userID, candidates)
	} else {
		// 数据稀疏，使用基于内容的推荐
		return m.contentBasedModel.Predict(ctx, userID, candidates)
	}
}

// cascadePrediction 级联预测（逐级过滤）
func (m *HybridRecommendationModel) cascadePrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	// 第一级：基于内容的过滤
	contentPreds, err := m.contentBasedModel.Predict(ctx, userID, candidates)
	if err != nil {
		return nil, err
	}
	
	// 选择前50%的候选
	topContentCandidates := make([]uuid.UUID, 0, len(contentPreds)/2)
	for i := 0; i < len(contentPreds)/2 && i < len(contentPreds); i++ {
		topContentCandidates = append(topContentCandidates, contentPreds[i].ContentID)
	}
	
	// 第二级：协同过滤精排
	if len(topContentCandidates) > 0 {
		return m.collaborativeModel.Predict(ctx, userID, topContentCandidates)
	}
	
	return contentPreds, nil
}

// mixedPrediction 混合预测（不同候选使用不同模型）
func (m *HybridRecommendationModel) mixedPrediction(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	predictions := make([]Prediction, 0, len(candidates))
	
	// 将候选分成三组
	groupSize := len(candidates) / 3
	
	// 第一组使用协同过滤
	if groupSize > 0 {
		group1 := candidates[:groupSize]
		preds1, err := m.collaborativeModel.Predict(ctx, userID, group1)
		if err == nil {
			predictions = append(predictions, preds1...)
		}
	}
	
	// 第二组使用基于内容的推荐
	if groupSize > 0 && len(candidates) > groupSize {
		group2 := candidates[groupSize : 2*groupSize]
		preds2, err := m.contentBasedModel.Predict(ctx, userID, group2)
		if err == nil {
			predictions = append(predictions, preds2...)
		}
	}
	
	// 第三组使用深度学习
	if len(candidates) > 2*groupSize {
		group3 := candidates[2*groupSize:]
		preds3, err := m.deepLearningModel.Predict(ctx, userID, group3)
		if err == nil {
			predictions = append(predictions, preds3...)
		}
	}
	
	// 按分数排序
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	return predictions, nil
}

// calculatePopularityScore 计算流行度分数
func (m *HybridRecommendationModel) calculatePopularityScore(contentID uuid.UUID) float64 {
	// 这里应该从数据库或缓存中获取流行度数据
	// 简化实现，返回随机值
	return 0.5
}

// calculateDiversityScore 计算多样性分数
func (m *HybridRecommendationModel) calculateDiversityScore(contentID uuid.UUID, userID uuid.UUID) float64 {
	// 这里应该计算与用户历史推荐的多样性
	// 简化实现，返回随机值
	return 0.3
}

// calculateUserDataDensity 计算用户数据密度
func (m *HybridRecommendationModel) calculateUserDataDensity(userID uuid.UUID) float64 {
	// 这里应该计算用户的交互数据密度
	// 简化实现，返回随机值
	return 0.5
}

// combineExplanations 合并解释
func (m *HybridRecommendationModel) combineExplanations(explanations []string) string {
	if len(explanations) == 0 {
		return "基于多种算法的综合推荐"
	}
	
	if len(explanations) == 1 {
		return explanations[0]
	}
	
	return fmt.Sprintf("综合考虑: %s等因素", explanations[0])
}

// diversityReranking 多样性重排序
func (m *HybridRecommendationModel) diversityReranking(predictions []Prediction, userID uuid.UUID) []Prediction {
	if len(predictions) <= 1 {
		return predictions
	}
	
	// 简单的多样性重排序：确保前几个推荐有一定的多样性
	reranked := make([]Prediction, 0, len(predictions))
	used := make(map[uuid.UUID]bool)
	
	// 先选择最高分的
	reranked = append(reranked, predictions[0])
	used[predictions[0].ContentID] = true
	
	// 然后在剩余的中选择多样性高的
	for len(reranked) < len(predictions) {
		bestIdx := -1
		bestScore := -1.0
		
		for i, pred := range predictions {
			if used[pred.ContentID] {
				continue
			}
			
			// 计算多样性调整后的分数
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

// calculateDiversityBonus 计算多样性奖励
func (m *HybridRecommendationModel) calculateDiversityBonus(contentID uuid.UUID, selected []Prediction) float64 {
	// 简化实现：如果与已选择的内容不同，给予奖励
	return 0.1
}

// optimizeHybridWeights 优化混合权重
func (m *HybridRecommendationModel) optimizeHybridWeights(data *TrainingData) {
	// 使用网格搜索优化权重
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
				
				// 临时设置权重
				originalWeights := m.weights
				m.weights = testWeights
				
				// 评估准确率
				accuracy := m.evaluateHybridModel(data)
				
				if accuracy > bestAccuracy {
					bestAccuracy = accuracy
					bestWeights = testWeights
				}
				
				// 恢复原权重
				m.weights = originalWeights
			}
		}
	}
	
	// 设置最佳权重
	m.weights = bestWeights
}

// evaluateHybridModel 评估混合模型
func (m *HybridRecommendationModel) evaluateHybridModel(data *TrainingData) float64 {
	// 简化的评估实现
	return (m.collaborativeModel.GetModelInfo().Accuracy*m.weights.Collaborative +
		m.contentBasedModel.GetModelInfo().Accuracy*m.weights.ContentBased +
		0.8*m.weights.DeepLearning) // 假设深度学习模型准确率为0.8
}

// Train 训练深度学习模型
func (m *DeepLearningModel) Train(ctx context.Context, data *TrainingData) error {
	// 初始化神经网络
	m.neuralNetwork.initializeWeights()
	
	// 准备训练数据
	trainX, trainY := m.prepareTrainingData(data)
	
	// 训练神经网络
	for epoch := 0; epoch < m.neuralNetwork.epochs; epoch++ {
		for i := 0; i < len(trainX); i++ {
			// 前向传播
			output := m.neuralNetwork.forward(trainX[i])
			
			// 反向传播
			m.neuralNetwork.backward(trainX[i], trainY[i], output)
		}
	}
	
	m.isTrained = true
	m.modelInfo.TrainedAt = time.Now()
	m.modelInfo.Accuracy = 0.8 // 简化的准确率
	
	return nil
}

// Predict 深度学习预测
func (m *DeepLearningModel) Predict(ctx context.Context, userID uuid.UUID, candidates []uuid.UUID) ([]Prediction, error) {
	if !m.isTrained {
		return nil, fmt.Errorf("model not trained")
	}
	
	predictions := make([]Prediction, 0, len(candidates))
	
	for _, contentID := range candidates {
		// 构建输入特征
		input := m.buildInputFeatures(userID, contentID)
		
		// 预测
		output := m.neuralNetwork.forward(input)
		score := output[0]
		
		prediction := Prediction{
			ContentID:   contentID,
			Score:       math.Max(0, math.Min(score, 1.0)),
			Confidence:  0.7, // 简化的置信度
			Explanation: "基于深度学习模型的推荐",
			Factors: map[string]float64{
				"neural_network": score,
			},
		}
		
		predictions = append(predictions, prediction)
	}
	
	// 按分数排序
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	return predictions, nil
}

// GetModelInfo 获取深度学习模型信息
func (m *DeepLearningModel) GetModelInfo() ModelInfo {
	return m.modelInfo
}

// initializeWeights 初始化权重
func (nn *NeuralNetwork) initializeWeights() {
	// 简化的权重初始化
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
				nn.weights[i][j][k] = (math.Mod(float64(i*j*k+1), 1000) - 500) / 1000.0 // 简化的随机初始化
			}
		}
		
		for j := 0; j < layers[i+1]; j++ {
			nn.biases[i][j] = (math.Mod(float64(i*j+1), 1000) - 500) / 1000.0
		}
	}
}

// forward 前向传播
func (nn *NeuralNetwork) forward(input []float64) []float64 {
	current := input
	
	for layer := 0; layer < len(nn.weights); layer++ {
		next := make([]float64, len(nn.biases[layer]))
		
		// 矩阵乘法
		for j := 0; j < len(next); j++ {
			sum := nn.biases[layer][j]
			for i := 0; i < len(current); i++ {
				sum += current[i] * nn.weights[layer][i][j]
			}
			
			// 激活函数
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

// backward 反向传播（简化实现）
func (nn *NeuralNetwork) backward(input []float64, target float64, output []float64) {
	// 简化的反向传播实现
	error := output[0] - target
	
	// 更新最后一层的权重（简化）
	lastLayer := len(nn.weights) - 1
	for i := 0; i < len(nn.weights[lastLayer]); i++ {
		for j := 0; j < len(nn.weights[lastLayer][i]); j++ {
			gradient := error * input[i%len(input)]
			nn.weights[lastLayer][i][j] -= nn.learningRate * gradient
		}
	}
}

// prepareTrainingData 准备训练数据
func (m *DeepLearningModel) prepareTrainingData(data *TrainingData) ([][]float64, []float64) {
	trainX := make([][]float64, 0, len(data.Ratings))
	trainY := make([]float64, 0, len(data.Ratings))
	
	for _, rating := range data.Ratings {
		input := m.buildInputFeatures(rating.UserID, rating.ContentID)
		trainX = append(trainX, input)
		trainY = append(trainY, rating.Rating/5.0) // 归一化到0-1
	}
	
	return trainX, trainY
}

// buildInputFeatures 构建输入特征
func (m *DeepLearningModel) buildInputFeatures(userID, contentID uuid.UUID) []float64 {
	// 简化的特征构建
	features := make([]float64, m.neuralNetwork.inputSize)
	
	// 用户ID哈希特征
	userHash := float64(userID.ID() % 1000) / 1000.0
	features[0] = userHash
	
	// 内容ID哈希特征
	contentHash := float64(contentID.ID() % 1000) / 1000.0
	features[1] = contentHash
	
	// 填充其他特征（简化实现）
	for i := 2; i < len(features); i++ {
		features[i] = math.Mod(float64(i)*userHash*contentHash, 1.0)
	}
	
	return features
}