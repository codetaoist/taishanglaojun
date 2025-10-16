package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codetaoist/taishanglaojun/core-services/monitoring/interfaces"
	"github.com/codetaoist/taishanglaojun/core-services/monitoring/models"
)

// NotificationManager ?
type NotificationManager struct {
	// 
	config *NotificationManagerConfig
	
	// 
	channels map[string]interfaces.NotificationChannel
	
	// 
	history []*models.NotificationHistory
	
	// 
	queue chan *models.Notification
	
	// 
	retryQueue chan *RetryNotification
	
	// ?
	mutex sync.RWMutex
	
	// ?
	running bool
	
	// ?
	ctx    context.Context
	cancel context.CancelFunc
	
	// 
	stats *NotificationManagerStats
	
	// ?
	rateLimiter *RateLimiter
	
	// 
	templateEngine *TemplateEngine
}

// NotificationManagerConfig ?
type NotificationManagerConfig struct {
	// 
	QueueSize     int           `yaml:"queue_size"`
	WorkerCount   int           `yaml:"worker_count"`
	BatchSize     int           `yaml:"batch_size"`
	BatchTimeout  time.Duration `yaml:"batch_timeout"`
	
	// 
	MaxRetries    int           `yaml:"max_retries"`
	RetryInterval time.Duration `yaml:"retry_interval"`
	RetryBackoff  float64       `yaml:"retry_backoff"`
	
	// 
	SendTimeout time.Duration `yaml:"send_timeout"`
	
	// 
	RateLimit     int           `yaml:"rate_limit"`
	RatePeriod    time.Duration `yaml:"rate_period"`
	BurstLimit    int           `yaml:"burst_limit"`
	
	// 
	HistoryRetention time.Duration `yaml:"history_retention"`
	MaxHistorySize   int           `yaml:"max_history_size"`
	
	// 
	TemplateDir string `yaml:"template_dir"`
	
	// 
	Channels map[string]interface{} `yaml:"channels"`
}

// RetryNotification 
type RetryNotification struct {
	Notification *models.Notification `json:"notification"`
	Attempts     int                  `json:"attempts"`
	LastAttempt  time.Time            `json:"last_attempt"`
	NextAttempt  time.Time            `json:"next_attempt"`
	Error        string               `json:"error"`
}

// NotificationManagerStats ?
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

// ChannelStats 
type ChannelStats struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Enabled      bool      `json:"enabled"`
	TotalSent    uint64    `json:"total_sent"`
	TotalFailed  uint64    `json:"total_failed"`
	LastSentTime time.Time `json:"last_sent_time"`
	LastError    string    `json:"last_error"`
}

// NewNotificationManager ?
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

// Initialize ?
func (nm *NotificationManager) Initialize() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	// ?
	if err := nm.templateEngine.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize template engine: %w", err)
	}
	
	// 
	if err := nm.initializeChannels(); err != nil {
		return fmt.Errorf("failed to initialize channels: %w", err)
	}
	
	return nil
}

// initializeChannels 
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

// createChannel 
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

// Start ?
func (nm *NotificationManager) Start() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	if nm.running {
		return fmt.Errorf("notification manager is already running")
	}
	
	// 
	for i := 0; i < nm.config.WorkerCount; i++ {
		go nm.worker()
	}
	
	// 
	go nm.retryWorker()
	
	// 
	go nm.cleanupWorker()
	
	nm.running = true
	nm.stats.Running = true
	
	return nil
}

// Stop ?
func (nm *NotificationManager) Stop() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	if !nm.running {
		return nil
	}
	
	// ?
	nm.cancel()
	
	// 
	close(nm.queue)
	close(nm.retryQueue)
	
	nm.running = false
	nm.stats.Running = false
	
	return nil
}

// worker 
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

// processNotification 
func (nm *NotificationManager) processNotification(notification *models.Notification) {
	// ?
	if !nm.rateLimiter.Allow() {
		nm.retryLater(notification, "rate limited")
		return
	}
	
	// 
	if err := nm.renderNotification(notification); err != nil {
		nm.recordFailure(notification, fmt.Sprintf("template render failed: %v", err))
		return
	}
	
	// 
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

// renderNotification 
func (nm *NotificationManager) renderNotification(notification *models.Notification) error {
	// 
	if notification.Title != "" {
		rendered, err := nm.templateEngine.Render(notification.Title, notification)
		if err != nil {
			return fmt.Errorf("failed to render title: %w", err)
		}
		notification.Title = rendered
	}
	
	// 
	if notification.Message != "" {
		rendered, err := nm.templateEngine.Render(notification.Message, notification)
		if err != nil {
			return fmt.Errorf("failed to render message: %w", err)
		}
		notification.Message = rendered
	}
	
	return nil
}

// getMatchingChannels ?
func (nm *NotificationManager) getMatchingChannels(notification *models.Notification) []interfaces.NotificationChannel {
	var channels []interfaces.NotificationChannel
	
	// 
	if len(notification.Channels) > 0 {
		for _, channelName := range notification.Channels {
			if channel, exists := nm.channels[channelName]; exists {
				channels = append(channels, channel)
			}
		}
		return channels
	}
	
	// ?
	for _, channel := range nm.channels {
		if nm.channelMatches(channel, notification) {
			channels = append(channels, channel)
		}
	}
	
	return channels
}

// channelMatches 
func (nm *NotificationManager) channelMatches(channel interfaces.NotificationChannel, notification *models.Notification) bool {
	// 
	// 
	return true
}

// sendToChannel 
func (nm *NotificationManager) sendToChannel(channel interfaces.NotificationChannel, notification *models.Notification) error {
	ctx, cancel := context.WithTimeout(nm.ctx, nm.config.SendTimeout)
	defer cancel()
	
	return channel.Send(ctx, notification)
}

// retryWorker 
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

// processRetries 
func (nm *NotificationManager) processRetries() {
	// 
}

// processRetry 
func (nm *NotificationManager) processRetry(retry *RetryNotification) {
	if time.Now().Before(retry.NextAttempt) {
		// 
		select {
		case nm.retryQueue <- retry:
		default:
			// ?
		}
		return
	}
	
	if retry.Attempts >= nm.config.MaxRetries {
		// 
		nm.recordFailure(retry.Notification, fmt.Sprintf("max retries exceeded: %s", retry.Error))
		return
	}
	
	// 
	nm.processNotification(retry.Notification)
}

// retryLater 
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
		// ?
		nm.recordFailure(notification, "retry queue full")
	}
}

// recordSuccess 
func (nm *NotificationManager) recordSuccess(notification *models.Notification) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	// 
	nm.stats.TotalSent++
	nm.stats.LastSentTime = time.Now()
	
	// 
	history := &models.NotificationHistory{
		ID:           nm.generateHistoryID(),
		NotificationID: notification.ID,
		Status:       "sent",
		Message:      "Notification sent successfully",
		Timestamp:    time.Now(),
	}
	
	nm.addHistory(history)
}

// recordFailure 
func (nm *NotificationManager) recordFailure(notification *models.Notification, errorMsg string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	
	// 
	nm.stats.TotalFailed++
	
	// 
	history := &models.NotificationHistory{
		ID:           nm.generateHistoryID(),
		NotificationID: notification.ID,
		Status:       "failed",
		Message:      errorMsg,
		Timestamp:    time.Now(),
	}
	
	nm.addHistory(history)
}

// updateChannelStats 
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

// addHistory 
func (nm *NotificationManager) addHistory(history *models.NotificationHistory) {
	nm.history = append(nm.history, history)
	
	// 
	if len(nm.history) > nm.config.MaxHistorySize {
		nm.history = nm.history[1:]
	}
}

// cleanupWorker 
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

// cleanup 
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

// generateHistoryID ID
func (nm *NotificationManager) generateHistoryID() string {
	return fmt.Sprintf("history_%d", time.Now().UnixNano())
}

//  interfaces.NotificationManager 

// Send 
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

// AddChannel 
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

// RemoveChannel 
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

// GetChannel 
func (nm *NotificationManager) GetChannel(name string) (interfaces.NotificationChannel, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	channel, exists := nm.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel %s not found", name)
	}
	
	return channel, nil
}

// ListChannels 
func (nm *NotificationManager) ListChannels() []interfaces.NotificationChannel {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	channels := make([]interfaces.NotificationChannel, 0, len(nm.channels))
	for _, channel := range nm.channels {
		channels = append(channels, channel)
	}
	
	return channels
}

// GetHistory 
func (nm *NotificationManager) GetHistory(limit int) ([]*models.NotificationHistory, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	if limit <= 0 || limit > len(nm.history) {
		limit = len(nm.history)
	}
	
	// 
	start := len(nm.history) - limit
	return nm.history[start:], nil
}

// GetStats 
func (nm *NotificationManager) GetStats() *NotificationManagerStats {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	
	nm.stats.QueueSize = len(nm.queue)
	nm.stats.RetryQueueSize = len(nm.retryQueue)
	nm.stats.LastUpdateTime = time.Now()
	
	return nm.stats
}

// IsRunning 
func (nm *NotificationManager) IsRunning() bool {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()
	return nm.running
}

