package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// CrossModalService 跨模态推理服�?
type CrossModalService struct {
	providerManager *providers.Manager
	imageService    ImageService
	audioService    AudioService
	textService     TextService
	embeddingCache  map[string][]float64
	config          CrossModalConfig
	logger          *zap.Logger
}

// CrossModalConfig 跨模态推理配�?
type CrossModalConfig struct {
	EmbeddingDimension    int     `json:"embedding_dimension" yaml:"embedding_dimension"`
	SimilarityThreshold   float64 `json:"similarity_threshold" yaml:"similarity_threshold"`
	MaxCacheSize          int     `json:"max_cache_size" yaml:"max_cache_size"`
	EnableSemanticSearch  bool    `json:"enable_semantic_search" yaml:"enable_semantic_search"`
	EnableCrossModalAlign bool    `json:"enable_cross_modal_align" yaml:"enable_cross_modal_align"`
	DefaultProvider       string  `json:"default_provider" yaml:"default_provider"`
	DefaultModel          string  `json:"default_model" yaml:"default_model"`
}

// CrossModalRequest 跨模态推理请�?
type CrossModalRequest struct {
	ID          string                    `json:"id"`
	UserID      string                    `json:"user_id"`
	SessionID   string                    `json:"session_id"`
	Type        CrossModalInferenceType   `json:"type"`
	Inputs      []CrossModalInput         `json:"inputs"`
	Query       string                    `json:"query,omitempty"`
	Config      CrossModalInferenceConfig `json:"config"`
	Timestamp   time.Time                 `json:"timestamp"`
}

// CrossModalInferenceType 跨模态推理类�?
type CrossModalInferenceType string

const (
	InferenceTypeSemanticSearch   CrossModalInferenceType = "semantic_search"   // 语义搜索
	InferenceTypeContentMatching  CrossModalInferenceType = "content_matching"  // 内容匹配
	InferenceTypeCrossModalAlign  CrossModalInferenceType = "cross_modal_align" // 跨模态对�?
	InferenceTypeMultiModalQA     CrossModalInferenceType = "multimodal_qa"     // 多模态问�?
	InferenceTypeSceneUnderstand  CrossModalInferenceType = "scene_understand"  // 场景理解
	InferenceTypeEmotionAnalysis  CrossModalInferenceType = "emotion_analysis"  // 情感分析
	InferenceTypeContentGenerate  CrossModalInferenceType = "content_generate"  // 内容生成
)

// CrossModalInput 跨模态输�?
type CrossModalInput struct {
	ID       string                 `json:"id"`
	Type     models.InputType       `json:"type"`
	Content  interface{}            `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Weight   float64                `json:"weight"` // 输入权重
}

// CrossModalInferenceConfig 跨模态推理配�?
type CrossModalInferenceConfig struct {
	Provider           string                 `json:"provider"`
	Model              string                 `json:"model"`
	Temperature        float32                `json:"temperature"`
	MaxResults         int                    `json:"max_results"`
	SimilarityThreshold float64               `json:"similarity_threshold"`
	EnableExplanation  bool                   `json:"enable_explanation"`
	CustomParams       map[string]interface{} `json:"custom_params"`
}

// CrossModalResponse 跨模态推理响�?
type CrossModalResponse struct {
	ID            string                  `json:"id"`
	RequestID     string                  `json:"request_id"`
	Type          CrossModalInferenceType `json:"type"`
	Results       []CrossModalResult      `json:"results"`
	Explanation   string                  `json:"explanation,omitempty"`
	Confidence    float64                 `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata      map[string]interface{}  `json:"metadata"`
	Timestamp     time.Time               `json:"timestamp"`
}

// CrossModalResult 跨模态推理结�?
type CrossModalResult struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Content     interface{}            `json:"content"`
	Similarity  float64                `json:"similarity"`
	Confidence  float64                `json:"confidence"`
	Explanation string                 `json:"explanation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SemanticEmbedding 语义嵌入
type SemanticEmbedding struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Vector    []float64 `json:"vector"`
	Content   string    `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`
}

// NewCrossModalService 创建跨模态推理服�?
func NewCrossModalService(
	providerManager *providers.Manager,
	logger *zap.Logger,
) *CrossModalService {
	return &CrossModalService{
		providerManager: providerManager,
		imageService:    nil, // TODO: 注入图像服务
		audioService:    nil, // TODO: 注入音频服务
		textService:     nil, // TODO: 注入文本服务
		embeddingCache:  make(map[string][]float64),
		config: CrossModalConfig{
			EmbeddingDimension:    768,
			SimilarityThreshold:   0.7,
			MaxCacheSize:          1000,
			EnableSemanticSearch:  true,
			EnableCrossModalAlign: true,
			DefaultProvider:       "openai",
			DefaultModel:          "gpt-4",
		},
		logger: logger,
	}
}

// ProcessCrossModalInference 处理跨模态推理请�?
func (s *CrossModalService) ProcessCrossModalInference(ctx context.Context, req *CrossModalRequest) (*CrossModalResponse, error) {
	startTime := time.Now()

	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 设置默认�?
	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	req.Timestamp = time.Now()

	// 根据推理类型处理请求
	var results []CrossModalResult
	var explanation string
	var confidence float64
	var err error

	switch req.Type {
	case InferenceTypeSemanticSearch:
		results, confidence, err = s.performSemanticSearch(ctx, req)
	case InferenceTypeContentMatching:
		results, confidence, err = s.performContentMatching(ctx, req)
	case InferenceTypeCrossModalAlign:
		results, confidence, err = s.performCrossModalAlignment(ctx, req)
	case InferenceTypeMultiModalQA:
		results, confidence, err = s.performMultiModalQA(ctx, req)
	case InferenceTypeSceneUnderstand:
		results, confidence, err = s.performSceneUnderstanding(ctx, req)
	case InferenceTypeEmotionAnalysis:
		results, confidence, err = s.performEmotionAnalysis(ctx, req)
	case InferenceTypeContentGenerate:
		results, confidence, err = s.performContentGeneration(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported inference type: %s", req.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// 生成解释（如果启用）
	if req.Config.EnableExplanation {
		explanation, _ = s.generateExplanation(ctx, req, results)
	}

	// 创建响应
	response := &CrossModalResponse{
		ID:             uuid.New().String(),
		RequestID:      req.ID,
		Type:           req.Type,
		Results:        results,
		Explanation:    explanation,
		Confidence:     confidence,
		ProcessingTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"provider":     req.Config.Provider,
			"model":        req.Config.Model,
			"input_count":  len(req.Inputs),
			"result_count": len(results),
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// performSemanticSearch 执行语义搜索
func (s *CrossModalService) performSemanticSearch(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	if req.Query == "" {
		return nil, 0, fmt.Errorf("query is required for semantic search")
	}

	// 获取查询的嵌入向�?
	queryEmbedding, err := s.getTextEmbedding(ctx, req.Query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get query embedding: %w", err)
	}

	var results []CrossModalResult
	var totalSimilarity float64

	// 为每个输入计算相似度
	for _, input := range req.Inputs {
		embedding, err := s.getInputEmbedding(ctx, input)
		if err != nil {
			continue // 跳过无法处理的输�?
		}

		similarity := s.calculateCosineSimilarity(queryEmbedding, embedding)
		if similarity >= req.Config.SimilarityThreshold {
			result := CrossModalResult{
				ID:         uuid.New().String(),
				Type:       string(input.Type),
				Content:    input.Content,
				Similarity: similarity,
				Confidence: similarity,
				Metadata: map[string]interface{}{
					"input_id": input.ID,
					"weight":   input.Weight,
				},
			}
			results = append(results, result)
			totalSimilarity += similarity
		}
	}

	// 按相似度排序
	s.sortResultsBySimilarity(results)

	// 限制结果数量
	if len(results) > req.Config.MaxResults {
		results = results[:req.Config.MaxResults]
	}

	confidence := 0.0
	if len(results) > 0 {
		confidence = totalSimilarity / float64(len(results))
	}

	return results, confidence, nil
}

// performContentMatching 执行内容匹配
func (s *CrossModalService) performContentMatching(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	if len(req.Inputs) < 2 {
		return nil, 0, fmt.Errorf("content matching requires at least 2 inputs")
	}

	var results []CrossModalResult
	var totalConfidence float64

	// 计算所有输入对之间的相似度
	for i := 0; i < len(req.Inputs); i++ {
		for j := i + 1; j < len(req.Inputs); j++ {
			input1 := req.Inputs[i]
			input2 := req.Inputs[j]

			similarity, err := s.calculateInputSimilarity(ctx, input1, input2)
			if err != nil {
				continue
			}

			if similarity >= req.Config.SimilarityThreshold {
				result := CrossModalResult{
					ID:         uuid.New().String(),
					Type:       "content_match",
					Content: map[string]interface{}{
						"input1":     input1,
						"input2":     input2,
						"match_type": s.getMatchType(input1.Type, input2.Type),
					},
					Similarity: similarity,
					Confidence: similarity,
					Metadata: map[string]interface{}{
						"input1_id": input1.ID,
						"input2_id": input2.ID,
					},
				}
				results = append(results, result)
				totalConfidence += similarity
			}
		}
	}

	confidence := 0.0
	if len(results) > 0 {
		confidence = totalConfidence / float64(len(results))
	}

	return results, confidence, nil
}

// performCrossModalAlignment 执行跨模态对�?
func (s *CrossModalService) performCrossModalAlignment(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	// 按模态类型分组输�?
	modalGroups := s.groupInputsByModality(req.Inputs)

	var results []CrossModalResult
	var totalConfidence float64

	// 在不同模态之间寻找对�?
	for modality1, inputs1 := range modalGroups {
		for modality2, inputs2 := range modalGroups {
			if modality1 >= modality2 { // 避免重复比较
				continue
			}

			alignments, confidence := s.findModalityAlignments(ctx, inputs1, inputs2)
			results = append(results, alignments...)
			totalConfidence += confidence
		}
	}

	avgConfidence := 0.0
	if len(results) > 0 {
		avgConfidence = totalConfidence / float64(len(results))
	}

	return results, avgConfidence, nil
}

// performMultiModalQA 执行多模态问�?
func (s *CrossModalService) performMultiModalQA(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	if req.Query == "" {
		return nil, 0, fmt.Errorf("query is required for multimodal QA")
	}

	// 构建多模态上下文（暂时不使用�?
	// multiModalContext := s.buildMultiModalContext(req.Inputs)

	// 调用AI提供商进行问�?
	provider, err := s.getProvider(req.Config.Provider)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get provider: %w", err)
	}
	
	// 构建多模态请�?
	multimodalReq := &models.MultimodalRequest{
		ID:        req.ID,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Type:      models.MultimodalTypeChat,
		Inputs:    s.convertToMultimodalInputs(req.Inputs, req.Query),
		Config: models.MultimodalConfig{
			Provider:    req.Config.Provider,
			Model:       req.Config.Model,
			Temperature: req.Config.Temperature,
			MaxTokens:   2000,
		},
	}

	// 调用多模态服�?
	response, err := s.callMultiModalProvider(ctx, provider, multimodalReq)
	if err != nil {
		return nil, 0, fmt.Errorf("multimodal QA failed: %w", err)
	}

	// 转换响应为跨模态结�?
	results := s.convertToCrossModalResults(response)
	confidence := s.calculateResponseConfidence(response)

	return results, confidence, nil
}

// performSceneUnderstanding 执行场景理解
func (s *CrossModalService) performSceneUnderstanding(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	var results []CrossModalResult
	var totalConfidence float64

	// 分析每个输入的场景信�?
	for _, input := range req.Inputs {
		sceneInfo, confidence, err := s.analyzeSceneFromInput(ctx, input)
		if err != nil {
			continue
		}

		result := CrossModalResult{
			ID:         uuid.New().String(),
			Type:       "scene_analysis",
			Content:    sceneInfo,
			Confidence: confidence,
			Metadata: map[string]interface{}{
				"input_id":   input.ID,
				"input_type": input.Type,
			},
		}
		results = append(results, result)
		totalConfidence += confidence
	}

	// 综合场景理解
	if len(results) > 1 {
		combinedScene, confidence := s.combineSceneAnalysis(results)
		combinedResult := CrossModalResult{
			ID:         uuid.New().String(),
			Type:       "combined_scene",
			Content:    combinedScene,
			Confidence: confidence,
			Metadata: map[string]interface{}{
				"source_count": len(results),
			},
		}
		results = append(results, combinedResult)
		totalConfidence += confidence
	}

	avgConfidence := 0.0
	if len(results) > 0 {
		avgConfidence = totalConfidence / float64(len(results))
	}

	return results, avgConfidence, nil
}

// performEmotionAnalysis 执行情感分析
func (s *CrossModalService) performEmotionAnalysis(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	var results []CrossModalResult
	var totalConfidence float64

	// 分析每个输入的情感信�?
	for _, input := range req.Inputs {
		emotion, confidence, err := s.analyzeEmotionFromInput(ctx, input)
		if err != nil {
			continue
		}

		result := CrossModalResult{
			ID:         uuid.New().String(),
			Type:       "emotion_analysis",
			Content:    emotion,
			Confidence: confidence,
			Metadata: map[string]interface{}{
				"input_id":   input.ID,
				"input_type": input.Type,
			},
		}
		results = append(results, result)
		totalConfidence += confidence
	}

	// 综合情感分析
	if len(results) > 1 {
		combinedEmotion, confidence := s.combineEmotionAnalysis(results)
		combinedResult := CrossModalResult{
			ID:         uuid.New().String(),
			Type:       "combined_emotion",
			Content:    combinedEmotion,
			Confidence: confidence,
			Metadata: map[string]interface{}{
				"source_count": len(results),
			},
		}
		results = append(results, combinedResult)
		totalConfidence += confidence
	}

	avgConfidence := 0.0
	if len(results) > 0 {
		avgConfidence = totalConfidence / float64(len(results))
	}

	return results, avgConfidence, nil
}

// performContentGeneration 执行内容生成
func (s *CrossModalService) performContentGeneration(ctx context.Context, req *CrossModalRequest) ([]CrossModalResult, float64, error) {
	// 基于输入内容生成新的多模态内�?
	provider, err := s.getProvider(req.Config.Provider)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get provider: %w", err)
	}

	// 构建生成请求
	multimodalReq := &models.MultimodalRequest{
		ID:        req.ID,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Type:      models.MultimodalTypeGeneration,
		Inputs:    s.convertToMultimodalInputs(req.Inputs, req.Query),
		Config: models.MultimodalConfig{
			Provider:    req.Config.Provider,
			Model:       req.Config.Model,
			Temperature: req.Config.Temperature,
			MaxTokens:   2000,
		},
	}

	// 调用生成服务
	response, err := s.callMultiModalProvider(ctx, provider, multimodalReq)
	if err != nil {
		return nil, 0, fmt.Errorf("content generation failed: %w", err)
	}

	// 转换响应
	results := s.convertToCrossModalResults(response)
	confidence := s.calculateResponseConfidence(response)

	return results, confidence, nil
}

// 辅助方法

// validateRequest 验证请求
func (s *CrossModalService) validateRequest(req *CrossModalRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if len(req.Inputs) == 0 {
		return fmt.Errorf("inputs cannot be empty")
	}
	if req.Config.Provider == "" {
		req.Config.Provider = s.config.DefaultProvider
	}
	if req.Config.Model == "" {
		req.Config.Model = s.config.DefaultModel
	}
	if req.Config.MaxResults <= 0 {
		req.Config.MaxResults = 10
	}
	if req.Config.SimilarityThreshold <= 0 {
		req.Config.SimilarityThreshold = s.config.SimilarityThreshold
	}
	return nil
}

// getProvider 获取指定的AI提供�?
func (s *CrossModalService) getProvider(providerName string) (providers.AIProvider, error) {
	provider, err := s.providerManager.GetProvider(providerName)
	if err != nil {
		return nil, fmt.Errorf("provider %s not found: %w", providerName, err)
	}
	return provider, nil
}

// getTextEmbedding 获取文本嵌入向量
func (s *CrossModalService) getTextEmbedding(ctx context.Context, text string) ([]float64, error) {
	// 检查缓�?
	if embedding, exists := s.embeddingCache[text]; exists {
		return embedding, nil
	}

	// 调用AI提供商获取嵌�?
	provider, err := s.getProvider(s.config.DefaultProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}
	embedding32, err := provider.Embed(ctx, text)
	if err != nil {
		return nil, err
	}

	// 转换为float64
	embedding := make([]float64, len(embedding32))
	for i, v := range embedding32 {
		embedding[i] = float64(v)
	}

	// 缓存结果
	s.cacheEmbedding(text, embedding)

	return embedding, nil
}

// getInputEmbedding 获取输入的嵌入向�?
func (s *CrossModalService) getInputEmbedding(ctx context.Context, input CrossModalInput) ([]float64, error) {
	switch input.Type {
	case models.InputTypeText:
		if textContent, ok := input.Content.(models.TextInput); ok {
			return s.getTextEmbedding(ctx, textContent.Content)
		}
		if textStr, ok := input.Content.(string); ok {
			return s.getTextEmbedding(ctx, textStr)
		}
	case models.InputTypeImage:
		return s.getImageEmbedding(ctx, input)
	case models.InputTypeAudio:
		return s.getAudioEmbedding(ctx, input)
	}
	return nil, fmt.Errorf("unsupported input type: %s", input.Type)
}

// getImageEmbedding 获取图像嵌入向量
func (s *CrossModalService) getImageEmbedding(ctx context.Context, input CrossModalInput) ([]float64, error) {
	// 这里应该调用图像编码器获取嵌入向�?
	// 暂时返回模拟数据
	return s.generateMockEmbedding(), nil
}

// getAudioEmbedding 获取音频嵌入向量
func (s *CrossModalService) getAudioEmbedding(ctx context.Context, input CrossModalInput) ([]float64, error) {
	// 这里应该调用音频编码器获取嵌入向�?
	// 暂时返回模拟数据
	return s.generateMockEmbedding(), nil
}

// calculateCosineSimilarity 计算余弦相似�?
func (s *CrossModalService) calculateCosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// calculateInputSimilarity 计算输入相似�?
func (s *CrossModalService) calculateInputSimilarity(ctx context.Context, input1, input2 CrossModalInput) (float64, error) {
	embedding1, err := s.getInputEmbedding(ctx, input1)
	if err != nil {
		return 0, err
	}

	embedding2, err := s.getInputEmbedding(ctx, input2)
	if err != nil {
		return 0, err
	}

	return s.calculateCosineSimilarity(embedding1, embedding2), nil
}

// sortResultsBySimilarity 按相似度排序结果
func (s *CrossModalService) sortResultsBySimilarity(results []CrossModalResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// getMatchType 获取匹配类型
func (s *CrossModalService) getMatchType(type1, type2 models.InputType) string {
	if type1 == type2 {
		return fmt.Sprintf("same_modality_%s", type1)
	}
	return fmt.Sprintf("cross_modality_%s_%s", type1, type2)
}

// groupInputsByModality 按模态分组输�?
func (s *CrossModalService) groupInputsByModality(inputs []CrossModalInput) map[models.InputType][]CrossModalInput {
	groups := make(map[models.InputType][]CrossModalInput)
	for _, input := range inputs {
		groups[input.Type] = append(groups[input.Type], input)
	}
	return groups
}

// findModalityAlignments 寻找模态对�?
func (s *CrossModalService) findModalityAlignments(ctx context.Context, inputs1, inputs2 []CrossModalInput) ([]CrossModalResult, float64) {
	var results []CrossModalResult
	var totalConfidence float64

	for _, input1 := range inputs1 {
		for _, input2 := range inputs2 {
			similarity, err := s.calculateInputSimilarity(ctx, input1, input2)
			if err != nil {
				continue
			}

			if similarity >= s.config.SimilarityThreshold {
				result := CrossModalResult{
					ID:   uuid.New().String(),
					Type: "modality_alignment",
					Content: map[string]interface{}{
						"input1":      input1,
						"input2":      input2,
						"alignment":   similarity,
					},
					Similarity: similarity,
					Confidence: similarity,
					Metadata: map[string]interface{}{
						"modality1": input1.Type,
						"modality2": input2.Type,
					},
				}
				results = append(results, result)
				totalConfidence += similarity
			}
		}
	}

	return results, totalConfidence
}

// buildMultiModalContext 构建多模态上下文
func (s *CrossModalService) buildMultiModalContext(inputs []CrossModalInput) string {
	var contextParts []string
	for _, input := range inputs {
		switch input.Type {
		case models.InputTypeText:
			if textContent, ok := input.Content.(models.TextInput); ok {
				contextParts = append(contextParts, textContent.Content)
			}
		case models.InputTypeImage:
			contextParts = append(contextParts, "[图像内容]")
		case models.InputTypeAudio:
			contextParts = append(contextParts, "[音频内容]")
		}
	}
	return strings.Join(contextParts, " ")
}

// convertToMultimodalInputs 转换为多模态输�?
func (s *CrossModalService) convertToMultimodalInputs(inputs []CrossModalInput, query string) []models.MultimodalInput {
	var multimodalInputs []models.MultimodalInput

	// 添加查询文本
	if query != "" {
		multimodalInputs = append(multimodalInputs, models.MultimodalInput{
			Type: models.InputTypeText,
			Content: models.TextInput{
				Content: query,
			},
		})
	}

	// 转换输入
	for _, input := range inputs {
		// 转换metadata类型
		customMetadata := make(map[string]string)
		for k, v := range input.Metadata {
			if str, ok := v.(string); ok {
				customMetadata[k] = str
			} else {
				customMetadata[k] = fmt.Sprintf("%v", v)
			}
		}
		
		multimodalInput := models.MultimodalInput{
			Type:     input.Type,
			Content:  input.Content,
			Metadata: models.InputMetadata{
				Custom: customMetadata,
			},
		}
		multimodalInputs = append(multimodalInputs, multimodalInput)
	}

	return multimodalInputs
}

// callMultiModalProvider 调用多模态提供商
func (s *CrossModalService) callMultiModalProvider(ctx context.Context, provider providers.AIProvider, req *models.MultimodalRequest) (*models.MultimodalResponse, error) {
	// 这里应该调用实际的多模态服�?
	// 暂时返回模拟响应
	return &models.MultimodalResponse{
		ID:        uuid.New().String(),
		RequestID: req.ID,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Type:      req.Type,
		Outputs: []models.MultimodalOutput{
			{
				Type: models.OutputTypeText,
				Content: models.TextOutput{
					Content: "这是一个模拟的多模态响�?,
				},
			},
		},
		CreatedAt: time.Now(),
		Status:    "completed",
	}, nil
}

// convertToCrossModalResults 转换为跨模态结�?
func (s *CrossModalService) convertToCrossModalResults(response *models.MultimodalResponse) []CrossModalResult {
	var results []CrossModalResult
	for _, output := range response.Outputs {
		result := CrossModalResult{
			ID:         uuid.New().String(),
			Type:       string(output.Type),
			Content:    output.Content,
			Confidence: 0.8, // 默认置信�?
			Metadata: map[string]interface{}{
				"response_id": response.ID,
			},
		}
		results = append(results, result)
	}
	return results
}

// calculateResponseConfidence 计算响应置信�?
func (s *CrossModalService) calculateResponseConfidence(response *models.MultimodalResponse) float64 {
	// 基于响应状态和内容计算置信�?
	if response.Status == "completed" {
		return 0.8
	}
	return 0.5
}

// analyzeSceneFromInput 从输入分析场�?
func (s *CrossModalService) analyzeSceneFromInput(ctx context.Context, input CrossModalInput) (map[string]interface{}, float64, error) {
	// 这里应该根据输入类型调用相应的场景分析服�?
	sceneInfo := map[string]interface{}{
		"type":        "unknown",
		"description": "场景分析结果",
		"confidence":  0.7,
	}
	return sceneInfo, 0.7, nil
}

// combineSceneAnalysis 合并场景分析
func (s *CrossModalService) combineSceneAnalysis(results []CrossModalResult) (map[string]interface{}, float64) {
	combined := map[string]interface{}{
		"type":        "combined_scene",
		"description": "综合场景分析结果",
		"sources":     len(results),
	}
	return combined, 0.8
}

// analyzeEmotionFromInput 从输入分析情�?
func (s *CrossModalService) analyzeEmotionFromInput(ctx context.Context, input CrossModalInput) (map[string]interface{}, float64, error) {
	// 这里应该根据输入类型调用相应的情感分析服�?
	emotion := map[string]interface{}{
		"primary":   "neutral",
		"secondary": []string{"calm", "focused"},
		"intensity": 0.5,
	}
	return emotion, 0.7, nil
}

// combineEmotionAnalysis 合并情感分析
func (s *CrossModalService) combineEmotionAnalysis(results []CrossModalResult) (map[string]interface{}, float64) {
	combined := map[string]interface{}{
		"primary":   "neutral",
		"secondary": []string{"mixed"},
		"intensity": 0.6,
		"sources":   len(results),
	}
	return combined, 0.75
}

// generateExplanation 生成解释
func (s *CrossModalService) generateExplanation(ctx context.Context, req *CrossModalRequest, results []CrossModalResult) (string, error) {
	explanation := fmt.Sprintf("基于%d个输入进�?s推理，生成了%d个结果�?, 
		len(req.Inputs), req.Type, len(results))
	return explanation, nil
}

// cacheEmbedding 缓存嵌入向量
func (s *CrossModalService) cacheEmbedding(key string, embedding []float64) {
	if len(s.embeddingCache) >= s.config.MaxCacheSize {
		// 简单的LRU策略：删除第一个元�?
		for k := range s.embeddingCache {
			delete(s.embeddingCache, k)
			break
		}
	}
	s.embeddingCache[key] = embedding
}

// generateMockEmbedding 生成模拟嵌入向量
func (s *CrossModalService) generateMockEmbedding() []float64 {
	embedding := make([]float64, s.config.EmbeddingDimension)
	for i := range embedding {
		embedding[i] = math.Sin(float64(i)) * 0.5
	}
	return embedding
}

// TextService 文本服务接口
type TextService interface {
	AnalyzeText(ctx context.Context, text string) (map[string]interface{}, error)
	GetTextEmbedding(ctx context.Context, text string) ([]float64, error)
}
