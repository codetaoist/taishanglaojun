package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	
	"github.com/codetaoist/taishanglaojun/core-services/ai-integration/voice"
)

// OpenAIProvider OpenAI语音服务提供�?
type OpenAIProvider struct {
	config     OpenAIConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey      string        `json:"api_key" yaml:"api_key"`
	BaseURL     string        `json:"base_url" yaml:"base_url"`
	Model       string        `json:"model" yaml:"model"`
	TTSModel    string        `json:"tts_model" yaml:"tts_model"`
	STTModel    string        `json:"stt_model" yaml:"stt_model"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	MaxRetries  int           `json:"max_retries" yaml:"max_retries"`
	Temperature float64       `json:"temperature" yaml:"temperature"`
}

// OpenAISTTRequest OpenAI语音识别请求
type OpenAISTTRequest struct {
	Model               string  `json:"model"`
	Language            string  `json:"language,omitempty"`
	Prompt              string  `json:"prompt,omitempty"`
	ResponseFormat      string  `json:"response_format,omitempty"`
	Temperature         float64 `json:"temperature,omitempty"`
	TimestampGranularities []string `json:"timestamp_granularities,omitempty"`
}

// OpenAISTTResponse OpenAI语音识别响应
type OpenAISTTResponse struct {
	Text     string                    `json:"text"`
	Language string                    `json:"language,omitempty"`
	Duration float64                   `json:"duration,omitempty"`
	Words    []OpenAIWordInfo          `json:"words,omitempty"`
	Segments []OpenAISegmentInfo       `json:"segments,omitempty"`
}

// OpenAIWordInfo OpenAI词语信息
type OpenAIWordInfo struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// OpenAISegmentInfo OpenAI片段信息
type OpenAISegmentInfo struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Temperature      float64 `json:"temperature"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}

// OpenAITTSRequest OpenAI语音合成请求
type OpenAITTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

// NewOpenAIProvider 创建OpenAI提供�?
func NewOpenAIProvider(config OpenAIConfig, logger *zap.Logger) *OpenAIProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.STTModel == "" {
		config.STTModel = "whisper-1"
	}
	if config.TTSModel == "" {
		config.TTSModel = "tts-1"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &OpenAIProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// SpeechToText 语音识别
func (p *OpenAIProvider) SpeechToText(ctx context.Context, audio voice.AudioInput) (*voice.SpeechToTextResult, error) {
	// 创建multipart表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加音频文件
	audioWriter, err := writer.CreateFormFile("file", fmt.Sprintf("audio.%s", audio.Format))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	
	if _, err := audioWriter.Write(audio.Data); err != nil {
		return nil, fmt.Errorf("failed to write audio data: %w", err)
	}

	// 添加其他参数
	if err := writer.WriteField("model", p.config.STTModel); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}

	if audio.Language != "" {
		if err := writer.WriteField("language", audio.Language); err != nil {
			return nil, fmt.Errorf("failed to write language field: %w", err)
		}
	}

	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}

	if err := writer.WriteField("timestamp_granularities[]", "word"); err != nil {
		return nil, fmt.Errorf("failed to write timestamp_granularities field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请�?
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var openaiResp OpenAISTTResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换为标准格�?
	result := &voice.SpeechToTextResult{
		ID:        uuid.New().String(),
		RequestID: audio.ID,
		Text:      openaiResp.Text,
		Confidence: 0.95, // OpenAI不提供置信度，使用默认�?
		Language:  openaiResp.Language,
		Words:     make([]voice.WordInfo, len(openaiResp.Words)),
		Segments:  make([]voice.TextSegment, len(openaiResp.Segments)),
		Timestamp: time.Now(),
		IsFinal:   true,
	}

	// 转换词语信息
	for i, word := range openaiResp.Words {
		result.Words[i] = voice.WordInfo{
			Word:       word.Word,
			StartTime:  time.Duration(word.Start * float64(time.Second)),
			EndTime:    time.Duration(word.End * float64(time.Second)),
			Confidence: 0.95,
		}
	}

	// 转换片段信息
	for i, segment := range openaiResp.Segments {
		result.Segments[i] = voice.TextSegment{
			Text:      segment.Text,
			StartTime: time.Duration(segment.Start * float64(time.Second)),
			EndTime:   time.Duration(segment.End * float64(time.Second)),
			Speaker:   "unknown",
		}
	}

	return result, nil
}

// TextToSpeech 语音合成
func (p *OpenAIProvider) TextToSpeech(ctx context.Context, text voice.TextInput) (*voice.TextToSpeechResult, error) {
	// 准备请求数据
	reqData := OpenAITTSRequest{
		Model:          p.config.TTSModel,
		Input:          text.Text,
		Voice:          p.getVoiceName(text.Voice),
		ResponseFormat: string(voice.FormatMP3),
		Speed:          text.Speed,
	}

	if reqData.Speed == 0 {
		reqData.Speed = 1.0
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/audio/speech", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请�?
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 读取音频数据
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	// 创建结果
	result := &voice.TextToSpeechResult{
		ID:         uuid.New().String(),
		RequestID:  text.ID,
		AudioData:  audioData,
		Format:     voice.FormatMP3,
		SampleRate: 24000, // OpenAI TTS默认采样�?
		Channels:   1,
		Voice: voice.Voice{
			ID:       reqData.Voice,
			Name:     reqData.Voice,
			Language: text.Language,
			Gender:   p.getVoiceGender(reqData.Voice),
		},
		Timestamp: time.Now(),
	}

	return result, nil
}

// getVoiceName 获取语音名称
func (p *OpenAIProvider) getVoiceName(voiceID string) string {
	if voiceID == "" {
		return "alloy" // 默认语音
	}
	
	// OpenAI支持的语�?
	supportedVoices := []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"}
	for _, voice := range supportedVoices {
		if voice == voiceID {
			return voiceID
		}
	}
	
	return "alloy"
}

// getVoiceGender 获取语音性别
func (p *OpenAIProvider) getVoiceGender(voiceID string) string {
	genderMap := map[string]string{
		"alloy":   "neutral",
		"echo":    "male",
		"fable":   "female",
		"onyx":    "male",
		"nova":    "female",
		"shimmer": "female",
	}
	
	if gender, exists := genderMap[voiceID]; exists {
		return gender
	}
	
	return "neutral"
}

// GetSupportedLanguages 获取支持的语言
func (p *OpenAIProvider) GetSupportedLanguages() []voice.Language {
	return []voice.Language{
		{Code: "en", Name: "English", NativeName: "English", Region: "US", IsSupported: true},
		{Code: "zh", Name: "Chinese", NativeName: "中文", Region: "CN", IsSupported: true},
		{Code: "es", Name: "Spanish", NativeName: "Español", Region: "ES", IsSupported: true},
		{Code: "fr", Name: "French", NativeName: "Français", Region: "FR", IsSupported: true},
		{Code: "de", Name: "German", NativeName: "Deutsch", Region: "DE", IsSupported: true},
		{Code: "ja", Name: "Japanese", NativeName: "日本�?, Region: "JP", IsSupported: true},
		{Code: "ko", Name: "Korean", NativeName: "한국�?, Region: "KR", IsSupported: true},
		{Code: "pt", Name: "Portuguese", NativeName: "Português", Region: "PT", IsSupported: true},
		{Code: "ru", Name: "Russian", NativeName: "Русский", Region: "RU", IsSupported: true},
		{Code: "ar", Name: "Arabic", NativeName: "العربية", Region: "SA", IsSupported: true},
		{Code: "hi", Name: "Hindi", NativeName: "हिन्दी", Region: "IN", IsSupported: true},
		{Code: "it", Name: "Italian", NativeName: "Italiano", Region: "IT", IsSupported: true},
		{Code: "nl", Name: "Dutch", NativeName: "Nederlands", Region: "NL", IsSupported: true},
		{Code: "pl", Name: "Polish", NativeName: "Polski", Region: "PL", IsSupported: true},
		{Code: "tr", Name: "Turkish", NativeName: "Türkçe", Region: "TR", IsSupported: true},
	}
}

// GetSupportedVoices 获取支持的语�?
func (p *OpenAIProvider) GetSupportedVoices() []voice.Voice {
	return []voice.Voice{
		{
			ID:       "alloy",
			Name:     "Alloy",
			Language: "en",
			Gender:   "neutral",
			Age:      "adult",
			Style:    "natural",
			Quality:  "high",
		},
		{
			ID:       "echo",
			Name:     "Echo",
			Language: "en",
			Gender:   "male",
			Age:      "adult",
			Style:    "natural",
			Quality:  "high",
		},
		{
			ID:       "fable",
			Name:     "Fable",
			Language: "en",
			Gender:   "female",
			Age:      "adult",
			Style:    "natural",
			Quality:  "high",
		},
		{
			ID:       "onyx",
			Name:     "Onyx",
			Language: "en",
			Gender:   "male",
			Age:      "adult",
			Style:    "natural",
			Quality:  "high",
		},
		{
			ID:       "nova",
			Name:     "Nova",
			Language: "en",
			Gender:   "female",
			Age:      "adult",
			Style:    "natural",
			Quality:  "high",
		},
		{
			ID:       "shimmer",
			Name:     "Shimmer",
			Language: "en",
			Gender:   "female",
			Age:      "adult",
			Style:    "natural",
			Quality:  "high",
		},
	}
}

// HealthCheck 健康检�?
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
