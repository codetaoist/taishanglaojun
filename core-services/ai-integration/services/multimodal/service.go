package multimodal

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// Service 多模态服务
type Service struct {
	config           Config
	processorManager *ProcessorManager
	providers        map[string]providers.AIProvider
}

// NewService 创建多模态服务
func NewService(config Config) *Service {
	return &Service{
		config:           config,
		processorManager: NewProcessorManager(),
		providers:        make(map[string]providers.AIProvider),
	}
}

// NewServiceWithManager 使用自定义处理器管理器创建服务
func NewServiceWithManager(config Config, manager *ProcessorManager) *Service {
	return &Service{
		config:           config,
		processorManager: manager,
		providers:        make(map[string]providers.AIProvider),
	}
}

// RegisterProvider 注册AI提供者
func (s *Service) RegisterProvider(name string, provider providers.AIProvider) {
	s.providers[name] = provider
}

// GetProvider 获取AI提供者
func (s *Service) GetProvider(name string) (providers.AIProvider, error) {
	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return provider, nil
}

// Chat 聊天功能
func (s *Service) Chat(
	ctx context.Context,
	providerName string,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return s.processorManager.Process(ctx, ProcessorTypeChat, provider, inputs, config)
}

// Analyze 分析功能
func (s *Service) Analyze(
	ctx context.Context,
	providerName string,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return s.processorManager.Process(ctx, ProcessorTypeAnalysis, provider, inputs, config)
}

// Generate 生成功能
func (s *Service) Generate(
	ctx context.Context,
	providerName string,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return s.processorManager.Process(ctx, ProcessorTypeGeneration, provider, inputs, config)
}

// Translate 翻译功能
func (s *Service) Translate(
	ctx context.Context,
	providerName string,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return s.processorManager.Process(ctx, ProcessorTypeTranslation, provider, inputs, config)
}

// Search 搜索功能
func (s *Service) Search(
	ctx context.Context,
	providerName string,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return s.processorManager.Process(ctx, ProcessorTypeSearch, provider, inputs, config)
}

// ProcessWithType 使用指定类型的处理器处理
func (s *Service) ProcessWithType(
	ctx context.Context,
	processorType ProcessorType,
	providerName string,
	inputs []models.MultimodalInput,
	config models.MultimodalConfig,
) ([]models.MultimodalOutput, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return s.processorManager.Process(ctx, processorType, provider, inputs, config)
}

// ProcessBatch 批量处理
func (s *Service) ProcessBatch(
	ctx context.Context,
	requests []BatchRequest,
) ([]BatchResponse, error) {
	var processRequests []ProcessRequest
	
	for _, req := range requests {
		provider, err := s.GetProvider(req.ProviderName)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider '%s': %w", req.ProviderName, err)
		}

		processRequests = append(processRequests, ProcessRequest{
			ProcessorType: req.ProcessorType,
			Provider:      provider,
			Inputs:        req.Inputs,
			Config:        req.Config,
		})
	}

	responses, err := s.processorManager.ProcessBatch(ctx, processRequests)
	if err != nil {
		return nil, err
	}

	var batchResponses []BatchResponse
	for i, resp := range responses {
		batchResponses = append(batchResponses, BatchResponse{
			Index:        resp.Index,
			ProviderName: requests[i].ProviderName,
			Outputs:      resp.Outputs,
			Error:        resp.Error,
		})
	}

	return batchResponses, nil
}

// ProcessConcurrent 并发处理
func (s *Service) ProcessConcurrent(
	ctx context.Context,
	requests []BatchRequest,
) ([]BatchResponse, error) {
	var processRequests []ProcessRequest
	
	for _, req := range requests {
		provider, err := s.GetProvider(req.ProviderName)
		if err != nil {
			return nil, fmt.Errorf("failed to get provider '%s': %w", req.ProviderName, err)
		}

		processRequests = append(processRequests, ProcessRequest{
			ProcessorType: req.ProcessorType,
			Provider:      provider,
			Inputs:        req.Inputs,
			Config:        req.Config,
		})
	}

	responses, err := s.processorManager.ProcessConcurrent(ctx, processRequests)
	if err != nil {
		return nil, err
	}

	var batchResponses []BatchResponse
	for i, resp := range responses {
		batchResponses = append(batchResponses, BatchResponse{
			Index:        resp.Index,
			ProviderName: requests[i].ProviderName,
			Outputs:      resp.Outputs,
			Error:        resp.Error,
		})
	}

	return batchResponses, nil
}

// ValidateInputs 验证输入
func (s *Service) ValidateInputs(inputs []models.MultimodalInput, config models.MultimodalConfig) error {
	return s.processorManager.ValidateInputs(inputs, config)
}

// PreprocessInputs 预处理输入
func (s *Service) PreprocessInputs(ctx context.Context, inputs []models.MultimodalInput) ([]models.MultimodalInput, error) {
	return s.processorManager.PreprocessInputs(ctx, inputs)
}

// PostprocessOutputs 后处理输出
func (s *Service) PostprocessOutputs(
	ctx context.Context,
	outputs []models.MultimodalOutput,
	expectedTypes []models.OutputType,
) ([]models.MultimodalOutput, error) {
	return s.processorManager.PostprocessOutputs(ctx, outputs, expectedTypes)
}

// GetProcessorInfo 获取处理器信息
func (s *Service) GetProcessorInfo(processorType ProcessorType) (ProcessorInfo, error) {
	return s.processorManager.GetProcessorInfo(processorType)
}

// ListProcessors 列出所有处理器
func (s *Service) ListProcessors() []ProcessorType {
	return s.processorManager.ListProcessors()
}

// ListProviders 列出所有提供者
func (s *Service) ListProviders() []string {
	var providers []string
	for name := range s.providers {
		providers = append(providers, name)
	}
	return providers
}

// GetConfig 获取配置
func (s *Service) GetConfig() Config {
	return s.config
}

// UpdateConfig 更新配置
func (s *Service) UpdateConfig(config Config) {
	s.config = config
}

// RegisterProcessor 注册自定义处理器
func (s *Service) RegisterProcessor(processorType ProcessorType, processor Processor) {
	s.processorManager.RegisterProcessor(processorType, processor)
}

// UnregisterProcessor 注销处理器
func (s *Service) UnregisterProcessor(processorType ProcessorType) {
	s.processorManager.UnregisterProcessor(processorType)
}

// Health 健康检查
func (s *Service) Health(ctx context.Context) error {
	// 检查处理器管理器
	if s.processorManager == nil {
		return fmt.Errorf("processor manager is nil")
	}

	// 检查是否有可用的处理器
	processors := s.processorManager.ListProcessors()
	if len(processors) == 0 {
		return fmt.Errorf("no processors available")
	}

	// 检查是否有可用的提供者
	if len(s.providers) == 0 {
		return fmt.Errorf("no providers available")
	}

	return nil
}

// Shutdown 关闭服务
func (s *Service) Shutdown(ctx context.Context) error {
	// 这里可以添加清理逻辑
	// 例如：关闭连接、清理资源等
	return nil
}

// BatchRequest 批量请求
type BatchRequest struct {
	ProcessorType ProcessorType
	ProviderName  string
	Inputs        []models.MultimodalInput
	Config        models.MultimodalConfig
}

// BatchResponse 批量响应
type BatchResponse struct {
	Index        int
	ProviderName string
	Outputs      []models.MultimodalOutput
	Error        error
}

// ServiceStats 服务统计
type ServiceStats struct {
	ProcessorCount int
	ProviderCount  int
	ProcessorTypes []ProcessorType
	ProviderNames  []string
}

// GetStats 获取服务统计信息
func (s *Service) GetStats() ServiceStats {
	return ServiceStats{
		ProcessorCount: len(s.processorManager.ListProcessors()),
		ProviderCount:  len(s.providers),
		ProcessorTypes: s.processorManager.ListProcessors(),
		ProviderNames:  s.ListProviders(),
	}
}