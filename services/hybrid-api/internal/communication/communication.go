package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/hybrid-api/internal/lifecycle"
	"go.uber.org/zap"
)

// MessageType 消息类型
type MessageType string

const (
	TypeRequest  MessageType = "request"
	TypeResponse MessageType = "response"
	TypeEvent    MessageType = "event"
	TypeBroadcast MessageType = "broadcast"
)

// Message 消息结构
type Message struct {
	ID          string                 `json:"id"`
	Type        MessageType            `json:"type"`
	From        string                 `json:"from"`
	To          string                 `json:"to,omitempty"`
	Subject     string                 `json:"subject"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
	CorrelationID string               `json:"correlationId,omitempty"`
	ReplyTo     string                 `json:"replyTo,omitempty"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
}

// MessageHandler 消息处理器
type MessageHandler interface {
	HandleMessage(ctx context.Context, msg Message) (Message, error)
	CanHandle(subject string) bool
}

// MessageFilter 消息过滤器
type MessageFilter interface {
	Allow(msg Message) bool
}

// PluginCommunicationManager 插件通信管理器
type PluginCommunicationManager struct {
	handlers    map[string][]MessageHandler
	filters     []MessageFilter
	messages    chan Message
	responses   map[string]chan Message
	mu          sync.RWMutex
	logger      *zap.Logger
	lifecycleMgr *lifecycle.PluginLifecycleManager
	timeout     time.Duration
}

// NewPluginCommunicationManager 创建插件通信管理器
func NewPluginCommunicationManager(
	lifecycleMgr *lifecycle.PluginLifecycleManager,
	logger *zap.Logger,
) *PluginCommunicationManager {
	mgr := &PluginCommunicationManager{
		handlers:     make(map[string][]MessageHandler),
		filters:      make([]MessageFilter, 0),
		messages:     make(chan Message, 1000),
		responses:    make(map[string]chan Message),
		logger:       logger,
		lifecycleMgr: lifecycleMgr,
		timeout:      30 * time.Second,
	}

	// 启动消息处理循环
	go mgr.processMessages()

	return mgr
}

// RegisterHandler 注册消息处理器
func (m *PluginCommunicationManager) RegisterHandler(pluginID string, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.handlers[pluginID]; !exists {
		m.handlers[pluginID] = make([]MessageHandler, 0)
	}

	m.handlers[pluginID] = append(m.handlers[pluginID], handler)

	m.logger.Info("Message handler registered",
		zap.String("plugin_id", pluginID),
		zap.String("subject", fmt.Sprintf("%T", handler)))
}

// UnregisterHandler 注销消息处理器
func (m *PluginCommunicationManager) UnregisterHandler(pluginID string, handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	handlers, exists := m.handlers[pluginID]
	if !exists {
		return
	}

	// 查找并移除处理器
	for i, h := range handlers {
		if h == handler {
			m.handlers[pluginID] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	m.logger.Info("Message handler unregistered",
		zap.String("plugin_id", pluginID))
}

// AddFilter 添加消息过滤器
func (m *PluginCommunicationManager) AddFilter(filter MessageFilter) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.filters = append(m.filters, filter)
}

// SendMessage 发送消息
func (m *PluginCommunicationManager) SendMessage(ctx context.Context, msg Message) error {
	// 验证消息
	if err := m.validateMessage(msg); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// 应用过滤器
	if !m.applyFilters(msg) {
		return fmt.Errorf("message filtered out")
	}

	// 设置时间戳
	msg.Timestamp = time.Now()

	// 发送消息到处理队列
	select {
	case m.messages <- msg:
		return nil
	default:
		return fmt.Errorf("message queue is full")
	}
}

// SendRequest 发送请求并等待响应
func (m *PluginCommunicationManager) SendRequest(ctx context.Context, from, to, subject string, payload map[string]interface{}) (Message, error) {
	// 创建请求消息
	msgID := generateMessageID()
	correlationID := generateMessageID()
	
	msg := Message{
		ID:           msgID,
		Type:         TypeRequest,
		From:         from,
		To:           to,
		Subject:      subject,
		Payload:      payload,
		Timestamp:    time.Now(),
		CorrelationID: correlationID,
	}

	// 创建响应通道
	respChan := make(chan Message, 1)
	m.mu.Lock()
	m.responses[correlationID] = respChan
	m.mu.Unlock()

	// 确保在函数返回时清理响应通道
	defer func() {
		m.mu.Lock()
		delete(m.responses, correlationID)
		m.mu.Unlock()
		close(respChan)
	}()

	// 发送请求
	if err := m.SendMessage(ctx, msg); err != nil {
		return Message{}, fmt.Errorf("failed to send request: %w", err)
	}

	// 等待响应
	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(m.timeout):
		return Message{}, fmt.Errorf("request timeout after %v", m.timeout)
	case <-ctx.Done():
		return Message{}, fmt.Errorf("request canceled: %w", ctx.Err())
	}
}

// SendEvent 发送事件
func (m *PluginCommunicationManager) SendEvent(ctx context.Context, from, subject string, payload map[string]interface{}) error {
	msg := Message{
		ID:        generateMessageID(),
		Type:      TypeEvent,
		From:      from,
		Subject:   subject,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	return m.SendMessage(ctx, msg)
}

// SendBroadcast 发送广播消息
func (m *PluginCommunicationManager) SendBroadcast(ctx context.Context, from, subject string, payload map[string]interface{}) error {
	msg := Message{
		ID:        generateMessageID(),
		Type:      TypeBroadcast,
		From:      from,
		Subject:   subject,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	return m.SendMessage(ctx, msg)
}

// processMessages 处理消息循环
func (m *PluginCommunicationManager) processMessages() {
	for msg := range m.messages {
		go m.handleMessage(msg)
	}
}

// handleMessage 处理消息
func (m *PluginCommunicationManager) handleMessage(msg Message) {
	ctx := context.Background()

	switch msg.Type {
	case TypeRequest:
		m.handleRequest(ctx, msg)
	case TypeEvent:
		m.handleEvent(ctx, msg)
	case TypeBroadcast:
		m.handleBroadcast(ctx, msg)
	case TypeResponse:
		m.handleResponse(ctx, msg)
	default:
		m.logger.Warn("Unknown message type",
			zap.String("type", string(msg.Type)),
			zap.String("message_id", msg.ID))
	}
}

// handleRequest 处理请求
func (m *PluginCommunicationManager) handleRequest(ctx context.Context, msg Message) {
	m.mu.RLock()
	handlers, exists := m.handlers[msg.To]
	m.mu.RUnlock()

	if !exists {
		m.logger.Warn("No handlers found for plugin",
			zap.String("plugin_id", msg.To),
			zap.String("message_id", msg.ID))
		return
	}

	// 查找能处理该主题的处理器
	var handler MessageHandler
	for _, h := range handlers {
		if h.CanHandle(msg.Subject) {
			handler = h
			break
		}
	}

	if handler == nil {
		m.logger.Warn("No handler found for subject",
			zap.String("plugin_id", msg.To),
			zap.String("subject", msg.Subject),
			zap.String("message_id", msg.ID))
		return
	}

	// 处理请求
	resp, err := handler.HandleMessage(ctx, msg)
	if err != nil {
		m.logger.Error("Failed to handle request",
			zap.String("plugin_id", msg.To),
			zap.String("subject", msg.Subject),
			zap.String("message_id", msg.ID),
			zap.Error(err))
		return
	}

	// 发送响应
	respMsg := Message{
		ID:           generateMessageID(),
		Type:         TypeResponse,
		From:         msg.To,
		To:           msg.From,
		Subject:      msg.Subject,
		Payload:      resp.Payload,
		Timestamp:    time.Now(),
		CorrelationID: msg.CorrelationID,
	}

	if err := m.SendMessage(ctx, respMsg); err != nil {
		m.logger.Error("Failed to send response",
			zap.String("plugin_id", msg.To),
			zap.String("message_id", msg.ID),
			zap.Error(err))
	}
}

// handleEvent 处理事件
func (m *PluginCommunicationManager) handleEvent(ctx context.Context, msg Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 事件可以被多个插件处理
	for pluginID, handlers := range m.handlers {
		// 跳过发送者自己
		if pluginID == msg.From {
			continue
		}

		// 检查插件是否正在运行
		instance, err := m.lifecycleMgr.GetPluginInstance(pluginID)
		if err != nil || instance.State != lifecycle.StateRunning {
			continue
		}

		// 查找能处理该事件的处理器
		for _, handler := range handlers {
			if handler.CanHandle(msg.Subject) {
				go func(h MessageHandler, pID string) {
					if _, err := h.HandleMessage(ctx, msg); err != nil {
						m.logger.Error("Failed to handle event",
							zap.String("plugin_id", pID),
							zap.String("subject", msg.Subject),
							zap.String("message_id", msg.ID),
							zap.Error(err))
					}
				}(handler, pluginID)
			}
		}
	}
}

// handleBroadcast 处理广播
func (m *PluginCommunicationManager) handleBroadcast(ctx context.Context, msg Message) {
	m.handleEvent(ctx, msg) // 广播和事件处理逻辑相同
}

// handleResponse 处理响应
func (m *PluginCommunicationManager) handleResponse(ctx context.Context, msg Message) {
	m.mu.RLock()
	respChan, exists := m.responses[msg.CorrelationID]
	m.mu.RUnlock()

	if !exists {
		m.logger.Warn("No response channel found",
			zap.String("correlation_id", msg.CorrelationID),
			zap.String("message_id", msg.ID))
		return
	}

	// 发送响应到通道
	select {
	case respChan <- msg:
	default:
		m.logger.Warn("Response channel is full",
			zap.String("correlation_id", msg.CorrelationID),
			zap.String("message_id", msg.ID))
	}
}

// validateMessage 验证消息
func (m *PluginCommunicationManager) validateMessage(msg Message) error {
	if msg.ID == "" {
		return fmt.Errorf("message ID is required")
	}

	if msg.Type == "" {
		return fmt.Errorf("message type is required")
	}

	if msg.From == "" {
		return fmt.Errorf("message sender is required")
	}

	if msg.Subject == "" {
		return fmt.Errorf("message subject is required")
	}

	// 检查消息是否已过期
	if msg.ExpiresAt != nil && time.Now().After(*msg.ExpiresAt) {
		return fmt.Errorf("message has expired")
	}

	return nil
}

// applyFilters 应用过滤器
func (m *PluginCommunicationManager) applyFilters(msg Message) bool {
	m.mu.RLock()
	filters := make([]MessageFilter, len(m.filters))
	copy(filters, m.filters)
	m.mu.RUnlock()

	for _, filter := range filters {
		if !filter.Allow(msg) {
			return false
		}
	}

	return true
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// DefaultMessageFilter 默认消息过滤器
type DefaultMessageFilter struct {
	allowedSubjects map[string]bool
	blockedPlugins  map[string]bool
}

// NewDefaultMessageFilter 创建默认消息过滤器
func NewDefaultMessageFilter() *DefaultMessageFilter {
	return &DefaultMessageFilter{
		allowedSubjects: make(map[string]bool),
		blockedPlugins:  make(map[string]bool),
	}
}

// AllowSubject 允许特定主题
func (f *DefaultMessageFilter) AllowSubject(subject string) {
	f.allowedSubjects[subject] = true
}

// BlockPlugin 阻止特定插件
func (f *DefaultMessageFilter) BlockPlugin(pluginID string) {
	f.blockedPlugins[pluginID] = true
}

// Allow 实现MessageFilter接口
func (f *DefaultMessageFilter) Allow(msg Message) bool {
	// 检查是否被阻止的插件
	if f.blockedPlugins[msg.From] {
		return false
	}

	// 如果有允许的主题列表，检查主题是否在列表中
	if len(f.allowedSubjects) > 0 && !f.allowedSubjects[msg.Subject] {
		return false
	}

	return true
}

// LoggingMessageHandler 日志消息处理器
type LoggingMessageHandler struct {
	pluginID string
	logger   *zap.Logger
}

// NewLoggingMessageHandler 创建日志消息处理器
func NewLoggingMessageHandler(pluginID string, logger *zap.Logger) *LoggingMessageHandler {
	return &LoggingMessageHandler{
		pluginID: pluginID,
		logger:   logger,
	}
}

// HandleMessage 实现MessageHandler接口
func (h *LoggingMessageHandler) HandleMessage(ctx context.Context, msg Message) (Message, error) {
	payload, _ := json.Marshal(msg.Payload)
	h.logger.Info("Message received",
		zap.String("plugin_id", h.pluginID),
		zap.String("message_id", msg.ID),
		zap.String("type", string(msg.Type)),
		zap.String("from", msg.From),
		zap.String("subject", msg.Subject),
		zap.String("payload", string(payload)))

	// 返回确认响应
	return Message{
		ID:        generateMessageID(),
		Type:      TypeResponse,
		From:      h.pluginID,
		To:        msg.From,
		Subject:   msg.Subject,
		Payload:   map[string]interface{}{"received": true},
		Timestamp: time.Now(),
	}, nil
}

// CanHandle 实现MessageHandler接口
func (h *LoggingMessageHandler) CanHandle(subject string) bool {
	return true // 处理所有主题
}