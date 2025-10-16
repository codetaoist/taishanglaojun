package multimodal

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
)

// DefaultInputValidator 默认输入验证器
type DefaultInputValidator struct{}

// NewDefaultInputValidator 创建默认输入验证器
func NewDefaultInputValidator() *DefaultInputValidator {
	return &DefaultInputValidator{}
}

// Validate 验证输入
func (v *DefaultInputValidator) Validate(inputs []models.MultimodalInput, config models.MultimodalConfig) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no inputs provided")
	}

	// 验证每个输入
	for i, input := range inputs {
		if err := v.validateInput(input, i); err != nil {
			return fmt.Errorf("input %d validation failed: %w", i, err)
		}
	}

	// 验证配置
	if err := v.validateConfig(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

// validateInput 验证单个输入
func (v *DefaultInputValidator) validateInput(input models.MultimodalInput, index int) error {
	// 验证输入类型
	if input.Type == "" {
		return fmt.Errorf("input type is required")
	}

	// 验证输入内容
	if input.Content == nil {
		return fmt.Errorf("input content is required")
	}

	// 根据类型验证具体内容
	switch input.Type {
	case models.InputTypeText:
		return v.validateTextInput(input.Content)
	case models.InputTypeImage:
		return v.validateImageInput(input.Content)
	case models.InputTypeAudio:
		return v.validateAudioInput(input.Content)
	case models.InputTypeVideo:
		return v.validateVideoInput(input.Content)
	default:
		return fmt.Errorf("unsupported input type: %s", input.Type)
	}
}

// validateTextInput 验证文本输入
func (v *DefaultInputValidator) validateTextInput(content interface{}) error {
	textInput, ok := content.(*models.TextInput)
	if !ok {
		return fmt.Errorf("invalid text input type")
	}

	if textInput.Content == "" {
		return fmt.Errorf("text content is required")
	}

	// 检查文本长度
	if len(textInput.Content) > 100000 { // 100KB限制
		return fmt.Errorf("text content too long (max 100KB)")
	}

	return nil
}

// validateImageInput 验证图像输入
func (v *DefaultInputValidator) validateImageInput(content interface{}) error {
	imageInput, ok := content.(*models.ImageInput)
	if !ok {
		return fmt.Errorf("invalid image input type")
	}

	if imageInput.URL == "" {
		return fmt.Errorf("image URL is required")
	}

	// 验证URL格式
	if err := v.validateURL(imageInput.URL); err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}

	// 验证图像格式
	if imageInput.Format != "" {
		validFormats := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"}
		if !v.isValidFormat(imageInput.Format, validFormats) {
			return fmt.Errorf("unsupported image format: %s", imageInput.Format)
		}
	}

	// 验证图像尺寸
	if imageInput.Dimensions.Width < 0 || imageInput.Dimensions.Height < 0 {
		return fmt.Errorf("invalid image dimensions")
	}

	// 检查图像尺寸限制
	if imageInput.Dimensions.Width > 10000 || imageInput.Dimensions.Height > 10000 {
		return fmt.Errorf("image dimensions too large (max 10000x10000)")
	}

	return nil
}

// validateAudioInput 验证音频输入
func (v *DefaultInputValidator) validateAudioInput(content interface{}) error {
	audioInput, ok := content.(*models.AudioInput)
	if !ok {
		return fmt.Errorf("invalid audio input type")
	}

	if audioInput.URL == "" {
		return fmt.Errorf("audio URL is required")
	}

	// 验证URL格式
	if err := v.validateURL(audioInput.URL); err != nil {
		return fmt.Errorf("invalid audio URL: %w", err)
	}

	// 验证音频格式
	if audioInput.Format != "" {
		validFormats := []string{"mp3", "wav", "flac", "aac", "ogg", "m4a"}
		if !v.isValidFormat(audioInput.Format, validFormats) {
			return fmt.Errorf("unsupported audio format: %s", audioInput.Format)
		}
	}

	// 验证音频时长
	if audioInput.Duration < 0 {
		return fmt.Errorf("invalid audio duration")
	}

	// 检查音频时长限制（最大1小时）
	if audioInput.Duration > 3600 {
		return fmt.Errorf("audio duration too long (max 1 hour)")
	}

	return nil
}

// validateVideoInput 验证视频输入
func (v *DefaultInputValidator) validateVideoInput(content interface{}) error {
	videoInput, ok := content.(*models.VideoInput)
	if !ok {
		return fmt.Errorf("invalid video input type")
	}

	if videoInput.URL == "" {
		return fmt.Errorf("video URL is required")
	}

	// 验证URL格式
	if err := v.validateURL(videoInput.URL); err != nil {
		return fmt.Errorf("invalid video URL: %w", err)
	}

	// 验证视频格式
	if videoInput.Format != "" {
		validFormats := []string{"mp4", "avi", "mov", "wmv", "flv", "webm", "mkv"}
		if !v.isValidFormat(videoInput.Format, validFormats) {
			return fmt.Errorf("unsupported video format: %s", videoInput.Format)
		}
	}

	// 验证视频时长
	if videoInput.Duration < 0 {
		return fmt.Errorf("invalid video duration")
	}

	// 检查视频时长限制（最大2小时）
	if videoInput.Duration > 7200 {
		return fmt.Errorf("video duration too long (max 2 hours)")
	}

	// 验证视频分辨率
	if videoInput.Dimensions.Width <= 0 || videoInput.Dimensions.Height <= 0 {
		return fmt.Errorf("video dimensions are required and must be positive")
	}

	return nil
}

// validateConfig 验证配置
func (v *DefaultInputValidator) validateConfig(config models.MultimodalConfig) error {
	// 验证模型名称
	if config.Model == "" {
		return fmt.Errorf("model is required")
	}

	// 验证最大令牌数
	if config.MaxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}

	if config.MaxTokens > 100000 {
		return fmt.Errorf("max tokens too large (max 100000)")
	}

	// 验证温度参数
	if config.Temperature < 0 || config.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}

	// 验证期望输出类型
	if len(config.ExpectedOutputs) > 0 {
		for _, outputType := range config.ExpectedOutputs {
			if !v.isValidOutputType(outputType) {
				return fmt.Errorf("invalid expected output type: %s", outputType)
			}
		}
	}

	return nil
}

// validateURL 验证URL格式
func (v *DefaultInputValidator) validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL is empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL scheme is required")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" && parsedURL.Scheme != "file" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	return nil
}

// isValidFormat 检查格式是否有效
func (v *DefaultInputValidator) isValidFormat(format string, validFormats []string) bool {
	format = strings.ToLower(format)
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}

// isValidResolution 检查分辨率格式是否有效
func (v *DefaultInputValidator) isValidResolution(resolution string) bool {
	// 支持格式：1920x1080, 1280x720等
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return false
	}

	// 简单检查是否为数字格式
	for _, part := range parts {
		if part == "" {
			return false
		}
		// 这里可以添加更严格的数字验证
	}

	return true
}

// isValidOutputType 检查输出类型是否有效
func (v *DefaultInputValidator) isValidOutputType(outputType models.OutputType) bool {
	validTypes := []models.OutputType{
		models.OutputTypeText,
		models.OutputTypeImage,
		models.OutputTypeAudio,
		models.OutputTypeVideo,
	}

	for _, validType := range validTypes {
		if outputType == validType {
			return true
		}
	}

	return false
}

// ValidateInputCount 验证输入数量
func (v *DefaultInputValidator) ValidateInputCount(inputs []models.MultimodalInput, maxCount int) error {
	if len(inputs) > maxCount {
		return fmt.Errorf("too many inputs: %d (max %d)", len(inputs), maxCount)
	}
	return nil
}

// ValidateInputSize 验证输入总大小
func (v *DefaultInputValidator) ValidateInputSize(inputs []models.MultimodalInput, maxSizeMB int) error {
	// 这里可以实现更复杂的大小计算逻辑
	// 目前只是一个简单的示例
	totalSize := 0
	for _, input := range inputs {
		switch input.Type {
		case models.InputTypeText:
			if textInput, ok := input.Content.(*models.TextInput); ok {
				totalSize += len(textInput.Content)
			}
		}
	}

	maxSizeBytes := maxSizeMB * 1024 * 1024
	if totalSize > maxSizeBytes {
		return fmt.Errorf("total input size too large: %d bytes (max %d MB)", totalSize, maxSizeMB)
	}

	return nil
}

// ValidateRequest 验证请求
func (v *DefaultInputValidator) ValidateRequest(req *models.MultimodalRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if len(req.Inputs) == 0 {
		return fmt.Errorf("no inputs provided in request")
	}

	// 验证输入
	for i, input := range req.Inputs {
		if err := v.validateInput(input, i); err != nil {
			return fmt.Errorf("input %d validation failed: %w", i, err)
		}
	}

	// 验证配置
	if err := v.validateConfig(req.Config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

// ValidateInput 验证单个输入（接口要求的方法）
func (v *DefaultInputValidator) ValidateInput(input interface{}, inputType string) error {
	// 将 interface{} 转换为 MultimodalInput
	multimodalInput, ok := input.(models.MultimodalInput)
	if !ok {
		return fmt.Errorf("invalid input type, expected MultimodalInput")
	}

	// 验证输入类型是否匹配
	if string(multimodalInput.Type) != inputType {
		return fmt.Errorf("input type mismatch: expected %s, got %s", inputType, multimodalInput.Type)
	}

	// 使用现有的验证逻辑
	return v.validateInput(multimodalInput, 0)
}