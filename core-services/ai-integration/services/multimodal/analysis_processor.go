package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// AnalysisProcessor 分析处理器
type AnalysisProcessor struct {
	*BaseProcessor
}

// NewAnalysisProcessor 创建分析处理器
func NewAnalysisProcessor(
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *AnalysisProcessor {
	return &AnalysisProcessor{
		BaseProcessor: NewBaseProcessor(
			string(ProcessorTypeAnalysis),
			validator,
			preprocessor,
			postprocessor,
			errorHandler,
		),
	}
}

// Process 处理分析请求
func (ap *AnalysisProcessor) Process(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// 验证输入
	if err := ap.Validate(inputs, config); err != nil {
		return nil, err
	}

	// 预处理输入
	processedInputs, err := ap.PreprocessInputs(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	var outputs []models.MultimodalOutput

	// 处理每个输入
	for _, input := range processedInputs {
		output, err := ap.processInput(ctx, provider, input, config)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	// 后处理输出
	finalOutputs, err := ap.PostprocessOutputs(ctx, outputs, config.ExpectedOutputs)
	if err != nil {
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	return finalOutputs, nil
}

// processInput 处理单个输入
func (ap *AnalysisProcessor) processInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	switch input.Type {
	case models.InputTypeText:
		return ap.processTextAnalysis(ctx, provider, input, config)
	case models.InputTypeImage:
		return ap.processImageAnalysis(ctx, provider, input, config)
	case models.InputTypeAudio:
		return ap.processAudioAnalysis(ctx, provider, input, config)
	case models.InputTypeVideo:
		return ap.processVideoAnalysis(ctx, provider, input, config)
	default:
		return models.MultimodalOutput{}, fmt.Errorf("unsupported input type for analysis: %s", input.Type)
	}
}

// processTextAnalysis 处理文本分析
func (ap *AnalysisProcessor) processTextAnalysis(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 构建分析请求
	analyzeReq := &providers.AnalyzeRequest{
		Content: textInput.Content,
		Type:    "sentiment", // 使用具体的分析类型
		UserID:  "", // 可以从 config 中获取
	}

	// 调用AI提供者
	var analyzeResp *providers.AnalyzeResponse
	err := ap.ProcessWithRetry(ctx, func() error {
		var err error
		analyzeResp, err = provider.Analyze(ctx, *analyzeReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("text analysis failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content: fmt.Sprintf("Analysis Results:\nType: %s\nResult: %s\nConfidence: %.2f\nDetails: %v",
				analyzeResp.Type,
				analyzeResp.Result,
				analyzeResp.Confidence,
				analyzeResp.Details,
			),
			Language: "en", // 默认语言
			Format:   "text",
		},
		Metadata: models.OutputMetadata{
			MimeType: "text/plain",
			Quality: models.QualityMetrics{
				Relevance: analyzeResp.Confidence,
				Accuracy:  analyzeResp.Confidence,
			},
		},
	}

	return output, nil
}

// processImageAnalysis 处理图像分析
func (ap *AnalysisProcessor) processImageAnalysis(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	imageInput, ok := input.Content.(*models.ImageInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid image input content")
	}

	// 构建图像分析请求
	analyzeReq := &providers.ImageAnalyzeRequest{
		ImageURL: imageInput.URL,
		Prompt:   "Analyze this image and provide detailed information about objects, scenes, colors, composition, and any text present",
		Features: []string{"objects", "text", "colors", "description"},
		UserID:   "", // 可以从 config 中获取
	}

	// 调用AI提供者
	var analyzeResp *providers.ImageAnalyzeResponse
	err := ap.ProcessWithRetry(ctx, func() error {
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
			MimeType: "text/plain",
			Quality: models.QualityMetrics{
				Relevance: 0.8, // 默认相关性
				Accuracy:  0.8, // 默认准确性
			},
		},
	}

	return output, nil
}

// processAudioAnalysis 处理音频分析
func (ap *AnalysisProcessor) processAudioAnalysis(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	audioInput, ok := input.Content.(*models.AudioInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid audio input content")
	}

	// 构建分析请求（使用聊天接口进行音频分析）
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze this audio file and provide information about speech content, emotions, language, and audio quality: %s", audioInput.URL),
			},
		},
		Temperature: 0.3, // 较低的温度以获得更一致的分析结果
		MaxTokens:   config.MaxTokens,
		UserID:      "", // 可以从 config 中获取
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := ap.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("audio analysis failed: %w", err)
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
				Relevance: 0.8, // 默认相关性
				Accuracy:  0.8, // 默认准确性
			},
		},
	}

	return output, nil
}

// processVideoAnalysis 处理视频分析
func (ap *AnalysisProcessor) processVideoAnalysis(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	videoInput, ok := input.Content.(*models.VideoInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid video input content")
	}

	// 构建分析请求（使用聊天接口进行视频分析）
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Analyze this video file and provide information about visual content, scenes, actions, objects, and overall narrative: %s", videoInput.URL),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: 0.3, // 较低的温度以获得更一致的分析结果
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := ap.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("video analysis failed: %w", err)
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