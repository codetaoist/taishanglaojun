package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// TranslationProcessor 翻译处理器
type TranslationProcessor struct {
	*BaseProcessor
}

// NewTranslationProcessor 创建翻译处理器
func NewTranslationProcessor(
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *TranslationProcessor {
	return &TranslationProcessor{
		BaseProcessor: NewBaseProcessor(
			string(ProcessorTypeTranslation),
			validator,
			preprocessor,
			postprocessor,
			errorHandler,
		),
	}
}

// Process 处理翻译请求
func (tp *TranslationProcessor) Process(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// 验证输入
	if err := tp.Validate(inputs, config); err != nil {
		return nil, err
	}

	// 预处理输入
	processedInputs, err := tp.PreprocessInputs(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	var outputs []models.MultimodalOutput

	// 处理每个输入
	for _, input := range processedInputs {
		output, err := tp.processInput(ctx, provider, input, config)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	// 后处理输出
	finalOutputs, err := tp.PostprocessOutputs(ctx, outputs, config.ExpectedOutputs)
	if err != nil {
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	return finalOutputs, nil
}

// processInput 处理单个输入
func (tp *TranslationProcessor) processInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	switch input.Type {
	case models.InputTypeText:
		return tp.processTextTranslation(ctx, provider, input, config)
	case models.InputTypeImage:
		return tp.processImageTranslation(ctx, provider, input, config)
	case models.InputTypeAudio:
		return tp.processAudioTranslation(ctx, provider, input, config)
	case models.InputTypeVideo:
		return tp.processVideoTranslation(ctx, provider, input, config)
	default:
		return models.MultimodalOutput{}, fmt.Errorf("unsupported input type for translation: %s", input.Type)
	}
}

// getTargetLanguage 获取目标语言
func (tp *TranslationProcessor) getTargetLanguage(config models.MultimodalConfig) string {
	// 首先检查 CustomConfig 中的 language 设置
	if config.CustomConfig != nil {
		if lang, ok := config.CustomConfig["language"].(string); ok && lang != "" {
			return lang
		}
		if lang, ok := config.CustomConfig["target_language"].(string); ok && lang != "" {
			return lang
		}
	}
	
	// 检查 AudioConfig 中的 Language 设置
	if config.AudioConfig.Language != "" {
		return config.AudioConfig.Language
	}
	
	// 默认为英语
	return "English"
}

// processTextTranslation 处理文本翻译
func (tp *TranslationProcessor) processTextTranslation(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 获取目标语言
	targetLanguage := tp.getTargetLanguage(config)

	// 构建翻译请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Translate the following text to %s: %s", targetLanguage, textInput.Content),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.3, // 较低的温度以获得更准确的翻译
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := tp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("text translation failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content: chatResp.Message.Content,
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Custom: map[string]string{
				"translation_type": "text",
				"source_text":      textInput.Content,
				"target_language":  targetLanguage,
				"model":            "translation",
				"tokens_used":      fmt.Sprintf("%d", chatResp.Usage.TotalTokens),
			},
		},
	}

	return output, nil
}

// processImageTranslation 处理图像翻译（先分析图像内容，再翻译）
func (tp *TranslationProcessor) processImageTranslation(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	imageInput, ok := input.Content.(*models.ImageInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid image input content")
	}

	// 获取目标语言
	targetLanguage := tp.getTargetLanguage(config)

	// 第一步：分析图像内容
	analyzeReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Describe the content of this image in detail, including any text that appears in the image: %s", imageInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.3,
		MaxTokens:   config.MaxTokens,
	}

	var analyzeResp *providers.ChatResponse
	err := tp.ProcessWithRetry(ctx, func() error {
		var err error
		analyzeResp, err = provider.Chat(ctx, *analyzeReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image analysis for translation failed: %w", err)
	}

	// 第二步：翻译图像描述
	translateReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Translate the following image description to %s: %s", targetLanguage, analyzeResp.Message.Content),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.3,
		MaxTokens:   config.MaxTokens,
	}

	var translateResp *providers.ChatResponse
	err = tp.ProcessWithRetry(ctx, func() error {
		var err error
		translateResp, err = provider.Chat(ctx, *translateReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image description translation failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content: translateResp.Message.Content,
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Custom: map[string]string{
				"translation_type":    "image",
				"image_url":           imageInput.URL,
				"original_description": analyzeResp.Message.Content,
				"target_language":     targetLanguage,
				"model":               "translation",
				"tokens_used":         fmt.Sprintf("%d", analyzeResp.Usage.TotalTokens + translateResp.Usage.TotalTokens),
			},
		},
	}

	return output, nil
}

// processAudioTranslation 处理音频翻译
func (tp *TranslationProcessor) processAudioTranslation(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	audioInput, ok := input.Content.(*models.AudioInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid audio input content")
	}

	// 获取目标语言
	targetLanguage := tp.getTargetLanguage(config)

	// 构建音频翻译请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Translate the audio content from this URL to %s: %s", targetLanguage, audioInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.3,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := tp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("audio translation failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  chatResp.Message.Content,
			Language: targetLanguage,
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

// processVideoTranslation 处理视频翻译
func (tp *TranslationProcessor) processVideoTranslation(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	videoInput, ok := input.Content.(*models.VideoInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid video input content")
	}

	// 获取目标语言
	targetLanguage := tp.getTargetLanguage(config)

	// 构建视频翻译请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Translate the video content from this URL to %s: %s", targetLanguage, videoInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.3,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := tp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("video translation failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  chatResp.Message.Content,
			Language: targetLanguage,
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