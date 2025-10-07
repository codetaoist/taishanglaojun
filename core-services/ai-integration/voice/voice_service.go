package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VoiceService 语音服务接口
type VoiceService interface {
	// 语音识别
	SpeechToText(ctx context.Context, audio AudioInput) (*SpeechToTextResult, error)
	StreamSpeechToText(ctx context.Context, audioStream <-chan AudioChunk) (<-chan *SpeechToTextResult, error)
	
	// 语音合成
	TextToSpeech(ctx context.Context, text TextInput) (*TextToSpeechResult, error)
	StreamTextToSpeech(ctx context.Context, textStream <-chan string) (<-chan AudioChunk, error)
	
	// 语音对话
	StartVoiceConversation(ctx context.Context, config ConversationConfig) (*VoiceConversation, error)
	
	// 语音分析
	AnalyzeVoice(ctx context.Context, audio AudioInput) (*VoiceAnalysisResult, error)
	
	// 配置管理
	UpdateConfig(config VoiceConfig) error
	GetSupportedLanguages() []Language
	GetSupportedVoices() []Voice
}

// AudioFormat 音频格式
type AudioFormat string

const (
	FormatWAV  AudioFormat = "wav"
	FormatMP3  AudioFormat = "mp3"
	FormatFLAC AudioFormat = "flac"
	FormatOGG  AudioFormat = "ogg"
	FormatAAC  AudioFormat = "aac"
	FormatPCM  AudioFormat = "pcm"
)

// AudioInput 音频输入
type AudioInput struct {
	ID          string                 `json:"id"`
	Data        []byte                 `json:"data"`
	Format      AudioFormat            `json:"format"`
	SampleRate  int                    `json:"sample_rate"`
	Channels    int                    `json:"channels"`
	BitDepth    int                    `json:"bit_depth"`
	Duration    time.Duration          `json:"duration"`
	Language    string                 `json:"language"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
}

// AudioChunk 音频块
type AudioChunk struct {
	ID        string      `json:"id"`
	Data      []byte      `json:"data"`
	Sequence  int         `json:"sequence"`
	IsLast    bool        `json:"is_last"`
	Timestamp time.Time   `json:"timestamp"`
}

// TextInput 文本输入
type TextInput struct {
	ID        string                 `json:"id"`
	Text      string                 `json:"text"`
	Language  string                 `json:"language"`
	Voice     string                 `json:"voice"`
	Speed     float64                `json:"speed"`
	Pitch     float64                `json:"pitch"`
	Volume    float64                `json:"volume"`
	Emotion   string                 `json:"emotion"`
	Style     string                 `json:"style"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id"`
	SessionID string                 `json:"session_id"`
}

// SpeechToTextResult 语音识别结果
type SpeechToTextResult struct {
	ID           string                 `json:"id"`
	RequestID    string                 `json:"request_id"`
	Text         string                 `json:"text"`
	Confidence   float64                `json:"confidence"`
	Language     string                 `json:"language"`
	Alternatives []TextAlternative      `json:"alternatives"`
	Words        []WordInfo             `json:"words"`
	Segments     []TextSegment          `json:"segments"`
	Metadata     map[string]interface{} `json:"metadata"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Timestamp    time.Time              `json:"timestamp"`
	IsPartial    bool                   `json:"is_partial"`
	IsFinal      bool                   `json:"is_final"`
}

// TextAlternative 文本候选
type TextAlternative struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// WordInfo 词语信息
type WordInfo struct {
	Word       string        `json:"word"`
	StartTime  time.Duration `json:"start_time"`
	EndTime    time.Duration `json:"end_time"`
	Confidence float64       `json:"confidence"`
}

// TextSegment 文本片段
type TextSegment struct {
	Text      string        `json:"text"`
	StartTime time.Duration `json:"start_time"`
	EndTime   time.Duration `json:"end_time"`
	Speaker   string        `json:"speaker"`
}

// TextToSpeechResult 语音合成结果
type TextToSpeechResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	AudioData      []byte                 `json:"audio_data"`
	Format         AudioFormat            `json:"format"`
	SampleRate     int                    `json:"sample_rate"`
	Channels       int                    `json:"channels"`
	Duration       time.Duration          `json:"duration"`
	Voice          Voice                  `json:"voice"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
}

// VoiceAnalysisResult 语音分析结果
type VoiceAnalysisResult struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	Emotion        EmotionAnalysis        `json:"emotion"`
	Speaker        SpeakerInfo            `json:"speaker"`
	Quality        AudioQuality           `json:"quality"`
	Language       LanguageDetection      `json:"language"`
	Sentiment      SentimentAnalysis      `json:"sentiment"`
	Keywords       []string               `json:"keywords"`
	Topics         []string               `json:"topics"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
}

// EmotionAnalysis 情感分析
type EmotionAnalysis struct {
	Primary    string             `json:"primary"`
	Confidence float64            `json:"confidence"`
	Emotions   map[string]float64 `json:"emotions"`
}

// SpeakerInfo 说话人信息
type SpeakerInfo struct {
	ID         string  `json:"id"`
	Gender     string  `json:"gender"`
	Age        string  `json:"age"`
	Accent     string  `json:"accent"`
	Confidence float64 `json:"confidence"`
}

// AudioQuality 音频质量
type AudioQuality struct {
	Score       float64 `json:"score"`
	NoiseLevel  float64 `json:"noise_level"`
	Clarity     float64 `json:"clarity"`
	Volume      float64 `json:"volume"`
	Distortion  float64 `json:"distortion"`
}

// LanguageDetection 语言检测
type LanguageDetection struct {
	Primary    string             `json:"primary"`
	Confidence float64            `json:"confidence"`
	Languages  map[string]float64 `json:"languages"`
}

// SentimentAnalysis 情感分析
type SentimentAnalysis struct {
	Polarity   string  `json:"polarity"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// Language 语言
type Language struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	NativeName  string `json:"native_name"`
	Region      string `json:"region"`
	IsSupported bool   `json:"is_supported"`
}

// Voice 语音
type Voice struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Language string   `json:"language"`
	Gender   string   `json:"gender"`
	Age      string   `json:"age"`
	Style    string   `json:"style"`
	Quality  string   `json:"quality"`
	Samples  []string `json:"samples"`
}

// VoiceConfig 语音配置
type VoiceConfig struct {
	// 语音识别配置
	STT STTConfig `json:"stt" yaml:"stt"`
	
	// 语音合成配置
	TTS TTSConfig `json:"tts" yaml:"tts"`
	
	// 通用配置
	DefaultLanguage string        `json:"default_language" yaml:"default_language"`
	MaxAudioSize    int64         `json:"max_audio_size" yaml:"max_audio_size"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryCount      int           `json:"retry_count" yaml:"retry_count"`
	
	// 提供商配置
	Providers map[string]ProviderConfig `json:"providers" yaml:"providers"`
}

// STTConfig 语音识别配置
type STTConfig struct {
	Provider       string        `json:"provider" yaml:"provider"`
	Model          string        `json:"model" yaml:"model"`
	Language       string        `json:"language" yaml:"language"`
	EnableProfanityFilter bool   `json:"enable_profanity_filter" yaml:"enable_profanity_filter"`
	EnableWordTimestamps  bool   `json:"enable_word_timestamps" yaml:"enable_word_timestamps"`
	EnableSpeakerDiarization bool `json:"enable_speaker_diarization" yaml:"enable_speaker_diarization"`
	MaxAlternatives int          `json:"max_alternatives" yaml:"max_alternatives"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
}

// TTSConfig 语音合成配置
type TTSConfig struct {
	Provider    string        `json:"provider" yaml:"provider"`
	Voice       string        `json:"voice" yaml:"voice"`
	Language    string        `json:"language" yaml:"language"`
	Speed       float64       `json:"speed" yaml:"speed"`
	Pitch       float64       `json:"pitch" yaml:"pitch"`
	Volume      float64       `json:"volume" yaml:"volume"`
	Format      AudioFormat   `json:"format" yaml:"format"`
	SampleRate  int           `json:"sample_rate" yaml:"sample_rate"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name     string                 `json:"name" yaml:"name"`
	Endpoint string                 `json:"endpoint" yaml:"endpoint"`
	APIKey   string                 `json:"api_key" yaml:"api_key"`
	Region   string                 `json:"region" yaml:"region"`
	Params   map[string]interface{} `json:"params" yaml:"params"`
}

// ConversationConfig 对话配置
type ConversationConfig struct {
	ID              string        `json:"id"`
	Language        string        `json:"language"`
	Voice           string        `json:"voice"`
	EnableVAD       bool          `json:"enable_vad"`       // 语音活动检测
	EnableNLP       bool          `json:"enable_nlp"`       // 自然语言处理
	EnableEmotion   bool          `json:"enable_emotion"`   // 情感分析
	StreamingMode   bool          `json:"streaming_mode"`   // 流式模式
	AutoResponse    bool          `json:"auto_response"`    // 自动回复
	MaxDuration     time.Duration `json:"max_duration"`     // 最大对话时长
	SilenceTimeout  time.Duration `json:"silence_timeout"`  // 静音超时
	UserID          string        `json:"user_id"`
	SessionID       string        `json:"session_id"`
}

// VoiceConversation 语音对话
type VoiceConversation struct {
	ID            string                 `json:"id"`
	Config        ConversationConfig     `json:"config"`
	Status        ConversationStatus     `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration"`
	Messages      []ConversationMessage  `json:"messages"`
	Metadata      map[string]interface{} `json:"metadata"`
	
	// 控制通道
	AudioInput    chan AudioChunk        `json:"-"`
	AudioOutput   chan AudioChunk        `json:"-"`
	TextInput     chan string            `json:"-"`
	TextOutput    chan string            `json:"-"`
	Control       chan ConversationControl `json:"-"`
	Events        chan ConversationEvent `json:"-"`
}

// ConversationStatus 对话状态
type ConversationStatus string

const (
	StatusIdle       ConversationStatus = "idle"
	StatusListening  ConversationStatus = "listening"
	StatusProcessing ConversationStatus = "processing"
	StatusSpeaking   ConversationStatus = "speaking"
	StatusPaused     ConversationStatus = "paused"
	StatusEnded      ConversationStatus = "ended"
	StatusError      ConversationStatus = "error"
)

// ConversationMessage 对话消息
type ConversationMessage struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Content   string                 `json:"content"`
	Speaker   string                 `json:"speaker"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MessageType 消息类型
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeSystem    MessageType = "system"
)

// ConversationControl 对话控制
type ConversationControl struct {
	Action    ControlAction `json:"action"`
	Timestamp time.Time     `json:"timestamp"`
}

// ControlAction 控制动作
type ControlAction string

const (
	ActionStart  ControlAction = "start"
	ActionPause  ControlAction = "pause"
	ActionResume ControlAction = "resume"
	ActionStop   ControlAction = "stop"
	ActionMute   ControlAction = "mute"
	ActionUnmute ControlAction = "unmute"
)

// ConversationEvent 对话事件
type ConversationEvent struct {
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventType 事件类型
type EventType string

const (
	EventSpeechStart    EventType = "speech_start"
	EventSpeechEnd      EventType = "speech_end"
	EventTextReceived   EventType = "text_received"
	EventResponseStart  EventType = "response_start"
	EventResponseEnd    EventType = "response_end"
	EventError          EventType = "error"
	EventStatusChange   EventType = "status_change"
)

// DefaultVoiceService 默认语音服务实现
type DefaultVoiceService struct {
	config    VoiceConfig
	providers map[string]VoiceProvider
	logger    *zap.Logger
}

// VoiceProvider 语音提供商接口
type VoiceProvider interface {
	SpeechToText(ctx context.Context, audio AudioInput) (*SpeechToTextResult, error)
	TextToSpeech(ctx context.Context, text TextInput) (*TextToSpeechResult, error)
	GetSupportedLanguages() []Language
	GetSupportedVoices() []Voice
	HealthCheck(ctx context.Context) error
}

// NewDefaultVoiceService 创建默认语音服务
func NewDefaultVoiceService(config VoiceConfig, logger *zap.Logger) *DefaultVoiceService {
	return &DefaultVoiceService{
		config:    config,
		providers: make(map[string]VoiceProvider),
		logger:    logger,
	}
}

// RegisterProvider 注册语音提供商
func (s *DefaultVoiceService) RegisterProvider(name string, provider VoiceProvider) {
	s.providers[name] = provider
}

// SpeechToText 语音识别
func (s *DefaultVoiceService) SpeechToText(ctx context.Context, audio AudioInput) (*SpeechToTextResult, error) {
	provider, err := s.getSTTProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.SpeechToText(ctx, audio)
	if err != nil {
		s.logger.Error("Speech to text failed", 
			zap.String("audio_id", audio.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Speech to text completed",
		zap.String("audio_id", audio.ID),
		zap.String("text", result.Text),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// TextToSpeech 语音合成
func (s *DefaultVoiceService) TextToSpeech(ctx context.Context, text TextInput) (*TextToSpeechResult, error) {
	provider, err := s.getTTSProvider()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	result, err := provider.TextToSpeech(ctx, text)
	if err != nil {
		s.logger.Error("Text to speech failed",
			zap.String("text_id", text.ID),
			zap.Error(err))
		return nil, err
	}

	result.ProcessingTime = time.Since(startTime)
	result.Timestamp = time.Now()

	s.logger.Info("Text to speech completed",
		zap.String("text_id", text.ID),
		zap.String("text", text.Text),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// StreamSpeechToText 流式语音识别
func (s *DefaultVoiceService) StreamSpeechToText(ctx context.Context, audioStream <-chan AudioChunk) (<-chan *SpeechToTextResult, error) {
	resultChan := make(chan *SpeechToTextResult, 100)

	go func() {
		defer close(resultChan)

		var audioBuffer []byte
		var sequence int

		for {
			select {
			case <-ctx.Done():
				return
			case chunk, ok := <-audioStream:
				if !ok {
					// 处理最后的音频数据
					if len(audioBuffer) > 0 {
						s.processAudioBuffer(ctx, audioBuffer, sequence, true, resultChan)
					}
					return
				}

				audioBuffer = append(audioBuffer, chunk.Data...)
				sequence = chunk.Sequence

				// 如果是最后一块或缓冲区足够大，则处理
				if chunk.IsLast || len(audioBuffer) >= 16000 { // 1秒的音频数据
					s.processAudioBuffer(ctx, audioBuffer, sequence, chunk.IsLast, resultChan)
					audioBuffer = nil
				}
			}
		}
	}()

	return resultChan, nil
}

// processAudioBuffer 处理音频缓冲区
func (s *DefaultVoiceService) processAudioBuffer(ctx context.Context, buffer []byte, sequence int, isFinal bool, resultChan chan<- *SpeechToTextResult) {
	audio := AudioInput{
		ID:        uuid.New().String(),
		Data:      buffer,
		Format:    FormatPCM,
		SampleRate: 16000,
		Channels:  1,
		Timestamp: time.Now(),
	}

	result, err := s.SpeechToText(ctx, audio)
	if err != nil {
		s.logger.Error("Stream speech to text failed", zap.Error(err))
		return
	}

	result.IsPartial = !isFinal
	result.IsFinal = isFinal

	select {
	case resultChan <- result:
	case <-ctx.Done():
		return
	}
}

// getSTTProvider 获取语音识别提供商
func (s *DefaultVoiceService) getSTTProvider() (VoiceProvider, error) {
	providerName := s.config.STT.Provider
	if providerName == "" {
		return nil, fmt.Errorf("no STT provider configured")
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("STT provider %s not found", providerName)
	}

	return provider, nil
}

// getTTSProvider 获取语音合成提供商
func (s *DefaultVoiceService) getTTSProvider() (VoiceProvider, error) {
	providerName := s.config.TTS.Provider
	if providerName == "" {
		return nil, fmt.Errorf("no TTS provider configured")
	}

	provider, exists := s.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("TTS provider %s not found", providerName)
	}

	return provider, nil
}

// UpdateConfig 更新配置
func (s *DefaultVoiceService) UpdateConfig(config VoiceConfig) error {
	s.config = config
	return nil
}

// GetSupportedLanguages 获取支持的语言
func (s *DefaultVoiceService) GetSupportedLanguages() []Language {
	var languages []Language
	for _, provider := range s.providers {
		languages = append(languages, provider.GetSupportedLanguages()...)
	}
	return languages
}

// GetSupportedVoices 获取支持的语音
func (s *DefaultVoiceService) GetSupportedVoices() []Voice {
	var voices []Voice
	for _, provider := range s.providers {
		voices = append(voices, provider.GetSupportedVoices()...)
	}
	return voices
}