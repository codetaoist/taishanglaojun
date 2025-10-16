package multimodal

import (
	"context"
	"fmt"
	"strings"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
)

// DefaultInputPreprocessor 默认输入预处理器
type DefaultInputPreprocessor struct{}

// NewDefaultInputPreprocessor 创建默认输入预处理器
func NewDefaultInputPreprocessor() *DefaultInputPreprocessor {
	return &DefaultInputPreprocessor{}
}

// PreprocessInputs 预处理输入
func (p *DefaultInputPreprocessor) PreprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error) {
	var processedInputs []models.MultimodalInput

	for i, input := range inputs {
		processedInput, err := p.preprocessInput(ctx, input, i)
		if err != nil {
			return nil, fmt.Errorf("preprocessing input %d failed: %w", i, err)
		}
		processedInputs = append(processedInputs, processedInput)
	}

	return processedInputs, nil
}

// preprocessInput 预处理单个输入
func (p *DefaultInputPreprocessor) preprocessInput(ctx context.Context, input models.MultimodalInput, index int) (models.MultimodalInput, error) {
	switch input.Type {
	case models.InputTypeText:
		return p.preprocessTextInput(ctx, input)
	case models.InputTypeImage:
		return p.preprocessImageInput(ctx, input)
	case models.InputTypeAudio:
		return p.preprocessAudioInput(ctx, input)
	case models.InputTypeVideo:
		return p.preprocessVideoInput(ctx, input)
	default:
		return input, nil // 不支持的类型直接返回
	}
}

// preprocessTextInput 预处理文本输入
func (p *DefaultInputPreprocessor) preprocessTextInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	textInput, ok := input.Content.(*models.TextInput)
	if !ok {
		return input, fmt.Errorf("invalid text input type")
	}

	// 创建副本以避免修改原始输入
	processedTextInput := &models.TextInput{
		Content:  textInput.Content,
		Language: textInput.Language,
		Format:   textInput.Format,
	}

	// 清理和标准化文本
	processedTextInput.Content = p.cleanText(processedTextInput.Content)

	// 检测语言（如果未指定）
	if processedTextInput.Language == "" {
		processedTextInput.Language = p.detectLanguage(processedTextInput.Content)
	}

	// 设置默认格式
	if processedTextInput.Format == "" {
		processedTextInput.Format = "plain"
	}

	// 创建处理后的输入
	processedInput := models.MultimodalInput{
		Type:     input.Type,
		Content:  processedTextInput,
		Metadata: p.copyMetadata(input.Metadata),
	}

	// 添加预处理元数据
	if processedInput.Metadata.Custom == nil {
		processedInput.Metadata.Custom = make(map[string]string)
	}
	processedInput.Metadata.Custom["preprocessed"] = "true"
	processedInput.Metadata.Custom["original_length"] = fmt.Sprintf("%d", len(textInput.Content))
	processedInput.Metadata.Custom["processed_length"] = fmt.Sprintf("%d", len(processedTextInput.Content))

	return processedInput, nil
}

// preprocessImageInput 预处理图像输入
func (p *DefaultInputPreprocessor) preprocessImageInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	imageInput, ok := input.Content.(*models.ImageInput)
	if !ok {
		return input, fmt.Errorf("invalid image input type")
	}

	// 创建副本
	processedImageInput := &models.ImageInput{
		URL:         imageInput.URL,
		Format:      imageInput.Format,
		Dimensions:  imageInput.Dimensions,
		Description: imageInput.Description,
	}

	// 标准化URL
	processedImageInput.URL = p.normalizeURL(processedImageInput.URL)

	// 推断格式（如果未指定）
	if processedImageInput.Format == "" {
		processedImageInput.Format = p.inferImageFormat(processedImageInput.URL)
	}

	// 清理描述
	if processedImageInput.Description != "" {
		processedImageInput.Description = p.cleanText(processedImageInput.Description)
	}

	// 创建处理后的输入
	processedInput := models.MultimodalInput{
		Type:     input.Type,
		Content:  processedImageInput,
		Metadata: p.copyMetadata(input.Metadata),
	}

	// 添加预处理元数据
	if processedInput.Metadata.Custom == nil {
		processedInput.Metadata.Custom = make(map[string]string)
	}
	processedInput.Metadata.Custom["preprocessed"] = "true"
	processedInput.Metadata.Custom["inferred_format"] = processedImageInput.Format

	return processedInput, nil
}

// PreprocessAudioInput 预处理音频输入（接口要求的方法）
func (p *DefaultInputPreprocessor) PreprocessAudioInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	// 验证输入类型是否为音频
	if input.Type != models.InputTypeAudio {
		return input, fmt.Errorf("input type is not audio")
	}

	// 调用现有的预处理逻辑
	processedInput, err := p.preprocessAudioInput(ctx, input)
	if err != nil {
		return input, fmt.Errorf("failed to preprocess audio input: %w", err)
	}

	return processedInput, nil
}

// preprocessAudioInput 预处理音频输入
func (p *DefaultInputPreprocessor) preprocessAudioInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	audioInput, ok := input.Content.(*models.AudioInput)
	if !ok {
		return input, fmt.Errorf("invalid audio input type")
	}

	// 创建副本
	processedAudioInput := &models.AudioInput{
		URL:      audioInput.URL,
		Format:   audioInput.Format,
		Duration: audioInput.Duration,
		Language: audioInput.Language,
	}

	// 标准化URL
	processedAudioInput.URL = p.normalizeURL(processedAudioInput.URL)

	// 推断格式（如果未指定）
	if processedAudioInput.Format == "" {
		processedAudioInput.Format = p.inferAudioFormat(processedAudioInput.URL)
	}

	// 设置默认语言（如果未指定）
	if processedAudioInput.Language == "" {
		processedAudioInput.Language = "auto" // 自动检测语言
	}

	// 创建处理后的输入
	processedInput := models.MultimodalInput{
		Type:     input.Type,
		Content:  processedAudioInput,
		Metadata: p.copyMetadata(input.Metadata),
	}

	// 添加预处理元数据
	if processedInput.Metadata.Custom == nil {
		processedInput.Metadata.Custom = make(map[string]string)
	}
	processedInput.Metadata.Custom["preprocessed"] = "true"
	processedInput.Metadata.Custom["inferred_format"] = processedAudioInput.Format

	return processedInput, nil
}

// preprocessVideoInput 预处理视频输入
func (p *DefaultInputPreprocessor) preprocessVideoInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	videoInput, ok := input.Content.(*models.VideoInput)
	if !ok {
		return input, fmt.Errorf("invalid video input type")
	}

	// 创建副本
	processedVideoInput := &models.VideoInput{
		URL:        videoInput.URL,
		Format:     videoInput.Format,
		Duration:   videoInput.Duration,
		Dimensions: videoInput.Dimensions,
		FrameRate:  videoInput.FrameRate,
	}

	// 标准化URL
	processedVideoInput.URL = p.normalizeURL(processedVideoInput.URL)

	// 推断格式（如果未指定）
	if processedVideoInput.Format == "" {
		processedVideoInput.Format = p.inferVideoFormat(processedVideoInput.URL)
	}

	// 设置默认帧率
	if processedVideoInput.FrameRate == 0 {
		processedVideoInput.FrameRate = 30 // 默认30fps
	}

	// 创建处理后的输入
	processedInput := models.MultimodalInput{
		Type:     input.Type,
		Content:  processedVideoInput,
		Metadata: p.copyMetadata(input.Metadata),
	}

	// 添加预处理元数据
	if processedInput.Metadata.Custom == nil {
		processedInput.Metadata.Custom = make(map[string]string)
	}
	processedInput.Metadata.Custom["preprocessed"] = "true"
	processedInput.Metadata.Custom["inferred_format"] = processedVideoInput.Format

	return processedInput, nil
}

// cleanText 清理文本
func (p *DefaultInputPreprocessor) cleanText(text string) string {
	// 移除多余的空白字符
	text = strings.TrimSpace(text)
	
	// 替换多个连续的空格为单个空格
	text = strings.Join(strings.Fields(text), " ")
	
	// 移除控制字符（保留换行符和制表符）
	var cleaned strings.Builder
	for _, r := range text {
		if r >= 32 || r == '\n' || r == '\t' {
			cleaned.WriteRune(r)
		}
	}
	
	return cleaned.String()
}

// detectLanguage 检测语言（简单实现）
func (p *DefaultInputPreprocessor) detectLanguage(text string) string {
	// 这里是一个简单的语言检测实现
	// 在实际应用中，可以使用更复杂的语言检测库
	
	// 检测中文字符
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return "zh"
		}
	}
	
	// 默认为英语
	return "en"
}

// normalizeURL 标准化URL
func (p *DefaultInputPreprocessor) normalizeURL(url string) string {
	// 移除URL中的多余空格
	url = strings.TrimSpace(url)
	
	// 这里可以添加更多的URL标准化逻辑
	// 例如：统一协议、移除多余的斜杠等
	
	return url
}

// inferImageFormat 从URL推断图像格式
func (p *DefaultInputPreprocessor) inferImageFormat(url string) string {
	url = strings.ToLower(url)
	
	if strings.Contains(url, ".jpg") || strings.Contains(url, ".jpeg") {
		return "jpeg"
	}
	if strings.Contains(url, ".png") {
		return "png"
	}
	if strings.Contains(url, ".gif") {
		return "gif"
	}
	if strings.Contains(url, ".bmp") {
		return "bmp"
	}
	if strings.Contains(url, ".webp") {
		return "webp"
	}
	if strings.Contains(url, ".svg") {
		return "svg"
	}
	
	return "unknown"
}

// inferAudioFormat 从URL推断音频格式
func (p *DefaultInputPreprocessor) inferAudioFormat(url string) string {
	url = strings.ToLower(url)
	
	if strings.Contains(url, ".mp3") {
		return "mp3"
	}
	if strings.Contains(url, ".wav") {
		return "wav"
	}
	if strings.Contains(url, ".flac") {
		return "flac"
	}
	if strings.Contains(url, ".aac") {
		return "aac"
	}
	if strings.Contains(url, ".ogg") {
		return "ogg"
	}
	if strings.Contains(url, ".m4a") {
		return "m4a"
	}
	
	return "unknown"
}

// inferVideoFormat 从URL推断视频格式
func (p *DefaultInputPreprocessor) inferVideoFormat(url string) string {
	url = strings.ToLower(url)
	
	if strings.Contains(url, ".mp4") {
		return "mp4"
	}
	if strings.Contains(url, ".avi") {
		return "avi"
	}
	if strings.Contains(url, ".mov") {
		return "mov"
	}
	if strings.Contains(url, ".wmv") {
		return "wmv"
	}
	if strings.Contains(url, ".flv") {
		return "flv"
	}
	if strings.Contains(url, ".webm") {
		return "webm"
	}
	if strings.Contains(url, ".mkv") {
		return "mkv"
	}
	
	return "unknown"
}

// normalizeResolution 标准化分辨率格式
func (p *DefaultInputPreprocessor) normalizeResolution(resolution string) string {
	// 移除空格并转换为小写
	resolution = strings.ToLower(strings.ReplaceAll(resolution, " ", ""))
	
	// 标准化常见分辨率
	switch resolution {
	case "1920*1080", "1920×1080":
		return "1920x1080"
	case "1280*720", "1280×720":
		return "1280x720"
	case "3840*2160", "3840×2160":
		return "3840x2160"
	case "1366*768", "1366×768":
		return "1366x768"
	}
	
	// 替换其他分隔符为标准的'x'
	resolution = strings.ReplaceAll(resolution, "*", "x")
	resolution = strings.ReplaceAll(resolution, "×", "x")
	
	return resolution
}

// copyMetadata 复制元数据
func (p *DefaultInputPreprocessor) copyMetadata(metadata models.InputMetadata) models.InputMetadata {
	// 复制 Custom 字段
	var copiedCustom map[string]string
	if metadata.Custom != nil {
		copiedCustom = make(map[string]string)
		for k, v := range metadata.Custom {
			copiedCustom[k] = v
		}
	}
	
	return models.InputMetadata{
		MimeType:   metadata.MimeType,
		Size:       metadata.Size,
		Duration:   metadata.Duration,
		Dimensions: metadata.Dimensions,
		Language:   metadata.Language,
		Encoding:   metadata.Encoding,
		Quality:    metadata.Quality,
		Custom:     copiedCustom,
	}
}

// PreprocessTextInput 预处理文本输入（接口要求的方法）
func (p *DefaultInputPreprocessor) PreprocessTextInput(input models.MultimodalInput) (models.MultimodalInput, error) {
	// 验证输入类型是否为文本
	if input.Type != models.InputTypeText {
		return input, fmt.Errorf("input type is not text")
	}

	// 调用现有的预处理逻辑
	processedInput, err := p.preprocessTextInput(context.Background(), input)
	if err != nil {
		return input, fmt.Errorf("failed to preprocess text input: %w", err)
	}

	return processedInput, nil
}

// PreprocessImageInput 预处理图像输入（接口要求的方法）
func (p *DefaultInputPreprocessor) PreprocessImageInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	// 验证输入类型是否为图像
	if input.Type != models.InputTypeImage {
		return input, fmt.Errorf("input type is not image")
	}

	// 调用现有的预处理逻辑
	processedInput, err := p.preprocessImageInput(ctx, input)
	if err != nil {
		return input, fmt.Errorf("failed to preprocess image input: %w", err)
	}

	return processedInput, nil
}

// PreprocessVideoInput 预处理视频输入（接口要求的方法）
func (p *DefaultInputPreprocessor) PreprocessVideoInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	// 验证输入类型是否为视频
	if input.Type != models.InputTypeVideo {
		return input, fmt.Errorf("input type is not video")
	}

	// 调用现有的预处理逻辑
	processedInput, err := p.preprocessVideoInput(ctx, input)
	if err != nil {
		return input, fmt.Errorf("failed to preprocess video input: %w", err)
	}

	return processedInput, nil
}