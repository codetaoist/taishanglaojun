package services

import (
	"context"
	"fmt"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/providers"
)

// AIService AI服务
type AIService struct {
	manager *providers.Manager
}

// NewAIService 创建新的AI服务
func NewAIService(manager *providers.Manager) *AIService {
	return &AIService{
		manager: manager,
	}
}

// GetEmbedding 获取文本嵌入向量
func (s *AIService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Embed(ctx, text)
}

// Chat 聊天对话
func (s *AIService) Chat(ctx context.Context, req providers.ChatRequest) (*providers.ChatResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Chat(ctx, req)
}

// Generate 生成内容
func (s *AIService) Generate(ctx context.Context, req providers.GenerateRequest) (*providers.GenerateResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Generate(ctx, req)
}

// Analyze 分析内容
func (s *AIService) Analyze(ctx context.Context, req providers.AnalyzeRequest) (*providers.AnalyzeResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Analyze(ctx, req)
}

// IntentRecognition 意图识别
func (s *AIService) IntentRecognition(ctx context.Context, req providers.IntentRequest) (*providers.IntentResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.IntentRecognition(ctx, req)
}

// SentimentAnalysis 情感分析
func (s *AIService) SentimentAnalysis(ctx context.Context, req providers.SentimentRequest) (*providers.SentimentResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.SentimentAnalysis(ctx, req)
}

// GenerateSummary 生成摘要
func (s *AIService) GenerateSummary(ctx context.Context, req providers.ChatRequest) (*providers.ChatResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	// 添加摘要生成的系统提示
	req.Messages = append(req.Messages, providers.Message{
		Role:    "system",
		Content: "请为以下内容生成简洁的摘要",
	})

	return provider.Chat(ctx, req)
}

// GenerateExplanation 生成解释
func (s *AIService) GenerateExplanation(ctx context.Context, req providers.ChatRequest) (*providers.ChatResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	// 添加解释生成的系统提示
	req.Messages = append(req.Messages, providers.Message{
		Role:    "system",
		Content: "请详细解释以下内容",
	})

	return provider.Chat(ctx, req)
}

// GenerateTranslation 生成翻译
func (s *AIService) GenerateTranslation(ctx context.Context, req providers.ChatRequest) (*providers.ChatResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	// 添加翻译的系统提示
	req.Messages = append(req.Messages, providers.Message{
		Role:    "system",
		Content: "请翻译以下内容",
	})

	return provider.Chat(ctx, req)
}

// ExtractKeywords 提取关键词
func (s *AIService) ExtractKeywords(ctx context.Context, req providers.AnalyzeRequest) (*providers.AnalyzeResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Analyze(ctx, req)
}

// CalculateSimilarity 计算相似度
func (s *AIService) CalculateSimilarity(ctx context.Context, req providers.AnalyzeRequest) (*providers.AnalyzeResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Analyze(ctx, req)
}

// GenerateEmbedding 生成嵌入向量
func (s *AIService) GenerateEmbedding(ctx context.Context, req providers.AnalyzeRequest) (*providers.AnalyzeResponse, error) {
	provider, err := s.manager.GetDefaultProvider()
	if err != nil {
		return nil, fmt.Errorf("获取AI提供者失败: %w", err)
	}

	return provider.Analyze(ctx, req)
}

