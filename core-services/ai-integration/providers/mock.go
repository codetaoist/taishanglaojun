package providers

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MockProvider 模拟AI提供者，用于开发和测试
type MockProvider struct {
	logger *zap.Logger
}

// MockConfig 模拟提供者配置
type MockConfig struct {
	Enabled bool `yaml:"enabled"`
}

// NewMockProvider 创建新的模拟AI提供者
func NewMockProvider(logger *zap.Logger) *MockProvider {
	return &MockProvider{
		logger: logger,
	}
}

// GetName 返回提供者名称
func (m *MockProvider) GetName() string {
	return "mock"
}

// IsHealthy 检查提供者健康状态
func (m *MockProvider) IsHealthy(ctx context.Context) bool {
	return true
}

// Chat 处理对话请求
func (m *MockProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	m.logger.Info("Mock AI processing chat request", zap.String("user_id", req.UserID))
	
	// 模拟处理时间
	time.Sleep(500 * time.Millisecond)
	
	// 获取最后一条用户消息
	var userMessage string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userMessage = req.Messages[i].Content
			break
		}
	}
	
	response := &ChatResponse{
		Message: Message{
			Role:    "assistant",
			Content: m.generateChatResponse(userMessage),
		},
		Usage: Usage{
			PromptTokens:     100,
			CompletionTokens: 200,
			TotalTokens:      300,
			Duration:         500 * time.Millisecond,
		},
		SessionID: req.SessionID,
	}
	
	return response, nil
}

// Generate 处理文本生成请求
func (m *MockProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	m.logger.Info("Mock AI generating text", zap.String("prompt", req.Prompt))
	
	// 模拟处理时间
	time.Sleep(500 * time.Millisecond)
	
	content := m.generateTextResponse(req.Prompt)
	
	response := &GenerateResponse{
		Content: content,
		Usage: Usage{
			PromptTokens:     len(req.Prompt) / 4, // 粗略估算
			CompletionTokens: len(content) / 4,
			TotalTokens:      (len(req.Prompt) + len(content)) / 4,
			Duration:         500 * time.Millisecond,
		},
	}
	
	return response, nil
}

// Analyze 处理分析请求
func (m *MockProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	m.logger.Info("Mock AI analyzing content", zap.String("type", req.Type))
	
	// 模拟处理时间
	time.Sleep(300 * time.Millisecond)
	
	response := &AnalyzeResponse{
		Type:       req.Type,
		Confidence: 0.85,
		Result:     m.generateAnalysisResult(req.Content, req.Type),
		Details:    []string{"这是模拟分析结果", "包含详细信息", "用于开发测试"},
		Usage: Usage{
			PromptTokens:     len(req.Content) / 4,
			CompletionTokens: 50,
			TotalTokens:      (len(req.Content) / 4) + 50,
			Duration:         300 * time.Millisecond,
		},
	}
	
	return response, nil
}

// Embed 处理文本嵌入请求
func (m *MockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	m.logger.Info("Mock AI generating embeddings", zap.Int("text_length", len(text)))
	
	// 模拟处理时间
	time.Sleep(200 * time.Millisecond)
	
	// 生成模拟的768维向量
	embedding := make([]float32, 768)
	for i := range embedding {
		embedding[i] = float32(i%100) / 100.0
	}
	
	return embedding, nil
}

// GetModels 返回支持的模型列表
func (m *MockProvider) GetModels() []string {
	return []string{"mock-gpt-3.5", "mock-gpt-4", "mock-claude"}
}

// generateChatResponse 生成对话响应
func (m *MockProvider) generateChatResponse(userMessage string) string {
	if strings.Contains(userMessage, "智慧") || strings.Contains(userMessage, "wisdom") {
		return "感谢您对传统智慧的关注。中华文化博大精深，每一句古语都蕴含着深刻的人生哲理。您想了解哪方面的智慧呢？"
	} else if strings.Contains(userMessage, "解读") || strings.Contains(userMessage, "interpret") {
		return "我很乐意为您解读这段内容。这其中蕴含的智慧可以从多个角度来理解..."
	} else if strings.Contains(userMessage, "推荐") || strings.Contains(userMessage, "recommend") {
		return "基于您的兴趣，我推荐您了解以下相关内容..."
	}
	
	return "这是一个模拟的AI对话响应。在实际环境中，这里会返回真实的AI生成内容。您的问题很有趣，让我们继续探讨吧！"
}

// generateTextResponse 生成文本响应
func (m *MockProvider) generateTextResponse(prompt string) string {
	if strings.Contains(prompt, "解读") || strings.Contains(prompt, "interpret") {
		return m.generateInterpretation(prompt)
	} else if strings.Contains(prompt, "推荐") || strings.Contains(prompt, "recommend") {
		return m.generateRecommendation(prompt)
	} else if strings.Contains(prompt, "分析") || strings.Contains(prompt, "analysis") {
		return m.generateAnalysis(prompt)
	}
	
	return "这是一个模拟的AI文本生成响应。在实际环境中，这里会返回真实的AI生成内容。"
}

// generateAnalysisResult 生成分析结果
func (m *MockProvider) generateAnalysisResult(content, analysisType string) string {
	switch analysisType {
	case "sentiment":
		return "积极"
	case "keywords":
		return "智慧,文化,传统,哲理"
	case "classification":
		return "文化教育类"
	default:
		return "综合分析结果"
	}
}

// generateInterpretation 生成智慧解读响应
func (m *MockProvider) generateInterpretation(prompt string) string {
	return `这是一句充满智慧的话语，体现了中华文化的深厚底蕴。

**核心含义：**
这句话蕴含着深刻的人生哲理，提醒我们在面对困难和挑战时，要保持内心的平静与智慧。

**文化背景：**
这体现了中华传统文化中"修身养性"的理念，强调通过内在修养来应对外在变化。

**现代启示：**
在当今快节奏的生活中，这句话提醒我们要：
1. 保持内心的宁静与专注
2. 用智慧而非情绪来处理问题
3. 在变化中寻找不变的真理

**实践建议：**
可以通过冥想、读书、反思等方式来加深对这句话的理解和实践。`
}

// generateRecommendation 生成相关推荐响应
func (m *MockProvider) generateRecommendation(prompt string) string {
	return `基于您的兴趣，我为您推荐以下相关智慧内容：

**相关经典：**
1. 《道德经》- 老子的智慧结晶，与此理念高度契合
2. 《论语》- 孔子关于修身养性的教导
3. 《庄子》- 逍遥自在的人生哲学

**相关主题：**
- 内心修养与自我提升
- 传统文化中的人生智慧
- 现代生活中的古典哲学应用

**延伸阅读：**
建议深入了解中华传统文化中关于"静心"、"修身"、"养性"的相关内容，这些都能帮助您更好地理解和实践这些智慧。`
}

// generateAnalysis 生成分析响应
func (m *MockProvider) generateAnalysis(prompt string) string {
	return `**智慧分析报告**

**语言特点：**
- 用词精炼，寓意深远
- 体现了中华文化的含蓄之美
- 具有很强的哲理性和指导性

**思想层次：**
1. 表层含义：字面意思的直接理解
2. 深层含义：蕴含的人生哲理
3. 实践意义：对现代生活的指导价值

**文化价值：**
这句话承载着丰富的文化内涵，是中华优秀传统文化的重要组成部分，值得我们深入学习和传承。

**现代意义：**
在当今社会，这样的智慧更显珍贵，能够帮助人们在浮躁的环境中找到内心的平静和方向。`
}