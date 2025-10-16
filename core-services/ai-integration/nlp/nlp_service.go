package nlp

import (
	"context"
	"time"
)

// NLPService 
type NLPService interface {
	// 
	TokenizeText(ctx context.Context, input TextInput) (*TokenizationResult, error)
	AnalyzeSentiment(ctx context.Context, input TextInput) (*SentimentAnalysisResult, error)
	ExtractEntities(ctx context.Context, input TextInput) (*EntityExtractionResult, error)
	ClassifyText(ctx context.Context, input TextInput) (*TextClassificationResult, error)

	// 
	AnalyzeSemantics(ctx context.Context, input TextInput) (*SemanticAnalysisResult, error)
	ExtractKeywords(ctx context.Context, input TextInput) (*KeywordExtractionResult, error)
	SummarizeText(ctx context.Context, input TextInput) (*TextSummarizationResult, error)
	AnalyzeIntent(ctx context.Context, input TextInput) (*IntentAnalysisResult, error)

	// 
	GenerateText(ctx context.Context, input TextGenerationInput) (*TextGenerationResult, error)
	TranslateText(ctx context.Context, input TranslationInput) (*TranslationResult, error)
	ParaphraseText(ctx context.Context, input TextInput) (*ParaphraseResult, error)

	// 
	ProcessConversation(ctx context.Context, input ConversationInput) (*ConversationResult, error)
	AnswerQuestion(ctx context.Context, input QuestionAnsweringInput) (*QuestionAnsweringResult, error)

	// 
	BatchProcess(ctx context.Context, inputs []TextInput, operations []NLPOperation) (*BatchNLPResult, error)

	// 
	UpdateConfig(config NLPConfig) error
	GetSupportedLanguages() []Language
	GetSupportedOperations() []OperationType
}

// TextInput 
type TextInput struct {
	ID       string                 `json:"id"`
	Text     string                 `json:"text"`
	Language Language               `json:"language"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Language 
type Language string

const (
	LanguageAuto     Language = "auto"
	LanguageEnglish  Language = "en"
	LanguageChinese  Language = "zh"
	LanguageJapanese Language = "ja"
	LanguageKorean   Language = "ko"
	LanguageFrench   Language = "fr"
	LanguageGerman   Language = "de"
	LanguageSpanish  Language = "es"
	LanguageRussian  Language = "ru"
	LanguageArabic   Language = "ar"
)

// OperationType 
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

// TokenizationResult 
type TokenizationResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Tokens         []Token                `json:"tokens"`
	Sentences      []Sentence             `json:"sentences"`
	TotalTokens    int                    `json:"total_tokens"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Token 
type Token struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Lemma    string            `json:"lemma"`
	POS      string            `json:"pos"`   // 
	Tag      string            `json:"tag"`   // 
	Start    int               `json:"start"` // 
	End      int               `json:"end"`   // 
	Features map[string]string `json:"features"`
}

// Sentence 
type Sentence struct {
	ID     string  `json:"id"`
	Text   string  `json:"text"`
	Start  int     `json:"start"`
	End    int     `json:"end"`
	Tokens []Token `json:"tokens"`
}

// SentimentAnalysisResult 
type SentimentAnalysisResult struct {
	ID                 string                 `json:"id"`
	RequestID          string                 `json:"request_id"`
	OverallSentiment   Sentiment              `json:"overall_sentiment"`
	SentenceSentiments []SentenceSentiment    `json:"sentence_sentiments"`
	Confidence         float64                `json:"confidence"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	Timestamp          time.Time              `json:"timestamp"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// Sentiment 
type Sentiment struct {
	Label      string  `json:"label"`      // positive, negative, neutral
	Score      float64 `json:"score"`      // -1.0 to 1.0
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
}

// SentenceSentiment 
type SentenceSentiment struct {
	SentenceID string    `json:"sentence_id"`
	Text       string    `json:"text"`
	Sentiment  Sentiment `json:"sentiment"`
}

// EntityExtractionResult 
type EntityExtractionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Entities       []Entity               `json:"entities"`
	TotalEntities  int                    `json:"total_entities"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Entity 
type Entity struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text"`
	Label      string                 `json:"label"` // PERSON, ORG, LOC, MISC, etc.
	Start      int                    `json:"start"`
	End        int                    `json:"end"`
	Confidence float64                `json:"confidence"`
	Properties map[string]interface{} `json:"properties"`
}

// TextClassificationResult 
type TextClassificationResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Categories     []Category             `json:"categories"`
	TopCategory    Category               `json:"top_category"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Category 
type Category struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Score      float64                `json:"score"`
	Confidence float64                `json:"confidence"`
	Properties map[string]interface{} `json:"properties"`
}

// SemanticAnalysisResult 
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

// SemanticRole 
type SemanticRole struct {
	ID         string  `json:"id"`
	Predicate  string  `json:"predicate"`
	Agent      string  `json:"agent"`
	Patient    string  `json:"patient"`
	Role       string  `json:"role"`
	Confidence float64 `json:"confidence"`
}

// Dependency 
type Dependency struct {
	ID         string  `json:"id"`
	Head       string  `json:"head"`
	Dependent  string  `json:"dependent"`
	Relation   string  `json:"relation"`
	Confidence float64 `json:"confidence"`
}

// Concept 
type Concept struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text"`
	Type       string                 `json:"type"`
	Confidence float64                `json:"confidence"`
	Properties map[string]interface{} `json:"properties"`
}

// Relation 
type Relation struct {
	ID         string  `json:"id"`
	Subject    string  `json:"subject"`
	Predicate  string  `json:"predicate"`
	Object     string  `json:"object"`
	Confidence float64 `json:"confidence"`
}

// KeywordExtractionResult 
type KeywordExtractionResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Keywords       []Keyword              `json:"keywords"`
	Phrases        []KeyPhrase            `json:"phrases"`
	TotalKeywords  int                    `json:"total_keywords"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Keyword 
type Keyword struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text"`
	Score      float64                `json:"score"`
	Frequency  int                    `json:"frequency"`
	Relevance  float64                `json:"relevance"`
	Properties map[string]interface{} `json:"properties"`
}

// KeyPhrase 
type KeyPhrase struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text"`
	Score      float64                `json:"score"`
	Start      int                    `json:"start"`
	End        int                    `json:"end"`
	Properties map[string]interface{} `json:"properties"`
}

// TextSummarizationResult 
type TextSummarizationResult struct {
	ID                 string                 `json:"id"`
	RequestID          string                 `json:"request_id"`
	Summary            string                 `json:"summary"`
	ExtractedSentences []string               `json:"extracted_sentences"`
	CompressionRatio   float64                `json:"compression_ratio"`
	Relevance          float64                `json:"relevance"`
	ProcessingTime     time.Duration          `json:"processing_time"`
	Timestamp          time.Time              `json:"timestamp"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// IntentAnalysisResult 
type IntentAnalysisResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Intent         Intent                 `json:"intent"`
	Entities       []Entity               `json:"entities"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Intent 
type Intent struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Category   string                 `json:"category"`
	Confidence float64                `json:"confidence"`
	Parameters map[string]interface{} `json:"parameters"`
}

// TextGenerationInput 
type TextGenerationInput struct {
	ID            string                 `json:"id"`
	Prompt        string                 `json:"prompt"`
	MaxLength     int                    `json:"max_length"`
	Temperature   float64                `json:"temperature"`
	TopP          float64                `json:"top_p"`
	TopK          int                    `json:"top_k"`
	StopSequences []string               `json:"stop_sequences"`
	Language      Language               `json:"language"`
	Style         string                 `json:"style"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// TextGenerationResult 
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

// TranslationInput 
type TranslationInput struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text"`
	SourceLang Language               `json:"source_lang"`
	TargetLang Language               `json:"target_lang"`
	Domain     string                 `json:"domain"`
	Formality  string                 `json:"formality"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TranslationResult 
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

// ParaphraseResult 
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

// ConversationInput 
type ConversationInput struct {
	ID        string                 `json:"id"`
	Message   string                 `json:"message"`
	Context   []ConversationTurn     `json:"context"`
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
	Language  Language               `json:"language"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ConversationTurn 
type ConversationTurn struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"` // user, assistant, system
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ConversationResult 
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

// QuestionAnsweringInput 
type QuestionAnsweringInput struct {
	ID         string                 `json:"id"`
	Question   string                 `json:"question"`
	Context    string                 `json:"context"`
	Documents  []Document             `json:"documents"`
	Language   Language               `json:"language"`
	AnswerType string                 `json:"answer_type"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Document 
type Document struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Source   string                 `json:"source"`
	Metadata map[string]interface{} `json:"metadata"`
}

// QuestionAnsweringResult 
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

// Answer 
type Answer struct {
	ID         string  `json:"id"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Source     string  `json:"source"`
}

// Source 
type Source struct {
	ID         string  `json:"id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	Snippet    string  `json:"snippet"`
	Relevance  float64 `json:"relevance"`
	URL        string  `json:"url"`
}

// NLPOperation NLP
type NLPOperation struct {
	ID         string                 `json:"id"`
	Type       OperationType          `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Order      int                    `json:"order"`
}

// BatchNLPResult NLP
type BatchNLPResult struct {
	ID              string                 `json:"id"`
	TotalInputs     int                    `json:"total_inputs"`
	ProcessedInputs int                    `json:"processed_inputs"`
	FailedInputs    int                    `json:"failed_inputs"`
	Results         []interface{}          `json:"results"`
	Errors          []BatchNLPError        `json:"errors"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// BatchNLPError NLP
type BatchNLPError struct {
	InputID string `json:"input_id"`
	Error   string `json:"error"`
}

// NLPConfig NLP
type NLPConfig struct {
	DefaultLanguage Language                  `json:"default_language"`
	MaxTextLength   int                       `json:"max_text_length"`
	MaxBatchSize    int                       `json:"max_batch_size"`
	EnableCache     bool                      `json:"enable_cache"`
	CacheExpiry     time.Duration             `json:"cache_expiry"`
	Timeout         time.Duration             `json:"timeout"`
	Providers       map[string]ProviderConfig `json:"providers"`
}

// ProviderConfig 
type ProviderConfig struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
	Config   map[string]interface{} `json:"config"`
}

// NLPProvider NLP
type NLPProvider interface {
	// 
	TokenizeText(ctx context.Context, input TextInput) (*TokenizationResult, error)
	AnalyzeSentiment(ctx context.Context, input TextInput) (*SentimentAnalysisResult, error)
	ExtractEntities(ctx context.Context, input TextInput) (*EntityExtractionResult, error)
	ClassifyText(ctx context.Context, input TextInput) (*TextClassificationResult, error)

	// 
	AnalyzeSemantics(ctx context.Context, input TextInput) (*SemanticAnalysisResult, error)
	ExtractKeywords(ctx context.Context, input TextInput) (*KeywordExtractionResult, error)
	SummarizeText(ctx context.Context, input TextInput) (*TextSummarizationResult, error)
	AnalyzeIntent(ctx context.Context, input TextInput) (*IntentAnalysisResult, error)

	// 
	GenerateText(ctx context.Context, input TextGenerationInput) (*TextGenerationResult, error)
	TranslateText(ctx context.Context, input TranslationInput) (*TranslationResult, error)
	ParaphraseText(ctx context.Context, input TextInput) (*ParaphraseResult, error)

	// 
	ProcessConversation(ctx context.Context, input ConversationInput) (*ConversationResult, error)
	AnswerQuestion(ctx context.Context, input QuestionAnsweringInput) (*QuestionAnsweringResult, error)

	// 
	GetSupportedLanguages() []Language
	GetSupportedOperations() []OperationType
	HealthCheck(ctx context.Context) error
}

// 

// CreateTextInput 
func CreateTextInput(id, text string, language Language) TextInput {
	return TextInput{
		ID:       id,
		Text:     text,
		Language: language,
		Metadata: make(map[string]interface{}),
	}
}

// CreateNLPOperation NLP
func CreateNLPOperation(opType OperationType, parameters map[string]interface{}, order int) NLPOperation {
	return NLPOperation{
		ID:         generateID(),
		Type:       opType,
		Parameters: parameters,
		Order:      order,
	}
}

// generateID ID
func generateID() string {
	// IDUUID
	return "nlp_" + time.Now().Format("20060102150405")
}

