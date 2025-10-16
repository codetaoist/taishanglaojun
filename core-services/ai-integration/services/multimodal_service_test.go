package services

import (
	"context"
	"testing"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/models"
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIProvider 模拟AI提供者
type MockAIProvider struct {
	mock.Mock
}

func (m *MockAIProvider) Chat(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.ChatResponse), args.Error(1)
}

func (m *MockAIProvider) Generate(ctx context.Context, req *providers.GenerateRequest) (*providers.GenerateResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.GenerateResponse), args.Error(1)
}

func (m *MockAIProvider) Analyze(ctx context.Context, req *providers.AnalyzeRequest) (*providers.AnalyzeResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.AnalyzeResponse), args.Error(1)
}

func (m *MockAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockAIProvider) IntentRecognition(ctx context.Context, req providers.IntentRequest) (*providers.IntentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.IntentResponse), args.Error(1)
}

func (m *MockAIProvider) SentimentAnalysis(ctx context.Context, req providers.SentimentRequest) (*providers.SentimentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.SentimentResponse), args.Error(1)
}

func (m *MockAIProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAIProvider) GetModels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAIProvider) GetCapabilities() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAIProvider) GenerateImage(ctx context.Context, req *providers.ImageGenerateRequest) (*providers.ImageGenerateResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.ImageGenerateResponse), args.Error(1)
}

func (m *MockAIProvider) AnalyzeImage(ctx context.Context, req *providers.ImageAnalyzeRequest) (*providers.ImageAnalyzeResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.ImageAnalyzeResponse), args.Error(1)
}

func (m *MockAIProvider) EditImage(ctx context.Context, req *providers.ImageEditRequest) (*providers.ImageEditResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*providers.ImageEditResponse), args.Error(1)
}

// MockFileService 模拟文件服务
type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) SaveFile(ctx context.Context, data []byte, filename string) (string, error) {
	args := m.Called(ctx, data, filename)
	return args.String(0), args.Error(1)
}

func (m *MockFileService) GetFile(ctx context.Context, url string) ([]byte, error) {
	args := m.Called(ctx, url)
	return args.Get(0).([]byte), args.Error(1)
}

// MockAudioService 模拟音频服务
type MockAudioService struct {
	mock.Mock
}

func (m *MockAudioService) GetDuration(data []byte) (float64, error) {
	args := m.Called(data)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockAudioService) ProcessAudio(data []byte, format string) ([]byte, error) {
	args := m.Called(data, format)
	return args.Get(0).([]byte), args.Error(1)
}

// MockImageService 模拟图像服务
type MockImageService struct {
	mock.Mock
}

func (m *MockImageService) GetDimensions(data []byte) (models.ImageDimensions, error) {
	args := m.Called(data)
	return args.Get(0).(models.ImageDimensions), args.Error(1)
}

func (m *MockImageService) ProcessImage(data []byte, format string) ([]byte, error) {
	args := m.Called(data, format)
	return args.Get(0).([]byte), args.Error(1)
}

// MockVideoService 模拟视频服务
type MockVideoService struct {
	mock.Mock
}

func (m *MockVideoService) GetInfo(data []byte) (*VideoInfo, error) {
	args := m.Called(data)
	return args.Get(0).(*VideoInfo), args.Error(1)
}

func (m *MockVideoService) ProcessVideo(data []byte, format string) ([]byte, error) {
	args := m.Called(data, format)
	return args.Get(0).([]byte), args.Error(1)
}

// MockRepository 模拟存储库
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveMultimodalSession(ctx context.Context, session *models.MultimodalSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockRepository) GetMultimodalSession(ctx context.Context, sessionID string) (*models.MultimodalSession, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.MultimodalSession), args.Error(1)
}

func (m *MockRepository) SaveMultimodalMessage(ctx context.Context, message *models.MultimodalMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockRepository) GetMultimodalMessages(ctx context.Context, sessionID string) ([]*models.MultimodalMessage, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*models.MultimodalMessage), args.Error(1)
}

// 创建测试用的多模态服务
func createTestMultimodalService() (*MultimodalService, *MockAIProvider, *MockFileService, *MockAudioService, *MockImageService, *MockVideoService, *MockRepository) {
	mockProvider := &MockAIProvider{}
	mockFileService := &MockFileService{}
	mockAudioService := &MockAudioService{}
	mockImageService := &MockImageService{}
	mockVideoService := &MockVideoService{}
	mockRepository := &MockRepository{}

	service := NewMultimodalService(
		map[string]providers.AIProvider{"test": mockProvider},
		mockProvider,
		mockRepository,
		mockFileService,
		mockAudioService,
		mockImageService,
		mockVideoService,
	)

	return service, mockProvider, mockFileService, mockAudioService, mockImageService, mockVideoService, mockRepository
}

// TestMultimodalChat 测试多模态聊天功能
func TestMultimodalChat(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟响应
	mockProvider.On("Chat", mock.Anything, mock.AnythingOfType("*providers.ChatRequest")).Return(
		&providers.ChatResponse{
			Content: "Hello! How can I help you?",
			Model:   "test-model",
			Usage: providers.TokenUsage{
				TotalTokens: 10,
			},
		}, nil)

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "Hello",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model:       "test-model",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalChat(ctx, mockProvider, inputs, config)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, outputs, 1)
	assert.Equal(t, models.OutputTypeText, outputs[0].Type)
	assert.Equal(t, "Hello! How can I help you?", outputs[0].TextOutput.Content)

	mockProvider.AssertExpectations(t)
}

// TestMultimodalAnalysis 测试多模态分析功能
func TestMultimodalAnalysis(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟响应
	mockProvider.On("Analyze", mock.Anything, mock.AnythingOfType("*providers.AnalyzeRequest")).Return(
		&providers.AnalyzeResponse{
			Summary: "This is a test analysis",
			Model:   "test-model",
			Usage: providers.TokenUsage{
				TotalTokens: 15,
			},
		}, nil)

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "Analyze this text",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model: "test-model",
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalAnalysis(ctx, mockProvider, inputs, config)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, outputs, 1)
	assert.Equal(t, models.OutputTypeText, outputs[0].Type)
	assert.Equal(t, "This is a test analysis", outputs[0].TextOutput.Content)

	mockProvider.AssertExpectations(t)
}

// TestMultimodalGeneration 测试多模态生成功能
func TestMultimodalGeneration(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟响应
	mockProvider.On("Generate", mock.Anything, mock.AnythingOfType("*providers.GenerateRequest")).Return(
		&providers.GenerateResponse{
			Text:  "Generated text content",
			Model: "test-model",
			Usage: providers.TokenUsage{
				TotalTokens: 20,
			},
		}, nil)

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "Generate something",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model:       "test-model",
		OutputTypes: []models.MultimodalOutputType{models.OutputTypeText},
		Temperature: 0.8,
		MaxTokens:   150,
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalGeneration(ctx, mockProvider, inputs, config)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, outputs, 1)
	assert.Equal(t, models.OutputTypeText, outputs[0].Type)
	assert.Equal(t, "Generated text content", outputs[0].TextOutput.Content)

	mockProvider.AssertExpectations(t)
}

// TestMultimodalTranslation 测试多模态翻译功能（使用Generate方法模拟翻译）
func TestMultimodalTranslation(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟响应
	mockProvider.On("Generate", mock.Anything, mock.AnythingOfType("providers.GenerateRequest")).Return(
		&providers.GenerateResponse{
			Content: "Translated text",
			Usage: providers.Usage{
				TotalTokens: 12,
			},
		}, nil)

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "Hello world",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model:          "test-model",
		TargetLanguage: "zh",
		SourceLanguage: "en",
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalTranslation(ctx, mockProvider, inputs, config)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, outputs, 1)
	assert.Equal(t, models.OutputTypeText, outputs[0].Type)
	assert.Equal(t, "Translated text", outputs[0].TextOutput.Content)
	assert.Equal(t, "zh", outputs[0].TextOutput.Language)

	mockProvider.AssertExpectations(t)
}

// TestMultimodalSearch 测试多模态搜索功能
func TestMultimodalSearch(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟响应
	mockProvider.On("Embed", mock.Anything, "search query").Return(
		[]float32{0.1, 0.2, 0.3, 0.4, 0.5}, nil)

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "search query",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model: "test-model",
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalSearch(ctx, mockProvider, inputs, config)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, outputs, 1)
	assert.Equal(t, models.OutputTypeText, outputs[0].Type)
	assert.Contains(t, outputs[0].TextOutput.Content, "Search results for: search query")
	assert.Contains(t, outputs[0].TextOutput.Content, "Embedding dimensions: 5")

	mockProvider.AssertExpectations(t)
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟错误响应
	mockProvider.On("Chat", mock.Anything, mock.AnythingOfType("*providers.ChatRequest")).Return(
		(*providers.ChatResponse)(nil), assert.AnError)

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "Hello",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model: "test-model",
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalChat(ctx, mockProvider, inputs, config)

	// 验证错误处理
	assert.Error(t, err)
	assert.Nil(t, outputs)
	assert.IsType(t, &MultimodalError{}, err)

	multimodalErr := err.(*MultimodalError)
	assert.Equal(t, ErrorTypeProvider, multimodalErr.Type)

	mockProvider.AssertExpectations(t)
}

// TestValidationError 测试输入验证错误
func TestValidationError(t *testing.T) {
	service, _, _, _, _, _, _ := createTestMultimodalService()

	// 测试空输入
	ctx := context.Background()
	outputs, err := service.callMultimodalChat(ctx, nil, nil, models.MultimodalConfig{})

	// 验证验证错误
	assert.Error(t, err)
	assert.Nil(t, outputs)
	assert.IsType(t, &MultimodalError{}, err)

	multimodalErr := err.(*MultimodalError)
	assert.Equal(t, ErrorTypeValidation, multimodalErr.Type)
	assert.False(t, multimodalErr.Retry)
}

// TestRetryMechanism 测试重试机制
func TestRetryMechanism(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置第一次调用失败，第二次成功
	mockProvider.On("Chat", mock.Anything, mock.AnythingOfType("*providers.ChatRequest")).Return(
		(*providers.ChatResponse)(nil), assert.AnError).Once()

	mockProvider.On("Chat", mock.Anything, mock.AnythingOfType("*providers.ChatRequest")).Return(
		&providers.ChatResponse{
			Content: "Success after retry",
			Model:   "test-model",
			Usage: providers.TokenUsage{
				TotalTokens: 10,
			},
		}, nil).Once()

	// 创建测试请求
	inputs := []models.MultimodalInput{
		{
			Type: models.InputTypeText,
			TextInput: &models.TextInput{
				Content:  "Hello",
				Language: "en",
			},
		},
	}

	config := models.MultimodalConfig{
		Model: "test-model",
	}

	// 执行测试
	ctx := context.Background()
	outputs, err := service.callMultimodalChat(ctx, mockProvider, inputs, config)

	// 验证重试成功
	assert.NoError(t, err)
	assert.Len(t, outputs, 1)
	assert.Equal(t, "Success after retry", outputs[0].TextOutput.Content)

	mockProvider.AssertExpectations(t)
}

// TestStreamingResponse 测试流式响应
func TestStreamingResponse(t *testing.T) {
	service, mockProvider, _, _, _, _, _ := createTestMultimodalService()

	// 设置模拟响应
	mockProvider.On("Chat", mock.Anything, mock.AnythingOfType("*providers.ChatRequest")).Return(
		&providers.ChatResponse{
			Content: "Streaming response",
			Model:   "test-model",
			Usage: providers.TokenUsage{
				TotalTokens: 10,
			},
		}, nil)

	// 创建测试请求
	request := &models.MultimodalRequest{
		Type: models.MultimodalTypeChat,
		Inputs: []models.MultimodalInput{
			{
				Type: models.InputTypeText,
				TextInput: &models.TextInput{
					Content:  "Hello",
					Language: "en",
				},
			},
		},
		Config: models.MultimodalConfig{
			Model: "test-model",
		},
	}

	// 执行测试
	ctx := context.Background()
	outputChan, err := service.streamAIProvider(ctx, mockProvider, request)

	// 验证流式响应
	assert.NoError(t, err)
	assert.NotNil(t, outputChan)

	// 读取流式输出
	select {
	case output := <-outputChan:
		assert.Equal(t, models.OutputTypeText, output.Type)
		assert.Equal(t, "Streaming response", output.TextOutput.Content)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for streaming response")
	}

	mockProvider.AssertExpectations(t)
}