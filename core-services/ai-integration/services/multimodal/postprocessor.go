package multimodal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
)

// DefaultOutputPostprocessor 默认输出后处理器
type DefaultOutputPostprocessor struct{}

// NewDefaultOutputPostprocessor 创建默认输出后处理器
func NewDefaultOutputPostprocessor() *DefaultOutputPostprocessor {
	return &DefaultOutputPostprocessor{}
}

// PostprocessOutputs 后处理输出
func (p *DefaultOutputPostprocessor) PostprocessOutputs(ctx context.Context, outputs []models.MultimodalOutput, expectedOutputs []models.OutputType) ([]models.MultimodalOutput, error) {
	var processedOutputs []models.MultimodalOutput

	for i, output := range outputs {
		processedOutput, err := p.postprocessOutput(ctx, output, expectedOutputs, i)
		if err != nil {
			return nil, fmt.Errorf("postprocessing output %d failed: %w", i, err)
		}
		processedOutputs = append(processedOutputs, processedOutput)
	}

	// 过滤和排序输出
	filteredOutputs := p.filterOutputsByExpectedOutputs(processedOutputs, expectedOutputs)
	
	// 合并相似输出
	mergedOutputs := p.mergeSimilarOutputs(filteredOutputs)

	return mergedOutputs, nil
}

// postprocessOutput 后处理单个输出
func (p *DefaultOutputPostprocessor) postprocessOutput(ctx context.Context, output models.MultimodalOutput, expectedOutputs []models.OutputType, index int) (models.MultimodalOutput, error) {
	switch output.Type {
	case models.OutputTypeText:
		return p.postprocessTextOutput(ctx, output)
	case models.OutputTypeImage:
		return p.postprocessImageOutput(ctx, output)
	case models.OutputTypeAudio:
		return p.postprocessAudioOutput(ctx, output)
	case models.OutputTypeVideo:
		return p.postprocessVideoOutput(ctx, output)
	default:
		return output, nil // 不支持的类型直接返回
	}
}

// postprocessTextOutput 后处理文本输出
func (p *DefaultOutputPostprocessor) postprocessTextOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	textOutput, ok := output.Content.(*models.TextOutput)
	if !ok {
		return output, fmt.Errorf("invalid text output type")
	}

	// 创建副本
	processedTextOutput := &models.TextOutput{
		Content:  textOutput.Content,
		Language: textOutput.Language,
		Format:   textOutput.Format,
	}

	// 清理和格式化文本
	processedTextOutput.Content = p.cleanAndFormatText(processedTextOutput.Content)

	// 检测语言（如果未指定）
	if processedTextOutput.Language == "" {
		processedTextOutput.Language = p.detectLanguage(processedTextOutput.Content)
	}

	// 设置默认格式
	if processedTextOutput.Format == "" {
		processedTextOutput.Format = "plain"
	}

	// 创建处理后的输出
	processedOutput := models.MultimodalOutput{
		Type:    output.Type,
		Content: processedTextOutput,
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"postprocessed":      "true",
				"original_length":    fmt.Sprintf("%d", len(textOutput.Content)),
				"processed_length":   fmt.Sprintf("%d", len(processedTextOutput.Content)),
				"word_count":         fmt.Sprintf("%d", p.countWords(processedTextOutput.Content)),
				"character_count":    fmt.Sprintf("%d", len(processedTextOutput.Content)),
			},
		},
	}

	// 复制原有的自定义元数据
	if output.Metadata.Custom != nil {
		for k, v := range output.Metadata.Custom {
			processedOutput.Metadata.Custom[k] = v
		}
	}

	return processedOutput, nil
}

// postprocessImageOutput 后处理图像输出
func (p *DefaultOutputPostprocessor) postprocessImageOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	imageOutput, ok := output.Content.(*models.ImageOutput)
	if !ok {
		return output, fmt.Errorf("invalid image output type")
	}

	// 创建副本
	processedImageOutput := &models.ImageOutput{
		URL:         imageOutput.URL,
		Format:      imageOutput.Format,
		Dimensions:  imageOutput.Dimensions,
		Description: imageOutput.Description,
	}

	// 清理描述
	if processedImageOutput.Description != "" {
		processedImageOutput.Description = p.cleanAndFormatText(processedImageOutput.Description)
	}

	// 验证图像URL
	if processedImageOutput.URL != "" {
		processedImageOutput.URL = p.normalizeURL(processedImageOutput.URL)
	}

	// 创建处理后的输出
	processedOutput := models.MultimodalOutput{
		Type:    output.Type,
		Content: processedImageOutput,
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"postprocessed":    "true",
				"postprocessed_at": time.Now().Format(time.RFC3339),
				"original_url":     imageOutput.URL,
				"normalized_url":   processedImageOutput.URL,
			},
		},
	}

	// 复制原始元数据
	if output.Metadata.Custom != nil {
		for k, v := range output.Metadata.Custom {
			processedOutput.Metadata.Custom[k] = v
		}
	}
	processedOutput.Metadata.Custom["aspect_ratio"] = p.calculateAspectRatio(processedImageOutput.Dimensions.Width, processedImageOutput.Dimensions.Height)

	return processedOutput, nil
}

// postprocessAudioOutput 后处理音频输出
func (p *DefaultOutputPostprocessor) postprocessAudioOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	audioOutput, ok := output.Content.(*models.AudioOutput)
	if !ok {
		return output, fmt.Errorf("invalid audio output type")
	}

	// 创建副本
	processedAudioOutput := &models.AudioOutput{
		URL:        audioOutput.URL,
		Format:     audioOutput.Format,
		Duration:   audioOutput.Duration,
		SampleRate: audioOutput.SampleRate,
		Language:   audioOutput.Language,
	}

	// 验证音频URL
	if processedAudioOutput.URL != "" {
		processedAudioOutput.URL = p.normalizeURL(processedAudioOutput.URL)
	}

	// 创建处理后的输出
	processedOutput := models.MultimodalOutput{
		Type:    output.Type,
		Content: processedAudioOutput,
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"postprocessed":      "true",
				"postprocessed_at":   time.Now().Format(time.RFC3339),
				"original_duration":  fmt.Sprintf("%.2f", audioOutput.Duration),
				"formatted_duration": fmt.Sprintf("%.2f", processedAudioOutput.Duration),
			},
		},
	}

	// 复制原始元数据
	if output.Metadata.Custom != nil {
		for k, v := range output.Metadata.Custom {
			processedOutput.Metadata.Custom[k] = v
		}
	}
	processedOutput.Metadata.Custom["duration_formatted"] = p.formatDuration(processedAudioOutput.Duration)

	return processedOutput, nil
}

// postprocessVideoOutput 后处理视频输出
func (p *DefaultOutputPostprocessor) postprocessVideoOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	videoOutput, ok := output.Content.(*models.VideoOutput)
	if !ok {
		return output, fmt.Errorf("invalid video output type")
	}

	// 创建副本
	processedVideoOutput := &models.VideoOutput{
		URL:        videoOutput.URL,
		Format:     videoOutput.Format,
		Duration:   videoOutput.Duration,
		Dimensions: videoOutput.Dimensions,
		FrameRate:  videoOutput.FrameRate,
	}

	// 视频输出没有 Description 字段，跳过描述处理

	// 验证视频URL
	if processedVideoOutput.URL != "" {
		processedVideoOutput.URL = p.normalizeURL(processedVideoOutput.URL)
	}

	// 视频输出使用 Dimensions 字段而不是 Resolution，跳过分辨率处理

	// 创建处理后的输出
	processedOutput := models.MultimodalOutput{
		Type:    output.Type,
		Content: processedVideoOutput,
		Metadata: models.OutputMetadata{
			Custom: map[string]string{
				"postprocessed":      "true",
				"postprocessed_at":   time.Now().Format(time.RFC3339),
				"original_duration":  fmt.Sprintf("%.2f", videoOutput.Duration),
				"formatted_duration": fmt.Sprintf("%.2f", processedVideoOutput.Duration),
			},
		},
	}

	// 复制原始元数据
	if output.Metadata.Custom != nil {
		for k, v := range output.Metadata.Custom {
			processedOutput.Metadata.Custom[k] = v
		}
	}
	processedOutput.Metadata.Custom["duration_formatted"] = p.formatDuration(processedVideoOutput.Duration)

	return processedOutput, nil
}

// cleanAndFormatText 清理和格式化文本
func (p *DefaultOutputPostprocessor) cleanAndFormatText(text string) string {
	// 移除多余的空白字符
	text = strings.TrimSpace(text)
	
	// 替换多个连续的空格为单个空格
	text = strings.Join(strings.Fields(text), " ")
	
	// 移除多余的换行符
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}
	text = strings.Join(cleanedLines, "\n")
	
	// 修复常见的标点符号问题
	text = p.fixPunctuation(text)
	
	return text
}

// fixPunctuation 修复标点符号
func (p *DefaultOutputPostprocessor) fixPunctuation(text string) string {
	// 移除句号前的空格
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")
	text = strings.ReplaceAll(text, " !", "!")
	text = strings.ReplaceAll(text, " ?", "?")
	text = strings.ReplaceAll(text, " :", ":")
	text = strings.ReplaceAll(text, " ;", ";")
	
	// 确保句号后有空格
	text = strings.ReplaceAll(text, ".", ". ")
	text = strings.ReplaceAll(text, ",", ", ")
	text = strings.ReplaceAll(text, "!", "! ")
	text = strings.ReplaceAll(text, "?", "? ")
	text = strings.ReplaceAll(text, ":", ": ")
	text = strings.ReplaceAll(text, ";", "; ")
	
	// 移除多余的空格
	text = strings.Join(strings.Fields(text), " ")
	
	return text
}

// detectLanguage 检测语言
func (p *DefaultOutputPostprocessor) detectLanguage(text string) string {
	// 检测中文字符
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return "zh"
		}
	}
	
	// 默认为英语
	return "en"
}

// countWords 计算单词数
func (p *DefaultOutputPostprocessor) countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// normalizeURL 标准化URL
func (p *DefaultOutputPostprocessor) normalizeURL(url string) string {
	return strings.TrimSpace(url)
}

// normalizeResolution 标准化分辨率
func (p *DefaultOutputPostprocessor) normalizeResolution(resolution string) string {
	resolution = strings.ToLower(strings.ReplaceAll(resolution, " ", ""))
	resolution = strings.ReplaceAll(resolution, "*", "x")
	resolution = strings.ReplaceAll(resolution, "×", "x")
	return resolution
}

// calculateAspectRatio 计算宽高比
func (p *DefaultOutputPostprocessor) calculateAspectRatio(width, height int) string {
	if width == 0 || height == 0 {
		return "unknown"
	}
	
	// 计算最大公约数
	gcd := p.gcd(width, height)
	aspectWidth := width / gcd
	aspectHeight := height / gcd
	
	return fmt.Sprintf("%d:%d", aspectWidth, aspectHeight)
}

// gcd 计算最大公约数
func (p *DefaultOutputPostprocessor) gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// formatDuration 格式化时长
func (p *DefaultOutputPostprocessor) formatDuration(duration float64) string {
	if duration <= 0 {
		return "0s"
	}
	
	hours := int(duration) / 3600
	minutes := (int(duration) % 3600) / 60
	seconds := int(duration) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}



// filterOutputsByExpectedOutputs 根据期望输出过滤输出
func (p *DefaultOutputPostprocessor) filterOutputsByExpectedOutputs(outputs []models.MultimodalOutput, expectedOutputs []models.OutputType) []models.MultimodalOutput {
	if len(expectedOutputs) == 0 {
		return outputs // 如果没有指定期望输出，返回所有输出
	}
	
	var filtered []models.MultimodalOutput
	for _, output := range outputs {
		for _, expectedOutput := range expectedOutputs {
			if output.Type == expectedOutput {
				filtered = append(filtered, output)
				break
			}
		}
	}
	
	return filtered
}

// mergeSimilarOutputs 合并相似输出
func (p *DefaultOutputPostprocessor) mergeSimilarOutputs(outputs []models.MultimodalOutput) []models.MultimodalOutput {
	if len(outputs) <= 1 {
		return outputs
	}
	
	// 这里可以实现更复杂的合并逻辑
	// 目前只是简单地返回原始输出
	return outputs
}

// ValidateOutput 验证输出
func (p *DefaultOutputPostprocessor) ValidateOutput(output models.MultimodalOutput) error {
	if output.Content == nil {
		return fmt.Errorf("output content is nil")
	}
	
	switch output.Type {
	case models.OutputTypeText:
		if _, ok := output.Content.(*models.TextOutput); !ok {
			return fmt.Errorf("invalid text output content type")
		}
	case models.OutputTypeImage:
		if _, ok := output.Content.(*models.ImageOutput); !ok {
			return fmt.Errorf("invalid image output content type")
		}
	case models.OutputTypeAudio:
		if _, ok := output.Content.(*models.AudioOutput); !ok {
			return fmt.Errorf("invalid audio output content type")
		}
	case models.OutputTypeVideo:
		if _, ok := output.Content.(*models.VideoOutput); !ok {
			return fmt.Errorf("invalid video output content type")
		}
	default:
		return fmt.Errorf("unsupported output type: %s", output.Type)
	}
	
	return nil
}

// ConvertOutput 转换输出（接口要求的方法）
func (p *DefaultOutputPostprocessor) ConvertOutput(ctx context.Context, output models.MultimodalOutput, targetType models.MultimodalOutputType) (models.MultimodalOutput, error) {
	// 根据目标类型进行转换
	switch targetType {
	case models.OutputTypeText:
		return p.convertToTextOutput(ctx, output)
	case models.OutputTypeImage:
		return p.convertToImageOutput(ctx, output)
	case models.OutputTypeAudio:
		return p.convertToAudioOutput(ctx, output)
	case models.OutputTypeVideo:
		return p.convertToVideoOutput(ctx, output)
	default:
		// 如果不需要特殊转换，直接返回原输出
		return output, nil
	}
}

// convertToTextOutput 转换为文本输出
func (p *DefaultOutputPostprocessor) convertToTextOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	if output.Type == models.OutputTypeText {
		return output, nil // 已经是文本类型，直接返回
	}

	// 创建新的文本输出
	textContent := ""
	switch output.Type {
	case models.OutputTypeImage:
		if imageOutput, ok := output.Content.(*models.ImageOutput); ok {
			textContent = fmt.Sprintf("Image: %s (Format: %s, Size: %dx%d)", 
				imageOutput.URL, imageOutput.Format, imageOutput.Dimensions.Width, imageOutput.Dimensions.Height)
		}
	case models.OutputTypeAudio:
		if audioOutput, ok := output.Content.(*models.AudioOutput); ok {
			textContent = fmt.Sprintf("Audio: %s (Format: %s, Duration: %.2fs)", 
				audioOutput.URL, audioOutput.Format, audioOutput.Duration)
		}
	case models.OutputTypeVideo:
		if videoOutput, ok := output.Content.(*models.VideoOutput); ok {
			textContent = fmt.Sprintf("Video: %s (Format: %s, Duration: %.2fs, Size: %dx%d)", 
				videoOutput.URL, videoOutput.Format, videoOutput.Duration, videoOutput.Dimensions.Width, videoOutput.Dimensions.Height)
		}
	default:
		textContent = fmt.Sprintf("Output of type: %s", output.Type)
	}

	return models.MultimodalOutput{
		Type: models.OutputTypeText,
		Content: &models.TextOutput{
			Content:  textContent,
			Language: "en",
			Format:   "plain",
		},
		Metadata: output.Metadata,
	}, nil
}

// convertToImageOutput 转换为图像输出
func (p *DefaultOutputPostprocessor) convertToImageOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	if output.Type == models.OutputTypeImage {
		return output, nil // 已经是图像类型，直接返回
	}

	// 对于非图像类型，无法直接转换，返回错误
	return output, fmt.Errorf("cannot convert %s output to image", output.Type)
}

// convertToAudioOutput 转换为音频输出
func (p *DefaultOutputPostprocessor) convertToAudioOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	if output.Type == models.OutputTypeAudio {
		return output, nil // 已经是音频类型，直接返回
	}

	// 对于非音频类型，无法直接转换，返回错误
	return output, fmt.Errorf("cannot convert %s output to audio", output.Type)
}

// convertToVideoOutput 转换为视频输出
func (p *DefaultOutputPostprocessor) convertToVideoOutput(ctx context.Context, output models.MultimodalOutput) (models.MultimodalOutput, error) {
	if output.Type == models.OutputTypeVideo {
		return output, nil // 已经是视频类型，直接返回
	}

	// 对于非视频类型，无法直接转换，返回错误
	return output, fmt.Errorf("cannot convert %s output to video", output.Type)
}