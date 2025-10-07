package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/taishanglaojun/core-services/ai-integration/nlp"
	"go.uber.org/zap"
)

// OpenAINLPProvider OpenAI NLP服务提供商
type OpenAINLPProvider struct {
	config     OpenAINLPConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// OpenAINLPConfig OpenAI NLP配置
type OpenAINLPConfig struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Model       string        `json:"model"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// OpenAIRequest OpenAI请求结构
type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []OpenAIMessage        `json:"messages"`
	Temperature float32                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Functions   []OpenAIFunction       `json:"functions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OpenAIMessage OpenAI消息结构
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIFunction OpenAI函数结构
type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenAIResponse OpenAI响应结构
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice OpenAI选择结构
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage OpenAI使用情况结构
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewOpenAINLPProvider 创建OpenAI NLP提供商
func NewOpenAINLPProvider(config OpenAINLPConfig, logger *zap.Logger) *OpenAINLPProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 1000
	}

	return &OpenAINLPProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// TokenizeText 分词
func (p *OpenAINLPProvider) TokenizeText(ctx context.Context, input nlp.TextInput) (*nlp.TokenizationResult, error) {
	prompt := fmt.Sprintf("Please tokenize the following text and return the tokens as a JSON array: %s", input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析响应
	tokens := strings.Fields(response) // 简化实现
	
	return &nlp.TokenizationResult{
		ID:          input.ID,
		Tokens:      tokens,
		TotalTokens: len(tokens),
		Language:    input.Language,
		Metadata:    make(map[string]interface{}),
	}, nil
}

// AnalyzeSentiment 情感分析
func (p *OpenAINLPProvider) AnalyzeSentiment(ctx context.Context, input nlp.TextInput) (*nlp.SentimentAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the sentiment of the following text and return a JSON response with sentiment (positive/negative/neutral), confidence (0-1), and detailed scores:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的情感分析结果解析
	sentiment := p.parseSentiment(response)
	
	return &nlp.SentimentAnalysisResult{
		ID:               input.ID,
		OverallSentiment: sentiment,
		Confidence:       0.85, // 简化实现
		DetailedScores: nlp.SentimentScores{
			Positive: 0.7,
			Negative: 0.2,
			Neutral:  0.1,
		},
		Language: input.Language,
		Metadata: make(map[string]interface{}),
	}, nil
}

// ExtractEntities 实体提取
func (p *OpenAINLPProvider) ExtractEntities(ctx context.Context, input nlp.TextInput) (*nlp.EntityExtractionResult, error) {
	prompt := fmt.Sprintf(`Extract named entities from the following text and return them as JSON with entity type, text, and position:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的实体提取结果解析
	entities := p.parseEntities(response, input.Text)
	
	return &nlp.EntityExtractionResult{
		ID:            input.ID,
		Entities:      entities,
		TotalEntities: len(entities),
		Language:      input.Language,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// ClassifyText 文本分类
func (p *OpenAINLPProvider) ClassifyText(ctx context.Context, input nlp.TextInput) (*nlp.TextClassificationResult, error) {
	prompt := fmt.Sprintf(`Classify the following text into categories and return the top categories with confidence scores:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的分类结果解析
	categories := p.parseCategories(response)
	
	return &nlp.TextClassificationResult{
		ID:          input.ID,
		Categories:  categories,
		TopCategory: categories[0], // 假设第一个是最高分
		Language:    input.Language,
		Metadata:    make(map[string]interface{}),
	}, nil
}

// AnalyzeSemantics 语义分析
func (p *OpenAINLPProvider) AnalyzeSemantics(ctx context.Context, input nlp.TextInput) (*nlp.SemanticAnalysisResult, error) {
	prompt := fmt.Sprintf(`Perform semantic analysis on the following text and provide complexity, coherence, and semantic relationships:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的语义分析结果
	return &nlp.SemanticAnalysisResult{
		ID:         input.ID,
		Complexity: 0.6,
		Coherence:  0.8,
		Topics:     []string{"general", "informative"},
		Concepts:   []nlp.Concept{{Name: "main_concept", Confidence: 0.9}},
		Relations:  []nlp.SemanticRelation{{Subject: "text", Predicate: "contains", Object: "information"}},
		Language:   input.Language,
		Metadata:   make(map[string]interface{}),
	}, nil
}

// ExtractKeywords 关键词提取
func (p *OpenAINLPProvider) ExtractKeywords(ctx context.Context, input nlp.TextInput) (*nlp.KeywordExtractionResult, error) {
	prompt := fmt.Sprintf(`Extract the most important keywords from the following text with relevance scores:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的关键词提取结果
	keywords := p.parseKeywords(response)
	
	return &nlp.KeywordExtractionResult{
		ID:            input.ID,
		Keywords:      keywords,
		TotalKeywords: len(keywords),
		Language:      input.Language,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// SummarizeText 文本摘要
func (p *OpenAINLPProvider) SummarizeText(ctx context.Context, input nlp.TextInput) (*nlp.TextSummarizationResult, error) {
	prompt := fmt.Sprintf(`Summarize the following text in a concise manner:
Text: %s`, input.Text)
	
	summary, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.TextSummarizationResult{
		ID:               input.ID,
		Summary:          summary,
		CompressionRatio: float64(len(summary)) / float64(len(input.Text)),
		Relevance:        0.9,
		KeyPoints:        []string{summary}, // 简化实现
		Language:         input.Language,
		Metadata:         make(map[string]interface{}),
	}, nil
}

// AnalyzeIntent 意图分析
func (p *OpenAINLPProvider) AnalyzeIntent(ctx context.Context, input nlp.TextInput) (*nlp.IntentAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the intent of the following text and identify the main purpose:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的意图分析结果
	intent := p.parseIntent(response)
	
	return &nlp.IntentAnalysisResult{
		ID:         input.ID,
		Intent:     intent,
		Confidence: 0.85,
		Entities:   []nlp.IntentEntity{},
		Language:   input.Language,
		Metadata:   make(map[string]interface{}),
	}, nil
}

// GenerateText 文本生成
func (p *OpenAINLPProvider) GenerateText(ctx context.Context, input nlp.TextGenerationInput) (*nlp.TextGenerationResult, error) {
	prompt := input.Prompt
	if input.Context != "" {
		prompt = fmt.Sprintf("Context: %s\nPrompt: %s", input.Context, input.Prompt)
	}
	
	generatedText, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.TextGenerationResult{
		ID:            input.ID,
		GeneratedText: generatedText,
		Quality:       0.9,
		Creativity:    0.8,
		Coherence:     0.9,
		Language:      input.Language,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// TranslateText 文本翻译
func (p *OpenAINLPProvider) TranslateText(ctx context.Context, input nlp.TranslationInput) (*nlp.TranslationResult, error) {
	prompt := fmt.Sprintf(`Translate the following text from %s to %s:
Text: %s`, input.SourceLang, input.TargetLang, input.Text)
	
	translatedText, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.TranslationResult{
		ID:             input.ID,
		TranslatedText: translatedText,
		SourceLang:     input.SourceLang,
		TargetLang:     input.TargetLang,
		Confidence:     0.9,
		Alternatives:   []string{},
		Metadata:       make(map[string]interface{}),
	}, nil
}

// ParaphraseText 文本改写
func (p *OpenAINLPProvider) ParaphraseText(ctx context.Context, input nlp.TextInput) (*nlp.ParaphraseResult, error) {
	prompt := fmt.Sprintf(`Paraphrase the following text in multiple ways while maintaining the original meaning:
Text: %s`, input.Text)
	
	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 简化的改写结果
	paraphrases := []string{response} // 实际中应该解析多个改写版本
	
	return &nlp.ParaphraseResult{
		ID:           input.ID,
		Paraphrases:  paraphrases,
		Similarity:   0.85,
		Diversity:    0.7,
		Language:     input.Language,
		Metadata:     make(map[string]interface{}),
	}, nil
}

// ProcessConversation 处理对话
func (p *OpenAINLPProvider) ProcessConversation(ctx context.Context, input nlp.ConversationInput) (*nlp.ConversationResult, error) {
	// 构建对话历史
	messages := make([]OpenAIMessage, 0)
	for _, msg := range input.History {
		messages = append(messages, OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	
	// 添加当前消息
	messages = append(messages, OpenAIMessage{
		Role:    "user",
		Content: input.Message,
	})

	response, err := p.callOpenAIWithMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	return &nlp.ConversationResult{
		ID:         input.ID,
		Response:   response,
		Confidence: 0.9,
		Intent:     nlp.Intent{Name: "conversation", Confidence: 0.8},
		Entities:   []nlp.IntentEntity{},
		Context:    make(map[string]interface{}),
		Language:   input.Language,
		Metadata:   make(map[string]interface{}),
	}, nil
}

// AnswerQuestion 问答
func (p *OpenAINLPProvider) AnswerQuestion(ctx context.Context, input nlp.QuestionAnsweringInput) (*nlp.QuestionAnsweringResult, error) {
	prompt := fmt.Sprintf(`Based on the following context, answer the question:
Context: %s
Question: %s`, input.Context, input.Question)
	
	answer, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.QuestionAnsweringResult{
		ID:         input.ID,
		Answer:     answer,
		Confidence: 0.9,
		Sources:    []nlp.AnswerSource{{Text: input.Context, Confidence: 0.9}},
		Language:   input.Language,
		Metadata:   make(map[string]interface{}),
	}, nil
}

// GetSupportedLanguages 获取支持的语言
func (p *OpenAINLPProvider) GetSupportedLanguages() []nlp.Language {
	return []nlp.Language{
		nlp.LanguageEnglish,
		nlp.LanguageChinese,
		nlp.LanguageSpanish,
		nlp.LanguageFrench,
		nlp.LanguageGerman,
		nlp.LanguageJapanese,
		nlp.LanguageKorean,
		nlp.LanguageAuto,
	}
}

// GetSupportedOperations 获取支持的操作
func (p *OpenAINLPProvider) GetSupportedOperations() []nlp.OperationType {
	return []nlp.OperationType{
		nlp.OpTokenization,
		nlp.OpSentiment,
		nlp.OpEntityExtraction,
		nlp.OpClassification,
		nlp.OpSemanticAnalysis,
		nlp.OpKeywordExtraction,
		nlp.OpSummarization,
		nlp.OpIntentAnalysis,
		nlp.OpTextGeneration,
		nlp.OpTranslation,
		nlp.OpParaphrase,
		nlp.OpConversation,
		nlp.OpQuestionAnswering,
	}
}

// HealthCheck 健康检查
func (p *OpenAINLPProvider) HealthCheck(ctx context.Context) error {
	_, err := p.callOpenAI(ctx, "Hello")
	return err
}

// 私有方法

// callOpenAI 调用OpenAI API
func (p *OpenAINLPProvider) callOpenAI(ctx context.Context, prompt string) (string, error) {
	messages := []OpenAIMessage{
		{Role: "user", Content: prompt},
	}
	return p.callOpenAIWithMessages(ctx, messages)
}

// callOpenAIWithMessages 使用消息调用OpenAI API
func (p *OpenAINLPProvider) callOpenAIWithMessages(ctx context.Context, messages []OpenAIMessage) (string, error) {
	request := OpenAIRequest{
		Model:       p.config.Model,
		Messages:    messages,
		Temperature: p.config.Temperature,
		MaxTokens:   p.config.MaxTokens,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var response OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// 解析方法（简化实现）

func (p *OpenAINLPProvider) parseSentiment(response string) nlp.Sentiment {
	response = strings.ToLower(response)
	if strings.Contains(response, "positive") {
		return nlp.Sentiment{Label: "positive", Score: 0.8}
	} else if strings.Contains(response, "negative") {
		return nlp.Sentiment{Label: "negative", Score: 0.8}
	}
	return nlp.Sentiment{Label: "neutral", Score: 0.6}
}

func (p *OpenAINLPProvider) parseEntities(response, text string) []nlp.Entity {
	// 简化实现 - 实际中需要更复杂的解析
	return []nlp.Entity{
		{
			Text:       "example",
			Type:       "MISC",
			StartIndex: 0,
			EndIndex:   7,
			Confidence: 0.9,
		},
	}
}

func (p *OpenAINLPProvider) parseCategories(response string) []nlp.Category {
	// 简化实现
	return []nlp.Category{
		{Name: "general", Confidence: 0.8},
		{Name: "informative", Confidence: 0.6},
	}
}

func (p *OpenAINLPProvider) parseKeywords(response string) []nlp.Keyword {
	// 简化实现
	words := strings.Fields(response)
	keywords := make([]nlp.Keyword, 0)
	for i, word := range words {
		if i >= 5 { // 限制关键词数量
			break
		}
		keywords = append(keywords, nlp.Keyword{
			Text:      word,
			Relevance: 0.8 - float64(i)*0.1,
			Type:      "general",
		})
	}
	return keywords
}

func (p *OpenAINLPProvider) parseIntent(response string) nlp.Intent {
	// 简化实现
	return nlp.Intent{
		Name:        "general_inquiry",
		Confidence:  0.8,
		Description: "General information request",
	}
}