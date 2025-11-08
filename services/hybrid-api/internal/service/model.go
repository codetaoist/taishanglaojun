package service

import (
	"context"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/models"
)

// ModelService 定义模型服务的接口
type ModelService interface {
	// 连接管理
	Connect(ctx context.Context, config *models.ModelConfig) error
	Disconnect(ctx context.Context) error
	Health(ctx context.Context) error

	// 模型管理
	ListModels(ctx context.Context) ([]*models.ModelInfo, error)
	GetModel(ctx context.Context, modelID string) (*models.ModelInfo, error)
	LoadModel(ctx context.Context, modelID string) error
	UnloadModel(ctx context.Context, modelID string) error

	// 文本生成
	GenerateText(ctx context.Context, request *models.TextGenerationRequest) (*models.TextGenerationResponse, error)
	GenerateTextStream(ctx context.Context, request *models.TextGenerationRequest) (<-chan *models.TextGenerationChunk, error)

	// 嵌入生成
	GenerateEmbedding(ctx context.Context, request *models.EmbeddingRequest) (*models.EmbeddingResponse, error)
	GenerateEmbeddings(ctx context.Context, request *models.EmbeddingsRequest) (*models.EmbeddingsResponse, error)

	// 对话管理
	CreateConversation(ctx context.Context, request *models.CreateConversationRequest) (*models.Conversation, error)
	GetConversation(ctx context.Context, conversationID string) (*models.Conversation, error)
	ListConversations(ctx context.Context, userID string) ([]*models.Conversation, error)
	UpdateConversation(ctx context.Context, conversationID string, request *models.UpdateConversationRequest) (*models.Conversation, error)
	DeleteConversation(ctx context.Context, conversationID string) error

	// 消息管理
	AddMessage(ctx context.Context, request *models.AddMessageRequest) (*models.Message, error)
	GetMessages(ctx context.Context, conversationID string, limit int, offset int) ([]*models.Message, error)
	UpdateMessage(ctx context.Context, messageID string, request *models.UpdateMessageRequest) (*models.Message, error)
	DeleteMessage(ctx context.Context, messageID string) error

	// 工具调用
	ExecuteTool(ctx context.Context, request *models.ToolExecutionRequest) (*models.ToolExecutionResponse, error)

	// 模型微调
	CreateFineTuningJob(ctx context.Context, request *models.CreateFineTuningJobRequest) (*models.FineTuningJob, error)
	GetFineTuningJob(ctx context.Context, jobID string) (*models.FineTuningJob, error)
	ListFineTuningJobs(ctx context.Context) ([]*models.FineTuningJob, error)
	CancelFineTuningJob(ctx context.Context, jobID string) error

	// 获取服务信息
	GetServiceInfo(ctx context.Context) (*models.ModelServiceInfo, error)
}

// ModelServiceFactory 模型服务工厂接口
type ModelServiceFactory interface {
	CreateService(config *models.ModelConfig) (ModelService, error)
	GetSupportedProviders() []string
}

// DefaultModelServiceFactory 默认模型服务工厂实现
type DefaultModelServiceFactory struct{}

// NewModelServiceFactory 创建模型服务工厂
func NewModelServiceFactory() ModelServiceFactory {
	return &DefaultModelServiceFactory{}
}

// CreateService 根据配置创建模型服务
func (f *DefaultModelServiceFactory) CreateService(config *models.ModelConfig) (ModelService, error) {
	switch config.Provider() {
	case "openai":
		return NewOpenAIService(config), nil
	case "huggingface":
		return NewHuggingFaceService(config), nil
	case "ollama":
		return NewOllamaService(config), nil
	case "azure":
		return NewAzureService(config), nil
	case "anthropic":
		return NewAnthropicService(config), nil
	case "cohere":
		return NewCohereService(config), nil
	case "palm":
		return NewPalmService(config), nil
	default:
		return NewOpenAIService(config), nil // 默认使用OpenAI
	}
}

// GetSupportedProviders 获取支持的提供商列表
func (f *DefaultModelServiceFactory) GetSupportedProviders() []string {
	return []string{"openai", "huggingface", "ollama", "azure", "anthropic", "cohere", "palm"}
}

// 以下为占位符实现，实际项目中需要完整实现

// NewAzureService 创建Azure OpenAI服务
func NewAzureService(config *models.ModelConfig) ModelService {
	// TODO: 实现Azure OpenAI服务
	return NewOpenAIService(config) // 临时使用OpenAI服务作为占位符
}

// NewAnthropicService 创建Anthropic服务
func NewAnthropicService(config *models.ModelConfig) ModelService {
	// TODO: 实现Anthropic服务
	return NewOpenAIService(config) // 临时使用OpenAI服务作为占位符
}

// NewCohereService 创建Cohere服务
func NewCohereService(config *models.ModelConfig) ModelService {
	// TODO: 实现Cohere服务
	return NewOpenAIService(config) // 临时使用OpenAI服务作为占位符
}

// NewPalmService 创建PaLM服务
func NewPalmService(config *models.ModelConfig) ModelService {
	// TODO: 实现PaLM服务
	return NewOpenAIService(config) // 临时使用OpenAI服务作为占位符
}