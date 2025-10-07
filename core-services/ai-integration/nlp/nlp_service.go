package nlp

import (
	"context"
	"time"
)

// NLPService 自然语言处理服务接口
type NLPService interface {
	// 基础文本处理
	TokenizeText(ctx context.Context, input TextInput) (*TokenizationResult, error)
	AnalyzeSentiment(ctx context.Context, input TextInput) (*SentimentAnalysisResult, error)
	ExtractEntities(ctx context.Context, input TextInput) (*EntityExtractionResult, error)
	ClassifyText(ctx context.Context, input TextInput) (*TextClassificationResult, error)
	
	// 高级语义理解
	AnalyzeSemantics(ctx context.Context, input TextInput) (*SemanticAnalysisResult, error)
	ExtractKeywords(ctx context.Context, input TextInput) (*KeywordExtractionResult, error)
	SummarizeText(ctx context.Context, input TextInput) (*TextSummarizationResult, error)
	AnalyzeIntent(ctx context.Context, input TextInput) (*IntentAnalysisResult, error)
	
	// 文本生成与转换
	GenerateText(ctx context.Context, input TextGenerationInput) (*TextGenerationResult, error)
	TranslateText(ctx context.Context, input TranslationInput) (*TranslationResult, error)
	ParaphraseText(ctx context.Context, input TextInput) (*ParaphraseResult, error)
	
	// 对话与问答
	ProcessConversation(ctx context.Context, input ConversationInput) (*ConversationResult, error)
	AnswerQuestion(ctx context.Context, input QuestionAnsweringInput) (*QuestionAnsweringResult, error)
	
	// 批量处理
	BatchProcess(ctx context.Context, inputs []TextInput, operations []NLPOperation) (*BatchNLPResult, error)
	
	// 配置与管理
	UpdateConfig(config NLPConfig) error
	GetSupportedLanguages() []Language
	GetSupportedOperations() []OperationType
}

// TextInput 文本输入
type TextInput struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Language Language          `json:"language"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Language 语言类型
type Language string

const (
	LanguageAuto    Language = "auto"
	LanguageEnglish Language = "en"
	LanguageChinese Language = "zh"
	LanguageJapanese Language = "ja"
	LanguageKorean  Language = "ko"
	LanguageFrench  Language = "fr"
	LanguageGerman  Language = "de"
	LanguageSpanish Language = "es"
	LanguageRussian Language = "ru"
	LanguageArabic  Language = "ar"
)

// OperationType 操作类型
type OperationType string

const (
	OpTokenization      OperationType = "tokenization"
	OpSentiment         OperationType = "sentiment"
	OpEntityExtraction  OperationType = "entity_extraction"
	OpClassification    OperationType = "classification"
	OpSemanticAnalysis  OperationType = "semantic_analysis"
	OpKeywordExtraction OperationType = "keyword_extraction"
	OpSummarization     OperationType = "summarization"
	OpIntentAnalysis    OperationType = "intent_analysis"
	OpTextGeneration    OperationType = "text_generation"
	OpTranslation       OperationType = "translation"
	OpParaphrase        OperationType = "paraphrase"
	OpConversation      OperationType = "conversation"
	OpQuestionAnswering OperationType = "question_answering"
)

// TokenizationResult 分词结果
type TokenizationResult struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Tokens       []Token                `json:"tokens"`
	Sentences    []Sentence             `json:"sentences"`
	TotalTokens  int                    `json:"total_tokens"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Token 词元
type Token struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Lemma    string    `json:"lemma"`
	POS      string    `json:"pos"`      // 词性
	Tag      string    `json:"tag"`      // 标签
	Start    int       `json:"start"`    // 起始位置
	End      int       `json:"end"`      // 结束位置
	Features map[string]string `json:"features"`
}

// Sentence 句子
type Sentence struct {
	ID     string  `json:"id"`
	Text   string  `json:"text"`
	Start  int     `json:"start"`
	End    int     `json:"end"`
	Tokens []Token `json:"tokens"`
}

// SentimentAnalysisResult 情感分析结果
type SentimentAnalysisResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	OverallSentiment Sentiment            `json:"overall_sentiment"`
	SentenceSentiments []SentenceSentiment `json:"sentence_sentiments"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Sentiment 情感
type Sentiment struct {
	Label      string  `json:"label"`      // positive, negative, neutral
	Score      float64 `json:"score"`      // -1.0 to 1.0
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
}

// SentenceSentiment 句子情感
type SentenceSentiment struct {
	SentenceID string    `json:"sentence_id"`
	Text       string    `json:"text"`
	Sentiment  Sentiment `json:"sentiment"`
}

// EntityExtractionResult 实体提取结果
type EntityExtractionResult struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Entities     []Entity               `json:"entities"`
	TotalEntities int                   `json:"total_entities"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Entity 实体
type Entity struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Label      string  `json:"label"`      // PERSON, ORG, LOC, MISC, etc.
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float64 `json:"confidence"`
	Properties map[string]interface{} `json:"properties"`
}

// TextClassificationResult 文本分类结果
type TextClassificationResult struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Categories   []Category             `json:"categories"`
	TopCategory  Category               `json:"top_category"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Category 分类
type Category struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
	Properties map[string]interface{} `json:"properties"`
}

// SemanticAnalysisResult 语义分析结果
type SemanticAnalysisResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	SemanticRoles  []SemanticRole         `json:"semantic_roles"`
	Dependencies   []Dependency           `json:"dependencies"`
	Concepts       []Concept              `json:"concepts"`
	Relations      []Relation             `json:"relations"`
	Complexity     float64                `json:"complexity"`
	Coherence      float64                `json:"coherence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// SemanticRole 语义角色
type SemanticRole struct {
	ID       string `json:"id"`
	Predicate string `json:"predicate"`
	Agent    string `json:"agent"`
	Patient  string `json:"patient"`
	Role     string `json:"role"`
	Confidence float64 `json:"confidence"`
}

// Dependency 依存关系
type Dependency struct {
	ID         string `json:"id"`
	Head       string `json:"head"`
	Dependent  string `json:"dependent"`
	Relation   string `json:"relation"`
	Confidence float64 `json:"confidence"`
}

// Concept 概念
type Concept struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Properties map[string]interface{} `json:"properties"`
}

// Relation 关系
type Relation struct {
	ID         string  `json:"id"`
	Subject    string  `json:"subject"`
	Predicate  string  `json:"predicate"`
	Object     string  `json:"object"`
	Confidence float64 `json:"confidence"`
}

// KeywordExtractionResult 关键词提取结果
type KeywordExtractionResult struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Keywords     []Keyword              `json:"keywords"`
	Phrases      []KeyPhrase            `json:"phrases"`
	TotalKeywords int                   `json:"total_keywords"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Keyword 关键词
type Keyword struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	Frequency  int     `json:"frequency"`
	Relevance  float64 `json:"relevance"`
	Properties map[string]interface{} `json:"properties"`
}

// KeyPhrase 关键短语
type KeyPhrase struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Properties map[string]interface{} `json:"properties"`
}

// TextSummarizationResult 文本摘要结果
type TextSummarizationResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Summary        string                 `json:"summary"`
	ExtractedSentences []string           `json:"extracted_sentences"`
	CompressionRatio float64              `json:"compression_ratio"`
	Relevance      float64                `json:"relevance"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// IntentAnalysisResult 意图分析结果
type IntentAnalysisResult struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Intent       Intent                 `json:"intent"`
	Entities     []Entity               `json:"entities"`
	Confidence   float64                `json:"confidence"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Intent 意图
type Intent struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TextGenerationInput 文本生成输入
type TextGenerationInput struct {
	ID           string            `json:"id"`
	Prompt       string            `json:"prompt"`
	MaxLength    int               `json:"max_length"`
	Temperature  float64           `json:"temperature"`
	TopP         float64           `json:"top_p"`
	TopK         int               `json:"top_k"`
	StopSequences []string         `json:"stop_sequences"`
	Language     Language          `json:"language"`
	Style        string            `json:"style"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TextGenerationResult 文本生成结果
type TextGenerationResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	GeneratedText  string                 `json:"generated_text"`
	Alternatives   []string               `json:"alternatives"`
	Quality        float64                `json:"quality"`
	Coherence      float64                `json:"coherence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// TranslationInput 翻译输入
type TranslationInput struct {
	ID           string            `json:"id"`
	Text         string            `json:"text"`
	SourceLang   Language          `json:"source_lang"`
	TargetLang   Language          `json:"target_lang"`
	Domain       string            `json:"domain"`
	Formality    string            `json:"formality"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TranslationResult 翻译结果
type TranslationResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	TranslatedText string                 `json:"translated_text"`
	Alternatives   []string               `json:"alternatives"`
	Confidence     float64                `json:"confidence"`
	Quality        float64                `json:"quality"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ParaphraseResult 改写结果
type ParaphraseResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Paraphrases    []string               `json:"paraphrases"`
	BestParaphrase string                 `json:"best_paraphrase"`
	Similarity     float64                `json:"similarity"`
	Quality        float64                `json:"quality"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ConversationInput 对话输入
type ConversationInput struct {
	ID           string            `json:"id"`
	Message      string            `json:"message"`
	Context      []ConversationTurn `json:"context"`
	UserID       string            `json:"user_id"`
	SessionID    string            `json:"session_id"`
	Language     Language          `json:"language"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ConversationTurn 对话轮次
type ConversationTurn struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`      // user, assistant, system
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ConversationResult 对话结果
type ConversationResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Response       string                 `json:"response"`
	Intent         Intent                 `json:"intent"`
	Entities       []Entity               `json:"entities"`
	Sentiment      Sentiment              `json:"sentiment"`
	Confidence     float64                `json:"confidence"`
	NextActions    []string               `json:"next_actions"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// QuestionAnsweringInput 问答输入
type QuestionAnsweringInput struct {
	ID           string            `json:"id"`
	Question     string            `json:"question"`
	Context      string            `json:"context"`
	Documents    []Document        `json:"documents"`
	Language     Language          `json:"language"`
	AnswerType   string            `json:"answer_type"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Document 文档
type Document struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Source   string            `json:"source"`
	Metadata map[string]interface{} `json:"metadata"`
}

// QuestionAnsweringResult 问答结果
type QuestionAnsweringResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Answer         string                 `json:"answer"`
	Alternatives   []Answer               `json:"alternatives"`
	Confidence     float64                `json:"confidence"`
	Sources        []Source               `json:"sources"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Answer 答案
type Answer struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Source     string  `json:"source"`
}

// Source 来源
type Source struct {
	ID         string  `json:"id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Snippet    string  `json:"snippet"`
	Relevance  float64 `json:"relevance"`
	URL        string  `json:"url"`
}

// NLPOperation NLP操作
type NLPOperation struct {
	ID         string                 `json:"id"`
	Type       OperationType          `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Order      int                    `json:"order"`
}

// BatchNLPResult 批量NLP结果
type BatchNLPResult struct {
	ID               string                 `json:"id"`
	TotalInputs      int                    `json:"total_inputs"`
	ProcessedInputs  int                    `json:"processed_inputs"`
	FailedInputs     int                    `json:"failed_inputs"`
	Results          []interface{}          `json:"results"`
	Errors           []BatchNLPError        `json:"errors"`
	ProcessingTime   time.Duration          `json:"processing_time"`
	Timestamp        time.Time              `json:"timestamp"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// BatchNLPError 批量NLP错误
type BatchNLPError struct {
	InputID string `json:"input_id"`
	Error   string `json:"error"`
}

// NLPConfig NLP配置
type NLPConfig struct {
	DefaultLanguage    Language          `json:"default_language"`
	MaxTextLength      int               `json:"max_text_length"`
	MaxBatchSize       int               `json:"max_batch_size"`
	EnableCache        bool              `json:"enable_cache"`
	CacheExpiry        time.Duration     `json:"cache_expiry"`
	Timeout            time.Duration     `json:"timeout"`
	Providers          map[string]ProviderConfig `json:"providers"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Enabled    bool                   `json:"enabled"`
	Priority   int                    `json:"priority"`
	Config     map[string]interface{} `json:"config"`
}

// NLPProvider NLP提供商接口
type NLPProvider interface {
	// 基础功能
	TokenizeText(ctx context.Context, input TextInput) (*TokenizationResult, error)
	AnalyzeSentiment(ctx context.Context, input TextInput) (*SentimentAnalysisResult, error)
	ExtractEntities(ctx context.Context, input TextInput) (*EntityExtractionResult, error)
	ClassifyText(ctx context.Context, input TextInput) (*TextClassificationResult, error)
	
	// 高级功能
	AnalyzeSemantics(ctx context.Context, input TextInput) (*SemanticAnalysisResult, error)
	ExtractKeywords(ctx context.Context, input TextInput) (*KeywordExtractionResult, error)
	SummarizeText(ctx context.Context, input TextInput) (*TextSummarizationResult, error)
	AnalyzeIntent(ctx context.Context, input TextInput) (*IntentAnalysisResult, error)
	
	// 生成功能
	GenerateText(ctx context.Context, input TextGenerationInput) (*TextGenerationResult, error)
	TranslateText(ctx context.Context, input TranslationInput) (*TranslationResult, error)
	ParaphraseText(ctx context.Context, input TextInput) (*ParaphraseResult, error)
	
	// 对话功能
	ProcessConversation(ctx context.Context, input ConversationInput) (*ConversationResult, error)
	AnswerQuestion(ctx context.Context, input QuestionAnsweringInput) (*QuestionAnsweringResult, error)
	
	// 元数据
	GetSupportedLanguages() []Language
	GetSupportedOperations() []OperationType
	HealthCheck(ctx context.Context) error
}

// 辅助函数

// CreateTextInput 创建文本输入
func CreateTextInput(id, text string, language Language) TextInput {
	return TextInput{
		ID:       id,
		Text:     text,
		Language: language,
		Metadata: make(map[string]interface{}),
	}
}

// CreateNLPOperation 创建NLP操作
func CreateNLPOperation(opType OperationType, parameters map[string]interface{}, order int) NLPOperation {
	return NLPOperation{
		ID:         generateID(),
		Type:       opType,
		Parameters: parameters,
		Order:      order,
	}
}

// generateID 生成ID
func generateID() string {
	// 简单的ID生成，实际实现中可以使用UUID
	return "nlp_" + time.Now().Format("20060102150405")
}