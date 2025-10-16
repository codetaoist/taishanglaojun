package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// SearchProcessor 搜索处理器
type SearchProcessor struct {
	*BaseProcessor
}

// NewSearchProcessor 创建搜索处理器
func NewSearchProcessor(
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *SearchProcessor {
	return &SearchProcessor{
		BaseProcessor: NewBaseProcessor(
			string(ProcessorTypeSearch),
			validator,
			preprocessor,
			postprocessor,
			errorHandler,
		),
	}
}

// Process 处理搜索请求
func (sp *SearchProcessor) Process(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// 验证输入
	if err := sp.Validate(inputs, config); err != nil {
		return nil, err
	}

	// 预处理输入
	processedInputs, err := sp.PreprocessInputs(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	var outputs []models.MultimodalOutput

	// 处理每个输入
	for _, input := range processedInputs {
		output, err := sp.processInput(ctx, provider, input, config)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	// 后处理输出
	finalOutputs, err := sp.PostprocessOutputs(ctx, outputs, config.ExpectedOutputs)
	if err != nil {
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	return finalOutputs, nil
}

// processInput 处理单个输入
func (sp *SearchProcessor) processInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	switch input.Type {
	case models.InputTypeText:
		return sp.processTextSearch(ctx, provider, input, config)
	case models.InputTypeImage:
		return sp.processImageSearch(ctx, provider, input, config)
	case models.InputTypeAudio:
		return sp.processAudioSearch(ctx, provider, input, config)
	case models.InputTypeVideo:
		return sp.processVideoSearch(ctx, provider, input, config)
	default:
		return models.MultimodalOutput{}, fmt.Errorf("unsupported input type for search: %s", input.Type)
	}
}

// processTextSearch 处理文本搜索
func (sp *SearchProcessor) processTextSearch(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 构建搜索请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Search for information about: %s. Provide comprehensive and relevant results.", textInput.Content),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.7,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := sp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("text search failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  chatResp.Message.Content,
			Language: "en",
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Quality: models.QualityMetrics{
				Relevance: 0.8,
				Accuracy:  0.8,
			},
		},
	}

	return output, nil
}

// processImageSearch 处理图像搜索
func (sp *SearchProcessor) processImageSearch(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	imageInput, ok := input.Content.(*models.ImageInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid image input content")
	}

	// 第一步：分析图像内容
	analyzeReq := &providers.ImageAnalyzeRequest{
		ImageURL: imageInput.URL,
		Prompt:   "Analyze this image and describe its content in detail for search purposes",
		Features: []string{"objects", "text", "colors", "description"},
		UserID:   "",
	}

	var analyzeResp *providers.ImageAnalyzeResponse
	err := sp.ProcessWithRetry(ctx, func() error {
		var err error
		analyzeResp, err = provider.AnalyzeImage(ctx, *analyzeReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image analysis for search failed: %w", err)
	}

	// 第二步：基于图像分析结果进行搜索
	searchReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Based on this image analysis, search for related information: %s", analyzeResp.Description),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.7,
		MaxTokens:   config.MaxTokens,
	}

	var searchResp *providers.ChatResponse
	err = sp.ProcessWithRetry(ctx, func() error {
		var err error
		searchResp, err = provider.Chat(ctx, *searchReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image search failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  searchResp.Message.Content,
			Language: "en",
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Quality: models.QualityMetrics{
				Relevance: 0.8,
				Accuracy:  0.8,
			},
		},
	}

	return output, nil
}

// processAudioSearch 处理音频搜索
func (sp *SearchProcessor) processAudioSearch(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	audioInput, ok := input.Content.(*models.AudioInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid audio input content")
	}

	// 构建音频搜索请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze the audio content from this URL and search for related information: %s", audioInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.7,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := sp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("audio search failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  chatResp.Message.Content,
			Language: "en",
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Duration: audioInput.Duration,
			Quality: models.QualityMetrics{
				Relevance: 0.8,
				Accuracy:  0.8,
			},
		},
	}

	return output, nil
}

// processVideoSearch 处理视频搜索
func (sp *SearchProcessor) processVideoSearch(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	videoInput, ok := input.Content.(*models.VideoInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid video input content")
	}

	// 构建视频搜索请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze the video content from this URL and search for related information: %s", videoInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.7,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := sp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("video search failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  chatResp.Message.Content,
			Language: "en",
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType:   "text/plain",
			Duration:   videoInput.Duration,
			Dimensions: videoInput.Dimensions,
			Quality: models.QualityMetrics{
				Relevance: 0.8,
				Accuracy:  0.8,
			},
		},
	}

	return output, nil
}

// SearchByKeywords 根据关键词搜索
func (sp *SearchProcessor) SearchByKeywords(ctx context.Context, provider providers.AIProvider, keywords []string, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	if len(keywords) == 0 {
		return models.MultimodalOutput{}, fmt.Errorf("no keywords provided for search")
	}

	// 构建关键词搜索请求
	keywordStr := ""
	for i, keyword := range keywords {
		if i > 0 {
			keywordStr += ", "
		}
		keywordStr += keyword
	}

	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Search for comprehensive information about these keywords: %s. Provide detailed and relevant results.", keywordStr),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.7,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := sp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("keyword search failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  chatResp.Message.Content,
			Language: "en",
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Quality: models.QualityMetrics{
				Relevance: 0.8,
				Accuracy:  0.8,
			},
		},
	}

	return output, nil
}