package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"errors"
	"math"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/repositories"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/services/multimodal"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MultimodalServiceConfig 多模态服务配置（保持向后兼容）
type MultimodalServiceConfig struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	MaxRetryDelay time.Duration `json:"max_retry_delay"`
	Timeout       time.Duration `json:"timeout"`
}

// DefaultMultimodalServiceConfig 默认配置
func DefaultMultimodalServiceConfig() *MultimodalServiceConfig {
	return &MultimodalServiceConfig{
		MaxRetries:    3,
		RetryDelay:    time.Second,
		MaxRetryDelay: 30 * time.Second,
		Timeout:       60 * time.Second,
	}
}

// MultimodalError 多模态服务错误（保持向后兼容）
type MultimodalError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Retry   bool   `json:"retry"`
}

func (e *MultimodalError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Type, e.Code, e.Message)
}

// 错误类型常量
const (
	ErrorTypeValidation = "validation"
	ErrorTypeProvider   = "provider"
	ErrorTypeNetwork    = "network"
	ErrorTypeTimeout    = "timeout"
	ErrorTypeInternal   = "internal"
)

// MultimodalService 多模态服务（适配器模式）
type MultimodalService struct {
	// 新的模块化服务
	service *multimodal.Service
	
	// 保持向后兼容的字段
	aiProviders  map[string]providers.AIProvider
	aiProvider   providers.AIProvider // 默认AI提供者
	repository   repositories.Repository
	fileService  FileService
	audioService AudioService
	imageService ImageService
	videoService VideoService
	config       *MultimodalServiceConfig
}

// NewMultimodalService 创建新的多模态AI服务
func NewMultimodalService(
	aiProviders map[string]providers.AIProvider,
	defaultProvider providers.AIProvider,
	repository repositories.Repository,
	fileService FileService,
	audioService AudioService,
	imageService ImageService,
	videoService VideoService,
) *MultimodalService {
	return NewMultimodalServiceWithConfig(
		aiProviders,
		defaultProvider,
		repository,
		fileService,
		audioService,
		imageService,
		videoService,
		nil,
	)
}

// NewMultimodalServiceWithConfig 使用自定义配置创建多模态服务实例
func NewMultimodalServiceWithConfig(
	aiProviders map[string]providers.AIProvider,
	defaultProvider providers.AIProvider,
	repository repositories.Repository,
	fileService FileService,
	audioService AudioService,
	imageService ImageService,
	videoService VideoService,
	config *MultimodalServiceConfig,
) *MultimodalService {
	if config == nil {
		config = DefaultMultimodalServiceConfig()
	}

	// 转换配置到新的格式
	newConfig := multimodal.Config{
		MaxRetries:    config.MaxRetries,
		RetryDelay:    config.RetryDelay,
		MaxRetryDelay: config.MaxRetryDelay,
		Timeout:       config.Timeout,
	}

	// 创建新的模块化服务
	factory := multimodal.NewServiceFactory()
	service := factory.CreateServiceWithConfig(newConfig)

	// 注册AI提供者
	for name, provider := range aiProviders {
		service.RegisterProvider(name, provider)
	}
	if defaultProvider != nil {
		service.RegisterProvider("default", defaultProvider)
	}

	return &MultimodalService{
		service:      service,
		aiProviders:  aiProviders,
		aiProvider:   defaultProvider,
		repository:   repository,
		fileService:  fileService,
		audioService: audioService,
		imageService: imageService,
		videoService: videoService,
		config:       config,
	}
}

// ProcessMultimodalRequest 处理多模态请求（保持向后兼容）
func (s *MultimodalService) ProcessMultimodalRequest(ctx context.Context, req *models.MultimodalRequest) (*models.MultimodalResponse, error) {
	startTime := time.Now()
	
	if req == nil {
		return nil, &MultimodalError{
			Code:    "INVALID_REQUEST",
			Message: "request cannot be nil",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// 确定使用的提供者
	providerName := req.Config.Provider
	if providerName == "" {
		providerName = "default"
	}

	// 根据请求类型选择处理器
	var outputs []models.MultimodalOutput
	var err error

	switch req.Type {
	case "chat":
		outputs, err = s.service.Chat(ctx, providerName, req.Inputs, req.Config)
	case "analysis":
		outputs, err = s.service.Analyze(ctx, providerName, req.Inputs, req.Config)
	case "generation":
		outputs, err = s.service.Generate(ctx, providerName, req.Inputs, req.Config)
	case "translation":
		outputs, err = s.service.Translate(ctx, providerName, req.Inputs, req.Config)
	case "search":
		outputs, err = s.service.Search(ctx, providerName, req.Inputs, req.Config)
	default:
		return nil, &MultimodalError{
			Code:    "UNSUPPORTED_TYPE",
			Message: fmt.Sprintf("unsupported request type: %s", req.Type),
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	if err != nil {
		return nil, s.wrapProviderError(err, string(req.Type))
	}

	// 构建响应
	response := &models.MultimodalResponse{
		ID:        uuid.New().String(),
		RequestID: req.ID,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Type:      req.Type,
		Outputs:   outputs,
		CreatedAt: time.Now(),
		Status:    "completed",
		Metadata: models.ResponseMetadata{
			Provider:       providerName,
			ProcessingTime: time.Since(startTime),
			TokensUsed:     calculateTokenUsage(req.Inputs, outputs),
			Model:          req.Config.Model,
		},
	}

	return response, nil
}

// ProcessFileUpload 处理文件上传（保持向后兼容）
func (s *MultimodalService) ProcessFileUpload(ctx *gin.Context) (*models.MultimodalInput, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, &MultimodalError{
			Code:    "FILE_UPLOAD_ERROR",
			Message: fmt.Sprintf("failed to get uploaded file: %v", err),
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}
	defer file.Close()

	// 读取文件数据
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, &MultimodalError{
			Code:    "FILE_READ_ERROR",
			Message: fmt.Sprintf("failed to read file data: %v", err),
			Type:    ErrorTypeInternal,
			Retry:   false,
		}
	}

	// 确定输入类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	inputType := s.determineInputType(contentType)
	format := s.extractFormat(header.Filename)

	// 保存文件
	var url string
	if s.fileService != nil {
		url, err = s.fileService.SaveFile(ctx, data, header.Filename)
		if err != nil {
			return nil, &MultimodalError{
				Code:    "FILE_SAVE_ERROR",
				Message: fmt.Sprintf("failed to save file: %v", err),
				Type:    ErrorTypeInternal,
				Retry:   true,
			}
		}
	}

	// 创建多模态输入
	input := &models.MultimodalInput{
		Type:    inputType,
		Content: string(data),
		Metadata: models.InputMetadata{
			MimeType: contentType,
			Size:     int64(len(data)),
			Custom: map[string]string{
				"filename": header.Filename,
				"url":      url,
				"format":   format,
			},
		},
	}

	// 根据类型添加特定元数据
	switch inputType {
	case models.InputTypeImage:
		if s.imageService != nil {
			if dimensions, err := s.imageService.GetDimensions(data); err == nil {
				input.Metadata.Dimensions = dimensions
			}
		}
	case models.InputTypeAudio:
		if s.audioService != nil {
			if duration, err := s.audioService.GetDuration(data); err == nil {
				input.Metadata.Duration = duration
			}
		}
	case models.InputTypeVideo:
		if s.videoService != nil {
			if info, err := s.videoService.GetInfo(data); err == nil {
				input.Metadata.Duration = info.Duration.Seconds()
				input.Metadata.Dimensions = models.ImageDimensions{
					Width:  info.Width,
					Height: info.Height,
				}
				if input.Metadata.Custom == nil {
					input.Metadata.Custom = make(map[string]string)
				}
				input.Metadata.Custom["video_info"] = fmt.Sprintf("format:%s,bitrate:%d,codec:%s", info.Format, info.Bitrate, info.VideoCodec)
			}
		}
	}

	return input, nil
}

// StreamMultimodalResponse 流式处理多模态响应（保持向后兼容）
func (s *MultimodalService) StreamMultimodalResponse(ctx context.Context, req *models.MultimodalRequest, responseChan chan<- *models.MultimodalResponse) error {
	defer close(responseChan)

	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return err
	}

	// 确定使用的提供者
	providerName := req.Config.Provider
	if providerName == "" {
		providerName = "default"
	}

	// 处理请求
	outputs, err := s.ProcessMultimodalRequest(ctx, req)
	if err != nil {
		return err
	}

	// 发送响应
	select {
	case responseChan <- outputs:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// 图像相关方法（保持向后兼容）

// GenerateImage 生成图像
func (s *MultimodalService) GenerateImage(ctx context.Context, req *providers.ImageGenerateRequest) (*providers.ImageGenerateResponse, error) {
	if err := s.validateImageGenerateRequest(req); err != nil {
		return nil, err
	}

	// 转换为多模态输入
	inputs := []models.MultimodalInput{
		{
			Type:    models.InputTypeText,
			Content: req.Prompt,
			Metadata: models.InputMetadata{
				Custom: map[string]string{
					"style":   req.Style,
					"size":    req.Size,
					"quality": req.Quality,
				},
			},
		},
	}

	config := models.MultimodalConfig{
		Model:           req.Model,
		MaxTokens:       1000,
		Temperature:     0.7,
		ExpectedOutputs: []models.OutputType{models.OutputTypeImage},
	}

	// 使用生成处理器
	providerName := "default" // 使用默认提供者
	outputs, err := s.service.Generate(ctx, providerName, inputs, config)
	if err != nil {
		return nil, s.wrapProviderError(err, "image_generation")
	}

	if len(outputs) == 0 || outputs[0].Type != models.OutputTypeImage {
		return nil, &MultimodalError{
			Code:    "NO_IMAGE_OUTPUT",
			Message: "no image output generated",
			Type:    ErrorTypeInternal,
			Retry:   true,
		}
	}

	// 构建响应
	imageOutput, ok := outputs[0].Content.(*models.ImageOutput)
	if !ok {
		return nil, &MultimodalError{
			Code:    "INVALID_OUTPUT",
			Message: "expected image output",
			Type:    ErrorTypeInternal,
			Retry:   false,
		}
	}

	response := &providers.ImageGenerateResponse{
		Images: []providers.GeneratedImage{
			{
				URL:    imageOutput.URL,
				Format: imageOutput.Format,
				Width:  imageOutput.Dimensions.Width,
				Height: imageOutput.Dimensions.Height,
			},
		},
		Usage: providers.Usage{
			TotalTokens: 100, // 默认token使用量
		},
	}

	// 记录日志
	s.logImageGeneration(ctx, req.UserID, req.Prompt, response)

	return response, nil
}

// AnalyzeImage 分析图像
func (s *MultimodalService) AnalyzeImage(ctx context.Context, req *providers.ImageAnalyzeRequest) (*providers.ImageAnalyzeResponse, error) {
	if err := s.validateImageAnalyzeRequest(req); err != nil {
		return nil, err
	}

	// 预处理图像数据
	if err := s.preprocessImageData(req); err != nil {
		return nil, err
	}

	// 转换为多模态输入
	imageInput := &models.ImageInput{
		Data: req.ImageData,
		URL:  req.ImageURL,
	}
	
	inputs := []models.MultimodalInput{
		{
			Type:    models.InputTypeImage,
			Content: imageInput,
			Metadata: models.InputMetadata{
				MimeType: "image/jpeg", // 默认MIME类型
				Custom: map[string]string{
					"url": req.ImageURL,
				},
			},
		},
	}

	if req.Prompt != "" {
		inputs = append(inputs, models.MultimodalInput{
			Type:    models.InputTypeText,
			Content: req.Prompt,
		})
	}

	config := models.MultimodalConfig{
		Model:           "gpt-4-vision-preview", // 默认模型
		MaxTokens:       1000,                   // 默认最大token数
		Temperature:     0.3,
		ExpectedOutputs: []models.OutputType{models.OutputTypeText},
	}

	// 使用分析处理器
	outputs, err := s.service.Analyze(ctx, "default", inputs, config)
	if err != nil {
		return nil, s.wrapProviderError(err, "image_analysis")
	}

	if len(outputs) == 0 || outputs[0].Type != models.OutputTypeText {
		return nil, &MultimodalError{
			Code:    "NO_TEXT_OUTPUT",
			Message: "no text output generated",
			Type:    ErrorTypeInternal,
			Retry:   true,
		}
	}

	// 从输出中提取文本内容
	textOutput, ok := outputs[0].Content.(*models.TextOutput)
	if !ok {
		return nil, &MultimodalError{
			Code:    "INVALID_OUTPUT_TYPE",
			Message: "expected text output from image analysis",
			Type:    ErrorTypeInternal,
			Retry:   false,
		}
	}

	// 构建响应
	response := &providers.ImageAnalyzeResponse{
		Description: textOutput.Content,
		Usage: providers.Usage{
			TotalTokens: 100, // 默认token使用量
		},
		RequestID: fmt.Sprintf("analyze_%d", time.Now().Unix()),
	}

	// 记录日志
	s.logImageAnalysis(ctx, req.UserID, req.Prompt, response)

	return response, nil
}

// EditImage 编辑图像
func (s *MultimodalService) EditImage(ctx context.Context, req *providers.ImageEditRequest) (*providers.ImageEditResponse, error) {
	if err := s.validateImageEditRequest(req); err != nil {
		return nil, err
	}

	// 预处理图像数据
	if err := s.preprocessImageEditData(req); err != nil {
		return nil, err
	}

	// 转换为多模态输入
	// 构建图像输入
	imageInput := &models.ImageInput{
		Data: req.ImageData,
		URL:  req.ImageURL,
	}

	inputs := []models.MultimodalInput{
		{
			Type:    models.InputTypeImage,
			Content: imageInput,
			Metadata: models.InputMetadata{
				MimeType: "image/jpeg", // 默认MIME类型
				Custom: map[string]string{
					"url": req.ImageURL,
				},
			},
		},
		{
			Type:    models.InputTypeText,
			Content: &models.TextInput{Content: req.Prompt},
		},
	}

	if req.MaskData != nil && len(req.MaskData) > 0 {
		maskInput := &models.ImageInput{
			Data: req.MaskData,
		}
		inputs = append(inputs, models.MultimodalInput{
			Type:    models.InputTypeImage,
			Content: maskInput,
			Metadata: models.InputMetadata{
				MimeType: "image/jpeg",
				Custom: map[string]string{
					"type": "mask",
				},
			},
		})
	}

	config := models.MultimodalConfig{
		Model:           "dall-e-2", // 默认图像编辑模型
		MaxTokens:       1000,
		Temperature:     0.7,
		ExpectedOutputs: []models.OutputType{models.OutputTypeImage},
	}

	// 使用生成处理器
	outputs, err := s.service.Generate(ctx, "default", inputs, config)
	if err != nil {
		return nil, s.wrapProviderError(err, "image_edit")
	}

	if len(outputs) == 0 || outputs[0].Type != models.OutputTypeImage {
		return nil, &MultimodalError{
			Code:    "NO_IMAGE_OUTPUT",
			Message: "no image output generated",
			Type:    ErrorTypeInternal,
			Retry:   true,
		}
	}

	// 从输出中提取图像内容
	imageOutput, ok := outputs[0].Content.(*models.ImageOutput)
	if !ok {
		return nil, &MultimodalError{
			Code:    "INVALID_OUTPUT_TYPE",
			Message: "expected image output from image editing",
			Type:    ErrorTypeInternal,
			Retry:   false,
		}
	}

	// 构建响应
	response := &providers.ImageEditResponse{
		Images: []providers.GeneratedImage{
			{
				URL:    imageOutput.URL,
				Width:  imageOutput.Dimensions.Width,
				Height: imageOutput.Dimensions.Height,
				Format: imageOutput.Format,
			},
		},
		Usage: providers.Usage{
			TotalTokens: 100, // 默认token使用量
		},
		RequestID: fmt.Sprintf("edit_%d", time.Now().Unix()),
		Model:     "dall-e-2",
	}

	// 记录日志
	s.logImageEdit(ctx, req.UserID, req.Prompt, response)

	return response, nil
}

// 辅助方法（保持向后兼容）

// withRetry 重试机制包装器
func (s *MultimodalService) withRetry(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		// 创建带超时的上下文
		_, cancel := context.WithTimeout(ctx, s.config.Timeout)
		
		// 执行操作
		err := operation()
		cancel()
		
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 检查是否应该重试
		if !s.shouldRetryError(err) || attempt == s.config.MaxRetries {
			break
		}
		
		// 计算重试延迟（指数退避）
		delay := s.calculateRetryDelay(attempt)
		
		// 等待重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}
	
	return lastErr
}

// shouldRetryError 判断错误是否应该重试
func (s *MultimodalService) shouldRetryError(err error) bool {
	if err == nil {
		return false
	}
	
	// 检查是否是多模态错误
	var multimodalErr *MultimodalError
	if errors.As(err, &multimodalErr) {
		return multimodalErr.Retry
	}
	
	// 检查常见的可重试错误
	errStr := strings.ToLower(err.Error())
	retryableErrors := []string{
		"timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
		"rate limit",
		"service unavailable",
		"internal server error",
		"bad gateway",
		"gateway timeout",
	}
	
	for _, retryableErr := range retryableErrors {
		if strings.Contains(errStr, retryableErr) {
			return true
		}
	}
	
	return false
}

// calculateRetryDelay 计算重试延迟（指数退避）
func (s *MultimodalService) calculateRetryDelay(attempt int) time.Duration {
	delay := time.Duration(math.Pow(2, float64(attempt))) * s.config.RetryDelay
	if delay > s.config.MaxRetryDelay {
		delay = s.config.MaxRetryDelay
	}
	return delay
}

// wrapProviderError 包装提供者错误
func (s *MultimodalService) wrapProviderError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// 如果已经是MultimodalError，直接返回
	var multimodalErr *MultimodalError
	if errors.As(err, &multimodalErr) {
		return err
	}

	// 根据错误类型进行分类
	errStr := strings.ToLower(err.Error())
	
	var errorType string
	var retry bool
	
	switch {
	case strings.Contains(errStr, "timeout"):
		errorType = ErrorTypeTimeout
		retry = true
	case strings.Contains(errStr, "network") || strings.Contains(errStr, "connection"):
		errorType = ErrorTypeNetwork
		retry = true
	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		errorType = ErrorTypeValidation
		retry = false
	default:
		errorType = ErrorTypeProvider
		retry = true
	}

	return &MultimodalError{
		Code:    "PROVIDER_ERROR",
		Message: fmt.Sprintf("%s operation failed: %v", operation, err),
		Type:    errorType,
		Retry:   retry,
	}
}

// validateInput 验证输入
func (s *MultimodalService) validateInput(input interface{}, inputType string) error {
	if input == nil {
		return &MultimodalError{
			Code:    "INVALID_INPUT",
			Message: fmt.Sprintf("%s input cannot be nil", inputType),
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}
	return nil
}

// validateRequest 验证请求
func (s *MultimodalService) validateRequest(req *models.MultimodalRequest) error {
	if req == nil {
		return &MultimodalError{
			Code:    "INVALID_REQUEST",
			Message: "request cannot be nil",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	if len(req.Inputs) == 0 {
		return &MultimodalError{
			Code:    "EMPTY_INPUTS",
			Message: "request inputs cannot be empty",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	return nil
}

// preprocessInputs 预处理输入
func (s *MultimodalService) preprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error) {
	return s.service.PreprocessInputs(ctx, inputs)
}

// postprocessOutputs 后处理输出
func (s *MultimodalService) postprocessOutputs(ctx context.Context, outputs []models.MultimodalOutput, expectedOutputs []string) ([]models.MultimodalOutput, error) {
	var expectedTypes []models.OutputType
	for _, output := range expectedOutputs {
		switch output {
		case "text":
			expectedTypes = append(expectedTypes, models.OutputTypeText)
		case "image":
			expectedTypes = append(expectedTypes, models.OutputTypeImage)
		case "audio":
			expectedTypes = append(expectedTypes, models.OutputTypeAudio)
		case "video":
			expectedTypes = append(expectedTypes, models.OutputTypeVideo)
		}
	}
	
	return s.service.PostprocessOutputs(ctx, outputs, expectedTypes)
}

// determineInputType 确定输入类型
func (s *MultimodalService) determineInputType(contentType string) models.InputType {
	switch {
	case strings.HasPrefix(contentType, "text/"):
		return models.InputTypeText
	case strings.HasPrefix(contentType, "image/"):
		return models.InputTypeImage
	case strings.HasPrefix(contentType, "audio/"):
		return models.InputTypeAudio
	case strings.HasPrefix(contentType, "video/"):
		return models.InputTypeVideo
	default:
		return models.InputTypeText
	}
}

// extractFormat 提取格式
func (s *MultimodalService) extractFormat(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

// detectLanguage 检测语言
func (s *MultimodalService) detectLanguage(text string) string {
	// 简单的语言检测逻辑
	if len(text) == 0 {
		return "unknown"
	}
	
	// 检测中文
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return "zh"
		}
	}
	
	return "en"
}

// calculateTokenUsage 计算token使用量
func calculateTokenUsage(inputs []models.MultimodalInput, outputs []models.MultimodalOutput) int {
	totalTokens := 0
	
	// 计算输入tokens
	for _, input := range inputs {
		switch input.Type {
		case models.InputTypeText:
			// 简单估算：每4个字符约等于1个token
			if textInput, ok := input.Content.(*models.TextInput); ok {
				totalTokens += len(textInput.Content) / 4
			} else {
				totalTokens += 50 // 默认值
			}
		case models.InputTypeImage:
			// 图像固定token数
			totalTokens += 85
		case models.InputTypeAudio:
			// 音频按时长计算
			if durationStr, ok := input.Metadata.Custom["duration"]; ok {
				if duration, err := strconv.ParseFloat(durationStr, 64); err == nil {
					totalTokens += int(duration * 10) // 每秒10个token
				} else {
					totalTokens += 100 // 默认值
				}
			} else {
				totalTokens += 100 // 默认值
			}
		case models.InputTypeVideo:
			// 视频按时长计算
			if durationStr, ok := input.Metadata.Custom["duration"]; ok {
				if duration, err := strconv.ParseFloat(durationStr, 64); err == nil {
					totalTokens += int(duration * 20) // 每秒20个token
				} else {
					totalTokens += 200 // 默认值
				}
			} else {
				totalTokens += 200 // 默认值
			}
		}
	}
	
	// 计算输出tokens（简单估算）
	for _, output := range outputs {
		switch output.Type {
		case models.OutputTypeText:
			if textOutput, ok := output.Content.(*models.TextOutput); ok {
				totalTokens += len(textOutput.Content) / 4
			} else {
				totalTokens += 50 // 默认值
			}
		case models.OutputTypeImage:
			totalTokens += 100 // 图像输出固定token数
		default:
			totalTokens += 50 // 其他类型默认值
		}
	}
	
	return totalTokens
}

// 服务接口定义（保持向后兼容）

type FileService interface {
	SaveFile(ctx context.Context, data []byte, filename string) (string, error)
	GetFile(ctx context.Context, url string) ([]byte, error)
}

type AudioService interface {
	GetDuration(data []byte) (float64, error)
	ProcessAudio(data []byte, format string) ([]byte, error)
}

type ImageService interface {
	GetDimensions(data []byte) (models.ImageDimensions, error)
	ProcessImage(data []byte, format string) ([]byte, error)
}

type VideoService interface {
	GetInfo(data []byte) (*VideoInfo, error)
	ProcessVideo(data []byte, format string) ([]byte, error)
}



// 验证方法（保持向后兼容）

func (s *MultimodalService) validateImageGenerateRequest(req *providers.ImageGenerateRequest) error {
	if req == nil {
		return &MultimodalError{
			Code:    "INVALID_REQUEST",
			Message: "image generate request cannot be nil",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	if req.Prompt == "" {
		return &MultimodalError{
			Code:    "EMPTY_PROMPT",
			Message: "prompt cannot be empty",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	// 注意：ImageGenerateRequest 结构体没有 Config 和 Model 字段
	// 这些配置将在调用时设置

	if req.Size == "" {
		req.Size = "1024x1024"
	}

	if req.Quality == "" {
		req.Quality = "standard"
	}

	return nil
}

func (s *MultimodalService) validateImageAnalyzeRequest(req *providers.ImageAnalyzeRequest) error {
	if req == nil {
		return &MultimodalError{
			Code:    "INVALID_REQUEST",
			Message: "image analyze request cannot be nil",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	if len(req.ImageData) == 0 && req.ImageURL == "" {
		return &MultimodalError{
			Code:    "MISSING_IMAGE",
			Message: "either image_data or image_url must be provided",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	// 注意：ImageAnalyzeRequest 结构体没有 Config、Model、MaxTokens 字段
	// 这些配置将在调用时设置

	return nil
}

func (s *MultimodalService) validateImageEditRequest(req *providers.ImageEditRequest) error {
	if req == nil {
		return &MultimodalError{
			Code:    "INVALID_REQUEST",
			Message: "image edit request cannot be nil",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	if len(req.ImageData) == 0 && req.ImageURL == "" {
		return &MultimodalError{
			Code:    "MISSING_IMAGE",
			Message: "either image_data or image_url must be provided",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	if req.Prompt == "" {
		return &MultimodalError{
			Code:    "EMPTY_PROMPT",
			Message: "prompt cannot be empty",
			Type:    ErrorTypeValidation,
			Retry:   false,
		}
	}

	// 注意：ImageEditRequest 结构体没有 Config 和 Model 字段
	// 这些配置将在调用时设置

	if req.Size == "" {
		req.Size = "1024x1024"
	}

	return nil
}

// 预处理方法（保持向后兼容）

func (s *MultimodalService) preprocessImageData(req *providers.ImageAnalyzeRequest) error {
	// 如果提供了URL但没有数据，尝试下载
	if req.ImageURL != "" && len(req.ImageData) == 0 && s.fileService != nil {
		data, err := s.fileService.GetFile(context.Background(), req.ImageURL)
		if err != nil {
			return &MultimodalError{
				Code:    "IMAGE_DOWNLOAD_ERROR",
				Message: fmt.Sprintf("failed to download image: %v", err),
				Type:    ErrorTypeNetwork,
				Retry:   true,
			}
		}
		req.ImageData = data
	}

	return nil
}

func (s *MultimodalService) preprocessImageEditData(req *providers.ImageEditRequest) error {
	// 如果提供了URL但没有数据，尝试下载
	if req.ImageURL != "" && len(req.ImageData) == 0 && s.fileService != nil {
		data, err := s.fileService.GetFile(context.Background(), req.ImageURL)
		if err != nil {
			return &MultimodalError{
				Code:    "IMAGE_DOWNLOAD_ERROR",
				Message: fmt.Sprintf("failed to download image: %v", err),
				Type:    ErrorTypeNetwork,
				Retry:   true,
			}
		}
		req.ImageData = data
	}

	return nil
}

// 日志方法（保持向后兼容）

func (s *MultimodalService) logImageGeneration(ctx context.Context, userID, prompt string, response *providers.ImageGenerateResponse) {
	// 实现日志记录逻辑
}

func (s *MultimodalService) logImageAnalysis(ctx context.Context, userID, prompt string, response *providers.ImageAnalyzeResponse) {
	// 实现日志记录逻辑
}

func (s *MultimodalService) logImageEdit(ctx context.Context, userID, prompt string, response *providers.ImageEditResponse) {
	// 实现日志记录逻辑
}

// 获取底层服务（用于高级用法）
func (s *MultimodalService) GetModularService() *multimodal.Service {
	return s.service
}
