﻿package voice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VoiceService 
type VoiceService interface {
	// 
	SpeechToText(ctx context.Context, audio AudioInput) (*SpeechToTextResult, error)
	StreamSpeechToText(ctx context.Context, audioStream <-chan AudioChunk) (<-chan *SpeechToTextResult, error)
	
	// 
	TextToSpeech(ctx context.Context, text TextInput) (*TextToSpeechResult, error)
	StreamTextToSpeech(ctx context.Context, textStream <-chan string) (<-chan AudioChunk, error)
	
	// 
	StartVoiceConversation(ctx context.Context, config ConversationConfig) (*VoiceConversation, error)
	
	// 
	AnalyzeVoice(ctx context.Context, audio AudioInput) (*VoiceAnalysisResult, error)
	
	// 
	UpdateConfig(config VoiceConfig) error
	GetSupportedLanguages() []Language
	GetSupportedVoices() []Voice
}

// AudioFormat 
type AudioFormat string

const (
	FormatWAV  AudioFormat = "wav"
	FormatMP3  AudioFormat = "mp3"
	FormatFLAC AudioFormat = "flac"
	FormatOGG  AudioFormat = "ogg"
	FormatAAC  AudioFormat = "aac"
	FormatPCM  AudioFormat = "pcm"
)

// AudioInput 
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

// AudioChunk ?
type AudioChunk struct {
	ID        string      `json:"id"`
	Data      []byte      `json:"data"`
	Sequence  int         `json:"sequence"`
	IsLast    bool        `json:"is_last"`
	Timestamp time.Time   `json:"timestamp"`
}

// TextInput 
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

// SpeechToTextResult 
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

// TextAlternative ?
type TextAlternative struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// WordInfo 
type WordInfo struct {
	Word       string        `json:"word"`
	StartTime  time.Duration `json:"start_time"`
	EndTime    time.Duration `json:"end_time"`
	Confidence float64       `json:"confidence"`
}

// TextSegment 
type TextSegment struct {
	Text      string        `json:"text"`
	StartTime time.Duration `json:"start_time"`
	EndTime   time.Duration `json:"end_time"`
	Speaker   string        `json:"speaker"`
}

// TextToSpeechResult 
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

// VoiceAnalysisResult 
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

// EmotionAnalysis 
type EmotionAnalysis struct {
	Primary    string             `json:"primary"`
	Confidence float64            `json:"confidence"`
	Emotions   map[string]float64 `json:"emotions"`
}

// SpeakerInfo ?
type SpeakerInfo struct {
	ID         string  `json:"id"`
	Gender     string  `json:"gender"`
	Age        string  `json:"age"`
	Accent     string  `json:"accent"`
	Confidence float64 `json:"confidence"`
}

// AudioQuality 
type AudioQuality struct {
	Score       float64 `json:"score"`
	NoiseLevel  float64 `json:"noise_level"`
	Clarity     float64 `json:"clarity"`
	Volume      float64 `json:"volume"`
	Distortion  float64 `json:"distortion"`
}

// LanguageDetection ?
type LanguageDetection struct {
	Primary    string             `json:"primary"`
	Confidence float64            `json:"confidence"`
	Languages  map[string]float64 `json:"languages"`
}

// SentimentAnalysis 
type SentimentAnalysis struct {
	Polarity   string  `json:"polarity"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// Language 
type Language struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	NativeName  string `json:"native_name"`
	Region      string `json:"region"`
	IsSupported bool   `json:"is_supported"`
}

// Voice 
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

// VoiceConfig 
type VoiceConfig struct {
	// 
	STT STTConfig `json:"stt" yaml:"stt"`
	
	// 
	TTS TTSConfig `json:"tts" yaml:"tts"`
	
	// 
	DefaultLanguage string        `json:"default_language" yaml:"default_language"`
	MaxAudioSize    int64         `json:"max_audio_size" yaml:"max_audio_size"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	RetryCount      int           `json:"retry_count" yaml:"retry_count"`
	
	// ?
	Providers map[string]ProviderConfig `json:"providers" yaml:"providers"`
}

// STTConfig 
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

// TTSConfig 
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

// ProviderConfig ?
type ProviderConfig struct {
	Name     string                 `json:"name" yaml:"name"`
	Endpoint string                 `json:"endpoint" yaml:"endpoint"`
	APIKey   string                 `json:"api_key" yaml:"api_key"`
	Region   string                 `json:"region" yaml:"region"`
	Params   map[string]interface{} `json:"params" yaml:"params"`
}

// ConversationConfig 
type ConversationConfig struct {
	ID              string        `json:"id"`
	Language        string        `json:"language"`
	Voice           string        `json:"voice"`
	EnableVAD       bool          `json:"enable_vad"`       // ?
	EnableNLP       bool          `json:"enable_nlp"`       // 
	EnableEmotion   bool          `json:"enable_emotion"`   // 
	StreamingMode   bool          `json:"streaming_mode"`   // 
	AutoResponse    bool          `json:"auto_response"`    // 
	MaxDuration     time.Duration `json:"max_duration"`     // ?
	SilenceTimeout  time.Duration `json:"silence_timeout"`  // 
	UserID          string        `json:"user_id"`
	SessionID       string        `json:"session_id"`
}

// VoiceConversation 
type VoiceConversation struct {
	ID            string                 `json:"id"`
	Config        ConversationConfig     `json:"config"`
	Status        ConversationStatus     `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration"`
	Messages      []ConversationMessage  `json:"messages"`
	Metadata      map[string]interface{} `json:"metadata"`
	
	// 
	AudioInput    chan AudioChunk        `json:"-"`
	AudioOutput   chan AudioChunk        `json:"-"`
	TextInput     chan string            `json:"-"`
	TextOutput    chan string            `json:"-"`
	Control       chan ConversationControl `json:"-"`
	Events        chan ConversationEvent `json:"-"`
}

// ConversationStatus ?
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

// ConversationMessage 
type ConversationMessage struct {
	ID        string                 `json:"id"`
	Type      MessageType            `json:"type"`
	Content   string                 `json:"content"`
	Speaker   string                 `json:"speaker"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MessageType 
type MessageType string

const (
	MessageTypeUser      MessageType = "user"
	MessageTypeAssistant MessageType = "assistant"
	MessageTypeSystem    MessageType = "system"
)

// ConversationControl 
type ConversationControl struct {
	Action    ControlAction `json:"action"`
	Timestamp time.Time     `json:"timestamp"`
}

// ControlAction 
type ControlAction string

const (
	ActionStart  ControlAction = "start"
	ActionPause  ControlAction = "pause"
	ActionResume ControlAction = "resume"
	ActionStop   ControlAction = "stop"
	ActionMute   ControlAction = "mute"
	ActionUnmute ControlAction = "unmute"
)

// ConversationEvent 
type ConversationEvent struct {
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventType 
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

// DefaultVoiceService 
type DefaultVoiceService struct {
	config    VoiceConfig
	providers map[string]VoiceProvider
	logger    *zap.Logger
}

// VoiceProvider ?
type VoiceProvider interface {
	SpeechToText(ctx context.Context, audio AudioInput) (*SpeechToTextResult, error)
	TextToSpeech(ctx context.Context, text TextInput) (*TextToSpeechResult, error)
	GetSupportedLanguages() []Language
	GetSupportedVoices() []Voice
	HealthCheck(ctx context.Context) error
}

// NewDefaultVoiceService 
func NewDefaultVoiceService(config VoiceConfig, logger *zap.Logger) *DefaultVoiceService {
	return &DefaultVoiceService{
		config:    config,
		providers: make(map[string]VoiceProvider),
		logger:    logger,
	}
}

// RegisterProvider ?
func (s *DefaultVoiceService) RegisterProvider(name string, provider VoiceProvider) {
	s.providers[name] = provider
}

// SpeechToText 
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

// TextToSpeech 
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

// StreamSpeechToText 
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
					// 
					if len(audioBuffer) > 0 {
						s.processAudioBuffer(ctx, audioBuffer, sequence, true, resultChan)
					}
					return
				}

				audioBuffer = append(audioBuffer, chunk.Data...)
				sequence = chunk.Sequence

				// 㹻
				if chunk.IsLast || len(audioBuffer) >= 16000 { // 1
					s.processAudioBuffer(ctx, audioBuffer, sequence, chunk.IsLast, resultChan)
					audioBuffer = nil
				}
			}
		}
	}()

	return resultChan, nil
}

// processAudioBuffer ?
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

// getSTTProvider ?
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

// getTTSProvider ?
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

// UpdateConfig 
func (s *DefaultVoiceService) UpdateConfig(config VoiceConfig) error {
	s.config = config
	return nil
}

// GetSupportedLanguages 
func (s *DefaultVoiceService) GetSupportedLanguages() []Language {
	var languages []Language
	for _, provider := range s.providers {
		languages = append(languages, provider.GetSupportedLanguages()...)
	}
	return languages
}

// GetSupportedVoices ?
func (s *DefaultVoiceService) GetSupportedVoices() []Voice {
	var voices []Voice
	for _, provider := range s.providers {
		voices = append(voices, provider.GetSupportedVoices()...)
	}
	return voices
}

