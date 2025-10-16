package multimodal

import (
	"context"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// Processor 处理器基础接口
type Processor interface {
	// Process 处理多模态请求
	Process(ctx context.Context, provider providers.AIProvider, inputs []models.MultimodalInput, config models.MultimodalConfig) ([]models.MultimodalOutput, error)
	
	// GetType 获取处理器类型
	GetType() string
	
	// Validate 验证输入是否适合此处理器
	Validate(inputs []models.MultimodalInput, config models.MultimodalConfig) error
}

// InputValidator 输入验证器接口
type InputValidator interface {
	// Validate 验证输入和配置
	Validate(inputs []models.MultimodalInput, config models.MultimodalConfig) error
	
	// ValidateRequest 验证请求
	ValidateRequest(req *models.MultimodalRequest) error
	
	// ValidateInput 验证单个输入
	ValidateInput(input interface{}, inputType string) error
}

// InputPreprocessor 输入预处理器接口
type InputPreprocessor interface {
	// PreprocessInputs 预处理输入列表
	PreprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error)
	
	// PreprocessTextInput 预处理文本输入
	PreprocessTextInput(input models.MultimodalInput) (models.MultimodalInput, error)
	
	// PreprocessImageInput 预处理图像输入
	PreprocessImageInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error)
	
	// PreprocessAudioInput 预处理音频输入
	PreprocessAudioInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error)
	
	// PreprocessVideoInput 预处理视频输入
	PreprocessVideoInput(ctx context.Context, input models.MultimodalInput) (models.MultimodalInput, error)
}

// OutputPostprocessor 输出后处理器接口
type OutputPostprocessor interface {
	// PostprocessOutputs 后处理输出列表
	PostprocessOutputs(ctx context.Context, outputs []models.MultimodalOutput, expectedOutputs []models.OutputType) ([]models.MultimodalOutput, error)
	
	// ConvertOutput 转换输出类型
	ConvertOutput(ctx context.Context, output models.MultimodalOutput, targetType models.MultimodalOutputType) (models.MultimodalOutput, error)
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	// HandleError 处理错误
	HandleError(ctx context.Context, err error, operation string) error
	
	// AggregateErrors 聚合多个错误
	AggregateErrors(errors []error, operation string) error
	
	// WrapProviderError 包装提供者错误
	WrapProviderError(err error, operation string) error
	
	// ShouldRetryError 判断是否应该重试错误
	ShouldRetryError(err error) bool
	
	// CalculateRetryDelay 计算重试延迟
	CalculateRetryDelay(attempt int) time.Duration
}

// ProcessorManagerInterface 处理器管理器接口
type ProcessorManagerInterface interface {
	// GetProcessor 根据请求类型获取处理器
	GetProcessor(requestType ProcessorType) (Processor, error)
	
	// RegisterProcessor 注册处理器
	RegisterProcessor(requestType ProcessorType, processor Processor)
	
	// ListProcessors 列出所有处理器
	ListProcessors() []ProcessorType
}