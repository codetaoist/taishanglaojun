package multimodal

import (
	"time"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// ServiceFactory 服务工厂
type ServiceFactory struct {
	defaultConfig Config
}

// NewServiceFactory 创建服务工厂
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{
		defaultConfig: *DefaultConfig(),
	}
}

// NewServiceFactoryWithConfig 使用自定义配置创建服务工厂
func NewServiceFactoryWithConfig(config Config) *ServiceFactory {
	return &ServiceFactory{
		defaultConfig: config,
	}
}

// CreateService 创建标准服务
func (f *ServiceFactory) CreateService() *Service {
	return f.CreateServiceWithConfig(f.defaultConfig)
}

// CreateServiceWithConfig 使用指定配置创建服务
func (f *ServiceFactory) CreateServiceWithConfig(config Config) *Service {
	// 创建默认组件
	validator := NewDefaultInputValidator()
	preprocessor := NewDefaultInputPreprocessor()
	postprocessor := NewDefaultOutputPostprocessor()
	errorHandler := NewDefaultErrorHandler(3, time.Second, 30*time.Second)

	// 创建处理器管理器
	manager := NewProcessorManagerWithComponents(
		validator,
		preprocessor,
		postprocessor,
		errorHandler,
	)

	// 注册默认处理器
	f.registerDefaultProcessors(manager, validator, preprocessor, postprocessor, errorHandler)

	// 创建服务
	return NewServiceWithManager(config, manager)
}

// CreateServiceWithProviders 创建服务并注册提供者
func (f *ServiceFactory) CreateServiceWithProviders(providers map[string]providers.AIProvider) *Service {
	service := f.CreateService()
	
	for name, provider := range providers {
		service.RegisterProvider(name, provider)
	}
	
	return service
}

// CreateServiceWithCustomComponents 使用自定义组件创建服务
func (f *ServiceFactory) CreateServiceWithCustomComponents(
	config Config,
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) *Service {
	// 创建处理器管理器
	manager := NewProcessorManagerWithComponents(
		validator,
		preprocessor,
		postprocessor,
		errorHandler,
	)

	// 注册默认处理器
	f.registerDefaultProcessors(manager, validator, preprocessor, postprocessor, errorHandler)

	// 创建服务
	return NewServiceWithManager(config, manager)
}

// CreateMinimalService 创建最小化服务（仅包含基本处理器）
func (f *ServiceFactory) CreateMinimalService() *Service {
	config := Config{
		MaxRetries:    1,
		RetryDelay:    1000,
		MaxRetryDelay: 5000,
		Timeout:       30000,
	}

	// 创建默认组件
	validator := NewDefaultInputValidator()
	preprocessor := NewDefaultInputPreprocessor()
	postprocessor := NewDefaultOutputPostprocessor()
	errorHandler := NewDefaultErrorHandler(3, time.Second, 30*time.Second)

	// 创建处理器管理器
	manager := NewProcessorManagerWithComponents(
		validator,
		preprocessor,
		postprocessor,
		errorHandler,
	)

	// 只注册聊天处理器
	chatProcessor := NewChatProcessor(validator, preprocessor, postprocessor, errorHandler)
	manager.RegisterProcessor(ProcessorTypeChat, chatProcessor)

	// 创建服务
	return NewServiceWithManager(config, manager)
}

// CreateFullService 创建完整服务（包含所有处理器和组件）
func (f *ServiceFactory) CreateFullService() *Service {
	return f.CreateServiceWithConfig(f.defaultConfig)
}

// registerDefaultProcessors 注册默认处理器
func (f *ServiceFactory) registerDefaultProcessors(
	manager *ProcessorManager,
	validator InputValidator,
	preprocessor InputPreprocessor,
	postprocessor OutputPostprocessor,
	errorHandler ErrorHandler,
) {
	// 创建并注册聊天处理器
	chatProcessor := NewChatProcessor(validator, preprocessor, postprocessor, errorHandler)
	manager.RegisterProcessor(ProcessorTypeChat, chatProcessor)

	// 创建并注册分析处理器
	analysisProcessor := NewAnalysisProcessor(validator, preprocessor, postprocessor, errorHandler)
	manager.RegisterProcessor(ProcessorTypeAnalysis, analysisProcessor)

	// 创建并注册生成处理器
	generationProcessor := NewGenerationProcessor(validator, preprocessor, postprocessor, errorHandler)
	manager.RegisterProcessor(ProcessorTypeGeneration, generationProcessor)

	// 创建并注册翻译处理器
	translationProcessor := NewTranslationProcessor(validator, preprocessor, postprocessor, errorHandler)
	manager.RegisterProcessor(ProcessorTypeTranslation, translationProcessor)

	// 创建并注册搜索处理器
	searchProcessor := NewSearchProcessor(validator, preprocessor, postprocessor, errorHandler)
	manager.RegisterProcessor(ProcessorTypeSearch, searchProcessor)
}

// SetDefaultConfig 设置默认配置
func (f *ServiceFactory) SetDefaultConfig(config Config) {
	f.defaultConfig = config
}

// GetDefaultConfig 获取默认配置
func (f *ServiceFactory) GetDefaultConfig() Config {
	return f.defaultConfig
}

// ServiceBuilder 服务构建器
type ServiceBuilder struct {
	config        Config
	validator     InputValidator
	preprocessor  InputPreprocessor
	postprocessor OutputPostprocessor
	errorHandler  ErrorHandler
	providers     map[string]providers.AIProvider
	processors    map[ProcessorType]Processor
}

// NewServiceBuilder 创建服务构建器
func NewServiceBuilder() *ServiceBuilder {
	return &ServiceBuilder{
		config:     *DefaultConfig(),
		providers:  make(map[string]providers.AIProvider),
		processors: make(map[ProcessorType]Processor),
	}
}

// WithConfig 设置配置
func (b *ServiceBuilder) WithConfig(config Config) *ServiceBuilder {
	b.config = config
	return b
}

// WithValidator 设置验证器
func (b *ServiceBuilder) WithValidator(validator InputValidator) *ServiceBuilder {
	b.validator = validator
	return b
}

// WithPreprocessor 设置预处理器
func (b *ServiceBuilder) WithPreprocessor(preprocessor InputPreprocessor) *ServiceBuilder {
	b.preprocessor = preprocessor
	return b
}

// WithPostprocessor 设置后处理器
func (b *ServiceBuilder) WithPostprocessor(postprocessor OutputPostprocessor) *ServiceBuilder {
	b.postprocessor = postprocessor
	return b
}

// WithErrorHandler 设置错误处理器
func (b *ServiceBuilder) WithErrorHandler(errorHandler ErrorHandler) *ServiceBuilder {
	b.errorHandler = errorHandler
	return b
}

// WithProvider 添加提供者
func (b *ServiceBuilder) WithProvider(name string, provider providers.AIProvider) *ServiceBuilder {
	b.providers[name] = provider
	return b
}

// WithProcessor 添加自定义处理器
func (b *ServiceBuilder) WithProcessor(processorType ProcessorType, processor Processor) *ServiceBuilder {
	b.processors[processorType] = processor
	return b
}

// Build 构建服务
func (b *ServiceBuilder) Build() *Service {
	// 设置默认组件
	if b.validator == nil {
		b.validator = NewDefaultInputValidator()
	}
	if b.preprocessor == nil {
		b.preprocessor = NewDefaultInputPreprocessor()
	}
	if b.postprocessor == nil {
		b.postprocessor = NewDefaultOutputPostprocessor()
	}
	if b.errorHandler == nil {
		b.errorHandler = NewDefaultErrorHandler(3, time.Second, 30*time.Second)
	}

	// 创建处理器管理器
	manager := NewProcessorManagerWithComponents(
		b.validator,
		b.preprocessor,
		b.postprocessor,
		b.errorHandler,
	)

	// 注册自定义处理器
	for processorType, processor := range b.processors {
		manager.RegisterProcessor(processorType, processor)
	}

	// 如果没有自定义处理器，注册默认处理器
	if len(b.processors) == 0 {
		factory := NewServiceFactory()
		factory.registerDefaultProcessors(manager, b.validator, b.preprocessor, b.postprocessor, b.errorHandler)
	}

	// 创建服务
	service := NewServiceWithManager(b.config, manager)

	// 注册提供者
	for name, provider := range b.providers {
		service.RegisterProvider(name, provider)
	}

	return service
}

// 预定义的服务配置

// CreateChatOnlyService 创建仅支持聊天的服务
func CreateChatOnlyService() *Service {
	return NewServiceBuilder().
		WithProcessor(ProcessorTypeChat, NewChatProcessor(
			NewDefaultInputValidator(),
			NewDefaultInputPreprocessor(),
			NewDefaultOutputPostprocessor(),
			NewDefaultErrorHandler(3, time.Second, 30*time.Second),
		)).
		Build()
}

// CreateAnalysisOnlyService 创建仅支持分析的服务
func CreateAnalysisOnlyService() *Service {
	return NewServiceBuilder().
		WithProcessor(ProcessorTypeAnalysis, NewAnalysisProcessor(
			NewDefaultInputValidator(),
			NewDefaultInputPreprocessor(),
			NewDefaultOutputPostprocessor(),
			NewDefaultErrorHandler(3, time.Second, 30*time.Second),
		)).
		Build()
}

// CreateGenerationOnlyService 创建仅支持生成的服务
func CreateGenerationOnlyService() *Service {
	return NewServiceBuilder().
		WithProcessor(ProcessorTypeGeneration, NewGenerationProcessor(
			NewDefaultInputValidator(),
			NewDefaultInputPreprocessor(),
			NewDefaultOutputPostprocessor(),
			NewDefaultErrorHandler(3, time.Second, 30*time.Second),
		)).
		Build()
}

// CreateHighPerformanceService 创建高性能服务配置
func CreateHighPerformanceService() *Service {
	config := Config{
		MaxRetries:    5,
		RetryDelay:    500,
		MaxRetryDelay: 10000,
		Timeout:       60000,
	}

	return NewServiceBuilder().
		WithConfig(config).
		Build()
}

// CreateLowLatencyService 创建低延迟服务配置
func CreateLowLatencyService() *Service {
	config := Config{
		MaxRetries:    1,
		RetryDelay:    100,
		MaxRetryDelay: 1000,
		Timeout:       5000,
	}

	return NewServiceBuilder().
		WithConfig(config).
		Build()
}