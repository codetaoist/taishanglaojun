package voice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConversationManager 对话管理器
type ConversationManager struct {
	voiceService   VoiceService
	nlpService     NLPService
	conversations  map[string]*VoiceConversation
	mutex          sync.RWMutex
	logger         *zap.Logger
	
	// 配置
	defaultConfig  ConversationConfig
	maxConcurrent  int
	cleanupInterval time.Duration
}

// NLPService 自然语言处理服务接口
type NLPService interface {
	ProcessText(ctx context.Context, text string, userID string) (string, error)
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error)
	ExtractKeywords(ctx context.Context, text string) ([]string, error)
	DetectIntent(ctx context.Context, text string) (string, error)
}

// NewConversationManager 创建对话管理器
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

	// 启动清理协程
	go manager.cleanupRoutine()

	return manager
}

// StartVoiceConversation 开始语音对话
func (m *ConversationManager) StartVoiceConversation(ctx context.Context, config ConversationConfig) (*VoiceConversation, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查并发限制
	if len(m.conversations) >= m.maxConcurrent {
		return nil, fmt.Errorf("maximum concurrent conversations reached: %d", m.maxConcurrent)
	}

	// 合并配置
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

	// 创建对话
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

	// 存储对话
	m.conversations[config.ID] = conversation

	// 启动对话处理协程
	go m.handleConversation(ctx, conversation)

	m.logger.Info("Voice conversation started",
		zap.String("conversation_id", config.ID),
		zap.String("user_id", config.UserID),
		zap.String("session_id", config.SessionID))

	return conversation, nil
}

// handleConversation 处理对话
func (m *ConversationManager) handleConversation(ctx context.Context, conv *VoiceConversation) {
	defer m.cleanupConversation(conv.ID)

	// 创建对话上下文
	convCtx, cancel := context.WithTimeout(ctx, conv.Config.MaxDuration)
	defer cancel()

	// 发送开始事件
	m.sendEvent(conv, EventType("conversation_started"), map[string]interface{}{
		"conversation_id": conv.ID,
		"start_time":     conv.StartTime,
	})

	// 启动各个处理协程
	var wg sync.WaitGroup

	// 音频输入处理
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.handleAudioInput(convCtx, conv)
	}()

	// 文本输入处理
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.handleTextInput(convCtx, conv)
	}()

	// 控制命令处理
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.handleControlCommands(convCtx, conv)
	}()

	// 静音检测
	if conv.Config.SilenceTimeout > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.handleSilenceDetection(convCtx, conv)
		}()
	}

	// 等待所有协程结束
	wg.Wait()

	// 更新对话状态
	conv.Status = StatusEnded
	endTime := time.Now()
	conv.EndTime = &endTime
	conv.Duration = endTime.Sub(conv.StartTime)

	// 发送结束事件
	m.sendEvent(conv, EventType("conversation_ended"), map[string]interface{}{
		"conversation_id": conv.ID,
		"end_time":       endTime,
		"duration":       conv.Duration,
	})

	m.logger.Info("Voice conversation ended",
		zap.String("conversation_id", conv.ID),
		zap.Duration("duration", conv.Duration))
}

// handleAudioInput 处理音频输入
func (m *ConversationManager) handleAudioInput(ctx context.Context, conv *VoiceConversation) {
	var audioBuffer []byte
	var lastActivity time.Time = time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case chunk, ok := <-conv.AudioInput:
			if !ok {
				return
			}

			lastActivity = time.Now()
			audioBuffer = append(audioBuffer, chunk.Data...)

			// 语音活动检测
			if conv.Config.EnableVAD && m.detectVoiceActivity(chunk.Data) {
				conv.Status = StatusListening
				m.sendEvent(conv, EventSpeechStart, map[string]interface{}{
					"timestamp": time.Now(),
				})
			}

			// 如果是最后一块或缓冲区足够大，则处理
			if chunk.IsLast || len(audioBuffer) >= 32000 { // 2秒的音频数据
				if len(audioBuffer) > 0 {
					m.processAudioChunk(ctx, conv, audioBuffer)
					audioBuffer = nil
				}
			}
		}
	}
}

// processAudioChunk 处理音频块
func (m *ConversationManager) processAudioChunk(ctx context.Context, conv *VoiceConversation, audioData []byte) {
	conv.Status = StatusProcessing

	// 创建音频输入
	audio := AudioInput{
		ID:        uuid.New().String(),
		Data:      audioData,
		Format:    FormatPCM,
		SampleRate: 16000,
		Channels:  1,
		Language:  conv.Config.Language,
		UserID:    conv.Config.UserID,
		SessionID: conv.Config.SessionID,
		Timestamp: time.Now(),
	}

	// 语音识别
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

	// 添加用户消息
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

	// 发送文本接收事件
	m.sendEvent(conv, EventTextReceived, map[string]interface{}{
		"text":       result.Text,
		"confidence": result.Confidence,
	})

	// 如果启用自动回复，则处理文本
	if conv.Config.AutoResponse {
		select {
		case conv.TextInput <- result.Text:
		case <-ctx.Done():
			return
		}
	}
}

// handleTextInput 处理文本输入
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

// processTextInput 处理文本输入
func (m *ConversationManager) processTextInput(ctx context.Context, conv *VoiceConversation, text string) {
	conv.Status = StatusProcessing

	// 发送响应开始事件
	m.sendEvent(conv, EventResponseStart, map[string]interface{}{
		"input_text": text,
	})

	var response string
	var err error

	// 如果启用NLP，则使用NLP服务处理
	if conv.Config.EnableNLP && m.nlpService != nil {
		response, err = m.nlpService.ProcessText(ctx, text, conv.Config.UserID)
		if err != nil {
			m.logger.Error("NLP processing failed",
				zap.String("conversation_id", conv.ID),
				zap.Error(err))
			response = "抱歉，我现在无法理解您的话，请稍后再试。"
		}
	} else {
		// 简单的回复逻辑
		response = m.generateSimpleResponse(text)
	}

	// 添加助手消息
	assistantMessage := ConversationMessage{
		ID:        uuid.New().String(),
		Type:      MessageTypeAssistant,
		Content:   response,
		Speaker:   "assistant",
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, assistantMessage)

	// 语音合成
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

	// 发送音频输出
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

	// 发送响应结束事件
	m.sendEvent(conv, EventResponseEnd, map[string]interface{}{
		"response_text": response,
		"audio_duration": ttsResult.Duration,
	})

	conv.Status = StatusListening
}

// handleControlCommands 处理控制命令
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

// processControlCommand 处理控制命令
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
		// 实现静音逻辑
	case ActionUnmute:
		// 实现取消静音逻辑
	}

	m.sendEvent(conv, EventStatusChange, map[string]interface{}{
		"action": control.Action,
		"status": conv.Status,
	})
}

// handleSilenceDetection 处理静音检测
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
				// 静音超时，结束对话
				conv.Status = StatusEnded
				return
			}
		}
	}
}

// detectVoiceActivity 检测语音活动
func (m *ConversationManager) detectVoiceActivity(audioData []byte) bool {
	// 简单的音量检测
	var sum int64
	for _, sample := range audioData {
		sum += int64(sample * sample)
	}
	
	rms := float64(sum) / float64(len(audioData))
	threshold := 1000.0 // 可调整的阈值
	
	return rms > threshold
}

// generateSimpleResponse 生成简单回复
func (m *ConversationManager) generateSimpleResponse(text string) string {
	responses := []string{
		"我听到您说：" + text,
		"您刚才说的是：" + text,
		"我理解您的意思了。",
		"请继续说。",
		"还有什么我可以帮助您的吗？",
	}
	
	// 简单的随机选择
	return responses[len(text)%len(responses)]
}

// sendEvent 发送事件
func (m *ConversationManager) sendEvent(conv *VoiceConversation, eventType EventType, data map[string]interface{}) {
	event := ConversationEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case conv.Events <- event:
	default:
		// 事件通道满了，丢弃事件
		m.logger.Warn("Event channel full, dropping event",
			zap.String("conversation_id", conv.ID),
			zap.String("event_type", string(eventType)))
	}
}

// GetConversation 获取对话
func (m *ConversationManager) GetConversation(id string) (*VoiceConversation, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	conv, exists := m.conversations[id]
	return conv, exists
}

// ListConversations 列出所有对话
func (m *ConversationManager) ListConversations() []*VoiceConversation {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	conversations := make([]*VoiceConversation, 0, len(m.conversations))
	for _, conv := range m.conversations {
		conversations = append(conversations, conv)
	}
	
	return conversations
}

// StopConversation 停止对话
func (m *ConversationManager) StopConversation(id string) error {
	m.mutex.RLock()
	conv, exists := m.conversations[id]
	m.mutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("conversation %s not found", id)
	}

	// 发送停止命令
	select {
	case conv.Control <- ConversationControl{
		Action:    ActionStop,
		Timestamp: time.Now(),
	}:
	default:
		// 控制通道满了，直接设置状态
		conv.Status = StatusEnded
	}

	return nil
}

// cleanupConversation 清理对话
func (m *ConversationManager) cleanupConversation(id string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if conv, exists := m.conversations[id]; exists {
		// 关闭所有通道
		close(conv.AudioInput)
		close(conv.AudioOutput)
		close(conv.TextInput)
		close(conv.TextOutput)
		close(conv.Control)
		close(conv.Events)
		
		// 从映射中删除
		delete(m.conversations, id)
		
		m.logger.Info("Conversation cleaned up",
			zap.String("conversation_id", id))
	}
}

// cleanupRoutine 清理协程
func (m *ConversationManager) cleanupRoutine() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		
		var toDelete []string
		for id, conv := range m.conversations {
			// 清理已结束或超时的对话
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