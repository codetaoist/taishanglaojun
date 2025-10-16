package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// ChatProcessor 聊天处理器
type ChatProcessor struct {
	*BaseProcessor
}

// NewChatProcessor 创建聊天处理器
func NewChatProcessor(
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *ChatProcessor {
	return &ChatProcessor{
		BaseProcessor: NewBaseProcessor(
			string(ProcessorTypeChat),
			validator,
			preprocessor,
			postprocessor,
			errorHandler,
		),
	}
}

// Process 处理聊天请求
func (cp *ChatProcessor) Process(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// 验证输入
	if err := cp.Validate(inputs, config); err != nil {
		return nil, err
	}

	// 预处理输入
	processedInputs, err := cp.PreprocessInputs(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	var outputs []models.MultimodalOutput

	// 处理每个输入
	for _, input := range processedInputs {
		output, err := cp.processInput(ctx, provider, input, config)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	// 后处理输出
	finalOutputs, err := cp.PostprocessOutputs(ctx, outputs, config.ExpectedOutputs)
	if err != nil {
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	return finalOutputs, nil
}

// processInput 处理单个输入
func (cp *ChatProcessor) processInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	switch input.Type {
	case models.InputTypeText:
		return cp.processTextInput(ctx, provider, input, config)
	case models.InputTypeImage:
		return cp.processImageInput(ctx, provider, input, config)
	case models.InputTypeAudio:
		return cp.processAudioInput(ctx, provider, input, config)
	case models.InputTypeVideo:
		return cp.processVideoInput(ctx, provider, input, config)
	default:
		return models.MultimodalOutput{}, fmt.Errorf("unsupported input type: %s", input.Type)
	}
}

// processTextInput 处理文本输入
func (cp *ChatProcessor) processTextInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 构建聊天请求
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: textInput.Content,
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := cp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("text chat failed: %w", err)
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
				Relevance: 0.9,
				Accuracy:  0.9,
			},
		},
	}

	return output, nil
}

// processImageInput 处理图像输入
func (cp *ChatProcessor) processImageInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	imageInput, ok := input.Content.(*models.ImageInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid image input content")
	}

	// 构建图像分析请求
	analyzeReq := &providers.ImageAnalyzeRequest{
		ImageURL: imageInput.URL,
		Prompt:   "Analyze this image and describe what you see",
		Features: []string{"objects", "text", "colors", "description"},
		UserID:   "",
	}

	// 调用AI提供者
	var analyzeResp *providers.ImageAnalyzeResponse
	err := cp.ProcessWithRetry(ctx, func() error {
		var err error
		analyzeResp, err = provider.AnalyzeImage(ctx, *analyzeReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image analysis failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  analyzeResp.Description,
			Language: "en",
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType:   "text/plain",
			Dimensions: imageInput.Dimensions,
			Quality: models.QualityMetrics{
				Relevance: 0.9,
				Accuracy:  0.8,
			},
		},
	}

	return output, nil
}

// processAudioInput 处理音频输入
func (cp *ChatProcessor) processAudioInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	audioInput, ok := input.Content.(*models.AudioInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid audio input content")
	}

	// 构建聊天请求（假设音频已转换为文本或可以直接处理）
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Process this audio file: %s", audioInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := cp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("audio chat failed: %w", err)
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

// processVideoInput 处理视频输入
func (cp *ChatProcessor) processVideoInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	videoInput, ok := input.Content.(*models.VideoInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid video input content")
	}

	// 构建聊天请求（假设视频可以直接处理）
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Process this video file: %s", videoInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := cp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("video chat failed: %w", err)
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