package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConversationManager 
type ConversationManager struct {
	voiceService  VoiceService
	nlpService    NLPService
	conversations map[string]*VoiceConversation
	mutex         sync.RWMutex
	logger        *zap.Logger

	// 
	defaultConfig   ConversationConfig
	maxConcurrent   int
	cleanupInterval time.Duration
}

// NLPService 
type NLPService interface {
	ProcessText(ctx context.Context, text string, userID string) (string, error)
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error)
	ExtractKeywords(ctx context.Context, text string) ([]string, error)
	DetectIntent(ctx context.Context, text string) (string, error)
}

// NewConversationManager 
func NewConversationManager(voiceService VoiceService, nlpService NLPService, logger *zap.Logger) *ConversationManager {
	manager := &ConversationManager{
		voiceService:    voiceService,
		nlpService:      nlpService,
		conversations:   make(map[string]*VoiceConversation),
		logger:          logger,
		maxConcurrent:   100,
		cleanupInterval: 5 * time.Minute,
		defaultConfig: ConversationConfig{
			Language:       "zh",
			Voice:          "alloy",
			EnableVAD:      true,
			EnableNLP:      true,
			EnableEmotion:  true,
			StreamingMode:  true,
			AutoResponse:   true,
			MaxDuration:    30 * time.Minute,
			SilenceTimeout: 10 * time.Second,
		},
	}

	// 
	go manager.cleanupRoutine()

	return manager
}

// StartVoiceConversation 
func (m *ConversationManager) StartVoiceConversation(ctx context.Context, config ConversationConfig) (*VoiceConversation, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 鲢
	if len(m.conversations) >= m.maxConcurrent {
		return nil, fmt.Errorf("maximum concurrent conversations reached: %d", m.maxConcurrent)
	}

	// 
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	if config.Language == "" {
		config.Language = m.defaultConfig.Language
	}
	if config.Voice == "" {
		config.Voice = m.defaultConfig.Voice
	}
	if config.MaxDuration == 0 {
		config.MaxDuration = m.defaultConfig.MaxDuration
	}
	if config.SilenceTimeout == 0 {
		config.SilenceTimeout = m.defaultConfig.SilenceTimeout
	}

	// 
	conversation := &VoiceConversation{
		ID:          config.ID,
		Config:      config,
		Status:      StatusIdle,
		StartTime:   time.Now(),
		Messages:    make([]ConversationMessage, 0),
		Metadata:    make(map[string]interface{}),
		AudioInput:  make(chan AudioChunk, 100),
		AudioOutput: make(chan AudioChunk, 100),
		TextInput:   make(chan string, 100),
		TextOutput:  make(chan string, 100),
		Control:     make(chan ConversationControl, 10),
		Events:      make(chan ConversationEvent, 100),
	}

	// 洢
	m.conversations[config.ID] = conversation

	// 
	go m.handleConversation(ctx, conversation)

	m.logger.Info("Voice conversation started",
		zap.String("conversation_id", config.ID),
		zap.String("user_id", config.UserID),
		zap.String("session_id", config.SessionID))

	return conversation, nil
}

// handleConversation 
func (m *ConversationManager) handleConversation(ctx context.Context, conv *VoiceConversation) {
	defer m.cleanupConversation(conv.ID)

	// 
	convCtx, cancel := context.WithTimeout(ctx, conv.Config.MaxDuration)
	defer cancel()

	// 
	m.sendEvent(conv, EventType("conversation_started"), map[string]interface{}{
		"conversation_id": conv.ID,
		"start_time":      conv.StartTime,
	})

	// 
	var wg sync.WaitGroup

	// 
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.handleAudioInput(convCtx, conv)
	}()

	// 
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.handleTextInput(convCtx, conv)
	}()

	// 
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.handleControlCommands(convCtx, conv)
	}()

	// 
	if conv.Config.SilenceTimeout > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.handleSilenceDetection(convCtx, conv)
		}()
	}

	// 
	wg.Wait()

	// 
	conv.Status = StatusEnded
	endTime := time.Now()
	conv.EndTime = &endTime
	conv.Duration = endTime.Sub(conv.StartTime)

	// 
	m.sendEvent(conv, EventType("conversation_ended"), map[string]interface{}{
		"conversation_id": conv.ID,
		"end_time":        endTime,
		"duration":        conv.Duration,
	})

	m.logger.Info("Voice conversation ended",
		zap.String("conversation_id", conv.ID),
		zap.Duration("duration", conv.Duration))
}

// handleAudioInput 
func (m *ConversationManager) handleAudioInput(ctx context.Context, conv *VoiceConversation) {
	var audioBuffer []byte

	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-conv.AudioInput:
			if !ok {
				return
			}

			audioBuffer = append(audioBuffer, chunk.Data...)

			// 
			if conv.Config.EnableVAD && m.detectVoiceActivity(chunk.Data) {
				conv.Status = StatusListening
				m.sendEvent(conv, EventSpeechStart, map[string]interface{}{
					"timestamp": time.Now(),
				})
			}

			// 㹻
			if chunk.IsLast || len(audioBuffer) >= 32000 { // 2
				if len(audioBuffer) > 0 {
					m.processAudioChunk(ctx, conv, audioBuffer)
					audioBuffer = nil
				}
			}
		}
	}
}

// processAudioChunk 
func (m *ConversationManager) processAudioChunk(ctx context.Context, conv *VoiceConversation, audioData []byte) {
	conv.Status = StatusProcessing

	// 
	audio := AudioInput{
		ID:         uuid.New().String(),
		Data:       audioData,
		Format:     FormatPCM,
		SampleRate: 16000,
		Channels:   1,
		Language:   conv.Config.Language,
		UserID:     conv.Config.UserID,
		SessionID:  conv.Config.SessionID,
		Timestamp:  time.Now(),
	}

	// 
	result, err := m.voiceService.SpeechToText(ctx, audio)
	if err != nil {
		m.logger.Error("Speech to text failed",
			zap.String("conversation_id", conv.ID),
			zap.Error(err))
		m.sendEvent(conv, EventError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if result.Text == "" {
		return
	}

	// 
	userMessage := ConversationMessage{
		ID:        uuid.New().String(),
		Type:      MessageTypeUser,
		Content:   result.Text,
		Speaker:   "user",
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"confidence": result.Confidence,
			"language":   result.Language,
		},
	}
	conv.Messages = append(conv.Messages, userMessage)

	// 
	m.sendEvent(conv, EventTextReceived, map[string]interface{}{
		"text":       result.Text,
		"confidence": result.Confidence,
	})

	// 
	if conv.Config.AutoResponse {
		select {
		case conv.TextInput <- result.Text:
		case <-ctx.Done():
			return
		}
	}
}

// handleTextInput 
func (m *ConversationManager) handleTextInput(ctx context.Context, conv *VoiceConversation) {
	for {
		select {
		case <-ctx.Done():
			return
		case text, ok := <-conv.TextInput:
			if !ok {
				return
			}

			m.processTextInput(ctx, conv, text)
		}
	}
}

// processTextInput 
func (m *ConversationManager) processTextInput(ctx context.Context, conv *VoiceConversation, text string) {
	conv.Status = StatusProcessing

	// 
	m.sendEvent(conv, EventResponseStart, map[string]interface{}{
		"input_text": text,
	})

	var response string
	var err error

	// NLPNLP
	if conv.Config.EnableNLP && m.nlpService != nil {
		response, err = m.nlpService.ProcessText(ctx, text, conv.Config.UserID)
		if err != nil {
			m.logger.Error("NLP processing failed",
				zap.String("conversation_id", conv.ID),
				zap.Error(err))
			response = ""
		}
	} else {
		// 
		response = m.generateSimpleResponse(text)
	}

	// 
	assistantMessage := ConversationMessage{
		ID:        uuid.New().String(),
		Type:      MessageTypeAssistant,
		Content:   response,
		Speaker:   "assistant",
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, assistantMessage)

	// 
	textInput := TextInput{
		ID:        uuid.New().String(),
		Text:      response,
		Language:  conv.Config.Language,
		Voice:     conv.Config.Voice,
		UserID:    conv.Config.UserID,
		SessionID: conv.Config.SessionID,
		Timestamp: time.Now(),
	}

	ttsResult, err := m.voiceService.TextToSpeech(ctx, textInput)
	if err != nil {
		m.logger.Error("Text to speech failed",
			zap.String("conversation_id", conv.ID),
			zap.Error(err))
		m.sendEvent(conv, EventError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 
	conv.Status = StatusSpeaking
	audioChunk := AudioChunk{
		ID:        uuid.New().String(),
		Data:      ttsResult.AudioData,
		Sequence:  1,
		IsLast:    true,
		Timestamp: time.Now(),
	}

	select {
	case conv.AudioOutput <- audioChunk:
	case <-ctx.Done():
		return
	}

	// 
	m.sendEvent(conv, EventResponseEnd, map[string]interface{}{
		"response_text":  response,
		"audio_duration": ttsResult.Duration,
	})

	conv.Status = StatusListening
}

// handleControlCommands 
func (m *ConversationManager) handleControlCommands(ctx context.Context, conv *VoiceConversation) {
	for {
		select {
		case <-ctx.Done():
			return
		case control, ok := <-conv.Control:
			if !ok {
				return
			}

			m.processControlCommand(conv, control)
		}
	}
}

// processControlCommand 
func (m *ConversationManager) processControlCommand(conv *VoiceConversation, control ConversationControl) {
	switch control.Action {
	case ActionStart:
		conv.Status = StatusListening
	case ActionPause:
		conv.Status = StatusPaused
	case ActionResume:
		conv.Status = StatusListening
	case ActionStop:
		conv.Status = StatusEnded
	case ActionMute:
		// 
	case ActionUnmute:
		// 
	}

	m.sendEvent(conv, EventStatusChange, map[string]interface{}{
		"action": control.Action,
		"status": conv.Status,
	})
}

// handleSilenceDetection 
func (m *ConversationManager) handleSilenceDetection(ctx context.Context, conv *VoiceConversation) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	lastActivity := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Since(lastActivity) > conv.Config.SilenceTimeout {
				// 
				conv.Status = StatusEnded
				return
			}
		}
	}
}

// detectVoiceActivity 
func (m *ConversationManager) detectVoiceActivity(audioData []byte) bool {
	// 
	var sum int64
	for _, sample := range audioData {
		sum += int64(sample * sample)
	}

	rms := float64(sum) / float64(len(audioData))
	threshold := 1000.0 // 

	return rms > threshold
}

// generateSimpleResponse 
func (m *ConversationManager) generateSimpleResponse(text string) string {
	responses := []string{
		"" + text,
		"" + text,
		"",
		"",
		"",
	}

	// 
	return responses[len(text)%len(responses)]
}

// sendEvent 
func (m *ConversationManager) sendEvent(conv *VoiceConversation, eventType EventType, data map[string]interface{}) {
	event := ConversationEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case conv.Events <- event:
	default:
		// 
		m.logger.Warn("Event channel full, dropping event",
			zap.String("conversation_id", conv.ID),
			zap.String("event_type", string(eventType)))
	}
}

// GetConversation 
func (m *ConversationManager) GetConversation(id string) (*VoiceConversation, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conv, exists := m.conversations[id]
	return conv, exists
}

// ListConversations 
func (m *ConversationManager) ListConversations() []*VoiceConversation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conversations := make([]*VoiceConversation, 0, len(m.conversations))
	for _, conv := range m.conversations {
		conversations = append(conversations, conv)
	}

	return conversations
}

// StopConversation 
func (m *ConversationManager) StopConversation(id string) error {
	m.mutex.RLock()
	conv, exists := m.conversations[id]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("conversation %s not found", id)
	}

	// ?
	select {
	case conv.Control <- ConversationControl{
		Action:    ActionStop,
		Timestamp: time.Now(),
	}:
	default:
		// ?
		conv.Status = StatusEnded
	}

	return nil
}

// cleanupConversation 
func (m *ConversationManager) cleanupConversation(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if conv, exists := m.conversations[id]; exists {
		// 
		close(conv.AudioInput)
		close(conv.AudioOutput)
		close(conv.TextInput)
		close(conv.TextOutput)
		close(conv.Control)
		close(conv.Events)

		// 
		delete(m.conversations, id)

		m.logger.Info("Conversation cleaned up",
			zap.String("conversation_id", id))
	}
}

// cleanupRoutine 
func (m *ConversationManager) cleanupRoutine() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()

		var toDelete []string
		for id, conv := range m.conversations {
			// ?
			if conv.Status == StatusEnded ||
				(conv.EndTime == nil && time.Since(conv.StartTime) > conv.Config.MaxDuration) {
				toDelete = append(toDelete, id)
			}
		}

		for _, id := range toDelete {
			m.cleanupConversation(id)
		}

		m.mutex.Unlock()
	}
}

