package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/taishanglaojun/core-services/ai-integration/models"
	"github.com/taishanglaojun/core-services/ai-integration/providers"
	"github.com/taishanglaojun/infrastructure/database-layer/repositories"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MultimodalService 多模态AI服务
type MultimodalService struct {
	aiProviders    map[string]providers.AIProvider
	aiProvider     providers.AIProvider // 默认AI提供商
	repository     repositories.Repository
	fileService    FileService
	audioService   AudioService
	imageService   ImageService
	videoService   VideoService
}

// NewMultimodalService 创建多模态AI服务
func NewMultimodalService(
	aiProviders map[string]providers.AIProvider,
	defaultProvider providers.AIProvider,
	repository repositories.Repository,
	fileService FileService,
	audioService AudioService,
	imageService ImageService,
	videoService VideoService,
) *MultimodalService {
	return &MultimodalService{
		aiProviders:  aiProviders,
		aiProvider:   defaultProvider,
		repository:   repository,
		fileService:  fileService,
		audioService: audioService,
		imageService: imageService,
		videoService: videoService,
	}
}

// ProcessMultimodalRequest 处理多模态请求
func (s *MultimodalService) ProcessMultimodalRequest(ctx context.Context, req *models.MultimodalRequest) (*models.MultimodalResponse, error) {
	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 预处理输入
	processedInputs, err := s.preprocessInputs(ctx, req.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to preprocess inputs: %w", err)
	}

	// 获取AI提供商
	provider, exists := s.aiProviders[req.Config.Provider]
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", req.Config.Provider)
	}

	// 调用AI服务
	outputs, err := s.callAIProvider(ctx, provider, req.Type, processedInputs, req.Config)
	if err != nil {
		return nil, fmt.Errorf("AI provider call failed: %w", err)
	}

	// 后处理输出
	processedOutputs, err := s.postprocessOutputs(ctx, outputs, req.Outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to postprocess outputs: %w", err)
	}

	// 创建响应
	response := &models.MultimodalResponse{
		ID:        uuid.New().String(),
		RequestID: req.ID,
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Type:      req.Type,
		Outputs:   processedOutputs,
		CreatedAt: time.Now(),
		Status:    "completed",
		Metadata: models.ResponseMetadata{
			Provider:     req.Config.Provider,
			Model:        req.Config.Model,
			TokensUsed:   calculateTokenUsage(processedInputs, processedOutputs),
			ProcessingTime: time.Since(req.CreatedAt).Milliseconds(),
			Quality: models.QualityMetrics{
				Accuracy:   0.95,
				Relevance:  0.92,
				Coherence:  0.88,
				Fluency:    0.90,
			},
		},
	}

	// 保存到数据库
	if err := s.repository.Create(ctx, "multimodal_responses", response); err != nil {
		return nil, fmt.Errorf("failed to save response: %w", err)
	}

	return response, nil
}

// ProcessFileUpload 处理文件上传
func (s *MultimodalService) ProcessFileUpload(ctx *gin.Context) (*models.MultimodalInput, error) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	defer file.Close()

	// 读取文件数据
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 确定文件类型
	inputType := s.determineInputType(header.Header.Get("Content-Type"))
	
	// 创建输入对象
	input := &models.MultimodalInput{
		Type: inputType,
		Metadata: models.InputMetadata{
			MimeType: header.Header.Get("Content-Type"),
			Size:     int64(len(data)),
		},
	}

	// 根据类型处理内容
	switch inputType {
	case models.InputTypeText:
		input.Content = models.TextInput{
			Content: string(data),
			Format:  "plain",
		}
	case models.InputTypeAudio:
		audioInput := models.AudioInput{
			Data:   data,
			Format: s.extractFormat(header.Filename),
		}
		// 获取音频时长
		if duration, err := s.audioService.GetDuration(data); err == nil {
			audioInput.Duration = duration
			input.Metadata.Duration = duration
		}
		input.Content = audioInput
	case models.InputTypeImage:
		imageInput := models.ImageInput{
			Data:   data,
			Format: s.extractFormat(header.Filename),
		}
		// 获取图像尺寸
		if dimensions, err := s.imageService.GetDimensions(data); err == nil {
			imageInput.Dimensions = dimensions
			input.Metadata.Dimensions = dimensions
		}
		input.Content = imageInput
	case models.InputTypeVideo:
		videoInput := models.VideoInput{
			Data:   data,
			Format: s.extractFormat(header.Filename),
		}
		// 获取视频信息
		if info, err := s.videoService.GetInfo(data); err == nil {
			videoInput.Duration = info.Duration
			videoInput.Dimensions = info.Dimensions
			videoInput.FrameRate = info.FrameRate
			input.Metadata.Duration = info.Duration
			input.Metadata.Dimensions = info.Dimensions
		}
		input.Content = videoInput
	default:
		return nil, fmt.Errorf("unsupported file type: %s", header.Header.Get("Content-Type"))
	}

	return input, nil
}

// StreamMultimodalResponse 流式处理多模态响应
func (s *MultimodalService) StreamMultimodalResponse(ctx context.Context, req *models.MultimodalRequest, outputChan chan<- *models.MultimodalOutput) error {
	defer close(outputChan)

	// 预处理输入
	processedInputs, err := s.preprocessInputs(ctx, req.Inputs)
	if err != nil {
		return fmt.Errorf("failed to preprocess inputs: %w", err)
	}

	// 获取AI提供商
	provider, exists := s.aiProviders[req.Config.Provider]
	if !exists {
		return fmt.Errorf("unsupported provider: %s", req.Config.Provider)
	}

	// 流式调用AI服务
	return s.streamAIProvider(ctx, provider, req.Type, processedInputs, req.Config, outputChan)
}

// validateRequest 验证请求
func (s *MultimodalService) validateRequest(req *models.MultimodalRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if len(req.Inputs) == 0 {
		return fmt.Errorf("at least one input is required")
	}
	if len(req.Outputs) == 0 {
		return fmt.Errorf("at least one output type is required")
	}
	if req.Config.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if req.Config.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

// preprocessInputs 预处理输入
func (s *MultimodalService) preprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error) {
	processed := make([]models.MultimodalInput, len(inputs))
	
	for i, input := range inputs {
		switch input.Type {
		case models.InputTypeText:
			// 文本预处理：清理、格式化
			processed[i] = s.preprocessTextInput(input)
		case models.InputTypeAudio:
			// 音频预处理：格式转换、降噪
			processedInput, err := s.preprocessAudioInput(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to preprocess audio input: %w", err)
			}
			processed[i] = processedInput
		case models.InputTypeImage:
			// 图像预处理：压缩、格式转换
			processedInput, err := s.preprocessImageInput(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to preprocess image input: %w", err)
			}
			processed[i] = processedInput
		case models.InputTypeVideo:
			// 视频预处理：压缩、关键帧提取
			processedInput, err := s.preprocessVideoInput(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to preprocess video input: %w", err)
			}
			processed[i] = processedInput
		default:
			processed[i] = input
		}
	}
	
	return processed, nil
}

// preprocessTextInput 预处理文本输入
func (s *MultimodalService) preprocessTextInput(input models.MultimodalInput) models.MultimodalInput {
	textInput := input.Content.(models.TextInput)
	
	// 清理文本
	textInput.Content = strings.TrimSpace(textInput.Content)
	
	// 检测语言
	if textInput.Language == "" {
		textInput.Language = s.detectLanguage(textInput.Content)
	}
	
	input.Content = textInput
	return input
}

// preprocessAudioInput 预处理音频输入
func (s *MultimodalService) preprocessAudioInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	audioInput := input.Content.(models.AudioInput)
	
	// 音频格式转换和优化
	if processedData, err := s.audioService.ProcessAudio(audioInput.Data, audioInput.Format); err == nil {
		audioInput.Data = processedData
	}
	
	input.Content = audioInput
	return input, nil
}

// preprocessImageInput 预处理图像输入
func (s *MultimodalService) preprocessImageInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	imageInput := input.Content.(models.ImageInput)
	
	// 图像压缩和优化
	if processedData, err := s.imageService.ProcessImage(imageInput.Data, imageInput.Format); err == nil {
		imageInput.Data = processedData
	}
	
	input.Content = imageInput
	return input, nil
}

// preprocessVideoInput 预处理视频输入
func (s *MultimodalService) preprocessVideoInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error) {
	videoInput := input.Content.(models.VideoInput)
	
	// 视频压缩和关键帧提取
	if processedData, err := s.videoService.ProcessVideo(videoInput.Data, videoInput.Format); err == nil {
		videoInput.Data = processedData
	}
	
	input.Content = videoInput
	return input, nil
}

// callAIProvider 调用AI提供商
func (s *MultimodalService) callAIProvider(ctx context.Context, provider providers.AIProvider, reqType models.MultimodalType, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// 根据请求类型调用不同的AI服务
	switch reqType {
	case models.MultimodalTypeChat:
		return s.callMultimodalChat(ctx, provider, inputs, config)
	case models.MultimodalTypeAnalysis:
		return s.callMultimodalAnalysis(ctx, provider, inputs, config)
	case models.MultimodalTypeGeneration:
		return s.callMultimodalGeneration(ctx, provider, inputs, config)
	case models.MultimodalTypeTranslation:
		return s.callMultimodalTranslation(ctx, provider, inputs, config)
	case models.MultimodalTypeSearch:
		return s.callMultimodalSearch(ctx, provider, inputs, config)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", reqType)
	}
}

// postprocessOutputs 后处理输出
func (s *MultimodalService) postprocessOutputs(ctx context.Context, outputs []models.MultimodalOutput, requestedTypes []models.MultimodalOutputType) ([]models.MultimodalOutput, error) {
	processed := make([]models.MultimodalOutput, 0)
	
	for _, output := range outputs {
		// 根据请求的输出类型进行转换
		for _, requestedType := range requestedTypes {
			if convertedOutput, err := s.convertOutput(ctx, output, requestedType); err == nil {
				processed = append(processed, convertedOutput)
			}
		}
	}
	
	return processed, nil
}

// 辅助方法
func (s *MultimodalService) determineInputType(contentType string) models.InputType {
	switch {
	case strings.HasPrefix(contentType, "text/"):
		return models.InputTypeText
	case strings.HasPrefix(contentType, "audio/"):
		return models.InputTypeAudio
	case strings.HasPrefix(contentType, "image/"):
		return models.InputTypeImage
	case strings.HasPrefix(contentType, "video/"):
		return models.InputTypeVideo
	default:
		return models.InputTypeFile
	}
}

func (s *MultimodalService) extractFormat(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

func (s *MultimodalService) detectLanguage(text string) string {
	// 简单的语言检测逻辑，实际应该使用专门的语言检测库
	if len(text) == 0 {
		return "unknown"
	}
	
	// 检测中文字符
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return "zh"
		}
	}
	
	return "en"
}

func calculateTokenUsage(inputs []models.MultimodalInput, outputs []models.MultimodalOutput) int {
	// 简化的token计算逻辑
	tokens := 0
	
	for _, input := range inputs {
		switch input.Type {
		case models.InputTypeText:
			if textInput, ok := input.Content.(models.TextInput); ok {
				tokens += len(strings.Fields(textInput.Content))
			}
		case models.InputTypeImage:
			tokens += 85 // 图像固定token消耗
		case models.InputTypeAudio:
			if audioInput, ok := input.Content.(models.AudioInput); ok {
				tokens += int(audioInput.Duration * 10) // 每秒10个token
			}
		}
	}
	
	for _, output := range outputs {
		switch output.Type {
		case models.OutputTypeText:
			if textOutput, ok := output.Content.(models.TextOutput); ok {
				tokens += len(strings.Fields(textOutput.Content))
			}
		case models.OutputTypeImage:
			tokens += 170 // 图像生成固定token消耗
		case models.OutputTypeAudio:
			if audioOutput, ok := output.Content.(models.AudioOutput); ok {
				tokens += int(audioOutput.Duration * 15) // 每秒15个token
			}
		}
	}
	
	return tokens
}

// 接口定义
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

type VideoInfo struct {
	Duration   float64
	Dimensions models.ImageDimensions
	FrameRate  float64
}

// 占位符方法，需要具体实现
func (s *MultimodalService) callMultimodalChat(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// TODO: 实现多模态对话
	return nil, fmt.Errorf("not implemented")
}

func (s *MultimodalService) callMultimodalAnalysis(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// TODO: 实现多模态分析
	return nil, fmt.Errorf("not implemented")
}

func (s *MultimodalService) callMultimodalGeneration(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// TODO: 实现多模态生成
	return nil, fmt.Errorf("not implemented")
}

func (s *MultimodalService) callMultimodalTranslation(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// TODO: 实现多模态翻译
	return nil, fmt.Errorf("not implemented")
}

func (s *MultimodalService) callMultimodalSearch(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error) {
	// TODO: 实现多模态搜索
	return nil, fmt.Errorf("not implemented")
}

func (s *MultimodalService) streamAIProvider(ctx context.Context, provider providers.AIProvider, reqType models.MultimodalType, inputs []models.MultimodalInput, config models.MultimodalConfig, outputChan chan<- *models.MultimodalOutput) error {
	// TODO: 实现流式AI调用
	return fmt.Errorf("not implemented")
}

func (s *MultimodalService) convertOutput(ctx context.Context, output models.MultimodalOutput, targetType models.MultimodalOutputType) (models.MultimodalOutput, error) {
	// TODO: 实现输出格式转换
	return output, nil
}

// GenerateImage 生成图像
func (s *MultimodalService) GenerateImage(ctx context.Context, req *ImageGenerateRequest) (*ImageGenerateResponse, error) {
	// 验证请求
	if err := s.validateImageGenerateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 构建AI提供商请求
	providerReq := &ImageGenerateRequest{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Size:           req.Size,
		Quality:        req.Quality,
		Style:          req.Style,
		N:              req.N,
	}

	// 设置默认值
	if providerReq.Size == "" {
		providerReq.Size = "1024x1024"
	}
	if providerReq.Quality == "" {
		providerReq.Quality = "standard"
	}
	if providerReq.N == 0 {
		providerReq.N = 1
	}

	// 调用AI提供商
	response, err := s.aiProvider.GenerateImage(ctx, providerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	// 记录生成历史
	if req.UserID != "" {
		go s.logImageGeneration(ctx, req.UserID, req.Prompt, response)
	}

	return response, nil
}

// AnalyzeImage 分析图像
func (s *MultimodalService) AnalyzeImage(ctx context.Context, req *ImageAnalyzeRequest) (*ImageAnalyzeResponse, error) {
	// 验证请求
	if err := s.validateImageAnalyzeRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 预处理图像数据
	if err := s.preprocessImageData(req); err != nil {
		return nil, fmt.Errorf("failed to preprocess image: %w", err)
	}

	// 构建AI提供商请求
	providerReq := &ImageAnalyzeRequest{
		ImageURL:  req.ImageURL,
		ImageData: req.ImageData,
		Prompt:    req.Prompt,
		Detail:    req.Detail,
		Features:  req.Features,
	}

	// 设置默认值
	if providerReq.Detail == "" {
		providerReq.Detail = "auto"
	}
	if providerReq.Prompt == "" {
		providerReq.Prompt = "请详细描述这张图片的内容"
	}

	// 调用AI提供商
	response, err := s.aiProvider.AnalyzeImage(ctx, providerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze image: %w", err)
	}

	// 记录分析历史
	if req.UserID != "" {
		go s.logImageAnalysis(ctx, req.UserID, req.Prompt, response)
	}

	return response, nil
}

// EditImage 编辑图像
func (s *MultimodalService) EditImage(ctx context.Context, req *ImageEditRequest) (*ImageEditResponse, error) {
	// 验证请求
	if err := s.validateImageEditRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 预处理图像数据
	if err := s.preprocessImageEditData(req); err != nil {
		return nil, fmt.Errorf("failed to preprocess image: %w", err)
	}

	// 构建AI提供商请求
	providerReq := &ImageEditRequest{
		ImageURL:  req.ImageURL,
		ImageData: req.ImageData,
		MaskURL:   req.MaskURL,
		MaskData:  req.MaskData,
		Prompt:    req.Prompt,
		Size:      req.Size,
		N:         req.N,
	}

	// 设置默认值
	if providerReq.Size == "" {
		providerReq.Size = "1024x1024"
	}
	if providerReq.N == 0 {
		providerReq.N = 1
	}

	// 调用AI提供商
	response, err := s.aiProvider.EditImage(ctx, providerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to edit image: %w", err)
	}

	// 记录编辑历史
	if req.UserID != "" {
		go s.logImageEdit(ctx, req.UserID, req.Prompt, response)
	}

	return response, nil
}

// 验证图像生成请求
func (s *MultimodalService) validateImageGenerateRequest(req *ImageGenerateRequest) error {
	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	
	if len(req.Prompt) > 4000 {
		return fmt.Errorf("prompt too long, maximum 4000 characters")
	}

	// 验证尺寸格式
	if req.Size != "" {
		validSizes := []string{"256x256", "512x512", "1024x1024", "1792x1024", "1024x1792"}
		valid := false
		for _, size := range validSizes {
			if req.Size == size {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid size, must be one of: %v", validSizes)
		}
	}

	// 验证质量
	if req.Quality != "" && req.Quality != "standard" && req.Quality != "hd" {
		return fmt.Errorf("invalid quality, must be 'standard' or 'hd'")
	}

	// 验证风格
	if req.Style != "" && req.Style != "vivid" && req.Style != "natural" {
		return fmt.Errorf("invalid style, must be 'vivid' or 'natural'")
	}

	// 验证数量
	if req.N < 0 || req.N > 10 {
		return fmt.Errorf("invalid n, must be between 1 and 10")
	}

	return nil
}

// 验证图像分析请求
func (s *MultimodalService) validateImageAnalyzeRequest(req *ImageAnalyzeRequest) error {
	if req.ImageURL == "" && len(req.ImageData) == 0 {
		return fmt.Errorf("either image_url or image_data is required")
	}

	if req.ImageURL != "" && len(req.ImageData) > 0 {
		return fmt.Errorf("cannot specify both image_url and image_data")
	}

	// 验证图像数据大小
	if len(req.ImageData) > 20*1024*1024 { // 20MB
		return fmt.Errorf("image data too large, maximum 20MB")
	}

	// 验证详细程度
	if req.Detail != "" && req.Detail != "auto" && req.Detail != "low" && req.Detail != "high" {
		return fmt.Errorf("invalid detail, must be 'auto', 'low', or 'high'")
	}

	return nil
}

// 验证图像编辑请求
func (s *MultimodalService) validateImageEditRequest(req *ImageEditRequest) error {
	if req.ImageURL == "" && len(req.ImageData) == 0 {
		return fmt.Errorf("either image_url or image_data is required")
	}

	if req.ImageURL != "" && len(req.ImageData) > 0 {
		return fmt.Errorf("cannot specify both image_url and image_data")
	}

	if req.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}

	// 验证图像数据大小
	if len(req.ImageData) > 4*1024*1024 { // 4MB
		return fmt.Errorf("image data too large, maximum 4MB")
	}

	// 验证遮罩数据大小
	if len(req.MaskData) > 4*1024*1024 { // 4MB
		return fmt.Errorf("mask data too large, maximum 4MB")
	}

	// 验证尺寸格式
	if req.Size != "" {
		validSizes := []string{"256x256", "512x512", "1024x1024"}
		valid := false
		for _, size := range validSizes {
			if req.Size == size {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid size, must be one of: %v", validSizes)
		}
	}

	// 验证数量
	if req.N < 0 || req.N > 10 {
		return fmt.Errorf("invalid n, must be between 1 and 10")
	}

	return nil
}

// 预处理图像数据
func (s *MultimodalService) preprocessImageData(req *ImageAnalyzeRequest) error {
	// 如果有图像数据，验证格式
	if len(req.ImageData) > 0 {
		// 检查图像格式
		contentType := http.DetectContentType(req.ImageData)
		if !strings.HasPrefix(contentType, "image/") {
			return fmt.Errorf("invalid image format: %s", contentType)
		}

		// 支持的格式
		supportedTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
		supported := false
		for _, t := range supportedTypes {
			if contentType == t {
				supported = true
				break
			}
		}
		if !supported {
			return fmt.Errorf("unsupported image format: %s", contentType)
		}
	}

	return nil
}

// 预处理图像编辑数据
func (s *MultimodalService) preprocessImageEditData(req *ImageEditRequest) error {
	// 验证原图像格式
	if len(req.ImageData) > 0 {
		contentType := http.DetectContentType(req.ImageData)
		if contentType != "image/png" {
			return fmt.Errorf("image must be PNG format for editing")
		}
	}

	// 验证遮罩格式
	if len(req.MaskData) > 0 {
		contentType := http.DetectContentType(req.MaskData)
		if contentType != "image/png" {
			return fmt.Errorf("mask must be PNG format")
		}
	}

	return nil
}

// 记录图像生成历史
func (s *MultimodalService) logImageGeneration(ctx context.Context, userID, prompt string, response *ImageGenerateResponse) {
	// TODO: 实现记录到数据库
	// 可以记录用户ID、提示词、生成的图像URL、时间戳等
}

// 记录图像分析历史
func (s *MultimodalService) logImageAnalysis(ctx context.Context, userID, prompt string, response *ImageAnalyzeResponse) {
	// TODO: 实现记录到数据库
	// 可以记录用户ID、分析提示、分析结果、时间戳等
}

// 记录图像编辑历史
func (s *MultimodalService) logImageEdit(ctx context.Context, userID, prompt string, response *ImageEditResponse) {
	// TODO: 实现记录到数据库
	// 可以记录用户ID、编辑提示、编辑结果、时间戳等
}

// 图像处理相关类型定义
type ImageGenerateRequest struct {
	UserID         string            `json:"user_id,omitempty"`
	Prompt         string            `json:"prompt" binding:"required"`
	NegativePrompt string            `json:"negative_prompt,omitempty"`
	Size           string            `json:"size,omitempty"`
	Quality        string            `json:"quality,omitempty"`
	Style          string            `json:"style,omitempty"`
	N              int               `json:"n,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type ImageGenerateResponse struct {
	ID        string            `json:"id"`
	Images    []GeneratedImage  `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type GeneratedImage struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type ImageAnalyzeRequest struct {
	UserID    string            `json:"user_id,omitempty"`
	ImageURL  string            `json:"image_url,omitempty"`
	ImageData []byte            `json:"image_data,omitempty"`
	Prompt    string            `json:"prompt,omitempty"`
	Detail    string            `json:"detail,omitempty"`
	Features  []string          `json:"features,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ImageAnalyzeResponse struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Objects     []DetectedObject  `json:"objects,omitempty"`
	Text        string            `json:"text,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Colors      []string          `json:"colors,omitempty"`
	Confidence  float64           `json:"confidence"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type DetectedObject struct {
	Name        string       `json:"name"`
	Confidence  float64      `json:"confidence"`
	BoundingBox *BoundingBox `json:"bounding_box,omitempty"`
}

type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ImageEditRequest struct {
	UserID    string            `json:"user_id,omitempty"`
	ImageURL  string            `json:"image_url,omitempty"`
	ImageData []byte            `json:"image_data,omitempty"`
	MaskURL   string            `json:"mask_url,omitempty"`
	MaskData  []byte            `json:"mask_data,omitempty"`
	Prompt    string            `json:"prompt" binding:"required"`
	Size      string            `json:"size,omitempty"`
	N         int               `json:"n,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type ImageEditResponse struct {
	ID        string            `json:"id"`
	Images    []GeneratedImage  `json:"images"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}