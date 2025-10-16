package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// GenerationProcessor 生成处理器
type GenerationProcessor struct {
	*BaseProcessor
}

// NewGenerationProcessor 创建生成处理器
func NewGenerationProcessor(
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *GenerationProcessor {
	return &GenerationProcessor{
		BaseProcessor: NewBaseProcessor(
			string(ProcessorTypeGeneration),
			validator,
			preprocessor,
			postprocessor,
			errorHandler,
		),
	}
}

// Process 处理生成请求
func (gp *GenerationProcessor) Process(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// 验证输入
	if err := gp.Validate(inputs, config); err != nil {
		return nil, err
	}

	// 预处理输入
	processedInputs, err := gp.PreprocessInputs(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	var outputs []models.MultimodalOutput

	// 处理每个输入
	for _, input := range processedInputs {
		output, err := gp.processInput(ctx, provider, input, config)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	// 后处理输出
	finalOutputs, err := gp.PostprocessOutputs(ctx, outputs, config.ExpectedOutputs)
	if err != nil {
		return nil, fmt.Errorf("postprocessing failed: %w", err)
	}

	return finalOutputs, nil
}

// processInput 处理单个输入
func (gp *GenerationProcessor) processInput(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	switch input.Type {
	case models.InputTypeText:
		return gp.processTextGeneration(ctx, provider, input, config)
	case models.InputTypeImage:
		return gp.processImageGeneration(ctx, provider, input, config)
	default:
		return models.MultimodalOutput{}, fmt.Errorf("unsupported input type for generation: %s", input.Type)
	}
}

// processTextGeneration 处理文本生成
func (gp *GenerationProcessor) processTextGeneration(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 构建生成请求
	generateReq := &providers.GenerateRequest{
		Prompt:      textInput.Content,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var generateResp *providers.GenerateResponse
	err := gp.ProcessWithRetry(ctx, func() error {
		var err error
		generateResp, err = provider.Generate(ctx, *generateReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("text generation failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content: generateResp.Content,
		},
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"generation_type": "text",
				"model":           config.Model,
				"tokens_used":     fmt.Sprintf("%d", generateResp.Usage.TotalTokens),
				"temperature":     fmt.Sprintf("%.2f", config.Temperature),
			},
		},
	}

	return output, nil
}

// processImageGeneration 处理图像生成
func (gp *GenerationProcessor) processImageGeneration(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	// 根据输入类型确定生成方式
	switch input.Type {
	case models.InputTypeText:
		return gp.processTextToImageGeneration(ctx, provider, input, config)
	case models.InputTypeImage:
		return gp.processImageToImageGeneration(ctx, provider, input, config)
	default:
		return models.MultimodalOutput{}, fmt.Errorf("unsupported input type for image generation: %s", input.Type)
	}
}

// processTextToImageGeneration 处理文本到图像生成
func (gp *GenerationProcessor) processTextToImageGeneration(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 构建图像生成请求
	imageGenReq := &providers.ImageGenerateRequest{
		Prompt:  textInput.Content,
		Model:   config.Model,
		Size:    "1024x1024", // 默认尺寸
		Style:   "realistic", // 默认风格
		Quality: "standard",  // 默认质量
		Count:   1,           // 生成一张图片
		UserID:  "",
	}

	// 调用AI提供者
	var imageGenResp *providers.ImageGenerateResponse
	err := gp.ProcessWithRetry(ctx, func() error {
		var err error
		imageGenResp, err = provider.GenerateImage(ctx, *imageGenReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image generation failed: %w", err)
	}

	// 构建输出
	var imageURL string
	var width, height int
	var format string
	if len(imageGenResp.Images) > 0 {
		imageURL = imageGenResp.Images[0].URL
		width = imageGenResp.Images[0].Width
		height = imageGenResp.Images[0].Height
		format = imageGenResp.Images[0].Format
	} else {
		width = 1024
		height = 1024
		format = "png"
	}

	output := models.MultimodalOutput{
		Type: models.OutputTypeImage,
		Content: &models.ImageOutput{
			URL:    imageURL,
			Dimensions: models.ImageDimensions{
				Width:  width,
				Height: height,
			},
			Format: format,
		},
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"generation_type": "text_to_image",
				"model":           config.Model,
				"prompt":          textInput.Content,
				"size":            imageGenReq.Size,
				"style":           imageGenReq.Style,
				"quality":         imageGenReq.Quality,
			},
		},
	}

	return output, nil
}

// processImageToImageGeneration 处理图像到图像生成
func (gp *GenerationProcessor) processImageToImageGeneration(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	imageInput, ok := input.Content.(*models.ImageInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid image input content")
	}

	// 构建图像编辑请求
	imageEditReq := &providers.ImageEditRequest{
		ImageURL: imageInput.URL,
		Prompt:   "Transform this image based on the style and content requirements",
		Size:     "1024x1024",
		Count:    1,
		UserID:   "",
	}

	// 调用AI提供者
	var imageEditResp *providers.ImageEditResponse
	err := gp.ProcessWithRetry(ctx, func() error {
		var err error
		imageEditResp, err = provider.EditImage(ctx, *imageEditReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("image editing failed: %w", err)
	}

	// 构建输出
	var imageURL string
	var width, height int
	var format string
	if len(imageEditResp.Images) > 0 {
		imageURL = imageEditResp.Images[0].URL
		width = imageEditResp.Images[0].Width
		height = imageEditResp.Images[0].Height
		format = imageEditResp.Images[0].Format
	} else {
		imageURL = ""
		width = 1024
		height = 1024
		format = "png"
	}

	output := models.MultimodalOutput{
		Type: models.OutputTypeImage,
		Content: &models.ImageOutput{
			URL: imageURL,
			Dimensions: models.ImageDimensions{
				Width:  width,
				Height: height,
			},
			Format: format,
		},
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"generation_type": "image_to_image",
				"model":           config.Model,
				"original_url":    imageInput.URL,
				"size":            imageEditReq.Size,
			},
		},
	}

	return output, nil
}

// processAudioGeneration 处理音频生成
func (gp *GenerationProcessor) processAudioGeneration(ctx context.Context, provider providers.AIProvider, input models.MultimodalInput, config models.MultimodalConfig) (models.MultimodalOutput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return models.MultimodalOutput{}, fmt.Errorf("invalid text input content")
	}

	// 使用聊天接口生成音频相关内容
	chatReq := &providers.ChatRequest{
		Messages: []providers.Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Generate audio content based on this text: %s", textInput.Content),
			},
		},
		UserID:      "",
		SessionID:   "",
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
	}

	// 调用AI提供者
	var chatResp *providers.ChatResponse
	err := gp.ProcessWithRetry(ctx, func() error {
		var err error
		chatResp, err = provider.Chat(ctx, *chatReq)
		return err
	}, 3)

	if err != nil {
		return models.MultimodalOutput{}, fmt.Errorf("audio generation failed: %w", err)
	}

	// 构建输出
	output := models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content: chatResp.Message.Content,
		},
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"generation_type": "audio_script",
				"model":           config.Model,
				"tokens_used":     fmt.Sprintf("%d", chatResp.Usage.TotalTokens),
				"original_text":   textInput.Content,
			},
		},
	}

	return output, nil
}