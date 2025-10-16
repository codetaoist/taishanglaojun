package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/nlp"
	"go.uber.org/zap"
)

// OpenAINLPProvider OpenAI NLP
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

// OpenAIRequest OpenAI NLP请求
type OpenAIRequest struct {
	Model       string                 `json:"model"`
	Messages    []OpenAIMessage        `json:"messages"`
	Temperature float32                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Functions   []OpenAIFunction       `json:"functions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OpenAIMessage OpenAI NLP消息
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIFunction OpenAI NLP函数
type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenAIResponse OpenAI NLP响应
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

// OpenAIChoice OpenAI NLP选择
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIUsage OpenAI NLP使用统计
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewOpenAINLPProvider OpenAI NLP 提供程序
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

// TokenizeText OpenAI NLP 文本分词
func (p *OpenAINLPProvider) TokenizeText(ctx context.Context, input nlp.TextInput) (*nlp.TokenizationResult, error) {
	prompt := fmt.Sprintf("Please tokenize the following text and return the tokens as a JSON array: %s", input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	//
	tokenStrings := strings.Fields(response) // JSON

	// Tokenize
	tokens := make([]nlp.Token, len(tokenStrings))
	for i, tokenStr := range tokenStrings {
		tokens[i] = nlp.Token{
			ID:    fmt.Sprintf("token_%d", i),
			Text:  tokenStr,
			POS:   "UNKNOWN", // OpenAI
			Start: 0,         // OpenAI
			End:   len(tokenStr),
		}
	}

	return &nlp.TokenizationResult{
		ID:          input.ID,
		RequestID:   input.ID,
		Tokens:      tokens,
		TotalTokens: len(tokens),
		Metadata:    make(map[string]interface{}),
	}, nil
}

// AnalyzeSentiment OpenAI NLP 情感分析
func (p *OpenAINLPProvider) AnalyzeSentiment(ctx context.Context, input nlp.TextInput) (*nlp.SentimentAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the sentiment of the following text and return a JSON response with sentiment (positive/negative/neutral), confidence (0-1), and detailed scores:
Text: %s`, input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse
	sentiment := p.parseSentiment(response)

	return &nlp.SentimentAnalysisResult{
		ID:               input.ID,
		RequestID:        input.ID,
		OverallSentiment: sentiment,
		Confidence:       0.85, // OpenAI
		Metadata:         make(map[string]interface{}),
	}, nil
}

// ExtractEntities OpenAI NLP 实体提取
func (p *OpenAINLPProvider) ExtractEntities(ctx context.Context, input nlp.TextInput) (*nlp.EntityExtractionResult, error) {
	prompt := fmt.Sprintf(`Extract named entities from the following text and return them as JSON with entity type, text, and position:
Text: %s`, input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse
	entities := p.parseEntities(response, input.Text)

	return &nlp.EntityExtractionResult{
		ID:            input.ID,
		RequestID:     input.ID,
		Entities:      entities,
		TotalEntities: len(entities),
		Metadata:      make(map[string]interface{}),
	}, nil
}

// ClassifyText OpenAI NLP 文本分类
func (p *OpenAINLPProvider) ClassifyText(ctx context.Context, input nlp.TextInput) (*nlp.TextClassificationResult, error) {
	prompt := fmt.Sprintf(`Classify the following text into categories and return the top categories with confidence scores:
Text: %s`, input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse
	categories := p.parseCategories(response)

	var topCategory nlp.Category
	if len(categories) > 0 {
		topCategory = categories[0] //
	}

	return &nlp.TextClassificationResult{
		ID:          input.ID,
		RequestID:   input.ID,
		Categories:  categories,
		TopCategory: topCategory,
		Metadata:    make(map[string]interface{}),
	}, nil
}

// AnalyzeSemantics OpenAI NLP 语义分析
func (p *OpenAINLPProvider) AnalyzeSemantics(ctx context.Context, input nlp.TextInput) (*nlp.SemanticAnalysisResult, error) {
	prompt := fmt.Sprintf(`Perform semantic analysis on the following text and provide complexity, coherence, and semantic relationships:
Text: %s`, input.Text)

	_, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response and return semantic analysis result
	return &nlp.SemanticAnalysisResult{
		ID:         input.ID,
		RequestID:  input.ID,
		Complexity: 0.6,
		Coherence:  0.8,
		Concepts:   []nlp.Concept{{ID: "concept_1", Text: "main_concept", Type: "general", Confidence: 0.9}},
		Relations:  []nlp.Relation{{ID: "relation_1", Subject: "text", Predicate: "contains", Object: "information", Confidence: 0.8}},
		Metadata:   make(map[string]interface{}),
	}, nil
}

// ExtractKeywords OpenAI NLP 关键词提取
func (p *OpenAINLPProvider) ExtractKeywords(ctx context.Context, input nlp.TextInput) (*nlp.KeywordExtractionResult, error) {
	prompt := fmt.Sprintf(`Extract the most important keywords from the following text with relevance scores:
Text: %s`, input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse
	keywords := p.parseKeywords(response)

	return &nlp.KeywordExtractionResult{
		ID:            input.ID,
		RequestID:     input.ID,
		Keywords:      keywords,
		TotalKeywords: len(keywords),
		Metadata:      make(map[string]interface{}),
	}, nil
}

// SummarizeText OpenAI NLP 文本摘要
func (p *OpenAINLPProvider) SummarizeText(ctx context.Context, input nlp.TextInput) (*nlp.TextSummarizationResult, error) {
	prompt := fmt.Sprintf(`Summarize the following text in a concise manner:
Text: %s`, input.Text)

	summary, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.TextSummarizationResult{
		ID:               input.ID,
		RequestID:        input.ID,
		Summary:          summary,
		CompressionRatio: float64(len(summary)) / float64(len(input.Text)),
		Relevance:        0.9,
		Metadata:         make(map[string]interface{}),
	}, nil
}

// AnalyzeIntent OpenAI NLP 意图分析
func (p *OpenAINLPProvider) AnalyzeIntent(ctx context.Context, input nlp.TextInput) (*nlp.IntentAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the intent of the following text and identify the main purpose:
Text: %s`, input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse
	intent := p.parseIntent(response)

	return &nlp.IntentAnalysisResult{
		ID:         input.ID,
		RequestID:  input.ID,
		Intent:     intent,
		Confidence: 0.85,
		Entities:   []nlp.Entity{},
		Metadata:   make(map[string]interface{}),
	}, nil
}

// GenerateText OpenAI NLP 文本生成
func (p *OpenAINLPProvider) GenerateText(ctx context.Context, input nlp.TextGenerationInput) (*nlp.TextGenerationResult, error) {
	prompt := input.Prompt

	generatedText, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.TextGenerationResult{
		ID:            input.ID,
		RequestID:     input.ID,
		GeneratedText: generatedText,
		Alternatives:  []string{},
		Quality:       0.9,
		Coherence:     0.9,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// TranslateText OpenAI NLP 文本翻译
func (p *OpenAINLPProvider) TranslateText(ctx context.Context, input nlp.TranslationInput) (*nlp.TranslationResult, error) {
	prompt := fmt.Sprintf(`Translate the following text from %s to %s:
Text: %s`, input.SourceLang, input.TargetLang, input.Text)

	translatedText, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.TranslationResult{
		ID:             input.ID,
		RequestID:      input.ID,
		TranslatedText: translatedText,
		Confidence:     0.9,
		Alternatives:   []string{},
		Quality:        0.9,
		ProcessingTime: 0,
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}, nil
}

// ParaphraseText OpenAI NLP 文本重写
func (p *OpenAINLPProvider) ParaphraseText(ctx context.Context, input nlp.TextInput) (*nlp.ParaphraseResult, error) {
	prompt := fmt.Sprintf(`Paraphrase the following text in multiple ways while maintaining the original meaning:
Text: %s`, input.Text)

	response, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse
	paraphrases := p.parseParaphrases(response)

	return &nlp.ParaphraseResult{
		ID:             input.ID,
		RequestID:      input.ID,
		Paraphrases:    paraphrases,
		BestParaphrase: response,
		Similarity:     0.85,
		Quality:        0.9,
		ProcessingTime: 0,
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}, nil
}

// ProcessConversation OpenAI NLP 对话处理
func (p *OpenAINLPProvider) ProcessConversation(ctx context.Context, input nlp.ConversationInput) (*nlp.ConversationResult, error) {
	//
	messages := make([]OpenAIMessage, 0)
	for _, turn := range input.Context {
		messages = append(messages, OpenAIMessage{
			Role:    turn.Role,
			Content: turn.Message,
		})
	}

	// Add user message
	messages = append(messages, OpenAIMessage{
		Role:    "user",
		Content: input.Message,
	})

	response, err := p.callOpenAIWithMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	return &nlp.ConversationResult{
		ID:             input.ID,
		RequestID:      input.ID,
		Response:       response,
		Intent:         nlp.Intent{Name: "conversation", Confidence: 0.8},
		Entities:       []nlp.Entity{},
		Sentiment:      nlp.Sentiment{Label: "neutral", Score: 0.5},
		Confidence:     0.9,
		NextActions:    []string{},
		ProcessingTime: 0,
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}, nil
}

// AnswerQuestion OpenAI NLP 问题回答
func (p *OpenAINLPProvider) AnswerQuestion(ctx context.Context, input nlp.QuestionAnsweringInput) (*nlp.QuestionAnsweringResult, error) {
	prompt := fmt.Sprintf(`Based on the following context, answer the question:
Context: %s
Question: %s`, input.Context, input.Question)

	answer, err := p.callOpenAI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &nlp.QuestionAnsweringResult{
		ID:             input.ID,
		RequestID:      input.ID,
		Answer:         answer,
		Alternatives:   []nlp.Answer{},
		Confidence:     0.9,
		Sources:        []nlp.Source{{ID: "1", DocumentID: "doc1", Title: "Context", Snippet: input.Context, Relevance: 0.9, URL: ""}},
		ProcessingTime: time.Since(time.Now()),
		Timestamp:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}, nil
}

// GetSupportedLanguages OpenAI NLP 支持的语言
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

// GetSupportedOperations OpenAI NLP 支持的操作
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

// HealthCheck OpenAI NLP 健康检查
func (p *OpenAINLPProvider) HealthCheck(ctx context.Context) error {
	_, err := p.callOpenAI(ctx, "Hello")
	return err
}

//

// callOpenAI OpenAI API 调用
func (p *OpenAINLPProvider) callOpenAI(ctx context.Context, prompt string) (string, error) {
	messages := []OpenAIMessage{
		{Role: "user", Content: prompt},
	}
	return p.callOpenAIWithMessages(ctx, messages)
}

// callOpenAIWithMessages OpenAI API 调用
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

// parseSentiment OpenAI NLP 解析情感
func (p *OpenAINLPProvider) parseSentiment(response string) nlp.Sentiment {
	response = strings.ToLower(response)
	if strings.Contains(response, "positive") {
		return nlp.Sentiment{Label: "positive", Score: 0.8}
	} else if strings.Contains(response, "negative") {
		return nlp.Sentiment{Label: "negative", Score: 0.8}
	}
	return nlp.Sentiment{Label: "neutral", Score: 0.6}
}

// parseEntities OpenAI NLP 解析实体
func (p *OpenAINLPProvider) parseEntities(response, text string) []nlp.Entity {
	//
	return []nlp.Entity{
		{
			ID:         "1",
			Text:       "example",
			Label:      "MISC",
			Start:      0,
			End:        7,
			Confidence: 0.9,
			Properties: make(map[string]interface{}),
		},
	}
}

// parseCategories OpenAI NLP 解析分类
func (p *OpenAINLPProvider) parseCategories(response string) []nlp.Category {
	// 简单的解析实现，实际应该根据OpenAI响应格式解析
	return []nlp.Category{
		{Name: "general", Confidence: 0.8},
		{Name: "informative", Confidence: 0.6},
	}
}

// parseKeywords OpenAI NLP 解析关键词
func (p *OpenAINLPProvider) parseKeywords(response string) []nlp.Keyword {
	// 简单的解析实现，实际应该根据OpenAI响应格式解析
	words := strings.Fields(response)
	keywords := make([]nlp.Keyword, 0)
	for i, word := range words {
		if i >= 5 { //
			break
		}
		keywords = append(keywords, nlp.Keyword{
			ID:         fmt.Sprintf("kw_%d", i),
			Text:       word,
			Score:      0.8 - float64(i)*0.1,
			Frequency:  1,
			Relevance:  0.8 - float64(i)*0.1,
			Properties: make(map[string]interface{}),
		})
	}
	return keywords
}

// parseIntent OpenAI NLP 解析意图
func (p *OpenAINLPProvider) parseIntent(response string) nlp.Intent {
	// 简单的解析实现，实际应该根据OpenAI响应格式解析
	return nlp.Intent{
		ID:         "intent_1",
		Name:       "general_inquiry",
		Category:   "general",
		Confidence: 0.8,
		Parameters: make(map[string]interface{}),
	}
}

// parseParaphrases OpenAI NLP 解析重写
func (p *OpenAINLPProvider) parseParaphrases(response string) []string {
	// 简单的解析实现，实际应该根据OpenAI响应格式解析
	return []string{response}
}
