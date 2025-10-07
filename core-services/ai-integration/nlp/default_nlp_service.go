package nlp

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DefaultNLPService 默认NLP服务实现
type DefaultNLPService struct {
	config    NLPConfig
	providers map[string]NLPProvider
	cache     NLPCache
	logger    *zap.Logger
	mutex     sync.RWMutex
}

// NLPCache NLP缓存接口
type NLPCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiry time.Duration)
	Delete(key string)
	Clear()
}

// NewDefaultNLPService 创建默认NLP服务
func NewDefaultNLPService(config NLPConfig, cache NLPCache, logger *zap.Logger) *DefaultNLPService {
	return &DefaultNLPService{
		config:    config,
		providers: make(map[string]NLPProvider),
		cache:     cache,
		logger:    logger,
	}
}

// RegisterProvider 注册NLP提供商
func (s *DefaultNLPService) RegisterProvider(name string, provider NLPProvider) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.providers[name] = provider
}

// TokenizeText 分词
func (s *DefaultNLPService) TokenizeText(ctx context.Context, input TextInput) (*TokenizationResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("tokenize:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*TokenizationResult); ok {
				s.logger.Debug("Tokenization result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpTokenization, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.TokenizeText(ctx, input)
	if err != nil {
		s.logger.Error("Tokenization failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("tokenize:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Tokenization completed",
		zap.String("input_id", input.ID),
		zap.Int("total_tokens", result.TotalTokens),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// AnalyzeSentiment 情感分析
func (s *DefaultNLPService) AnalyzeSentiment(ctx context.Context, input TextInput) (*SentimentAnalysisResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("sentiment:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*SentimentAnalysisResult); ok {
				s.logger.Debug("Sentiment analysis result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpSentiment, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.AnalyzeSentiment(ctx, input)
	if err != nil {
		s.logger.Error("Sentiment analysis failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("sentiment:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Sentiment analysis completed",
		zap.String("input_id", input.ID),
		zap.String("sentiment", result.OverallSentiment.Label),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// ExtractEntities 实体提取
func (s *DefaultNLPService) ExtractEntities(ctx context.Context, input TextInput) (*EntityExtractionResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("entities:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*EntityExtractionResult); ok {
				s.logger.Debug("Entity extraction result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpEntityExtraction, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.ExtractEntities(ctx, input)
	if err != nil {
		s.logger.Error("Entity extraction failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("entities:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Entity extraction completed",
		zap.String("input_id", input.ID),
		zap.Int("total_entities", result.TotalEntities),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// ClassifyText 文本分类
func (s *DefaultNLPService) ClassifyText(ctx context.Context, input TextInput) (*TextClassificationResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("classify:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*TextClassificationResult); ok {
				s.logger.Debug("Text classification result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpClassification, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.ClassifyText(ctx, input)
	if err != nil {
		s.logger.Error("Text classification failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("classify:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Text classification completed",
		zap.String("input_id", input.ID),
		zap.String("top_category", result.TopCategory.Name),
		zap.Float64("confidence", result.TopCategory.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// AnalyzeSemantics 语义分析
func (s *DefaultNLPService) AnalyzeSemantics(ctx context.Context, input TextInput) (*SemanticAnalysisResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("semantics:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*SemanticAnalysisResult); ok {
				s.logger.Debug("Semantic analysis result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpSemanticAnalysis, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.AnalyzeSemantics(ctx, input)
	if err != nil {
		s.logger.Error("Semantic analysis failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("semantics:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Semantic analysis completed",
		zap.String("input_id", input.ID),
		zap.Float64("complexity", result.Complexity),
		zap.Float64("coherence", result.Coherence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// ExtractKeywords 关键词提取
func (s *DefaultNLPService) ExtractKeywords(ctx context.Context, input TextInput) (*KeywordExtractionResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("keywords:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*KeywordExtractionResult); ok {
				s.logger.Debug("Keyword extraction result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpKeywordExtraction, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.ExtractKeywords(ctx, input)
	if err != nil {
		s.logger.Error("Keyword extraction failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("keywords:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Keyword extraction completed",
		zap.String("input_id", input.ID),
		zap.Int("total_keywords", result.TotalKeywords),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// SummarizeText 文本摘要
func (s *DefaultNLPService) SummarizeText(ctx context.Context, input TextInput) (*TextSummarizationResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("summary:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*TextSummarizationResult); ok {
				s.logger.Debug("Text summarization result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpSummarization, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.SummarizeText(ctx, input)
	if err != nil {
		s.logger.Error("Text summarization failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("summary:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Text summarization completed",
		zap.String("input_id", input.ID),
		zap.Float64("compression_ratio", result.CompressionRatio),
		zap.Float64("relevance", result.Relevance),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// AnalyzeIntent 意图分析
func (s *DefaultNLPService) AnalyzeIntent(ctx context.Context, input TextInput) (*IntentAnalysisResult, error) {
	// 检查缓存
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("intent:%s:%s", input.Language, s.generateTextHash(input.Text))
		if cached, found := s.cache.Get(cacheKey); found {
			if result, ok := cached.(*IntentAnalysisResult); ok {
				s.logger.Debug("Intent analysis result found in cache",
					zap.String("input_id", input.ID))
				return result, nil
			}
		}
	}

	provider, err := s.getBestProvider(OpIntentAnalysis, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.AnalyzeIntent(ctx, input)
	if err != nil {
		s.logger.Error("Intent analysis failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	// 缓存结果
	if s.config.EnableCache {
		cacheKey := fmt.Sprintf("intent:%s:%s", input.Language, s.generateTextHash(input.Text))
		s.cache.Set(cacheKey, result, s.config.CacheExpiry)
	}

	s.logger.Info("Intent analysis completed",
		zap.String("input_id", input.ID),
		zap.String("intent", result.Intent.Name),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// GenerateText 文本生成
func (s *DefaultNLPService) GenerateText(ctx context.Context, input TextGenerationInput) (*TextGenerationResult, error) {
	provider, err := s.getBestProvider(OpTextGeneration, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.GenerateText(ctx, input)
	if err != nil {
		s.logger.Error("Text generation failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Text generation completed",
		zap.String("input_id", input.ID),
		zap.Int("generated_length", len(result.GeneratedText)),
		zap.Float64("quality", result.Quality),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// TranslateText 文本翻译
func (s *DefaultNLPService) TranslateText(ctx context.Context, input TranslationInput) (*TranslationResult, error) {
	provider, err := s.getBestProvider(OpTranslation, input.SourceLang)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.TranslateText(ctx, input)
	if err != nil {
		s.logger.Error("Text translation failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Text translation completed",
		zap.String("input_id", input.ID),
		zap.String("source_lang", string(input.SourceLang)),
		zap.String("target_lang", string(input.TargetLang)),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// ParaphraseText 文本改写
func (s *DefaultNLPService) ParaphraseText(ctx context.Context, input TextInput) (*ParaphraseResult, error) {
	provider, err := s.getBestProvider(OpParaphrase, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.ParaphraseText(ctx, input)
	if err != nil {
		s.logger.Error("Text paraphrase failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Text paraphrase completed",
		zap.String("input_id", input.ID),
		zap.Int("paraphrases_count", len(result.Paraphrases)),
		zap.Float64("similarity", result.Similarity),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// ProcessConversation 处理对话
func (s *DefaultNLPService) ProcessConversation(ctx context.Context, input ConversationInput) (*ConversationResult, error) {
	provider, err := s.getBestProvider(OpConversation, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.ProcessConversation(ctx, input)
	if err != nil {
		s.logger.Error("Conversation processing failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Conversation processing completed",
		zap.String("input_id", input.ID),
		zap.String("user_id", input.UserID),
		zap.String("session_id", input.SessionID),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// AnswerQuestion 问答
func (s *DefaultNLPService) AnswerQuestion(ctx context.Context, input QuestionAnsweringInput) (*QuestionAnsweringResult, error) {
	provider, err := s.getBestProvider(OpQuestionAnswering, input.Language)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.AnswerQuestion(ctx, input)
	if err != nil {
		s.logger.Error("Question answering failed",
			zap.String("input_id", input.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Question answering completed",
		zap.String("input_id", input.ID),
		zap.Float64("confidence", result.Confidence),
		zap.Int("sources_count", len(result.Sources)),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// BatchProcess 批量处理
func (s *DefaultNLPService) BatchProcess(ctx context.Context, inputs []TextInput, operations []NLPOperation) (*BatchNLPResult, error) {
	if len(inputs) > s.config.MaxBatchSize {
		return nil, fmt.Errorf("batch size %d exceeds maximum %d", len(inputs), s.config.MaxBatchSize)
	}

	startTime := time.Now()
	result := &BatchNLPResult{
		ID:          uuid.New().String(),
		TotalInputs: len(inputs),
		Results:     make([]interface{}, 0),
		Errors:      make([]BatchNLPError, 0),
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// 按操作类型分组处理
	for _, operation := range operations {
		for _, input := range inputs {
			res, err := s.processOperation(ctx, input, operation)
			if err != nil {
				result.FailedInputs++
				result.Errors = append(result.Errors, BatchNLPError{
					InputID: input.ID,
					Error:   err.Error(),
				})
			} else {
				result.ProcessedInputs++
				result.Results = append(result.Results, res)
			}
		}
	}

	result.ProcessingTime = time.Since(startTime)

	s.logger.Info("Batch processing completed",
		zap.Int("total_inputs", result.TotalInputs),
		zap.Int("processed_inputs", result.ProcessedInputs),
		zap.Int("failed_inputs", result.FailedInputs),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// UpdateConfig 更新配置
func (s *DefaultNLPService) UpdateConfig(config NLPConfig) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.config = config
	return nil
}

// GetSupportedLanguages 获取支持的语言
func (s *DefaultNLPService) GetSupportedLanguages() []Language {
	languages := make(map[Language]bool)
	
	for _, provider := range s.providers {
		for _, lang := range provider.GetSupportedLanguages() {
			languages[lang] = true
		}
	}
	
	result := make([]Language, 0, len(languages))
	for lang := range languages {
		result = append(result, lang)
	}
	
	return result
}

// GetSupportedOperations 获取支持的操作
func (s *DefaultNLPService) GetSupportedOperations() []OperationType {
	operations := make(map[OperationType]bool)
	
	for _, provider := range s.providers {
		for _, op := range provider.GetSupportedOperations() {
			operations[op] = true
		}
	}
	
	result := make([]OperationType, 0, len(operations))
	for op := range operations {
		result = append(result, op)
	}
	
	return result
}

// 私有方法

// getBestProvider 获取最佳提供商
func (s *DefaultNLPService) getBestProvider(operation OperationType, language Language) (NLPProvider, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 收集支持该操作和语言的提供商
	type providerScore struct {
		name     string
		provider NLPProvider
		score    int
	}

	candidates := make([]providerScore, 0)

	for name, provider := range s.providers {
		// 检查是否支持操作
		supportedOps := provider.GetSupportedOperations()
		supportsOp := false
		for _, op := range supportedOps {
			if op == operation {
				supportsOp = true
				break
			}
		}

		if !supportsOp {
			continue
		}

		// 检查是否支持语言
		supportedLangs := provider.GetSupportedLanguages()
		supportsLang := false
		for _, lang := range supportedLangs {
			if lang == language || lang == LanguageAuto {
				supportsLang = true
				break
			}
		}

		if !supportsLang {
			continue
		}

		// 计算分数（基于配置中的优先级）
		score := 0
		if config, exists := s.config.Providers[name]; exists && config.Enabled {
			score = config.Priority
		}

		candidates = append(candidates, providerScore{
			name:     name,
			provider: provider,
			score:    score,
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no provider found for operation %s and language %s", operation, language)
	}

	// 按分数排序，选择最高分的提供商
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	return candidates[0].provider, nil
}

// processOperation 处理单个操作
func (s *DefaultNLPService) processOperation(ctx context.Context, input TextInput, operation NLPOperation) (interface{}, error) {
	switch operation.Type {
	case OpTokenization:
		return s.TokenizeText(ctx, input)
	case OpSentiment:
		return s.AnalyzeSentiment(ctx, input)
	case OpEntityExtraction:
		return s.ExtractEntities(ctx, input)
	case OpClassification:
		return s.ClassifyText(ctx, input)
	case OpSemanticAnalysis:
		return s.AnalyzeSemantics(ctx, input)
	case OpKeywordExtraction:
		return s.ExtractKeywords(ctx, input)
	case OpSummarization:
		return s.SummarizeText(ctx, input)
	case OpIntentAnalysis:
		return s.AnalyzeIntent(ctx, input)
	case OpParaphrase:
		return s.ParaphraseText(ctx, input)
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operation.Type)
	}
}

// generateTextHash 生成文本哈希
func (s *DefaultNLPService) generateTextHash(text string) string {
	// 简单的哈希实现，实际中可以使用更复杂的哈希算法
	if len(text) > 50 {
		return fmt.Sprintf("%x", text[:50])
	}
	return fmt.Sprintf("%x", text)
}