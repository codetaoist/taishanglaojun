package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/taishanglaojun/core-services/monitoring/models"
)

// NotificationManager 通知管理器
type NotificationManager struct {
	// 配置
	config *NotificationManagerConfig
	
	// 通知渠道
	channels map[string]interfaces.NotificationChannel
	
	// 通知历史
	history []*models.NotificationHistory
	
	// 通知队列
	queue chan *models.Notification
	
	// 重试队列
	retryQueue chan *RetryNotification
	
	// 同步锁
	mutex sync.RWMutex
	
	// 运行状态
	running bool
	
	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
	
	// 统计信息
	stats *NotificationManagerStats
	
	// 限流器
	rateLimiter *RateLimiter
	
	// 模板引擎
	templateEngine *TemplateEngine
}

// NotificationManagerConfig 通知管理器配置
type NotificationManagerConfig struct {
	// 队列配置
	QueueSize     int           `yaml:"queue_size"`
	WorkerCount   int           `yaml:"worker_count"`
	BatchSize     int           `yaml:"batch_size"`
	BatchTimeout  time.Duration `yaml:"batch_timeout"`
	
	// 重试配置
	MaxRetries    int           `yaml:"max_retries"`
	RetryInterval time.Duration `yaml:"retry_interval"`
	RetryBackoff  float64       `yaml:"retry_backoff"`
	
	// 超时配置
	SendTimeout time.Duration `yaml:"send_timeout"`
	
	// 限流配置
	RateLimit     int           `yaml:"rate_limit"`
	RatePeriod    time.Duration `yaml:"rate_period"`
	BurstLimit    int           `yaml:"burst_limit"`
	
	// 历史配置
	HistoryRetention time.Duration `yaml:"history_retention"`
	MaxHistorySize   int           `yaml:"max_history_size"`
	
	// 模板配置
	TemplateDir string `yaml:"template_dir"`
	
	// 渠道配置
	Channels map[string]interface{} `yaml:"channels"`
}

// RetryNotification 重试通知
type RetryNotification struct {
	Notification *models.Notification `json:"notification"`
	Attempts     int                  `json:"attempts"`
	LastAttempt  time.Time            `json:"last_attempt"`
	NextAttempt  time.Time            `json:"next_attempt"`
	Error        string               `json:"error"`
}

// NotificationManagerStats 通知管理器统计信息
type NotificationManagerStats struct {
	Running              bool      `json:"running"`
	QueueSize            int       `json:"queue_size"`
	RetryQueueSize       int       `json:"retry_queue_size"`
	TotalSent            uint64    `json:"total_sent"`
	TotalFailed          uint64    `json:"total_failed"`
	TotalRetries         uint64    `json:"total_retries"`
	ChannelStats         map[string]*ChannelStats `json:"channel_stats"`
	LastSentTime         time.Time `json:"last_sent_time"`
	LastUpdateTime       time.Time `json:"last_update_time"`
}

// ChannelStats 渠道统计信息
type ChannelStats struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Enabled      bool      `json:"enabled"`
	TotalSent    uint64    `json:"total_sent"`
	TotalFailed  uint64    `json:"total_failed"`
	LastSentTime time.Time `json:"last_sent_time"`
	LastError    string    `json:"last_error"`
}

// NewNotificationManager 创建通知管理器
func NewNotificationManager(config *NotificationManagerConfig) *NotificationManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &NotificationManager{
		config:         config,
		channels:       make(map[string]interfaces.NotificationChannel),
		history:        make([]*models.NotificationHistory, 0),
		queue:          make(chan *models.Notification, config.QueueSize),
		retryQueue:     make(chan *RetryNotification, config.QueueSize),
		ctx:            ctx,
		cancel:         cancel,
		running:        false,
		stats:          &NotificationManagerStats{
			ChannelStats: make(map[string]*ChannelStats),
		},
		rateLimiter:    NewRateLimiter(config.RateLimit, config.RatePeriod, config.BurstLimit),
		templateEngine: NewTemplateEngine(config.TemplateDir),
	}
}

// Initialize 初始化通知管理器
func (nm *NotificationManager) Initialize() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	// 初始化模板引擎
	if err := nm.templateEngine.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize template engine: %w", err)
	}
	
	// 初始化通知渠道
	if err := nm.initializeChannels(); err != nil {
		return fmt.Errorf("failed to initialize channels: %w", err)
	}
	
	return nil
}

// initializeChannels 初始化通知渠道
func (nm *NotificationManager) initializeChannels() error {
	for name, config := range nm.config.Channels {
		channel, err := nm.createChannel(name, config)
		if err != nil {
			return fmt.Errorf("failed to create channel %s: %w", name, err)
		}
		
		nm.channels[name] = channel
		nm.stats.ChannelStats[name] = &ChannelStats{
			Name:    name,
			Type:    channel.GetType(),
			Enabled: true,
		}
	}
	
	return nil
}

// createChannel 创建通知渠道
func (nm *NotificationManager) createChannel(name string, config interface{}) (interfaces.NotificationChannel, error) {
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid channel config for %s", name)
	}
	
	channelType, ok := configMap["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing channel type for %s", name)
	}
	
	switch channelType {
	case "email":
		return NewEmailChannel(name, configMap)
	case "webhook":
		return NewWebhookChannel(name, configMap)
	case "slack":
		return NewSlackChannel(name, configMap)
	case "dingtalk":
		return NewDingTalkChannel(name, configMap)
	case "wechat":
		return NewWeChatChannel(name, configMap)
	case "sms":
		return NewSMSChannel(name, configMap)
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// Start 启动通知管理器
func (nm *NotificationManager) Start() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	if nm.running {
		return fmt.Errorf("notification manager is already running")
	}
	
	// 启动工作协程
	for i := 0; i < nm.config.WorkerCount; i++ {
		go nm.worker()
	}
	
	// 启动重试协程
	go nm.retryWorker()
	
	// 启动清理协程
	go nm.cleanupWorker()
	
	nm.running = true
	nm.stats.Running = true
	
	return nil
}

// Stop 停止通知管理器
func (nm *NotificationManager) Stop() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	if !nm.running {
		return nil
	}
	
	// 取消上下文
	nm.cancel()
	
	// 关闭队列
	close(nm.queue)
	close(nm.retryQueue)
	
	nm.running = false
	nm.stats.Running = false
	
	return nil
}

// worker 工作协程
func (nm *NotificationManager) worker() {
	for {
		select {
		case <-nm.ctx.Done():
			return
		case notification := <-nm.queue:
			if notification != nil {
				nm.processNotification(notification)
			}
		}
	}
}

// processNotification 处理通知
func (nm *NotificationManager) processNotification(notification *models.Notification) {
	// 检查限流
	if !nm.rateLimiter.Allow() {
		nm.retryLater(notification, "rate limited")
		return
	}
	
	// 渲染模板
	if err := nm.renderNotification(notification); err != nil {
		nm.recordFailure(notification, fmt.Sprintf("template render failed: %v", err))
		return
	}
	
	// 发送到所有匹配的渠道
	channels := nm.getMatchingChannels(notification)
	if len(channels) == 0 {
		nm.recordFailure(notification, "no matching channels found")
		return
	}
	
	success := false
	var lastError error
	
	for _, channel := range channels {
		if err := nm.sendToChannel(channel, notification); err != nil {
			lastError = err
			nm.updateChannelStats(channel.GetName(), false, err.Error())
		} else {
			success = true
			nm.updateChannelStats(channel.GetName(), true, "")
		}
	}
	
	if success {
		nm.recordSuccess(notification)
	} else {
		nm.retryLater(notification, lastError.Error())
	}
}

// renderNotification 渲染通知模板
func (nm *NotificationManager) renderNotification(notification *models.Notification) error {
	// 渲染标题
	if notification.Title != "" {
		rendered, err := nm.templateEngine.Render(notification.Title, notification)
		if err != nil {
			return fmt.Errorf("failed to render title: %w", err)
		}
		notification.Title = rendered
	}
	
	// 渲染消息
	if notification.Message != "" {
		rendered, err := nm.templateEngine.Render(notification.Message, notification)
		if err != nil {
			return fmt.Errorf("failed to render message: %w", err)
		}
		notification.Message = rendered
	}
	
	return nil
}

// getMatchingChannels 获取匹配的渠道
func (nm *NotificationManager) getMatchingChannels(notification *models.Notification) []interfaces.NotificationChannel {
	var channels []interfaces.NotificationChannel
	
	// 如果指定了渠道，只使用指定的渠道
	if len(notification.Channels) > 0 {
		for _, channelName := range notification.Channels {
			if channel, exists := nm.channels[channelName]; exists {
				channels = append(channels, channel)
			}
		}
		return channels
	}
	
	// 否则根据严重级别和标签匹配渠道
	for _, channel := range nm.channels {
		if nm.channelMatches(channel, notification) {
			channels = append(channels, channel)
		}
	}
	
	return channels
}

// channelMatches 检查渠道是否匹配通知
func (nm *NotificationManager) channelMatches(channel interfaces.NotificationChannel, notification *models.Notification) bool {
	// 这里可以实现更复杂的匹配逻辑
	// 例如根据严重级别、标签、时间等条件匹配
	return true
}

// sendToChannel 发送到指定渠道
func (nm *NotificationManager) sendToChannel(channel interfaces.NotificationChannel, notification *models.Notification) error {
	ctx, cancel := context.WithTimeout(nm.ctx, nm.config.SendTimeout)
	defer cancel()
	
	return channel.Send(ctx, notification)
}

// retryWorker 重试工作协程
func (nm *NotificationManager) retryWorker() {
	ticker := time.NewTicker(nm.config.RetryInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.processRetries()
		case retryNotification := <-nm.retryQueue:
			if retryNotification != nil {
				nm.processRetry(retryNotification)
			}
		}
	}
}

// processRetries 处理重试
func (nm *NotificationManager) processRetries() {
	// 这里可以实现定期检查重试队列的逻辑
}

// processRetry 处理单个重试
func (nm *NotificationManager) processRetry(retry *RetryNotification) {
	if time.Now().Before(retry.NextAttempt) {
		// 还没到重试时间，重新放入队列
		select {
		case nm.retryQueue <- retry:
		default:
			// 队列满了，丢弃
		}
		return
	}
	
	if retry.Attempts >= nm.config.MaxRetries {
		// 超过最大重试次数，记录失败
		nm.recordFailure(retry.Notification, fmt.Sprintf("max retries exceeded: %s", retry.Error))
		return
	}
	
	// 重新处理通知
	nm.processNotification(retry.Notification)
}

// retryLater 稍后重试
func (nm *NotificationManager) retryLater(notification *models.Notification, errorMsg string) {
	retry := &RetryNotification{
		Notification: notification,
		Attempts:     1,
		LastAttempt:  time.Now(),
		NextAttempt:  time.Now().Add(nm.config.RetryInterval),
		Error:        errorMsg,
	}
	
	select {
	case nm.retryQueue <- retry:
		nm.stats.TotalRetries++
	default:
		// 重试队列满了，直接记录失败
		nm.recordFailure(notification, "retry queue full")
	}
}

// recordSuccess 记录成功
func (nm *NotificationManager) recordSuccess(notification *models.Notification) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	// 更新统计信息
	nm.stats.TotalSent++
	nm.stats.LastSentTime = time.Now()
	
	// 记录历史
	history := &models.NotificationHistory{
		ID:           nm.generateHistoryID(),
		NotificationID: notification.ID,
		Status:       "sent",
		Message:      "Notification sent successfully",
		Timestamp:    time.Now(),
	}
	
	nm.addHistory(history)
}

// recordFailure 记录失败
func (nm *NotificationManager) recordFailure(notification *models.Notification, errorMsg string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	// 更新统计信息
	nm.stats.TotalFailed++
	
	// 记录历史
	history := &models.NotificationHistory{
		ID:           nm.generateHistoryID(),
		NotificationID: notification.ID,
		Status:       "failed",
		Message:      errorMsg,
		Timestamp:    time.Now(),
	}
	
	nm.addHistory(history)
}

// updateChannelStats 更新渠道统计信息
func (nm *NotificationManager) updateChannelStats(channelName string, success bool, errorMsg string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	stats, exists := nm.stats.ChannelStats[channelName]
	if !exists {
		return
	}
	
	if success {
		stats.TotalSent++
		stats.LastSentTime = time.Now()
		stats.LastError = ""
	} else {
		stats.TotalFailed++
		stats.LastError = errorMsg
	}
}

// addHistory 添加历史记录
func (nm *NotificationManager) addHistory(history *models.NotificationHistory) {
	nm.history = append(nm.history, history)
	
	// 限制历史记录数量
	if len(nm.history) > nm.config.MaxHistorySize {
		nm.history = nm.history[1:]
	}
}

// cleanupWorker 清理工作协程
func (nm *NotificationManager) cleanupWorker() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.cleanup()
		}
	}
}

// cleanup 清理过期数据
func (nm *NotificationManager) cleanup() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	if nm.config.HistoryRetention <= 0 {
		return
	}
	
	cutoff := time.Now().Add(-nm.config.HistoryRetention)
	var newHistory []*models.NotificationHistory
	
	for _, history := range nm.history {
		if history.Timestamp.After(cutoff) {
			newHistory = append(newHistory, history)
		}
	}
	
	nm.history = newHistory
}

// generateHistoryID 生成历史ID
func (nm *NotificationManager) generateHistoryID() string {
	return fmt.Sprintf("history_%d", time.Now().UnixNano())
}

// 实现 interfaces.NotificationManager 接口

// Send 发送通知
func (nm *NotificationManager) Send(ctx context.Context, notification *models.Notification) error {
	if !nm.running {
		return fmt.Errorf("notification manager is not running")
	}
	
	select {
	case nm.queue <- notification:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("notification queue is full")
	}
}

// AddChannel 添加通知渠道
func (nm *NotificationManager) AddChannel(channel interfaces.NotificationChannel) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	name := channel.GetName()
	if _, exists := nm.channels[name]; exists {
		return fmt.Errorf("channel %s already exists", name)
	}
	
	nm.channels[name] = channel
	nm.stats.ChannelStats[name] = &ChannelStats{
		Name:    name,
		Type:    channel.GetType(),
		Enabled: true,
	}
	
	return nil
}

// RemoveChannel 移除通知渠道
func (nm *NotificationManager) RemoveChannel(name string) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	if _, exists := nm.channels[name]; !exists {
		return fmt.Errorf("channel %s not found", name)
	}
	
	delete(nm.channels, name)
	delete(nm.stats.ChannelStats, name)
	
	return nil
}

// GetChannel 获取通知渠道
func (nm *NotificationManager) GetChannel(name string) (interfaces.NotificationChannel, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	channel, exists := nm.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel %s not found", name)
	}
	
	return channel, nil
}

// ListChannels 列出所有通知渠道
func (nm *NotificationManager) ListChannels() []interfaces.NotificationChannel {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	channels := make([]interfaces.NotificationChannel, 0, len(nm.channels))
	for _, channel := range nm.channels {
		channels = append(channels, channel)
	}
	
	return channels
}

// GetHistory 获取通知历史
func (nm *NotificationManager) GetHistory(limit int) ([]*models.NotificationHistory, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	if limit <= 0 || limit > len(nm.history) {
		limit = len(nm.history)
	}
	
	// 返回最新的记录
	start := len(nm.history) - limit
	return nm.history[start:], nil
}

// GetStats 获取统计信息
func (nm *NotificationManager) GetStats() *NotificationManagerStats {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	nm.stats.QueueSize = len(nm.queue)
	nm.stats.RetryQueueSize = len(nm.retryQueue)
	nm.stats.LastUpdateTime = time.Now()
	
	return nm.stats
}

// IsRunning 检查是否运行中
func (nm *NotificationManager) IsRunning() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.running
}