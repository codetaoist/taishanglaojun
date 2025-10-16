package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// ProcessorManager 处理器管理器
type ProcessorManager struct {
	processors   map[ProcessorType]Processor
	validator    InputValidator
	preprocessor InputPreprocessor
	postprocessor OutputPostprocessor
	errorHandler ErrorHandler
}

// NewProcessorManager 创建处理器管理器
func NewProcessorManager() *ProcessorManager {
	// 创建默认的支持模块
	validator := NewDefaultInputValidator()
	preprocessor := NewDefaultInputPreprocessor()
	postprocessor := NewDefaultOutputPostprocessor()
	errorHandler := NewDefaultErrorHandler(3, DefaultConfig().RetryDelay, DefaultConfig().MaxRetryDelay)

	// 创建处理器管理器
	pm := &ProcessorManager{
		processors:    make(map[ProcessorType]Processor),
		validator:     validator,
		preprocessor:  preprocessor,
		postprocessor: postprocessor,
		errorHandler:  errorHandler,
	}

	// 注册默认处理器
	pm.registerDefaultProcessors()

	return pm
}

// NewProcessorManagerWithComponents 使用自定义组件创建处理器管理器
func NewProcessorManagerWithComponents(
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *ProcessorManager {
	pm := &ProcessorManager{
		processors:    make(map[ProcessorType]Processor),
		validator:     validator,
		preprocessor:  preprocessor,
		postprocessor: postprocessor,
		errorHandler:  errorHandler,
	}

	// 注册默认处理器
	pm.registerDefaultProcessors()

	return pm
}

// registerDefaultProcessors 注册默认处理器
func (pm *ProcessorManager) registerDefaultProcessors() {
	// 注册聊天处理器
	chatProcessor := NewChatProcessor(
		pm.validator,
		pm.preprocessor,
		pm.postprocessor,
		pm.errorHandler,
	)
	pm.RegisterProcessor(ProcessorTypeChat, chatProcessor)

	// 注册分析处理器
	analysisProcessor := NewAnalysisProcessor(
		pm.validator,
		pm.preprocessor,
		pm.postprocessor,
		pm.errorHandler,
	)
	pm.RegisterProcessor(ProcessorTypeAnalysis, analysisProcessor)

	// 注册生成处理器
	generationProcessor := NewGenerationProcessor(
		pm.validator,
		pm.preprocessor,
		pm.postprocessor,
		pm.errorHandler,
	)
	pm.RegisterProcessor(ProcessorTypeGeneration, generationProcessor)

	// 注册翻译处理器
	translationProcessor := NewTranslationProcessor(
		pm.validator,
		pm.preprocessor,
		pm.postprocessor,
		pm.errorHandler,
	)
	pm.RegisterProcessor(ProcessorTypeTranslation, translationProcessor)

	// 注册搜索处理器
	searchProcessor := NewSearchProcessor(
		pm.validator,
		pm.preprocessor,
		pm.postprocessor,
		pm.errorHandler,
	)
	pm.RegisterProcessor(ProcessorTypeSearch, searchProcessor)
}

// RegisterProcessor 注册处理器
func (pm *ProcessorManager) RegisterProcessor(processorType ProcessorType, processor Processor) {
	pm.processors[processorType] = processor
}

// UnregisterProcessor 注销处理器
func (pm *ProcessorManager) UnregisterProcessor(processorType ProcessorType) {
	delete(pm.processors, processorType)
}

// GetProcessor 获取处理器
func (pm *ProcessorManager) GetProcessor(processorType ProcessorType) (Processor, error) {
	processor, exists := pm.processors[processorType]
	if !exists {
		return nil, fmt.Errorf("processor type '%s' not found", processorType)
	}
	return processor, nil
}

// HasProcessor 检查是否有指定类型的处理器
func (pm *ProcessorManager) HasProcessor(processorType ProcessorType) bool {
	_, exists := pm.processors[processorType]
	return exists
}

// ListProcessors 列出所有已注册的处理器类型
func (pm *ProcessorManager) ListProcessors() []ProcessorType {
	var types []ProcessorType
	for processorType := range pm.processors {
		types = append(types, processorType)
	}
	return types
}

// Process 处理请求
func (pm *ProcessorManager) Process(
	ctx context.Context,
	processorType ProcessorType,
	provider providers.AIProvider,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	// 获取处理器
	processor, err := pm.GetProcessor(processorType)
	if err != nil {
		return nil, pm.errorHandler.HandleError(ctx, err, "get_processor")
	}

	// 处理请求
	outputs, err := processor.Process(ctx, provider, inputs, config)
	if err != nil {
		return nil, pm.errorHandler.HandleError(ctx, err, fmt.Sprintf("process_%s", processorType))
	}

	return outputs, nil
}

// ProcessWithFallback 使用备用处理器处理请求
func (pm *ProcessorManager) ProcessWithFallback(
	ctx context.Context,
	primaryType ProcessorType,
	fallbackType ProcessorType,
	provider providers.AIProvider,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	// 尝试使用主要处理器
	outputs, err := pm.Process(ctx, primaryType, provider, inputs, config)
	if err == nil {
		return outputs, nil
	}

	// 如果主要处理器失败，尝试备用处理器
	if pm.HasProcessor(fallbackType) {
		fallbackOutputs, fallbackErr := pm.Process(ctx, fallbackType, provider, inputs, config)
		if fallbackErr == nil {
			return fallbackOutputs, nil
		}

		// 聚合错误
		return nil, pm.errorHandler.AggregateErrors([]error{err, fallbackErr}, "process_with_fallback")
	}

	return nil, err
}

// ProcessBatch 批量处理请求
func (pm *ProcessorManager) ProcessBatch(
	ctx context.Context,
	requests []ProcessRequest,
) ([]ProcessResponse, error) {
	var responses []ProcessResponse

	for i, request := range requests {
		outputs, err := pm.Process(
			ctx,
			request.ProcessorType,
			request.Provider,
			request.Inputs,
			request.Config,
		)

		response := ProcessResponse{
			Index:   i,
			Outputs: outputs,
			Error:   err,
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// ProcessConcurrent 并发处理请求
func (pm *ProcessorManager) ProcessConcurrent(
	ctx context.Context,
	requests []ProcessRequest,
) ([]ProcessResponse, error) {
	responses := make([]ProcessResponse, len(requests))
	errChan := make(chan error, len(requests))
	
	// 启动并发处理
	for i, request := range requests {
		go func(index int, req ProcessRequest) {
			outputs, err := pm.Process(
				ctx,
				req.ProcessorType,
				req.Provider,
				req.Inputs,
				req.Config,
			)

			responses[index] = ProcessResponse{
				Index:   index,
				Outputs: outputs,
				Error:   err,
			}

			errChan <- err
		}(i, request)
	}

	// 等待所有处理完成
	var errors []error
	for i := 0; i < len(requests); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	// 如果有错误，聚合它们
	if len(errors) > 0 {
		aggregatedErr := pm.errorHandler.AggregateErrors(errors, "process_concurrent")
		return responses, aggregatedErr
	}

	return responses, nil
}

// ValidateInputs 验证输入
func (pm *ProcessorManager) ValidateInputs(inputs []models.MultimodalInput, config models.MultimodalConfig) error {
	return pm.validator.Validate(inputs, config)
}

// PreprocessInputs 预处理输入
func (pm *ProcessorManager) PreprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error) {
	return pm.preprocessor.PreprocessInputs(ctx, inputs)
}

// PostprocessOutputs 后处理输出
func (pm *ProcessorManager) PostprocessOutputs(
	ctx context.Context,
	outputs []models.MultimodalOutput,
	expectedTypes []models.OutputType,
) ([]models.MultimodalOutput, error) {
	return pm.postprocessor.PostprocessOutputs(ctx, outputs, expectedTypes)
}

// GetProcessorInfo 获取处理器信息
func (pm *ProcessorManager) GetProcessorInfo(processorType ProcessorType) (ProcessorInfo, error) {
	_, err := pm.GetProcessor(processorType)
	if err != nil {
		return ProcessorInfo{}, err
	}

	return ProcessorInfo{
		Type:        processorType, // 直接使用传入的 processorType 而不是 processor.GetType()
		Description: pm.getProcessorDescription(processorType),
		Capabilities: pm.getProcessorCapabilities(processorType),
	}, nil
}

// getProcessorDescription 获取处理器描述
func (pm *ProcessorManager) getProcessorDescription(processorType ProcessorType) string {
	switch processorType {
	case ProcessorTypeChat:
		return "Handles conversational interactions with multimodal inputs"
	case ProcessorTypeAnalysis:
		return "Analyzes and extracts insights from multimodal content"
	case ProcessorTypeGeneration:
		return "Generates new content based on multimodal inputs"
	case ProcessorTypeTranslation:
		return "Translates content across different languages and modalities"
	case ProcessorTypeSearch:
		return "Searches for information based on multimodal queries"
	default:
		return "Unknown processor type"
	}
}

// getProcessorCapabilities 获取处理器能力
func (pm *ProcessorManager) getProcessorCapabilities(processorType ProcessorType) []string {
	switch processorType {
	case ProcessorTypeChat:
		return []string{"text", "image", "audio", "video", "conversation"}
	case ProcessorTypeAnalysis:
		return []string{"text_analysis", "image_analysis", "audio_analysis", "video_analysis", "sentiment", "entities"}
	case ProcessorTypeGeneration:
		return []string{"text_generation", "image_generation", "audio_generation", "creative_writing"}
	case ProcessorTypeTranslation:
		return []string{"text_translation", "image_translation", "audio_translation", "video_translation", "multilingual"}
	case ProcessorTypeSearch:
		return []string{"text_search", "image_search", "audio_search", "video_search", "semantic_search"}
	default:
		return []string{}
	}
}

// ProcessRequest 处理请求结构
type ProcessRequest struct {
	ProcessorType ProcessorType
	Provider      providers.AIProvider
	Inputs        []models.MultimodalInput
	Config        models.MultimodalConfig
}

// ProcessResponse 处理响应结构
type ProcessResponse struct {
	Index   int
	Outputs []models.MultimodalOutput
	Error   error
}

// ProcessorInfo 处理器信息
type ProcessorInfo struct {
	Type         ProcessorType
	Description  string
	Capabilities []string
}